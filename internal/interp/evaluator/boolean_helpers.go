package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Boolean Helper Method Implementations
// ============================================================================
// Task 3.5.102d: Migrate boolean helper methods from Interpreter to Evaluator.
//
// These implementations avoid the adapter by directly manipulating runtime values.
// The goal is to remove EvalNode delegation for common boolean operations.

// evalBooleanHelper evaluates a built-in boolean helper method directly in the evaluator.
// Returns the result value, or nil if this helper is not handled here (should fall through
// to the adapter).
//
// Task 3.5.102d: Handles boolean operations that don't require the builtins Context.
func (e *Evaluator) evalBooleanHelper(spec string, selfValue Value, args []Value, node ast.Node) Value {
	switch spec {
	case "__boolean_tostring":
		return e.evalBooleanToString(selfValue, args, node)

	default:
		// Not a boolean helper we handle - return nil to signal fallthrough to adapter
		return nil
	}
}

// evalBooleanToString implements Boolean.ToString() helper method.
// Converts the boolean to "True" or "False" string (Pascal-style capitalization).
func (e *Evaluator) evalBooleanToString(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 0 {
		return e.newError(node, "Boolean.ToString does not take arguments")
	}

	boolVal, ok := selfValue.(*runtime.BooleanValue)
	if !ok {
		return e.newError(node, "Boolean.ToString requires boolean receiver")
	}

	if boolVal.Value {
		return &runtime.StringValue{Value: "True"}
	}
	return &runtime.StringValue{Value: "False"}
}
