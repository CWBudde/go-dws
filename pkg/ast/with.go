package ast

import (
	"bytes"
	"strings"
)

// WithStatement represents DWScript's local declaration with statement.
// Example:
//
//	with x := 1, y : Integer = 2 do begin
//	  PrintLn(x + y);
//	end;
type WithStatement struct {
	Body         Statement
	Declarations []*VarDeclStatement
	BaseNode
}

func (ws *WithStatement) statementNode() {}

func (ws *WithStatement) String() string {
	var out bytes.Buffer

	out.WriteString("with ")
	declarations := make([]string, 0, len(ws.Declarations))
	for _, decl := range ws.Declarations {
		declarations = append(declarations, strings.TrimPrefix(decl.String(), "var "))
	}
	out.WriteString(strings.Join(declarations, ", "))
	out.WriteString(" do ")
	if ws.Body != nil {
		out.WriteString(ws.Body.String())
	}

	return out.String()
}
