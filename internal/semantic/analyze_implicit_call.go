package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

// checkImplicitCallArity reports a statement that names a routine without an
// argument list when that routine requires arguments. DWScript reads a bare
// routine name in statement position as a call — `Test;` is `Test()` — so a
// routine with required parameters produces `More arguments expected`, anchored
// at the name.
//
// Only statement position is covered. In an expression the same name may be a
// reference rather than a call (`var f := Test`), and which one it is depends on
// the expected type.
func (a *Analyzer) checkImplicitCallArity(expr ast.Expression) {
	// A failed parse leaves recovered fragments behind whose shape says nothing
	// about the source; upstream stops at the syntax error and never reaches the
	// implicit call. The constant-instruction hint is suppressed for the same
	// reason.
	if expr == nil || a.parseHadErrors {
		return
	}
	switch e := expr.(type) {
	case *ast.Identifier:
		if pos, missing := a.implicitCallNeedsArguments(e); missing {
			if overloadedBuiltins[ident.Normalize(e.Value)] {
				a.addNoOverloadedVersion(a.builtinCanonicalName(e.Value), pos)
				return
			}
			a.addMoreArgumentsExpected(pos)
		}
	case *ast.MemberAccessExpression:
		if pos, missing := a.implicitMemberCallNeedsArguments(e); missing {
			a.addMoreArgumentsExpected(pos)
		}
	}
}

// implicitCallNeedsArguments reports a bare name that resolves to a routine with
// required parameters.
func (a *Analyzer) implicitCallNeedsArguments(identifier *ast.Identifier) (token.Position, bool) {
	pos := identifier.Token.Pos

	if sym, ok := a.symbols.Resolve(identifier.Value); ok {
		// Inside a function's own body the name is an alias for Result, not a
		// recursive call.
		if a.currentFunction != nil && ident.Equal(a.currentFunction.Name.Value, identifier.Value) {
			return pos, false
		}
		// An overload set has no type of its own, so it is answered from its
		// members: the bare name is a call to whichever overload takes none.
		if sym.IsOverloadSet {
			return pos, !overloadSetAcceptsNoArguments(a.symbols.GetOverloadSet(identifier.Value))
		}
		funcType, isFunc := sym.Type.(*types.FunctionType)
		if !isFunc {
			return pos, false
		}
		return pos, requiredParamCount(funcType) > 0
	}

	// A method of the enclosing class, reachable by bare name.
	if a.currentClass != nil {
		if _, found := a.currentClass.GetMethod(identifier.Value); found {
			return pos, !classAcceptsNoArguments(a.currentClass, identifier.Value)
		}
	}

	// A built-in. The registry is the only source that knows the minimum arity;
	// isBuiltinFunction answers a different question and would misreport the
	// names it does not describe.
	if sig, ok := a.builtinRegistry.GetSignature(identifier.Value); ok {
		return pos, sig.MinArgs > 0
	}

	return pos, false
}

// implicitMemberCallNeedsArguments reports `Obj.Method;` where Method requires
// arguments. Array helpers are excluded: they are magic methods with their own
// diagnostic anchor, handled in analyzeArrayMemberAccess.
func (a *Analyzer) implicitMemberCallNeedsArguments(expr *ast.MemberAccessExpression) (token.Position, bool) {
	if expr.Member == nil || expr.Object == nil {
		return token.Position{}, false
	}
	pos := expr.Member.Token.Pos

	objectType := a.inferMemberObjectType(expr.Object)
	if objectType == nil {
		return pos, false
	}
	if _, isArray := types.GetUnderlyingType(objectType).(*types.ArrayType); isArray {
		return pos, false
	}
	classType := metaExprClassType(objectType)
	if classType == nil {
		if underlying, isClass := types.GetUnderlyingType(objectType).(*types.ClassType); isClass {
			classType = underlying
		}
	}
	if classType == nil {
		return pos, false
	}
	if classType.HasConstructor(expr.Member.Value) {
		return pos, false
	}
	if _, found := classType.GetMethod(expr.Member.Value); !found {
		return pos, false
	}
	return pos, !classAcceptsNoArguments(classType, expr.Member.Value)
}

// classAcceptsNoArguments reports whether any method of that name in the class
// hierarchy can be called without arguments. An overload set is spread across
// the hierarchy — a subclass may add `Test(a: Integer)` while the parameterless
// `Test` it overloads lives in a parent — so every level is consulted, not just
// the one GetMethod stops at.
func classAcceptsNoArguments(classType *types.ClassType, name string) bool {
	for class := classType; class != nil; class = class.Parent {
		for _, overload := range class.GetMethodOverloads(name) {
			if overload != nil && requiredParamCount(overload.Signature) == 0 {
				return true
			}
		}
		if signature, found := class.Methods[ident.Normalize(name)]; found &&
			requiredParamCount(signature) == 0 {
			return true
		}
	}
	return false
}

// inferMemberObjectType types a member access's receiver without emitting
// diagnostics for it. The receiver has already been analyzed by the time the
// implicit-call check runs, so re-analyzing it here would duplicate every
// diagnostic it produced.
func (a *Analyzer) inferMemberObjectType(object ast.Expression) types.Type {
	if _, isSelf := object.(*ast.SelfExpression); isSelf {
		if a.currentClass == nil {
			return nil
		}
		if a.inClassMethod {
			return types.NewClassOfType(a.currentClass)
		}
		return a.currentClass
	}
	objIdent, ok := object.(*ast.Identifier)
	if !ok {
		return nil
	}
	if sym, found := a.symbols.Resolve(objIdent.Value); found {
		return sym.Type
	}
	if classType := a.getClassType(objIdent.Value); classType != nil {
		return types.NewClassOfType(classType)
	}
	return nil
}

// builtinCanonicalName returns a built-in's declared spelling, which upstream
// quotes in diagnostics whatever casing the source used.
func (a *Analyzer) builtinCanonicalName(name string) string {
	if a.builtinRegistry != nil {
		if info, ok := a.builtinRegistry.Get(name); ok && info.Name != "" {
			return info.Name
		}
	}
	return name
}

// overloadSetAcceptsNoArguments reports whether any member of an overload set
// can be called without arguments.
func overloadSetAcceptsNoArguments(overloads []*Symbol) bool {
	for _, overload := range overloads {
		funcType, isFunc := overload.Type.(*types.FunctionType)
		if isFunc && requiredParamCount(funcType) == 0 {
			return true
		}
	}
	return false
}
