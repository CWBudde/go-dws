package interp

import (
	"math"
	"math/bits"
)

// builtinFactorial implements the Factorial() built-in function.
// It computes the factorial of a non-negative integer: n! = n * (n-1) * ... * 1
// Factorial(n: Integer): Integer
func (i *Interpreter) builtinFactorial(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Factorial() expects exactly 1 argument, got %d", len(args))
	}

	intVal, ok := args[0].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Factorial() expects Integer, got %s", args[0].Type())
	}

	n := intVal.Value

	// Validate n >= 0
	if n < 0 {
		return i.newErrorWithLocation(i.currentNode, "Factorial() expects non-negative integer, got %d", n)
	}

	// Handle overflow: 20! is the largest factorial that fits in int64
	// 21! = 51090942171709440000 which overflows int64
	if n > 20 {
		return i.newErrorWithLocation(i.currentNode, "Factorial() overflow: %d! is too large for Integer", n)
	}

	// Calculate factorial
	result := int64(1)
	for j := int64(2); j <= n; j++ {
		result *= j
	}

	return &IntegerValue{Value: result}
}

// builtinGcd implements the Gcd() built-in function.
// It computes the Greatest Common Divisor using Euclidean algorithm.
// Gcd(a, b: Integer): Integer
func (i *Interpreter) builtinGcd(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "Gcd() expects exactly 2 arguments, got %d", len(args))
	}

	aVal, ok1 := args[0].(*IntegerValue)
	bVal, ok2 := args[1].(*IntegerValue)

	if !ok1 || !ok2 {
		return i.newErrorWithLocation(i.currentNode, "Gcd() expects Integer arguments, got %s and %s", args[0].Type(), args[1].Type())
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

	return &IntegerValue{Value: a}
}

// builtinLcm implements the Lcm() built-in function.
// It computes the Least Common Multiple: lcm(a, b) = |a * b| / gcd(a, b)
// Lcm(a, b: Integer): Integer
func (i *Interpreter) builtinLcm(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "Lcm() expects exactly 2 arguments, got %d", len(args))
	}

	aVal, ok1 := args[0].(*IntegerValue)
	bVal, ok2 := args[1].(*IntegerValue)

	if !ok1 || !ok2 {
		return i.newErrorWithLocation(i.currentNode, "Lcm() expects Integer arguments, got %s and %s", args[0].Type(), args[1].Type())
	}

	a := aVal.Value
	b := bVal.Value

	// Handle special case: if either is 0, result is 0
	if a == 0 || b == 0 {
		return &IntegerValue{Value: 0}
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

	return &IntegerValue{Value: result}
}

// builtinIsPrime implements the IsPrime() built-in function.
// It checks if a number is prime using trial division.
// IsPrime(n: Integer): Boolean
func (i *Interpreter) builtinIsPrime(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "IsPrime() expects exactly 1 argument, got %d", len(args))
	}

	intVal, ok := args[0].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "IsPrime() expects Integer, got %s", args[0].Type())
	}

	n := intVal.Value

	// Numbers less than 2 are not prime
	if n < 2 {
		return &BooleanValue{Value: false}
	}

	// 2 and 3 are prime
	if n == 2 || n == 3 {
		return &BooleanValue{Value: true}
	}

	// Even numbers (except 2) are not prime
	if n%2 == 0 {
		return &BooleanValue{Value: false}
	}

	// Multiples of 3 (except 3) are not prime
	if n%3 == 0 {
		return &BooleanValue{Value: false}
	}

	// Check divisibility by numbers of form 6k ± 1 up to √n
	// This is an optimization: all primes > 3 are of form 6k ± 1
	sqrtN := int64(math.Sqrt(float64(n)))
	for i := int64(5); i <= sqrtN; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return &BooleanValue{Value: false}
		}
	}

	return &BooleanValue{Value: true}
}

// builtinLeastFactor implements the LeastFactor() built-in function.
// It finds the smallest prime factor of n.
// LeastFactor(n: Integer): Integer
func (i *Interpreter) builtinLeastFactor(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "LeastFactor() expects exactly 1 argument, got %d", len(args))
	}

	intVal, ok := args[0].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "LeastFactor() expects Integer, got %s", args[0].Type())
	}

	n := intVal.Value

	// Handle special cases
	if n <= 1 {
		return &IntegerValue{Value: 1}
	}

	// Check for divisibility by 2
	if n%2 == 0 {
		return &IntegerValue{Value: 2}
	}

	// Check for divisibility by 3
	if n%3 == 0 {
		return &IntegerValue{Value: 3}
	}

	// Check divisibility by numbers of form 6k ± 1 up to √n
	sqrtN := int64(math.Sqrt(float64(n)))
	for i := int64(5); i <= sqrtN; i += 6 {
		if n%i == 0 {
			return &IntegerValue{Value: i}
		}
		if n%(i+2) == 0 {
			return &IntegerValue{Value: i + 2}
		}
	}

	// If no factor found, n is prime
	return &IntegerValue{Value: n}
}

