// Package ast defines the Abstract Syntax Tree node types for DWScript.
// This file contains AST nodes for interface declarations.
package ast

import (
	"bytes"
	"strings"

	"github.com/cwbudde/go-dws/pkg/token"
)

// ============================================================================
// Interface Method Declaration
// ============================================================================

// InterfaceMethodDecl represents a method declaration within an interface.
// DWScript syntax:
//
//	procedure Hello;
//	function GetValue: Integer;
//	procedure SetValue(x: Integer);
//
// Note: Interface methods only declare signatures, they have no body.
type InterfaceMethodDecl struct {
	BaseNode
	Name       *Identifier
	ReturnType *TypeAnnotation
	Parameters []*Parameter
}

func (imd *InterfaceMethodDecl) statementNode() {}

func (i *InterfaceMethodDecl) End() token.Position {
	if i.EndPos.Line != 0 {
		return i.EndPos
	}
	return i.Token.Pos
}

func (imd *InterfaceMethodDecl) String() string {
	var out bytes.Buffer

	// Write "function" or "procedure" keyword based on whether there's a return type
	if imd.ReturnType != nil {
		out.WriteString("function ")
	} else {
		out.WriteString("procedure ")
	}

	// Write method name
	out.WriteString(imd.Name.String())

	// Write parameters if present
	if len(imd.Parameters) > 0 {
		out.WriteString("(")
		params := []string{}
		for _, p := range imd.Parameters {
			params = append(params, p.String())
		}
		out.WriteString(strings.Join(params, "; "))
		out.WriteString(")")
	}

	// Write return type for functions
	if imd.ReturnType != nil {
		out.WriteString(": ")
		out.WriteString(imd.ReturnType.String())
	}

	return out.String()
}

// ============================================================================
// Interface Declaration
// ============================================================================

// InterfaceDecl represents an interface declaration in DWScript.
// DWScript syntax:
//
//	type IMyInterface = interface
//	  procedure Hello;
//	  function GetValue: Integer;
//	end;
//
//	type IDescendent = interface(IBase)
//	  procedure AdditionalMethod;
//	end;
type InterfaceDecl struct {
	BaseNode
	Name         *Identifier
	Parent       *Identifier
	ExternalName string
	Methods      []*InterfaceMethodDecl
	IsExternal   bool
}

func (i *InterfaceDecl) End() token.Position {
	if i.EndPos.Line != 0 {
		return i.EndPos
	}
	return i.Token.Pos
}

func (id *InterfaceDecl) statementNode() {}
func (id *InterfaceDecl) String() string {
	var out bytes.Buffer

	out.WriteString("type ")
	out.WriteString(id.Name.String())
	out.WriteString(" = interface")

	// Add parent interface if present (interface inheritance)
	if id.Parent != nil {
		out.WriteString("(")
		out.WriteString(id.Parent.String())
		out.WriteString(")")
	}

	out.WriteString("\n")

	// Add method declarations
	for _, method := range id.Methods {
		out.WriteString("  ")
		out.WriteString(method.String())
		out.WriteString(";\n")
	}

	out.WriteString("end")

	return out.String()
}
