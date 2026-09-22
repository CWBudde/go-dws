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

// valueToJSONValue snapshots a runtime Value for JSON text serialization with access to
// the evaluator and execution context, so class/record serialization can run
// property getters and a custom Stringify override. It is the context-aware
// counterpart of the package-level ValueToJSONValue and is used by the
// JSON.Stringify/Serialize/PrettyStringify handlers.
func (e *Evaluator) valueToJSONValue(val Value, node ast.Node, ctx *ExecutionContext) *jsonTextValue {
	if val == nil {
		return jsonTextFromValue(jsonvalue.NewNull())
	}

	// Unwrap Variant so a boxed object/record/array is serialized structurally.
	if wrapper, ok := val.(runtime.VariantWrapper); ok {
		unwrapped := wrapper.UnwrapVariant()
		if unwrapped == nil {
			return jsonTextFromValue(jsonvalue.NewNull())
		}
		val = unwrapped
	}

	switch v := val.(type) {
	case *runtime.JSONValue:
		// Snapshot live JSON without adopting or moving its children.
		// A nil node represents JSON undefined, which serializes as null.
		if v == nil || v.Value == nil {
			return jsonTextFromValue(jsonvalue.NewNull())
		}
		return jsonTextFromValue(v.Value)
	case *runtime.AssociativeArrayValue:
		return e.associativeArrayToJSON(v, node, ctx)
	case *runtime.ArrayValue:
		arr := newJSONTextArray()
		for _, elem := range v.Elements {
			arr.elements = append(arr.elements, e.valueToJSONValue(elem, node, ctx))
			if ctx.Exception() != nil {
				break
			}
		}
		return arr
	case *runtime.RecordValue:
		return e.recordToJSON(v, node, ctx)
	case *runtime.ObjectInstance:
		return e.objectToJSON(v, node, ctx)
	case *runtime.SetValue:
		return jsonTextFromValue(setToJSON(v))
	default:
		// Primitives, JSON passthrough, and nil are handled by the
		// context-free converter.
		return jsonTextFromValue(ValueToJSONValue(val))
	}
}

// associativeArrayToJSON follows JSONScript.StringifyAssociativeArray: base
// scalar keys become object names in bucket order, while other key types
// produce an empty object. Values use the contextual converter so nested
// objects retain property getter and custom Stringify behavior.
func (e *Evaluator) associativeArrayToJSON(assoc *runtime.AssociativeArrayValue, node ast.Node, ctx *ExecutionContext) *jsonTextValue {
	result := newJSONTextObject()
	if assoc == nil || assoc.KeyType() == nil || ctx.Exception() != nil {
		return result
	}
	keyType := types.GetUnderlyingType(assoc.KeyType())
	if !types.IsBasicType(keyType) && keyType.TypeKind() != "VARIANT" {
		return result
	}
	for _, key := range assoc.Keys() {
		value, _ := assoc.Get(key)
		jv := e.valueToJSONValue(value, node, ctx)
		if ctx.Exception() != nil {
			return result
		}
		result.appendMember(key.String(), jv)
	}
	return result
}

// jsonMember retains one textual name and serialized value; names may repeat.
type jsonMember struct {
	jv   *jsonTextValue
	name string
}

// objectToJSON serializes a class instance to a JSON object, matching DWScript's
// JSON connector: public members only (private/protected excluded), fields and
// non-indexed readable properties, ordered most-derived class first and
// ordinally within each class level. A custom parameterless String-returning
// Stringify method, if present, replaces the composite serialization.
func (e *Evaluator) objectToJSON(obj *runtime.ObjectInstance, node ast.Node, ctx *ExecutionContext) *jsonTextValue {
	if obj == nil || obj.Class == nil || obj.Destroyed {
		return jsonTextFromValue(jsonvalue.NewNull())
	}

	if jv, ok := e.objectCustomStringify(obj, node, ctx); ok {
		return jv
	}

	result := newJSONTextObject()
	if ctx.Exception() != nil {
		return result
	}
	seen := make(map[string]bool)

	for cur := obj.Class; cur != nil; cur = cur.GetParent() {
		var members []jsonMember
		levelSeen := make(map[string]bool)

		add := func(name string, jv *jsonTextValue) {
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
				if ctx.Exception() != nil {
					return result
				}
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
			if ctx.Exception() != nil {
				return result
			}
			if isError(res) {
				continue
			}
			// An `external 'name'` clause renames the property's JSON key.
			name := prop.Name
			if pInfo.ExternalName != "" {
				name = pInfo.ExternalName
			}
			add(name, e.valueToJSONValue(res, node, ctx))
			if ctx.Exception() != nil {
				return result
			}
		}

		sort.Slice(members, func(i, j int) bool { return members[i].name < members[j].name })
		for _, m := range members {
			seen[ident.Normalize(m.name)] = true
			result.appendMember(m.name, m.jv)
		}
	}

	return result
}

