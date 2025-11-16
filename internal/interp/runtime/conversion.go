package runtime

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ============================================================================
// Value Conversion Utilities
// ============================================================================
//
// These helpers provide safe, consistent value conversions with proper error
// handling. They reduce code duplication and make conversion logic explicit.
//
// Usage:
//   intVal, err := ToInteger(someValue)
//   floatVal, err := ToFloat(someValue)
//   strVal := ToString(someValue) // never fails
//   boolVal, err := ToBoolean(someValue)
// ============================================================================

// ToInteger converts a value to an integer.
// Handles: IntegerValue (direct), FloatValue (truncate), StringValue (parse), BooleanValue (0/1)
func ToInteger(v Value) (int64, error) {
	if v == nil {
		return 0, NewConversionError(v, "INTEGER", "nil value")
	}

	switch val := v.(type) {
	case *IntegerValue:
		return val.Value, nil
	case *FloatValue:
		// Truncate to integer
		return int64(val.Value), nil
	case *StringValue:
		// Try to parse as integer
		if i, err := strconv.ParseInt(val.Value, 10, 64); err == nil {
			return i, nil
		}
		return 0, NewConversionError(v, "INTEGER", fmt.Sprintf("cannot parse '%s' as integer", val.Value))
	case *BooleanValue:
		if val.Value {
			return 1, nil
		}
		return 0, nil
	case NumericValue:
		// Try using the interface
		if i, ok := val.AsInteger(); ok {
			return i, nil
		}
	}

	return 0, NewConversionError(v, "INTEGER", fmt.Sprintf("type %s cannot be converted", v.Type()))
}

// ToFloat converts a value to a float.
// Handles: FloatValue (direct), IntegerValue (convert), StringValue (parse), BooleanValue (0.0/1.0)
func ToFloat(v Value) (float64, error) {
	if v == nil {
		return 0.0, NewConversionError(v, "FLOAT", "nil value")
	}

	switch val := v.(type) {
	case *FloatValue:
		return val.Value, nil
	case *IntegerValue:
		return float64(val.Value), nil
	case *StringValue:
		// Try to parse as float
		if f, err := strconv.ParseFloat(val.Value, 64); err == nil {
			return f, nil
		}
		return 0.0, NewConversionError(v, "FLOAT", fmt.Sprintf("cannot parse '%s' as float", val.Value))
	case *BooleanValue:
		if val.Value {
			return 1.0, nil
		}
		return 0.0, nil
	case NumericValue:
		// Try using the interface
		if f, ok := val.AsFloat(); ok {
			return f, nil
		}
	}

	return 0.0, NewConversionError(v, "FLOAT", fmt.Sprintf("type %s cannot be converted", v.Type()))
}

// ToString converts any value to a string representation.
// This never fails - all values can be converted to strings.
func ToString(v Value) string {
	if v == nil {
		return "nil"
	}
	return v.String()
}

// ToBoolean converts a value to a boolean.
// Handles: BooleanValue (direct), IntegerValue (0=false, other=true), StringValue (parse)
func ToBoolean(v Value) (bool, error) {
	if v == nil {
		return false, NewConversionError(v, "BOOLEAN", "nil value")
	}

	switch val := v.(type) {
	case *BooleanValue:
		return val.Value, nil
	case *IntegerValue:
		return val.Value != 0, nil
	case *FloatValue:
		return val.Value != 0.0, nil
	case *StringValue:
		// DWScript string to boolean conversion
		lower := strings.ToLower(val.Value)
		switch lower {
		case "true", "yes", "1", "t", "y":
			return true, nil
		case "false", "no", "0", "f", "n", "":
			return false, nil
		default:
			return false, NewConversionError(v, "BOOLEAN", fmt.Sprintf("cannot parse '%s' as boolean", val.Value))
		}
	}

	return false, NewConversionError(v, "BOOLEAN", fmt.Sprintf("type %s cannot be converted", v.Type()))
}

// ============================================================================
// Numeric Conversion Utilities
// ============================================================================

// ToNumeric converts a value to a NumericValue interface.
// Returns the value if it implements NumericValue, or an error otherwise.
func ToNumeric(v Value) (NumericValue, error) {
	if v == nil {
		return nil, NewConversionError(v, "NUMERIC", "nil value")
	}

	if num, ok := v.(NumericValue); ok {
		return num, nil
	}

	return nil, NewConversionError(v, "NUMERIC", fmt.Sprintf("type %s is not numeric", v.Type()))
}

// GetNumericValue extracts a numeric value as int64 or float64.
// Prefers integer if the value represents a whole number.
func GetNumericValue(v Value) (isInt bool, intVal int64, floatVal float64, err error) {
	num, err := ToNumeric(v)
	if err != nil {
		return false, 0, 0.0, err
	}

	// Try integer first
	if i, ok := num.AsInteger(); ok {
		// Check if we should use integer or float
		if f, ok := num.AsFloat(); ok {
			// If float and integer are the same, prefer integer
			if float64(i) == f {
				return true, i, 0.0, nil
			}
			// Otherwise use float
			return false, 0, f, nil
		}
		return true, i, 0.0, nil
	}

	// Fall back to float
	if f, ok := num.AsFloat(); ok {
		return false, 0, f, nil
	}

	return false, 0, 0.0, NewConversionError(v, "NUMERIC", "failed to extract numeric value")
}

// ============================================================================
// Safe Arithmetic Operations
// ============================================================================

