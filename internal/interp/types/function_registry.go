// Package types provides the type system for the DWScript interpreter.
// This file implements FunctionRegistry for managing function overloads and qualified names.
package types

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// FunctionRegistry manages function declarations including overloads and
// qualified names (unit-scoped functions).
//
// The registry provides:
// - Case-insensitive function name lookup
// - Multiple overloads per function name
// - Qualified name support (Unit.Function)
// - Builtin function registration and lookup
// - Efficient lookup operations
type FunctionRegistry struct {
	// functions maps function names to overload lists (case-insensitive)
	functions *ident.Map[[]*FunctionEntry]

	// qualifiedFunctions maps "unitname.functionname" to overload lists (case-insensitive)
	// This supports explicit unit qualification (e.g., Math.Abs)
	qualifiedFunctions *ident.Map[[]*FunctionEntry]

	// builtins is the registry for built-in functions
	// Defaults to builtins.DefaultRegistry, but can be customized per instance
	builtins *builtins.Registry
}

// FunctionEntry represents a registered function with metadata.
type FunctionEntry struct {
	// Decl is the function declaration AST node
	Decl *ast.FunctionDecl

	// UnitName is the unit this function belongs to (empty string for global)
	UnitName string

	// Name is the original case-sensitive function name
	Name string
}

// NewFunctionRegistry creates a new empty function registry.
// The builtin registry defaults to builtins.DefaultRegistry.
// Use NewFunctionRegistryWithBuiltins to provide a custom builtin registry.
func NewFunctionRegistry() *FunctionRegistry {
	return NewFunctionRegistryWithBuiltins(builtins.DefaultRegistry)
}

// NewFunctionRegistryWithBuiltins creates a new function registry with a custom builtin registry.
// This is useful for testing or when you want to use a different set of builtin functions.
func NewFunctionRegistryWithBuiltins(builtinReg *builtins.Registry) *FunctionRegistry {
	return &FunctionRegistry{
		functions:          ident.NewMap[[]*FunctionEntry](),
		qualifiedFunctions: ident.NewMap[[]*FunctionEntry](),
		builtins:           builtinReg,
	}
}

// Register adds a function to the registry.
// Multiple functions with the same name can be registered (overloading).
// The name is stored case-insensitively.
func (r *FunctionRegistry) Register(name string, fn *ast.FunctionDecl) {
	if fn == nil {
		return
	}

	entry := &FunctionEntry{
		Decl:     fn,
		UnitName: "",
		Name:     name,
	}

	existing, _ := r.functions.Get(name)
	r.functions.Set(name, append(existing, entry))
}

// RegisterWithUnit adds a function to the registry with an associated unit name.
// This allows for qualified lookups (UnitName.FunctionName).
// The function is registered both globally and in the qualified namespace.
func (r *FunctionRegistry) RegisterWithUnit(unitName, functionName string, fn *ast.FunctionDecl) {
	if fn == nil {
		return
	}

	entry := &FunctionEntry{
		Decl:     fn,
		UnitName: unitName,
		Name:     functionName,
	}

	// Register in global namespace
	existing, _ := r.functions.Get(functionName)
	r.functions.Set(functionName, append(existing, entry))

	// Register in qualified namespace
	qualifiedKey := unitName + "." + functionName
	existingQual, _ := r.qualifiedFunctions.Get(qualifiedKey)
	r.qualifiedFunctions.Set(qualifiedKey, append(existingQual, entry))
}

// Lookup returns all overloads for the given function name.
// The lookup is case-insensitive. Returns nil if no functions found.
func (r *FunctionRegistry) Lookup(name string) []*ast.FunctionDecl {
	entries, ok := r.functions.Get(name)
	if !ok || len(entries) == 0 {
		return nil
	}

	result := make([]*ast.FunctionDecl, len(entries))
	for i, entry := range entries {
		result[i] = entry.Decl
	}
	return result
}

// LookupQualified returns all overloads for a qualified function name (Unit.Function).
// The lookup is case-insensitive. Returns nil if no functions found.
func (r *FunctionRegistry) LookupQualified(unitName, functionName string) []*ast.FunctionDecl {
	qualifiedKey := unitName + "." + functionName
	entries, ok := r.qualifiedFunctions.Get(qualifiedKey)
	if !ok || len(entries) == 0 {
		return nil
	}

	result := make([]*ast.FunctionDecl, len(entries))
	for i, entry := range entries {
		result[i] = entry.Decl
	}
	return result
}

// Exists checks if any function with the given name exists in the registry.
// The check is case-insensitive.
func (r *FunctionRegistry) Exists(name string) bool {
	return r.functions.Has(name)
}

// ExistsQualified checks if a qualified function exists (Unit.Function).
// The check is case-insensitive.
func (r *FunctionRegistry) ExistsQualified(unitName, functionName string) bool {
	qualifiedKey := unitName + "." + functionName
	return r.qualifiedFunctions.Has(qualifiedKey)
}

// GetOverloadCount returns the number of overloads for the given function name.
// Returns 0 if the function doesn't exist.
func (r *FunctionRegistry) GetOverloadCount(name string) int {
	entries, _ := r.functions.Get(name)
	return len(entries)
}

