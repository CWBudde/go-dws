package evaluator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
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

	for idx, dimExpr := range dimensions {
		dimValue := e.Eval(dimExpr, ctx)
		if isError(dimValue) {
			return nil, dimValue
		}

		if dimValue.Type() != "INTEGER" {
			return nil, e.newError(node, "array dimension must be an integer, got %s", dimValue.Type())
		}

		dimSize, err := e.extractIntegerValue(dimValue)
		if err != nil {
			return nil, e.newError(node, "dimension %d: %v", idx, err)
		}

		// Zero-size dynamic arrays are legal in DWScript (new Integer[0]).
		if dimSize < 0 {
			return nil, e.newError(node, "array dimension must be non-negative, got %d", dimSize)
		}

		dimSizes[idx] = dimSize
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
	if e.SemanticInfo() != nil {
		if typeAnnot := e.SemanticInfo().GetType(node); typeAnnot != nil && typeAnnot.Name != "" {
			isSetAnnotation := false
			if resolvedType, err := e.ResolveTypeWithContext(typeAnnot.Name, ctx); err == nil {
				_, isSetAnnotation = types.GetUnderlyingType(resolvedType).(*types.SetType)
			} else if e.parseInlineSetType(typeAnnot.Name) != nil {
				// Inline "set of X" annotations may not resolve as named types.
				isSetAnnotation = true
			}
			if isSetAnnotation {
				setLit := &ast.SetLiteral{
					Elements:            node.Elements,
					TypedExpressionBase: node.TypedExpressionBase,
				}

				// Preserve the type annotation for set inference (esp. for empty `[]`).
				e.SemanticInfo().SetType(setLit, typeAnnot)
				defer e.SemanticInfo().ClearType(setLit)

				return e.evalSetLiteralDirect(setLit, ctx)
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

	// Validate static array size (against expanded element count)
	if err := e.validateStaticArraySize(arrayType, len(coercedElements), node); err != nil {
		return err
	}

	// Build runtime array
	return e.buildRuntimeArray(arrayType, coercedElements)
}

// evaluateArrayElements evaluates all elements in an array literal, expanding
// ordinal range elements ([1..3] → 1,2,3; [3..1] → 3,2,1).
func (e *Evaluator) evaluateArrayElements(node *ast.ArrayLiteralExpression, arrayType *types.ArrayType, ctx *ExecutionContext) ([]Value, []types.Type, Value) {
	evaluatedElements := make([]Value, 0, len(node.Elements))
	elementTypes := make([]types.Type, 0, len(node.Elements))

	for _, elem := range node.Elements {
		if rangeExpr, isRange := elem.(*ast.RangeExpression); isRange {
			expanded, errVal := e.expandArrayRangeElement(rangeExpr, ctx)
			if errVal != nil {
				return nil, nil, errVal
			}
			for _, val := range expanded {
				evaluatedElements = append(evaluatedElements, val)
				elementTypes = append(elementTypes, GetValueType(val))
			}
			continue
		}

		val := e.evaluateSingleArrayElement(elem, arrayType, ctx)
		if isError(val) {
			return nil, nil, val
		}
		evaluatedElements = append(evaluatedElements, val)
		elementTypes = append(elementTypes, GetValueType(val))
	}

	return evaluatedElements, elementTypes, nil
}

// expandArrayRangeElement evaluates a range element of an array constructor
// into its sequence of values. Descending ranges expand in descending order.
func (e *Evaluator) expandArrayRangeElement(rangeExpr *ast.RangeExpression, ctx *ExecutionContext) ([]Value, Value) {
	startVal := e.Eval(rangeExpr.Start, ctx)
	if isError(startVal) {
		return nil, startVal
	}
	endVal := e.Eval(rangeExpr.RangeEnd, ctx)
	if isError(endVal) {
		return nil, endVal
	}

	if startEnum, ok := unwrapVariant(startVal).(*runtime.EnumValue); ok {
		enumType, err := e.lookupEnumType(startEnum.TypeName)
		if err != nil {
			return nil, e.newError(rangeExpr, "%s", err.Error())
		}
		endEnum, ok := unwrapVariant(endVal).(*runtime.EnumValue)
		if !ok {
			return nil, e.newError(rangeExpr, "range bounds must have the same type")
		}
		startIdx, err := runtime.EnumValueIndex(startEnum, enumType)
		if err != nil {
			return nil, e.newError(rangeExpr, "%s", err.Error())
		}
		endIdx, err := runtime.EnumValueIndex(endEnum, enumType)
		if err != nil {
			return nil, e.newError(rangeExpr, "%s", err.Error())
		}
		values := make([]Value, 0, absInt(endIdx-startIdx)+1)
		for idx := startIdx; ; idx += signInt(endIdx - startIdx) {
			val, err := runtime.EnumValueAtIndex(startEnum.TypeName, enumType, idx)
			if err != nil {
				return nil, e.newError(rangeExpr, "%s", err.Error())
			}
			values = append(values, val)
			if idx == endIdx {
				break
			}
		}
		return values, nil
	}

	startOrd, err := runtime.GetOrdinalValue(unwrapVariant(startVal))
	if err != nil {
		return nil, e.newError(rangeExpr, "range bounds must be ordinal: %s", err.Error())
	}
	endOrd, err := runtime.GetOrdinalValue(unwrapVariant(endVal))
	if err != nil {
		return nil, e.newError(rangeExpr, "range bounds must be ordinal: %s", err.Error())
	}
	values := make([]Value, 0, absInt(endOrd-startOrd)+1)
	for ord := startOrd; ; ord += signInt(endOrd - startOrd) {
		values = append(values, &runtime.IntegerValue{Value: int64(ord)})
		if ord == endOrd {
			break
		}
	}
	return values, nil
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// signInt returns the iteration step direction for a range (+1, -1, or +1 for
// single-element ranges so the loop terminates).
func signInt(n int) int {
	if n < 0 {
		return -1
	}
	return 1
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
		if recordLit, ok := elem.(*ast.RecordLiteralExpression); ok && recordLit.TypeName == nil {
			if recordType, ok := types.GetUnderlyingType(arrayType.ElementType).(*types.RecordType); ok {
				prev := ctx.RecordTypeContext()
				ctx.SetRecordTypeContext(recordType.Name)
				val := e.Eval(recordLit, ctx)
				ctx.SetRecordTypeContext(prev)
				return val
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
	copy(runtimeElements, coercedElements)
	return &runtime.ArrayValue{ArrayType: arrayType, Elements: runtimeElements}
}

// getArrayTypeFromAnnotation retrieves the array type from semantic info.
func (e *Evaluator) getArrayTypeFromAnnotation(node *ast.ArrayLiteralExpression, ctx *ExecutionContext) *types.ArrayType {
	if e.SemanticInfo() == nil {
		return nil
	}

	typeAnnot := e.SemanticInfo().GetType(node)
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
	evaluatedElements := make([]Value, 0, len(node.Elements))
	elementTypes := make([]types.Type, 0, len(node.Elements))

	for _, elem := range node.Elements {
		if rangeExpr, isRange := elem.(*ast.RangeExpression); isRange {
			expanded, errVal := e.expandArrayRangeElement(rangeExpr, ctx)
			if errVal != nil {
				return errVal
			}
			for _, val := range expanded {
				evaluatedElements = append(evaluatedElements, val)
				elementTypes = append(elementTypes, GetValueType(val))
			}
			continue
		}

		var val Value

		// Try nested array evaluation with expected type
		if elemLit, ok := elem.(*ast.ArrayLiteralExpression); ok {
			if expectedElemArr, ok := arrayType.ElementType.(*types.ArrayType); ok {
				val = e.evalArrayLiteralWithType(elemLit, expectedElemArr, ctx)
			}
		}
		if val == nil {
			if recordLit, ok := elem.(*ast.RecordLiteralExpression); ok && recordLit.TypeName == nil {
				if recordType, ok := types.GetUnderlyingType(arrayType.ElementType).(*types.RecordType); ok {
					prev := ctx.RecordTypeContext()
					ctx.SetRecordTypeContext(recordType.Name)
					val = e.Eval(recordLit, ctx)
					ctx.SetRecordTypeContext(prev)
				}
			}
		}

		if val == nil {
			val = e.Eval(elem, ctx)
		}

		if isError(val) {
			return val
		}
		evaluatedElements = append(evaluatedElements, val)
		elementTypes = append(elementTypes, GetValueType(val))
	}

	coercedElements, errVal := e.coerceElementsToType(arrayType, evaluatedElements, elementTypes, node)
	if errVal != nil {
		return errVal
	}

	if arrayType.IsStatic() {
		expectedSize := arrayType.Size()
		if len(evaluatedElements) != expectedSize {
			return e.newError(node, "array literal has %d elements, expected %d", len(evaluatedElements), expectedSize)
		}
	}

	runtimeElements := make([]runtime.Value, len(coercedElements))
	copy(runtimeElements, coercedElements)
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
		// Enum values carry only a type name; resolve it via the type system.
		if enumVal, ok := unwrapVariant(val).(*runtime.EnumValue); ok {
			if enumType, err := e.lookupEnumType(enumVal.TypeName); err == nil {
				valType = enumType
			}
		}
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
	case "__array_indexof":
		return e.evalArrayIndexOf(selfValue, args, node)
	case "__array_setlength":
		return e.evalArraySetLength(selfValue, args, node)
	case "__array_join":
		return e.evalArrayJoinHelper(selfValue, args, node)
	case "__string_array_join":
		return e.evalStringArrayJoin(selfValue, args, node)
	case "__array_map":
		return e.evalArrayMap(selfValue, args, node)
	case "__array_move":
		return e.evalArrayMove(selfValue, args, node)
	case "__array_reverse":
		return e.evalArrayReverse(selfValue, args, node)
	case "__array_sort":
		return e.evalArraySortMethod(selfValue, args, node)
	case "__array_insert":
		return e.evalArrayInsert(selfValue, args, node)
	case "__array_copy":
		return e.evalArrayCopyMethod(selfValue, args, node)
	case "__array_remove":
		return e.evalArrayRemove(selfValue, args, node)
	case "__array_contains":
		return e.evalArrayContains(selfValue, args, node)
	case "__array_foreach":
		return e.evalArrayForEach(selfValue, args, node)
	case "__array_filter":
		return e.evalArrayFilter(selfValue, args, node)
	case "__array_clear":
		return e.evalArrayClear(selfValue, args, node)
	case "__array_peek":
		return e.evalArrayPeek(selfValue, args, node)
	default:
		return nil
	}
}

// arrayMethodNamePos returns the source position of the method/member name in a
// helper call, so bound-check exceptions point at the method name (matching
// DWScript's diagnostics) rather than the receiver expression.
func arrayMethodNamePos(node ast.Node) token.Position {
	switch n := node.(type) {
	case *ast.MethodCallExpression:
		if n.Method != nil {
			return n.Method.Token.Pos
		}
	case *ast.MemberAccessExpression:
		if n.Member != nil {
			return n.Member.Token.Pos
		}
	}
	if node != nil {
		return node.Pos()
	}
	return token.Position{}
}

// raiseArrayBoundExceeded sets a catchable "bound exceeded" exception, matching
// DWScript's message format, and returns nil so the surrounding try/except can
// intercept it.
func (e *Evaluator) raiseArrayBoundExceeded(node ast.Node, index int, upper bool) Value {
	ctx := e.currentContext
	pos := arrayMethodNamePos(node)
	boundWord := "Lower"
	if upper {
		boundWord = "Upper"
	}
	message := fmt.Sprintf("%s bound exceeded! Index %d", boundWord, index)
	if ctx != nil {
		if routine := currentRoutineName(ctx); routine != "" {
			message += " in " + routine
		}
	}
	message = fmt.Sprintf("%s [line: %d, column: %d]", message, pos.Line, pos.Column)
	if ctx != nil {
		exc := e.createException("Exception", message, &pos, ctx)
		ctx.SetException(exc)
	}
	return e.nilValue()
}

// indexBracketPos returns the position of the closing bracket of an index
// expression (one column before its end), matching DWScript's bounds
// diagnostics for a[i] reads and writes.
func indexBracketPos(node ast.Node) token.Position {
	if node == nil {
		return token.Position{}
	}
	pos := node.End()
	if pos.Column > 1 {
		pos.Column--
	}
	return pos
}

// raiseIndexBoundExceeded sets a catchable "bound exceeded" exception for an
// out-of-bounds a[i] access, pointing at the closing bracket of the index
// expression, and returns nil so the surrounding try/except can intercept it.
func (e *Evaluator) raiseIndexBoundExceeded(node ast.Node, index int, upper bool) Value {
	return e.raiseIndexBoundExceededAt(indexBracketPos(node), index, upper)
}

// raiseIndexBoundExceededAt is raiseIndexBoundExceeded with an explicit position.
func (e *Evaluator) raiseIndexBoundExceededAt(pos token.Position, index int, upper bool) Value {
	ctx := e.currentContext
	boundWord := "Lower"
	if upper {
		boundWord = "Upper"
	}
	message := fmt.Sprintf("%s bound exceeded! Index %d", boundWord, index)
	if ctx != nil {
		if routine := currentRoutineName(ctx); routine != "" {
			message += " in " + routine
		}
	}
	message = fmt.Sprintf("%s [line: %d, column: %d]", message, pos.Line, pos.Column)
	if ctx != nil {
		exc := e.createException("Exception", message, &pos, ctx)
		ctx.SetException(exc)
	}
	return e.nilValue()
}

// raisePositiveCountExpected sets a catchable exception for a negative count
// argument, matching DWScript's message format.
func (e *Evaluator) raisePositiveCountExpected(node ast.Node, count int) Value {
	ctx := e.currentContext
	pos := arrayMethodNamePos(node)
	message := fmt.Sprintf("Positive count expected (got %d) [line: %d, column: %d]", count, pos.Line, pos.Column)
	if ctx != nil {
		exc := e.createException("Exception", message, &pos, ctx)
		ctx.SetException(exc)
	}
	return e.nilValue()
}

// evalArrayMove relocates the element at index `from` to index `to`, shifting the
// intervening elements. Both indices must be within bounds.
func (e *Evaluator) evalArrayMove(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 2 {
		return e.newError(node, "Array.Move expects exactly 2 arguments, got %d", len(args))
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.Move requires array receiver")
	}

	fromInt, ok := args[0].(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "Array.Move first argument must be Integer, got %s", args[0].Type())
	}
	toInt, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "Array.Move second argument must be Integer, got %s", args[1].Type())
	}

	from := int(fromInt.Value)
	to := int(toInt.Value)
	arrayLen := len(arrVal.Elements)

	if from < 0 {
		return e.raiseArrayBoundExceeded(node, from, false)
	}
	if from >= arrayLen {
		return e.raiseArrayBoundExceeded(node, from, true)
	}
	if to < 0 {
		return e.raiseArrayBoundExceeded(node, to, false)
	}
	if to >= arrayLen {
		return e.raiseArrayBoundExceeded(node, to, true)
	}

	if from != to {
		elem := arrVal.Elements[from]
		arrVal.Elements = append(arrVal.Elements[:from], arrVal.Elements[from+1:]...)
		// Re-insert at target index (indices stay valid because from and to are
		// both < arrayLen and we removed exactly one element).
		arrVal.Elements = append(arrVal.Elements, nil)
		copy(arrVal.Elements[to+1:], arrVal.Elements[to:])
		arrVal.Elements[to] = elem
	}

	return e.nilValue()
}

// evalArrayReverse reverses the array in place and returns it (enabling chaining).
func (e *Evaluator) evalArrayReverse(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 0 {
		return e.newError(node, "Array.Reverse expects no arguments, got %d", len(args))
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.Reverse requires array receiver")
	}

	runtime.ArrayHelperReverse(arrVal)
	return arrVal
}

// evalArraySortMethod sorts the array in place, either by natural order (no
// argument) or by a supplied comparator function, and returns the array.
func (e *Evaluator) evalArraySortMethod(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) > 1 {
		return e.newError(node, "Array.Sort expects at most 1 argument, got %d", len(args))
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.Sort requires array receiver")
	}

	if len(args) == 0 {
		runtime.ArrayHelperSort(arrVal)
		return arrVal
	}

	funcPtr, ok := args[0].(*runtime.FunctionPointerValue)
	if !ok {
		return e.newError(node, "Array.Sort expects function pointer as argument, got %s", args[0].Type())
	}

	var sortErr Value
	sort.SliceStable(arrVal.Elements, func(i, j int) bool {
		if sortErr != nil {
			return false
		}
		result := e.EvalFunctionPointer(funcPtr, []Value{arrVal.Elements[i], arrVal.Elements[j]})
		if isError(result) {
			sortErr = result
			return false
		}
		cmp, ok := result.(*runtime.IntegerValue)
		if !ok {
			sortErr = e.newError(node, "Array.Sort comparator must return Integer, got %s", result.Type())
			return false
		}
		return cmp.Value < 0
	})
	if sortErr != nil {
		return sortErr
	}

	return arrVal
}

// evalArrayInsert inserts a value at the given index, shifting later elements.
// The index may range over 0..Length (appending when equal to Length).
func (e *Evaluator) evalArrayInsert(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 2 {
		return e.newError(node, "Array.Insert expects exactly 2 arguments, got %d", len(args))
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.Insert requires array receiver")
	}

	if arrVal.ArrayType != nil && !arrVal.ArrayType.IsDynamic() {
		return e.newError(node, "Insert() can only be used with dynamic arrays, not static arrays")
	}

	indexInt, ok := args[0].(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "Array.Insert index must be Integer, got %s", args[0].Type())
	}
	index := int(indexInt.Value)
	arrayLen := len(arrVal.Elements)

	if index < 0 {
		return e.raiseArrayBoundExceeded(node, index, false)
	}
	if index > arrayLen {
		return e.raiseArrayBoundExceeded(node, index, true)
	}

	value := runtime.CopyValue(args[1])
	arrVal.Elements = append(arrVal.Elements, nil)
	copy(arrVal.Elements[index+1:], arrVal.Elements[index:])
	arrVal.Elements[index] = value

	return e.nilValue()
}

