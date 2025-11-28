package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// This file contains visitor methods for operator expression AST nodes.
// Operators include binary operators (+, -, *, /, etc.) and unary operators (-, not, @).

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

	case left.Type() == "SET" && right.Type() == "SET":
		return e.evalSetBinaryOp(node.Operator, left, right, node)

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

// VisitAddressOfExpression evaluates an address-of expression (@funcName or @obj.method).
// Creates function/method pointers that can be called later or assigned to variables.
//
// Task 3.5.37: Full migration of address-of operator evaluation.
//
// **SYNTAX FORMS**:
//   - `@FunctionName` - Creates a function pointer to a standalone function
//   - `@object.MethodName` - Creates a method pointer bound to an object instance
//
// **FUNCTION POINTERS**:
//   - Regular function/procedure references
//   - Resolved via function registry (case-insensitive lookup)
//   - For overloaded functions, the first overload is used
//   - The function pointer captures the closure environment
//
// **METHOD POINTERS** (procedure/function of object):
//   - Method references bound to a specific object instance
//   - The object is evaluated and stored in the pointer's SelfObject field
//   - When called, Self is bound to the captured object
//   - Enables callback patterns: `var callback: TNotifyEvent := @myObj.OnClick;`
//
// **TYPE INFORMATION**:
//   - Function pointers include type information (parameter types, return type)
//   - Method pointers are marked as "of object" type
//   - Used for type checking during assignment and calls
func (e *Evaluator) VisitAddressOfExpression(node *ast.AddressOfExpression, ctx *ExecutionContext) Value {
	// The operator should be an identifier (function/procedure name) or member access (for methods)
	switch operand := node.Operator.(type) {
	case *ast.Identifier:
		// Regular function/procedure pointer: @FunctionName
		// Task 3.5.122: Direct function lookup and pointer creation without adapter
		funcNameLower := ident.Normalize(operand.Value)
		overloads := e.FunctionRegistry().Lookup(funcNameLower)
		if len(overloads) == 0 {
			return e.newError(node, "undefined function or procedure: %s", operand.Value)
		}

		// For overloaded functions, use the first overload
		// Note: Function pointers cannot represent overload sets, only single functions
		function := overloads[0]

		// Build the function pointer type and create the value
		pointerType := buildFunctionPointerType(function)
		return &runtime.FunctionPointerValue{
			Function:    function,
			Closure:     ctx.Env(),
			PointerType: pointerType,
		}

	case *ast.MemberAccessExpression:
		// Method pointer: @object.MethodName
		// First evaluate the object
		objectVal := e.Eval(operand.Object, ctx)
		if isError(objectVal) {
			return objectVal
		}

		// Get the method name
		methodName := operand.Member.Value

		// Task 3.5.123: Use ObjectValue.CreateMethodPointer with callback pattern
		if objVal, ok := objectVal.(ObjectValue); ok {
			if methodPtr, created := objVal.CreateMethodPointer(methodName, func(methodDecl any) Value {
				return e.adapter.CreateBoundMethodPointer(objectVal, methodDecl)
			}); created {
				return methodPtr
			}
			// Method not found
			return e.newError(node, "undefined method: %s.%s", objVal.ClassName(), methodName)
		}

		// Non-object type - cannot create method pointer
		return e.newError(node, "method pointer requires an object instance, got %s", objectVal.Type())

	default:
		return e.newError(node, "address-of operator requires function or method name, got %T", operand)
	}
}
