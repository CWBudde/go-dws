package builtins

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
)

// =============================================================================
// Basic String Operations
// =============================================================================

// Concat concatenates multiple strings or arrays.
// It dispatches based on the type of the first argument:
// - If the first argument is a string, it concatenates strings
// - If the first argument is an array, it concatenates arrays
//
// Signature:
//   - Concat(str1, str2, ...) -> String
//   - Concat(arr1, arr2, ...) -> Array
//
// Example:
//
//	var s := Concat("Hello", " ", "World"); // "Hello World"
//	var a := Concat([1, 2], [3, 4]); // [1, 2, 3, 4]
func Concat(ctx Context, args []Value) Value {
	if len(args) == 0 {
		return ctx.NewError("Concat() expects at least 1 argument, got 0")
	}

	// Check if first argument is an array - if so, dispatch to array concatenation
	if _, ok := args[0].(*runtime.ArrayValue); ok {
		// Import ConcatArrays from array.go
		return ConcatArrays(ctx, args)
	}

	// Otherwise, concatenate strings using the Context helper
	return ctx.ConcatStrings(args)
}

// Pos implements the Pos() built-in function.
// It finds the position of a substring within a string.
// Pos(substr, str [, offset]) - returns 1-based position (0 if not found)
func Pos(ctx Context, args []Value) Value {
	if len(args) != 2 && len(args) != 3 {
		return ctx.NewError("Pos() expects 2 or 3 arguments, got %d", len(args))
	}

	// First argument: substring to find
	substrVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("Pos() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string to search in
	strVal, ok := args[1].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("Pos() expects string as second argument, got %s", args[1].Type())
	}

	// Optional third argument: offset (1-based)
	if len(args) == 3 {
		offsetVal, ok := ctx.ToInt64(args[2])
		if !ok {
			return ctx.NewError("Pos() expects integer as third argument, got %s", args[2].Type())
		}
		// Delegate to PosEx for offset handling
		return PosEx(ctx, []Value{substrVal, strVal, &runtime.IntegerValue{Value: offsetVal}})
	}

	substr := substrVal.Value
	str := strVal.Value

	// Handle empty substring - returns 1 (found at start)
	if len(substr) == 0 {
		return &runtime.IntegerValue{Value: 1}
	}

	// Find the substring
	index := strings.Index(str, substr)

	// Convert to 1-based index (or 0 if not found)
	if index == -1 {
		return &runtime.IntegerValue{Value: 0}
	}

	return &runtime.IntegerValue{Value: int64(index + 1)}
}

// UpperCase implements the UpperCase() built-in function.
// It converts a string to uppercase.
// UpperCase(str) - returns uppercase version of the string
func UpperCase(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("UpperCase() expects exactly 1 argument, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("UpperCase() expects string as argument, got %s", args[0].Type())
	}

	return &runtime.StringValue{Value: strings.ToUpper(strVal.Value)}
}

// LowerCase implements the LowerCase() built-in function.
// It converts a string to lowercase.
// LowerCase(str) - returns lowercase version of the string
func LowerCase(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("LowerCase() expects exactly 1 argument, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("LowerCase() expects string as argument, got %s", args[0].Type())
	}

	return &runtime.StringValue{Value: strings.ToLower(strVal.Value)}
}

// ASCIIUpperCase implements the ASCIIUpperCase() built-in function.
// It converts a string to uppercase using ASCII-only conversion.
// ASCIIUpperCase(str) - returns uppercase version using ASCII rules
func ASCIIUpperCase(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("ASCIIUpperCase() expects exactly 1 argument, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("ASCIIUpperCase() expects string as argument, got %s", args[0].Type())
	}

	// ASCII-only uppercase conversion
	// Only converts ASCII letters (a-z) to uppercase (A-Z)
	result := make([]byte, len(strVal.Value))
	for idx, b := range []byte(strVal.Value) {
		if b >= 'a' && b <= 'z' {
			result[idx] = b - 32 // Convert to uppercase
		} else {
			result[idx] = b
		}
	}

	return &runtime.StringValue{Value: string(result)}
}

