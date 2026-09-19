package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// getValueType returns the types.Type for a runtime Value.
//
// This method maps runtime values to their corresponding semantic types,
// enabling the semantic analyzer's overload resolution to work with
// evaluated argument values.
func (e *Evaluator) getValueType(val Value) types.Type {
	if val == nil {
		return types.NIL
	}

	switch v := val.(type) {
	case *runtime.IntegerValue:
		return types.INTEGER
	case *runtime.FloatValue:
		return types.FLOAT
	case *runtime.StringValue:
		return types.STRING
	case *runtime.BooleanValue:
		return types.BOOLEAN
	case *runtime.NilValue:
		return types.NIL
	case *runtime.VariantValue:
		return types.VARIANT
	case *runtime.ArrayValue:
		if v.ArrayType != nil {
			return v.ArrayType
		}
		return types.NIL
	case *runtime.ObjectInstance:
		if v.Class != nil && v.Class.GetClassType() != nil {
			return v.Class.GetClassType()
		}
		return types.NIL
	case *runtime.RecordValue:
		if v.RecordType != nil {
			return v.RecordType
		}
		return types.NIL
	case *runtime.FunctionPointerValue:
		return e.functionPointerValueType(v)
	default:
		// For ObjectInstance - use helper method (was e.adapter.GetClassMetadataFromValue)
		if metadata := e.getClassMetadataFromValue(val); metadata != nil {
			return e.classTypeFromMetadata(metadata)
		}
		return types.NIL
	}
}

// functionPointerValueType returns the static pointer type of a function or
// method pointer value, so an overload can be resolved from already-evaluated
// argument values.
//
// Both construction paths populate PointerType; the MethodMetadata fallback
// covers pointers built from a canonical runtime callable whose signature was
// resolved after the pointer value itself.
//
// A bound pointer (SelfObject != nil) is always reported as a MethodPointerType,
// even when it was constructed with a plain FunctionPointerType: the pointer's
// identity decides whether a `procedure(...) of object` parameter can accept it,
// and buildFunctionPointerType cannot express that distinction.
func (e *Evaluator) functionPointerValueType(fp *runtime.FunctionPointerValue) types.Type {
	if fp == nil {
		return types.NIL
	}
	if fp.PointerType != nil {
		return methodPointerIfBound(fp.PointerType, fp.SelfObject != nil)
	}
	if fp.Callable != nil {
		paramTypes := make([]types.Type, 0, len(fp.Callable.Parameters))
		for _, param := range fp.Callable.Parameters {
			if param.Type == nil {
				return types.NIL
			}
			paramTypes = append(paramTypes, param.Type)
		}
		var returnType types.Type
		if fp.Callable.ReturnType != nil && fp.Callable.ReturnType != types.VOID {
			returnType = fp.Callable.ReturnType
		}
		if fp.SelfObject != nil {
			return types.NewMethodPointerType(paramTypes, returnType)
		}
		return types.NewFunctionPointerType(paramTypes, returnType)
	}
	return types.NIL
}

// methodPointerIfBound promotes a plain function pointer type to a method
// pointer type when the value it describes carries a bound receiver.
func methodPointerIfBound(pointerType types.Type, bound bool) types.Type {
	if !bound {
		return pointerType
	}
	funcPtr, ok := types.GetUnderlyingType(pointerType).(*types.FunctionPointerType)
	if !ok {
		// Already a MethodPointerType (or not a pointer at all): leave it alone.
		return pointerType
	}
	return types.NewMethodPointerType(funcPtr.Parameters, funcPtr.ReturnType)
}

// classTypeFromMetadata builds a types.ClassType from runtime.ClassMetadata.
//
// This recursively builds the class type hierarchy by looking up parent
// metadata from the runtime.ClassMetadata's ParentMetadata pointer.
func (e *Evaluator) classTypeFromMetadata(metadata *runtime.ClassMetadata) types.Type {
	if metadata == nil {
		return types.NIL
	}
	if e.typeSystem != nil {
		if class := e.typeSystem.LookupClass(metadata.Name); class != nil && class.GetClassType() != nil {
			return class.GetClassType()
		}
	}

	var parentType *types.ClassType
	if metadata.Parent != nil {
		if pt := e.classTypeFromMetadata(metadata.Parent); pt != nil {
			if ct, ok := pt.(*types.ClassType); ok {
				parentType = ct
			}
		}
	}

	return types.NewClassType(metadata.Name, parentType)
}

