package evaluator

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// This file contains visitor methods for expression AST nodes.
// Visitor pattern implementation for expressions.
//
// Expressions evaluate to values and can be nested (e.g., binary expressions
// contain left and right sub-expressions).

// ErrorValue represents a runtime error.
type ErrorValue struct {
	Message string
}

func (e *ErrorValue) Type() string   { return "ERROR" }
func (e *ErrorValue) String() string { return "ERROR: " + e.Message }

// newError creates a new error value with optional formatting.
func (e *Evaluator) newError(_ ast.Node, format string, args ...interface{}) Value {
	return &ErrorValue{Message: fmt.Sprintf(format, args...)}
}

// isError checks if a value is an error.
func isError(val Value) bool {
	if val != nil {
		return val.Type() == "ERROR"
	}
	return false
}

// VisitIdentifier evaluates an identifier (variable reference).
func (e *Evaluator) VisitIdentifier(node *ast.Identifier, ctx *ExecutionContext) Value {
	// Self keyword refers to current object instance
	if node.Value == "Self" {
		val, ok := ctx.Env().Get("Self")
		if !ok {
			return e.newError(node, "Self used outside method context")
		}
		// Environment stores interface{}, cast to Value
		if selfVal, ok := val.(Value); ok {
			return selfVal
		}
		return e.newError(node, "Self has invalid type")
	}

	// Try to find identifier in current environment (variables, parameters, constants)
	val, ok := e.adapter.GetVariable(node.Value, ctx)
	if ok {
		// Variable found - return immediately for basic primitives
		switch val.(type) {
		case *runtime.IntegerValue, *runtime.FloatValue, *runtime.StringValue, *runtime.BooleanValue, *runtime.NilValue:
			return val
		}

		// For complex value types (ExternalVarValue, LazyThunk, ReferenceValue, arrays, objects, records),
		// delegate to adapter for full processing
		return e.adapter.EvalNode(node)
	}

	// Check if we're in an instance method context (Self is bound)
	// When Self is bound, identifiers can refer to instance fields, class variables,
	// properties, methods (auto-invoked if zero params), or ClassName/ClassType
	if selfRaw, selfOk := ctx.Env().Get("Self"); selfOk {
		if _, ok := selfRaw.(Value); ok {
			return e.adapter.EvalNode(node)
		}
	}

	// Check if we're in a class method context (__CurrentClass__ is bound)
	// Identifiers can refer to ClassName, ClassType, or class variables
	if currentClassRaw, hasCurrentClass := ctx.Env().Get("__CurrentClass__"); hasCurrentClass {
		if _, ok := currentClassRaw.(Value); ok {
			return e.adapter.EvalNode(node)
		}
	}

	// Check if this identifier is a user-defined function name
	// Functions are auto-invoked if they have zero parameters, or converted to function pointers if they have parameters
	funcNameLower := strings.ToLower(node.Value)
	if overloads, exists := e.adapter.LookupFunction(funcNameLower); exists && len(overloads) > 0 {
		return e.adapter.EvalNode(node)
	}

	// Check if this identifier is a class name (metaclass reference)
	if e.adapter.HasClass(node.Value) {
		return e.adapter.EvalNode(node)
	}

	// Final check: delegate to adapter for built-in functions or error if undefined
	return e.adapter.EvalNode(node)
}

