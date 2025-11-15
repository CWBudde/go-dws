// Package ast defines the Abstract Syntax Tree node types for DWScript.
package ast

import (
	"fmt"

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

// String returns a simple string representation for debugging.
// For formatted output, use the printer package.
func (fd *FunctionDecl) String() string {
	return fmt.Sprintf("FunctionDecl(%s)", fd.Name.Value)
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
	BaseNode
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode() {}

// End returns the end position, preferring ReturnValue's end if EndPos not set.
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
