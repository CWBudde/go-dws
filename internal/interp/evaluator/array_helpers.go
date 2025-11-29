package evaluator

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
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

// ============================================================================
// Array Literal Direct Evaluation
// ============================================================================
//
// Task 3.5.83: Direct array literal evaluation without adapter.EvalNode().
// This eliminates double-evaluation by using pre-evaluated elements.
// ============================================================================

// evalArrayLiteralDirect evaluates an array literal expression directly.
// This is the main entry point for array literal evaluation without adapter delegation.
//
// Process:
//  1. Get type annotation from semanticInfo (if available)
//  2. Evaluate all elements (with expected type for nested arrays)
//  3. Infer type if no annotation exists
//  4. Coerce elements to target type
//  5. Validate static array bounds
//  6. Create ArrayValue directly via adapter.CreateArrayValue
//
// Returns the ArrayValue or an error Value.
func (e *Evaluator) evalArrayLiteralDirect(node *ast.ArrayLiteralExpression, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(node, "nil array literal")
	}

	// Task 3.5.105e: Check context first for type inference from assignment target
	if ctx.ArrayTypeContext() != nil {
		return e.evalArrayLiteralWithExpectedType(node, ctx.ArrayTypeContext(), ctx)
	}

	// Step 1: Get type annotation from semanticInfo
	arrayType := e.getArrayTypeFromAnnotation(node, ctx)

	// Step 2: Evaluate all elements
	elementCount := len(node.Elements)
	evaluatedElements := make([]Value, elementCount)
	elementTypes := make([]types.Type, elementCount)

	for idx, elem := range node.Elements {
		var val Value

		// If we have an expected array type and element is an array literal,
		// evaluate it with the expected element type for proper nested array typing.
		if arrayType != nil {
			if elemLit, ok := elem.(*ast.ArrayLiteralExpression); ok {
				if expectedElemArr, ok := arrayType.ElementType.(*types.ArrayType); ok {
					val = e.evalArrayLiteralWithExpectedType(elemLit, expectedElemArr, ctx)
				}
			}
		}

		// Regular evaluation if not a nested array literal
		if val == nil {
			val = e.Eval(elem, ctx)
		}

		if isError(val) {
			return val
		}
		evaluatedElements[idx] = val
		elementTypes[idx] = GetValueType(val)
	}

	// Step 3: Infer type if no annotation
	if arrayType == nil {
		inferred := e.inferArrayTypeFromElements(node, elementTypes)
		if inferred == nil {
			if elementCount == 0 {
				return e.newError(node, "cannot infer type for empty array literal")
			}
			return e.newError(node, "cannot determine array type for literal")
		}
		arrayType = inferred
	}

	// Step 4: Coerce elements to target type
	coercedElements, errVal := e.coerceElementsToType(arrayType, evaluatedElements, elementTypes, node)
	if errVal != nil {
		return errVal
	}

	// Step 5: Validate static array bounds
	if arrayType.IsStatic() {
		expectedSize := arrayType.Size()
		if elementCount != expectedSize {
			return e.newError(node, "array literal has %d elements, expected %d", elementCount, expectedSize)
		}
	}

	// Step 6: Create ArrayValue directly
	// Task 3.5.127: Create array value directly without adapter
	runtimeElements := make([]runtime.Value, len(coercedElements))
	for i, elem := range coercedElements {
		runtimeElements[i] = elem.(runtime.Value)
	}
	return &runtime.ArrayValue{ArrayType: arrayType, Elements: runtimeElements}
}

// getArrayTypeFromAnnotation retrieves the array type from semantic info annotations.
// Returns nil if no annotation exists or cannot be resolved.
//
// Task 3.5.106: Updated to take context for adapter-free type resolution.
func (e *Evaluator) getArrayTypeFromAnnotation(node *ast.ArrayLiteralExpression, ctx *ExecutionContext) *types.ArrayType {
	if e.semanticInfo == nil {
		return nil
	}

	typeAnnot := e.semanticInfo.GetType(node)
	if typeAnnot == nil || typeAnnot.Name == "" {
		return nil
	}

	// Resolve the type name to an ArrayType using context-aware resolution
	resolved, err := e.ResolveTypeWithContext(typeAnnot.Name, ctx)
	if err != nil {
		return nil
	}

	if arrayType, ok := resolved.(*types.ArrayType); ok {
		return arrayType
	}

	// Check underlying type for type aliases
	if underlying := types.GetUnderlyingType(resolved); underlying != nil {
		if arrayType, ok := underlying.(*types.ArrayType); ok {
			return arrayType
		}
	}

	return nil
}

