package semantic

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
)

// Symbol represents a symbol in the symbol table (variable or function)
type Symbol struct {
	Type             types.Type
	Overloads        []*Symbol   // List of overloaded function symbols (nil for non-overloaded)
	Value            interface{} // Compile-time constant value (nil for non-constants)
	Name             string
	ReadOnly         bool
	IsConst          bool
	IsOverloadSet    bool // True if this symbol represents multiple overloaded functions
	HasOverloadDirective bool // True if function has explicit 'overload' directive (Task 9.58)
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
//   - Function exists without overload directive on either declaration (Task 9.58)
//   - Exact duplicate signature exists (Task 9.59)
//   - Ambiguous overload with default parameters (Task 9.62)
func (st *SymbolTable) DefineOverload(name string, funcType *types.FunctionType, hasOverloadDirective bool) error {
	lowerName := strings.ToLower(name)
	existing, exists := st.symbols[lowerName]

	if !exists {
		// First declaration - create new symbol (may or may not be overloaded later)
		st.symbols[lowerName] = &Symbol{
			Name:                 name,
			Type:                 funcType,
			ReadOnly:             false,
			IsConst:              false,
			IsOverloadSet:        false, // Not an overload set yet
			Overloads:            nil,
			HasOverloadDirective: hasOverloadDirective,
		}
		return nil
	}

	// Check if existing symbol is a function (or overload set of functions)
	if !existing.IsOverloadSet {
		if _, isFuncType := existing.Type.(*types.FunctionType); !isFuncType {
			return fmt.Errorf("'%s' is already declared as a non-function symbol", name)
		}
	}

	// Task 9.58: Validate overload directive consistency
	// If an existing function has the overload directive, or we're adding one with it,
	// then all must have it (with exception: the last implementation can omit it in some cases)
	if existing.IsOverloadSet {
		// Check all existing overloads for directive consistency
		for _, overload := range existing.Overloads {
			if overload.HasOverloadDirective && !hasOverloadDirective {
				return fmt.Errorf("overloaded %s '%s' must be marked with the 'overload' directive",
					getFunctionKind(funcType), name)
			}
		}
	} else {
		// Second overload - both first and second must have directive (or neither, for now)
		if existing.HasOverloadDirective && !hasOverloadDirective {
			return fmt.Errorf("overloaded %s '%s' must be marked with the 'overload' directive",
				getFunctionKind(funcType), name)
		}
		if !existing.HasOverloadDirective && hasOverloadDirective {
			// First one didn't have it, but second does - this is also an error
			// However, DWScript allows this in some cases, so we'll be lenient here
			// and just update the first one to mark it as part of an overload set
			existing.HasOverloadDirective = true
		}
	}

	// Task 9.59: Check for duplicate signature in existing overloads
	// Note: In DWScript, functions with same parameters but different return types
	// are allowed as overloads. Only if BOTH signature AND return type match is it a duplicate.
	if existing.IsOverloadSet {
		for _, overload := range existing.Overloads {
			existingFunc := overload.Type.(*types.FunctionType)
			// Check if signatures are equal (same parameters)
			if SignaturesEqual(existingFunc, funcType) {
				// Signatures match - check if return types also match
				if existingFunc.ReturnType.Equals(funcType.ReturnType) {
					// True duplicate - same signature AND same return type
					if hasOverloadDirective && overload.HasOverloadDirective {
						return fmt.Errorf("there is already a method with name \"%s\"", name)
					}
					return fmt.Errorf("'%s' already declared", name)
				}
				// Signatures match but return types differ - this is allowed in DWScript
			}
		}
	} else {
		// Check if new signature is different from existing
		existingFunc := existing.Type.(*types.FunctionType)
		if SignaturesEqual(existingFunc, funcType) {
			// Signatures match - check if return types also match
			if existingFunc.ReturnType.Equals(funcType.ReturnType) {
				// True duplicate - same signature AND same return type
				if hasOverloadDirective && existing.HasOverloadDirective {
					return fmt.Errorf("there is already a method with name \"%s\"", name)
				}
				return fmt.Errorf("'%s' already declared", name)
			}
			// Signatures match but return types differ - this is allowed in DWScript
		}
	}

	// Task 9.62: Check for ambiguous overloads (especially with default parameters)
	if err := st.checkAmbiguousOverload(name, funcType, existing); err != nil {
		return err
	}

	// Add the new overload
	if existing.IsOverloadSet {
		// Add to existing overload set
		existing.Overloads = append(existing.Overloads, &Symbol{
			Name:                 name,
			Type:                 funcType,
			ReadOnly:             false,
			IsConst:              false,
			IsOverloadSet:        false,
			Overloads:            nil,
			HasOverloadDirective: hasOverloadDirective,
		})
	} else {
		// Convert to overload set (this is the second overload)
		firstOverload := &Symbol{
			Name:                 existing.Name,
			Type:                 existing.Type,
			ReadOnly:             false,
			IsConst:              false,
			IsOverloadSet:        false,
			Overloads:            nil,
			HasOverloadDirective: existing.HasOverloadDirective,
		}
		secondOverload := &Symbol{
			Name:                 name,
			Type:                 funcType,
			ReadOnly:             false,
			IsConst:              false,
			IsOverloadSet:        false,
			Overloads:            nil,
			HasOverloadDirective: hasOverloadDirective,
		}
		existing.IsOverloadSet = true
		existing.Overloads = []*Symbol{firstOverload, secondOverload}
		existing.Type = nil // The overload set itself doesn't have a single type
	}

	return nil
}

// getFunctionKind returns "procedure" or "function" based on return type
func getFunctionKind(funcType *types.FunctionType) string {
	if funcType.ReturnType == types.VOID {
		return "procedure"
	}
	return "function"
}

// checkAmbiguousOverload checks if a new overload would be ambiguous with existing overloads
// Task 9.62: Detect ambiguous overloads, especially with default parameters
func (st *SymbolTable) checkAmbiguousOverload(name string, newSig *types.FunctionType, existing *Symbol) error {
	// Get all existing signatures
	var existingSigs []*types.FunctionType
	if existing.IsOverloadSet {
		for _, overload := range existing.Overloads {
			existingSigs = append(existingSigs, overload.Type.(*types.FunctionType))
		}
	} else {
		existingSigs = []*types.FunctionType{existing.Type.(*types.FunctionType)}
	}

	// Check if the new signature could be ambiguous with any existing signature
	// due to default parameters
	for _, existingSig := range existingSigs {
		if isAmbiguous(newSig, existingSig) {
			return fmt.Errorf("overload of \"%s\" will be ambiguous with a previously declared version", name)
		}
	}

	return nil
}

// isAmbiguous checks if two function signatures are ambiguous due to default parameters
// Two signatures are ambiguous if there's a call that could match both
// Note: In DWScript, same signature with different return types is allowed (not ambiguous)
func isAmbiguous(sig1, sig2 *types.FunctionType) bool {
	// If only return types differ, not ambiguous (DWScript allows this)
	// The disambiguation happens based on the context where the result is used
	if SignaturesEqual(sig1, sig2) && !sig1.ReturnType.Equals(sig2.ReturnType) {
		return false // Not ambiguous - only return types differ
	}

	// Get the parameter counts and minimum required counts
	params1 := len(sig1.Parameters)
	params2 := len(sig2.Parameters)

	// Count required parameters (non-default) for each signature
	required1 := params1
	required2 := params2

	if sig1.DefaultValues != nil {
		for i := len(sig1.DefaultValues) - 1; i >= 0; i-- {
			if sig1.DefaultValues[i] != nil {
				required1 = i
			} else {
				break
			}
		}
	}

	if sig2.DefaultValues != nil {
		for i := len(sig2.DefaultValues) - 1; i >= 0; i-- {
			if sig2.DefaultValues[i] != nil {
				required2 = i
			} else {
				break
			}
		}
	}

	// Check if there's an overlap in the number of arguments that both could accept
	minArgs1 := required1
	maxArgs1 := params1
	minArgs2 := required2
	maxArgs2 := params2

	// For each possible argument count, check if both signatures could match
	for argCount := 0; argCount <= max(maxArgs1, maxArgs2); argCount++ {
		// Check if both signatures accept this argument count
		canAccept1 := argCount >= minArgs1 && argCount <= maxArgs1
		canAccept2 := argCount >= minArgs2 && argCount <= maxArgs2

		if canAccept1 && canAccept2 {
			// Both signatures accept this number of arguments
			// Check if the parameter types and modifiers are compatible
			allMatch := true
			for i := 0; i < argCount; i++ {
				if i < params1 && i < params2 {
					// Compare the types
					if !sig1.Parameters[i].Equals(sig2.Parameters[i]) {
						allMatch = false
						break
					}
					// Also check parameter modifiers - different modifiers make signatures distinct
					if i < len(sig1.VarParams) && i < len(sig2.VarParams) {
						if sig1.VarParams[i] != sig2.VarParams[i] {
							allMatch = false
							break
						}
					}
					if i < len(sig1.ConstParams) && i < len(sig2.ConstParams) {
						if sig1.ConstParams[i] != sig2.ConstParams[i] {
							allMatch = false
							break
						}
					}
					if i < len(sig1.LazyParams) && i < len(sig2.LazyParams) {
						if sig1.LazyParams[i] != sig2.LazyParams[i] {
							allMatch = false
							break
						}
					}
				}
			}

			if allMatch {
				// This argument count would match both signatures - ambiguous!
				return true
			}
		}
	}

	return false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
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

// AllSymbols returns all symbols in the current scope and all outer scopes.
// Used for symbol extraction for LSP features (Task 10.15).
func (st *SymbolTable) AllSymbols() map[string]*Symbol {
	result := make(map[string]*Symbol)

	// Collect symbols from outer scopes first (so current scope can override)
	if st.outer != nil {
		for name, sym := range st.outer.AllSymbols() {
			result[name] = sym
		}
	}

	// Add symbols from current scope (overrides outer symbols with same name)
	for name, sym := range st.symbols {
		result[name] = sym
	}

	return result
}
