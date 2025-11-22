package semantic

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

// ValidationPass implements Pass 3: Semantic Validation
//
// **Purpose**: Type-check all expressions, statements, and control flow now that
// all types are fully resolved. This is the "main" semantic analysis pass that
// ensures the program is semantically correct.
//
// **Responsibilities**:
// - Type-check all expressions (binary ops, unary ops, calls, indexing, member access)
// - Validate variable declarations
// - Validate assignments (type compatibility, const violations)
// - Type-check function calls (argument types, count)
// - Validate return statements (type compatibility, presence in all code paths)
// - Validate control flow statements (break/continue only in loops, etc.)
// - Check abstract method implementations in concrete classes
// - Validate interface method implementations
// - Validate visibility rules (private, protected, public)
// - Check constructor/destructor rules
// - Validate operator overloads
// - Check property getter/setter compatibility
// - Validate exception handling (raise, try/except/finally)
// - Annotate AST nodes with resolved types (store in SemanticInfo)
//
// **What it does NOT do**:
// - Resolve type names (already done in Pass 2)
// - Validate contracts (requires/ensures/invariant - done in Pass 4)
//
// **Dependencies**: Pass 2 (Type Resolution)
//
// **Inputs**:
// - TypeRegistry with fully resolved types
// - Symbols with resolved type references
// - Complete inheritance hierarchies
//
// **Outputs**:
// - SemanticInfo with type annotations on AST nodes
// - Errors for type mismatches, undefined variables, invalid operations
//
// **Example**:
//
//	var x: Integer;
//	var y: String;
//	x := y; // ERROR: Cannot assign String to Integer
//
//	class TFoo = class
//	  procedure Bar; virtual; abstract;
//	end;
//
//	class TBaz = class(TFoo)
//	end; // ERROR: TBaz must implement abstract method Bar
type ValidationPass struct{}

// NewValidationPass creates a new semantic validation pass.
func NewValidationPass() *ValidationPass {
	return &ValidationPass{}
}

// Name returns the name of this pass.
func (p *ValidationPass) Name() string {
	return "Pass 3: Semantic Validation"
}

// Run executes the semantic validation pass.
func (p *ValidationPass) Run(program *ast.Program, ctx *PassContext) error {
	validator := &statementValidator{
		ctx:     ctx,
		program: program,
	}

	// Walk all statements and validate them
	for _, stmt := range program.Statements {
		validator.validateStatement(stmt)
	}

	// Validate that concrete classes implement all abstract methods
	validator.validateAbstractImplementations()

	// Validate that classes correctly implement their interfaces
	validator.validateInterfaceImplementations()

	return nil
}

// statementValidator validates statements and expressions
type statementValidator struct {
	ctx     *PassContext
	program *ast.Program
}

// validateStatement validates a single statement
func (v *statementValidator) validateStatement(stmt ast.Statement) {
	if stmt == nil {
		return
	}

	switch s := stmt.(type) {
	case *ast.VarDeclStatement:
		v.validateVarDecl(s)
	case *ast.ConstDecl:
		v.validateConstDecl(s)
	case *ast.AssignmentStatement:
		v.validateAssignment(s)
	case *ast.ReturnStatement:
		v.validateReturn(s)
	case *ast.BreakStatement:
		v.validateBreak(s)
	case *ast.ContinueStatement:
		v.validateContinue(s)
	case *ast.ExitStatement:
		v.validateExit(s)
	case *ast.ExpressionStatement:
		// Type-check the expression
		v.checkExpression(s.Expression)
	case *ast.IfStatement:
		v.validateIf(s)
	case *ast.WhileStatement:
		v.validateWhile(s)
	case *ast.RepeatStatement:
		v.validateRepeat(s)
	case *ast.ForStatement:
		v.validateFor(s)
	case *ast.ForInStatement:
		v.validateForIn(s)
	case *ast.CaseStatement:
		v.validateCase(s)
	case *ast.BlockStatement:
		v.validateBlock(s)
	case *ast.FunctionDecl:
		v.validateFunction(s)
	case *ast.RaiseStatement:
		v.validateRaise(s)
	case *ast.TryStatement:
		v.validateTry(s)
	// Type and class declarations don't need validation here
	// (they were handled in Pass 1 and Pass 2)
	case *ast.ClassDecl, *ast.InterfaceDecl, *ast.EnumDecl, *ast.RecordDecl:
		// Skip - structural validation done in earlier passes
	case *ast.OperatorDecl, *ast.HelperDecl, *ast.SetDecl, *ast.ArrayDecl, *ast.TypeDeclaration:
		// Skip - type declarations validated in earlier passes
	case *ast.UsesClause, *ast.UnitDeclaration:
		// Skip - these are structural and don't need validation
	default:
		// Unknown statement type - skip
	}
}

// validateVarDecl validates a variable declaration
func (v *statementValidator) validateVarDecl(stmt *ast.VarDeclStatement) {
	// Resolve the type
	var varType types.Type
	if stmt.Type != nil {
		varType = v.resolveTypeExpression(stmt.Type)
		// In dual mode, skip validation for complex types that we can't resolve yet
		// (arrays, sets, function pointers, etc.). The old analyzer will handle them.
		// TODO: Implement full type resolution when we remove dual mode
		if varType == nil {
			// Complex type - skip validation
			return
		}
	}

	// If there's an initializer, check type compatibility with context-aware inference
	if stmt.Value != nil {
		// Pass expected type for context-aware type inference
		valueType := v.checkExpressionWithExpectedType(stmt.Value, varType)
		if varType != nil && valueType != nil {
			if !v.typesCompatible(varType, valueType) {
				v.ctx.AddError("cannot initialize variable of type %s with value of type %s",
					varType, valueType)
			}
		} else if varType == nil {
			// Type inference from initializer
			varType = valueType
		}
	}

	// Update symbol table with resolved type
	for _, name := range stmt.Names {
		if name != nil {
			existing, _ := v.ctx.Symbols.Resolve(name.Value)
			if existing != nil {
				existing.Type = varType
			}

			// Also add to the current scope (for local variables in functions)
			// This allows proper scoped variable resolution
			if varType != nil {
				v.ctx.DefineInCurrentScope(name.Value, varType)
			}
		}
	}
}

// validateAssignment validates an assignment statement
func (v *statementValidator) validateAssignment(stmt *ast.AssignmentStatement) {
	// Check target (left-hand side)
	targetType := v.checkExpression(stmt.Target)
	if targetType == nil {
		return // Error already reported
	}

	// Handle compound assignment operators (+=, -=, *=, /=, etc.)
	if v.isCompoundAssignment(stmt.Operator) {
		v.validateCompoundAssignment(stmt, targetType)
		return
	}

	// Regular assignment (:=)
	// Check value (right-hand side) with context-aware inference
	valueType := v.checkExpressionWithExpectedType(stmt.Value, targetType)
	if valueType == nil {
		return // Error already reported
	}

	// Check type compatibility
	if !v.typesCompatible(targetType, valueType) {
		v.ctx.AddError("cannot assign %s to %s", valueType, targetType)
	}

	// Check if target is assignable (not const, not readonly)
	// TODO: Implement const/readonly checking
}

// isCompoundAssignment checks if the operator is a compound assignment
func (v *statementValidator) isCompoundAssignment(op token.TokenType) bool {
	return op == token.PLUS_ASSIGN ||
		op == token.MINUS_ASSIGN ||
		op == token.TIMES_ASSIGN ||
		op == token.DIVIDE_ASSIGN ||
		op == token.PERCENT_ASSIGN ||
		op == token.CARET_ASSIGN ||
		op == token.AT_ASSIGN ||
		op == token.TILDE_ASSIGN
}

// validateCompoundAssignment validates compound assignment operators (+=, -=, *=, /=, etc.)
func (v *statementValidator) validateCompoundAssignment(stmt *ast.AssignmentStatement, targetType types.Type) {
	// Check value (right-hand side)
	valueType := v.checkExpression(stmt.Value)
	if valueType == nil {
		return // Error already reported
	}

	// Get the corresponding binary operator
	binaryOp := v.compoundToBinaryOperator(stmt.Operator)
	if binaryOp == "" {
		v.ctx.AddError("unsupported compound assignment operator %s", stmt.Operator)
		return
	}

	// Validate the operation: target OP value
	// For example, x += y is equivalent to x := x + y
	resultType := v.checkBinaryOperation(targetType, binaryOp, valueType)
	if resultType == nil {
		v.ctx.AddError("operator %s not applicable to types %s and %s", binaryOp, targetType, valueType)
		return
	}

	// Check that the result type is compatible with the target type
	// For example: var x: Integer; x += 5.0 should fail if Float can't be assigned to Integer
	if !v.typesCompatible(targetType, resultType) {
		v.ctx.AddError("cannot assign %s to %s", resultType, targetType)
	}
}

// compoundToBinaryOperator converts compound assignment operator to binary operator
func (v *statementValidator) compoundToBinaryOperator(op token.TokenType) string {
	switch op {
	case token.PLUS_ASSIGN:
		return "+"
	case token.MINUS_ASSIGN:
		return "-"
	case token.TIMES_ASSIGN:
		return "*"
	case token.DIVIDE_ASSIGN:
		return "/"
	case token.PERCENT_ASSIGN:
		return "mod"
	case token.CARET_ASSIGN:
		return "^"
	case token.AT_ASSIGN:
		return "@"
	case token.TILDE_ASSIGN:
		return "~"
	default:
		return ""
	}
}

// checkBinaryOperation validates a binary operation and returns the result type
func (v *statementValidator) checkBinaryOperation(leftType types.Type, operator string, rightType types.Type) types.Type {
	// Resolve to underlying types
	leftResolved := types.GetUnderlyingType(leftType)
	rightResolved := types.GetUnderlyingType(rightType)

	// String concatenation: + (check before numeric to handle strings)
	if operator == "+" {
		if v.isStringType(leftResolved) && v.isStringType(rightResolved) {
			return types.STRING
		}
	}

	// Arithmetic operators: +, -, *, /
	if operator == "+" || operator == "-" || operator == "*" || operator == "/" {
		// Both operands must be numeric
		if !v.isNumericType(leftResolved) || !v.isNumericType(rightResolved) {
			return nil
		}

		// Result type promotion: Float > Integer
		if v.isFloatType(leftResolved) || v.isFloatType(rightResolved) {
			return types.FLOAT
		}
		return types.INTEGER
	}

	// Modulo operator: mod (%)
	if operator == "mod" {
		// Both operands must be integers
		if !v.isIntegerType(leftResolved) || !v.isIntegerType(rightResolved) {
			return nil
		}
		return types.INTEGER
	}

	// Power operator: ^
	if operator == "^" {
		if !v.isNumericType(leftResolved) || !v.isNumericType(rightResolved) {
			return nil
		}
		return types.FLOAT
	}

	// Unsupported operator for these types
	return nil
}