// AddNumeric safely adds two numeric values.
// Returns a new value of the appropriate type (integer or float).
func AddNumeric(left, right Value) (Value, error) {
	leftNum, err := ToNumeric(left)
	if err != nil {
		return nil, err
	}
	rightNum, err := ToNumeric(right)
	if err != nil {
		return nil, err
	}

	// If both are actually IntegerValue types, return integer result
	leftIntVal, leftIsIntType := left.(*IntegerValue)
	rightIntVal, rightIsIntType := right.(*IntegerValue)
	if leftIsIntType && rightIsIntType {
		// Check for overflow
		if (rightIntVal.Value > 0 && leftIntVal.Value > math.MaxInt64-rightIntVal.Value) ||
			(rightIntVal.Value < 0 && leftIntVal.Value < math.MinInt64-rightIntVal.Value) {
			return nil, NewArithmeticError("integer overflow in addition")
		}
		return NewInteger(leftIntVal.Value + rightIntVal.Value), nil
	}

	// Otherwise, use float
	leftFloat, _ := leftNum.AsFloat()
	rightFloat, _ := rightNum.AsFloat()
	return NewFloat(leftFloat + rightFloat), nil
}

// SubtractNumeric safely subtracts two numeric values.
func SubtractNumeric(left, right Value) (Value, error) {
	leftNum, err := ToNumeric(left)
	if err != nil {
		return nil, err
	}
	rightNum, err := ToNumeric(right)
	if err != nil {
		return nil, err
	}

	// If both are actually IntegerValue types, return integer result
	leftIntVal, leftIsIntType := left.(*IntegerValue)
	rightIntVal, rightIsIntType := right.(*IntegerValue)
	if leftIsIntType && rightIsIntType {
		// Check for overflow
		if (rightIntVal.Value < 0 && leftIntVal.Value > math.MaxInt64+rightIntVal.Value) ||
			(rightIntVal.Value > 0 && leftIntVal.Value < math.MinInt64+rightIntVal.Value) {
			return nil, NewArithmeticError("integer overflow in subtraction")
		}
		return NewInteger(leftIntVal.Value - rightIntVal.Value), nil
	}

	// Otherwise, use float
	leftFloat, _ := leftNum.AsFloat()
	rightFloat, _ := rightNum.AsFloat()
	return NewFloat(leftFloat - rightFloat), nil
}

// MultiplyNumeric safely multiplies two numeric values.
func MultiplyNumeric(left, right Value) (Value, error) {
	leftNum, err := ToNumeric(left)
	if err != nil {
		return nil, err
	}
	rightNum, err := ToNumeric(right)
	if err != nil {
		return nil, err
	}

	// If both are actually IntegerValue types, return integer result
	leftIntVal, leftIsIntType := left.(*IntegerValue)
	rightIntVal, rightIsIntType := right.(*IntegerValue)
	if leftIsIntType && rightIsIntType {
		// Check for overflow (simplified check)
		if leftIntVal.Value != 0 && rightIntVal.Value != 0 {
			result := leftIntVal.Value * rightIntVal.Value
			if result/leftIntVal.Value != rightIntVal.Value {
				return nil, NewArithmeticError("integer overflow in multiplication")
			}
		}
		return NewInteger(leftIntVal.Value * rightIntVal.Value), nil
	}

	// Otherwise, use float
	leftFloat, _ := leftNum.AsFloat()
	rightFloat, _ := rightNum.AsFloat()
	return NewFloat(leftFloat * rightFloat), nil
}

// DivideNumeric safely divides two numeric values.
// Always returns a float result (like DWScript '/').
func DivideNumeric(left, right Value) (Value, error) {
	leftNum, err := ToNumeric(left)
	if err != nil {
		return nil, err
	}
	rightNum, err := ToNumeric(right)
	if err != nil {
		return nil, err
	}

	leftFloat, _ := leftNum.AsFloat()
	rightFloat, _ := rightNum.AsFloat()

	if rightFloat == 0.0 {
		return nil, NewArithmeticError("division by zero")
	}

	return NewFloat(leftFloat / rightFloat), nil
}

// IntDivideNumeric performs integer division (like DWScript 'div').
// Always returns an integer result.
func IntDivideNumeric(left, right Value) (Value, error) {
	leftInt, err := ToInteger(left)
	if err != nil {
		return nil, err
	}
	rightInt, err := ToInteger(right)
	if err != nil {
		return nil, err
	}

	if rightInt == 0 {
		return nil, NewArithmeticError("division by zero")
	}

	return NewInteger(leftInt / rightInt), nil
}

// ModNumeric performs modulo operation (like DWScript 'mod').
func ModNumeric(left, right Value) (Value, error) {
	leftInt, err := ToInteger(left)
	if err != nil {
		return nil, err
	}
	rightInt, err := ToInteger(right)
	if err != nil {
		return nil, err
	}

	if rightInt == 0 {
		return nil, NewArithmeticError("modulo by zero")
	}

	return NewInteger(leftInt % rightInt), nil
}

// ============================================================================
// String Conversion Helpers (for compatibility)
// ============================================================================

// IntToStr converts an integer to a string.
func IntToStr(i int64) string {
	return strconv.FormatInt(i, 10)
}

// FloatToStr converts a float to a string.
func FloatToStr(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}

// BoolToStr converts a boolean to a string ("True" or "False").
func BoolToStr(b bool) string {
	if b {
		return "True"
	}
	return "False"
}
