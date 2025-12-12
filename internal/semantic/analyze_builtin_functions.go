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

	// CHECK REGISTRY FOR EXISTENCE
	// Note: Some builtins (Inc, Dec, Swap, Default, etc.) are not in the registry
	// because they need special AST-level handling or aren't implemented yet in the runtime.
	// We still have detailed analyzers for these in the switch below.
	//
	// We do NOT perform arity validation here because:
	// 1. Many functions have optional parameters with complex rules
	// 2. The detailed analyzers below provide better error messages
	// 3. Registry signatures may be out of sync with actual semantic rules
	// The registry is used primarily by the interpreter for runtime validation.

	// Emit a hint when the case of a built-in differs from its declaration.
	if lowerName == "assigned" && name != "Assigned" {
		pos := callExpr.Function.Pos()
		a.addCaseMismatchHint(name, "Assigned", pos)
	}

	// Dispatch to detailed type-checking methods
	// This handles ALL builtins, including those not in the registry
	switch lowerName {
	// I/O Functions
	case "println", "print":
		return a.analyzePrintLn(args, callExpr), true

	// Type Conversion
	case "ord", "integer":
		return a.analyzeOrd(args, callExpr), true
	case "inttostr":
		return a.analyzeIntToStr(args, callExpr), true
	case "inttobin":
		return a.analyzeIntToBin(args, callExpr), true
	case "inttohex":
		return a.analyzeIntToHex(args, callExpr), true
	case "strtoint":
		return a.analyzeStrToInt(args, callExpr), true
	case "booltostr":
		return a.analyzeBoolToStr(args, callExpr), true
	case "strtofloat":
		return a.analyzeStrToFloat(args, callExpr), true
	case "vartostr":
		return a.analyzeVarToStr(args, callExpr), true
	case "floattostr":
		return a.analyzeFloatToStr(args, callExpr), true
	case "floattostrf":
		return a.analyzeFloatToStrF(args, callExpr), true
	case "strtobool":
		return a.analyzeStrToBool(args, callExpr), true
	case "strtointdef":
		return a.analyzeStrToIntDef(args, callExpr), true
	case "strtofloatdef":
		return a.analyzeStrToFloatDef(args, callExpr), true
	case "trystrtoint":
		return a.analyzeTryStrToInt(args, callExpr), true
	case "trystrtofloat":
		return a.analyzeTryStrToFloat(args, callExpr), true
	case "hextoint":
		return a.analyzeHexToInt(args, callExpr), true
	case "bintoint":
		return a.analyzeBinToInt(args, callExpr), true
	case "vartointdef":
		return a.analyzeVarToIntDef(args, callExpr), true
	case "vartofloatdef":
		return a.analyzeVarToFloatDef(args, callExpr), true
	case "chr":
		return a.analyzeChr(args, callExpr), true
	case "default":
		return a.analyzeDefault(args, callExpr), true
	case "charat":
		return a.analyzeCharAt(args, callExpr), true
	case "bytesizetostr":
		return a.analyzeByteSizeToStr(args, callExpr), true
	case "gettext", "_":
		return a.analyzeGetText(args, callExpr), true

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

	// String Functions
	case "length":
		return a.analyzeLength(args, callExpr), true
	case "copy":
		return a.analyzeCopy(args, callExpr), true
	case "substr":
		return a.analyzeSubStr(args, callExpr), true
	case "concat":
		return a.analyzeConcat(args, callExpr), true
	case "pos":
		return a.analyzePos(args, callExpr), true
	case "uppercase", "asciiuppercase", "ansiuppercase":
		return a.analyzeUpperCase(args, callExpr), true
	case "lowercase", "asciilowercase", "ansilowercase":
		return a.analyzeLowerCase(args, callExpr), true
	case "trim":
		return a.analyzeTrim(args, callExpr), true
	case "trimleft":
		return a.analyzeTrimLeft(args, callExpr), true
	case "trimright":
		return a.analyzeTrimRight(args, callExpr), true
	case "stringreplace":
		return a.analyzeStringReplace(args, callExpr), true
	case "strreplace":
		return a.analyzeStringReplace(args, callExpr), true
	case "strreplacemacros":
		return a.analyzeStrReplaceMacros(args, callExpr), true
	case "stringofchar":
		return a.analyzeStringOfChar(args, callExpr), true
	case "format":
		return a.analyzeFormat(args, callExpr), true
	case "insert":
		return a.analyzeInsert(args, callExpr), true
	case "substring":
		return a.analyzeSubString(args, callExpr), true
	case "leftstr":
		return a.analyzeLeftStr(args, callExpr), true
	case "rightstr":
		return a.analyzeRightStr(args, callExpr), true
	case "midstr":
		return a.analyzeMidStr(args, callExpr), true
	case "strbeginswith":
		return a.analyzeStrBeginsWith(args, callExpr), true
	case "strendswith":
		return a.analyzeStrEndsWith(args, callExpr), true
	case "strcontains":
		return a.analyzeStrContains(args, callExpr), true
	case "posex":
		return a.analyzePosEx(args, callExpr), true
	case "revpos":
		return a.analyzeRevPos(args, callExpr), true
	case "strfind":
		return a.analyzeStrFind(args, callExpr), true
	case "strsplit":
		return a.analyzeStrSplit(args, callExpr), true
	case "strjoin":
		return a.analyzeStrJoin(args, callExpr), true
	case "strarraypack":
		return a.analyzeStrArrayPack(args, callExpr), true
	case "strbefore":
		return a.analyzeStrBefore(args, callExpr), true
	case "strbeforelast":
		return a.analyzeStrBeforeLast(args, callExpr), true
	case "strafter":
		return a.analyzeStrAfter(args, callExpr), true
	case "strafterlast":
		return a.analyzeStrAfterLast(args, callExpr), true
	case "strbetween":
		return a.analyzeStrBetween(args, callExpr), true
	case "isdelimiter":
		return a.analyzeIsDelimiter(args, callExpr), true
	case "lastdelimiter":
		return a.analyzeLastDelimiter(args, callExpr), true
	case "finddelimiter":
		return a.analyzeFindDelimiter(args, callExpr), true
	case "padleft":
		return a.analyzePadLeft(args, callExpr), true
	case "padright":
		return a.analyzePadRight(args, callExpr), true
	case "strdeleteleft", "deleteleft":
		return a.analyzeStrDeleteLeft(args, callExpr), true
	case "strdeleteright", "deleteright":
		return a.analyzeStrDeleteRight(args, callExpr), true
	case "reversestring":
		return a.analyzeReverseString(args, callExpr), true
	case "quotedstr":
		return a.analyzeQuotedStr(args, callExpr), true
	case "stringofstring", "dupestring":
		return a.analyzeStringOfString(args, callExpr), true
	case "normalizestring", "normalize":
		return a.analyzeNormalizeString(args, callExpr), true
	case "stripaccents":
		return a.analyzeStripAccents(args, callExpr), true
	case "sametext":
		return a.analyzeSameText(args, callExpr), true
	case "comparetext":
		return a.analyzeCompareText(args, callExpr), true
	case "comparestr":
		return a.analyzeCompareStr(args, callExpr), true
	case "ansicomparetext":
		return a.analyzeAnsiCompareText(args, callExpr), true
	case "ansicomparestr":
		return a.analyzeAnsiCompareStr(args, callExpr), true
	case "comparelocalestr":
		return a.analyzeCompareLocaleStr(args, callExpr), true
	case "strmatches":
		return a.analyzeStrMatches(args, callExpr), true
	case "strisascii":
		return a.analyzeStrIsASCII(args, callExpr), true

	// Encoding/Escaping Functions
	case "strtohtml":
		return a.analyzeStrToHtml(args, callExpr), true
	case "strtohtmlattribute":
		return a.analyzeStrToHtmlAttribute(args, callExpr), true
	case "strtojson":
		return a.analyzeStrToJSON(args, callExpr), true
	case "strtocsstext":
		return a.analyzeStrToCSSText(args, callExpr), true
	case "strtoxml":
		return a.analyzeStrToXML(args, callExpr), true

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
	case "setrandseed":
		return a.analyzeSetRandSeed(args, callExpr), true
	case "isnan":
		return a.analyzeIsNaN(args, callExpr), true

	// Math Functions - Exponential/Logarithmic
	case "exp":
		return a.analyzeExp(args, callExpr), true
	case "ln":
		return a.analyzeLn(args, callExpr), true
	case "log2":
		return a.analyzeLog2(args, callExpr), true
	case "log10":
		return a.analyzeLog10(args, callExpr), true
	case "logn":
		return a.analyzeLogN(args, callExpr), true
	case "pi":
		return a.analyzePi(args, callExpr), true
	case "sign":
		return a.analyzeSign(args, callExpr), true
	case "odd":
		return a.analyzeOdd(args, callExpr), true
	case "frac":
		return a.analyzeFrac(args, callExpr), true
	case "int":
		return a.analyzeInt(args, callExpr), true
	case "infinity":
		return a.analyzeInfinity(args, callExpr), true
	case "nan":
		return a.analyzeNaN(args, callExpr), true
	case "isfinite":
		return a.analyzeIsFinite(args, callExpr), true
	case "isinfinite":
		return a.analyzeIsInfinite(args, callExpr), true
	case "intpower":
		return a.analyzeIntPower(args, callExpr), true
	case "randseed":
		return a.analyzeRandSeed(args, callExpr), true
	case "randg":
		return a.analyzeRandG(args, callExpr), true
	case "divmod":
		return a.analyzeDivMod(args, callExpr), true

	// Math Functions - Advanced
	case "factorial":
		return a.analyzeFactorial(args, callExpr), true
	case "gcd":
		return a.analyzeGcd(args, callExpr), true
	case "lcm":
		return a.analyzeLcm(args, callExpr), true
	case "isprime":
		return a.analyzeIsPrime(args, callExpr), true
	case "leastfactor":
		return a.analyzeLeastFactor(args, callExpr), true
	case "popcount":
		return a.analyzePopCount(args, callExpr), true
	case "testbit":
		return a.analyzeTestBit(args, callExpr), true
	case "haversine":
		return a.analyzeHaversine(args, callExpr), true
	case "comparenum":
		return a.analyzeCompareNum(args, callExpr), true

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
	case "assigned":
		return a.analyzeAssigned(args, callExpr), true
	case "swap":
		return a.analyzeSwap(args, callExpr), true

	// Date/Time Functions - Current time
	case "now":
		return a.analyzeNow(args, callExpr), true
	case "date":
		return a.analyzeDate(args, callExpr), true
	case "time":
		return a.analyzeTime(args, callExpr), true
	case "utcdatetime":
		return a.analyzeUTCDateTime(args, callExpr), true
	case "unixtime":
		return a.analyzeUnixTime(args, callExpr), true
	case "unixtimemsec":
		return a.analyzeUnixTimeMSec(args, callExpr), true

	// Date/Time Functions - Encoding
	case "encodedate":
		return a.analyzeEncodeDate(args, callExpr), true
	case "encodetime":
		return a.analyzeEncodeTime(args, callExpr), true
	case "encodedatetime":
		return a.analyzeEncodeDateTime(args, callExpr), true

	// Date/Time Functions - Decoding
	case "decodedate":
		return a.analyzeDecodeDate(args, callExpr), true
	case "decodetime":
		return a.analyzeDecodeTime(args, callExpr), true

	// Date/Time Functions - Component extraction
	case "yearof":
		return a.analyzeYearOf(args, callExpr), true
	case "monthof":
		return a.analyzeMonthOf(args, callExpr), true
	case "dayof":
		return a.analyzeDayOf(args, callExpr), true
	case "hourof":
		return a.analyzeHourOf(args, callExpr), true
	case "minuteof":
		return a.analyzeMinuteOf(args, callExpr), true
	case "secondof":
		return a.analyzeSecondOf(args, callExpr), true
	case "dayofweek":
		return a.analyzeDayOfWeek(args, callExpr), true
	case "dayoftheweek":
		return a.analyzeDayOfTheWeek(args, callExpr), true
	case "dayofyear":
		return a.analyzeDayOfYear(args, callExpr), true
	case "weeknumber":
		return a.analyzeWeekNumber(args, callExpr), true
	case "yearofweek":
		return a.analyzeYearOfWeek(args, callExpr), true

	// Date/Time Functions - Formatting
	case "formatdatetime":
		return a.analyzeFormatDateTime(args, callExpr), true
	case "datetimetostr":
		return a.analyzeDateTimeToStr(args, callExpr), true
	case "datetostr":
		return a.analyzeDateToStr(args, callExpr), true
	case "timetostr":
		return a.analyzeTimeToStr(args, callExpr), true
	case "datetoiso8601":
		return a.analyzeDateToISO8601(args, callExpr), true
	case "datetimetoiso8601":
		return a.analyzeDateTimeToISO8601(args, callExpr), true
	case "datetimetorfc822":
		return a.analyzeDateTimeToRFC822(args, callExpr), true

	// Date/Time Functions - Parsing
	case "strtodate":
		return a.analyzeStrToDate(args, callExpr), true
	case "strtodatetime":
		return a.analyzeStrToDateTime(args, callExpr), true
	case "strtotime":
		return a.analyzeStrToTime(args, callExpr), true
	case "iso8601todatetime":
		return a.analyzeISO8601ToDateTime(args, callExpr), true
	case "rfc822todatetime":
		return a.analyzeRFC822ToDateTime(args, callExpr), true

	// Date/Time Functions - Incrementing
	case "incyear":
		return a.analyzeIncYear(args, callExpr), true
	case "incmonth":
		return a.analyzeIncMonth(args, callExpr), true
	case "incday":
		return a.analyzeIncDay(args, callExpr), true
	case "inchour":
		return a.analyzeIncHour(args, callExpr), true
	case "incminute":
		return a.analyzeIncMinute(args, callExpr), true
	case "incsecond":
		return a.analyzeIncSecond(args, callExpr), true

	// Date/Time Functions - Difference
	case "daysbetween":
		return a.analyzeDaysBetween(args, callExpr), true
	case "hoursbetween":
		return a.analyzeHoursBetween(args, callExpr), true
	case "minutesbetween":
		return a.analyzeMinutesBetween(args, callExpr), true
	case "secondsbetween":
		return a.analyzeSecondsBetween(args, callExpr), true

	// Date/Time Functions - Special
	case "isleapyear":
		return a.analyzeIsLeapYear(args, callExpr), true
	case "firstdayofyear":
		return a.analyzeFirstDayOfYear(args, callExpr), true
	case "firstdayofnextyear":
		return a.analyzeFirstDayOfNextYear(args, callExpr), true
	case "firstdayofmonth":
		return a.analyzeFirstDayOfMonth(args, callExpr), true
	case "firstdayofnextmonth":
		return a.analyzeFirstDayOfNextMonth(args, callExpr), true
	case "firstdayofweek":
		return a.analyzeFirstDayOfWeek(args, callExpr), true

	// Date/Time Functions - Unix time conversion
	case "unixtimetodatetime":
		return a.analyzeUnixTimeToDateTime(args, callExpr), true
	case "unixtimemsectodatetime":
		return a.analyzeUnixTimeMSecToDateTime(args, callExpr), true
	case "datetimetounixtime":
		return a.analyzeDateTimeToUnixTime(args, callExpr), true
	case "datetimetounixtimemsec":
		return a.analyzeDateTimeToUnixTimeMSec(args, callExpr), true

	// JSON Functions
	case "parsejson":
		return a.analyzeParseJSON(args, callExpr), true
	case "tojson":
		return a.analyzeToJSON(args, callExpr), true
	case "tojsonformatted":
		return a.analyzeToJSONFormatted(args, callExpr), true
	case "jsonhasfield":
		return a.analyzeJSONHasField(args, callExpr), true
	case "jsonkeys":
		return a.analyzeJSONKeys(args, callExpr), true
	case "jsonvalues":
		return a.analyzeJSONValues(args, callExpr), true
	case "jsonlength":
		return a.analyzeJSONLength(args, callExpr), true

	// Variant Functions
	case "vartype":
		return a.analyzeVarType(args, callExpr), true
	case "varisnull":
		return a.analyzeVarIsNull(args, callExpr), true
	case "varisempty":
		return a.analyzeVarIsEmpty(args, callExpr), true
	case "varisclear":
		return a.analyzeVarIsClear(args, callExpr), true
	case "varisarray":
		return a.analyzeVarIsArray(args, callExpr), true
	case "varisstr":
		return a.analyzeVarIsStr(args, callExpr), true
	case "varisnumeric":
		return a.analyzeVarIsNumeric(args, callExpr), true
	case "vartoint":
		return a.analyzeVarToInt(args, callExpr), true
	case "vartofloat":
		return a.analyzeVarToFloat(args, callExpr), true
	case "varastype":
		return a.analyzeVarAsType(args, callExpr), true

	default:
		// Not a built-in function
		return nil, false
	}
}

