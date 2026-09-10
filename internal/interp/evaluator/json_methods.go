package evaluator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/jsonvalue"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// jsonValueOf extracts the underlying jsonvalue.Value from a (possibly
// variant-boxed) JSON value, or nil if v is not a JSON value.
func jsonValueOf(v Value) *jsonvalue.Value {
	if v == nil {
		return nil
	}
	return extractJSONValueViaReflection(unwrapVariant(v))
}

// coerceToJSONVariant auto-boxes a base scalar (or nil/unassigned) into a JSON
// immediate when it is assigned to a JSONVariant target. Values that are already
// JSON pass through unchanged.
func coerceToJSONVariant(v Value) Value {
	if v == nil {
		return boxJSON(jsonvalue.NewUndefined())
	}
	if isJSONBoxed(v) {
		return v
	}
	switch u := unwrapVariant(v).(type) {
	case *runtime.StringValue:
		return boxJSON(jsonvalue.NewString(u.Value))
	case *runtime.IntegerValue:
		return boxJSON(jsonvalue.NewInt64(u.Value))
	case *runtime.FloatValue:
		return boxJSON(jsonvalue.NewNumber(u.Value))
	case *runtime.BooleanValue:
		return boxJSON(jsonvalue.NewBoolean(u.Value))
	case *runtime.NullValue:
		return boxJSON(jsonvalue.NewNull())
	case *runtime.NilValue:
		return boxJSON(jsonvalue.NewNull())
	case *runtime.UnassignedValue:
		return boxJSON(jsonvalue.NewUndefined())
	default:
		return v
	}
}

// isJSONBoxed reports whether v (after unwrapping a Variant) is a JSON value.
func isJSONBoxed(v Value) bool {
	if v == nil {
		return false
	}
	return runtime.KindOf(unwrapVariant(v)) == runtime.KindJSON
}

// boxJSON wraps a jsonvalue.Value into a variant-boxed JSONValue, materializing a
// missing (nil) value as Undefined so it remains browsable.
func boxJSON(jv *jsonvalue.Value) Value {
	if jv == nil {
		jv = jsonvalue.NewUndefined()
	}
	return runtime.BoxVariantWithJSON(jv)
}

// jsonTypeName maps a jsonvalue.Kind to the DWScript ValueType string.
func jsonTypeName(jv *jsonvalue.Value) string {
	kind := jsonvalue.KindUndefined
	if jv != nil {
		kind = jv.Kind()
	}
	switch kind {
	case jsonvalue.KindNull:
		return "Null"
	case jsonvalue.KindObject:
		return "Object"
	case jsonvalue.KindArray:
		return "Array"
	case jsonvalue.KindString:
		return "String"
	case jsonvalue.KindNumber, jsonvalue.KindInt64:
		return "Number"
	case jsonvalue.KindBoolean:
		return "Boolean"
	default:
		return "Undefined"
	}
}

// jsonElementCount returns the number of elements for arrays/objects, else 0.
func jsonElementCount(jv *jsonvalue.Value) int {
	if jv == nil {
		return 0
	}
	switch jv.Kind() {
	case jsonvalue.KindArray:
		return jv.ArrayLen()
	case jsonvalue.KindObject:
		return len(jv.ObjectKeys())
	default:
		return 0
	}
}

// jsonAssignValue converts a script value into the jsonvalue stored on assignment
// into a JSON container. An undefined/unassigned/nil value becomes JSON null (a
// container slot cannot hold Undefined); a JSON value is stored by reference so
// later mutation is visible.
func jsonAssignValue(v Value) *jsonvalue.Value {
	u := unwrapVariant(v)
	if u == nil {
		return jsonvalue.NewNull()
	}
	switch runtime.KindOf(u) {
	case runtime.KindUnassigned, runtime.KindNil, runtime.KindNull:
		return jsonvalue.NewNull()
	case runtime.KindJSON:
		jv := extractJSONValueViaReflection(u)
		if jv == nil || jv.Kind() == jsonvalue.KindUndefined {
			return jsonvalue.NewNull()
		}
		return jv
	default:
		return ValueToJSONValue(u)
	}
}

