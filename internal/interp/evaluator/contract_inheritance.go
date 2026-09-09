package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// contractSource is one class's contribution to a method's contract chain: the
// declaration carrying the conditions, plus the name a failure message must
// report. DWScript names the class that *declares* a condition, not the
// receiver's dynamic class, so an inherited `require` still reports
// "TBase.Check" when the call was dispatched to a TSubChild instance.
type contractSource struct {
	fn          *ast.FunctionDecl
	className   string
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

	chain := []contractSource{newContractSource(fn, defining.GetName())}
	seen := map[*ast.FunctionDecl]bool{fn: true}

	for current := defining; current != nil; {
		parent := current.GetParent()
		if parent == nil {
			break
		}
		decl := ancestorMethodDecl(parent, fn, seen)
		if decl == nil {
			break
		}
		seen[decl] = true

		owner := ownerOfMethodDecl(parent, decl)
		if owner == nil {
			owner = parent
		}
		chain = append(chain, newContractSource(decl, owner.GetName()))
		current = owner
	}

	return chain
}

// ancestorMethodDecl finds the declaration fn overrides, starting at parent.
// A name is not enough to identify it: an ancestor may declare several
// overloads, and inheriting a sibling overload's condition would evaluate it
// against arguments it was never written for. Class methods live in their own
// table, which the instance lookup never searches.
func ancestorMethodDecl(
	parent runtime.IClassInfo,
	fn *ast.FunctionDecl,
	seen map[*ast.FunctionDecl]bool,
) *ast.FunctionDecl {
	overloads := parent.GetMethodOverloads(fn.Name.Value)
	if fn.IsClassMethod {
		overloads = parent.GetClassMethodOverloads(fn.Name.Value)
	}
	if decl := bestSignatureMatch(overloads, fn, seen); decl != nil {
		return decl
	}

	// A method that is not overloaded is registered in the name-indexed table
	// only, so the overload set above is empty for it.
	single := parent.LookupMethod(fn.Name.Value)
	if fn.IsClassMethod {
		single = parent.LookupClassMethod(fn.Name.Value)
	}
	decl := runtime.MethodDeclaration(single)
	if decl == nil || seen[decl] {
		return nil
	}
	return decl
}

// bestSignatureMatch picks the overload fn overrides: same parameter types when
// one matches, otherwise the first with the same arity. Returns nil when no
// candidate has fn's arity, which is safer than inheriting a mismatched
// contract.
func bestSignatureMatch(
	candidates []*runtime.MethodMetadata,
	fn *ast.FunctionDecl,
	seen map[*ast.FunctionDecl]bool,
) *ast.FunctionDecl {
	var byArity *ast.FunctionDecl
	for _, candidate := range candidates {
		decl := runtime.MethodDeclaration(candidate)
		if decl == nil || seen[decl] || len(decl.Parameters) != len(fn.Parameters) {
			continue
		}
		if parameterTypesMatch(decl, fn) {
			return decl
		}
		if byArity == nil {
			byArity = decl
		}
	}
	return byArity
}

// parameterTypesMatch compares two declarations' parameter types by their
// written form, which is what distinguishes overloads of the same arity.
func parameterTypesMatch(candidate, fn *ast.FunctionDecl) bool {
	for idx, param := range candidate.Parameters {
		own := fn.Parameters[idx]
		if param.Type == nil || own.Type == nil {
			if (param.Type == nil) != (own.Type == nil) {
				return false
			}
			continue
		}
		if !ident.Equal(param.Type.String(), own.Type.String()) {
			return false
		}
	}
	return true
}

