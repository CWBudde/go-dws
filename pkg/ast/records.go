// Package ast defines the Abstract Syntax Tree node types for DWScript.
// This file contains AST nodes for record types.
package ast

import (
	"bytes"
	"strings"

	"github.com/cwbudde/go-dws/pkg/token"
)

// RecordVisibilitySection records one visibility specifier written inside a
// record body, in source order. DWScript diagnoses redundant specifiers and
// rejects "protected" inside records, so the specifier keyword is kept as
// written (lower-cased) rather than collapsed onto Visibility.
type RecordVisibilitySection struct {
	// Specifier is the keyword as written, lower-cased: "private", "public",
	// "published" or "protected".
	Specifier string
	// Pos is the position of the specifier keyword.
	Pos token.Position
}

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
//	  const Origin = 0;               // Record constant
//	  class var Count: Integer;       // Class variable (shared across all instances)
//	  function MethodName: ReturnType;
//	end;
type RecordDecl struct {
	Name       *Identifier
	Fields     []*FieldDecl
	Methods    []*FunctionDecl
	Properties []RecordPropertyDecl
	Constants  []*ConstDecl
	ClassVars  []*FieldDecl
	// TypeParams holds the generic type-parameter names for a generic record
	// (e.g. ["A", "B"] for `type TRec<A,B> = record ... end;`). Empty for
	// non-generic records. Generic records are monomorphized before analysis.
	TypeParams []string
	// VisibilitySections lists the visibility specifiers written inside the
	// record body, in source order. It carries no child nodes.
	VisibilitySections []RecordVisibilitySection `ast:"skip"`
	BaseNode
	// EndKeywordPos is the position of the record's closing `end` keyword.
	EndKeywordPos token.Position `ast:"skip"`
}

func (rd *RecordDecl) statementNode() {}
func (rd *RecordDecl) String() string {
	var out bytes.Buffer

	out.WriteString("type ")
	out.WriteString(rd.Name.String())
	out.WriteString(" = record\n")

	// Add constants
	for _, constant := range rd.Constants {
		out.WriteString("  ")
		if constant.IsClassConst {
			out.WriteString("class ")
		}
		out.WriteString("const ")
		out.WriteString(constant.Name.String())
		if constant.Type != nil {
			out.WriteString(": ")
			out.WriteString(constant.Type.String())
		}
		out.WriteString(" = ")
		out.WriteString(constant.Value.String())
		out.WriteString(";\n")
	}

	// Add class variables
	for _, classVar := range rd.ClassVars {
		out.WriteString("  class var ")
		out.WriteString(classVar.String())
		out.WriteString(";\n")
	}

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
// Also supports array properties: property Name[Index: Type]: Type read Field;
// Note: Renamed from PropertyDecl to avoid conflict with class PropertyDecl
type RecordPropertyDecl struct {
	Type       TypeExpression
	Name       *Identifier
	ReadField  string
	WriteField string
	// ReadExpr holds an expression-based read specifier: read (2*Field).
	// When set, ReadField is empty.
	ReadExpr Expression
	// WriteStmt holds an expression-based write specifier: write (Field := Value)
	// or a normalized parenthesized lvalue write. When set, WriteField is empty.
	WriteStmt   Statement
	IndexParams []*Parameter
	BaseNode
	IsDefault bool
	// IsExternal is true for `property Name: Type external 'JsonKey' ...`.
	// ExternalName holds the quoted name, which replaces the declared name when
	// the record is serialized (JSON.Stringify).
	IsExternal   bool
	ExternalName string
	// DeprecatedMessage is the text of a `deprecated 'msg'` directive written
	// after the declaration; a bare `deprecated;` leaves it empty and sets
	// IsDeprecated alone.
	DeprecatedMessage string
	IsDeprecated      bool
	// IsClassProperty is true for `class property Name: Type ...` declared inside
	// a record body. A class property is backed by a class var rather than by an
	// instance field, and is reachable through both the record type and a value.
	IsClassProperty bool
	// IsAutoProperty is true when the property was declared without read/write
	// specifiers (e.g. `property Alpha: Integer;`). The parser desugars it to
	// read/write the synthesized backing member `F<Name>` and, while assembling
	// the record body, also synthesizes that member (see
	// Parser.addRecordAutoPropertyBackingField). Mirrors PropertyDecl.IsAutoProperty.
	IsAutoProperty bool
}

func (pd RecordPropertyDecl) String() string {
	var out bytes.Buffer

	out.WriteString("property ")
	out.WriteString(pd.Name.String())

	// Add index parameters for array properties
	if len(pd.IndexParams) > 0 {
		out.WriteString("[")
		for i, param := range pd.IndexParams {
			if i > 0 {
				out.WriteString("; ")
			}
			out.WriteString(param.String())
		}
		out.WriteString("]")
	}

	out.WriteString(": ")
	out.WriteString(pd.Type.String())

	if pd.ReadField != "" {
		out.WriteString(" read ")
		out.WriteString(pd.ReadField)
	} else if pd.ReadExpr != nil {
		out.WriteString(" read (")
		out.WriteString(pd.ReadExpr.String())
		out.WriteString(")")
	}

	if pd.WriteField != "" {
		out.WriteString(" write ")
		out.WriteString(pd.WriteField)
	} else if pd.WriteStmt != nil {
		out.WriteString(" write (")
		out.WriteString(pd.WriteStmt.String())
		out.WriteString(")")
	}

	if pd.IsDefault {
		out.WriteString("; default")
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
	Value Expression
	Name  *Identifier
	BaseNode
}

func (fi *FieldInitializer) statementNode() {}

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
	TypeName *Identifier
	Fields   []*FieldInitializer
	BaseNode
}

func (rle *RecordLiteralExpression) expressionNode() {}
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

// AnonymousRecordExpression represents DWScript's anonymous record constructor
// expression:
//
//	record a := 1; b := 'x'; end
//	record "i*i" := i * i; "2i" := 2 * i; end
//	record Field := 123 end
//
// Unlike RecordLiteralExpression, which is written with parentheses and needs an
// expected record type from its context, this form is structurally typed: the
// field names and the inferred types of their values fully describe the record,
// so it can stand alone as an expression (for example as an argument to
// JSON.Stringify).
//
// Field names may be written as identifiers or as string literals; a quoted name
// is kept verbatim in the Name identifier so that names which are not valid
// identifiers ("i*i", "2i") survive into serialization.
type AnonymousRecordExpression struct {
	Fields []*FieldInitializer
	BaseNode
}

func (are *AnonymousRecordExpression) expressionNode() {}

// String returns a string representation of the anonymous record expression.
func (are *AnonymousRecordExpression) String() string {
	var out bytes.Buffer

	out.WriteString("record ")
	for _, field := range are.Fields {
		if field.Name != nil {
			out.WriteString(field.Name.String())
			out.WriteString(" := ")
		}
		if field.Value != nil {
			out.WriteString(field.Value.String())
		}
		out.WriteString("; ")
	}
	out.WriteString("end")

	return out.String()
}
