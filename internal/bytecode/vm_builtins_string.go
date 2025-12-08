package bytecode

import (
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
	"golang.org/x/text/unicode/norm"
)

// registerStringBuiltins registers all string manipulation functions
func (vm *VM) registerStringBuiltins() {
	vm.builtins["Copy"] = builtinCopy
	vm.builtins["SubStr"] = builtinSubStr
	vm.builtins["SubString"] = builtinSubString
	vm.builtins["LeftStr"] = builtinLeftStr
	vm.builtins["RightStr"] = builtinRightStr
	vm.builtins["MidStr"] = builtinMidStr
	vm.builtins["StrBeginsWith"] = builtinStrBeginsWith
	vm.builtins["StrEndsWith"] = builtinStrEndsWith
	vm.builtins["StrContains"] = builtinStrContains
	vm.builtins["PosEx"] = builtinPosEx
	vm.builtins["RevPos"] = builtinRevPos
	vm.builtins["StrFind"] = builtinStrFind
	vm.builtins["StrSplit"] = builtinStrSplit
	vm.builtins["StrJoin"] = builtinStrJoin
	vm.builtins["StrArrayPack"] = builtinStrArrayPack
	vm.builtins["StrBefore"] = builtinStrBefore
	vm.builtins["StrBeforeLast"] = builtinStrBeforeLast
	vm.builtins["StrAfter"] = builtinStrAfter
	vm.builtins["StrAfterLast"] = builtinStrAfterLast
	vm.builtins["StrBetween"] = builtinStrBetween
	vm.builtins["IsDelimiter"] = builtinIsDelimiter
	vm.builtins["LastDelimiter"] = builtinLastDelimiter
	vm.builtins["FindDelimiter"] = builtinFindDelimiter
	vm.builtins["PadLeft"] = builtinPadLeft
	vm.builtins["PadRight"] = builtinPadRight
	vm.builtins["StrDeleteLeft"] = builtinStrDeleteLeft
	vm.builtins["DeleteLeft"] = builtinStrDeleteLeft
	vm.builtins["StrDeleteRight"] = builtinStrDeleteRight
	vm.builtins["DeleteRight"] = builtinStrDeleteRight
	vm.builtins["ReverseString"] = builtinReverseString
	vm.builtins["QuotedStr"] = builtinQuotedStr
	vm.builtins["StringOfString"] = builtinStringOfString
	vm.builtins["DupeString"] = builtinDupeString
	vm.builtins["NormalizeString"] = builtinNormalizeString
	vm.builtins["Normalize"] = builtinNormalizeString
	vm.builtins["StripAccents"] = builtinStripAccents
	vm.builtins["SameText"] = builtinSameText
	vm.builtins["CompareText"] = builtinCompareText
	vm.builtins["CompareStr"] = builtinCompareStr
	vm.builtins["AnsiCompareText"] = builtinAnsiCompareText
	vm.builtins["AnsiCompareStr"] = builtinAnsiCompareStr
	vm.builtins["CompareLocaleStr"] = builtinCompareLocaleStr
	vm.builtins["StrMatches"] = builtinStrMatches
	vm.builtins["StrIsASCII"] = builtinStrIsASCII
	vm.builtins["Ord"] = builtinOrd
	vm.builtins["Chr"] = builtinChr
	vm.builtins["UpperCase"] = builtinUpperCase
	vm.builtins["LowerCase"] = builtinLowerCase
	vm.builtins["ASCIIUpperCase"] = builtinASCIIUpperCase
	vm.builtins["ASCIILowerCase"] = builtinASCIILowerCase
	vm.builtins["AnsiUpperCase"] = builtinAnsiUpperCase
	vm.builtins["AnsiLowerCase"] = builtinAnsiLowerCase
	vm.builtins["CharAt"] = builtinCharAt
	vm.builtins["ByteSizeToStr"] = builtinByteSizeToStr
	vm.builtins["GetText"] = builtinGetText
	vm.builtins["_"] = builtinGetText
}

