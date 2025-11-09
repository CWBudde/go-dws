package semantic

import (
	"testing"
)

// ============================================================================
// Built-in Math Functions Tests
// ============================================================================
// These tests cover the built-in mathematical functions to improve
// coverage of analyze_builtin_math.go (currently at 0-62% coverage)

// Abs function tests
func TestBuiltinAbs_Integer(t *testing.T) {
	input := `
		var x := Abs(-42);
		var y := Abs(42);
	`
	expectNoErrors(t, input)
}

func TestBuiltinAbs_Float(t *testing.T) {
	input := `
		var x := Abs(-3.14);
		var y := Abs(3.14);
	`
	expectNoErrors(t, input)
}

func TestBuiltinAbs_InvalidType(t *testing.T) {
	input := `
		var x := Abs('hello');
	`
	expectError(t, input, "numeric")
}

// Min/Max function tests
func TestBuiltinMin_TwoIntegers(t *testing.T) {
	input := `
		var result := Min(5, 10);
	`
	expectNoErrors(t, input)
}

func TestBuiltinMin_TwoFloats(t *testing.T) {
	input := `
		var result := Min(3.14, 2.71);
	`
	expectNoErrors(t, input)
}

func TestBuiltinMax_TwoIntegers(t *testing.T) {
	input := `
		var result := Max(5, 10);
	`
	expectNoErrors(t, input)
}

func TestBuiltinMax_TwoFloats(t *testing.T) {
	input := `
		var result := Max(3.14, 2.71);
	`
	expectNoErrors(t, input)
}

func TestBuiltinMinMax_InvalidArgCount(t *testing.T) {
	input := `
		var result := Min(5);
	`
	expectError(t, input, "argument")
}

// Clamp function tests
func TestBuiltinClampInt_Basic(t *testing.T) {
	input := `
		var result := ClampInt(50, 0, 100);
	`
	expectNoErrors(t, input)
}

func TestBuiltinClampInt_BelowMin(t *testing.T) {
	input := `
		var result := ClampInt(-10, 0, 100);
	`
	expectNoErrors(t, input)
}

func TestBuiltinClampInt_AboveMax(t *testing.T) {
	input := `
		var result := ClampInt(150, 0, 100);
	`
	expectNoErrors(t, input)
}

func TestBuiltinClamp_Float(t *testing.T) {
	input := `
		var result := Clamp(3.14, 0.0, 5.0);
	`
	expectNoErrors(t, input)
}

// Sqr and Sqrt function tests
func TestBuiltinSqr_Integer(t *testing.T) {
	input := `
		var result := Sqr(5);
	`
	expectNoErrors(t, input)
}

func TestBuiltinSqr_Float(t *testing.T) {
	input := `
		var result := Sqr(3.5);
	`
	expectNoErrors(t, input)
}

func TestBuiltinSqrt_Basic(t *testing.T) {
	input := `
		var result := Sqrt(25.0);
	`
	expectNoErrors(t, input)
}

func TestBuiltinSqrt_Negative(t *testing.T) {
	// Should analyze without error (runtime error check)
	input := `
		var result := Sqrt(-1.0);
	`
	expectNoErrors(t, input)
}

// Power function tests
func TestBuiltinPower_Basic(t *testing.T) {
	input := `
		var result := Power(2.0, 8.0);
	`
	expectNoErrors(t, input)
}

func TestBuiltinPower_Negative(t *testing.T) {
	input := `
		var result := Power(2.0, -2.0);
	`
	expectNoErrors(t, input)
}

// Trigonometric function tests
func TestBuiltinSin_Basic(t *testing.T) {
	input := `
		var result := Sin(0.0);
	`
	expectNoErrors(t, input)
}

func TestBuiltinCos_Basic(t *testing.T) {
	input := `
		var result := Cos(0.0);
	`
	expectNoErrors(t, input)
}

func TestBuiltinTan_Basic(t *testing.T) {
	input := `
		var result := Tan(0.0);
	`
	expectNoErrors(t, input)
}

func TestBuiltinArcSin_Basic(t *testing.T) {
	input := `
		var result := ArcSin(0.5);
	`
	expectNoErrors(t, input)
}

