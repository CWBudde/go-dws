package interp

import (
	"strings"
	"testing"
)

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
