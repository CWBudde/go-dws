package interp

import (
	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// Phase 3.5.4 - Phase 2A: Function call system adapter methods
// These methods implement the InterpreterAdapter interface for function calls.

// convertEvaluatorArgs converts a slice of evaluator.Value to interp.Value.
// This is used by adapter methods when delegating to internal functions.
func convertEvaluatorArgs(args []evaluator.Value) []Value {
	interpArgs := make([]Value, len(args))
	for idx, arg := range args {
		interpArgs[idx] = arg
	}
	return interpArgs
}

// CallFunctionPointer executes a function pointer with given arguments.
// DEPRECATED: Task 3.5.121 - Use FunctionPointerCallable.Invoke + ExecuteFunctionPointerCall instead.
func (i *Interpreter) CallFunctionPointer(funcPtr evaluator.Value, args []evaluator.Value, node ast.Node) evaluator.Value {
	// Convert evaluator.Value to interp.Value (they're the same interface)
	fp, ok := funcPtr.(*FunctionPointerValue)
	if !ok {
		return i.newErrorWithLocation(node, "invalid function pointer type: expected FunctionPointerValue, got %T", funcPtr)
	}

	return i.callFunctionPointer(fp, convertEvaluatorArgs(args), node)
}

// ExecuteFunctionPointerCall executes a function pointer with the given metadata.
// Task 3.5.121: Low-level execution method used by FunctionPointerCallable.Invoke callback.
// This handles the interpreter-dependent parts of function pointer invocation.
func (i *Interpreter) ExecuteFunctionPointerCall(metadata evaluator.FunctionPointerMetadata, args []evaluator.Value, node ast.Node) evaluator.Value {
	interpArgs := convertEvaluatorArgs(args)

	// Handle lambda execution
	if metadata.IsLambda {
		lambda, ok := metadata.Lambda.(*ast.LambdaExpression)
		if !ok {
			return i.newErrorWithLocation(node, "invalid lambda type in function pointer metadata")
		}

		closureEnv, ok := metadata.Closure.(*Environment)
		if !ok {
			return i.newErrorWithLocation(node, "invalid closure type in function pointer metadata")
		}

		return i.callLambda(lambda, closureEnv, interpArgs, node)
	}

	// Handle regular function pointer execution
	fn, ok := metadata.Function.(*ast.FunctionDecl)
	if !ok {
		return i.newErrorWithLocation(node, "invalid function type in function pointer metadata")
	}

	// If this is a method pointer, set up Self binding
	if metadata.SelfObject != nil {
		funcEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = funcEnv

		// Bind Self to the captured object
		i.env.Define("Self", metadata.SelfObject)

		// Call the function
		result := i.callUserFunction(fn, interpArgs)

		// Restore environment
		i.env = savedEnv

		return result
	}

	// Regular function pointer - just call the function directly
	return i.callUserFunction(fn, interpArgs)
}

// CallUserFunction executes a user-defined function.
func (i *Interpreter) CallUserFunction(fn *ast.FunctionDecl, args []evaluator.Value) evaluator.Value {
	return i.callUserFunction(fn, convertEvaluatorArgs(args))
}

// CallBuiltinFunction executes a built-in function by name.
func (i *Interpreter) CallBuiltinFunction(name string, args []evaluator.Value) evaluator.Value {
	return i.callBuiltinFunction(name, convertEvaluatorArgs(args))
}

// LookupFunction finds a function by name in the function registry.
func (i *Interpreter) LookupFunction(name string) ([]*ast.FunctionDecl, bool) {
	// DWScript is case-insensitive, so normalize to lowercase
	normalizedName := ident.Normalize(name)
	functions, ok := i.functions[normalizedName]
	return functions, ok
}

// ===== Task 3.5.97: User Function Call Methods =====

// CallUserFunctionWithOverloads calls a user-defined function with overload resolution.
// Task 3.5.97a: Enables evaluator to call user functions without using EvalNode.
func (i *Interpreter) CallUserFunctionWithOverloads(callExpr *ast.CallExpression, funcName *ast.Identifier) evaluator.Value {
	// This method encapsulates the logic from evalCallExpression lines 210-265

	funcNameLower := ident.Normalize(funcName.Value)
	overloads, exists := i.functions[funcNameLower]
	if !exists || len(overloads) == 0 {
		return i.newErrorWithLocation(callExpr, "function '%s' not found", funcName.Value)
	}

	// Resolve overload based on argument types and get cached evaluated arguments
	fn, cachedArgs, err := i.resolveOverload(funcNameLower, overloads, callExpr.Arguments)
	if err != nil {
		return i.newErrorWithLocation(callExpr, "%s", err.Error())
	}

	// Prepare arguments - lazy parameters get LazyThunks, var parameters get References
	args := make([]Value, len(callExpr.Arguments))
	for idx, arg := range callExpr.Arguments {
		// Check parameter flags
		isLazy := idx < len(fn.Parameters) && fn.Parameters[idx].IsLazy
		isByRef := idx < len(fn.Parameters) && fn.Parameters[idx].ByRef

		if isLazy {
			// For lazy parameters, create a LazyThunk
			args[idx] = NewLazyThunk(arg, i.env, i)
		} else if isByRef {
			// For var parameters, create a reference
			if argIdent, ok := arg.(*ast.Identifier); ok {
				if val, exists := i.env.Get(argIdent.Value); exists {
					if refVal, isRef := val.(*ReferenceValue); isRef {
						args[idx] = refVal // Pass through existing reference
					} else {
						args[idx] = &ReferenceValue{Env: i.env, VarName: argIdent.Value}
					}
				} else {
					args[idx] = &ReferenceValue{Env: i.env, VarName: argIdent.Value}
				}
			} else {
				return i.newErrorWithLocation(arg, "var parameter requires a variable, got %T", arg)
			}
		} else {
			// For regular parameters, use the cached value from overload resolution
			args[idx] = cachedArgs[idx]
		}
	}
	return i.callUserFunction(fn, args)
}

// CallImplicitSelfMethod calls a method on the implicit Self object.
// Task 3.5.97b: Enables evaluator to call implicit Self methods without using EvalNode.
func (i *Interpreter) CallImplicitSelfMethod(callExpr *ast.CallExpression, funcName *ast.Identifier) evaluator.Value {
	// This method encapsulates the logic from evalCallExpression lines 267-291

	selfVal, ok := i.env.Get("Self")
	if !ok {
		return i.newErrorWithLocation(callExpr, "no Self context for implicit method call")
	}

	obj, isObj := AsObject(selfVal)
	if !isObj {
		return i.newErrorWithLocation(callExpr, "Self is not an object")
	}

	if obj.GetMethod(funcName.Value) == nil {
		return i.newErrorWithLocation(callExpr, "method '%s' not found on Self", funcName.Value)
	}

	// Create a method call expression: Self.MethodName(args)
	mc := &ast.MethodCallExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: callExpr.Token,
			},
		},
		Object: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: funcName.Token,
				},
			},
			Value: "Self",
		},
		Method:    funcName,
		Arguments: callExpr.Arguments,
	}
	return i.evalMethodCall(mc)
}

