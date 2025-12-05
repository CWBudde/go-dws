package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Rounding & Conversion Built-in Function Analysis
// ============================================================================
// This file contains analyzers for rounding and conversion functions:
// - Round, Trunc, Ceil, Floor
// - Frac, Int
// - Unsigned32

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

// analyzeTrunc analyzes the Trunc built-in function.
// Trunc takes one numeric argument and always returns Integer.
func (a *Analyzer) analyzeTrunc(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Trunc' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType != types.INTEGER && argType != types.FLOAT {
			a.addError("function 'Trunc' expects Integer or Float as argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeCeil analyzes the Ceil built-in function.
// Ceil takes one numeric argument and always returns Integer.
func (a *Analyzer) analyzeCeil(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Ceil' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType != types.INTEGER && argType != types.FLOAT {
			a.addError("function 'Ceil' expects Integer or Float as argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeFloor analyzes the Floor built-in function.
// Floor takes one numeric argument and always returns Integer.
func (a *Analyzer) analyzeFloor(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Floor' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType != types.INTEGER && argType != types.FLOAT {
			a.addError("function 'Floor' expects Integer or Float as argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeFrac analyzes the Frac built-in function.
// Frac takes a Float or Integer and returns the fractional part as Float.
func (a *Analyzer) analyzeFrac(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Frac' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.FLOAT && argType != types.INTEGER {
		a.addError("function 'Frac' expects Float or Integer, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.FLOAT
}

// analyzeInt analyzes the Int built-in function.
// Int takes a Float or Integer and returns the integer part as Float.
func (a *Analyzer) analyzeInt(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Int' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.FLOAT && argType != types.INTEGER {
		a.addError("function 'Int' expects Float or Integer, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.FLOAT
}

// analyzeUnsigned32 analyzes the Unsigned32 built-in function.
// Unsigned32 converts Integer to unsigned 32-bit representation.
func (a *Analyzer) analyzeUnsigned32(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Unsigned32' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.INTEGER {
		a.addError("function 'Unsigned32' expects Integer argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}
