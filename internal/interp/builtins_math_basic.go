package interp

import (
	"math"
	"math/rand"
	"time"
)

// builtinAbs implements the Abs() built-in function.
// It returns the absolute value of a number.
// Abs(x) - returns absolute value (Integer → Integer, Float → Float)
func (i *Interpreter) builtinAbs(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Abs() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	switch v := arg.(type) {
	case *IntegerValue:
		if v.Value < 0 {
			return &IntegerValue{Value: -v.Value}
		}
		return v
	case *FloatValue:
		return &FloatValue{Value: math.Abs(v.Value)}
	default:
		return i.newErrorWithLocation(i.currentNode, "Abs() expects Integer or Float, got %s", arg.Type())
	}
}

// builtinMin implements the Min() built-in function.
// It returns the minimum of two numbers.
// Min(a, b) - supports Integer and Float (mixed allowed)
func (i *Interpreter) builtinMin(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "Min() expects exactly 2 arguments, got %d", len(args))
	}

	left := args[0]
	right := args[1]

	switch l := left.(type) {
	case *IntegerValue:
		// Integer-Integer
		if r, ok := right.(*IntegerValue); ok {
			if l.Value < r.Value {
				return l
			}
			return r
		}
		// Integer-Float (promote to float)
		if r, ok := right.(*FloatValue); ok {
			leftFloat := float64(l.Value)
			if leftFloat < r.Value {
				return &FloatValue{Value: leftFloat}
			}
			return r
		}
	case *FloatValue:
		// Float-Float
		if r, ok := right.(*FloatValue); ok {
			if l.Value < r.Value {
				return l
			}
			return r
		}
		// Float-Integer (promote integer)
		if r, ok := right.(*IntegerValue); ok {
			rightFloat := float64(r.Value)
			if l.Value < rightFloat {
				return l
			}
			return &FloatValue{Value: rightFloat}
		}
	}

	return i.newErrorWithLocation(i.currentNode, "Min() expects Integer or Float arguments, got %s and %s", left.Type(), right.Type())
}

// builtinMax implements the Max() built-in function.
// It returns the maximum of two numbers.
// Max(a, b) - supports Integer and Float (mixed allowed)
func (i *Interpreter) builtinMax(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "Max() expects exactly 2 arguments, got %d", len(args))
	}

	left := args[0]
	right := args[1]

	switch l := left.(type) {
	case *IntegerValue:
		// Integer-Integer
		if r, ok := right.(*IntegerValue); ok {
			if l.Value > r.Value {
				return l
			}
			return r
		}
		// Integer-Float
		if r, ok := right.(*FloatValue); ok {
			leftFloat := float64(l.Value)
			if leftFloat > r.Value {
				return &FloatValue{Value: leftFloat}
			}
			return r
		}
	case *FloatValue:
		// Float-Float
		if r, ok := right.(*FloatValue); ok {
			if l.Value > r.Value {
				return l
			}
			return r
		}
		// Float-Integer
		if r, ok := right.(*IntegerValue); ok {
			rightFloat := float64(r.Value)
			if l.Value > rightFloat {
				return l
			}
			return &FloatValue{Value: rightFloat}
		}
	}

	return i.newErrorWithLocation(i.currentNode, "Max() expects Integer or Float arguments, got %s and %s", left.Type(), right.Type())
}

// builtinSqr implements the Sqr() built-in function.
// It returns the square of a number.
// Sqr(x) - returns x * x (Integer → Integer, Float → Float)
func (i *Interpreter) builtinSqr(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Sqr() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	switch v := arg.(type) {
	case *IntegerValue:
		return &IntegerValue{Value: v.Value * v.Value}
	case *FloatValue:
		return &FloatValue{Value: v.Value * v.Value}
	default:
		return i.newErrorWithLocation(i.currentNode, "Sqr() expects Integer or Float, got %s", arg.Type())
	}
}

