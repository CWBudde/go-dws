// Package types provides the type system for the DWScript interpreter.
// This file implements FunctionRegistry for managing function overloads and qualified names.
//
// Task 3.4.3: Create FunctionRegistry with overload support
package types

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// FunctionRegistry manages function declarations including overloads and
// qualified names (unit-scoped functions).
//
// The registry provides:
// - Case-insensitive function name lookup
// - Multiple overloads per function name
// - Qualified name support (Unit.Function)
// - Efficient lookup operations
type FunctionRegistry struct {
	// functions maps normalized function names to overload lists
	functions map[string][]*FunctionEntry

	// qualifiedFunctions maps "unitname.functionname" (normalized) to overload lists
	// This supports explicit unit qualification (e.g., Math.Abs)
	qualifiedFunctions map[string][]*FunctionEntry
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
func NewFunctionRegistry() *FunctionRegistry {
	return &FunctionRegistry{
		functions:          make(map[string][]*FunctionEntry),
		qualifiedFunctions: make(map[string][]*FunctionEntry),
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

	key := ident.Normalize(name)
	r.functions[key] = append(r.functions[key], entry)
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
	funcKey := ident.Normalize(functionName)
	r.functions[funcKey] = append(r.functions[funcKey], entry)

	// Register in qualified namespace
	qualifiedKey := ident.Normalize(unitName + "." + functionName)
	r.qualifiedFunctions[qualifiedKey] = append(r.qualifiedFunctions[qualifiedKey], entry)
}

// Lookup returns all overloads for the given function name.
// The lookup is case-insensitive. Returns nil if no functions found.
func (r *FunctionRegistry) Lookup(name string) []*ast.FunctionDecl {
	entries := r.functions[ident.Normalize(name)]
	if len(entries) == 0 {
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
	qualifiedKey := ident.Normalize(unitName + "." + functionName)
	entries := r.qualifiedFunctions[qualifiedKey]
	if len(entries) == 0 {
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
	_, exists := r.functions[ident.Normalize(name)]
	return exists
}

// ExistsQualified checks if a qualified function exists (Unit.Function).
// The check is case-insensitive.
func (r *FunctionRegistry) ExistsQualified(unitName, functionName string) bool {
	qualifiedKey := ident.Normalize(unitName + "." + functionName)
	_, exists := r.qualifiedFunctions[qualifiedKey]
	return exists
}

// GetOverloadCount returns the number of overloads for the given function name.
// Returns 0 if the function doesn't exist.
func (r *FunctionRegistry) GetOverloadCount(name string) int {
	return len(r.functions[ident.Normalize(name)])
}

// GetOverloadCountQualified returns the number of overloads for a qualified function.
// Returns 0 if the function doesn't exist.
func (r *FunctionRegistry) GetOverloadCountQualified(unitName, functionName string) int {
	qualifiedKey := ident.Normalize(unitName + "." + functionName)
	return len(r.qualifiedFunctions[qualifiedKey])
}

// GetAllFunctions returns a map of all registered functions.
// The map uses normalized keys and contains all overloads for each function.
// The returned map should not be modified directly.
func (r *FunctionRegistry) GetAllFunctions() map[string][]*ast.FunctionDecl {
	result := make(map[string][]*ast.FunctionDecl, len(r.functions))
	for key, entries := range r.functions {
		overloads := make([]*ast.FunctionDecl, len(entries))
		for i, entry := range entries {
			overloads[i] = entry.Decl
		}
		result[key] = overloads
	}
	return result
}

// GetFunctionNames returns a slice of all registered function names (original case).
// Each name appears once, even if it has multiple overloads.
// The names are not sorted.
func (r *FunctionRegistry) GetFunctionNames() []string {
	seen := make(map[string]bool)
	names := []string{}

	for _, entries := range r.functions {
		if len(entries) > 0 {
			name := entries[0].Name
			if !seen[ident.Normalize(name)] {
				names = append(names, name)
				seen[ident.Normalize(name)] = true
			}
		}
	}

	return names
}

// GetFunctionsInUnit returns all function names that belong to a specific unit.
// Returns a map of normalized function names to overload lists.
func (r *FunctionRegistry) GetFunctionsInUnit(unitName string) map[string][]*ast.FunctionDecl {
	result := make(map[string][]*ast.FunctionDecl)
	unitNormalized := ident.Normalize(unitName)

	for _, entries := range r.functions {
		var unitFuncs []*ast.FunctionDecl
		for _, entry := range entries {
			if ident.Normalize(entry.UnitName) == unitNormalized {
				unitFuncs = append(unitFuncs, entry.Decl)
			}
		}
		if len(unitFuncs) > 0 && len(entries) > 0 {
			funcKey := ident.Normalize(entries[0].Name)
			result[funcKey] = unitFuncs
		}
	}

	return result
}

// Count returns the total number of unique function names in the registry.
// Overloads of the same function are counted as one.
func (r *FunctionRegistry) Count() int {
	return len(r.functions)
}

// TotalOverloads returns the total number of function declarations in the registry.
// This counts all overloads separately.
func (r *FunctionRegistry) TotalOverloads() int {
	total := 0
	for _, entries := range r.functions {
		total += len(entries)
	}
	return total
}

// Clear removes all functions from the registry.
func (r *FunctionRegistry) Clear() {
	r.functions = make(map[string][]*FunctionEntry)
	r.qualifiedFunctions = make(map[string][]*FunctionEntry)
}

// RemoveFunction removes all overloads of a function by name.
// The removal is case-insensitive.
// Returns true if the function was found and removed, false otherwise.
func (r *FunctionRegistry) RemoveFunction(name string) bool {
	key := ident.Normalize(name)
	if _, exists := r.functions[key]; exists {
		delete(r.functions, key)

		// Also remove from qualified namespace
		for qualKey := range r.qualifiedFunctions {
			parts := strings.Split(qualKey, ".")
			if len(parts) == 2 && ident.Normalize(parts[1]) == key {
				delete(r.qualifiedFunctions, qualKey)
			}
		}
		return true
	}
	return false
}

// FindFunctionsByParameterCount returns all functions that have at least one overload
// with the specified number of parameters.
func (r *FunctionRegistry) FindFunctionsByParameterCount(paramCount int) map[string][]*ast.FunctionDecl {
	result := make(map[string][]*ast.FunctionDecl)

	for key, entries := range r.functions {
		var matching []*ast.FunctionDecl
		for _, entry := range entries {
			if len(entry.Decl.Parameters) == paramCount {
				matching = append(matching, entry.Decl)
			}
		}
		if len(matching) > 0 {
			result[key] = matching
		}
	}

	return result
}

// GetFunctionMetadata returns metadata about a function without returning the AST.
// This is useful for introspection without pulling in full declarations.
func (r *FunctionRegistry) GetFunctionMetadata(name string) []FunctionMetadata {
	entries := r.functions[ident.Normalize(name)]
	if len(entries) == 0 {
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
	entries := r.functions[ident.Normalize(name)]

	if len(entries) == 0 {
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
