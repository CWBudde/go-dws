package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Float Helper Method Implementations
// ============================================================================
// Task 3.5.102c: Migrate float helper methods from Interpreter to Evaluator.
//
// These implementations avoid the adapter by directly manipulating runtime values.
// The goal is to remove EvalNode delegation for common float operations.

// evalFloatHelper evaluates a built-in float helper method directly in the evaluator.
// Returns the result value, or nil if this helper is not handled here (should fall through
// to the adapter).
//
// Task 3.5.102c: Handles float operations that don't require the builtins Context.
func (e *Evaluator) evalFloatHelper(spec string, selfValue Value, args []Value, node ast.Node) Value {
	switch spec {
	case "__float_tostring_prec":
		return e.evalFloatToStringPrec(selfValue, args, node)

	case "__float_tostring_default":
		return e.evalFloatToStringDefault(selfValue, args, node)

	default:
		// Not a float helper we handle - return nil to signal fallthrough to adapter
		return nil
	}
}

// evalFloatToStringPrec implements Float.ToString(precision) helper method.
// Converts the float to a string with the specified number of decimal places.
func (e *Evaluator) evalFloatToStringPrec(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 1 {
		return e.newError(node, "Float.ToString expects exactly 1 argument")
	}

	floatVal, ok := selfValue.(*runtime.FloatValue)
	if !ok {
		return e.newError(node, "Float.ToString requires float receiver")
	}

	precVal, ok := args[0].(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "Float.ToString precision must be Integer, got %s", args[0].Type())
	}

	precision := int(precVal.Value)
	if precision < 0 {
		precision = 0
	}

	return &runtime.StringValue{Value: fmt.Sprintf("%.*f", precision, floatVal.Value)}
}

// evalFloatToStringDefault implements Float.ToString property (no arguments).
// Converts the float to a string using Go's %g format (compact representation).
func (e *Evaluator) evalFloatToStringDefault(selfValue Value, args []Value, node ast.Node) Value {
	// This is a property read, so args should be empty
	if len(args) != 0 {
		return e.newError(node, "Float.ToString property does not take arguments")
	}

	floatVal, ok := selfValue.(*runtime.FloatValue)
	if !ok {
		return e.newError(node, "Float.ToString property requires float receiver")
	}

	return &runtime.StringValue{Value: fmt.Sprintf("%g", floatVal.Value)}
}
