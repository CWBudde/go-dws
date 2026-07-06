package types

// JSONVariantType represents the DWScript JSON connector value type (JSONVariant).
//
// Unlike a plain Variant, a JSONVariant supports dynamic browsing: arbitrary member
// access (v.foo), arbitrary indexing (v['k'], v[3]), and the JSON value methods
// (TypeName, Length, Clone, Extend, Add, ...). Every such access yields another
// JSONVariant, so chains like v.a.b[0].TypeName() type-check without knowing the
// JSON structure at compile time. A JSONVariant implicitly converts to the base
// scalar types and to Variant.
//
// Keeping this distinct from VARIANT is required for DWScript parity: a plain
// Variant must still reject member access (see JSONConnectorFail/coalesce_typ),
// while a JSONVariant must permit it.
//
// See reference/dwscript-original/Source/dwsJSONConnector.pas.
type JSONVariantType struct{}

func (t *JSONVariantType) String() string   { return "JSONVariant" }
func (t *JSONVariantType) TypeKind() string { return "JSON_VARIANT" }

func (t *JSONVariantType) Equals(other Type) bool {
	other = GetUnderlyingType(other)
	_, ok := other.(*JSONVariantType)
	return ok
}

// JSON_VARIANT is the singleton JSONVariant type instance.
var JSON_VARIANT = &JSONVariantType{}

// IsJSONVariant reports whether t resolves to the JSONVariant connector type.
func IsJSONVariant(t Type) bool {
	if t == nil {
		return false
	}
	return GetUnderlyingType(t).TypeKind() == "JSON_VARIANT"
}
