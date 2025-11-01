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
				// Evaluate arguments
				args := make([]Value, len(expr.Arguments))
				for idx, arg := range expr.Arguments {
					argVal := i.Eval(arg)
					if isError(argVal) {
						return argVal
					}
					args[idx] = argVal
				}
				// Call through the function pointer
				return i.callFunctionPointer(funcPtr, args, expr)
			}
		}
	}

	// Check if this is a unit-qualified function call (UnitName.FunctionName)
	if memberAccess, ok := expr.Function.(*ast.MemberAccessExpression); ok {
		if unitIdent, ok := memberAccess.Object.(*ast.Identifier); ok {
			// This could be a unit-qualified call: UnitName.FunctionName()
			if i.unitRegistry != nil {
				if _, exists := i.unitRegistry.GetUnit(unitIdent.Value); exists {
					// Resolve the qualified function
					fn, err := i.ResolveQualifiedFunction(unitIdent.Value, memberAccess.Member.Value)
					if err == nil {
						// Found the function - evaluate arguments and call it
						args := make([]Value, len(expr.Arguments))
						for idx, arg := range expr.Arguments {
							val := i.Eval(arg)
							if isError(val) {
								return val
							}
							args[idx] = val
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
		// Evaluate all arguments
		args := make([]Value, len(expr.Arguments))
		for idx, arg := range expr.Arguments {
			val := i.Eval(arg)
			if isError(val) {
				return val
			}
			args[idx] = val
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

// callBuiltin calls a built-in function by name.
func (i *Interpreter) callBuiltin(name string, args []Value) Value {
	// Task 9.32: Check for external Go functions first
	if i.externalFunctions != nil {
		if extFunc, ok := i.externalFunctions.Get(name); ok {
			return i.callExternalFunction(extFunc, args)
		}
	}

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
	case "Format":
		return i.builtinFormat(args)
	case "Abs":
		return i.builtinAbs(args)
	case "Min":
		return i.builtinMin(args)
	case "Max":
		return i.builtinMax(args)
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
	default:
		return i.newErrorWithLocation(i.currentNode, "undefined function: %s", name)
	}
}

// callExternalFunction calls an external Go function registered via FFI (Task 9.32).
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
	// Check argument count matches parameter count
	if len(args) != len(fn.Parameters) {
		return newError("wrong number of arguments: expected %d, got %d",
			len(fn.Parameters), len(args))
	}

	// Create a new environment for the function scope
	funcEnv := NewEnclosedEnvironment(i.env)
	savedEnv := i.env
	i.env = funcEnv

	// Push function name onto call stack for stack traces
	i.callStack = append(i.callStack, fn.Name.Value)
	// Ensure it's popped when function exits (even if exception occurs)
	defer func() {
		if len(i.callStack) > 0 {
			i.callStack = i.callStack[:len(i.callStack)-1]
		}
	}()

	// Bind parameters to arguments
	for idx, param := range fn.Parameters {
		arg := args[idx]

		// Task 8.19b: Apply implicit conversion if parameter has a type and types don't match
		if param.Type != nil {
			paramTypeName := param.Type.Name
			if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
				arg = converted
			}
		}

		if param.ByRef {
			// By-reference parameter - we need to handle this specially
			// For now, we'll pass by value (TODO: implement proper by-ref support)
			i.env.Define(param.Name.Value, arg)
		} else {
			// By-value parameter
			i.env.Define(param.Name.Value, arg)
		}
	}

	// For functions (not procedures), initialize the Result variable
	if fn.ReturnType != nil {
		// Initialize Result based on return type
		var resultValue Value = &NilValue{}

		// Check if return type is a record
		returnTypeName := fn.ReturnType.Name
		recordTypeKey := "__record_type_" + returnTypeName
		if typeVal, ok := i.env.Get(recordTypeKey); ok {
			if rtv, ok := typeVal.(*RecordTypeValue); ok {
				resultValue = NewRecordValue(rtv.RecordType)
			}
		}

		i.env.Define("Result", resultValue)
		// Also define the function name as an alias for Result (DWScript style)
		i.env.Define(fn.Name.Value, resultValue)
	}

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
		// Check both Result and function name variable
		// Prioritize whichever one was actually set (not nil)
		resultVal, resultOk := i.env.Get("Result")
		fnNameVal, fnNameOk := i.env.Get(fn.Name.Value)

		// Use whichever variable is not nil, preferring Result if both are set
		if resultOk && resultVal.Type() != "NIL" {
			returnValue = resultVal
		} else if fnNameOk && fnNameVal.Type() != "NIL" {
			returnValue = fnNameVal
		} else if resultOk {
			// Result exists but is nil - use it
			returnValue = resultVal
		} else if fnNameOk {
			// Function name exists but is nil - use it
			returnValue = fnNameVal
		} else {
			// Neither exists (shouldn't happen)
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

	// Push lambda marker onto call stack for stack traces
	i.callStack = append(i.callStack, "<lambda>")
	defer func() {
		if len(i.callStack) > 0 {
			i.callStack = i.callStack[:len(i.callStack)-1]
		}
	}()

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
		// Initialize Result based on return type
		var resultValue Value = &NilValue{}

		// Check if return type is a record
		returnTypeName := lambda.ReturnType.Name
		recordTypeKey := "__record_type_" + returnTypeName
		if typeVal, ok := i.env.Get(recordTypeKey); ok {
			if rtv, ok := typeVal.(*RecordTypeValue); ok {
				resultValue = NewRecordValue(rtv.RecordType)
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
