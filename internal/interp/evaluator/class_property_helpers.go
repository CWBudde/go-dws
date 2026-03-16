package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

func (e *Evaluator) evalClassPropertyRead(
	classInfo runtime.IClassInfo,
	propInfo *types.PropertyInfo,
	node ast.Node,
	ctx *ExecutionContext,
) Value {
	if propInfo.IsIndexed {
		return e.newError(node, "indexed class property '%s' requires index arguments", propInfo.Name)
	}

	switch propInfo.ReadKind {
	case types.PropAccessField:
		if classVarValue, _ := classInfo.LookupClassVar(propInfo.ReadSpec); classVarValue != nil {
			return classVarValue
		}

		method := classInfo.LookupClassMethod(propInfo.ReadSpec)
		if method == nil {
			return e.newError(node, "class property '%s' read specifier '%s' not found as class variable or class method", propInfo.Name, propInfo.ReadSpec)
		}
		return e.executeClassPropertyMethod(classInfo, method, nil, node, ctx)

	case types.PropAccessMethod:
		method := classInfo.LookupClassMethod(propInfo.ReadSpec)
		if method == nil {
			return e.newError(node, "class property '%s' getter method '%s' not found", propInfo.Name, propInfo.ReadSpec)
		}
		return e.executeClassPropertyMethod(classInfo, method, nil, node, ctx)

	default:
		return e.newError(node, "class property '%s' has no read access", propInfo.Name)
	}
}

func (e *Evaluator) evalClassPropertyWrite(
	classInfo runtime.IClassInfo,
	propInfo *types.PropertyInfo,
	value Value,
	node ast.Node,
	ctx *ExecutionContext,
) Value {
	if propInfo.IsIndexed {
		return e.newError(node, "indexed class property '%s' requires index arguments", propInfo.Name)
	}
	if propInfo.WriteKind == types.PropAccessNone {
		return e.newError(node, "class property '%s' is read-only", propInfo.Name)
	}

	switch propInfo.WriteKind {
	case types.PropAccessField:
		if e.setClassVarValue(classInfo, propInfo.WriteSpec, value) {
			return value
		}

		method := classInfo.LookupClassMethod(propInfo.WriteSpec)
		if method == nil {
			return e.newError(node, "class property '%s' write specifier '%s' not found as class variable or class method", propInfo.Name, propInfo.WriteSpec)
		}
		return e.executeClassPropertyMethod(classInfo, method, []Value{value}, node, ctx)

	case types.PropAccessMethod:
		method := classInfo.LookupClassMethod(propInfo.WriteSpec)
		if method == nil {
			return e.newError(node, "class property '%s' setter method '%s' not found", propInfo.Name, propInfo.WriteSpec)
		}
		return e.executeClassPropertyMethod(classInfo, method, []Value{value}, node, ctx)

	default:
		return e.newError(node, "class property '%s' has no write access", propInfo.Name)
	}
}

func (e *Evaluator) executeClassPropertyMethod(
	classInfo runtime.IClassInfo,
	method *ast.FunctionDecl,
	args []Value,
	node ast.Node,
	ctx *ExecutionContext,
) Value {
	ctx.PushEnv()
	defer ctx.PopEnv()
	scope := newBindingScope()
	defer scope.cleanup(e, ctx.Env())

	e.bindClassVarsForProperty(classInfo, ctx, scope)

	if len(args) > 0 && len(method.Parameters) > 0 {
		scope.defineOwned(e, ctx, method.Parameters[0].Name.Value, args[0])
	}

	if method.ReturnType != nil {
		returnType, err := e.ResolveTypeFromAnnotation(method.ReturnType)
		if err != nil {
			return e.newError(node, "failed to resolve return type: %v", err)
		}
		defaultVal := e.GetDefaultValue(returnType)
		scope.defineOwned(e, ctx, "Result", defaultVal)
		scope.defineExposed(ctx, method.Name.Value, e.newResultAlias(ctx.Env()))
	}

	result := e.Eval(method.Body, ctx)
	if isError(result) {
		return result
	}

	e.syncClassVarsFromEnv(classInfo, ctx.Env())

	if method.ReturnType != nil {
		returnValue := e.extractReturnValue(method.Name.Value, ctx)
		return e.retainValueForBinding(returnValue, ctx)
	}

	if len(args) > 0 {
		return args[0]
	}
	return e.nilValue()
}

