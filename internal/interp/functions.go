package interp

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// evalCallExpression evaluates a function call expression.
func (i *Interpreter) evalCallExpression(expr *ast.CallExpression) Value {
	// Check if this is a function pointer call
	// If the function expression is an identifier that resolves to a FunctionPointerValue,
	// we need to call through the pointer
	if funcIdent, ok := expr.Function.(*ast.Identifier); ok {
		// Try to resolve as a variable (might be a function pointer variable)
		if val, exists := i.env.Get(funcIdent.Value); exists {
			// Check if it's a function pointer
			if funcPtr, isFuncPtr := val.(*FunctionPointerValue); isFuncPtr {
				// Prepare arguments - check for lazy and var parameters in the function pointer's declaration
				args := make([]Value, len(expr.Arguments))
				for idx, arg := range expr.Arguments {
					// Check parameter flags (only for regular function pointers, not lambdas)
					isLazy := false
					isByRef := false
					if funcPtr.Function != nil && idx < len(funcPtr.Function.Parameters) {
						isLazy = funcPtr.Function.Parameters[idx].IsLazy
						isByRef = funcPtr.Function.Parameters[idx].ByRef
					}

					if isLazy {
						// For lazy parameters, create a LazyThunk
						args[idx] = NewLazyThunk(arg, i.env, i)
					} else if isByRef {
						// Task 9.35: For var parameters, create a reference or pass through existing reference
						if argIdent, ok := arg.(*ast.Identifier); ok {
							if val, exists := i.env.Get(argIdent.Value); exists {
								if refVal, isRef := val.(*ReferenceValue); isRef {
									args[idx] = refVal // Pass through existing reference
								} else {
									args[idx] = &ReferenceValue{Env: i.env, VarName: argIdent.Value}
								}
							} else {
								args[idx] = &ReferenceValue{Env: i.env, VarName: argIdent.Value}
							}
						} else {
							return i.newErrorWithLocation(arg, "var parameter requires a variable, got %T", arg)
						}
					} else {
						// For regular parameters, evaluate immediately
						argVal := i.Eval(arg)
						if isError(argVal) {
							return argVal
						}
						args[idx] = argVal
					}
				}
				// Call through the function pointer
				return i.callFunctionPointer(funcPtr, args, expr)
			}
		}
	}

	// Check if this is a unit-qualified function call (UnitName.FunctionName) or record method call
	if memberAccess, ok := expr.Function.(*ast.MemberAccessExpression); ok {
		// First, evaluate the object part to see what we're dealing with
		objVal := i.Eval(memberAccess.Object)
		if isError(objVal) {
			return objVal
		}

		// Task 9.7b: Check if this is a record method call (record.Method(...))
		// Implementation Strategy (Task 9.7a): Integrate with class method system (Option 2)
		// Records are value types, but methods work similarly to class methods.
		// Key difference: For mutating methods (procedures), we need copy-back semantics.
		if recVal, ok := objVal.(*RecordValue); ok {
			return i.evalRecordMethodCall(recVal, memberAccess, expr.Arguments, memberAccess.Object)
		}

		// Check if the left side is a unit identifier (for qualified access: UnitName.FunctionName)
		if unitIdent, ok := memberAccess.Object.(*ast.Identifier); ok {
			// This could be a unit-qualified call: UnitName.FunctionName()
			if i.unitRegistry != nil {
				if _, exists := i.unitRegistry.GetUnit(unitIdent.Value); exists {
					// Resolve the qualified function
					fn, err := i.ResolveQualifiedFunction(unitIdent.Value, memberAccess.Member.Value)
					if err == nil {
						// Prepare arguments - lazy parameters get LazyThunks, var parameters get References, regular parameters get evaluated
						args := make([]Value, len(expr.Arguments))
						for idx, arg := range expr.Arguments {
							// Check parameter flags
							isLazy := idx < len(fn.Parameters) && fn.Parameters[idx].IsLazy
							isByRef := idx < len(fn.Parameters) && fn.Parameters[idx].ByRef

							if isLazy {
								// For lazy parameters, create a LazyThunk
								args[idx] = NewLazyThunk(arg, i.env, i)
							} else if isByRef {
								// Task 9.35: For var parameters, create a reference or pass through existing reference
								if argIdent, ok := arg.(*ast.Identifier); ok {
									if val, exists := i.env.Get(argIdent.Value); exists {
										if refVal, isRef := val.(*ReferenceValue); isRef {
											args[idx] = refVal // Pass through existing reference
										} else {
											args[idx] = &ReferenceValue{Env: i.env, VarName: argIdent.Value}
										}
									} else {
										args[idx] = &ReferenceValue{Env: i.env, VarName: argIdent.Value}
									}
								} else {
									return i.newErrorWithLocation(arg, "var parameter requires a variable, got %T", arg)
								}
							} else {
								// For regular parameters, evaluate immediately
								val := i.Eval(arg)
								if isError(val) {
									return val
								}
								args[idx] = val
							}
						}
						return i.callUserFunction(fn, args)
					}
					// Function not found in unit
					return i.newErrorWithLocation(expr, "function '%s' not found in unit '%s'", memberAccess.Member.Value, unitIdent.Value)
				}
			}
		}
		// Not a unit-qualified call - could be a method call, let it fall through
		// to be handled as a method call on an object
		return i.newErrorWithLocation(expr, "cannot call member expression that is not a method or unit-qualified function")
	}

	// Get the function name
	funcName, ok := expr.Function.(*ast.Identifier)
	if !ok {
		return newError("function call requires identifier or qualified name, got %T", expr.Function)
	}

	// Check if it's a user-defined function first
	if fn, exists := i.functions[funcName.Value]; exists {
		// Prepare arguments - lazy parameters get LazyThunks, var parameters get References, regular parameters get evaluated
		args := make([]Value, len(expr.Arguments))
		for idx, arg := range expr.Arguments {
			// Check parameter flags
			isLazy := idx < len(fn.Parameters) && fn.Parameters[idx].IsLazy
			isByRef := idx < len(fn.Parameters) && fn.Parameters[idx].ByRef

			if isLazy {
				// For lazy parameters, create a LazyThunk with the unevaluated expression
				// and the current environment (captured from call site)
				args[idx] = NewLazyThunk(arg, i.env, i)
			} else if isByRef {
				// Task 9.35: For var parameters, create a reference to the variable
				// instead of copying its value
				if argIdent, ok := arg.(*ast.Identifier); ok {
					// Check if the variable is already a reference (var parameter passed through)
					if val, exists := i.env.Get(argIdent.Value); exists {
						if refVal, isRef := val.(*ReferenceValue); isRef {
							// Already a reference - pass it through
							args[idx] = refVal
						} else {
							// Regular variable - create a reference
							args[idx] = &ReferenceValue{
								Env:     i.env,
								VarName: argIdent.Value,
							}
						}
					} else {
						// Variable doesn't exist - create reference anyway (will error on access)
						args[idx] = &ReferenceValue{
							Env:     i.env,
							VarName: argIdent.Value,
						}
					}
				} else {
					// Var parameter must be a variable reference
					return i.newErrorWithLocation(arg, "var parameter requires a variable, got %T", arg)
				}
			} else {
				// For regular parameters, evaluate the expression immediately
				val := i.Eval(arg)
				if isError(val) {
					return val
				}
				args[idx] = val
			}
		}
		return i.callUserFunction(fn, args)
	}

	// Check if this is an instance method call within the current context (implicit Self)
	if selfVal, ok := i.env.Get("Self"); ok {
		if obj, isObj := AsObject(selfVal); isObj {
			if obj.GetMethod(funcName.Value) != nil {
				mc := &ast.MethodCallExpression{
					Token:     expr.Token,
					Object:    &ast.Identifier{Token: funcName.Token, Value: "Self"},
					Method:    funcName,
					Arguments: expr.Arguments,
				}
				return i.evalMethodCall(mc)
			}
		}
	}

	// Check if this is a built-in function with var parameters
	// These functions need the AST node for the first argument to modify it in place
	if funcName.Value == "Inc" || funcName.Value == "Dec" || funcName.Value == "Insert" ||
		(funcName.Value == "Delete" && len(expr.Arguments) == 3) ||
		funcName.Value == "DecodeDate" || funcName.Value == "DecodeTime" {
		return i.callBuiltinWithVarParam(funcName.Value, expr.Arguments)
	}

	// Task 9.2d: Check if this is an external function with var parameters
	// We need to check BEFORE evaluating args to create ReferenceValues
	if i.externalFunctions != nil {
		if extFunc, ok := i.externalFunctions.Get(funcName.Value); ok {
			varParams := extFunc.Wrapper.GetVarParams()

			// Prepare arguments - create ReferenceValues for var parameters
			args := make([]Value, len(expr.Arguments))
			for idx, arg := range expr.Arguments {
				isVarParam := idx < len(varParams) && varParams[idx]

				if isVarParam {
					// Task 9.2d: For var parameters, create a reference
					if argIdent, ok := arg.(*ast.Identifier); ok {
						if val, exists := i.env.Get(argIdent.Value); exists {
							if refVal, isRef := val.(*ReferenceValue); isRef {
								args[idx] = refVal // Pass through existing reference
							} else {
								args[idx] = &ReferenceValue{Env: i.env, VarName: argIdent.Value}
							}
						} else {
							args[idx] = &ReferenceValue{Env: i.env, VarName: argIdent.Value}
						}
					} else {
						return i.newErrorWithLocation(arg, "var parameter requires a variable, got %T", arg)
					}
				} else {
					// For regular parameters, evaluate immediately
					val := i.Eval(arg)
					if isError(val) {
						return val
					}
					args[idx] = val
				}
			}

			return i.callExternalFunction(extFunc, args)
		}
	}

	// Otherwise, try built-in functions
	// Evaluate all arguments
	args := make([]Value, len(expr.Arguments))
	for idx, arg := range expr.Arguments {
		val := i.Eval(arg)
		if isError(val) {
			return val
		}
		args[idx] = val
	}

	return i.callBuiltin(funcName.Value, args)
}

