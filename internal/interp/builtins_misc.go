package interp

import (
	"github.com/cwbudde/go-dws/pkg/ast"
)

// Miscellaneous Built-in Functions
// Length, Copy, array operations, debugging, utilities

// builtinLength implements the Length() built-in function.
// It returns the number of elements in an array or characters in a string.
func (i *Interpreter) builtinLength(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Length() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Handle array values
	if arrayVal, ok := arg.(*ArrayValue); ok {
		// Return the number of elements in the array
		// For both static and dynamic arrays, this is len(Elements)
		return &IntegerValue{Value: int64(len(arrayVal.Elements))}
	}

	// Handle string values
	if strVal, ok := arg.(*StringValue); ok {
		// Return the number of characters in the string (not byte length)
		return &IntegerValue{Value: int64(runeLength(strVal.Value))}
	}

	return i.newErrorWithLocation(i.currentNode, "Length() expects array or string, got %s", arg.Type())
}

// builtinCopy implements the Copy() built-in function.
// It returns a substring of a string.
// Copy(str, index) - index is 1-based, copies from index to end
// Copy(str, index, count) - index is 1-based, count is number of characters
// Copy(arr) - creates a deep copy of an array
func (i *Interpreter) builtinCopy(args []Value) Value {
	// Handle array copy: Copy(arr) - 1 argument
	if len(args) == 1 {
		if arrVal, ok := args[0].(*ArrayValue); ok {
			return i.builtinArrayCopy(arrVal)
		}
		return i.newErrorWithLocation(i.currentNode, "Copy() with 1 argument expects array, got %s", args[0].Type())
	}

	// Handle string copy: Copy(str, index) or Copy(str, index, count)
	if len(args) < 2 || len(args) > 3 {
		return i.newErrorWithLocation(i.currentNode, "Copy() expects 1 argument (array) or 2-3 arguments (string), got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Copy() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: index (1-based)
	indexVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Copy() expects integer as second argument, got %s", args[1].Type())
	}

	str := strVal.Value
	index := indexVal.Value // 1-based

	// Third argument: count (optional, defaults to rest of string)
	var count = int64(len([]rune(str))) // Default: copy to end
	if len(args) == 3 {
		countVal, ok := args[2].(*IntegerValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "Copy() expects integer as third argument, got %s", args[2].Type())
		}
		count = countVal.Value
	}

	// Use rune-based slicing to handle UTF-8 correctly
	result := runeSliceFrom(str, int(index), int(count))
	return &StringValue{Value: result}
}

// builtinIndexOf implements the IndexOf() built-in function for arrays.
//
// Returns 0-based index of first occurrence (0 = first element)
// Returns -1 if not found
func (i *Interpreter) builtinIndexOf(args []Value) Value {
	// Validate argument count: 2 or 3 arguments
	if len(args) < 2 || len(args) > 3 {
		return i.newErrorWithLocation(i.currentNode, "IndexOf() expects 2 or 3 arguments, got %d", len(args))
	}

	// First argument must be array
	arr, ok := args[0].(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "IndexOf() expects array as first argument, got %s", args[0].Type())
	}

	// Second argument is the value to search for (any type)
	searchValue := args[1]

	// Third argument (optional) is start index (0-based for internal use)
	startIndex := 0
	if len(args) == 3 {
		startIndexVal, ok := args[2].(*IntegerValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "IndexOf() expects integer as third argument, got %s", args[2].Type())
		}
		startIndex = int(startIndexVal.Value)
	}

	return i.builtinArrayIndexOf(arr, searchValue, startIndex)
}

// builtinContains implements the Contains() built-in function for arrays.
// Contains(arr, value)
//
// Returns true if array contains value, false otherwise
func (i *Interpreter) builtinContains(args []Value) Value {
	// Validate argument count: 2 arguments
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "Contains() expects 2 arguments, got %d", len(args))
	}

	// First argument must be array
	arr, ok := args[0].(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Contains() expects array as first argument, got %s", args[0].Type())
	}

	// Second argument is the value to search for (any type)
	searchValue := args[1]

	return i.builtinArrayContains(arr, searchValue)
}

