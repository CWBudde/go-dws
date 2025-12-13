package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/contracts"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// EvaluateDefaultParameters fills in missing optional arguments with default values.
// Evaluates defaults in the caller's context for variables in caller's scope.
// Returns error if argument count is invalid or evaluation fails.
func (e *Evaluator) EvaluateDefaultParameters(
	fn *ast.FunctionDecl,
	args []Value,
	ctx *ExecutionContext,
) ([]Value, error) {
	// Calculate required parameter count (parameters without defaults)
	requiredParams := 0
	for _, param := range fn.Parameters {
		if param.DefaultValue == nil {
			requiredParams++
		}
	}

	// Check argument count is within valid range
	if len(args) < requiredParams {
		return nil, fmt.Errorf("wrong number of arguments: expected at least %d, got %d",
			requiredParams, len(args))
	}
	if len(args) > len(fn.Parameters) {
		return nil, fmt.Errorf("wrong number of arguments: expected at most %d, got %d",
			len(fn.Parameters), len(args))
	}

	// If all arguments provided, return as-is
	if len(args) == len(fn.Parameters) {
		return args, nil
	}

	// Fill in missing optional arguments with default values
	// Evaluate default expressions in the caller's environment (ctx)
	result := make([]Value, len(fn.Parameters))
	copy(result, args)

	for idx := len(args); idx < len(fn.Parameters); idx++ {
		param := fn.Parameters[idx]
		if param.DefaultValue == nil {
			// This should never happen due to requiredParams check above
			return nil, fmt.Errorf("internal error: missing required parameter at index %d", idx)
		}

		// Evaluate default value in caller's environment
		defaultVal := e.Eval(param.DefaultValue, ctx)
		if isError(defaultVal) {
			return nil, fmt.Errorf("error evaluating default for parameter '%s': %v",
				param.Name.Value, defaultVal)
		}
		result[idx] = defaultVal
	}

	return result, nil
}

// ImplicitConversionFunc is a callback type for implicit type conversion.
type ImplicitConversionFunc = contracts.ImplicitConversionFunc

// BindFunctionParameters binds function parameters to arguments with implicit conversion.
// Var (ByRef) parameters skip conversion to preserve ReferenceValue.
func (e *Evaluator) BindFunctionParameters(
	fn *ast.FunctionDecl,
	args []Value,
	ctx *ExecutionContext,
	converter ImplicitConversionFunc,
) error {
	env := ctx.Env()

	for idx, param := range fn.Parameters {
		arg := args[idx]

		// For var parameters, arg should already be a ReferenceValue
		// Don't apply implicit conversion to references - they need to stay as references
		if !param.ByRef {
			// Apply implicit conversion if parameter has a type and converter is provided
			if param.Type != nil && converter != nil {
				paramTypeName := param.Type.String()
				if converted, ok := converter(arg, paramTypeName); ok {
					arg = converted
				}
			}
		}

		// Store the argument in the function's environment
		// For var parameters, this will be a ReferenceValue
		// For regular parameters, this will be the actual value (possibly converted)
		env.Define(param.Name.Value, arg)
	}

	return nil
}

// DefaultValueFunc is a callback type for getting the default value for a return type.
type DefaultValueFunc = contracts.DefaultValueFunc

// FunctionNameAliasFunc is a callback type for creating the function name alias.
// In DWScript, assigning to either Result or the function name sets the return value.
type FunctionNameAliasFunc = contracts.FunctionNameAliasFunc

// InitializeResultVariable initializes the Result variable for functions (not procedures).
// Creates function name alias to point to Result.
func (e *Evaluator) InitializeResultVariable(
	fn *ast.FunctionDecl,
	ctx *ExecutionContext,
	defaultValueGetter DefaultValueFunc,
	aliasCreator FunctionNameAliasFunc,
) error {
	// Procedures (no return type) don't have a Result variable
	if fn.ReturnType == nil {
		return nil
	}

	env := ctx.Env()

	// Get the default value for the return type
	var resultValue Value
	if defaultValueGetter != nil {
		returnTypeName := fn.ReturnType.String()
		resultValue = defaultValueGetter(returnTypeName)
	} else {
		// Default to NilValue if no callback provided
		resultValue = &runtime.NilValue{}
	}

	// Define Result in the function's environment
	env.Define("Result", resultValue)

	// Create function name alias if callback provided
	// In DWScript, assigning to either Result or the function name sets the return value
	if aliasCreator != nil {
		aliasValue := aliasCreator(fn.Name.Value, env)
		env.Define(fn.Name.Value, aliasValue)
	}

	return nil
}

