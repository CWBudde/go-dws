package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

// statelessBuiltins names the built-in routines DWScript registers as stateless:
// deterministic, free of side effects, and therefore constant whenever every
// argument is constant. Upstream decides this without evaluating the call, which
// is why `StrToInt('A')` is still reported as a constant instruction even though
// evaluating it would fail.
//
// The set is deliberately limited to pure conversions and pure math. Anything
// that reads or writes mutable state — Random, Now, the Var* inspectors, the
// JSON helpers, Print — is excluded, and so is anything whose result depends on
// locale or format settings.
var statelessBuiltins = map[string]bool{
	"inttostr": true, "inttohex": true, "inttobin": true,
	"strtoint": true, "strtointdef": true, "strtofloat": true, "strtofloatdef": true,
	"strtobool": true, "booltostr": true, "floattostr": true,
	"chr": true, "ord": true, "abs": true, "sqr": true, "sqrt": true,
	"sin": true, "cos": true, "tan": true, "exp": true, "ln": true, "log2": true,
	"round": true, "trunc": true, "frac": true, "power": true, "sign": true,
	"min": true, "max": true, "maxint": true, "minint": true,
	"uppercase": true, "lowercase": true, "trim": true, "trimleft": true, "trimright": true,
	"length": true, "copy": true, "pos": true, "substr": true, "substring": true,
	"leftstr": true, "rightstr": true, "midstr": true, "reversestring": true,
	"quotedstr": true, "comparestr": true, "comparetext": true, "sametext": true,
	"high": true, "low": true,
}

// hintConstantInstruction reports a statement whose expression DWScript can prove
// constant. Such a statement computes a value and discards it, so upstream emits
// `Constant Instruction - has no effect`, anchored at the start of the
// expression.
func (a *Analyzer) hintConstantInstruction(expr ast.Expression) {
	if expr == nil || a.hintsLevel < HintsLevelNormal {
		return
	}
	// A failed parse leaves stray expression statements behind — `array_error8`
	// recovers a bare `)` into one — and upstream, having stopped at the syntax
	// error, never reaches the hint. Suppressing it here keeps recovery noise out
	// of the output.
	if a.parseHadErrors {
		return
	}
	if !a.isConstantInstruction(expr) {
		return
	}
	pos := constInstructionPos(expr)
	a.addHint("Constant Instruction - has no effect [line: %d, column: %d]", pos.Line, pos.Column)
}

// constInstructionPos returns the leftmost position of an expression. A binary
// expression's own Pos is its operator token, but upstream anchors the hint at
// the start of the whole instruction, so the left operand is followed down.
func constInstructionPos(expr ast.Expression) token.Position {
	for {
		binary, ok := expr.(*ast.BinaryExpression)
		if !ok || binary.Left == nil {
			return expr.Pos()
		}
		expr = binary.Left
	}
}

// isConstantInstruction reports whether an expression is constant in the sense
// upstream uses for the constant-instruction hint. It is a structural test, not
// an evaluation: it never folds, so it stays usable on expressions whose folding
// would raise.
func (a *Analyzer) isConstantInstruction(expr ast.Expression) bool {
	switch e := expr.(type) {
	case *ast.IntegerLiteral, *ast.FloatLiteral, *ast.StringLiteral,
		*ast.BooleanLiteral, *ast.CharLiteral:
		return true
	case *ast.Identifier:
		if sym, ok := a.symbols.Resolve(e.Value); ok {
			return sym.IsConst
		}
		return a.isConstantBareBuiltin(e.Value)
	case *ast.UnaryExpression:
		return a.isConstantInstruction(e.Right)
	case *ast.BinaryExpression:
		return a.isConstantInstruction(e.Left) && a.isConstantInstruction(e.Right)
	case *ast.CallExpression:
		return a.isConstantBuiltinCall(e)
	case *ast.MemberAccessExpression:
		return a.isConstantArrayMember(e)
	}
	return false
}

// isConstantBuiltinCall reports a call to a stateless built-in whose arguments
// are all constant. A user routine that shadows the built-in name disqualifies
// the call: its body may do anything.
func (a *Analyzer) isConstantBuiltinCall(call *ast.CallExpression) bool {
	funcIdent, ok := call.Function.(*ast.Identifier)
	if !ok || !statelessBuiltins[ident.Normalize(funcIdent.Value)] {
		return false
	}
	if _, shadowed := a.symbols.Resolve(funcIdent.Value); shadowed {
		return false
	}
	for _, arg := range call.Arguments {
		if !a.isConstantInstruction(arg) {
			return false
		}
	}
	return true
}

// isConstantArrayMember reports `a.Low`, `a.High`, `a.Length` and `a.Count` on a
// static array, whose bounds are fixed at compile time. The same members on a
// dynamic array depend on the instance and are not constant.
func (a *Analyzer) isConstantArrayMember(expr *ast.MemberAccessExpression) bool {
	if expr.Member == nil {
		return false
	}
	objIdent, ok := expr.Object.(*ast.Identifier)
	if !ok {
		return false
	}
	sym, found := a.symbols.Resolve(objIdent.Value)
	if !found || sym.Type == nil {
		return false
	}
	arrayType, isArray := types.GetUnderlyingType(sym.Type).(*types.ArrayType)
	if !isArray || !arrayType.IsStatic() {
		return false
	}
	member, _ := types.LookupBuiltinHelper("array", ident.Normalize(expr.Member.Value))
	switch member.Operation {
	case types.HelperArrayLow, types.HelperArrayHigh, types.HelperArrayLength, types.HelperArrayCount:
		return true
	}
	return false
}

// isConstantBareBuiltin reports a built-in named without an argument list that
// upstream still folds to a constant, such as `MaxInt;`. The name must be
// stateless and callable with no arguments; MaxInt also accepts two, so the
// maximum arity says nothing and only the minimum is consulted.
func (a *Analyzer) isConstantBareBuiltin(name string) bool {
	if !statelessBuiltins[ident.Normalize(name)] {
		return false
	}
	sig, ok := a.builtinRegistry.GetSignature(name)
	return ok && sig.MinArgs == 0
}
