package evaluator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Array Type Inference
// ============================================================================

// getArrayElementType determines the element type for an array literal.
// Checks for explicit type annotation, otherwise infers from runtime values.
func (e *Evaluator) getArrayElementType(node *ast.ArrayLiteralExpression, values []Value) *types.ArrayType {
	// TODO: Add annotation lookup when type resolution is implemented
	return e.inferArrayTypeFromValues(values, node)
}

// inferArrayTypeFromValues infers the array element type from runtime values.
// Rules: same type → that type, Integer+Float → Float, mixed → Variant, empty/all nil → error.
func (e *Evaluator) inferArrayTypeFromValues(values []Value, node *ast.ArrayLiteralExpression) *types.ArrayType {
	if len(values) == 0 {
		return nil
	}

	var commonElementType types.Type

	for i, val := range values {
		valType := GetValueType(val)

		if i == 0 {
			commonElementType = valType
			if commonElementType == nil {
				continue
			}
		} else {
			commonElementType = commonType(commonElementType, valType)

			if commonElementType != nil && commonElementType.TypeKind() == "VARIANT" {
				break
			}
		}
	}

	if commonElementType == nil {
		return nil
	}

	return types.NewDynamicArrayType(commonElementType)
}

// coerceArrayElements validates that all array elements can be coerced to the target element type.
// Validates compatibility without performing actual value transformation.
func (e *Evaluator) coerceArrayElements(elements []Value, targetElementType types.Type, node ast.Node) Value {
	if targetElementType == nil {
		return nil
	}

	for i, elem := range elements {
		if err := e.validateCoercion(elem, targetElementType, node, i); err != nil {
			return err
		}
	}

	return nil
}

// validateCoercion checks if a value can be coerced to the target type.
func (e *Evaluator) validateCoercion(val Value, targetType types.Type, node ast.Node, index int) Value {
	if val == nil {
		if isReferenceType(targetType) {
			return nil
		}
		return e.newError(node, "element %d: cannot use nil in array of %s", index, targetType.String())
	}

	sourceType := GetValueType(val)

	if sourceType != nil && sourceType.Equals(targetType) {
		return nil
	}

	targetKind := targetType.TypeKind()

	switch targetKind {
	case "FLOAT":
		if sourceType != nil && sourceType.TypeKind() == "INTEGER" {
			return nil
		}
		return e.newError(node, "element %d: cannot coerce %s to Float in array literal", index, sourceType.String())

	case "VARIANT":
		return nil

	default:
		if sourceType == nil {
			return e.newError(node, "element %d: cannot use nil in array of %s", index, targetType.String())
		}

		return e.newError(node, "element %d: cannot coerce %s to %s in array literal",
			index, sourceType.String(), targetType.String())
	}
}

// isReferenceType checks if a type can be nil (class, interface, array, string).
func isReferenceType(t types.Type) bool {
	if t == nil {
		return false
	}
	kind := t.TypeKind()
	return kind == "CLASS" || kind == "INTERFACE" || kind == "ARRAY" || kind == "STRING"
}