// builtinPower implements the Power() built-in function.
// It raises a number to a power.
// Power(base, exponent) - returns base^exponent as Float
func (i *Interpreter) builtinPower(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "Power() expects exactly 2 arguments, got %d", len(args))
	}

	base := args[0]
	exponent := args[1]

	var baseFloat, exponentFloat float64

	switch v := base.(type) {
	case *IntegerValue:
		baseFloat = float64(v.Value)
	case *FloatValue:
		baseFloat = v.Value
	default:
		return i.newErrorWithLocation(i.currentNode, "Power() expects Integer or Float as base, got %s", base.Type())
	}

	switch v := exponent.(type) {
	case *IntegerValue:
		exponentFloat = float64(v.Value)
	case *FloatValue:
		exponentFloat = v.Value
	default:
		return i.newErrorWithLocation(i.currentNode, "Power() expects Integer or Float as exponent, got %s", exponent.Type())
	}

	// Use math.Pow() - this handles all special cases including 0^0 = 1
	result := math.Pow(baseFloat, exponentFloat)
	return &FloatValue{Value: result}
}

// builtinSqrt implements the Sqrt() built-in function.
// It returns the square root of a number.
// Sqrt(x) - returns sqrt(x) as Float (always returns Float, even for Integer input)
func (i *Interpreter) builtinSqrt(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Sqrt() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	var value float64

	// Handle Integer - convert to float
	if intVal, ok := arg.(*IntegerValue); ok {
		if intVal.Value < 0 {
			return i.newErrorWithLocation(i.currentNode, "Sqrt() of negative number (%d)", intVal.Value)
		}
		value = float64(intVal.Value)
	} else if floatVal, ok := arg.(*FloatValue); ok {
		// Handle Float
		if floatVal.Value < 0 {
			return i.newErrorWithLocation(i.currentNode, "Sqrt() of negative number (%f)", floatVal.Value)
		}
		value = floatVal.Value
	} else {
		return i.newErrorWithLocation(i.currentNode, "Sqrt() expects Integer or Float as argument, got %s", arg.Type())
	}

	return &FloatValue{Value: math.Sqrt(value)}
}

// builtinRandom implements the Random() built-in function.
// It returns a random Float between 0.0 (inclusive) and 1.0 (exclusive).
// Random() - returns random Float in [0, 1)
func (i *Interpreter) builtinRandom(args []Value) Value {
	if len(args) != 0 {
		return i.newErrorWithLocation(i.currentNode, "Random() expects no arguments, got %d", len(args))
	}

	return &FloatValue{Value: i.rand.Float64()}
}

// builtinRandomize implements the Randomize() built-in procedure.
// It seeds the random number generator with the current time.
// Randomize() - seeds RNG with current time (no return value)
func (i *Interpreter) builtinRandomize(args []Value) Value {
	if len(args) != 0 {
		return i.newErrorWithLocation(i.currentNode, "Randomize() expects no arguments, got %d", len(args))
	}

	// Re-seed the random number generator with current time and store the seed
	seed := time.Now().UnixNano()
	i.rand.Seed(seed)
	i.randSeed = seed
	return &NilValue{}
}

// builtinExp implements the Exp() built-in function.
// It returns e raised to the power of x.
// Exp(x) - returns e^x as Float (always returns Float, even for Integer input)
func (i *Interpreter) builtinExp(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Exp() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	var value float64

	// Handle Integer - convert to float
	if intVal, ok := arg.(*IntegerValue); ok {
		value = float64(intVal.Value)
	} else if floatVal, ok := arg.(*FloatValue); ok {
		// Handle Float
		value = floatVal.Value
	} else {
		return i.newErrorWithLocation(i.currentNode, "Exp() expects Integer or Float as argument, got %s", arg.Type())
	}

	return &FloatValue{Value: math.Exp(value)}
}

