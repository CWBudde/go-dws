package builtins

import (
	"math"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// =============================================================================
// Trigonometric Functions
// =============================================================================

// Sin implements the Sin() built-in function.
// It returns the sine of a number (in radians).
// Sin(x) - returns sine as Float (always returns Float, even for Integer input)
func Sin(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Sin() expects exactly 1 argument, got %d", len(args))
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
		return ctx.NewError("Sin() expects Integer or Float as argument, got %s", arg.Type())
	}

	return &runtime.FloatValue{Value: math.Sin(value)}
}

// Cos implements the Cos() built-in function.
// It returns the cosine of a number (in radians).
// Cos(x) - returns cosine as Float (always returns Float, even for Integer input)
func Cos(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Cos() expects exactly 1 argument, got %d", len(args))
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
		return ctx.NewError("Cos() expects Integer or Float as argument, got %s", arg.Type())
	}

	return &runtime.FloatValue{Value: math.Cos(value)}
}

// Tan implements the Tan() built-in function.
// It returns the tangent of a number (in radians).
// Tan(x) - returns tangent as Float (always returns Float, even for Integer input)
func Tan(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Tan() expects exactly 1 argument, got %d", len(args))
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
		return ctx.NewError("Tan() expects Integer or Float as argument, got %s", arg.Type())
	}

	return &runtime.FloatValue{Value: math.Tan(value)}
}

// ArcSin implements the ArcSin() built-in function.
// It returns the arcsine (inverse sine) of a number.
// ArcSin(x) - returns arcsine in radians as Float (x must be in [-1, 1])
func ArcSin(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("ArcSin() expects exactly 1 argument, got %d", len(args))
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
		return ctx.NewError("ArcSin() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Check domain: x must be in [-1, 1]
	if value < -1.0 || value > 1.0 {
		return ctx.NewError("ArcSin() argument must be in [-1, 1], got %f", value)
	}

	return &runtime.FloatValue{Value: math.Asin(value)}
}

// ArcCos implements the ArcCos() built-in function.
// It returns the arccosine (inverse cosine) of a number.
// ArcCos(x) - returns arccosine in radians as Float (x must be in [-1, 1])
func ArcCos(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("ArcCos() expects exactly 1 argument, got %d", len(args))
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
		return ctx.NewError("ArcCos() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Check domain: x must be in [-1, 1]
	if value < -1.0 || value > 1.0 {
		return ctx.NewError("ArcCos() argument must be in [-1, 1], got %f", value)
	}

	return &runtime.FloatValue{Value: math.Acos(value)}
}

// ArcTan implements the ArcTan() built-in function.
// It returns the arctangent (inverse tangent) of a number.
// ArcTan(x) - returns arctangent in radians as Float
func ArcTan(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("ArcTan() expects exactly 1 argument, got %d", len(args))
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
		return ctx.NewError("ArcTan() expects Integer or Float as argument, got %s", arg.Type())
	}

	return &runtime.FloatValue{Value: math.Atan(value)}
}

// ArcTan2 implements the ArcTan2() built-in function.
// It returns the arctangent of y/x, using the signs of both to determine the quadrant.
// ArcTan2(y, x) - returns arctangent in radians as Float
func ArcTan2(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("ArcTan2() expects exactly 2 arguments, got %d", len(args))
	}

	y := args[0]
	x := args[1]

	var yFloat, xFloat float64

	// Handle y parameter
	if intVal, ok := y.(*runtime.IntegerValue); ok {
		yFloat = float64(intVal.Value)
	} else if floatVal, ok := y.(*runtime.FloatValue); ok {
		yFloat = floatVal.Value
	} else {
		return ctx.NewError("ArcTan2() expects Integer or Float as first argument, got %s", y.Type())
	}

	// Handle x parameter
	if intVal, ok := x.(*runtime.IntegerValue); ok {
		xFloat = float64(intVal.Value)
	} else if floatVal, ok := x.(*runtime.FloatValue); ok {
		xFloat = floatVal.Value
	} else {
		return ctx.NewError("ArcTan2() expects Integer or Float as second argument, got %s", x.Type())
	}

	return &runtime.FloatValue{Value: math.Atan2(yFloat, xFloat)}
}

