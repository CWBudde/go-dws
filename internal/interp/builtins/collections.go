package builtins
import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// Collection built-in functions for DWScript.
// This file contains higher-order array functions that work with callbacks:
// Map, Filter, Reduce, ForEach, Every, Some, Find, FindIndex.
//
// Task 3.7.7: Migrate array and collection functions to internal/interp/builtins/ package.

// Map transforms each element of an array using a callback function.
//
// Signature: Map(array, lambda) -> array
// - array: The source array to transform
// - lambda: A function that takes one element and returns the transformed value
//
// Returns: New array with transformed elements
//
// Example:
//
//	var numbers := [1, 2, 3, 4, 5];
//	var doubled := Map(numbers, lambda(x: Integer): Integer => x * 2);
//	// Result: [2, 4, 6, 8, 10]
func Map(ctx Context, args []Value) Value {
	// Validate argument count
	if len(args) != 2 {
		return ctx.NewError("Map() expects 2 arguments (array, lambda), got %d", len(args))
	}

	// First argument must be an array
	arrayVal, ok := args[0].(*runtime.ArrayValue)
	if !ok {
		return ctx.NewError("Map() first argument must be an array, got %T", args[0])
	}

	// Second argument must be a function pointer (lambda)
	lambdaVal, ok := args[1].(*runtime.FunctionPointerValue)
	if !ok {
		return ctx.NewError("Map() second argument must be a lambda/function, got %T", args[1])
	}

	// Create result array with same capacity
	resultElements := make([]Value, len(arrayVal.Elements))

	// Apply lambda to each element
	for idx, element := range arrayVal.Elements {
		// Call the lambda with the current element
		result := ctx.EvalFunctionPointer(lambdaVal, []Value{element})

		// Check for errors
		if _, isErr := result.(*runtime.ErrorValue); isErr {
			return result
		}

		// Store the transformed value
		resultElements[idx] = result
	}

	// Create and return new array with transformed elements
	return &runtime.ArrayValue{
		Elements:  resultElements,
		ArrayType: arrayVal.ArrayType,
	}
}

// Filter creates a new array containing only elements that match a predicate.
//
// Signature: Filter(array, predicate) -> array
// - array: The source array to filter
// - predicate: A function that takes one element and returns Boolean (true to keep)
//
// Returns: New array with only elements where predicate returned true
//
// Example:
//
//	var numbers := [1, 2, 3, 4, 5];
//	var evens := Filter(numbers, lambda(x: Integer): Boolean => (x mod 2) = 0);
//	// Result: [2, 4]
func Filter(ctx Context, args []Value) Value {
	// Validate argument count
	if len(args) != 2 {
		return ctx.NewError("Filter() expects 2 arguments (array, predicate), got %d", len(args))
	}

	// First argument must be an array
	arrayVal, ok := args[0].(*runtime.ArrayValue)
	if !ok {
		return ctx.NewError("Filter() first argument must be an array, got %T", args[0])
	}

	// Second argument must be a function pointer (lambda)
	predicateVal, ok := args[1].(*runtime.FunctionPointerValue)
	if !ok {
		return ctx.NewError("Filter() second argument must be a lambda/function, got %T", args[1])
	}

	// Create result array (will grow as needed)
	var resultElements []Value

	// Apply predicate to each element
	for _, element := range arrayVal.Elements {
		// Call the predicate with the current element
		result := ctx.EvalFunctionPointer(predicateVal, []Value{element})

		// Check for errors
		if _, isErr := result.(*runtime.ErrorValue); isErr {
			return result
		}

		// Check for nil result
		if result == nil {
			return ctx.NewError("Filter() predicate returned nil")
		}

		// Predicate must return boolean
		boolResult, ok := result.(*runtime.BooleanValue)
		if !ok {
			return ctx.NewError("Filter() predicate must return Boolean, got %T", result)
		}

		// If predicate is true, keep this element
		if boolResult.Value {
			resultElements = append(resultElements, element)
		}
	}

	// Create and return new array with filtered elements
	return &runtime.ArrayValue{
		Elements:  resultElements,
		ArrayType: arrayVal.ArrayType,
	}
}

// Reduce reduces an array to a single value using an accumulator function.
//
// Signature: Reduce(array, lambda, initial) -> value
// - array: The source array to reduce
// - lambda: A function that takes (accumulator, element) and returns new accumulator
// - initial: The initial value of the accumulator
//
// Returns: Final accumulated value
//
// Example:
//
//	var numbers := [1, 2, 3, 4, 5];
//	var sum := Reduce(numbers, lambda(acc, x: Integer): Integer => acc + x, 0);
//	// Result: 15
func Reduce(ctx Context, args []Value) Value {
	// Validate argument count
	if len(args) != 3 {
		return ctx.NewError("Reduce() expects 3 arguments (array, lambda, initial), got %d", len(args))
	}

	// First argument must be an array
	arrayVal, ok := args[0].(*runtime.ArrayValue)
	if !ok {
		return ctx.NewError("Reduce() first argument must be an array, got %T", args[0])
	}

	// Second argument must be a function pointer (lambda)
	lambdaVal, ok := args[1].(*runtime.FunctionPointerValue)
	if !ok {
		return ctx.NewError("Reduce() second argument must be a lambda/function, got %T", args[1])
	}

	// Third argument is the initial accumulator value
	accumulator := args[2]

	// Apply lambda to each element with accumulator
	for _, element := range arrayVal.Elements {
		// Call the lambda with (accumulator, element)
		result := ctx.EvalFunctionPointer(lambdaVal, []Value{accumulator, element})

		// Check for errors
		if _, isErr := result.(*runtime.ErrorValue); isErr {
			return result
		}

		// Update accumulator with result
		accumulator = result
	}

	// Return final accumulated value
	return accumulator
}

