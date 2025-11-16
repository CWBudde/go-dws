package evaluator

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// This file contains visitor methods for literal AST nodes.
// Phase 3.5.2: Visitor pattern implementation for literals.
//
// Literals are the simplest nodes - they directly create runtime values
// without needing access to the interpreter state.

// VisitIntegerLiteral evaluates an integer literal node.
func (e *Evaluator) VisitIntegerLiteral(node *ast.IntegerLiteral, ctx *ExecutionContext) Value {
	// Phase 3.5.4.1: Direct creation of IntegerValue from literal
	// Integer literals are the simplest case - just wrap the parsed value
	return &runtime.IntegerValue{Value: node.Value}
}

// VisitFloatLiteral evaluates a float literal node.
func (e *Evaluator) VisitFloatLiteral(node *ast.FloatLiteral, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Direct creation
	return e.adapter.EvalNode(node)
}

// VisitStringLiteral evaluates a string literal node.
func (e *Evaluator) VisitStringLiteral(node *ast.StringLiteral, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Direct creation
	return e.adapter.EvalNode(node)
}

// VisitBooleanLiteral evaluates a boolean literal node.
func (e *Evaluator) VisitBooleanLiteral(node *ast.BooleanLiteral, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Direct creation
	return e.adapter.EvalNode(node)
}

// VisitCharLiteral evaluates a character literal node.
// In DWScript, character literals are treated as single-character strings.
func (e *Evaluator) VisitCharLiteral(node *ast.CharLiteral, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Convert char to string value
	return e.adapter.EvalNode(node)
}

// VisitNilLiteral evaluates a nil literal node.
func (e *Evaluator) VisitNilLiteral(node *ast.NilLiteral, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Return nil value
	return e.adapter.EvalNode(node)
}
