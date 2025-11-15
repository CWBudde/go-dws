// Package ast defines the Abstract Syntax Tree node types for DWScript.
// This file contains AST nodes for object-oriented programming features.
package ast

import (
	"bytes"
	"strings"

	"github.com/cwbudde/go-dws/pkg/token"
)

// ============================================================================
// Visibility
// ============================================================================

// Visibility represents the access level of class members (fields and methods).
// DWScript supports three visibility levels: private, protected, and public.
type Visibility int

const (
	// VisibilityPrivate means the member is only accessible within the same class.
	VisibilityPrivate Visibility = iota

	// VisibilityProtected means the member is accessible within the same class
	// and all descendant classes.
	VisibilityProtected

	// VisibilityPublic means the member is accessible from anywhere.
	VisibilityPublic
)

// String returns the string representation of the visibility level.
func (v Visibility) String() string {
	switch v {
	case VisibilityPrivate:
		return "private"
	case VisibilityProtected:
		return "protected"
	case VisibilityPublic:
		return "public"
	default:
		return "unknown"
	}
}

// ============================================================================
// Class Declaration
// ============================================================================

// ClassDecl represents a class declaration in DWScript.
// DWScript syntax:
//
//	type TClassName = class(TParent)
//	  // fields and methods
//	end;
//	type TAbstract = class abstract
//	  // abstract class (cannot be instantiated)
//	end;
//	type TPartial = partial class
//	  // partial class (can be declared multiple times)
//	end;
type ClassDecl struct {
	BaseNode
	Constructor  *FunctionDecl
	Name         *Identifier
	Parent       *Identifier
	Destructor   *FunctionDecl
	ExternalName string
	Interfaces   []*Identifier
	Operators    []*OperatorDecl
	Properties   []*PropertyDecl
	Methods      []*FunctionDecl
	Fields       []*FieldDecl
	Constants    []*ConstDecl // Class constants
	IsAbstract   bool
	IsExternal   bool
	IsPartial    bool // True if declared with 'partial' keyword
}

