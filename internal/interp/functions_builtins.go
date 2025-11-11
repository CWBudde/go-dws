package interp

import (
	"github.com/cwbudde/go-dws/internal/ast"
)

// callBuiltin dispatches built-in and external functions by name.
// It normalizes the function name for DWScript's case-insensitive matching,
// and routes to the appropriate built-in implementation or external Go function.
func (i *Interpreter) callBuiltin(name string, args []Value) Value {
	// Check for external Go functions first
	if i.externalFunctions != nil {
		if extFunc, ok := i.externalFunctions.Get(name); ok {
			return i.callExternalFunction(extFunc, args)
		}
	}

	// Normalize function name for case-insensitive matching (DWScript is case-insensitive)
	name = normalizeBuiltinName(name)

	switch name {
	case "PrintLn":
		return i.builtinPrintLn(args)
	case "Print":
		return i.builtinPrint(args)
	case "Ord":
		return i.builtinOrd(args)
	case "Integer":
		return i.builtinInteger(args)
	case "Length":
		return i.builtinLength(args)
	case "Copy":
		return i.builtinCopy(args)
	case "Concat":
		return i.builtinConcat(args)
	case "IndexOf":
		return i.builtinIndexOf(args)
	case "Contains":
		return i.builtinContains(args)
	case "Reverse":
		return i.builtinReverse(args)
	case "Sort":
		return i.builtinSort(args)
	case "Pos":
		return i.builtinPos(args)
	case "UpperCase":
		return i.builtinUpperCase(args)
	case "LowerCase":
		return i.builtinLowerCase(args)
	case "Trim":
		return i.builtinTrim(args)
	case "TrimLeft":
		return i.builtinTrimLeft(args)
	case "TrimRight":
		return i.builtinTrimRight(args)
	case "StringReplace":
		return i.builtinStringReplace(args)
	case "StringOfChar":
		return i.builtinStringOfChar(args)
	case "SubStr":
		return i.builtinSubStr(args)
	case "SubString":
		return i.builtinSubString(args)
	case "LeftStr":
		return i.builtinLeftStr(args)
	case "RightStr":
		return i.builtinRightStr(args)
	case "MidStr":
		return i.builtinMidStr(args)
	case "StrBeginsWith":
		return i.builtinStrBeginsWith(args)
	case "StrEndsWith":
		return i.builtinStrEndsWith(args)
	case "StrContains":
		return i.builtinStrContains(args)
	case "PosEx":
		return i.builtinPosEx(args)
	case "RevPos":
		return i.builtinRevPos(args)
	case "StrFind":
		return i.builtinStrFind(args)
	case "StrSplit":
		return i.builtinStrSplit(args)
	case "StrJoin":
		return i.builtinStrJoin(args)
	case "StrArrayPack":
		return i.builtinStrArrayPack(args)
	case "StrBefore":
		return i.builtinStrBefore(args)
	case "StrBeforeLast":
		return i.builtinStrBeforeLast(args)
	case "StrAfter":
		return i.builtinStrAfter(args)
	case "StrAfterLast":
		return i.builtinStrAfterLast(args)
	case "StrBetween":
		return i.builtinStrBetween(args)
	case "IsDelimiter":
		return i.builtinIsDelimiter(args)
	case "LastDelimiter":
		return i.builtinLastDelimiter(args)
	case "FindDelimiter":
		return i.builtinFindDelimiter(args)
	case "PadLeft":
		return i.builtinPadLeft(args)
	case "PadRight":
		return i.builtinPadRight(args)
	case "StrDeleteLeft":
		return i.builtinStrDeleteLeft(args)
	case "DeleteLeft":
		return i.builtinStrDeleteLeft(args)
	case "StrDeleteRight":
		return i.builtinStrDeleteRight(args)
	case "DeleteRight":
		return i.builtinStrDeleteRight(args)
	case "ReverseString":
		return i.builtinReverseString(args)
	case "QuotedStr":
		return i.builtinQuotedStr(args)
	case "StringOfString":
		return i.builtinStringOfString(args)
	case "DupeString":
		return i.builtinDupeString(args)
	case "NormalizeString":
		return i.builtinNormalizeString(args)
	case "Normalize":
		return i.builtinNormalizeString(args)
	case "StripAccents":
		return i.builtinStripAccents(args)
	case "Format":
		return i.builtinFormat(args)
	case "Abs":
		return i.builtinAbs(args)
	case "Min":
		return i.builtinMin(args)
	case "Max":
		return i.builtinMax(args)
	case "ClampInt":
		return i.builtinClampInt(args)
	case "Clamp":
		return i.builtinClamp(args)
	case "Sqr":
		return i.builtinSqr(args)
	case "Power":
		return i.builtinPower(args)
	case "Sqrt":
		return i.builtinSqrt(args)
	case "Sin":
		return i.builtinSin(args)
	case "Cos":
		return i.builtinCos(args)
	case "Tan":
		return i.builtinTan(args)
	case "Random":
		return i.builtinRandom(args)
	case "Randomize":
		return i.builtinRandomize(args)
	case "SetRandSeed":
		return i.builtinSetRandSeed(args)
	case "Pi":
		return i.builtinPi(args)
	case "Sign":
		return i.builtinSign(args)
	case "Odd":
		return i.builtinOdd(args)
	case "Frac":
		return i.builtinFrac(args)
	case "Int":
		return i.builtinInt(args)
	case "Log10":
		return i.builtinLog10(args)
	case "LogN":
		return i.builtinLogN(args)
	case "Infinity":
		return i.builtinInfinity(args)
	case "NaN":
		return i.builtinNaN(args)
	case "IsFinite":
		return i.builtinIsFinite(args)
	case "IsInfinite":
		return i.builtinIsInfinite(args)
	case "IntPower":
		return i.builtinIntPower(args)
	case "RandSeed":
		return i.builtinRandSeed(args)
	case "RandG":
		return i.builtinRandG(args)
	case "IsNaN":
		return i.builtinIsNaN(args)
	case "Exp":
		return i.builtinExp(args)
	case "Ln":
		return i.builtinLn(args)
	case "Log2":
		return i.builtinLog2(args)
	case "Round":
		return i.builtinRound(args)
	case "Trunc":
		return i.builtinTrunc(args)
	case "Ceil":
		return i.builtinCeil(args)
	case "Floor":
		return i.builtinFloor(args)
	case "RandomInt":
		return i.builtinRandomInt(args)
	case "Unsigned32":
		return i.builtinUnsigned32(args)
	case "MaxInt":
		return i.builtinMaxInt(args)
	case "MinInt":
		return i.builtinMinInt(args)
	case "DegToRad":
		return i.builtinDegToRad(args)
	case "RadToDeg":
		return i.builtinRadToDeg(args)
	case "ArcSin":
		return i.builtinArcSin(args)
	case "ArcCos":
		return i.builtinArcCos(args)
	case "ArcTan":
		return i.builtinArcTan(args)
	case "ArcTan2":
		return i.builtinArcTan2(args)
	case "CoTan":
		return i.builtinCoTan(args)
	case "Hypot":
		return i.builtinHypot(args)
	case "Sinh":
		return i.builtinSinh(args)
	case "Cosh":
		return i.builtinCosh(args)
	case "Tanh":
		return i.builtinTanh(args)
	case "ArcSinh":
		return i.builtinArcSinh(args)
	case "ArcCosh":
		return i.builtinArcCosh(args)
	case "ArcTanh":
		return i.builtinArcTanh(args)
	case "Low":
		return i.builtinLow(args)
	case "High":
		return i.builtinHigh(args)
	case "SetLength":
		return i.builtinSetLength(args)
	case "Add":
		return i.builtinAdd(args)
	case "Delete":
		return i.builtinDelete(args)
	case "IntToStr":
		return i.builtinIntToStr(args)
	case "IntToBin":
		return i.builtinIntToBin(args)
	case "IntToHex":
		return i.builtinIntToHex(args)
	case "StrToInt":
		return i.builtinStrToInt(args)
	case "FloatToStr":
		return i.builtinFloatToStr(args)
	case "StrToFloat":
		return i.builtinStrToFloat(args)
	case "StrToIntDef":
		return i.builtinStrToIntDef(args)
	case "StrToFloatDef":
		return i.builtinStrToFloatDef(args)
	case "StrToBool":
		return i.builtinStrToBool(args)
	case "BoolToStr":
		return i.builtinBoolToStr(args)
	case "Chr":
		return i.builtinChr(args)
	case "Succ":
		return i.builtinSucc(args)
	case "Pred":
		return i.builtinPred(args)
	case "Assert":
		return i.builtinAssert(args)
	case "Assigned":
		return i.builtinAssigned(args)
	// Higher-order functions for working with arrays and lambdas
	case "Map":
		return i.builtinMap(args)
	case "Filter":
		return i.builtinFilter(args)
	case "Reduce":
		return i.builtinReduce(args)
	case "ForEach":
		return i.builtinForEach(args)
	case "Every":
		return i.builtinEvery(args)
	case "Some":
		return i.builtinSome(args)
	case "Find":
		return i.builtinFind(args)
	case "FindIndex":
		return i.builtinFindIndex(args)
	case "Slice":
		return i.builtinSlice(args)
	// Current date/time functions
	case "Now":
		return i.builtinNow(args)
	case "Date":
		return i.builtinDate(args)
	case "Time":
		return i.builtinTime(args)
	case "UTCDateTime":
		return i.builtinUTCDateTime(args)
	// Date encoding functions
	case "EncodeDate":
		return i.builtinEncodeDate(args)
	case "EncodeTime":
		return i.builtinEncodeTime(args)
	case "EncodeDateTime":
		return i.builtinEncodeDateTime(args)
	// Component extraction functions
	case "YearOf":
		return i.builtinYearOf(args)
	case "MonthOf":
		return i.builtinMonthOf(args)
	case "DayOf":
		return i.builtinDayOf(args)
	case "HourOf":
		return i.builtinHourOf(args)
	case "MinuteOf":
		return i.builtinMinuteOf(args)
	case "SecondOf":
		return i.builtinSecondOf(args)
	case "DayOfWeek":
		return i.builtinDayOfWeek(args)
	case "DayOfTheWeek":
		return i.builtinDayOfTheWeek(args)
	case "DayOfYear":
		return i.builtinDayOfYear(args)
	case "WeekNumber":
		return i.builtinWeekNumber(args)
	case "YearOfWeek":
		return i.builtinYearOfWeek(args)
	// Formatting functions
	case "FormatDateTime":
		return i.builtinFormatDateTime(args)
	case "DateTimeToStr":
		return i.builtinDateTimeToStr(args)
	case "DateToStr":
		return i.builtinDateToStr(args)
	case "TimeToStr":
		return i.builtinTimeToStr(args)
	case "DateToISO8601":
		return i.builtinDateToISO8601(args)
	case "DateTimeToISO8601":
		return i.builtinDateTimeToISO8601(args)
	case "DateTimeToRFC822":
		return i.builtinDateTimeToRFC822(args)
	// Parsing functions
	case "StrToDate":
		return i.builtinStrToDate(args)
	case "StrToDateTime":
		return i.builtinStrToDateTime(args)
	case "StrToTime":
		return i.builtinStrToTime(args)
	case "ISO8601ToDateTime":
		return i.builtinISO8601ToDateTime(args)
	case "RFC822ToDateTime":
		return i.builtinRFC822ToDateTime(args)
	// Incrementing functions
	case "IncYear":
		return i.builtinIncYear(args)
	case "IncMonth":
		return i.builtinIncMonth(args)
	case "IncDay":
		return i.builtinIncDay(args)
	case "IncHour":
		return i.builtinIncHour(args)
	case "IncMinute":
		return i.builtinIncMinute(args)
	case "IncSecond":
		return i.builtinIncSecond(args)
	// Date difference functions
	case "DaysBetween":
		return i.builtinDaysBetween(args)
	case "HoursBetween":
		return i.builtinHoursBetween(args)
	case "MinutesBetween":
		return i.builtinMinutesBetween(args)
	case "SecondsBetween":
		return i.builtinSecondsBetween(args)
	// Special date functions
	case "IsLeapYear":
		return i.builtinIsLeapYear(args)
	case "FirstDayOfYear":
		return i.builtinFirstDayOfYear(args)
	case "FirstDayOfNextYear":
		return i.builtinFirstDayOfNextYear(args)
	case "FirstDayOfMonth":
		return i.builtinFirstDayOfMonth(args)
	case "FirstDayOfNextMonth":
		return i.builtinFirstDayOfNextMonth(args)
	case "FirstDayOfWeek":
		return i.builtinFirstDayOfWeek(args)
	// Unix time functions
	case "UnixTime":
		return i.builtinUnixTime(args)
	case "UnixTimeMSec":
		return i.builtinUnixTimeMSec(args)
	case "UnixTimeToDateTime":
		return i.builtinUnixTimeToDateTime(args)
	case "DateTimeToUnixTime":
		return i.builtinDateTimeToUnixTime(args)
	case "UnixTimeMSecToDateTime":
		return i.builtinUnixTimeMSecToDateTime(args)
	case "DateTimeToUnixTimeMSec":
		return i.builtinDateTimeToUnixTimeMSec(args)
	// Variant introspection functions
	case "VarType":
		return i.builtinVarType(args)
	case "VarIsNull":
		return i.builtinVarIsNull(args)
	case "VarIsEmpty":
		return i.builtinVarIsEmpty(args)
	case "VarIsNumeric":
		return i.builtinVarIsNumeric(args)
	// Variant conversion functions
	case "VarToStr":
		return i.builtinVarToStr(args)
	case "VarToInt":
		return i.builtinVarToInt(args)
	case "VarToFloat":
		return i.builtinVarToFloat(args)
	case "VarAsType":
		return i.builtinVarAsType(args)
	case "VarClear":
		return i.builtinVarClear(args)
	// JSON parsing functions
	case "ParseJSON":
		return i.builtinParseJSON(args)
	// JSON serialization functions
	case "ToJSON":
		return i.builtinToJSON(args)
	case "ToJSONFormatted":
		return i.builtinToJSONFormatted(args)
	// JSON object access functions
	case "JSONHasField":
		return i.builtinJSONHasField(args)
	case "JSONKeys":
		return i.builtinJSONKeys(args)
	case "JSONValues":
		return i.builtinJSONValues(args)
	// JSON array length function
	case "JSONLength":
		return i.builtinJSONLength(args)
	// Exception Enhancements - GetStackTrace() built-in
	case "GetStackTrace":
		return i.builtinGetStackTrace(args)
	// Debugging Information - GetCallStack() built-in
	case "GetCallStack":
		return i.builtinGetCallStack(args)
	default:
		return i.newErrorWithLocation(i.currentNode, "undefined function: %s", name)
	}
}

