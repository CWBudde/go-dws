package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/semantic"
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
	case *runtime.RecordValue:
		if v.RecordType != nil {
			return v.RecordType
		}
		return types.NIL
	default:
		// For ObjectInstance - use helper method (was e.adapter.GetClassMetadataFromValue)
		if metadata := e.getClassMetadataFromValue(val); metadata != nil {
			return e.classTypeFromMetadata(metadata)
		}
		return types.NIL
	}
}

// classTypeFromMetadata builds a types.ClassType from runtime.ClassMetadata.
//
// This recursively builds the class type hierarchy by looking up parent
// metadata from the runtime.ClassMetadata's ParentMetadata pointer.
func (e *Evaluator) classTypeFromMetadata(metadata *runtime.ClassMetadata) types.Type {
	if metadata == nil {
		return types.NIL
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
		paramType, err := e.resolveTypeName(param.Type.String(), ctx)
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
		if rt, err := e.resolveTypeName(fn.ReturnType.String(), ctx); err == nil {
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
// This method skips evaluation for lazy parameters (they're wrapped later by
// PrepareUserFunctionArgs). Non-lazy parameters are evaluated and cached to
// prevent double-evaluation.
//
// Returns the cached argument values where:
//   - Non-lazy parameters: evaluated Value
//   - Lazy parameters: nil (to be wrapped as LazyThunk later)
func (e *Evaluator) ResolveOverloadFast(
	fn *ast.FunctionDecl,
	argExprs []ast.Expression,
	ctx *ExecutionContext,
) ([]Value, error) {
	argValues := make([]Value, len(argExprs))

	for idx, argExpr := range argExprs {
		// Check if this parameter is lazy
		isLazy := idx < len(fn.Parameters) && fn.Parameters[idx].IsLazy
		if isLazy {
			// Don't evaluate lazy parameters - mark as nil
			// PrepareUserFunctionArgs will wrap them later
			argValues[idx] = nil
		} else {
			// Set record type context if argument is anonymous record literal
			contextSet := false
			if idx < len(fn.Parameters) && fn.Parameters[idx].Type != nil {
				paramType := fn.Parameters[idx].Type.String()
				if recordLit, ok := argExpr.(*ast.RecordLiteralExpression); ok && recordLit.TypeName == nil {
					if e.typeSystem.HasRecord(paramType) {
						ctx.SetRecordTypeContext(paramType)
						contextSet = true
					}
				}
			}

			// Evaluate non-lazy parameters
			val := e.Eval(argExpr, ctx)

			if contextSet {
				ctx.ClearRecordTypeContext()
			}

			if isError(val) {
				return nil, fmt.Errorf("error evaluating argument %d: %v", idx+1, val)
			}
			argValues[idx] = val
		}
	}

	return argValues, nil
}

// ResolveOverloadMultiple resolves which overload to call when multiple exist.
//
// This method:
//  1. Evaluates all arguments to determine their types
//  2. Builds semantic Symbol candidates from AST function declarations
//  3. Calls semantic.ResolveOverload to find the best match
//  4. Returns the matching function declaration and cached argument values
//
// Returns an error if no overload matches the provided arguments.
func (e *Evaluator) ResolveOverloadMultiple(
	funcName string,
	overloads []*ast.FunctionDecl,
	argExprs []ast.Expression,
	ctx *ExecutionContext,
) (*ast.FunctionDecl, []Value, error) {
	// 1. Evaluate all arguments to get types
	argTypes := make([]types.Type, len(argExprs))
	argValues := make([]Value, len(argExprs))

	for idx, argExpr := range argExprs {
		// For overload resolution, we need to determine the best matching function
		// first, but we don't know parameter types yet. We evaluate without context
		// initially to determine types.
		val := e.Eval(argExpr, ctx)
		if isError(val) {
			return nil, nil, fmt.Errorf("error evaluating argument %d: %v", idx+1, val)
		}
		argTypes[idx] = e.getValueType(val)
		argValues[idx] = val
	}

	// 2. Build semantic symbols from overloads
	candidates := make([]*semantic.Symbol, len(overloads))
	for idx, fn := range overloads {
		funcType := e.extractFunctionType(fn, ctx)
		if funcType == nil {
			return nil, nil, fmt.Errorf("unable to extract function type for overload %d of '%s'", idx+1, funcName)
		}
		candidates[idx] = &semantic.Symbol{
			Name:                 fn.Name.Value,
			Type:                 funcType,
			HasOverloadDirective: fn.IsOverload,
		}
	}

	// 3. Use semantic analyzer's overload resolution
	selected, err := semantic.ResolveOverload(candidates, argTypes)
	if err != nil {
		return nil, nil, fmt.Errorf("There is no overloaded version of \"%s\" that can be called with these arguments", funcName)
	}

	// 4. Find matching declaration
	selectedType, ok := selected.Type.(*types.FunctionType)
	if !ok {
		return nil, nil, fmt.Errorf("internal error: selected symbol type is not FunctionType")
	}
	for _, fn := range overloads {
		fnType := e.extractFunctionType(fn, ctx)
		if fnType != nil && semantic.SignaturesEqual(fnType, selectedType) &&
			fnType.ReturnType.Equals(selectedType.ReturnType) {
			return fn, argValues, nil
		}
	}

	return nil, nil, fmt.Errorf("internal error: resolved overload not found in candidate list")
}
