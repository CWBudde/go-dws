package builtins

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// =============================================================================
// Variant Functions Tests
// =============================================================================

func TestVarType(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected int64
	}{
		{
			name:     "integer value",
			args:     []Value{&runtime.IntegerValue{Value: 42}},
			expected: 3, // varInteger
		},
		{
			name:     "float value",
			args:     []Value{&runtime.FloatValue{Value: 3.14}},
			expected: 5, // varDouble
		},
		{
			name:     "string value",
			args:     []Value{&runtime.StringValue{Value: "hello"}},
			expected: 256, // varString
		},
		{
			name:     "boolean value",
			args:     []Value{&runtime.BooleanValue{Value: true}},
			expected: 11, // varBoolean
		},
		{
			name:     "nil value",
			args:     []Value{&runtime.NilValue{}},
			expected: 0, // varEmpty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VarType(ctx, tt.args)
			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("VarType() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

func TestVarIsNull(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected bool
	}{
		{
			name:     "nil value is null",
			args:     []Value{&runtime.NilValue{}},
			expected: true,
		},
		{
			name:     "integer is not null",
			args:     []Value{&runtime.IntegerValue{Value: 42}},
			expected: false,
		},
		{
			name:     "string is not null",
			args:     []Value{&runtime.StringValue{Value: "hello"}},
			expected: false,
		},
		{
			name:     "zero is not null",
			args:     []Value{&runtime.IntegerValue{Value: 0}},
			expected: false,
		},
		{
			name:     "empty string is not null",
			args:     []Value{&runtime.StringValue{Value: ""}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VarIsNull(ctx, tt.args)
			boolVal, ok := result.(*runtime.BooleanValue)
			if !ok {
				t.Fatalf("expected BooleanValue, got %T", result)
			}
			if boolVal.Value != tt.expected {
				t.Errorf("VarIsNull() = %v, want %v", boolVal.Value, tt.expected)
			}
		})
	}
}

func TestVarIsEmpty(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected bool
	}{
		{
			name:     "nil value is empty",
			args:     []Value{&runtime.NilValue{}},
			expected: true,
		},
		{
			name:     "integer is not empty",
			args:     []Value{&runtime.IntegerValue{Value: 42}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VarIsEmpty(ctx, tt.args)
			boolVal, ok := result.(*runtime.BooleanValue)
			if !ok {
				t.Fatalf("expected BooleanValue, got %T", result)
			}
			if boolVal.Value != tt.expected {
				t.Errorf("VarIsEmpty() = %v, want %v", boolVal.Value, tt.expected)
			}
		})
	}
}

func TestVarIsClear(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected bool
	}{
		{
			name:     "nil value is clear",
			args:     []Value{&runtime.NilValue{}},
			expected: true,
		},
		{
			name:     "integer is not clear",
			args:     []Value{&runtime.IntegerValue{Value: 42}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VarIsClear(ctx, tt.args)
			boolVal, ok := result.(*runtime.BooleanValue)
			if !ok {
				t.Fatalf("expected BooleanValue, got %T", result)
			}
			if boolVal.Value != tt.expected {
				t.Errorf("VarIsClear() = %v, want %v", boolVal.Value, tt.expected)
			}
		})
	}
}

func TestVarIsArray(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected bool
	}{
		{
			name: "array value is array",
			args: []Value{&runtime.ArrayValue{
				Elements: []Value{
					&runtime.IntegerValue{Value: 1},
					&runtime.IntegerValue{Value: 2},
				},
			}},
			expected: true,
		},
		{
			name:     "integer is not array",
			args:     []Value{&runtime.IntegerValue{Value: 42}},
			expected: false,
		},
		{
			name:     "string is not array",
			args:     []Value{&runtime.StringValue{Value: "hello"}},
			expected: false,
		},
		{
			name:     "nil is not array",
			args:     []Value{&runtime.NilValue{}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VarIsArray(ctx, tt.args)
			boolVal, ok := result.(*runtime.BooleanValue)
			if !ok {
				t.Fatalf("expected BooleanValue, got %T", result)
			}
			if boolVal.Value != tt.expected {
				t.Errorf("VarIsArray() = %v, want %v", boolVal.Value, tt.expected)
			}
		})
	}
}

func TestVarIsStr(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected bool
	}{
		{
			name:     "string value is string",
			args:     []Value{&runtime.StringValue{Value: "hello"}},
			expected: true,
		},
		{
			name:     "empty string is string",
			args:     []Value{&runtime.StringValue{Value: ""}},
			expected: true,
		},
		{
			name:     "integer is not string",
			args:     []Value{&runtime.IntegerValue{Value: 42}},
			expected: false,
		},
		{
			name:     "nil is not string",
			args:     []Value{&runtime.NilValue{}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VarIsStr(ctx, tt.args)
			boolVal, ok := result.(*runtime.BooleanValue)
			if !ok {
				t.Fatalf("expected BooleanValue, got %T", result)
			}
			if boolVal.Value != tt.expected {
				t.Errorf("VarIsStr() = %v, want %v", boolVal.Value, tt.expected)
			}
		})
	}
}