// evalArrayLiteralWithExpectedType evaluates a nested array literal with an expected type.
// This ensures nested array literals get the correct element type from their parent.
func (e *Evaluator) evalArrayLiteralWithExpectedType(node *ast.ArrayLiteralExpression, expected *types.ArrayType, ctx *ExecutionContext) Value {
	if expected == nil {
		return e.evalArrayLiteralDirect(node, ctx)
	}

	// Temporarily set type annotation for this evaluation
	if e.semanticInfo == nil {
		// No semantic info - just use expected type directly
		return e.evalArrayLiteralWithType(node, expected, ctx)
	}

	// Save and restore type annotation
	prevType := e.semanticInfo.GetType(node)
	annotation := &ast.TypeAnnotation{Token: node.Token, Name: expected.String()}
	e.semanticInfo.SetType(node, annotation)

	result := e.evalArrayLiteralDirect(node, ctx)

	// Restore previous annotation
	if prevType != nil {
		e.semanticInfo.SetType(node, prevType)
	} else {
		e.semanticInfo.ClearType(node)
	}

	return result
}

// evalArrayLiteralWithType evaluates an array literal with a known array type.
// Used when semanticInfo is not available.
func (e *Evaluator) evalArrayLiteralWithType(node *ast.ArrayLiteralExpression, arrayType *types.ArrayType, ctx *ExecutionContext) Value {
	elementCount := len(node.Elements)
	evaluatedElements := make([]Value, elementCount)
	elementTypes := make([]types.Type, elementCount)

	for idx, elem := range node.Elements {
		var val Value

		// Handle nested array literals with expected element type
		if elemLit, ok := elem.(*ast.ArrayLiteralExpression); ok {
			if expectedElemArr, ok := arrayType.ElementType.(*types.ArrayType); ok {
				val = e.evalArrayLiteralWithType(elemLit, expectedElemArr, ctx)
			}
		}

		if val == nil {
			val = e.Eval(elem, ctx)
		}

		if isError(val) {
			return val
		}
		evaluatedElements[idx] = val
		elementTypes[idx] = GetValueType(val)
	}

	// Coerce elements
	coercedElements, errVal := e.coerceElementsToType(arrayType, evaluatedElements, elementTypes, node)
	if errVal != nil {
		return errVal
	}

	// Validate static array bounds
	if arrayType.IsStatic() {
		expectedSize := arrayType.Size()
		if elementCount != expectedSize {
			return e.newError(node, "array literal has %d elements, expected %d", elementCount, expectedSize)
		}
	}

	// Task 3.5.127: Create array value directly without adapter
	runtimeElements := make([]runtime.Value, len(coercedElements))
	for i, elem := range coercedElements {
		runtimeElements[i] = elem.(runtime.Value)
	}
	return &runtime.ArrayValue{ArrayType: arrayType, Elements: runtimeElements}
}

// inferArrayTypeFromElements infers an array type from evaluated element types.
// Returns a dynamic array type with the common element type, or nil if inference fails.
func (e *Evaluator) inferArrayTypeFromElements(node *ast.ArrayLiteralExpression, elementTypes []types.Type) *types.ArrayType {
	if len(elementTypes) == 0 {
		return nil
	}

	var inferred types.Type

	for idx, elemType := range elementTypes {
		if elemType == nil {
			continue
		}

		underlying := types.GetUnderlyingType(elemType)
		if underlying == types.NIL {
			continue
		}

		if inferred == nil {
			inferred = underlying
			continue
		}

		if inferred.Equals(underlying) {
			continue
		}

		// Numeric promotion: Integer + Float → Float
		if inferred.Equals(types.INTEGER) && underlying.Equals(types.FLOAT) {
			inferred = types.FLOAT
			continue
		}
		if inferred.Equals(types.FLOAT) && underlying.Equals(types.INTEGER) {
			continue
		}

		// Incompatible types - could return Variant or error
		// For now, return nil to match existing behavior
		_ = idx // Suppress unused variable warning
		return nil
	}

	if inferred == nil {
		return nil
	}

	// Create a static array type for literals (value semantics)
	size := len(node.Elements)
	if size == 0 {
		return types.NewDynamicArrayType(types.GetUnderlyingType(inferred))
	}
	return types.NewStaticArrayType(types.GetUnderlyingType(inferred), 0, size-1)
}

