package evaluator

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// String Helper Method Implementations
// ============================================================================
// Task 3.5.102a: Migrate simple string helper methods from Interpreter to Evaluator.
//
// These implementations avoid the adapter by directly manipulating runtime values.
// The goal is to remove EvalNode delegation for common string operations.

// evalStringHelper evaluates a built-in string helper method directly in the evaluator.
// Returns the result value, or nil if this helper is not handled here (should fall through
// to the adapter).
//
// Task 3.5.102a: Handles simple string operations that don't require the builtins Context.
func (e *Evaluator) evalStringHelper(spec string, selfValue Value, args []Value, node ast.Node) Value {
	switch spec {
	case "__string_toupper":
		return e.evalStringToUpper(selfValue, args, node)

	case "__string_tolower":
		return e.evalStringToLower(selfValue, args, node)

	case "__string_length":
		return e.evalStringLength(selfValue, node)

	case "__string_tostring":
		return e.evalStringToString(selfValue, args, node)

	default:
		// Not a simple string helper - return nil to signal fallthrough to adapter
		return nil
	}
}

// evalStringToUpper implements String.ToUpper() helper method.
// Converts the string to uppercase using Go's strings.ToUpper.
func (e *Evaluator) evalStringToUpper(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 0 {
		return e.newError(node, "String.ToUpper does not take arguments")
	}

	strVal, ok := selfValue.(*runtime.StringValue)
	if !ok {
		return e.newError(node, "String.ToUpper requires string receiver")
	}

	return &runtime.StringValue{Value: strings.ToUpper(strVal.Value)}
}

// evalStringToLower implements String.ToLower() helper method.
// Converts the string to lowercase using Go's strings.ToLower.
func (e *Evaluator) evalStringToLower(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 0 {
		return e.newError(node, "String.ToLower does not take arguments")
	}

	strVal, ok := selfValue.(*runtime.StringValue)
	if !ok {
		return e.newError(node, "String.ToLower requires string receiver")
	}

	return &runtime.StringValue{Value: strings.ToLower(strVal.Value)}
}

// evalStringLength implements String.Length property read.
// Returns the length of the string as an integer.
func (e *Evaluator) evalStringLength(selfValue Value, node ast.Node) Value {
	strVal, ok := selfValue.(*runtime.StringValue)
	if !ok {
		return e.newError(node, "String.Length property requires string receiver")
	}

	return &runtime.IntegerValue{Value: int64(len(strVal.Value))}
}

// evalStringToString implements String.ToString() helper method.
// Returns the string itself (identity function).
func (e *Evaluator) evalStringToString(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 0 {
		return e.newError(node, "String.ToString does not take arguments")
	}

	strVal, ok := selfValue.(*runtime.StringValue)
	if !ok {
		return e.newError(node, "String.ToString requires string receiver")
	}

	// Identity - return the same string value
	return strVal
}
