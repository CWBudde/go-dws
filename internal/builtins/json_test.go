package builtins

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// =============================================================================
// JSON Functions Tests
// =============================================================================

func TestParseJSON(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name      string
		args      []Value
		expectErr bool
	}{
		{
			name: "valid JSON string",
			args: []Value{
				&runtime.StringValue{Value: `{"key": "value"}`},
			},
			expectErr: false,
		},
		{
			name: "empty JSON object",
			args: []Value{
				&runtime.StringValue{Value: "{}"},
			},
			expectErr: false,
		},
		{
			name: "JSON array",
			args: []Value{
				&runtime.StringValue{Value: "[1, 2, 3]"},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseJSON(ctx, tt.args)

			if tt.expectErr {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %s", result.Type())
				}
			} else {
				// Mock returns nil, which is acceptable for this test
				// In real context, it would return a parsed JSON value
			}
		})
	}

	// Test error cases
	t.Run("no arguments", func(t *testing.T) {
		result := ParseJSON(ctx, []Value{})
		if result.Type() != "ERROR" {
			t.Errorf("ParseJSON with no args should return error, got %s", result.Type())
		}
	})

	t.Run("too many arguments", func(t *testing.T) {
		result := ParseJSON(ctx, []Value{
			&runtime.StringValue{Value: "{}"},
			&runtime.StringValue{Value: "extra"},
		})
		if result.Type() != "ERROR" {
			t.Errorf("ParseJSON with too many args should return error, got %s", result.Type())
		}
	})

	t.Run("non-string argument", func(t *testing.T) {
		result := ParseJSON(ctx, []Value{&runtime.IntegerValue{Value: 42}})
		if result.Type() != "ERROR" {
			t.Errorf("ParseJSON with non-string should return error, got %s", result.Type())
		}
	})
}

func TestToJSON(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name string
		args []Value
	}{
		{
			name: "integer value",
			args: []Value{&runtime.IntegerValue{Value: 42}},
		},
		{
			name: "string value",
			args: []Value{&runtime.StringValue{Value: "hello"}},
		},
		{
			name: "boolean value",
			args: []Value{&runtime.BooleanValue{Value: true}},
		},
		{
			name: "nil value",
			args: []Value{&runtime.NilValue{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToJSON(ctx, tt.args)
			// Mock returns "{}" which is acceptable
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value == "" {
				t.Error("ToJSON should not return empty string")
			}
		})
	}

	// Test error cases
	t.Run("no arguments", func(t *testing.T) {
		result := ToJSON(ctx, []Value{})
		if result.Type() != "ERROR" {
			t.Errorf("ToJSON with no args should return error, got %s", result.Type())
		}
	})

	t.Run("too many arguments", func(t *testing.T) {
		result := ToJSON(ctx, []Value{
			&runtime.IntegerValue{Value: 1},
			&runtime.IntegerValue{Value: 2},
		})
		if result.Type() != "ERROR" {
			t.Errorf("ToJSON with too many args should return error, got %s", result.Type())
		}
	})
}