// VisitBinaryExpression evaluates a binary expression (e.g., a + b, x == y).
func (e *Evaluator) VisitBinaryExpression(node *ast.BinaryExpression, ctx *ExecutionContext) Value {
	// Handle short-circuit operators first (special evaluation order)
	switch node.Operator {
	case "??":
		return e.evalCoalesceOp(node, ctx)
	case "and":
		return e.evalAndOp(node, ctx)
	case "or":
		return e.evalOrOp(node, ctx)
	}

	// Evaluate both operands for non-short-circuit operators
	left := e.Eval(node.Left, ctx)
	if isError(left) {
		return left
	}
	if left == nil {
		return e.newError(node.Left, "left operand evaluated to nil")
	}

	right := e.Eval(node.Right, ctx)
	if isError(right) {
		return right
	}
	if right == nil {
		return e.newError(node.Right, "right operand evaluated to nil")
	}

	// Try operator overloading first (custom operators for objects)
	if result, ok := e.tryBinaryOperator(node.Operator, left, right, node); ok {
		return result
	}

	// Handle 'in' operator (membership testing)
	if node.Operator == "in" {
		return e.evalInOperator(left, right, node)
	}

	// Handle operations based on operand types
	// Check for Variant FIRST (Variant operations take precedence)
	if left.Type() == "VARIANT" || right.Type() == "VARIANT" {
		return e.evalVariantBinaryOp(node.Operator, left, right, node)
	}

	// Type-specific binary operations
	switch {
	case left.Type() == "INTEGER" && right.Type() == "INTEGER":
		return e.evalIntegerBinaryOp(node.Operator, left, right, node)

	case left.Type() == "FLOAT" || right.Type() == "FLOAT":
		return e.evalFloatBinaryOp(node.Operator, left, right, node)

	case left.Type() == "STRING" && right.Type() == "STRING":
		return e.evalStringBinaryOp(node.Operator, left, right, node)

	// Allow string concatenation with RTTI_TYPEINFO
	case (left.Type() == "STRING" && right.Type() == "RTTI_TYPEINFO") ||
		(left.Type() == "RTTI_TYPEINFO" && right.Type() == "STRING"):
		if node.Operator == "+" {
			return &runtime.StringValue{Value: left.String() + right.String()}
		}
		return e.newError(node, "type mismatch: %s %s %s", left.Type(), node.Operator, right.Type())

	case left.Type() == "BOOLEAN" && right.Type() == "BOOLEAN":
		return e.evalBooleanBinaryOp(node.Operator, left, right, node)

	// Enum comparisons
	case left.Type() == "ENUM" && right.Type() == "ENUM":
		return e.evalEnumBinaryOp(node.Operator, left, right, node)

	// Object, interface, class, and nil comparisons (= and <>)
	case node.Operator == "=" || node.Operator == "<>":
		return e.evalEqualityComparison(node.Operator, left, right, node)

	default:
		return e.newError(node, "type mismatch: %s %s %s", left.Type(), node.Operator, right.Type())
	}
}

// VisitUnaryExpression evaluates a unary expression (e.g., -x, not b).
func (e *Evaluator) VisitUnaryExpression(node *ast.UnaryExpression, ctx *ExecutionContext) Value {
	// Evaluate the operand
	operand := e.Eval(node.Right, ctx)
	if isError(operand) {
		return operand
	}

	// Try operator overloading first (custom operators for objects)
	if result, ok := e.tryUnaryOperator(node.Operator, operand, node); ok {
		return result
	}

	// Handle standard unary operators
	switch node.Operator {
	case "-":
		return e.evalMinusUnaryOp(operand, node)
	case "+":
		return e.evalPlusUnaryOp(operand, node)
	case "not":
		return e.evalNotUnaryOp(operand, node)
	default:
		return e.newError(node, "unknown operator: %s%s", node.Operator, operand.Type())
	}
}

