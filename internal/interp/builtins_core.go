package interp

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
)

// builtinPrintLn implements the PrintLn built-in function.
// It prints all arguments followed by a newline.
// Like DWScript's WriteLn, arguments are concatenated directly.
func (i *Interpreter) builtinPrintLn(args []Value) Value {
	// If output is nil, silently discard output (some tests use New(nil))
	if i.output == nil {
		return &NilValue{}
	}
	for _, arg := range args {
		// Handle nil arguments
		if arg == nil {
			fmt.Fprint(i.output, "<nil>")
		} else {
			fmt.Fprint(i.output, arg.String())
		}
	}
	fmt.Fprintln(i.output)
	return &NilValue{}
}

// builtinPrint implements the Print built-in function.
// It prints all arguments without a newline.
// Like DWScript's Write, arguments are concatenated directly.
func (i *Interpreter) builtinPrint(args []Value) Value {
	// If output is nil, silently discard output (some tests use New(nil))
	if i.output == nil {
		return &NilValue{}
	}
	for _, arg := range args {
		// Handle nil arguments
		if arg == nil {
			fmt.Fprint(i.output, "<nil>")
		} else {
			fmt.Fprint(i.output, arg.String())
		}
	}
	return &NilValue{}
}

// builtinOrd implements the Ord() built-in function.
// It returns the ordinal value of an enum, boolean, or character.
// Task 8.51: Ord() function for enums
func (i *Interpreter) builtinOrd(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Ord() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Handle enum values
	if enumVal, ok := arg.(*EnumValue); ok {
		return &IntegerValue{Value: int64(enumVal.OrdinalValue)}
	}

	// Handle boolean values (False=0, True=1)
	if boolVal, ok := arg.(*BooleanValue); ok {
		if boolVal.Value {
			return &IntegerValue{Value: 1}
		}
		return &IntegerValue{Value: 0}
	}

	// Handle integer values (pass through)
	if intVal, ok := arg.(*IntegerValue); ok {
		return intVal
	}

	return i.newErrorWithLocation(i.currentNode, "Ord() expects enum, boolean, or integer, got %s", arg.Type())
}

// builtinInteger implements the Integer() cast function.
// It converts values to integers.
// Task 8.52: Integer() cast for enums
func (i *Interpreter) builtinInteger(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Integer() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Handle enum values
	if enumVal, ok := arg.(*EnumValue); ok {
		return &IntegerValue{Value: int64(enumVal.OrdinalValue)}
	}

	// Handle integer values (pass through)
	if intVal, ok := arg.(*IntegerValue); ok {
		return intVal
	}

	// Handle float values (truncate)
	if floatVal, ok := arg.(*FloatValue); ok {
		return &IntegerValue{Value: int64(floatVal.Value)}
	}

	// Handle boolean values
	if boolVal, ok := arg.(*BooleanValue); ok {
		if boolVal.Value {
			return &IntegerValue{Value: 1}
		}
		return &IntegerValue{Value: 0}
	}

	return i.newErrorWithLocation(i.currentNode, "Integer() cannot convert %s to integer", arg.Type())
}

// builtinLength implements the Length() built-in function.
// It returns the number of elements in an array or characters in a string.
// Task 8.130: Length() function for arrays
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
		// Return the number of characters in the string
		return &IntegerValue{Value: int64(len(strVal.Value))}
	}

	return i.newErrorWithLocation(i.currentNode, "Length() expects array or string, got %s", arg.Type())
}

// builtinCopy implements the Copy() built-in function.
// It returns a substring of a string.
// Copy(str, index, count) - index is 1-based, count is number of characters
// Copy(arr) - creates a deep copy of an array
// Task 8.183: Copy() function for strings
// Task 9.67: Copy() function for arrays
func (i *Interpreter) builtinCopy(args []Value) Value {
	// Handle array copy: Copy(arr) - 1 argument
	if len(args) == 1 {
		if arrVal, ok := args[0].(*ArrayValue); ok {
			return i.builtinArrayCopy(arrVal)
		}
		return i.newErrorWithLocation(i.currentNode, "Copy() with 1 argument expects array, got %s", args[0].Type())
	}

	// Handle string copy: Copy(str, index, count) - 3 arguments
	if len(args) != 3 {
		return i.newErrorWithLocation(i.currentNode, "Copy() expects either 1 argument (array) or 3 arguments (string), got %d", len(args))
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

	// Third argument: count
	countVal, ok := args[2].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Copy() expects integer as third argument, got %s", args[2].Type())
	}

	str := strVal.Value
	index := indexVal.Value // 1-based
	count := countVal.Value

	// Handle edge cases
	// If index is <= 0, return empty string (1-based indexing, so 0 and negative are invalid)
	if index <= 0 {
		return &StringValue{Value: ""}
	}

	// If count is <= 0, return empty string
	if count <= 0 {
		return &StringValue{Value: ""}
	}

	// Convert to 0-based index for Go
	startIdx := int(index - 1)

	// If start index is beyond string length, return empty string
	if startIdx >= len(str) {
		return &StringValue{Value: ""}
	}

	// Calculate end index
	endIdx := startIdx + int(count)

	// If end index goes beyond string length, adjust it
	if endIdx > len(str) {
		endIdx = len(str)
	}

	// Extract substring
	result := str[startIdx:endIdx]

	return &StringValue{Value: result}
}

