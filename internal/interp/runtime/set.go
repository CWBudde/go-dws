package runtime

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
)

// IntRange represents an integer range for lazy set storage.
// Used by SetValue to efficiently represent large contiguous ranges
// without allocating individual elements.
//
// Phase 3.5.4 - Type Migration: Migrated from internal/interp to runtime/
// to enable evaluator package to work with set values directly.
type IntRange struct {
	Start int
	End   int
}

// SetValue represents a set value in DWScript.
// Sets are collections of enumerated values or integers with set operations.
//
// Phase 3.5.4 - Type Migration: Migrated from internal/interp to runtime/
// to enable evaluator package to work with set values directly.
//
// Storage Strategy:
//   - Small sets (< 64 elements): Bitmask in Elements field
//   - Large sets: Map storage in MapStore field
//   - Contiguous ranges: Lazy storage in Ranges field
//
// Example: type TColors = set of TColor;
// Task 9.8: Supports both bitmask and map storage.
type SetValue struct {
	SetType  *types.SetType // The set type metadata
	MapStore map[int]bool   // Map storage for large sets
	Ranges   []IntRange     // Lazy storage for contiguous ranges
	Elements uint64         // Bitmask storage for small sets (up to 64 elements)
}

// Type returns "SET".
func (s *SetValue) Type() string {
	return "SET"
}

// String returns the string representation of the set.
// Format: [element1, element2, ...] or [] for empty set
// Task 9.8: Works with both bitmask and map storage.
func (s *SetValue) String() string {
	// Quick check for empty set (both storage types)
	if s.SetType.StorageKind == types.SetStorageBitmask && s.Elements == 0 {
		return "[]"
	}
	if s.SetType.StorageKind == types.SetStorageMap && len(s.MapStore) == 0 {
		return "[]"
	}

	var elements []string

	// Task 9.226: Handle different element types for display
	if s.SetType != nil && s.SetType.ElementType != nil {
		// For enum sets, show enum names in order
		if enumType, ok := s.SetType.ElementType.(*types.EnumType); ok {
			for _, ordinal := range s.Ordinals() {
				if name := enumType.GetEnumName(ordinal); name != "" {
					elements = append(elements, name)
				} else {
					elements = append(elements, fmt.Sprintf("%d", ordinal))
				}
			}
		} else {
			// For non-enum sets (Integer, String, Boolean), show ordinal values
			// Collect ordinals that are in the set
			ordinals := s.Ordinals()

			// Convert ordinals to strings based on element type
			for _, ord := range ordinals {
				elements = append(elements, fmt.Sprintf("%d", ord))
			}
		}
	}

	if len(elements) == 0 {
		return "[]"
	}

	return "[" + strings.Join(elements, ", ") + "]"
}

// Ordinals returns a sorted slice of all ordinal values stored in the set.
// Works for both bitmask and map storage and expands lazy ranges.
func (s *SetValue) Ordinals() []int {
	seen := make(map[int]bool)

	switch s.SetType.StorageKind {
	case types.SetStorageBitmask:
		for i := 0; i < 64; i++ {
			if s.HasElement(i) {
				seen[i] = true
			}
		}
	case types.SetStorageMap:
		for ordinal := range s.MapStore {
			seen[ordinal] = true
		}
	}

	for _, r := range s.Ranges {
		if r.Start <= r.End {
			for v := r.Start; v <= r.End; v++ {
				seen[v] = true
			}
		} else {
			for v := r.Start; v >= r.End; v-- {
				seen[v] = true
			}
		}
	}

	ordinals := make([]int, 0, len(seen))
	for ord := range seen {
		ordinals = append(ordinals, ord)
	}
	sort.Ints(ordinals)
	return ordinals
}

// HasElement checks if an element with the given ordinal value is in the set.
// Task 9.8: Supports both bitmask and map storage.
// Also checks lazy ranges for large integer ranges.
func (s *SetValue) HasElement(ordinal int) bool {
	if ordinal < 0 {
		return false // Negative ordinals are invalid
	}

	// First check lazy ranges (most common for large sets)
	for _, r := range s.Ranges {
		if r.Start <= r.End {
			// Forward range
			if ordinal >= r.Start && ordinal <= r.End {
				return true
			}
		} else {
			// Reverse range
			if ordinal >= r.End && ordinal <= r.Start {
				return true
			}
		}
	}

	// Choose storage backend based on set type
	switch s.SetType.StorageKind {
	case types.SetStorageBitmask:
		if ordinal >= 64 {
			return false // Out of range for bitset
		}
		mask := uint64(1) << uint(ordinal)
		return (s.Elements & mask) != 0

	case types.SetStorageMap:
		return s.MapStore[ordinal]

	default:
		return false
	}
}

// AddElement adds an element with the given ordinal value to the set.
// This mutates the set in place (used for Include).
// Task 9.8: Supports both bitmask and map storage.
func (s *SetValue) AddElement(ordinal int) {
	if ordinal < 0 {
		return // Negative ordinals are invalid
	}

	// Choose storage backend based on set type
	switch s.SetType.StorageKind {
	case types.SetStorageBitmask:
		if ordinal >= 64 {
			return // Out of range for bitset
		}
		mask := uint64(1) << uint(ordinal)
		s.Elements |= mask

	case types.SetStorageMap:
		s.MapStore[ordinal] = true
	}
}

// RemoveElement removes an element with the given ordinal value from the set.
// This mutates the set in place (used for Exclude).
// Task 9.8: Supports both bitmask and map storage.
func (s *SetValue) RemoveElement(ordinal int) {
	if ordinal < 0 {
		return // Negative ordinals are invalid
	}

	// Choose storage backend based on set type
	switch s.SetType.StorageKind {
	case types.SetStorageBitmask:
		if ordinal >= 64 {
			return // Out of range for bitset
		}
		mask := uint64(1) << uint(ordinal)
		s.Elements &^= mask // AND NOT to clear the bit

	case types.SetStorageMap:
		delete(s.MapStore, ordinal)
	}
}

// NewSetValue creates a new empty SetValue with the given set type.
// Task 9.8: Initializes the appropriate storage backend (bitmask or map).
func NewSetValue(setType *types.SetType) *SetValue {
	sv := &SetValue{
		SetType:  setType,
		Elements: 0,
	}

	// Initialize map storage if needed for large enums
	if setType.StorageKind == types.SetStorageMap {
		sv.MapStore = make(map[int]bool)
	}

	return sv
}

// GetSetElementTypeName returns the element type name for error messages.
func (s *SetValue) GetSetElementTypeName() string {
	if s.SetType == nil || s.SetType.ElementType == nil {
		return "Unknown"
	}
	return s.SetType.ElementType.String()
}