// builtinReverse implements the Reverse() built-in function for arrays.
// Reverse(arr)
//
// Reverses array elements in place
func (i *Interpreter) builtinReverse(args []Value) Value {
	// Validate argument count: 1 argument
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Reverse() expects 1 argument, got %d", len(args))
	}

	// First argument must be array
	arr, ok := args[0].(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Reverse() expects array as argument, got %s", args[0].Type())
	}

	return i.builtinArrayReverse(arr)
}

// builtinSort implements the Sort() built-in function for arrays.
//
// Sorts array elements in place using default comparison or custom comparator
func (i *Interpreter) builtinSort(args []Value) Value {
	// Validate argument count: 1 or 2 arguments
	if len(args) < 1 || len(args) > 2 {
		return i.newErrorWithLocation(i.currentNode, "Sort() expects 1 or 2 arguments, got %d", len(args))
	}

	// First argument must be array
	arr, ok := args[0].(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Sort() expects array as first argument, got %s", args[0].Type())
	}

	// If only 1 argument, use default sorting
	if len(args) == 1 {
		return i.builtinArraySort(arr)
	}

	// Second argument must be a comparator function pointer
	comparator, ok := args[1].(*FunctionPointerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Sort() expects function pointer as second argument, got %s", args[1].Type())
	}

	return i.builtinArraySortWithComparator(arr, comparator)
}

// builtinSetLength implements the SetLength() built-in function for AST expressions (var-param version).
// It resizes a dynamic array or string to the specified length.
func (i *Interpreter) builtinSetLength(args []ast.Expression) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "SetLength() expects exactly 2 arguments, got %d", len(args))
	}

	// Use evaluateLValue to support identifiers, indexed arrays, member access, etc.
	currentVal, assignFunc, err := i.evaluateLValue(args[0])
	if err != nil {
		return i.newErrorWithLocation(i.currentNode, "SetLength() first argument must be a variable: %s", err.Error())
	}

	// Dereference if it's a var parameter (ReferenceValue)
	if ref, isRef := currentVal.(*ReferenceValue); isRef {
		actualVal, err := ref.Dereference()
		if err != nil {
			return i.newErrorWithLocation(i.currentNode, "%s", err.Error())
		}
		currentVal = actualVal
	}

	// Evaluate the second argument (new length)
	lengthVal := i.Eval(args[1])
	if isError(lengthVal) {
		return lengthVal
	}

	lengthInt, ok := lengthVal.(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "SetLength() expects integer as second argument, got %s", lengthVal.Type())
	}

	newLength := int(lengthInt.Value)
	if newLength < 0 {
		return i.newErrorWithLocation(i.currentNode, "SetLength() expects non-negative length, got %d", newLength)
	}

	// Handle arrays
	if arrayVal, ok := currentVal.(*ArrayValue); ok {
		// Check that it's a dynamic array
		if arrayVal.ArrayType == nil {
			return i.newErrorWithLocation(i.currentNode, "array has no type information")
		}

		if arrayVal.ArrayType.IsStatic() {
			return i.newErrorWithLocation(i.currentNode, "SetLength() can only be used with dynamic arrays, not static arrays")
		}

		currentLength := len(arrayVal.Elements)

		if newLength != currentLength {
			if newLength < currentLength {
				// Truncate the slice
				arrayVal.Elements = arrayVal.Elements[:newLength]
			} else {
				// Extend the slice with nil values
				additional := make([]Value, newLength-currentLength)
				arrayVal.Elements = append(arrayVal.Elements, additional...)
			}
		}

		return &NilValue{}
	}

	// Handle strings
	if strVal, ok := currentVal.(*StringValue); ok {
		// Use rune-based SetLength to handle UTF-8 correctly
		// This truncates or pads with spaces to match DWScript behavior
		newStr := runeSetLength(strVal.Value, newLength)

		// Create new StringValue
		newValue := &StringValue{Value: newStr}

		// Use the assignment function to update the string
		if err := assignFunc(newValue); err != nil {
			return i.newErrorWithLocation(i.currentNode, "failed to update string variable: %s", err)
		}

		return &NilValue{}
	}

	return i.newErrorWithLocation(i.currentNode, "SetLength() expects array or string as first argument, got %s", currentVal.Type())
}