func builtinCopy(vm *VM, args []Value) (Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return NilValue(), vm.runtimeError("Copy expects 2 or 3 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("Copy expects a string as first argument")
	}
	if !args[1].IsInt() {
		return NilValue(), vm.runtimeError("Copy expects an integer as second argument")
	}

	str := args[0].AsString()
	start := int(args[1].AsInt()) - 1 // DWScript uses 1-based indexing

	if start < 0 {
		start = 0
	}
	if start >= len(str) {
		return StringValue(""), nil
	}

	length := len(str) - start
	if len(args) == 3 {
		if !args[2].IsInt() {
			return NilValue(), vm.runtimeError("Copy expects an integer as third argument")
		}
		length = int(args[2].AsInt())
	}

	if start+length > len(str) {
		length = len(str) - start
	}

	return StringValue(str[start : start+length]), nil
}

func builtinSubStr(vm *VM, args []Value) (Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return NilValue(), vm.runtimeError("SubStr expects 2 or 3 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("SubStr expects a string as first argument")
	}
	if !args[1].IsInt() {
		return NilValue(), vm.runtimeError("SubStr expects an integer as second argument")
	}

	str := args[0].AsString()
	start := int(args[1].AsInt()) - 1 // DWScript uses 1-based indexing

	if start < 0 {
		start = 0
	}
	if start >= len(str) {
		return StringValue(""), nil
	}

	// Default length is to end of string (MaxInt in DWScript)
	length := len(str) - start
	if len(args) == 3 {
		if !args[2].IsInt() {
			return NilValue(), vm.runtimeError("SubStr expects an integer as third argument")
		}
		length = int(args[2].AsInt())
	}

	if start+length > len(str) {
		length = len(str) - start
	}

	return StringValue(str[start : start+length]), nil
}

func builtinSubString(vm *VM, args []Value) (Value, error) {
	if len(args) != 3 {
		return NilValue(), vm.runtimeError("SubString expects 3 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("SubString expects a string as first argument")
	}
	if !args[1].IsInt() {
		return NilValue(), vm.runtimeError("SubString expects an integer as second argument")
	}
	if !args[2].IsInt() {
		return NilValue(), vm.runtimeError("SubString expects an integer as third argument")
	}

	str := args[0].AsString()
	start := int(args[1].AsInt()) // 1-based
	end := int(args[2].AsInt())   // 1-based, end-exclusive

	if start < 1 {
		start = 1
	}

	// Calculate length from start and end positions (end-exclusive)
	length := end - start
	if length <= 0 {
		return StringValue(""), nil
	}

	// Convert to runes for UTF-8 support
	runes := []rune(str)
	startIdx := start - 1

	if startIdx >= len(runes) {
		return StringValue(""), nil
	}

	endIdx := startIdx + length
	if endIdx > len(runes) {
		endIdx = len(runes)
	}

	return StringValue(string(runes[startIdx:endIdx])), nil
}

func builtinLeftStr(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("LeftStr expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("LeftStr expects a string as first argument")
	}
	if !args[1].IsInt() {
		return NilValue(), vm.runtimeError("LeftStr expects an integer as second argument")
	}

	str := args[0].AsString()
	count := int(args[1].AsInt())

	if count <= 0 {
		return StringValue(""), nil
	}

	// Convert to runes for UTF-8 support
	runes := []rune(str)
	if count > len(runes) {
		count = len(runes)
	}

	return StringValue(string(runes[:count])), nil
}

func builtinRightStr(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("RightStr expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("RightStr expects a string as first argument")
	}
	if !args[1].IsInt() {
		return NilValue(), vm.runtimeError("RightStr expects an integer as second argument")
	}

	str := args[0].AsString()
	count := int(args[1].AsInt())

	if count <= 0 {
		return StringValue(""), nil
	}

	// Convert to runes for UTF-8 support
	runes := []rune(str)
	strLen := len(runes)

	if count >= strLen {
		return StringValue(str), nil
	}

	start := strLen - count
	return StringValue(string(runes[start:])), nil
}

func builtinMidStr(vm *VM, args []Value) (Value, error) {
	// MidStr is an alias for SubStr
	return builtinSubStr(vm, args)
}

func builtinStrBeginsWith(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("StrBeginsWith expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrBeginsWith expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("StrBeginsWith expects a string as second argument")
	}

	str := args[0].AsString()
	prefix := args[1].AsString()

	// DWScript treats empty prefixes as false (see JS RTL implementation)
	result := len(prefix) != 0 && len(str) >= len(prefix) && str[:len(prefix)] == prefix
	return BoolValue(result), nil
}

func builtinStrEndsWith(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("StrEndsWith expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrEndsWith expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("StrEndsWith expects a string as second argument")
	}

	str := args[0].AsString()
	suffix := args[1].AsString()

	result := len(suffix) == 0 || (len(str) >= len(suffix) && str[len(str)-len(suffix):] == suffix)
	return BoolValue(result), nil
}

func builtinStrContains(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("StrContains expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrContains expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("StrContains expects a string as second argument")
	}

	str := args[0].AsString()
	substr := args[1].AsString()

	// Empty substring is always contained
	if len(substr) == 0 {
		return BoolValue(true), nil
	}

	// Check if substring exists
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return BoolValue(true), nil
		}
	}

	return BoolValue(false), nil
}

