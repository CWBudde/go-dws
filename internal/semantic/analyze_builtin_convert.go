package semantic

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Type Conversion Built-in Function Analysis
// ============================================================================

// analyzeIntToStr analyzes the IntToStr built-in function.
// IntToStr takes one or two arguments (value, [base]) and returns a string.
// The base parameter is optional and defaults to 10.
func (a *Analyzer) analyzeIntToStr(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) < 1 || len(args) > 2 {
		a.addError("function 'IntToStr' expects 1 or 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze the first argument and verify it's Integer or a subrange of Integer
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.INTEGER {
		// Check if it's a subrange type with Integer base
		if subrange, ok := argType.(*types.SubrangeType); ok {
			if subrange.BaseType != types.INTEGER {
				a.addError("function 'IntToStr' expects Integer as first argument, got %s at %s",
					argType.String(), callExpr.Token.Pos.String())
			}
		} else {
			a.addError("function 'IntToStr' expects Integer as first argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	// Analyze optional second argument (base) - must be Integer
	if len(args) == 2 {
		baseType := a.analyzeExpression(args[1])
		if baseType != nil && baseType != types.INTEGER {
			// Check if it's a subrange type with Integer base
			if subrange, ok := baseType.(*types.SubrangeType); ok {
				if subrange.BaseType != types.INTEGER {
					a.addError("function 'IntToStr' expects Integer as second argument (base), got %s at %s",
						baseType.String(), callExpr.Token.Pos.String())
				}
			} else {
				a.addError("function 'IntToStr' expects Integer as second argument (base), got %s at %s",
					baseType.String(), callExpr.Token.Pos.String())
			}
		}
	}
	return types.STRING
}

// analyzeIntToBin analyzes the IntToBin built-in function.
// IntToBin takes one or two integer arguments (value, [digits]) and returns a string.
// The digits parameter is optional and defaults to the minimum required digits.
func (a *Analyzer) analyzeIntToBin(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) < 1 || len(args) > 2 {
		a.addError("function 'IntToBin' expects 1 or 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze first argument (value) - must be Integer or subrange of Integer
	argType1 := a.analyzeExpression(args[0])
	if argType1 != nil && argType1 != types.INTEGER {
		// Check if it's a subrange type with Integer base
		if subrange, ok := argType1.(*types.SubrangeType); ok {
			if subrange.BaseType != types.INTEGER {
				a.addError("function 'IntToBin' expects Integer as first argument, got %s at %s",
					argType1.String(), callExpr.Token.Pos.String())
			}
		} else {
			a.addError("function 'IntToBin' expects Integer as first argument, got %s at %s",
				argType1.String(), callExpr.Token.Pos.String())
		}
	}
	// Analyze optional second argument (digits) - must be Integer
	if len(args) >= 2 {
		argType2 := a.analyzeExpression(args[1])
		if argType2 != nil && argType2 != types.INTEGER {
			// Check if it's a subrange type with Integer base
			if subrange, ok := argType2.(*types.SubrangeType); ok {
				if subrange.BaseType != types.INTEGER {
					a.addError("function 'IntToBin' expects Integer as second argument, got %s at %s",
						argType2.String(), callExpr.Token.Pos.String())
				}
			} else {
				a.addError("function 'IntToBin' expects Integer as second argument, got %s at %s",
					argType2.String(), callExpr.Token.Pos.String())
			}
		}
	}
	return types.STRING
}

// analyzeIntToHex analyzes the IntToHex built-in function.
// IntToHex takes one or two integer arguments (value, [digits]) and returns a string.
// The digits parameter is optional and defaults to the minimum required digits.
func (a *Analyzer) analyzeIntToHex(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) < 1 || len(args) > 2 {
		a.addError("function 'IntToHex' expects 1 or 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze first argument (value) - must be Integer or subrange of Integer
	argType1 := a.analyzeExpression(args[0])
	if argType1 != nil && argType1 != types.INTEGER {
		// Check if it's a subrange type with Integer base
		if subrange, ok := argType1.(*types.SubrangeType); ok {
			if subrange.BaseType != types.INTEGER {
				a.addError("function 'IntToHex' expects Integer as first argument, got %s at %s",
					argType1.String(), callExpr.Token.Pos.String())
			}
		} else {
			a.addError("function 'IntToHex' expects Integer as first argument, got %s at %s",
				argType1.String(), callExpr.Token.Pos.String())
		}
	}
	// Analyze optional second argument (digits) - must be Integer
	if len(args) >= 2 {
		argType2 := a.analyzeExpression(args[1])
		if argType2 != nil && argType2 != types.INTEGER {
			// Check if it's a subrange type with Integer base
			if subrange, ok := argType2.(*types.SubrangeType); ok {
				if subrange.BaseType != types.INTEGER {
					a.addError("function 'IntToHex' expects Integer as second argument, got %s at %s",
						argType2.String(), callExpr.Token.Pos.String())
				}
			} else {
				a.addError("function 'IntToHex' expects Integer as second argument, got %s at %s",
					argType2.String(), callExpr.Token.Pos.String())
			}
		}
	}
	return types.STRING
}

// analyzeStrToInt analyzes the StrToInt built-in function.
// StrToInt takes one or two arguments (string, [base]) and returns an integer.
// The base parameter is optional and defaults to 10.
func (a *Analyzer) analyzeStrToInt(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) < 1 || len(args) > 2 {
		a.addError("function 'StrToInt' expects 1 or 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze the first argument and verify it's String
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.STRING {
		a.addError("function 'StrToInt' expects String as first argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	// Analyze optional second argument (base) - must be Integer
	if len(args) == 2 {
		baseType := a.analyzeExpression(args[1])
		if baseType != nil && baseType != types.INTEGER {
			// Check if it's a subrange type with Integer base
			if subrange, ok := baseType.(*types.SubrangeType); ok {
				if subrange.BaseType != types.INTEGER {
					a.addError("function 'StrToInt' expects Integer as second argument (base), got %s at %s",
						baseType.String(), callExpr.Token.Pos.String())
				}
			} else {
				a.addError("function 'StrToInt' expects Integer as second argument (base), got %s at %s",
					baseType.String(), callExpr.Token.Pos.String())
			}
		}
	}
	return types.INTEGER
}

// analyzeBoolToStr analyzes the BoolToStr built-in function.
// BoolToStr takes one boolean argument and returns a string.
func (a *Analyzer) analyzeBoolToStr(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'BoolToStr' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze the argument and verify it's Boolean
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.BOOLEAN {
		a.addError("function 'BoolToStr' expects Boolean as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeStrToFloat analyzes the StrToFloat built-in function.
// StrToFloat takes one string argument and returns a float.
func (a *Analyzer) analyzeStrToFloat(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'StrToFloat' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	// Analyze the argument and verify it's String
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.STRING {
		a.addError("function 'StrToFloat' expects String as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.FLOAT
}

// analyzeVarToStr analyzes the VarToStr built-in function.
// VarToStr takes one variant argument and returns a string.
func (a *Analyzer) analyzeVarToStr(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'VarToStr' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze the argument (accepts any type - variant)
	a.analyzeExpression(args[0])
	return types.STRING
}

// analyzeFloatToStr analyzes the FloatToStr built-in function.
// FloatToStr takes one float argument and returns a string.
func (a *Analyzer) analyzeFloatToStr(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'FloatToStr' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze the argument and verify it's Float
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.FLOAT {
		a.addError("function 'FloatToStr' expects Float as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

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

// analyzeStrToBool analyzes the StrToBool built-in function.
// StrToBool takes one string argument and returns a boolean.
func (a *Analyzer) analyzeStrToBool(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'StrToBool' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}
	// Analyze the argument and verify it's String
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.STRING {
		a.addError("function 'StrToBool' expects String as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.BOOLEAN
}

// analyzeChr analyzes the Chr built-in function.
// Chr takes one integer argument (character code) and returns a string.
func (a *Analyzer) analyzeChr(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Chr' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze the argument and verify it's Integer
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.INTEGER {
		a.addError("function 'Chr' expects Integer as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeStrToIntDef analyzes the StrToIntDef built-in function.
// StrToIntDef takes two or three arguments (string, default, [base]) and returns an integer.
// The base parameter is optional and defaults to 10.
func (a *Analyzer) analyzeStrToIntDef(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) < 2 || len(args) > 3 {
		a.addError("function 'StrToIntDef' expects 2 or 3 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze first argument (string)
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'StrToIntDef' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (default integer value)
	defaultType := a.analyzeExpression(args[1])
	if defaultType != nil && defaultType != types.INTEGER {
		a.addError("function 'StrToIntDef' expects integer as second argument, got %s at %s",
			defaultType.String(), callExpr.Token.Pos.String())
	}
	// Analyze optional third argument (base) - must be Integer
	if len(args) == 3 {
		baseType := a.analyzeExpression(args[2])
		if baseType != nil && baseType != types.INTEGER {
			// Check if it's a subrange type with Integer base
			if subrange, ok := baseType.(*types.SubrangeType); ok {
				if subrange.BaseType != types.INTEGER {
					a.addError("function 'StrToIntDef' expects Integer as third argument (base), got %s at %s",
						baseType.String(), callExpr.Token.Pos.String())
				}
			} else {
				a.addError("function 'StrToIntDef' expects Integer as third argument (base), got %s at %s",
					baseType.String(), callExpr.Token.Pos.String())
			}
		}
	}
	return types.INTEGER
}

// analyzeStrToFloatDef analyzes the StrToFloatDef built-in function.
// StrToFloatDef takes a string and a default float value, returning a float.
func (a *Analyzer) analyzeStrToFloatDef(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'StrToFloatDef' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	// Analyze first argument (string)
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'StrToFloatDef' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	// Analyze second argument (default float value)
	defaultType := a.analyzeExpression(args[1])
	if defaultType != nil && defaultType != types.FLOAT {
		a.addError("function 'StrToFloatDef' expects float as second argument, got %s at %s",
			defaultType.String(), callExpr.Token.Pos.String())
	}
	return types.FLOAT
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

// analyzeHexToInt analyzes the HexToInt built-in function.
// HexToInt takes one string argument (hexadecimal) and returns an integer.
func (a *Analyzer) analyzeHexToInt(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'HexToInt' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze the argument and verify it's String
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.STRING {
		a.addError("function 'HexToInt' expects String as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeBinToInt analyzes the BinToInt built-in function.
// BinToInt takes one string argument (binary) and returns an integer.
func (a *Analyzer) analyzeBinToInt(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'BinToInt' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze the argument and verify it's String
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.STRING {
		a.addError("function 'BinToInt' expects String as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeVarToIntDef analyzes the VarToIntDef built-in function.
// VarToIntDef takes any variant value and a default integer, returning an integer.
func (a *Analyzer) analyzeVarToIntDef(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'VarToIntDef' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze first argument (variant - any type)
	a.analyzeExpression(args[0])
	// Analyze second argument (default integer value)
	defaultType := a.analyzeExpression(args[1])
	if defaultType != nil && defaultType != types.INTEGER {
		a.addError("function 'VarToIntDef' expects integer as second argument, got %s at %s",
			defaultType.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeVarToFloatDef analyzes the VarToFloatDef built-in function.
// VarToFloatDef takes any variant value and a default float, returning a float.
func (a *Analyzer) analyzeVarToFloatDef(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'VarToFloatDef' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	// Analyze first argument (variant - any type)
	a.analyzeExpression(args[0])
	// Analyze second argument (default float value)
	defaultType := a.analyzeExpression(args[1])
	if defaultType != nil && defaultType != types.FLOAT {
		a.addError("function 'VarToFloatDef' expects float as second argument, got %s at %s",
			defaultType.String(), callExpr.Token.Pos.String())
	}
	return types.FLOAT
}
