package interp

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// indexArray performs array indexing with bounds checking.
func (i *Interpreter) indexArray(arr *ArrayValue, index int, expr *ast.IndexExpression) Value {
	if arr.ArrayType == nil {
		return i.newErrorWithLocation(expr, "array has no type information")
	}

	// Convert logical index to physical index
	var physicalIndex int
	if arr.ArrayType.IsStatic() {
		// Static array: check bounds and adjust for low bound
		lowBound := *arr.ArrayType.LowBound
		highBound := *arr.ArrayType.HighBound

		if index < lowBound || index > highBound {
			return i.newErrorWithLocation(expr, "index out of bounds: %d (bounds are %d..%d)", index, lowBound, highBound)
		}

		physicalIndex = index - lowBound
	} else {
		// Dynamic array: zero-based indexing
		if index < 0 || index >= len(arr.Elements) {
			return i.newErrorWithLocation(expr, "index out of bounds: %d (array length is %d)", index, len(arr.Elements))
		}

		physicalIndex = index
	}

	// Check physical bounds
	if physicalIndex < 0 || physicalIndex >= len(arr.Elements) {
		return i.newErrorWithLocation(expr, "index out of bounds: physical index %d, length %d", physicalIndex, len(arr.Elements))
	}

	// Return the element
	elem := arr.Elements[physicalIndex]
	if elem == nil {
		// Return properly typed zero value for uninitialized elements
		// This allows operators like NOT to work correctly with type information
		return getZeroValueForType(arr.ArrayType.ElementType, nil)
	}

	return elem
}

// indexString performs string indexing (returns a single-character string).
func (i *Interpreter) indexString(str *StringValue, index int, expr *ast.IndexExpression) Value {
	// DWScript strings are 1-indexed
	// Use rune-based indexing to handle UTF-8 correctly
	strLen := runeLength(str.Value)
	if index < 1 || index > strLen {
		return i.newErrorWithLocation(expr, "string index out of bounds: %d (string length is %d)", index, strLen)
	}

	// Get the character at the given position
	char, ok := runeAt(str.Value, index)
	if !ok {
		return i.newErrorWithLocation(expr, "string index out of bounds: %d", index)
	}
	return &StringValue{Value: string(char)}
}

// indexJSON performs JSON value indexing.
// Support both object property access (string index) and array element access (integer index).
//
// For JSON objects: obj['propertyName'] returns the value or nil if not found
// For JSON arrays: arr[index] returns the element or nil if out of bounds
func (i *Interpreter) indexJSON(jsonVal *JSONValue, indexVal Value, expr *ast.IndexExpression) Value {
	if jsonVal.Value == nil {
		return jsonValueToVariant(nil) // nil/null JSON value
	}

	kind := jsonVal.Value.Kind()

	// JSON Object: support string indexing
	if kind == 2 { // KindObject
		// Index must be a string for object property access
		indexStr, ok := indexVal.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(expr, "JSON object index must be a string, got %s", indexVal.Type())
		}

		// Get the property value (returns nil if not found)
		propValue := jsonVal.Value.ObjectGet(indexStr.Value)

		// Convert the JSON value to a Variant (nil becomes JSON null)
		return jsonValueToVariant(propValue)
	}

	// JSON Array: support integer indexing
	if kind == 3 { // KindArray
		// Index must be an integer
		indexInt, ok := indexVal.(*IntegerValue)
		if !ok {
			return i.newErrorWithLocation(expr, "JSON array index must be an integer, got %s", indexVal.Type())
		}

		index := int(indexInt.Value)

		// Get the array element (returns nil if out of bounds)
		elemValue := jsonVal.Value.ArrayGet(index)

		// Convert the JSON value to a Variant (nil becomes JSON null)
		return jsonValueToVariant(elemValue)
	}

	// Not an object or array
	return i.newErrorWithLocation(expr, "cannot index JSON %s", kind)
}