func builtinPosEx(vm *VM, args []Value) (Value, error) {
	if len(args) != 3 {
		return NilValue(), vm.runtimeError("PosEx expects 3 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("PosEx expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("PosEx expects a string as second argument")
	}
	if !args[2].IsInt() {
		return NilValue(), vm.runtimeError("PosEx expects an integer as third argument")
	}

	needle := args[0].AsString()
	haystack := args[1].AsString()
	offset := int(args[2].AsInt()) // 1-based

	// Handle invalid offset first
	if offset < 1 {
		return IntValue(0), nil
	}

	// Handle empty needle - returns 0 (not found)
	if len(needle) == 0 {
		return IntValue(0), nil
	}

	// Convert to runes for UTF-8 support
	haystackRunes := []rune(haystack)
	needleRunes := []rune(needle)

	// Adjust offset to 0-based
	startIdx := offset - 1

	// If offset is beyond the string length, not found
	if startIdx >= len(haystackRunes) {
		return IntValue(0), nil
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
			return IntValue(int64(i + 1)), nil
		}
	}

	// Not found
	return IntValue(0), nil
}

func builtinRevPos(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("RevPos expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("RevPos expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("RevPos expects a string as second argument")
	}

	needle := args[0].AsString()
	haystack := args[1].AsString()

	// Handle empty needle - return 0 (not found)
	if len(needle) == 0 {
		return IntValue(0), nil
	}

	// Convert to runes for UTF-8 support
	haystackRunes := []rune(haystack)
	needleRunes := []rune(needle)

	// Search backwards for the last occurrence
	for i := len(haystackRunes) - len(needleRunes); i >= 0; i-- {
		match := true
		for j := 0; j < len(needleRunes); j++ {
			if haystackRunes[i+j] != needleRunes[j] {
				match = false
				break
			}
		}
		if match {
			// Return 1-based position
			return IntValue(int64(i + 1)), nil
		}
	}

	// Not found
	return IntValue(0), nil
}

