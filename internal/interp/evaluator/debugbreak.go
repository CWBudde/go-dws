package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// VisitDebugBreakExpression executes the intrinsic's empty runtime operation,
// matching DWScript's TDebugBreakExpr.EvalNoResult.
func (e *Evaluator) VisitDebugBreakExpression(_ *ast.DebugBreakExpression, _ *ExecutionContext) Value {
	return &runtime.NilValue{}
}
