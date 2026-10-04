package evaluator

import (
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// resolveHelperOverload is the fallback for callers without semantic bindings.
// Checked calls retain compiler selection and never enter this value-based path.
func (e *Evaluator) resolveHelperOverload(overloads []*ast.FunctionDecl, args []Value, ctx *ExecutionContext) *ast.FunctionDecl {
	argTypes := make([]types.Type, len(args))
	for i, arg := range args {
		argTypes[i] = e.getValueType(arg)
		if argTypes[i] == nil {
			return nil
		}
	}
	signatures := make([]types.Type, len(overloads))
	for i, declaration := range overloads {
		view := boundHelperCallView(declaration)
		signatures[i] = e.extractFunctionType(&view, ctx)
		if signatures[i] == nil {
			return nil
		}
	}
	selected, err := types.ResolveOverload(signatures, argTypes)
	if err != nil {
		return nil
	}
	return overloads[selected]
}

func boundHelperCallView(declaration *ast.FunctionDecl) ast.FunctionDecl {
	view := *declaration
	if view.IsHelper && len(view.Parameters) > 0 {
		// Function helpers declare Self in their AST parameter list; bound
		// calls supply it separately from the written arguments.
		view.Parameters = view.Parameters[1:]
		view.IsHelper = false
	}
	return view
}

// selectedHelperCall retrieves the compiler's declaring owner and overload
// slot. Out-of-line implementations replace that slot without changing identity.
// Unchecked calls and builtin helpers retain their existing dispatch paths.
func (e *Evaluator) selectedHelperCall(member *ast.Identifier) (*HelperMethodResult, bool) {
	info := e.SemanticInfo()
	if info == nil {
		return nil, false
	}
	annotation := info.GetType(member)
	if annotation == nil || !strings.HasPrefix(annotation.Name, "__helper_call:") {
		return nil, false
	}
	var helper HelperInfo
	if owner, ok := info.GetResolvedType(member).(*types.HelperType); ok && owner.TargetType != nil {
		for _, candidate := range e.typeSystem.LookupHelpers(owner.TargetType.String()) {
			if !ident.Equal(candidate.Name, owner.Name) {
				continue
			}
			// Function helpers can share their synthetic owner name even on
			// the same target. Their source declaration supplies the identity.
			if declaration, functionHelper := owner.Decl.(*ast.FunctionDecl); functionHelper && candidate.Methods[ident.Normalize(member.Value)] != declaration {
				continue
			}
			helper = candidate
			break
		}
	} else {
		helper = e.helperAnnotationOwner(annotation.Name, "__helper_call:")
	}
	method := e.findHelperMethodInHelper(helper, member.Value)
	if method == nil || method.Method == nil {
		return nil, true
	}
	_, payload, _ := strings.Cut(annotation.Name, ":")
	_, ordinal, _ := strings.Cut(payload, ":")
	index, err := strconv.Atoi(ordinal)
	if err != nil || index < 0 || index >= len(method.Overloads) {
		return nil, true
	}
	method.Method = method.Overloads[index]
	method.Overloads = []*ast.FunctionDecl{method.Method}
	return method, true
}

func (e *Evaluator) evalSelectedHelperCall(method *HelperMethodResult, self Value, expressions []ast.Expression, node ast.Node, ctx *ExecutionContext) Value {
	if method == nil {
		return e.newError(node, "selected helper method not implemented")
	}
	view := boundHelperCallView(method.Method)
	cached, err := e.ResolveOverloadFast(&view, expressions, ctx)
	var args []Value
	if err == nil {
		args, err = e.PrepareUserFunctionArgs(&view, expressions, cached, ctx, node)
	}
	if ctx.Exception() != nil {
		return e.nilValue()
	}
	if err != nil {
		return e.newError(node, "%s", err)
	}
	return e.CallASTHelperMethod(method.OwnerHelper, method.Method, self, args, node, ctx)
}

func (e *Evaluator) evalSelectedBoundHelperCall(member *ast.Identifier, expressions []ast.Expression, node ast.Node, ctx *ExecutionContext) (Value, bool) {
	method, bound := e.selectedHelperCall(member)
	if !bound {
		return nil, false
	}
	var self Value
	if value, found := ctx.Env().Get("Self"); found {
		self = value
	}
	if self == nil && method != nil {
		// Static helper bodies have no Self binding, but can call other
		// statically bound methods in the same helper.
		self = &runtime.TypeMetaValue{TypeInfo: method.OwnerHelper.TargetType, TypeName: method.OwnerHelper.TargetType.String()}
	}
	return e.evalSelectedHelperCall(method, self, expressions, node, ctx), true
}
