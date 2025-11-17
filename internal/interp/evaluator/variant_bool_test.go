package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// mockVariantValue simulates VariantValue without importing internal/interp
// (which would create circular dependency). It implements runtime.VariantWrapper.
type mockVariantValue struct {
	wrapped runtime.Value
}

func (m *mockVariantValue) Type() string {
	return "VARIANT"
}

func (m *mockVariantValue) String() string {
	if m.wrapped == nil {
		return "Unassigned"
	}
	return m.wrapped.String()
}

func (m *mockVariantValue) UnwrapVariant() runtime.Value {
	if m.wrapped == nil {
		return &mockUnassignedValue{}
	}
	return m.wrapped
}

// mockUnassignedValue simulates UnassignedValue
type mockUnassignedValue struct{}

func (m *mockUnassignedValue) Type() string   { return "UNASSIGNED" }
func (m *mockUnassignedValue) String() string { return "Unassigned" }

func TestIsTruthy_Variants(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		expected bool
	}{
		// Variant wrapping boolean values
		{
			name:     "Variant(True)",
			value:    &mockVariantValue{wrapped: &runtime.BooleanValue{Value: true}},
			expected: true,
		},
		{
			name:     "Variant(False)",
			value:    &mockVariantValue{wrapped: &runtime.BooleanValue{Value: false}},
			expected: false,
		},

		// Variant wrapping integer values
		{
			name:     "Variant(1)",
			value:    &mockVariantValue{wrapped: &runtime.IntegerValue{Value: 1}},
			expected: true,
		},
		{
			name:     "Variant(42)",
			value:    &mockVariantValue{wrapped: &runtime.IntegerValue{Value: 42}},
			expected: true,
		},
		{
			name:     "Variant(-5)",
			value:    &mockVariantValue{wrapped: &runtime.IntegerValue{Value: -5}},
			expected: true,
		},
		{
			name:     "Variant(0)",
			value:    &mockVariantValue{wrapped: &runtime.IntegerValue{Value: 0}},
			expected: false,
		},

		// Variant wrapping float values
		{
			name:     "Variant(3.14)",
			value:    &mockVariantValue{wrapped: &runtime.FloatValue{Value: 3.14}},
			expected: true,
		},
		{
			name:     "Variant(-2.5)",
			value:    &mockVariantValue{wrapped: &runtime.FloatValue{Value: -2.5}},
			expected: true,
		},
		{
			name:     "Variant(0.0)",
			value:    &mockVariantValue{wrapped: &runtime.FloatValue{Value: 0.0}},
			expected: false,
		},

		// Variant wrapping string values
		{
			name:     "Variant('hello')",
			value:    &mockVariantValue{wrapped: &runtime.StringValue{Value: "hello"}},
			expected: true,
		},
		{
			name:     "Variant('0')",
			value:    &mockVariantValue{wrapped: &runtime.StringValue{Value: "0"}},
			expected: true, // Non-empty string is truthy
		},
		{
			name:     "Variant('')",
			value:    &mockVariantValue{wrapped: &runtime.StringValue{Value: ""}},
			expected: false,
		},

		// Variant wrapping nil/unassigned
		{
			name:     "Variant(nil)",
			value:    &mockVariantValue{wrapped: nil},
			expected: false,
		},
		{
			name:     "Variant(Unassigned)",
			value:    &mockVariantValue{wrapped: &mockUnassignedValue{}},
			expected: false,
		},

		// Nested variants (should unwrap recursively)
		{
			name: "Variant(Variant(True))",
			value: &mockVariantValue{
				wrapped: &mockVariantValue{
					wrapped: &runtime.BooleanValue{Value: true},
				},
			},
			expected: true,
		},
		{
			name: "Variant(Variant(0))",
			value: &mockVariantValue{
				wrapped: &mockVariantValue{
					wrapped: &runtime.IntegerValue{Value: 0},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTruthy(tt.value)
			if result != tt.expected {
				t.Errorf("isTruthy(%s) = %v, expected %v",
					tt.name, result, tt.expected)
			}
		})
	}
}

