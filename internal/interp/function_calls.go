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
	copy(interpArgs, args)
	return interpArgs
}

// CallFunctionPointer executes a function pointer with given arguments.
// DEPRECATED: Use FunctionPointerCallable.Invoke + ExecuteFunctionPointerCall instead.
func (i *Interpreter) CallFunctionPointer(funcPtr evaluator.Value, args []evaluator.Value, node ast.Node) evaluator.Value {
	// Convert evaluator.Value to interp.Value (they're the same interface)
	fp, ok := funcPtr.(*FunctionPointerValue)
	if !ok {
		return i.newErrorWithLocation(node, "invalid function pointer type: expected FunctionPointerValue, got %T", funcPtr)
	}

	return i.callFunctionPointer(fp, convertEvaluatorArgs(args), node)
}

// ExecuteFunctionPointerCall executes a function pointer with the given metadata.
// Low-level execution method used by FunctionPointerCallable.Invoke callback.
// Handles the interpreter-dependent parts of function pointer invocation.
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

	// Use ExecuteUserFunction instead of callUserFunction
	callbacks := i.createUserFunctionCallbacks()

	// If this is a method pointer, set up Self binding in a wrapper environment
	// before calling ExecuteUserFunction. The function's enclosed environment
	// will inherit Self from this wrapper.
	if metadata.SelfObject != nil {
		// Phase 3.1.4: unified scope management
		defer i.PushScope()()

		// Bind Self to the captured object
		i.Env().Define("Self", metadata.SelfObject)

		// Sync i.ctx.Env() with i.Env() before calling ExecuteUserFunction.
		// ExecuteUserFunction creates its function environment from ctx.Env(), so we need
		// ctx.Env() to see the Self binding we just set up in i.Env().
		// Note: PushScope already synced i.ctx.env, but we save it here for clarity
		savedCtxEnv := i.ctx.Env()

		// Call the function via ExecuteUserFunction
		result, err := i.evaluatorInstance.ExecuteUserFunction(fn, args, i.ctx, callbacks)

		// Restore context environment
		i.ctx.SetEnv(savedCtxEnv)

		if err != nil {
			return i.newErrorWithLocation(node, "%s", err.Error())
		}
		return result
	}

	// Regular function pointer - call via ExecuteUserFunction
	result, err := i.evaluatorInstance.ExecuteUserFunction(fn, args, i.ctx, callbacks)
	if err != nil {
		return i.newErrorWithLocation(node, "%s", err.Error())
	}
	return result
}

// CallUserFunction executes a user-defined function.
func (i *Interpreter) CallUserFunction(fn *ast.FunctionDecl, args []evaluator.Value) evaluator.Value {
	return i.executeUserFunctionViaEvaluator(fn, convertEvaluatorArgs(args))
}

// CallBuiltinFunction REMOVED
// The evaluator now implements builtins.Context and calls builtins directly
// via builtins.DefaultRegistry.Lookup() instead of using this adapter method.
// See visitor_expressions_functions.go:331 and visitor_expressions_identifiers.go:195

// LookupFunction REMOVED - zero callers in evaluator package
// Evaluator now uses FunctionRegistry.Resolve() for function lookup

// ===== User Function Call Methods =====

