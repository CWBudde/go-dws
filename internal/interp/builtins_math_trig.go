package interp

import (
	"math"
)

// builtinSin implements the Sin() built-in function.
// It returns the sine of a number (in radians).
// Sin(x) - returns sine as Float (always returns Float, even for Integer input)
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

// builtinDegToRad implements the DegToRad() built-in function.
// It converts degrees to radians.
// DegToRad(degrees) - returns radians as Float
func (i *Interpreter) builtinDegToRad(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "DegToRad() expects exactly 1 argument, got %d", len(args))
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
		return i.newErrorWithLocation(i.currentNode, "DegToRad() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Convert degrees to radians: radians = degrees * π / 180
	return &FloatValue{Value: value * math.Pi / 180.0}
}

// builtinRadToDeg implements the RadToDeg() built-in function.
// It converts radians to degrees.
// RadToDeg(radians) - returns degrees as Float
func (i *Interpreter) builtinRadToDeg(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "RadToDeg() expects exactly 1 argument, got %d", len(args))
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
		return i.newErrorWithLocation(i.currentNode, "RadToDeg() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Convert radians to degrees: degrees = radians * 180 / π
	return &FloatValue{Value: value * 180.0 / math.Pi}
}

// builtinArcSin implements the ArcSin() built-in function.
// It returns the arcsine (inverse sine) of a number.
// ArcSin(x) - returns arcsine in radians as Float (x must be in [-1, 1])
func (i *Interpreter) builtinArcSin(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "ArcSin() expects exactly 1 argument, got %d", len(args))
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
		return i.newErrorWithLocation(i.currentNode, "ArcSin() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Check domain: x must be in [-1, 1]
	if value < -1.0 || value > 1.0 {
		return i.newErrorWithLocation(i.currentNode, "ArcSin() argument must be in [-1, 1], got %f", value)
	}

	return &FloatValue{Value: math.Asin(value)}
}

// builtinArcCos implements the ArcCos() built-in function.
// It returns the arccosine (inverse cosine) of a number.
// ArcCos(x) - returns arccosine in radians as Float (x must be in [-1, 1])
func (i *Interpreter) builtinArcCos(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "ArcCos() expects exactly 1 argument, got %d", len(args))
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
		return i.newErrorWithLocation(i.currentNode, "ArcCos() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Check domain: x must be in [-1, 1]
	if value < -1.0 || value > 1.0 {
		return i.newErrorWithLocation(i.currentNode, "ArcCos() argument must be in [-1, 1], got %f", value)
	}

	return &FloatValue{Value: math.Acos(value)}
}

// builtinArcTan implements the ArcTan() built-in function.
// It returns the arctangent (inverse tangent) of a number.
// ArcTan(x) - returns arctangent in radians as Float
func (i *Interpreter) builtinArcTan(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "ArcTan() expects exactly 1 argument, got %d", len(args))
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
		return i.newErrorWithLocation(i.currentNode, "ArcTan() expects Integer or Float as argument, got %s", arg.Type())
	}

	return &FloatValue{Value: math.Atan(value)}
}

// builtinArcTan2 implements the ArcTan2() built-in function.
// It returns the arctangent of y/x, using the signs of both to determine the quadrant.
// ArcTan2(y, x) - returns arctangent in radians as Float
func (i *Interpreter) builtinArcTan2(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "ArcTan2() expects exactly 2 arguments, got %d", len(args))
	}

	y := args[0]
	x := args[1]

	var yFloat, xFloat float64

	// Handle y parameter
	if intVal, ok := y.(*IntegerValue); ok {
		yFloat = float64(intVal.Value)
	} else if floatVal, ok := y.(*FloatValue); ok {
		yFloat = floatVal.Value
	} else {
		return i.newErrorWithLocation(i.currentNode, "ArcTan2() expects Integer or Float as first argument, got %s", y.Type())
	}

	// Handle x parameter
	if intVal, ok := x.(*IntegerValue); ok {
		xFloat = float64(intVal.Value)
	} else if floatVal, ok := x.(*FloatValue); ok {
		xFloat = floatVal.Value
	} else {
		return i.newErrorWithLocation(i.currentNode, "ArcTan2() expects Integer or Float as second argument, got %s", x.Type())
	}

	return &FloatValue{Value: math.Atan2(yFloat, xFloat)}
}

// builtinCoTan implements the CoTan() built-in function.
// It returns the cotangent of a number (in radians).
// CoTan(x) - returns cotangent as Float (1 / Tan(x))
func (i *Interpreter) builtinCoTan(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "CoTan() expects exactly 1 argument, got %d", len(args))
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
		return i.newErrorWithLocation(i.currentNode, "CoTan() expects Integer or Float as argument, got %s", arg.Type())
	}

	// CoTan(x) = 1 / Tan(x) = Cos(x) / Sin(x)
	tan := math.Tan(value)
	if tan == 0 {
		return i.newErrorWithLocation(i.currentNode, "CoTan() undefined at x=%f (division by zero)", value)
	}

	return &FloatValue{Value: 1.0 / tan}
}

