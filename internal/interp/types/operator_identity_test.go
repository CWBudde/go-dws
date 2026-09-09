package types

import (
	"testing"

	coretypes "github.com/cwbudde/go-dws/internal/types"
)

func TestConversionPathUsesTypedIdentityAndDeclarationOrder(t *testing.T) {
	first := coretypes.NewClassType("First", nil)
	second := coretypes.NewClassType("Second", nil)
	registry := NewConversionRegistry()
	for _, pair := range [][2]coretypes.Type{
		{coretypes.INTEGER, first}, {coretypes.INTEGER, second},
		{first, coretypes.INTEGER}, {first, coretypes.STRING}, {second, coretypes.STRING},
	} {
		if err := registry.Register(&ConversionEntry{From: pair[0], To: pair[1], Implicit: true}); err != nil {
			t.Fatal(err)
		}
	}
	path := registry.FindConversionPath(coretypes.INTEGER, coretypes.STRING, 2)
	if len(path) != 3 || path[1] != first {
		t.Fatalf("expected first declared shortest path, got %v", path)
	}
	if path := registry.FindConversionPath(coretypes.INTEGER, coretypes.STRING, 1); path != nil {
		t.Fatalf("depth limit ignored: %v", path)
	}
	if _, ok := registry.FindImplicit(coretypes.NewClassType("FIRST", nil), coretypes.STRING); !ok {
		t.Fatal("case-insensitive nominal lookup failed")
	}
	if err := registry.Register(&ConversionEntry{From: coretypes.NewClassType("first", nil), To: coretypes.STRING, Implicit: true}); err == nil {
		t.Fatal("duplicate conversion accepted")
	}
}

func TestRuntimeOperatorsShareTypedMatching(t *testing.T) {
	registry := NewOperatorRegistry()
	parent := coretypes.NewClassType("Parent", nil)
	child := coretypes.NewClassType("Child", parent)
	base := &OperatorEntry{Operator: "+", OperandTypes: []coretypes.Type{parent, coretypes.INTEGER}}
	exact := &OperatorEntry{Operator: "+", OperandTypes: []coretypes.Type{child, coretypes.INTEGER}}
	for _, entry := range []*OperatorEntry{base, exact} {
		if err := registry.Register(entry); err != nil {
			t.Fatal(err)
		}
	}
	if got, ok := registry.Lookup("+", []coretypes.Type{child, coretypes.INTEGER}); !ok || got != exact {
		t.Fatal("compatible entry shadowed exact entry")
	}
	grandchild := coretypes.NewClassType("Grandchild", child)
	if got, ok := registry.Lookup("+", []coretypes.Type{grandchild, coretypes.INTEGER}); !ok || got != exact {
		t.Fatal("compatible entries must prefer the nearest ancestor")
	}
}

func TestRuntimeOperatorsAncestorExactBeforeArrayCompatibility(t *testing.T) {
	registry := NewOperatorRegistry()
	parent := coretypes.NewClassType("Parent", nil)
	child := coretypes.NewClassType("Child", parent)
	integers := coretypes.NewDynamicArrayType(coretypes.INTEGER)
	broad := &OperatorEntry{Operator: "+", OperandTypes: []coretypes.Type{parent, coretypes.NewDynamicArrayType(coretypes.VARIANT)}}
	exactArray := &OperatorEntry{Operator: "+", OperandTypes: []coretypes.Type{parent, integers}}
	for _, entry := range []*OperatorEntry{broad, exactArray} {
		if err := registry.Register(entry); err != nil {
			t.Fatal(err)
		}
	}
	if got, ok := registry.Lookup("+", []coretypes.Type{child, integers}); !ok || got != exactArray {
		t.Fatal("compatible array shadowed exact signature at same ancestor distance")
	}
}