// evalArrayCopyMethod returns a new dynamic array with a copied slice of the
// receiver: Copy, Copy(startIndex), or Copy(startIndex, count). With no
// arguments the whole array is duplicated. An out-of-range start index raises a
// catchable bound exception (as DWScript does); a count larger than the number
// of available elements is clamped, but a negative count is an error.
func (e *Evaluator) evalArrayCopyMethod(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) > 2 {
		return e.newError(node, "Array.Copy expects at most 2 arguments, got %d", len(args))
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.Copy requires array receiver")
	}

	arrayLen := len(arrVal.Elements)

	start := 0
	if len(args) >= 1 {
		startInt, ok := args[0].(*runtime.IntegerValue)
		if !ok {
			return e.newError(node, "Array.Copy startIndex must be Integer, got %s", args[0].Type())
		}
		start = int(startInt.Value)
	}

	count := arrayLen - start
	if len(args) == 2 {
		countInt, ok := args[1].(*runtime.IntegerValue)
		if !ok {
			return e.newError(node, "Array.Copy count must be Integer, got %s", args[1].Type())
		}
		count = int(countInt.Value)
	}

	// Validate the start index against the array bounds. A whole-array copy of an
	// empty array (no explicit start) is permitted and yields an empty array.
	if len(args) >= 1 || arrayLen > 0 {
		if start < 0 {
			return e.raiseArrayBoundExceeded(node, start, false)
		}
		if start >= arrayLen {
			return e.raiseArrayBoundExceeded(node, start, true)
		}
	}

	if count < 0 {
		return e.raisePositiveCountExpected(node, count)
	}

	end := start + count
	if end > arrayLen {
		end = arrayLen
	}

	var elementType types.Type = types.VARIANT
	if arrVal.ArrayType != nil && arrVal.ArrayType.ElementType != nil {
		elementType = arrVal.ArrayType.ElementType
	}
	newArray := &runtime.ArrayValue{
		ArrayType: types.NewDynamicArrayType(elementType),
		Elements:  make([]Value, 0, end-start),
	}
	for idx := start; idx < end; idx++ {
		newArray.Elements = append(newArray.Elements, runtime.CopyValue(arrVal.Elements[idx]))
	}

	return newArray
}

