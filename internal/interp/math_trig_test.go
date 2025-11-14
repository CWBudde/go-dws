package interp

import (
	"strings"
	"testing"
)

// TestBuiltinAbs_Integers tests Abs() with integer values.
// Abs(x) returns the absolute value of x
func TestBuiltinSin_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name: "Sin(0)",
			input: `
begin
	Sin(0);
end
			`,
			expected: 0.0,
		},
		{
			name: "Sin(10) - from reference test",
			input: `
begin
	Sin(10);
end
			`,
			expected: -0.5440211108893699,
		},
		{
			name: "Sin(π/2) ≈ Sin(1.5708)",
			input: `
begin
	Sin(1.5708);
end
			`,
			expected: 0.9999999999932537,
		},
		{
			name: "Sin with Integer argument",
			input: `
var x: Integer := 0;
begin
	Sin(x);
end
			`,
			expected: 0.0,
		},
		{
			name: "Sin with Float variable",
			input: `
var f: Float := 10.0;
begin
	Sin(f);
end
			`,
			expected: -0.5440211108893699,
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
				t.Errorf("Sin() = %.16f, want %.16f", floatVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinCos_BasicUsage tests Cos() with basic values.
// Cos(x) returns the cosine of x (in radians) as a Float
func TestBuiltinCos_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name: "Cos(0)",
			input: `
begin
	Cos(0);
end
			`,
			expected: 1.0,
		},
		{
			name: "Cos(10) - from reference test",
			input: `
begin
	Cos(10);
end
			`,
			expected: -0.8390715290764524,
		},
		{
			name: "Cos(π) ≈ Cos(3.14159)",
			input: `
begin
	Cos(3.14159);
end
			`,
			expected: -0.9999999999964793,
		},
		{
			name: "Cos with Integer argument",
			input: `
var x: Integer := 0;
begin
	Cos(x);
end
			`,
			expected: 1.0,
		},
		{
			name: "Cos with Float variable",
			input: `
var f: Float := 10.0;
begin
	Cos(f);
end
			`,
			expected: -0.8390715290764524,
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
				t.Errorf("Cos() = %.16f, want %.16f", floatVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinTan_BasicUsage tests Tan() with basic values.
// Tan(x) returns the tangent of x (in radians) as a Float
func TestBuiltinTan_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name: "Tan(0)",
			input: `
begin
	Tan(0);
end
			`,
			expected: 0.0,
		},
		{
			name: "Tan(10) - from reference test",
			input: `
begin
	Tan(10);
end
			`,
			expected: 0.6483608274590867,
		},
		{
			name: "Tan(π/4) ≈ Tan(0.7854)",
			input: `
begin
	Tan(0.7854);
end
			`,
			expected: 1.0000036732118496,
		},
		{
			name: "Tan with Integer argument",
			input: `
var x: Integer := 0;
begin
	Tan(x);
end
			`,
			expected: 0.0,
		},
		{
			name: "Tan with Float variable",
			input: `
var f: Float := 10.0;
begin
	Tan(f);
end
			`,
			expected: 0.6483608274590867,
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
				t.Errorf("Tan() = %.16f, want %.16f", floatVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinTrig_Errors tests trigonometric functions' error cases.
func TestBuiltinTrig_Errors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "Sin - no arguments",
			input: `
begin
	Sin();
end
			`,
			expectedError: "Sin() expects exactly 1 argument",
		},
		{
			name: "Sin - too many arguments",
			input: `
begin
	Sin(1, 2);
end
			`,
			expectedError: "Sin() expects exactly 1 argument",
		},
		{
			name: "Sin - string argument",
			input: `
begin
	Sin("hello");
end
			`,
			expectedError: "Sin() expects Integer or Float",
		},
		{
			name: "Cos - no arguments",
			input: `
begin
	Cos();
end
			`,
			expectedError: "Cos() expects exactly 1 argument",
		},
		{
			name: "Cos - string argument",
			input: `
begin
	Cos("hello");
end
			`,
			expectedError: "Cos() expects Integer or Float",
		},
		{
			name: "Tan - no arguments",
			input: `
begin
	Tan();
end
			`,
			expectedError: "Tan() expects exactly 1 argument",
		},
		{
			name: "Tan - boolean argument",
			input: `
begin
	Tan(true);
end
			`,
			expectedError: "Tan() expects Integer or Float",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			errVal, ok := result.(*ErrorValue)
			if !ok {
				t.Fatalf("expected error, got=%T (%+v)", result, result)
			}

			if !strings.Contains(errVal.Message, tt.expectedError) {
				t.Errorf("error message = %q, want to contain %q", errVal.Message, tt.expectedError)
			}
		})
	}
}

