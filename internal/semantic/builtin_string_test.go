package semantic

import (
	"testing"
)

// ============================================================================
// Built-in String Functions Tests
// ============================================================================
// These tests cover the built-in string manipulation functions to improve
// coverage of analyze_builtin_string.go (currently at 0-57% coverage)

// Length function tests
func TestBuiltinLength_String(t *testing.T) {
	input := `
		var s := 'hello';
		var len := Length(s);
	`
	expectNoErrors(t, input)
}

func TestBuiltinLength_Array(t *testing.T) {
	input := `
		var arr: array of Integer := [1, 2, 3];
		var len := Length(arr);
	`
	expectNoErrors(t, input)
}

func TestBuiltinLength_InvalidArgument(t *testing.T) {
	input := `
		var x := 42;
		var len := Length(x);
	`
	expectError(t, input, "length")
}

// Copy function tests
func TestBuiltinCopy_Basic(t *testing.T) {
	input := `
		var s := 'hello world';
		var sub := Copy(s, 7, 5);
	`
	expectNoErrors(t, input)
}

func TestBuiltinCopy_TwoArgs(t *testing.T) {
	input := `
		var s := 'hello';
		var sub := Copy(s, 2);
	`
	expectNoErrors(t, input)
}

func TestBuiltinCopy_InvalidArgCount(t *testing.T) {
	input := `
		var sub := Copy('hello');
	`
	expectError(t, input, "argument")
}

func TestBuiltinCopy_InvalidType(t *testing.T) {
	input := `
		var sub := Copy(42, 1, 5);
	`
	expectError(t, input, "string")
}

// Concat function tests
func TestBuiltinConcat_TwoStrings(t *testing.T) {
	input := `
		var result := Concat('hello', ' world');
	`
	expectNoErrors(t, input)
}

func TestBuiltinConcat_MultipleStrings(t *testing.T) {
	input := `
		var result := Concat('a', 'b', 'c', 'd');
	`
	expectNoErrors(t, input)
}

func TestBuiltinConcat_NoArgs(t *testing.T) {
	input := `
		var result := Concat();
	`
	expectError(t, input, "argument")
}

// Pos function tests
func TestBuiltinPos_Basic(t *testing.T) {
	input := `
		var position := Pos('world', 'hello world');
	`
	expectNoErrors(t, input)
}

func TestBuiltinPos_NotFound(t *testing.T) {
	input := `
		var position := Pos('xyz', 'hello world');
	`
	expectNoErrors(t, input)
}

func TestBuiltinPos_InvalidArgs(t *testing.T) {
	input := `
		var position := Pos(42, 'hello');
	`
	expectError(t, input, "string")
}

// UpperCase function tests
func TestBuiltinUpperCase_Basic(t *testing.T) {
	input := `
		var s := UpperCase('hello');
	`
	expectNoErrors(t, input)
}

func TestBuiltinUpperCase_Empty(t *testing.T) {
	input := `
		var s := UpperCase('');
	`
	expectNoErrors(t, input)
}

func TestBuiltinUpperCase_InvalidType(t *testing.T) {
	input := `
		var s := UpperCase(42);
	`
	expectError(t, input, "string")
}

// LowerCase function tests
func TestBuiltinLowerCase_Basic(t *testing.T) {
	input := `
		var s := LowerCase('HELLO');
	`
	expectNoErrors(t, input)
}

func TestBuiltinLowerCase_Mixed(t *testing.T) {
	input := `
		var s := LowerCase('HeLLo WoRLD');
	`
	expectNoErrors(t, input)
}

// Trim functions tests
func TestBuiltinTrim_Basic(t *testing.T) {
	input := `
		var s := Trim('  hello  ');
	`
	expectNoErrors(t, input)
}

func TestBuiltinTrimLeft_Basic(t *testing.T) {
	input := `
		var s := TrimLeft('  hello  ');
	`
	expectNoErrors(t, input)
}

func TestBuiltinTrimRight_Basic(t *testing.T) {
	input := `
		var s := TrimRight('  hello  ');
	`
	expectNoErrors(t, input)
}

func TestBuiltinTrim_InvalidType(t *testing.T) {
	input := `
		var s := Trim(123);
	`
	expectError(t, input, "string")
}

