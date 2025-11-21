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
		expected string
		kind     ControlFlowKind
	}{
		{name: "none", expected: "none", kind: FlowNone},
		{name: "break", expected: "break", kind: FlowBreak},
		{name: "continue", expected: "continue", kind: FlowContinue},
		{name: "exit", expected: "exit", kind: FlowExit},
		{name: "return", expected: "return", kind: FlowReturn},
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

// TestExecutionContext_NewExecutionContextWithMaxDepth tests creating a context with custom max depth.
func TestExecutionContext_NewExecutionContextWithMaxDepth(t *testing.T) {
	env := newMockEnvironment()
	maxDepth := 512

	ctx := NewExecutionContextWithMaxDepth(env, maxDepth)

	if ctx.Env() != env {
		t.Errorf("NewExecutionContextWithMaxDepth() env mismatch")
	}

	// Verify the max depth is set correctly
	callStack := ctx.GetCallStack()
	if callStack.MaxDepth() != maxDepth {
		t.Errorf("NewExecutionContextWithMaxDepth() max depth = %d, want %d", callStack.MaxDepth(), maxDepth)
	}

	if ctx.CallStackDepth() != 0 {
		t.Errorf("NewExecutionContextWithMaxDepth() call stack depth = %d, want 0", ctx.CallStackDepth())
	}
}

// TestExecutionContext_GetCallStack tests getting the CallStack instance.
func TestExecutionContext_GetCallStack(t *testing.T) {
	env := newMockEnvironment()
	ctx := NewExecutionContext(env)

	callStack := ctx.GetCallStack()
	if callStack == nil {
		t.Fatalf("GetCallStack() returned nil")
	}

	// Verify it's the same instance
	callStack2 := ctx.GetCallStack()
	if callStack != callStack2 {
		t.Errorf("GetCallStack() returns different instances")
	}

	// Verify we can use it directly
	if callStack.Depth() != 0 {
		t.Errorf("GetCallStack().Depth() = %d, want 0", callStack.Depth())
	}

	// Push a frame and verify
	err := callStack.Push("testFunc", "test.dws", &lexer.Position{Line: 1, Column: 1})
	if err != nil {
		t.Errorf("GetCallStack().Push() error = %v", err)
	}

	if callStack.Depth() != 1 {
		t.Errorf("GetCallStack().Depth() after push = %d, want 1", callStack.Depth())
	}
}

// TestExecutionContext_CallStack_Deprecated tests the deprecated CallStack() method.
func TestExecutionContext_CallStack_Deprecated(t *testing.T) {
	env := newMockEnvironment()
	ctx := NewExecutionContext(env)

	// Push some frames
	pos := lexer.Position{Line: 1, Column: 1}
	ctx.PushCallStack(errors.NewStackFrame("func1", "file1.dws", &pos))
	ctx.PushCallStack(errors.NewStackFrame("func2", "file2.dws", &pos))

	// Use deprecated CallStack() method
	frames := ctx.CallStack()

	if len(frames) != 2 {
		t.Errorf("CallStack() returned %d frames, want 2", len(frames))
	}

	// Verify the frames are returned in correct order
	if frames[0].FunctionName != "func1" {
		t.Errorf("CallStack()[0].FunctionName = %s, want func1", frames[0].FunctionName)
	}
	if frames[1].FunctionName != "func2" {
		t.Errorf("CallStack()[1].FunctionName = %s, want func2", frames[1].FunctionName)
	}
}

// TestExecutionContext_PushEnv tests pushing a new environment.
func TestExecutionContext_PushEnv(t *testing.T) {
	env := newMockEnvironment()
	env.Define("x", 10)

	ctx := NewExecutionContext(env)

	// Push a new environment
	newEnv := ctx.PushEnv()

	if newEnv == nil {
		t.Fatalf("PushEnv() returned nil")
	}

	// The current environment should be the new one
	if ctx.Env() != newEnv {
		t.Errorf("PushEnv() did not set current environment to new environment")
	}

	// The new environment should be enclosed by the original
	// Define a variable in the new environment
	newEnv.Define("y", 20)

	// Should be able to access parent variable
	val, ok := newEnv.Get("x")
	if !ok || val != 10 {
		t.Errorf("PushEnv() new environment cannot access parent variable x")
	}

	// Should be able to access new variable
	val, ok = newEnv.Get("y")
	if !ok || val != 20 {
		t.Errorf("PushEnv() new environment cannot access its own variable y")
	}
}

// TestExecutionContext_PopEnv tests popping an environment.
func TestExecutionContext_PopEnv(t *testing.T) {
	env := newMockEnvironment()
	env.Define("x", 10)

	ctx := NewExecutionContext(env)

	// Push a new environment
	newEnv := ctx.PushEnv()
	newEnv.Define("y", 20)

	// Pop the environment
	restored := ctx.PopEnv()

	if restored != env {
		t.Errorf("PopEnv() did not restore original environment")
	}

	if ctx.Env() != env {
		t.Errorf("PopEnv() did not set current environment to restored environment")
	}

	// Original environment should still have x
	val, ok := env.Get("x")
	if !ok || val != 10 {
		t.Errorf("PopEnv() corrupted original environment")
	}

	// Original environment should not have y (was only in child)
	_, ok = env.Get("y")
	if ok {
		t.Errorf("PopEnv() leaked child variable to parent")
	}
}

// TestExecutionContext_PushPopEnv_Multiple tests multiple push/pop operations.
func TestExecutionContext_PushPopEnv_Multiple(t *testing.T) {
	env := newMockEnvironment()
	env.Define("level", 0)

	ctx := NewExecutionContext(env)

	// Push multiple environments
	env1 := ctx.PushEnv()
	env1.Define("level", 1)

	env2 := ctx.PushEnv()
	env2.Define("level", 2)

	env3 := ctx.PushEnv()
	env3.Define("level", 3)

	// Verify we're at level 3
	val, _ := ctx.Env().Get("level")
	if val != 3 {
		t.Errorf("After 3 pushes, level = %v, want 3", val)
	}

	// Pop back to level 2
	ctx.PopEnv()
	val, _ = ctx.Env().Get("level")
	if val != 2 {
		t.Errorf("After 1 pop, level = %v, want 2", val)
	}

	// Pop back to level 1
	ctx.PopEnv()
	val, _ = ctx.Env().Get("level")
	if val != 1 {
		t.Errorf("After 2 pops, level = %v, want 1", val)
	}

	// Pop back to level 0
	ctx.PopEnv()
	val, _ = ctx.Env().Get("level")
	if val != 0 {
		t.Errorf("After 3 pops, level = %v, want 0", val)
	}

	// Verify we're back at the original environment
	if ctx.Env() != env {
		t.Errorf("After all pops, not back to original environment")
	}
}

// TestExecutionContext_PopEnv_Empty tests popping when stack is empty.
func TestExecutionContext_PopEnv_Empty(t *testing.T) {
	env := newMockEnvironment()
	ctx := NewExecutionContext(env)

	// Pop without pushing should return current environment
	restored := ctx.PopEnv()

	if restored != env {
		t.Errorf("PopEnv() on empty stack should return current environment")
	}

	if ctx.Env() != env {
		t.Errorf("PopEnv() on empty stack changed current environment")
	}
}

// TestExecutionContext_GetOldValue tests retrieving old values.
func TestExecutionContext_GetOldValue(t *testing.T) {
	env := newMockEnvironment()
	ctx := NewExecutionContext(env)

	// Getting old value from empty stack should return nil, false
	val, ok := ctx.GetOldValue("x")
	if ok || val != nil {
		t.Errorf("GetOldValue() on empty stack = (%v, %v), want (nil, false)", val, ok)
	}

	// Push old values
	oldVals1 := map[string]interface{}{"x": 10, "y": 20}
	ctx.PushOldValues(oldVals1)

	// Get existing value
	val, ok = ctx.GetOldValue("x")
	if !ok || val != 10 {
		t.Errorf("GetOldValue(\"x\") = (%v, %v), want (10, true)", val, ok)
	}

	val, ok = ctx.GetOldValue("y")
	if !ok || val != 20 {
		t.Errorf("GetOldValue(\"y\") = (%v, %v), want (20, true)", val, ok)
	}

	// Get non-existing value
	val, ok = ctx.GetOldValue("z")
	if ok || val != nil {
		t.Errorf("GetOldValue(\"z\") = (%v, %v), want (nil, false)", val, ok)
	}

	// Push another set of old values
	oldVals2 := map[string]interface{}{"a": 100, "b": 200}
	ctx.PushOldValues(oldVals2)

	// Should get from top of stack (most recent)
	val, ok = ctx.GetOldValue("a")
	if !ok || val != 100 {
		t.Errorf("GetOldValue(\"a\") from top of stack = (%v, %v), want (100, true)", val, ok)
	}

	// Should not get from previous frame
	val, ok = ctx.GetOldValue("x")
	if ok {
		t.Errorf("GetOldValue(\"x\") should not find value from previous frame")
	}

	// Pop the top frame
	ctx.PopOldValues()

	// Now should get from the first frame again
	val, ok = ctx.GetOldValue("x")
	if !ok || val != 10 {
		t.Errorf("GetOldValue(\"x\") after pop = (%v, %v), want (10, true)", val, ok)
	}
}
