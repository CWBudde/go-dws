package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// CallStack manages the function call stack for execution tracking.
// It provides stack overflow detection and comprehensive stack trace support.
//
// Phase 3.3.3: Extract call stack management from Interpreter.
type CallStack struct {
	frames   errors.StackTrace
	maxDepth int
}

// NewCallStack creates a new call stack with the given maximum depth.
// If maxDepth is 0 or negative, DefaultMaxRecursionDepth (1024) is used.
func NewCallStack(maxDepth int) *CallStack {
	if maxDepth <= 0 {
		maxDepth = 1024 // DefaultMaxRecursionDepth
	}
	return &CallStack{
		frames:   errors.NewStackTrace(),
		maxDepth: maxDepth,
	}
}

// Push adds a new frame to the call stack.
// Returns an error if pushing would exceed the maximum depth (stack overflow).
func (cs *CallStack) Push(functionName string, sourceFile string, pos *lexer.Position) error {
	if len(cs.frames) >= cs.maxDepth {
		return fmt.Errorf("stack overflow: maximum recursion depth (%d) exceeded in function '%s'", cs.maxDepth, functionName)
	}

	frame := errors.NewStackFrame(functionName, sourceFile, pos)
	cs.frames = append(cs.frames, frame)
	return nil
}

// Pop removes the most recent frame from the call stack.
// If the stack is empty, this is a no-op.
func (cs *CallStack) Pop() {
	if len(cs.frames) > 0 {
		cs.frames = cs.frames[:len(cs.frames)-1]
	}
}

// Current returns the current (most recent) stack frame, or nil if the stack is empty.
func (cs *CallStack) Current() *errors.StackFrame {
	if len(cs.frames) == 0 {
		return nil
	}
	return &cs.frames[len(cs.frames)-1]
}

// Depth returns the current depth of the call stack.
func (cs *CallStack) Depth() int {
	return len(cs.frames)
}

// Frames returns a copy of all stack frames.
// The frames are returned in the order they were pushed (oldest to newest).
func (cs *CallStack) Frames() errors.StackTrace {
	// Return a copy to prevent external modification
	frames := make(errors.StackTrace, len(cs.frames))
	copy(frames, cs.frames)
	return frames
}

// MaxDepth returns the maximum allowed depth of the call stack.
func (cs *CallStack) MaxDepth() int {
	return cs.maxDepth
}

// SetMaxDepth updates the maximum allowed depth of the call stack.
// If maxDepth is 0 or negative, it's set to DefaultMaxRecursionDepth (1024).
func (cs *CallStack) SetMaxDepth(maxDepth int) {
	if maxDepth <= 0 {
		maxDepth = 1024 // DefaultMaxRecursionDepth
	}
	cs.maxDepth = maxDepth
}

// IsEmpty returns true if the call stack has no frames.
func (cs *CallStack) IsEmpty() bool {
	return len(cs.frames) == 0
}

// WillOverflow returns true if pushing one more frame would exceed the maximum depth.
func (cs *CallStack) WillOverflow() bool {
	return len(cs.frames) >= cs.maxDepth
}

// Clear removes all frames from the call stack.
func (cs *CallStack) Clear() {
	cs.frames = errors.NewStackTrace()
}

// String returns a string representation of the call stack.
// This is useful for debugging and error messages.
func (cs *CallStack) String() string {
	return cs.frames.String()
}

// Clone creates a shallow copy of the call stack.
// The frames slice is copied, but the frames themselves are shared.
func (cs *CallStack) Clone() *CallStack {
	frames := make(errors.StackTrace, len(cs.frames))
	copy(frames, cs.frames)
	return &CallStack{
		frames:   frames,
		maxDepth: cs.maxDepth,
	}
}

// FormatError formats an error message with the current call stack.
// This is useful for creating rich error messages with stack traces.
func (cs *CallStack) FormatError(message string) string {
	if len(cs.frames) == 0 {
		return message
	}
	return fmt.Sprintf("%s\n\nCall stack:\n%s", message, cs.String())
}

// GetFrameAt returns the frame at the given index (0-based, from oldest to newest).
// Returns nil if the index is out of bounds.
func (cs *CallStack) GetFrameAt(index int) *errors.StackFrame {
	if index < 0 || index >= len(cs.frames) {
		return nil
	}
	return &cs.frames[index]
}

// FindFrame searches for the first frame with the given function name.
// Returns the frame and its index, or nil and -1 if not found.
func (cs *CallStack) FindFrame(functionName string) (*errors.StackFrame, int) {
	for i, frame := range cs.frames {
		if frame.FunctionName == functionName {
			return &frame, i
		}
	}
	return nil, -1
}

// ContainsFunction returns true if any frame in the stack has the given function name.
func (cs *CallStack) ContainsFunction(functionName string) bool {
	_, index := cs.FindFrame(functionName)
	return index != -1
}
