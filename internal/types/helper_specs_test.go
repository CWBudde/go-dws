package types

import "testing"

func TestBuiltinHelperCatalog(t *testing.T) {
	for _, target := range []string{"array", "String", "Integer", "Float", "Boolean", "array of String", "enum"} {
		helper := NewBuiltinHelper(target)
		if helper == nil {
			t.Fatalf("missing builtin helper for %s", target)
		}
		for name, spec := range helper.BuiltinMethods {
			member, ok := LookupBuiltinHelper(target, name)
			if !ok || string(member.Operation) != spec {
				t.Errorf("%s.%s does not resolve to %s", target, name, spec)
			}
		}
	}
	for _, name := range []string{"Length", "COUNT", "hIgH", "low", "Contains", "Filter", "Sort", "Map"} {
		if _, ok := LookupBuiltinHelper("ARRAY", name); !ok {
			t.Errorf("missing case-insensitive array helper %s", name)
		}
	}
	if _, ok := LookupBuiltinHelper("array", "missing"); ok {
		t.Fatal("unknown member resolved")
	}
}

func TestBuiltinHelperCatalog_IsolatedRegistrations(t *testing.T) {
	first := NewBuiltinHelper("String")
	first.Properties["length"].ReadSpec = "changed"
	first.Methods["padleft"].Parameters[0] = BOOLEAN
	delete(first.BuiltinMethods, "trim")
	second := NewBuiltinHelper("String")
	if second.Properties["length"].ReadSpec != "__string_length" || second.Methods["padleft"].Parameters[0] != INTEGER || second.BuiltinMethods["trim"] == "" {
		t.Fatal("mutating one registration changed another")
	}
}
