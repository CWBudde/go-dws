package semantic

import "github.com/cwbudde/go-dws/internal/types"

// exportImplementsForward snapshots the local branch DefineOverload will bind
// before replacement clears its forward state. A non-overloaded forward binds
// by name even when its header disagrees; explicit overloads bind a signature.
func (st *SymbolTable) exportImplementsForward(name string, signature *types.FunctionType) bool {
	existing, ok := st.symbols.Get(name)
	if !ok {
		return false
	}
	if !existing.IsOverloadSet && !existing.HasOverloadDirective {
		_, function := existing.Type.(*types.FunctionType)
		return existing.IsForward && function
	}
	candidates := existing.Overloads
	if !existing.IsOverloadSet {
		candidates = []*Symbol{existing}
	}
	for _, overload := range candidates {
		forward, function := overload.Type.(*types.FunctionType)
		if !overload.IsForward || !function {
			continue
		}
		if overload.HasOverloadDirective && forwardSignaturesMatch(forward, signature) ||
			!overload.HasOverloadDirective && SignaturesEqual(forward, signature) &&
				forward.ReturnType.Equals(signature.ReturnType) && defaultParametersMatch(forward, signature) {
			return true
		}
	}
	return false
}

// matchingExplicitForward returns the selected declaration before binding
// consumes its forward state, including its original default signature snapshot.
func (st *SymbolTable) matchingExplicitForward(name string, signature *types.FunctionType) *Symbol {
	existing, ok := st.symbols.Get(name)
	if !ok {
		return nil
	}
	candidates := existing.Overloads
	if !existing.IsOverloadSet {
		candidates = []*Symbol{existing}
	}
	for _, overload := range candidates {
		forward, function := overload.Type.(*types.FunctionType)
		if overload.IsForward && overload.HasOverloadDirective && function && forwardSignaturesMatch(forward, signature) {
			return overload
		}
	}
	return nil
}
