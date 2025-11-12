package interp

import (
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// builtinSameText implements the SameText() built-in function.
// It performs case-insensitive string comparison for equality.
// SameText(str1, str2) - returns true if strings are equal ignoring case
func (i *Interpreter) builtinSameText(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "SameText() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string 1
	str1Val, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "SameText() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string 2
	str2Val, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "SameText() expects string as second argument, got %s", args[1].Type())
	}

	str1 := str1Val.Value
	str2 := str2Val.Value

	// Case-insensitive comparison using Unicode case folding
	result := strings.EqualFold(str1, str2)
	return &BooleanValue{Value: result}
}

// builtinCompareText implements the CompareText() built-in function.
// It performs case-insensitive string comparison.
// CompareText(str1, str2) - returns -1 if str1 < str2, 0 if equal, 1 if str1 > str2
func (i *Interpreter) builtinCompareText(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "CompareText() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string 1
	str1Val, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "CompareText() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string 2
	str2Val, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "CompareText() expects string as second argument, got %s", args[1].Type())
	}

	str1 := strings.ToLower(str1Val.Value)
	str2 := strings.ToLower(str2Val.Value)

	// Compare strings
	if str1 < str2 {
		return &IntegerValue{Value: -1}
	} else if str1 > str2 {
		return &IntegerValue{Value: 1}
	}
	return &IntegerValue{Value: 0}
}

// builtinCompareStr implements the CompareStr() built-in function.
// It performs case-sensitive string comparison.
// CompareStr(str1, str2) - returns -1 if str1 < str2, 0 if equal, 1 if str1 > str2
func (i *Interpreter) builtinCompareStr(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "CompareStr() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string 1
	str1Val, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "CompareStr() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string 2
	str2Val, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "CompareStr() expects string as second argument, got %s", args[1].Type())
	}

	str1 := str1Val.Value
	str2 := str2Val.Value

	// Compare strings
	if str1 < str2 {
		return &IntegerValue{Value: -1}
	} else if str1 > str2 {
		return &IntegerValue{Value: 1}
	}
	return &IntegerValue{Value: 0}
}

// builtinAnsiCompareText implements the AnsiCompareText() built-in function.
// It performs ANSI case-insensitive string comparison.
// AnsiCompareText(str1, str2) - returns -1 if str1 < str2, 0 if equal, 1 if str1 > str2
func (i *Interpreter) builtinAnsiCompareText(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "AnsiCompareText() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string 1
	str1Val, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "AnsiCompareText() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string 2
	str2Val, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "AnsiCompareText() expects string as second argument, got %s", args[1].Type())
	}

	str1 := strings.ToLower(str1Val.Value)
	str2 := strings.ToLower(str2Val.Value)

	// For ANSI comparison, use simple byte-wise comparison after lowercasing
	if str1 < str2 {
		return &IntegerValue{Value: -1}
	} else if str1 > str2 {
		return &IntegerValue{Value: 1}
	}
	return &IntegerValue{Value: 0}
}

// builtinAnsiCompareStr implements the AnsiCompareStr() built-in function.
// It performs ANSI case-sensitive string comparison.
// AnsiCompareStr(str1, str2) - returns -1 if str1 < str2, 0 if equal, 1 if str1 > str2
func (i *Interpreter) builtinAnsiCompareStr(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "AnsiCompareStr() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string 1
	str1Val, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "AnsiCompareStr() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string 2
	str2Val, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "AnsiCompareStr() expects string as second argument, got %s", args[1].Type())
	}

	str1 := str1Val.Value
	str2 := str2Val.Value

	// ANSI case-sensitive comparison
	if str1 < str2 {
		return &IntegerValue{Value: -1}
	} else if str1 > str2 {
		return &IntegerValue{Value: 1}
	}
	return &IntegerValue{Value: 0}
}

// builtinCompareLocaleStr implements the CompareLocaleStr() built-in function.
// It performs locale-aware string comparison.
// CompareLocaleStr(str1, str2) - uses default locale, case-insensitive
// CompareLocaleStr(str1, str2, locale) - uses specified locale, case-insensitive
// CompareLocaleStr(str1, str2, locale, caseSensitive) - with case sensitivity control
func (i *Interpreter) builtinCompareLocaleStr(args []Value) Value {
	if len(args) < 2 || len(args) > 4 {
		return i.newErrorWithLocation(i.currentNode, "CompareLocaleStr() expects 2 to 4 arguments, got %d", len(args))
	}

	// First argument: string 1
	str1Val, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "CompareLocaleStr() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string 2
	str2Val, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "CompareLocaleStr() expects string as second argument, got %s", args[1].Type())
	}

	str1 := str1Val.Value
	str2 := str2Val.Value

	// Default locale is English
	locale := "en"
	caseSensitive := false

	// Optional third argument: locale
	if len(args) >= 3 {
		localeVal, ok := args[2].(*StringValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "CompareLocaleStr() expects string as third argument, got %s", args[2].Type())
		}
		locale = localeVal.Value
	}

	// Optional fourth argument: case sensitivity
	if len(args) == 4 {
		csVal, ok := args[3].(*BooleanValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "CompareLocaleStr() expects boolean as fourth argument, got %s", args[3].Type())
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
	col := collate.New(tag)
	if !caseSensitive {
		col = collate.New(tag, collate.IgnoreCase)
	}

	// Compare strings
	result := col.CompareString(str1, str2)
	if result < 0 {
		return &IntegerValue{Value: -1}
	} else if result > 0 {
		return &IntegerValue{Value: 1}
	}
	return &IntegerValue{Value: 0}
}

// builtinStrMatches implements the StrMatches() built-in function.
// It performs wildcard pattern matching.
// StrMatches(str, mask) - returns true if string matches pattern (* and ? wildcards)
func (i *Interpreter) builtinStrMatches(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "StrMatches() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: string to match
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrMatches() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: pattern/mask
	maskVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrMatches() expects string as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	mask := maskVal.Value

	// Use filepath.Match for wildcard pattern matching
	// Note: filepath.Match uses * for zero or more characters and ? for single character
	matched, err := filepath.Match(mask, str)
	if err != nil {
		// Invalid pattern
		return i.newErrorWithLocation(i.currentNode, "StrMatches() invalid pattern: %s", err.Error())
	}

	return &BooleanValue{Value: matched}
}

// builtinStrIsASCII implements the StrIsASCII() built-in function.
// It checks if a string contains only ASCII characters.
// StrIsASCII(str) - returns true if string is pure ASCII (all chars < 128)
func (i *Interpreter) builtinStrIsASCII(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "StrIsASCII() expects exactly 1 argument, got %d", len(args))
	}

	// First argument: string to check
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrIsASCII() expects string as argument, got %s", args[0].Type())
	}

	str := strVal.Value

	// Check if all characters are ASCII (< 128)
	for _, r := range str {
		if r > unicode.MaxASCII {
			return &BooleanValue{Value: false}
		}
	}

	return &BooleanValue{Value: true}
}