// evalArrayRemove removes the first element equal to the given value at or after
// the optional startIndex (default 0) and returns that element's index, or -1 if
// no matching element is found.
func (e *Evaluator) evalArrayRemove(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) < 1 || len(args) > 2 {
		return e.newError(node, "Array.Remove expects 1 or 2 arguments, got %d", len(args))
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.Remove requires array receiver")
	}

	if arrVal.ArrayType != nil && !arrVal.ArrayType.IsDynamic() {
		return e.newError(node, "Remove() can only be used with dynamic arrays, not static arrays")
	}

	startIndex := 0
	if len(args) == 2 {
		startInt, ok := args[1].(*runtime.IntegerValue)
		if !ok {
			return e.newError(node, "Array.Remove startIndex must be Integer, got %s", args[1].Type())
		}
		startIndex = int(startInt.Value)
	}
	if startIndex < 0 {
		startIndex = 0
	}

	target := args[0]
	for idx := startIndex; idx < len(arrVal.Elements); idx++ {
		if runtime.ValuesEqual(arrVal.Elements[idx], target) {
			copy(arrVal.Elements[idx:], arrVal.Elements[idx+1:])
			last := len(arrVal.Elements) - 1
			arrVal.Elements[last] = nil // release the reference so it can be GC'd
			arrVal.Elements = arrVal.Elements[:last]
			return &runtime.IntegerValue{Value: int64(idx)}
		}
	}

	return &runtime.IntegerValue{Value: -1}
}

