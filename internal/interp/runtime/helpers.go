package runtime

import (
	"strings"
)

// ============================================================================
// Type Checking Utilities
// ============================================================================
//
// These helpers provide safe type assertions and checks, reducing the need
// for repeated type assertion code throughout the interpreter.
//
// Usage:
//   if IsInteger(val) { ... }
//   if intVal := AsInteger(val); intVal != nil { ... }
//   if IsNil(val) { ... }
// ============================================================================

// IsInteger checks if a value is an IntegerValue.
func IsInteger(v Value) bool {
	_, ok := v.(*IntegerValue)
	return ok
}

// AsInteger returns the value as *IntegerValue if possible, nil otherwise.
func AsInteger(v Value) *IntegerValue {
	if i, ok := v.(*IntegerValue); ok {
		return i
	}
	return nil
}

// IsFloat checks if a value is a FloatValue.
func IsFloat(v Value) bool {
	_, ok := v.(*FloatValue)
	return ok
}

// AsFloat returns the value as *FloatValue if possible, nil otherwise.
func AsFloat(v Value) *FloatValue {
	if f, ok := v.(*FloatValue); ok {
		return f
	}
	return nil
}

// IsString checks if a value is a StringValue.
func IsString(v Value) bool {
	_, ok := v.(*StringValue)
	return ok
}

// AsString returns the value as *StringValue if possible, nil otherwise.
func AsString(v Value) *StringValue {
	if s, ok := v.(*StringValue); ok {
		return s
	}
	return nil
}

// IsBoolean checks if a value is a BooleanValue.
func IsBoolean(v Value) bool {
	_, ok := v.(*BooleanValue)
	return ok
}

// AsBoolean returns the value as *BooleanValue if possible, nil otherwise.
func AsBoolean(v Value) *BooleanValue {
	if b, ok := v.(*BooleanValue); ok {
		return b
	}
	return nil
}

// IsNil checks if a value is a NilValue.
func IsNil(v Value) bool {
	_, ok := v.(*NilValue)
	return ok
}

// AsNil returns the value as *NilValue if possible, nil otherwise.
func AsNil(v Value) *NilValue {
	if n, ok := v.(*NilValue); ok {
		return n
	}
	return nil
}

// IsNull checks if a value is a NullValue (variant null).
func IsNull(v Value) bool {
	_, ok := v.(*NullValue)
	return ok
}

// IsUnassigned checks if a value is an UnassignedValue (variant unassigned).
func IsUnassigned(v Value) bool {
	_, ok := v.(*UnassignedValue)
	return ok
}

// IsNumeric checks if a value implements NumericValue.
func IsNumeric(v Value) bool {
	_, ok := v.(NumericValue)
	return ok
}

// AsNumeric returns the value as NumericValue if possible, nil otherwise.
func AsNumeric(v Value) NumericValue {
	if n, ok := v.(NumericValue); ok {
		return n
	}
	return nil
}

// IsComparable checks if a value implements ComparableValue.
func IsComparable(v Value) bool {
	_, ok := v.(ComparableValue)
	return ok
}

// AsComparable returns the value as ComparableValue if possible, nil otherwise.
func AsComparable(v Value) ComparableValue {
	if c, ok := v.(ComparableValue); ok {
		return c
	}
	return nil
}

// IsOrderable checks if a value implements OrderableValue.
func IsOrderable(v Value) bool {
	_, ok := v.(OrderableValue)
	return ok
}

// AsOrderable returns the value as OrderableValue if possible, nil otherwise.
func AsOrderable(v Value) OrderableValue {
	if o, ok := v.(OrderableValue); ok {
		return o
	}
	return nil
}

// IsIndexable checks if a value implements IndexableValue.
func IsIndexable(v Value) bool {
	_, ok := v.(IndexableValue)
	return ok
}

// AsIndexable returns the value as IndexableValue if possible, nil otherwise.
func AsIndexable(v Value) IndexableValue {
	if i, ok := v.(IndexableValue); ok {
		return i
	}
	return nil
}

// IsCopyable checks if a value implements CopyableValue.
func IsCopyable(v Value) bool {
	_, ok := v.(CopyableValue)
	return ok
}

// AsCopyable returns the value as CopyableValue if possible, nil otherwise.
func AsCopyable(v Value) CopyableValue {
	if c, ok := v.(CopyableValue); ok {
		return c
	}
	return nil
}

// IsReferenceType checks if a value implements ReferenceType.
func IsReferenceType(v Value) bool {
	_, ok := v.(ReferenceType)
	return ok
}

// AsReferenceType returns the value as ReferenceType if possible, nil otherwise.
func AsReferenceType(v Value) ReferenceType {
	if r, ok := v.(ReferenceType); ok {
		return r
	}
	return nil
}

// ============================================================================
// Truthiness Utilities
// ============================================================================