// assignJSONMember implements `jsonValue.member := value`.
func (e *Evaluator) assignJSONMember(jv *jsonvalue.Value, name string, value Value, node ast.Node, ctx *ExecutionContext) Value {
	if jv == nil || jv.Kind() == jsonvalue.KindUndefined {
		e.builtinContext(ctx).RaiseException("Exception", fmt.Sprintf(`Cannot set member "%s" of Undefined`, name), nil)
		return &runtime.NilValue{}
	}
	switch jv.Kind() {
	case jsonvalue.KindObject:
		jv.ObjectSet(name, jsonAssignValue(value))
	case jsonvalue.KindArray:
		e.builtinContext(ctx).RaiseException("Exception", fmt.Sprintf(`Invalid array member "%s"`, name), nil)
	default:
		e.builtinContext(ctx).RaiseException("Exception", fmt.Sprintf(`Cannot set member "%s" of Immediate`, name), nil)
	}
	return &runtime.NilValue{}
}

// assignJSONIndex implements `jsonValue[index] := value` for objects (string key)
// and arrays (integer index, auto-extending with nulls).
func (e *Evaluator) assignJSONIndex(jv *jsonvalue.Value, index Value, value Value, node ast.Node, ctx *ExecutionContext) Value {
	if jv == nil || jv.Kind() == jsonvalue.KindUndefined {
		e.builtinContext(ctx).RaiseException("Exception", "Cannot set items of Undefined", nil)
		return &runtime.NilValue{}
	}
	idx := unwrapVariant(index)
	switch jv.Kind() {
	case jsonvalue.KindObject:
		jv.ObjectSet(jsonArgString(idx), jsonAssignValue(value))
	case jsonvalue.KindArray:
		i, ok := ExtractIntegerIndex(idx)
		if !ok || i < 0 {
			return e.newError(node, "JSON array index must be a non-negative integer")
		}
		// Reparent before sizing the array: detaching the incoming node from
		// this very array would otherwise shrink it back under the new index.
		child := jsonAssignValue(value)
		child.Detach()
		for jv.ArrayLen() <= i {
			jv.ArrayAppend(jsonvalue.NewNull())
		}
		jv.ArraySet(i, child)
	default:
		e.builtinContext(ctx).RaiseException("Exception", fmt.Sprintf("Cannot set items of %s", jsonTypeName(jv)), nil)
	}
	return &runtime.NilValue{}
}

// evalJSONValueMember handles member access on a JSON value (v.foo, v.length).
// Any member other than the special `length` yields the object field value (or an
// Undefined JSON value when absent), so chains such as v.a.b stay browsable.
func (e *Evaluator) evalJSONValueMember(jv *jsonvalue.Value, memberName string) Value {
	if ident.Equal(memberName, "length") {
		if jv != nil {
			switch jv.Kind() {
			case jsonvalue.KindArray:
				return boxJSON(jsonvalue.NewInt64(int64(jv.ArrayLen())))
			case jsonvalue.KindString:
				return boxJSON(jsonvalue.NewInt64(int64(len([]rune(jv.StringValue())))))
			}
		}
	}
	if jv != nil && jv.Kind() == jsonvalue.KindObject {
		return boxJSON(jv.ObjectGet(memberName))
	}
	return boxJSON(nil)
}