// evalArrayContains reports whether the array holds an element equal to value.
func (e *Evaluator) evalArrayContains(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 1 {
		return e.newError(node, "Array.Contains expects exactly 1 argument, got %d", len(args))
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.Contains requires array receiver")
	}

	for _, elem := range arrVal.Elements {
		if runtime.ValuesEqual(elem, args[0]) {
			return &runtime.BooleanValue{Value: true}
		}
	}
	return &runtime.BooleanValue{Value: false}
}

// evalArrayForEach invokes the supplied procedure for each element.
func (e *Evaluator) evalArrayForEach(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 1 {
		return e.newError(node, "Array.ForEach expects exactly 1 argument, got %d", len(args))
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.ForEach requires array receiver")
	}

	funcPtr, ok := args[0].(*runtime.FunctionPointerValue)
	if !ok {
		return e.newError(node, "Array.ForEach expects function pointer as argument, got %s", args[0].Type())
	}

	for _, elem := range arrVal.Elements {
		result := e.EvalFunctionPointer(funcPtr, []Value{elem})
		if isError(result) {
			return result
		}
	}

	return e.nilValue()
}

// evalArrayFilter returns a new dynamic array containing the elements for which
// the supplied predicate returns True.
func (e *Evaluator) evalArrayFilter(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 1 {
		return e.newError(node, "Array.Filter expects exactly 1 argument, got %d", len(args))
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.Filter requires array receiver")
	}

	funcPtr, ok := args[0].(*runtime.FunctionPointerValue)
	if !ok {
		return e.newError(node, "Array.Filter expects function pointer as argument, got %s", args[0].Type())
	}

	var elementType types.Type = types.VARIANT
	if arrVal.ArrayType != nil && arrVal.ArrayType.ElementType != nil {
		elementType = arrVal.ArrayType.ElementType
	}
	newArray := &runtime.ArrayValue{
		ArrayType: types.NewDynamicArrayType(elementType),
		Elements:  make([]Value, 0, len(arrVal.Elements)),
	}
	for _, elem := range arrVal.Elements {
		result := e.EvalFunctionPointer(funcPtr, []Value{elem})
		if isError(result) {
			return result
		}
		keep, ok := result.(*runtime.BooleanValue)
		if !ok {
			return e.newError(node, "Array.Filter predicate must return Boolean, got %s", result.Type())
		}
		if keep.Value {
			newArray.Elements = append(newArray.Elements, runtime.CopyValue(elem))
		}
	}

	return newArray
}

