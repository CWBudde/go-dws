package bytecode

import (
	"math"
	"math/bits"
	"math/rand"
	"time"
)

// registerMathBuiltins registers all math functions
func (vm *VM) registerMathBuiltins() {
	vm.builtins["pi"] = builtinPi
	vm.builtins["sign"] = builtinSign
	vm.builtins["odd"] = builtinOdd
	vm.builtins["frac"] = builtinFrac
	vm.builtins["int"] = builtinInt
	vm.builtins["log10"] = builtinLog10
	vm.builtins["logn"] = builtinLogN
	vm.builtins["infinity"] = builtinInfinity
	vm.builtins["nan"] = builtinNaN
	vm.builtins["isfinite"] = builtinIsFinite
	vm.builtins["isinfinite"] = builtinIsInfinite
	vm.builtins["intpower"] = builtinIntPower
	vm.builtins["random"] = builtinRandom
	vm.builtins["randomint"] = builtinRandomInt
	vm.builtins["randseed"] = builtinRandSeed
	vm.builtins["randg"] = builtinRandG
	vm.builtins["setrandseed"] = builtinSetRandSeed
	vm.builtins["randomize"] = builtinRandomize
	vm.builtins["factorial"] = builtinFactorial
	vm.builtins["gcd"] = builtinGcd
	vm.builtins["lcm"] = builtinLcm
	vm.builtins["isprime"] = builtinIsPrime
	vm.builtins["leastfactor"] = builtinLeastFactor
	vm.builtins["popcount"] = builtinPopCount
	vm.builtins["testbit"] = builtinTestBit
	vm.builtins["haversine"] = builtinHaversine
	vm.builtins["comparenum"] = builtinCompareNum
}

// Math Functions

func builtinPi(vm *VM, args []Value) (Value, error) {
	if len(args) != 0 {
		return NilValue(), vm.runtimeError("Pi expects no arguments, got %d", len(args))
	}
	return FloatValue(math.Pi), nil
}

func builtinSign(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("Sign expects 1 argument, got %d", len(args))
	}
	arg := args[0]

	var floatVal float64
	if arg.IsFloat() {
		floatVal = arg.AsFloat()
	} else if arg.IsInt() {
		floatVal = float64(arg.AsInt())
	} else {
		return NilValue(), vm.runtimeError("Sign expects Float or Integer, got %s", arg.Type.String())
	}

	if floatVal > 0 {
		return IntValue(1), nil
	} else if floatVal < 0 {
		return IntValue(-1), nil
	}
	return IntValue(0), nil
}

func builtinOdd(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("Odd expects 1 argument, got %d", len(args))
	}
	arg := args[0]

	if !arg.IsInt() {
		return NilValue(), vm.runtimeError("Odd expects Integer, got %s", arg.Type.String())
	}

	return BoolValue(arg.AsInt()%2 != 0), nil
}

func builtinFrac(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("Frac expects 1 argument, got %d", len(args))
	}
	arg := args[0]

	var floatVal float64
	if arg.IsFloat() {
		floatVal = arg.AsFloat()
	} else if arg.IsInt() {
		floatVal = float64(arg.AsInt())
	} else {
		return NilValue(), vm.runtimeError("Frac expects Float or Integer, got %s", arg.Type.String())
	}

	// Fractional part = x - floor(x)
	_, frac := math.Modf(floatVal)
	return FloatValue(frac), nil
}

func builtinInt(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("Int expects 1 argument, got %d", len(args))
	}
	arg := args[0]

	var floatVal float64
	if arg.IsFloat() {
		floatVal = arg.AsFloat()
	} else if arg.IsInt() {
		floatVal = float64(arg.AsInt())
	} else {
		return NilValue(), vm.runtimeError("Int expects Float or Integer, got %s", arg.Type.String())
	}

	// Int() returns the integer part (truncated towards zero) as a Float
	return FloatValue(math.Trunc(floatVal)), nil
}

func builtinLog10(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("Log10 expects 1 argument, got %d", len(args))
	}
	arg := args[0]

	var floatVal float64
	if arg.IsFloat() {
		floatVal = arg.AsFloat()
	} else if arg.IsInt() {
		floatVal = float64(arg.AsInt())
	} else {
		return NilValue(), vm.runtimeError("Log10 expects Float or Integer, got %s", arg.Type.String())
	}

	if floatVal <= 0 {
		return NilValue(), vm.runtimeError("Log10 argument must be positive, got %f", floatVal)
	}

	return FloatValue(math.Log10(floatVal)), nil
}

