package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Advanced Math Built-in Function Analysis
// ============================================================================

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
