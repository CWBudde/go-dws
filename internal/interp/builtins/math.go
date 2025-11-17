package builtins

import (
	"math"
	"math/bits"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// NOTE: math/rand and time imports will be needed when random functions are implemented
// after Context interface is extended with RandSource(), GetRandSeed(), and SetRandSeed() methods.

// =============================================================================
// Basic Math Functions
// =============================================================================

// Abs implements the Abs() built-in function.
// It returns the absolute value of a number.
// Abs(x) - returns absolute value (Integer → Integer, Float → Float)
func Abs(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Abs() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	switch v := arg.(type) {
	case *runtime.IntegerValue:
		if v.Value < 0 {
			return &runtime.IntegerValue{Value: -v.Value}
		}
		return v
	case *runtime.FloatValue:
		return &runtime.FloatValue{Value: math.Abs(v.Value)}
	default:
		return ctx.NewError("Abs() expects Integer or Float, got %s", arg.Type())
	}
}

// Min implements the Min() built-in function.
// It returns the minimum of two numbers.
// Min(a, b) - supports Integer and Float (mixed allowed)
func Min(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("Min() expects exactly 2 arguments, got %d", len(args))
	}

	left := args[0]
	right := args[1]

	switch l := left.(type) {
	case *runtime.IntegerValue:
		// Integer-Integer
		if r, ok := right.(*runtime.IntegerValue); ok {
			if l.Value < r.Value {
				return l
			}
			return r
		}
		// Integer-Float (promote to float)
		if r, ok := right.(*runtime.FloatValue); ok {
			leftFloat := float64(l.Value)
			if leftFloat < r.Value {
				return &runtime.FloatValue{Value: leftFloat}
			}
			return r
		}
	case *runtime.FloatValue:
		// Float-Float
		if r, ok := right.(*runtime.FloatValue); ok {
			if l.Value < r.Value {
				return l
			}
			return r
		}
		// Float-Integer (promote integer)
		if r, ok := right.(*runtime.IntegerValue); ok {
			rightFloat := float64(r.Value)
			if l.Value < rightFloat {
				return l
			}
			return &runtime.FloatValue{Value: rightFloat}
		}
	}

	return ctx.NewError("Min() expects Integer or Float arguments, got %s and %s", left.Type(), right.Type())
}

// Max implements the Max() built-in function.
// It returns the maximum of two numbers.
// Max(a, b) - supports Integer and Float (mixed allowed)
func Max(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("Max() expects exactly 2 arguments, got %d", len(args))
	}

	left := args[0]
	right := args[1]

	switch l := left.(type) {
	case *runtime.IntegerValue:
		// Integer-Integer
		if r, ok := right.(*runtime.IntegerValue); ok {
			if l.Value > r.Value {
				return l
			}
			return r
		}
		// Integer-Float
		if r, ok := right.(*runtime.FloatValue); ok {
			leftFloat := float64(l.Value)
			if leftFloat > r.Value {
				return &runtime.FloatValue{Value: leftFloat}
			}
			return r
		}
	case *runtime.FloatValue:
		// Float-Float
		if r, ok := right.(*runtime.FloatValue); ok {
			if l.Value > r.Value {
				return l
			}
			return r
		}
		// Float-Integer
		if r, ok := right.(*runtime.IntegerValue); ok {
			rightFloat := float64(r.Value)
			if l.Value > rightFloat {
				return l
			}
			return &runtime.FloatValue{Value: rightFloat}
		}
	}

	return ctx.NewError("Max() expects Integer or Float arguments, got %s and %s", left.Type(), right.Type())
}

// Sqr implements the Sqr() built-in function.
// It returns the square of a number.
// Sqr(x) - returns x * x (Integer → Integer, Float → Float)
func Sqr(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Sqr() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	switch v := arg.(type) {
	case *runtime.IntegerValue:
		return &runtime.IntegerValue{Value: v.Value * v.Value}
	case *runtime.FloatValue:
		return &runtime.FloatValue{Value: v.Value * v.Value}
	default:
		return ctx.NewError("Sqr() expects Integer or Float, got %s", arg.Type())
	}
}

