package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// indexedClassPropertyKnown classifies without executing the receiver. Generic
// array/string assignments must retain their existing RHS-first evaluation.
func (e *Evaluator) indexedClassPropertyKnown(receiver ast.Expression, name string, ctx *ExecutionContext) bool {
	// TypeCastAccessor indexing needs a separate selected static-owner path.
	if call, ok := receiver.(*ast.CallExpression); ok && e.typeSystem != nil {
		if id, ok := call.Function.(*ast.Identifier); ok && e.typeSystem.HasClass(id.Value) {
			return false
		}
	}
	if class := e.indexedPropertyReceiverClass(receiver, ctx); class != nil {
		return classHasCapturableProperty(class, name)
	}
	if value, found := e.peekClassReceiver(receiver, ctx); found {
		if accessor, ok := value.(ObjectValue); ok {
			if named, ok := accessor.(PropertyAccessor); ok {
				if name == "" {
					return capturablePropertyMetadata(named.GetDefaultProperty())
				}
				prop := named.LookupProperty(name)
				return capturablePropertyMetadata(prop)
			}
		}
		if class, ok := value.(ClassMetaValue); ok && class.GetClassInfo() != nil {
			prop := class.GetClassInfo().LookupProperty(name)
			return capturablePropertyMetadata(prop)
		}
	}
	return e.uncheckedIndexedPropertyReceiverKnown(receiver, name, ctx)
}

func (e *Evaluator) indexedPropertyReceiverClass(receiver ast.Expression, ctx *ExecutionContext) *types.ClassType {
	if typ := e.resolvedExpressionType(receiver, ctx); typ != nil {
		switch typ := types.GetUnderlyingType(typ).(type) {
		case *types.ClassType:
			return typ
		case *types.ClassOfType:
			return typ.ClassType
		}
	}
	return nil
}

func classHasCapturableProperty(class *types.ClassType, name string) bool {
	if name != "" {
		prop, ok := class.GetProperty(name)
		return ok && capturableIndexedProperty(prop)
	}
	for current := class; current != nil; current = current.Parent {
		for _, prop := range current.Properties {
			if prop.IsDefault {
				return capturableIndexedProperty(prop)
			}
		}
	}
	return false
}

func (e *Evaluator) uncheckedIndexedPropertyReceiverKnown(receiver ast.Expression, name string, ctx *ExecutionContext) bool {
	// An unchecked named routine still has a declared return annotation.
	function := receiver
	if call, ok := receiver.(*ast.CallExpression); ok {
		function = call.Function
	}
	if id, ok := function.(*ast.Identifier); ok && e.typeSystem != nil {
		declarations := e.typeSystem.LookupFunctions(id.Value)
		if local := e.lookupLocalFunctions(id.Value, ctx); local != nil {
			declarations = local.Decls
		}
		if value, found := ctx.Env().Get(id.Value); found {
			if callable, ok := value.(FunctionPointerCallable); ok {
				if decl, ok := callable.GetFunctionDecl().(*ast.FunctionDecl); ok {
					declarations = []*ast.FunctionDecl{decl}
				}
			}
		}
		if len(declarations) == 0 {
			return false
		}
		for _, decl := range declarations {
			annotation, ok := decl.ReturnType.(*ast.TypeAnnotation)
			if !ok {
				return false
			}
			class := e.typeSystem.LookupClass(annotation.Name)
			if class == nil {
				return false
			}
			var prop *runtime.PropertyInfo
			if name == "" {
				prop = class.GetDefaultProperty()
			} else {
				prop = class.LookupProperty(name)
			}
			if !capturablePropertyMetadata(prop) {
				return false
			}
		}
		return true
	}
	return false
}

