// Package ast defines the Abstract Syntax Tree node types for DWScript.
package ast

import (
	"bytes"
	"strings"

	"github.com/cwbudde/go-dws/lexer"
)

// Parameter represents a function parameter.
// Examples:
//
//	x: Integer
//	var s: String
//	a, b: Float
type Parameter struct {
	Token lexer.Token      // The parameter name token
	Name  *Identifier      // The parameter name
	Type  *TypeAnnotation  // The type annotation
	ByRef bool             // True for var parameters (pass by reference)
}

func (p *Parameter) String() string {
	result := ""
	if p.ByRef {
		result += "var "
	}
	result += p.Name.String() + ": " + p.Type.String()
	return result
}

// FunctionDecl represents a function or procedure declaration.
// Examples:
//
//	function Add(a: Integer, b: Integer): Integer; begin ... end;
//	procedure Hello; begin ... end;
type FunctionDecl struct {
	Token      lexer.Token      // The 'function' or 'procedure' token
	Name       *Identifier      // The function name
	Parameters []*Parameter     // The function parameters
	ReturnType *TypeAnnotation  // The return type (nil for procedures)
	Body       *BlockStatement  // The function body
}

func (fd *FunctionDecl) statementNode()       {}
func (fd *FunctionDecl) TokenLiteral() string { return fd.Token.Literal }
func (fd *FunctionDecl) Pos() lexer.Position  { return fd.Token.Pos }
func (fd *FunctionDecl) String() string {
	var out bytes.Buffer

	// Write function or procedure keyword
	out.WriteString(fd.Token.Literal)
	out.WriteString(" ")
	out.WriteString(fd.Name.String())

	// Write parameters - always show parentheses if there's a return type
	if fd.ReturnType != nil || len(fd.Parameters) > 0 {
		out.WriteString("(")
		params := []string{}
		for _, p := range fd.Parameters {
			params = append(params, p.String())
		}
		out.WriteString(strings.Join(params, "; "))
		out.WriteString(")")
	}

	// Write return type for functions
	if fd.ReturnType != nil {
		out.WriteString(": ")
		out.WriteString(fd.ReturnType.String())
	}

	// Write body
	out.WriteString(" ")
	out.WriteString(fd.Body.String())

	return out.String()
}

// ReturnStatement represents a return statement in a function.
// In DWScript, functions return via:
//   - Result := value (the Result variable)
//   - FunctionName := value (assigning to function name)
//   - exit (to exit early without explicit return)
//
// Examples:
//
//	Result := 42
//	Add := a + b
//	exit
type ReturnStatement struct {
	Token       lexer.Token // The 'Result', function name, or 'exit' token
	ReturnValue Expression  // The return value (nil for exit without value)
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) Pos() lexer.Position  { return rs.Token.Pos }
func (rs *ReturnStatement) String() string {
	if rs.ReturnValue == nil {
		return rs.Token.Literal
	}
	return rs.Token.Literal + " := " + rs.ReturnValue.String()
}