// Power implements the Power() built-in function.
// It raises a number to a power.
// Power(base, exponent) - returns base^exponent as Float
func Power(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("Power() expects exactly 2 arguments, got %d", len(args))
	}

	base := args[0]
	exponent := args[1]

	var baseFloat, exponentFloat float64

	switch v := base.(type) {
	case *runtime.IntegerValue:
		baseFloat = float64(v.Value)
	case *runtime.FloatValue:
		baseFloat = v.Value
	default:
		return ctx.NewError("Power() expects Integer or Float as base, got %s", base.Type())
	}

	switch v := exponent.(type) {
	case *runtime.IntegerValue:
		exponentFloat = float64(v.Value)
	case *runtime.FloatValue:
		exponentFloat = v.Value
	default:
		return ctx.NewError("Power() expects Integer or Float as exponent, got %s", exponent.Type())
	}

	// Use math.Pow() - this handles all special cases including 0^0 = 1
	result := math.Pow(baseFloat, exponentFloat)
	return &runtime.FloatValue{Value: result}
}

// Sqrt implements the Sqrt() built-in function.
// It returns the square root of a number.
// Sqrt(x) - returns sqrt(x) as Float (always returns Float, even for Integer input)
func Sqrt(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Sqrt() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	var value float64

	// Handle Integer - convert to float
	if intVal, ok := arg.(*runtime.IntegerValue); ok {
		if intVal.Value < 0 {
			return ctx.NewError("Sqrt() of negative number (%d)", intVal.Value)
		}
		value = float64(intVal.Value)
	} else if floatVal, ok := arg.(*runtime.FloatValue); ok {
		// Handle Float
		if floatVal.Value < 0 {
			return ctx.NewError("Sqrt() of negative number (%f)", floatVal.Value)
		}
		value = floatVal.Value
	} else {
		return ctx.NewError("Sqrt() expects Integer or Float as argument, got %s", arg.Type())
	}

	return &runtime.FloatValue{Value: math.Sqrt(value)}
}

// Exp implements the Exp() built-in function.
// It returns e raised to the power of x.
// Exp(x) - returns e^x as Float (always returns Float, even for Integer input)
func Exp(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Exp() expects exactly 1 argument, got %d", len(args))
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
		return ctx.NewError("Exp() expects Integer or Float as argument, got %s", arg.Type())
	}

	return &runtime.FloatValue{Value: math.Exp(value)}
}

// Ln implements the Ln() built-in function.
// It returns the natural logarithm (base e) of x.
// Ln(x) - returns natural logarithm as Float (always returns Float, even for Integer input)
func Ln(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Ln() expects exactly 1 argument, got %d", len(args))
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
		return ctx.NewError("Ln() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Check for non-positive numbers (Ln is undefined for x <= 0)
	if value <= 0 {
		return ctx.NewError("Ln() of non-positive number (%f)", value)
	}

	return &runtime.FloatValue{Value: math.Log(value)}
}

// Log2 implements the Log2() built-in function.
// It returns the base-2 logarithm of x.
// Log2(x) - returns base-2 logarithm as Float (always returns Float, even for Integer input)
func Log2(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Log2() expects exactly 1 argument, got %d", len(args))
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
		return ctx.NewError("Log2() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Check for non-positive numbers (Log2 is undefined for x <= 0)
	if value <= 0 {
		return ctx.NewError("Log2() of non-positive number (%f)", value)
	}

	return &runtime.FloatValue{Value: math.Log2(value)}
}

// Log10 implements the Log10() built-in function.
// It returns the base-10 logarithm.
// Log10(x: Float): Float
func Log10(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Log10() expects exactly 1 argument, got %d", len(args))
	}

	var floatVal float64
	switch v := args[0].(type) {
	case *runtime.FloatValue:
		floatVal = v.Value
	case *runtime.IntegerValue:
		floatVal = float64(v.Value)
	default:
		return ctx.NewError("Log10() expects Float or Integer, got %s", args[0].Type())
	}

	if floatVal <= 0 {
		return ctx.NewError("Log10() argument must be positive, got %f", floatVal)
	}

	return &runtime.FloatValue{Value: math.Log10(floatVal)}
}

// LogN implements the LogN() built-in function.
// It returns the logarithm with a custom base.
// LogN(x, base: Float): Float
func LogN(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("LogN() expects exactly 2 arguments, got %d", len(args))
	}

	var xVal, baseVal float64
	switch v := args[0].(type) {
	case *runtime.FloatValue:
		xVal = v.Value
	case *runtime.IntegerValue:
		xVal = float64(v.Value)
	default:
		return ctx.NewError("LogN() expects Float or Integer as first argument, got %s", args[0].Type())
	}

	switch v := args[1].(type) {
	case *runtime.FloatValue:
		baseVal = v.Value
	case *runtime.IntegerValue:
		baseVal = float64(v.Value)
	default:
		return ctx.NewError("LogN() expects Float or Integer as second argument, got %s", args[1].Type())
	}

	if xVal <= 0 {
		return ctx.NewError("LogN() first argument must be positive, got %f", xVal)
	}
	if baseVal <= 0 || baseVal == 1 {
		return ctx.NewError("LogN() base must be positive and not equal to 1, got %f", baseVal)
	}

	// LogN(x, base) = Log(x) / Log(base)
	return &runtime.FloatValue{Value: math.Log(xVal) / math.Log(baseVal)}
}

