package builtins

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// =============================================================================
// System Functions Tests
// =============================================================================

func TestGetStackTrace(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name string
		args []Value
	}{
		{
			name: "no arguments",
			args: []Value{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStackTrace(ctx, tt.args)
			// Should return a string (even if empty from mock context)
			if _, ok := result.(*runtime.StringValue); !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
		})
	}

	// Test error case
	t.Run("too many arguments", func(t *testing.T) {
		result := GetStackTrace(ctx, []Value{&runtime.IntegerValue{Value: 1}})
		if result.Type() != "ERROR" {
			t.Errorf("GetStackTrace with args should return error, got %s", result.Type())
		}
	})
}

func TestGetCallStack(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name string
		args []Value
	}{
		{
			name: "no arguments",
			args: []Value{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetCallStack(ctx, tt.args)
			// Should return some value (array or empty from mock context)
			if result == nil {
				t.Fatal("GetCallStack should not return nil")
			}
		})
	}

	// Test error case
	t.Run("too many arguments", func(t *testing.T) {
		result := GetCallStack(ctx, []Value{&runtime.IntegerValue{Value: 1}})
		if result.Type() != "ERROR" {
			t.Errorf("GetCallStack with args should return error, got %s", result.Type())
		}
	})
}

func TestAssigned(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected bool
	}{
		{
			name:     "nil is not assigned",
			args:     []Value{&runtime.NilValue{}},
			expected: false,
		},
		{
			name:     "integer is assigned",
			args:     []Value{&runtime.IntegerValue{Value: 42}},
			expected: true,
		},
		{
			name:     "string is assigned",
			args:     []Value{&runtime.StringValue{Value: "hello"}},
			expected: true,
		},
		{
			name:     "zero is assigned",
			args:     []Value{&runtime.IntegerValue{Value: 0}},
			expected: true,
		},
		{
			name:     "empty string is assigned",
			args:     []Value{&runtime.StringValue{Value: ""}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Assigned(ctx, tt.args)
			boolVal, ok := result.(*runtime.BooleanValue)
			if !ok {
				t.Fatalf("expected BooleanValue, got %T", result)
			}
			if boolVal.Value != tt.expected {
				t.Errorf("Assigned() = %v, want %v", boolVal.Value, tt.expected)
			}
		})
	}
}

func TestAssert(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name        string
		args        []Value
		expectNil   bool
		expectError bool
	}{
		{
			name: "true condition passes",
			args: []Value{
				&runtime.BooleanValue{Value: true},
			},
			expectNil: true,
		},
		{
			name: "true condition with message passes",
			args: []Value{
				&runtime.BooleanValue{Value: true},
				&runtime.StringValue{Value: "custom message"},
			},
			expectNil: true,
		},
		{
			name: "false condition fails",
			args: []Value{
				&runtime.BooleanValue{Value: false},
			},
			expectError: false, // Will call RaiseAssertionFailed on context
		},
		{
			name: "non-boolean first arg",
			args: []Value{
				&runtime.IntegerValue{Value: 1},
			},
			expectError: true,
		},
		{
			name: "non-string second arg with false condition",
			args: []Value{
				&runtime.BooleanValue{Value: false},
				&runtime.IntegerValue{Value: 42},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Assert(ctx, tt.args)

			if tt.expectError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %s", result.Type())
				}
				return
			}

			if tt.expectNil {
				if result.Type() != "NIL" {
					t.Errorf("expected NIL, got %s", result.Type())
				}
			}
		})
	}
}

