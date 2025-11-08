// Copyright (c) 2024 MeKo-Tech
// SPDX-License-Identifier: MIT

// Package dwscript provides structured error handling for DWScript compilation and execution.
//
// # Error Message Format Standards
//
// The go-dws error system follows these conventions for IDE integration and consistency:
//
// 1. **Position Information**: Always separate from message text
//   - Error messages should NOT contain position info in the text (e.g., no "at 5:10")
//   - Position is stored in Error.Line, Error.Column fields (1-indexed)
//   - Format when displayed: "severity at line:column: message [CODE]"
//   - Example: "error at 10:5: undefined variable 'x' [E_UNDEFINED_VAR]"
//
// 2. **Message Clarity**: Clear, concise, actionable
//   - Start with what went wrong (e.g., "undefined variable", "type mismatch")
//   - Include relevant context (variable names, types, etc.)
//   - Avoid redundant information
//   - Good: "undefined variable 'x'"
//   - Bad: "error: undefined variable 'x' at line 10 column 5"
//
// 3. **Severity Levels**:
//   - SeverityError: Critical errors preventing compilation/execution
//   - SeverityWarning: Non-critical issues that should be addressed
//   - SeverityInfo: Informational messages
//   - SeverityHint: Subtle suggestions for improvement
//
// 4. **Error Codes**: Optional but recommended for programmatic handling
//   - Format: E_* for errors, W_* for warnings
//   - Examples: E_UNDEFINED_VAR, E_TYPE_MISMATCH, W_UNUSED_VAR
//   - Allows IDEs to provide specific quick fixes
//
// 5. **Error Spans**: Use Length field to highlight problematic code
//   - Length indicates how many characters are affected
//   - 0 means point error (no specific span)
//   - Allows IDEs to show squiggly lines under exact problematic code
//
// # LSP Integration
//
// The Error struct is designed to map directly to LSP Diagnostic:
//   - Error.Line, Error.Column → diagnostic.range.start
//   - Error.Length → diagnostic.range.end (start + length)
//   - Error.Severity → diagnostic.severity
//   - Error.Code → diagnostic.code
//   - Error.Message → diagnostic.message
//
// # Example Usage
//
//	// Creating a structured error
//	err := dwscript.NewError(
//	    "undefined variable 'x'",
//	    10,    // line
//	    5,     // column
//	    1,     // length (1 character)
//	    dwscript.SeverityError,
//	    "E_UNDEFINED_VAR",
//	)
//
//	// Formatting for console output
//	fmt.Println(err.Error())
//	// Output: error at 10:5: undefined variable 'x' [E_UNDEFINED_VAR]
//
//	// Accessing structured fields for IDE integration
//	if err.IsError() {
//	    showDiagnostic(err.Line, err.Column, err.Length, err.Message)
//	}
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
	Message  string
	Code     string
	Line     int
	Column   int
	Length   int
	Severity ErrorSeverity
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
