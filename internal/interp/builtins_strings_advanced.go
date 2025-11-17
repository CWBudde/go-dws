package interp

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
)

// builtinStrSplit implements the StrSplit() built-in function.
// It splits a string into an array of strings using a delimiter.
// If the delimiter is empty, returns an array with the original string as the sole element.
func (i *Interpreter) builtinStrSplit(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "StrSplit() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string to split
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrSplit() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: delimiter
	delimVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrSplit() expects string as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	delim := delimVal.Value

	// Handle empty delimiter - return array with single element (the original string)
	if len(delim) == 0 {
		elements := []Value{&StringValue{Value: str}}
		return &ArrayValue{
			Elements:  elements,
			ArrayType: types.NewDynamicArrayType(types.STRING),
		}
	}

	// Split the string
	parts := strings.Split(str, delim)

	// Convert to array of StringValue
	elements := make([]Value, len(parts))
	for idx, part := range parts {
		elements[idx] = &StringValue{Value: part}
	}

	return &ArrayValue{
		Elements:  elements,
		ArrayType: types.NewDynamicArrayType(types.STRING),
	}
}

// builtinStrJoin implements the StrJoin() built-in function.
// It joins an array of strings into a single string using a delimiter.
// StrJoin(array, delimiter) - returns joined string
func (i *Interpreter) builtinStrJoin(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "StrJoin() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: array of strings
	arrVal, ok := args[0].(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrJoin() expects array as first argument, got %s", args[0].Type())
	}

	// Second argument: delimiter
	delimVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrJoin() expects string as second argument, got %s", args[1].Type())
	}

	delim := delimVal.Value

	// Convert array elements to strings
	parts := make([]string, len(arrVal.Elements))
	for idx, elem := range arrVal.Elements {
		strElem, ok := elem.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "StrJoin() expects array of strings, got %s at index %d", elem.Type(), idx)
		}
		parts[idx] = strElem.Value
	}

	// Join the strings
	result := strings.Join(parts, delim)
	return &StringValue{Value: result}
}

// builtinStrArrayPack implements the StrArrayPack() built-in function.
// It removes empty strings from an array.
// StrArrayPack(array) - returns array with empty strings removed
func (i *Interpreter) builtinStrArrayPack(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "StrArrayPack() expects exactly 1 argument, got %d", len(args))
	}

	// First argument: array of strings
	arrVal, ok := args[0].(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrArrayPack() expects array as argument, got %s", args[0].Type())
	}

	// Filter out empty strings
	var packed []Value
	for _, elem := range arrVal.Elements {
		strElem, ok := elem.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "StrArrayPack() expects array of strings, got %s", elem.Type())
		}
		if strElem.Value != "" {
			packed = append(packed, strElem)
		}
	}

	return &ArrayValue{
		Elements:  packed,
		ArrayType: types.NewDynamicArrayType(types.STRING),
	}
}

// builtinStrBefore implements the StrBefore() built-in function.
// It returns the substring before the first occurrence of a delimiter.
// StrBefore(str, delimiter) - returns substring before first delimiter (empty if not found)
func (i *Interpreter) builtinStrBefore(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "StrBefore() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrBefore() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: delimiter
	delimVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrBefore() expects string as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	delim := delimVal.Value

	// Handle empty delimiter - return the full string
	if len(delim) == 0 {
		return &StringValue{Value: str}
	}

	// Find the first occurrence of delimiter
	index := strings.Index(str, delim)
	if index == -1 {
		// Delimiter not found - return the full string
		return &StringValue{Value: str}
	}

	// Return substring before delimiter
	return &StringValue{Value: str[:index]}
}

// builtinStrBeforeLast implements the StrBeforeLast() built-in function.
// It returns the substring before the last occurrence of a delimiter.
// StrBeforeLast(str, delimiter) - returns substring before last delimiter (empty if not found)
func (i *Interpreter) builtinStrBeforeLast(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "StrBeforeLast() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrBeforeLast() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: delimiter
	delimVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrBeforeLast() expects string as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	delim := delimVal.Value

	// Handle empty delimiter - return the full string
	if len(delim) == 0 {
		return &StringValue{Value: str}
	}

	// Find the last occurrence of delimiter
	index := strings.LastIndex(str, delim)
	if index == -1 {
		// Delimiter not found - return the full string
		return &StringValue{Value: str}
	}

	// Return substring before last delimiter
	return &StringValue{Value: str[:index]}
}

// builtinStrAfter implements the StrAfter() built-in function.
// It returns the substring after the first occurrence of a delimiter.
// StrAfter(str, delimiter) - returns substring after first delimiter (empty if not found)
func (i *Interpreter) builtinStrAfter(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "StrAfter() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrAfter() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: delimiter
	delimVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrAfter() expects string as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	delim := delimVal.Value

	// Handle empty delimiter - return empty string
	if len(delim) == 0 {
		return &StringValue{Value: ""}
	}

	// Find the first occurrence of delimiter
	index := strings.Index(str, delim)
	if index == -1 {
		// Delimiter not found - return empty string
		return &StringValue{Value: ""}
	}

	// Return substring after delimiter
	return &StringValue{Value: str[index+len(delim):]}
}

