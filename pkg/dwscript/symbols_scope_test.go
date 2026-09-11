package dwscript

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/token"
)

// scopeFixtureSource declares a global that a function local shadows, plus a
// parameter and a global-only variable, so one compilation covers every scope
// kind Symbols() reports.
const scopeFixtureSource = `
	var shared: Integer := 1;
	var onlyGlobal: String := 'g';

	function Compute(factor: Integer): Integer;
	var shared: String;
	begin
		shared := 'local';
		Result := factor * 2;
	end;
`

// compileScopeFixture compiles scopeFixtureSource and returns the reported
// symbols grouped by name, together with a lookup that selects the occurrence
// belonging to a given scope. Shadowed names have more than one occurrence.
func compileScopeFixture(t *testing.T) (map[string][]Symbol, func(name, scope string) (Symbol, bool)) {
	t.Helper()

	engine, err := New(WithTypeCheck(true))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	program, err := engine.Compile(scopeFixtureSource)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	byName := make(map[string][]Symbol)
	for _, sym := range program.Symbols() {
		byName[sym.Name] = append(byName[sym.Name], sym)
	}

	findScope := func(name, scope string) (Symbol, bool) {
		for _, sym := range byName[name] {
			if sym.Scope == scope {
				return sym, true
			}
		}
		return Symbol{}, false
	}

	return byName, findScope
}

// TestProgram_Symbols_Scope covers the Scope field for globals, a function
// parameter and a function local.
func TestProgram_Symbols_Scope(t *testing.T) {
	byName, findScope := compileScopeFixture(t)

	// A plain global.
	if sym, ok := findScope("onlyGlobal", "global"); !ok {
		t.Errorf("Expected global symbol \"onlyGlobal\", got %+v", byName["onlyGlobal"])
	} else if sym.Type != "String" {
		t.Errorf("Expected onlyGlobal to have type String, got %q", sym.Type)
	}

	// A function is itself a global.
	if _, ok := findScope("Compute", "global"); !ok {
		t.Errorf("Expected function \"Compute\" at global scope, got %+v", byName["Compute"])
	}

	// A function parameter is scoped to the function.
	if sym, ok := findScope("factor", "Compute"); !ok {
		t.Errorf("Expected parameter \"factor\" scoped to Compute, got %+v", byName["factor"])
	} else if sym.Type != "Integer" {
		t.Errorf("Expected factor to have type Integer, got %q", sym.Type)
	}

	// The implicit Result variable is a function local.
	if _, ok := findScope("Result", "Compute"); !ok {
		t.Errorf("Expected \"Result\" scoped to Compute, got %+v", byName["Result"])
	}
}

// TestProgram_Symbols_ShadowedLocal pins the case the old flattened
// AllSymbols() dropped: a local shadowing a global must be reported alongside
// it, with its own scope and type.
func TestProgram_Symbols_ShadowedLocal(t *testing.T) {
	byName, findScope := compileScopeFixture(t)

	globalShared, okGlobal := findScope("shared", "global")
	localShared, okLocal := findScope("shared", "Compute")
	if !okGlobal || !okLocal {
		t.Fatalf("Expected \"shared\" at both global and Compute scope, got %+v", byName["shared"])
	}
	if globalShared.Type != "Integer" {
		t.Errorf("Expected global 'shared' to be Integer, got %q", globalShared.Type)
	}
	if localShared.Type != "String" {
		t.Errorf("Expected local 'shared' to be String, got %q", localShared.Type)
	}
	if globalShared.Scope == localShared.Scope {
		t.Errorf("Expected distinct scopes for the shadowed symbol, both are %q", globalShared.Scope)
	}
}

// TestProgram_Symbols_Position checks that declaration positions are reported.
func TestProgram_Symbols_Position(t *testing.T) {
	source := "var alpha: Integer := 1;\nvar beta: String := 'b';\n"

	engine, err := New(WithTypeCheck(true))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	program, err := engine.Compile(source)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	want := map[string]token.Position{
		"alpha": {Line: 1, Column: 5},
		"beta":  {Line: 2, Column: 5},
	}

	seen := map[string]bool{}
	for _, sym := range program.Symbols() {
		expected, tracked := want[sym.Name]
		if !tracked {
			continue
		}
		seen[sym.Name] = true
		if sym.Position.Line != expected.Line || sym.Position.Column != expected.Column {
			t.Errorf("Symbol %q: expected position %d:%d, got %d:%d",
				sym.Name, expected.Line, expected.Column, sym.Position.Line, sym.Position.Column)
		}
	}
	for name := range want {
		if !seen[name] {
			t.Errorf("Expected symbol %q not found", name)
		}
	}
}
