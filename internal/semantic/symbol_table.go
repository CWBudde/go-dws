package semantic

import (
	"fmt"
	"sort"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

// Symbol represents a symbol in the symbol table (variable or function)
type Symbol struct {
	Type types.Type
	// ClassFieldOwner is set only on the synthesized bindings that make a class's
	// own fields visible by bare name inside a method body or a property
	// expression accessor. It names the declaring class, so a bare-name reference
	// can be attributed back to the field for unused-private-field tracking; a
	// local that shadows a field is an ordinary symbol and leaves it nil.
	ClassFieldOwner       *types.ClassType
	Value                 interface{}
	Name                  string
	DeprecationMessage    string
	Documentation         string
	Overloads             []*Symbol
	Usages                []token.Position
	DeclPosition          token.Position
	IsForward             bool
	HasOverloadDirective  bool
	IsOverloadSet         bool
	IsConst               bool
	IsDeprecated          bool
	ReadOnly              bool
	SuppressUnusedWarning bool
	// IsLoopVariable marks a `for` control variable. Writing to one is legal but
	// draws DWScript's `Assignment to FOR-Loop variable` warning, so the flag has
	// to survive until the body is analyzed; it is not the same as ReadOnly.
	IsLoopVariable bool
}

// SymbolTable manages symbols and scopes during semantic analysis.
// Unlike the interpreter's symbol table, this one tracks compile-time
// type information for variables and functions.
type SymbolTable struct {
	// exportedTypes accompanies a unit interface scope when it is imported.
	exportedTypes map[string]types.Type
	// Current scope's symbols (case-insensitive via ident.Map)
	symbols *ident.Map[*Symbol]

	// Parent scope (nil for global scope)
	outer *SymbolTable

	// children holds the nested scopes that were kept alive for post-analysis
	// inspection. It stays empty unless this scope is retained (see Retain):
	// the analyzer creates and discards many short-lived scopes, so scopes are
	// not linked to their parent by default.
	children []*SymbolTable

	// scopeName names the construct that owns this scope, such as the function
	// whose body it is. Nested scopes inherit it; it is empty for the global
	// scope.
	scopeName string

	// depth is the nesting level of this scope: 0 for the global scope, one
	// more than the enclosing scope otherwise.
	depth int

	// retain records that this scope, and every scope nested inside it,
	// survives analysis.
	retain bool
}

// ScopedSymbol pairs a symbol with the scope that declared it.
type ScopedSymbol struct {
	Symbol *Symbol
	// Scope names the construct owning the declaring scope (a function or a
	// lambda). It is empty for the global scope.
	Scope string
	// Depth is the nesting level of the declaring scope: 0 for the global scope.
	Depth int
}

// NewSymbolTable creates a new symbol table
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		symbols: ident.NewMap[*Symbol](),
		outer:   nil,
	}
}

// NewEnclosedSymbolTable creates a new symbol table enclosed by an outer scope.
// The new scope sits one level deeper than its parent and inherits the parent's
// scope name. If the parent is retained, so is the child, and the child is
// linked into the parent's children.
func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	st := NewSymbolTable()
	st.outer = outer
	if outer != nil {
		st.depth = outer.depth + 1
		st.scopeName = outer.scopeName
		if outer.retain {
			st.retain = true
			outer.children = append(outer.children, st)
		}
	}
	return st
}

// Retain marks this scope, and every scope later nested inside it, as surviving
// analysis so that it can be inspected afterwards (for example by
// Program.Symbols). name identifies the owning construct and is inherited by
// nested scopes.
//
// Retention is opt-in rather than the default because the analyzer allocates
// throwaway scopes - a per-call scope for unit-qualified calls, for instance -
// that would otherwise accumulate for the lifetime of the analyzer.
func (st *SymbolTable) Retain(name string) {
	st.retain = true
	st.scopeName = name
}

