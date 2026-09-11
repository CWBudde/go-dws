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
	ctx *ExecutionContext,
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
		keys, values := assoc.TakeEntries()
		for i, v := range values {
			e.releaseAssociativeEntry(keys[i], v)
		}
		return &runtime.NilValue{}, true
	case "delete":
		if len(args) != 1 {
			return e.newError(node, "Delete expects 1 argument, got %d", len(args)), true
		}
		key, errVal := e.coerceAssociativeKey(assoc, args[0], ctx)
		if errVal != nil {
			return errVal, true
		}
		storedKey, storedValue, removed := assoc.DeleteEntry(key)
		if removed {
			e.releaseAssociativeEntry(storedKey, storedValue)
		}
		return &runtime.BooleanValue{Value: removed}, true
	}
	return nil, false
}

// coerceAssociativeKey unwraps a Variant index and converts it to the array's
// declared key type. A Variant holding an Integer used against an
// `array [String] of ...` must be stored (and looked up) as the String '123',
// otherwise a['123'] would miss the slot written through the Variant.
//
// Returns (key, nil) on success. A failed variant cast raises a catchable
// exception on ctx and returns (nil, errorValue).
func (e *Evaluator) coerceAssociativeKey(
	assoc *runtime.AssociativeArrayValue,
	indexVal Value,
	ctx *ExecutionContext,
) (Value, Value) {
	key := unwrapVariant(indexVal)
	keyType := assoc.KeyType()
	if key == nil || keyType == nil {
		return key, nil
	}

	// Only Variant-held indices are coerced; a genuinely mistyped key keeps its
	// strict behavior.
	if _, wasVariant := indexVal.(runtime.VariantWrapper); !wasVariant {
		return key, nil
	}

	if types.OperatorTypesEqual(runtime.LanguageType(key), keyType) {
		return key, nil
	}

	if converted, ok := e.TryImplicitConversion(key, keyType, ctx); ok {
		return converted, nil
	}

	converted, errVal := e.coerceValueToKind(key, keyType.TypeKind(), nil, ctx)
	if errVal != nil {
		return nil, errVal
	}
	if converted != nil {
		return converted, nil
	}
	return key, nil
}

// storeAssociativeEntry writes value at key, keeping the ARC bookkeeping of the
// map's slots correct: the newly stored key and value are retained and the
// value displaced by an overwrite is released (running its destructor when it
// held the last reference).
func (e *Evaluator) storeAssociativeEntry(
	assoc *runtime.AssociativeArrayValue,
	key Value,
	value Value,
	ctx *ExecutionContext,
) {
	e.retainValueForBinding(value, ctx)
	prev, replaced := assoc.Set(key, value)
	if replaced {
		// The key slot already holds a retained copy of the key; only the
		// displaced value loses a reference.
		e.releaseValueForBinding(prev)
		return
	}
	e.retainValueForBinding(key, ctx)
}

// releaseAssociativeEntry drops the map's references to one removed entry.
// The value is released before the key, matching DWScript's destructor order
// for Delete and Clear.
func (e *Evaluator) releaseAssociativeEntry(key, value Value) {
	e.releaseValueForBinding(value)
	e.releaseValueForBinding(key)
}

// releaseAssociativeContents empties an associative array and drops the map's
// references to every entry. Keys are released before values, matching
// DWScript's destructor order when the owning binding dies (as opposed to the
// value-first order of an explicit Delete/Clear).
func (e *Evaluator) releaseAssociativeContents(assoc *runtime.AssociativeArrayValue) {
	keys, values := assoc.TakeEntries()
	for _, k := range keys {
		e.releaseValueForBinding(k)
	}
	for _, v := range values {
		e.releaseValueForBinding(v)
	}
}

// releaseAssociativeBindings runs the program-scope finalization for
// associative arrays: every map still bound in env drops the references it
// holds, so objects kept alive only by a map slot get their destructor at
// program end. Plain object bindings are deliberately untouched here — global
// object finalization is a separate concern.
func (e *Evaluator) releaseAssociativeBindings(env *runtime.Environment) {
	if env == nil || e.engineState == nil || e.engineState.RefCountManager == nil {
		return
	}
	env.Range(func(_ string, value Value) bool {
		if assoc, ok := value.(*runtime.AssociativeArrayValue); ok {
			e.releaseAssociativeContents(assoc)
		}
		return true
	})
}
