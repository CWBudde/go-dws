// Package ast defines the Abstract Syntax Tree node types for DWScript.
// This file contains AST nodes for array type annotations, array literals, and array indexing.
package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/lexer"
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
	Token     lexer.Token
}

// statementNode implements the Statement interface
func (ad *ArrayDecl) statementNode() {}

// TokenLiteral returns the literal value of the token
func (ad *ArrayDecl) TokenLiteral() string {
	return ad.Token.Literal
}

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

// Pos returns the position of the array declaration in the source code
func (ad *ArrayDecl) Pos() lexer.Position {
	return ad.Token.Pos
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
	ElementType *TypeAnnotation
	LowBound    *int
	HighBound   *int
	Token       lexer.Token
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
		out.WriteString(fmt.Sprintf("[%d..%d]", *ata.LowBound, *ata.HighBound))
	}

	out.WriteString(" of ")
	if ata.ElementType != nil {
		out.WriteString(ata.ElementType.String())
	}

	return out.String()
}

// Pos returns the position of the array type annotation in the source code
func (ata *ArrayTypeAnnotation) Pos() lexer.Position {
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
// ArrayLiteral
// ============================================================================

// ArrayLiteral represents an array literal expression.
// Examples:
//   - [1, 2, 3]        // array with elements
//   - []               // empty array
//   - ['a', 'b', 'c']  // array of strings
type ArrayLiteral struct {
	Type     *TypeAnnotation
	Elements []Expression
	Token    lexer.Token
}

// expressionNode implements the Expression interface
func (al *ArrayLiteral) expressionNode() {}

// TokenLiteral returns the literal value of the token
func (al *ArrayLiteral) TokenLiteral() string {
	return al.Token.Literal
}

// String returns a string representation of the array literal
func (al *ArrayLiteral) String() string {
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

// Pos returns the position of the array literal in the source code
func (al *ArrayLiteral) Pos() lexer.Position {
	return al.Token.Pos
}

// GetType returns the inferred type annotation
func (al *ArrayLiteral) GetType() *TypeAnnotation {
	return al.Type
}

// SetType sets the type annotation
func (al *ArrayLiteral) SetType(typ *TypeAnnotation) {
	al.Type = typ
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
	Type  *TypeAnnotation
	Token lexer.Token
}

// expressionNode implements the Expression interface
func (ie *IndexExpression) expressionNode() {}

// TokenLiteral returns the literal value of the token
func (ie *IndexExpression) TokenLiteral() string {
	return ie.Token.Literal
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

// Pos returns the position of the index expression in the source code
func (ie *IndexExpression) Pos() lexer.Position {
	return ie.Token.Pos
}

// GetType returns the inferred type annotation
func (ie *IndexExpression) GetType() *TypeAnnotation {
	return ie.Type
}

// SetType sets the type annotation
func (ie *IndexExpression) SetType(typ *TypeAnnotation) {
	ie.Type = typ
}
