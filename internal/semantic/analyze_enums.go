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
		Scoped:       decl.Scoped,
		Flags:        decl.Flags,
	}

	// Register enum values and calculate ordinal values
	currentOrdinal := 0
	flagBitPosition := 0               // For flags enums, track the bit position (2^n)
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
		var ordinalValue int
		if enumValue.Value != nil {
			// Explicit value provided
			ordinalValue = *enumValue.Value
			if decl.Flags {
				// For flags, update bit position based on explicit value
				// Find the bit position of the explicit value
				for bitPos := 0; bitPos < 64; bitPos++ {
					if (1 << bitPos) == ordinalValue {
						flagBitPosition = bitPos + 1
						break
					}
				}
			} else {
				// For regular enums, update current ordinal
				currentOrdinal = ordinalValue + 1
			}
		} else {
			// Implicit value
			if decl.Flags {
				// Flags use power-of-2 values: 1, 2, 4, 8, 16, ...
				ordinalValue = 1 << flagBitPosition
				flagBitPosition++
			} else {
				// Regular enums use sequential values
				ordinalValue = currentOrdinal
				currentOrdinal++
			}
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
	}

	// Register the enum type (use lowercase key for case-insensitive lookup)
	a.enums[strings.ToLower(enumName)] = enumType

	// Register each enum value as a constant in the symbol table
	// For scoped enums (enum/flags keyword), skip global registration -
	// values are only accessible via qualified access (Type.Value)
	if !decl.Scoped {
		for valueName, ordinalValue := range enumType.Values {
			// Store enum values as constants with the enum type and ordinal value
			// This allows type checking: var color: TColor := Red;
			// and const array initialization: const arr = (Red, Green, Blue);
			a.symbols.DefineConst(valueName, enumType, ordinalValue)
		}
	}

	// Register enum type name as an identifier
	// This allows the type name to be used as a runtime value in expressions
	// like High(TColor) or Low(TColor)
	a.symbols.Define(enumName, enumType)

	// Create implicit helper for scoped enum access (TColor.Red)
	// This enables accessing enum values via the type name while maintaining
	// backward compatibility with unscoped access (Red)
	a.createEnumScopedAccessHelper(enumName, enumType)
}

// createEnumScopedAccessHelper creates an implicit helper for an enum type
// that allows scoped access to enum values (e.g., TColor.Red).
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

	// Add Low and High as class constants
	// Low returns the minimum ordinal value
	// High returns the maximum ordinal value
	lowValue := enumType.Low()
	highValue := enumType.High()
	helper.ClassConsts["low"] = lowValue
	helper.ClassConsts["high"] = highValue

	// Also add Low and High as methods so they can be called with parentheses
	// e.g., MyEnum.Low() or MyEnum.High()
	lowMethod := &types.FunctionType{
		Parameters:    []types.Type{},
		ReturnType:    types.INTEGER,
		DefaultValues: nil,
	}
	highMethod := &types.FunctionType{
		Parameters:    []types.Type{},
		ReturnType:    types.INTEGER,
		DefaultValues: nil,
	}
	helper.Methods["low"] = lowMethod
	helper.Methods["high"] = highMethod

	// Add ByName method for string-to-enum conversion
	// e.g., MyEnum.ByName('a') returns the ordinal value of 'a'
	byNameMethod := &types.FunctionType{
		Parameters:    []types.Type{types.STRING},
		ReturnType:    types.INTEGER,
		DefaultValues: nil,
	}
	helper.Methods["byname"] = byNameMethod

	// Register the helper for this enum type
	targetTypeName := strings.ToLower(enumType.String())
	if a.helpers[targetTypeName] == nil {
		a.helpers[targetTypeName] = make([]*types.HelperType, 0)
	}
	a.helpers[targetTypeName] = append(a.helpers[targetTypeName], helper)
}
