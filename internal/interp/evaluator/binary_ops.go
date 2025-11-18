package evaluator

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// This file contains binary operator evaluation methods for the Evaluator.
// Task 3.5.19: Migrated from Interpreter expressions_binary.go

// ============================================================================
// Short-Circuit Operators
// ============================================================================

// evalCoalesceOp evaluates the coalesce operator (??) with short-circuit evaluation.
// Returns the left operand if it's "truthy", otherwise evaluates and returns the right operand.
func (e *Evaluator) evalCoalesceOp(node *ast.BinaryExpression, ctx *ExecutionContext) Value {
	// Evaluate left operand
	left := e.Eval(node.Left, ctx)
	if isError(left) {
		return left
	}
	if left == nil {
		return e.newError(node.Left, "left operand evaluated to nil")
	}

	// Check if left is "falsey" (default/zero value for its type)
	if !isFalsey(left) {
		// Left is truthy, return it (don't evaluate right)
		return left
	}

	// Left is falsey, evaluate and return right operand
	right := e.Eval(node.Right, ctx)
	if isError(right) {
		return right
	}
	if right == nil {
		return e.newError(node.Right, "right operand evaluated to nil")
	}

	return right
}

// evalAndOp evaluates the 'and' operator with short-circuit evaluation.
// For booleans: left and right (short-circuit if left is false).
// For integers: bitwise AND.
func (e *Evaluator) evalAndOp(node *ast.BinaryExpression, ctx *ExecutionContext) Value {
	// Evaluate left operand
	left := e.Eval(node.Left, ctx)
	if isError(left) {
		return left
	}
	if left == nil {
		return e.newError(node.Left, "left operand evaluated to nil")
	}

	// For integers, 'and' is bitwise AND (not short-circuit)
	if left.Type() == "INTEGER" {
		// Need to evaluate right operand for bitwise operation
		right := e.Eval(node.Right, ctx)
		if isError(right) {
			return right
		}
		if right == nil {
			return e.newError(node.Right, "right operand evaluated to nil")
		}
		return e.evalIntegerBinaryOp("and", left, right, node)
	}

	// For booleans, short-circuit evaluation
	if left.Type() == "BOOLEAN" {
		leftBool, ok := left.(*runtime.BooleanValue)
		if !ok {
			return e.newError(node.Left, "expected boolean for 'and' operator")
		}

		// Short-circuit: if left is false, return false without evaluating right
		if !leftBool.Value {
			return &runtime.BooleanValue{Value: false}
		}

		// Left is true, evaluate right
		right := e.Eval(node.Right, ctx)
		if isError(right) {
			return right
		}
		if right == nil {
			return e.newError(node.Right, "right operand evaluated to nil")
		}

		rightBool, ok := right.(*runtime.BooleanValue)
		if !ok {
			return e.newError(node.Right, "expected boolean for 'and' operator")
		}

		return &runtime.BooleanValue{Value: rightBool.Value}
	}

	// Handle Variant types
	if left.Type() == "VARIANT" {
		// Unwrap and try again
		left = unwrapVariant(left)
		if left.Type() == "BOOLEAN" {
			leftBool, ok := left.(*runtime.BooleanValue)
			if ok && !leftBool.Value {
				// Short-circuit
				return &runtime.BooleanValue{Value: false}
			}
		}
		// Fall through to evaluate right operand
		right := e.Eval(node.Right, ctx)
		if isError(right) {
			return right
		}
		if right == nil {
			return e.newError(node.Right, "right operand evaluated to nil")
		}
		return e.evalVariantBinaryOp("and", left, right, node)
	}

	return e.newError(node, "type mismatch: 'and' operator requires boolean or integer operands")
}

