package interp

import (
	"testing"
)

// TestNewSymbolTable tests creating a new symbol table
func TestNewSymbolTable(t *testing.T) {
	st := NewSymbolTable()

	if st == nil {
		t.Fatal("NewSymbolTable() returned nil")
	}
}

// TestDefineSymbol tests defining a symbol in the table
func TestDefineSymbol(t *testing.T) {
	st := NewSymbolTable()

	sym := st.Define("x")

	if sym == nil {
		t.Fatal("Define() returned nil")
	}

	if sym.Name != "x" {
		t.Errorf("symbol name = %q, want 'x'", sym.Name)
	}

	if sym.Type != SymbolVariable {
		t.Errorf("symbol type = %v, want SymbolVariable", sym.Type)
	}
}

// TestResolveSymbol tests resolving a symbol from the table
func TestResolveSymbol(t *testing.T) {
	st := NewSymbolTable()

	st.Define("x")

	sym, ok := st.Resolve("x")
	if !ok {
		t.Fatal("Resolve() failed to find defined symbol 'x'")
	}

	if sym.Name != "x" {
		t.Errorf("resolved symbol name = %q, want 'x'", sym.Name)
	}
}

// TestResolveUndefinedSymbol tests resolving a symbol that doesn't exist
func TestResolveUndefinedSymbol(t *testing.T) {
	st := NewSymbolTable()

	_, ok := st.Resolve("undefined")
	if ok {
		t.Error("Resolve() should return false for undefined symbol")
	}
}

// TestSymbolRedeclaration tests that redefining a symbol updates it
func TestSymbolRedeclaration(t *testing.T) {
	st := NewSymbolTable()

	sym1 := st.Define("x")
	sym2 := st.Define("x")

	if sym1.Name != sym2.Name {
		t.Error("Redefining symbol should return symbol with same name")
	}
}

// TestEnclosedSymbolTable tests creating enclosed (nested) symbol tables
func TestEnclosedSymbolTable(t *testing.T) {
	outer := NewSymbolTable()
	outer.Define("x")

	inner := NewEnclosedSymbolTable(outer)

	if inner == nil {
		t.Fatal("NewEnclosedSymbolTable() returned nil")
	}

	// Inner should be able to resolve from outer
	sym, ok := inner.Resolve("x")
	if !ok {
		t.Fatal("inner table should resolve symbol from outer table")
	}

	if sym.Name != "x" {
		t.Errorf("resolved symbol name = %q, want 'x'", sym.Name)
	}
}

// TestNestedSymbolScope tests that inner scopes shadow outer scopes
func TestNestedSymbolScope(t *testing.T) {
	outer := NewSymbolTable()
	outer.Define("x")

	inner := NewEnclosedSymbolTable(outer)
	inner.Define("x") // Shadow outer 'x'
	inner.Define("y") // Inner-only symbol

	// Resolve from inner - should get inner's 'x'
	sym, ok := inner.Resolve("x")
	if !ok {
		t.Fatal("failed to resolve 'x' in inner scope")
	}

	if sym.Scope != ScopeLocal {
		t.Errorf("inner 'x' should have ScopeLocal, got %v", sym.Scope)
	}

	// Resolve from outer - should get outer's 'x'
	sym, ok = outer.Resolve("x")
	if !ok {
		t.Fatal("failed to resolve 'x' in outer scope")
	}

	if sym.Scope != ScopeGlobal {
		t.Errorf("outer 'x' should have ScopeGlobal, got %v", sym.Scope)
	}

	// 'y' should not be visible in outer scope
	_, ok = outer.Resolve("y")
	if ok {
		t.Error("outer scope should not see inner-only symbol 'y'")
	}
}

// TestSymbolTypes tests different symbol types
func TestSymbolTypes(t *testing.T) {
	st := NewSymbolTable()

	// Define a variable
	varSym := st.Define("x")
	if varSym.Type != SymbolVariable {
		t.Errorf("variable symbol type = %v, want SymbolVariable", varSym.Type)
	}

	// Define a function
	funcSym := st.DefineFunction("Add")
	if funcSym.Type != SymbolFunction {
		t.Errorf("function symbol type = %v, want SymbolFunction", funcSym.Type)
	}
}

// TestUpdateSymbol tests updating an existing symbol
func TestUpdateSymbol(t *testing.T) {
	st := NewSymbolTable()

	sym := st.Define("x")
	originalIndex := sym.Index

	// Update should succeed
	ok := st.Update("x")
	if !ok {
		t.Error("Update() should succeed for defined symbol")
	}

	// Verify symbol still exists with same index
	updated, ok := st.Resolve("x")
	if !ok {
		t.Fatal("symbol should still exist after update")
	}

	if updated.Index != originalIndex {
		t.Errorf("updated symbol index = %d, want %d", updated.Index, originalIndex)
	}
}

// TestUpdateUndefinedSymbol tests updating a symbol that doesn't exist
func TestUpdateUndefinedSymbol(t *testing.T) {
	st := NewSymbolTable()

	ok := st.Update("undefined")
	if ok {
		t.Error("Update() should return false for undefined symbol")
	}
}

// TestSymbolScopes tests symbol scope tracking
func TestSymbolScopes(t *testing.T) {
	global := NewSymbolTable()
	global.Define("global")

	local := NewEnclosedSymbolTable(global)
	local.Define("local")

	// Check global symbol has global scope
	sym, ok := global.Resolve("global")
	if !ok {
		t.Fatal("failed to resolve global symbol")
	}
	if sym.Scope != ScopeGlobal {
		t.Errorf("global symbol scope = %v, want ScopeGlobal", sym.Scope)
	}

	// Check local symbol has local scope when resolved from local table
	sym, ok = local.Resolve("local")
	if !ok {
		t.Fatal("failed to resolve local symbol")
	}
	if sym.Scope != ScopeLocal {
		t.Errorf("local symbol scope = %v, want ScopeLocal", sym.Scope)
	}

	// Check global symbol is marked as free when resolved from local scope
	sym, ok = local.Resolve("global")
	if !ok {
		t.Fatal("failed to resolve global symbol from local scope")
	}
	if sym.Scope != ScopeGlobal {
		t.Errorf("global symbol resolved from local scope = %v, want ScopeGlobal", sym.Scope)
	}
}
