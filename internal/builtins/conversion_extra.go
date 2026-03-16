package builtins

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// HexToInt converts a hexadecimal string to an integer.
// HexToInt(hexa: String): Integer
func HexToInt(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("HexToInt() expects exactly 1 argument, got %d", len(args))
	}

	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("HexToInt() expects String as argument, got %s", args[0].Type())
	}

	s := strings.TrimSpace(strVal.Value)
	s = strings.TrimPrefix(s, "$")
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")

	intValue, err := strconv.ParseInt(s, 16, 64)
	if err != nil {
		msg := fmt.Sprintf("%q is not a valid hexadecimal number", strVal.Value)
		if node := ctx.CurrentNode(); node != nil {
			if posNode, ok := node.(interface{ Pos() lexer.Position }); ok {
				pos := posNode.Pos()
				if raiser, ok := ctx.(interface {
					RaiseException(className, message string, pos any)
				}); ok {
					raiser.RaiseException("EConvertError", msg, pos)
				}
			}
		}
		return ctx.NewError(msg)
	}

	return &runtime.IntegerValue{Value: intValue}
}

// BinToInt converts a binary string to an integer.
// BinToInt(binary: String): Integer
func BinToInt(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("BinToInt() expects exactly 1 argument, got %d", len(args))
	}

	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("BinToInt() expects String as argument, got %s", args[0].Type())
	}

	s := strings.TrimSpace(strVal.Value)
	s = strings.TrimPrefix(s, "%")
	s = strings.TrimPrefix(s, "0b")
	s = strings.TrimPrefix(s, "0B")

	intValue, err := strconv.ParseInt(s, 2, 64)
	if err != nil {
		return ctx.NewError("%q is not a valid binary number", strVal.Value)
	}

	return &runtime.IntegerValue{Value: intValue}
}

// VarToIntDef converts a variant-like value to integer, returning a default on failure.
// VarToIntDef(v: Variant, default: Integer): Integer
func VarToIntDef(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("VarToIntDef() expects exactly 2 arguments, got %d", len(args))
	}

	defaultVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("VarToIntDef() expects Integer as second argument, got %s", args[1].Type())
	}

	arg := ctx.UnwrapVariant(args[0])
	switch v := arg.(type) {
	case *runtime.IntegerValue:
		return &runtime.IntegerValue{Value: v.Value}
	case *runtime.SubrangeValue:
		return &runtime.IntegerValue{Value: int64(v.Value)}
	case *runtime.FloatValue:
		return &runtime.IntegerValue{Value: int64(v.Value)}
	case *runtime.BooleanValue:
		if v.Value {
			return &runtime.IntegerValue{Value: 1}
		}
		return &runtime.IntegerValue{Value: 0}
	case *runtime.StringValue:
		if intValue, ok := ctx.ParseInt(v.Value, 10); ok {
			return &runtime.IntegerValue{Value: intValue}
		}
	}

	return &runtime.IntegerValue{Value: defaultVal.Value}
}

// VarToFloatDef converts a variant-like value to float, returning a default on failure.
// VarToFloatDef(v: Variant, default: Float): Float
func VarToFloatDef(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("VarToFloatDef() expects exactly 2 arguments, got %d", len(args))
	}

	defaultFloat, ok := ctx.ToFloat64(args[1])
	if !ok {
		return ctx.NewError("VarToFloatDef() expects Float as second argument, got %s", args[1].Type())
	}

	arg := ctx.UnwrapVariant(args[0])
	switch v := arg.(type) {
	case *runtime.FloatValue:
		return &runtime.FloatValue{Value: v.Value}
	case *runtime.IntegerValue:
		return &runtime.FloatValue{Value: float64(v.Value)}
	case *runtime.SubrangeValue:
		return &runtime.FloatValue{Value: float64(v.Value)}
	case *runtime.BooleanValue:
		if v.Value {
			return &runtime.FloatValue{Value: 1}
		}
		return &runtime.FloatValue{Value: 0}
	case *runtime.StringValue:
		if floatValue, ok := ctx.ParseFloat(v.Value); ok {
			return &runtime.FloatValue{Value: floatValue}
		}
	}

	return &runtime.FloatValue{Value: defaultFloat}
}
