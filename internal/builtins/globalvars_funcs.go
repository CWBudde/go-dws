package builtins

import (
	"math"
	"strings"
	"time"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// GlobalVars built-in function bindings
// ============================================================================
//
// The functions in this file adapt DefaultGlobalVars to the Context-based
// BuiltinFunc calling convention. Functions with var parameters
// (TryReadGlobalVar, GlobalQueuePull/Pop/Peek/First) cannot be expressed here
// because the registry hands builtins already-evaluated values; they are
// dispatched by the evaluator instead and only registered here so that
// semantic analysis knows their signatures.

// globalVarName extracts a string argument, reporting a DWScript-style error.
func globalVarName(ctx Context, fnName string, value Value) (string, Value) {
	str, ok := ctx.UnwrapVariant(value).(*runtime.StringValue)
	if !ok {
		return "", ctx.NewError("%s() expects a String name, got %s", fnName, value.Type())
	}
	return str.Value, nil
}

// optionalMask returns args[0] as a mask, defaulting to "*" when absent.
func optionalMask(ctx Context, fnName string, args []Value) (string, Value) {
	if len(args) == 0 {
		return "*", nil
	}
	return globalVarName(ctx, fnName, args[0])
}

// storableArg converts an argument into a storable Variant.
//
// JSON values are flattened to their serialized text first: a global only ever
// holds a simple Variant, and DWScript's JSON serialization already yields a
// String, so a stored JSON document reads back as its textual form.
func storableArg(ctx Context, value Value) (GlobalVarValue, Value) {
	unwrapped := ctx.UnwrapVariant(value)
	if runtime.KindOf(unwrapped) == runtime.KindJSON {
		text, err := ctx.ValueToJSON(unwrapped, false)
		if err != nil {
			return GlobalVarValue{}, ctx.NewError("%s", err.Error())
		}
		return GlobalVarValue{Kind: GlobalVarString, Str: text}, nil
	}
	stored, err := FromRuntimeValue(unwrapped)
	if err != nil {
		return GlobalVarValue{}, ctx.NewError("%s", err.Error())
	}
	return stored, nil
}

// ----------------------------------------------------------------------------
// Global variables
// ----------------------------------------------------------------------------

// WriteGlobalVar implements WriteGlobalVar(name, value [, expirationSeconds]).
// The optional third argument gives the value a lifetime in seconds; omitting
// it (or passing zero) clears any previous expiration.
func WriteGlobalVar(ctx Context, args []Value) Value {
	if len(args) < 2 || len(args) > 3 {
		return ctx.NewError("WriteGlobalVar() expects 2 or 3 arguments, got %d", len(args))
	}
	name, errVal := globalVarName(ctx, "WriteGlobalVar", args[0])
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
			return ctx.NewError("WriteGlobalVar() expects a Float expiration, got %s", args[2].Type())
		}
		expire = seconds
	}
	DefaultGlobalVars.Write(name, stored, expire)
	return &runtime.NilValue{}
}

// ReadGlobalVar implements ReadGlobalVar(name): Variant.
// A missing or expired global reads as Unassigned.
func ReadGlobalVar(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("ReadGlobalVar() expects exactly 1 argument, got %d", len(args))
	}
	name, errVal := globalVarName(ctx, "ReadGlobalVar", args[0])
	if errVal != nil {
		return errVal
	}
	stored, _ := DefaultGlobalVars.Read(name)
	return stored.ToRuntimeValue()
}

// ReadGlobalVarDef implements ReadGlobalVarDef(name, default): Variant.
// The default is returned when the global is missing or has expired.
func ReadGlobalVarDef(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("ReadGlobalVarDef() expects exactly 2 arguments, got %d", len(args))
	}
	name, errVal := globalVarName(ctx, "ReadGlobalVarDef", args[0])
	if errVal != nil {
		return errVal
	}
	stored, found := DefaultGlobalVars.Read(name)
	if !found {
		return ctx.UnwrapVariant(args[1])
	}
	return stored.ToRuntimeValue()
}

// DeleteGlobalVar implements DeleteGlobalVar(name): Boolean.
// It reports whether the global existed.
func DeleteGlobalVar(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("DeleteGlobalVar() expects exactly 1 argument, got %d", len(args))
	}
	name, errVal := globalVarName(ctx, "DeleteGlobalVar", args[0])
	if errVal != nil {
		return errVal
	}
	return &runtime.BooleanValue{Value: DefaultGlobalVars.Delete(name)}
}

