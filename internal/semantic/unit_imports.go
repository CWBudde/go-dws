package semantic

import (
	"fmt"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

// SetSemanticInfo shares the compilation's metadata across unit analyzers.
func (a *Analyzer) SetSemanticInfo(info *ast.SemanticInfo) { a.semanticInfo = info }

// GetUnitSymbols returns only the public declarations of an analyzed unit.
func (a *Analyzer) GetUnitSymbols(name string) *SymbolTable {
	return a.unitSymbols[ident.Normalize(name)]
}

// ImportUnitSymbols makes a unit's exported declarations available to a program.
func (a *Analyzer) ImportUnitSymbols(name string, symbols *SymbolTable) error {
	if symbols == nil {
		return fmt.Errorf("unit '%s' has no analyzed symbols", name)
	}
	a.unitSymbols[ident.Normalize(name)] = symbols
	symbols.symbols.Range(func(name string, symbol *Symbol) bool {
		a.symbols.symbols.Set(name, symbol)
		return true
	})
	for typeName, typ := range symbols.exportedTypes {
		for _, importedName := range []string{typeName, name + "." + typeName} {
			if existing, ok := a.typeRegistry.Resolve(importedName); ok && existing == typ {
				continue
			}
			if err := a.typeRegistry.Register(importedName, typ, token.Position{}, 2); err != nil {
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
					conflict = fmt.Errorf("symbol conflict: '%s' is exported by both '%s' and '%s'", symbol.Name, previous, name.Value)
					return false
				}
				imported[key] = name.Value
				return true
			})
			if conflict != nil {
				return conflict
			}
			if err := a.ImportUnitSymbols(name.Value, symbols); err != nil {
				return err
			}
		}
	}
	return nil
}
