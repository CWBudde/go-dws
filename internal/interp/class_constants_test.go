package interp

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// Helper function to evaluate class constant tests with output capture
func testEvalClassConst(input string) (Value, string) {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		panic("Parser errors: " + joinParserErrorsNewline(p.Errors()))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)
	return result, buf.String()
}

// TestClassConstantInClassMethod tests that class constants can be accessed
// within class methods (static methods).
func TestClassConstantInClassMethod(t *testing.T) {
	source := `
		type TCounter = class
			class const MAX_COUNT = 100;
			class const MIN_COUNT = 0;
			class var FCount: Integer;

			class procedure PrintMax;
			class function GetMax: Integer;
			class function IsAtMax: Boolean;
		end;

		class procedure TCounter.PrintMax;
		begin
			PrintLn('MAX: ' + IntToStr(MAX_COUNT));
		end;

		class function TCounter.GetMax: Integer;
		begin
			Result := MAX_COUNT;
		end;

		class function TCounter.IsAtMax: Boolean;
		begin
			Result := FCount >= MAX_COUNT;
		end;

		begin
			TCounter.FCount := 50;
			PrintLn(TCounter.IsAtMax);  // Should print: False
			TCounter.FCount := 100;
			PrintLn(TCounter.IsAtMax);  // Should print: True
			PrintLn(TCounter.GetMax);   // Should print: 100
		end.
	`

	result, output := testEvalClassConst(source)
	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "False\nTrue\n100\n"
	if output != expected {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expected, output)
	}
}

// TestClassConstantInInstanceMethod tests that class constants can be accessed
// within instance methods.
func TestClassConstantInInstanceMethod(t *testing.T) {
	source := `
		type TValidator = class
			class const MAX_VALUE = 999;
			class const MIN_VALUE = 1;
			FValue: Integer;

			constructor Create(AValue: Integer);
			function IsValid: Boolean;
			function GetRange: String;
		end;

		constructor TValidator.Create(AValue: Integer);
		begin
			FValue := AValue;
		end;

		function TValidator.IsValid: Boolean;
		begin
			Result := (FValue >= MIN_VALUE) and (FValue <= MAX_VALUE);
		end;

		function TValidator.GetRange: String;
		begin
			Result := IntToStr(MIN_VALUE) + ' to ' + IntToStr(MAX_VALUE);
		end;

		var
			v1, v2: TValidator;
		begin
			v1 := TValidator.Create(50);
			v2 := TValidator.Create(1000);

			PrintLn(v1.IsValid);      // Should print: True
			PrintLn(v2.IsValid);      // Should print: False
			PrintLn(v1.GetRange);     // Should print: 1 to 999
		end.
	`

	result, output := testEvalClassConst(source)
	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "True\nFalse\n1 to 999\n"
	if output != expected {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expected, output)
	}
}

// TestClassConstantInheritance tests that class constants are inherited.
func TestClassConstantInheritance(t *testing.T) {
	source := `
		type TBase = class
			class const BASE_VALUE = 42;
			class function GetBaseValue: Integer;
		end;

		type TDerived = class(TBase)
			class const DERIVED_VALUE = 100;
			class function GetSum: Integer;
		end;

		class function TBase.GetBaseValue: Integer;
		begin
			Result := BASE_VALUE;
		end;

		class function TDerived.GetSum: Integer;
		begin
			// Can access both base and derived constants
			Result := BASE_VALUE + DERIVED_VALUE;
		end;

		begin
			PrintLn(TBase.GetBaseValue);      // Should print: 42
			PrintLn(TDerived.GetBaseValue);   // Should print: 42 (inherited)
			PrintLn(TDerived.GetSum);         // Should print: 142
		end.
	`

	result, output := testEvalClassConst(source)
	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "42\n42\n142\n"
	if output != expected {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expected, output)
	}
}