// builtinPopCount implements the PopCount() built-in function.
// It counts the number of set bits (1s) in the binary representation.
// PopCount(n: Integer): Integer
func (i *Interpreter) builtinPopCount(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "PopCount() expects exactly 1 argument, got %d", len(args))
	}

	intVal, ok := args[0].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "PopCount() expects Integer, got %s", args[0].Type())
	}

	// Use Go's built-in bits.OnesCount64 for unsigned 64-bit integers
	// Convert to uint64 to count bits
	count := bits.OnesCount64(uint64(intVal.Value))

	return &IntegerValue{Value: int64(count)}
}

// builtinTestBit implements the TestBit() built-in function.
// It tests if a specific bit is set in a number.
// TestBit(value: Integer, bit: Integer): Boolean
func (i *Interpreter) builtinTestBit(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "TestBit() expects exactly 2 arguments, got %d", len(args))
	}

	valueVal, ok1 := args[0].(*IntegerValue)
	bitVal, ok2 := args[1].(*IntegerValue)

	if !ok1 || !ok2 {
		return i.newErrorWithLocation(i.currentNode, "TestBit() expects Integer arguments, got %s and %s", args[0].Type(), args[1].Type())
	}

	value := valueVal.Value
	bit := bitVal.Value

	// Validate bit position (0-63 for int64)
	if bit < 0 || bit >= 64 {
		return i.newErrorWithLocation(i.currentNode, "TestBit() bit position must be in range 0-63, got %d", bit)
	}

	// Test the bit: (value >> bit) & 1
	isSet := (value >> uint(bit)) & 1
	return &BooleanValue{Value: isSet != 0}
}

// builtinHaversine implements the Haversine() built-in function.
// It calculates the great-circle distance between two points on a sphere
// given their latitudes and longitudes in degrees.
// Result is in kilometers (Earth radius = 6371 km).
// Haversine(lat1, lon1, lat2, lon2: Float): Float
func (i *Interpreter) builtinHaversine(args []Value) Value {
	if len(args) != 4 {
		return i.newErrorWithLocation(i.currentNode, "Haversine() expects exactly 4 arguments, got %d", len(args))
	}

	// Extract and convert all arguments to float64
	var lat1, lon1, lat2, lon2 float64

	// First argument (lat1)
	switch v := args[0].(type) {
	case *FloatValue:
		lat1 = v.Value
	case *IntegerValue:
		lat1 = float64(v.Value)
	default:
		return i.newErrorWithLocation(i.currentNode, "Haversine() expects Float or Integer arguments, got %s", args[0].Type())
	}

	// Second argument (lon1)
	switch v := args[1].(type) {
	case *FloatValue:
		lon1 = v.Value
	case *IntegerValue:
		lon1 = float64(v.Value)
	default:
		return i.newErrorWithLocation(i.currentNode, "Haversine() expects Float or Integer arguments, got %s", args[1].Type())
	}

	// Third argument (lat2)
	switch v := args[2].(type) {
	case *FloatValue:
		lat2 = v.Value
	case *IntegerValue:
		lat2 = float64(v.Value)
	default:
		return i.newErrorWithLocation(i.currentNode, "Haversine() expects Float or Integer arguments, got %s", args[2].Type())
	}

	// Fourth argument (lon2)
	switch v := args[3].(type) {
	case *FloatValue:
		lon2 = v.Value
	case *IntegerValue:
		lon2 = float64(v.Value)
	default:
		return i.newErrorWithLocation(i.currentNode, "Haversine() expects Float or Integer arguments, got %s", args[3].Type())
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

	return &FloatValue{Value: distance}
}

// builtinCompareNum implements the CompareNum() built-in function.
// It compares two numbers and returns -1, 0, or 1.
// CompareNum(a, b: Float): Integer
func (i *Interpreter) builtinCompareNum(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "CompareNum() expects exactly 2 arguments, got %d", len(args))
	}

	// Extract first argument
	var a float64
	switch v := args[0].(type) {
	case *FloatValue:
		a = v.Value
	case *IntegerValue:
		a = float64(v.Value)
	default:
		return i.newErrorWithLocation(i.currentNode, "CompareNum() expects Float or Integer arguments, got %s", args[0].Type())
	}

	// Extract second argument
	var b float64
	switch v := args[1].(type) {
	case *FloatValue:
		b = v.Value
	case *IntegerValue:
		b = float64(v.Value)
	default:
		return i.newErrorWithLocation(i.currentNode, "CompareNum() expects Float or Integer arguments, got %s", args[1].Type())
	}

	// Handle NaN: NaN is considered equal to NaN, and less than all other values
	aIsNaN := math.IsNaN(a)
	bIsNaN := math.IsNaN(b)

	if aIsNaN && bIsNaN {
		return &IntegerValue{Value: 0} // Both NaN, equal
	}
	if aIsNaN {
		return &IntegerValue{Value: -1} // NaN is less than any number
	}
	if bIsNaN {
		return &IntegerValue{Value: 1} // Any number is greater than NaN
	}

	// Regular comparison
	if a < b {
		return &IntegerValue{Value: -1}
	} else if a > b {
		return &IntegerValue{Value: 1}
	}
	return &IntegerValue{Value: 0}
}