// IsTruthy determines if a value is "truthy" in boolean context.
// DWScript rules:
//   - BooleanValue: use the value
//   - IntegerValue/FloatValue: non-zero is true
//   - StringValue: non-empty is true
//   - NilValue: false
//   - Everything else: true (objects exist)
func IsTruthy(v Value) bool {
	if v == nil {
		return false
	}

	switch val := v.(type) {
	case *BooleanValue:
		return val.Value
	case *IntegerValue:
		return val.Value != 0
	case *FloatValue:
		return val.Value != 0.0
	case *StringValue:
		return val.Value != ""
	case *NilValue:
		return false
	case *NullValue:
		return false
	case *UnassignedValue:
		return false
	default:
		// Objects, arrays, etc. are truthy if they exist
		return true
	}
}

// IsFalsy is the inverse of IsTruthy.
func IsFalsy(v Value) bool {
	return !IsTruthy(v)
}

// ============================================================================
// Value Comparison Utilities
// ============================================================================

// Equal compares two values for equality.
// Uses ComparableValue interface if available, falls back to type-specific comparison.
func Equal(left, right Value) (bool, error) {
	// Handle nil cases
	if left == nil && right == nil {
		return true, nil
	}
	if left == nil || right == nil {
		return false, nil
	}

	// Try using ComparableValue interface
	if lComp, ok := left.(ComparableValue); ok {
		return lComp.Equals(right)
	}

	// Fall back to simple type check
	if left.Type() != right.Type() {
		return false, nil
	}

	// String comparison
	return left.String() == right.String(), nil
}

// NotEqual is the inverse of Equal.
func NotEqual(left, right Value) (bool, error) {
	eq, err := Equal(left, right)
	if err != nil {
		return false, err
	}
	return !eq, nil
}

// LessThan compares if left < right.
func LessThan(left, right Value) (bool, error) {
	// Try using OrderableValue interface
	if lOrd, ok := left.(OrderableValue); ok {
		cmp, err := lOrd.CompareTo(right)
		if err != nil {
			return false, err
		}
		return cmp < 0, nil
	}

	return false, NewComparisonError(left, right, "<")
}

// LessThanOrEqual compares if left <= right.
func LessThanOrEqual(left, right Value) (bool, error) {
	// Try using OrderableValue interface
	if lOrd, ok := left.(OrderableValue); ok {
		cmp, err := lOrd.CompareTo(right)
		if err != nil {
			return false, err
		}
		return cmp <= 0, nil
	}

	return false, NewComparisonError(left, right, "<=")
}

// GreaterThan compares if left > right.
func GreaterThan(left, right Value) (bool, error) {
	// Try using OrderableValue interface
	if lOrd, ok := left.(OrderableValue); ok {
		cmp, err := lOrd.CompareTo(right)
		if err != nil {
			return false, err
		}
		return cmp > 0, nil
	}

	return false, NewComparisonError(left, right, ">")
}

// GreaterThanOrEqual compares if left >= right.
func GreaterThanOrEqual(left, right Value) (bool, error) {
	// Try using OrderableValue interface
	if lOrd, ok := left.(OrderableValue); ok {
		cmp, err := lOrd.CompareTo(right)
		if err != nil {
			return false, err
		}
		return cmp >= 0, nil
	}

	return false, NewComparisonError(left, right, ">=")
}

// ============================================================================
// String Utilities
// ============================================================================

// StringConcat concatenates two values as strings.
func StringConcat(left, right Value) Value {
	return NewString(ToString(left) + ToString(right))
}

// StringContains checks if a string contains a substring (case-sensitive).
func StringContains(str, substr string) bool {
	return strings.Contains(str, substr)
}

// StringContainsInsensitive checks if a string contains a substring (case-insensitive).
func StringContainsInsensitive(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

// StringEquals compares two strings (case-sensitive).
func StringEquals(left, right string) bool {
	return left == right
}

// StringEqualsInsensitive compares two strings (case-insensitive).
func StringEqualsInsensitive(left, right string) bool {
	return strings.EqualFold(left, right)
}

// ============================================================================
// Value Creation Shortcuts
// ============================================================================

// ValuesEqual is a convenience function to compare two values and return boolean.
// Returns false if comparison errors occur.
func ValuesEqual(left, right Value) bool {
	eq, err := Equal(left, right)
	if err != nil {
		return false
	}
	return eq
}

// CopyValue safely copies a value if it implements CopyableValue.
// Returns the original value if it's not copyable.
func CopyValue(v Value) Value {
	if copyable := AsCopyable(v); copyable != nil {
		return copyable.Copy()
	}
	// Return as-is if not copyable
	return v
}

// ============================================================================
// Type Name Utilities
// ============================================================================

// GetTypeName returns a human-readable type name for a value.
func GetTypeName(v Value) string {
	if v == nil {
		return "nil"
	}
	return v.Type()
}

// TypesMatch checks if two values have the same type.
func TypesMatch(left, right Value) bool {
	if left == nil && right == nil {
		return true
	}
	if left == nil || right == nil {
		return false
	}
	return left.Type() == right.Type()
}
