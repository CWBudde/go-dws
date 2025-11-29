package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Index Operations
// ============================================================================
//
// Task 3.5.81: Index operation helpers for VisitIndexExpression.
// Migrated from interp/array.go to support direct evaluation without adapter.
// ============================================================================

// IndexArray performs array indexing with bounds checking.
// Returns the element at the given index or an error if out of bounds.
//
// For static arrays, the index is checked against low/high bounds and
// converted to a physical index. For dynamic arrays, zero-based indexing
// is used.
func (e *Evaluator) IndexArray(arr *runtime.ArrayValue, index int, node ast.Node) Value {
	if arr.ArrayType == nil {
		return e.newError(node, "array has no type information")
	}

	// Convert logical index to physical index
	var physicalIndex int
	if arr.ArrayType.IsStatic() {
		// Static array: check bounds and adjust for low bound
		lowBound := *arr.ArrayType.LowBound
		highBound := *arr.ArrayType.HighBound

		if index < lowBound || index > highBound {
			return e.newError(node, "index out of bounds: %d (bounds are %d..%d)", index, lowBound, highBound)
		}

		physicalIndex = index - lowBound
	} else {
		// Dynamic array: zero-based indexing
		if index < 0 || index >= len(arr.Elements) {
			return e.newError(node, "index out of bounds: %d (array length is %d)", index, len(arr.Elements))
		}

		physicalIndex = index
	}

	// Check physical bounds
	if physicalIndex < 0 || physicalIndex >= len(arr.Elements) {
		return e.newError(node, "index out of bounds: physical index %d, length %d", physicalIndex, len(arr.Elements))
	}

	// Return the element
	elem := arr.Elements[physicalIndex]
	if elem == nil {
		// Return properly typed zero value for uninitialized elements
		return e.getZeroValueForType(arr.ArrayType.ElementType)
	}

	return elem
}

// IndexString performs string indexing (returns a single-character string).
// DWScript strings are 1-indexed.
func (e *Evaluator) IndexString(str *runtime.StringValue, index int, node ast.Node) Value {
	// DWScript strings are 1-indexed
	// Use rune-based indexing to handle UTF-8 correctly
	strLen := RuneLength(str.Value)
	if index < 1 || index > strLen {
		return e.newError(node, "string index out of bounds: %d (string length is %d)", index, strLen)
	}

	// Get the character at the given position
	char, ok := RuneAt(str.Value, index)
	if !ok {
		return e.newError(node, "string index out of bounds: %d", index)
	}
	return &runtime.StringValue{Value: string(char)}
}

// Note: JSON indexing is delegated to adapter since JSONValue and VariantValue
// are in the interp package and we can't import them here without circular deps.
// JSON handling will use val.Type() == "JSON" check and delegate to adapter.

// getZeroValueForType returns the zero/default value for a given type.
// This is used when accessing uninitialized array elements and record field initialization.
// Task 3.5.128e: Extended to support all DWScript types for record field initialization.
func (e *Evaluator) getZeroValueForType(t types.Type) runtime.Value {
	if t == nil {
		return &runtime.NilValue{}
	}

	switch t.TypeKind() {
	case "INTEGER":
		return &runtime.IntegerValue{Value: 0}
	case "FLOAT":
		return &runtime.FloatValue{Value: 0.0}
	case "STRING":
		return &runtime.StringValue{Value: ""}
	case "BOOLEAN":
		return &runtime.BooleanValue{Value: false}
	case "ARRAY":
		// For array types, create an empty array with proper type
		if arrayType, ok := t.(*types.ArrayType); ok {
			return runtime.NewArrayValue(arrayType, nil)
		}
		return &runtime.NilValue{}
	case "RECORD":
		// Task 3.5.128e: Recursively create nested records with zero-initialized fields
		if recordType, ok := t.(*types.RecordType); ok {
			// Look up metadata for nested record (returns any, need type assertion)
			metadataAny := e.typeSystem.LookupRecordMetadata(recordType.Name)
			var nestedMetadata *runtime.RecordMetadata
			if metadataAny != nil {
				if md, ok := metadataAny.(*runtime.RecordMetadata); ok {
					nestedMetadata = md
				}
			}

			// Create zero-initialized nested record
			zeroInit := func(nestedFieldName string, nestedFieldType types.Type) runtime.Value {
				return e.getZeroValueForType(nestedFieldType)
			}
			return runtime.NewRecordValueWithInitializer(recordType, nestedMetadata, zeroInit)
		}
		return &runtime.NilValue{}
	case "INTERFACE":
		// Task 3.5.128e: Interface fields initialize as nil (use adapter for proper InterfaceInstance)
		// InterfaceInstance is in interp package, so we return nil here
		// The adapter will handle interface field initialization if needed
		return &runtime.NilValue{}
	case "CLASS":
		// Task 3.5.128e: Class fields initialize as nil
		if classType, ok := t.(*types.ClassType); ok {
			return &runtime.NilValue{ClassType: classType.Name}
		}
		return &runtime.NilValue{}
	case "VARIANT":
		// Task 3.5.128e: Variant fields initialize as nil (VariantValue is in interp package)
		// For now, return nil - the adapter will handle variant initialization if needed
		return &runtime.NilValue{}
	default:
		// For other types, return nil
		return &runtime.NilValue{}
	}
}

