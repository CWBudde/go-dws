package semantic

import (
	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Array Analysis
// ============================================================================

// analyzeArrayDecl analyzes an array type declaration
func (a *Analyzer) analyzeArrayDecl(decl *ast.ArrayDecl) {
	defer func() {
		if decl != nil && decl.Name != nil {
			a.recordDeclaredType(decl, decl.Name.Value)
		}
	}()

	if decl == nil {
		return
	}

	arrayName := decl.Name.Value

	// Check if array type is already declared
	// Use lowercase for case-insensitive duplicate check
	if a.hasType(arrayName) {
		a.addError("%s", errors.FormatNameAlreadyExists(arrayName, decl.Token.Pos.Line, decl.Token.Pos.Column))
		return
	}

	// Validate array type
	arrayType := decl.ArrayType
	if arrayType == nil {
		a.addError("invalid array type declaration at %s", decl.Token.Pos.String())
		return
	}

	// Resolve the element type using resolveType helper
	elementTypeName := getTypeExpressionName(arrayType.ElementType)
	elementType, err := a.resolveType(elementTypeName)
	if err != nil {
		a.addError("unknown type '%s' at %s", elementTypeName, decl.Token.Pos.String())
		return
	}

	// Create the array type
	var arrType *types.ArrayType
	if arrayType.IsDynamic() {
		arrType = types.NewDynamicArrayType(elementType)
	} else {
		lowBound, highBound, indexType, ok := a.resolveOrdinalArrayBounds(arrayType.LowBound, arrayType.HighBound)
		if !ok {
			return
		}
		if indexType != nil && indexType.TypeKind() != "INTEGER" {
			arrType = types.NewStaticArrayTypeWithIndexType(elementType, indexType, lowBound, highBound)
		} else {
			arrType = types.NewStaticArrayType(elementType, lowBound, highBound)
		}
	}

	// Register the array type in the arrays registry
	// Use lowercase key for case-insensitive lookup
	a.registerTypeWithPos(arrayName, arrType, decl.Token.Pos)
}

// analyzeIndexExpression analyzes an array/string indexing expression
func (a *Analyzer) analyzeIndexExpression(expr *ast.IndexExpression) types.Type {
	if expr == nil {
		return nil
	}

	// Special-case indexed properties: obj.Prop[index]
	if memberAccess, ok := expr.Left.(*ast.MemberAccessExpression); ok {
		if propType := a.analyzeIndexedPropertyAccess(memberAccess, expr); propType != nil {
			return propType
		}
	}

	// Multi-index properties: obj.Prop[i, j] parses as obj.Prop[i][j], so the
	// whole chain has to be resolved against one property declaration.
	if propType, handled := a.analyzeMultiIndexPropertyAccess(expr); handled {
		return propType
	}

	// Analyze the left side (what's being indexed)
	leftType := a.analyzeExpression(expr.Left)
	if leftType == nil {
		// Error already reported
		return nil
	}

	// Allow default indexed properties on classes (obj[index] -> obj.DefaultProperty[index])
	if classType, ok := types.GetUnderlyingType(leftType).(*types.ClassType); ok {
		if defaultProp := a.getDefaultClassProperty(classType); defaultProp != nil {
			expectedIndexTypes := a.getIndexedPropertyParamTypes(defaultProp, classType)
			if len(expectedIndexTypes) > 0 {
				indexType := a.analyzeExpressionWithExpectedType(expr.Index, expectedIndexTypes[0])
				if indexType != nil && !a.canAssign(indexType, expectedIndexTypes[0]) {
					a.addStructuredError(NewArrayIndexError(expr.Index.Pos(), expectedIndexTypes[0].String(), indexType.String()))
					return defaultProp.Type
				}
			} else {
				a.analyzeExpression(expr.Index)
			}
			return defaultProp.Type
		}
		a.addStructuredError(NewNoDefaultPropertyError(expr.Token.Pos, classType.Name))
		return nil
	}

	// Associative array indexing: a[key] where key is assignable to KeyType,
	// yielding the element type. New keys are legal (validated on assignment),
	// so a read of a missing key returns the element's zero value at runtime.
	if assoc, isAssoc := types.GetUnderlyingType(leftType).(*types.AssociativeArrayType); isAssoc {
		keyType := a.analyzeExpression(expr.Index)
		if keyType == nil {
			return nil
		}
		if !types.GetUnderlyingType(keyType).Equals(types.VARIANT) && !a.canAssign(keyType, assoc.KeyType) {
			a.addStructuredError(NewArrayIndexError(expr.Index.Pos(), assoc.KeyType.String(), semanticTypeNameForDiagnostic(keyType)))
			return nil
		}
		return assoc.ElementType
	}

	// Check if left side is an array type
	arrayType, ok := leftType.(*types.ArrayType)
	if !ok {
		// Also check for string indexing
		if leftType.Equals(types.STRING) {
			// String indexing returns a string (single character)
			// Check index type
			indexType := a.analyzeExpression(expr.Index)
			if indexType != nil && !indexType.Equals(types.INTEGER) {
				a.addStructuredError(NewArrayIndexError(expr.Index.Pos(), "Integer", semanticTypeNameForDiagnostic(indexType)))
				return nil
			}
			return types.STRING
		}

		// JSONVariant indexing (v['key'], v[3]) yields another JSONVariant.
		if types.IsJSONVariant(leftType) {
			a.analyzeExpression(expr.Index)
			return types.JSON_VARIANT
		}

		// Allow indexing of Variant types (can contain JSON objects/arrays)
		// At runtime, the interpreter will handle JSON object property access and array indexing
		if leftType.Equals(types.VARIANT) {
			// Analyze the index expression (can be string or integer)
			a.analyzeExpression(expr.Index)
			// Result type is Variant since we don't know the JSON structure at compile time
			return types.VARIANT
		}

		// Check if this is a record type with a default property
		if recordType, isRecord := leftType.(*types.RecordType); isRecord {
			// Look for a default property (marked with IsDefault)
			var defaultProp *types.RecordPropertyInfo
			for _, propInfo := range recordType.Properties {
				if propInfo.IsDefault {
					defaultProp = propInfo
					break
				}
			}

			if defaultProp != nil {
				// Analyze the index expression
				// TODO: Validate index type matches property index parameter types
				a.analyzeExpression(expr.Index)
				return defaultProp.Type
			}
		}

		a.addStructuredError(NewCannotIndexTypeError(expr.Token.Pos, leftType.String()))
		return nil
	}

	// Analyze the index expression
	indexType := a.analyzeExpression(expr.Index)
	if indexType == nil {
		// Error already reported
		return nil
	}

	if isZeroArgIntToStrCall(expr.Index) {
		pos := expr.Index.Pos()
		a.addError(`There is no overloaded version of "IntToStr" that can be called with these arguments at %s`,
			pos.String())
		a.addStructuredError(NewArrayIndexError(pos, "Integer", "Any Type"))
		return nil
	}

	if arrayType.IndexType != nil {
		if types.GetUnderlyingType(indexType).Equals(types.VARIANT) {
			return arrayType.ElementType
		}
		expectedIndexType := types.GetUnderlyingType(arrayType.IndexType)
		indexUnderlying := types.GetUnderlyingType(indexType)
		if !indexUnderlying.Equals(expectedIndexType) {
			pos := expr.Index.Pos()
			a.addStructuredError(NewArrayIndexError(pos, expectedIndexType.String(), semanticTypeNameForDiagnostic(indexType)))
			return nil
		}
		return arrayType.ElementType
	}

	// Index must be an integer for ordinary arrays. Variants are allowed and
	// are validated at runtime.
	indexUnderlying := types.GetUnderlyingType(indexType)
	if !indexUnderlying.Equals(types.VARIANT) && !indexUnderlying.Equals(types.INTEGER) {
		pos := expr.Index.Pos()
		a.addStructuredError(NewArrayIndexError(pos, "Integer", semanticTypeNameForDiagnostic(indexType)))
		return nil
	}

	// Return the element type of the array
	return arrayType.ElementType
}

func isZeroArgIntToStrCall(expr ast.Expression) bool {
	callExpr, ok := expr.(*ast.CallExpression)
	if !ok {
		return false
	}
	identExpr, ok := callExpr.Function.(*ast.Identifier)
	if !ok {
		return false
	}
	return ident.Equal(identExpr.Value, "IntToStr") && len(callExpr.Arguments) == 0
}

func (a *Analyzer) constantArrayIndex(expr ast.Expression) (int, bool) {
	if expr == nil {
		return 0, false
	}
	if value, err := a.evaluateConstant(expr); err == nil {
		switch v := value.(type) {
		case int:
			return v, true
		case bool:
			if v {
				return 1, true
			}
			return 0, true
		case string:
			if len(v) == 1 {
				return int([]rune(v)[0]), true
			}
		}
	}
	if value, err := a.evaluateConstantInt(expr); err == nil {
		return value, true
	}
	return 0, false
}

// analyzeIndexedPropertyAccess handles expressions like obj.Prop[index]
// by validating the index type against the property's signature and returning the property type.
func (a *Analyzer) analyzeIndexedPropertyAccess(memberAccess *ast.MemberAccessExpression, expr *ast.IndexExpression) types.Type {
	// Determine the object type for the member access
	objectType := a.analyzeExpression(memberAccess.Object)
	if objectType == nil {
		return nil
	}

	objectResolved := types.GetUnderlyingType(objectType)
	isMetaclass := false
	if metaclassType, ok := objectResolved.(*types.ClassOfType); ok {
		objectResolved = metaclassType.ClassType
		isMetaclass = true
	}

	memberName := ident.Normalize(memberAccess.Member.Value)

	// Handle class instance properties
	if classType, ok := objectResolved.(*types.ClassType); ok {
		if propInfo, found := classType.GetProperty(memberName); found {
			if !propInfo.IsIndexed {
				// Not an indexed property – let general indexing rules apply to the property type
				return nil
			}

			// Reaching an indexed property through a class name is legal only when the
			// accessor needs no instance. This mirrors the rule the plain member-access
			// path applies in analyzeClassMemberAccess; without it, semantic analysis
			// accepted what the evaluator could not execute.
			if isMetaclass && !a.checkIndexedPropertyMetaclassAccess(classType, propInfo, memberAccess) {
				return nil
			}

			expectedIndexTypes := a.getIndexedPropertyParamTypes(propInfo, classType)
			if len(expectedIndexTypes) > 0 {
				indexType := a.analyzeExpressionWithExpectedType(expr.Index, expectedIndexTypes[0])
				if indexType != nil && !a.canAssign(indexType, expectedIndexTypes[0]) {
					a.addStructuredError(NewArrayIndexError(expr.Index.Pos(), expectedIndexTypes[0].String(), indexType.String()))
					return propInfo.Type
				}
			} else {
				a.analyzeExpression(expr.Index)
			}
			return propInfo.Type
		}
	}

	// Not an indexed property access
	return nil
}

// analyzeMultiIndexPropertyAccess resolves an indexed property that takes more
// than one index. `obj.Prop[i, j]` parses as the chain `obj.Prop[i][j]`, so the
// indices must be collected back together and checked against the single
// property declaration at the root of the chain. Returns handled=false when the
// expression is not such a chain, leaving the ordinary indexing rules to apply.
func (a *Analyzer) analyzeMultiIndexPropertyAccess(expr *ast.IndexExpression) (types.Type, bool) {
	// Collect the chain outermost-first, then reverse into declaration order.
	indices := []ast.Expression{expr.Index}
	root := expr.Left
	for {
		inner, ok := root.(*ast.IndexExpression)
		if !ok {
			break
		}
		indices = append(indices, inner.Index)
		root = inner.Left
	}
	if len(indices) < 2 {
		return nil, false
	}
	for i, j := 0, len(indices)-1; i < j; i, j = i+1, j-1 {
		indices[i], indices[j] = indices[j], indices[i]
	}

	memberAccess, ok := root.(*ast.MemberAccessExpression)
	if !ok {
		return nil, false
	}
	objectType := a.analyzeExpression(memberAccess.Object)
	if objectType == nil {
		return nil, false
	}
	objectResolved := types.GetUnderlyingType(objectType)
	if metaclassType, ok := objectResolved.(*types.ClassOfType); ok {
		objectResolved = metaclassType.ClassType
	}
	classType, ok := objectResolved.(*types.ClassType)
	if !ok {
		return nil, false
	}
	propInfo, found := classType.GetProperty(ident.Normalize(memberAccess.Member.Value))
	if !found || !propInfo.IsIndexed {
		return nil, false
	}

	expectedIndexTypes := a.getIndexedPropertyParamTypes(propInfo, classType)
	// Only claim the chain when its length matches the declared arity; a shorter
	// or longer chain indexes into the property's own (array) result type.
	if len(expectedIndexTypes) != len(indices) {
		return nil, false
	}

	for i, indexExpr := range indices {
		indexType := a.analyzeExpressionWithExpectedType(indexExpr, expectedIndexTypes[i])
		if indexType != nil && !a.canAssign(indexType, expectedIndexTypes[i]) {
			a.addStructuredError(NewArrayIndexError(indexExpr.Pos(), expectedIndexTypes[i].String(), indexType.String()))
			return propInfo.Type, true
		}
	}
	return propInfo.Type, true
}

// getDefaultClassProperty walks the class hierarchy to find a default property, if any.
func (a *Analyzer) getDefaultClassProperty(classType *types.ClassType) *types.PropertyInfo {
	for current := classType; current != nil; current = current.Parent {
		for _, propInfo := range current.Properties {
			if propInfo.IsDefault {
				return propInfo
			}
		}
	}
	return nil
}

// getIndexedPropertyParamTypes tries to determine the index parameter types for an indexed property.
// Preference order:
//  1. Getter method parameters (all parameters are index parameters)
//  2. Setter method parameters (all but the last parameter are index parameters)
//
// If no method information is available, returns nil.
func (a *Analyzer) getIndexedPropertyParamTypes(propInfo *types.PropertyInfo, classType *types.ClassType) []types.Type {
	// Declared index parameters are authoritative and, unlike an accessor
	// method's signature, are also available for expression-based accessors.
	if len(propInfo.IndexParamTypes) > 0 {
		return propInfo.IndexParamTypes
	}

	// Use getter signature if it is a method
	if propInfo.ReadKind == types.PropAccessMethod && propInfo.ReadSpec != "" {
		if methodType, found := classType.GetMethod(ident.Normalize(propInfo.ReadSpec)); found {
			return methodType.Parameters
		}
	}

	// Use setter signature if it is a method (exclude the value parameter)
	if propInfo.WriteKind == types.PropAccessMethod && propInfo.WriteSpec != "" {
		if methodType, found := classType.GetMethod(ident.Normalize(propInfo.WriteSpec)); found {
			if len(methodType.Parameters) > 0 {
				return methodType.Parameters[:len(methodType.Parameters)-1]
			}
			return []types.Type{}
		}
	}

	return nil
}

// analyzeNewArrayExpression analyzes array instantiation with 'new' keyword
//
// Examples:
//   - new Integer[16]           // 1D array
//   - new String[10, 20]        // 2D array
//   - new Float[Length(arr)+1]  // Expression-based size
func (a *Analyzer) analyzeNewArrayExpression(expr *ast.NewArrayExpression) types.Type {
	if expr == nil {
		return nil
	}

	// Resolve the element type name
	elementTypeName := expr.ElementTypeName.Value
	elementType, err := a.resolveType(elementTypeName)
	if err != nil {
		a.addError("unknown type '%s' at %s", elementTypeName, expr.ElementTypeName.Pos().String())
		return nil
	}

	a.semanticInfo.SetResolvedType(expr.ElementTypeName, elementType)

	// Validate each dimension expression is an integer
	for i, dimExpr := range expr.Dimensions {
		dimType := a.analyzeExpression(dimExpr)
		if dimType == nil {
			// Error already reported by analyzeExpression
			continue
		}

		// Dimension must be integer
		if !dimType.Equals(types.INTEGER) {
			a.addStructuredError(NewArrayDimensionTypeError(dimExpr.Pos(), i+1, dimType.String()))
			return nil
		}
	}

	// Construct the result type (nested arrays for multi-dimensional)
	// For 1D: array of ElementType
	// For 2D: array of (array of ElementType)
	// For 3D: array of (array of (array of ElementType))
	resultType := elementType
	for range expr.Dimensions {
		resultType = types.NewDynamicArrayType(resultType)
	}

	return resultType
}

// checkIndexedPropertyMetaclassAccess validates reading an indexed instance property
// through a class name. Only a class-method accessor is reachable without an
// instance; a field-, expression- or instance-method-backed accessor is diagnosed
// with the same messages the non-indexed metaclass path uses. Class properties are
// always allowed. Returns false when a diagnostic was emitted.
func (a *Analyzer) checkIndexedPropertyMetaclassAccess(
	classType *types.ClassType,
	propInfo *types.PropertyInfo,
	memberAccess *ast.MemberAccessExpression,
) bool {
	if propInfo.IsClassProperty {
		return true
	}
	if a.indexedWriteTargetMember == ast.Expression(memberAccess) {
		// The target of a plain assignment is never read; checkIndexedPropertyWriteTarget
		// has already validated the setter.
		return true
	}
	pos := memberAccess.Member.Token.Pos
	switch propInfo.ReadKind {
	case types.PropAccessField, types.PropAccessMethod:
		// Resolve the class-method flag across the hierarchy: an inherited class
		// method is absent from the derived class's own ClassMethodFlags map, but
		// the evaluator's LookupClassMethod finds it.
		if propInfo.ReadSpec != "" && a.isClassMethodInHierarchy(classType, propInfo.ReadSpec) {
			return true
		}
		a.addStructuredError(NewPropertyReadShouldBeStaticMethodError(pos))
		a.addStructuredError(NewClassMethodOrConstructorExpectedError(pos))
		return false
	default:
		a.addStructuredError(NewObjectReferenceNeededError(pos))
		return false
	}
}

// indexedPropertyOfMemberAccess resolves the indexed property a member access names,
// or nil when the expression is not a member access on a class, or names no indexed
// property. It does not analyze the member access.
func (a *Analyzer) indexedPropertyOfMemberAccess(expr ast.Expression) *types.PropertyInfo {
	memberAccess, ok := expr.(*ast.MemberAccessExpression)
	if !ok {
		return nil
	}
	objectType := a.analyzeExpression(memberAccess.Object)
	if objectType == nil {
		return nil
	}
	objectResolved := types.GetUnderlyingType(objectType)
	if metaclassType, ok := objectResolved.(*types.ClassOfType); ok {
		objectResolved = metaclassType.ClassType
	}
	classType, ok := objectResolved.(*types.ClassType)
	if !ok {
		return nil
	}
	propInfo, found := classType.GetProperty(ident.Normalize(memberAccess.Member.Value))
	if !found || !propInfo.IsIndexed {
		return nil
	}
	return propInfo
}

// checkIndexedPropertyWriteTarget validates writing an indexed property, including
// through a class name, where only a class-method setter is reachable. It is the
// write counterpart of checkIndexedPropertyMetaclassAccess. Returns false when a
// diagnostic was emitted, so the caller can stop before the read-side analysis
// reports the same problem a second time.
func (a *Analyzer) checkIndexedPropertyWriteTarget(target ast.Expression, propInfo *types.PropertyInfo) bool {
	memberAccess, ok := target.(*ast.MemberAccessExpression)
	if !ok {
		return true
	}
	pos := memberAccess.Member.Token.Pos
	if propInfo.WriteKind == types.PropAccessNone {
		a.addStructuredError(NewReadOnlyPropertyError(pos, memberAccess.Member.Value))
		return false
	}

	objectType := a.analyzeExpression(memberAccess.Object)
	classOf, isMetaclass := types.GetUnderlyingType(objectType).(*types.ClassOfType)
	if !isMetaclass || propInfo.IsClassProperty || classOf.ClassType == nil {
		return true
	}
	switch propInfo.WriteKind {
	case types.PropAccessField, types.PropAccessMethod:
		if propInfo.WriteSpec != "" && a.isClassMethodInHierarchy(classOf.ClassType, propInfo.WriteSpec) {
			return true
		}
		a.addStructuredError(NewPropertyWriteShouldBeStaticMethodError(pos))
		a.addStructuredError(NewClassMethodOrConstructorExpectedError(pos))
	default:
		a.addStructuredError(NewObjectReferenceNeededError(pos))
	}
	return false
}
