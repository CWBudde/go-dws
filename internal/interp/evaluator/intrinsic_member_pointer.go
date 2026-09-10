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

// bindClassMethodPointer builds a pointer to a user-declared class method on a
// class reference. It also covers the parameterless case, which the
// CreateClassMethodPointer factories decline because the auto-invoke path
// normally owns it: in a pointer context the declared method must still win
// over an intrinsic member of the same name.
func (e *Evaluator) bindClassMethodPointer(
	classMeta ClassMetaValue,
	memberName string,
	receiver Value,
	ctx *ExecutionContext,
) (Value, bool) {
	if classMeta == nil || !classMeta.HasClassMethod(memberName) {
		return nil, false
	}
	bind := func(methodDecl *runtime.MethodMetadata) Value {
		return e.createFunctionPointerFromDecl(methodDecl, receiver, ctx)
	}
	if ptr, created := classMeta.CreateClassMethodPointer(memberName, bind); created {
		return ptr, true
	}
	// Reuse the parameterless lookup, binding the declaration instead of
	// executing it (nested classes decline parameterless methods above).
	return classMeta.InvokeParameterlessClassMethod(memberName, bind)
}

// classMetaMemberPointer resolves `TClass.Member` in a pointer context: a
// user-declared class method owns the name, otherwise the intrinsic member
// (ClassName / ClassType) is captured.
func (e *Evaluator) classMetaMemberPointer(
	classMeta ClassMetaValue,
	memberName string,
	receiver Value,
	node ast.Expression,
	ctx *ExecutionContext,
) Value {
	if ptr, ok := e.bindClassMethodPointer(classMeta, memberName, receiver, ctx); ok {
		return ptr
	}
	return e.newIntrinsicClassMemberPointer(receiver, memberName, node, ctx)
}

// instanceMemberPointer resolves `obj.Member` in a pointer context for one of
// the intrinsic member names. An instance method wins, then a class method
// reached through the instance (bound to the class reference, as elsewhere),
// and only then the intrinsic itself.
func (e *Evaluator) instanceMemberPointer(
	objVal ObjectValue,
	receiver Value,
	memberName string,
	node ast.Expression,
	ctx *ExecutionContext,
) Value {
	if methodDecl := objVal.GetMethodDecl(memberName); methodDecl != nil {
		return e.createFunctionPointerFromDecl(methodDecl, receiver, ctx)
	}
	if methodDecl := objVal.GetClassMethodDecl(memberName); methodDecl != nil {
		return e.createFunctionPointerFromDecl(methodDecl, e.classSelfForInstance(objVal, receiver), ctx)
	}
	return e.newIntrinsicClassMemberPointer(receiver, memberName, node, ctx)
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
