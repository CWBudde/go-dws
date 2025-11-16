package interp

import (
	"testing"
)

// ============================================================================
// Property Access Benchmarks
// ============================================================================

// BenchmarkPropertyReadField benchmarks field-backed property reads.
func BenchmarkPropertyReadField(b *testing.B) {
	input := `
type
	TPoint = class
	private
		FX: Integer;
		FY: Integer;
	public
		property X: Integer read FX write FX;
		property Y: Integer read FY write FY;
	end;

var p: TPoint;
var i, sum: Integer;
begin
	p := TPoint.Create;
	p.X := 10;
	p.Y := 20;

	sum := 0;
	for i := 1 to 1000 do
		sum := sum + p.X + p.Y;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// BenchmarkPropertyWriteField benchmarks field-backed property writes.
func BenchmarkPropertyWriteField(b *testing.B) {
	input := `
type
	TCounter = class
	private
		FValue: Integer;
	public
		property Value: Integer read FValue write FValue;
	end;

var c: TCounter;
var i: Integer;
begin
	c := TCounter.Create;

	for i := 1 to 1000 do
		c.Value := i;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// BenchmarkPropertyReadMethod benchmarks method-backed property reads.
func BenchmarkPropertyReadMethod(b *testing.B) {
	input := `
type
	TBox = class
	private
		FWidth: Integer;
		FHeight: Integer;
		function GetArea: Integer;
	public
		constructor Create(w, h: Integer);
		property Area: Integer read GetArea;
	end;

constructor TBox.Create(w, h: Integer);
begin
	FWidth := w;
	FHeight := h;
end;

function TBox.GetArea: Integer;
begin
	Result := FWidth * FHeight;
end;

var box: TBox;
var i, sum: Integer;
begin
	box := TBox.Create(10, 20);

	sum := 0;
	for i := 1 to 1000 do
		sum := sum + box.Area;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// BenchmarkPropertyWriteMethod benchmarks method-backed property writes.
func BenchmarkPropertyWriteMethod(b *testing.B) {
	input := `
type
	TValidator = class
	private
		FValue: Integer;
		procedure SetValue(v: Integer);
	public
		property Value: Integer read FValue write SetValue;
	end;

procedure TValidator.SetValue(v: Integer);
begin
	if v >= 0 then
		FValue := v
	else
		FValue := 0;
end;

var v: TValidator;
var i: Integer;
begin
	v := TValidator.Create;

	for i := 1 to 1000 do
		v.Value := i;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// ============================================================================
// Built-in Function Benchmarks
// ============================================================================

// BenchmarkBuiltinStringFunctions benchmarks string built-in functions.
func BenchmarkBuiltinStringFunctions(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{"Length", `
begin
	var s: String;
	var i, len: Integer;
	s := 'Hello World';
	for i := 1 to 1000 do
		len := Length(s);
end.
`},
		{"Pos", `
begin
	var s: String;
	var i, pos: Integer;
	s := 'Hello World';
	for i := 1 to 1000 do
		pos := Pos('World', s);
end.
`},
		{"Copy", `
begin
	var s, sub: String;
	var i: Integer;
	s := 'Hello World';
	for i := 1 to 1000 do
		sub := Copy(s, 1, 5);
end.
`},
		{"UpperCase", `
begin
	var s, upper: String;
	var i: Integer;
	s := 'Hello World';
	for i := 1 to 1000 do
		upper := UpperCase(s);
end.
`},
		{"LowerCase", `
begin
	var s, lower: String;
	var i: Integer;
	s := 'Hello World';
	for i := 1 to 1000 do
		lower := LowerCase(s);
end.
`},
		{"Trim", `
begin
	var s, trimmed: String;
	var i: Integer;
	s := '  Hello World  ';
	for i := 1 to 1000 do
		trimmed := Trim(s);
end.
`},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				runProgram(tc.input)
			}
		})
	}
}

// BenchmarkBuiltinMathFunctions benchmarks math built-in functions.
func BenchmarkBuiltinMathFunctions(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{"Abs", `
begin
	var i: Integer;
	var result: Integer;
	for i := 1 to 1000 do
		result := Abs(-42);
end.
`},
		{"Min", `
begin
	var i: Integer;
	var result: Integer;
	for i := 1 to 1000 do
		result := Min(10, 20);
end.
`},
		{"Max", `
begin
	var i: Integer;
	var result: Integer;
	for i := 1 to 1000 do
		result := Max(10, 20);
end.
`},
		{"Sqr", `
begin
	var i: Integer;
	var result: Integer;
	for i := 1 to 1000 do
		result := Sqr(7);
end.
`},
		{"Sqrt", `
begin
	var i: Integer;
	var result: Float;
	for i := 1 to 1000 do
		result := Sqrt(49.0);
end.
`},
		{"Power", `
begin
	var i: Integer;
	var result: Float;
	for i := 1 to 1000 do
		result := Power(2.0, 8.0);
end.
`},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				runProgram(tc.input)
			}
		})
	}
}