func TestVariantToBool(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		expected bool
	}{
		// Direct runtime values (already unwrapped)
		{
			name:     "BooleanValue(true)",
			value:    &runtime.BooleanValue{Value: true},
			expected: true,
		},
		{
			name:     "IntegerValue(1)",
			value:    &runtime.IntegerValue{Value: 1},
			expected: true,
		},
		{
			name:     "IntegerValue(0)",
			value:    &runtime.IntegerValue{Value: 0},
			expected: false,
		},
		{
			name:     "FloatValue(3.14)",
			value:    &runtime.FloatValue{Value: 3.14},
			expected: true,
		},
		{
			name:     "FloatValue(0.0)",
			value:    &runtime.FloatValue{Value: 0.0},
			expected: false,
		},
		{
			name:     "StringValue('test')",
			value:    &runtime.StringValue{Value: "test"},
			expected: true,
		},
		{
			name:     "StringValue('')",
			value:    &runtime.StringValue{Value: ""},
			expected: false,
		},
		{
			name:     "NilValue",
			value:    &runtime.NilValue{},
			expected: false,
		},

		// Wrapped variants (tests unwrapping)
		{
			name:     "Variant wrapping True",
			value:    &mockVariantValue{wrapped: &runtime.BooleanValue{Value: true}},
			expected: true,
		},
		{
			name:     "Variant wrapping 42",
			value:    &mockVariantValue{wrapped: &runtime.IntegerValue{Value: 42}},
			expected: true,
		},
		{
			name:     "Variant wrapping 0",
			value:    &mockVariantValue{wrapped: &runtime.IntegerValue{Value: 0}},
			expected: false,
		},

		// Edge cases
		{
			name:     "nil value",
			value:    nil,
			expected: false,
		},
		{
			name:     "Unassigned",
			value:    &mockUnassignedValue{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := variantToBool(tt.value)
			if result != tt.expected {
				t.Errorf("variantToBool(%s) = %v, expected %v",
					tt.name, result, tt.expected)
			}
		})
	}
}

// TestDebugVariantUnwrapping debugs the unwrapping process
func TestDebugVariantUnwrapping(t *testing.T) {
	intVal := &runtime.IntegerValue{Value: 0}
	t.Logf("Direct IntegerValue(0): Type=%T, Value=%+v", intVal, intVal)

	variant := &mockVariantValue{wrapped: intVal}
	t.Logf("Variant Type: %s", variant.Type())

	// Check if it implements VariantWrapper
	if vw, ok := interface{}(variant).(runtime.VariantWrapper); ok {
		t.Log("✓ Implements runtime.VariantWrapper")
		unwrapped := vw.UnwrapVariant()
		t.Logf("Unwrapped: Type=%T, Value=%+v", unwrapped, unwrapped)

		// Check type of unwrapped value
		if iv, ok := unwrapped.(*runtime.IntegerValue); ok {
			t.Logf("✓ Unwrapped is *runtime.IntegerValue with value %d", iv.Value)
		} else {
			t.Errorf("✗ Unwrapped is not *runtime.IntegerValue, it's %T", unwrapped)
		}
	} else {
		t.Error("✗ Does NOT implement runtime.VariantWrapper")
	}

	// Now test variantToBool
	result := variantToBool(variant)
	t.Logf("variantToBool(variant) = %v (expected false)", result)
}

// TestVariantBooleanCoercionBugFix verifies the fix for the PR #178 bug
// where Variants always evaluated to false in conditionals
func TestVariantBooleanCoercionBugFix(t *testing.T) {
	t.Run("Bug: Variant(True) was returning false", func(t *testing.T) {
		// Before fix: This would return false (BUG)
		// After fix: This should return true
		variant := &mockVariantValue{
			wrapped: &runtime.BooleanValue{Value: true},
		}
		result := isTruthy(variant)
		if !result {
			t.Error("BUG: Variant wrapping True should be truthy, but got false")
		}
	})

	t.Run("Bug: Variant(non-zero integer) was returning false", func(t *testing.T) {
		// Before fix: This would return false (BUG)
		// After fix: This should return true
		variant := &mockVariantValue{
			wrapped: &runtime.IntegerValue{Value: 42},
		}
		result := isTruthy(variant)
		if !result {
			t.Error("BUG: Variant wrapping non-zero integer should be truthy, but got false")
		}
	})

	t.Run("Bug: Variant(zero) correctly returns false", func(t *testing.T) {
		variant := &mockVariantValue{
			wrapped: &runtime.IntegerValue{Value: 0},
		}
		result := isTruthy(variant)
		if result {
			t.Error("Variant wrapping zero should be falsy, but got true")
		}
	})
}
