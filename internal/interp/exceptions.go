package interp

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Exception Value Representation
// ============================================================================

// ExceptionValue represents an exception object at runtime.
// It holds the exception class type, the message, position, and the call stack at the point of raise.
// Task 3.5.43: Migrated to use ClassMetadata instead of direct ClassInfo dependency.
type ExceptionValue struct {
	Metadata  *runtime.ClassMetadata // AST-free class metadata (Task 3.5.43)
	Instance  *ObjectInstance
	Message   string
	Position  *lexer.Position   // Position where the exception was raised (for error reporting)
	CallStack errors.StackTrace // Stack trace at the point the exception was raised

	// Deprecated: Use Metadata instead. Will be removed in Phase 3.5.44.
	// Kept temporarily for backward compatibility during migration.
	ClassInfo *ClassInfo
}

// Type returns the type of this exception value.
// Task 3.5.43: Updated to prefer Metadata over ClassInfo.
func (e *ExceptionValue) Type() string {
	// Prefer metadata if available
	if e.Metadata != nil {
		return e.Metadata.Name
	}
	// Fallback to ClassInfo for backward compatibility
	if e.ClassInfo != nil {
		return e.ClassInfo.Name
	}
	return "EXCEPTION"
}

// Inspect returns a string representation of the exception.
// Task 3.5.43: Updated to prefer Metadata over ClassInfo.
func (e *ExceptionValue) Inspect() string {
	// Phase 3.5.44: Add nil check to prevent panic
	if e == nil {
		return "EXCEPTION: <nil>"
	}
	// Prefer metadata if available
	if e.Metadata != nil {
		return fmt.Sprintf("%s: %s", e.Metadata.Name, e.Message)
	}
	// Fallback to ClassInfo for backward compatibility
	if e.ClassInfo != nil {
		return fmt.Sprintf("%s: %s", e.ClassInfo.Name, e.Message)
	}
	return fmt.Sprintf("EXCEPTION: %s", e.Message)
}

// ============================================================================
// Built-in Exception Classes Registration
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
	// Create a minimal Create constructor AST node
	// The nil body means it just initializes fields with defaults
	createConstructor := &ast.FunctionDecl{
		Name: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Create"},
				},
			},
			Value: "Create",
		},
		Parameters:    []*ast.Parameter{},                                 // No parameters
		ReturnType:    nil,                                                // Constructors don't have explicit return types
		Body:          &ast.BlockStatement{Statements: []ast.Statement{}}, // Empty body
		IsConstructor: true,
	}
	// Use lowercase key for case-insensitive constructor matching
	objectClass.Constructors["create"] = createConstructor
	objectClass.ConstructorOverloads["create"] = []*ast.FunctionDecl{createConstructor}

	// Use lowercase key for O(1) case-insensitive lookup
	i.classes[strings.ToLower("TObject")] = objectClass
	// Also register in TypeSystem for shared access
	i.typeSystem.RegisterClass("TObject", objectClass)

	// Register Exception base class
	exceptionClass := NewClassInfo("Exception")
	exceptionClass.Parent = objectClass // Exception inherits from TObject
	exceptionClass.Fields["Message"] = types.STRING
	exceptionClass.IsAbstract = false
	exceptionClass.IsExternal = false

	// Task 3.5.43: Set parent metadata for hierarchy checks
	exceptionClass.Metadata.Parent = objectClass.Metadata
	exceptionClass.Metadata.ParentName = "TObject"

	// Task 3.5.40: Populate metadata for exception fields
	messageMeta := &runtime.FieldMetadata{
		Name:       "Message",
		TypeName:   "String",
		Type:       types.STRING,
		Visibility: runtime.FieldVisibilityPublic,
	}
	runtime.AddFieldToClass(exceptionClass.Metadata, messageMeta)

	// Add Create constructor - just a placeholder, will be handled specially
	exceptionClass.Constructors["Create"] = nil

	// Use lowercase key for O(1) case-insensitive lookup
	i.classes[strings.ToLower("Exception")] = exceptionClass
	// Also register in TypeSystem for shared access
	i.typeSystem.RegisterClassWithParent("Exception", exceptionClass, "TObject")

	// Register standard exception types
	standardExceptions := []string{
		"EConvertError",
		"ERangeError",
		"EDivByZero",
		"EAssertionFailed",
		"EInvalidOp",
		"EScriptStackOverflow",
	}

	for _, excName := range standardExceptions {
		excClass := NewClassInfo(excName)
		excClass.Parent = exceptionClass
		excClass.Fields["Message"] = types.STRING
		excClass.IsAbstract = false
		excClass.IsExternal = false

		// Task 3.5.43: Set parent metadata for hierarchy checks
		excClass.Metadata.Parent = exceptionClass.Metadata
		excClass.Metadata.ParentName = "Exception"

		// Task 3.5.40: Populate metadata for exception fields
		messageMeta := &runtime.FieldMetadata{
			Name:       "Message",
			TypeName:   "String",
			Type:       types.STRING,
			Visibility: runtime.FieldVisibilityPublic,
		}
		runtime.AddFieldToClass(excClass.Metadata, messageMeta)

		// Inherit Create constructor
		excClass.Constructors["Create"] = nil

		// Use lowercase key for O(1) case-insensitive lookup
		i.classes[strings.ToLower(excName)] = excClass
		// Also register in TypeSystem for shared access
		i.typeSystem.RegisterClassWithParent(excName, excClass, "Exception")
	}

	// Register EHost exception wrapper for host runtime errors.
	eHostClass := NewClassInfo("EHost")
	eHostClass.Parent = exceptionClass
	eHostClass.Fields["Message"] = types.STRING
	eHostClass.Fields["ExceptionClass"] = types.STRING
	eHostClass.IsAbstract = false
	eHostClass.IsExternal = false

	// Task 3.5.43: Set parent metadata for hierarchy checks
	eHostClass.Metadata.Parent = exceptionClass.Metadata
	eHostClass.Metadata.ParentName = "Exception"

	// Task 3.5.40: Populate metadata for EHost fields
	messageMeta2 := &runtime.FieldMetadata{
		Name:       "Message",
		TypeName:   "String",
		Type:       types.STRING,
		Visibility: runtime.FieldVisibilityPublic,
	}
	runtime.AddFieldToClass(eHostClass.Metadata, messageMeta2)

	exceptionClassMeta := &runtime.FieldMetadata{
		Name:       "ExceptionClass",
		TypeName:   "String",
		Type:       types.STRING,
		Visibility: runtime.FieldVisibilityPublic,
	}
	runtime.AddFieldToClass(eHostClass.Metadata, exceptionClassMeta)

	eHostClass.Constructors["Create"] = nil

	// Use lowercase key for O(1) case-insensitive lookup
	i.classes[strings.ToLower("EHost")] = eHostClass
	// Also register in TypeSystem for shared access
	i.typeSystem.RegisterClassWithParent("EHost", eHostClass, "Exception")
}

