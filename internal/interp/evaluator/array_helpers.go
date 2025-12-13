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

	// Disambiguation: `[...]` can represent a set literal when semantic analysis expects a SET.
	// Some contexts (notably empty literals `[]`) otherwise look like an empty array literal.
	if e.semanticInfo != nil {
		if typeAnnot := e.semanticInfo.GetType(node); typeAnnot != nil && typeAnnot.Name != "" {
			if resolvedType, err := e.ResolveTypeWithContext(typeAnnot.Name, ctx); err == nil {
				if _, ok := types.GetUnderlyingType(resolvedType).(*types.SetType); ok {
					setLit := &ast.SetLiteral{
						Elements:            node.Elements,
						TypedExpressionBase: node.TypedExpressionBase,
					}

					// Preserve the type annotation for set inference (esp. for empty `[]`).
					e.semanticInfo.SetType(setLit, typeAnnot)
					defer e.semanticInfo.ClearType(setLit)

					return e.evalSetLiteralDirect(setLit, ctx)
				}
			}
		}
	}

	// Use context type if available
	if ctx.ArrayTypeContext() != nil {
		return e.evalArrayLiteralWithExpectedType(node, ctx.ArrayTypeContext(), ctx)
	}

	// Get or infer array type
	arrayType := e.getArrayTypeFromAnnotation(node, ctx)
	evaluatedElements, elementTypes, err := e.evaluateArrayElements(node, arrayType, ctx)
	if err != nil {
		return err
	}

	// Infer type if not explicitly provided
	arrayType, err = e.ensureArrayType(arrayType, node, elementTypes)
	if err != nil {
		return err
	}

	// Coerce elements to target type
	coercedElements, err := e.coerceElementsToType(arrayType, evaluatedElements, elementTypes, node)
	if err != nil {
		return err
	}

	// Validate static array size
	if err := e.validateStaticArraySize(arrayType, len(node.Elements), node); err != nil {
		return err
	}

	// Build runtime array
	return e.buildRuntimeArray(arrayType, coercedElements)
}

// evaluateArrayElements evaluates all elements in an array literal.
func (e *Evaluator) evaluateArrayElements(node *ast.ArrayLiteralExpression, arrayType *types.ArrayType, ctx *ExecutionContext) ([]Value, []types.Type, Value) {
	elementCount := len(node.Elements)
	evaluatedElements := make([]Value, elementCount)
	elementTypes := make([]types.Type, elementCount)

	for idx, elem := range node.Elements {
		val := e.evaluateSingleArrayElement(elem, arrayType, ctx)
		if isError(val) {
			return nil, nil, val
		}
		evaluatedElements[idx] = val
		elementTypes[idx] = GetValueType(val)
	}

	return evaluatedElements, elementTypes, nil
}

// evaluateSingleArrayElement evaluates a single array element with optional type context.
func (e *Evaluator) evaluateSingleArrayElement(elem ast.Expression, arrayType *types.ArrayType, ctx *ExecutionContext) Value {
	// Try nested array evaluation with expected element type
	if arrayType != nil {
		if elemLit, ok := elem.(*ast.ArrayLiteralExpression); ok {
			if expectedElemArr, ok := arrayType.ElementType.(*types.ArrayType); ok {
				return e.evalArrayLiteralWithExpectedType(elemLit, expectedElemArr, ctx)
			}
		}
	}

	return e.Eval(elem, ctx)
}

// ensureArrayType returns the array type, inferring it if necessary.
func (e *Evaluator) ensureArrayType(arrayType *types.ArrayType, node *ast.ArrayLiteralExpression, elementTypes []types.Type) (*types.ArrayType, Value) {
	if arrayType != nil {
		return arrayType, nil
	}

	inferred := e.inferArrayTypeFromElements(node, elementTypes)
	if inferred == nil {
		if len(node.Elements) == 0 {
			return nil, e.newError(node, "cannot infer type for empty array literal")
		}
		return nil, e.newError(node, "cannot determine array type for literal")
	}

	return inferred, nil
}

// validateStaticArraySize checks that static array literals have the correct element count.
func (e *Evaluator) validateStaticArraySize(arrayType *types.ArrayType, elementCount int, node ast.Node) Value {
	if !arrayType.IsStatic() {
		return nil
	}

	expectedSize := arrayType.Size()
	if elementCount != expectedSize {
		return e.newError(node, "array literal has %d elements, expected %d", elementCount, expectedSize)
	}

	return nil
}

