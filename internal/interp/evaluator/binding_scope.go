package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// bindingScope tracks scope-owned bindings inside pushed evaluator environments.
// It distinguishes owned bindings from values merely exposed into the env as
// aliases/views of existing runtime state.
type bindingScope struct {
	owned map[string]struct{}
}

func newBindingScope() *bindingScope {
	return &bindingScope{
		owned: make(map[string]struct{}),
	}
}

func (s *bindingScope) defineOwned(e *Evaluator, ctx *ExecutionContext, name string, value Value) Value {
	if s == nil {
		return value
	}
	value = e.retainValueForBinding(value, ctx)
	ctx.Env().Define(name, value)
	s.owned[ident.Normalize(name)] = struct{}{}
	return value
}

func (s *bindingScope) defineExposed(ctx *ExecutionContext, name string, value Value) {
	if s == nil {
		return
	}
	ctx.Env().Define(name, value)
}

func (s *bindingScope) cleanup(e *Evaluator, env *runtime.Environment) {
	if s == nil || env == nil {
		return
	}
	for name := range s.owned {
		value, ok := env.GetLocal(name)
		if !ok {
			continue
		}
		e.releaseValueForBinding(value)
	}
}
