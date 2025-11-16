// Package ast defines the Abstract Syntax Tree node types for DWScript.
// This file contains AST nodes for operator declarations (Stage 8).
package ast

import (
	"bytes"
	"strings"

	"github.com/cwbudde/go-dws/pkg/token"
)

// OperatorKind identifies the category of an operator declaration.
// Stage 8 requires tracking global operators, class operators, and conversions.
type OperatorKind int

const (
	// OperatorKindGlobal represents a standalone operator declaration.
	OperatorKindGlobal OperatorKind = iota
	// OperatorKindClass represents a class operator declared within a class body.
	OperatorKindClass
	// OperatorKindConversion represents an implicit/explicit conversion operator.
	OperatorKindConversion
)

// String returns a human-readable representation of the operator kind.
func (k OperatorKind) String() string {
	switch k {
	case OperatorKindGlobal:
		return "global"
	case OperatorKindClass:
		return "class"
	case OperatorKindConversion:
		return "conversion"
	default:
		return "unknown"
	}
}

// OperatorDecl represents an operator declaration in DWScript.
//
// Examples:
//
//	operator + (String, Integer) : String uses StrPlusInt;
//	operator implicit (Integer) : String uses IntToStr;
//	class operator += String uses AppendString;
type OperatorDecl struct {
	ReturnType     TypeExpression
	Binding        *Identifier
	OperatorSymbol string
	OperandTypes   []TypeExpression
	OperatorToken  token.Token
	BaseNode
	Kind       OperatorKind
	Arity      int
	Visibility Visibility
}

func (od *OperatorDecl) statementNode() {}

// String renders the operator declaration in DWScript syntax (without trailing semicolon).
func (od *OperatorDecl) String() string {
	var out bytes.Buffer

	if od.Kind == OperatorKindClass {
		out.WriteString("class ")
	}

	out.WriteString("operator ")

	if od.OperatorSymbol != "" {
		out.WriteString(od.OperatorSymbol)
		out.WriteString(" ")
	}

	// Render operand list
	if len(od.OperandTypes) > 0 {
		if od.Kind == OperatorKindClass && !od.hasParenthesizedOperands() {
			// Class operators in DWScript typically omit parentheses around operand types.
			operands := []string{}
			for _, operand := range od.OperandTypes {
				operands = append(operands, operand.String())
			}
			out.WriteString(strings.Join(operands, ", "))
			out.WriteString(" ")
		} else {
			operands := []string{}
			for _, operand := range od.OperandTypes {
				operands = append(operands, operand.String())
			}
			out.WriteString("(")
			out.WriteString(strings.Join(operands, ", "))
			out.WriteString(")")
			out.WriteString(" ")
		}
	}

	// Render return type (if any)
	if od.ReturnType != nil && od.ReturnType.String() != "" {
		out.WriteString(": ")
		out.WriteString(od.ReturnType.String())
		out.WriteString(" ")
	}

	// Render binding
	if od.Binding != nil {
		out.WriteString("uses ")
		out.WriteString(od.Binding.String())
	}

	return strings.TrimSpace(out.String())
}

// hasParenthesizedOperands determines whether the operator should render operands in parentheses.
// Global and conversion operators always use parentheses; class operators omit them by default.
func (od *OperatorDecl) hasParenthesizedOperands() bool {
	return od.Kind != OperatorKindClass
}
