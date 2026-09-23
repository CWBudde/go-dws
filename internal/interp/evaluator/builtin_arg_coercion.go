package evaluator

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
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
		// A record reaching a Variant parameter goes through a user-defined
		// `operator implicit (TRec) : Variant` when one is registered, so
		// PrintLn(rec) prints what the operator returns rather than the
		// default record dump. Gated on records so the common argument kinds
		// never touch the conversion registry.
		if sig.ParamTypes[i].TypeKind() == "VARIANT" {
			if _, isRecord := unwrapVariant(args[i]).(*runtime.RecordValue); isRecord {
				if converted, ok := e.TryImplicitConversion(unwrapVariant(args[i]), types.VARIANT, ctx); ok {
					args[i] = converted
				}
			}
			continue
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
	if resolved := e.SemanticInfo().GetResolvedType(expr); resolved != nil {
		return types.GetUnderlyingType(resolved) == types.VARIANT
	}
	typeAnnot := e.SemanticInfo().GetType(expr)
	if typeAnnot == nil {
		return false
	}
	return ident.Equal(typeAnnot.Name, "Variant")
}

// coerceValueToKind converts a basic runtime value to the given type kind.
// Returns (nil, nil) when no replacement applies. A matching Boolean is still
// returned to replace a possible Variant wrapper. A failed cast raises a
// catchable exception and returns a non-nil error value.
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
	if _, ok := arg.(*runtime.IntegerValue); ok {
		return nil, nil
	}
	if n, ok := scalarToInt64(arg); ok {
		return &runtime.IntegerValue{Value: n}, nil
	}
	if _, ok := arg.(*runtime.StringValue); ok {
		return nil, e.raiseVariantCastException("Could not cast variant from String to Integer", funcName, ctx)
	}
	return nil, nil
}

func (e *Evaluator) coerceToFloat(arg Value, funcName *ast.Identifier, ctx *ExecutionContext) (Value, Value) {
	if _, ok := arg.(*runtime.FloatValue); ok {
		return nil, nil
	}
	// Builtins with a FLOAT/TDateTime signature type-assert *FloatValue,
	// so a Variant-held Integer must be widened at the call boundary.
	if f, ok := scalarToFloat64(arg); ok {
		return &runtime.FloatValue{Value: f}, nil
	}
	if v, ok := arg.(*runtime.StringValue); ok {
		return nil, e.raiseVariantCastException(
			fmt.Sprintf("%q is not a valid floating point value", v.Value), funcName, ctx)
	}
	return nil, nil
}

// scalarToInt64 converts a Variant's scalar content to Integer the way
// DWScript's VariantToInt64 does: Floats round half to even (Delphi's Round),
// Booleans are 0/1 and Strings must hold a decimal integer.
func scalarToInt64(v Value) (int64, bool) {
	switch v := v.(type) {
	case *runtime.IntegerValue:
		return v.Value, true
	case *runtime.FloatValue:
		return int64(math.RoundToEven(v.Value)), true
	case *runtime.BooleanValue:
		if v.Value {
			return 1, true
		}
		return 0, true
	case *runtime.StringValue:
		n, err := strconv.ParseInt(strings.TrimSpace(v.Value), 10, 64)
		return n, err == nil
	}
	return 0, false
}

// scalarToFloat64 converts a Variant's scalar content to Float.
func scalarToFloat64(v Value) (float64, bool) {
	switch v := v.(type) {
	case *runtime.FloatValue:
		return v.Value, true
	case *runtime.IntegerValue:
		return float64(v.Value), true
	case *runtime.BooleanValue:
		if v.Value {
			return 1, true
		}
		return 0, true
	case *runtime.StringValue:
		f, err := strconv.ParseFloat(strings.TrimSpace(v.Value), 64)
		return f, err == nil
	}
	return 0, false
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
		// Publish the unwrapped value even when its type already matches:
		// the original argument can still be a Variant wrapper.
		return v, nil
	case *runtime.UnassignedValue, *runtime.NullValue, *runtime.NilValue:
		return &runtime.BooleanValue{Value: false}, nil
	case *runtime.IntegerValue:
		return &runtime.BooleanValue{Value: v.Value != 0}, nil
	case *runtime.EnumValue:
		return &runtime.BooleanValue{Value: v.OrdinalValue != 0}, nil
	case *runtime.SubrangeValue:
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
