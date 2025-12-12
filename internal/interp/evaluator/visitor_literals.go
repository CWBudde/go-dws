package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// This file contains visitor methods for literal AST nodes.
// Literals directly create runtime values without interpreter state access.

// VisitIntegerLiteral evaluates an integer literal node.
func (e *Evaluator) VisitIntegerLiteral(node *ast.IntegerLiteral, ctx *ExecutionContext) Value {
	return &runtime.IntegerValue{Value: node.Value}
}

// VisitFloatLiteral evaluates a float literal node.
func (e *Evaluator) VisitFloatLiteral(node *ast.FloatLiteral, ctx *ExecutionContext) Value {
	return &runtime.FloatValue{Value: node.Value}
}

// VisitStringLiteral evaluates a string literal node.
func (e *Evaluator) VisitStringLiteral(node *ast.StringLiteral, ctx *ExecutionContext) Value {
	return &runtime.StringValue{Value: node.Value}
}

// VisitBooleanLiteral evaluates a boolean literal node.
func (e *Evaluator) VisitBooleanLiteral(node *ast.BooleanLiteral, ctx *ExecutionContext) Value {
	return &runtime.BooleanValue{Value: node.Value}
}

// VisitCharLiteral evaluates a character literal node.
// Character literals are treated as single-character strings.
func (e *Evaluator) VisitCharLiteral(node *ast.CharLiteral, ctx *ExecutionContext) Value {
	return &runtime.StringValue{Value: string(node.Value)}
}

// VisitNilLiteral evaluates a nil literal node.
func (e *Evaluator) VisitNilLiteral(node *ast.NilLiteral, ctx *ExecutionContext) Value {
	return &runtime.NilValue{}
}
