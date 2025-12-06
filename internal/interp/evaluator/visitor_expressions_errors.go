package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// This file contains error handling utilities for the evaluator visitor methods.

// newError creates a new error value with optional formatting and location information.
func (e *Evaluator) newError(node ast.Node, format string, args ...any) Value {
	message := fmt.Sprintf(format, args...)

	// Add location information if node is available
	// Format matches RuntimeError.String() format: "[line: N, column: M]"
	if node != nil {
		pos := node.Pos()
		if pos.Line > 0 {
			message = fmt.Sprintf("%s [line: %d, column: %d]", message, pos.Line, pos.Column)
		}
	}

	return &runtime.ErrorValue{Message: message}
}

// isError checks if a value is an error.
func isError(val Value) bool {
	if val != nil {
		return val.Type() == "ERROR"
	}
	return false
}