// BenchmarkBuiltinArrayFunctions benchmarks array built-in functions.
func BenchmarkBuiltinArrayFunctions(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{"SetLength", `
begin
	var arr: array of Integer;
	var i: Integer;
	for i := 1 to 100 do
		SetLength(arr, 100);
end.
`},
		{"Low", `
begin
	var arr: array of Integer;
	var i, low: Integer;
	SetLength(arr, 100);
	for i := 1 to 1000 do
		low := Low(arr);
end.
`},
		{"High", `
begin
	var arr: array of Integer;
	var i, high: Integer;
	SetLength(arr, 100);
	for i := 1 to 1000 do
		high := High(arr);
end.
`},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				runProgram(tc.input)
			}
		})
	}
}

// BenchmarkBuiltinConversionFunctions benchmarks conversion built-ins.
func BenchmarkBuiltinConversionFunctions(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{"IntToStr", `
begin
	var i: Integer;
	var s: String;
	for i := 1 to 1000 do
		s := IntToStr(42);
end.
`},
		{"StrToInt", `
begin
	var i: Integer;
	var n: Integer;
	for i := 1 to 1000 do
		n := StrToInt('42');
end.
`},
		{"FloatToStr", `
begin
	var i: Integer;
	var s: String;
	for i := 1 to 1000 do
		s := FloatToStr(3.14);
end.
`},
		{"StrToFloat", `
begin
	var i: Integer;
	var f: Float;
	for i := 1 to 1000 do
		f := StrToFloat('3.14');
end.
`},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				runProgram(tc.input)
			}
		})
	}
}

// ============================================================================
// Exception Handling Benchmarks
// ============================================================================