// normalizeBuiltinName normalizes a builtin function name to its canonical form
// for case-insensitive matching (DWScript is case-insensitive).
func normalizeBuiltinName(name string) string {
	// Create a lowercase version for comparison
	lower := strings.ToLower(name)

	// Map of lowercase names to canonical names
	// This allows case-insensitive function calls
	canonicalNames := map[string]string{
		"println": "PrintLn", "print": "Print", "ord": "Ord", "integer": "Integer",
		"length": "Length", "copy": "Copy", "concat": "Concat", "indexof": "IndexOf",
		"contains": "Contains", "reverse": "Reverse", "sort": "Sort", "pos": "Pos",
		"uppercase": "UpperCase", "lowercase": "LowerCase", "trim": "Trim",
		"trimleft": "TrimLeft", "trimright": "TrimRight", "stringreplace": "StringReplace",
		"stringofchar": "StringOfChar", "format": "Format", "abs": "Abs", "min": "Min", "max": "Max",
		"maxint": "MaxInt", "minint": "MinInt", "sqr": "Sqr", "power": "Power",
		"sqrt": "Sqrt", "sin": "Sin", "cos": "Cos", "tan": "Tan",
		"degtorad": "DegToRad", "radtodeg": "RadToDeg", "arcsin": "ArcSin",
		"arccos": "ArcCos", "arctan": "ArcTan", "arctan2": "ArcTan2",
		"cotan": "CoTan", "hypot": "Hypot", "sinh": "Sinh", "cosh": "Cosh",
		"tanh": "Tanh", "arcsinh": "ArcSinh", "arccosh": "ArcCosh", "arctanh": "ArcTanh",
		"random": "Random", "randomint": "RandomInt", "randomize": "Randomize",
		"unsigned32": "Unsigned32", "exp": "Exp", "ln": "Ln", "log2": "Log2",
		"round": "Round", "trunc": "Trunc", "ceil": "Ceil", "floor": "Floor",
		"low": "Low", "high": "High", "setlength": "SetLength", "add": "Add",
		"delete": "Delete", "inttostr": "IntToStr", "inttobin": "IntToBin",
		"strtoint": "StrToInt", "floattostr": "FloatToStr", "booltostr": "BoolToStr",
		"strtofloat": "StrToFloat", "inc": "Inc", "dec": "Dec", "succ": "Succ",
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
		"isleapyear":     "IsLeapYear",
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
	}

	// Return canonical name if found, otherwise return original name
	if canonical, ok := canonicalNames[lower]; ok {
		return canonical
	}
	return name
}