// validateReturn validates a return statement
func (v *statementValidator) validateReturn(stmt *ast.ReturnStatement) {
	// Check if we're in a function
	if v.ctx.CurrentFunction == nil {
		v.ctx.AddError("return statement outside of function")
		return
	}

	// Get the function declaration
	fnDecl, ok := v.ctx.CurrentFunction.(*ast.FunctionDecl)
	if !ok {
		// Shouldn't happen, but be defensive
		return
	}

	// Check if this is a function (has return type declaration) or procedure (no return type)
	// Note: We check fnDecl.ReturnType != nil, NOT the resolved type, because
	// resolveTypeExpression may return nil for complex types (arrays, sets, function pointers, etc.)
	// that we can't fully resolve in this pass.
	isFunction := fnDecl.ReturnType != nil

	// Get function name for error messages (may be empty for lambdas)
	funcName := "lambda"
	if fnDecl.Name != nil {
		funcName = fnDecl.Name.Value
	}

	if isFunction {
		// Function: must have a return value
		if stmt.ReturnValue == nil {
			// Try to resolve the return type for a better error message
			expectedReturnType := v.resolveTypeExpression(fnDecl.ReturnType)
			if expectedReturnType != nil {
				v.ctx.AddError("function %s must return a value of type %s",
					funcName, expectedReturnType)
			} else {
				v.ctx.AddError("function %s must return a value",
					funcName)
			}
			return
		}

		// Type-check the return value if we can resolve the expected type
		expectedReturnType := v.resolveTypeExpression(fnDecl.ReturnType)
		if expectedReturnType != nil {
			actualReturnType := v.checkExpression(stmt.ReturnValue)
			if actualReturnType == nil {
				// Expression type couldn't be determined (error already reported)
				return
			}

			// Validate type compatibility
			if !v.typesCompatible(expectedReturnType, actualReturnType) {
				v.ctx.AddError("cannot return value of type %s from function %s (expected %s)",
					actualReturnType, funcName, expectedReturnType)
			}
		} else {
			// Can't resolve return type (complex type like array, set, etc.)
			// Still validate the return expression exists and is valid
			v.checkExpression(stmt.ReturnValue)
		}
	} else {
		// Procedure: should not have a return value
		// Skip this check for lambdas in dual mode (old analyzer handles it)
		if stmt.ReturnValue != nil && funcName != "lambda" {
			v.ctx.AddError("procedure %s cannot return a value (use a function instead)",
				funcName)
		}
	}
}

// validateBreak validates a break statement
func (v *statementValidator) validateBreak(stmt *ast.BreakStatement) {
	if v.ctx.LoopDepth == 0 && !v.ctx.InLoop {
		v.ctx.AddError("break statement outside of loop")
	}
}

// validateContinue validates a continue statement
func (v *statementValidator) validateContinue(stmt *ast.ContinueStatement) {
	if v.ctx.LoopDepth == 0 && !v.ctx.InLoop {
		v.ctx.AddError("continue statement outside of loop")
	}
}

// validateIf validates an if statement
func (v *statementValidator) validateIf(stmt *ast.IfStatement) {
	// Check condition is boolean
	condType := v.checkExpression(stmt.Condition)
	if condType != nil && !v.isBoolean(condType) {
		v.ctx.AddError("if condition must be boolean, got %s", condType)
	}

	// Validate consequence
	v.validateStatement(stmt.Consequence)

	// Validate alternative if present
	if stmt.Alternative != nil {
		v.validateStatement(stmt.Alternative)
	}
}

// validateWhile validates a while loop
func (v *statementValidator) validateWhile(stmt *ast.WhileStatement) {
	// Check condition is boolean
	condType := v.checkExpression(stmt.Condition)
	if condType != nil && !v.isBoolean(condType) {
		v.ctx.AddError("while condition must be boolean, got %s", condType)
	}

	// Validate body with loop context
	v.ctx.LoopDepth++
	v.ctx.InLoop = true
	v.validateStatement(stmt.Body)
	v.ctx.LoopDepth--
	if v.ctx.LoopDepth == 0 {
		v.ctx.InLoop = false
	}
}

// validateFor validates a for loop
func (v *statementValidator) validateFor(stmt *ast.ForStatement) {
	// Validate start value
	if stmt.Start != nil {
		startType := v.checkExpression(stmt.Start)
		if startType != nil && !v.isInteger(startType) {
			v.ctx.AddError("for loop start value must be integer, got %s", startType)
		}
	}

	// Validate end value
	if stmt.EndValue != nil {
		endType := v.checkExpression(stmt.EndValue)
		if endType != nil && !v.isInteger(endType) {
			v.ctx.AddError("for loop end value must be integer, got %s", endType)
		}
	}

	// Validate step (if present)
	if stmt.Step != nil {
		stepType := v.checkExpression(stmt.Step)
		if stepType != nil && !v.isInteger(stepType) {
			v.ctx.AddError("for loop step value must be integer, got %s", stepType)
		}
	}

	// Set the for loop variable context
	oldForLoopVar := v.ctx.CurrentForLoopVar
	if stmt.Variable != nil {
		v.ctx.CurrentForLoopVar = stmt.Variable.Value
	}
	defer func() { v.ctx.CurrentForLoopVar = oldForLoopVar }()

	// Validate body with loop context
	v.ctx.LoopDepth++
	v.ctx.InLoop = true
	v.validateStatement(stmt.Body)
	v.ctx.LoopDepth--
	if v.ctx.LoopDepth == 0 {
		v.ctx.InLoop = false
	}
}

// validateBlock validates a block statement
func (v *statementValidator) validateBlock(stmt *ast.BlockStatement) {
	for _, s := range stmt.Statements {
		v.validateStatement(s)
	}
}

// validateFunction validates a function declaration
func (v *statementValidator) validateFunction(decl *ast.FunctionDecl) {
	// Set current function context
	oldFunction := v.ctx.CurrentFunction
	v.ctx.CurrentFunction = decl
	defer func() { v.ctx.CurrentFunction = oldFunction }()

	// If this is a method (has a class name), set the current class context
	// This allows field access within the method body
	oldClass := v.ctx.CurrentClass
	if decl.ClassName != nil {
		// Look up the class type
		classTypeName := ident.Normalize(decl.ClassName.Value)
		if classType, ok := v.ctx.TypeRegistry.Resolve(classTypeName); ok {
			if ct, ok := classType.(*types.ClassType); ok {
				v.ctx.CurrentClass = ct
			}
		}
	}
	defer func() { v.ctx.CurrentClass = oldClass }()

	// Push a new function scope for local variables and parameters
	v.ctx.PushScope(ScopeFunction)
	defer v.ctx.PopScope()

	// Add function parameters to the current scope
	for _, param := range decl.Parameters {
		if param != nil && param.Name != nil {
			var paramType types.Type = types.VARIANT // default type
			if param.Type != nil {
				resolvedType := v.resolveTypeExpression(param.Type)
				if resolvedType != nil {
					paramType = resolvedType
				}
			}
			v.ctx.DefineInCurrentScope(param.Name.Value, paramType)
		}
	}

	// Validate function body
	if decl.Body != nil {
		v.validateStatement(decl.Body)

		// Check if this is a function (has return type)
		if decl.ReturnType != nil {
			// Validate that all code paths return a value
			// Note: This check is disabled in dual mode to avoid conflicts with old analyzer
			// TODO: Re-enable once old analyzer is removed
			if false && !v.allPathsReturn(decl.Body) {
				expectedReturnType := v.resolveTypeExpression(decl.ReturnType)
				if expectedReturnType != nil {
					v.ctx.AddError("not all code paths return a value in function %s (expected %s)",
						decl.Name.Value, expectedReturnType)
				} else {
					v.ctx.AddError("not all code paths return a value in function %s",
						decl.Name.Value)
				}
			}
		}
	}
}

// allPathsReturn checks if all code paths in a statement return a value.
// This is used to validate that functions (not procedures) return on all paths.
func (v *statementValidator) allPathsReturn(stmt ast.Statement) bool {
	if stmt == nil {
		return false
	}

	switch s := stmt.(type) {
	case *ast.ReturnStatement:
		// Return statement always returns
		return true

	case *ast.ExitStatement:
		// Exit statement also terminates the function
		return true

	case *ast.RaiseStatement:
		// Raise statement terminates execution
		return true

	case *ast.BlockStatement:
		// Check if any statement in the block guarantees a return
		for _, blockStmt := range s.Statements {
			if v.allPathsReturn(blockStmt) {
				return true
			}
		}
		return false

	case *ast.IfStatement:
		// If-else: both branches must return
		if s.Alternative != nil {
			return v.allPathsReturn(s.Consequence) && v.allPathsReturn(s.Alternative)
		}
		// If without else doesn't guarantee return
		return false

	case *ast.CaseStatement:
		// Case statement: all branches (including else) must return
		if s.Else == nil {
			// No else branch, so not all paths covered
			return false
		}

		// Check all case branches
		for _, branch := range s.Cases {
			if !v.allPathsReturn(branch.Statement) {
				return false
			}
		}

		// Check else branch
		return v.allPathsReturn(s.Else)

	case *ast.TryStatement:
		// Try-except-finally: complex control flow
		// The try block must return, AND:
		// - If there are exception handlers, they all must return
		// - The finally block doesn't affect return (it always executes)

		tryReturns := v.allPathsReturn(s.TryBlock)
		if !tryReturns {
			return false
		}

		// If there's an except clause, check all exception paths
		if s.ExceptClause != nil {
			// If there are specific handlers, all must return
			if len(s.ExceptClause.Handlers) > 0 {
				for _, handler := range s.ExceptClause.Handlers {
					if !v.allPathsReturn(handler.Statement) {
						return false
					}
				}
			}
			// If there's an else block (catch-all handler), it must also return
			if s.ExceptClause.ElseBlock != nil {
				if !v.allPathsReturn(s.ExceptClause.ElseBlock) {
					return false
				}
			}
		}

		// Finally block doesn't affect whether we return
		// (it executes regardless)
		return true

	case *ast.WhileStatement, *ast.ForStatement, *ast.RepeatStatement:
		// Loops don't guarantee execution (might be skipped)
		// So they don't guarantee a return
		return false

	default:
		// Other statements don't return
		return false
	}
}

// validateConstDecl validates a constant declaration
func (v *statementValidator) validateConstDecl(stmt *ast.ConstDecl) {
	// Resolve declared type if specified
	var declaredType types.Type
	if stmt.Type != nil {
		declaredType = v.resolveTypeExpression(stmt.Type)
	}

	// Type-check the initializer with context-aware inference
	var finalType types.Type = declaredType
	if stmt.Value != nil {
		valueType := v.checkExpressionWithExpectedType(stmt.Value, declaredType)

		// If type is specified, validate compatibility
		if declaredType != nil && valueType != nil {
			if !v.typesCompatible(declaredType, valueType) {
				v.ctx.AddError("cannot initialize constant of type %s with value of type %s",
					declaredType, valueType)
			}
		} else if declaredType == nil {
			// Type inference from initializer
			finalType = valueType
		}
	}

	// Add constant to the current scope
	if stmt.Name != nil && finalType != nil {
		v.ctx.DefineInCurrentScope(stmt.Name.Value, finalType)
	}
}

// validateExit validates an exit statement
func (v *statementValidator) validateExit(stmt *ast.ExitStatement) {
	if v.ctx.CurrentFunction == nil {
		v.ctx.AddError("exit statement outside of function")
	}
}

// validateRepeat validates a repeat-until loop
func (v *statementValidator) validateRepeat(stmt *ast.RepeatStatement) {
	// Validate condition is boolean
	condType := v.checkExpression(stmt.Condition)
	if condType != nil && !v.isBoolean(condType) {
		v.ctx.AddError("repeat condition must be boolean, got %s", condType)
	}

	// Validate body with loop context
	v.ctx.LoopDepth++
	v.ctx.InLoop = true
	v.validateStatement(stmt.Body)
	v.ctx.LoopDepth--
	if v.ctx.LoopDepth == 0 {
		v.ctx.InLoop = false
	}
}

// validateForIn validates a for-in loop
func (v *statementValidator) validateForIn(stmt *ast.ForInStatement) {
	// Validate collection expression
	collectionType := v.checkExpression(stmt.Collection)
	if collectionType != nil {
		// TODO: Validate collection is iterable (array, set, string)
		_ = collectionType
	}

	// Set the for loop variable context
	oldForLoopVar := v.ctx.CurrentForLoopVar
	if stmt.Variable != nil {
		v.ctx.CurrentForLoopVar = stmt.Variable.Value
	}
	defer func() { v.ctx.CurrentForLoopVar = oldForLoopVar }()

	// Validate body with loop context
	v.ctx.LoopDepth++
	v.ctx.InLoop = true
	v.validateStatement(stmt.Body)
	v.ctx.LoopDepth--
	if v.ctx.LoopDepth == 0 {
		v.ctx.InLoop = false
	}
}