// isBuiltinFunction checks if a name refers to a built-in function.
// Helper for checking parameterless built-in function calls.
func (i *Interpreter) isBuiltinFunction(name string) bool {
	// Check external functions first
	if i.externalFunctions != nil {
		if _, ok := i.externalFunctions.Get(name); ok {
			return true
		}
	}

	// Check built-in functions (same list as in callBuiltin switch)
	switch name {
	case "PrintLn", "Print", "Ord", "Integer", "Length", "Copy", "Concat",
		"IndexOf", "Contains", "Reverse", "Sort", "Pos", "UpperCase",
		"LowerCase", "Trim", "TrimLeft", "TrimRight", "StringReplace", "StringOfChar",
		"SubStr", "SubString", "LeftStr", "RightStr", "MidStr",
		"StrBeginsWith", "StrEndsWith", "StrContains", "PosEx", "RevPos", "StrFind",
		"StrSplit", "StrJoin", "StrArrayPack",
		"StrBefore", "StrBeforeLast", "StrAfter", "StrAfterLast", "StrBetween",
		"IsDelimiter", "LastDelimiter", "FindDelimiter",
		"PadLeft", "PadRight", "StrDeleteLeft", "DeleteLeft", "StrDeleteRight", "DeleteRight",
		"ReverseString", "QuotedStr", "StringOfString", "DupeString",
		"NormalizeString", "Normalize", "StripAccents",
		"Format", "Abs", "Min", "Max", "Sqr", "Power", "Sqrt", "Sin",
		"Cos", "Tan", "Random", "Randomize", "Exp", "Ln", "Round",
		"Trunc", "Frac", "Chr", "SetLength", "High", "Low", "Assigned",
		"DegToRad", "RadToDeg", "ArcSin", "ArcCos", "ArcTan", "ArcTan2",
		"CoTan", "Hypot", "Sinh", "Cosh", "Tanh", "ArcSinh", "ArcCosh", "ArcTanh",
		"TypeOf", "SizeOf", "TypeName", "Delete", "StrToInt", "StrToFloat",
		"IntToStr", "FloatToStr", "FloatToStrF", "BoolToStr", "StrToBool",
		"VarToStr", "VarToInt", "VarToFloat", "VarAsType", "VarIsNull", "VarIsEmpty", "VarIsNumeric", "VarType", "VarClear",
		"Include", "Exclude", "Map", "Filter", "Reduce", "ForEach",
		"MaxInt", "MinInt",
		"Now", "Date", "Time", "UTCDateTime", "EncodeDate", "EncodeTime",
		"EncodeDateTime", "YearOf", "MonthOf", "DayOf", "HourOf", "MinuteOf",
		"SecondOf", "MillisecondOf", "DayOfWeek", "DayOfYear", "WeekOfYear",
		"DateTimeToStr", "DateToStr", "TimeToStr", "FormatDateTime",
		"IncYear", "IncMonth", "IncWeek", "IncDay", "IncHour", "IncMinute",
		"IncSecond", "IncMillisecond", "DaysBetween", "HoursBetween",
		"MinutesBetween", "SecondsBetween", "MillisecondsBetween",
		"IsLeapYear", "DaysInMonth", "DaysInYear", "StartOfDay", "EndOfDay",
		"StartOfMonth", "EndOfMonth", "StartOfYear", "EndOfYear", "IsToday",
		"IsYesterday", "IsTomorrow", "IsSameDay", "CompareDate", "CompareTime",
		"CompareDateTime", "ParseJSON", "ToJSON", "ToJSONFormatted",
		"JSONHasField", "JSONKeys", "JSONValues", "JSONLength",
		"GetStackTrace", "GetCallStack":
		return true
	default:
		return false
	}
}