// GetOverloadCountQualified returns the number of overloads for a qualified function.
// Returns 0 if the function doesn't exist.
func (r *FunctionRegistry) GetOverloadCountQualified(unitName, functionName string) int {
	qualifiedKey := unitName + "." + functionName
	entries, _ := r.qualifiedFunctions.Get(qualifiedKey)
	return len(entries)
}

// GetAllFunctions returns a map of all registered functions.
// The map uses normalized keys and contains all overloads for each function.
// The returned map should not be modified directly.
func (r *FunctionRegistry) GetAllFunctions() map[string][]*ast.FunctionDecl {
	result := make(map[string][]*ast.FunctionDecl, r.functions.Len())
	r.functions.Range(func(name string, entries []*FunctionEntry) bool {
		overloads := make([]*ast.FunctionDecl, len(entries))
		for i, entry := range entries {
			overloads[i] = entry.Decl
		}
		result[ident.Normalize(name)] = overloads
		return true
	})
	return result
}

// GetFunctionNames returns a slice of all registered function names (original case).
// Each name appears once, even if it has multiple overloads.
// The names are not sorted.
func (r *FunctionRegistry) GetFunctionNames() []string {
	return r.functions.Keys()
}

// GetFunctionsInUnit returns all function names that belong to a specific unit.
// Returns a map of normalized function names to overload lists.
func (r *FunctionRegistry) GetFunctionsInUnit(unitName string) map[string][]*ast.FunctionDecl {
	result := make(map[string][]*ast.FunctionDecl)

	r.functions.Range(func(name string, entries []*FunctionEntry) bool {
		var unitFuncs []*ast.FunctionDecl
		for _, entry := range entries {
			if ident.Equal(entry.UnitName, unitName) {
				unitFuncs = append(unitFuncs, entry.Decl)
			}
		}
		if len(unitFuncs) > 0 {
			funcKey := ident.Normalize(name)
			result[funcKey] = unitFuncs
		}
		return true
	})

	return result
}

// Count returns the total number of unique function names in the registry.
// Overloads of the same function are counted as one.
func (r *FunctionRegistry) Count() int {
	return r.functions.Len()
}

// TotalOverloads returns the total number of function declarations in the registry.
// This counts all overloads separately.
func (r *FunctionRegistry) TotalOverloads() int {
	total := 0
	r.functions.Range(func(_ string, entries []*FunctionEntry) bool {
		total += len(entries)
		return true
	})
	return total
}

// Clear removes all functions from the registry.
func (r *FunctionRegistry) Clear() {
	r.functions.Clear()
	r.qualifiedFunctions.Clear()
}

// RemoveFunction removes all overloads of a function by name.
// The removal is case-insensitive.
// Returns true if the function was found and removed, false otherwise.
func (r *FunctionRegistry) RemoveFunction(name string) bool {
	if !r.functions.Has(name) {
		return false
	}
	r.functions.Delete(name)

	// Also remove from qualified namespace
	// Collect keys to delete first (can't delete while iterating)
	normalizedName := ident.Normalize(name)
	var keysToDelete []string
	r.qualifiedFunctions.Range(func(qualKey string, _ []*FunctionEntry) bool {
		parts := strings.Split(qualKey, ".")
		if len(parts) == 2 && ident.Normalize(parts[1]) == normalizedName {
			keysToDelete = append(keysToDelete, qualKey)
		}
		return true
	})
	for _, key := range keysToDelete {
		r.qualifiedFunctions.Delete(key)
	}
	return true
}

// FindFunctionsByParameterCount returns all functions that have at least one overload
// with the specified number of parameters.
func (r *FunctionRegistry) FindFunctionsByParameterCount(paramCount int) map[string][]*ast.FunctionDecl {
	result := make(map[string][]*ast.FunctionDecl)

	r.functions.Range(func(name string, entries []*FunctionEntry) bool {
		var matching []*ast.FunctionDecl
		for _, entry := range entries {
			if len(entry.Decl.Parameters) == paramCount {
				matching = append(matching, entry.Decl)
			}
		}
		if len(matching) > 0 {
			result[ident.Normalize(name)] = matching
		}
		return true
	})

	return result
}

// GetFunctionMetadata returns metadata about a function without returning the AST.
// This is useful for introspection without pulling in full declarations.
func (r *FunctionRegistry) GetFunctionMetadata(name string) []FunctionMetadata {
	entries, ok := r.functions.Get(name)
	if !ok || len(entries) == 0 {
		return nil
	}

	result := make([]FunctionMetadata, len(entries))
	for i, entry := range entries {
		result[i] = FunctionMetadata{
			Name:           entry.Name,
			UnitName:       entry.UnitName,
			ParameterCount: len(entry.Decl.Parameters),
			IsForward:      entry.Decl.Body == nil,
		}
	}
	return result
}

// FunctionMetadata contains summary information about a function.
type FunctionMetadata struct {
	Name           string
	UnitName       string
	ParameterCount int
	IsForward      bool // true if declaration without body
}

