package builtins

import "github.com/cwbudde/go-dws/internal/types"

// DefaultRegistry is the default global registry of all built-in functions.
// Populated on initialization with all standard DWScript built-ins:
// 234 functions across Math, String, DateTime, Conversion, Encoding, JSON,
// Type, Array, Collections, Variant, I/O, and System categories.
//
// Var-param functions (Inc, Dec, Swap, DivMod, SetLength) remain in the
// interpreter as they require AST-level access to modify variables in-place.
var DefaultRegistry *Registry

func init() {
	DefaultRegistry = NewRegistry()
	RegisterAll(DefaultRegistry)
}

// RegisterAll registers all built-in functions with the given registry.
// Functions are organized by category: Math, String, DateTime, Conversion,
// Encoding, JSON, Type, Array, Collections, Variant, I/O, and System.
func RegisterAll(r *Registry) {
	RegisterMathFunctions(r)
	RegisterStringFunctions(r)
	RegisterDateTimeFunctions(r)
	RegisterConversionFunctions(r)
	RegisterEncodingFunctions(r)
	RegisterJSONFunctions(r)
	RegisterTypeFunctions(r)
	RegisterIOFunctions(r)
	RegisterVariantFunctions(r)
	RegisterArrayFunctions(r)
	RegisterCollectionFunctions(r)
	RegisterSystemFunctions(r)
}

