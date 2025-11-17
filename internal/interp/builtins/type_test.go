package builtins

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// =============================================================================
// Type Introspection Functions Tests
// =============================================================================

func TestTypeOf(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected string
	}{
		{
			name:     "integer type",
			args:     []Value{&runtime.IntegerValue{Value: 42}},
			expected: "INTEGER",
		},
		{
			name:     "float type",
			args:     []Value{&runtime.FloatValue{Value: 3.14}},
			expected: "FLOAT",
		},
		{
			name:     "string type",
			args:     []Value{&runtime.StringValue{Value: "hello"}},
			expected: "STRING",
		},
		{
			name:     "boolean type",
			args:     []Value{&runtime.BooleanValue{Value: true}},
			expected: "BOOLEAN",
		},
		{
			name:     "nil type",
			args:     []Value{&runtime.NilValue{}},
			expected: "NIL",
		},
		{
			name: "array type",
			args: []Value{&runtime.ArrayValue{
				Elements: []Value{
					&runtime.IntegerValue{Value: 1},
					&runtime.IntegerValue{Value: 2},
				},
			}},
			expected: "ARRAY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TypeOf(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("TypeOf() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestTypeOfClass(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected string
	}{
		{
			name:     "non-object value returns empty string",
			args:     []Value{&runtime.IntegerValue{Value: 42}},
			expected: "",
		},
		{
			name:     "string value returns empty string",
			args:     []Value{&runtime.StringValue{Value: "hello"}},
			expected: "",
		},
		{
			name:     "nil value returns empty string",
			args:     []Value{&runtime.NilValue{}},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TypeOfClass(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("TypeOfClass() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// Test error cases
func TestTypeFunctionsErrors(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		funcName string
		fn       func(Context, []Value) Value
		args     []Value
	}{
		{
			name:     "TypeOf with no args",
			funcName: "TypeOf",
			fn:       TypeOf,
			args:     []Value{},
		},
		{
			name:     "TypeOf with too many args",
			funcName: "TypeOf",
			fn:       TypeOf,
			args:     []Value{&runtime.IntegerValue{Value: 1}, &runtime.IntegerValue{Value: 2}},
		},
		{
			name:     "TypeOfClass with no args",
			funcName: "TypeOfClass",
			fn:       TypeOfClass,
			args:     []Value{},
		},
		{
			name:     "TypeOfClass with too many args",
			funcName: "TypeOfClass",
			fn:       TypeOfClass,
			args:     []Value{&runtime.IntegerValue{Value: 1}, &runtime.IntegerValue{Value: 2}},
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
