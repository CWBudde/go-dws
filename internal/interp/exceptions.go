package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Exception Value Representation
// ============================================================================

// ExceptionValue represents an exception object at runtime.
// It holds the exception class type, the message, and the call stack at the point of raise.
type ExceptionValue struct {
	ClassInfo *ClassInfo
	Instance  *ObjectInstance
	Message   string
	CallStack []string // Function names in the call stack when exception was raised
}

// Type returns the type of this exception value.
func (e *ExceptionValue) Type() string {
	return e.ClassInfo.Name
}

// Inspect returns a string representation of the exception.
func (e *ExceptionValue) Inspect() string {
	return fmt.Sprintf("%s: %s", e.ClassInfo.Name, e.Message)
}

// ============================================================================
// Built-in Exception Classes Registration (Tasks 8.203-8.204)
// ============================================================================

// registerBuiltinExceptions registers the Exception base class and standard exception types.
func (i *Interpreter) registerBuiltinExceptions() {
	// Register TObject as the root base class for all classes
	// This is required for DWScript compatibility - all classes ultimately inherit from TObject
	objectClass := NewClassInfo("TObject")
	objectClass.Parent = nil // Root of the class hierarchy
	objectClass.IsAbstract = false
	objectClass.IsExternal = false

	// Add basic TObject constructor
	objectClass.Constructors["Create"] = nil // Placeholder, handled specially

	i.classes["TObject"] = objectClass

	// Task 8.203: Register Exception base class
	exceptionClass := NewClassInfo("Exception")
	exceptionClass.Parent = objectClass // Exception inherits from TObject
	exceptionClass.Fields["Message"] = types.STRING
	exceptionClass.IsAbstract = false
	exceptionClass.IsExternal = false

	// Add Create constructor - just a placeholder, will be handled specially
	exceptionClass.Constructors["Create"] = nil

	i.classes["Exception"] = exceptionClass

	// Task 8.204: Register standard exception types
	standardExceptions := []string{
		"EConvertError",
		"ERangeError",
		"EDivByZero",
		"EAssertionFailed",
		"EInvalidOp",
	}

	for _, excName := range standardExceptions {
		excClass := NewClassInfo(excName)
		excClass.Parent = exceptionClass
		excClass.Fields["Message"] = types.STRING
		excClass.IsAbstract = false
		excClass.IsExternal = false

		// Inherit Create constructor
		excClass.Constructors["Create"] = nil

		i.classes[excName] = excClass
	}

	// Task 9.51a: Register EHost exception wrapper for host runtime errors.
	eHostClass := NewClassInfo("EHost")
	eHostClass.Parent = exceptionClass
	eHostClass.Fields["Message"] = types.STRING
	eHostClass.Fields["ExceptionClass"] = types.STRING
	eHostClass.IsAbstract = false
	eHostClass.IsExternal = false
	eHostClass.Constructors["Create"] = nil

	i.classes["EHost"] = eHostClass
}

// ============================================================================
// Exception Handling Evaluation (Tasks 8.213-8.216)
// ============================================================================

// evalTryStatement evaluates a try/except/finally statement.
// Task 8.213: Implement evalTryStatement
func (i *Interpreter) evalTryStatement(stmt *ast.TryStatement) Value {
	// Set up finally block to run at the end
	if stmt.FinallyClause != nil {
		defer func() {
			// Save the current exception state
			savedExc := i.exception

			// Set ExceptObject to the current exception in finally block
			oldExceptObject, _ := i.env.Get("ExceptObject")
			if savedExc != nil {
				i.env.Set("ExceptObject", savedExc.Instance)
			}

			// Clear exception so finally block can execute
			i.exception = nil
			// Execute finally block
			i.evalBlockStatement(stmt.FinallyClause.Block)

			// If finally raised a new exception, keep it (replaces original)
			// If finally completed normally, restore the original exception
			if i.exception == nil {
				// Finally completed normally, restore original exception
				i.exception = savedExc
			}
			// else: finally raised an exception, keep it (it replaces the original)

			// Restore ExceptObject
			i.env.Set("ExceptObject", oldExceptObject)
		}()
	}

	// Execute try block
	i.evalBlockStatement(stmt.TryBlock)

	// If an exception occurred, try to handle it
	if i.exception != nil {
		if stmt.ExceptClause != nil {
			i.evalExceptClause(stmt.ExceptClause)
		}
		// If exception is still active after except clause, it will propagate
	}

	return nil
}

