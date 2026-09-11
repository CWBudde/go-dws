package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/token"
)

func TestSymbolTable_Depth(t *testing.T) {
	global := NewSymbolTable()
	if got := global.Depth(); got != 0 {
		t.Errorf("global scope depth = %d, want 0", got)
	}

	inner := NewEnclosedSymbolTable(global)
	if got := inner.Depth(); got != 1 {
		t.Errorf("inner scope depth = %d, want 1", got)
	}

	nested := NewEnclosedSymbolTable(inner)
	if got := nested.Depth(); got != 2 {
		t.Errorf("nested scope depth = %d, want 2", got)
	}
}

func TestSymbolTable_RetainPropagatesToChildren(t *testing.T) {
	global := NewSymbolTable()

	// Scopes are not retained by default.
	throwaway := NewEnclosedSymbolTable(global)
	if len(global.Children()) != 0 {
		t.Errorf("global should not collect children of unretained scopes, got %d", len(global.Children()))
	}
	if throwaway.ScopeName() != "" {
		t.Errorf("unretained scope name = %q, want empty", throwaway.ScopeName())
	}

	body := NewEnclosedSymbolTable(global)
	body.Retain("Compute")
	if body.ScopeName() != "Compute" {
		t.Errorf("scope name = %q, want %q", body.ScopeName(), "Compute")
	}

	block := NewEnclosedSymbolTable(body)
	if len(body.Children()) != 1 || body.Children()[0] != block {
		t.Fatalf("retained scope should collect its child scopes, got %d children", len(body.Children()))
	}
	if block.ScopeName() != "Compute" {
		t.Errorf("nested scope name = %q, want inherited %q", block.ScopeName(), "Compute")
	}
	if block.Depth() != 2 {
		t.Errorf("nested scope depth = %d, want 2", block.Depth())
	}
}

func TestSymbolTable_AllSymbolsWithScope_KeepsShadowedSymbols(t *testing.T) {
	global := NewSymbolTable()
	global.Define("shared", types.INTEGER, token.Position{Line: 1, Column: 5})
	global.Define("onlyGlobal", types.STRING, token.Position{Line: 2, Column: 5})

	body := NewEnclosedSymbolTable(global)
	body.Retain("Compute")
	body.Define("shared", types.STRING, token.Position{Line: 5, Column: 6})

	scoped := body.AllSymbolsWithScope()
	if len(scoped) != 3 {
		t.Fatalf("AllSymbolsWithScope returned %d symbols, want 3: %+v", len(scoped), scoped)
	}

	// Globals come first, ordered by declaration position.
	if scoped[0].Symbol.Name != "shared" || scoped[0].Depth != 0 || scoped[0].Scope != "" {
		t.Errorf("first entry = %+v, want global 'shared' at depth 0", scoped[0])
	}
	if scoped[1].Symbol.Name != "onlyGlobal" || scoped[1].Depth != 0 {
		t.Errorf("second entry = %+v, want global 'onlyGlobal' at depth 0", scoped[1])
	}

	// The shadowing local is preserved rather than overwriting the global.
	last := scoped[2]
	if last.Symbol.Name != "shared" || last.Depth != 1 || last.Scope != "Compute" {
		t.Errorf("third entry = %+v, want local 'shared' at depth 1 in Compute", last)
	}
	if !last.Symbol.Type.Equals(types.STRING) {
		t.Errorf("local 'shared' type = %v, want String", last.Symbol.Type)
	}
	if !scoped[0].Symbol.Type.Equals(types.INTEGER) {
		t.Errorf("global 'shared' type = %v, want Integer", scoped[0].Symbol.Type)
	}

	// AllSymbols still flattens and drops the shadowed global.
	flat := body.AllSymbols()
	if len(flat) != 2 {
		t.Errorf("AllSymbols returned %d symbols, want 2 (flattened)", len(flat))
	}
}

func TestSymbolTable_NestedSymbolsWithScope(t *testing.T) {
	global := NewSymbolTable()
	global.Define("g", types.INTEGER, token.Position{Line: 1, Column: 1})

	body := NewEnclosedSymbolTable(global)
	body.Retain("Compute")
	body.DefineParameter("factor", types.INTEGER, token.Position{Line: 3, Column: 18}, false)

	block := NewEnclosedSymbolTable(body)
	block.Define("temp", types.STRING, token.Position{Line: 5, Column: 7})

	nested := body.NestedSymbolsWithScope()
	if len(nested) != 2 {
		t.Fatalf("NestedSymbolsWithScope returned %d symbols, want 2: %+v", len(nested), nested)
	}
	if nested[0].Symbol.Name != "factor" || nested[0].Depth != 1 {
		t.Errorf("first nested entry = %+v, want 'factor' at depth 1", nested[0])
	}
	if nested[1].Symbol.Name != "temp" || nested[1].Depth != 2 {
		t.Errorf("second nested entry = %+v, want 'temp' at depth 2", nested[1])
	}
	for _, entry := range nested {
		if entry.Scope != "Compute" {
			t.Errorf("entry %q scope = %q, want %q", entry.Symbol.Name, entry.Scope, "Compute")
		}
	}
}

func TestAnalyzer_RetainedScopes(t *testing.T) {
	a := NewAnalyzer()
	if len(a.RetainedScopes()) != 0 {
		t.Fatalf("fresh analyzer has %d retained scopes, want 0", len(a.RetainedScopes()))
	}

	inner := NewEnclosedSymbolTable(a.GetSymbolTable())
	a.retainScope(inner, "Compute")

	retained := a.RetainedScopes()
	if len(retained) != 1 || retained[0] != inner {
		t.Fatalf("RetainedScopes = %+v, want the registered scope", retained)
	}
	if retained[0].ScopeName() != "Compute" {
		t.Errorf("retained scope name = %q, want %q", retained[0].ScopeName(), "Compute")
	}
}
