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
	"github.com/cwbudde/go-dws/pkg/token"
)

// coerceBuiltinArgsToSignature converts arguments whose dynamic type differs
// from a builtin's declared parameter type. Variant-typed expressions pass the
// semantic analyzer but reach the builtin with their dynamic type (Integer,
// Float, String, Boolean); DWScript applies variant casts at the call
// boundary. Failed casts raise a catchable exception on ctx and return a nil
// placeholder value; the caller must check ctx.Exception().
func (e *Evaluator) coerceBuiltinArgsToSignature(funcName *ast.Identifier, argExprs []ast.Expression, args []Value, ctx *ExecutionContext) Value {
	sig, ok := builtins.DefaultRegistry.GetSignature(funcName.Value)
	if !ok || sig == nil {
		return nil
	}

	for i := range args {
		if i >= len(sig.ParamTypes) || sig.ParamTypes[i] == nil {
			break
		}
		// Only apply variant casts to arguments whose static (declared) type
		// is Variant; other mismatches keep their strict runtime errors.
		if i >= len(argExprs) || !e.exprIsStaticVariant(argExprs[i]) {
			continue
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

// exprIsStaticVariant reports whether the semantic analyzer annotated the
// expression's static type as Variant.
func (e *Evaluator) exprIsStaticVariant(expr ast.Expression) bool {
	if e.SemanticInfo() == nil || expr == nil {
		return false
	}
	typeAnnot := e.SemanticInfo().GetType(expr)
	if typeAnnot == nil {
		return false
	}
	return ident.Equal(typeAnnot.Name, "Variant")
}

// coerceValueToKind converts a basic runtime value to the given type kind.
// Returns (nil, nil) when no conversion applies (value already matches or the
// kind is not a basic type). A failed cast raises a catchable exception and
// returns a non-nil error value.
func (e *Evaluator) coerceValueToKind(arg Value, kind string, funcName *ast.Identifier, ctx *ExecutionContext) (Value, Value) {
	switch kind {
	case "INTEGER":
		return e.coerceToInteger(arg, funcName, ctx)
	case "FLOAT":
		return e.coerceToFloat(arg, funcName, ctx)
	case "STRING":
		return coerceToString(arg)
	case "BOOLEAN":
		return coerceToBoolean(arg)
	}
	return nil, nil
}

func (e *Evaluator) coerceToInteger(arg Value, funcName *ast.Identifier, ctx *ExecutionContext) (Value, Value) {
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
	return nil, nil
}

func (e *Evaluator) coerceToFloat(arg Value, funcName *ast.Identifier, ctx *ExecutionContext) (Value, Value) {
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
	return nil, nil
}

func coerceToString(arg Value) (Value, Value) {
	switch arg.(type) {
	case *runtime.StringValue:
		return nil, nil
	case *runtime.IntegerValue, *runtime.FloatValue, *runtime.BooleanValue:
		return &runtime.StringValue{Value: convertToString(arg)}, nil
	}
	return nil, nil
}

func coerceToBoolean(arg Value) (Value, Value) {
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
	return nil, nil
}

// stringToBoolCast applies DWScript's String→Boolean variant cast: recognized
// "true"-ish spellings and non-zero numbers are True, anything else is False.
func stringToBoolCast(s string) bool {
	s = strings.TrimSpace(s)
	if ident.Equal(s, "true") || ident.Equal(s, "t") || ident.Equal(s, "y") || ident.Equal(s, "yes") {
		return true
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f != 0
	}
	return false
}

// ExtractIndexWithVariantCast extracts an integer index, applying DWScript
// variant casts: Boolean→0/1 and Float/Integer/Enum ordinals are accepted via
// ExtractIntegerIndex; a numeric String converts to its value; a non-numeric
// String raises a catchable "Could not cast variant from String to Integer"
// exception. Returns ok=false (with the exception set on ctx) on failure.
func (e *Evaluator) ExtractIndexWithVariantCast(indexVal Value, ctx *ExecutionContext) (int, bool) {
	indexVal = unwrapVariant(indexVal)
	if idx, ok := ExtractIntegerIndex(indexVal); ok {
		return idx, true
	}
	switch v := indexVal.(type) {
	case *runtime.FloatValue:
		return int(math.Round(v.Value)), true
	case *runtime.StringValue:
		if n, err := strconv.ParseInt(strings.TrimSpace(v.Value), 10, 64); err == nil {
			return int(n), true
		}
		if ctx == nil {
			ctx = e.currentContext
		}
		if ctx != nil {
			ctx.SetException(e.createException("Exception",
				"Could not cast variant from String to Integer", nil, ctx))
		}
		return 0, false
	}
	return 0, false
}

// raiseVariantCastException raises a catchable exception for a failed variant
// cast, matching DWScript's message format. When a call-site identifier is
// available the routine name and source location are appended
// ("<msg> in <routine> [line: N, column: M]"); otherwise the plain message
// is used.
func (e *Evaluator) raiseVariantCastException(message string, funcName *ast.Identifier, ctx *ExecutionContext) Value {
	if ctx == nil {
		ctx = e.currentContext
	}
	var posPtr *token.Position
	if funcName != nil {
		pos := funcName.Token.Pos
		if routine := currentRoutineName(ctx); routine != "" {
			message += " in " + routine
		}
		message = fmt.Sprintf("%s [line: %d, column: %d]", message, pos.Line, pos.Column)
		posPtr = &pos
	}
	if ctx != nil {
		ctx.SetException(e.createException("Exception", message, posPtr, ctx))
	}
	return e.nilValue()
}
