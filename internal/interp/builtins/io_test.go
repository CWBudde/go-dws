package builtins

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// mockIOContext is a test context that captures output
type mockIOContext struct {
	*mockContext
	output bytes.Buffer
}

func newMockIOContext() *mockIOContext {
	return &mockIOContext{
		mockContext: newMockContext(),
	}
}

func (m *mockIOContext) Write(s string) {
	m.output.WriteString(s)
}

func (m *mockIOContext) WriteLine(s string) {
	m.output.WriteString(s + "\n")
}

func TestPrint(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		args     []Value
	}{
		{
			name:     "Single string",
			args:     []Value{&runtime.StringValue{Value: "hello"}},
			expected: "hello",
		},
		{
			name: "Multiple strings",
			args: []Value{
				&runtime.StringValue{Value: "hello"},
				&runtime.StringValue{Value: " "},
				&runtime.StringValue{Value: "world"},
			},
			expected: "hello world",
		},
		{
			name:     "Integer",
			args:     []Value{&runtime.IntegerValue{Value: 42}},
			expected: "42",
		},
		{
			name:     "Boolean",
			args:     []Value{&runtime.BooleanValue{Value: true}},
			expected: "True",
		},
		{
			name:     "Nil argument",
			args:     []Value{nil},
			expected: "<nil>",
		},
		{
			name: "Mixed types",
			args: []Value{
				&runtime.StringValue{Value: "Answer: "},
				&runtime.IntegerValue{Value: 42},
			},
			expected: "Answer: 42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newMockIOContext()
			result := Print(ctx, tt.args)

			// Check that result is NilValue
			if _, ok := result.(*runtime.NilValue); !ok {
				t.Errorf("Print() should return NilValue, got %T", result)
			}

			// Check output
			if ctx.output.String() != tt.expected {
				t.Errorf("Print() output = %q, want %q", ctx.output.String(), tt.expected)
			}
		})
	}
}

func TestPrintLn(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		args     []Value
	}{
		{
			name:     "Single string",
			args:     []Value{&runtime.StringValue{Value: "hello"}},
			expected: "hello\n",
		},
		{
			name: "Multiple strings",
			args: []Value{
				&runtime.StringValue{Value: "hello"},
				&runtime.StringValue{Value: " "},
				&runtime.StringValue{Value: "world"},
			},
			expected: "hello world\n",
		},
		{
			name:     "Integer",
			args:     []Value{&runtime.IntegerValue{Value: 42}},
			expected: "42\n",
		},
		{
			name:     "Empty (just newline)",
			args:     []Value{},
			expected: "\n",
		},
		{
			name:     "Nil argument",
			args:     []Value{nil},
			expected: "<nil>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newMockIOContext()
			result := PrintLn(ctx, tt.args)

			// Check that result is NilValue
			if _, ok := result.(*runtime.NilValue); !ok {
				t.Errorf("PrintLn() should return NilValue, got %T", result)
			}

			// Check output
			if ctx.output.String() != tt.expected {
				t.Errorf("PrintLn() output = %q, want %q", ctx.output.String(), tt.expected)
			}
		})
	}
}

func TestPrintAndPrintLnSequence(t *testing.T) {
	ctx := newMockIOContext()

	// Print without newline
	Print(ctx, []Value{&runtime.StringValue{Value: "Hello"}})
	// Print without newline
	Print(ctx, []Value{&runtime.StringValue{Value: " "}})
	// Print without newline
	Print(ctx, []Value{&runtime.StringValue{Value: "World"}})
	// PrintLn adds newline
	PrintLn(ctx, []Value{})

	expected := "Hello World\n"
	if ctx.output.String() != expected {
		t.Errorf("Sequence output = %q, want %q", ctx.output.String(), expected)
	}
}
