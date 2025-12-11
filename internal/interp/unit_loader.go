package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/units"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// LoadUnit loads a DWScript unit by name, using the interpreter's unit registry.
// The unit file is searched in the provided search paths (or registry's default paths if nil).
// After loading, the unit's interface symbols are processed but not yet imported into the
// current environment - use ImportUnitSymbols() to make them available.
//
// This function:
//  1. Delegates to the unit registry to find and parse the unit file
//  2. Checks for circular dependencies
//  3. Recursively loads any units that this unit depends on (via uses clauses)
//  4. Returns the loaded unit or an error
//
// The unit is cached in the registry, so subsequent LoadUnit calls for the same
// unit name return the cached instance.
//
// Example:
//
//	unit, err := interp.LoadUnit("MathUtils", []string{"./lib"})
//	if err != nil {
//	    // Handle error
//	}
//	// Unit is loaded but symbols not imported yet
//	err = interp.ImportUnitSymbols(unit)
//	// Now MathUtils functions are available
func (i *Interpreter) LoadUnit(name string, searchPaths []string) (*units.Unit, error) {
	// Ensure unit registry is initialized
	if i.unitRegistry == nil {
		return nil, fmt.Errorf("unit registry not initialized - call SetUnitRegistry first")
	}

	// Delegate to the registry to handle file search, parsing, and dependency loading
	unit, err := i.unitRegistry.LoadUnit(name, searchPaths)
	if err != nil {
		return nil, fmt.Errorf("failed to load unit '%s': %w", name, err)
	}

	// Track that this unit has been loaded (for initialization/finalization order)
	i.trackLoadedUnit(name)

	return unit, nil
}

// SetUnitRegistry sets the unit registry for this interpreter.
// The registry manages loaded units and handles dependency resolution.
// This must be called before LoadUnit can be used.
//
// Example:
//
//	registry := units.NewUnitRegistry([]string{"./lib", "./units"})
//	interp.SetUnitRegistry(registry)
func (i *Interpreter) SetUnitRegistry(registry *units.UnitRegistry) {
	i.unitRegistry = registry
	// Phase 3.5.1: Also update the evaluator's unit registry
	// (once we move unit loading to evaluator, this will be the only place it's set)
	// For now, we keep both in sync during the migration
	if i.evaluatorInstance != nil {
		i.evaluatorInstance.SetUnitRegistry(registry)
	}
}

// SetSource sets the source code and filename for enhanced error messages.
// Allows runtime errors to display source code snippets.
// Task 3.3.4: Now only updates evaluator's config (Interpreter no longer stores these fields).
func (i *Interpreter) SetSource(source, filename string) {
	if i.evaluatorInstance != nil {
		cfg := i.evaluatorInstance.Config()
		cfg.SourceCode = source
		cfg.SourceFile = filename
	}
}

// GetUnitRegistry returns the interpreter's unit registry.
// Returns nil if no registry has been set.
func (i *Interpreter) GetUnitRegistry() *units.UnitRegistry {
	return i.unitRegistry
}

// trackLoadedUnit records that a unit has been loaded.
// This maintains the load order for proper initialization/finalization sequencing.
func (i *Interpreter) trackLoadedUnit(name string) {
	// Check if already tracked (avoid duplicates)
	for _, loaded := range i.loadedUnits {
		if loaded == name {
			return
		}
	}
	i.loadedUnits = append(i.loadedUnits, name)
}

// IsUnitLoaded checks if a unit has been loaded by name.
// The check is case-insensitive (as DWScript is case-insensitive).
func (i *Interpreter) IsUnitLoaded(name string) bool {
	if i.unitRegistry == nil {
		return false
	}
	_, exists := i.unitRegistry.GetUnit(name)
	return exists
}

// ListLoadedUnits returns the names of all loaded units in load order.
func (i *Interpreter) ListLoadedUnits() []string {
	result := make([]string, len(i.loadedUnits))
	copy(result, i.loadedUnits)
	return result
}

