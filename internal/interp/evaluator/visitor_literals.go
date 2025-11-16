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
	// Phase 3.5.4.2: Direct creation of FloatValue from literal
	return &runtime.FloatValue{Value: node.Value}
}

// VisitStringLiteral evaluates a string literal node.
func (e *Evaluator) VisitStringLiteral(node *ast.StringLiteral, ctx *ExecutionContext) Value {
	// Phase 3.5.4.3: Direct creation of StringValue from literal
	return &runtime.StringValue{Value: node.Value}
}

// VisitBooleanLiteral evaluates a boolean literal node.
func (e *Evaluator) VisitBooleanLiteral(node *ast.BooleanLiteral, ctx *ExecutionContext) Value {
	// Phase 3.5.4.4: Direct creation of BooleanValue from literal
	return &runtime.BooleanValue{Value: node.Value}
}

// VisitCharLiteral evaluates a character literal node.
// In DWScript, character literals are treated as single-character strings.
func (e *Evaluator) VisitCharLiteral(node *ast.CharLiteral, ctx *ExecutionContext) Value {
	// Phase 3.5.4.5: Convert char rune to string value
	// Character literals in DWScript are syntactic sugar for single-char strings
	return &runtime.StringValue{Value: string(node.Value)}
}

// VisitNilLiteral evaluates a nil literal node.
func (e *Evaluator) VisitNilLiteral(node *ast.NilLiteral, ctx *ExecutionContext) Value {
	// Phase 3.5.4.6: Return new nil value instance
	// NilValue can have an associated ClassType, but literals don't specify one
	return &runtime.NilValue{}
}