// Unsigned32 implements the Unsigned32() built-in function.
// It converts an Integer to its unsigned 32-bit representation.
// Unsigned32(x) - converts Integer to unsigned 32-bit value (wraps around)
func Unsigned32(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Unsigned32() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Only accept Integer argument
	intVal, ok := arg.(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("Unsigned32() expects Integer as argument, got %s", arg.Type())
	}

	// Convert to uint32 (truncates to lower 32 bits) then back to int64
	// This mimics Cardinal(value) in Delphi: wraps around within uint32 range
	result := int64(uint32(intVal.Value))
	return &runtime.IntegerValue{Value: result}
}

// MaxInt implements the MaxInt() built-in function.
// It returns the maximum integer constant or the maximum of two integers.
// MaxInt() - returns math.MaxInt64 (9223372036854775807)
// MaxInt(a, b) - returns maximum of two Integer values
func MaxInt(ctx Context, args []Value) Value {
	// No arguments - return maximum integer constant
	if len(args) == 0 {
		return &runtime.IntegerValue{Value: math.MaxInt64}
	}

	// Two arguments - return maximum of two integers
	if len(args) == 2 {
		left, ok1 := args[0].(*runtime.IntegerValue)
		right, ok2 := args[1].(*runtime.IntegerValue)

		if !ok1 || !ok2 {
			return ctx.NewError("MaxInt() expects Integer arguments, got %s and %s", args[0].Type(), args[1].Type())
		}

		if left.Value > right.Value {
			return left
		}
		return right
	}

	// Invalid number of arguments
	return ctx.NewError("MaxInt() expects 0 or 2 arguments, got %d", len(args))
}

// MinInt implements the MinInt() built-in function.
// It returns the minimum integer constant or the minimum of two integers.
// MinInt() - returns math.MinInt64 (-9223372036854775808)
// MinInt(a, b) - returns minimum of two Integer values
func MinInt(ctx Context, args []Value) Value {
	// No arguments - return minimum integer constant
	if len(args) == 0 {
		return &runtime.IntegerValue{Value: math.MinInt64}
	}

	// Two arguments - return minimum of two integers
	if len(args) == 2 {
		left, ok1 := args[0].(*runtime.IntegerValue)
		right, ok2 := args[1].(*runtime.IntegerValue)

		if !ok1 || !ok2 {
			return ctx.NewError("MinInt() expects Integer arguments, got %s and %s", args[0].Type(), args[1].Type())
		}

		if left.Value < right.Value {
			return left
		}
		return right
	}

	// Invalid number of arguments
	return ctx.NewError("MinInt() expects 0 or 2 arguments, got %d", len(args))
}

