// Package ast defines the Abstract Syntax Tree node types for DWScript.
package ast

import (
	"bytes"
	"strings"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// Parameter represents a function parameter.
// Examples:
//
//	x: Integer
//	var s: String
//	a, b: Float
type Parameter struct {
	Name  *Identifier
	Type  *TypeAnnotation
	Token lexer.Token
	ByRef bool
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
//	class function GetCount: Integer; static; begin ... end;  // class method
//	function DoWork(): Integer; virtual; begin ... end;  // virtual method
//	function DoWork(): Integer; override; begin ... end;  // override method
//	function GetArea(): Float; abstract;  // abstract method
type FunctionDecl struct {
	Name          *Identifier
	ClassName     *Identifier
	ReturnType    *TypeAnnotation
	Body          *BlockStatement
	ExternalName  string
	Parameters    []*Parameter
	Token         lexer.Token
	Visibility    Visibility
	IsDestructor  bool
	IsConstructor bool
	IsVirtual     bool
	IsOverride    bool
	IsAbstract    bool
	IsExternal    bool
	IsClassMethod bool
}

func (fd *FunctionDecl) statementNode()       {}
func (fd *FunctionDecl) TokenLiteral() string { return fd.Token.Literal }
func (fd *FunctionDecl) Pos() lexer.Position  { return fd.Token.Pos }
func (fd *FunctionDecl) String() string {
	var out bytes.Buffer

	// Write "class" prefix for class methods
	if fd.IsClassMethod {
		out.WriteString("class ")
	}

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

	// Write virtual/override/abstract directives
	if fd.IsVirtual {
		out.WriteString("; virtual")
	}
	if fd.IsOverride {
		out.WriteString("; override")
	}
	if fd.IsAbstract {
		out.WriteString("; abstract")
	}

	// Write body (abstract methods have no body)
	if fd.Body != nil {
		out.WriteString(" ")
		out.WriteString(fd.Body.String())
	}

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
	ReturnValue Expression
	Token       lexer.Token
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
