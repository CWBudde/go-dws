package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
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
	if obj, ok := receiver.(ObjectValue); ok && !obj.HasProperty(member.Member.Value) && obj.GetField(member.Member.Value) != nil {
		fieldName := member.Member.Value
		getter := func() (runtime.Value, error) {
			value := obj.GetField(fieldName)
			if value == nil {
				return nil, fmt.Errorf("field '%s' not found in class '%s'", fieldName, obj.ClassName())
			}
			return value, nil
		}
		return runtime.NewReferenceValue(member.String(), getter, func(value runtime.Value) error {
			return assign(value)
		}), nil
	}
	return newAssignedReference(member.String(), current, assign), nil
}
