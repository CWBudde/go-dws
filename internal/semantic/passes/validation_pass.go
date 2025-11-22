package passes

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
//   var x: Integer;
//   var y: String;
//   x := y; // ERROR: Cannot assign String to Integer
//
//   class TFoo = class
//     procedure Bar; virtual; abstract;
//   end;
//
//   class TBaz = class(TFoo)
//   end; // ERROR: TBaz must implement abstract method Bar
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
	case *ast.AssignmentStatement:
		v.validateAssignment(s)
	case *ast.ReturnStatement:
		v.validateReturn(s)
	case *ast.BreakStatement:
		v.validateBreak(s)
	case *ast.ContinueStatement:
		v.validateContinue(s)
	case *ast.ExpressionStatement:
		// Type-check the expression
		v.checkExpression(s.Expression)
	case *ast.IfStatement:
		v.validateIf(s)
	case *ast.WhileStatement:
		v.validateWhile(s)
	case *ast.ForStatement:
		v.validateFor(s)
	case *ast.BlockStatement:
		v.validateBlock(s)
	case *ast.FunctionDecl:
		v.validateFunction(s)
	// Class/interface/enum/record declarations don't need validation here
	// (they were handled in Pass 1 and Pass 2)
	case *ast.ClassDecl, *ast.InterfaceDecl, *ast.EnumDecl, *ast.RecordDecl:
		// Skip - structural validation done in earlier passes
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
	case *ast.Identifier:
		return v.checkIdentifier(e)
	case *ast.BinaryExpression:
		return v.checkBinaryExpression(e)
	case *ast.UnaryExpression:
		return v.checkUnaryExpression(e)
	case *ast.CallExpression:
		return v.checkCallExpression(e)
	case *ast.IndexExpression:
		return v.checkIndexExpression(e)
	case *ast.MemberAccessExpression:
		return v.checkMemberAccessExpression(e)
	default:
		// Unknown expression type
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
