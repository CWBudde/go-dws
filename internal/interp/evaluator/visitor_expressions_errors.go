package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/pkg/ast"
)

// This file contains error handling utilities for the evaluator visitor methods.

// ErrorValue represents a runtime error.
type ErrorValue struct {
	Message string
}

func (e *ErrorValue) Type() string   { return "ERROR" }
func (e *ErrorValue) String() string { return "ERROR: " + e.Message }

// newError creates a new error value with optional formatting and location information.
func (e *Evaluator) newError(node ast.Node, format string, args ...any) Value {
	message := fmt.Sprintf(format, args...)

	// Add location information if node is available
	if node != nil {
		pos := node.Pos()
		if pos.Line > 0 {
			message = fmt.Sprintf("%s at line %d, column: %d", message, pos.Line, pos.Column)
		}
	}

	return &ErrorValue{Message: message}
}

// isError checks if a value is an error.
func isError(val Value) bool {
	if val != nil {
		return val.Type() == "ERROR"
	}
	return false
}
