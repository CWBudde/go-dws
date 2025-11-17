package builtins

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
)

// Array built-in functions for DWScript.
// This file contains simple array operations: Length, Copy, Low, High, IndexOf,
// Contains, Reverse, Sort, Add, Delete, SetLength.
//
// Task 3.7.7: Migrate array and collection functions to internal/interp/builtins/ package.

// Length returns the number of elements in an array or characters in a string.
//
// Signature: Length(arr/str) -> Integer
// - arr: Array value to measure
// - str: String value to measure
//
// Returns: Number of elements (for arrays) or characters (for strings)
//
// Example:
//
//	var arr := [1, 2, 3];
//	PrintLn(Length(arr)); // Output: 3
func Length(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Length() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Handle arrays via Context helper
	if length, ok := ctx.GetArrayLength(arg); ok {
		// Return integer value
		return &runtime.IntegerValue{Value: length}
	}

	// Handle strings
	if strVal, ok := arg.(*runtime.StringValue); ok {
		// Count runes (characters), not bytes
		length := int64(0)
		for range strVal.Value {
			length++
		}
		return &runtime.IntegerValue{Value: length}
	}

	return ctx.NewError("Length() expects array or string, got %T", arg)
}

// Copy creates a deep copy of an array or returns a substring.
//
// Signature:
//   - Copy(arr) -> array (array copy)
//   - Copy(str, index, count) -> string (substring)
//
// Example:
//
//	var arr := [1, 2, 3];
//	var arr2 := Copy(arr);
//
//	var str := "Hello";
//	var sub := Copy(str, 1, 3); // "Hel"
func Copy(ctx Context, args []Value) Value {
	// Handle array copy: Copy(arr) - 1 argument
	if len(args) == 1 {
		return ctx.ArrayCopy(args[0])
	}

	// Handle string copy: Copy(str, index, count) - 3 arguments
	if len(args) != 3 {
		return ctx.NewError("Copy() expects either 1 argument (array) or 3 arguments (string), got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("Copy() expects string as first argument, got %T", args[0])
	}

	// Second argument: index (1-based)
	var indexInt int64
	if intVal, ok := args[1].(*runtime.IntegerValue); ok {
		indexInt = intVal.Value
	} else if i64, ok := ctx.ToInt64(args[1]); ok {
		indexInt = i64
	} else {
		return ctx.NewError("Copy() expects integer as second argument, got %T", args[1])
	}

	// Third argument: count
	var countInt int64
	if intVal, ok := args[2].(*runtime.IntegerValue); ok {
		countInt = intVal.Value
	} else if i64, ok := ctx.ToInt64(args[2]); ok {
		countInt = i64
	} else {
		return ctx.NewError("Copy() expects integer as third argument, got %T", args[2])
	}

	// Convert string to rune slice for proper indexing
	runes := []rune(strVal.Value)
	strLen := int64(len(runes))

	// Handle 1-based indexing
	index := indexInt - 1
	count := countInt

	// Bounds checking
	if index < 0 || count < 0 {
		return &runtime.StringValue{Value: ""}
	}

	if index >= strLen {
		return &runtime.StringValue{Value: ""}
	}

	// Calculate end position
	end := index + count
	if end > strLen {
		end = strLen
	}

	// Extract substring
	result := string(runes[index:end])
	return &runtime.StringValue{Value: result}
}

// Low returns the lower bound of an array or the lowest value of an enum/type.
//
// Signature: Low(arr/enum/type) -> Value
// - arr: Array value
// - enum: Enum value
// - type: Type meta-value (e.g., Integer, Boolean)
//
// Returns: Lower bound (0 for dynamic arrays, LowBound for static arrays, lowest enum value, or type minimum)
//
// Example:
//
//	var arr: array [1..5] of Integer;
//	PrintLn(Low(arr)); // Output: 1
//	PrintLn(Low(Integer)); // Output: -9223372036854775808
func Low(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Low() expects exactly 1 argument, got %d", len(args))
	}

	// Use the Context helper which handles arrays, enums, and type meta-values
	result, err := ctx.GetLowBound(args[0])
	if err != nil {
		return ctx.NewError("Low() failed: %v", err)
	}
	return result
}

