package interp

import (
	"strings"
	"testing"
)

// ============================================================================
// Subrange Type Runtime Tests
// ============================================================================

// TestSubrangeTypeDeclaration tests declaring subrange types at runtime
func TestSubrangeTypeDeclaration(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{
			name:      "Basic digit subrange",
			input:     "type TDigit = 0..9; PrintLn('ok');",
			expectErr: false,
		},
		{
			name:      "Percentage subrange",
			input:     "type TPercent = 0..100; PrintLn('ok');",
			expectErr: false,
		},
		{
			name:      "Negative range",
			input:     "type TTemperature = -40..50; PrintLn('ok');",
			expectErr: false,
		},
		{
			name:      "Error: low > high",
			input:     "type TBad = 10..5;",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := testEval(tt.input)
			hasError := isError(val)

			if hasError != tt.expectErr {
				if tt.expectErr {
					t.Errorf("Expected error but got none")
				} else {
					t.Errorf("Unexpected error: %v", val)
				}
			}
		})
	}
}

// TestSubrangeVariableDeclaration tests declaring variables with subrange types
func TestSubrangeVariableDeclaration(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{
			name: "Variable without initializer",
			input: `
				type TDigit = 0..9;
				var digit: TDigit;
				PrintLn('ok');
			`,
			expectErr: false,
		},
		{
			name: "Variable with valid initializer",
			input: `
				type TDigit = 0..9;
				var digit: TDigit := 5;
				PrintLn('ok');
			`,
			expectErr: false,
		},
		{
			name: "Variable with out-of-range initializer",
			input: `
				type TDigit = 0..9;
				var digit: TDigit := 99;
			`,
			expectErr: true,
		},
		{
			name: "Variable with negative initializer in negative range",
			input: `
				type TTemp = -40..50;
				var temp: TTemp := -20;
				PrintLn('ok');
			`,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := testEval(tt.input)
			hasError := isError(val)

			if hasError != tt.expectErr {
				if tt.expectErr {
					t.Errorf("Expected error but got none")
				} else {
					t.Errorf("Unexpected error: %v", val)
				}
			}
		})
	}
}

// TestSubrangeAssignment tests assigning values to subrange variables
func TestSubrangeAssignment(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		errorMsg  string
		expectErr bool
	}{
		{
			name: "Valid assignment within range",
			input: `
				type TDigit = 0..9;
				var digit: TDigit;
				digit := 5;
				PrintLn('ok');
			`,
			expectErr: false,
		},
		{
			name: "Assignment at low bound",
			input: `
				type TDigit = 0..9;
				var digit: TDigit;
				digit := 0;
				PrintLn('ok');
			`,
			expectErr: false,
		},
		{
			name: "Assignment at high bound",
			input: `
				type TDigit = 0..9;
				var digit: TDigit;
				digit := 9;
				PrintLn('ok');
			`,
			expectErr: false,
		},
		{
			name: "Assignment below range",
			input: `
				type TDigit = 0..9;
				var digit: TDigit;
				digit := -1;
			`,
			expectErr: true,
			errorMsg:  "out of range",
		},
		{
			name: "Assignment above range",
			input: `
				type TDigit = 0..9;
				var digit: TDigit;
				digit := 10;
			`,
			expectErr: true,
			errorMsg:  "out of range",
		},
		{
			name: "Assignment far above range",
			input: `
				type TDigit = 0..9;
				var digit: TDigit;
				digit := 999;
			`,
			expectErr: true,
			errorMsg:  "out of range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := testEval(tt.input)
			hasError := isError(val)

			if hasError != tt.expectErr {
				if tt.expectErr {
					t.Errorf("Expected error but got none")
				} else {
					t.Errorf("Unexpected error: %v", val)
				}
			}

			if tt.expectErr && tt.errorMsg != "" {
				if errVal, ok := val.(*ErrorValue); ok {
					if !strings.Contains(errVal.Message, tt.errorMsg) {
						t.Errorf("Expected error containing %q, got: %v", tt.errorMsg, errVal.Message)
					}
				}
			}
		})
	}
}

