package semantic

import (
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// DWScript reports every argument-count problem with one of three fixed
// sentences, without naming the routine or the counts. The wording carries no
// detail on purpose: upstream raises it from the argument reader, which knows
// only that it ran out of arguments or was handed too many.
const (
	msgMoreArgumentsExpected = "More arguments expected"
	msgTooManyArguments      = "Too many arguments"
	msgNoArgumentsExpected   = "No arguments expected"
)

// addMoreArgumentsExpected reports a call that supplies fewer arguments than the
// routine requires.
func (a *Analyzer) addMoreArgumentsExpected(pos token.Position) {
	a.addError("%s at %s", msgMoreArgumentsExpected, pos.String())
}

// addTooManyArguments reports a call that supplies more arguments than the
// routine accepts. Upstream uses this whenever the routine takes at least one
// argument; see addNoArgumentsExpected for the parameterless case.
func (a *Analyzer) addTooManyArguments(pos token.Position) {
	a.addError("%s at %s", msgTooManyArguments, pos.String())
}

// addNoArgumentsExpected reports arguments passed to a routine that declares
// none.
func (a *Analyzer) addNoArgumentsExpected(pos token.Position) {
	a.addError("%s at %s", msgNoArgumentsExpected, pos.String())
}

// addArgumentCountError picks the wording upstream uses for the given counts.
//
// It is used by the call sites that name a routine — plain calls, class,
// interface, record and helper methods, constructors and `new`. The
// specialized built-in analyzers (Length, Low, DecodeDate, FloatToStrF, …), the
// signature-driven registry path in reportBuiltinArity, and function-pointer
// calls still describe their own counts; converting those is a separate,
// separately measurable change, since each carries its own per-built-in
// diagnostic policy.
func (a *Analyzer) addArgumentCountError(pos token.Position, got, minWanted, maxWanted int) {
	switch {
	case got < minWanted:
		a.addMoreArgumentsExpected(pos)
	case maxWanted == 0:
		a.addNoArgumentsExpected(pos)
	default:
		a.addTooManyArguments(pos)
	}
}

// overloadedBuiltins names the built-ins DWScript declares as an overload set,
// one signature per operand type, rather than as a single magic function. A call
// they cannot match reports that no overload accepts these arguments instead of
// naming a count, and the dedicated analyzers in analyze_builtin_math_basic.go
// report the same sentence for an argument of the wrong type.
var overloadedBuiltins = map[string]bool{
	"abs": true,
	"max": true,
	"min": true,
	"sqr": true,
}

// addNoOverloadedVersion reports a call that matches none of a routine's
// overloads.
func (a *Analyzer) addNoOverloadedVersion(name string, pos token.Position) {
	a.addError(`There is no overloaded version of "%s" that can be called with these arguments at %s`,
		name, pos.String())
}

// callNamePos returns the position upstream anchors an argument-count
// diagnostic at: the token naming the routine being called, not the opening
// parenthesis and not the start of the whole expression. For `c.Test(1)` that
// is `Test`, for `Test(1)` it is `Test`.
func callNamePos(fn ast.Expression, fallback token.Position) token.Position {
	switch callee := fn.(type) {
	case nil:
		return fallback
	case *ast.Identifier:
		return callee.Token.Pos
	case *ast.MemberAccessExpression:
		if callee.Member != nil {
			return callee.Member.Token.Pos
		}
	}
	return fallback
}