func builtinStrFind(vm *VM, args []Value) (Value, error) {
	if len(args) != 3 {
		return NilValue(), vm.runtimeError("StrFind expects 3 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrFind expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("StrFind expects a string as second argument")
	}
	if !args[2].IsInt() {
		return NilValue(), vm.runtimeError("StrFind expects an integer as third argument")
	}

	// StrFind(str, substr, fromIndex) maps to PosEx(substr, str, fromIndex)
	// Reorder arguments
	reorderedArgs := []Value{
		args[1], // substr becomes first arg (needle)
		args[0], // str becomes second arg (haystack)
		args[2], // fromIndex stays as third arg (offset)
	}

	return builtinPosEx(vm, reorderedArgs)
}

func builtinStrSplit(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("StrSplit expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrSplit expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("StrSplit expects a string as second argument")
	}

	str := args[0].AsString()
	delim := args[1].AsString()

	// Handle empty delimiter - return array with single element (the original string)
	if len(delim) == 0 {
		elements := []Value{StringValue(str)}
		return ArrayValue(NewArrayInstance(elements)), nil
	}

	// Split the string
	parts := strings.Split(str, delim)

	// Convert to array of Value
	elements := make([]Value, len(parts))
	for idx, part := range parts {
		elements[idx] = StringValue(part)
	}

	return ArrayValue(NewArrayInstance(elements)), nil
}

func builtinStrJoin(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("StrJoin expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsArray() {
		return NilValue(), vm.runtimeError("StrJoin expects an array as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("StrJoin expects a string as second argument")
	}

	arr := args[0].AsArray()
	delim := args[1].AsString()

	// Convert array elements to strings
	parts := make([]string, len(arr.elements))
	for idx, elem := range arr.elements {
		if !elem.IsString() {
			return NilValue(), vm.runtimeError("StrJoin expects array of strings, got %s at index %d", elem.Type.String(), idx)
		}
		parts[idx] = elem.AsString()
	}

	// Join the strings
	result := strings.Join(parts, delim)
	return StringValue(result), nil
}

func builtinStrArrayPack(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("StrArrayPack expects 1 argument, got %d", len(args))
	}
	if !args[0].IsArray() {
		return NilValue(), vm.runtimeError("StrArrayPack expects an array as argument")
	}

	arr := args[0].AsArray()

	// Filter out empty strings
	var packed []Value
	for _, elem := range arr.elements {
		if !elem.IsString() {
			return NilValue(), vm.runtimeError("StrArrayPack expects array of strings, got %s", elem.Type.String())
		}
		if elem.AsString() != "" {
			packed = append(packed, elem)
		}
	}

	return ArrayValue(NewArrayInstance(packed)), nil
}

func builtinStrBefore(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("StrBefore expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrBefore expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("StrBefore expects a string as second argument")
	}

	str := args[0].AsString()
	delim := args[1].AsString()

	// Handle empty delimiter - return original string
	if len(delim) == 0 {
		return StringValue(str), nil
	}

	// Find the first occurrence of delimiter
	index := strings.Index(str, delim)
	if index == -1 {
		// Delimiter not found - return the full string
		return StringValue(str), nil
	}

	// Return substring before delimiter
	return StringValue(str[:index]), nil
}

func builtinStrBeforeLast(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("StrBeforeLast expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrBeforeLast expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("StrBeforeLast expects a string as second argument")
	}

	str := args[0].AsString()
	delim := args[1].AsString()

	// Handle empty delimiter - return original string
	if len(delim) == 0 {
		return StringValue(str), nil
	}

	// Find the last occurrence of delimiter
	index := strings.LastIndex(str, delim)
	if index == -1 {
		// Delimiter not found - return the full string
		return StringValue(str), nil
	}

	// Return substring before last delimiter
	return StringValue(str[:index]), nil
}

func builtinStrAfter(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("StrAfter expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrAfter expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("StrAfter expects a string as second argument")
	}

	str := args[0].AsString()
	delim := args[1].AsString()

	// Handle empty delimiter - return empty string
	if len(delim) == 0 {
		return StringValue(""), nil
	}

	// Find the first occurrence of delimiter
	index := strings.Index(str, delim)
	if index == -1 {
		// Delimiter not found - return empty string
		return StringValue(""), nil
	}

	// Return substring after delimiter
	return StringValue(str[index+len(delim):]), nil
}

func builtinStrAfterLast(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("StrAfterLast expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrAfterLast expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("StrAfterLast expects a string as second argument")
	}

	str := args[0].AsString()
	delim := args[1].AsString()

	// Handle empty delimiter - return empty string
	if len(delim) == 0 {
		return StringValue(""), nil
	}

	// Find the last occurrence of delimiter
	index := strings.LastIndex(str, delim)
	if index == -1 {
		// Delimiter not found - return empty string
		return StringValue(""), nil
	}

	// Return substring after last delimiter
	return StringValue(str[index+len(delim):]), nil
}

func builtinStrBetween(vm *VM, args []Value) (Value, error) {
	if len(args) != 3 {
		return NilValue(), vm.runtimeError("StrBetween expects 3 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrBetween expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("StrBetween expects a string as second argument")
	}
	if !args[2].IsString() {
		return NilValue(), vm.runtimeError("StrBetween expects a string as third argument")
	}

	str := args[0].AsString()
	start := args[1].AsString()
	stop := args[2].AsString()

	// Handle empty delimiters - return empty string
	if len(start) == 0 || len(stop) == 0 {
		return StringValue(""), nil
	}

	// Find the first occurrence of start delimiter
	startIdx := strings.Index(str, start)
	if startIdx == -1 {
		// Start delimiter not found - return empty string
		return StringValue(""), nil
	}

	// Search for stop delimiter after the start delimiter
	searchFrom := startIdx + len(start)
	if searchFrom >= len(str) {
		// No room for stop delimiter - return empty string
		return StringValue(""), nil
	}

	stopIdx := strings.Index(str[searchFrom:], stop)
	if stopIdx == -1 {
		// Stop delimiter not found - return empty string
		return StringValue(""), nil
	}

	// Adjust stopIdx to be relative to the original string
	stopIdx += searchFrom

	// Return substring between start and stop delimiters
	return StringValue(str[searchFrom:stopIdx]), nil
}

func builtinIsDelimiter(vm *VM, args []Value) (Value, error) {
	if len(args) != 3 {
		return NilValue(), vm.runtimeError("IsDelimiter expects 3 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("IsDelimiter expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("IsDelimiter expects a string as second argument")
	}
	if !args[2].IsInt() {
		return NilValue(), vm.runtimeError("IsDelimiter expects an integer as third argument")
	}

	delims := args[0].AsString()
	str := args[1].AsString()
	index := args[2].AsInt() // 1-based

	// Handle invalid index
	if index < 1 {
		return BoolValue(false), nil
	}

	// Convert to rune-based indexing for UTF-8 support
	strRunes := []rune(str)

	// Check if index is within bounds (1-based)
	if int(index) > len(strRunes) {
		return BoolValue(false), nil
	}

	// Get the character at the specified position (convert to 0-based)
	ch := strRunes[index-1]

	// Check if the character is in the delimiter string
	result := strings.ContainsRune(delims, ch)
	return BoolValue(result), nil
}

func builtinLastDelimiter(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("LastDelimiter expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("LastDelimiter expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("LastDelimiter expects a string as second argument")
	}

	delims := args[0].AsString()
	str := args[1].AsString()

	// Convert to rune-based for UTF-8 support
	strRunes := []rune(str)

	// Search from the end for any delimiter character
	for i := len(strRunes) - 1; i >= 0; i-- {
		if strings.ContainsRune(delims, strRunes[i]) {
			// Return 1-based position
			return IntValue(int64(i + 1)), nil
		}
	}

	// No delimiter found
	return IntValue(0), nil
}

func builtinFindDelimiter(vm *VM, args []Value) (Value, error) {
	if len(args) != 3 {
		return NilValue(), vm.runtimeError("FindDelimiter expects 3 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("FindDelimiter expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("FindDelimiter expects a string as second argument")
	}
	if !args[2].IsInt() {
		return NilValue(), vm.runtimeError("FindDelimiter expects an integer as third argument")
	}

	delims := args[0].AsString()
	str := args[1].AsString()
	startIndex := args[2].AsInt() // 1-based

	// Handle invalid start index
	if startIndex < 1 {
		return IntValue(-1), nil
	}

	// Convert to rune-based for UTF-8 support
	strRunes := []rune(str)

	// Adjust to 0-based index
	startIdx := int(startIndex) - 1

	// Check if start index is within bounds
	if startIdx >= len(strRunes) {
		return IntValue(-1), nil
	}

	// Search from startIdx for any delimiter character
	for i := startIdx; i < len(strRunes); i++ {
		if strings.ContainsRune(delims, strRunes[i]) {
			// Return 1-based position
			return IntValue(int64(i + 1)), nil
		}
	}

	// No delimiter found - return -1
	return IntValue(-1), nil
}

func builtinPadLeft(vm *VM, args []Value) (Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return NilValue(), vm.runtimeError("PadLeft expects 2 or 3 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("PadLeft expects a string as first argument")
	}
	if !args[1].IsInt() {
		return NilValue(), vm.runtimeError("PadLeft expects an integer as second argument")
	}

	str := args[0].AsString()
	count := int(args[1].AsInt())
	padChar := " "

	if len(args) == 3 {
		if !args[2].IsString() {
			return NilValue(), vm.runtimeError("PadLeft expects a string as third argument")
		}
		padCharStr := args[2].AsString()
		if len(padCharStr) > 0 {
			runes := []rune(padCharStr)
			padChar = string(runes[0])
		}
	}

	strRunes := []rune(str)
	strLen := len(strRunes)
	if strLen >= count {
		return StringValue(str), nil
	}

	padding := strings.Repeat(padChar, count-strLen)
	return StringValue(padding + str), nil
}

func builtinPadRight(vm *VM, args []Value) (Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return NilValue(), vm.runtimeError("PadRight expects 2 or 3 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("PadRight expects a string as first argument")
	}
	if !args[1].IsInt() {
		return NilValue(), vm.runtimeError("PadRight expects an integer as second argument")
	}

	str := args[0].AsString()
	count := int(args[1].AsInt())
	padChar := " "

	if len(args) == 3 {
		if !args[2].IsString() {
			return NilValue(), vm.runtimeError("PadRight expects a string as third argument")
		}
		padCharStr := args[2].AsString()
		if len(padCharStr) > 0 {
			runes := []rune(padCharStr)
			padChar = string(runes[0])
		}
	}

	strRunes := []rune(str)
	strLen := len(strRunes)
	if strLen >= count {
		return StringValue(str), nil
	}

	padding := strings.Repeat(padChar, count-strLen)
	return StringValue(str + padding), nil
}

func builtinStrDeleteLeft(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("StrDeleteLeft expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrDeleteLeft expects a string as first argument")
	}
	if !args[1].IsInt() {
		return NilValue(), vm.runtimeError("StrDeleteLeft expects an integer as second argument")
	}

	str := args[0].AsString()
	count := int(args[1].AsInt())
	if count <= 0 {
		return StringValue(str), nil
	}

	strRunes := []rune(str)
	strLen := len(strRunes)
	if count >= strLen {
		return StringValue(""), nil
	}

	return StringValue(string(strRunes[count:])), nil
}

func builtinStrDeleteRight(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("StrDeleteRight expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrDeleteRight expects a string as first argument")
	}
	if !args[1].IsInt() {
		return NilValue(), vm.runtimeError("StrDeleteRight expects an integer as second argument")
	}

	str := args[0].AsString()
	count := int(args[1].AsInt())
	if count <= 0 {
		return StringValue(str), nil
	}

	strRunes := []rune(str)
	strLen := len(strRunes)
	if count >= strLen {
		return StringValue(""), nil
	}

	return StringValue(string(strRunes[:strLen-count])), nil
}

func builtinReverseString(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("ReverseString expects 1 argument, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("ReverseString expects a string as argument")
	}

	str := args[0].AsString()
	strRunes := []rune(str)
	for i, j := 0, len(strRunes)-1; i < j; i, j = i+1, j-1 {
		strRunes[i], strRunes[j] = strRunes[j], strRunes[i]
	}

	return StringValue(string(strRunes)), nil
}

func builtinQuotedStr(vm *VM, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return NilValue(), vm.runtimeError("QuotedStr expects 1 or 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("QuotedStr expects a string as first argument")
	}

	str := args[0].AsString()
	quoteChar := "'"

	if len(args) == 2 {
		if !args[1].IsString() {
			return NilValue(), vm.runtimeError("QuotedStr expects a string as second argument")
		}
		quoteCharStr := args[1].AsString()
		if len(quoteCharStr) > 0 {
			runes := []rune(quoteCharStr)
			quoteChar = string(runes[0])
		}
	}

	escaped := strings.ReplaceAll(str, quoteChar, quoteChar+quoteChar)
	return StringValue(quoteChar + escaped + quoteChar), nil
}

func builtinStringOfString(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("StringOfString expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StringOfString expects a string as first argument")
	}
	if !args[1].IsInt() {
		return NilValue(), vm.runtimeError("StringOfString expects an integer as second argument")
	}

	str := args[0].AsString()
	count := int(args[1].AsInt())
	if count <= 0 {
		return StringValue(""), nil
	}

	return StringValue(strings.Repeat(str, count)), nil
}

func builtinDupeString(vm *VM, args []Value) (Value, error) {
	return builtinStringOfString(vm, args)
}

func builtinNormalizeString(vm *VM, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return NilValue(), vm.runtimeError("NormalizeString expects 1 or 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("NormalizeString expects a string as first argument")
	}

	str := args[0].AsString()
	form := "NFC"

	if len(args) == 2 {
		if !args[1].IsString() {
			return NilValue(), vm.runtimeError("NormalizeString expects a string as second argument")
		}
		form = strings.ToUpper(args[1].AsString())
	}

	result := normalizeStringUnicode(str, form)
	return StringValue(result), nil
}

func builtinStripAccents(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("StripAccents expects 1 argument, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StripAccents expects a string as argument")
	}

	str := args[0].AsString()
	result := stripStringAccents(str)
	return StringValue(result), nil
}

