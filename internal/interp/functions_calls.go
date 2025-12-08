package interp

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	pkgident "github.com/cwbudde/go-dws/pkg/ident"
)

// wrapLazyArgument creates a thunk for a lazy parameter, capturing the current environment.
func (i *Interpreter) wrapLazyArgument(arg ast.Expression) Value {
	return NewLazyThunk(arg, i.env, i)
}

// evalCallExpression evaluates a DWScript function call expression.
//
// This method handles:
//   - Direct function calls (user-defined and built-in)
//   - Function pointer calls, including proper handling of lazy and var parameters
//   - Member access calls (unit-qualified functions, record/class methods)
//   - Overload resolution for user-defined functions
//   - Error handling for invalid calls, argument mismatches, and unsupported cases
//
// For function pointers, it checks parameter flags to support lazy evaluation and by-reference passing.
// For member access, it distinguishes between unit-qualified calls and record/class method invocations.
// Returns a Value representing the result of the function call, or an ErrorValue on failure.
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
						// For lazy parameters, reuse existing thunks to avoid self-recursive wrapping
						args[idx] = i.wrapLazyArgument(arg)
					} else if isByRef {
						// For var parameters, create a reference or pass through existing reference
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

		// Check if this is a record method call (record.Method(...))
		// Implementation Strategy: Integrate with class method system (Option 2)
		// Records are value types, but methods work similarly to class methods.
		// Key difference: For mutating methods (procedures), we need copy-back semantics.
		if recVal, ok := objVal.(*RecordValue); ok {
			return i.evalRecordMethodCall(recVal, memberAccess, expr.Arguments, memberAccess.Object)
		}

		// Check if this is an interface method call (interface.Method(...))
		if ifaceInst, ok := objVal.(*InterfaceInstance); ok {
			// Dispatch to the underlying object
			if ifaceInst.Object == nil {
				return i.newErrorWithLocation(expr, "cannot call method on nil interface")
			}
			// Call the method on the underlying object by temporarily swapping the variable
			if objIdent, ok := memberAccess.Object.(*ast.Identifier); ok {
				savedVal, exists := i.env.Get(objIdent.Value)
				if exists {
					// Temporarily set to underlying object
					_ = i.env.Set(objIdent.Value, ifaceInst.Object)
					// Use defer to ensure restoration even if method call panics or returns early
					defer func() { _ = i.env.Set(objIdent.Value, savedVal) }()

					// Create a method call expression
					mc := &ast.MethodCallExpression{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: expr.Token,
							},
						},
						Object:    memberAccess.Object,
						Method:    memberAccess.Member,
						Arguments: expr.Arguments,
					}
					return i.evalMethodCall(mc)
				}
			}
			return i.newErrorWithLocation(expr, "interface method call requires identifier")
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
								// For lazy parameters, reuse existing thunks to avoid self-recursive wrapping
								args[idx] = i.wrapLazyArgument(arg)
							} else if isByRef {
								// For var parameters, create a reference or pass through existing reference
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
						return i.executeUserFunctionViaEvaluator(fn, args)
					}
					// Function not found in unit
					return i.newErrorWithLocation(expr, "function '%s' not found in unit '%s'", memberAccess.Member.Value, unitIdent.Value)
				}
			}
		}

		// Check if this is a class constructor call (TClass.Create(...))
		// When calling TObj.Create(args), the parser creates CallExpression with MemberAccessExpression
		if ident, ok := memberAccess.Object.(*ast.Identifier); ok {
			// Check if this identifier refers to a class (case-insensitive)
			var classInfo *ClassInfo
			for className, class := range i.classes {
				if pkgident.Equal(className, ident.Value) {
					classInfo = class
					break
				}
			}
			if classInfo != nil {
				// This is a class constructor/method call - convert to MethodCallExpression
				mc := &ast.MethodCallExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: expr.Token,
						},
					},
					Object:    ident,
					Method:    memberAccess.Member,
					Arguments: expr.Arguments,
				}
				return i.evalMethodCall(mc)
			}
		}

		// Not a unit-qualified call - could be a method call, let it fall through
		// to be handled as a method call on an object
		return i.newErrorWithLocation(expr, "cannot call member expression that is not a method or unit-qualified function")
	}

	// Get the function name
	if _, isIdent := expr.Function.(*ast.Identifier); !isIdent {
		funcVal := i.Eval(expr.Function)
		if isError(funcVal) {
			return funcVal
		}

		if funcPtr, isFuncPtr := funcVal.(*FunctionPointerValue); isFuncPtr {
			args := make([]Value, len(expr.Arguments))
			for idx, arg := range expr.Arguments {
				isLazy := false
				isByRef := false
				if funcPtr.Function != nil && idx < len(funcPtr.Function.Parameters) {
					isLazy = funcPtr.Function.Parameters[idx].IsLazy
					isByRef = funcPtr.Function.Parameters[idx].ByRef
				}

				if isLazy {
					args[idx] = i.wrapLazyArgument(arg)
				} else if isByRef {
					if argIdent, ok := arg.(*ast.Identifier); ok {
						if val, exists := i.env.Get(argIdent.Value); exists {
							if refVal, isRef := val.(*ReferenceValue); isRef {
								args[idx] = refVal
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
					val := i.Eval(arg)
					if isError(val) {
						return val
					}
					args[idx] = val
				}
			}
			return i.callFunctionPointer(funcPtr, args, expr)
		}

		if classVal, ok := funcVal.(*ClassValue); ok {
			// Calling a metaclass value with no arguments returns the class reference itself.
			if len(expr.Arguments) != 0 {
				return i.newErrorWithLocation(expr, "class reference call does not take arguments")
			}
			return classVal
		}

		return i.newErrorWithLocation(expr, "function call requires identifier or qualified name, got %T", expr.Function)
	}

	funcName := expr.Function.(*ast.Identifier)

	// Check if it's a user-defined function first
	// DWScript is case-insensitive, so normalize the function name to lowercase
	funcNameLower := pkgident.Normalize(funcName.Value)
	if overloads, exists := i.functions[funcNameLower]; exists && len(overloads) > 0 {
		// Resolve overload based on argument types and get cached evaluated arguments
		fn, cachedArgs, err := i.resolveOverload(funcNameLower, overloads, expr.Arguments)
		if err != nil {
			return i.newErrorWithLocation(expr, "%s", err.Error())
		}

		// Prepare arguments - lazy parameters get LazyThunks, var parameters get References, regular parameters use cached values
		args := make([]Value, len(expr.Arguments))
		for idx, arg := range expr.Arguments {
			// Check parameter flags
			isLazy := idx < len(fn.Parameters) && fn.Parameters[idx].IsLazy
			isByRef := idx < len(fn.Parameters) && fn.Parameters[idx].ByRef

			if isLazy {
				// For lazy parameters, create a LazyThunk with the unevaluated expression
				// and the current environment (captured from call site)
				args[idx] = i.wrapLazyArgument(arg)
			} else if isByRef {
				// For var parameters, create a reference to the variable
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
				// For regular parameters, use the cached value from overload resolution
				// This prevents double evaluation of function arguments
				args[idx] = cachedArgs[idx]
			}
		}
		return i.executeUserFunctionViaEvaluator(fn, args)
	}

	// Check if this is an instance method call within the current context (implicit Self)
	if selfVal, ok := i.env.Get("Self"); ok {
		if obj, isObj := AsObject(selfVal); isObj {
			if obj.GetMethod(funcName.Value) != nil {
				mc := &ast.MethodCallExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: expr.Token,
						},
					},
					Object: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: funcName.Token,
							},
						},
						Value: "Self",
					},
					Method:    funcName,
					Arguments: expr.Arguments,
				}
				return i.evalMethodCall(mc)
			}
		}
	}

	// Check if this is a static method call within the current record context
	if recordVal, ok := i.env.Get("__CurrentRecord__"); ok {
		if rtv, isRecord := recordVal.(*RecordTypeValue); isRecord {
			// Check if the function name matches a static method (case-insensitive)
			methodNameLower := pkgident.Normalize(funcName.Value)
			if overloads, exists := rtv.ClassMethodOverloads[methodNameLower]; exists && len(overloads) > 0 {
				// Found a static method - convert to qualified call
				mc := &ast.MethodCallExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: expr.Token,
						},
					},
					Object: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: funcName.Token,
							},
						},
						Value: rtv.RecordType.Name,
					},
					Method:    funcName,
					Arguments: expr.Arguments,
				}
				return i.evalMethodCall(mc)
			}
		}
	}

	if funcName.Value == "TryStrToInt" || funcName.Value == "TryStrToFloat" ||
		funcName.Value == "DecodeDate" || funcName.Value == "DecodeTime" {
		// Delegate to the evaluator which handles these with the visitor pattern
		return i.evaluatorInstance.Eval(expr, i.ctx)
	}

	// Check if this is a built-in function with var parameters
	// These functions need the AST node for the first argument to modify it in place
	if funcName.Value == "Inc" || funcName.Value == "Dec" || funcName.Value == "Insert" ||
		(funcName.Value == "Delete" && len(expr.Arguments) == 3) ||
		funcName.Value == "Swap" || funcName.Value == "DivMod" ||
		funcName.Value == "SetLength" {
		return i.callBuiltinWithVarParam(funcName.Value, expr.Arguments)
	}

	// Check if this is an external function with var parameters
	// We need to check BEFORE evaluating args to create ReferenceValues
	if i.externalFunctions != nil {
		if extFunc, ok := i.externalFunctions.Get(funcName.Value); ok {
			varParams := extFunc.Wrapper.GetVarParams()
			paramTypes := extFunc.Wrapper.GetParamTypes()

			// Prepare arguments - create ReferenceValues for var parameters
			args := make([]Value, len(expr.Arguments))
			for idx, arg := range expr.Arguments {
				isVarParam := idx < len(varParams) && varParams[idx]

				if isVarParam {
					// For var parameters, create a reference
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
					// For regular parameters, evaluate with type context if available
					var val Value
					if idx < len(paramTypes) {
						// Parse the parameter type string and provide context for type inference
						expectedType, _ := i.parseTypeString(paramTypes[idx])
						val = i.EvalWithExpectedType(arg, expectedType)
					} else {
						val = i.Eval(arg)
					}
					if isError(val) {
						return val
					}
					args[idx] = val
				}
			}

			return i.callExternalFunction(extFunc, args)
		}
	}

	// Check if this is Default() function (expects unevaluated type identifier)
	// Default(Integer) should pass "Integer" as string, not evaluate it
	// This must be handled before evaluating arguments, similar to type casts
	if funcName.Value == "Default" && len(expr.Arguments) == 1 {
		if defaultValue := i.evalDefaultFunction(expr.Arguments[0]); defaultValue != nil {
			return defaultValue
		}
	}

	// Check if this is a type cast (TypeName(expression))
	// Type casts look like function calls but the "function" name is actually a type name
	if len(expr.Arguments) == 1 {
		if castValue := i.evalTypeCast(funcName.Value, expr.Arguments[0]); castValue != nil || i.exception != nil {
			return castValue
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
		if i.exception != nil {
			// Exception raised while evaluating arguments - skip function execution
			return &NilValue{}
		}
		args[idx] = val
	}

	// Set currentNode to the function name (which has the position of the call)
	// This ensures that built-in functions like Assert() report the correct position
	i.currentNode = funcName
	return i.callBuiltin(funcName.Value, args)
}

