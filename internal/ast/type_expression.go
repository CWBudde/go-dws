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
//
// Task 9.51: Created to support inline array type syntax
// Task 9.54: Extended to support static array bounds
type ArrayTypeNode struct {
	Token       lexer.Token    // The 'array' token
	ElementType TypeExpression // The element type (can be any type expression)
	LowBound    *int           // Low bound for static arrays (nil for dynamic)
	HighBound   *int           // High bound for static arrays (nil for dynamic)
}

// String returns a string representation of the array type.
func (at *ArrayTypeNode) String() string {
	if at == nil || at.ElementType == nil {
		return "array of <invalid>"
	}

	// Static array with bounds: array[1..10] of Integer
	if at.IsStatic() {
		return "array[" +
			intToString(*at.LowBound) + ".." +
			intToString(*at.HighBound) + "] of " +
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

// Size returns the number of elements in a static array.
// Returns -1 for dynamic arrays.
func (at *ArrayTypeNode) Size() int {
	if at.IsDynamic() {
		return -1
	}
	return *at.HighBound - *at.LowBound + 1
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

// intToString converts an integer to a string.
// Helper function for String() method.
func intToString(n int) string {
	if n < 0 {
		return "-" + intToString(-n)
	}
	if n < 10 {
		return string(rune('0' + n))
	}
	return intToString(n/10) + string(rune('0'+n%10))
}