// linkedToRetainedParent reports whether this scope is already reachable from
// its enclosing scope's children, which NewEnclosedSymbolTable arranges for
// every scope opened inside a retained one. Callers that enumerate retained
// scopes and recurse into Children use it to avoid visiting such a scope twice.
func (st *SymbolTable) linkedToRetainedParent() bool {
	if st.outer == nil || !st.outer.retain {
		return false
	}
	for _, child := range st.outer.children {
		if child == st {
			return true
		}
	}
	return false
}

// Depth returns the nesting level of this scope: 0 for the global scope.
func (st *SymbolTable) Depth() int {
	return st.depth
}

// ScopeName returns the name of the construct owning this scope, or "" when the
// scope is global or unnamed.
func (st *SymbolTable) ScopeName() string {
	return st.scopeName
}

// Children returns the retained scopes nested directly inside this one.
func (st *SymbolTable) Children() []*SymbolTable {
	return st.children
}

// LocalSymbols returns the symbols declared in this scope only, ordered by
// declaration position and then by name so that the result is deterministic.
func (st *SymbolTable) LocalSymbols() []ScopedSymbol {
	result := make([]ScopedSymbol, 0, st.symbols.Len())
	st.symbols.Range(func(_ string, sym *Symbol) bool {
		result = append(result, ScopedSymbol{Symbol: sym, Scope: st.scopeName, Depth: st.depth})
		return true
	})
	sortScopedSymbols(result)
	return result
}

// AllSymbolsWithScope returns the symbols visible from this scope, walking the
// enclosing chain from the outermost scope inwards. Unlike AllSymbols it does
// not flatten the chain: every scope contributes its own symbols, each tagged
// with the depth and name of the scope that declared it, so a local that
// shadows an outer symbol appears alongside the symbol it shadows.
func (st *SymbolTable) AllSymbolsWithScope() []ScopedSymbol {
	var result []ScopedSymbol
	if st.outer != nil {
		result = st.outer.AllSymbolsWithScope()
	}
	return append(result, st.LocalSymbols()...)
}

// NestedSymbolsWithScope returns the symbols declared in this scope and,
// recursively, in every retained scope nested inside it. The enclosing chain is
// not included.
func (st *SymbolTable) NestedSymbolsWithScope() []ScopedSymbol {
	result := st.LocalSymbols()
	for _, child := range st.children {
		result = append(result, child.NestedSymbolsWithScope()...)
	}
	return result
}

