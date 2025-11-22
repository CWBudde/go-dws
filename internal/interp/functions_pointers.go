package interp

import (
	"strings"

	"github.com/cwbudde/go-dws/pkg/ast"
)

// callFunctionPointer executes a function pointer call, handling lambdas, method pointers, and regular function pointers.
// It sets up closure environments for lambdas and binds Self for method pointers as needed.
func (i *Interpreter) callFunctionPointer(funcPtr *FunctionPointerValue, args []Value, node ast.Node) Value {
	// Enhanced to handle lambda closures

	// Check if this is a lambda or a regular function pointer
	if funcPtr.Lambda != nil {
		// Lambda closure - call with closure environment
		closure, ok := funcPtr.Closure.(*Environment)
		if !ok {
			return i.newErrorWithLocation(node, "invalid closure type in function pointer")
		}
		return i.callLambda(funcPtr.Lambda, closure, args, node)
	}

	// Regular function pointer
	if funcPtr.Function == nil {
		return i.newErrorWithLocation(node, "function pointer is nil")
	}

	// If this is a method pointer, we need to set up the Self binding
	if funcPtr.SelfObject != nil {
		// Create a new environment with Self bound
		funcEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = funcEnv

		// Bind Self to the captured object
		i.env.Define("Self", funcPtr.SelfObject)

		// Call the function
		result := i.callUserFunction(funcPtr.Function, args)

		// Restore environment
		i.env = savedEnv

		return result
	}

	// Regular function pointer - just call the function directly
	return i.callUserFunction(funcPtr.Function, args)
}

// callLambda executes a lambda expression with its captured closure environment.
//
// The key difference from regular functions is that lambdas execute within their
// closure environment, allowing them to access captured variables from outer scopes.
//
// Parameters:
//   - lambda: The lambda expression AST node
//   - closureEnv: The environment captured when the lambda was created
//   - args: The argument values passed to the lambda
//   - node: AST node for error reporting
//
// Variable Capture Semantics:
//   - Captured variables are accessed by reference (not copied)
//   - Changes to captured variables inside the lambda affect the outer scope
//   - The environment chain naturally provides this behavior
func (i *Interpreter) callLambda(lambda *ast.LambdaExpression, closureEnv *Environment, args []Value, node ast.Node) Value {
	// Check argument count matches parameter count
	if len(args) != len(lambda.Parameters) {
		return i.newErrorWithLocation(node, "wrong number of arguments for lambda: expected %d, got %d",
			len(lambda.Parameters), len(args))
	}

	// Create a new environment for the lambda scope
	// CRITICAL: Use closureEnv as parent, NOT i.env
	// This gives the lambda access to captured variables
	lambdaEnv := NewEnclosedEnvironment(closureEnv)
	savedEnv := i.env
	i.env = lambdaEnv

	// Check recursion depth before pushing to call stack
	if i.ctx.GetCallStack().WillOverflow() {
		i.env = savedEnv // Restore environment before raising exception
		return i.raiseMaxRecursionExceeded()
	}

	// Push lambda marker onto call stack for stack traces
	i.pushCallStack("<lambda>")
	defer i.popCallStack()

	// Bind parameters to arguments
	for idx, param := range lambda.Parameters {
		arg := args[idx]

		// Apply implicit conversion if parameter has a type and types don't match
		if param.Type != nil {
			paramTypeName := param.Type.String()
			if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
				arg = converted
			}
		}

		// Note: Lambdas don't support by-ref parameters (for now)
		// All parameters are by-value
		i.env.Define(param.Name.Value, arg)
	}

	// For functions (not procedures), initialize the Result variable
	if lambda.ReturnType != nil {
		// Initialize Result based on return type with appropriate defaults
		returnType := i.resolveTypeFromAnnotation(lambda.ReturnType)
		var resultValue = i.getDefaultValue(returnType)

		// Check if return type is a record (overrides default)
		returnTypeName := lambda.ReturnType.String()
		recordTypeKey := "__record_type_" + strings.ToLower(returnTypeName)
		if typeVal, ok := i.env.Get(recordTypeKey); ok {
			if rtv, ok := typeVal.(*RecordTypeValue); ok {
				// Use createRecordValue for proper nested record initialization
				resultValue = i.createRecordValue(rtv.RecordType, rtv.Methods)
			}
		}

		i.env.Define("Result", resultValue)
	}

	// Execute the lambda body
	bodyResult := i.Eval(lambda.Body)

	// If an error occurred during execution, propagate it
	if isError(bodyResult) {
		i.env = savedEnv
		return bodyResult
	}

	// If an exception was raised during lambda execution, propagate it immediately
	if i.exception != nil {
		i.env = savedEnv
		return &NilValue{}
	}

	// Handle exit signal
	if i.ctx.ControlFlow().IsExit() {
		i.ctx.ControlFlow().Clear()
	}

	// Extract return value
	var returnValue Value
	if lambda.ReturnType != nil {
		// Lambda has a return type - get the Result value
		resultVal, resultOk := i.env.Get("Result")

		if resultOk && resultVal.Type() != "NIL" {
			returnValue = resultVal
		} else if resultOk {
			returnValue = resultVal
		} else {
			returnValue = &NilValue{}
		}
	} else {
		// Procedure lambda - no return value
		returnValue = &NilValue{}
	}

	// Restore environment
	i.env = savedEnv

	return returnValue
}

// evalRecordMethodCall evaluates a method call on a record value (record.Method(...)).
//
// Records are value types in DWScript (unlike classes which are reference types).
// This means:
//   - Methods execute with Self bound to the record instance
//   - For mutating methods (procedures that modify Self), we need copy-back semantics
//   - No inheritance - simple method lookup in RecordType.Methods
//
// Parameters:
//   - recVal: The record instance to call the method on
//   - memberAccess: The member access expression containing the method name
//   - argExprs: The argument expressions to evaluate
//
// Example:
//
//	type TPoint = record
//	  X, Y: Integer;
//	  function Distance: Float;
//	  begin
//	    Result := Sqrt(X*X + Y*Y);
//	  end;
//	  procedure Move(dx, dy: Integer);
//	  begin
//	    X := X + dx;
//	    Y := Y + dy;
//	  end;
//	end;
//
//	var p: TPoint;
//	p.X := 3; p.Y := 4;
//	var d := p.Distance();  // Returns 5.0
//	p.Move(1, 1);           // Modifies p to (4, 5)