// evalArrayLiteral evaluates an array literal expression at runtime.
func (i *Interpreter) evalArrayLiteral(lit *ast.ArrayLiteralExpression) Value {
	if lit == nil {
		return &ErrorValue{Message: "nil array literal"}
	}

	// If semantic analysis annotated this literal as a set (array literal used for set context),
	// evaluate it via set literal handling so empty set assignments work.
	if i.semanticInfo() != nil {
		if typeAnnot := i.semanticInfo().GetType(lit); typeAnnot != nil && typeAnnot.Name != "" {
			var resolved types.Type
			if typeVal, err := i.resolveType(typeAnnot.Name); err == nil {
				resolved = typeVal
			}

			var setType *types.SetType
			if resolved != nil {
				if _, isArray := resolved.(*types.ArrayType); !isArray {
					if candidate, ok := types.GetUnderlyingType(resolved).(*types.SetType); ok {
						setType = candidate
					}
				}
			} else {
				setType = i.parseInlineSetType(typeAnnot.Name)
			}

			if setType != nil {
				setLit := &ast.SetLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token:  lit.Token,
							EndPos: lit.End(),
						},
					},
					Elements: lit.Elements,
				}

				// Propagate the type annotation so evalSetLiteral can infer element type.
				i.semanticInfo().SetType(setLit, typeAnnot)
				return i.evalSetLiteral(setLit)
			}
		}
	}

	// If a type annotation exists, resolve it first so we can evaluate elements
	// with the expected element type (important for nested array literals).
	arrayType, errVal := i.arrayTypeFromLiteral(lit)
	if errVal != nil {
		return errVal
	}

	elementCount := len(lit.Elements)
	evaluated := make([]Value, elementCount)
	elementTypes := make([]types.Type, elementCount)

	for idx, elem := range lit.Elements {
		var val Value

		// If we already know the array type and the element is an array literal,
		// evaluate it with the expected element type to avoid incompatible nested arrays.
		if arrayType != nil {
			if elemLit, ok := elem.(*ast.ArrayLiteralExpression); ok {
				if expectedElemArr, ok := arrayType.ElementType.(*types.ArrayType); ok {
					val = i.evalArrayLiteralWithExpected(elemLit, expectedElemArr)
				}
			}
		}

		if val == nil {
			val = i.Eval(elem)
		}

		if isError(val) {
			return val
		}
		evaluated[idx] = val
		elementTypes[idx] = i.typeFromValue(val)
	}

	if arrayType == nil {
		inferred, errVal := i.inferArrayTypeFromValues(lit, elementTypes)
		if errVal != nil {
			return errVal
		}
		if inferred == nil {
			if elementCount == 0 {
				return i.newErrorWithLocation(lit, "cannot infer type for empty array literal")
			}
			return i.newErrorWithLocation(lit, "cannot determine array type for literal")
		}
		arrayType = inferred
	}

	coerced, errVal := i.coerceArrayElements(arrayType, evaluated, elementTypes, lit)
	if errVal != nil {
		return errVal
	}

	var elements []Value
	if arrayType.IsStatic() {
		expectedSize := arrayType.Size()
		if elementCount != expectedSize {
			return i.newErrorWithLocation(lit, "array literal has %d elements, expected %d", elementCount, expectedSize)
		}
		elements = make([]Value, expectedSize)
		copy(elements, coerced)
	} else {
		elements = append([]Value(nil), coerced...)
	}

	return &ArrayValue{
		ArrayType: arrayType,
		Elements:  elements,
	}
}

// arrayTypeFromLiteral resolves the array type for a literal using its type annotation, if available.
func (i *Interpreter) arrayTypeFromLiteral(lit *ast.ArrayLiteralExpression) (*types.ArrayType, Value) {
	var typeAnnot *ast.TypeAnnotation
	if i.semanticInfo() != nil {
		typeAnnot = i.semanticInfo().GetType(lit)
	}
	if typeAnnot == nil || typeAnnot.Name == "" {
		return nil, nil
	}

	return i.arrayTypeByName(typeAnnot.Name, lit)
}

// inferArrayTypeFromValues infers a dynamic array type based on evaluated element types.
func (i *Interpreter) inferArrayTypeFromValues(lit *ast.ArrayLiteralExpression, elementTypes []types.Type) (*types.ArrayType, Value) {
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

		if inferred.Equals(types.INTEGER) && underlying.Equals(types.FLOAT) {
			inferred = types.FLOAT
			continue
		}

		if inferred.Equals(types.FLOAT) && underlying.Equals(types.INTEGER) {
			continue
		}

		return nil, i.newErrorWithLocation(lit.Elements[idx], "array element %d has incompatible type (got %s, expected %s)",
			idx+1, underlying.String(), inferred.String())
	}

	if inferred == nil {
		return nil, nil
	}

	// Prefer static array type when inferring without an explicit annotation.
	// This gives value semantics (copy on assignment) matching DWScript behavior
	// for array literals assigned to implicitly-typed variables.
	size := len(lit.Elements)
	if size == 0 {
		return types.NewDynamicArrayType(types.GetUnderlyingType(inferred)), nil
	}
	return types.NewStaticArrayType(types.GetUnderlyingType(inferred), 0, size-1), nil
}

