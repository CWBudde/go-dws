package builtins

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// System and miscellaneous built-in functions for DWScript.
// This file contains utility functions that don't fit into other categories:
// - Stack introspection: GetStackTrace, GetCallStack
// - Ordinal functions: Succ, Pred, Ord, Integer
// - Type conversion with defaults: StrToIntDef, StrToFloatDef
// - Runtime utilities: Assigned, Assert
// - String formatting: Format (complex)

// GetStackTrace returns a formatted string representation of the current call stack.
//
// Signature: GetStackTrace() -> String
//
// Returns: String containing the formatted call stack with function names and line numbers
//
// Example:
//
//	try
//	  RaiseException();
//	except
//	  on E: Exception do
//	    PrintLn(GetStackTrace());
//	end;
func GetStackTrace(ctx Context, args []Value) Value {
	// Validate argument count
	if len(args) != 0 {
		return ctx.NewError("GetStackTrace() expects no arguments, got %d", len(args))
	}

	// Get the call stack as a formatted string
	stackTrace := ctx.GetCallStackString()

	return &runtime.StringValue{Value: stackTrace}
}

// GetCallStack returns the current call stack as an array of records.
//
// Signature: GetCallStack() -> array of record
//
// Returns: Array of records, each containing:
//   - FunctionName: String - name of the function
//   - Line: Integer - line number
//   - Column: Integer - column number
//
// Example:
//
//	var stack := GetCallStack();
//	for var i := 0 to High(stack) do
//	  PrintLn(stack[i].FunctionName + ' at line ' + IntToStr(stack[i].Line));
func GetCallStack(ctx Context, args []Value) Value {
	// Validate argument count
	if len(args) != 0 {
		return ctx.NewError("GetCallStack() expects no arguments, got %d", len(args))
	}

	// Get the call stack as an array of records
	callStack := ctx.GetCallStackArray()

	return callStack
}

// Assigned checks if a value is assigned (not nil).
//
// Signature: Assigned(value: Variant) -> Boolean
//
// Returns: true if value is not nil, false otherwise
//
// Example:
//
//	var obj: TMyClass;
//	if Assigned(obj) then
//	  obj.DoSomething();
func Assigned(ctx Context, args []Value) Value {
	// Validate argument count
	if len(args) != 1 {
		return ctx.NewError("Assigned() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Check if argument is assigned (not nil)
	isAssigned := ctx.IsAssigned(arg)

	return &runtime.BooleanValue{Value: isAssigned}
}

// Assert validates a condition and raises an EAssertionFailed exception if false.
//
// Signature: Assert(condition: Boolean)
// Signature: Assert(condition: Boolean, message: String)
//
// Parameters:
//   - condition: Boolean condition to check
//   - message: Optional custom error message
//
// Returns: nil if condition is true
// Raises: EAssertionFailed exception if condition is false
//
// Example:
//
//	Assert(x > 0);
//	Assert(x > 0, 'x must be positive');
func Assert(ctx Context, args []Value) Value {
	// Validate argument count (1-2 arguments)
	if len(args) < 1 || len(args) > 2 {
		return ctx.NewError("Assert() expects 1-2 arguments, got %d", len(args))
	}

	// First argument must be Boolean
	condition, ok := args[0].(*runtime.BooleanValue)
	if !ok {
		return ctx.NewError("Assert() first argument must be Boolean, got %T", args[0])
	}

	// If condition is true, assertion passes - return nil
	if condition.Value {
		return &runtime.NilValue{}
	}

	// Condition is false - raise EAssertionFailed exception
	// Get optional custom message
	var customMessage string
	if len(args) == 2 {
		customMsg, ok := args[1].(*runtime.StringValue)
		if !ok {
			return ctx.NewError("Assert() second argument must be String, got %T", args[1])
		}
		customMessage = customMsg.Value
	}

	// Raise the assertion failed exception via Context
	ctx.RaiseAssertionFailed(customMessage)

	// This return should never be reached (exception raised above)
	return nil
}

// Integer converts values to integers.
// NOTE: Ord() is already defined in ordinal.go
//
// Signature: Integer(value: Variant) -> Integer
//
// Parameters:
//   - value: Enum, Integer, Float, or Boolean
//
// Returns: Integer value
//
// Example:
//
//	var f := 3.7;
//	PrintLn(Integer(f));  // Prints 4 (rounded)
func Integer(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Integer() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Use Context helper for conversion
	intVal, ok := ctx.ToInt64(arg)
	if !ok {
		return ctx.NewError("Integer() cannot convert %T to integer", arg)
	}

	return &runtime.IntegerValue{Value: intVal}
}

// StrToIntDef converts a string to an integer, returning a default value if invalid.
//
// Signature: StrToIntDef(s: String, default: Integer) -> Integer
// Signature: StrToIntDef(s: String, default: Integer, base: Integer) -> Integer
//
// Parameters:
//   - s: String to convert
//   - default: Default value to return on error
//   - base: Optional base (2-36), default is 10
//
// Returns: Integer value or default on error
//
// Example:
//
//	var x := StrToIntDef('abc', -1);  // Returns -1
//	var y := StrToIntDef('42', 0);    // Returns 42
//	var z := StrToIntDef('1010', 0, 2);  // Returns 10 (binary)
func StrToIntDef(ctx Context, args []Value) Value {
	if len(args) < 2 || len(args) > 3 {
		return ctx.NewError("StrToIntDef() expects 2 or 3 arguments, got %d", len(args))
	}

	// First argument must be a string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrToIntDef() expects string as first argument, got %T", args[0])
	}

	// Second argument must be an integer (the default value)
	defaultVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("StrToIntDef() expects integer as second argument, got %T", args[1])
	}

	// Default base is 10
	base := 10

	// If third argument is provided, it specifies the base
	if len(args) == 3 {
		baseVal, ok := ctx.ToInt64(args[2])
		if !ok {
			return ctx.NewError("StrToIntDef() expects integer as third argument (base), got %T", args[2])
		}
		base = int(baseVal)

		// Validate base range (2-36)
		if base < 2 || base > 36 {
			return ctx.NewError("StrToIntDef() base must be between 2 and 36, got %d", base)
		}
	}

	// Try to parse the string using Context helper
	intValue, ok := ctx.ParseInt(strVal.Value, base)
	if !ok {
		// Return the default value on error
		return &runtime.IntegerValue{Value: defaultVal.Value}
	}

	return &runtime.IntegerValue{Value: intValue}
}

