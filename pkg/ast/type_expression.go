package ast

import "github.com/cwbudde/go-dws/pkg/token"

// TypeExpression represents any type expression in DWScript.
// This interface unifies simple types (TypeAnnotation), function pointer types
// (FunctionPointerTypeNode), and array types (ArrayTypeNode).
//
// This allows types to be parsed and used inline in variable declarations,
// parameter lists, and other contexts without requiring type aliases.
type TypeExpression interface {
	Node
	typeExpressionNode() // Marker method to identify type expressions
}

// Ensure existing types implement TypeExpression
var (
	_ TypeExpression = (*TypeAnnotation)(nil)
	_ TypeExpression = (*FunctionPointerTypeNode)(nil)
	_ TypeExpression = (*ArrayTypeNode)(nil)
	_ TypeExpression = (*SetTypeNode)(nil)
	_ TypeExpression = (*ClassOfTypeNode)(nil)
)

// ArrayTypeNode represents an array type in inline type expressions.
// Supports both dynamic arrays (no bounds) and static arrays (with bounds).
//
// Examples:
//   - array of Integer (dynamic)
//   - array[1..10] of Integer (static)
//   - array of String (dynamic)
//   - array[0..99] of String (static)
//   - array of array of Integer (nested dynamic arrays)
//   - array[1..5] of array[1..10] of Integer (nested static arrays)
//   - array of function(x: Integer): Boolean (array of function pointers)
//   - array[TEnum] of String (enum-indexed array)
type ArrayTypeNode struct {
	ElementType TypeExpression
	LowBound    Expression
	HighBound   Expression
	IndexType   TypeExpression // For enum-indexed arrays: array[TEnum] of Type
	Token       token.Token
	EndPos      token.Position
}

// String returns a string representation of the array type.
func (at *ArrayTypeNode) String() string {
	if at == nil || at.ElementType == nil {
		return "array of <invalid>"
	}

	// Enum-indexed array: array[TEnum] of Integer
	if at.IndexType != nil {
		return "array[" + at.IndexType.String() + "] of " + at.ElementType.String()
	}

	// Static array with bounds: array[1..10] of Integer
	if at.IsStatic() {
		return "array[" +
			at.LowBound.String() + ".." +
			at.HighBound.String() + "] of " +
			at.ElementType.String()
	}

	// Dynamic array: array of Integer
	return "array of " + at.ElementType.String()
}

// IsDynamic returns true if this is a dynamic array (no bounds specified).
func (at *ArrayTypeNode) IsDynamic() bool {
	return at.IndexType == nil && (at.LowBound == nil || at.HighBound == nil)
}

// IsStatic returns true if this is a static array (bounds specified).
func (at *ArrayTypeNode) IsStatic() bool {
	return !at.IsDynamic()
}

// IsEnumIndexed returns true if this is an enum-indexed array.
func (at *ArrayTypeNode) IsEnumIndexed() bool {
	return at.IndexType != nil
}

// TokenLiteral returns the literal value of the token.
func (at *ArrayTypeNode) TokenLiteral() string {
	return at.Token.Literal
}

// Pos returns the position of the node in the source code.
func (at *ArrayTypeNode) Pos() token.Position {
	return at.Token.Pos
}

// End returns the end position of the node in the source code.
func (at *ArrayTypeNode) End() token.Position {
	if at.EndPos.Line != 0 {
		return at.EndPos
	}
	if at.ElementType != nil {
		return at.ElementType.End()
	}
	return at.Token.Pos
}

// typeExpressionNode marks this as a type expression
func (at *ArrayTypeNode) typeExpressionNode() {}

// SetTypeNode represents a set type in inline type expressions.
//
// Examples:
//   - set of TEnum (set of named enum type)
//   - set of (A, B, C) (set of inline anonymous enum)
//   - set of 2..1000 (set of inline subrange - if supported)
type SetTypeNode struct {
	ElementType TypeExpression
	Token       token.Token
	EndPos      token.Position
}

// String returns a string representation of the set type.
func (st *SetTypeNode) String() string {
	if st == nil || st.ElementType == nil {
		return "set of <invalid>"
	}
	return "set of " + st.ElementType.String()
}

// TokenLiteral returns the literal value of the token.
func (st *SetTypeNode) TokenLiteral() string {
	return st.Token.Literal
}

// Pos returns the position of the node in the source code.
func (st *SetTypeNode) Pos() token.Position {
	return st.Token.Pos
}

// End returns the end position of the node in the source code.
func (st *SetTypeNode) End() token.Position {
	if st.EndPos.Line != 0 {
		return st.EndPos
	}
	if st.ElementType != nil {
		return st.ElementType.End()
	}
	return st.Token.Pos
}

// typeExpressionNode marks this as a type expression
func (st *SetTypeNode) typeExpressionNode() {}

// ClassOfTypeNode represents a metaclass type in inline type expressions.
// A metaclass type is a reference to a class type itself, not an instance.
//
// Examples:
//   - class of TMyClass (metaclass/class reference type)
//   - class of TObject (metaclass of base class)
//
// Usage:
//
//	var cls: class of TMyClass;  // cls holds a reference to a class type
//	cls := TMyClass;              // assign class reference
//	cls := TDerivedClass;         // can assign derived class
//	obj := cls.Create;            // call constructor through metaclass
type ClassOfTypeNode struct {
	ClassType TypeExpression
	Token     token.Token
	EndPos    token.Position
}

// String returns a string representation of the metaclass type.
func (ct *ClassOfTypeNode) String() string {
	if ct == nil || ct.ClassType == nil {
		return "class of <invalid>"
	}
	return "class of " + ct.ClassType.String()
}

// TokenLiteral returns the literal value of the token.
func (ct *ClassOfTypeNode) TokenLiteral() string {
	return ct.Token.Literal
}

// Pos returns the position of the node in the source code.
func (ct *ClassOfTypeNode) Pos() token.Position {
	return ct.Token.Pos
}

// End returns the end position of the node in the source code.
func (ct *ClassOfTypeNode) End() token.Position {
	if ct.EndPos.Line != 0 {
		return ct.EndPos
	}
	if ct.ClassType != nil {
		return ct.ClassType.End()
	}
	return ct.Token.Pos
}

// typeExpressionNode marks this as a type expression
func (ct *ClassOfTypeNode) typeExpressionNode() {}
