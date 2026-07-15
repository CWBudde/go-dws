package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
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
	if method.IsClassMethod {
		if classVal, err := e.typeSystem.CreateClassValue(classInfo.GetName()); err == nil && classVal != nil {
			if classMeta, ok := classVal.(ClassMetaValue); ok {
				return e.executeClassMethodDirect(classMeta, method, args, node, ctx)
			}
		}
		return e.newError(node, "class method execution requires runtime class value")
	}

	return e.executeMethodWithClassInfo(self, classInfo, method, args, ctx)
}

// executeMethodWithClassInfo executes a method body with an explicitly supplied
// class context. This allows non-virtual methods to run with a nil Self
// (DWScript permits calling non-virtual methods on nil references; the class
// context then comes from the receiver's static type).
func (e *Evaluator) executeMethodWithClassInfo(
	self Value,
	classInfo runtime.IClassInfo,
	method *ast.FunctionDecl,
	args []Value,
	ctx *ExecutionContext,
) Value {
	ctx.PushEnv()
	defer ctx.PopEnv()

	ctx.Env().Define("Self", self)
	ctx.Env().Define("__CurrentMethod__", &runtime.StringValue{Value: method.Name.Value})
	if defining := definingClassOf(classInfo, method); defining != nil {
		ctx.Env().Define("__CurrentMethodClass__", &runtime.StringValue{Value: defining.GetName()})
	}
	if classVal, err := e.typeSystem.CreateClassValue(classInfo.GetName()); err == nil && classVal != nil {
		if classMeta, ok := classVal.(ClassMetaValue); ok {
			ctx.Env().Define("__CurrentClass__", classMeta)
		}
	}
	e.bindClassConstantsForMethod(classInfo, ctx)

	return e.ExecuteUserFunctionDirect(method, args, ctx)
}

// methodDeclOwner is implemented by class infos that can report whether they
// (and not an ancestor) declare a given method declaration.
type methodDeclOwner interface {
	OwnsMethodDecl(fn *ast.FunctionDecl) bool
}

// definingClassOf walks the class hierarchy to find the class that declares
// the given method. `inherited` inside the method resolves relative to this
// class, not the receiver's dynamic class. Method implementations are
// propagated (pointer-shared) down to descendants during registration, so the
// HIGHEST ancestor owning the declaration is the true definer.
func definingClassOf(classInfo runtime.IClassInfo, method *ast.FunctionDecl) runtime.IClassInfo {
	var defining runtime.IClassInfo
	for current := classInfo; current != nil; current = current.GetParent() {
		owner, ok := current.(methodDeclOwner)
		if !ok {
			break
		}
		if owner.OwnsMethodDecl(method) {
			defining = current
		}
	}
	if defining == nil {
		return classInfo
	}
	return defining
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
	bindClassInfo := classInfo
	if defining := definingClassOf(classInfo, method); defining != nil {
		if method.IsStatic {
			bindClassInfo = defining
			if definingValue, err := e.typeSystem.CreateClassValue(defining.GetName()); err == nil && definingValue != nil {
				if value, ok := definingValue.(Value); ok {
					classValue = value
				}
			}
		}
	}

	ctx.PushEnv()
	defer ctx.PopEnv()

	ctx.Env().Define("Self", classValue)
	ctx.Env().Define("__CurrentClass__", classValue)
	ctx.Env().Define("__CurrentMethod__", &runtime.StringValue{Value: method.Name.Value})
	if defining := definingClassOf(bindClassInfo, method); defining != nil {
		ctx.Env().Define("__CurrentMethodClass__", &runtime.StringValue{Value: defining.GetName()})
	}
	e.bindClassConstantsForMethod(bindClassInfo, ctx)

	return e.ExecuteUserFunctionDirect(method, args, ctx)
}

// currentMethodClassName returns the name of the class that declares the
// currently executing method, or "" when not in a method context. This is the
// static scope in which bare identifiers (implicit Self members) resolve,
// which matters for shadowed fields (a subclass redeclaring a parent's field
// name keeps both storage slots).
func currentMethodClassName(ctx *ExecutionContext) string {
	if ctx == nil || ctx.Env() == nil {
		return ""
	}
	if raw, ok := ctx.Env().Get("__CurrentMethodClass__"); ok {
		if strVal, ok := raw.(*runtime.StringValue); ok {
			return strVal.Value
		}
	}
	return ""
}

// staticClassNameOf returns the static class name of an object-valued
// expression, used to resolve shadowed fields against the reference's static
// type. Self resolves to the declaring class of the current method; other
// expressions use the semantic analyzer's type annotation. Returns "" when no
// static type is known (the caller then falls back to the dynamic class).
func (e *Evaluator) staticClassNameOf(expr ast.Expression, ctx *ExecutionContext) string {
	if identExpr, ok := expr.(*ast.Identifier); ok && ident.Equal(identExpr.Value, "Self") {
		return currentMethodClassName(ctx)
	}
	if e.SemanticInfo() != nil {
		if annot := e.SemanticInfo().GetType(expr); annot != nil {
			return annot.Name
		}
	}
	return ""
}

// getFieldWithStaticClass reads a field honoring field-shadowing semantics
// when the receiver is a concrete object instance; otherwise it falls back to
// the plain dynamic lookup.
func getFieldWithStaticClass(objVal ObjectValue, name string, staticClassName string) Value {
	if objInst, ok := objVal.(*runtime.ObjectInstance); ok {
		return objInst.GetFieldFromClass(name, staticClassName)
	}
	return objVal.GetField(name)
}

// identifierIsInstanceMember reports whether name refers to an instance member
// (field, property, or method) of the class whose method is currently
// executing. Used to raise "Object not instantiated" when a method body runs
// with Self = nil and touches instance state.
func (e *Evaluator) identifierIsInstanceMember(name string, ctx *ExecutionContext) bool {
	_, classMeta, ok := currentClassMetaValue(ctx)
	if !ok || classMeta == nil {
		return false
	}
	classInfo := classMeta.GetClassInfo()
	if classInfo == nil {
		return false
	}
	if classInfo.FieldExists(ident.Normalize(name)) {
		return true
	}
	if classInfo.LookupProperty(name) != nil {
		return true
	}
	if method := classInfo.LookupMethod(name); method != nil && !method.IsClassMethod {
		return true
	}
	return false
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
