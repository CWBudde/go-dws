package semantic

import (
	"testing"
)

// ============================================================================
// Built-in Variant Functions Tests
// ============================================================================
// These tests cover the built-in variant manipulation functions to improve
// coverage of analyze_builtin_variant.go (currently at 0% coverage)

// VarType function tests
func TestBuiltinVarType_Integer(t *testing.T) {
	input := `
		var v: Variant := 42;
		var vt := VarType(v);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVarType_String(t *testing.T) {
	input := `
		var v: Variant := 'hello';
		var vt := VarType(v);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVarType_Float(t *testing.T) {
	input := `
		var v: Variant := 3.14;
		var vt := VarType(v);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVarType_Boolean(t *testing.T) {
	input := `
		var v: Variant := True;
		var vt := VarType(v);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVarType_InvalidType(t *testing.T) {
	input := `
		var vt := VarType(42);
	`
	expectError(t, input, "variant")
}

// VarIsNull function tests
func TestBuiltinVarIsNull_Null(t *testing.T) {
	input := `
		var v: Variant;
		var isNull := VarIsNull(v);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVarIsNull_NotNull(t *testing.T) {
	input := `
		var v: Variant := 42;
		var isNull := VarIsNull(v);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVarIsNull_InvalidType(t *testing.T) {
	input := `
		var isNull := VarIsNull('hello');
	`
	expectError(t, input, "variant")
}

// VarIsEmpty function tests
func TestBuiltinVarIsEmpty_Empty(t *testing.T) {
	input := `
		var v: Variant;
		var isEmpty := VarIsEmpty(v);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVarIsEmpty_NotEmpty(t *testing.T) {
	input := `
		var v: Variant := 0;
		var isEmpty := VarIsEmpty(v);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVarIsEmpty_String(t *testing.T) {
	input := `
		var v: Variant := '';
		var isEmpty := VarIsEmpty(v);
	`
	expectNoErrors(t, input)
}

// VarIsNumeric function tests
func TestBuiltinVarIsNumeric_Integer(t *testing.T) {
	input := `
		var v: Variant := 42;
		var isNumeric := VarIsNumeric(v);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVarIsNumeric_Float(t *testing.T) {
	input := `
		var v: Variant := 3.14;
		var isNumeric := VarIsNumeric(v);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVarIsNumeric_String(t *testing.T) {
	input := `
		var v: Variant := 'hello';
		var isNumeric := VarIsNumeric(v);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVarIsNumeric_NumericString(t *testing.T) {
	input := `
		var v: Variant := '123';
		var isNumeric := VarIsNumeric(v);
	`
	expectNoErrors(t, input)
}

// VarToInt function tests
func TestBuiltinVarToInt_Integer(t *testing.T) {
	input := `
		var v: Variant := 42;
		var n := VarToInt(v);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVarToInt_Float(t *testing.T) {
	input := `
		var v: Variant := 3.14;
		var n := VarToInt(v);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVarToInt_String(t *testing.T) {
	// Should analyze without error (runtime check)
	input := `
		var v: Variant := '42';
		var n := VarToInt(v);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVarToInt_InvalidType(t *testing.T) {
	input := `
		var n := VarToInt(42);
	`
	expectError(t, input, "variant")
}

// VarToFloat function tests
func TestBuiltinVarToFloat_Float(t *testing.T) {
	input := `
		var v: Variant := 3.14;
		var f := VarToFloat(v);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVarToFloat_Integer(t *testing.T) {
	input := `
		var v: Variant := 42;
		var f := VarToFloat(v);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVarToFloat_String(t *testing.T) {
	// Should analyze without error (runtime check)
	input := `
		var v: Variant := '3.14';
		var f := VarToFloat(v);
	`
	expectNoErrors(t, input)
}

// VarAsType function tests
func TestBuiltinVarAsType_Integer(t *testing.T) {
	input := `
		var v: Variant := 42;
		var n := VarAsType(v, 'Integer');
	`
	expectNoErrors(t, input)
}

func TestBuiltinVarAsType_String(t *testing.T) {
	input := `
		var v: Variant := 'hello';
		var s := VarAsType(v, 'String');
	`
	expectNoErrors(t, input)
}

func TestBuiltinVarAsType_Float(t *testing.T) {
	input := `
		var v: Variant := 3.14;
		var f := VarAsType(v, 'Float');
	`
	expectNoErrors(t, input)
}

func TestBuiltinVarAsType_InvalidArgCount(t *testing.T) {
	input := `
		var v: Variant;
		var result := VarAsType(v);
	`
	expectError(t, input, "argument")
}

// Combined variant operations tests
func TestBuiltinVariant_TypeChecking(t *testing.T) {
	input := `
		var v: Variant := 42;
		var vt := VarType(v);
		var isNull := VarIsNull(v);
		var isEmpty := VarIsEmpty(v);
		var isNumeric := VarIsNumeric(v);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVariant_Conversion(t *testing.T) {
	input := `
		var v: Variant := 3.14;
		var n := VarToInt(v);
		var f := VarToFloat(v);
		var s := VarToStr(v);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVariant_InFunction(t *testing.T) {
	input := `
		function ProcessVariant(v: Variant): String;
		begin
			if VarIsNull(v) then
				Result := 'null'
			else if VarIsNumeric(v) then
				Result := 'numeric: ' + VarToStr(v)
			else
				Result := 'other: ' + VarToStr(v);
		end;

		var result := ProcessVariant(42);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVariant_Conditionals(t *testing.T) {
	input := `
		var v: Variant := 42;
		if VarIsNumeric(v) then
		begin
			var n := VarToInt(v);
			PrintLn(n);
		end;
	`
	expectNoErrors(t, input)
}

func TestBuiltinVariant_ArrayOfVariants(t *testing.T) {
	input := `
		var arr: array of Variant;
		SetLength(arr, 3);
		arr[0] := 42;
		arr[1] := 'hello';
		arr[2] := 3.14;

		for i := 0 to 2 do
		begin
			var vt := VarType(arr[i]);
		end;
	`
	expectNoErrors(t, input)
}

func TestBuiltinVariant_TypeAssertion(t *testing.T) {
	input := `
		var v: Variant := 'hello';
		if VarType(v) = VarType('string') then
			PrintLn('It is a string');
	`
	expectNoErrors(t, input)
}

// Edge cases
func TestBuiltinVariant_UninitializedVariant(t *testing.T) {
	input := `
		var v: Variant;
		var isNull := VarIsNull(v);
		var isEmpty := VarIsEmpty(v);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVariant_ZeroValue(t *testing.T) {
	input := `
		var v: Variant := 0;
		var isNull := VarIsNull(v);
		var isEmpty := VarIsEmpty(v);
		var isNumeric := VarIsNumeric(v);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVariant_EmptyString(t *testing.T) {
	input := `
		var v: Variant := '';
		var isNull := VarIsNull(v);
		var isEmpty := VarIsEmpty(v);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVariant_BooleanValues(t *testing.T) {
	input := `
		var vTrue: Variant := True;
		var vFalse: Variant := False;
		var vtTrue := VarType(vTrue);
		var vtFalse := VarType(vFalse);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVariant_ConversionRoundTrip(t *testing.T) {
	input := `
		var v1: Variant := 42;
		var n := VarToInt(v1);
		var v2: Variant := n;
		var n2 := VarToInt(v2);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVariant_WithJSON(t *testing.T) {
	input := `
		var v: Variant := ParseJSON('{"test": 123}');
		var vt := VarType(v);
		var isNull := VarIsNull(v);
	`
	expectNoErrors(t, input)
}

func TestBuiltinVariant_AsParameters(t *testing.T) {
	input := `
		procedure LogVariant(v: Variant);
		begin
			PrintLn('Type: ' + IntToStr(VarType(v)));
			PrintLn('Value: ' + VarToStr(v));
		end;

		var v: Variant := 'test';
		LogVariant(v);
	`
	expectNoErrors(t, input)
}
