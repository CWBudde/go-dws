package interp

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
		return i.newErrorWithLocation(i.currentNode, "Map() expects 2 arguments (array, lambda), got %d", len(args))
	}

	// First argument must be an array
	arrayVal, ok := args[0].(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Map() first argument must be an array, got %s", args[0].Type())
	}

	// Second argument must be a function pointer (lambda)
	lambdaVal, ok := args[1].(*FunctionPointerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Map() second argument must be a lambda/function, got %s", args[1].Type())
	}

	// Create result array with same capacity
	resultElements := make([]Value, len(arrayVal.Elements))

	// Apply lambda to each element
	for idx, element := range arrayVal.Elements {
		// Call the lambda with the current element
		callArgs := []Value{element}
		result := i.callFunctionPointer(lambdaVal, callArgs, i.currentNode)

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
		return i.newErrorWithLocation(i.currentNode, "Filter() expects 2 arguments (array, predicate), got %d", len(args))
	}

	// First argument must be an array
	arrayVal, ok := args[0].(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Filter() first argument must be an array, got %s", args[0].Type())
	}

	// Second argument must be a function pointer (lambda)
	predicateVal, ok := args[1].(*FunctionPointerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Filter() second argument must be a lambda/function, got %s", args[1].Type())
	}

	// Create result array (will grow as needed)
	var resultElements []Value

	// Apply predicate to each element
	for _, element := range arrayVal.Elements {
		// Call the predicate with the current element
		callArgs := []Value{element}
		result := i.callFunctionPointer(predicateVal, callArgs, i.currentNode)

		// Check for errors
		if isError(result) {
			return result
		}

		// Check for nil result
		if result == nil {
			return i.newErrorWithLocation(i.currentNode, "Filter() predicate returned nil")
		}

		// Predicate must return boolean
		boolResult, ok := result.(*BooleanValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "Filter() predicate must return Boolean, got %s", result.Type())
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
		return i.newErrorWithLocation(i.currentNode, "Reduce() expects 3 arguments (array, lambda, initial), got %d", len(args))
	}

	// First argument must be an array
	arrayVal, ok := args[0].(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Reduce() first argument must be an array, got %s", args[0].Type())
	}

	// Second argument must be a function pointer (lambda)
	lambdaVal, ok := args[1].(*FunctionPointerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Reduce() second argument must be a lambda/function, got %s", args[1].Type())
	}

	// Third argument is the initial accumulator value
	accumulator := args[2]

	// Apply lambda to each element with accumulator
	for _, element := range arrayVal.Elements {
		// Call the lambda with (accumulator, element)
		callArgs := []Value{accumulator, element}
		result := i.callFunctionPointer(lambdaVal, callArgs, i.currentNode)

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
		return i.newErrorWithLocation(i.currentNode, "ForEach() expects 2 arguments (array, lambda), got %d", len(args))
	}

	// First argument must be an array
	arrayVal, ok := args[0].(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "ForEach() first argument must be an array, got %s", args[0].Type())
	}

	// Second argument must be a function pointer (lambda)
	lambdaVal, ok := args[1].(*FunctionPointerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "ForEach() second argument must be a lambda/function, got %s", args[1].Type())
	}

	// Execute lambda for each element
	for _, element := range arrayVal.Elements {
		// Call the lambda with the current element
		callArgs := []Value{element}
		result := i.callFunctionPointer(lambdaVal, callArgs, i.currentNode)

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
