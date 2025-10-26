package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Type Alias Support (Task 9.21)
// ============================================================================

// TypeAliasValue stores a type alias definition
type TypeAliasValue struct {
	Name        string
	AliasedType types.Type
}

func (tv *TypeAliasValue) Type() string {
	return "TYPE_ALIAS"
}

func (tv *TypeAliasValue) String() string {
	return fmt.Sprintf("type %s = %s", tv.Name, tv.AliasedType.String())
}

// evalTypeDeclaration evaluates a type declaration (Task 9.21)
// Handles type aliases: type TUserID = Integer;
func (i *Interpreter) evalTypeDeclaration(decl *ast.TypeDeclaration) Value {
	if decl == nil {
		return &ErrorValue{Message: "nil type declaration"}
	}

	// Handle type aliases
	if decl.IsAlias {
		// Resolve the aliased type
		aliasedType, err := i.resolveType(decl.AliasedType.Name)
		if err != nil {
			return &ErrorValue{Message: fmt.Sprintf("unknown type '%s' in type alias", decl.AliasedType.Name)}
		}

		// Create TypeAliasValue and register it
		typeAlias := &TypeAliasValue{
			Name:        decl.Name.Value,
			AliasedType: aliasedType,
		}

		// Store in environment with special prefix
		typeKey := "__type_alias_" + decl.Name.Value
		i.env.Define(typeKey, typeAlias)

		return &NilValue{}
	}

	// Non-alias type declarations (future)
	return &ErrorValue{Message: "non-alias type declarations not yet supported"}
}