// validateCase validates a case statement
func (v *statementValidator) validateCase(stmt *ast.CaseStatement) {
	// Validate selector expression
	selectorType := v.checkExpression(stmt.Expression)
	if selectorType == nil {
		return
	}

	// Validate each case branch
	for _, branch := range stmt.Cases {
		// Validate case values
		for _, value := range branch.Values {
			valueType := v.checkExpression(value)
			if valueType != nil && !v.typesCompatible(selectorType, valueType) {
				v.ctx.AddError("case value type %s incompatible with selector type %s",
					valueType, selectorType)
			}
		}

		// Validate branch statement
		v.validateStatement(branch.Statement)
	}

	// Validate else clause if present
	if stmt.Else != nil {
		v.validateStatement(stmt.Else)
	}
}

// validateRaise validates a raise statement
func (v *statementValidator) validateRaise(stmt *ast.RaiseStatement) {
	// If there's an exception expression, validate it
	if stmt.Exception != nil {
		exceptionType := v.checkExpression(stmt.Exception)

		// TODO: Validate exception type is or derives from Exception
		_ = exceptionType
	} else {
		// Re-raise without argument is only valid inside exception handler
		if !v.ctx.InExceptionHandler {
			v.ctx.AddError("'raise' without exception object can only be used inside exception handler")
		}
	}
}

// validateTry validates a try-except-finally statement
func (v *statementValidator) validateTry(stmt *ast.TryStatement) {
	// Validate try block
	v.validateStatement(stmt.TryBlock)

	// Validate except clause
	if stmt.ExceptClause != nil {
		oldInHandler := v.ctx.InExceptionHandler
		v.ctx.InExceptionHandler = true

		// Validate exception handlers
		for _, handler := range stmt.ExceptClause.Handlers {
			// Validate exception type if specified
			if handler.ExceptionType != nil {
				exceptionType := v.resolveTypeExpression(handler.ExceptionType)
				// TODO: Validate it's an Exception class
				_ = exceptionType
			}

			// Validate handler statement
			v.validateStatement(handler.Statement)
		}

		// Validate else block in except clause if present
		if stmt.ExceptClause.ElseBlock != nil {
			v.validateStatement(stmt.ExceptClause.ElseBlock)
		}

		v.ctx.InExceptionHandler = oldInHandler
	}

	// Validate finally clause if present
	if stmt.FinallyClause != nil && stmt.FinallyClause.Block != nil {
		oldInFinally := v.ctx.InFinallyBlock
		v.ctx.InFinallyBlock = true
		v.validateStatement(stmt.FinallyClause.Block)
		v.ctx.InFinallyBlock = oldInFinally
	}
}

// validateAbstractImplementations validates that concrete classes implement abstract methods
func (v *statementValidator) validateAbstractImplementations() {
	// Get all class types
	allTypes := v.ctx.TypeRegistry.AllDescriptors()

	for _, desc := range allTypes {
		classType, ok := desc.Type.(*types.ClassType)
		if !ok || classType.IsAbstract {
			continue // Skip non-classes and abstract classes
		}

		// Check for unimplemented abstract methods
		unimplemented := v.getUnimplementedAbstractMethods(classType)
		if len(unimplemented) > 0 {
			// Report error for each unimplemented abstract method
			for _, methodName := range unimplemented {
				v.ctx.AddError("concrete class '%s' does not implement abstract method '%s'",
					classType.Name, methodName)
			}
		}
	}
}

// validateInterfaceImplementations validates that classes correctly implement their interfaces
func (v *statementValidator) validateInterfaceImplementations() {
	// Get all class types
	allTypes := v.ctx.TypeRegistry.AllDescriptors()

	for _, desc := range allTypes {
		classType, ok := desc.Type.(*types.ClassType)
		if !ok {
			continue
		}

		// Check each interface the class claims to implement
		for _, iface := range classType.Interfaces {
			// Get all methods required by the interface
			requiredMethods := types.GetAllInterfaceMethods(iface)

			// Check if the class implements all required methods
			for methodName, methodType := range requiredMethods {
				// Look up method in class
				classMethod, found := classType.GetMethod(methodName)
				if !found {
					v.ctx.AddError("class '%s' does not implement interface method '%s' from '%s'",
						classType.Name, methodName, iface.Name)
					continue
				}

				// Check method signature compatibility
				if !v.methodSignaturesCompatible(classMethod, methodType) {
					v.ctx.AddError("class '%s' method '%s' has incompatible signature for interface '%s'",
						classType.Name, methodName, iface.Name)
				}
			}
		}
	}
}

// getUnimplementedAbstractMethods returns a list of abstract methods that are not implemented
func (v *statementValidator) getUnimplementedAbstractMethods(classType *types.ClassType) []string {
	unimplemented := []string{}

	// Collect all abstract methods from parent chain
	abstractMethods := v.collectAbstractMethods(classType.Parent)

	// Check which ones are not implemented in this class
	for methodName := range abstractMethods {
		lowerMethodName := ident.Normalize(methodName)
		hasOwnMethod := len(classType.MethodOverloads[lowerMethodName]) > 0

		if !hasOwnMethod {
			// Method not defined in this class at all
			unimplemented = append(unimplemented, methodName)
		} else {
			// Method is defined - check if it's still abstract or reintroduced
			if isReintroduce, exists := classType.ReintroduceMethods[lowerMethodName]; exists && isReintroduce {
				// Method reintroduces (hides) parent method without implementing it
				unimplemented = append(unimplemented, methodName)
			} else if isAbstract, exists := classType.AbstractMethods[lowerMethodName]; exists && isAbstract {
				// Still abstract in this class
				unimplemented = append(unimplemented, methodName)
			}
		}
	}

	return unimplemented
}

// collectAbstractMethods recursively collects all abstract methods from the parent chain
func (v *statementValidator) collectAbstractMethods(parent *types.ClassType) map[string]bool {
	abstractMethods := make(map[string]bool)

	if parent == nil {
		return abstractMethods
	}

	// Collect abstract methods from this parent
	for methodName, isAbstract := range parent.AbstractMethods {
		if isAbstract {
			abstractMethods[methodName] = true
		}
	}

	// Recursively collect from grandparents
	grandparentMethods := v.collectAbstractMethods(parent.Parent)
	for methodName := range grandparentMethods {
		// Only add if not overridden (not abstract) in this parent
		lowerMethodName := ident.Normalize(methodName)
		if isAbstract, exists := parent.AbstractMethods[lowerMethodName]; !exists || isAbstract {
			abstractMethods[methodName] = true
		}
	}

	return abstractMethods
}

// methodSignaturesCompatible checks if two method signatures are compatible
func (v *statementValidator) methodSignaturesCompatible(m1, m2 *types.FunctionType) bool {
	// Check parameter count
	if len(m1.Parameters) != len(m2.Parameters) {
		return false
	}

	// Check parameter types
	for i := range m1.Parameters {
		if !v.typesCompatible(m1.Parameters[i], m2.Parameters[i]) {
			return false
		}
	}

	// Check return type
	if m1.ReturnType == nil && m2.ReturnType == nil {
		return true
	}
	if m1.ReturnType == nil || m2.ReturnType == nil {
		return false
	}

	return v.typesCompatible(m1.ReturnType, m2.ReturnType)
}

// checkExpression type-checks an expression and returns its type
func (v *statementValidator) checkExpression(expr ast.Expression) types.Type {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return types.INTEGER
	case *ast.FloatLiteral:
		return types.FLOAT
	case *ast.StringLiteral:
		return types.STRING
	case *ast.BooleanLiteral:
		return types.BOOLEAN
	case *ast.CharLiteral:
		// Character literals are treated as single-character strings in DWScript
		return types.STRING
	case *ast.NilLiteral:
		return types.NIL
	case *ast.Identifier:
		return v.checkIdentifier(e)
	case *ast.BinaryExpression:
		return v.checkBinaryExpression(e)
	case *ast.UnaryExpression:
		return v.checkUnaryExpression(e)
	case *ast.GroupedExpression:
		return v.checkExpression(e.Expression)
	case *ast.CallExpression:
		return v.checkCallExpression(e)
	case *ast.IndexExpression:
		return v.checkIndexExpression(e)
	case *ast.MemberAccessExpression:
		return v.checkMemberAccessExpression(e)
	case *ast.MethodCallExpression:
		return v.checkMethodCallExpression(e)
	case *ast.NewExpression:
		return v.checkNewExpression(e)
	case *ast.NewArrayExpression:
		return v.checkNewArrayExpression(e)
	case *ast.ArrayLiteralExpression:
		return v.checkArrayLiteral(e, nil)
	case *ast.RecordLiteralExpression:
		return v.checkRecordLiteral(e, nil)
	case *ast.SetLiteral:
		return v.checkSetLiteral(e, nil)
	case *ast.IsExpression:
		return v.checkIsExpression(e)
	case *ast.AsExpression:
		return v.checkAsExpression(e)
	case *ast.ImplementsExpression:
		return v.checkImplementsExpression(e)
	case *ast.IfExpression:
		return v.checkIfExpression(e)
	case *ast.SelfExpression:
		return v.checkSelfExpression(e)
	case *ast.InheritedExpression:
		return v.checkInheritedExpression(e)
	case *ast.AddressOfExpression:
		return v.checkAddressOfExpression(e)
	case *ast.LambdaExpression:
		return v.checkLambdaExpression(e, nil)
	case *ast.OldExpression:
		return v.checkOldExpression(e)
	default:
		// Unknown expression type - log for debugging but don't error
		// Some expression types may not need validation
		return nil
	}
}

// checkExpressionWithExpectedType type-checks an expression with optional expected type context.
// This enables context-sensitive type inference for expressions that benefit from knowing
// the expected type (e.g., lambda parameters, nil literals, record literals, integer→float).
//
// Context-aware analysis is used in:
//   - Variable declarations: var x: T := <expr>  (expected type = T)
//   - Assignments: x := <expr>                    (expected type = type of x)
//   - Function arguments: f(<expr>)               (expected type = parameter type)
//   - Return statements: return <expr>            (expected type = function return type)
//
// Parameters:
//   - expr: The expression to analyze
//   - expectedType: The expected type from context (may be nil if no context available)
//
// Returns:
//   - The actual type of the expression, or nil if analysis failed
func (v *statementValidator) checkExpressionWithExpectedType(expr ast.Expression, expectedType types.Type) types.Type {
	if expr == nil {
		return nil
	}

	// Handle expressions that benefit from expected type context
	switch e := expr.(type) {
	case *ast.NilLiteral:
		// If expected type is a class, interface, or function pointer, return that type
		// instead of generic NIL for better type specificity
		if expectedType != nil {
			underlyingType := types.GetUnderlyingType(expectedType)
			typeKind := underlyingType.TypeKind()
			if typeKind == "CLASS" || typeKind == "INTERFACE" || typeKind == "FUNCTION_POINTER" {
				return expectedType
			}
		}
		return types.NIL

	case *ast.IntegerLiteral:
		// If expected type is Float, treat integer literal as float for better compatibility
		if expectedType != nil {
			underlyingType := types.GetUnderlyingType(expectedType)
			if underlyingType.TypeKind() == "FLOAT" {
				return types.FLOAT
			}
		}
		return types.INTEGER

	case *ast.RecordLiteralExpression:
		// Record literals can use expected type to validate fields
		return v.checkRecordLiteral(e, expectedType)

	case *ast.ArrayLiteralExpression:
		// Array literals can infer element type from expected array type
		return v.checkArrayLiteral(e, expectedType)

	case *ast.SetLiteral:
		// Set literals can infer element type from expected set type
		return v.checkSetLiteral(e, expectedType)

	case *ast.LambdaExpression:
		// Lambda expressions can infer parameter types from expected function pointer type
		return v.checkLambdaExpression(e, expectedType)

	default:
		// For all other expression types, use regular type checking without context
		return v.checkExpression(expr)
	}
}

