package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// analyzeImplicitIndexedCompatibilityRead binds a directly named empty call
// followed by one declared index group to the lexical method owner's property.
// Self remains dynamic for virtual dispatch; local and ordinary member shadows
// retain their existing call paths.
func (a *Analyzer) analyzeImplicitIndexedCompatibilityRead(expr *ast.IndexExpression, call *ast.CallExpression, nodes []*ast.IndexExpression) (types.Type, bool) {
	name, prop, scope := a.implicitIndexedCompatibilityProperty(call)
	if prop == nil {
		return nil, false
	}
	if scope.classMethodStatic {
		a.addIdentifierCaseHint(name, prop.Name)
		// ReadName rejects the absent Self before consuming compatibility '('.
		a.addPunctuationStop(name.Token.Pos, "Object reference needed to read/write an object field")
		return nil, true
	}
	if !prop.IsReintroduce {
		a.analyzeOrdinaryIndexedPropertyCall(&ast.MethodCallExpression{
			BaseNode: call.BaseNode, Method: name, ParenPos: call.ParenPos,
		}, prop)
		return nil, true
	}
	a.addIdentifierCaseHint(name, prop.Name)
	a.warnDeprecatedPropertyUsage(prop, name.Token.Pos)
	a.addPropertyBracketHint(prop, call.ParenPos)
	pos := call.End()
	pos.Column--
	pos.Offset--
	if prop.ReadKind == types.PropAccessNone {
		a.addStructuredError(NewWriteOnlyPropertyError(pos, prop.Name))
		a.addStructuredError(NewGenericError(nodes[len(nodes)-1].Token.Pos, "Array expected"))
		return prop.Type, true
	}
	indices := indexedPropertyArguments(nodes)
	if !a.checkIndexedCompatibilityArguments(a.currentClass, prop, indices, pos, a.inClassMethod, expr) {
		return nil, true
	}
	read := &ast.MemberAccessExpression{
		BaseNode: call.BaseNode, Member: name,
		Object: &ast.SelfExpression{BaseNode: name.BaseNode},
	}
	a.semanticInfo.SetIndexedPropertyRead(expr, &ast.IndexedPropertyReadBinding{
		Read: &ast.InheritedPropertyReadBinding{
			Read: read, Property: prop, Owner: compatibilityPropertyOwner(a.currentClass, prop),
		},
		Indices: indices,
	})
	return prop.Type, true
}

func (a *Analyzer) implicitIndexedCompatibilityProperty(call *ast.CallExpression) (*ast.Identifier, *types.PropertyInfo, *SymbolTable) {
	name, ok := call.Function.(*ast.Identifier)
	if !ok || a.currentClass == nil || call.ParenPos.Line == 0 || len(call.Arguments) != 0 || call.Token.Type != token.RPAREN {
		return nil, nil, nil
	}
	scope := a.implicitPropertyMethodScope(name.Value)
	if scope == nil || a.hasHelperMethod(a.currentImplicitSelfType(), name.Value) != nil {
		return nil, nil, nil
	}
	prop := propertyForCompatibilityCall(a.currentClass, name.Value)
	if prop == nil || !prop.IsIndexed || isFunctionPointerType(prop.Type) {
		return nil, nil, nil
	}
	return name, prop, scope
}