// evalOrOp evaluates the 'or' operator with short-circuit evaluation.
// For booleans: left or right (short-circuit if left is true).
// For integers: bitwise OR.
func (e *Evaluator) evalOrOp(node *ast.BinaryExpression, ctx *ExecutionContext) Value {
	// Evaluate left operand
	left := e.Eval(node.Left, ctx)
	if isError(left) {
		return left
	}
	if left == nil {
		return e.newError(node.Left, "left operand evaluated to nil")
	}

	// For integers, 'or' is bitwise OR (not short-circuit)
	if left.Type() == "INTEGER" {
		// Need to evaluate right operand for bitwise operation
		right := e.Eval(node.Right, ctx)
		if isError(right) {
			return right
		}
		if right == nil {
			return e.newError(node.Right, "right operand evaluated to nil")
		}
		return e.evalIntegerBinaryOp("or", left, right, node)
	}

	// For booleans, short-circuit evaluation
	if left.Type() == "BOOLEAN" {
		leftBool, ok := left.(*runtime.BooleanValue)
		if !ok {
			return e.newError(node.Left, "expected boolean for 'or' operator")
		}

		// Short-circuit: if left is true, return true without evaluating right
		if leftBool.Value {
			return &runtime.BooleanValue{Value: true}
		}

		// Left is false, evaluate right
		right := e.Eval(node.Right, ctx)
		if isError(right) {
			return right
		}
		if right == nil {
			return e.newError(node.Right, "right operand evaluated to nil")
		}

		rightBool, ok := right.(*runtime.BooleanValue)
		if !ok {
			return e.newError(node.Right, "expected boolean for 'or' operator")
		}

		return &runtime.BooleanValue{Value: rightBool.Value}
	}

	// Handle Variant types
	if left.Type() == "VARIANT" {
		// Unwrap and try again
		left = unwrapVariant(left)
		if left.Type() == "BOOLEAN" {
			leftBool, ok := left.(*runtime.BooleanValue)
			if ok && leftBool.Value {
				// Short-circuit
				return &runtime.BooleanValue{Value: true}
			}
		}
		// Fall through to evaluate right operand
		right := e.Eval(node.Right, ctx)
		if isError(right) {
			return right
		}
		if right == nil {
			return e.newError(node.Right, "right operand evaluated to nil")
		}
		return e.evalVariantBinaryOp("or", left, right, node)
	}

	return e.newError(node, "type mismatch: 'or' operator requires boolean or integer operands")
}

// ============================================================================
// Type-Specific Binary Operations
// ============================================================================