// checkIdentifier checks an identifier reference
func (v *statementValidator) checkIdentifier(expr *ast.Identifier) types.Type {
	if v.ctx == nil {
		return nil // Defensive check
	}

	// Check for special implicit variables first
	idLower := strings.ToLower(expr.Value)

	// Handle the implicit 'Result' variable in functions/methods
	if idLower == "result" {
		if v.ctx.CurrentFunction != nil {
			// Try to get the return type from the current function
			if fnDecl, ok := v.ctx.CurrentFunction.(*ast.FunctionDecl); ok && fnDecl != nil {
				if fnDecl.ReturnType != nil {
					// This is a function with explicit return type
					returnType := v.resolveTypeExpression(fnDecl.ReturnType)
					if returnType != nil {
						return returnType
					}
				}
				// Return Variant as a safe fallback for:
				// - Lambdas without explicit return type (type inference)
				// - Unresolved return types (complex types)
				return types.VARIANT
			}
		}
		// If not in a function, fall through to normal resolution
	}

	// Check if it's the current for loop variable
	if v.ctx.CurrentForLoopVar != "" {
		if strings.EqualFold(expr.Value, v.ctx.CurrentForLoopVar) {
			// For loop variables are always Integer in DWScript
			return types.INTEGER
		}
	}

	// First, try to resolve in the scoped symbol tables (local variables, parameters)
	// This searches the current scope and all parent scopes up to global
	if scopedType, found := v.ctx.LookupInScopes(expr.Value); found {
		return scopedType
	}

	// Try to resolve in the global symbol table (for global variables and functions)
	// This is separate from scopes to maintain compatibility with Pass 1 & 2
	if v.ctx.Symbols != nil {
		if symbol, ok := v.ctx.Symbols.Resolve(expr.Value); ok {
			return symbol.Type
		}
	}

	// If we're in a method/constructor, check if this is a field access (unqualified)
	if v.ctx.CurrentClass != nil {
		// Look up the field in the current class (including inherited fields)
		classType := v.ctx.CurrentClass
		fieldNameLower := strings.ToLower(expr.Value)

		// Check current class and all parent classes for the field
		for classType != nil {
			// Check if it's a field
			if fieldType, found := classType.Fields[fieldNameLower]; found {
				// Unqualified field access - should use Self.FieldName for clarity
				// but it's allowed in DWScript
				return fieldType
			}

			// Check parent class
			if classType.Parent != nil {
				classType = classType.Parent
			} else {
				break
			}
		}
	}

	// Check if this is a type name in the TypeRegistry
	// Type names can be used in member access expressions (e.g., TClassName.Create())
	// or in type operators (e.g., obj is TClassName)
	if v.ctx.TypeRegistry != nil {
		if typ, ok := v.ctx.TypeRegistry.Resolve(expr.Value); ok {
			// This is a valid type name - return the type itself
			// (will be used in member access for constructors/class methods)
			return typ
		}
	}

	// Variable not found in any scope or type registry
	v.ctx.AddError("undefined variable '%s'", expr.Value)
	return nil
}

// checkBinaryExpression checks a binary expression
func (v *statementValidator) checkBinaryExpression(expr *ast.BinaryExpression) types.Type {
	leftType := v.checkExpression(expr.Left)
	rightType := v.checkExpression(expr.Right)

	if leftType == nil || rightType == nil {
		return nil // Error already reported
	}

	// Check operator compatibility
	switch expr.Operator {
	case "+":
		// Task 6.1.2.2: + can be used for set union, string concatenation, or numeric addition
		// Check for set union first
		leftSetType, leftIsSet := leftType.(*types.SetType)
		rightSetType, rightIsSet := rightType.(*types.SetType)
		if leftIsSet && rightIsSet {
			// Set union - both operands must have the same element type
			if !leftSetType.ElementType.Equals(rightSetType.ElementType) {
				v.ctx.AddError("set union requires sets with the same element type, got set of %s and set of %s",
					leftSetType.ElementType, rightSetType.ElementType)
				return nil
			}
			return leftSetType
		}
		if leftIsSet || rightIsSet {
			v.ctx.AddError("operator + cannot mix set and non-set types")
			return nil
		}

		// String concatenation
		if v.isString(leftType) || v.isString(rightType) {
			// String concatenation - both operands should be strings (or convertible to string)
			return types.STRING
		}

		// Numeric addition
		if !v.isNumeric(leftType) || !v.isNumeric(rightType) {
			v.ctx.AddError("operator + requires numeric, string, or set types, got %s and %s",
				leftType, rightType)
			return nil
		}
		// Result type is the "wider" of the two operands
		if v.isFloat(leftType) || v.isFloat(rightType) {
			return types.FLOAT
		}
		return types.INTEGER

	case "-":
		// Task 6.1.2.2: - can be used for set difference or numeric subtraction
		// Check for set difference first
		leftSetType, leftIsSet := leftType.(*types.SetType)
		rightSetType, rightIsSet := rightType.(*types.SetType)
		if leftIsSet && rightIsSet {
			// Set difference - both operands must have the same element type
			if !leftSetType.ElementType.Equals(rightSetType.ElementType) {
				v.ctx.AddError("set difference requires sets with the same element type, got set of %s and set of %s",
					leftSetType.ElementType, rightSetType.ElementType)
				return nil
			}
			return leftSetType
		}
		if leftIsSet || rightIsSet {
			v.ctx.AddError("operator - cannot mix set and non-set types")
			return nil
		}

		// Numeric subtraction
		if !v.isNumeric(leftType) || !v.isNumeric(rightType) {
			v.ctx.AddError("operator - requires numeric or set types, got %s and %s",
				leftType, rightType)
			return nil
		}
		// Result type is the "wider" of the two operands
		if v.isFloat(leftType) || v.isFloat(rightType) {
			return types.FLOAT
		}
		return types.INTEGER

	case "*":
		// Task 6.1.2.2: * can be used for set intersection or numeric multiplication
		// Check for set intersection first
		leftSetType, leftIsSet := leftType.(*types.SetType)
		rightSetType, rightIsSet := rightType.(*types.SetType)
		if leftIsSet && rightIsSet {
			// Set intersection - both operands must have the same element type
			if !leftSetType.ElementType.Equals(rightSetType.ElementType) {
				v.ctx.AddError("set intersection requires sets with the same element type, got set of %s and set of %s",
					leftSetType.ElementType, rightSetType.ElementType)
				return nil
			}
			return leftSetType
		}
		if leftIsSet || rightIsSet {
			v.ctx.AddError("operator * cannot mix set and non-set types")
			return nil
		}

		// Numeric multiplication
		if !v.isNumeric(leftType) || !v.isNumeric(rightType) {
			v.ctx.AddError("operator * requires numeric or set types, got %s and %s",
				leftType, rightType)
			return nil
		}
		// Result type is the "wider" of the two operands
		if v.isFloat(leftType) || v.isFloat(rightType) {
			return types.FLOAT
		}
		return types.INTEGER

	case "/":
		// Arithmetic operators require numeric types
		if !v.isNumeric(leftType) || !v.isNumeric(rightType) {
			v.ctx.AddError("operator / requires numeric types, got %s and %s",
				leftType, rightType)
			return nil
		}
		// Result type is the "wider" of the two operands
		if v.isFloat(leftType) || v.isFloat(rightType) {
			return types.FLOAT
		}
		return types.INTEGER

	case "div", "mod":
		// Integer division and modulo require integer types
		if !v.isInteger(leftType) || !v.isInteger(rightType) {
			v.ctx.AddError("operator %s requires integer types, got %s and %s",
				expr.Operator, leftType, rightType)
			return nil
		}
		return types.INTEGER

	case "=", "<>", "<", "<=", ">", ">=":
		// Comparison operators return boolean
		// Allow numeric types to be compared with each other (Integer vs Float)
		if v.isNumeric(leftType) && v.isNumeric(rightType) {
			return types.BOOLEAN
		}
		// For non-numeric types, require compatibility
		if !v.typesCompatible(leftType, rightType) {
			v.ctx.AddError("cannot compare %s with %s", leftType, rightType)
		}
		return types.BOOLEAN

	case "and", "or", "xor":
		// Logical operators require boolean
		if !v.isBoolean(leftType) || !v.isBoolean(rightType) {
			v.ctx.AddError("logical operator %s requires boolean operands, got %s and %s",
				expr.Operator, leftType, rightType)
			return nil
		}
		return types.BOOLEAN

	case "in":
		// Task 6.1.2.2: Set membership operator
		// Left operand should be an ordinal value
		// Right operand should be a set type
		// The element type of the set should match the type of the left operand
		rightSetType, rightIsSet := rightType.(*types.SetType)
		if !rightIsSet {
			v.ctx.AddError("right operand of 'in' must be a set type, got %s", rightType.String())
			return nil
		}

		// Left operand must be an ordinal type
		if !types.IsOrdinalType(leftType) {
			v.ctx.AddError("left operand of 'in' must be an ordinal type, got %s", leftType.String())
			return nil
		}

		// The element type of the set should match the type of the left operand
		if !leftType.Equals(rightSetType.ElementType) {
			v.ctx.AddError("type mismatch in 'in' operator: expected %s, got %s",
				rightSetType.ElementType.String(), leftType.String())
			return nil
		}

		return types.BOOLEAN

	default:
		v.ctx.AddError("unknown binary operator: %s", expr.Operator)
		return nil
	}
}

// checkUnaryExpression checks a unary expression
func (v *statementValidator) checkUnaryExpression(expr *ast.UnaryExpression) types.Type {
	operandType := v.checkExpression(expr.Right)
	if operandType == nil {
		return nil
	}

	switch expr.Operator {
	case "-", "+":
		if !v.isNumeric(operandType) {
			v.ctx.AddError("unary %s requires numeric type, got %s", expr.Operator, operandType)
			return nil
		}
		return operandType
	case "not":
		if !v.isBoolean(operandType) {
			v.ctx.AddError("not operator requires boolean, got %s", operandType)
			return nil
		}
		return types.BOOLEAN
	default:
		v.ctx.AddError("unknown unary operator: %s", expr.Operator)
		return nil
	}
}

