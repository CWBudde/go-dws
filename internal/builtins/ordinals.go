package builtins

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
)

// =============================================================================
// Helper functions for enum operations with Context interface
// =============================================================================

// getEnumTypeForContext retrieves enum type metadata using Context interface.
// This is used by Succ/Pred which use the regular Context (not VarParamContext).
func getEnumTypeForContext(ctx Context, val *runtime.EnumValue) (*types.EnumType, Value) {
	enumMetadata := ctx.GetEnumMetadata(val.TypeName)
	if enumMetadata == nil {
		return nil, ctx.NewError("enum type metadata not found for %s", val.TypeName)
	}

	etv, ok := enumMetadata.(EnumTypeValueAccessor)
	if !ok {
		return nil, ctx.NewError("invalid enum type metadata for %s", val.TypeName)
	}
	return etv.GetEnumType(), nil
}

// findEnumPositionForContext finds the position of an enum value using Context interface.
func findEnumPositionForContext(ctx Context, enumType *types.EnumType, val *runtime.EnumValue) (int, Value) {
	for idx, name := range enumType.OrderedNames {
		if name == val.ValueName {
			return idx, nil
		}
	}
	return -1, ctx.NewError("enum value '%s' not found in type '%s'", val.ValueName, val.TypeName)
}

// Succ implements the Succ() built-in function.
// It returns the successor of an ordinal value (Integer or Enum):
// Succ(x) or Succ(x, delta).
func Succ(ctx Context, args []Value) Value {
	return stepOrdinal(ctx, "Succ", args, 1)
}

// Pred implements the Pred() built-in function.
// It returns the predecessor of an ordinal value (Integer or Enum):
// Pred(x) or Pred(x, delta).
func Pred(ctx Context, args []Value) Value {
	return stepOrdinal(ctx, "Pred", args, -1)
}

// stepOrdinal moves an ordinal value by direction*delta, where delta is the
// optional second argument (default 1). Variant deltas arrive already cast to
// Integer by the call-boundary coercion of the Integer parameter.
func stepOrdinal(ctx Context, name string, args []Value, direction int64) Value {
	if len(args) < 1 || len(args) > 2 {
		return ctx.NewError("%s() expects 1-2 arguments, got %d", name, len(args))
	}

	delta := int64(1)
	if len(args) == 2 {
		deltaVal := args[1]
		if variant, ok := deltaVal.(*runtime.VariantValue); ok {
			deltaVal = variant.UnwrapVariant()
		}
		deltaInt, ok := deltaVal.(*runtime.IntegerValue)
		if !ok {
			return ctx.NewError("%s() delta must be Integer, got %s", name, args[1].Type())
		}
		delta = deltaInt.Value
	}
	step := direction * delta

	switch val := args[0].(type) {
	case *runtime.IntegerValue:
		return &runtime.IntegerValue{Value: val.Value + step}

	case *runtime.EnumValue:
		return stepEnumValue(ctx, name, val, step)

	default:
		return ctx.NewError("%s() expects Integer or Enum, got %s", name, args[0].Type())
	}
}

// stepEnumValue returns the enum value step positions away from val.
func stepEnumValue(ctx Context, name string, val *runtime.EnumValue, step int64) Value {
	enumType, errVal := getEnumTypeForContext(ctx, val)
	if errVal != nil {
		return errVal
	}

	currentPos, errVal := findEnumPositionForContext(ctx, enumType, val)
	if errVal != nil {
		return errVal
	}

	target := int64(currentPos) + step
	if target >= int64(len(enumType.OrderedNames)) {
		return ctx.NewError("%s() cannot get successor of maximum enum value", name)
	}
	if target < 0 {
		return ctx.NewError("%s() cannot get predecessor of minimum enum value", name)
	}

	valueName := enumType.OrderedNames[target]
	return &runtime.EnumValue{
		TypeName:     val.TypeName,
		ValueName:    valueName,
		OrdinalValue: enumType.Values[valueName],
	}
}