// evalIntegerBinaryOp evaluates binary operations on integers.
func (e *Evaluator) evalIntegerBinaryOp(op string, left, right Value, node ast.Node) Value {
	// Unwrap Variant values before processing
	left = unwrapVariant(left)
	right = unwrapVariant(right)

	// Safe type assertions with error handling
	leftInt, ok := left.(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "expected integer, got %s", left.Type())
	}
	rightInt, ok := right.(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "expected integer, got %s", right.Type())
	}

	leftVal := leftInt.Value
	rightVal := rightInt.Value

	switch op {
	case "+":
		return &runtime.IntegerValue{Value: leftVal + rightVal}
	case "-":
		return &runtime.IntegerValue{Value: leftVal - rightVal}
	case "*":
		return &runtime.IntegerValue{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return e.newError(node, "division by zero: %d / %d", leftVal, rightVal)
		}
		// Integer division in DWScript uses / for float division
		return &runtime.FloatValue{Value: float64(leftVal) / float64(rightVal)}
	case "div":
		if rightVal == 0 {
			return e.newError(node, "division by zero: %d div %d", leftVal, rightVal)
		}
		return &runtime.IntegerValue{Value: leftVal / rightVal}
	case "mod":
		if rightVal == 0 {
			return e.newError(node, "modulo by zero: %d mod %d", leftVal, rightVal)
		}
		return &runtime.IntegerValue{Value: leftVal % rightVal}
	case "shl":
		if rightVal < 0 {
			return e.newError(node, "negative shift amount")
		}
		return &runtime.IntegerValue{Value: leftVal << uint(rightVal)}
	case "shr":
		if rightVal < 0 {
			return e.newError(node, "negative shift amount")
		}
		return &runtime.IntegerValue{Value: leftVal >> uint(rightVal)}
	case "sar":
		if rightVal < 0 {
			return e.newError(node, "negative shift amount")
		}
		// Arithmetic shift right (sign-preserving)
		return &runtime.IntegerValue{Value: leftVal >> uint(rightVal)}
	case "and":
		// Bitwise AND for integers
		return &runtime.IntegerValue{Value: leftVal & rightVal}
	case "or":
		// Bitwise OR for integers
		return &runtime.IntegerValue{Value: leftVal | rightVal}
	case "xor":
		// Bitwise XOR for integers
		return &runtime.IntegerValue{Value: leftVal ^ rightVal}
	case "=":
		return &runtime.BooleanValue{Value: leftVal == rightVal}
	case "<>":
		return &runtime.BooleanValue{Value: leftVal != rightVal}
	case "<":
		return &runtime.BooleanValue{Value: leftVal < rightVal}
	case ">":
		return &runtime.BooleanValue{Value: leftVal > rightVal}
	case "<=":
		return &runtime.BooleanValue{Value: leftVal <= rightVal}
	case ">=":
		return &runtime.BooleanValue{Value: leftVal >= rightVal}
	default:
		return e.newError(node, "unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

// evalFloatBinaryOp evaluates binary operations on floats.
// Handles mixed integer/float operations by converting to float.
func (e *Evaluator) evalFloatBinaryOp(op string, left, right Value, node ast.Node) Value {
	var leftVal, rightVal float64

	// Unwrap Variant values before processing
	left = unwrapVariant(left)
	right = unwrapVariant(right)

	// Convert left operand to float
	switch v := left.(type) {
	case *runtime.FloatValue:
		leftVal = v.Value
	case *runtime.IntegerValue:
		leftVal = float64(v.Value)
	default:
		return e.newError(node, "type error in float operation: expected FLOAT or INTEGER, got %s", left.Type())
	}

	// Convert right operand to float
	switch v := right.(type) {
	case *runtime.FloatValue:
		rightVal = v.Value
	case *runtime.IntegerValue:
		rightVal = float64(v.Value)
	default:
		return e.newError(node, "type error in float operation: expected FLOAT or INTEGER, got %s", right.Type())
	}

	switch op {
	case "+":
		return &runtime.FloatValue{Value: leftVal + rightVal}
	case "-":
		return &runtime.FloatValue{Value: leftVal - rightVal}
	case "*":
		return &runtime.FloatValue{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return e.newError(node, "division by zero: %v / %v", leftVal, rightVal)
		}
		return &runtime.FloatValue{Value: leftVal / rightVal}
	case "=":
		return &runtime.BooleanValue{Value: leftVal == rightVal}
	case "<>":
		return &runtime.BooleanValue{Value: leftVal != rightVal}
	case "<":
		return &runtime.BooleanValue{Value: leftVal < rightVal}
	case ">":
		return &runtime.BooleanValue{Value: leftVal > rightVal}
	case "<=":
		return &runtime.BooleanValue{Value: leftVal <= rightVal}
	case ">=":
		return &runtime.BooleanValue{Value: leftVal >= rightVal}
	default:
		return e.newError(node, "unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

// evalStringBinaryOp evaluates binary operations on strings.
func (e *Evaluator) evalStringBinaryOp(op string, left, right Value, node ast.Node) Value {
	// Unwrap Variant values before processing
	left = unwrapVariant(left)
	right = unwrapVariant(right)

	// Safe type assertions with error handling
	leftStr, ok := left.(*runtime.StringValue)
	if !ok {
		return e.newError(node, "expected string, got %s", left.Type())
	}
	rightStr, ok := right.(*runtime.StringValue)
	if !ok {
		return e.newError(node, "expected string, got %s", right.Type())
	}

	leftVal := leftStr.Value
	rightVal := rightStr.Value

	switch op {
	case "+":
		return &runtime.StringValue{Value: leftVal + rightVal}
	case "=":
		return &runtime.BooleanValue{Value: leftVal == rightVal}
	case "<>":
		return &runtime.BooleanValue{Value: leftVal != rightVal}
	case "<":
		return &runtime.BooleanValue{Value: leftVal < rightVal}
	case ">":
		return &runtime.BooleanValue{Value: leftVal > rightVal}
	case "<=":
		return &runtime.BooleanValue{Value: leftVal <= rightVal}
	case ">=":
		return &runtime.BooleanValue{Value: leftVal >= rightVal}
	default:
		return e.newError(node, "unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

// evalBooleanBinaryOp evaluates binary operations on booleans.
func (e *Evaluator) evalBooleanBinaryOp(op string, left, right Value, node ast.Node) Value {
	// Unwrap Variant values before processing
	left = unwrapVariant(left)
	right = unwrapVariant(right)

	// Safe type assertions with error handling
	leftBool, ok := left.(*runtime.BooleanValue)
	if !ok {
		return e.newError(node, "expected boolean, got %s", left.Type())
	}
	rightBool, ok := right.(*runtime.BooleanValue)
	if !ok {
		return e.newError(node, "expected boolean, got %s", right.Type())
	}

	leftVal := leftBool.Value
	rightVal := rightBool.Value

	switch op {
	case "and":
		return &runtime.BooleanValue{Value: leftVal && rightVal}
	case "or":
		return &runtime.BooleanValue{Value: leftVal || rightVal}
	case "xor":
		return &runtime.BooleanValue{Value: leftVal != rightVal}
	case "=":
		return &runtime.BooleanValue{Value: leftVal == rightVal}
	case "<>":
		return &runtime.BooleanValue{Value: leftVal != rightVal}
	default:
		return e.newError(node, "unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

// evalEnumBinaryOp evaluates binary operations on enum values.
// Enums are compared by their ordinal values.
// Note: Enum types may not be migrated to runtime package yet.
func (e *Evaluator) evalEnumBinaryOp(op string, left, right Value, node ast.Node) Value {
	// Delegate to adapter for now since EnumValue is in interp package
	// Full migration requires EnumValue to be in runtime package
	return e.adapter.EvalNode(node)
}

// ============================================================================
// Complex Type Comparisons
// ============================================================================

// evalEqualityComparison handles = and <> operators for complex types.
// Supports: nil, objects, interfaces, classes, RTTI, sets, arrays, records.
func (e *Evaluator) evalEqualityComparison(op string, left, right Value, node ast.Node) Value {
	// This handles object/interface/class/RTTI/set/array/record comparisons
	// Delegate to adapter for now since these types are still in interp package
	// Full migration requires all these value types in runtime package
	return e.adapter.EvalNode(node)
}

// ============================================================================
// Operator Overloading and Special Operators
// ============================================================================

// tryBinaryOperator attempts to use custom operator overloading.
// Returns (result, true) if operator found, or (nil, false) if not found.
func (e *Evaluator) tryBinaryOperator(operator string, left, right Value, node ast.Node) (Value, bool) {
	// Operator overloading requires access to:
	// - ObjectInstance.Class.lookupOperator() for instance operators
	// - globalOperators registry for global operators
	// These are in interp package and haven't been migrated yet
	// Delegate to adapter for now
	return nil, false
}

// evalInOperator evaluates the 'in' operator for membership testing.
// Supports: arrays, sets, strings, subranges.
func (e *Evaluator) evalInOperator(value, container Value, node ast.Node) Value {
	// The 'in' operator is complex and handles:
	// - Array membership: x in arrayVar
	// - Set membership: x in setVar
	// - String substring: 'ab' in 'abc'
	// - Subrange checking: 5 in [1..10]
	// Delegate to adapter for now since these types are still in interp package
	return e.adapter.EvalNode(node)
}

// evalVariantBinaryOp handles binary operations with Variant operands.
// Variants require complex type coercion and unwrapping.
func (e *Evaluator) evalVariantBinaryOp(op string, left, right Value, node ast.Node) Value {
	// Variant operations are extremely complex:
	// - Unwrap Variant values
	// - Determine underlying types
	// - Apply type coercion rules
	// - Perform operation on unwrapped values
	// - Special handling for nil/unassigned variants
	// Delegate to adapter for now since VariantValue is in interp package
	return e.adapter.EvalNode(node)
}