// validateArrayLiteralSize checks element count matches static array bounds.
func (e *Evaluator) validateArrayLiteralSize(arrayType *types.ArrayType, elementCount int, node ast.Node) Value {
	if arrayType == nil || !arrayType.IsStatic() {
		return nil
	}

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

// evaluateDimensions evaluates dimension expressions for array allocation.
// Each dimension must be a positive integer.
func (e *Evaluator) evaluateDimensions(dimensions []ast.Expression, ctx *ExecutionContext, node ast.Node) ([]int, Value) {
	if len(dimensions) == 0 {
		return nil, e.newError(node, "new array expression must have at least one dimension")
	}

	dimSizes := make([]int, len(dimensions))

	for i, dimExpr := range dimensions {
		dimValue := e.Eval(dimExpr, ctx)
		if isError(dimValue) {
			return nil, dimValue
		}

		if dimValue.Type() != "INTEGER" {
			return nil, e.newError(node, "dimension %d: expected Integer, got %s", i, dimValue.Type())
		}

		dimSize, err := e.extractIntegerValue(dimValue)
		if err != nil {
			return nil, e.newError(node, "dimension %d: %v", i, err)
		}

		if dimSize <= 0 {
			return nil, e.newError(node, "dimension %d: array size must be positive, got %d", i, dimSize)
		}

		dimSizes[i] = dimSize
	}

	return dimSizes, nil
}

// extractIntegerValue extracts an int from an IntegerValue via string parsing.
func (e *Evaluator) extractIntegerValue(val Value) (int, error) {
	strVal := val.String()

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

// evalArrayLiteralDirect evaluates an array literal without adapter delegation.
// Gets type from annotation or context, evaluates elements, coerces to target type, validates bounds.
func (e *Evaluator) evalArrayLiteralDirect(node *ast.ArrayLiteralExpression, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(node, "nil array literal")
	}

	if ctx.ArrayTypeContext() != nil {
		return e.evalArrayLiteralWithExpectedType(node, ctx.ArrayTypeContext(), ctx)
	}

	arrayType := e.getArrayTypeFromAnnotation(node, ctx)

	elementCount := len(node.Elements)
	evaluatedElements := make([]Value, elementCount)
	elementTypes := make([]types.Type, elementCount)

	for idx, elem := range node.Elements {
		var val Value

		// Try nested array evaluation with expected type
		if arrayType != nil {
			if elemLit, ok := elem.(*ast.ArrayLiteralExpression); ok {
				if expectedElemArr, ok := arrayType.ElementType.(*types.ArrayType); ok {
					val = e.evalArrayLiteralWithExpectedType(elemLit, expectedElemArr, ctx)
				}
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

	coercedElements, errVal := e.coerceElementsToType(arrayType, evaluatedElements, elementTypes, node)
	if errVal != nil {
		return errVal
	}

	if arrayType.IsStatic() {
		expectedSize := arrayType.Size()
		if elementCount != expectedSize {
			return e.newError(node, "array literal has %d elements, expected %d", elementCount, expectedSize)
		}
	}

	runtimeElements := make([]runtime.Value, len(coercedElements))
	for i, elem := range coercedElements {
		runtimeElements[i] = elem.(runtime.Value)
	}
	return &runtime.ArrayValue{ArrayType: arrayType, Elements: runtimeElements}
}

// getArrayTypeFromAnnotation retrieves the array type from semantic info.
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

// evalArrayLiteralWithExpectedType evaluates nested array literal with expected type from parent.
func (e *Evaluator) evalArrayLiteralWithExpectedType(node *ast.ArrayLiteralExpression, expected *types.ArrayType, ctx *ExecutionContext) Value {
	if expected == nil {
		return e.evalArrayLiteralDirect(node, ctx)
	}

	if e.semanticInfo == nil {
		return e.evalArrayLiteralWithType(node, expected, ctx)
	}

	// Temporarily set type annotation
	prevType := e.semanticInfo.GetType(node)
	annotation := &ast.TypeAnnotation{Token: node.Token, Name: expected.String()}
	e.semanticInfo.SetType(node, annotation)

	result := e.evalArrayLiteralDirect(node, ctx)

	if prevType != nil {
		e.semanticInfo.SetType(node, prevType)
	} else {
		e.semanticInfo.ClearType(node)
	}

	return result
}

// evalArrayLiteralWithType evaluates array literal with known type (when semanticInfo unavailable).
func (e *Evaluator) evalArrayLiteralWithType(node *ast.ArrayLiteralExpression, arrayType *types.ArrayType, ctx *ExecutionContext) Value {
	elementCount := len(node.Elements)
	evaluatedElements := make([]Value, elementCount)
	elementTypes := make([]types.Type, elementCount)

	for idx, elem := range node.Elements {
		var val Value

		// Try nested array evaluation with expected type
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

	coercedElements, errVal := e.coerceElementsToType(arrayType, evaluatedElements, elementTypes, node)
	if errVal != nil {
		return errVal
	}

	if arrayType.IsStatic() {
		expectedSize := arrayType.Size()
		if elementCount != expectedSize {
			return e.newError(node, "array literal has %d elements, expected %d", elementCount, expectedSize)
		}
	}

	runtimeElements := make([]runtime.Value, len(coercedElements))
	for i, elem := range coercedElements {
		runtimeElements[i] = elem.(runtime.Value)
	}
	return &runtime.ArrayValue{ArrayType: arrayType, Elements: runtimeElements}
}

// inferArrayTypeFromElements infers type from element types (same type → that type, Integer+Float → Float).
func (e *Evaluator) inferArrayTypeFromElements(node *ast.ArrayLiteralExpression, elementTypes []types.Type) *types.ArrayType {
	if len(elementTypes) == 0 {
		return nil
	}

	var inferred types.Type

	for _, elemType := range elementTypes {
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

		return nil
	}

	if inferred == nil {
		return nil
	}

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

		if underlyingElementType.Equals(types.VARIANT) {
			coerced[idx] = runtime.BoxVariant(val)
			continue
		}

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

		if underlyingElementType.Equals(valType) {
			coerced[idx] = val
			continue
		}

		if underlyingElementType.Equals(types.FLOAT) && valType.Equals(types.INTEGER) {
			coerced[idx] = e.castToFloat(val)
			continue
		}

		if valType.TypeKind() == "ARRAY" && underlyingElementType.TypeKind() == "ARRAY" {
			if types.IsCompatible(valType, underlyingElementType) || types.IsCompatible(underlyingElementType, valType) {
				coerced[idx] = val
				continue
			}
		}

		if types.IsCompatible(valType, underlyingElementType) {
			coerced[idx] = val
			continue
		}

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

// evalArrayHelper evaluates built-in array helper methods.
// Returns result or nil if not handled (falls through to adapter).
func (e *Evaluator) evalArrayHelper(spec string, selfValue Value, args []Value, node ast.Node) Value {
	switch spec {
	case "__array_length", "__array_count":
		return e.evalArrayLengthHelper(selfValue, args, node)
	case "__array_high":
		return e.evalArrayHigh(selfValue, args, node)
	case "__array_low":
		return e.evalArrayLow(selfValue, args, node)
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
	case "__array_join":
		return e.evalArrayJoinHelper(selfValue, args, node)
	case "__string_array_join":
		return e.evalStringArrayJoin(selfValue, args, node)
	default:
		return nil
	}
}

// ============================================================================
// Array Properties
// ============================================================================

// evalArrayLengthHelper implements Array.Length and Array.Count.
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

// evalArrayHigh returns highest valid index (declared high bound for static, Length-1 for dynamic).
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

// evalArrayLow returns lowest valid index (declared low bound for static, 0 for dynamic).
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
// Simple Array Methods
// ============================================================================

// evalArrayAdd appends element to dynamic array.
func (e *Evaluator) evalArrayAdd(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 1 {
		return e.newError(node, "Array.Add expects exactly 1 argument")
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.Add requires array receiver")
	}

	if arrVal.ArrayType != nil && !arrVal.ArrayType.IsDynamic() {
		return e.newError(node, "Add() can only be used with dynamic arrays, not static arrays")
	}

	arrVal.Elements = append(arrVal.Elements, args[0])

	return &runtime.NilValue{}
}

// evalArrayPush appends element to dynamic array, copying records to avoid aliasing.
func (e *Evaluator) evalArrayPush(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 1 {
		return e.newError(node, "Array.Push expects exactly 1 argument")
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.Push requires array receiver")
	}

	if arrVal.ArrayType != nil && !arrVal.ArrayType.IsDynamic() {
		return e.newError(node, "Push() can only be used with dynamic arrays, not static arrays")
	}

	valueToAdd := args[0]

	// Copy records (value types) to avoid aliasing in collections
	if copyable, ok := valueToAdd.(interface{ Copy() Value }); ok {
		valueToAdd = copyable.Copy()
	}

	arrVal.Elements = append(arrVal.Elements, valueToAdd)

	return &runtime.NilValue{}
}

// evalArrayPop removes and returns the last element from a dynamic array.
func (e *Evaluator) evalArrayPop(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 0 {
		return e.newError(node, "Array.Pop expects no arguments, got %d", len(args))
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.Pop requires array receiver")
	}

	if arrVal.ArrayType != nil && !arrVal.ArrayType.IsDynamic() {
		return e.newError(node, "Pop() can only be used with dynamic arrays, not static arrays")
	}

	if len(arrVal.Elements) == 0 {
		return e.newError(node, "Pop() called on empty array")
	}

	lastElement := arrVal.Elements[len(arrVal.Elements)-1]
	arrVal.Elements = arrVal.Elements[:len(arrVal.Elements)-1]

	return lastElement
}

// evalArraySwap swaps two elements in the array.
func (e *Evaluator) evalArraySwap(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 2 {
		return e.newError(node, "Array.Swap expects exactly 2 arguments, got %d", len(args))
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.Swap requires array receiver")
	}

	iInt, ok := args[0].(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "Array.Swap first argument must be Integer, got %s", args[0].Type())
	}
	iIdx := int(iInt.Value)

	jInt, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "Array.Swap second argument must be Integer, got %s", args[1].Type())
	}
	jIdx := int(jInt.Value)

	arrayLen := len(arrVal.Elements)
	if iIdx < 0 || iIdx >= arrayLen {
		return e.newError(node, "Array.Swap first index %d out of bounds (0..%d)", iIdx, arrayLen-1)
	}
	if jIdx < 0 || jIdx >= arrayLen {
		return e.newError(node, "Array.Swap second index %d out of bounds (0..%d)", jIdx, arrayLen-1)
	}

	arrVal.Elements[iIdx], arrVal.Elements[jIdx] = arrVal.Elements[jIdx], arrVal.Elements[iIdx]

	return &runtime.NilValue{}
}

// evalArrayDelete removes 1 or more elements from a dynamic array.
func (e *Evaluator) evalArrayDelete(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) < 1 || len(args) > 2 {
		return e.newError(node, "Array.Delete expects 1 or 2 arguments, got %d", len(args))
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.Delete requires array receiver")
	}

	if arrVal.ArrayType != nil && !arrVal.ArrayType.IsDynamic() {
		return e.newError(node, "Delete() can only be used with dynamic arrays, not static arrays")
	}

	indexInt, ok := args[0].(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "Array.Delete index must be Integer, got %s", args[0].Type())
	}
	index := int(indexInt.Value)

	count := 1
	if len(args) == 2 {
		countInt, ok := args[1].(*runtime.IntegerValue)
		if !ok {
			return e.newError(node, "Array.Delete count must be Integer, got %s", args[1].Type())
		}
		count = int(countInt.Value)
	}

	arrayLen := len(arrVal.Elements)
	if index < 0 || index >= arrayLen {
		return e.newError(node, "Array.Delete index %d out of bounds (0..%d)", index, arrayLen-1)
	}
	if count < 0 {
		return e.newError(node, "Array.Delete count must be non-negative, got %d", count)
	}

	endIndex := index + count
	if endIndex > arrayLen {
		endIndex = arrayLen
	}

	arrVal.Elements = append(arrVal.Elements[:index], arrVal.Elements[endIndex:]...)

	return &runtime.NilValue{}
}

// ============================================================================
// Join Methods
// ============================================================================

// evalArrayJoinHelper implements Array.Join(separator) method.
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

// ============================================================================
// Array Manipulation Helpers
// ============================================================================
//
// Standalone functions (not Evaluator methods) to allow reuse from both
// Interpreter and Evaluator without creating circular dependencies.
// ============================================================================

// ValuesEqual compares two Values for equality, handling variant unwrapping,
// nil comparisons, type checking, and recursive record field comparison.
func ValuesEqual(a, b Value) bool {
	if varVal, ok := a.(*runtime.VariantValue); ok {
		a = varVal.Value
	}
	if varVal, ok := b.(*runtime.VariantValue); ok {
		b = varVal.Value
	}

	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	if a.Type() != b.Type() {
		return false
	}

	switch left := a.(type) {
	case *runtime.IntegerValue:
		right, ok := b.(*runtime.IntegerValue)
		if !ok {
			return false
		}
		return left.Value == right.Value

	case *runtime.FloatValue:
		right, ok := b.(*runtime.FloatValue)
		if !ok {
			return false
		}
		return left.Value == right.Value

	case *runtime.StringValue:
		right, ok := b.(*runtime.StringValue)
		if !ok {
			return false
		}
		return left.Value == right.Value

	case *runtime.BooleanValue:
		right, ok := b.(*runtime.BooleanValue)
		if !ok {
			return false
		}
		return left.Value == right.Value

	case *runtime.NilValue:
		return true

	case *runtime.RecordValue:
		right, ok := b.(*runtime.RecordValue)
		if !ok {
			return false
		}
		return recordsEqualInternal(left, right)

	default:
		return a.String() == b.String()
	}
}

// recordsEqualInternal recursively compares two RecordValue instances for equality.
func recordsEqualInternal(left, right *runtime.RecordValue) bool {
	if left.RecordType.Name != right.RecordType.Name {
		return false
	}

	for fieldName := range left.RecordType.Fields {
		leftVal, leftExists := left.Fields[fieldName]
		rightVal, rightExists := right.Fields[fieldName]

		if !leftExists || !rightExists {
			return false
		}

		if !ValuesEqual(leftVal, rightVal) {
			return false
		}
	}

	return true
}

// RecordsEqual checks if two RecordValues are equal by comparing all fields.
func RecordsEqual(left, right Value) bool {
	return ValuesEqual(left, right)
}

// ArrayHelperCopy creates a deep copy of an array.
// For arrays of objects, references are shallow copied (not the objects themselves).
func ArrayHelperCopy(arr *runtime.ArrayValue) Value {
	newArray := &runtime.ArrayValue{
		ArrayType: arr.ArrayType,
		Elements:  make([]runtime.Value, len(arr.Elements)),
	}
	copy(newArray.Elements, arr.Elements)
	return newArray
}

// ArrayHelperIndexOf searches an array for a value starting from startIndex.
// Returns the 0-based index (>= 0) or -1 if not found.
func ArrayHelperIndexOf(arr *runtime.ArrayValue, value Value, startIndex int) Value {
	if startIndex < 0 || startIndex >= len(arr.Elements) {
		return &runtime.IntegerValue{Value: -1}
	}

	for idx := startIndex; idx < len(arr.Elements); idx++ {
		if ValuesEqual(arr.Elements[idx], value) {
			return &runtime.IntegerValue{Value: int64(idx)}
		}
	}

	return &runtime.IntegerValue{Value: -1}
}

// ArrayHelperContains checks if an array contains a specific value.
func ArrayHelperContains(arr *runtime.ArrayValue, value Value) Value {
	// Use IndexOf to check if value exists
	// IndexOf returns >= 0 if found (0-based indexing), -1 if not found
	result := ArrayHelperIndexOf(arr, value, 0)
	intResult, ok := result.(*runtime.IntegerValue)
	if !ok {
		// Should never happen, but handle error case
		return &runtime.BooleanValue{Value: false}
	}

	// Return true if found (index >= 0), false otherwise
	return &runtime.BooleanValue{Value: intResult.Value >= 0}
}

// ArrayHelperReverse reverses an array in place.
func ArrayHelperReverse(arr *runtime.ArrayValue) Value {
	elements := arr.Elements
	n := len(elements)

	// Swap elements from both ends
	for left := 0; left < n/2; left++ {
		right := n - 1 - left
		elements[left], elements[right] = elements[right], elements[left]
	}

	// Return nil (procedure with no return value)
	return &runtime.NilValue{}
}

// ArrayHelperSort sorts an array in place.
// Supports Integer (numeric), Float (numeric), String (lexicographic), and Boolean (false < true).
func ArrayHelperSort(arr *runtime.ArrayValue) Value {
	elements := arr.Elements
	n := len(elements)

	// Empty or single element arrays are already sorted
	if n <= 1 {
		return &runtime.NilValue{}
	}

	// Determine element type from first element
	firstElem := elements[0]

	// Sort based on element type
	switch firstElem.(type) {
	case *runtime.IntegerValue:
		// Numeric sort for integers
		sort.Slice(elements, func(i, j int) bool {
			left, leftOk := elements[i].(*runtime.IntegerValue)
			right, rightOk := elements[j].(*runtime.IntegerValue)
			if !leftOk || !rightOk {
				return false
			}
			return left.Value < right.Value
		})

	case *runtime.FloatValue:
		// Numeric sort for floats
		sort.Slice(elements, func(i, j int) bool {
			left, leftOk := elements[i].(*runtime.FloatValue)
			right, rightOk := elements[j].(*runtime.FloatValue)
			if !leftOk || !rightOk {
				return false
			}
			return left.Value < right.Value
		})

	case *runtime.StringValue:
		// Lexicographic sort for strings
		sort.Slice(elements, func(i, j int) bool {
			left, leftOk := elements[i].(*runtime.StringValue)
			right, rightOk := elements[j].(*runtime.StringValue)
			if !leftOk || !rightOk {
				return false
			}
			return left.Value < right.Value
		})

	case *runtime.BooleanValue:
		// Boolean sort: false < true
		sort.Slice(elements, func(i, j int) bool {
			left, leftOk := elements[i].(*runtime.BooleanValue)
			right, rightOk := elements[j].(*runtime.BooleanValue)
			if !leftOk || !rightOk {
				return false
			}
			// false (false < true) sorts before true
			return !left.Value && right.Value
		})

	default:
		// For other types, we can't sort - just return nil
		return &runtime.NilValue{}
	}

	return &runtime.NilValue{}
}

// ArrayHelperConcatArrays concatenates multiple arrays into a new array.
// The result array type is taken from the first array.
func ArrayHelperConcatArrays(arrays []*runtime.ArrayValue) Value {
	// Collect all elements from all arrays
	var resultElements []runtime.Value
	var firstArrayType *types.ArrayType

	for _, arrayVal := range arrays {
		// Store the type of the first array to use for the result
		if firstArrayType == nil && arrayVal.ArrayType != nil {
			firstArrayType = arrayVal.ArrayType
		}

		// Append all elements from this array
		resultElements = append(resultElements, arrayVal.Elements...)
	}

	// Create and return new array with concatenated elements
	return &runtime.ArrayValue{
		Elements:  resultElements,
		ArrayType: firstArrayType,
	}
}

// ArrayHelperSlice extracts a slice from an array.
// Indices are adjusted relative to the array's low bound (e.g., low bound 1: start=1 extracts first element).
func ArrayHelperSlice(arr *runtime.ArrayValue, startIdx, endIdx int64) Value {
	// Get the low bound of the array
	lowBound := int64(0)
	if arr.ArrayType != nil && arr.ArrayType.LowBound != nil {
		lowBound = int64(*arr.ArrayType.LowBound)
	}

	// Adjust indices to be relative to the array's low bound
	start := int(startIdx - lowBound)
	end := int(endIdx - lowBound)

	// Validate indices
	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = 0
	}
	if end > len(arr.Elements) {
		end = len(arr.Elements)
	}
	if start > end {
		start = end
	}

	// Extract the slice
	resultElements := make([]runtime.Value, end-start)
	copy(resultElements, arr.Elements[start:end])

	// Create and return new array with sliced elements
	return &runtime.ArrayValue{
		Elements:  resultElements,
		ArrayType: arr.ArrayType,
	}
}
