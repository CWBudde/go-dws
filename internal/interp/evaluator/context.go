package evaluator

import (
	"fmt"
	"io"
	"math/rand"

	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
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

// ExceptionGetter is a callback to read the current exception from external storage.
// Used to sync exception state between interpreter and evaluator.
type ExceptionGetter func() any

// ExceptionSetter is a callback to write an exception to external storage.
// Used to sync exception state between interpreter and evaluator.
type ExceptionSetter func(any)

// ExecutionContext holds all execution state that was previously scattered
// throughout the Interpreter struct. This separation makes the execution
// state explicit and easier to manage.
//
// Phase 3.1.3: Now uses concrete *runtime.Environment instead of interface.
// This eliminates the need for EnvironmentAdapter.
//
// Phase 3.2: Added exception callbacks for syncing with interpreter's i.exception.
type ExecutionContext struct {
	env               *runtime.Environment
	exception         any
	handlerException  any
	callStack         *CallStack
	controlFlow       *ControlFlow
	propContext       *PropertyEvalContext
	arrayTypeContext  *types.ArrayType
	evaluator         *Evaluator
	recordTypeContext string
	envStack          []*runtime.Environment
	oldValuesStack    []map[string]any
	// Exception callbacks for unified exception handling (Phase 3.2)
	exceptionGetter ExceptionGetter
	exceptionSetter ExceptionSetter
}

// NewExecutionContext creates a new execution context with the given environment.
// The call stack is created with the default maximum depth (1024).
func NewExecutionContext(env *runtime.Environment) *ExecutionContext {
	return &ExecutionContext{
		env:            env,
		envStack:       make([]*runtime.Environment, 0),
		callStack:      NewCallStack(0), // 0 uses default max depth (1024)
		controlFlow:    NewControlFlow(),
		propContext:    NewPropertyEvalContext(),
		oldValuesStack: make([]map[string]any, 0),
	}
}

// NewExecutionContextWithMaxDepth creates a new execution context with a custom max call stack depth.
func NewExecutionContextWithMaxDepth(env *runtime.Environment, maxDepth int) *ExecutionContext {
	return &ExecutionContext{
		env:            env,
		envStack:       make([]*runtime.Environment, 0),
		callStack:      NewCallStack(maxDepth),
		controlFlow:    NewControlFlow(),
		propContext:    NewPropertyEvalContext(),
		oldValuesStack: make([]map[string]interface{}, 0),
	}
}

// NewExecutionContextWithCallbacks creates a new execution context with exception callbacks.
// This allows the interpreter to provide its own exception storage (i.exception) while
// the evaluator uses ctx.Exception()/SetException() seamlessly.
// Phase 3.2: Enables unified exception handling between interpreter and evaluator.
func NewExecutionContextWithCallbacks(env *runtime.Environment, maxDepth int, getter ExceptionGetter, setter ExceptionSetter) *ExecutionContext {
	return &ExecutionContext{
		env:             env,
		envStack:        make([]*runtime.Environment, 0),
		callStack:       NewCallStack(maxDepth),
		controlFlow:     NewControlFlow(),
		propContext:     NewPropertyEvalContext(),
		oldValuesStack:  make([]map[string]interface{}, 0),
		exceptionGetter: getter,
		exceptionSetter: setter,
	}
}

// Env returns the current runtime environment.
func (ctx *ExecutionContext) Env() *runtime.Environment {
	return ctx.env
}

// SetEnv updates the runtime environment.
func (ctx *ExecutionContext) SetEnv(env *runtime.Environment) {
	ctx.env = env
}

// PushEnv creates a new enclosed environment and pushes the current environment onto the stack.
// This is used for entering a new scope (loops, blocks, try-except handlers).
// The current environment is saved on the stack, and a new enclosed environment becomes current.
// Returns the new environment for convenience.
//
// Phase 3.5.4 - Phase 2D: Environment Scoping infrastructure.
func (ctx *ExecutionContext) PushEnv() *runtime.Environment {
	// Save the current environment on the stack
	ctx.envStack = append(ctx.envStack, ctx.env)

	// Create a new enclosed environment
	newEnv := runtime.NewEnclosedEnvironment(ctx.env)
	ctx.env = newEnv

	return newEnv
}

// PopEnv restores the previous environment from the stack.
// This is used for exiting a scope (loops, blocks, try-except handlers).
// Returns the restored environment, or the current environment if the stack is empty.
//
// Phase 3.5.4 - Phase 2D: Environment Scoping infrastructure.
func (ctx *ExecutionContext) PopEnv() *runtime.Environment {
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
// If exception callbacks are configured (Phase 3.2), uses external storage.
func (ctx *ExecutionContext) Exception() interface{} {
	if ctx.exceptionGetter != nil {
		return ctx.exceptionGetter()
	}
	return ctx.exception
}

// SetException sets the current active exception.
// If exception callbacks are configured (Phase 3.2), uses external storage.
func (ctx *ExecutionContext) SetException(exc interface{}) {
	if ctx.exceptionSetter != nil {
		ctx.exceptionSetter(exc)
		return
	}
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
	envStackCopy := make([]*runtime.Environment, len(ctx.envStack))
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
		evaluator:         ctx.evaluator,         // Task 3.5.41: Copy evaluator reference
		exceptionGetter:   ctx.exceptionGetter,   // Phase 3.2: Copy exception callbacks
		exceptionSetter:   ctx.exceptionSetter,   // Phase 3.2: Copy exception callbacks
	}
}

// Reset clears the execution context state for reuse.
// This is useful when you want to reset the context without creating a new one.
func (ctx *ExecutionContext) Reset() {
	ctx.envStack = make([]*runtime.Environment, 0)
	ctx.callStack.Clear()
	ctx.controlFlow.Clear()
	ctx.exception = nil
	ctx.handlerException = nil
	ctx.oldValuesStack = make([]map[string]interface{}, 0)
	ctx.propContext = NewPropertyEvalContext()
}

// SetEvaluator sets the evaluator reference in the execution context.
// Task 3.5.41: Enables access to RefCountManager from assignment helpers.
func (ctx *ExecutionContext) SetEvaluator(evaluator *Evaluator) {
	ctx.evaluator = evaluator
}

// RefCountManager returns the reference counting manager from the evaluator.
// Task 3.5.41: Enables assignment helpers to manage object lifecycles.
func (ctx *ExecutionContext) RefCountManager() runtime.RefCountManager {
	if ctx.evaluator != nil {
		return ctx.evaluator.RefCountManager()
	}
	return nil
}

// Package-level documentation
// This package provides execution context management for the DWScript interpreter.
// The ExecutionContext separates execution state from the Interpreter struct,
// improving maintainability and making execution state explicit.
//
// Phase 3.3.1: Extract execution state from Interpreter into ExecutionContext.
// Phase 3.3.2: Implement explicit control flow handling with ControlFlow type.
// Phase 3.3.3: Create CallStack abstraction with stack overflow detection.

// ============================================================================
// Context Interface Implementation
// ============================================================================
//
// This section implements the builtins.Context interface methods for the Evaluator.
// These methods provide core functionality for built-in functions:
// - Error creation with location information
// - Current AST node access
// - Random number generation
// - I/O operations
// - Value inspection
// ============================================================================

// ============================================================================
// Core State & Error Methods
// ============================================================================

// NewError creates an error value with location information from the current node.
// It formats the message using fmt.Sprintf semantics.
//
// This implements the builtins.Context interface method NewError().
func (e *Evaluator) NewError(format string, args ...interface{}) Value {
	return e.newError(e.currentNode, format, args...)
}

// Note: CurrentNode() is already implemented in evaluator.go:1053

// RandSource returns the random number generator for built-in functions
// like Random(), RandomInt(), and RandG().
//
// This implements the builtins.Context interface method RandSource().
func (e *Evaluator) RandSource() *rand.Rand {
	return e.rand
}

// ============================================================================
// Random Number Methods
// ============================================================================

// GetRandSeed returns the current random number generator seed value.
// Used by the RandSeed() built-in function.
//
// This implements the builtins.Context interface method GetRandSeed().
func (e *Evaluator) GetRandSeed() int64 {
	return e.randSeed
}

// SetRandSeed sets the random number generator seed.
// Used by the SetRandSeed() and Randomize() built-in functions.
//
// This implements the builtins.Context interface method SetRandSeed().
func (e *Evaluator) SetRandSeed(seed int64) {
	e.randSeed = seed
	e.rand.Seed(seed)
}

// ============================================================================
// I/O Methods
// ============================================================================

// Write outputs a string to the configured output writer without a newline.
// Used by the Print() built-in function.
//
// This implements the builtins.Context interface method Write().
func (e *Evaluator) Write(s string) {
	if e.output != nil {
		io.WriteString(e.output, s)
	}
}

// WriteLine outputs a string to the configured output writer with a newline.
// Used by the PrintLn() built-in function.
//
// This implements the builtins.Context interface method WriteLine().
func (e *Evaluator) WriteLine(s string) {
	if e.output != nil {
		fmt.Fprintln(e.output, s)
	}
}

// ============================================================================
// Value Inspection Method
// ============================================================================

// IsAssigned checks if a Variant value has been assigned (is not uninitialized).
// Returns true if the value is assigned, false if it's an uninitialized Variant.
//
// This implements the builtins.Context interface method IsAssigned().
func (e *Evaluator) IsAssigned(value Value) bool {
	// Check if it's nil
	if value == nil {
		return false
	}

	// Check if it's an InterfaceInstance
	if intfVal, ok := value.(*runtime.InterfaceInstance); ok {
		// Interface is assigned if its Object field is not nil
		return intfVal.Object != nil
	}

	// Check if it's a VariantWrapper (runtime.VariantWrapper interface)
	if wrapper, ok := value.(runtime.VariantWrapper); ok {
		unwrapped := wrapper.UnwrapVariant()
		// Uninitialized variants unwrap to nil
		return unwrapped != nil
	}

	// Non-variant values are always considered assigned
	return true
}

// ============================================================================
// Call Stack Methods
// ============================================================================

// GetCallStackString returns a formatted string representation of the current call stack.
// This implements the builtins.Context interface.
func (e *Evaluator) GetCallStackString() string {
	if e.currentContext == nil {
		return ""
	}
	return e.currentContext.GetCallStack().String()
}

// GetCallStackArray returns the current call stack as an array of records.
// This implements the builtins.Context interface.
func (e *Evaluator) GetCallStackArray() Value {
	// Handle nil context
	if e.currentContext == nil {
		return &runtime.ArrayValue{
			Elements:  []runtime.Value{},
			ArrayType: types.NewDynamicArrayType(types.VARIANT),
		}
	}

	// Get frames from call stack
	frames := e.currentContext.GetCallStack().Frames()
	elements := make([]runtime.Value, len(frames))

	for idx, frame := range frames {
		// Create a record with FunctionName, Line, Column fields
		fields := make(map[string]runtime.Value)
		fields["FunctionName"] = &runtime.StringValue{Value: frame.FunctionName}

		// Extract line and column from Position
		if frame.Position != nil {
			fields["Line"] = &runtime.IntegerValue{Value: int64(frame.Position.Line)}
			fields["Column"] = &runtime.IntegerValue{Value: int64(frame.Position.Column)}
		} else {
			fields["Line"] = &runtime.IntegerValue{Value: 0}
			fields["Column"] = &runtime.IntegerValue{Value: 0}
		}

		elements[idx] = &runtime.RecordValue{
			Fields:     fields,
			RecordType: nil, // Anonymous record (no type metadata needed)
		}
	}

	// Create and return the array
	return &runtime.ArrayValue{
		Elements:  elements,
		ArrayType: types.NewDynamicArrayType(types.VARIANT),
	}
}

// ============================================================================
// Exception Raising Methods
// ============================================================================

// RaiseAssertionFailed raises an EAssertionFailed exception with an optional custom message.
// The exception includes position information from the current node.
// This implements the builtins.Context interface.
func (e *Evaluator) RaiseAssertionFailed(customMessage string) {
	// Delegate to the adapter since exception handling is still in the Interpreter.
	// The adapter will create the EAssertionFailed exception instance with proper
	// position information and custom message, then store it in i.exception.
	e.exceptionMgr.RaiseAssertionFailed(customMessage)
}

// ============================================================================
// Function Pointer Delegation
// ============================================================================

// EvalFunctionPointer executes a function pointer with given arguments.
//
// IMPORTANT: This method delegates to the adapter (not migrated to Evaluator).
// This is the ONLY Context method that still uses adapter delegation.
//
// This implements the builtins.Context interface by delegating to the adapter's
// CallFunctionPointer method.
func (e *Evaluator) EvalFunctionPointer(funcPtr Value, args []Value) Value {
	// Delegate to adapter.CallFunctionPointer with current node for error reporting
	return e.oopEngine.CallFunctionPointer(funcPtr, args, e.currentNode)
}
