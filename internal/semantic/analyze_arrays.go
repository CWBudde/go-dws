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
