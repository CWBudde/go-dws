package semantic

import "github.com/cwbudde/go-dws/pkg/ast"

// recordDeclaredType connects runtime registration to the same nominal type used
// during analysis. It runs after declaration construction has completed.
func (a *Analyzer) recordDeclaredType(node ast.Node, name string) {
	resolved, err := a.resolveType(name)
	if err != nil || resolved == nil {
		return
	}
	a.semanticInfo.SetResolvedType(node, resolved)
	switch declaration := node.(type) {
	case *ast.ArrayDecl:
		a.semanticInfo.SetResolvedType(declaration.ArrayType, resolved)
	case *ast.TypeDeclaration:
		if declaration.FunctionPointerType != nil {
			a.semanticInfo.SetResolvedType(declaration.FunctionPointerType, resolved)
		}
	}
}
