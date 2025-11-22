package builtins

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// ============================================================================
// Ordinal Built-in Functions
// ============================================================================
//
// This file contains ordinal value manipulation functions that have been
// migrated from internal/interp to use the Context interface pattern.
//
// Functions in this file:
//   - Ord: Returns the ordinal value of enum/boolean/char/integer
//   - Chr: Converts integer character code to string
//   - Succ: Returns successor of ordinal value
//   - Pred: Returns predecessor of ordinal value
//
// Note: Inc() and Dec() cannot be migrated using this pattern as they take
// AST expressions (lvalues) and modify them in place.

// Ord returns the ordinal value of an enum, boolean, character, or integer.
// Ord(value): Integer
//
// Supported types:
//   - Enum: returns the ordinal value (position) of the enum member
//   - Boolean: False → 0, True → 1
//   - Integer: returns the value unchanged
//   - String (character): returns Unicode code point of first character
//   - Empty string: returns 0
//
// Example:
//
//	Ord('A')  // Returns: 65
//	Ord(True) // Returns: 1
func Ord(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Ord() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Handle enum values (using Context helper to avoid circular dependency)
	if ordinal, ok := ctx.GetEnumOrdinal(arg); ok {
		return &runtime.IntegerValue{Value: ordinal}
	}

	// Handle boolean values (False=0, True=1)
	if boolVal, ok := arg.(*runtime.BooleanValue); ok {
		if boolVal.Value {
			return &runtime.IntegerValue{Value: 1}
		}
		return &runtime.IntegerValue{Value: 0}
	}

	// Handle integer values (pass through)
	if intVal, ok := arg.(*runtime.IntegerValue); ok {
		return intVal
	}

	// Handle string values (characters)
	// In DWScript, character literals are single-character strings
	if strVal, ok := arg.(*runtime.StringValue); ok {
		if len(strVal.Value) == 0 {
			// Empty string returns 0
			return &runtime.IntegerValue{Value: 0}
		}
		// Return the Unicode code point of the first character
		runes := []rune(strVal.Value)
		return &runtime.IntegerValue{Value: int64(runes[0])}
	}

	return ctx.NewError("Ord() expects enum, boolean, integer, or string, got %s", arg.Type())
}
