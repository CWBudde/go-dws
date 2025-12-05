package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Basic Math Built-in Function Analysis
// ============================================================================
// This file contains analyzers for basic mathematical operations including:
// - Abs, Min, Max
// - Clamp, ClampInt, MinInt, MaxInt
// - Sqr, Sign, Odd
// - DivMod

// analyzeAbs analyzes the Abs built-in function.
// Abs takes one numeric argument and returns the same type.
func (a *Analyzer) analyzeAbs(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Abs' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType != types.INTEGER && argType != types.FLOAT {
			a.addError("function 'Abs' expects numeric (Integer or Float) argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
			return types.INTEGER
		}
		return argType
	}
	return types.INTEGER
}

// analyzeMin analyzes the Min built-in function.
// Min takes two numeric arguments and returns the smaller value.
func (a *Analyzer) analyzeMin(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'Min' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	arg1Type := a.analyzeExpression(args[0])
	arg2Type := a.analyzeExpression(args[1])

	if arg1Type != nil && arg2Type != nil {
		if (arg1Type != types.INTEGER && arg1Type != types.FLOAT) ||
			(arg2Type != types.INTEGER && arg2Type != types.FLOAT) {
			a.addError("function 'Min' expects Integer or Float arguments, got %s and %s at %s",
				arg1Type.String(), arg2Type.String(), callExpr.Token.Pos.String())
			return types.INTEGER
		}
		if arg1Type == types.INTEGER && arg2Type == types.INTEGER {
			return types.INTEGER
		}
		return types.FLOAT
	}
	return types.INTEGER
}

// analyzeMax analyzes the Max built-in function.
// Max takes two numeric arguments and returns the larger value.
func (a *Analyzer) analyzeMax(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'Max' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	arg1Type := a.analyzeExpression(args[0])
	arg2Type := a.analyzeExpression(args[1])

	if arg1Type != nil && arg2Type != nil {
		if (arg1Type != types.INTEGER && arg1Type != types.FLOAT) ||
			(arg2Type != types.INTEGER && arg2Type != types.FLOAT) {
			a.addError("function 'Max' expects Integer or Float arguments, got %s and %s at %s",
				arg1Type.String(), arg2Type.String(), callExpr.Token.Pos.String())
			return types.INTEGER
		}
		if arg1Type == types.INTEGER && arg2Type == types.INTEGER {
			return types.INTEGER
		}
		return types.FLOAT
	}
	return types.INTEGER
}

// analyzeClampInt analyzes the ClampInt built-in function.
// ClampInt takes three integer arguments: value, min, max.
func (a *Analyzer) analyzeClampInt(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 3 {
		a.addError("function 'ClampInt' expects 3 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	arg1Type := a.analyzeExpression(args[0])
	arg2Type := a.analyzeExpression(args[1])
	arg3Type := a.analyzeExpression(args[2])

	if arg1Type != nil && arg2Type != nil && arg3Type != nil {
		if arg1Type != types.INTEGER || arg2Type != types.INTEGER || arg3Type != types.INTEGER {
			a.addError("function 'ClampInt' expects Integer arguments, got %s, %s, %s at %s",
				arg1Type.String(), arg2Type.String(), arg3Type.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeClamp analyzes the Clamp built-in function.
// Clamp takes three numeric arguments: value, min, max (returns Float).
func (a *Analyzer) analyzeClamp(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 3 {
		a.addError("function 'Clamp' expects 3 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	arg1Type := a.analyzeExpression(args[0])
	arg2Type := a.analyzeExpression(args[1])
	arg3Type := a.analyzeExpression(args[2])

	if arg1Type != nil && arg2Type != nil && arg3Type != nil {
		if (arg1Type != types.INTEGER && arg1Type != types.FLOAT) ||
			(arg2Type != types.INTEGER && arg2Type != types.FLOAT) ||
			(arg3Type != types.INTEGER && arg3Type != types.FLOAT) {
			a.addError("function 'Clamp' expects Integer or Float arguments, got %s, %s, %s at %s",
				arg1Type.String(), arg2Type.String(), arg3Type.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeMaxInt analyzes the MaxInt built-in function.
// MaxInt() with 0 args returns max integer constant.
// MaxInt(a, b, ...) with 2+ args returns maximum of all integers (variadic).
func (a *Analyzer) analyzeMaxInt(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) == 0 {
		return types.INTEGER
	}
	for i, arg := range args {
		argType := a.analyzeExpression(arg)
		if argType != nil && argType != types.INTEGER {
			a.addError("function 'MaxInt' expects Integer arguments, got %s at argument %d at %s",
				argType.String(), i+1, callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeMinInt analyzes the MinInt built-in function.
// MinInt() with 0 args returns min integer constant.
// MinInt(a, b, ...) with 2+ args returns minimum of all integers (variadic).
func (a *Analyzer) analyzeMinInt(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) == 0 {
		return types.INTEGER
	}
	for i, arg := range args {
		argType := a.analyzeExpression(arg)
		if argType != nil && argType != types.INTEGER {
			a.addError("function 'MinInt' expects Integer arguments, got %s at argument %d at %s",
				argType.String(), i+1, callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeSqr analyzes the Sqr built-in function.
// Sqr takes one numeric argument and returns x*x, preserving type.
func (a *Analyzer) analyzeSqr(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Sqr' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType != types.INTEGER && argType != types.FLOAT {
			a.addError("function 'Sqr' expects Integer or Float as argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
			return types.INTEGER
		}
		return argType
	}
	return types.INTEGER
}

// analyzeSign analyzes the Sign built-in function.
// Sign takes a Float or Integer and returns an Integer (-1, 0, or 1).
func (a *Analyzer) analyzeSign(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Sign' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.FLOAT && argType != types.INTEGER {
		a.addError("function 'Sign' expects Float or Integer, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeOdd analyzes the Odd built-in function.
// Odd takes an Integer and returns a Boolean.
func (a *Analyzer) analyzeOdd(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Odd' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.INTEGER {
		a.addError("function 'Odd' expects Integer, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.BOOLEAN
}

// analyzeDivMod analyzes the DivMod built-in procedure.
// DivMod(dividend, divisor: Integer; var quotient, remainder: Integer)
func (a *Analyzer) analyzeDivMod(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 4 {
		a.addError("function 'DivMod' expects 4 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return nil
	}

	dividendType := a.analyzeExpression(args[0])
	if dividendType != nil && dividendType != types.INTEGER {
		a.addError("function 'DivMod' expects Integer as first argument, got %s at %s",
			dividendType.String(), callExpr.Token.Pos.String())
	}

	divisorType := a.analyzeExpression(args[1])
	if divisorType != nil && divisorType != types.INTEGER {
		a.addError("function 'DivMod' expects Integer as second argument, got %s at %s",
			divisorType.String(), callExpr.Token.Pos.String())
	}

	quotientType := a.analyzeExpression(args[2])
	if quotientType != nil && quotientType != types.INTEGER {
		a.addError("function 'DivMod' expects Integer as third argument, got %s at %s",
			quotientType.String(), callExpr.Token.Pos.String())
	}

	remainderType := a.analyzeExpression(args[3])
	if remainderType != nil && remainderType != types.INTEGER {
		a.addError("function 'DivMod' expects Integer as fourth argument, got %s at %s",
			remainderType.String(), callExpr.Token.Pos.String())
	}

	return nil
}
