package units

import (
	"fmt"
	"os"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// UnitRegistry manages loaded units and handles unit dependencies.
// It caches compiled units to avoid reloading and detects circular dependencies.
type UnitRegistry struct {
	// units maps normalized unit names to loaded Unit instances
	units map[string]*Unit

	// loading tracks units currently being loaded to detect circular dependencies
	loading map[string]bool

	// cache stores parsed units to speed up repeated loads
	cache *UnitCache

	// loadingChain tracks the current dependency chain for error reporting
	loadingChain []string

	// searchPaths are directories to search for unit files
	searchPaths []string
}

// NewUnitRegistry creates a new unit registry with the given search paths.
// If no search paths are provided, only the current directory is searched.
func NewUnitRegistry(searchPaths []string) *UnitRegistry {
	if searchPaths == nil {
		searchPaths = []string{"."}
	}
	return &UnitRegistry{
		units:        make(map[string]*Unit),
		loading:      make(map[string]bool),
		loadingChain: make([]string, 0),
		searchPaths:  searchPaths,
		cache:        NewUnitCache(),
	}
}

// RegisterUnit registers a unit in the registry.
// Returns an error if a unit with the same name (case-insensitive) is already registered.
func (r *UnitRegistry) RegisterUnit(name string, unit *Unit) error {
	normalized := ident.Normalize(name)

	// Check if already registered
	if _, exists := r.units[normalized]; exists {
		return fmt.Errorf("unit '%s' is already registered", name)
	}

	// Register the unit
	r.units[normalized] = unit
	return nil
}

// GetUnit retrieves a unit by name from the registry.
// Returns the unit and true if found, nil and false otherwise.
// The name lookup is case-insensitive.
func (r *UnitRegistry) GetUnit(name string) (*Unit, bool) {
	normalized := ident.Normalize(name)
	unit, exists := r.units[normalized]
	return unit, exists
}

// LoadUnit loads a unit by name, searching in the configured search paths.
// The unit is parsed, its dependencies are recursively loaded, and it's registered in the registry.
// Returns an error if:
//   - The unit file cannot be found
//   - The unit file cannot be parsed
//   - A circular dependency is detected
//   - A dependency cannot be loaded
func (r *UnitRegistry) LoadUnit(name string, searchPaths []string) (*Unit, error) {
	normalized := ident.Normalize(name)

	// Check if already loaded in registry
	if unit, exists := r.units[normalized]; exists {
		return unit, nil
	}

	// Check compilation cache before parsing
	if cachedUnit, found := r.cache.Get(normalized); found {
		// Unit found in cache and is still valid - use it
		r.units[normalized] = cachedUnit
		return cachedUnit, nil
	}

	// Check for circular dependency
	if r.loading[normalized] {
		// Build the cycle path for better error reporting
		cyclePath := append(r.loadingChain, name)
		return nil, fmt.Errorf("circular dependency detected: %s", strings.Join(cyclePath, " -> "))
	}

	// Mark as loading and add to chain
	r.loading[normalized] = true
	r.loadingChain = append(r.loadingChain, name)
	defer func() {
		// Clean up loading state when done
		delete(r.loading, normalized)
		if len(r.loadingChain) > 0 {
			r.loadingChain = r.loadingChain[:len(r.loadingChain)-1]
		}
	}()

	// Merge search paths (prefer provided paths, then registry's default paths)
	paths := searchPaths
	if len(paths) == 0 {
		paths = r.searchPaths
	}

	// Find the unit file
	filePath, err := FindUnit(name, paths)
	if err != nil {
		return nil, fmt.Errorf("cannot load unit '%s': %w", name, err)
	}

	// Read the source file
	source, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read unit file '%s': %w", filePath, err)
	}

	// Parse the unit file
	l := lexer.New(string(source))
	p := parser.New(l)
	program := p.ParseProgram()

	// Check for parsing errors
	if len(p.Errors()) > 0 {
		errorMessages := make([]string, len(p.Errors()))
		for i, err := range p.Errors() {
			errorMessages[i] = err.Error()
		}
		return nil, fmt.Errorf("parse errors in unit '%s': %s", name, strings.Join(errorMessages, "; "))
	}

	// Extract the unit declaration from the program
	// A unit file should have exactly one statement: the UnitDeclaration
	if len(program.Statements) == 0 {
		return nil, fmt.Errorf("unit file '%s' is empty", filePath)
	}

	unitDecl, ok := program.Statements[0].(*ast.UnitDeclaration)
	if !ok {
		// This might be a program file, not a unit file
		return nil, fmt.Errorf("file '%s' is not a unit (expected 'unit' declaration)", filePath)
	}

	// Create the unit from the parsed declaration
	unit := NewUnit(unitDecl.Name.Value, filePath)

	// Extract sections from the parsed AST
	unit.InterfaceSection = unitDecl.InterfaceSection
	unit.ImplementationSection = unitDecl.ImplementationSection
	unit.InitializationSection = unitDecl.InitSection
	unit.FinalizationSection = unitDecl.FinalSection

	// Extract uses clauses from interface section
	if unitDecl.InterfaceSection != nil {
		for _, stmt := range unitDecl.InterfaceSection.Statements {
			if usesClause, ok := stmt.(*ast.UsesClause); ok {
				for _, unitIdent := range usesClause.Units {
					unit.Uses = append(unit.Uses, unitIdent.Value)
				}
			}
		}
	}

	// Extract uses clauses from implementation section
	if unitDecl.ImplementationSection != nil {
		for _, stmt := range unitDecl.ImplementationSection.Statements {
			if usesClause, ok := stmt.(*ast.UsesClause); ok {
				for _, unitIdent := range usesClause.Units {
					// Only add if not already in the uses list from interface
					found := false
					for _, existing := range unit.Uses {
						if strings.EqualFold(existing, unitIdent.Value) {
							found = true
							break
						}
					}
					if !found {
						unit.Uses = append(unit.Uses, unitIdent.Value)
					}
				}
			}
		}
	}

	// Load dependencies recursively (if any)
	for _, depName := range unit.Uses {
		_, err := r.LoadUnit(depName, paths)
		if err != nil {
			return nil, fmt.Errorf("failed to load dependency '%s' for unit '%s': %w", depName, name, err)
		}
	}

	// Register the unit
	if err := r.RegisterUnit(name, unit); err != nil {
		return nil, err
	}

	// Add to compilation cache
	r.cache.Put(normalized, unit, filePath)

	return unit, nil
}

