package evaluator

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/jsonvalue"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// isJSONNamespaceObject reports whether the member-access/call object is the
// built-in `JSON` static class used as a namespace, i.e. an identifier "JSON"
// that is not shadowed by a local variable.
func (e *Evaluator) isJSONNamespaceObject(obj ast.Expression, ctx *ExecutionContext) bool {
	identObj, ok := obj.(*ast.Identifier)
	if !ok || !ident.Equal(identObj.Value, "JSON") {
		return false
	}
	_, exists := ctx.Env().Get(identObj.Value)
	return !exists
}

// evalJSONNamespaceCall dispatches JSON.<method>(args). argExprs is nil for bare
// parameterless access such as `JSON.NewArray`.
func (e *Evaluator) evalJSONNamespaceCall(method string, argExprs []ast.Expression, node ast.Node, ctx *ExecutionContext) Value {
	args := make([]Value, len(argExprs))
	for i, argExpr := range argExprs {
		v := e.Eval(argExpr, ctx)
		if isError(v) {
			return v
		}
		args[i] = v
	}

	switch ident.Normalize(method) {
	case "newobject":
		return runtime.BoxVariantWithJSON(jsonvalue.NewObject())
	case "newarray":
		return runtime.BoxVariantWithJSON(jsonvalue.NewArray())
	case "parse":
		return e.jsonParse(jsonArgString(argValue(args, 0)), node)
	case "parseutf8":
		return e.jsonParse(decodeUTF8Bytes(jsonArgString(argValue(args, 0))), node)
	case "serialize":
		return e.jsonParse(jsonvalue.Stringify(e.valueToJSONValue(argValue(args, 0), node, ctx)), node)
	case "stringify":
		return &runtime.StringValue{Value: jsonvalue.Stringify(e.valueToJSONValue(argValue(args, 0), node, ctx))}
	case "stringifyutf8":
		return &runtime.StringValue{Value: encodeUTF8Bytes(jsonvalue.Stringify(e.valueToJSONValue(argValue(args, 0), node, ctx)))}
	case "prettystringify":
		indent := "\t"
		if len(args) >= 2 {
			indent = jsonArgString(args[1])
		}
		return &runtime.StringValue{Value: jsonvalue.StringifyPretty(e.valueToJSONValue(argValue(args, 0), node, ctx), indent)}
	case "parseintegerarray":
		nullVal := int64(0)
		if len(args) >= 2 {
			if i, ok := ExtractIntegerIndex(args[1]); ok {
				nullVal = int64(i)
			}
		}
		return e.jsonParseTypedArray(jsonArgString(argValue(args, 0)), "int", nullVal, node)
	case "parsefloatarray":
		return e.jsonParseTypedArray(jsonArgString(argValue(args, 0)), "float", 0, node)
	case "parsestringarray":
		return e.jsonParseTypedArray(jsonArgString(argValue(args, 0)), "string", 0, node)
	default:
		return e.newError(node, "There is no accessible member with name %q for type JSON", method)
	}
}

// jsonParseTypedArray parses a JSON array string into a dynamic array of the
// given scalar kind ("int"/"float"/"string"). JSON nulls become nullVal (integers),
// 0 (floats), or "" (strings).
func (e *Evaluator) jsonParseTypedArray(s, kind string, nullVal int64, node ast.Node) Value {
	jv, err := jsonvalue.Parse(strings.TrimSpace(s))
	if err != nil {
		return e.newError(node, "JSON parse error: %s", err.Error())
	}
	elems := make([]Value, 0)
	if jv != nil && jv.Kind() == jsonvalue.KindArray {
		for _, item := range jv.ArrayElements() {
			elems = append(elems, jsonScalarToTyped(item, kind, nullVal))
		}
	}
	return newTypedArray(elems, kind)
}

// jsonScalarToTyped converts a JSON scalar to the Integer/Float/String runtime
// value used by the Parse*Array helpers, substituting the null placeholder.
func jsonScalarToTyped(item *jsonvalue.Value, kind string, nullVal int64) Value {
	isNull := item == nil || item.Kind() == jsonvalue.KindNull || item.Kind() == jsonvalue.KindUndefined
	switch kind {
	case "int":
		if isNull {
			return &runtime.IntegerValue{Value: nullVal}
		}
		if i, ok := (&runtime.JSONValue{Value: item}).AsInteger(); ok {
			return &runtime.IntegerValue{Value: i}
		}
		return &runtime.IntegerValue{Value: nullVal}
	case "float":
		if isNull {
			return &runtime.FloatValue{Value: 0}
		}
		if f, ok := (&runtime.JSONValue{Value: item}).AsFloat(); ok {
			return &runtime.FloatValue{Value: f}
		}
		return &runtime.FloatValue{Value: 0}
	default: // string
		if isNull {
			return &runtime.StringValue{Value: ""}
		}
		return &runtime.StringValue{Value: item.StringValue()}
	}
}

// newTypedArray wraps elements in a dynamic ArrayValue of the given element kind.
func newTypedArray(elems []Value, kind string) Value {
	var elemType types.Type
	switch kind {
	case "int":
		elemType = types.INTEGER
	case "float":
		elemType = types.FLOAT
	default:
		elemType = types.STRING
	}
	runtimeElems := make([]runtime.Value, len(elems))
	copy(runtimeElems, elems)
	return &runtime.ArrayValue{
		Elements:  runtimeElems,
		ArrayType: types.NewDynamicArrayType(elemType),
	}
}

// jsonParse parses s, returning an Undefined JSONVariant for an empty/blank input
// (matching DWScript's JSON.Parse(”)).
func (e *Evaluator) jsonParse(s string, node ast.Node) Value {
	if strings.TrimSpace(s) == "" {
		return runtime.BoxVariantWithJSON(jsonvalue.NewUndefined())
	}
	jv, err := jsonvalue.Parse(s)
	if err != nil {
		return e.newError(node, "JSON parse error: %s", err.Error())
	}
	return runtime.BoxVariantWithJSON(jv)
}

// argValue returns args[i] or nil when out of range.
func argValue(args []Value, i int) Value {
	if i < 0 || i >= len(args) {
		return nil
	}
	return args[i]
}

// jsonArgString extracts a Go string from a script argument (a String value keeps
// its raw content; other values use their string rendering).
func jsonArgString(v Value) string {
	if v == nil {
		return ""
	}
	v = unwrapVariant(v)
	if sv, ok := v.(*runtime.StringValue); ok {
		return sv.Value
	}
	return v.String()
}

// encodeUTF8Bytes reinterprets the UTF-8 bytes of s as a Latin-1 code-point
// sequence, matching DWScript's *UTF8 helpers that treat the byte stream as raw.
func encodeUTF8Bytes(s string) string {
	bytes := []byte(s)
	runes := make([]rune, len(bytes))
	for i, b := range bytes {
		runes[i] = rune(b)
	}
	return string(runes)
}

// decodeUTF8Bytes is the inverse of encodeUTF8Bytes: it packs each code point
// (assumed to be a byte) back into a UTF-8 byte stream and decodes it.
func decodeUTF8Bytes(s string) string {
	bytes := make([]byte, 0, len(s))
	for _, r := range s {
		bytes = append(bytes, byte(r))
	}
	return string(bytes)
}