// callBuiltinFunction calls a built-in function by name with the given arguments.
// Helper for calling parameterless built-in functions from identifier context.
func (i *Interpreter) callBuiltinFunction(name string, args []Value) Value {
	return i.callBuiltin(name, args)
}

// callExternalFunction calls an external Go function registered via FFI
// It uses the existing FFI error handling infrastructure to safely call the Go function
// and convert any errors or panics to DWScript exceptions.
func (i *Interpreter) callExternalFunction(extFunc *ExternalFunctionValue, args []Value) Value {
	// Set interpreter reference for callback support
	// This allows the FFI wrapper to create Go callbacks that call back into DWScript
	extFunc.Wrapper.SetInterpreter(i)

	// Use the existing callExternalFunctionSafe wrapper which handles panics
	// and converts them to EHost exceptions (from ffi_errors.go)
	return i.callExternalFunctionSafe(func() (Value, error) {
		// Call the wrapped Go function
		return extFunc.Wrapper.Call(args)
	})
}

// callBuiltinWithVarParam calls a built-in function that requires var parameters.
// These functions need access to the AST nodes to modify variables in place.
func (i *Interpreter) callBuiltinWithVarParam(name string, args []ast.Expression) Value {
	switch name {
	case "Inc":
		return i.builtinInc(args)
	case "Dec":
		return i.builtinDec(args)
	case "Insert":
		return i.builtinInsert(args)
	case "Delete":
		return i.builtinDeleteString(args)
	// Date decoding functions with var parameters
	case "DecodeDate":
		return i.builtinDecodeDate(args)
	case "DecodeTime":
		return i.builtinDecodeTime(args)
	case "Swap":
		return i.builtinSwap(args)
	case "DivMod":
		return i.builtinDivMod(args)
	default:
		return i.newErrorWithLocation(i.currentNode, "undefined var-param function: %s", name)
	}
}

// callUserFunction calls a user-defined function.
// It creates a new environment, binds parameters to arguments, executes the body,
// and extracts the return value from the Result variable or function name variable.
