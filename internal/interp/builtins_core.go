package interp

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
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
// It returns the ordinal value of an enum, boolean, character, or integer.
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

	// Handle string values (characters)
	// In DWScript, character literals are single-character strings
	if strVal, ok := arg.(*StringValue); ok {
		if len(strVal.Value) == 0 {
			// Empty string returns 0 (per DWScript fixtures: ord.pas line 4)
			return &IntegerValue{Value: 0}
		}
		// Return the Unicode code point of the first character
		runes := []rune(strVal.Value)
		return &IntegerValue{Value: int64(runes[0])}
	}

	return i.newErrorWithLocation(i.currentNode, "Ord() expects enum, boolean, integer, or string, got %s", arg.Type())
}

// builtinInteger implements the Integer() cast function.
// It converts values to integers.
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

	// Use rune-based slicing to handle UTF-8 correctly
	result := runeSliceFrom(str, int(index), int(count))
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

// builtinLow implements the Low() built-in function.
// It returns the lower bound of an array or the lowest value of an enum type.
func (i *Interpreter) builtinLow(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Low() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Handle type meta-values (type names as values)
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
		enumTypeKey := "__enum_type_" + strings.ToLower(enumVal.TypeName)
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
func (i *Interpreter) builtinHigh(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "High() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Handle type meta-values (type names as values)
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
		// Normalize to lowercase for case-insensitive lookups
		enumTypeKey := "__enum_type_" + strings.ToLower(enumVal.TypeName)
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

// builtinIntToStr implements the IntToStr() built-in function.
// It converts an integer to its string representation.
// IntToStr(i: Integer): String
// IntToStr(i: Integer, base: Integer): String (base 2-36)
func (i *Interpreter) builtinIntToStr(args []Value) Value {
	if len(args) < 1 || len(args) > 2 {
		return i.newErrorWithLocation(i.currentNode, "IntToStr() expects 1 or 2 arguments, got %d", len(args))
	}

	// First argument must be an integer or subrange value
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

	// Default base is 10
	base := 10

	// If second argument is provided, it specifies the base
	if len(args) == 2 {
		switch v := args[1].(type) {
		case *IntegerValue:
			base = int(v.Value)
		case *SubrangeValue:
			base = int(v.Value)
		default:
			return i.newErrorWithLocation(i.currentNode, "IntToStr() expects integer as second argument (base), got %s", args[1].Type())
		}

		// Validate base range (2-36)
		if base < 2 || base > 36 {
			return i.newErrorWithLocation(i.currentNode, "IntToStr() base must be between 2 and 36, got %d", base)
		}
	}

	// Convert integer to string using Go's strconv with specified base
	result := strconv.FormatInt(intValue, base)
	return &StringValue{Value: result}
}

// builtinIntToBin implements the IntToBin() built-in function.
// It converts an integer to its binary string representation with specified width.
// IntToBin(v: Integer, digits: Integer): String
func (i *Interpreter) builtinIntToBin(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "IntToBin() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: integer value to convert
	var value int64
	switch v := args[0].(type) {
	case *IntegerValue:
		value = v.Value
	case *SubrangeValue:
		value = int64(v.Value)
	default:
		return i.newErrorWithLocation(i.currentNode, "IntToBin() expects integer as first argument, got %s", args[0].Type())
	}

	// Second argument: minimum number of digits
	var digits int64
	switch d := args[1].(type) {
	case *IntegerValue:
		digits = d.Value
	case *SubrangeValue:
		digits = int64(d.Value)
	default:
		return i.newErrorWithLocation(i.currentNode, "IntToBin() expects integer as second argument, got %s", args[1].Type())
	}

	// Convert to binary string
	// For negative numbers, we use two's complement representation (64-bit)
	var result string
	if value < 0 {
		// Negative numbers: use full 64-bit two's complement
		// This matches DWScript behavior
		uValue := uint64(value)
		for i := 63; i >= 0; i-- {
			if (uValue & (1 << uint(i))) != 0 {
				result += "1"
			} else {
				result += "0"
			}
		}
	} else {
		// Positive numbers: convert using bit operations
		// Build string from least significant bit to most significant
		remaining := digits
		temp := value

		// Build the binary representation
		var bits []byte
		for temp != 0 || remaining > 0 {
			if (temp & 1) == 1 {
				bits = append(bits, '1')
			} else {
				bits = append(bits, '0')
			}
			temp >>= 1
			remaining--
		}

		// Reverse the bits to get correct order (most significant first)
		for i := len(bits) - 1; i >= 0; i-- {
			result += string(bits[i])
		}

		// Handle special case of zero with no bits
		if result == "" {
			result = "0"
		}
	}

	return &StringValue{Value: result}
}

// builtinStrToInt implements the StrToInt() built-in function.
// It converts a string to an integer, raising an error if the string is invalid.
// StrToInt(s: String): Integer
// StrToInt(s: String, base: Integer): Integer (base 2-36)
func (i *Interpreter) builtinStrToInt(args []Value) Value {
	if len(args) < 1 || len(args) > 2 {
		return i.newErrorWithLocation(i.currentNode, "StrToInt() expects 1 or 2 arguments, got %d", len(args))
	}

	// First argument must be a string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrToInt() expects string argument, got %s", args[0].Type())
	}

	// Default base is 10
	base := 10

	// If second argument is provided, it specifies the base
	if len(args) == 2 {
		switch v := args[1].(type) {
		case *IntegerValue:
			base = int(v.Value)
		case *SubrangeValue:
			base = int(v.Value)
		default:
			return i.newErrorWithLocation(i.currentNode, "StrToInt() expects integer as second argument (base), got %s", args[1].Type())
		}

		// Validate base range (2-36)
		if base < 2 || base > 36 {
			return i.newErrorWithLocation(i.currentNode, "StrToInt() base must be between 2 and 36, got %d", base)
		}
	}

	// Try to parse the string as an integer
	// Use strings.TrimSpace to handle leading/trailing whitespace
	s := strings.TrimSpace(strVal.Value)

	// Empty string is an error
	if s == "" {
		return i.newErrorWithLocation(i.currentNode, "empty string is not a valid integer")
	}

	// Use strconv.ParseInt for strict parsing (doesn't accept partial matches)
	intValue, err := strconv.ParseInt(s, base, 64)
	if err != nil {
		return i.newErrorWithLocation(i.currentNode, "'%s' is not a valid integer", strVal.Value)
	}

	return &IntegerValue{Value: intValue}
}

// builtinFloatToStr implements the FloatToStr() built-in function.
// It converts a float to its string representation.
// FloatToStr(f: Float): String
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

// builtinStrToIntDef implements the StrToIntDef() built-in function.
// It converts a string to an integer, returning a default value if the string is invalid.
// StrToIntDef(s: String, default: Integer): Integer
// StrToIntDef(s: String, default: Integer, base: Integer): Integer (base 2-36)
func (i *Interpreter) builtinStrToIntDef(args []Value) Value {
	if len(args) < 2 || len(args) > 3 {
		return i.newErrorWithLocation(i.currentNode, "StrToIntDef() expects 2 or 3 arguments, got %d", len(args))
	}

	// First argument must be a string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrToIntDef() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument must be an integer (the default value)
	defaultVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrToIntDef() expects integer as second argument, got %s", args[1].Type())
	}

	// Default base is 10
	base := 10

	// If third argument is provided, it specifies the base
	if len(args) == 3 {
		switch v := args[2].(type) {
		case *IntegerValue:
			base = int(v.Value)
		case *SubrangeValue:
			base = int(v.Value)
		default:
			return i.newErrorWithLocation(i.currentNode, "StrToIntDef() expects integer as third argument (base), got %s", args[2].Type())
		}

		// Validate base range (2-36)
		if base < 2 || base > 36 {
			return i.newErrorWithLocation(i.currentNode, "StrToIntDef() base must be between 2 and 36, got %d", base)
		}
	}

	// Try to parse the string as an integer
	s := strings.TrimSpace(strVal.Value)

	// Use strconv.ParseInt for strict parsing (doesn't accept partial matches)
	intValue, err := strconv.ParseInt(s, base, 64)
	if err != nil {
		// Return the default value on error
		return &IntegerValue{Value: defaultVal.Value}
	}

	return &IntegerValue{Value: intValue}
}

