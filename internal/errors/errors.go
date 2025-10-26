// Package errors provides error formatting utilities for the DWScript compiler.
// It formats compiler errors with source context, line/column information,
// and visual indicators (carets) pointing to the error location.
package errors

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// CompilerError represents a single compilation error with position and context.
type CompilerError struct {
	Message string
	Source  string
	File    string
	Pos     lexer.Position
}

// NewCompilerError creates a new compiler error.
func NewCompilerError(pos lexer.Position, message, source, file string) *CompilerError {
	return &CompilerError{
		Pos:     pos,
		Message: message,
		Source:  source,
		File:    file,
	}
}

// Error implements the error interface.
func (e *CompilerError) Error() string {
	return e.Format(false)
}

// Format formats the error message with source context.
// If color is true, ANSI color codes are used for terminal output.
func (e *CompilerError) Format(color bool) string {
	var sb strings.Builder

	// File and position header
	if e.File != "" {
		sb.WriteString(fmt.Sprintf("Error in %s:%d:%d\n", e.File, e.Pos.Line, e.Pos.Column))
	} else {
		sb.WriteString(fmt.Sprintf("Error at line %d:%d\n", e.Pos.Line, e.Pos.Column))
	}

	// Extract the relevant source line
	sourceLine := e.getSourceLine(e.Pos.Line)
	if sourceLine != "" {
		// Line number and source
		lineNumStr := fmt.Sprintf("%4d | ", e.Pos.Line)
		sb.WriteString(lineNumStr)
		sb.WriteString(sourceLine)
		sb.WriteString("\n")

		// Caret indicator
		sb.WriteString(strings.Repeat(" ", len(lineNumStr)+e.Pos.Column-1))
		if color {
			sb.WriteString("\033[1;31m") // Red bold
		}
		sb.WriteString("^")
		if color {
			sb.WriteString("\033[0m") // Reset
		}
		sb.WriteString("\n")
	}

	// Error message
	if color {
		sb.WriteString("\033[1m") // Bold
	}
	sb.WriteString(e.Message)
	if color {
		sb.WriteString("\033[0m") // Reset
	}

	return sb.String()
}

// getSourceLine extracts a specific line from the source code.
// Lines are 1-indexed.
func (e *CompilerError) getSourceLine(lineNum int) string {
	if e.Source == "" {
		return ""
	}

	lines := strings.Split(e.Source, "\n")
	if lineNum < 1 || lineNum > len(lines) {
		return ""
	}

	return lines[lineNum-1]
}

// getSourceContext extracts multiple lines around the error for context.
// Returns lines from (lineNum - contextBefore) to (lineNum + contextAfter).
func (e *CompilerError) getSourceContext(lineNum, contextBefore, contextAfter int) []string {
	if e.Source == "" {
		return nil
	}

	lines := strings.Split(e.Source, "\n")
	if lineNum < 1 || lineNum > len(lines) {
		return nil
	}

	start := lineNum - contextBefore
	if start < 1 {
		start = 1
	}

	end := lineNum + contextAfter
	if end > len(lines) {
		end = len(lines)
	}

	return lines[start-1 : end]
}

