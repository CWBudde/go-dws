package interp

// SymbolType represents the type of symbol (variable, function, etc.)
type SymbolType int

const (
	SymbolVariable SymbolType = iota
	SymbolFunction
)

// SymbolScope represents the scope level of a symbol
type SymbolScope int

const (
	ScopeGlobal SymbolScope = iota
	ScopeLocal
	ScopeFree
)

// Symbol represents a symbol in the symbol table
type Symbol struct {
	Name  string
	Type  SymbolType
	Scope SymbolScope
	Index int
}

// SymbolTable manages symbols in a scope
type SymbolTable struct {
	outer *SymbolTable
	store map[string]*Symbol
	numDefinitions int
}

// NewSymbolTable creates a new symbol table
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		store: make(map[string]*Symbol),
	}
}

// NewEnclosedSymbolTable creates a new symbol table enclosed by an outer scope
func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	st := NewSymbolTable()
	st.outer = outer
	return st
}

// Define defines a new symbol in the symbol table
func (st *SymbolTable) Define(name string) *Symbol {
	sym := &Symbol{
		Name:  name,
		Type:  SymbolVariable,
		Index: st.numDefinitions,
	}

	// Determine scope based on whether we have an outer table
	if st.outer == nil {
		sym.Scope = ScopeGlobal
	} else {
		sym.Scope = ScopeLocal
	}

	st.store[name] = sym
	st.numDefinitions++
	return sym
}

// DefineFunction defines a new function symbol in the symbol table
func (st *SymbolTable) DefineFunction(name string) *Symbol {
	sym := st.Define(name)
	sym.Type = SymbolFunction
	return sym
}

// Resolve looks up a symbol by name in the current and outer scopes
func (st *SymbolTable) Resolve(name string) (*Symbol, bool) {
	sym, ok := st.store[name]
	if ok {
		return sym, true
	}

	// If not found locally, try outer scope
	if st.outer != nil {
		sym, ok := st.outer.Resolve(name)
		if ok {
			return sym, true
		}
	}

	return nil, false
}

// Update updates an existing symbol (marks it as assigned)
func (st *SymbolTable) Update(name string) bool {
	_, ok := st.store[name]
	if ok {
		return true
	}

	// Try outer scope
	if st.outer != nil {
		return st.outer.Update(name)
	}

	return false
}
