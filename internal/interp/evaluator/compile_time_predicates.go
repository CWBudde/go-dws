package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// builtinCompileTimePredicate returns the value the semantic analyzer folded for
// a compile-time predicate call such as Declared or ConditionalDefined.
//
// Those intrinsics answer questions about the compilation itself, so they are
// resolved during semantic analysis and published per call site through the
// semantic metadata table. The evaluator only reads the folded answer; without
// semantic information the call cannot be answered at all.
func (e *Evaluator) builtinCompileTimePredicate(funcName *ast.Identifier, node ast.Node) Value {
	if info := e.SemanticInfo(); info != nil {
		if folded, ok := info.GetSymbol(funcName).(bool); ok {
			return &runtime.BooleanValue{Value: folded}
		}
	}
	return e.newError(node, "%s is a compile-time function and requires semantic analysis", funcName.Value)
}
