package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// selectHelperCallOverload separates candidate selection from argument checking.
// Even one marked declaration is an overload set. Recoverable child errors are
// read before reporting the enclosing no-match at the member, without falling
// back to another helper or to an inaccessible-member diagnostic.
func (a *Analyzer) selectHelperCallOverload(candidates []*types.FunctionType, marked bool, args []ast.Expression, pos token.Position, name string) int {
	if len(candidates) == 1 && !marked {
		return 0
	}
	argTypes := make([]types.Type, len(args))
	failed := false
	for i, arg := range args {
		if len(candidates) == 1 && i < len(candidates[0].Parameters) {
			signature := candidates[0]
			strict := i < len(signature.StrictParams) && signature.StrictParams[i]
			argTypes[i] = a.analyzeArgumentForParameter(arg, signature.Parameters[i], strict)
		} else {
			argTypes[i] = a.analyzeOverloadArgument(arg)
		}
		failed = failed || argTypes[i] == nil
	}
	signatures := make([]types.Type, len(candidates))
	for i, candidate := range candidates {
		signatures[i] = candidate
	}
	selected := -1
	if !failed {
		if index, err := types.ResolveOverload(signatures, argTypes); err == nil {
			selected = index
		}
	}
	if selected < 0 {
		diagnostic := NewNoOverloadMatchError(pos, name)
		diagnostic.AfterChildren = true
		a.addStructuredError(diagnostic)
		return -1
	}
	return selected
}

// analyzeBoundHelperCall uses the already-visible lexical overload set, rather
// than reopening helper lookup and exposing later declarations or other owners.
func (a *Analyzer) analyzeBoundHelperCall(member *ast.Identifier, args []ast.Expression, symbol *Symbol) (types.Type, bool) {
	if a.currentHelperType == nil {
		return nil, false
	}
	symbols := []*Symbol{symbol}
	if symbol.IsOverloadSet {
		symbols = a.symbols.GetOverloadSet(member.Value)
	}
	candidates := make([]*types.FunctionType, len(symbols))
	marked := false
	for i, candidate := range symbols {
		signature, ok := candidate.Type.(*types.FunctionType)
		if !ok || !a.isBoundHelperMethod(member.Value, signature) {
			return nil, false
		}
		candidates[i] = signature
		marked = marked || candidate.HasOverloadDirective
	}
	selected := a.selectHelperCallOverload(candidates, marked, args, member.Token.Pos, symbol.Name)
	if selected < 0 {
		return nil, true
	}
	signature := candidates[selected]
	a.analyzeHelperCallArguments(a.currentHelperType.TargetType, member.Value, signature, args, member.Token.Pos)
	return signature.ReturnType, true
}
