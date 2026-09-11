package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
)

// TestIsFalsey tests the isFalsey helper function with various value types.
func TestIsFalsey(t *testing.T) {
	tests := []struct {
		value    Value
		name     string
		expected bool
	}{
		// Nil values
		{name: "nil value", value: nil, expected: true},
		{name: "NilValue", value: &runtime.NilValue{}, expected: true},

		// Integer values
		{name: "zero integer", value: &runtime.IntegerValue{Value: 0}, expected: true},
		{name: "positive integer", value: &runtime.IntegerValue{Value: 42}, expected: false},
		{name: "negative integer", value: &runtime.IntegerValue{Value: -5}, expected: false},

		// Float values
		{name: "zero float", value: &runtime.FloatValue{Value: 0.0}, expected: true},
		{name: "positive float", value: &runtime.FloatValue{Value: 3.14}, expected: false},
		{name: "negative float", value: &runtime.FloatValue{Value: -2.5}, expected: false},

		// String values
		{name: "empty string", value: &runtime.StringValue{Value: ""}, expected: true},
		{name: "non-empty string", value: &runtime.StringValue{Value: "hello"}, expected: false},

		// Boolean values
		{name: "false boolean", value: &runtime.BooleanValue{Value: false}, expected: true},
		{name: "true boolean", value: &runtime.BooleanValue{Value: true}, expected: false},

		// Array values
		{
			name: "empty array",
			value: &runtime.ArrayValue{
				ArrayType: types.NewDynamicArrayType(types.INTEGER),
				Elements:  []Value{},
			},
			expected: true, // Empty arrays should be falsey
		},
		{
			name: "non-empty array",
			value: &runtime.ArrayValue{
				ArrayType: types.NewDynamicArrayType(types.INTEGER),
				Elements: []Value{
					&runtime.IntegerValue{Value: 1},
					&runtime.IntegerValue{Value: 2},
				},
			},
			expected: false, // Non-empty arrays should be truthy
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsFalsey(tt.value)
			if result != tt.expected {
				t.Errorf("IsFalsey(%v) = %v, expected %v", tt.value, result, tt.expected)
			}
		})
	}
}

// TestIsFalseyWithVariant tests isFalsey with Variant values.
func TestIsFalseyWithVariant(t *testing.T) {
	// Create a variant wrapping an empty array
	emptyArray := &runtime.ArrayValue{
		ArrayType: types.NewDynamicArrayType(types.INTEGER),
		Elements:  []Value{},
	}

	variantWithEmptyArray := &runtime.VariantValue{Value: emptyArray}

	if !IsFalsey(variantWithEmptyArray) {
		t.Error("Variant wrapping empty array should be falsey")
	}

	// Create a variant wrapping a non-empty array
	nonEmptyArray := &runtime.ArrayValue{
		ArrayType: types.NewDynamicArrayType(types.INTEGER),
		Elements: []Value{
			&runtime.IntegerValue{Value: 1},
		},
	}

	variantWithNonEmptyArray := &runtime.VariantValue{Value: nonEmptyArray}

	if IsFalsey(variantWithNonEmptyArray) {
		t.Error("Variant wrapping non-empty array should be truthy")
	}
}

// TestUnwrapVariant tests the unwrapVariant helper function.
func TestUnwrapVariant(t *testing.T) {
	// Test unwrapping a variant with an integer
	innerInt := &runtime.IntegerValue{Value: 42}
	variantInt := &runtime.VariantValue{Value: innerInt}

	unwrapped := unwrapVariant(variantInt)
	if unwrapped != innerInt {
		t.Errorf("unwrapVariant should return inner value, got %v", unwrapped)
	}

	// Test unwrapping a non-variant value
	plainInt := &runtime.IntegerValue{Value: 10}
	unwrappedPlain := unwrapVariant(plainInt)
	if unwrappedPlain != plainInt {
		t.Errorf("unwrapVariant should return the value itself for non-variants, got %v", unwrappedPlain)
	}

	// Test unwrapping a nil variant
	// Note: VariantValue.UnwrapVariant() returns UnassignedValue for nil values,
	// not NilValue. The unwrapVariant helper doesn't convert UnassignedValue to NilValue
	// because UnassignedValue is not nil.
	nilVariant := &runtime.VariantValue{Value: nil}
	unwrappedNil := unwrapVariant(nilVariant)
	if _, ok := unwrappedNil.(*runtime.UnassignedValue); !ok {
		t.Errorf("unwrapVariant of nil variant should return UnassignedValue, got %T", unwrappedNil)
	}
}