// StrToFloatDef converts a string to a float, returning a default value if invalid.
//
// Signature: StrToFloatDef(s: String, default: Float) -> Float
//
// Parameters:
//   - s: String to convert
//   - default: Default value to return on error
//
// Returns: Float value or default on error
//
// Example:
//
//	var x := StrToFloatDef('abc', -1.0);  // Returns -1.0
//	var y := StrToFloatDef('3.14', 0.0);  // Returns 3.14
func StrToFloatDef(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("StrToFloatDef() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument must be a string
	strVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("StrToFloatDef() expects string as first argument, got %T", args[0])
	}

	// Second argument must be a float (the default value)
	defaultVal, ok := args[1].(*runtime.FloatValue)
	if !ok {
		return ctx.NewError("StrToFloatDef() expects float as second argument, got %T", args[1])
	}

	// Try to parse the string using Context helper
	floatValue, ok := ctx.ParseFloat(strVal.Value)
	if !ok {
		// Return the default value on error
		return &runtime.FloatValue{Value: defaultVal.Value}
	}

	return &runtime.FloatValue{Value: floatValue}
}

// NOTE: Succ() and Pred() are already defined in ordinal.go

// Format formats a string using format specifiers and an array of values.
//
// Signature: Format(formatStr: String, args: array of const) -> String
//
// Supports: %s (string), %d (integer), %f (float), %% (literal %)
// Optional: width and precision (%5d, %.2f, %8.2f)
//
// Example:
//
//	var name := 'World';
//	var count := 42;
//	PrintLn(Format('Hello %s! Count: %d', [name, count]));
func Format(ctx Context, args []Value) Value {
	// Expect exactly 2 arguments: format string and array of values
	if len(args) != 2 {
		return ctx.NewError("Format() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: format string
	fmtVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return ctx.NewError("Format() expects string as first argument, got %T", args[0])
	}

	// Second argument: array of values
	arrVal, ok := args[1].(*runtime.ArrayValue)
	if !ok {
		return ctx.NewError("Format() expects array as second argument, got %T", args[1])
	}

	// Delegate to Context helper for actual formatting
	result, err := ctx.FormatString(fmtVal.Value, arrVal.Elements)
	if err != nil {
		// Format error should raise an exception that can be caught by try/except
		baseMsg := "Format invalid or incompatible with argument"
		msg := baseMsg

		// Get position from current node for error reporting
		node := ctx.CurrentNode()
		var pos interface{}
		if node != nil {
			nodePos := node.Pos()
			pos = nodePos
			// Include position in message like DWScript does
			msg = fmt.Sprintf("%s [line: %d, column: %d]", baseMsg, nodePos.Line, nodePos.Column)
		}

		if raiser, ok := ctx.(interface {
			RaiseException(className, message string, pos any)
		}); ok {
			raiser.RaiseException("EDelphi", msg, pos)
		}
		return ctx.NewError("EDelphi: " + msg)
	}

	return &runtime.StringValue{Value: result}
}
