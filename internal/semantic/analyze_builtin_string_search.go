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

// analyzeStrBeginsWith analyzes the StrBeginsWith built-in function.
// StrBeginsWith(str, prefix) - returns boolean
func (a *Analyzer) analyzeStrBeginsWith(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'StrBeginsWith' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}
	// Analyze first argument (string)
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'StrBeginsWith' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (prefix - string)
	prefixType := a.analyzeExpression(args[1])
	if prefixType != nil && prefixType != types.STRING {
		a.addError("function 'StrBeginsWith' expects string as second argument, got %s at %s",
			prefixType.String(), callExpr.Token.Pos.String())
	}
	return types.BOOLEAN
}

// analyzeStrEndsWith analyzes the StrEndsWith built-in function.
// StrEndsWith(str, suffix) - returns boolean
func (a *Analyzer) analyzeStrEndsWith(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'StrEndsWith' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}
	// Analyze first argument (string)
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'StrEndsWith' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (suffix - string)
	suffixType := a.analyzeExpression(args[1])
	if suffixType != nil && suffixType != types.STRING {
		a.addError("function 'StrEndsWith' expects string as second argument, got %s at %s",
			suffixType.String(), callExpr.Token.Pos.String())
	}
	return types.BOOLEAN
}

// analyzeStrContains analyzes the StrContains built-in function.
// StrContains(str, substring) - returns boolean
func (a *Analyzer) analyzeStrContains(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'StrContains' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}
	// Analyze first argument (string)
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'StrContains' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (substring - string)
	substrType := a.analyzeExpression(args[1])
	if substrType != nil && substrType != types.STRING {
		a.addError("function 'StrContains' expects string as second argument, got %s at %s",
			substrType.String(), callExpr.Token.Pos.String())
	}
	return types.BOOLEAN
}

