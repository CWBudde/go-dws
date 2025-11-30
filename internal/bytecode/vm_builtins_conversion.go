package bytecode

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// registerConversionBuiltins registers all type conversion functions
func (vm *VM) registerConversionBuiltins() {
	vm.builtins["IntToStr"] = builtinIntToStr
	vm.builtins["FloatToStr"] = builtinFloatToStr
	vm.builtins["StrToInt"] = builtinStrToInt
	vm.builtins["StrToFloat"] = builtinStrToFloat
	vm.builtins["StrToIntDef"] = builtinStrToIntDef
	vm.builtins["StrToFloatDef"] = builtinStrToFloatDef
	vm.builtins["Integer"] = builtinInteger
	vm.builtins["Float"] = builtinFloat
	vm.builtins["String"] = builtinString
	vm.builtins["Boolean"] = builtinBoolean
}

// String to/from conversion functions

func builtinIntToStr(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("IntToStr expects 1 argument, got %d", len(args))
	}
	if !args[0].IsInt() {
		return NilValue(), vm.runtimeError("IntToStr expects an integer argument")
	}
	return StringValue(fmt.Sprintf("%d", args[0].AsInt())), nil
}

func builtinFloatToStr(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("FloatToStr expects 1 argument, got %d", len(args))
	}
	// Accept both float and integer types (auto-convert integer to float)
	if !args[0].IsFloat() && !args[0].IsInt() {
		return NilValue(), vm.runtimeError("FloatToStr expects a numeric argument")
	}
	// AsFloat() handles conversion for both types
	return StringValue(fmt.Sprintf("%g", args[0].AsFloat())), nil
}

func builtinStrToInt(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("StrToInt expects 1 argument, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrToInt expects a string argument")
	}
	var val int64
	_, err := fmt.Sscanf(args[0].AsString(), "%d", &val)
	if err != nil {
		return NilValue(), vm.runtimeError("StrToInt: invalid integer string")
	}
	return IntValue(val), nil
}

func builtinStrToFloat(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("StrToFloat expects 1 argument, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrToFloat expects a string argument")
	}
	var val float64
	_, err := fmt.Sscanf(args[0].AsString(), "%f", &val)
	if err != nil {
		return NilValue(), vm.runtimeError("StrToFloat: invalid float string")
	}
	return FloatValue(val), nil
}

func builtinStrToIntDef(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("StrToIntDef expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrToIntDef expects a string as first argument")
	}
	if !args[1].IsInt() {
		return NilValue(), vm.runtimeError("StrToIntDef expects an integer as second argument")
	}
	// Try to parse the string as an integer
	s := strings.TrimSpace(args[0].AsString())
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		// Return default value on error
		return args[1], nil
	}
	return IntValue(val), nil
}

func builtinStrToFloatDef(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("StrToFloatDef expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrToFloatDef expects a string as first argument")
	}
	if !args[1].IsFloat() && !args[1].IsInt() {
		return NilValue(), vm.runtimeError("StrToFloatDef expects a float as second argument")
	}
	// Try to parse the string as a float
	s := strings.TrimSpace(args[0].AsString())
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		// Return default value on error (coerce int to float if needed)
		if args[1].IsInt() {
			return FloatValue(float64(args[1].AsInt())), nil
		}
		return args[1], nil
	}
	return FloatValue(val), nil
}

// Type cast built-in functions

func builtinInteger(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("Integer expects 1 argument, got %d", len(args))
	}
	arg := args[0]

	switch arg.Type {
	case ValueInt:
		return arg, nil
	case ValueFloat:
		return IntValue(int64(math.Round(arg.AsFloat()))), nil
	case ValueBool:
		if arg.AsBool() {
			return IntValue(1), nil
		}
		return IntValue(0), nil
	case ValueString:
		var val int64
		_, err := fmt.Sscanf(arg.AsString(), "%d", &val)
		if err != nil {
			return NilValue(), vm.runtimeError("cannot convert string '%s' to Integer", arg.AsString())
		}
		return IntValue(val), nil
	default:
		return NilValue(), vm.runtimeError("cannot cast %s to Integer", arg.Type.String())
	}
}

func builtinFloat(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("Float expects 1 argument, got %d", len(args))
	}
	arg := args[0]

	switch arg.Type {
	case ValueFloat:
		return arg, nil
	case ValueInt:
		return FloatValue(float64(arg.AsInt())), nil
	case ValueBool:
		if arg.AsBool() {
			return FloatValue(1.0), nil
		}
		return FloatValue(0.0), nil
	case ValueString:
		var val float64
		_, err := fmt.Sscanf(arg.AsString(), "%f", &val)
		if err != nil {
			return NilValue(), vm.runtimeError("cannot convert string '%s' to Float", arg.AsString())
		}
		return FloatValue(val), nil
	default:
		return NilValue(), vm.runtimeError("cannot cast %s to Float", arg.Type.String())
	}
}

func builtinString(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("String expects 1 argument, got %d", len(args))
	}
	arg := args[0]

	switch arg.Type {
	case ValueString:
		return arg, nil
	case ValueInt:
		return StringValue(fmt.Sprintf("%d", arg.AsInt())), nil
	case ValueFloat:
		return StringValue(fmt.Sprintf("%g", arg.AsFloat())), nil
	case ValueBool:
		if arg.AsBool() {
			return StringValue("True"), nil
		}
		return StringValue("False"), nil
	default:
		return NilValue(), vm.runtimeError("cannot cast %s to String", arg.Type.String())
	}
}

func builtinBoolean(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("Boolean expects 1 argument, got %d", len(args))
	}
	arg := args[0]

	switch arg.Type {
	case ValueBool:
		return arg, nil
	case ValueInt:
		return BoolValue(arg.AsInt() != 0), nil
	case ValueFloat:
		return BoolValue(arg.AsFloat() != 0.0), nil
	case ValueString:
		s := strings.ToLower(strings.TrimSpace(arg.AsString()))
		if s == "true" || s == "1" {
			return BoolValue(true), nil
		}
		if s == "false" || s == "0" || s == "" {
			return BoolValue(false), nil
		}
		return NilValue(), vm.runtimeError("cannot convert string '%s' to Boolean", arg.AsString())
	default:
		return NilValue(), vm.runtimeError("cannot cast %s to Boolean", arg.Type.String())
	}
}
