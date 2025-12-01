package runtime

import "fmt"

// EnumValue represents an enumerated value in DWScript.
// Enums are named constants with associated ordinal values.
//
// Phase 3.5.4 - Type Migration: Migrated from internal/interp to runtime/
// to enable evaluator package to work with enum values directly.
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
