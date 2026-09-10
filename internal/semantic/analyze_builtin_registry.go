package semantic

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// analyzeRegisteredBuiltin checks ordinary calls against shared signatures.
// Diagnostic policies preserve expression-error order and recovery behavior.
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
	if result == nil && !style.nilProcedureResult {
		result = types.VOID
	}
	if !sig.AcceptsArgCount(len(args)) {
		a.reportBuiltinArity(sig, style, diagnosticName, len(args), call)
		if style.variantArityResult {
			result = types.VARIANT
		}
		if style.arityRecovery == stopAfterArityError {
			return result, true
		}
	}
	analyzedArgs := args
	if style.arityRecovery == recoverDeclaredArguments && len(args) > len(sig.ParamTypes) {
		analyzedArgs = args[:len(sig.ParamTypes)]
	}
	if style.analyzeAllFirst {
		actual := make([]types.Type, len(analyzedArgs))
		for index, arg := range analyzedArgs {
			actual[index] = a.analyzeExpression(arg)
		}
		a.checkBuiltinArguments(sig, style, diagnosticName, actual, call)
	} else {
		for index, arg := range analyzedArgs {
			actual := a.analyzeExpression(arg)
			a.checkBuiltinArgument(sig, style, diagnosticName, index, actual, call)
		}
	}
	return result, true
}

func (a *Analyzer) reportBuiltinArity(sig *builtins.FunctionSignature, style builtinDiagnosticStyle, name string, count int, call *ast.CallExpression) {
	if style.silentMissingArguments && count == 0 {
		return
	}
	if style.arityFormat != "" {
		a.addError(style.arityFormat, call.Token.Pos.String())
		return
	}
	a.addError("function '%s' expects %s%s, got %d at %s", name,
		builtinArityDescription(sig, style), style.aritySuffix, count, call.Token.Pos.String())
}

func builtinArityDescription(sig *builtins.FunctionSignature, style builtinDiagnosticStyle) string {
	if len(sig.AllowedArgCounts) != 0 {
		counts := make([]string, len(sig.AllowedArgCounts))
		for index, count := range sig.AllowedArgCounts {
			counts[index] = strconv.Itoa(count)
		}
		return strings.Join(counts, " or ") + " arguments"
	}
	switch {
	case sig.MinArgs == sig.MaxArgs:
		if sig.MinArgs == 1 {
			return "1 argument"
		}
		if sig.MinArgs == 0 && style.noArguments {
			return "no arguments"
		}
		return fmt.Sprintf("%d arguments", sig.MinArgs)
	case sig.MaxArgs < 0:
		return fmt.Sprintf("at least %d arguments", sig.MinArgs)
	case style.optionalOr:
		return fmt.Sprintf("%d or %d arguments", sig.MinArgs, sig.MaxArgs)
	default:
		return fmt.Sprintf("%d to %d arguments", sig.MinArgs, sig.MaxArgs)
	}
}

func (a *Analyzer) checkBuiltinArguments(sig *builtins.FunctionSignature, style builtinDiagnosticStyle, name string, actual []types.Type, call *ast.CallExpression) {
	if style.groupFormat != "" {
		valid := true
		values := make([]any, 0, len(actual)+1)
		for index, typ := range actual {
			if typ == nil {
				return
			}
			values = append(values, typ.String())
			valid = sig.AcceptsArgument(index, typ) && valid
		}
		if !valid {
			values = append(values, call.Token.Pos.String())
			a.addError(style.groupFormat, values...)
		}
		return
	}
	for index, typ := range actual {
		a.checkBuiltinArgument(sig, style, name, index, typ, call)
	}
}

func (a *Analyzer) checkBuiltinArgument(sig *builtins.FunctionSignature, style builtinDiagnosticStyle, name string, index int, actual types.Type, call *ast.CallExpression) {
	paramIndex := index
	if paramIndex >= len(sig.ParamTypes) && (sig.IsVariadic || style.arityRecovery == recoverAllArguments) {
		paramIndex = len(sig.ParamTypes) - 1
	}
	if paramIndex < 0 || sig.AcceptsArgument(paramIndex, actual) {
		return
	}
	pos := call.Token.Pos.String()
	if style.arrayElementFormat != "" && paramIndex == 0 {
		if array, ok := actual.(*types.ArrayType); ok {
			a.addError(style.arrayElementFormat, array.ElementType.String(), pos)
			return
		}
	}
	if format := builtinArgumentFormat(style, paramIndex, index); format != "" {
		a.addError(format, actual.String(), pos)
		return
	}
	expected := sig.ParamTypes[paramIndex]
	description := expected.String()
	if expected == types.FLOAT {
		description = "Integer or Float"
	}
	position := "argument"
	if len(sig.ParamTypes) > 1 {
		position = fmt.Sprintf("argument %d", index+1)
	}
	expectation := description + " as " + position
	if paramIndex < len(style.arguments) && style.arguments[paramIndex] != "" {
		expectation = style.arguments[paramIndex]
	}
	a.addError("function '%s' expects %s, got %s at %s", name, expectation, actual.String(), pos)
}

// parameterlessBuiltinType returns the result type of a builtin that takes no
// arguments at all, so a bare identifier such as `Random` or `Now` types as an
// implicit call (`Random*0`, `var t := Now`) instead of falling back to VOID.
//
// Only strictly parameterless functions qualify: a signature with optional or
// variadic parameters says nothing about whether the bare name means a call or
// a reference.
//
// Procedures are the exception. A procedure has no result, so its bare name can
// only ever mean a call, never a reference; `CleanupGlobalVars;` is a statement
// in DWScript exactly like `Randomize;`. Any non-variadic procedure whose
// parameters are all optional therefore types as VOID here.
func (a *Analyzer) parameterlessBuiltinType(name string) (types.Type, bool) {
	sig, ok := a.builtinRegistry.GetSignature(name)
	if !ok || sig.IsVariadic || sig.MinArgs != 0 {
		return nil, false
	}
	if sig.ReturnType == nil {
		return types.VOID, true
	}
	if sig.MaxArgs != 0 {
		return nil, false
	}
	return sig.ReturnType, true
}

func builtinArgumentFormat(style builtinDiagnosticStyle, paramIndex, index int) string {
	formatIndex := paramIndex
	repeat := style.arityRecovery == recoverAllArguments
	if len(style.argumentFormats) == 1 && strings.Contains(style.argumentFormats[0], "{index}") {
		repeat = true
	}
	if formatIndex >= len(style.argumentFormats) && repeat {
		formatIndex = len(style.argumentFormats) - 1
	}
	if formatIndex < 0 || formatIndex >= len(style.argumentFormats) {
		return ""
	}
	return strings.ReplaceAll(style.argumentFormats[formatIndex], "{index}", strconv.Itoa(index+1))
}
