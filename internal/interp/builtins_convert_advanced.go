package interp

import (
	"fmt"
	"strconv"
	"strings"
)

// builtinHexToInt implements the HexToInt() built-in function.
// It converts a hexadecimal string to an integer.
// HexToInt(hexa: String): Integer
func (i *Interpreter) builtinHexToInt(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "HexToInt() expects exactly 1 argument, got %d", len(args))
	}

	// Argument must be a string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "HexToInt() expects string argument, got %s", args[0].Type())
	}

	// Clean the hex string - remove $ or 0x prefix if present
	s := strings.TrimSpace(strVal.Value)
	s = strings.TrimPrefix(s, "$")
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")

	// Parse as hexadecimal (base 16)
	intValue, err := strconv.ParseInt(s, 16, 64)
	if err != nil {
		// Raise an exception that can be caught by try/except
		msg := fmt.Sprintf("'%s' is not a valid hexadecimal number", strVal.Value)
		if i.currentNode != nil {
			pos := i.currentNode.Pos()
			i.raiseException("EConvertError", msg, &pos)
		} else {
			i.raiseException("EConvertError", msg, nil)
		}
		return &IntegerValue{Value: 0}
	}

	return &IntegerValue{Value: intValue}
}

// builtinBinToInt implements the BinToInt() built-in function.
// It converts a binary string to an integer.
// BinToInt(binary: String): Integer
func (i *Interpreter) builtinBinToInt(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "BinToInt() expects exactly 1 argument, got %d", len(args))
	}

	// Argument must be a string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "BinToInt() expects string argument, got %s", args[0].Type())
	}

	// Clean the binary string - remove % or 0b prefix if present
	s := strings.TrimSpace(strVal.Value)
	s = strings.TrimPrefix(s, "%")
	s = strings.TrimPrefix(s, "0b")
	s = strings.TrimPrefix(s, "0B")

	// Parse as binary (base 2)
	intValue, err := strconv.ParseInt(s, 2, 64)
	if err != nil {
		return i.newErrorWithLocation(i.currentNode, "'%s' is not a valid binary number", strVal.Value)
	}

	return &IntegerValue{Value: intValue}
}

// builtinVarToIntDef implements the VarToIntDef() built-in function.
// It converts a variant to an integer, returning a default value if conversion fails.
// VarToIntDef(v: Variant, default: Integer): Integer
func (i *Interpreter) builtinVarToIntDef(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "VarToIntDef() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: value to convert (can be any type)
	varVal := args[0]

	// Second argument must be an integer (the default value)
	defaultVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "VarToIntDef() expects integer as second argument, got %s", args[1].Type())
	}

	// Try to convert the value to an integer
	switch v := varVal.(type) {
	case *IntegerValue:
		return &IntegerValue{Value: v.Value}
	case *SubrangeValue:
		return &IntegerValue{Value: int64(v.Value)}
	case *FloatValue:
		return &IntegerValue{Value: int64(v.Value)}
	case *BooleanValue:
		if v.Value {
			return &IntegerValue{Value: 1}
		}
		return &IntegerValue{Value: 0}
	case *StringValue:
		// Try to parse as integer
		s := strings.TrimSpace(v.Value)
		intValue, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return &IntegerValue{Value: defaultVal.Value}
		}
		return &IntegerValue{Value: intValue}
	case *VariantValue:
		// Recursively convert the underlying value
		return i.builtinVarToIntDef([]Value{v.Value, defaultVal})
	default:
		// Cannot convert - return default
		return &IntegerValue{Value: defaultVal.Value}
	}
}

// builtinVarToFloatDef implements the VarToFloatDef() built-in function.
// It converts a variant to a float, returning a default value if conversion fails.
// VarToFloatDef(v: Variant, default: Float): Float
func (i *Interpreter) builtinVarToFloatDef(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "VarToFloatDef() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: value to convert (can be any type)
	varVal := args[0]

	// Second argument must be a float (the default value)
	defaultVal, ok := args[1].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "VarToFloatDef() expects float as second argument, got %s", args[1].Type())
	}

	// Try to convert the value to a float
	switch v := varVal.(type) {
	case *FloatValue:
		return &FloatValue{Value: v.Value}
	case *IntegerValue:
		return &FloatValue{Value: float64(v.Value)}
	case *SubrangeValue:
		return &FloatValue{Value: float64(v.Value)}
	case *BooleanValue:
		if v.Value {
			return &FloatValue{Value: 1.0}
		}
		return &FloatValue{Value: 0.0}
	case *NullValue:
		// Delphi/DWScript semantics: Null → 0 for VarToFloatDef
		return &FloatValue{Value: 0.0}
	case *NilValue:
		// Treat nil object references like Null for float conversion
		return &FloatValue{Value: 0.0}
	case *UnassignedValue:
		// Unassigned keeps the provided default
		return &FloatValue{Value: defaultVal.Value}
	case *StringValue:
		// Try to parse as float
		if floatValue, ok := parseFloatWithFallback(v.Value); ok {
			return &FloatValue{Value: floatValue}
		}
		return &FloatValue{Value: defaultVal.Value}
	case *VariantValue:
		// Recursively convert the underlying value
		if v.Value == nil {
			return &FloatValue{Value: defaultVal.Value}
		}
		// Null should map to 0, not default
		if _, isNull := v.Value.(*NullValue); isNull {
			return &FloatValue{Value: 0.0}
		}
		if _, isUnassigned := v.Value.(*UnassignedValue); isUnassigned {
			return &FloatValue{Value: defaultVal.Value}
		}
		return i.builtinVarToFloatDef([]Value{v.Value, defaultVal})
	default:
		// Cannot convert - return default
		return &FloatValue{Value: defaultVal.Value}
	}
}

// parseFloatWithFallback parses a float string, accepting both '.' and ',' as decimal separators.
func parseFloatWithFallback(s string) (float64, bool) {
	s = strings.TrimSpace(s)

	if floatValue, err := strconv.ParseFloat(s, 64); err == nil {
		return floatValue, true
	}

	// Fallback: accept comma as decimal separator when dot is absent
	if strings.Contains(s, ",") && !strings.Contains(s, ".") {
		if floatValue, err := strconv.ParseFloat(strings.ReplaceAll(s, ",", "."), 64); err == nil {
			return floatValue, true
		}
	}

	return 0.0, false
}