// evalExceptClause evaluates an except clause.
// Task 8.214: Implement evalExceptClause
func (i *Interpreter) evalExceptClause(clause *ast.ExceptClause) {
	if i.exception == nil {
		// No exception to handle
		return
	}

	// Save the current exception
	exc := i.exception

	// If no handlers, this is a bare except - catches all
	if len(clause.Handlers) == 0 {
		i.exception = nil // Clear the exception
		return
	}

	// Try each handler in order
	for _, handler := range clause.Handlers {
		if i.matchesExceptionType(exc, handler.ExceptionType) {
			// Create new scope for exception variable
			oldEnv := i.env
			i.env = NewEnclosedEnvironment(i.env)

			// Bind exception variable
			if handler.Variable != nil {
				// Use Define instead of Set to create a new variable in the current scope
				i.env.Define(handler.Variable.Value, exc.Instance)
			}

			// Save the current handlerException (for nested handlers)
			savedHandlerException := i.handlerException

			// Save exception for bare raise to access
			i.handlerException = exc

			// Set ExceptObject to the current exception
			// Save old ExceptObject value to restore later
			oldExceptObject, _ := i.env.Get("ExceptObject")
			i.env.Set("ExceptObject", exc.Instance)

			// Temporarily clear exception to allow handler to execute
			i.exception = nil

			// Execute handler statement
			// Use Eval directly, not evalStatement
			i.Eval(handler.Statement)

			// After handler executes:
			// - If i.exception is still nil, handler completed normally
			// - If i.exception is not nil, handler raised/re-raised

			// Restore handler exception context (for nested handlers)
			i.handlerException = savedHandlerException

			// Restore ExceptObject
			i.env.Set("ExceptObject", oldExceptObject)

			// Restore environment
			i.env = oldEnv

			// If handler raised an exception (including bare raise), it's now in i.exception
			// If handler completed normally, i.exception is nil
			// Either way, we're done with this handler
			return
		}
	}

	// No handler matched - execute else block if present
	if clause.ElseBlock != nil {
		// Clear the exception before executing else block
		i.exception = nil
		i.evalBlockStatement(clause.ElseBlock)
	}
	// If no else block, exception remains active and will propagate
}

// matchesExceptionType checks if an exception matches a handler's type.
// Task 8.215: Implement exception type matching
func (i *Interpreter) matchesExceptionType(exc *ExceptionValue, typeAnnotation *ast.TypeAnnotation) bool {
	if typeAnnotation == nil {
		return true // Bare handler catches all
	}

	typeName := typeAnnotation.Name

	// Check if exception class matches or inherits from handler type
	currentClass := exc.ClassInfo
	for currentClass != nil {
		if currentClass.Name == typeName {
			return true
		}
		// Check parent class
		currentClass = currentClass.Parent
	}

	return false
}

// evalRaiseStatement evaluates a raise statement.
// Task 8.216: Implement evalRaiseStatement
func (i *Interpreter) evalRaiseStatement(stmt *ast.RaiseStatement) Value {
	// Bare raise - re-raise current exception
	if stmt.Exception == nil {
		// Use the exception saved by evalExceptClause
		if i.handlerException != nil {
			// Re-raise the exception
			i.exception = i.handlerException
			return nil
		}

		panic("runtime error: bare raise with no active exception")
	}

	// Evaluate exception expression
	excVal := i.Eval(stmt.Exception)

	// Should be an object instance
	obj, ok := excVal.(*ObjectInstance)
	if !ok {
		panic(fmt.Sprintf("runtime error: raise requires exception object, got %s", excVal.Type()))
	}

	// Get the class info
	classInfo := obj.Class

	// Create exception value
	// Extract message from the object's Message field
	message := ""
	if msgVal, ok := obj.Fields["Message"]; ok {
		if strVal, ok := msgVal.(*StringValue); ok {
			message = strVal.Value
		}
	}

	// Capture current call stack (make a copy to avoid slice aliasing)
	callStack := make([]string, len(i.callStack))
	copy(callStack, i.callStack)

	i.exception = &ExceptionValue{
		ClassInfo: classInfo,
		Message:   message,
		Instance:  obj,
		CallStack: callStack,
	}

	return nil
}

// evalStatement is a helper that evaluates statements and checks for exceptions.
// It returns early if an exception is active (for stack unwinding).
func (i *Interpreter) evalStatement(stmt ast.Statement) Value {
	// Check if exception is active - if so, unwind the stack
	if i.exception != nil {
		return nil
	}

	// Evaluate the statement
	return i.Eval(stmt)
}

// isExceptionClass checks if a class is an Exception or inherits from Exception.
// Task 8.218: Helper for special exception constructor handling.
func (i *Interpreter) isExceptionClass(classInfo *ClassInfo) bool {
	current := classInfo
	for current != nil {
		if current.Name == "Exception" {
			return true
		}
		current = current.Parent
	}
	return false
}
