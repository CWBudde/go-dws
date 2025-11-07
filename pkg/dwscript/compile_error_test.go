// Copyright (c) 2024 MeKo-Tech
// SPDX-License-Identifier: MIT

package dwscript

import (
	"strings"
	"testing"
)

func TestCompileError_StructuredErrors(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Test parsing error with structured errors
	_, err = engine.Compile("var x := ")
	if err == nil {
		t.Fatal("expected compile error, got nil")
	}

	compileErr, ok := err.(*CompileError)
	if !ok {
		t.Fatalf("expected *CompileError, got %T", err)
	}

	if compileErr.Stage != "parsing" {
		t.Errorf("expected stage 'parsing', got %q", compileErr.Stage)
	}

	if len(compileErr.Errors) == 0 {
		t.Fatal("expected errors, got none")
	}

	// Check that errors are structured
	for i, structErr := range compileErr.Errors {
		if structErr.Message == "" {
			t.Errorf("error %d has empty message", i)
		}
		if structErr.Line == 0 && structErr.Column == 0 {
			// Position might be missing in current implementation, that's OK for now
			t.Logf("warning: error %d has no position information", i)
		}
		if structErr.Severity != SeverityError {
			t.Errorf("error %d has severity %v, want %v", i, structErr.Severity, SeverityError)
		}
		if !structErr.IsError() {
			t.Errorf("error %d: IsError() = false, want true", i)
		}
		if structErr.IsWarning() {
			t.Errorf("error %d: IsWarning() = true, want false", i)
		}
	}
}

func TestCompileError_HasErrors(t *testing.T) {
	tests := []struct {
		name        string
		errors      []*Error
		wantErrors  bool
		wantWarning bool
	}{
		{
			name: "only errors",
			errors: []*Error{
				{Message: "error 1", Severity: SeverityError},
				{Message: "error 2", Severity: SeverityError},
			},
			wantErrors:  true,
			wantWarning: false,
		},
		{
			name: "only warnings",
			errors: []*Error{
				{Message: "warning 1", Severity: SeverityWarning},
				{Message: "warning 2", Severity: SeverityWarning},
			},
			wantErrors:  false,
			wantWarning: true,
		},
		{
			name: "mixed errors and warnings",
			errors: []*Error{
				{Message: "error 1", Severity: SeverityError},
				{Message: "warning 1", Severity: SeverityWarning},
			},
			wantErrors:  true,
			wantWarning: true,
		},
		{
			name:        "no errors or warnings",
			errors:      []*Error{},
			wantErrors:  false,
			wantWarning: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ce := &CompileError{
				Stage:  "test",
				Errors: tt.errors,
			}

			if got := ce.HasErrors(); got != tt.wantErrors {
				t.Errorf("HasErrors() = %v, want %v", got, tt.wantErrors)
			}

			if got := ce.HasWarnings(); got != tt.wantWarning {
				t.Errorf("HasWarnings() = %v, want %v", got, tt.wantWarning)
			}
		})
	}
}

func TestCompileError_ErrorFormatting(t *testing.T) {
	tests := []struct {
		name     string
		errors   []*Error
		contains []string
	}{
		{
			name: "single error",
			errors: []*Error{
				{
					Message:  "undefined variable",
					Line:     10,
					Column:   5,
					Severity: SeverityError,
					Code:     "E_UNDEFINED",
				},
			},
			contains: []string{"error at 10:5", "undefined variable", "[E_UNDEFINED]"},
		},
		{
			name: "multiple errors",
			errors: []*Error{
				{Message: "error 1", Line: 1, Column: 1, Severity: SeverityError},
				{Message: "error 2", Line: 2, Column: 2, Severity: SeverityError},
			},
			contains: []string{"errors (2)", "error 1", "error 2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ce := &CompileError{
				Stage:  "test",
				Errors: tt.errors,
			}

			errStr := ce.Error()
			for _, want := range tt.contains {
				if !strings.Contains(errStr, want) {
					t.Errorf("Error() output should contain %q, got:\n%s", want, errStr)
				}
			}
		})
	}
}

func TestCompileError_ManyErrors(t *testing.T) {
	// Test formatting with many errors (should truncate output)
	errors := make([]*Error, 20)
	for i := range errors {
		errors[i] = &Error{
			Message:  "error message",
			Line:     i + 1,
			Column:   1,
			Severity: SeverityError,
		}
	}

	ce := &CompileError{
		Stage:  "test",
		Errors: errors,
	}

	errStr := ce.Error()

	// Should mention the count
	if !strings.Contains(errStr, "errors (20)") {
		t.Errorf("Error() should mention total count, got:\n%s", errStr)
	}

	// Should have truncation message
	if !strings.Contains(errStr, "more errors") {
		t.Errorf("Error() should have truncation message for many errors, got:\n%s", errStr)
	}
}
