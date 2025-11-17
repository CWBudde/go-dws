package interp

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/interp/builtins"
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
		return builtins.Pos(i, args)
	case "UpperCase":
		return builtins.UpperCase(i, args)
	case "LowerCase":
		return builtins.LowerCase(i, args)
	case "ASCIIUpperCase":
		return builtins.ASCIIUpperCase(i, args)
	case "ASCIILowerCase":
		return builtins.ASCIILowerCase(i, args)
	case "AnsiUpperCase":
		return builtins.AnsiUpperCase(i, args)
	case "AnsiLowerCase":
		return builtins.AnsiLowerCase(i, args)
	case "Trim":
		return builtins.Trim(i, args)
	case "TrimLeft":
		return builtins.TrimLeft(i, args)
	case "TrimRight":
		return builtins.TrimRight(i, args)
	case "StringReplace":
		return builtins.StringReplace(i, args)
	case "StringOfChar":
		return builtins.StringOfChar(i, args)
	case "SubStr":
		return builtins.SubStr(i, args)
	case "SubString":
		return builtins.SubString(i, args)
	case "LeftStr":
		return builtins.LeftStr(i, args)
	case "RightStr":
		return builtins.RightStr(i, args)
	case "MidStr":
		return builtins.MidStr(i, args)
	case "StrBeginsWith":
		return builtins.StrBeginsWith(i, args)
	case "StrEndsWith":
		return builtins.StrEndsWith(i, args)
	case "StrContains":
		return builtins.StrContains(i, args)
	case "PosEx":
		return builtins.PosEx(i, args)
	case "RevPos":
		return builtins.RevPos(i, args)
	case "StrFind":
		return builtins.StrFind(i, args)
	case "StrSplit":
		return i.builtinStrSplit(args)
	case "StrJoin":
		return i.builtinStrJoin(args)
	case "StrArrayPack":
		return i.builtinStrArrayPack(args)
	case "StrBefore":
		return builtins.StrBefore(i, args)
	case "StrBeforeLast":
		return builtins.StrBeforeLast(i, args)
	case "StrAfter":
		return builtins.StrAfter(i, args)
	case "StrAfterLast":
		return builtins.StrAfterLast(i, args)
	case "StrBetween":
		return builtins.StrBetween(i, args)
	case "IsDelimiter":
		return builtins.IsDelimiter(i, args)
	case "LastDelimiter":
		return builtins.LastDelimiter(i, args)
	case "FindDelimiter":
		return builtins.FindDelimiter(i, args)
	case "PadLeft":
		return builtins.PadLeft(i, args)
	case "PadRight":
		return builtins.PadRight(i, args)
	case "StrDeleteLeft":
		return builtins.StrDeleteLeft(i, args)
	case "DeleteLeft":
		return builtins.StrDeleteLeft(i, args)
	case "StrDeleteRight":
		return builtins.StrDeleteRight(i, args)
	case "DeleteRight":
		return builtins.StrDeleteRight(i, args)
	case "ReverseString":
		return builtins.ReverseString(i, args)
	case "QuotedStr":
		return builtins.QuotedStr(i, args)
	case "StringOfString":
		return builtins.StringOfString(i, args)
	case "DupeString":
		return builtins.DupeString(i, args)
	case "NormalizeString":
		return builtins.NormalizeString(i, args)
	case "Normalize":
		return builtins.NormalizeString(i, args)
	case "StripAccents":
		return builtins.StripAccents(i, args)
	case "SameText":
		return builtins.SameText(i, args)
	case "CompareText":
		return builtins.CompareText(i, args)
	case "CompareStr":
		return builtins.CompareStr(i, args)
	case "AnsiCompareText":
		return builtins.AnsiCompareText(i, args)
	case "AnsiCompareStr":
		return builtins.AnsiCompareStr(i, args)
	case "CompareLocaleStr":
		return builtins.CompareLocaleStr(i, args)
	case "StrMatches":
		return builtins.StrMatches(i, args)
	case "StrIsASCII":
		return builtins.StrIsASCII(i, args)
	case "Format":
		return i.builtinFormat(args)
	case "Abs":
		return builtins.Abs(i, args)
	case "Min":
		return builtins.Min(i, args)
	case "Max":
		return builtins.Max(i, args)
	case "ClampInt":
		return builtins.ClampInt(i, args)
	case "Clamp":
		return builtins.Clamp(i, args)
	case "Sqr":
		return builtins.Sqr(i, args)
	case "Power":
		return builtins.Power(i, args)
	case "Sqrt":
		return builtins.Sqrt(i, args)
	case "Sin":
		return builtins.Sin(i, args)
	case "Cos":
		return builtins.Cos(i, args)
	case "Tan":
		return builtins.Tan(i, args)
	case "Random":
		return i.builtinRandom(args)
	case "Randomize":
		return i.builtinRandomize(args)
	case "SetRandSeed":
		return i.builtinSetRandSeed(args)
	case "Pi":
		return builtins.Pi(i, args)
	case "Sign":
		return builtins.Sign(i, args)
	case "Odd":
		return builtins.Odd(i, args)
	case "Frac":
		return builtins.Frac(i, args)
	case "Int":
		return builtins.Int(i, args)
	case "Log10":
		return builtins.Log10(i, args)
	case "LogN":
		return builtins.LogN(i, args)
	case "Infinity":
		return builtins.Infinity(i, args)
	case "NaN":
		return builtins.NaN(i, args)
	case "IsFinite":
		return builtins.IsFinite(i, args)
	case "IsInfinite":
		return builtins.IsInfinite(i, args)
	case "IntPower":
		return builtins.IntPower(i, args)
	case "RandSeed":
		return i.builtinRandSeed(args)
	case "RandG":
		return i.builtinRandG(args)
	case "IsNaN":
		return builtins.IsNaN(i, args)
	case "Exp":
		return builtins.Exp(i, args)
	case "Ln":
		return builtins.Ln(i, args)
	case "Log2":
		return builtins.Log2(i, args)
	case "Round":
		return builtins.Round(i, args)
	case "Trunc":
		return builtins.Trunc(i, args)
	case "Ceil":
		return builtins.Ceil(i, args)
	case "Floor":
		return builtins.Floor(i, args)
	case "RandomInt":
		return i.builtinRandomInt(args)
	case "Unsigned32":
		return builtins.Unsigned32(i, args)
	case "MaxInt":
		return builtins.MaxInt(i, args)
	case "MinInt":
		return builtins.MinInt(i, args)
	case "DegToRad":
		return builtins.DegToRad(i, args)
	case "RadToDeg":
		return builtins.RadToDeg(i, args)
	case "ArcSin":
		return builtins.ArcSin(i, args)
	case "ArcCos":
		return builtins.ArcCos(i, args)
	case "ArcTan":
		return builtins.ArcTan(i, args)
	case "ArcTan2":
		return builtins.ArcTan2(i, args)
	case "CoTan":
		return builtins.CoTan(i, args)
	case "Hypot":
		return builtins.Hypot(i, args)
	case "Sinh":
		return builtins.Sinh(i, args)
	case "Cosh":
		return builtins.Cosh(i, args)
	case "Tanh":
		return builtins.Tanh(i, args)
	case "ArcSinh":
		return builtins.ArcSinh(i, args)
	case "ArcCosh":
		return builtins.ArcCosh(i, args)
	case "ArcTanh":
		return builtins.ArcTanh(i, args)
	case "Low":
		return i.builtinLow(args)
	case "High":
		return i.builtinHigh(args)
	case "SetLength":
		// SetLength is a var-param function, shouldn't be called with evaluated args
		return i.newErrorWithLocation(i.currentNode, "SetLength should be called as var-param function")
	case "Add":
		return i.builtinAdd(args)
	case "Delete":
		return i.builtinDelete(args)
	case "IntToStr":
		return i.builtinIntToStr(args)
	case "IntToBin":
		return i.builtinIntToBin(args)
	case "IntToHex":
		return builtins.IntToHex(i, args)
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
		return builtins.StrToBool(i, args)
	case "BoolToStr":
		return i.builtinBoolToStr(args)
	case "HexToInt":
		return i.builtinHexToInt(args)
	case "BinToInt":
		return i.builtinBinToInt(args)
	case "VarToIntDef":
		return i.builtinVarToIntDef(args)
	case "VarToFloatDef":
		return i.builtinVarToFloatDef(args)
	case "Chr":
		return builtins.Chr(i, args)
	case "CharAt":
		return builtins.CharAt(i, args)
	case "ByteSizeToStr":
		return builtins.ByteSizeToStr(i, args)
	case "GetText":
		return builtins.GetText(i, args)
	case "_":
		return builtins.Underscore(i, args)
	case "Succ":
		return i.builtinSucc(args)
	case "Pred":
		return i.builtinPred(args)
	case "Assert":
		return i.builtinAssert(args)
	case "Assigned":
		return i.builtinAssigned(args)
	// RTTI functions
	case "TypeOf":
		return i.builtinTypeOf(args)
	case "TypeOfClass":
		return i.builtinTypeOfClass(args)
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
		return builtins.Now(i, args)
	case "Date":
		return builtins.Date(i, args)
	case "Time":
		return builtins.Time(i, args)
	case "UTCDateTime":
		return builtins.UTCDateTime(i, args)
	// Date encoding functions
	case "EncodeDate":
		return builtins.EncodeDate(i, args)
	case "EncodeTime":
		return builtins.EncodeTime(i, args)
	case "EncodeDateTime":
		return builtins.EncodeDateTime(i, args)
	// Component extraction functions
	case "YearOf":
		return builtins.YearOf(i, args)
	case "MonthOf":
		return builtins.MonthOf(i, args)
	case "DayOf":
		return builtins.DayOf(i, args)
	case "HourOf":
		return builtins.HourOf(i, args)
	case "MinuteOf":
		return builtins.MinuteOf(i, args)
	case "SecondOf":
		return builtins.SecondOf(i, args)
	case "DayOfWeek":
		return builtins.DayOfWeek(i, args)
	case "DayOfTheWeek":
		return builtins.DayOfTheWeek(i, args)
	case "DayOfYear":
		return builtins.DayOfYear(i, args)
	case "WeekNumber":
		return builtins.WeekNumber(i, args)
	case "YearOfWeek":
		return builtins.YearOfWeek(i, args)
	// Formatting functions
	case "FormatDateTime":
		return builtins.FormatDateTime(i, args)
	case "DateTimeToStr":
		return builtins.DateTimeToStr(i, args)
	case "DateToStr":
		return builtins.DateToStr(i, args)
	case "TimeToStr":
		return builtins.TimeToStr(i, args)
	case "DateToISO8601":
		return builtins.DateToISO8601(i, args)
	case "DateTimeToISO8601":
		return builtins.DateTimeToISO8601(i, args)
	case "DateTimeToRFC822":
		return builtins.DateTimeToRFC822(i, args)
	// Parsing functions
	case "StrToDate":
		return builtins.StrToDate(i, args)
	case "StrToDateTime":
		return builtins.StrToDateTime(i, args)
	case "StrToTime":
		return builtins.StrToTime(i, args)
	case "ISO8601ToDateTime":
		return builtins.ISO8601ToDateTime(i, args)
	case "RFC822ToDateTime":
		return builtins.RFC822ToDateTime(i, args)
	// Incrementing functions
	case "IncYear":
		return builtins.IncYear(i, args)
	case "IncMonth":
		return builtins.IncMonth(i, args)
	case "IncDay":
		return builtins.IncDay(i, args)
	case "IncHour":
		return builtins.IncHour(i, args)
	case "IncMinute":
		return builtins.IncMinute(i, args)
	case "IncSecond":
		return builtins.IncSecond(i, args)
	// Date difference functions
	case "DaysBetween":
		return builtins.DaysBetween(i, args)
	case "HoursBetween":
		return builtins.HoursBetween(i, args)
	case "MinutesBetween":
		return builtins.MinutesBetween(i, args)
	case "SecondsBetween":
		return builtins.SecondsBetween(i, args)
	// Special date functions
	case "IsLeapYear":
		return builtins.IsLeapYear(i, args)
	case "FirstDayOfYear":
		return builtins.FirstDayOfYear(i, args)
	case "FirstDayOfNextYear":
		return builtins.FirstDayOfNextYear(i, args)
	case "FirstDayOfMonth":
		return builtins.FirstDayOfMonth(i, args)
	case "FirstDayOfNextMonth":
		return builtins.FirstDayOfNextMonth(i, args)
	case "FirstDayOfWeek":
		return builtins.FirstDayOfWeek(i, args)
	// Unix time functions
	case "UnixTime":
		return builtins.UnixTime(i, args)
	case "UnixTimeMSec":
		return builtins.UnixTimeMSec(i, args)
	case "UnixTimeToDateTime":
		return builtins.UnixTimeToDateTime(i, args)
	case "DateTimeToUnixTime":
		return builtins.DateTimeToUnixTime(i, args)
	case "UnixTimeMSecToDateTime":
		return builtins.UnixTimeMSecToDateTime(i, args)
	case "DateTimeToUnixTimeMSec":
		return builtins.DateTimeToUnixTimeMSec(i, args)
	// Variant introspection functions
	case "VarType":
		return i.builtinVarType(args)
	case "VarIsNull":
		return i.builtinVarIsNull(args)
	case "VarIsEmpty":
		return i.builtinVarIsEmpty(args)
	case "VarIsClear":
		return i.builtinVarIsClear(args)
	case "VarIsArray":
		return i.builtinVarIsArray(args)
	case "VarIsStr":
		return i.builtinVarIsStr(args)
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
	// Encoding/Escaping functions (Phase 9.17.6)
	case "StrToHtml":
		return i.builtinStrToHtml(args)
	case "StrToHtmlAttribute":
		return i.builtinStrToHtmlAttribute(args)
	case "StrToJSON":
		return i.builtinStrToJSON(args)
	case "StrToCSSText":
		return i.builtinStrToCSSText(args)
	case "StrToXML":
		return i.builtinStrToXML(args)
	// Advanced Math functions (Phase 9.23)
	case "Factorial":
		return builtins.Factorial(i, args)
	case "Gcd":
		return builtins.Gcd(i, args)
	case "Lcm":
		return builtins.Lcm(i, args)
	case "IsPrime":
		return builtins.IsPrime(i, args)
	case "LeastFactor":
		return builtins.LeastFactor(i, args)
	case "PopCount":
		return builtins.PopCount(i, args)
	case "TestBit":
		return builtins.TestBit(i, args)
	case "Haversine":
		return builtins.Haversine(i, args)
	case "CompareNum":
		return builtins.CompareNum(i, args)
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
		"LowerCase", "ASCIIUpperCase", "ASCIILowerCase", "AnsiUpperCase", "AnsiLowerCase",
		"Trim", "TrimLeft", "TrimRight", "StringReplace", "StringOfChar",
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
		"Trunc", "Frac", "Chr", "CharAt", "ByteSizeToStr", "GetText", "_", "SetLength", "High", "Low", "Assigned",
		"DegToRad", "RadToDeg", "ArcSin", "ArcCos", "ArcTan", "ArcTan2",
		"CoTan", "Hypot", "Sinh", "Cosh", "Tanh", "ArcSinh", "ArcCosh", "ArcTanh",
		"TypeOf", "TypeOfClass", "SizeOf", "TypeName", "Delete", "StrToInt", "StrToFloat",
		"IntToStr", "IntToBin", "IntToHex", "FloatToStr", "FloatToStrF", "BoolToStr", "StrToBool",
		"HexToInt", "BinToInt", "StrToIntDef", "StrToFloatDef", "TryStrToInt", "TryStrToFloat",
		"VarToStr", "VarToInt", "VarToFloat", "VarToIntDef", "VarToFloatDef",
		"VarAsType", "VarIsNull", "VarIsEmpty", "VarIsNumeric", "VarType", "VarClear",
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
		"GetStackTrace", "GetCallStack",
		"StrToHtml", "StrToHtmlAttribute", "StrToJSON", "StrToCSSText", "StrToXML",
		"Factorial", "Gcd", "Lcm", "IsPrime", "LeastFactor", "PopCount", "TestBit",
		"Haversine", "CompareNum":
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
	case "TryStrToInt":
		return i.builtinTryStrToInt(args)
	case "TryStrToFloat":
		return i.builtinTryStrToFloat(args)
	case "SetLength":
		return i.builtinSetLength(args)
	default:
		return i.newErrorWithLocation(i.currentNode, "undefined var-param function: %s", name)
	}
}

// callUserFunction calls a user-defined function.
// It creates a new environment, binds parameters to arguments, executes the body,
// and extracts the return value from the Result variable or function name variable.
