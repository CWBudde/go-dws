package evaluator

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Debug Introspection Built-ins
// ============================================================================
//
// CurrentSourceCodeLocation, CallerSourceCodeLocation and CurrentStackTrace read
// the live call stack, so they cannot go through the value-only builtin
// registry: each one needs the position of its own occurrence in the source.
//
// The call stack stores, per frame, the routine being executed and the position
// it was called from. So for a stack pushed as
//
//	frames = [Test @ 11:1, Test2 @ 8:4]        // Test2 is executing
//
// "the current routine" is the top frame's name, and "where the current routine
// was called from" is the top frame's position.

// evalDebugBuiltin evaluates one of the call-stack introspection built-ins.
// The second result reports whether name is one of them.
func (e *Evaluator) evalDebugBuiltin(name string, node ast.Node, ctx *ExecutionContext) (Value, bool) {
	switch {
	case ident.Equal(name, builtins.CurrentSourceCodeLocationName):
		return e.currentSourceCodeLocation(node, ctx), true
	case ident.Equal(name, builtins.CallerSourceCodeLocationName):
		return e.callerSourceCodeLocation(ctx), true
	case ident.Equal(name, builtins.CurrentStackTraceName):
		return e.currentStackTrace(node, ctx), true
	default:
		return nil, false
	}
}

// currentSourceCodeLocation reports where the expression itself sits, labeled
// with the routine that contains it (empty in the main program).
func (e *Evaluator) currentSourceCodeLocation(node ast.Node, ctx *ExecutionContext) Value {
	line := int64(0)
	if node != nil {
		line = int64(node.Pos().Line)
	}
	return builtins.NewSourceCodeLocation(builtins.MainModuleName, line, currentRoutineName(ctx))
}

// callerSourceCodeLocation reports where the enclosing routine was called from,
// labeled with the routine that made that call. In the main program there is no
// caller, and DWScript answers with a wholly empty location.
func (e *Evaluator) callerSourceCodeLocation(ctx *ExecutionContext) Value {
	frames := callStackFrames(ctx)
	if len(frames) == 0 {
		return builtins.NewSourceCodeLocation("", 0, "")
	}

	callSite := frames[len(frames)-1]
	line := int64(0)
	if callSite.Position != nil {
		line = int64(callSite.Position.Line)
	}

	caller := ""
	if len(frames) >= 2 {
		caller = frames[len(frames)-2].FunctionName
	}
	return builtins.NewSourceCodeLocation(builtins.MainModuleName, line, caller)
}

// currentStackTrace renders the live call stack the way DWScript reports it for
// an unhandled exception, with one extra innermost line for the routine that
// asked, positioned at the CurrentStackTrace expression itself.
func (e *Evaluator) currentStackTrace(node ast.Node, ctx *ExecutionContext) Value {
	here := errors.StackFrame{FunctionName: currentRoutineName(ctx)}
	if node != nil {
		pos := node.Pos()
		here.Position = &pos
	}

	trace := here.String()
	if below := callStackFrames(ctx).DWScriptString(); below != "" {
		trace += "\n" + below
	}
	return &runtime.StringValue{Value: strings.TrimSuffix(trace, "\n")}
}

// callStackFrames returns the live frames, oldest first.
func callStackFrames(ctx *ExecutionContext) errors.StackTrace {
	if ctx == nil {
		return nil
	}
	stack := ctx.GetCallStack()
	if stack == nil {
		return nil
	}
	return stack.Frames()
}
