package interp

import (
	"fmt"
	"math"
)

// expressions_binary.go contains legacy binary operation helpers.
// These are still used by variant_ops.go for type-specific operations.

// evalIntegerBinaryOp evaluates binary operations on integers.
func (i *Interpreter) evalIntegerBinaryOp(op string, left, right Value) Value {
	left = unwrapVariant(left)
	right = unwrapVariant(right)

	// Safe type assertions with error handling
	leftInt, ok := left.(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "expected integer, got %s", left.Type())
	}
	rightInt, ok := right.(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "expected integer, got %s", right.Type())
	}

	leftVal := leftInt.Value
	rightVal := rightInt.Value

	switch op {
	case "+":
		return &IntegerValue{Value: leftVal + rightVal}
	case "-":
		return &IntegerValue{Value: leftVal - rightVal}
	case "*":
		return &IntegerValue{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			// Enhanced error with operand values
			return i.NewRuntimeError(
				i.evaluatorInstance.CurrentNode(),
				"division_by_zero",
				"Division by zero",
				map[string]string{
					"left":  fmt.Sprintf("%d", leftVal),
					"right": fmt.Sprintf("%d", rightVal),
				},
			)
		}
		// Integer division in DWScript uses / for float division
		// We'll convert to float for division
		return &FloatValue{Value: float64(leftVal) / float64(rightVal)}
	case "div":
		if rightVal == 0 {
			// Enhanced error with operand values
			return i.NewRuntimeError(
				i.evaluatorInstance.CurrentNode(),
				"division_by_zero",
				"Division by zero",
				map[string]string{
					"left":  fmt.Sprintf("%d", leftVal),
					"right": fmt.Sprintf("%d", rightVal),
				},
			)
		}
		return &IntegerValue{Value: leftVal / rightVal}
	case "mod":
		if rightVal == 0 {
			// Enhanced error with operand values
			return i.NewRuntimeError(
				i.evaluatorInstance.CurrentNode(),
				"modulo_by_zero",
				"Division by zero",
				map[string]string{
					"left":  fmt.Sprintf("%d", leftVal),
					"right": fmt.Sprintf("%d", rightVal),
				},
			)
		}
		return &IntegerValue{Value: leftVal % rightVal}
	case "shl":
		if rightVal < 0 {
			return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "negative shift amount")
		}
		// Shift left - multiply by 2^rightVal
		return &IntegerValue{Value: leftVal << uint(rightVal)}
	case "shr":
		if rightVal < 0 {
			return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "negative shift amount")
		}
		// Shift right - divide by 2^rightVal (logical shift)
		return &IntegerValue{Value: leftVal >> uint(rightVal)}
	case "sar":
		if rightVal < 0 {
			return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "negative shift amount")
		}
		// Arithmetic shift right - sign-preserving shift
		// In Go, >> on signed integers is already arithmetic (sign-preserving)
		return &IntegerValue{Value: leftVal >> uint(rightVal)}
	case "and":
		// Bitwise AND for integers
		return &IntegerValue{Value: leftVal & rightVal}
	case "or":
		// Bitwise OR for integers
		return &IntegerValue{Value: leftVal | rightVal}
	case "xor":
		// Bitwise XOR for integers
		return &IntegerValue{Value: leftVal ^ rightVal}
	case "=":
		return &BooleanValue{Value: leftVal == rightVal}
	case "<>":
		return &BooleanValue{Value: leftVal != rightVal}
	case "<":
		return &BooleanValue{Value: leftVal < rightVal}
	case ">":
		return &BooleanValue{Value: leftVal > rightVal}
	case "<=":
		return &BooleanValue{Value: leftVal <= rightVal}
	case ">=":
		return &BooleanValue{Value: leftVal >= rightVal}
	default:
		return i.newTypeError(i.evaluatorInstance.CurrentNode(), "unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

// evalFloatBinaryOp evaluates binary operations on floats.
// Handles mixed integer/float operations by converting to float.
func (i *Interpreter) evalFloatBinaryOp(op string, left, right Value) Value {
	var leftVal, rightVal float64

	// Unwrap Variant values before processing
	left = unwrapVariant(left)
	right = unwrapVariant(right)

	// Convert left operand to float
	switch v := left.(type) {
	case *FloatValue:
		leftVal = v.Value
	case *IntegerValue:
		leftVal = float64(v.Value)
	default:
		return i.newTypeError(i.evaluatorInstance.CurrentNode(), "type error in float operation: expected FLOAT or INTEGER, got %s", left.Type())
	}

	// Convert right operand to float
	switch v := right.(type) {
	case *FloatValue:
		rightVal = v.Value
	case *IntegerValue:
		rightVal = float64(v.Value)
	default:
		return i.newTypeError(i.evaluatorInstance.CurrentNode(), "type error in float operation: expected FLOAT or INTEGER, got %s", right.Type())
	}

	switch op {
	case "+":
		return &FloatValue{Value: leftVal + rightVal}
	case "-":
		return &FloatValue{Value: leftVal - rightVal}
	case "*":
		return &FloatValue{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			// DWScript returns infinity (or NaN when 0/0) for float division by zero
			if leftVal == 0 {
				return &FloatValue{Value: math.NaN()}
			}
			sign := 1.0
			if leftVal < 0 {
				sign = -1.0
			}
			return &FloatValue{Value: math.Inf(int(sign))}
		}
		return &FloatValue{Value: leftVal / rightVal}
	case "=":
		return &BooleanValue{Value: leftVal == rightVal}
	case "<>":
		return &BooleanValue{Value: leftVal != rightVal}
	case "<":
		return &BooleanValue{Value: leftVal < rightVal}
	case ">":
		return &BooleanValue{Value: leftVal > rightVal}
	case "<=":
		return &BooleanValue{Value: leftVal <= rightVal}
	case ">=":
		return &BooleanValue{Value: leftVal >= rightVal}
	default:
		return i.newTypeError(i.evaluatorInstance.CurrentNode(), "unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

// evalStringBinaryOp evaluates binary operations on strings.
func (i *Interpreter) evalStringBinaryOp(op string, left, right Value) Value {
	// Unwrap Variant values before processing
	left = unwrapVariant(left)
	right = unwrapVariant(right)

	// Safe type assertions with error handling
	leftStr, ok := left.(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "expected string, got %s", left.Type())
	}
	rightStr, ok := right.(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "expected string, got %s", right.Type())
	}

	leftVal := leftStr.Value
	rightVal := rightStr.Value

	switch op {
	case "+":
		return &StringValue{Value: leftVal + rightVal}
	case "=":
		return &BooleanValue{Value: leftVal == rightVal}
	case "<>":
		return &BooleanValue{Value: leftVal != rightVal}
	case "<":
		return &BooleanValue{Value: leftVal < rightVal}
	case ">":
		return &BooleanValue{Value: leftVal > rightVal}
	case "<=":
		return &BooleanValue{Value: leftVal <= rightVal}
	case ">=":
		return &BooleanValue{Value: leftVal >= rightVal}
	default:
		return i.newTypeError(i.evaluatorInstance.CurrentNode(), "unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

// evalBooleanBinaryOp evaluates binary operations on booleans.
func (i *Interpreter) evalBooleanBinaryOp(op string, left, right Value) Value {
	// Unwrap Variant values before processing
	left = unwrapVariant(left)
	right = unwrapVariant(right)

	// Safe type assertions with error handling
	leftBool, ok := left.(*BooleanValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "expected boolean, got %s", left.Type())
	}
	rightBool, ok := right.(*BooleanValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "expected boolean, got %s", right.Type())
	}

	leftVal := leftBool.Value
	rightVal := rightBool.Value

	switch op {
	case "and":
		return &BooleanValue{Value: leftVal && rightVal}
	case "or":
		return &BooleanValue{Value: leftVal || rightVal}
	case "xor":
		return &BooleanValue{Value: leftVal != rightVal}
	case "=":
		return &BooleanValue{Value: leftVal == rightVal}
	case "<>":
		return &BooleanValue{Value: leftVal != rightVal}
	default:
		return i.newTypeError(i.evaluatorInstance.CurrentNode(), "unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}
