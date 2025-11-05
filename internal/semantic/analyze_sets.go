package semantic

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Set Analysis
// ============================================================================

// analyzeSetDecl analyzes a set type declaration
// Task 8.99: Register set types in symbol table
func (a *Analyzer) analyzeSetDecl(decl *ast.SetDecl) {
	if decl == nil {
		return
	}

	setName := decl.Name.Value

	// Check if set type is already declared
	if _, exists := a.sets[setName]; exists {
		a.addError("set type '%s' already declared at %s", setName, decl.Token.Pos.String())
		return
	}

	// Resolve the element type (must be an enum type)
	// Task 8.100: Validate set element types (must be enum or small integer range)
	elementTypeName := decl.ElementType.Name

	// First check if it's an enum type
	enumType, exists := a.enums[elementTypeName]
	if !exists {
		a.addError("unknown type '%s' at %s", elementTypeName, decl.Token.Pos.String())
		return
	}

	// Create the set type
	setType := types.NewSetType(enumType)

	// Task 8.99: Register the set type
	a.sets[setName] = setType
}

// analyzeSetLiteralWithContext analyzes a set literal expression with optional type context
// Task 8.101: Type-check set literals (elements match set's element type)
