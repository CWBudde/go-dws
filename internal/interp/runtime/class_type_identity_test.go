package runtime

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
)

func TestClassInfo_ResolvedTypeOwnership(t *testing.T) {
	parent := NewClassInfo("TBase")
	unchecked := NewClassInfo("TUnchecked")
	unchecked.SetParentClass(parent)
	if unchecked.GetClassType().Parent != parent.GetClassType() {
		t.Fatal("unchecked class identity lost its parent")
	}

	shared := types.NewClassType("TChecked", nil)
	checked := NewClassInfo("TChecked")
	checked.SetResolvedType(shared)
	checked.SetParentClass(parent)
	if checked.GetClassType() != shared {
		t.Fatal("reconstructed semantic class identity")
	}
	if shared.Parent != nil {
		t.Fatal("runtime binding mutated shared semantic metadata")
	}
}