// IsNaN implements the IsNaN() built-in function.
// It checks if a float value is NaN (Not a Number).
// IsNaN(value: Float): Boolean
func IsNaN(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("IsNaN() expects exactly 1 argument, got %d", len(args))
	}

	// Argument must be Float
	floatVal, ok := args[0].(*runtime.FloatValue)
	if !ok {
		// If not a float, it's not NaN
		return &runtime.BooleanValue{Value: false}
	}

	// Check if the value is NaN
	return &runtime.BooleanValue{Value: math.IsNaN(floatVal.Value)}
}

// Pi returns the mathematical constant π (Pi).
// Pi: Float
func Pi(ctx Context, args []Value) Value {
	if len(args) != 0 {
		return ctx.NewError("Pi expects no arguments, got %d", len(args))
	}
	return &runtime.FloatValue{Value: math.Pi}
}

// Sign implements the Sign() built-in function.
// It returns -1, 0, or 1 based on the sign of the number.
// Sign(x: Float): Integer
func Sign(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Sign() expects exactly 1 argument, got %d", len(args))
	}

	var floatVal float64
	switch v := args[0].(type) {
	case *runtime.FloatValue:
		floatVal = v.Value
	case *runtime.IntegerValue:
		floatVal = float64(v.Value)
	default:
		return ctx.NewError("Sign() expects Float or Integer, got %s", args[0].Type())
	}

	if floatVal > 0 {
		return &runtime.IntegerValue{Value: 1}
	} else if floatVal < 0 {
		return &runtime.IntegerValue{Value: -1}
	}
	return &runtime.IntegerValue{Value: 0}
}

// Odd implements the Odd() built-in function.
// It checks if an integer is odd.
// Odd(x: Integer): Boolean
func Odd(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Odd() expects exactly 1 argument, got %d", len(args))
	}

	intVal, ok := args[0].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("Odd() expects Integer, got %s", args[0].Type())
	}

	return &runtime.BooleanValue{Value: intVal.Value%2 != 0}
}

// Infinity returns the Infinity constant.
// Infinity: Float
func Infinity(ctx Context, args []Value) Value {
	if len(args) != 0 {
		return ctx.NewError("Infinity expects no arguments, got %d", len(args))
	}
	return &runtime.FloatValue{Value: math.Inf(1)}
}

// NaN returns the NaN (Not-a-Number) constant.
// NaN: Float
func NaN(ctx Context, args []Value) Value {
	if len(args) != 0 {
		return ctx.NewError("NaN expects no arguments, got %d", len(args))
	}
	return &runtime.FloatValue{Value: math.NaN()}
}

// IsFinite implements the IsFinite() built-in function.
// It checks if a number is finite (not infinite and not NaN).
// IsFinite(x: Float): Boolean
func IsFinite(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("IsFinite() expects exactly 1 argument, got %d", len(args))
	}

	var floatVal float64
	switch v := args[0].(type) {
	case *runtime.FloatValue:
		floatVal = v.Value
	case *runtime.IntegerValue:
		floatVal = float64(v.Value)
	default:
		return ctx.NewError("IsFinite() expects Float or Integer, got %s", args[0].Type())
	}

	// A number is finite if it's not infinite and not NaN
	isFinite := !math.IsInf(floatVal, 0) && !math.IsNaN(floatVal)
	return &runtime.BooleanValue{Value: isFinite}
}

// IsInfinite implements the IsInfinite() built-in function.
// It checks if a number is infinite.
// IsInfinite(x: Float): Boolean
func IsInfinite(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("IsInfinite() expects exactly 1 argument, got %d", len(args))
	}

	var floatVal float64
	switch v := args[0].(type) {
	case *runtime.FloatValue:
		floatVal = v.Value
	case *runtime.IntegerValue:
		floatVal = float64(v.Value)
	default:
		return ctx.NewError("IsInfinite() expects Float or Integer, got %s", args[0].Type())
	}

	return &runtime.BooleanValue{Value: math.IsInf(floatVal, 0)}
}

