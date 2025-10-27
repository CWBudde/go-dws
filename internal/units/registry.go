package units

import (
	"fmt"
	"os"
	"strings"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// UnitRegistry manages loaded units and handles unit dependencies.
// It caches compiled units to avoid reloading and detects circular dependencies.
type UnitRegistry struct {
	// units maps normalized unit names to loaded Unit instances
	units map[string]*Unit

	// loading tracks units currently being loaded to detect circular dependencies
	loading map[string]bool

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
	}
}

// RegisterUnit registers a unit in the registry.
// Returns an error if a unit with the same name (case-insensitive) is already registered.
func (r *UnitRegistry) RegisterUnit(name string, unit *Unit) error {
	normalized := strings.ToLower(name)

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
	normalized := strings.ToLower(name)
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
	normalized := strings.ToLower(name)

	// Check if already loaded (cached)
	if unit, exists := r.units[normalized]; exists {
		return unit, nil
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
	_ = p.ParseProgram() // TODO: Use program AST in tasks 9.108-9.110

	// Check for parsing errors
	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("parse errors in unit '%s': %s", name, strings.Join(p.Errors(), "; "))
	}

	// Create the unit
	// TODO: In later tasks (9.108-9.110), we'll parse proper unit declarations
	// For now, we create a basic unit structure
	unit := NewUnit(name, filePath)

	// TODO: Extract interface/implementation sections from parsed AST
	// TODO: Extract uses clauses and load dependencies recursively
	// TODO: Build symbol table from interface section

	// Register the unit
	if err := r.RegisterUnit(name, unit); err != nil {
		return nil, err
	}

	return unit, nil
}

// UnregisterUnit removes a unit from the registry.
// This is primarily useful for testing or when reloading a unit.
func (r *UnitRegistry) UnregisterUnit(name string) {
	normalized := strings.ToLower(name)
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
			normalizedDep := strings.ToLower(depName)

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
