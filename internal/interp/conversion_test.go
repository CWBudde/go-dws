package interp

import "testing"

// TestBuiltinIntToStr_BasicUsage tests IntToStr() with basic integer to string conversion.
// IntToStr(i) - converts integer to string representation
func TestBuiltinIntToStr_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Positive integer",
			input: `
begin
	IntToStr(42);
end
			`,
			expected: "42",
		},
		{
			name: "Negative integer",
			input: `
begin
	IntToStr(-123);
end
			`,
			expected: "-123",
		},
		{
			name: "Zero",
			input: `
begin
	IntToStr(0);
end
			`,
			expected: "0",
		},
		{
			name: "Large positive number",
			input: `
begin
	IntToStr(999999);
end
			`,
			expected: "999999",
		},
		{
			name: "Large negative number",
			input: `
begin
	IntToStr(-999999);
end
			`,
			expected: "-999999",
		},
		{
			name: "With variable",
			input: `
var x: Integer := 42;
begin
	IntToStr(x);
end
			`,
			expected: "42",
		},
		{
			name: "Single digit",
			input: `
begin
	IntToStr(5);
end
			`,
			expected: "5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			strVal, ok := result.(*StringValue)
			if !ok {
				t.Fatalf("result is not *StringValue. got=%T (%+v)", result, result)
			}

			if strVal.Value != tt.expected {
				t.Errorf("IntToStr() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinIntToStr_InExpressions tests using IntToStr() in various expressions.
func TestBuiltinIntToStr_InExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "IntToStr in string concatenation",
			input: `
var x: Integer := 42;
begin
	"Value: " + IntToStr(x);
end
			`,
			expected: "Value: 42",
		},
		{
			name: "IntToStr with arithmetic expression",
			input: `
var x: Integer := 10;
var y: Integer := 32;
begin
	IntToStr(x + y);
end
			`,
			expected: "42",
		},
		{
			name: "Multiple IntToStr calls",
			input: `
begin
	IntToStr(1) + ", " + IntToStr(2) + ", " + IntToStr(3);
end
			`,
			expected: "1, 2, 3",
		},
		{
			name: "IntToStr with negative result",
			input: `
var x: Integer := 10;
var y: Integer := 20;
begin
	IntToStr(x - y);
end
			`,
			expected: "-10",
		},
		{
			name: "Nested in function call",
			input: `
begin
	UpperCase(IntToStr(42));
end
			`,
			expected: "42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			strVal, ok := result.(*StringValue)
			if !ok {
				t.Fatalf("result is not *StringValue. got=%T (%+v)", result, result)
			}

			if strVal.Value != tt.expected {
				t.Errorf("IntToStr() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinIntToStr_ErrorCases tests error handling for IntToStr().
func TestBuiltinIntToStr_ErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "No arguments",
			input: `
begin
	IntToStr();
end
			`,
		},
		{
			name: "Too many arguments",
			input: `
begin
	IntToStr(42, 100);
end
			`,
		},
		{
			name: "Non-integer argument - string",
			input: `
var s: String := "hello";
begin
	IntToStr(s);
end
			`,
		},
		{
			name: "Non-integer argument - float",
			input: `
var f: Float := 3.14;
begin
	IntToStr(f);
end
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			// Should return an error
			if !isError(result) {
				t.Errorf("expected error for invalid IntToStr() call, got %T: %+v", result, result)
			}
		})
	}
}

// TestBuiltinStrToInt_BasicUsage tests StrToInt() with basic string to integer conversion.
// StrToInt(s) - converts string to integer, raises error on invalid input
func TestBuiltinStrToInt_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Positive integer string",
			input: `
begin
	StrToInt("42");
end
			`,
			expected: 42,
		},
		{
			name: "Negative integer string",
			input: `
begin
	StrToInt("-123");
end
			`,
			expected: -123,
		},
		{
			name: "Zero string",
			input: `
begin
	StrToInt("0");
end
			`,
			expected: 0,
		},
		{
			name: "Large positive number",
			input: `
begin
	StrToInt("999999");
end
			`,
			expected: 999999,
		},
		{
			name: "Large negative number",
			input: `
begin
	StrToInt("-999999");
end
			`,
			expected: -999999,
		},
		{
			name: "With variable",
			input: `
var s: String := "42";
begin
	StrToInt(s);
end
			`,
			expected: 42,
		},
		{
			name: "Single digit",
			input: `
begin
	StrToInt("5");
end
			`,
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			intVal, ok := result.(*IntegerValue)
			if !ok {
				t.Fatalf("result is not *IntegerValue. got=%T (%+v)", result, result)
			}

			if intVal.Value != tt.expected {
				t.Errorf("StrToInt() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinStrToInt_InExpressions tests using StrToInt() in various expressions.
func TestBuiltinStrToInt_InExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "StrToInt in arithmetic",
			input: `
begin
	StrToInt("10") + StrToInt("32");
end
			`,
			expected: 42,
		},
		{
			name: "StrToInt with multiplication",
			input: `
var s: String := "6";
begin
	StrToInt(s) * 7;
end
			`,
			expected: 42,
		},
		{
			name: "StrToInt result used in comparison",
			input: `
begin
	if StrToInt("42") = 42 then
		100
	else
		0;
end
			`,
			expected: 100,
		},
		{
			name: "Round-trip conversion",
			input: `
var x: Integer := 42;
begin
	StrToInt(IntToStr(x));
end
			`,
			expected: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			intVal, ok := result.(*IntegerValue)
			if !ok {
				t.Fatalf("result is not *IntegerValue. got=%T (%+v)", result, result)
			}

			if intVal.Value != tt.expected {
				t.Errorf("StrToInt() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinStrToInt_ErrorCases tests error handling for StrToInt().
func TestBuiltinStrToInt_ErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "No arguments",
			input: `
begin
	StrToInt();
end
			`,
		},
		{
			name: "Too many arguments",
			input: `
begin
	StrToInt("42", "100");
end
			`,
		},
		{
			name: "Non-string argument - integer",
			input: `
var x: Integer := 42;
begin
	StrToInt(x);
end
			`,
		},
		{
			name: "Invalid string - letters",
			input: `
begin
	StrToInt("hello");
end
			`,
		},
		{
			name: "Invalid string - mixed",
			input: `
begin
	StrToInt("42abc");
end
			`,
		},
		{
			name: "Empty string",
			input: `
begin
	StrToInt("");
end
			`,
		},
		{
			name: "Invalid string - whitespace only",
			input: `
begin
	StrToInt("   ");
end
			`,
		},
		{
			name: "Invalid string - float format",
			input: `
begin
	StrToInt("3.14");
end
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			// Should return an error
			if !isError(result) {
				t.Errorf("expected error for invalid StrToInt() call, got %T: %+v", result, result)
			}
		})
	}
}

// TestBuiltinFloatToStr_BasicUsage tests FloatToStr() with basic float to string conversion.
// FloatToStr(f) - converts float to string representation
func TestBuiltinFloatToStr_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Positive float",
			input: `
begin
	FloatToStr(3.14);
end
			`,
			expected: "3.14",
		},
		{
			name: "Negative float",
			input: `
begin
	FloatToStr(-2.5);
end
			`,
			expected: "-2.5",
		},
		{
			name: "Zero",
			input: `
begin
	FloatToStr(0.0);
end
			`,
			expected: "0",
		},
		{
			name: "Integer as float",
			input: `
begin
	FloatToStr(42.0);
end
			`,
			expected: "42",
		},
		{
			name: "Large float",
			input: `
begin
	FloatToStr(123456.789);
end
			`,
			expected: "123456.789",
		},
		{
			name: "Small decimal",
			input: `
begin
	FloatToStr(0.5);
end
			`,
			expected: "0.5",
		},
		{
			name: "With variable",
			input: `
var x: Float := 3.14159;
begin
	FloatToStr(x);
end
			`,
			expected: "3.14159",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			strVal, ok := result.(*StringValue)
			if !ok {
				t.Fatalf("result is not *StringValue. got=%T (%+v)", result, result)
			}

			if strVal.Value != tt.expected {
				t.Errorf("FloatToStr() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinFloatToStr_InExpressions tests using FloatToStr() in various expressions.
func TestBuiltinFloatToStr_InExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "FloatToStr in string concatenation",
			input: `
var x: Float := 3.14;
begin
	"Pi is approximately " + FloatToStr(x);
end
			`,
			expected: "Pi is approximately 3.14",
		},
		{
			name: "FloatToStr with arithmetic expression",
			input: `
var x: Float := 10.5;
var y: Float := 2.0;
begin
	FloatToStr(x * y);
end
			`,
			expected: "21",
		},
		{
			name: "Multiple FloatToStr calls",
			input: `
begin
	FloatToStr(1.1) + ", " + FloatToStr(2.2);
end
			`,
			expected: "1.1, 2.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			strVal, ok := result.(*StringValue)
			if !ok {
				t.Fatalf("result is not *StringValue. got=%T (%+v)", result, result)
			}

			if strVal.Value != tt.expected {
				t.Errorf("FloatToStr() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinFloatToStr_ErrorCases tests error handling for FloatToStr().
func TestBuiltinFloatToStr_ErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "No arguments",
			input: `
begin
	FloatToStr();
end
			`,
		},
		{
			name: "Too many arguments",
			input: `
begin
	FloatToStr(3.14, 2.0);
end
			`,
		},
		{
			name: "Non-float argument - string",
			input: `
var s: String := "hello";
begin
	FloatToStr(s);
end
			`,
		},
		{
			name: "Non-float argument - integer",
			input: `
var x: Integer := 42;
begin
	FloatToStr(x);
end
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			// Should return an error
			if !isError(result) {
				t.Errorf("expected error for invalid FloatToStr() call, got %T: %+v", result, result)
			}
		})
	}
}

// TestBuiltinStrToFloat_BasicUsage tests StrToFloat() with basic string to float conversion.
// StrToFloat(s) - converts string to float, raises error on invalid input
func TestBuiltinStrToFloat_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name: "Positive float string",
			input: `
begin
	StrToFloat("3.14");
end
			`,
			expected: 3.14,
		},
		{
			name: "Negative float string",
			input: `
begin
	StrToFloat("-2.5");
end
			`,
			expected: -2.5,
		},
		{
			name: "Zero string",
			input: `
begin
	StrToFloat("0.0");
end
			`,
			expected: 0.0,
		},
		{
			name: "Integer string",
			input: `
begin
	StrToFloat("42");
end
			`,
			expected: 42.0,
		},
		{
			name: "Small decimal",
			input: `
begin
	StrToFloat("0.5");
end
			`,
			expected: 0.5,
		},
		{
			name: "With variable",
			input: `
var s: String := "3.14159";
begin
	StrToFloat(s);
end
			`,
			expected: 3.14159,
		},
		{
			name: "Scientific notation",
			input: `
begin
	StrToFloat("1.5e2");
end
			`,
			expected: 150.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			floatVal, ok := result.(*FloatValue)
			if !ok {
				t.Fatalf("result is not *FloatValue. got=%T (%+v)", result, result)
			}

			if floatVal.Value != tt.expected {
				t.Errorf("StrToFloat() = %f, want %f", floatVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinStrToFloat_InExpressions tests using StrToFloat() in various expressions.
func TestBuiltinStrToFloat_InExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name: "StrToFloat in arithmetic",
			input: `
begin
	StrToFloat("3.14") + StrToFloat("2.86");
end
			`,
			expected: 6.0,
		},
		{
			name: "StrToFloat with multiplication",
			input: `
var s: String := "2.5";
begin
	StrToFloat(s) * 4.0;
end
			`,
			expected: 10.0,
		},
		{
			name: "Round-trip conversion",
			input: `
var x: Float := 3.14;
begin
	StrToFloat(FloatToStr(x));
end
			`,
			expected: 3.14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			floatVal, ok := result.(*FloatValue)
			if !ok {
				t.Fatalf("result is not *FloatValue. got=%T (%+v)", result, result)
			}

			if floatVal.Value != tt.expected {
				t.Errorf("StrToFloat() = %f, want %f", floatVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinStrToFloat_ErrorCases tests error handling for StrToFloat().
func TestBuiltinStrToFloat_ErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "No arguments",
			input: `
begin
	StrToFloat();
end
			`,
		},
		{
			name: "Too many arguments",
			input: `
begin
	StrToFloat("3.14", "2.0");
end
			`,
		},
		{
			name: "Non-string argument - integer",
			input: `
var x: Integer := 42;
begin
	StrToFloat(x);
end
			`,
		},
		{
			name: "Invalid string - letters",
			input: `
begin
	StrToFloat("hello");
end
			`,
		},
		{
			name: "Invalid string - mixed",
			input: `
begin
	StrToFloat("3.14abc");
end
			`,
		},
		{
			name: "Empty string",
			input: `
begin
	StrToFloat("");
end
			`,
		},
		{
			name: "Invalid string - whitespace only",
			input: `
begin
	StrToFloat("   ");
end
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			// Should return an error
			if !isError(result) {
				t.Errorf("expected error for invalid StrToFloat() call, got %T: %+v", result, result)
			}
		})
	}
}

// TestBuiltinBoolToStr_BasicUsage tests BoolToStr() with basic boolean to string conversion.
// BoolToStr(b) - converts boolean to string representation
// Task 9.245: BoolToStr built-in function
func TestBuiltinBoolToStr_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "True literal",
			input: `
begin
	BoolToStr(True);
end
			`,
			expected: "True",
		},
		{
			name: "False literal",
			input: `
begin
	BoolToStr(False);
end
			`,
			expected: "False",
		},
		{
			name: "With boolean variable - True",
			input: `
var b: Boolean := True;
begin
	BoolToStr(b);
end
			`,
			expected: "True",
		},
		{
			name: "With boolean variable - False",
			input: `
var b: Boolean := False;
begin
	BoolToStr(b);
end
			`,
			expected: "False",
		},
		{
			name: "Comparison expression - True",
			input: `
begin
	BoolToStr(5 > 3);
end
			`,
			expected: "True",
		},
		{
			name: "Comparison expression - False",
			input: `
begin
	BoolToStr(5 < 3);
end
			`,
			expected: "False",
		},
		{
			name: "Equality expression - True",
			input: `
begin
	BoolToStr(42 = 42);
end
			`,
			expected: "True",
		},
		{
			name: "Equality expression - False",
			input: `
begin
	BoolToStr(42 = 0);
end
			`,
			expected: "False",
		},
		{
			name: "Logical expression - True",
			input: `
begin
	BoolToStr(True and True);
end
			`,
			expected: "True",
		},
		{
			name: "Logical expression - False",
			input: `
begin
	BoolToStr(True and False);
end
			`,
			expected: "False",
		},
		{
			name: "Not expression - True",
			input: `
begin
	BoolToStr(not False);
end
			`,
			expected: "True",
		},
		{
			name: "Not expression - False",
			input: `
begin
	BoolToStr(not True);
end
			`,
			expected: "False",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			strVal, ok := result.(*StringValue)
			if !ok {
				t.Fatalf("result is not *StringValue. got=%T (%+v)", result, result)
			}

			if strVal.Value != tt.expected {
				t.Errorf("BoolToStr() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinBoolToStr_InExpressions tests using BoolToStr() in various expressions.
func TestBuiltinBoolToStr_InExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "BoolToStr in string concatenation",
			input: `
var b: Boolean := True;
begin
	"The value is: " + BoolToStr(b);
end
			`,
			expected: "The value is: True",
		},
		{
			name: "BoolToStr with comparison",
			input: `
var x: Integer := 10;
var y: Integer := 20;
begin
	BoolToStr(x < y) + " is correct";
end
			`,
			expected: "True is correct",
		},
		{
			name: "Multiple BoolToStr calls",
			input: `
begin
	BoolToStr(True) + ", " + BoolToStr(False);
end
			`,
			expected: "True, False",
		},
		{
			name: "BoolToStr nested in UpperCase",
			input: `
begin
	UpperCase(BoolToStr(True));
end
			`,
			expected: "TRUE",
		},
		{
			name: "BoolToStr nested in LowerCase",
			input: `
begin
	LowerCase(BoolToStr(False));
end
			`,
			expected: "false",
		},
		{
			name: "Complex boolean expression",
			input: `
var a: Integer := 5;
var b: Integer := 10;
var c: Integer := 15;
begin
	BoolToStr((a < b) and (b < c));
end
			`,
			expected: "True",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			strVal, ok := result.(*StringValue)
			if !ok {
				t.Fatalf("result is not *StringValue. got=%T (%+v)", result, result)
			}

			if strVal.Value != tt.expected {
				t.Errorf("BoolToStr() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinBoolToStr_ErrorCases tests error handling for BoolToStr().
func TestBuiltinBoolToStr_ErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "No arguments",
			input: `
begin
	BoolToStr();
end
			`,
		},
		{
			name: "Too many arguments",
			input: `
begin
	BoolToStr(True, False);
end
			`,
		},
		{
			name: "Non-boolean argument - integer",
			input: `
var x: Integer := 42;
begin
	BoolToStr(x);
end
			`,
		},
		{
			name: "Non-boolean argument - string",
			input: `
var s: String := "hello";
begin
	BoolToStr(s);
end
			`,
		},
		{
			name: "Non-boolean argument - float",
			input: `
var f: Float := 3.14;
begin
	BoolToStr(f);
end
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			// Should return an error
			if !isError(result) {
				t.Errorf("expected error for invalid BoolToStr() call, got %T: %+v", result, result)
			}
		})
	}
}

// TestBuiltinBoolToStr_PracticalUseCases tests real-world usage scenarios.
func TestBuiltinBoolToStr_PracticalUseCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Debug output with boolean",
			input: `
var enabled: Boolean := True;
begin
	"Feature enabled: " + BoolToStr(enabled);
end
			`,
			expected: "Feature enabled: True",
		},
		{
			name: "Validation result",
			input: `
var age: Integer := 25;
begin
	"Is adult: " + BoolToStr(age >= 18);
end
			`,
			expected: "Is adult: True",
		},
		{
			name: "Multiple conditions",
			input: `
var x: Integer := 10;
begin
	"x is positive: " + BoolToStr(x > 0) + ", x is even: " + BoolToStr((x mod 2) = 0);
end
			`,
			expected: "x is positive: True, x is even: True",
		},
		{
			name: "Replaces if-then-else workaround",
			input: `
var success: Boolean := True;
begin
	BoolToStr(success);
end
			`,
			expected: "True",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			strVal, ok := result.(*StringValue)
			if !ok {
				t.Fatalf("result is not *StringValue. got=%T (%+v)", result, result)
			}

			if strVal.Value != tt.expected {
				t.Errorf("BoolToStr() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}
