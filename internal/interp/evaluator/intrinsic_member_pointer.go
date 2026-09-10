package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// TObject's intrinsic, parameterless members. They have no declaration behind
// them, so a pointer to one is dispatched by name against its bound receiver.
const (
	intrinsicClassName = "ClassName"
	intrinsicClassType = "ClassType"
)

// isIntrinsicClassMemberName reports whether name is one of TObject's intrinsic
// parameterless members that may be captured as a pointer.
func isIntrinsicClassMemberName(name string) bool {
	return ident.Equal(name, intrinsicClassName) || ident.Equal(name, intrinsicClassType)
}

// newIntrinsicClassMemberPointer captures ClassName / ClassType on a class
// reference or an object instance as a parameterless pointer, deferring the
// lookup until the pointer is invoked (`@TObject.ClassType`, or
// `a.Add(TObject.ClassName)` where the element type is `function : String`).
//
// The analyzer's resolved signature is adopted when present so a call through
// the pointer keeps the declared result type; the fallback is a bare
// parameterless signature, which is all the runtime dispatch needs.
func (e *Evaluator) newIntrinsicClassMemberPointer(receiver Value, memberName string, node ast.Expression, ctx *ExecutionContext) Value {
	var pointerType *types.FunctionPointerType
	switch resolved := e.resolvedExpressionType(node, ctx).(type) {
	case *types.MethodPointerType:
		pointerType = &resolved.FunctionPointerType
	case *types.FunctionPointerType:
		pointerType = resolved
	}
	if pointerType == nil || len(pointerType.Parameters) != 0 {
		pointerType = types.NewFunctionPointerType(nil, nil)
	}

	return &runtime.FunctionPointerValue{
		SelfObject:      receiver,
		IntrinsicMember: memberName,
		PointerType:     pointerType,
	}
}

// invokeIntrinsicClassMember evaluates a pointer captured by
// newIntrinsicClassMemberPointer against its bound receiver.
func (e *Evaluator) invokeIntrinsicClassMember(ptr *runtime.FunctionPointerValue, args []Value, node ast.Node) Value {
	if len(args) != 0 {
		return e.newError(node, "wrong number of arguments for %s: expected 0, got %d", ptr.IntrinsicMember, len(args))
	}
	if ptr.SelfObject == nil {
		return e.newError(node, "%s pointer has no bound receiver", ptr.IntrinsicMember)
	}

	if classMeta, ok := ptr.SelfObject.(ClassMetaValue); ok {
		if ident.Equal(ptr.IntrinsicMember, intrinsicClassName) {
			return &runtime.StringValue{Value: classMeta.GetClassName()}
		}
		// A class reference is its own ClassType.
		return ptr.SelfObject
	}

	if objVal, ok := ptr.SelfObject.(ObjectValue); ok {
		if ident.Equal(ptr.IntrinsicMember, intrinsicClassName) {
			return &runtime.StringValue{Value: objVal.ClassName()}
		}
		classVal, err := e.typeSystem.CreateClassValue(objVal.ClassName())
		if err != nil {
			return e.newError(node, "%s", err.Error())
		}
		return classVal
	}

	return e.newError(node, "%s is not available for %s", ptr.IntrinsicMember, ptr.SelfObject.Type())
}