// evalArrayClear empties a dynamic array.
func (e *Evaluator) evalArrayClear(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 0 {
		return e.newError(node, "Array.Clear expects no arguments, got %d", len(args))
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.Clear requires array receiver")
	}

	if arrVal.ArrayType != nil && !arrVal.ArrayType.IsDynamic() {
		return e.newError(node, "Clear() can only be used with dynamic arrays, not static arrays")
	}

	// Release element references so they can be GC'd before reslicing to empty.
	for idx := range arrVal.Elements {
		arrVal.Elements[idx] = nil
	}
	arrVal.Elements = arrVal.Elements[:0]
	return e.nilValue()
}

// evalArrayPeek returns the last element without removing it.
func (e *Evaluator) evalArrayPeek(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 0 {
		return e.newError(node, "Array.Peek expects no arguments, got %d", len(args))
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.Peek requires array receiver")
	}

	if len(arrVal.Elements) == 0 {
		// DWScript reports Peek on an empty array as an out-of-bounds access.
		return e.raiseArrayBoundExceeded(node, 0, true)
	}

	return arrVal.Elements[len(arrVal.Elements)-1]
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
	if len(args) == 0 {
		return e.newError(node, "Array.Add expects at least 1 argument")
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.Add requires array receiver")
	}

	if arrVal.ArrayType != nil && !arrVal.ArrayType.IsDynamic() {
		return e.newError(node, "Add() can only be used with dynamic arrays, not static arrays")
	}

	e.appendArrayArgs(arrVal, args)

	return &runtime.NilValue{}
}

