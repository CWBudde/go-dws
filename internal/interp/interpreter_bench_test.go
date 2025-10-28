package interp

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
)

// helper function to run the full pipeline: lexer -> parser -> analyzer -> interpreter
func runProgram(input string) {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		return // ignore parse errors for benchmarking
	}

	analyzer := semantic.NewAnalyzer()
	err := analyzer.Analyze(program)
	if err != nil {
		return // ignore semantic errors for benchmarking
	}

	output := &bytes.Buffer{}
	interp := New(output)
	_ = interp.Eval(program)
}

// BenchmarkInterpreter benchmarks the overall interpreter performance
func BenchmarkInterpreter(b *testing.B) {
	input := `
function Fibonacci(n: Integer): Integer;
begin
	if n <= 1 then
		Result := n
	else
		Result := Fibonacci(n-1) + Fibonacci(n-2);
end;

var i, result: Integer;
begin
	for i := 0 to 15 do
	begin
		result := Fibonacci(i);
	end;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// BenchmarkInterpreterFibonacci benchmarks recursive function calls
func BenchmarkInterpreterFibonacci(b *testing.B) {
	// Test different Fibonacci numbers
	testCases := []struct {
		name  string
		input string
		n     int
	}{
		{"Fib10", 10, `
function Fibonacci(n: Integer): Integer;
begin
	if n <= 1 then
		Result := n
	else
		Result := Fibonacci(n-1) + Fibonacci(n-2);
end;

begin
	var x: Integer;
	x := Fibonacci(10);
end.
`},
		{"Fib15", 15, `
function Fibonacci(n: Integer): Integer;
begin
	if n <= 1 then
		Result := n
	else
		Result := Fibonacci(n-1) + Fibonacci(n-2);
end;

begin
	var x: Integer;
	x := Fibonacci(15);
end.
`},
		{"Fib20", 20, `
function Fibonacci(n: Integer): Integer;
begin
	if n <= 1 then
		Result := n
	else
		Result := Fibonacci(n-1) + Fibonacci(n-2);
end;

begin
	var x: Integer;
	x := Fibonacci(20);
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

// BenchmarkInterpreterPrimes benchmarks loops and conditionals
func BenchmarkInterpreterPrimes(b *testing.B) {
	input := `
function IsPrime(n: Integer): Boolean;
var i: Integer;
begin
	if n <= 1 then
	begin
		Result := False;
		Exit;
	end;

	Result := True;
	for i := 2 to n - 1 do
	begin
		if (n mod i) = 0 then
		begin
			Result := False;
			Break;
		end;
	end;
end;

var n, count: Integer;
begin
	count := 0;
	for n := 2 to 100 do
	begin
		if IsPrime(n) then
			count := count + 1;
	end;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// BenchmarkInterpreterLoops benchmarks different loop types
func BenchmarkInterpreterLoops(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{"ForLoop", `
begin
	var sum, i: Integer;
	sum := 0;
	for i := 1 to 1000 do
		sum := sum + i;
end.
`},
		{"WhileLoop", `
begin
	var sum, i: Integer;
	sum := 0;
	i := 1;
	while i <= 1000 do
	begin
		sum := sum + i;
		i := i + 1;
	end;
end.
`},
		{"RepeatUntil", `
begin
	var sum, i: Integer;
	sum := 0;
	i := 1;
	repeat
		sum := sum + i;
		i := i + 1;
	until i > 1000;
end.
`},
		{"NestedLoops", `
begin
	var sum, i, j: Integer;
	sum := 0;
	for i := 1 to 50 do
		for j := 1 to 50 do
			sum := sum + 1;
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

// BenchmarkInterpreterArithmetic benchmarks arithmetic operations
func BenchmarkInterpreterArithmetic(b *testing.B) {
	input := `
begin
	var a, b, c, d, e: Integer;
	var i: Integer;

	a := 1;
	b := 2;
	c := 3;
	d := 4;
	e := 5;

	for i := 1 to 1000 do
	begin
		a := b + c;
		b := c - d;
		c := d * e;
		d := a div 2;
		e := b mod 3;
	end;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// BenchmarkInterpreterStrings benchmarks string operations
func BenchmarkInterpreterStrings(b *testing.B) {
	input := `
begin
	var s1, s2, s3: String;
	var i, len: Integer;

	s1 := 'Hello';
	s2 := 'World';

	for i := 1 to 100 do
	begin
		s3 := s1 + ' ' + s2;
		len := Length(s3);
		s3 := UpperCase(s3);
		s3 := LowerCase(s3);
	end;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// BenchmarkInterpreterArrays benchmarks array operations
func BenchmarkInterpreterArrays(b *testing.B) {
	input := `
begin
	var arr: array of Integer;
	var i, sum: Integer;

	SetLength(arr, 100);

	for i := 0 to 99 do
		arr[i] := i * 2;

	sum := 0;
	for i := 0 to 99 do
		sum := sum + arr[i];

	for i := 0 to 99 do
		arr[i] := arr[i] + 1;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// BenchmarkInterpreterClasses benchmarks class instantiation and method calls
func BenchmarkInterpreterClasses(b *testing.B) {
	input := `
type
	TCounter = class
	private
		FValue: Integer;
	public
		constructor Create(initial: Integer);
		procedure Increment;
		procedure Decrement;
		function GetValue: Integer;
		procedure Reset;
	end;

constructor TCounter.Create(initial: Integer);
begin
	FValue := initial;
end;

procedure TCounter.Increment;
begin
	FValue := FValue + 1;
end;

procedure TCounter.Decrement;
begin
	FValue := FValue - 1;
end;

function TCounter.GetValue: Integer;
begin
	Result := FValue;
end;

procedure TCounter.Reset;
begin
	FValue := 0;
end;

var counter: TCounter;
var i: Integer;
begin
	counter := TCounter.Create(0);

	for i := 1 to 100 do
	begin
		counter.Increment;
		counter.Increment;
		counter.Decrement;
	end;

	counter.Reset;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// BenchmarkInterpreterRecords benchmarks record operations
func BenchmarkInterpreterRecords(b *testing.B) {
	input := `
type
	TPoint = record
		X: Integer;
		Y: Integer;
	end;

	TRectangle = record
		TopLeft: TPoint;
		BottomRight: TPoint;
	end;

function CalculateArea(rect: TRectangle): Integer;
begin
	Result := (rect.BottomRight.X - rect.TopLeft.X) *
	          (rect.BottomRight.Y - rect.TopLeft.Y);
end;

var rect: TRectangle;
var i, area: Integer;
begin
	rect.TopLeft.X := 0;
	rect.TopLeft.Y := 0;

	for i := 1 to 100 do
	begin
		rect.BottomRight.X := i;
		rect.BottomRight.Y := i;
		area := CalculateArea(rect);
	end;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// BenchmarkInterpreterConditionals benchmarks conditional statements
func BenchmarkInterpreterConditionals(b *testing.B) {
	input := `
begin
	var i, positive, negative, zero: Integer;

	positive := 0;
	negative := 0;
	zero := 0;

	for i := -500 to 500 do
	begin
		if i > 0 then
			positive := positive + 1
		else if i < 0 then
			negative := negative + 1
		else
			zero := zero + 1;
	end;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// BenchmarkInterpreterFunctionCalls benchmarks function call overhead
func BenchmarkInterpreterFunctionCalls(b *testing.B) {
	input := `
function Add(a, b: Integer): Integer;
begin
	Result := a + b;
end;

function Multiply(a, b: Integer): Integer;
begin
	Result := a * b;
end;

function Calculate(x, y, z: Integer): Integer;
begin
	Result := Add(Multiply(x, y), z);
end;

var i, result: Integer;
begin
	for i := 1 to 1000 do
		result := Calculate(i, i + 1, i + 2);
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runProgram(input)
	}
}

// BenchmarkInterpreterVariableAccess benchmarks variable access patterns
func BenchmarkInterpreterVariableAccess(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{"LocalVars", `
begin
	var a, b, c, d, e: Integer;
	var i: Integer;

	for i := 1 to 1000 do
	begin
		a := 1;
		b := a + 1;
		c := b + 1;
		d := c + 1;
		e := d + 1;
	end;
end.
`},
		{"GlobalVars", `
var a, b, c, d, e: Integer;
begin
	var i: Integer;

	for i := 1 to 1000 do
	begin
		a := 1;
		b := a + 1;
		c := b + 1;
		d := c + 1;
		e := d + 1;
	end;
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

// BenchmarkInterpreterFullPipeline benchmarks the complete pipeline with various programs
func BenchmarkInterpreterFullPipeline(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{"TinyProgram", `
begin
	var x: Integer;
	x := 42;
end.
`},
		{"SmallProgram", `
var sum, i: Integer;
begin
	sum := 0;
	for i := 1 to 100 do
		sum := sum + i;
end.
`},
		{"MediumProgram", `
function Factorial(n: Integer): Integer;
begin
	if n <= 1 then
		Result := 1
	else
		Result := n * Factorial(n - 1);
end;

var i, result: Integer;
begin
	for i := 1 to 10 do
		result := Factorial(i);
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
