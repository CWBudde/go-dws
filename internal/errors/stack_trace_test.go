package errors

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

func TestStackFrame_String(t *testing.T) {
	tests := []struct {
		name     string
		frame    StackFrame
		expected string
	}{
		{
			name: "Frame with position",
			frame: StackFrame{
				FunctionName: "MyFunction",
				FileName:     "test.dws",
				Position:     &lexer.Position{Line: 10, Column: 5},
			},
			expected: "MyFunction [line: 10, column: 5]",
		},
		{
			name: "Frame without position",
			frame: StackFrame{
				FunctionName: "MyFunction",
				FileName:     "test.dws",
				Position:     nil,
			},
			expected: "MyFunction",
		},
		{
			name: "Frame with method name",
			frame: StackFrame{
				FunctionName: "TMyClass.MyMethod",
				FileName:     "test.dws",
				Position:     &lexer.Position{Line: 42, Column: 15},
			},
			expected: "TMyClass.MyMethod [line: 42, column: 15]",
		},
		{
			name: "Frame with lambda",
			frame: StackFrame{
				FunctionName: "<lambda>",
				FileName:     "",
				Position:     &lexer.Position{Line: 7, Column: 1},
			},
			expected: "<lambda> [line: 7, column: 1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.frame.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestStackTrace_String(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		trace    StackTrace
	}{
		{
			name:     "Empty stack trace",
			trace:    StackTrace{},
			expected: "",
		},
		{
			name: "Single frame",
			trace: StackTrace{
				{FunctionName: "Main", Position: &lexer.Position{Line: 1, Column: 1}},
			},
			expected: "Main [line: 1, column: 1]",
		},
		{
			name: "Multiple frames",
			trace: StackTrace{
				{FunctionName: "Main", Position: &lexer.Position{Line: 20, Column: 1}},
				{FunctionName: "Foo", Position: &lexer.Position{Line: 15, Column: 5}},
				{FunctionName: "Bar", Position: &lexer.Position{Line: 10, Column: 3}},
			},
			expected: "Bar [line: 10, column: 3]\nFoo [line: 15, column: 5]\nMain [line: 20, column: 1]",
		},
		{
			name: "Frames with and without position",
			trace: StackTrace{
				{FunctionName: "Main", Position: &lexer.Position{Line: 20, Column: 1}},
				{FunctionName: "Foo", Position: nil},
			},
			expected: "Foo\nMain [line: 20, column: 1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.trace.String()
			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestStackTrace_Reverse(t *testing.T) {
	original := StackTrace{
		{FunctionName: "First", Position: &lexer.Position{Line: 1, Column: 1}},
		{FunctionName: "Second", Position: &lexer.Position{Line: 2, Column: 1}},
		{FunctionName: "Third", Position: &lexer.Position{Line: 3, Column: 1}},
	}

	reversed := original.Reverse()

	// Check that order is reversed
	if reversed[0].FunctionName != "Third" {
		t.Errorf("Expected first frame to be 'Third', got %q", reversed[0].FunctionName)
	}
	if reversed[1].FunctionName != "Second" {
		t.Errorf("Expected second frame to be 'Second', got %q", reversed[1].FunctionName)
	}
	if reversed[2].FunctionName != "First" {
		t.Errorf("Expected third frame to be 'First', got %q", reversed[2].FunctionName)
	}

	// Check that original is unchanged
	if original[0].FunctionName != "First" {
		t.Errorf("Original stack trace was modified")
	}
}

func TestStackTrace_Top(t *testing.T) {
	tests := []struct {
		expected *string
		name     string
		trace    StackTrace
	}{
		{
			name:     "Empty stack",
			trace:    StackTrace{},
			expected: nil,
		},
		{
			name: "Single frame",
			trace: StackTrace{
				{FunctionName: "Main", Position: &lexer.Position{Line: 1, Column: 1}},
			},
			expected: stringPtr("Main"),
		},
		{
			name: "Multiple frames",
			trace: StackTrace{
				{FunctionName: "Main", Position: &lexer.Position{Line: 20, Column: 1}},
				{FunctionName: "Foo", Position: &lexer.Position{Line: 15, Column: 5}},
				{FunctionName: "Bar", Position: &lexer.Position{Line: 10, Column: 3}},
			},
			expected: stringPtr("Bar"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			top := tt.trace.Top()
			if tt.expected == nil {
				if top != nil {
					t.Errorf("Expected nil, got %v", top)
				}
			} else {
				if top == nil {
					t.Errorf("Expected %q, got nil", *tt.expected)
				} else if top.FunctionName != *tt.expected {
					t.Errorf("Expected %q, got %q", *tt.expected, top.FunctionName)
				}
			}
		})
	}
}

func TestStackTrace_Bottom(t *testing.T) {
	tests := []struct {
		expected *string
		name     string
		trace    StackTrace
	}{
		{
			name:     "Empty stack",
			trace:    StackTrace{},
			expected: nil,
		},
		{
			name: "Single frame",
			trace: StackTrace{
				{FunctionName: "Main", Position: &lexer.Position{Line: 1, Column: 1}},
			},
			expected: stringPtr("Main"),
		},
		{
			name: "Multiple frames",
			trace: StackTrace{
				{FunctionName: "Main", Position: &lexer.Position{Line: 20, Column: 1}},
				{FunctionName: "Foo", Position: &lexer.Position{Line: 15, Column: 5}},
				{FunctionName: "Bar", Position: &lexer.Position{Line: 10, Column: 3}},
			},
			expected: stringPtr("Main"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bottom := tt.trace.Bottom()
			if tt.expected == nil {
				if bottom != nil {
					t.Errorf("Expected nil, got %v", bottom)
				}
			} else {
				if bottom == nil {
					t.Errorf("Expected %q, got nil", *tt.expected)
				} else if bottom.FunctionName != *tt.expected {
					t.Errorf("Expected %q, got %q", *tt.expected, bottom.FunctionName)
				}
			}
		})
	}
}

func TestStackTrace_Depth(t *testing.T) {
	tests := []struct {
		name     string
		trace    StackTrace
		expected int
	}{
		{
			name:     "Empty stack",
			trace:    StackTrace{},
			expected: 0,
		},
		{
			name: "Single frame",
			trace: StackTrace{
				{FunctionName: "Main"},
			},
			expected: 1,
		},
		{
			name: "Multiple frames",
			trace: StackTrace{
				{FunctionName: "Main"},
				{FunctionName: "Foo"},
				{FunctionName: "Bar"},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			depth := tt.trace.Depth()
			if depth != tt.expected {
				t.Errorf("Expected depth %d, got %d", tt.expected, depth)
			}
		})
	}
}

func TestNewStackFrame(t *testing.T) {
	pos := &lexer.Position{Line: 42, Column: 13}
	frame := NewStackFrame("TestFunc", "test.dws", pos)

	if frame.FunctionName != "TestFunc" {
		t.Errorf("Expected FunctionName 'TestFunc', got %q", frame.FunctionName)
	}
	if frame.FileName != "test.dws" {
		t.Errorf("Expected FileName 'test.dws', got %q", frame.FileName)
	}
	if frame.Position != pos {
		t.Errorf("Expected position %v, got %v", pos, frame.Position)
	}
}

func TestNewStackTrace(t *testing.T) {
	trace := NewStackTrace()

	if trace == nil {
		t.Error("NewStackTrace returned nil")
	}
	if len(trace) != 0 {
		t.Errorf("Expected empty stack trace, got length %d", len(trace))
	}
}

func TestStackTrace_RealWorldScenario(t *testing.T) {
	// Simulate a call stack: Main -> ProcessData -> ValidateInput
	trace := StackTrace{
		{FunctionName: "Main", FileName: "main.dws", Position: &lexer.Position{Line: 50, Column: 1}},
		{FunctionName: "ProcessData", FileName: "main.dws", Position: &lexer.Position{Line: 30, Column: 5}},
		{FunctionName: "ValidateInput", FileName: "main.dws", Position: &lexer.Position{Line: 10, Column: 3}},
	}

	// Test string representation (should show most recent first)
	expected := "ValidateInput [line: 10, column: 3]\nProcessData [line: 30, column: 5]\nMain [line: 50, column: 1]"
	result := trace.String()
	if result != expected {
		t.Errorf("Stack trace string doesn't match.\nExpected:\n%s\nGot:\n%s", expected, result)
	}

	// Test depth
	if trace.Depth() != 3 {
		t.Errorf("Expected depth 3, got %d", trace.Depth())
	}

	// Test top (most recent call)
	top := trace.Top()
	if top == nil || top.FunctionName != "ValidateInput" {
		t.Errorf("Expected top to be ValidateInput, got %v", top)
	}

	// Test bottom (original caller)
	bottom := trace.Bottom()
	if bottom == nil || bottom.FunctionName != "Main" {
		t.Errorf("Expected bottom to be Main, got %v", bottom)
	}
}

func TestStackTrace_StringFormatMatchesDWScript(t *testing.T) {
	// Test that our format matches DWScript's expected format from fixtures
	// Example from testdata/fixtures/SimpleScripts/stacktrace.txt:
	// ThisOneBombs [line: 3, column: 20]
	// CallsABomb [line: 8, column: 4]
	trace := StackTrace{
		{FunctionName: "CallsABomb", Position: &lexer.Position{Line: 8, Column: 4}},
		{FunctionName: "ThisOneBombs", Position: &lexer.Position{Line: 3, Column: 20}},
	}

	result := trace.String()
	lines := strings.Split(result, "\n")

	// Check each line matches the DWScript format
	if lines[0] != "ThisOneBombs [line: 3, column: 20]" {
		t.Errorf("First line doesn't match DWScript format: %q", lines[0])
	}
	if lines[1] != "CallsABomb [line: 8, column: 4]" {
		t.Errorf("Second line doesn't match DWScript format: %q", lines[1])
	}
}

// Helper function for tests
func stringPtr(s string) *string {
	return &s
}
