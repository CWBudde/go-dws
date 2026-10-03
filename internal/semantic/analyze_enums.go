package semantic

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Enum Analysis
// ============================================================================

// analyzeEnumDecl analyzes an enum type declaration
func (a *Analyzer) analyzeEnumDecl(decl *ast.EnumDecl) {
	defer func() {
		if decl != nil && decl.Name != nil {
			a.recordDeclaredType(decl, decl.Name.Value)
		}
	}()

	if decl == nil {
		return
	}

	enumName := decl.Name.Value

	// Check if enum is already declared
	// Use lowercase for case-insensitive duplicate check
	if a.hasType(enumName) {
		a.addError("%s", errors.FormatNameAlreadyExists(enumName, decl.Token.Pos.Line, decl.Token.Pos.Column))
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

	elements := make(map[string]*Symbol, len(decl.Values))
	a.enumElements[enumType] = elements
	if a.symbols.enumNamespaces == nil {
		a.symbols.enumNamespaces = make(map[*types.EnumType]map[string]*Symbol)
	}
	a.symbols.enumNamespaces[enumType] = elements
	// Register enum values and calculate ordinal values
	currentOrdinal := 0
	flagBitPosition := 0               // For flags enums, track the bit position (2^n)
	usedNames := make(map[string]bool) // Track used names to detect duplicates

	for _, enumValue := range decl.Values {
		valueName := enumValue.Name

		// Check for duplicate value names
		if usedNames[ident.Normalize(valueName)] {
			a.addError("%s", errors.FormatNameAlreadyExists(valueName, decl.Token.Pos.Line, decl.Token.Pos.Column))
			continue
		}
		usedNames[ident.Normalize(valueName)] = true

		// Determine ordinal value (explicit or implicit)
		var ordinalValue int
		if enumValue.ValueExpr != nil || enumValue.Value != nil {
			var err error
			ordinalValue, err = a.enumOrdinalValue(enumValue)
			if err != nil {
				a.addError("enum value '%s': %v", valueName, err)
				continue
			}

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

		// Register the enum value
		enumType.Values[valueName] = ordinalValue
		enumType.OrderedNames = append(enumType.OrderedNames, valueName)
		// Elements precede their enclosing type and share one identity between
		// the enum namespace and (for classic enums) the source local table.
		binding := ast.EnumElementBinding{
			EnumType: enumType, Name: valueName, Ordinal: ordinalValue,
			IsDeprecated: enumValue.IsDeprecated, DeprecationMessage: enumValue.DeprecatedMessage,
		}
		sym := &Symbol{
			Name: valueName, Type: enumType, Value: ordinalValue, IsConst: true, ReadOnly: true,
			SuppressUnusedWarning: true, DeclPosition: decl.Token.Pos, EnumElement: &binding,
			IsDeprecated: binding.IsDeprecated, DeprecationMessage: binding.DeprecationMessage,
		}
		elements[ident.Normalize(valueName)] = sym
		if !decl.Scoped {
			a.symbols.put(valueName, sym)
		}

	}

	a.registerTypeWithPos(enumName, enumType, decl.Token.Pos)

	// Register enum type name as an identifier
	// This allows the type name to be used as a runtime value in expressions
	// like High(TColor) or Low(TColor)
	a.symbols.DefineEnumTypeName(enumName, enumType, decl.Token.Pos)

	// Create implicit helper for scoped enum access (TColor.Red)
	// This enables accessing enum values via the type name while maintaining
	// backward compatibility with unscoped access (Red)
	a.createEnumScopedAccessHelper(enumName, enumType)
}

// enumOrdinalValue re-evaluates source expressions against semantic bindings;
// parser convenience ordinals cannot capture duplicate-preserving lookup.
func (a *Analyzer) enumOrdinalValue(value ast.EnumValue) (int, error) {
	if value.ValueExpr == nil {
		return *value.Value, nil
	}
	a.analyzeExpression(value.ValueExpr)
	constant, err := a.evaluateConstant(value.ValueExpr)
	if err != nil {
		return 0, fmt.Errorf("must be constant: %w", err)
	}
	switch v := constant.(type) {
	case int:
		return v, nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case string:
		runes := []rune(v)
		if len(runes) == 1 {
			return int(runes[0]), nil
		}
	}
	return 0, fmt.Errorf("must be ordinal")
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
		helper.ClassConsts[ident.Normalize(valueName)] = ordinalValue
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
	a.registerHelper(ident.Normalize(enumType.String()), helper)
}
