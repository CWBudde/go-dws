package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// EvaluateDefaultParameters fills in missing optional arguments with default values.
//
// This method:
//  1. Validates argument count against required/optional parameters
//  2. Evaluates default expressions for missing arguments in the caller's context
//  3. Returns a complete slice of argument values
//
// Default values are evaluated in the caller's environment (ctx), not a new scope.
// This is important for defaults that reference variables in the caller's scope.
//
// Returns an error if:
//   - Too few arguments (less than required parameters)
//   - Too many arguments (more than total parameters)
//   - Default value evaluation fails
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
//
// The function should attempt to convert the value to the target type name.
// Returns (converted value, true) if conversion happened, (original, false) otherwise.
type ImplicitConversionFunc func(value Value, targetTypeName string) (Value, bool)

// BindFunctionParameters binds function parameters to arguments in the execution context.
//
// This method:
//  1. Iterates over function parameters and corresponding arguments
//  2. For var (ByRef) parameters: binds the value directly without conversion
//  3. For regular parameters: applies implicit conversion if a converter is provided
//  4. Defines each parameter in the execution context's environment
//
// The converter callback is used for implicit type conversion. It is not called for:
//   - Var (ByRef) parameters (they keep their ReferenceValue)
//   - Parameters without type annotations
//   - When converter is nil
//
// This mirrors the logic from functions_user.go lines 70-90.
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
// Task 3.5.144b.3: Callback pattern allows complex type resolution to remain in interpreter.
//
// The function receives the return type name (from fn.ReturnType.String()) and should
// return the appropriate default value. The interpreter callback handles:
//   - Basic type defaults (Integer→0, Float→0.0, String→"", Boolean→false)
//   - Record type initialization (creates empty record with nested initialization)
//   - Array type initialization (creates empty array)
//   - Interface type initialization (creates InterfaceInstance with nil object)
type DefaultValueFunc func(returnTypeName string) Value

// FunctionNameAliasFunc is a callback type for creating the function name alias.
// Task 3.5.144b.3: Callback pattern allows ReferenceValue creation to remain in interpreter.
//
// In DWScript, assigning to either Result or the function name sets the return value.
// This callback creates a ReferenceValue that points to "Result" in the function's environment.
// The interpreter callback handles the environment-specific ReferenceValue creation.
//
// Task 3.5.1d: Changed signature to accept the function environment directly,
// since ExecuteUserFunction doesn't swap i.env like the old callUserFunction did.
type FunctionNameAliasFunc func(funcName string, funcEnv *runtime.Environment) Value

// InitializeResultVariable initializes the Result variable for functions.
// Task 3.5.144b.3: Extract Result variable initialization from callUserFunction.
//
// This method:
//  1. Checks if the function has a return type (procedures skip initialization)
//  2. Gets the default value for the return type via callback
//  3. Defines "Result" in the execution context's environment
//  4. Creates function name alias via callback (if provided)
//
// The callbacks allow complex type resolution and ReferenceValue creation to
// remain in the interpreter, while the evaluator handles the control flow.
//
// This mirrors the logic from functions_user.go lines 92-138.
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

// CheckPreconditions is a public wrapper around checkPreconditions for external use.
// Task 3.5.144b.5: Exposes private method from contracts.go for adapter-free access.
func (e *Evaluator) CheckPreconditions(
	funcName string,
	preConditions *ast.PreConditions,
	ctx *ExecutionContext,
) Value {
	return e.checkPreconditions(funcName, preConditions, ctx)
}

// CheckPostconditions is a public wrapper around checkPostconditions for external use.
// Task 3.5.144b.6: Exposes private method from contracts.go for adapter-free access.
func (e *Evaluator) CheckPostconditions(
	funcName string,
	postConditions *ast.PostConditions,
	ctx *ExecutionContext,
) Value {
	return e.checkPostconditions(funcName, postConditions, ctx)
}

// CaptureOldValues is a public wrapper around captureOldValues for external use.
// Task 3.5.144b.6: Exposes private method from contracts.go for adapter-free access.
func (e *Evaluator) CaptureOldValues(
	funcDecl *ast.FunctionDecl,
	ctx *ExecutionContext,
) map[string]Value {
	return e.captureOldValues(funcDecl, ctx)
}

// CleanupInterfaceReferencesFunc is a callback type for cleaning up interface references.
// Task 3.5.144b.10: Callback pattern allows interface cleanup to remain in interpreter.
//
// The function receives the environment and should clean up any interface/object references
// that need to be released when the function scope ends. This includes:
//   - Decrementing reference counts on InterfaceInstance values
//   - Calling destructors on ObjectInstance values if ref count reaches 0
//   - Cleaning up method pointers that hold object references
type CleanupInterfaceReferencesFunc func(env *runtime.Environment)

// TryImplicitConversionReturnFunc is a callback type for implicit return type conversion.
// Task 3.5.144b.8: Callback pattern allows return value conversion to remain in interpreter.
//
// The function receives the return value and target type name, and should attempt
// to convert the value to the target type. Returns (converted value, true) if
// conversion happened, (original, false) otherwise.
type TryImplicitConversionReturnFunc func(returnValue Value, expectedReturnType string) (Value, bool)

// IncrementInterfaceRefCountFunc is a callback type for incrementing interface reference counts.
// Task 3.5.144b.8: Callback pattern for interface reference counting logic.
//
// When returning an interface value from a function, the ref count needs to be incremented
// for the caller's reference. This will be balanced by cleanup releasing Result after return.
type IncrementInterfaceRefCountFunc func(returnValue Value)

