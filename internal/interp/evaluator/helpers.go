package evaluator

import (
	"unicode/utf8"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// This file contains shared helper functions for the evaluator.
// Task 3.5.9: Extracted from visitor_statements.go for reusability.

// IsTruthy determines if a value is considered "true" for conditional logic.
// Task 9.35: Support Variant→Boolean implicit conversion.
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
// Task 9.35: Variant→Boolean coercion rules:
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

// ValuesEqual compares two values for equality.
// This is used by case statements to match values.
// Phase 3.5.4.41: Migrated from Interpreter.valuesEqual()
func ValuesEqual(left, right Value) bool {
	// Unwrap VariantValue if present (check by type name since VariantValue is in interp package)
	// For now, we don't handle Variant unwrapping in the evaluator
	// This will be handled when Variant types are migrated

	// Handle nil values (uninitialized variants)
	if left == nil && right == nil {
		return true // Both uninitialized variants are equal
	}
	if left == nil || right == nil {
		return false // One is nil, the other is not
	}

	// Handle same type comparisons
	if left.Type() != right.Type() {
		return false
	}

	switch l := left.(type) {
	case *runtime.IntegerValue:
		r, ok := right.(*runtime.IntegerValue)
		if !ok {
			return false
		}
		return l.Value == r.Value
	case *runtime.FloatValue:
		r, ok := right.(*runtime.FloatValue)
		if !ok {
			return false
		}
		return l.Value == r.Value
	case *runtime.StringValue:
		r, ok := right.(*runtime.StringValue)
		if !ok {
			return false
		}
		return l.Value == r.Value
	case *runtime.BooleanValue:
		r, ok := right.(*runtime.BooleanValue)
		if !ok {
			return false
		}
		return l.Value == r.Value
	case *runtime.NilValue:
		return true // nil == nil
	default:
		// For other types (RecordValue, etc.), use string comparison as fallback
		// Phase 3.5.4.41: Record comparison delegated to later migration
		return left.String() == right.String()
	}
}

// IsInRange checks if value is within the range [start, end] inclusive.
// Supports Integer, Float, String (character), and Enum values.
// Phase 3.5.4.41: Migrated from Interpreter.isInRange()
func IsInRange(value, start, end Value) bool {
	// Unwrap VariantValue if present - delegated to later migration
	// For now, assume values are not wrapped in Variant

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
		// Phase 3.5.4.41: Enum range checking delegated to later migration
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