// CallImplicitSelfMethod calls a method on the implicit Self object.
// Enables evaluator to call implicit Self methods without using EvalNode.
func (i *Interpreter) CallImplicitSelfMethod(callExpr *ast.CallExpression, funcName *ast.Identifier) evaluator.Value {
	// This method encapsulates the logic from evalCallExpression lines 267-291

	selfVal, ok := i.Env().Get("Self")
	if !ok {
		return i.newErrorWithLocation(callExpr, "no Self context for implicit method call")
	}

	if recVal, ok := selfVal.(*RecordValue); ok {
		if !RecordHasMethod(recVal, funcName.Value) {
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
// Enables evaluator to call record static methods without using EvalNode.
func (i *Interpreter) CallRecordStaticMethod(callExpr *ast.CallExpression, funcName *ast.Identifier) evaluator.Value {
	// This method encapsulates the logic from evalCallExpression lines 293+

	recordVal, ok := i.Env().Get("__CurrentRecord__")
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

// DispatchRecordStaticMethod dispatches a static method call on a record type.
// Task 3.5.146: Simpler adapter method that takes record type name directly.
// The evaluator already verified that the record type exists and has the static method
// via the RecordTypeMetaValue interface.
func (i *Interpreter) DispatchRecordStaticMethod(recordTypeName string, callExpr *ast.CallExpression, funcName *ast.Identifier) evaluator.Value {
	// Create a method call expression: RecordTypeName.MethodName(args)
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
			Value: recordTypeName,
		},
		Method:    funcName,
		Arguments: callExpr.Arguments,
	}
	return i.evalMethodCall(mc)
}

// CallExternalFunction calls an external (Go) function with var parameter support.
// Task 3.2.9: Adapter method to handle external function dispatch from evaluator.
// This encapsulates the logic from evalCallExpression lines 393-439.
func (i *Interpreter) CallExternalFunction(funcName string, argExprs []ast.Expression, node ast.Node) evaluator.Value {
	// Check if this is an external function with var parameters
	if i.evaluatorInstance.ExternalFunctions() == nil {
		return i.newErrorWithLocation(node, "external function registry not initialized")
	}

	// Type-assert to concrete type to access Get method
	registry, ok := i.evaluatorInstance.ExternalFunctions().(*ExternalFunctionRegistry)
	if !ok {
		return i.newErrorWithLocation(node, "external function registry type mismatch")
	}

	extFunc, ok := registry.Get(funcName)
	if !ok {
		return i.newErrorWithLocation(node, "external function '%s' not found", funcName)
	}

	varParams := extFunc.Wrapper.GetVarParams()
	paramTypes := extFunc.Wrapper.GetParamTypes()

	// Prepare arguments - create ReferenceValues for var parameters
	args := make([]Value, len(argExprs))
	for idx, arg := range argExprs {
		isVarParam := idx < len(varParams) && varParams[idx]

		if isVarParam {
			// For var parameters, create a reference
			if argIdent, ok := arg.(*ast.Identifier); ok {
				if val, exists := i.Env().Get(argIdent.Value); exists {
					if refVal, isRef := val.(*ReferenceValue); isRef {
						args[idx] = refVal // Pass through existing reference
					} else {
						args[idx] = &ReferenceValue{Env: i.Env(), VarName: argIdent.Value}
					}
				} else {
					args[idx] = &ReferenceValue{Env: i.Env(), VarName: argIdent.Value}
				}
			} else {
				return i.newErrorWithLocation(arg, "var parameter requires a variable, got %T", arg)
			}
		} else {
			// For regular parameters, evaluate with type context if available
			var val Value
			if idx < len(paramTypes) {
				// Parse the parameter type string and provide context for type inference
				expectedType, _ := i.parseTypeString(paramTypes[idx])
				val = i.EvalWithExpectedType(arg, expectedType)
			} else {
				val = i.Eval(arg)
			}
			if isError(val) {
				return val
			}
			args[idx] = val
		}
	}

	return i.callExternalFunction(extFunc, args)
}

// ===== Task 3.5.7: Function Declaration Methods =====

// EvalMethodImplementation handles method implementation registration for classes/records.
// Task 3.5.7: Delegated from Evaluator.VisitFunctionDecl because it requires ClassInfo
// internals (VMT rebuild, descendant propagation).
func (i *Interpreter) EvalMethodImplementation(fn *ast.FunctionDecl) evaluator.Value {
	if fn == nil || fn.ClassName == nil {
		return i.newErrorWithLocation(fn, "EvalMethodImplementation requires a method declaration with ClassName")
	}

	typeName := fn.ClassName.Value

	// Check if class first (case-insensitive lookup)
	classInfo, isClass := i.classes[ident.Normalize(typeName)]
	if isClass {
		i.evalClassMethodImplementation(fn, classInfo)
		return &NilValue{}
	}

	// Check if record (case-insensitive lookup)
	recordInfo, isRecord := i.records[ident.Normalize(typeName)]
	if isRecord {
		i.evalRecordMethodImplementation(fn, recordInfo)
		return &NilValue{}
	}

	return i.newErrorWithLocation(fn, "type '%s' not found for method '%s'", typeName, fn.Name.Value)
}
