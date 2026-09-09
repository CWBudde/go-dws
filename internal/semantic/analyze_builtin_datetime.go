package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Date/Time Built-in Function Analysis
// ============================================================================

// analyzeDecodeDate analyzes the DecodeDate built-in procedure.
// DecodeDate takes 4 arguments (dt, var year, var month, var day) and returns void.
func (a *Analyzer) analyzeDecodeDate(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 4 {
		a.addError("function 'DecodeDate' expects 4 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	// First argument: TDateTime (Float)
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'DecodeDate' expects Float/TDateTime as first argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	// Other arguments are var parameters (year, month, day) - just analyze them
	for i := 1; i < len(args); i++ {
		a.analyzeExpression(args[i])
	}
	return types.VOID
}

// analyzeDecodeTime analyzes the DecodeTime built-in procedure.
// DecodeTime takes 5 arguments (dt, var hour, var minute, var second, var msec) and returns void.
func (a *Analyzer) analyzeDecodeTime(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 5 {
		a.addError("function 'DecodeTime' expects 5 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	// First argument: TDateTime (Float)
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'DecodeTime' expects Float/TDateTime as first argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	// Other arguments are var parameters (hour, minute, second, msec) - just analyze them
	for i := 1; i < len(args); i++ {
		a.analyzeExpression(args[i])
	}
	return types.VOID
}