// raiseMaxRecursionExceeded raises an EScriptStackOverflow exception when the
// maximum recursion depth is exceeded. This prevents infinite recursion and
// stack overflow errors.
func (i *Interpreter) raiseMaxRecursionExceeded() Value {
	message := fmt.Sprintf("Maximal recursion exceeded (%d)", i.maxRecursionDepth)

	// Capture current call stack
	callStack := make(errors.StackTrace, len(i.callStack))
	copy(callStack, i.callStack)

	// Look up EScriptStackOverflow class
	// PR #147: Use lowercase key for O(1) case-insensitive lookup
	stackOverflowClass, ok := i.classes[strings.ToLower("EScriptStackOverflow")]
	if !ok {
		// Fall back to Exception if EScriptStackOverflow isn't registered
		if baseClass, exists := i.classes[strings.ToLower("Exception")]; exists {
			stackOverflowClass = baseClass
		} else {
			// As a last resort, return NilValue without setting exception
			return &NilValue{}
		}
	}

	// Create exception instance
	instance := NewObjectInstance(stackOverflowClass)
	instance.SetField("Message", &StringValue{Value: message})

	// Set the exception
	// Position is nil for internally-raised exceptions like recursion overflow
	// Task 3.5.43: Populate Metadata field from ClassInfo
	i.exception = &ExceptionValue{
		Metadata:  stackOverflowClass.Metadata,
		Instance:  instance,
		Message:   message,
		Position:  nil,
		CallStack: callStack,
		ClassInfo: stackOverflowClass, // Deprecated: backward compatibility
	}

	return &NilValue{}
}

// ============================================================================
// Exception Handling Evaluation
// ============================================================================

// evalTryStatement evaluates a try/except/finally statement.
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
// Task 3.5.43: Updated to prefer Metadata over ClassInfo for hierarchy checks.
func (i *Interpreter) matchesExceptionType(exc *ExceptionValue, typeExpr ast.TypeExpression) bool {
	if typeExpr == nil {
		return true // Bare handler catches all
	}

	typeName := typeExpr.String()

	// Prefer metadata if available
	if exc.Metadata != nil {
		// Check if exception class matches or inherits from handler type
		currentMetadata := exc.Metadata
		for currentMetadata != nil {
			if currentMetadata.Name == typeName {
				return true
			}
			// Check parent class metadata
			currentMetadata = currentMetadata.Parent
		}
		return false
	}

	// Fallback to ClassInfo for backward compatibility
	if exc.ClassInfo != nil {
		// Check if exception class matches or inherits from handler type
		currentClass := exc.ClassInfo
		for currentClass != nil {
			if currentClass.Name == typeName {
				return true
			}
			// Check parent class
			currentClass = currentClass.Parent
		}
	}

	return false
}

// evalRaiseStatement evaluates a raise statement.
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
	// Task 3.5.40: Use GetField for proper normalization
	message := ""
	if msgVal := obj.GetField("Message"); msgVal != nil {
		if strVal, ok := msgVal.(*StringValue); ok {
			message = strVal.Value
		}
	}

	// Capture current call stack (make a copy to avoid slice aliasing)
	callStack := make(errors.StackTrace, len(i.callStack))
	copy(callStack, i.callStack)

	// Capture position of the raise statement
	pos := stmt.Token.Pos

	// Task 3.5.43: Populate Metadata field from ClassInfo
	i.exception = &ExceptionValue{
		Metadata:  classInfo.Metadata,
		Message:   message,
		Instance:  obj,
		Position:  &pos,
		CallStack: callStack,
		ClassInfo: classInfo, // Deprecated: backward compatibility
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