func builtinSameText(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("SameText expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("SameText expects string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("SameText expects string as second argument")
	}

	str1 := args[0].AsString()
	str2 := args[1].AsString()
	result := strings.EqualFold(str1, str2)
	return BoolValue(result), nil
}

func builtinCompareText(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("CompareText expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("CompareText expects string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("CompareText expects string as second argument")
	}

	str1 := strings.ToLower(args[0].AsString())
	str2 := strings.ToLower(args[1].AsString())

	if str1 < str2 {
		return IntValue(-1), nil
	} else if str1 > str2 {
		return IntValue(1), nil
	}
	return IntValue(0), nil
}

func builtinCompareStr(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("CompareStr expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("CompareStr expects string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("CompareStr expects string as second argument")
	}

	str1 := args[0].AsString()
	str2 := args[1].AsString()

	if str1 < str2 {
		return IntValue(-1), nil
	} else if str1 > str2 {
		return IntValue(1), nil
	}
	return IntValue(0), nil
}

func builtinAnsiCompareText(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("AnsiCompareText expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("AnsiCompareText expects string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("AnsiCompareText expects string as second argument")
	}

	str1 := strings.ToLower(args[0].AsString())
	str2 := strings.ToLower(args[1].AsString())

	if str1 < str2 {
		return IntValue(-1), nil
	} else if str1 > str2 {
		return IntValue(1), nil
	}
	return IntValue(0), nil
}

