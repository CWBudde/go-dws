package evaluator

import (
	"errors"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// preparePropertyIndices captures each original operand before evaluating the
// next. In particular, a later operand cannot rebind a captured array slot.
func (e *Evaluator) preparePropertyIndices(prop *types.PropertyInfo, expressions []ast.Expression, node ast.Node, ctx *ExecutionContext) ([]Value, Value) {
	if len(prop.IndexParamModes) != 0 && len(prop.IndexParamModes) != indexedPropertyArity(prop) {
		return nil, e.newError(node, "indexed property '%s' has incomplete index mode metadata", prop.Name)
	}
	if err := e.checkIndexedPropertyArity(prop, len(expressions), node); err != nil {
		return nil, err
	}
	args := make([]Value, len(expressions))
	for i, expr := range expressions {
		if ctx.Exception() != nil {
			return nil, e.nilValue()
		}
		if prop.IndexMode(i) == types.PropertyIndexVar {
			ref, err := e.prepareByRefArgument(expr, ctx)
			// Keep the raised script exception, including its original message.
			if ctx.Exception() != nil {
				return nil, e.nilValue()
			}
			if err != nil {
				var original *argumentValueError
				if errors.As(err, &original) {
					return nil, original.value
				}
				return nil, e.newError(node, "%s", err)
			}
			args[i] = ref
		} else {
			args[i] = e.Eval(expr, ctx)
		}
		if isError(args[i]) {
			return nil, args[i]
		}
		if ctx.Exception() != nil {
			return nil, e.nilValue()
		}
	}
	return args, nil
}