// CoTan implements the CoTan() built-in function.
// It returns the cotangent of a number (in radians).
// CoTan(x) - returns cotangent as Float (1 / Tan(x))
func CoTan(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("CoTan() expects exactly 1 argument, got %d", len(args))
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
		return ctx.NewError("CoTan() expects Integer or Float as argument, got %s", arg.Type())
	}

	// CoTan(x) = 1 / Tan(x) = Cos(x) / Sin(x)
	tan := math.Tan(value)
	if tan == 0 {
		return ctx.NewError("CoTan() undefined at x=%f (division by zero)", value)
	}

	return &runtime.FloatValue{Value: 1.0 / tan}
}

// Hypot implements the Hypot() built-in function.
// It returns the hypotenuse: sqrt(x² + y²)
// Hypot(x, y) - returns hypotenuse as Float
func Hypot(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("Hypot() expects exactly 2 arguments, got %d", len(args))
	}

	x := args[0]
	y := args[1]

	var xFloat, yFloat float64

	// Handle x parameter
	if intVal, ok := x.(*runtime.IntegerValue); ok {
		xFloat = float64(intVal.Value)
	} else if floatVal, ok := x.(*runtime.FloatValue); ok {
		xFloat = floatVal.Value
	} else {
		return ctx.NewError("Hypot() expects Integer or Float as first argument, got %s", x.Type())
	}

	// Handle y parameter
	if intVal, ok := y.(*runtime.IntegerValue); ok {
		yFloat = float64(intVal.Value)
	} else if floatVal, ok := y.(*runtime.FloatValue); ok {
		yFloat = floatVal.Value
	} else {
		return ctx.NewError("Hypot() expects Integer or Float as second argument, got %s", y.Type())
	}

	return &runtime.FloatValue{Value: math.Hypot(xFloat, yFloat)}
}

// Sinh implements the Sinh() built-in function.
// It returns the hyperbolic sine of a number.
// Sinh(x) - returns hyperbolic sine as Float
func Sinh(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Sinh() expects exactly 1 argument, got %d", len(args))
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
		return ctx.NewError("Sinh() expects Integer or Float as argument, got %s", arg.Type())
	}

	return &runtime.FloatValue{Value: math.Sinh(value)}
}

// Cosh implements the Cosh() built-in function.
// It returns the hyperbolic cosine of a number.
// Cosh(x) - returns hyperbolic cosine as Float
func Cosh(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Cosh() expects exactly 1 argument, got %d", len(args))
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
		return ctx.NewError("Cosh() expects Integer or Float as argument, got %s", arg.Type())
	}

	return &runtime.FloatValue{Value: math.Cosh(value)}
}

// Tanh implements the Tanh() built-in function.
// It returns the hyperbolic tangent of a number.
// Tanh(x) - returns hyperbolic tangent as Float
func Tanh(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Tanh() expects exactly 1 argument, got %d", len(args))
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
		return ctx.NewError("Tanh() expects Integer or Float as argument, got %s", arg.Type())
	}

	return &runtime.FloatValue{Value: math.Tanh(value)}
}

// ArcSinh implements the ArcSinh() built-in function.
// It returns the inverse hyperbolic sine of a number.
// ArcSinh(x) - returns inverse hyperbolic sine as Float
func ArcSinh(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("ArcSinh() expects exactly 1 argument, got %d", len(args))
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
		return ctx.NewError("ArcSinh() expects Integer or Float as argument, got %s", arg.Type())
	}

	return &runtime.FloatValue{Value: math.Asinh(value)}
}

// ArcCosh implements the ArcCosh() built-in function.
// It returns the inverse hyperbolic cosine of a number.
// ArcCosh(x) - returns inverse hyperbolic cosine as Float (x must be >= 1)
func ArcCosh(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("ArcCosh() expects exactly 1 argument, got %d", len(args))
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
		return ctx.NewError("ArcCosh() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Check domain: x must be >= 1
	if value < 1.0 {
		return ctx.NewError("ArcCosh() argument must be >= 1, got %f", value)
	}

	return &runtime.FloatValue{Value: math.Acosh(value)}
}

// ArcTanh implements the ArcTanh() built-in function.
// It returns the inverse hyperbolic tangent of a number.
// ArcTanh(x) - returns inverse hyperbolic tangent as Float (x must be in (-1, 1))
func ArcTanh(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("ArcTanh() expects exactly 1 argument, got %d", len(args))
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
		return ctx.NewError("ArcTanh() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Check domain: x must be in (-1, 1)
	if value <= -1.0 || value >= 1.0 {
		return ctx.NewError("ArcTanh() argument must be in (-1, 1), got %f", value)
	}

	return &runtime.FloatValue{Value: math.Atanh(value)}
}
