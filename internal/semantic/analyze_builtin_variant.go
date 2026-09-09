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

	argType := a.analyzeExpression(args[0])
	// JSONVariant participates in Variant introspection.
	if argType != nil && argType != types.VARIANT && !types.IsJSONVariant(argType) {
		a.addError("function '%s' expects Variant argument, got %s at %s",
			"VarType", argType.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}