// ASCIILowerCase implements the ASCIILowerCase() built-in function.
// It converts a string to lowercase using ASCII-only conversion.
// ASCIILowerCase(str) - returns lowercase version using ASCII rules
func ASCIILowerCase(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("ASCIILowerCase() expects exactly 1 argument, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("ASCIILowerCase() expects string as argument, got %s", args[0].Type())
	}

	// ASCII-only lowercase conversion
	// Only converts ASCII letters (A-Z) to lowercase (a-z)
	result := make([]byte, len(strVal.Value))
	for idx, b := range []byte(strVal.Value) {
		if b >= 'A' && b <= 'Z' {
			result[idx] = b + 32 // Convert to lowercase
		} else {
			result[idx] = b
		}
	}

	return &runtime.StringValue{Value: string(result)}
}

// AnsiUpperCase implements the AnsiUpperCase() built-in function.
// It is an alias for UpperCase() - converts a string to uppercase.
// AnsiUpperCase(str) - returns uppercase version of the string
func AnsiUpperCase(ctx Context, args []Value) Value {
	// AnsiUpperCase is just an alias for UpperCase
	return UpperCase(ctx, args)
}

// AnsiLowerCase implements the AnsiLowerCase() built-in function.
// It is an alias for LowerCase() - converts a string to lowercase.
// AnsiLowerCase(str) - returns lowercase version of the string
func AnsiLowerCase(ctx Context, args []Value) Value {
	// AnsiLowerCase is just an alias for LowerCase
	return LowerCase(ctx, args)
}

// Trim implements the Trim() built-in function.
// Trim(str) - removes leading/trailing whitespace
// Trim(str, left, right) - removes left/right character counts
func Trim(ctx Context, args []Value) Value {
	if len(args) != 1 && len(args) != 3 {
		return ctx.NewError("Trim() expects 1 or 3 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("Trim() expects string as first argument, got %s", args[0].Type())
	}

	if len(args) == 1 {
		return &runtime.StringValue{Value: strings.TrimSpace(strVal.Value)}
	}

	leftVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("Trim() expects integer as second argument, got %s", args[1].Type())
	}
	rightVal, ok := args[2].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("Trim() expects integer as third argument, got %s", args[2].Type())
	}

	runes := []rune(strVal.Value)
	left := int(leftVal.Value)
	right := int(rightVal.Value)
	if left < 0 {
		left = 0
	}
	if right < 0 {
		right = 0
	}
	if left+right >= len(runes) {
		return &runtime.StringValue{Value: ""}
	}
	return &runtime.StringValue{Value: string(runes[left : len(runes)-right])}
}

// TrimLeft implements the TrimLeft() built-in function.
// It removes leading whitespace or a number of characters.
// TrimLeft(str) - returns string with leading whitespace removed
// TrimLeft(str, count) - removes count leading characters
func TrimLeft(ctx Context, args []Value) Value {
	if len(args) != 1 && len(args) != 2 {
		return ctx.NewError("TrimLeft() expects 1 or 2 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("TrimLeft() expects string as argument, got %s", args[0].Type())
	}

	if len(args) == 1 {
		return &runtime.StringValue{Value: strings.TrimLeft(strVal.Value, " \t\n\r")}
	}

	countVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("TrimLeft() expects integer as second argument, got %s", args[1].Type())
	}
	count := int(countVal.Value)
	if count < 0 {
		count = 0
	}
	runes := []rune(strVal.Value)
	if count >= len(runes) {
		return &runtime.StringValue{Value: ""}
	}
	return &runtime.StringValue{Value: string(runes[count:])}
}

