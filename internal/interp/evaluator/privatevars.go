package evaluator

import (
	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// readPrivateVarBuiltinWins compares overload signatures without executing the
// default expression. Other builtin arbitration uses dynamic argument types,
// but doing so here would discard ReadPrivateVar's lazy-default contract.
func (e *Evaluator) readPrivateVarBuiltinWins(overloads []*ast.FunctionDecl, node *ast.CallExpression, ctx *ExecutionContext) bool {
	if len(node.Arguments) < 1 || len(node.Arguments) > 2 {
		return false
	}
	argTypes, unknownFallback := e.privateVarArgumentTypes(node.Arguments, ctx)
	if argTypes == nil {
		return false
	}
	builtinType := types.NewFunctionTypeWithMetadata(
		[]types.Type{types.STRING, types.VARIANT}, nil, []interface{}{nil, struct{}{}},
		nil, nil, nil, types.VARIANT,
	)
	builtinDistance := types.SignatureDistance(argTypes, builtinType)
	if builtinDistance < 0 {
		return false
	}
	for _, decl := range overloads {
		userType := e.extractMethodType(decl, ctx)
		if userType == nil {
			return false
		}
		if unknownFallback && len(userType.Parameters) > 1 && userType.Parameters[1] != types.VARIANT {
			// Without a semantic fallback type, a more specific user parameter
			// may match. Leave that ambiguous case to ordinary arbitration.
			probe := []types.Type{argTypes[0], userType.Parameters[1]}
			if types.SignatureDistance(probe, userType) >= 0 {
				return false
			}
		}
		if distance := types.SignatureDistance(argTypes, userType); distance >= 0 && distance <= builtinDistance {
			return false
		}
	}
	return true
}

// privateVarArgumentTypes obtains overload types without evaluating the default.
// A missing name type prevents arbitration; an unknown fallback accepts Variant.
func (e *Evaluator) privateVarArgumentTypes(args []ast.Expression, ctx *ExecutionContext) ([]types.Type, bool) {
	argTypes := make([]types.Type, len(args))
	unknownFallback := false
	for idx, arg := range args {
		argTypes[idx] = e.resolvedExpressionType(arg, ctx)
		if fn, ok := argTypes[idx].(*types.FunctionType); ok {
			argTypes[idx] = fn.ReturnType // Parameterless routines auto-invoke in value context.
		}
		if argTypes[idx] == nil {
			if idx == 0 {
				return nil, false
			}
			argTypes[idx] = types.VARIANT
			unknownFallback = true
		}
	}
	return argTypes, unknownFallback
}

// builtinReadPrivateVar evaluates the fallback only when the private value is
// absent. Ordinary registry dispatch would evaluate it before looking up the key.
func (e *Evaluator) builtinReadPrivateVar(node *ast.CallExpression, name *ast.Identifier, ctx *ExecutionContext) Value {
	if len(node.Arguments) < 1 || len(node.Arguments) > 2 {
		return e.newError(node, "ReadPrivateVar() expects 1 or 2 arguments, got %d", len(node.Arguments))
	}
	unit := ctx.CurrentUnit()
	if unit == "" {
		ctx.SetException(e.createException("Exception", "Private variables cannot be referred from main module", nil, ctx))
		return &runtime.NilValue{}
	}
	key := e.evalValueContextExpression(node.Arguments[0], ctx)
	if isError(key) || ctx.Exception() != nil {
		return key
	}
	args := []Value{key}
	if errVal := e.coerceBuiltinArgsToSignature(name, node.Arguments[:1], args, ctx); errVal != nil {
		return errVal
	}
	if ctx.Exception() != nil {
		return &runtime.NilValue{}
	}
	keyString, ok := unwrapVariant(args[0]).(*runtime.StringValue)
	if !ok {
		return e.newError(node, "ReadPrivateVar() expects a String name, got %s", key.Type())
	}
	stored, found := builtins.DefaultPrivateVars.Read(unit, keyString.Value)
	if found || len(node.Arguments) == 1 {
		return stored.ToRuntimeValue()
	}
	value := e.evalValueContextExpression(node.Arguments[1], ctx)
	if isError(value) || ctx.Exception() != nil {
		return value
	}
	// Preserve the ordinary builtin boundary's record-to-Variant conversion
	// for the default expression, without evaluating it on a successful read.
	args = append(args, value)
	if errVal := e.coerceBuiltinArgsToSignature(name, node.Arguments, args, ctx); errVal != nil {
		return errVal
	}
	if ctx.Exception() != nil {
		return &runtime.NilValue{}
	}
	return unwrapVariant(args[1])
}
