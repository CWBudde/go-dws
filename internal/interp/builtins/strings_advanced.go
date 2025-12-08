package builtins

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// StrBefore implements the StrBefore() built-in function.
// It returns the substring before the first occurrence of a delimiter.
// StrBefore(str, delimiter) - returns full string if delimiter not found
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

	// Handle empty delimiter - return the full string
	if len(delim) == 0 {
		return &runtime.StringValue{Value: str}
	}

	// Find the first occurrence of delimiter
	index := strings.Index(str, delim)
	if index == -1 {
		// Delimiter not found - return the full string
		return &runtime.StringValue{Value: str}
	}

	// Return substring before delimiter
	return &runtime.StringValue{Value: str[:index]}
}

// StrBeforeLast implements the StrBeforeLast() built-in function.
// It returns the substring before the last occurrence of a delimiter.
// StrBeforeLast(str, delimiter) - returns full string if delimiter not found
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

	// Handle empty delimiter - return the full string
	if len(delim) == 0 {
		return &runtime.StringValue{Value: str}
	}

	// Find the last occurrence of delimiter
	index := strings.LastIndex(str, delim)
	if index == -1 {
		// Delimiter not found - return the full string
		return &runtime.StringValue{Value: str}
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

	// Handle empty delimiters following DWScript semantics:
	// - Empty start -> not found
	// - Empty stop  -> return everything after the start delimiter (if present)
	if len(start) == 0 {
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

	// Empty stop delimiter means "to the end"
	stopIdx := -1
	if len(stop) > 0 {
		stopIdx = strings.Index(str[searchFrom:], stop)
	}
	if stopIdx == -1 {
		// Stop delimiter not found - return substring from start to end
		return &runtime.StringValue{Value: str[searchFrom:]}
	}

	// Adjust stopIdx to be relative to the original string
	stopIdx += searchFrom

	// Return substring between start and stop delimiters
	return &runtime.StringValue{Value: str[searchFrom:stopIdx]}
}

// StrReplaceMacros replaces macros delimited by start/end tokens using a map of replacements.
// StrReplaceMacros(str, macrosArray, startDelim [, endDelim])
// macrosArray is an array of strings in pairs: [key1, val1, key2, val2, ...]
func StrReplaceMacros(ctx Context, args []Value) Value {
	if len(args) < 2 || len(args) > 4 {
		return ctx.NewError("StrReplaceMacros() expects 2 to 4 arguments, got %d", len(args))
	}

	textVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrReplaceMacros() expects string as first argument, got %s", args[0].Type())
	}
	macroArr, ok := args[1].(*runtime.ArrayValue)
	if !ok {
		return ctx.NewError("StrReplaceMacros() expects array as second argument, got %s", args[1].Type())
	}

	startDelim := ""
	endDelim := ""
	if len(args) >= 3 {
		if s, ok := args[2].(*runtime.StringValue); ok {
			startDelim = s.Value
		} else {
			return ctx.NewError("StrReplaceMacros() expects string as third argument, got %s", args[2].Type())
		}
	}
	if len(args) == 4 {
		if s, ok := args[3].(*runtime.StringValue); ok {
			endDelim = s.Value
		} else {
			return ctx.NewError("StrReplaceMacros() expects string as fourth argument, got %s", args[3].Type())
		}
	}

	macros := []string{}
	for _, elem := range macroArr.Elements {
		if sv, ok := elem.(*runtime.StringValue); ok {
			macros = append(macros, sv.Value)
		} else {
			return ctx.NewError("StrReplaceMacros() expects array of strings for macros")
		}
	}

	text := textVal.Value
	// Replace using pairs
	for i := 0; i+1 < len(macros); i += 2 {
		key := macros[i]
		val := macros[i+1]

		var pattern string
		switch {
		case startDelim != "" && endDelim != "":
			pattern = startDelim + key + endDelim
		case startDelim != "":
			pattern = startDelim + key + startDelim
		default:
			pattern = key
		}

		if pattern == "" {
			// Avoid infinite loops when pattern is empty
			continue
		}

		text = strings.ReplaceAll(text, pattern, val)
	}

	return &runtime.StringValue{Value: text}
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
// FindDelimiter(delims, str, startIndex) - returns 1-based position (-1 if not found)
func FindDelimiter(ctx Context, args []Value) Value {
	if len(args) != 2 && len(args) != 3 {
		return ctx.NewError("FindDelimiter() expects 2 or 3 arguments, got %d", len(args))
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
	startIndex := 1
	if len(args) == 3 {
		startIndexVal, ok := ctx.ToInt64(args[2])
		if !ok {
			return ctx.NewError("FindDelimiter() expects integer as third argument, got %s", args[2].Type())
		}
		startIndex = int(startIndexVal)
	}

	delims := delimsVal.Value
	str := strVal.Value
	// Handle invalid start index
	if startIndex < 1 {
		return &runtime.IntegerValue{Value: -1}
	}

	// Convert to rune-based for UTF-8 support
	strRunes := []rune(str)

	// Adjust to 0-based index
	startIdx := startIndex - 1

	// Check if start index is within bounds
	if startIdx >= len(strRunes) {
		return &runtime.IntegerValue{Value: -1}
	}

	// Search from startIdx for any delimiter character
	for i := startIdx; i < len(strRunes); i++ {
		if strings.ContainsRune(delims, strRunes[i]) {
			// Return 1-based position
			return &runtime.IntegerValue{Value: int64(i + 1)}
		}
	}

	// No delimiter found - return -1
	return &runtime.IntegerValue{Value: -1}
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