func TestVarIsNumeric(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected bool
	}{
		{
			name:     "integer is numeric",
			args:     []Value{&runtime.IntegerValue{Value: 42}},
			expected: true,
		},
		{
			name:     "float is numeric",
			args:     []Value{&runtime.FloatValue{Value: 3.14}},
			expected: true,
		},
		{
			name:     "zero is numeric",
			args:     []Value{&runtime.IntegerValue{Value: 0}},
			expected: true,
		},
		{
			name:     "string is not numeric",
			args:     []Value{&runtime.StringValue{Value: "42"}},
			expected: false,
		},
		{
			name:     "boolean is not numeric",
			args:     []Value{&runtime.BooleanValue{Value: true}},
			expected: false,
		},
		{
			name:     "nil is not numeric",
			args:     []Value{&runtime.NilValue{}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VarIsNumeric(ctx, tt.args)
			boolVal, ok := result.(*runtime.BooleanValue)
			if !ok {
				t.Fatalf("expected BooleanValue, got %T", result)
			}
			if boolVal.Value != tt.expected {
				t.Errorf("VarIsNumeric() = %v, want %v", boolVal.Value, tt.expected)
			}
		})
	}
}

func TestVarToStr(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		expected string
		args     []Value
	}{
		{
			name:     "integer to string",
			args:     []Value{&runtime.IntegerValue{Value: 42}},
			expected: "42",
		},
		{
			name:     "float to string",
			args:     []Value{&runtime.FloatValue{Value: 3.14}},
			expected: "3.14",
		},
		{
			name:     "string to string",
			args:     []Value{&runtime.StringValue{Value: "hello"}},
			expected: "hello",
		},
		{
			name:     "boolean true to string",
			args:     []Value{&runtime.BooleanValue{Value: true}},
			expected: "True",
		},
		{
			name:     "boolean false to string",
			args:     []Value{&runtime.BooleanValue{Value: false}},
			expected: "False",
		},
		{
			name:     "nil to empty string",
			args:     []Value{&runtime.NilValue{}},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VarToStr(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("VarToStr() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestVarToInt(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected int64
		isError  bool
	}{
		{
			name:     "integer to int",
			args:     []Value{&runtime.IntegerValue{Value: 42}},
			expected: 42,
		},
		{
			name:     "float to int (truncate)",
			args:     []Value{&runtime.FloatValue{Value: 3.9}},
			expected: 3,
		},
		{
			name:     "string number to int",
			args:     []Value{&runtime.StringValue{Value: "123"}},
			expected: 123,
		},
		{
			name:     "boolean true to int",
			args:     []Value{&runtime.BooleanValue{Value: true}},
			expected: 1,
		},
		{
			name:     "boolean false to int",
			args:     []Value{&runtime.BooleanValue{Value: false}},
			expected: 0,
		},
		{
			name:     "nil to zero",
			args:     []Value{&runtime.NilValue{}},
			expected: 0,
		},
		{
			name:    "invalid string to int",
			args:    []Value{&runtime.StringValue{Value: "abc"}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VarToInt(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("VarToInt() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

func TestVarToFloat(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected float64
		isError  bool
	}{
		{
			name:     "float to float",
			args:     []Value{&runtime.FloatValue{Value: 3.14}},
			expected: 3.14,
		},
		{
			name:     "integer to float",
			args:     []Value{&runtime.IntegerValue{Value: 42}},
			expected: 42.0,
		},
		{
			name:     "string number to float",
			args:     []Value{&runtime.StringValue{Value: "3.14"}},
			expected: 3.14,
		},
		{
			name:     "boolean true to float",
			args:     []Value{&runtime.BooleanValue{Value: true}},
			expected: 1.0,
		},
		{
			name:     "boolean false to float",
			args:     []Value{&runtime.BooleanValue{Value: false}},
			expected: 0.0,
		},
		{
			name:     "nil to zero",
			args:     []Value{&runtime.NilValue{}},
			expected: 0.0,
		},
		{
			name:    "invalid string to float",
			args:    []Value{&runtime.StringValue{Value: "abc"}},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VarToFloat(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			floatVal, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}
			if floatVal.Value != tt.expected {
				t.Errorf("VarToFloat() = %f, want %f", floatVal.Value, tt.expected)
			}
		})
	}
}

func TestVarAsType(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		value       Value
		expectedVal interface{}
		name        string
		targetType  int64
		isError     bool
	}{
		{
			name:        "string to integer",
			value:       &runtime.StringValue{Value: "42"},
			targetType:  3, // varInteger
			expectedVal: int64(42),
		},
		{
			name:        "integer to float",
			value:       &runtime.IntegerValue{Value: 42},
			targetType:  5, // varDouble
			expectedVal: 42.0,
		},
		{
			name:        "integer to string",
			value:       &runtime.IntegerValue{Value: 42},
			targetType:  256, // varString
			expectedVal: "42",
		},
		{
			name:        "integer to boolean (non-zero)",
			value:       &runtime.IntegerValue{Value: 42},
			targetType:  11, // varBoolean
			expectedVal: true,
		},
		{
			name:        "integer to boolean (zero)",
			value:       &runtime.IntegerValue{Value: 0},
			targetType:  11, // varBoolean
			expectedVal: false,
		},
		{
			name:        "string to boolean (non-empty)",
			value:       &runtime.StringValue{Value: "hello"},
			targetType:  11, // varBoolean
			expectedVal: true,
		},
		{
			name:        "nil to integer",
			value:       &runtime.NilValue{},
			targetType:  3, // varInteger
			expectedVal: int64(0),
		},
		{
			name:        "nil to string",
			value:       &runtime.NilValue{},
			targetType:  256, // varString
			expectedVal: "",
		},
		{
			name:       "unsupported type code",
			value:      &runtime.IntegerValue{Value: 42},
			targetType: 999,
			isError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []Value{
				tt.value,
				&runtime.IntegerValue{Value: tt.targetType},
			}
			result := VarAsType(ctx, args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			switch expected := tt.expectedVal.(type) {
			case int64:
				intVal, ok := result.(*runtime.IntegerValue)
				if !ok {
					t.Fatalf("expected IntegerValue, got %T", result)
				}
				if intVal.Value != expected {
					t.Errorf("VarAsType() = %d, want %d", intVal.Value, expected)
				}
			case float64:
				floatVal, ok := result.(*runtime.FloatValue)
				if !ok {
					t.Fatalf("expected FloatValue, got %T", result)
				}
				if floatVal.Value != expected {
					t.Errorf("VarAsType() = %f, want %f", floatVal.Value, expected)
				}
			case string:
				strVal, ok := result.(*runtime.StringValue)
				if !ok {
					t.Fatalf("expected StringValue, got %T", result)
				}
				if strVal.Value != expected {
					t.Errorf("VarAsType() = %q, want %q", strVal.Value, expected)
				}
			case bool:
				boolVal, ok := result.(*runtime.BooleanValue)
				if !ok {
					t.Fatalf("expected BooleanValue, got %T", result)
				}
				if boolVal.Value != expected {
					t.Errorf("VarAsType() = %v, want %v", boolVal.Value, expected)
				}
			}
		})
	}
}

func TestVarClear(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name string
		args []Value
	}{
		{
			name: "clear integer",
			args: []Value{&runtime.IntegerValue{Value: 42}},
		},
		{
			name: "clear string",
			args: []Value{&runtime.StringValue{Value: "hello"}},
		},
		{
			name: "clear nil",
			args: []Value{&runtime.NilValue{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VarClear(ctx, tt.args)
			// VarClear should return a NilValue
			if result.Type() != "NIL" {
				t.Errorf("VarClear() should return NIL, got %s", result.Type())
			}
		})
	}
}

// Test error cases for all variant functions
func TestVariantFunctionsErrors(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		funcName string
		fn       func(Context, []Value) Value
		args     []Value
	}{
		{
			name:     "VarType with no args",
			funcName: "VarType",
			fn:       VarType,
			args:     []Value{},
		},
		{
			name:     "VarType with too many args",
			funcName: "VarType",
			fn:       VarType,
			args:     []Value{&runtime.IntegerValue{Value: 1}, &runtime.IntegerValue{Value: 2}},
		},
		{
			name:     "VarIsNull with no args",
			funcName: "VarIsNull",
			fn:       VarIsNull,
			args:     []Value{},
		},
		{
			name:     "VarToInt with no args",
			funcName: "VarToInt",
			fn:       VarToInt,
			args:     []Value{},
		},
		{
			name:     "VarToFloat with no args",
			funcName: "VarToFloat",
			fn:       VarToFloat,
			args:     []Value{},
		},
		{
			name:     "VarAsType with no args",
			funcName: "VarAsType",
			fn:       VarAsType,
			args:     []Value{},
		},
		{
			name:     "VarAsType with one arg",
			funcName: "VarAsType",
			fn:       VarAsType,
			args:     []Value{&runtime.IntegerValue{Value: 42}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(ctx, tt.args)
			if result.Type() != "ERROR" {
				t.Errorf("%s should return error for invalid args, got %s", tt.funcName, result.Type())
			}
		})
	}
}
