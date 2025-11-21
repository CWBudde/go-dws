package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// analyzeFactorial analyzes the Factorial() built-in function.
// Factorial(n: Integer): Integer
func (a *Analyzer) analyzeFactorial(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Factorial' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}

	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.INTEGER {
		a.addError("function 'Factorial' expects Integer argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}

	return types.INTEGER
}

// analyzeGcd analyzes the Gcd() built-in function.
// Gcd(a, b: Integer): Integer
func (a *Analyzer) analyzeGcd(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'Gcd' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}

	argType1 := a.analyzeExpression(args[0])
	argType2 := a.analyzeExpression(args[1])

	if argType1 != nil && argType1 != types.INTEGER {
		a.addError("function 'Gcd' expects Integer as first argument, got %s at %s",
			argType1.String(), callExpr.Token.Pos.String())
	}

	if argType2 != nil && argType2 != types.INTEGER {
		a.addError("function 'Gcd' expects Integer as second argument, got %s at %s",
			argType2.String(), callExpr.Token.Pos.String())
	}

	return types.INTEGER
}

// analyzeLcm analyzes the Lcm() built-in function.
// Lcm(a, b: Integer): Integer
func (a *Analyzer) analyzeLcm(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'Lcm' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}

	argType1 := a.analyzeExpression(args[0])
	argType2 := a.analyzeExpression(args[1])

	if argType1 != nil && argType1 != types.INTEGER {
		a.addError("function 'Lcm' expects Integer as first argument, got %s at %s",
			argType1.String(), callExpr.Token.Pos.String())
	}

	if argType2 != nil && argType2 != types.INTEGER {
		a.addError("function 'Lcm' expects Integer as second argument, got %s at %s",
			argType2.String(), callExpr.Token.Pos.String())
	}

	return types.INTEGER
}

// analyzeIsPrime analyzes the IsPrime() built-in function.
// IsPrime(n: Integer): Boolean
func (a *Analyzer) analyzeIsPrime(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'IsPrime' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}

	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.INTEGER {
		a.addError("function 'IsPrime' expects Integer argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}

	return types.BOOLEAN
}

// analyzeLeastFactor analyzes the LeastFactor() built-in function.
// LeastFactor(n: Integer): Integer
func (a *Analyzer) analyzeLeastFactor(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'LeastFactor' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}

	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.INTEGER {
		a.addError("function 'LeastFactor' expects Integer argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}

	return types.INTEGER
}

// analyzePopCount analyzes the PopCount() built-in function.
// PopCount(n: Integer): Integer
func (a *Analyzer) analyzePopCount(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'PopCount' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}

	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.INTEGER {
		a.addError("function 'PopCount' expects Integer argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}

	return types.INTEGER
}

// analyzeTestBit analyzes the TestBit() built-in function.
// TestBit(value: Integer, bit: Integer): Boolean
func (a *Analyzer) analyzeTestBit(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'TestBit' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}

	argType1 := a.analyzeExpression(args[0])
	argType2 := a.analyzeExpression(args[1])

	if argType1 != nil && argType1 != types.INTEGER {
		a.addError("function 'TestBit' expects Integer as first argument, got %s at %s",
			argType1.String(), callExpr.Token.Pos.String())
	}

	if argType2 != nil && argType2 != types.INTEGER {
		a.addError("function 'TestBit' expects Integer as second argument, got %s at %s",
			argType2.String(), callExpr.Token.Pos.String())
	}

	return types.BOOLEAN
}

// analyzeHaversine analyzes the Haversine() built-in function.
// Haversine(lat1, lon1, lat2, lon2: Float): Float
func (a *Analyzer) analyzeHaversine(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 4 {
		a.addError("function 'Haversine' expects 4 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}

	// Analyze all four arguments - they should be numeric (Integer or Float)
	for i, arg := range args {
		argType := a.analyzeExpression(arg)
		if argType != nil && !types.IsNumericType(argType) {
			a.addError("function 'Haversine' expects numeric argument %d, got %s at %s",
				i+1, argType.String(), callExpr.Token.Pos.String())
		}
	}

	return types.FLOAT
}

// analyzeCompareNum analyzes the CompareNum() built-in function.
// CompareNum(a, b: Float): Integer
func (a *Analyzer) analyzeCompareNum(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'CompareNum' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}

	argType1 := a.analyzeExpression(args[0])
	argType2 := a.analyzeExpression(args[1])

	if argType1 != nil && !types.IsNumericType(argType1) {
		a.addError("function 'CompareNum' expects numeric first argument, got %s at %s",
			argType1.String(), callExpr.Token.Pos.String())
	}

	if argType2 != nil && !types.IsNumericType(argType2) {
		a.addError("function 'CompareNum' expects numeric second argument, got %s at %s",
			argType2.String(), callExpr.Token.Pos.String())
	}

	return types.INTEGER
}
