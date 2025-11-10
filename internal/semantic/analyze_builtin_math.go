package semantic

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Math Built-in Function Analysis
// ============================================================================

// analyzeAbs analyzes the Abs built-in function.
// Abs takes one numeric argument and returns the same type.
func (a *Analyzer) analyzeAbs(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Abs' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER // Default to INTEGER on error
	}
	// Analyze the argument and verify it's Integer or Float
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType != types.INTEGER && argType != types.FLOAT {
			a.addError("function 'Abs' expects Integer or Float as argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
			return types.INTEGER
		}
		// Return the same type as the input
		return argType
	}
	return types.INTEGER // Default to INTEGER if type is unknown
}

// analyzeMin analyzes the Min built-in function.
// Min takes two numeric arguments and returns the smaller value.
func (a *Analyzer) analyzeMin(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'Min' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER // Default to INTEGER on error
	}
	// Analyze both arguments and verify they're Integer or Float
	arg1Type := a.analyzeExpression(args[0])
	arg2Type := a.analyzeExpression(args[1])

	if arg1Type != nil && arg2Type != nil {
		// Verify both are numeric
		if (arg1Type != types.INTEGER && arg1Type != types.FLOAT) ||
			(arg2Type != types.INTEGER && arg2Type != types.FLOAT) {
			a.addError("function 'Min' expects Integer or Float arguments, got %s and %s at %s",
				arg1Type.String(), arg2Type.String(), callExpr.Token.Pos.String())
			return types.INTEGER
		}
		// If both Integer, return Integer; otherwise return Float
		if arg1Type == types.INTEGER && arg2Type == types.INTEGER {
			return types.INTEGER
		}
		return types.FLOAT
	}
	return types.INTEGER // Default to INTEGER if type is unknown
}

// analyzeMax analyzes the Max built-in function.
// Max takes two numeric arguments and returns the larger value.
func (a *Analyzer) analyzeMax(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'Max' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER // Default to INTEGER on error
	}
	// Analyze both arguments and verify they're Integer or Float
	arg1Type := a.analyzeExpression(args[0])
	arg2Type := a.analyzeExpression(args[1])

	if arg1Type != nil && arg2Type != nil {
		// Verify both are numeric
		if (arg1Type != types.INTEGER && arg1Type != types.FLOAT) ||
			(arg2Type != types.INTEGER && arg2Type != types.FLOAT) {
			a.addError("function 'Max' expects Integer or Float arguments, got %s and %s at %s",
				arg1Type.String(), arg2Type.String(), callExpr.Token.Pos.String())
			return types.INTEGER
		}
		// If both Integer, return Integer; otherwise return Float
		if arg1Type == types.INTEGER && arg2Type == types.INTEGER {
			return types.INTEGER
		}
		return types.FLOAT
	}
	return types.INTEGER // Default to INTEGER if type is unknown
}

