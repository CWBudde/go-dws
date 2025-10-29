package interp

import (
	"math"
	"math/rand"
	"time"
)

// builtinAbs implements the Abs() built-in function.
// It returns the absolute value of a number.
// Abs(x) - returns absolute value (Integer → Integer, Float → Float)
// Task 8.185: Abs() function for math operations
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
// Task 8.185: Min() function for math operations
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
// Task 8.185: Max() function for math operations
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
// Task 8.185: Sqr() function for math operations
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
// Task 8.185: Power() function for math operations
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
// Task 8.185: Sqrt() function for math operations
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

// builtinSin implements the Sin() built-in function.
// It returns the sine of a number (in radians).
// Sin(x) - returns sine as Float (always returns Float, even for Integer input)
// Task 8.185: Sin() function for trigonometric operations
func (i *Interpreter) builtinSin(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Sin() expects exactly 1 argument, got %d", len(args))
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
		return i.newErrorWithLocation(i.currentNode, "Sin() expects Integer or Float as argument, got %s", arg.Type())
	}

	return &FloatValue{Value: math.Sin(value)}
}

// builtinCos implements the Cos() built-in function.
// It returns the cosine of a number (in radians).
// Cos(x) - returns cosine as Float (always returns Float, even for Integer input)
// Task 8.185: Cos() function for trigonometric operations
func (i *Interpreter) builtinCos(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Cos() expects exactly 1 argument, got %d", len(args))
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
		return i.newErrorWithLocation(i.currentNode, "Cos() expects Integer or Float as argument, got %s", arg.Type())
	}

	return &FloatValue{Value: math.Cos(value)}
}

// builtinTan implements the Tan() built-in function.
// It returns the tangent of a number (in radians).
// Tan(x) - returns tangent as Float (always returns Float, even for Integer input)
// Task 8.185: Tan() function for trigonometric operations
func (i *Interpreter) builtinTan(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Tan() expects exactly 1 argument, got %d", len(args))
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
		return i.newErrorWithLocation(i.currentNode, "Tan() expects Integer or Float as argument, got %s", arg.Type())
	}

	return &FloatValue{Value: math.Tan(value)}
}

// builtinRandom implements the Random() built-in function.
// It returns a random Float between 0.0 (inclusive) and 1.0 (exclusive).
// Random() - returns random Float in [0, 1)
// Task 8.185: Random() function for random number generation
func (i *Interpreter) builtinRandom(args []Value) Value {
	if len(args) != 0 {
		return i.newErrorWithLocation(i.currentNode, "Random() expects no arguments, got %d", len(args))
	}

	return &FloatValue{Value: i.rand.Float64()}
}

// builtinRandomize implements the Randomize() built-in procedure.
// It seeds the random number generator with the current time.
// Randomize() - seeds RNG with current time (no return value)
// Task 8.185: Randomize() procedure for random number generation
func (i *Interpreter) builtinRandomize(args []Value) Value {
	if len(args) != 0 {
		return i.newErrorWithLocation(i.currentNode, "Randomize() expects no arguments, got %d", len(args))
	}

	// Re-seed the random number generator with current time
	i.rand.Seed(time.Now().UnixNano())
	return &NilValue{}
}

// builtinExp implements the Exp() built-in function.
// It returns e raised to the power of x.
// Exp(x) - returns e^x as Float (always returns Float, even for Integer input)
// Task 8.185: Exp() function for exponential operations
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
// Task 8.185: Ln() function for logarithmic operations
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

// builtinRound implements the Round() built-in function.
// It rounds a number to the nearest integer.
// Round(x) - returns rounded value as Integer (always returns Integer)
// Task 8.185: Round() function for rounding operations
func (i *Interpreter) builtinRound(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Round() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	var value float64

	// Handle Integer - already an integer, just return it
	if intVal, ok := arg.(*IntegerValue); ok {
		return &IntegerValue{Value: intVal.Value}
	} else if floatVal, ok := arg.(*FloatValue); ok {
		// Handle Float
		value = floatVal.Value
	} else {
		return i.newErrorWithLocation(i.currentNode, "Round() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Round to nearest integer
	rounded := math.Round(value)
	return &IntegerValue{Value: int64(rounded)}
}

// builtinTrunc implements the Trunc() built-in function.
// It truncates a number towards zero (removes the decimal part).
// Trunc(x) - returns truncated value as Integer (always returns Integer)
// Task 8.185: Trunc() function for truncation operations
func (i *Interpreter) builtinTrunc(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Trunc() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	var value float64

	// Handle Integer - already an integer, just return it
	if intVal, ok := arg.(*IntegerValue); ok {
		return &IntegerValue{Value: intVal.Value}
	} else if floatVal, ok := arg.(*FloatValue); ok {
		// Handle Float
		value = floatVal.Value
	} else {
		return i.newErrorWithLocation(i.currentNode, "Trunc() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Truncate toward zero
	truncated := math.Trunc(value)
	return &IntegerValue{Value: int64(truncated)}
}

// builtinCeil implements the Ceil() built-in function.
// It returns the smallest integer greater than or equal to x.
// Ceil(x) - returns ceiling value as Integer (always returns Integer)
// Task 8.185: Ceil() function for rounding operations
func (i *Interpreter) builtinCeil(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Ceil() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	var value float64

	// Handle Integer - already an integer, just return it
	if intVal, ok := arg.(*IntegerValue); ok {
		return &IntegerValue{Value: intVal.Value}
	} else if floatVal, ok := arg.(*FloatValue); ok {
		// Handle Float
		value = floatVal.Value
	} else {
		return i.newErrorWithLocation(i.currentNode, "Ceil() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Ceil returns float64 - convert to integer
	ceiling := math.Ceil(value)
	return &IntegerValue{Value: int64(ceiling)}
}

// builtinFloor implements the Floor() built-in function.
// It returns the largest integer less than or equal to x.
// Floor(x) - returns floor value as Integer (always returns Integer)
// Task 8.185: Floor() function for rounding operations
func (i *Interpreter) builtinFloor(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Floor() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	var value float64

	// Handle Integer - already an integer, just return it
	if intVal, ok := arg.(*IntegerValue); ok {
		return &IntegerValue{Value: intVal.Value}
	} else if floatVal, ok := arg.(*FloatValue); ok {
		// Handle Float
		value = floatVal.Value
	} else {
		return i.newErrorWithLocation(i.currentNode, "Floor() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Floor returns float64 - convert to integer
	floor := math.Floor(value)
	return &IntegerValue{Value: int64(floor)}
}

// builtinRandomInt implements the RandomInt() built-in function.
// It returns a random integer between 0 (inclusive) and max (exclusive).
// RandomInt(max) - max must be positive
// Task 8.185: RandomInt() function for random number generation
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
