package semantic

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Array Analysis
// ============================================================================

// analyzeArrayDecl analyzes an array type declaration
// Task 8.126: Register array types and validate bounds
func (a *Analyzer) analyzeArrayDecl(decl *ast.ArrayDecl) {
	if decl == nil {
		return
	}

	arrayName := decl.Name.Value

	// Check if array type is already declared
	if _, exists := a.arrays[arrayName]; exists {
		a.addError("type '%s' already declared at %s", arrayName, decl.Token.Pos.String())
		return
	}

	// Validate array type
	arrayType := decl.ArrayType
	if arrayType == nil {
		a.addError("invalid array type declaration at %s", decl.Token.Pos.String())
		return
	}

	// Validate bounds if static array
	if arrayType.LowBound != nil && arrayType.HighBound != nil {
		if *arrayType.LowBound > *arrayType.HighBound {
			a.addError("invalid array bounds: lower bound (%d) > upper bound (%d) at %s",
				*arrayType.LowBound, *arrayType.HighBound, decl.Token.Pos.String())
			return
		}
	}

	// Resolve the element type using resolveType helper
	elementTypeName := arrayType.ElementType.Name
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
		arrType = types.NewStaticArrayType(elementType, *arrayType.LowBound, *arrayType.HighBound)
	}

	// Register the array type in the arrays registry
	a.arrays[arrayName] = arrType
}

// analyzeIndexExpression analyzes an array/string indexing expression
// Task 8.126: Type-check array indexing (index must be integer, result is element type)
func (a *Analyzer) analyzeIndexExpression(expr *ast.IndexExpression) types.Type {
	if expr == nil {
		return nil
	}

	// Analyze the left side (what's being indexed)
	leftType := a.analyzeExpression(expr.Left)
	if leftType == nil {
		// Error already reported
		return nil
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
				a.addError("string index must be integer, got %s at %s",
					indexType.String(), expr.Index.Pos().String())
				return nil
			}
			return types.STRING
		}

		a.addError("cannot index non-array type %s at %s",
			leftType.String(), expr.Token.Pos.String())
		return nil
	}

	// Analyze the index expression
	indexType := a.analyzeExpression(expr.Index)
	if indexType == nil {
		// Error already reported
		return nil
	}

	// Index must be integer
	if !indexType.Equals(types.INTEGER) {
		a.addError("array index must be integer, got %s at %s",
			indexType.String(), expr.Index.Pos().String())
		return nil
	}

	// Return the element type of the array
	return arrayType.ElementType
}

// analyzeNewArrayExpression analyzes array instantiation with 'new' keyword
// Task 9.162: Validate dimensions are integers and construct array type
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

	// Validate each dimension expression is an integer
	for i, dimExpr := range expr.Dimensions {
		dimType := a.analyzeExpression(dimExpr)
		if dimType == nil {
			// Error already reported by analyzeExpression
			continue
		}

		// Dimension must be integer
		if !dimType.Equals(types.INTEGER) {
			a.addError("array dimension %d must be integer, got %s at %s",
				i+1, dimType.String(), dimExpr.Pos().String())
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

// analyzeArrayLiteral analyzes an array literal expression.
// Task 9.186: Type inference, element validation, numeric promotion.
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
		// Task 9.156 & 9.225: Allow empty arrays for array of const / array of Variant (Format function)
		lit.SetType(&ast.TypeAnnotation{
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
			// Task 9.225 & 9.235: Allow any type when expected element type is Variant (array of const)
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
		lit.SetType(&ast.TypeAnnotation{
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
	arrayType := types.NewDynamicArrayType(elementUnderlying)

	lit.SetType(&ast.TypeAnnotation{
		Token: lit.Token,
		Name:  arrayType.String(),
	})

	return arrayType
}
