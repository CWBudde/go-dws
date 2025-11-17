package builtins

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
	"golang.org/x/text/unicode/norm"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// =============================================================================
// Basic String Operations
// =============================================================================

// TODO: Concat - Requires ArrayValue to be moved to runtime package
// Original signature: func (i *Interpreter) builtinConcat(args []Value) Value
// This function dispatches to builtinConcatArrays for array arguments, which
// requires ArrayValue support.

// Pos implements the Pos() built-in function.
// It finds the position of a substring within a string.
// Pos(substr, str) - returns 1-based position (0 if not found)
func Pos(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("Pos() expects exactly 2 arguments, got %d", len(args))
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
// It removes leading and trailing whitespace from a string.
// Trim(str) - returns string with whitespace removed from both ends
func Trim(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Trim() expects exactly 1 argument, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("Trim() expects string as argument, got %s", args[0].Type())
	}

	return &runtime.StringValue{Value: strings.TrimSpace(strVal.Value)}
}

// TrimLeft implements the TrimLeft() built-in function.
// It removes leading whitespace from a string.
// TrimLeft(str) - returns string with leading whitespace removed
func TrimLeft(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("TrimLeft() expects exactly 1 argument, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("TrimLeft() expects string as argument, got %s", args[0].Type())
	}

	// Use TrimLeft to remove leading whitespace
	return &runtime.StringValue{Value: strings.TrimLeft(strVal.Value, " \t\n\r")}
}

