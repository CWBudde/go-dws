package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/ast"
)

// This file contains visitor methods for expression AST nodes.
// Phase 3.5.2: Visitor pattern implementation for expressions.
//
// Expressions evaluate to values and can be nested (e.g., binary expressions
// contain left and right sub-expressions).

// ErrorValue represents a runtime error (temporary definition to avoid circular imports).
type ErrorValue struct {
	Message string
}

func (e *ErrorValue) Type() string   { return "ERROR" }
func (e *ErrorValue) String() string { return "ERROR: " + e.Message }

// newError creates a new error value with optional formatting.
// TODO: Add location information from node in Phase 3.6 (error handling improvements)
func (e *Evaluator) newError(_ ast.Node, format string, args ...interface{}) Value {
	return &ErrorValue{Message: fmt.Sprintf(format, args...)}
}

// isError checks if a value is an error.
func isError(val Value) bool {
	if val != nil {
		return val.Type() == "ERROR"
	}
	return false
}

// VisitIdentifier evaluates an identifier (variable reference).
func (e *Evaluator) VisitIdentifier(node *ast.Identifier, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	// Future: Move variable lookup logic here
	return e.adapter.EvalNode(node)
}

// VisitBinaryExpression evaluates a binary expression (e.g., a + b, x == y).
func (e *Evaluator) VisitBinaryExpression(node *ast.BinaryExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	// Future: Move operator evaluation logic here
	return e.adapter.EvalNode(node)
}

// VisitUnaryExpression evaluates a unary expression (e.g., -x, not b).
func (e *Evaluator) VisitUnaryExpression(node *ast.UnaryExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	return e.adapter.EvalNode(node)
}

// VisitAddressOfExpression evaluates an address-of expression (@funcName).
func (e *Evaluator) VisitAddressOfExpression(node *ast.AddressOfExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2A: Function lookup is available via adapter.LookupFunction
	// TODO: Migrate evalAddressOfExpression logic to use adapter.LookupFunction
	return e.adapter.EvalNode(node)
}

// VisitGroupedExpression evaluates a grouped expression (parenthesized).
func (e *Evaluator) VisitGroupedExpression(node *ast.GroupedExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.4.11: Grouped expressions just evaluate their inner expression
	// Parentheses are only for precedence, they don't change the value
	return e.Eval(node.Expression, ctx)
}

// VisitCallExpression evaluates a function call expression.
func (e *Evaluator) VisitCallExpression(node *ast.CallExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2A: Function call infrastructure is available via adapter
	// (CallFunctionPointer, CallUserFunction, CallBuiltinFunction, LookupFunction)
	// TODO: Migrate evalCallExpression logic from Interpreter to use these adapter methods
	return e.adapter.EvalNode(node)
}

// VisitNewExpression evaluates a 'new' expression (object instantiation).
func (e *Evaluator) VisitNewExpression(node *ast.NewExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	return e.adapter.EvalNode(node)
}

// VisitMemberAccessExpression evaluates member access (obj.field, obj.method).
func (e *Evaluator) VisitMemberAccessExpression(node *ast.MemberAccessExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	// Future: Move member access logic here
	return e.adapter.EvalNode(node)
}

// VisitMethodCallExpression evaluates a method call (obj.Method(args)).
func (e *Evaluator) VisitMethodCallExpression(node *ast.MethodCallExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2A: Function call infrastructure is available via adapter
	// TODO: Migrate method call logic to use adapter.CallUserFunction
	return e.adapter.EvalNode(node)
}

// VisitInheritedExpression evaluates an 'inherited' expression.
func (e *Evaluator) VisitInheritedExpression(node *ast.InheritedExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2A: Function call infrastructure is available via adapter
	// TODO: Migrate inherited call logic to use adapter.CallUserFunction
	return e.adapter.EvalNode(node)
}

// VisitSelfExpression evaluates a 'Self' expression.
func (e *Evaluator) VisitSelfExpression(node *ast.SelfExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2A: Self context will be needed for method calls
	// TODO: Add Self context to ExecutionContext or Environment
	return e.adapter.EvalNode(node)
}

// VisitEnumLiteral evaluates an enum literal (EnumType.Value).
func (e *Evaluator) VisitEnumLiteral(node *ast.EnumLiteral, ctx *ExecutionContext) Value {
	// Phase 3.5.4.12: Enum literals are looked up in the environment
	// The semantic analyzer validates enum types and values exist
	if node == nil {
		return e.newError(node, "nil enum literal")
	}

	valueName := node.ValueName

	// Look up the value in the environment
	val, ok := ctx.Env().Get(valueName)
	if !ok {
		return e.newError(node, "undefined enum value '%s'", valueName)
	}

	// Environment stores interface{}, cast to Value
	// The semantic analyzer ensures this is a valid enum value
	if value, ok := val.(Value); ok {
		return value
	}

	// Should never happen if semantic analysis passed
	return e.newError(node, "enum value '%s' has invalid type", valueName)
}

// VisitRecordLiteralExpression evaluates a record literal expression.
func (e *Evaluator) VisitRecordLiteralExpression(node *ast.RecordLiteralExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	return e.adapter.EvalNode(node)
}

// VisitSetLiteral evaluates a set literal [value1, value2, ...].
func (e *Evaluator) VisitSetLiteral(node *ast.SetLiteral, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	return e.adapter.EvalNode(node)
}

// VisitArrayLiteralExpression evaluates an array literal [1, 2, 3].
func (e *Evaluator) VisitArrayLiteralExpression(node *ast.ArrayLiteralExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	return e.adapter.EvalNode(node)
}

// VisitIndexExpression evaluates an index expression array[index].
func (e *Evaluator) VisitIndexExpression(node *ast.IndexExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	return e.adapter.EvalNode(node)
}

// VisitNewArrayExpression evaluates a new array expression.
func (e *Evaluator) VisitNewArrayExpression(node *ast.NewArrayExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	return e.adapter.EvalNode(node)
}

// VisitLambdaExpression evaluates a lambda expression (closure).
func (e *Evaluator) VisitLambdaExpression(node *ast.LambdaExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2A: Lambda creation needs to capture closure environment
	// TODO: Migrate lambda creation logic to create FunctionPointerValue with closure
	return e.adapter.EvalNode(node)
}

// VisitIsExpression evaluates an 'is' type checking expression.
func (e *Evaluator) VisitIsExpression(node *ast.IsExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	return e.adapter.EvalNode(node)
}

// VisitAsExpression evaluates an 'as' type casting expression.
func (e *Evaluator) VisitAsExpression(node *ast.AsExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	return e.adapter.EvalNode(node)
}

// VisitImplementsExpression evaluates an 'implements' interface checking expression.
func (e *Evaluator) VisitImplementsExpression(node *ast.ImplementsExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	return e.adapter.EvalNode(node)
}

// VisitIfExpression evaluates an inline if-then-else expression.
func (e *Evaluator) VisitIfExpression(node *ast.IfExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	return e.adapter.EvalNode(node)
}

// VisitOldExpression evaluates an 'old' expression (used in postconditions).
func (e *Evaluator) VisitOldExpression(node *ast.OldExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	return e.adapter.EvalNode(node)
}
