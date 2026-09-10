package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// executeNativeClassMethod runs a built-in class method whose body is written
// in Go (runtime.MethodMetadata.Native) rather than in DWScript.
//
// A native method reports failure by returning an error; that error becomes a
// catchable DWScript Exception whose Message carries the error text plus the
// call site position, matching what script-level `raise` produces.
func (e *Evaluator) executeNativeClassMethod(
	classInfo runtime.IClassInfo,
	callable *runtime.MethodMetadata,
	args []Value,
	node ast.Node,
	ctx *ExecutionContext,
) Value {
	if len(args) != len(callable.Parameters) {
		return e.newError(node, "%s.%s expects %d argument(s), got %d",
			classInfo.GetName(), callable.Name, len(callable.Parameters), len(args))
	}

	result, err := callable.Native(classInfo, args)
	if err != nil {
		return e.raiseNativeMethodError(err, node, ctx)
	}
	if result == nil {
		return &runtime.NullValue{}
	}
	return result
}

// raiseNativeMethodError converts a native method's error into a DWScript
// exception that `try ... except` can catch, and returns the error value that
// unwinds the current evaluation.
func (e *Evaluator) raiseNativeMethodError(err error, node ast.Node, ctx *ExecutionContext) Value {
	message := err.Error()

	// DWScript qualifies the message with the routine the failing statement
	// belongs to, and points at the statement rather than the sub-expression.
	if name := enclosingRoutineName(ctx); name != "" {
		message += " in " + name
	}

	reportNode := node
	if ctx != nil {
		if stmt := ctx.CurrentStatement(); stmt != nil {
			reportNode = stmt
		}
	}

	var pos *lexer.Position
	if reportNode != nil {
		nodePos := reportNode.Pos()
		if nodePos.Line > 0 {
			message = fmt.Sprintf("%s [line: %d, column: %d]", message, nodePos.Line, nodePos.Column)
			pos = &lexer.Position{Line: nodePos.Line, Column: nodePos.Column}
		}
	}

	if ctx != nil {
		ctx.SetException(e.createException("Exception", message, pos, ctx))
	}
	return &runtime.ErrorValue{Message: message}
}

// enclosingRoutineName returns the name of the routine whose body is currently
// executing, or "" at the top level of the main program.
func enclosingRoutineName(ctx *ExecutionContext) string {
	if ctx == nil {
		return ""
	}
	stack := ctx.GetCallStack()
	if stack == nil {
		return ""
	}
	frame := stack.Current()
	if frame == nil {
		return ""
	}
	return frame.FunctionName
}
