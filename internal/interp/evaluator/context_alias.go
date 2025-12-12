package evaluator

import "github.com/cwbudde/go-dws/internal/interp/runtime"

// Type aliases to keep evaluator public surface stable while moving runtime state
// into the runtime package (and out of the interpreter/evaluator layering).

type (
	CallStack           = runtime.CallStack
	ControlFlow         = runtime.ControlFlow
	ControlFlowKind     = runtime.ControlFlowKind
	ExecutionContext    = runtime.ExecutionContext
	ExceptionGetter     = runtime.ExceptionGetter
	ExceptionSetter     = runtime.ExceptionSetter
	PropertyEvalContext = runtime.PropertyEvalContext
)

const (
	FlowNone     = runtime.FlowNone
	FlowBreak    = runtime.FlowBreak
	FlowContinue = runtime.FlowContinue
	FlowExit     = runtime.FlowExit
	FlowReturn   = runtime.FlowReturn
)

func NewCallStack(maxDepth int) *CallStack { return runtime.NewCallStack(maxDepth) }
func NewControlFlow() *ControlFlow         { return runtime.NewControlFlow() }
func NewPropertyEvalContext() *PropertyEvalContext {
	return runtime.NewPropertyEvalContext()
}

func NewExecutionContext(env *runtime.Environment) *ExecutionContext {
	return runtime.NewExecutionContext(env)
}

func NewExecutionContextWithMaxDepth(env *runtime.Environment, maxDepth int) *ExecutionContext {
	return runtime.NewExecutionContextWithMaxDepth(env, maxDepth)
}

func NewExecutionContextWithCallbacks(env *runtime.Environment, maxDepth int, getter ExceptionGetter, setter ExceptionSetter) *ExecutionContext {
	return runtime.NewExecutionContextWithCallbacks(env, maxDepth, getter, setter)
}
