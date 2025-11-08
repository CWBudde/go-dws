// Copyright (c) 2024 MeKo-Tech
// SPDX-License-Identifier: MIT

package dwscript

import (
	"testing"
)

func TestErrorSeverity_String(t *testing.T) {
	tests := []struct {
		want     string
		severity ErrorSeverity
	}{
		{"error", SeverityError},
		{"warning", SeverityWarning},
		{"info", SeverityInfo},
		{"hint", SeverityHint},
		{"unknown", ErrorSeverity(999)},
	}

	for _, tt := range tests {
		got := tt.severity.String()
		if got != tt.want {
			t.Errorf("ErrorSeverity(%d).String() = %q, want %q", tt.severity, got, tt.want)
		}
	}
}

func TestError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "error with code",
			err: &Error{
				Message:  "undefined variable 'x'",
				Line:     10,
				Column:   5,
				Length:   1,
				Severity: SeverityError,
				Code:     "E_UNDEFINED_VAR",
			},
			want: "error at 10:5: undefined variable 'x' [E_UNDEFINED_VAR]",
		},
		{
			name: "error without code",
			err: &Error{
				Message:  "unexpected token",
				Line:     1,
				Column:   1,
				Length:   5,
				Severity: SeverityError,
			},
			want: "error at 1:1: unexpected token",
		},
		{
			name: "warning with code",
			err: &Error{
				Message:  "unused variable 'y'",
				Line:     20,
				Column:   8,
				Length:   1,
				Severity: SeverityWarning,
				Code:     "W_UNUSED_VAR",
			},
			want: "warning at 20:8: unused variable 'y' [W_UNUSED_VAR]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewError(t *testing.T) {
	err := NewError("test message", 5, 10, 3, SeverityError, "E_TEST")

	if err.Message != "test message" {
		t.Errorf("Message = %q, want %q", err.Message, "test message")
	}
	if err.Line != 5 {
		t.Errorf("Line = %d, want %d", err.Line, 5)
	}
	if err.Column != 10 {
		t.Errorf("Column = %d, want %d", err.Column, 10)
	}
	if err.Length != 3 {
		t.Errorf("Length = %d, want %d", err.Length, 3)
	}
	if err.Severity != SeverityError {
		t.Errorf("Severity = %v, want %v", err.Severity, SeverityError)
	}
	if err.Code != "E_TEST" {
		t.Errorf("Code = %q, want %q", err.Code, "E_TEST")
	}
}

func TestNewErrorFromPosition(t *testing.T) {
	err := NewErrorFromPosition("test message", 1, 2, 3)

	if err.Message != "test message" {
		t.Errorf("Message = %q, want %q", err.Message, "test message")
	}
	if err.Line != 1 {
		t.Errorf("Line = %d, want %d", err.Line, 1)
	}
	if err.Column != 2 {
		t.Errorf("Column = %d, want %d", err.Column, 2)
	}
	if err.Length != 3 {
		t.Errorf("Length = %d, want %d", err.Length, 3)
	}
	if err.Severity != SeverityError {
		t.Errorf("Severity = %v, want %v", err.Severity, SeverityError)
	}
}

func TestNewWarning(t *testing.T) {
	warn := NewWarning("test warning", 15, 20, 5, "W_TEST")

	if warn.Message != "test warning" {
		t.Errorf("Message = %q, want %q", warn.Message, "test warning")
	}
	if warn.Line != 15 {
		t.Errorf("Line = %d, want %d", warn.Line, 15)
	}
	if warn.Column != 20 {
		t.Errorf("Column = %d, want %d", warn.Column, 20)
	}
	if warn.Length != 5 {
		t.Errorf("Length = %d, want %d", warn.Length, 5)
	}
	if warn.Severity != SeverityWarning {
		t.Errorf("Severity = %v, want %v", warn.Severity, SeverityWarning)
	}
	if warn.Code != "W_TEST" {
		t.Errorf("Code = %q, want %q", warn.Code, "W_TEST")
	}
}

func TestError_IsError(t *testing.T) {
	tests := []struct {
		name     string
		severity ErrorSeverity
		want     bool
	}{
		{"error", SeverityError, true},
		{"warning", SeverityWarning, false},
		{"info", SeverityInfo, false},
		{"hint", SeverityHint, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{Severity: tt.severity}
			got := err.IsError()
			if got != tt.want {
				t.Errorf("IsError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestError_IsWarning(t *testing.T) {
	tests := []struct {
		name     string
		severity ErrorSeverity
		want     bool
	}{
		{"error", SeverityError, false},
		{"warning", SeverityWarning, true},
		{"info", SeverityInfo, false},
		{"hint", SeverityHint, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{Severity: tt.severity}
			got := err.IsWarning()
			if got != tt.want {
				t.Errorf("IsWarning() = %v, want %v", got, tt.want)
			}
		})
	}
}