// coerceElementsToType coerces all elements to the target array element type.
// Handles Integer→Float promotion and Variant boxing.
func (e *Evaluator) coerceElementsToType(arrayType *types.ArrayType, values []Value, valueTypes []types.Type, node *ast.ArrayLiteralExpression) ([]Value, Value) {
	coerced := make([]Value, len(values))

	elementType := arrayType.ElementType
	if elementType == nil {
		return nil, e.newError(node, "array literal has no element type information")
	}
	underlyingElementType := types.GetUnderlyingType(elementType)

	for idx, val := range values {
		var valType types.Type
		if idx < len(valueTypes) && valueTypes[idx] != nil {
			valType = types.GetUnderlyingType(valueTypes[idx])
		}

		// Box values when expected element type is Variant
		if underlyingElementType.Equals(types.VARIANT) {
			coerced[idx] = e.adapter.BoxVariant(val)
			continue
		}

		// Handle nil values
		if val != nil && val.Type() == "NIL" {
			switch underlyingElementType.TypeKind() {
			case "CLASS", "INTERFACE", "ARRAY":
				coerced[idx] = val
				continue
			default:
				elemNode := node
				if idx < len(node.Elements) {
					elemNode = &ast.ArrayLiteralExpression{Elements: []ast.Expression{node.Elements[idx]}}
				}
				return nil, e.newError(elemNode, "cannot assign nil to %s", underlyingElementType.String())
			}
		}

		if valType == nil {
			elemNode := node
			if idx < len(node.Elements) {
				elemNode = &ast.ArrayLiteralExpression{Elements: []ast.Expression{node.Elements[idx]}}
			}
			return nil, e.newError(elemNode, "cannot determine type for array element %d", idx+1)
		}

		// Exact type match
		if underlyingElementType.Equals(valType) {
			coerced[idx] = val
			continue
		}

		// Integer → Float promotion
		if underlyingElementType.Equals(types.FLOAT) && valType.Equals(types.INTEGER) {
			// Convert integer to float directly
			coerced[idx] = e.castToFloat(val)
			continue
		}

		// Array compatibility check
		if valType.TypeKind() == "ARRAY" && underlyingElementType.TypeKind() == "ARRAY" {
			if types.IsCompatible(valType, underlyingElementType) || types.IsCompatible(underlyingElementType, valType) {
				coerced[idx] = val
				continue
			}
		}

		// General compatibility check
		if types.IsCompatible(valType, underlyingElementType) {
			coerced[idx] = val
			continue
		}

		// Incompatible type
		elemNode := node
		if idx < len(node.Elements) {
			elemNode = &ast.ArrayLiteralExpression{Elements: []ast.Expression{node.Elements[idx]}}
		}
		return nil, e.newError(elemNode, "array element %d has incompatible type (got %s, expected %s)",
			idx+1, val.Type(), underlyingElementType.String())
	}

	return coerced, nil
}

// ============================================================================
// Array Helper Method Implementations
// ============================================================================
// Task 3.5.102e: Migrate array helper methods from Interpreter to Evaluator.
//
// These implementations avoid the adapter by directly manipulating runtime values.
// The goal is to remove EvalNode delegation for common array operations.
//
// Split into sub-tasks:
// - 3.5.102e1: Array properties (Length, Count, High, Low)
// - 3.5.102e2: Simple array methods (Add, Push, Pop, Swap, Delete)
// - 3.5.102e3: Join methods (Join, string array Join)

// evalArrayHelper evaluates a built-in array helper method directly in the evaluator.
// Returns the result value, or nil if this helper is not handled here (should fall through
// to the adapter).
func (e *Evaluator) evalArrayHelper(spec string, selfValue Value, args []Value, node ast.Node) Value {
	switch spec {
	// Task 3.5.102e1: Array properties
	case "__array_length", "__array_count":
		return e.evalArrayLengthHelper(selfValue, args, node)
	case "__array_high":
		return e.evalArrayHigh(selfValue, args, node)
	case "__array_low":
		return e.evalArrayLow(selfValue, args, node)

	// Task 3.5.102e2: Simple array methods
	case "__array_add":
		return e.evalArrayAdd(selfValue, args, node)
	case "__array_push":
		return e.evalArrayPush(selfValue, args, node)
	case "__array_pop":
		return e.evalArrayPop(selfValue, args, node)
	case "__array_swap":
		return e.evalArraySwap(selfValue, args, node)
	case "__array_delete":
		return e.evalArrayDelete(selfValue, args, node)

	// Task 3.5.102e3: Join methods
	case "__array_join":
		return e.evalArrayJoinHelper(selfValue, args, node)
	case "__string_array_join":
		return e.evalStringArrayJoin(selfValue, args, node)

	default:
		// Not an array helper we handle - return nil to signal fallthrough to adapter
		return nil
	}
}

