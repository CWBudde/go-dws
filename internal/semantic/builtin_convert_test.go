package semantic

import (
	"testing"
)

// ============================================================================
// Built-in Conversion Functions Tests
// ============================================================================
// These tests cover the built-in type conversion functions to improve
// coverage of analyze_builtin_convert.go (currently at 0-40% coverage)

// IntToStr function tests
func TestBuiltinIntToStr_Positive(t *testing.T) {
	input := `
		var s := IntToStr(42);
	`
	expectNoErrors(t, input)
}

func TestBuiltinIntToStr_Negative(t *testing.T) {
	input := `
		var s := IntToStr(-42);
	`
	expectNoErrors(t, input)
}

func TestBuiltinIntToStr_Zero(t *testing.T) {
	input := `
		var s := IntToStr(0);
	`
	expectNoErrors(t, input)
}

func TestBuiltinIntToStr_InvalidType(t *testing.T) {
	input := `
		var s := IntToStr('hello');
	`
	expectError(t, input, "integer")
}

// IntToBin function tests
func TestBuiltinIntToBin_Positive(t *testing.T) {
	input := `
		var s := IntToBin(42);
	`
	expectNoErrors(t, input)
}

func TestBuiltinIntToBin_WithDigits(t *testing.T) {
	input := `
		var s := IntToBin(42, 8);
	`
	expectNoErrors(t, input)
}

func TestBuiltinIntToBin_Zero(t *testing.T) {
	input := `
		var s := IntToBin(0);
	`
	expectNoErrors(t, input)
}

// IntToHex function tests
func TestBuiltinIntToHex_Positive(t *testing.T) {
	input := `
		var s := IntToHex(255);
	`
	expectNoErrors(t, input)
}

func TestBuiltinIntToHex_WithDigits(t *testing.T) {
	input := `
		var s := IntToHex(255, 4);
	`
	expectNoErrors(t, input)
}

func TestBuiltinIntToHex_Zero(t *testing.T) {
	input := `
		var s := IntToHex(0);
	`
	expectNoErrors(t, input)
}

// StrToInt function tests
func TestBuiltinStrToInt_Positive(t *testing.T) {
	input := `
		var n := StrToInt('42');
	`
	expectNoErrors(t, input)
}

func TestBuiltinStrToInt_Negative(t *testing.T) {
	input := `
		var n := StrToInt('-42');
	`
	expectNoErrors(t, input)
}

func TestBuiltinStrToInt_InvalidString(t *testing.T) {
	// Should analyze without error (runtime check)
	input := `
		var n := StrToInt('not a number');
	`
	expectNoErrors(t, input)
}

func TestBuiltinStrToInt_InvalidType(t *testing.T) {
	input := `
		var n := StrToInt(42);
	`
	expectError(t, input, "string")
}

// FloatToStr function tests
func TestBuiltinFloatToStr_Basic(t *testing.T) {
	input := `
		var s := FloatToStr(3.14);
	`
	expectNoErrors(t, input)
}

func TestBuiltinFloatToStr_Negative(t *testing.T) {
	input := `
		var s := FloatToStr(-2.71828);
	`
	expectNoErrors(t, input)
}

func TestBuiltinFloatToStr_Zero(t *testing.T) {
	input := `
		var s := FloatToStr(0.0);
	`
	expectNoErrors(t, input)
}

// FloatToStrF function tests
func TestBuiltinFloatToStrF_Fixed(t *testing.T) {
	input := `
		var s := FloatToStrF(3.14159, 2);
	`
	expectNoErrors(t, input)
}

func TestBuiltinFloatToStrF_WithFormat(t *testing.T) {
	input := `
		var s := FloatToStrF(1234.5678, 2);
	`
	expectNoErrors(t, input)
}

func TestBuiltinFloatToStrF_InvalidArgCount(t *testing.T) {
	input := `
		var s := FloatToStrF(3.14);
	`
	expectError(t, input, "argument")
}

// StrToFloat function tests
func TestBuiltinStrToFloat_Basic(t *testing.T) {
	input := `
		var f := StrToFloat('3.14');
	`
	expectNoErrors(t, input)
}

func TestBuiltinStrToFloat_Negative(t *testing.T) {
	input := `
		var f := StrToFloat('-2.71828');
	`
	expectNoErrors(t, input)
}

func TestBuiltinStrToFloat_Integer(t *testing.T) {
	input := `
		var f := StrToFloat('42');
	`
	expectNoErrors(t, input)
}

func TestBuiltinStrToFloat_Scientific(t *testing.T) {
	input := `
		var f := StrToFloat('1.23e-4');
	`
	expectNoErrors(t, input)
}

func TestBuiltinStrToFloat_InvalidString(t *testing.T) {
	// Should analyze without error (runtime check)
	input := `
		var f := StrToFloat('not a number');
	`
	expectNoErrors(t, input)
}

// BoolToStr function tests
func TestBuiltinBoolToStr_True(t *testing.T) {
	input := `
		var s := BoolToStr(True);
	`
	expectNoErrors(t, input)
}

func TestBuiltinBoolToStr_False(t *testing.T) {
	input := `
		var s := BoolToStr(False);
	`
	expectNoErrors(t, input)
}

func TestBuiltinBoolToStr_Expression(t *testing.T) {
	input := `
		var s := BoolToStr(5 > 3);
	`
	expectNoErrors(t, input)
}

func TestBuiltinBoolToStr_InvalidType(t *testing.T) {
	input := `
		var s := BoolToStr(42);
	`
	expectError(t, input, "boolean")
}

