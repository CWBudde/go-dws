package builtins

// DefaultRegistry is the default global registry of all built-in functions.
// It's populated on package initialization with all standard DWScript built-ins.
var DefaultRegistry *Registry

func init() {
	DefaultRegistry = NewRegistry()
	RegisterAll(DefaultRegistry)
}

// RegisterAll registers all built-in functions with the given registry.
// This allows for creating custom registries with a different set of functions.
func RegisterAll(r *Registry) {
	RegisterMathFunctions(r)
	RegisterStringFunctions(r)
	RegisterDateTimeFunctions(r)
	RegisterConversionFunctions(r)
}

// RegisterMathFunctions registers all mathematical built-in functions.
func RegisterMathFunctions(r *Registry) {
	// Basic math functions
	r.Register("Abs", Abs, CategoryMath, "Returns the absolute value of a number")
	r.Register("Min", Min, CategoryMath, "Returns the minimum of two numbers")
	r.Register("Max", Max, CategoryMath, "Returns the maximum of two numbers")
	r.Register("ClampInt", ClampInt, CategoryMath, "Clamps an integer value between min and max")
	r.Register("Clamp", Clamp, CategoryMath, "Clamps a value between min and max")
	r.Register("Sqr", Sqr, CategoryMath, "Returns the square of a number")
	r.Register("Power", Power, CategoryMath, "Returns base raised to the power of exponent")
	r.Register("Sqrt", Sqrt, CategoryMath, "Returns the square root of a number")
	r.Register("Pi", Pi, CategoryMath, "Returns the value of π (pi)")
	r.Register("Sign", Sign, CategoryMath, "Returns the sign of a number (-1, 0, or 1)")
	r.Register("Odd", Odd, CategoryMath, "Returns true if the number is odd")
	r.Register("Frac", Frac, CategoryMath, "Returns the fractional part of a float")
	r.Register("Int", Int, CategoryMath, "Returns the integer part of a float")
	r.Register("Round", Round, CategoryMath, "Rounds a float to the nearest integer")
	r.Register("Trunc", Trunc, CategoryMath, "Truncates a float to an integer")
	r.Register("Ceil", Ceil, CategoryMath, "Returns the ceiling (smallest integer >= value)")
	r.Register("Floor", Floor, CategoryMath, "Returns the floor (largest integer <= value)")
	r.Register("Unsigned32", Unsigned32, CategoryMath, "Converts a signed integer to unsigned 32-bit")
	r.Register("MaxInt", MaxInt, CategoryMath, "Returns the maximum value of multiple integers")
	r.Register("MinInt", MinInt, CategoryMath, "Returns the minimum value of multiple integers")

	// Advanced math functions
	r.Register("Factorial", Factorial, CategoryMath, "Returns the factorial of n")
	r.Register("Gcd", Gcd, CategoryMath, "Returns the greatest common divisor")
	r.Register("Lcm", Lcm, CategoryMath, "Returns the least common multiple")
	r.Register("IsPrime", IsPrime, CategoryMath, "Returns true if n is a prime number")
	r.Register("LeastFactor", LeastFactor, CategoryMath, "Returns the smallest prime factor")
	r.Register("PopCount", PopCount, CategoryMath, "Returns the number of set bits")
	r.Register("TestBit", TestBit, CategoryMath, "Tests if a specific bit is set")
	r.Register("Haversine", Haversine, CategoryMath, "Calculates the haversine distance")
	r.Register("CompareNum", CompareNum, CategoryMath, "Compares two numbers (-1, 0, 1)")

	// Exponential and logarithmic functions
	r.Register("Exp", Exp, CategoryMath, "Returns e raised to the power of x")
	r.Register("Ln", Ln, CategoryMath, "Returns the natural logarithm")
	r.Register("Log2", Log2, CategoryMath, "Returns the base-2 logarithm")
	r.Register("Log10", Log10, CategoryMath, "Returns the base-10 logarithm")
	r.Register("LogN", LogN, CategoryMath, "Returns the logarithm with custom base")
	r.Register("IntPower", IntPower, CategoryMath, "Returns base raised to integer exponent")

	// Special values
	r.Register("Infinity", Infinity, CategoryMath, "Returns positive infinity")
	r.Register("NaN", NaN, CategoryMath, "Returns NaN (Not a Number)")
	r.Register("IsFinite", IsFinite, CategoryMath, "Returns true if value is finite")
	r.Register("IsInfinite", IsInfinite, CategoryMath, "Returns true if value is infinite")
	r.Register("IsNaN", IsNaN, CategoryMath, "Returns true if value is NaN")

	// Trigonometric functions
	r.Register("Sin", Sin, CategoryMath, "Returns the sine of x (radians)")
	r.Register("Cos", Cos, CategoryMath, "Returns the cosine of x (radians)")
	r.Register("Tan", Tan, CategoryMath, "Returns the tangent of x (radians)")
	r.Register("CoTan", CoTan, CategoryMath, "Returns the cotangent of x (radians)")
	r.Register("ArcSin", ArcSin, CategoryMath, "Returns the arcsine of x")
	r.Register("ArcCos", ArcCos, CategoryMath, "Returns the arccosine of x")
	r.Register("ArcTan", ArcTan, CategoryMath, "Returns the arctangent of x")
	r.Register("ArcTan2", ArcTan2, CategoryMath, "Returns the arctangent of y/x")
	r.Register("DegToRad", DegToRad, CategoryMath, "Converts degrees to radians")
	r.Register("RadToDeg", RadToDeg, CategoryMath, "Converts radians to degrees")
	r.Register("Hypot", Hypot, CategoryMath, "Returns the hypotenuse (sqrt(x²+y²))")

	// Hyperbolic functions
	r.Register("Sinh", Sinh, CategoryMath, "Returns the hyperbolic sine")
	r.Register("Cosh", Cosh, CategoryMath, "Returns the hyperbolic cosine")
	r.Register("Tanh", Tanh, CategoryMath, "Returns the hyperbolic tangent")
	r.Register("ArcSinh", ArcSinh, CategoryMath, "Returns the inverse hyperbolic sine")
	r.Register("ArcCosh", ArcCosh, CategoryMath, "Returns the inverse hyperbolic cosine")
	r.Register("ArcTanh", ArcTanh, CategoryMath, "Returns the inverse hyperbolic tangent")

	// Random number functions (TODO: Currently implemented in Interpreter, not yet migrated)
	// r.Register("Random", Random, CategoryMath, "Returns a random float between 0 and 1")
	// r.Register("RandomInt", RandomInt, CategoryMath, "Returns a random integer in range")
	// r.Register("Randomize", Randomize, CategoryMath, "Seeds the random number generator")
	// r.Register("SetRandSeed", SetRandSeed, CategoryMath, "Sets the random number seed")
	// r.Register("RandSeed", RandSeed, CategoryMath, "Returns the current random seed")
	// r.Register("RandG", RandG, CategoryMath, "Returns a random Gaussian value")
}