// builtinIndexOf implements the IndexOf() built-in function for arrays.
// Tasks 9.69-9.70: IndexOf(arr, value) and IndexOf(arr, value, startIndex)
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
// Task 9.72: Contains(arr, value)
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
// Task 9.74: Reverse(arr)
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
// Task 9.76: Sort(arr) - sorts using default comparison
// Task 9.33: Sort(arr, comparator) - sorts using custom comparator function
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

// builtinLow implements the Low() built-in function.
// It returns the lower bound of an array or the lowest value of an enum type.
// Task 8.132: Low() function for arrays
// Task 9.31: Low() function for enums
// Task 9.134: Low() function for type meta-values
func (i *Interpreter) builtinLow(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Low() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Task 9.134: Handle type meta-values (type names as values)
	if typeMetaVal, ok := arg.(*TypeMetaValue); ok {
		// Handle built-in types
		switch typeMetaVal.TypeInfo {
		case types.INTEGER:
			return &IntegerValue{Value: math.MinInt64}
		case types.FLOAT:
			return &FloatValue{Value: -math.MaxFloat64}
		case types.BOOLEAN:
			return &BooleanValue{Value: false}
		}

		// Handle enum types
		if enumType, ok := typeMetaVal.TypeInfo.(*types.EnumType); ok {
			if len(enumType.OrderedNames) == 0 {
				return i.newErrorWithLocation(i.currentNode, "enum type '%s' has no values", typeMetaVal.TypeName)
			}

			// Return the first enum value
			firstValueName := enumType.OrderedNames[0]
			firstOrdinal := enumType.Values[firstValueName]

			return &EnumValue{
				TypeName:     typeMetaVal.TypeName,
				ValueName:    firstValueName,
				OrdinalValue: firstOrdinal,
			}
		}

		return i.newErrorWithLocation(i.currentNode, "Low() not supported for type %s", typeMetaVal.TypeName)
	}

	// Handle array values
	if arrayVal, ok := arg.(*ArrayValue); ok {
		if arrayVal.ArrayType == nil {
			return i.newErrorWithLocation(i.currentNode, "array has no type information")
		}

		// For static arrays, return LowBound
		// For dynamic arrays, return 0
		if arrayVal.ArrayType.IsStatic() {
			return &IntegerValue{Value: int64(*arrayVal.ArrayType.LowBound)}
		}
		return &IntegerValue{Value: 0}
	}

	// Handle enum values
	if enumVal, ok := arg.(*EnumValue); ok {
		// Look up the enum type metadata
		enumTypeKey := "__enum_type_" + enumVal.TypeName
		typeVal, ok := i.env.Get(enumTypeKey)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "enum type '%s' not found", enumVal.TypeName)
		}

		enumTypeVal, ok := typeVal.(*EnumTypeValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "invalid enum type metadata for '%s'", enumVal.TypeName)
		}

		enumType := enumTypeVal.EnumType
		if len(enumType.OrderedNames) == 0 {
			return i.newErrorWithLocation(i.currentNode, "enum type '%s' has no values", enumVal.TypeName)
		}

		// Return the first enum value
		firstValueName := enumType.OrderedNames[0]
		firstOrdinal := enumType.Values[firstValueName]

		return &EnumValue{
			TypeName:     enumVal.TypeName,
			ValueName:    firstValueName,
			OrdinalValue: firstOrdinal,
		}
	}

	return i.newErrorWithLocation(i.currentNode, "Low() expects array, enum, or type name, got %s", arg.Type())
}