// builtinStrToFloatDef implements the StrToFloatDef() built-in function.
// It converts a string to a float, returning a default value if the string is invalid.
// StrToFloatDef(s: String, default: Float): Float
func (i *Interpreter) builtinStrToFloatDef(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "StrToFloatDef() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument must be a string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrToFloatDef() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument must be a float (the default value)
	defaultVal, ok := args[1].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrToFloatDef() expects float as second argument, got %s", args[1].Type())
	}

	// Try to parse the string as a float
	s := strings.TrimSpace(strVal.Value)

	// Use strconv.ParseFloat for strict parsing (doesn't accept partial matches)
	floatValue, err := strconv.ParseFloat(s, 64)
	if err != nil {
		// Return the default value on error
		return &FloatValue{Value: defaultVal.Value}
	}

	return &FloatValue{Value: floatValue}
}

// builtinBoolToStr implements the BoolToStr() built-in function.
// It converts a boolean to its string representation ("True" or "False").
// BoolToStr(b: Boolean): String
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

// builtinTypeOf implements the TypeOf() built-in function.
// Task 9.25.1: TypeOf(value): TTypeInfo
// Returns runtime type information for the given value.
// Can accept any value (object, class, primitive, etc.)
func (i *Interpreter) builtinTypeOf(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "TypeOf() expects exactly 1 argument, got %d", len(args))
	}

	val := args[0]

	// Get type ID and type name based on value type
	typeID, typeName := i.getTypeIDAndName(val)

	return &RTTITypeInfoValue{
		TypeID:   typeID,
		TypeName: typeName,
		TypeInfo: i.getValueType(val),
	}
}

