package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func (e *Evaluator) executeObjectMethodDirect(
	self Value,
	methodDecl any,
	args []Value,
	node ast.Node,
	ctx *ExecutionContext,
) Value {
	method, ok := methodDecl.(*ast.FunctionDecl)
	if !ok {
		return e.newError(node, "invalid method declaration type")
	}

	classInfo := e.classInfoForMethodSelf(self)
	if classInfo == nil {
		return e.newError(node, "method execution requires class context")
	}

	ctx.PushEnv()
	defer ctx.PopEnv()

	ctx.Env().Define("Self", self)
	ctx.Env().Define("__CurrentMethod__", &runtime.StringValue{Value: method.Name.Value})
	if classVal, err := e.typeSystem.CreateClassValue(classInfo.GetName()); err == nil && classVal != nil {
		if classMeta, ok := classVal.(ClassMetaValue); ok {
			ctx.Env().Define("__CurrentClass__", classMeta)
		}
	}
	e.bindClassConstantsForMethod(classInfo, ctx)

	return e.ExecuteUserFunctionDirect(method, args, ctx)
}

func (e *Evaluator) classInfoForMethodSelf(self Value) runtime.IClassInfo {
	switch val := self.(type) {
	case *runtime.ObjectInstance:
		return val.Class
	case *runtime.InterfaceInstance:
		if val.Object != nil {
			return val.Object.Class
		}
	}

	return nil
}

func (e *Evaluator) bindClassConstantsForMethod(classInfo runtime.IClassInfo, ctx *ExecutionContext) {
	for _, cls := range classInfoHierarchy(classInfo) {
		meta := cls.GetMetadata()
		if meta == nil || meta.Constants == nil {
			continue
		}
		for name, value := range meta.Constants {
			if runtimeValue, ok := value.(Value); ok {
				ctx.Env().Define(name, runtimeValue)
			}
		}
	}
}

func (e *Evaluator) executeClassMethodDirect(
	classMeta ClassMetaValue,
	methodDecl any,
	args []Value,
	node ast.Node,
	ctx *ExecutionContext,
) Value {
	method, ok := methodDecl.(*ast.FunctionDecl)
	if !ok {
		return e.newError(node, "invalid class method declaration type")
	}

	classInfo := classMeta.GetClassInfo()
	if classInfo == nil {
		return e.newError(node, "class method execution requires class context")
	}

	classValue, ok := classMeta.(Value)
	if !ok {
		return e.newError(node, "class method execution requires runtime class value")
	}

	ctx.PushEnv()
	defer ctx.PopEnv()

	ctx.Env().Define("Self", classValue)
	ctx.Env().Define("__CurrentClass__", classValue)
	ctx.Env().Define("__CurrentMethod__", &runtime.StringValue{Value: method.Name.Value})
	e.bindClassConstantsForMethod(classInfo, ctx)

	return e.ExecuteUserFunctionDirect(method, args, ctx)
}

func currentClassMetaValue(ctx *ExecutionContext) (Value, ClassMetaValue, bool) {
	if ctx == nil || ctx.Env() == nil {
		return nil, nil, false
	}

	if currentClassRaw, ok := ctx.Env().Get("__CurrentClass__"); ok {
		if classValue, classMeta, ok := classMetaValueFromRaw(currentClassRaw); ok {
			return classValue, classMeta, true
		}
	}

	if selfRaw, ok := ctx.Env().Get("Self"); ok {
		if classValue, classMeta, ok := classMetaValueFromRaw(selfRaw); ok {
			return classValue, classMeta, true
		}
	}

	return nil, nil, false
}

func classMetaValueFromRaw(raw any) (Value, ClassMetaValue, bool) {
	classValue, ok := raw.(Value)
	if !ok || classValue == nil {
		return nil, nil, false
	}

	if classMeta, ok := raw.(ClassMetaValue); ok && classMeta != nil {
		return classValue, classMeta, true
	}

	if obj, ok := raw.(ObjectValue); ok {
		if classMeta, ok := obj.GetClassType().(ClassMetaValue); ok && classMeta != nil {
			return classValue, classMeta, true
		}
	}

	return nil, nil, false
}
