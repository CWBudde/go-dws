// Package ast defines the Abstract Syntax Tree node types for DWScript.
// This file contains AST nodes for set type declarations and set literals (Stage 8).
package ast

import (
	"bytes"
	"strings"
)

// ============================================================================
// Set Declaration
// ============================================================================

// SetDecl represents a set type declaration.
// Examples:
//   - type TDays = set of TWeekday;
//   - var s: set of (Mon, Tue, Wed);  // inline with anonymous enum
type SetDecl struct {
	BaseNode
	Name        *Identifier
	ElementType *TypeAnnotation
}

// statementNode implements the Statement interface
func (sd *SetDecl) statementNode() {}

// String returns a string representation of the set declaration
func (sd *SetDecl) String() string {
	var out bytes.Buffer

	if sd.Name != nil {
		out.WriteString("type ")
		out.WriteString(sd.Name.Value)
		out.WriteString(" = ")
	}

	out.WriteString("set of ")
	if sd.ElementType != nil {
		out.WriteString(sd.ElementType.String())
	}

	return out.String()
}

// ============================================================================
// Set Literal Expression
// ============================================================================

// SetLiteral represents a set literal expression.
// Examples:
//   - [one, two, three]        // set with elements
//   - [one..five]              // set with range (not yet implemented)
//   - []                       // empty set
type SetLiteral struct {
	TypedExpressionBase
	Elements []Expression
}

// expressionNode implements the Expression interface
func (sl *SetLiteral) expressionNode() {}

// String returns a string representation of the set literal
func (sl *SetLiteral) String() string {
	var out bytes.Buffer

	out.WriteString("[")

	elements := []string{}
	for _, elem := range sl.Elements {
		elements = append(elements, elem.String())
	}
	out.WriteString(strings.Join(elements, ", "))

	out.WriteString("]")

	return out.String()
}
