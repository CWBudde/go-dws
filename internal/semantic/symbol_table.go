package semantic

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

// Symbol represents a symbol in the symbol table (variable or function)
type Symbol struct {
	Type                 types.Type
	Value                interface{}
	Name                 string
	Overloads            []*Symbol
	ReadOnly             bool
	IsConst              bool
	IsOverloadSet        bool
	HasOverloadDirective bool
	IsForward            bool

	// LSP support fields
	DeclPosition       token.Position   // Position where symbol was declared
	Usages             []token.Position // All usage positions for go-to-reference
	Documentation      string           // Doc comment text
	IsDeprecated       bool             // Whether marked as deprecated
	DeprecationMessage string           // Deprecation warning message
}

// SymbolTable manages symbols and scopes during semantic analysis.
// Unlike the interpreter's symbol table, this one tracks compile-time
// type information for variables and functions.
type SymbolTable struct {
	// Current scope's symbols (case-insensitive via ident.Map)
	symbols *ident.Map[*Symbol]

	// Parent scope (nil for global scope)
	outer *SymbolTable
}

// NewSymbolTable creates a new symbol table
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		symbols: ident.NewMap[*Symbol](),
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
// DWScript is case-insensitive, handled by ident.Map
func (st *SymbolTable) Define(name string, typ types.Type, pos token.Position) {
	st.symbols.Set(name, &Symbol{
		Name:         name, // Keep original case for error messages
		Type:         typ,
		ReadOnly:     false,
		IsConst:      false,
		DeclPosition: pos,
		Usages:       make([]token.Position, 0),
	})
}

// DefineReadOnly defines a new read-only variable symbol in the current scope
func (st *SymbolTable) DefineReadOnly(name string, typ types.Type, pos token.Position) {
	st.symbols.Set(name, &Symbol{
		Name:         name, // Keep original case for error messages
		Type:         typ,
		ReadOnly:     true,
		IsConst:      false,
		DeclPosition: pos,
		Usages:       make([]token.Position, 0),
	})
}

// DefineConst defines a new constant symbol in the current scope
func (st *SymbolTable) DefineConst(name string, typ types.Type, value interface{}, pos token.Position) {
	st.symbols.Set(name, &Symbol{
		Name:         name, // Keep original case for error messages
		Type:         typ,
		ReadOnly:     true,
		IsConst:      true,
		Value:        value,
		DeclPosition: pos,
		Usages:       make([]token.Position, 0),
	})
}

// DefineFunction defines a new function symbol in the current scope
func (st *SymbolTable) DefineFunction(name string, funcType *types.FunctionType, pos token.Position) {
	st.symbols.Set(name, &Symbol{
		Name:         name, // Keep original case for error messages
		Type:         funcType,
		ReadOnly:     false, // Functions are not assignable
		IsConst:      false,
		DeclPosition: pos,
		Usages:       make([]token.Position, 0),
	})
}