func TestInteger(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected int64
		isError  bool
	}{
		{
			name:     "integer to integer",
			args:     []Value{&runtime.IntegerValue{Value: 42}},
			expected: 42,
		},
		{
			name:     "float to integer",
			args:     []Value{&runtime.FloatValue{Value: 3.7}},
			expected: 3,
		},
		{
			name:     "boolean true to integer",
			args:     []Value{&runtime.BooleanValue{Value: true}},
			expected: 1,
		},
		{
			name:     "boolean false to integer",
			args:     []Value{&runtime.BooleanValue{Value: false}},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Integer(ctx, tt.args)

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
				t.Errorf("Integer() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

func TestStrToIntDef(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected int64
		isError  bool
	}{
		{
			name: "valid decimal string",
			args: []Value{
				&runtime.StringValue{Value: "42"},
				&runtime.IntegerValue{Value: -1},
			},
			expected: 42,
		},
		{
			name: "invalid string returns default",
			args: []Value{
				&runtime.StringValue{Value: "abc"},
				&runtime.IntegerValue{Value: -1},
			},
			expected: -1,
		},
		{
			name: "empty string returns default",
			args: []Value{
				&runtime.StringValue{Value: ""},
				&runtime.IntegerValue{Value: 999},
			},
			expected: 999,
		},
		{
			name: "binary string with base 2",
			args: []Value{
				&runtime.StringValue{Value: "1010"},
				&runtime.IntegerValue{Value: -1},
				&runtime.IntegerValue{Value: 2},
			},
			expected: 10,
		},
		{
			name: "hex string with base 16",
			args: []Value{
				&runtime.StringValue{Value: "FF"},
				&runtime.IntegerValue{Value: -1},
				&runtime.IntegerValue{Value: 16},
			},
			expected: 255,
		},
		{
			name: "invalid base returns error",
			args: []Value{
				&runtime.StringValue{Value: "42"},
				&runtime.IntegerValue{Value: -1},
				&runtime.IntegerValue{Value: 1},
			},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrToIntDef(ctx, tt.args)

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
				t.Errorf("StrToIntDef() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

func TestStrToFloatDef(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected float64
	}{
		{
			name: "valid float string",
			args: []Value{
				&runtime.StringValue{Value: "3.14"},
				&runtime.FloatValue{Value: -1.0},
			},
			expected: 3.14,
		},
		{
			name: "invalid string returns default",
			args: []Value{
				&runtime.StringValue{Value: "abc"},
				&runtime.FloatValue{Value: -1.0},
			},
			expected: -1.0,
		},
		{
			name: "empty string returns default",
			args: []Value{
				&runtime.StringValue{Value: ""},
				&runtime.FloatValue{Value: 999.0},
			},
			expected: 999.0,
		},
		{
			name: "integer string",
			args: []Value{
				&runtime.StringValue{Value: "42"},
				&runtime.FloatValue{Value: -1.0},
			},
			expected: 42.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrToFloatDef(ctx, tt.args)
			floatVal, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", result)
			}
			if floatVal.Value != tt.expected {
				t.Errorf("StrToFloatDef() = %f, want %f", floatVal.Value, tt.expected)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected string
		isError  bool
	}{
		{
			name: "simple string format",
			args: []Value{
				&runtime.StringValue{Value: "Hello %s"},
				&runtime.ArrayValue{
					Elements: []Value{
						&runtime.StringValue{Value: "World"},
					},
				},
			},
			expected: "Hello World",
		},
		{
			name: "integer format",
			args: []Value{
				&runtime.StringValue{Value: "Number: %d"},
				&runtime.ArrayValue{
					Elements: []Value{
						&runtime.IntegerValue{Value: 42},
					},
				},
			},
			expected: "Number: 42",
		},
		{
			name: "multiple values",
			args: []Value{
				&runtime.StringValue{Value: "%s: %d"},
				&runtime.ArrayValue{
					Elements: []Value{
						&runtime.StringValue{Value: "Count"},
						&runtime.IntegerValue{Value: 10},
					},
				},
			},
			expected: "Count: 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Format(ctx, tt.args)

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("Format() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

// Test error cases for all system functions
func TestSystemFunctionsErrors(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		funcName string
		fn       func(Context, []Value) Value
		args     []Value
	}{
		{
			name:     "Assigned with no args",
			funcName: "Assigned",
			fn:       Assigned,
			args:     []Value{},
		},
		{
			name:     "Assert with no args",
			funcName: "Assert",
			fn:       Assert,
			args:     []Value{},
		},
		{
			name:     "Assert with too many args",
			funcName: "Assert",
			fn:       Assert,
			args:     []Value{&runtime.BooleanValue{Value: true}, &runtime.StringValue{Value: "a"}, &runtime.IntegerValue{Value: 1}},
		},
		{
			name:     "Integer with no args",
			funcName: "Integer",
			fn:       Integer,
			args:     []Value{},
		},
		{
			name:     "StrToIntDef with one arg",
			funcName: "StrToIntDef",
			fn:       StrToIntDef,
			args:     []Value{&runtime.StringValue{Value: "42"}},
		},
		{
			name:     "StrToFloatDef with no args",
			funcName: "StrToFloatDef",
			fn:       StrToFloatDef,
			args:     []Value{},
		},
		{
			name:     "StrToFloatDef with one arg",
			funcName: "StrToFloatDef",
			fn:       StrToFloatDef,
			args:     []Value{&runtime.StringValue{Value: "3.14"}},
		},
		{
			name:     "Format with no args",
			funcName: "Format",
			fn:       Format,
			args:     []Value{},
		},
		{
			name:     "Format with one arg",
			funcName: "Format",
			fn:       Format,
			args:     []Value{&runtime.StringValue{Value: "test"}},
		},
		{
			name:     "Format with non-string first arg",
			funcName: "Format",
			fn:       Format,
			args:     []Value{&runtime.IntegerValue{Value: 42}, &runtime.ArrayValue{Elements: []Value{}}},
		},
		{
			name:     "Format with non-array second arg",
			funcName: "Format",
			fn:       Format,
			args:     []Value{&runtime.StringValue{Value: "test"}, &runtime.IntegerValue{Value: 42}},
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
