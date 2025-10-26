package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/ast"
	"github.com/cwbudde/go-dws/types"
)

// ============================================================================
// Enum Declaration Evaluation (Task 8.48)
// ============================================================================

// evalEnumDeclaration evaluates an enum type declaration.
// It registers the enum type and its values in the interpreter's symbol table.
func (i *Interpreter) evalEnumDeclaration(decl *ast.EnumDecl) Value {
	if decl == nil {
		return &ErrorValue{Message: "nil enum declaration"}
	}

	enumName := decl.Name.Value

	// Build the enum type from the declaration
	enumValues := make(map[string]int)
	orderedNames := make([]string, 0, len(decl.Values))

	// Calculate ordinal values (explicit or implicit)
	currentOrdinal := 0
	for _, enumValue := range decl.Values {
		valueName := enumValue.Name

		// Determine ordinal value
		ordinalValue := currentOrdinal
		if enumValue.Value != nil {
			ordinalValue = *enumValue.Value
		}

		// Store the value
		enumValues[valueName] = ordinalValue
		orderedNames = append(orderedNames, valueName)

		// Next implicit value
		currentOrdinal = ordinalValue + 1
	}

	// Create the enum type
	enumType := types.NewEnumType(enumName, enumValues, orderedNames)

	// Store enum type in the interpreter's registry (for type resolution)
	if i.classes == nil {
		i.classes = make(map[string]*ClassInfo)
	}
	// Note: We don't have a dedicated enum registry in the interpreter yet,
	// so we'll use the environment to store enum types temporarily.
	// A better approach would be to add an 'enums' map to the Interpreter struct.

	// Register each enum value in the symbol table as a constant
	for valueName, ordinalValue := range enumValues {
		enumVal := &EnumValue{
			TypeName:     enumName,
			ValueName:    valueName,
			OrdinalValue: ordinalValue,
		}
		i.env.Define(valueName, enumVal)
	}

	// Store enum type metadata in environment with special key
	// This allows variable declarations to resolve the type
	enumTypeKey := "__enum_type_" + enumName
	i.env.Define(enumTypeKey, &EnumTypeValue{EnumType: enumType})

	return &NilValue{}
}

// EnumTypeValue is an internal value type used to store enum type metadata
// in the interpreter's environment.
type EnumTypeValue struct {
	EnumType *types.EnumType
}

// Type returns "ENUM_TYPE".
func (e *EnumTypeValue) Type() string {
	return "ENUM_TYPE"
}

// String returns the enum type name.
func (e *EnumTypeValue) String() string {
	return e.EnumType.Name
}

// ============================================================================
// Enum Literal Evaluation (Task 8.50)
// ============================================================================

// evalEnumLiteral evaluates an enum literal expression.
// Examples: Red, TColor.Green
func (i *Interpreter) evalEnumLiteral(literal *ast.EnumLiteral) Value {
	if literal == nil {
		return &ErrorValue{Message: "nil enum literal"}
	}

	// For scoped references (e.g., TColor.Red), the semantic analyzer
	// should have already validated that the enum type exists and the
	// value is valid. We just need to look up the value.

	valueName := literal.ValueName

	// Look up the value in the environment
	val, ok := i.env.Get(valueName)
	if !ok {
		return &ErrorValue{
			Message: fmt.Sprintf("undefined enum value '%s'", valueName),
		}
	}

	// Verify it's an enum value
	enumVal, ok := val.(*EnumValue)
	if !ok {
		return &ErrorValue{
			Message: fmt.Sprintf("'%s' is not an enum value (got %s)", valueName, val.Type()),
		}
	}

	// If it's a scoped reference, verify the type matches
	if literal.EnumName != "" && enumVal.TypeName != literal.EnumName {
		return &ErrorValue{
			Message: fmt.Sprintf("enum value '%s' does not belong to type '%s'", valueName, literal.EnumName),
		}
	}

	return enumVal
}
