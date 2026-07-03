// Package ast defines the Abstract Syntax Tree node types for DWScript.
// This file contains AST support for generic types.
package ast

import (
	"strings"

	"github.com/cwbudde/go-dws/pkg/token"
)

// GenericTypeRef represents a reference to a generic type used in expression
// position, such as `TTest<Integer>` in `TTest<Integer>.Identity(10)`. It is
// produced by the parser when it recognizes the generic-instantiation pattern
// and is rewritten to a plain Identifier (carrying the mangled specialized
// name) during monomorphization, so downstream phases never see it.
type GenericTypeRef struct {
	// Base is the generic type name (e.g. "TTest").
	Base *Identifier
	// TypeArgs holds the concrete type arguments (e.g. Integer, String).
	TypeArgs []TypeExpression
	BaseNode
}

func (g *GenericTypeRef) expressionNode() {}

// TokenLiteral returns the base type name's token literal.
func (g *GenericTypeRef) TokenLiteral() string {
	if g.Base != nil {
		return g.Base.TokenLiteral()
	}
	return g.Token.Literal
}

// Pos returns the position of the base type name.
func (g *GenericTypeRef) Pos() token.Position {
	if g.Base != nil {
		return g.Base.Pos()
	}
	return g.Token.Pos
}

// String renders the generic type reference (e.g. "TTest<Integer,String>").
func (g *GenericTypeRef) String() string {
	base := ""
	if g.Base != nil {
		base = g.Base.Value
	}
	return MangleGenericName(base, g.TypeArgs)
}

// MangleGenericName produces the canonical specialized type name for a generic
// instantiation, e.g. MangleGenericName("TTest", [Integer, String]) yields
// "TTest<Integer,String>". The format (angle brackets, comma-separated, no
// spaces) matches DWScript's ClassName output for specialized generic classes.
func MangleGenericName(base string, args []TypeExpression) string {
	if len(args) == 0 {
		return base
	}
	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = typeExpressionName(arg)
	}
	return base + "<" + strings.Join(parts, ",") + ">"
}

// typeExpressionName renders a type expression to its canonical name for use in
// mangled generic names. Nested generic arguments are rendered recursively.
func typeExpressionName(expr TypeExpression) string {
	switch t := expr.(type) {
	case *TypeAnnotation:
		if len(t.TypeArgs) > 0 {
			return MangleGenericName(t.Name, t.TypeArgs)
		}
		if t.InlineType != nil && t.Name == "" {
			return typeExpressionName(t.InlineType)
		}
		return t.Name
	case nil:
		return ""
	default:
		return strings.TrimSpace(expr.String())
	}
}
