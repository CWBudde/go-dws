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