// TestSubrangeCrossAssignment tests assigning between different subrange types
func TestSubrangeCrossAssignment(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{
			name: "Assign subrange to subrange (same type)",
			input: `
				type TDigit = 0..9;
				var d1: TDigit := 5;
				var d2: TDigit;
				d2 := d1;
				PrintLn('ok');
			`,
			expectErr: false,
		},
		{
			name: "Assign subrange to subrange (different type, valid value)",
			input: `
				type TDigit = 0..9;
				type TSmallDigit = 0..5;
				var digit: TDigit := 3;
				var small: TSmallDigit;
				small := digit;
				PrintLn('ok');
			`,
			expectErr: false,
		},
		{
			name: "Assign subrange to subrange (different type, invalid value)",
			input: `
				type TDigit = 0..9;
				type TSmallDigit = 0..5;
				var digit: TDigit := 8;
				var small: TSmallDigit;
				small := digit;
			`,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := testEval(tt.input)
			hasError := isError(val)

			if hasError != tt.expectErr {
				if tt.expectErr {
					t.Errorf("Expected error but got none")
				} else {
					t.Errorf("Unexpected error: %v", val)
				}
			}
		})
	}
}

// TestSubrangeToIntegerCoercion tests that subrange values can be assigned to integer variables
func TestSubrangeToIntegerCoercion(t *testing.T) {
	input := `
		type TDigit = 0..9;
		var digit: TDigit := 5;
		var number: Integer;
		number := digit;
		PrintLn(number);
	`

	_, output := testEvalWithOutput(input)
	expected := "5\n"
	if output != expected {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}

// TestIntegerToSubrangeCoercion tests that integer values can be assigned to subrange variables (with validation)
func TestIntegerToSubrangeCoercion(t *testing.T) {
	input := `
		type TDigit = 0..9;
		var number: Integer := 7;
		var digit: TDigit;
		digit := number;
		PrintLn(digit);
	`

	_, output := testEvalWithOutput(input)
	expected := "7\n"
	if output != expected {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}

// TestSubrangeInFunction tests using subrange types in functions
func TestSubrangeInFunction(t *testing.T) {
	input := `
		type TDigit = 0..9;

		function GetDigit(): TDigit;
		begin
			Result := 5;
		end;

		var d: TDigit;
		d := GetDigit();
		PrintLn(d);
	`

	_, output := testEvalWithOutput(input)
	expected := "5\n"
	if output != expected {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}

// TestSubrangeInLoop tests using subrange variables in loops
func TestSubrangeInLoop(t *testing.T) {
	input := `
		type TDigit = 0..9;
		var digit: TDigit;
		var i: Integer;

		for i := 0 to 3 do
		begin
			digit := i;
			Print(digit);
			Print(' ');
		end;
	`

	_, output := testEvalWithOutput(input)
	expected := "0 1 2 3 "
	if output != expected {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}

// TestSubrangeRuntimeValidation tests that validation happens at runtime
func TestSubrangeRuntimeValidation(t *testing.T) {
	input := `
		type TDigit = 0..9;
		var digit: TDigit;
		var x: Integer := 15;
		digit := x; // Should fail at runtime
	`

	val := testEval(input)
	if !isError(val) {
		t.Error("Expected runtime error for out-of-range value, got none")
	}

	if errVal, ok := val.(*ErrorValue); ok {
		if !strings.Contains(errVal.Message, "out of range") {
			t.Errorf("Expected error containing 'out of range', got: %v", errVal.Message)
		}
	}
}

// TestMultipleSubrangeTypes tests using multiple subrange types in one program
func TestMultipleSubrangeTypes(t *testing.T) {
	input := `
		type TDigit = 0..9;
		type TPercent = 0..100;
		type TTemperature = -40..50;

		var digit: TDigit := 5;
		var percent: TPercent := 75;
		var temp: TTemperature := -10;

		PrintLn(digit);
		PrintLn(percent);
		PrintLn(temp);
	`

	_, output := testEvalWithOutput(input)
	expected := "5\n75\n-10\n"
	if output != expected {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}
