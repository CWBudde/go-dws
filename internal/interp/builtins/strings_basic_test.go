package builtins

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// =============================================================================
// Basic String Functions Tests
// =============================================================================

func TestConcat(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		expected string
		args     []Value
		isError  bool
	}{
		{
			name: "two strings",
			args: []Value{
				&runtime.StringValue{Value: "Hello"},
				&runtime.StringValue{Value: "World"},
			},
			expected: "HelloWorld",
		},
		{
			name: "three strings",
			args: []Value{
				&runtime.StringValue{Value: "a"},
				&runtime.StringValue{Value: "b"},
				&runtime.StringValue{Value: "c"},
			},
			expected: "abc",
		},
		{
			name:    "no arguments",
			args:    []Value{},
			isError: true,
		},
		{
			name: "empty strings",
			args: []Value{
				&runtime.StringValue{Value: ""},
				&runtime.StringValue{Value: ""},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Concat(ctx, tt.args)

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
				t.Errorf("Concat() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestPos(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected int64
	}{
		{
			name: "substring at beginning",
			args: []Value{
				&runtime.StringValue{Value: "Hello"},
				&runtime.StringValue{Value: "Hello World"},
			},
			expected: 1,
		},
		{
			name: "substring in middle",
			args: []Value{
				&runtime.StringValue{Value: "World"},
				&runtime.StringValue{Value: "Hello World"},
			},
			expected: 7,
		},
		{
			name: "substring not found",
			args: []Value{
				&runtime.StringValue{Value: "xyz"},
				&runtime.StringValue{Value: "Hello World"},
			},
			expected: 0,
		},
		{
			name: "empty substring",
			args: []Value{
				&runtime.StringValue{Value: ""},
				&runtime.StringValue{Value: "Hello"},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Pos(ctx, tt.args)
			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("Pos() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

func TestUpperCase(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		expected string
		args     []Value
	}{
		{
			name:     "lowercase to uppercase",
			args:     []Value{&runtime.StringValue{Value: "hello"}},
			expected: "HELLO",
		},
		{
			name:     "mixed case",
			args:     []Value{&runtime.StringValue{Value: "HeLLo WoRLd"}},
			expected: "HELLO WORLD",
		},
		{
			name:     "already uppercase",
			args:     []Value{&runtime.StringValue{Value: "HELLO"}},
			expected: "HELLO",
		},
		{
			name:     "empty string",
			args:     []Value{&runtime.StringValue{Value: ""}},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UpperCase(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("UpperCase() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestLowerCase(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		expected string
		args     []Value
	}{
		{
			name:     "uppercase to lowercase",
			args:     []Value{&runtime.StringValue{Value: "HELLO"}},
			expected: "hello",
		},
		{
			name:     "mixed case",
			args:     []Value{&runtime.StringValue{Value: "HeLLo WoRLd"}},
			expected: "hello world",
		},
		{
			name:     "already lowercase",
			args:     []Value{&runtime.StringValue{Value: "hello"}},
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LowerCase(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("LowerCase() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestTrim(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		expected string
		args     []Value
	}{
		{
			name:     "trim spaces",
			args:     []Value{&runtime.StringValue{Value: "  hello  "}},
			expected: "hello",
		},
		{
			name:     "no spaces",
			args:     []Value{&runtime.StringValue{Value: "hello"}},
			expected: "hello",
		},
		{
			name:     "only spaces",
			args:     []Value{&runtime.StringValue{Value: "   "}},
			expected: "",
		},
		{
			name:     "tabs and spaces",
			args:     []Value{&runtime.StringValue{Value: "\t  hello  \t"}},
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Trim(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("Trim() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestTrimLeft(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		expected string
		args     []Value
	}{
		{
			name:     "trim left spaces",
			args:     []Value{&runtime.StringValue{Value: "  hello  "}},
			expected: "hello  ",
		},
		{
			name:     "no left spaces",
			args:     []Value{&runtime.StringValue{Value: "hello  "}},
			expected: "hello  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TrimLeft(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("TrimLeft() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestTrimRight(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		expected string
		args     []Value
	}{
		{
			name:     "trim right spaces",
			args:     []Value{&runtime.StringValue{Value: "  hello  "}},
			expected: "  hello",
		},
		{
			name:     "no right spaces",
			args:     []Value{&runtime.StringValue{Value: "  hello"}},
			expected: "  hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TrimRight(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("TrimRight() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestStringReplace(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		expected string
		args     []Value
	}{
		{
			name: "replace all occurrences",
			args: []Value{
				&runtime.StringValue{Value: "hello world hello"},
				&runtime.StringValue{Value: "hello"},
				&runtime.StringValue{Value: "hi"},
			},
			expected: "hi world hi",
		},
		{
			name: "replace with empty string",
			args: []Value{
				&runtime.StringValue{Value: "hello world"},
				&runtime.StringValue{Value: "world"},
				&runtime.StringValue{Value: ""},
			},
			expected: "hello ",
		},
		{
			name: "no match",
			args: []Value{
				&runtime.StringValue{Value: "hello world"},
				&runtime.StringValue{Value: "xyz"},
				&runtime.StringValue{Value: "abc"},
			},
			expected: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringReplace(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("StringReplace() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestStringOfChar(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		expected string
		args     []Value
		isError  bool
	}{
		{
			name: "repeat 'a' 5 times",
			args: []Value{
				&runtime.StringValue{Value: "a"},
				&runtime.IntegerValue{Value: 5},
			},
			expected: "aaaaa",
		},
		{
			name: "repeat first char of 'ab' 3 times",
			args: []Value{
				&runtime.StringValue{Value: "ab"},
				&runtime.IntegerValue{Value: 3},
			},
			expected: "aaa",
		},
		{
			name: "zero repetitions",
			args: []Value{
				&runtime.StringValue{Value: "a"},
				&runtime.IntegerValue{Value: 0},
			},
			expected: "",
		},
		{
			name: "negative repetitions",
			args: []Value{
				&runtime.StringValue{Value: "a"},
				&runtime.IntegerValue{Value: -1},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringOfChar(ctx, tt.args)

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
				t.Errorf("StringOfChar() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestSubStr(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		expected string
		args     []Value
	}{
		{
			name: "normal substring",
			args: []Value{
				&runtime.StringValue{Value: "Hello World"},
				&runtime.IntegerValue{Value: 1},
				&runtime.IntegerValue{Value: 5},
			},
			expected: "Hello",
		},
		{
			name: "from middle",
			args: []Value{
				&runtime.StringValue{Value: "Hello World"},
				&runtime.IntegerValue{Value: 7},
				&runtime.IntegerValue{Value: 5},
			},
			expected: "World",
		},
		{
			name: "length exceeds string",
			args: []Value{
				&runtime.StringValue{Value: "Hello"},
				&runtime.IntegerValue{Value: 1},
				&runtime.IntegerValue{Value: 100},
			},
			expected: "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SubStr(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("SubStr() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestIntToHex(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		expected string
		args     []Value
	}{
		{
			name: "255 with width 0",
			args: []Value{
				&runtime.IntegerValue{Value: 255},
				&runtime.IntegerValue{Value: 0},
			},
			expected: "FF",
		},
		{
			name: "255 with width 4",
			args: []Value{
				&runtime.IntegerValue{Value: 255},
				&runtime.IntegerValue{Value: 4},
			},
			expected: "00FF",
		},
		{
			name: "zero with width 0",
			args: []Value{
				&runtime.IntegerValue{Value: 0},
				&runtime.IntegerValue{Value: 0},
			},
			expected: "0",
		},
		{
			name: "16 with width 0",
			args: []Value{
				&runtime.IntegerValue{Value: 16},
				&runtime.IntegerValue{Value: 0},
			},
			expected: "10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IntToHex(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("IntToHex() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestStrToBool(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected bool
	}{
		{
			name:     "true",
			args:     []Value{&runtime.StringValue{Value: "true"}},
			expected: true,
		},
		{
			name:     "True (case insensitive)",
			args:     []Value{&runtime.StringValue{Value: "True"}},
			expected: true,
		},
		{
			name:     "1",
			args:     []Value{&runtime.StringValue{Value: "1"}},
			expected: true,
		},
		{
			name:     "false",
			args:     []Value{&runtime.StringValue{Value: "false"}},
			expected: false,
		},
		{
			name:     "0",
			args:     []Value{&runtime.StringValue{Value: "0"}},
			expected: false,
		},
		{
			name:     "empty string",
			args:     []Value{&runtime.StringValue{Value: ""}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrToBool(ctx, tt.args)
			boolVal, ok := result.(*runtime.BooleanValue)
			if !ok {
				t.Fatalf("expected BooleanValue, got %T", result)
			}
			if boolVal.Value != tt.expected {
				t.Errorf("StrToBool() = %v, want %v", boolVal.Value, tt.expected)
			}
		})
	}
}

func TestSubString(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		expected string
		args     []Value
	}{
		{
			name: "substring from position",
			args: []Value{
				&runtime.StringValue{Value: "Hello World"},
				&runtime.IntegerValue{Value: 1},
				&runtime.IntegerValue{Value: 5},
			},
			expected: "Hello",
		},
		{
			name: "substring from middle",
			args: []Value{
				&runtime.StringValue{Value: "Hello World"},
				&runtime.IntegerValue{Value: 7},
				&runtime.IntegerValue{Value: 11},
			},
			expected: "World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SubString(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("SubString() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestLeftStr(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		expected string
		args     []Value
	}{
		{
			name: "left 5 chars",
			args: []Value{
				&runtime.StringValue{Value: "Hello World"},
				&runtime.IntegerValue{Value: 5},
			},
			expected: "Hello",
		},
		{
			name: "left 0 chars",
			args: []Value{
				&runtime.StringValue{Value: "Hello"},
				&runtime.IntegerValue{Value: 0},
			},
			expected: "",
		},
		{
			name: "left exceeds length",
			args: []Value{
				&runtime.StringValue{Value: "Hi"},
				&runtime.IntegerValue{Value: 10},
			},
			expected: "Hi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LeftStr(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("LeftStr() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestRightStr(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		expected string
		args     []Value
	}{
		{
			name: "right 5 chars",
			args: []Value{
				&runtime.StringValue{Value: "Hello World"},
				&runtime.IntegerValue{Value: 5},
			},
			expected: "World",
		},
		{
			name: "right 0 chars",
			args: []Value{
				&runtime.StringValue{Value: "Hello"},
				&runtime.IntegerValue{Value: 0},
			},
			expected: "",
		},
		{
			name: "right exceeds length",
			args: []Value{
				&runtime.StringValue{Value: "Hi"},
				&runtime.IntegerValue{Value: 10},
			},
			expected: "Hi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RightStr(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("RightStr() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}

func TestStrBeginsWith(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected bool
	}{
		{
			name: "starts with prefix",
			args: []Value{
				&runtime.StringValue{Value: "Hello World"},
				&runtime.StringValue{Value: "Hello"},
			},
			expected: true,
		},
		{
			name: "does not start with prefix",
			args: []Value{
				&runtime.StringValue{Value: "Hello World"},
				&runtime.StringValue{Value: "World"},
			},
			expected: false,
		},
		{
			name: "empty prefix",
			args: []Value{
				&runtime.StringValue{Value: "Hello"},
				&runtime.StringValue{Value: ""},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrBeginsWith(ctx, tt.args)
			boolVal, ok := result.(*runtime.BooleanValue)
			if !ok {
				t.Fatalf("expected BooleanValue, got %T", result)
			}
			if boolVal.Value != tt.expected {
				t.Errorf("StrBeginsWith() = %v, want %v", boolVal.Value, tt.expected)
			}
		})
	}
}

func TestStrEndsWith(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected bool
	}{
		{
			name: "ends with suffix",
			args: []Value{
				&runtime.StringValue{Value: "Hello World"},
				&runtime.StringValue{Value: "World"},
			},
			expected: true,
		},
		{
			name: "does not end with suffix",
			args: []Value{
				&runtime.StringValue{Value: "Hello World"},
				&runtime.StringValue{Value: "Hello"},
			},
			expected: false,
		},
		{
			name: "empty suffix",
			args: []Value{
				&runtime.StringValue{Value: "Hello"},
				&runtime.StringValue{Value: ""},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrEndsWith(ctx, tt.args)
			boolVal, ok := result.(*runtime.BooleanValue)
			if !ok {
				t.Fatalf("expected BooleanValue, got %T", result)
			}
			if boolVal.Value != tt.expected {
				t.Errorf("StrEndsWith() = %v, want %v", boolVal.Value, tt.expected)
			}
		})
	}
}

func TestStrContains(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected bool
	}{
		{
			name: "contains substring",
			args: []Value{
				&runtime.StringValue{Value: "Hello World"},
				&runtime.StringValue{Value: "lo Wo"},
			},
			expected: true,
		},
		{
			name: "does not contain substring",
			args: []Value{
				&runtime.StringValue{Value: "Hello World"},
				&runtime.StringValue{Value: "xyz"},
			},
			expected: false,
		},
		{
			name: "empty substring",
			args: []Value{
				&runtime.StringValue{Value: "Hello"},
				&runtime.StringValue{Value: ""},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrContains(ctx, tt.args)
			boolVal, ok := result.(*runtime.BooleanValue)
			if !ok {
				t.Fatalf("expected BooleanValue, got %T", result)
			}
			if boolVal.Value != tt.expected {
				t.Errorf("StrContains() = %v, want %v", boolVal.Value, tt.expected)
			}
		})
	}
}

func TestPosEx(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		args     []Value
		expected int64
	}{
		{
			name: "find from start",
			args: []Value{
				&runtime.StringValue{Value: "hello"},
				&runtime.StringValue{Value: "hello hello hello"},
				&runtime.IntegerValue{Value: 1},
			},
			expected: 1,
		},
		{
			name: "find from position",
			args: []Value{
				&runtime.StringValue{Value: "hello"},
				&runtime.StringValue{Value: "hello hello hello"},
				&runtime.IntegerValue{Value: 2},
			},
			expected: 7,
		},
		{
			name: "not found",
			args: []Value{
				&runtime.StringValue{Value: "xyz"},
				&runtime.StringValue{Value: "hello world"},
				&runtime.IntegerValue{Value: 1},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PosEx(ctx, tt.args)
			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("PosEx() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

func TestByteSizeToStr(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		expected string
		args     []Value
	}{
		{
			name:     "bytes",
			args:     []Value{&runtime.IntegerValue{Value: 500}},
			expected: "500 bytes",
		},
		{
			name:     "kilobytes",
			args:     []Value{&runtime.IntegerValue{Value: 1024}},
			expected: "1.00 KB",
		},
		{
			name:     "megabytes",
			args:     []Value{&runtime.IntegerValue{Value: 1048576}},
			expected: "1.00 MB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ByteSizeToStr(ctx, tt.args)
			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("ByteSizeToStr() = %q, want %q", strVal.Value, tt.expected)
			}
		})
	}
}