// RegisterStringFunctions registers all string manipulation built-in functions.
func RegisterStringFunctions(r *Registry) {
	// Basic string functions
	r.Register("Pos", Pos, CategoryString, "Finds the position of a substring")
	r.Register("UpperCase", UpperCase, CategoryString, "Converts string to uppercase")
	r.Register("LowerCase", LowerCase, CategoryString, "Converts string to lowercase")
	r.Register("ASCIIUpperCase", ASCIIUpperCase, CategoryString, "Converts ASCII characters to uppercase")
	r.Register("ASCIILowerCase", ASCIILowerCase, CategoryString, "Converts ASCII characters to lowercase")
	r.Register("AnsiUpperCase", AnsiUpperCase, CategoryString, "Converts ANSI string to uppercase")
	r.Register("AnsiLowerCase", AnsiLowerCase, CategoryString, "Converts ANSI string to lowercase")
	r.Register("Trim", Trim, CategoryString, "Removes leading and trailing whitespace")
	r.Register("TrimLeft", TrimLeft, CategoryString, "Removes leading whitespace")
	r.Register("TrimRight", TrimRight, CategoryString, "Removes trailing whitespace")
	r.Register("StringReplace", StringReplace, CategoryString, "Replaces occurrences of a substring")
	r.Register("StringOfChar", StringOfChar, CategoryString, "Creates a string of repeated characters")
	r.Register("SubStr", SubStr, CategoryString, "Extracts a substring")
	r.Register("SubString", SubString, CategoryString, "Extracts a substring (alias)")
	r.Register("LeftStr", LeftStr, CategoryString, "Returns leftmost characters")
	r.Register("RightStr", RightStr, CategoryString, "Returns rightmost characters")
	r.Register("MidStr", MidStr, CategoryString, "Extracts middle substring")
	r.Register("Chr", Chr, CategoryString, "Converts character code to string")

	// String search functions
	r.Register("StrBeginsWith", StrBeginsWith, CategoryString, "Checks if string starts with prefix")
	r.Register("StrEndsWith", StrEndsWith, CategoryString, "Checks if string ends with suffix")
	r.Register("StrContains", StrContains, CategoryString, "Checks if string contains substring")
	r.Register("PosEx", PosEx, CategoryString, "Finds position with start index")
	r.Register("RevPos", RevPos, CategoryString, "Finds last position of substring")
	r.Register("StrFind", StrFind, CategoryString, "Finds substring in string")

	// Advanced string functions
	r.Register("StrBefore", StrBefore, CategoryString, "Returns text before delimiter")
	r.Register("StrBeforeLast", StrBeforeLast, CategoryString, "Returns text before last delimiter")
	r.Register("StrAfter", StrAfter, CategoryString, "Returns text after delimiter")
	r.Register("StrAfterLast", StrAfterLast, CategoryString, "Returns text after last delimiter")
	r.Register("StrBetween", StrBetween, CategoryString, "Extracts text between delimiters")
	r.Register("IsDelimiter", IsDelimiter, CategoryString, "Checks if character is a delimiter")
	r.Register("LastDelimiter", LastDelimiter, CategoryString, "Finds last delimiter position")
	r.Register("FindDelimiter", FindDelimiter, CategoryString, "Finds first delimiter position")
	r.Register("PadLeft", PadLeft, CategoryString, "Pads string on the left")
	r.Register("PadRight", PadRight, CategoryString, "Pads string on the right")
	r.Register("StrDeleteLeft", StrDeleteLeft, CategoryString, "Deletes characters from left")
	r.Register("DeleteLeft", StrDeleteLeft, CategoryString, "Deletes characters from left (alias)")
	r.Register("StrDeleteRight", StrDeleteRight, CategoryString, "Deletes characters from right")
	r.Register("DeleteRight", StrDeleteRight, CategoryString, "Deletes characters from right (alias)")
	r.Register("ReverseString", ReverseString, CategoryString, "Reverses a string")
	r.Register("QuotedStr", QuotedStr, CategoryString, "Returns quoted string")
	r.Register("StringOfString", StringOfString, CategoryString, "Repeats a string")
	r.Register("DupeString", DupeString, CategoryString, "Duplicates a string")
	r.Register("NormalizeString", NormalizeString, CategoryString, "Normalizes Unicode string")
	r.Register("Normalize", NormalizeString, CategoryString, "Normalizes Unicode string (alias)")
	r.Register("StripAccents", StripAccents, CategoryString, "Removes accents from characters")
	r.Register("ByteSizeToStr", ByteSizeToStr, CategoryString, "Formats byte size as human-readable string")
	r.Register("GetText", GetText, CategoryString, "Localizes text (i18n)")
	r.Register("CharAt", CharAt, CategoryString, "Returns character at index")
	r.Register("Underscore", Underscore, CategoryString, "Converts string to underscore_case")

	// String comparison functions
	r.Register("SameText", SameText, CategoryString, "Case-insensitive string equality")
	r.Register("CompareText", CompareText, CategoryString, "Case-insensitive string comparison")
	r.Register("CompareStr", CompareStr, CategoryString, "Case-sensitive string comparison")
	r.Register("AnsiCompareText", AnsiCompareText, CategoryString, "ANSI case-insensitive comparison")
	r.Register("AnsiCompareStr", AnsiCompareStr, CategoryString, "ANSI case-sensitive comparison")
	r.Register("CompareLocaleStr", CompareLocaleStr, CategoryString, "Locale-aware string comparison")
	r.Register("StrMatches", StrMatches, CategoryString, "Tests if string matches pattern")
	r.Register("StrIsASCII", StrIsASCII, CategoryString, "Checks if string is ASCII only")
}

