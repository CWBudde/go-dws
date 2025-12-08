package evaluator

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// ============================================================================
// String Concatenation Methods
// ============================================================================
//
// This file implements the string concatenation method of the builtins.Context
// interface for the Evaluator:
// - ConcatStrings(): Concatenate multiple string values
//
// Used by the polymorphic Concat() built-in function for string operations.
//
// Phase 3.5.143 - Phase IV: Complex Methods
// ============================================================================

// ConcatStrings concatenates multiple string values into a single string.
// This implements the builtins.Context interface.
func (e *Evaluator) ConcatStrings(args []Value) Value {
	// Build the concatenated string
	var result strings.Builder

	for idx, arg := range args {
		strVal, ok := arg.(*runtime.StringValue)
		if !ok {
			return e.NewError("Concat() expects string as argument %d, got %s", idx+1, arg.Type())
		}
		result.WriteString(strVal.Value)
	}

	return &runtime.StringValue{Value: result.String()}
}
