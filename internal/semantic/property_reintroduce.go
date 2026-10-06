package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

// propertyForCompatibilityCall respects a descendant's ordinary member shadow.
func propertyForCompatibilityCall(class *types.ClassType, name string) *types.PropertyInfo {
	key := ident.Normalize(name)
	for current := class; current != nil; current = current.Parent {
		for propName, prop := range current.Properties {
			if ident.Equal(propName, name) {
				return prop
			}
		}
		if _, exists := current.Methods[key]; exists {
			return nil
		}
		if _, exists := current.Fields[key]; exists {
			return nil
		}
		if _, exists := current.ClassVars[key]; exists {
			return nil
		}
	}
	return nil
}

func (a *Analyzer) analyzePropertyCompatibilityRead(expr *ast.MethodCallExpression, class *types.ClassType) (types.Type, bool) {
	prop := propertyForCompatibilityCall(class, expr.Method.Value)
	if prop == nil || prop.IsIndexed || isFunctionPointerType(prop.Type) {
		return nil, false
	}
	read := &ast.MemberAccessExpression{BaseNode: expr.BaseNode, Object: expr.Object, Member: expr.Method}
	a.addIdentifierCaseHint(expr.Method, prop.Name)
	if !prop.IsReintroduce {
		a.addPunctuationStop(expr.ParenPos, "Not a method")
		return nil, true
	}
	// ReadPropertyExpr consumes a compatibility pair before performing the read.
	a.addHintAt(expr.ParenPos, "Property %q reintroduced a method, you should remove empty brackets () [line: %d, column: %d]", prop.Name, expr.ParenPos.Line, expr.ParenPos.Column)
	if expr.FirstArgumentToken.Type != token.RPAREN {
		a.addStructuredError(NewGenericError(expr.FirstArgumentToken.Pos, `")" expected`))
	}
	result := a.analyzeMemberAccessExpression(read)
	a.semanticInfo.SetPropertyRead(expr, read)
	return result, true
}

func (a *Analyzer) analyzeIncompleteMemberCall(_ *ast.MethodCallExpression) types.Type {
	// The parser retains the authoritative stop, including through enclosing
	// unfinished calls. Never validate a truncated argument list or run end checks.
	a.compileStopped = true
	return nil
}
