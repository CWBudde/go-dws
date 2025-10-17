package interp

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/lexer"
	"github.com/cwbudde/go-dws/parser"
)

// TestErrorMessagesIncludeLocation tests that runtime errors include line/column information
func TestErrorMessagesIncludeLocation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // Expected substring in error message
	}{
		{
			name:     "undefined variable error includes location",
			input:    "var x: Integer := 5;\nvar y: Integer := z;",
			expected: "line 2",
		},
		{
			name:     "division by zero includes location",
			input:    "var x: Integer := 10;\nvar y: Integer := x / 0;",
			expected: "line 2",
		},
		{
			name:     "type mismatch in operation includes location",
			input:    "var x: Integer := 5;\nvar y: String := 'hello';\nvar z: Integer := x + y;",
			expected: "line 3",
		},
		{
			name:     "undefined function includes location",
			input:    "var x: Integer := unknownFunc(5);",
			expected: "line 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			interp := New(nil)
			result := interp.Eval(program)

			if result.Type() != "ERROR" {
				t.Fatalf("expected error, got %s: %s", result.Type(), result.String())
			}

			errorMsg := result.String()
			if !strings.Contains(errorMsg, tt.expected) {
				t.Errorf("error message does not include location info\nExpected to contain: %s\nGot: %s",
					tt.expected, errorMsg)
			}
		})
	}
}

// TestErrorMessageFormat tests the format of error messages
func TestErrorMessageFormat(t *testing.T) {
	input := "var x: Integer := y;"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	interp := New(nil)
	result := interp.Eval(program)

	if result.Type() != "ERROR" {
		t.Fatalf("expected error, got %s", result.Type())
	}

	errorMsg := result.String()

	// Check that error includes key components
	if !strings.Contains(errorMsg, "ERROR") {
		t.Error("error message should start with ERROR")
	}

	// Should include line number
	if !strings.Contains(errorMsg, "line") {
		t.Error("error message should include line number")
	}

	// Should include column number
	if !strings.Contains(errorMsg, "column") || !strings.Contains(errorMsg, "col") {
		t.Error("error message should include column information")
	}
}
