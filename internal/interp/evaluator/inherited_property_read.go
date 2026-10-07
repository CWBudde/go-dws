package evaluator

import (
	"fmt"
	"strconv"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// evalInheritedPropertyRead retains the resolved accessor identity independently
// from both the property descriptor's owner and dynamic Self.
func (e *Evaluator) evalInheritedPropertyRead(binding *ast.InheritedPropertyReadBinding, ctx *ExecutionContext) Value {
	prop := binding.Property
	owner := e.typeSystem.LookupClass(prop.ReadOwner)
	if owner == nil {
		return e.newError(binding.Read, "property accessor owner '%s' not found", prop.ReadOwner)
	}
	receiver := e.Eval(binding.Read.Object, ctx)
	if isError(receiver) {
		return receiver
	}
	if prop.ReadKind == types.PropAccessExpression {
		return e.withInheritedPropertyContext(prop, binding.Read, ctx, func() Value {
			return e.evalInheritedPropertyExpression(receiver, owner, prop, binding.Read, ctx)
		})
	}
	if prop.ReadStorage != types.PropStorageNone {
		return e.evalInheritedPropertyStorage(receiver, owner, prop, binding.Read)
	}
	method := owner.LookupMethod(prop.ReadSpec)
	if method == nil {
		method = owner.LookupClassMethod(prop.ReadSpec)
	}
	if method == nil {
		return e.newError(binding.Read, "property getter '%s' not found", prop.ReadSpec)
	}
	dynamicClass := e.helperReceiverClassInfo(receiver)
	if dynamicClass == nil {
		if method.IsVirtual || method.IsOverride || method.IsAbstract {
			return e.newError(binding.Read, "Object not instantiated")
		}
		dynamicClass = owner
		if _, currentClass, ok := currentClassMetaValue(ctx); ok {
			dynamicClass = currentClass.GetClassInfo()
		}
	}
	method = inheritedPropertyGetter(owner, dynamicClass, method)
	return e.withInheritedPropertyContext(prop, binding.Read, ctx, func() Value {
		return e.executeInheritedPropertyGetter(receiver, dynamicClass, method, prop, binding.Read, ctx)
	})
}

func (e *Evaluator) evalInheritedPropertyStorage(receiver Value, owner runtime.IClassInfo, prop *types.PropertyInfo, node ast.Node) Value {
	switch prop.ReadStorage {
	case types.PropStorageField:
		if object, ok := receiver.(ObjectValue); ok {
			if value := getFieldWithStaticClass(object, prop.ReadSpec, owner.GetName()); value != nil {
				return value
			}
		}
		return e.newError(node, "Object not instantiated")
	case types.PropStorageClassVar:
		if value, _ := owner.LookupClassVar(prop.ReadSpec); value != nil {
			return value
		}
	case types.PropStorageConstant:
		if constants, ok := owner.(runtime.ClassConstantProvider); ok {
			if value, found := constants.GetClassConstant(prop.ReadSpec); found {
				return value
			}
		}
	}
	return e.newError(node, "property reader '%s' not found", prop.ReadSpec)
}

// Walk past restarted chains to retain the most-derived override of the
// original slot. A name lookup or fallback to the original declaration would
// either enter the new chain or lose overrides preceding its restart.
func inheritedPropertyGetter(owner, dynamicClass runtime.IClassInfo, method *runtime.MethodMetadata) *runtime.MethodMetadata {
	if !method.IsVirtual && !method.IsOverride && !method.IsAbstract {
		return method
	}
	signature := ident.Normalize(method.Name) + "_" + strconv.Itoa(len(method.Parameters))
	original := owner.GetVirtualMethodTable()[signature]
	for current := dynamicClass; current != nil; current = current.GetParent() {
		entry := current.GetVirtualMethodTable()[signature]
		if entry != nil && entry.Method != nil && !isDifferentVirtualChain(entry, original) {
			return entry.Method
		}
	}
	return method
}

func (e *Evaluator) withInheritedPropertyContext(prop *types.PropertyInfo, node ast.Node, ctx *ExecutionContext, read func() Value) Value {
	propCtx := ctx.PropContext()
	key := fmt.Sprintf("%p", prop)
	for _, active := range propCtx.PropertyChain {
		if active == key {
			return e.newError(node, "circular property reference detected: %s", prop.Name)
		}
	}
	propCtx.PropertyChain = append(propCtx.PropertyChain, key)
	savedInGetter := propCtx.InPropertyGetter
	propCtx.InPropertyGetter = true
	defer func() {
		propCtx.PropertyChain = propCtx.PropertyChain[:len(propCtx.PropertyChain)-1]
		propCtx.InPropertyGetter = savedInGetter
	}()
	return read()
}

func (e *Evaluator) executeInheritedPropertyGetter(receiver Value, class runtime.IClassInfo, method *runtime.MethodMetadata, prop *types.PropertyInfo, node ast.Node, ctx *ExecutionContext) Value {
	args, err := e.buildIndexDirectiveArgs(prop)
	if err != nil {
		return e.newError(node, "%s", err)
	}
	if method.IsClassMethod {
		value, err := e.typeSystem.CreateClassValue(class.GetName())
		if err != nil {
			return e.newError(node, "%s", err)
		}
		if meta, ok := value.(ClassMetaValue); ok {
			return e.executeClassMethodDirect(meta, method, args, node, ctx)
		}
		return e.newError(node, "class getter requires a class receiver")
	}
	return e.executeMethodWithClassInfo(receiver, class, method, args, ctx)
}

func (e *Evaluator) evalInheritedPropertyExpression(receiver Value, owner runtime.IClassInfo, prop *types.PropertyInfo, node ast.Node, ctx *ExecutionContext) Value {
	expression, ok := prop.ReadExpr.(ast.Expression)
	if !ok {
		return e.newError(node, "invalid property expression")
	}
	ctx.PushEnv()
	defer ctx.PopEnv()
	scope := newBindingScope()
	defer scope.cleanup(e, ctx.Env())
	scope.defineExposed(ctx, "Self", receiver)
	scope.defineExposed(ctx, "__CurrentMethodClass__", &runtime.StringValue{Value: owner.GetName()})
	e.bindClassVarsForProperty(owner, ctx, scope)
	e.bindClassConstantsForMethod(owner, ctx)
	if object, ok := receiver.(ObjectValue); ok {
		for _, class := range classInfoHierarchy(owner) {
			if metadata := class.GetMetadata(); metadata != nil {
				for name := range metadata.Fields {
					if value := getFieldWithStaticClass(object, name, owner.GetName()); value != nil {
						scope.defineExposed(ctx, name, value)
					}
				}
			}
		}
	}
	return e.Eval(expression, ctx)
}
