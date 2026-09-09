package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	pkgident "github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Type Conversion Built-in Function Analysis
// ============================================================================

// analyzeFloatToStrF analyzes the FloatToStrF built-in function.
// FloatToStrF takes 2 or 4 arguments (value, format, [precision, digits]) and returns a string.
// The precision and digits parameters are optional.
func (a *Analyzer) analyzeFloatToStrF(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 && len(args) != 4 {
		a.addError("function 'FloatToStrF' expects 2 or 4 arguments (value, format, [precision, digits]), got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// First argument: Float value
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'FloatToStrF' expects Float as first argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	// Second argument: format (Integer)
	if len(args) > 1 {
		argType := a.analyzeExpression(args[1])
		if argType != nil && argType != types.INTEGER {
			a.addError("function 'FloatToStrF' expects Integer as second argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	// Third argument: precision (Integer) - optional
	if len(args) > 2 {
		argType := a.analyzeExpression(args[2])
		if argType != nil && argType != types.INTEGER {
			a.addError("function 'FloatToStrF' expects Integer as third argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	// Fourth argument: digits (Integer) - optional
	if len(args) > 3 {
		argType := a.analyzeExpression(args[3])
		if argType != nil && argType != types.INTEGER {
			a.addError("function 'FloatToStrF' expects Integer as fourth argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.STRING
}

// analyzeDefault analyzes the Default built-in function.
// Default takes one argument (a type identifier) and returns the default value for that type.
// Default(Integer) returns 0, Default(String) returns "", Default(Boolean) returns False, etc.
func (a *Analyzer) analyzeDefault(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Default' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.NIL
	}

	// The argument should be a type identifier
	// For now, we'll handle it as an identifier and look up the type
	ident, ok := args[0].(*ast.Identifier)
	if !ok {
		a.addError("function 'Default' expects a type name as argument at %s",
			callExpr.Token.Pos.String())
		return types.NIL
	}

	// Look up the type by name
	typeName := ident.Value

	// Return the appropriate type based on the type name
	switch typeName {
	case "Integer", "Int64", "Byte", "Word", "Cardinal", "SmallInt", "ShortInt", "LongWord":
		return types.INTEGER
	case "Float", "Double", "Single", "Extended", "Currency":
		return types.FLOAT
	case "String", "UnicodeString", "AnsiString":
		return types.STRING
	case "Boolean":
		return types.BOOLEAN
	case "Variant":
		return types.VARIANT
	default:
		// Check if it's a valid type in the symbol table
		// This could be a class, record, enum, or other custom type
		lowerName := pkgident.Normalize(typeName)
		if _, exists := a.symbols.Resolve(lowerName); exists {
			// Valid custom type - all default to nil/zero values
			// Return VARIANT as a safe fallback type for analysis
			return types.VARIANT
		}
		// Unknown type - report error instead of silently succeeding
		a.addError("function 'Default' received unknown type '%s' at %s",
			typeName, callExpr.Token.Pos.String())
		return types.NIL
	}
}

// analyzeTryStrToInt analyzes the TryStrToInt built-in function.
// TryStrToInt takes two or three arguments (string, [base,] var value) and returns a boolean.
// The base parameter is optional and defaults to 10.
func (a *Analyzer) analyzeTryStrToInt(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) < 2 || len(args) > 3 {
		a.addError("function 'TryStrToInt' expects 2 or 3 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}
	// Analyze first argument (string)
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'TryStrToInt' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	if len(args) == 2 {
		// TryStrToInt(str, var value) - base defaults to 10
		// Second argument should be a var parameter (identifier)
		// We just verify it exists and is an integer
		if ident, ok := args[1].(*ast.Identifier); ok {
			// Check if variable exists in symbol table
			if sym, exists := a.symbols.Resolve(ident.Value); exists {
				if sym.Type != types.INTEGER {
					a.addError("function 'TryStrToInt' expects var Integer parameter, got %s at %s",
						sym.Type.String(), callExpr.Token.Pos.String())
				}
			}
		}
	} else {
		// TryStrToInt(str, base, var value)
		// Analyze second argument (base) - must be Integer
		baseType := a.analyzeExpression(args[1])
		if baseType != nil && baseType != types.INTEGER {
			// Check if it's a subrange type with Integer base
			if subrange, ok := baseType.(*types.SubrangeType); ok {
				if subrange.BaseType != types.INTEGER {
					a.addError("function 'TryStrToInt' expects Integer as second argument (base), got %s at %s",
						baseType.String(), callExpr.Token.Pos.String())
				}
			} else {
				a.addError("function 'TryStrToInt' expects Integer as second argument (base), got %s at %s",
					baseType.String(), callExpr.Token.Pos.String())
			}
		}
		// Third argument should be a var parameter (identifier)
		if ident, ok := args[2].(*ast.Identifier); ok {
			if sym, exists := a.symbols.Resolve(ident.Value); exists {
				if sym.Type != types.INTEGER {
					a.addError("function 'TryStrToInt' expects var Integer parameter, got %s at %s",
						sym.Type.String(), callExpr.Token.Pos.String())
				}
			}
		}
	}
	return types.BOOLEAN
}

// analyzeTryStrToFloat analyzes the TryStrToFloat built-in function.
// TryStrToFloat takes two arguments (string, var value) and returns a boolean.
func (a *Analyzer) analyzeTryStrToFloat(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'TryStrToFloat' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}
	// Analyze first argument (string)
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'TryStrToFloat' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	// Second argument should be a var parameter (identifier)
	// We just verify it exists and is a float
	if ident, ok := args[1].(*ast.Identifier); ok {
		// Check if variable exists in symbol table
		if sym, exists := a.symbols.Resolve(ident.Value); exists {
			if sym.Type != types.FLOAT {
				a.addError("function 'TryStrToFloat' expects var Float parameter, got %s at %s",
					sym.Type.String(), callExpr.Token.Pos.String())
			}
		}
	}
	return types.BOOLEAN
}