// TrimRight implements the TrimRight() built-in function.
// It removes trailing whitespace or a number of characters.
// TrimRight(str) - returns string with trailing whitespace removed
// TrimRight(str, count) - removes count trailing characters
func TrimRight(ctx Context, args []Value) Value {
	if len(args) != 1 && len(args) != 2 {
		return ctx.NewError("TrimRight() expects 1 or 2 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("TrimRight() expects string as argument, got %s", args[0].Type())
	}

	if len(args) == 1 {
		return &runtime.StringValue{Value: strings.TrimRight(strVal.Value, " \t\n\r")}
	}

	countVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("TrimRight() expects integer as second argument, got %s", args[1].Type())
	}
	count := int(countVal.Value)
	if count < 0 {
		count = 0
	}
	runes := []rune(strVal.Value)
	if count >= len(runes) {
		return &runtime.StringValue{Value: ""}
	}
	return &runtime.StringValue{Value: string(runes[:len(runes)-count])}
}

// StringReplace implements the StringReplace() built-in function.
// It replaces occurrences of a substring within a string.
// StringReplace(str, old, new) - replaces all occurrences of old with new
// StringReplace(str, old, new, count) - replaces count occurrences (count=-1 means all)
func StringReplace(ctx Context, args []Value) Value {
	// Accept 3 or 4 arguments
	if len(args) < 3 || len(args) > 4 {
		return ctx.NewError("StringReplace() expects 3 or 4 arguments, got %d", len(args))
	}

	// First argument: string to search in
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StringReplace() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: old substring
	oldVal, ok := args[1].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StringReplace() expects string as second argument, got %s", args[1].Type())
	}

	// Third argument: new substring
	newVal, ok := args[2].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StringReplace() expects string as third argument, got %s", args[2].Type())
	}

	str := strVal.Value
	old := oldVal.Value
	new := newVal.Value

	// Default count: -1 means replace all
	count := -1

	// Optional fourth argument: count
	if len(args) == 4 {
		countVal, ok := args[3].(*runtime.IntegerValue)
		if !ok {
			return ctx.NewError("StringReplace() expects integer as fourth argument, got %s", args[3].Type())
		}
		count = int(countVal.Value)
	}

	// Handle edge cases
	// Empty old string: return original (can't replace nothing)
	if len(old) == 0 {
		return &runtime.StringValue{Value: str}
	}

	// Count is 0 or negative (except -1): no replacement
	if count == 0 || (count < 0 && count != -1) {
		return &runtime.StringValue{Value: str}
	}

	// Perform replacement
	var result string
	if count == -1 {
		result = strings.ReplaceAll(str, old, new)
	} else {
		result = strings.Replace(str, old, new, count)
	}

	return &runtime.StringValue{Value: result}
}

// StrReplace is an alias for StringReplace for DWScript compatibility.
func StrReplace(ctx Context, args []Value) Value {
	return StringReplace(ctx, args)
}

// StringOfChar implements the StringOfChar() built-in function.
// It creates a string by repeating a character N times.
// StringOfChar(ch, count) - returns a string with ch repeated count times
func StringOfChar(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("StringOfChar() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: character (string)
	charVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StringOfChar() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: count (integer)
	countVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("StringOfChar() expects integer as second argument, got %s", args[1].Type())
	}

	count := int(countVal.Value)

	// Handle edge cases
	// If count <= 0, return empty string
	if count <= 0 {
		return &runtime.StringValue{Value: ""}
	}

	// Extract the first character from the string
	// If the string is empty, return empty string
	if runeLength(charVal.Value) == 0 {
		return &runtime.StringValue{Value: ""}
	}

	// Get the first character (rune-based to handle UTF-8)
	firstRune, _ := runeAt(charVal.Value, 1)
	ch := string(firstRune)

	// Use strings.Repeat to create the repeated string
	result := strings.Repeat(ch, count)

	return &runtime.StringValue{Value: result}
}

