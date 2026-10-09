package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/token"
)

func TestDefineOverload_SingletonForwardPreservedOnRejectedNewMember(t *testing.T) {
	st := NewSymbolTable()
	integer := types.NewFunctionType([]types.Type{types.INTEGER}, types.VOID)
	if err := st.DefineOverload("P", integer, true, true, token.Position{Line: 1, Column: 11}); err != nil {
		t.Fatal(err)
	}
	stringSignature := types.NewFunctionType([]types.Type{types.STRING}, types.VOID)
	if err := st.DefineOverload("p", stringSignature, false, false, token.Position{Line: 2, Column: 11}); err == nil || err.Error() != "Overloaded procedure \"p\" must be marked with the \"overload\" directive" {
		t.Fatalf("new member without overload: %v", err)
	}
	forwards := st.UnimplementedForwards()
	if len(forwards) != 1 || !forwards[0].IsForward || forwards[0].Type != integer {
		t.Fatalf("rejected member consumed or replaced pending forward: %+v", forwards)
	}
	if err := st.DefineOverload("P", integer, false, false, token.Position{Line: 3, Column: 11}); err != nil {
		t.Fatalf("matching implementation without repeated overload: %v", err)
	}
	if forwards := st.UnimplementedForwards(); len(forwards) != 0 {
		t.Fatalf("selected forward remained pending: %+v", forwards)
	}
}

func TestDefineOverload_SingletonForwardOnlySelectedCandidateCleared(t *testing.T) {
	st := NewSymbolTable()
	integer := types.NewFunctionType([]types.Type{types.INTEGER}, types.VOID)
	stringSignature := types.NewFunctionType([]types.Type{types.STRING}, types.VOID)
	boolean := types.NewFunctionType([]types.Type{types.BOOLEAN}, types.VOID)
	if err := st.DefineOverload("P", integer, true, true, token.Position{Line: 1, Column: 11}); err != nil {
		t.Fatal(err)
	}
	if err := st.DefineOverload("P", stringSignature, true, false, token.Position{Line: 2, Column: 11}); err != nil {
		t.Fatalf("new differing-type implementation: %v", err)
	}
	if forwards := st.UnimplementedForwards(); len(forwards) != 1 || forwards[0].Type != integer {
		t.Fatalf("new member changed pending original: %+v", forwards)
	}
	if err := st.DefineOverload("P", boolean, true, true, token.Position{Line: 3, Column: 11}); err != nil {
		t.Fatal(err)
	}
	if err := st.DefineOverload("p", integer, false, false, token.Position{Line: 4, Column: 11}); err != nil {
		t.Fatalf("matching original implementation: %v", err)
	}
	symbol, ok := st.Resolve("P")
	if !ok || !symbol.IsOverloadSet || len(symbol.Overloads) != 3 {
		t.Fatalf("overload candidates: %+v", symbol)
	}
	if symbol.Overloads[0].IsForward || symbol.Overloads[1].IsForward || !symbol.Overloads[2].IsForward {
		t.Fatalf("only selected Integer forward should clear: %+v", symbol.Overloads)
	}
	if forwards := st.UnimplementedForwards(); len(forwards) != 1 || forwards[0].Type != boolean {
		t.Fatalf("unselected Boolean forward changed: %+v", forwards)
	}
}