// getBuiltinReturnType returns the return type for a built-in function WITHOUT analyzing arguments.
// This is used by the validation pass to check if something is a built-in
// and get its return type, without triggering argument analysis (which would use the wrong scope).
// For functions where the return type depends on arguments (e.g., Default), returns VARIANT.
//
// IMPORTANT: This function must stay in sync with analyzeBuiltinFunction above.
// All builtins recognized by analyzeBuiltinFunction should also be recognized here.
func (a *Analyzer) getBuiltinReturnType(name string) (types.Type, bool) {
	// Normalize function name to lowercase for case-insensitive matching
	lowerName := ident.Normalize(name)

	switch lowerName {
	// ========================================================================
	// I/O Functions - return VOID
	// ========================================================================
	case "println", "print":
		return types.VOID, true

	// ========================================================================
	// Type Conversion Functions
	// ========================================================================
	case "ord", "integer", "strtoint", "hextoint", "bintoint", "strtointdef", "vartointdef", "vartoint":
		return types.INTEGER, true
	case "inttostr", "inttobin", "inttohex", "booltostr", "vartostr", "floattostr", "floattostrf",
		"bytesizetostr", "gettext", "_", "chr", "charat":
		return types.STRING, true
	case "strtofloat", "strtofloatdef", "vartofloatdef", "vartofloat":
		return types.FLOAT, true
	case "strtobool":
		return types.BOOLEAN, true
	case "trystrtoint", "trystrtofloat":
		return types.BOOLEAN, true
	case "default", "varastype":
		return types.VARIANT, true // Return type depends on arguments

	// ========================================================================
	// Array Functions
	// ========================================================================
	case "low", "high":
		return types.INTEGER, true // For most cases, arrays return Integer bounds
	case "length":
		return types.INTEGER, true
	case "setlength", "add", "delete":
		return types.VOID, true

	// ========================================================================
	// String Functions
	// ========================================================================
	case "copy", "substr", "substring", "concat":
		return types.STRING, true
	case "pos", "posex", "revpos", "strfind":
		return types.INTEGER, true
	case "uppercase", "asciiuppercase", "ansiuppercase",
		"lowercase", "asciilowercase", "ansilowercase":
		return types.STRING, true
	case "trim", "trimleft", "trimright":
		return types.STRING, true
	case "stringreplace", "stringofchar", "format":
		return types.STRING, true
	case "insert":
		return types.VOID, true
	case "leftstr", "rightstr", "midstr":
		return types.STRING, true
	case "strbeginswith", "strendswith", "strcontains":
		return types.BOOLEAN, true
	case "strsplit":
		return types.VARIANT, true // Returns array of string
	case "strjoin", "strarraypack":
		return types.STRING, true
	case "strbefore", "strbeforelast", "strafter", "strafterlast", "strbetween":
		return types.STRING, true
	case "isdelimiter":
		return types.BOOLEAN, true
	case "lastdelimiter", "finddelimiter":
		return types.INTEGER, true
	case "padleft", "padright":
		return types.STRING, true
	case "strdeleteleft", "deleteleft", "strdeleteright", "deleteright":
		return types.STRING, true
	case "reversestring":
		return types.STRING, true
	case "quotedstr":
		return types.STRING, true
	case "stringofstring", "dupestring":
		return types.STRING, true
	case "normalizestring", "normalize", "stripaccents":
		return types.STRING, true
	case "sametext":
		return types.BOOLEAN, true
	case "comparetext", "comparestr", "ansicomparetext", "ansicomparestr", "comparelocalestr":
		return types.INTEGER, true
	case "strmatches", "strisascii":
		return types.BOOLEAN, true

	// ========================================================================
	// Encoding/Escaping Functions
	// ========================================================================
	case "strtohtml", "strtohtmlattribute", "strtojson", "strtocsstext", "strtoxml":
		return types.STRING, true

	// ========================================================================
	// Math Functions - Basic
	// ========================================================================
	case "abs", "sqr", "sqrt", "power":
		return types.FLOAT, true
	case "min", "max", "clamp", "clampint", "minint", "maxint":
		return types.VARIANT, true // Return type depends on arguments

	// ========================================================================
	// Math Functions - Trigonometric
	// ========================================================================
	case "sin", "cos", "tan", "degtorad", "radtodeg":
		return types.FLOAT, true
	case "arcsin", "arccos", "arctan", "arctan2", "cotan", "hypot":
		return types.FLOAT, true

	// ========================================================================
	// Math Functions - Hyperbolic
	// ========================================================================
	case "sinh", "cosh", "tanh", "arcsinh", "arccosh", "arctanh":
		return types.FLOAT, true

	// ========================================================================
	// Math Functions - Random
	// ========================================================================
	case "random":
		return types.FLOAT, true
	case "randomint":
		return types.INTEGER, true
	case "unsigned32":
		return types.INTEGER, true
	case "randomize":
		return types.VOID, true
	case "setrandseed":
		return types.VOID, true
	case "isnan", "isfinite", "isinfinite":
		return types.BOOLEAN, true

	// ========================================================================
	// Math Functions - Exponential/Logarithmic
	// ========================================================================
	case "exp", "ln", "log2", "log10", "logn":
		return types.FLOAT, true
	case "pi", "infinity", "nan":
		return types.FLOAT, true
	case "sign":
		return types.INTEGER, true
	case "odd":
		return types.BOOLEAN, true
	case "frac", "int":
		return types.FLOAT, true
	case "intpower":
		return types.FLOAT, true
	case "randseed":
		return types.INTEGER, true
	case "randg":
		return types.FLOAT, true
	case "divmod":
		return types.VOID, true // Modifies var parameters

	// ========================================================================
	// Math Functions - Advanced
	// ========================================================================
	case "factorial":
		return types.INTEGER, true
	case "gcd", "lcm":
		return types.INTEGER, true
	case "isprime":
		return types.BOOLEAN, true
	case "leastfactor":
		return types.INTEGER, true
	case "popcount", "testbit":
		return types.INTEGER, true
	case "haversine":
		return types.FLOAT, true
	case "comparenum":
		return types.INTEGER, true

	// ========================================================================
	// Math Functions - Rounding
	// ========================================================================
	case "round", "trunc", "ceil", "floor":
		return types.INTEGER, true

	// ========================================================================
	// Math Functions - Ordinal
	// ========================================================================
	case "inc", "dec":
		return types.VOID, true
	case "succ", "pred":
		return types.VARIANT, true // Return type matches argument type
	case "assigned":
		return types.BOOLEAN, true
	case "swap":
		return types.VOID, true

	// ========================================================================
	// Date/Time Functions - Current time
	// ========================================================================
	case "now", "date", "time", "utcdatetime":
		return types.FLOAT, true // TDateTime is Float
	case "unixtime", "unixtimemsec":
		return types.INTEGER, true

	// ========================================================================
	// Date/Time Functions - Encoding
	// ========================================================================
	case "encodedate", "encodetime", "encodedatetime":
		return types.FLOAT, true // TDateTime is Float

	// ========================================================================
	// Date/Time Functions - Decoding
	// ========================================================================
	case "decodedate", "decodetime":
		return types.VOID, true // Modifies var parameters

	// ========================================================================
	// Date/Time Functions - Component extraction
	// ========================================================================
	case "yearof", "monthof", "dayof", "hourof", "minuteof", "secondof":
		return types.INTEGER, true
	case "dayofweek", "dayoftheweek", "dayofyear", "weeknumber", "yearofweek":
		return types.INTEGER, true

	// ========================================================================
	// Date/Time Functions - Formatting
	// ========================================================================
	case "formatdatetime", "datetimetostr", "datetostr", "timetostr":
		return types.STRING, true
	case "datetoiso8601", "datetimetoiso8601", "datetimetorfc822":
		return types.STRING, true

	// ========================================================================
	// Date/Time Functions - Parsing
	// ========================================================================
	case "strtodate", "strtodatetime", "strtotime":
		return types.FLOAT, true // TDateTime is Float
	case "iso8601todatetime", "rfc822todatetime":
		return types.FLOAT, true // TDateTime is Float

	// ========================================================================
	// Date/Time Functions - Incrementing
	// ========================================================================
	case "incyear", "incmonth", "incday", "inchour", "incminute", "incsecond":
		return types.FLOAT, true // TDateTime is Float

	// ========================================================================
	// Date/Time Functions - Difference
	// ========================================================================
	case "daysbetween", "hoursbetween", "minutesbetween", "secondsbetween":
		return types.INTEGER, true

	// ========================================================================
	// Date/Time Functions - Special
	// ========================================================================
	case "isleapyear":
		return types.BOOLEAN, true
	case "firstdayofyear", "firstdayofnextyear", "firstdayofmonth",
		"firstdayofnextmonth", "firstdayofweek":
		return types.FLOAT, true // TDateTime is Float

	// ========================================================================
	// Date/Time Functions - Unix time conversion
	// ========================================================================
	case "unixtimetodatetime", "unixtimemsectodatetime":
		return types.FLOAT, true // TDateTime is Float
	case "datetimetounixtime", "datetimetounixtimemsec":
		return types.INTEGER, true

	// ========================================================================
	// JSON Functions
	// ========================================================================
	case "parsejson":
		return types.VARIANT, true
	case "tojson", "tojsonformatted":
		return types.STRING, true
	case "jsonhasfield":
		return types.BOOLEAN, true
	case "jsonkeys", "jsonvalues":
		return types.VARIANT, true // Returns array
	case "jsonlength":
		return types.INTEGER, true

	// ========================================================================
	// Variant Functions
	// ========================================================================
	case "vartype":
		return types.INTEGER, true
	case "varisnull", "varisempty", "varisclear", "varisarray", "varisstr", "varisnumeric":
		return types.BOOLEAN, true

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
