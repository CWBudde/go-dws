package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

type inheritedIndexRecovery struct {
	expression *ast.InheritedExpression
	result     types.Type
}

// analyzeInheritedIndexedPropertyRead keeps the parent descriptor relative to
// the lexical method owner. Self and virtual getter dispatch remain dynamic.
func (a *Analyzer) analyzeInheritedIndexedPropertyRead(expr *ast.IndexExpression, inherited *ast.InheritedExpression, nodes []*ast.IndexExpression) (types.Type, bool) {
	parent, prop := a.inheritedIndexedProperty(inherited)
	if prop == nil {
		return nil, false
	}
	if a.inheritedIndexRecovery != nil && a.inheritedIndexRecovery.expression == inherited {
		return nil, false
	}
	if a.isInheritedIndexWriteTarget(expr, inherited) {
		return nil, false
	}
	// ReadInherited does not pass through the ordinary member case checker.
	a.warnDeprecatedPropertyUsage(prop, inherited.Method.Token.Pos)
	pos := inherited.Method.Token.Pos
	if prop.IsReintroduce {
		pos = inherited.AfterNamePos
	}
	if inherited.IsCall {
		if !prop.IsReintroduce {
			if prop.ReadKind == types.PropAccessNone {
				a.addStructuredError(NewWriteOnlyPropertyError(pos, prop.Name))
			} else {
				a.addMemberCallCountError(&types.FunctionType{Parameters: a.getIndexedPropertyParamTypes(prop, parent)}, 0, pos)
			}
			a.addPunctuationStop(inherited.ParenPos, "Not a method")
			return nil, true
		}
		a.addPropertyBracketHint(prop, inherited.ParenPos)
		pos = inherited.End()
		pos.Column--
		pos.Offset-- // The consumed compatibility ')', before declared indices.
	}
	if prop.ReadKind == types.PropAccessNone {
		a.addStructuredError(NewWriteOnlyPropertyError(pos, prop.Name))
		if expr == nil {
			return prop.Type, true
		}
		previous := a.inheritedIndexRecovery
		a.inheritedIndexRecovery = &inheritedIndexRecovery{expression: inherited, result: prop.Type}
		defer func() { a.inheritedIndexRecovery = previous }()
		return a.analyzeIndexExpression(expr), true
	}
	indices := indexedPropertyArguments(nodes)
	if !a.checkIndexedCompatibilityArguments(parent, prop, indices, pos, a.inClassMethod, expr) {
		return nil, true
	}
	read := &ast.MemberAccessExpression{
		BaseNode: inherited.BaseNode, Member: inherited.Method,
		Object: &ast.SelfExpression{BaseNode: inherited.BaseNode},
	}
	binding := &ast.InheritedPropertyReadBinding{
		Read: read, Property: prop, Owner: compatibilityPropertyOwner(parent, prop),
	}
	if expr == nil {
		a.semanticInfo.SetInheritedPropertyRead(inherited, binding)
	} else {
		a.semanticInfo.SetIndexedPropertyRead(expr, &ast.IndexedPropertyReadBinding{Read: binding, Indices: indices})
	}
	return prop.Type, true
}

func (a *Analyzer) isInheritedIndexWriteTarget(expr *ast.IndexExpression, inherited *ast.InheritedExpression) bool {
	for target := a.indexedAssignmentTarget; target != nil; {
		if target == expr || (expr == nil && target.Left == inherited) {
			return true
		}
		inner, ok := target.Left.(*ast.IndexExpression)
		if !ok {
			break
		}
		target = inner
	}
	return false
}

func (a *Analyzer) inheritedIndexedProperty(inherited *ast.InheritedExpression) (*types.ClassType, *types.PropertyInfo) {
	if inherited.Method == nil || len(inherited.Arguments) != 0 || a.currentHelperType != nil || a.currentClass == nil || a.currentClass.Parent == nil {
		return nil, nil
	}
	for scope := a.symbols; scope != nil; scope = scope.outer {
		if scope.classMethodOwner != nil {
			if scope.classMethodStatic {
				return nil, nil
			}
			break
		}
	}
	parent := a.currentClass.Parent
	prop := propertyForCompatibilityCall(parent, inherited.Method.Value)
	if prop == nil || !prop.IsIndexed || isFunctionPointerType(prop.Type) {
		return nil, nil
	}
	return parent, prop
}
