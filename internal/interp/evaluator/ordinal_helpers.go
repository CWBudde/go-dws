package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Ordinal Value Utilities
// ============================================================================
//
// Task 3.5.77: Ordinal helpers moved to evaluator package for direct access
// during set/array literal evaluation. These functions extract ordinal values
// from runtime values and determine their types.
// ============================================================================

// GetOrdinalValue extracts the ordinal value from any ordinal type value.
// Ordinal types include: Integer, Enum, String (single character), Boolean.
// Returns the ordinal value and an error if the value is not an ordinal type.

func GetOrdinalValue(val Value) (int, error) {
	switch v := val.(type) {
	case *runtime.IntegerValue:
		// Integer values are their own ordinals
		return int(v.Value), nil

	case *runtime.EnumValue:
		// Enum values have an ordinal value field
		return v.OrdinalValue, nil

	case *runtime.StringValue:
		// String values represent characters - use the first character's Unicode code point
		// Note: We must count runes (characters), not bytes, since UTF-8 encoding
		// can use multiple bytes per character (e.g., chr(255) = 'ÿ' uses 2 bytes)
		runes := []rune(v.Value)
		if len(runes) == 0 {
			return 0, fmt.Errorf("cannot get ordinal value of empty string")
		}
		if len(runes) > 1 {
			return 0, fmt.Errorf("cannot get ordinal value of multi-character string '%s'", v.Value)
		}
		// Return the Unicode code point of the single character
		return int(runes[0]), nil

	case *runtime.BooleanValue:
		// Boolean: False=0, True=1
		if v.Value {
			return 1, nil
		}
		return 0, nil

	default:
		return 0, fmt.Errorf("value of type %s is not an ordinal type", val.Type())
	}
}

// GetOrdinalType extracts the Type from a runtime value.
// Returns the appropriate type for the value to use in set types.
func GetOrdinalType(val Value) types.Type {
	switch val.(type) {
	case *runtime.IntegerValue:
		return types.INTEGER
	case *runtime.EnumValue:
		// For enum values, we need the specific enum type
		// This is handled separately in evalSetLiteral
		return nil
	case *runtime.StringValue:
		// Character literals are represented as strings
		return types.STRING
	case *runtime.BooleanValue:
		return types.BOOLEAN
	default:
		return nil
	}
}
