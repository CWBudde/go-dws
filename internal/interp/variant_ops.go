package interp

import (
	"github.com/cwbudde/go-dws/pkg/ast"
)

// variant_ops.go contains Variant-specific binary operations.
// Extracted from expressions_binary.go as part of Phase 3.8.4.2.

// evalVariantBinaryOp evaluates binary operations on Variant values.
//
// Variant operations follow these rules:
//   - Unwrap operands to get actual runtime values
//   - Apply numeric promotion (Integer + Float → Float)
//   - Support string concatenation with + operator
//   - Raise runtime error if types are incompatible
func (i *Interpreter) evalVariantBinaryOp(op string, left, right Value, node ast.Node) Value {
	// Task 9.4.3: Check if either operand is an uninitialized Variant (before unwrapping)
	// An uninitialized variant is a VariantValue with Value == nil.
	// After unwrapping via unwrapVariant(), this nil becomes an UnassignedValue object.
	// This is distinct from a VariantValue explicitly containing an UnassignedValue/NullValue/NilValue.
	leftUnassignedVariant := false
	rightUnassignedVariant := false
	if leftVar, ok := left.(*VariantValue); ok && leftVar.Value == nil {
		leftUnassignedVariant = true
	}
	if rightVar, ok := right.(*VariantValue); ok && rightVar.Value == nil {
		rightUnassignedVariant = true
	}

	// Unwrap Variant values to get the actual runtime values
	leftVal := unwrapVariant(left)
	rightVal := unwrapVariant(right)

	// Task 9.4.1: Check for Null/Unassigned/Nil values (after unwrapping)
	_, leftIsNil := leftVal.(*NilValue)
	_, rightIsNil := rightVal.(*NilValue)
	_, leftIsNull := leftVal.(*NullValue)
	_, rightIsNull := rightVal.(*NullValue)
	_, leftIsUnassigned := leftVal.(*UnassignedValue)
	_, rightIsUnassigned := rightVal.(*UnassignedValue)

	leftIsNullish := leftIsNil || leftIsNull || leftIsUnassigned
	rightIsNullish := rightIsNil || rightIsNull || rightIsUnassigned

	// For comparison operators
	// Task 9.4.3: Complex comparison semantics for Null/Unassigned variants:
	// - Uninitialized variant (VariantValue with Value==nil): equals falsey values (0, false, '', etc.)
	//   and also equals other nullish values (Unassigned, Null, Nil).
	// - Explicit Unassigned/Null/Nil value: only equals other nullish values (Unassigned, Null, Nil),
	//   does NOT equal falsey values.
	//   Example: var v: Variant; (uninitialized) will equal 0, but var v: Variant := Unassigned; (explicitly assigned) will NOT equal 0.
	if op == "=" || op == "<>" {
		// Case 1: Both are nullish (Null/nil/Unassigned) or unassigned variants -> equal
		if (leftIsNullish || leftUnassignedVariant) && (rightIsNullish || rightUnassignedVariant) {
			return &BooleanValue{Value: op == "="}
		}

		// Case 2: One is an UNASSIGNED variant (not just nullish), check if other is falsey
		// Only unassigned variants (not Null/nil) equal falsey values
		if leftUnassignedVariant && !rightIsNullish {
			result := isFalsey(rightVal)
			if op == "=" {
				return &BooleanValue{Value: result}
			}
			return &BooleanValue{Value: !result}
		}
		if rightUnassignedVariant && !leftIsNullish {
			result := isFalsey(leftVal)
			if op == "=" {
				return &BooleanValue{Value: result}
			}
			return &BooleanValue{Value: !result}
		}

		// Case 3: One is nullish (but not unassigned variant), the other is not -> not equal
		if leftIsNullish || rightIsNullish {
			return &BooleanValue{Value: op == "<>"}
		}
	}

	// Error if either operand is nullish for non-comparison operators
	if leftIsNullish {
		return i.newErrorWithLocation(node, "cannot perform operation on unassigned Variant")
	}
	if rightIsNullish {
		return i.newErrorWithLocation(node, "cannot perform operation on unassigned Variant")
	}

	leftType := leftVal.Type()
	rightType := rightVal.Type()

	// Dispatch based on unwrapped types
	switch {
	// Both integers
	case leftType == "INTEGER" && rightType == "INTEGER":
		return i.evalIntegerBinaryOp(op, leftVal, rightVal)

	// Either is float → promote to float
	case leftType == "FLOAT" || rightType == "FLOAT":
		return i.evalFloatBinaryOp(op, leftVal, rightVal)

	// Both strings
	case leftType == "STRING" && rightType == "STRING":
		return i.evalStringBinaryOp(op, leftVal, rightVal)

	// Both booleans
	case leftType == "BOOLEAN" && rightType == "BOOLEAN":
		return i.evalBooleanBinaryOp(op, leftVal, rightVal)

	// String + any type → string concatenation (for + operator only)
	case op == "+" && (leftType == "STRING" || rightType == "STRING"):
		leftStr := i.convertToString(leftVal)
		rightStr := i.convertToString(rightVal)
		return &StringValue{Value: leftStr + rightStr}

	// Numeric type mismatch → try conversion
	case isNumericType(leftType) && isNumericType(rightType):
		// This shouldn't happen since we handle Integer and Float above,
		// but included for completeness
		return i.evalFloatBinaryOp(op, leftVal, rightVal)

	// For comparison operators, try comparing as strings
	case (op == "=" || op == "<>" || op == "<" || op == ">" || op == "<=" || op == ">="):
		// Convert both to strings and compare
		leftStr := i.convertToString(leftVal)
		rightStr := i.convertToString(rightVal)
		return i.evalStringBinaryOp(op, &StringValue{Value: leftStr}, &StringValue{Value: rightStr})

	// For boolean operators with mixed numeric/boolean types, coerce to boolean
	case (op == "and" || op == "or" || op == "xor") &&
		((leftType == "BOOLEAN" && (rightType == "INTEGER" || rightType == "FLOAT")) ||
			(rightType == "BOOLEAN" && (leftType == "INTEGER" || leftType == "FLOAT")) ||
			((leftType == "INTEGER" || leftType == "FLOAT") && (rightType == "INTEGER" || rightType == "FLOAT"))):
		// Coerce both operands to boolean
		leftBool := variantToBool(leftVal)
		rightBool := variantToBool(rightVal)
		result := i.evalBooleanBinaryOp(op, &BooleanValue{Value: leftBool}, &BooleanValue{Value: rightBool})
		// Wrap result in Variant since at least one operand was a Variant
		return BoxVariant(result)

	default:
		return i.newErrorWithLocation(node, "incompatible Variant types for operator %s: %s and %s",
			op, leftType, rightType)
	}
}

