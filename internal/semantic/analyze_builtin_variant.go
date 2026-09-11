package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
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
	// A literal names a Variant type for comparison; a String variable does not.
	if _, ok := args[0].(*ast.StringLiteral); ok {
		return types.INTEGER
	}

	// DWScript boxes any value into a Variant for a Variant parameter, so
	// VarType accepts every analyzable expression type.
	a.analyzeExpression(args[0])
	return types.INTEGER
}

// analyzeVarToStr analyzes the VarToStr built-in function.
// The conversion is accepted for every type, but DWScript hints when a
// dedicated conversion (or no conversion at all) reads better.
func (a *Analyzer) analyzeVarToStr(args []ast.Expression, callExpr *ast.CallExpression) (types.Type, bool) {
	result, handled := a.analyzeRegisteredBuiltin("vartostr", args, callExpr)
	if !handled || len(args) != 1 || a.hintsLevel < HintsLevelNormal {
		return result, handled
	}

	argType := a.semanticInfo.GetResolvedType(args[0])
	if argType == nil {
		return result, handled
	}
	pos := callExpr.Function.Pos()
	switch types.GetUnderlyingType(argType) {
	case types.INTEGER:
		a.addHint("Prefer .ToString or IntToStr() [line: %d, column: %d]", pos.Line, pos.Column)
	case types.FLOAT:
		a.addHint("Prefer .ToString or FloatToStr() [line: %d, column: %d]", pos.Line, pos.Column)
	case types.STRING:
		a.addHint("Redundant function call [line: %d, column: %d]", pos.Line, pos.Column)
	}
	return result, handled
}