// =============================================================================
// Random Number Functions
// NOTE: These functions require access to the random number generator and seed.
// The Context interface needs to be extended with:
//   - RandSource() *rand.Rand (or similar) to access the RNG
//   - GetRandSeed() int64 to get the current seed
//   - SetRandSeed(seed int64) to set the seed
// =============================================================================

// Random implements the Random() built-in function.
// It returns a random Float between 0.0 (inclusive) and 1.0 (exclusive).
// Random() - returns random Float in [0, 1)
// TODO: Requires Context.RandSource() method
func Random(ctx Context, args []Value) Value {
	if len(args) != 0 {
		return ctx.NewError("Random() expects no arguments, got %d", len(args))
	}

	// TODO: Need to access random source from context
	// return &runtime.FloatValue{Value: ctx.RandSource().Float64()}
	panic("Random() not yet implemented - requires Context.RandSource() method")
}

// Randomize implements the Randomize() built-in procedure.
// It seeds the random number generator with the current time.
// Randomize() - seeds RNG with current time (no return value)
// TODO: Requires Context.SetRandSeed() method
func Randomize(ctx Context, args []Value) Value {
	if len(args) != 0 {
		return ctx.NewError("Randomize() expects no arguments, got %d", len(args))
	}

	// TODO: Need to set random seed in context
	// seed := time.Now().UnixNano()
	// ctx.SetRandSeed(seed)
	// return &runtime.NilValue{}
	panic("Randomize() not yet implemented - requires Context.SetRandSeed() method")
}

// RandomInt implements the RandomInt() built-in function.
// It returns a random integer between 0 (inclusive) and max (exclusive).
// RandomInt(max) - max must be positive
// TODO: Requires Context.RandSource() method
func RandomInt(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("RandomInt() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Only accept Integer argument
	intVal, ok := arg.(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("RandomInt() expects Integer as argument, got %s", arg.Type())
	}

	max := intVal.Value

	// Validate max > 0
	if max <= 0 {
		return ctx.NewError("RandomInt() expects max > 0, got %d", max)
	}

	// TODO: Need to access random source from context
	// randomValue := ctx.RandSource().Intn(int(max))
	// return &runtime.IntegerValue{Value: int64(randomValue)}
	panic("RandomInt() not yet implemented - requires Context.RandSource() method")
}

// SetRandSeed implements the SetRandSeed() built-in function.
// It sets the seed for the random number generator.
// SetRandSeed(seed: Integer)
// TODO: Requires Context.SetRandSeed() method
func SetRandSeed(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("SetRandSeed() expects exactly 1 argument, got %d", len(args))
	}

	// Argument must be Integer
	seedVal, ok := args[0].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("SetRandSeed() expects Integer, got %s", args[0].Type())
	}

	// TODO: Need to set random seed in context
	// ctx.SetRandSeed(seedVal.Value)
	// return &runtime.NilValue{}
	_ = seedVal // avoid unused variable error
	panic("SetRandSeed() not yet implemented - requires Context.SetRandSeed() method")
}

// RandSeed implements the RandSeed() built-in function.
// It returns the current random seed value.
// RandSeed: Integer
// TODO: Requires Context.GetRandSeed() method
func RandSeed(ctx Context, args []Value) Value {
	if len(args) != 0 {
		return ctx.NewError("RandSeed expects no arguments, got %d", len(args))
	}

	// TODO: Need to get random seed from context
	// return &runtime.IntegerValue{Value: ctx.GetRandSeed()}
	panic("RandSeed() not yet implemented - requires Context.GetRandSeed() method")
}

// RandG implements the RandG() built-in function.
// It returns a Gaussian (normal) distributed random number with mean=0 and stddev=1.
// Uses the Box-Muller transform.
// RandG(): Float
// TODO: Requires Context.RandSource() method
func RandG(ctx Context, args []Value) Value {
	if len(args) != 0 {
		return ctx.NewError("RandG() expects no arguments, got %d", len(args))
	}

	// TODO: Need to access random source from context
	// Box-Muller transform to generate Gaussian distributed random numbers
	// Generate two uniform random numbers in (0, 1]
	// rng := ctx.RandSource()
	// u1 := rng.Float64()
	// u2 := rng.Float64()
	//
	// // Ensure u1 is not zero or near-zero to avoid log(0)
	// if u1 < 1e-10 {
	// 	u1 = 1e-10
	// }
	//
	// // Box-Muller transform
	// z0 := math.Sqrt(-2.0*math.Log(u1)) * math.Cos(2.0*math.Pi*u2)
	//
	// return &runtime.FloatValue{Value: z0}
	panic("RandG() not yet implemented - requires Context.RandSource() method")
}

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
	arg := unwrapVariant(args[0])
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

