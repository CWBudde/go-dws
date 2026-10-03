package evaluator

import "github.com/cwbudde/go-dws/internal/interp/runtime"

// capturedLazyArgument keeps ordinary thunk evaluation unchanged while allowing
// lazy reads to propagate supplier exceptions to their consuming context.
type capturedLazyArgument struct {
	*runtime.LazyThunk
	captured       *ExecutionContext
	contextualEval func() Value
}

func (argument *capturedLazyArgument) evaluateInContext(ctx *ExecutionContext) Value {
	value := argument.contextualEval()
	if argument.captured != ctx && argument.captured.Exception() != nil {
		ctx.SetException(argument.captured.Exception())
		argument.captured.SetException(nil)
	}
	return value
}

func forceLazyArgument(thunk LazyEvaluator, ctx *ExecutionContext) Value {
	if contextual, ok := thunk.(interface{ evaluateInContext(*ExecutionContext) Value }); ok {
		return contextual.evaluateInContext(ctx)
	}
	return thunk.Evaluate()
}
