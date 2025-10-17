package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// TestRunWithSemanticErrors tests that semantic errors are reported before execution
func TestRunWithSemanticErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectError   bool
		errorContains string
	}{
		{
			name:          "type mismatch error",
			input:         "var x: Integer := 'hello';",
			expectError:   true,
			errorContains: "cannot assign String to Integer",
		},
		{
			name:          "undefined variable error",
			input:         "var x: Integer := y;",
			expectError:   true,
			errorContains: "undefined variable 'y'",
		},
		{
			name:        "valid program",
			input:       "var x: Integer := 5; PrintLn(x);",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up flags
			evalExpr = tt.input
			dumpAST = false
			trace = false

			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			// Run the script
			err := runScript(nil, []string{})

			// Restore stderr
			w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			buf.ReadFrom(r)
			stderr := buf.String()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				if !strings.Contains(stderr, tt.errorContains) {
					t.Errorf("expected error to contain %q, got: %s", tt.errorContains, stderr)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v, stderr: %s", err, stderr)
				}
			}
		})
	}
}

// TestRunWithSemanticAnalysisFlag tests the --type-check flag
func TestRunWithSemanticAnalysisFlag(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		enableTypeCheck bool
		expectError    bool
	}{
		{
			name:            "type error with type checking enabled",
			input:           "var x: Integer := 'hello';",
			enableTypeCheck: true,
			expectError:     true,
		},
		{
			name:            "type error with type checking disabled",
			input:           "var x: Integer := 'hello';",
			enableTypeCheck: false,
			expectError:     false, // Should pass parsing, fail at runtime
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up flags
			evalExpr = tt.input
			typeCheck = tt.enableTypeCheck
			dumpAST = false
			trace = false

			// Capture output
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			// Run the script
			err := runScript(nil, []string{})

			// Restore stderr
			w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			buf.ReadFrom(r)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			}
		})
	}
}
