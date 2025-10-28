package ast

import "github.com/cwbudde/go-dws/internal/lexer"

// TypeExpression represents any type expression in DWScript.
// This interface unifies simple types (TypeAnnotation), function pointer types
// (FunctionPointerTypeNode), and array types (ArrayTypeNode).
//
// This allows types to be parsed and used inline in variable declarations,
// parameter lists, and other contexts without requiring type aliases.
//
// Task 9.49: Created to support inline type expressions
type TypeExpression interface {
	Node
	typeExpressionNode() // Marker method to identify type expressions
}

// Ensure existing types implement TypeExpression
var (
	_ TypeExpression = (*TypeAnnotation)(nil)
	_ TypeExpression = (*FunctionPointerTypeNode)(nil)
	_ TypeExpression = (*ArrayTypeNode)(nil)
)

// ArrayTypeNode represents a dynamic array type: `array of ElementType`
//
// Examples:
//   - array of Integer
//   - array of String
//   - array of array of Integer (nested arrays)
//   - array of function(x: Integer): Boolean (array of function pointers)
//
// Task 9.51: Created to support inline array type syntax
type ArrayTypeNode struct {
	Token       lexer.Token     // The 'array' token
	ElementType TypeExpression  // The element type (can be any type expression)
}

// String returns a string representation of the array type.
func (at *ArrayTypeNode) String() string {
	if at == nil || at.ElementType == nil {
		return "array of <invalid>"
	}
	return "array of " + at.ElementType.String()
}

// TokenLiteral returns the literal value of the token.
func (at *ArrayTypeNode) TokenLiteral() string {
	return at.Token.Literal
}

// Pos returns the position of the node in the source code.
func (at *ArrayTypeNode) Pos() lexer.Position {
	return at.Token.Pos
}

// typeExpressionNode marks this as a type expression
func (at *ArrayTypeNode) typeExpressionNode() {}
