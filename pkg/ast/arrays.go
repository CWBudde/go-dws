// Package ast defines the Abstract Syntax Tree node types for DWScript.
// This file contains AST nodes for array type annotations, array literals, and array indexing.
package ast

import (
	"bytes"
	"strings"

	"github.com/cwbudde/go-dws/pkg/token"
)

// ============================================================================
// ArrayDecl
// ============================================================================

// ArrayDecl represents an array type declaration statement.
// Examples:
//   - type TMyArray = array[1..10] of Integer;
//   - type TDynamic = array of String;
type ArrayDecl struct {
	Name      *Identifier
	ArrayType *ArrayTypeAnnotation
	BaseNode
}

// statementNode implements the Statement interface
func (ad *ArrayDecl) statementNode() {}

// String returns a string representation of the array declaration
func (ad *ArrayDecl) String() string {
	var out bytes.Buffer

	out.WriteString("type ")
	if ad.Name != nil {
		out.WriteString(ad.Name.Value)
	}
	out.WriteString(" = ")
	if ad.ArrayType != nil {
		out.WriteString(ad.ArrayType.String())
	}

	return out.String()
}

// ============================================================================
// ArrayTypeAnnotation
// ============================================================================

// ArrayTypeAnnotation represents an array type annotation.
// DWScript supports both static arrays (with bounds) and dynamic arrays.
// Examples:
//   - array[1..10] of Integer (static, with bounds)
//   - array of String (dynamic, no bounds)
type ArrayTypeAnnotation struct {
	ElementType TypeExpression // Can be TypeAnnotation, ArrayTypeNode, FunctionPointerTypeNode, etc.
	LowBound    Expression
	HighBound   Expression
	Token       token.Token
	EndPos      token.Position
}

func (a *ArrayTypeAnnotation) End() token.Position {
	if a.EndPos.Line != 0 {
		return a.EndPos
	}
	return a.Token.Pos
}

// TokenLiteral returns the literal value of the token
func (ata *ArrayTypeAnnotation) TokenLiteral() string {
	return ata.Token.Literal
}

// String returns a string representation of the array type annotation
func (ata *ArrayTypeAnnotation) String() string {
	var out bytes.Buffer

	out.WriteString("array")

	if ata.LowBound != nil && ata.HighBound != nil {
		out.WriteString("[")
		out.WriteString(ata.LowBound.String())
		out.WriteString("..")
		out.WriteString(ata.HighBound.String())
		out.WriteString("]")
	}

	out.WriteString(" of ")
	if ata.ElementType != nil {
		out.WriteString(ata.ElementType.String())
	}

	return out.String()
}

// Pos returns the position of the array type annotation in the source code
func (ata *ArrayTypeAnnotation) Pos() token.Position {
	return ata.Token.Pos
}

// IsDynamic returns true if this is a dynamic array (no bounds)
func (ata *ArrayTypeAnnotation) IsDynamic() bool {
	return ata.LowBound == nil && ata.HighBound == nil
}

// IsStatic returns true if this is a static array (with bounds)
func (ata *ArrayTypeAnnotation) IsStatic() bool {
	return !ata.IsDynamic()
}

// ============================================================================
// ArrayLiteralExpression
// ============================================================================

// ArrayLiteralExpression represents an array literal expression.
// Examples:
//   - [1, 2, 3]        // array with elements
//   - []               // empty array
//   - ['a', 'b', 'c']  // array of strings
type ArrayLiteralExpression struct {
	Elements []Expression
	TypedExpressionBase
}

// expressionNode implements the Expression interface
func (al *ArrayLiteralExpression) expressionNode() {}

// String returns a string representation of the array literal
func (al *ArrayLiteralExpression) String() string {
	var out bytes.Buffer

	out.WriteString("[")

	elements := []string{}
	for _, elem := range al.Elements {
		elements = append(elements, elem.String())
	}
	out.WriteString(strings.Join(elements, ", "))

	out.WriteString("]")

	return out.String()
}

// ============================================================================
// IndexExpression
// ============================================================================

// IndexExpression represents an array/string indexing operation.
// Examples:
//   - arr[i]      // simple indexing
//   - arr[0]      // literal index
//   - arr[i + 1]  // expression index
//   - arr[i][j]   // nested indexing
type IndexExpression struct {
	Left  Expression
	Index Expression
	TypedExpressionBase
}

// expressionNode implements the Expression interface
func (ie *IndexExpression) expressionNode() {}

// Pos returns the position from the Left expression
func (ie *IndexExpression) Pos() token.Position {
	return ie.Left.Pos()
}

// String returns a string representation of the index expression
func (ie *IndexExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("]")
	out.WriteString(")")

	return out.String()
}

// ============================================================================
// NewArrayExpression
// ============================================================================

// NewArrayExpression represents dynamic array instantiation with the 'new' keyword.
// This is distinct from NewExpression (class instantiation) and ArrayLiteralExpression (literal values).
//
// DWScript syntax:
//   - new Integer[16]                      // 1D array with 16 elements
//   - new Integer[10, 20]                  // 2D array (10x20)
//   - new String[Length(s)+1]              // Size from expression
//   - new Integer[aScale*12+1, aScale*12+1] // 2D with computed sizes
//
// Related AST nodes:
//   - NewExpression: for class instantiation (new ClassName(args))
//   - ArrayLiteralExpression: for literal array values ([1, 2, 3])
//   - ArrayTypeAnnotation: for array type declarations
type NewArrayExpression struct {
	ElementTypeName *Identifier
	Dimensions      []Expression
	TypedExpressionBase
}

// expressionNode implements the Expression interface
func (nae *NewArrayExpression) expressionNode() {}

// String returns a string representation of the new array expression
// Examples:
//   - "new Integer[16]"
//   - "new String[10, 20]"
//   - "new Float[Length(arr)+1]"
func (nae *NewArrayExpression) String() string {
	var out bytes.Buffer

	out.WriteString("new ")
	if nae.ElementTypeName != nil {
		out.WriteString(nae.ElementTypeName.Value)
	}
	out.WriteString("[")

	dimStrings := []string{}
	for _, dim := range nae.Dimensions {
		dimStrings = append(dimStrings, dim.String())
	}
	out.WriteString(strings.Join(dimStrings, ", "))

	out.WriteString("]")

	return out.String()
}
