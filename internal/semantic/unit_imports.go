package semantic

import (
	"fmt"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// SetSemanticInfo shares the compilation's metadata across unit analyzers.
func (a *Analyzer) SetSemanticInfo(info *ast.SemanticInfo) { a.semanticInfo = info }

// GetUnitSymbols returns only the public declarations of an analyzed unit.
func (a *Analyzer) GetUnitSymbols(name string) *SymbolTable {
	return a.unitSymbols[ident.Normalize(name)]
}

// ProvideUnitSymbols makes analyzed unit metadata available for a later uses
// clause without making its symbols or types visible in the source scope.
func (a *Analyzer) ProvideUnitSymbols(name string, symbols *SymbolTable) error {
	if symbols == nil {
		return fmt.Errorf("unit '%s' has no analyzed symbols", name)
	}
	a.availableUnitSymbols[ident.Normalize(name)] = symbols
	return nil
}

func (a *Analyzer) activateSourceUnit(name string) {
	a.symbols.registerUnitEntry(name)
	if symbols, ok := a.availableUnitSymbols[ident.Normalize(name)]; ok {
		if err := a.ImportUnitSymbols(name, symbols); err != nil {
			a.addError("%v", err)
		}
	}
}

// ImportUnitSymbols makes a unit's exported declarations available to a program.
func (a *Analyzer) ImportUnitSymbols(name string, symbols *SymbolTable) error {
	if symbols == nil {
		return fmt.Errorf("unit '%s' has no analyzed symbols", name)
	}
	a.unitSymbols[ident.Normalize(name)] = symbols
	// Imported declarations remain in their own lookup table: their names must
	// not perturb the program's local duplicate search.
	// Repeated uses repositions a parent rather than duplicating it.
	parents := []*SymbolTable{symbols}
	for _, parent := range a.symbols.importedParents {
		if parent != symbols {
			parents = append(parents, parent)
		}
	}
	a.symbols.importedParents = parents
	a.typeRegistry.activeImports[ident.Normalize(name)] = true
	for enumType, elements := range symbols.enumNamespaces {
		a.enumElements[enumType] = elements
	}
	for typeName, typ := range symbols.exportedTypes {
		for _, importedName := range []string{typeName, name + "." + typeName} {
			if existing, ok := a.typeRegistry.Resolve(importedName); ok && existing == typ {
				continue
			}
			if err := a.typeRegistry.registerImported(importedName, typ, name); err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *Analyzer) importUnitUses(section *ast.BlockStatement, available map[string]*SymbolTable, imported map[string]string) error {
	if section == nil {
		return nil
	}
	for _, stmt := range section.Statements {
		uses, ok := stmt.(*ast.UsesClause)
		if !ok {
			continue
		}
		for _, name := range uses.Units {
			symbols, ok := available[ident.Normalize(name.Value)]
			if !ok {
				return fmt.Errorf("unit '%s' not found (required by uses clause)", name.Value)
			}
			var conflict error
			symbols.symbols.Range(func(key string, symbol *Symbol) bool {
				key = ident.Normalize(key)
				if previous, exists := imported[key]; exists && !ident.Equal(previous, name.Value) {
					previousSymbol, _ := available[ident.Normalize(previous)].symbols.Get(key)
					if symbol.EnumElement != nil && previousSymbol != nil && previousSymbol.EnumElement != nil {
						imported[key] = name.Value
						return true
					}
					conflict = fmt.Errorf("symbol conflict: '%s' is exported by both '%s' and '%s'", symbol.Name, previous, name.Value)
					return false
				}
				imported[key] = name.Value
				return true
			})
			if conflict != nil {
				return conflict
			}
			a.symbols.registerUnitEntry(name.Value)
			if err := a.ImportUnitSymbols(name.Value, symbols); err != nil {
				return err
			}
		}
	}
	return nil
}

// importedUnitNamespace recognizes an active unit only when the receiver's
// selected lexical identity is that namespace, not a parameter/local/type.
func (a *Analyzer) importedUnitNamespace(name string) (*SymbolTable, bool) {
	unit, active := a.unitSymbols[ident.Normalize(name)]
	if !active {
		return nil, false
	}
	if sym, found := a.symbols.resolveIdentity(name, true); found && !sym.isUnitName {
		return nil, false
	}
	return unit, true
}
