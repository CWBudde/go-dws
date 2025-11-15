// Package ast defines the Abstract Syntax Tree node types for DWScript.
package ast

import (
	"bytes"
	"strings"

	"github.com/cwbudde/go-dws/pkg/token"
)

// Parameter represents a function parameter.
// Examples:
//
//	x: Integer
//	var s: String
//	const data: array of Integer
//	lazy expr: Integer
//	a, b: Float
//	prefix: String = 'Hello'  // optional parameter with default value
//
// Lazy parameters capture expressions, not values. The expression is re-evaluated
// each time the parameter is accessed within the function body, enabling patterns
// like Jensen's Device, conditional evaluation, and deferred computation.
//
// Const parameters are passed by const-reference, preventing modification while
// avoiding copy overhead for large types like arrays and records.
//
// Optional parameters have default values. When a function is called without
// providing a value for an optional parameter, the default expression is evaluated
// in the caller's context. Optional parameters must come after all required parameters.
type Parameter struct {
	DefaultValue Expression
	Name         *Identifier
	Type         *TypeAnnotation
	Token        token.Token
	EndPos       token.Position
	IsLazy       bool
	ByRef        bool
	IsConst      bool
}

func (p *Parameter) String() string {
	result := ""
	if p.IsConst {
		result += "const "
	}
	if p.IsLazy {
		result += "lazy "
	}
	if p.ByRef {
		result += "var "
	}

	// Handle shorthand syntax (no parameter name)
	if p.Name == nil {
		result += p.Type.String()
	} else {
		result += p.Name.String() + ": " + p.Type.String()
	}

	// Add default value if present
	if p.DefaultValue != nil {
		result += " = " + p.DefaultValue.String()
	}

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
//
// With contracts:
//
//	function DotProduct(a, b: array of Float): Float;
//	require
//	   a.Length = b.Length;
//	begin
//	   // ... implementation
//	end;
//	ensure
//	   Result >= 0;
type FunctionDecl struct {
	BaseNode
	ClassName         *Identifier
	ReturnType        *TypeAnnotation
	Body              *BlockStatement
	PreConditions     *PreConditions
	PostConditions    *PostConditions
	Name              *Identifier
	ExternalName      string
	CallingConvention string // e.g., "register", "pascal", "cdecl", "safecall", "stdcall"
	DeprecatedMessage string // Optional message if deprecated
	Parameters        []*Parameter
	Visibility        Visibility
	IsDestructor      bool
	IsVirtual         bool
	IsOverride        bool
	IsReintroduce     bool // True if marked as reintroduce (hides parent method)
	IsAbstract        bool
	IsExternal        bool
	IsClassMethod     bool
	IsOverload        bool
	IsForward         bool
	IsConstructor     bool
	IsDeprecated      bool // True if marked as deprecated
}

func (fd *FunctionDecl) statementNode() {}
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

	// Write virtual/override/abstract/overload/forward directives
	if fd.IsVirtual {
		out.WriteString("; virtual")
	}
	if fd.IsOverride {
		out.WriteString("; override")
	}
	if fd.IsAbstract {
		out.WriteString("; abstract")
	}
	if fd.IsOverload {
		out.WriteString("; overload")
	}
	if fd.IsForward {
		out.WriteString("; forward")
	}

	// Write calling convention directive
	if fd.CallingConvention != "" {
		out.WriteString("; ")
		out.WriteString(fd.CallingConvention)
	}

	// Write deprecated directive
	if fd.IsDeprecated {
		out.WriteString("; deprecated")
		if fd.DeprecatedMessage != "" {
			out.WriteString(" '")
			out.WriteString(fd.DeprecatedMessage)
			out.WriteString("'")
		}
	}

	// Write preconditions if present
	if fd.PreConditions != nil {
		out.WriteString("\n")
		out.WriteString(fd.PreConditions.String())
	}

	// Write body (abstract methods have no body)
	if fd.Body != nil {
		out.WriteString(" ")
		out.WriteString(fd.Body.String())
	}

	// Write postconditions if present
	if fd.PostConditions != nil {
		out.WriteString("\n")
		out.WriteString(fd.PostConditions.String())
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
	Token       token.Token
	EndPos      token.Position
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) Pos() token.Position  { return rs.Token.Pos }
func (rs *ReturnStatement) End() token.Position {
	if rs.EndPos.Line != 0 {
		return rs.EndPos
	}
	if rs.ReturnValue != nil {
		return rs.ReturnValue.End()
	}
	return rs.Token.Pos
}
func (rs *ReturnStatement) String() string {
	if rs.ReturnValue == nil {
		return rs.Token.Literal
	}
	return rs.Token.Literal + " := " + rs.ReturnValue.String()
}
