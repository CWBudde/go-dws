package semantic

import (
	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Array Analysis
// ============================================================================

// analyzeArrayDecl analyzes an array type declaration
func (a *Analyzer) analyzeArrayDecl(decl *ast.ArrayDecl) {
	if decl == nil {
		return
	}

	arrayName := decl.Name.Value

	// Check if array type is already declared
	// Use lowercase for case-insensitive duplicate check
	// Task 6.1.1.3: Use TypeRegistry instead of scattered maps
	if a.hasType(arrayName) {
		a.addError("type '%s' already declared at %s", arrayName, decl.Token.Pos.String())
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
		// Evaluate bound expressions at semantic analysis time
		lowBound, err := a.evaluateConstantInt(arrayType.LowBound)
		if err != nil {
			a.addError("array lower bound must be a compile-time constant integer at %s: %v",
				decl.Token.Pos.String(), err)
			return
		}
		highBound, err := a.evaluateConstantInt(arrayType.HighBound)
		if err != nil {
			a.addError("array upper bound must be a compile-time constant integer at %s: %v",
				decl.Token.Pos.String(), err)
			return
		}

		// Validate bounds
		if lowBound > highBound {
			a.addError("array lower bound (%d) cannot be greater than upper bound (%d) at %s",
				lowBound, highBound, decl.Token.Pos.String())
			return
		}

		arrType = types.NewStaticArrayType(elementType, lowBound, highBound)
	}

	// Register the array type in the arrays registry
	// Use lowercase key for case-insensitive lookup
	// Task 6.1.1.3: Use TypeRegistry instead of scattered maps
	a.registerTypeWithPos(arrayName, arrType, decl.Token.Pos)
}

// analyzeIndexExpression analyzes an array/string indexing expression
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

	// Index must be integer or enum (enums are ordinal types in DWScript)
	if !indexType.Equals(types.INTEGER) && indexType.TypeKind() != "ENUM" {
		pos := expr.Index.Pos()
		a.addError("%s", errors.FormatArrayIndexError("Integer", indexType.String(), pos.Line, pos.Column))
		return nil
	}

	// Return the element type of the array
	return arrayType.ElementType
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