// InitializeUnits executes initialization blocks for all loaded units in dependency order.
// Units with no dependencies are initialized first, followed by units that depend on them.
// Each unit is initialized exactly once, even if this method is called multiple times.
//
// This should be called after all units are loaded but before the main program executes.
//
// Returns an error if:
//   - A circular dependency is detected
//   - A unit's initialization block encounters a runtime error
//   - A dependency is missing
//
// Example initialization order:
//
//	If Unit C uses Unit B, and Unit B uses Unit A:
//	  1. Unit A initialized
//	  2. Unit B initialized
//	  3. Unit C initialized
func (i *Interpreter) InitializeUnits() error {
	if i.unitRegistry == nil {
		// No units to initialize
		return nil
	}

	// Compute the initialization order using topological sort
	order, err := i.unitRegistry.ComputeInitializationOrder()
	if err != nil {
		return fmt.Errorf("failed to compute initialization order: %w", err)
	}

	// Initialize each unit in dependency order
	for _, unitName := range order {
		// Skip if already initialized
		if i.initializedUnits[unitName] {
			continue
		}

		// Get the unit from the registry
		unit, exists := i.unitRegistry.GetUnit(unitName)
		if !exists {
			return fmt.Errorf("unit '%s' not found in registry", unitName)
		}

		// Execute the initialization section if it exists
		if unit.InitializationSection != nil {
			// Evaluate the initialization block
			result := i.Eval(unit.InitializationSection)

			// Check if initialization raised an exception
			if i.exception != nil {
				return fmt.Errorf("exception in initialization of unit '%s': %v", unitName, i.exception)
			}

			// Check for errors (though Eval typically returns error values, not Go errors)
			if errVal, ok := result.(*ErrorValue); ok {
				return fmt.Errorf("error initializing unit '%s': %s", unitName, errVal.Message)
			}
		}

		// Mark as initialized
		i.initializedUnits[unitName] = true
	}

	return nil
}

// FinalizeUnits executes finalization blocks for all loaded units in reverse dependency order.
// Units are finalized in the opposite order of initialization (LIFO - Last In, First Out).
// This ensures that units are cleaned up in the reverse order they were set up.
//
// This should be called at program exit, after the main program has finished executing.
//
// Errors during finalization are collected but do not stop the finalization process.
// All units will attempt to finalize even if earlier units encounter errors.
//
// Returns the first error encountered during finalization, or nil if all succeeded.
//
// Example finalization order:
//
//	If initialization order was [A, B, C], finalization order is [C, B, A]
func (i *Interpreter) FinalizeUnits() error {
	if i.unitRegistry == nil {
		// No units to finalize
		return nil
	}

	// Get initialization order
	order, err := i.unitRegistry.ComputeInitializationOrder()
	if err != nil {
		// If we can't compute order, finalize in reverse load order
		order = i.loadedUnits
	}

	var firstError error

	// Finalize in reverse order (LIFO)
	for idx := len(order) - 1; idx >= 0; idx-- {
		unitName := order[idx]

		// Get the unit from the registry
		unit, exists := i.unitRegistry.GetUnit(unitName)
		if !exists {
			// Unit was unloaded or not found - skip
			continue
		}

		// Execute the finalization section if it exists
		if unit.FinalizationSection != nil {
			// Evaluate the finalization block
			result := i.Eval(unit.FinalizationSection)

			// Capture errors but continue finalizing other units
			if i.exception != nil {
				if firstError == nil {
					firstError = fmt.Errorf("exception in finalization of unit '%s': %v", unitName, i.exception)
				}
				// Clear the exception to allow other finalizations to proceed
				i.exception = nil
			}

			if errVal, ok := result.(*ErrorValue); ok {
				if firstError == nil {
					firstError = fmt.Errorf("error finalizing unit '%s': %s", unitName, errVal.Message)
				}
			}
		}
	}

	return firstError
}