func (e *Evaluator) bindClassVarsForProperty(classInfo runtime.IClassInfo, ctx *ExecutionContext, scope *bindingScope) {
	chain := classInfoHierarchy(classInfo)
	for _, cls := range chain {
		for name, value := range cls.GetClassVarsMap() {
			scope.defineExposed(ctx, name, value)
		}
	}
}

func (e *Evaluator) syncClassVarsFromEnv(classInfo runtime.IClassInfo, env *runtime.Environment) {
	for _, cls := range classInfoHierarchy(classInfo) {
		for name := range cls.GetClassVarsMap() {
			if val, ok := env.Get(name); ok {
				cls.GetClassVarsMap()[name] = val
			}
		}
	}
}

func (e *Evaluator) setClassVarValue(classInfo runtime.IClassInfo, name string, value Value) bool {
	_, owner := classInfo.LookupClassVar(name)
	if owner == nil {
		return false
	}

	classVars := owner.GetClassVarsMap()
	if classVars == nil {
		return false
	}
	if _, exists := classVars[name]; exists {
		classVars[name] = value
		return true
	}
	for existingName := range classVars {
		if ident.Equal(existingName, name) {
			classVars[existingName] = value
			return true
		}
	}
	return false
}

func (e *Evaluator) newResultAlias(env *runtime.Environment) Value {
	getter := func() (Value, error) {
		val, ok := env.Get("Result")
		if !ok {
			return &runtime.NilValue{}, nil
		}
		return val.(Value), nil
	}
	setter := func(val Value) error {
		return env.Set("Result", val)
	}
	return runtime.NewReferenceValue("Result", getter, setter)
}

// executeRecordIndexedPropertyRead executes a record indexed property getter method.
//
// This is the evaluator-owned replacement for oopEngine.ExecuteRecordPropertyRead.
// Instead of creating a synthetic AST call and binding temp vars in the interpreter's
// environment, we resolve the getter method directly and call it with the indices as args.
func (e *Evaluator) executeRecordIndexedPropertyRead(record Value, propInfoAny any, indices []Value, node ast.Node, ctx *ExecutionContext) Value {
	recVal, ok := record.(RecordInstanceValue)
	if !ok {
		return e.newError(node, "indexed property read requires a record value")
	}
	propInfo, ok := propInfoAny.(*types.RecordPropertyInfo)
	if !ok {
		return e.newError(node, "internal error: expected *types.RecordPropertyInfo for indexed property read")
	}
	if propInfo.ReadField == "" {
		return e.newError(node, "default property is write-only")
	}
	methodDecl, found := recVal.GetRecordMethod(propInfo.ReadField)
	if !found {
		return e.newError(node, "default property read accessor '%s' is not a method", propInfo.ReadField)
	}
	return e.callRecordMethod(recVal, methodDecl, indices, node, ctx)
}

func classInfoHierarchy(classInfo runtime.IClassInfo) []runtime.IClassInfo {
	if classInfo == nil {
		return nil
	}

	var reversed []runtime.IClassInfo
	for current := classInfo; current != nil; current = current.GetParent() {
		reversed = append(reversed, current)
	}

	hierarchy := make([]runtime.IClassInfo, 0, len(reversed))
	for idx := len(reversed) - 1; idx >= 0; idx-- {
		hierarchy = append(hierarchy, reversed[idx])
	}
	return hierarchy
}
