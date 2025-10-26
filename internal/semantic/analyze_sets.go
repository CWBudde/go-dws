package semantic

import (
	"github.com/cwbudde/go-dws/ast"
	"github.com/cwbudde/go-dws/types"
)

// ============================================================================
// Set Analysis (Task 8.99-8.104)
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
func (a *Analyzer) analyzeSetLiteralWithContext(lit *ast.SetLiteral, expectedType types.Type) types.Type {
	if lit == nil {
		return nil
	}

	// If we have an expected type, it should be a SetType
	var expectedSetType *types.SetType
	if expectedType != nil {
		var ok bool
		expectedSetType, ok = expectedType.(*types.SetType)
		if !ok {
			a.addError("set literal cannot be assigned to non-set type %s at %s",
				expectedType.String(), lit.Token.Pos.String())
			return nil
		}
	}

	// Empty set literal
	if len(lit.Elements) == 0 {
		if expectedSetType != nil {
			return expectedSetType
		}
		// Empty set without context - cannot infer type
		a.addError("cannot infer type for empty set literal at %s", lit.Token.Pos.String())
		return nil
	}

	// Analyze all elements and check they are of the same enum type
	var elementEnumType *types.EnumType
	for i, elem := range lit.Elements {
		elemType := a.analyzeExpression(elem)
		if elemType == nil {
			// Error already reported
			continue
		}

		// Element must be an enum type
		enumType, ok := elemType.(*types.EnumType)
		if !ok {
			a.addError("set element must be an enum value, got %s at %s",
				elemType.String(), elem.Pos().String())
			continue
		}

		// First element determines the enum type
		if i == 0 {
			elementEnumType = enumType
		} else {
			// All elements must be of the same enum type
			if !enumType.Equals(elementEnumType) {
				a.addError("type mismatch in set literal: expected %s, got %s at %s",
					elementEnumType.String(), enumType.String(), elem.Pos().String())
			}
		}
	}

	if elementEnumType == nil {
		// All elements had errors
		return nil
	}

	// If we have an expected set type, verify the element type matches
	if expectedSetType != nil {
		if !elementEnumType.Equals(expectedSetType.ElementType) {
			a.addError("type mismatch in set literal: expected set of %s, got set of %s at %s",
				expectedSetType.ElementType.String(), elementEnumType.String(), lit.Token.Pos.String())
			return expectedSetType // Return expected type to continue analysis
		}
		return expectedSetType
	}

	// Create and return a new set type based on inferred element type
	return types.NewSetType(elementEnumType)
}
