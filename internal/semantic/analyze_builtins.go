package semantic

import (
	"github.com/cwbudde/go-dws/internal/ast"
)

// ============================================================================
// Expression Analysis
// ============================================================================

// isLValue checks if an expression is an lvalue (can be assigned to).
// Task 9.2b: Used to validate arguments to var parameters.
// An lvalue is:
//   - An identifier (variable)
//   - An index expression (array[i], string[i])
//   - A member access expression (record.field, object.field)
func (a *Analyzer) isLValue(expr ast.Expression) bool {
	switch expr.(type) {
	case *ast.Identifier:
		return true
	case *ast.IndexExpression:
		return true
	case *ast.MemberAccessExpression:
		return true
	default:
		return false
	}
}

// isBuiltinFunction checks if a name refers to a built-in function.
// Task 9.132: Helper for semantic analysis of parameterless built-in function calls.
func (a *Analyzer) isBuiltinFunction(name string) bool {
	// List of all built-in functions that can be called without parentheses
	// This should match the list in the interpreter's isBuiltinFunction
	switch name {
	case "PrintLn", "Print", "Ord", "Integer", "Length", "Copy", "Concat",
		"IndexOf", "Contains", "Reverse", "Sort", "Pos", "UpperCase",
		"LowerCase", "Trim", "TrimLeft", "TrimRight", "StringReplace", "StringOfChar",
		"Format", "Abs", "Min", "Max", "Sqr", "Power", "Sqrt", "Sin",
		"Cos", "Tan", "Random", "Randomize", "Exp", "Ln", "Log2", "Round",
		"Trunc", "Frac", "Chr", "SetLength", "High", "Low", "Assigned",
		"DegToRad", "RadToDeg", "ArcSin", "ArcCos", "ArcTan", "ArcTan2",
		"CoTan", "Hypot", "Sinh", "Cosh", "Tanh", "ArcSinh", "ArcCosh", "ArcTanh",
		"TypeOf", "SizeOf", "TypeName", "Delete", "StrToInt", "StrToFloat",
		"IntToStr", "IntToBin", "FloatToStr", "FloatToStrF", "BoolToStr", "StrToBool",
		"VarToStr", "VarIsNull", "VarIsEmpty", "VarType", "VarClear",
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

// analyzeInheritedExpression analyzes an inherited expression and returns its type.
