package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Trigonometric Built-in Function Analysis
// ============================================================================
// This file contains analyzers for trigonometric and hyperbolic functions:
// - Basic trig: Sin, Cos, Tan, CoTan
// - Inverse trig: ArcSin, ArcCos, ArcTan, ArcTan2
// - Hyperbolic: Sinh, Cosh, Tanh
// - Inverse hyperbolic: ArcSinh, ArcCosh, ArcTanh
// - Angle conversion: DegToRad, RadToDeg
// - Hypot

// analyzeArcTan2 analyzes the ArcTan2 built-in function.
// ArcTan2 calculates the angle from the X axis to a point.
func (a *Analyzer) analyzeArcTan2(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'ArcTan2' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	yType := a.analyzeExpression(args[0])
	xType := a.analyzeExpression(args[1])
	if yType != nil && yType != types.INTEGER && yType != types.FLOAT {
		a.addError("function 'ArcTan2' expects Integer or Float as first argument, got %s at %s",
			yType.String(), callExpr.Token.Pos.String())
	}
	if xType != nil && xType != types.INTEGER && xType != types.FLOAT {
		a.addError("function 'ArcTan2' expects Integer or Float as second argument, got %s at %s",
			xType.String(), callExpr.Token.Pos.String())
	}
	return types.FLOAT
}

// analyzeHypot analyzes the Hypot built-in function.
// Hypot calculates the Euclidean distance: sqrt(x*x + y*y).
func (a *Analyzer) analyzeHypot(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'Hypot' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	xType := a.analyzeExpression(args[0])
	yType := a.analyzeExpression(args[1])
	if xType != nil && xType != types.INTEGER && xType != types.FLOAT {
		a.addError("function 'Hypot' expects Integer or Float as first argument, got %s at %s",
			xType.String(), callExpr.Token.Pos.String())
	}
	if yType != nil && yType != types.INTEGER && yType != types.FLOAT {
		a.addError("function 'Hypot' expects Integer or Float as second argument, got %s at %s",
			yType.String(), callExpr.Token.Pos.String())
	}
	return types.FLOAT
}
