package semantic

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Type Conversion Built-in Function Analysis
// ============================================================================

// analyzeIntToStr analyzes the IntToStr built-in function.
// IntToStr takes one integer argument and returns a string.
func (a *Analyzer) analyzeIntToStr(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'IntToStr' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze the argument and verify it's Integer or a subrange of Integer
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.INTEGER {
		// Check if it's a subrange type with Integer base
		if subrange, ok := argType.(*types.SubrangeType); ok {
			if subrange.BaseType != types.INTEGER {
				a.addError("function 'IntToStr' expects Integer as argument, got %s at %s",
					argType.String(), callExpr.Token.Pos.String())
			}
		} else {
			a.addError("function 'IntToStr' expects Integer as argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.STRING
}

// analyzeIntToBin analyzes the IntToBin built-in function.
// IntToBin takes two integer arguments (value, digits) and returns a string.
func (a *Analyzer) analyzeIntToBin(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'IntToBin' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze first argument (value) - must be Integer or subrange of Integer
	argType1 := a.analyzeExpression(args[0])
	if argType1 != nil && argType1 != types.INTEGER {
		// Check if it's a subrange type with Integer base
		if subrange, ok := argType1.(*types.SubrangeType); ok {
			if subrange.BaseType != types.INTEGER {
				a.addError("function 'IntToBin' expects Integer as first argument, got %s at %s",
					argType1.String(), callExpr.Token.Pos.String())
			}
		} else {
			a.addError("function 'IntToBin' expects Integer as first argument, got %s at %s",
				argType1.String(), callExpr.Token.Pos.String())
		}
	}
	// Analyze second argument (digits) - must be Integer
	argType2 := a.analyzeExpression(args[1])
	if argType2 != nil && argType2 != types.INTEGER {
		// Check if it's a subrange type with Integer base
		if subrange, ok := argType2.(*types.SubrangeType); ok {
			if subrange.BaseType != types.INTEGER {
				a.addError("function 'IntToBin' expects Integer as second argument, got %s at %s",
					argType2.String(), callExpr.Token.Pos.String())
			}
		} else {
			a.addError("function 'IntToBin' expects Integer as second argument, got %s at %s",
				argType2.String(), callExpr.Token.Pos.String())
		}
	}
	return types.STRING
}

// analyzeStrToInt analyzes the StrToInt built-in function.
// StrToInt takes one string argument and returns an integer.
func (a *Analyzer) analyzeStrToInt(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'StrToInt' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze the argument and verify it's String
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.STRING {
		a.addError("function 'StrToInt' expects String as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeBoolToStr analyzes the BoolToStr built-in function.
// BoolToStr takes one boolean argument and returns a string.
func (a *Analyzer) analyzeBoolToStr(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'BoolToStr' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze the argument and verify it's Boolean
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.BOOLEAN {
		a.addError("function 'BoolToStr' expects Boolean as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeStrToFloat analyzes the StrToFloat built-in function.
// StrToFloat takes one string argument and returns a float.
func (a *Analyzer) analyzeStrToFloat(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'StrToFloat' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	// Analyze the argument and verify it's String
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.STRING {
		a.addError("function 'StrToFloat' expects String as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.FLOAT
}

// analyzeVarToStr analyzes the VarToStr built-in function.
// VarToStr takes one variant argument and returns a string.
func (a *Analyzer) analyzeVarToStr(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'VarToStr' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze the argument (accepts any type - variant)
	a.analyzeExpression(args[0])
	return types.STRING
}

// analyzeFloatToStr analyzes the FloatToStr built-in function.
// FloatToStr takes one float argument and returns a string.
func (a *Analyzer) analyzeFloatToStr(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'FloatToStr' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze the argument and verify it's Float
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.FLOAT {
		a.addError("function 'FloatToStr' expects Float as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeFloatToStrF analyzes the FloatToStrF built-in function.
// FloatToStrF takes three arguments (value, format, precision, digits) and returns a string.
func (a *Analyzer) analyzeFloatToStrF(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 4 {
		a.addError("function 'FloatToStrF' expects 4 arguments (value, format, precision, digits), got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// First argument: Float value
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'FloatToStrF' expects Float as first argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	// Second argument: format (Integer)
	if len(args) > 1 {
		argType := a.analyzeExpression(args[1])
		if argType != nil && argType != types.INTEGER {
			a.addError("function 'FloatToStrF' expects Integer as second argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	// Third argument: precision (Integer)
	if len(args) > 2 {
		argType := a.analyzeExpression(args[2])
		if argType != nil && argType != types.INTEGER {
			a.addError("function 'FloatToStrF' expects Integer as third argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	// Fourth argument: digits (Integer)
	if len(args) > 3 {
		argType := a.analyzeExpression(args[3])
		if argType != nil && argType != types.INTEGER {
			a.addError("function 'FloatToStrF' expects Integer as fourth argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.STRING
}

// analyzeStrToBool analyzes the StrToBool built-in function.
// StrToBool takes one string argument and returns a boolean.
func (a *Analyzer) analyzeStrToBool(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'StrToBool' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}
	// Analyze the argument and verify it's String
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.STRING {
		a.addError("function 'StrToBool' expects String as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.BOOLEAN
}

// analyzeChr analyzes the Chr built-in function.
// Chr takes one integer argument (character code) and returns a string.
func (a *Analyzer) analyzeChr(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Chr' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze the argument and verify it's Integer
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.INTEGER {
		a.addError("function 'Chr' expects Integer as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}
