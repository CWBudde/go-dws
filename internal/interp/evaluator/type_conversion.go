// Package evaluator provides type conversion helpers for implicit type conversions.
package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
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
// This helper enables the evaluator to execute implicit conversion operators
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
// that uses the evaluator's native TryImplicitConversion for parameter conversion.
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
	// Use evaluator's native TryImplicitConversion for parameter conversion if needed
	implicitConversion := func(value Value, targetTypeName string) (Value, bool) {
		return e.TryImplicitConversion(value, targetTypeName, ctx)
	}

	callbacks := &ConversionCallbacks{
		ImplicitConversion: implicitConversion,
		// EnvSyncer is nil - ExecuteUserFunction uses adapter.EvalNode which handles env
	}

	return e.ExecuteConversionFunction(fn, arg, ctx, callbacks)
}

// TryImplicitConversion attempts to apply an implicit conversion from value to targetTypeName.
//
// The conversion logic follows this priority order:
//  1. Direct conversion: Look up a registered implicit conversion from source to target type
//  2. Chained conversion: Find a path of implicit conversions (max 3 steps)
//  3. Built-in conversions: Integer→Float (widening), Enum→Integer (ordinal)
//
// Parameters:
//   - value: The value to convert
//   - targetTypeName: The target type name (as it appears in type annotations)
//   - ctx: The execution context for function calls
//
// Returns:
//   - (convertedValue, true) if conversion was found and applied
//   - (original value, false) if no conversion was needed or available
func (e *Evaluator) TryImplicitConversion(value Value, targetTypeName string, ctx *ExecutionContext) (Value, bool) {
	// Handle nil value
	if value == nil {
		return nil, false
	}

	sourceTypeName := value.Type()

	// No conversion needed if types already match
	if sourceTypeName == targetTypeName {
		return value, false
	}

	// Normalize type names for conversion lookup (to match how they're registered)
	normalizedSource := interptypes.NormalizeTypeAnnotation(sourceTypeName)
	normalizedTarget := interptypes.NormalizeTypeAnnotation(targetTypeName)

	// Try direct conversion first (using TypeSystem's ConversionRegistry)
	entry, found := e.typeSystem.Conversions().FindImplicit(normalizedSource, normalizedTarget)
	if found {
		result, ok := e.executeConversionEntry(entry, value, ctx)
		if ok {
			return result, true
		}
		// If conversion function execution failed, continue to try other methods
	}

	// Try chained conversion if direct conversion not found
	const maxConversionChainDepth = 3
	path := e.typeSystem.Conversions().FindConversionPath(normalizedSource, normalizedTarget, maxConversionChainDepth)
	if len(path) >= 2 {
		result, ok := e.executeConversionChain(path, value, ctx)
		if ok {
			return result, true
		}
	}

	// Built-in conversions (no registry entry needed)

	// Integer → Float is always allowed in Pascal/Delphi (automatic widening)
	if normalizedSource == "integer" && normalizedTarget == "float" {
		if intVal, ok := value.(*runtime.IntegerValue); ok {
			return &runtime.FloatValue{Value: float64(intVal.Value)}, true
		}
	}

	// Enum → Integer implicit conversion
	if enumVal, ok := value.(*runtime.EnumValue); ok && normalizedTarget == "integer" {
		return &runtime.IntegerValue{Value: int64(enumVal.OrdinalValue)}, true
	}

	return value, false
}

// executeConversionEntry executes a single conversion entry (direct conversion).
//
// This helper looks up the conversion function by its binding name and executes it.
//
// Returns:
//   - (convertedValue, true) if conversion succeeded
//   - (nil, false) if conversion failed
func (e *Evaluator) executeConversionEntry(entry *interptypes.ConversionEntry, value Value, ctx *ExecutionContext) (Value, bool) {
	if entry == nil {
		return nil, false
	}

	// Look up the conversion function using TypeSystem's FunctionRegistry
	overloads := e.typeSystem.LookupFunctions(entry.BindingName)
	if len(overloads) == 0 {
		// This should not happen if semantic analysis passed
		return nil, false
	}
	fn := overloads[0]

	// Execute the conversion function using our existing helper
	result, err := e.ExecuteConversionFunctionSimple(fn, value, ctx)
	if err != nil {
		// Conversion function execution failed
		return nil, false
	}

	// Check if result is an error value
	if isErrorValue(result) {
		return nil, false
	}

	return result, true
}

// executeConversionChain applies a sequence of conversions along a path.
//
// This helper iterates through the conversion path, applying each step sequentially.
// The path is a slice of type names where each adjacent pair represents a conversion.
//
// Example: path = ["integer", "mytype", "float"] means:
//  1. Convert integer → mytype
//  2. Convert mytype → float
//
// Returns:
//   - (convertedValue, true) if all conversions in the chain succeeded
//   - (nil, false) if any conversion in the chain failed
func (e *Evaluator) executeConversionChain(path []string, value Value, ctx *ExecutionContext) (Value, bool) {
	if len(path) < 2 {
		return nil, false
	}

	currentValue := value
	for idx := 0; idx < len(path)-1; idx++ {
		fromType := path[idx]
		toType := path[idx+1]

		// Find the conversion function for this step
		stepEntry, stepFound := e.typeSystem.Conversions().FindImplicit(fromType, toType)
		if !stepFound {
			// Path is broken - this shouldn't happen if findConversionPath worked correctly
			return nil, false
		}

		// Execute this conversion step
		result, ok := e.executeConversionEntry(stepEntry, currentValue, ctx)
		if !ok {
			return nil, false
		}

		currentValue = result
	}

	return currentValue, true
}

// isErrorValue checks if a value is an error value.
// This is a local helper to avoid importing the full interp package.
func isErrorValue(val Value) bool {
	if val == nil {
		return false
	}
	return val.Type() == "ERROR"
}