// builtinLn implements the Ln() built-in function.
// It returns the natural logarithm (base e) of x.
// Ln(x) - returns natural logarithm as Float (always returns Float, even for Integer input)
func (i *Interpreter) builtinLn(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Ln() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	var value float64

	// Handle Integer - convert to float
	if intVal, ok := arg.(*IntegerValue); ok {
		value = float64(intVal.Value)
	} else if floatVal, ok := arg.(*FloatValue); ok {
		// Handle Float
		value = floatVal.Value
	} else {
		return i.newErrorWithLocation(i.currentNode, "Ln() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Check for non-positive numbers (Ln is undefined for x <= 0)
	if value <= 0 {
		return i.newErrorWithLocation(i.currentNode, "Ln() of non-positive number (%f)", value)
	}

	return &FloatValue{Value: math.Log(value)}
}

// builtinLog2 implements the Log2() built-in function.
// It returns the base-2 logarithm of x.
// Log2(x) - returns base-2 logarithm as Float (always returns Float, even for Integer input)
func (i *Interpreter) builtinLog2(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Log2() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	var value float64

	// Handle Integer - convert to float
	if intVal, ok := arg.(*IntegerValue); ok {
		value = float64(intVal.Value)
	} else if floatVal, ok := arg.(*FloatValue); ok {
		// Handle Float
		value = floatVal.Value
	} else {
		return i.newErrorWithLocation(i.currentNode, "Log2() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Check for non-positive numbers (Log2 is undefined for x <= 0)
	if value <= 0 {
		return i.newErrorWithLocation(i.currentNode, "Log2() of non-positive number (%f)", value)
	}

	return &FloatValue{Value: math.Log2(value)}
}

// builtinRandomInt implements the RandomInt() built-in function.
// It returns a random integer between 0 (inclusive) and max (exclusive).
// RandomInt(max) - max must be positive
func (i *Interpreter) builtinRandomInt(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "RandomInt() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Only accept Integer argument
	intVal, ok := arg.(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "RandomInt() expects Integer as argument, got %s", arg.Type())
	}

	max := intVal.Value

	// Validate max > 0
	if max <= 0 {
		return i.newErrorWithLocation(i.currentNode, "RandomInt() expects max > 0, got %d", max)
	}

	// Generate random integer in range [0, max)
	randomValue := rand.Intn(int(max))
	return &IntegerValue{Value: int64(randomValue)}
}

// builtinUnsigned32 implements the Unsigned32() built-in function.
// It converts an Integer to its unsigned 32-bit representation.
// Unsigned32(x) - converts Integer to unsigned 32-bit value (wraps around)
func (i *Interpreter) builtinUnsigned32(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Unsigned32() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Only accept Integer argument
	intVal, ok := arg.(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Unsigned32() expects Integer as argument, got %s", arg.Type())
	}

	// Convert to uint32 (truncates to lower 32 bits) then back to int64
	// This mimics Cardinal(value) in Delphi: wraps around within uint32 range
	result := int64(uint32(intVal.Value))
	return &IntegerValue{Value: result}
}

// builtinMaxInt implements the MaxInt() built-in function.
// It returns the maximum integer constant or the maximum of two integers.
// MaxInt() - returns math.MaxInt64 (9223372036854775807)
// MaxInt(a, b) - returns maximum of two Integer values
func (i *Interpreter) builtinMaxInt(args []Value) Value {
	// No arguments - return maximum integer constant
	if len(args) == 0 {
		return &IntegerValue{Value: math.MaxInt64}
	}

	// Two arguments - return maximum of two integers
	if len(args) == 2 {
		left, ok1 := args[0].(*IntegerValue)
		right, ok2 := args[1].(*IntegerValue)

		if !ok1 || !ok2 {
			return i.newErrorWithLocation(i.currentNode, "MaxInt() expects Integer arguments, got %s and %s", args[0].Type(), args[1].Type())
		}

		if left.Value > right.Value {
			return left
		}
		return right
	}

	// Invalid number of arguments
	return i.newErrorWithLocation(i.currentNode, "MaxInt() expects 0 or 2 arguments, got %d", len(args))
}

// builtinMinInt implements the MinInt() built-in function.
// It returns the minimum integer constant or the minimum of two integers.
// MinInt() - returns math.MinInt64 (-9223372036854775808)
// MinInt(a, b) - returns minimum of two Integer values
func (i *Interpreter) builtinMinInt(args []Value) Value {
	// No arguments - return minimum integer constant
	if len(args) == 0 {
		return &IntegerValue{Value: math.MinInt64}
	}

	// Two arguments - return minimum of two integers
	if len(args) == 2 {
		left, ok1 := args[0].(*IntegerValue)
		right, ok2 := args[1].(*IntegerValue)

		if !ok1 || !ok2 {
			return i.newErrorWithLocation(i.currentNode, "MinInt() expects Integer arguments, got %s and %s", args[0].Type(), args[1].Type())
		}

		if left.Value < right.Value {
			return left
		}
		return right
	}

	// Invalid number of arguments
	return i.newErrorWithLocation(i.currentNode, "MinInt() expects 0 or 2 arguments, got %d", len(args))
}

// builtinIsNaN implements the IsNaN() built-in function.
// It checks if a float value is NaN (Not a Number).
// IsNaN(value: Float): Boolean
func (i *Interpreter) builtinIsNaN(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "IsNaN() expects exactly 1 argument, got %d", len(args))
	}

	// Argument must be Float
	floatVal, ok := args[0].(*FloatValue)
	if !ok {
		// If not a float, it's not NaN
		return &BooleanValue{Value: false}
	}

	// Check if the value is NaN
	return &BooleanValue{Value: math.IsNaN(floatVal.Value)}
}

// builtinSetRandSeed implements the SetRandSeed() built-in function.
// It sets the seed for the random number generator.
// SetRandSeed(seed: Integer)
func (i *Interpreter) builtinSetRandSeed(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "SetRandSeed() expects exactly 1 argument, got %d", len(args))
	}

	// Argument must be Integer
	seedVal, ok := args[0].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "SetRandSeed() expects Integer, got %s", args[0].Type())
	}

	// Set the seed for the random number generator and store it
	i.rand.Seed(seedVal.Value)
	i.randSeed = seedVal.Value

	return &NilValue{}
}