// sortScopedSymbols orders symbols by declaration position, then by name.
func sortScopedSymbols(symbols []ScopedSymbol) {
	sort.SliceStable(symbols, func(i, j int) bool {
		a, b := symbols[i].Symbol, symbols[j].Symbol
		if a.DeclPosition.Line != b.DeclPosition.Line {
			return a.DeclPosition.Line < b.DeclPosition.Line
		}
		if a.DeclPosition.Column != b.DeclPosition.Column {
			return a.DeclPosition.Column < b.DeclPosition.Column
		}
		return a.Name < b.Name
	})
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

// DefineClassField defines a synthesized binding that exposes a class field by
// bare name in the current scope, recording the declaring class on the symbol.
func (st *SymbolTable) DefineClassField(name string, typ types.Type, owner *types.ClassType) {
	st.symbols.Set(name, &Symbol{
		Name:            name, // Keep original case for error messages
		Type:            typ,
		ClassFieldOwner: owner,
		DeclPosition:    token.Position{},
		Usages:          make([]token.Position, 0),
	})
}

// DefineReadOnly defines a new read-only variable symbol in the current scope
func (st *SymbolTable) DefineReadOnly(name string, typ types.Type, pos token.Position) {
	st.symbols.Set(name, &Symbol{
		Name:                  name, // Keep original case for error messages
		Type:                  typ,
		ReadOnly:              true,
		IsConst:               false,
		DeclPosition:          pos,
		Usages:                make([]token.Position, 0),
		SuppressUnusedWarning: true,
	})
}

// DefineParameter defines a new parameter symbol in the current scope.
func (st *SymbolTable) DefineParameter(name string, typ types.Type, pos token.Position, readOnly bool) {
	st.symbols.Set(name, &Symbol{
		Name:                  name,
		Type:                  typ,
		ReadOnly:              readOnly,
		IsConst:               false,
		DeclPosition:          pos,
		Usages:                make([]token.Position, 0),
		SuppressUnusedWarning: true,
	})
}

// DefineLoopVariable defines a new loop control variable that should not
// participate in unused-variable warnings.
func (st *SymbolTable) DefineLoopVariable(name string, typ types.Type, pos token.Position) {
	st.symbols.Set(name, &Symbol{
		Name:                  name,
		Type:                  typ,
		ReadOnly:              false,
		IsConst:               false,
		DeclPosition:          pos,
		Usages:                make([]token.Position, 0),
		SuppressUnusedWarning: true,
		IsLoopVariable:        true,
	})
}

// DefineConst defines a new constant symbol in the current scope
func (st *SymbolTable) DefineConst(name string, typ types.Type, value interface{}, pos token.Position) {
	st.symbols.Set(name, &Symbol{
		Name:                  name, // Keep original case for error messages
		Type:                  typ,
		ReadOnly:              true,
		IsConst:               true,
		Value:                 value,
		DeclPosition:          pos,
		Usages:                make([]token.Position, 0),
		SuppressUnusedWarning: true,
	})
}

// DefineFunction defines a new function symbol in the current scope
func (st *SymbolTable) DefineFunction(name string, funcType *types.FunctionType, pos token.Position) {
	st.symbols.Set(name, &Symbol{
		Name:                  name, // Keep original case for error messages
		Type:                  funcType,
		ReadOnly:              false, // Functions are not assignable
		IsConst:               false,
		DeclPosition:          pos,
		Usages:                make([]token.Position, 0),
		SuppressUnusedWarning: true,
	})
}

// MarkDeprecated records a `deprecated` directive on an already-defined symbol.
// It is separate from DefineOverload because the directive belongs to the name,
// not to one signature: a forward declaration, its implementation and every
// overload share a single symbol, and a reference to the name is what upstream
// warns about.
func (st *SymbolTable) MarkDeprecated(name, message string) {
	sym, ok := st.symbols.Get(name)
	if !ok {
		if st.outer != nil {
			st.outer.MarkDeprecated(name, message)
		}
		return
	}
	sym.IsDeprecated = true
	if message != "" {
		sym.DeprecationMessage = message
	}
}

// DefineOverload defines a new function overload or adds to an existing overload set.
// Returns error if function exists without overload directive, has duplicate signature,
// is ambiguous with default parameters, or forward declaration doesn't match implementation.
func (st *SymbolTable) DefineOverload(
	name string,
	funcType *types.FunctionType,
	hasOverloadDirective bool,
	isForward bool,
	pos token.Position,
) error {
	existing, exists := st.symbols.Get(name)

	if !exists {
		st.defineNewSymbol(name, funcType, hasOverloadDirective, isForward, pos)
		return nil
	}

	if err := st.ensureFunctionSymbol(name, existing); err != nil {
		return err
	}

	// Handle simple forward replacement (non-overload set)
	if !existing.IsOverloadSet && existing.IsForward && !isForward {
		return st.replaceForwardWithImplementation(name, existing, funcType, hasOverloadDirective)
	}

	if !existing.IsForward && isForward {
		return fmt.Errorf("forward declaration for '%s' must come before implementation", name)
	}

	if err := st.validateOverloadDirectives(name, existing, hasOverloadDirective, isForward); err != nil {
		return err
	}

	// Check for duplicate signatures and resolve forward declarations in overload sets
	resolved, err := st.checkSignaturesAndResolveForward(name, existing, funcType, hasOverloadDirective, isForward)
	if err != nil {
		return err
	}
	if resolved {
		return nil
	}

	// Check for ambiguous overloads with default parameters
	if err := st.checkAmbiguousOverload(name, funcType, existing); err != nil {
		return err
	}

	return st.addOverloadToSet(name, existing, funcType, hasOverloadDirective, isForward, pos)
}

func (st *SymbolTable) defineNewSymbol(name string, funcType *types.FunctionType, hasOverloadDirective, isForward bool, pos token.Position) {
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
}

func (st *SymbolTable) ensureFunctionSymbol(name string, existing *Symbol) error {
	if !existing.IsOverloadSet {
		if _, isFuncType := existing.Type.(*types.FunctionType); !isFuncType {
			return fmt.Errorf("'%s' is already declared as a non-function symbol", name)
		}
	}
	return nil
}

func (st *SymbolTable) replaceForwardWithImplementation(name string, existing *Symbol, funcType *types.FunctionType, hasOverloadDirective bool) error {
	// This is an implementation following a forward declaration
	existingFunc, ok := existing.Type.(*types.FunctionType)
	if !ok {
		return fmt.Errorf("symbol '%s' is not a function type", name)
	}

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

func (st *SymbolTable) validateOverloadDirectives(name string, existing *Symbol, hasOverloadDirective, isForward bool) error {
	if isForward {
		// Current is a forward - check all existing overloads for directive consistency
		if existing.IsOverloadSet {
			for _, overload := range existing.Overloads {
				if overload.HasOverloadDirective && !hasOverloadDirective {
					firstFuncType, ok := existing.Overloads[0].Type.(*types.FunctionType)
					if !ok {
						return fmt.Errorf("expected function type for overload, but got %T", existing.Overloads[0].Type)
					}
					return fmt.Errorf("overloaded %s \"%s\" must be marked with the \"overload\" directive",
						getFunctionKind(firstFuncType), name)
				}
			}
		} else {
			// Second overload (both forwards) - both first and second must have directive (or neither, for now)
			if existing.HasOverloadDirective && !hasOverloadDirective {
				existingFuncType, ok := existing.Type.(*types.FunctionType)
				if !ok {
					return fmt.Errorf("expected function type for existing symbol '%s', but got %T", name, existing.Type)
				}
				return fmt.Errorf("overloaded %s \"%s\" must be marked with the \"overload\" directive",
					getFunctionKind(existingFuncType), name)
			}
			if !existing.HasOverloadDirective && hasOverloadDirective {
				// First one didn't have it, but second does - this is also an error
				// However, DWScript allows this in some cases, so we'll be lenient here
				// and just update the first one to mark it as part of an overload set
				existing.HasOverloadDirective = true
			}
		}
		return nil
	}

	// Current is an implementation
	if !existing.IsOverloadSet {
		// Not an overload set yet - check if first has directive but second doesn't
		if existing.HasOverloadDirective && !hasOverloadDirective {
			existingFuncType, ok := existing.Type.(*types.FunctionType)
			if !ok {
				return fmt.Errorf("expected function type for existing symbol '%s', but got %T", name, existing.Type)
			}
			return fmt.Errorf("overloaded %s \"%s\" must be marked with the \"overload\" directive",
				getFunctionKind(existingFuncType), name)
		}
		if !existing.HasOverloadDirective && hasOverloadDirective {
			// First one didn't have it, but second does
			existing.HasOverloadDirective = true
		}
	}
	// For existing.IsOverloadSet case, we check directive requirement in addOverloadToSet
	// because we need to know if it matches a forward first (which is checked in checkSignaturesAndResolveForward)

	return nil
}

func (st *SymbolTable) checkSignaturesAndResolveForward(
	name string,
	existing *Symbol,
	funcType *types.FunctionType,
	hasOverloadDirective, isForward bool,
) (bool, error) {
	if existing.IsOverloadSet {
		for i, overload := range existing.Overloads {
			existingFunc, ok := overload.Type.(*types.FunctionType)
			if !ok {
				// This should ideally not happen if 'overload' is indeed part of an overload set
				return false, fmt.Errorf("expected function type for overload in set, but got %T", overload.Type)
			}
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
							return false, fmt.Errorf("implementation has 'overload' directive but forward declaration does not for '%s'", name)
						}
						existing.Overloads[i].IsForward = false
						existing.Overloads[i].Type = funcType
						return true, nil
					}

					// Check for duplicate forwards
					if overload.IsForward && isForward {
						return false, fmt.Errorf("duplicate forward declaration for '%s'", name)
					}

					if hasDefaults1 && !hasDefaults2 {
						return false, fmt.Errorf("there is already a method with name \"%s\"", name)
					} else if !hasDefaults1 && hasDefaults2 {
						continue
					} else if defaultParametersMatch(existingFunc, funcType) {
						return false, fmt.Errorf("there is already a method with name \"%s\"", name)
					} else {
						continue
					}
				}
				// Signatures match but return types differ - this is a valid overload
			}
		}
	} else {
		// Check if new signature is different from existing
		existingFunc, ok := existing.Type.(*types.FunctionType)
		if !ok {
			// This should ideally not happen if 'existing' is indeed a function symbol
			return false, fmt.Errorf("expected function type for existing symbol, but got %T", existing.Type)
		}
		if SignaturesEqual(existingFunc, funcType) {
			if existingFunc.ReturnType.Equals(funcType.ReturnType) {
				hasDefaults1 := hasDefaultParameters(existingFunc)
				hasDefaults2 := hasDefaultParameters(funcType)

				if hasDefaults1 && !hasDefaults2 {
					return false, fmt.Errorf("there is already a method with name \"%s\"", name)
				} else if !hasDefaults1 && hasDefaults2 {
					// New adds default parameters - this will be caught by ambiguity check
				} else {
					if defaultParametersMatch(existingFunc, funcType) {
						return false, fmt.Errorf("there is already a method with name \"%s\"", name)
					}
				}
			}
		}
	}
	return false, nil
}

