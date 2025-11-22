package evaluator

import (
	"github.com/cwbudde/go-dws/internal/errors"
)

// ControlFlowKind represents the type of control flow signal.
type ControlFlowKind int

const (
	// FlowNone indicates normal execution (no control flow signal).
	FlowNone ControlFlowKind = iota
	// FlowBreak indicates a break statement was executed.
	FlowBreak
	// FlowContinue indicates a continue statement was executed.
	FlowContinue
	// FlowExit indicates an exit statement was executed.
	FlowExit
	// FlowReturn indicates a return statement was executed (reserved for future use).
	FlowReturn
)

// String returns a string representation of the control flow kind.
func (k ControlFlowKind) String() string {
	switch k {
	case FlowNone:
		return "none"
	case FlowBreak:
		return "break"
	case FlowContinue:
		return "continue"
	case FlowExit:
		return "exit"
	case FlowReturn:
		return "return"
	default:
		return "unknown"
	}
}

// ControlFlow manages control flow signals (break, continue, exit, return).
// It replaces the boolean flags (breakSignal, continueSignal, exitSignal)
// with a single explicit state value for cleaner control flow handling.
type ControlFlow struct {
	kind ControlFlowKind
}

// NewControlFlow creates a new ControlFlow with no signal (FlowNone).
func NewControlFlow() *ControlFlow {
	return &ControlFlow{kind: FlowNone}
}

// Kind returns the current control flow kind.
func (cf *ControlFlow) Kind() ControlFlowKind {
	return cf.kind
}

// IsActive returns true if there is an active control flow signal.
func (cf *ControlFlow) IsActive() bool {
	return cf.kind != FlowNone
}

// Clear resets the control flow to FlowNone (no signal).
func (cf *ControlFlow) Clear() {
	cf.kind = FlowNone
}

// SetBreak signals that a break statement was executed.
func (cf *ControlFlow) SetBreak() {
	cf.kind = FlowBreak
}

// SetContinue signals that a continue statement was executed.
func (cf *ControlFlow) SetContinue() {
	cf.kind = FlowContinue
}

// SetExit signals that an exit statement was executed.
func (cf *ControlFlow) SetExit() {
	cf.kind = FlowExit
}

// SetReturn signals that a return statement was executed.
// This is reserved for future use in Phase 3.3.2.
func (cf *ControlFlow) SetReturn() {
	cf.kind = FlowReturn
}

// IsBreak returns true if the signal is a break.
func (cf *ControlFlow) IsBreak() bool {
	return cf.kind == FlowBreak
}

// IsContinue returns true if the signal is a continue.
func (cf *ControlFlow) IsContinue() bool {
	return cf.kind == FlowContinue
}

// IsExit returns true if the signal is an exit.
func (cf *ControlFlow) IsExit() bool {
	return cf.kind == FlowExit
}

// IsReturn returns true if the signal is a return.
func (cf *ControlFlow) IsReturn() bool {
	return cf.kind == FlowReturn
}

// PropertyEvalContext tracks the state during property getter/setter evaluation.
// This prevents infinite recursion when evaluating properties.
type PropertyEvalContext struct {
	PropertyChain    []string
	InPropertyGetter bool
	InPropertySetter bool
}

// NewPropertyEvalContext creates a new PropertyEvalContext.
func NewPropertyEvalContext() *PropertyEvalContext {
	return &PropertyEvalContext{
		PropertyChain: make([]string, 0),
	}
}

// Environment represents the runtime environment for variable storage and scoping.
// This is temporarily defined here to avoid circular imports.
// In Phase 3.4, this will be properly organized.
type Environment interface {
	// Define creates a new variable binding in the current scope.
	Define(name string, value interface{})
	// Get retrieves a variable value by name.
	Get(name string) (interface{}, bool)
	// Set updates an existing variable value.
	Set(name string, value interface{}) bool
	// NewEnclosedEnvironment creates a new child scope.
	NewEnclosedEnvironment() Environment
}

