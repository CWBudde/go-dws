package semantic

import (
	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
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
	if a.hasType(setName) {
		a.addError("%s", errors.FormatNameAlreadyExists(setName, decl.Token.Pos.Line, decl.Token.Pos.Column))
		return
	}

	// Resolve the element type (must be an enum type)
	// Validate set element types (must be enum or small integer range)
	elementTypeName := getTypeExpressionName(decl.ElementType)

	// First check if it's an enum type
	enumType := a.getEnumType(elementTypeName)
	if enumType == nil {
		a.addError("unknown type '%s' at %s", elementTypeName, decl.Token.Pos.String())
		return
	}

	// Create the set type
	setType := types.NewSetType(enumType)

	// Register the set type
	// Use lowercase key for case-insensitive lookup
	a.registerTypeWithPos(setName, setType, decl.Token.Pos)
}