// extractFunctionType extracts a types.FunctionType from an ast.FunctionDecl.
//
// This method converts AST function declarations to semantic FunctionType
// objects, extracting parameter types, names, modifiers (lazy/var/const),
// default values, and return type.
//
// Returns nil if any parameter type cannot be resolved.
func (e *Evaluator) extractFunctionType(fn *ast.FunctionDecl, ctx *ExecutionContext) *types.FunctionType {
	paramTypes := make([]types.Type, len(fn.Parameters))
	paramNames := make([]string, len(fn.Parameters))
	lazyParams := make([]bool, len(fn.Parameters))
	varParams := make([]bool, len(fn.Parameters))
	constParams := make([]bool, len(fn.Parameters))
	defaultValues := make([]interface{}, len(fn.Parameters))

	for idx, param := range fn.Parameters {
		if param.Type == nil {
			return nil // Invalid function - missing type annotation
		}

		// Use evaluator's existing resolveTypeName for type resolution
		paramType, err := e.ResolveTypeFromAnnotation(param.Type, ctx)
		if err != nil {
			return nil
		}

		paramTypes[idx] = paramType
		paramNames[idx] = param.Name.Value
		lazyParams[idx] = param.IsLazy
		varParams[idx] = param.ByRef
		constParams[idx] = param.IsConst
		defaultValues[idx] = param.DefaultValue
	}

	var returnType types.Type = types.VOID
	if fn.ReturnType != nil {
		if rt, err := e.ResolveTypeFromAnnotation(fn.ReturnType, ctx); err == nil {
			returnType = rt
		}
	}

	return types.NewFunctionTypeWithMetadata(
		paramTypes, paramNames, defaultValues,
		lazyParams, varParams, constParams,
		returnType,
	)
}

// ResolveOverloadFast handles single-overload case efficiently.
//
// Arguments are prepared in source order: var parameters capture references,
// ordinary parameters cache values, and lazy parameters remain unevaluated.
//
// Returns the cached argument values where:
//   - Var parameters: captured ReferenceAccessor
//   - Ordinary parameters: evaluated Value
//   - Lazy parameters: nil (to be wrapped as LazyThunk later)
func (e *Evaluator) ResolveOverloadFast(
	fn *ast.FunctionDecl,
	argExprs []ast.Expression,
	ctx *ExecutionContext,
) ([]Value, error) {
	argValues := make([]Value, len(argExprs))

	for idx, argExpr := range argExprs {
		if ctx.Exception() != nil {
			return nil, fmt.Errorf("argument evaluation raised an exception")
		}
		// Check if this parameter is lazy
		isLazy := idx < len(fn.Parameters) && fn.Parameters[idx].IsLazy
		if isLazy {
			// Don't evaluate lazy parameters - mark as nil
			// PrepareUserFunctionArgs will wrap them later
			argValues[idx] = nil
		} else if idx < len(fn.Parameters) && fn.Parameters[idx].ByRef {
			ref, err := e.prepareByRefArgument(argExpr, ctx)
			if err != nil {
				return nil, err
			}
			if ctx.Exception() != nil {
				return nil, fmt.Errorf("argument evaluation raised an exception")
			}
			argValues[idx] = ref
		} else {
			// Set record type context if argument is anonymous record literal
			previousRecordType := ctx.RecordTypeContext()
			contextSet := false
			if idx < len(fn.Parameters) && fn.Parameters[idx].Type != nil {
				paramType := e.recordTypeFromAnnotation(fn.Parameters[idx].Type, ctx)
				if recordLit, ok := argExpr.(*ast.RecordLiteralExpression); ok && recordLit.TypeName == nil {
					if paramType != nil {
						ctx.SetRecordTypeContext(paramType)
						contextSet = true
					}
				}
			}

			// The parameter's own array type (or none) governs array literal
			// arguments; an enclosing assignment's array context must not leak
			// into the call (e.g. F('x', [1, 2]) inside a TItemArray literal).
			prevArrayCtx := ctx.ArrayTypeContext()
			ctx.ClearArrayTypeContext()
			if idx < len(fn.Parameters) && fn.Parameters[idx].Type != nil {
				if paramType, err := e.ResolveTypeFromAnnotation(fn.Parameters[idx].Type, ctx); err == nil {
					if arrType, ok := types.GetUnderlyingType(paramType).(*types.ArrayType); ok {
						ctx.SetArrayTypeContext(arrType)
					}
				}
			}

			// Evaluate non-lazy parameters
			val := e.Eval(argExpr, ctx)
			ctx.SetArrayTypeContext(prevArrayCtx)

			if contextSet {
				ctx.SetRecordTypeContext(previousRecordType)
			}

			if isError(val) {
				return nil, fmt.Errorf("error evaluating argument %d: %v", idx+1, val)
			}
			if ctx.Exception() != nil {
				return nil, fmt.Errorf("argument evaluation raised an exception")
			}
			argValues[idx] = val
		}
	}

	return argValues, nil
}

