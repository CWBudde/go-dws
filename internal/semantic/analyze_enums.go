package semantic

import (
	"strings"

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

	// Check if enum is already declared
	// Use lowercase for case-insensitive duplicate check
	if _, exists := a.enums[strings.ToLower(enumName)]; exists {
		a.addError("enum type '%s' already declared at %s", enumName, decl.Token.Pos.String())
		return
	}

	// Create the enum type
	enumType := &types.EnumType{
		Name:         enumName,
		Values:       make(map[string]int),
		OrderedNames: make([]string, 0, len(decl.Values)),
	}

	// Register enum values and calculate ordinal values
	currentOrdinal := 0
	usedValues := make(map[int]string) // Track used ordinal values to detect duplicates
	usedNames := make(map[string]bool) // Track used names to detect duplicates

	for _, enumValue := range decl.Values {
		valueName := enumValue.Name

		// Check for duplicate value names
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

		// Check for duplicate ordinal values
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

	// Register the enum type (use lowercase key for case-insensitive lookup)
	a.enums[strings.ToLower(enumName)] = enumType

	// Register each enum value as a constant in the symbol table
	for valueName, ordinalValue := range enumType.Values {
		// Store enum values as constants with the enum type
		// For now, we'll store them as the enum type itself
		// This allows type checking: var color: TColor := Red;
		_ = ordinalValue // We don't need the ordinal value for type checking
		a.symbols.Define(valueName, enumType)
	}

	// Register enum type name as an identifier
	// This allows the type name to be used as a runtime value in expressions
	// like High(TColor) or Low(TColor)
	a.symbols.Define(enumName, enumType)

	// Task 9.54: Create implicit helper for scoped enum access (TColor.Red)
	// This enables accessing enum values via the type name while maintaining
	// backward compatibility with unscoped access (Red)
	a.createEnumScopedAccessHelper(enumName, enumType)
}

// createEnumScopedAccessHelper creates an implicit helper for an enum type
// that allows scoped access to enum values (e.g., TColor.Red).
// Task 9.54: Enable TE1.val1 syntax for enum values
func (a *Analyzer) createEnumScopedAccessHelper(enumName string, enumType *types.EnumType) {
	// Create a helper type for this specific enum
	helperName := "__" + enumName + "_ScopedAccessHelper"
	helper := types.NewHelperType(helperName, enumType, false)

	// Add each enum value as a class constant on the helper
	// This allows TColor.Red to resolve to the Red constant
	for valueName, ordinalValue := range enumType.Values {
		// Store the ordinal value as the constant value
		// Normalize to lowercase for case-insensitive lookup
		helper.ClassConsts[strings.ToLower(valueName)] = ordinalValue
	}

	// Register the helper for this enum type
	targetTypeName := enumType.String()
	if a.helpers[targetTypeName] == nil {
		a.helpers[targetTypeName] = make([]*types.HelperType, 0)
	}
	a.helpers[targetTypeName] = append(a.helpers[targetTypeName], helper)
}