// UnregisterUnit removes a unit from the registry.
// This is primarily useful for testing or when reloading a unit.
func (r *UnitRegistry) UnregisterUnit(name string) {
	normalized := ident.Normalize(name)
	delete(r.units, normalized)
}

// Clear removes all units from the registry.
// This is primarily useful for testing.
func (r *UnitRegistry) Clear() {
	r.units = make(map[string]*Unit)
	r.loading = make(map[string]bool)
}

// ListUnits returns a list of all registered unit names.
func (r *UnitRegistry) ListUnits() []string {
	names := make([]string, 0, len(r.units))
	for _, unit := range r.units {
		names = append(names, unit.Name)
	}
	return names
}

// GetCache returns the unit registry's compilation cache
func (r *UnitRegistry) GetCache() *UnitCache {
	return r.cache
}

// InvalidateCache invalidates a specific unit in the cache
func (r *UnitRegistry) InvalidateCache(name string) {
	normalized := ident.Normalize(name)
	r.cache.Invalidate(normalized)
}

// ClearCache clears all entries from the compilation cache
func (r *UnitRegistry) ClearCache() {
	r.cache.Clear()
}

// ComputeInitializationOrder returns the order in which units should be initialized
// using topological sort (Kahn's algorithm). Units with no dependencies are initialized
// first, followed by units that depend on them.
//
// Returns an error if a circular dependency is detected.
//
// Example:
//
//	If Unit B uses Unit A, and Unit C uses Unit B, the order will be: [A, B, C]
func (r *UnitRegistry) ComputeInitializationOrder() ([]string, error) {
	// Build dependency graph
	// inDegree tracks how many dependencies each unit has
	inDegree := make(map[string]int)
	// adjacency list: unit -> units that depend on it
	dependents := make(map[string][]string)

	// Initialize in-degree for all units
	for unitName := range r.units {
		inDegree[unitName] = 0
	}

	// Build the graph
	for unitName, unit := range r.units {
		for _, depName := range unit.Uses {
			normalizedDep := ident.Normalize(depName)

			// Check if the dependency exists
			if _, exists := r.units[normalizedDep]; !exists {
				return nil, fmt.Errorf("unit '%s' depends on '%s', which is not loaded", unitName, depName)
			}

			// Add edge: depName -> unitName (unitName depends on depName)
			dependents[normalizedDep] = append(dependents[normalizedDep], unitName)
			inDegree[unitName]++
		}
	}

	// Kahn's algorithm: start with units that have no dependencies
	queue := make([]string, 0)
	for unitName, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, unitName)
		}
	}

	// Process units in topological order
	initOrder := make([]string, 0, len(r.units))

	for len(queue) > 0 {
		// Remove a unit with no dependencies
		current := queue[0]
		queue = queue[1:]
		// Append the unit's actual name (with original case), not the normalized key
		initOrder = append(initOrder, r.units[current].Name)

		// Reduce in-degree for all units that depend on current
		for _, dependent := range dependents[current] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				// This unit now has all its dependencies satisfied
				queue = append(queue, dependent)
			}
		}
	}

	// If we didn't process all units, there's a cycle
	if len(initOrder) != len(r.units) {
		// Find a unit involved in the cycle for error reporting
		remaining := make([]string, 0)
		for unitName, degree := range inDegree {
			if degree > 0 {
				remaining = append(remaining, unitName)
			}
		}
		return nil, fmt.Errorf("circular dependency detected among units: %s", strings.Join(remaining, ", "))
	}

	return initOrder, nil
}
