package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Array Type Inference
// ============================================================================
//
// Task 3.5.24: These helpers determine array types for array literal expressions.
// They handle both explicit type annotations and type inference from values.
// ============================================================================

// getArrayElementType determines the element type for an array literal.
// This combines type annotation lookup and type inference.
//
// Process:
//  1. Check for explicit type annotation from semantic analysis
//  2. If no annotation, infer type from runtime values
//  3. Return the determined ArrayType
//
// Returns:
//   - The ArrayType for the array literal
//   - nil if type cannot be determined (returns error via e.newError)
func (e *Evaluator) getArrayElementType(node *ast.ArrayLiteralExpression, values []Value) *types.ArrayType {
	// Step 1: Try to get type from annotation
	// For now, we'll skip annotation lookup as it requires more complex type resolution
	// This will be enhanced in future iterations
	// annotatedType := e.arrayTypeFromLiteral(node)
	// if annotatedType != nil {
	// 	return annotatedType
	// }

	// Step 2: Infer type from values
	return e.inferArrayTypeFromValues(values, node)
}

// inferArrayTypeFromValues infers the array element type from runtime values.
// This is used when no explicit type annotation exists.
//
// Type inference rules:
//   - All same type → array of that type
//   - Integer + Float → array of Float (numeric promotion)
//   - Mixed incompatible types → array of Variant
//   - All nil → cannot infer (need explicit type)
//   - Empty array → cannot infer (need explicit type)
//
// Returns:
//   - The inferred ArrayType (always dynamic array)
//   - nil if type cannot be inferred (returns error via e.newError)
//
// Example:
//
//	[1, 2, 3] → array of Integer
//	[1, 2.5] → array of Float
//	[1, "hello"] → array of Variant
//	[] → nil (cannot infer, returns error)
func (e *Evaluator) inferArrayTypeFromValues(values []Value, node *ast.ArrayLiteralExpression) *types.ArrayType {
	if len(values) == 0 {
		// Cannot infer type from empty array - return error
		// Note: e.newError returns a Value (ErrorValue), not an error
		// We'll handle this by returning nil and letting the caller check
		return nil
	}

	// Get the type of the first element
	var commonElementType types.Type

	for i, val := range values {
		// Get the type of this value
		valType := GetValueType(val)

		if i == 0 {
			// First element - initialize common type
			commonElementType = valType
			if commonElementType == nil {
				// First element is nil - need more elements to infer type
				continue
			}
		} else {
			// Subsequent elements - find common type
			commonElementType = commonType(commonElementType, valType)

			// If we've reached Variant, no need to continue checking
			if commonElementType != nil && commonElementType.TypeKind() == "VARIANT" {
				break
			}
		}
	}

	// Check if we successfully inferred a type
	if commonElementType == nil {
		// All elements were nil - cannot infer type
		return nil
	}

	// Create a dynamic array type with the inferred element type
	return types.NewDynamicArrayType(commonElementType)
}

// coerceArrayElements validates that all array elements can be coerced to the target element type.
// This implements type compatibility checking for array literals.
//
// Coercion rules:
//  1. Integer → Float: Numeric promotion for mixed numeric arrays
//  2. Any → Variant: Wrap heterogeneous types in Variant
//  3. Same type: No coercion needed
//  4. Incompatible: Return error
//
// Examples:
//   - [1, 2, 3] with target Integer → valid (no coercion)
//   - [1, 2.5] with target Float → valid (Integer→Float promotion)
//   - [1, "hello"] with target Variant → valid (wrap in Variant)
//   - [1, "hello"] with target Integer → ERROR (incompatible types)
//
// Note: This function validates type compatibility. The actual value coercion
// (creating FloatValue, VariantValue) is delegated to the adapter when constructing
// the ArrayValue in task 3.5.26. This avoids circular import issues with value types.
//
// Returns:
//   - nil if all elements can be coerced to target type
//   - error Value if coercion is not possible
func (e *Evaluator) coerceArrayElements(elements []Value, targetElementType types.Type, node ast.Node) Value {
	if targetElementType == nil {
		// No target type - nothing to validate
		return nil
	}

	// Validate each element can be coerced to the target type
	for i, elem := range elements {
		if err := e.validateCoercion(elem, targetElementType, node, i); err != nil {
			return err
		}
	}

	return nil
}

