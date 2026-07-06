package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// analyzeAssociativeArrayMemberAccess handles bare member access (no parens) on
// an associative array receiver: a.Keys, a.Length, a.Count, a.Clear.
func (a *Analyzer) analyzeAssociativeArrayMemberAccess(expr *ast.MemberAccessExpression, assoc *types.AssociativeArrayType) types.Type {
	if expr == nil || expr.Member == nil {
		return nil
	}
	switch ident.Normalize(expr.Member.Value) {
	case "keys":
		return types.NewDynamicArrayType(assoc.KeyType)
	case "length", "count":
		return types.INTEGER
	case "clear":
		return types.VOID
	}
	return nil
}

// analyzeAssociativeArrayMethodCall handles method-call syntax on an associative
// array receiver: a.Keys, a.Length, a.Count, a.Clear, a.Delete(key).
func (a *Analyzer) analyzeAssociativeArrayMethodCall(expr *ast.MethodCallExpression, assoc *types.AssociativeArrayType) types.Type {
	if expr == nil || expr.Method == nil {
		return nil
	}
	switch ident.Normalize(expr.Method.Value) {
	case "keys":
		if len(expr.Arguments) != 0 {
			a.addArrayHelperNoArgs(expr)
		}
		return types.NewDynamicArrayType(assoc.KeyType)
	case "length", "count":
		if len(expr.Arguments) != 0 {
			a.addArrayHelperNoArgs(expr)
		}
		return types.INTEGER
	case "clear":
		if len(expr.Arguments) != 0 {
			a.addArrayHelperNoArgs(expr)
		}
		return types.VOID
	case "delete":
		// Delete(key) removes the entry and returns Boolean (was it present).
		if len(expr.Arguments) == 0 {
			a.addArrayHelperTooFewArgs(expr)
			return types.BOOLEAN
		}
		if len(expr.Arguments) > 1 {
			a.addArrayHelperTooManyArgs(expr)
		}
		if argType := a.analyzeExpressionWithExpectedType(expr.Arguments[0], assoc.KeyType); argType != nil &&
			!types.GetUnderlyingType(argType).Equals(types.VARIANT) && !a.canAssign(argType, assoc.KeyType) {
			a.addArrayHelperParamTypeExpectedAt(expr.Arguments[0].Pos(), assoc.KeyType, argType)
		}
		return types.BOOLEAN
	}
	return nil
}
