package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// raiseContractException creates an exception and sets it in the context via adapter.
func (e *Evaluator) raiseContractException(className, message string, node ast.Node, ctx *ExecutionContext) {
	// Get call stack for exception
	callStack := ctx.CallStack()

	// Look up exception class via TypeSystem (case-insensitive)
	classMetadata := e.typeSystem.LookupClass(className)

	// If class not found, try to fall back to base Exception class
	if classMetadata == nil && className != "Exception" {
		classMetadata = e.typeSystem.LookupClass("Exception")
	}

	// Use adapter bridge constructor to create exception (avoids import cycle)
	exc := e.exceptionMgr.CreateContractException(className, message, node, classMetadata, callStack)

	// Set exception in context
	ctx.SetException(exc)
}

// checkPreconditions evaluates all preconditions of a function.
// If any condition fails, it raises an exception directly.
//
// Returns nil on success, error value if evaluation fails.
func (e *Evaluator) checkPreconditions(funcName string, preConditions *ast.PreConditions, ctx *ExecutionContext) Value {
	if preConditions == nil {
		return nil
	}

	for _, condition := range preConditions.Conditions {
		// Evaluate the test expression using visitor pattern
		result := e.Eval(condition.Test, ctx)

		// Check for evaluation errors
		if isError(result) {
			return result
		}

		// Check that the result is a boolean
		boolVal, ok := result.(*runtime.BooleanValue)
		if !ok {
			return e.newError(condition.Test, "precondition test must evaluate to boolean, got %s", result.Type())
		}

		// If the condition failed, raise an exception
		if !boolVal.Value {
			// Build error message - clean up parentheses from expression
			message := cleanContractMessage(condition.Test.String())

			// Evaluate custom message if provided
			if condition.Message != nil {
				msgVal := e.Eval(condition.Message, ctx)
				if isError(msgVal) {
					return msgVal
				}
				if strVal, ok := msgVal.(*runtime.StringValue); ok {
					message = strVal.Value
				}
			}

			// Format the exception message in DWScript format:
			// "Pre-condition failed in FuncName [line: X, column: Y], condition_expr"
			condPos := condition.Pos()
			fullMessage := fmt.Sprintf("Pre-condition failed in %s [line: %d, column: %d], %s",
				funcName, condPos.Line, condPos.Column, message)

			// Raise exception directly (no adapter!)
			e.raiseContractException("Exception", fullMessage, condition.Test, ctx)
			return nil
		}
	}

	return nil
}

// captureOldValues traverses postconditions to find all OldExpression nodes
// and captures their current values from the environment.
// This must be called BEFORE the function body executes.
func (e *Evaluator) captureOldValues(funcDecl *ast.FunctionDecl, ctx *ExecutionContext) map[string]Value {
	oldValues := make(map[string]Value)

	// If there are no postconditions, no need to capture anything
	if funcDecl.PostConditions == nil {
		return oldValues
	}

	// Traverse all postconditions and find OldExpression nodes
	for _, condition := range funcDecl.PostConditions.Conditions {
		e.findOldExpressions(condition.Test, ctx, oldValues)
		// Note: Message expressions can also contain old expressions
		if condition.Message != nil {
			e.findOldExpressions(condition.Message, ctx, oldValues)
		}
	}

	return oldValues
}