// StrToBool function tests
func TestBuiltinStrToBool_True(t *testing.T) {
	input := `
		var b := StrToBool('True');
	`
	expectNoErrors(t, input)
}

func TestBuiltinStrToBool_False(t *testing.T) {
	input := `
		var b := StrToBool('False');
	`
	expectNoErrors(t, input)
}

func TestBuiltinStrToBool_One(t *testing.T) {
	input := `
		var b := StrToBool('1');
	`
	expectNoErrors(t, input)
}

func TestBuiltinStrToBool_Zero(t *testing.T) {
	input := `
		var b := StrToBool('0');
	`
	expectNoErrors(t, input)
}

func TestBuiltinStrToBool_InvalidString(t *testing.T) {
	// Should analyze without error (runtime check)
	input := `
		var b := StrToBool('maybe');
	`
	expectNoErrors(t, input)
}

// Chr function tests
func TestBuiltinChr_Basic(t *testing.T) {
	input := `
		var c := Chr(65);
	`
	expectNoErrors(t, input)
}

func TestBuiltinChr_Zero(t *testing.T) {
	input := `
		var c := Chr(0);
	`
	expectNoErrors(t, input)
}

func TestBuiltinChr_HighValue(t *testing.T) {
	input := `
		var c := Chr(255);
	`
	expectNoErrors(t, input)
}

func TestBuiltinChr_InvalidType(t *testing.T) {
	input := `
		var c := Chr('A');
	`
	expectError(t, input, "integer")
}

// VarToStr function tests
func TestBuiltinVarToStr_String(t *testing.T) {
	input := `
		var v: Variant := 'hello';
		var s := VarToStr(v);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVarToStr_Integer(t *testing.T) {
	input := `
		var v: Variant := 42;
		var s := VarToStr(v);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVarToStr_Float(t *testing.T) {
	input := `
		var v: Variant := 3.14;
		var s := VarToStr(v);
	`
	expectNoErrors(t, input)
}

// Combined conversion tests
func TestBuiltinConvert_RoundTrip(t *testing.T) {
	input := `
		var n := 42;
		var s := IntToStr(n);
		var n2 := StrToInt(s);
	`
	expectNoErrors(t, input)
}

func TestBuiltinConvert_FloatRoundTrip(t *testing.T) {
	input := `
		var f := 3.14159;
		var s := FloatToStr(f);
		var f2 := StrToFloat(s);
	`
	expectNoErrors(t, input)
}

func TestBuiltinConvert_BoolRoundTrip(t *testing.T) {
	input := `
		var b := True;
		var s := BoolToStr(b);
		var b2 := StrToBool(s);
	`
	expectNoErrors(t, input)
}

func TestBuiltinConvert_HexConversion(t *testing.T) {
	input := `
		var n := 255;
		var hex := IntToHex(n, 2);
		PrintLn(hex);
	`
	expectNoErrors(t, input)
}

func TestBuiltinConvert_BinaryConversion(t *testing.T) {
	input := `
		var n := 42;
		var bin := IntToBin(n, 8);
		PrintLn(bin);
	`
	expectNoErrors(t, input)
}

func TestBuiltinConvert_InExpressions(t *testing.T) {
	input := `
		var result := 'Value: ' + IntToStr(42) + ', Float: ' + FloatToStr(3.14);
	`
	expectNoErrors(t, input)
}

func TestBuiltinConvert_AsParameters(t *testing.T) {
	input := `
		procedure LogValue(msg: String; value: Integer);
		begin
			PrintLn(msg + ': ' + IntToStr(value));
		end;

		LogValue('Count', 42);
	`
	expectNoErrors(t, input)
}

// Edge cases
func TestBuiltinConvert_EmptyString(t *testing.T) {
	// Should analyze without error (runtime check)
	input := `
		var n := StrToInt('');
	`
	expectNoErrors(t, input)
}

func TestBuiltinConvert_MaxInt(t *testing.T) {
	input := `
		var s := IntToStr(2147483647);
	`
	expectNoErrors(t, input)
}

func TestBuiltinConvert_MinInt(t *testing.T) {
	input := `
		var s := IntToStr(-2147483648);
	`
	expectNoErrors(t, input)
}

func TestBuiltinConvert_LargeFloat(t *testing.T) {
	input := `
		var s := FloatToStr(1.23456789012345e100);
	`
	expectNoErrors(t, input)
}

func TestBuiltinConvert_SmallFloat(t *testing.T) {
	input := `
		var s := FloatToStr(1.23456789012345e-100);
	`
	expectNoErrors(t, input)
}

func TestBuiltinConvert_CharacterCodes(t *testing.T) {
	input := `
		var a := Chr(65);
		var z := Chr(90);
		var zero := Chr(48);
		var nine := Chr(57);
	`
	expectNoErrors(t, input)
}

func TestBuiltinConvert_InFunction(t *testing.T) {
	input := `
		function FormatNumber(n: Integer): String;
		begin
			Result := 'Number: ' + IntToStr(n);
		end;

		var s := FormatNumber(42);
	`
	expectNoErrors(t, input)
}

func TestBuiltinConvert_MultipleConversions(t *testing.T) {
	input := `
		var n := 42;
		var s1 := IntToStr(n);
		var s2 := IntToHex(n);
		var s3 := IntToBin(n);
		PrintLn(s1 + ' = 0x' + s2 + ' = 0b' + s3);
	`
	expectNoErrors(t, input)
}