// ForEach executes a function for each element of an array (for side effects).
//
// Signature: ForEach(array, lambda)
// - array: The source array to iterate
// - lambda: A function that takes one element (return value ignored)
//
// Returns: nil (this function is used for side effects only)
//
// Example:
//
//	var numbers := [1, 2, 3];
//	ForEach(numbers, lambda(x: Integer) begin PrintLn(x); end);
//	// Output: 1\n2\n3
func ForEach(ctx Context, args []Value) Value {
	// Validate argument count
	if len(args) != 2 {
		return ctx.NewError("ForEach() expects 2 arguments (array, lambda), got %d", len(args))
	}

	// First argument must be an array
	arrayVal, ok := args[0].(*runtime.ArrayValue)
	if !ok {
		return ctx.NewError("ForEach() first argument must be an array, got %T", args[0])
	}

	// Second argument must be a function pointer (lambda)
	lambdaVal, ok := args[1].(*runtime.FunctionPointerValue)
	if !ok {
		return ctx.NewError("ForEach() second argument must be a lambda/function, got %T", args[1])
	}

	// Execute lambda for each element
	for _, element := range arrayVal.Elements {
		// Call the lambda with the current element
		result := ctx.EvalFunctionPointer(lambdaVal, []Value{element})

		// Check for errors
		if _, isErr := result.(*runtime.ErrorValue); isErr {
			return result
		}
	}

	// ForEach returns nil (used for side effects)
	return &runtime.NilValue{}
}

// Every checks if all elements of an array match a predicate.
//
// Signature: Every(array, predicate) -> Boolean
// - array: The source array to check
// - predicate: A function that takes one element and returns Boolean
//
// Returns: true if all elements match predicate, false otherwise (short-circuits on first false)
//
// Example:
//
//	var numbers := [2, 4, 6, 8];
//	var allEven := Every(numbers, lambda(x: Integer): Boolean => (x mod 2) = 0);
//	// Result: true
func Every(ctx Context, args []Value) Value {
	// Validate argument count
	if len(args) != 2 {
		return ctx.NewError("Every() expects 2 arguments (array, predicate), got %d", len(args))
	}

	// First argument must be an array
	arrayVal, ok := args[0].(*runtime.ArrayValue)
	if !ok {
		return ctx.NewError("Every() first argument must be an array, got %T", args[0])
	}

	// Second argument must be a function pointer (lambda)
	predicateVal, ok := args[1].(*runtime.FunctionPointerValue)
	if !ok {
		return ctx.NewError("Every() second argument must be a lambda/function, got %T", args[1])
	}

	// Apply predicate to each element, short-circuit on first false
	for _, element := range arrayVal.Elements {
		// Call the predicate with the current element
		result := ctx.EvalFunctionPointer(predicateVal, []Value{element})

		// Check for errors
		if _, isErr := result.(*runtime.ErrorValue); isErr {
			return result
		}

		// Check for nil result
		if result == nil {
			return ctx.NewError("Every() predicate returned nil")
		}

		// Predicate must return boolean
		boolResult, ok := result.(*runtime.BooleanValue)
		if !ok {
			return ctx.NewError("Every() predicate must return Boolean, got %T", result)
		}

		// If predicate is false, return false immediately (short-circuit)
		if !boolResult.Value {
			return &runtime.BooleanValue{Value: false}
		}
	}

	// All elements passed the predicate
	return &runtime.BooleanValue{Value: true}
}

