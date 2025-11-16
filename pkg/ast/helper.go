// Package ast defines the Abstract Syntax Tree node types for DWScript.
// This file contains AST nodes for helper types.
package ast

import (
	"bytes"
	"strings"

	"github.com/cwbudde/go-dws/pkg/token"
)

// ============================================================================
// Helper Declaration
// ============================================================================

// HelperDecl represents a helper type declaration in DWScript.
// Helpers extend existing types with additional methods, properties, and class members
// without modifying the original type declaration.
//
// DWScript syntax supports two variants:
//
//	type TStringHelper = record helper for String
//	  function ToUpper: String;
//	  property Length: Integer read GetLength;
//	end;
//
//	type TIntHelper = helper for Integer
//	  function IsEven: Boolean;
//	end;
//
// Helpers can also inherit from other helpers:
//
//	type TParentHelper = helper for String
//	  function ToUpper: String;
//	end;
//
//	type TChildHelper = helper(TParentHelper) for String
//	  function ToLower: String;
//	end;
//
// Helpers can contain:
//   - Methods (functions and procedures)
//   - Properties
//   - Class variables (static fields)
//   - Class constants
//   - Visibility sections (private/public)
//
// Usage example:
//
//	var s := 'hello';
//	PrintLn(s.ToUpper());  // Calls TStringHelper.ToUpper with s as Self
type HelperDecl struct {
	ForType        TypeExpression
	Name           *Identifier
	ParentHelper   *Identifier
	Methods        []*FunctionDecl
	Properties     []*PropertyDecl
	ClassVars      []*FieldDecl
	ClassConsts    []*ConstDecl
	PrivateMembers []Statement
	PublicMembers  []Statement
	BaseNode
	IsRecordHelper bool
}

func (h *HelperDecl) End() token.Position {
	if h.EndPos.Line != 0 {
		return h.EndPos
	}
	return h.Token.Pos
}

func (hd *HelperDecl) statementNode() {}
func (hd *HelperDecl) String() string {
	var out bytes.Buffer

	out.WriteString("type ")
	out.WriteString(hd.Name.String())
	out.WriteString(" = ")

	if hd.IsRecordHelper {
		out.WriteString("record ")
	}

	out.WriteString("helper")

	// Include parent helper if present
	if hd.ParentHelper != nil {
		out.WriteString("(")
		out.WriteString(hd.ParentHelper.String())
		out.WriteString(")")
	}

	out.WriteString(" for ")
	out.WriteString(hd.ForType.String())
	out.WriteString("\n")

	// Track if we've written any private members
	hasPrivateSection := len(hd.PrivateMembers) > 0
	hasPublicSection := len(hd.PublicMembers) > 0

	// Write private section if exists
	if hasPrivateSection {
		out.WriteString("  private\n")
		for _, member := range hd.PrivateMembers {
			out.WriteString("    ")
			memberStr := member.String()
			// Indent multi-line declarations
			out.WriteString(strings.ReplaceAll(memberStr, "\n", "\n    "))
			out.WriteString(";\n")
		}
	}

	// Write public section if exists
	if hasPublicSection {
		out.WriteString("  public\n")
		for _, member := range hd.PublicMembers {
			out.WriteString("    ")
			memberStr := member.String()
			// Indent multi-line declarations
			out.WriteString(strings.ReplaceAll(memberStr, "\n", "\n    "))
			out.WriteString(";\n")
		}
	}

	// If no visibility sections, write all members at root level
	if !hasPrivateSection && !hasPublicSection {
		// Write class constants
		for _, classConst := range hd.ClassConsts {
			out.WriteString("  class ")
			out.WriteString(classConst.String())
			out.WriteString(";\n")
		}

		// Write class variables
		for _, classVar := range hd.ClassVars {
			out.WriteString("  ")
			out.WriteString(classVar.String())
			out.WriteString(";\n")
		}

		// Write methods
		for _, method := range hd.Methods {
			out.WriteString("  ")
			methodStr := method.String()
			// Indent multi-line method declarations
			out.WriteString(strings.ReplaceAll(methodStr, "\n", "\n  "))
			out.WriteString(";\n")
		}

		// Write properties
		for _, property := range hd.Properties {
			out.WriteString("  ")
			if property != nil {
				out.WriteString(property.String())
			}
			out.WriteString(";\n")
		}
	}

	out.WriteString("end")

	return out.String()
}