// CleanupGlobalVars implements CleanupGlobalVars([mask]).
// Without a mask every global is removed.
func CleanupGlobalVars(ctx Context, args []Value) Value {
	if len(args) > 1 {
		return ctx.NewError("CleanupGlobalVars() expects at most 1 argument, got %d", len(args))
	}
	mask, errVal := optionalMask(ctx, "CleanupGlobalVars", args)
	if errVal != nil {
		return errVal
	}
	DefaultGlobalVars.Cleanup(mask)
	return &runtime.NilValue{}
}

// GlobalVarsNames implements GlobalVarsNames(mask): array of String.
func GlobalVarsNames(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("GlobalVarsNames() expects exactly 1 argument, got %d", len(args))
	}
	mask, errVal := globalVarName(ctx, "GlobalVarsNames", args[0])
	if errVal != nil {
		return errVal
	}
	return ctx.CreateStringArray(DefaultGlobalVars.Names(mask))
}

// GlobalVarsNamesCommaText implements GlobalVarsNamesCommaText: String.
// It returns every global name joined by commas.
func GlobalVarsNamesCommaText(ctx Context, args []Value) Value {
	if len(args) != 0 {
		return ctx.NewError("GlobalVarsNamesCommaText expects no arguments, got %d", len(args))
	}
	return &runtime.StringValue{Value: strings.Join(DefaultGlobalVars.Names("*"), ",")}
}

// IncrementGlobalVar implements IncrementGlobalVar(name [, delta [, expiration]]): Integer.
// It returns the incremented value. The expiration is always reset, so
// incrementing without an explicit lifetime clears a previous one.
func IncrementGlobalVar(ctx Context, args []Value) Value {
	if len(args) < 1 || len(args) > 3 {
		return ctx.NewError("IncrementGlobalVar() expects 1 to 3 arguments, got %d", len(args))
	}
	name, errVal := globalVarName(ctx, "IncrementGlobalVar", args[0])
	if errVal != nil {
		return errVal
	}
	delta := int64(1)
	if len(args) >= 2 {
		value, ok := ctx.ToInt64(ctx.UnwrapVariant(args[1]))
		if !ok {
			return ctx.NewError("IncrementGlobalVar() expects an Integer delta, got %s", args[1].Type())
		}
		delta = value
	}
	expire := 0.0
	if len(args) == 3 {
		seconds, ok := ctx.ToFloat64(ctx.UnwrapVariant(args[2]))
		if !ok {
			return ctx.NewError("IncrementGlobalVar() expects a Float expiration, got %s", args[2].Type())
		}
		expire = seconds
	}
	return &runtime.IntegerValue{Value: DefaultGlobalVars.Increment(name, delta, expire)}
}

// CompareExchangeGlobalVar implements
// CompareExchangeGlobalVar(name, value, comparand): Variant.
// The value is written only when the current value equals the comparand; the
// previous value is returned either way.
func CompareExchangeGlobalVar(ctx Context, args []Value) Value {
	if len(args) != 3 {
		return ctx.NewError("CompareExchangeGlobalVar() expects exactly 3 arguments, got %d", len(args))
	}
	name, errVal := globalVarName(ctx, "CompareExchangeGlobalVar", args[0])
	if errVal != nil {
		return errVal
	}
	value, errVal := storableArg(ctx, args[1])
	if errVal != nil {
		return errVal
	}
	comparand, errVal := storableArg(ctx, args[2])
	if errVal != nil {
		return errVal
	}
	return DefaultGlobalVars.CompareExchange(name, value, comparand).ToRuntimeValue()
}

// SaveGlobalVarsToString implements SaveGlobalVarsToString: String.
// The result is an opaque snapshot accepted by LoadGlobalVarsFromString.
func SaveGlobalVarsToString(ctx Context, args []Value) Value {
	if len(args) != 0 {
		return ctx.NewError("SaveGlobalVarsToString expects no arguments, got %d", len(args))
	}
	return &runtime.StringValue{Value: DefaultGlobalVars.SaveToString()}
}

// LoadGlobalVarsFromString implements LoadGlobalVarsFromString(data).
// It replaces every global with the snapshot's contents. An empty string
// clears the store; a payload with an unrecognized header raises
// "Invalid file tag".
func LoadGlobalVarsFromString(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("LoadGlobalVarsFromString() expects exactly 1 argument, got %d", len(args))
	}
	data, errVal := globalVarName(ctx, "LoadGlobalVarsFromString", args[0])
	if errVal != nil {
		return errVal
	}
	if err := DefaultGlobalVars.LoadFromString(data); err != nil {
		return ctx.NewError("%s", err.Error())
	}
	return &runtime.NilValue{}
}

