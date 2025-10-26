package ast

import "github.com/cwbudde/go-dws/lexer"

// TypeAnnotation represents a type annotation in the AST.
// This is used for variable declarations, parameters, and return types.
// Example: `: Integer` in `var x: Integer := 5;`
type TypeAnnotation struct {
	Token lexer.Token // The ':' token or type name token
	Name  string      // The type name (e.g., "Integer", "String")
}

// String returns the string representation of the type annotation
func (ta *TypeAnnotation) String() string {
	if ta == nil {
		return ""
	}
	return ta.Name
}

// TypedExpression is an interface that all expressions with type information must implement.
// This allows the semantic analyzer to attach type information to expressions.
type TypedExpression interface {
	Expression
	// GetType returns the type of this expression (nil if not yet determined)
	GetType() *TypeAnnotation
	// SetType sets the type of this expression
	SetType(typ *TypeAnnotation)
}
