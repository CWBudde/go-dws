package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/pkg/ast"
)

// EvaluateDefaultParameters fills in missing optional arguments with default values.
// Task 3.5.144b.1: Extract default parameter evaluation from callUserFunction.
//
// This method:
//  1. Validates argument count against required/optional parameters
//  2. Evaluates default expressions for missing arguments in the caller's context
//  3. Returns a complete slice of argument values
//
// Default values are evaluated in the caller's environment (ctx), not a new scope.
// This is important for defaults that reference variables in the caller's scope.
//
// Returns an error if:
//   - Too few arguments (less than required parameters)
//   - Too many arguments (more than total parameters)
//   - Default value evaluation fails
func (e *Evaluator) EvaluateDefaultParameters(
	fn *ast.FunctionDecl,
	args []Value,
	ctx *ExecutionContext,
) ([]Value, error) {
	// Calculate required parameter count (parameters without defaults)
	requiredParams := 0
	for _, param := range fn.Parameters {
		if param.DefaultValue == nil {
			requiredParams++
		}
	}

	// Check argument count is within valid range
	if len(args) < requiredParams {
		return nil, fmt.Errorf("wrong number of arguments: expected at least %d, got %d",
			requiredParams, len(args))
	}
	if len(args) > len(fn.Parameters) {
		return nil, fmt.Errorf("wrong number of arguments: expected at most %d, got %d",
			len(fn.Parameters), len(args))
	}

	// If all arguments provided, return as-is
	if len(args) == len(fn.Parameters) {
		return args, nil
	}

	// Fill in missing optional arguments with default values
	// Evaluate default expressions in the caller's environment (ctx)
	result := make([]Value, len(fn.Parameters))
	copy(result, args)

	for idx := len(args); idx < len(fn.Parameters); idx++ {
		param := fn.Parameters[idx]
		if param.DefaultValue == nil {
			// This should never happen due to requiredParams check above
			return nil, fmt.Errorf("internal error: missing required parameter at index %d", idx)
		}

		// Evaluate default value in caller's environment
		defaultVal := e.Eval(param.DefaultValue, ctx)
		if isError(defaultVal) {
			return nil, fmt.Errorf("error evaluating default for parameter '%s': %v",
				param.Name.Value, defaultVal)
		}
		result[idx] = defaultVal
	}

	return result, nil
}
