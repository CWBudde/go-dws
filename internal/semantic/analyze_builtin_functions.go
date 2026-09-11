package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Built-in Function Analysis
// ============================================================================

// analyzeBuiltinFunction analyzes built-in function calls.
// Returns (resultType, true) if the function is a recognized built-in,
// or (nil, false) if it's not a built-in function.
func (a *Analyzer) analyzeBuiltinFunction(name string, args []ast.Expression, callExpr *ast.CallExpression) (types.Type, bool) {
	// Normalize function name to lowercase for case-insensitive matching
	lowerName := ident.Normalize(name)

	// AST-dependent and polymorphic intrinsics retain explicit analyzers.
	// Ordinary calls consume registry signatures and diagnostic policies.

	// Emit a hint when the case of a built-in differs from its declaration.
	if lowerName == "assigned" && name != "Assigned" {
		pos := callExpr.Function.Pos()
		a.addCaseMismatchHint(name, "Assigned", pos)
	}

	// Dispatch only where specialized semantic rules are needed.
	switch lowerName {
	// These calls have dedicated analysis later in analyzeFunctionCall. Their
	// result depends on argument types (array elements or the accumulator), or
	// their AST/diagnostic rules require that path before ordinary validation.
	case "map", "filter", "reduce", "foreach", "every", "some", "find", "findindex", "slice",
		"getcallstack", "assert":
		return nil, false

	// FloatToStrF is a compile-time compatibility intrinsic without a runtime
	// implementation or callable registry entry.
	case "floattostrf":
		return a.analyzeFloatToStrF(args, callExpr), true
	case "trystrtoint":
		return a.analyzeTryStrToInt(args, callExpr), true
	case "trystrtofloat":
		return a.analyzeTryStrToFloat(args, callExpr), true
	case "default":
		return a.analyzeDefault(args, callExpr), true
	// Compile-time predicates: resolved during analysis and folded to a constant.
	case "declared":
		return a.analyzeDeclared(name, args, callExpr), true
	case "conditionaldefined":
		return a.analyzeConditionalDefined(name, args, callExpr), true
	case "charat":
		return a.analyzeCharAt(args, callExpr), true

	// Array Functions
	case "low":
		return a.analyzeLow(args, callExpr), true
	case "high":
		return a.analyzeHigh(args, callExpr), true
	case "setlength":
		return a.analyzeSetLength(args, callExpr), true
	case "add":
		return a.analyzeAdd(args, callExpr), true
	case "delete":
		return a.analyzeDelete(args, callExpr), true
	case "include", "exclude":
		return a.analyzeIncludeExclude(name, args, callExpr), true

	// String Functions
	case "length":
		return a.analyzeLength(args, callExpr), true
	case "copy":
		return a.analyzeCopy(args, callExpr), true
	case "concat":
		return a.analyzeConcat(args, callExpr), true
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
	case "sqr":
		return a.analyzeSqr(args, callExpr), true

	// By-reference arithmetic
	case "divmod":
		return a.analyzeDivMod(args, callExpr), true

	// Math Functions - Ordinal
	case "inc":
		return a.analyzeInc(args, callExpr), true
	case "dec":
		return a.analyzeDec(args, callExpr), true
	case "succ":
		return a.analyzeSucc(args, callExpr), true
	case "pred":
		return a.analyzePred(args, callExpr), true
	case "swap":
		return a.analyzeSwap(args, callExpr), true

	// Date/Time Functions - Decoding
	case "decodedate":
		return a.analyzeDecodeDate(args, callExpr), true
	case "decodetime":
		return a.analyzeDecodeTime(args, callExpr), true

	// Variant Functions
	case "vartype":
		return a.analyzeVarType(args, callExpr), true
	case "vartostr":
		return a.analyzeVarToStr(args, callExpr)

	// Integer() reaches a set through the conversion registry rather than the
	// type-cast path, so the set's bitmask-width rule is applied here.
	case "integer":
		result, handled := a.analyzeRegisteredBuiltin(name, args, callExpr)
		if handled && len(args) == 1 {
			if argType := a.semanticInfo.GetResolvedType(args[0]); argType != nil {
				if setType, isSet := types.GetUnderlyingType(argType).(*types.SetType); isSet {
					a.checkSetIntegerCastWidth(setType, args[0].Pos())
				}
			}
		}
		return result, handled

	default:
		return a.analyzeRegisteredBuiltin(name, args, callExpr)
	}
}

// getBuiltinReturnType looks up the declared result without analyzing arguments.
// Intrinsics without runtime signatures retain their compile-time result here.
func (a *Analyzer) getBuiltinReturnType(name string) (types.Type, bool) {
	switch ident.Normalize(name) {
	case "floattostrf":
		return types.STRING, true
	case "trystrtoint", "trystrtofloat", "declared", "conditionaldefined":
		return types.BOOLEAN, true
	case "default":
		return types.VARIANT, true
	case "include", "exclude", "insert", "divmod", "inc", "dec", "swap", "decodedate", "decodetime":
		return types.VOID, true
	}
	signature, ok := a.builtinRegistry.GetSignature(name)
	if !ok {
		return nil, false
	}
	if signature.ReturnType == nil {
		return types.VOID, true
	}
	return signature.ReturnType, true
}
