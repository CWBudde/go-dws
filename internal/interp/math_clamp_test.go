package interp

import (
	"strings"
	"testing"
)

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