func builtinLogN(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("LogN expects 2 arguments, got %d", len(args))
	}

	// First argument (x)
	var xVal float64
	if args[0].IsFloat() {
		xVal = args[0].AsFloat()
	} else if args[0].IsInt() {
		xVal = float64(args[0].AsInt())
	} else {
		return NilValue(), vm.runtimeError("LogN expects Float or Integer as first argument, got %s", args[0].Type.String())
	}

	// Second argument (base)
	var baseVal float64
	if args[1].IsFloat() {
		baseVal = args[1].AsFloat()
	} else if args[1].IsInt() {
		baseVal = float64(args[1].AsInt())
	} else {
		return NilValue(), vm.runtimeError("LogN expects Float or Integer as second argument, got %s", args[1].Type.String())
	}

	if xVal <= 0 {
		return NilValue(), vm.runtimeError("LogN first argument must be positive, got %f", xVal)
	}
	if baseVal <= 0 || baseVal == 1 {
		return NilValue(), vm.runtimeError("LogN base must be positive and not equal to 1, got %f", baseVal)
	}

	// LogN(x, base) = Log(x) / Log(base)
	return FloatValue(math.Log(xVal) / math.Log(baseVal)), nil
}

// MEDIUM PRIORITY Math Functions

func builtinInfinity(vm *VM, args []Value) (Value, error) {
	if len(args) != 0 {
		return NilValue(), vm.runtimeError("Infinity expects no arguments, got %d", len(args))
	}
	return FloatValue(math.Inf(1)), nil
}

func builtinNaN(vm *VM, args []Value) (Value, error) {
	if len(args) != 0 {
		return NilValue(), vm.runtimeError("NaN expects no arguments, got %d", len(args))
	}
	return FloatValue(math.NaN()), nil
}

func builtinIsFinite(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("IsFinite expects 1 argument, got %d", len(args))
	}

	var floatVal float64
	if args[0].IsFloat() {
		floatVal = args[0].AsFloat()
	} else if args[0].IsInt() {
		floatVal = float64(args[0].AsInt())
	} else {
		return NilValue(), vm.runtimeError("IsFinite expects Float or Integer, got %s", args[0].Type.String())
	}

	return BoolValue(!math.IsInf(floatVal, 0) && !math.IsNaN(floatVal)), nil
}

func builtinIsInfinite(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("IsInfinite expects 1 argument, got %d", len(args))
	}

	var floatVal float64
	if args[0].IsFloat() {
		floatVal = args[0].AsFloat()
	} else if args[0].IsInt() {
		floatVal = float64(args[0].AsInt())
	} else {
		return NilValue(), vm.runtimeError("IsInfinite expects Float or Integer, got %s", args[0].Type.String())
	}

	return BoolValue(math.IsInf(floatVal, 0)), nil
}

func builtinIntPower(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("IntPower expects 2 arguments, got %d", len(args))
	}

	// First argument (base) - Float or Integer
	var baseVal float64
	if args[0].IsFloat() {
		baseVal = args[0].AsFloat()
	} else if args[0].IsInt() {
		baseVal = float64(args[0].AsInt())
	} else {
		return NilValue(), vm.runtimeError("IntPower expects Float or Integer as first argument, got %s", args[0].Type.String())
	}

	// Second argument (exponent) - Integer only
	if !args[1].IsInt() {
		return NilValue(), vm.runtimeError("IntPower expects Integer as second argument, got %s", args[1].Type.String())
	}
	expVal := args[1].AsInt()

	// Calculate power using exponentiation by squaring for integer exponents
	result := 1.0
	base := baseVal
	exp := expVal

	if exp < 0 {
		base = 1.0 / base
		exp = -exp
	}

	for exp > 0 {
		if exp%2 == 1 {
			result *= base
		}
		base *= base
		exp /= 2
	}

	return FloatValue(result), nil
}

func builtinRandom(vm *VM, args []Value) (Value, error) {
	if len(args) != 0 {
		return NilValue(), vm.runtimeError("Random expects no arguments, got %d", len(args))
	}
	return FloatValue(vm.rand.Float64()), nil
}

func builtinRandomInt(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("RandomInt expects 1 argument, got %d", len(args))
	}

	if !args[0].IsInt() {
		return NilValue(), vm.runtimeError("RandomInt expects Integer argument, got %s", args[0].Type.String())
	}

	max := args[0].AsInt()
	if max <= 0 {
		return NilValue(), vm.runtimeError("RandomInt expects max > 0, got %d", max)
	}

	return IntValue(int64(vm.rand.Intn(int(max)))), nil
}

func builtinRandSeed(vm *VM, args []Value) (Value, error) {
	if len(args) != 0 {
		return NilValue(), vm.runtimeError("RandSeed expects no arguments, got %d", len(args))
	}
	return IntValue(vm.randSeed), nil
}

