package runtime

import "github.com/cwbudde/go-dws/internal/types"

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
