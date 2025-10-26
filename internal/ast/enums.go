package ast

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// ============================================================================
// Enum Declaration (Task 8.33)
// ============================================================================

// EnumValue represents a single value in an enum declaration.
// Example: Red, Green = 5, Blue
type EnumValue struct {
	Name  string // Value name (e.g., "Red", "Green")
	Value *int   // Optional explicit ordinal value (nil for implicit values)
}

// EnumDecl represents an enum type declaration.
// Examples:
//   - type TColor = (Red, Green, Blue);
//   - type TEnum = (One = 1, Two = 5);
//   - type TEnum = enum (One, Two);
type EnumDecl struct {
	Token  lexer.Token // The 'type' token
	Name   *Identifier // Enum type name
	Values []EnumValue // List of enum values
}

// statementNode implements the Statement interface
func (ed *EnumDecl) statementNode() {}

// TokenLiteral returns the literal value of the token
func (ed *EnumDecl) TokenLiteral() string {
	return ed.Token.Literal
}

// String returns a string representation of the enum declaration
func (ed *EnumDecl) String() string {
	var out strings.Builder

	out.WriteString("type ")
	out.WriteString(ed.Name.Value)
	out.WriteString(" = (")

	for i, val := range ed.Values {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(val.Name)
		if val.Value != nil {
			out.WriteString(fmt.Sprintf(" = %d", *val.Value))
		}
	}

	out.WriteString(")")

	return out.String()
}

// Pos returns the position of the enum declaration in the source code
func (ed *EnumDecl) Pos() lexer.Position {
	return ed.Token.Pos
}

// ============================================================================
// Enum Literal Expression (Task 8.34)
// ============================================================================

// EnumLiteral represents an enum value reference in an expression.
// Examples:
//   - Red (direct reference)
//   - TColor.Red (scoped reference)
type EnumLiteral struct {
	Token     lexer.Token // The identifier token
	EnumName  string      // Optional enum type name (for scoped reference like TColor.Red)
	ValueName string      // The enum value name (e.g., "Red")
}

// expressionNode implements the Expression interface
func (el *EnumLiteral) expressionNode() {}

// TokenLiteral returns the literal value of the token
func (el *EnumLiteral) TokenLiteral() string {
	return el.Token.Literal
}

// String returns a string representation of the enum literal
func (el *EnumLiteral) String() string {
	if el.EnumName != "" {
		return el.EnumName + "." + el.ValueName
	}
	return el.ValueName
}

// Pos returns the position of the enum literal in the source code
func (el *EnumLiteral) Pos() lexer.Position {
	return el.Token.Pos
}
