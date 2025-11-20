package builtins

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// =============================================================================
// Previously Uncovered String Functions Tests
// =============================================================================

func TestASCIIUpperCase(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase string",
			input:    "hello",
			expected: "HELLO",
		},
		{
			name:     "mixed case",
			input:    "Hello World",
			expected: "HELLO WORLD",
		},
		{
			name:     "already uppercase",
			input:    "HELLO",
			expected: "HELLO",
		},
		{
			name:     "with numbers",
			input:    "test123",
			expected: "TEST123",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ASCIIUpperCase(ctx, []Value{&runtime.StringValue{Value: tt.input}})

			if result.Type() == "ERROR" {
				t.Errorf("ASCIIUpperCase() returned error: %v", result)
				return
			}

			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("ASCIIUpperCase(%q) = %q, want %q", tt.input, strVal.Value, tt.expected)
			}
		})
	}

	// Test errors
	result := ASCIIUpperCase(ctx, []Value{})
	if result.Type() != "ERROR" {
		t.Errorf("ASCIIUpperCase() with 0 arguments should error")
	}
}

func TestASCIILowerCase(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "uppercase string",
			input:    "HELLO",
			expected: "hello",
		},
		{
			name:     "mixed case",
			input:    "Hello World",
			expected: "hello world",
		},
		{
			name:     "already lowercase",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "with numbers",
			input:    "TEST123",
			expected: "test123",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ASCIILowerCase(ctx, []Value{&runtime.StringValue{Value: tt.input}})

			if result.Type() == "ERROR" {
				t.Errorf("ASCIILowerCase() returned error: %v", result)
				return
			}

			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("ASCIILowerCase(%q) = %q, want %q", tt.input, strVal.Value, tt.expected)
			}
		})
	}

	// Test errors
	result := ASCIILowerCase(ctx, []Value{})
	if result.Type() != "ERROR" {
		t.Errorf("ASCIILowerCase() with 0 arguments should error")
	}
}

func TestAnsiUpperCase(t *testing.T) {
	ctx := newMockContext()

	result := AnsiUpperCase(ctx, []Value{&runtime.StringValue{Value: "hello"}})
	if result.Type() == "ERROR" {
		t.Errorf("AnsiUpperCase() returned error: %v", result)
	}

	// Test errors
	result = AnsiUpperCase(ctx, []Value{})
	if result.Type() != "ERROR" {
		t.Errorf("AnsiUpperCase() with 0 arguments should error")
	}
}

func TestAnsiLowerCase(t *testing.T) {
	ctx := newMockContext()

	result := AnsiLowerCase(ctx, []Value{&runtime.StringValue{Value: "HELLO"}})
	if result.Type() == "ERROR" {
		t.Errorf("AnsiLowerCase() returned error: %v", result)
	}

	// Test errors
	result = AnsiLowerCase(ctx, []Value{})
	if result.Type() != "ERROR" {
		t.Errorf("AnsiLowerCase() with 0 arguments should error")
	}
}

func TestMidStr(t *testing.T) {
	ctx := newMockContext()

	result := MidStr(ctx, []Value{
		&runtime.StringValue{Value: "Hello World"},
		&runtime.IntegerValue{Value: 2},
		&runtime.IntegerValue{Value: 5},
	})
	if result.Type() == "ERROR" {
		t.Errorf("MidStr() returned error: %v", result)
	}

	// Test errors
	result = MidStr(ctx, []Value{})
	if result.Type() != "ERROR" {
		t.Errorf("MidStr() with 0 arguments should error")
	}
}

func TestRevPos(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		substr   string
		str      string
		expected int64
	}{
		{
			name:     "substring at end",
			substr:   "World",
			str:      "Hello World",
			expected: 7,
		},
		{
			name:     "substring not found",
			substr:   "xyz",
			str:      "Hello World",
			expected: 0,
		},
		{
			name:     "empty substring",
			substr:   "",
			str:      "Hello",
			expected: 6, // RevPos returns len(haystack) + 1 for empty needle
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RevPos(ctx, []Value{
				&runtime.StringValue{Value: tt.substr},
				&runtime.StringValue{Value: tt.str},
			})

			if result.Type() == "ERROR" {
				t.Errorf("RevPos() returned error: %v", result)
				return
			}

			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("RevPos(%q, %q) = %d, want %d", tt.substr, tt.str, intVal.Value, tt.expected)
			}
		})
	}

	// Test errors
	result := RevPos(ctx, []Value{})
	if result.Type() != "ERROR" {
		t.Errorf("RevPos() with 0 arguments should error")
	}
}

func TestStrFind(t *testing.T) {
	ctx := newMockContext()

	result := StrFind(ctx, []Value{
		&runtime.StringValue{Value: "Hello World"}, // str
		&runtime.StringValue{Value: "World"},       // substr
		&runtime.IntegerValue{Value: 1},            // fromIndex
	})
	if result.Type() == "ERROR" {
		t.Errorf("StrFind() returned error: %v", result)
	}

	// Test errors
	result = StrFind(ctx, []Value{})
	if result.Type() != "ERROR" {
		t.Errorf("StrFind() with 0 arguments should error")
	}
}

func TestGetText(t *testing.T) {
	ctx := newMockContext()

	result := GetText(ctx, []Value{
		&runtime.StringValue{Value: "key"},
	})
	if result.Type() == "ERROR" {
		t.Errorf("GetText() returned error: %v", result)
	}

	// Test errors
	result = GetText(ctx, []Value{})
	if result.Type() != "ERROR" {
		t.Errorf("GetText() with 0 arguments should error")
	}
}

