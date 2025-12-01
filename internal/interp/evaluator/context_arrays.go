package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Array Construction Methods
// ============================================================================
//
// This file implements the array construction and operation methods of the
// builtins.Context interface for the Evaluator:
//
// Construction:
// - CreateStringArray(): Create array of strings from []string
// - CreateVariantArray(): Create array of Variants from []Value
//
// Operations:
// - GetBuiltinArrayLength(): Get array length
// - SetArrayLength(): Resize dynamic arrays
// - ArrayCopy(): Deep copy array
// - ArrayReverse(): Reverse array in place
// - ArraySort(): Sort array in place
//
// These methods are used by JSON and other built-in functions to construct
// and manipulate arrays dynamically.
// ============================================================================

// CreateStringArray creates an array of strings from a slice of string values.
// Used by JSON functions to return string arrays (e.g., JSONKeys).
//
// This implements the builtins.Context interface method CreateStringArray().
func (e *Evaluator) CreateStringArray(values []string) Value {
	// Convert strings to StringValue elements (as runtime.Value)
	elements := make([]runtime.Value, len(values))
	for idx, str := range values {
		elements[idx] = &runtime.StringValue{Value: str}
	}

	// Create array type: array of String
	arrayType := types.NewDynamicArrayType(types.STRING)

	return &runtime.ArrayValue{
		Elements:  elements,
		ArrayType: arrayType,
	}
}

// CreateVariantArray creates an array of Variants from a slice of values.
// Used by JSON functions to return variant arrays (e.g., JSONValues).
//
// This implements the builtins.Context interface method CreateVariantArray().
func (e *Evaluator) CreateVariantArray(values []Value) Value {
	// Convert evaluator.Value slice to runtime.Value slice
	elements := make([]runtime.Value, len(values))
	for idx, val := range values {
		elements[idx] = val.(runtime.Value)
	}

	// Create array type: array of Variant
	arrayType := types.NewDynamicArrayType(types.VARIANT)

	return &runtime.ArrayValue{
		Elements:  elements,
		ArrayType: arrayType,
	}
}

// ============================================================================
// Array Operation Methods
// ============================================================================

// GetBuiltinArrayLength returns the length of an array for builtin functions.
// This implements the builtins.Context interface.
func (e *Evaluator) GetBuiltinArrayLength(value Value) (int64, bool) {
	arrayVal, ok := value.(*runtime.ArrayValue)
	if !ok {
		return 0, false
	}
	return int64(len(arrayVal.Elements)), true
}

// SetArrayLength resizes a dynamic array to the specified length.
// This implements the builtins.Context interface.
func (e *Evaluator) SetArrayLength(array Value, newLength int) error {
	// Handle arrays
	arrayVal, ok := array.(*runtime.ArrayValue)
	if !ok {
		return fmt.Errorf("SetArrayLength() expects array, got %s", array.Type())
	}

	// Check that it's a dynamic array
	if arrayVal.ArrayType == nil {
		return fmt.Errorf("array has no type information")
	}

	if arrayVal.ArrayType.IsStatic() {
		return fmt.Errorf("SetArrayLength() can only be used with dynamic arrays, not static arrays")
	}

	currentLength := len(arrayVal.Elements)

	if newLength != currentLength {
		if newLength < currentLength {
			// Truncate the slice
			arrayVal.Elements = arrayVal.Elements[:newLength]
		} else {
			// Extend the slice with nil values
			additional := make([]runtime.Value, newLength-currentLength)
			arrayVal.Elements = append(arrayVal.Elements, additional...)
		}
	}

	return nil
}

// ArrayCopy creates a deep copy of an array value.
// This implements the builtins.Context interface.
func (e *Evaluator) ArrayCopy(array Value) Value {
	arrayVal, ok := array.(*runtime.ArrayValue)
	if !ok {
		return e.newError(nil, "ArrayCopy() expects array, got %s", array.Type())
	}

	return ArrayHelperCopy(arrayVal)
}

// ArrayReverse reverses the elements of an array in place.
// This implements the builtins.Context interface.
func (e *Evaluator) ArrayReverse(array Value) Value {
	arrayVal, ok := array.(*runtime.ArrayValue)
	if !ok {
		return e.newError(nil, "ArrayReverse() expects array, got %s", array.Type())
	}

	return ArrayHelperReverse(arrayVal)
}

// ArraySort sorts the elements of an array in place using default comparison.
// This implements the builtins.Context interface.
func (e *Evaluator) ArraySort(array Value) Value {
	arrayVal, ok := array.(*runtime.ArrayValue)
	if !ok {
		return e.newError(nil, "ArraySort() expects array, got %s", array.Type())
	}

	return ArrayHelperSort(arrayVal)
}
