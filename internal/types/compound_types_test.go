// Package types_test contains tests for compound types (arrays, records, sets, enums).
// Task 9.10a: Comprehensive tests for SetType with focus on storage kind selection.
package types

import (
	"fmt"
	"testing"
)

// ============================================================================
// SetType Tests
// ============================================================================

func TestSetType(t *testing.T) {
	// Create enum type for testing
	colorEnum := &EnumType{
		Name:         "TColor",
		Values:       map[string]int{"Red": 0, "Green": 1, "Blue": 2},
		OrderedNames: []string{"Red", "Green", "Blue"},
	}

	t.Run("Create basic set type", func(t *testing.T) {
		// Task 8.81: Create SetType with NewSetType factory
		setType := NewSetType(colorEnum)

		// Test String() method - should be "set of TColor"
		expected := "set of TColor"
		if setType.String() != expected {
			t.Errorf("String() = %v, want %v", setType.String(), expected)
		}

		// Test TypeKind() method
		if setType.TypeKind() != "SET" {
			t.Errorf("TypeKind() = %v, want SET", setType.TypeKind())
		}

		// Test ElementType field
		if setType.ElementType != colorEnum {
			t.Error("ElementType should match the provided enum type")
		}
	})

	t.Run("Set type equality", func(t *testing.T) {
		// Task 8.81: Test Equals() method
		enum1 := &EnumType{
			Name:         "TColor",
			Values:       map[string]int{"Red": 0, "Green": 1, "Blue": 2},
			OrderedNames: []string{"Red", "Green", "Blue"},
		}
		enum2 := &EnumType{
			Name:         "TSize",
			Values:       map[string]int{"Small": 0, "Medium": 1, "Large": 2},
			OrderedNames: []string{"Small", "Medium", "Large"},
		}

		set1 := NewSetType(enum1)
		set2 := NewSetType(enum1)
		set3 := NewSetType(enum2)

		// Same element type sets should be equal
		if !set1.Equals(set2) {
			t.Error("Sets with same element type should be equal")
		}

		// Different element type sets should not be equal
		if set1.Equals(set3) {
			t.Error("Sets with different element types should not be equal")
		}

		// Set should not equal other types
		if set1.Equals(INTEGER) {
			t.Error("SetType should not equal INTEGER")
		}

		// Set should not equal enum type
		if set1.Equals(enum1) {
			t.Error("SetType should not equal EnumType")
		}
	})

	t.Run("Set type with nil element type", func(t *testing.T) {
		// Task 8.83: Validation - ensure nil enum type is handled
		// For now, we allow it to be created (validation will be added in semantic analysis)
		setType := NewSetType(nil)
		if setType == nil {
			t.Error("NewSetType should not return nil even with nil element type")
		}
		if setType.ElementType != nil {
			t.Error("ElementType should be nil when created with nil")
		}
	})

	t.Run("Set type with different enum instances", func(t *testing.T) {
		// Even if values are same, different enum instances should work
		enum1 := &EnumType{
			Name:         "TColor",
			Values:       map[string]int{"Red": 0},
			OrderedNames: []string{"Red"},
		}
		enum2 := &EnumType{
			Name:         "TColor",
			Values:       map[string]int{"Red": 0},
			OrderedNames: []string{"Red"},
		}

		set1 := NewSetType(enum1)
		set2 := NewSetType(enum2)

		// Sets should be equal because enums have same name (nominal typing)
		if !set1.Equals(set2) {
			t.Error("Sets with same-named enums should be equal")
		}
	})

	t.Run("Set type string representation", func(t *testing.T) {
		// Test various enum types to ensure String() works correctly
		tests := []struct {
			name     string
			enum     *EnumType
			expected string
		}{
			{
				name: "simple enum",
				enum: &EnumType{
					Name:         "TStatus",
					Values:       map[string]int{"Ok": 0, "Error": 1},
					OrderedNames: []string{"Ok", "Error"},
				},
				expected: "set of TStatus",
			},
			{
				name: "weekday enum",
				enum: &EnumType{
					Name:         "TWeekday",
					Values:       map[string]int{"Mon": 0, "Tue": 1, "Wed": 2},
					OrderedNames: []string{"Mon", "Tue", "Wed"},
				},
				expected: "set of TWeekday",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				setType := NewSetType(tt.enum)
				if setType.String() != tt.expected {
					t.Errorf("String() = %v, want %v", setType.String(), tt.expected)
				}
			})
		}
	})
}

// ============================================================================
// SetType Storage Kind Tests (Task 9.10a)
// ============================================================================

// TestSetStorageKind_SmallEnums tests that small enums use bitmask storage.
// Task 9.10a: Verify storage kind selection for enums ≤64 values.
func TestSetStorageKind_SmallEnums(t *testing.T) {
	tests := []struct {
		name        string
		enumSize    int
		wantStorage SetStorageKind
	}{
		{"1 element", 1, SetStorageBitmask},
		{"2 elements", 2, SetStorageBitmask},
		{"10 elements", 10, SetStorageBitmask},
		{"32 elements", 32, SetStorageBitmask},
		{"63 elements", 63, SetStorageBitmask},
		{"64 elements (boundary)", 64, SetStorageBitmask},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create enum with specified size
			enumType := createEnumWithSize(t, "TSmallEnum", tt.enumSize)

			// Create set type
			setType := NewSetType(enumType)

			// Verify storage kind
			if setType.StorageKind != tt.wantStorage {
				t.Errorf("Storage kind = %v, want %v for enum with %d elements",
					setType.StorageKind, tt.wantStorage, tt.enumSize)
			}
		})
	}
}

