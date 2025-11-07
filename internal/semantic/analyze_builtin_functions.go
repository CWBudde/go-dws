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

// analyzeLength analyzes the Length built-in function.
// Length takes one argument (array or string) and returns an integer.
func (a *Analyzer) analyzeLength(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Length' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze the argument
	argType := a.analyzeExpression(args[0])
	// Verify it's an array or string
	if argType != nil {
		if _, isArray := argType.(*types.ArrayType); !isArray {
			if argType != types.STRING {
				a.addError("function 'Length' expects array or string, got %s at %s",
					argType.String(), callExpr.Token.Pos.String())
			}
		}
	}
	return types.INTEGER
}

// analyzeCopy analyzes the Copy built-in function.
// Copy has two overloads:
//   - Copy(arr) - returns copy of array
//   - Copy(str, index, count) - returns substring
func (a *Analyzer) analyzeCopy(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) == 1 {
		// Copy(arr) - array copy overload
		arrType := a.analyzeExpression(args[0])
		if arrType != nil {
			if arrayType, ok := arrType.(*types.ArrayType); ok {
				// Return the same array type
				return arrayType
			}
			a.addError("function 'Copy' with 1 argument expects array, got %s at %s",
				arrType.String(), callExpr.Token.Pos.String())
		}
		// Return a generic array type as fallback
		return types.NewDynamicArrayType(types.INTEGER)
	}

	if len(args) != 3 {
		a.addError("function 'Copy' expects either 1 argument (array) or 3 arguments (string), got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}

	// Copy(str, index, count) - string copy overload
	// Analyze the first argument (string)
	strType := a.analyzeExpression(args[0])
	if strType != nil && strType != types.STRING {
		a.addError("function 'Copy' expects string as first argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	// Analyze the second argument (index - integer)
	indexType := a.analyzeExpression(args[1])
	if indexType != nil && indexType != types.INTEGER {
		a.addError("function 'Copy' expects integer as second argument, got %s at %s",
			indexType.String(), callExpr.Token.Pos.String())
	}
	// Analyze the third argument (count - integer)
	countType := a.analyzeExpression(args[2])
	if countType != nil && countType != types.INTEGER {
		a.addError("function 'Copy' expects integer as third argument, got %s at %s",
			countType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeConcat analyzes the Concat built-in function.
// Concat takes at least one argument (all strings) and returns a string.
func (a *Analyzer) analyzeConcat(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) == 0 {
		a.addError("function 'Concat' expects at least 1 argument, got 0 at %s",
			callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze all arguments and verify they're strings
	for i, arg := range args {
		argType := a.analyzeExpression(arg)
		if argType != nil && argType != types.STRING {
			a.addError("function 'Concat' expects string as argument %d, got %s at %s",
				i+1, argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.STRING
}

// analyzePos analyzes the Pos built-in function.
// Pos takes two string arguments and returns an integer.
func (a *Analyzer) analyzePos(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'Pos' expects 2 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.INTEGER
	}
	// Analyze the first argument (substring)
	substrType := a.analyzeExpression(args[0])
	if substrType != nil && substrType != types.STRING {
		a.addError("function 'Pos' expects string as first argument, got %s at %s",
			substrType.String(), callExpr.Token.Pos.String())
	}
	// Analyze the second argument (string to search in)
	strType := a.analyzeExpression(args[1])
	if strType != nil && strType != types.STRING {
		a.addError("function 'Pos' expects string as second argument, got %s at %s",
			strType.String(), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeUpperCase analyzes the UpperCase built-in function.
// UpperCase takes one string argument and returns a string.
func (a *Analyzer) analyzeUpperCase(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'UpperCase' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze the argument and verify it's a string
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.STRING {
		a.addError("function 'UpperCase' expects string as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeLowerCase analyzes the LowerCase built-in function.
// LowerCase takes one string argument and returns a string.
func (a *Analyzer) analyzeLowerCase(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'LowerCase' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze the argument and verify it's a string
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.STRING {
		a.addError("function 'LowerCase' expects string as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeTrim analyzes the Trim built-in function.
// Trim takes one string argument and returns a string.
func (a *Analyzer) analyzeTrim(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'Trim' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze the argument and verify it's a string
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.STRING {
		a.addError("function 'Trim' expects string as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}

// analyzeTrimLeft analyzes the TrimLeft built-in function.
// TrimLeft takes one string argument and returns a string.
func (a *Analyzer) analyzeTrimLeft(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'TrimLeft' expects 1 argument, got %d at %s",
			len(args), callExpr.Token.Pos.String())
		return types.STRING
	}
	// Analyze the argument and verify it's a string
	argType := a.analyzeExpression(args[0])
	if argType != nil && argType != types.STRING {
		a.addError("function 'TrimLeft' expects string as argument, got %s at %s",
			argType.String(), callExpr.Token.Pos.String())
	}
	return types.STRING
}
