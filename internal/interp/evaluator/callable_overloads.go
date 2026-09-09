package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
)

// selectCallableOverload keeps the selected canonical binding through dispatch.
func (e *Evaluator) selectCallableOverload(className, methodName string, overloads []*runtime.MethodMetadata, args []Value, ctx *ExecutionContext) (*runtime.MethodMetadata, error) {
	if len(overloads) == 0 {
		return nil, fmt.Errorf("method '%s' not found in class '%s'", methodName, className)
	}
	if len(overloads) == 1 {
		return overloads[0], nil
	}
	argumentTypes := make([]types.Type, len(args))
	for i, argument := range args {
		argumentTypes[i] = e.runtimeValueType(argument)
	}
	candidates := make([]types.Type, 0, len(overloads))
	bindings := make([]*runtime.MethodMetadata, 0, len(overloads))
	for _, method := range overloads {
		if signature := e.callableFunctionType(method, ctx); signature != nil {
			candidates = append(candidates, signature)
			bindings = append(bindings, method)
		}
	}
	if len(candidates) != 0 {
		if selected, err := types.ResolveOverload(candidates, argumentTypes); err == nil {
			return bindings[selected], nil
		}
	}
	// Preserve unchecked execution's argument-count and default-parameter fallback.
	for _, method := range overloads {
		if method != nil && method.ParamCount() == len(args) {
			return method, nil
		}
	}
	for _, method := range overloads {
		if method != nil && len(args) <= method.ParamCount() && len(args) >= method.RequiredParamCount() {
			return method, nil
		}
	}
	return overloads[0], nil
}

func (e *Evaluator) callableFunctionType(method *runtime.MethodMetadata, ctx *ExecutionContext) *types.FunctionType {
	if method == nil {
		return nil
	}
	parameters := make([]types.Type, len(method.Parameters))
	names := make([]string, len(parameters))
	defaults := make([]any, len(parameters))
	lazy := make([]bool, len(parameters))
	byRef := make([]bool, len(parameters))
	constant := make([]bool, len(parameters))
	for i, parameter := range method.Parameters {
		if parameter.Type == nil {
			if method.Declaration != nil {
				return e.extractMethodType(method.Declaration, ctx)
			}
			return nil
		}
		parameters[i], names[i], defaults[i] = parameter.Type, parameter.Name, parameter.DefaultValue
		lazy[i], byRef[i], constant[i] = parameter.IsLazy, parameter.ByRef, parameter.IsConst
	}
	result := method.ReturnType
	if result == nil {
		if method.IsFunction() && method.Declaration != nil {
			return e.extractMethodType(method.Declaration, ctx)
		}
		result = types.VOID
	}
	return types.NewFunctionTypeWithMetadata(parameters, names, defaults, lazy, byRef, constant, result)
}
