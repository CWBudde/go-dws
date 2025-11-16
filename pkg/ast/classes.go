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

// String returns a full class declaration string.
// For formatted output, use the printer package.
func (cd *ClassDecl) String() string {
	var out bytes.Buffer

	out.WriteString("type ")
	out.WriteString(cd.Name.String())
	out.WriteString(" = ")

	// Add modifiers before "class"
	if cd.IsPartial {
		out.WriteString("partial ")
	}

	out.WriteString("class")

	// Add modifiers after "class"
	if cd.IsAbstract {
		out.WriteString(" abstract")
	}
	if cd.IsExternal {
		out.WriteString(" external")
	}

	// Add parent and/or interfaces
	if cd.Parent != nil || len(cd.Interfaces) > 0 {
		out.WriteString("(")
		items := []string{}
		if cd.Parent != nil {
			items = append(items, cd.Parent.String())
		}
		for _, intf := range cd.Interfaces {
			items = append(items, intf.String())
		}
		out.WriteString(strings.Join(items, ", "))
		out.WriteString(")")
	}

	out.WriteString("\n")

	// Add constants
	for _, constant := range cd.Constants {
		out.WriteString("  ")
		out.WriteString(constant.String())
		out.WriteString(";\n")
	}

	// Add fields
	for _, field := range cd.Fields {
		out.WriteString("  ")
		out.WriteString(field.String())
		out.WriteString(";\n")
	}

	// Add properties
	for _, property := range cd.Properties {
		out.WriteString("  ")
		out.WriteString(property.String())
		out.WriteString(";\n")
	}

	// Add constructor
	if cd.Constructor != nil {
		out.WriteString("  ")
		methodStr := cd.Constructor.String()
		out.WriteString(strings.ReplaceAll(methodStr, "\n", "\n  "))
		out.WriteString(";\n")
	}

	// Add methods
	for _, method := range cd.Methods {
		out.WriteString("  ")
		methodStr := method.String()
		out.WriteString(strings.ReplaceAll(methodStr, "\n", "\n  "))
		out.WriteString(";\n")
	}

	// Add destructor
	if cd.Destructor != nil {
		out.WriteString("  ")
		methodStr := cd.Destructor.String()
		out.WriteString(strings.ReplaceAll(methodStr, "\n", "\n  "))
		out.WriteString(";\n")
	}

	// Add operators
	for _, operator := range cd.Operators {
		out.WriteString("  ")
		out.WriteString(operator.String())
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
//	Field: String = 'hello';         // instance field with initializer
//	class var Count: Integer;         // class variable (static field)
//	class var Count: Integer := 42;   // class variable with initialization
//	property PropertyName: Type read FFieldName write FFieldName;
type FieldDecl struct {
	BaseNode
	Name       *Identifier
	Type       TypeExpression
	Visibility Visibility
	IsClassVar bool
	InitValue  Expression // Optional initialization value for instance fields and class variables
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
	TypedExpressionBase
	ClassName *Identifier  // The class name (e.g., TAnimal, TPerson)
	Arguments []Expression // Constructor arguments
}

func (ne *NewExpression) expressionNode() {}

// TokenLiteral returns the class name's token literal
func (ne *NewExpression) TokenLiteral() string {
	if ne.ClassName != nil {
		return ne.ClassName.TokenLiteral()
	}
	return ne.Token.Literal
}

// Pos returns the position from the ClassName
func (ne *NewExpression) Pos() token.Position {
	if ne.ClassName != nil {
		return ne.ClassName.Pos()
	}
	return ne.Token.Pos
}
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
	TypedExpressionBase
	Object Expression
	Member *Identifier
}

func (ma *MemberAccessExpression) expressionNode() {}

// Pos returns the start position from the Object expression
func (ma *MemberAccessExpression) Pos() token.Position {
	return ma.Object.Pos()
}
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
	TypedExpressionBase
	Object    Expression
	Method    *Identifier
	Arguments []Expression
}

func (mc *MethodCallExpression) expressionNode() {}

// Pos returns the start position from the Object expression
func (mc *MethodCallExpression) Pos() token.Position {
	return mc.Object.Pos()
}
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
	TypedExpressionBase
	Method    *Identifier
	Arguments []Expression
	IsCall    bool
	IsMember  bool
}

func (ie *InheritedExpression) expressionNode() {}
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

// SelfExpression represents the Self keyword in class methods.
// Used to refer to the current instance (in instance methods) or
// the current class (in class methods).
//
// DWScript syntax:
//
//	Self
//	Self.FieldName
//	Self.MethodName(args)
//	Self.ClassName
type SelfExpression struct {
	TypedExpressionBase
	Token token.Token // The 'self' token
}

func (se *SelfExpression) expressionNode() {}
func (se *SelfExpression) String() string {
	return "Self"
}
