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

// newError creates a new error value with optional formatting.
func (e *Evaluator) newError(_ ast.Node, format string, args ...interface{}) Value {
	return &ErrorValue{Message: fmt.Sprintf(format, args...)}
}

// isError checks if a value is an error.
func isError(val Value) bool {
	if val != nil {
		return val.Type() == "ERROR"
	}
	return false
}
