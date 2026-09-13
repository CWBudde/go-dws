package builtins

import "github.com/cwbudde/go-dws/internal/interp/runtime"

// privateVarUnit validates the lexical owner of a private-variable operation.
func privateVarUnit(ctx Context) (string, Value) {
	unit := ctx.CurrentUnit()
	if unit == "" {
		const message = "Private variables cannot be referred from main module"
		if raiser, ok := ctx.(interface {
			RaiseException(className, message string, pos any)
		}); ok {
			raiser.RaiseException("Exception", message, nil)
			return "", &runtime.NilValue{}
		}
		return "", ctx.NewError(message)
	}
	return unit, nil
}

// WritePrivateVar stores a simple Variant in the declaring unit's private store.
// The result is true for a new or expired entry, false for a live replacement.
func WritePrivateVar(ctx Context, args []Value) Value {
	if len(args) < 2 || len(args) > 3 {
		return ctx.NewError("WritePrivateVar() expects 2 or 3 arguments, got %d", len(args))
	}
	unit, errVal := privateVarUnit(ctx)
	if errVal != nil {
		return errVal
	}
	name, errVal := globalVarName(ctx, "WritePrivateVar", args[0])
	if errVal != nil {
		return errVal
	}
	stored, errVal := storableArg(ctx, args[1])
	if errVal != nil {
		return errVal
	}
	expire := 0.0
	if len(args) == 3 {
		seconds, ok := ctx.ToFloat64(ctx.UnwrapVariant(args[2]))
		if !ok {
			return ctx.NewError("WritePrivateVar() expects a Float expiration, got %s", args[2].Type())
		}
		expire = seconds
	}
	return &runtime.BooleanValue{Value: DefaultPrivateVars.Write(unit, name, stored, expire)}
}

// ReadPrivateVar reads a unit-private value, returning the optional default or
// Unassigned when absent. Direct script calls evaluate the default lazily in
// the evaluator; this binding also supports callers with already evaluated args.
func ReadPrivateVar(ctx Context, args []Value) Value {
	if len(args) < 1 || len(args) > 2 {
		return ctx.NewError("ReadPrivateVar() expects 1 or 2 arguments, got %d", len(args))
	}
	unit, errVal := privateVarUnit(ctx)
	if errVal != nil {
		return errVal
	}
	name, errVal := globalVarName(ctx, "ReadPrivateVar", args[0])
	if errVal != nil {
		return errVal
	}
	stored, found := DefaultPrivateVars.Read(unit, name)
	if !found && len(args) == 2 {
		return ctx.UnwrapVariant(args[1])
	}
	return stored.ToRuntimeValue()
}

// PrivateVarsNames returns the sorted names matching mask in the current unit.
func PrivateVarsNames(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("PrivateVarsNames() expects exactly 1 argument, got %d", len(args))
	}
	unit, errVal := privateVarUnit(ctx)
	if errVal != nil {
		return errVal
	}
	mask, errVal := globalVarName(ctx, "PrivateVarsNames", args[0])
	if errVal != nil {
		return errVal
	}
	return ctx.CreateStringArray(DefaultPrivateVars.Names(unit, mask))
}

// CleanupPrivateVars deletes private entries matching an optional mask.
// The omitted mask defaults to "*"; an empty mask matches only the empty name.
func CleanupPrivateVars(ctx Context, args []Value) Value {
	if len(args) > 1 {
		return ctx.NewError("CleanupPrivateVars() expects at most 1 argument, got %d", len(args))
	}
	unit, errVal := privateVarUnit(ctx)
	if errVal != nil {
		return errVal
	}
	mask, errVal := optionalMask(ctx, "CleanupPrivateVars", args)
	if errVal != nil {
		return errVal
	}
	DefaultPrivateVars.Cleanup(unit, mask)
	return &runtime.NilValue{}
}
