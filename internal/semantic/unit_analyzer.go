package semantic

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// AnalyzeUnit performs semantic analysis on a unit declaration without dependencies.
// For units with uses clauses, use AnalyzeUnitWithDependencies instead.
// It validates that:
//  1. All interface declarations have matching implementations
//  2. Implementation signatures match interface signatures exactly
//  3. Types are valid and consistent
//
// The analyzed unit's exported symbols are added to the analyzer's symbol table.
func (a *Analyzer) AnalyzeUnit(unit *ast.UnitDeclaration) error {
	return a.AnalyzeUnitWithDependencies(unit, nil)
}

// AnalyzeUnitWithDependencies performs semantic analysis on a unit declaration with dependency resolution.
// The availableUnits map provides symbol tables for units that can be imported via uses clauses.
// Keys are unit names (case-insensitive).
//
// It validates that:
//  1. Uses clauses can be resolved (all imported units exist)
//  2. Imported symbols don't conflict
//  3. All interface declarations have matching implementations
//  4. Implementation signatures match interface signatures exactly
//  5. Types are valid and consistent
//
// The analyzed unit's exported symbols (and imported symbols) are added to the analyzer's symbol table.
func (a *Analyzer) AnalyzeUnitWithDependencies(unit *ast.UnitDeclaration, availableUnits map[string]*SymbolTable) error {
	if unit == nil {
		return fmt.Errorf("cannot analyze nil unit")
	}

	// Step 0: Store available units for qualified access (UnitName.Symbol)
	if availableUnits != nil {
		for unitName, unitSymbols := range availableUnits {
			a.unitSymbols[unitName] = unitSymbols
		}
	}

	// Step 1: Process uses clauses and import symbols from dependencies
	// Track which symbols are imported from which unit to detect conflicts
	importedSymbols := make(map[string]string) // symbol name -> source unit name

	if unit.InterfaceSection != nil {
		for _, stmt := range unit.InterfaceSection.Statements {
			if usesClause, ok := stmt.(*ast.UsesClause); ok {
				// Process each unit in the uses clause
				for _, unitIdent := range usesClause.Units {
					unitName := strings.ToLower(unitIdent.Value)

					// Look up the unit's symbols
					if availableUnits == nil {
						return fmt.Errorf("unit '%s' uses '%s', but no units are available", unit.Name.Value, unitIdent.Value)
					}

					unitSymbols, found := availableUnits[unitName]
					if !found {
						return fmt.Errorf("unit '%s' not found (required by uses clause)", unitIdent.Value)
					}

					// Import all symbols from the used unit
					for symbolName, symbol := range unitSymbols.symbols {
						// Check for conflicts
						if existingSource, exists := importedSymbols[symbolName]; exists {
							return fmt.Errorf("symbol conflict: '%s' is exported by both '%s' and '%s'",
								symbol.Name, existingSource, unitIdent.Value)
						}

						// Import the symbol
						a.symbols.symbols[symbolName] = symbol
						importedSymbols[symbolName] = unitIdent.Value
					}
				}
			}
		}
	}

	// Create a separate symbol table for the unit's interface (exported symbols)
	interfaceSymbols := NewSymbolTable()

	// Step 1: Analyze interface section and collect function signatures
	// Interface section contains declarations only (no implementations)
	interfaceFunctions := make(map[string]*ast.FunctionDecl)
	if unit.InterfaceSection != nil {
		for _, stmt := range unit.InterfaceSection.Statements {
			switch decl := stmt.(type) {
			case *ast.FunctionDecl:
				// Validate function declaration
				if decl.Name == nil {
					a.addError("function declaration missing name")
					continue
				}

				// Store function signature for later validation
				normalizedName := strings.ToLower(decl.Name.Value)
				interfaceFunctions[normalizedName] = decl

				// Build function type from parameters and return type
				funcType, err := a.buildFunctionType(decl)
				if err != nil {
					a.addError("invalid function signature for '%s': %v", decl.Name.Value, err)
					continue
				}

				// Add to interface symbol table (exported)
				interfaceSymbols.DefineFunction(decl.Name.Value, funcType)

			// TODO: Handle other declaration types (type declarations, constants, etc.)
			default:
				// For now, skip non-function declarations
			}
		}
	}

	// Step 2: Analyze implementation section and validate against interface
	implementedFunctions := make(map[string]bool)
	if unit.ImplementationSection != nil {
		for _, stmt := range unit.ImplementationSection.Statements {
			switch decl := stmt.(type) {
			case *ast.FunctionDecl:
				if decl.Name == nil {
					a.addError("function implementation missing name")
					continue
				}

				normalizedName := strings.ToLower(decl.Name.Value)

				// Check if this function was declared in the interface
				interfaceDecl, hasInterfaceDecl := interfaceFunctions[normalizedName]
				if hasInterfaceDecl {
					// Validate that signatures match
					if err := a.validateFunctionSignatureMatch(interfaceDecl, decl); err != nil {
						a.addError("implementation of '%s' doesn't match interface: %v", decl.Name.Value, err)
						continue
					}

					// Mark as implemented
					implementedFunctions[normalizedName] = true
				}

				// Analyze the function implementation
				// (For now, we just validate the structure; full body analysis comes later)
				if decl.Body != nil {
					// TODO: Analyze function body when ready
					// For now, just verify it's present
				}

			default:
				// Implementation-only declarations (not in interface)
			}
		}
	}

	// Step 3: Verify all interface functions have implementations
	for name, interfaceFunc := range interfaceFunctions {
		if !implementedFunctions[name] {
			a.addError("interface function '%s' has no implementation", interfaceFunc.Name.Value)
		}
	}

	// Step 4: Import interface symbols into the analyzer's symbol table
	// These become the unit's public API
	for name, symbol := range interfaceSymbols.symbols {
		a.symbols.symbols[name] = symbol
	}

	// Return accumulated errors
	if len(a.errors) > 0 {
		return &AnalysisError{Errors: a.errors}
	}

	return nil
}

