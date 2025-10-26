package semantic

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Enum Analysis
// ============================================================================

// analyzeEnumDecl analyzes an enum type declaration
func (a *Analyzer) analyzeEnumDecl(decl *ast.EnumDecl) {
	if decl == nil {
		return
	}

	enumName := decl.Name.Value

	// Task 8.45: Check if enum is already declared
	if _, exists := a.enums[enumName]; exists {
		a.addError("enum type '%s' already declared at %s", enumName, decl.Token.Pos.String())
		return
	}

	// Create the enum type
	enumType := &types.EnumType{
		Name:         enumName,
		Values:       make(map[string]int),
		OrderedNames: make([]string, 0, len(decl.Values)),
	}

	// Task 8.44: Register enum values and calculate ordinal values
	// Task 8.45: Validate uniqueness and range
	currentOrdinal := 0
	usedValues := make(map[int]string) // Track used ordinal values to detect duplicates
	usedNames := make(map[string]bool) // Track used names to detect duplicates

	for _, enumValue := range decl.Values {
		valueName := enumValue.Name

		// Task 8.45: Check for duplicate value names
		if usedNames[valueName] {
			a.addError("duplicate enum value '%s' in enum '%s' at %s",
				valueName, enumName, decl.Token.Pos.String())
			continue
		}
		usedNames[valueName] = true

		// Determine ordinal value (explicit or implicit)
		ordinalValue := currentOrdinal
		if enumValue.Value != nil {
			ordinalValue = *enumValue.Value
		}

		// Task 8.45: Check for duplicate ordinal values
		if existingName, exists := usedValues[ordinalValue]; exists {
			a.addError("duplicate enum ordinal value %d in enum '%s' (values '%s' and '%s') at %s",
				ordinalValue, enumName, existingName, valueName, decl.Token.Pos.String())
			continue
		}
		usedValues[ordinalValue] = valueName

		// Register the enum value
		enumType.Values[valueName] = ordinalValue
		enumType.OrderedNames = append(enumType.OrderedNames, valueName)

		// Update current ordinal for next implicit value
		currentOrdinal = ordinalValue + 1
	}

	// Task 8.43: Register the enum type
	a.enums[enumName] = enumType

	// Task 8.44: Register each enum value as a constant in the symbol table
	for valueName, ordinalValue := range enumType.Values {
		// Store enum values as constants with the enum type
		// For now, we'll store them as the enum type itself
		// This allows type checking: var color: TColor := Red;
		_ = ordinalValue // We don't need the ordinal value for type checking
		a.symbols.Define(valueName, enumType)
	}
}
