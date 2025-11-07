// Copyright (c) 2024 MeKo-Tech
// SPDX-License-Identifier: MIT

package dwscript

import (
	"fmt"
)

// ErrorSeverity represents the severity level of an error or warning.
type ErrorSeverity int

const (
	// SeverityError represents a critical error that prevents compilation or execution.
	SeverityError ErrorSeverity = iota
	// SeverityWarning represents a non-critical issue that should be addressed.
	SeverityWarning
	// SeverityInfo represents informational messages.
	SeverityInfo
	// SeverityHint represents subtle suggestions for code improvement.
	SeverityHint
)

// String returns the string representation of the severity level.
func (s ErrorSeverity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	case SeverityInfo:
		return "info"
	case SeverityHint:
		return "hint"
	default:
		return "unknown"
	}
}

// Error represents a structured error with position information and metadata.
// This type is designed for LSP integration and provides rich error information
// for IDEs and other tooling.
//
// Position fields (Line and Column) use 1-based indexing to match LSP conventions
// and standard editor behavior. For example:
//   - Line 1, Column 1 refers to the first character in the file
//   - Line 5, Column 10 refers to the 10th character on the 5th line
//
// The Length field indicates the span of the error in characters, allowing
// tools to highlight the exact portion of code that caused the error.
type Error struct {
	// Message is the human-readable error description.
	Message string

	// Line is the 1-based line number where the error occurred.
	// Line 1 is the first line of the file.
	Line int

	// Column is the 1-based column number where the error occurred.
	// Column 1 is the first character on the line.
	Column int

	// Length is the length of the error span in characters.
	// This allows tools to highlight the exact problematic code.
	// A length of 0 indicates a point error (no specific span).
	Length int

	// Severity indicates whether this is an error, warning, info, or hint.
	Severity ErrorSeverity

	// Code is an optional error code for programmatic error handling.
	// Examples: "E_UNDEFINED_VAR", "W_UNUSED_VAR", "E_TYPE_MISMATCH"
	Code string
}

// Error implements the error interface.
// It formats the error in a human-readable format suitable for console output.
func (e *Error) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("%s at %d:%d: %s [%s]",
			e.Severity, e.Line, e.Column, e.Message, e.Code)
	}
	return fmt.Sprintf("%s at %d:%d: %s",
		e.Severity, e.Line, e.Column, e.Message)
}

// NewError creates a new Error with the given parameters.
// This is a convenience constructor for creating errors programmatically.
func NewError(message string, line, column, length int, severity ErrorSeverity, code string) *Error {
	return &Error{
		Message:  message,
		Line:     line,
		Column:   column,
		Length:   length,
		Severity: severity,
		Code:     code,
	}
}

// NewErrorFromPosition creates a new error from a message and position information.
// If the length is not known, pass 0.
func NewErrorFromPosition(message string, line, column, length int) *Error {
	return &Error{
		Message:  message,
		Line:     line,
		Column:   column,
		Length:   length,
		Severity: SeverityError,
	}
}

// NewWarning creates a new warning (non-critical issue).
func NewWarning(message string, line, column, length int, code string) *Error {
	return &Error{
		Message:  message,
		Line:     line,
		Column:   column,
		Length:   length,
		Severity: SeverityWarning,
		Code:     code,
	}
}

// IsError returns true if this is an error (not a warning, info, or hint).
func (e *Error) IsError() bool {
	return e.Severity == SeverityError
}

// IsWarning returns true if this is a warning.
func (e *Error) IsWarning() bool {
	return e.Severity == SeverityWarning
}