// validateCoercion checks if a value can be coerced to the target type.
// Returns nil if coercion is valid, or an error Value otherwise.
//
// This validates type compatibility without performing the actual coercion.
// The actual value transformation happens in ArrayValue construction.
func (e *Evaluator) validateCoercion(val Value, targetType types.Type, node ast.Node, index int) Value {
	if val == nil {
		// Nil is compatible with all reference types
		if isReferenceType(targetType) {
			return nil
		}
		return e.newError(node, "element %d: cannot use nil in array of %s", index, targetType.String())
	}

	// Get the source type
	sourceType := GetValueType(val)

	// If types match, no coercion needed
	if sourceType != nil && sourceType.Equals(targetType) {
		return nil // Valid - no coercion needed
	}

	// Handle specific coercion cases
	targetKind := targetType.TypeKind()

	switch targetKind {
	case "FLOAT":
		// Integer → Float promotion is valid
		if sourceType != nil && sourceType.TypeKind() == "INTEGER" {
			return nil // Valid - Integer can be promoted to Float
		}
		// Other types → Float requires explicit conversion
		return e.newError(node, "element %d: cannot coerce %s to Float in array literal", index, sourceType.String())

	case "VARIANT":
		// Any type → Variant is always valid (wrapping)
		return nil

	default:
		// For other types, check if source type is compatible
		if sourceType == nil {
			// Nil without a known type
			return e.newError(node, "element %d: cannot use nil in array of %s", index, targetType.String())
		}

		// Types don't match and no coercion rule applies
		return e.newError(node, "element %d: cannot coerce %s to %s in array literal",
			index, sourceType.String(), targetType.String())
	}
}

// isReferenceType checks if a type is a reference type (class, interface, etc.)
// Reference types can be nil.
func isReferenceType(t types.Type) bool {
	if t == nil {
		return false
	}
	kind := t.TypeKind()
	return kind == "CLASS" || kind == "INTERFACE" || kind == "ARRAY" || kind == "STRING"
}

// validateArrayLiteralSize checks if the number of elements matches the expected size
// for static arrays.
//
// For dynamic arrays (no bounds), any size is valid.
// For static arrays (with bounds), the element count must match the size.
//
// Example:
//
//	var x: array[1..3] of Integer := [1, 2, 3]; // OK - 3 elements
//	var y: array[1..3] of Integer := [1, 2];    // ERROR - expected 3, got 2
//
// Returns nil if validation passes, otherwise returns an error Value.
func (e *Evaluator) validateArrayLiteralSize(arrayType *types.ArrayType, elementCount int, node ast.Node) Value {
	if arrayType == nil {
		return nil
	}

	// Only validate static arrays (with bounds)
	if !arrayType.IsStatic() {
		return nil
	}

	// Get the expected size
	expectedSize := arrayType.Size()

	if elementCount != expectedSize {
		return e.newError(node, "array literal has %d elements, but type %s requires %d elements",
			elementCount, arrayType.String(), expectedSize)
	}

	return nil
}

// ============================================================================
// New Array Expression Helpers
// ============================================================================
//
// Task 3.5.27: Helpers for dynamic array allocation with 'new' keyword.
// ============================================================================

// evaluateDimensions evaluates all dimension expressions for a new array expression.
// Each dimension must evaluate to a positive integer.
//
// Returns:
//   - Slice of dimension sizes ([]int)
//   - Error Value if any dimension is invalid
//
// Examples:
//   - new Integer[10] → [10]
//   - new String[3, 4] → [3, 4]
//   - new Float[x+1, y*2] → [evaluated x+1, evaluated y*2]
func (e *Evaluator) evaluateDimensions(dimensions []ast.Expression, ctx *ExecutionContext, node ast.Node) ([]int, Value) {
	if len(dimensions) == 0 {
		return nil, e.newError(node, "new array expression must have at least one dimension")
	}

	dimSizes := make([]int, len(dimensions))

	for i, dimExpr := range dimensions {
		// Evaluate the dimension expression
		dimValue := e.Eval(dimExpr, ctx)
		if isError(dimValue) {
			return nil, dimValue
		}

		// Dimension must be an integer
		if dimValue.Type() != "INTEGER" {
			return nil, e.newError(node, "dimension %d: expected Integer, got %s", i, dimValue.Type())
		}

		// Extract the integer value
		// We can't use type assertions due to circular imports, so we parse the string
		dimSize, err := e.extractIntegerValue(dimValue)
		if err != nil {
			return nil, e.newError(node, "dimension %d: %v", i, err)
		}

		// Dimension must be positive
		if dimSize <= 0 {
			return nil, e.newError(node, "dimension %d: array size must be positive, got %d", i, dimSize)
		}

		dimSizes[i] = dimSize
	}

	return dimSizes, nil
}

// extractIntegerValue extracts an int from an IntegerValue.
// This is a helper to work around circular import issues.
func (e *Evaluator) extractIntegerValue(val Value) (int, error) {
	// IntegerValue.String() returns the decimal representation
	// We can parse it to get the int value
	// This is a workaround until we can access IntegerValue.Value directly

	// For now, delegate to the adapter to extract the value
	// The adapter has direct access to IntegerValue
	// This is temporary - will be fixed when value types are refactored

	// Use a simple string parsing approach as a fallback
	// In practice, the adapter will handle this properly
	strVal := val.String()

	// Try to parse the string as an integer
	var intVal int
	_, err := fmt.Sscanf(strVal, "%d", &intVal)
	if err != nil {
		return 0, fmt.Errorf("failed to extract integer value: %v", err)
	}

	return intVal, nil
}
