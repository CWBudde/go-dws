package semantic

import (
	"testing"
)

// ============================================================================
// Expression Type Checking Tests
// ============================================================================

func TestArithmeticExpressions(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"integer addition", "var x := 3 + 5;", true},
		{"float addition", "var x := 3.14 + 2.86;", true},
		{"mixed addition", "var x := 3 + 2.5;", true}, // Integer + Float -> Float
		{"integer subtraction", "var x := 10 - 5;", true},
		{"integer multiplication", "var x := 4 * 5;", true},
		{"float division", "var x := 10.0 / 2.0;", true},
		{"string + number", "var x: String; x := 'hello' + 5;", false},
		{"number + string", "var x: Integer; x := 5 + 'hello';", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				expectNoErrors(t, tt.input)
			} else {
				_, err := analyzeSource(t, tt.input)
				if err == nil {
					t.Errorf("expected error for: %s", tt.input)
				}
			}
		})
	}
}

func TestStringConcatenation(t *testing.T) {
	input := `var s := 'hello' + ' ' + 'world';`
	expectNoErrors(t, input)
}

func TestStringConcatenationError(t *testing.T) {
	input := `var s := 'hello' + 42;`
	expectError(t, input, "string concatenation requires both operands to be strings")
}

func TestIntegerOperations(t *testing.T) {
	input := `
		var a := 10 div 3;
		var b := 10 mod 3;
	`
	expectNoErrors(t, input)
}

func TestIntegerOperationsError(t *testing.T) {
	input := `var x := 3.14 div 2.0;`
	expectError(t, input, "requires integer operands")
}

func TestComparisonOperations(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"integer equality", "var b := 3 = 5;", true},
		{"integer inequality", "var b := 3 <> 5;", true},
		{"integer less than", "var b := 3 < 5;", true},
		{"integer greater than", "var b := 3 > 5;", true},
		{"string equality", "var b := 'a' = 'b';", true},
		{"string comparison", "var b := 'a' < 'b';", true},
		{"float comparison", "var b := 3.14 > 2.86;", true},
		{"mixed comparison", "var b := 3 < 2.5;", true}, // Integer vs Float
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				expectNoErrors(t, tt.input)
			} else {
				_, err := analyzeSource(t, tt.input)
				if err == nil {
					t.Errorf("expected error for: %s", tt.input)
				}
			}
		})
	}
}

func TestLogicalOperations(t *testing.T) {
	input := `
		var a := true and false;
		var b := true or false;
		var c := true xor false;
		var d := not true;
	`
	expectNoErrors(t, input)
}

func TestLogicalOperationsError(t *testing.T) {
	tests := []string{
		"var x := 3 and 5;",
		"var x := 'hello' or 'world';",
		"var x := not 42;",
	}

	for _, input := range tests {
		expectError(t, input, "boolean")
	}
}

func TestUnaryOperations(t *testing.T) {
	input := `
		var a := -5;
		var b := +3.14;
		var c := not true;
	`
	expectNoErrors(t, input)
}

func TestUnaryOperationsError(t *testing.T) {
	tests := []struct {
		input string
		error string
	}{
		{"var x := -'hello';", "numeric operand"},
		{"var x := not 42;", "boolean operand"},
	}

	for _, tt := range tests {
		expectError(t, tt.input, tt.error)
	}
}

// ============================================================================
// Format Built-in Function Tests (Task 9.51a)
// ============================================================================

func TestFormatBuiltInFunction(t *testing.T) {
	// Test valid Format calls with a declared array variable
	input := `
		type TIntArray = array [0..10] of Integer;
		var arr: TIntArray;
		var result: String;
		begin
			result := Format('Hello %d', arr);
		end;
	`
	expectNoErrors(t, input)
}

func TestFormatWrongNumberOfArguments(t *testing.T) {
	tests := []string{
		`var result := Format('test');`,           // Only 1 argument
		`var result := Format('test', arr, 123);`, // 3 arguments
	}

	for _, input := range tests {
		expectError(t, input, "Format() expects exactly 2 arguments")
	}
}

func TestFormatWrongArgumentTypes(t *testing.T) {
	tests := []struct {
		input string
		error string
	}{
		{
			`var result := Format(42, arr);`, // First arg not string
			"Format() expects string as first argument",
		},
		{
			`var result := Format('test', 42);`, // Second arg not array
			"Format() expects array as second argument",
		},
	}

	for _, tt := range tests {
		expectError(t, tt.input, tt.error)
	}
}