func TestBuiltinArcCos_Basic(t *testing.T) {
	input := `
		var result := ArcCos(0.5);
	`
	expectNoErrors(t, input)
}

func TestBuiltinArcTan_Basic(t *testing.T) {
	input := `
		var result := ArcTan(1.0);
	`
	expectNoErrors(t, input)
}

func TestBuiltinArcTan2_Basic(t *testing.T) {
	input := `
		var result := ArcTan2(1.0, 1.0);
	`
	expectNoErrors(t, input)
}

// Angle conversion tests
func TestBuiltinDegToRad_Basic(t *testing.T) {
	input := `
		var result := DegToRad(180.0);
	`
	expectNoErrors(t, input)
}

func TestBuiltinRadToDeg_Basic(t *testing.T) {
	input := `
		var result := RadToDeg(3.14159);
	`
	expectNoErrors(t, input)
}

// Hyperbolic function tests
func TestBuiltinSinh_Basic(t *testing.T) {
	input := `
		var result := Sinh(1.0);
	`
	expectNoErrors(t, input)
}

func TestBuiltinCosh_Basic(t *testing.T) {
	input := `
		var result := Cosh(1.0);
	`
	expectNoErrors(t, input)
}

func TestBuiltinTanh_Basic(t *testing.T) {
	input := `
		var result := Tanh(1.0);
	`
	expectNoErrors(t, input)
}

func TestBuiltinArcSinh_Basic(t *testing.T) {
	input := `
		var result := ArcSinh(1.0);
	`
	expectNoErrors(t, input)
}

func TestBuiltinArcCosh_Basic(t *testing.T) {
	input := `
		var result := ArcCosh(2.0);
	`
	expectNoErrors(t, input)
}

func TestBuiltinArcTanh_Basic(t *testing.T) {
	input := `
		var result := ArcTanh(0.5);
	`
	expectNoErrors(t, input)
}

// Logarithm and exponential tests
func TestBuiltinExp_Basic(t *testing.T) {
	input := `
		var result := Exp(1.0);
	`
	expectNoErrors(t, input)
}

func TestBuiltinLn_Basic(t *testing.T) {
	input := `
		var result := Ln(2.718281828);
	`
	expectNoErrors(t, input)
}

func TestBuiltinLog2_Basic(t *testing.T) {
	input := `
		var result := Log2(8.0);
	`
	expectNoErrors(t, input)
}

// Rounding function tests
func TestBuiltinRound_Basic(t *testing.T) {
	input := `
		var result := Round(3.7);
	`
	expectNoErrors(t, input)
}

func TestBuiltinRound_Negative(t *testing.T) {
	input := `
		var result := Round(-3.7);
	`
	expectNoErrors(t, input)
}

func TestBuiltinTrunc_Basic(t *testing.T) {
	input := `
		var result := Trunc(3.7);
	`
	expectNoErrors(t, input)
}

func TestBuiltinFloor_Basic(t *testing.T) {
	input := `
		var result := Floor(3.7);
	`
	expectNoErrors(t, input)
}

func TestBuiltinCeil_Basic(t *testing.T) {
	input := `
		var result := Ceil(3.2);
	`
	expectNoErrors(t, input)
}

// Inc/Dec function tests
func TestBuiltinInc_NoStep(t *testing.T) {
	input := `
		var x := 5;
		Inc(x);
	`
	expectNoErrors(t, input)
}

func TestBuiltinInc_WithStep(t *testing.T) {
	input := `
		var x := 5;
		Inc(x, 3);
	`
	expectNoErrors(t, input)
}

func TestBuiltinDec_NoStep(t *testing.T) {
	input := `
		var x := 5;
		Dec(x);
	`
	expectNoErrors(t, input)
}

func TestBuiltinDec_WithStep(t *testing.T) {
	input := `
		var x := 5;
		Dec(x, 2);
	`
	expectNoErrors(t, input)
}

func TestBuiltinInc_NotVariable(t *testing.T) {
	input := `
		Inc(5);
	`
	expectError(t, input, "variable")
}

// Succ/Pred function tests
func TestBuiltinSucc_Integer(t *testing.T) {
	input := `
		var result := Succ(5);
	`
	expectNoErrors(t, input)
}

