package types

import "testing"

func TestOperatorTypedIdentity(t *testing.T) {
	parent := NewClassType("TBase", nil)
	child := NewClassType("TChild", parent)
	registry := NewOperatorRegistry()
	base := &OperatorSignature{Operator: "+", OperandTypes: []Type{parent, INTEGER}, Binding: "base"}
	exact := &OperatorSignature{Operator: "+", OperandTypes: []Type{child, INTEGER}, Binding: "child"}
	for _, signature := range []*OperatorSignature{base, exact} {
		if err := registry.Register(signature); err != nil {
			t.Fatal(err)
		}
	}
	if got, ok := registry.Lookup("+", []Type{child, INTEGER}); !ok || got != exact {
		t.Fatal("exact overload must precede compatible overload")
	}
	if got, ok := registry.Lookup("+", []Type{NewClassType("tbase", nil), INTEGER}); !ok || got != base {
		t.Fatal("nominal type lookup must ignore identifier case")
	}
	if _, ok := registry.Lookup("+", []Type{NewClassType("Other", nil), INTEGER}); ok {
		t.Fatal("unrelated class matched")
	}
}

func TestOperatorCompatibleClassAndArray(t *testing.T) {
	parent := NewClassType("TBase", nil)
	child := NewClassType("TChild", NewClassType("tbase", nil))
	if !OperatorTypesCompatible(child, parent) {
		t.Fatal("case-insensitive ancestor did not match")
	}
	if OperatorTypesCompatible(parent, child) {
		t.Fatal("base matched derived operand")
	}
	if !OperatorTypesCompatible(NewStaticArrayType(INTEGER, -2, 2), NewDynamicArrayType(VARIANT)) {
		t.Fatal("static array must match array of const")
	}
	if OperatorTypesCompatible(nil, INTEGER) {
		t.Fatal("unresolved operand matched integer")
	}
}
