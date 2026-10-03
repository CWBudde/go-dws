package types

import "testing"

func TestRecordMeta_NominalCompatibility(t *testing.T) {
	record := NewRecordType("R", map[string]Type{"x": INTEGER})
	sameShape := NewRecordType("R", map[string]Type{"x": INTEGER})
	meta := NewRecordMetaType(record)
	alias := NewRecordMetaType(&TypeAlias{Name: "A", AliasedType: record})
	if !meta.Equals(alias) || !alias.Equals(meta) {
		t.Fatal("aliases must preserve record identity")
	}
	if meta.Equals(NewRecordMetaType(sameShape)) {
		t.Fatal("distinct declarations must remain distinct")
	}
	if meta.Equals(record) || record.Equals(meta) {
		t.Fatal("instance and meta must differ")
	}
	if GetUnderlyingType(meta) != meta {
		t.Fatal("meta must survive ordinary alias unwrapping")
	}
}