// builtinAdd implements the Add() built-in function.
// It appends an element to the end of a dynamic array.
func (i *Interpreter) builtinAdd(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "Add() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument must be a dynamic array
	arrayArg := args[0]
	arrayVal, ok := arrayArg.(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Add() expects array as first argument, got %s", arrayArg.Type())
	}

	// Check that it's a dynamic array
	if arrayVal.ArrayType == nil {
		return i.newErrorWithLocation(i.currentNode, "array has no type information")
	}

	if arrayVal.ArrayType.IsStatic() {
		return i.newErrorWithLocation(i.currentNode, "Add() can only be used with dynamic arrays, not static arrays")
	}

	// Second argument is the element to add
	element := args[1]

	// Append the element to the array
	arrayVal.Elements = append(arrayVal.Elements, element)

	return &NilValue{}
}

// builtinDelete implements the Delete() built-in function.
// It removes an element at the specified index from a dynamic array.
func (i *Interpreter) builtinDelete(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "Delete() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument must be a dynamic array
	arrayArg := args[0]
	arrayVal, ok := arrayArg.(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Delete() expects array as first argument, got %s", arrayArg.Type())
	}

	// Check that it's a dynamic array
	if arrayVal.ArrayType == nil {
		return i.newErrorWithLocation(i.currentNode, "array has no type information")
	}

	if arrayVal.ArrayType.IsStatic() {
		return i.newErrorWithLocation(i.currentNode, "Delete() can only be used with dynamic arrays, not static arrays")
	}

	// Second argument must be an integer (the index)
	indexArg := args[1]
	indexInt, ok := indexArg.(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Delete() expects integer as second argument, got %s", indexArg.Type())
	}

	index := int(indexInt.Value)

	// Validate index bounds (0-based for dynamic arrays)
	if index < 0 || index >= len(arrayVal.Elements) {
		return i.newErrorWithLocation(i.currentNode, "Delete() index out of bounds: %d (array length is %d)", index, len(arrayVal.Elements))
	}

	// Remove the element at index by slicing
	// Create a new slice with the element removed
	arrayVal.Elements = append(arrayVal.Elements[:index], arrayVal.Elements[index+1:]...)

	return &NilValue{}
}

// builtinGetStackTrace implements the GetStackTrace() built-in function.
// It returns the current call stack as a formatted string.
// GetStackTrace(): String
func (i *Interpreter) builtinGetStackTrace(args []Value) Value {
	if len(args) != 0 {
		return i.newErrorWithLocation(i.currentNode, "GetStackTrace() expects no arguments, got %d", len(args))
	}

	// Return the current call stack as a string
	// The String() method of StackTrace formats it nicely with one frame per line
	return &StringValue{Value: i.callStack.String()}
}

// builtinGetCallStack implements the GetCallStack() built-in function.
// It returns the current call stack as an array of records containing frame information.
// GetCallStack(): array of record
func (i *Interpreter) builtinGetCallStack(args []Value) Value {
	if len(args) != 0 {
		return i.newErrorWithLocation(i.currentNode, "GetCallStack() expects no arguments, got %d", len(args))
	}

	// Create an array to hold the stack frames
	frames := make([]Value, 0, len(i.callStack))

	// Convert each stack frame to a record
	for _, frame := range i.callStack {
		// Create a record value with FunctionName, FileName, and Line fields
		recordFields := make(map[string]Value)
		recordFields["FunctionName"] = &StringValue{Value: frame.FunctionName}
		recordFields["FileName"] = &StringValue{Value: frame.FileName}

		// Set line and column information
		if frame.Position != nil {
			recordFields["Line"] = &IntegerValue{Value: int64(frame.Position.Line)}
			recordFields["Column"] = &IntegerValue{Value: int64(frame.Position.Column)}
		} else {
			recordFields["Line"] = &IntegerValue{Value: 0}
			recordFields["Column"] = &IntegerValue{Value: 0}
		}

		// Create a record value
		// Note: We use a simple map-based structure since we don't have a formal record type here
		// In practice, this will be accessed via dynamic field access
		recordValue := &RecordValue{
			Fields: recordFields,
		}

		frames = append(frames, recordValue)
	}

	// Return as a dynamic array
	return &ArrayValue{
		Elements: frames,
	}
}

