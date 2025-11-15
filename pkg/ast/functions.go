// Package ast defines the Abstract Syntax Tree node types for DWScript.
package ast

import (
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
	IsDefault         bool // True if constructor is marked as default
	IsDeprecated      bool // True if marked as deprecated
}

func (fd *FunctionDecl) statementNode() {}

// String returns the function signature.
// For formatted output, use the printer package.
func (fd *FunctionDecl) String() string {
	var result strings.Builder

	// Add class method modifier
	if fd.IsClassMethod {
		result.WriteString("class ")
	}

	// Add function/procedure keyword
	if fd.IsConstructor {
		result.WriteString("constructor ")
	} else if fd.IsDestructor {
		result.WriteString("destructor ")
	} else if fd.ReturnType != nil {
		result.WriteString("function ")
	} else {
		result.WriteString("procedure ")
	}

	// Add function name
	result.WriteString(fd.Name.Value)

	// Add parameters (include parentheses only if there are parameters OR if there's a return type)
	if len(fd.Parameters) > 0 || fd.ReturnType != nil {
		result.WriteString("(")
		for i, param := range fd.Parameters {
			if i > 0 {
				result.WriteString("; ")
			}
			result.WriteString(param.String())
		}
		result.WriteString(")")
	}

	// Add return type for functions
	if fd.ReturnType != nil && !fd.IsConstructor {
		result.WriteString(": ")
		result.WriteString(fd.ReturnType.String())
	}

	// Add modifiers
	var modifiers []string
	if fd.IsVirtual {
		modifiers = append(modifiers, "virtual")
	}
	if fd.IsOverride {
		modifiers = append(modifiers, "override")
	}
	if fd.IsReintroduce {
		modifiers = append(modifiers, "reintroduce")
	}
	if fd.IsAbstract {
		modifiers = append(modifiers, "abstract")
	}
	if fd.IsOverload {
		modifiers = append(modifiers, "overload")
	}
	if fd.IsExternal {
		modifiers = append(modifiers, "external")
	}
	if fd.IsForward {
		modifiers = append(modifiers, "forward")
	}
	if fd.IsDeprecated {
		modifiers = append(modifiers, "deprecated")
	}

	if len(modifiers) > 0 {
		result.WriteString("; ")
		result.WriteString(strings.Join(modifiers, "; "))
	}

	// Add preconditions if present
	if fd.PreConditions != nil {
		result.WriteString("\n")
		result.WriteString(fd.PreConditions.String())
	}

	// Add body if present
	if fd.Body != nil {
		if fd.PreConditions == nil {
			result.WriteString(" ")
		}
		result.WriteString(fd.Body.String())
	}

	// Add postconditions if present
	if fd.PostConditions != nil {
		result.WriteString("\n")
		result.WriteString(fd.PostConditions.String())
	}

	return result.String()
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
