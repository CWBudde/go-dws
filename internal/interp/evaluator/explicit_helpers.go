package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// namedHelperReceiver recognizes declarations without mistaking a variable
// shadowing a helper name for an explicit helper invocation.
func (e *Evaluator) namedHelperReceiver(object ast.Expression, ctx *ExecutionContext) HelperInfo {
	name, ok := object.(*ast.Identifier)
	if !ok {
		return nil
	}
	if info := e.SemanticInfo(); info != nil {
		if resolved := info.GetResolvedType(object); resolved != nil {
			if helper, ok := resolved.(*types.HelperType); ok {
				return e.lookupMutableHelper(helper.Name)
			}
			return nil
		}
	}
	// VisitHelperDecl binds the helper name to a TypeMetaValue; only another
	// binding is a variable shadowing the helper.
	if value, exists := ctx.Env().Get(name.Value); exists {
		if _, isTypeMeta := value.(*runtime.TypeMetaValue); !isTypeMeta {
			return nil
		}
	}
	return e.lookupMutableHelper(name.Value)
}

func (e *Evaluator) readExplicitHelperMember(helper HelperInfo, member *ast.Identifier, node ast.Node, ctx *ExecutionContext) Value {
	if prop, owner, ok := helper.GetProperty(member.Value); ok {
		self := &runtime.TypeMetaValue{TypeInfo: helper.TargetType, TypeName: helper.TargetType.String()}
		return e.executeHelperPropertyRead(owner, prop, self, node, ctx)
	}
	for owner := helper; owner != nil; owner = owner.ParentHelper {
		for name, value := range owner.ClassConsts {
			if ident.Equal(name, member.Value) {
				return value
			}
		}
		for name, value := range owner.ClassVars {
			if ident.Equal(name, member.Value) {
				return value
			}
		}
	}
	return e.newError(node, "member '%s' not found in helper '%s'", member.Value, helper.Name)
}

// evalExplicitHelperCall uses a temporary signature view so ordinary argument
// preparation sees the receiver exactly where it was written. Parsed ASTs and
// helper declaration signatures remain unchanged.
func (e *Evaluator) evalExplicitHelperCall(helper HelperInfo, member *ast.Identifier, expressions []ast.Expression, node ast.Node, ctx *ExecutionContext) (Value, bool) {
	method := e.findHelperMethodInHelper(helper, member.Value)
	if method == nil || method.Method == nil {
		return nil, false
	}
	helper = method.OwnerHelper
	views, originals := e.explicitHelperCallViews(helper, method, node)
	selected := views[0]
	var cached []Value
	var err error
	if lazyView := e.selectExplicitLazyHelper(views, expressions, ctx); lazyView != nil {
		selected = lazyView
		cached, err = e.ResolveOverloadFast(selected, expressions, ctx)
	} else if len(views) == 1 {
		cached, err = e.ResolveOverloadFast(selected, expressions, ctx)
	} else {
		selected, cached, err = e.ResolveOverloadMultiple(helper.Name+"."+member.Value, views, expressions, ctx)
	}
	if err != nil {
		if ctx.Exception() != nil {
			return e.nilValue(), true
		}
		return e.newError(node, "%s", err), true
	}
	args, err := e.PrepareUserFunctionArgs(selected, expressions, cached, ctx, node)
	if err == nil {
		args, err = e.EvaluateDefaultParameters(selected, args, ctx)
	}
	if err != nil {
		if ctx.Exception() != nil {
			return e.nilValue(), true
		}
		return e.newError(node, "%s", err), true
	}
	return e.callExplicitHelperMethod(helper, originals[selected], args, node, ctx), true
}

func (e *Evaluator) explicitHelperCallViews(helper HelperInfo, method *HelperMethodResult, node ast.Node) ([]*ast.FunctionDecl, map[*ast.FunctionDecl]*ast.FunctionDecl) {
	overloads := method.Overloads
	if len(overloads) == 0 {
		overloads = []*ast.FunctionDecl{method.Method}
	}
	views := make([]*ast.FunctionDecl, len(overloads))
	originals := make(map[*ast.FunctionDecl]*ast.FunctionDecl, len(overloads))
	for i, original := range overloads {
		view := *original
		view.IsHelper = false
		view.Parameters = append([]*ast.Parameter(nil), original.Parameters...)
		if explicitHelperNeedsReceiver(helper.TargetType, original) && !original.IsHelper {
			receiverType := e.explicitHelperReceiverAnnotation(helper, original, node)
			self := &ast.Parameter{Name: &ast.Identifier{Value: "Self"}, Type: receiverType}
			view.Parameters = append([]*ast.Parameter{self}, view.Parameters...)
		}
		views[i] = &view
		originals[&view] = original
	}
	return views, originals
}