// builtinHigh implements the High() built-in function.
// It returns the upper bound of an array or the highest value of an enum type.
// Task 8.133: High() function for arrays
// Task 9.32: High() function for enums
// Task 9.134: High() function for type meta-values
func (i *Interpreter) builtinHigh(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "High() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Task 9.134: Handle type meta-values (type names as values)
	if typeMetaVal, ok := arg.(*TypeMetaValue); ok {
		// Handle built-in types
		switch typeMetaVal.TypeInfo {
		case types.INTEGER:
			return &IntegerValue{Value: math.MaxInt64}
		case types.FLOAT:
			return &FloatValue{Value: math.MaxFloat64}
		case types.BOOLEAN:
			return &BooleanValue{Value: true}
		}

		// Handle enum types
		if enumType, ok := typeMetaVal.TypeInfo.(*types.EnumType); ok {
			if len(enumType.OrderedNames) == 0 {
				return i.newErrorWithLocation(i.currentNode, "enum type '%s' has no values", typeMetaVal.TypeName)
			}

			// Return the last enum value
			lastValueName := enumType.OrderedNames[len(enumType.OrderedNames)-1]
			lastOrdinal := enumType.Values[lastValueName]

			return &EnumValue{
				TypeName:     typeMetaVal.TypeName,
				ValueName:    lastValueName,
				OrdinalValue: lastOrdinal,
			}
		}

		return i.newErrorWithLocation(i.currentNode, "High() not supported for type %s", typeMetaVal.TypeName)
	}

	// Handle array values
	if arrayVal, ok := arg.(*ArrayValue); ok {
		if arrayVal.ArrayType == nil {
			return i.newErrorWithLocation(i.currentNode, "array has no type information")
		}

		// For static arrays, return HighBound
		// For dynamic arrays, return Length - 1
		if arrayVal.ArrayType.IsStatic() {
			return &IntegerValue{Value: int64(*arrayVal.ArrayType.HighBound)}
		}
		// Dynamic array: High = Length - 1
		return &IntegerValue{Value: int64(len(arrayVal.Elements) - 1)}
	}

	// Handle enum values
	if enumVal, ok := arg.(*EnumValue); ok {
		// Look up the enum type metadata
		enumTypeKey := "__enum_type_" + enumVal.TypeName
		typeVal, ok := i.env.Get(enumTypeKey)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "enum type '%s' not found", enumVal.TypeName)
		}

		enumTypeVal, ok := typeVal.(*EnumTypeValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "invalid enum type metadata for '%s'", enumVal.TypeName)
		}

		enumType := enumTypeVal.EnumType
		if len(enumType.OrderedNames) == 0 {
			return i.newErrorWithLocation(i.currentNode, "enum type '%s' has no values", enumVal.TypeName)
		}

		// Return the last enum value
		lastValueName := enumType.OrderedNames[len(enumType.OrderedNames)-1]
		lastOrdinal := enumType.Values[lastValueName]

		return &EnumValue{
			TypeName:     enumVal.TypeName,
			ValueName:    lastValueName,
			OrdinalValue: lastOrdinal,
		}
	}

	return i.newErrorWithLocation(i.currentNode, "High() expects array, enum, or type name, got %s", arg.Type())
}

// builtinSetLength implements the SetLength() built-in function.
// It resizes a dynamic array to the specified length.
// Task 8.131: SetLength() function for dynamic arrays
func (i *Interpreter) builtinSetLength(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "SetLength() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument must be a dynamic array
	arrayArg := args[0]
	arrayVal, ok := arrayArg.(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "SetLength() expects array as first argument, got %s", arrayArg.Type())
	}

	// Check that it's a dynamic array
	if arrayVal.ArrayType == nil {
		return i.newErrorWithLocation(i.currentNode, "array has no type information")
	}

	if arrayVal.ArrayType.IsStatic() {
		return i.newErrorWithLocation(i.currentNode, "SetLength() can only be used with dynamic arrays, not static arrays")
	}

	// Second argument must be an integer (the new length)
	lengthArg := args[1]
	lengthInt, ok := lengthArg.(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "SetLength() expects integer as second argument, got %s", lengthArg.Type())
	}

	newLength := int(lengthInt.Value)
	if newLength < 0 {
		return i.newErrorWithLocation(i.currentNode, "SetLength() expects non-negative length, got %d", newLength)
	}

	currentLength := len(arrayVal.Elements)

	if newLength == currentLength {
		// No change
		return &NilValue{}
	}

	if newLength < currentLength {
		// Truncate the slice
		arrayVal.Elements = arrayVal.Elements[:newLength]
		return &NilValue{}
	}

	// Extend the slice with nil values
	additional := make([]Value, newLength-currentLength)
	arrayVal.Elements = append(arrayVal.Elements, additional...)

	return &NilValue{}
}

