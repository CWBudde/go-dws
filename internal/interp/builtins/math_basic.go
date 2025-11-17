package builtins

import (
	"math"

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