// TestIsInRange_VariantUnwrap verifies that range checks unwrap Variant-wrapped
// selectors and bounds (case v of 11..12 with v: Variant, see fixture case_variant).
func TestIsInRange_VariantUnwrap(t *testing.T) {
	tests := []struct {
		value    Value
		start    Value
		end      Value
		name     string
		expected bool
	}{
		{
			name:     "variant integer inside range",
			value:    &runtime.VariantValue{Value: &runtime.IntegerValue{Value: 12}},
			start:    &runtime.IntegerValue{Value: 11},
			end:      &runtime.IntegerValue{Value: 12},
			expected: true,
		},
		{
			name:     "variant integer outside range",
			value:    &runtime.VariantValue{Value: &runtime.IntegerValue{Value: 10}},
			start:    &runtime.IntegerValue{Value: 11},
			end:      &runtime.IntegerValue{Value: 12},
			expected: false,
		},
		{
			name:     "variant bounds",
			value:    &runtime.IntegerValue{Value: 5},
			start:    &runtime.VariantValue{Value: &runtime.IntegerValue{Value: 1}},
			end:      &runtime.VariantValue{Value: &runtime.IntegerValue{Value: 9}},
			expected: true,
		},
		{
			name:     "variant string in char range",
			value:    &runtime.VariantValue{Value: &runtime.StringValue{Value: "c"}},
			start:    &runtime.StringValue{Value: "a"},
			end:      &runtime.StringValue{Value: "z"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsInRange(tt.value, tt.start, tt.end); got != tt.expected {
				t.Errorf("IsInRange() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// newTestEnumType builds enum declaration metadata from an ordered list of
// member names and their declared ordinals, preserving declaration order.
func newTestEnumType(name string, names []string, ordinals []int) *types.EnumType {
	values := make(map[string]int, len(names))
	for i, n := range names {
		values[n] = ordinals[i]
	}

	return &types.EnumType{Name: name, Values: values, OrderedNames: names}
}

// TestIsInRange_Enum verifies enum range checking in case statements
// (see fixture SimpleScripts/case_range_enum).
func TestIsInRange_Enum(t *testing.T) {
	// Implicit, contiguous ordinals: TColor = (Red, Green, Blue, Yellow).
	color := newTestEnumType("TColor",
		[]string{"Red", "Green", "Blue", "Yellow"},
		[]int{0, 1, 2, 3})
	// Explicit, non-monotonic ordinals: TDisj = (A = 1, B = 10, C = 2).
	disj := newTestEnumType("TDisj", []string{"A", "B", "C"}, []int{1, 10, 2})
	// A distinct enumeration that shares member names and ordinals with TColor.
	other := newTestEnumType("TOther", []string{"Red", "Green", "Blue"}, []int{0, 1, 2})

	enumOf := func(et *types.EnumType, name string) *runtime.EnumValue {
		return &runtime.EnumValue{
			EnumType:     et,
			TypeName:     et.Name,
			ValueName:    name,
			OrdinalValue: et.Values[name],
		}
	}

	tests := []struct {
		value    Value
		start    Value
		end      Value
		name     string
		expected bool
	}{
		{
			name:     "implicit ordinals inside range",
			value:    enumOf(color, "Green"),
			start:    enumOf(color, "Red"),
			end:      enumOf(color, "Blue"),
			expected: true,
		},
		{
			name:     "implicit ordinals on lower bound",
			value:    enumOf(color, "Red"),
			start:    enumOf(color, "Red"),
			end:      enumOf(color, "Blue"),
			expected: true,
		},
		{
			name:     "implicit ordinals on upper bound",
			value:    enumOf(color, "Blue"),
			start:    enumOf(color, "Red"),
			end:      enumOf(color, "Blue"),
			expected: true,
		},
		{
			name:     "implicit ordinals outside range",
			value:    enumOf(color, "Yellow"),
			start:    enumOf(color, "Red"),
			end:      enumOf(color, "Blue"),
			expected: false,
		},
		{
			name:     "variant-wrapped enum selector",
			value:    &runtime.VariantValue{Value: enumOf(color, "Green")},
			start:    enumOf(color, "Red"),
			end:      enumOf(color, "Blue"),
			expected: true,
		},
		{
			// Declaration order puts B (ordinal 10) between A and C, so A..C
			// covers it even though 10 falls outside the ordinal span [1, 2].
			name:     "explicit non-contiguous ordinals use declaration order",
			value:    enumOf(disj, "B"),
			start:    enumOf(disj, "A"),
			end:      enumOf(disj, "C"),
			expected: true,
		},
		{
			name:     "explicit non-contiguous ordinals outside declaration range",
			value:    enumOf(disj, "C"),
			start:    enumOf(disj, "A"),
			end:      enumOf(disj, "B"),
			expected: false,
		},
		{
			name:     "mismatched enum types never match",
			value:    enumOf(other, "Green"),
			start:    enumOf(color, "Red"),
			end:      enumOf(color, "Blue"),
			expected: false,
		},
		{
			name:     "mismatched enum bound never matches",
			value:    enumOf(color, "Green"),
			start:    enumOf(color, "Red"),
			end:      enumOf(other, "Blue"),
			expected: false,
		},
		{
			name:     "reversed bounds never match",
			value:    enumOf(color, "Green"),
			start:    enumOf(color, "Blue"),
			end:      enumOf(color, "Red"),
			expected: false,
		},
		{
			name:     "integer bounds compare declared ordinals",
			value:    enumOf(color, "Green"),
			start:    &runtime.IntegerValue{Value: 0},
			end:      &runtime.IntegerValue{Value: 2},
			expected: true,
		},
		{
			name:     "integer bounds outside declared ordinals",
			value:    enumOf(disj, "B"),
			start:    &runtime.IntegerValue{Value: 1},
			end:      &runtime.IntegerValue{Value: 2},
			expected: false,
		},
		{
			name:     "integer selector with enum bounds",
			value:    &runtime.IntegerValue{Value: 1},
			start:    enumOf(color, "Red"),
			end:      enumOf(color, "Blue"),
			expected: true,
		},
		{
			name:     "unsupported bound type",
			value:    enumOf(color, "Green"),
			start:    &runtime.StringValue{Value: "Red"},
			end:      &runtime.StringValue{Value: "Blue"},
			expected: false,
		},
		{
			// Without declaration metadata the comparison degrades to the
			// declared ordinals, which is all the values carry.
			name:     "missing metadata falls back to ordinals",
			value:    &runtime.EnumValue{TypeName: "TColor", ValueName: "Green", OrdinalValue: 1},
			start:    &runtime.EnumValue{TypeName: "tcolor", ValueName: "Red", OrdinalValue: 0},
			end:      &runtime.EnumValue{TypeName: "TCOLOR", ValueName: "Blue", OrdinalValue: 2},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsInRange(tt.value, tt.start, tt.end); got != tt.expected {
				t.Errorf("IsInRange() = %v, want %v", got, tt.expected)
			}
		})
	}
}
