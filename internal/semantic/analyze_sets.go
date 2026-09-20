package semantic

import (
	"sort"

	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Set Analysis
// ============================================================================

// analyzeSetDecl analyzes a set type declaration
func (a *Analyzer) analyzeSetDecl(decl *ast.SetDecl) {
	defer func() {
		if decl != nil && decl.Name != nil {
			a.recordDeclaredType(decl, decl.Name.Value)
		}
	}()

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

// setTypeDiagnosticName renders a set type the way DWScript names it in a
// diagnostic: by the symbol the program declared it under, so
// `type TMySet = set of TMyEnum` is reported as "TMySet" and not by its
// structure (SetOfFail/invalid_method). An inline set (`var t : set of (a, b)`)
// has no declared name and keeps the structural spelling.
func (a *Analyzer) setTypeDiagnosticName(t types.Type) string {
	if t == nil {
		return semanticTypeNameForDiagnostic(t)
	}
	if _, isSet := types.GetUnderlyingType(t).(*types.SetType); !isSet {
		return semanticTypeNameForDiagnostic(t)
	}
	names := append([]string(nil), a.typeRegistry.TypesByKind("SET")...)
	sort.Strings(names)
	for _, name := range names {
		if declared, found := a.typeRegistry.Resolve(name); found && declared == t {
			return name
		}
	}
	return semanticTypeNameForDiagnostic(t)
}