// analyzePosEx analyzes the PosEx built-in function.
// PosEx(needle, haystack, offset) - returns integer (1-based position, 0 if not found)
func (a *Analyzer) analyzePosEx(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 3 {
		a.addError("function 'PosEx' expects 3 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze first argument (needle - string)
	needleType := a.analyzeExpression(args[0])
	if needleType != nil && needleType != types.STRING {
		a.addError("function 'PosEx' expects string as first argument, got %s at %s",
			needleType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (haystack - string)
	haystackType := a.analyzeExpression(args[1])
	if haystackType != nil && haystackType != types.STRING {
		a.addError("function 'PosEx' expects string as second argument, got %s at %s",
			haystackType.String(), callExpr.Token.Pos.String())
	}
	// Analyze third argument (offset - integer)
	offsetType := a.analyzeExpression(args[2])
	if offsetType != nil && offsetType != types.INTEGER {
		a.addError("function 'PosEx' expects integer as third argument, got %s at %s",
			offsetType.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeRevPos analyzes the RevPos built-in function.
// RevPos(needle, haystack) - returns integer (1-based position of last occurrence, 0 if not found)
func (a *Analyzer) analyzeRevPos(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'RevPos' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze first argument (needle - string)
	needleType := a.analyzeExpression(args[0])
	if needleType != nil && needleType != types.STRING {
		a.addError("function 'RevPos' expects string as first argument, got %s at %s",
			needleType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (haystack - string)
	haystackType := a.analyzeExpression(args[1])
	if haystackType != nil && haystackType != types.STRING {
		a.addError("function 'RevPos' expects string as second argument, got %s at %s",
			haystackType.String(), callExpr.Token.Pos.String())
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

// analyzeIsDelimiter analyzes the IsDelimiter built-in function.
// IsDelimiter(delims, str, index) - returns boolean
func (a *Analyzer) analyzeIsDelimiter(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 3 {
		a.addError("function 'IsDelimiter' expects 3 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}
	// Analyze first argument (delims - string)
	delimsType := a.analyzeExpression(args[0])
	if delimsType != nil && delimsType != types.STRING {
		a.addError("function 'IsDelimiter' expects string as first argument, got %s at %s",
			delimsType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (str - string)
	strType := a.analyzeExpression(args[1])
	if strType != nil && strType != types.STRING {
		a.addError("function 'IsDelimiter' expects string as second argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	// Analyze third argument (index - integer)
	indexType := a.analyzeExpression(args[2])
	if indexType != nil && indexType != types.INTEGER {
		a.addError("function 'IsDelimiter' expects integer as third argument, got %s at %s",
			indexType.String(), callExpr.Token.Pos.String())
	}
	return types.BOOLEAN
}

// analyzeLastDelimiter analyzes the LastDelimiter built-in function.
// LastDelimiter(delims, str) - returns integer
func (a *Analyzer) analyzeLastDelimiter(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'LastDelimiter' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze first argument (delims - string)
	delimsType := a.analyzeExpression(args[0])
	if delimsType != nil && delimsType != types.STRING {
		a.addError("function 'LastDelimiter' expects string as first argument, got %s at %s",
			delimsType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (str - string)
	strType := a.analyzeExpression(args[1])
	if strType != nil && strType != types.STRING {
		a.addError("function 'LastDelimiter' expects string as second argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
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

// analyzeSameText analyzes the SameText built-in function.
// SameText takes 2 arguments: SameText(str1, str2)
func (a *Analyzer) analyzeSameText(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'SameText' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}
	str1Type := a.analyzeExpression(args[0])
	if str1Type != nil && str1Type != types.STRING {
		a.addError("function 'SameText' expects string as first argument, got %s at %s",
			str1Type.String(), callExpr.Token.Pos.String())
	}
	str2Type := a.analyzeExpression(args[1])
	if str2Type != nil && str2Type != types.STRING {
		a.addError("function 'SameText' expects string as second argument, got %s at %s",
			str2Type.String(), callExpr.Token.Pos.String())
	}
	return types.BOOLEAN
}

// analyzeCompareText analyzes the CompareText built-in function.
// CompareText takes 2 arguments: CompareText(str1, str2)
func (a *Analyzer) analyzeCompareText(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'CompareText' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	str1Type := a.analyzeExpression(args[0])
	if str1Type != nil && str1Type != types.STRING {
		a.addError("function 'CompareText' expects string as first argument, got %s at %s",
			str1Type.String(), callExpr.Token.Pos.String())
	}
	str2Type := a.analyzeExpression(args[1])
	if str2Type != nil && str2Type != types.STRING {
		a.addError("function 'CompareText' expects string as second argument, got %s at %s",
			str2Type.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeCompareStr analyzes the CompareStr built-in function.
// CompareStr takes 2 arguments: CompareStr(str1, str2)
func (a *Analyzer) analyzeCompareStr(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'CompareStr' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	str1Type := a.analyzeExpression(args[0])
	if str1Type != nil && str1Type != types.STRING {
		a.addError("function 'CompareStr' expects string as first argument, got %s at %s",
			str1Type.String(), callExpr.Token.Pos.String())
	}
	str2Type := a.analyzeExpression(args[1])
	if str2Type != nil && str2Type != types.STRING {
		a.addError("function 'CompareStr' expects string as second argument, got %s at %s",
			str2Type.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeAnsiCompareText analyzes the AnsiCompareText built-in function.
// AnsiCompareText takes 2 arguments: AnsiCompareText(str1, str2)
func (a *Analyzer) analyzeAnsiCompareText(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'AnsiCompareText' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	str1Type := a.analyzeExpression(args[0])
	if str1Type != nil && str1Type != types.STRING {
		a.addError("function 'AnsiCompareText' expects string as first argument, got %s at %s",
			str1Type.String(), callExpr.Token.Pos.String())
	}
	str2Type := a.analyzeExpression(args[1])
	if str2Type != nil && str2Type != types.STRING {
		a.addError("function 'AnsiCompareText' expects string as second argument, got %s at %s",
			str2Type.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeAnsiCompareStr analyzes the AnsiCompareStr built-in function.
// AnsiCompareStr takes 2 arguments: AnsiCompareStr(str1, str2)
func (a *Analyzer) analyzeAnsiCompareStr(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'AnsiCompareStr' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	str1Type := a.analyzeExpression(args[0])
	if str1Type != nil && str1Type != types.STRING {
		a.addError("function 'AnsiCompareStr' expects string as first argument, got %s at %s",
			str1Type.String(), callExpr.Token.Pos.String())
	}
	str2Type := a.analyzeExpression(args[1])
	if str2Type != nil && str2Type != types.STRING {
		a.addError("function 'AnsiCompareStr' expects string as second argument, got %s at %s",
			str2Type.String(), callExpr.Token.Pos.String())
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

// analyzeStrMatches analyzes the StrMatches built-in function.
// StrMatches takes 2 arguments: StrMatches(str, mask)
func (a *Analyzer) analyzeStrMatches(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'StrMatches' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'StrMatches' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	maskType := a.analyzeExpression(args[1])
	if maskType != nil && maskType != types.STRING {
		a.addError("function 'StrMatches' expects string as second argument, got %s at %s",
			maskType.String(), callExpr.Token.Pos.String())
	}
	return types.BOOLEAN
}