// SubStr implements the SubStr() built-in function.
// It extracts a substring from a string with a length parameter.
// SubStr(str, start) - returns substring from start to end (1-based)
// SubStr(str, start, length) - returns length characters starting at start (1-based)
// Note: Different from SubString which takes an end position instead of length.
func SubStr(ctx Context, args []Value) Value {
	// Accept 2 or 3 arguments
	if len(args) < 2 || len(args) > 3 {
		return ctx.NewError("SubStr() expects 2 or 3 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("SubStr() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: start position (1-based)
	startVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("SubStr() expects integer as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	start := startVal.Value // 1-based

	// Third argument (optional): length
	// Default is MaxInt (meaning "to end of string")
	length := int64(1<<31 - 1) // MaxInt for 32-bit (matches DWScript behavior)
	if len(args) == 3 {
		lengthVal, ok := args[2].(*runtime.IntegerValue)
		if !ok {
			return ctx.NewError("SubStr() expects integer as third argument, got %s", args[2].Type())
		}
		length = lengthVal.Value
	}

	// Use rune-based slicing to handle UTF-8 correctly
	// This is the same logic as Copy()
	result := runeSliceFrom(str, int(start), int(length))
	return &runtime.StringValue{Value: result}
}

// NOTE: Format() is implemented below in the "Advanced String Operations" section.

// NOTE: Insert() and Delete() for strings are var-param functions implemented in the
// interpreter (internal/interp/builtins_var_param.go) because they require AST-level
// access to modify variables in-place. They cannot be migrated to this builtins package
// since the Context interface only provides evaluated values, not AST nodes.
//
// Implemented var-param string functions:
//   - Insert(source, var dest, pos) - builtinInsert() in internal/interp/builtins_var_param.go
//   - Delete(var s, pos, count) - builtinDeleteString() in internal/interp/builtins_var_param.go
//
// These are routed through callBuiltinWithVarParam() in functions_builtins.go.

// Chr implements the Chr() built-in function.
// It converts an integer character code to a single-character string.
// Chr(code: Integer): String
func Chr(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Chr() expects exactly 1 argument, got %d", len(args))
	}

	// Argument must be Integer
	intVal, ok := args[0].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("Chr() expects Integer argument, got %s", args[0].Type())
	}

	// Check if the code is in valid range (0-1114111 for Unicode)
	if intVal.Value < 0 || intVal.Value > 0x10FFFF {
		return ctx.NewError("Chr() code %d out of valid Unicode range (0-1114111)", intVal.Value)
	}

	return &runtime.StringValue{Value: string(rune(intVal.Value))}
}

// IntToHex implements the IntToHex() built-in function.
// It converts an integer to a hexadecimal string with specified minimum number of digits.
// IntToHex(value: Integer, digits: Integer): String
func IntToHex(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("IntToHex() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument must be Integer (the value to convert)
	intVal, ok := args[0].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("IntToHex() first argument must be Integer, got %s", args[0].Type())
	}

	// Second argument must be Integer (minimum number of digits)
	digitsVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("IntToHex() second argument must be Integer, got %s", args[1].Type())
	}

	// Convert to hexadecimal string with uppercase letters
	hexStr := fmt.Sprintf("%X", uint64(intVal.Value))

	// Pad with zeros if necessary to reach minimum digit count
	if digitsVal.Value > 0 && int64(len(hexStr)) < digitsVal.Value {
		// Pad with leading zeros
		hexStr = strings.Repeat("0", int(digitsVal.Value)-len(hexStr)) + hexStr
	}

	return &runtime.StringValue{Value: hexStr}
}