// ============================================================================
// Task 3.5.102e1: Array Properties
// ============================================================================

// evalArrayLengthHelper implements Array.Length and Array.Count property.
// Returns the number of elements in the array.
func (e *Evaluator) evalArrayLengthHelper(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 0 {
		return e.newError(node, "Array.Length property does not take arguments")
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.Length property requires array receiver")
	}

	return &runtime.IntegerValue{Value: int64(len(arrVal.Elements))}
}

// evalArrayHigh implements Array.High property.
// Returns the highest valid index of the array.
// For static arrays, returns the declared high bound.
// For dynamic arrays, returns Length - 1.
func (e *Evaluator) evalArrayHigh(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 0 {
		return e.newError(node, "Array.High property does not take arguments")
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.High property requires array receiver")
	}

	if arrVal.ArrayType != nil && arrVal.ArrayType.IsStatic() {
		return &runtime.IntegerValue{Value: int64(*arrVal.ArrayType.HighBound)}
	}
	return &runtime.IntegerValue{Value: int64(len(arrVal.Elements) - 1)}
}

// evalArrayLow implements Array.Low property.
// Returns the lowest valid index of the array.
// For static arrays, returns the declared low bound.
// For dynamic arrays, returns 0.
func (e *Evaluator) evalArrayLow(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 0 {
		return e.newError(node, "Array.Low property does not take arguments")
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.Low property requires array receiver")
	}

	if arrVal.ArrayType != nil && arrVal.ArrayType.IsStatic() {
		return &runtime.IntegerValue{Value: int64(*arrVal.ArrayType.LowBound)}
	}
	return &runtime.IntegerValue{Value: 0}
}

// ============================================================================
// Task 3.5.102e2: Simple Array Methods
// ============================================================================

// evalArrayAdd implements Array.Add(value) method.
// Adds an element to a dynamic array.
func (e *Evaluator) evalArrayAdd(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 1 {
		return e.newError(node, "Array.Add expects exactly 1 argument")
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.Add requires array receiver")
	}

	// Check if it's a dynamic array (static arrays cannot use Add)
	if arrVal.ArrayType != nil && !arrVal.ArrayType.IsDynamic() {
		return e.newError(node, "Add() can only be used with dynamic arrays, not static arrays")
	}

	// Append the element
	arrVal.Elements = append(arrVal.Elements, args[0])

	return &runtime.NilValue{}
}

// evalArrayPush implements Array.Push(value) method.
// Pushes an element onto a dynamic array (same as Add, but copies records).
func (e *Evaluator) evalArrayPush(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 1 {
		return e.newError(node, "Array.Push expects exactly 1 argument")
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.Push requires array receiver")
	}

	// Check if it's a dynamic array (static arrays cannot use Push)
	if arrVal.ArrayType != nil && !arrVal.ArrayType.IsDynamic() {
		return e.newError(node, "Push() can only be used with dynamic arrays, not static arrays")
	}

	valueToAdd := args[0]

	// If pushing a record, make a copy to avoid aliasing issues
	// Records are value types and should be copied when added to collections
	// Check if value implements Copyable interface (RecordValue does)
	if copyable, ok := valueToAdd.(interface{ Copy() Value }); ok {
		valueToAdd = copyable.Copy()
	}

	arrVal.Elements = append(arrVal.Elements, valueToAdd)

	return &runtime.NilValue{}
}

// evalArrayPop implements Array.Pop() method.
// Removes and returns the last element from a dynamic array.
func (e *Evaluator) evalArrayPop(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 0 {
		return e.newError(node, "Array.Pop expects no arguments, got %d", len(args))
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.Pop requires array receiver")
	}

	// Check if it's a dynamic array (static arrays cannot use Pop)
	if arrVal.ArrayType != nil && !arrVal.ArrayType.IsDynamic() {
		return e.newError(node, "Pop() can only be used with dynamic arrays, not static arrays")
	}

	// Check if array is empty
	if len(arrVal.Elements) == 0 {
		return e.newError(node, "Pop() called on empty array")
	}

	// Get the last element
	lastElement := arrVal.Elements[len(arrVal.Elements)-1]

	// Remove the last element
	arrVal.Elements = arrVal.Elements[:len(arrVal.Elements)-1]

	return lastElement
}