func builtinAnsiCompareStr(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("AnsiCompareStr expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("AnsiCompareStr expects string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("AnsiCompareStr expects string as second argument")
	}

	str1 := args[0].AsString()
	str2 := args[1].AsString()

	if str1 < str2 {
		return IntValue(-1), nil
	} else if str1 > str2 {
		return IntValue(1), nil
	}
	return IntValue(0), nil
}

func builtinCompareLocaleStr(vm *VM, args []Value) (Value, error) {
	if len(args) < 2 || len(args) > 4 {
		return NilValue(), vm.runtimeError("CompareLocaleStr expects 2 to 4 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("CompareLocaleStr expects string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("CompareLocaleStr expects string as second argument")
	}

	str1 := args[0].AsString()
	str2 := args[1].AsString()

	// Default locale is English
	locale := "en"
	caseSensitive := false

	// Optional third argument: locale
	if len(args) >= 3 {
		if !args[2].IsString() {
			return NilValue(), vm.runtimeError("CompareLocaleStr expects string as third argument")
		}
		locale = args[2].AsString()
	}

	// Optional fourth argument: case sensitivity
	if len(args) == 4 {
		if !args[3].IsBool() {
			return NilValue(), vm.runtimeError("CompareLocaleStr expects boolean as fourth argument")
		}
		caseSensitive = args[3].AsBool()
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
		return IntValue(-1), nil
	} else if result > 0 {
		return IntValue(1), nil
	}
	return IntValue(0), nil
}

func builtinStrMatches(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("StrMatches expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrMatches expects string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("StrMatches expects string as second argument")
	}

	str := args[0].AsString()
	mask := args[1].AsString()

	matched := wildcardMatch(str, mask)
	return BoolValue(matched), nil
}

func builtinStrIsASCII(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("StrIsASCII expects 1 argument, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrIsASCII expects string as argument")
	}

	str := args[0].AsString()
	for _, r := range str {
		if r > unicode.MaxASCII {
			return BoolValue(false), nil
		}
	}
	return BoolValue(true), nil
}