// builtinStrAfterLast implements the StrAfterLast() built-in function.
// It returns the substring after the last occurrence of a delimiter.
// StrAfterLast(str, delimiter) - returns substring after last delimiter (empty if not found)
func (i *Interpreter) builtinStrAfterLast(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "StrAfterLast() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrAfterLast() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: delimiter
	delimVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrAfterLast() expects string as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	delim := delimVal.Value

	// Handle empty delimiter - return empty string
	if len(delim) == 0 {
		return &StringValue{Value: ""}
	}

	// Find the last occurrence of delimiter
	index := strings.LastIndex(str, delim)
	if index == -1 {
		// Delimiter not found - return empty string
		return &StringValue{Value: ""}
	}

	// Return substring after last delimiter
	return &StringValue{Value: str[index+len(delim):]}
}

// builtinStrBetween implements the StrBetween() built-in function.
// It returns the substring between first occurrence of start and first occurrence of stop after start.
// StrBetween(str, start, stop) - returns substring between start and stop delimiters
func (i *Interpreter) builtinStrBetween(args []Value) Value {
	if len(args) != 3 {
		return i.newErrorWithLocation(i.currentNode, "StrBetween() expects exactly 3 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrBetween() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: start delimiter
	startVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrBetween() expects string as second argument, got %s", args[1].Type())
	}

	// Third argument: stop delimiter
	stopVal, ok := args[2].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrBetween() expects string as third argument, got %s", args[2].Type())
	}

	str := strVal.Value
	start := startVal.Value
	stop := stopVal.Value

	// Handle empty delimiters - return empty string
	if len(start) == 0 || len(stop) == 0 {
		return &StringValue{Value: ""}
	}

	// Find the first occurrence of start delimiter
	startIdx := strings.Index(str, start)
	if startIdx == -1 {
		// Start delimiter not found - return empty string
		return &StringValue{Value: ""}
	}

	// Search for stop delimiter after the start delimiter
	searchFrom := startIdx + len(start)
	if searchFrom >= len(str) {
		// No room for stop delimiter - return empty string
		return &StringValue{Value: ""}
	}

	stopIdx := strings.Index(str[searchFrom:], stop)
	if stopIdx == -1 {
		// Stop delimiter not found - return empty string
		return &StringValue{Value: ""}
	}

	// Adjust stopIdx to be relative to the original string
	stopIdx += searchFrom

	// Return substring between start and stop delimiters
	return &StringValue{Value: str[searchFrom:stopIdx]}
}

// builtinIsDelimiter implements the IsDelimiter() built-in function.
// It checks if the character at a given position is one of the specified delimiters.
// IsDelimiter(delims, str, index) - returns true if char at index is a delimiter (1-based index)
func (i *Interpreter) builtinIsDelimiter(args []Value) Value {
	if len(args) != 3 {
		return i.newErrorWithLocation(i.currentNode, "IsDelimiter() expects exactly 3 arguments, got %d", len(args))
	}

	// First argument: delimiter characters
	delimsVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "IsDelimiter() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string to check
	strVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "IsDelimiter() expects string as second argument, got %s", args[1].Type())
	}

	// Third argument: index (1-based)
	indexVal, ok := args[2].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "IsDelimiter() expects integer as third argument, got %s", args[2].Type())
	}

	delims := delimsVal.Value
	str := strVal.Value
	index := int(indexVal.Value) // 1-based

	// Handle invalid index
	if index < 1 {
		return &BooleanValue{Value: false}
	}

	// Convert to rune-based indexing for UTF-8 support
	strRunes := []rune(str)

	// Check if index is within bounds (1-based)
	if index > len(strRunes) {
		return &BooleanValue{Value: false}
	}

	// Get the character at the specified position (convert to 0-based)
	ch := strRunes[index-1]

	// Check if the character is in the delimiter string
	result := strings.ContainsRune(delims, ch)
	return &BooleanValue{Value: result}
}

// builtinLastDelimiter implements the LastDelimiter() built-in function.
// It finds the position of the last occurrence of any delimiter character.
// LastDelimiter(delims, str) - returns 1-based position of last delimiter (0 if not found)
func (i *Interpreter) builtinLastDelimiter(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "LastDelimiter() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: delimiter characters
	delimsVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "LastDelimiter() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string to search
	strVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "LastDelimiter() expects string as second argument, got %s", args[1].Type())
	}

	delims := delimsVal.Value
	str := strVal.Value

	// Convert to rune-based for UTF-8 support
	strRunes := []rune(str)

	// Search from the end for any delimiter character
	for i := len(strRunes) - 1; i >= 0; i-- {
		if strings.ContainsRune(delims, strRunes[i]) {
			// Return 1-based position
			return &IntegerValue{Value: int64(i + 1)}
		}
	}

	// No delimiter found
	return &IntegerValue{Value: 0}
}

