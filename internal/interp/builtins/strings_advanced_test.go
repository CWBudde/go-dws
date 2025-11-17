package builtins

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// =============================================================================
// Advanced String Functions Tests
// =============================================================================

func TestStrBefore(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected string
	}{
		{
			name: "text before delimiter",
			args: []Value{
				&runtime.StringValue{Value: "Hello@World"},
				&runtime.StringValue{Value: "@"},
			},
			expected: "Hello",
		},
		{
			name: "delimiter not found",
			args: []Value{
				&runtime.StringValue{Value: "HelloWorld"},
				&runtime.StringValue{Value: "@"},
			},
			expected: "HelloWorld",
		},
		{
			name: "delimiter at start",
			args: []Value{
				&runtime.StringValue{Value: "@World"},
				&runtime.StringValue{Value: "@"},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrBefore(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("StrBefore() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestStrAfter(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected string
	}{
		{
			name: "text after delimiter",
			args: []Value{
				&runtime.StringValue{Value: "Hello@World"},
				&runtime.StringValue{Value: "@"},
			},
			expected: "World",
		},
		{
			name: "delimiter not found",
			args: []Value{
				&runtime.StringValue{Value: "HelloWorld"},
				&runtime.StringValue{Value: "@"},
			},
			expected: "",
		},
		{
			name: "delimiter at end",
			args: []Value{
				&runtime.StringValue{Value: "Hello@"},
				&runtime.StringValue{Value: "@"},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrAfter(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("StrAfter() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestStrBeforeLast(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected string
	}{
		{
			name: "text before last delimiter",
			args: []Value{
				&runtime.StringValue{Value: "a@b@c"},
				&runtime.StringValue{Value: "@"},
			},
			expected: "a@b",
		},
		{
			name: "delimiter not found",
			args: []Value{
				&runtime.StringValue{Value: "abc"},
				&runtime.StringValue{Value: "@"},
			},
			expected: "abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrBeforeLast(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("StrBeforeLast() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestStrAfterLast(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected string
	}{
		{
			name: "text after last delimiter",
			args: []Value{
				&runtime.StringValue{Value: "a@b@c"},
				&runtime.StringValue{Value: "@"},
			},
			expected: "c",
		},
		{
			name: "delimiter not found",
			args: []Value{
				&runtime.StringValue{Value: "abc"},
				&runtime.StringValue{Value: "@"},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrAfterLast(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("StrAfterLast() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestStrBetween(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected string
	}{
		{
			name: "text between delimiters",
			args: []Value{
				&runtime.StringValue{Value: "Hello[World]End"},
				&runtime.StringValue{Value: "["},
				&runtime.StringValue{Value: "]"},
			},
			expected: "World",
		},
		{
			name: "start delimiter not found",
			args: []Value{
				&runtime.StringValue{Value: "HelloWorld]End"},
				&runtime.StringValue{Value: "["},
				&runtime.StringValue{Value: "]"},
			},
			expected: "",
		},
		{
			name: "end delimiter not found",
			args: []Value{
				&runtime.StringValue{Value: "Hello[WorldEnd"},
				&runtime.StringValue{Value: "["},
				&runtime.StringValue{Value: "]"},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrBetween(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("StrBetween() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestIsDelimiter(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected bool
	}{
		{
			name: "is delimiter",
			args: []Value{
				&runtime.StringValue{Value: " ,.;"},
				&runtime.StringValue{Value: "Hello,World"},
				&runtime.IntegerValue{Value: 6},
			},
			expected: true,
		},
		{
			name: "not delimiter",
			args: []Value{
				&runtime.StringValue{Value: " ,.;"},
				&runtime.StringValue{Value: "Hello World"},
				&runtime.IntegerValue{Value: 1},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDelimiter(ctx, tt.args)
			boolVal, ok := result.(*runtime.BooleanValue)
			if !ok {
				t.Fatalf("expected BooleanValue, got %T", result)
			}
			if boolVal.Value != tt.expected {
				t.Errorf("IsDelimiter() = %v, want %v", boolVal.Value, tt.expected)
			}
		})
	}
}

func TestPadLeft(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected string
	}{
		{
			name: "pad with spaces",
			args: []Value{
				&runtime.StringValue{Value: "hello"},
				&runtime.IntegerValue{Value: 10},
			},
			expected: "     hello",
		},
		{
			name: "pad with custom char",
			args: []Value{
				&runtime.StringValue{Value: "hello"},
				&runtime.IntegerValue{Value: 10},
				&runtime.StringValue{Value: "0"},
			},
			expected: "00000hello",
		},
		{
			name: "no padding needed",
			args: []Value{
				&runtime.StringValue{Value: "hello"},
				&runtime.IntegerValue{Value: 3},
			},
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PadLeft(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("PadLeft() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestPadRight(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected string
	}{
		{
			name: "pad with spaces",
			args: []Value{
				&runtime.StringValue{Value: "hello"},
				&runtime.IntegerValue{Value: 10},
			},
			expected: "hello     ",
		},
		{
			name: "pad with custom char",
			args: []Value{
				&runtime.StringValue{Value: "hello"},
				&runtime.IntegerValue{Value: 10},
				&runtime.StringValue{Value: "0"},
			},
			expected: "hello00000",
		},
		{
			name: "no padding needed",
			args: []Value{
				&runtime.StringValue{Value: "hello"},
				&runtime.IntegerValue{Value: 3},
			},
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PadRight(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("PadRight() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestStrDeleteLeft(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected string
	}{
		{
			name: "delete left 5 chars",
			args: []Value{
				&runtime.StringValue{Value: "Hello World"},
				&runtime.IntegerValue{Value: 5},
			},
			expected: " World",
		},
		{
			name: "delete all",
			args: []Value{
				&runtime.StringValue{Value: "Hello"},
				&runtime.IntegerValue{Value: 10},
			},
			expected: "",
		},
		{
			name: "delete zero",
			args: []Value{
				&runtime.StringValue{Value: "Hello"},
				&runtime.IntegerValue{Value: 0},
			},
			expected: "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrDeleteLeft(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("StrDeleteLeft() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestStrDeleteRight(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected string
	}{
		{
			name: "delete right 5 chars",
			args: []Value{
				&runtime.StringValue{Value: "Hello World"},
				&runtime.IntegerValue{Value: 5},
			},
			expected: "Hello ",
		},
		{
			name: "delete all",
			args: []Value{
				&runtime.StringValue{Value: "Hello"},
				&runtime.IntegerValue{Value: 10},
			},
			expected: "",
		},
		{
			name: "delete zero",
			args: []Value{
				&runtime.StringValue{Value: "Hello"},
				&runtime.IntegerValue{Value: 0},
			},
			expected: "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrDeleteRight(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("StrDeleteRight() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestReverseString(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected string
	}{
		{
			name:     "simple string",
			args:     []Value{&runtime.StringValue{Value: "hello"}},
			expected: "olleh",
		},
		{
			name:     "empty string",
			args:     []Value{&runtime.StringValue{Value: ""}},
			expected: "",
		},
		{
			name:     "single char",
			args:     []Value{&runtime.StringValue{Value: "a"}},
			expected: "a",
		},
		{
			name:     "with spaces",
			args:     []Value{&runtime.StringValue{Value: "hello world"}},
			expected: "dlrow olleh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReverseString(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("ReverseString() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestQuotedStr(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected string
	}{
		{
			name:     "simple string",
			args:     []Value{&runtime.StringValue{Value: "hello"}},
			expected: "'hello'",
		},
		{
			name:     "empty string",
			args:     []Value{&runtime.StringValue{Value: ""}},
			expected: "''",
		},
		{
			name:     "string with quotes",
			args:     []Value{&runtime.StringValue{Value: "say \"hi\""}},
			expected: "'say \"hi\"'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := QuotedStr(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("QuotedStr() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestDupeString(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected string
	}{
		{
			name: "duplicate 3 times",
			args: []Value{
				&runtime.StringValue{Value: "ab"},
				&runtime.IntegerValue{Value: 3},
			},
			expected: "ababab",
		},
		{
			name: "duplicate 0 times",
			args: []Value{
				&runtime.StringValue{Value: "ab"},
				&runtime.IntegerValue{Value: 0},
			},
			expected: "",
		},
		{
			name: "duplicate 1 time",
			args: []Value{
				&runtime.StringValue{Value: "hello"},
				&runtime.IntegerValue{Value: 1},
			},
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DupeString(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("DupeString() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}
