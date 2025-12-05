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

// analyzeSin analyzes the Sin built-in function.
// Sin takes one numeric argument and always returns Float.
func (a *Analyzer) analyzeSin(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Sin' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType != types.INTEGER && argType != types.FLOAT {
			a.addError("function 'Sin' expects Integer or Float as argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeCos analyzes the Cos built-in function.
// Cos takes one numeric argument and always returns Float.
func (a *Analyzer) analyzeCos(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Cos' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType != types.INTEGER && argType != types.FLOAT {
			a.addError("function 'Cos' expects Integer or Float as argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeTan analyzes the Tan built-in function.
// Tan takes one numeric argument and always returns Float.
func (a *Analyzer) analyzeTan(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Tan' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType != types.INTEGER && argType != types.FLOAT {
			a.addError("function 'Tan' expects Integer or Float as argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeCoTan analyzes the CoTan built-in function.
// CoTan calculates the cotangent.
func (a *Analyzer) analyzeCoTan(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'CoTan' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.INTEGER && argType != types.FLOAT {
		a.addError("function 'CoTan' expects Integer or Float as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.FLOAT
}

// analyzeDegToRad analyzes the DegToRad built-in function
// DegToRad converts degrees to radians.
func (a *Analyzer) analyzeDegToRad(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'DegToRad' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.INTEGER && argType != types.FLOAT {
		a.addError("function 'DegToRad' expects Integer or Float as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.FLOAT
}

// analyzeRadToDeg analyzes the RadToDeg built-in function.
// RadToDeg converts radians to degrees.
func (a *Analyzer) analyzeRadToDeg(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'RadToDeg' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.INTEGER && argType != types.FLOAT {
		a.addError("function 'RadToDeg' expects Integer or Float as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.FLOAT
}

// analyzeArcSin analyzes the ArcSin built-in function.
// ArcSin calculates the inverse sine.
func (a *Analyzer) analyzeArcSin(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'ArcSin' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.INTEGER && argType != types.FLOAT {
		a.addError("function 'ArcSin' expects Integer or Float as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.FLOAT
}

// analyzeArcCos analyzes the ArcCos built-in function.
// ArcCos calculates the inverse cosine.
func (a *Analyzer) analyzeArcCos(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'ArcCos' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.INTEGER && argType != types.FLOAT {
		a.addError("function 'ArcCos' expects Integer or Float as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.FLOAT
}

// analyzeArcTan analyzes the ArcTan built-in function.
// ArcTan calculates the inverse tangent.
func (a *Analyzer) analyzeArcTan(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'ArcTan' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.INTEGER && argType != types.FLOAT {
		a.addError("function 'ArcTan' expects Integer or Float as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.FLOAT
}

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

// analyzeSinh analyzes the Sinh built-in function.
// Sinh calculates the hyperbolic sine.
func (a *Analyzer) analyzeSinh(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Sinh' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.INTEGER && argType != types.FLOAT {
		a.addError("function 'Sinh' expects Integer or Float as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.FLOAT
}

// analyzeCosh analyzes the Cosh built-in function.
// Cosh calculates the hyperbolic cosine.
func (a *Analyzer) analyzeCosh(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Cosh' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.INTEGER && argType != types.FLOAT {
		a.addError("function 'Cosh' expects Integer or Float as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.FLOAT
}

// analyzeTanh analyzes the Tanh built-in function.
// Tanh calculates the hyperbolic tangent.
func (a *Analyzer) analyzeTanh(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Tanh' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.INTEGER && argType != types.FLOAT {
		a.addError("function 'Tanh' expects Integer or Float as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.FLOAT
}

// analyzeArcSinh analyzes the ArcSinh built-in function.
// ArcSinh calculates the inverse hyperbolic sine.
func (a *Analyzer) analyzeArcSinh(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'ArcSinh' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.INTEGER && argType != types.FLOAT {
		a.addError("function 'ArcSinh' expects Integer or Float as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.FLOAT
}

// analyzeArcCosh analyzes the ArcCosh built-in function.
// ArcCosh calculates the inverse hyperbolic cosine.
func (a *Analyzer) analyzeArcCosh(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'ArcCosh' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.INTEGER && argType != types.FLOAT {
		a.addError("function 'ArcCosh' expects Integer or Float as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.FLOAT
}

// analyzeArcTanh analyzes the ArcTanh built-in function.
// ArcTanh calculates the inverse hyperbolic tangent.
func (a *Analyzer) analyzeArcTanh(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'ArcTanh' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.INTEGER && argType != types.FLOAT {
		a.addError("function 'ArcTanh' expects Integer or Float as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.FLOAT
}
