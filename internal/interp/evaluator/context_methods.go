package evaluator

import (
	"fmt"
	"io"
	"math/rand"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Context Interface Implementation (evaluator methods)
// ============================================================================

// NewError creates an error value with location information from the current node.
func (e *Evaluator) NewError(format string, args ...interface{}) Value {
	return e.newError(e.currentNode, format, args...)
}

// Note: CurrentNode() is already implemented in evaluator.go.

// RandSource returns the random number generator for built-in functions.
func (e *Evaluator) RandSource() *rand.Rand {
	return e.rand
}

// GetRandSeed returns the current random number generator seed value.
func (e *Evaluator) GetRandSeed() int64 {
	return e.randSeed
}

// SetRandSeed sets the random number generator seed.
func (e *Evaluator) SetRandSeed(seed int64) {
	e.randSeed = seed
	e.rand.Seed(seed)
}

// Write outputs a string to the configured output writer without a newline.
func (e *Evaluator) Write(s string) {
	if e.output != nil {
		io.WriteString(e.output, s)
	}
}

// WriteLine outputs a string to the configured output writer with a newline.
func (e *Evaluator) WriteLine(s string) {
	if e.output != nil {
		fmt.Fprintln(e.output, s)
	}
}

// IsAssigned checks if a Variant value has been assigned (is not uninitialized).
func (e *Evaluator) IsAssigned(value Value) bool {
	if value == nil {
		return false
	}

	if intfVal, ok := value.(*runtime.InterfaceInstance); ok {
		return intfVal.Object != nil
	}

	if wrapper, ok := value.(runtime.VariantWrapper); ok {
		unwrapped := wrapper.UnwrapVariant()
		return unwrapped != nil
	}

	return true
}

// GetCallStackString returns a formatted string representation of the current call stack.
func (e *Evaluator) GetCallStackString() string {
	if e.currentContext == nil {
		return ""
	}
	return e.currentContext.GetCallStack().String()
}

// GetCallStackArray returns the current call stack as an array of records.
func (e *Evaluator) GetCallStackArray() Value {
	if e.currentContext == nil {
		return &runtime.ArrayValue{
			Elements:  []runtime.Value{},
			ArrayType: types.NewDynamicArrayType(types.VARIANT),
		}
	}

	frames := e.currentContext.GetCallStack().Frames()
	elements := make([]runtime.Value, len(frames))

	for idx, frame := range frames {
		fields := make(map[string]runtime.Value)
		fields["FunctionName"] = &runtime.StringValue{Value: frame.FunctionName}

		if frame.Position != nil {
			fields["Line"] = &runtime.IntegerValue{Value: int64(frame.Position.Line)}
			fields["Column"] = &runtime.IntegerValue{Value: int64(frame.Position.Column)}
		} else {
			fields["Line"] = &runtime.IntegerValue{Value: 0}
			fields["Column"] = &runtime.IntegerValue{Value: 0}
		}

		elements[idx] = &runtime.RecordValue{
			Fields:     fields,
			RecordType: nil,
		}
	}

	return &runtime.ArrayValue{
		Elements:  elements,
		ArrayType: types.NewDynamicArrayType(types.VARIANT),
	}
}

// RaiseAssertionFailed raises an EAssertionFailed exception with an optional custom message.
func (e *Evaluator) RaiseAssertionFailed(customMessage string) {
	e.exceptionMgr.RaiseAssertionFailed(customMessage)
}

// EvalFunctionPointer executes a function pointer with given arguments.
func (e *Evaluator) EvalFunctionPointer(funcPtr Value, args []Value) Value {
	return e.oopEngine.CallFunctionPointer(funcPtr, args, e.currentNode)
}
