package runtime

import (
	"testing"
)

// metaClassInfo is a metadata-backed IClassInfo for field-shadowing tests.
type metaClassInfo struct {
	mockClassInfo
	meta   *ClassMetadata
	parent *metaClassInfo
}

func (m *metaClassInfo) GetMetadata() *ClassMetadata { return m.meta }
func (m *metaClassInfo) GetParent() IClassInfo {
	if m.parent == nil {
		return nil
	}
	return m.parent
}

func newMetaClass(name string, parent *metaClassInfo, fieldNames ...string) *metaClassInfo {
	meta := NewClassMetadata(name)
	if parent != nil {
		meta.Parent = parent.meta
		meta.ParentName = parent.meta.Name
	}
	for _, fieldName := range fieldNames {
		AddFieldToClass(meta, &FieldMetadata{
			Name:       fieldName,
			TypeName:   "Integer",
			Visibility: FieldVisibilityPublic,
		})
	}
	return &metaClassInfo{
		mockClassInfo: mockClassInfo{name: name},
		meta:          meta,
		parent:        parent,
	}
}

func intVal(v int64) Value { return &IntegerValue{Value: v} }

// TestFieldShadowing_PerClassSlots verifies that when a subclass redeclares a
// parent field with the same name, both storage slots exist and are selected
// by the static class of the reference (DWScript field shadowing semantics).
func TestFieldShadowing_PerClassSlots(t *testing.T) {
	base := newMetaClass("TBase", nil, "Field")
	child := newMetaClass("TChild", base, "Field")

	obj := NewObjectInstance(child)

	// Write through each static class: distinct slots.
	obj.SetFieldFromClass("Field", intVal(1), "TBase")
	obj.SetFieldFromClass("Field", intVal(2), "TChild")

	tests := []struct {
		name       string
		className  string
		wantResult int64
	}{
		{"read via TBase static class", "TBase", 1},
		{"read via TChild static class", "TChild", 2},
		{"read via dynamic class (empty)", "", 2},
		{"read via unknown class falls back to dynamic", "TOther", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := obj.GetFieldFromClass("Field", tt.className)
			iv, ok := got.(*IntegerValue)
			if !ok {
				t.Fatalf("expected IntegerValue, got %T (%v)", got, got)
			}
			if iv.Value != tt.wantResult {
				t.Errorf("got %d, want %d", iv.Value, tt.wantResult)
			}
		})
	}
}

// TestFieldShadowing_CaseInsensitive verifies shadowed slots resolve with
// case-insensitive class and field names.
func TestFieldShadowing_CaseInsensitive(t *testing.T) {
	base := newMetaClass("TBase", nil, "Field")
	child := newMetaClass("TChild", base, "Field")

	obj := NewObjectInstance(child)
	obj.SetFieldFromClass("FIELD", intVal(10), "tbase")

	got := obj.GetFieldFromClass("field", "TBASE")
	iv, ok := got.(*IntegerValue)
	if !ok || iv.Value != 10 {
		t.Fatalf("case-insensitive shadowed field read failed, got %v", got)
	}
	// The child slot is untouched (unset), so a dynamic read finds nothing.
	if got := obj.GetField("Field"); got != nil {
		t.Fatalf("expected unset child slot, got %v", got)
	}
}

// TestFieldNoShadowing_SingleSlot verifies that without redeclaration a field
// uses a single slot regardless of the lookup class (the common case).
func TestFieldNoShadowing_SingleSlot(t *testing.T) {
	base := newMetaClass("TBase", nil, "Field")
	child := newMetaClass("TChild", base) // no redeclaration

	obj := NewObjectInstance(child)
	obj.SetFieldFromClass("Field", intVal(5), "TChild")

	for _, className := range []string{"", "TBase", "TChild"} {
		got := obj.GetFieldFromClass("Field", className)
		iv, ok := got.(*IntegerValue)
		if !ok || iv.Value != 5 {
			t.Fatalf("lookup class %q: expected 5, got %v", className, got)
		}
	}

	// Plain SetField/GetField keep working on the same slot.
	obj.SetField("Field", intVal(6))
	if got := obj.GetFieldFromClass("Field", "TBase"); got.(*IntegerValue).Value != 6 {
		t.Fatalf("expected 6 after plain SetField, got %v", got)
	}
	// The slot key is the plain normalized name (legacy layout).
	if _, ok := obj.Fields["field"]; !ok {
		t.Fatalf("expected plain normalized key 'field' in Fields map, keys: %v", mapKeys(obj.Fields))
	}
}

// TestFieldShadowing_ThreeLevels verifies slots across a three-level chain
// where only the outer classes declare the field.
func TestFieldShadowing_ThreeLevels(t *testing.T) {
	base := newMetaClass("TBase", nil, "Field")
	middle := newMetaClass("TMiddle", base) // inherits, no redeclaration
	child := newMetaClass("TChild", middle, "Field")

	obj := NewObjectInstance(child)
	obj.SetFieldFromClass("Field", intVal(1), "TBase")
	obj.SetFieldFromClass("Field", intVal(3), "TChild")

	// TMiddle does not declare the field: lookup walks up to TBase's slot.
	got := obj.GetFieldFromClass("Field", "TMiddle")
	if iv, ok := got.(*IntegerValue); !ok || iv.Value != 1 {
		t.Fatalf("TMiddle lookup should reach TBase slot (1), got %v", got)
	}
	got = obj.GetFieldFromClass("Field", "")
	if iv, ok := got.(*IntegerValue); !ok || iv.Value != 3 {
		t.Fatalf("dynamic lookup should reach TChild slot (3), got %v", got)
	}
}

func mapKeys(m map[string]Value) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