// =============================================================================
// Advanced Math Functions
// =============================================================================

// Factorial implements the Factorial() built-in function.
// It computes the factorial of a non-negative integer: n! = n * (n-1) * ... * 1
// Factorial(n: Integer): Integer
func Factorial(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("Factorial() expects exactly 1 argument, got %d", len(args))
	}

	intVal, ok := args[0].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("Factorial() expects Integer, got %s", args[0].Type())
	}

	n := intVal.Value

	// Validate n >= 0
	if n < 0 {
		return ctx.NewError("Factorial() expects non-negative integer, got %d", n)
	}

	// Handle overflow: 20! is the largest factorial that fits in int64
	// 21! = 51090942171709440000 which overflows int64
	if n > 20 {
		return ctx.NewError("Factorial() overflow: %d! is too large for Integer", n)
	}

	// Calculate factorial
	result := int64(1)
	for j := int64(2); j <= n; j++ {
		result *= j
	}

	return &runtime.IntegerValue{Value: result}
}

// Gcd implements the Gcd() built-in function.
// It computes the Greatest Common Divisor using Euclidean algorithm.
// Gcd(a, b: Integer): Integer
func Gcd(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("Gcd() expects exactly 2 arguments, got %d", len(args))
	}

	aVal, ok1 := args[0].(*runtime.IntegerValue)
	bVal, ok2 := args[1].(*runtime.IntegerValue)

	if !ok1 || !ok2 {
		return ctx.NewError("Gcd() expects Integer arguments, got %s and %s", args[0].Type(), args[1].Type())
	}

	a := aVal.Value
	b := bVal.Value

	// Take absolute values
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}

	// Euclidean algorithm
	for b != 0 {
		a, b = b, a%b
	}

	return &runtime.IntegerValue{Value: a}
}

// Lcm implements the Lcm() built-in function.
// It computes the Least Common Multiple: lcm(a, b) = |a * b| / gcd(a, b)
// Lcm(a, b: Integer): Integer
func Lcm(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("Lcm() expects exactly 2 arguments, got %d", len(args))
	}

	aVal, ok1 := args[0].(*runtime.IntegerValue)
	bVal, ok2 := args[1].(*runtime.IntegerValue)

	if !ok1 || !ok2 {
		return ctx.NewError("Lcm() expects Integer arguments, got %s and %s", args[0].Type(), args[1].Type())
	}

	a := aVal.Value
	b := bVal.Value

	// Handle special case: if either is 0, result is 0
	if a == 0 || b == 0 {
		return &runtime.IntegerValue{Value: 0}
	}

	// Take absolute values
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}

	// Compute GCD using Euclidean algorithm
	gcdA, gcdB := a, b
	for gcdB != 0 {
		gcdA, gcdB = gcdB, gcdA%gcdB
	}

	// LCM = |a * b| / gcd(a, b)
	// To avoid overflow, compute (a / gcd) * b
	result := (a / gcdA) * b

	return &runtime.IntegerValue{Value: result}
}

