// Package ast defines the Abstract Syntax Tree node types for DWScript.
// This file contains AST nodes for function pointer types and address-of expressions (Stage 9).
package ast

import (
	"bytes"
	"strings"

	"github.com/cwbudde/go-dws/pkg/token"
)

// FunctionPointerTypeNode represents a function or procedure pointer type annotation in DWScript.
// Function pointers can be stored in variables, passed as parameters, and called indirectly.
//
// Examples:
//   - type TComparator = function(a, b: Integer): Integer;
//   - type TCallback = procedure(msg: String);
//   - type TNotifyEvent = procedure(Sender: TObject) of object;
//
// This is a type annotation node that can appear in type declarations, variable declarations,
// and parameter lists.
type FunctionPointerTypeNode struct {
	Parameters []*Parameter    // Parameter list
	ReturnType *TypeAnnotation // Return type (nil for procedures)
	Token      token.Token     // The 'function' or 'procedure' token
	OfObject   bool            // True for method pointers (procedure/function of object)
	EndPos     token.Position
}

// String returns a string representation of the function pointer type.
// Examples:
//   - "function(Integer, String): Boolean"
//   - "procedure(Integer)"
//   - "function(Integer): Boolean of object"
func (fpt *FunctionPointerTypeNode) String() string {
	var out bytes.Buffer

	// Write function or procedure keyword
	if fpt.ReturnType == nil {
		out.WriteString("procedure")
	} else {
		out.WriteString("function")
	}

	// Write parameters
	out.WriteString("(")
	paramStrs := []string{}
	for _, param := range fpt.Parameters {
		paramStrs = append(paramStrs, param.String())
	}
	// Note: DWScript uses semicolons to separate parameters, not commas
	out.WriteString(strings.Join(paramStrs, "; "))
	out.WriteString(")")

	// Write return type for functions
	if fpt.ReturnType != nil {
		out.WriteString(": ")
		out.WriteString(fpt.ReturnType.String())
	}

	// Write "of object" for method pointers
	if fpt.OfObject {
		out.WriteString(" of object")
	}

	return out.String()
}

// TokenLiteral returns the literal value of the token.
func (fpt *FunctionPointerTypeNode) TokenLiteral() string {
	return fpt.Token.Literal
}

// Pos returns the position of the node in the source code.
func (fpt *FunctionPointerTypeNode) Pos() token.Position {
	return fpt.Token.Pos
}

// End returns the end position of the node in the source code.
func (fpt *FunctionPointerTypeNode) End() token.Position {
	if fpt.EndPos.Line != 0 {
		return fpt.EndPos
	}
	// Try to calculate from return type
	if fpt.ReturnType != nil {
		return fpt.ReturnType.End()
	}
	return fpt.Token.Pos
}

// typeExpressionNode marks this as a type expression
func (fpt *FunctionPointerTypeNode) typeExpressionNode() {}

// AddressOfExpression represents the address-of operator (@) applied to a function or procedure.
// This is used to get a function pointer from a function/procedure name.
//
// Examples:
//   - @Ascending
//   - @MyCallback
//   - @TMyClass.MyMethod
//
// The @ operator takes a function/procedure identifier and produces a function pointer value
// that can be assigned to variables, passed as parameters, or stored in data structures.
type AddressOfExpression struct {
	Operator Expression      // The target function/procedure (usually an Identifier or member access)
	Type     *TypeAnnotation // Type of the resulting function pointer
	Token    token.Token     // The @ token
	EndPos   token.Position
}

func (ao *AddressOfExpression) expressionNode()      {}
func (ao *AddressOfExpression) TokenLiteral() string { return ao.Token.Literal }
func (ao *AddressOfExpression) Pos() token.Position  { return ao.Token.Pos }
func (ao *AddressOfExpression) End() token.Position {
	if ao.EndPos.Line != 0 {
		return ao.EndPos
	}
	if ao.Operator != nil {
		return ao.Operator.End()
	}
	return ao.Token.Pos
}
func (ao *AddressOfExpression) GetType() *TypeAnnotation    { return ao.Type }
func (ao *AddressOfExpression) SetType(typ *TypeAnnotation) { ao.Type = typ }

// String returns a string representation of the address-of expression.
func (ao *AddressOfExpression) String() string {
	var out bytes.Buffer
	out.WriteString("@")
	out.WriteString(ao.Operator.String())
	return out.String()
}