// builtinHypot implements the Hypot() built-in function.
// It returns the hypotenuse: sqrt(x² + y²)
// Hypot(x, y) - returns hypotenuse as Float
func (i *Interpreter) builtinHypot(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "Hypot() expects exactly 2 arguments, got %d", len(args))
	}

	x := args[0]
	y := args[1]

	var xFloat, yFloat float64

	// Handle x parameter
	if intVal, ok := x.(*IntegerValue); ok {
		xFloat = float64(intVal.Value)
	} else if floatVal, ok := x.(*FloatValue); ok {
		xFloat = floatVal.Value
	} else {
		return i.newErrorWithLocation(i.currentNode, "Hypot() expects Integer or Float as first argument, got %s", x.Type())
	}

	// Handle y parameter
	if intVal, ok := y.(*IntegerValue); ok {
		yFloat = float64(intVal.Value)
	} else if floatVal, ok := y.(*FloatValue); ok {
		yFloat = floatVal.Value
	} else {
		return i.newErrorWithLocation(i.currentNode, "Hypot() expects Integer or Float as second argument, got %s", y.Type())
	}

	return &FloatValue{Value: math.Hypot(xFloat, yFloat)}
}

// builtinSinh implements the Sinh() built-in function.
// It returns the hyperbolic sine of a number.
// Sinh(x) - returns hyperbolic sine as Float
func (i *Interpreter) builtinSinh(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Sinh() expects exactly 1 argument, got %d", len(args))
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
		return i.newErrorWithLocation(i.currentNode, "Sinh() expects Integer or Float as argument, got %s", arg.Type())
	}

	return &FloatValue{Value: math.Sinh(value)}
}

// builtinCosh implements the Cosh() built-in function.
// It returns the hyperbolic cosine of a number.
// Cosh(x) - returns hyperbolic cosine as Float
func (i *Interpreter) builtinCosh(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Cosh() expects exactly 1 argument, got %d", len(args))
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
		return i.newErrorWithLocation(i.currentNode, "Cosh() expects Integer or Float as argument, got %s", arg.Type())
	}

	return &FloatValue{Value: math.Cosh(value)}
}

// builtinTanh implements the Tanh() built-in function.
// It returns the hyperbolic tangent of a number.
// Tanh(x) - returns hyperbolic tangent as Float
func (i *Interpreter) builtinTanh(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Tanh() expects exactly 1 argument, got %d", len(args))
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
		return i.newErrorWithLocation(i.currentNode, "Tanh() expects Integer or Float as argument, got %s", arg.Type())
	}

	return &FloatValue{Value: math.Tanh(value)}
}

// builtinArcSinh implements the ArcSinh() built-in function.
// It returns the inverse hyperbolic sine of a number.
// ArcSinh(x) - returns inverse hyperbolic sine as Float
func (i *Interpreter) builtinArcSinh(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "ArcSinh() expects exactly 1 argument, got %d", len(args))
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
		return i.newErrorWithLocation(i.currentNode, "ArcSinh() expects Integer or Float as argument, got %s", arg.Type())
	}

	return &FloatValue{Value: math.Asinh(value)}
}

// builtinArcCosh implements the ArcCosh() built-in function.
// It returns the inverse hyperbolic cosine of a number.
// ArcCosh(x) - returns inverse hyperbolic cosine as Float (x must be >= 1)
func (i *Interpreter) builtinArcCosh(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "ArcCosh() expects exactly 1 argument, got %d", len(args))
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
		return i.newErrorWithLocation(i.currentNode, "ArcCosh() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Check domain: x must be >= 1
	if value < 1.0 {
		return i.newErrorWithLocation(i.currentNode, "ArcCosh() argument must be >= 1, got %f", value)
	}

	return &FloatValue{Value: math.Acosh(value)}
}

// builtinArcTanh implements the ArcTanh() built-in function.
// It returns the inverse hyperbolic tangent of a number.
// ArcTanh(x) - returns inverse hyperbolic tangent as Float (x must be in (-1, 1))
func (i *Interpreter) builtinArcTanh(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "ArcTanh() expects exactly 1 argument, got %d", len(args))
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
		return i.newErrorWithLocation(i.currentNode, "ArcTanh() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Check domain: x must be in (-1, 1)
	if value <= -1.0 || value >= 1.0 {
		return i.newErrorWithLocation(i.currentNode, "ArcTanh() argument must be in (-1, 1), got %f", value)
	}

	return &FloatValue{Value: math.Atanh(value)}
}