// StrToBool implements the StrToBool() built-in function.
// It converts a string to a boolean value.
// Accepts: 'True', 'False', '1', '0', 'Yes', 'No' (case-insensitive)
// StrToBool(s: String): Boolean
func StrToBool(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("StrToBool() expects exactly 1 argument, got %d", len(args))
	}

	// First argument must be String
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrToBool() expects String, got %s", args[0].Type())
	}

	// Normalize to lowercase for case-insensitive matching
	s := strings.ToLower(strings.TrimSpace(strVal.Value))

	// Numeric strings: parse as integer, zero is False, non-zero is True
	if num, err := strconv.ParseInt(s, 10, 64); err == nil {
		return &runtime.BooleanValue{Value: num != 0}
	}

	// Check for true values
	switch s {
	case "true", "1", "yes", "t", "y":
		return &runtime.BooleanValue{Value: true}
	case "false", "0", "no", "f", "n":
		return &runtime.BooleanValue{Value: false}
	default:
		// For invalid strings, return false (DWScript behavior)
		return &runtime.BooleanValue{Value: false}
	}
}

// SubString implements the SubString() built-in function.
// It extracts a substring from a string using start and end positions.
// SubString(str, start, end) - returns substring from start to end (1-based, inclusive)
// Note: Different from SubStr which takes a length parameter instead of end position.
func SubString(ctx Context, args []Value) Value {
	if len(args) != 3 {
		return ctx.NewError("SubString() expects exactly 3 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("SubString() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: start position (1-based)
	startVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("SubString() expects integer as second argument, got %s", args[1].Type())
	}

	// Third argument: end position (1-based, inclusive)
	endVal, ok := args[2].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("SubString() expects integer as third argument, got %s", args[2].Type())
	}

	str := strVal.Value
	start := int(startVal.Value) // 1-based
	end := int(endVal.Value)     // 1-based, inclusive

	// Calculate length using end as an exclusive position (DWScript semantics):
	// SubString(str, 3, 7) returns characters at positions 3..6 → length 4.
	length := end - start

	// Handle edge cases
	if length <= 0 {
		return &runtime.StringValue{Value: ""}
	}

	// Use rune-based slicing to handle UTF-8 correctly
	result := runeSliceFrom(str, start, length)
	return &runtime.StringValue{Value: result}
}

