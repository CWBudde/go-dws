package errors

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// StackFrame represents a single frame in a call stack.
// It captures the function being executed and its location in the source code.
type StackFrame struct {
	Position     *lexer.Position
	FunctionName string
	FileName     string
}

// String returns a formatted string representation of the stack frame.
// Format matches DWScript: "FunctionName [line: N, column: M]"
// If position is not available, returns just the function name.
func (sf StackFrame) String() string {
	if sf.Position == nil {
		return sf.FunctionName
	}
	return fmt.Sprintf("%s [line: %d, column: %d]",
		sf.FunctionName, sf.Position.Line, sf.Position.Column)
}

// StackTrace represents a complete call stack as a sequence of frames.
// Frames are ordered from oldest (bottom of stack) to newest (top of stack).
type StackTrace []StackFrame

// String returns a formatted string representation of the entire stack trace.
// Each frame is printed on a separate line.
func (st StackTrace) String() string {
	if len(st) == 0 {
		return ""
	}

	var sb strings.Builder
	for i := len(st) - 1; i >= 0; i-- {
		sb.WriteString(st[i].String())
		if i > 0 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// Reverse returns a new StackTrace with frames in reverse order.
// This is useful when you need to display the stack with the most recent call first.
func (st StackTrace) Reverse() StackTrace {
	reversed := make(StackTrace, len(st))
	for i, frame := range st {
		reversed[len(st)-1-i] = frame
	}
	return reversed
}

// Top returns the most recent (top) frame in the stack, or nil if empty.
func (st StackTrace) Top() *StackFrame {
	if len(st) == 0 {
		return nil
	}
	return &st[len(st)-1]
}

// Bottom returns the oldest (bottom) frame in the stack, or nil if empty.
func (st StackTrace) Bottom() *StackFrame {
	if len(st) == 0 {
		return nil
	}
	return &st[0]
}

// Depth returns the number of frames in the stack.
func (st StackTrace) Depth() int {
	return len(st)
}

// NewStackFrame creates a new stack frame with the given function name and position.
func NewStackFrame(functionName string, fileName string, position *lexer.Position) StackFrame {
	return StackFrame{
		FunctionName: functionName,
		FileName:     fileName,
		Position:     position,
	}
}

// NewStackTrace creates a new empty stack trace.
func NewStackTrace() StackTrace {
	return make(StackTrace, 0)
}
