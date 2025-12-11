package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Compound Assignment Helpers
// ============================================================================
//
// This file contains helpers for compound member and index assignments.
// These were previously delegated to adapter.EvalNode() but are now
// implemented directly in the evaluator using read-modify-write pattern.
// ============================================================================

// evalCompoundMemberAssignment handles compound assignment to member access.
// Example: obj.field += value
//
// Pattern: Read current value → apply operation → write back
func (e *Evaluator) evalCompoundMemberAssignment(
	memberAccess *ast.MemberAccessExpression,
	stmt *ast.AssignmentStatement,
	ctx *ExecutionContext,
) Value {
	// Read current value via member access
	currentValue := e.VisitMemberAccessExpression(memberAccess, ctx)
	if isError(currentValue) {
		return currentValue
	}

	// Check for exception during read
	if ctx.Exception() != nil {
		return &runtime.NilValue{}
	}

	// Evaluate RHS
	rightValue := e.Eval(stmt.Value, ctx)
	if isError(rightValue) {
		return rightValue
	}

	// Check for exception during RHS evaluation
	if ctx.Exception() != nil {
		return &runtime.NilValue{}
	}

	// Apply compound operation
	result := e.applyCompoundOperation(stmt.Operator, currentValue, rightValue, stmt)
	if isError(result) {
		return result
	}

	// Write back via member assignment
	return e.evalMemberAssignmentDirect(memberAccess, result, stmt, ctx)
}

// evalCompoundIndexAssignment handles compound assignment to indexed access.
// Example: arr[i] += value
//
// Pattern: Read current value → apply operation → write back
func (e *Evaluator) evalCompoundIndexAssignment(
	indexExpr *ast.IndexExpression,
	stmt *ast.AssignmentStatement,
	ctx *ExecutionContext,
) Value {
	// Read current value via index access
	currentValue := e.VisitIndexExpression(indexExpr, ctx)
	if isError(currentValue) {
		return currentValue
	}

	// Check for exception during read
	if ctx.Exception() != nil {
		return &runtime.NilValue{}
	}

	// Evaluate RHS
	rightValue := e.Eval(stmt.Value, ctx)
	if isError(rightValue) {
		return rightValue
	}

	// Check for exception during RHS evaluation
	if ctx.Exception() != nil {
		return &runtime.NilValue{}
	}

	// Apply compound operation
	result := e.applyCompoundOperation(stmt.Operator, currentValue, rightValue, stmt)
	if isError(result) {
		return result
	}

	// Write back via index assignment
	return e.evalIndexAssignmentDirect(indexExpr, result, stmt, ctx)
}