// checkCallExpression checks a function call
func (v *statementValidator) checkCallExpression(expr *ast.CallExpression) types.Type {
	// Handle member access (method calls like obj.Method())
	if memberAccess, ok := expr.Function.(*ast.MemberAccessExpression); ok {
		// Analyze the member access to get the method type
		methodType := v.checkMemberAccessExpression(memberAccess)
		if methodType == nil {
			return nil
		}

		// Verify it's a function type
		funcType, ok := methodType.(*types.FunctionType)
		if !ok {
			// Could be auto-invoked already, just check arguments
			for _, arg := range expr.Arguments {
				v.checkExpression(arg)
			}
			return methodType
		}

		// Use method name for error messages
		return v.validateFunctionCall(memberAccess.Member.Value, funcType, expr.Arguments)
	}

	// Handle regular function calls (identifier-based)
	funcIdent, ok := expr.Function.(*ast.Identifier)
	if !ok {
		v.ctx.AddError("function call must use identifier or member access")
		return nil
	}

	// Look up function in symbol table
	sym, ok := v.ctx.Symbols.Resolve(funcIdent.Value)
	if !ok {
		// Check if it's a built-in function
		if v.ctx.BuiltinChecker != nil {
			// Task 6.1.2.1: Check if function is a built-in WITHOUT delegating argument analysis
			// The old analyzer doesn't have access to the validation pass's scoped symbol tables,
			// so we need to validate arguments ourselves using checkExpression
			if resultType, isBuiltin := v.ctx.BuiltinChecker.IsBuiltinFunction(funcIdent.Value); isBuiltin {
				// Validate arguments ourselves using the validation pass's scope
				for _, arg := range expr.Arguments {
					v.checkExpression(arg)
				}
				return resultType
			}
		}

		// Check if it's a method in the current class
		if v.ctx.CurrentClass != nil {
			if methodType, found := v.ctx.CurrentClass.GetMethod(funcIdent.Value); found {
				// Found a method - validate it like a function call
				return v.validateFunctionCall(funcIdent.Value, methodType, expr.Arguments)
			}
		}

		v.ctx.AddError("undefined function '%s'", funcIdent.Value)

		// Still check arguments for type errors
		for _, arg := range expr.Arguments {
			v.checkExpression(arg)
		}
		return nil
	}

	// Get function type
	funcType, ok := sym.Type.(*types.FunctionType)
	if !ok {
		// Task 6.1.2.1: Check if it's a function pointer type
		underlyingType := types.GetUnderlyingType(sym.Type)
		if funcPtrType, isFuncPtr := underlyingType.(*types.FunctionPointerType); isFuncPtr {
			return v.validateFunctionPointerCall(funcIdent.Value, funcPtrType, expr.Arguments)
		}
		if methodPtrType, isMethodPtr := underlyingType.(*types.MethodPointerType); isMethodPtr {
			return v.validateFunctionPointerCall(funcIdent.Value, &methodPtrType.FunctionPointerType, expr.Arguments)
		}

		v.ctx.AddError("'%s' is not a function", funcIdent.Value)
		return nil
	}

	return v.validateFunctionCall(funcIdent.Value, funcType, expr.Arguments)
}

// validateFunctionCall validates a function call with the given function type and arguments
func (v *statementValidator) validateFunctionCall(funcName string, funcType *types.FunctionType, args []ast.Expression) types.Type {
	// Validate argument count (handle variadic functions)
	if funcType.IsVariadic {
		// Variadic function: must have at least as many arguments as non-variadic parameters
		nonVariadicCount := len(funcType.Parameters)
		if len(args) < nonVariadicCount {
			v.ctx.AddError("function '%s' expects at least %d argument(s), got %d",
				funcName, nonVariadicCount, len(args))
		}
	} else {
		// Non-variadic function: exact match required
		if len(args) != len(funcType.Parameters) {
			v.ctx.AddError("function '%s' expects %d argument(s), got %d",
				funcName, len(funcType.Parameters), len(args))
		}
	}

	// Validate argument types
	for i, arg := range args {
		var expectedType types.Type

		if i < len(funcType.Parameters) {
			// Non-variadic parameter
			expectedType = funcType.Parameters[i]
		} else if funcType.IsVariadic {
			// Variadic parameter - use VariadicType
			expectedType = funcType.VariadicType
		} else {
			// Too many arguments for non-variadic function (error already reported)
			break
		}

		argType := v.checkExpression(arg)
		if argType != nil && expectedType != nil {
			if !v.typesCompatible(expectedType, argType) {
				v.ctx.AddError("argument %d has type %s, expected %s",
					i+1, argType, expectedType)
			}
		}
	}

	return funcType.ReturnType
}

// validateFunctionPointerCall validates a call expression through a function pointer variable.
// Task 6.1.2.1: Handle function pointer calls in the validation pass
func (v *statementValidator) validateFunctionPointerCall(varName string, funcPtrType *types.FunctionPointerType, args []ast.Expression) types.Type {
	// Validate argument count
	if len(args) != len(funcPtrType.Parameters) {
		v.ctx.AddError("function pointer '%s' expects %d argument(s), got %d",
			varName, len(funcPtrType.Parameters), len(args))
	}

	// Validate argument types
	for i, arg := range args {
		if i >= len(funcPtrType.Parameters) {
			// Too many arguments - error already reported
			v.checkExpression(arg) // Still check the argument for errors
			continue
		}

		expectedType := funcPtrType.Parameters[i]
		argType := v.checkExpression(arg)
		if argType != nil && expectedType != nil {
			if !v.typesCompatible(expectedType, argType) {
				v.ctx.AddError("argument %d has type %s, expected %s",
					i+1, argType, expectedType)
			}
		}
	}

	// Return the function pointer's return type
	if funcPtrType.ReturnType != nil {
		return funcPtrType.ReturnType
	}
	return types.VOID
}

// checkIndexExpression checks an array/string index operation
func (v *statementValidator) checkIndexExpression(expr *ast.IndexExpression) types.Type {
	// TODO: Implement array indexing validation
	v.checkExpression(expr.Left)
	v.checkExpression(expr.Index)
	return nil // Unknown element type
}

// checkMemberAccessExpression checks a member access operation
func (v *statementValidator) checkMemberAccessExpression(expr *ast.MemberAccessExpression) types.Type {
	// Analyze the object expression
	objectType := v.checkExpression(expr.Object)
	if objectType == nil {
		return nil
	}

	// Get member name (case-insensitive)
	memberName := expr.Member.Value
	memberNameLower := ident.Normalize(memberName)

	// Resolve type aliases to get the underlying type
	objectTypeResolved := types.GetUnderlyingType(objectType)

	// Handle record type
	if recordType, ok := objectTypeResolved.(*types.RecordType); ok {
		// Check for class methods (static methods) on record type
		if recordType.HasClassMethod(memberNameLower) {
			classMethod := recordType.GetClassMethod(memberNameLower)
			if classMethod != nil {
				return classMethod
			}
		}

		// Check for instance fields
		fieldType, found := recordType.Fields[memberNameLower]
		if found {
			return fieldType
		}

		v.ctx.AddError("record '%s' has no member '%s'", recordType.Name, memberName)
		return nil
	}

	// Handle interface type
	if ifaceType, ok := objectTypeResolved.(*types.InterfaceType); ok {
		allMethods := types.GetAllInterfaceMethods(ifaceType)
		if methodType, hasMethod := allMethods[memberNameLower]; hasMethod {
			return methodType
		}
		v.ctx.AddError("interface '%s' has no method '%s'", ifaceType.Name, memberName)
		return nil
	}

	// Handle metaclass type (class of T) - convert to base class
	if metaclassType, ok := objectTypeResolved.(*types.ClassOfType); ok {
		if metaclassType.ClassType != nil {
			objectTypeResolved = metaclassType.ClassType
		}
	}

	// Handle class type
	classType, ok := objectTypeResolved.(*types.ClassType)
	if !ok {
		// Handle enum .Value property
		if _, isEnum := objectTypeResolved.(*types.EnumType); isEnum {
			if memberNameLower == "value" {
				return types.INTEGER
			}
			// Continue to check helpers for other properties/methods on enums
		}

		// For non-class/record types, check helpers for properties and methods
		// Prefer helper properties before methods so that property-style access
		// (e.g., i.ToString) resolves correctly when parentheses are omitted
		_, helperProp := v.hasHelperProperty(objectType, memberName)
		if helperProp != nil {
			return helperProp.Type
		}

		_, helperMethod := v.hasHelperMethod(objectType, memberName)
		if helperMethod != nil {
			// Auto-invoke parameterless helper methods when accessed without ()
			// This allows arr.Pop to work the same as arr.Pop()
			if len(helperMethod.Parameters) == 0 {
				// Parameterless method - auto-invoke and return the return type
				return helperMethod.ReturnType
			}
			// Method has parameters - return the method type for deferred invocation
			return helperMethod
		}

		// Check for helper class constants (for scoped enum access like TColor.Red)
		_, helperConst := v.hasHelperClassConst(objectType, memberName)
		if helperConst != nil {
			// For enum types, the constant is the enum value, so return the enum type itself
			if _, isEnum := objectTypeResolved.(*types.EnumType); isEnum {
				return objectType
			}
			// For other types, we'd need to determine the constant's type
			// For now, return the object type (conservative approach)
			return objectType
		}

		v.ctx.AddError("member access on type %s requires a helper, got no helper with member '%s'",
			objectType.String(), memberName)
		return nil
	}

	// Handle built-in TObject properties
	if memberNameLower == "classname" {
		return types.STRING
	}
	if memberNameLower == "classtype" {
		return types.NewClassOfType(classType)
	}

	// Look up field in class (including inherited fields)
	fieldType, found := classType.GetField(memberNameLower)
	if found {
		// Check field visibility
		fieldOwner := v.getFieldOwner(classType, memberNameLower)
		if fieldOwner != nil {
			visibility, hasVisibility := fieldOwner.FieldVisibility[memberNameLower]
			if hasVisibility {
				if !v.checkMemberVisibility(fieldOwner, visibility, "field", memberName) {
					return nil
				}
			}
		}
		return fieldType
	}

	// Look up class variable
	classVarType, foundClassVar := classType.GetClassVar(memberNameLower)
	if foundClassVar {
		// Check class variable visibility
		classVarOwner := v.getClassVarOwner(classType, memberNameLower)
		if classVarOwner != nil {
			visibility, hasVisibility := classVarOwner.ClassVarVisibility[memberNameLower]
			if hasVisibility {
				if !v.checkMemberVisibility(classVarOwner, visibility, "class variable", memberName) {
					return nil
				}
			}
		}
		return classVarType
	}

	// Look up property
	propInfo, propFound := classType.GetProperty(memberNameLower)
	if propFound {
		return propInfo.Type
	}

	// Check for constructors
	constructorOverloads := classType.GetConstructorOverloads(memberNameLower)
	if len(constructorOverloads) > 0 {
		// Check if there's a parameterless constructor
		hasParameterless := false
		for _, ctor := range constructorOverloads {
			if len(ctor.Signature.Parameters) == 0 {
				hasParameterless = true
				break
			}
		}

		// Parameterless constructor - auto-invoke and return class type
		if hasParameterless {
			return classType
		}

		// Constructor with parameters - return method pointer type
		if len(constructorOverloads) == 1 {
			return types.NewMethodPointerType(constructorOverloads[0].Signature.Parameters, classType)
		}
		return types.NewMethodPointerType([]types.Type{}, classType)
	}

	// Look up method in class
	methodType, found := classType.GetMethod(memberNameLower)
	if found {
		// Check method visibility
		methodOwner := v.getMethodOwner(classType, memberNameLower)
		if methodOwner != nil {
			visibility, hasVisibility := methodOwner.MethodVisibility[memberNameLower]
			if hasVisibility {
				if !v.checkMemberVisibility(methodOwner, visibility, "method", memberName) {
					return nil
				}
			}
		}

		// Parameterless methods are auto-invoked
		if len(methodType.Parameters) == 0 {
			if methodType.ReturnType == nil {
				return types.VOID
			}
			return methodType.ReturnType
		}

		// Methods with parameters return method pointer
		return types.NewMethodPointerType(methodType.Parameters, methodType.ReturnType)
	}

	// Check helpers for properties and methods on class types
	_, helperProp := v.hasHelperProperty(objectType, memberName)
	if helperProp != nil {
		return helperProp.Type
	}

	_, helperMethod := v.hasHelperMethod(objectType, memberName)
	if helperMethod != nil {
		// Auto-invoke parameterless helper methods
		if len(helperMethod.Parameters) == 0 {
			return helperMethod.ReturnType
		}
		return helperMethod
	}

	// Check for helper class constants
	_, helperConst := v.hasHelperClassConst(objectType, memberName)
	if helperConst != nil {
		// Return the appropriate type for the constant
		// For now, return the object type (conservative approach)
		return objectType
	}

	// Member not found
	v.ctx.AddError("class '%s' has no member '%s'", classType.Name, memberName)
	return nil
}

