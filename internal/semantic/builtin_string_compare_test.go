package semantic

import (
	"testing"
)

// ============================================================================
// String Comparison Functions Tests (Phase 9.17.4)
// ============================================================================

// SameText function tests
func TestBuiltinSameText_Basic(t *testing.T) {
	input := `
		var result := SameText('hello', 'HELLO');
	`
	expectNoErrors(t, input)
}

func TestBuiltinSameText_Different(t *testing.T) {
	input := `
		var result := SameText('hello', 'world');
	`
	expectNoErrors(t, input)
}

func TestBuiltinSameText_WrongArgCount(t *testing.T) {
	input := `
		var result := SameText('hello');
	`
	expectError(t, input, "2 arguments")
}

func TestBuiltinSameText_WrongType(t *testing.T) {
	input := `
		var result := SameText(42, 'hello');
	`
	expectError(t, input, "string")
}

// CompareText function tests
func TestBuiltinCompareText_Basic(t *testing.T) {
	input := `
		var result := CompareText('apple', 'BANANA');
	`
	expectNoErrors(t, input)
}

func TestBuiltinCompareText_WrongArgCount(t *testing.T) {
	input := `
		var result := CompareText('hello');
	`
	expectError(t, input, "2 arguments")
}

// CompareStr function tests
func TestBuiltinCompareStr_Basic(t *testing.T) {
	input := `
		var result := CompareStr('apple', 'banana');
	`
	expectNoErrors(t, input)
}

func TestBuiltinCompareStr_WrongType(t *testing.T) {
	input := `
		var result := CompareStr(42, 'banana');
	`
	expectError(t, input, "string")
}

// AnsiCompareText function tests
func TestBuiltinAnsiCompareText_Basic(t *testing.T) {
	input := `
		var result := AnsiCompareText('apple', 'BANANA');
	`
	expectNoErrors(t, input)
}

// AnsiCompareStr function tests
func TestBuiltinAnsiCompareStr_Basic(t *testing.T) {
	input := `
		var result := AnsiCompareStr('apple', 'banana');
	`
	expectNoErrors(t, input)
}

// CompareLocaleStr function tests
func TestBuiltinCompareLocaleStr_TwoArgs(t *testing.T) {
	input := `
		var result := CompareLocaleStr('apple', 'banana');
	`
	expectNoErrors(t, input)
}

func TestBuiltinCompareLocaleStr_ThreeArgs(t *testing.T) {
	input := `
		var result := CompareLocaleStr('apple', 'banana', 'en-US');
	`
	expectNoErrors(t, input)
}

func TestBuiltinCompareLocaleStr_FourArgs(t *testing.T) {
	input := `
		var result := CompareLocaleStr('apple', 'banana', 'en-US', true);
	`
	expectNoErrors(t, input)
}

func TestBuiltinCompareLocaleStr_WrongArgCount(t *testing.T) {
	input := `
		var result := CompareLocaleStr('hello');
	`
	expectError(t, input, "2 to 4 arguments")
}

// StrMatches function tests
func TestBuiltinStrMatches_Basic(t *testing.T) {
	input := `
		var result := StrMatches('hello.txt', '*.txt');
	`
	expectNoErrors(t, input)
}

func TestBuiltinStrMatches_Question(t *testing.T) {
	input := `
		var result := StrMatches('hello', 'hel?o');
	`
	expectNoErrors(t, input)
}

func TestBuiltinStrMatches_WrongArgCount(t *testing.T) {
	input := `
		var result := StrMatches('hello');
	`
	expectError(t, input, "2 arguments")
}

func TestBuiltinStrMatches_WrongType(t *testing.T) {
	input := `
		var result := StrMatches(42, '*.txt');
	`
	expectError(t, input, "string")
}

// StrIsASCII function tests
func TestBuiltinStrIsASCII_Basic(t *testing.T) {
	input := `
		var result := StrIsASCII('hello');
	`
	expectNoErrors(t, input)
}

func TestBuiltinStrIsASCII_WrongArgCount(t *testing.T) {
	input := `
		var result := StrIsASCII('hello', 'world');
	`
	expectError(t, input, "1 argument")
}

func TestBuiltinStrIsASCII_WrongType(t *testing.T) {
	input := `
		var result := StrIsASCII(42);
	`
	expectError(t, input, "string")
}