func (cd *ClassDecl) statementNode() {}
func (cd *ClassDecl) String() string {
	var out bytes.Buffer

	out.WriteString("type ")
	out.WriteString(cd.Name.String())
	out.WriteString(" = ")

	// Add partial keyword if this is a partial class
	if cd.IsPartial {
		out.WriteString("partial ")
	}

	out.WriteString("class")

	// Add parent class and/or interfaces if present
	// Syntax: class(TParent, IInterface1, IInterface2) or class(IInterface) with no parent
	if cd.Parent != nil || len(cd.Interfaces) > 0 {
		out.WriteString("(")

		// Add parent first if present
		if cd.Parent != nil {
			out.WriteString(cd.Parent.String())
			// Add comma if there are also interfaces
			if len(cd.Interfaces) > 0 {
				out.WriteString(", ")
			}
		}

		// Add interfaces
		for i, iface := range cd.Interfaces {
			out.WriteString(iface.String())
			if i < len(cd.Interfaces)-1 {
				out.WriteString(", ")
			}
		}

		out.WriteString(")")
	}

	// Add abstract keyword if this is an abstract class
	if cd.IsAbstract {
		out.WriteString(" abstract")
	}

	out.WriteString("\n")

	// Add fields
	for _, field := range cd.Fields {
		out.WriteString("  ")
		out.WriteString(field.String())
		out.WriteString(";\n")
	}

	// Add constants
	for _, constant := range cd.Constants {
		out.WriteString("  ")
		out.WriteString(constant.String())
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

	// Add class operators
	for _, operator := range cd.Operators {
		out.WriteString("  ")
		if operator != nil {
			out.WriteString(operator.String())
		}
		out.WriteString(";\n")
	}

	// Add properties
	for _, property := range cd.Properties {
		out.WriteString("  ")
		if property != nil {
			out.WriteString(property.String())
		}
		out.WriteString("\n")
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
// Field Declaration
// ============================================================================

// FieldDecl represents a field (member variable) declaration in a class.
// DWScript syntax:
//
//	FFieldName: Type;                // instance field
//	class var Count: Integer;         // class variable (static field)
//	class var Count: Integer := 42;   // class variable with initialization
//	property PropertyName: Type read FFieldName write FFieldName;
type FieldDecl struct {
	BaseNode
	Name       *Identifier
	Type       TypeExpression
	Visibility Visibility
	IsClassVar bool
	InitValue  Expression // Optional initialization value for class variables
}

func (fd *FieldDecl) statementNode()       {}
func (fd *FieldDecl) TokenLiteral() string { return fd.Name.TokenLiteral() }
func (fd *FieldDecl) Pos() token.Position  { return fd.Name.Pos() }
func (fd *FieldDecl) String() string {
	var out bytes.Buffer

	if fd.IsClassVar {
		out.WriteString("class var ")
	}
	out.WriteString(fd.Name.String())
	if fd.Type != nil {
		out.WriteString(": ")
		out.WriteString(fd.Type.String())
	}
	if fd.InitValue != nil {
		out.WriteString(" := ")
		out.WriteString(fd.InitValue.String())
	}

	return out.String()
}

// ============================================================================
// Object Creation Expression
// ============================================================================

// NewExpression represents CLASS instantiation in DWScript using the 'new' keyword.
// This AST node is ONLY for creating objects from classes, not arrays.
//
// DWScript syntax:
//   - new TClassName(arg1, arg2)         // Create object with constructor arguments
//   - new TClassName()                   // Create object with no arguments
//   - TClassName.Create(arg1, arg2)      // Alternative syntax (also supported)
//
// For array instantiation with 'new', see NewArrayExpression instead.
//
// Related AST nodes:
//   - NewArrayExpression: for array instantiation (new Integer[16])
//   - ClassDeclaration: for class type definitions
//   - ConstructorDeclaration: for constructor method definitions
type NewExpression struct {
	ClassName *Identifier     // The class name (e.g., TAnimal, TPerson)
	Type      *TypeAnnotation // Inferred class type (for semantic analysis)
	Arguments []Expression    // Constructor arguments
	Token     token.Token     // The 'new' token or class name token
	EndPos    token.Position
}

func (n *NewExpression) End() token.Position {
	if n.EndPos.Line != 0 {
		return n.EndPos
	}
	return n.Token.Pos
}

func (ne *NewExpression) expressionNode()             {}
func (ne *NewExpression) TokenLiteral() string        { return ne.ClassName.TokenLiteral() }
func (ne *NewExpression) Pos() token.Position         { return ne.ClassName.Pos() }
func (ne *NewExpression) GetType() *TypeAnnotation    { return ne.Type }
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
// Member Access Expression
// ============================================================================

// MemberAccessExpression represents accessing a field or method of an object.
// DWScript syntax:
//
//	obj.field
//	obj.method
//	obj.field1.field2
type MemberAccessExpression struct {
	Object Expression
	Member *Identifier
	Type   *TypeAnnotation
	Token  token.Token
	EndPos token.Position
}

func (m *MemberAccessExpression) End() token.Position {
	if m.EndPos.Line != 0 {
		return m.EndPos
	}
	return m.Token.Pos
}

func (ma *MemberAccessExpression) expressionNode()             {}
func (ma *MemberAccessExpression) TokenLiteral() string        { return ma.Token.Literal }
func (ma *MemberAccessExpression) Pos() token.Position         { return ma.Object.Pos() }
func (ma *MemberAccessExpression) GetType() *TypeAnnotation    { return ma.Type }
func (ma *MemberAccessExpression) SetType(typ *TypeAnnotation) { ma.Type = typ }
func (ma *MemberAccessExpression) String() string {
	var out bytes.Buffer

	out.WriteString(ma.Object.String())
	out.WriteString(".")
	out.WriteString(ma.Member.String())

	return out.String()
}

// ============================================================================
// Method Call Expression
// ============================================================================

// MethodCallExpression represents calling a method on an object.
// DWScript syntax:
//
//	obj.MethodName(arg1, arg2)
//	obj.MethodName()
type MethodCallExpression struct {
	Object    Expression
	Method    *Identifier
	Type      *TypeAnnotation
	Arguments []Expression
	Token     token.Token
	EndPos    token.Position
}

func (m *MethodCallExpression) End() token.Position {
	if m.EndPos.Line != 0 {
		return m.EndPos
	}
	return m.Token.Pos
}

func (mc *MethodCallExpression) expressionNode()             {}
func (mc *MethodCallExpression) TokenLiteral() string        { return mc.Token.Literal }
func (mc *MethodCallExpression) Pos() token.Position         { return mc.Object.Pos() }
func (mc *MethodCallExpression) GetType() *TypeAnnotation    { return mc.Type }
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

// InheritedExpression represents a call to the parent class's implementation.
// Used in overridden methods to call the base class method.
//
// DWScript syntax:
//
//	inherited MethodName(args)
//	inherited MethodName
//	inherited
type InheritedExpression struct {
	Method    *Identifier
	Type      *TypeAnnotation
	Arguments []Expression
	Token     token.Token
	EndPos    token.Position
	IsCall    bool
	IsMember  bool
}

func (i *InheritedExpression) End() token.Position {
	if i.EndPos.Line != 0 {
		return i.EndPos
	}
	return i.Token.Pos
}

func (ie *InheritedExpression) expressionNode()             {}
func (ie *InheritedExpression) TokenLiteral() string        { return ie.Token.Literal }
func (ie *InheritedExpression) Pos() token.Position         { return ie.Token.Pos }
func (ie *InheritedExpression) GetType() *TypeAnnotation    { return ie.Type }
func (ie *InheritedExpression) SetType(typ *TypeAnnotation) { ie.Type = typ }
func (ie *InheritedExpression) String() string {
	var out bytes.Buffer
	out.WriteString("inherited")

	if ie.Method != nil {
		out.WriteString(" ")
		out.WriteString(ie.Method.String())

		if ie.IsCall {
			out.WriteString("(")
			args := []string{}
			for _, arg := range ie.Arguments {
				args = append(args, arg.String())
			}
			out.WriteString(strings.Join(args, ", "))
			out.WriteString(")")
		}
	}

	return out.String()
}
