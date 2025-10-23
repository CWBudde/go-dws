package interp

import (
	"strings"
	"testing"
)

// TestBuiltinAbs_Integers tests Abs() with integer values.
// Abs(x) returns the absolute value of x
// Task 8.185: Abs() function for integers
func TestBuiltinAbs_Integers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Negative integer",
			input: `
var vi: Integer := -2;
begin
	Abs(vi);
end
			`,
			expected: 2,
		},
		{
			name: "Positive integer",
			input: `
var vi: Integer := 2;
begin
	Abs(vi);
end
			`,
			expected: 2,
		},
		{
			name: "Zero",
			input: `
begin
	Abs(0);
end
			`,
			expected: 0,
		},
		{
			name: "Negative literal",
			input: `
begin
	Abs(-42);
end
			`,
			expected: 42,
		},
		{
			name: "Positive literal",
			input: `
begin
	Abs(42);
end
			`,
			expected: 42,
		},
		{
			name: "Expression with negative result",
			input: `
begin
	Abs(5 - 10);
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
				t.Errorf("Abs() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinAbs_Floats tests Abs() with float values.
// Task 8.185: Abs() function for floats
func TestBuiltinAbs_Floats(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name: "Negative float",
			input: `
var vf: Float := -1.5;
begin
	Abs(vf);
end
			`,
			expected: 1.5,
		},
		{
			name: "Positive float",
			input: `
var vf: Float := 1.5;
begin
	Abs(vf);
end
			`,
			expected: 1.5,
		},
		{
			name: "Zero float",
			input: `
begin
	Abs(0.0);
end
			`,
			expected: 0.0,
		},
		{
			name: "Negative float literal",
			input: `
begin
	Abs(-3.14159);
end
			`,
			expected: 3.14159,
		},
		{
			name: "Positive float literal",
			input: `
begin
	Abs(2.71828);
end
			`,
			expected: 2.71828,
		},
		{
			name: "Expression with negative result",
			input: `
begin
	Abs(1.5 - 3.0);
end
			`,
			expected: 1.5,
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
				t.Errorf("Abs() = %f, want %f", floatVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinAbs_Assignment tests Abs() in variable assignments.
// Task 8.185: Abs() function in assignments
func TestBuiltinAbs_Assignment(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedInt *int64
		expectedFlt *float64
	}{
		{
			name: "Assign integer result",
			input: `
var i: Integer := Abs(-5);
begin
	i;
end
			`,
			expectedInt: ptr(int64(5)),
		},
		{
			name: "Assign float result",
			input: `
var f: Float := Abs(-2.5);
begin
	f;
end
			`,
			expectedFlt: ptr(2.5),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)

			if tt.expectedInt != nil {
				intVal, ok := result.(*IntegerValue)
				if !ok {
					t.Fatalf("result is not *IntegerValue. got=%T (%+v)", result, result)
				}
				if intVal.Value != *tt.expectedInt {
					t.Errorf("Abs() = %d, want %d", intVal.Value, *tt.expectedInt)
				}
			}

			if tt.expectedFlt != nil {
				floatVal, ok := result.(*FloatValue)
				if !ok {
					t.Fatalf("result is not *FloatValue. got=%T (%+v)", result, result)
				}
				if floatVal.Value != *tt.expectedFlt {
					t.Errorf("Abs() = %f, want %f", floatVal.Value, *tt.expectedFlt)
				}
			}
		})
	}
}

// TestBuiltinAbs_Errors tests Abs() error cases.
// Task 8.185: Abs() error handling
func TestBuiltinAbs_Errors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "No arguments",
			input: `
begin
	Abs();
end
			`,
			expectedError: "Abs() expects exactly 1 argument",
		},
		{
			name: "Too many arguments",
			input: `
begin
	Abs(1, 2);
end
			`,
			expectedError: "Abs() expects exactly 1 argument",
		},
		{
			name: "String argument",
			input: `
begin
	Abs("hello");
end
			`,
			expectedError: "Abs() expects Integer or Float",
		},
		{
			name: "Boolean argument",
			input: `
begin
	Abs(true);
end
			`,
			expectedError: "Abs() expects Integer or Float",
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

// ptr is a helper function to create pointers for test expectations.
func ptr[T any](v T) *T {
	return &v
}

// TestBuiltinSqrt_BasicUsage tests Sqrt() with basic numeric values.
// Sqrt(x) returns the square root of x as a Float
// Task 8.185: Sqrt() function for math operations
func TestBuiltinSqrt_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name: "Perfect square integer",
			input: `
begin
	Sqrt(16);
end
			`,
			expected: 4.0,
		},
		{
			name: "Perfect square float",
			input: `
begin
	Sqrt(2.25);
end
			`,
			expected: 1.5,
		},
		{
			name: "Non-perfect square",
			input: `
begin
	Sqrt(2);
end
			`,
			expected: 1.4142135623730951,
		},
		{
			name: "Zero",
			input: `
begin
	Sqrt(0);
end
			`,
			expected: 0.0,
		},
		{
			name: "One",
			input: `
begin
	Sqrt(1);
end
			`,
			expected: 1.0,
		},
		{
			name: "Large number",
			input: `
begin
	Sqrt(256);
end
			`,
			expected: 16.0,
		},
		{
			name: "Small decimal",
			input: `
begin
	Sqrt(0.25);
end
			`,
			expected: 0.5,
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
				t.Errorf("Sqrt() = %f, want %f", floatVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinSqrt_WithVariables tests Sqrt() with variables.
// Task 8.185: Sqrt() function with variables
func TestBuiltinSqrt_WithVariables(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name: "Integer variable",
			input: `
var x: Integer := 9;
begin
	Sqrt(x);
end
			`,
			expected: 3.0,
		},
		{
			name: "Float variable",
			input: `
var f: Float := 56.25;
begin
	Sqrt(f);
end
			`,
			expected: 7.5,
		},
		{
			name: "Expression as argument",
			input: `
begin
	Sqrt(3 * 3);
end
			`,
			expected: 3.0,
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
				t.Errorf("Sqrt() = %f, want %f", floatVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinSqrt_Assignment tests Sqrt() in variable assignments.
// Task 8.185: Sqrt() function in assignments
func TestBuiltinSqrt_Assignment(t *testing.T) {
	input := `
var result: Float := Sqrt(16);
begin
	result;
end
	`
	result := testEval(input)

	floatVal, ok := result.(*FloatValue)
	if !ok {
		t.Fatalf("result is not *FloatValue. got=%T (%+v)", result, result)
	}

	if floatVal.Value != 4.0 {
		t.Errorf("Sqrt() = %f, want %f", floatVal.Value, 4.0)
	}
}

// TestBuiltinSqrt_Errors tests Sqrt() error cases.
// Task 8.185: Sqrt() error handling
func TestBuiltinSqrt_Errors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "No arguments",
			input: `
begin
	Sqrt();
end
			`,
			expectedError: "Sqrt() expects exactly 1 argument",
		},
		{
			name: "Too many arguments",
			input: `
begin
	Sqrt(4, 9);
end
			`,
			expectedError: "Sqrt() expects exactly 1 argument",
		},
		{
			name: "String argument",
			input: `
begin
	Sqrt("hello");
end
			`,
			expectedError: "Sqrt() expects Integer or Float",
		},
		{
			name: "Boolean argument",
			input: `
begin
	Sqrt(true);
end
			`,
			expectedError: "Sqrt() expects Integer or Float",
		},
		{
			name: "Negative number",
			input: `
begin
	Sqrt(-4);
end
			`,
			expectedError: "Sqrt() of negative number",
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

// TestBuiltinSin_BasicUsage tests Sin() with basic values.
// Sin(x) returns the sine of x (in radians) as a Float
// Task 8.185: Sin() function for trigonometric operations
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
// Task 8.185: Cos() function for trigonometric operations
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
// Task 8.185: Tan() function for trigonometric operations
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
// Task 8.185: Trigonometric function error handling
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
// Task 8.185: Exp() function for exponential operations
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
// Task 8.185: Ln() function for logarithmic operations
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
// Task 8.185: Verify Ln(Exp(x)) = x
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
// Task 8.185: Ln and Exp error handling
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
// Task 8.185: Round() function for rounding operations
func TestBuiltinRound_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Round(1.55) - from reference test",
			input: `
begin
	Round(1.55);
end
			`,
			expected: 2,
		},
		{
			name: "Round(-1.55) - from reference test",
			input: `
begin
	Round(-1.55);
end
			`,
			expected: -2,
		},
		{
			name: "Round(2.4) - rounds down",
			input: `
begin
	Round(2.4);
end
			`,
			expected: 2,
		},
		{
			name: "Round(2.5) - rounds up",
			input: `
begin
	Round(2.5);
end
			`,
			expected: 3,
		},
		{
			name: "Round(-2.5) - rounds down (away from zero)",
			input: `
begin
	Round(-2.5);
end
			`,
			expected: -3,
		},
		{
			name: "Round(0.0)",
			input: `
begin
	Round(0.0);
end
			`,
			expected: 0,
		},
		{
			name: "Round with Integer argument",
			input: `
var x: Integer := 5;
begin
	Round(x);
end
			`,
			expected: 5,
		},
		{
			name: "Round with Float variable",
			input: `
var f: Float := 3.7;
begin
	Round(f);
end
			`,
			expected: 4,
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
				t.Errorf("Round() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinTrunc_BasicUsage tests Trunc() with basic values.
// Trunc(x) truncates towards zero (removes decimal part) and returns Integer
// Task 8.185: Trunc() function for truncation operations
func TestBuiltinTrunc_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Trunc(1.55) - from reference test",
			input: `
begin
	Trunc(1.55);
end
			`,
			expected: 1,
		},
		{
			name: "Trunc(-1.55) - from reference test",
			input: `
begin
	Trunc(-1.55);
end
			`,
			expected: -1,
		},
		{
			name: "Trunc(2.9) - truncates towards zero",
			input: `
begin
	Trunc(2.9);
end
			`,
			expected: 2,
		},
		{
			name: "Trunc(-2.9) - truncates towards zero",
			input: `
begin
	Trunc(-2.9);
end
			`,
			expected: -2,
		},
		{
			name: "Trunc(0.0)",
			input: `
begin
	Trunc(0.0);
end
			`,
			expected: 0,
		},
		{
			name: "Trunc(5.1) - removes decimal",
			input: `
begin
	Trunc(5.1);
end
			`,
			expected: 5,
		},
		{
			name: "Trunc with Integer argument",
			input: `
var x: Integer := 7;
begin
	Trunc(x);
end
			`,
			expected: 7,
		},
		{
			name: "Trunc with Float variable",
			input: `
var f: Float := -3.8;
begin
	Trunc(f);
end
			`,
			expected: -3,
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
				t.Errorf("Trunc() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinRoundTrunc_Errors tests Round() and Trunc() error cases.
// Task 8.185: Round/Trunc error handling
func TestBuiltinRoundTrunc_Errors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "Round - no arguments",
			input: `
begin
	Round();
end
			`,
			expectedError: "Round() expects exactly 1 argument",
		},
		{
			name: "Round - too many arguments",
			input: `
begin
	Round(1.5, 2.5);
end
			`,
			expectedError: "Round() expects exactly 1 argument",
		},
		{
			name: "Round - string argument",
			input: `
begin
	Round("hello");
end
			`,
			expectedError: "Round() expects Integer or Float",
		},
		{
			name: "Trunc - no arguments",
			input: `
begin
	Trunc();
end
			`,
			expectedError: "Trunc() expects exactly 1 argument",
		},
		{
			name: "Trunc - too many arguments",
			input: `
begin
	Trunc(1.5, 2.5);
end
			`,
			expectedError: "Trunc() expects exactly 1 argument",
		},
		{
			name: "Trunc - boolean argument",
			input: `
begin
	Trunc(true);
end
			`,
			expectedError: "Trunc() expects Integer or Float",
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

// TestBuiltinRandom_BasicUsage tests Random() function.
// Random() returns a random Float between 0 and 1 (exclusive)
// Task 8.185: Random() function for random number generation
func TestBuiltinRandom_BasicUsage(t *testing.T) {
	// Test that Random returns values in range [0, 1)
	input := `
var i: Integer;
var allInRange := true;
for i := 1 to 100 do begin
	var r := Random();
	if (r < 0.0) or (r >= 1.0) then
		allInRange := false;
end;
begin
	allInRange;
end
	`
	result := testEval(input)

	boolVal, ok := result.(*BooleanValue)
	if !ok {
		t.Fatalf("result is not *BooleanValue. got=%T (%+v)", result, result)
	}

	if !boolVal.Value {
		t.Errorf("Random() produced value outside [0, 1) range")
	}
}

// TestBuiltinRandom_ReturnType tests that Random() returns Float.
// Task 8.185: Random() return type validation
func TestBuiltinRandom_ReturnType(t *testing.T) {
	input := `
begin
	Random();
end
	`
	result := testEval(input)

	_, ok := result.(*FloatValue)
	if !ok {
		t.Fatalf("Random() did not return *FloatValue. got=%T (%+v)", result, result)
	}
}

// TestBuiltinRandom_Variation tests that Random() produces different values.
// Task 8.185: Random() produces varied output
func TestBuiltinRandom_Variation(t *testing.T) {
	input := `
var r1 := Random();
var r2 := Random();
var r3 := Random();
begin
	(r1 <> r2) and (r2 <> r3);
end
	`
	result := testEval(input)

	boolVal, ok := result.(*BooleanValue)
	if !ok {
		t.Fatalf("result is not *BooleanValue. got=%T (%+v)", result, result)
	}

	if !boolVal.Value {
		t.Errorf("Random() produced identical consecutive values (very unlikely)")
	}
}

// TestBuiltinRandomize_BasicUsage tests Randomize() function.
// Randomize() seeds the random number generator
// Task 8.185: Randomize() function
func TestBuiltinRandomize_BasicUsage(t *testing.T) {
	// Test that Randomize can be called
	input := `
begin
	Randomize();
end
	`
	result := testEval(input)

	// Randomize returns nil/nothing
	_, ok := result.(*NilValue)
	if !ok {
		t.Fatalf("Randomize() did not return *NilValue. got=%T (%+v)", result, result)
	}
}

// TestBuiltinRandom_Errors tests Random() error cases.
// Task 8.185: Random() error handling
func TestBuiltinRandom_Errors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "Too many arguments",
			input: `
begin
	Random(5);
end
			`,
			expectedError: "Random() expects no arguments",
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

// TestBuiltinRandomize_Errors tests Randomize() error cases.
// Task 8.185: Randomize() error handling
func TestBuiltinRandomize_Errors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "Too many arguments",
			input: `
begin
	Randomize(5);
end
			`,
			expectedError: "Randomize() expects no arguments",
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
