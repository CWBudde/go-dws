package builtins

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// =============================================================================
// String Comparison Functions Tests
// =============================================================================

func TestSameText(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected bool
	}{
		{
			name: "same text different case",
			args: []Value{
				&runtime.StringValue{Value: "Hello"},
				&runtime.StringValue{Value: "hello"},
			},
			expected: true,
		},
		{
			name: "same text same case",
			args: []Value{
				&runtime.StringValue{Value: "Hello"},
				&runtime.StringValue{Value: "Hello"},
			},
			expected: true,
		},
		{
			name: "different text",
			args: []Value{
				&runtime.StringValue{Value: "Hello"},
				&runtime.StringValue{Value: "World"},
			},
			expected: false,
		},
		{
			name: "empty strings",
			args: []Value{
				&runtime.StringValue{Value: ""},
				&runtime.StringValue{Value: ""},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SameText(ctx, tt.args)
			boolVal, ok := result.(*runtime.BooleanValue)
			if !ok {
				t.Fatalf("expected BooleanValue, got %T", result)
			}
			if boolVal.Value != tt.expected {
				t.Errorf("SameText() = %v, want %v", boolVal.Value, tt.expected)
			}
		})
	}
}

func TestCompareText(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected int64
	}{
		{
			name: "equal strings",
			args: []Value{
				&runtime.StringValue{Value: "Hello"},
				&runtime.StringValue{Value: "hello"},
			},
			expected: 0,
		},
		{
			name: "first less than second",
			args: []Value{
				&runtime.StringValue{Value: "apple"},
				&runtime.StringValue{Value: "banana"},
			},
			expected: -1,
		},
		{
			name: "first greater than second",
			args: []Value{
				&runtime.StringValue{Value: "zebra"},
				&runtime.StringValue{Value: "apple"},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareText(ctx, tt.args)
			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("CompareText() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

func TestCompareStr(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected int64
	}{
		{
			name: "equal strings",
			args: []Value{
				&runtime.StringValue{Value: "Hello"},
				&runtime.StringValue{Value: "Hello"},
			},
			expected: 0,
		},
		{
			name: "different case - not equal",
			args: []Value{
				&runtime.StringValue{Value: "Hello"},
				&runtime.StringValue{Value: "hello"},
			},
			expected: -1,
		},
		{
			name: "first less than second",
			args: []Value{
				&runtime.StringValue{Value: "apple"},
				&runtime.StringValue{Value: "banana"},
			},
			expected: -1,
		},
		{
			name: "first greater than second",
			args: []Value{
				&runtime.StringValue{Value: "zebra"},
				&runtime.StringValue{Value: "apple"},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareStr(ctx, tt.args)
			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("CompareStr() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

func TestAnsiCompareText(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected int64
	}{
		{
			name: "equal strings",
			args: []Value{
				&runtime.StringValue{Value: "Hello"},
				&runtime.StringValue{Value: "hello"},
			},
			expected: 0,
		},
		{
			name: "first less than second",
			args: []Value{
				&runtime.StringValue{Value: "apple"},
				&runtime.StringValue{Value: "banana"},
			},
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AnsiCompareText(ctx, tt.args)
			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("AnsiCompareText() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

func TestAnsiCompareStr(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected int64
	}{
		{
			name: "equal strings",
			args: []Value{
				&runtime.StringValue{Value: "Hello"},
				&runtime.StringValue{Value: "Hello"},
			},
			expected: 0,
		},
		{
			name: "different case - not equal",
			args: []Value{
				&runtime.StringValue{Value: "Hello"},
				&runtime.StringValue{Value: "hello"},
			},
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AnsiCompareStr(ctx, tt.args)
			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("AnsiCompareStr() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

func TestStrMatches(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected bool
	}{
		{
			name: "simple wildcard match",
			args: []Value{
				&runtime.StringValue{Value: "hello.txt"},
				&runtime.StringValue{Value: "*.txt"},
			},
			expected: true,
		},
		{
			name: "no match",
			args: []Value{
				&runtime.StringValue{Value: "hello.doc"},
				&runtime.StringValue{Value: "*.txt"},
			},
			expected: false,
		},
		{
			name: "question mark wildcard",
			args: []Value{
				&runtime.StringValue{Value: "hello"},
				&runtime.StringValue{Value: "h?llo"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrMatches(ctx, tt.args)
			boolVal, ok := result.(*runtime.BooleanValue)
			if !ok {
				t.Fatalf("expected BooleanValue, got %T", result)
			}
			if boolVal.Value != tt.expected {
				t.Errorf("StrMatches() = %v, want %v", boolVal.Value, tt.expected)
			}
		})
	}
}

func TestStrIsASCII(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected bool
	}{
		{
			name:     "ASCII string",
			args:     []Value{&runtime.StringValue{Value: "Hello World"}},
			expected: true,
		},
		{
			name:     "empty string",
			args:     []Value{&runtime.StringValue{Value: ""}},
			expected: true,
		},
		{
			name:     "non-ASCII string",
			args:     []Value{&runtime.StringValue{Value: "Hello 世界"}},
			expected: false,
		},
		{
			name:     "ASCII with numbers",
			args:     []Value{&runtime.StringValue{Value: "Test123"}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrIsASCII(ctx, tt.args)
			boolVal, ok := result.(*runtime.BooleanValue)
			if !ok {
				t.Fatalf("expected BooleanValue, got %T", result)
			}
			if boolVal.Value != tt.expected {
				t.Errorf("StrIsASCII() = %v, want %v", boolVal.Value, tt.expected)
			}
		})
	}
}
