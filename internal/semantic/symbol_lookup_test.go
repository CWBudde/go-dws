package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/token"
)

// Expectations replay pinned dwsSymbols.pas SortSymbols/AddSymbolDirect/FindLocal.
// They are algorithm tests, not claims of executing an upstream binary.
func TestDeclarationTableDuplicateIdentity(t *testing.T) {
	first, second, third := &Symbol{Name: "Hello"}, &Symbol{Name: "hello"}, &Symbol{Name: "HELLO"}
	table := declarationTable{}
	table.add(first)
	table.add(second)
	table.add(third)
	if got, _ := table.find("hello"); got != second {
		t.Fatal("unsorted midpoint pivot selected wrong identity")
	}
	fourth := &Symbol{Name: "Hello"}
	table.add(fourth)
	if got, _ := table.find("HELLO"); got != fourth {
		t.Fatal("equal-name insertion did not stop at midpoint")
	}
	fifth := &Symbol{Name: "hello"}
	table.add(fifth)
	if got, _ := table.find("Hello"); got != fourth {
		t.Fatal("lookup used first/last equal entry")
	}
	if len(table.entries) != 5 {
		t.Fatal("duplicates lost")
	}
}

func TestDeclarationTableOrderedParameters(t *testing.T) {
	first, second := &Symbol{Name: "Hello"}, &Symbol{Name: "hello"}
	table := declarationTable{ordered: true}
	table.add(first)
	table.add(second)
	if got, _ := table.find("HELLO"); got != first {
		t.Fatal("ordered table must retain first identity")
	}
}

func TestImportedParentInspectionMatchesResolution(t *testing.T) {
	first, second := NewSymbolTable(), NewSymbolTable()
	first.DefineConst("Shared", types.INTEGER, 1, token.Position{})
	second.DefineConst("Shared", types.INTEGER, 2, token.Position{})
	table := NewSymbolTable()
	table.importedParents = []*SymbolTable{second, first}
	selected, _ := table.Resolve("Shared")
	if got := table.AllSymbols()["shared"]; got != selected {
		t.Fatal("flattened inspection disagrees with parent lookup")
	}
	table.outer = first
	table.outerBeforeImports = true
	selected, _ = table.Resolve("Shared")
	if got := table.AllSymbols()["shared"]; got != selected {
		t.Fatal("own interface lost priority during inspection")
	}
}
