// Package interp provides array manipulation functions for the DWScript interpreter.
package interp

import (
	"sort"
)

// builtinArrayCopy creates a deep copy of an array.
//
// For dynamic arrays: creates new array with same elements
// For static arrays: copies elements to new array
// For arrays of objects: shallow copy references (as per spec)
func (i *Interpreter) builtinArrayCopy(arr *ArrayValue) Value {
	// Create a new ArrayValue with the same type
	newArray := &ArrayValue{
		ArrayType: arr.ArrayType,
		Elements:  make([]Value, len(arr.Elements)),
	}

	// Deep copy the elements slice
	// Note: For object references, this is a shallow copy (references are copied, not objects)
	copy(newArray.Elements, arr.Elements)

	return newArray
}

// builtinArrayIndexOf searches an array for a value and returns its 0-based index.
//
// Returns 0-based index of first occurrence (0 = first element)
// Returns -1 if not found or invalid startIndex
// Uses 0-based indexing (standard for dynamic arrays in Pascal/Delphi)
func (i *Interpreter) builtinArrayIndexOf(arr *ArrayValue, value Value, startIndex int) Value {
	// Validate startIndex bounds
	if startIndex < 0 || startIndex >= len(arr.Elements) {
		return &IntegerValue{Value: -1}
	}

	// Search array from startIndex onwards
	for idx := startIndex; idx < len(arr.Elements); idx++ {
		if i.valuesEqual(arr.Elements[idx], value) {
			// Return 0-based index
			return &IntegerValue{Value: int64(idx)}
		}
	}

	// Not found
	return &IntegerValue{Value: -1}
}

// builtinArrayContains checks if an array contains a specific value.
//
// Returns true if value is found in array, false otherwise
// Uses builtinArrayIndexOf internally
func (i *Interpreter) builtinArrayContains(arr *ArrayValue, value Value) Value {
	// Use IndexOf to check if value exists
	// IndexOf returns >= 0 if found (0-based indexing), -1 if not found
	result := i.builtinArrayIndexOf(arr, value, 0)
	intResult, ok := result.(*IntegerValue)
	if !ok {
		// Should never happen, but handle error case
		return &BooleanValue{Value: false}
	}

	// Return true if found (index >= 0), false otherwise
	return &BooleanValue{Value: intResult.Value >= 0}
}

// builtinArrayReverse reverses an array in place.
//
// Modifies array by reversing elements in place
// Swaps elements from both ends moving inward
// Returns nil
func (i *Interpreter) builtinArrayReverse(arr *ArrayValue) Value {
	elements := arr.Elements
	n := len(elements)

	// Swap elements from both ends
	for left := 0; left < n/2; left++ {
		right := n - 1 - left
		elements[left], elements[right] = elements[right], elements[left]
	}

	// Return nil (procedure with no return value)
	return &NilValue{}
}

// builtinArraySort sorts an array in place using default comparison.
//
// Sorts integers numerically, strings lexicographically
// Uses Go's sort.Slice() for efficient sorting
// Returns nil
func (i *Interpreter) builtinArraySort(arr *ArrayValue) Value {
	elements := arr.Elements
	n := len(elements)

	// Empty or single element arrays are already sorted
	if n <= 1 {
		return &NilValue{}
	}

	// Determine element type from first element
	firstElem := elements[0]

	// Sort based on element type
	switch firstElem.(type) {
	case *IntegerValue:
		// Numeric sort for integers
		sort.Slice(elements, func(i, j int) bool {
			left, leftOk := elements[i].(*IntegerValue)
			right, rightOk := elements[j].(*IntegerValue)
			if !leftOk || !rightOk {
				return false
			}
			return left.Value < right.Value
		})

	case *FloatValue:
		// Numeric sort for floats
		sort.Slice(elements, func(i, j int) bool {
			left, leftOk := elements[i].(*FloatValue)
			right, rightOk := elements[j].(*FloatValue)
			if !leftOk || !rightOk {
				return false
			}
			return left.Value < right.Value
		})

	case *StringValue:
		// Lexicographic sort for strings
		sort.Slice(elements, func(i, j int) bool {
			left, leftOk := elements[i].(*StringValue)
			right, rightOk := elements[j].(*StringValue)
			if !leftOk || !rightOk {
				return false
			}
			return left.Value < right.Value
		})

	case *BooleanValue:
		// Boolean sort: false < true
		sort.Slice(elements, func(i, j int) bool {
			left, leftOk := elements[i].(*BooleanValue)
			right, rightOk := elements[j].(*BooleanValue)
			if !leftOk || !rightOk {
				return false
			}
			// false (false < true) sorts before true
			return !left.Value && right.Value
		})

	default:
		// For other types, we can't sort - just return nil
		return &NilValue{}
	}

	return &NilValue{}
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
