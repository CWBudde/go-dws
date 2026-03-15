package interp

import (
	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// callBuiltin dispatches built-in and external functions by name.
// It normalizes the function name for DWScript's case-insensitive matching,
// and routes to the appropriate built-in implementation or external Go function.
func (i *Interpreter) callBuiltin(name string, args []Value) Value {
	// Check for external Go functions first
	if registry := i.externalFunctions(); registry != nil {
		if extFunc, ok := registry.Get(name); ok {
			return i.callExternalFunction(extFunc, args)
		}
	}

	// Check the built-in function registry (case-insensitive lookup)
	if fn, ok := builtins.DefaultRegistry.Lookup(name); ok {
		return fn(i, args)
	}

	// Normalize function name for case-insensitive matching (DWScript is case-insensitive)
	name = normalizeBuiltinName(name)

	// Functions not yet migrated to the builtins registry
	// These are either:
	// - Not yet migrated (HexToInt, BinToInt, VarToIntDef, VarToFloatDef)
	// - Need interpreter access for callbacks (Map, Filter, Reduce, etc.)
	switch name {
	case "HexToInt":
		return i.builtinHexToInt(args)
	case "BinToInt":
		return i.builtinBinToInt(args)
	case "VarToIntDef":
		return i.builtinVarToIntDef(args)
	case "VarToFloatDef":
		return i.builtinVarToFloatDef(args)
	// Higher-order functions for working with arrays and lambdas
	// These need interpreter access for callback evaluation
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
	default:
		return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "undefined function: %s", name)
	}
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
// The implementations are in internal/builtins/var_param.go.
func (i *Interpreter) callBuiltinWithVarParam(name string, args []ast.Expression) Value {
	if fn, ok := builtins.VarParamFunctions[name]; ok {
		return fn(i, args)
	}
	return i.newErrorWithLocation(i.evaluatorInstance.CurrentNode(), "undefined var-param function: %s", name)
}

func normalizeBuiltinName(name string) string {
	lower := ident.Normalize(name)

	canonicalNames := map[string]string{
		"println": "PrintLn", "print": "Print", "ord": "Ord", "integer": "Integer",
		"length": "Length", "copy": "Copy", "concat": "Concat", "indexof": "IndexOf",
		"contains": "Contains", "reverse": "Reverse", "sort": "Sort", "pos": "Pos",
		"uppercase": "UpperCase", "lowercase": "LowerCase",
		"asciiuppercase": "ASCIIUpperCase", "asciilowercase": "ASCIILowerCase",
		"ansiuppercase": "AnsiUpperCase", "ansilowercase": "AnsiLowerCase",
		"trim": "Trim", "trimleft": "TrimLeft", "trimright": "TrimRight",
		"stringreplace": "StringReplace", "strreplace": "StrReplace", "strreplacemacros": "StrReplaceMacros",
		"stringofchar": "StringOfChar", "substr": "SubStr", "substring": "SubString",
		"leftstr": "LeftStr", "rightstr": "RightStr", "midstr": "MidStr",
		"strbeginswith": "StrBeginsWith", "strendswith": "StrEndsWith", "strcontains": "StrContains",
		"posex": "PosEx", "revpos": "RevPos", "strfind": "StrFind",
		"format": "Format", "abs": "Abs", "min": "Min", "max": "Max",
		"maxint": "MaxInt", "minint": "MinInt", "sqr": "Sqr", "power": "Power",
		"sqrt": "Sqrt", "sin": "Sin", "cos": "Cos", "tan": "Tan",
		"degtorad": "DegToRad", "radtodeg": "RadToDeg", "arcsin": "ArcSin",
		"arccos": "ArcCos", "arctan": "ArcTan", "arctan2": "ArcTan2",
		"cotan": "CoTan", "hypot": "Hypot", "sinh": "Sinh", "cosh": "Cosh",
		"tanh": "Tanh", "arcsinh": "ArcSinh", "arccosh": "ArcCosh", "arctanh": "ArcTanh",
		"random": "Random", "randomint": "RandomInt", "randomize": "Randomize",
		"setrandseed": "SetRandSeed", "randseed": "RandSeed", "randg": "RandG",
		"pi": "Pi", "sign": "Sign", "odd": "Odd",
		"frac": "Frac", "int": "Int", "log10": "Log10", "logn": "LogN",
		"infinity": "Infinity", "nan": "NaN", "isfinite": "IsFinite",
		"isinfinite": "IsInfinite", "intpower": "IntPower", "divmod": "DivMod",
		"isnan":      "IsNaN",
		"unsigned32": "Unsigned32", "exp": "Exp", "ln": "Ln", "log2": "Log2",
		"round": "Round", "trunc": "Trunc", "ceil": "Ceil", "floor": "Floor",
		"low": "Low", "high": "High", "setlength": "SetLength", "add": "Add",
		"delete": "Delete", "inttostr": "IntToStr", "inttobin": "IntToBin",
		"strtoint": "StrToInt", "floattostr": "FloatToStr", "booltostr": "BoolToStr",
		"strtofloat": "StrToFloat", "strtobool": "StrToBool",
		"strtointdef": "StrToIntDef", "strtofloatdef": "StrToFloatDef",
		"chr": "Chr", "charat": "CharAt", "bytesizetostr": "ByteSizeToStr",
		"gettext": "GetText", "_": "_",
		"inc": "Inc", "dec": "Dec", "succ": "Succ",
		"pred": "Pred", "assert": "Assert", "insert": "Insert",
		"map": "Map", "filter": "Filter", "reduce": "Reduce", "foreach": "ForEach",
		"now": "Now", "date": "Date", "time": "Time", "utcdatetime": "UTCDateTime",
		"unixtime": "UnixTime", "unixtimemsec": "UnixTimeMSec",
		"encodedate": "EncodeDate", "encodetime": "EncodeTime", "encodedatetime": "EncodeDateTime",
		"decodedate": "DecodeDate", "decodetime": "DecodeTime",
		"yearof": "YearOf", "monthof": "MonthOf", "dayof": "DayOf",
		"hourof": "HourOf", "minuteof": "MinuteOf", "secondof": "SecondOf",
		"dayofweek": "DayOfWeek", "dayoftheweek": "DayOfTheWeek",
		"dayofyear": "DayOfYear", "weeknumber": "WeekNumber", "yearofweek": "YearOfWeek",
		"formatdatetime": "FormatDateTime", "datetimetostr": "DateTimeToStr",
		"datetostr": "DateToStr", "timetostr": "TimeToStr",
		"datetoiso8601": "DateToISO8601", "datetimetoiso8601": "DateTimeToISO8601",
		"datetimetorfc822": "DateTimeToRFC822",
		"strtodate":        "StrToDate", "strtodatetime": "StrToDateTime", "strtotime": "StrToTime",
		"iso8601todatetime": "ISO8601ToDateTime", "rfc822todatetime": "RFC822ToDateTime",
		"incyear": "IncYear", "incmonth": "IncMonth", "incday": "IncDay",
		"inchour": "IncHour", "incminute": "IncMinute", "incsecond": "IncSecond",
		"daysbetween": "DaysBetween", "hoursbetween": "HoursBetween",
		"minutesbetween": "MinutesBetween", "secondsbetween": "SecondsBetween",
		"isleapyear": "IsLeapYear", "swap": "Swap",
		"firstdayofyear": "FirstDayOfYear", "firstdayofnextyear": "FirstDayOfNextYear",
		"firstdayofmonth": "FirstDayOfMonth", "firstdayofnextmonth": "FirstDayOfNextMonth",
		"firstdayofweek":     "FirstDayOfWeek",
		"unixtimetodatetime": "UnixTimeToDateTime", "datetimetounixtime": "DateTimeToUnixTime",
		"unixtimemsectodatetime": "UnixTimeMSecToDateTime", "datetimetounixtimemsec": "DateTimeToUnixTimeMSec",
		"vartype": "VarType", "varisnull": "VarIsNull", "varisempty": "VarIsEmpty",
		"varisnumeric": "VarIsNumeric", "vartostr": "VarToStr", "vartoint": "VarToInt",
		"vartofloat": "VarToFloat", "varastype": "VarAsType", "varclear": "VarClear",
		"parsejson": "ParseJSON", "tojson": "ToJSON", "tojsonformatted": "ToJSONFormatted",
		"jsonhasfield": "JSONHasField", "jsonkeys": "JSONKeys", "jsonvalues": "JSONValues",
		"jsonlength":    "JSONLength",
		"getstacktrace": "GetStackTrace", "getcallstack": "GetCallStack",
		"sametext": "SameText", "comparetext": "CompareText", "comparestr": "CompareStr",
		"ansicomparetext": "AnsiCompareText", "ansicomparestr": "AnsiCompareStr",
		"comparelocalestr": "CompareLocaleStr", "strmatches": "StrMatches",
		"strisascii": "StrIsASCII",
		"strtohtml":  "StrToHtml", "strtohtmlattribute": "StrToHtmlAttribute",
		"strtojson": "StrToJSON", "strtocsstext": "StrToCSSText", "strtoxml": "StrToXML",
	}

	if canonical, ok := canonicalNames[lower]; ok {
		return canonical
	}
	return name
}
