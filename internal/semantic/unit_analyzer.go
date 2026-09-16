package semantic

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

// warnUnitNameFileMismatch emits DWScript's warning when the declared unit
// name does not match the source file's base name.
func (a *Analyzer) warnUnitNameFileMismatch(unit *ast.UnitDeclaration) {
	if unit == nil || unit.Name == nil || a.sourceFile == "" || a.parseHadErrors {
		return
	}
	base := filepath.Base(a.sourceFile)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	if !ident.Equal(unit.Name.Value, base) {
		a.addWarning("Unit name does not match file name [line: %d, column: %d]",
			unit.Name.Token.Pos.Line, unit.Name.Token.Pos.Column)
	}
}

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

	a.warnUnitNameFileMismatch(unit)
	previousInUnit := a.inUnitDecl
	a.inUnitDecl = true
	defer func() { a.inUnitDecl = previousInUnit }()

	normalizedUnits := make(map[string]*SymbolTable, len(availableUnits))
	for name, symbols := range availableUnits {
		normalizedUnits[ident.Normalize(name)] = symbols
		a.unitSymbols[ident.Normalize(name)] = symbols
	}
	imported := make(map[string]string)
	if err := a.importUnitUses(unit.InterfaceSection, normalizedUnits, imported); err != nil {
		return err
	}
	publicSection := unit.InterfaceSection
	implementationSection := unit.ImplementationSection
	// The parser places section-less declarations in an implicit implementation
	// block. These declarations are public, including their executable bodies.
	// A leading uses clause has its own synthetic interface block; explicit
	// implementation sections remain private even without an interface section.
	if (publicSection == nil || publicSection.Token.Type == token.USES) &&
		implementationSection != nil && implementationSection.Token.Type != token.IMPLEMENTATION {
		publicSection = implementationSection
		implementationSection = nil
		if err := a.importUnitUses(publicSection, normalizedUnits, imported); err != nil {
			return err
		}
	}

	// An enclosed scope separates the public API from builtins and dependencies.
	a.symbols = NewEnclosedSymbolTable(a.symbols)
	exports := a.symbols
	beforeTypes := a.typeRegistry.AllTypes()
	interfaceFunctions := make(map[string][]*ast.FunctionDecl)
	var bodies []*ast.FunctionDecl
	if publicSection != nil {
		for _, stmt := range publicSection.Statements {
			if decl, ok := stmt.(*ast.FunctionDecl); ok && decl.ClassName == nil {
				if decl.Name == nil {
					a.addError("function declaration missing name")
					continue
				}
				name := ident.Normalize(decl.Name.Value)
				interfaceFunctions[name] = append(interfaceFunctions[name], decl)
				a.registerFunctionSignature(decl)
				if decl.Body != nil {
					bodies = append(bodies, decl)
				}
			} else {
				a.analyzeStatement(stmt)
			}
		}
	}
	exports.exportedTypes = make(map[string]types.Type)
	for name, typ := range a.typeRegistry.AllTypes() {
		if _, existed := beforeTypes[name]; !existed {
			exports.exportedTypes[name] = typ
		}
	}
	publicSymbols := NewSymbolTable()
	publicSymbols.exportedTypes = exports.exportedTypes
	exports.symbols.Range(func(name string, symbol *Symbol) bool {
		publicSymbols.symbols.Set(name, symbol)
		return true
	})
	a.unitSymbols[ident.Normalize(unit.Name.Value)] = publicSymbols

	// Private declarations and implementation-only imports never become exports.
	a.symbols = NewEnclosedSymbolTable(exports)
	if err := a.importUnitUses(implementationSection, normalizedUnits, imported); err != nil {
		return err
	}
	implemented := make(map[*ast.FunctionDecl]bool)
	if implementationSection != nil {
		for _, stmt := range implementationSection.Statements {
			decl, ok := stmt.(*ast.FunctionDecl)
			if !ok || decl.ClassName != nil {
				a.analyzeStatement(stmt)
				continue
			}
			if decl.Name == nil {
				a.addError("function implementation missing name")
				continue
			}
			candidates := interfaceFunctions[ident.Normalize(decl.Name.Value)]
			matched := false
			for _, candidate := range candidates {
				if a.validateFunctionSignatureMatch(candidate, decl) == nil {
					implemented[candidate] = true
					matched = true
					break
				}
			}
			if len(candidates) > 0 && !matched {
				a.addError("implementation of '%s' doesn't match interface: %v", decl.Name.Value, a.validateFunctionSignatureMatch(candidates[0], decl))
				continue
			}
			if !matched {
				a.registerFunctionSignature(decl)
			}
			bodies = append(bodies, decl)
		}
	}
	for _, decl := range bodies {
		if decl.Body == nil {
			continue
		}
		funcType, err := a.buildFunctionType(decl)
		if err != nil {
			a.addError("invalid function signature for '%s': %v", decl.Name.Value, err)
			continue
		}
		returnType := funcType.ReturnType
		if returnType == nil {
			returnType = types.VOID
		}
		a.analyzeFunctionBody(decl, funcType.Parameters, returnType)
	}
	for _, candidates := range interfaceFunctions {
		for _, decl := range candidates {
			if !implemented[decl] && !decl.IsExternal && decl.Body == nil {
				a.addError("interface function '%s' has no implementation", decl.Name.Value)
			}
		}
	}
	for _, section := range []*ast.BlockStatement{unit.InitSection, unit.FinalSection} {
		if section != nil {
			for _, stmt := range section.Statements {
				a.analyzeStatement(stmt)
			}
		}
	}
	if a.hasActualErrors() {
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
	normalizedUnitName := ident.Normalize(unitName)

	// Look up the unit's symbol table
	unitSymbols, found := a.unitSymbols[normalizedUnitName]
	if !found {
		return nil, fmt.Errorf("unit '%s' not found or not imported", unitName)
	}

	// Look up the symbol within that unit (case-insensitive via ident.Map)
	symbol, found := unitSymbols.symbols.Get(symbolName)
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

		paramType, err := a.resolveTypeExpression(param.Type)
		if err != nil {
			return nil, fmt.Errorf("unknown type '%s' for parameter '%s': %v", getTypeExpressionName(param.Type), param.Name.Value, err)
		}

		funcType.Parameters = append(funcType.Parameters, paramType)
	}

	// Resolve return type
	if decl.ReturnType != nil {
		returnType, err := a.resolveTypeExpression(decl.ReturnType)
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
