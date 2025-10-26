package semantic

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
)

// Symbol represents a symbol in the symbol table (variable or function)
type Symbol struct {
	Name     string
	Type     types.Type
	ReadOnly bool // True for const variables and exception handler variables (Task 8.207)
	IsConst  bool // True for const declarations (Task 8.254)
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
// DWScript is case-insensitive, so we normalize to lowercase for lookup
func (st *SymbolTable) Define(name string, typ types.Type) {
	st.symbols[strings.ToLower(name)] = &Symbol{
		Name:     name, // Keep original case for error messages
		Type:     typ,
		ReadOnly: false,
		IsConst:  false,
	}
}

// DefineReadOnly defines a new read-only variable symbol in the current scope (Task 8.207)
func (st *SymbolTable) DefineReadOnly(name string, typ types.Type) {
	st.symbols[strings.ToLower(name)] = &Symbol{
		Name:     name, // Keep original case for error messages
		Type:     typ,
		ReadOnly: true,
		IsConst:  false,
	}
}

// DefineConst defines a new constant symbol in the current scope (Task 8.254)
func (st *SymbolTable) DefineConst(name string, typ types.Type) {
	st.symbols[strings.ToLower(name)] = &Symbol{
		Name:     name, // Keep original case for error messages
		Type:     typ,
		ReadOnly: true,
		IsConst:  true,
	}
}

// DefineFunction defines a new function symbol in the current scope
func (st *SymbolTable) DefineFunction(name string, funcType *types.FunctionType) {
	st.symbols[strings.ToLower(name)] = &Symbol{
		Name:     name, // Keep original case for error messages
		Type:     funcType,
		ReadOnly: false, // Functions are not assignable
		IsConst:  false,
	}
}

// Resolve looks up a symbol by name in the current and outer scopes
// DWScript is case-insensitive, so we normalize to lowercase for lookup
func (st *SymbolTable) Resolve(name string) (*Symbol, bool) {
	// Check current scope (case-insensitive)
	sym, ok := st.symbols[strings.ToLower(name)]
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
// (not in any parent scope). DWScript is case-insensitive.
func (st *SymbolTable) IsDeclaredInCurrentScope(name string) bool {
	_, ok := st.symbols[strings.ToLower(name)]
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
