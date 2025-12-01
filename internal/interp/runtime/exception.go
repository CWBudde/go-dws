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
	Metadata  *ClassMetadata    // AST-free class metadata
	Instance  *ObjectInstance   // Exception object instance
	Message   string            // Exception message
	Position  *lexer.Position   // Position where the exception was raised (for error reporting)
	CallStack errors.StackTrace // Stack trace at the point the exception was raised

	// Deprecated: Use Metadata instead. Will be removed in Phase 3.5.44.
	// Kept temporarily for backward compatibility during migration.
	// Using any to avoid import cycle during migration period.
	ClassInfo any
}

// Type returns the type of this exception value.
func (e *ExceptionValue) Type() string {
	// Prefer metadata if available
	if e.Metadata != nil {
		return e.Metadata.Name
	}
	// Fallback to ClassInfo for backward compatibility
	if e.ClassInfo != nil {
		// During migration, ClassInfo may be set - extract name if possible
		// This is a temporary workaround until Phase 3.5.44
		return "EXCEPTION"
	}
	return "EXCEPTION"
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
	// Phase 3.5.44: Add nil check to prevent panic
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
