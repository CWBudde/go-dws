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

	// An index chain rooted at a member (`holder.X[k]`) may be an indexed
	// property, which VisitIndexExpression handles as a whole because it needs
	// the flattened index list. Ordinary members — a field or a non-indexed
	// property holding an associative array — must still take the vivifying
	// path, so only genuine indexed properties are handed over.
	if base, _ := CollectIndices(idx); isMemberRootedBase(base) {
		if ma, isMember := base.(*ast.MemberAccessExpression); isMember && e.memberIsIndexedProperty(ma, ctx) {
			return e.Eval(expr, ctx)
		}
	}

	return e.resolveIndexedLValueContainer(idx, ctx)
}

// resolveIndexedLValueContainer resolves one level of an index chain that
// resolveLValueContainer has already cleared for vivification.
func (e *Evaluator) resolveIndexedLValueContainer(idx *ast.IndexExpression, ctx *ExecutionContext) Value {
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

// memberIsIndexedProperty reports whether `obj.Member` names something that
// VisitIndexExpression reads as an indexed property rather than as a plain
// container that happens to be indexed afterwards. Only those forms must skip
// vivification; an ordinary field or a non-indexed property that holds an
// associative array is resolved through the vivifying path like any other
// container.
//
// The object expression is evaluated here and again by the indexed-property
// path, mirroring what evalIndexAssignmentDirect already does for the write
// side of the same forms.
func (e *Evaluator) memberIsIndexedProperty(ma *ast.MemberAccessExpression, ctx *ExecutionContext) bool {
	if ma.Object == nil || ma.Member == nil {
		return true
	}
	objVal := e.Eval(ma.Object, ctx)
	if isError(objVal) || ctx.Exception() != nil {
		// Let the ordinary path report the failure.
		return true
	}
	// An indexed property reached through a class name has its own handling.
	if _, isMeta := objVal.(ClassMetaValue); isMeta {
		return true
	}
	propDesc := lookupPropertyDescriptor(objVal, ma.Member.Value)
	if propDesc == nil {
		if intf, isIntf := objVal.(InterfaceInstanceValue); isIntf {
			if underlying := intf.GetUnderlyingObjectValue(); underlying != nil {
				objVal = underlying
				propDesc = lookupPropertyDescriptor(objVal, ma.Member.Value)
			}
		}
	}
	if propDesc == nil {
		return false
	}
	// Records route every indexed property read through ReadIndexedProperty,
	// not just the ones declared with index parameters.
	if _, isRecord := objVal.(RecordInstanceValue); isRecord {
		return true
	}
	return propDesc.IsIndexed
}

// lookupPropertyDescriptor returns the property named name on v, or nil when v
// exposes no properties or has no such property.
func lookupPropertyDescriptor(v Value, name string) *PropertyDescriptor {
	accessor, ok := v.(PropertyAccessor)
	if !ok {
		return nil
	}
	return accessor.LookupProperty(name)
}

// isMemberRootedBase reports whether an index chain is rooted at a member
// access or a class name, the forms that VisitIndexExpression routes to
// indexed-property handling rather than plain container indexing.
func isMemberRootedBase(base ast.Expression) bool {
	_, ok := base.(*ast.MemberAccessExpression)
	return ok
}
