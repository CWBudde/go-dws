package runtime

import (
	"github.com/cwbudde/go-dws/internal/jsonvalue"
	"github.com/cwbudde/go-dws/internal/types"
)

// VariantValue represents a Variant value in DWScript.
// Variant is DWScript's dynamic type that can hold any value at runtime.
type VariantValue struct {
	Value      Value      // The wrapped runtime value
	ActualType types.Type // The actual type of the wrapped value (for type checking)
}

// Type returns "VARIANT" to identify this as a Variant value.
func (v *VariantValue) Type() string {
	return "VARIANT"
}

// String returns the string representation of the wrapped value.
func (v *VariantValue) String() string {
	if v.Value == nil {
		return "" // DWScript prints an unassigned Variant as nothing at all
	}
	return v.Value.String()
}

// GetVariantValue returns the wrapped value.
func (v *VariantValue) GetVariantValue() Value {
	return v.Value
}

// UnwrapVariant returns the underlying wrapped value or UnassignedValue if nil.
func (v *VariantValue) UnwrapVariant() Value {
	if v.Value == nil {
		return &UnassignedValue{}
	}
	return v.Value
}

// IsUninitialized returns true if the variant has no wrapped value (Value == nil).
func (v *VariantValue) IsUninitialized() bool {
	return v.Value == nil
}

// BoxVariant wraps any Value in a VariantValue for dynamic typing.
// Prevents double-wrapping - if the value is already a Variant, returns it as-is.
func BoxVariant(value Value) *VariantValue {
	if value == nil {
		return &VariantValue{Value: nil, ActualType: nil}
	}

	// If already a Variant, return as-is (no double-wrapping)
	if variant, ok := value.(*VariantValue); ok {
		return variant
	}

	// Map runtime Value type to semantic types.Type
	var actualType types.Type
	switch KindOf(value) {
	case KindInteger:
		actualType = types.INTEGER
	case KindFloat:
		actualType = types.FLOAT
	case KindString:
		actualType = types.STRING
	case KindBoolean:
		actualType = types.BOOLEAN
	case KindByteBuffer:
		actualType = types.BYTE_BUFFER
	case KindNil:
		actualType = nil // nil has no type
	case KindNull:
		actualType = nil
	case KindUnassigned:
		actualType = nil
	default:
		actualType = nil
	}

	return &VariantValue{
		Value:      value,
		ActualType: actualType,
	}
}

// BoxVariantWithJSON wraps a jsonvalue.Value in a VariantValue containing a JSONValue.
//
// This creates the necessary JSONValue wrapper and boxes it in a Variant, making JSON
// values available for variant operations. JSON has no specific ActualType since it's
// a dynamic type.
//
// Examples:
//   - BoxVariantWithJSON(jsonVal) → VariantValue{Value: JSONValue{jsonVal}, ActualType: nil}
//   - BoxVariantWithJSON(nil) → VariantValue{Value: nil, ActualType: nil}
func BoxVariantWithJSON(jv *jsonvalue.Value) *VariantValue {
	if jv == nil {
		return &VariantValue{Value: nil, ActualType: nil}
	}
	return &VariantValue{
		Value:      &JSONValue{Value: jv},
		ActualType: nil, // JSON is a dynamic type
	}
}
