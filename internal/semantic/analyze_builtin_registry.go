package semantic

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// analyzeRegisteredBuiltin checks ordinary calls from the shared registry.
// Intrinsics and builtins with specialized semantic rules are handled by
// analyzeBuiltinFunction before reaching this path. Diagnostic styles preserve
// legacy wording without duplicating signature validation.
func (a *Analyzer) analyzeRegisteredBuiltin(name string, args []ast.Expression, call *ast.CallExpression) (types.Type, bool) {
	info, ok := a.builtinRegistry.Get(name)
	if !ok || info.Signature == nil {
		return nil, false
	}
	sig := info.Signature
	style := builtinDiagnosticStyles[ident.Normalize(name)]
	diagnosticName := info.Name
	if style.name != "" {
		diagnosticName = style.name
	}
	result := sig.ReturnType
	if result == nil {
		result = types.VOID
	}
	if len(args) < sig.MinArgs || sig.MaxArgs >= 0 && len(args) > sig.MaxArgs {
		var expected string
		switch {
		case sig.MinArgs == sig.MaxArgs:
			expected = fmt.Sprintf("%d arguments", sig.MinArgs)
			if sig.MinArgs == 1 {
				expected = "1 argument"
			} else if sig.MinArgs == 0 && style.noArguments {
				expected = "no arguments"
			}
		case sig.MaxArgs < 0:
			expected = fmt.Sprintf("at least %d arguments", sig.MinArgs)
		default:
			expected = fmt.Sprintf("%d to %d arguments", sig.MinArgs, sig.MaxArgs)
			if style.optionalOr {
				expected = fmt.Sprintf("%d or %d arguments", sig.MinArgs, sig.MaxArgs)
			}
		}
		a.addError("function '%s' expects %s, got %d at %s", diagnosticName, expected, len(args), call.Token.Pos.String())
		return result, true
	}
	for index, arg := range args {
		actual := a.analyzeExpression(arg)
		if actual == nil || len(sig.ParamTypes) == 0 {
			continue
		}
		paramIndex := index
		if paramIndex >= len(sig.ParamTypes) {
			if !sig.IsVariadic {
				continue
			}
			paramIndex = len(sig.ParamTypes) - 1
		}
		expected := sig.ParamTypes[paramIndex]
		if expected == nil || expected == types.VARIANT {
			continue
		}
		valid := expected.Equals(actual)
		description := expected.String()
		if expected == types.FLOAT {
			valid = types.INTEGER.Equals(actual) || types.FLOAT.Equals(actual)
			description = "Integer or Float"
		}
		if !valid {
			position := "argument"
			if len(sig.ParamTypes) > 1 {
				position = fmt.Sprintf("argument %d", index+1)
			}
			expectation := description + " as " + position
			if paramIndex < len(style.arguments) && style.arguments[paramIndex] != "" {
				expectation = style.arguments[paramIndex]
			}
			a.addError("function '%s' expects %s, got %s at %s", diagnosticName, expectation, actual.String(), call.Token.Pos.String())
		}
	}
	return result, true
}

// parameterlessBuiltinType returns the result type of a builtin that takes no
// arguments at all, so a bare identifier such as `Random` or `Now` types as an
// implicit call (`Random*0`, `var t := Now`) instead of falling back to VOID.
//
// Only strictly parameterless functions qualify: a signature with optional or
// variadic parameters says nothing about whether the bare name means a call or
// a reference, and procedures (nil ReturnType) stay VOID as before.
func (a *Analyzer) parameterlessBuiltinType(name string) (types.Type, bool) {
	sig, ok := a.builtinRegistry.GetSignature(name)
	if !ok || sig.ReturnType == nil {
		return nil, false
	}
	if sig.IsVariadic || sig.MinArgs != 0 || sig.MaxArgs != 0 {
		return nil, false
	}
	return sig.ReturnType, true
}
