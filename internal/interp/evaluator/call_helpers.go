package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func (e *Evaluator) executeFunctionPointerDirect(funcPtr Value, args []Value, node ast.Node, ctx *ExecutionContext) Value {
	callable, ok := funcPtr.(FunctionPointerCallable)
	if !ok {
		return e.newError(node, "invalid function pointer type: got %T", funcPtr)
	}
	if callable.IsNil() {
		exc := e.createException("Exception", "Function pointer is nil", nil, ctx)
		ctx.SetException(exc)
		return e.nilValue()
	}

	if callable.IsLambda() {
		lambda, _ := callable.GetLambdaExpr().(*ast.LambdaExpression)
		closureEnv, _ := callable.GetClosure().(*runtime.Environment)
		if lambda == nil || closureEnv == nil {
			return e.newError(node, "invalid lambda function pointer")
		}
		return e.executeLambdaDirect(lambda, closureEnv, args, node, ctx)
	}

	fn, _ := callable.GetFunctionDecl().(*ast.FunctionDecl)
	if fn == nil {
		return e.newError(node, "function pointer is nil")
	}

	callCtx := ctx
	restore := func() {}
	if selfObj := callable.GetSelfObject(); selfObj != nil {
		callCtx = ctx.Clone()
		callCtx.SetEnv(runtime.NewEnclosedEnvironment(ctx.Env()))
		callCtx.Env().Define("Self", selfObj)
		restore = e.bindCurrentMethodIfNamed(callCtx, fn.Name.Value)
		defer restore()
	}

	return e.ExecuteUserFunctionDirect(fn, args, callCtx)
}

func (e *Evaluator) executeLambdaDirect(
	lambda *ast.LambdaExpression,
	closureEnv *runtime.Environment,
	args []Value,
	node ast.Node,
	ctx *ExecutionContext,
) Value {
	if len(args) != len(lambda.Parameters) {
		return e.newError(node, "wrong number of arguments for lambda: expected %d, got %d",
			len(lambda.Parameters), len(args))
	}

	lambdaEnv := runtime.NewEnclosedEnvironment(closureEnv)
	lambdaCtx := ctx.Clone()
	lambdaCtx.SetEnv(lambdaEnv)

	if lambdaCtx.GetCallStack().WillOverflow() {
		return e.raiseRecursionExceeded(ctx)
	}

	if err := lambdaCtx.GetCallStack().Push("<lambda>", e.SourceFile(), nil); err != nil {
		return e.newError(node, "%s", err.Error())
	}
	defer lambdaCtx.GetCallStack().Pop()

	for idx, param := range lambda.Parameters {
		arg := args[idx]
		if param.Type != nil {
			paramTypeName := param.Type.String()
			if converted, ok := e.TryImplicitConversion(arg, paramTypeName, lambdaCtx); ok {
				arg = converted
			}
		}
		lambdaEnv.Define(param.Name.Value, arg)
	}

	if lambda.ReturnType != nil || lambda.IsShorthand {
		var resultValue Value = e.nilValue()
		if lambda.ReturnType != nil {
			returnType, err := e.ResolveTypeFromAnnotation(lambda.ReturnType)
			if err != nil {
				return e.newError(node, "failed to resolve lambda return type: %v", err)
			}
			if returnType != nil && returnType.TypeKind() == "RECORD" {
				resultValue = e.getZeroValueForType(returnType)
			} else {
				resultValue = e.GetDefaultValue(returnType)
			}
		}
		lambdaEnv.Define("Result", resultValue)
	}

	bodyResult := e.Eval(lambda.Body, lambdaCtx)
	if isError(bodyResult) {
		return bodyResult
	}
	if lambdaCtx.Exception() != nil {
		ctx.SetException(lambdaCtx.Exception())
		return e.nilValue()
	}
	if lambdaCtx.ControlFlow().IsExit() {
		lambdaCtx.ControlFlow().Clear()
	}

	if lambda.ReturnType != nil || lambda.IsShorthand {
		if resultVal, ok := lambdaEnv.Get("Result"); ok {
			if value, ok := resultVal.(Value); ok {
				return value
			}
			return e.newError(node, "lambda Result is not a runtime value")
		}
	}
	return e.nilValue()
}