func TestBuiltinPred_Integer(t *testing.T) {
	input := `
		var result := Pred(5);
	`
	expectNoErrors(t, input)
}

func TestBuiltinSucc_Enum(t *testing.T) {
	input := `
		type TColor = (Red, Green, Blue);
		var c := Succ(Red);
	`
	expectNoErrors(t, input)
}

// Random function tests
func TestBuiltinRandom_NoArgs(t *testing.T) {
	input := `
		var result := Random();
	`
	expectNoErrors(t, input)
}

func TestBuiltinRandomInt_Basic(t *testing.T) {
	input := `
		var result := RandomInt(100);
	`
	expectNoErrors(t, input)
}

func TestBuiltinRandomize_Basic(t *testing.T) {
	input := `
		Randomize();
	`
	expectNoErrors(t, input)
}

func TestBuiltinSetRandSeed_Basic(t *testing.T) {
	input := `
		SetRandSeed(12345);
	`
	expectNoErrors(t, input)
}

// Other math functions
func TestBuiltinCoTan_Basic(t *testing.T) {
	input := `
		var result := CoTan(1.0);
	`
	expectNoErrors(t, input)
}

func TestBuiltinHypot_Basic(t *testing.T) {
	input := `
		var result := Hypot(3.0, 4.0);
	`
	expectNoErrors(t, input)
}

func TestBuiltinUnsigned32_Basic(t *testing.T) {
	input := `
		var result := Unsigned32(-1);
	`
	expectNoErrors(t, input)
}

func TestBuiltinMaxInt_Basic(t *testing.T) {
	input := `
		var result := MaxInt(5, 10, 3, 8);
	`
	expectNoErrors(t, input)
}

func TestBuiltinMinInt_Basic(t *testing.T) {
	input := `
		var result := MinInt(5, 10, 3, 8);
	`
	expectNoErrors(t, input)
}

func TestBuiltinSwap_Basic(t *testing.T) {
	input := `
		var a := 5;
		var b := 10;
		Swap(a, b);
	`
	expectNoErrors(t, input)
}

func TestBuiltinSwap_NotVariables(t *testing.T) {
	input := `
		Swap(5, 10);
	`
	expectError(t, input, "variable")
}

func TestBuiltinIsNaN_Basic(t *testing.T) {
	input := `
		var result := IsNaN(3.14);
	`
	expectNoErrors(t, input)
}

func TestBuiltinAssigned_Basic(t *testing.T) {
	input := `
		type TMyClass = class
		end;
		var obj: TMyClass;
		var result := Assigned(obj);
	`
	expectNoErrors(t, input)
}

// Combined math operations tests
func TestBuiltinMath_ChainedOperations(t *testing.T) {
	input := `
		var x := Abs(Sin(DegToRad(45.0)));
		var y := Round(Sqrt(Power(3.0, 2.0) + Power(4.0, 2.0)));
	`
	expectNoErrors(t, input)
}

func TestBuiltinMath_InExpressions(t *testing.T) {
	input := `
		var result := (Max(5, 10) + Min(3, 7)) * Abs(-2);
		var isPositive := result > 0;
	`
	expectNoErrors(t, input)
}

func TestBuiltinMath_AsParameters(t *testing.T) {
	input := `
		function Distance(x1, y1, x2, y2: Float): Float;
		begin
			Result := Sqrt(Power(x2 - x1, 2.0) + Power(y2 - y1, 2.0));
		end;

		var dist := Distance(0.0, 0.0, 3.0, 4.0);
	`
	expectNoErrors(t, input)
}

// Edge cases
func TestBuiltinMath_DivisionByZero(t *testing.T) {
	// Should analyze without error (runtime check)
	input := `
		var result := 10.0 / 0.0;
	`
	expectNoErrors(t, input)
}

func TestBuiltinMath_Overflow(t *testing.T) {
	// Should analyze without error (runtime check)
	input := `
		var result := Power(10.0, 1000.0);
	`
	expectNoErrors(t, input)
}

func TestBuiltinMath_Constants(t *testing.T) {
	input := `
		var pi := 3.14159265359;
		var e := 2.71828182846;
		var circle := 2.0 * pi * 5.0;
		var euler := Exp(1.0);
	`
	expectNoErrors(t, input)
}