// normalizeBuiltinName normalizes a builtin function name to its canonical form
// for case-insensitive matching (DWScript is case-insensitive).
func normalizeBuiltinName(name string) string {
	// Create a lowercase version for comparison
	lower := pkgident.Normalize(name)

	// Map of lowercase names to canonical names
	// This allows case-insensitive function calls
	canonicalNames := map[string]string{
		"println": "PrintLn", "print": "Print", "ord": "Ord", "integer": "Integer",
		"length": "Length", "copy": "Copy", "concat": "Concat", "indexof": "IndexOf",
		"contains": "Contains", "reverse": "Reverse", "sort": "Sort", "pos": "Pos",
		"uppercase": "UpperCase", "lowercase": "LowerCase",
		"asciiuppercase": "ASCIIUpperCase", "asciilowercase": "ASCIILowerCase",
		"ansiuppercase": "AnsiUpperCase", "ansilowercase": "AnsiLowerCase",
		"trim":     "Trim",
		"trimleft": "TrimLeft", "trimright": "TrimRight", "stringreplace": "StringReplace", "strreplace": "StrReplace", "strreplacemacros": "StrReplaceMacros",
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

	// Return canonical name if found, otherwise return original name
	if canonical, ok := canonicalNames[lower]; ok {
		return canonical
	}
	return name
}

// parseTypeString parses a type string (e.g., "Integer", "array of String") into a types.Type.
// Returns nil if the type cannot be parsed.
func (i *Interpreter) parseTypeString(typeStr string) (types.Type, error) {
	return i.resolveType(typeStr)
}

// callBuiltin calls a built-in function by name.