func (e *Evaluator) executeQualifiedFunctionCall(unitName string, member *ast.Identifier, argsExpr []ast.Expression, node ast.Node, ctx *ExecutionContext) Value {
	if e.UnitRegistry() == nil {
		return e.newError(node, "unit registry not initialized")
	}
	if _, exists := e.UnitRegistry().GetUnit(unitName); !exists {
		return e.newError(node, "unit '%s' not loaded", unitName)
	}

	overloads := e.typeSystem.LookupQualifiedFunction(unitName, member.Value)
	if len(overloads) == 0 {
		overloads = e.typeSystem.LookupFunctions(member.Value)
	}
	if len(overloads) == 0 {
		return e.newError(node, "function '%s' not found in unit '%s'", member.Value, unitName)
	}

	var (
		fn         *ast.FunctionDecl
		cachedArgs []Value
		err        error
	)
	if len(overloads) == 1 {
		fn = overloads[0]
		cachedArgs, err = e.ResolveOverloadFast(fn, argsExpr, ctx)
	} else {
		fn, cachedArgs, err = e.ResolveOverloadMultiple(unitName+"."+member.Value, overloads, argsExpr, ctx)
	}
	if err != nil {
		return e.newError(node, "%s", err.Error())
	}

	args, err := e.PrepareUserFunctionArgs(fn, argsExpr, cachedArgs, ctx, node)
	if err != nil {
		return e.newError(node, "%s", err.Error())
	}
	return e.ExecuteUserFunctionDirect(fn, args, ctx)
}

func (e *Evaluator) executeImplicitSelfCall(node *ast.CallExpression, funcName *ast.Identifier, ctx *ExecutionContext) Value {
	selfRaw, ok := ctx.Env().Get("Self")
	if !ok {
		return e.newError(node, "no Self context for implicit method call")
	}

	selfVal, ok := selfRaw.(Value)
	if !ok {
		return e.newError(node, "Self has invalid type")
	}

	args := make([]Value, len(node.Arguments))
	for i, arg := range node.Arguments {
		val := e.Eval(arg, ctx)
		if isError(val) {
			return val
		}
		args[i] = val
	}

	switch self := selfVal.(type) {
	case ClassMetaValue:
		return e.callClassMethod(self, funcName.Value, args, node, ctx)
	case RecordInstanceValue:
		if methodDecl, found := self.GetRecordMethod(funcName.Value); found {
			return e.callRecordMethod(self, methodDecl, args, node, ctx)
		}
		return e.newError(node, "method '%s' not found on Self", funcName.Value)
	default:
		return e.DispatchMethodCall(selfVal, funcName.Value, args, &ast.MethodCallExpression{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{Token: node.Token},
			},
			Object: &ast.Identifier{Value: "Self"},
			Method: funcName,
		}, ctx)
	}
}

func (e *Evaluator) executeInheritedCallDirect(self Value, methodName string, args []Value, node ast.Node, ctx *ExecutionContext) Value {
	if objVal, ok := self.(ObjectValue); ok {
		return objVal.CallInheritedMethod(methodName, args, func(methodDecl any, methodArgs []Value) Value {
			return e.executeObjectMethodDirect(self, methodDecl, methodArgs, node, ctx)
		})
	}

	if classMeta, ok := self.(ClassMetaValue); ok {
		classInfo, ok := classMeta.GetClassInfo().(runtime.IClassInfo)
		if !ok || classInfo == nil || classInfo.GetParent() == nil {
			return e.newError(node, "class '%s' has no parent class", classMeta.GetClassName())
		}
		methodDecl := classInfo.GetParent().LookupClassMethod(methodName)
		if methodDecl == nil {
			return e.newError(node, "method, property, or field '%s' not found in parent class '%s'", methodName, classInfo.GetParent().GetName())
		}
		return e.executeClassMethodDirect(classMeta, methodDecl, args, node, ctx)
	}

	return e.newError(node, "inherited requires Self to be an object or class instance")
}

func (e *Evaluator) bindCurrentMethodIfNamed(ctx *ExecutionContext, methodName string) func() {
	if methodName == "" || ctx == nil || ctx.Env() == nil {
		return func() {}
	}
	ctx.Env().Define("__CurrentMethod__", &runtime.StringValue{Value: methodName})
	return func() {}
}