// builtinAssigned implements the Assigned() built-in function.
// It checks if a pointer/object/variant is nil.
// Returns false if nil, true otherwise.
// Assigned(value): Boolean
func (i *Interpreter) builtinAssigned(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Assigned() expects exactly 1 argument, got %d", len(args))
	}

	val := args[0]

	// Check if the value is nil or represents a nil value
	switch v := val.(type) {
	case *NilValue:
		return &BooleanValue{Value: false}
	case *ObjectInstance:
		// Check if object is nil (nil object reference)
		if v == nil {
			return &BooleanValue{Value: false}
		}
		return &BooleanValue{Value: true}
	case *VariantValue:
		// Variant is assigned if it's not nil
		if v.Value == nil {
			return &BooleanValue{Value: false}
		}
		// Check if the variant contains a nil value
		if _, isNil := v.Value.(*NilValue); isNil {
			return &BooleanValue{Value: false}
		}
		return &BooleanValue{Value: true}
	case *ArrayValue:
		// Arrays are assigned unless they're nil
		if v == nil {
			return &BooleanValue{Value: false}
		}
		return &BooleanValue{Value: true}
	default:
		// All other types are considered assigned if they're not nil
		if val == nil {
			return &BooleanValue{Value: false}
		}
		return &BooleanValue{Value: true}
	}
}

// builtinSwap implements the Swap() built-in function.
// It swaps the values of two variables: Swap(var a, var b)
func (i *Interpreter) builtinSwap(args []ast.Expression) Value {
	// Validate argument count (exactly 2 arguments)
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "Swap() expects exactly 2 arguments, got %d", len(args))
	}

	// Both arguments must be identifiers (variable names)
	var1Ident, ok1 := args[0].(*ast.Identifier)
	var2Ident, ok2 := args[1].(*ast.Identifier)

	if !ok1 {
		return i.newErrorWithLocation(i.currentNode, "Swap() first argument must be a variable, got %T", args[0])
	}
	if !ok2 {
		return i.newErrorWithLocation(i.currentNode, "Swap() second argument must be a variable, got %T", args[1])
	}

	var1Name := var1Ident.Value
	var2Name := var2Ident.Value

	// Get current values from environment
	val1, exists1 := i.env.Get(var1Name)
	if !exists1 {
		return i.newErrorWithLocation(i.currentNode, "undefined variable: %s", var1Name)
	}

	val2, exists2 := i.env.Get(var2Name)
	if !exists2 {
		return i.newErrorWithLocation(i.currentNode, "undefined variable: %s", var2Name)
	}

	// Handle var parameters (ReferenceValue) for first variable
	var ref1 *ReferenceValue
	if ref, isRef := val1.(*ReferenceValue); isRef {
		ref1 = ref
		// Dereference to get the actual value
		actualVal, err := ref.Dereference()
		if err != nil {
			return i.newErrorWithLocation(i.currentNode, "%s", err.Error())
		}
		val1 = actualVal
	}

	// Handle var parameters (ReferenceValue) for second variable
	var ref2 *ReferenceValue
	if ref, isRef := val2.(*ReferenceValue); isRef {
		ref2 = ref
		// Dereference to get the actual value
		actualVal, err := ref.Dereference()
		if err != nil {
			return i.newErrorWithLocation(i.currentNode, "%s", err.Error())
		}
		val2 = actualVal
	}

	// Swap the values
	// If first variable is a var parameter, write through the reference
	if ref1 != nil {
		if err := ref1.Assign(val2); err != nil {
			return i.newErrorWithLocation(i.currentNode, "failed to update variable %s: %s", var1Name, err)
		}
	} else {
		if err := i.env.Set(var1Name, val2); err != nil {
			return i.newErrorWithLocation(i.currentNode, "failed to update variable %s: %s", var1Name, err)
		}
	}

	// If second variable is a var parameter, write through the reference
	if ref2 != nil {
		if err := ref2.Assign(val1); err != nil {
			return i.newErrorWithLocation(i.currentNode, "failed to update variable %s: %s", var2Name, err)
		}
	} else {
		if err := i.env.Set(var2Name, val1); err != nil {
			return i.newErrorWithLocation(i.currentNode, "failed to update variable %s: %s", var2Name, err)
		}
	}

	return &NilValue{}
}