// DefineOverload defines a new function overload or adds to an existing overload set.
// Returns error if function exists without overload directive, has duplicate signature,
// is ambiguous with default parameters, or forward declaration doesn't match implementation.
func (st *SymbolTable) DefineOverload(name string, funcType *types.FunctionType, hasOverloadDirective bool, isForward bool, pos token.Position) error {
	existing, exists := st.symbols.Get(name)

	if !exists {
		// First declaration - create new symbol (may or may not be overloaded later)
		st.symbols.Set(name, &Symbol{
			Name:                 name,
			Type:                 funcType,
			ReadOnly:             false,
			IsConst:              false,
			IsOverloadSet:        false, // Not an overload set yet
			Overloads:            nil,
			HasOverloadDirective: hasOverloadDirective,
			IsForward:            isForward, // Track if this is a forward declaration
			DeclPosition:         pos,
			Usages:               make([]token.Position, 0),
		})
		return nil
	}

	// Check if existing symbol is a function (or overload set of functions)
	if !existing.IsOverloadSet {
		if _, isFuncType := existing.Type.(*types.FunctionType); !isFuncType {
			return fmt.Errorf("'%s' is already declared as a non-function symbol", name)
		}
	}

	// If existing is a forward declaration and current is an implementation,
	// replace the forward with the implementation
	if !existing.IsOverloadSet && existing.IsForward && !isForward {
		// This is an implementation following a forward declaration
		existingFunc := existing.Type.(*types.FunctionType)

		// Validate that signatures match (including default parameters)
		if !SignaturesEqual(existingFunc, funcType) {
			return fmt.Errorf("implementation signature for '%s' does not match forward declaration", name)
		}
		if !existingFunc.ReturnType.Equals(funcType.ReturnType) {
			return fmt.Errorf("implementation return type for '%s' does not match forward declaration", name)
		}
		if !defaultParametersMatch(existingFunc, funcType) {
			return fmt.Errorf("implementation signature for '%s' does not match forward declaration", name)
		}

		// Validate overload directive consistency
		// Forward can have 'overload' and implementation can omit it (DWScript allows this)
		// But if forward omits 'overload', implementation should too
		if !existing.HasOverloadDirective && hasOverloadDirective {
			return fmt.Errorf("implementation has 'overload' directive but forward declaration does not for '%s'", name)
		}
		// Note: We allow existing.HasOverloadDirective && !hasOverloadDirective (forward has overload, impl doesn't)

		// Replace forward declaration with implementation
		existing.IsForward = false
		existing.Type = funcType // Update to implementation's type (in case of minor differences)
		return nil
	}

	// If existing is an implementation and current is forward, that's an error
	// But allow multiple forward declarations with different signatures (overloads)
	if !existing.IsForward && isForward {
		return fmt.Errorf("forward declaration for '%s' must come before implementation", name)
	}
	// Note: We allow multiple forward declarations with different signatures (overloads)
	// The duplicate forward check is handled below in the duplicate signature check

	// Validate overload directive consistency. Skip if current is implementation (might be replacing forward)
	// If an existing function has the overload directive, or we're adding one with it,
	// then all must have it (with exception: implementations can omit it if forwards have it)
	if !isForward {
		// Current is an implementation - check if it's being added to an overload set
		// If adding to an overload set (not replacing a forward), it must have overload directive
		if existing.IsOverloadSet {
			// Check if this implementation matches any forward in the set
			matchesForward := false
			for _, overload := range existing.Overloads {
				if overload.IsForward {
					existingFunc := overload.Type.(*types.FunctionType)
					// For forward matching, we need EXACT match including default parameters
					if SignaturesEqual(existingFunc, funcType) && existingFunc.ReturnType.Equals(funcType.ReturnType) &&
						defaultParametersMatch(existingFunc, funcType) {
						matchesForward = true
						break
					}
				}
			}
			// If not replacing a forward, must have overload directive
			if !matchesForward && !hasOverloadDirective {
				// Use the function kind from the FIRST overload in the set
				firstFuncType := existing.Overloads[0].Type.(*types.FunctionType)
				return fmt.Errorf("Overloaded %s \"%s\" must be marked with the \"overload\" directive",
					getFunctionKind(firstFuncType), name)
			}
		} else {
			// Not an overload set yet - check if first has directive but second doesn't
			if existing.HasOverloadDirective && !hasOverloadDirective {
				// Use the function kind from the EXISTING (first) function
				existingFuncType := existing.Type.(*types.FunctionType)
				return fmt.Errorf("Overloaded %s \"%s\" must be marked with the \"overload\" directive",
					getFunctionKind(existingFuncType), name)
			}
			if !existing.HasOverloadDirective && hasOverloadDirective {
				// First one didn't have it, but second does
				// DWScript allows this in some cases, so we'll be lenient here
				// and just update the first one to mark it as part of an overload set
				existing.HasOverloadDirective = true
			}
		}
	} else if existing.IsOverloadSet {
		// Current is a forward - check all existing overloads for directive consistency
		for _, overload := range existing.Overloads {
			if overload.HasOverloadDirective && !hasOverloadDirective {
				// Use the function kind from the FIRST overload in the set
				firstFuncType := existing.Overloads[0].Type.(*types.FunctionType)
				return fmt.Errorf("Overloaded %s \"%s\" must be marked with the \"overload\" directive",
					getFunctionKind(firstFuncType), name)
			}
		}
	} else {
		// Second overload (both forwards) - both first and second must have directive (or neither, for now)
		if existing.HasOverloadDirective && !hasOverloadDirective {
			// Use the function kind from the EXISTING (first) function
			existingFuncType := existing.Type.(*types.FunctionType)
			return fmt.Errorf("Overloaded %s \"%s\" must be marked with the \"overload\" directive",
				getFunctionKind(existingFuncType), name)
		}
		if !existing.HasOverloadDirective && hasOverloadDirective {
			// First one didn't have it, but second does - this is also an error
			// However, DWScript allows this in some cases, so we'll be lenient here
			// and just update the first one to mark it as part of an overload set
			existing.HasOverloadDirective = true
		}
	}

	// Check for duplicate signature in existing overloads
	if existing.IsOverloadSet {
		for i, overload := range existing.Overloads {
			existingFunc := overload.Type.(*types.FunctionType)
			if SignaturesEqual(existingFunc, funcType) {
				if existingFunc.ReturnType.Equals(funcType.ReturnType) {
					hasDefaults1 := hasDefaultParameters(existingFunc)
					hasDefaults2 := hasDefaultParameters(funcType)

					// Forward + implementation pair: check if default parameters match
					if overload.IsForward && !isForward {
						if !defaultParametersMatch(existingFunc, funcType) {
							continue
						}
						// Forward can have 'overload'; implementation can omit it
						if !overload.HasOverloadDirective && hasOverloadDirective {
							return fmt.Errorf("implementation has 'overload' directive but forward declaration does not for '%s'", name)
						}
						existing.Overloads[i].IsForward = false
						existing.Overloads[i].Type = funcType
						return nil
					}

					// Check for duplicate forwards (both are forward declarations with same signature)
					if overload.IsForward && isForward {
						return fmt.Errorf("duplicate forward declaration for '%s'", name)
					}

					if hasDefaults1 && !hasDefaults2 {
						return fmt.Errorf("There is already a method with name \"%s\"", name)
					} else if !hasDefaults1 && hasDefaults2 {
						continue
					} else if defaultParametersMatch(existingFunc, funcType) {
						return fmt.Errorf("There is already a method with name \"%s\"", name)
					} else {
						continue
					}
				}
				// Signatures match but return types differ - this is a valid overload
				// (procedures vs functions, or functions with different return types)
				// Continue to allow this overload
			}
		}
	} else {
		// Check if new signature is different from existing
		existingFunc := existing.Type.(*types.FunctionType)
		if SignaturesEqual(existingFunc, funcType) {
			// Signatures match - check if return types also match
			if existingFunc.ReturnType.Equals(funcType.ReturnType) {
				// Signatures and return types match - check default parameters
				// Rule: If existing has defaults but new doesn't, it's a duplicate (new is redundant)
				//       If existing doesn't have defaults but new does, it's ambiguous (new expands but creates ambiguity)
				hasDefaults1 := hasDefaultParameters(existingFunc)
				hasDefaults2 := hasDefaultParameters(funcType)

				if hasDefaults1 && !hasDefaults2 {
					return fmt.Errorf("There is already a method with name \"%s\"", name)
				} else if !hasDefaults1 && hasDefaults2 {
					// New adds default parameters - this will be caught by ambiguity check
					// Don't return duplicate error
				} else {
					// Both have defaults or neither has defaults - check if they match exactly
					if defaultParametersMatch(existingFunc, funcType) {
						// Exact match - true duplicate
						return fmt.Errorf("There is already a method with name \"%s\"", name)
					}
					// Different default values - will be caught by ambiguity check
				}
			}
			// Return types differ - valid overload (procedures vs functions)
		}
	}

	// Check for ambiguous overloads with default parameters
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
			IsForward:            isForward, // Track forward declarations in overload sets
			DeclPosition:         pos,
			Usages:               make([]token.Position, 0),
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
			IsForward:            existing.IsForward, // Preserve forward status
			DeclPosition:         existing.DeclPosition,
			Usages:               existing.Usages,
			Documentation:        existing.Documentation,
			IsDeprecated:         existing.IsDeprecated,
			DeprecationMessage:   existing.DeprecationMessage,
		}
		secondOverload := &Symbol{
			Name:                 name,
			Type:                 funcType,
			ReadOnly:             false,
			IsConst:              false,
			IsOverloadSet:        false,
			Overloads:            nil,
			HasOverloadDirective: hasOverloadDirective,
			IsForward:            isForward, // Track forward declarations
			DeclPosition:         pos,
			Usages:               make([]token.Position, 0),
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

// checkAmbiguousOverload checks if a new overload would be ambiguous with existing overloads,
// especially with default parameters.
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
			return fmt.Errorf("Overload of \"%s\" will be ambiguous with a previously declared version", name)
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

// GetOverloadSet retrieves all overloads for a function name.
// Returns slice of all overload symbols, single-element slice for non-overloaded functions, or nil.
func (st *SymbolTable) GetOverloadSet(name string) []*Symbol {
	sym, ok := st.symbols.Get(name)
	if !ok {
		// Check outer scope recursively (like Resolve does)
		if st.outer != nil {
			return st.outer.GetOverloadSet(name)
		}
		return nil
	}

	if sym.IsOverloadSet {
		return sym.Overloads
	}

	// Non-overloaded function - return as single-element slice
	return []*Symbol{sym}
}

// Resolve looks up a symbol by name in the current and outer scopes (case-insensitive).
func (st *SymbolTable) Resolve(name string) (*Symbol, bool) {
	sym, ok := st.symbols.Get(name)
	if ok {
		return sym, true
	}
	if st.outer != nil {
		return st.outer.Resolve(name)
	}
	return nil, false
}

// IsDeclaredInCurrentScope checks if a symbol is declared in the current scope (case-insensitive).
func (st *SymbolTable) IsDeclaredInCurrentScope(name string) bool {
	return st.symbols.Has(name)
}

// PushScope creates a new nested scope (managed by Analyzer).
func (st *SymbolTable) PushScope() {}

// PopScope returns to the parent scope (managed by Analyzer).
func (st *SymbolTable) PopScope() {}

// AllSymbols returns all symbols in the current and outer scopes (keys normalized).
func (st *SymbolTable) AllSymbols() map[string]*Symbol {
	result := make(map[string]*Symbol)
	if st.outer != nil {
		for name, sym := range st.outer.AllSymbols() {
			result[name] = sym
		}
	}
	st.symbols.Range(func(name string, sym *Symbol) bool {
		result[ident.Normalize(name)] = sym
		return true
	})
	return result
}

// RecordUsage records a usage of a symbol at the given position (for LSP find-references).
func (st *SymbolTable) RecordUsage(name string, pos token.Position) {
	sym, ok := st.symbols.Get(name)
	if ok {
		sym.Usages = append(sym.Usages, pos)
		return
	}
	if st.outer != nil {
		st.outer.RecordUsage(name, pos)
	}
}

// FindDefinition finds the definition of a symbol by name (case-insensitive).
func (st *SymbolTable) FindDefinition(name string) (*Symbol, token.Position, bool) {
	sym, ok := st.symbols.Get(name)
	if ok {
		return sym, sym.DeclPosition, true
	}
	if st.outer != nil {
		return st.outer.FindDefinition(name)
	}
	return nil, token.Position{}, false
}

// FindReferences returns all usage positions for a given symbol name (case-insensitive).
func (st *SymbolTable) FindReferences(name string) []token.Position {
	sym, ok := st.symbols.Get(name)
	if ok {
		refs := make([]token.Position, len(sym.Usages))
		copy(refs, sym.Usages)
		return refs
	}
	if st.outer != nil {
		return st.outer.FindReferences(name)
	}
	return nil
}

// UnusedSymbols returns all declared but unused symbols in the current scope.
func (st *SymbolTable) UnusedSymbols() []*Symbol {
	var unused []*Symbol
	st.symbols.Range(func(name string, sym *Symbol) bool {
		if len(sym.Usages) == 0 {
			if sym.IsOverloadSet {
				anyUsed := false
				for _, overload := range sym.Overloads {
					if len(overload.Usages) > 0 {
						anyUsed = true
						break
					}
				}
				if !anyUsed {
					unused = append(unused, sym)
				}
			} else {
				unused = append(unused, sym)
			}
		}
		return true
	})
	return unused
}

// hasDefaultParameters checks if a function signature has any default parameters
func hasDefaultParameters(sig *types.FunctionType) bool {
	if sig.DefaultValues == nil {
		return false
	}
	for _, defaultVal := range sig.DefaultValues {
		if defaultVal != nil {
			return true
		}
	}
	return false
}

// defaultParametersMatch checks if two function signatures have matching default parameters.
// Used for forward declaration matching - forwards and implementations must match exactly.
func defaultParametersMatch(sig1, sig2 *types.FunctionType) bool {
	if len(sig1.Parameters) != len(sig2.Parameters) {
		return false
	}
	for i := 0; i < len(sig1.Parameters); i++ {
		has1 := i < len(sig1.DefaultValues) && sig1.DefaultValues[i] != nil
		has2 := i < len(sig2.DefaultValues) && sig2.DefaultValues[i] != nil
		if has1 != has2 {
			return false
		}
	}
	return true
}
