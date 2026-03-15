package interp

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// raiseException raises an exception with the given class name and message.
func (i *Interpreter) raiseException(className, message string, pos *lexer.Position) {
	excClass := i.lookupRegisteredClassInfo(className)
	if excClass == nil {
		excClass = i.lookupRegisteredClassInfo("Exception")
	}
	if excClass == nil {
		// This shouldn't happen, but handle it gracefully
		i.ctx.SetException(&runtime.ExceptionValue{
			ClassInfo: NewClassInfo(className),
			Instance:  nil,
			Message:   message,
			Position:  pos,
			CallStack: i.ctx.CallStack(),
		})
		return
	}

	// Create an instance of the exception class
	instance := NewObjectInstance(excClass)
	instance.SetField("Message", &StringValue{Value: message})

	// Set the exception
	i.ctx.SetException(&runtime.ExceptionValue{
		ClassInfo: excClass,
		Instance:  instance,
		Message:   message,
		Position:  pos,
		CallStack: i.ctx.CallStack(),
	})
}

// getOldValue retrieves a captured old value by identifier name.
// Returns the value and true if found, or nil and false if not found.
func (i *Interpreter) getOldValue(identName string) (Value, bool) {
	val, exists := i.ctx.GetOldValue(identName)
	if !exists {
		return nil, false
	}
	if typedVal, ok := val.(Value); ok {
		return typedVal, true
	}
	return nil, false
}
