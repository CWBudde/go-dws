package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// VisitRaiseStatement evaluates a raise statement (exception throwing).
func (e *Evaluator) VisitRaiseStatement(node *ast.RaiseStatement, ctx *ExecutionContext) Value {
	// Bare raise - re-raise current exception
	if node.Exception == nil {
		// Use the exception saved by evalExceptClause
		if ctx.HandlerException() != nil {
			// Re-raise the exception
			ctx.SetException(ctx.HandlerException())
			return nil
		}

		panic("runtime error: bare raise with no active exception")
	}

	excVal := e.Eval(node.Exception, ctx)
	if isError(excVal) {
		return excVal
	}

	// Raising a nil exception reference raises "Object not instantiated",
	// reported just past the raised expression (DWScript reports the parser's
	// position after consuming the expression).
	if excVal == nil || runtime.KindOf(excVal) == runtime.KindNil {
		pos := raisedExpressionEndPos(node.Exception)
		message := fmt.Sprintf("Object not instantiated [line: %d, column: %d]", pos.Line, pos.Column)
		ctx.SetException(e.createException("Exception", message, nil, ctx))
		return nil
	}

	// DWScript reports an unhandled raise just past the raised expression (the
	// parser's position after consuming it) ...
	pos := node.Exception.End()

	// `raise ExceptObject` inside a handler re-raises the exception already in
	// flight rather than starting a new one: it keeps the original message,
	// position and "user raised" status, and only records where it was re-raised
	// (SimpleScripts/re_raise re-raises a runtime division-by-zero and still
	// expects "Division by zero", not "User defined exception: ...").
	if reRaised := reRaisedException(ctx, excVal); reRaised != nil {
		reRaised.ReRaisePos = &pos
		// The reported call stack is the one at the re-raise, not at the original
		// raise: DWScript lists the frames the exception is escaping through now.
		reRaised.CallStack = ctx.GetCallStack().Frames()
		ctx.SetException(reRaised)
		return nil
	}

	excObj := e.createExceptionFromObject(excVal, ctx, &pos)
	if excValue, ok := excObj.(*runtime.ExceptionValue); ok {
		excValue.UserRaised = true
		// ... but the innermost stack-trace frame is the site where the
		// exception object was constructed, at the constructor's name token.
		originPos := raiseSitePos(node.Exception)
		excValue.OriginPos = &originPos
	}
	ctx.SetException(excObj)

	return nil
}

// raisedExpressionEndPos approximates the source position immediately after a
// raised expression (identifiers advance by their length; other expressions
// fall back to their start position).
func raisedExpressionEndPos(expr ast.Expression) token.Position {
	pos := expr.Pos()
	if identExpr, ok := expr.(*ast.Identifier); ok {
		pos.Column += len(identExpr.Value)
	}
	return pos
}

// reRaisedException reports the exception `raise expr` is re-raising: the one the
// enclosing handler is already processing, identified by the object instance bound
// to the handler variable and to ExceptObject. Returns nil when expr names a
// different object, which is an ordinary raise.
func reRaisedException(ctx *ExecutionContext, raised Value) *runtime.ExceptionValue {
	handled, ok := ctx.HandlerException().(*runtime.ExceptionValue)
	if !ok || handled == nil || handled.Instance == nil {
		return nil
	}
	inst, ok := raised.(*runtime.ObjectInstance)
	if !ok || inst != handled.Instance {
		return nil
	}
	return handled
}
