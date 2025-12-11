package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/builtins"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// ============================================================================
// Bounds Operation Methods
// ============================================================================
//
// This file implements the bounds operation methods of the builtins.Context
// interface for the Evaluator:
// - GetLowBound(): Get lower bound of array, enum, or type meta-value
// - GetHighBound(): Get upper bound of array, enum, or type meta-value
//
// Polymorphic behavior:
// - Arrays: Return array bounds from ArrayType or dynamic bounds
// - Enums: Return first/last enum value as bound
// - Type meta-values: Return bounds for built-in types or enum bounds
//
// These methods use the builtins.Context interface which the Interpreter
// already implements in builtins_context.go. We cast the adapter to
// builtins.Context to access these methods without duplicating code.
//
// ============================================================================

// GetLowBound returns the lower bound for arrays, enums, or type meta-values.
// This implements the builtins.Context interface.
func (e *Evaluator) GetLowBound(value Value) (Value, error) {
	// Cast adapter to builtins.Context to access the existing implementation.
	// The Interpreter implements builtins.Context with these methods.
	ctx, ok := e.coreEvaluator.(builtins.Context)
	if !ok {
		return nil, fmt.Errorf("GetLowBound: adapter does not implement builtins.Context")
	}
	// Convert evaluator.Value to runtime.Value (which is builtins.Value)
	return ctx.GetLowBound(runtime.Value(value))
}

// GetHighBound returns the upper bound for arrays, enums, or type meta-values.
// This implements the builtins.Context interface.
func (e *Evaluator) GetHighBound(value Value) (Value, error) {
	// Cast adapter to builtins.Context to access the existing implementation
	ctx, ok := e.coreEvaluator.(builtins.Context)
	if !ok {
		return nil, fmt.Errorf("GetHighBound: adapter does not implement builtins.Context")
	}
	// Convert evaluator.Value to runtime.Value (which is builtins.Value)
	return ctx.GetHighBound(runtime.Value(value))
}
