package types

import "testing"

func TestPropertyIndexModes_LegacyAndDeclared(t *testing.T) {
	var zero PropertyIndexMode
	if zero != PropertyIndexValue {
		t.Fatal("zero mode must preserve ordinary value semantics")
	}
	legacy := &PropertyInfo{}
	if legacy.IndexMode(0) != PropertyIndexValue {
		t.Fatal("legacy metadata must default to value")
	}
	prop := &PropertyInfo{IndexParamModes: []PropertyIndexMode{PropertyIndexVar, PropertyIndexConst, PropertyIndexValue}}
	for i, mode := range prop.IndexParamModes {
		if prop.IndexMode(i) != mode {
			t.Fatalf("mode %d changed", i)
		}
	}
	if prop.IndexMode(3) != PropertyIndexValue {
		t.Fatal("missing mode must default to value")
	}
}