// ResolveOverloadMultiple selects an overload and caches each argument once.
// Concrete checked var calls select first; dynamic calls retain runtime type
// selection and capture array storage while evaluating arguments in source order.
func (e *Evaluator) ResolveOverloadMultiple(funcName string, overloads []*ast.FunctionDecl, argExprs []ast.Expression, ctx *ExecutionContext) (*ast.FunctionDecl, []Value, error) {
	candidates := make([]types.Type, len(overloads))
	hasVar := false
	for idx, fn := range overloads {
		funcType := e.extractFunctionType(fn, ctx)
		if funcType == nil {
			return nil, nil, fmt.Errorf("unable to extract function type for overload %d of '%s'", idx+1, funcName)
		}
		candidates[idx] = funcType
		for _, parameter := range fn.Parameters {
			hasVar = hasVar || parameter.ByRef
		}
	}
	if selected, ok := e.selectStaticVarOverload(candidates, argExprs, hasVar, ctx); ok {
		fn := overloads[selected]
		args, err := e.ResolveOverloadFast(fn, argExprs, ctx)
		return fn, args, err
	}

	argTypes := make([]types.Type, len(argExprs))
	argValues := make([]Value, len(argExprs))
	capturedArrays := make([]*capturedArrayArgument, len(argExprs))
	previousArrayContext := ctx.ArrayTypeContext()
	ctx.ClearArrayTypeContext()
	defer ctx.SetArrayTypeContext(previousArrayContext)
	for idx, expr := range argExprs {
		allVar, anyVar := e.overloadArgumentModes(candidates, expr, idx, ctx)
		argument, err := e.evaluateOverloadArgument(expr, allVar, anyVar, ctx)
		if err != nil {
			return nil, nil, err
		}
		argValues[idx], argTypes[idx], capturedArrays[idx] = argument.value, argument.typ, argument.captured
	}
	ctx.SetArrayTypeContext(previousArrayContext)
	selected, err := types.ResolveOverload(candidates, argTypes)
	if err != nil {
		return nil, nil, fmt.Errorf("There is no overloaded version of %q that can be called with these arguments", funcName) //nolint:staticcheck // Preserve the DWScript diagnostic capitalization.
	}
	fn := overloads[selected]
	for idx, captured := range capturedArrays {
		if captured != nil && idx < len(fn.Parameters) && fn.Parameters[idx].ByRef {
			ref, err := e.bindCapturedOverloadArray(captured, ctx)
			if err != nil {
				return nil, nil, err
			}
			argValues[idx] = ref
		}
	}
	return fn, argValues, nil
}