func builtinRandG(vm *VM, args []Value) (Value, error) {
	if len(args) != 0 {
		return NilValue(), vm.runtimeError("RandG expects no arguments, got %d", len(args))
	}

	// Generate Gaussian random number using Box-Muller transform
	u1 := vm.rand.Float64()
	u2 := vm.rand.Float64()

	// Ensure u1 is not zero or near-zero to avoid log(0)
	if u1 < 1e-10 {
		u1 = 1e-10
	}

	// Box-Muller transform
	z0 := math.Sqrt(-2.0*math.Log(u1)) * math.Cos(2.0*math.Pi*u2)

	return FloatValue(z0), nil
}

func builtinSetRandSeed(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("SetRandSeed expects 1 argument, got %d", len(args))
	}

	if !args[0].IsInt() {
		return NilValue(), vm.runtimeError("SetRandSeed expects Integer, got %s", args[0].Type.String())
	}

	seed := args[0].AsInt()
	vm.randSeed = seed
	vm.rand = rand.New(rand.NewSource(seed))

	return NilValue(), nil
}

func builtinRandomize(vm *VM, args []Value) (Value, error) {
	if len(args) != 0 {
		return NilValue(), vm.runtimeError("Randomize expects no arguments, got %d", len(args))
	}

	seed := time.Now().UnixNano()
	vm.randSeed = seed
	vm.rand = rand.New(rand.NewSource(seed))

	return NilValue(), nil
}

// ============================================================================
// Advanced Math Functions (Phase 9.23)
// ============================================================================

// builtinFactorial implements the Factorial() built-in function.
// Factorial(n: Integer): Integer
func builtinFactorial(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("Factorial expects 1 argument, got %d", len(args))
	}
	if !args[0].IsInt() {
		return NilValue(), vm.runtimeError("Factorial expects Integer argument")
	}

	n := args[0].AsInt()
	if n < 0 {
		return NilValue(), vm.runtimeError("Factorial expects non-negative integer, got %d", n)
	}

	// Handle overflow: 20! is the largest factorial that fits in int64
	if n > 20 {
		return NilValue(), vm.runtimeError("Factorial overflow: %d! is too large for Integer", n)
	}

	result := int64(1)
	for i := int64(2); i <= n; i++ {
		result *= i
	}

	return IntValue(result), nil
}

// builtinGcd implements the Gcd() built-in function.
// Gcd(a, b: Integer): Integer
func builtinGcd(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("Gcd expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsInt() || !args[1].IsInt() {
		return NilValue(), vm.runtimeError("Gcd expects Integer arguments")
	}

	a := args[0].AsInt()
	b := args[1].AsInt()

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

	return IntValue(a), nil
}

// builtinLcm implements the Lcm() built-in function.
// Lcm(a, b: Integer): Integer
func builtinLcm(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("Lcm expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsInt() || !args[1].IsInt() {
		return NilValue(), vm.runtimeError("Lcm expects Integer arguments")
	}

	a := args[0].AsInt()
	b := args[1].AsInt()

	// Handle special case: if either is 0, result is 0
	if a == 0 || b == 0 {
		return IntValue(0), nil
	}

	// Take absolute values
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}

	// Compute GCD
	gcdA, gcdB := a, b
	for gcdB != 0 {
		gcdA, gcdB = gcdB, gcdA%gcdB
	}

	// LCM = |a * b| / gcd(a, b)
	result := (a / gcdA) * b

	return IntValue(result), nil
}

// builtinIsPrime implements the IsPrime() built-in function.
// IsPrime(n: Integer): Boolean
func builtinIsPrime(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("IsPrime expects 1 argument, got %d", len(args))
	}
	if !args[0].IsInt() {
		return NilValue(), vm.runtimeError("IsPrime expects Integer argument")
	}

	n := args[0].AsInt()

	// Numbers less than 2 are not prime
	if n < 2 {
		return BoolValue(false), nil
	}

	// 2 and 3 are prime
	if n == 2 || n == 3 {
		return BoolValue(true), nil
	}

	// Even numbers (except 2) are not prime
	if n%2 == 0 {
		return BoolValue(false), nil
	}

	// Multiples of 3 (except 3) are not prime
	if n%3 == 0 {
		return BoolValue(false), nil
	}

	// Check divisibility by numbers of form 6k ± 1 up to √n
	sqrtN := int64(math.Sqrt(float64(n)))
	for i := int64(5); i <= sqrtN; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return BoolValue(false), nil
		}
	}

	return BoolValue(true), nil
}

