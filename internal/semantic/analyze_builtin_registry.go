package semantic

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// analyzeRegisteredBuiltin checks ordinary calls from the shared registry.
// Intrinsics and builtins with specialized diagnostics are handled by
// analyzeBuiltinFunction before reaching this path.
func (a *Analyzer) analyzeRegisteredBuiltin(name string, args []ast.Expression, call *ast.CallExpression) (types.Type, bool) {
	info, ok := a.builtinRegistry.Get(name)
	if !ok || info.Signature == nil {
		return nil, false
	}
	sig := info.Signature
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
			}
		case sig.MaxArgs < 0:
			expected = fmt.Sprintf("at least %d arguments", sig.MinArgs)
		default:
			expected = fmt.Sprintf("%d to %d arguments", sig.MinArgs, sig.MaxArgs)
		}
		a.addError("function '%s' expects %s, got %d at %s", info.Name, expected, len(args), call.Token.Pos.String())
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
			a.addError("function '%s' expects %s as %s, got %s at %s", info.Name, description, position, actual.String(), call.Token.Pos.String())
		}
	}
	return result, true
}