// TrimRight implements the TrimRight() built-in function.
// It removes trailing whitespace from a string.
// TrimRight(str) - returns string with trailing whitespace removed
func TrimRight(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("TrimRight() expects exactly 1 argument, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("TrimRight() expects string as argument, got %s", args[0].Type())
	}

	// Use TrimRight to remove trailing whitespace
	return &runtime.StringValue{Value: strings.TrimRight(strVal.Value, " \t\n\r")}
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

// TODO: Format - Requires ArrayValue to be moved to runtime package
// Original signature: func (i *Interpreter) builtinFormat(args []Value) Value
// This function expects an array of values as the second argument.

// TODO: Insert - Requires special handling (takes []ast.Expression, modifies variable in-place)
// Original signature: func (i *Interpreter) builtinInsert(args []ast.Expression) Value
// This procedure modifies a string variable in-place and requires AST evaluation.

// TODO: DeleteString - Requires special handling (takes []ast.Expression, modifies variable in-place)
// Original signature: func (i *Interpreter) builtinDeleteString(args []ast.Expression) Value
// This procedure modifies a string variable in-place and requires AST evaluation.

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

	// Convert to rune and then to string
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

	// Check for true values
	switch s {
	case "true", "1", "yes", "t", "y":
		return &runtime.BooleanValue{Value: true}
	case "false", "0", "no", "f", "n":
		return &runtime.BooleanValue{Value: false}
	default:
		return ctx.NewError("StrToBool() invalid boolean string: '%s'", strVal.Value)
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

	// Calculate length from start and end positions
	// SubString(str, 3, 7) should return 5 characters (positions 3, 4, 5, 6, 7)
	length := end - start + 1

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

	// Handle empty needle - returns length + 1 (not found behavior for DWScript)
	if len(needle) == 0 {
		return &runtime.IntegerValue{Value: int64(runeLength(haystack) + 1)}
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
// StrFind(str, substr, fromIndex) - returns 1-based position (0 if not found)
func StrFind(ctx Context, args []Value) Value {
	if len(args) != 3 {
		return ctx.NewError("StrFind() expects exactly 3 arguments, got %d", len(args))
	}

	// StrFind(str, substr, fromIndex) maps to PosEx(substr, str, fromIndex)
	// Need to reorder arguments
	reorderedArgs := []Value{
		args[1], // substr becomes first arg (needle)
		args[0], // str becomes second arg (haystack)
		args[2], // fromIndex stays as third arg (offset)
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

	if absSize < KB {
		result = fmt.Sprintf("%d bytes", int64(size))
	} else if absSize < MB {
		result = fmt.Sprintf("%.2f KB", size/KB)
	} else if absSize < GB {
		result = fmt.Sprintf("%.2f MB", size/MB)
	} else if absSize < TB {
		result = fmt.Sprintf("%.2f GB", size/GB)
	} else {
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

// TODO: StrSplit - Requires ArrayValue to be moved to runtime package
// Original signature: func (i *Interpreter) builtinStrSplit(args []Value) Value
// This function returns an ArrayValue.

// TODO: StrJoin - Requires ArrayValue to be moved to runtime package
// Original signature: func (i *Interpreter) builtinStrJoin(args []Value) Value
// This function takes an ArrayValue as the first argument.

// TODO: StrArrayPack - Requires ArrayValue to be moved to runtime package
// Original signature: func (i *Interpreter) builtinStrArrayPack(args []Value) Value
// This function takes and returns an ArrayValue.

// StrBefore implements the StrBefore() built-in function.
// It returns the substring before the first occurrence of a delimiter.
// StrBefore(str, delimiter) - returns substring before first delimiter (empty if not found)
func StrBefore(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("StrBefore() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrBefore() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: delimiter
	delimVal, ok := args[1].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrBefore() expects string as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	delim := delimVal.Value

	// Handle empty delimiter - return empty string
	if len(delim) == 0 {
		return &runtime.StringValue{Value: ""}
	}

	// Find the first occurrence of delimiter
	index := strings.Index(str, delim)
	if index == -1 {
		// Delimiter not found - return empty string
		return &runtime.StringValue{Value: ""}
	}

	// Return substring before delimiter
	return &runtime.StringValue{Value: str[:index]}
}

// StrBeforeLast implements the StrBeforeLast() built-in function.
// It returns the substring before the last occurrence of a delimiter.
// StrBeforeLast(str, delimiter) - returns substring before last delimiter (empty if not found)
func StrBeforeLast(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("StrBeforeLast() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrBeforeLast() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: delimiter
	delimVal, ok := args[1].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrBeforeLast() expects string as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	delim := delimVal.Value

	// Handle empty delimiter - return empty string
	if len(delim) == 0 {
		return &runtime.StringValue{Value: ""}
	}

	// Find the last occurrence of delimiter
	index := strings.LastIndex(str, delim)
	if index == -1 {
		// Delimiter not found - return empty string
		return &runtime.StringValue{Value: ""}
	}

	// Return substring before last delimiter
	return &runtime.StringValue{Value: str[:index]}
}

// StrAfter implements the StrAfter() built-in function.
// It returns the substring after the first occurrence of a delimiter.
// StrAfter(str, delimiter) - returns substring after first delimiter (empty if not found)
func StrAfter(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("StrAfter() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrAfter() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: delimiter
	delimVal, ok := args[1].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrAfter() expects string as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	delim := delimVal.Value

	// Handle empty delimiter - return empty string
	if len(delim) == 0 {
		return &runtime.StringValue{Value: ""}
	}

	// Find the first occurrence of delimiter
	index := strings.Index(str, delim)
	if index == -1 {
		// Delimiter not found - return empty string
		return &runtime.StringValue{Value: ""}
	}

	// Return substring after delimiter
	return &runtime.StringValue{Value: str[index+len(delim):]}
}

// StrAfterLast implements the StrAfterLast() built-in function.
// It returns the substring after the last occurrence of a delimiter.
// StrAfterLast(str, delimiter) - returns substring after last delimiter (empty if not found)
func StrAfterLast(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("StrAfterLast() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrAfterLast() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: delimiter
	delimVal, ok := args[1].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrAfterLast() expects string as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	delim := delimVal.Value

	// Handle empty delimiter - return empty string
	if len(delim) == 0 {
		return &runtime.StringValue{Value: ""}
	}

	// Find the last occurrence of delimiter
	index := strings.LastIndex(str, delim)
	if index == -1 {
		// Delimiter not found - return empty string
		return &runtime.StringValue{Value: ""}
	}

	// Return substring after last delimiter
	return &runtime.StringValue{Value: str[index+len(delim):]}
}

// StrBetween implements the StrBetween() built-in function.
// It returns the substring between first occurrence of start and first occurrence of stop after start.
// StrBetween(str, start, stop) - returns substring between start and stop delimiters
func StrBetween(ctx Context, args []Value) Value {
	if len(args) != 3 {
		return ctx.NewError("StrBetween() expects exactly 3 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrBetween() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: start delimiter
	startVal, ok := args[1].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrBetween() expects string as second argument, got %s", args[1].Type())
	}

	// Third argument: stop delimiter
	stopVal, ok := args[2].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrBetween() expects string as third argument, got %s", args[2].Type())
	}

	str := strVal.Value
	start := startVal.Value
	stop := stopVal.Value

	// Handle empty delimiters - return empty string
	if len(start) == 0 || len(stop) == 0 {
		return &runtime.StringValue{Value: ""}
	}

	// Find the first occurrence of start delimiter
	startIdx := strings.Index(str, start)
	if startIdx == -1 {
		// Start delimiter not found - return empty string
		return &runtime.StringValue{Value: ""}
	}

	// Search for stop delimiter after the start delimiter
	searchFrom := startIdx + len(start)
	if searchFrom >= len(str) {
		// No room for stop delimiter - return empty string
		return &runtime.StringValue{Value: ""}
	}

	stopIdx := strings.Index(str[searchFrom:], stop)
	if stopIdx == -1 {
		// Stop delimiter not found - return empty string
		return &runtime.StringValue{Value: ""}
	}

	// Adjust stopIdx to be relative to the original string
	stopIdx += searchFrom

	// Return substring between start and stop delimiters
	return &runtime.StringValue{Value: str[searchFrom:stopIdx]}
}

// IsDelimiter implements the IsDelimiter() built-in function.
// It checks if the character at a given position is one of the specified delimiters.
// IsDelimiter(delims, str, index) - returns true if char at index is a delimiter (1-based index)
func IsDelimiter(ctx Context, args []Value) Value {
	if len(args) != 3 {
		return ctx.NewError("IsDelimiter() expects exactly 3 arguments, got %d", len(args))
	}

	// First argument: delimiter characters
	delimsVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("IsDelimiter() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string to check
	strVal, ok := args[1].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("IsDelimiter() expects string as second argument, got %s", args[1].Type())
	}

	// Third argument: index (1-based)
	indexVal, ok := args[2].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("IsDelimiter() expects integer as third argument, got %s", args[2].Type())
	}

	delims := delimsVal.Value
	str := strVal.Value
	index := int(indexVal.Value) // 1-based

	// Handle invalid index
	if index < 1 {
		return &runtime.BooleanValue{Value: false}
	}

	// Convert to rune-based indexing for UTF-8 support
	strRunes := []rune(str)

	// Check if index is within bounds (1-based)
	if index > len(strRunes) {
		return &runtime.BooleanValue{Value: false}
	}

	// Get the character at the specified position (convert to 0-based)
	ch := strRunes[index-1]

	// Check if the character is in the delimiter string
	result := strings.ContainsRune(delims, ch)
	return &runtime.BooleanValue{Value: result}
}

// LastDelimiter implements the LastDelimiter() built-in function.
// It finds the position of the last occurrence of any delimiter character.
// LastDelimiter(delims, str) - returns 1-based position of last delimiter (0 if not found)
func LastDelimiter(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("LastDelimiter() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: delimiter characters
	delimsVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("LastDelimiter() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string to search
	strVal, ok := args[1].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("LastDelimiter() expects string as second argument, got %s", args[1].Type())
	}

	delims := delimsVal.Value
	str := strVal.Value

	// Convert to rune-based for UTF-8 support
	strRunes := []rune(str)

	// Search from the end for any delimiter character
	for i := len(strRunes) - 1; i >= 0; i-- {
		if strings.ContainsRune(delims, strRunes[i]) {
			// Return 1-based position
			return &runtime.IntegerValue{Value: int64(i + 1)}
		}
	}

	// No delimiter found
	return &runtime.IntegerValue{Value: 0}
}

// FindDelimiter implements the FindDelimiter() built-in function.
// It finds the position of the first occurrence of any delimiter character, starting from an index.
// FindDelimiter(delims, str, startIndex) - returns 1-based position (0 if not found)
func FindDelimiter(ctx Context, args []Value) Value {
	if len(args) != 3 {
		return ctx.NewError("FindDelimiter() expects exactly 3 arguments, got %d", len(args))
	}

	// First argument: delimiter characters
	delimsVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("FindDelimiter() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string to search
	strVal, ok := args[1].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("FindDelimiter() expects string as second argument, got %s", args[1].Type())
	}

	// Third argument: start index (1-based)
	startIndexVal, ok := args[2].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("FindDelimiter() expects integer as third argument, got %s", args[2].Type())
	}

	delims := delimsVal.Value
	str := strVal.Value
	startIndex := int(startIndexVal.Value) // 1-based

	// Handle invalid start index
	if startIndex < 1 {
		return &runtime.IntegerValue{Value: 0}
	}

	// Convert to rune-based for UTF-8 support
	strRunes := []rune(str)

	// Adjust to 0-based index
	startIdx := startIndex - 1

	// Check if start index is within bounds
	if startIdx >= len(strRunes) {
		return &runtime.IntegerValue{Value: 0}
	}

	// Search from startIdx for any delimiter character
	for i := startIdx; i < len(strRunes); i++ {
		if strings.ContainsRune(delims, strRunes[i]) {
			// Return 1-based position
			return &runtime.IntegerValue{Value: int64(i + 1)}
		}
	}

	// No delimiter found
	return &runtime.IntegerValue{Value: 0}
}

// PadLeft implements the PadLeft() built-in function.
// It pads a string on the left to reach a minimum width.
// PadLeft(str, count) - pads with spaces
// PadLeft(str, count, char) - pads with specified character
func PadLeft(ctx Context, args []Value) Value {
	if len(args) < 2 || len(args) > 3 {
		return ctx.NewError("PadLeft() expects 2 or 3 arguments, got %d", len(args))
	}

	// First argument: string to pad
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("PadLeft() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: count (minimum width)
	countVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("PadLeft() expects integer as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	count := int(countVal.Value)

	// Default padding character is space
	padChar := " "

	// Optional third argument: padding character
	if len(args) == 3 {
		charVal, ok := args[2].(*runtime.StringValue)
		if !ok {
			return ctx.NewError("PadLeft() expects string as third argument, got %s", args[2].Type())
		}
		// Get the first character (rune-based to handle UTF-8)
		if runeLength(charVal.Value) > 0 {
			firstRune, _ := runeAt(charVal.Value, 1)
			padChar = string(firstRune)
		}
	}

	// Get the current length in runes
	strLen := runeLength(str)

	// If string is already at or beyond the desired length, return as-is
	if strLen >= count {
		return &runtime.StringValue{Value: str}
	}

	// Pad on the left
	padding := strings.Repeat(padChar, count-strLen)
	return &runtime.StringValue{Value: padding + str}
}

// PadRight implements the PadRight() built-in function.
// It pads a string on the right to reach a minimum width.
// PadRight(str, count) - pads with spaces
// PadRight(str, count, char) - pads with specified character
func PadRight(ctx Context, args []Value) Value {
	if len(args) < 2 || len(args) > 3 {
		return ctx.NewError("PadRight() expects 2 or 3 arguments, got %d", len(args))
	}

	// First argument: string to pad
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("PadRight() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: count (minimum width)
	countVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("PadRight() expects integer as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	count := int(countVal.Value)

	// Default padding character is space
	padChar := " "

	// Optional third argument: padding character
	if len(args) == 3 {
		charVal, ok := args[2].(*runtime.StringValue)
		if !ok {
			return ctx.NewError("PadRight() expects string as third argument, got %s", args[2].Type())
		}
		// Get the first character (rune-based to handle UTF-8)
		if runeLength(charVal.Value) > 0 {
			firstRune, _ := runeAt(charVal.Value, 1)
			padChar = string(firstRune)
		}
	}

	// Get the current length in runes
	strLen := runeLength(str)

	// If string is already at or beyond the desired length, return as-is
	if strLen >= count {
		return &runtime.StringValue{Value: str}
	}

	// Pad on the right
	padding := strings.Repeat(padChar, count-strLen)
	return &runtime.StringValue{Value: str + padding}
}

// StrDeleteLeft implements the StrDeleteLeft() built-in function.
// It deletes N leftmost characters from a string.
// StrDeleteLeft(str, count) - removes first count characters
func StrDeleteLeft(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("StrDeleteLeft() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrDeleteLeft() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: count
	countVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("StrDeleteLeft() expects integer as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	count := int(countVal.Value)

	// Handle edge cases
	if count <= 0 {
		return &runtime.StringValue{Value: str}
	}

	// Convert to rune-based for UTF-8 support
	strRunes := []rune(str)
	strLen := len(strRunes)

	// If count >= length, return empty string
	if count >= strLen {
		return &runtime.StringValue{Value: ""}
	}

	// Return substring from count to end
	return &runtime.StringValue{Value: string(strRunes[count:])}
}

// StrDeleteRight implements the StrDeleteRight() built-in function.
// It deletes N rightmost characters from a string.
// StrDeleteRight(str, count) - removes last count characters
func StrDeleteRight(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("StrDeleteRight() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrDeleteRight() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: count
	countVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("StrDeleteRight() expects integer as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	count := int(countVal.Value)

	// Handle edge cases
	if count <= 0 {
		return &runtime.StringValue{Value: str}
	}

	// Convert to rune-based for UTF-8 support
	strRunes := []rune(str)
	strLen := len(strRunes)

	// If count >= length, return empty string
	if count >= strLen {
		return &runtime.StringValue{Value: ""}
	}

	// Return substring from start to (length - count)
	return &runtime.StringValue{Value: string(strRunes[:strLen-count])}
}

// ReverseString implements the ReverseString() built-in function.
// It reverses the character order of a string.
// ReverseString(str) - returns reversed string
func ReverseString(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("ReverseString() expects exactly 1 argument, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("ReverseString() expects string as argument, got %s", args[0].Type())
	}

	str := strVal.Value

	// Convert to rune-based for UTF-8 support
	strRunes := []rune(str)

	// Reverse the runes
	for i, j := 0, len(strRunes)-1; i < j; i, j = i+1, j-1 {
		strRunes[i], strRunes[j] = strRunes[j], strRunes[i]
	}

	return &runtime.StringValue{Value: string(strRunes)}
}

// QuotedStr implements the QuotedStr() built-in function.
// It adds quotes around a string, escaping internal quotes by doubling them.
// QuotedStr(str) - uses single quotes (default)
// QuotedStr(str, quoteChar) - uses specified quote character
func QuotedStr(ctx Context, args []Value) Value {
	if len(args) < 1 || len(args) > 2 {
		return ctx.NewError("QuotedStr() expects 1 or 2 arguments, got %d", len(args))
	}

	// First argument: string to quote
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("QuotedStr() expects string as first argument, got %s", args[0].Type())
	}

	str := strVal.Value

	// Default quote character is single quote
	quoteChar := "'"

	// Optional second argument: quote character
	if len(args) == 2 {
		quoteCharVal, ok := args[1].(*runtime.StringValue)
		if !ok {
			return ctx.NewError("QuotedStr() expects string as second argument, got %s", args[1].Type())
		}
		// Get the first character (rune-based to handle UTF-8)
		if runeLength(quoteCharVal.Value) > 0 {
			firstRune, _ := runeAt(quoteCharVal.Value, 1)
			quoteChar = string(firstRune)
		}
	}

	// Escape internal quotes by doubling them
	escaped := strings.ReplaceAll(str, quoteChar, quoteChar+quoteChar)

	// Wrap with quotes
	return &runtime.StringValue{Value: quoteChar + escaped + quoteChar}
}

// StringOfString implements the StringOfString() built-in function.
// It repeats a string N times.
// StringOfString(str, count) - returns string repeated count times
func StringOfString(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("StringOfString() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string to repeat
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StringOfString() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: count (number of repetitions)
	countVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("StringOfString() expects integer as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	count := int(countVal.Value)

	// Handle edge cases
	if count <= 0 {
		return &runtime.StringValue{Value: ""}
	}

	// Repeat the string
	result := strings.Repeat(str, count)
	return &runtime.StringValue{Value: result}
}

// DupeString implements the DupeString() built-in function.
// It is an alias for StringOfString - repeats a string N times.
// DupeString(str, count) - returns string repeated count times
func DupeString(ctx Context, args []Value) Value {
	// DupeString is just an alias for StringOfString
	return StringOfString(ctx, args)
}

// NormalizeString implements the NormalizeString() built-in function.
// It normalizes a string to Unicode Normal Form.
// NormalizeString(str) - normalizes to NFC (default)
// NormalizeString(str, form) - normalizes to specified form (NFC, NFD, NFKC, NFKD)
func NormalizeString(ctx Context, args []Value) Value {
	if len(args) < 1 || len(args) > 2 {
		return ctx.NewError("NormalizeString() expects 1 or 2 arguments, got %d", len(args))
	}

	// First argument: string to normalize
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("NormalizeString() expects string as first argument, got %s", args[0].Type())
	}

	str := strVal.Value

	// Default normalization form is NFC
	form := "NFC"

	// Optional second argument: normalization form
	if len(args) == 2 {
		formVal, ok := args[1].(*runtime.StringValue)
		if !ok {
			return ctx.NewError("NormalizeString() expects string as second argument, got %s", args[1].Type())
		}
		form = strings.ToUpper(formVal.Value)
	}

	// Apply normalization
	result := normalizeUnicode(str, form)
	return &runtime.StringValue{Value: result}
}

// StripAccents implements the StripAccents() built-in function.
// It removes diacritical marks from a string.
// StripAccents(str) - returns string with accents removed
func StripAccents(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("StripAccents() expects exactly 1 argument, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StripAccents() expects string as argument, got %s", args[0].Type())
	}

	str := strVal.Value

	// Strip accents using NFD normalization and then removing combining marks
	result := stripAccents(str)
	return &runtime.StringValue{Value: result}
}

// =============================================================================
// String Comparison Operations
// =============================================================================

// SameText implements the SameText() built-in function.
// It performs case-insensitive string comparison for equality.
// SameText(str1, str2) - returns true if strings are equal ignoring case
func SameText(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("SameText() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string 1
	str1Val, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("SameText() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string 2
	str2Val, ok := args[1].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("SameText() expects string as second argument, got %s", args[1].Type())
	}

	str1 := str1Val.Value
	str2 := str2Val.Value

	// Case-insensitive comparison using Unicode case folding
	result := strings.EqualFold(str1, str2)
	return &runtime.BooleanValue{Value: result}
}

// CompareText implements the CompareText() built-in function.
// It performs case-insensitive string comparison.
// CompareText(str1, str2) - returns -1 if str1 < str2, 0 if equal, 1 if str1 > str2
func CompareText(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("CompareText() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string 1
	str1Val, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("CompareText() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string 2
	str2Val, ok := args[1].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("CompareText() expects string as second argument, got %s", args[1].Type())
	}

	str1 := strings.ToLower(str1Val.Value)
	str2 := strings.ToLower(str2Val.Value)

	// Compare strings
	if str1 < str2 {
		return &runtime.IntegerValue{Value: -1}
	} else if str1 > str2 {
		return &runtime.IntegerValue{Value: 1}
	}
	return &runtime.IntegerValue{Value: 0}
}

// CompareStr implements the CompareStr() built-in function.
// It performs case-sensitive string comparison.
// CompareStr(str1, str2) - returns -1 if str1 < str2, 0 if equal, 1 if str1 > str2
func CompareStr(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("CompareStr() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string 1
	str1Val, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("CompareStr() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string 2
	str2Val, ok := args[1].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("CompareStr() expects string as second argument, got %s", args[1].Type())
	}

	str1 := str1Val.Value
	str2 := str2Val.Value

	// Compare strings
	if str1 < str2 {
		return &runtime.IntegerValue{Value: -1}
	} else if str1 > str2 {
		return &runtime.IntegerValue{Value: 1}
	}
	return &runtime.IntegerValue{Value: 0}
}

// AnsiCompareText implements the AnsiCompareText() built-in function.
// It performs ANSI case-insensitive string comparison.
// AnsiCompareText(str1, str2) - returns -1 if str1 < str2, 0 if equal, 1 if str1 > str2
func AnsiCompareText(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("AnsiCompareText() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string 1
	str1Val, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("AnsiCompareText() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string 2
	str2Val, ok := args[1].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("AnsiCompareText() expects string as second argument, got %s", args[1].Type())
	}

	str1 := strings.ToLower(str1Val.Value)
	str2 := strings.ToLower(str2Val.Value)

	// For ANSI comparison, use simple byte-wise comparison after lowercasing
	if str1 < str2 {
		return &runtime.IntegerValue{Value: -1}
	} else if str1 > str2 {
		return &runtime.IntegerValue{Value: 1}
	}
	return &runtime.IntegerValue{Value: 0}
}

// AnsiCompareStr implements the AnsiCompareStr() built-in function.
// It performs ANSI case-sensitive string comparison.
// AnsiCompareStr(str1, str2) - returns -1 if str1 < str2, 0 if equal, 1 if str1 > str2
func AnsiCompareStr(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("AnsiCompareStr() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string 1
	str1Val, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("AnsiCompareStr() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string 2
	str2Val, ok := args[1].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("AnsiCompareStr() expects string as second argument, got %s", args[1].Type())
	}

	str1 := str1Val.Value
	str2 := str2Val.Value

	// ANSI case-sensitive comparison
	if str1 < str2 {
		return &runtime.IntegerValue{Value: -1}
	} else if str1 > str2 {
		return &runtime.IntegerValue{Value: 1}
	}
	return &runtime.IntegerValue{Value: 0}
}

// CompareLocaleStr implements the CompareLocaleStr() built-in function.
// It performs locale-aware string comparison.
// CompareLocaleStr(str1, str2) - uses default locale, case-insensitive
// CompareLocaleStr(str1, str2, locale) - uses specified locale, case-insensitive
// CompareLocaleStr(str1, str2, locale, caseSensitive) - with case sensitivity control
func CompareLocaleStr(ctx Context, args []Value) Value {
	if len(args) < 2 || len(args) > 4 {
		return ctx.NewError("CompareLocaleStr() expects 2 to 4 arguments, got %d", len(args))
	}

	// First argument: string 1
	str1Val, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("CompareLocaleStr() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string 2
	str2Val, ok := args[1].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("CompareLocaleStr() expects string as second argument, got %s", args[1].Type())
	}

	str1 := str1Val.Value
	str2 := str2Val.Value

	// Default locale is English
	locale := "en"
	caseSensitive := false

	// Optional third argument: locale
	if len(args) >= 3 {
		localeVal, ok := args[2].(*runtime.StringValue)
		if !ok {
			return ctx.NewError("CompareLocaleStr() expects string as third argument, got %s", args[2].Type())
		}
		locale = localeVal.Value
	}

	// Optional fourth argument: case sensitivity
	if len(args) == 4 {
		csVal, ok := args[3].(*runtime.BooleanValue)
		if !ok {
			return ctx.NewError("CompareLocaleStr() expects boolean as fourth argument, got %s", args[3].Type())
		}
		caseSensitive = csVal.Value
	}

	// Parse the locale tag
	tag, err := language.Parse(locale)
	if err != nil {
		// Fall back to English if locale is invalid
		tag = language.English
	}

	// Create collator with appropriate options
	var col *collate.Collator
	if !caseSensitive {
		col = collate.New(tag, collate.IgnoreCase)
	} else {
		col = collate.New(tag)
	}

	// Compare strings
	result := col.CompareString(str1, str2)
	if result < 0 {
		return &runtime.IntegerValue{Value: -1}
	} else if result > 0 {
		return &runtime.IntegerValue{Value: 1}
	}
	return &runtime.IntegerValue{Value: 0}
}

// StrMatches implements the StrMatches() built-in function.
// It performs wildcard pattern matching.
// StrMatches(str, mask) - returns true if string matches pattern (* and ? wildcards)
func StrMatches(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("StrMatches() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string to match
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrMatches() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: pattern/mask
	maskVal, ok := args[1].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrMatches() expects string as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	mask := maskVal.Value

	// Use cross-platform wildcard pattern matching
	// Supports * (zero or more characters) and ? (single character)
	matched := wildcardMatch(str, mask)

	return &runtime.BooleanValue{Value: matched}
}

// StrIsASCII implements the StrIsASCII() built-in function.
// It checks if a string contains only ASCII characters.
// StrIsASCII(str) - returns true if string is pure ASCII (all chars < 128)
func StrIsASCII(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("StrIsASCII() expects exactly 1 argument, got %d", len(args))
	}

	// First argument: string to check
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrIsASCII() expects string as argument, got %s", args[0].Type())
	}

	str := strVal.Value

	// Check if all characters are ASCII (< 128)
	for _, r := range str {
		if r > unicode.MaxASCII {
			return &runtime.BooleanValue{Value: false}
		}
	}

	return &runtime.BooleanValue{Value: true}
}

// =============================================================================
// Helper Functions
// =============================================================================

// runeLength returns the number of Unicode characters (runes) in a string,
// not the byte length. This is important for UTF-8 strings where characters
// can be multiple bytes.
func runeLength(s string) int {
	return utf8.RuneCountInString(s)
}

// runeAt returns the rune at the given 1-based index in the string.
// Returns the rune and true if the index is valid, or 0 and false otherwise.
// DWScript uses 1-based indexing, so index 1 returns the first character.
func runeAt(s string, index int) (rune, bool) {
	if index < 1 {
		return 0, false
	}

	runes := []rune(s)
	if index > len(runes) {
		return 0, false
	}

	return runes[index-1], true
}

// runeSliceFrom returns a substring starting from a 1-based position and taking count characters.
// This is commonly used in DWScript's Copy function: Copy(str, start, count)
func runeSliceFrom(s string, start, count int) string {
	if start < 1 || count <= 0 {
		return ""
	}

	runes := []rune(s)
	length := len(runes)

	startIdx := start - 1 // Convert to 0-based
	if startIdx >= length {
		return ""
	}

	endIdx := startIdx + count
	if endIdx > length {
		endIdx = length
	}

	return string(runes[startIdx:endIdx])
}

// normalizeUnicode normalizes a string to the specified Unicode normalization form.
// Supported forms: NFC, NFD, NFKC, NFKD
func normalizeUnicode(s string, form string) string {
	switch form {
	case "NFC":
		return norm.NFC.String(s)
	case "NFD":
		return norm.NFD.String(s)
	case "NFKC":
		return norm.NFKC.String(s)
	case "NFKD":
		return norm.NFKD.String(s)
	default:
		// Default to NFC if form is unknown
		return norm.NFC.String(s)
	}
}

// stripAccents removes diacritical marks from a string.
// It works by normalizing the string to NFD (decomposed form) and then
// removing all combining marks (which include accents).
func stripAccents(s string) string {
	// Normalize to NFD (decomposed form)
	normalized := norm.NFD.String(s)

	// Filter out combining marks
	var result []rune
	for _, r := range normalized {
		if !unicode.Is(unicode.Mn, r) {
			result = append(result, r)
		}
	}

	return string(result)
}

// wildcardMatch performs wildcard pattern matching.
// Supports * (zero or more characters) and ? (single character).
// This is a cross-platform implementation that doesn't have the
// path separator issues of filepath.Match.
func wildcardMatch(str, pattern string) bool {
	return wildcardMatchImpl([]rune(str), []rune(pattern), 0, 0)
}

// wildcardMatchImpl is the recursive implementation of wildcard matching.
func wildcardMatchImpl(str, pattern []rune, si, pi int) bool {
	// Both string and pattern exhausted - match
	if si == len(str) && pi == len(pattern) {
		return true
	}

	// Pattern exhausted but string not - no match
	if pi == len(pattern) {
		return false
	}

	// Handle * wildcard
	if pattern[pi] == '*' {
		// Skip consecutive *
		for pi < len(pattern) && pattern[pi] == '*' {
			pi++
		}
		// * at end matches everything
		if pi == len(pattern) {
			return true
		}
		// Try matching zero or more characters
		for si <= len(str) {
			if wildcardMatchImpl(str, pattern, si, pi) {
				return true
			}
			si++
		}
		return false
	}

	// String exhausted but pattern has non-* characters - no match
	if si == len(str) {
		return false
	}

	// Handle ? wildcard or exact character match
	if pattern[pi] == '?' || pattern[pi] == str[si] {
		return wildcardMatchImpl(str, pattern, si+1, pi+1)
	}

	// No match
	return false
}
