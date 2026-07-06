package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// isJSONNamespace reports whether obj is the built-in `JSON` static class used as
// a namespace (JSON.Parse, JSON.Stringify, ...). It is treated as a namespace only
// when no user symbol named JSON shadows it.
func (a *Analyzer) isJSONNamespace(obj ast.Expression) bool {
	identExpr, ok := obj.(*ast.Identifier)
	if !ok || !ident.Equal(identExpr.Value, "JSON") {
		return false
	}
	_, resolved := a.symbols.Resolve(identExpr.Value)
	return !resolved
}

// jsonNamespaceMemberType returns the result type of a JSON static method /
// factory (used for both bare access like JSON.NewArray and calls like
// JSON.Parse(s)).
func jsonNamespaceMemberType(methodName string) types.Type {
	switch ident.Normalize(methodName) {
	case "parse", "parseutf8", "serialize", "newobject", "newarray":
		return types.JSON_VARIANT
	case "stringify", "stringifyutf8", "prettystringify":
		return types.STRING
	case "parseintegerarray":
		return types.NewDynamicArrayType(types.INTEGER)
	case "parsefloatarray":
		return types.NewDynamicArrayType(types.FLOAT)
	case "parsestringarray":
		return types.NewDynamicArrayType(types.STRING)
	default:
		// Unknown JSON member: lenient for now (error-detection parity is a later
		// phase). Treat as a browsable JSON value.
		return types.JSON_VARIANT
	}
}

// analyzeJSONNamespaceResult analyzes the arguments of a JSON.<method>(args) call
// and returns the method's result type.
func (a *Analyzer) analyzeJSONNamespaceResult(methodName string, args []ast.Expression) types.Type {
	for _, arg := range args {
		a.analyzeExpression(arg)
	}
	return jsonNamespaceMemberType(methodName)
}

// analyzeJSONMethodResult analyzes the arguments of a method call on a JSONVariant
// receiver (v.TypeName(), v.Add(x), ...) and returns the connector method's result
// type.
func (a *Analyzer) analyzeJSONMethodResult(methodName string, args []ast.Expression) types.Type {
	for _, arg := range args {
		a.analyzeExpression(arg)
	}
	switch ident.Normalize(methodName) {
	case "typename", "elementname", "tostring":
		return types.STRING
	case "defined":
		return types.BOOLEAN
	case "length", "low", "high", "add", "push":
		return types.INTEGER
	case "clone":
		return types.JSON_VARIANT
	case "extend", "addfrom", "delete", "swap":
		return types.VOID
	default:
		// Unknown method on a JSONVariant: treat as a browsable value.
		return types.JSON_VARIANT
	}
}