// callBuiltin calls a built-in function by name.
func (i *Interpreter) callBuiltin(name string, args []Value) Value {
	// Task 9.32: Check for external Go functions first
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
	case "StrToInt":
		return i.builtinStrToInt(args)
	case "FloatToStr":
		return i.builtinFloatToStr(args)
	case "StrToFloat":
		return i.builtinStrToFloat(args)
	case "BoolToStr":
		return i.builtinBoolToStr(args)
	case "Succ":
		return i.builtinSucc(args)
	case "Pred":
		return i.builtinPred(args)
	case "Assert":
		return i.builtinAssert(args)
	// Task 9.227: Higher-order functions for working with arrays and lambdas
	case "Map":
		return i.builtinMap(args)
	case "Filter":
		return i.builtinFilter(args)
	case "Reduce":
		return i.builtinReduce(args)
	case "ForEach":
		return i.builtinForEach(args)
	// Task 9.95-9.97: Current date/time functions
	case "Now":
		return i.builtinNow(args)
	case "Date":
		return i.builtinDate(args)
	case "Time":
		return i.builtinTime(args)
	case "UTCDateTime":
		return i.builtinUTCDateTime(args)
	// Task 9.99-9.101: Date encoding functions
	case "EncodeDate":
		return i.builtinEncodeDate(args)
	case "EncodeTime":
		return i.builtinEncodeTime(args)
	case "EncodeDateTime":
		return i.builtinEncodeDateTime(args)
	// Task 9.105: Component extraction functions
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
	// Task 9.107-9.109: Formatting functions
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
	// Task 9.110-9.111: Parsing functions
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
	// Task 9.113: Incrementing functions
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
	// Task 9.114: Date difference functions
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
	// Task 9.232: Variant introspection functions
	case "VarType":
		return i.builtinVarType(args)
	case "VarIsNull":
		return i.builtinVarIsNull(args)
	case "VarIsEmpty":
		return i.builtinVarIsEmpty(args)
	case "VarIsNumeric":
		return i.builtinVarIsNumeric(args)
	// Task 9.233: Variant conversion functions
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
	// Task 9.91: JSON parsing functions
	case "ParseJSON":
		return i.builtinParseJSON(args)
	// Task 9.94-9.95: JSON serialization functions
	case "ToJSON":
		return i.builtinToJSON(args)
	case "ToJSONFormatted":
		return i.builtinToJSONFormatted(args)
	// Task 9.98-9.100: JSON object access functions
	case "JSONHasField":
		return i.builtinJSONHasField(args)
	case "JSONKeys":
		return i.builtinJSONKeys(args)
	case "JSONValues":
		return i.builtinJSONValues(args)
	// Task 9.102: JSON array length function
	case "JSONLength":
		return i.builtinJSONLength(args)
	// Task 9.114: Exception Enhancements - GetStackTrace() built-in
	case "GetStackTrace":
		return i.builtinGetStackTrace(args)
	// Task 9.116: Debugging Information - GetCallStack() built-in
	case "GetCallStack":
		return i.builtinGetCallStack(args)
	default:
		return i.newErrorWithLocation(i.currentNode, "undefined function: %s", name)
	}
}

// isBuiltinFunction checks if a name refers to a built-in function.
// Task 9.132: Helper for checking parameterless built-in function calls.
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
// Task 9.132: Helper for calling parameterless built-in functions from identifier context.
func (i *Interpreter) callBuiltinFunction(name string, args []Value) Value {
	return i.callBuiltin(name, args)
}

// callExternalFunction calls an external Go function registered via FFI
// It uses the existing FFI error handling infrastructure to safely call the Go function
// and convert any errors or panics to DWScript exceptions.
func (i *Interpreter) callExternalFunction(extFunc *ExternalFunctionValue, args []Value) Value {
	// Use the existing callExternalFunctionSafe wrapper which handles panics
	// and converts them to EHost exceptions (from ffi_errors.go)
	return i.callExternalFunctionSafe(func() (Value, error) {
		// Call the wrapped Go function
		return extFunc.Wrapper.Call(args)
	})
}

// callBuiltinWithVarParam calls a built-in function that requires var parameters.
// These functions need access to the AST nodes to modify variables in place.
// Task 9.24: Support for Inc/Dec which need to modify the first argument.
// Task 9.43: Support for Insert which needs to modify the second argument.
// Task 9.44: Support for Delete (string mode) which needs to modify the first argument.
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
	// Task 9.103-9.104: Date decoding functions with var parameters
	case "DecodeDate":
		return i.builtinDecodeDate(args)
	case "DecodeTime":
		return i.builtinDecodeTime(args)
	default:
		return i.newErrorWithLocation(i.currentNode, "undefined var-param function: %s", name)
	}
}