// captureIndexedClassPropertyTarget retains the selected descriptor and caller
// storage through RHS evaluation. Unsupported targets remain on the old path.
func (e *Evaluator) captureIndexedClassPropertyTarget(target *ast.IndexExpression, stmt *ast.AssignmentStatement, ctx *ExecutionContext) (func(Value) Value, Value) {
	base, expressions := CollectIndices(target)
	receiver := target.Left
	name := ""
	if member, ok := base.(*ast.MemberAccessExpression); ok {
		receiver, name = member.Object, member.Member.Value
	} else {
		// Preserve the existing one-index default-property boundary.
		expressions = []ast.Expression{target.Index}
	}
	if !e.indexedClassPropertyKnown(receiver, name, ctx) {
		return nil, nil
	}
	obj := e.Eval(receiver, ctx)
	obj = e.normalizeMemberReceiver(obj, receiver, target, ctx)
	if isError(obj) || ctx.Exception() != nil {
		return nil, obj
	}
	descriptor := lookupIndexedPropertyDescriptor(obj, name)
	class, isClass := obj.(ClassMetaValue)
	var classProperty *runtime.PropertyInfo
	if isClass && class.GetClassInfo() != nil {
		classProperty = class.GetClassInfo().LookupProperty(name)
	}
	if classProperty != nil {
		prop, ok := unwrapPropertyInfo(classProperty.Impl)
		if !ok {
			return nil, e.newError(target, "invalid property info type")
		}
		indices, err := e.preparePropertyIndices(prop, expressions, target, ctx)
		if err != nil || ctx.Exception() != nil {
			return nil, err
		}
		return func(value Value) Value {
			result, _ := e.evalClassMetaIndexedPropertyWriteDescriptor(obj, class, classProperty, indices, value, stmt, ctx)
			return result
		}, nil
	}
	if descriptor == nil {
		return nil, e.newError(target, "indexed property '%s' not found", name)
	}
	prop, ok := unwrapPropertyInfo(descriptor.Impl)
	if !ok {
		return nil, e.newError(target, "invalid property info type")
	}
	indices, err := e.preparePropertyIndices(prop, expressions, target, ctx)
	if err != nil || ctx.Exception() != nil {
		return nil, err
	}
	return func(value Value) Value {
		return e.evalIndexedPropertyAssignmentDescriptor(obj, descriptor, indices, value, stmt, ctx)
	}, nil
}

// lookupIndexedPropertyDescriptor selects the named or default descriptor from
// the captured receiver, without evaluating any source expression.
func lookupIndexedPropertyDescriptor(obj Value, name string) *runtime.PropertyDescriptor {
	accessor, ok := obj.(PropertyAccessor)
	if !ok {
		return nil
	}
	if name == "" {
		return accessor.GetDefaultProperty()
	}
	return accessor.LookupProperty(name)
}

// peekClassReceiver reads plain storage only; properties and callable receivers
// are deliberately excluded because their evaluation can execute script code.
func (e *Evaluator) peekClassReceiver(expr ast.Expression, ctx *ExecutionContext) (Value, bool) {
	switch expr := expr.(type) {
	case *ast.Identifier:
		value, found := ctx.Env().Get(expr.Value)
		if found && value != nil {
			return value, true
		}
	case *ast.MemberAccessExpression:
		value, found := e.peekClassReceiver(expr.Object, ctx)
		if object, ok := value.(ObjectValue); found && ok {
			if !object.HasProperty(expr.Member.Value) {
				field := object.GetField(expr.Member.Value)
				return field, field != nil
			}
		}
	}
	return nil, false
}

func capturableIndexedProperty(prop *types.PropertyInfo) bool {
	// Indexed field writes have no supported evaluator dispatch. Leave their
	// historical path untouched until the independent field-writer slice.
	return prop != nil && prop.IsIndexed && prop.WriteKind != types.PropAccessField
}

func capturablePropertyMetadata(metadata any) bool {
	switch metadata := metadata.(type) {
	case *runtime.PropertyInfo:
		if metadata == nil {
			return false
		}
	case *runtime.PropertyDescriptor:
		if metadata == nil {
			return false
		}
	}
	prop, ok := unwrapPropertyInfo(metadata)
	return ok && capturableIndexedProperty(prop)
}
