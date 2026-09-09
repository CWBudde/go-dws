package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Basic Math Built-in Function Analysis
// ============================================================================

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
		// Variant arguments are coerced at runtime.
		if arg1Type == types.VARIANT || arg2Type == types.VARIANT {
			return types.VARIANT
		}
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
		// Variant arguments are coerced at runtime.
		if arg1Type == types.VARIANT || arg2Type == types.VARIANT {
			return types.VARIANT
		}
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

// analyzeDivMod analyzes the DivMod built-in procedure.
// DivMod(dividend, divisor: Integer; var quotient, remainder: Integer)
//
//nolint:unparam // returns nil for procedures (void functions)
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