// callUserFunction calls a user-defined function.
// It creates a new environment, binds parameters to arguments, executes the body,
// and extracts the return value from the Result variable or function name variable.
func (i *Interpreter) callUserFunction(fn *ast.FunctionDecl, args []Value) Value {
	// Task 9.1c: Calculate required parameter count (parameters without defaults)
	requiredParams := 0
	for _, param := range fn.Parameters {
		if param.DefaultValue == nil {
			requiredParams++
		}
	}

	// Check argument count is within valid range
	if len(args) < requiredParams {
		return newError("wrong number of arguments: expected at least %d, got %d",
			requiredParams, len(args))
	}
	if len(args) > len(fn.Parameters) {
		return newError("wrong number of arguments: expected at most %d, got %d",
			len(fn.Parameters), len(args))
	}

	// Task 9.1c: Fill in missing optional arguments with default values
	// Evaluate default expressions in the CALLER'S environment
	if len(args) < len(fn.Parameters) {
		savedEnv := i.env // Save caller's environment
		for idx := len(args); idx < len(fn.Parameters); idx++ {
			param := fn.Parameters[idx]
			if param.DefaultValue == nil {
				// This should never happen due to requiredParams check above
				return newError("internal error: missing required parameter at index %d", idx)
			}
			// Evaluate default value in caller's environment
			defaultVal := i.Eval(param.DefaultValue)
			if isError(defaultVal) {
				return defaultVal
			}
			args = append(args, defaultVal)
		}
		i.env = savedEnv // Restore environment (should be unchanged, but be safe)
	}

	// Create a new environment for the function scope
	funcEnv := NewEnclosedEnvironment(i.env)
	savedEnv := i.env
	i.env = funcEnv

	// Task 9.5: Check recursion depth before pushing to call stack
	if len(i.callStack) >= i.maxRecursionDepth {
		i.env = savedEnv // Restore environment before raising exception
		return i.raiseMaxRecursionExceeded()
	}

	// Push function name onto call stack for stack traces (Task 9.108)
	i.pushCallStack(fn.Name.Value)
	// Ensure it's popped when function exits (even if exception occurs)
	defer i.popCallStack()

	// Bind parameters to arguments
	for idx, param := range fn.Parameters {
		arg := args[idx]

		// Task 9.35: For var parameters, arg should already be a ReferenceValue
		// Don't apply implicit conversion to references - they need to stay as references
		if !param.ByRef {
			// Task 8.19b: Apply implicit conversion if parameter has a type and types don't match
			if param.Type != nil {
				paramTypeName := param.Type.Name
				if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
					arg = converted
				}
			}
		}

		// Store the argument in the function's environment
		// For var parameters, this will be a ReferenceValue
		// For regular parameters, this will be the actual value
		i.env.Define(param.Name.Value, arg)
	}

	// For functions (not procedures), initialize the Result variable
	if fn.ReturnType != nil {
		// Task 9.221: Initialize Result based on return type with appropriate defaults
		returnType := i.resolveTypeFromAnnotation(fn.ReturnType)
		var resultValue Value = i.getDefaultValue(returnType)

		// Check if return type is a record (overrides default)
		returnTypeName := fn.ReturnType.Name
		// Task 9.225: Normalize to lowercase for case-insensitive lookups
		recordTypeKey := "__record_type_" + strings.ToLower(returnTypeName)
		if typeVal, ok := i.env.Get(recordTypeKey); ok {
			if rtv, ok := typeVal.(*RecordTypeValue); ok {
				// Task 9.7e1: Use createRecordValue for proper nested record initialization
				resultValue = i.createRecordValue(rtv.RecordType, rtv.Methods)
			}
		}

		// Task 9.218: Check if return type is an array (overrides default)
		// Array return types should be initialized to empty arrays, not NIL
		// This allows methods like .Add() and .High to work on the Result variable
		// Task 9.225: Normalize to lowercase for case-insensitive lookups
		arrayTypeKey := "__array_type_" + strings.ToLower(returnTypeName)
		if typeVal, ok := i.env.Get(arrayTypeKey); ok {
			if atv, ok := typeVal.(*ArrayTypeValue); ok {
				resultValue = NewArrayValue(atv.ArrayType)
			}
		} else if strings.HasPrefix(returnTypeName, "array of ") || strings.HasPrefix(returnTypeName, "array[") {
			// Task 9.218: Handle inline array return types like "array of Integer"
			// For inline array types, create the array type directly from the type name
			elementTypeName := strings.TrimPrefix(returnTypeName, "array of ")
			if elementTypeName != returnTypeName {
				// Dynamic array: "array of Integer" -> elementTypeName = "Integer"
				elementType, err := i.resolveType(elementTypeName)
				if err == nil {
					arrayType := types.NewDynamicArrayType(elementType)
					resultValue = NewArrayValue(arrayType)
				}
			}
			// TODO: Handle static inline arrays like "array[1..10] of Integer"
			// For now, those should use named types
		}

		i.env.Define("Result", resultValue)
		// Also define the function name as an alias for Result
		// In DWScript, assigning to either Result or the function name sets the return value
		// We implement this by making the function name a reference to Result
		// This ensures that assigning to either one updates the same underlying value
		i.env.Define(fn.Name.Value, &ReferenceValue{Env: i.env, VarName: "Result"})
	}

	// Task 9.147: Check preconditions before executing function body
	if fn.PreConditions != nil {
		if err := i.checkPreconditions(fn.Name.Value, fn.PreConditions, i.env); err != nil {
			i.env = savedEnv
			return err
		}
	}

	// Task 9.146: Capture old values for postcondition evaluation
	oldValues := i.captureOldValues(fn, i.env)
	i.pushOldValues(oldValues)
	// Ensure old values are popped even if function exits early
	defer i.popOldValues()

	// Execute the function body
	if fn.Body == nil {
		// Function has no body (forward declaration) - this is an error
		i.env = savedEnv
		return newError("function '%s' has no body", fn.Name.Value)
	}

	i.Eval(fn.Body)

	// If an exception was raised during function execution, propagate it immediately
	if i.exception != nil {
		i.env = savedEnv
		return &NilValue{} // Return NilValue - actual value doesn't matter when exception is active
	}

	// Task 8.235n: Handle exit signal
	// If exit was called, clear the signal (don't propagate to caller)
	if i.exitSignal {
		i.exitSignal = false
		// Exit was called, function returns immediately with current Result value
	}

	// Extract return value
	var returnValue Value
	if fn.ReturnType != nil {
		// In DWScript, you can assign to either Result or the function name to set the return value
		// We implement the function name as a ReferenceValue pointing to Result
		// So we just need to get Result's value
		resultVal, resultOk := i.env.Get("Result")

		if resultOk {
			returnValue = resultVal
		} else {
			// Result not found (shouldn't happen)
			returnValue = &NilValue{}
		}

		// Task 8.19c: Apply implicit conversion if return type doesn't match
		if returnValue.Type() != "NIL" {
			expectedReturnType := fn.ReturnType.Name
			if converted, ok := i.tryImplicitConversion(returnValue, expectedReturnType); ok {
				returnValue = converted
			}
		}
	} else {
		// Procedure - no return value
		returnValue = &NilValue{}
	}

	// Task 9.149: Check postconditions after function body executes
	// Note: old values are available via oldValuesStack during postcondition evaluation
	if fn.PostConditions != nil {
		if err := i.checkPostconditions(fn.Name.Value, fn.PostConditions, i.env); err != nil {
			i.env = savedEnv
			return err
		}
	}

	// Restore the original environment
	i.env = savedEnv

	return returnValue
}