// evalJSONMethodCall dispatches a method call on a JSON value receiver.
func (e *Evaluator) evalJSONMethodCall(recv Value, method string, args []Value, node ast.Node, ctx *ExecutionContext) Value {
	jv := jsonValueOf(recv)

	switch ident.Normalize(method) {
	case "typename":
		return &runtime.StringValue{Value: jsonTypeName(jv)}
	case "tostring":
		if jv == nil {
			jv = jsonvalue.NewUndefined()
		}
		return &runtime.StringValue{Value: jsonvalue.Stringify(jv)}
	case "defined":
		defined := jv != nil && jv.Kind() != jsonvalue.KindUndefined
		return &runtime.BooleanValue{Value: defined}
	case "length":
		return &runtime.IntegerValue{Value: int64(jsonElementCount(jv))}
	case "low":
		return &runtime.IntegerValue{Value: 0}
	case "high":
		return &runtime.IntegerValue{Value: int64(jsonElementCount(jv) - 1)}
	case "elementname":
		return e.jsonElementName(jv, args)
	case "clone":
		if jv == nil {
			return boxJSON(nil)
		}
		return boxJSON(jv.Clone())
	case "add", "push":
		return e.jsonArrayAdd(jv, args, node, ctx)
	case "addfrom":
		return e.jsonArrayAddFrom(jv, args, node)
	case "extend":
		return e.jsonExtend(jv, args, node)
	case "delete":
		return e.jsonDelete(jv, args, node)
	case "swap":
		return e.jsonSwap(jv, args, node)
	default:
		return e.newError(node, `Method "%s" not found in connector "JSON Connector 2.0"`, method)
	}
}

func (e *Evaluator) jsonElementName(jv *jsonvalue.Value, args []Value) Value {
	idx, _ := ExtractIntegerIndex(argValue(args, 0))
	if jv != nil {
		switch jv.Kind() {
		case jsonvalue.KindObject:
			keys := jv.ObjectKeys()
			if idx >= 0 && idx < len(keys) {
				return &runtime.StringValue{Value: keys[idx]}
			}
		case jsonvalue.KindArray:
			if idx >= 0 && idx < jv.ArrayLen() {
				return &runtime.StringValue{Value: strconv.Itoa(idx)}
			}
		}
	}
	return &runtime.StringValue{Value: ""}
}

func (e *Evaluator) jsonArrayAdd(jv *jsonvalue.Value, args []Value, node ast.Node, ctx *ExecutionContext) Value {
	if jv == nil || jv.Kind() != jsonvalue.KindArray {
		e.builtinContext(ctx).RaiseException("Exception", fmt.Sprintf("JSON method Add() unsupported for type %s", jsonTypeName(jv)), nil)
		return &runtime.NilValue{}
	}
	for _, arg := range args {
		if u := unwrapVariant(arg); u == nil || runtime.KindOf(u) == runtime.KindUnassigned {
			e.builtinContext(ctx).RaiseException("Exception", "JSON Array Add() unsupported type", nil)
			return &runtime.NilValue{}
		}
		jv.ArrayAppend(ValueToJSONValue(arg))
	}
	return &runtime.IntegerValue{Value: int64(jv.ArrayLen())}
}

// jsonArrayAddFrom appends the source array's elements to the receiver and empties
// the source (a move), matching DWScript's AddFrom semantics.
func (e *Evaluator) jsonArrayAddFrom(jv *jsonvalue.Value, args []Value, node ast.Node) Value {
	if jv == nil || jv.Kind() != jsonvalue.KindArray {
		return e.newError(node, "JSON AddFrom requires an array value")
	}
	src := jsonValueOf(argValue(args, 0))
	if src != nil && src.Kind() == jsonvalue.KindArray {
		for _, elem := range src.ArrayElements() {
			jv.ArrayAppend(elem)
		}
		src.ClearArray()
	}
	return &runtime.NilValue{}
}

