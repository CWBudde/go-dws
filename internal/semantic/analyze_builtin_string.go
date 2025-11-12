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

// analyzeConcat analyzes the Concat built-in function.
// Concat takes at least one argument (all strings or all arrays) and returns a string or array.
func (a *Analyzer) analyzeConcat(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) == 0 {
		a.addError("function 'Concat' expects at least 1 argument, got 0 at %s",
			callExpr.Token.Pos.String())
		return types.STRING
	}

	// Check the type of the first argument to determine if we're concatenating strings or arrays
	firstArgType := a.analyzeExpression(args[0])

	// If first argument is an array, all arguments must be arrays
	if arrType, ok := firstArgType.(*types.ArrayType); ok {
		for i := 1; i < len(args); i++ {
			argType := a.analyzeExpression(args[i])
			if _, ok := argType.(*types.ArrayType); !ok {
				a.addError("function 'Concat' expects array as argument %d, got %s at %s",
					i+1, argType.String(), callExpr.Token.Pos.String())
			}
		}
		return arrType
	}

	// Otherwise, all arguments must be strings
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

	// Validate that the first argument is a single character if it's a string literal
	if stringLit, ok := args[0].(*ast.StringLiteral); ok {
		// Count runes to handle UTF-8 correctly
		runeCount := len([]rune(stringLit.Value))
		if runeCount != 1 {
			a.addError("function 'StringOfChar' expects a single char as first argument, got string of length %d at %s",
				runeCount, callExpr.Token.Pos.String())
		}
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
	// Pass ARRAY_OF_CONST (array of Variant) as expected type
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

// analyzeStrBeginsWith analyzes the StrBeginsWith built-in function.
// StrBeginsWith(str, prefix) - returns boolean
func (a *Analyzer) analyzeStrBeginsWith(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'StrBeginsWith' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}
	// Analyze first argument (string)
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'StrBeginsWith' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (prefix - string)
	prefixType := a.analyzeExpression(args[1])
	if prefixType != nil && prefixType != types.STRING {
		a.addError("function 'StrBeginsWith' expects string as second argument, got %s at %s",
			prefixType.String(), callExpr.Token.Pos.String())
	}
	return types.BOOLEAN
}

// analyzeStrEndsWith analyzes the StrEndsWith built-in function.
// StrEndsWith(str, suffix) - returns boolean
func (a *Analyzer) analyzeStrEndsWith(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'StrEndsWith' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}
	// Analyze first argument (string)
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'StrEndsWith' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (suffix - string)
	suffixType := a.analyzeExpression(args[1])
	if suffixType != nil && suffixType != types.STRING {
		a.addError("function 'StrEndsWith' expects string as second argument, got %s at %s",
			suffixType.String(), callExpr.Token.Pos.String())
	}
	return types.BOOLEAN
}

// analyzeStrContains analyzes the StrContains built-in function.
// StrContains(str, substring) - returns boolean
func (a *Analyzer) analyzeStrContains(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'StrContains' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}
	// Analyze first argument (string)
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'StrContains' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (substring - string)
	substrType := a.analyzeExpression(args[1])
	if substrType != nil && substrType != types.STRING {
		a.addError("function 'StrContains' expects string as second argument, got %s at %s",
			substrType.String(), callExpr.Token.Pos.String())
	}
	return types.BOOLEAN
}

