package evaluator

import (
	"github.com/cwbudde/go-dws/pkg/ast"
)

// callRecordMethod executes a record method in the evaluator.
//
// Record methods are user-defined methods attached to record types.
// They execute with Self bound to the record instance.
//
// Execution semantics:
// - Create new environment (child of current)
// - Bind Self to record instance
// - Bind method parameters from args
// - Initialize Result variable (if method has return type)
// - Execute method body
// - Extract and return Result
func (e *Evaluator) callRecordMethod(
	record RecordInstanceValue,
	method *ast.FunctionDecl,
	args []Value,
	node ast.Node,
	ctx *ExecutionContext,
) Value {
	// 1. Validate parameter count
	if len(args) != len(method.Parameters) {
		return e.newError(node,
			"method '%s' expects %d parameters, got %d",
			method.Name.Value, len(method.Parameters), len(args))
	}

	// 2. Create method environment (child of current context)
	ctx.PushEnv()
	defer ctx.PopEnv()

	// 3. Bind Self to record instance
	// This allows the method body to access record fields via Self.FieldName
	ctx.Env().Define("Self", record)

	// 4. Bind method parameters
	for i, param := range method.Parameters {
		paramName := param.Name.Value
		ctx.Env().Define(paramName, args[i])
	}

	// 5. Initialize Result variable
	// DWScript uses implicit Result variable for function return values
	if method.ReturnType != nil {
		// Resolve return type and get default value
		returnType, err := e.ResolveTypeFromAnnotation(method.ReturnType)
		if err != nil {
			return e.newError(node, "failed to resolve return type: %v", err)
		}
		defaultVal := e.GetDefaultValue(returnType)
		ctx.Env().Define("Result", defaultVal)
		// Also define method name as alias for Result (Pascal convention)
		ctx.Env().Define(method.Name.Value, defaultVal)
	}

	// 6. Execute method body in new environment
	result := e.Eval(method.Body, ctx)
	if isError(result) {
		return result
	}

	// 7. Check for early return/exit
	// Control flow is managed by ExecutionContext.ControlFlow()
	if ctx.ControlFlow().IsExit() {
		// Exit propagates up the call stack
		return e.nilValue()
	}

	// 8. Extract Result variable (if method has return type)
	if method.ReturnType != nil {
		return e.extractReturnValue(method.Name.Value, ctx)
	}

	// Procedure (no return type) - return nil
	return e.nilValue()
}
