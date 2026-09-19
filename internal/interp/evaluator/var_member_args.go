package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// prepareCapturedMemberArgument retains one receiver for a member var argument.
// Plain object fields remain live: writes through another alias or a later
// argument must be visible when the callee reads its var parameter.
func (e *Evaluator) prepareCapturedMemberArgument(member *ast.MemberAccessExpression, ctx *ExecutionContext) (Value, error) {
	receiver := e.resolveLValueContainer(member.Object, ctx)
	if err := argumentEvaluationError(receiver, ctx); err != nil {
		return nil, err
	}
	if ref, ok := receiver.(ReferenceAccessor); ok {
		value, err := ref.Dereference()
		if err != nil {
			return nil, err
		}
		receiver = value
	}
	current, assign, err := e.evaluateResolvedMemberTarget(member, receiver, ctx, false)
	if err != nil {
		return nil, fmt.Errorf("var parameter requires a variable, got %T", member)
	}
	if err := argumentEvaluationError(current, ctx); err != nil {
		return nil, err
	}
	if ref, ok := current.(ReferenceAccessor); ok {
		return ref, nil
	}
	if obj, ok := receiver.(ObjectValue); ok && !obj.HasProperty(member.Member.Value) && objectFieldDeclared(obj, member.Member.Value) {
		fieldName := member.Member.Value
		getter := func() (runtime.Value, error) {
			// A declared field that was never set reads as nil.
			if value := obj.GetField(fieldName); value != nil {
				return value, nil
			}
			return &runtime.NilValue{}, nil
		}
		return runtime.NewReferenceValue(member.String(), getter, func(value runtime.Value) error {
			return assign(value)
		}), nil
	}
	return newAssignedReference(member.String(), current, assign), nil
}

// objectFieldDeclared reports whether obj has a plain field of the given name,
// using class metadata so a declared but not yet initialized field counts.
func objectFieldDeclared(obj ObjectValue, fieldName string) bool {
	if obj.GetField(fieldName) != nil {
		return true
	}
	inst, ok := obj.(*runtime.ObjectInstance)
	return ok && inst.Class != nil && inst.Class.FieldExists(ident.Normalize(fieldName))
}
