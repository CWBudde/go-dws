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
	// Verify it's an array or string. A JSONVariant is accepted and, matching
	// DWScript, Length() returns the length of its string cast (e.g. "[]" → 2),
	// not the container element count (that is a.Length()/a.length).
	if argType != nil {
		if _, isArray := argType.(*types.ArrayType); !isArray {
			if argType != types.STRING && !types.IsJSONVariant(argType) {
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
		a.addMoreArgumentsExpected(callNamePos(callExpr.Function, callExpr.Token.Pos))
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
		a.addArgumentCountError(callNamePos(callExpr.Function, callExpr.Token.Pos), len(args), 2, 2)
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