// ExtractIntegerIndex extracts an integer index from a Value.
// Returns the index and true if successful, or 0 and false if the value
// is not an integer or enum type.
func ExtractIntegerIndex(indexVal Value) (int, bool) {
	switch iv := indexVal.(type) {
	case *runtime.IntegerValue:
		return int(iv.Value), true
	case *runtime.BooleanValue:
		if iv.Value {
			return 1, true
		}
		return 0, true
	case *runtime.EnumValue:
		return iv.OrdinalValue, true
	default:
		return 0, false
	}
}

// ============================================================================
// Multi-Dimensional Array Construction
// ============================================================================
//
// Task 3.5.82: Helpers for creating multi-dimensional arrays from new expressions.
// ============================================================================

// CreateMultiDimArray creates a multi-dimensional array with the given dimensions.
// For 1D arrays, creates a single array with the specified size.
// For multi-dimensional arrays, recursively creates nested arrays.
//
// Example:
//
//	new Integer[10] → single 1D array with 10 elements
//	new String[3, 4] → 3x4 nested arrays
func (e *Evaluator) CreateMultiDimArray(elementType types.Type, dimensions []int) *runtime.ArrayValue {
	if len(dimensions) == 0 {
		// This shouldn't happen, but handle gracefully
		return &runtime.ArrayValue{
			ArrayType: types.NewDynamicArrayType(elementType),
			Elements:  []runtime.Value{},
		}
	}

	size := dimensions[0]

	if len(dimensions) == 1 {
		// Base case: 1D array
		arrayType := types.NewDynamicArrayType(elementType)

		// Create elements filled with zero values
		elements := make([]runtime.Value, size)
		for idx := 0; idx < size; idx++ {
			elements[idx] = e.getZeroValueForType(elementType)
		}

		return &runtime.ArrayValue{
			ArrayType: arrayType,
			Elements:  elements,
		}
	}

	// Recursive case: multi-dimensional array
	// The element type for this level is an array of the remaining dimensions
	innerElementType := buildArrayTypeForDimensions(elementType, dimensions[1:])

	// Create the outer array type
	arrayType := types.NewDynamicArrayType(innerElementType)

	// Create elements, each being an array of the remaining dimensions
	elements := make([]runtime.Value, size)
	for idx := 0; idx < size; idx++ {
		elements[idx] = e.CreateMultiDimArray(elementType, dimensions[1:])
	}

	return &runtime.ArrayValue{
		ArrayType: arrayType,
		Elements:  elements,
	}
}

// buildArrayTypeForDimensions builds an array type for the given dimensions.
// For example, dimensions [3, 4] with elementType Integer produces:
// array of array of Integer
func buildArrayTypeForDimensions(elementType types.Type, dimensions []int) types.Type {
	if len(dimensions) == 0 {
		return elementType
	}

	// Build from innermost to outermost
	currentType := elementType
	for range dimensions {
		currentType = types.NewDynamicArrayType(currentType)
	}

	return currentType
}
