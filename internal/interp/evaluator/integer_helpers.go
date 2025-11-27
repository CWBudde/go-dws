package evaluator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Integer Helper Method Implementations
// ============================================================================
// Task 3.5.102b: Migrate integer helper methods from Interpreter to Evaluator.
//
// These implementations avoid the adapter by directly manipulating runtime values.
// The goal is to remove EvalNode delegation for common integer operations.

// evalIntegerHelper evaluates a built-in integer helper method directly in the evaluator.
// Returns the result value, or nil if this helper is not handled here (should fall through
// to the adapter).
//
// Task 3.5.102b: Handles integer operations that don't require the builtins Context.
func (e *Evaluator) evalIntegerHelper(spec string, selfValue Value, args []Value, node ast.Node) Value {
	switch spec {
	case "__integer_tostring":
		return e.evalIntegerToString(selfValue, args, node)

	case "__integer_tohexstring":
		return e.evalIntegerToHexString(selfValue, args, node)

	default:
		// Not an integer helper we handle - return nil to signal fallthrough to adapter
		return nil
	}
}

// evalIntegerToString implements Integer.ToString() / Integer.ToString(base) helper method.
// Converts the integer to a string representation in the specified base (2-36).
// If no base is provided, defaults to base 10.
func (e *Evaluator) evalIntegerToString(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) > 1 {
		return e.newError(node, "Integer.ToString expects 0 or 1 argument")
	}

	intVal, ok := selfValue.(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "Integer.ToString requires integer receiver")
	}

	base := 10
	if len(args) == 1 {
		baseVal, ok := args[0].(*runtime.IntegerValue)
		if !ok {
			return e.newError(node, "Integer.ToString base must be Integer, got %s", args[0].Type())
		}
		base = int(baseVal.Value)
		if base < 2 || base > 36 {
			return e.newError(node, "Integer.ToString base must be between 2 and 36, got %d", base)
		}
	}

	return &runtime.StringValue{Value: strconv.FormatInt(intVal.Value, base)}
}

// evalIntegerToHexString implements Integer.ToHexString(digits) helper method.
// Converts the integer to an uppercase hexadecimal string, zero-padded to the
// specified number of digits.
func (e *Evaluator) evalIntegerToHexString(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 1 {
		return e.newError(node, "Integer.ToHexString expects exactly 1 argument")
	}

	intVal, ok := selfValue.(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "Integer.ToHexString requires integer receiver")
	}

	digitsVal, ok := args[0].(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "Integer.ToHexString digits must be Integer, got %s", args[0].Type())
	}

	digits := int(digitsVal.Value)
	if digits < 0 {
		digits = 0
	}

	// Format as uppercase hex
	hexStr := fmt.Sprintf("%X", intVal.Value)

	// Pad with leading zeros if needed
	if len(hexStr) < digits {
		hexStr = strings.Repeat("0", digits-len(hexStr)) + hexStr
	}

	return &runtime.StringValue{Value: hexStr}
}
