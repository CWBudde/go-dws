package ast

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/pkg/token"
)

// ============================================================================
// Enum Declaration
// ============================================================================

// EnumValue represents a single value in an enum declaration.
// Example: Red, Green = 5, Blue
type EnumValue struct {
	Value             *int
	Name              string
	DeprecatedMessage string // Optional message if deprecated
	IsDeprecated      bool   // True if marked as deprecated
}

// EnumDecl represents an enum type declaration.
// Examples:
//   - type TColor = (Red, Green, Blue);
//   - type TEnum = (One = 1, Two = 5);
//   - type TEnum = enum (One, Two);      // scoped enum
//   - type TFlags = flags (a, b, c);     // flags enum (bit flags)
type EnumDecl struct {
	BaseNode
	Name   *Identifier
	Values []EnumValue
	Scoped bool // True if declared with 'enum' or 'flags' keyword (requires qualified access)
	Flags  bool // True if declared with 'flags' keyword (uses power-of-2 values)
}

// statementNode implements the Statement interface
func (ed *EnumDecl) statementNode() {}

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
		if val.IsDeprecated {
			out.WriteString(" deprecated")
			if val.DeprecatedMessage != "" {
				out.WriteString(" '")
				out.WriteString(val.DeprecatedMessage)
				out.WriteString("'")
			}
		}
		if val.Value != nil {
			out.WriteString(fmt.Sprintf(" = %d", *val.Value))
		}
	}

	out.WriteString(")")

	return out.String()
}

// ============================================================================
// Enum Literal Expression
// ============================================================================

// EnumLiteral represents an enum value reference in an expression.
// Examples:
//   - Red (direct reference)
//   - TColor.Red (scoped reference)
type EnumLiteral struct {
	EnumName  string
	ValueName string
	Token     token.Token
	EndPos    token.Position
}

func (e *EnumLiteral) End() token.Position {
	if e.EndPos.Line != 0 {
		return e.EndPos
	}
	return e.Token.Pos
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
func (el *EnumLiteral) Pos() token.Position {
	return el.Token.Pos
}
