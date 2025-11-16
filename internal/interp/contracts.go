package interp

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// raiseException raises an exception with the given class name and message.
func (i *Interpreter) raiseException(className, message string, pos *lexer.Position) {
	// Get the exception class
	// PR #147: Use lowercase key for O(1) case-insensitive lookup
	excClass, ok := i.classes[strings.ToLower(className)]
	if !ok {
		// Fallback to base Exception if class not found
		excClass, ok = i.classes[strings.ToLower("Exception")]
		if !ok {
			// This shouldn't happen, but handle it gracefully
			i.exception = &ExceptionValue{
				ClassInfo: NewClassInfo(className),
				Instance:  nil,
				Message:   message,
				Position:  pos,
				CallStack: i.callStack,
			}
			return
		}
	}

	// Create an instance of the exception class
	instance := NewObjectInstance(excClass)
	instance.SetField("Message", &StringValue{Value: message})

	// Set the exception
	i.exception = &ExceptionValue{
		ClassInfo: excClass,
		Instance:  instance,
		Message:   message,
		Position:  pos,
		CallStack: i.callStack,
	}
}

// captureOldValues traverses postconditions to find all OldExpression nodes
// and captures their current values from the environment.
// This must be called BEFORE the function body executes.
func (i *Interpreter) captureOldValues(funcDecl *ast.FunctionDecl, env *Environment) map[string]Value {
	oldValues := make(map[string]Value)

	// If there are no postconditions, no need to capture anything
	if funcDecl.PostConditions == nil {
		return oldValues
	}

	// Traverse all postconditions and find OldExpression nodes
	for _, condition := range funcDecl.PostConditions.Conditions {
		i.findOldExpressions(condition.Test, env, oldValues)
		// Note: Message expressions can also contain old expressions
		if condition.Message != nil {
			i.findOldExpressions(condition.Message, env, oldValues)
		}
	}

	return oldValues
}

// findOldExpressions recursively searches an expression tree for OldExpression nodes
// and captures their values.
func (i *Interpreter) findOldExpressions(expr ast.Expression, env *Environment, oldValues map[string]Value) {
	if expr == nil {
		return
	}

	switch node := expr.(type) {
	case *ast.OldExpression:
		// Capture the value of this identifier
		identName := node.Identifier.Value
		if _, exists := oldValues[identName]; !exists {
			// Only capture once per identifier
			val, ok := env.Get(identName)
			if ok {
				oldValues[identName] = val
			}
			// If not found, we'll let the runtime evaluation handle the error
		}

	case *ast.BinaryExpression:
		i.findOldExpressions(node.Left, env, oldValues)
		i.findOldExpressions(node.Right, env, oldValues)

	case *ast.UnaryExpression:
		i.findOldExpressions(node.Right, env, oldValues)

	case *ast.CallExpression:
		i.findOldExpressions(node.Function, env, oldValues)
		for _, arg := range node.Arguments {
			i.findOldExpressions(arg, env, oldValues)
		}

	case *ast.IndexExpression:
		i.findOldExpressions(node.Left, env, oldValues)
		i.findOldExpressions(node.Index, env, oldValues)

	case *ast.MemberAccessExpression:
		i.findOldExpressions(node.Object, env, oldValues)

	case *ast.ArrayLiteralExpression:
		for _, elem := range node.Elements {
			i.findOldExpressions(elem, env, oldValues)
		}

	case *ast.SetLiteral:
		for _, elem := range node.Elements {
			i.findOldExpressions(elem, env, oldValues)
		}

	case *ast.RecordLiteralExpression:
		for _, field := range node.Fields {
			i.findOldExpressions(field.Value, env, oldValues)
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

// checkPreconditions evaluates all preconditions of a function.
// If any condition fails, it raises an exception.
func (i *Interpreter) checkPreconditions(funcName string, preConditions *ast.PreConditions, env *Environment) Value {
	if preConditions == nil {
		return nil
	}

	for _, condition := range preConditions.Conditions {
		// Evaluate the test expression
		result := i.Eval(condition.Test)

		// Check for evaluation errors
		if isError(result) {
			return result
		}

		// Check that the result is a boolean
		boolVal, ok := result.(*BooleanValue)
		if !ok {
			return newError("precondition test must evaluate to boolean, got %s", result.Type())
		}

		// If the condition failed, raise an exception
		if !boolVal.Value {
			// Build error message
			message := condition.Test.String()

			// Evaluate custom message if provided
			if condition.Message != nil {
				msgVal := i.Eval(condition.Message)
				if isError(msgVal) {
					return msgVal
				}
				if strVal, ok := msgVal.(*StringValue); ok {
					message = strVal.Value
				}
			}

			// Format the exception message in DWScript format:
			// "Pre-condition failed in FuncName [line: X, column: Y], condition_expr"
			fullMessage := fmt.Sprintf("Pre-condition failed in %s [line: %d, column: %d], %s",
				funcName, condition.Token.Pos.Line, condition.Token.Pos.Column, message)

			// Raise an Exception
			i.raiseException("Exception", fullMessage, &condition.Token.Pos)
			return nil
		}
	}

	return nil
}

// checkPostconditions evaluates all postconditions of a function.
// If any condition fails, it raises an exception.
// This must be called AFTER the function body executes, with oldValues available.
func (i *Interpreter) checkPostconditions(funcName string, postConditions *ast.PostConditions, env *Environment) Value {
	if postConditions == nil {
		return nil
	}

	for _, condition := range postConditions.Conditions {
		// Evaluate the test expression (can reference old values via oldValuesStack)
		result := i.Eval(condition.Test)

		// Check for evaluation errors
		if isError(result) {
			return result
		}

		// Check that the result is a boolean
		boolVal, ok := result.(*BooleanValue)
		if !ok {
			return newError("postcondition test must evaluate to boolean, got %s", result.Type())
		}

		// If the condition failed, raise an exception
		if !boolVal.Value {
			// Build error message
			message := condition.Test.String()

			// Evaluate custom message if provided
			if condition.Message != nil {
				msgVal := i.Eval(condition.Message)
				if isError(msgVal) {
					return msgVal
				}
				if strVal, ok := msgVal.(*StringValue); ok {
					message = strVal.Value
				}
			}

			// Format the exception message in DWScript format:
			// "Post-condition failed in FuncName [line: X, column: Y], condition_expr"
			fullMessage := fmt.Sprintf("Post-condition failed in %s [line: %d, column: %d], %s",
				funcName, condition.Token.Pos.Line, condition.Token.Pos.Column, message)

			// Raise an Exception
			i.raiseException("Exception", fullMessage, &condition.Token.Pos)
			return nil
		}
	}

	return nil
}

// pushOldValues pushes captured old values onto the stack for postcondition evaluation.
func (i *Interpreter) pushOldValues(oldValues map[string]Value) {
	i.oldValuesStack = append(i.oldValuesStack, oldValues)
}

// popOldValues removes the top old values map from the stack.
func (i *Interpreter) popOldValues() {
	if len(i.oldValuesStack) > 0 {
		i.oldValuesStack = i.oldValuesStack[:len(i.oldValuesStack)-1]
	}
}

// getOldValue retrieves a captured old value by identifier name.
// Returns the value and true if found, or nil and false if not found.
func (i *Interpreter) getOldValue(identName string) (Value, bool) {
	// Check the top of the stack (most recent function call)
	if len(i.oldValuesStack) > 0 {
		topMap := i.oldValuesStack[len(i.oldValuesStack)-1]
		if val, exists := topMap[identName]; exists {
			return val, true
		}
	}
	return nil, false
}
