package interp

import (
	"strings"
	"testing"
)

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