// appendArrayArgs appends each argument to the array. An argument that is itself
// an array of the receiver's element type is flattened (DWScript's Add/Push
// accept either an element or an array of elements).
func (e *Evaluator) appendArrayArgs(arrVal *runtime.ArrayValue, args []Value) {
	elemKind := ""
	if arrVal.ArrayType != nil && arrVal.ArrayType.ElementType != nil {
		elemKind = types.GetUnderlyingType(arrVal.ArrayType.ElementType).TypeKind()
	}
	appendOne := func(v Value) bool {
		// Coerce Variant values to a basic element type (DWScript variant
		// casts); a failed cast raises a catchable exception.
		switch elemKind {
		case "INTEGER", "FLOAT", "STRING", "BOOLEAN":
			unwrapped := unwrapVariant(v)
			converted, errVal := e.coerceValueToKind(unwrapped, elemKind, nil, e.currentContext)
			if errVal != nil {
				return false
			}
			if converted != nil {
				v = converted
			} else {
				v = unwrapped
			}
		}
		arrVal.Elements = append(arrVal.Elements, runtime.CopyValue(v))
		return true
	}
	for _, arg := range args {
		if arrArg, ok := arg.(*runtime.ArrayValue); ok && e.shouldFlattenArrayArg(arrVal, arrArg) {
			for _, el := range arrArg.Elements {
				if !appendOne(el) {
					return
				}
			}
			continue
		}
		if !appendOne(arg) {
			return
		}
	}
}

