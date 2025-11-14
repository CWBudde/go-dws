package semantic

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
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

// analyzeSubStr analyzes the SubStr built-in function.
// SubStr takes 2 or 3 arguments:
//   - SubStr(str, start) - returns substring from start to end (1-based)
//   - SubStr(str, start, length) - returns length characters starting at start (1-based)
//
// Note: Different from Copy which has an array overload, SubStr only works with strings.
func (a *Analyzer) analyzeSubStr(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) < 2 || len(args) > 3 {
		a.addError("function 'SubStr' expects 2 or 3 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}

	// Analyze the first argument (string)
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'SubStr' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}

	// Analyze the second argument (start position - integer)
	startType := a.analyzeExpression(args[1])
	if startType != nil && startType != types.INTEGER {
		a.addError("function 'SubStr' expects integer as second argument, got %s at %s",
			startType.String(), callExpr.Token.Pos.String())
	}

	// Analyze the third argument (length - integer) if present
	if len(args) == 3 {
		lengthType := a.analyzeExpression(args[2])
		if lengthType != nil && lengthType != types.INTEGER {
			a.addError("function 'SubStr' expects integer as third argument, got %s at %s",
				lengthType.String(), callExpr.Token.Pos.String())
		}
	}

	return types.STRING
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

// analyzeSubString analyzes the SubString built-in function.
// SubString(str, start, end) - returns substring from start to end (1-based, inclusive)
func (a *Analyzer) analyzeSubString(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 3 {
		a.addError("function 'SubString' expects 3 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze first argument (string)
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'SubString' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (start - integer)
	startType := a.analyzeExpression(args[1])
	if startType != nil && startType != types.INTEGER {
		a.addError("function 'SubString' expects integer as second argument, got %s at %s",
			startType.String(), callExpr.Token.Pos.String())
	}
	// Analyze third argument (end - integer)
	endType := a.analyzeExpression(args[2])
	if endType != nil && endType != types.INTEGER {
		a.addError("function 'SubString' expects integer as third argument, got %s at %s",
			endType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeLeftStr analyzes the LeftStr built-in function.
// LeftStr(str, count) - returns leftmost count characters
func (a *Analyzer) analyzeLeftStr(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'LeftStr' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze first argument (string)
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'LeftStr' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (count - integer)
	countType := a.analyzeExpression(args[1])
	if countType != nil && countType != types.INTEGER {
		a.addError("function 'LeftStr' expects integer as second argument, got %s at %s",
			countType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeRightStr analyzes the RightStr built-in function.
// RightStr(str, count) - returns rightmost count characters
func (a *Analyzer) analyzeRightStr(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'RightStr' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze first argument (string)
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'RightStr' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (count - integer)
	countType := a.analyzeExpression(args[1])
	if countType != nil && countType != types.INTEGER {
		a.addError("function 'RightStr' expects integer as second argument, got %s at %s",
			countType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeMidStr analyzes the MidStr built-in function.
// MidStr(str, start, count) - alias for SubStr
func (a *Analyzer) analyzeMidStr(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	// MidStr is an alias for SubStr, so use the same analysis
	return a.analyzeSubStr(args, callExpr)
}

// analyzeStrBefore analyzes the StrBefore built-in function.
// StrBefore(str, delimiter) - returns string
func (a *Analyzer) analyzeStrBefore(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'StrBefore' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze first argument (str - string)
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'StrBefore' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (delimiter - string)
	delimType := a.analyzeExpression(args[1])
	if delimType != nil && delimType != types.STRING {
		a.addError("function 'StrBefore' expects string as second argument, got %s at %s",
			delimType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeStrBeforeLast analyzes the StrBeforeLast built-in function.
// StrBeforeLast(str, delimiter) - returns string
func (a *Analyzer) analyzeStrBeforeLast(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'StrBeforeLast' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze first argument (str - string)
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'StrBeforeLast' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (delimiter - string)
	delimType := a.analyzeExpression(args[1])
	if delimType != nil && delimType != types.STRING {
		a.addError("function 'StrBeforeLast' expects string as second argument, got %s at %s",
			delimType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeStrAfter analyzes the StrAfter built-in function.
// StrAfter(str, delimiter) - returns string
func (a *Analyzer) analyzeStrAfter(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'StrAfter' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze first argument (str - string)
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'StrAfter' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (delimiter - string)
	delimType := a.analyzeExpression(args[1])
	if delimType != nil && delimType != types.STRING {
		a.addError("function 'StrAfter' expects string as second argument, got %s at %s",
			delimType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeStrAfterLast analyzes the StrAfterLast built-in function.
// StrAfterLast(str, delimiter) - returns string
func (a *Analyzer) analyzeStrAfterLast(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'StrAfterLast' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze first argument (str - string)
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'StrAfterLast' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (delimiter - string)
	delimType := a.analyzeExpression(args[1])
	if delimType != nil && delimType != types.STRING {
		a.addError("function 'StrAfterLast' expects string as second argument, got %s at %s",
			delimType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeStrBetween analyzes the StrBetween built-in function.
// StrBetween(str, start, stop) - returns string
func (a *Analyzer) analyzeStrBetween(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 3 {
		a.addError("function 'StrBetween' expects 3 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze first argument (str - string)
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'StrBetween' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (start - string)
	startType := a.analyzeExpression(args[1])
	if startType != nil && startType != types.STRING {
		a.addError("function 'StrBetween' expects string as second argument, got %s at %s",
			startType.String(), callExpr.Token.Pos.String())
	}
	// Analyze third argument (stop - string)
	stopType := a.analyzeExpression(args[2])
	if stopType != nil && stopType != types.STRING {
		a.addError("function 'StrBetween' expects string as third argument, got %s at %s",
			stopType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzePadLeft analyzes the PadLeft built-in function.
// PadLeft takes 2 or 3 arguments: PadLeft(str, count) or PadLeft(str, count, char)
func (a *Analyzer) analyzePadLeft(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) < 2 || len(args) > 3 {
		a.addError("function 'PadLeft' expects 2 or 3 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'PadLeft' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	countType := a.analyzeExpression(args[1])
	if countType != nil && countType != types.INTEGER {
		a.addError("function 'PadLeft' expects integer as second argument, got %s at %s",
			countType.String(), callExpr.Token.Pos.String())
	}
	if len(args) == 3 {
		charType := a.analyzeExpression(args[2])
		if charType != nil && charType != types.STRING {
			a.addError("function 'PadLeft' expects string as third argument, got %s at %s",
				charType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.STRING
}

// analyzePadRight analyzes the PadRight built-in function.
// PadRight takes 2 or 3 arguments: PadRight(str, count) or PadRight(str, count, char)
func (a *Analyzer) analyzePadRight(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) < 2 || len(args) > 3 {
		a.addError("function 'PadRight' expects 2 or 3 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'PadRight' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	countType := a.analyzeExpression(args[1])
	if countType != nil && countType != types.INTEGER {
		a.addError("function 'PadRight' expects integer as second argument, got %s at %s",
			countType.String(), callExpr.Token.Pos.String())
	}
	if len(args) == 3 {
		charType := a.analyzeExpression(args[2])
		if charType != nil && charType != types.STRING {
			a.addError("function 'PadRight' expects string as third argument, got %s at %s",
				charType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.STRING
}

// analyzeStrDeleteLeft analyzes the StrDeleteLeft/DeleteLeft built-in function.
// StrDeleteLeft takes 2 arguments: StrDeleteLeft(str, count)
func (a *Analyzer) analyzeStrDeleteLeft(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'StrDeleteLeft' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'StrDeleteLeft' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	countType := a.analyzeExpression(args[1])
	if countType != nil && countType != types.INTEGER {
		a.addError("function 'StrDeleteLeft' expects integer as second argument, got %s at %s",
			countType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeStrDeleteRight analyzes the StrDeleteRight/DeleteRight built-in function.
// StrDeleteRight takes 2 arguments: StrDeleteRight(str, count)
func (a *Analyzer) analyzeStrDeleteRight(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'StrDeleteRight' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'StrDeleteRight' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	countType := a.analyzeExpression(args[1])
	if countType != nil && countType != types.INTEGER {
		a.addError("function 'StrDeleteRight' expects integer as second argument, got %s at %s",
			countType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeReverseString analyzes the ReverseString built-in function.
// ReverseString takes 1 argument: ReverseString(str)
func (a *Analyzer) analyzeReverseString(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'ReverseString' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'ReverseString' expects string as argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeNormalizeString analyzes the NormalizeString/Normalize built-in function.
// NormalizeString takes 1 or 2 arguments: NormalizeString(str) or NormalizeString(str, form)
func (a *Analyzer) analyzeNormalizeString(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) < 1 || len(args) > 2 {
		a.addError("function 'NormalizeString' expects 1 or 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'NormalizeString' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	if len(args) == 2 {
		formType := a.analyzeExpression(args[1])
		if formType != nil && formType != types.STRING {
			a.addError("function 'NormalizeString' expects string as second argument, got %s at %s",
				formType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.STRING
}

// analyzeStripAccents analyzes the StripAccents built-in function.
// StripAccents takes 1 argument: StripAccents(str)
func (a *Analyzer) analyzeStripAccents(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'StripAccents' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'StripAccents' expects string as argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
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
	return types.STRING
}