// builtinFindDelimiter implements the FindDelimiter() built-in function.
// It finds the position of the first occurrence of any delimiter character, starting from an index.
// FindDelimiter(delims, str, startIndex) - returns 1-based position (0 if not found)
func (i *Interpreter) builtinFindDelimiter(args []Value) Value {
	if len(args) != 3 {
		return i.newErrorWithLocation(i.currentNode, "FindDelimiter() expects exactly 3 arguments, got %d", len(args))
	}

	// First argument: delimiter characters
	delimsVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "FindDelimiter() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string to search
	strVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "FindDelimiter() expects string as second argument, got %s", args[1].Type())
	}

	// Third argument: start index (1-based)
	startIndexVal, ok := args[2].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "FindDelimiter() expects integer as third argument, got %s", args[2].Type())
	}

	delims := delimsVal.Value
	str := strVal.Value
	startIndex := int(startIndexVal.Value) // 1-based

	// Handle invalid start index
	if startIndex < 1 {
		return &IntegerValue{Value: 0}
	}

	// Convert to rune-based for UTF-8 support
	strRunes := []rune(str)

	// Adjust to 0-based index
	startIdx := startIndex - 1

	// Check if start index is within bounds
	if startIdx >= len(strRunes) {
		return &IntegerValue{Value: 0}
	}

	// Search from startIdx for any delimiter character
	for i := startIdx; i < len(strRunes); i++ {
		if strings.ContainsRune(delims, strRunes[i]) {
			// Return 1-based position
			return &IntegerValue{Value: int64(i + 1)}
		}
	}

	// No delimiter found
	return &IntegerValue{Value: 0}
}

// builtinPadLeft implements the PadLeft() built-in function.
// It pads a string on the left to reach a minimum width.
// PadLeft(str, count) - pads with spaces
// PadLeft(str, count, char) - pads with specified character
func (i *Interpreter) builtinPadLeft(args []Value) Value {
	if len(args) < 2 || len(args) > 3 {
		return i.newErrorWithLocation(i.currentNode, "PadLeft() expects 2 or 3 arguments, got %d", len(args))
	}

	// First argument: string to pad
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "PadLeft() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: count (minimum width)
	countVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "PadLeft() expects integer as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	count := int(countVal.Value)

	// Default padding character is space
	padChar := " "

	// Optional third argument: padding character
	if len(args) == 3 {
		charVal, ok := args[2].(*StringValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "PadLeft() expects string as third argument, got %s", args[2].Type())
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
		return &StringValue{Value: str}
	}

	// Pad on the left
	padding := strings.Repeat(padChar, count-strLen)
	return &StringValue{Value: padding + str}
}

// builtinPadRight implements the PadRight() built-in function.
// It pads a string on the right to reach a minimum width.
// PadRight(str, count) - pads with spaces
// PadRight(str, count, char) - pads with specified character
func (i *Interpreter) builtinPadRight(args []Value) Value {
	if len(args) < 2 || len(args) > 3 {
		return i.newErrorWithLocation(i.currentNode, "PadRight() expects 2 or 3 arguments, got %d", len(args))
	}

	// First argument: string to pad
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "PadRight() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: count (minimum width)
	countVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "PadRight() expects integer as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	count := int(countVal.Value)

	// Default padding character is space
	padChar := " "

	// Optional third argument: padding character
	if len(args) == 3 {
		charVal, ok := args[2].(*StringValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "PadRight() expects string as third argument, got %s", args[2].Type())
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
		return &StringValue{Value: str}
	}

	// Pad on the right
	padding := strings.Repeat(padChar, count-strLen)
	return &StringValue{Value: str + padding}
}

// builtinStrDeleteLeft implements the StrDeleteLeft() built-in function.
// It deletes N leftmost characters from a string.
// StrDeleteLeft(str, count) - removes first count characters
func (i *Interpreter) builtinStrDeleteLeft(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "StrDeleteLeft() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrDeleteLeft() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: count
	countVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrDeleteLeft() expects integer as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	count := int(countVal.Value)

	// Handle edge cases
	if count <= 0 {
		return &StringValue{Value: str}
	}

	// Convert to rune-based for UTF-8 support
	strRunes := []rune(str)
	strLen := len(strRunes)

	// If count >= length, return empty string
	if count >= strLen {
		return &StringValue{Value: ""}
	}

	// Return substring from count to end
	return &StringValue{Value: string(strRunes[count:])}
}

