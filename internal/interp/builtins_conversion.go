package interp

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Type Conversion Built-in Functions
// Ord, Chr, Integer, IntToStr, StrToInt, FloatToStr, StrToFloat, etc.

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

	// Handle float values (round to nearest integer)
	if floatVal, ok := arg.(*FloatValue); ok {
		return &IntegerValue{Value: int64(math.Round(floatVal.Value))}
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
