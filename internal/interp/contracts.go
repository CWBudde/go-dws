package interp

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// raiseException raises an exception with the given class name and message.
func (i *Interpreter) raiseException(className, message string, pos *lexer.Position) {
	// Get the exception class
	// PR #147: Use normalized key for O(1) case-insensitive lookup
	excClass, ok := i.classes[ident.Normalize(className)]
	if !ok {
		// Fallback to base Exception if class not found
		excClass, ok = i.classes[ident.Normalize("Exception")]
		if !ok {
			// This shouldn't happen, but handle it gracefully
			i.exception = &runtime.ExceptionValue{
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
	i.exception = &runtime.ExceptionValue{
		ClassInfo: excClass,
		Instance:  instance,
		Message:   message,
		Position:  pos,
		CallStack: i.callStack,
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