// RegisterMathFunctions registers all mathematical built-in functions.
func RegisterMathFunctions(r *Registry) {
	I := types.INTEGER
	F := types.FLOAT
	B := types.BOOLEAN
	V := types.VARIANT // Polymorphic functions like Abs, Min, Max

	// Basic math functions
	r.RegisterWithSignature("Abs", Abs, CategoryMath, "Returns the absolute value of a number",
		Sig([]types.Type{V}, V)) // Polymorphic: Integer->Integer or Float->Float
	r.RegisterWithSignature("Min", Min, CategoryMath, "Returns the minimum of two numbers",
		Sig([]types.Type{V, V}, V))
	r.RegisterWithSignature("Max", Max, CategoryMath, "Returns the maximum of two numbers",
		Sig([]types.Type{V, V}, V))
	r.RegisterWithSignature("ClampInt", ClampInt, CategoryMath, "Clamps an integer value between min and max",
		Sig([]types.Type{I, I, I}, I))
	r.RegisterWithSignature("Clamp", Clamp, CategoryMath, "Clamps a value between min and max",
		Sig([]types.Type{F, F, F}, F))
	r.RegisterWithSignature("Sqr", Sqr, CategoryMath, "Returns the square of a number",
		Sig([]types.Type{V}, V))
	r.RegisterWithSignature("Power", Power, CategoryMath, "Returns base raised to the power of exponent",
		Sig([]types.Type{F, F}, F))
	r.RegisterWithSignature("Sqrt", Sqrt, CategoryMath, "Returns the square root of a number",
		Sig([]types.Type{F}, F))
	r.RegisterWithSignature("Pi", Pi, CategoryMath, "Returns the value of π (pi)",
		Sig(nil, F))
	r.RegisterWithSignature("Sign", Sign, CategoryMath, "Returns the sign of a number (-1, 0, or 1)",
		Sig([]types.Type{V}, I))
	r.RegisterWithSignature("Odd", Odd, CategoryMath, "Returns true if the number is odd",
		Sig([]types.Type{I}, B))
	r.RegisterWithSignature("Frac", Frac, CategoryMath, "Returns the fractional part of a float",
		Sig([]types.Type{F}, F))
	r.RegisterWithSignature("Int", Int, CategoryMath, "Returns the integer part of a float",
		Sig([]types.Type{F}, F))
	r.RegisterWithSignature("Round", Round, CategoryMath, "Rounds a float to the nearest integer",
		Sig([]types.Type{F}, I))
	r.RegisterWithSignature("Trunc", Trunc, CategoryMath, "Truncates a float to an integer",
		Sig([]types.Type{F}, I))
	r.RegisterWithSignature("Ceil", Ceil, CategoryMath, "Returns the ceiling (smallest integer >= value)",
		Sig([]types.Type{F}, I))
	r.RegisterWithSignature("Floor", Floor, CategoryMath, "Returns the floor (largest integer <= value)",
		Sig([]types.Type{F}, I))
	r.RegisterWithSignature("Unsigned32", Unsigned32, CategoryMath, "Converts a signed integer to unsigned 32-bit",
		Sig([]types.Type{I}, I))
	r.RegisterWithSignature("MaxInt", MaxInt, CategoryMath, "Returns the maximum value of multiple integers",
		SigVariadic([]types.Type{I}, I, 1))
	r.RegisterWithSignature("MinInt", MinInt, CategoryMath, "Returns the minimum value of multiple integers",
		SigVariadic([]types.Type{I}, I, 1))

	// Advanced math functions
	r.RegisterWithSignature("Factorial", Factorial, CategoryMath, "Returns the factorial of n",
		Sig([]types.Type{I}, I))
	r.RegisterWithSignature("Gcd", Gcd, CategoryMath, "Returns the greatest common divisor",
		Sig([]types.Type{I, I}, I))
	r.RegisterWithSignature("Lcm", Lcm, CategoryMath, "Returns the least common multiple",
		Sig([]types.Type{I, I}, I))
	r.RegisterWithSignature("IsPrime", IsPrime, CategoryMath, "Returns true if n is a prime number",
		Sig([]types.Type{I}, B))
	r.RegisterWithSignature("LeastFactor", LeastFactor, CategoryMath, "Returns the smallest prime factor",
		Sig([]types.Type{I}, I))
	r.RegisterWithSignature("PopCount", PopCount, CategoryMath, "Returns the number of set bits",
		Sig([]types.Type{I}, I))
	r.RegisterWithSignature("TestBit", TestBit, CategoryMath, "Tests if a specific bit is set",
		Sig([]types.Type{I, I}, B))
	r.RegisterWithSignature("Haversine", Haversine, CategoryMath, "Calculates the haversine distance",
		Sig([]types.Type{F, F, F, F}, F))
	r.RegisterWithSignature("CompareNum", CompareNum, CategoryMath, "Compares two numbers (-1, 0, 1)",
		Sig([]types.Type{V, V}, I))

	// Exponential and logarithmic functions
	r.RegisterWithSignature("Exp", Exp, CategoryMath, "Returns e raised to the power of x",
		Sig([]types.Type{F}, F))
	r.RegisterWithSignature("Ln", Ln, CategoryMath, "Returns the natural logarithm",
		Sig([]types.Type{F}, F))
	r.RegisterWithSignature("Log2", Log2, CategoryMath, "Returns the base-2 logarithm",
		Sig([]types.Type{F}, F))
	r.RegisterWithSignature("Log10", Log10, CategoryMath, "Returns the base-10 logarithm",
		Sig([]types.Type{F}, F))
	r.RegisterWithSignature("LogN", LogN, CategoryMath, "Returns the logarithm with custom base",
		Sig([]types.Type{F, F}, F))
	r.RegisterWithSignature("IntPower", IntPower, CategoryMath, "Returns base raised to integer exponent",
		Sig([]types.Type{F, I}, F))

	// Special values
	r.RegisterWithSignature("Infinity", Infinity, CategoryMath, "Returns positive infinity",
		Sig(nil, F))
	r.RegisterWithSignature("NaN", NaN, CategoryMath, "Returns NaN (Not a Number)",
		Sig(nil, F))
	r.RegisterWithSignature("IsFinite", IsFinite, CategoryMath, "Returns true if value is finite",
		Sig([]types.Type{F}, B))
	r.RegisterWithSignature("IsInfinite", IsInfinite, CategoryMath, "Returns true if value is infinite",
		Sig([]types.Type{F}, B))
	r.RegisterWithSignature("IsNaN", IsNaN, CategoryMath, "Returns true if value is NaN",
		Sig([]types.Type{F}, B))

	// Trigonometric functions
	r.RegisterWithSignature("Sin", Sin, CategoryMath, "Returns the sine of x (radians)",
		Sig([]types.Type{F}, F))
	r.RegisterWithSignature("Cos", Cos, CategoryMath, "Returns the cosine of x (radians)",
		Sig([]types.Type{F}, F))
	r.RegisterWithSignature("Tan", Tan, CategoryMath, "Returns the tangent of x (radians)",
		Sig([]types.Type{F}, F))
	r.RegisterWithSignature("CoTan", CoTan, CategoryMath, "Returns the cotangent of x (radians)",
		Sig([]types.Type{F}, F))
	r.RegisterWithSignature("ArcSin", ArcSin, CategoryMath, "Returns the arcsine of x",
		Sig([]types.Type{F}, F))
	r.RegisterWithSignature("ArcCos", ArcCos, CategoryMath, "Returns the arccosine of x",
		Sig([]types.Type{F}, F))
	r.RegisterWithSignature("ArcTan", ArcTan, CategoryMath, "Returns the arctangent of x",
		Sig([]types.Type{F}, F))
	r.RegisterWithSignature("ArcTan2", ArcTan2, CategoryMath, "Returns the arctangent of y/x",
		Sig([]types.Type{F, F}, F))
	r.RegisterWithSignature("DegToRad", DegToRad, CategoryMath, "Converts degrees to radians",
		Sig([]types.Type{F}, F))
	r.RegisterWithSignature("RadToDeg", RadToDeg, CategoryMath, "Converts radians to degrees",
		Sig([]types.Type{F}, F))
	r.RegisterWithSignature("Hypot", Hypot, CategoryMath, "Returns the hypotenuse (sqrt(x²+y²))",
		Sig([]types.Type{F, F}, F))

	// Hyperbolic functions
	r.RegisterWithSignature("Sinh", Sinh, CategoryMath, "Returns the hyperbolic sine",
		Sig([]types.Type{F}, F))
	r.RegisterWithSignature("Cosh", Cosh, CategoryMath, "Returns the hyperbolic cosine",
		Sig([]types.Type{F}, F))
	r.RegisterWithSignature("Tanh", Tanh, CategoryMath, "Returns the hyperbolic tangent",
		Sig([]types.Type{F}, F))
	r.RegisterWithSignature("ArcSinh", ArcSinh, CategoryMath, "Returns the inverse hyperbolic sine",
		Sig([]types.Type{F}, F))
	r.RegisterWithSignature("ArcCosh", ArcCosh, CategoryMath, "Returns the inverse hyperbolic cosine",
		Sig([]types.Type{F}, F))
	r.RegisterWithSignature("ArcTanh", ArcTanh, CategoryMath, "Returns the inverse hyperbolic tangent",
		Sig([]types.Type{F}, F))

	// Random number functions
	r.RegisterWithSignature("Random", Random, CategoryMath, "Returns a random float between 0 and 1",
		Sig(nil, F))
	r.RegisterWithSignature("RandomInt", RandomInt, CategoryMath, "Returns a random integer in range",
		Sig([]types.Type{I}, I))
	r.RegisterWithSignature("Randomize", Randomize, CategoryMath, "Seeds the random number generator",
		Sig(nil, nil)) // Procedure
	r.RegisterWithSignature("SetRandSeed", SetRandSeed, CategoryMath, "Sets the random number seed",
		Sig([]types.Type{I}, nil)) // Procedure
	r.RegisterWithSignature("RandSeed", RandSeed, CategoryMath, "Returns the current random seed",
		Sig(nil, I))
	r.RegisterWithSignature("RandG", RandG, CategoryMath, "Returns a random Gaussian value",
		Sig([]types.Type{F, F}, F))
}

