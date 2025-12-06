package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
)

// TestIsFalsey tests the isFalsey helper function with various value types.
func TestIsFalsey(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		expected bool
	}{
		// Nil values
		{"nil value", nil, true},
		{"NilValue", &runtime.NilValue{}, true},

		// Integer values
		{"zero integer", &runtime.IntegerValue{Value: 0}, true},
		{"positive integer", &runtime.IntegerValue{Value: 42}, false},
		{"negative integer", &runtime.IntegerValue{Value: -5}, false},

		// Float values
		{"zero float", &runtime.FloatValue{Value: 0.0}, true},
		{"positive float", &runtime.FloatValue{Value: 3.14}, false},
		{"negative float", &runtime.FloatValue{Value: -2.5}, false},

		// String values
		{"empty string", &runtime.StringValue{Value: ""}, true},
		{"non-empty string", &runtime.StringValue{Value: "hello"}, false},

		// Boolean values
		{"false boolean", &runtime.BooleanValue{Value: false}, true},
		{"true boolean", &runtime.BooleanValue{Value: true}, false},

		// Array values - THE KEY TEST FOR TASK 3.8.3.0b
		{
			"empty array",
			&runtime.ArrayValue{
				ArrayType: types.NewDynamicArrayType(types.INTEGER),
				Elements:  []Value{},
			},
			true, // Empty arrays should be falsey
		},
		{
			"non-empty array",
			&runtime.ArrayValue{
				ArrayType: types.NewDynamicArrayType(types.INTEGER),
				Elements: []Value{
					&runtime.IntegerValue{Value: 1},
					&runtime.IntegerValue{Value: 2},
				},
			},
			false, // Non-empty arrays should be truthy
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isFalsey(tt.value)
			if result != tt.expected {
				t.Errorf("isFalsey(%v) = %v, expected %v", tt.value, result, tt.expected)
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

	if !isFalsey(variantWithEmptyArray) {
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

	if isFalsey(variantWithNonEmptyArray) {
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
