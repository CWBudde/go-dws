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
		if overload.IsForward && function && SignaturesEqual(forward, signature) &&
			forward.ReturnType.Equals(signature.ReturnType) && defaultParametersMatch(forward, signature) {
			return true
		}
	}
	return false
}
