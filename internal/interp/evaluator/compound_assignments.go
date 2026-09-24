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
	if obj, prop, indices, handled, err := e.resolveInterfaceIndexedProperty(indexExpr, ctx); handled {
		if err != nil {
			return err
		}
		current := e.readInterfaceIndexedProperty(obj, prop, indices, indexExpr, ctx)
		if isError(current) || ctx.Exception() != nil {
			return current
		}
		right := e.Eval(stmt.Value, ctx)
		if isError(right) || ctx.Exception() != nil {
			return right
		}
		result := e.applyCompoundOperation(stmt.Operator, current, right, stmt, ctx)
		if isError(result) || ctx.Exception() != nil {
			return result
		}
		return e.writeInterfaceIndexedProperty(obj, prop, indices, result, stmt, ctx)
	}
	if e.interfacePropertyResultIndex(indexExpr, ctx) {
		return e.evalCompoundPropertyResultIndex(indexExpr, stmt, ctx)
	}
	// A member-rooted index can name an indexed property, whose indices are
	// arguments to its accessors. Capture its receiver before choosing that path.
	if base, _ := CollectIndices(indexExpr); base != nil {
		if member, ok := base.(*ast.MemberAccessExpression); ok {
			return e.evalCompoundMemberRootedIndex(member, indexExpr, stmt, ctx)
		}
	}
	return e.evalCompoundPropertyResultIndex(indexExpr, stmt, ctx)
}

// evalCompoundMemberRootedIndex captures a member receiver once. A member can
// be either an indexed property or an ordinary array-valued field/property;
// both must keep the same receiver and index values across read and write.
func (e *Evaluator) evalCompoundMemberRootedIndex(member *ast.MemberAccessExpression, indexExpr *ast.IndexExpression, stmt *ast.AssignmentStatement, ctx *ExecutionContext) Value {
	obj := e.normalizeMemberReceiver(e.Eval(member.Object, ctx), member.Object, member, ctx)
	if isError(obj) || ctx.Exception() != nil {
		return obj
	}

	if result, handled := e.evalCompoundMemberIndexedProperty(obj, member, indexExpr, stmt, ctx); handled {
		return result
	}

	container := e.readResolvedMember(member, obj, ctx)
	if isError(container) || ctx.Exception() != nil {
		return container
	}
	// For an ordinary member, each index is a separate array/associative
	// access. Vivify intermediate associative slots just as a plain lvalue does.
	indices := collectIndexNodes(indexExpr)
	for _, idx := range indices[:len(indices)-1] {
		key := e.Eval(idx.Index, ctx)
		if isError(key) || ctx.Exception() != nil {
			return key
		}
		if assoc, ok := derefAssociative(container); ok {
			container = e.vivifyAssociativeSlot(assoc, key, ctx)
		} else {
			container = e.readResolvedIndex(container, key, idx, ctx)
		}
		if isError(container) || ctx.Exception() != nil {
			return container
		}
	}
	return e.evalCompoundCapturedIndex(container, indices[len(indices)-1], stmt, ctx)
}

func (e *Evaluator) evalCompoundMemberIndexedProperty(obj Value, member *ast.MemberAccessExpression, indexExpr *ast.IndexExpression, stmt *ast.AssignmentStatement, ctx *ExecutionContext) (Value, bool) {
	if accessor, ok := obj.(runtime.PropertyAccessor); ok {
		if prop := accessor.LookupProperty(member.Member.Value); prop != nil {
			_, isRecord := obj.(RecordInstanceValue)
			if prop.IsIndexed || isRecord {
				return e.evalCompoundNamedIndexedProperty(obj, prop, indexExpr, stmt, ctx), true
			}
		}
	}
	if class, ok := obj.(ClassMetaValue); ok {
		if info := class.GetClassInfo(); info != nil {
			if prop := info.LookupProperty(member.Member.Value); prop != nil && prop.IsIndexed {
				return e.evalCompoundClassIndexedProperty(obj, class, member.Member.Value, indexExpr, stmt, ctx), true
			}
		}
	}
	return nil, false
}

func collectIndexNodes(expr *ast.IndexExpression) []*ast.IndexExpression {
	var reversed []*ast.IndexExpression
	for current := expr; current != nil; {
		reversed = append(reversed, current)
		next, ok := current.Left.(*ast.IndexExpression)
		if !ok {
			break
		}
		current = next
	}
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	return reversed
}

