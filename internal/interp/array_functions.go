// Package interp provides array manipulation functions for the DWScript interpreter.
package interp

import (
	"sort"

	"github.com/cwbudde/go-dws/internal/interp/evaluator"
)

// builtinArrayCopy creates a deep copy of an array.
//
// For dynamic arrays: creates new array with same elements
// For static arrays: copies elements to new array
// For arrays of objects: shallow copy references (as per spec)
func (i *Interpreter) builtinArrayCopy(arr *ArrayValue) Value {
	return evaluator.ArrayHelperCopy(arr)
}

// builtinArrayIndexOf searches an array for a value and returns its 0-based index.
//
// Returns 0-based index of first occurrence (0 = first element)
// Returns -1 if not found or invalid startIndex
// Uses 0-based indexing (standard for dynamic arrays in Pascal/Delphi)
func (i *Interpreter) builtinArrayIndexOf(arr *ArrayValue, value Value, startIndex int) Value {
	return evaluator.ArrayHelperIndexOf(arr, value, startIndex)
}

// builtinArrayContains checks if an array contains a specific value.
//
// Returns true if value is found in array, false otherwise
// Uses builtinArrayIndexOf internally
func (i *Interpreter) builtinArrayContains(arr *ArrayValue, value Value) Value {
	return evaluator.ArrayHelperContains(arr, value)
}

// builtinArrayReverse reverses an array in place.
//
// Modifies array by reversing elements in place
// Swaps elements from both ends moving inward
// Returns nil
func (i *Interpreter) builtinArrayReverse(arr *ArrayValue) Value {
	return evaluator.ArrayHelperReverse(arr)
}

// builtinArraySort sorts an array in place using default comparison.
//
// Sorts integers numerically, strings lexicographically
// Uses Go's sort.Slice() for efficient sorting
// Returns nil
func (i *Interpreter) builtinArraySort(arr *ArrayValue) Value {
	return evaluator.ArrayHelperSort(arr)
}

// builtinArraySortWithComparator sorts an array in place using a custom comparator function.
//
// The comparator function must:
// - Accept 2 parameters of the same type as the array elements
// - Return Integer: -1 (a < b), 0 (a == b), 1 (a > b)
//
// Uses Go's sort.Slice() with custom comparison function
// Returns nil
func (i *Interpreter) builtinArraySortWithComparator(arr *ArrayValue, comparator *FunctionPointerValue) Value {
	elements := arr.Elements
	n := len(elements)

	// Empty or single element arrays are already sorted
	if n <= 1 {
		return &NilValue{}
	}

	// Validate comparator signature
	var paramCount int
	if comparator.Function != nil {
		paramCount = len(comparator.Function.Parameters)
	} else if comparator.Lambda != nil {
		paramCount = len(comparator.Lambda.Parameters)
	}

	if paramCount != 2 {
		return i.newErrorWithLocation(i.currentNode, "Sort() comparator must accept 2 parameters, got %d", paramCount)
	}

	// Sort using the comparator
	var sortErr Value
	sort.Slice(elements, func(idx1, idx2 int) bool {
		// If we've already encountered an error, don't continue sorting
		if sortErr != nil {
			return false
		}

		// Call comparator with the two elements
		callArgs := []Value{elements[idx1], elements[idx2]}
		result := i.callFunctionPointer(comparator, callArgs, i.currentNode)

		// Check for error
		if _, isErr := result.(*ErrorValue); isErr {
			sortErr = result
			return false
		}

		// Result must be an integer
		intResult, ok := result.(*IntegerValue)
		if !ok {
			sortErr = i.newErrorWithLocation(i.currentNode, "Sort() comparator must return Integer, got %s", result.Type())
			return false
		}

		// Return true if first element should come before second (a < b means -1)
		return intResult.Value < 0
	})

	// If an error occurred during sorting, return it
	if sortErr != nil {
		return sortErr
	}

	return &NilValue{}
}
