package builtins

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// mockEnumValue is a test-only enum value
type mockEnumValue struct {
	ordinal int64
}

func (m *mockEnumValue) Type() string   { return "ENUM" }
func (m *mockEnumValue) String() string { return "MockEnum" }

// mockOrdContext extends mockContext to handle enum ordinals
type mockOrdContext struct {
	*mockContext
}

func newMockOrdContext() *mockOrdContext {
	return &mockOrdContext{
		mockContext: newMockContext(),
	}
}

func (m *mockOrdContext) GetEnumOrdinal(value Value) (int64, bool) {
	if enumVal, ok := value.(*mockEnumValue); ok {
		return enumVal.ordinal, true
	}
	return 0, false
}

func TestOrd(t *testing.T) {
	tests := []struct {
		name     string
		args     []Value
		expected int64
	}{
		{
			name:     "Boolean True",
			args:     []Value{&runtime.BooleanValue{Value: true}},
			expected: 1,
		},
		{
			name:     "Boolean False",
			args:     []Value{&runtime.BooleanValue{Value: false}},
			expected: 0,
		},
		{
			name:     "Integer passthrough",
			args:     []Value{&runtime.IntegerValue{Value: 42}},
			expected: 42,
		},
		{
			name:     "Character 'A'",
			args:     []Value{&runtime.StringValue{Value: "A"}},
			expected: 65,
		},
		{
			name:     "Character '0'",
			args:     []Value{&runtime.StringValue{Value: "0"}},
			expected: 48,
		},
		{
			name:     "Unicode character '€'",
			args:     []Value{&runtime.StringValue{Value: "€"}},
			expected: 8364,
		},
		{
			name:     "Empty string",
			args:     []Value{&runtime.StringValue{Value: ""}},
			expected: 0,
		},
		{
			name:     "Multi-character string (first char)",
			args:     []Value{&runtime.StringValue{Value: "ABC"}},
			expected: 65,
		},
		{
			name:     "Enum value",
			args:     []Value{&mockEnumValue{ordinal: 0}},
			expected: 0,
		},
		{
			name:     "Enum value with ordinal 5",
			args:     []Value{&mockEnumValue{ordinal: 5}},
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newMockOrdContext()
			result := Ord(ctx, tt.args)

			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("Ord() returned %T, expected IntegerValue", result)
			}

			if intVal.Value != tt.expected {
				t.Errorf("Ord() = %d, want %d", intVal.Value, tt.expected)
			}
		})
	}
}

func TestOrd_ErrorCases(t *testing.T) {
	tests := []struct {
		name string
		args []Value
	}{
		{
			name: "Wrong number of arguments (0)",
			args: []Value{},
		},
		{
			name: "Wrong number of arguments (2)",
			args: []Value{&runtime.IntegerValue{Value: 1}, &runtime.IntegerValue{Value: 2}},
		},
		{
			name: "Invalid type (Float)",
			args: []Value{&runtime.FloatValue{Value: 3.14}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newMockContext()
			result := Ord(ctx, tt.args)

			if result.Type() != "ERROR" {
				t.Errorf("expected error for invalid Ord() call, got %s: %v", result.Type(), result)
			}
		})
	}
}

func TestChr(t *testing.T) {
	tests := []struct {
		name     string
		code     int64
		expected string
	}{
		{
			name:     "Character code for 'A'",
			code:     65,
			expected: "A",
		},
		{
			name:     "Character code for '0'",
			code:     48,
			expected: "0",
		},
		{
			name:     "Character code for space",
			code:     32,
			expected: " ",
		},
		{
			name:     "Unicode euro sign",
			code:     8364,
			expected: "€",
		},
		{
			name:     "Null character",
			code:     0,
			expected: "\x00",
		},
		{
			name:     "High Unicode (emoji)",
			code:     0x1F600, // 😀
			expected: "😀",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newMockContext()
			args := []Value{&runtime.IntegerValue{Value: tt.code}}
			result := Chr(ctx, args)

			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("Chr() returned %T, expected StringValue", result)
			}

			if strVal.Value != tt.expected {
				t.Errorf("Chr(%d) = %q, want %q", tt.code, strVal.Value, tt.expected)
			}
		})
	}
}

func TestChr_ErrorCases(t *testing.T) {
	tests := []struct {
		name string
		args []Value
	}{
		{
			name: "Wrong number of arguments (0)",
			args: []Value{},
		},
		{
			name: "Wrong number of arguments (2)",
			args: []Value{&runtime.IntegerValue{Value: 65}, &runtime.IntegerValue{Value: 66}},
		},
		{
			name: "Non-integer argument",
			args: []Value{&runtime.StringValue{Value: "65"}},
		},
		{
			name: "Negative code",
			args: []Value{&runtime.IntegerValue{Value: -1}},
		},
		{
			name: "Code above Unicode range",
			args: []Value{&runtime.IntegerValue{Value: 0x110000}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newMockContext()
			result := Chr(ctx, tt.args)

			if result.Type() != "ERROR" {
				t.Errorf("expected error for invalid Chr() call, got %s: %v", result.Type(), result)
			}
		})
	}
}

func TestOrdChrRoundTrip(t *testing.T) {
	// Test that Chr(Ord(x)) == x for single characters
	testChars := []string{"A", "z", "0", "9", " ", "€", "😀"}

	for _, char := range testChars {
		t.Run("RoundTrip_"+char, func(t *testing.T) {
			ctx := newMockContext()

			// Get ordinal value
			ordResult := Ord(ctx, []Value{&runtime.StringValue{Value: char}})
			ordInt, ok := ordResult.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("Ord() failed to return integer")
			}

			// Convert back to character
			chrResult := Chr(ctx, []Value{ordInt})
			chrStr, ok := chrResult.(*runtime.StringValue)
			if !ok {
				t.Fatalf("Chr() failed to return string")
			}

			if chrStr.Value != char {
				t.Errorf("RoundTrip failed: Chr(Ord(%q)) = %q, want %q", char, chrStr.Value, char)
			}
		})
	}
}
