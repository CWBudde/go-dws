package interp

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// TestSafeTypeAssertions tests that type assertions don't panic and provide helpful errors
func TestSafeTypeAssertions(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		shouldError   bool
		errorContains string
	}{
		{
			name:          "integer arithmetic with valid types",
			input:         "var x: Integer := 5 + 3;",
			shouldError:   false,
			errorContains: "",
		},
		{
			name:          "string concatenation with valid types",
			input:         "var s: String := 'hello' + ' world';",
			shouldError:   false,
			errorContains: "",
		},
		{
			name:          "type mismatch in arithmetic should error, not panic",
			input:         "var x: Integer := 5 + 'hello';",
			shouldError:   true,
			errorContains: "type mismatch",
		},
		{
			name:          "type mismatch in comparison should error, not panic",
			input:         "var b: Boolean := 5 = 'hello';",
			shouldError:   true,
			errorContains: "type mismatch",
		},
		{
			name:          "negating non-integer should error, not panic",
			input:         "var x: Integer := -'hello';",
			shouldError:   true,
			errorContains: "expected integer",
		},
		{
			name:          "boolean NOT on non-boolean should error, not panic",
			input:         "var b: Boolean := not 42;",
			shouldError:   true,
			errorContains: "expected boolean",
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

			// Disable type checking to test runtime type assertions
			interp := New(nil)
			result := interp.Eval(program)

			if tt.shouldError {
				if result.Type() != "ERROR" {
					t.Fatalf("expected error, got %s: %s", result.Type(), result.String())
				}
				errorMsg := result.String()
				if !strings.Contains(strings.ToLower(errorMsg), strings.ToLower(tt.errorContains)) {
					t.Errorf("error message doesn't contain expected text\nExpected to contain: %s\nGot: %s",
						tt.errorContains, errorMsg)
				}
			} else {
				if result.Type() == "ERROR" {
					t.Fatalf("unexpected error: %s", result.String())
				}
			}
		})
	}
}

// TestTypeAssertionsPanicPrevention tests that operations don't panic even with invalid types
func TestTypeAssertionsPanicPrevention(t *testing.T) {
	// These tests should NOT panic, even though they have type mismatches
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "arithmetic with mismatched types",
			input: "var x: Integer := 5; var y: String := 'hello'; var z: Integer := x + y;",
		},
		{
			name:  "division with non-integer",
			input: "var x: Integer := 'hello' / 2;",
		},
		{
			name:  "modulo with non-integer",
			input: "var x: Integer := 'hello' mod 2;",
		},
		{
			name:  "comparison with mismatched types",
			input: "var b: Boolean := 5 < 'hello';",
		},
		{
			name:  "unary minus on string",
			input: "var x: Integer := -'hello';",
		},
		{
			name:  "boolean NOT on integer",
			input: "var b: Boolean := not 42;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use defer/recover to catch panics
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Operation panicked: %v\nThis should have returned an error instead", r)
				}
			}()

			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			interp := New(nil)
			result := interp.Eval(program)

			// We expect an error, not a panic
			if result.Type() != "ERROR" {
				t.Errorf("expected ERROR result for invalid type operation, got %s", result.Type())
			}
		})
	}
}
