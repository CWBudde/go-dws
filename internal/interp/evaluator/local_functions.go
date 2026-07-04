// Package evaluator: scoped (nested) function declarations.
//
// DWScript allows functions to be declared inside another function's body.
// Such nested functions are local to the enclosing scope: they hide any
// same-named outer functions or methods for the duration of the enclosing
// call and must not leak into the global function registry.
package evaluator

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// localFunctionSetPrefix keys env entries holding nested function overload sets.
const localFunctionSetPrefix = "__localfns_"

// LocalFunctionSet is an environment-scoped set of nested function overloads.
// It lives in the enclosing function's environment, so it disappears when the
// enclosing call returns.
type LocalFunctionSet struct {
	Name  string
	Decls []*ast.FunctionDecl
}

// Type implements Value.
func (s *LocalFunctionSet) Type() string { return "LOCAL_FUNCTION_SET" }

// String implements Value.
func (s *LocalFunctionSet) String() string {
	names := make([]string, len(s.Decls))
	for i := range s.Decls {
		names[i] = s.Name
	}
	return "local function " + strings.Join(names, ", ")
}

// localFunctionSetKey returns the env key for a nested function name.
func localFunctionSetKey(name string) string {
	return localFunctionSetPrefix + ident.Normalize(name)
}

// defineLocalFunction records a nested function declaration in the current
// environment, appending to (or replacing within) the local overload set.
func (e *Evaluator) defineLocalFunction(node *ast.FunctionDecl, ctx *ExecutionContext) {
	key := localFunctionSetKey(node.Name.Value)
	// Only extend a set declared in the SAME scope (GetLocal): sibling nested
	// overloads share one set, while a set owned by an outer function must be
	// hidden by a fresh local set, not appended to.
	if existingRaw, ok := ctx.Env().GetLocal(key); ok {
		if set, ok := existingRaw.(*LocalFunctionSet); ok {
			for i, decl := range set.Decls {
				if parametersMatchLocal(decl.Parameters, node.Parameters) {
					set.Decls[i] = node
					return
				}
			}
			set.Decls = append(set.Decls, node)
			return
		}
	}
	ctx.Env().Define(key, &LocalFunctionSet{Name: node.Name.Value, Decls: []*ast.FunctionDecl{node}})
}

// lookupLocalFunctions returns the nested function overload set visible from
// the current environment, or nil.
func (e *Evaluator) lookupLocalFunctions(name string, ctx *ExecutionContext) *LocalFunctionSet {
	if raw, ok := ctx.Env().Get(localFunctionSetKey(name)); ok {
		if set, ok := raw.(*LocalFunctionSet); ok {
			return set
		}
	}
	return nil
}

// parametersMatchLocal reports whether two parameter lists have identical
// declared types (used to replace a redeclared nested function).
func parametersMatchLocal(a, b []*ast.Parameter) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		aType, bType := "", ""
		if a[i].Type != nil {
			aType = ident.Normalize(a[i].Type.String())
		}
		if b[i].Type != nil {
			bType = ident.Normalize(b[i].Type.String())
		}
		if aType != bType {
			return false
		}
	}
	return true
}

