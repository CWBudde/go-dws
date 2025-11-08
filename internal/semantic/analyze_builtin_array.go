package semantic

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Array Built-in Function Analysis
// ============================================================================

// analyzeLow analyzes the Low built-in function.
// Low takes one argument (array, enum, or type meta-value) and returns a value of the appropriate type.
func (a *Analyzer) analyzeLow(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Low' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze the argument
	argType := a.analyzeExpression(args[0])
	// Verify it's an array, enum, or basic type (type meta-value)
	if argType != nil {
		if _, isArray := argType.(*types.ArrayType); isArray {
			// For arrays, return Integer
			return types.INTEGER
		}
		if enumType, isEnum := argType.(*types.EnumType); isEnum {
			// For enums, return the same enum type
			return enumType
		}
		// Handle type meta-values (Integer, Float, Boolean, String)
		switch argType {
		case types.INTEGER:
			return types.INTEGER
		case types.FLOAT:
			return types.FLOAT
		case types.BOOLEAN:
			return types.BOOLEAN
		case types.STRING:
			// String doesn't have a low value, but we allow it for consistency
			return types.INTEGER
		}
		// Neither array, enum, nor type meta-value
		a.addError("function 'Low' expects array, enum, or type name, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeHigh analyzes the High built-in function.
// High takes one argument (array, enum, or type meta-value) and returns a value of the appropriate type.
func (a *Analyzer) analyzeHigh(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'High' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze the argument
	argType := a.analyzeExpression(args[0])
	// Verify it's an array, enum, or basic type (type meta-value)
	if argType != nil {
		if _, isArray := argType.(*types.ArrayType); isArray {
			// For arrays, return Integer
			return types.INTEGER
		}
		if enumType, isEnum := argType.(*types.EnumType); isEnum {
			// For enums, return the same enum type
			return enumType
		}
		// Handle type meta-values (Integer, Float, Boolean, String)
		switch argType {
		case types.INTEGER:
			return types.INTEGER
		case types.FLOAT:
			return types.FLOAT
		case types.BOOLEAN:
			return types.BOOLEAN
		case types.STRING:
			// String doesn't have a high value, but we allow it for consistency
			return types.INTEGER
		}
		// Neither array, enum, nor type meta-value
		a.addError("function 'High' expects array, enum, or type name, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeSetLength analyzes the SetLength built-in procedure.
// SetLength takes two arguments (array, length) and returns void.
func (a *Analyzer) analyzeSetLength(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'SetLength' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.VOID
	}
	// Analyze the first argument (array)
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if _, isArray := argType.(*types.ArrayType); !isArray {
			a.addError("function 'SetLength' expects array as first argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	// Analyze the second argument (integer)
	lengthType := a.analyzeExpression(args[1])
	if lengthType != nil && lengthType != types.INTEGER {
		a.addError("function 'SetLength' expects integer as second argument, got %s at %s",
			lengthType.String(), callExpr.Token.Pos.String())
	}
	return types.VOID
}

// analyzeAdd analyzes the Add built-in procedure.
// Add takes two arguments (array, element) and returns void.
func (a *Analyzer) analyzeAdd(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'Add' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.VOID
	}
	// Analyze the first argument (array)
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if _, isArray := argType.(*types.ArrayType); !isArray {
			a.addError("function 'Add' expects array as first argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	// Analyze the second argument (element to add)
	a.analyzeExpression(args[1])
	return types.VOID
}

// analyzeDelete analyzes the Delete built-in procedure.
// Delete has two overloads:
//   - Delete(array, index) - for arrays (2 args)
//   - Delete(string, pos, count) - for strings (3 args)
func (a *Analyzer) analyzeDelete(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) == 2 {
		// Array delete: Delete(array, index)
		argType := a.analyzeExpression(args[0])
		if argType != nil {
			if _, isArray := argType.(*types.ArrayType); !isArray {
				a.addError("function 'Delete' expects array as first argument for 2-argument form, got %s at %s",
					argType.String(), callExpr.Token.Pos.String())
			}
		}
		indexType := a.analyzeExpression(args[1])
		if indexType != nil && indexType != types.INTEGER {
			a.addError("function 'Delete' expects integer as second argument, got %s at %s",
				indexType.String(), callExpr.Token.Pos.String())
		}
		return types.VOID
	} else if len(args) == 3 {
		// String delete: Delete(string, pos, count)
		if _, ok := args[0].(*ast.Identifier); !ok {
			a.addError("function 'Delete' first argument must be a variable at %s",
				callExpr.Token.Pos.String())
		} else {
			strType := a.analyzeExpression(args[0])
			if strType != nil && strType != types.STRING {
				a.addError("function 'Delete' first argument must be String for 3-argument form, got %s at %s",
					strType.String(), callExpr.Token.Pos.String())
			}
		}
		posType := a.analyzeExpression(args[1])
		if posType != nil && posType != types.INTEGER {
			a.addError("function 'Delete' second argument must be Integer, got %s at %s",
				posType.String(), callExpr.Token.Pos.String())
		}
		countType := a.analyzeExpression(args[2])
		if countType != nil && countType != types.INTEGER {
			a.addError("function 'Delete' third argument must be Integer, got %s at %s",
				countType.String(), callExpr.Token.Pos.String())
		}
		return types.VOID
	} else {
		a.addError("function 'Delete' expects 2 or 3 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.VOID
	}
}
