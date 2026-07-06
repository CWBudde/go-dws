package evaluator

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/jsonvalue"
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
		return e.jsonParse(jsonvalue.Stringify(ValueToJSONValue(argValue(args, 0))), node)
	case "stringify":
		return &runtime.StringValue{Value: jsonvalue.Stringify(ValueToJSONValue(argValue(args, 0)))}
	case "stringifyutf8":
		return &runtime.StringValue{Value: encodeUTF8Bytes(jsonvalue.Stringify(ValueToJSONValue(argValue(args, 0))))}
	case "prettystringify":
		indent := "\t"
		if len(args) >= 2 {
			indent = jsonArgString(args[1])
		}
		return &runtime.StringValue{Value: jsonvalue.StringifyPretty(ValueToJSONValue(argValue(args, 0)), indent)}
	default:
		return e.newError(node, "There is no accessible member with name %q for type JSON", method)
	}
}

// jsonParse parses s, returning an Undefined JSONVariant for an empty/blank input
// (matching DWScript's JSON.Parse('')).
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
