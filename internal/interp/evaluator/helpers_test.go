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

		// Array values - THE KEY TEST FOR TASK 3.8.3.0b
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