// callFunctionPointer calls a function through a function pointer.
// Task 9.166: Implement function pointer call execution.
//
// This handles both regular function pointers and method pointers.
// For method pointers, it binds the Self object before calling.
func (i *Interpreter) callFunctionPointer(funcPtr *FunctionPointerValue, args []Value, node ast.Node) Value {
	// Task 9.223: Enhanced to handle lambda closures

	// Check if this is a lambda or a regular function pointer
	if funcPtr.Lambda != nil {
		// Lambda closure - call with closure environment
		return i.callLambda(funcPtr.Lambda, funcPtr.Closure, args, node)
	}

	// Regular function pointer
	if funcPtr.Function == nil {
		return i.newErrorWithLocation(node, "function pointer is nil")
	}

	// If this is a method pointer, we need to set up the Self binding
	if funcPtr.SelfObject != nil {
		// Create a new environment with Self bound
		funcEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = funcEnv

		// Bind Self to the captured object
		i.env.Define("Self", funcPtr.SelfObject)

		// Call the function
		result := i.callUserFunction(funcPtr.Function, args)

		// Restore environment
		i.env = savedEnv

		return result
	}

	// Regular function pointer - just call the function directly
	return i.callUserFunction(funcPtr.Function, args)
}

// callLambda executes a lambda expression with its captured closure environment.
// Task 9.223: Closure invocation - executes lambda body with closure environment.
// Task 9.224: Variable capture - the closure environment provides reference semantics.
//
// The key difference from regular functions is that lambdas execute within their
// closure environment, allowing them to access captured variables from outer scopes.
//
// Parameters:
//   - lambda: The lambda expression AST node
//   - closureEnv: The environment captured when the lambda was created
//   - args: The argument values passed to the lambda
//   - node: AST node for error reporting
//
// Variable Capture Semantics:
//   - Captured variables are accessed by reference (not copied)
//   - Changes to captured variables inside the lambda affect the outer scope
//   - The environment chain naturally provides this behavior
func (i *Interpreter) callLambda(lambda *ast.LambdaExpression, closureEnv *Environment, args []Value, node ast.Node) Value {
	// Check argument count matches parameter count
	if len(args) != len(lambda.Parameters) {
		return i.newErrorWithLocation(node, "wrong number of arguments for lambda: expected %d, got %d",
			len(lambda.Parameters), len(args))
	}

	// Create a new environment for the lambda scope
	// CRITICAL: Use closureEnv as parent, NOT i.env
	// This gives the lambda access to captured variables
	lambdaEnv := NewEnclosedEnvironment(closureEnv)
	savedEnv := i.env
	i.env = lambdaEnv

	// Task 9.6: Check recursion depth before pushing to call stack
	if len(i.callStack) >= i.maxRecursionDepth {
		i.env = savedEnv // Restore environment before raising exception
		return i.raiseMaxRecursionExceeded()
	}

	// Push lambda marker onto call stack for stack traces (Task 9.108)
	i.pushCallStack("<lambda>")
	defer i.popCallStack()

	// Bind parameters to arguments
	for idx, param := range lambda.Parameters {
		arg := args[idx]

		// Apply implicit conversion if parameter has a type and types don't match
		if param.Type != nil {
			paramTypeName := param.Type.Name
			if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
				arg = converted
			}
		}

		// Note: Lambdas don't support by-ref parameters (for now)
		// All parameters are by-value
		i.env.Define(param.Name.Value, arg)
	}

	// For functions (not procedures), initialize the Result variable
	if lambda.ReturnType != nil {
		// Task 9.221: Initialize Result based on return type with appropriate defaults
		returnType := i.resolveTypeFromAnnotation(lambda.ReturnType)
		var resultValue Value = i.getDefaultValue(returnType)

		// Check if return type is a record (overrides default)
		returnTypeName := lambda.ReturnType.Name
		// Task 9.225: Normalize to lowercase for case-insensitive lookups
		recordTypeKey := "__record_type_" + strings.ToLower(returnTypeName)
		if typeVal, ok := i.env.Get(recordTypeKey); ok {
			if rtv, ok := typeVal.(*RecordTypeValue); ok {
				// Task 9.7e1: Use createRecordValue for proper nested record initialization
				resultValue = i.createRecordValue(rtv.RecordType, rtv.Methods)
			}
		}

		i.env.Define("Result", resultValue)
	}

	// Execute the lambda body
	bodyResult := i.Eval(lambda.Body)

	// If an error occurred during execution, propagate it
	if isError(bodyResult) {
		i.env = savedEnv
		return bodyResult
	}

	// If an exception was raised during lambda execution, propagate it immediately
	if i.exception != nil {
		i.env = savedEnv
		return &NilValue{}
	}

	// Handle exit signal
	if i.exitSignal {
		i.exitSignal = false
	}

	// Extract return value
	var returnValue Value
	if lambda.ReturnType != nil {
		// Lambda has a return type - get the Result value
		resultVal, resultOk := i.env.Get("Result")

		if resultOk && resultVal.Type() != "NIL" {
			returnValue = resultVal
		} else if resultOk {
			returnValue = resultVal
		} else {
			returnValue = &NilValue{}
		}
	} else {
		// Procedure lambda - no return value
		returnValue = &NilValue{}
	}

	// Restore environment
	i.env = savedEnv

	return returnValue
}

