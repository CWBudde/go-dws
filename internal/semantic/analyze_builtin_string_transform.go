package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// String Built-in Function Analysis - Transform Functions
// ============================================================================
// Functions for transforming, modifying, and extracting strings

// analyzeCopy analyzes the Copy built-in function.
// Copy has multiple overloads:
//   - Copy(arr) - returns copy of array
//   - Copy(str, index) - returns substring from index to end
//   - Copy(str, index, count) - returns substring of length count starting at index
func (a *Analyzer) analyzeCopy(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) == 1 {
		// Copy(arr) - array copy overload
		arrType := a.analyzeExpression(args[0])
		if arrType != nil {
			if arrayType, ok := arrType.(*types.ArrayType); ok {
				// Return the same array type
				return arrayType
			}
			a.addError("function 'Copy' with 1 argument expects array, got %s at %s",
				arrType.String(), callExpr.Token.Pos.String())
		}
		// Return a generic array type as fallback
		return types.NewDynamicArrayType(types.INTEGER)
	}

	if len(args) == 2 {
		// Copy(str, index) - string copy from index to end
		strType := a.analyzeExpression(args[0])
		if strType != nil && strType != types.STRING {
			a.addError("function 'Copy' expects string as first argument, got %s at %s",
				strType.String(), callExpr.Token.Pos.String())
		}
		indexType := a.analyzeExpression(args[1])
		if indexType != nil && indexType != types.INTEGER {
			a.addError("function 'Copy' expects integer as second argument, got %s at %s",
				indexType.String(), callExpr.Token.Pos.String())
		}
		return types.STRING
	}

	if len(args) != 3 {
		a.addError("function 'Copy' expects 1, 2, or 3 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}

	// Copy(str, index, count) - string copy overload
	// Analyze the first argument (string)
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'Copy' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	// Analyze the second argument (index - integer)
	indexType := a.analyzeExpression(args[1])
	if indexType != nil && indexType != types.INTEGER {
		a.addError("function 'Copy' expects integer as second argument, got %s at %s",
			indexType.String(), callExpr.Token.Pos.String())
	}
	// Analyze the third argument (count - integer)
	countType := a.analyzeExpression(args[2])
	if countType != nil && countType != types.INTEGER {
		a.addError("function 'Copy' expects integer as third argument, got %s at %s",
			countType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeInsert analyzes the Insert built-in procedure.
// Insert takes 3 arguments: source string, target variable, position.
func (a *Analyzer) analyzeInsert(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 3 {
		a.addError("function 'Insert' expects 3 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.VOID
	}
	sourceType := a.analyzeExpression(args[0])
	if sourceType != nil && sourceType != types.STRING {
		a.addError("function 'Insert' first argument must be String, got %s at %s",
			sourceType.String(), callExpr.Token.Pos.String())
	}
	if _, ok := args[1].(*ast.Identifier); !ok {
		a.addError("function 'Insert' second argument must be a variable at %s",
			callExpr.Token.Pos.String())
	} else {
		targetType := a.analyzeExpression(args[1])
		if targetType != nil && targetType != types.STRING {
			a.addError("function 'Insert' second argument must be String, got %s at %s",
				targetType.String(), callExpr.Token.Pos.String())
		}
	}
	posType := a.analyzeExpression(args[2])
	if posType != nil && posType != types.INTEGER {
		a.addError("function 'Insert' third argument must be Integer, got %s at %s",
			posType.String(), callExpr.Token.Pos.String())
	}
	return types.VOID
}

// analyzeCharAt analyzes the CharAt built-in function.
// CharAt takes 2 arguments: CharAt(str, position)
func (a *Analyzer) analyzeCharAt(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'CharAt' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'CharAt' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	posType := a.analyzeExpression(args[1])
	if posType != nil && posType != types.INTEGER {
		a.addError("function 'CharAt' expects integer as second argument, got %s at %s",
			posType.String(), callExpr.Token.Pos.String())
	}
	// Emit deprecation warning to mirror DWScript behavior
	a.addWarning("\"CharAt\" has been deprecated [line: %d, column: %d]",
		callExpr.Token.Pos.Line, 9)
	return types.STRING
}
