package evaluator

import (
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// helperMethodPointer retains helper ownership as well as the captured receiver.
// Calling it reuses helper execution, including helper constants and Self binding.
type helperMethodPointer struct {
	helper HelperInfo
	method *ast.FunctionDecl
	runtime.FunctionPointerValue
	explicit bool
}

// functionPointerOf also exposes the pointer embedded in a helper method pointer,
// so binding retention and assignment checks treat both alike.
func functionPointerOf(value Value) (*runtime.FunctionPointerValue, bool) {
	switch pointer := value.(type) {
	case *runtime.FunctionPointerValue:
		return pointer, true
	case *helperMethodPointer:
		return &pointer.FunctionPointerValue, true
	}
	return nil, false
}

func (e *Evaluator) captureHelperMember(node *ast.MemberAccessExpression, ctx *ExecutionContext) (Value, bool) {
	if e.engineState == nil {
		return nil, false
	}
	info := e.SemanticInfo()
	if info == nil {
		return nil, false
	}
	annotation := info.GetType(node.Member)
	if annotation == nil || !strings.HasPrefix(annotation.Name, "__helper_member:") {
		return nil, false
	}
	helper := e.helperAnnotationOwner(annotation.Name, "__helper_member:")
	if helper == nil {
		return e.newError(node, "helper method not found"), true
	}
	explicit := e.namedHelperReceiver(node.Object, ctx) != nil
	var receiver Value
	if !explicit {
		receiver = e.normalizeMemberReceiver(e.Eval(node.Object, ctx), node.Object, node, ctx)
	}
	if isError(receiver) || ctx.Exception() != nil {
		return receiver, true
	}
	helper, method := e.helperValueMethod(node, helper, receiver, annotation.Name)
	if method == nil || method.Method == nil {
		return e.newError(node, "helper method not implemented"), true
	}
	pointer := e.helperValuePointerType(node, method.Method, ctx)
	return &helperMethodPointer{
		FunctionPointerValue: runtime.FunctionPointerValue{Function: method.Method, SelfObject: receiver, PointerType: pointer},
		helper:               helper,
		method:               method.Method,
		explicit:             explicit,
	}, true
}

func (e *Evaluator) captureHelperIdentifier(node *ast.Identifier, ctx *ExecutionContext) (Value, bool) {
	if e.engineState == nil {
		return nil, false
	}
	info := e.SemanticInfo()
	if info == nil {
		return nil, false
	}
	annotation := info.GetType(node)
	if annotation == nil || !strings.HasPrefix(annotation.Name, "__helper_body_member:") {
		return nil, false
	}
	helper := e.helperAnnotationOwner(annotation.Name, "__helper_body_member:")
	method := e.findHelperMethodInHelper(helper, node.Value)
	if method == nil || method.Method == nil {
		return e.newError(node, "helper method not implemented"), true
	}
	if !info.IsImplicitCall(node) {
		selectAnnotatedHelperMethod(method, annotation.Name)
	}
	receiver, _ := ctx.Env().Get("Self")
	var pointer *types.FunctionPointerType
	if resolved, ok := info.GetResolvedType(node).(*types.FunctionPointerType); ok && !info.IsImplicitCall(node) {
		pointer = resolved
	}
	value := &helperMethodPointer{FunctionPointerValue: runtime.FunctionPointerValue{Function: method.Method, SelfObject: receiver, PointerType: pointer}, helper: helper, method: method.Method}
	if info.IsImplicitCall(node) {
		return e.executeFunctionPointerDirect(value, nil, node, ctx), true
	}
	return value, true
}

func (e *Evaluator) helperAnnotationOwner(annotation, prefix string) HelperInfo {
	payload := strings.TrimPrefix(annotation, prefix)
	owner, _, _ := strings.Cut(payload, ":")
	return e.lookupMutableHelper(owner)
}

func selectAnnotatedHelperMethod(method *HelperMethodResult, annotation string) {
	_, payload, _ := strings.Cut(annotation, ":")
	_, ordinal, _ := strings.Cut(payload, ":")
	index, err := strconv.Atoi(ordinal)
	if err == nil && index >= 0 && index < len(method.Overloads) {
		method.Method = method.Overloads[index]
	}
}

func (e *Evaluator) implicitHelperValueMethod(node *ast.MemberAccessExpression, receiver Value) *HelperMethodResult {
	// Preserve the existing runtime declaration-order policy and static alias,
	// field and property binding while the general selection audit stays open.
	if receiverType := e.SemanticInfo().GetResolvedType(node.Object); receiverType != nil {
		for _, candidate := range orderedHelpersForLookup(e.typeSystem.LookupHelpers(receiverType.String())) {
			if method := e.findHelperMethodInHelper(candidate, node.Member.Value); method != nil && method.Method != nil {
				return method
			}
		}
	}
	return e.FindHelperMethod(receiver, node.Member.Value)
}

func (e *Evaluator) helperValuePointerType(node *ast.MemberAccessExpression, method *ast.FunctionDecl, ctx *ExecutionContext) *types.FunctionPointerType {
	if resolved, ok := e.SemanticInfo().GetResolvedType(node).(*types.FunctionPointerType); ok && !e.SemanticInfo().IsImplicitCall(node) {
		return resolved
	}
	pointer := e.buildFunctionPointerType(method, ctx)
	if pointer != nil && method.IsHelper {
		copyType := *pointer
		copyType.Parameters = pointer.Parameters[1:]
		pointer = &copyType
	}
	return pointer
}

// GetFunctionDecl exposes parameter flags in written-argument order. Explicit
// helper references include Self; bound function helpers omit their first param.
func (pointer *helperMethodPointer) GetFunctionDecl() any {
	typ := pointer.PointerType
	if typ == nil {
		return pointer.method
	}
	declaration := *pointer.method
	declaration.Parameters = make([]*ast.Parameter, len(typ.Parameters))
	for i := range declaration.Parameters {
		parameter := &ast.Parameter{}
		if i < len(typ.LazyParams) {
			parameter.IsLazy = typ.LazyParams[i]
		}
		if i < len(typ.VarParams) {
			parameter.ByRef = typ.VarParams[i]
		}
		declaration.Parameters[i] = parameter
	}
	return &declaration
}

func (e *Evaluator) helperValueMethod(node *ast.MemberAccessExpression, helper HelperInfo, receiver Value, annotation string) (HelperInfo, *HelperMethodResult) {
	method := e.findHelperMethodInHelper(helper, node.Member.Value)
	if e.SemanticInfo().IsImplicitCall(node) {
		if selected := e.implicitHelperValueMethod(node, receiver); selected != nil && selected.Method != nil {
			method, helper = selected, selected.OwnerHelper
		}
	}
	if method == nil || method.Method == nil {
		return helper, method
	}
	if signature, ok := e.SemanticInfo().GetResolvedType(node.Member).(*types.FunctionType); ok {
		for _, candidate := range method.Overloads {
			if len(candidate.Parameters) == len(signature.Parameters) {
				method.Method = candidate
				break
			}
		}
	}
	if !e.SemanticInfo().IsImplicitCall(node) {
		selectAnnotatedHelperMethod(method, annotation)
	}

	return helper, method
}