// analyzeClampInt analyzes the ClampInt built-in function.
// ClampInt takes three integer arguments: value, min, max.
func (a *Analyzer) analyzeClampInt(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 3 {
		a.addError("function 'ClampInt' expects 3 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER // Default to INTEGER on error
	}
	// Analyze all three arguments and verify they're Integer
	arg1Type := a.analyzeExpression(args[0])
	arg2Type := a.analyzeExpression(args[1])
	arg3Type := a.analyzeExpression(args[2])

	if arg1Type != nil && arg2Type != nil && arg3Type != nil {
		// Verify all are Integer (not Float)
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
		return types.FLOAT // Default to FLOAT on error
	}
	// Analyze all three arguments and verify they're numeric
	arg1Type := a.analyzeExpression(args[0])
	arg2Type := a.analyzeExpression(args[1])
	arg3Type := a.analyzeExpression(args[2])

	if arg1Type != nil && arg2Type != nil && arg3Type != nil {
		// Verify all are numeric
		if (arg1Type != types.INTEGER && arg1Type != types.FLOAT) ||
			(arg2Type != types.INTEGER && arg2Type != types.FLOAT) ||
			(arg3Type != types.INTEGER && arg3Type != types.FLOAT) {
			a.addError("function 'Clamp' expects Integer or Float arguments, got %s, %s, %s at %s",
				arg1Type.String(), arg2Type.String(), arg3Type.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT // Always returns Float
}

// analyzeMaxInt analyzes the MaxInt built-in function.
// MaxInt() with 0 args returns max integer constant.
// MaxInt(a, b, ...) with 2+ args returns maximum of all integers (variadic).
func (a *Analyzer) analyzeMaxInt(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) == 0 {
		return types.INTEGER
	}
	// Variadic: accepts any number of arguments >= 2
	// Analyze all arguments and verify they're all Integer
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
	// Variadic: accepts any number of arguments >= 2
	// Analyze all arguments and verify they're all Integer
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
		return types.INTEGER // Default to INTEGER on error
	}
	// Analyze the argument and verify it's Integer or Float
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType != types.INTEGER && argType != types.FLOAT {
			a.addError("function 'Sqr' expects Integer or Float as argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
			return types.INTEGER
		}
		// Return the same type as the input
		return argType
	}
	return types.INTEGER // Default to INTEGER if type is unknown
}

// analyzePower analyzes the Power built-in function.
// Power takes two numeric arguments and always returns Float.
func (a *Analyzer) analyzePower(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'Power' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT // Default to FLOAT on error
	}
	// Analyze both arguments and verify they're Integer or Float
	arg1Type := a.analyzeExpression(args[0])
	arg2Type := a.analyzeExpression(args[1])

	if arg1Type != nil && arg2Type != nil {
		// Verify both are numeric
		if (arg1Type != types.INTEGER && arg1Type != types.FLOAT) ||
			(arg2Type != types.INTEGER && arg2Type != types.FLOAT) {
			a.addError("function 'Power' expects Integer or Float arguments, got %s and %s at %s",
				arg1Type.String(), arg2Type.String(), callExpr.Token.Pos.String())
		}
	}
	// Always returns Float
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
	// Analyze the argument and verify it's Integer or Float
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType != types.INTEGER && argType != types.FLOAT {
			a.addError("function 'Sqrt' expects Integer or Float as argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	// Always returns Float
	return types.FLOAT
}

// analyzeSin analyzes the Sin built-in function.
// Sin takes one numeric argument and always returns Float.
func (a *Analyzer) analyzeSin(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Sin' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	// Analyze the argument and verify it's Integer or Float
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType != types.INTEGER && argType != types.FLOAT {
			a.addError("function 'Sin' expects Integer or Float as argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	// Always returns Float
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
	// Analyze the argument and verify it's Integer or Float
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType != types.INTEGER && argType != types.FLOAT {
			a.addError("function 'Cos' expects Integer or Float as argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	// Always returns Float
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
	// Analyze the argument and verify it's Integer or Float
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType != types.INTEGER && argType != types.FLOAT {
			a.addError("function 'Tan' expects Integer or Float as argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	// Always returns Float
	return types.FLOAT
}

// analyzeDegToRad analyzes the DegToRad built-in function (Task 9.232).
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

// analyzeRadToDeg analyzes the RadToDeg built-in function (Task 9.232).
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

// analyzeArcSin analyzes the ArcSin built-in function (Task 9.232).
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

// analyzeArcCos analyzes the ArcCos built-in function (Task 9.232).
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

// analyzeArcTan analyzes the ArcTan built-in function (Task 9.232).
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

// analyzeArcTan2 analyzes the ArcTan2 built-in function (Task 9.232).
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

// analyzeCoTan analyzes the CoTan built-in function (Task 9.232).
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

// analyzeHypot analyzes the Hypot built-in function (Task 9.232).
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

// analyzeSinh analyzes the Sinh built-in function (Task 9.232).
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

// analyzeCosh analyzes the Cosh built-in function (Task 9.232).
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

// analyzeTanh analyzes the Tanh built-in function (Task 9.232).
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

// analyzeArcSinh analyzes the ArcSinh built-in function (Task 9.232).
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

// analyzeArcCosh analyzes the ArcCosh built-in function (Task 9.232).
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

// analyzeArcTanh analyzes the ArcTanh built-in function (Task 9.232).
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

// analyzeRandom analyzes the Random built-in function.
// Random takes no arguments and always returns Float.
func (a *Analyzer) analyzeRandom(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 0 {
		a.addError("function 'Random' expects no arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	// Always returns Float
	return types.FLOAT
}

// analyzeRandomInt analyzes the RandomInt built-in function.
// RandomInt takes one Integer argument and returns random Integer in [0, max).
func (a *Analyzer) analyzeRandomInt(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'RandomInt' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER // Default to INTEGER on error
	}
	// Analyze argument and verify it's Integer
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.INTEGER {
		a.addError("function 'RandomInt' expects Integer argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	// Always returns Integer
	return types.INTEGER
}

// analyzeUnsigned32 analyzes the Unsigned32 built-in function (Task 9.219).
// Unsigned32 converts Integer to unsigned 32-bit representation.
func (a *Analyzer) analyzeUnsigned32(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Unsigned32' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER // Default to INTEGER on error
	}
	// Analyze argument and verify it's Integer
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.INTEGER {
		a.addError("function 'Unsigned32' expects Integer argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	// Always returns Integer (holding uint32 value)
	return types.INTEGER
}

// analyzeRandomize analyzes the Randomize built-in procedure.
// Randomize takes no arguments and returns nothing (nil/void).
func (a *Analyzer) analyzeRandomize(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 0 {
		a.addError("function 'Randomize' expects no arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	// Returns nil/void (no meaningful return value)
	return nil
}

// analyzeExp analyzes the Exp built-in function.
// Exp takes one numeric argument and always returns Float.
func (a *Analyzer) analyzeExp(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Exp' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	// Analyze the argument and verify it's Integer or Float
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType != types.INTEGER && argType != types.FLOAT {
			a.addError("function 'Exp' expects Integer or Float as argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	// Always returns Float
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
	// Analyze the argument and verify it's Integer or Float
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType != types.INTEGER && argType != types.FLOAT {
			a.addError("function 'Ln' expects Integer or Float as argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	// Always returns Float
	return types.FLOAT
}

// analyzeLog2 analyzes the Log2 built-in function (Task 9.38).
// Log2 takes one numeric argument and always returns Float.
func (a *Analyzer) analyzeLog2(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Log2' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	// Analyze the argument and verify it's Integer or Float
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType != types.INTEGER && argType != types.FLOAT {
			a.addError("function 'Log2' expects Integer or Float as argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	// Always returns Float
	return types.FLOAT
}

// analyzeRound analyzes the Round built-in function.
// Round takes one numeric argument and always returns Integer.
func (a *Analyzer) analyzeRound(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Round' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze the argument and verify it's Integer or Float
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType != types.INTEGER && argType != types.FLOAT {
			a.addError("function 'Round' expects Integer or Float as argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	// Always returns Integer
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
	// Analyze the argument and verify it's Integer or Float
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType != types.INTEGER && argType != types.FLOAT {
			a.addError("function 'Trunc' expects Integer or Float as argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	// Always returns Integer
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
	// Analyze the argument and verify it's Integer or Float
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType != types.INTEGER && argType != types.FLOAT {
			a.addError("function 'Ceil' expects Integer or Float as argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	// Always returns Integer
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
	// Analyze the argument and verify it's Integer or Float
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType != types.INTEGER && argType != types.FLOAT {
			a.addError("function 'Floor' expects Integer or Float as argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	// Always returns Integer
	return types.INTEGER
}

// analyzeInc analyzes the Inc built-in procedure.
// Inc takes 1-2 arguments: variable and optional delta.
func (a *Analyzer) analyzeInc(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) < 1 || len(args) > 2 {
		a.addError("function 'Inc' expects 1-2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.VOID
	}
	// First argument must be a variable (Identifier)
	if _, ok := args[0].(*ast.Identifier); !ok {
		a.addError("function 'Inc' first argument must be a variable at %s",
			callExpr.Token.Pos.String())
	} else {
		// Analyze the variable to get its type
		varType := a.analyzeExpression(args[0])
		// Must be Integer or Enum
		if varType != nil {
			if varType != types.INTEGER {
				if _, isEnum := varType.(*types.EnumType); !isEnum {
					a.addError("function 'Inc' expects Integer or Enum variable, got %s at %s",
						varType.String(), callExpr.Token.Pos.String())
				}
			}
		}
	}
	// If there's a second argument (delta), it must be Integer
	if len(args) == 2 {
		deltaType := a.analyzeExpression(args[1])
		if deltaType != nil && deltaType != types.INTEGER {
			a.addError("function 'Inc' delta must be Integer, got %s at %s",
				deltaType.String(), callExpr.Token.Pos.String())
		}
	}
	// In DWScript, Inc returns the incremented value (like prefix ++ in C)
	// This allows Inc to be used in expressions: arr[Inc(k)]
	return types.INTEGER
}

// analyzeDec analyzes the Dec built-in function.
// Dec takes 1-2 arguments: variable and optional delta.
// Returns the decremented value (like prefix -- in C).
func (a *Analyzer) analyzeDec(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) < 1 || len(args) > 2 {
		a.addError("function 'Dec' expects 1-2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.VOID
	}
	// First argument must be a variable (Identifier)
	if _, ok := args[0].(*ast.Identifier); !ok {
		a.addError("function 'Dec' first argument must be a variable at %s",
			callExpr.Token.Pos.String())
	} else {
		// Analyze the variable to get its type
		varType := a.analyzeExpression(args[0])
		// Must be Integer or Enum
		if varType != nil {
			if varType != types.INTEGER {
				if _, isEnum := varType.(*types.EnumType); !isEnum {
					a.addError("function 'Dec' expects Integer or Enum variable, got %s at %s",
						varType.String(), callExpr.Token.Pos.String())
				}
			}
		}
	}
	// If there's a second argument (delta), it must be Integer
	if len(args) == 2 {
		deltaType := a.analyzeExpression(args[1])
		if deltaType != nil && deltaType != types.INTEGER {
			a.addError("function 'Dec' delta must be Integer, got %s at %s",
				deltaType.String(), callExpr.Token.Pos.String())
		}
	}
	// In DWScript, Dec returns the decremented value (like prefix -- in C)
	// This allows Dec to be used in expressions: arr[Dec(k)]
	return types.INTEGER
}

// analyzeSucc analyzes the Succ built-in function.
// Succ takes 1 argument: ordinal value and returns the successor.
func (a *Analyzer) analyzeSucc(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Succ' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze the argument to get its type
	argType := a.analyzeExpression(args[0])
	// Must be Integer or Enum
	if argType != nil {
		if argType == types.INTEGER {
			return types.INTEGER
		}
		if enumType, isEnum := argType.(*types.EnumType); isEnum {
			return enumType
		}
		a.addError("function 'Succ' expects Integer or Enum, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzePred analyzes the Pred built-in function.
// Pred takes 1 argument: ordinal value and returns the predecessor.
func (a *Analyzer) analyzePred(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Pred' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze the argument to get its type
	argType := a.analyzeExpression(args[0])
	// Must be Integer or Enum
	if argType != nil {
		if argType == types.INTEGER {
			return types.INTEGER
		}
		if enumType, isEnum := argType.(*types.EnumType); isEnum {
			return enumType
		}
		a.addError("function 'Pred' expects Integer or Enum, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeAssigned analyzes the Assigned built-in function.
// Assigned takes 1 argument and checks if a pointer/object/variant is nil.
// Returns Boolean.
func (a *Analyzer) analyzeAssigned(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Assigned' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}

	// Analyze the argument - Assigned can take any type
	// Objects, arrays, variants, etc. can all be checked
	a.analyzeExpression(args[0])

	// Always returns Boolean
	return types.BOOLEAN
}

// analyzeSwap analyzes the Swap built-in function.
// Swap takes 2 var arguments and swaps their values.
// Swap(var a, var b)
func (a *Analyzer) analyzeSwap(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'Swap' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return nil
	}

	// Both arguments must be identifiers (variables)
	for i, arg := range args {
		if _, ok := arg.(*ast.Identifier); !ok {
			a.addError("function 'Swap' argument %d must be a variable at %s",
				i+1, callExpr.Token.Pos.String())
		}
	}

	// Analyze both arguments to validate they exist and get their types
	type1 := a.analyzeExpression(args[0])
	type2 := a.analyzeExpression(args[1])

	// Both arguments must have compatible types
	if type1 != nil && type2 != nil {
		if !type1.Equals(type2) {
			a.addError("function 'Swap' arguments must have compatible types, got %s and %s at %s",
				type1.String(), type2.String(), callExpr.Token.Pos.String())
		}
	}

	// Swap doesn't return a value (procedure)
	return nil
}

// analyzeIsNaN analyzes the IsNaN built-in function.
// IsNaN takes 1 Float argument and returns Boolean.
// IsNaN(value: Float): Boolean
func (a *Analyzer) analyzeIsNaN(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'IsNaN' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}

	// Analyze the argument
	argType := a.analyzeExpression(args[0])

	// Argument should be Float (but we're lenient - will return false for non-floats at runtime)
	if argType != nil && argType != types.FLOAT {
		// Don't error - just check at runtime
		// This allows IsNaN to be called on any type
	}

	return types.BOOLEAN
}

// analyzeSetRandSeed analyzes the SetRandSeed built-in function.
// SetRandSeed takes 1 Integer argument and sets the random seed.
// SetRandSeed(seed: Integer)
func (a *Analyzer) analyzeSetRandSeed(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'SetRandSeed' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return nil
	}

	// Analyze the argument
	argType := a.analyzeExpression(args[0])

	// Argument must be Integer
	if argType != nil && argType != types.INTEGER {
		a.addError("function 'SetRandSeed' expects Integer, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}

	// SetRandSeed doesn't return a value (procedure)
	return nil
}
