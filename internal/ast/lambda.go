// Package ast defines the Abstract Syntax Tree node types for DWScript.
package ast

import (
	"bytes"
	"strings"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// LambdaExpression represents a lambda/anonymous function expression.
// Supports both full and shorthand syntax:
//
// Full syntax:
//
//	lambda(x: Integer, y: Integer): Integer begin Result := x + y; end
//
// Shorthand syntax (desugared to full internally):
//
//	lambda(x, y) => x + y
//
// The lambda creates a closure capturing variables from outer scopes.
// Closure capture semantics are handled during semantic analysis (tasks 9.216-9.220).
//
// Examples:
//
//	var doubled := Map(numbers, lambda(x: Integer): Integer begin Result := x * 2; end);
//	var squared := Map(numbers, lambda(x) => x * x);
//	var printer := lambda() begin PrintLn('Hello'); end;
type LambdaExpression struct {
	// Token is the 'lambda' keyword token
	Token lexer.Token

	// Parameters is the parameter list (may have optional types)
	Parameters []*Parameter

	// ReturnType is the optional return type annotation
	// If nil, type will be inferred during semantic analysis
	ReturnType *TypeAnnotation

	// Body is the lambda function body (always normalized to BlockStatement)
	// Shorthand syntax (=>) is desugared to a block with a single return statement
	Body *BlockStatement

	// Type is the inferred/assigned function pointer type
	// Set during semantic analysis
	Type *TypeAnnotation

	// IsShorthand indicates if this lambda was originally written with => syntax
	// Used for String() method to preserve original syntax in output
	IsShorthand bool
}

// expressionNode marks this as an Expression node
func (le *LambdaExpression) expressionNode() {}

// TokenLiteral returns the token literal ('lambda')
func (le *LambdaExpression) TokenLiteral() string {
	return le.Token.Literal
}

// Pos returns the position of the lambda keyword
func (le *LambdaExpression) Pos() lexer.Position {
	return le.Token.Pos
}

// GetType returns the inferred/assigned type annotation
func (le *LambdaExpression) GetType() *TypeAnnotation {
	return le.Type
}

// SetType sets the type annotation
func (le *LambdaExpression) SetType(typ *TypeAnnotation) {
	le.Type = typ
}

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
	out.WriteString(strings.Join(params, ", "))
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
