package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Literal Analysis
// ============================================================================
// This file consolidates analysis of literal expressions:
// - Array literals
// - Record literals
// - Set literals

// analyzeArrayLiteral analyzes an array literal expression.
// Type inference, element validation, numeric promotion.
func (a *Analyzer) analyzeArrayLiteral(lit *ast.ArrayLiteralExpression, expectedType types.Type) types.Type {
	if lit == nil {
		return nil
	}

	originalExpectedType := expectedType

	var expectedArrayType *types.ArrayType
	if expectedType != nil {
		if arr, ok := types.GetUnderlyingType(expectedType).(*types.ArrayType); ok {
			expectedArrayType = arr
		} else {
			a.addError("array literal cannot be assigned to non-array type %s at %s",
				expectedType.String(), lit.Token.Pos.String())
			return nil
		}
	}

	// Empty literal requires explicit context
	if len(lit.Elements) == 0 {
		if expectedArrayType == nil {
			a.addError("cannot infer type for empty array literal at %s", lit.Token.Pos.String())
			return nil
		}
		// Allow empty arrays for array of const / array of Variant (Format function)
		a.semanticInfo.SetType(lit, &ast.TypeAnnotation{
			Token: lit.Token,
			Name:  originalExpectedType.String(),
		})
		return originalExpectedType
	}

	var inferredElementType types.Type
	hasErrors := false

	for idx, elem := range lit.Elements {
		var elementExpected types.Type
		if expectedArrayType != nil {
			elementExpected = expectedArrayType.ElementType
		}

		elemType := a.analyzeExpressionWithExpectedType(elem, elementExpected)
		if elemType == nil {
			hasErrors = true
			continue
		}

		if expectedArrayType != nil {
			// This enables heterogeneous arrays like ['string', 123, 3.14] for Format()
			// Migrated from CONST to VARIANT for proper dynamic typing
			elemTypeUnderlying := types.GetUnderlyingType(expectedArrayType.ElementType)
			if elemTypeUnderlying.TypeKind() == "VARIANT" {
				// Accept any element type for array of Variant
				continue
			}

			if !a.canAssign(elemType, expectedArrayType.ElementType) {
				a.addError("array element %d has type %s, expected %s at %s",
					idx+1, elemType.String(), expectedArrayType.ElementType.String(), elem.Pos().String())
				hasErrors = true
			}
			continue
		}

		if inferredElementType == nil {
			inferredElementType = elemType
			continue
		}

		underlyingCurrent := types.GetUnderlyingType(elemType)
		underlyingInferred := types.GetUnderlyingType(inferredElementType)

		if underlyingInferred.Equals(underlyingCurrent) {
			continue
		}

		// If the current element fits in the inferred type, keep the inferred type.
		if a.canAssign(elemType, inferredElementType) {
			continue
		}

		// If we can widen the inferred type to the current element, do so.
		if a.canAssign(inferredElementType, elemType) {
			inferredElementType = elemType
			continue
		}

		// Attempt numeric promotion (e.g., Integer + Float -> Float)
		if promoted := types.PromoteTypes(underlyingInferred, underlyingCurrent); promoted != nil {
			inferredElementType = promoted
			continue
		}

		a.addError("incompatible element types in array literal: %s and %s at %s",
			underlyingInferred.String(), underlyingCurrent.String(), elem.Pos().String())
		hasErrors = true
	}

	if hasErrors {
		return nil
	}

	if expectedArrayType != nil {
		a.semanticInfo.SetType(lit, &ast.TypeAnnotation{
			Token: lit.Token,
			Name:  originalExpectedType.String(),
		})
		return originalExpectedType
	}

	if inferredElementType == nil {
		a.addError("unable to infer element type for array literal at %s", lit.Token.Pos.String())
		return nil
	}

	elementUnderlying := types.GetUnderlyingType(inferredElementType)
	size := len(lit.Elements)
	arrayType := types.NewStaticArrayType(elementUnderlying, 0, size-1)

	a.semanticInfo.SetType(lit, &ast.TypeAnnotation{
		Token: lit.Token,
		Name:  arrayType.String(),
	})

	return arrayType
}

