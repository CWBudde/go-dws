package dwscript

import (
	"fmt"
	"testing"
)

// symbolKey identifies one reported symbol. Two entries sharing a key are the
// same declaration reported twice, not two distinct symbols: a shadowing local
// differs in Scope, and an overload differs in Type.
func symbolKey(s Symbol) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s", s.Kind, s.Name, s.Scope, s.Type, s.Position.String())
}

// assertNoDuplicateSymbols compiles source and fails if Program.Symbols reports
// the same declaration more than once.
func assertNoDuplicateSymbols(t *testing.T, source string) {
	t.Helper()

	engine, err := New(WithTypeCheck(true))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	program, err := engine.Compile(source)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	counts := make(map[string]int)
	for _, sym := range program.Symbols() {
		counts[symbolKey(sym)]++
	}
	for key, n := range counts {
		if n > 1 {
			t.Errorf("symbol reported %d times, expected once: %s", n, key)
		}
	}
}

// TestProgram_Symbols_NoDuplicates_LambdaInFunction covers a lambda analyzed
// inside an already-retained function body. The lambda's scope is reachable
// both as a child of the function's retained scope and, before the fix, as a
// retained scope of its own, which reported its parameters and Result twice.
func TestProgram_Symbols_NoDuplicates_LambdaInFunction(t *testing.T) {
	assertNoDuplicateSymbols(t, `
		function Outer(a: Integer): Integer;
		var f: function(x: Integer): Integer;
		begin
			f := lambda(x: Integer) => x + a;
			Result := f(a);
		end;
	`)
}

// TestProgram_Symbols_NoDuplicates_NestedLambda covers a lambda nested inside
// another lambda, i.e. two levels of retained scope.
func TestProgram_Symbols_NoDuplicates_NestedLambda(t *testing.T) {
	assertNoDuplicateSymbols(t, `
		function Outer(a: Integer): Integer;
		var f: function(x: Integer): Integer;
		begin
			f := lambda(x: Integer) => (lambda(y: Integer) => y * 2)(x) + a;
			Result := f(a);
		end;
	`)
}

// TestProgram_Symbols_NoDuplicates_TopLevelLambda guards the case the fix must
// not regress: a lambda at program scope has no retained parent, so it must
// still be reported (exactly once).
func TestProgram_Symbols_NoDuplicates_TopLevelLambda(t *testing.T) {
	source := `
		var f: function(x: Integer): Integer;
		f := lambda(x: Integer) => x * 3;
		PrintLn(f(2));
	`
	assertNoDuplicateSymbols(t, source)

	engine, err := New(WithTypeCheck(true))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	program, err := engine.Compile(source)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}
	found := false
	for _, sym := range program.Symbols() {
		if sym.Scope == "lambda" && sym.Name == "x" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected the top-level lambda's parameter to still be reported, got %+v", program.Symbols())
	}
}

// TestProgram_Symbols_NoDuplicates_Functions covers plain function bodies,
// which are retained at top level and must keep being reported once each.
func TestProgram_Symbols_NoDuplicates_Functions(t *testing.T) {
	assertNoDuplicateSymbols(t, `
		var shared: Integer := 1;

		function First(a: Integer): Integer;
		var local: String;
		begin
			local := 'x';
			Result := a + shared;
		end;

		function Second(b: Integer): Integer;
		begin
			Result := First(b) * 2;
		end;
	`)
}
