package interp

import (
	"strings"
	"testing"
)

// TestBuiltinAbs_Integers tests Abs() with integer values.
// Abs(x) returns the absolute value of x
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
func TestBuiltinAbs_Assignment(t *testing.T) {
	tests := []struct {
		expectedInt *int64
		expectedFlt *float64
		name        string
		input       string
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

// TestBuiltinMin_Integers tests Min() with integer values.
// Min(a, b) returns the smaller of two values
func TestBuiltinMin_Integers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Min(5, 10) returns 5",
			input: `
begin
	Min(5, 10);
end
			`,
			expected: 5,
		},
		{
			name: "Min(10, 5) returns 5",
			input: `
begin
	Min(10, 5);
end
			`,
			expected: 5,
		},
		{
			name: "Min with negative numbers",
			input: `
begin
	Min(-5, -10);
end
			`,
			expected: -10,
		},
		{
			name: "Min with negative and positive",
			input: `
begin
	Min(-5, 10);
end
			`,
			expected: -5,
		},
		{
			name: "Min with equal values",
			input: `
begin
	Min(7, 7);
end
			`,
			expected: 7,
		},
		{
			name: "Min with zero",
			input: `
begin
	Min(0, 5);
end
			`,
			expected: 0,
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
				t.Errorf("Min() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinMin_Floats tests Min() with float values.
func TestBuiltinMin_Floats(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name: "Min(3.14, 2.71)",
			input: `
begin
	Min(3.14, 2.71);
end
			`,
			expected: 2.71,
		},
		{
			name: "Min(2.71, 3.14)",
			input: `
begin
	Min(2.71, 3.14);
end
			`,
			expected: 2.71,
		},
		{
			name: "Min with negative floats",
			input: `
begin
	Min(-1.5, -2.5);
end
			`,
			expected: -2.5,
		},
		{
			name: "Min with equal floats",
			input: `
begin
	Min(1.5, 1.5);
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
				t.Errorf("Min() = %f, want %f", floatVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinMin_MixedTypes tests Min() with mixed Integer/Float types.
func TestBuiltinMin_MixedTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name: "Min(5, 3.14) promotes to Float",
			input: `
begin
	Min(5, 3.14);
end
			`,
			expected: 3.14,
		},
		{
			name: "Min(3.14, 5) promotes to Float",
			input: `
begin
	Min(3.14, 5);
end
			`,
			expected: 3.14,
		},
		{
			name: "Min(-5, -3.14)",
			input: `
begin
	Min(-5, -3.14);
end
			`,
			expected: -5.0,
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
				t.Errorf("Min() = %f, want %f", floatVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinMax_Integers tests Max() with integer values.
// Max(a, b) returns the larger of two values
func TestBuiltinMax_Integers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Max(5, 10) returns 10",
			input: `
begin
	Max(5, 10);
end
			`,
			expected: 10,
		},
		{
			name: "Max(10, 5) returns 10",
			input: `
begin
	Max(10, 5);
end
			`,
			expected: 10,
		},
		{
			name: "Max with negative numbers",
			input: `
begin
	Max(-5, -10);
end
			`,
			expected: -5,
		},
		{
			name: "Max with negative and positive",
			input: `
begin
	Max(-5, 10);
end
			`,
			expected: 10,
		},
		{
			name: "Max with equal values",
			input: `
begin
	Max(7, 7);
end
			`,
			expected: 7,
		},
		{
			name: "Max with zero",
			input: `
begin
	Max(0, -5);
end
			`,
			expected: 0,
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
				t.Errorf("Max() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinMax_Floats tests Max() with float values.
func TestBuiltinMax_Floats(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name: "Max(3.14, 2.71)",
			input: `
begin
	Max(3.14, 2.71);
end
			`,
			expected: 3.14,
		},
		{
			name: "Max(2.71, 3.14)",
			input: `
begin
	Max(2.71, 3.14);
end
			`,
			expected: 3.14,
		},
		{
			name: "Max with negative floats",
			input: `
begin
	Max(-1.5, -2.5);
end
			`,
			expected: -1.5,
		},
		{
			name: "Max with equal floats",
			input: `
begin
	Max(1.5, 1.5);
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
				t.Errorf("Max() = %f, want %f", floatVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinMax_MixedTypes tests Max() with mixed Integer/Float types.
func TestBuiltinMax_MixedTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name: "Max(5, 3.14) promotes to Float",
			input: `
begin
	Max(5, 3.14);
end
			`,
			expected: 5.0,
		},
		{
			name: "Max(3.14, 5) promotes to Float",
			input: `
begin
	Max(3.14, 5);
end
			`,
			expected: 5.0,
		},
		{
			name: "Max(-5, -3.14)",
			input: `
begin
	Max(-5, -3.14);
end
			`,
			expected: -3.14,
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
				t.Errorf("Max() = %f, want %f", floatVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinMinMax_Errors tests Min() and Max() error cases.
func TestBuiltinMinMax_Errors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "Min - no arguments",
			input: `
begin
	Min();
end
			`,
			expectedError: "Min() expects exactly 2 arguments",
		},
		{
			name: "Min - one argument",
			input: `
begin
	Min(5);
end
			`,
			expectedError: "Min() expects exactly 2 arguments",
		},
		{
			name: "Min - too many arguments",
			input: `
begin
	Min(1, 2, 3);
end
			`,
			expectedError: "Min() expects exactly 2 arguments",
		},
		{
			name: "Min - string argument",
			input: `
begin
	Min("hello", 5);
end
			`,
			expectedError: "Min() expects Integer or Float",
		},
		{
			name: "Min - boolean argument",
			input: `
begin
	Min(true, 5);
end
			`,
			expectedError: "Min() expects Integer or Float",
		},
		{
			name: "Max - no arguments",
			input: `
begin
	Max();
end
			`,
			expectedError: "Max() expects exactly 2 arguments",
		},
		{
			name: "Max - one argument",
			input: `
begin
	Max(5);
end
			`,
			expectedError: "Max() expects exactly 2 arguments",
		},
		{
			name: "Max - too many arguments",
			input: `
begin
	Max(1, 2, 3);
end
			`,
			expectedError: "Max() expects exactly 2 arguments",
		},
		{
			name: "Max - string argument",
			input: `
begin
	Max(5, "hello");
end
			`,
			expectedError: "Max() expects Integer or Float",
		},
		{
			name: "Max - boolean argument",
			input: `
begin
	Max(false, 5);
end
			`,
			expectedError: "Max() expects Integer or Float",
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

// TestBuiltinSqr_Integers tests Sqr() with integer values.
// Sqr(x) returns x * x
func TestBuiltinSqr_Integers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Sqr(5) returns 25",
			input: `
begin
	Sqr(5);
end
			`,
			expected: 25,
		},
		{
			name: "Sqr(0) returns 0",
			input: `
begin
	Sqr(0);
end
			`,
			expected: 0,
		},
		{
			name: "Sqr(1) returns 1",
			input: `
begin
	Sqr(1);
end
			`,
			expected: 1,
		},
		{
			name: "Sqr(-3) returns 9",
			input: `
begin
	Sqr(-3);
end
			`,
			expected: 9,
		},
		{
			name: "Sqr(10) returns 100",
			input: `
begin
	Sqr(10);
end
			`,
			expected: 100,
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
				t.Errorf("Sqr() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinSqr_Floats tests Sqr() with float values.
func TestBuiltinSqr_Floats(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name: "Sqr(3.0) returns 9.0",
			input: `
begin
	Sqr(3.0);
end
			`,
			expected: 9.0,
		},
		{
			name: "Sqr(2.5) returns 6.25",
			input: `
begin
	Sqr(2.5);
end
			`,
			expected: 6.25,
		},
		{
			name: "Sqr(-1.5) returns 2.25",
			input: `
begin
	Sqr(-1.5);
end
			`,
			expected: 2.25,
		},
		{
			name: "Sqr(0.5) returns 0.25",
			input: `
begin
	Sqr(0.5);
end
			`,
			expected: 0.25,
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
				t.Errorf("Sqr() = %f, want %f", floatVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinPower_BasicUsage tests Power() with basic values.
// Power(x, y) returns x^y as Float
func TestBuiltinPower_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name: "Power(2, 8) returns 256.0",
			input: `
begin
	Power(2, 8);
end
			`,
			expected: 256.0,
		},
		{
			name: "Power(2.0, 0.5) returns sqrt(2)",
			input: `
begin
	Power(2.0, 0.5);
end
			`,
			expected: 1.4142135623730951,
		},
		{
			name: "Power(2, -1) returns 0.5",
			input: `
begin
	Power(2, -1);
end
			`,
			expected: 0.5,
		},
		{
			name: "Power(10, 3) returns 1000.0",
			input: `
begin
	Power(10, 3);
end
			`,
			expected: 1000.0,
		},
		{
			name: "Power(5, 0) returns 1.0",
			input: `
begin
	Power(5, 0);
end
			`,
			expected: 1.0,
		},
		{
			name: "Power(3.0, 2.0) returns 9.0",
			input: `
begin
	Power(3.0, 2.0);
end
			`,
			expected: 9.0,
		},
		{
			name: "Power(4, 0.5) returns 2.0",
			input: `
begin
	Power(4, 0.5);
end
			`,
			expected: 2.0,
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
				t.Errorf("Power() = %.16f, want %.16f", floatVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinPower_SpecialCases tests Power() special cases.
func TestBuiltinPower_SpecialCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name: "Power(0, 0) returns 1.0 (by convention)",
			input: `
begin
	Power(0, 0);
end
			`,
			expected: 1.0,
		},
		{
			name: "Power(0, 5) returns 0.0",
			input: `
begin
	Power(0, 5);
end
			`,
			expected: 0.0,
		},
		{
			name: "Power(1, 100) returns 1.0",
			input: `
begin
	Power(1, 100);
end
			`,
			expected: 1.0,
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
				t.Errorf("Power() = %.16f, want %.16f", floatVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinSqrPower_Errors tests Sqr() and Power() error cases.
func TestBuiltinSqrPower_Errors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "Sqr - no arguments",
			input: `
begin
	Sqr();
end
			`,
			expectedError: "Sqr() expects exactly 1 argument",
		},
		{
			name: "Sqr - too many arguments",
			input: `
begin
	Sqr(5, 10);
end
			`,
			expectedError: "Sqr() expects exactly 1 argument",
		},
		{
			name: "Sqr - string argument",
			input: `
begin
	Sqr("hello");
end
			`,
			expectedError: "Sqr() expects Integer or Float",
		},
		{
			name: "Power - no arguments",
			input: `
begin
	Power();
end
			`,
			expectedError: "Power() expects exactly 2 arguments",
		},
		{
			name: "Power - one argument",
			input: `
begin
	Power(5);
end
			`,
			expectedError: "Power() expects exactly 2 arguments",
		},
		{
			name: "Power - too many arguments",
			input: `
begin
	Power(2, 3, 4);
end
			`,
			expectedError: "Power() expects exactly 2 arguments",
		},
		{
			name: "Power - string argument",
			input: `
begin
	Power("hello", 5);
end
			`,
			expectedError: "Power() expects Integer or Float",
		},
		{
			name: "Power - boolean argument",
			input: `
begin
	Power(2, true);
end
			`,
			expectedError: "Power() expects Integer or Float",
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

// TestBuiltinCeil_BasicUsage tests Ceil() with basic values.
// Ceil(x) rounds up to the nearest integer
func TestBuiltinCeil_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Ceil(3.2) returns 4",
			input: `
begin
	Ceil(3.2);
end
			`,
			expected: 4,
		},
		{
			name: "Ceil(3.8) returns 4",
			input: `
begin
	Ceil(3.8);
end
			`,
			expected: 4,
		},
		{
			name: "Ceil(-3.2) returns -3",
			input: `
begin
	Ceil(-3.2);
end
			`,
			expected: -3,
		},
		{
			name: "Ceil(-3.8) returns -3",
			input: `
begin
	Ceil(-3.8);
end
			`,
			expected: -3,
		},
		{
			name: "Ceil(5.0) returns 5",
			input: `
begin
	Ceil(5.0);
end
			`,
			expected: 5,
		},
		{
			name: "Ceil(0.0) returns 0",
			input: `
begin
	Ceil(0.0);
end
			`,
			expected: 0,
		},
		{
			name: "Ceil(0.1) returns 1",
			input: `
begin
	Ceil(0.1);
end
			`,
			expected: 1,
		},
		{
			name: "Ceil(-0.1) returns 0",
			input: `
begin
	Ceil(-0.1);
end
			`,
			expected: 0,
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
				t.Errorf("Ceil() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinFloor_BasicUsage tests Floor() with basic values.
// Floor(x) rounds down to the nearest integer
func TestBuiltinFloor_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Floor(3.8) returns 3",
			input: `
begin
	Floor(3.8);
end
			`,
			expected: 3,
		},
		{
			name: "Floor(3.2) returns 3",
			input: `
begin
	Floor(3.2);
end
			`,
			expected: 3,
		},
		{
			name: "Floor(-3.8) returns -4",
			input: `
begin
	Floor(-3.8);
end
			`,
			expected: -4,
		},
		{
			name: "Floor(-3.2) returns -4",
			input: `
begin
	Floor(-3.2);
end
			`,
			expected: -4,
		},
		{
			name: "Floor(5.0) returns 5",
			input: `
begin
	Floor(5.0);
end
			`,
			expected: 5,
		},
		{
			name: "Floor(0.0) returns 0",
			input: `
begin
	Floor(0.0);
end
			`,
			expected: 0,
		},
		{
			name: "Floor(0.9) returns 0",
			input: `
begin
	Floor(0.9);
end
			`,
			expected: 0,
		},
		{
			name: "Floor(-0.1) returns -1",
			input: `
begin
	Floor(-0.1);
end
			`,
			expected: -1,
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
				t.Errorf("Floor() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinCeilFloor_WithIntegers tests Ceil() and Floor() with Integer arguments.
func TestBuiltinCeilFloor_WithIntegers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Ceil with Integer argument",
			input: `
var x: Integer := 5;
begin
	Ceil(x);
end
			`,
			expected: 5,
		},
		{
			name: "Floor with Integer argument",
			input: `
var x: Integer := 5;
begin
	Floor(x);
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
				t.Errorf("result = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinCeilFloor_Errors tests Ceil() and Floor() error cases.
func TestBuiltinCeilFloor_Errors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "Ceil - no arguments",
			input: `
begin
	Ceil();
end
			`,
			expectedError: "Ceil() expects exactly 1 argument",
		},
		{
			name: "Ceil - too many arguments",
			input: `
begin
	Ceil(3.5, 4.5);
end
			`,
			expectedError: "Ceil() expects exactly 1 argument",
		},
		{
			name: "Ceil - string argument",
			input: `
begin
	Ceil("hello");
end
			`,
			expectedError: "Ceil() expects Integer or Float",
		},
		{
			name: "Ceil - boolean argument",
			input: `
begin
	Ceil(true);
end
			`,
			expectedError: "Ceil() expects Integer or Float",
		},
		{
			name: "Floor - no arguments",
			input: `
begin
	Floor();
end
			`,
			expectedError: "Floor() expects exactly 1 argument",
		},
		{
			name: "Floor - too many arguments",
			input: `
begin
	Floor(3.5, 4.5);
end
			`,
			expectedError: "Floor() expects exactly 1 argument",
		},
		{
			name: "Floor - string argument",
			input: `
begin
	Floor("hello");
end
			`,
			expectedError: "Floor() expects Integer or Float",
		},
		{
			name: "Floor - boolean argument",
			input: `
begin
	Floor(false);
end
			`,
			expectedError: "Floor() expects Integer or Float",
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

// TestBuiltinRandomInt_BasicUsage tests RandomInt() returns values in range [0, max).
// RandomInt(max) returns a random Integer in range [0, max)
func TestBuiltinRandomInt_BasicUsage(t *testing.T) {
	input := `
var i: Integer;
var allInRange := true;
for i := 1 to 100 do begin
	var r := RandomInt(10);
	if (r < 0) or (r >= 10) then
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
		t.Errorf("RandomInt(10) produced value outside [0, 10) range")
	}
}

// TestBuiltinRandomInt_ReturnType tests that RandomInt() returns Integer.
func TestBuiltinRandomInt_ReturnType(t *testing.T) {
	input := `
begin
	RandomInt(10);
end
	`
	result := testEval(input)

	_, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("RandomInt() did not return *IntegerValue. got=%T (%+v)", result, result)
	}
}

// TestBuiltinRandomInt_Variation tests that RandomInt() produces different values.
func TestBuiltinRandomInt_Variation(t *testing.T) {
	input := `
var r1 := RandomInt(1000);
var r2 := RandomInt(1000);
var r3 := RandomInt(1000);
begin
	(r1 <> r2) or (r2 <> r3);
end
	`
	result := testEval(input)

	boolVal, ok := result.(*BooleanValue)
	if !ok {
		t.Fatalf("result is not *BooleanValue. got=%T (%+v)", result, result)
	}

	// With max=1000, it's extremely unlikely all three are the same
	if !boolVal.Value {
		t.Errorf("RandomInt() produced identical consecutive values (very unlikely)")
	}
}

// TestBuiltinRandomInt_MaxOne tests RandomInt(1) always returns 0.
func TestBuiltinRandomInt_MaxOne(t *testing.T) {
	input := `
var i: Integer;
var allZero := true;
for i := 1 to 10 do begin
	var r := RandomInt(1);
	if r <> 0 then
		allZero := false;
end;
begin
	allZero;
end
	`
	result := testEval(input)

	boolVal, ok := result.(*BooleanValue)
	if !ok {
		t.Fatalf("result is not *BooleanValue. got=%T (%+v)", result, result)
	}

	if !boolVal.Value {
		t.Errorf("RandomInt(1) should always return 0")
	}
}

// TestBuiltinRandomInt_Errors tests RandomInt() error cases.
func TestBuiltinRandomInt_Errors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "No arguments",
			input: `
begin
	RandomInt();
end
			`,
			expectedError: "RandomInt() expects exactly 1 argument",
		},
		{
			name: "Too many arguments",
			input: `
begin
	RandomInt(10, 20);
end
			`,
			expectedError: "RandomInt() expects exactly 1 argument",
		},
		{
			name: "Max is zero",
			input: `
begin
	RandomInt(0);
end
			`,
			expectedError: "RandomInt() expects max > 0",
		},
		{
			name: "Max is negative",
			input: `
begin
	RandomInt(-5);
end
			`,
			expectedError: "RandomInt() expects max > 0",
		},
		{
			name: "String argument",
			input: `
begin
	RandomInt("hello");
end
			`,
			expectedError: "RandomInt() expects Integer",
		},
		{
			name: "Float argument",
			input: `
begin
	RandomInt(10.5);
end
			`,
			expectedError: "RandomInt() expects Integer",
		},
		{
			name: "Boolean argument",
			input: `
begin
	RandomInt(true);
end
			`,
			expectedError: "RandomInt() expects Integer",
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

// TestBuiltinUnsigned32_BasicUsage tests Unsigned32() with typical values.
// Unsigned32(x) converts signed Integer to unsigned 32-bit representation
func TestBuiltinUnsigned32_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Zero",
			input: `
begin
	Unsigned32(0);
end
			`,
			expected: 0,
		},
		{
			name: "Positive value within range",
			input: `
begin
	Unsigned32(255);
end
			`,
			expected: 255,
		},
		{
			name: "Max positive int32",
			input: `
begin
	Unsigned32(2147483647);
end
			`,
			expected: 2147483647,
		},
		{
			name: "Small positive value",
			input: `
begin
	Unsigned32(1);
end
			`,
			expected: 1,
		},
		{
			name: "Large positive value",
			input: `
begin
	Unsigned32(1000000);
end
			`,
			expected: 1000000,
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
				t.Errorf("Unsigned32() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinUnsigned32_NegativeValues tests Unsigned32() wrapping behavior with negative values.
func TestBuiltinUnsigned32_NegativeValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Negative one wraps to max uint32",
			input: `
begin
	Unsigned32(-1);
end
			`,
			expected: 4294967295, // 0xFFFFFFFF
		},
		{
			name: "Min negative int32 wraps correctly",
			input: `
begin
	Unsigned32(-2147483648);
end
			`,
			expected: 2147483648, // 0x80000000
		},
		{
			name: "Small negative value",
			input: `
begin
	Unsigned32(-100);
end
			`,
			expected: 4294967196, // 0xFFFFFF9C
		},
		{
			name: "Negative value in expression",
			input: `
begin
	Unsigned32(-5);
end
			`,
			expected: 4294967291, // 0xFFFFFFFB
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
				t.Errorf("Unsigned32() = %d (0x%X), want %d (0x%X)",
					intVal.Value, intVal.Value, tt.expected, tt.expected)
			}
		})
	}
}

// TestBuiltinUnsigned32_OverflowBehavior tests Unsigned32() with values larger than uint32.
func TestBuiltinUnsigned32_OverflowBehavior(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Max uint32 value",
			input: `
var i: Integer := $FFFFFFFF;
begin
	Unsigned32(i);
end
			`,
			expected: 4294967295, // 0xFFFFFFFF
		},
		{
			name: "Max uint32 + 1 wraps to 0",
			input: `
var i: Integer := $FFFFFFFF;
begin
	Unsigned32(i + 1);
end
			`,
			expected: 0,
		},
		{
			name: "Value with upper bits set truncates",
			input: `
var i: Integer := $100000001;
begin
	Unsigned32(i);
end
			`,
			expected: 1, // Only lower 32 bits: 0x00000001
		},
		{
			name: "Large value truncates correctly",
			input: `
var i: Integer := $1FFFFFFFF;
begin
	Unsigned32(i);
end
			`,
			expected: 4294967295, // Lower 32 bits: 0xFFFFFFFF
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
				t.Errorf("Unsigned32() = %d (0x%X), want %d (0x%X)",
					intVal.Value, intVal.Value, tt.expected, tt.expected)
			}
		})
	}
}

// TestBuiltinUnsigned32_WithVariables tests Unsigned32() with variables and expressions.
func TestBuiltinUnsigned32_WithVariables(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Variable holding positive value",
			input: `
var i: Integer := 42;
begin
	Unsigned32(i);
end
			`,
			expected: 42,
		},
		{
			name: "Variable holding negative value",
			input: `
var i: Integer := -1;
begin
	Unsigned32(i);
end
			`,
			expected: 4294967295,
		},
		{
			name: "Expression result",
			input: `
begin
	Unsigned32(10 - 15);
end
			`,
			expected: 4294967291, // Unsigned32(-5)
		},
		{
			name: "Multiple conversions in sequence",
			input: `
var i: Integer := 1;
var u1 := Unsigned32(i);
i := -1;
var u2 := Unsigned32(i);
begin
	u1 + u2;
end
			`,
			expected: 4294967296, // 1 + 4294967295
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
				t.Errorf("result = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinUnsigned32_Errors tests Unsigned32() error cases.
func TestBuiltinUnsigned32_Errors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "No arguments",
			input: `
begin
	Unsigned32();
end
			`,
			expectedError: "Unsigned32() expects exactly 1 argument",
		},
		{
			name: "Too many arguments",
			input: `
begin
	Unsigned32(10, 20);
end
			`,
			expectedError: "Unsigned32() expects exactly 1 argument",
		},
		{
			name: "String argument",
			input: `
begin
	Unsigned32("hello");
end
			`,
			expectedError: "Unsigned32() expects Integer",
		},
		{
			name: "Float argument",
			input: `
begin
	Unsigned32(10.5);
end
			`,
			expectedError: "Unsigned32() expects Integer",
		},
		{
			name: "Boolean argument",
			input: `
begin
	Unsigned32(true);
end
			`,
			expectedError: "Unsigned32() expects Integer",
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

// TestBuiltinClampInt_BasicUsage tests ClampInt() with various value ranges.
// ClampInt(value, min, max) clamps value to [min, max]
func TestBuiltinClampInt_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Value within range",
			input: `
begin
	ClampInt(5, 1, 10);
end
			`,
			expected: 5,
		},
		{
			name: "Value below min",
			input: `
begin
	ClampInt(0, 1, 10);
end
			`,
			expected: 1,
		},
		{
			name: "Value above max",
			input: `
begin
	ClampInt(15, 1, 10);
end
			`,
			expected: 10,
		},
		{
			name: "Value equals min",
			input: `
begin
	ClampInt(1, 1, 10);
end
			`,
			expected: 1,
		},
		{
			name: "Value equals max",
			input: `
begin
	ClampInt(10, 1, 10);
end
			`,
			expected: 10,
		},
		{
			name: "Negative range - value in range",
			input: `
begin
	ClampInt(-5, -10, -1);
end
			`,
			expected: -5,
		},
		{
			name: "Negative range - value below min",
			input: `
begin
	ClampInt(-15, -10, -1);
end
			`,
			expected: -10,
		},
		{
			name: "Negative range - value above max",
			input: `
begin
	ClampInt(0, -10, -1);
end
			`,
			expected: -1,
		},
		{
			name: "Zero in range",
			input: `
begin
	ClampInt(0, -5, 5);
end
			`,
			expected: 0,
		},
		{
			name: "Single value range",
			input: `
begin
	ClampInt(10, 5, 5);
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
				t.Errorf("ClampInt() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinClampInt_WithVariables tests ClampInt() with variables and expressions.
func TestBuiltinClampInt_WithVariables(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name: "Variable arguments",
			input: `
var value: Integer := 15;
var min: Integer := 1;
var max: Integer := 10;
begin
	ClampInt(value, min, max);
end
			`,
			expected: 10,
		},
		{
			name: "Expression as value",
			input: `
begin
	ClampInt(5 + 10, 1, 10);
end
			`,
			expected: 10,
		},
		{
			name: "From death_star.pas example",
			input: `
var intensity: Integer := 4;
var cShades: String := '.:-=+*#%@';
begin
	ClampInt(intensity + 1, 1, Length(cShades));
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
				t.Errorf("ClampInt() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinClampInt_Errors tests ClampInt() error cases.
func TestBuiltinClampInt_Errors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "No arguments",
			input: `
begin
	ClampInt();
end
			`,
			expectedError: "ClampInt() expects exactly 3 arguments",
		},
		{
			name: "Too few arguments",
			input: `
begin
	ClampInt(5, 1);
end
			`,
			expectedError: "ClampInt() expects exactly 3 arguments",
		},
		{
			name: "Too many arguments",
			input: `
begin
	ClampInt(5, 1, 10, 15);
end
			`,
			expectedError: "ClampInt() expects exactly 3 arguments",
		},
		{
			name: "String value argument",
			input: `
begin
	ClampInt("hello", 1, 10);
end
			`,
			expectedError: "ClampInt() expects Integer",
		},
		{
			name: "Float value argument",
			input: `
begin
	ClampInt(5.5, 1, 10);
end
			`,
			expectedError: "ClampInt() expects Integer",
		},
		{
			name: "Float min argument",
			input: `
begin
	ClampInt(5, 1.5, 10);
end
			`,
			expectedError: "ClampInt() expects Integer",
		},
		{
			name: "Float max argument",
			input: `
begin
	ClampInt(5, 1, 10.5);
end
			`,
			expectedError: "ClampInt() expects Integer",
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

// TestBuiltinClamp_BasicUsage tests Clamp() with various float value ranges.
// Clamp(value, min, max) clamps value to [min, max] and returns Float
func TestBuiltinClamp_BasicUsage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name: "Float value within range",
			input: `
begin
	Clamp(5.5, 1.0, 10.0);
end
			`,
			expected: 5.5,
		},
		{
			name: "Float value below min",
			input: `
begin
	Clamp(0.5, 1.0, 10.0);
end
			`,
			expected: 1.0,
		},
		{
			name: "Float value above max",
			input: `
begin
	Clamp(15.5, 1.0, 10.0);
end
			`,
			expected: 10.0,
		},
		{
			name: "Float value equals min",
			input: `
begin
	Clamp(1.0, 1.0, 10.0);
end
			`,
			expected: 1.0,
		},
		{
			name: "Float value equals max",
			input: `
begin
	Clamp(10.0, 1.0, 10.0);
end
			`,
			expected: 10.0,
		},
		{
			name: "Negative float range",
			input: `
begin
	Clamp(-5.5, -10.0, -1.0);
end
			`,
			expected: -5.5,
		},
		{
			name: "Zero in range",
			input: `
begin
	Clamp(0.0, -5.0, 5.0);
end
			`,
			expected: 0.0,
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
				t.Errorf("Clamp() = %f, want %f", floatVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinClamp_MixedTypes tests Clamp() with mixed Integer/Float arguments.
func TestBuiltinClamp_MixedTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name: "Integer value, float bounds",
			input: `
begin
	Clamp(5, 1.0, 10.0);
end
			`,
			expected: 5.0,
		},
		{
			name: "Float value, integer bounds",
			input: `
begin
	Clamp(5.5, 1, 10);
end
			`,
			expected: 5.5,
		},
		{
			name: "All integers (promoted to Float)",
			input: `
begin
	Clamp(5, 1, 10);
end
			`,
			expected: 5.0,
		},
		{
			name: "Integer value clamped by integer min",
			input: `
begin
	Clamp(0, 1, 10);
end
			`,
			expected: 1.0,
		},
		{
			name: "Mixed types - value exceeds max",
			input: `
begin
	Clamp(15, 1.0, 10);
end
			`,
			expected: 10.0,
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
				t.Errorf("Clamp() = %f, want %f", floatVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinClamp_WithVariables tests Clamp() with variables and expressions.
func TestBuiltinClamp_WithVariables(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name: "Variable arguments",
			input: `
var value: Float := 15.5;
var min: Float := 1.0;
var max: Float := 10.0;
begin
	Clamp(value, min, max);
end
			`,
			expected: 10.0,
		},
		{
			name: "Expression as value",
			input: `
begin
	Clamp(5.5 + 10.0, 1.0, 10.0);
end
			`,
			expected: 10.0,
		},
		{
			name: "Integer variables",
			input: `
var value: Integer := 15;
var min: Integer := 1;
var max: Integer := 10;
begin
	Clamp(value, min, max);
end
			`,
			expected: 10.0,
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
				t.Errorf("Clamp() = %f, want %f", floatVal.Value, tt.expected)
			}
		})
	}
}

// TestBuiltinClamp_Errors tests Clamp() error cases.
func TestBuiltinClamp_Errors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "No arguments",
			input: `
begin
	Clamp();
end
			`,
			expectedError: "Clamp() expects exactly 3 arguments",
		},
		{
			name: "Too few arguments",
			input: `
begin
	Clamp(5.0, 1.0);
end
			`,
			expectedError: "Clamp() expects exactly 3 arguments",
		},
		{
			name: "Too many arguments",
			input: `
begin
	Clamp(5.0, 1.0, 10.0, 15.0);
end
			`,
			expectedError: "Clamp() expects exactly 3 arguments",
		},
		{
			name: "String value argument",
			input: `
begin
	Clamp("hello", 1.0, 10.0);
end
			`,
			expectedError: "Clamp() expects Integer or Float",
		},
		{
			name: "Boolean value argument",
			input: `
begin
	Clamp(true, 1.0, 10.0);
end
			`,
			expectedError: "Clamp() expects Integer or Float",
		},
		{
			name: "String min argument",
			input: `
begin
	Clamp(5.0, "hello", 10.0);
end
			`,
			expectedError: "Clamp() expects Integer or Float",
		},
		{
			name: "Boolean max argument",
			input: `
begin
	Clamp(5.0, 1.0, false);
end
			`,
			expectedError: "Clamp() expects Integer or Float",
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
