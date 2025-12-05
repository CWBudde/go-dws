package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Math Utility Built-in Function Analysis
// ============================================================================
// This file contains analyzers for utility math functions:
// - Inc, Dec, Succ, Pred
// - Random, RandomInt, Randomize, SetRandSeed, RandSeed, RandG
// - Assigned, Swap

// analyzeInc analyzes the Inc built-in procedure.
// Inc takes 1-2 arguments: variable and optional delta.
func (a *Analyzer) analyzeInc(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) < 1 || len(args) > 2 {
		a.addError("function 'Inc' expects 1-2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.VOID
	}
	if !a.isLValue(args[0]) {
		a.addError("function 'Inc' first argument must be a variable (identifier, array element, or field) at %s",
			callExpr.Token.Pos.String())
		return types.VOID
	}

	varType := a.analyzeExpression(args[0])
	if varType != nil {
		if varType != types.INTEGER {
			if _, isEnum := varType.(*types.EnumType); !isEnum {
				a.addError("function 'Inc' expects Integer or Enum variable, got %s at %s",
					varType.String(), callExpr.Token.Pos.String())
			}
		}
	}
	if len(args) == 2 {
		deltaType := a.analyzeExpression(args[1])
		if deltaType != nil && deltaType != types.INTEGER {
			a.addError("function 'Inc' delta must be Integer, got %s at %s",
				deltaType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeDec analyzes the Dec built-in function.
// Dec takes 1-2 arguments: variable and optional delta.
// Returns the decremented value (like prefix -- in C).
func (a *Analyzer) analyzeDec(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) < 1 || len(args) > 2 {
		a.addError("function 'Dec' expects 1-2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.VOID
	}
	if !a.isLValue(args[0]) {
		a.addError("function 'Dec' first argument must be a variable (identifier, array element, or field) at %s",
			callExpr.Token.Pos.String())
	} else {
		varType := a.analyzeExpression(args[0])
		if varType != nil {
			if varType != types.INTEGER {
				if _, isEnum := varType.(*types.EnumType); !isEnum {
					a.addError("function 'Dec' expects Integer or Enum variable, got %s at %s",
						varType.String(), callExpr.Token.Pos.String())
				}
			}
		}
	}
	if len(args) == 2 {
		deltaType := a.analyzeExpression(args[1])
		if deltaType != nil && deltaType != types.INTEGER {
			a.addError("function 'Dec' delta must be Integer, got %s at %s",
				deltaType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeSucc analyzes the Succ built-in function.
// Succ takes 1 argument: ordinal value and returns the successor.
func (a *Analyzer) analyzeSucc(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Succ' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType == types.INTEGER {
			return types.INTEGER
		}
		if enumType, isEnum := argType.(*types.EnumType); isEnum {
			return enumType
		}
		a.addError("function 'Succ' expects Integer or Enum, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzePred analyzes the Pred built-in function.
// Pred takes 1 argument: ordinal value and returns the predecessor.
func (a *Analyzer) analyzePred(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Pred' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if argType == types.INTEGER {
			return types.INTEGER
		}
		if enumType, isEnum := argType.(*types.EnumType); isEnum {
			return enumType
		}
		a.addError("function 'Pred' expects Integer or Enum, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeAssigned analyzes the Assigned built-in function.
// Assigned takes 1 argument and checks if a pointer/object/variant is nil.
// Returns Boolean.
func (a *Analyzer) analyzeAssigned(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Assigned' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.BOOLEAN
	}
	a.analyzeExpression(args[0])
	return types.BOOLEAN
}

// analyzeSwap analyzes the Swap built-in function.
// Swap takes 2 var arguments and swaps their values.
func (a *Analyzer) analyzeSwap(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'Swap' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return nil
	}

	for i, arg := range args {
		if _, ok := arg.(*ast.Identifier); !ok {
			a.addError("function 'Swap' argument %d must be a variable at %s",
				i+1, callExpr.Token.Pos.String())
		}
	}

	type1 := a.analyzeExpression(args[0])
	type2 := a.analyzeExpression(args[1])

	if type1 != nil && type2 != nil {
		if !type1.Equals(type2) {
			a.addError("function 'Swap' arguments must have compatible types, got %s and %s at %s",
				type1.String(), type2.String(), callExpr.Token.Pos.String())
		}
	}

	return nil
}

// analyzeRandom analyzes the Random built-in function.
// Random takes no arguments and always returns Float.
func (a *Analyzer) analyzeRandom(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 0 {
		a.addError("function 'Random' expects no arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	return types.FLOAT
}

// analyzeRandomInt analyzes the RandomInt built-in function.
// RandomInt takes one Integer argument and returns random Integer in [0, max).
func (a *Analyzer) analyzeRandomInt(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'RandomInt' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.INTEGER {
		a.addError("function 'RandomInt' expects Integer argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeRandomize analyzes the Randomize built-in procedure.
// Randomize takes no arguments and returns nothing (nil/void).
func (a *Analyzer) analyzeRandomize(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 0 {
		a.addError("function 'Randomize' expects no arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	return nil
}

// analyzeSetRandSeed analyzes the SetRandSeed built-in function.
// SetRandSeed takes 1 Integer argument and sets the random seed.
func (a *Analyzer) analyzeSetRandSeed(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'SetRandSeed' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return nil
	}
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.INTEGER {
		a.addError("function 'SetRandSeed' expects Integer, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return nil
}

// analyzeRandSeed analyzes the RandSeed built-in function.
// RandSeed: Integer
func (a *Analyzer) analyzeRandSeed(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 0 {
		a.addError("function 'RandSeed' expects no arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	return types.INTEGER
}

// analyzeRandG analyzes the RandG built-in function.
// RandG: Float
func (a *Analyzer) analyzeRandG(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 0 {
		a.addError("function 'RandG' expects no arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.FLOAT
	}
	return types.FLOAT
}
