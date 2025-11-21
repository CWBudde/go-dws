package interp

import (
	"math"

	"github.com/cwbudde/go-dws/pkg/ast"
)

// builtinRound implements the Round() built-in function.
// It rounds a number to the nearest integer.
// Round(x) - returns rounded value as Integer (always returns Integer)
// Task 9.4.5: Now supports Variant arguments.
func (i *Interpreter) builtinRound(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Round() expects exactly 1 argument, got %d", len(args))
	}

	// Task 9.4.5: Unwrap Variant if necessary
	arg := unwrapVariant(args[0])
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

	// Round to nearest integer using banker's rounding (round-half-to-even)
	// DWScript uses banker's rounding: 16.5 → 16, 17.5 → 18
	rounded := math.RoundToEven(value)
	return &IntegerValue{Value: int64(rounded)}
}

// builtinTrunc implements the Trunc() built-in function.
// It truncates a number towards zero (removes the decimal part).
// Trunc(x) - returns truncated value as Integer (always returns Integer)
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

// builtinClampInt implements the ClampInt() built-in function.
// It clamps an integer value to a range [min, max].
// ClampInt(value, min, max: Integer): Integer
func (i *Interpreter) builtinClampInt(args []Value) Value {
	if len(args) != 3 {
		return i.newErrorWithLocation(i.currentNode, "ClampInt() expects exactly 3 arguments, got %d", len(args))
	}

	// Extract value
	valueArg := args[0]
	valueInt, ok := valueArg.(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "ClampInt() expects Integer for argument 1, got %s", valueArg.Type())
	}
	value := valueInt.Value

	// Extract min
	minArg := args[1]
	minInt, ok := minArg.(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "ClampInt() expects Integer for argument 2, got %s", minArg.Type())
	}
	min := minInt.Value

	// Extract max
	maxArg := args[2]
	maxInt, ok := maxArg.(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "ClampInt() expects Integer for argument 3, got %s", maxArg.Type())
	}
	max := maxInt.Value

	// Clamp the value
	result := value
	if result < min {
		result = min
	} else if result > max {
		result = max
	}

	return &IntegerValue{Value: result}
}

// builtinClamp implements the Clamp() built-in function.
// It clamps a float value to a range [min, max].
// Clamp(value, min, max: Float): Float
// Accepts mixed Integer/Float arguments (converts to Float)
func (i *Interpreter) builtinClamp(args []Value) Value {
	if len(args) != 3 {
		return i.newErrorWithLocation(i.currentNode, "Clamp() expects exactly 3 arguments, got %d", len(args))
	}

	// Extract value (convert to float)
	var value float64
	switch v := args[0].(type) {
	case *IntegerValue:
		value = float64(v.Value)
	case *FloatValue:
		value = v.Value
	default:
		return i.newErrorWithLocation(i.currentNode, "Clamp() expects Integer or Float for argument 1, got %s", args[0].Type())
	}

	// Extract min (convert to float)
	var min float64
	switch v := args[1].(type) {
	case *IntegerValue:
		min = float64(v.Value)
	case *FloatValue:
		min = v.Value
	default:
		return i.newErrorWithLocation(i.currentNode, "Clamp() expects Integer or Float for argument 2, got %s", args[1].Type())
	}

	// Extract max (convert to float)
	var max float64
	switch v := args[2].(type) {
	case *IntegerValue:
		max = float64(v.Value)
	case *FloatValue:
		max = v.Value
	default:
		return i.newErrorWithLocation(i.currentNode, "Clamp() expects Integer or Float for argument 3, got %s", args[2].Type())
	}

	// Clamp the value
	result := value
	if result < min {
		result = min
	} else if result > max {
		result = max
	}

	return &FloatValue{Value: result}
}

// builtinFrac implements the Frac() built-in function.
// It returns the fractional part of a number.
// Frac(x: Float): Float
func (i *Interpreter) builtinFrac(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Frac() expects exactly 1 argument, got %d", len(args))
	}

	var floatVal float64
	switch v := args[0].(type) {
	case *FloatValue:
		floatVal = v.Value
	case *IntegerValue:
		floatVal = float64(v.Value)
	default:
		return i.newErrorWithLocation(i.currentNode, "Frac() expects Float or Integer, got %s", args[0].Type())
	}

	// Fractional part = x - floor(x)
	// For negative numbers: Frac(-2.3) = -2.3 - (-3) = 0.7
	_, frac := math.Modf(floatVal)
	return &FloatValue{Value: frac}
}

// builtinInt implements the Int() built-in function.
// It returns the integer part of a number as a Float (different from Trunc).
// Int(x: Float): Float
func (i *Interpreter) builtinInt(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Int() expects exactly 1 argument, got %d", len(args))
	}

	var floatVal float64
	switch v := args[0].(type) {
	case *FloatValue:
		floatVal = v.Value
	case *IntegerValue:
		floatVal = float64(v.Value)
	default:
		return i.newErrorWithLocation(i.currentNode, "Int() expects Float or Integer, got %s", args[0].Type())
	}

	// Int() returns the integer part (truncated towards zero) as a Float
	return &FloatValue{Value: math.Trunc(floatVal)}
}