// coerceArrayElements ensures evaluated values conform to the array's element type.
func (i *Interpreter) coerceArrayElements(arrayType *types.ArrayType, values []Value, valueTypes []types.Type, lit *ast.ArrayLiteralExpression) ([]Value, Value) {
	coerced := make([]Value, len(values))

	elementType := arrayType.ElementType
	if elementType == nil {
		return nil, i.newErrorWithLocation(lit, "array literal has no element type information")
	}
	underlyingElementType := types.GetUnderlyingType(elementType)

	for idx, val := range values {
		var valType types.Type
		if idx < len(valueTypes) && valueTypes[idx] != nil {
			valType = types.GetUnderlyingType(valueTypes[idx])
		}

		// Box values when expected element type is Variant
		// This enables heterogeneous arrays like [1, "hello", 3.14, true] for Format()
		// Replaces the old CONST workaround with proper Variant boxing
		if underlyingElementType.Equals(types.VARIANT) {
			coerced[idx] = BoxVariant(val)
			continue
		}

		if _, isNil := val.(*NilValue); isNil {
			switch underlyingElementType.TypeKind() {
			case "CLASS", "INTERFACE", "ARRAY":
				coerced[idx] = val
				continue
			default:
				return nil, i.newErrorWithLocation(lit.Elements[idx], "cannot assign nil to %s", underlyingElementType.String())
			}
		}

		if valType == nil {
			return nil, i.newErrorWithLocation(lit.Elements[idx], "cannot determine type for array element %d", idx+1)
		}

		if underlyingElementType.Equals(valType) {
			coerced[idx] = val
			continue
		}

		if underlyingElementType.Equals(types.FLOAT) && valType.Equals(types.INTEGER) {
			if intVal, ok := val.(*IntegerValue); ok {
				coerced[idx] = &FloatValue{Value: float64(intVal.Value)}
				continue
			}
		}

		if valType.TypeKind() == "ARRAY" && underlyingElementType.TypeKind() == "ARRAY" {
			if arrayVal, ok := val.(*ArrayValue); ok && arrayVal.ArrayType != nil {
				if types.IsCompatible(arrayVal.ArrayType, underlyingElementType) || types.IsCompatible(underlyingElementType, arrayVal.ArrayType) {
					coerced[idx] = val
					continue
				}
			}
		}

		if types.IsCompatible(valType, underlyingElementType) {
			coerced[idx] = val
			continue
		}

		return nil, i.newErrorWithLocation(lit.Elements[idx], "array element %d has incompatible type (got %s, expected %s)",
			idx+1, val.Type(), underlyingElementType.String())
	}

	return coerced, nil
}

// typeFromValue maps a runtime value to its compile-time type representation.
func (i *Interpreter) typeFromValue(val Value) types.Type {
	switch v := val.(type) {
	case *IntegerValue:
		return types.INTEGER
	case *FloatValue:
		return types.FLOAT
	case *StringValue:
		return types.STRING
	case *BooleanValue:
		return types.BOOLEAN
	case *NilValue:
		return types.NIL
	case *ArrayValue:
		return v.ArrayType
	case *EnumValue:
		if enumMetadata := i.typeSystem.LookupEnumMetadata(v.TypeName); enumMetadata != nil {
			if etv, ok := enumMetadata.(*EnumTypeValue); ok {
				return etv.EnumType
			}
		}
		return nil
	case *ObjectInstance:
		// Need concrete ClassInfo for classTypeForName
		concreteClass, ok := v.Class.(*ClassInfo)
		if !ok {
			return nil
		}
		return i.classTypeForName(concreteClass)
	case *RecordValue:
		return v.RecordType
	default:
		return nil
	}
}

// arrayTypeByName resolves an array type by name or inline signature.
func (i *Interpreter) arrayTypeByName(typeName string, node ast.Node) (*types.ArrayType, Value) {
	if typeName == "" {
		return nil, nil
	}

	if arr := i.parseInlineArrayType(typeName); arr != nil {
		return arr, nil
	}

	resolved, err := i.resolveType(typeName)
	if err != nil {
		return nil, i.newErrorWithLocation(node, "unknown array type '%s'", typeName)
	}

	if arr, ok := resolved.(*types.ArrayType); ok {
		return arr, nil
	}

	if arr, ok := types.GetUnderlyingType(resolved).(*types.ArrayType); ok {
		return arr, nil
	}

	return nil, i.newErrorWithLocation(node, "type '%s' is not an array type", typeName)
}

// evalArrayLiteralWithExpected evaluates an array literal with an expected array type.
func (i *Interpreter) evalArrayLiteralWithExpected(lit *ast.ArrayLiteralExpression, expected *types.ArrayType) Value {
	if expected == nil {
		return i.evalArrayLiteral(lit)
	}

	// Ensure semanticInfo exists for type annotation
	wasNil := i.semanticInfo() == nil
	if wasNil {
		i.engineState.SemanticInfo = ast.NewSemanticInfo()
	}

	// Temporarily set type annotation for evaluation
	prevType := i.semanticInfo().GetType(lit)
	annotation := &ast.TypeAnnotation{Token: lit.Token, Name: expected.String()}
	i.semanticInfo().SetType(lit, annotation)

	result := i.evalArrayLiteral(lit)

	// Restore previous type
	if prevType != nil {
		i.semanticInfo().SetType(lit, prevType)
	} else {
		i.semanticInfo().ClearType(lit)
	}

	// Clean up semanticInfo if we created it
	if wasNil {
		i.engineState.SemanticInfo = nil
	}

	return result
}

