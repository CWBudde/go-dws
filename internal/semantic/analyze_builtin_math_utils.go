package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Math Utility Built-in Function Analysis
// ============================================================================

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
		if !isOrdinalDeltaType(deltaType) {
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
		if !isOrdinalDeltaType(deltaType) {
			a.addError("function 'Dec' delta must be Integer, got %s at %s",
				deltaType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeSucc analyzes the Succ built-in function.
// Succ takes an ordinal value and an optional Integer or Variant delta, and
// returns the successor.
func (a *Analyzer) analyzeSucc(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	return a.analyzeOrdinalStep("Succ", args, callExpr)
}

// analyzePred analyzes the Pred built-in function.
// Pred takes an ordinal value and an optional Integer or Variant delta, and
// returns the predecessor.
func (a *Analyzer) analyzePred(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	return a.analyzeOrdinalStep("Pred", args, callExpr)
}

// analyzeOrdinalStep implements the shared Succ/Pred analysis. The result has
// the ordinal argument's type: Integer or the argument's enumeration.
func (a *Analyzer) analyzeOrdinalStep(name string, args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) < 1 || len(args) > 2 {
		a.addError("function '%s' expects 1-2 arguments, got %d at %s",
			name, len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	result := types.Type(types.INTEGER)
	argType := a.analyzeExpression(args[0])
	if argType != nil {
		if enumType, isEnum := argType.(*types.EnumType); isEnum {
			result = enumType
		} else if argType != types.INTEGER {
			a.addError("function '%s' expects Integer or Enum, got %s at %s",
				name, argType.String(), callExpr.Token.Pos.String())
		}
	}
	if len(args) == 2 {
		deltaType := a.analyzeExpression(args[1])
		if !isOrdinalDeltaType(deltaType) {
			a.addError("function '%s' delta must be Integer, got %s at %s",
				name, deltaType.String(), callExpr.Token.Pos.String())
		}
	}
	return result
}

// isOrdinalDeltaType reports whether t may be the delta of Inc, Dec, Succ or
// Pred. A Variant delta is converted to Integer at runtime; an unresolved type
// has already been reported.
func isOrdinalDeltaType(t types.Type) bool {
	return t == nil || t == types.INTEGER || t == types.VARIANT
}

// analyzeSwap analyzes the Swap built-in function.
// Swap takes 2 var arguments and swaps their values.
//
//nolint:unparam // returns nil for procedures (void functions)
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