// RegisterStringFunctions registers all string manipulation built-in functions.
func RegisterStringFunctions(r *Registry) {
	S := types.STRING
	I := types.INTEGER
	B := types.BOOLEAN
	V := types.VARIANT

	// Basic string functions
	r.RegisterWithSignature("Pos", Pos, CategoryString, "Finds the position of a substring",
		Sig([]types.Type{S, S}, I))
	r.RegisterWithSignature("UpperCase", UpperCase, CategoryString, "Converts string to uppercase",
		Sig([]types.Type{S}, S))
	r.RegisterWithSignature("LowerCase", LowerCase, CategoryString, "Converts string to lowercase",
		Sig([]types.Type{S}, S))
	r.RegisterWithSignature("ASCIIUpperCase", ASCIIUpperCase, CategoryString, "Converts ASCII characters to uppercase",
		Sig([]types.Type{S}, S))
	r.RegisterWithSignature("ASCIILowerCase", ASCIILowerCase, CategoryString, "Converts ASCII characters to lowercase",
		Sig([]types.Type{S}, S))
	r.RegisterWithSignature("AnsiUpperCase", AnsiUpperCase, CategoryString, "Converts ANSI string to uppercase",
		Sig([]types.Type{S}, S))
	r.RegisterWithSignature("AnsiLowerCase", AnsiLowerCase, CategoryString, "Converts ANSI string to lowercase",
		Sig([]types.Type{S}, S))
	r.RegisterWithSignature("Trim", Trim, CategoryString, "Removes leading and trailing whitespace",
		SigOptional([]types.Type{S, S}, S, 1))
	r.RegisterWithSignature("TrimLeft", TrimLeft, CategoryString, "Removes leading whitespace",
		SigOptional([]types.Type{S, S}, S, 1))
	r.RegisterWithSignature("TrimRight", TrimRight, CategoryString, "Removes trailing whitespace",
		SigOptional([]types.Type{S, S}, S, 1))
	r.RegisterWithSignature("StringReplace", StringReplace, CategoryString, "Replaces occurrences of a substring",
		Sig([]types.Type{S, S, S}, S))
	r.RegisterWithSignature("StrReplace", StrReplace, CategoryString, "Alias for StringReplace",
		Sig([]types.Type{S, S, S}, S))
	r.RegisterWithSignature("StrReplaceMacros", StrReplaceMacros, CategoryString, "Replaces macros delimited in a string",
		Sig([]types.Type{S, V, S, S}, S)) // (str, callback, startDelim, endDelim)
	r.RegisterWithSignature("StringOfChar", StringOfChar, CategoryString, "Creates a string of repeated characters",
		Sig([]types.Type{S, I}, S))
	r.RegisterWithSignature("SubStr", SubStr, CategoryString, "Extracts a substring",
		SigOptional([]types.Type{S, I, I}, S, 2))
	r.RegisterWithSignature("SubString", SubString, CategoryString, "Extracts a substring (alias)",
		SigOptional([]types.Type{S, I, I}, S, 2))
	r.RegisterWithSignature("LeftStr", LeftStr, CategoryString, "Returns leftmost characters",
		Sig([]types.Type{S, I}, S))
	r.RegisterWithSignature("RightStr", RightStr, CategoryString, "Returns rightmost characters",
		Sig([]types.Type{S, I}, S))
	r.RegisterWithSignature("MidStr", MidStr, CategoryString, "Extracts middle substring",
		SigOptional([]types.Type{S, I, I}, S, 2))
	r.RegisterWithSignature("Chr", Chr, CategoryString, "Converts character code to string",
		Sig([]types.Type{I}, S))

	// String search functions
	r.RegisterWithSignature("StrBeginsWith", StrBeginsWith, CategoryString, "Checks if string starts with prefix",
		Sig([]types.Type{S, S}, B))
	r.RegisterWithSignature("StrEndsWith", StrEndsWith, CategoryString, "Checks if string ends with suffix",
		Sig([]types.Type{S, S}, B))
	r.RegisterWithSignature("StrContains", StrContains, CategoryString, "Checks if string contains substring",
		Sig([]types.Type{S, S}, B))
	r.RegisterWithSignature("PosEx", PosEx, CategoryString, "Finds position with start index",
		Sig([]types.Type{S, S, I}, I))
	r.RegisterWithSignature("RevPos", RevPos, CategoryString, "Finds last position of substring",
		Sig([]types.Type{S, S}, I))
	r.RegisterWithSignature("StrFind", StrFind, CategoryString, "Finds substring in string",
		SigOptional([]types.Type{S, S, I}, I, 2))

	// Advanced string functions
	r.RegisterWithSignature("StrBefore", StrBefore, CategoryString, "Returns text before delimiter",
		Sig([]types.Type{S, S}, S))
	r.RegisterWithSignature("StrBeforeLast", StrBeforeLast, CategoryString, "Returns text before last delimiter",
		Sig([]types.Type{S, S}, S))
	r.RegisterWithSignature("StrAfter", StrAfter, CategoryString, "Returns text after delimiter",
		Sig([]types.Type{S, S}, S))
	r.RegisterWithSignature("StrAfterLast", StrAfterLast, CategoryString, "Returns text after last delimiter",
		Sig([]types.Type{S, S}, S))
	r.RegisterWithSignature("StrBetween", StrBetween, CategoryString, "Extracts text between delimiters",
		Sig([]types.Type{S, S, S}, S))
	r.RegisterWithSignature("StrSplit", StrSplit, CategoryString, "Splits string into array by delimiter",
		Sig([]types.Type{S, S}, V)) // Returns array of string
	r.RegisterWithSignature("StrJoin", StrJoin, CategoryString, "Joins array of strings with delimiter",
		Sig([]types.Type{V, S}, S)) // Array of string, delimiter
	r.RegisterWithSignature("StrArrayPack", StrArrayPack, CategoryString, "Removes empty strings from array",
		Sig([]types.Type{V}, V)) // Array -> Array
	r.RegisterWithSignature("IsDelimiter", IsDelimiter, CategoryString, "Checks if character is a delimiter",
		Sig([]types.Type{S, S, I}, B))
	r.RegisterWithSignature("LastDelimiter", LastDelimiter, CategoryString, "Finds last delimiter position",
		Sig([]types.Type{S, S}, I))
	r.RegisterWithSignature("FindDelimiter", FindDelimiter, CategoryString, "Finds first delimiter position",
		SigOptional([]types.Type{S, S, I}, I, 2))
	r.RegisterWithSignature("PadLeft", PadLeft, CategoryString, "Pads string on the left",
		SigOptional([]types.Type{S, I, S}, S, 2))
	r.RegisterWithSignature("PadRight", PadRight, CategoryString, "Pads string on the right",
		SigOptional([]types.Type{S, I, S}, S, 2))
	r.RegisterWithSignature("StrDeleteLeft", StrDeleteLeft, CategoryString, "Deletes characters from left",
		Sig([]types.Type{S, I}, S))
	r.RegisterWithSignature("DeleteLeft", StrDeleteLeft, CategoryString, "Deletes characters from left (alias)",
		Sig([]types.Type{S, I}, S))
	r.RegisterWithSignature("StrDeleteRight", StrDeleteRight, CategoryString, "Deletes characters from right",
		Sig([]types.Type{S, I}, S))
	r.RegisterWithSignature("DeleteRight", StrDeleteRight, CategoryString, "Deletes characters from right (alias)",
		Sig([]types.Type{S, I}, S))
	r.RegisterWithSignature("ReverseString", ReverseString, CategoryString, "Reverses a string",
		Sig([]types.Type{S}, S))
	r.RegisterWithSignature("QuotedStr", QuotedStr, CategoryString, "Returns quoted string",
		SigOptional([]types.Type{S, S}, S, 1))
	r.RegisterWithSignature("StringOfString", StringOfString, CategoryString, "Repeats a string",
		Sig([]types.Type{S, I}, S))
	r.RegisterWithSignature("DupeString", DupeString, CategoryString, "Duplicates a string",
		Sig([]types.Type{S, I}, S))
	r.RegisterWithSignature("NormalizeString", NormalizeString, CategoryString, "Normalizes Unicode string",
		SigOptional([]types.Type{S, S}, S, 1))
	r.RegisterWithSignature("Normalize", NormalizeString, CategoryString, "Normalizes Unicode string (alias)",
		SigOptional([]types.Type{S, S}, S, 1))
	r.RegisterWithSignature("StripAccents", StripAccents, CategoryString, "Removes accents from characters",
		Sig([]types.Type{S}, S))
	r.RegisterWithSignature("ByteSizeToStr", ByteSizeToStr, CategoryString, "Formats byte size as human-readable string",
		Sig([]types.Type{I}, S))
	r.RegisterWithSignature("GetText", GetText, CategoryString, "Localizes text (i18n)",
		Sig([]types.Type{S}, S))
	r.RegisterWithSignature("CharAt", CharAt, CategoryString, "Returns character at index",
		Sig([]types.Type{S, I}, S))
	r.RegisterWithSignature("Underscore", Underscore, CategoryString, "Converts string to underscore_case",
		Sig([]types.Type{S}, S))
	r.RegisterWithSignature("_", Underscore, CategoryString, "Converts string to underscore_case (alias for Underscore)",
		Sig([]types.Type{S}, S))

	// String comparison functions
	r.RegisterWithSignature("SameText", SameText, CategoryString, "Case-insensitive string equality",
		Sig([]types.Type{S, S}, B))
	r.RegisterWithSignature("CompareText", CompareText, CategoryString, "Case-insensitive string comparison",
		Sig([]types.Type{S, S}, I))
	r.RegisterWithSignature("CompareStr", CompareStr, CategoryString, "Case-sensitive string comparison",
		Sig([]types.Type{S, S}, I))
	r.RegisterWithSignature("AnsiCompareText", AnsiCompareText, CategoryString, "ANSI case-insensitive comparison",
		Sig([]types.Type{S, S}, I))
	r.RegisterWithSignature("AnsiCompareStr", AnsiCompareStr, CategoryString, "ANSI case-sensitive comparison",
		Sig([]types.Type{S, S}, I))
	r.RegisterWithSignature("CompareLocaleStr", CompareLocaleStr, CategoryString, "Locale-aware string comparison",
		Sig([]types.Type{S, S}, I))
	r.RegisterWithSignature("StrMatches", StrMatches, CategoryString, "Tests if string matches pattern",
		Sig([]types.Type{S, S}, B))
	r.RegisterWithSignature("StrIsASCII", StrIsASCII, CategoryString, "Checks if string is ASCII only",
		Sig([]types.Type{S}, B))
}