// builtinRandSeed implements the RandSeed() built-in function.
// It returns the current random seed value.
// RandSeed: Integer
func (i *Interpreter) builtinRandSeed(args []Value) Value {
	if len(args) != 0 {
		return i.newErrorWithLocation(i.currentNode, "RandSeed expects no arguments, got %d", len(args))
	}
	return &IntegerValue{Value: i.randSeed}
}

// builtinPi returns the mathematical constant π (Pi).
// Pi: Float
func (i *Interpreter) builtinPi(args []Value) Value {
	if len(args) != 0 {
		return i.newErrorWithLocation(i.currentNode, "Pi expects no arguments, got %d", len(args))
	}
	return &FloatValue{Value: math.Pi}
}

// builtinSign implements the Sign() built-in function.
// It returns -1, 0, or 1 based on the sign of the number.
// Sign(x: Float): Integer
func (i *Interpreter) builtinSign(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Sign() expects exactly 1 argument, got %d", len(args))
	}

	var floatVal float64
	switch v := args[0].(type) {
	case *FloatValue:
		floatVal = v.Value
	case *IntegerValue:
		floatVal = float64(v.Value)
	default:
		return i.newErrorWithLocation(i.currentNode, "Sign() expects Float or Integer, got %s", args[0].Type())
	}

	if floatVal > 0 {
		return &IntegerValue{Value: 1}
	} else if floatVal < 0 {
		return &IntegerValue{Value: -1}
	}
	return &IntegerValue{Value: 0}
}

// builtinOdd implements the Odd() built-in function.
// It checks if an integer is odd.
// Odd(x: Integer): Boolean
func (i *Interpreter) builtinOdd(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Odd() expects exactly 1 argument, got %d", len(args))
	}

	intVal, ok := args[0].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Odd() expects Integer, got %s", args[0].Type())
	}

	return &BooleanValue{Value: intVal.Value%2 != 0}
}

// builtinLog10 implements the Log10() built-in function.
// It returns the base-10 logarithm.
// Log10(x: Float): Float
func (i *Interpreter) builtinLog10(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Log10() expects exactly 1 argument, got %d", len(args))
	}

	var floatVal float64
	switch v := args[0].(type) {
	case *FloatValue:
		floatVal = v.Value
	case *IntegerValue:
		floatVal = float64(v.Value)
	default:
		return i.newErrorWithLocation(i.currentNode, "Log10() expects Float or Integer, got %s", args[0].Type())
	}

	if floatVal <= 0 {
		return i.newErrorWithLocation(i.currentNode, "Log10() argument must be positive, got %f", floatVal)
	}

	return &FloatValue{Value: math.Log10(floatVal)}
}

