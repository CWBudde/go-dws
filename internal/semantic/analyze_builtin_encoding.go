package semantic

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Encoding/Escaping Built-in Function Analysis (Phase 9.17.6)
// ============================================================================

// analyzeStrToHtml analyzes the StrToHtml built-in function.
// StrToHtml takes one string argument and returns a string.
func (a *Analyzer) analyzeStrToHtml(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'StrToHtml' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze the argument and verify it's String
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.STRING {
		a.addError("function 'StrToHtml' expects String as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeStrToHtmlAttribute analyzes the StrToHtmlAttribute built-in function.
// StrToHtmlAttribute takes one string argument and returns a string.
func (a *Analyzer) analyzeStrToHtmlAttribute(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'StrToHtmlAttribute' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze the argument and verify it's String
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.STRING {
		a.addError("function 'StrToHtmlAttribute' expects String as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeStrToJSON analyzes the StrToJSON built-in function.
// StrToJSON takes one string argument and returns a string.
func (a *Analyzer) analyzeStrToJSON(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'StrToJSON' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze the argument and verify it's String
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.STRING {
		a.addError("function 'StrToJSON' expects String as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeStrToCSSText analyzes the StrToCSSText built-in function.
// StrToCSSText takes one string argument and returns a string.
func (a *Analyzer) analyzeStrToCSSText(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'StrToCSSText' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze the argument and verify it's String
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.STRING {
		a.addError("function 'StrToCSSText' expects String as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeStrToXML analyzes the StrToXML built-in function.
// StrToXML takes one or two arguments (string, [mode]) and returns a string.
// The mode parameter is optional and defaults to 0 (standard XML encoding).
func (a *Analyzer) analyzeStrToXML(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) < 1 || len(args) > 2 {
		a.addError("function 'StrToXML' expects 1 or 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze the first argument and verify it's String
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.STRING {
		a.addError("function 'StrToXML' expects String as first argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	// Analyze optional second argument (mode) - must be Integer
	if len(args) == 2 {
		modeType := a.analyzeExpression(args[1])
		if modeType != nil && modeType != types.INTEGER {
			// Check if it's a subrange type with Integer base
			if subrange, ok := modeType.(*types.SubrangeType); ok {
				if subrange.BaseType != types.INTEGER {
					a.addError("function 'StrToXML' expects Integer as second argument (mode), got %s at %s",
						modeType.String(), callExpr.Token.Pos.String())
				}
			} else {
				a.addError("function 'StrToXML' expects Integer as second argument (mode), got %s at %s",
					modeType.String(), callExpr.Token.Pos.String())
			}
		}
	}
	return types.STRING
}
