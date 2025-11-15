package semantic

import (
	"strings"

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
		if _, isArray := types.GetUnderlyingType(argType).(*types.ArrayType); isArray {
			// For arrays, return Integer
			return types.INTEGER
		}
		if enumType, isEnum := types.GetUnderlyingType(argType).(*types.EnumType); isEnum {
			// For enums, return the same enum type
			return enumType
		}
		if argType == types.STRING {
			return types.INTEGER
		}

		// Handle type meta-values (Integer, Float, Boolean, String)
		if a.isTypeMetaValueExpression(args[0]) {
			switch argType {
			case types.INTEGER:
				return types.INTEGER
			case types.FLOAT:
				return types.FLOAT
			case types.BOOLEAN:
				return types.BOOLEAN
			case types.STRING:
				return types.INTEGER
			}
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
		if _, isArray := types.GetUnderlyingType(argType).(*types.ArrayType); isArray {
			// For arrays, return Integer
			return types.INTEGER
		}
		if enumType, isEnum := types.GetUnderlyingType(argType).(*types.EnumType); isEnum {
			// For enums, return the same enum type
			return enumType
		}
		if argType == types.STRING {
			return types.INTEGER
		}
		// Handle type meta-values (Integer, Float, Boolean, String)
		if a.isTypeMetaValueExpression(args[0]) {
			switch argType {
			case types.INTEGER:
				return types.INTEGER
			case types.FLOAT:
				return types.FLOAT
			case types.BOOLEAN:
				return types.BOOLEAN
			case types.STRING:
				return types.INTEGER
			}
		}
		// Neither array, enum, nor type meta-value
		a.addError("function 'High' expects array, enum, or type name, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeSetLength analyzes the SetLength built-in procedure.
// SetLength takes two arguments (array or string, length) and returns void.
// DWScript supports SetLength on both dynamic arrays and strings.
func (a *Analyzer) analyzeSetLength(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'SetLength' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.VOID
	}
	// Analyze the first argument (array or string)
	argType := a.analyzeExpression(args[0])
	var arrayType *types.ArrayType
	if argType != nil {
		underlyingType := types.GetUnderlyingType(argType)

		// Check if it's an array type
		if at, isArray := underlyingType.(*types.ArrayType); isArray {
			arrayType = at
			if !arrayType.IsDynamic() {
				a.addError("function 'SetLength' expects dynamic array as first argument, got %s at %s",
					argType.String(), callExpr.Token.Pos.String())
			}
		} else if underlyingType != types.STRING {
			// Not an array and not a string - error
			a.addError("function 'SetLength' expects array or string as first argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
		// If it's a string, no additional validation needed (strings are implicitly dynamic)
	}
	// Analyze the second argument (integer)
	lengthType := a.analyzeExpression(args[1])
	if lengthType != nil && lengthType != types.INTEGER {
		a.addError("function 'SetLength' expects integer as second argument, got %s at %s",
			lengthType.String(), callExpr.Token.Pos.String())
	}

	// When array type is known, ensure we're operating on a dynamic array
	if arrayType != nil && !arrayType.IsDynamic() {
		a.addError("function 'SetLength' expects dynamic array as first argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
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
	var arrayType *types.ArrayType
	if argType != nil {
		if at, isArray := types.GetUnderlyingType(argType).(*types.ArrayType); isArray {
			arrayType = at
		} else {
			a.addError("function 'Add' expects array as first argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	// Analyze the second argument (element to add)
	var elementType types.Type
	if arrayType != nil {
		elementType = arrayType.ElementType
	}
	valueType := a.analyzeExpressionWithExpectedType(args[1], elementType)
	if arrayType != nil && valueType != nil && elementType != nil && !a.canAssign(elementType, valueType) {
		a.addError("function 'Add' expects element of type %s, got %s at %s",
			elementType.String(), valueType.String(), callExpr.Token.Pos.String())
	}
	return types.VOID
}

// analyzeDelete analyzes the Delete built-in procedure.
// Delete has multiple overloads:
//   - Delete(array, index) - deletes single element (2 args)
//   - Delete(array, index, count) - deletes count elements (3 args)
//   - Delete(string, pos, count) - deletes count characters (3 args)
func (a *Analyzer) analyzeDelete(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 3 {
		a.addError("function 'Delete' expects 3 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.VOID
	}

	// 3-argument form: Delete(array, index, count) or Delete(string, pos, count)
	if _, ok := args[0].(*ast.Identifier); !ok {
		a.addError("function 'Delete' first argument must be a variable at %s",
			callExpr.Token.Pos.String())
	} else {
		firstArgType := a.analyzeExpression(args[0])
		if firstArgType != nil {
			isArray := false
			isString := false
			if arrayType, ok := types.GetUnderlyingType(firstArgType).(*types.ArrayType); ok {
				isArray = true
				if !arrayType.IsDynamic() {
					a.addError("function 'Delete' expects dynamic array as first argument, got %s at %s",
						firstArgType.String(), callExpr.Token.Pos.String())
				}
			} else if firstArgType == types.STRING {
				isString = true
			}

			if !isArray && !isString {
				a.addError("function 'Delete' first argument must be String or array for 3-argument form, got %s at %s",
					firstArgType.String(), callExpr.Token.Pos.String())
			}
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
}

// isTypeMetaValueExpression checks if the provided expression refers to a type identifier rather than a value.
func (a *Analyzer) isTypeMetaValueExpression(expr ast.Expression) bool {
	ident, ok := expr.(*ast.Identifier)
	if !ok {
		return false
	}

	if _, exists := a.symbols.Resolve(ident.Value); exists {
		return false
	}

	lower := strings.ToLower(ident.Value)

	if _, ok := a.enums[lower]; ok {
		return true
	}
	if _, ok := a.classes[lower]; ok {
		return true
	}
	if _, ok := a.interfaces[lower]; ok {
		return true
	}
	if _, ok := a.typeAliases[lower]; ok {
		return true
	}
	if _, ok := a.records[lower]; ok {
		return true
	}
	if _, ok := a.sets[lower]; ok {
		return true
	}

	switch lower {
	case "integer", "float", "boolean", "string":
		return true
	}

	return false
}