// BenchmarkExceptionRaise benchmarks raising exceptions.
func BenchmarkExceptionRaise(b *testing.B) {
	input := `
var i: Integer;
begin
	for i := 1 to 100 do
	begin
		try
			raise Exception.Create('Error');
		except
			on E: Exception do
				; // Do nothing
		end;
	end;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// BenchmarkExceptionTryCatch benchmarks try-except overhead.
func BenchmarkExceptionTryCatch(b *testing.B) {
	input := `
function SafeDivide(a, b: Integer): Integer;
begin
	try
		if b = 0 then
			raise Exception.Create('Division by zero');
		Result := a div b;
	except
		on E: Exception do
			Result := 0;
	end;
end;

var i, result: Integer;
begin
	for i := 1 to 1000 do
		result := SafeDivide(100, 5);
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// BenchmarkExceptionFinally benchmarks try-finally.
func BenchmarkExceptionFinally(b *testing.B) {
	input := `
var i, counter: Integer;
begin
	counter := 0;
	for i := 1 to 1000 do
	begin
		try
			counter := counter + 1;
		finally
			counter := counter + 1;
		end;
	end;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// ============================================================================
// Type Operation Benchmarks
// ============================================================================

// BenchmarkTypeChecks benchmarks 'is' type checks.
func BenchmarkTypeChecks(b *testing.B) {
	input := `
type
	TBase = class
	end;

	TDerived = class(TBase)
	end;

var base: TBase;
var derived: TDerived;
var i: Integer;
var check: Boolean;
begin
	derived := TDerived.Create;
	base := derived;

	for i := 1 to 1000 do
		check := base is TDerived;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// BenchmarkTypeCasts benchmarks 'as' type casts.
func BenchmarkTypeCasts(b *testing.B) {
	input := `
type
	TBase = class
	end;

	TDerived = class(TBase)
		Value: Integer;
	end;

var base: TBase;
var derived: TDerived;
var i: Integer;
begin
	derived := TDerived.Create;
	base := derived;

	for i := 1 to 1000 do
	begin
		var d: TDerived;
		d := base as TDerived;
	end;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// BenchmarkTypeOf benchmarks TypeOf operation.
func BenchmarkTypeOf(b *testing.B) {
	input := `
type
	TMyClass = class
		Value: Integer;
	end;

var obj: TMyClass;
var i: Integer;
var typeName: String;
begin
	obj := TMyClass.Create;

	for i := 1 to 1000 do
		typeName := TypeOf(obj);
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// ============================================================================
// Variant Operation Benchmarks
// ============================================================================

// BenchmarkVariantAssignment benchmarks variant assignments.
func BenchmarkVariantAssignment(b *testing.B) {
	input := `
var v: Variant;
var i: Integer;
begin
	for i := 1 to 1000 do
	begin
		v := 42;
		v := 'Hello';
		v := 3.14;
		v := True;
	end;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// BenchmarkVariantOperations benchmarks variant arithmetic.
func BenchmarkVariantOperations(b *testing.B) {
	input := `
var v1, v2, v3: Variant;
var i: Integer;
begin
	v1 := 10;
	v2 := 20;

	for i := 1 to 1000 do
	begin
		v3 := v1 + v2;
		v3 := v1 * v2;
		v3 := v2 - v1;
	end;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// ============================================================================
// Interface Operation Benchmarks
// ============================================================================

// BenchmarkInterfaceMethodCall benchmarks interface method calls.
func BenchmarkInterfaceMethodCall(b *testing.B) {
	input := `
type
	ICounter = interface
		procedure Increment;
		function GetValue: Integer;
	end;

	TCounter = class(ICounter)
	private
		FValue: Integer;
	public
		procedure Increment;
		function GetValue: Integer;
	end;

procedure TCounter.Increment;
begin
	FValue := FValue + 1;
end;

function TCounter.GetValue: Integer;
begin
	Result := FValue;
end;

var intf: ICounter;
var i, val: Integer;
begin
	intf := TCounter.Create;

	for i := 1 to 1000 do
	begin
		intf.Increment;
		val := intf.GetValue;
	end;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// BenchmarkInterfaceImplementsCheck benchmarks 'implements' checks.
func BenchmarkInterfaceImplementsCheck(b *testing.B) {
	input := `
type
	ICounter = interface
		function GetValue: Integer;
	end;

	TCounter = class(ICounter)
		function GetValue: Integer;
	end;

function TCounter.GetValue: Integer;
begin
	Result := 42;
end;

var obj: TCounter;
var i: Integer;
var check: Boolean;
begin
	obj := TCounter.Create;

	for i := 1 to 1000 do
		check := obj implements ICounter;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// ============================================================================
// Enum and Set Operation Benchmarks
// ============================================================================

// BenchmarkEnumOperations benchmarks enum operations.
func BenchmarkEnumOperations(b *testing.B) {
	input := `
type
	TColor = (clRed, clGreen, clBlue, clYellow, clBlack, clWhite);

var color: TColor;
var i: Integer;
var ord: Integer;
begin
	color := clRed;

	for i := 1 to 1000 do
	begin
		color := clBlue;
		ord := Ord(color);
		color := TColor(ord + 1);
	end;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// BenchmarkSetInclusion benchmarks 'in' operator for sets.
func BenchmarkSetInclusion(b *testing.B) {
	input := `
type
	TColor = (clRed, clGreen, clBlue, clYellow, clBlack, clWhite);
	TColors = set of TColor;

var colors: TColors;
var i: Integer;
var check: Boolean;
begin
	colors := [clRed, clBlue, clWhite];

	for i := 1 to 1000 do
	begin
		check := clRed in colors;
		check := clGreen in colors;
		check := clBlue in colors;
	end;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// ============================================================================
// Object Creation and Destruction Benchmarks
// ============================================================================

// BenchmarkObjectCreation benchmarks object instantiation.
func BenchmarkObjectCreation(b *testing.B) {
	input := `
type
	TSimple = class
		Value: Integer;
	end;

var i: Integer;
var obj: TSimple;
begin
	for i := 1 to 100 do
		obj := TSimple.Create;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// BenchmarkObjectWithConstructor benchmarks constructor calls.
func BenchmarkObjectWithConstructor(b *testing.B) {
	input := `
type
	TPoint = class
	private
		FX, FY: Integer;
	public
		constructor Create(x, y: Integer);
	end;

constructor TPoint.Create(x, y: Integer);
begin
	FX := x;
	FY := y;
end;

var i: Integer;
var p: TPoint;
begin
	for i := 1 to 100 do
		p := TPoint.Create(10, 20);
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// BenchmarkInheritedMethodCall benchmarks inherited method calls.
func BenchmarkInheritedMethodCall(b *testing.B) {
	input := `
type
	TBase = class
		function GetValue: Integer; virtual;
	end;

	TDerived = class(TBase)
		function GetValue: Integer; override;
	end;

function TBase.GetValue: Integer;
begin
	Result := 42;
end;

function TDerived.GetValue: Integer;
begin
	Result := inherited GetValue + 1;
end;

var obj: TDerived;
var i, val: Integer;
begin
	obj := TDerived.Create;

	for i := 1 to 1000 do
		val := obj.GetValue;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// ============================================================================
// Lambda and Function Pointer Benchmarks
// ============================================================================

// BenchmarkLambdaCall benchmarks lambda function calls.
func BenchmarkLambdaCall(b *testing.B) {
	input := `
type
	TFunc = function(x: Integer): Integer;

var f: TFunc;
var i, result: Integer;
begin
	f := lambda(x: Integer): Integer => x * 2;

	for i := 1 to 1000 do
		result := f(i);
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// BenchmarkFunctionPointerCall benchmarks function pointer calls.
func BenchmarkFunctionPointerCall(b *testing.B) {
	input := `
function Double(x: Integer): Integer;
begin
	Result := x * 2;
end;

type
	TFunc = function(x: Integer): Integer;

var f: TFunc;
var i, result: Integer;
begin
	f := @Double;

	for i := 1 to 1000 do
		result := f(i);
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}