// builtinAdd implements the Add() built-in function.
// It appends an element to the end of a dynamic array.
// Task 8.134: Add() function for dynamic arrays
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
// Task 8.135: Delete() function for dynamic arrays
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

// builtinIntToStr implements the IntToStr() built-in function.
// It converts an integer to its string representation.
// IntToStr(i: Integer): String
// Task 8.187: Type conversion functions
// Task 9.102: Support subrange types (subrange values should be assignable to Integer)
func (i *Interpreter) builtinIntToStr(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "IntToStr() expects exactly 1 argument, got %d", len(args))
	}

	// Argument must be an integer or subrange value
	var intValue int64
	switch v := args[0].(type) {
	case *IntegerValue:
		intValue = v.Value
	case *SubrangeValue:
		// Subrange values are assignable to Integer (coercion)
		intValue = int64(v.Value)
	default:
		return i.newErrorWithLocation(i.currentNode, "IntToStr() expects integer argument, got %s", args[0].Type())
	}

	// Convert integer to string using Go's strconv
	result := fmt.Sprintf("%d", intValue)
	return &StringValue{Value: result}
}

// builtinStrToInt implements the StrToInt() built-in function.
// It converts a string to an integer, raising an error if the string is invalid.
// StrToInt(s: String): Integer
// Task 8.187: Type conversion functions
func (i *Interpreter) builtinStrToInt(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "StrToInt() expects exactly 1 argument, got %d", len(args))
	}

	// Argument must be a string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrToInt() expects string argument, got %s", args[0].Type())
	}

	// Try to parse the string as an integer
	// Use strings.TrimSpace to handle leading/trailing whitespace
	s := strings.TrimSpace(strVal.Value)

	// Use strconv.ParseInt for strict parsing (doesn't accept partial matches)
	intValue, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return i.newErrorWithLocation(i.currentNode, "'%s' is not a valid integer", strVal.Value)
	}

	return &IntegerValue{Value: intValue}
}

// builtinFloatToStr implements the FloatToStr() built-in function.
// It converts a float to its string representation.
// FloatToStr(f: Float): String
// Task 8.187: Type conversion functions
func (i *Interpreter) builtinFloatToStr(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "FloatToStr() expects exactly 1 argument, got %d", len(args))
	}

	// Argument can be Float or Integer (implicit coercion)
	var floatValue float64
	switch v := args[0].(type) {
	case *FloatValue:
		floatValue = v.Value
	case *IntegerValue:
		// Implicit Integer→Float coercion
		floatValue = float64(v.Value)
	default:
		return i.newErrorWithLocation(i.currentNode, "FloatToStr() expects float argument, got %s", args[0].Type())
	}

	// Convert float to string using Go's strconv
	// Use 'g' format for general representation (like DWScript's FloatToStr)
	result := fmt.Sprintf("%g", floatValue)
	return &StringValue{Value: result}
}

// builtinStrToFloat implements the StrToFloat() built-in function.
// It converts a string to a float, raising an error if the string is invalid.
// StrToFloat(s: String): Float
// Task 8.187: Type conversion functions
func (i *Interpreter) builtinStrToFloat(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "StrToFloat() expects exactly 1 argument, got %d", len(args))
	}

	// Argument must be a string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrToFloat() expects string argument, got %s", args[0].Type())
	}

	// Try to parse the string as a float
	s := strings.TrimSpace(strVal.Value)

	// Use strconv.ParseFloat for strict parsing (doesn't accept partial matches)
	floatValue, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return i.newErrorWithLocation(i.currentNode, "'%s' is not a valid float", strVal.Value)
	}

	return &FloatValue{Value: floatValue}
}

// builtinBoolToStr implements the BoolToStr() built-in function.
// It converts a boolean to its string representation ("True" or "False").
// BoolToStr(b: Boolean): String
// Task 9.245: Type conversion functions
func (i *Interpreter) builtinBoolToStr(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "BoolToStr() expects exactly 1 argument, got %d", len(args))
	}

	// Argument must be a boolean
	boolVal, ok := args[0].(*BooleanValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "BoolToStr() expects boolean argument, got %s", args[0].Type())
	}

	// Convert boolean to string
	// DWScript uses "True" and "False" (capitalized)
	if boolVal.Value {
		return &StringValue{Value: "True"}
	}
	return &StringValue{Value: "False"}
}
