package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Nested lvalue resolution and associative-array vivification
// ============================================================================
//
// Reading `a[k]` on an associative array with a missing key yields the element
// type's zero value without inserting anything, which is correct for a pure
// rvalue read (`PrintLn(a['nope'])` must not grow the map). It is wrong when
// `a[k]` is only an intermediate step of a larger lvalue, because the caller
// then mutates a value that is not stored anywhere:
//
//	sa[1][1] := 123;          // writes into a throwaway static array
//	ra[2].S  := 'hello';      // writes into a throwaway record
//	a['x'].Add('y');          // mutates a throwaway dynamic array
//
// In those positions DWScript inserts the slot first (vivification) and then
// operates on the stored value. resolveLValueContainer implements that: it
// resolves the *container* part of a nested lvalue, taking the vivifying path
// through associative arrays and the ordinary read path through everything
// else.
// ============================================================================

// resolveLValueContainer evaluates expr as the container of a nested lvalue.
//
// For a plain expression this is just Eval: arrays, objects and records are
// already handed back as live references, so mutating the result is visible to
// the owner. For an IndexExpression the base is resolved recursively and a
// missing associative-array key is vivified, so that the value handed back is
// the one actually stored in the map.
func (e *Evaluator) resolveLValueContainer(expr ast.Expression, ctx *ExecutionContext) Value {
	idx, ok := expr.(*ast.IndexExpression)
	if !ok || idx.Left == nil || idx.Index == nil {
		return e.Eval(expr, ctx)
	}

	// Indexed properties and other member-rooted forms have their own handling
	// in VisitIndexExpression; leave them to it.
	if base, _ := CollectIndices(idx); isMemberRootedBase(base) {
		return e.Eval(expr, ctx)
	}

	container := e.resolveLValueContainer(idx.Left, ctx)
	if isError(container) {
		return container
	}
	if ctx.Exception() != nil {
		return &runtime.NilValue{}
	}

	assoc, ok := derefAssociative(container)
	if !ok {
		// Not an associative array: ordinary read semantics already yield a live
		// container for every mutable kind (array, object, record, JSON).
		return e.indexResolvedValue(container, idx, ctx)
	}

	indexVal := e.Eval(idx.Index, ctx)
	if isError(indexVal) {
		return indexVal
	}
	if ctx.Exception() != nil {
		return &runtime.NilValue{}
	}

	return e.vivifyAssociativeSlot(assoc, indexVal, ctx)
}

// vivifyAssociativeSlot returns the value stored at key, inserting the element
// type's zero value first when the key is absent. The returned value is always
// the one held by the map, so mutating it (a nested index/member write, or an
// in-place method such as Add) is visible through the map.
func (e *Evaluator) vivifyAssociativeSlot(
	assoc *runtime.AssociativeArrayValue,
	indexVal Value,
	ctx *ExecutionContext,
) Value {
	key := unwrapVariant(indexVal)
	if stored, present := assoc.Get(key); present {
		return stored
	}
	zero := e.getZeroValueForType(assoc.ElementType(), ctx)
	assoc.Set(key, zero)
	// Read back rather than returning `zero` directly, so the slot the caller
	// mutates is unambiguously the one the map holds.
	stored, _ := assoc.Get(key)
	return stored
}

// derefAssociative unwraps variants and var-parameter references and reports
// whether the result is an associative array.
func derefAssociative(v Value) (*runtime.AssociativeArrayValue, bool) {
	if ref, isRef := v.(ReferenceAccessor); isRef {
		deref, err := ref.Dereference()
		if err != nil {
			return nil, false
		}
		v = deref
	}
	assoc, ok := unwrapVariant(v).(*runtime.AssociativeArrayValue)
	return assoc, ok
}

// isMemberRootedBase reports whether an index chain is rooted at a member
// access or a class name, the forms that VisitIndexExpression routes to
// indexed-property handling rather than plain container indexing.
func isMemberRootedBase(base ast.Expression) bool {
	_, ok := base.(*ast.MemberAccessExpression)
	return ok
}
