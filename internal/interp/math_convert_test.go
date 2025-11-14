package interp

import (
	"strings"
	"testing"
)

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

// TestBuiltinCeil_BasicUsage tests Ceil() function.
// Ceil(x) returns the smallest integer greater than or equal to x
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
