package interp

import (
	"github.com/cwbudde/go-dws/internal/interp/evaluator"
)

// builtinMap implements the Map() built-in function.
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
func (i *Interpreter) builtinMap(args []Value) Value {
	// Validate argument count
	if len(args) != 2 {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Map() expects 2 arguments (array, lambda), got %d", len(args))
	}

	// First argument must be an array
	arrayVal, ok := args[0].(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Map() first argument must be an array, got %s", args[0].Type())
	}

	// Second argument must be a function pointer (lambda)
	lambdaVal, ok := args[1].(*FunctionPointerValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Map() second argument must be a lambda/function, got %s", args[1].Type())
	}

	// Create result array with same capacity
	resultElements := make([]Value, len(arrayVal.Elements))

	// Apply lambda to each element
	for idx, element := range arrayVal.Elements {
		// Call the lambda with the current element
		callArgs := []Value{element}
		result := i.callFunctionPointer(lambdaVal, callArgs, i.evaluatorInstance.CurrentNode())

		// Check for errors
		if isError(result) {
			return result
		}

		// Store the transformed value
		resultElements[idx] = result
	}

	// Create and return new array with transformed elements
	return &ArrayValue{
		Elements:  resultElements,
		ArrayType: arrayVal.ArrayType,
	}
}

// builtinFilter implements the Filter() built-in function.
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
func (i *Interpreter) builtinFilter(args []Value) Value {
	// Validate argument count
	if len(args) != 2 {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Filter() expects 2 arguments (array, predicate), got %d", len(args))
	}

	// First argument must be an array
	arrayVal, ok := args[0].(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Filter() first argument must be an array, got %s", args[0].Type())
	}

	// Second argument must be a function pointer (lambda)
	predicateVal, ok := args[1].(*FunctionPointerValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Filter() second argument must be a lambda/function, got %s", args[1].Type())
	}

	// Create result array (will grow as needed)
	var resultElements []Value

	// Apply predicate to each element
	for _, element := range arrayVal.Elements {
		// Call the predicate with the current element
		callArgs := []Value{element}
		result := i.callFunctionPointer(predicateVal, callArgs, i.evaluatorInstance.CurrentNode())

		// Check for errors
		if isError(result) {
			return result
		}

		// Check for nil result
		if result == nil {
			return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Filter() predicate returned nil")
		}

		// Predicate must return boolean
		boolResult, ok := result.(*BooleanValue)
		if !ok {
			return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Filter() predicate must return Boolean, got %s", result.Type())
		}

		// If predicate is true, keep this element
		if boolResult.Value {
			resultElements = append(resultElements, element)
		}
	}

	// Create and return new array with filtered elements
	return &ArrayValue{
		Elements:  resultElements,
		ArrayType: arrayVal.ArrayType,
	}
}

// builtinReduce implements the Reduce() built-in function.
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
func (i *Interpreter) builtinReduce(args []Value) Value {
	// Validate argument count
	if len(args) != 3 {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Reduce() expects 3 arguments (array, lambda, initial), got %d", len(args))
	}

	// First argument must be an array
	arrayVal, ok := args[0].(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Reduce() first argument must be an array, got %s", args[0].Type())
	}

	// Second argument must be a function pointer (lambda)
	lambdaVal, ok := args[1].(*FunctionPointerValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Reduce() second argument must be a lambda/function, got %s", args[1].Type())
	}

	// Third argument is the initial accumulator value
	accumulator := args[2]

	// Apply lambda to each element with accumulator
	for _, element := range arrayVal.Elements {
		// Call the lambda with (accumulator, element)
		callArgs := []Value{accumulator, element}
		result := i.callFunctionPointer(lambdaVal, callArgs, i.evaluatorInstance.CurrentNode())

		// Check for errors
		if isError(result) {
			return result
		}

		// Update accumulator with result
		accumulator = result
	}

	// Return final accumulated value
	return accumulator
}

