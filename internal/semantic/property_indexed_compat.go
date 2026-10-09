package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

// analyzeIndexedCompatibilityRead recognizes one declared bracket group after
// an explicitly named empty compatibility call. Later bracket groups index the
// returned value, so they must pass through the ordinary recursive index path.
func (a *Analyzer) analyzeIndexedCompatibilityRead(expr *ast.IndexExpression) (types.Type, bool) {
	nodes := []*ast.IndexExpression{expr}
	root := expr.Left
	for nodes[len(nodes)-1].CommaPos.Line != 0 {
		inner, ok := root.(*ast.IndexExpression)
		if !ok {
			return nil, false
		}
		nodes = append(nodes, inner)
		root = inner.Left
	}
	if inherited, ok := root.(*ast.InheritedExpression); ok {
		return a.analyzeInheritedIndexedPropertyRead(expr, inherited, nodes)
	}
	if call, ok := root.(*ast.CallExpression); ok {
		return a.analyzeImplicitIndexedCompatibilityRead(expr, call, nodes)
	}
	call, ok := root.(*ast.MethodCallExpression)
	if !ok || call.Incomplete || len(call.Arguments) != 0 {
		return nil, false
	}
	class, prop, metaclass, ok := a.indexedCompatibilityReceiver(call)
	if !ok {
		return nil, false
	}
	if !prop.IsReintroduce {
		a.analyzeOrdinaryIndexedPropertyCall(call, prop)
		return nil, true
	}
	if call.FirstArgumentToken.Type != token.RPAREN {
		return nil, false
	}
	a.addIdentifierCaseHint(call.Method, prop.Name)
	a.warnDeprecatedPropertyUsage(prop, call.Method.Token.Pos)
	a.addPropertyBracketHint(prop, call.ParenPos)
	pos := call.End()
	pos.Column--
	pos.Offset-- // Upstream captures the consumed compatibility ')'.
	if prop.ReadKind == types.PropAccessNone {
		a.addStructuredError(NewWriteOnlyPropertyError(pos, prop.Name))
		// Scalar recovery never reads declared indices after a missing getter.
		a.addStructuredError(NewGenericError(nodes[len(nodes)-1].Token.Pos, "Array expected"))
		return prop.Type, true
	}
	indices := indexedPropertyArguments(nodes)
	if !a.checkIndexedCompatibilityArguments(class, prop, indices, pos, metaclass, expr) {
		return nil, true
	}
	read := &ast.MemberAccessExpression{BaseNode: call.BaseNode, Object: call.Object, Member: call.Method}
	a.semanticInfo.SetIndexedPropertyRead(expr, &ast.IndexedPropertyReadBinding{
		Read: &ast.InheritedPropertyReadBinding{Read: read, Property: prop, Owner: compatibilityPropertyOwner(class, prop)}, Indices: indices,
	})
	return prop.Type, true
}

func indexedPropertyArguments(nodes []*ast.IndexExpression) []ast.Expression {
	if len(nodes) == 1 && nodes[0].Empty {
		return nil
	}
	indices := make([]ast.Expression, len(nodes))
	for i, node := range nodes {
		indices[len(nodes)-1-i] = node.Index
	}
	return indices
}

func (a *Analyzer) addPropertyBracketHint(prop *types.PropertyInfo, pos token.Position) {
	a.addHintAt(pos, "Property %q reintroduced a method, you should remove empty brackets () [line: %d, column: %d]", prop.Name, pos.Line, pos.Column)
}

func (a *Analyzer) analyzeOrdinaryIndexedPropertyCall(call *ast.MethodCallExpression, prop *types.PropertyInfo) {
	a.addIdentifierCaseHint(call.Method, prop.Name)
	a.warnDeprecatedPropertyUsage(prop, call.Method.Token.Pos)
	if prop.ReadKind == types.PropAccessNone {
		a.addStructuredError(NewWriteOnlyPropertyError(call.Method.Token.Pos, prop.Name))
	} else {
		a.addMemberCallCountError(&types.FunctionType{Parameters: prop.IndexParamTypes}, 0, call.Method.Token.Pos)
	}
	a.addPunctuationStop(call.ParenPos, "Not a method")
}

func indexedCompatibilityClassReader(class *types.ClassType, prop *types.PropertyInfo) bool {
	if prop.ReadKind == types.PropAccessExpression {
		return prop.IsClassProperty
	}
	for current := class; current != nil; current = current.Parent {
		if ident.Equal(current.Name, prop.ReadOwner) {
			return current.ClassMethodFlags[ident.Normalize(prop.ReadSpec)]
		}
	}
	return false
}

// indexedCompatibilityReceiver must analyze the receiver to classify the
// member. When the member is no compatibility candidate, the analyzed type is
// handed to the ordinary method-call path instead of being analyzed again.
func (a *Analyzer) indexedCompatibilityReceiver(call *ast.MethodCallExpression) (*types.ClassType, *types.PropertyInfo, bool, bool) {
	if a.isNamespaceReceiver(call.Object) {
		return nil, nil, false, false
	}
	analyzed := a.analyzeExpression(call.Object)
	objectType := a.implicitCallTypePreview(call.Object, analyzed)
	resolved := types.GetUnderlyingType(objectType)
	class, ok := resolved.(*types.ClassType)
	metaclass := false
	if meta, isMeta := resolved.(*types.ClassOfType); isMeta {
		class, ok, metaclass = meta.ClassType, true, true
	}
	var prop *types.PropertyInfo
	if ok && a.hasHelperMethod(objectType, call.Method.Value) == nil {
		prop = propertyForCompatibilityCall(class, call.Method.Value)
	}
	if prop == nil || !prop.IsIndexed || isFunctionPointerType(prop.Type) {
		if a.probedReceivers == nil {
			a.probedReceivers = make(map[ast.Expression]types.Type)
		}
		a.probedReceivers[call.Object] = analyzed
		return nil, nil, false, false
	}
	a.applyImplicitCallType(call.Object, analyzed)
	return class, prop, metaclass, true
}

// isNamespaceReceiver mirrors the receivers analyzeMethodCallExpression
// resolves as namespaces before analyzing them as expressions.
func (a *Analyzer) isNamespaceReceiver(object ast.Expression) bool {
	if name, ok := object.(*ast.Identifier); ok {
		if _, imported := a.importedUnitNamespace(name.Value); imported {
			return true
		}
	}
	return a.isJSONNamespace(object) || a.isDefaultNamespace(object)
}

// analyzeProbedReceiver consumes a receiver type a speculative probe already
// analyzed, and otherwise analyzes the receiver.
func (a *Analyzer) analyzeProbedReceiver(object ast.Expression) types.Type {
	if typ, ok := a.probedReceivers[object]; ok {
		delete(a.probedReceivers, object)
		return typ
	}
	return a.analyzeExpression(object)
}

func (a *Analyzer) checkIndexedCompatibilityArguments(class *types.ClassType, prop *types.PropertyInfo, indices []ast.Expression, pos token.Position, metaclass bool, list *ast.IndexExpression) bool {
	args, stopped := a.readPropertyIndexArguments(prop, indices, list)
	if stopped {
		return false
	}
	if metaclass && !indexedCompatibilityClassReader(class, prop) {
		readerError := NewPropertyReadShouldBeStaticMethodError(pos)
		readerError.AfterChildren = true
		a.addStructuredError(readerError)
		classError := NewClassMethodOrConstructorExpectedError(pos)
		classError.AfterChildren = true
		a.addStructuredError(classError)
	}
	a.checkPropertyReadArguments(prop, args, a.propertyClassAccessor(class, prop, false), pos)
	return true
}
