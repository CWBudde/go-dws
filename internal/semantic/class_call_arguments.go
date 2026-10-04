package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// selectMemberCallOverload keeps overload directives meaningful even when only
// one declaration is visible. Unmarked single signatures defer to type checking.
func (a *Analyzer) selectMemberCallOverload(overloads []*types.MethodInfo, args []ast.Expression, name string, pos token.Position) *types.MethodInfo {
	if len(overloads) == 1 && !overloads[0].HasOverloadDirective {
		return overloads[0]
	}
	argTypes := make([]types.Type, len(args))
	failed := false
	for i, arg := range args {
		argTypes[i] = a.analyzeOverloadArgument(arg)
		if argTypes[i] == nil {
			failed = true
		}
	}
	if failed {
		return nil
	}
	candidates := make([]*Symbol, len(overloads))
	for i, method := range overloads {
		candidates[i] = &Symbol{Type: method.Signature}
	}
	selected, err := ResolveOverload(candidates, argTypes)
	if err != nil {
		diagnostic := NewNoOverloadMatchError(pos, name)
		diagnostic.AfterChildren = true
		a.addStructuredError(diagnostic)
		return nil
	}
	for i, candidate := range candidates {
		if candidate == selected {
			return overloads[i]
		}
	}
	return nil
}

// analyzeClassCallArguments reads every supplied argument before validating the
// signature. Upstream records child errors while reading arguments, then checks
// types and finally counts; only errors from that type check suppress arity.
func (a *Analyzer) analyzeClassCallArguments(signature *types.FunctionType, args []ast.Expression, pos token.Position) {
	a.analyzeMemberCallArguments(signature, args, pos, false)
}

// analyzeMemberCallArguments also handles record methods, whose implicit Self
// occupies argument 0 upstream. It shifts both diagnostic indices and the lookup
// into written argument positions; the final argument falls back to the method.
func (a *Analyzer) analyzeMemberCallArguments(signature *types.FunctionType, args []ast.Expression, pos token.Position, hasSelfArgument bool) {
	argTypes := make([]types.Type, len(args))
	failed := make([]bool, len(args))
	for i, arg := range args {
		mark := len(a.errors)
		if i < len(signature.Parameters) {
			strict := i < len(signature.StrictParams) && signature.StrictParams[i]
			argTypes[i] = a.analyzeArgumentForParameter(arg, signature.Parameters[i], strict)
		} else {
			argTypes[i] = a.analyzeExpression(arg)
		}
		failed[i] = argTypes[i] == nil || a.errorsSince(mark)
	}

	mark := len(a.errors)
	for i, argType := range argTypes {
		if i >= len(signature.Parameters) {
			break
		}
		strict := i < len(signature.StrictParams) && signature.StrictParams[i]
		if !failed[i] && !a.argumentMatchesParameter(argType, signature.Parameters[i], strict) {
			index, argPos := i, args[i].Pos()
			if hasSelfArgument {
				index++
				argPos = pos
				if index < len(args) {
					argPos = args[index].Pos()
				}
			}
			a.addArgumentTypeError(index, semanticTypeNameForDiagnostic(signature.Parameters[i]), argType, argPos)
		}
	}
	if !a.errorsSince(mark) {
		a.addMemberCallCountError(signature, len(args), pos)
	}
}

// addMemberCallCountError is for ordinary named members. Unlike specialized
// array helpers, upstream uses Too many arguments even for parameterless methods.
func (a *Analyzer) addMemberCallCountError(signature *types.FunctionType, count int, pos token.Position) {
	var message string
	switch {
	case count < requiredParamCount(signature):
		message = msgMoreArgumentsExpected
	case count > len(signature.Parameters):
		message = msgTooManyArguments
	default:
		return
	}
	err := NewGenericError(pos, message)
	err.Type = ErrorArgumentCount
	err.AfterChildren = true
	a.addStructuredError(err)
}