// shouldFlattenArrayArg reports whether an array argument to Add/Push should be
// flattened into individual elements rather than appended as a single element.
// It is added as a single element only when the receiver's element type is
// itself an array compatible with the argument.
func (e *Evaluator) shouldFlattenArrayArg(receiver, arg *runtime.ArrayValue) bool {
	if receiver.ArrayType == nil || receiver.ArrayType.ElementType == nil {
		return true
	}
	elemType := receiver.ArrayType.ElementType
	// Only an element type that is itself an array can absorb an array
	// argument as a single element. In particular, `array of Variant`
	// flattens array arguments (DWScript appends the elements, boxed).
	if _, elemIsArray := types.GetUnderlyingType(elemType).(*types.ArrayType); !elemIsArray {
		return true
	}
	if arg.ArrayType != nil && (types.IsCompatible(arg.ArrayType, elemType) || types.IsCompatible(elemType, arg.ArrayType)) {
		return false
	}
	return true
}

// evalArrayPush appends element to dynamic array, copying records to avoid aliasing.
func (e *Evaluator) evalArrayPush(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) == 0 {
		return e.newError(node, "Array.Push expects at least 1 argument")
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.Push requires array receiver")
	}

	if arrVal.ArrayType != nil && !arrVal.ArrayType.IsDynamic() {
		return e.newError(node, "Push() can only be used with dynamic arrays, not static arrays")
	}

	e.appendArrayArgs(arrVal, args)

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
		// DWScript reports Pop on an empty array as an out-of-bounds access.
		return e.raiseArrayBoundExceeded(node, 0, true)
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

	return arrVal
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

