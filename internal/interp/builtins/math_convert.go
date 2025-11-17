package builtins

import (
	"math"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// =============================================================================
// Conversion Functions (Angle and Rounding)
// =============================================================================

// DegToRad implements the DegToRad() built-in function.
// It converts degrees to radians.
// DegToRad(degrees) - returns radians as Float
func DegToRad(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("DegToRad() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	var value float64

	// Handle Integer - convert to float
	if intVal, ok := arg.(*runtime.IntegerValue); ok {
		value = float64(intVal.Value)
	} else if floatVal, ok := arg.(*runtime.FloatValue); ok {
		// Handle Float
		value = floatVal.Value
	} else {
		return ctx.NewError("DegToRad() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Convert degrees to radians: radians = degrees * π / 180
	return &runtime.FloatValue{Value: value * math.Pi / 180.0}
}

// RadToDeg implements the RadToDeg() built-in function.
// It converts radians to degrees.
// RadToDeg(radians) - returns degrees as Float
func RadToDeg(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("RadToDeg() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	var value float64

	// Handle Integer - convert to float
	if intVal, ok := arg.(*runtime.IntegerValue); ok {
		value = float64(intVal.Value)
	} else if floatVal, ok := arg.(*runtime.FloatValue); ok {
		// Handle Float
		value = floatVal.Value
	} else {
		return ctx.NewError("RadToDeg() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Convert radians to degrees: degrees = radians * 180 / π
	return &runtime.FloatValue{Value: value * 180.0 / math.Pi}
}

// Round implements the Round() built-in function.
// It rounds a number to the nearest integer.
// Round(x) - returns rounded value as Integer (always returns Integer)
// Task 9.4.5: Now supports Variant arguments.
func Round(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Round() expects exactly 1 argument, got %d", len(args))
	}

	// Task 9.4.5: Unwrap Variant if necessary
	arg := ctx.UnwrapVariant(args[0])
	var value float64

	// Handle Integer - already an integer, just return it
	if intVal, ok := arg.(*runtime.IntegerValue); ok {
		return &runtime.IntegerValue{Value: intVal.Value}
	} else if floatVal, ok := arg.(*runtime.FloatValue); ok {
		// Handle Float
		value = floatVal.Value
	} else {
		return ctx.NewError("Round() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Round to nearest integer using banker's rounding (round-half-to-even)
	// DWScript uses banker's rounding: 16.5 → 16, 17.5 → 18
	rounded := math.RoundToEven(value)
	return &runtime.IntegerValue{Value: int64(rounded)}
}

// Trunc implements the Trunc() built-in function.
// It truncates a number towards zero (removes the decimal part).
// Trunc(x) - returns truncated value as Integer (always returns Integer)
func Trunc(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Trunc() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	var value float64

	// Handle Integer - already an integer, just return it
	if intVal, ok := arg.(*runtime.IntegerValue); ok {
		return &runtime.IntegerValue{Value: intVal.Value}
	} else if floatVal, ok := arg.(*runtime.FloatValue); ok {
		// Handle Float
		value = floatVal.Value
	} else {
		return ctx.NewError("Trunc() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Truncate toward zero
	truncated := math.Trunc(value)
	return &runtime.IntegerValue{Value: int64(truncated)}
}

// Ceil implements the Ceil() built-in function.
// It returns the smallest integer greater than or equal to x.
// Ceil(x) - returns ceiling value as Integer (always returns Integer)
func Ceil(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Ceil() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	var value float64

	// Handle Integer - already an integer, just return it
	if intVal, ok := arg.(*runtime.IntegerValue); ok {
		return &runtime.IntegerValue{Value: intVal.Value}
	} else if floatVal, ok := arg.(*runtime.FloatValue); ok {
		// Handle Float
		value = floatVal.Value
	} else {
		return ctx.NewError("Ceil() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Ceil returns float64 - convert to integer
	ceiling := math.Ceil(value)
	return &runtime.IntegerValue{Value: int64(ceiling)}
}

// Floor implements the Floor() built-in function.
// It returns the largest integer less than or equal to x.
// Floor(x) - returns floor value as Integer (always returns Integer)
func Floor(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Floor() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	var value float64

	// Handle Integer - already an integer, just return it
	if intVal, ok := arg.(*runtime.IntegerValue); ok {
		return &runtime.IntegerValue{Value: intVal.Value}
	} else if floatVal, ok := arg.(*runtime.FloatValue); ok {
		// Handle Float
		value = floatVal.Value
	} else {
		return ctx.NewError("Floor() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Floor returns float64 - convert to integer
	floor := math.Floor(value)
	return &runtime.IntegerValue{Value: int64(floor)}
}

// ClampInt implements the ClampInt() built-in function.
// It clamps an integer value to a range [min, max].
// ClampInt(value, min, max: Integer): Integer
func ClampInt(ctx Context, args []Value) Value {
	if len(args) != 3 {
		return ctx.NewError("ClampInt() expects exactly 3 arguments, got %d", len(args))
	}

	// Extract value
	valueArg := args[0]
	valueInt, ok := valueArg.(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("ClampInt() expects Integer for argument 1, got %s", valueArg.Type())
	}
	value := valueInt.Value

	// Extract min
	minArg := args[1]
	minInt, ok := minArg.(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("ClampInt() expects Integer for argument 2, got %s", minArg.Type())
	}
	min := minInt.Value

	// Extract max
	maxArg := args[2]
	maxInt, ok := maxArg.(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("ClampInt() expects Integer for argument 3, got %s", maxArg.Type())
	}
	max := maxInt.Value

	// Clamp the value
	result := value
	if result < min {
		result = min
	} else if result > max {
		result = max
	}

	return &runtime.IntegerValue{Value: result}
}

// Clamp implements the Clamp() built-in function.
// It clamps a float value to a range [min, max].
// Clamp(value, min, max: Float): Float
// Accepts mixed Integer/Float arguments (converts to Float)
func Clamp(ctx Context, args []Value) Value {
	if len(args) != 3 {
		return ctx.NewError("Clamp() expects exactly 3 arguments, got %d", len(args))
	}

	// Extract value (convert to float)
	var value float64
	switch v := args[0].(type) {
	case *runtime.IntegerValue:
		value = float64(v.Value)
	case *runtime.FloatValue:
		value = v.Value
	default:
		return ctx.NewError("Clamp() expects Integer or Float for argument 1, got %s", args[0].Type())
	}

	// Extract min (convert to float)
	var min float64
	switch v := args[1].(type) {
	case *runtime.IntegerValue:
		min = float64(v.Value)
	case *runtime.FloatValue:
		min = v.Value
	default:
		return ctx.NewError("Clamp() expects Integer or Float for argument 2, got %s", args[1].Type())
	}

	// Extract max (convert to float)
	var max float64
	switch v := args[2].(type) {
	case *runtime.IntegerValue:
		max = float64(v.Value)
	case *runtime.FloatValue:
		max = v.Value
	default:
		return ctx.NewError("Clamp() expects Integer or Float for argument 3, got %s", args[2].Type())
	}

	// Clamp the value
	result := value
	if result < min {
		result = min
	} else if result > max {
		result = max
	}

	return &runtime.FloatValue{Value: result}
}

// Frac implements the Frac() built-in function.
// It returns the fractional part of a number.
// Frac(x: Float): Float
func Frac(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Frac() expects exactly 1 argument, got %d", len(args))
	}

	var floatVal float64
	switch v := args[0].(type) {
	case *runtime.FloatValue:
		floatVal = v.Value
	case *runtime.IntegerValue:
		floatVal = float64(v.Value)
	default:
		return ctx.NewError("Frac() expects Float or Integer, got %s", args[0].Type())
	}

	// Fractional part = x - floor(x)
	// For negative numbers: Frac(-2.3) = -2.3 - (-3) = 0.7
	_, frac := math.Modf(floatVal)
	return &runtime.FloatValue{Value: frac}
}

// Int implements the Int() built-in function.
// It returns the integer part of a number as a Float (different from Trunc).
// Int(x: Float): Float
func Int(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Int() expects exactly 1 argument, got %d", len(args))
	}

	var floatVal float64
	switch v := args[0].(type) {
	case *runtime.FloatValue:
		floatVal = v.Value
	case *runtime.IntegerValue:
		floatVal = float64(v.Value)
	default:
		return ctx.NewError("Int() expects Float or Integer, got %s", args[0].Type())
	}

	// Int() returns the integer part (truncated towards zero) as a Float
	return &runtime.FloatValue{Value: math.Trunc(floatVal)}
}

// IntPower implements the IntPower() built-in function.
// It computes base^exponent using integer exponentiation (faster than Power for integer exponents).
// IntPower(base: Float, exponent: Integer): Float
func IntPower(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("IntPower() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: base (Float or Integer)
	var base float64
	switch v := args[0].(type) {
	case *runtime.FloatValue:
		base = v.Value
	case *runtime.IntegerValue:
		base = float64(v.Value)
	default:
		return ctx.NewError("IntPower() expects Float or Integer as first argument, got %s", args[0].Type())
	}

	// Second argument: exponent (Integer)
	exponentVal, ok := args[1].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("IntPower() expects Integer as second argument, got %s", args[1].Type())
	}
	exponent := exponentVal.Value

	// Handle special cases
	if exponent == 0 {
		return &runtime.FloatValue{Value: 1.0}
	}
	if exponent < 0 {
		// For negative exponents, compute 1 / base^(-exponent)
		base = 1.0 / base
		exponent = -exponent
	}

	// Fast integer exponentiation using exponentiation by squaring
	result := 1.0
	for exponent > 0 {
		if exponent%2 == 1 {
			result *= base
		}
		base *= base
		exponent /= 2
	}

	return &runtime.FloatValue{Value: result}
}