// convertToString converts a Value to its string representation.
// Used for Variant string concatenation and comparison.
func (i *Interpreter) convertToString(val Value) string {
	if val == nil {
		return ""
	}
	return val.String()
}

// isFalsey checks if a value is considered "falsey" (default/zero value for its type).
// Falsey values: 0 (integer), 0.0 (float), "" (empty string), false (boolean), nil, empty arrays.
//
// NOTE: This is a legacy version kept for compatibility with variant_ops.go.
// New code should use evaluator.IsFalsey() instead.
func isFalsey(val Value) bool {
	// Task 9.4.4: Handle nil (from unassigned variants)
	if val == nil {
		return true
	}

	switch v := val.(type) {
	case *IntegerValue:
		return v.Value == 0
	case *FloatValue:
		return v.Value == 0.0
	case *StringValue:
		return v.Value == ""
	case *BooleanValue:
		return !v.Value
	case *NilValue:
		return true
	case *NullValue:
		// Task 9.4.1: Null is always falsey
		return true
	case *UnassignedValue:
		// Task 9.4.1: Unassigned is always falsey
		return true
	case *ArrayValue:
		return len(v.Elements) == 0
	case *VariantValue:
		// Task 9.4.4: Unwrap variant and check inner value
		// If variant.Value is nil, the nil check above will return true
		return isFalsey(v.Value)
	default:
		// Other types (objects, classes, etc.) are truthy if non-nil
		return false
	}
}

// isNumericType checks if a type is numeric (INTEGER or FLOAT).
func isNumericType(typeStr string) bool {
	return typeStr == "INTEGER" || typeStr == "FLOAT"
}
