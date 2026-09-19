package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// methodArgumentReference retains the selected signature's type independently
// of the referenced storage. A later argument may resize the array before
// dispatch, but a stale reference must fail only when the method accesses it.
type methodArgumentReference struct {
	ReferenceAccessor
	argumentType types.Type
}

func (*methodArgumentReference) ValueKind() runtime.ValueKind { return runtime.KindReference }

// prepareVarMethodArguments resolves method signatures that require references
// before evaluating their arguments. Dispatch still owns execution, including
// virtual and statically bound method selection.
func (e *Evaluator) prepareVarMethodArguments(obj Value, node *ast.MethodCallExpression, ctx *ExecutionContext) (*ast.FunctionDecl, []Value, bool, error) {
	declarations := e.methodArgumentDeclarations(obj, node, ctx)
	name := node.Method.Value
	hasVar := false
	for _, declaration := range declarations {
		for _, parameter := range declaration.Parameters {
			hasVar = hasVar || parameter.ByRef
		}
	}
	if !hasVar {
		return nil, nil, false, nil
	}
	var selected *ast.FunctionDecl
	var cached []Value
	var err error
	if len(declarations) == 1 {
		selected = declarations[0]
		cached, err = e.ResolveOverloadFast(selected, node.Arguments, ctx)
	} else {
		selected, cached, err = e.ResolveOverloadMultiple(name, declarations, node.Arguments, ctx)
	}
	if err != nil {
		return nil, nil, true, err
	}
	prepared, err := e.PrepareUserFunctionArgs(selected, node.Arguments, cached, ctx, node)
	if err != nil {
		return nil, nil, true, err
	}
	for index, argument := range prepared {
		if index >= len(selected.Parameters) || !selected.Parameters[index].ByRef {
			continue
		}
		reference, ok := argument.(ReferenceAccessor)
		if !ok {
			continue
		}
		argumentType, err := e.ResolveTypeFromAnnotation(selected.Parameters[index].Type, ctx)
		if err != nil {
			return nil, nil, true, err
		}
		prepared[index] = &methodArgumentReference{ReferenceAccessor: reference, argumentType: argumentType}
	}
	return selected, prepared, true, nil
}

// methodArgumentDeclarations collects the signatures visible to method dispatch.
func (e *Evaluator) methodArgumentDeclarations(obj Value, node *ast.MethodCallExpression, ctx *ExecutionContext) []*ast.FunctionDecl {
	var declarations []*ast.FunctionDecl
	var methods []*runtime.MethodMetadata
	name := node.Method.Value
	switch receiver := obj.(type) {
	case InterfaceInstanceValue:
		if underlying := receiver.GetUnderlyingObjectValue(); underlying != nil {
			return e.methodArgumentDeclarations(underlying, node, ctx)
		}
	case *runtime.RecordTypeValue:
		declarations = receiver.ClassMethodOverloads[ident.Normalize(name)]
	case *runtime.RecordValue:
		declarations = receiver.GetRecordMethodOverloads(name)
	case *runtime.ObjectInstance:
		if receiver.Class != nil {
			if _, method := e.staticallyDispatchedMethod(receiver.Class, name, len(node.Arguments), node, ctx); method != nil {
				methods = []*runtime.MethodMetadata{method}
			} else {
				methods = append(methods, receiver.Class.GetMethodOverloads(name)...)
				methods = append(methods, receiver.Class.GetClassMethodOverloads(name)...)
			}
		}
	case ClassMetaValue:
		if class := receiver.GetClassInfo(); class != nil {
			methods = append(methods, class.GetConstructorOverloads(name)...)
			methods = append(methods, class.GetClassMethodOverloads(name)...)
		}
	}
	for _, method := range methods {
		if declaration := runtime.MethodDeclaration(method); declaration != nil {
			declarations = append(declarations, declaration)
		}
	}
	return declarations
}
