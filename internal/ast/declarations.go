package ast

import (
	"bytes"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// ConstDecl represents a constant declaration statement.
// Constants are immutable values that can be used throughout the program.
// Examples:
//
//	const PI = 3.14;
//	const MAX_USERS: Integer = 1000;
//	const APP_NAME = 'MyApp';
type ConstDecl struct {
	Value Expression
	Name  *Identifier
	Type  *TypeAnnotation
	Token lexer.Token
}

func (cd *ConstDecl) statementNode()       {}
func (cd *ConstDecl) TokenLiteral() string { return cd.Token.Literal }
func (cd *ConstDecl) Pos() lexer.Position  { return cd.Token.Pos }
func (cd *ConstDecl) String() string {
	var out bytes.Buffer

	out.WriteString("const ")
	out.WriteString(cd.Name.String())

	if cd.Type != nil {
		out.WriteString(": ")
		out.WriteString(cd.Type.String())
	}

	out.WriteString(" = ")
	out.WriteString(cd.Value.String())

	return out.String()
}