// IsPrime implements the IsPrime() built-in function.
// It checks if a number is prime using trial division.
// IsPrime(n: Integer): Boolean
func IsPrime(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("IsPrime() expects exactly 1 argument, got %d", len(args))
	}

	intVal, ok := args[0].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("IsPrime() expects Integer, got %s", args[0].Type())
	}

	n := intVal.Value

	// Numbers less than 2 are not prime
	if n < 2 {
		return &runtime.BooleanValue{Value: false}
	}

	// 2 and 3 are prime
	if n == 2 || n == 3 {
		return &runtime.BooleanValue{Value: true}
	}

	// Even numbers (except 2) are not prime
	if n%2 == 0 {
		return &runtime.BooleanValue{Value: false}
	}

	// Multiples of 3 (except 3) are not prime
	if n%3 == 0 {
		return &runtime.BooleanValue{Value: false}
	}

	// Check divisibility by numbers of form 6k ± 1 up to √n
	// This is an optimization: all primes > 3 are of form 6k ± 1
	sqrtN := int64(math.Sqrt(float64(n)))
	for i := int64(5); i <= sqrtN; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return &runtime.BooleanValue{Value: false}
		}
	}

	return &runtime.BooleanValue{Value: true}
}

// LeastFactor implements the LeastFactor() built-in function.
// It finds the smallest prime factor of n.
// LeastFactor(n: Integer): Integer
func LeastFactor(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("LeastFactor() expects exactly 1 argument, got %d", len(args))
	}

	intVal, ok := args[0].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("LeastFactor() expects Integer, got %s", args[0].Type())
	}

	n := intVal.Value

	// Handle special cases
	if n <= 1 {
		return &runtime.IntegerValue{Value: 1}
	}

	// Check for divisibility by 2
	if n%2 == 0 {
		return &runtime.IntegerValue{Value: 2}
	}

	// Check for divisibility by 3
	if n%3 == 0 {
		return &runtime.IntegerValue{Value: 3}
	}

	// Check divisibility by numbers of form 6k ± 1 up to √n
	sqrtN := int64(math.Sqrt(float64(n)))
	for i := int64(5); i <= sqrtN; i += 6 {
		if n%i == 0 {
			return &runtime.IntegerValue{Value: i}
		}
		if n%(i+2) == 0 {
			return &runtime.IntegerValue{Value: i + 2}
		}
	}

	// If no factor found, n is prime
	return &runtime.IntegerValue{Value: n}
}

// PopCount implements the PopCount() built-in function.
// It counts the number of set bits (1s) in the binary representation.
// PopCount(n: Integer): Integer
func PopCount(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("PopCount() expects exactly 1 argument, got %d", len(args))
	}

	intVal, ok := args[0].(*runtime.IntegerValue)
	if !ok {
		return ctx.NewError("PopCount() expects Integer, got %s", args[0].Type())
	}

	// Use Go's built-in bits.OnesCount64 for unsigned 64-bit integers
	// Convert to uint64 to count bits
	count := bits.OnesCount64(uint64(intVal.Value))

	return &runtime.IntegerValue{Value: int64(count)}
}

// TestBit implements the TestBit() built-in function.
// It tests if a specific bit is set in a number.
// TestBit(value: Integer, bit: Integer): Boolean
func TestBit(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("TestBit() expects exactly 2 arguments, got %d", len(args))
	}

	valueVal, ok1 := args[0].(*runtime.IntegerValue)
	bitVal, ok2 := args[1].(*runtime.IntegerValue)

	if !ok1 || !ok2 {
		return ctx.NewError("TestBit() expects Integer arguments, got %s and %s", args[0].Type(), args[1].Type())
	}

	value := valueVal.Value
	bit := bitVal.Value

	// Validate bit position (0-63 for int64)
	if bit < 0 || bit >= 64 {
		return ctx.NewError("TestBit() bit position must be in range 0-63, got %d", bit)
	}

	// Test the bit: (value >> bit) & 1
	isSet := (value >> uint(bit)) & 1
	return &runtime.BooleanValue{Value: isSet != 0}
}