func (st *SymbolTable) addOverloadToSet(name string, existing *Symbol, funcType *types.FunctionType, hasOverloadDirective, isForward bool, pos token.Position) error {
	// Check if we are adding to an existing overload set (and not replacing a forward, which is handled before).
	// If so, we must have the overload directive.
	if existing.IsOverloadSet && !isForward && !hasOverloadDirective {
		// Use the function kind from the FIRST overload in the set
		firstFuncType, ok := existing.Overloads[0].Type.(*types.FunctionType)
		if !ok {
			return fmt.Errorf("expected function type for first overload in set, but got %T", existing.Overloads[0].Type)
		}
		return fmt.Errorf("overloaded %s \"%s\" must be marked with the \"overload\" directive",
			getFunctionKind(firstFuncType), name)
	}

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
			IsForward:            isForward,
			DeclPosition:         pos,
			Usages:               make([]token.Position, 0),
		})
	} else {
		// Convert to overload set
		firstOverload := &Symbol{
			Name:                 existing.Name,
			Type:                 existing.Type,
			ReadOnly:             false,
			IsConst:              false,
			IsOverloadSet:        false,
			Overloads:            nil,
			HasOverloadDirective: existing.HasOverloadDirective,
			IsForward:            existing.IsForward,
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
			IsForward:            isForward,
			DeclPosition:         pos,
			Usages:               make([]token.Position, 0),
		}
		existing.IsOverloadSet = true
		existing.Overloads = []*Symbol{firstOverload, secondOverload}
		existing.Type = nil
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
			funcType, ok := overload.Type.(*types.FunctionType)
			if !ok {
				// This indicates an internal inconsistency, should not happen
				return fmt.Errorf("expected function type for overload, but got %T", overload.Type)
			}
			existingSigs = append(existingSigs, funcType)
		}
	} else {
		funcType, ok := existing.Type.(*types.FunctionType)
		if !ok {
			// This indicates an internal inconsistency, should not happen
			return fmt.Errorf("expected function type for existing symbol, but got %T", existing.Type)
		}
		existingSigs = []*types.FunctionType{funcType}
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
