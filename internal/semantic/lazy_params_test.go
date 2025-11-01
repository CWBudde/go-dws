package semantic

import (
	"testing"
)

// ============================================================================
// Lazy Parameter Declaration Tests
// ============================================================================

func TestLazyParameterDeclaration(t *testing.T) {
	input := `
		function Test(lazy x: Integer): Integer;
		begin
			Result := x + 1;
		end;
	`
	expectNoErrors(t, input)
}

func TestLazyParameterWithRegularParameter(t *testing.T) {
	input := `
		procedure Log(level: Integer; lazy msg: String);
		begin
			if level > 0 then
				PrintLn(msg);
		end;
	`
	expectNoErrors(t, input)
}

func TestMultipleLazyParameters(t *testing.T) {
	input := `
		function IfThen(cond: Boolean; lazy trueVal: Integer; lazy falseVal: Integer): Integer;
		begin
			if cond then
				Result := trueVal
			else
				Result := falseVal;
		end;
	`
	expectNoErrors(t, input)
}

func TestLazyParameterWithSharedType(t *testing.T) {
	input := `
		function If(cond: Boolean; lazy trueVal, falseVal: Integer): Integer;
		begin
			if cond then
				Result := trueVal
			else
				Result := falseVal;
		end;
	`
	expectNoErrors(t, input)
}

// ============================================================================
// Lazy Parameter Error Tests
// ============================================================================

// Note: lazy+var conflict is already validated at the parser level
// (see TestParameterErrors in parser_test.go), so we don't test it here

// ============================================================================
// Lazy Parameter Call-Site Tests
// ============================================================================

func TestLazyParameterCallSite(t *testing.T) {
	input := `
		function IfThen(cond: Boolean; lazy trueVal: Integer; lazy falseVal: Integer): Integer;
		begin
			if cond then
				Result := trueVal
			else
				Result := falseVal;
		end;

		var x: Integer := IfThen(true, 10, 20);
	`
	expectNoErrors(t, input)
}

func TestLazyParameterWithExpression(t *testing.T) {
	input := `
		function Test(lazy x: Integer): Integer;
		begin
			Result := x;
		end;

		var a: Integer := 5;
		var b: Integer := 3;
		var result: Integer := Test(a + b * 2);
	`
	expectNoErrors(t, input)
}

func TestLazyParameterTypeChecking(t *testing.T) {
	input := `
		function Test(lazy x: Integer): Integer;
		begin
			Result := x;
		end;

		var result: Integer := Test('hello');
	`
	expectError(t, input, "has type String, expected Integer")
}

func TestLazyParameterWithVariableReference(t *testing.T) {
	input := `
		function Compute(lazy expr: Float): Float;
		begin
			Result := expr * 2.0;
		end;

		var value: Float := 3.14;
		var result: Float := Compute(value);
	`
	expectNoErrors(t, input)
}

func TestLazyParameterMixedWithRegular(t *testing.T) {
	input := `
		function Process(name: String; lazy value: Integer): Integer;
		begin
			PrintLn(name);
			Result := value + 1;
		end;

		var x: Integer := Process('test', 42);
	`
	expectNoErrors(t, input)
}

func TestLazyParameterCallSiteTypeError(t *testing.T) {
	input := `
		procedure Log(level: Integer; lazy msg: String);
		begin
			if level > 0 then
				PrintLn(msg);
		end;

		begin
			Log(1, 42);
		end;
	`
	expectError(t, input, "has type Integer, expected String")
}