// ============================================================================
// Array Instantiation with new Keyword
// ============================================================================

// createMultiDimArray creates a multi-dimensional array with the given dimensions.
// For 1D arrays, creates a single array with the specified size.
// For multi-dimensional arrays, recursively creates nested arrays.
func (i *Interpreter) createMultiDimArray(elementType types.Type, dimensions []int) *ArrayValue {
	if len(dimensions) == 0 {
		// This shouldn't happen, but handle gracefully
		return &ArrayValue{
			ArrayType: types.NewDynamicArrayType(elementType),
			Elements:  []Value{},
		}
	}

	size := dimensions[0]

	if len(dimensions) == 1 {
		// Base case: 1D array
		// Create array type
		arrayType := types.NewDynamicArrayType(elementType)

		// Create elements filled with zero values
		elements := make([]Value, size)
		for idx := 0; idx < size; idx++ {
			elements[idx] = i.createZeroValueForType(elementType)
		}

		return &ArrayValue{
			ArrayType: arrayType,
			Elements:  elements,
		}
	}

	// Recursive case: multi-dimensional array
	// The element type for this level is an array of the remaining dimensions
	innerElementType := i.buildArrayTypeForDimensions(elementType, dimensions[1:])

	// Create the outer array type
	arrayType := types.NewDynamicArrayType(innerElementType)

	// Create elements, each being an array of the remaining dimensions
	elements := make([]Value, size)
	for idx := 0; idx < size; idx++ {
		elements[idx] = i.createMultiDimArray(elementType, dimensions[1:])
	}

	return &ArrayValue{
		ArrayType: arrayType,
		Elements:  elements,
	}
}

// buildArrayTypeForDimensions builds an array type for the given dimensions.
// For example, dimensions [3, 4] with elementType Integer produces:
// array of array of Integer
func (i *Interpreter) buildArrayTypeForDimensions(elementType types.Type, dimensions []int) types.Type {
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

// ensureClassPropertyInstance makes sure a field-backed class property referenced by a MemberAccessExpression
// is initialized. If the backing field is nil, it instantiates the class and assigns it.
func (i *Interpreter) ensureClassPropertyInstance(ma *ast.MemberAccessExpression) Value {
	if ma == nil {
		return nil
	}

	ownerVal := i.Eval(ma.Object)
	if isError(ownerVal) {
		return ownerVal
	}

	ownerObj, ok := AsObject(ownerVal)
	if !ok {
		return ownerVal
	}

	propInfo := ownerObj.Class.LookupProperty(ma.Member.Value)
	if propInfo == nil {
		return ownerVal
	}

	// Extract the actual *types.PropertyInfo from the Impl field
	typesPropInfo, ok := propInfo.Impl.(*types.PropertyInfo)
	if !ok {
		return ownerVal
	}

	classType, isClass := typesPropInfo.Type.(*types.ClassType)
	if !isClass || typesPropInfo.ReadKind != types.PropAccessField || typesPropInfo.ReadSpec == "" {
		return ownerVal
	}

	currentVal := ownerObj.GetField(typesPropInfo.ReadSpec)
	if currentVal != nil && currentVal.Type() != "NIL" {
		return currentVal
	}

	if classInfo := i.resolveClassInfoByName(classType.Name); classInfo != nil {
		newInst := NewObjectInstance(classInfo)
		ownerObj.SetField(propInfo.ReadSpec, newInst)
		return newInst
	}

	return ownerVal
}

// createZeroValueForType creates a zero value for the given type.
// This is similar to createZeroValue but works with types.Type instead of ast.TypeAnnotation.
func (i *Interpreter) createZeroValueForType(typ types.Type) Value {
	if typ == nil {
		return &NilValue{}
	}

	switch typ {
	case types.INTEGER:
		return &IntegerValue{Value: 0}
	case types.FLOAT:
		return &FloatValue{Value: 0.0}
	case types.STRING:
		return &StringValue{Value: ""}
	case types.BOOLEAN:
		return &BooleanValue{Value: false}
	default:
		// For complex types (arrays, records, etc.), initialize them properly
		if arrayType, ok := typ.(*types.ArrayType); ok {
			return NewArrayValue(arrayType)
		}
		// Initialize record types properly for array elements
		if recordType, ok := typ.(*types.RecordType); ok {
			return i.createRecordValue(recordType)
		}
		return &NilValue{}
	}
}
