package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// resolveInterfaceIndexedProperty uses the static interface contract to capture
// a receiver and its index arguments once, including for compound assignments.
func (e *Evaluator) resolveInterfaceIndexedProperty(node *ast.IndexExpression, ctx *ExecutionContext) (Value, *types.PropertyInfo, []Value, bool, Value) {
	receiver, prop, indices := e.interfaceIndexedPropertyContract(node)
	if prop == nil || len(indices) > len(prop.IndexParamTypes) {
		return nil, nil, nil, false, nil
	}
	obj := e.Eval(receiver, ctx)
	obj = e.normalizeMemberReceiver(obj, receiver, node, ctx)
	if isError(obj) || ctx.Exception() != nil {
		return nil, nil, nil, true, obj
	}
	vals := make([]Value, len(indices))
	for n, index := range indices {
		vals[n] = e.Eval(index, ctx)
		if isError(vals[n]) || ctx.Exception() != nil {
			return nil, nil, nil, true, vals[n]
		}
	}
	return obj, prop, vals, true, nil
}

func (e *Evaluator) interfaceIndexedPropertyContract(node *ast.IndexExpression) (ast.Expression, *types.PropertyInfo, []ast.Expression) {
	if e.SemanticInfo() == nil {
		return nil, nil, nil
	}
	root, indices := CollectIndices(node)
	receiver := root
	var prop *types.PropertyInfo
	if member, ok := root.(*ast.MemberAccessExpression); ok {
		if iface := e.interfacePropertyReceiverType(member.Object); iface != nil {
			prop = iface.GetProperty(member.Member.Value)
			if prop != nil && prop.IsIndexed {
				receiver = member.Object
			}
		}
	}
	if prop == nil || !prop.IsIndexed {
		iface := e.interfacePropertyReceiverType(root)
		if iface == nil {
			return nil, nil, nil
		}
		prop = iface.GetDefaultProperty()
	}
	if prop == nil || !prop.IsIndexed {
		return nil, nil, nil
	}
	return receiver, prop, indices
}

// interfacePropertyReceiverType mirrors the implicit invocation applied by the
// analyzer when a parameterless routine is the indexed property's receiver.
func (e *Evaluator) interfacePropertyReceiverType(expr ast.Expression) *types.InterfaceType {
	typ := types.GetUnderlyingType(e.SemanticInfo().GetResolvedType(expr))
	switch callable := typ.(type) {
	case *types.FunctionType:
		if len(callable.Parameters) == 0 {
			typ = types.GetUnderlyingType(callable.ReturnType)
		}
	case *types.FunctionPointerType:
		if len(callable.Parameters) == 0 {
			typ = types.GetUnderlyingType(callable.ReturnType)
		}
	case *types.MethodPointerType:
		if len(callable.Parameters) == 0 {
			typ = types.GetUnderlyingType(callable.ReturnType)
		}
	}
	if iface, ok := typ.(*types.InterfaceType); ok {
		return iface
	}
	return nil
}

func (e *Evaluator) readInterfaceIndexedProperty(obj Value, prop *types.PropertyInfo, indices []Value, node ast.Node, ctx *ExecutionContext) Value {
	if iface, ok := obj.(*runtime.InterfaceInstance); ok {
		if iface.Object == nil {
			return e.newError(node, "interface is nil")
		}
		obj = iface.Object
	}
	return e.executeIndexedPropertyRead(obj, prop, indices, node, ctx)
}

// interfacePropertyResultIndex distinguishes indexing an accessor's array result
// from the arguments supplied to the accessor itself.
func (e *Evaluator) interfacePropertyResultIndex(node *ast.IndexExpression) bool {
	_, prop, indices := e.interfaceIndexedPropertyContract(node)
	return prop != nil && len(indices) > len(prop.IndexParamTypes)
}

func (e *Evaluator) writeInterfaceIndexedProperty(obj Value, prop *types.PropertyInfo, indices []Value, value Value, stmt *ast.AssignmentStatement, ctx *ExecutionContext) Value {
	descriptor := &runtime.PropertyDescriptor{Name: prop.Name, IsIndexed: prop.IsIndexed, IsDefault: prop.IsDefault, ReadSpec: prop.ReadSpec, WriteSpec: prop.WriteSpec, Impl: prop}
	return e.evalIndexedPropertyAssignmentDescriptor(obj, descriptor, indices, value, stmt, ctx)
}
