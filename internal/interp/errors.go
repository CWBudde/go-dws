package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/errors"
	interpErrors "github.com/cwbudde/go-dws/internal/interp/errors"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// ErrorValue represents a runtime error that implements the Value interface.
// It wraps an InterpreterError for better error handling and categorization.
type ErrorValue struct {
	Err     *interpErrors.InterpreterError
	Message string
}

func (e *ErrorValue) Type() string   { return "ERROR" }
func (e *ErrorValue) String() string { return "ERROR: " + e.Message }

// newError creates a new ErrorValue with an Internal error.
// Deprecated: Use newTypeError, newRuntimeError, newUndefinedError, or newInternalError instead.
func newError(format string, args ...interface{}) *ErrorValue {
	msg := fmt.Sprintf(format, args...)
	return &ErrorValue{
		Message: msg,
		Err:     interpErrors.NewInternalError(nil, msg, ""),
	}
}

// newErrorWithLocation creates a new ErrorValue with location information from a node.
func (i *Interpreter) newErrorWithLocation(node ast.Node, format string, args ...interface{}) *ErrorValue {
	message := fmt.Sprintf(format, args...)

	// Try to get location information from the node's token
	if node != nil {
		tokenLiteral := node.TokenLiteral()
		if tokenLiteral != "" {
			// Extract token information - we need to get the actual token from the node
			location := i.getLocationFromNode(node)
			if location != "" {
				message = fmt.Sprintf("%s %s", message, location)
			}
		}
	}

	return &ErrorValue{Message: message}
}

// getLocationFromNode extracts location information from an AST node
func (i *Interpreter) getLocationFromNode(node ast.Node) string {
	if node == nil {
		return ""
	}
	pos := node.Pos()
	return fmt.Sprintf("[line: %d, column: %d]", pos.Line, pos.Column)
}

// ContractFailureError represents a contract violation (precondition or postcondition failure).
type ContractFailureError struct {
	FunctionName  string
	ConditionType string // "Pre-condition" or "Post-condition"
	ConditionExpr string // The condition expression that failed
	CustomMessage string // Optional custom error message
	Line          int
	Column        int
}

func (e *ContractFailureError) Type() string { return "ERROR" }

func (e *ContractFailureError) String() string {
	location := fmt.Sprintf("[%d:%d]", e.Line, e.Column)
	message := e.CustomMessage
	if message == "" {
		message = e.ConditionExpr
	}
	return fmt.Sprintf("%s failed in %s %s: %s", e.ConditionType, e.FunctionName, location, message)
}

// newContractError creates a new ContractFailureError.
func newContractError(funcName, condType string, condition *ast.Condition) *ContractFailureError {
	var message string
	if condition.Message != nil {
		// Message will be evaluated at runtime and passed in
		message = ""
	}

	return &ContractFailureError{
		FunctionName:  funcName,
		ConditionType: condType,
		ConditionExpr: condition.Test.String(),
		CustomMessage: message,
		Line:          condition.Token.Pos.Line,
		Column:        condition.Token.Pos.Column,
	}
}

// isError checks if a value is an error.
func isError(val Value) bool {
	if val != nil {
		return val.Type() == "ERROR"
	}
	return false
}

// Helper functions for creating categorized errors

// newTypeError creates a type error with position information from a node.
func (i *Interpreter) newTypeError(node ast.Node, format string, args ...interface{}) *ErrorValue {
	msg := fmt.Sprintf(format, args...)
	var pos *token.Position
	var expr string
	if node != nil {
		p := convertLexerToTokenPos(i.getPositionFromNode(node))
		pos = &p
		expr = node.String()
	}

	interpErr := interpErrors.NewTypeError(pos, msg, expr)
	return &ErrorValue{
		Message: interpErr.Error(),
		Err:     interpErr,
	}
}

// newRuntimeError creates a runtime error with position information from a node.
func (i *Interpreter) newRuntimeError(node ast.Node, format string, args ...interface{}) *ErrorValue {
	msg := fmt.Sprintf(format, args...)
	var pos *token.Position
	var expr string
	if node != nil {
		p := convertLexerToTokenPos(i.getPositionFromNode(node))
		pos = &p
		expr = node.String()
	}

	interpErr := interpErrors.NewRuntimeError(pos, msg, expr)
	return &ErrorValue{
		Message: interpErr.Error(),
		Err:     interpErr,
	}
}

