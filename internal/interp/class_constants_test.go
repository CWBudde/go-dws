package interp

import (
	"testing"
)

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

	result, output := testEvalWithOutput(source)
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

	result, output := testEvalWithOutput(source)
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

	result, output := testEvalWithOutput(source)
	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "42\n42\n142\n"
	if output != expected {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expected, output)
	}
}

// TestClassConstantInInheritedMethod tests that class constants can be accessed
// within inherited method calls.
func TestClassConstantInInheritedMethod(t *testing.T) {
	source := `
		type TBase = class
			class const BASE_LIMIT = 100;
			FValue: Integer;

			constructor Create;
			function Validate: Boolean;
		end;

		type TDerived = class(TBase)
			class const DERIVED_LIMIT = 200;

			function Validate: Boolean; override;
		end;

		constructor TBase.Create;
		begin
			FValue := 0;
		end;

		function TBase.Validate: Boolean;
		begin
			// Base method should access BASE_LIMIT
			Result := FValue < BASE_LIMIT;
		end;

		function TDerived.Validate: Boolean;
		begin
			// First check derived limit
			if FValue >= DERIVED_LIMIT then
			begin
				Result := False;
			end
			else
			begin
				// Call parent's validation which uses BASE_LIMIT
				Result := inherited Validate;
			end;
		end;

		var
			d: TDerived;
		begin
			d := TDerived.Create;
			d.FValue := 50;
			PrintLn(d.Validate);    // Should print: True (50 < 100)

			d.FValue := 150;
			PrintLn(d.Validate);    // Should print: False (150 >= 100)

			d.FValue := 250;
			PrintLn(d.Validate);    // Should print: False (250 >= 200)
		end.
	`

	result, output := testEvalWithOutput(source)
	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "True\nFalse\nFalse\n"
	if output != expected {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expected, output)
	}
}