// evalRecordMethodCall evaluates a method call on a record value (record.Method(...)).
// Task 9.7b-9.7c: Implement record method invocation and resolution.
//
// Records are value types in DWScript (unlike classes which are reference types).
// This means:
//   - Methods execute with Self bound to the record instance
//   - For mutating methods (procedures that modify Self), we need copy-back semantics
//   - No inheritance - simple method lookup in RecordType.Methods
//
// Parameters:
//   - recVal: The record instance to call the method on
//   - memberAccess: The member access expression containing the method name
//   - argExprs: The argument expressions to evaluate
//
// Example:
//
//	type TPoint = record
//	  X, Y: Integer;
//	  function Distance: Float;
//	  begin
//	    Result := Sqrt(X*X + Y*Y);
//	  end;
//	  procedure Move(dx, dy: Integer);
//	  begin
//	    X := X + dx;
//	    Y := Y + dy;
//	  end;
//	end;
//
//	var p: TPoint;
//	p.X := 3; p.Y := 4;
//	var d := p.Distance();  // Returns 5.0
//	p.Move(1, 1);           // Modifies p to (4, 5)
func (i *Interpreter) evalRecordMethodCall(recVal *RecordValue, memberAccess *ast.MemberAccessExpression, argExprs []ast.Expression, objExpr ast.Expression) Value {
	methodName := memberAccess.Member.Value

	// Task 9.7c: Method resolution - lookup in RecordValue.Methods
	// No inheritance needed for records (unlike classes)
	if !recVal.HasMethod(methodName) {
		// Check if helpers provide this method
		helper, helperMethod, builtinSpec := i.findHelperMethod(recVal, methodName)
		if helperMethod == nil && builtinSpec == "" {
			if recVal.RecordType != nil {
				return i.newErrorWithLocation(memberAccess, "method '%s' not found in record type '%s' (no helper found)",
					methodName, recVal.RecordType.Name)
			}
			return i.newErrorWithLocation(memberAccess, "method '%s' not found (no helper found)", methodName)
		}

		// Evaluate method arguments
		args := make([]Value, len(argExprs))
		for idx, arg := range argExprs {
			val := i.Eval(arg)
			if isError(val) {
				return val
			}
			args[idx] = val
		}

		// Call the helper method
		return i.callHelperMethod(helper, helperMethod, builtinSpec, recVal, args, memberAccess)
	}

	method := recVal.GetMethod(methodName)
	if method == nil {
		return i.newErrorWithLocation(memberAccess, "method '%s' not found in record type '%s'",
			methodName, recVal.RecordType.Name)
	}

	// Evaluate method arguments
	args := make([]Value, len(argExprs))
	for idx, arg := range argExprs {
		val := i.Eval(arg)
		if isError(val) {
			return val
		}
		args[idx] = val
	}

	// Check argument count matches parameter count
	if len(args) != len(method.Parameters) {
		return i.newErrorWithLocation(memberAccess, "wrong number of arguments for method '%s': expected %d, got %d",
			methodName, len(method.Parameters), len(args))
	}

	// Create method environment with Self bound to the record
	// IMPORTANT: Records are value types, so we need to work with a copy
	// For mutating methods, we'll need to copy back changes to the original
	methodEnv := NewEnclosedEnvironment(i.env)
	savedEnv := i.env
	i.env = methodEnv

	// Make a copy of the record for the method execution
	// This implements value semantics - the method works on a copy
	recordCopy := recVal.Copy()

	// Bind Self to the record copy
	i.env.Define("Self", recordCopy)

	// Task 9.7e1: Bind all record fields to environment so they can be accessed directly
	// This allows code like "X := X + dx" to work without needing "Self.X"
	// Similar to how class property expressions bind fields (see objects.go:431-435)
	for fieldName, fieldValue := range recordCopy.Fields {
		i.env.Define(fieldName, fieldValue)
	}

	// Task 9.7: Check recursion depth before pushing to call stack
	if len(i.callStack) >= i.maxRecursionDepth {
		i.env = savedEnv // Restore environment before raising exception
		return i.raiseMaxRecursionExceeded()
	}

	// Push method name onto call stack for stack traces (Task 9.108)
	fullMethodName := recVal.RecordType.Name + "." + methodName
	i.pushCallStack(fullMethodName)
	defer i.popCallStack()

	// Bind method parameters to arguments with implicit conversion
	for idx, param := range method.Parameters {
		arg := args[idx]

		// Task 8.19b: Apply implicit conversion if parameter has a type and types don't match
		if param.Type != nil {
			paramTypeName := param.Type.Name
			if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
				arg = converted
			}
		}

		if param.ByRef {
			// By-reference parameter (TODO: implement proper by-ref support)
			i.env.Define(param.Name.Value, arg)
		} else {
			// By-value parameter
			i.env.Define(param.Name.Value, arg)
		}
	}

	// For functions (not procedures), initialize the Result variable
	if method.ReturnType != nil {
		// Task 9.221: Initialize Result based on return type with appropriate defaults
		returnType := i.resolveTypeFromAnnotation(method.ReturnType)
		var resultValue Value = i.getDefaultValue(returnType)

		// Check if return type is a record (overrides default)
		returnTypeName := method.ReturnType.Name
		// Task 9.225: Normalize to lowercase for case-insensitive lookups
		recordTypeKey := "__record_type_" + strings.ToLower(returnTypeName)
		if typeVal, ok := i.env.Get(recordTypeKey); ok {
			if rtv, ok := typeVal.(*RecordTypeValue); ok {
				// Task 9.7e1: Use createRecordValue for proper nested record initialization
				resultValue = i.createRecordValue(rtv.RecordType, rtv.Methods)
			}
		}

		i.env.Define("Result", resultValue)
		// Also define the method name as an alias for Result
		// In DWScript, assigning to either Result or the method name sets the return value
		i.env.Define(method.Name.Value, &ReferenceValue{Env: i.env, VarName: "Result"})
	}

	// Execute method body
	if method.Body == nil {
		i.env = savedEnv
		return i.newErrorWithLocation(memberAccess, "method '%s' has no body", methodName)
	}

	bodyResult := i.Eval(method.Body)

	// If an error occurred during execution, propagate it
	if isError(bodyResult) {
		i.env = savedEnv
		return bodyResult
	}

	// If an exception was raised during method execution, propagate it immediately
	if i.exception != nil {
		i.env = savedEnv
		return &NilValue{}
	}

	// Handle exit signal
	if i.exitSignal {
		i.exitSignal = false
	}

	// Extract return value
	var returnValue Value
	if method.ReturnType != nil {
		// Method has a return type - get the Result value
		resultVal, resultOk := i.env.Get("Result")
		methodNameVal, methodNameOk := i.env.Get(method.Name.Value)

		// Use whichever variable is not nil, preferring Result if both are set
		if resultOk && resultVal.Type() != "NIL" {
			returnValue = resultVal
		} else if methodNameOk && methodNameVal.Type() != "NIL" {
			returnValue = methodNameVal
		} else if resultOk {
			returnValue = resultVal
		} else if methodNameOk {
			returnValue = methodNameVal
		} else {
			returnValue = &NilValue{}
		}

		// Task 8.19c: Apply implicit conversion if return type doesn't match
		if returnValue.Type() != "NIL" {
			expectedReturnType := method.ReturnType.Name
			if converted, ok := i.tryImplicitConversion(returnValue, expectedReturnType); ok {
				returnValue = converted
			}
		}
	} else {
		// Procedure - no return value
		// But we need to handle copy-back for mutating procedures
		returnValue = &NilValue{}

		// Copy-back semantics for procedures:
		// If the method is a procedure (no return type), it may have modified Self.
		// We need to update the original record with the modified fields.
		// However, since we evaluated the object expression already, we can't directly
		// modify the original. This is a limitation of the current approach.
		//
		// TODO: For full copy-back semantics, we would need to:
		// 1. Track the lvalue (variable) that holds the record
		// 2. Update that variable with the modified record copy
		//
		// For now, we return the modified copy and rely on assignment handling.
	}

	// Task 9.7e1: Copy modified field values back from environment to record copy
	// This ensures that any field modifications made during method execution are preserved
	for fieldName := range recordCopy.Fields {
		if updatedVal, exists := i.env.Get(fieldName); exists {
			recordCopy.Fields[fieldName] = updatedVal
		}
	}

	// Restore environment
	i.env = savedEnv

	// Task 9.7e1: Update the original variable with the modified record copy
	// This implements proper value semantics for records - mutations persist
	// Check if the object expression is a simple identifier (variable)
	if ident, ok := objExpr.(*ast.Identifier); ok {
		// Update the variable in the environment with the modified copy
		// This makes mutations visible: p.SetCoords(10, 20) updates p
		i.env.Set(ident.Value, recordCopy)
	}

	return returnValue
}