// builtinForEach implements the ForEach() built-in function.
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
func (i *Interpreter) builtinForEach(args []Value) Value {
	// Validate argument count
	if len(args) != 2 {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "ForEach() expects 2 arguments (array, lambda), got %d", len(args))
	}

	// First argument must be an array
	arrayVal, ok := args[0].(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "ForEach() first argument must be an array, got %s", args[0].Type())
	}

	// Second argument must be a function pointer (lambda)
	lambdaVal, ok := args[1].(*FunctionPointerValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "ForEach() second argument must be a lambda/function, got %s", args[1].Type())
	}

	// Execute lambda for each element
	for _, element := range arrayVal.Elements {
		// Call the lambda with the current element
		callArgs := []Value{element}
		result := i.callFunctionPointer(lambdaVal, callArgs, i.evaluatorInstance.CurrentNode())

		// Check for errors
		if isError(result) {
			return result
		}

		// Check if exception was raised
		if i.exception != nil {
			return &NilValue{} // Exception propagation
		}
	}

	// ForEach returns nil (used for side effects)
	return &NilValue{}
}

// builtinEvery implements the Every() built-in function.
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
func (i *Interpreter) builtinEvery(args []Value) Value {
	// Validate argument count
	if len(args) != 2 {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Every() expects 2 arguments (array, predicate), got %d", len(args))
	}

	// First argument must be an array
	arrayVal, ok := args[0].(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Every() first argument must be an array, got %s", args[0].Type())
	}

	// Second argument must be a function pointer (lambda)
	predicateVal, ok := args[1].(*FunctionPointerValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Every() second argument must be a lambda/function, got %s", args[1].Type())
	}

	// Apply predicate to each element, short-circuit on first false
	for _, element := range arrayVal.Elements {
		// Call the predicate with the current element
		callArgs := []Value{element}
		result := i.callFunctionPointer(predicateVal, callArgs, i.evaluatorInstance.CurrentNode())

		// Check for errors
		if isError(result) {
			return result
		}

		// Check if exception was raised
		if i.exception != nil {
			return &BooleanValue{Value: false}
		}

		// Check for nil result
		if result == nil {
			return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Every() predicate returned nil")
		}

		// Predicate must return boolean
		boolResult, ok := result.(*BooleanValue)
		if !ok {
			return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Every() predicate must return Boolean, got %s", result.Type())
		}

		// If predicate is false, return false immediately (short-circuit)
		if !boolResult.Value {
			return &BooleanValue{Value: false}
		}
	}

	// All elements passed the predicate
	return &BooleanValue{Value: true}
}

// builtinSome implements the Some() built-in function.
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
func (i *Interpreter) builtinSome(args []Value) Value {
	// Validate argument count
	if len(args) != 2 {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Some() expects 2 arguments (array, predicate), got %d", len(args))
	}

	// First argument must be an array
	arrayVal, ok := args[0].(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Some() first argument must be an array, got %s", args[0].Type())
	}

	// Second argument must be a function pointer (lambda)
	predicateVal, ok := args[1].(*FunctionPointerValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Some() second argument must be a lambda/function, got %s", args[1].Type())
	}

	// Apply predicate to each element, short-circuit on first true
	for _, element := range arrayVal.Elements {
		// Call the predicate with the current element
		callArgs := []Value{element}
		result := i.callFunctionPointer(predicateVal, callArgs, i.evaluatorInstance.CurrentNode())

		// Check for errors
		if isError(result) {
			return result
		}

		// Check if exception was raised
		if i.exception != nil {
			return &BooleanValue{Value: false}
		}

		// Check for nil result
		if result == nil {
			return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Some() predicate returned nil")
		}

		// Predicate must return boolean
		boolResult, ok := result.(*BooleanValue)
		if !ok {
			return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Some() predicate must return Boolean, got %s", result.Type())
		}

		// If predicate is true, return true immediately (short-circuit)
		if boolResult.Value {
			return &BooleanValue{Value: true}
		}
	}

	// No elements passed the predicate
	return &BooleanValue{Value: false}
}

