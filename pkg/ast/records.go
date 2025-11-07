// Package ast defines the Abstract Syntax Tree node types for DWScript.
// This file contains AST nodes for record types.
package ast

import (
	"bytes"
	"strings"

	"github.com/cwbudde/go-dws/pkg/token"
)

// ============================================================================
// Record Declaration
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
	Token      token.Token
	EndPos     token.Position
}

func (r *RecordDecl) End() token.Position {
	if r.EndPos.Line != 0 {
		return r.EndPos
	}
	return r.Token.Pos
}

func (rd *RecordDecl) statementNode()       {}
func (rd *RecordDecl) TokenLiteral() string { return rd.Token.Literal }
func (rd *RecordDecl) Pos() token.Position  { return rd.Token.Pos }
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
	Token      token.Token
	EndPos     token.Position
}

func (r *RecordPropertyDecl) End() token.Position {
	if r.EndPos.Line != 0 {
		return r.EndPos
	}
	return r.Token.Pos
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
// Record Literal
// ============================================================================

// FieldInitializer represents a single field initialization in a record literal.
// DWScript syntax: fieldName: value
// The Name can be nil for positional initialization (not yet implemented).
type FieldInitializer struct {
	Name   *Identifier // Field name (nil for positional initialization)
	Value  Expression  // Field value expression
	Token  token.Token // The field name token or first token of value
	EndPos token.Position
}

func (f *FieldInitializer) End() token.Position {
	if f.EndPos.Line != 0 {
		return f.EndPos
	}
	return f.Token.Pos
}

// String returns a string representation of the field initializer.
func (fi *FieldInitializer) String() string {
	var out bytes.Buffer

	// Named field
	if fi.Name != nil {
		out.WriteString(fi.Name.String())
		out.WriteString(": ")
	}

	// Field value
	out.WriteString(fi.Value.String())

	return out.String()
}

// RecordLiteralExpression represents a record literal expression.
// DWScript supports both anonymous and typed record literals:
//   - Anonymous: (x: 10; y: 20)
//   - Typed: TPoint(x: 10; y: 20)
//   - Semicolons or commas as separators: (a: 1; b: 2) or (a: 1, b: 2)
//
// Examples from Death_Star.dws:
//   - const big : TSphere = (cx: 20; cy: 20; cz: 0; r: 20);
//   - const small : TSphere = (cx: 7; cy: 7; cz: -10; r: 15);
type RecordLiteralExpression struct {
	TypeName *Identifier         // Optional type name (nil for anonymous records)
	Fields   []*FieldInitializer // Field initializers
	Token    token.Token         // The '(' token or type name token
	EndPos   token.Position
}

func (r *RecordLiteralExpression) End() token.Position {
	if r.EndPos.Line != 0 {
		return r.EndPos
	}
	return r.Token.Pos
}

func (rle *RecordLiteralExpression) expressionNode()      {}
func (rle *RecordLiteralExpression) TokenLiteral() string { return rle.Token.Literal }
func (rle *RecordLiteralExpression) Pos() token.Position  { return rle.Token.Pos }
func (rle *RecordLiteralExpression) String() string {
	var out bytes.Buffer

	// Add type name if present
	if rle.TypeName != nil {
		out.WriteString(rle.TypeName.String())
	}

	out.WriteString("(")

	// Add fields with semicolon separator (DWScript convention)
	for i, field := range rle.Fields {
		if i > 0 {
			out.WriteString("; ")
		}
		out.WriteString(field.String())
	}

	out.WriteString(")")

	return out.String()
}