// builtinTypeOfClass implements the TypeOfClass() built-in function.
// Task 9.25.2: TypeOfClass(classRef: TClass): TTypeInfo
// Returns type information for a class reference (metaclass).
func (i *Interpreter) builtinTypeOfClass(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "TypeOfClass() expects exactly 1 argument, got %d", len(args))
	}

	val := args[0]

	// TypeOfClass expects a class reference (ClassValue or ClassInfoValue)
	var classInfo *ClassInfo
	switch v := val.(type) {
	case *ClassValue:
		classInfo = v.ClassInfo
	case *ClassInfoValue:
		classInfo = v.ClassInfo
	default:
		return i.newErrorWithLocation(i.currentNode, "TypeOfClass() expects a class reference, got %s", val.Type())
	}

	if classInfo == nil {
		return i.newErrorWithLocation(i.currentNode, "TypeOfClass() received nil class reference")
	}

	// Generate type ID for the class
	typeID := i.getClassTypeID(classInfo.Name)

	return &RTTITypeInfoValue{
		TypeID:   typeID,
		TypeName: classInfo.Name,
		TypeInfo: nil, // Could be enhanced with class type metadata
	}
}

// getTypeIDAndName returns a unique type ID and human-readable type name for a value.
// Task 9.25: Helper for TypeOf() implementation.
func (i *Interpreter) getTypeIDAndName(val Value) (int, string) {
	switch v := val.(type) {
	case *IntegerValue:
		return 1, "Integer"
	case *FloatValue:
		return 2, "Float"
	case *StringValue:
		return 3, "String"
	case *BooleanValue:
		return 4, "Boolean"
	case *ObjectInstance:
		if v.Class != nil {
			return i.getClassTypeID(v.Class.Name), v.Class.Name
		}
		return 100, "TObject"
	case *ClassValue:
		if v.ClassInfo != nil {
			return i.getClassTypeID(v.ClassInfo.Name), v.ClassInfo.Name
		}
		return 100, "TObject"
	case *ClassInfoValue:
		if v.ClassInfo != nil {
			return i.getClassTypeID(v.ClassInfo.Name), v.ClassInfo.Name
		}
		return 100, "TObject"
	case *ArrayValue:
		return 10, "Array"
	case *RecordValue:
		if v.RecordType != nil {
			return i.getRecordTypeID(v.RecordType.Name), v.RecordType.Name
		}
		return 20, "Record"
	case *EnumValue:
		return i.getEnumTypeID(v.TypeName), v.TypeName
	case *SetValue:
		return 30, "Set"
	case *VariantValue:
		// For Variant, return the type of the contained value
		if v.Value != nil {
			return i.getTypeIDAndName(v.Value)
		}
		return 40, "Variant"
	case *NilValue:
		return 0, "Nil"
	case *TypeMetaValue:
		return 50, v.TypeName
	default:
		return 999, "Unknown"
	}
}

// getClassTypeID returns a unique type ID for a class name.
// Type IDs start at 1000 for classes.
// Uses a registry to ensure unique IDs and handles case-insensitivity.
func (i *Interpreter) getClassTypeID(className string) int {
	// Normalize to lowercase for case-insensitive comparison
	normalizedName := strings.ToLower(className)

	// Check if we already have an ID for this class
	if id, exists := i.classTypeIDRegistry[normalizedName]; exists {
		return id
	}

	// Assign new ID and store in registry
	id := i.nextClassTypeID
	i.classTypeIDRegistry[normalizedName] = id
	i.nextClassTypeID++
	return id
}

// getRecordTypeID returns a unique type ID for a record name.
// Type IDs start at 200000 for records.
// Uses a registry to ensure unique IDs and handles case-insensitivity.
func (i *Interpreter) getRecordTypeID(recordName string) int {
	// Normalize to lowercase for case-insensitive comparison
	normalizedName := strings.ToLower(recordName)

	// Check if we already have an ID for this record
	if id, exists := i.recordTypeIDRegistry[normalizedName]; exists {
		return id
	}

	// Assign new ID and store in registry
	id := i.nextRecordTypeID
	i.recordTypeIDRegistry[normalizedName] = id
	i.nextRecordTypeID++
	return id
}

// getEnumTypeID returns a unique type ID for an enum name.
// Type IDs start at 300000 for enums.
// Uses a registry to ensure unique IDs and handles case-insensitivity.
func (i *Interpreter) getEnumTypeID(enumName string) int {
	// Normalize to lowercase for case-insensitive comparison
	normalizedName := strings.ToLower(enumName)

	// Check if we already have an ID for this enum
	if id, exists := i.enumTypeIDRegistry[normalizedName]; exists {
		return id
	}

	// Assign new ID and store in registry
	id := i.nextEnumTypeID
	i.enumTypeIDRegistry[normalizedName] = id
	i.nextEnumTypeID++
	return id
}