// Helper functions

// resolveTypeExpression resolves a type expression to a concrete type
func (v *statementValidator) resolveTypeExpression(typeExpr ast.TypeExpression) types.Type {
	if typeExpr == nil {
		return nil
	}

	switch te := typeExpr.(type) {
	case *ast.TypeAnnotation:
		typ, _ := v.ctx.TypeRegistry.Resolve(te.Name)
		return typ

	case *ast.FunctionPointerTypeNode:
		// Task 6.1.2.1: Handle function pointer type expressions
		// Resolve parameter types
		paramTypes := make([]types.Type, 0, len(te.Parameters))
		for _, param := range te.Parameters {
			if param.Type != nil {
				paramType := v.resolveTypeExpression(param.Type)
				if paramType != nil {
					paramTypes = append(paramTypes, paramType)
				} else {
					// Default to VARIANT for unknown types
					paramTypes = append(paramTypes, types.VARIANT)
				}
			} else {
				paramTypes = append(paramTypes, types.VARIANT)
			}
		}

		// Resolve return type (nil for procedures)
		var returnType types.Type
		if te.ReturnType != nil {
			returnType = v.resolveTypeExpression(te.ReturnType)
		}

		// Create the function pointer type
		if te.OfObject {
			return types.NewMethodPointerType(paramTypes, returnType)
		}
		return types.NewFunctionPointerType(paramTypes, returnType)

	case *ast.SetTypeNode:
		// Task 6.1.2.2: Handle set type expressions
		// Resolve the element type
		if te.ElementType == nil {
			v.ctx.AddError("set type must have an element type")
			return nil
		}

		elementType := v.resolveTypeExpression(te.ElementType)
		if elementType == nil {
			v.ctx.AddError("cannot resolve element type for set")
			return nil
		}

		// Create the set type (NewSetType automatically determines storage kind)
		return types.NewSetType(elementType)

	default:
		// TODO: Handle other type expressions (arrays, etc.)
		return nil
	}
}

// typesCompatible checks if two types are compatible for assignment
func (v *statementValidator) typesCompatible(target, source types.Type) bool {
	if target == nil || source == nil {
		return false
	}

	// Resolve type aliases to get underlying types
	targetUnderlying := types.GetUnderlyingType(target)
	sourceUnderlying := types.GetUnderlyingType(source)

	// Variant can accept any type
	if targetUnderlying.Equals(types.VARIANT) {
		return true
	}

	// Check if types are equal (including underlying types)
	if targetUnderlying.Equals(sourceUnderlying) {
		return true
	}

	// Task 6.1.2.1: Check function pointer type compatibility
	if targetFP, ok := targetUnderlying.(*types.FunctionPointerType); ok {
		// Source must also be a function pointer type
		if sourceFP, ok := sourceUnderlying.(*types.FunctionPointerType); ok {
			return v.functionPointersCompatible(targetFP, sourceFP)
		}
		// Method pointer can be assigned to function pointer (but not vice versa)
		if sourceMP, ok := sourceUnderlying.(*types.MethodPointerType); ok {
			return v.functionPointersCompatible(targetFP, &sourceMP.FunctionPointerType)
		}
	}
	if targetMP, ok := targetUnderlying.(*types.MethodPointerType); ok {
		// Only method pointers can be assigned to method pointers
		if sourceMP, ok := sourceUnderlying.(*types.MethodPointerType); ok {
			return v.functionPointersCompatible(&targetMP.FunctionPointerType, &sourceMP.FunctionPointerType)
		}
		// Function pointer cannot be assigned to method pointer
		return false
	}

	// TODO: Handle subtype relationships (class inheritance), numeric conversions, etc.
	return false
}

// functionPointersCompatible checks if two function pointer types are compatible.
// Task 6.1.2.1: Two function pointers are compatible if:
// - They have the same number of parameters
// - Each parameter type is compatible
// - Return types are compatible (or both are procedures with no return)
func (v *statementValidator) functionPointersCompatible(target, source *types.FunctionPointerType) bool {
	// Check parameter count
	if len(target.Parameters) != len(source.Parameters) {
		return false
	}

	// Check each parameter type
	for i, targetParam := range target.Parameters {
		sourceParam := source.Parameters[i]
		if !targetParam.Equals(sourceParam) {
			return false
		}
	}

	// Check return type
	// Both nil means both are procedures
	if target.ReturnType == nil && source.ReturnType == nil {
		return true
	}
	// One nil and one not means incompatible (function vs procedure)
	if target.ReturnType == nil || source.ReturnType == nil {
		return false
	}
	// Both have return types - check compatibility
	return target.ReturnType.Equals(source.ReturnType)
}

// isNumeric checks if a type is numeric (Integer or Float)
func (v *statementValidator) isNumeric(t types.Type) bool {
	return v.isInteger(t) || v.isFloat(t)
}

// isInteger checks if a type is Integer
func (v *statementValidator) isInteger(t types.Type) bool {
	_, ok := t.(*types.IntegerType)
	return ok
}

// isFloat checks if a type is Float
func (v *statementValidator) isFloat(t types.Type) bool {
	_, ok := t.(*types.FloatType)
	return ok
}

// isBoolean checks if a type is Boolean
func (v *statementValidator) isBoolean(t types.Type) bool {
	_, ok := t.(*types.BooleanType)
	return ok
}

// isString checks if a type is String
func (v *statementValidator) isString(t types.Type) bool {
	_, ok := t.(*types.StringType)
	return ok
}

// ============================================================================
// Additional Expression Validation Methods
// ============================================================================

// checkMethodCallExpression checks a method call expression
func (v *statementValidator) checkMethodCallExpression(expr *ast.MethodCallExpression) types.Type {
	// Check the object expression
	objType := v.checkExpression(expr.Object)
	if objType == nil {
		return nil
	}

	// Check all arguments
	for _, arg := range expr.Arguments {
		v.checkExpression(arg)
	}

	// TODO: Validate method exists and argument types match
	return nil
}

// checkNewExpression checks a class instantiation expression
func (v *statementValidator) checkNewExpression(expr *ast.NewExpression) types.Type {
	// Resolve the class type from the class name
	if expr.ClassName == nil {
		v.ctx.AddError("'new' expression missing class name")
		return nil
	}

	classType, ok := v.ctx.TypeRegistry.Resolve(expr.ClassName.Value)
	if !ok || classType == nil {
		v.ctx.AddError("cannot resolve class '%s' in 'new' expression", expr.ClassName.Value)
		return nil
	}

	// Check constructor arguments
	for _, arg := range expr.Arguments {
		v.checkExpression(arg)
	}

	// TODO: Validate constructor exists and argument types match
	return classType
}

// checkNewArrayExpression checks an array instantiation expression
func (v *statementValidator) checkNewArrayExpression(expr *ast.NewArrayExpression) types.Type {
	// Check dimension expressions
	for _, dim := range expr.Dimensions {
		dimType := v.checkExpression(dim)
		if dimType != nil && !v.isInteger(dimType) {
			v.ctx.AddError("array dimension must be integer, got %s", dimType)
		}
	}

	// TODO: Return proper array type
	return nil
}

// checkArrayLiteral checks an array literal expression
func (v *statementValidator) checkArrayLiteral(expr *ast.ArrayLiteralExpression, expectedType types.Type) types.Type {
	if expr == nil {
		return nil
	}

	// Task 6.1.2.2: Check if expected type is a set type
	// DWScript uses [] for both arrays and sets, so we need to handle both cases
	if expectedType != nil {
		underlyingExpected := types.GetUnderlyingType(expectedType)
		if setType, isSet := underlyingExpected.(*types.SetType); isSet {
			// Convert array literal expression to set literal for validation
			// Create a temporary SetLiteral to reuse the set validation logic
			setLiteral := &ast.SetLiteral{
				Elements: expr.Elements,
			}
			return v.checkSetLiteral(setLiteral, setType)
		}
	}

	var expectedArrayType *types.ArrayType
	if expectedType != nil {
		if arr, ok := types.GetUnderlyingType(expectedType).(*types.ArrayType); ok {
			expectedArrayType = arr
		} else {
			v.ctx.AddError("array literal cannot be assigned to non-array type %s", expectedType.String())
			return nil
		}
	}

	// Empty literal requires explicit context
	if len(expr.Elements) == 0 {
		if expectedArrayType == nil {
			v.ctx.AddError("cannot infer type for empty array literal")
			return nil
		}
		// Allow empty arrays for array of const / array of Variant (Format function)
		return expectedType
	}

	var inferredElementType types.Type
	hasErrors := false

	for idx, elem := range expr.Elements {
		var elementExpected types.Type
		if expectedArrayType != nil {
			elementExpected = expectedArrayType.ElementType
		}

		elemType := v.checkExpressionWithExpectedType(elem, elementExpected)
		if elemType == nil {
			hasErrors = true
			continue
		}

		if expectedArrayType != nil {
			// This enables heterogeneous arrays like ['string', 123, 3.14] for Format()
			elemTypeUnderlying := types.GetUnderlyingType(expectedArrayType.ElementType)
			if elemTypeUnderlying.TypeKind() == "VARIANT" {
				// Accept any element type for array of Variant
				continue
			}

			if !v.typesCompatible(expectedArrayType.ElementType, elemType) {
				v.ctx.AddError("array element %d has type %s, expected %s",
					idx+1, elemType.String(), expectedArrayType.ElementType.String())
				hasErrors = true
			}
			continue
		}

		if inferredElementType == nil {
			inferredElementType = elemType
			continue
		}

		underlyingCurrent := types.GetUnderlyingType(elemType)
		underlyingInferred := types.GetUnderlyingType(inferredElementType)

		if underlyingInferred.Equals(underlyingCurrent) {
			continue
		}

		// If the current element fits in the inferred type, keep the inferred type.
		if v.typesCompatible(inferredElementType, elemType) {
			continue
		}

		// If we can widen the inferred type to the current element, do so.
		if v.typesCompatible(elemType, inferredElementType) {
			inferredElementType = elemType
			continue
		}

		// Attempt numeric promotion (e.g., Integer + Float -> Float)
		if promoted := types.PromoteTypes(underlyingInferred, underlyingCurrent); promoted != nil {
			inferredElementType = promoted
			continue
		}

		v.ctx.AddError("incompatible element types in array literal: %s and %s",
			underlyingInferred.String(), underlyingCurrent.String())
		hasErrors = true
	}

	if hasErrors {
		return nil
	}

	if expectedArrayType != nil {
		return expectedType
	}

	if inferredElementType == nil {
		v.ctx.AddError("unable to infer element type for array literal")
		return nil
	}

	elementUnderlying := types.GetUnderlyingType(inferredElementType)
	arrayType := types.NewDynamicArrayType(elementUnderlying)

	return arrayType
}

