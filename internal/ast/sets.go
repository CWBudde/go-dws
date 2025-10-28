// Package ast defines the Abstract Syntax Tree node types for DWScript.
// This file contains AST nodes for set type declarations and set literals (Stage 8).
package ast

import (
	"bytes"
	"strings"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// ============================================================================
// Set Declaration
// ============================================================================

// SetDecl represents a set type declaration.
// Examples:
//   - type TDays = set of TWeekday;
//   - var s: set of (Mon, Tue, Wed);  // inline with anonymous enum
type SetDecl struct {
	Name        *Identifier
	ElementType *TypeAnnotation
	Token       lexer.Token
}

// statementNode implements the Statement interface
func (sd *SetDecl) statementNode() {}

// TokenLiteral returns the literal value of the token
func (sd *SetDecl) TokenLiteral() string {
	return sd.Token.Literal
}

// String returns a string representation of the set declaration
func (sd *SetDecl) String() string {
	var out bytes.Buffer

	if sd.Name != nil {
		out.WriteString("type ")
		out.WriteString(sd.Name.Value)
		out.WriteString(" = ")
	}

	out.WriteString("set of ")
	if sd.ElementType != nil {
		out.WriteString(sd.ElementType.String())
	}

	return out.String()
}

// Pos returns the position of the set declaration in the source code
func (sd *SetDecl) Pos() lexer.Position {
	return sd.Token.Pos
}

// ============================================================================
// Set Literal Expression
// ============================================================================

// SetLiteral represents a set literal expression.
// Examples:
//   - [one, two, three]        // set with elements
//   - [one..five]              // set with range (not yet implemented)
//   - []                       // empty set
type SetLiteral struct {
	Type     *TypeAnnotation
	Elements []Expression
	Token    lexer.Token
}

// expressionNode implements the Expression interface
func (sl *SetLiteral) expressionNode() {}

// TokenLiteral returns the literal value of the token
func (sl *SetLiteral) TokenLiteral() string {
	return sl.Token.Literal
}

// String returns a string representation of the set literal
func (sl *SetLiteral) String() string {
	var out bytes.Buffer

	out.WriteString("[")

	elements := []string{}
	for _, elem := range sl.Elements {
		elements = append(elements, elem.String())
	}
	out.WriteString(strings.Join(elements, ", "))

	out.WriteString("]")

	return out.String()
}

// Pos returns the position of the set literal in the source code
func (sl *SetLiteral) Pos() lexer.Position {
	return sl.Token.Pos
}

// GetType returns the inferred type annotation
func (sl *SetLiteral) GetType() *TypeAnnotation {
	return sl.Type
}

// SetType sets the type annotation
func (sl *SetLiteral) SetType(typ *TypeAnnotation) {
	sl.Type = typ
}