// maybeCallBuiltinOverload arbitrates between a user overload set and a
// same-named builtin function. In DWScript, declaring `function IntToStr(...);
// overload;` extends the builtin's overload set rather than replacing it: the
// builtin still wins when its signature matches the arguments strictly better.
//
// Returns (result, true) when the call was fully handled here (either builtin
// or user overload executed), or (nil, false) to fall through to the regular
// user-function path.
func (e *Evaluator) maybeCallBuiltinOverload(funcName string, overloads []*ast.FunctionDecl, node *ast.CallExpression, ctx *ExecutionContext) (Value, bool) {
	info, ok := builtins.DefaultRegistry.Get(funcName)
	if !ok || info == nil || info.Signature == nil {
		return nil, false
	}
	// Only arbitrate when every user overload opted into overloading and no
	// overload needs unevaluated arguments (lazy/var).
	for _, fn := range overloads {
		if !fn.IsOverload {
			return nil, false
		}
		for _, p := range fn.Parameters {
			if p.IsLazy || p.ByRef {
				return nil, false
			}
		}
	}

	sig := info.Signature
	if sig.IsVariadic || sig.MaxArgs < 0 || sig.MaxArgs != len(sig.ParamTypes) {
		return nil, false
	}

	// Evaluate arguments once and derive their runtime types.
	args := make([]Value, len(node.Arguments))
	argTypes := make([]types.Type, len(node.Arguments))
	for i, arg := range node.Arguments {
		val := e.Eval(arg, ctx)
		if isError(val) {
			return val, true
		}
		args[i] = val
		argTypes[i] = e.runtimeValueType(val)
	}

	// Distance to the builtin's signature (params beyond MinArgs are optional).
	defaults := make([]interface{}, len(sig.ParamTypes))
	for i := range defaults {
		if i >= sig.MinArgs {
			defaults[i] = struct{}{} // non-nil marks the parameter optional
		}
	}
	builtinType := types.NewFunctionTypeWithMetadata(
		sig.ParamTypes, make([]string, len(sig.ParamTypes)), defaults,
		make([]bool, len(sig.ParamTypes)), make([]bool, len(sig.ParamTypes)), make([]bool, len(sig.ParamTypes)),
		sig.ReturnType,
	)
	builtinDist := semantic.SignatureDistance(argTypes, builtinType)

	// Best user overload distance.
	bestUserDist := -1
	var bestUser *ast.FunctionDecl
	for _, fn := range overloads {
		fnType := e.extractMethodType(fn)
		if fnType == nil {
			continue
		}
		if dist := semantic.SignatureDistance(argTypes, fnType); dist >= 0 && (bestUserDist < 0 || dist < bestUserDist) {
			bestUserDist = dist
			bestUser = fn
		}
	}

	// Builtin wins only on a strictly better match.
	if builtinDist >= 0 && (bestUser == nil || builtinDist < bestUserDist) {
		return info.Function(e, args), true
	}
	if bestUser != nil {
		prepared, err := e.PrepareUserFunctionArgs(bestUser, node.Arguments, args, ctx, node)
		if err != nil {
			return e.newError(node, "%s", err.Error()), true
		}
		return e.ExecuteUserFunctionDirect(bestUser, prepared, ctx), true
	}
	return nil, false
}

// userMethodHidesBuiltin reports whether the receiver's class declares a
// method with the given name that can be invoked with zero arguments,
// hiding the same-named builtin TObject member (e.g. ClassName).
func (e *Evaluator) userMethodHidesBuiltin(obj Value, memberName string) bool {
	objInst, ok := obj.(*runtime.ObjectInstance)
	if !ok || objInst.Class == nil {
		return false
	}
	for _, decl := range objInst.Class.GetMethodOverloads(memberName) {
		if astCallableWithNoArgs(decl) {
			return true
		}
	}
	for _, decl := range objInst.Class.GetClassMethodOverloads(memberName) {
		if astCallableWithNoArgs(decl) {
			return true
		}
	}
	return false
}

// astCallableWithNoArgs reports whether a declaration can be called without
// arguments (no parameters, or all parameters defaulted).
func astCallableWithNoArgs(decl *ast.FunctionDecl) bool {
	for _, param := range decl.Parameters {
		if param.DefaultValue == nil {
			return false
		}
	}
	return true
}

// callLocalFunctionSet resolves and invokes a nested function overload set.
func (e *Evaluator) callLocalFunctionSet(set *LocalFunctionSet, argExprs []ast.Expression, node ast.Node, ctx *ExecutionContext) Value {
	var fn *ast.FunctionDecl
	var cachedArgs []Value
	var err error

	if len(set.Decls) == 1 {
		fn = set.Decls[0]
		cachedArgs, err = e.ResolveOverloadFast(fn, argExprs, ctx)
	} else {
		fn, cachedArgs, err = e.ResolveOverloadMultiple(ident.Normalize(set.Name), set.Decls, argExprs, ctx)
	}
	if err != nil {
		return e.newError(node, "%s", err.Error())
	}

	args, err := e.PrepareUserFunctionArgs(fn, argExprs, cachedArgs, ctx, node)
	if err != nil {
		return e.newError(node, "%s", err.Error())
	}
	return e.ExecuteUserFunctionDirect(fn, args, ctx)
}
