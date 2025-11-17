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

	// Check the built-in function registry (case-insensitive lookup)
	if fn, ok := builtins.DefaultRegistry.Lookup(name); ok {
		return fn(i, args)
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
	case "StrSplit":
		return i.builtinStrSplit(args)
	case "StrJoin":
		return i.builtinStrJoin(args)
	case "StrArrayPack":
		return i.builtinStrArrayPack(args)
	case "Format":
		return i.builtinFormat(args)
	case "Random":
		return i.builtinRandom(args)
	case "Randomize":
		return i.builtinRandomize(args)
	case "SetRandSeed":
		return i.builtinSetRandSeed(args)
	case "RandSeed":
		return i.builtinRandSeed(args)
	case "RandG":
		return i.builtinRandG(args)
	case "RandomInt":
		return i.builtinRandomInt(args)
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
	// Date encoding functions
	// Component extraction functions
	// Formatting functions
	// Parsing functions
	// Incrementing functions
	// Date difference functions
	// Special date functions
	// Unix time functions
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