func (e *Evaluator) evalCompoundCapturedIndex(container Value, indexExpr *ast.IndexExpression, stmt *ast.AssignmentStatement, ctx *ExecutionContext) Value {
	index := e.Eval(indexExpr.Index, ctx)
	if isError(index) || ctx.Exception() != nil {
		return index
	}
	current := e.readResolvedIndex(container, index, indexExpr, ctx)
	if isError(current) || ctx.Exception() != nil {
		return current
	}
	right := e.Eval(stmt.Value, ctx)
	if isError(right) || ctx.Exception() != nil {
		return right
	}
	result := e.applyCompoundOperation(stmt.Operator, current, right, stmt, ctx)
	if isError(result) || ctx.Exception() != nil {
		return result
	}
	return e.assignResolvedIndex(container, index, result, stmt, ctx)
}

func (e *Evaluator) evalCompoundNamedIndexedProperty(obj Value, prop *runtime.PropertyDescriptor, indexExpr *ast.IndexExpression, stmt *ast.AssignmentStatement, ctx *ExecutionContext) Value {
	_, expressions := CollectIndices(indexExpr)
	indices := make([]Value, len(expressions))
	for i, expr := range expressions {
		indices[i] = e.Eval(expr, ctx)
		if isError(indices[i]) || ctx.Exception() != nil {
			return indices[i]
		}
	}
	var current Value
	if record, ok := obj.(RecordInstanceValue); ok {
		current = record.ReadIndexedProperty(prop.Impl, indices, func(pi any, idx []Value) Value {
			return e.executeRecordIndexedPropertyRead(obj, pi, idx, indexExpr, ctx)
		})
	} else if object, ok := obj.(ObjectValue); ok {
		current = object.ReadIndexedProperty(prop.Impl, indices, func(pi any, idx []Value) Value {
			return e.executeIndexedPropertyRead(obj, pi, idx, indexExpr, ctx)
		})
	} else {
		return e.newError(indexExpr, "cannot read indexed property from non-object value")
	}
	if isError(current) || ctx.Exception() != nil {
		return current
	}
	right := e.Eval(stmt.Value, ctx)
	if isError(right) || ctx.Exception() != nil {
		return right
	}
	result := e.applyCompoundOperation(stmt.Operator, current, right, stmt, ctx)
	if isError(result) || ctx.Exception() != nil {
		return result
	}
	return e.evalIndexedPropertyAssignmentDescriptor(obj, prop, indices, result, stmt, ctx)
}

func (e *Evaluator) evalCompoundClassIndexedProperty(obj Value, class ClassMetaValue, memberName string, indexExpr *ast.IndexExpression, stmt *ast.AssignmentStatement, ctx *ExecutionContext) Value {
	_, expressions := CollectIndices(indexExpr)
	indices := make([]Value, len(expressions))
	for i, expr := range expressions {
		indices[i] = e.Eval(expr, ctx)
		if isError(indices[i]) || ctx.Exception() != nil {
			return indices[i]
		}
	}
	current, _ := e.evalClassMetaIndexedPropertyValues(obj, class, memberName, indices, indexExpr, ctx)
	if isError(current) || ctx.Exception() != nil {
		return current
	}
	right := e.Eval(stmt.Value, ctx)
	if isError(right) || ctx.Exception() != nil {
		return right
	}
	result := e.applyCompoundOperation(stmt.Operator, current, right, stmt, ctx)
	if isError(result) || ctx.Exception() != nil {
		return result
	}
	written, _ := e.evalClassMetaIndexedPropertyWriteValues(obj, class, memberName, indices, result, stmt, ctx)
	return written
}

// evalCompoundPropertyResultIndex captures the returned container and final
// index so the accessor and all index expressions execute only once.
func (e *Evaluator) evalCompoundPropertyResultIndex(indexExpr *ast.IndexExpression, stmt *ast.AssignmentStatement, ctx *ExecutionContext) Value {
	container := e.resolveLValueContainer(indexExpr.Left, ctx)
	if isError(container) || ctx.Exception() != nil {
		return container
	}
	return e.evalCompoundCapturedIndex(container, indexExpr, stmt, ctx)
}
