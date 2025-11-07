package semantic

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Built-in Function Analysis
// ============================================================================

// analyzeBuiltinFunction analyzes built-in function calls.
// Returns (resultType, true) if the function is a recognized built-in,
// or (nil, false) if it's not a built-in function.
func (a *Analyzer) analyzeBuiltinFunction(name string, args []ast.Expression, callExpr *ast.CallExpression) (types.Type, bool) {
	// Normalize function name to lowercase for case-insensitive matching
	lowerName := strings.ToLower(name)

	// Dispatch to specific analyzer based on function name
	switch lowerName {
	// I/O Functions
	case "println", "print":
		return a.analyzePrintLn(args, callExpr), true

	// Type Conversion
	case "ord", "integer":
		return a.analyzeOrd(args, callExpr), true

	// String Functions
	case "length":
		return a.analyzeLength(args, callExpr), true
	case "copy":
		return a.analyzeCopy(args, callExpr), true
	case "concat":
		return a.analyzeConcat(args, callExpr), true
	case "pos":
		return a.analyzePos(args, callExpr), true
	case "uppercase":
		return a.analyzeUpperCase(args, callExpr), true
	case "lowercase":
		return a.analyzeLowerCase(args, callExpr), true
	case "trim":
		return a.analyzeTrim(args, callExpr), true
	case "trimleft":
		return a.analyzeTrimLeft(args, callExpr), true
	case "trimright":
		return a.analyzeTrimRight(args, callExpr), true
	case "stringreplace":
		return a.analyzeStringReplace(args, callExpr), true
	case "stringofchar":
		return a.analyzeStringOfChar(args, callExpr), true
	case "format":
		return a.analyzeFormat(args, callExpr), true
	case "insert":
		return a.analyzeInsert(args, callExpr), true

	// Math Functions - Basic
	case "abs":
		return a.analyzeAbs(args, callExpr), true
	case "min":
		return a.analyzeMin(args, callExpr), true
	case "max":
		return a.analyzeMax(args, callExpr), true
	case "clampint":
		return a.analyzeClampInt(args, callExpr), true
	case "clamp":
		return a.analyzeClamp(args, callExpr), true
	case "maxint":
		return a.analyzeMaxInt(args, callExpr), true
	case "minint":
		return a.analyzeMinInt(args, callExpr), true
	case "sqr":
		return a.analyzeSqr(args, callExpr), true
	case "power":
		return a.analyzePower(args, callExpr), true
	case "sqrt":
		return a.analyzeSqrt(args, callExpr), true

	// Math Functions - Trigonometric
	case "sin":
		return a.analyzeSin(args, callExpr), true
	case "cos":
		return a.analyzeCos(args, callExpr), true
	case "tan":
		return a.analyzeTan(args, callExpr), true
	case "degtorad":
		return a.analyzeDegToRad(args, callExpr), true
	case "radtodeg":
		return a.analyzeRadToDeg(args, callExpr), true
	case "arcsin":
		return a.analyzeArcSin(args, callExpr), true
	case "arccos":
		return a.analyzeArcCos(args, callExpr), true
	case "arctan":
		return a.analyzeArcTan(args, callExpr), true
	case "arctan2":
		return a.analyzeArcTan2(args, callExpr), true
	case "cotan":
		return a.analyzeCoTan(args, callExpr), true
	case "hypot":
		return a.analyzeHypot(args, callExpr), true

	// Math Functions - Hyperbolic
	case "sinh":
		return a.analyzeSinh(args, callExpr), true
	case "cosh":
		return a.analyzeCosh(args, callExpr), true
	case "tanh":
		return a.analyzeTanh(args, callExpr), true
	case "arcsinh":
		return a.analyzeArcSinh(args, callExpr), true
	case "arccosh":
		return a.analyzeArcCosh(args, callExpr), true
	case "arctanh":
		return a.analyzeArcTanh(args, callExpr), true

	// Math Functions - Random
	case "random":
		return a.analyzeRandom(args, callExpr), true
	case "randomint":
		return a.analyzeRandomInt(args, callExpr), true
	case "unsigned32":
		return a.analyzeUnsigned32(args, callExpr), true
	case "randomize":
		return a.analyzeRandomize(args, callExpr), true

	// Math Functions - Exponential/Logarithmic
	case "exp":
		return a.analyzeExp(args, callExpr), true
	case "ln":
		return a.analyzeLn(args, callExpr), true
	case "log2":
		return a.analyzeLog2(args, callExpr), true

	// Math Functions - Rounding
	case "round":
		return a.analyzeRound(args, callExpr), true
	case "trunc":
		return a.analyzeTrunc(args, callExpr), true
	case "ceil":
		return a.analyzeCeil(args, callExpr), true
	case "floor":
		return a.analyzeFloor(args, callExpr), true

	// Math Functions - Ordinal
	case "inc":
		return a.analyzeInc(args, callExpr), true
	case "dec":
		return a.analyzeDec(args, callExpr), true
	case "succ":
		return a.analyzeSucc(args, callExpr), true
	case "pred":
		return a.analyzePred(args, callExpr), true

	default:
		// Not a built-in function
		return nil, false
	}
}

// ============================================================================
// Individual Built-in Function Analyzers
// ============================================================================

// analyzePrintLn analyzes the PrintLn/Print built-in function.
// These functions accept any number of arguments of any type and return void.
func (a *Analyzer) analyzePrintLn(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	// Analyze arguments for side effects (but accept any type)
	for _, arg := range args {
		a.analyzeExpression(arg)
	}
	return types.VOID
}

// analyzeOrd analyzes the Ord/Integer built-in function.
// These functions take one argument and return an integer.
func (a *Analyzer) analyzeOrd(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Ord' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze the argument
	a.analyzeExpression(args[0])
	return types.INTEGER
}
