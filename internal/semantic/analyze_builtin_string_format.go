package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// String Built-in Function Analysis - Format Functions
// ============================================================================
// Functions for formatting, building, and constructing strings

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

	// Validate that the first argument is at most a single character if it's a string literal
	// Empty string is allowed (DWScript treats it as space character)
	if stringLit, ok := args[0].(*ast.StringLiteral); ok {
		// Count runes to handle UTF-8 correctly
		runeCount := len([]rune(stringLit.Value))
		if runeCount > 1 {
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

// analyzeByteSizeToStr analyzes the ByteSizeToStr built-in function.
// ByteSizeToStr takes 1 argument: ByteSizeToStr(size)
func (a *Analyzer) analyzeByteSizeToStr(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'ByteSizeToStr' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	sizeType := a.analyzeExpression(args[0])
	if sizeType != nil && sizeType != types.INTEGER {
		a.addError("function 'ByteSizeToStr' expects integer as argument, got %s at %s",
			sizeType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeGetText analyzes the GetText/_() built-in function.
// GetText takes 1 argument: GetText(str)
func (a *Analyzer) analyzeGetText(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'GetText' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'GetText' expects string as argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}
