package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// contractSource is one class's contribution to a method's contract chain: the
// declaration carrying the conditions, plus the name a failure message must
// report. DWScript names the class that *declares* a condition, not the
// receiver's dynamic class, so an inherited `require` still reports
// "TBase.Check" when the call was dispatched to a TSubChild instance.
type contractSource struct {
	fn          *ast.FunctionDecl
	routineName string
}

// contractChainKey identifies a resolved chain. The declaration alone is not
// enough: the same declaration is shared with every descendant that does not
// override it, and each descendant reaches a different set of ancestors.
type contractChainKey struct {
	decl  *ast.FunctionDecl
	class string
}

// contractChain returns the contract sources that apply to a call of fn, the
// executing declaration first and the base-most ancestor last. It returns nil
// when fn is not a method — a free function, or a local function nested inside
// a method body, which must keep its bare name and its own conditions only.
//
// Chains are resolved once per (declaration, class) pair: the walk is pure
// class-registry structure, which does not change once a program is running.
func (e *Evaluator) contractChain(fn *ast.FunctionDecl, ctx *ExecutionContext) []contractSource {
	if fn == nil || fn.Name == nil {
		return nil
	}

	// The declaring class of the executing method, bound by
	// executeMethodWithClassInfo / executeClassMethodDirect. Empty for every
	// non-method call, which is the common case, so this is also the fast path.
	className := currentMethodClassName(ctx)
	if className == "" {
		return nil
	}

	key := contractChainKey{decl: fn, class: className}
	if chain, ok := e.contractChains[key]; ok {
		return chain
	}

	chain := e.buildContractChain(fn, className)
	if e.contractChains == nil {
		e.contractChains = make(map[contractChainKey][]contractSource)
	}
	e.contractChains[key] = chain
	return chain
}

// buildContractChain walks the class hierarchy from the class declaring fn up
// to the root, collecting each distinct declaration of the same method.
func (e *Evaluator) buildContractChain(fn *ast.FunctionDecl, className string) []contractSource {
	classInfo := e.typeSystem.LookupClass(className)
	if classInfo == nil {
		return nil
	}

	defining := ownerOfMethodDecl(classInfo, fn)
	if defining == nil {
		// No class in the chain declares fn: it is a local function declared
		// inside a method body, executing with the method's class still bound.
		return nil
	}

	chain := []contractSource{{fn: fn, routineName: qualifiedName(defining.GetName(), fn)}}
	seen := map[*ast.FunctionDecl]bool{fn: true}

	for current := defining; current != nil; {
		parent := current.GetParent()
		if parent == nil {
			break
		}
		decl := runtime.MethodDeclaration(parent.LookupMethod(fn.Name.Value))
		if decl == nil || seen[decl] {
			break
		}
		seen[decl] = true

		owner := ownerOfMethodDecl(parent, decl)
		if owner == nil {
			owner = parent
		}
		chain = append(chain, contractSource{fn: decl, routineName: qualifiedName(owner.GetName(), decl)})
		current = owner
	}

	return chain
}

// ownerOfMethodDecl returns the highest class in the hierarchy that declares
// decl, or nil when no class does. Method implementations are pointer-shared
// down to descendants during registration, so the highest owner is the class
// that actually wrote the declaration.
func ownerOfMethodDecl(classInfo runtime.IClassInfo, decl *ast.FunctionDecl) runtime.IClassInfo {
	var owner runtime.IClassInfo
	for current := classInfo; current != nil; current = current.GetParent() {
		asOwner, ok := current.(methodDeclOwner)
		if !ok {
			break
		}
		if asOwner.OwnsMethodDecl(decl) {
			owner = current
		}
	}
	return owner
}

// qualifiedName renders "TClass.Method", the form DWScript uses in contract
// failure messages.
func qualifiedName(className string, fn *ast.FunctionDecl) string {
	if className == "" {
		return fn.Name.Value
	}
	return className + "." + fn.Name.Value
}
