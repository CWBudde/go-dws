package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// mockEnvironment is a simple mock implementation of the Environment interface.
type mockEnvironment struct {
	bindings map[string]interface{}
	outer    *mockEnvironment
}

func newMockEnvironment() *mockEnvironment {
	return &mockEnvironment{
		bindings: make(map[string]interface{}),
	}
}

func (e *mockEnvironment) Define(name string, value interface{}) {
	e.bindings[name] = value
}

func (e *mockEnvironment) Get(name string) (interface{}, bool) {
	val, ok := e.bindings[name]
	if !ok && e.outer != nil {
		return e.outer.Get(name)
	}
	return val, ok
}

func (e *mockEnvironment) Set(name string, value interface{}) bool {
	if _, ok := e.bindings[name]; ok {
		e.bindings[name] = value
		return true
	}
	if e.outer != nil {
		return e.outer.Set(name, value)
	}
	return false
}

func (e *mockEnvironment) NewEnclosedEnvironment() Environment {
	child := newMockEnvironment()
	child.outer = e
	return child
}

// TestControlFlow_Kind tests the ControlFlowKind type.
func TestControlFlow_Kind(t *testing.T) {
	tests := []struct {
		name     string
		kind     ControlFlowKind
		expected string
	}{
		{"none", FlowNone, "none"},
		{"break", FlowBreak, "break"},
		{"continue", FlowContinue, "continue"},
		{"exit", FlowExit, "exit"},
		{"return", FlowReturn, "return"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.kind.String(); got != tt.expected {
				t.Errorf("ControlFlowKind.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestControlFlow_NewControlFlow tests creating a new ControlFlow.
func TestControlFlow_NewControlFlow(t *testing.T) {
	cf := NewControlFlow()
	if cf.Kind() != FlowNone {
		t.Errorf("NewControlFlow() kind = %v, want %v", cf.Kind(), FlowNone)
	}
	if cf.IsActive() {
		t.Errorf("NewControlFlow() IsActive() = true, want false")
	}
}

// TestControlFlow_SetBreak tests setting a break signal.
func TestControlFlow_SetBreak(t *testing.T) {
	cf := NewControlFlow()
	cf.SetBreak()

	if cf.Kind() != FlowBreak {
		t.Errorf("SetBreak() kind = %v, want %v", cf.Kind(), FlowBreak)
	}
	if !cf.IsActive() {
		t.Errorf("SetBreak() IsActive() = false, want true")
	}
	if !cf.IsBreak() {
		t.Errorf("SetBreak() IsBreak() = false, want true")
	}
	if cf.IsContinue() || cf.IsExit() || cf.IsReturn() {
		t.Errorf("SetBreak() set wrong signal type")
	}
}

// TestControlFlow_SetContinue tests setting a continue signal.
func TestControlFlow_SetContinue(t *testing.T) {
	cf := NewControlFlow()
	cf.SetContinue()

	if cf.Kind() != FlowContinue {
		t.Errorf("SetContinue() kind = %v, want %v", cf.Kind(), FlowContinue)
	}
	if !cf.IsActive() {
		t.Errorf("SetContinue() IsActive() = false, want true")
	}
	if !cf.IsContinue() {
		t.Errorf("SetContinue() IsContinue() = false, want true")
	}
	if cf.IsBreak() || cf.IsExit() || cf.IsReturn() {
		t.Errorf("SetContinue() set wrong signal type")
	}
}

// TestControlFlow_SetExit tests setting an exit signal.
func TestControlFlow_SetExit(t *testing.T) {
	cf := NewControlFlow()
	cf.SetExit()

	if cf.Kind() != FlowExit {
		t.Errorf("SetExit() kind = %v, want %v", cf.Kind(), FlowExit)
	}
	if !cf.IsActive() {
		t.Errorf("SetExit() IsActive() = false, want true")
	}
	if !cf.IsExit() {
		t.Errorf("SetExit() IsExit() = false, want true")
	}
	if cf.IsBreak() || cf.IsContinue() || cf.IsReturn() {
		t.Errorf("SetExit() set wrong signal type")
	}
}

// TestControlFlow_SetReturn tests setting a return signal.
func TestControlFlow_SetReturn(t *testing.T) {
	cf := NewControlFlow()
	cf.SetReturn()

	if cf.Kind() != FlowReturn {
		t.Errorf("SetReturn() kind = %v, want %v", cf.Kind(), FlowReturn)
	}
	if !cf.IsActive() {
		t.Errorf("SetReturn() IsActive() = false, want true")
	}
	if !cf.IsReturn() {
		t.Errorf("SetReturn() IsReturn() = false, want true")
	}
	if cf.IsBreak() || cf.IsContinue() || cf.IsExit() {
		t.Errorf("SetReturn() set wrong signal type")
	}
}

// TestControlFlow_Clear tests clearing a control flow signal.
func TestControlFlow_Clear(t *testing.T) {
	cf := NewControlFlow()

	// Set a signal and clear it
	cf.SetBreak()
	if !cf.IsActive() {
		t.Errorf("SetBreak() should set active signal")
	}

	cf.Clear()
	if cf.Kind() != FlowNone {
		t.Errorf("Clear() kind = %v, want %v", cf.Kind(), FlowNone)
	}
	if cf.IsActive() {
		t.Errorf("Clear() IsActive() = true, want false")
	}
}

// TestExecutionContext_NewExecutionContext tests creating a new context.
func TestExecutionContext_NewExecutionContext(t *testing.T) {
	env := newMockEnvironment()
	ctx := NewExecutionContext(env)

	if ctx.Env() != env {
		t.Errorf("NewExecutionContext() env mismatch")
	}
	if ctx.CallStackDepth() != 0 {
		t.Errorf("NewExecutionContext() call stack depth = %d, want 0", ctx.CallStackDepth())
	}
	if ctx.ControlFlow().IsActive() {
		t.Errorf("NewExecutionContext() has active control flow, want none")
	}
	if ctx.Exception() != nil {
		t.Errorf("NewExecutionContext() exception = %v, want nil", ctx.Exception())
	}
	if ctx.HandlerException() != nil {
		t.Errorf("NewExecutionContext() handler exception = %v, want nil", ctx.HandlerException())
	}
}

// TestExecutionContext_Env tests environment management.
func TestExecutionContext_Env(t *testing.T) {
	env1 := newMockEnvironment()
	env2 := newMockEnvironment()
	ctx := NewExecutionContext(env1)

	if ctx.Env() != env1 {
		t.Errorf("Env() = %v, want %v", ctx.Env(), env1)
	}

	ctx.SetEnv(env2)
	if ctx.Env() != env2 {
		t.Errorf("SetEnv() did not update environment")
	}
}

// TestExecutionContext_CallStack tests call stack management.
func TestExecutionContext_CallStack(t *testing.T) {
	env := newMockEnvironment()
	ctx := NewExecutionContext(env)

	// Initial depth should be 0
	if depth := ctx.CallStackDepth(); depth != 0 {
		t.Errorf("initial CallStackDepth() = %d, want 0", depth)
	}

	// Push frames
	pos := lexer.Position{Line: 1, Column: 1}
	frame1 := errors.NewStackFrame("func1", "", &pos)
	frame2 := errors.NewStackFrame("func2", "", &pos)

	ctx.PushCallStack(frame1)
	if depth := ctx.CallStackDepth(); depth != 1 {
		t.Errorf("CallStackDepth() after push = %d, want 1", depth)
	}

	ctx.PushCallStack(frame2)
	if depth := ctx.CallStackDepth(); depth != 2 {
		t.Errorf("CallStackDepth() after second push = %d, want 2", depth)
	}

	// Pop frames
	ctx.PopCallStack()
	if depth := ctx.CallStackDepth(); depth != 1 {
		t.Errorf("CallStackDepth() after pop = %d, want 1", depth)
	}

	ctx.PopCallStack()
	if depth := ctx.CallStackDepth(); depth != 0 {
		t.Errorf("CallStackDepth() after second pop = %d, want 0", depth)
	}

	// Popping empty stack should not panic
	ctx.PopCallStack()
	if depth := ctx.CallStackDepth(); depth != 0 {
		t.Errorf("CallStackDepth() after popping empty stack = %d, want 0", depth)
	}
}

// TestExecutionContext_Exception tests exception management.
func TestExecutionContext_Exception(t *testing.T) {
	env := newMockEnvironment()
	ctx := NewExecutionContext(env)

	// Initial exception should be nil
	if exc := ctx.Exception(); exc != nil {
		t.Errorf("initial Exception() = %v, want nil", exc)
	}

	// Set exception
	testExc := "test exception"
	ctx.SetException(testExc)
	if exc := ctx.Exception(); exc != testExc {
		t.Errorf("Exception() after SetException() = %v, want %v", exc, testExc)
	}

	// Handler exception
	if handlerExc := ctx.HandlerException(); handlerExc != nil {
		t.Errorf("initial HandlerException() = %v, want nil", handlerExc)
	}

	testHandlerExc := "test handler exception"
	ctx.SetHandlerException(testHandlerExc)
	if handlerExc := ctx.HandlerException(); handlerExc != testHandlerExc {
		t.Errorf("HandlerException() after SetHandlerException() = %v, want %v", handlerExc, testHandlerExc)
	}
}

// TestExecutionContext_OldValuesStack tests old values stack management.
func TestExecutionContext_OldValuesStack(t *testing.T) {
	env := newMockEnvironment()
	ctx := NewExecutionContext(env)

	// Push values
	oldVals1 := map[string]interface{}{"x": 1, "y": 2}
	oldVals2 := map[string]interface{}{"a": 10, "b": 20}

	ctx.PushOldValues(oldVals1)
	ctx.PushOldValues(oldVals2)

	// Pop values in LIFO order
	popped2 := ctx.PopOldValues()
	if len(popped2) != 2 || popped2["a"] != 10 || popped2["b"] != 20 {
		t.Errorf("PopOldValues() first pop = %v, want %v", popped2, oldVals2)
	}

	popped1 := ctx.PopOldValues()
	if len(popped1) != 2 || popped1["x"] != 1 || popped1["y"] != 2 {
		t.Errorf("PopOldValues() second pop = %v, want %v", popped1, oldVals1)
	}

	// Popping empty stack should return nil
	popped3 := ctx.PopOldValues()
	if popped3 != nil {
		t.Errorf("PopOldValues() on empty stack = %v, want nil", popped3)
	}
}

// TestExecutionContext_PropContext tests property context management.
func TestExecutionContext_PropContext(t *testing.T) {
	env := newMockEnvironment()
	ctx := NewExecutionContext(env)

	propCtx := ctx.PropContext()
	if propCtx == nil {
		t.Errorf("PropContext() = nil, want non-nil")
	}
	if propCtx.InPropertyGetter {
		t.Errorf("PropContext() InPropertyGetter = true, want false")
	}
	if propCtx.InPropertySetter {
		t.Errorf("PropContext() InPropertySetter = true, want false")
	}
	if len(propCtx.PropertyChain) != 0 {
		t.Errorf("PropContext() PropertyChain length = %d, want 0", len(propCtx.PropertyChain))
	}
}

// TestExecutionContext_Clone tests context cloning.
func TestExecutionContext_Clone(t *testing.T) {
	env := newMockEnvironment()
	ctx := NewExecutionContext(env)

	// Set up some state
	ctx.ControlFlow().SetBreak()
	ctx.SetException("test exception")
	pos := lexer.Position{Line: 1, Column: 1}
	ctx.PushCallStack(errors.NewStackFrame("func1", "", &pos))

	// Clone the context
	cloned := ctx.Clone()

	// Verify shared references
	if cloned.Env() != ctx.Env() {
		t.Errorf("Clone() env is not shared")
	}
	if cloned.ControlFlow() != ctx.ControlFlow() {
		t.Errorf("Clone() control flow is not shared")
	}
	if cloned.Exception() != ctx.Exception() {
		t.Errorf("Clone() exception is not shared")
	}

	// Modifying one should affect the other (shallow copy)
	ctx.ControlFlow().SetExit()
	if !cloned.ControlFlow().IsExit() {
		t.Errorf("Clone() control flow modification not reflected")
	}
}

// TestExecutionContext_Reset tests context reset.
func TestExecutionContext_Reset(t *testing.T) {
	env := newMockEnvironment()
	ctx := NewExecutionContext(env)

	// Set up some state
	ctx.ControlFlow().SetBreak()
	ctx.SetException("test exception")
	ctx.SetHandlerException("handler exception")
	pos := lexer.Position{Line: 1, Column: 1}
	ctx.PushCallStack(errors.NewStackFrame("func1", "", &pos))
	ctx.PushOldValues(map[string]interface{}{"x": 1})

	// Reset the context
	ctx.Reset()

	// Verify everything is reset
	if ctx.CallStackDepth() != 0 {
		t.Errorf("Reset() call stack depth = %d, want 0", ctx.CallStackDepth())
	}
	if ctx.ControlFlow().IsActive() {
		t.Errorf("Reset() control flow is active, want none")
	}
	if ctx.Exception() != nil {
		t.Errorf("Reset() exception = %v, want nil", ctx.Exception())
	}
	if ctx.HandlerException() != nil {
		t.Errorf("Reset() handler exception = %v, want nil", ctx.HandlerException())
	}
	if ctx.PopOldValues() != nil {
		t.Errorf("Reset() old values stack not empty")
	}

	// Environment should not be reset
	if ctx.Env() != env {
		t.Errorf("Reset() changed environment")
	}
}

// TestPropertyEvalContext_NewPropertyEvalContext tests creating a new property context.
func TestPropertyEvalContext_NewPropertyEvalContext(t *testing.T) {
	propCtx := NewPropertyEvalContext()

	if propCtx == nil {
		t.Errorf("NewPropertyEvalContext() = nil, want non-nil")
	}
	if propCtx.InPropertyGetter {
		t.Errorf("NewPropertyEvalContext() InPropertyGetter = true, want false")
	}
	if propCtx.InPropertySetter {
		t.Errorf("NewPropertyEvalContext() InPropertySetter = true, want false")
	}
	if len(propCtx.PropertyChain) != 0 {
		t.Errorf("NewPropertyEvalContext() PropertyChain length = %d, want 0", len(propCtx.PropertyChain))
	}
}

// TestControlFlow_StateTransitions tests control flow state transitions.
func TestControlFlow_StateTransitions(t *testing.T) {
	cf := NewControlFlow()

	// None -> Break
	cf.SetBreak()
	if !cf.IsBreak() {
		t.Errorf("SetBreak() failed")
	}

	// Break -> Continue
	cf.SetContinue()
	if !cf.IsContinue() || cf.IsBreak() {
		t.Errorf("SetContinue() failed or Break still set")
	}

	// Continue -> Exit
	cf.SetExit()
	if !cf.IsExit() || cf.IsContinue() {
		t.Errorf("SetExit() failed or Continue still set")
	}

	// Exit -> Return
	cf.SetReturn()
	if !cf.IsReturn() || cf.IsExit() {
		t.Errorf("SetReturn() failed or Exit still set")
	}

	// Return -> None
	cf.Clear()
	if cf.IsActive() || cf.Kind() != FlowNone {
		t.Errorf("Clear() failed")
	}
}