// buildRuntimeArray creates a runtime ArrayValue from coerced elements.
func (e *Evaluator) buildRuntimeArray(arrayType *types.ArrayType, coercedElements []Value) Value {
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

	// Avoid recursion: evalArrayLiteralDirect consults ctx.ArrayTypeContext(), which in assignment
	// contexts often points back to the same expected type. Evaluate directly with the expected type.
	return e.evalArrayLiteralWithType(node, expected, ctx)
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
	elementType := arrayType.ElementType
	if elementType == nil {
		return nil, e.newError(node, "array literal has no element type information")
	}
	underlyingElementType := types.GetUnderlyingType(elementType)

	coerced := make([]Value, len(values))
	for idx, val := range values {
		coercedVal, err := e.coerceSingleElement(val, valueTypes, idx, underlyingElementType, node)
		if err != nil {
			return nil, err
		}
		coerced[idx] = coercedVal
	}

	return coerced, nil
}

// coerceSingleElement coerces a single array element to the target type.
func (e *Evaluator) coerceSingleElement(val Value, valueTypes []types.Type, idx int, targetType types.Type, node *ast.ArrayLiteralExpression) (Value, Value) {
	// Handle Variant target type (accepts any value)
	if targetType.Equals(types.VARIANT) {
		return runtime.BoxVariant(val), nil
	}

	// Handle nil values
	if val != nil && val.Type() == "NIL" {
		return e.handleNilElement(val, idx, targetType, node)
	}

	// Get source type
	var valType types.Type
	if idx < len(valueTypes) && valueTypes[idx] != nil {
		valType = types.GetUnderlyingType(valueTypes[idx])
	}
	if valType == nil {
		return nil, e.elementError(node, idx, "cannot determine type for array element %d", idx+1)
	}

	// Check for exact type match
	if targetType.Equals(valType) {
		return val, nil
	}

	// Handle numeric promotion (Integer → Float)
	if targetType.Equals(types.FLOAT) && valType.Equals(types.INTEGER) {
		return e.castToFloat(val), nil
	}

	// Handle array type compatibility
	if valType.TypeKind() == "ARRAY" && targetType.TypeKind() == "ARRAY" {
		if types.IsCompatible(valType, targetType) || types.IsCompatible(targetType, valType) {
			return val, nil
		}
	}

	// Handle general type compatibility
	if types.IsCompatible(valType, targetType) {
		return val, nil
	}

	// Type mismatch error
	return nil, e.elementError(node, idx, "array element %d has incompatible type (got %s, expected %s)",
		idx+1, val.Type(), targetType.String())
}

// handleNilElement validates and returns nil values for reference types.
func (e *Evaluator) handleNilElement(val Value, idx int, targetType types.Type, node *ast.ArrayLiteralExpression) (Value, Value) {
	switch targetType.TypeKind() {
	case "CLASS", "INTERFACE", "ARRAY":
		return val, nil
	default:
		return nil, e.elementError(node, idx, "cannot assign nil to %s", targetType.String())
	}
}

// elementError creates an error with proper node context for an array element.
func (e *Evaluator) elementError(node *ast.ArrayLiteralExpression, idx int, format string, args ...interface{}) Value {
	elemNode := node
	if idx < len(node.Elements) {
		elemNode = &ast.ArrayLiteralExpression{Elements: []ast.Expression{node.Elements[idx]}}
	}
	return e.newError(elemNode, format, args...)
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
	a = unwrapVariant(a)
	b = unwrapVariant(b)

	// Handle nil cases
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Type must match
	if a.Type() != b.Type() {
		return false
	}

	return compareValuesByType(a, b)
}

// compareValuesByType compares two non-nil values of the same type.
func compareValuesByType(a, b Value) bool {
	switch left := a.(type) {
	case *runtime.IntegerValue:
		return compareInteger(left, b)
	case *runtime.FloatValue:
		return compareFloat(left, b)
	case *runtime.StringValue:
		return compareString(left, b)
	case *runtime.BooleanValue:
		return compareBoolean(left, b)
	case *runtime.NilValue:
		return true
	case *runtime.RecordValue:
		return compareRecord(left, b)
	default:
		return a.String() == b.String()
	}
}

// compareInteger compares two integer values.
func compareInteger(left *runtime.IntegerValue, b Value) bool {
	right, ok := b.(*runtime.IntegerValue)
	return ok && left.Value == right.Value
}

// compareFloat compares two float values.
func compareFloat(left *runtime.FloatValue, b Value) bool {
	right, ok := b.(*runtime.FloatValue)
	return ok && left.Value == right.Value
}

// compareString compares two string values.
func compareString(left *runtime.StringValue, b Value) bool {
	right, ok := b.(*runtime.StringValue)
	return ok && left.Value == right.Value
}

// compareBoolean compares two boolean values.
func compareBoolean(left *runtime.BooleanValue, b Value) bool {
	right, ok := b.(*runtime.BooleanValue)
	return ok && left.Value == right.Value
}

// compareRecord compares two record values.
func compareRecord(left *runtime.RecordValue, b Value) bool {
	right, ok := b.(*runtime.RecordValue)
	return ok && recordsEqualInternal(left, right)
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
