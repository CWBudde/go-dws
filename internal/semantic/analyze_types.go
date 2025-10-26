package semantic

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Type Declaration Analysis (Task 9.19)
// ============================================================================

// analyzeTypeDeclaration analyzes a type declaration statement
// Handles type aliases: type TUserID = Integer;
func (a *Analyzer) analyzeTypeDeclaration(decl *ast.TypeDeclaration) {
	if decl == nil {
		return
	}

	// Check if type name already exists
	if _, err := a.resolveType(decl.Name.Value); err == nil {
		a.addError("type '%s' already declared at %s", decl.Name.Value, decl.Token.Pos.String())
		return
	}

	// Handle type aliases
	if decl.IsAlias {
		// Resolve the aliased type
		aliasedType, err := a.resolveType(decl.AliasedType.Name)
		if err != nil {
			a.addError("unknown type '%s' in type alias at %s", decl.AliasedType.Name, decl.Token.Pos.String())
			return
		}

		// Create TypeAlias and register it
		typeAlias := &types.TypeAlias{
			Name:        decl.Name.Value,
			AliasedType: aliasedType,
		}

		a.typeAliases[decl.Name.Value] = typeAlias
	}
}