// TestSetStorageKind_LargeEnums tests that large enums use map storage.
// Task 9.10a: Verify storage kind selection for enums >64 values.
func TestSetStorageKind_LargeEnums(t *testing.T) {
	tests := []struct {
		name        string
		enumSize    int
		wantStorage SetStorageKind
	}{
		{"65 elements (boundary)", 65, SetStorageMap},
		{"70 elements", 70, SetStorageMap},
		{"100 elements", 100, SetStorageMap},
		{"128 elements", 128, SetStorageMap},
		{"200 elements", 200, SetStorageMap},
		{"500 elements", 500, SetStorageMap},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create enum with specified size
			enumType := createEnumWithSize(t, "TLargeEnum", tt.enumSize)

			// Create set type
			setType := NewSetType(enumType)

			// Verify storage kind
			if setType.StorageKind != tt.wantStorage {
				t.Errorf("Storage kind = %v, want %v for enum with %d elements",
					setType.StorageKind, tt.wantStorage, tt.enumSize)
			}

			// Verify enum size
			if len(enumType.OrderedNames) != tt.enumSize {
				t.Errorf("Enum size = %d, want %d", len(enumType.OrderedNames), tt.enumSize)
			}
		})
	}
}

// TestSetStorageKind_Boundary tests the exact boundary between bitmask and map storage.
// Task 9.10a: Ensure 64→65 element transition changes storage kind.
func TestSetStorageKind_Boundary(t *testing.T) {
	// Create enum with exactly 64 elements (should use bitmask)
	enum64 := createEnumWithSize(t, "TEnum64", 64)
	set64 := NewSetType(enum64)

	if set64.StorageKind != SetStorageBitmask {
		t.Errorf("64-element enum should use bitmask, got %v", set64.StorageKind)
	}

	// Create enum with exactly 65 elements (should use map)
	enum65 := createEnumWithSize(t, "TEnum65", 65)
	set65 := NewSetType(enum65)

	if set65.StorageKind != SetStorageMap {
		t.Errorf("65-element enum should use map, got %v", set65.StorageKind)
	}

	// Verify they're not equal (different storage kinds)
	if set64.StorageKind == set65.StorageKind {
		t.Error("64 and 65 element sets should have different storage kinds")
	}
}

// TestSetType_EqualityWithDifferentStorageKinds tests that sets with different
// storage kinds can still be equal if their element types are compatible.
// Task 9.10a: Verify type equality is independent of storage strategy.
func TestSetType_EqualityWithDifferentStorageKinds(t *testing.T) {
	// Create two enums with same name but different sizes
	smallEnum := &EnumType{
		Name:         "TValues",
		Values:       map[string]int{"A": 0, "B": 1, "C": 2},
		OrderedNames: []string{"A", "B", "C"},
	}

	largeEnum := createEnumWithSize(t, "TDifferent", 100)

	smallSet := NewSetType(smallEnum)
	largeSet := NewSetType(largeEnum)

	// Verify different storage kinds
	if smallSet.StorageKind != SetStorageBitmask {
		t.Error("Small set should use bitmask")
	}
	if largeSet.StorageKind != SetStorageMap {
		t.Error("Large set should use map")
	}

	// Sets with different element types should not be equal
	if smallSet.Equals(largeSet) {
		t.Error("Sets with different element types should not be equal regardless of storage")
	}

	// Now test two large sets with same element type
	largeEnum1 := createEnumWithSize(t, "TSame", 100)
	largeEnum2 := &EnumType{
		Name:         "TSame", // Same name
		Values:       largeEnum1.Values,
		OrderedNames: largeEnum1.OrderedNames,
	}

	largeSet1 := NewSetType(largeEnum1)
	largeSet2 := NewSetType(largeEnum2)

	// Both should use map storage
	if largeSet1.StorageKind != SetStorageMap || largeSet2.StorageKind != SetStorageMap {
		t.Error("Both large sets should use map storage")
	}

	// Should be equal (same element type name)
	if !largeSet1.Equals(largeSet2) {
		t.Error("Sets with same element type name should be equal")
	}
}

// TestSetType_LargeEnumString tests String() method with large enums.
// Task 9.10a: Ensure string representation works for large sets.
func TestSetType_LargeEnumString(t *testing.T) {
	tests := []struct {
		name     string
		enumName string
		enumSize int
		expected string
	}{
		{"100 elements", "TLarge100", 100, "set of TLarge100"},
		{"200 elements", "TLarge200", 200, "set of TLarge200"},
		{"500 elements", "TLarge500", 500, "set of TLarge500"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enumType := createEnumWithSize(t, tt.enumName, tt.enumSize)
			setType := NewSetType(enumType)

			if setType.String() != tt.expected {
				t.Errorf("String() = %v, want %v", setType.String(), tt.expected)
			}

			// Verify it's using map storage
			if setType.StorageKind != SetStorageMap {
				t.Errorf("Set with %d elements should use map storage", tt.enumSize)
			}
		})
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// createEnumWithSize creates an EnumType with the specified number of elements.
// Elements are named E00, E01, E02, ... E99, E100, etc.
// Task 9.10a: Helper for generating test enums of various sizes.
func createEnumWithSize(t *testing.T, name string, size int) *EnumType {
	t.Helper()

	if size < 0 {
		t.Fatalf("Invalid enum size: %d", size)
	}

	values := make(map[string]int, size)
	orderedNames := make([]string, size)

	for i := 0; i < size; i++ {
		elementName := fmt.Sprintf("E%02d", i)
		values[elementName] = i
		orderedNames[i] = elementName
	}

	return &EnumType{
		Name:         name,
		Values:       values,
		OrderedNames: orderedNames,
	}
}