func TestUnderscore(t *testing.T) {
	ctx := newMockContext()

	result := Underscore(ctx, []Value{
		&runtime.StringValue{Value: "HelloWorld"},
	})
	if result.Type() == "ERROR" {
		t.Errorf("Underscore() returned error: %v", result)
	}

	// Test errors
	result = Underscore(ctx, []Value{})
	if result.Type() != "ERROR" {
		t.Errorf("Underscore() with 0 arguments should error")
	}
}

func TestCharAt(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name     string
		str      string
		expected string
		index    int64
		isError  bool
	}{
		{
			name:     "first character",
			str:      "Hello",
			index:    1,
			expected: "H",
		},
		{
			name:     "middle character",
			str:      "Hello",
			index:    3,
			expected: "l",
		},
		{
			name:     "index out of bounds",
			str:      "Hello",
			index:    10,
			expected: "", // CharAt returns empty string for out-of-bounds
		},
		{
			name:     "zero index",
			str:      "Hello",
			index:    0,
			expected: "", // CharAt returns empty string for index < 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CharAt(ctx, []Value{
				&runtime.StringValue{Value: tt.str},
				&runtime.IntegerValue{Value: tt.index},
			})

			if tt.isError {
				if result.Type() != "ERROR" {
					t.Errorf("expected error, got %v", result)
				}
				return
			}

			if result.Type() == "ERROR" {
				t.Errorf("CharAt() returned error: %v", result)
				return
			}

			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected StringValue, got %T", result)
			}
			if strVal.Value != tt.expected {
				t.Errorf("CharAt(%q, %d) = %q, want %q", tt.str, tt.index, strVal.Value, tt.expected)
			}
		})
	}

	// Test errors
	result := CharAt(ctx, []Value{})
	if result.Type() != "ERROR" {
		t.Errorf("CharAt() with 0 arguments should error")
	}
}

func TestCompareLocaleStr(t *testing.T) {
	ctx := newMockContext()

	result := CompareLocaleStr(ctx, []Value{
		&runtime.StringValue{Value: "Hello"},
		&runtime.StringValue{Value: "World"},
	})
	if result.Type() == "ERROR" {
		t.Errorf("CompareLocaleStr() returned error: %v", result)
	}

	// Test errors
	result = CompareLocaleStr(ctx, []Value{})
	if result.Type() != "ERROR" {
		t.Errorf("CompareLocaleStr() with 0 arguments should error")
	}
}

func TestLastDelimiter(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name       string
		delimiters string
		str        string
		expected   int64
	}{
		{
			name:       "find last delimiter",
			delimiters: "/\\",
			str:        "path/to\\file",
			expected:   8, // position of backslash
		},
		{
			name:       "no delimiter found",
			delimiters: "/",
			str:        "no-delimiters",
			expected:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LastDelimiter(ctx, []Value{
				&runtime.StringValue{Value: tt.delimiters},
				&runtime.StringValue{Value: tt.str},
			})

			if result.Type() == "ERROR" {
				t.Errorf("LastDelimiter() returned error: %v", result)
				return
			}

			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("LastDelimiter(%q, %q) = %d, want %d", tt.delimiters, tt.str, intVal.Value, tt.expected)
			}
		})
	}
}

func TestFindDelimiter(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name       string
		delimiters string
		str        string
		startPos   int64
		expected   int64
	}{
		{
			name:       "find delimiter from start",
			delimiters: "/",
			str:        "path/to/file",
			startPos:   1,
			expected:   5,
		},
		{
			name:       "no delimiter found",
			delimiters: "/",
			str:        "no-delimiters",
			startPos:   1,
			expected:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindDelimiter(ctx, []Value{
				&runtime.StringValue{Value: tt.delimiters},
				&runtime.StringValue{Value: tt.str},
				&runtime.IntegerValue{Value: tt.startPos},
			})

			if result.Type() == "ERROR" {
				t.Errorf("FindDelimiter() returned error: %v", result)
				return
			}

			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T", result)
			}
			if intVal.Value != tt.expected {
				t.Errorf("FindDelimiter(%q, %q, %d) = %d, want %d", tt.delimiters, tt.str, tt.startPos, intVal.Value, tt.expected)
			}
		})
	}
}

func TestNormalizeString(t *testing.T) {
	ctx := newMockContext()

	result := NormalizeString(ctx, []Value{
		&runtime.StringValue{Value: "  Hello   World  "},
	})
	if result.Type() == "ERROR" {
		t.Errorf("NormalizeString() returned error: %v", result)
	}

	// Test errors
	result = NormalizeString(ctx, []Value{})
	if result.Type() != "ERROR" {
		t.Errorf("NormalizeString() with 0 arguments should error")
	}
}

func TestStripAccents(t *testing.T) {
	ctx := newMockContext()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "string with accents",
			input: "café",
		},
		{
			name:  "plain string",
			input: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripAccents(ctx, []Value{
				&runtime.StringValue{Value: tt.input},
			})
			if result.Type() == "ERROR" {
				t.Errorf("StripAccents() returned error: %v", result)
			}
		})
	}

	// Test errors
	result := StripAccents(ctx, []Value{})
	if result.Type() != "ERROR" {
		t.Errorf("StripAccents() with 0 arguments should error")
	}
}