// builtinStrDeleteRight implements the StrDeleteRight() built-in function.
// It deletes N rightmost characters from a string.
// StrDeleteRight(str, count) - removes last count characters
func (i *Interpreter) builtinStrDeleteRight(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "StrDeleteRight() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrDeleteRight() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: count
	countVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrDeleteRight() expects integer as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	count := int(countVal.Value)

	// Handle edge cases
	if count <= 0 {
		return &StringValue{Value: str}
	}

	// Convert to rune-based for UTF-8 support
	strRunes := []rune(str)
	strLen := len(strRunes)

	// If count >= length, return empty string
	if count >= strLen {
		return &StringValue{Value: ""}
	}

	// Return substring from start to (length - count)
	return &StringValue{Value: string(strRunes[:strLen-count])}
}

// builtinReverseString implements the ReverseString() built-in function.
// It reverses the character order of a string.
// ReverseString(str) - returns reversed string
func (i *Interpreter) builtinReverseString(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "ReverseString() expects exactly 1 argument, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "ReverseString() expects string as argument, got %s", args[0].Type())
	}

	str := strVal.Value

	// Convert to rune-based for UTF-8 support
	strRunes := []rune(str)

	// Reverse the runes
	for i, j := 0, len(strRunes)-1; i < j; i, j = i+1, j-1 {
		strRunes[i], strRunes[j] = strRunes[j], strRunes[i]
	}

	return &StringValue{Value: string(strRunes)}
}

// builtinQuotedStr implements the QuotedStr() built-in function.
// It adds quotes around a string, escaping internal quotes by doubling them.
// QuotedStr(str) - uses single quotes (default)
// QuotedStr(str, quoteChar) - uses specified quote character
func (i *Interpreter) builtinQuotedStr(args []Value) Value {
	if len(args) < 1 || len(args) > 2 {
		return i.newErrorWithLocation(i.currentNode, "QuotedStr() expects 1 or 2 arguments, got %d", len(args))
	}

	// First argument: string to quote
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "QuotedStr() expects string as first argument, got %s", args[0].Type())
	}

	str := strVal.Value

	// Default quote character is single quote
	quoteChar := "'"

	// Optional second argument: quote character
	if len(args) == 2 {
		quoteCharVal, ok := args[1].(*StringValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "QuotedStr() expects string as second argument, got %s", args[1].Type())
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
	return &StringValue{Value: quoteChar + escaped + quoteChar}
}

// builtinStringOfString implements the StringOfString() built-in function.
// It repeats a string N times.
// StringOfString(str, count) - returns string repeated count times
func (i *Interpreter) builtinStringOfString(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "StringOfString() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string to repeat
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StringOfString() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: count (number of repetitions)
	countVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StringOfString() expects integer as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	count := int(countVal.Value)

	// Handle edge cases
	if count <= 0 {
		return &StringValue{Value: ""}
	}

	// Repeat the string
	result := strings.Repeat(str, count)
	return &StringValue{Value: result}
}

// builtinDupeString implements the DupeString() built-in function.
// It is an alias for StringOfString - repeats a string N times.
// DupeString(str, count) - returns string repeated count times
func (i *Interpreter) builtinDupeString(args []Value) Value {
	// DupeString is just an alias for StringOfString
	return i.builtinStringOfString(args)
}

// builtinNormalizeString implements the NormalizeString() built-in function.
// It normalizes a string to Unicode Normal Form.
// NormalizeString(str) - normalizes to NFC (default)
// NormalizeString(str, form) - normalizes to specified form (NFC, NFD, NFKC, NFKD)
func (i *Interpreter) builtinNormalizeString(args []Value) Value {
	if len(args) < 1 || len(args) > 2 {
		return i.newErrorWithLocation(i.currentNode, "NormalizeString() expects 1 or 2 arguments, got %d", len(args))
	}

	// First argument: string to normalize
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "NormalizeString() expects string as first argument, got %s", args[0].Type())
	}

	str := strVal.Value

	// Default normalization form is NFC
	form := "NFC"

	// Optional second argument: normalization form
	if len(args) == 2 {
		formVal, ok := args[1].(*StringValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "NormalizeString() expects string as second argument, got %s", args[1].Type())
		}
		form = strings.ToUpper(formVal.Value)
	}

	// Apply normalization
	result := normalizeUnicode(str, form)
	return &StringValue{Value: result}
}

// builtinStripAccents implements the StripAccents() built-in function.
// It removes diacritical marks from a string.
// StripAccents(str) - returns string with accents removed
func (i *Interpreter) builtinStripAccents(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "StripAccents() expects exactly 1 argument, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StripAccents() expects string as argument, got %s", args[0].Type())
	}

	str := strVal.Value

	// Strip accents using NFD normalization and then removing combining marks
	result := stripAccents(str)
	return &StringValue{Value: result}
}
