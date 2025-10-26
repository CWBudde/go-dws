package semantic

import "github.com/cwbudde/go-dws/types"

// Symbol represents a symbol in the symbol table (variable or function)
type Symbol struct {
	Name     string
	Type     types.Type
	ReadOnly bool // True for const variables and exception handler variables (Task 8.207)
}

// SymbolTable manages symbols and scopes during semantic analysis.
// Unlike the interpreter's symbol table, this one tracks compile-time
// type information for variables and functions.
type SymbolTable struct {
	// Current scope's symbols
	symbols map[string]*Symbol

	// Parent scope (nil for global scope)
	outer *SymbolTable
}

// NewSymbolTable creates a new symbol table
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		symbols: make(map[string]*Symbol),
		outer:   nil,
	}
}

// NewEnclosedSymbolTable creates a new symbol table enclosed by an outer scope
func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	st := NewSymbolTable()
	st.outer = outer
	return st
}

// Define defines a new variable symbol in the current scope
func (st *SymbolTable) Define(name string, typ types.Type) {
	st.symbols[name] = &Symbol{
		Name:     name,
		Type:     typ,
		ReadOnly: false,
	}
}

// DefineReadOnly defines a new read-only variable symbol in the current scope (Task 8.207)
func (st *SymbolTable) DefineReadOnly(name string, typ types.Type) {
	st.symbols[name] = &Symbol{
		Name:     name,
		Type:     typ,
		ReadOnly: true,
	}
}

// DefineFunction defines a new function symbol in the current scope
func (st *SymbolTable) DefineFunction(name string, funcType *types.FunctionType) {
	st.symbols[name] = &Symbol{
		Name:     name,
		Type:     funcType,
		ReadOnly: false, // Functions are not assignable
	}
}

// Resolve looks up a symbol by name in the current and outer scopes
func (st *SymbolTable) Resolve(name string) (*Symbol, bool) {
	// Check current scope
	sym, ok := st.symbols[name]
	if ok {
		return sym, true
	}

	// Check outer scope
	if st.outer != nil {
		return st.outer.Resolve(name)
	}

	return nil, false
}

// IsDeclaredInCurrentScope checks if a symbol is declared in the current scope
// (not in any parent scope)
func (st *SymbolTable) IsDeclaredInCurrentScope(name string) bool {
	_, ok := st.symbols[name]
	return ok
}

// PushScope creates a new nested scope
func (st *SymbolTable) PushScope() {
	// The Analyzer will manage this by creating a new SymbolTable
	// This is a helper method for clarity
}

// PopScope returns to the parent scope
func (st *SymbolTable) PopScope() {
	// The Analyzer will manage this by using the outer reference
	// This is a helper method for clarity
}