func TestToJSONFormatted(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name string
		args []Value
	}{
		{
			name: "value with default indent",
			args: []Value{&runtime.IntegerValue{Value: 42}},
		},
		{
			name: "value with custom indent",
			args: []Value{
				&runtime.IntegerValue{Value: 42},
				&runtime.IntegerValue{Value: 4},
			},
		},
		{
			name: "value with zero indent",
			args: []Value{
				&runtime.IntegerValue{Value: 42},
				&runtime.IntegerValue{Value: 0},
			},
		},
		{
			name: "value with negative indent (becomes 0)",
			args: []Value{
				&runtime.IntegerValue{Value: 42},
				&runtime.IntegerValue{Value: -5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToJSONFormatted(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value == "" {
				t.Error("ToJSONFormatted should not return empty string")
			}
		})
	}

	// Test error cases
	t.Run("no arguments", func(t *testing.T) {
		result := ToJSONFormatted(ctx, []Value{})
		if result.Type() != "ERROR" {
			t.Errorf("ToJSONFormatted with no args should return error, got %s", result.Type())
		}
	})

	t.Run("too many arguments", func(t *testing.T) {
		result := ToJSONFormatted(ctx, []Value{
			&runtime.IntegerValue{Value: 1},
			&runtime.IntegerValue{Value: 2},
			&runtime.IntegerValue{Value: 3},
		})
		if result.Type() != "ERROR" {
			t.Errorf("ToJSONFormatted with too many args should return error, got %s", result.Type())
		}
	})

	t.Run("non-integer indent", func(t *testing.T) {
		result := ToJSONFormatted(ctx, []Value{
			&runtime.IntegerValue{Value: 42},
			&runtime.StringValue{Value: "not a number"},
		})
		if result.Type() != "ERROR" {
			t.Errorf("ToJSONFormatted with non-integer indent should return error, got %s", result.Type())
		}
	})
}

func TestJSONHasField(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected bool
	}{
		{
			name: "check field existence",
			args: []Value{
				&runtime.NilValue{}, // Mock object
				&runtime.StringValue{Value: "name"},
			},
			expected: false, // Mock returns false
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JSONHasField(ctx, tt.args)
			boolVal, ok := result.(*runtime.BooleanValue)
			if !ok {
				t.Fatalf("expected BooleanValue, got %T", result)
			}
			if boolVal.Value != tt.expected {
				t.Errorf("JSONHasField() = %v, want %v", boolVal.Value, tt.expected)
			}
		})
	}

	// Test error cases
	t.Run("no arguments", func(t *testing.T) {
		result := JSONHasField(ctx, []Value{})
		if result.Type() != "ERROR" {
			t.Errorf("JSONHasField with no args should return error, got %s", result.Type())
		}
	})

	t.Run("one argument", func(t *testing.T) {
		result := JSONHasField(ctx, []Value{&runtime.NilValue{}})
		if result.Type() != "ERROR" {
			t.Errorf("JSONHasField with one arg should return error, got %s", result.Type())
		}
	})

	t.Run("too many arguments", func(t *testing.T) {
		result := JSONHasField(ctx, []Value{
			&runtime.NilValue{},
			&runtime.StringValue{Value: "field"},
			&runtime.StringValue{Value: "extra"},
		})
		if result.Type() != "ERROR" {
			t.Errorf("JSONHasField with too many args should return error, got %s", result.Type())
		}
	})

	t.Run("non-string field name", func(t *testing.T) {
		result := JSONHasField(ctx, []Value{
			&runtime.NilValue{},
			&runtime.IntegerValue{Value: 42},
		})
		if result.Type() != "ERROR" {
			t.Errorf("JSONHasField with non-string field should return error, got %s", result.Type())
		}
	})
}

func TestJSONKeys(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name string
		args []Value
	}{
		{
			name: "get keys from object",
			args: []Value{&runtime.NilValue{}}, // Mock object
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JSONKeys(ctx, tt.args)
			// Mock returns "mock array" string
			if result == nil {
				t.Fatal("JSONKeys should not return nil")
			}
		})
	}

	// Test error cases
	t.Run("no arguments", func(t *testing.T) {
		result := JSONKeys(ctx, []Value{})
		if result.Type() != "ERROR" {
			t.Errorf("JSONKeys with no args should return error, got %s", result.Type())
		}
	})

	t.Run("too many arguments", func(t *testing.T) {
		result := JSONKeys(ctx, []Value{
			&runtime.NilValue{},
			&runtime.NilValue{},
		})
		if result.Type() != "ERROR" {
			t.Errorf("JSONKeys with too many args should return error, got %s", result.Type())
		}
	})
}

func TestJSONValues(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name string
		args []Value
	}{
		{
			name: "get values from object",
			args: []Value{&runtime.NilValue{}}, // Mock object
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JSONValues(ctx, tt.args)
			// Mock returns "mock array" string
			if result == nil {
				t.Fatal("JSONValues should not return nil")
			}
		})
	}

	// Test error cases
	t.Run("no arguments", func(t *testing.T) {
		result := JSONValues(ctx, []Value{})
		if result.Type() != "ERROR" {
			t.Errorf("JSONValues with no args should return error, got %s", result.Type())
		}
	})

	t.Run("too many arguments", func(t *testing.T) {
		result := JSONValues(ctx, []Value{
			&runtime.NilValue{},
			&runtime.NilValue{},
		})
		if result.Type() != "ERROR" {
			t.Errorf("JSONValues with too many args should return error, got %s", result.Type())
		}
	})
}

func TestJSONLength(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected int64
	}{
		{
			name:     "get length of object",
			args:     []Value{&runtime.NilValue{}}, // Mock object
			expected: 0,                            // Mock returns 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JSONLength(ctx, tt.args)
			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("JSONLength() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}

	// Test error cases
	t.Run("no arguments", func(t *testing.T) {
		result := JSONLength(ctx, []Value{})
		if result.Type() != "ERROR" {
			t.Errorf("JSONLength with no args should return error, got %s", result.Type())
		}
	})

	t.Run("too many arguments", func(t *testing.T) {
		result := JSONLength(ctx, []Value{
			&runtime.NilValue{},
			&runtime.NilValue{},
		})
		if result.Type() != "ERROR" {
			t.Errorf("JSONLength with too many args should return error, got %s", result.Type())
		}
	})
}
