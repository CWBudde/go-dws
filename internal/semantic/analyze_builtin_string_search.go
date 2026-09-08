package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// String Built-in Function Analysis - Search Functions
// ============================================================================
// Functions for searching, finding, and comparing strings

// analyzePos analyzes the Pos built-in function.
// Pos takes two string arguments and an optional integer offset and returns an integer.
func (a *Analyzer) analyzePos(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) < 2 || len(args) > 3 {
		a.addError("function 'Pos' expects 2 or 3 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze the first argument (substring)
	substrType := a.analyzeExpression(args[0])
	if substrType != nil && substrType != types.STRING {
		a.addError("function 'Pos' expects string as first argument, got %s at %s",
			substrType.String(), callExpr.Token.Pos.String())
	}
	// Analyze the second argument (string to search in)
	strType := a.analyzeExpression(args[1])
	if strType != nil && strType != types.STRING {
		a.addError("function 'Pos' expects string as second argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	if len(args) == 3 {
		offsetType := a.analyzeExpression(args[2])
		if offsetType != nil && offsetType != types.INTEGER {
			a.addError("function 'Pos' expects integer as third argument, got %s at %s",
				offsetType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeStrFind analyzes the StrFind built-in function.
// StrFind(str, substr [, fromIndex]) - returns integer (1-based position, 0 if not found)
func (a *Analyzer) analyzeStrFind(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) < 2 || len(args) > 3 {
		a.addError("function 'StrFind' expects 2 or 3 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze first argument (str - string)
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'StrFind' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (substr - string)
	substrType := a.analyzeExpression(args[1])
	if substrType != nil && substrType != types.STRING {
		a.addError("function 'StrFind' expects string as second argument, got %s at %s",
			substrType.String(), callExpr.Token.Pos.String())
	}
	if len(args) == 3 {
		// Analyze third argument (fromIndex - integer)
		fromIndexType := a.analyzeExpression(args[2])
		if fromIndexType != nil && fromIndexType != types.INTEGER {
			a.addError("function 'StrFind' expects integer as third argument, got %s at %s",
				fromIndexType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeFindDelimiter analyzes the FindDelimiter built-in function.
// FindDelimiter(delims, str, startIndex) - returns integer
func (a *Analyzer) analyzeFindDelimiter(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) < 2 || len(args) > 3 {
		a.addError("function 'FindDelimiter' expects 2 or 3 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze first argument (delims - string)
	delimsType := a.analyzeExpression(args[0])
	if delimsType != nil && delimsType != types.STRING {
		a.addError("function 'FindDelimiter' expects string as first argument, got %s at %s",
			delimsType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (str - string)
	strType := a.analyzeExpression(args[1])
	if strType != nil && strType != types.STRING {
		a.addError("function 'FindDelimiter' expects string as second argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	if len(args) == 3 {
		// Analyze third argument (startIndex - integer)
		startIndexType := a.analyzeExpression(args[2])
		if startIndexType != nil && startIndexType != types.INTEGER {
			a.addError("function 'FindDelimiter' expects integer as third argument, got %s at %s",
				startIndexType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeCompareLocaleStr analyzes the CompareLocaleStr built-in function.
// CompareLocaleStr takes 2 to 4 arguments: CompareLocaleStr(str1, str2 [, locale [, caseSensitive]])
func (a *Analyzer) analyzeCompareLocaleStr(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) < 2 || len(args) > 4 {
		a.addError("function 'CompareLocaleStr' expects 2 to 4 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	str1Type := a.analyzeExpression(args[0])
	if str1Type != nil && str1Type != types.STRING {
		a.addError("function 'CompareLocaleStr' expects string as first argument, got %s at %s",
			str1Type.String(), callExpr.Token.Pos.String())
	}
	str2Type := a.analyzeExpression(args[1])
	if str2Type != nil && str2Type != types.STRING {
		a.addError("function 'CompareLocaleStr' expects string as second argument, got %s at %s",
			str2Type.String(), callExpr.Token.Pos.String())
	}
	if len(args) >= 3 {
		localeType := a.analyzeExpression(args[2])
		if localeType != nil && localeType != types.STRING {
			a.addError("function 'CompareLocaleStr' expects string as third argument, got %s at %s",
				localeType.String(), callExpr.Token.Pos.String())
		}
	}
	if len(args) == 4 {
		csType := a.analyzeExpression(args[3])
		if csType != nil && csType != types.BOOLEAN {
			a.addError("function 'CompareLocaleStr' expects boolean as fourth argument, got %s at %s",
				csType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}
