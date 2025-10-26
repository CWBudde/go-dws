// Package interp provides array manipulation functions for the DWScript interpreter.
package interp

// builtinArrayCopy creates a deep copy of an array.
// Task 9.67: Implement Copy(arr) for arrays
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
