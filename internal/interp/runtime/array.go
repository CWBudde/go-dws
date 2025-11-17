package runtime

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
)

// ArrayValue represents an array value in DWScript.
// DWScript supports both static arrays (with fixed bounds) and dynamic arrays (resizable).
//
// Phase 3.5.4 - Type Migration: Migrated from internal/interp to runtime/
// to enable evaluator package to work with array values directly.
//
// Examples:
//   - Static: array[1..10] of Integer
//   - Dynamic: array of String
//
// Static arrays are pre-allocated with the specified size.
// Dynamic arrays start empty and can grow/shrink at runtime.
type ArrayValue struct {
	ArrayType *types.ArrayType // The array type metadata
	Elements  []Value          // The runtime elements (slice)
}

// Type returns "ARRAY".
func (a *ArrayValue) Type() string {
	return "ARRAY"
}

// String returns the string representation of the array.
// Format: [element1, element2, ...] or [] for empty array
func (a *ArrayValue) String() string {
	if len(a.Elements) == 0 {
		return "[]"
	}

	var elements []string
	for _, elem := range a.Elements {
		if elem != nil {
			elements = append(elements, elem.String())
		} else {
			elements = append(elements, "nil")
		}
	}

	return "[" + strings.Join(elements, ", ") + "]"
}

// ArrayElementInitializer is a function type for initializing array elements.
// This allows the interp package to provide custom initialization logic
// for complex element types (records, nested arrays, etc.) without creating
// circular dependencies.
//
// Phase 3.5.4: Used to break circular dependency between runtime and interp packages.
type ArrayElementInitializer func(elementType types.Type, index int) Value

// NewArrayValue creates a new ArrayValue with the given array type.
// For static arrays, pre-allocates elements (initialized to nil unless initializer provided).
// For dynamic arrays, creates an empty array.
//
// The optional initializer function is called for each element to provide custom
// initialization logic. If nil, elements are initialized to nil.
func NewArrayValue(arrayType *types.ArrayType, initializer ArrayElementInitializer) *ArrayValue {
	var elements []Value

	if arrayType.IsStatic() {
		// Static array: pre-allocate with size
		size := arrayType.Size()
		elements = make([]Value, size)

		// Task 9.56: For nested arrays, initialize each element as an array
		// Task 9.36: For record elements, initialize each element as a record
		if initializer != nil && arrayType.ElementType != nil {
			for i := 0; i < size; i++ {
				elements[i] = initializer(arrayType.ElementType, i)
			}
		}
		// Otherwise elements are nil (will be filled with zero values or explicit assignments)
	} else {
		// Dynamic array: start empty
		elements = make([]Value, 0)
	}

	return &ArrayValue{
		ArrayType: arrayType,
		Elements:  elements,
	}
}
