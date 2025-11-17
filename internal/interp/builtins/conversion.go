package builtins

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// ============================================================================
// Type Conversion Built-in Functions (Task 3.7.3)
// ============================================================================
//
// This file contains basic type conversion functions that have been migrated
// from internal/interp to use the Context interface pattern.
//
// Functions in this file:
//   - IntToStr: Convert integer to string (with optional base)
//   - IntToBin: Convert integer to binary string
//   - StrToInt: Convert string to integer
//   - StrToFloat: Convert string to float
//   - FloatToStr: Convert float to string
//   - BoolToStr: Convert boolean to string
//
// These functions use the Context helper methods (ToInt64, ToFloat64, etc.)
// to handle type coercion from SubrangeValue, EnumValue, and other types.

// IntToStr converts an integer to its string representation.
// IntToStr(i: Integer): String
// IntToStr(i: Integer, base: Integer): String (base 2-36)
func IntToStr(ctx Context, args []Value) Value {
	if len(args) < 1 || len(args) > 2 {
		return ctx.NewError("IntToStr() expects 1 or 2 arguments, got %d", len(args))
	}

	// First argument must be an integer (or subrange/enum)
	intValue, ok := ctx.ToInt64(args[0])
	if !ok {
		return ctx.NewError("IntToStr() expects integer argument, got %s", args[0].Type())
	}

	// Default base is 10
	base := 10

	// If second argument is provided, it specifies the base
	if len(args) == 2 {
		baseValue, ok := ctx.ToInt64(args[1])
		if !ok {
			return ctx.NewError("IntToStr() expects integer as second argument (base), got %s", args[1].Type())
		}
		base = int(baseValue)

		// Validate base range (2-36)
		if base < 2 || base > 36 {
			return ctx.NewError("IntToStr() base must be between 2 and 36, got %d", base)
		}
	}

	// Convert integer to string using Go's strconv with specified base
	result := strconv.FormatInt(intValue, base)
	return &runtime.StringValue{Value: result}
}

// IntToBin converts an integer to its binary string representation with specified width.
// IntToBin(v: Integer, digits: Integer): String
func IntToBin(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("IntToBin() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: integer value to convert
	value, ok := ctx.ToInt64(args[0])
	if !ok {
		return ctx.NewError("IntToBin() expects integer as first argument, got %s", args[0].Type())
	}

	// Second argument: minimum number of digits
	digits, ok := ctx.ToInt64(args[1])
	if !ok {
		return ctx.NewError("IntToBin() expects integer as second argument, got %s", args[1].Type())
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

	return &runtime.StringValue{Value: result}
}

// StrToInt converts a string to an integer, raising an error if the string is invalid.
// StrToInt(s: String): Integer
// StrToInt(s: String, base: Integer): Integer (base 2-36)
func StrToInt(ctx Context, args []Value) Value {
	if len(args) < 1 || len(args) > 2 {
		return ctx.NewError("StrToInt() expects 1 or 2 arguments, got %d", len(args))
	}

	// First argument must be a string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrToInt() expects string argument, got %s", args[0].Type())
	}

	// Default base is 10
	base := 10

	// If second argument is provided, it specifies the base
	if len(args) == 2 {
		baseValue, ok := ctx.ToInt64(args[1])
		if !ok {
			return ctx.NewError("StrToInt() expects integer as second argument (base), got %s", args[1].Type())
		}
		base = int(baseValue)

		// Validate base range (2-36)
		if base < 2 || base > 36 {
			return ctx.NewError("StrToInt() base must be between 2 and 36, got %d", base)
		}
	}

	// Try to parse the string as an integer
	// Use strings.TrimSpace to handle leading/trailing whitespace
	s := strings.TrimSpace(strVal.Value)

	// Special handling for hex prefix (0x or $)
	if base == 16 {
		s = strings.TrimPrefix(s, "0x")
		s = strings.TrimPrefix(s, "0X")
		s = strings.TrimPrefix(s, "$")
	}

	// Use strconv.ParseInt for strict parsing
	intValue, err := strconv.ParseInt(s, base, 64)
	if err != nil {
		return ctx.NewError("'%s' is not a valid integer", strVal.Value)
	}

	return &runtime.IntegerValue{Value: intValue}
}

// StrToFloat converts a string to a float, raising an error if the string is invalid.
// StrToFloat(s: String): Float
func StrToFloat(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("StrToFloat() expects exactly 1 argument, got %d", len(args))
	}

	// Argument must be a string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrToFloat() expects string argument, got %s", args[0].Type())
	}

	// Try to parse the string as a float
	s := strings.TrimSpace(strVal.Value)

	// Use strconv.ParseFloat for strict parsing (doesn't accept partial matches)
	floatValue, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return ctx.NewError("'%s' is not a valid float", strVal.Value)
	}

	return &runtime.FloatValue{Value: floatValue}
}

// FloatToStr converts a float to its string representation.
// FloatToStr(f: Float): String
func FloatToStr(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("FloatToStr() expects exactly 1 argument, got %d", len(args))
	}

	// Argument can be Float or Integer (implicit coercion)
	floatValue, ok := ctx.ToFloat64(args[0])
	if !ok {
		return ctx.NewError("FloatToStr() expects float argument, got %s", args[0].Type())
	}

	// Convert float to string using Go's fmt
	// Use 'g' format for general representation (like DWScript's FloatToStr)
	result := fmt.Sprintf("%g", floatValue)
	return &runtime.StringValue{Value: result}
}

// BoolToStr converts a boolean to its string representation.
// BoolToStr(b: Boolean): String
func BoolToStr(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("BoolToStr() expects exactly 1 argument, got %d", len(args))
	}

	// Argument must be a boolean
	boolValue, ok := ctx.ToBool(args[0])
	if !ok {
		return ctx.NewError("BoolToStr() expects boolean argument, got %s", args[0].Type())
	}

	// Convert boolean to string
	// DWScript uses "True" and "False" (capitalized)
	if boolValue {
		return &runtime.StringValue{Value: "True"}
	}
	return &runtime.StringValue{Value: "False"}
}
