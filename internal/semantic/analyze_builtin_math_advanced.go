package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Advanced Math Built-in Function Analysis
// ============================================================================
// This file contains analyzers for advanced mathematical functions:
// - Power, IntPower, Sqrt
// - Exp, Ln, Log2, Log10, LogN
// - Pi, Infinity, NaN
// - IsNaN, IsFinite, IsInfinite

// analyzePower analyzes the Power built-in function.
// Power takes two numeric arguments and always returns Float.
func (a *Analyzer) analyzePower(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'Power' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	arg1Type := a.analyzeExpression(args[0])
	arg2Type := a.analyzeExpression(args[1])

	if arg1Type != nil && arg2Type != nil {
		if (arg1Type != types.INTEGER && arg1Type != types.FLOAT) ||
			(arg2Type != types.INTEGER && arg2Type != types.FLOAT) {
			a.addError("function 'Power' expects Integer or Float arguments, got %s and %s at %s",
				arg1Type.String(), arg2Type.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeIntPower analyzes the IntPower built-in function.
// IntPower takes base (Float or Integer) and exponent (Integer) and returns Float.
func (a *Analyzer) analyzeIntPower(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'IntPower' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}

	argType1 := a.analyzeExpression(args[0])
	if argType1 != nil && argType1 != types.FLOAT && argType1 != types.INTEGER {
		a.addError("function 'IntPower' expects Float or Integer as first argument, got %s at %s",
			argType1.String(), callExpr.Token.Pos.String())
	}

	argType2 := a.analyzeExpression(args[1])
	if argType2 != nil && argType2 != types.INTEGER {
		a.addError("function 'IntPower' expects Integer as second argument, got %s at %s",
			argType2.String(), callExpr.Token.Pos.String())
	}

	return types.FLOAT
}

// analyzeSqrt analyzes the Sqrt built-in function.
// Sqrt takes one numeric argument and always returns Float.
func (a *Analyzer) analyzeSqrt(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Sqrt' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType != types.INTEGER && argType != types.FLOAT {
			a.addError("function 'Sqrt' expects Integer or Float as argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeExp analyzes the Exp built-in function.
// Exp takes one numeric argument and always returns Float.
func (a *Analyzer) analyzeExp(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Exp' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType != types.INTEGER && argType != types.FLOAT {
			a.addError("function 'Exp' expects Integer or Float as argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeLn analyzes the Ln built-in function.
// Ln takes one numeric argument and always returns Float.
func (a *Analyzer) analyzeLn(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Ln' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType != types.INTEGER && argType != types.FLOAT {
			a.addError("function 'Ln' expects Integer or Float as argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeLog2 analyzes the Log2 built-in function.
// Log2 takes one numeric argument and always returns Float.
func (a *Analyzer) analyzeLog2(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Log2' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType != types.INTEGER && argType != types.FLOAT {
			a.addError("function 'Log2' expects Integer or Float as argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeLog10 analyzes the Log10 built-in function.
// Log10 takes a Float or Integer and returns the base-10 logarithm as Float.
func (a *Analyzer) analyzeLog10(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Log10' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.FLOAT && argType != types.INTEGER {
		a.addError("function 'Log10' expects Float or Integer, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.FLOAT
}

// analyzeLogN analyzes the LogN built-in function.
// LogN takes two Float or Integer arguments and returns the logarithm with custom base as Float.
func (a *Analyzer) analyzeLogN(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'LogN' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}

	argType1 := a.analyzeExpression(args[0])
	if argType1 != nil && argType1 != types.FLOAT && argType1 != types.INTEGER {
		a.addError("function 'LogN' expects Float or Integer as first argument, got %s at %s",
			argType1.String(), callExpr.Token.Pos.String())
	}

	argType2 := a.analyzeExpression(args[1])
	if argType2 != nil && argType2 != types.FLOAT && argType2 != types.INTEGER {
		a.addError("function 'LogN' expects Float or Integer as second argument, got %s at %s",
			argType2.String(), callExpr.Token.Pos.String())
	}

	return types.FLOAT
}

// analyzePi analyzes the Pi built-in constant.
// Pi: Float
func (a *Analyzer) analyzePi(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 0 {
		a.addError("function 'Pi' expects no arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	return types.FLOAT
}

// analyzeInfinity analyzes the Infinity built-in constant.
// Infinity: Float
func (a *Analyzer) analyzeInfinity(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 0 {
		a.addError("function 'Infinity' expects no arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	return types.FLOAT
}

// analyzeNaN analyzes the NaN built-in constant.
// NaN: Float
func (a *Analyzer) analyzeNaN(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 0 {
		a.addError("function 'NaN' expects no arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	return types.FLOAT
}

// analyzeIsNaN analyzes the IsNaN built-in function.
// IsNaN takes 1 Float argument and returns Boolean.
func (a *Analyzer) analyzeIsNaN(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'IsNaN' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.FLOAT {
		// Don't error - just check at runtime
		// This allows IsNaN to be called on any type
	}
	return types.BOOLEAN
}

// analyzeIsFinite analyzes the IsFinite built-in function.
// IsFinite takes a Float or Integer and returns a Boolean.
func (a *Analyzer) analyzeIsFinite(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'IsFinite' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.FLOAT && argType != types.INTEGER {
		a.addError("function 'IsFinite' expects Float or Integer, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.BOOLEAN
}

// analyzeIsInfinite analyzes the IsInfinite built-in function.
// IsInfinite takes a Float or Integer and returns a Boolean.
func (a *Analyzer) analyzeIsInfinite(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'IsInfinite' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.FLOAT && argType != types.INTEGER {
		a.addError("function 'IsInfinite' expects Float or Integer, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.BOOLEAN
}
