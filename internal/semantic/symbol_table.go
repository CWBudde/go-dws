package semantic

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
)

// Symbol represents a symbol in the symbol table (variable or function)
type Symbol struct {
	Type          types.Type
	Overloads     []*Symbol   // List of overloaded function symbols (nil for non-overloaded)
	Value         interface{} // Compile-time constant value (nil for non-constants)
	Name          string
	ReadOnly      bool
	IsConst       bool
	IsOverloadSet bool // True if this symbol represents multiple overloaded functions
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

// DefineReadOnly defines a new read-only variable symbol in the current scope
func (st *SymbolTable) DefineReadOnly(name string, typ types.Type) {
	st.symbols[strings.ToLower(name)] = &Symbol{
		Name:     name, // Keep original case for error messages
		Type:     typ,
		ReadOnly: true,
		IsConst:  false,
	}
}

// DefineConst defines a new constant symbol in the current scope
func (st *SymbolTable) DefineConst(name string, typ types.Type, value interface{}) {
	st.symbols[strings.ToLower(name)] = &Symbol{
		Name:     name, // Keep original case for error messages
		Type:     typ,
		ReadOnly: true,
		IsConst:  true,
		Value:    value,
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

// DefineOverload defines a new function overload or adds to an existing overload set.
//
// Parameters:
//   - name: Function name
//   - funcType: Function signature
//   - hasOverloadDirective: Whether the function declaration has the 'overload' directive
//
// Returns error if:
//   - Function exists without overload directive on either declaration
//   - Exact duplicate signature exists (even with overload directive)
func (st *SymbolTable) DefineOverload(name string, funcType *types.FunctionType, hasOverloadDirective bool) error {
	lowerName := strings.ToLower(name)
	existing, exists := st.symbols[lowerName]

	if !exists {
		// First declaration - create new symbol (may or may not be overloaded later)
		st.symbols[lowerName] = &Symbol{
			Name:          name,
			Type:          funcType,
			ReadOnly:      false,
			IsConst:       false,
			IsOverloadSet: false, // Not an overload set yet
			Overloads:     nil,
		}
		return nil
	}

	// Check if existing symbol is a function (or overload set of functions)
	if !existing.IsOverloadSet {
		if _, isFuncType := existing.Type.(*types.FunctionType); !isFuncType {
			return fmt.Errorf("'%s' is already declared as a non-function symbol", name)
		}
	}

	// Both the existing and new declaration must have overload directive
	// Note: We can't check the existing one's directive here since we don't store it.
	// This is a simplified version - a full implementation would track the directive.
	// For now, we assume if we're calling DefineOverload, both have the directive.
	if !hasOverloadDirective {
		return fmt.Errorf("function '%s' already declared; use 'overload' directive for multiple definitions", name)
	}

	// Check for duplicate signature in existing overloads
	if existing.IsOverloadSet {
		for _, overload := range existing.Overloads {
			if overload.Type.Equals(funcType) {
				return fmt.Errorf("duplicate function signature for overloaded function '%s'", name)
			}
		}
		// Add to existing overload set
		existing.Overloads = append(existing.Overloads, &Symbol{
			Name:          name,
			Type:          funcType,
			ReadOnly:      false,
			IsConst:       false,
			IsOverloadSet: false,
			Overloads:     nil,
		})
	} else {
		// Check if new signature is different from existing
		if existing.Type.Equals(funcType) {
			return fmt.Errorf("duplicate function signature for overloaded function '%s'", name)
		}
		// Convert to overload set (this is the second overload)
		firstOverload := &Symbol{
			Name:          existing.Name,
			Type:          existing.Type,
			ReadOnly:      false,
			IsConst:       false,
			IsOverloadSet: false,
			Overloads:     nil,
		}
		secondOverload := &Symbol{
			Name:          name,
			Type:          funcType,
			ReadOnly:      false,
			IsConst:       false,
			IsOverloadSet: false,
			Overloads:     nil,
		}
		existing.IsOverloadSet = true
		existing.Overloads = []*Symbol{firstOverload, secondOverload}
		existing.Type = nil // The overload set itself doesn't have a single type
	}

	return nil
}

// GetOverloadSet retrieves all overloads for a given function name.
//
// Returns:
//   - For overloaded functions: slice of all overload symbols
//   - For non-overloaded functions: single-element slice
//   - For non-existent functions: nil
func (st *SymbolTable) GetOverloadSet(name string) []*Symbol {
	lowerName := strings.ToLower(name)
	sym, ok := st.symbols[lowerName]
	if !ok {
		return nil
	}

	if sym.IsOverloadSet {
		return sym.Overloads
	}

	// Non-overloaded function - return as single-element slice
	return []*Symbol{sym}
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
