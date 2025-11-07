package semantic

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// JSON Built-in Function Analysis
// ============================================================================

// analyzeParseJSON analyzes the ParseJSON built-in function.
// ParseJSON takes one string argument and returns a Variant containing a JSONValue.
func (a *Analyzer) analyzeParseJSON(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'ParseJSON' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.VARIANT
	}
	// Analyze the argument (should be a string)
	argType := a.analyzeExpression(args[0])
	if argType != nil && !argType.Equals(types.STRING) {
		a.addError("ParseJSON expects String argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	// Returns Variant containing a JSONValue
	return types.VARIANT
}

// analyzeToJSON analyzes the ToJSON built-in function.
// ToJSON takes one argument of any type and returns a string.
func (a *Analyzer) analyzeToJSON(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'ToJSON' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze the argument (can be any type)
	a.analyzeExpression(args[0])
	// Returns String
	return types.STRING
}

// analyzeToJSONFormatted analyzes the ToJSONFormatted built-in function.
// ToJSONFormatted takes two arguments (value, indent) and returns a string.
func (a *Analyzer) analyzeToJSONFormatted(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'ToJSONFormatted' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze first argument (can be any type)
	a.analyzeExpression(args[0])
	// Analyze second argument (should be Integer)
	indentType := a.analyzeExpression(args[1])
	if indentType != nil && !indentType.Equals(types.INTEGER) {
		a.addError("ToJSONFormatted expects Integer as second argument, got %s at %s",
			indentType.String(), callExpr.Token.Pos.String())
	}
	// Returns String
	return types.STRING
}

// analyzeJSONHasField analyzes the JSONHasField built-in function.
// JSONHasField takes two arguments (json object, field name) and returns a boolean.
func (a *Analyzer) analyzeJSONHasField(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'JSONHasField' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}
	// Analyze both arguments (object, field name)
	a.analyzeExpression(args[0])
	fieldType := a.analyzeExpression(args[1])
	// Second argument must be String
	if fieldType != nil && !fieldType.Equals(types.STRING) {
		a.addError("JSONHasField expects String as second argument, got %s at %s",
			fieldType.String(), callExpr.Token.Pos.String())
	}
	// Returns Boolean
	return types.BOOLEAN
}

// analyzeJSONKeys analyzes the JSONKeys built-in function.
// JSONKeys takes one argument (json object) and returns an array of strings.
func (a *Analyzer) analyzeJSONKeys(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'JSONKeys' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.NewDynamicArrayType(types.STRING)
	}
	// Analyze argument
	a.analyzeExpression(args[0])
	// Returns array of String
	return types.NewDynamicArrayType(types.STRING)
}

// analyzeJSONValues analyzes the JSONValues built-in function.
// JSONValues takes one argument (json object) and returns an array of variants.
func (a *Analyzer) analyzeJSONValues(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'JSONValues' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.NewDynamicArrayType(types.VARIANT)
	}
	// Analyze argument
	a.analyzeExpression(args[0])
	// Returns array of Variant
	return types.NewDynamicArrayType(types.VARIANT)
}

// analyzeJSONLength analyzes the JSONLength built-in function.
// JSONLength takes one argument (json object/array) and returns an integer.
func (a *Analyzer) analyzeJSONLength(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'JSONLength' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze argument
	a.analyzeExpression(args[0])
	// Returns Integer
	return types.INTEGER
}
