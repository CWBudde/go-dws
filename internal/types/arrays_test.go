package types

import (
	"testing"
)

// ============================================================================
// Array Type Tests
// ============================================================================

func TestArrayType(t *testing.T) {
	t.Run("Dynamic array", func(t *testing.T) {
		at := NewDynamicArrayType(INTEGER)
		expected := "array of Integer"
		if at.String() != expected {
			t.Errorf("String() = %v, want %v", at.String(), expected)
		}
		if !at.IsDynamic() {
			t.Error("Should be dynamic")
		}
		if at.IsStatic() {
			t.Error("Should not be static")
		}
		if at.Size() != -1 {
			t.Errorf("Size() = %v, want -1", at.Size())
		}
	})

	t.Run("Static array", func(t *testing.T) {
		at := NewStaticArrayType(STRING, 1, 10)
		expected := "array[1..10] of String"
		if at.String() != expected {
			t.Errorf("String() = %v, want %v", at.String(), expected)
		}
		if at.IsDynamic() {
			t.Error("Should not be dynamic")
		}
		if !at.IsStatic() {
			t.Error("Should be static")
		}
		if at.Size() != 10 {
			t.Errorf("Size() = %v, want 10", at.Size())
		}
	})
}

func TestArrayTypeEquality(t *testing.T) {
	tests := []struct {
		a        Type
		b        Type
		name     string
		expected bool
	}{
		{
			a:        NewDynamicArrayType(INTEGER),
			b:        NewDynamicArrayType(INTEGER),
			name:     "Same dynamic arrays",
			expected: true,
		},
		{
			a:        NewDynamicArrayType(INTEGER),
			b:        NewDynamicArrayType(FLOAT),
			name:     "Different element types (dynamic)",
			expected: false,
		},
		{
			a:        NewStaticArrayType(STRING, 1, 10),
			b:        NewStaticArrayType(STRING, 1, 10),
			name:     "Same static arrays",
			expected: true,
		},
		{
			a:        NewStaticArrayType(INTEGER, 1, 10),
			b:        NewStaticArrayType(INTEGER, 0, 9),
			name:     "Different bounds",
			expected: false,
		},
		{
			a:        NewDynamicArrayType(INTEGER),
			b:        NewStaticArrayType(INTEGER, 1, 10),
			name:     "Dynamic vs static",
			expected: false,
		},
		{
			a:        NewDynamicArrayType(INTEGER),
			b:        INTEGER,
			name:     "Array vs non-array",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := tt.a.Equals(tt.b); result != tt.expected {
				t.Errorf("Equals() = %v, want %v", result, tt.expected)
			}
		})
	}
}
