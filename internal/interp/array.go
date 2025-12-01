package interp

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Array Declaration Evaluation
// ============================================================================

// evalArrayDeclaration evaluates an array type declaration.
// Example: type TMyArray = array[1..10] of Integer;
func (i *Interpreter) evalArrayDeclaration(decl *ast.ArrayDecl) Value {
	if decl == nil {
		return &ErrorValue{Message: "nil array declaration"}
	}

	arrayName := decl.Name.Value

	// Build the array type from the declaration
	arrayTypeAnnotation := decl.ArrayType
	if arrayTypeAnnotation == nil {
		return i.newErrorWithLocation(decl, "invalid array type declaration")
	}

	// Resolve the element type
	elementTypeName := arrayTypeAnnotation.ElementType.String()
	elementType, err := i.resolveType(elementTypeName)
	if err != nil {
		return i.newErrorWithLocation(decl, "unknown element type '%s'", elementTypeName)
	}

	// Create the array type
	var arrayType *types.ArrayType
	if arrayTypeAnnotation.IsDynamic() {
		arrayType = types.NewDynamicArrayType(elementType)
	} else {
		// Evaluate bound expressions at runtime
		lowBoundVal := i.Eval(arrayTypeAnnotation.LowBound)
		if isError(lowBoundVal) {
			return lowBoundVal
		}
		highBoundVal := i.Eval(arrayTypeAnnotation.HighBound)
		if isError(highBoundVal) {
			return highBoundVal
		}

		// Extract integer values
		lowBound, ok := lowBoundVal.(*IntegerValue)
		if !ok {
			return i.newErrorWithLocation(decl, "array lower bound must be an integer")
		}
		highBound, ok := highBoundVal.(*IntegerValue)
		if !ok {
			return i.newErrorWithLocation(decl, "array upper bound must be an integer")
		}

		arrayType = types.NewStaticArrayType(elementType, int(lowBound.Value), int(highBound.Value))
	}

	// Register array type in TypeSystem only
	i.typeSystem.RegisterArrayType(arrayName, arrayType)

	return &NilValue{} // Type declarations don't return a value
}

