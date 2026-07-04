package evaluator

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// coerceBuiltinArgsToSignature converts arguments whose dynamic type differs
// from a builtin's declared parameter type. Variant-typed expressions pass the
// semantic analyzer but reach the builtin with their dynamic type (Integer,
// Float, String, Boolean); DWScript applies variant casts at the call
// boundary. Failed casts raise a catchable exception on ctx and return a nil
// placeholder value; the caller must check ctx.Exception().
func (e *Evaluator) coerceBuiltinArgsToSignature(funcName *ast.Identifier, args []Value, ctx *ExecutionContext) Value {
	sig, ok := builtins.DefaultRegistry.GetSignature(funcName.Value)
	if !ok || sig == nil {
		return nil
	}

	for i := range args {
		if i >= len(sig.ParamTypes) || sig.ParamTypes[i] == nil {
			break
		}
		paramKind := sig.ParamTypes[i].TypeKind()
		arg := unwrapVariant(args[i])
		if arg == nil {
			continue
		}

		converted, errVal := e.coerceValueToKind(arg, paramKind, funcName, ctx)
		if errVal != nil {
			return errVal
		}
		if converted != nil {
			args[i] = converted
		}
	}
	return nil
}

// coerceValueToKind converts a basic runtime value to the given type kind.
// Returns (nil, nil) when no conversion applies (value already matches or the
// kind is not a basic type). A failed cast raises a catchable exception and
// returns a non-nil error value.
func (e *Evaluator) coerceValueToKind(arg Value, kind string, funcName *ast.Identifier, ctx *ExecutionContext) (Value, Value) {
	switch kind {
	case "INTEGER":
		switch v := arg.(type) {
		case *runtime.IntegerValue:
			return nil, nil
		case *runtime.FloatValue:
			return &runtime.IntegerValue{Value: int64(math.Round(v.Value))}, nil
		case *runtime.BooleanValue:
			if v.Value {
				return &runtime.IntegerValue{Value: 1}, nil
			}
			return &runtime.IntegerValue{Value: 0}, nil
		case *runtime.StringValue:
			if n, err := strconv.ParseInt(strings.TrimSpace(v.Value), 10, 64); err == nil {
				return &runtime.IntegerValue{Value: n}, nil
			}
			return nil, e.raiseVariantCastException("Could not cast variant from String to Integer", funcName, ctx)
		}
	case "FLOAT":
		switch v := arg.(type) {
		case *runtime.FloatValue, *runtime.IntegerValue:
			return nil, nil
		case *runtime.BooleanValue:
			if v.Value {
				return &runtime.FloatValue{Value: 1}, nil
			}
			return &runtime.FloatValue{Value: 0}, nil
		case *runtime.StringValue:
			if f, err := strconv.ParseFloat(strings.TrimSpace(v.Value), 64); err == nil {
				return &runtime.FloatValue{Value: f}, nil
			}
			return nil, e.raiseVariantCastException(
				fmt.Sprintf("%q is not a valid floating point value", v.Value), funcName, ctx)
		}
	case "STRING":
		switch arg.(type) {
		case *runtime.StringValue:
			return nil, nil
		case *runtime.IntegerValue, *runtime.FloatValue, *runtime.BooleanValue:
			return &runtime.StringValue{Value: convertToString(arg)}, nil
		}
	case "BOOLEAN":
		switch v := arg.(type) {
		case *runtime.BooleanValue:
			return nil, nil
		case *runtime.IntegerValue:
			return &runtime.BooleanValue{Value: v.Value != 0}, nil
		case *runtime.FloatValue:
			return &runtime.BooleanValue{Value: v.Value != 0}, nil
		case *runtime.StringValue:
			return &runtime.BooleanValue{Value: stringToBoolCast(v.Value)}, nil
		}
	}
	return nil, nil
}

// stringToBoolCast applies DWScript's String→Boolean variant cast: recognized
// "true"-ish spellings and non-zero numbers are True, anything else is False.
func stringToBoolCast(s string) bool {
	s = strings.TrimSpace(s)
	switch {
	case ident.Equal(s, "true") || ident.Equal(s, "t") || ident.Equal(s, "y") || ident.Equal(s, "yes"):
		return true
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f != 0
	}
	return false
}

// raiseVariantCastException raises a catchable exception for a failed variant
// cast at the builtin call site, matching DWScript's message format
// ("<msg> in <routine> [line: N, column: M]").
func (e *Evaluator) raiseVariantCastException(message string, funcName *ast.Identifier, ctx *ExecutionContext) Value {
	pos := funcName.Token.Pos
	if routine := currentRoutineName(ctx); routine != "" {
		message += " in " + routine
	}
	message = fmt.Sprintf("%s [line: %d, column: %d]", message, pos.Line, pos.Column)
	ctx.SetException(e.createException("Exception", message, &pos, ctx))
	return e.nilValue()
}
