package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
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
		if varType == nil {
			v.ctx.AddError("undefined type in variable declaration")
			return
		}
	}

	// If there's an initializer, check type compatibility
	if stmt.Value != nil {
		valueType := v.checkExpression(stmt.Value)
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

	// Check value (right-hand side)
	valueType := v.checkExpression(stmt.Value)
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

// validateReturn validates a return statement
func (v *statementValidator) validateReturn(stmt *ast.ReturnStatement) {
	// Check if we're in a function
	if v.ctx.CurrentFunction == nil {
		v.ctx.AddError("return statement outside of function")
		return
	}

	// Type-check return value
	if stmt.ReturnValue != nil {
		returnType := v.checkExpression(stmt.ReturnValue)
		// TODO: Compare with function's return type
		_ = returnType
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

	// Validate function body
	if decl.Body != nil {
		v.validateStatement(decl.Body)
	}

	// TODO: Validate that all code paths return a value (if function, not procedure)
}

// validateConstDecl validates a constant declaration
func (v *statementValidator) validateConstDecl(stmt *ast.ConstDecl) {
	// Type-check the initializer
	if stmt.Value != nil {
		valueType := v.checkExpression(stmt.Value)

		// If type is specified, validate compatibility
		if stmt.Type != nil {
			declaredType := v.resolveTypeExpression(stmt.Type)
			if declaredType != nil && valueType != nil {
				if !v.typesCompatible(declaredType, valueType) {
					v.ctx.AddError("cannot initialize constant of type %s with value of type %s",
						declaredType, valueType)
				}
			}
		}
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
		// TODO: Walk the inheritance chain and check for abstract methods
		_ = classType
	}
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
		return v.checkArrayLiteral(e)
	case *ast.RecordLiteralExpression:
		return v.checkRecordLiteral(e)
	case *ast.SetLiteral:
		return v.checkSetLiteral(e)
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
		return v.checkLambdaExpression(e)
	case *ast.OldExpression:
		return v.checkOldExpression(e)
	default:
		// Unknown expression type - log for debugging but don't error
		// Some expression types may not need validation
		return nil
	}
}

// checkIdentifier checks an identifier reference
func (v *statementValidator) checkIdentifier(expr *ast.Identifier) types.Type {
	symbol, ok := v.ctx.Symbols.Resolve(expr.Value)
	if !ok {
		v.ctx.AddError("undefined variable '%s'", expr.Value)
		return nil
	}
	return symbol.Type
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
	case "+", "-", "*", "/":
		// Arithmetic operators require numeric types
		if !v.isNumeric(leftType) || !v.isNumeric(rightType) {
			v.ctx.AddError("operator %s requires numeric types, got %s and %s",
				expr.Operator, leftType, rightType)
			return nil
		}
		// Result type is the "wider" of the two operands
		if v.isFloat(leftType) || v.isFloat(rightType) {
			return types.FLOAT
		}
		return types.INTEGER

	case "=", "<>", "<", "<=", ">", ">=":
		// Comparison operators return boolean
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
	// TODO: Implement full function call validation
	// For now, just type-check arguments
	for _, arg := range expr.Arguments {
		v.checkExpression(arg)
	}
	return nil // Unknown return type
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
	// TODO: Implement member access validation
	v.checkExpression(expr.Object)
	return nil // Unknown member type
}

// Helper functions

// resolveTypeExpression resolves a type expression to a concrete type
func (v *statementValidator) resolveTypeExpression(typeExpr ast.TypeExpression) types.Type {
	if typeAnnot, ok := typeExpr.(*ast.TypeAnnotation); ok {
		typ, _ := v.ctx.TypeRegistry.Resolve(typeAnnot.Name)
		return typ
	}
	// TODO: Handle other type expressions (arrays, function pointers, etc.)
	return nil
}

// typesCompatible checks if two types are compatible for assignment
func (v *statementValidator) typesCompatible(target, source types.Type) bool {
	if target == nil || source == nil {
		return false
	}
	return target.Equals(source)
	// TODO: Handle subtype relationships, conversions, etc.
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
func (v *statementValidator) checkArrayLiteral(expr *ast.ArrayLiteralExpression) types.Type {
	// Check all elements
	for _, elem := range expr.Elements {
		v.checkExpression(elem)
	}

	// TODO: Infer common element type and return array type
	return nil
}

// checkRecordLiteral checks a record literal expression
func (v *statementValidator) checkRecordLiteral(expr *ast.RecordLiteralExpression) types.Type {
	// Check field values
	for _, field := range expr.Fields {
		v.checkExpression(field.Value)
	}

	// TODO: Validate record type and field compatibility
	return nil
}

// checkSetLiteral checks a set literal expression
func (v *statementValidator) checkSetLiteral(expr *ast.SetLiteral) types.Type {
	// Check all elements
	for _, elem := range expr.Elements {
		v.checkExpression(elem)
	}

	// TODO: Infer set element type and return set type
	return nil
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
func (v *statementValidator) checkLambdaExpression(expr *ast.LambdaExpression) types.Type {
	// Mark that we're in a lambda
	oldInLambda := v.ctx.InLambda
	v.ctx.InLambda = true
	defer func() { v.ctx.InLambda = oldInLambda }()

	// Validate lambda body
	if expr.Body != nil {
		v.validateStatement(expr.Body)
	}

	// TODO: Return function pointer type
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