// StringReplace function tests
func TestBuiltinStringReplace_Basic(t *testing.T) {
	input := `
		var s := StringReplace('hello world', 'world', 'there');
	`
	expectNoErrors(t, input)
}

func TestBuiltinStringReplace_Multiple(t *testing.T) {
	input := `
		var s := StringReplace('hello hello hello', 'hello', 'hi');
	`
	expectNoErrors(t, input)
}

func TestBuiltinStringReplace_InvalidArgCount(t *testing.T) {
	input := `
		var s := StringReplace('hello', 'hi');
	`
	expectError(t, input, "argument")
}

// StringOfChar function tests
func TestBuiltinStringOfChar_Basic(t *testing.T) {
	input := `
		var s := StringOfChar('*', 10);
	`
	expectNoErrors(t, input)
}

func TestBuiltinStringOfChar_Zero(t *testing.T) {
	input := `
		var s := StringOfChar('x', 0);
	`
	expectNoErrors(t, input)
}

func TestBuiltinStringOfChar_InvalidCharType(t *testing.T) {
	input := `
		var s := StringOfChar('hello', 5);
	`
	expectError(t, input, "char")
}

func TestBuiltinStringOfChar_InvalidCountType(t *testing.T) {
	input := `
		var s := StringOfChar('x', 'five');
	`
	expectError(t, input, "integer")
}

// Format function tests
func TestBuiltinFormat_Basic(t *testing.T) {
	input := `
		var s := Format('Hello %s', ['world']);
	`
	expectNoErrors(t, input)
}

func TestBuiltinFormat_MultipleArgs(t *testing.T) {
	input := `
		var s := Format('Name: %s, Age: %d', ['John', 25]);
	`
	expectNoErrors(t, input)
}

func TestBuiltinFormat_NoArgs(t *testing.T) {
	input := `
		var s := Format('Hello world', []);
	`
	expectNoErrors(t, input)
}

func TestBuiltinFormat_InvalidArgCount(t *testing.T) {
	input := `
		var s := Format('hello');
	`
	expectError(t, input, "argument")
}

// Insert function tests
func TestBuiltinInsert_Basic(t *testing.T) {
	input := `
		var s := 'hello world';
		Insert('beautiful ', s, 7);
	`
	expectNoErrors(t, input)
}

func TestBuiltinInsert_Beginning(t *testing.T) {
	input := `
		var s := 'world';
		Insert('hello ', s, 1);
	`
	expectNoErrors(t, input)
}

func TestBuiltinInsert_InvalidArgCount(t *testing.T) {
	input := `
		var s := 'hello';
		Insert('x', s);
	`
	expectError(t, input, "argument")
}

func TestBuiltinInsert_InvalidType(t *testing.T) {
	input := `
		var s := 'hello';
		Insert(42, s, 1);
	`
	expectError(t, input, "string")
}

// Combined string operations tests
func TestBuiltinString_ChainedOperations(t *testing.T) {
	input := `
		var s := '  HELLO WORLD  ';
		s := Trim(s);
		s := LowerCase(s);
		var pos := Pos('world', s);
		var len := Length(s);
	`
	expectNoErrors(t, input)
}

func TestBuiltinString_InExpressions(t *testing.T) {
	input := `
		var s := UpperCase('hello') + ' ' + LowerCase('WORLD');
		var result := Length(s) > 5;
	`
	expectNoErrors(t, input)
}

func TestBuiltinString_AsParameters(t *testing.T) {
	input := `
		function ProcessString(s: String): String;
		begin
			Result := UpperCase(Trim(s));
		end;

		var result := ProcessString('  hello  ');
	`
	expectNoErrors(t, input)
}

// Edge cases
func TestBuiltinString_EmptyString(t *testing.T) {
	input := `
		var len := Length('');
		var upper := UpperCase('');
		var lower := LowerCase('');
		var trimmed := Trim('');
	`
	expectNoErrors(t, input)
}

func TestBuiltinString_SpecialCharacters(t *testing.T) {
	input := `
		var s := 'Line1' + #13#10 + 'Line2';
		var len := Length(s);
	`
	expectNoErrors(t, input)
}
