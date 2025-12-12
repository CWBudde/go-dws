// Package interp provides array manipulation functions for the DWScript interpreter.
package interp

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// builtinArrayCopy creates a deep copy of an array.
//
// For dynamic arrays: creates new array with same elements
// For static arrays: copies elements to new array
// For arrays of objects: shallow copy references (as per spec)
func (i *Interpreter) builtinArrayCopy(arr *ArrayValue) Value {
	return runtime.ArrayHelperCopy(arr)
}

// builtinArrayIndexOf searches an array for a value and returns its 0-based index.
//
// Returns 0-based index of first occurrence (0 = first element)
// Returns -1 if not found or invalid startIndex
// Uses 0-based indexing (standard for dynamic arrays in Pascal/Delphi)
func (i *Interpreter) builtinArrayIndexOf(arr *ArrayValue, value Value, startIndex int) Value {
	return runtime.ArrayHelperIndexOf(arr, value, startIndex)
}

// builtinArrayReverse reverses an array in place.
//
// Modifies array by reversing elements in place
// Swaps elements from both ends moving inward
// Returns nil
func (i *Interpreter) builtinArrayReverse(arr *ArrayValue) Value {
	return runtime.ArrayHelperReverse(arr)
}

// builtinArraySort sorts an array in place using default comparison.
//
// Sorts integers numerically, strings lexicographically
// Uses Go's sort.Slice() for efficient sorting
// Returns nil
func (i *Interpreter) builtinArraySort(arr *ArrayValue) Value {
	return runtime.ArrayHelperSort(arr)
}
