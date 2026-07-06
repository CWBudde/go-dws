package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// evalAssociativeArrayMethod dispatches the built-in methods of an associative
// array (`array [K] of E`): Keys, Length/Count, Clear, and Delete(key). It
// serves both bare member access (a.Keys, a.Clear) and call syntax
// (a.Delete(k)). Returns (value, true) when handled; (nil, false) otherwise so
// normal dispatch can proceed.
func (e *Evaluator) evalAssociativeArrayMethod(
	assoc *runtime.AssociativeArrayValue,
	name string,
	args []Value,
	node ast.Node,
) (Value, bool) {
	switch ident.Normalize(name) {
	case "keys":
		if len(args) != 0 {
			return e.newError(node, "Keys expects no arguments, got %d", len(args)), true
		}
		// Return the keys as a dynamic array (chainable: .Map/.Sort/.Join).
		keys := assoc.Keys()
		return &runtime.ArrayValue{
			ArrayType: types.NewDynamicArrayType(assoc.KeyType()),
			Elements:  keys,
		}, true
	case "length", "count":
		if len(args) != 0 {
			return e.newError(node, "%s expects no arguments, got %d", name, len(args)), true
		}
		return &runtime.IntegerValue{Value: int64(assoc.Len())}, true
	case "clear":
		if len(args) != 0 {
			return e.newError(node, "Clear expects no arguments, got %d", len(args)), true
		}
		assoc.Clear()
		return &runtime.NilValue{}, true
	case "delete":
		if len(args) != 1 {
			return e.newError(node, "Delete expects 1 argument, got %d", len(args)), true
		}
		removed := assoc.Delete(unwrapVariant(args[0]))
		return &runtime.BooleanValue{Value: removed}, true
	}
	return nil, false
}
