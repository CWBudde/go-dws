package evaluator

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// This file contains error handling utilities for the evaluator visitor methods.

const readOnlyPropertyWriteMessage = "Cannot set a value for a read-only property"

// newError creates a new error value with optional formatting and location information.
func (e *Evaluator) newError(node ast.Node, format string, args ...any) Value {
	message := fmt.Sprintf(format, args...)

	// Add location information if node is available
	// Format matches RuntimeError.String() format: "[line: N, column: M]"
	if node != nil {
		pos := node.Pos()
		if pos.Line > 0 {
			message = fmt.Sprintf("%s [line: %d, column: %d]", message, pos.Line, pos.Column)
		}
	}

	return &runtime.ErrorValue{Message: message}
}

// isError checks if a value is an error.
func isError(val Value) bool {
	if val != nil {
		return val.Type() == "ERROR"
	}
	return false
}

// spliceRoutineNameIntoError inserts " in <routine>" before the trailing
// " [line: N, column: M]" location suffix of a runtime error message, matching
// DWScript's format for runtime errors raised inside a routine, e.g.
// "Object not instantiated in TMyObj.Proc [line: 11, column: 12]".
func spliceRoutineNameIntoError(message, routine string) string {
	if routine == "" {
		return message
	}
	if idx := strings.LastIndex(message, " [line: "); idx >= 0 {
		return message[:idx] + " in " + routine + message[idx:]
	}
	return message + " in " + routine
}

// raiseErrorValueAsException converts a runtime ErrorValue into a catchable
// script exception (DWScript runtime errors are catchable with try/except).
// If routine is non-empty it is spliced into the message before the location
// suffix ("<msg> in <routine> [line: ...]"), matching DWScript semantics where
// the message is formed at the raise point inside the routine.
func (e *Evaluator) raiseErrorValueAsException(errVal Value, routine string, ctx *ExecutionContext) {
	message := ""
	if ev, ok := errVal.(*runtime.ErrorValue); ok {
		message = ev.Message
	} else if errVal != nil {
		message = errVal.String()
	}
	message = spliceRoutineNameIntoError(message, routine)
	ctx.SetException(e.createException("Exception", message, nil, ctx))
}

// currentRoutineName returns the qualified name of the routine currently on
// top of the call stack, or "" when executing at program level.
func currentRoutineName(ctx *ExecutionContext) string {
	frame := ctx.GetCallStack().Current()
	if frame == nil {
		return ""
	}
	return frame.FunctionName
}