// builtinIntPower implements the IntPower() built-in function.
// It computes base^exponent using integer exponentiation (faster than Power for integer exponents).
// IntPower(base: Float, exponent: Integer): Float
func (i *Interpreter) builtinIntPower(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "IntPower() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: base (Float or Integer)
	var base float64
	switch v := args[0].(type) {
	case *FloatValue:
		base = v.Value
	case *IntegerValue:
		base = float64(v.Value)
	default:
		return i.newErrorWithLocation(i.currentNode, "IntPower() expects Float or Integer as first argument, got %s", args[0].Type())
	}

	// Second argument: exponent (Integer)
	exponentVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "IntPower() expects Integer as second argument, got %s", args[1].Type())
	}
	exponent := exponentVal.Value

	// Handle special cases
	if exponent == 0 {
		return &FloatValue{Value: 1.0}
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

	return &FloatValue{Value: result}
}

// builtinDivMod implements the DivMod() built-in procedure.
// It computes both the quotient and remainder of integer division.
// DivMod(dividend, divisor: Integer; var quotient, remainder: Integer)
// Note: This function is called from functions_builtins.go with special handling for var parameters
func (i *Interpreter) builtinDivMod(args []ast.Expression) Value {
	// Validate argument count (exactly 4 arguments)
	if len(args) != 4 {
		return i.newErrorWithLocation(i.currentNode, "DivMod() expects exactly 4 arguments, got %d", len(args))
	}

	// Evaluate first two arguments (dividend and divisor)
	dividendVal := i.Eval(args[0])
	if isError(dividendVal) {
		return dividendVal
	}
	dividendInt, ok1 := dividendVal.(*IntegerValue)
	if !ok1 {
		return i.newErrorWithLocation(i.currentNode, "DivMod() expects integer as first argument, got %s", dividendVal.Type())
	}

	divisorVal := i.Eval(args[1])
	if isError(divisorVal) {
		return divisorVal
	}
	divisorInt, ok2 := divisorVal.(*IntegerValue)
	if !ok2 {
		return i.newErrorWithLocation(i.currentNode, "DivMod() expects integer as second argument, got %s", divisorVal.Type())
	}

	// Check for division by zero
	if divisorInt.Value == 0 {
		return i.newErrorWithLocation(i.currentNode, "DivMod() division by zero")
	}

	// Calculate quotient and remainder
	quotient := dividendInt.Value / divisorInt.Value
	remainder := dividendInt.Value % divisorInt.Value

	// Last two arguments must be identifiers (variable names for var parameters)
	quotientIdent, ok3 := args[2].(*ast.Identifier)
	if !ok3 {
		return i.newErrorWithLocation(i.currentNode, "DivMod() third argument must be a variable, got %T", args[2])
	}
	remainderIdent, ok4 := args[3].(*ast.Identifier)
	if !ok4 {
		return i.newErrorWithLocation(i.currentNode, "DivMod() fourth argument must be a variable, got %T", args[3])
	}

	// Get variable names
	quotientVarName := quotientIdent.Value
	remainderVarName := remainderIdent.Value

	// Check if variables exist and handle ReferenceValue (var parameters)
	quotientVar, exists1 := i.env.Get(quotientVarName)
	if !exists1 {
		return i.newErrorWithLocation(i.currentNode, "undefined variable: %s", quotientVarName)
	}
	remainderVar, exists2 := i.env.Get(remainderVarName)
	if !exists2 {
		return i.newErrorWithLocation(i.currentNode, "undefined variable: %s", remainderVarName)
	}

	// Handle var parameters (ReferenceValue)
	quotientResult := &IntegerValue{Value: quotient}
	remainderResult := &IntegerValue{Value: remainder}

	if refQuot, isRef := quotientVar.(*ReferenceValue); isRef {
		if err := refQuot.Assign(quotientResult); err != nil {
			return i.newErrorWithLocation(i.currentNode, "failed to update variable %s: %s", quotientVarName, err)
		}
	} else {
		if err := i.env.Set(quotientVarName, quotientResult); err != nil {
			return i.newErrorWithLocation(i.currentNode, "failed to update variable %s: %s", quotientVarName, err)
		}
	}

	if refRem, isRef := remainderVar.(*ReferenceValue); isRef {
		if err := refRem.Assign(remainderResult); err != nil {
			return i.newErrorWithLocation(i.currentNode, "failed to update variable %s: %s", remainderVarName, err)
		}
	} else {
		if err := i.env.Set(remainderVarName, remainderResult); err != nil {
			return i.newErrorWithLocation(i.currentNode, "failed to update variable %s: %s", remainderVarName, err)
		}
	}

	return &NilValue{}
}
