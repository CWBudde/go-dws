package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Rounding & Conversion Built-in Function Analysis
// ============================================================================

// analyzeRound analyzes the Round built-in function.
// Round takes one numeric argument and always returns Integer.
func (a *Analyzer) analyzeRound(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Round' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType != types.INTEGER && argType != types.FLOAT && argType != types.VARIANT {
			a.addError("function 'Round' expects Integer, Float, or Variant as argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}
