package ast

import "github.com/cwbudde/go-dws/internal/lexer"

// TypeExpression represents any type expression in DWScript.
// This interface unifies simple types (TypeAnnotation), function pointer types
// (FunctionPointerTypeNode), and array types (ArrayTypeNode).
//
// This allows types to be parsed and used inline in variable declarations,
// parameter lists, and other contexts without requiring type aliases.
type TypeExpression interface {
	Node
	typeExpressionNode() // Marker method to identify type expressions
}

// Ensure existing types implement TypeExpression
var (
	_ TypeExpression = (*TypeAnnotation)(nil)
	_ TypeExpression = (*FunctionPointerTypeNode)(nil)
	_ TypeExpression = (*ArrayTypeNode)(nil)
	_ TypeExpression = (*SetTypeNode)(nil)
)

// ArrayTypeNode represents an array type in inline type expressions.
// Supports both dynamic arrays (no bounds) and static arrays (with bounds).
//
// Examples:
//   - array of Integer (dynamic)
//   - array[1..10] of Integer (static)
//   - array of String (dynamic)
//   - array[0..99] of String (static)
//   - array of array of Integer (nested dynamic arrays)
//   - array[1..5] of array[1..10] of Integer (nested static arrays)
//   - array of function(x: Integer): Boolean (array of function pointers)
type ArrayTypeNode struct {
	Token       lexer.Token    // The 'array' token
	ElementType TypeExpression // The element type (can be any type expression)
	LowBound    Expression     // Low bound for static arrays (nil for dynamic)
	HighBound   Expression     // High bound for static arrays (nil for dynamic)
}

// String returns a string representation of the array type.
func (at *ArrayTypeNode) String() string {
	if at == nil || at.ElementType == nil {
		return "array of <invalid>"
	}

	// Static array with bounds: array[1..10] of Integer
	if at.IsStatic() {
		return "array[" +
			at.LowBound.String() + ".." +
			at.HighBound.String() + "] of " +
			at.ElementType.String()
	}

	// Dynamic array: array of Integer
	return "array of " + at.ElementType.String()
}

// IsDynamic returns true if this is a dynamic array (no bounds specified).
func (at *ArrayTypeNode) IsDynamic() bool {
	return at.LowBound == nil || at.HighBound == nil
}

// IsStatic returns true if this is a static array (bounds specified).
func (at *ArrayTypeNode) IsStatic() bool {
	return !at.IsDynamic()
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

// SetTypeNode represents a set type in inline type expressions.
//
// Examples:
//   - set of TEnum (set of named enum type)
//   - set of (A, B, C) (set of inline anonymous enum)
//   - set of 2..1000 (set of inline subrange - if supported)
type SetTypeNode struct {
	Token       lexer.Token    // The 'set' token
	ElementType TypeExpression // The element type (enum or subrange)
}

// String returns a string representation of the set type.
func (st *SetTypeNode) String() string {
	if st == nil || st.ElementType == nil {
		return "set of <invalid>"
	}
	return "set of " + st.ElementType.String()
}

// TokenLiteral returns the literal value of the token.
func (st *SetTypeNode) TokenLiteral() string {
	return st.Token.Literal
}

// Pos returns the position of the node in the source code.
func (st *SetTypeNode) Pos() lexer.Position {
	return st.Token.Pos
}

// typeExpressionNode marks this as a type expression
func (st *SetTypeNode) typeExpressionNode() {}