// ----------------------------------------------------------------------------
// Global queues
// ----------------------------------------------------------------------------

// GlobalQueuePush implements GlobalQueuePush(name, value): Boolean.
// The value is appended to the back of the queue.
func GlobalQueuePush(ctx Context, args []Value) Value {
	return globalQueueAdd(ctx, "GlobalQueuePush", args, DefaultGlobalVars.QueuePush)
}

// GlobalQueueInsert implements GlobalQueueInsert(name, value): Boolean.
// The value is prepended to the front of the queue.
func GlobalQueueInsert(ctx Context, args []Value) Value {
	return globalQueueAdd(ctx, "GlobalQueueInsert", args, DefaultGlobalVars.QueueInsert)
}

// globalQueueAdd is the shared body of GlobalQueuePush and GlobalQueueInsert.
func globalQueueAdd(ctx Context, fnName string, args []Value, add func(string, GlobalVarValue)) Value {
	if len(args) != 2 {
		return ctx.NewError("%s() expects exactly 2 arguments, got %d", fnName, len(args))
	}
	name, errVal := globalVarName(ctx, fnName, args[0])
	if errVal != nil {
		return errVal
	}
	stored, errVal := storableArg(ctx, args[1])
	if errVal != nil {
		return errVal
	}
	add(name, stored)
	return &runtime.BooleanValue{Value: true}
}

// GlobalQueueLength implements GlobalQueueLength(name): Integer.
func GlobalQueueLength(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("GlobalQueueLength() expects exactly 1 argument, got %d", len(args))
	}
	name, errVal := globalVarName(ctx, "GlobalQueueLength", args[0])
	if errVal != nil {
		return errVal
	}
	return &runtime.IntegerValue{Value: int64(DefaultGlobalVars.QueueLength(name))}
}

// CleanupGlobalQueues implements CleanupGlobalQueues([mask]).
func CleanupGlobalQueues(ctx Context, args []Value) Value {
	if len(args) > 1 {
		return ctx.NewError("CleanupGlobalQueues() expects at most 1 argument, got %d", len(args))
	}
	mask, errVal := optionalMask(ctx, "CleanupGlobalQueues", args)
	if errVal != nil {
		return errVal
	}
	DefaultGlobalVars.CleanupQueues(mask)
	return &runtime.NilValue{}
}

// GlobalQueueSnapshot implements GlobalQueueSnapshot(name): array of Variant.
// Elements are ordered front first and the queue is left untouched.
func GlobalQueueSnapshot(ctx Context, args []Value) Value {
	entries, errVal := globalQueueSnapshotEntries(ctx, "GlobalQueueSnapshot", args)
	if errVal != nil {
		return errVal
	}
	values := make([]Value, len(entries))
	for idx, entry := range entries {
		values[idx] = entry.ToRuntimeValue()
	}
	return ctx.CreateVariantArray(values)
}

// GlobalQueueSnapshotIntegers implements GlobalQueueSnapshotIntegers(name): array of Integer.
// Floats are rounded to the nearest even integer; strings raise a cast error.
func GlobalQueueSnapshotIntegers(ctx Context, args []Value) Value {
	entries, errVal := globalQueueSnapshotEntries(ctx, "GlobalQueueSnapshotIntegers", args)
	if errVal != nil {
		return errVal
	}
	elements := make([]Value, len(entries))
	for idx, entry := range entries {
		number, ok := globalVarAsInteger(entry)
		if !ok {
			return ctx.NewError("Could not cast variant from %s to Integer", globalVarTypeName(entry))
		}
		elements[idx] = &runtime.IntegerValue{Value: number}
	}
	return newTypedArray(elements, types.INTEGER)
}

// GlobalQueueSnapshotFloats implements GlobalQueueSnapshotFloats(name): array of Float.
func GlobalQueueSnapshotFloats(ctx Context, args []Value) Value {
	entries, errVal := globalQueueSnapshotEntries(ctx, "GlobalQueueSnapshotFloats", args)
	if errVal != nil {
		return errVal
	}
	elements := make([]Value, len(entries))
	for idx, entry := range entries {
		number, ok := globalVarAsFloat(entry)
		if !ok {
			return ctx.NewError("Could not cast variant from %s to Float", globalVarTypeName(entry))
		}
		elements[idx] = &runtime.FloatValue{Value: number}
	}
	return newTypedArray(elements, types.FLOAT)
}