// RegisterDateTimeFunctions registers all date/time built-in functions.
func RegisterDateTimeFunctions(r *Registry) {
	F := types.FLOAT // DateTime stored as Float
	I := types.INTEGER
	S := types.STRING
	B := types.BOOLEAN

	// Date/time creation
	r.RegisterWithSignature("EncodeDate", EncodeDate, CategoryDateTime, "Creates date from year, month, day",
		Sig([]types.Type{I, I, I}, F))
	r.RegisterWithSignature("EncodeTime", EncodeTime, CategoryDateTime, "Creates time from hour, minute, second",
		SigOptional([]types.Type{I, I, I, I}, F, 3)) // msec optional
	r.RegisterWithSignature("EncodeDateTime", EncodeDateTime, CategoryDateTime, "Creates datetime from components",
		SigOptional([]types.Type{I, I, I, I, I, I, I}, F, 6)) // msec optional
	r.RegisterWithSignature("Now", Now, CategoryDateTime, "Returns current date and time",
		Sig(nil, F))
	r.RegisterWithSignature("Date", Date, CategoryDateTime, "Returns current date",
		Sig(nil, F))
	r.RegisterWithSignature("Time", Time, CategoryDateTime, "Returns current time",
		Sig(nil, F))
	r.RegisterWithSignature("UTCDateTime", UTCDateTime, CategoryDateTime, "Returns current UTC datetime",
		Sig(nil, F))

	// Date/time arithmetic
	r.RegisterWithSignature("IncYear", IncYear, CategoryDateTime, "Adds years to a date",
		SigOptional([]types.Type{F, I}, F, 1))
	r.RegisterWithSignature("IncMonth", IncMonth, CategoryDateTime, "Adds months to a date",
		SigOptional([]types.Type{F, I}, F, 1))
	r.RegisterWithSignature("IncDay", IncDay, CategoryDateTime, "Adds days to a date",
		SigOptional([]types.Type{F, I}, F, 1))
	r.RegisterWithSignature("IncHour", IncHour, CategoryDateTime, "Adds hours to a datetime",
		SigOptional([]types.Type{F, I}, F, 1))
	r.RegisterWithSignature("IncMinute", IncMinute, CategoryDateTime, "Adds minutes to a datetime",
		SigOptional([]types.Type{F, I}, F, 1))
	r.RegisterWithSignature("IncSecond", IncSecond, CategoryDateTime, "Adds seconds to a datetime",
		SigOptional([]types.Type{F, I}, F, 1))
	r.RegisterWithSignature("DaysBetween", DaysBetween, CategoryDateTime, "Returns days between two dates",
		Sig([]types.Type{F, F}, I))
	r.RegisterWithSignature("HoursBetween", HoursBetween, CategoryDateTime, "Returns hours between two datetimes",
		Sig([]types.Type{F, F}, I))
	r.RegisterWithSignature("MinutesBetween", MinutesBetween, CategoryDateTime, "Returns minutes between two datetimes",
		Sig([]types.Type{F, F}, I))
	r.RegisterWithSignature("SecondsBetween", SecondsBetween, CategoryDateTime, "Returns seconds between two datetimes",
		Sig([]types.Type{F, F}, I))

	// Date/time formatting
	r.RegisterWithSignature("FormatDateTime", FormatDateTime, CategoryDateTime, "Formats datetime with format string",
		Sig([]types.Type{S, F}, S))
	r.RegisterWithSignature("DateTimeToStr", DateTimeToStr, CategoryDateTime, "Converts datetime to string",
		Sig([]types.Type{F}, S))
	r.RegisterWithSignature("DateToStr", DateToStr, CategoryDateTime, "Converts date to string",
		Sig([]types.Type{F}, S))
	r.RegisterWithSignature("TimeToStr", TimeToStr, CategoryDateTime, "Converts time to string",
		Sig([]types.Type{F}, S))
	r.RegisterWithSignature("DateToISO8601", DateToISO8601, CategoryDateTime, "Converts date to ISO8601 format",
		Sig([]types.Type{F}, S))
	r.RegisterWithSignature("DateTimeToISO8601", DateTimeToISO8601, CategoryDateTime, "Converts datetime to ISO8601 format",
		Sig([]types.Type{F}, S))
	r.RegisterWithSignature("DateTimeToRFC822", DateTimeToRFC822, CategoryDateTime, "Converts datetime to RFC822 format",
		Sig([]types.Type{F}, S))

	// Date/time parsing
	r.RegisterWithSignature("StrToDate", StrToDate, CategoryDateTime, "Parses string to date",
		Sig([]types.Type{S}, F))
	r.RegisterWithSignature("StrToDateTime", StrToDateTime, CategoryDateTime, "Parses string to datetime",
		Sig([]types.Type{S}, F))
	r.RegisterWithSignature("StrToTime", StrToTime, CategoryDateTime, "Parses string to time",
		Sig([]types.Type{S}, F))
	r.RegisterWithSignature("ISO8601ToDateTime", ISO8601ToDateTime, CategoryDateTime, "Parses ISO8601 to datetime",
		Sig([]types.Type{S}, F))
	r.RegisterWithSignature("RFC822ToDateTime", RFC822ToDateTime, CategoryDateTime, "Parses RFC822 to datetime",
		Sig([]types.Type{S}, F))

	// Unix time conversions
	r.RegisterWithSignature("UnixTime", UnixTime, CategoryDateTime, "Returns current Unix timestamp",
		Sig(nil, I))
	r.RegisterWithSignature("UnixTimeMSec", UnixTimeMSec, CategoryDateTime, "Returns current Unix timestamp in milliseconds",
		Sig(nil, I))
	r.RegisterWithSignature("UnixTimeToDateTime", UnixTimeToDateTime, CategoryDateTime, "Converts Unix timestamp to datetime",
		Sig([]types.Type{I}, F))
	r.RegisterWithSignature("DateTimeToUnixTime", DateTimeToUnixTime, CategoryDateTime, "Converts datetime to Unix timestamp",
		Sig([]types.Type{F}, I))
	r.RegisterWithSignature("UnixTimeMSecToDateTime", UnixTimeMSecToDateTime, CategoryDateTime, "Converts Unix milliseconds to datetime",
		Sig([]types.Type{I}, F))
	r.RegisterWithSignature("DateTimeToUnixTimeMSec", DateTimeToUnixTimeMSec, CategoryDateTime, "Converts datetime to Unix milliseconds",
		Sig([]types.Type{F}, I))

	// Date/time information
	r.RegisterWithSignature("YearOf", YearOf, CategoryDateTime, "Extracts year from datetime",
		Sig([]types.Type{F}, I))
	r.RegisterWithSignature("MonthOf", MonthOf, CategoryDateTime, "Extracts month from datetime",
		Sig([]types.Type{F}, I))
	r.RegisterWithSignature("DayOf", DayOf, CategoryDateTime, "Extracts day from datetime",
		Sig([]types.Type{F}, I))
	r.RegisterWithSignature("HourOf", HourOf, CategoryDateTime, "Extracts hour from datetime",
		Sig([]types.Type{F}, I))
	r.RegisterWithSignature("MinuteOf", MinuteOf, CategoryDateTime, "Extracts minute from datetime",
		Sig([]types.Type{F}, I))
	r.RegisterWithSignature("SecondOf", SecondOf, CategoryDateTime, "Extracts second from datetime",
		Sig([]types.Type{F}, I))
	r.RegisterWithSignature("DayOfWeek", DayOfWeek, CategoryDateTime, "Returns day of week (0=Sunday)",
		Sig([]types.Type{F}, I))
	r.RegisterWithSignature("DayOfTheWeek", DayOfTheWeek, CategoryDateTime, "Returns day of week (1=Monday)",
		Sig([]types.Type{F}, I))
	r.RegisterWithSignature("DayOfYear", DayOfYear, CategoryDateTime, "Returns day of year (1-366)",
		Sig([]types.Type{F}, I))
	r.RegisterWithSignature("WeekNumber", WeekNumber, CategoryDateTime, "Returns ISO week number",
		Sig([]types.Type{F}, I))
	r.RegisterWithSignature("YearOfWeek", YearOfWeek, CategoryDateTime, "Returns year of ISO week",
		Sig([]types.Type{F}, I))
	r.RegisterWithSignature("IsLeapYear", IsLeapYear, CategoryDateTime, "Checks if year is a leap year",
		Sig([]types.Type{I}, B))
	r.RegisterWithSignature("FirstDayOfYear", FirstDayOfYear, CategoryDateTime, "Returns first day of year",
		Sig([]types.Type{F}, F))
	r.RegisterWithSignature("FirstDayOfNextYear", FirstDayOfNextYear, CategoryDateTime, "Returns first day of next year",
		Sig([]types.Type{F}, F))
	r.RegisterWithSignature("FirstDayOfMonth", FirstDayOfMonth, CategoryDateTime, "Returns first day of month",
		Sig([]types.Type{F}, F))
	r.RegisterWithSignature("FirstDayOfNextMonth", FirstDayOfNextMonth, CategoryDateTime, "Returns first day of next month",
		Sig([]types.Type{F}, F))
	r.RegisterWithSignature("FirstDayOfWeek", FirstDayOfWeek, CategoryDateTime, "Returns first day of ISO week",
		Sig([]types.Type{F}, F))
}