func builtinOrd(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("Ord expects 1 argument, got %d", len(args))
	}
	arg := args[0]
	if arg.IsString() {
		s := arg.AsString()
		if len(s) == 0 {
			return IntValue(0), nil
		}
		return IntValue(int64(s[0])), nil
	}
	if arg.IsInt() {
		// For enums and other types
		return arg, nil
	}
	return NilValue(), vm.runtimeError("Ord expects a string or integer argument")
}

func builtinChr(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("Chr expects 1 argument, got %d", len(args))
	}
	if !args[0].IsInt() {
		return NilValue(), vm.runtimeError("Chr expects an integer argument")
	}

	c := args[0].AsInt()

	// Check if the code is in valid range (0-1114111 for Unicode)
	if c < 0 || c > 0x10FFFF {
		return NilValue(), vm.runtimeError("Chr code %d out of valid Unicode range (0-1114111)", c)
	}

	// Return UTF-8 encoded character (Go native)
	// NOTE: Unlike DWScript (which uses UTF-16), go-dws uses UTF-8 for all strings.
	// See docs/string-encoding.md for details.
	return StringValue(string(rune(c))), nil
}

func normalizeStringUnicode(s string, form string) string {
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

// stripStringAccents removes diacritical marks from a string.
// It works by normalizing the string to NFD (decomposed form) and then
// removing all combining marks (which include accents).
func stripStringAccents(s string) string {
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

// compareLocaleStrSimple performs a simplified locale-aware string comparison.
// For the bytecode VM, we use a simple case-insensitive comparison.
func compareLocaleStrSimple(str1, str2 string) int {
	s1 := strings.ToLower(str1)
	s2 := strings.ToLower(str2)
	if s1 < s2 {
		return -1
	} else if s1 > s2 {
		return 1
	}
	return 0
}

// wildcardMatch performs wildcard pattern matching.
// Supports * (zero or more characters) and ? (single character).
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

// builtinUpperCase implements the UpperCase() built-in function.
// It converts a string to uppercase.
func builtinUpperCase(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("UpperCase expects 1 argument, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("UpperCase expects a string argument")
	}
	return StringValue(strings.ToUpper(args[0].AsString())), nil
}

// builtinLowerCase implements the LowerCase() built-in function.
// It converts a string to lowercase.
func builtinLowerCase(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("LowerCase expects 1 argument, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("LowerCase expects a string argument")
	}
	return StringValue(strings.ToLower(args[0].AsString())), nil
}

// builtinASCIIUpperCase implements the ASCIIUpperCase() built-in function.
// It converts a string to uppercase using ASCII-only conversion.
func builtinASCIIUpperCase(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("ASCIIUpperCase expects 1 argument, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("ASCIIUpperCase expects a string argument")
	}

	str := args[0].AsString()
	result := make([]byte, len(str))
	for i, b := range []byte(str) {
		if b >= 'a' && b <= 'z' {
			result[i] = b - 32 // Convert to uppercase
		} else {
			result[i] = b
		}
	}

	return StringValue(string(result)), nil
}

// builtinASCIILowerCase implements the ASCIILowerCase() built-in function.
// It converts a string to lowercase using ASCII-only conversion.
func builtinASCIILowerCase(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("ASCIILowerCase expects 1 argument, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("ASCIILowerCase expects a string argument")
	}

	str := args[0].AsString()
	result := make([]byte, len(str))
	for i, b := range []byte(str) {
		if b >= 'A' && b <= 'Z' {
			result[i] = b + 32 // Convert to lowercase
		} else {
			result[i] = b
		}
	}

	return StringValue(string(result)), nil
}

// builtinAnsiUpperCase implements the AnsiUpperCase() built-in function.
// It is an alias for UpperCase().
func builtinAnsiUpperCase(vm *VM, args []Value) (Value, error) {
	return builtinUpperCase(vm, args)
}

// builtinAnsiLowerCase implements the AnsiLowerCase() built-in function.
// It is an alias for LowerCase().
func builtinAnsiLowerCase(vm *VM, args []Value) (Value, error) {
	return builtinLowerCase(vm, args)
}

// builtinCharAt implements the CharAt() built-in function.
// It returns the character at the specified position in a string (1-based).
func builtinCharAt(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("CharAt expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("CharAt expects string as first argument")
	}
	if !args[1].IsInt() {
		return NilValue(), vm.runtimeError("CharAt expects integer as second argument")
	}

	// Use SubStr to get a single character
	return builtinSubStr(vm, []Value{args[0], args[1], IntValue(1)})
}

// builtinByteSizeToStr implements the ByteSizeToStr() built-in function.
// It formats a byte size into a human-readable string (KB, MB, GB, TB).
func builtinByteSizeToStr(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("ByteSizeToStr expects 1 argument, got %d", len(args))
	}
	if !args[0].IsInt() {
		return NilValue(), vm.runtimeError("ByteSizeToStr expects an integer argument")
	}

	size := float64(args[0].AsInt())

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

	return StringValue(result), nil
}

// builtinGetText implements the GetText() and _() built-in functions.
// It is a localization/translation function that returns the input string unchanged.
func builtinGetText(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("GetText expects 1 argument, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("GetText expects a string argument")
	}

	// For now, just return the input string unchanged
	// In a full implementation, this would look up translations
	return args[0], nil
}