// evalArraySwap implements Array.Swap(i, j) method.
// Swaps two elements in the array.
func (e *Evaluator) evalArraySwap(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 2 {
		return e.newError(node, "Array.Swap expects exactly 2 arguments, got %d", len(args))
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.Swap requires array receiver")
	}

	// Get index i
	iInt, ok := args[0].(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "Array.Swap first argument must be Integer, got %s", args[0].Type())
	}
	iIdx := int(iInt.Value)

	// Get index j
	jInt, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "Array.Swap second argument must be Integer, got %s", args[1].Type())
	}
	jIdx := int(jInt.Value)

	// Validate indices
	arrayLen := len(arrVal.Elements)
	if iIdx < 0 || iIdx >= arrayLen {
		return e.newError(node, "Array.Swap first index %d out of bounds (0..%d)", iIdx, arrayLen-1)
	}
	if jIdx < 0 || jIdx >= arrayLen {
		return e.newError(node, "Array.Swap second index %d out of bounds (0..%d)", jIdx, arrayLen-1)
	}

	// Swap elements
	arrVal.Elements[iIdx], arrVal.Elements[jIdx] = arrVal.Elements[jIdx], arrVal.Elements[iIdx]

	return &runtime.NilValue{}
}

// evalArrayDelete implements Array.Delete(index) or Array.Delete(index, count) method.
// Removes elements from a dynamic array.
func (e *Evaluator) evalArrayDelete(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) < 1 || len(args) > 2 {
		return e.newError(node, "Array.Delete expects 1 or 2 arguments, got %d", len(args))
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.Delete requires array receiver")
	}

	// Check if it's a dynamic array (static arrays cannot use Delete)
	if arrVal.ArrayType != nil && !arrVal.ArrayType.IsDynamic() {
		return e.newError(node, "Delete() can only be used with dynamic arrays, not static arrays")
	}

	// Get the index
	indexInt, ok := args[0].(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "Array.Delete index must be Integer, got %s", args[0].Type())
	}
	index := int(indexInt.Value)

	// Get the count (default to 1 if not specified)
	count := 1
	if len(args) == 2 {
		countInt, ok := args[1].(*runtime.IntegerValue)
		if !ok {
			return e.newError(node, "Array.Delete count must be Integer, got %s", args[1].Type())
		}
		count = int(countInt.Value)
	}

	// Validate index and count
	arrayLen := len(arrVal.Elements)
	if index < 0 || index >= arrayLen {
		return e.newError(node, "Array.Delete index %d out of bounds (0..%d)", index, arrayLen-1)
	}
	if count < 0 {
		return e.newError(node, "Array.Delete count must be non-negative, got %d", count)
	}

	// Calculate end index (don't go beyond array length)
	endIndex := index + count
	if endIndex > arrayLen {
		endIndex = arrayLen
	}

	// Delete elements by slicing
	arrVal.Elements = append(arrVal.Elements[:index], arrVal.Elements[endIndex:]...)

	return &runtime.NilValue{}
}

// ============================================================================
// Task 3.5.102e3: Join Methods
// ============================================================================

// evalArrayJoinHelper implements Array.Join(separator) method.
// Joins array elements into a string using the separator.
func (e *Evaluator) evalArrayJoinHelper(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 1 {
		return e.newError(node, "Array.Join expects exactly 1 argument")
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.Join requires array receiver")
	}

	sep, ok := args[0].(*runtime.StringValue)
	if !ok {
		return e.newError(node, "Array.Join separator must be String, got %s", args[0].Type())
	}

	var b strings.Builder
	for idx, elem := range arrVal.Elements {
		if idx > 0 {
			b.WriteString(sep.Value)
		}
		if elem == nil {
			continue
		}
		b.WriteString(elem.String())
	}

	return &runtime.StringValue{Value: b.String()}
}

// evalStringArrayJoin implements string array Join(separator) method.
// Joins string array elements into a string using the separator.
func (e *Evaluator) evalStringArrayJoin(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 1 {
		return e.newError(node, "String array Join expects exactly 1 argument")
	}

	separator, ok := args[0].(*runtime.StringValue)
	if !ok {
		return e.newError(node, "Join separator must be String, got %s", args[0].Type())
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Join helper requires string array receiver")
	}

	var builder strings.Builder
	for idx, elem := range arrVal.Elements {
		strElem, ok := elem.(*runtime.StringValue)
		if !ok {
			return e.newError(node, "Join requires elements of type String")
		}
		if idx > 0 {
			builder.WriteString(separator.Value)
		}
		builder.WriteString(strElem.Value)
	}

	return &runtime.StringValue{Value: builder.String()}
}