// newRuntimeErrorWithValues creates a runtime error with runtime values for debugging.
func (i *Interpreter) newRuntimeErrorWithValues(node ast.Node, message string, values map[string]string) *ErrorValue {
	var pos *token.Position
	var expr string
	if node != nil {
		p := convertLexerToTokenPos(i.getPositionFromNode(node))
		pos = &p
		expr = node.String()
	}

	interpErr := interpErrors.NewRuntimeErrorWithValues(pos, expr, message, values)
	return &ErrorValue{
		Message: interpErr.Error(),
		Err:     interpErr,
	}
}

// newUndefinedError creates an undefined entity error with position information from a node.
func (i *Interpreter) newUndefinedError(node ast.Node, format string, args ...interface{}) *ErrorValue {
	msg := fmt.Sprintf(format, args...)
	var pos *token.Position
	var expr string
	if node != nil {
		p := convertLexerToTokenPos(i.getPositionFromNode(node))
		pos = &p
		expr = node.String()
	}

	interpErr := interpErrors.NewUndefinedError(pos, msg, expr)
	return &ErrorValue{
		Message: interpErr.Error(),
		Err:     interpErr,
	}
}

// newInternalError creates an internal interpreter error with position information from a node.
func (i *Interpreter) newInternalError(node ast.Node, format string, args ...interface{}) *ErrorValue {
	msg := fmt.Sprintf(format, args...)
	var pos *token.Position
	var expr string
	if node != nil {
		p := convertLexerToTokenPos(i.getPositionFromNode(node))
		pos = &p
		expr = node.String()
	}

	interpErr := interpErrors.NewInternalError(pos, msg, expr)
	return &ErrorValue{
		Message: interpErr.Error(),
		Err:     interpErr,
	}
}

// newContractErrorFromInterpreterError creates a contract error with position information.
func (i *Interpreter) newContractErrorFromInterpreterError(node ast.Node, format string, args ...interface{}) *ErrorValue {
	msg := fmt.Sprintf(format, args...)
	var pos *token.Position
	var expr string
	if node != nil {
		p := convertLexerToTokenPos(i.getPositionFromNode(node))
		pos = &p
		expr = node.String()
	}

	interpErr := interpErrors.NewContractError(pos, msg, expr)
	return &ErrorValue{
		Message: interpErr.Error(),
		Err:     interpErr,
	}
}

// convertLexerToTokenPos converts a lexer.Position to token.Position.
func convertLexerToTokenPos(pos lexer.Position) token.Position {
	return pos
}

// RuntimeError represents a structured runtime error with rich context
type RuntimeError struct {
	Message    string
	Pos        *lexer.Position
	Expression string            // The expression that failed
	Values     map[string]string // Runtime values as strings
	SourceCode string            // Full source code
	SourceFile string            // Source filename
	ErrorType  string            // Error classification
	CallStack  errors.StackTrace // Call stack at time of error
}

// Type implements the Value interface
func (r *RuntimeError) Type() string { return "ERROR" }

// String implements the Value interface
func (r *RuntimeError) String() string {
	if r.Pos != nil {
		return fmt.Sprintf("Runtime error at line %d: %s", r.Pos.Line, r.Message)
	}
	return fmt.Sprintf("Runtime error: %s", r.Message)
}

// ToCompilerError converts a RuntimeError to a CompilerError for display
func (r *RuntimeError) ToCompilerError() *errors.CompilerError {
	if r.Pos == nil || r.SourceCode == "" {
		// Fall back to simple error if no position info
		return nil
	}

	message := r.Message

	// Add runtime values if available
	if len(r.Values) > 0 {
		message += "\n"
		for name, value := range r.Values {
			message += fmt.Sprintf("  %s = %s\n", name, value)
		}
	}

	return errors.NewCompilerError(*r.Pos, message, r.SourceCode, r.SourceFile)
}

// NewRuntimeError creates a new structured runtime error
func (i *Interpreter) NewRuntimeError(node ast.Node, errorType, message string, values map[string]string) *RuntimeError {
	var pos *lexer.Position
	var expr string

	if node != nil {
		// Extract position from node
		p := i.getPositionFromNode(node)
		pos = &p
		expr = node.String()
	}

	return &RuntimeError{
		Message:    message,
		Pos:        pos,
		Expression: expr,
		Values:     values,
		SourceCode: i.sourceCode,
		SourceFile: i.sourceFile,
		ErrorType:  errorType,
		CallStack:  i.callStack,
	}
}

// getPositionFromNode extracts position from an AST node
func (i *Interpreter) getPositionFromNode(node ast.Node) lexer.Position {
	if node == nil {
		return lexer.Position{}
	}
	return node.Pos()
}