// High returns the upper bound of an array or the highest value of an enum/type.
//
// Signature: High(arr/enum/type) -> Value
// - arr: Array value
// - enum: Enum value
// - type: Type meta-value (e.g., Integer, Boolean)
//
// Returns: Upper bound (Length-1 for dynamic arrays, HighBound for static arrays, highest enum value, or type maximum)
//
// Example:
//
//	var arr: array [1..5] of Integer;
//	PrintLn(High(arr)); // Output: 5
//	PrintLn(High(Boolean)); // Output: true
func High(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("High() expects exactly 1 argument, got %d", len(args))
	}

	// Use the Context helper which handles arrays, enums, and type meta-values
	result, err := ctx.GetHighBound(args[0])
	if err != nil {
		return ctx.NewError("High() failed: %v", err)
	}
	return result
}

// IndexOf returns the index of the first occurrence of a value in an array.
//
// Signature: IndexOf(arr, value[, startIndex]) -> Integer
// - arr: Array to search
// - value: Value to find
// - startIndex: Optional starting index (0-based)
//
// Returns: 0-based index of first occurrence, or -1 if not found
//
// Example:
//
//	var arr := [10, 20, 30, 20];
//	PrintLn(IndexOf(arr, 20)); // Output: 1
//	PrintLn(IndexOf(arr, 20, 2)); // Output: 3
func IndexOf(ctx Context, args []Value) Value {
	// Validate argument count: 2 or 3 arguments
	if len(args) < 2 || len(args) > 3 {
		return ctx.NewError("IndexOf() expects 2 or 3 arguments, got %d", len(args))
	}

	// First argument must be array
	arr, ok := args[0].(*runtime.ArrayValue)
	if !ok {
		return ctx.NewError("IndexOf() expects array as first argument, got %T", args[0])
	}

	// Second argument is the value to search for (any type)
	searchValue := args[1]

	// Third argument (optional) is start index (0-based)
	startIndex := 0
	if len(args) == 3 {
		var startIndexInt int64
		if intVal, ok := args[2].(*runtime.IntegerValue); ok {
			startIndexInt = intVal.Value
		} else if i64, ok := ctx.ToInt64(args[2]); ok {
			startIndexInt = i64
		} else {
			return ctx.NewError("IndexOf() expects integer as third argument, got %T", args[2])
		}
		startIndex = int(startIndexInt)
	}

	// Validate startIndex bounds
	if startIndex < 0 || startIndex >= len(arr.Elements) {
		return &runtime.IntegerValue{Value: -1}
	}

	// Search for the value starting from startIndex
	for i := startIndex; i < len(arr.Elements); i++ {
		if valuesEqual(arr.Elements[i], searchValue) {
			return &runtime.IntegerValue{Value: int64(i)}
		}
	}

	// Not found
	return &runtime.IntegerValue{Value: -1}
}

// Contains checks if an array contains a specific value.
//
// Signature: Contains(arr, value) -> Boolean
// - arr: Array to search
// - value: Value to find
//
// Returns: true if value is found, false otherwise
//
// Example:
//
//	var arr := [10, 20, 30];
//	PrintLn(Contains(arr, 20)); // Output: true
//	PrintLn(Contains(arr, 99)); // Output: false
func Contains(ctx Context, args []Value) Value {
	// Validate argument count: 2 arguments
	if len(args) != 2 {
		return ctx.NewError("Contains() expects 2 arguments, got %d", len(args))
	}

	// First argument must be array
	arr, ok := args[0].(*runtime.ArrayValue)
	if !ok {
		return ctx.NewError("Contains() expects array as first argument, got %T", args[0])
	}

	// Second argument is the value to search for
	searchValue := args[1]

	// Search for the value
	for _, element := range arr.Elements {
		if valuesEqual(element, searchValue) {
			return &runtime.BooleanValue{Value: true}
		}
	}

	return &runtime.BooleanValue{Value: false}
}

