// Package ast defines the Abstract Syntax Tree node types for DWScript.
package ast

import (
	"bytes"
	"strings"
)

// LambdaExpression represents a lambda/anonymous function expression.
// Supports both full and shorthand syntax:
//
// Full syntax:
//
//	lambda(x: Integer; y: Integer): Integer begin Result := x + y; end
//	lambda(x, y: Integer): Integer begin Result := x + y; end
//
// Shorthand syntax (desugared to full internally):
//
//	lambda(x, y: Integer) => x + y
//
// The lambda creates a closure capturing variables from outer scopes.
// Closure capture semantics are handled during semantic analysis.
//
// Examples:
//
//	var doubled := Map(numbers, lambda(x: Integer): Integer begin Result := x * 2; end);
//	var squared := Map(numbers, lambda(x: Integer) => x * x);
//	var printer := lambda() begin PrintLn('Hello'); end;
type LambdaExpression struct {
	TypedExpressionBase
	ReturnType   *TypeAnnotation
	Body         *BlockStatement
	Parameters   []*Parameter
	CapturedVars []string
	IsShorthand  bool
}

// expressionNode marks this as an Expression node
func (le *LambdaExpression) expressionNode() {}

// String returns a string representation of the lambda expression
// Preserves the original syntax (full vs shorthand) for readability
func (le *LambdaExpression) String() string {
	var out bytes.Buffer

	out.WriteString("lambda")

	// Write parameter list
	out.WriteString("(")
	params := make([]string, len(le.Parameters))
	for i, param := range le.Parameters {
		params[i] = param.String()
	}
	out.WriteString(strings.Join(params, "; "))
	out.WriteString(")")

	// Write return type if present
	if le.ReturnType != nil {
		out.WriteString(": ")
		out.WriteString(le.ReturnType.String())
	}

	// Write body in appropriate syntax
	if le.IsShorthand {
		// Shorthand syntax: lambda(x) => expr
		out.WriteString(" => ")
		// Extract the expression from the desugared block
		// The block should contain a single return/expression statement
		if le.Body != nil && len(le.Body.Statements) > 0 {
			// If it's a return statement, extract the expression
			if returnStmt, ok := le.Body.Statements[0].(*ReturnStatement); ok && returnStmt.ReturnValue != nil {
				out.WriteString(returnStmt.ReturnValue.String())
			} else {
				// Otherwise just print the statement
				out.WriteString(le.Body.Statements[0].String())
			}
		}
	} else {
		// Full syntax: lambda(x: Integer) begin ... end
		out.WriteString(" ")
		if le.Body != nil && len(le.Body.Statements) > 0 {
			out.WriteString("begin ")
			for i, stmt := range le.Body.Statements {
				if i > 0 {
					out.WriteString("; ")
				}
				out.WriteString(stmt.String())
			}
			out.WriteString("; end")
		} else {
			out.WriteString("begin end")
		}
	}

	return out.String()
}