// ImportUnitSymbols imports exported symbols from a unit into the current environment.
// This makes the unit's interface section declarations (functions, types, constants, etc.)
// available in the current scope without requiring qualified names.
//
// Symbol conflicts are handled by keeping the first definition (first import wins).
// Subsequent imports of the same symbol name are silently ignored.
//
// Example:
//
//	unit MathUtils:
//	  function Add(x, y: Integer): Integer;
//
//	After ImportUnitSymbols(mathUnit):
//	  Add(3, 5)  // Can call without "MathUtils." prefix
//
// Returns an error if the unit is not properly loaded or if symbol extraction fails.
func (i *Interpreter) ImportUnitSymbols(unit *units.Unit) error {
	if unit == nil {
		return fmt.Errorf("cannot import symbols from nil unit")
	}

	// Process the interface section to extract exported symbols
	// The interface section contains declarations that should be made available
	if unit.InterfaceSection == nil {
		// Unit has no interface section - nothing to import
		return nil
	}

	// First, evaluate interface section declarations (function signatures, types, etc.)
	for _, stmt := range unit.InterfaceSection.Statements {
		// Skip uses clauses - they're handled during unit loading
		if _, ok := stmt.(*ast.UsesClause); ok {
			continue
		}

		// Process the declaration
		_ = i.Eval(stmt)

		// Check for errors during symbol import
		if i.exception != nil {
			return fmt.Errorf("exception while importing symbols from unit '%s': %v", unit.Name, i.exception)
		}
	}

	// Now evaluate the implementation section to get function bodies
	// The implementation section contains the actual function implementations
	// that correspond to the declarations in the interface section
	if unit.ImplementationSection != nil {
		for _, stmt := range unit.ImplementationSection.Statements {
			// Skip uses clauses
			if _, ok := stmt.(*ast.UsesClause); ok {
				continue
			}

			// Process implementation (function bodies, private functions, etc.)
			_ = i.Eval(stmt)

			// Check for errors
			if i.exception != nil {
				return fmt.Errorf("exception while importing implementations from unit '%s': %v", unit.Name, i.exception)
			}
		}
	}

	return nil
}

// ResolveQualifiedFunction resolves a qualified function identifier (UnitName.FunctionName)
// to the function declaration. This allows calling functions from a unit using the unit prefix.
//
// Example:
//
//	fn, err := interp.ResolveQualifiedFunction("MathUtils", "Add")
//	// fn is the *ast.FunctionDecl for Add from MathUtils
//
// Returns the function declaration and nil on success, or nil and an error if:
//   - The unit is not loaded
//   - The function is not found in the unit's interface
//   - The unit has no exported functions
func (i *Interpreter) ResolveQualifiedFunction(unitName, functionName string) (*ast.FunctionDecl, error) {
	if i.unitRegistry == nil {
		return nil, fmt.Errorf("unit registry not initialized")
	}

	// Get the unit from the registry
	_, exists := i.unitRegistry.GetUnit(unitName)
	if !exists {
		return nil, fmt.Errorf("unit '%s' not loaded", unitName)
	}

	// Look up the function in the global function registry
	// Note: The current implementation stores all functions globally.
	// TODO: This needs to be enhanced once units properly maintain their own symbol tables
	// (see tasks 9.108-9.110 where unit parsing is improved)
	// For now, we assume the function was imported and is available globally.
	// DWScript is case-insensitive, so normalize the function name
	if overloads, ok := i.functions[ident.Normalize(functionName)]; ok && len(overloads) > 0 {
		// TODO: Verify this function actually belongs to this unit once we have proper
		fn := overloads[0]
		// unit-scoped symbol tables
		return fn, nil
	}

	return nil, fmt.Errorf("function '%s' not found in unit '%s'", functionName, unitName)
}

// ResolveQualifiedVariable resolves a qualified variable/constant identifier
// (UnitName.VariableName) to its value.
//
// Returns the value and nil on success, or nil and an error if not found.
func (i *Interpreter) ResolveQualifiedVariable(unitName, variableName string) (Value, error) {
	if i.unitRegistry == nil {
		return nil, fmt.Errorf("unit registry not initialized")
	}

	// Get the unit from the registry
	_, exists := i.unitRegistry.GetUnit(unitName)
	if !exists {
		return nil, fmt.Errorf("unit '%s' not loaded", unitName)
	}

	// Try to find in the environment (for constants, variables)
	if val, ok := i.Env().Get(variableName); ok {
		return val, nil
	}

	return nil, fmt.Errorf("variable '%s' not found in unit '%s'", variableName, unitName)
}
