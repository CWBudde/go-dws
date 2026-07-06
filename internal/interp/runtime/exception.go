// Package runtime provides runtime value types for the DWScript interpreter.
// This file contains ExceptionValue, the runtime representation of exceptions.
package runtime

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// ExceptionValue represents an exception object at runtime.
// It holds the exception class type, the message, position, and the call stack at the point of raise.
type ExceptionValue struct {
	ClassInfo any             // Deprecated: Use Metadata instead.
	Metadata  *ClassMetadata  // AST-free class metadata
	Instance  *ObjectInstance // Exception object instance
	Position  *lexer.Position // Position where the exception was raised (for error reporting)
	Message   string          // Exception message

	CallStack errors.StackTrace // Stack trace at the point the exception was raised

	// UserRaised marks exceptions raised by an explicit `raise` statement in
	// script code. DWScript reports these unhandled as
	// "User defined exception: <message>"; runtime errors keep their message.
	UserRaised bool
}

// Type returns the type of this exception value.
func (e *ExceptionValue) Type() string {
	// Prefer metadata if available
	if e.Metadata != nil {
		return e.Metadata.Name
	}
	// Fallback to ClassInfo for backward compatibility
	if e.ClassInfo != nil {
		return "EXCEPTION"
	}
	return "EXCEPTION"
}

// StackTraceString renders the call stack for DWScript's Exception.StackTrace
// property: one line per active routine (innermost first), each labeled with the
// routine that made the call/raise and positioned at that call site. The raise
// site itself is the innermost frame (labeled by the raising routine). The
// outermost frame is the main program, whose label is empty.
func (e *ExceptionValue) StackTraceString() string {
	if e == nil {
		return ""
	}
	frames := make(errors.StackTrace, len(e.CallStack), len(e.CallStack)+1)
	copy(frames, e.CallStack)
	if e.Position != nil {
		// Append the raise site as the innermost frame. Its label is supplied by
		// the preceding (raising) frame in DWScriptString, so the name is unused.
		frames = append(frames, errors.NewStackFrame("", "", e.Position))
	}
	return frames.DWScriptString()
}

// GetInstance returns the ObjectInstance from this exception.
// Returns interface{} to avoid circular import issues with evaluator package.
func (e *ExceptionValue) GetInstance() interface{} {
	if e == nil {
		return nil
	}
	return e.Instance
}

// Inspect returns a string representation of the exception.
func (e *ExceptionValue) Inspect() string {
	if e == nil {
		return "EXCEPTION: <nil>"
	}
	// Prefer metadata if available
	if e.Metadata != nil {
		return fmt.Sprintf("%s: %s", e.Metadata.Name, e.Message)
	}
	// Fallback for backward compatibility
	return fmt.Sprintf("EXCEPTION: %s", e.Message)
}

// NewException creates a new exception with class metadata and message.
// This is the primary constructor for exceptions in the runtime.
func NewException(metadata *ClassMetadata, instance *ObjectInstance, message string, pos *lexer.Position, callStack errors.StackTrace) *ExceptionValue {
	return &ExceptionValue{
		Metadata:  metadata,
		Instance:  instance,
		Message:   message,
		Position:  pos,
		CallStack: callStack,
	}
}

// NewExceptionFromObject wraps an existing ObjectInstance as an exception.
// Used for raise statements with object instances.
func NewExceptionFromObject(instance *ObjectInstance, message string, pos *lexer.Position, callStack errors.StackTrace) *ExceptionValue {
	// Extract metadata from instance
	var metadata *ClassMetadata
	if instance != nil && instance.Class != nil {
		metadata = instance.Class.GetMetadata()
	}

	return &ExceptionValue{
		Metadata:  metadata,
		Instance:  instance,
		Message:   message,
		Position:  pos,
		CallStack: callStack,
	}
}