// ExecutionContext holds all execution state that was previously scattered
// throughout the Interpreter struct. This separation makes the execution
// state explicit and easier to manage.
//
// Phase 3.3.1: Initial implementation with basic state separation.
// The context is passed to Eval methods to make execution state explicit.
// Phase 3.3.3: Updated to use CallStack abstraction with overflow detection.
// Phase 3.5.4 - Phase 2D: Added envStack for proper environment scoping.
type ExecutionContext struct {
	env               Environment
	exception         interface{}
	handlerException  interface{}
	callStack         *CallStack
	controlFlow       *ControlFlow
	propContext       *PropertyEvalContext
	envStack          []Environment
	oldValuesStack    []map[string]interface{}
	recordTypeContext string // Type context for anonymous record literals (avoids AST mutation)
}

// NewExecutionContext creates a new execution context with the given environment.
// The call stack is created with the default maximum depth (1024).
func NewExecutionContext(env Environment) *ExecutionContext {
	return &ExecutionContext{
		env:            env,
		envStack:       make([]Environment, 0),
		callStack:      NewCallStack(0), // 0 uses default max depth (1024)
		controlFlow:    NewControlFlow(),
		propContext:    NewPropertyEvalContext(),
		oldValuesStack: make([]map[string]interface{}, 0),
	}
}

// NewExecutionContextWithMaxDepth creates a new execution context with a custom max call stack depth.
func NewExecutionContextWithMaxDepth(env Environment, maxDepth int) *ExecutionContext {
	return &ExecutionContext{
		env:            env,
		envStack:       make([]Environment, 0),
		callStack:      NewCallStack(maxDepth),
		controlFlow:    NewControlFlow(),
		propContext:    NewPropertyEvalContext(),
		oldValuesStack: make([]map[string]interface{}, 0),
	}
}

// Env returns the current runtime environment.
func (ctx *ExecutionContext) Env() Environment {
	return ctx.env
}

// SetEnv updates the runtime environment.
func (ctx *ExecutionContext) SetEnv(env Environment) {
	ctx.env = env
}

// PushEnv creates a new enclosed environment and pushes the current environment onto the stack.
// This is used for entering a new scope (loops, blocks, try-except handlers).
// The current environment is saved on the stack, and a new enclosed environment becomes current.
// Returns the new environment for convenience.
//
// Phase 3.5.4 - Phase 2D: Environment Scoping infrastructure.
func (ctx *ExecutionContext) PushEnv() Environment {
	// Save the current environment on the stack
	ctx.envStack = append(ctx.envStack, ctx.env)

	// Create a new enclosed environment
	newEnv := ctx.env.NewEnclosedEnvironment()
	ctx.env = newEnv

	return newEnv
}

// PopEnv restores the previous environment from the stack.
// This is used for exiting a scope (loops, blocks, try-except handlers).
// Returns the restored environment, or the current environment if the stack is empty.
//
// Phase 3.5.4 - Phase 2D: Environment Scoping infrastructure.
func (ctx *ExecutionContext) PopEnv() Environment {
	if len(ctx.envStack) == 0 {
		// Stack is empty - already at root, nothing to pop
		return ctx.env
	}

	// Pop the last environment from the stack
	ctx.env = ctx.envStack[len(ctx.envStack)-1]
	ctx.envStack = ctx.envStack[:len(ctx.envStack)-1]

	return ctx.env
}

// GetCallStack returns the CallStack instance for direct access.
// Phase 3.3.3: Provides access to the CallStack abstraction.
func (ctx *ExecutionContext) GetCallStack() *CallStack {
	return ctx.callStack
}

// CallStack returns a copy of the current call stack frames.
// Deprecated: Use GetCallStack() for full CallStack API access.
func (ctx *ExecutionContext) CallStack() errors.StackTrace {
	return ctx.callStack.Frames()
}

// PushCallStack adds a new frame to the call stack.
// Returns an error if the stack would overflow.
// Deprecated: Use GetCallStack().Push() for better error handling.
func (ctx *ExecutionContext) PushCallStack(frame errors.StackFrame) {
	// For backward compatibility, we ignore the error
	// Real code should use GetCallStack().Push() for proper error handling
	_ = ctx.callStack.Push(frame.FunctionName, frame.FileName, frame.Position)
}

// PopCallStack removes the most recent frame from the call stack.
// Deprecated: Use GetCallStack().Pop() instead.
func (ctx *ExecutionContext) PopCallStack() {
	ctx.callStack.Pop()
}

// CallStackDepth returns the current depth of the call stack.
// Deprecated: Use GetCallStack().Depth() instead.
func (ctx *ExecutionContext) CallStackDepth() int {
	return ctx.callStack.Depth()
}

