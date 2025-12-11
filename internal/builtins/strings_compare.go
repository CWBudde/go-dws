package builtins

import (
	"strings"
	"unicode"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
	"golang.org/x/text/unicode/norm"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

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

	// Special-case French collation to match DWScript fixtures
	if strings.HasPrefix(strings.ToLower(locale), "fr") {
		cmp := compareFrench(str1, str2, caseSensitive)
		return &runtime.IntegerValue{Value: int64(cmp)}
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

// compareFrench implements a simplified French collation rule set used in fixtures:
// strings are compared case-insensitively by default, ignoring accents for the
// primary ordering. If the base strings are identical, the last accent decides
// ordering with precedence: none < circumflex < grave < acute.
// If last accents are equal, compare all accents from right to left.
func compareFrench(a, b string, caseSensitive bool) int {
	if !caseSensitive {
		a = strings.ToLower(a)
		b = strings.ToLower(b)
	}

	baseA := stripAccentsLocal(a)
	baseB := stripAccentsLocal(b)
	if baseA < baseB {
		return -1
	}
	if baseA > baseB {
		return 1
	}

	// Compare accents from right to left
	accentsA := getAccentWeights(a)
	accentsB := getAccentWeights(b)

	// Compare from rightmost accent
	maxLen := len(accentsA)
	if len(accentsB) > maxLen {
		maxLen = len(accentsB)
	}

	for i := 0; i < maxLen; i++ {
		var wA, wB int
		if i < len(accentsA) {
			wA = accentsA[len(accentsA)-1-i]
		}
		if i < len(accentsB) {
			wB = accentsB[len(accentsB)-1-i]
		}
		if wA < wB {
			return -1
		}
		if wA > wB {
			return 1
		}
	}
	return 0
}

func stripAccentsLocal(s string) string {
	decomposed := norm.NFD.String(s)
	var result []rune
	for _, r := range decomposed {
		if !unicode.Is(unicode.Mn, r) {
			result = append(result, r)
		}
	}
	return string(result)
}

// getAccentWeights returns a slice of accent weights for each accented character
// in the string, in left-to-right order. Weight: none=0, circumflex=1, grave=2, acute=3.
func getAccentWeights(s string) []int {
	decomposed := []rune(norm.NFD.String(s))
	var weights []int
	for i := 0; i < len(decomposed); i++ {
		r := decomposed[i]
		// Skip combining marks
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		// Check if next rune is a combining mark (accent)
		weight := 0
		if i+1 < len(decomposed) && unicode.Is(unicode.Mn, decomposed[i+1]) {
			switch decomposed[i+1] {
			case 0x0302: // circumflex
				weight = 1
			case 0x0300: // grave
				weight = 2
			case 0x0301: // acute
				weight = 3
			default:
				weight = 1
			}
			weights = append(weights, weight)
		}
	}
	return weights
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