// findOldExpressions recursively searches an expression tree for OldExpression nodes
// and captures their values.
func (e *Evaluator) findOldExpressions(expr ast.Expression, ctx *ExecutionContext, oldValues map[string]Value) {
	if expr == nil {
		return
	}

	switch node := expr.(type) {
	case *ast.OldExpression:
		// Capture the value of this identifier
		identName := node.Identifier.Value
		if _, exists := oldValues[identName]; !exists {
			// Only capture once per identifier
			val, ok := ctx.Env().Get(identName)
			if ok {
				// If the value is a reference (var parameter), dereference it
				// to capture the actual value, not the reference itself
				if refVal, ok := val.(Value); ok {
					if refAccessor, isRef := refVal.(ReferenceAccessor); isRef {
						derefVal, err := refAccessor.Dereference()
						if err != nil {
							// Store nil if dereference fails - error will be reported at evaluation time
							oldValues[identName] = &runtime.NilValue{}
						} else {
							oldValues[identName] = derefVal
						}
					} else {
						oldValues[identName] = refVal
					}
				}
			}
			// If not found, we'll let the runtime evaluation handle the error
		}

	case *ast.BinaryExpression:
		e.findOldExpressions(node.Left, ctx, oldValues)
		e.findOldExpressions(node.Right, ctx, oldValues)

	case *ast.UnaryExpression:
		e.findOldExpressions(node.Right, ctx, oldValues)

	case *ast.CallExpression:
		e.findOldExpressions(node.Function, ctx, oldValues)
		for _, arg := range node.Arguments {
			e.findOldExpressions(arg, ctx, oldValues)
		}

	case *ast.IndexExpression:
		e.findOldExpressions(node.Left, ctx, oldValues)
		e.findOldExpressions(node.Index, ctx, oldValues)

	case *ast.MemberAccessExpression:
		e.findOldExpressions(node.Object, ctx, oldValues)

	case *ast.ArrayLiteralExpression:
		for _, elem := range node.Elements {
			e.findOldExpressions(elem, ctx, oldValues)
		}

	case *ast.SetLiteral:
		for _, elem := range node.Elements {
			e.findOldExpressions(elem, ctx, oldValues)
		}

	case *ast.RecordLiteralExpression:
		for _, field := range node.Fields {
			e.findOldExpressions(field.Value, ctx, oldValues)
		}

	case *ast.LambdaExpression:
		// Lambda bodies can't reference old from outer postconditions
		// (they have their own scope), so we don't traverse them
		return

	// Literals and identifiers don't contain old expressions
	case *ast.IntegerLiteral, *ast.FloatLiteral, *ast.StringLiteral,
		*ast.BooleanLiteral, *ast.Identifier, *ast.NilLiteral:
		return

		// Add more cases as needed for other expression types
	}
}

// checkPostconditions evaluates all postconditions of a function.
// If any condition fails, it raises an exception.
// This must be called AFTER the function body executes, with oldValues available.
//
// Returns nil on success, error value if evaluation fails.
func (e *Evaluator) checkPostconditions(funcName string, postConditions *ast.PostConditions, ctx *ExecutionContext) Value {
	if postConditions == nil {
		return nil
	}

	for _, condition := range postConditions.Conditions {
		// Evaluate the test expression (can reference old values via ctx.GetOldValue)
		result := e.Eval(condition.Test, ctx)

		// Check for evaluation errors
		if isError(result) {
			return result
		}

		// Check that the result is a boolean
		boolVal, ok := result.(*runtime.BooleanValue)
		if !ok {
			return e.newError(condition.Test, "postcondition test must evaluate to boolean, got %s", result.Type())
		}

		// If the condition failed, raise an exception
		if !boolVal.Value {
			// Build error message - clean up parentheses from expression
			message := cleanContractMessage(condition.Test.String())

			// Evaluate custom message if provided
			if condition.Message != nil {
				msgVal := e.Eval(condition.Message, ctx)
				if isError(msgVal) {
					return msgVal
				}
				if strVal, ok := msgVal.(*runtime.StringValue); ok {
					message = strVal.Value
				}
			}

			// Format the exception message in DWScript format:
			// "Post-condition failed in FuncName [line: X, column: Y], condition_expr"
			condPos := condition.Pos()
			fullMessage := fmt.Sprintf("Post-condition failed in %s [line: %d, column: %d], %s",
				funcName, condPos.Line, condPos.Column, message)

			// Raise exception directly (no adapter!)
			e.raiseContractException("Exception", fullMessage, condition.Test, ctx)
			return nil
		}
	}

	return nil
}