// analyzePosEx analyzes the PosEx built-in function.
// PosEx(needle, haystack, offset) - returns integer (1-based position, 0 if not found)
func (a *Analyzer) analyzePosEx(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 3 {
		a.addError("function 'PosEx' expects 3 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze first argument (needle - string)
	needleType := a.analyzeExpression(args[0])
	if needleType != nil && needleType != types.STRING {
		a.addError("function 'PosEx' expects string as first argument, got %s at %s",
			needleType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (haystack - string)
	haystackType := a.analyzeExpression(args[1])
	if haystackType != nil && haystackType != types.STRING {
		a.addError("function 'PosEx' expects string as second argument, got %s at %s",
			haystackType.String(), callExpr.Token.Pos.String())
	}
	// Analyze third argument (offset - integer)
	offsetType := a.analyzeExpression(args[2])
	if offsetType != nil && offsetType != types.INTEGER {
		a.addError("function 'PosEx' expects integer as third argument, got %s at %s",
			offsetType.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeRevPos analyzes the RevPos built-in function.
// RevPos(needle, haystack) - returns integer (1-based position of last occurrence, 0 if not found)
func (a *Analyzer) analyzeRevPos(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'RevPos' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze first argument (needle - string)
	needleType := a.analyzeExpression(args[0])
	if needleType != nil && needleType != types.STRING {
		a.addError("function 'RevPos' expects string as first argument, got %s at %s",
			needleType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (haystack - string)
	haystackType := a.analyzeExpression(args[1])
	if haystackType != nil && haystackType != types.STRING {
		a.addError("function 'RevPos' expects string as second argument, got %s at %s",
			haystackType.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeStrFind analyzes the StrFind built-in function.
// StrFind(str, substr, fromIndex) - returns integer (1-based position, 0 if not found)
func (a *Analyzer) analyzeStrFind(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 3 {
		a.addError("function 'StrFind' expects 3 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze first argument (str - string)
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'StrFind' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (substr - string)
	substrType := a.analyzeExpression(args[1])
	if substrType != nil && substrType != types.STRING {
		a.addError("function 'StrFind' expects string as second argument, got %s at %s",
			substrType.String(), callExpr.Token.Pos.String())
	}
	// Analyze third argument (fromIndex - integer)
	fromIndexType := a.analyzeExpression(args[2])
	if fromIndexType != nil && fromIndexType != types.INTEGER {
		a.addError("function 'StrFind' expects integer as third argument, got %s at %s",
			fromIndexType.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeStrSplit analyzes the StrSplit built-in function.
// StrSplit(str, delimiter) - returns array of string
func (a *Analyzer) analyzeStrSplit(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'StrSplit' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.NewDynamicArrayType(types.STRING)
	}
	// Analyze first argument (str - string)
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'StrSplit' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (delimiter - string)
	delimType := a.analyzeExpression(args[1])
	if delimType != nil && delimType != types.STRING {
		a.addError("function 'StrSplit' expects string as second argument, got %s at %s",
			delimType.String(), callExpr.Token.Pos.String())
	}
	return types.NewDynamicArrayType(types.STRING)
}

// analyzeStrJoin analyzes the StrJoin built-in function.
// StrJoin(array, delimiter) - returns string
func (a *Analyzer) analyzeStrJoin(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'StrJoin' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze first argument (array of strings)
	arrType := a.analyzeExpression(args[0])
	if arrType != nil {
		arrayType, ok := arrType.(*types.ArrayType)
		if !ok {
			a.addError("function 'StrJoin' expects array as first argument, got %s at %s",
				arrType.String(), callExpr.Token.Pos.String())
		} else if arrayType.ElementType != types.STRING {
			a.addError("function 'StrJoin' expects array of string, got array of %s at %s",
				arrayType.ElementType.String(), callExpr.Token.Pos.String())
		}
	}
	// Analyze second argument (delimiter - string)
	delimType := a.analyzeExpression(args[1])
	if delimType != nil && delimType != types.STRING {
		a.addError("function 'StrJoin' expects string as second argument, got %s at %s",
			delimType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeStrArrayPack analyzes the StrArrayPack built-in function.
// StrArrayPack(array) - returns array of string
func (a *Analyzer) analyzeStrArrayPack(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'StrArrayPack' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.NewDynamicArrayType(types.STRING)
	}
	// Analyze first argument (array of strings)
	arrType := a.analyzeExpression(args[0])
	if arrType != nil {
		arrayType, ok := arrType.(*types.ArrayType)
		if !ok {
			a.addError("function 'StrArrayPack' expects array as argument, got %s at %s",
				arrType.String(), callExpr.Token.Pos.String())
		} else if arrayType.ElementType != types.STRING {
			a.addError("function 'StrArrayPack' expects array of string, got array of %s at %s",
				arrayType.ElementType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.NewDynamicArrayType(types.STRING)
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

// analyzeIsDelimiter analyzes the IsDelimiter built-in function.
// IsDelimiter(delims, str, index) - returns boolean
func (a *Analyzer) analyzeIsDelimiter(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 3 {
		a.addError("function 'IsDelimiter' expects 3 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}
	// Analyze first argument (delims - string)
	delimsType := a.analyzeExpression(args[0])
	if delimsType != nil && delimsType != types.STRING {
		a.addError("function 'IsDelimiter' expects string as first argument, got %s at %s",
			delimsType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (str - string)
	strType := a.analyzeExpression(args[1])
	if strType != nil && strType != types.STRING {
		a.addError("function 'IsDelimiter' expects string as second argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	// Analyze third argument (index - integer)
	indexType := a.analyzeExpression(args[2])
	if indexType != nil && indexType != types.INTEGER {
		a.addError("function 'IsDelimiter' expects integer as third argument, got %s at %s",
			indexType.String(), callExpr.Token.Pos.String())
	}
	return types.BOOLEAN
}

// analyzeLastDelimiter analyzes the LastDelimiter built-in function.
// LastDelimiter(delims, str) - returns integer
func (a *Analyzer) analyzeLastDelimiter(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'LastDelimiter' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze first argument (delims - string)
	delimsType := a.analyzeExpression(args[0])
	if delimsType != nil && delimsType != types.STRING {
		a.addError("function 'LastDelimiter' expects string as first argument, got %s at %s",
			delimsType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (str - string)
	strType := a.analyzeExpression(args[1])
	if strType != nil && strType != types.STRING {
		a.addError("function 'LastDelimiter' expects string as second argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeFindDelimiter analyzes the FindDelimiter built-in function.
// FindDelimiter(delims, str, startIndex) - returns integer
func (a *Analyzer) analyzeFindDelimiter(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 3 {
		a.addError("function 'FindDelimiter' expects 3 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze first argument (delims - string)
	delimsType := a.analyzeExpression(args[0])
	if delimsType != nil && delimsType != types.STRING {
		a.addError("function 'FindDelimiter' expects string as first argument, got %s at %s",
			delimsType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (str - string)
	strType := a.analyzeExpression(args[1])
	if strType != nil && strType != types.STRING {
		a.addError("function 'FindDelimiter' expects string as second argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	// Analyze third argument (startIndex - integer)
	startIndexType := a.analyzeExpression(args[2])
	if startIndexType != nil && startIndexType != types.INTEGER {
		a.addError("function 'FindDelimiter' expects integer as third argument, got %s at %s",
			startIndexType.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
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

// analyzeQuotedStr analyzes the QuotedStr built-in function.
// QuotedStr takes 1 or 2 arguments: QuotedStr(str) or QuotedStr(str, quoteChar)
func (a *Analyzer) analyzeQuotedStr(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) < 1 || len(args) > 2 {
		a.addError("function 'QuotedStr' expects 1 or 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'QuotedStr' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	if len(args) == 2 {
		quoteType := a.analyzeExpression(args[1])
		if quoteType != nil && quoteType != types.STRING {
			a.addError("function 'QuotedStr' expects string as second argument, got %s at %s",
				quoteType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.STRING
}

// analyzeStringOfString analyzes the StringOfString/DupeString built-in function.
// StringOfString takes 2 arguments: StringOfString(str, count)
func (a *Analyzer) analyzeStringOfString(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'StringOfString' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'StringOfString' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	countType := a.analyzeExpression(args[1])
	if countType != nil && countType != types.INTEGER {
		a.addError("function 'StringOfString' expects integer as second argument, got %s at %s",
			countType.String(), callExpr.Token.Pos.String())
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

// analyzeSameText analyzes the SameText built-in function.
// SameText takes 2 arguments: SameText(str1, str2)
func (a *Analyzer) analyzeSameText(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'SameText' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}
	str1Type := a.analyzeExpression(args[0])
	if str1Type != nil && str1Type != types.STRING {
		a.addError("function 'SameText' expects string as first argument, got %s at %s",
			str1Type.String(), callExpr.Token.Pos.String())
	}
	str2Type := a.analyzeExpression(args[1])
	if str2Type != nil && str2Type != types.STRING {
		a.addError("function 'SameText' expects string as second argument, got %s at %s",
			str2Type.String(), callExpr.Token.Pos.String())
	}
	return types.BOOLEAN
}

// analyzeCompareText analyzes the CompareText built-in function.
// CompareText takes 2 arguments: CompareText(str1, str2)
func (a *Analyzer) analyzeCompareText(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'CompareText' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	str1Type := a.analyzeExpression(args[0])
	if str1Type != nil && str1Type != types.STRING {
		a.addError("function 'CompareText' expects string as first argument, got %s at %s",
			str1Type.String(), callExpr.Token.Pos.String())
	}
	str2Type := a.analyzeExpression(args[1])
	if str2Type != nil && str2Type != types.STRING {
		a.addError("function 'CompareText' expects string as second argument, got %s at %s",
			str2Type.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeCompareStr analyzes the CompareStr built-in function.
// CompareStr takes 2 arguments: CompareStr(str1, str2)
func (a *Analyzer) analyzeCompareStr(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'CompareStr' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	str1Type := a.analyzeExpression(args[0])
	if str1Type != nil && str1Type != types.STRING {
		a.addError("function 'CompareStr' expects string as first argument, got %s at %s",
			str1Type.String(), callExpr.Token.Pos.String())
	}
	str2Type := a.analyzeExpression(args[1])
	if str2Type != nil && str2Type != types.STRING {
		a.addError("function 'CompareStr' expects string as second argument, got %s at %s",
			str2Type.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeAnsiCompareText analyzes the AnsiCompareText built-in function.
// AnsiCompareText takes 2 arguments: AnsiCompareText(str1, str2)
func (a *Analyzer) analyzeAnsiCompareText(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'AnsiCompareText' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	str1Type := a.analyzeExpression(args[0])
	if str1Type != nil && str1Type != types.STRING {
		a.addError("function 'AnsiCompareText' expects string as first argument, got %s at %s",
			str1Type.String(), callExpr.Token.Pos.String())
	}
	str2Type := a.analyzeExpression(args[1])
	if str2Type != nil && str2Type != types.STRING {
		a.addError("function 'AnsiCompareText' expects string as second argument, got %s at %s",
			str2Type.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeAnsiCompareStr analyzes the AnsiCompareStr built-in function.
// AnsiCompareStr takes 2 arguments: AnsiCompareStr(str1, str2)
func (a *Analyzer) analyzeAnsiCompareStr(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'AnsiCompareStr' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	str1Type := a.analyzeExpression(args[0])
	if str1Type != nil && str1Type != types.STRING {
		a.addError("function 'AnsiCompareStr' expects string as first argument, got %s at %s",
			str1Type.String(), callExpr.Token.Pos.String())
	}
	str2Type := a.analyzeExpression(args[1])
	if str2Type != nil && str2Type != types.STRING {
		a.addError("function 'AnsiCompareStr' expects string as second argument, got %s at %s",
			str2Type.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeCompareLocaleStr analyzes the CompareLocaleStr built-in function.
// CompareLocaleStr takes 2 to 4 arguments: CompareLocaleStr(str1, str2 [, locale [, caseSensitive]])
func (a *Analyzer) analyzeCompareLocaleStr(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) < 2 || len(args) > 4 {
		a.addError("function 'CompareLocaleStr' expects 2 to 4 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	str1Type := a.analyzeExpression(args[0])
	if str1Type != nil && str1Type != types.STRING {
		a.addError("function 'CompareLocaleStr' expects string as first argument, got %s at %s",
			str1Type.String(), callExpr.Token.Pos.String())
	}
	str2Type := a.analyzeExpression(args[1])
	if str2Type != nil && str2Type != types.STRING {
		a.addError("function 'CompareLocaleStr' expects string as second argument, got %s at %s",
			str2Type.String(), callExpr.Token.Pos.String())
	}
	if len(args) >= 3 {
		localeType := a.analyzeExpression(args[2])
		if localeType != nil && localeType != types.STRING {
			a.addError("function 'CompareLocaleStr' expects string as third argument, got %s at %s",
				localeType.String(), callExpr.Token.Pos.String())
		}
	}
	if len(args) == 4 {
		csType := a.analyzeExpression(args[3])
		if csType != nil && csType != types.BOOLEAN {
			a.addError("function 'CompareLocaleStr' expects boolean as fourth argument, got %s at %s",
				csType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeStrMatches analyzes the StrMatches built-in function.
// StrMatches takes 2 arguments: StrMatches(str, mask)
func (a *Analyzer) analyzeStrMatches(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'StrMatches' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'StrMatches' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	maskType := a.analyzeExpression(args[1])
	if maskType != nil && maskType != types.STRING {
		a.addError("function 'StrMatches' expects string as second argument, got %s at %s",
			maskType.String(), callExpr.Token.Pos.String())
	}
	return types.BOOLEAN
}

// analyzeStrIsASCII analyzes the StrIsASCII built-in function.
// StrIsASCII takes 1 argument: StrIsASCII(str)
func (a *Analyzer) analyzeStrIsASCII(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'StrIsASCII' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'StrIsASCII' expects string as argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	return types.BOOLEAN
}
