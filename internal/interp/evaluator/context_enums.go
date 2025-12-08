package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/builtins"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// ============================================================================
// Enum Operation Methods
// ============================================================================
//
// This file implements the enum operation methods of the builtins.Context
// interface for the Evaluator:
// - GetEnumOrdinal(): Get ordinal value of an enum
// - GetEnumSuccessor(): Get next enum value in sequence
// - GetEnumPredecessor(): Get previous enum value in sequence
// - GetJSONVarType(): Get VarType code for JSON values
//
// These methods use the builtins.Context interface which the Interpreter
// already implements in builtins_context.go. We cast the adapter to
// builtins.Context to access these methods without duplicating code.
// ============================================================================

// GetEnumOrdinal returns the ordinal value of an enum Value.
// This implements the builtins.Context interface.
func (e *Evaluator) GetEnumOrdinal(value Value) (int64, bool) {
	// Cast adapter to builtins.Context to access the existing implementation.
	// The Interpreter implements builtins.Context with these methods.
	ctx, ok := e.adapter.(builtins.Context)
	if !ok {
		return 0, false
	}
	// Convert evaluator.Value to runtime.Value (which is builtins.Value)
	return ctx.GetEnumOrdinal(runtime.Value(value))
}

// GetEnumSuccessor returns the successor of an enum value.
// This implements the builtins.Context interface.
// Task 3.7.8: Helper for Succ() function.
func (e *Evaluator) GetEnumSuccessor(enumVal Value) (Value, error) {
	// Cast adapter to builtins.Context to access the existing implementation
	ctx, ok := e.adapter.(builtins.Context)
	if !ok {
		return nil, nil
	}
	// Convert evaluator.Value to runtime.Value (which is builtins.Value)
	return ctx.GetEnumSuccessor(runtime.Value(enumVal))
}

// GetEnumPredecessor returns the predecessor of an enum value.
// This implements the builtins.Context interface.
// Task 3.7.8: Helper for Pred() function.
func (e *Evaluator) GetEnumPredecessor(enumVal Value) (Value, error) {
	// Cast adapter to builtins.Context to access the existing implementation
	ctx, ok := e.adapter.(builtins.Context)
	if !ok {
		return nil, nil
	}
	// Convert evaluator.Value to runtime.Value (which is builtins.Value)
	return ctx.GetEnumPredecessor(runtime.Value(enumVal))
}

// GetJSONVarType returns the VarType code for a JSON value based on its kind.
// This implements the builtins.Context interface.
// Task 3.7.5: Helper for VarType() function to handle JSON values.
func (e *Evaluator) GetJSONVarType(value Value) (int64, bool) {
	// Cast adapter to builtins.Context to access the existing implementation
	ctx, ok := e.adapter.(builtins.Context)
	if !ok {
		return 0, false
	}
	// Convert evaluator.Value to runtime.Value (which is builtins.Value)
	return ctx.GetJSONVarType(runtime.Value(value))
}

// GetEnumMetadata retrieves enum type metadata by type name.
// This implements the builtins.Context interface.
// Used by Succ/Pred builtins to navigate enum ordinals.
func (e *Evaluator) GetEnumMetadata(typeName string) builtins.Value {
	// Cast adapter to builtins.Context to access the existing implementation
	ctx, ok := e.adapter.(builtins.Context)
	if !ok {
		return nil
	}
	return ctx.GetEnumMetadata(typeName)
}