// checkRecordLiteral checks a record literal expression
func (v *statementValidator) checkRecordLiteral(expr *ast.RecordLiteralExpression, expectedType types.Type) types.Type {
	if expr == nil {
		return nil
	}

	var recordType *types.RecordType

	// Check if this is a typed record literal (has TypeName)
	if expr.TypeName != nil {
		// Typed record literal: TPoint(x: 10; y: 20)
		// Look up the type by name
		typeName := expr.TypeName.Value
		resolvedType, found := v.ctx.TypeRegistry.Resolve(typeName)
		if !found {
			v.ctx.AddError("unknown record type '%s' in record literal", typeName)
			return nil
		}

		var ok bool
		recordType, ok = resolvedType.(*types.RecordType)
		if !ok {
			v.ctx.AddError("'%s' is not a record type, got %s", typeName, resolvedType.String())
			return nil
		}

		// If expectedType is provided, verify it matches
		if expectedType != nil {
			if expectedRecordType, ok := expectedType.(*types.RecordType); ok {
				if expectedRecordType.Name != recordType.Name {
					v.ctx.AddError("record literal type '%s' does not match expected type '%s'",
						recordType.Name, expectedRecordType.Name)
					return nil
				}
			}
		}
	} else {
		// Anonymous record literal: (x: 10; y: 20)
		// Requires expectedType from context
		if expectedType == nil {
			v.ctx.AddError("anonymous record literal requires type context (use explicit type annotation or typed literal)")
			return nil
		}

		var ok bool
		recordType, ok = expectedType.(*types.RecordType)
		if !ok {
			v.ctx.AddError("record literal requires a record type, got %s", expectedType.String())
			return nil
		}
	}

	// Track which fields have been initialized
	initializedFields := make(map[string]bool)

	// Validate each field in the literal
	for _, field := range expr.Fields {
		// Skip positional fields (not yet implemented)
		if field.Name == nil {
			v.ctx.AddError("positional record field initialization not yet supported")
			continue
		}

		fieldName := field.Name.Value
		// Normalize field name to lowercase for case-insensitive comparison
		lowerFieldName := ident.Normalize(fieldName)

		// Check for duplicate field initialization
		if initializedFields[lowerFieldName] {
			v.ctx.AddError("duplicate field '%s' in record literal", fieldName)
			continue
		}
		initializedFields[lowerFieldName] = true

		// Check if field exists in record type
		expectedFieldType, exists := recordType.Fields[lowerFieldName]
		if !exists {
			v.ctx.AddError("field '%s' does not exist in record type '%s'", fieldName, recordType.Name)
			continue
		}

		// Type-check the field value with expected type context
		actualType := v.checkExpressionWithExpectedType(field.Value, expectedFieldType)
		if actualType == nil {
			continue
		}

		// Check type compatibility
		if !v.typesCompatible(expectedFieldType, actualType) {
			v.ctx.AddError("cannot assign %s to %s in field '%s'",
				actualType.String(), expectedFieldType.String(), fieldName)
		}
	}

	// Check for missing required fields (skip fields with default initializers)
	for fieldName := range recordType.Fields {
		if !initializedFields[fieldName] {
			// Check if the field has a default initializer
			if recordType.FieldsWithInit != nil && recordType.FieldsWithInit[fieldName] {
				// Field has a default initializer, so it's not required in the literal
				continue
			}
			v.ctx.AddError("missing required field '%s' in record literal", fieldName)
		}
	}

	return recordType
}

// checkSetLiteral checks a set literal expression
func (v *statementValidator) checkSetLiteral(expr *ast.SetLiteral, expectedType types.Type) types.Type {
	if expr == nil {
		return nil
	}

	// If we have an expected type, it should be a SetType
	var expectedSetType *types.SetType
	if expectedType != nil {
		var ok bool
		expectedSetType, ok = expectedType.(*types.SetType)
		if !ok {
			v.ctx.AddError("set literal cannot be assigned to non-set type %s", expectedType.String())
			return nil
		}
	}

	// Empty set literal
	if len(expr.Elements) == 0 {
		if expectedSetType != nil {
			return expectedSetType
		}
		// Empty set without context - cannot infer type
		v.ctx.AddError("cannot infer type for empty set literal")
		return nil
	}

	// Analyze all elements and check they are of the same ordinal type
	// Support all ordinal types (Integer, String/Char, Enum, Subrange)
	var elementType types.Type
	for i, elem := range expr.Elements {
		var elemType types.Type

		// Check if this is a range expression (e.g., 1..10 or 'a'..'z')
		if rangeExpr, isRange := elem.(*ast.RangeExpression); isRange {
			// Analyze start and end of range
			startType := v.checkExpression(rangeExpr.Start)
			endType := v.checkExpression(rangeExpr.RangeEnd)

			if startType == nil || endType == nil {
				// Error already reported
				continue
			}

			// Both bounds must be ordinal types
			if !types.IsOrdinalType(startType) {
				v.ctx.AddError("range start must be an ordinal type, got %s", startType.String())
				continue
			}
			if !types.IsOrdinalType(endType) {
				v.ctx.AddError("range end must be an ordinal type, got %s", endType.String())
				continue
			}

			// Both bounds must be the same type
			if !startType.Equals(endType) {
				v.ctx.AddError("range start and end must have the same type: got %s and %s",
					startType.String(), endType.String())
				continue
			}

			elemType = startType
		} else {
			// Regular element (not a range)
			elemType = v.checkExpression(elem)
			if elemType == nil {
				// Error already reported
				continue
			}

			// Element must be an ordinal type
			if !types.IsOrdinalType(elemType) {
				v.ctx.AddError("set element must be an ordinal value, got %s", elemType.String())
				continue
			}
		}

		// First element determines the element type
		if i == 0 {
			elementType = elemType
		} else {
			// All elements must be of the same ordinal type
			if !elemType.Equals(elementType) {
				v.ctx.AddError("type mismatch in set literal: expected %s, got %s",
					elementType.String(), elemType.String())
			}
		}
	}

	if elementType == nil {
		// All elements had errors
		return nil
	}

	// If we have an expected set type, verify the element type matches
	if expectedSetType != nil {
		if !elementType.Equals(expectedSetType.ElementType) {
			v.ctx.AddError("type mismatch in set literal: expected set of %s, got set of %s",
				expectedSetType.ElementType.String(), elementType.String())
			return expectedSetType // Return expected type to continue analysis
		}
		return expectedSetType
	}

	// Create and return a new set type based on inferred element type
	return types.NewSetType(elementType)
}

// checkIsExpression checks an 'is' type checking expression
func (v *statementValidator) checkIsExpression(expr *ast.IsExpression) types.Type {
	// Check left expression
	leftType := v.checkExpression(expr.Left)
	if leftType == nil {
		return nil
	}

	// If checking against boolean value (is True/is False)
	if expr.Right != nil {
		v.checkExpression(expr.Right)
		return types.BOOLEAN
	}

	// Type checking mode - validate target type is a class
	if expr.TargetType != nil {
		targetType := v.resolveTypeExpression(expr.TargetType)
		if targetType == nil {
			v.ctx.AddError("cannot resolve target type in 'is' expression")
			return nil
		}

		// Validate operand is a class or nil
		if leftType != types.NIL {
			leftUnderlying := types.GetUnderlyingType(leftType)
			if _, isClass := leftUnderlying.(*types.ClassType); !isClass {
				v.ctx.AddError("'is' operator requires class instance, got %s", leftType)
			}
		}

		// Validate target is a class type
		targetUnderlying := types.GetUnderlyingType(targetType)
		if _, isClass := targetUnderlying.(*types.ClassType); !isClass {
			v.ctx.AddError("'is' operator requires class type, got %s", targetType)
		}
	}

	return types.BOOLEAN
}

// checkAsExpression checks an 'as' type casting expression
func (v *statementValidator) checkAsExpression(expr *ast.AsExpression) types.Type {
	// Check left expression
	leftType := v.checkExpression(expr.Left)
	if leftType == nil {
		return nil
	}

	// Resolve target type
	targetType := v.resolveTypeExpression(expr.TargetType)
	if targetType == nil {
		v.ctx.AddError("cannot resolve target type in 'as' expression")
		return nil
	}

	// Validate target is class or interface
	targetUnderlying := types.GetUnderlyingType(targetType)
	_, isInterface := targetUnderlying.(*types.InterfaceType)
	_, isClass := targetUnderlying.(*types.ClassType)

	if !isInterface && !isClass {
		v.ctx.AddError("'as' operator requires class or interface type, got %s", targetType)
		return targetType // Return target type to prevent cascading errors
	}

	// Validate left type is a class or nil
	if leftType != types.NIL {
		leftUnderlying := types.GetUnderlyingType(leftType)
		if _, isClass := leftUnderlying.(*types.ClassType); !isClass {
			v.ctx.AddError("'as' operator requires class instance, got %s", leftType)
		}
	}

	// TODO: Validate inheritance/interface implementation relationship
	return targetType
}

// checkImplementsExpression checks an 'implements' interface checking expression
func (v *statementValidator) checkImplementsExpression(expr *ast.ImplementsExpression) types.Type {
	// Check left expression
	leftType := v.checkExpression(expr.Left)
	if leftType == nil {
		return nil
	}

	// Resolve target type
	targetType := v.resolveTypeExpression(expr.TargetType)
	if targetType == nil {
		v.ctx.AddError("cannot resolve target type in 'implements' expression")
		return nil
	}

	// Validate target is an interface
	targetUnderlying := types.GetUnderlyingType(targetType)
	if _, isInterface := targetUnderlying.(*types.InterfaceType); !isInterface {
		v.ctx.AddError("'implements' operator requires interface type, got %s", targetType)
	}

	// Validate left type is a class or nil
	if leftType != types.NIL {
		leftUnderlying := types.GetUnderlyingType(leftType)
		if _, isClass := leftUnderlying.(*types.ClassType); !isClass {
			v.ctx.AddError("'implements' operator requires class instance, got %s", leftType)
		}
	}

	return types.BOOLEAN
}

// checkIfExpression checks an inline if-then-else expression
func (v *statementValidator) checkIfExpression(expr *ast.IfExpression) types.Type {
	// Check condition is boolean
	condType := v.checkExpression(expr.Condition)
	if condType != nil && !v.isBoolean(condType) {
		v.ctx.AddError("if expression condition must be boolean, got %s", condType)
	}

	// Check consequence
	consequenceType := v.checkExpression(expr.Consequence)
	if consequenceType == nil {
		return nil
	}

	// If there's an alternative, check type compatibility
	if expr.Alternative != nil {
		alternativeType := v.checkExpression(expr.Alternative)
		if alternativeType != nil && !v.typesCompatible(consequenceType, alternativeType) {
			v.ctx.AddError("incompatible types in if-then-else: %s and %s",
				consequenceType, alternativeType)
		}
	}

	return consequenceType
}

// checkSelfExpression checks a 'Self' expression
func (v *statementValidator) checkSelfExpression(expr *ast.SelfExpression) types.Type {
	// 'Self' is only valid inside class methods
	if v.ctx.CurrentClass == nil {
		v.ctx.AddError("'Self' can only be used inside class methods")
		return nil
	}

	return v.ctx.CurrentClass
}

// checkInheritedExpression checks an 'inherited' expression
func (v *statementValidator) checkInheritedExpression(expr *ast.InheritedExpression) types.Type {
	// 'inherited' is only valid inside class methods
	if v.ctx.CurrentClass == nil {
		v.ctx.AddError("'inherited' can only be used inside class methods")
		return nil
	}

	// Validate that class has a parent
	if v.ctx.CurrentClass.Parent == nil {
		v.ctx.AddError("'inherited' used in class with no parent")
		return nil
	}

	// TODO: Return appropriate type based on inherited call
	return nil
}

// checkAddressOfExpression checks an address-of (@) expression
func (v *statementValidator) checkAddressOfExpression(expr *ast.AddressOfExpression) types.Type {
	// Check the operator expression (the function/method reference)
	operatorType := v.checkExpression(expr.Operator)
	if operatorType == nil {
		return nil
	}

	// TODO: Return function pointer type
	return nil
}