// EnvSyncerFunc syncs the interpreter's environment with the evaluator's function environment.
// Task 3.5.22d: This is called before executing the function body to ensure that when
// EvalNode is called back to the interpreter (e.g., for function pointer assignments),
// it uses the correct function environment, not the caller's environment.
// The returned function should be called to restore the original environment.
type EnvSyncerFunc func(funcEnv *runtime.Environment) func()

// UserFunctionCallbacks holds all callback functions needed for user function execution.
// Task 3.5.144b.11: Consolidates all callbacks into a single struct for cleaner API.
type UserFunctionCallbacks struct {
	// ImplicitConversion converts parameter values to match parameter types
	ImplicitConversion ImplicitConversionFunc

	// DefaultValueGetter returns the default value for a return type
	DefaultValueGetter DefaultValueFunc

	// FunctionNameAlias creates a reference value pointing to "Result"
	FunctionNameAlias FunctionNameAliasFunc

	// ReturnValueConverter applies implicit conversion to return values
	ReturnValueConverter TryImplicitConversionReturnFunc

	// InterfaceRefCounter increments ref counts for interface return values
	InterfaceRefCounter IncrementInterfaceRefCountFunc

	// InterfaceCleanup cleans up interface/object references when scope ends
	InterfaceCleanup CleanupInterfaceReferencesFunc

	// EnvSyncer syncs interpreter's i.env with evaluator's funcEnv during body execution
	// Task 3.5.22d: Ensures EvalNode callbacks use the correct environment
	EnvSyncer EnvSyncerFunc
}

// ExecuteUserFunction executes a user-defined function with all necessary setup and cleanup.
// Task 3.5.144b.11: Unified helper that replaces both callUserFunction and invokeParameterlessUserFunction.
//
// This method handles the complete function execution lifecycle:
//  1. Default parameter evaluation (if needed)
//  2. Parameter binding with implicit conversion
//  3. Result variable initialization for functions (not procedures)
//  4. Precondition checking
//  5. Old value capture for postconditions
//  6. Function body execution
//  7. Return value extraction and conversion
//  8. Postcondition checking
//  9. Interface cleanup
//
// The callbacks parameter provides all interpreter-dependent operations:
//   - ImplicitConversion: converts parameter values to match types
//   - DefaultValueGetter: returns default value for return types
//   - FunctionNameAlias: creates reference to Result variable
//   - ReturnValueConverter: converts return value to expected type
//   - InterfaceRefCounter: increments ref counts for interface returns
//   - InterfaceCleanup: cleans up interface/object references
//
// Error handling:
//   - Returns error if preconditions fail
//   - Returns error if postconditions fail
//   - Propagates exceptions via ExecutionContext
//   - Handles exit signal (clears it and returns current Result)
//
// Environment management:
//   - Creates new enclosed environment for function scope
//   - Pushes/pops call stack for recursion tracking
//   - Pushes/pops old values for postcondition evaluation
//   - Restores caller's environment on return
//
// This mirrors the logic from functions_user.go:14-238 (callUserFunction).
func (e *Evaluator) ExecuteUserFunction(
	fn *ast.FunctionDecl,
	args []Value,
	ctx *ExecutionContext,
	callbacks *UserFunctionCallbacks,
) (Value, error) {
	// Task 3.5.1d: Validate argument count (mirrors callUserFunction lines 16-31)
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

	// Create new context with function environment
	funcCtx := &ExecutionContext{
		env:               funcEnv,
		exception:         ctx.exception,
		handlerException:  ctx.handlerException,
		callStack:         ctx.callStack,
		controlFlow:       ctx.controlFlow,
		propContext:       ctx.propContext,
		arrayTypeContext:  ctx.arrayTypeContext,
		evaluator:         ctx.evaluator,
		recordTypeContext: ctx.recordTypeContext,
		envStack:          ctx.envStack,
		oldValuesStack:    ctx.oldValuesStack,
		exceptionGetter:   ctx.exceptionGetter,
		exceptionSetter:   ctx.exceptionSetter,
	}

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

	// Task 3.5.22d: Sync interpreter's i.env with funcEnv before body execution.
	// This ensures that when EvalNode is called back to the interpreter
	// (e.g., for function pointer assignments to Result), it uses funcEnv.
	var restoreEnv func()
	if callbacks.EnvSyncer != nil {
		restoreEnv = callbacks.EnvSyncer(funcEnv)
	}

	// Task 3.5.144b.7: Execute function body via adapter.
	// We use EvalNode instead of e.Eval because the evaluator doesn't fully support
	// all language features yet (e.g., class constructor calls like TClass.Create).
	// The adapter's EvalNode delegates to the interpreter's Eval which has full support.
	_ = e.adapter.EvalNode(fn.Body)

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

	// Task 3.5.144b.8: Extract return value
	var returnValue Value
	if fn.ReturnType != nil {
		// Get Result from function environment
		resultValInterface, resultOk := funcEnv.Get("Result")
		// Debug: Check what we got
		if resultOk {
			// Convert interface{} to Value
			if val, ok := resultValInterface.(Value); ok {
				returnValue = val
			} else {
				returnValue = &runtime.NilValue{}
			}
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

	// Task 3.5.144b.9: Check postconditions after function body executes
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

	// Task 3.5.144b.10: Clean up interface references before restoring environment
	if callbacks.InterfaceCleanup != nil {
		callbacks.InterfaceCleanup(funcEnv)
	}

	// Environment is automatically restored by using funcCtx instead of modifying e.ctx

	return returnValue, nil
}
