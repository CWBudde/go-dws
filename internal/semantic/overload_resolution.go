package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// SignaturesEqual compares signatures using the shared type-system rules.
func SignaturesEqual(left, right *types.FunctionType) bool {
	return types.SignaturesEqual(left, right)
}

// ResolveOverload selects a semantic symbol using the shared type-system rules.
func ResolveOverload(candidates []*Symbol, argTypes []types.Type) (*Symbol, error) {
	signatures := make([]types.Type, len(candidates))
	for index, candidate := range candidates {
		signatures[index] = candidate.Type
	}
	index, err := types.ResolveOverload(signatures, argTypes)
	if err != nil {
		return nil, err
	}
	return candidates[index], nil
}

// declaredMethodName returns the original-cased declared name of a method or
// constructor in the class hierarchy, or "" if unknown.
func (a *Analyzer) declaredMethodName(classType *types.ClassType, name string) string {
	normalized := ident.Normalize(name)
	for current := classType; current != nil; current = current.Parent {
		if declared, ok := current.MethodDeclNames[normalized]; ok {
			return declared
		}
	}
	return ""
}

// analyzeOverloadArgument analyzes a call argument for overload resolution.
// An empty array literal carries no inherent element type; it gets an
// "array of Variant" placeholder so it can match any array-typed parameter
// (the winning overload's parameter type then applies at the call).
func (a *Analyzer) analyzeOverloadArgument(arg ast.Expression) types.Type {
	if lit, ok := arg.(*ast.ArrayLiteralExpression); ok && len(lit.Elements) == 0 {
		return types.NewDynamicArrayType(types.VARIANT)
	}
	return a.analyzeExpression(arg)
}

// requiredParamCount returns the number of parameters without default values,
// i.e. the minimum number of arguments a call must supply.
func requiredParamCount(sig *types.FunctionType) int {
	if len(sig.DefaultValues) != len(sig.Parameters) {
		return len(sig.Parameters)
	}
	required := 0
	for _, def := range sig.DefaultValues {
		if def == nil {
			required++
		}
	}
	return required
}

// mergeDefaultValues copies parameter default values from a declaration signature
// into an implementation signature for parameters where the implementation did not
// respecify them. DWScript allows (and expects) implementations to omit defaults
// that were declared in the class/interface declaration.
func mergeDefaultValues(impl, decl *types.FunctionType) {
	if decl == nil || impl == nil || len(decl.DefaultValues) != len(decl.Parameters) {
		return
	}
	if len(impl.DefaultValues) != len(impl.Parameters) || len(impl.Parameters) != len(decl.Parameters) {
		return
	}
	for i, def := range decl.DefaultValues {
		if impl.DefaultValues[i] == nil {
			impl.DefaultValues[i] = def
		}
	}
}