// analyzeRecordLiteral analyzes a record literal expression.
func (a *Analyzer) analyzeRecordLiteral(lit *ast.RecordLiteralExpression, expectedType types.Type) types.Type {
	if lit == nil {
		return nil
	}

	var recordType *types.RecordType

	// Check if this is a typed record literal (has TypeName)
	if lit.TypeName != nil {
		// Typed record literal: TPoint(x: 10; y: 20)
		// Look up the type by name
		typeName := lit.TypeName.Value
		resolvedType, err := a.resolveType(typeName)
		if err != nil {
			a.addError("unknown record type '%s' in record literal", typeName)
			return nil
		}

		var ok bool
		recordType, ok = resolvedType.(*types.RecordType)
		if !ok {
			a.addError("'%s' is not a record type, got %s", typeName, resolvedType.String())
			return nil
		}

		// If expectedType is provided, verify it matches
		if expectedType != nil {
			if expectedRecordType, ok := expectedType.(*types.RecordType); ok {
				if expectedRecordType.Name != recordType.Name {
					a.addError("record literal type '%s' does not match expected type '%s'",
						recordType.Name, expectedRecordType.Name)
					return nil
				}
			}
		}
	} else {
		// Anonymous record literal: (x: 10; y: 20)
		// Requires expectedType from context
		if expectedType == nil {
			a.addError("anonymous record literal requires type context (use explicit type annotation or typed literal)")
			return nil
		}

		var ok bool
		recordType, ok = expectedType.(*types.RecordType)
		if !ok {
			a.addError("record literal requires a record type, got %s", expectedType.String())
			return nil
		}
	}

	// Track which fields have been initialized
	initializedFields := make(map[string]bool)

	// Validate each field in the literal
	for _, field := range lit.Fields {
		// Skip positional fields (not yet implemented)
		if field.Name == nil {
			a.addError("positional record field initialization not yet supported")
			continue
		}

		fieldName := field.Name.Value
		// Normalize field name to lowercase for case-insensitive comparison
		lowerFieldName := ident.Normalize(fieldName)

		// Check for duplicate field initialization
		if initializedFields[lowerFieldName] {
			a.addError("duplicate field '%s' in record literal", fieldName)
			continue
		}
		initializedFields[lowerFieldName] = true

		// Check if field exists in record type
		expectedFieldType, exists := recordType.Fields[lowerFieldName]
		if !exists {
			a.addError("field '%s' does not exist in record type '%s'", fieldName, recordType.Name)
			continue
		}

		// Type-check the field value
		actualType := a.analyzeExpression(field.Value)
		if actualType == nil {
			continue
		}

		// Check type compatibility
		if !a.canAssign(actualType, expectedFieldType) {
			a.addError("cannot assign %s to %s in field '%s'",
				actualType.String(), expectedFieldType.String(), fieldName)
		}
	}

	// Check for missing required fields (skip fields with default initializers)
	for fieldName := range recordType.Fields {
		if !initializedFields[fieldName] {
			// Check if the field has a default initializer
			if recordType.FieldsWithInit != nil && recordType.FieldsWithInit[fieldName] {
				// Field has a default initializer, so it's not required in the literal
				continue
			}
			a.addError("missing required field '%s' in record literal", fieldName)
		}
	}

	return recordType
}

// analyzeSetLiteralWithContext analyzes a set literal expression with optional type context
func (a *Analyzer) analyzeSetLiteralWithContext(lit *ast.SetLiteral, expectedType types.Type) types.Type {
	if lit == nil {
		return nil
	}

	// If we have an expected type, it should be a SetType
	var expectedSetType *types.SetType
	if expectedType != nil {
		var ok bool
		expectedSetType, ok = expectedType.(*types.SetType)
		if !ok {
			a.addError("set literal cannot be assigned to non-set type %s at %s",
				expectedType.String(), lit.Token.Pos.String())
			return nil
		}
	}

	// Empty set literal
	if len(lit.Elements) == 0 {
		if expectedSetType != nil {
			return expectedSetType
		}
		// Empty set without context - cannot infer type
		a.addError("cannot infer type for empty set literal at %s", lit.Token.Pos.String())
		return nil
	}

	// Analyze all elements and check they are of the same ordinal type
	// Support all ordinal types (Integer, String/Char, Enum, Subrange)
	var elementType types.Type
	for i, elem := range lit.Elements {
		var elemType types.Type

		// Check if this is a range expression (e.g., 1..10 or 'a'..'z')
		if rangeExpr, isRange := elem.(*ast.RangeExpression); isRange {
			// Analyze start and end of range
			startType := a.analyzeExpression(rangeExpr.Start)
			endType := a.analyzeExpression(rangeExpr.RangeEnd)

			if startType == nil || endType == nil {
				// Error already reported
				continue
			}

			// Both bounds must be ordinal types
			if !types.IsOrdinalType(startType) {
				a.addError("range start must be an ordinal type, got %s at %s",
					startType.String(), rangeExpr.Start.Pos().String())
				continue
			}
			if !types.IsOrdinalType(endType) {
				a.addError("range end must be an ordinal type, got %s at %s",
					endType.String(), rangeExpr.RangeEnd.Pos().String())
				continue
			}

			// Both bounds must be the same type
			if !startType.Equals(endType) {
				a.addError("range start and end must have the same type: got %s and %s at %s",
					startType.String(), endType.String(), rangeExpr.Pos().String())
				continue
			}

			elemType = startType
		} else {
			// Regular element (not a range)
			elemType = a.analyzeExpression(elem)
			if elemType == nil {
				// Error already reported
				continue
			}

			// Element must be an ordinal type
			if !types.IsOrdinalType(elemType) {
				a.addError("set element must be an ordinal value, got %s at %s",
					elemType.String(), elem.Pos().String())
				continue
			}
		}

		// First element determines the element type
		if i == 0 {
			elementType = elemType
		} else {
			// All elements must be of the same ordinal type
			if !elemType.Equals(elementType) {
				a.addError("type mismatch in set literal: expected %s, got %s at %s",
					elementType.String(), elemType.String(), elem.Pos().String())
			}
		}
	}

	if elementType == nil {
		// All elements had errors
		return nil
	}

	// If we have an expected set type, verify the element type matches
	if expectedSetType != nil {
		if !elementType.Equals(expectedSetType.ElementType) {
			a.addError("type mismatch in set literal: expected set of %s, got set of %s at %s",
				expectedSetType.ElementType.String(), elementType.String(), lit.Token.Pos.String())
			return expectedSetType // Return expected type to continue analysis
		}
		return expectedSetType
	}

	// Create and return a new set type based on inferred element type
	return types.NewSetType(elementType)
}