// Reverse reverses the elements of an array in place.
//
// Signature: Reverse(arr)
// - arr: Array to reverse (modified in place)
//
// Example:
//
//	var arr := [1, 2, 3];
//	Reverse(arr);
//	// arr is now [3, 2, 1]
func Reverse(ctx Context, args []Value) Value {
	// Validate argument count: 1 argument
	if len(args) != 1 {
		return ctx.NewError("Reverse() expects 1 argument, got %d", len(args))
	}

	// Delegate to Context helper
	return ctx.ArrayReverse(args[0])
}

// Sort sorts the elements of an array in place using default comparison.
//
// Signature: Sort(arr[, comparator])
// - arr: Array to sort (modified in place)
// - comparator: Optional comparison function
//
// Example:
//
//	var arr := [3, 1, 2];
//	Sort(arr);
//	// arr is now [1, 2, 3]
func Sort(ctx Context, args []Value) Value {
	// Validate argument count: 1 or 2 arguments
	if len(args) < 1 || len(args) > 2 {
		return ctx.NewError("Sort() expects 1 or 2 arguments, got %d", len(args))
	}

	// First argument must be array
	arr, ok := args[0].(*runtime.ArrayValue)
	if !ok {
		return ctx.NewError("Sort() expects array as first argument, got %T", args[0])
	}

	// If only 1 argument, use default sorting
	if len(args) == 1 {
		return ctx.ArraySort(args[0])
	}

	// Second argument must be a comparator function pointer
	comparator, ok := args[1].(*runtime.FunctionPointerValue)
	if !ok {
		return ctx.NewError("Sort() expects function pointer as second argument, got %T", args[1])
	}

	// Validate comparator signature - must accept exactly 2 parameters
	var paramCount int
	if comparator.Function != nil {
		paramCount = len(comparator.Function.Parameters)
	} else if comparator.Lambda != nil {
		paramCount = len(comparator.Lambda.Parameters)
	}

	if paramCount != 2 {
		return ctx.NewError("Sort() comparator must accept 2 parameters, got %d", paramCount)
	}

	// Sort with custom comparator using bubble sort
	// (This is simple but not the most efficient; can optimize later)
	elements := arr.Elements
	n := len(elements)

	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			// Call comparator(elements[j], elements[j+1])
			result := ctx.EvalFunctionPointer(comparator, []Value{elements[j], elements[j+1]})

			// Check for error
			if _, isErr := result.(*runtime.ErrorValue); isErr {
				return result
			}

			// Comparator must return integer:
			// < 0 if a < b
			// = 0 if a == b
			// > 0 if a > b
			var cmpInt int64
			if intVal, ok := result.(*runtime.IntegerValue); ok {
				cmpInt = intVal.Value
			} else if i64, ok := ctx.ToInt64(result); ok {
				cmpInt = i64
			} else {
				return ctx.NewError("Sort() comparator must return Integer, got %T", result)
			}

			// If elements[j] > elements[j+1], swap them
			if cmpInt > 0 {
				elements[j], elements[j+1] = elements[j+1], elements[j]
			}
		}
	}

	return &runtime.NilValue{}
}