// builtinLeastFactor implements the LeastFactor() built-in function.
// LeastFactor(n: Integer): Integer
func builtinLeastFactor(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("LeastFactor expects 1 argument, got %d", len(args))
	}
	if !args[0].IsInt() {
		return NilValue(), vm.runtimeError("LeastFactor expects Integer argument")
	}

	n := args[0].AsInt()

	// Handle special cases
	if n <= 1 {
		return IntValue(1), nil
	}

	// Check for divisibility by 2
	if n%2 == 0 {
		return IntValue(2), nil
	}

	// Check for divisibility by 3
	if n%3 == 0 {
		return IntValue(3), nil
	}

	// Check divisibility by numbers of form 6k ± 1 up to √n
	sqrtN := int64(math.Sqrt(float64(n)))
	for i := int64(5); i <= sqrtN; i += 6 {
		if n%i == 0 {
			return IntValue(i), nil
		}
		if n%(i+2) == 0 {
			return IntValue(i + 2), nil
		}
	}

	// If no factor found, n is prime
	return IntValue(n), nil
}

// builtinPopCount implements the PopCount() built-in function.
// PopCount(n: Integer): Integer
func builtinPopCount(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("PopCount expects 1 argument, got %d", len(args))
	}
	if !args[0].IsInt() {
		return NilValue(), vm.runtimeError("PopCount expects Integer argument")
	}

	// Use math/bits.OnesCount64 for counting set bits
	count := int64(bits.OnesCount64(uint64(args[0].AsInt())))

	return IntValue(count), nil
}

// builtinTestBit implements the TestBit() built-in function.
// TestBit(value: Integer, bit: Integer): Boolean
func builtinTestBit(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("TestBit expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsInt() || !args[1].IsInt() {
		return NilValue(), vm.runtimeError("TestBit expects Integer arguments")
	}

	value := args[0].AsInt()
	bit := args[1].AsInt()

	// Validate bit position (0-63 for int64)
	if bit < 0 || bit >= 64 {
		return NilValue(), vm.runtimeError("TestBit bit position must be in range 0-63, got %d", bit)
	}

	// Test the bit: (value >> bit) & 1
	isSet := (value >> uint(bit)) & 1
	return BoolValue(isSet != 0), nil
}

// builtinHaversine implements the Haversine() built-in function.
// Haversine(lat1, lon1, lat2, lon2: Float): Float
func builtinHaversine(vm *VM, args []Value) (Value, error) {
	if len(args) != 4 {
		return NilValue(), vm.runtimeError("Haversine expects 4 arguments, got %d", len(args))
	}

	// Extract and convert all arguments to float64
	var lat1, lon1, lat2, lon2 float64

	if args[0].IsFloat() {
		lat1 = args[0].AsFloat()
	} else if args[0].IsInt() {
		lat1 = float64(args[0].AsInt())
	} else {
		return NilValue(), vm.runtimeError("Haversine expects numeric arguments")
	}

	if args[1].IsFloat() {
		lon1 = args[1].AsFloat()
	} else if args[1].IsInt() {
		lon1 = float64(args[1].AsInt())
	} else {
		return NilValue(), vm.runtimeError("Haversine expects numeric arguments")
	}

	if args[2].IsFloat() {
		lat2 = args[2].AsFloat()
	} else if args[2].IsInt() {
		lat2 = float64(args[2].AsInt())
	} else {
		return NilValue(), vm.runtimeError("Haversine expects numeric arguments")
	}

	if args[3].IsFloat() {
		lon2 = args[3].AsFloat()
	} else if args[3].IsInt() {
		lon2 = float64(args[3].AsInt())
	} else {
		return NilValue(), vm.runtimeError("Haversine expects numeric arguments")
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

	return FloatValue(distance), nil
}

// builtinCompareNum implements the CompareNum() built-in function.
// CompareNum(a, b: Float): Integer
func builtinCompareNum(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("CompareNum expects 2 arguments, got %d", len(args))
	}

	// Extract first argument
	var a float64
	if args[0].IsFloat() {
		a = args[0].AsFloat()
	} else if args[0].IsInt() {
		a = float64(args[0].AsInt())
	} else {
		return NilValue(), vm.runtimeError("CompareNum expects numeric arguments")
	}

	// Extract second argument
	var b float64
	if args[1].IsFloat() {
		b = args[1].AsFloat()
	} else if args[1].IsInt() {
		b = float64(args[1].AsInt())
	} else {
		return NilValue(), vm.runtimeError("CompareNum expects numeric arguments")
	}

	// Handle NaN: NaN is considered equal to NaN, and less than all other values
	aIsNaN := math.IsNaN(a)
	bIsNaN := math.IsNaN(b)

	if aIsNaN && bIsNaN {
		return IntValue(0), nil // Both NaN, equal
	}
	if aIsNaN {
		return IntValue(-1), nil // NaN is less than any number
	}
	if bIsNaN {
		return IntValue(1), nil // Any number is greater than NaN
	}

	// Regular comparison
	if a < b {
		return IntValue(-1), nil
	} else if a > b {
		return IntValue(1), nil
	}
	return IntValue(0), nil
}
