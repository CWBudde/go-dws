// Package ast defines the Abstract Syntax Tree node types for DWScript.
// This file contains AST nodes for object-oriented programming features.
package ast

import (
	"bytes"
	"strings"

	"github.com/cwbudde/go-dws/lexer"
)

// ============================================================================
// Class Declaration (Task 7.7)
// ============================================================================

// ClassDecl represents a class declaration in DWScript.
// DWScript syntax:
//
//	type TClassName = class(TParent)
//	  // fields and methods
//	end;
type ClassDecl struct {
	Token       lexer.Token      // The 'type' token
	Name        *Identifier      // The class name (e.g., "TPoint", "TPerson")
	Parent      *Identifier      // The parent class name (optional, nil for root classes)
	Fields      []*FieldDecl     // Field declarations
	Methods     []*FunctionDecl  // Method declarations
	Constructor *FunctionDecl    // Constructor method (optional, usually named "Create")
	Destructor  *FunctionDecl    // Destructor method (optional, usually named "Destroy")
}

func (cd *ClassDecl) statementNode()       {}
func (cd *ClassDecl) TokenLiteral() string { return cd.Token.Literal }
func (cd *ClassDecl) Pos() lexer.Position  { return cd.Token.Pos }
func (cd *ClassDecl) String() string {
	var out bytes.Buffer

	out.WriteString("type ")
	out.WriteString(cd.Name.String())
	out.WriteString(" = class")

	// Add parent class if present
	if cd.Parent != nil {
		out.WriteString("(")
		out.WriteString(cd.Parent.String())
		out.WriteString(")")
	}

	out.WriteString("\n")

	// Add fields
	for _, field := range cd.Fields {
		out.WriteString("  ")
		out.WriteString(field.String())
		out.WriteString(";\n")
	}

	// Add methods
	for _, method := range cd.Methods {
		out.WriteString("  ")
		methodStr := method.String()
		// Indent multi-line method declarations
		out.WriteString(strings.ReplaceAll(methodStr, "\n", "\n  "))
		out.WriteString(";\n")
	}

	// Add constructor if present
	if cd.Constructor != nil {
		out.WriteString("  ")
		constructorStr := cd.Constructor.String()
		out.WriteString(strings.ReplaceAll(constructorStr, "\n", "\n  "))
		out.WriteString(";\n")
	}

	// Add destructor if present
	if cd.Destructor != nil {
		out.WriteString("  ")
		destructorStr := cd.Destructor.String()
		out.WriteString(strings.ReplaceAll(destructorStr, "\n", "\n  "))
		out.WriteString(";\n")
	}

	out.WriteString("end")

	return out.String()
}

// ============================================================================
// Field Declaration (Task 7.8)
// ============================================================================

// FieldDecl represents a field (member variable) declaration in a class.
// DWScript syntax:
//
//	FFieldName: Type;
//	property PropertyName: Type read FFieldName write FFieldName;
type FieldDecl struct {
	Token      lexer.Token      // The field name token
	Name       *Identifier      // The field name (e.g., "FValue", "X", "Y")
	Type       *TypeAnnotation  // The field type
	Visibility string           // Visibility: "public", "private", "protected"
}

func (fd *FieldDecl) statementNode()       {}
func (fd *FieldDecl) TokenLiteral() string { return fd.Name.TokenLiteral() }
func (fd *FieldDecl) Pos() lexer.Position  { return fd.Name.Pos() }
func (fd *FieldDecl) String() string {
	var out bytes.Buffer

	out.WriteString(fd.Name.String())
	out.WriteString(": ")
	out.WriteString(fd.Type.String())

	return out.String()
}

// ============================================================================
// Object Creation Expression (Task 7.9)
// ============================================================================

// NewExpression represents object instantiation in DWScript.
// DWScript syntax:
//
//	TClassName.Create(arg1, arg2)
//	or sometimes just:
//	TClassName.Create
type NewExpression struct {
	Token     lexer.Token      // The class name token
	ClassName *Identifier      // The class name (e.g., "TPoint")
	Arguments []Expression     // Constructor arguments
	Type      *TypeAnnotation  // The type (set by semantic analyzer)
}

func (ne *NewExpression) expressionNode()         {}
func (ne *NewExpression) TokenLiteral() string    { return ne.ClassName.TokenLiteral() }
func (ne *NewExpression) Pos() lexer.Position     { return ne.ClassName.Pos() }
func (ne *NewExpression) GetType() *TypeAnnotation { return ne.Type }
func (ne *NewExpression) SetType(typ *TypeAnnotation) { ne.Type = typ }
func (ne *NewExpression) String() string {
	var out bytes.Buffer

	out.WriteString(ne.ClassName.String())
	out.WriteString(".Create(")

	args := []string{}
	for _, arg := range ne.Arguments {
		args = append(args, arg.String())
	}
	out.WriteString(strings.Join(args, ", "))

	out.WriteString(")")

	return out.String()
}

// ============================================================================
// Member Access Expression (Task 7.10)
// ============================================================================

// MemberAccessExpression represents accessing a field or method of an object.
// DWScript syntax:
//
//	obj.field
//	obj.method
//	obj.field1.field2
type MemberAccessExpression struct {
	Token  lexer.Token      // The '.' token
	Object Expression       // The object expression (left side)
	Member *Identifier      // The member name (right side)
	Type   *TypeAnnotation  // The type (set by semantic analyzer)
}

func (ma *MemberAccessExpression) expressionNode()         {}
func (ma *MemberAccessExpression) TokenLiteral() string    { return ma.Token.Literal }
func (ma *MemberAccessExpression) Pos() lexer.Position     { return ma.Object.Pos() }
func (ma *MemberAccessExpression) GetType() *TypeAnnotation { return ma.Type }
func (ma *MemberAccessExpression) SetType(typ *TypeAnnotation) { ma.Type = typ }
func (ma *MemberAccessExpression) String() string {
	var out bytes.Buffer

	out.WriteString(ma.Object.String())
	out.WriteString(".")
	out.WriteString(ma.Member.String())

	return out.String()
}

// ============================================================================
// Method Call Expression (Task 7.11)
// ============================================================================

// MethodCallExpression represents calling a method on an object.
// DWScript syntax:
//
//	obj.MethodName(arg1, arg2)
//	obj.MethodName()
type MethodCallExpression struct {
	Token     lexer.Token      // The '.' token
	Object    Expression       // The object expression
	Method    *Identifier      // The method name
	Arguments []Expression     // The method arguments
	Type      *TypeAnnotation  // The return type (set by semantic analyzer)
}

func (mc *MethodCallExpression) expressionNode()         {}
func (mc *MethodCallExpression) TokenLiteral() string    { return mc.Token.Literal }
func (mc *MethodCallExpression) Pos() lexer.Position     { return mc.Object.Pos() }
func (mc *MethodCallExpression) GetType() *TypeAnnotation { return mc.Type }
func (mc *MethodCallExpression) SetType(typ *TypeAnnotation) { mc.Type = typ }
func (mc *MethodCallExpression) String() string {
	var out bytes.Buffer

	out.WriteString(mc.Object.String())
	out.WriteString(".")
	out.WriteString(mc.Method.String())
	out.WriteString("(")

	args := []string{}
	for _, arg := range mc.Arguments {
		args = append(args, arg.String())
	}
	out.WriteString(strings.Join(args, ", "))

	out.WriteString(")")

	return out.String()
}
