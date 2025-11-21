package errors

import (
	"fmt"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// ErrorCategory represents the category of an interpreter error.
type ErrorCategory string

const (
	// CategoryType represents type-related errors (type mismatches, invalid operations).
	CategoryType ErrorCategory = "Type"
	// CategoryRuntime represents runtime errors (division by zero, index out of bounds).
	CategoryRuntime ErrorCategory = "Runtime"
	// CategoryUndefined represents errors for undefined entities (variables, functions, types).
	CategoryUndefined ErrorCategory = "Undefined"
	// CategoryContract represents contract violation errors (precondition, postcondition).
	CategoryContract ErrorCategory = "Contract"
	// CategoryInternal represents internal interpreter errors (should never happen).
	CategoryInternal ErrorCategory = "Internal"
)

// InterpreterError represents a runtime error in the interpreter with rich context.
type InterpreterError struct {
	Err        error
	Pos        *token.Position
	Values     map[string]string
	Category   ErrorCategory
	Message    string
	Expression string
}

// Error implements the error interface.
func (e *InterpreterError) Error() string {
	if e.Pos != nil {
		return fmt.Sprintf("%s error at line %d, column %d: %s", e.Category, e.Pos.Line, e.Pos.Column, e.Message)
	}
	return fmt.Sprintf("%s error: %s", e.Category, e.Message)
}

// Unwrap implements error unwrapping for error chains.
func (e *InterpreterError) Unwrap() error {
	return e.Err
}

// NewTypeError creates a type-related error.
func NewTypeError(pos *token.Position, message string, expr string) *InterpreterError {
	return &InterpreterError{
		Category:   CategoryType,
		Message:    message,
		Pos:        pos,
		Expression: expr,
	}
}

// NewTypeErrorf creates a type-related error with formatting.
func NewTypeErrorf(pos *token.Position, expr string, format string, args ...interface{}) *InterpreterError {
	return &InterpreterError{
		Category:   CategoryType,
		Message:    fmt.Sprintf(format, args...),
		Pos:        pos,
		Expression: expr,
	}
}

// NewRuntimeError creates a runtime error.
func NewRuntimeError(pos *token.Position, message string, expr string) *InterpreterError {
	return &InterpreterError{
		Category:   CategoryRuntime,
		Message:    message,
		Pos:        pos,
		Expression: expr,
	}
}

// NewRuntimeErrorf creates a runtime error with formatting.
func NewRuntimeErrorf(pos *token.Position, expr string, format string, args ...interface{}) *InterpreterError {
	return &InterpreterError{
		Category:   CategoryRuntime,
		Message:    fmt.Sprintf(format, args...),
		Pos:        pos,
		Expression: expr,
	}
}

// NewRuntimeErrorWithValues creates a runtime error with runtime values for debugging.
func NewRuntimeErrorWithValues(pos *token.Position, expr string, message string, values map[string]string) *InterpreterError {
	return &InterpreterError{
		Category:   CategoryRuntime,
		Message:    message,
		Pos:        pos,
		Expression: expr,
		Values:     values,
	}
}

// NewUndefinedError creates an undefined entity error.
func NewUndefinedError(pos *token.Position, message string, expr string) *InterpreterError {
	return &InterpreterError{
		Category:   CategoryUndefined,
		Message:    message,
		Pos:        pos,
		Expression: expr,
	}
}

// NewUndefinedErrorf creates an undefined entity error with formatting.
func NewUndefinedErrorf(pos *token.Position, expr string, format string, args ...interface{}) *InterpreterError {
	return &InterpreterError{
		Category:   CategoryUndefined,
		Message:    fmt.Sprintf(format, args...),
		Pos:        pos,
		Expression: expr,
	}
}

// NewContractError creates a contract violation error.
func NewContractError(pos *token.Position, message string, expr string) *InterpreterError {
	return &InterpreterError{
		Category:   CategoryContract,
		Message:    message,
		Pos:        pos,
		Expression: expr,
	}
}

// NewContractErrorf creates a contract violation error with formatting.
func NewContractErrorf(pos *token.Position, expr string, format string, args ...interface{}) *InterpreterError {
	return &InterpreterError{
		Category:   CategoryContract,
		Message:    fmt.Sprintf(format, args...),
		Pos:        pos,
		Expression: expr,
	}
}

// NewInternalError creates an internal interpreter error.
func NewInternalError(pos *token.Position, message string, expr string) *InterpreterError {
	return &InterpreterError{
		Category:   CategoryInternal,
		Message:    message,
		Pos:        pos,
		Expression: expr,
	}
}

// NewInternalErrorf creates an internal interpreter error with formatting.
func NewInternalErrorf(pos *token.Position, expr string, format string, args ...interface{}) *InterpreterError {
	return &InterpreterError{
		Category:   CategoryInternal,
		Message:    fmt.Sprintf(format, args...),
		Pos:        pos,
		Expression: expr,
	}
}

// WrapError wraps an existing error with interpreter context.
func WrapError(err error, category ErrorCategory, pos *token.Position, expr string) *InterpreterError {
	return &InterpreterError{
		Category:   category,
		Message:    err.Error(),
		Pos:        pos,
		Expression: expr,
		Err:        err,
	}
}

// WrapErrorf wraps an existing error with additional message formatting.
func WrapErrorf(err error, category ErrorCategory, pos *token.Position, expr string, format string, args ...interface{}) *InterpreterError {
	return &InterpreterError{
		Category:   category,
		Message:    fmt.Sprintf(format, args...),
		Pos:        pos,
		Expression: expr,
		Err:        err,
	}
}

// PositionFromNode extracts position from an AST node.
func PositionFromNode(node ast.Node) *token.Position {
	if node == nil {
		return nil
	}

	// All AST nodes implement the Pos() method, so we can just use that
	pos := node.Pos()
	return &pos
}

// ExpressionFromNode returns a string representation of an AST node.
func ExpressionFromNode(node ast.Node) string {
	if node == nil {
		return ""
	}
	return node.String()
}
