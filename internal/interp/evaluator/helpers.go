package evaluator

import (
	"unicode/utf8"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// This file contains shared helper functions for the evaluator.

// IsTruthy determines if a value is considered "true" for conditional logic.
// Support Variant→Boolean implicit conversion.
// DWScript semantics for Variant→Boolean: empty/nil/zero → false, otherwise → true
func IsTruthy(val Value) bool {
	switch v := val.(type) {
	case *runtime.BooleanValue:
		return v.Value
	default:
		// Check if this is a Variant type by type name
		// (VariantValue is in internal/interp, not runtime, so we check by type string)
		if val.Type() == "VARIANT" {
			// For variants, we need to unwrap and check the underlying value
			// This requires accessing the Value field, but VariantValue is not imported here
			// Use VariantToBool helper
			return VariantToBool(val)
		}
		// In DWScript, only booleans and variants can be used in conditions
		// Non-boolean values in conditionals would be a type error
		// But we return false as a safe default
		return false
	}
}

// VariantToBool converts a variant value to boolean using DWScript semantics.
// Support Variant→Boolean coercion rules:
// - nil/null → false
// - Integer 0 → false, non-zero → true
// - Float 0.0 → false, non-zero → true
// - String "" → false, non-empty → true
// - Objects/arrays → true (non-nil)
func VariantToBool(val Value) bool {
	if val == nil {
		return false
	}

	// First, check if this is a VariantWrapper and unwrap it
	// This handles *VariantValue without needing to import it (avoids circular dependency)
	if wrapper, ok := val.(runtime.VariantWrapper); ok {
		unwrapped := wrapper.UnwrapVariant()
		// Recursively evaluate the unwrapped value
		return VariantToBool(unwrapped)
	}

	switch v := val.(type) {
	case *runtime.BooleanValue:
		return v.Value
	case *runtime.IntegerValue:
		return v.Value != 0
	case *runtime.FloatValue:
		return v.Value != 0.0
	case *runtime.StringValue:
		return v.Value != ""
	case *runtime.NilValue:
		return false
	default:
		// Check by type name for types not in runtime package
		switch val.Type() {
		case "NIL", "UNASSIGNED":
			return false
		default:
			// For objects, arrays, records, etc: non-nil → true
			return true
		}
	}
}

// ValuesEqual is now defined in array_helpers.go
// with full Variant unwrapping and RecordValue support.

// IsInRange checks if value is within the range [start, end] inclusive.
// Supports Integer, Float, String (character), and Enum values.
func IsInRange(value, start, end Value) bool {
	// Handle nil values (uninitialized variants)
	if value == nil || start == nil || end == nil {
		return false // Cannot perform range check with uninitialized variants
	}

	// Handle different value types
	switch v := value.(type) {
	case *runtime.IntegerValue:
		startInt, startOk := start.(*runtime.IntegerValue)
		endInt, endOk := end.(*runtime.IntegerValue)
		if startOk && endOk {
			return v.Value >= startInt.Value && v.Value <= endInt.Value
		}

	case *runtime.FloatValue:
		startFloat, startOk := start.(*runtime.FloatValue)
		endFloat, endOk := end.(*runtime.FloatValue)
		if startOk && endOk {
			return v.Value >= startFloat.Value && v.Value <= endFloat.Value
		}

	case *runtime.StringValue:
		// For strings, compare character by character
		startStr, startOk := start.(*runtime.StringValue)
		endStr, endOk := end.(*runtime.StringValue)
		// Use rune-based comparison to handle UTF-8 correctly
		if startOk && endOk && RuneLength(v.Value) == 1 && RuneLength(startStr.Value) == 1 && RuneLength(endStr.Value) == 1 {
			// Single character comparison (for 'A'..'Z' style ranges)
			charVal, _ := RuneAt(v.Value, 1)
			charStart, _ := RuneAt(startStr.Value, 1)
			charEnd, _ := RuneAt(endStr.Value, 1)
			return charVal >= charStart && charVal <= charEnd
		}
		// Fall back to string comparison for multi-char strings
		if startOk && endOk {
			return v.Value >= startStr.Value && v.Value <= endStr.Value
		}

	default:
		// For EnumValue and other types, check by type name
		// TODO: Implement enum range checking
		return false
	}

	return false
}

// RuneLength returns the number of Unicode characters (runes) in a string.
func RuneLength(s string) int {
	return utf8.RuneCountInString(s)
}

// RuneAt returns the rune at the given 1-based index in the string.
// Returns the rune and true if the index is valid, or 0 and false otherwise.
func RuneAt(s string, index int) (rune, bool) {
	if index < 1 {
		return 0, false
	}

	runes := []rune(s)
	if index > len(runes) {
		return 0, false
	}

	return runes[index-1], true
}

// RuneReplace replaces the rune at the given 1-based index with the given rune.
// Returns the new string and true if successful, or empty string and false otherwise.
func RuneReplace(s string, index int, replacement rune) (string, bool) {
	if index < 1 {
		return "", false
	}

	runes := []rune(s)
	if index > len(runes) {
		return "", false
	}

	runes[index-1] = replacement
	return string(runes), true
}

// IsFalsey determines if a value is "falsey" (default/zero value for its type).
// Exported for use by test utilities across packages.
func IsFalsey(val Value) bool {
	// Handle nil (from unassigned variants)
	if val == nil {
		return true
	}

	switch v := val.(type) {
	case *runtime.IntegerValue:
		return v.Value == 0
	case *runtime.FloatValue:
		return v.Value == 0.0
	case *runtime.StringValue:
		return v.Value == ""
	case *runtime.BooleanValue:
		return !v.Value
	case *runtime.NilValue:
		return true
	case *runtime.ArrayValue:
		// Empty arrays are falsey
		return len(v.Elements) == 0
	default:
		// Check by type name for types not in runtime package
		switch val.Type() {
		case "NIL", "UNASSIGNED", "NULL":
			return true
		case "VARIANT":
			// Variant values need to be unwrapped
			if wrapper, ok := val.(runtime.VariantWrapper); ok {
				return IsFalsey(wrapper.UnwrapVariant())
			}
			return false
		default:
			// Other types (objects, classes, etc.) are truthy if non-nil
			return false
		}
	}
}

// unwrapVariant unwraps a Variant value to its underlying value.
func unwrapVariant(value Value) Value {
	// Check if this is a VariantWrapper (runtime.VariantWrapper interface)
	if wrapper, ok := value.(runtime.VariantWrapper); ok {
		unwrapped := wrapper.UnwrapVariant()
		if unwrapped == nil {
			// Uninitialized variant becomes nil value
			return &runtime.NilValue{}
		}
		return unwrapped
	}
	return value
}
