package semantic

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Variant Built-in Function Analysis
// ============================================================================

// analyzeVarType analyzes the VarType built-in function.
// VarType takes one Variant argument and returns an integer type code.
func (a *Analyzer) analyzeVarType(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'VarType' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze the argument (can be Variant or any type)
	a.analyzeExpression(args[0])
	return types.INTEGER
}

// analyzeVarIsNull analyzes the VarIsNull built-in function.
// VarIsNull takes one Variant argument and returns a boolean.
func (a *Analyzer) analyzeVarIsNull(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'VarIsNull' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}
	// Analyze the argument (can be Variant or any type)
	a.analyzeExpression(args[0])
	return types.BOOLEAN
}

// analyzeVarIsEmpty analyzes the VarIsEmpty built-in function.
// VarIsEmpty takes one Variant argument and returns a boolean.
func (a *Analyzer) analyzeVarIsEmpty(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'VarIsEmpty' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}
	// Analyze the argument (can be Variant or any type)
	a.analyzeExpression(args[0])
	return types.BOOLEAN
}

// analyzeVarIsNumeric analyzes the VarIsNumeric built-in function.
// VarIsNumeric takes one Variant argument and returns a boolean.
func (a *Analyzer) analyzeVarIsNumeric(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'VarIsNumeric' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}
	// Analyze the argument (can be Variant or any type)
	a.analyzeExpression(args[0])
	return types.BOOLEAN
}

// analyzeVarToInt analyzes the VarToInt built-in function.
// VarToInt takes one argument and returns an integer.
func (a *Analyzer) analyzeVarToInt(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'VarToInt' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	a.analyzeExpression(args[0])
	return types.INTEGER
}

// analyzeVarToFloat analyzes the VarToFloat built-in function.
// VarToFloat takes one argument and returns a float.
func (a *Analyzer) analyzeVarToFloat(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'VarToFloat' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	a.analyzeExpression(args[0])
	return types.FLOAT
}

// analyzeVarAsType analyzes the VarAsType built-in function.
// VarAsType takes two arguments (value, type code) and returns a Variant.
func (a *Analyzer) analyzeVarAsType(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'VarAsType' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.VARIANT
	}
	a.analyzeExpression(args[0])
	argType := a.analyzeExpression(args[1])
	// Second argument should be Integer (type code)
	if argType != nil && !argType.Equals(types.INTEGER) {
		a.addError("VarAsType type code must be Integer, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.VARIANT
}