// RegisterDateTimeFunctions registers all date/time built-in functions.
func RegisterDateTimeFunctions(r *Registry) {
	// Date/time creation
	r.Register("EncodeDate", EncodeDate, CategoryDateTime, "Creates date from year, month, day")
	r.Register("EncodeTime", EncodeTime, CategoryDateTime, "Creates time from hour, minute, second")
	r.Register("EncodeDateTime", EncodeDateTime, CategoryDateTime, "Creates datetime from components")
	r.Register("Now", Now, CategoryDateTime, "Returns current date and time")
	r.Register("Date", Date, CategoryDateTime, "Returns current date")
	r.Register("Time", Time, CategoryDateTime, "Returns current time")
	r.Register("UTCDateTime", UTCDateTime, CategoryDateTime, "Returns current UTC datetime")

	// Date/time arithmetic
	r.Register("IncYear", IncYear, CategoryDateTime, "Adds years to a date")
	r.Register("IncMonth", IncMonth, CategoryDateTime, "Adds months to a date")
	r.Register("IncDay", IncDay, CategoryDateTime, "Adds days to a date")
	r.Register("IncHour", IncHour, CategoryDateTime, "Adds hours to a datetime")
	r.Register("IncMinute", IncMinute, CategoryDateTime, "Adds minutes to a datetime")
	r.Register("IncSecond", IncSecond, CategoryDateTime, "Adds seconds to a datetime")
	r.Register("DaysBetween", DaysBetween, CategoryDateTime, "Returns days between two dates")
	r.Register("HoursBetween", HoursBetween, CategoryDateTime, "Returns hours between two datetimes")
	r.Register("MinutesBetween", MinutesBetween, CategoryDateTime, "Returns minutes between two datetimes")
	r.Register("SecondsBetween", SecondsBetween, CategoryDateTime, "Returns seconds between two datetimes")

	// Date/time formatting
	r.Register("FormatDateTime", FormatDateTime, CategoryDateTime, "Formats datetime with format string")
	r.Register("DateTimeToStr", DateTimeToStr, CategoryDateTime, "Converts datetime to string")
	r.Register("DateToStr", DateToStr, CategoryDateTime, "Converts date to string")
	r.Register("TimeToStr", TimeToStr, CategoryDateTime, "Converts time to string")
	r.Register("DateToISO8601", DateToISO8601, CategoryDateTime, "Converts date to ISO8601 format")
	r.Register("DateTimeToISO8601", DateTimeToISO8601, CategoryDateTime, "Converts datetime to ISO8601 format")
	r.Register("DateTimeToRFC822", DateTimeToRFC822, CategoryDateTime, "Converts datetime to RFC822 format")

	// Date/time parsing
	r.Register("StrToDate", StrToDate, CategoryDateTime, "Parses string to date")
	r.Register("StrToDateTime", StrToDateTime, CategoryDateTime, "Parses string to datetime")
	r.Register("StrToTime", StrToTime, CategoryDateTime, "Parses string to time")
	r.Register("ISO8601ToDateTime", ISO8601ToDateTime, CategoryDateTime, "Parses ISO8601 to datetime")
	r.Register("RFC822ToDateTime", RFC822ToDateTime, CategoryDateTime, "Parses RFC822 to datetime")

	// Unix time conversions
	r.Register("UnixTime", UnixTime, CategoryDateTime, "Returns current Unix timestamp")
	r.Register("UnixTimeMSec", UnixTimeMSec, CategoryDateTime, "Returns current Unix timestamp in milliseconds")
	r.Register("UnixTimeToDateTime", UnixTimeToDateTime, CategoryDateTime, "Converts Unix timestamp to datetime")
	r.Register("DateTimeToUnixTime", DateTimeToUnixTime, CategoryDateTime, "Converts datetime to Unix timestamp")
	r.Register("UnixTimeMSecToDateTime", UnixTimeMSecToDateTime, CategoryDateTime, "Converts Unix milliseconds to datetime")
	r.Register("DateTimeToUnixTimeMSec", DateTimeToUnixTimeMSec, CategoryDateTime, "Converts datetime to Unix milliseconds")

	// Date/time information
	r.Register("YearOf", YearOf, CategoryDateTime, "Extracts year from datetime")
	r.Register("MonthOf", MonthOf, CategoryDateTime, "Extracts month from datetime")
	r.Register("DayOf", DayOf, CategoryDateTime, "Extracts day from datetime")
	r.Register("HourOf", HourOf, CategoryDateTime, "Extracts hour from datetime")
	r.Register("MinuteOf", MinuteOf, CategoryDateTime, "Extracts minute from datetime")
	r.Register("SecondOf", SecondOf, CategoryDateTime, "Extracts second from datetime")
	r.Register("DayOfWeek", DayOfWeek, CategoryDateTime, "Returns day of week (0=Sunday)")
	r.Register("DayOfTheWeek", DayOfTheWeek, CategoryDateTime, "Returns day of week (1=Monday)")
	r.Register("DayOfYear", DayOfYear, CategoryDateTime, "Returns day of year (1-366)")
	r.Register("WeekNumber", WeekNumber, CategoryDateTime, "Returns ISO week number")
	r.Register("YearOfWeek", YearOfWeek, CategoryDateTime, "Returns year of ISO week")
	r.Register("IsLeapYear", IsLeapYear, CategoryDateTime, "Checks if year is a leap year")
	r.Register("FirstDayOfYear", FirstDayOfYear, CategoryDateTime, "Returns first day of year")
	r.Register("FirstDayOfNextYear", FirstDayOfNextYear, CategoryDateTime, "Returns first day of next year")
	r.Register("FirstDayOfMonth", FirstDayOfMonth, CategoryDateTime, "Returns first day of month")
	r.Register("FirstDayOfNextMonth", FirstDayOfNextMonth, CategoryDateTime, "Returns first day of next month")
	r.Register("FirstDayOfWeek", FirstDayOfWeek, CategoryDateTime, "Returns first day of ISO week")
}

// RegisterConversionFunctions registers all type conversion built-in functions.
func RegisterConversionFunctions(r *Registry) {
	r.Register("IntToHex", IntToHex, CategoryConversion, "Converts integer to hexadecimal string")
	r.Register("StrToBool", StrToBool, CategoryConversion, "Converts string to boolean")
}