// setToJSON serializes a set value to a JSON array, matching DWScript's JSON
// connector: members are emitted in ascending ordinal order; a member with an
// enum name becomes a JSON string (qualified as "TEnum.member" for a scoped
// `enum`/`flags` type, bare for a plain `(...)` enum), and an ordinal with no
// name becomes a JSON number.
func setToJSON(s *runtime.SetValue) *jsonvalue.Value {
	arr := jsonvalue.NewArray()
	if s == nil || s.SetType == nil {
		return arr
	}
	// Resolve through any type alias so `set of <alias-to-enum>` still emits
	// enum member names rather than falling back to numeric ordinals.
	enumType, _ := types.GetUnderlyingType(s.SetType.ElementType).(*types.EnumType)
	for _, ordinal := range s.Ordinals() {
		var name string
		if enumType != nil {
			name = enumType.GetEnumName(ordinal)
		}
		switch {
		case name == "":
			arr.ArrayAppend(jsonvalue.NewInt64(int64(ordinal)))
		case enumType.Scoped:
			arr.ArrayAppend(jsonvalue.NewString(enumType.Name + "." + name))
		default:
			arr.ArrayAppend(jsonvalue.NewString(name))
		}
	}
	return arr
}

// objectCustomStringify returns the spliced result of a class's custom Stringify
// override, if it declares a parameterless, non-class, non-constructor method
// named Stringify returning String. The returned string is treated as raw JSON
// (re-parsed); on a parse failure it falls back to a JSON string.
func (e *Evaluator) objectCustomStringify(obj *runtime.ObjectInstance, node ast.Node, ctx *ExecutionContext) (*jsonTextValue, bool) {
	md := obj.Class.LookupMethod("Stringify")
	if md == nil {
		return nil, false
	}
	if len(md.Parameters) != 0 || md.IsClassMethod || md.IsConstructor || md.ReturnType == nil {
		return nil, false
	}
	if types.GetUnderlyingType(md.ReturnType) != types.STRING {
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

	if jv, err := parseJSONText(strings.TrimSpace(s)); err == nil {
		return jv, true
	}
	return jsonTextFromValue(jsonvalue.NewString(s)), true
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
func (e *Evaluator) recordToJSON(rec *runtime.RecordValue, node ast.Node, ctx *ExecutionContext) *jsonTextValue {
	result := newJSONTextObject()

	var members []jsonMember
	seen := make(map[string]bool)

	add := func(name string, jv *jsonTextValue) {
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
			if ctx.Exception() != nil {
				return result
			}
			if isError(res) {
				continue
			}
			// An `external 'name'` clause renames the property's JSON key.
			name := prop.Name
			if prop.ExternalName != "" {
				name = prop.ExternalName
			}
			add(name, e.valueToJSONValue(res, node, ctx))
			if ctx.Exception() != nil {
				return result
			}
		}
	}

	// Public fields. rec.Fields keys are normalized; recover the original
	// declared casing from the record type's FieldNames map.
	for fieldKey, fieldValue := range rec.Fields {
		if !recordFieldIsPublic(rec.RecordType, fieldKey) {
			continue
		}
		name := recordJSONFieldName(rec.RecordType, fieldKey)
		add(name, e.valueToJSONValue(fieldValue, node, ctx))
		if ctx.Exception() != nil {
			return result
		}
	}

	sort.Slice(members, func(i, j int) bool { return members[i].name < members[j].name })
	for _, m := range members {
		result.appendMember(m.name, m.jv)
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

// recordJSONFieldName recovers declaration spelling from normalized storage keys.
func recordJSONFieldName(recordType *types.RecordType, key string) string {
	if recordType != nil {
		if name, ok := recordType.FieldNames[ident.Normalize(key)]; ok {
			return name
		}
	}
	return key
}