func (e *Evaluator) evalArrayIndexOf(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) < 1 || len(args) > 2 {
		return e.newError(node, "Array.IndexOf expects 1 or 2 arguments, got %d", len(args))
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.IndexOf requires array receiver")
	}

	startIndex := 0
	if len(args) == 2 {
		startIndexVal, ok := args[1].(*runtime.IntegerValue)
		if !ok {
			return e.newError(node, "Array.IndexOf startIndex must be Integer, got %s", args[1].Type())
		}
		startIndex = int(startIndexVal.Value)
	}

	return runtime.ArrayHelperIndexOf(arrVal, args[0], startIndex)
}

func (e *Evaluator) evalArraySetLength(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 1 {
		return e.newError(node, "Array.SetLength expects exactly 1 argument")
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.SetLength requires array receiver")
	}
	if arrVal.ArrayType != nil && !arrVal.ArrayType.IsDynamic() {
		return e.newError(node, "SetLength() can only be used with dynamic arrays, not static arrays")
	}

	newLengthVal, ok := args[0].(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "Array.SetLength expects integer argument, got %s", args[0].Type())
	}

	newLength := int(newLengthVal.Value)
	if newLength < 0 {
		return e.newError(node, "Array.SetLength expects non-negative length, got %d", newLength)
	}

	currentLength := len(arrVal.Elements)
	switch {
	case newLength == currentLength:
		return &runtime.NilValue{}
	case newLength < currentLength:
		arrVal.Elements = arrVal.Elements[:newLength]
		return &runtime.NilValue{}
	}

	for idx := currentLength; idx < newLength; idx++ {
		if arrVal.ArrayType == nil || arrVal.ArrayType.ElementType == nil {
			arrVal.Elements = append(arrVal.Elements, &runtime.NilValue{})
			continue
		}
		arrVal.Elements = append(arrVal.Elements, e.GetDefaultValue(arrVal.ArrayType.ElementType))
	}

	return &runtime.NilValue{}
}

func (e *Evaluator) evalArrayMap(selfValue Value, args []Value, node ast.Node) Value {
	if len(args) != 1 {
		return e.newError(node, "Array.Map expects exactly 1 argument")
	}

	arrVal, ok := selfValue.(*runtime.ArrayValue)
	if !ok {
		return e.newError(node, "Array.Map requires array receiver")
	}

	funcPtr, ok := args[0].(*runtime.FunctionPointerValue)
	if !ok {
		return e.newError(node, "Array.Map expects function pointer as argument, got %s", args[0].Type())
	}

	resultElements := make([]Value, len(arrVal.Elements))
	for idx, element := range arrVal.Elements {
		result := e.EvalFunctionPointer(funcPtr, []Value{element})
		if isError(result) {
			return result
		}
		resultElements[idx] = runtime.CopyValue(result)
	}

	return &runtime.ArrayValue{
		Elements:  resultElements,
		ArrayType: arrVal.ArrayType,
	}
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
	case *runtime.ObjectInstance:
		// Object references compare by identity, not by content
		right, ok := b.(*runtime.ObjectInstance)
		return ok && left == right
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
// Value-semantic elements are copied so the result does not alias the source slice.
func ArrayHelperCopy(arr *runtime.ArrayValue) Value {
	newArray := &runtime.ArrayValue{
		ArrayType: arr.ArrayType,
		Elements:  make([]runtime.Value, len(arr.Elements)),
	}
	for idx, elem := range arr.Elements {
		newArray.Elements[idx] = runtime.CopyValue(elem)
	}
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
// Value-semantic elements are copied before storage to avoid aliasing.
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

		// Append copies of all elements from this array
		for _, elem := range arrayVal.Elements {
			resultElements = append(resultElements, runtime.CopyValue(elem))
		}
	}

	// Create and return new array with concatenated elements
	return &runtime.ArrayValue{
		Elements:  resultElements,
		ArrayType: firstArrayType,
	}
}

// ArrayHelperSlice extracts a slice from an array.
// Value-semantic elements are copied before storage to avoid aliasing.
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
	for idx, elem := range arr.Elements[start:end] {
		resultElements[idx] = runtime.CopyValue(elem)
	}

	// Create and return new array with sliced elements
	return &runtime.ArrayValue{
		Elements:  resultElements,
		ArrayType: arr.ArrayType,
	}
}