// RegisterConversionFunctions registers all type conversion built-in functions.
func RegisterConversionFunctions(r *Registry) {
	S := types.STRING
	I := types.INTEGER
	F := types.FLOAT
	B := types.BOOLEAN
	V := types.VARIANT
	r.RegisterWithSignature("IntToStr", IntToStr, CategoryConversion, "Converts integer to string",
		SigOptional([]types.Type{I, I}, S, 1)) // Optional base parameter
	r.RegisterWithSignature("IntToBin", IntToBin, CategoryConversion, "Converts integer to binary string",
		SigOptional([]types.Type{I, I}, S, 1)) // Optional digits parameter
	r.RegisterWithSignature("StrToInt", StrToInt, CategoryConversion, "Converts string to integer",
		Sig([]types.Type{S}, I))
	r.RegisterWithSignature("StrToFloat", StrToFloat, CategoryConversion, "Converts string to float",
		Sig([]types.Type{S}, F))
	r.RegisterWithSignature("FloatToStr", FloatToStr, CategoryConversion, "Converts float to string",
		SigOptional([]types.Type{F, I}, S, 1)) // Optional precision
	r.RegisterWithSignature("BoolToStr", BoolToStr, CategoryConversion, "Converts boolean to string",
		Sig([]types.Type{B}, S))

	// Hexadecimal conversion
	r.RegisterWithSignature("IntToHex", IntToHex, CategoryConversion, "Converts integer to hexadecimal string",
		Sig([]types.Type{I, I}, S))
	r.RegisterWithSignature("StrToBool", StrToBool, CategoryConversion, "Converts string to boolean",
		Sig([]types.Type{S}, B))
	r.RegisterWithSignature("Ord", Ord, CategoryConversion, "Returns ordinal value of enum/boolean/char",
		Sig([]types.Type{V}, I))
}