// newContractSource records a declaration together with the class that declares
// it: the name a failure reports, and the scope its conditions must resolve in.
func newContractSource(fn *ast.FunctionDecl, className string) contractSource {
	return contractSource{fn: fn, className: className, routineName: qualifiedName(className, fn)}
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

// preconditionSources returns the sources whose preconditions must run, base-most
// first. Upstream DWScript allows `require` on the root method only, so in
// practice at most one source contributes; ordering it root-first keeps the
// behavior defined if a script declares more.
func preconditionSources(chain []contractSource, fn *ast.FunctionDecl) []contractSource {
	if len(chain) == 0 {
		return ownConditionSource(fn, fn.PreConditions != nil)
	}

	sources := make([]contractSource, 0, len(chain))
	for i := len(chain) - 1; i >= 0; i-- {
		if chain[i].fn.PreConditions != nil {
			sources = append(sources, chain[i])
		}
	}
	return sources
}

// postconditionSources returns the sources whose postconditions must run, the
// executing declaration first and ancestors after. The order is observable:
// when a derived and an inherited postcondition both fail, DWScript reports the
// derived one (SimpleScripts/method_contracts).
func postconditionSources(chain []contractSource, fn *ast.FunctionDecl) []contractSource {
	if len(chain) == 0 {
		return ownConditionSource(fn, fn.PostConditions != nil)
	}

	sources := make([]contractSource, 0, len(chain))
	for _, source := range chain {
		if source.fn.PostConditions != nil {
			sources = append(sources, source)
		}
	}
	return sources
}

// ownConditionSource is the non-method fallback: the declaration's own
// conditions under its own (possibly out-of-line qualified) name.
func ownConditionSource(fn *ast.FunctionDecl, has bool) []contractSource {
	if !has {
		return nil
	}
	return []contractSource{{fn: fn, routineName: contractFuncName(fn)}}
}

// parameterAliases maps an ancestor declaration's parameter names to the values
// currently bound under the executing declaration's names. DWScript matches
// contract parameters by position, so an ancestor condition whose parameters are
// named differently from the override's must still see the call's arguments.
// Names that already agree are omitted, so the common case allocates nothing.
func parameterAliases(source, executing *ast.FunctionDecl, ctx *ExecutionContext) map[string]Value {
	if source == executing {
		return nil
	}

	var aliases map[string]Value
	for idx, param := range source.Parameters {
		if idx >= len(executing.Parameters) {
			break
		}
		own := executing.Parameters[idx]
		if param.Name == nil || own.Name == nil || ident.Equal(param.Name.Value, own.Name.Value) {
			continue
		}
		value, ok := ctx.Env().Get(own.Name.Value)
		if !ok {
			continue
		}
		if aliases == nil {
			aliases = make(map[string]Value, len(source.Parameters))
		}
		aliases[param.Name.Value] = value
	}
	return aliases
}

// evalInSourceScope runs body in the scope the source's conditions were written
// in: its declaring class, and its own parameter names bound to the executing
// call's argument values. For the executing declaration itself both already
// hold, so it runs body directly.
//
// The class matters because bare and Self member access resolve against the
// static class of the method being executed (see staticClassNameOf). Leaving
// the derived class bound would make an inherited condition read a field the
// derived class shadows -- the wrong storage slot -- rather than the one its
// own class declared.
func (e *Evaluator) evalInSourceScope(
	source contractSource,
	executing *ast.FunctionDecl,
	ctx *ExecutionContext,
	body func() Value,
) Value {
	if source.fn == executing {
		return body()
	}

	aliases := parameterAliases(source.fn, executing, ctx)
	rebindClass := source.className != "" && !ident.Equal(source.className, currentMethodClassName(ctx))
	if len(aliases) == 0 && !rebindClass {
		return body()
	}

	ctx.PushEnv()
	defer ctx.PopEnv()
	if rebindClass {
		ctx.Env().Define("__CurrentMethodClass__", &runtime.StringValue{Value: source.className})
	}
	for name, value := range aliases {
		ctx.Env().Define(name, value)
	}
	return body()
}

// checkContractPreconditions evaluates each source's preconditions in turn,
// stopping at the first failure (which has already raised the exception).
func (e *Evaluator) checkContractPreconditions(
	sources []contractSource,
	executing *ast.FunctionDecl,
	ctx *ExecutionContext,
) Value {
	for _, source := range sources {
		result := e.evalInSourceScope(source, executing, ctx, func() Value {
			return e.checkPreconditions(source.routineName, source.fn.PreConditions, ctx)
		})
		if isError(result) {
			return result
		}
		if ctx.Exception() != nil {
			return nil
		}
	}
	return nil
}

// checkContractPostconditions mirrors checkContractPreconditions for `ensure`.
func (e *Evaluator) checkContractPostconditions(
	sources []contractSource,
	executing *ast.FunctionDecl,
	ctx *ExecutionContext,
) Value {
	for _, source := range sources {
		result := e.evalInSourceScope(source, executing, ctx, func() Value {
			return e.checkPostconditions(source.routineName, source.fn.PostConditions, ctx)
		})
		if isError(result) {
			return result
		}
		if ctx.Exception() != nil {
			return nil
		}
	}
	return nil
}

// captureOldValuesForSources captures the `old` operands of every postcondition
// that will run, including inherited ones, before the body executes.
func (e *Evaluator) captureOldValuesForSources(
	sources []contractSource,
	executing *ast.FunctionDecl,
	ctx *ExecutionContext,
) map[string]Value {
	oldValues := make(map[string]Value)
	for _, source := range sources {
		e.evalInSourceScope(source, executing, ctx, func() Value {
			for _, condition := range source.fn.PostConditions.Conditions {
				e.findOldExpressions(condition.Test, ctx, oldValues)
				if condition.Message != nil {
					e.findOldExpressions(condition.Message, ctx, oldValues)
				}
			}
			return nil
		})
	}
	return oldValues
}
