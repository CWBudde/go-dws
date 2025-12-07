package runtime

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// TypeMetaValue represents a type reference in DWScript.
// This is used for type-as-value scenarios where a type itself is passed as a runtime value.
//
// Phase 3.5.4 - Type Migration: Migrated from internal/interp to runtime/
// to enable evaluator package to work with type metadata directly.
//
// Examples:
//   - Low(Integer) where `Integer` is a TypeMetaValue wrapping types.INTEGER
//   - High(TColor) where `TColor` is a TypeMetaValue wrapping the enum type
//   - for e in TColor do ... where `TColor` is a TypeMetaValue representing the enum type
//
// TypeMetaValue is distinct from RTTITypeInfoValue:
//   - TypeMetaValue: Represents a type at compile-time/runtime (type name as value)
//   - RTTITypeInfoValue: Runtime type information from TypeOf() introspection
type TypeMetaValue struct {
	TypeInfo types.Type // The type metadata (e.g., types.INTEGER, types.FLOAT, enum type)
	TypeName string     // The type name for display (e.g., "Integer", "TColor")
}

// Type returns "TYPE_META".
func (t *TypeMetaValue) Type() string {
	return "TYPE_META"
}

// String returns the type name.
func (t *TypeMetaValue) String() string {
	return t.TypeName
}

// IsEnumTypeMeta returns true if this type meta wraps an enum type.
func (t *TypeMetaValue) IsEnumTypeMeta() bool {
	if t.TypeInfo == nil {
		return false
	}
	_, isEnum := t.TypeInfo.(*types.EnumType)
	return isEnum
}

// EnumLow returns the lowest ordinal value of the enum type.
// Returns 0 if not an enum type.
func (t *TypeMetaValue) EnumLow() int {
	if t.TypeInfo == nil {
		return 0
	}
	enumType, ok := t.TypeInfo.(*types.EnumType)
	if !ok {
		return 0
	}
	return enumType.Low()
}

// EnumHigh returns the highest ordinal value of the enum type.
// Returns 0 if not an enum type.
func (t *TypeMetaValue) EnumHigh() int {
	if t.TypeInfo == nil {
		return 0
	}
	enumType, ok := t.TypeInfo.(*types.EnumType)
	if !ok {
		return 0
	}
	return enumType.High()
}

// EnumByName looks up an enum value by name (case-insensitive).
// Supports both simple names ('Red') and qualified names ('TColor.Red').
// Returns the ordinal value if found, or 0 if not found (DWScript behavior).
func (t *TypeMetaValue) EnumByName(name string) int {
	if t.TypeInfo == nil {
		return 0
	}
	enumType, ok := t.TypeInfo.(*types.EnumType)
	if !ok {
		return 0
	}

	// Empty string returns 0 (DWScript behavior - returns first enum ordinal value)
	if name == "" {
		return 0
	}

	// Check for qualified name (TypeName.ValueName)
	searchName := name
	parts := strings.Split(name, ".")
	if len(parts) == 2 {
		// Use the value name part
		searchName = parts[1]
	}

	// Look up the value (case-insensitive)
	for valueName, ordinalValue := range enumType.Values {
		if ident.Equal(valueName, searchName) {
			return ordinalValue
		}
	}

	// Value not found, return 0 (DWScript behavior)
	return 0
}

// GetEnumValue looks up an enum value by name and returns it as a runtime Value.
// Returns nil if the name is not found or this is not an enum type meta.
// Used for member access like TColor.Red.
func (t *TypeMetaValue) GetEnumValue(name string) Value {
	if t.TypeInfo == nil {
		return nil
	}
	enumType, ok := t.TypeInfo.(*types.EnumType)
	if !ok {
		return nil
	}

	// Look up the value (case-insensitive)
	for valueName, ordinalValue := range enumType.Values {
		if ident.Equal(valueName, name) {
			return &EnumValue{
				TypeName:     t.TypeName,
				ValueName:    valueName, // Use the canonical name from the type
				OrdinalValue: ordinalValue,
			}
		}
	}

	return nil
}
