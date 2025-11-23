package evaluator

import (
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// This file contains helper methods for call expression evaluation.
// Task 3.5.52: Migrated from adapter.EvalCallExpression calls to inline Evaluator logic.

// builtinsWithVarParams lists built-in functions that require var parameter handling.
// These functions modify their first argument in place.
var builtinsWithVarParams = map[string]bool{
	"inc":            true,
	"dec":            true,
	"insert":         true,
	"decodedate":     true,
	"decodetime":     true,
	"swap":           true,
	"divmod":         true,
	"trystrtoint":    true,
	"trystrtofloat":  true,
	"setlength":      true,
}

// isBuiltinWithVarParam checks if a function name is a builtin with var parameters.
func isBuiltinWithVarParam(name string) bool {
	return builtinsWithVarParams[ident.Normalize(name)]
}

// isDeleteWithVarParam checks if a Delete call has 3 arguments (the var param form).
// Delete(str, pos, count) modifies str in place.
func isDeleteWithVarParam(name string, argCount int) bool {
	return ident.Equal(name, "delete") && argCount == 3
}

// isDefaultFunction checks if the call is a Default(TypeName) function.
func isDefaultFunction(name string, argCount int) bool {
	return ident.Equal(name, "default") && argCount == 1
}

// isSingleArgCall checks if a call has exactly one argument.
// Used to identify potential type casts.
func isSingleArgCall(argCount int) bool {
	return argCount == 1
}

// evalFunctionPointerCall evaluates a function pointer call.
// Task 3.5.52: Handles lazy params, var params, and closure restoration.
func (e *Evaluator) evalFunctionPointerCall(
	node *ast.CallExpression,
	funcPtr Value,
	ctx *ExecutionContext,
) Value {
	// Delegate to adapter which handles:
	// - Closure environment restoration
	// - Lazy parameter creation (CreateLazyThunk)
	// - Var parameter creation (CreateReferenceValue)
	// - Regular parameter evaluation
	return e.adapter.CallFunctionPointerWithArgs(funcPtr, node.Arguments, ctx)
}

// evalRecordMethodCall evaluates a method call on a record value.
// Task 3.5.52: Delegates to adapter for record method dispatch.
func (e *Evaluator) evalRecordMethodCall(
	node *ast.CallExpression,
	memberAccess *ast.MemberAccessExpression,
	objVal Value,
	ctx *ExecutionContext,
) Value {
	return e.adapter.CallRecordMethod(objVal, memberAccess.Member.Value, node.Arguments, ctx)
}

// evalInterfaceMethodCall evaluates a method call on an interface value.
// Task 3.5.52: Delegates to adapter for interface method dispatch.
func (e *Evaluator) evalInterfaceMethodCall(
	node *ast.CallExpression,
	memberAccess *ast.MemberAccessExpression,
	objVal Value,
	ctx *ExecutionContext,
) Value {
	return e.adapter.CallInterfaceMethod(objVal, memberAccess.Member.Value, node.Arguments, ctx)
}

// evalObjectMethodCall evaluates a method call on an object value.
// Task 3.5.52: Delegates to adapter for object method dispatch.
func (e *Evaluator) evalObjectMethodCall(
	node *ast.CallExpression,
	memberAccess *ast.MemberAccessExpression,
	objVal Value,
	ctx *ExecutionContext,
) Value {
	return e.adapter.CallObjectMethod(objVal, memberAccess.Member.Value, node.Arguments, ctx)
}

// evalUnitQualifiedCall evaluates a unit-qualified function call (UnitName.FunctionName).
// Task 3.5.52: Delegates to adapter for unit function resolution and calling.
func (e *Evaluator) evalUnitQualifiedCall(
	node *ast.CallExpression,
	unitName string,
	funcName string,
	ctx *ExecutionContext,
) Value {
	return e.adapter.CallUnitFunction(unitName, funcName, node.Arguments, ctx)
}

// evalClassConstructorCall evaluates a class constructor call (TClass.Create).
// Task 3.5.52: Delegates to adapter for constructor resolution and calling.
func (e *Evaluator) evalClassConstructorCall(
	node *ast.CallExpression,
	className string,
	methodName string,
	ctx *ExecutionContext,
) Value {
	return e.adapter.CallClassMethod(className, methodName, node.Arguments, ctx)
}

// evalOverloadedFunctionCall evaluates a user-defined function call with potential overloads.
// Task 3.5.52: Delegates to adapter for overload resolution.
func (e *Evaluator) evalOverloadedFunctionCall(
	node *ast.CallExpression,
	funcName string,
	ctx *ExecutionContext,
) Value {
	return e.adapter.CallUserFunctionWithOverloads(funcName, node.Arguments, ctx)
}

// evalImplicitSelfMethodCall evaluates a method call using implicit Self.
// Task 3.5.52: When inside an instance method, MethodName() means Self.MethodName().
func (e *Evaluator) evalImplicitSelfMethodCall(
	node *ast.CallExpression,
	funcName *ast.Identifier,
	selfVal Value,
	ctx *ExecutionContext,
) Value {
	return e.adapter.CallImplicitSelfMethod(selfVal, funcName.Value, node.Arguments, ctx)
}

// evalRecordStaticMethodCall evaluates a static method call within a record context.
// Task 3.5.52: When inside a record static method, FuncName() calls TRecord.FuncName().
func (e *Evaluator) evalRecordStaticMethodCall(
	node *ast.CallExpression,
	funcName *ast.Identifier,
	recordVal Value,
	ctx *ExecutionContext,
) Value {
	return e.adapter.CallRecordStaticMethod(recordVal, funcName.Value, node.Arguments, ctx)
}

// evalBuiltinWithVarParam evaluates a builtin function that modifies its argument.
// Task 3.5.52: Handles Inc, Dec, Swap, SetLength, etc.
func (e *Evaluator) evalBuiltinWithVarParam(
	node *ast.CallExpression,
	funcName string,
	ctx *ExecutionContext,
) Value {
	return e.adapter.CallBuiltinWithVarParam(funcName, node.Arguments, ctx)
}

// evalExternalFunctionCall evaluates an external (Go) function call.
// Task 3.5.52: External functions may have var parameters.
func (e *Evaluator) evalExternalFunctionCall(
	node *ast.CallExpression,
	funcName string,
	ctx *ExecutionContext,
) Value {
	return e.adapter.CallExternalFunction(funcName, node.Arguments, ctx)
}

// evalDefaultFunctionCall evaluates a Default(TypeName) call.
// Task 3.5.52: Default expects unevaluated type identifier.
func (e *Evaluator) evalDefaultFunctionCall(
	node *ast.CallExpression,
	ctx *ExecutionContext,
) Value {
	if len(node.Arguments) != 1 {
		return e.newError(node, "Default() requires exactly one argument")
	}
	return e.adapter.EvalDefaultFunction(node.Arguments[0], ctx)
}

// evalTypeCastCall evaluates a type cast call (TypeName(expression)).
// Task 3.5.52: Type casts look like function calls but convert types.
func (e *Evaluator) evalTypeCastCall(
	node *ast.CallExpression,
	typeName string,
	ctx *ExecutionContext,
) Value {
	if len(node.Arguments) != 1 {
		return nil // Not a type cast
	}
	return e.adapter.EvalTypeCast(typeName, node.Arguments[0], ctx)
}

// evalBuiltinFunctionCall evaluates a standard built-in function call.
// Task 3.5.52: Evaluates all arguments first, then calls the builtin.
func (e *Evaluator) evalBuiltinFunctionCall(
	node *ast.CallExpression,
	funcName string,
	ctx *ExecutionContext,
) Value {
	// Evaluate all arguments
	args := make([]Value, len(node.Arguments))
	for idx, arg := range node.Arguments {
		val := e.Eval(arg, ctx)
		if isError(val) {
			return val
		}
		args[idx] = val
	}

	return e.adapter.CallBuiltinFunction(funcName, args)
}
