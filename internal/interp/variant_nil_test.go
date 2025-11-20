package interp

import (
	"bytes"
	"testing"
)

// TestValuesEqualWithNilVariants tests the fix for nil dereference when comparing uninitialized variants
func TestValuesEqualWithNilVariants(t *testing.T) {
	var buf bytes.Buffer
	interp := New(&buf)

	tests := []struct {
		left     Value
		right    Value
		name     string
		expected bool
	}{
		{
			name:     "two nil values should be equal",
			left:     nil,
			right:    nil,
			expected: true,
		},
		{
			name:     "nil and integer should not be equal",
			left:     nil,
			right:    &IntegerValue{Value: 42},
			expected: false,
		},
		{
			name:     "integer and nil should not be equal",
			left:     &IntegerValue{Value: 42},
			right:    nil,
			expected: false,
		},
		{
			name:     "uninitialized variant and nil should be equal",
			left:     &VariantValue{Value: nil},
			right:    nil,
			expected: true,
		},
		{
			name:     "two uninitialized variants should be equal",
			left:     &VariantValue{Value: nil},
			right:    &VariantValue{Value: nil},
			expected: true,
		},
		{
			name:     "uninitialized variant and integer should not be equal",
			left:     &VariantValue{Value: nil},
			right:    &IntegerValue{Value: 42},
			expected: false,
		},
		{
			name:     "initialized variant and integer should be compared by value",
			left:     &VariantValue{Value: &IntegerValue{Value: 42}},
			right:    &IntegerValue{Value: 42},
			expected: true,
		},
		{
			name:     "initialized variant with different value should not match",
			left:     &VariantValue{Value: &IntegerValue{Value: 42}},
			right:    &IntegerValue{Value: 10},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interp.valuesEqual(tt.left, tt.right)
			if result != tt.expected {
				t.Errorf("valuesEqual(%v, %v) = %v, expected %v", tt.left, tt.right, result, tt.expected)
			}
		})
	}
}

// TestIsInRangeWithNilVariants tests the fix for nil handling in range checks
func TestIsInRangeWithNilVariants(t *testing.T) {
	var buf bytes.Buffer
	interp := New(&buf)

	tests := []struct {
		value    Value
		start    Value
		end      Value
		name     string
		expected bool
	}{
		{
			name:     "nil value should not be in range",
			value:    nil,
			start:    &IntegerValue{Value: 1},
			end:      &IntegerValue{Value: 10},
			expected: false,
		},
		{
			name:     "nil start should return false",
			value:    &IntegerValue{Value: 5},
			start:    nil,
			end:      &IntegerValue{Value: 10},
			expected: false,
		},
		{
			name:     "nil end should return false",
			value:    &IntegerValue{Value: 5},
			start:    &IntegerValue{Value: 1},
			end:      nil,
			expected: false,
		},
		{
			name:     "uninitialized variant value should not be in range",
			value:    &VariantValue{Value: nil},
			start:    &IntegerValue{Value: 1},
			end:      &IntegerValue{Value: 10},
			expected: false,
		},
		{
			name:     "uninitialized variant start should return false",
			value:    &IntegerValue{Value: 5},
			start:    &VariantValue{Value: nil},
			end:      &IntegerValue{Value: 10},
			expected: false,
		},
		{
			name:     "all nil should return false",
			value:    nil,
			start:    nil,
			end:      nil,
			expected: false,
		},
		{
			name:     "initialized variant in range should return true",
			value:    &VariantValue{Value: &IntegerValue{Value: 5}},
			start:    &IntegerValue{Value: 1},
			end:      &IntegerValue{Value: 10},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interp.isInRange(tt.value, tt.start, tt.end)
			if result != tt.expected {
				t.Errorf("isInRange(%v, %v, %v) = %v, expected %v", tt.value, tt.start, tt.end, result, tt.expected)
			}
		})
	}
}
