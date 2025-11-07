package semantic

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// String Built-in Function Analysis
// ============================================================================

// analyzeLength analyzes the Length built-in function.
// Length takes one argument (array or string) and returns an integer.
func (a *Analyzer) analyzeLength(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Length' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze the argument
	argType := a.analyzeExpression(args[0])
	// Verify it's an array or string
	if argType != nil {
		if _, isArray := argType.(*types.ArrayType); !isArray {
			if argType != types.STRING {
				a.addError("function 'Length' expects array or string, got %s at %s",
					argType.String(), callExpr.Token.Pos.String())
			}
		}
	}
	return types.INTEGER
}

// analyzeCopy analyzes the Copy built-in function.
// Copy has two overloads:
//   - Copy(arr) - returns copy of array
//   - Copy(str, index, count) - returns substring
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

	if len(args) != 3 {
		a.addError("function 'Copy' expects either 1 argument (array) or 3 arguments (string), got %d at %s",
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

// analyzeConcat analyzes the Concat built-in function.
// Concat takes at least one argument (all strings) and returns a string.
func (a *Analyzer) analyzeConcat(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) == 0 {
		a.addError("function 'Concat' expects at least 1 argument, got 0 at %s",
			callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze all arguments and verify they're strings
	for i, arg := range args {
		argType := a.analyzeExpression(arg)
		if argType != nil && argType != types.STRING {
			a.addError("function 'Concat' expects string as argument %d, got %s at %s",
				i+1, argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.STRING
}

// analyzePos analyzes the Pos built-in function.
// Pos takes two string arguments and returns an integer.
func (a *Analyzer) analyzePos(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'Pos' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze the first argument (substring)
	substrType := a.analyzeExpression(args[0])
	if substrType != nil && substrType != types.STRING {
		a.addError("function 'Pos' expects string as first argument, got %s at %s",
			substrType.String(), callExpr.Token.Pos.String())
	}
	// Analyze the second argument (string to search in)
	strType := a.analyzeExpression(args[1])
	if strType != nil && strType != types.STRING {
		a.addError("function 'Pos' expects string as second argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeUpperCase analyzes the UpperCase built-in function.
// UpperCase takes one string argument and returns a string.
func (a *Analyzer) analyzeUpperCase(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'UpperCase' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze the argument and verify it's a string
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.STRING {
		a.addError("function 'UpperCase' expects string as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeLowerCase analyzes the LowerCase built-in function.
// LowerCase takes one string argument and returns a string.
func (a *Analyzer) analyzeLowerCase(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'LowerCase' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze the argument and verify it's a string
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.STRING {
		a.addError("function 'LowerCase' expects string as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeTrim analyzes the Trim built-in function.
// Trim takes one string argument and returns a string.
func (a *Analyzer) analyzeTrim(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Trim' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze the argument and verify it's a string
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.STRING {
		a.addError("function 'Trim' expects string as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeTrimLeft analyzes the TrimLeft built-in function.
// TrimLeft takes one string argument and returns a string.
func (a *Analyzer) analyzeTrimLeft(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'TrimLeft' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze the argument and verify it's a string
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.STRING {
		a.addError("function 'TrimLeft' expects string as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeTrimRight analyzes the TrimRight built-in function.
// TrimRight takes one string argument and returns a string.
func (a *Analyzer) analyzeTrimRight(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'TrimRight' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze the argument and verify it's a string
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.STRING {
		a.addError("function 'TrimRight' expects string as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeStringReplace analyzes the StringReplace built-in function.
// StringReplace takes 3-4 arguments: str, old, new, [count].
func (a *Analyzer) analyzeStringReplace(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) < 3 || len(args) > 4 {
		a.addError("function 'StringReplace' expects 3 or 4 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// First argument: string to search in
	arg1Type := a.analyzeExpression(args[0])
	if arg1Type != nil && arg1Type != types.STRING {
		a.addError("function 'StringReplace' expects string as first argument, got %s at %s",
			arg1Type.String(), callExpr.Token.Pos.String())
	}
	// Second argument: old substring
	arg2Type := a.analyzeExpression(args[1])
	if arg2Type != nil && arg2Type != types.STRING {
		a.addError("function 'StringReplace' expects string as second argument, got %s at %s",
			arg2Type.String(), callExpr.Token.Pos.String())
	}
	// Third argument: new substring
	arg3Type := a.analyzeExpression(args[2])
	if arg3Type != nil && arg3Type != types.STRING {
		a.addError("function 'StringReplace' expects string as third argument, got %s at %s",
			arg3Type.String(), callExpr.Token.Pos.String())
	}
	// Optional fourth argument: count (integer)
	if len(args) == 4 {
		arg4Type := a.analyzeExpression(args[3])
		if arg4Type != nil && arg4Type != types.INTEGER {
			a.addError("function 'StringReplace' expects integer as fourth argument, got %s at %s",
				arg4Type.String(), callExpr.Token.Pos.String())
		}
	}
	return types.STRING
}

// analyzeStringOfChar analyzes the StringOfChar built-in function.
// StringOfChar takes exactly 2 arguments: character (string) and count (integer).
func (a *Analyzer) analyzeStringOfChar(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'StringOfChar' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// First argument: character (string)
	arg1Type := a.analyzeExpression(args[0])
	if arg1Type != nil && arg1Type != types.STRING {
		a.addError("function 'StringOfChar' expects string as first argument, got %s at %s",
			arg1Type.String(), callExpr.Token.Pos.String())
	}
	// Second argument: count (integer)
	arg2Type := a.analyzeExpression(args[1])
	if arg2Type != nil && arg2Type != types.INTEGER {
		a.addError("function 'StringOfChar' expects integer as second argument, got %s at %s",
			arg2Type.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeFormat analyzes the Format built-in function.
// Format takes exactly 2 arguments: format string and array of values.
func (a *Analyzer) analyzeFormat(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("Format() expects exactly 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// First argument: format string (must be String)
	fmtType := a.analyzeExpression(args[0])
	if fmtType != nil && fmtType != types.STRING {
		a.addError("Format() expects string as first argument, got %s at %s",
			fmtType.String(), callExpr.Token.Pos.String())
	}
	// Second argument: array of values (must be Array type)
	// Task 9.156 & 9.236: Pass ARRAY_OF_CONST (array of Variant) as expected type
	// This allows heterogeneous arrays like ['string', 123, 3.14]
	arrType := a.analyzeExpressionWithExpectedType(args[1], types.ARRAY_OF_CONST)
	if arrType != nil {
		if _, isArray := arrType.(*types.ArrayType); !isArray {
			a.addError("Format() expects array as second argument, got %s at %s",
				arrType.String(), callExpr.Token.Pos.String())
		}
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