// CheckPreconditions evaluates preconditions for a function.
func (e *Evaluator) CheckPreconditions(
	funcName string,
	preConditions *ast.PreConditions,
	ctx *ExecutionContext,
) Value {
	return e.checkPreconditions(funcName, preConditions, ctx)
}

// CheckPostconditions evaluates postconditions for a function.
func (e *Evaluator) CheckPostconditions(
	funcName string,
	postConditions *ast.PostConditions,
	ctx *ExecutionContext,
) Value {
	return e.checkPostconditions(funcName, postConditions, ctx)
}

// CaptureOldValues captures variable values for postcondition evaluation.
func (e *Evaluator) CaptureOldValues(
	funcDecl *ast.FunctionDecl,
	ctx *ExecutionContext,
) map[string]Value {
	return e.captureOldValues(funcDecl, ctx)
}

// CleanupInterfaceReferencesFunc is a callback type for cleaning up interface references.
type CleanupInterfaceReferencesFunc = contracts.CleanupInterfaceReferencesFunc

// TryImplicitConversionReturnFunc is a callback type for implicit return type conversion.
type TryImplicitConversionReturnFunc = contracts.TryImplicitConversionReturnFunc

// IncrementInterfaceRefCountFunc is a callback type for incrementing interface reference counts.
type IncrementInterfaceRefCountFunc = contracts.IncrementInterfaceRefCountFunc

// EnvSyncerFunc syncs interpreter's environment with evaluator's function environment.
type EnvSyncerFunc = contracts.EnvSyncerFunc

// UserFunctionCallbacks holds all callback functions needed for user function execution.
type UserFunctionCallbacks = contracts.UserFunctionCallbacks

