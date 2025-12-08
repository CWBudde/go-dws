package interp

import (
	"github.com/cwbudde/go-dws/internal/interp/builtins"
	"github.com/cwbudde/go-dws/pkg/ast"
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

	// Functions not yet migrated to the builtins registry
	// These are either:
	// - Not yet migrated (HexToInt, BinToInt, VarToIntDef, VarToFloatDef, Succ, Pred)
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
	case "Succ":
		return i.builtinSucc(args)
	case "Pred":
		return i.builtinPred(args)
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
		"Trim", "TrimLeft", "TrimRight", "StringReplace", "StrReplace", "StrReplaceMacros", "StringOfChar",
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
	case "Swap":
		return i.builtinSwap(args)
	case "DivMod":
		return i.builtinDivMod(args)
	// TryStrToInt/TryStrToFloat migrated to evaluator
	// These functions are now handled by the evaluator's visitor pattern
	case "SetLength":
		return i.builtinSetLength(args)
	default:
		return i.newErrorWithLocation(i.currentNode, "undefined var-param function: %s", name)
	}
}