// ResolveQualifiedSymbol resolves a qualified symbol reference like "UnitName.SymbolName".
// Returns the symbol if found, or an error if the unit or symbol doesn't exist.
//
// This enables disambiguation when multiple units export symbols with the same name:
//
//	Math.Add(1, 2)     // Use Add from Math unit
//	Strings.Add(a, b)  // Use Add from Strings unit
func (a *Analyzer) ResolveQualifiedSymbol(unitName, symbolName string) (*Symbol, error) {
	// Normalize unit name for case-insensitive lookup
	normalizedUnitName := strings.ToLower(unitName)

	// Look up the unit's symbol table
	unitSymbols, found := a.unitSymbols[normalizedUnitName]
	if !found {
		return nil, fmt.Errorf("unit '%s' not found or not imported", unitName)
	}

	// Look up the symbol within that unit (case-insensitive)
	normalizedSymbolName := strings.ToLower(symbolName)
	symbol, found := unitSymbols.symbols[normalizedSymbolName]
	if !found {
		return nil, fmt.Errorf("symbol '%s' not found in unit '%s'", symbolName, unitName)
	}

	return symbol, nil
}

// buildFunctionType constructs a FunctionType from a function declaration.
// It resolves parameter types and the return type.
func (a *Analyzer) buildFunctionType(decl *ast.FunctionDecl) (*types.FunctionType, error) {
	funcType := &types.FunctionType{
		Parameters: make([]types.Type, 0),
	}

	// Resolve parameter types
	for _, param := range decl.Parameters {
		if param.Type == nil {
			return nil, fmt.Errorf("parameter '%s' missing type", param.Name.Value)
		}

		paramType, err := a.resolveType(getTypeExpressionName(param.Type))
		if err != nil {
			return nil, fmt.Errorf("unknown type '%s' for parameter '%s': %v", getTypeExpressionName(param.Type), param.Name.Value, err)
		}

		funcType.Parameters = append(funcType.Parameters, paramType)
	}

	// Resolve return type
	if decl.ReturnType != nil {
		returnType, err := a.resolveType(getTypeExpressionName(decl.ReturnType))
		if err != nil {
			return nil, fmt.Errorf("unknown return type '%s': %v", getTypeExpressionName(decl.ReturnType), err)
		}
		funcType.ReturnType = returnType
	}

	return funcType, nil
}

// validateFunctionSignatureMatch checks that two function declarations have matching signatures.
// Used to verify interface declarations match their implementations.
func (a *Analyzer) validateFunctionSignatureMatch(interfaceDecl, implDecl *ast.FunctionDecl) error {
	// Check parameter count
	if len(interfaceDecl.Parameters) != len(implDecl.Parameters) {
		return fmt.Errorf("parameter count mismatch: interface has %d, implementation has %d",
			len(interfaceDecl.Parameters), len(implDecl.Parameters))
	}

	// Check each parameter type
	for i := 0; i < len(interfaceDecl.Parameters); i++ {
		interfaceParam := interfaceDecl.Parameters[i]
		implParam := implDecl.Parameters[i]

		// Compare types (case-insensitive)
		if !ident.Equal(getTypeExpressionName(interfaceParam.Type), getTypeExpressionName(implParam.Type)) {
			return fmt.Errorf("parameter %d type mismatch: interface has '%s', implementation has '%s'",
				i+1, getTypeExpressionName(interfaceParam.Type), getTypeExpressionName(implParam.Type))
		}
	}

	// Check return type
	if interfaceDecl.ReturnType != nil && implDecl.ReturnType != nil {
		if !ident.Equal(getTypeExpressionName(interfaceDecl.ReturnType), getTypeExpressionName(implDecl.ReturnType)) {
			return fmt.Errorf("return type mismatch: interface has '%s', implementation has '%s'",
				getTypeExpressionName(interfaceDecl.ReturnType), getTypeExpressionName(implDecl.ReturnType))
		}
	} else if interfaceDecl.ReturnType != nil || implDecl.ReturnType != nil {
		return fmt.Errorf("return type mismatch: one has return type, the other doesn't")
	}

	return nil
}
