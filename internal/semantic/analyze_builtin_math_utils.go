package semantic

import (
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
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
		underlying := types.GetUnderlyingType(argType)
		if _, isEnum := underlying.(*types.EnumType); isEnum {
			result = argType
		} else if underlying != types.INTEGER {
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
// Pred, looking through aliases. A Variant delta is converted to Integer at
// runtime; an unresolved type has already been reported.
func isOrdinalDeltaType(t types.Type) bool {
	if t == nil {
		return true
	}
	underlying := types.GetUnderlyingType(t)
	return underlying == types.INTEGER || underlying == types.VARIANT
}

// analyzeSwap analyzes the Swap intrinsic. Upstream checks each data operand
// before comparing types, so an invalid operand suppresses the enclosing pair.
func (a *Analyzer) analyzeSwap(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) == 0 || len(args) > 2 {
		a.addError("function 'Swap' expects 2 arguments, got %d at %s", len(args), callExpr.Token.Pos.String())
		return types.VOID
	}

	firstType, firstValid := a.analyzeSwapArgument(args[0])
	if len(args) == 1 {
		pos := callExpr.Token.Pos
		if callExpr.Token.Type != lexer.RPAREN {
			pos = callExpr.End()
			pos.Column--
			pos.Offset--
		}
		a.addError("Syntax Error: \",\" expected [line: %d, column: %d]", pos.Line, pos.Column)
		return types.VOID
	}
	secondType, secondValid := a.analyzeSwapArgument(args[1])
	if firstValid && secondValid && !types.IsIdentical(types.GetUnderlyingType(firstType), types.GetUnderlyingType(secondType)) {
		a.addStructuredError(NewIncompatibleTypesPairError(callExpr.Function.Pos(),
			semanticTypeNameForDiagnostic(firstType), semanticTypeNameForDiagnostic(secondType)))
	}

	return types.VOID
}

func (a *Analyzer) analyzeSwapArgument(arg ast.Expression) (types.Type, bool) {
	mark := len(a.errors)
	argType := a.analyzeExpression(arg)
	if argType == nil || a.errorsSince(mark) {
		return argType, false
	}
	valid := a.isSwapDataArgument(arg, argType)
	if !valid {
		a.addError("Variable expected at %s", arg.Pos().String())
		return argType, false
	}
	a.markVarArgumentWritten(arg)
	return argType, true
}

// isSwapDataArgument follows the intrinsic's read-side data-expression check:
// a field reader remains data even without a property writer, while a getter
// method is a call result and cannot supply storage to swap.
func (a *Analyzer) isSwapDataArgument(arg ast.Expression, argType types.Type) bool {
	switch expr := arg.(type) {
	case *ast.Identifier:
		sym, found := a.symbols.Resolve(expr.Value)
		if !found || sym.ReadOnly || sym.IsConst {
			return false
		}
		_, routine := sym.Type.(*types.FunctionType)
		return !routine
	case *ast.AddressOfExpression:
		// Upstream folds explicit routine references to data expressions.
		return types.IsPointerType(argType)
	case *ast.IndexExpression:
		baseType := a.semanticInfo.GetResolvedType(expr.Left)
		if array, ok := types.GetUnderlyingType(baseType).(*types.ArrayType); ok && array.IsStatic() {
			return a.isSwapDataArgument(expr.Left, baseType)
		}
		if types.GetUnderlyingType(baseType) == types.STRING {
			return a.isSwapDataArgument(expr.Left, baseType)
		}
		return true
	case *ast.MemberAccessExpression:
		return a.isSwapMemberData(expr)
	}
	return false
}

func (a *Analyzer) isSwapMemberData(expr *ast.MemberAccessExpression) bool {
	baseType := types.GetUnderlyingType(a.semanticInfo.GetResolvedType(expr.Object))
	if meta, ok := baseType.(*types.ClassOfType); ok {
		baseType = meta.ClassType
	}
	if meta, ok := baseType.(*types.RecordMetaType); ok {
		baseType = meta.RecordType
	}
	name := ident.Normalize(expr.Member.Value)
	switch receiver := baseType.(type) {
	case *types.ClassType:
		if _, constant := receiver.GetConstant(name); constant {
			return false
		}
		if property, found := receiver.GetProperty(name); found {
			return property.ReadKind == types.PropAccessField
		}
		_, field := receiver.GetField(name)
		_, classVar := receiver.GetClassVar(name)
		return field || classVar
	case *types.RecordType:
		if property := receiver.GetProperty(name); property != nil {
			return property.ReadKind == types.PropAccessField && a.isSwapDataArgument(expr.Object, receiver)
		}
		if _, classVar := receiver.ClassVars[name]; classVar {
			return true
		}
		return receiver.GetFieldType(name) != nil && a.isSwapDataArgument(expr.Object, receiver)
	}
	return false
}
