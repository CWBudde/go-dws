package semantic

import (
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
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
	a.annotateHelperCall(a.currentHelperType, member, signature)
	a.analyzeHelperCallArguments(a.currentHelperType.TargetType, member.Value, signature, args, member.Token.Pos)
	return signature.ReturnType, true
}

// helperCallOwner keeps the first declaring user helper within a target tier,
// allowing a descendant to hide its ancestor. Alias/metatype tiers retain their
// existing specificity order; builtin declarations remain fallback.
func (a *Analyzer) helperCallOwner(receiver types.Type, name string) *types.HelperType {
	helpers := a.getHelpersForType(receiver)
	key := ident.Normalize(name)
	_, meta := recordReceiverType(receiver)
	declares := func(helper *types.HelperType) bool {
		if meta {
			return len(helper.ClassMethodOverloads[key]) > 0
		}
		return len(helperOwnOverloads(helper, key)) > 0
	}
	for i := len(helpers) - 1; i >= 0; i-- {
		tier := helpers[i]
		if !declares(tier) {
			continue
		}
		if tier.TargetType == nil {
			// Generic intrinsic helpers have no concrete target declaration.
			return tier
		}
		var winner *types.HelperType
		for _, helper := range helpers {
			if helper.TargetType == nil || !ident.Equal(helper.TargetType.String(), tier.TargetType.String()) || !declares(helper) || helper.BuiltinMethods[key] != "" {
				continue
			}
			if winner == nil || helperInherits(helper, winner) {
				winner = helper
			}
		}
		if winner != nil {
			return winner
		}
		return tier
	}
	return nil
}

func (a *Analyzer) declaredHelperCallName(receiver types.Type, name string) string {
	if owner := a.helperCallOwner(receiver, name); owner != nil {
		return owner.MethodDeclNames[ident.Normalize(name)]
	}
	return ""
}

func (a *Analyzer) isHelperCallClassMethod(receiver types.Type, name string) bool {
	owner := a.helperCallOwner(receiver, name)
	return owner != nil && owner.ClassMethods[ident.Normalize(name)]
}

// annotateHelperCall records the declaring helper and its original overload
// slot, which runtime implementation registration replaces in place. Function
// helpers use slot zero. Builtin calls retain their specialized dispatch.
func (a *Analyzer) annotateHelperCall(receiver types.Type, member *ast.Identifier, signature *types.FunctionType) {
	owner, declaration := a.helperSignatureDeclaration(receiver, member.Value, signature)
	if declaration == nil {
		return
	}
	annotation := strings.Replace(helperMemberAnnotation(owner, member.Value, signature), "__helper_member:", "__helper_call:", 1)
	a.semanticInfo.SetType(member, &ast.TypeAnnotation{Token: member.Token, Name: annotation})
	a.semanticInfo.SetResolvedType(member, owner)
}

func (a *Analyzer) annotateExplicitHelperCall(member *ast.Identifier, signature explicitHelperSignature) {
	a.semanticInfo.SetType(member, &ast.TypeAnnotation{Token: member.Token,
		Name: "__helper_call:" + signature.owner + ":" + strconv.Itoa(signature.index)})
	a.semanticInfo.SetResolvedType(member, a.getHelperType(signature.owner))
}

func (a *Analyzer) hasHelperCallBinding(member *ast.Identifier) bool {
	annotation := a.semanticInfo.GetType(member)
	return annotation != nil && strings.HasPrefix(annotation.Name, "__helper_call:")
}