// RegisterEncodingFunctions registers all encoding/escaping built-in functions.
func RegisterEncodingFunctions(r *Registry) {
	S := types.STRING
	r.RegisterWithSignature("StrToHtml", StrToHtml, CategoryEncoding, "Encodes string for HTML content",
		Sig([]types.Type{S}, S))
	r.RegisterWithSignature("StrToHtmlAttribute", StrToHtmlAttribute, CategoryEncoding, "Encodes string for HTML attributes",
		Sig([]types.Type{S}, S))
	r.RegisterWithSignature("StrToJSON", StrToJSON, CategoryEncoding, "Encodes string for JSON",
		Sig([]types.Type{S}, S))
	r.RegisterWithSignature("StrToCSSText", StrToCSSText, CategoryEncoding, "Encodes string for CSS text",
		Sig([]types.Type{S}, S))
	r.RegisterWithSignature("StrToXML", StrToXML, CategoryEncoding, "Encodes string for XML",
		Sig([]types.Type{S}, S))
}

// RegisterJSONFunctions registers all JSON manipulation built-in functions.
func RegisterJSONFunctions(r *Registry) {
	S := types.STRING
	I := types.INTEGER
	B := types.BOOLEAN
	V := types.VARIANT

	r.RegisterWithSignature("ParseJSON", ParseJSON, CategoryJSON, "Parses JSON string to Variant",
		Sig([]types.Type{S}, V))
	r.RegisterWithSignature("ToJSON", ToJSON, CategoryJSON, "Converts value to compact JSON string",
		Sig([]types.Type{V}, S))
	r.RegisterWithSignature("ToJSONFormatted", ToJSONFormatted, CategoryJSON, "Converts value to formatted JSON string",
		SigOptional([]types.Type{V, S}, S, 1)) // Optional indent string
	r.RegisterWithSignature("JSONHasField", JSONHasField, CategoryJSON, "Checks if JSON object has field",
		Sig([]types.Type{V, S}, B))
	r.RegisterWithSignature("JSONKeys", JSONKeys, CategoryJSON, "Returns keys of JSON object",
		Sig([]types.Type{V}, V)) // Returns array of string
	r.RegisterWithSignature("JSONValues", JSONValues, CategoryJSON, "Returns values of JSON object/array",
		Sig([]types.Type{V}, V)) // Returns array
	r.RegisterWithSignature("JSONLength", JSONLength, CategoryJSON, "Returns length of JSON array/object",
		Sig([]types.Type{V}, I))
}