// builtinFind implements the Find() built-in function.
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
func (i *Interpreter) builtinFind(args []Value) Value {
	// Validate argument count
	if len(args) != 2 {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Find() expects 2 arguments (array, predicate), got %d", len(args))
	}

	// First argument must be an array
	arrayVal, ok := args[0].(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Find() first argument must be an array, got %s", args[0].Type())
	}

	// Second argument must be a function pointer (lambda)
	predicateVal, ok := args[1].(*FunctionPointerValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Find() second argument must be a lambda/function, got %s", args[1].Type())
	}

	// Apply predicate to each element
	for _, element := range arrayVal.Elements {
		// Call the predicate with the current element
		callArgs := []Value{element}
		result := i.callFunctionPointer(predicateVal, callArgs, i.evaluatorInstance.CurrentNode())

		// Check for errors
		if isError(result) {
			return result
		}

		// Check if exception was raised
		if i.exception != nil {
			return &NilValue{}
		}

		// Check for nil result
		if result == nil {
			return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Find() predicate returned nil")
		}

		// Predicate must return boolean
		boolResult, ok := result.(*BooleanValue)
		if !ok {
			return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Find() predicate must return Boolean, got %s", result.Type())
		}

		// If predicate is true, return this element
		if boolResult.Value {
			return element
		}
	}

	// No element found
	return &NilValue{}
}

// builtinFindIndex implements the FindIndex() built-in function.
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
func (i *Interpreter) builtinFindIndex(args []Value) Value {
	// Validate argument count
	if len(args) != 2 {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "FindIndex() expects 2 arguments (array, predicate), got %d", len(args))
	}

	// First argument must be an array
	arrayVal, ok := args[0].(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "FindIndex() first argument must be an array, got %s", args[0].Type())
	}

	// Second argument must be a function pointer (lambda)
	predicateVal, ok := args[1].(*FunctionPointerValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "FindIndex() second argument must be a lambda/function, got %s", args[1].Type())
	}

	// Apply predicate to each element
	for idx, element := range arrayVal.Elements {
		// Call the predicate with the current element
		callArgs := []Value{element}
		result := i.callFunctionPointer(predicateVal, callArgs, i.evaluatorInstance.CurrentNode())

		// Check for errors
		if isError(result) {
			return result
		}

		// Check if exception was raised
		if i.exception != nil {
			return &IntegerValue{Value: -1}
		}

		// Check for nil result
		if result == nil {
			return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "FindIndex() predicate returned nil")
		}

		// Predicate must return boolean
		boolResult, ok := result.(*BooleanValue)
		if !ok {
			return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "FindIndex() predicate must return Boolean, got %s", result.Type())
		}

		// If predicate is true, return this index (adjusted for array's low bound)
		if boolResult.Value {
			// Calculate the actual index based on the array's low bound
			lowBound := int64(0)
			if arrayVal.ArrayType != nil && arrayVal.ArrayType.LowBound != nil {
				lowBound = int64(*arrayVal.ArrayType.LowBound)
			}
			actualIndex := lowBound + int64(idx)
			return &IntegerValue{Value: actualIndex}
		}
	}

	// No element found
	return &IntegerValue{Value: -1}
}

// builtinSlice implements the Slice() built-in function.
//
// Signature: Slice(array, start, end) -> array
// - array: The source array to slice
// - start: Starting index (inclusive)
// - end: Ending index (exclusive)
//
// Returns: New array containing elements from start to end-1
//
// Example:
//
//	var numbers := [1, 2, 3, 4, 5];
//	var slice := Slice(numbers, 1, 4);
//	// Result: [2, 3, 4]
func (i *Interpreter) builtinSlice(args []Value) Value {
	// Validate argument count
	if len(args) != 3 {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Slice() expects 3 arguments (array, start, end), got %d", len(args))
	}

	// First argument must be an array
	arrayVal, ok := args[0].(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Slice() first argument must be an array, got %s", args[0].Type())
	}

	// Second argument must be an integer (start index)
	startVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Slice() second argument (start) must be an Integer, got %s", args[1].Type())
	}

	// Third argument must be an integer (end index)
	endVal, ok := args[2].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "Slice() third argument (end) must be an Integer, got %s", args[2].Type())
	}

	// Delegate to standalone helper function
	return evaluator.ArrayHelperSlice(arrayVal, startVal.Value, endVal.Value)
}
