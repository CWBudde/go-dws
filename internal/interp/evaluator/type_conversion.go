// Package evaluator provides type conversion helpers for implicit type conversions.
//
// Task 3.5.22f: Create Evaluator Helper for Conversion Function Execution
package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ConversionCallbacks holds the callbacks needed for executing conversion functions.
// These callbacks are provided by the interpreter to handle interpreter-dependent operations.
type ConversionCallbacks struct {
	// ImplicitConversion is used for parameter type conversion during the function call.
	// Can be nil if no parameter conversion is needed.
	ImplicitConversion ImplicitConversionFunc

	// EnvSyncer syncs the interpreter's environment with the function environment.
	// This is required for proper function body execution.
	EnvSyncer EnvSyncerFunc
}

// ExecuteConversionFunction executes a user-defined conversion function with a single argument.
//
// Task 3.5.22f: This helper enables the evaluator to execute implicit conversion operators
// that are defined as functions (via the 'implicit operator' syntax).
//
// The conversion function is expected to:
//   - Take a single parameter (the value to convert)
//   - Return the converted value
//
// This method uses ExecuteUserFunction with a minimal callback set since conversion
// functions typically don't need complex features like return type initialization
// (they're simple single-argument functions).
//
// Parameters:
//   - fn: The conversion function declaration (must have exactly 1 parameter)
//   - arg: The value to convert
//   - ctx: The execution context
//   - callbacks: Optional callbacks for interpreter-dependent operations
//
// Returns:
//   - The converted value, or nil if an error occurred
//   - An error if the conversion fails
func (e *Evaluator) ExecuteConversionFunction(
	fn *ast.FunctionDecl,
	arg Value,
	ctx *ExecutionContext,
	callbacks *ConversionCallbacks,
) (Value, error) {
	// Validate that the function has exactly one parameter
	if len(fn.Parameters) != 1 {
		return nil, fmt.Errorf("conversion function '%s' must have exactly 1 parameter, got %d",
			fn.Name.Value, len(fn.Parameters))
	}

	// Validate that the function has a return type (conversions must return a value)
	if fn.ReturnType == nil {
		return nil, fmt.Errorf("conversion function '%s' must have a return type", fn.Name.Value)
	}

	// Prepare arguments as a slice
	args := []Value{arg}

	// Create a minimal callback set for conversion function execution
	// Conversion functions are simple: one parameter in, one value out
	userCallbacks := &UserFunctionCallbacks{
		// DefaultValueGetter returns the default for the return type
		DefaultValueGetter: func(returnTypeName string) Value {
			// For conversion functions, we typically expect a value to be returned
			// The default is nil, which will be overwritten by the function body
			return &runtime.NilValue{}
		},
	}

	// Copy callbacks from ConversionCallbacks if provided
	if callbacks != nil {
		if callbacks.ImplicitConversion != nil {
			userCallbacks.ImplicitConversion = callbacks.ImplicitConversion
		}
		if callbacks.EnvSyncer != nil {
			userCallbacks.EnvSyncer = callbacks.EnvSyncer
		}
	}

	// Execute the conversion function
	result, err := e.ExecuteUserFunction(fn, args, ctx, userCallbacks)
	if err != nil {
		return nil, fmt.Errorf("conversion function '%s' failed: %w", fn.Name.Value, err)
	}

	// Check for exceptions
	if ctx.Exception() != nil {
		return nil, fmt.Errorf("conversion function '%s' raised an exception", fn.Name.Value)
	}

	return result, nil
}

// ExecuteConversionFunctionSimple is a simplified version of ExecuteConversionFunction
// that uses the adapter for environment syncing.
//
// Task 3.5.22f: This version is for use when the full ConversionCallbacks are not available,
// falling back to adapter methods.
//
// Parameters:
//   - fn: The conversion function declaration (must have exactly 1 parameter)
//   - arg: The value to convert
//   - ctx: The execution context
//
// Returns:
//   - The converted value, or nil if an error occurred
//   - An error if the conversion fails
func (e *Evaluator) ExecuteConversionFunctionSimple(
	fn *ast.FunctionDecl,
	arg Value,
	ctx *ExecutionContext,
) (Value, error) {
	// Use adapter.TryImplicitConversion for parameter conversion if needed
	var implicitConversion ImplicitConversionFunc
	if e.adapter != nil {
		implicitConversion = func(value Value, targetTypeName string) (Value, bool) {
			return e.adapter.TryImplicitConversion(value, targetTypeName)
		}
	}

	callbacks := &ConversionCallbacks{
		ImplicitConversion: implicitConversion,
		// EnvSyncer is nil - ExecuteUserFunction uses adapter.EvalNode which handles env
	}

	return e.ExecuteConversionFunction(fn, arg, ctx, callbacks)
}