// checkLambdaExpression checks a lambda expression
func (v *statementValidator) checkLambdaExpression(expr *ast.LambdaExpression, expectedType types.Type) types.Type {
	if expr == nil {
		return nil
	}

	// If expectedType is provided and is a function pointer, use it for validation
	var expectedFuncType *types.FunctionPointerType
	if expectedType != nil {
		underlyingType := types.GetUnderlyingType(expectedType)
		if funcPtrType, ok := underlyingType.(*types.FunctionPointerType); ok {
			expectedFuncType = funcPtrType
		}
	}

	// If we have expected function type, validate parameter count
	if expectedFuncType != nil {
		if len(expr.Parameters) != len(expectedFuncType.Parameters) {
			v.ctx.AddError("lambda has %d parameter(s) but expected function type has %d parameter(s)",
				len(expr.Parameters), len(expectedFuncType.Parameters))
			return nil
		}
	}

	// Check for duplicate parameter names
	paramNames := make(map[string]bool)
	for _, param := range expr.Parameters {
		paramName := ident.Normalize(param.Name.Value)
		if paramNames[paramName] {
			v.ctx.AddError("duplicate parameter name '%s' in lambda", param.Name.Value)
			return nil
		}
		paramNames[paramName] = true
	}

	// Mark that we're in a lambda
	oldInLambda := v.ctx.InLambda
	v.ctx.InLambda = true
	defer func() { v.ctx.InLambda = oldInLambda }()

	// Set CurrentFunction so return statements can validate properly
	// Create a temporary FunctionDecl representing the lambda
	oldCurrentFunction := v.ctx.CurrentFunction
	lambdaFuncDecl := &ast.FunctionDecl{
		Parameters: expr.Parameters,
		ReturnType: expr.ReturnType,
		// Lambda body will be validated below
	}
	v.ctx.CurrentFunction = lambdaFuncDecl
	defer func() { v.ctx.CurrentFunction = oldCurrentFunction }()

	// Push a new function scope for lambda parameters and local variables
	v.ctx.PushScope(ScopeFunction)
	defer v.ctx.PopScope()

	// Add lambda parameters to the current scope
	for _, param := range expr.Parameters {
		if param != nil && param.Name != nil {
			var paramType types.Type = types.VARIANT // default type
			if param.Type != nil {
				resolvedType := v.resolveTypeExpression(param.Type)
				if resolvedType != nil {
					paramType = resolvedType
				}
			}
			// If we have an expected function type, use its parameter types
			if expectedFuncType != nil {
				paramIdx := -1
				for i, p := range expr.Parameters {
					if p == param {
						paramIdx = i
						break
					}
				}
				if paramIdx >= 0 && paramIdx < len(expectedFuncType.Parameters) {
					// Parameters is []Type, so directly use the type
					paramType = expectedFuncType.Parameters[paramIdx]
				}
			}
			v.ctx.DefineInCurrentScope(param.Name.Value, paramType)
		}
	}

	// Validate lambda body
	if expr.Body != nil {
		v.validateStatement(expr.Body)
	}

	// If expected type is available, return it for better type checking
	if expectedFuncType != nil {
		return expectedFuncType
	}

	// Otherwise return nil (full lambda type inference requires symbol table access)
	// This will be handled by the old analyzer until fully migrated
	return nil
}

// checkOldExpression checks an 'old' expression (for postconditions)
func (v *statementValidator) checkOldExpression(expr *ast.OldExpression) types.Type {
	// Check the identifier
	if expr.Identifier != nil {
		return v.checkIdentifier(expr.Identifier)
	}
	return nil
}

// ============================================================================
// Helper Support Methods
// ============================================================================

// getHelpersForType returns all helpers that extend the given type.
func (v *statementValidator) getHelpersForType(typ types.Type) []*types.HelperType {
	if typ == nil {
		return nil
	}

	// Look up helpers by the type's string representation (case-insensitive)
	typeName := ident.Normalize(typ.String())
	helpers := v.ctx.Helpers[typeName]

	// For array types, also include generic array helpers
	if _, isArray := typ.(*types.ArrayType); isArray {
		arrayHelpers := v.ctx.Helpers["array"]
		if arrayHelpers != nil {
			// Combine type-specific helpers with generic array helpers
			helpers = append(helpers, arrayHelpers...)
		}
	}

	// For enum types, also include generic enum helpers
	if _, isEnum := typ.(*types.EnumType); isEnum {
		enumHelpers := v.ctx.Helpers["enum"]
		if enumHelpers != nil {
			// Combine type-specific helpers with generic enum helpers
			helpers = append(helpers, enumHelpers...)
		}
	}

	return helpers
}

// hasHelperMethod checks if any helper for the given type defines the specified method.
// Returns the helper type and method if found.
func (v *statementValidator) hasHelperMethod(typ types.Type, methodName string) (*types.HelperType, *types.FunctionType) {
	helpers := v.getHelpersForType(typ)
	if helpers == nil {
		return nil, nil
	}

	// Check each helper in reverse order so user-defined helpers (added later)
	// take precedence over built-in helpers registered during initialization.
	methodNameLower := ident.Normalize(methodName)
	for idx := len(helpers) - 1; idx >= 0; idx-- {
		helper := helpers[idx]
		if method, ok := helper.Methods[methodNameLower]; ok {
			// For array types, specialize the method signature if needed
			// (e.g., Pop() should return the array's element type, not VARIANT)
			if arrayType, isArray := typ.(*types.ArrayType); isArray {
				// Check if this is the Pop method that needs specialization
				if methodNameLower == "pop" && method.ReturnType == types.VARIANT {
					// Create a specialized version with the actual element type
					specialized := types.NewFunctionType(method.Parameters, arrayType.ElementType)
					specialized.ParamNames = method.ParamNames
					specialized.DefaultValues = method.DefaultValues
					specialized.VarParams = method.VarParams
					specialized.ConstParams = method.ConstParams
					specialized.LazyParams = method.LazyParams
					specialized.IsVariadic = method.IsVariadic
					specialized.VariadicType = method.VariadicType
					return helper, specialized
				}
			}
			return helper, method
		}
	}

	return nil, nil
}

// hasHelperProperty checks if any helper for the given type defines the specified property.
// Returns the helper type and property if found.
func (v *statementValidator) hasHelperProperty(typ types.Type, propName string) (*types.HelperType, *types.PropertyInfo) {
	helpers := v.getHelpersForType(typ)
	if helpers == nil {
		return nil, nil
	}

	// Check each helper in reverse order (most recent first)
	propNameLower := ident.Normalize(propName)
	for idx := len(helpers) - 1; idx >= 0; idx-- {
		helper := helpers[idx]
		if prop, ok := helper.Properties[propNameLower]; ok {
			return helper, prop
		}
	}

	return nil, nil
}

// hasHelperClassConst checks if any helper for the given type defines the specified class constant.
// Returns the helper type and constant value if found.
func (v *statementValidator) hasHelperClassConst(typ types.Type, constName string) (*types.HelperType, interface{}) {
	helpers := v.getHelpersForType(typ)
	if helpers == nil {
		return nil, nil
	}

	// Check each helper in reverse order (most recent first)
	constNameLower := ident.Normalize(constName)
	for idx := len(helpers) - 1; idx >= 0; idx-- {
		helper := helpers[idx]
		if constVal, ok := helper.ClassConsts[constNameLower]; ok {
			return helper, constVal
		}
	}

	return nil, nil
}

// ============================================================================
// Visibility Checking Methods
// ============================================================================

// checkMemberVisibility checks if a class member is accessible from the current context.
// Returns true if the member is accessible, false otherwise (and adds an error).
func (v *statementValidator) checkMemberVisibility(
	definingClass *types.ClassType,
	visibility int,
	memberKind string, // "field", "method", "property", "class variable"
	memberName string,
) bool {
	// Import visibility constants from AST package
	const (
		visibilityPrivate   = 0 // ast.VisibilityPrivate
		visibilityProtected = 1 // ast.VisibilityProtected
		visibilityPublic    = 2 // ast.VisibilityPublic
	)

	// Public members are always accessible
	if visibility == visibilityPublic {
		return true
	}

	// If no current class context, only public members are accessible
	if v.ctx.CurrentClass == nil {
		v.ctx.AddError("cannot access %s %s '%s' of class %s from outside a class",
			visibilityString(visibility), memberKind, memberName, definingClass.Name)
		return false
	}

	// Private members: only accessible from the same class
	if visibility == visibilityPrivate {
		if v.ctx.CurrentClass == definingClass {
			return true
		}
		v.ctx.AddError("cannot access private %s '%s' of class %s from class %s",
			memberKind, memberName, definingClass.Name, v.ctx.CurrentClass.Name)
		return false
	}

	// Protected members: accessible from the same class or descendant classes
	if visibility == visibilityProtected {
		// Check if current class is the defining class or inherits from it
		if v.ctx.CurrentClass == definingClass || v.isDescendantOf(v.ctx.CurrentClass, definingClass) {
			return true
		}
		v.ctx.AddError("cannot access protected %s '%s' of class %s from class %s",
			memberKind, memberName, definingClass.Name, v.ctx.CurrentClass.Name)
		return false
	}

	// Unknown visibility level - be conservative and allow access
	return true
}

// isDescendantOf checks if childClass inherits from parentClass
func (v *statementValidator) isDescendantOf(childClass, parentClass *types.ClassType) bool {
	if childClass == nil || parentClass == nil {
		return false
	}

	// Walk up the inheritance chain
	current := childClass.Parent
	for current != nil {
		if current == parentClass {
			return true
		}
		current = current.Parent
	}

	return false
}

// getFieldOwner returns the class that declares a field, walking up the inheritance chain
func (v *statementValidator) getFieldOwner(class *types.ClassType, fieldName string) *types.ClassType {
	if class == nil {
		return nil
	}

	// Check if this class declares the field (case-insensitive)
	lowerFieldName := ident.Normalize(fieldName)
	if _, found := class.Fields[lowerFieldName]; found {
		return class
	}

	// Check parent classes
	return v.getFieldOwner(class.Parent, fieldName)
}

// getMethodOwner returns the class that declares a method, walking up the inheritance chain
func (v *statementValidator) getMethodOwner(class *types.ClassType, methodName string) *types.ClassType {
	if class == nil {
		return nil
	}

	// Use overload system
	methodKey := ident.Normalize(methodName)
	if _, found := class.MethodOverloads[methodKey]; found {
		return class
	}

	// Check parent classes
	return v.getMethodOwner(class.Parent, methodName)
}

// getClassVarOwner returns the class that declares a class variable, walking up the inheritance chain
func (v *statementValidator) getClassVarOwner(class *types.ClassType, classVarName string) *types.ClassType {
	if class == nil {
		return nil
	}

	// Check if this class declares the class variable (case-insensitive)
	lowerClassVarName := ident.Normalize(classVarName)
	if _, found := class.ClassVars[lowerClassVarName]; found {
		return class
	}

	// Check parent classes
	return v.getClassVarOwner(class.Parent, classVarName)
}

// visibilityString returns a human-readable string for a visibility level
func visibilityString(visibility int) string {
	switch visibility {
	case 0: // VisibilityPrivate
		return "private"
	case 1: // VisibilityProtected
		return "protected"
	case 2: // VisibilityPublic
		return "public"
	default:
		return "unknown"
	}
}

// ============================================================================
// Type Checking Helper Methods
// ============================================================================

// isNumericType checks if a type is numeric (Integer or Float)
func (v *statementValidator) isNumericType(t types.Type) bool {
	return v.isIntegerType(t) || v.isFloatType(t)
}

// isIntegerType checks if a type is Integer
func (v *statementValidator) isIntegerType(t types.Type) bool {
	if t == nil {
		return false
	}
	underlying := types.GetUnderlyingType(t)
	return underlying == types.INTEGER || underlying.TypeKind() == "INTEGER"
}

// isFloatType checks if a type is Float
func (v *statementValidator) isFloatType(t types.Type) bool {
	if t == nil {
		return false
	}
	underlying := types.GetUnderlyingType(t)
	return underlying == types.FLOAT || underlying.TypeKind() == "FLOAT"
}

// isStringType checks if a type is String
func (v *statementValidator) isStringType(t types.Type) bool {
	if t == nil {
		return false
	}
	underlying := types.GetUnderlyingType(t)
	return underlying == types.STRING || underlying.TypeKind() == "STRING"
}