// ValidateNoConflicts checks if adding a new function would create ambiguous overloads.
// This is a helper for semantic analysis.
// Returns nil if no conflicts, or an error describing the conflict.
func (r *FunctionRegistry) ValidateNoConflicts(name string, paramCount int, hasOverloadDirective bool) error {
	entries, ok := r.functions.Get(name)
	if !ok || len(entries) == 0 {
		// No existing function, no conflict
		return nil
	}

	// Check if any existing overload has the same parameter count
	for _, entry := range entries {
		if len(entry.Decl.Parameters) == paramCount {
			// Same parameter count - this could be ambiguous
			// In DWScript, overload directive is required when parameter counts match
			if !hasOverloadDirective && !entry.Decl.IsOverload {
				return fmt.Errorf("function '%s' with %d parameters already exists; use 'overload' directive", name, paramCount)
			}
		}
	}

	return nil
}

// ===== Builtin Function Support =====

// LookupBuiltin looks up a builtin function by name (case-insensitive).
// Returns the builtin function implementation and true if found, nil and false otherwise.
func (r *FunctionRegistry) LookupBuiltin(name string) (builtins.BuiltinFunc, bool) {
	if r.builtins == nil {
		return nil, false
	}
	return r.builtins.Lookup(name)
}

// GetBuiltinInfo retrieves the full FunctionInfo for a builtin function by name (case-insensitive).
// Returns the info and true if found, nil and false otherwise.
func (r *FunctionRegistry) GetBuiltinInfo(name string) (*builtins.FunctionInfo, bool) {
	if r.builtins == nil {
		return nil, false
	}
	return r.builtins.Get(name)
}

// IsBuiltin checks if a function name refers to a builtin function.
// The check is case-insensitive.
func (r *FunctionRegistry) IsBuiltin(name string) bool {
	if r.builtins == nil {
		return false
	}
	_, ok := r.builtins.Lookup(name)
	return ok
}

// GetBuiltinRegistry returns the underlying builtin registry.
// This is useful for advanced operations like registering custom builtins.
func (r *FunctionRegistry) GetBuiltinRegistry() *builtins.Registry {
	return r.builtins
}

// SetBuiltinRegistry sets the builtin registry to use.
// This is useful for testing or when you want to swap in a different builtin registry.
func (r *FunctionRegistry) SetBuiltinRegistry(reg *builtins.Registry) {
	r.builtins = reg
}

// LookupAny looks up a function by name, checking both user-defined and builtin functions.
// Returns:
//   - userDefined: AST declarations if it's a user-defined function, nil otherwise
//   - isBuiltin: true if the function is a builtin
//   - found: true if any function (user-defined or builtin) was found
//
// User-defined functions take precedence over builtin functions.
func (r *FunctionRegistry) LookupAny(name string) (userDefined []*ast.FunctionDecl, isBuiltin bool, found bool) {
	// Check user-defined functions first
	if decls := r.Lookup(name); len(decls) > 0 {
		return decls, false, true
	}

	// Check builtin functions
	if r.IsBuiltin(name) {
		return nil, true, true
	}

	return nil, false, false
}

// ExistsAny checks if a function with the given name exists in either user-defined or builtin registry.
// The check is case-insensitive.
func (r *FunctionRegistry) ExistsAny(name string) bool {
	return r.Exists(name) || r.IsBuiltin(name)
}

// ===== Declaration/Implementation Support =====

// RegisterOrReplace registers a function, replacing any existing declaration-only
// version if this is an implementation (has a body).
func (r *FunctionRegistry) RegisterOrReplace(name string, fn *ast.FunctionDecl) {
	if fn == nil {
		return
	}

	entry := &FunctionEntry{
		Decl:     fn,
		UnitName: "",
		Name:     name,
	}

	existing, ok := r.functions.Get(name)
	if !ok {
		r.functions.Set(name, []*FunctionEntry{entry})
		return
	}

	// If new function has body, try to replace matching declaration
	if fn.Body != nil {
		for idx, e := range existing {
			if parametersMatchFn(e.Decl.Parameters, fn.Parameters) {
				// Preserve virtual/override/reintroduce/abstract flags from declaration
				fn.IsVirtual = e.Decl.IsVirtual
				fn.IsOverride = e.Decl.IsOverride
				fn.IsReintroduce = e.Decl.IsReintroduce
				fn.IsAbstract = e.Decl.IsAbstract
				existing[idx] = entry
				r.functions.Set(name, existing)
				return
			}
		}
	}

	// No match found or no body, append
	r.functions.Set(name, append(existing, entry))
}

// parametersMatchFn checks if two parameter lists have matching signatures
// (same count and same parameter types).
func parametersMatchFn(params1, params2 []*ast.Parameter) bool {
	if len(params1) != len(params2) {
		return false
	}
	for i := range params1 {
		// Compare parameter types
		if params1[i].Type != nil && params2[i].Type != nil {
			if params1[i].Type.String() != params2[i].Type.String() {
				return false
			}
		} else if params1[i].Type != params2[i].Type {
			// One has type, other doesn't
			return false
		}
	}
	return true
}
