package runtime

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/types"
)

// EnumValue represents an enumerated value in DWScript.
// Enums are named constants with associated ordinal values.
// Enables evaluator package to work with enum values directly.
//
// Example: type TColor = (Red, Green, Blue);
// - TypeName: "TColor"
// - ValueName: "Red" (for Red), "Green" (for Green), etc.
// - OrdinalValue: 0 (for Red), 1 (for Green), 2 (for Blue)
type EnumValue struct {
	TypeName     string // Enum type name (e.g., "TColor")
	ValueName    string // Enum value name (e.g., "Red")
	OrdinalValue int    // The ordinal value (e.g., 0 for Red if implicit)
}

// NewEnumValue constructs an enum runtime value from an ordinal.
// If the ordinal does not map to a declared enum member, a DWScript-style
// placeholder name is used so the ordinal can still flow through execution.
func NewEnumValue(typeName string, enumType *types.EnumType, ordinal int) *EnumValue {
	valueName := ""
	if enumType != nil {
		valueName = enumType.GetEnumName(ordinal)
	}
	if valueName == "" {
		valueName = fmt.Sprintf("$%d", ordinal)
	}

	return &EnumValue{
		TypeName:     typeName,
		ValueName:    valueName,
		OrdinalValue: ordinal,
	}
}

// EnumValueIndex returns the declaration-order index of an enum value.
// It first matches by canonical value name, then falls back to ordinal lookup.
func EnumValueIndex(enumVal *EnumValue, enumType *types.EnumType) (int, error) {
	if enumVal == nil {
		return -1, fmt.Errorf("enum value is nil")
	}
	if enumType == nil {
		return -1, fmt.Errorf("enum type metadata is nil for %s", enumVal.TypeName)
	}

	if enumVal.ValueName != "" {
		for idx, name := range enumType.OrderedNames {
			if name == enumVal.ValueName {
				return idx, nil
			}
		}
	}

	for idx, name := range enumType.OrderedNames {
		if enumType.Values[name] == enumVal.OrdinalValue {
			return idx, nil
		}
	}

	return -1, fmt.Errorf("enum value '%s' (%d) not found in type '%s'", enumVal.ValueName, enumVal.OrdinalValue, enumVal.TypeName)
}

// EnumValueAtIndex returns the enum value at the given declaration-order index.
func EnumValueAtIndex(typeName string, enumType *types.EnumType, index int) (*EnumValue, error) {
	if enumType == nil {
		return nil, fmt.Errorf("enum type metadata is nil for %s", typeName)
	}
	if index < 0 || index >= len(enumType.OrderedNames) {
		return nil, fmt.Errorf("enum index %d out of bounds for type '%s'", index, typeName)
	}

	valueName := enumType.OrderedNames[index]
	return &EnumValue{
		TypeName:     typeName,
		ValueName:    valueName,
		OrdinalValue: enumType.Values[valueName],
	}, nil
}

// Type returns "ENUM".
func (e *EnumValue) Type() string {
	return "ENUM"
}

// String returns the enum value's ordinal value as a string.
// In DWScript, when an enum is converted to string (e.g., for Print()),
// it returns the ordinal value, not the name.
func (e *EnumValue) String() string {
	return fmt.Sprintf("%d", e.OrdinalValue)
}

// GetOrdinal returns the ordinal (integer) value of the enum.
func (e *EnumValue) GetOrdinal() int {
	return e.OrdinalValue
}

// GetEnumTypeName returns the enum type name (e.g., "TColor").
func (e *EnumValue) GetEnumTypeName() string {
	return e.TypeName
}

// ============================================================================
// EnumTypeValue - Enum Type Metadata
// ============================================================================

// EnumTypeValue represents enum type metadata in DWScript.
// It wraps a types.EnumType and is used to store enum type information
// in the TypeSystem registry.
//
// This type implements:
//   - Value interface (Type(), String())
//   - GetEnumType() for TypeSystem's EnumTypeValueAccessor interface
type EnumTypeValue struct {
	EnumType *types.EnumType
}

// Type returns "ENUM_TYPE".
func (e *EnumTypeValue) Type() string {
	return "ENUM_TYPE"
}

// String returns the enum type name.
func (e *EnumTypeValue) String() string {
	if e.EnumType == nil {
		return "<nil enum type>"
	}
	return e.EnumType.Name
}

// GetEnumType returns the underlying EnumType.
// This implements the EnumTypeValueAccessor interface expected by the TypeSystem.
func (e *EnumTypeValue) GetEnumType() *types.EnumType {
	return e.EnumType
}

// NewEnumTypeValue creates a new EnumTypeValue from an EnumType.
func NewEnumTypeValue(enumType *types.EnumType) *EnumTypeValue {
	return &EnumTypeValue{EnumType: enumType}
}