// jsonExtend merges another JSON value into the receiver: object keys are copied
// (overwriting) and array elements are appended as copies, leaving the source
// intact, matching DWScript's Extend semantics.
func (e *Evaluator) jsonExtend(jv *jsonvalue.Value, args []Value, node ast.Node) Value {
	src := jsonValueOf(argValue(args, 0))
	if jv == nil || src == nil {
		return &runtime.NilValue{}
	}
	if jv.Kind() == jsonvalue.KindObject && src.Kind() == jsonvalue.KindObject {
		for _, k := range src.ObjectKeys() {
			jv.ObjectSet(k, src.ObjectGet(k).Clone())
		}
	} else if jv.Kind() == jsonvalue.KindArray && src.Kind() == jsonvalue.KindArray {
		for _, elem := range src.ArrayElements() {
			jv.ArrayAppend(elem.Clone())
		}
	}
	return &runtime.NilValue{}
}

func (e *Evaluator) jsonDelete(jv *jsonvalue.Value, args []Value, node ast.Node) Value {
	if jv == nil {
		return &runtime.NilValue{}
	}
	key := unwrapVariant(argValue(args, 0))
	switch jv.Kind() {
	case jsonvalue.KindObject:
		jv.ObjectDelete(jsonArgString(key))
	case jsonvalue.KindArray:
		if idx, ok := ExtractIntegerIndex(key); ok {
			jv.ArrayDelete(idx)
		}
	}
	return &runtime.NilValue{}
}

func (e *Evaluator) jsonSwap(jv *jsonvalue.Value, args []Value, node ast.Node) Value {
	if jv == nil || jv.Kind() != jsonvalue.KindArray {
		return e.newError(node, "JSON Swap requires an array value")
	}
	i, iok := ExtractIntegerIndex(argValue(args, 0))
	j, jok := ExtractIntegerIndex(argValue(args, 1))
	n := jv.ArrayLen()
	if !iok || !jok || i < 0 || i >= n || j < 0 || j >= n {
		return e.newError(node, "Upper bound exceeded! Index %d", i)
	}
	jv.ArraySwap(i, j)
	return &runtime.NilValue{}
}

// jsonScalarFloat applies DWScript's JSON immediate → Float conversion. Numbers
// and booleans convert directly and strings are parsed (an empty or unparsable
// string yields 0, matching StrToFloatDef). Objects, arrays, null and undefined
// are not values and report ok=false, which the caller turns into DWScript's
// "Not a value" exception.
func jsonScalarFloat(jv *jsonvalue.Value) (float64, bool) {
	if jv == nil {
		return 0, false
	}
	switch jv.Kind() {
	case jsonvalue.KindInt64:
		return float64(jv.Int64Value()), true
	case jsonvalue.KindNumber:
		return jv.NumberValue(), true
	case jsonvalue.KindBoolean:
		if jv.BoolValue() {
			return 1, true
		}
		return 0, true
	case jsonvalue.KindString:
		f, err := strconv.ParseFloat(strings.TrimSpace(jv.StringValue()), 64)
		if err != nil {
			return 0, true
		}
		return f, true
	default:
		return 0, false
	}
}

// jsonScalarInteger applies DWScript's JSON immediate → Integer conversion. It
// follows jsonScalarFloat and truncates, so Integer(json '1.25') is 1.
func jsonScalarInteger(jv *jsonvalue.Value) (int64, bool) {
	if jv != nil && jv.Kind() == jsonvalue.KindInt64 {
		return jv.Int64Value(), true
	}
	f, ok := jsonScalarFloat(jv)
	if !ok {
		return 0, false
	}
	return int64(f), true
}

// jsonNotAValue raises DWScript's "Not a value" exception, which is what casting
// a JSON object, array or undefined node to a scalar produces. The position is
// the statement being executed, matching DWScript's statement-level reporting.
func (e *Evaluator) jsonNotAValue(ctx *ExecutionContext) Value {
	msg := "Not a value"
	if ctx != nil {
		if stmt := ctx.CurrentStatement(); stmt != nil {
			pos := stmt.Pos()
			msg = fmt.Sprintf("%s [line: %d, column: %d]", msg, pos.Line, pos.Column)
		}
	}
	e.builtinContext(ctx).RaiseException("Exception", msg, nil)
	return &runtime.NilValue{}
}