// evalIndexExpression evaluates array/string indexing: arr[i]
func (i *Interpreter) evalIndexExpression(expr *ast.IndexExpression) Value {
	if expr == nil {
		return &ErrorValue{Message: "nil index expression"}
	}

	// Check if this might be a multi-index property access
	// We only flatten indices if the base is a MemberAccessExpression (property access)
	// For regular array access like arr[i][j], we process each level separately
	base, indices := evaluator.CollectIndices(expr)

	// Check if this is indexed property access: obj.Property[index1, index2, ...]
	// Only flatten indices for property access, not for regular arrays
	if memberAccess, ok := base.(*ast.MemberAccessExpression); ok {
		// Evaluate the object being accessed
		objVal := i.Eval(memberAccess.Object)
		if isError(objVal) {
			return objVal
		}

		// Interface indexed property access
		if intfInst, ok := objVal.(*InterfaceInstance); ok {
			if intfInst.Object == nil {
				return i.newErrorWithLocation(expr, "Interface is nil")
			}
			if propInfo := intfInst.Interface.GetProperty(memberAccess.Member.Value); propInfo != nil && propInfo.IsIndexed {
				indexVals := make([]Value, len(indices))
				for idx, indexExpr := range indices {
					indexVals[idx] = i.Eval(indexExpr)
					if isError(indexVals[idx]) {
						return indexVals[idx]
					}
				}
				if obj, ok := AsObject(intfInst.Object); ok {
					// Task 3.5.20: Extract *types.PropertyInfo from Impl field
					typesPropertyInfo, ok := propInfo.Impl.(*types.PropertyInfo)
					if !ok {
						return i.newErrorWithLocation(expr, "invalid property info type")
					}
					return i.evalIndexedPropertyRead(obj, typesPropertyInfo, indexVals, expr)
				}
				return i.newErrorWithLocation(expr, "interface underlying object is not a class instance")
			}
			// unwrap for further checks
			objVal = intfInst.Object
		}

		// Check if it's a class instance with an indexed property
		if obj, ok := AsObject(objVal); ok {
			propInfo := obj.Class.LookupProperty(memberAccess.Member.Value)
			if propInfo != nil && propInfo.IsIndexed {
				// Extract the actual *types.PropertyInfo from the Impl field
				typesPropInfo, ok := propInfo.Impl.(*types.PropertyInfo)
				if !ok {
					return i.NewError("invalid property info implementation")
				}

				// This is a multi-index property access: flatten and evaluate ALL indices
				indexVals := make([]Value, len(indices))
				for idx, indexExpr := range indices {
					indexVals[idx] = i.Eval(indexExpr)
					if isError(indexVals[idx]) {
						return indexVals[idx]
					}
				}

				// Call indexed property read with all indices
				return i.evalIndexedPropertyRead(obj, typesPropInfo, indexVals, expr)
			}
		}

		// Check if it's a record instance with an indexed property
		if recordVal, ok := objVal.(*RecordValue); ok {
			memberNameLower := strings.ToLower(memberAccess.Member.Value)
			if propInfo, exists := recordVal.RecordType.Properties[memberNameLower]; exists {
				// This is an array property access
				if propInfo.ReadField != "" {
					// Evaluate all indices
					indexVals := make([]Value, len(indices))
					for idx, indexExpr := range indices {
						indexVals[idx] = i.Eval(indexExpr)
						if isError(indexVals[idx]) {
							return indexVals[idx]
						}
					}

					// Check if ReadField is a method (getter with parameters)
					if getterMethod := GetRecordMethod(recordVal, propInfo.ReadField); getterMethod != nil {
						// Call the getter method with indices as arguments
						methodCall := &ast.MethodCallExpression{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: expr.Token,
								},
							},
							Object:    memberAccess.Object,
							Method:    &ast.Identifier{Value: propInfo.ReadField, TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{Token: expr.Token}}},
							Arguments: make([]ast.Expression, len(indexVals)),
						}

						// Create temporary identifiers for each index value
						for idx, indexVal := range indexVals {
							tempName := fmt.Sprintf("__temp_index_%d__", idx)
							i.env.Define(tempName, indexVal)
							methodCall.Arguments[idx] = &ast.Identifier{
								Value:               tempName,
								TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{Token: expr.Token}},
							}
						}

						return i.evalMethodCall(methodCall)
					}

					return i.newErrorWithLocation(expr, "array property '%s' read accessor '%s' is not a method",
						memberAccess.Member.Value, propInfo.ReadField)
				}
				return i.newErrorWithLocation(expr, "property '%s' is write-only", memberAccess.Member.Value)
			}
		}
	}

	// Not a property access - this is regular array/string indexing
	// Process ONLY the outermost index, not all nested indices
	// This allows FData[x][y] to work as: (FData[x])[y]
	leftVal := i.Eval(expr.Left)
	if isError(leftVal) {
		return leftVal
	}

	// Evaluate the index for this level only
	indexVal := i.Eval(expr.Index)
	if isError(indexVal) {
		return indexVal
	}

	// Auto-instantiate class-typed fields when accessed through a property that returns nil.
	// This allows patterns like obj.Sub['x'] to work when Sub is a field-backed class property.
	if leftVal != nil && leftVal.Type() == "NIL" {
		if memberAccess, ok := expr.Left.(*ast.MemberAccessExpression); ok {
			ownerVal := i.Eval(memberAccess.Object)
			if !isError(ownerVal) {
				// If the owner is already an object, try to initialize its field-backed property.
				if ownerObj, ok := AsObject(ownerVal); ok {
					if propInfo := ownerObj.Class.LookupProperty(memberAccess.Member.Value); propInfo != nil {
						// Extract the actual *types.PropertyInfo from the Impl field
						if typesPropInfo, ok := propInfo.Impl.(*types.PropertyInfo); ok {
							if classType, ok := typesPropInfo.Type.(*types.ClassType); ok {
								if classInfo := i.resolveClassInfoByName(classType.Name); classInfo != nil {
									// Only field-backed properties can be initialized directly
									if typesPropInfo.ReadKind == types.PropAccessField && typesPropInfo.ReadSpec != "" {
										newInst := NewObjectInstance(classInfo)
										ownerObj.SetField(typesPropInfo.ReadSpec, newInst)
										leftVal = newInst
									}
								}
							}
						}
					}
				} else if ownerVal.Type() == "NIL" {
					// If the owner itself is a nil class property (e.g., data.Sub),
					// try to auto-create it based on its declaration chain.
					if maObj, ok := memberAccess.Object.(*ast.MemberAccessExpression); ok {
						if ensured := i.ensureClassPropertyInstance(maObj); ensured != nil && ensured.Type() != "NIL" {
							ownerVal = ensured
							if ownerObj, ok := AsObject(ensured); ok {
								if propInfo := ownerObj.Class.LookupProperty(memberAccess.Member.Value); propInfo != nil {
									// Extract the actual *types.PropertyInfo from the Impl field
									if typesPropInfo, ok := propInfo.Impl.(*types.PropertyInfo); ok {
										if classType, ok := typesPropInfo.Type.(*types.ClassType); ok {
											if classInfo := i.resolveClassInfoByName(classType.Name); classInfo != nil {
												if typesPropInfo.ReadKind == types.PropAccessField && typesPropInfo.ReadSpec != "" {
													newInst := NewObjectInstance(classInfo)
													ownerObj.SetField(typesPropInfo.ReadSpec, newInst)
													leftVal = newInst
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// Handle default indexed properties on interface values (e.g., intf['x'])
	if intfInst, ok := leftVal.(*InterfaceInstance); ok {
		if intfInst.Object == nil {
			return i.newErrorWithLocation(expr, "Interface is nil")
		}
		// Task 3.5.20: Use GetDefaultProperty() (public) instead of getDefaultProperty()
		if propInfo := intfInst.Interface.GetDefaultProperty(); propInfo != nil && propInfo.IsIndexed {
			if obj, ok := AsObject(intfInst.Object); ok {
				// Extract *types.PropertyInfo from Impl field
				typesPropertyInfo, ok := propInfo.Impl.(*types.PropertyInfo)
				if !ok {
					return i.newErrorWithLocation(expr, "invalid property info type")
				}
				return i.evalIndexedPropertyRead(obj, typesPropertyInfo, []Value{indexVal}, expr)
			}
			return i.newErrorWithLocation(expr, "interface underlying object is not a class instance")
		}
		// unwrap for further default property checks
		leftVal = intfInst.Object
	}

	// Check if left side is an object with a default property
	// This allows obj[index] to be equivalent to obj.DefaultProperty[index]
	if obj, ok := AsObject(leftVal); ok {
		defaultProp := obj.Class.GetDefaultProperty()
		if defaultProp != nil {
			// Extract the actual *types.PropertyInfo from the Impl field
			typesDefaultProp, ok := defaultProp.Impl.(*types.PropertyInfo)
			if !ok {
				return i.NewError("invalid default property info implementation")
			}

			// Route to the default indexed property
			// For now, we only support single-index default properties
			// Multi-index would need to collect all indices from nested IndexExpressions
			return i.evalIndexedPropertyRead(obj, typesDefaultProp, []Value{indexVal}, expr)
		}
	}

	// Check if left side is a record with a default property
	// This allows record[index] to be equivalent to record.DefaultProperty[index]
	if recordVal, ok := leftVal.(*RecordValue); ok {
		// Find the default property
		var defaultProp *types.RecordPropertyInfo
		for _, propInfo := range recordVal.RecordType.Properties {
			if propInfo.IsDefault {
				defaultProp = propInfo
				break
			}
		}

		if defaultProp != nil {
			// Call the getter method with the index
			if defaultProp.ReadField != "" {
				if getterMethod := GetRecordMethod(recordVal, defaultProp.ReadField); getterMethod != nil {
					// Call the getter method with index as argument
					methodCall := &ast.MethodCallExpression{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{Token: expr.Token},
						},
						Object: expr.Left,
						Method: &ast.Identifier{Value: defaultProp.ReadField, TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{Token: expr.Token}}},
						Arguments: []ast.Expression{
							&ast.Identifier{Value: "__temp_default_index__", TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{Token: expr.Token}}},
						},
					}
					i.env.Define("__temp_default_index__", indexVal)
					return i.evalMethodCall(methodCall)
				}
				return i.newErrorWithLocation(expr, "default property read accessor '%s' is not a method", defaultProp.ReadField)
			}
			return i.newErrorWithLocation(expr, "default property is write-only")
		}
	}

	// Check if left side is a JSON value (wrapped in Variant)
	// JSON objects support string indexing: obj['propertyName']
	// JSON arrays support integer indexing: arr[0]
	unwrapped := unwrapVariant(leftVal)
	if jsonVal, ok := unwrapped.(*JSONValue); ok {
		return i.indexJSON(jsonVal, indexVal, expr)
	}

	// Index must be an ordinal for arrays and strings
	var index int
	switch iv := indexVal.(type) {
	case *IntegerValue:
		index = int(iv.Value)
	case *EnumValue:
		index = iv.OrdinalValue
	case *BooleanValue:
		if iv.Value {
			index = 1
		}
	case *SubrangeValue:
		index = iv.Value
	default:
		return i.newErrorWithLocation(expr, "index must be an ordinal value, got %s", indexVal.Type())
	}

	// Check if left side is an array
	if arrayVal, ok := leftVal.(*ArrayValue); ok {
		return i.indexArray(arrayVal, index, expr)
	}

	// Check if left side is a string
	if strVal, ok := leftVal.(*StringValue); ok {
		return i.indexString(strVal, index, expr)
	}

	return i.newErrorWithLocation(expr, "cannot index type %s", leftVal.Type())
}

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
	if i.semanticInfo != nil {
		if typeAnnot := i.semanticInfo.GetType(lit); typeAnnot != nil && typeAnnot.Name != "" {
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
				i.semanticInfo.SetType(setLit, typeAnnot)
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
	if i.semanticInfo != nil {
		typeAnnot = i.semanticInfo.GetType(lit)
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
	wasNil := i.semanticInfo == nil
	if wasNil {
		i.semanticInfo = ast.NewSemanticInfo()
	}

	// Temporarily set type annotation for evaluation
	prevType := i.semanticInfo.GetType(lit)
	annotation := &ast.TypeAnnotation{Token: lit.Token, Name: expected.String()}
	i.semanticInfo.SetType(lit, annotation)

	result := i.evalArrayLiteral(lit)

	// Restore previous type
	if prevType != nil {
		i.semanticInfo.SetType(lit, prevType)
	} else {
		i.semanticInfo.ClearType(lit)
	}

	// Clean up semanticInfo if we created it
	if wasNil {
		i.semanticInfo = nil
	}

	return result
}

// ============================================================================
// Array Instantiation with new Keyword
// ============================================================================

// evalNewArrayExpression evaluates a new array expression.
// Example: new Integer[10] or new String[3, 4]
// Implement runtime support for dynamic array instantiation.
func (i *Interpreter) evalNewArrayExpression(expr *ast.NewArrayExpression) Value {
	if expr == nil {
		return &ErrorValue{Message: "nil new array expression"}
	}

	// Resolve the element type
	if expr.ElementTypeName == nil {
		return i.newErrorWithLocation(expr, "new array expression missing element type")
	}

	elementTypeName := expr.ElementTypeName.Value
	elementType, err := i.resolveType(elementTypeName)
	if err != nil {
		return i.newErrorWithLocation(expr, "unknown element type '%s': %s", elementTypeName, err)
	}

	// Evaluate each dimension expression to get integer sizes
	if len(expr.Dimensions) == 0 {
		return i.newErrorWithLocation(expr, "new array expression must have at least one dimension")
	}

	dimensions := make([]int, len(expr.Dimensions))
	for idx, dimExpr := range expr.Dimensions {
		dimVal := i.Eval(dimExpr)
		if isError(dimVal) {
			return dimVal
		}

		// Dimension must be an integer
		dimInt, ok := dimVal.(*IntegerValue)
		if !ok {
			return i.newErrorWithLocation(expr, "array dimension must be an integer, got %s", dimVal.Type())
		}

		// Validate dimension is positive
		if dimInt.Value <= 0 {
			return i.newErrorWithLocation(expr, "array dimension must be positive, got %d", dimInt.Value)
		}

		dimensions[idx] = int(dimInt.Value)
	}

	// Create the multi-dimensional array
	return i.createMultiDimArray(elementType, dimensions)
}

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