func (e *Evaluator) explicitHelperReceiverAnnotation(helper HelperInfo, original *ast.FunctionDecl, node ast.Node) ast.TypeExpression {
	receiverType := ast.TypeExpression(&ast.TypeAnnotation{Name: helper.TargetType.String()})
	// Reuse the original type expression where semantic metadata carries
	// bounds or nominal identity that a spelling alone would lose.
	if info := e.SemanticInfo(); info != nil {
		if named, ok := info.GetResolvedType(explicitHelperObject(node)).(*types.HelperType); ok {
			for owner := named; owner != nil; owner = owner.ParentHelper {
				if ident.Equal(owner.Name, helper.Name) {
					if declaration, ok := owner.Decl.(*ast.HelperDecl); ok {
						receiverType = declaration.ForType
					}
					break
				}
			}
		}
	}
	if original.IsClassMethod {
		if target, ok := types.GetUnderlyingType(helper.TargetType).(*types.ClassType); ok {
			receiverType = &ast.ClassOfTypeNode{ClassType: &ast.TypeAnnotation{Name: target.Name}}
		}
	}
	return receiverType
}

func (e *Evaluator) callExplicitHelperMethod(helper HelperInfo, original *ast.FunctionDecl, args []Value, node ast.Node, ctx *ExecutionContext) Value {
	var self Value = &runtime.TypeMetaValue{TypeInfo: helper.TargetType, TypeName: helper.TargetType.String()}
	if explicitHelperNeedsReceiver(helper.TargetType, original) {
		self, args = args[0], args[1:]
		if original.IsClassMethod {
			if _, record := types.GetUnderlyingType(helper.TargetType).(*types.RecordType); record {
				if _, ok := self.(*runtime.RecordTypeValue); !ok {
					return e.newError(node, "record type expected")
				}
			}
		}
	}
	return e.CallASTHelperMethod(helper, original, self, args, node, ctx)
}

// Select lazy overloads from analyzed argument types before any evaluation.
// The ordinary runtime overload resolver evaluates arguments to discover their
// types, which would invoke a lazy argument once before its thunk is used.
func (e *Evaluator) selectExplicitLazyHelper(views []*ast.FunctionDecl, args []ast.Expression, ctx *ExecutionContext) *ast.FunctionDecl {
	if len(views) < 2 {
		return nil
	}
	hasLazy := false
	candidates := make([]types.Type, len(views))
	for i, view := range views {
		for _, param := range view.Parameters {
			hasLazy = hasLazy || param.IsLazy
		}
		candidates[i] = e.extractFunctionType(view, ctx)
		if candidates[i] == nil {
			return nil
		}
	}
	if !hasLazy {
		return nil
	}
	argTypes := make([]types.Type, len(args))
	for i, arg := range args {
		argTypes[i] = e.resolvedExpressionType(arg, ctx)
		if argTypes[i] == nil {
			return nil
		}
	}
	selected, err := types.ResolveOverload(candidates, argTypes)
	if err != nil {
		return nil
	}
	return views[selected]
}

// Mirrors DWScript CreateHelperMethodExpr/CreateSelfParameter: class methods
// need Self only when nonstatic and extending a class, metaclass, or record.
func explicitHelperNeedsReceiver(target types.Type, method *ast.FunctionDecl) bool {
	if !method.IsClassMethod {
		return true
	}
	if method.IsStatic {
		return false
	}
	switch types.GetUnderlyingType(target).(type) {
	case *types.ClassType, *types.ClassOfType, *types.RecordType:
		return true
	default:
		return false
	}
}

func explicitHelperObject(node ast.Node) ast.Expression {
	switch expr := node.(type) {
	case *ast.MethodCallExpression:
		return expr.Object
	case *ast.MemberAccessExpression:
		return expr.Object
	default:
		return nil
	}
}