// callRecordStaticMethod executes a static record method (class function/procedure).
// Task 9.7f: Static record methods don't have Self binding and are called on the type, not instances.
// Example: TPoint.Origin() where Origin is declared as "class function Origin: TPoint"
//
// Parameters:
//   - rtv: The RecordTypeValue containing the static method
//   - method: The FunctionDecl AST node for the static method
//   - argExprs: The argument expressions from the call site
//   - callNode: The call node for error reporting
//
// Static methods behave like regular functions but are scoped to the record type.
// They cannot access instance fields (no Self) but can return values of the record type.
func (i *Interpreter) callRecordStaticMethod(rtv *RecordTypeValue, method *ast.FunctionDecl, argExprs []ast.Expression, callNode ast.Node) Value {
	methodName := method.Name.Value

	// Evaluate method arguments
	args := make([]Value, len(argExprs))
	for idx, arg := range argExprs {
		val := i.Eval(arg)
		if isError(val) {
			return val
		}
		args[idx] = val
	}

	// Check argument count matches parameter count
	if len(args) != len(method.Parameters) {
		return i.newErrorWithLocation(callNode, "wrong number of arguments for static method '%s': expected %d, got %d",
			methodName, len(method.Parameters), len(args))
	}

	// Create method environment (NO Self binding for static methods)
	methodEnv := NewEnclosedEnvironment(i.env)
	savedEnv := i.env
	i.env = methodEnv

	// Task 9.8: Check recursion depth before pushing to call stack
	if len(i.callStack) >= i.maxRecursionDepth {
		i.env = savedEnv // Restore environment before raising exception
		return i.raiseMaxRecursionExceeded()
	}

	// Push method name onto call stack for stack traces (Task 9.108)
	fullMethodName := rtv.RecordType.Name + "." + methodName
	i.pushCallStack(fullMethodName)
	defer i.popCallStack()

	// Bind method parameters to arguments with implicit conversion
	for idx, param := range method.Parameters {
		arg := args[idx]

		// Apply implicit conversion if parameter has a type and types don't match
		if param.Type != nil {
			paramTypeName := param.Type.Name
			if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
				arg = converted
			}
		}

		i.env.Define(param.Name.Value, arg)
	}

	// For functions (not procedures), initialize the Result variable
	if method.ReturnType != nil {
		// Task 9.221: Initialize Result based on return type with appropriate defaults
		returnType := i.resolveTypeFromAnnotation(method.ReturnType)
		var resultValue Value = i.getDefaultValue(returnType)

		// Check if return type is a record (overrides default)
		// Task 9.7f: For static record methods returning records, initialize Result properly
		returnTypeName := method.ReturnType.Name
		// Task 9.225: Normalize to lowercase for case-insensitive lookups
		recordTypeKey := "__record_type_" + strings.ToLower(returnTypeName)
		if typeVal, ok := i.env.Get(recordTypeKey); ok {
			if recordTV, ok := typeVal.(*RecordTypeValue); ok {
				// Return type is a record - create an instance
				resultValue = i.createRecordValue(recordTV.RecordType, recordTV.Methods)
			}
		}

		i.env.Define("Result", resultValue)
		// Also define the method name as an alias for Result
		// In DWScript, assigning to either Result or the method name sets the return value
		i.env.Define(methodName, &ReferenceValue{Env: i.env, VarName: "Result"})
	}

	// Execute method body
	result := i.Eval(method.Body)
	if isError(result) {
		i.env = savedEnv
		return result
	}

	// Extract return value (same logic as class methods)
	var returnValue Value
	if method.ReturnType != nil {
		// Check both Result and method name variable
		resultVal, resultOk := i.env.Get("Result")
		methodNameVal, methodNameOk := i.env.Get(methodName)

		// Use whichever variable is not nil, preferring Result if both are set
		if resultOk && resultVal.Type() != "NIL" {
			returnValue = resultVal
		} else if methodNameOk && methodNameVal.Type() != "NIL" {
			returnValue = methodNameVal
		} else if resultOk {
			returnValue = resultVal
		} else if methodNameOk {
			returnValue = methodNameVal
		} else {
			returnValue = &NilValue{}
		}

		// Apply implicit conversion if return type doesn't match
		if returnValue.Type() != "NIL" {
			expectedReturnType := method.ReturnType.Name
			if converted, ok := i.tryImplicitConversion(returnValue, expectedReturnType); ok {
				returnValue = converted
			}
		}
	} else {
		// Procedure - no return value
		returnValue = &NilValue{}
	}

	// Restore environment
	i.env = savedEnv

	return returnValue
}