// selectStaticVarOverload leaves dynamic Variant/object discrimination on the
// existing runtime path instead of changing how those overloads are selected.
func (e *Evaluator) selectStaticVarOverload(candidates []types.Type, args []ast.Expression, hasVar bool, ctx *ExecutionContext) (int, bool) {
	if !hasVar || e.engineState == nil {
		return 0, false
	}
	staticTypes := make([]types.Type, len(args))
	for idx, expr := range args {
		staticTypes[idx] = e.resolvedExpressionType(expr, ctx)
		if staticTypes[idx] == nil {
			return 0, false
		}
		switch types.GetUnderlyingType(staticTypes[idx]).TypeKind() {
		case "VARIANT", "CLASS", "INTERFACE":
			return 0, false
		}
	}
	selected, err := types.ResolveOverload(candidates, staticTypes)
	return selected, err == nil
}

// overloadArgumentModes excludes candidates incompatible with a concrete
// argument type before deciding whether its storage can be bound immediately.
func (e *Evaluator) overloadArgumentModes(candidates []types.Type, expr ast.Expression, index int, ctx *ExecutionContext) (allVar, anyVar bool) {
	var staticType types.Type
	if e.engineState != nil {
		staticType = e.resolvedExpressionType(expr, ctx)
	}
	if staticType != nil {
		switch types.GetUnderlyingType(staticType).TypeKind() {
		case "INTEGER", "FLOAT", "STRING", "BOOLEAN", "ARRAY", "RECORD":
		default:
			staticType = nil
		}
	}
	allVar, viable := true, false
	for _, candidate := range candidates {
		signature, ok := candidate.(*types.FunctionType)
		if !ok || index >= len(signature.Parameters) {
			continue
		}
		parameter := types.NewFunctionType([]types.Type{signature.Parameters[index]}, nil)
		if staticType != nil && types.SignatureDistance([]types.Type{staticType}, parameter) < 0 {
			continue
		}
		viable = true
		byRef := index < len(signature.VarParams) && signature.VarParams[index]
		allVar = allVar && byRef
		anyVar = anyVar || byRef
	}
	return allVar && viable, anyVar
}

type evaluatedOverloadArgument struct {
	value    Value
	typ      types.Type
	captured *capturedArrayArgument
}

func (e *Evaluator) evaluateOverloadArgument(expr ast.Expression, allVar, anyVar bool, ctx *ExecutionContext) (evaluatedOverloadArgument, error) {
	var argument evaluatedOverloadArgument
	if allVar {
		ref, err := e.prepareByRefArgument(expr, ctx)
		if err != nil {
			return argument, err
		}
		if err := argumentEvaluationError(ref, ctx); err != nil {
			return argument, err
		}
		accessor, ok := ref.(ReferenceAccessor)
		if !ok {
			return argument, fmt.Errorf("var argument did not produce a reference")
		}
		value, err := accessor.Dereference()
		if err != nil {
			e.raiseBoundExceededError(err, ctx)
			return argument, err
		}
		argument.value, argument.typ = ref, e.getValueType(value)
		return argument, argumentEvaluationError(value, ctx)
	}
	handled := false
	if anyVar {
		var err error
		argument.value, argument.captured, handled, err = e.captureOverloadArrayArgument(expr, ctx)
		if err != nil {
			return argument, err
		}
	}
	if !handled {
		argument.value = e.Eval(expr, ctx)
	}
	argument.typ = e.getValueType(argument.value)
	return argument, argumentEvaluationError(argument.value, ctx)
}

func (e *Evaluator) bindCapturedOverloadArray(captured *capturedArrayArgument, ctx *ExecutionContext) (Value, error) {
	container := captured.bindContainer()
	if err := argumentEvaluationError(container, ctx); err != nil {
		return nil, err
	}
	array, ok := container.(*runtime.ArrayValue)
	if !ok {
		return nil, fmt.Errorf("var parameter requires an array element")
	}
	if array == captured.array {
		return captured.reference, nil
	}
	return e.bindArrayElementReference(array, captured.index, captured.node, ctx)
}
