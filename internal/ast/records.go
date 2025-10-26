// Package ast defines the Abstract Syntax Tree node types for DWScript.
// This file contains AST nodes for record types.
package ast

import (
	"bytes"
	"strings"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// ============================================================================
// Record Declaration (Task 8.56)
// ============================================================================

// RecordDecl represents a record type declaration in DWScript.
// Records are value types (like structs) with fields and optionally methods/properties.
// DWScript syntax:
//
//	type TRecordName = record
//	  Field1: Type1;
//	  Field2: Type2;
//	  function MethodName: ReturnType;
//	end;
type RecordDecl struct {
	Name       *Identifier
	Fields     []*FieldDecl
	Methods    []*FunctionDecl
	Properties []RecordPropertyDecl
	Token      lexer.Token
}

func (rd *RecordDecl) statementNode()       {}
func (rd *RecordDecl) TokenLiteral() string { return rd.Token.Literal }
func (rd *RecordDecl) Pos() lexer.Position  { return rd.Token.Pos }
func (rd *RecordDecl) String() string {
	var out bytes.Buffer

	out.WriteString("type ")
	out.WriteString(rd.Name.String())
	out.WriteString(" = record\n")

	// Add fields
	for _, field := range rd.Fields {
		out.WriteString("  ")
		out.WriteString(field.String())
		out.WriteString(";\n")
	}

	// Add methods
	for _, method := range rd.Methods {
		out.WriteString("  ")
		methodStr := method.String()
		// Indent multi-line method declarations
		out.WriteString(strings.ReplaceAll(methodStr, "\n", "\n  "))
		out.WriteString(";\n")
	}

	// Add properties (if any)
	for _, prop := range rd.Properties {
		out.WriteString("  ")
		out.WriteString(prop.String())
		out.WriteString(";\n")
	}

	out.WriteString("end")

	return out.String()
}

// RecordPropertyDecl represents a property declaration in a record.
// DWScript syntax: property Name: Type read Field write Field;
// Note: Renamed from PropertyDecl to avoid conflict with class PropertyDecl
type RecordPropertyDecl struct {
	Name       *Identifier
	Type       *TypeAnnotation
	ReadField  string
	WriteField string
	Token      lexer.Token
}

func (pd RecordPropertyDecl) String() string {
	var out bytes.Buffer

	out.WriteString("property ")
	out.WriteString(pd.Name.String())
	out.WriteString(": ")
	out.WriteString(pd.Type.String())

	if pd.ReadField != "" {
		out.WriteString(" read ")
		out.WriteString(pd.ReadField)
	}

	if pd.WriteField != "" {
		out.WriteString(" write ")
		out.WriteString(pd.WriteField)
	}

	return out.String()
}

// ============================================================================
// Record Literal (Task 8.57)
// ============================================================================

// RecordField represents a single field initialization in a record literal.
// Can be either named (X: 10) or positional (10).
type RecordField struct {
	Value Expression
	Name  string
}

// RecordLiteral represents a record literal expression.
// Examples:
//   - Named: (X: 10, Y: 20)
//   - Positional: (10, 20)
//   - Typed: TPoint(X: 10, Y: 20) or TPoint(10, 20)
type RecordLiteral struct {
	TypeName string
	Fields   []RecordField
	Token    lexer.Token
}

func (rl *RecordLiteral) expressionNode()      {}
func (rl *RecordLiteral) TokenLiteral() string { return rl.Token.Literal }
func (rl *RecordLiteral) Pos() lexer.Position  { return rl.Token.Pos }
func (rl *RecordLiteral) String() string {
	var out bytes.Buffer

	// Add type name if present
	if rl.TypeName != "" {
		out.WriteString(rl.TypeName)
	}

	out.WriteString("(")

	// Add fields
	for i, field := range rl.Fields {
		if i > 0 {
			out.WriteString(", ")
		}

		// Named field
		if field.Name != "" {
			out.WriteString(field.Name)
			out.WriteString(": ")
		}

		// Field value
		out.WriteString(field.Value.String())
	}

	out.WriteString(")")

	return out.String()
}
