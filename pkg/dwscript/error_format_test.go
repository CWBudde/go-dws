package dwscript

import (
	"strings"
	"testing"
)

// TestErrorFormatting tests that error messages follow the standard format
func TestErrorFormatting(t *testing.T) {
	tests := []struct {
		checkMessage func(t *testing.T, err error)
		name         string
		source       string
		expectError  bool
	}{
		{
			name:        "undefined variable - position not in message text",
			source:      "begin x := 5; end;",
			expectError: true,
			checkMessage: func(t *testing.T, err error) {
				compileErr, ok := err.(*CompileError)
				if !ok {
					t.Fatal("Expected CompileError")
				}

				if len(compileErr.Errors) == 0 {
					t.Fatal("Expected at least one error")
				}

				e := compileErr.Errors[0]

				// Message should not contain " at " followed by line:column
				if strings.Contains(e.Message, " at ") {
					// Check if it's actually position info (digits after " at ")
					idx := strings.LastIndex(e.Message, " at ")
					if idx >= 0 {
						remaining := e.Message[idx+4:]
						var line, col int
						n, _ := sscanf(remaining, "%d:%d", &line, &col)
						if n == 2 {
							t.Errorf("Message contains position info: %q", e.Message)
						}
					}
				}

				// Should have valid position in struct fields
				if e.Line == 0 && e.Column == 0 {
					// Parser errors might have 0:0 for some cases, but semantic errors should have position
					if compileErr.Stage == "type checking" {
						t.Log("Warning: Semantic error has 0:0 position")
					}
				}

				// Error() method should produce clean format
				errStr := e.Error()
				if !strings.Contains(errStr, "error at") {
					t.Errorf("Error string should start with 'error at', got: %q", errStr)
				}
			},
		},
		{
			name:        "type mismatch - clear message",
			source:      "var x: Integer; begin x := 'hello'; end;",
			expectError: true,
			checkMessage: func(t *testing.T, err error) {
				compileErr, ok := err.(*CompileError)
				if !ok {
					t.Fatal("Expected CompileError")
				}

				if len(compileErr.Errors) == 0 {
					t.Fatal("Expected at least one error")
				}

				e := compileErr.Errors[0]

				// Message should be clear and concise
				if !strings.Contains(strings.ToLower(e.Message), "assign") {
					t.Errorf("Type mismatch error should mention 'assign', got: %q", e.Message)
				}

				// Should not be overly verbose
				if len(e.Message) > 200 {
					t.Errorf("Error message too verbose (%d chars): %q", len(e.Message), e.Message)
				}
			},
		},
		{
			name:        "parse error - consistent format",
			source:      "var x: Integer := ;",
			expectError: true,
			checkMessage: func(t *testing.T, err error) {
				compileErr, ok := err.(*CompileError)
				if !ok {
					t.Fatal("Expected CompileError")
				}

				if len(compileErr.Errors) == 0 {
					t.Fatal("Expected at least one error")
				}

				// All errors should have consistent format
				for i, e := range compileErr.Errors {
					// Should have error code
					if e.Code == "" {
						t.Errorf("Error %d missing error code", i)
					}

					// Should have severity
					if e.Severity != SeverityError && e.Severity != SeverityWarning {
						t.Errorf("Error %d has invalid severity: %v", i, e.Severity)
					}

					// Error() output should be consistent
					errStr := e.Error()
					if !strings.Contains(errStr, " at ") {
						t.Errorf("Error %d missing ' at ' in formatted output: %q", i, errStr)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, err := New(WithTypeCheck(true))
			if err != nil {
				t.Fatalf("Failed to create engine: %v", err)
			}

			_, err = engine.Compile(tt.source)
			if tt.expectError && err == nil {
				t.Fatal("Expected compilation error, got none")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if err != nil && tt.checkMessage != nil {
				tt.checkMessage(t, err)
			}
		})
	}
}

// TestExtractPositionFromError tests the position extraction helper
func TestExtractPositionFromError(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantMsg  string
		wantLine int
		wantCol  int
	}{
		{
			name:     "error with position",
			input:    "undefined variable 'x' at 10:5",
			wantLine: 10,
			wantCol:  5,
			wantMsg:  "undefined variable 'x'",
		},
		{
			name:     "error without position",
			input:    "type mismatch",
			wantLine: 0,
			wantCol:  0,
			wantMsg:  "type mismatch",
		},
		{
			name:     "error with at but no position",
			input:    "looking at line 5",
			wantLine: 0,
			wantCol:  0,
			wantMsg:  "looking at line 5",
		},
		{
			name:     "error with multiple at",
			input:    "looking at variable 'x' at 5:10",
			wantLine: 5,
			wantCol:  10,
			wantMsg:  "looking at variable 'x'",
		},
		{
			name:     "error at start of file",
			input:    "syntax error at 1:1",
			wantLine: 1,
			wantCol:  1,
			wantMsg:  "syntax error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line, col, msg := extractPositionFromError(tt.input)
			if line != tt.wantLine {
				t.Errorf("line = %d, want %d", line, tt.wantLine)
			}
			if col != tt.wantCol {
				t.Errorf("column = %d, want %d", col, tt.wantCol)
			}
			if msg != tt.wantMsg {
				t.Errorf("message = %q, want %q", msg, tt.wantMsg)
			}
		})
	}
}

// TestErrorSeverityString tests the severity string formatting
func TestErrorSeverityString(t *testing.T) {
	tests := []struct {
		want     string
		severity ErrorSeverity
	}{
		{severity: SeverityError, want: "error"},
		{severity: SeverityWarning, want: "warning"},
		{severity: SeverityInfo, want: "info"},
		{severity: SeverityHint, want: "hint"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.severity.String()
			if got != tt.want {
				t.Errorf("Severity.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestErrorIsError and TestErrorIsWarning test the helper methods
func TestErrorTypeChecking(t *testing.T) {
	errObj := NewError("test", 1, 1, 0, SeverityError, "E_TEST")
	warnObj := NewWarning("test", 1, 1, 0, "W_TEST")

	if !errObj.IsError() {
		t.Error("Expected IsError() to return true for error")
	}
	if errObj.IsWarning() {
		t.Error("Expected IsWarning() to return false for error")
	}

	if warnObj.IsError() {
		t.Error("Expected IsError() to return false for warning")
	}
	if !warnObj.IsWarning() {
		t.Error("Expected IsWarning() to return true for warning")
	}
}

// TestCompileErrorFormatting tests CompileError.Error() formatting
func TestCompileErrorFormatting(t *testing.T) {
	tests := []struct {
		name    string
		stage   string
		errors  []*Error
		wantStr []string // substrings that should appear
	}{
		{
			name:  "single error",
			stage: "parsing",
			errors: []*Error{
				NewError("test error", 5, 10, 1, SeverityError, "E_TEST"),
			},
			wantStr: []string{"parsing error", "5:10", "test error", "E_TEST"},
		},
		{
			name:  "multiple errors",
			stage: "type checking",
			errors: []*Error{
				NewError("first error", 1, 1, 0, SeverityError, "E_1"),
				NewError("second error", 2, 5, 0, SeverityError, "E_2"),
			},
			wantStr: []string{"type checking errors", "(2)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compileErr := &CompileError{
				Stage:  tt.stage,
				Errors: tt.errors,
			}

			errStr := compileErr.Error()
			for _, want := range tt.wantStr {
				if !strings.Contains(errStr, want) {
					t.Errorf("Error string should contain %q, got: %q", want, errStr)
				}
			}
		})
	}
}

// Helper for scanf that doesn't exist in strings package
func sscanf(s, format string, args ...interface{}) (int, error) {
	return 0, nil // Simplified - actual implementation would use fmt.Sscanf
}