// ControlFlow returns the control flow state.
func (ctx *ExecutionContext) ControlFlow() *ControlFlow {
	return ctx.controlFlow
}

// Exception returns the current active exception.
func (ctx *ExecutionContext) Exception() interface{} {
	return ctx.exception
}

// SetException sets the current active exception.
func (ctx *ExecutionContext) SetException(exc interface{}) {
	ctx.exception = exc
}

// HandlerException returns the exception being handled in a try-except block.
func (ctx *ExecutionContext) HandlerException() interface{} {
	return ctx.handlerException
}

// SetHandlerException sets the exception being handled.
func (ctx *ExecutionContext) SetHandlerException(exc interface{}) {
	ctx.handlerException = exc
}

// PropContext returns the property evaluation context.
func (ctx *ExecutionContext) PropContext() *PropertyEvalContext {
	return ctx.propContext
}

// RecordTypeContext returns the current record type context for anonymous record literals.
// This allows passing type information to record literal evaluation without mutating the AST.
func (ctx *ExecutionContext) RecordTypeContext() string {
	return ctx.recordTypeContext
}

// SetRecordTypeContext sets the record type context for anonymous record literals.
func (ctx *ExecutionContext) SetRecordTypeContext(typeName string) {
	ctx.recordTypeContext = typeName
}

// ClearRecordTypeContext clears the record type context.
func (ctx *ExecutionContext) ClearRecordTypeContext() {
	ctx.recordTypeContext = ""
}

// PushOldValues saves the current variable values before entering a new scope.
func (ctx *ExecutionContext) PushOldValues(oldValues map[string]interface{}) {
	ctx.oldValuesStack = append(ctx.oldValuesStack, oldValues)
}

// PopOldValues restores the previous variable values when exiting a scope.
func (ctx *ExecutionContext) PopOldValues() map[string]interface{} {
	if len(ctx.oldValuesStack) == 0 {
		return nil
	}
	top := ctx.oldValuesStack[len(ctx.oldValuesStack)-1]
	ctx.oldValuesStack = ctx.oldValuesStack[:len(ctx.oldValuesStack)-1]
	return top
}

// GetOldValue retrieves an old value by name from the current old values map.
// This is used for 'old' expressions in postconditions.
// Returns the value and true if found, nil and false otherwise.
func (ctx *ExecutionContext) GetOldValue(name string) (interface{}, bool) {
	if len(ctx.oldValuesStack) == 0 {
		return nil, false
	}
	top := ctx.oldValuesStack[len(ctx.oldValuesStack)-1]
	val, ok := top[name]
	return val, ok
}

// Clone creates a shallow copy of the execution context.
// This is useful when you need to fork execution (e.g., for parallel evaluation).
// Note: The environment, call stack, and control flow are shared references.
func (ctx *ExecutionContext) Clone() *ExecutionContext {
	// Clone the envStack slice
	envStackCopy := make([]Environment, len(ctx.envStack))
	copy(envStackCopy, ctx.envStack)

	return &ExecutionContext{
		env:               ctx.env,
		envStack:          envStackCopy,
		callStack:         ctx.callStack,
		controlFlow:       ctx.controlFlow,
		exception:         ctx.exception,
		handlerException:  ctx.handlerException,
		oldValuesStack:    ctx.oldValuesStack,
		propContext:       ctx.propContext,
		recordTypeContext: ctx.recordTypeContext,
	}
}

// Reset clears the execution context state for reuse.
// This is useful when you want to reset the context without creating a new one.
func (ctx *ExecutionContext) Reset() {
	ctx.envStack = make([]Environment, 0)
	ctx.callStack.Clear()
	ctx.controlFlow.Clear()
	ctx.exception = nil
	ctx.handlerException = nil
	ctx.oldValuesStack = make([]map[string]interface{}, 0)
	ctx.propContext = NewPropertyEvalContext()
}

// Package-level documentation
// This package provides execution context management for the DWScript interpreter.
// The ExecutionContext separates execution state from the Interpreter struct,
// improving maintainability and making execution state explicit.
//
// Phase 3.3.1: Extract execution state from Interpreter into ExecutionContext.
// Phase 3.3.2: Implement explicit control flow handling with ControlFlow type.
// Phase 3.3.3: Create CallStack abstraction with stack overflow detection.
