package evaluator

import (
	"sort"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/jsonvalue"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// valueToJSONValue converts a runtime Value to a jsonvalue.Value with access to
// the evaluator and execution context, so class/record serialization can run
// property getters and a custom Stringify override. It is the context-aware
// counterpart of the package-level ValueToJSONValue and is used by the
// JSON.Stringify/Serialize/PrettyStringify handlers.
func (e *Evaluator) valueToJSONValue(val Value, node ast.Node, ctx *ExecutionContext) *jsonvalue.Value {
	if val == nil {
		return jsonvalue.NewNull()
	}

	// Unwrap Variant so a boxed object/record/array is serialized structurally.
	if wrapper, ok := val.(runtime.VariantWrapper); ok {
		unwrapped := wrapper.UnwrapVariant()
		if unwrapped == nil {
			return jsonvalue.NewNull()
		}
		val = unwrapped
	}

	switch v := val.(type) {
	case *runtime.ArrayValue:
		arr := jsonvalue.NewArray()
		for _, elem := range v.Elements {
			arr.ArrayAppend(e.valueToJSONValue(elem, node, ctx))
		}
		return arr
	case *runtime.RecordValue:
		return e.recordToJSON(v, node, ctx)
	case *runtime.ObjectInstance:
		return e.objectToJSON(v, node, ctx)
	default:
		// Primitives, JSON passthrough, and nil are handled by the
		// context-free converter.
		return ValueToJSONValue(val)
	}
}

// jsonMember is a collected (name, value) pair pending ordinal sorting.
type jsonMember struct {
	jv   *jsonvalue.Value
	name string
}

// objectToJSON serializes a class instance to a JSON object, matching DWScript's
// JSON connector: public members only (private/protected excluded), fields and
// non-indexed readable properties, ordered most-derived class first and
// ordinally within each class level. A custom parameterless String-returning
// Stringify method, if present, replaces the composite serialization.
func (e *Evaluator) objectToJSON(obj *runtime.ObjectInstance, node ast.Node, ctx *ExecutionContext) *jsonvalue.Value {
	if obj == nil || obj.Class == nil || obj.Destroyed {
		return jsonvalue.NewNull()
	}

	if jv, ok := e.objectCustomStringify(obj, node, ctx); ok {
		return jv
	}

	result := jsonvalue.NewObject()
	seen := make(map[string]bool)

	for cur := obj.Class; cur != nil; cur = cur.GetParent() {
		var members []jsonMember
		levelSeen := make(map[string]bool)

		add := func(name string, jv *jsonvalue.Value) {
			norm := ident.Normalize(name)
			if seen[norm] || levelSeen[norm] {
				return
			}
			levelSeen[norm] = true
			members = append(members, jsonMember{name: name, jv: jv})
		}

		// Own fields at this class level (public only).
		if meta := cur.GetMetadata(); meta != nil {
			for _, fm := range meta.Fields {
				if fm.Visibility != runtime.FieldVisibilityPublic {
					continue
				}
				fv := obj.GetFieldFromClass(fm.Name, cur.GetName())
				add(fm.Name, e.valueToJSONValue(fv, node, ctx))
			}
		}

		// Own properties at this class level (non-indexed, readable).
		for _, prop := range ownProperties(cur) {
			if prop.IsIndexed {
				continue
			}
			pInfo, ok := unwrapPropertyInfo(prop)
			if !ok || pInfo.ReadKind == types.PropAccessNone {
				continue
			}
			res := e.executePropertyRead(obj, prop, node, ctx)
			if isError(res) {
				continue
			}
			add(prop.Name, e.valueToJSONValue(res, node, ctx))
		}

		sort.Slice(members, func(i, j int) bool { return members[i].name < members[j].name })
		for _, m := range members {
			seen[ident.Normalize(m.name)] = true
			result.ObjectSet(m.name, m.jv)
		}
	}

	return result
}

// objectCustomStringify returns the spliced result of a class's custom Stringify
// override, if it declares a parameterless, non-class, non-constructor method
// named Stringify returning String. The returned string is treated as raw JSON
// (re-parsed); on a parse failure it falls back to a JSON string.
func (e *Evaluator) objectCustomStringify(obj *runtime.ObjectInstance, node ast.Node, ctx *ExecutionContext) (*jsonvalue.Value, bool) {
	md := obj.Class.LookupMethod("Stringify")
	if md == nil {
		return nil, false
	}
	if len(md.Parameters) != 0 || md.IsClassMethod || md.IsConstructor || md.ReturnType == nil {
		return nil, false
	}
	if !ident.Equal(md.ReturnType.String(), "String") {
		return nil, false
	}

	res := e.executeObjectMethodDirect(obj, md, nil, node, ctx)
	if isError(res) {
		return nil, false
	}

	s, ok := jsonResultString(res)
	if !ok {
		return nil, false
	}

	if jv, err := jsonvalue.Parse(strings.TrimSpace(s)); err == nil {
		return jv, true
	}
	return jsonvalue.NewString(s), true
}

// jsonResultString extracts a Go string from a runtime string value, unwrapping
// a Variant if necessary.
func jsonResultString(val Value) (string, bool) {
	if wrapper, ok := val.(runtime.VariantWrapper); ok {
		if unwrapped := wrapper.UnwrapVariant(); unwrapped != nil {
			val = unwrapped
		}
	}
	if sv, ok := val.(*runtime.StringValue); ok {
		return sv.Value, true
	}
	return "", false
}

// recordToJSON serializes a record value to a JSON object, applying the same
// visibility, ordering, and getter-execution rules as class serialization.
func (e *Evaluator) recordToJSON(rec *runtime.RecordValue, node ast.Node, ctx *ExecutionContext) *jsonvalue.Value {
	result := jsonvalue.NewObject()

	var members []jsonMember
	seen := make(map[string]bool)

	add := func(name string, jv *jsonvalue.Value) {
		norm := ident.Normalize(name)
		if seen[norm] {
			return
		}
		seen[norm] = true
		members = append(members, jsonMember{name: name, jv: jv})
	}

	// Non-indexed readable properties.
	if rec.RecordType != nil {
		for _, prop := range rec.RecordType.Properties {
			if prop.IsIndexed || prop.ReadKind == types.PropAccessNone {
				continue
			}
			res := e.executeRecordPropertyRead(rec, prop, node, ctx)
			if isError(res) {
				continue
			}
			add(prop.Name, e.valueToJSONValue(res, node, ctx))
		}
	}

	// Public fields. rec.Fields keys are normalized; recover the original
	// declared casing from the record type's FieldNames map.
	for fieldKey, fieldValue := range rec.Fields {
		if !recordFieldIsPublic(rec.RecordType, fieldKey) {
			continue
		}
		name := fieldKey
		if rec.RecordType != nil {
			if orig, ok := rec.RecordType.FieldNames[ident.Normalize(fieldKey)]; ok {
				name = orig
			}
		}
		add(name, e.valueToJSONValue(fieldValue, node, ctx))
	}

	sort.Slice(members, func(i, j int) bool { return members[i].name < members[j].name })
	for _, m := range members {
		result.ObjectSet(m.name, m.jv)
	}

	return result
}

// recordFieldIsPublic reports whether a record field should be serialized. A
// field with no recorded visibility (e.g. anonymous records) is treated as
// public; a field explicitly declared private or protected is excluded.
func recordFieldIsPublic(rt *types.RecordType, fieldName string) bool {
	if rt == nil || rt.FieldVisibility == nil {
		return true
	}
	vis, ok := rt.FieldVisibility[ident.Normalize(fieldName)]
	if !ok {
		return true
	}
	return vis == int(ast.VisibilityPublic)
}

// ownPropertyLister is implemented by the concrete *interp.ClassInfo to expose
// the properties declared directly on a class level (not inherited). Asserted
// structurally on runtime.IClassInfo to avoid an import cycle.
type ownPropertyLister interface {
	GetOwnProperties() []*runtime.PropertyInfo
}

// ownProperties returns the properties declared directly on a class level, or
// nil if the class info does not support per-level enumeration.
func ownProperties(ci runtime.IClassInfo) []*runtime.PropertyInfo {
	if l, ok := ci.(ownPropertyLister); ok {
		return l.GetOwnProperties()
	}
	return nil
}