// builtinLogN implements the LogN() built-in function.
// It returns the logarithm with a custom base.
// LogN(x, base: Float): Float
func (i *Interpreter) builtinLogN(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "LogN() expects exactly 2 arguments, got %d", len(args))
	}

	var xVal, baseVal float64
	switch v := args[0].(type) {
	case *FloatValue:
		xVal = v.Value
	case *IntegerValue:
		xVal = float64(v.Value)
	default:
		return i.newErrorWithLocation(i.currentNode, "LogN() expects Float or Integer as first argument, got %s", args[0].Type())
	}

	switch v := args[1].(type) {
	case *FloatValue:
		baseVal = v.Value
	case *IntegerValue:
		baseVal = float64(v.Value)
	default:
		return i.newErrorWithLocation(i.currentNode, "LogN() expects Float or Integer as second argument, got %s", args[1].Type())
	}

	if xVal <= 0 {
		return i.newErrorWithLocation(i.currentNode, "LogN() first argument must be positive, got %f", xVal)
	}
	if baseVal <= 0 || baseVal == 1 {
		return i.newErrorWithLocation(i.currentNode, "LogN() base must be positive and not equal to 1, got %f", baseVal)
	}

	// LogN(x, base) = Log(x) / Log(base)
	return &FloatValue{Value: math.Log(xVal) / math.Log(baseVal)}
}

// builtinInfinity returns the Infinity constant.
// Infinity: Float
func (i *Interpreter) builtinInfinity(args []Value) Value {
	if len(args) != 0 {
		return i.newErrorWithLocation(i.currentNode, "Infinity expects no arguments, got %d", len(args))
	}
	return &FloatValue{Value: math.Inf(1)}
}

// builtinNaN returns the NaN (Not-a-Number) constant.
// NaN: Float
func (i *Interpreter) builtinNaN(args []Value) Value {
	if len(args) != 0 {
		return i.newErrorWithLocation(i.currentNode, "NaN expects no arguments, got %d", len(args))
	}
	return &FloatValue{Value: math.NaN()}
}

// builtinIsFinite implements the IsFinite() built-in function.
// It checks if a number is finite (not infinite and not NaN).
// IsFinite(x: Float): Boolean
func (i *Interpreter) builtinIsFinite(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "IsFinite() expects exactly 1 argument, got %d", len(args))
	}

	var floatVal float64
	switch v := args[0].(type) {
	case *FloatValue:
		floatVal = v.Value
	case *IntegerValue:
		floatVal = float64(v.Value)
	default:
		return i.newErrorWithLocation(i.currentNode, "IsFinite() expects Float or Integer, got %s", args[0].Type())
	}

	// A number is finite if it's not infinite and not NaN
	isFinite := !math.IsInf(floatVal, 0) && !math.IsNaN(floatVal)
	return &BooleanValue{Value: isFinite}
}

// builtinIsInfinite implements the IsInfinite() built-in function.
// It checks if a number is infinite.
// IsInfinite(x: Float): Boolean
func (i *Interpreter) builtinIsInfinite(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "IsInfinite() expects exactly 1 argument, got %d", len(args))
	}

	var floatVal float64
	switch v := args[0].(type) {
	case *FloatValue:
		floatVal = v.Value
	case *IntegerValue:
		floatVal = float64(v.Value)
	default:
		return i.newErrorWithLocation(i.currentNode, "IsInfinite() expects Float or Integer, got %s", args[0].Type())
	}

	return &BooleanValue{Value: math.IsInf(floatVal, 0)}
}

// builtinRandG implements the RandG() built-in function.
// It returns a Gaussian (normal) distributed random number with mean=0 and stddev=1.
// Uses the Box-Muller transform.
// RandG(): Float
func (i *Interpreter) builtinRandG(args []Value) Value {
	if len(args) != 0 {
		return i.newErrorWithLocation(i.currentNode, "RandG() expects no arguments, got %d", len(args))
	}

	// Box-Muller transform to generate Gaussian distributed random numbers
	// Generate two uniform random numbers in (0, 1]
	u1 := i.rand.Float64()
	u2 := i.rand.Float64()

	// Ensure u1 is not zero or near-zero to avoid log(0)
	if u1 < 1e-10 {
		u1 = 1e-10
	}

	// Box-Muller transform
	z0 := math.Sqrt(-2.0*math.Log(u1)) * math.Cos(2.0*math.Pi*u2)

	return &FloatValue{Value: z0}
}
