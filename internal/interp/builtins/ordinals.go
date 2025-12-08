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
// It returns the successor of an ordinal value (Integer or Enum).
func Succ(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Succ() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	switch val := arg.(type) {
	case *runtime.IntegerValue:
		return &runtime.IntegerValue{Value: val.Value + 1}

	case *runtime.EnumValue:
		return succEnumValue(ctx, val)

	default:
		return ctx.NewError("Succ() expects Integer or Enum, got %s", arg.Type())
	}
}

// succEnumValue computes the successor enum value for Succ().
func succEnumValue(ctx Context, val *runtime.EnumValue) Value {
	enumType, errVal := getEnumTypeForContext(ctx, val)
	if errVal != nil {
		return errVal
	}

	currentPos, errVal := findEnumPositionForContext(ctx, enumType, val)
	if errVal != nil {
		return errVal
	}

	if currentPos >= len(enumType.OrderedNames)-1 {
		return ctx.NewError("Succ() cannot get successor of maximum enum value")
	}

	nextValueName := enumType.OrderedNames[currentPos+1]
	nextOrdinal := enumType.Values[nextValueName]
	return &runtime.EnumValue{
		TypeName:     val.TypeName,
		ValueName:    nextValueName,
		OrdinalValue: nextOrdinal,
	}
}

// Pred implements the Pred() built-in function.
// It returns the predecessor of an ordinal value (Integer or Enum).
func Pred(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Pred() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	switch val := arg.(type) {
	case *runtime.IntegerValue:
		return &runtime.IntegerValue{Value: val.Value - 1}

	case *runtime.EnumValue:
		return predEnumValue(ctx, val)

	default:
		return ctx.NewError("Pred() expects Integer or Enum, got %s", arg.Type())
	}
}

// predEnumValue computes the predecessor enum value for Pred().
func predEnumValue(ctx Context, val *runtime.EnumValue) Value {
	enumType, errVal := getEnumTypeForContext(ctx, val)
	if errVal != nil {
		return errVal
	}

	currentPos, errVal := findEnumPositionForContext(ctx, enumType, val)
	if errVal != nil {
		return errVal
	}

	if currentPos <= 0 {
		return ctx.NewError("Pred() cannot get predecessor of minimum enum value")
	}

	prevValueName := enumType.OrderedNames[currentPos-1]
	prevOrdinal := enumType.Values[prevValueName]
	return &runtime.EnumValue{
		TypeName:     val.TypeName,
		ValueName:    prevValueName,
		OrdinalValue: prevOrdinal,
	}
}
