package semantic

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Set Analysis
// ============================================================================

// analyzeSetDecl analyzes a set type declaration
func (a *Analyzer) analyzeSetDecl(decl *ast.SetDecl) {
	if decl == nil {
		return
	}

	setName := decl.Name.Value

	// Check if set type is already declared
	// Use lowercase for case-insensitive duplicate check
	if _, exists := a.sets[strings.ToLower(setName)]; exists {
		a.addError("set type '%s' already declared at %s", setName, decl.Token.Pos.String())
		return
	}

	// Resolve the element type (must be an enum type)
	// Validate set element types (must be enum or small integer range)
	elementTypeName := getTypeExpressionName(decl.ElementType)

	// First check if it's an enum type
	// Use lowercase for case-insensitive lookup
	enumType, exists := a.enums[strings.ToLower(elementTypeName)]
	if !exists {
		a.addError("unknown type '%s' at %s", elementTypeName, decl.Token.Pos.String())
		return
	}

	// Create the set type
	setType := types.NewSetType(enumType)

	// Register the set type
	// Use lowercase key for case-insensitive lookup
	a.sets[strings.ToLower(setName)] = setType
}
