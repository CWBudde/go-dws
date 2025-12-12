package runtime

import (
	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/types"
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
	// FlowReturn indicates a return statement was executed.
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
	return &PropertyEvalContext{PropertyChain: make([]string, 0)}
}

// ExceptionGetter is a callback to read the current exception from external storage.
// Used to sync exception state between interpreter and evaluator.
type ExceptionGetter func() any

// ExceptionSetter is a callback to write an exception to external storage.
// Used to sync exception state between interpreter and evaluator.
type ExceptionSetter func(any)

// ExecutionContext holds all execution state for script evaluation.
//
//nolint:govet // Keep layout stable and readable; alignment optimization is not worth the churn here.
type ExecutionContext struct {
	handlerException  any
	exception         any
	arrayTypeContext  *types.ArrayType
	callStack         *CallStack
	controlFlow       *ControlFlow
	propContext       *PropertyEvalContext
	env               *Environment
	exceptionGetter   ExceptionGetter
	exceptionSetter   ExceptionSetter
	recordTypeContext string
	envStack          []*Environment
	oldValuesStack    []map[string]any
	refCountManager   RefCountManager
}

// NewExecutionContext creates a new execution context with the given environment.
// The call stack is created with the default maximum depth (1024).
func NewExecutionContext(env *Environment) *ExecutionContext {
	return &ExecutionContext{
		env:            env,
		envStack:       make([]*Environment, 0),
		callStack:      NewCallStack(0),
		controlFlow:    NewControlFlow(),
		propContext:    NewPropertyEvalContext(),
		oldValuesStack: make([]map[string]any, 0),
	}
}

// NewExecutionContextWithMaxDepth creates a new execution context with a custom max call stack depth.
func NewExecutionContextWithMaxDepth(env *Environment, maxDepth int) *ExecutionContext {
	return &ExecutionContext{
		env:            env,
		envStack:       make([]*Environment, 0),
		callStack:      NewCallStack(maxDepth),
		controlFlow:    NewControlFlow(),
		propContext:    NewPropertyEvalContext(),
		oldValuesStack: make([]map[string]any, 0),
	}
}

// NewExecutionContextWithCallbacks creates a new execution context with exception callbacks.
func NewExecutionContextWithCallbacks(env *Environment, maxDepth int, getter ExceptionGetter, setter ExceptionSetter) *ExecutionContext {
	return &ExecutionContext{
		env:             env,
		envStack:        make([]*Environment, 0),
		callStack:       NewCallStack(maxDepth),
		controlFlow:     NewControlFlow(),
		propContext:     NewPropertyEvalContext(),
		oldValuesStack:  make([]map[string]any, 0),
		exceptionGetter: getter,
		exceptionSetter: setter,
	}
}

// Env returns the current runtime environment.
func (ctx *ExecutionContext) Env() *Environment {
	return ctx.env
}

// SetEnv updates the runtime environment.
func (ctx *ExecutionContext) SetEnv(env *Environment) {
	ctx.env = env
}

// PushEnv creates a new enclosed environment and pushes the current environment onto the stack.
// Used for entering a new scope (loops, blocks, try-except handlers).
func (ctx *ExecutionContext) PushEnv() *Environment {
	ctx.envStack = append(ctx.envStack, ctx.env)
	newEnv := NewEnclosedEnvironment(ctx.env)
	ctx.env = newEnv
	return newEnv
}

// PopEnv restores the previous environment from the stack.
// Used for exiting a scope (loops, blocks, try-except handlers).
func (ctx *ExecutionContext) PopEnv() *Environment {
	if len(ctx.envStack) == 0 {
		return ctx.env
	}
	ctx.env = ctx.envStack[len(ctx.envStack)-1]
	ctx.envStack = ctx.envStack[:len(ctx.envStack)-1]
	return ctx.env
}

// GetCallStack returns the CallStack instance for direct access.
func (ctx *ExecutionContext) GetCallStack() *CallStack {
	return ctx.callStack
}

// CallStack returns a copy of the current call stack frames.
//
// Deprecated: Use GetCallStack() for full CallStack API access.
func (ctx *ExecutionContext) CallStack() errors.StackTrace {
	return ctx.callStack.Frames()
}

// PushCallStack adds a new frame to the call stack.
//
// Deprecated: Use GetCallStack().Push() for better error handling.
func (ctx *ExecutionContext) PushCallStack(frame errors.StackFrame) {
	_ = ctx.callStack.Push(frame.FunctionName, frame.FileName, frame.Position) //nolint:errcheck // Deprecated wrapper
}