// VisitAddressOfExpression evaluates an address-of expression (@funcName).
// Creates function/method pointers that can be called later or assigned to variables.
func (e *Evaluator) VisitAddressOfExpression(node *ast.AddressOfExpression, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitGroupedExpression evaluates a grouped expression (parenthesized).
func (e *Evaluator) VisitGroupedExpression(node *ast.GroupedExpression, ctx *ExecutionContext) Value {
	// Grouped expressions just evaluate their inner expression
	// Parentheses are only for precedence, they don't change the value
	return e.Eval(node.Expression, ctx)
}

// VisitCallExpression evaluates a function call expression.
func (e *Evaluator) VisitCallExpression(node *ast.CallExpression, ctx *ExecutionContext) Value {
	// Check for function pointer calls (require special handling for lazy and var parameters)
	if funcIdent, ok := node.Function.(*ast.Identifier); ok {
		if val, exists := e.adapter.GetVariable(funcIdent.Value, ctx); exists {
			if val.Type() == "FUNCTION_POINTER" || val.Type() == "LAMBDA" {
				return e.adapter.EvalNode(node)
			}
		}
	}

	// Check for member access calls: obj.Method(), UnitName.Func(), TClass.Create()
	if memberAccess, ok := node.Function.(*ast.MemberAccessExpression); ok {
		objVal := e.Eval(memberAccess.Object, ctx)
		if isError(objVal) {
			return objVal
		}

		// Delegate record, interface, and object method calls to adapter
		if objVal.Type() == "RECORD" || objVal.Type() == "INTERFACE" || objVal.Type() == "OBJECT" {
			return e.adapter.EvalNode(node)
		}

		// Check for unit-qualified or class constructor calls
		if ident, ok := memberAccess.Object.(*ast.Identifier); ok {
			if e.unitRegistry != nil || e.adapter.HasClass(ident.Value) {
				return e.adapter.EvalNode(node)
			}
		}

		return e.newError(node, "cannot call member expression that is not a method or unit-qualified function")
	}

	// Remaining call types require a simple identifier
	funcName, ok := node.Function.(*ast.Identifier)
	if !ok {
		return e.newError(node, "function call requires identifier or qualified name, got %T", node.Function)
	}

	// Check for user-defined functions (with potential overloading)
	funcNameLower := strings.ToLower(funcName.Value)
	if overloads, exists := e.adapter.LookupFunction(funcNameLower); exists && len(overloads) > 0 {
		return e.adapter.EvalNode(node)
	}

	// Check for implicit Self method calls (MethodName() is shorthand for Self.MethodName())
	if selfRaw, ok := ctx.Env().Get("Self"); ok {
		if selfVal, ok := selfRaw.(Value); ok {
			if selfVal.Type() == "OBJECT" || selfVal.Type() == "CLASS" {
				return e.adapter.EvalNode(node)
			}
		}
	}

	// Check for record static method calls
	if recordRaw, ok := ctx.Env().Get("__CurrentRecord__"); ok {
		if recordVal, ok := recordRaw.(Value); ok {
			if recordVal.Type() == "RECORD_TYPE" {
				return e.adapter.EvalNode(node)
			}
		}
	}

	// Check for built-in functions that need var parameter handling (modify arguments in place)
	switch funcNameLower {
	case "inc", "dec", "insert", "decodedate", "decodetime",
		"swap", "divmod", "trystrtoint", "trystrtofloat", "setlength":
		return e.adapter.EvalNode(node)
	case "delete":
		// Only the 3-parameter form needs var parameter handling
		if len(node.Arguments) == 3 {
			return e.adapter.EvalNode(node)
		}
	}

	// Check for external (Go) functions that may need var parameter handling
	if e.externalFunctions != nil {
		return e.adapter.EvalNode(node)
	}

	// Check for Default(TypeName) function which expects unevaluated type identifier
	if funcNameLower == "default" && len(node.Arguments) == 1 {
		return e.adapter.EvalNode(node)
	}

	// Try type cast for single-argument calls: TypeName(expression)
	if len(node.Arguments) == 1 {
		result := e.adapter.EvalNode(node)
		// If type cast succeeded or there's a real error (not "unknown function"), return it
		if result != nil && !isError(result) {
			return result
		}
		if isError(result) {
			if !strings.Contains(result.String(), "unknown function") &&
				!strings.Contains(result.String(), "undefined identifier") {
				return result
			}
		}
	}

	// Try built-in functions - evaluate all arguments first
	args := make([]Value, len(node.Arguments))
	for idx, arg := range node.Arguments {
		val := e.Eval(arg, ctx)
		if isError(val) {
			return val
		}
		args[idx] = val
	}

	// Call built-in function via adapter
	return e.adapter.CallBuiltinFunction(funcName.Value, args)
}

// VisitNewExpression evaluates a 'new' expression (object instantiation).
// Handles class lookup, field initialization, constructor dispatch, and interface wrapping.
func (e *Evaluator) VisitNewExpression(node *ast.NewExpression, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitMemberAccessExpression evaluates member access (obj.field, obj.method).
// Handles static access, unit-qualified access, instance fields, properties, helper methods,
// class metadata, and enum values.
func (e *Evaluator) VisitMemberAccessExpression(node *ast.MemberAccessExpression, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitMethodCallExpression evaluates a method call (obj.Method(args)).
// Handles virtual dispatch, overload resolution, Self binding, and interface unwrapping.
func (e *Evaluator) VisitMethodCallExpression(node *ast.MethodCallExpression, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitInheritedExpression evaluates an 'inherited' expression.
// Calls parent class method with proper context and Self preservation.
func (e *Evaluator) VisitInheritedExpression(node *ast.InheritedExpression, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitSelfExpression evaluates a 'Self' expression.
// Self refers to the current instance (in instance methods) or the current class (in class methods).
func (e *Evaluator) VisitSelfExpression(node *ast.SelfExpression, ctx *ExecutionContext) Value {
	selfVal, exists := ctx.Env().Get("Self")
	if !exists {
		return e.newError(node, "Self used outside method context")
	}

	val, ok := selfVal.(Value)
	if !ok {
		return e.newError(node, "Self has invalid type")
	}

	return val
}

// VisitEnumLiteral evaluates an enum literal (EnumType.Value).
func (e *Evaluator) VisitEnumLiteral(node *ast.EnumLiteral, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(node, "nil enum literal")
	}

	valueName := node.ValueName
	val, ok := ctx.Env().Get(valueName)
	if !ok {
		return e.newError(node, "undefined enum value '%s'", valueName)
	}

	if value, ok := val.(Value); ok {
		return value
	}

	return e.newError(node, "enum value '%s' has invalid type", valueName)
}

// VisitRecordLiteralExpression evaluates a record literal expression.
// Handles typed and anonymous record literals with field initialization and default values.
func (e *Evaluator) VisitRecordLiteralExpression(node *ast.RecordLiteralExpression, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitSetLiteral evaluates a set literal [value1, value2, ...].
// Handles simple elements, ranges, and mixed sets with proper type inference.
func (e *Evaluator) VisitSetLiteral(node *ast.SetLiteral, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitArrayLiteralExpression evaluates an array literal [1, 2, 3].
// Handles type inference, element coercion, and bounds validation for static and dynamic arrays.
func (e *Evaluator) VisitArrayLiteralExpression(node *ast.ArrayLiteralExpression, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(node, "nil array literal")
	}

	// Empty arrays need type annotation
	if len(node.Elements) == 0 {
		return e.adapter.EvalNode(node)
	}

	// Evaluate all element expressions
	elementCount := len(node.Elements)
	evaluatedElements := make([]Value, elementCount)

	for idx, elem := range node.Elements {
		val := e.Eval(elem, ctx)
		if isError(val) {
			return val
		}
		evaluatedElements[idx] = val
	}

	// Determine array element type
	arrayType := e.getArrayElementType(node, evaluatedElements)
	if arrayType == nil {
		return e.newError(node, "cannot infer type of array literal")
	}

	// Validate element type compatibility
	if err := e.coerceArrayElements(evaluatedElements, arrayType.ElementType, node); err != nil {
		return err
	}

	// Validate static array bounds
	if err := e.validateArrayLiteralSize(arrayType, elementCount, node); err != nil {
		return err
	}

	// Delegate final construction to adapter
	return e.adapter.EvalNode(node)
}

// VisitIndexExpression evaluates an index expression array[index].
// Handles array, string, property, and JSON indexing with bounds checking.
func (e *Evaluator) VisitIndexExpression(node *ast.IndexExpression, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(node, "nil index expression")
	}

	if node.Left == nil {
		return e.newError(node, "index expression missing base")
	}

	base := e.Eval(node.Left, ctx)
	if isError(base) {
		return base
	}

	if node.Index == nil {
		return e.newError(node, "index expression missing index")
	}

	index := e.Eval(node.Index, ctx)
	if isError(index) {
		return index
	}

	// Delegate indexing logic to adapter
	return e.adapter.EvalNode(node)
}

// VisitNewArrayExpression evaluates a new array expression.
// Handles dimension evaluation and multi-dimensional array construction with default values.
func (e *Evaluator) VisitNewArrayExpression(node *ast.NewArrayExpression, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(node, "nil new array expression")
	}

	if node.ElementTypeName == nil {
		return e.newError(node, "new array expression missing element type")
	}

	// Evaluate and validate dimensions
	_, err := e.evaluateDimensions(node.Dimensions, ctx, node)
	if err != nil {
		return err
	}

	// Delegate array construction to adapter
	return e.adapter.EvalNode(node)
}

// VisitLambdaExpression evaluates a lambda expression (closure).
// Creates a lambda that captures the current scope.
func (e *Evaluator) VisitLambdaExpression(node *ast.LambdaExpression, ctx *ExecutionContext) Value {
	return e.adapter.CreateLambda(node, ctx.Env())
}

// VisitIsExpression evaluates an 'is' type checking expression.
// Performs runtime type checking with class hierarchy and interface support.
func (e *Evaluator) VisitIsExpression(node *ast.IsExpression, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitAsExpression evaluates an 'as' type casting expression.
// Performs runtime type checking and casts, raising exception on failure.
func (e *Evaluator) VisitAsExpression(node *ast.AsExpression, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitImplementsExpression evaluates an 'implements' interface checking expression.
// Checks if an object's class implements a specified interface.
func (e *Evaluator) VisitImplementsExpression(node *ast.ImplementsExpression, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitIfExpression evaluates an inline if-then-else expression.
func (e *Evaluator) VisitIfExpression(node *ast.IfExpression, ctx *ExecutionContext) Value {
	condition := e.Eval(node.Condition, ctx)
	if isError(condition) {
		return condition
	}

	// Evaluate consequence if condition is truthy
	if IsTruthy(condition) {
		result := e.Eval(node.Consequence, ctx)
		if isError(result) {
			return result
		}
		return result
	}

	// Evaluate alternative if present
	if node.Alternative != nil {
		result := e.Eval(node.Alternative, ctx)
		if isError(result) {
			return result
		}
		return result
	}

	// No else clause - return default value for the consequence type
	var typeAnnot *ast.TypeAnnotation
	if e.semanticInfo != nil {
		typeAnnot = e.semanticInfo.GetType(node)
	}
	if typeAnnot == nil {
		return e.newError(node, "if expression missing type annotation")
	}

	// Return default value based on type name
	typeName := strings.ToLower(typeAnnot.Name)
	switch typeName {
	case "integer", "int64":
		return &runtime.IntegerValue{Value: 0}
	case "float", "float64", "double", "real":
		return &runtime.FloatValue{Value: 0.0}
	case "string":
		return &runtime.StringValue{Value: ""}
	case "boolean", "bool":
		return &runtime.BooleanValue{Value: false}
	default:
		return &runtime.NilValue{}
	}
}

// VisitOldExpression evaluates an 'old' expression (used in postconditions).
func (e *Evaluator) VisitOldExpression(node *ast.OldExpression, ctx *ExecutionContext) Value {
	identName := node.Identifier.Value
	oldValue, found := ctx.GetOldValue(identName)
	if !found {
		return e.newError(node, "old value for '%s' not captured (internal error)", identName)
	}
	return oldValue.(Value)
}