// RegisterTypeFunctions registers all type introspection built-in functions.
func RegisterTypeFunctions(r *Registry) {
	S := types.STRING
	V := types.VARIANT
	r.RegisterWithSignature("TypeOf", TypeOf, CategoryType, "Returns the type name of a value",
		Sig([]types.Type{V}, S))
	r.RegisterWithSignature("TypeOfClass", TypeOfClass, CategoryType, "Returns the class name of an object",
		Sig([]types.Type{V}, S))
}

// RegisterIOFunctions registers all I/O built-in functions.
func RegisterIOFunctions(r *Registry) {
	V := types.VARIANT
	r.RegisterWithSignature("Print", Print, CategoryIO, "Prints arguments without newline",
		SigVariadic([]types.Type{V}, nil, 0))
	r.RegisterWithSignature("PrintLn", PrintLn, CategoryIO, "Prints arguments with newline",
		SigVariadic([]types.Type{V}, nil, 0))
}

// RegisterVariantFunctions registers all Variant introspection and conversion built-in functions.
func RegisterVariantFunctions(r *Registry) {
	S := types.STRING
	I := types.INTEGER
	F := types.FLOAT
	B := types.BOOLEAN
	V := types.VARIANT

	// Variant type checking
	r.RegisterWithSignature("VarType", VarType, CategoryVariant, "Returns the type code of a Variant",
		Sig([]types.Type{V}, I))
	r.RegisterWithSignature("VarIsNull", VarIsNull, CategoryVariant, "Checks if Variant is unassigned",
		Sig([]types.Type{V}, B))
	r.RegisterWithSignature("VarIsEmpty", VarIsEmpty, CategoryVariant, "Checks if Variant is empty (alias for VarIsNull)",
		Sig([]types.Type{V}, B))
	r.RegisterWithSignature("VarIsClear", VarIsClear, CategoryVariant, "Checks if Variant is cleared (alias for VarIsNull)",
		Sig([]types.Type{V}, B))
	r.RegisterWithSignature("VarIsArray", VarIsArray, CategoryVariant, "Checks if Variant holds an array",
		Sig([]types.Type{V}, B))
	r.RegisterWithSignature("VarIsStr", VarIsStr, CategoryVariant, "Checks if Variant holds a string",
		Sig([]types.Type{V}, B))
	r.RegisterWithSignature("VarIsNumeric", VarIsNumeric, CategoryVariant, "Checks if Variant holds a numeric value",
		Sig([]types.Type{V}, B))

	// Variant conversion
	r.RegisterWithSignature("VarToStr", VarToStr, CategoryVariant, "Converts Variant to string",
		Sig([]types.Type{V}, S))
	r.RegisterWithSignature("VarToInt", VarToInt, CategoryVariant, "Converts Variant to integer",
		Sig([]types.Type{V}, I))
	r.RegisterWithSignature("VarToFloat", VarToFloat, CategoryVariant, "Converts Variant to float",
		Sig([]types.Type{V}, F))
	r.RegisterWithSignature("VarAsType", VarAsType, CategoryVariant, "Converts Variant to specified type code",
		Sig([]types.Type{V, I}, V))
	r.RegisterWithSignature("VarClear", VarClear, CategoryVariant, "Clears Variant to unassigned state",
		Sig([]types.Type{V}, nil)) // Procedure (modifies input)
}