// PopCallStack removes the most recent frame from the call stack.
//
// Deprecated: Use GetCallStack().Pop() instead.
func (ctx *ExecutionContext) PopCallStack() {
	ctx.callStack.Pop()
}

// CallStackDepth returns the current depth of the call stack.
//
// Deprecated: Use GetCallStack().Depth() instead.
func (ctx *ExecutionContext) CallStackDepth() int {
	return ctx.callStack.Depth()
}

// ControlFlow returns the control flow state.
func (ctx *ExecutionContext) ControlFlow() *ControlFlow {
	return ctx.controlFlow
}

// Exception returns the current active exception.
func (ctx *ExecutionContext) Exception() any {
	if ctx.exceptionGetter != nil {
		return ctx.exceptionGetter()
	}
	return ctx.exception
}

// SetException sets the current active exception.
func (ctx *ExecutionContext) SetException(exc any) {
	if ctx.exceptionSetter != nil {
		ctx.exceptionSetter(exc)
		return
	}
	ctx.exception = exc
}

// HandlerException returns the exception being handled in a try-except block.
func (ctx *ExecutionContext) HandlerException() any {
	return ctx.handlerException
}

// SetHandlerException sets the exception being handled.
func (ctx *ExecutionContext) SetHandlerException(exc any) {
	ctx.handlerException = exc
}

// PropContext returns the property evaluation context.
func (ctx *ExecutionContext) PropContext() *PropertyEvalContext {
	return ctx.propContext
}

// RecordTypeContext returns the current record type context for anonymous record literals.
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

// ArrayTypeContext returns the current array type context for array literal evaluation.
func (ctx *ExecutionContext) ArrayTypeContext() *types.ArrayType {
	return ctx.arrayTypeContext
}

// SetArrayTypeContext sets the array type context for array literal evaluation.
func (ctx *ExecutionContext) SetArrayTypeContext(arrayType *types.ArrayType) {
	ctx.arrayTypeContext = arrayType
}

// ClearArrayTypeContext clears the array type context.
func (ctx *ExecutionContext) ClearArrayTypeContext() {
	ctx.arrayTypeContext = nil
}

// PushOldValues saves the current variable values before entering a new scope.
func (ctx *ExecutionContext) PushOldValues(oldValues map[string]any) {
	ctx.oldValuesStack = append(ctx.oldValuesStack, oldValues)
}

// PopOldValues restores the previous variable values when exiting a scope.
func (ctx *ExecutionContext) PopOldValues() map[string]any {
	if len(ctx.oldValuesStack) == 0 {
		return nil
	}
	top := ctx.oldValuesStack[len(ctx.oldValuesStack)-1]
	ctx.oldValuesStack = ctx.oldValuesStack[:len(ctx.oldValuesStack)-1]
	return top
}

// GetOldValue retrieves an old value by name from the current old values map.
func (ctx *ExecutionContext) GetOldValue(name string) (any, bool) {
	if len(ctx.oldValuesStack) == 0 {
		return nil, false
	}
	top := ctx.oldValuesStack[len(ctx.oldValuesStack)-1]
	val, ok := top[name]
	return val, ok
}

// Clone creates a shallow copy of the execution context.
// Note: The environment, call stack, and control flow are shared references.
func (ctx *ExecutionContext) Clone() *ExecutionContext {
	envStackCopy := make([]*Environment, len(ctx.envStack))
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
		arrayTypeContext:  ctx.arrayTypeContext,
		exceptionGetter:   ctx.exceptionGetter,
		exceptionSetter:   ctx.exceptionSetter,
		refCountManager:   ctx.refCountManager,
	}
}

// Reset clears the execution context state for reuse.
func (ctx *ExecutionContext) Reset() {
	ctx.envStack = make([]*Environment, 0)
	ctx.callStack.Clear()
	ctx.controlFlow.Clear()
	ctx.exception = nil
	ctx.handlerException = nil
	ctx.oldValuesStack = make([]map[string]any, 0)
	ctx.propContext = NewPropertyEvalContext()
	ctx.recordTypeContext = ""
	ctx.arrayTypeContext = nil
}

// SetRefCountManager attaches a RefCountManager used by assignment helpers.
func (ctx *ExecutionContext) SetRefCountManager(mgr RefCountManager) {
	ctx.refCountManager = mgr
}

// RefCountManager returns the attached reference counting manager (if any).
func (ctx *ExecutionContext) RefCountManager() RefCountManager {
	return ctx.refCountManager
}
