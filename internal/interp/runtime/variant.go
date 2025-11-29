package runtime

import "github.com/cwbudde/go-dws/internal/types"

// VariantValue represents a Variant value in DWScript (Task 9.4).
// Task 3.5.139: Moved from internal/interp/value.go to runtime package for evaluator access.
//
// Variant is DWScript's dynamic type that can hold any value.
// Similar to Delphi's Variant type or Go's interface{}.
//
// When a variable is declared as Variant:
//
//	var v: Variant := 42;
//	// Creates: VariantValue{Value: IntegerValue{42}, ActualType: INTEGER}
//
//	v := 'hello';
//	// Creates: VariantValue{Value: StringValue{'hello'}, ActualType: STRING}
//
// See reference/dwscript-original/Source/dwsVariantFunctions.pas
type VariantValue struct {
	Value      Value      // The wrapped runtime value
	ActualType types.Type // The actual type of the wrapped value (for type checking)
}

// Type returns "VARIANT" to identify this as a Variant value.
func (v *VariantValue) Type() string {
	return "VARIANT"
}

// String returns the string representation by delegating to the wrapped value.
// This allows Variant values to be printed naturally.
func (v *VariantValue) String() string {
	if v.Value == nil {
		return "Unassigned" // Similar to Delphi's unassigned variant
	}
	return v.Value.String()
}

// GetVariantValue returns the wrapped value.
// Task 3.5.94: Implements VariantAccessor interface for type cast support.
func (v *VariantValue) GetVariantValue() Value {
	return v.Value
}

// UnwrapVariant returns the underlying wrapped value.
// This method implements the runtime.VariantWrapper interface, allowing
// the evaluator package to unwrap variants without circular dependencies.
// Returns UnassignedValue if the variant is nil/uninitialized.
func (v *VariantValue) UnwrapVariant() Value {
	if v.Value == nil {
		return &UnassignedValue{}
	}
	return v.Value
}

// IsUninitialized returns true if the variant has no wrapped value (Value == nil).
// Task 3.5.103f: Implements runtime.VariantWrapper interface for variant comparison semantics.
// An uninitialized variant equals falsey values, while a variant containing Unassigned does not.
func (v *VariantValue) IsUninitialized() bool {
	return v.Value == nil
}

// BoxVariant wraps any Value in a VariantValue for dynamic typing.
// Task 9.227: Implement VariantValue boxing in interpreter.
// Task 3.5.139: Moved to runtime package for direct evaluator access.
//
// Boxing preserves the original value and tracks its type for later unboxing.
// Prevents double-wrapping - if the value is already a Variant, returns it as-is.
//
// Examples:
//   - BoxVariant(&IntegerValue{42}) → VariantValue{Value: IntegerValue{42}, ActualType: INTEGER}
//   - BoxVariant(&StringValue{"hello"}) → VariantValue{Value: StringValue{"hello"}, ActualType: STRING}
//   - BoxVariant(nil) → VariantValue{Value: nil, ActualType: nil}
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
	switch value.Type() {
	case "INTEGER":
		actualType = types.INTEGER
	case "FLOAT":
		actualType = types.FLOAT
	case "STRING":
		actualType = types.STRING
	case "BOOLEAN":
		actualType = types.BOOLEAN
	case "NIL":
		actualType = nil // nil has no type
	case "NULL":
		actualType = nil // Task 9.4.1: Null has no specific type
	case "UNASSIGNED":
		actualType = nil // Task 9.4.1: Unassigned has no specific type
	// Complex types (arrays, records, objects) will be added as needed
	// For now, we store nil for ActualType and rely on Value.Type()
	default:
		actualType = nil
	}

	return &VariantValue{
		Value:      value,
		ActualType: actualType,
	}
}