// RegisterArrayFunctions registers all array built-in functions.
func RegisterArrayFunctions(r *Registry) {
	I := types.INTEGER
	B := types.BOOLEAN
	V := types.VARIANT

	r.RegisterWithSignature("Length", Length, CategoryArray, "Returns the number of elements in an array or characters in a string",
		Sig([]types.Type{V}, I))
	r.RegisterWithSignature("Copy", Copy, CategoryArray, "Creates a deep copy of an array or returns a substring",
		SigOptional([]types.Type{V, I, I}, V, 1))
	r.RegisterWithSignature("Low", Low, CategoryArray, "Returns the lower bound of an array or the lowest value of an enum/type",
		Sig([]types.Type{V}, I))
	r.RegisterWithSignature("High", High, CategoryArray, "Returns the upper bound of an array or the highest value of an enum/type",
		Sig([]types.Type{V}, I))
	r.RegisterWithSignature("IndexOf", IndexOf, CategoryArray, "Returns the index of the first occurrence of a value",
		SigOptional([]types.Type{V, V, I}, I, 2))
	r.RegisterWithSignature("Contains", Contains, CategoryArray, "Checks if an array contains a specific value",
		Sig([]types.Type{V, V}, B))
	r.RegisterWithSignature("Reverse", Reverse, CategoryArray, "Reverses the elements of an array in place",
		Sig([]types.Type{V}, nil)) // Procedure
	r.RegisterWithSignature("Sort", Sort, CategoryArray, "Sorts the elements of an array in place",
		SigOptional([]types.Type{V, V}, nil, 1)) // Optional comparator
	r.RegisterWithSignature("Add", Add, CategoryArray, "Appends an element to the end of a dynamic array",
		Sig([]types.Type{V, V}, I)) // Returns new length
	r.RegisterWithSignature("Delete", Delete, CategoryArray, "Removes an element at the specified index from a dynamic array",
		SigOptional([]types.Type{V, I, I}, nil, 2)) // Optional count
	r.RegisterWithSignature("SetLength", SetLength, CategoryArray, "Resizes a dynamic array or string to the specified length",
		Sig([]types.Type{V, I}, nil)) // Procedure
	r.RegisterWithSignature("Concat", Concat, CategoryArray, "Concatenates multiple strings or arrays into a new array/string",
		SigVariadic([]types.Type{V}, V, 1))
	r.RegisterWithSignature("Slice", Slice, CategoryArray, "Extracts a portion of an array",
		SigOptional([]types.Type{V, I, I}, V, 1))
}

// RegisterCollectionFunctions registers all collection (higher-order) built-in functions.
func RegisterCollectionFunctions(r *Registry) {
	I := types.INTEGER
	B := types.BOOLEAN
	V := types.VARIANT

	r.RegisterWithSignature("Map", Map, CategoryCollections, "Transforms each element of an array using a callback function",
		Sig([]types.Type{V, V}, V)) // (array, callback) -> array
	r.RegisterWithSignature("Filter", Filter, CategoryCollections, "Creates a new array containing only elements that match a predicate",
		Sig([]types.Type{V, V}, V)) // (array, predicate) -> array
	r.RegisterWithSignature("Reduce", Reduce, CategoryCollections, "Reduces an array to a single value using an accumulator function",
		SigOptional([]types.Type{V, V, V}, V, 2)) // (array, callback, initial?) -> value
	r.RegisterWithSignature("ForEach", ForEach, CategoryCollections, "Executes a function for each element of an array",
		Sig([]types.Type{V, V}, nil)) // (array, callback) -> void
	r.RegisterWithSignature("Every", Every, CategoryCollections, "Checks if all elements of an array match a predicate",
		Sig([]types.Type{V, V}, B)) // (array, predicate) -> boolean
	r.RegisterWithSignature("Some", Some, CategoryCollections, "Checks if any element of an array matches a predicate",
		Sig([]types.Type{V, V}, B)) // (array, predicate) -> boolean
	r.RegisterWithSignature("Find", Find, CategoryCollections, "Returns the first element that matches a predicate",
		Sig([]types.Type{V, V}, V)) // (array, predicate) -> element
	r.RegisterWithSignature("FindIndex", FindIndex, CategoryCollections, "Returns the index of the first element that matches a predicate",
		Sig([]types.Type{V, V}, I)) // (array, predicate) -> index
}

// RegisterSystemFunctions registers all system and miscellaneous built-in functions.
func RegisterSystemFunctions(r *Registry) {
	S := types.STRING
	I := types.INTEGER
	F := types.FLOAT
	B := types.BOOLEAN
	V := types.VARIANT

	// Debug/runtime introspection
	r.RegisterWithSignature("GetStackTrace", GetStackTrace, CategorySystem, "Returns a formatted string representation of the current call stack",
		Sig(nil, S))
	r.RegisterWithSignature("GetCallStack", GetCallStack, CategorySystem, "Returns the current call stack as an array of records",
		Sig(nil, V)) // Returns array of records
	r.RegisterWithSignature("Assigned", Assigned, CategorySystem, "Checks if a value is assigned (not nil)",
		Sig([]types.Type{V}, B))
	r.RegisterWithSignature("Assert", Assert, CategorySystem, "Validates a condition and raises EAssertionFailed if false",
		SigOptional([]types.Type{B, S}, nil, 1)) // (condition, message?) -> void

	// Type conversion
	r.RegisterWithSignature("Integer", Integer, CategoryConversion, "Converts a value to an integer",
		Sig([]types.Type{V}, I))
	r.RegisterWithSignature("StrToIntDef", StrToIntDef, CategoryConversion, "Converts a string to an integer with a default value",
		Sig([]types.Type{S, I}, I))
	r.RegisterWithSignature("StrToFloatDef", StrToFloatDef, CategoryConversion, "Converts a string to a float with a default value",
		Sig([]types.Type{S, F}, F))

	// String formatting
	r.RegisterWithSignature("Format", Format, CategoryString, "Formats a string using format specifiers",
		SigVariadic([]types.Type{S}, S, 1)) // (format, args...) -> string
}