// parseInlineArrayType parses an inline array type signature and creates an ArrayType.
// Task 9.56: Support for inline array type initialization in variable declarations.
//
// Examples:
//   - "array of Integer" -> DynamicArrayType
//   - "array[1..10] of String" -> StaticArrayType with bounds
//   - "array of array of Integer" -> Nested dynamic arrays
func (i *Interpreter) parseInlineArrayType(signature string) *types.ArrayType {
	var lowBound, highBound *int

	// Check if this is a static array with bounds
	if strings.HasPrefix(signature, "array[") {
		// Extract bounds: array[low..high] of Type
		endBracket := strings.Index(signature, "]")
		if endBracket == -1 {
			return nil
		}

		boundsStr := signature[6:endBracket] // Skip "array["
		parts := strings.Split(boundsStr, "..")
		if len(parts) != 2 {
			return nil
		}

		// Parse low bound
		low := 0
		if _, err := fmt.Sscanf(parts[0], "%d", &low); err != nil {
			return nil
		}
		lowBound = &low

		// Parse high bound
		high := 0
		if _, err := fmt.Sscanf(parts[1], "%d", &high); err != nil {
			return nil
		}
		highBound = &high

		// Skip past "] of "
		signature = signature[endBracket+1:]
	} else if strings.HasPrefix(signature, "array of ") {
		// Dynamic array: skip "array" to get " of ElementType"
		signature = signature[5:] // Skip "array"
	} else {
		return nil
	}

	// Now signature should be " of ElementType"
	if !strings.HasPrefix(signature, " of ") {
		return nil
	}

	// Extract element type name
	elementTypeName := strings.TrimSpace(signature[4:]) // Skip " of "

	// Get the element type (resolveType handles recursion for nested arrays)
	elementType, err := i.resolveType(elementTypeName)
	if err != nil || elementType == nil {
		return nil
	}

	// Create array type
	if lowBound != nil && highBound != nil {
		return types.NewStaticArrayType(elementType, *lowBound, *highBound)
	}
	return types.NewDynamicArrayType(elementType)
}

// parseInlineSetType parses inline set type syntax like "set of TEnumType".
// Returns the SetType, or nil if the string doesn't match the expected format.
// Task 9.214: Support set variable initialization
func (i *Interpreter) parseInlineSetType(signature string) *types.SetType {
	// Check for "set of " prefix
	if !strings.HasPrefix(signature, "set of ") {
		return nil
	}

	// Extract enum type name: "set of TColor" → "TColor"
	enumTypeName := strings.TrimSpace(signature[7:]) // Skip "set of "
	if enumTypeName == "" {
		return nil
	}

	// Look up the enum type in the environment
	// Enum types are stored with "__enum_type_" prefix
	// Task 9.225: Normalize to lowercase for case-insensitive lookups
	typeKey := "__enum_type_" + strings.ToLower(enumTypeName)
	typeVal, ok := i.env.Get(typeKey)
	if !ok {
		return nil
	}

	// Extract the EnumType from the EnumTypeValue
	enumTypeVal, ok := typeVal.(*EnumTypeValue)
	if !ok {
		return nil
	}

	// Create and return the set type
	return types.NewSetType(enumTypeVal.EnumType)
}

// resolveArrayTypeNode resolves an ArrayTypeNode directly from the AST.
// This avoids string conversion issues with parentheses in bound expressions like (-5).
// Task: Fix negative array bounds like array[-5..5]
func (i *Interpreter) resolveArrayTypeNode(arrayNode *ast.ArrayTypeNode) *types.ArrayType {
	if arrayNode == nil {
		return nil
	}

	// Resolve element type first
	var elementType types.Type

	// Check if element type is also an array (nested arrays)
	if nestedArray, ok := arrayNode.ElementType.(*ast.ArrayTypeNode); ok {
		elementType = i.resolveArrayTypeNode(nestedArray)
		if elementType == nil {
			return nil
		}
	} else {
		// Get element type name and resolve it
		var elementTypeName string
		if typeAnnot, ok := arrayNode.ElementType.(*ast.TypeAnnotation); ok {
			elementTypeName = typeAnnot.Name
		} else {
			elementTypeName = arrayNode.ElementType.String()
		}

		var err error
		elementType, err = i.resolveType(elementTypeName)
		if err != nil || elementType == nil {
			return nil
		}
	}

	// Check if dynamic or static array
	if arrayNode.IsDynamic() {
		return types.NewDynamicArrayType(elementType)
	}

	// Static array - evaluate bounds by interpreting the expressions
	// For constant expressions (literals, unary minus), we can evaluate them directly
	lowVal := i.Eval(arrayNode.LowBound)
	if isError(lowVal) {
		return nil
	}
	lowBound, ok := lowVal.(*IntegerValue)
	if !ok {
		return nil
	}

	highVal := i.Eval(arrayNode.HighBound)
	if isError(highVal) {
		return nil
	}
	highBound, ok := highVal.(*IntegerValue)
	if !ok {
		return nil
	}

	return types.NewStaticArrayType(elementType, int(lowBound.Value), int(highBound.Value))
}
