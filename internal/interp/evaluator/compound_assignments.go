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
	// Capture the receiver before the read and RHS. The RHS may replace the
	// variable holding it, and a constructor must not run again for the write.
	obj, setter, err := e.evaluateMemberAssignmentContainer(memberAccess.Object, ctx)
	if err != nil {
		return e.newError(stmt, "%s", err.Error())
	}
	obj = e.normalizeMemberReceiver(obj, memberAccess.Object, memberAccess, ctx)
	if isError(obj) {
		return obj
	}
	if ctx.Exception() != nil {
		return &runtime.NilValue{}
	}
	currentValue := e.readResolvedMember(memberAccess, obj, ctx)
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
	result := e.applyCompoundOperation(stmt.Operator, currentValue, rightValue, stmt, ctx)
	if isError(result) {
		return result
	}
	if ctx.Exception() != nil {
		return &runtime.NilValue{}
	}

	// Member storage owns its copy of a record result, as in simple assignment.
	if record, ok := result.(*runtime.RecordValue); ok {
		result = record.Copy()
	}
	return e.assignResolvedMember(memberAccess, result, stmt, obj, setter, ctx)
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
	result := e.applyCompoundOperation(stmt.Operator, currentValue, rightValue, stmt, ctx)
	if isError(result) {
		return result
	}

	// Write back via index assignment
	return e.evalIndexAssignmentDirect(indexExpr, result, stmt, ctx)
}