// Add appends an element to the end of a dynamic array.
//
// Signature: Add(arr, element)
// - arr: Dynamic array to modify
// - element: Element to append
//
// Example:
//
//	var arr: array of Integer;
//	Add(arr, 42);
func Add(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("Add() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument must be a dynamic array
	arrayArg := args[0]
	arrayVal, ok := arrayArg.(*runtime.ArrayValue)
	if !ok {
		return ctx.NewError("Add() expects array as first argument, got %T", arrayArg)
	}

	// Check that it's a dynamic array
	if arrayVal.ArrayType != nil && arrayVal.ArrayType.IsStatic() {
		return ctx.NewError("Add() can only be used with dynamic arrays, not static arrays")
	}

	// Second argument is the element to add
	element := args[1]

	// Append the element to the array
	arrayVal.Elements = append(arrayVal.Elements, element)

	return &runtime.NilValue{}
}

// Delete removes an element at the specified index from a dynamic array.
//
// Signature: Delete(arr, index)
// - arr: Dynamic array to modify
// - index: 0-based index of element to remove
//
// Example:
//
//	var arr := [10, 20, 30];
//	Delete(arr, 1); // Removes 20
func Delete(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("Delete() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument must be a dynamic array
	arrayArg := args[0]
	arrayVal, ok := arrayArg.(*runtime.ArrayValue)
	if !ok {
		return ctx.NewError("Delete() expects array as first argument, got %T", arrayArg)
	}

	// Check that it's a dynamic array
	if arrayVal.ArrayType != nil && arrayVal.ArrayType.IsStatic() {
		return ctx.NewError("Delete() can only be used with dynamic arrays, not static arrays")
	}

	// Second argument must be an integer (the index)
	var indexInt int64
	if intVal, ok := args[1].(*runtime.IntegerValue); ok {
		indexInt = intVal.Value
	} else if i64, ok := ctx.ToInt64(args[1]); ok {
		indexInt = i64
	} else {
		return ctx.NewError("Delete() expects integer as second argument, got %T", args[1])
	}

	index := int(indexInt)

	// Validate index bounds (0-based for dynamic arrays)
	if index < 0 || index >= len(arrayVal.Elements) {
		return ctx.NewError("Delete() index out of bounds: %d (array length is %d)", index, len(arrayVal.Elements))
	}

	// Remove the element at index by slicing
	arrayVal.Elements = append(arrayVal.Elements[:index], arrayVal.Elements[index+1:]...)

	return &runtime.NilValue{}
}

// SetLength resizes a dynamic array or string to the specified length.
//
// Note: This function requires special handling as it modifies variables.
// It should be called through the var-param mechanism in the interpreter.
//
// Signature: SetLength(var arr, newLength)
// - arr: Dynamic array or string variable to resize
// - newLength: New length
//
// Example:
//
//	var arr: array of Integer;
//	SetLength(arr, 10); // arr now has 10 elements
func SetLength(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("SetLength() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument is the array/string (will be modified by interpreter)
	target := args[0]

	// Second argument is the new length
	var lengthInt int64
	if intVal, ok := args[1].(*runtime.IntegerValue); ok {
		lengthInt = intVal.Value
	} else if i64, ok := ctx.ToInt64(args[1]); ok {
		lengthInt = i64
	} else {
		return ctx.NewError("SetLength() expects integer as second argument, got %T", args[1])
	}

	newLength := int(lengthInt)
	if newLength < 0 {
		return ctx.NewError("SetLength() expects non-negative length, got %d", newLength)
	}

	// Delegate to Context helper
	if err := ctx.SetArrayLength(target, newLength); err != nil {
		return ctx.NewError("SetLength() failed: %v", err)
	}

	return &runtime.NilValue{}
}

// ConcatArrays concatenates multiple arrays into a new array.
//
// Signature: Concat(arr1, arr2, ...) -> array
// - arr1, arr2, ...: Arrays to concatenate
//
// Returns: New array containing all elements
//
// Example:
//
//	var a := [1, 2];
//	var b := [3, 4];
//	var c := Concat(a, b); // [1, 2, 3, 4]
func ConcatArrays(ctx Context, args []Value) Value {
	// Validate argument count
	if len(args) < 1 {
		return ctx.NewError("Concat() expects at least 1 argument, got %d", len(args))
	}

	// Collect all elements from all arrays
	var resultElements []Value
	var firstArrayType *types.ArrayType

	for argIdx, arg := range args {
		// Each argument must be an array
		arrayVal, ok := arg.(*runtime.ArrayValue)
		if !ok {
			return ctx.NewError("Concat() argument %d must be an array, got %T", argIdx+1, arg)
		}

		// Store the type of the first array to use for the result
		if firstArrayType == nil && arrayVal.ArrayType != nil {
			firstArrayType = arrayVal.ArrayType
		}

		// Append all elements from this array
		resultElements = append(resultElements, arrayVal.Elements...)
	}

	// Create and return new array with concatenated elements
	return &runtime.ArrayValue{
		Elements:  resultElements,
		ArrayType: firstArrayType,
	}
}

// Slice extracts a portion of an array.
//
// Signature: Slice(arr, start, end) -> array
// - arr: Array to slice
// - start: Starting index (inclusive)
// - end: Ending index (exclusive)
//
// Returns: New array with elements from start to end-1
//
// Example:
//
//	var arr := [1, 2, 3, 4, 5];
//	var s := Slice(arr, 1, 4); // [2, 3, 4]
func Slice(ctx Context, args []Value) Value {
	// Validate argument count
	if len(args) != 3 {
		return ctx.NewError("Slice() expects 3 arguments (array, start, end), got %d", len(args))
	}

	// First argument must be an array
	arrayVal, ok := args[0].(*runtime.ArrayValue)
	if !ok {
		return ctx.NewError("Slice() first argument must be an array, got %T", args[0])
	}

	// Second argument must be an integer (start index)
	var startInt int64
	if intVal, ok := args[1].(*runtime.IntegerValue); ok {
		startInt = intVal.Value
	} else if i64, ok := ctx.ToInt64(args[1]); ok {
		startInt = i64
	} else {
		return ctx.NewError("Slice() second argument (start) must be an Integer, got %T", args[1])
	}

	// Third argument must be an integer (end index)
	var endInt int64
	if intVal, ok := args[2].(*runtime.IntegerValue); ok {
		endInt = intVal.Value
	} else if i64, ok := ctx.ToInt64(args[2]); ok {
		endInt = i64
	} else {
		return ctx.NewError("Slice() third argument (end) must be an Integer, got %T", args[2])
	}

	// Get the low bound of the array
	lowBound := int64(0)
	if arrayVal.ArrayType != nil && arrayVal.ArrayType.LowBound != nil {
		lowBound = int64(*arrayVal.ArrayType.LowBound)
	}

	// Adjust indices to be relative to the array's low bound
	start := int(startInt - lowBound)
	end := int(endInt - lowBound)

	// Validate indices
	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = 0
	}
	if end > len(arrayVal.Elements) {
		end = len(arrayVal.Elements)
	}
	if start > end {
		start = end
	}

	// Extract the slice
	resultElements := make([]Value, end-start)
	copy(resultElements, arrayVal.Elements[start:end])

	// Create and return new array with sliced elements
	return &runtime.ArrayValue{
		Elements:  resultElements,
		ArrayType: arrayVal.ArrayType,
	}
}

// valuesEqual compares two Value instances for equality.
// This is a helper function used by IndexOf and Contains.
func valuesEqual(a, b Value) bool {
	// Handle nil cases
	if a == nil || b == nil {
		return a == b
	}

	// Compare by type
	switch aVal := a.(type) {
	case *runtime.IntegerValue:
		if bVal, ok := b.(*runtime.IntegerValue); ok {
			return aVal.Value == bVal.Value
		}
	case *runtime.FloatValue:
		if bVal, ok := b.(*runtime.FloatValue); ok {
			return aVal.Value == bVal.Value
		}
	case *runtime.StringValue:
		if bVal, ok := b.(*runtime.StringValue); ok {
			return aVal.Value == bVal.Value
		}
	case *runtime.BooleanValue:
		if bVal, ok := b.(*runtime.BooleanValue); ok {
			return aVal.Value == bVal.Value
		}
	}

	// Default: not equal
	return false
}