// TestBuiltinExp_BasicUsage tests Exp() with basic values.
// Exp(x) returns e^x as a Float
func TestBuiltinExp_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name: "Exp(0) should be 1",
			input: `
begin
	Exp(0);
end
			`,
			expected: 1.0,
		},
		{
			name: "Exp(1) should be e",
			input: `
begin
	Exp(1);
end
			`,
			expected: 2.7182818284590451,
		},
		{
			name: "Exp(1.5) - from reference test",
			input: `
begin
	Exp(1.5);
end
			`,
			expected: 4.4816890703380654,
		},
		{
			name: "Exp with Integer argument",
			input: `
var x: Integer := 0;
begin
	Exp(x);
end
			`,
			expected: 1.0,
		},
		{
			name: "Exp with Float variable",
			input: `
var f: Float := 1.5;
begin
	Exp(f);
end
			`,
			expected: 4.4816890703380654,
		},
		{
			name: "Exp(5)",
			input: `
begin
	Exp(5);
end
			`,
			expected: 148.4131591025766,
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
				t.Errorf("Exp() = %.16f, want %.16f", floatVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinLn_BasicUsage tests Ln() with basic values.
// Ln(x) returns the natural logarithm (log base e) of x as a Float
func TestBuiltinLn_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name: "Ln(1) should be 0",
			input: `
begin
	Ln(1);
end
			`,
			expected: 0.0,
		},
		{
			name: "Ln(e) should be 1",
			input: `
begin
	Ln(2.718281828459045);
end
			`,
			expected: 1.0,
		},
		{
			name: "Ln(1.5) - from reference test",
			input: `
begin
	Ln(1.5);
end
			`,
			expected: 0.4054651081081644,
		},
		{
			name: "Ln with Integer argument",
			input: `
var x: Integer := 1;
begin
	Ln(x);
end
			`,
			expected: 0.0,
		},
		{
			name: "Ln with Float variable",
			input: `
var f: Float := 1.5;
begin
	Ln(f);
end
			`,
			expected: 0.4054651081081644,
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
				t.Errorf("Ln() = %.16f, want %.16f", floatVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinLnExp_Inverse tests that Ln and Exp are inverse functions.
func TestBuiltinLnExp_Inverse(t *testing.T) {
	input := `
begin
	Ln(Exp(5));
end
	`
	result := testEval(input)

	floatVal, ok := result.(*FloatValue)
	if !ok {
		t.Fatalf("result is not *FloatValue. got=%T (%+v)", result, result)
	}

	expected := 5.0
	if floatVal.Value != expected {
		t.Errorf("Ln(Exp(5)) = %.16f, want %.16f", floatVal.Value, expected)
	}
}

// TestBuiltinLnExp_Errors tests Ln and Exp error cases.
func TestBuiltinLnExp_Errors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "Exp - no arguments",
			input: `
begin
	Exp();
end
			`,
			expectedError: "Exp() expects exactly 1 argument",
		},
		{
			name: "Exp - too many arguments",
			input: `
begin
	Exp(1, 2);
end
			`,
			expectedError: "Exp() expects exactly 1 argument",
		},
		{
			name: "Exp - string argument",
			input: `
begin
	Exp("hello");
end
			`,
			expectedError: "Exp() expects Integer or Float",
		},
		{
			name: "Ln - no arguments",
			input: `
begin
	Ln();
end
			`,
			expectedError: "Ln() expects exactly 1 argument",
		},
		{
			name: "Ln - too many arguments",
			input: `
begin
	Ln(1, 2);
end
			`,
			expectedError: "Ln() expects exactly 1 argument",
		},
		{
			name: "Ln - string argument",
			input: `
begin
	Ln("hello");
end
			`,
			expectedError: "Ln() expects Integer or Float",
		},
		{
			name: "Ln - negative number",
			input: `
begin
	Ln(-1);
end
			`,
			expectedError: "Ln() of non-positive number",
		},
		{
			name: "Ln - zero",
			input: `
begin
	Ln(0);
end
			`,
			expectedError: "Ln() of non-positive number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			errVal, ok := result.(*ErrorValue)
			if !ok {
				t.Fatalf("expected error, got=%T (%+v)", result, result)
			}

			if !strings.Contains(errVal.Message, tt.expectedError) {
				t.Errorf("error message = %q, want to contain %q", errVal.Message, tt.expectedError)
			}
		})
	}
}

// TestBuiltinRound_BasicUsage tests Round() with basic values.
// Round(x) rounds to the nearest integer and returns Integer