// CallRecordStaticMethod calls a static method within a record context.
// Task 3.5.97c: Enables evaluator to call record static methods without using EvalNode.
func (i *Interpreter) CallRecordStaticMethod(callExpr *ast.CallExpression, funcName *ast.Identifier) evaluator.Value {
	// This method encapsulates the logic from evalCallExpression lines 293+

	recordVal, ok := i.env.Get("__CurrentRecord__")
	if !ok {
		return i.newErrorWithLocation(callExpr, "no __CurrentRecord__ context for static method call")
	}

	rtv, isRecord := recordVal.(*RecordTypeValue)
	if !isRecord {
		return i.newErrorWithLocation(callExpr, "__CurrentRecord__ is not a RecordTypeValue")
	}

	// Check if the function name matches a static method (case-insensitive)
	methodNameLower := ident.Normalize(funcName.Value)
	overloads, exists := rtv.ClassMethodOverloads[methodNameLower]
	if !exists || len(overloads) == 0 {
		return i.newErrorWithLocation(callExpr, "static method '%s' not found on record type '%s'", funcName.Value, rtv.RecordType.Name)
	}

	// Found a static method - convert to qualified call
	mc := &ast.MethodCallExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: callExpr.Token,
			},
		},
		Object: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: funcName.Token,
				},
			},
			Value: rtv.RecordType.Name,
		},
		Method:    funcName,
		Arguments: callExpr.Arguments,
	}
	return i.evalMethodCall(mc)
}
