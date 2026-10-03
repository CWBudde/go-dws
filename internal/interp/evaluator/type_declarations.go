package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// VisitTypeDeclaration evaluates a type declaration.
// Handles subrange types, function pointers, and type aliases.
func (e *Evaluator) VisitTypeDeclaration(node *ast.TypeDeclaration, ctx *ExecutionContext) Value {
	if node == nil {
		return e.newError(nil, "nil type declaration")
	}

	if node.IsSubrange {
		return e.evalSubrangeType(node, ctx)
	}
	if node.IsFunctionPointer {
		return e.evalFunctionPointerType(node, ctx)
	}
	if node.IsAlias {
		return e.evalTypeAlias(node, ctx)
	}

	return e.newError(node, "non-alias type declarations not yet supported")
}

// Evaluates subrange type (type TDigit = 0..9).
func (e *Evaluator) evalSubrangeType(node *ast.TypeDeclaration, ctx *ExecutionContext) Value {
	// Evaluate bounds
	lowBoundVal := e.Eval(node.LowBound, ctx)
	if isError(lowBoundVal) {
		return lowBoundVal
	}
	lowBoundIntVal, ok := lowBoundVal.(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "subrange low bound must be an integer")
	}
	lowBoundInt := int(lowBoundIntVal.Value)

	highBoundVal := e.Eval(node.HighBound, ctx)
	if isError(highBoundVal) {
		return highBoundVal
	}
	highBoundIntVal, ok := highBoundVal.(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "subrange high bound must be an integer")
	}
	highBoundInt := int(highBoundIntVal.Value)

	// Validate bounds
	if lowBoundInt > highBoundInt {
		return e.newError(node, "subrange low bound (%d) cannot be greater than high bound (%d)", lowBoundInt, highBoundInt)
	}

	// Create and register subrange type
	subrangeType := &types.SubrangeType{
		BaseType:  types.INTEGER,
		Name:      node.Name.Value,
		LowBound:  lowBoundInt,
		HighBound: highBoundInt,
	}

	if resolved, ok := e.resolvedSemanticType(node).(*types.SubrangeType); ok {
		subrangeType = resolved
	}
	e.typeSystem.RegisterSubrangeType(node.Name.Value, subrangeType)

	return &runtime.NilValue{}
}

// Evaluates function pointer type (type TCallback = procedure(x: Integer)).
func (e *Evaluator) evalFunctionPointerType(node *ast.TypeDeclaration, ctx *ExecutionContext) Value {
	if node.FunctionPointerType == nil {
		return e.newError(node, "function pointer type declaration has no type information")
	}

	resolvedType, err := e.ResolveTypeFromAnnotation(node.FunctionPointerType, ctx)
	if err != nil {
		return e.newError(node, "invalid function pointer type: %v", err)
	}

	// Register in TypeSystem
	if e.typeSystem != nil {
		e.typeSystem.RegisterFunctionPointerType(node.Name.Value, resolvedType)
	}

	// Legacy marker
	typeKey := "__funcptr_type_" + node.Name.Value
	ctx.Env().Define(typeKey, &runtime.StringValue{Value: "function_pointer_type"})

	return &runtime.NilValue{}
}

// Evaluates type alias (type TUserID = Integer).
func (e *Evaluator) evalTypeAlias(node *ast.TypeDeclaration, ctx *ExecutionContext) Value {
	aliasedType, err := e.ResolveTypeFromAnnotation(node.AliasedType, ctx)
	if err != nil {
		return e.newError(node, "unknown type '%s' in type alias", node.AliasedType.String())
	}
	if semanticType := e.resolvedSemanticType(node); semanticType != nil {
		aliasedType = semanticType
	}

	// Create and register type alias.
	typeAlias := &runtime.TypeAliasValue{
		Name:        node.Name.Value,
		AliasedType: aliasedType,
	}

	// If alias targets an enum, register the alias for scoped enum lookups.
	if enumType, ok := types.GetUnderlyingType(aliasedType).(*types.EnumType); ok {
		e.typeSystem.RegisterEnumType(node.Name.Value, runtime.NewEnumTypeValue(enumType))
	}

	typeKey := "__type_alias_" + ident.Normalize(node.Name.Value)
	ctx.Env().Define(typeKey, typeAlias)
	if record, ok := types.GetUnderlyingType(aliasedType).(*types.RecordType); ok {
		if value, found := ctx.Env().Get("__record_type_" + ident.Normalize(record.Name)); found {
			if canonical, ok := value.(*runtime.RecordTypeValue); ok {
				aliasValue := *canonical
				aliasValue.SourceType = aliasedType
				ctx.Env().Define(node.Name.Value, &aliasValue)
				return &runtime.NilValue{}
			}
		}
	}
	ctx.Env().Define(node.Name.Value, &runtime.TypeMetaValue{
		TypeInfo: aliasedType,
		TypeName: node.Name.Value,
	})

	return &runtime.NilValue{}
}