// LeftStr implements the LeftStr() built-in function.
// It returns the leftmost N characters of a string.
// LeftStr(str, count) - returns first count characters
func LeftStr(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("LeftStr() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("LeftStr() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: count
	countVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("LeftStr() expects integer as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	count := int(countVal.Value)

	// Handle edge cases
	if count <= 0 {
		return &runtime.StringValue{Value: ""}
	}

	// Use rune-based slicing to handle UTF-8 correctly
	// LeftStr is equivalent to SubStr(str, 1, count)
	result := runeSliceFrom(str, 1, count)
	return &runtime.StringValue{Value: result}
}

// RightStr implements the RightStr() built-in function.
// It returns the rightmost N characters of a string.
// RightStr(str, count) - returns last count characters
func RightStr(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("RightStr() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("RightStr() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: count
	countVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("RightStr() expects integer as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	count := int(countVal.Value)

	// Handle edge cases
	if count <= 0 {
		return &runtime.StringValue{Value: ""}
	}

	// Get the length of the string in runes (not bytes)
	strLen := runeLength(str)

	// If count >= length, return the whole string
	if count >= strLen {
		return &runtime.StringValue{Value: str}
	}

	// Calculate start position (1-based)
	// For a string of length 10, RightStr(str, 3) should return positions 8, 9, 10
	start := strLen - count + 1

	// Use rune-based slicing to handle UTF-8 correctly
	result := runeSliceFrom(str, start, count)
	return &runtime.StringValue{Value: result}
}

// MidStr implements the MidStr() built-in function.
// It is an alias for SubStr - extracts a substring with a length parameter.
// MidStr(str, start, count) - returns count characters starting at start (1-based)
func MidStr(ctx Context, args []Value) Value {
	// MidStr is just an alias for SubStr
	return SubStr(ctx, args)
}

// StrBeginsWith implements the StrBeginsWith() built-in function.
// It checks if a string starts with a given prefix.
// StrBeginsWith(str, prefix) - returns true if str starts with prefix
func StrBeginsWith(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("StrBeginsWith() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string to check
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrBeginsWith() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: prefix
	prefixVal, ok := args[1].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrBeginsWith() expects string as second argument, got %s", args[1].Type())
	}

	// An empty prefix returns False in DWScript fixtures.
	if len(prefixVal.Value) == 0 {
		return &runtime.BooleanValue{Value: false}
	}

	result := strings.HasPrefix(strVal.Value, prefixVal.Value)
	return &runtime.BooleanValue{Value: result}
}

// StrEndsWith implements the StrEndsWith() built-in function.
// It checks if a string ends with a given suffix.
// StrEndsWith(str, suffix) - returns true if str ends with suffix
func StrEndsWith(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("StrEndsWith() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string to check
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrEndsWith() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: suffix
	suffixVal, ok := args[1].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrEndsWith() expects string as second argument, got %s", args[1].Type())
	}

	// Empty suffix returns false per DWScript reference behavior.
	if len(suffixVal.Value) == 0 {
		return &runtime.BooleanValue{Value: false}
	}

	result := strings.HasSuffix(strVal.Value, suffixVal.Value)
	return &runtime.BooleanValue{Value: result}
}

// StrContains implements the StrContains() built-in function.
// It checks if a string contains a given substring.
// StrContains(str, substring) - returns true if str contains substring
func StrContains(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("StrContains() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string to check
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrContains() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: substring
	substrVal, ok := args[1].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrContains() expects string as second argument, got %s", args[1].Type())
	}

	result := strings.Contains(strVal.Value, substrVal.Value)
	return &runtime.BooleanValue{Value: result}
}

// PosEx implements the PosEx() built-in function.
// It finds the position of a substring within a string, starting from an offset.
// PosEx(needle, haystack, offset) - returns 1-based position (0 if not found)
func PosEx(ctx Context, args []Value) Value {
	if len(args) != 3 {
		return ctx.NewError("PosEx() expects exactly 3 arguments, got %d", len(args))
	}

	// First argument: substring to find (needle)
	needleVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("PosEx() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string to search in (haystack)
	haystackVal, ok := args[1].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("PosEx() expects string as second argument, got %s", args[1].Type())
	}

	// Third argument: offset (1-based starting position)
	offsetVal, ok := args[2].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("PosEx() expects integer as third argument, got %s", args[2].Type())
	}

	needle := needleVal.Value
	haystack := haystackVal.Value
	offset := int(offsetVal.Value) // 1-based

	// Handle invalid offset first (before empty needle check)
	// This prevents returning negative positions
	if offset < 1 {
		return &runtime.IntegerValue{Value: 0}
	}

	// Handle empty needle - returns 0 (not found)
	// This matches the original DWScript behavior
	if len(needle) == 0 {
		return &runtime.IntegerValue{Value: 0}
	}

	// Convert to rune-based indexing for UTF-8 support
	haystackRunes := []rune(haystack)
	needleRunes := []rune(needle)

	// Adjust offset to 0-based
	startIdx := offset - 1

	// If offset is beyond the string length, not found
	if startIdx >= len(haystackRunes) {
		return &runtime.IntegerValue{Value: 0}
	}

	// Search for the needle starting from offset
	for i := startIdx; i <= len(haystackRunes)-len(needleRunes); i++ {
		match := true
		for j := 0; j < len(needleRunes); j++ {
			if haystackRunes[i+j] != needleRunes[j] {
				match = false
				break
			}
		}
		if match {
			// Return 1-based position
			return &runtime.IntegerValue{Value: int64(i + 1)}
		}
	}

	// Not found
	return &runtime.IntegerValue{Value: 0}
}

// RevPos implements the RevPos() built-in function.
// It finds the last position of a substring within a string.
// RevPos(needle, haystack) - returns 1-based position of last occurrence (0 if not found)
func RevPos(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("RevPos() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: substring to find (needle)
	needleVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("RevPos() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string to search in (haystack)
	haystackVal, ok := args[1].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("RevPos() expects string as second argument, got %s", args[1].Type())
	}

	needle := needleVal.Value
	haystack := haystackVal.Value

	// Handle empty needle - DWScript returns length + 1
	if len(needle) == 0 {
		var length int64
		for range haystack {
			length++
		}
		return &runtime.IntegerValue{Value: length + 1}
	}

	// Find the last occurrence using strings.LastIndex
	index := strings.LastIndex(haystack, needle)

	// Convert to 1-based index (or 0 if not found)
	if index == -1 {
		return &runtime.IntegerValue{Value: 0}
	}

	// Convert byte index to rune index (for UTF-8 support)
	runeIndex := len([]rune(haystack[:index])) + 1
	return &runtime.IntegerValue{Value: int64(runeIndex)}
}

// StrFind implements the StrFind() built-in function.
// It is an alias for PosEx - finds substring with starting index.
// StrFind(str, substr [, fromIndex]) - returns 1-based position (0 if not found)
func StrFind(ctx Context, args []Value) Value {
	if len(args) != 2 && len(args) != 3 {
		return ctx.NewError("StrFind() expects 2 or 3 arguments, got %d", len(args))
	}

	fromIndex := &runtime.IntegerValue{Value: 1}
	if len(args) == 3 {
		offsetVal, ok := ctx.ToInt64(args[2])
		if !ok {
			return ctx.NewError("StrFind() expects integer as third argument, got %s", args[2].Type())
		}
		fromIndex = &runtime.IntegerValue{Value: offsetVal}
	}

	// StrFind(str, substr, fromIndex) maps to PosEx(substr, str, fromIndex)
	// Need to reorder arguments
	reorderedArgs := []Value{
		args[1], // substr becomes first arg (needle)
		args[0], // str becomes second arg (haystack)
		fromIndex,
	}

	return PosEx(ctx, reorderedArgs)
}

// ByteSizeToStr implements the ByteSizeToStr() built-in function.
// It formats a byte size into a human-readable string (KB, MB, GB, TB).
// ByteSizeToStr(size: Integer): String
func ByteSizeToStr(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("ByteSizeToStr() expects exactly 1 argument, got %d", len(args))
	}

	// First argument: byte size (integer)
	sizeVal, ok := args[0].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("ByteSizeToStr() expects integer as argument, got %s", args[0].Type())
	}

	size := float64(sizeVal.Value)

	// Define size units
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)

	// Format based on size
	var result string
	absSize := size
	if absSize < 0 {
		absSize = -absSize
	}

	switch {
	case absSize < KB:
		result = fmt.Sprintf("%d bytes", int64(size))
	case absSize < MB:
		result = fmt.Sprintf("%.2f KB", size/KB)
	case absSize < GB:
		result = fmt.Sprintf("%.2f MB", size/MB)
	case absSize < TB:
		result = fmt.Sprintf("%.2f GB", size/GB)
	default:
		result = fmt.Sprintf("%.2f TB", size/TB)
	}

	return &runtime.StringValue{Value: result}
}

// GetText implements the GetText() built-in function.
// It is a localization/translation function that returns the input string unchanged.
// In a full implementation, this would look up translations.
// GetText(str: String): String
func GetText(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("GetText() expects exactly 1 argument, got %d", len(args))
	}

	// First argument: string to translate
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("GetText() expects string as argument, got %s", args[0].Type())
	}

	// For now, just return the input string unchanged
	// In a full implementation, this would look up translations from a resource file
	return &runtime.StringValue{Value: strVal.Value}
}