// ExecuteUserFunction executes a user-defined function with all necessary setup and cleanup.
// Handles parameter binding, result initialization, preconditions, body execution,
// postconditions, and cleanup via callbacks.
func (e *Evaluator) ExecuteUserFunction(
	fn *ast.FunctionDecl,
	args []Value,
	ctx *ExecutionContext,
	callbacks *UserFunctionCallbacks,
) (Value, error) {
	// Validate argument count
	requiredParams := 0
	for _, param := range fn.Parameters {
		if param.DefaultValue == nil {
			requiredParams++
		}
	}
	if len(args) < requiredParams {
		return nil, fmt.Errorf("wrong number of arguments: expected at least %d, got %d",
			requiredParams, len(args))
	}
	if len(args) > len(fn.Parameters) {
		return nil, fmt.Errorf("wrong number of arguments: expected at most %d, got %d",
			len(fn.Parameters), len(args))
	}

	// Fill in missing optional arguments with default values
	// Note: Defaults may already be filled by evaluator.EvaluateDefaultParameters
	// in some call paths. This code handles other paths where defaults aren't filled.
	if len(args) < len(fn.Parameters) && callbacks.DefaultValueGetter != nil {
		for idx := len(args); idx < len(fn.Parameters); idx++ {
			param := fn.Parameters[idx]
			if param.DefaultValue == nil {
				return nil, fmt.Errorf("internal error: missing required parameter at index %d", idx)
			}
			// Evaluate default value in caller's context
			defaultVal := e.Eval(param.DefaultValue, ctx)
			if isError(defaultVal) {
				return nil, fmt.Errorf("error evaluating default value: %v", defaultVal)
			}
			args = append(args, defaultVal)
		}
	}

	// Create new environment for function scope
	funcEnv := runtime.NewEnclosedEnvironment(ctx.Env())

	// Create new context with function environment.
	// Clone preserves shared execution state (call stack, control flow, exception callbacks)
	// while allowing the environment to be swapped for the function scope.
	funcCtx := ctx.Clone()
	funcCtx.SetEnv(funcEnv)

	// Check recursion depth
	if funcCtx.GetCallStack().WillOverflow() {
		return nil, fmt.Errorf("maximum recursion depth exceeded")
	}

	// Push function name onto call stack
	// Note: CallStack.Push requires (functionName, sourceFile, pos)
	// For now, use empty string and nil for sourceFile and pos
	if err := funcCtx.GetCallStack().Push(fn.Name.Value, "", nil); err != nil {
		return nil, err
	}
	defer funcCtx.GetCallStack().Pop()

	// Bind parameters to arguments with implicit conversion
	if err := e.BindFunctionParameters(fn, args, funcCtx, callbacks.ImplicitConversion); err != nil {
		return nil, err
	}

	// Initialize Result variable for functions (not procedures)
	if err := e.InitializeResultVariable(fn, funcCtx, callbacks.DefaultValueGetter, callbacks.FunctionNameAlias); err != nil {
		return nil, err
	}

	// Check preconditions before executing function body
	if fn.PreConditions != nil {
		if err := e.CheckPreconditions(fn.Name.Value, fn.PreConditions, funcCtx); isError(err) {
			return nil, fmt.Errorf("precondition failed: %v", err)
		}
		// If exception was raised during precondition checking, propagate it
		if funcCtx.Exception() != nil {
			ctx.SetException(funcCtx.Exception())
			return &runtime.NilValue{}, nil
		}
	}

	// Capture old values for postcondition evaluation
	oldValues := e.CaptureOldValues(fn, funcCtx)
	// Convert to map[string]interface{} for PushOldValues
	oldValuesInterface := make(map[string]interface{}, len(oldValues))
	for k, v := range oldValues {
		oldValuesInterface[k] = v
	}
	funcCtx.PushOldValues(oldValuesInterface)
	defer funcCtx.PopOldValues()

	// Execute function body
	if fn.Body == nil {
		return nil, fmt.Errorf("function '%s' has no body", fn.Name.Value)
	}

	// Sync interpreter's environment before body execution
	var restoreEnv func()
	if callbacks.EnvSyncer != nil {
		restoreEnv = callbacks.EnvSyncer(funcEnv)
	}

	// Execute function body through the evaluator.
	// Any remaining not-yet-migrated constructs may still fall back to coreEvaluator
	// from within Evaluator.Eval, but this avoids an unconditional EvalNode ping-pong.
	_ = e.Eval(fn.Body, funcCtx)

	// Restore interpreter's environment after body execution
	if restoreEnv != nil {
		restoreEnv()
	}

	// If exception was raised, propagate it to caller's context
	if funcCtx.Exception() != nil {
		ctx.SetException(funcCtx.Exception())
		return &runtime.NilValue{}, nil
	}

	// If exit was called, clear the signal (don't propagate to caller)
	if funcCtx.ControlFlow().IsExit() {
		funcCtx.ControlFlow().Clear()
	}

	// Ensure var (ByRef) parameters are written back to the caller even when the
	// interpreter-backed body execution doesn't propagate assignments.
	for idx, param := range fn.Parameters {
		if !param.ByRef || idx >= len(args) {
			continue
		}

		// Only attempt write-back when the original argument is a reference.
		refArg, ok := args[idx].(ReferenceAccessor)
		if !ok {
			continue
		}

		// Fetch the current value of the parameter from the function environment.
		if paramVal, exists := funcEnv.Get(param.Name.Value); exists {
			switch v := paramVal.(type) {
			case ReferenceAccessor:
				if cur, err := v.Dereference(); err == nil {
					_ = refArg.Assign(cur)
				}
			case Value:
				_ = refArg.Assign(v)
			}
		}
	}

	// Extract return value
	var returnValue Value
	if fn.ReturnType != nil {
		// Get Result from function environment
		resultVal, resultOk := funcEnv.Get("Result")
		if resultOk {
			returnValue = resultVal
		} else {
			returnValue = &runtime.NilValue{}
		}

		// Increment ref count for interface return values (if callback provided)
		if callbacks.InterfaceRefCounter != nil {
			callbacks.InterfaceRefCounter(returnValue)
		}

		// Apply implicit conversion if return type doesn't match (if callback provided)
		if callbacks.ReturnValueConverter != nil && returnValue.Type() != "NIL" {
			expectedReturnType := fn.ReturnType.String()
			if converted, ok := callbacks.ReturnValueConverter(returnValue, expectedReturnType); ok {
				returnValue = converted
			}
		}
	} else {
		// Procedure - no return value
		returnValue = &runtime.NilValue{}
	}

	// Check postconditions after function body executes
	if fn.PostConditions != nil {
		if err := e.CheckPostconditions(fn.Name.Value, fn.PostConditions, funcCtx); isError(err) {
			return nil, fmt.Errorf("postcondition failed: %v", err)
		}
		// If exception was raised during postcondition checking, propagate it
		if funcCtx.Exception() != nil {
			ctx.SetException(funcCtx.Exception())
			return &runtime.NilValue{}, nil
		}
	}

	// Clean up interface references
	if callbacks.InterfaceCleanup != nil {
		callbacks.InterfaceCleanup(funcEnv)
	}

	// Environment is automatically restored by using funcCtx instead of modifying e.ctx

	return returnValue, nil
}
