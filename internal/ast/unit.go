// Package ast defines the Abstract Syntax Tree node types for DWScript.
package ast

import (
	"bytes"
	"strings"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// UnitDeclaration represents a DWScript unit (module) declaration.
// A unit is a separate compilation module that can be imported by other units or programs.
//
// Example:
//
//	unit MyLibrary;
//	interface
//	  uses System;
//	  function Add(x, y: Integer): Integer;
//	implementation
//	  function Add(x, y: Integer): Integer;
//	  begin
//	    Result := x + y;
//	  end;
//	initialization
//	  // Setup code
//	finalization
//	  // Cleanup code
//	end.
type UnitDeclaration struct {
	// Name is the unit's identifier
	Name *Identifier

	// InterfaceSection contains public declarations (types, functions, etc.)
	// These symbols are visible to units that use this unit
	InterfaceSection *BlockStatement

	// ImplementationSection contains private implementation code
	// These symbols are only visible within this unit
	ImplementationSection *BlockStatement

	// InitSection contains initialization code (optional)
	// Runs when the unit is loaded, before the main program
	InitSection *BlockStatement

	// FinalSection contains finalization code (optional)
	// Runs when the program exits, in reverse order of initialization
	FinalSection *BlockStatement

	// Token is the 'unit' keyword token
	Token lexer.Token
}

func (ud *UnitDeclaration) statementNode()       {}
func (ud *UnitDeclaration) TokenLiteral() string { return ud.Token.Literal }
func (ud *UnitDeclaration) Pos() lexer.Position  { return ud.Token.Pos }
func (ud *UnitDeclaration) String() string {
	var out bytes.Buffer

	// Unit declaration
	out.WriteString("unit ")
	out.WriteString(ud.Name.String())
	out.WriteString(";\n\n")

	// Interface section
	if ud.InterfaceSection != nil {
		out.WriteString("interface\n")
		for _, stmt := range ud.InterfaceSection.Statements {
			out.WriteString(stmt.String())
			out.WriteString("\n")
		}
		out.WriteString("\n")
	}

	// Implementation section
	if ud.ImplementationSection != nil {
		out.WriteString("implementation\n")
		for _, stmt := range ud.ImplementationSection.Statements {
			out.WriteString(stmt.String())
			out.WriteString("\n")
		}
		out.WriteString("\n")
	}

	// Initialization section
	if ud.InitSection != nil {
		out.WriteString("initialization\n")
		for _, stmt := range ud.InitSection.Statements {
			out.WriteString("  ")
			out.WriteString(strings.ReplaceAll(stmt.String(), "\n", "\n  "))
			out.WriteString("\n")
		}
		out.WriteString("\n")
	}

	// Finalization section
	if ud.FinalSection != nil {
		out.WriteString("finalization\n")
		for _, stmt := range ud.FinalSection.Statements {
			out.WriteString("  ")
			out.WriteString(strings.ReplaceAll(stmt.String(), "\n", "\n  "))
			out.WriteString("\n")
		}
		out.WriteString("\n")
	}

	out.WriteString("end.")

	return out.String()
}

// UsesClause represents a uses statement that imports other units.
// The uses clause can appear in both the interface and implementation sections.
//
// Examples:
//
//	uses System;
//	uses System, Math, Graphics;
//	uses System.Collections, System.IO;
type UsesClause struct {
	// Units is the list of unit names to import
	Units []*Identifier

	// Token is the 'uses' keyword token
	Token lexer.Token
}

func (uc *UsesClause) statementNode()       {}
func (uc *UsesClause) TokenLiteral() string { return uc.Token.Literal }
func (uc *UsesClause) Pos() lexer.Position  { return uc.Token.Pos }
func (uc *UsesClause) String() string {
	var out bytes.Buffer

	out.WriteString("uses ")

	unitNames := []string{}
	for _, unit := range uc.Units {
		unitNames = append(unitNames, unit.String())
	}

	out.WriteString(strings.Join(unitNames, ", "))
	out.WriteString(";")

	return out.String()
}