// Some checks if any element of an array matches a predicate.
//
// Signature: Some(array, predicate) -> Boolean
// - array: The source array to check
// - predicate: A function that takes one element and returns Boolean
//
// Returns: true if any element matches predicate, false otherwise (short-circuits on first true)
//
// Example:
//
//	var numbers := [1, 3, 5, 6, 7];
//	var hasEven := Some(numbers, lambda(x: Integer): Boolean => (x mod 2) = 0);
//	// Result: true
func Some(ctx Context, args []Value) Value {
	// Validate argument count
	if len(args) != 2 {
		return ctx.NewError("Some() expects 2 arguments (array, predicate), got %d", len(args))
	}

	// First argument must be an array
	arrayVal, ok := args[0].(*runtime.ArrayValue)
	if !ok {
		return ctx.NewError("Some() first argument must be an array, got %T", args[0])
	}

	// Second argument must be a function pointer (lambda)
	predicateVal, ok := args[1].(*runtime.FunctionPointerValue)
	if !ok {
		return ctx.NewError("Some() second argument must be a lambda/function, got %T", args[1])
	}

	// Apply predicate to each element, short-circuit on first true
	for _, element := range arrayVal.Elements {
		// Call the predicate with the current element
		result := ctx.EvalFunctionPointer(predicateVal, []Value{element})

		// Check for errors
		if _, isErr := result.(*runtime.ErrorValue); isErr {
			return result
		}

		// Check for nil result
		if result == nil {
			return ctx.NewError("Some() predicate returned nil")
		}

		// Predicate must return boolean
		boolResult, ok := result.(*runtime.BooleanValue)
		if !ok {
			return ctx.NewError("Some() predicate must return Boolean, got %T", result)
		}

		// If predicate is true, return true immediately (short-circuit)
		if boolResult.Value {
			return &runtime.BooleanValue{Value: true}
		}
	}

	// No elements passed the predicate
	return &runtime.BooleanValue{Value: false}
}

// Find returns the first element of an array that matches a predicate.
//
// Signature: Find(array, predicate) -> Variant
// - array: The source array to search
// - predicate: A function that takes one element and returns Boolean
//
// Returns: First element matching predicate, or nil if not found
//
// Example:
//
//	var numbers := [1, 2, 3, 4, 5];
//	var found := Find(numbers, lambda(x: Integer): Boolean => x > 3);
//	// Result: 4
func Find(ctx Context, args []Value) Value {
	// Validate argument count
	if len(args) != 2 {
		return ctx.NewError("Find() expects 2 arguments (array, predicate), got %d", len(args))
	}

	// First argument must be an array
	arrayVal, ok := args[0].(*runtime.ArrayValue)
	if !ok {
		return ctx.NewError("Find() first argument must be an array, got %T", args[0])
	}

	// Second argument must be a function pointer (lambda)
	predicateVal, ok := args[1].(*runtime.FunctionPointerValue)
	if !ok {
		return ctx.NewError("Find() second argument must be a lambda/function, got %T", args[1])
	}

	// Apply predicate to each element
	for _, element := range arrayVal.Elements {
		// Call the predicate with the current element
		result := ctx.EvalFunctionPointer(predicateVal, []Value{element})

		// Check for errors
		if _, isErr := result.(*runtime.ErrorValue); isErr {
			return result
		}

		// Check for nil result
		if result == nil {
			return ctx.NewError("Find() predicate returned nil")
		}

		// Predicate must return boolean
		boolResult, ok := result.(*runtime.BooleanValue)
		if !ok {
			return ctx.NewError("Find() predicate must return Boolean, got %T", result)
		}

		// If predicate is true, return this element
		if boolResult.Value {
			return element
		}
	}

	// No element found
	return &runtime.NilValue{}
}

// FindIndex returns the index of the first element that matches a predicate.
//
// Signature: FindIndex(array, predicate) -> Integer
// - array: The source array to search
// - predicate: A function that takes one element and returns Boolean
//
// Returns: Index of first element matching predicate, or -1 if not found
//
// Example:
//
//	var numbers := [1, 2, 3, 4, 5];
//	var index := FindIndex(numbers, lambda(x: Integer): Boolean => x > 3);
//	// Result: 3 (zero-based index of value 4)
func FindIndex(ctx Context, args []Value) Value {
	// Validate argument count
	if len(args) != 2 {
		return ctx.NewError("FindIndex() expects 2 arguments (array, predicate), got %d", len(args))
	}

	// First argument must be an array
	arrayVal, ok := args[0].(*runtime.ArrayValue)
	if !ok {
		return ctx.NewError("FindIndex() first argument must be an array, got %T", args[0])
	}

	// Second argument must be a function pointer (lambda)
	predicateVal, ok := args[1].(*runtime.FunctionPointerValue)
	if !ok {
		return ctx.NewError("FindIndex() second argument must be a lambda/function, got %T", args[1])
	}

	// Apply predicate to each element
	for idx, element := range arrayVal.Elements {
		// Call the predicate with the current element
		result := ctx.EvalFunctionPointer(predicateVal, []Value{element})

		// Check for errors
		if _, isErr := result.(*runtime.ErrorValue); isErr {
			return result
		}

		// Check for nil result
		if result == nil {
			return ctx.NewError("FindIndex() predicate returned nil")
		}

		// Predicate must return boolean
		boolResult, ok := result.(*runtime.BooleanValue)
		if !ok {
			return ctx.NewError("FindIndex() predicate must return Boolean, got %T", result)
		}

		// If predicate is true, return this index (adjusted for array's low bound)
		if boolResult.Value {
			// Calculate the actual index based on the array's low bound
			lowBound := int64(0)
			if arrayVal.ArrayType != nil && arrayVal.ArrayType.LowBound != nil {
				lowBound = int64(*arrayVal.ArrayType.LowBound)
			}
			actualIndex := lowBound + int64(idx)
			return &runtime.IntegerValue{Value: actualIndex}
		}
	}

	// No element found
	return &runtime.IntegerValue{Value: -1}
}
