package semantic

import (
	"fmt"

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
	indices := make([]ast.Expression, len(nodes))
	for i, node := range nodes {
		indices[len(nodes)-1-i] = node.Index
	}
	a.checkIndexedCompatibilityArguments(class, prop, indices, pos, metaclass)
	read := &ast.MemberAccessExpression{BaseNode: call.BaseNode, Object: call.Object, Member: call.Method}
	a.semanticInfo.SetIndexedPropertyRead(expr, &ast.IndexedPropertyReadBinding{
		Read: &ast.InheritedPropertyReadBinding{Read: read, Property: prop, Owner: compatibilityPropertyOwner(class, prop)}, Indices: indices,
	})
	return prop.Type, true
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

func (a *Analyzer) indexedCompatibilityReceiver(call *ast.MethodCallExpression) (*types.ClassType, *types.PropertyInfo, bool, bool) {
	objectType := a.analyzeExpression(call.Object)
	objectType = a.applyImplicitCallType(call.Object, objectType)
	resolved := types.GetUnderlyingType(objectType)
	class, ok := resolved.(*types.ClassType)
	metaclass := false
	if meta, isMeta := resolved.(*types.ClassOfType); isMeta {
		class, ok, metaclass = meta.ClassType, true, true
	}
	if !ok || a.hasHelperMethod(objectType, call.Method.Value) != nil {
		return nil, nil, false, false
	}
	prop := propertyForCompatibilityCall(class, call.Method.Value)
	if prop == nil || !prop.IsIndexed || isFunctionPointerType(prop.Type) {
		return nil, nil, false, false
	}
	return class, prop, metaclass, true
}

func (a *Analyzer) checkIndexedCompatibilityArguments(class *types.ClassType, prop *types.PropertyInfo, indices []ast.Expression, pos token.Position, metaclass bool) {
	expected := a.getIndexedPropertyParamTypes(prop, class)
	argTypes := make([]types.Type, len(indices))
	failed := make([]bool, len(indices))
	for i, index := range indices {
		mark := len(a.errors)
		if i < len(expected) {
			argTypes[i] = a.analyzeArgumentForParameter(index, expected[i], false)
		} else {
			argTypes[i] = a.analyzeExpression(index)
		}
		failed[i] = argTypes[i] == nil || a.errorsSince(mark)
	}
	if metaclass && !indexedCompatibilityClassReader(class, prop) {
		a.addStructuredError(NewPropertyReadShouldBeStaticMethodError(pos))
		a.addStructuredError(NewClassMethodOrConstructorExpectedError(pos))
	}
	mark := len(a.errors)
	for i, typ := range argTypes {
		if i >= len(expected) {
			break
		}
		if failed[i] || a.argumentMatchesParameter(typ, expected[i], false) {
			continue
		}
		message := fmt.Sprintf("Argument %d expects type %q instead of %q", i, semanticTypeNameForDiagnostic(expected[i]), semanticTypeNameForDiagnostic(typ))
		if _, void := typ.(*types.VoidType); void {
			message = fmt.Sprintf("Argument %d expects type %q", i, semanticTypeNameForDiagnostic(expected[i]))
		}
		diagnostic := NewGenericError(pos, message)
		diagnostic.AfterChildren = true
		a.addStructuredError(diagnostic)
	}
	if !a.errorsSince(mark) {
		a.addMemberCallCountError(&types.FunctionType{Parameters: expected}, len(indices), pos)
	}
}
