package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/ast"
)

// ErrorValue represents a runtime error.
type ErrorValue struct {
	Message string
}

func (e *ErrorValue) Type() string   { return "ERROR" }
func (e *ErrorValue) String() string { return "ERROR: " + e.Message }

// newError creates a new ErrorValue.
func newError(format string, args ...interface{}) *ErrorValue {
	return &ErrorValue{Message: fmt.Sprintf(format, args...)}
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
				message = fmt.Sprintf("%s at %s", message, location)
			}
		}
	}

	return &ErrorValue{Message: message}
}

// getLocationFromNode extracts location information from an AST node
func (i *Interpreter) getLocationFromNode(node ast.Node) string {
	// Try to extract token information from various node types
	switch n := node.(type) {
	case *ast.Identifier:
		return fmt.Sprintf("line %d, column %d", n.Token.Pos.Line, n.Token.Pos.Column)
	case *ast.IntegerLiteral:
		return fmt.Sprintf("line %d, column %d", n.Token.Pos.Line, n.Token.Pos.Column)
	case *ast.FloatLiteral:
		return fmt.Sprintf("line %d, column %d", n.Token.Pos.Line, n.Token.Pos.Column)
	case *ast.StringLiteral:
		return fmt.Sprintf("line %d, column %d", n.Token.Pos.Line, n.Token.Pos.Column)
	case *ast.BooleanLiteral:
		return fmt.Sprintf("line %d, column %d", n.Token.Pos.Line, n.Token.Pos.Column)
	case *ast.BinaryExpression:
		return fmt.Sprintf("line %d, column %d", n.Token.Pos.Line, n.Token.Pos.Column)
	case *ast.UnaryExpression:
		return fmt.Sprintf("line %d, column %d", n.Token.Pos.Line, n.Token.Pos.Column)
	case *ast.CallExpression:
		return fmt.Sprintf("line %d, column %d", n.Token.Pos.Line, n.Token.Pos.Column)
	case *ast.VarDeclStatement:
		return fmt.Sprintf("line %d, column %d", n.Token.Pos.Line, n.Token.Pos.Column)
	case *ast.AssignmentStatement:
		return fmt.Sprintf("line %d, column %d", n.Token.Pos.Line, n.Token.Pos.Column)
	}
	return ""
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