// FormatWithContext formats the error with surrounding source context.
func (e *CompilerError) FormatWithContext(contextLines int, color bool) string {
	var sb strings.Builder

	// File and position header
	if e.File != "" {
		sb.WriteString(fmt.Sprintf("Error in %s:%d:%d\n", e.File, e.Pos.Line, e.Pos.Column))
	} else {
		sb.WriteString(fmt.Sprintf("Error at line %d:%d\n", e.Pos.Line, e.Pos.Column))
	}

	// Get context lines
	contextLinesList := e.getSourceContext(e.Pos.Line, contextLines, contextLines)
	if len(contextLinesList) == 0 {
		// Fallback to single line
		return e.Format(color)
	}

	// Calculate starting line number
	startLine := e.Pos.Line - contextLines
	if startLine < 1 {
		startLine = 1
	}

	// Display context
	for i, line := range contextLinesList {
		currentLine := startLine + i
		lineNumStr := fmt.Sprintf("%4d | ", currentLine)

		// Highlight the error line
		if currentLine == e.Pos.Line {
			if color {
				sb.WriteString("\033[1m") // Bold
			}
			sb.WriteString(lineNumStr)
			sb.WriteString(line)
			if color {
				sb.WriteString("\033[0m") // Reset
			}
			sb.WriteString("\n")

			// Caret indicator
			sb.WriteString(strings.Repeat(" ", len(lineNumStr)+e.Pos.Column-1))
			if color {
				sb.WriteString("\033[1;31m") // Red bold
			}
			sb.WriteString("^")
			if color {
				sb.WriteString("\033[0m") // Reset
			}
			sb.WriteString("\n")
		} else {
			// Context lines (dimmed if color enabled)
			if color {
				sb.WriteString("\033[2m") // Dim
			}
			sb.WriteString(lineNumStr)
			sb.WriteString(line)
			if color {
				sb.WriteString("\033[0m") // Reset
			}
			sb.WriteString("\n")
		}
	}

	// Error message
	sb.WriteString("\n")
	if color {
		sb.WriteString("\033[1m") // Bold
	}
	sb.WriteString(e.Message)
	if color {
		sb.WriteString("\033[0m") // Reset
	}

	return sb.String()
}

// FormatErrors formats multiple compiler errors.
// Each error is formatted individually with source context.
func FormatErrors(errors []*CompilerError, color bool) string {
	if len(errors) == 0 {
		return ""
	}

	if len(errors) == 1 {
		return errors[0].Format(color)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Compilation failed with %d error(s):\n\n", len(errors)))

	for i, err := range errors {
		sb.WriteString(fmt.Sprintf("[Error %d of %d]\n", i+1, len(errors)))
		sb.WriteString(err.Format(color))
		if i < len(errors)-1 {
			sb.WriteString("\n\n")
		}
	}

	return sb.String()
}

// FormatErrorsWithContext formats multiple compiler errors with source context.
func FormatErrorsWithContext(errors []*CompilerError, contextLines int, color bool) string {
	if len(errors) == 0 {
		return ""
	}

	if len(errors) == 1 {
		return errors[0].FormatWithContext(contextLines, color)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Compilation failed with %d error(s):\n\n", len(errors)))

	for i, err := range errors {
		sb.WriteString(fmt.Sprintf("[Error %d of %d]\n", i+1, len(errors)))
		sb.WriteString(err.FormatWithContext(contextLines, color))
		if i < len(errors)-1 {
			sb.WriteString("\n\n")
		}
	}

	return sb.String()
}

// FromStringErrors converts string error messages to CompilerErrors.
// This is a helper for backward compatibility with existing error reporting.
// Position information must be extracted from the error string (format: "message at line:column").
func FromStringErrors(stringErrors []string, source, file string) []*CompilerError {
	errors := make([]*CompilerError, 0, len(stringErrors))

	for _, errStr := range stringErrors {
		// Try to extract position from error string
		pos, message := parseErrorString(errStr)
		errors = append(errors, NewCompilerError(pos, message, source, file))
	}

	return errors
}

// parseErrorString attempts to extract position information from an error string.
// Expected format: "...at LINE:COLUMN" or "message"
func parseErrorString(errStr string) (lexer.Position, string) {
	// Look for " at LINE:COLUMN" pattern
	atIndex := strings.LastIndex(errStr, " at ")
	if atIndex == -1 {
		// No position information found
		return lexer.Position{Line: 0, Column: 0}, errStr
	}

	// Extract position part
	posStr := errStr[atIndex+4:] // Skip " at "
	message := strings.TrimSpace(errStr[:atIndex])

	// Parse LINE:COLUMN
	var line, column int
	_, err := fmt.Sscanf(posStr, "%d:%d", &line, &column)
	if err != nil {
		// Failed to parse, return as-is
		return lexer.Position{Line: 0, Column: 0}, errStr
	}

	return lexer.Position{Line: line, Column: column}, message
}