// Underscore implements the _() built-in function.
// It is an alias for GetText() - a localization/translation function.
// _(str: String): String
func Underscore(ctx Context, args []Value) Value {
	// _() is just an alias for GetText()
	return GetText(ctx, args)
}

// CharAt implements the CharAt() built-in function.
// It returns the character at the specified position in a string (1-based).
// This function is deprecated in favor of SubStr().
// CharAt(s: String, x: Integer): String
func CharAt(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("CharAt() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("CharAt() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: position (1-based)
	posVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("CharAt() expects integer as second argument, got %s", args[1].Type())
	}

	// Use SubStr to get a single character
	// SubStr expects (str, start, length)
	subStrArgs := []Value{
		strVal,
		posVal,
		&runtime.IntegerValue{Value: 1}, // length = 1
	}

	return SubStr(ctx, subStrArgs)
}

// =============================================================================
// Advanced String Operations
// =============================================================================

// StrSplit implements the StrSplit() built-in function.
// It splits a string into an array of strings using a delimiter.
// If the delimiter is empty, returns an array with the original string as the sole element.
//
// Signature: StrSplit(str, delimiter) -> array of String
//
// Example:
//
//	var parts := StrSplit("a,b,c", ","); // ["a", "b", "c"]
func StrSplit(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("StrSplit() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string to split
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrSplit() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: delimiter
	delimVal, ok := args[1].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrSplit() expects string as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	delim := delimVal.Value

	var parts []string
	if len(delim) == 0 {
		if len(str) == 0 {
			parts = []string{}
		} else {
			for _, r := range []rune(str) {
				parts = append(parts, string(r))
			}
		}
	} else {
		parts = strings.Split(str, delim)
	}

	// Convert to array of StringValue
	elements := make([]Value, len(parts))
	for idx, part := range parts {
		elements[idx] = &runtime.StringValue{Value: part}
	}

	return &runtime.ArrayValue{
		Elements:  elements,
		ArrayType: types.NewDynamicArrayType(types.STRING),
	}
}

// StrJoin implements the StrJoin() built-in function.
// It joins an array of strings into a single string using a delimiter.
//
// Signature: StrJoin(array, delimiter) -> String
//
// Example:
//
//	var s := StrJoin(["a", "b", "c"], ","); // "a,b,c"
func StrJoin(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("StrJoin() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: array of strings
	arrVal, ok := args[0].(*runtime.ArrayValue)
	if !ok {
		return ctx.NewError("StrJoin() expects array as first argument, got %s", args[0].Type())
	}

	// Second argument: delimiter
	delimVal, ok := args[1].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrJoin() expects string as second argument, got %s", args[1].Type())
	}

	delim := delimVal.Value

	// Convert array elements to strings
	parts := make([]string, len(arrVal.Elements))
	for idx, elem := range arrVal.Elements {
		strElem, ok := elem.(*runtime.StringValue)
		if !ok {
			return ctx.NewError("StrJoin() expects array of strings, got %s at index %d", elem.Type(), idx)
		}
		parts[idx] = strElem.Value
	}

	// Join the strings
	result := strings.Join(parts, delim)
	return &runtime.StringValue{Value: result}
}

// StrArrayPack implements the StrArrayPack() built-in function.
// It removes empty strings from an array.
//
// Signature: StrArrayPack(array) -> array of String
//
// Example:
//
//	var packed := StrArrayPack(["a", "", "b", "", "c"]); // ["a", "b", "c"]
func StrArrayPack(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("StrArrayPack() expects exactly 1 argument, got %d", len(args))
	}

	// First argument: array of strings
	arrVal, ok := args[0].(*runtime.ArrayValue)
	if !ok {
		return ctx.NewError("StrArrayPack() expects array as argument, got %s", args[0].Type())
	}

	// Filter out empty strings
	var packed []Value
	for _, elem := range arrVal.Elements {
		strElem, ok := elem.(*runtime.StringValue)
		if !ok {
			return ctx.NewError("StrArrayPack() expects array of strings, got %s", elem.Type())
		}
		if strElem.Value != "" {
			packed = append(packed, strElem)
		}
	}

	return &runtime.ArrayValue{
		Elements:  packed,
		ArrayType: types.NewDynamicArrayType(types.STRING),
	}
}

// NOTE: Format() is implemented in system.go and registered there.