// Haversine implements the Haversine() built-in function.
// It calculates the great-circle distance between two points on a sphere
// given their latitudes and longitudes in degrees.
// Result is in kilometers (Earth radius = 6371 km).
// Haversine(lat1, lon1, lat2, lon2: Float): Float
func Haversine(ctx Context, args []Value) Value {
	if len(args) != 4 {
		return ctx.NewError("Haversine() expects exactly 4 arguments, got %d", len(args))
	}

	// Extract and convert all arguments to float64
	var lat1, lon1, lat2, lon2 float64

	// First argument (lat1)
	switch v := args[0].(type) {
	case *runtime.FloatValue:
		lat1 = v.Value
	case *runtime.IntegerValue:
		lat1 = float64(v.Value)
	default:
		return ctx.NewError("Haversine() expects Float or Integer arguments, got %s", args[0].Type())
	}

	// Second argument (lon1)
	switch v := args[1].(type) {
	case *runtime.FloatValue:
		lon1 = v.Value
	case *runtime.IntegerValue:
		lon1 = float64(v.Value)
	default:
		return ctx.NewError("Haversine() expects Float or Integer arguments, got %s", args[1].Type())
	}

	// Third argument (lat2)
	switch v := args[2].(type) {
	case *runtime.FloatValue:
		lat2 = v.Value
	case *runtime.IntegerValue:
		lat2 = float64(v.Value)
	default:
		return ctx.NewError("Haversine() expects Float or Integer arguments, got %s", args[2].Type())
	}

	// Fourth argument (lon2)
	switch v := args[3].(type) {
	case *runtime.FloatValue:
		lon2 = v.Value
	case *runtime.IntegerValue:
		lon2 = float64(v.Value)
	default:
		return ctx.NewError("Haversine() expects Float or Integer arguments, got %s", args[3].Type())
	}

	// Convert degrees to radians
	const degToRad = math.Pi / 180.0
	lat1Rad := lat1 * degToRad
	lon1Rad := lon1 * degToRad
	lat2Rad := lat2 * degToRad
	lon2Rad := lon2 * degToRad

	// Haversine formula
	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Earth radius in kilometers
	const earthRadiusKm = 6371.0
	distance := earthRadiusKm * c

	return &runtime.FloatValue{Value: distance}
}

// CompareNum implements the CompareNum() built-in function.
// It compares two numbers and returns -1, 0, or 1.
// CompareNum(a, b: Float): Integer
func CompareNum(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("CompareNum() expects exactly 2 arguments, got %d", len(args))
	}

	// Extract first argument
	var a float64
	switch v := args[0].(type) {
	case *runtime.FloatValue:
		a = v.Value
	case *runtime.IntegerValue:
		a = float64(v.Value)
	default:
		return ctx.NewError("CompareNum() expects Float or Integer arguments, got %s", args[0].Type())
	}

	// Extract second argument
	var b float64
	switch v := args[1].(type) {
	case *runtime.FloatValue:
		b = v.Value
	case *runtime.IntegerValue:
		b = float64(v.Value)
	default:
		return ctx.NewError("CompareNum() expects Float or Integer arguments, got %s", args[1].Type())
	}

	// Handle NaN: NaN is considered equal to NaN, and less than all other values
	aIsNaN := math.IsNaN(a)
	bIsNaN := math.IsNaN(b)

	if aIsNaN && bIsNaN {
		return &runtime.IntegerValue{Value: 0} // Both NaN, equal
	}
	if aIsNaN {
		return &runtime.IntegerValue{Value: -1} // NaN is less than any number
	}
	if bIsNaN {
		return &runtime.IntegerValue{Value: 1} // Any number is greater than NaN
	}

	// Regular comparison
	if a < b {
		return &runtime.IntegerValue{Value: -1}
	} else if a > b {
		return &runtime.IntegerValue{Value: 1}
	}
	return &runtime.IntegerValue{Value: 0}
}

// =============================================================================
// Special Functions (Skip for now)
// =============================================================================

// TODO: DivMod requires special handling as it takes []ast.Expression instead of []Value
// This function should remain in the Interpreter for now and be migrated separately.
// See builtins_math_convert.go:builtinDivMod for the original implementation.

// =============================================================================
// Helper Functions
// =============================================================================

// unwrapVariant unwraps a VariantValue to its underlying value.
// If the value is not a VariantValue, it returns the value unchanged.
// This is needed for Task 9.4.5: Variant support in built-in functions.
func unwrapVariant(v Value) Value {
	// TODO: Implement variant unwrapping when VariantValue is available in runtime package
	// For now, just return the value unchanged
	return v
}