// GlobalQueueSnapshotStrings implements GlobalQueueSnapshotStrings(name): array of String.
func GlobalQueueSnapshotStrings(ctx Context, args []Value) Value {
	entries, errVal := globalQueueSnapshotEntries(ctx, "GlobalQueueSnapshotStrings", args)
	if errVal != nil {
		return errVal
	}
	texts := make([]string, len(entries))
	for idx, entry := range entries {
		texts[idx] = globalVarAsString(entry)
	}
	return ctx.CreateStringArray(texts)
}

// globalQueueSnapshotEntries validates the single name argument shared by every
// snapshot function and returns the queue contents.
func globalQueueSnapshotEntries(ctx Context, fnName string, args []Value) ([]GlobalVarValue, Value) {
	if len(args) != 1 {
		return nil, ctx.NewError("%s() expects exactly 1 argument, got %d", fnName, len(args))
	}
	name, errVal := globalVarName(ctx, fnName, args[0])
	if errVal != nil {
		return nil, errVal
	}
	return DefaultGlobalVars.QueueSnapshot(name), nil
}

// newTypedArray builds a dynamic array runtime value with the given element type.
func newTypedArray(elements []Value, elementType types.Type) Value {
	return &runtime.ArrayValue{
		Elements:  elements,
		ArrayType: types.NewDynamicArrayType(elementType),
	}
}

// globalVarTypeName returns the DWScript type name used in cast diagnostics.
func globalVarTypeName(v GlobalVarValue) string {
	switch v.Kind {
	case GlobalVarInteger:
		return "Integer"
	case GlobalVarFloat:
		return "Float"
	case GlobalVarString:
		return "String"
	case GlobalVarBoolean:
		return "Boolean"
	case GlobalVarNull:
		return "Null"
	case GlobalVarUnassigned:
		return "Unassigned"
	}
	return "Variant"
}

// globalVarAsInteger converts a stored Variant to an integer using DWScript's
// banker's rounding for floats. Strings are not implicitly convertible.
func globalVarAsInteger(v GlobalVarValue) (int64, bool) {
	switch v.Kind {
	case GlobalVarInteger:
		return v.Int, true
	case GlobalVarFloat:
		return int64(math.RoundToEven(v.Float)), true
	case GlobalVarBoolean:
		if v.Bool {
			return 1, true
		}
		return 0, true
	case GlobalVarString, GlobalVarNull, GlobalVarUnassigned:
		return 0, false
	}
	return 0, false
}

// globalVarAsFloat converts a stored Variant to a float.
// Strings are not implicitly convertible.
func globalVarAsFloat(v GlobalVarValue) (float64, bool) {
	switch v.Kind {
	case GlobalVarInteger:
		return float64(v.Int), true
	case GlobalVarFloat:
		return v.Float, true
	case GlobalVarBoolean:
		if v.Bool {
			return 1, true
		}
		return 0, true
	case GlobalVarString, GlobalVarNull, GlobalVarUnassigned:
		return 0, false
	}
	return 0, false
}

// globalVarAsString renders a stored Variant using the runtime's own
// value formatting, so numbers match Print/String() output.
func globalVarAsString(v GlobalVarValue) string {
	if v.Kind == GlobalVarUnassigned {
		return ""
	}
	return v.ToRuntimeValue().String()
}

// ----------------------------------------------------------------------------
// Var-parameter placeholders
// ----------------------------------------------------------------------------

// varParamOnlyGlobalVarFunc builds the stub registered for functions the
// evaluator intercepts before argument evaluation. Reaching the stub means the
// call did not take the var-parameter path.
func varParamOnlyGlobalVarFunc(name string) BuiltinFunc {
	return func(ctx Context, _ []Value) Value {
		return ctx.NewError("%s() requires a variable as its last argument", name)
	}
}

// ----------------------------------------------------------------------------
// Sleep
// ----------------------------------------------------------------------------

// Sleep implements Sleep(milliseconds), suspending the current script.
// Negative or zero delays return immediately.
func Sleep(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Sleep() expects exactly 1 argument, got %d", len(args))
	}
	millis, ok := ctx.ToInt64(ctx.UnwrapVariant(args[0]))
	if !ok {
		return ctx.NewError("Sleep() expects an Integer, got %s", args[0].Type())
	}
	if millis > 0 {
		time.Sleep(time.Duration(millis) * time.Millisecond)
	}
	return &runtime.NilValue{}
}
