package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// BenchmarkParser benchmarks the overall parser performance with a realistic DWScript program
func BenchmarkParser(b *testing.B) {
	// Realistic DWScript program with functions, classes, and control flow
	input := `
function Fibonacci(n: Integer): Integer;
begin
	if n <= 1 then
		Result := n
	else
		Result := Fibonacci(n-1) + Fibonacci(n-2);
end;

function IsPrime(n: Integer): Boolean;
var
	i: Integer;
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

type
	TPerson = class
	private
		FName: String;
		FAge: Integer;
	public
		constructor Create(name: String; age: Integer);
		function GetName: String;
		procedure SetAge(age: Integer);
		property Name: String read FName write FName;
		property Age: Integer read FAge write SetAge;
	end;

constructor TPerson.Create(name: String; age: Integer);
begin
	FName := name;
	FAge := age;
end;

function TPerson.GetName: String;
begin
	Result := FName;
end;

procedure TPerson.SetAge(age: Integer);
begin
	if age >= 0 then
		FAge := age;
end;

var
	i, j, sum: Integer;
	person: TPerson;
begin
	sum := 0;
	for i := 1 to 10 do
		for j := 1 to 10 do
			sum := sum + i * j;

	person := TPerson.Create('John', 30);
	PrintLn(person.Name + ' is ' + IntToStr(person.Age));
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := lexer.New(input)
		p := New(l)
		_ = p.ParseProgram()
	}
}

// BenchmarkParserExpressions benchmarks parsing of various expressions
func BenchmarkParserExpressions(b *testing.B) {
	input := `
begin
	x := 1 + 2 * 3 - 4 / 5;
	y := (a + b) * (c - d);
	z := not flag and (value > 100) or (count <= 0);
	result := func1(arg1, arg2) + func2(arg3);
	arr[i] := obj.field.method();
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := lexer.New(input)
		p := New(l)
		_ = p.ParseProgram()
	}
}

// BenchmarkParserFunctions benchmarks parsing of function declarations
func BenchmarkParserFunctions(b *testing.B) {
	input := `
function Add(a, b: Integer): Integer;
begin
	Result := a + b;
end;

function Subtract(a, b: Integer): Integer;
begin
	Result := a - b;
end;

function Multiply(a, b: Integer): Integer;
begin
	Result := a * b;
end;

function Divide(a, b: Integer): Float;
begin
	if b <> 0 then
		Result := a / b
	else
		Result := 0.0;
end;

procedure PrintValues(x, y, z: Integer);
begin
	PrintLn(IntToStr(x));
	PrintLn(IntToStr(y));
	PrintLn(IntToStr(z));
end;
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := lexer.New(input)
		p := New(l)
		_ = p.ParseProgram()
	}
}

// BenchmarkParserClasses benchmarks parsing of class declarations
func BenchmarkParserClasses(b *testing.B) {
	input := `
type
	TAnimal = class
	private
		FName: String;
		FAge: Integer;
	public
		constructor Create(name: String; age: Integer);
		destructor Destroy; override;
		function GetName: String;
		function GetAge: Integer;
		procedure MakeSound; virtual; abstract;
		property Name: String read FName write FName;
		property Age: Integer read FAge write FAge;
	end;

	TDog = class(TAnimal)
	private
		FBreed: String;
	public
		constructor Create(name: String; age: Integer; breed: String);
		procedure MakeSound; override;
		property Breed: String read FBreed write FBreed;
	end;

	TCat = class(TAnimal)
	private
		FColor: String;
	public
		constructor Create(name: String; age: Integer; color: String);
		procedure MakeSound; override;
		property Color: String read FColor write FColor;
	end;

constructor TAnimal.Create(name: String; age: Integer);
begin
	FName := name;
	FAge := age;
end;

destructor TAnimal.Destroy;
begin
	// Cleanup
end;

function TAnimal.GetName: String;
begin
	Result := FName;
end;

function TAnimal.GetAge: Integer;
begin
	Result := FAge;
end;

constructor TDog.Create(name: String; age: Integer; breed: String);
begin
	inherited Create(name, age);
	FBreed := breed;
end;

procedure TDog.MakeSound;
begin
	PrintLn('Woof!');
end;

constructor TCat.Create(name: String; age: Integer; color: String);
begin
	inherited Create(name, age);
	FColor := color;
end;

procedure TCat.MakeSound;
begin
	PrintLn('Meow!');
end;
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := lexer.New(input)
		p := New(l)
		_ = p.ParseProgram()
	}
}

// BenchmarkParserControlFlow benchmarks parsing of control flow statements
func BenchmarkParserControlFlow(b *testing.B) {
	input := `
begin
	if x > 0 then
	begin
		PrintLn('Positive');
	end
	else if x < 0 then
	begin
		PrintLn('Negative');
	end
	else
	begin
		PrintLn('Zero');
	end;

	while count < 100 do
	begin
		count := count + 1;
		if count mod 2 = 0 then
			Continue;
		PrintLn(IntToStr(count));
	end;

	for i := 1 to 10 do
	begin
		for j := 1 to 10 do
		begin
			if i * j > 50 then
				Break;
			PrintLn(IntToStr(i * j));
		end;
	end;

	repeat
		value := value * 2;
	until value > 1000;

	case grade of
		'A': PrintLn('Excellent');
		'B': PrintLn('Good');
		'C': PrintLn('Fair');
		'D': PrintLn('Poor');
		'F': PrintLn('Fail');
	else
		PrintLn('Invalid grade');
	end;

	try
		result := Divide(x, y);
	except
		on E: Exception do
			PrintLn('Error: ' + E.Message);
	end;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := lexer.New(input)
		p := New(l)
		_ = p.ParseProgram()
	}
}

// BenchmarkParserTypes benchmarks parsing of type declarations
func BenchmarkParserTypes(b *testing.B) {
	input := `
type
	TStatus = (Pending, Active, Completed, Cancelled);
	TPriority = (Low, Medium, High, Critical);

	TPoint = record
		X: Integer;
		Y: Integer;
	end;

	TRectangle = record
		TopLeft: TPoint;
		BottomRight: TPoint;
	end;

	TIntArray = array of Integer;
	TStringArray = array of String;
	TMatrix = array of array of Float;

	TCallback = function(value: Integer): Boolean;
	TComparator = function(a, b: String): Integer;
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := lexer.New(input)
		p := New(l)
		_ = p.ParseProgram()
	}
}

// BenchmarkParserArrays benchmarks parsing of array operations
func BenchmarkParserArrays(b *testing.B) {
	input := `
begin
	arr := [1, 2, 3, 4, 5];
	matrix := [[1, 2, 3], [4, 5, 6], [7, 8, 9]];

	for i := 0 to High(arr) do
		arr[i] := arr[i] * 2;

	for i := 0 to High(matrix) do
		for j := 0 to High(matrix[i]) do
			matrix[i][j] := matrix[i][j] + 1;

	SetLength(arr, 10);
	arr[9] := 100;

	total := 0;
	for value in arr do
		total := total + value;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := lexer.New(input)
		p := New(l)
		_ = p.ParseProgram()
	}
}

// BenchmarkParserStrings benchmarks parsing of string operations
func BenchmarkParserStrings(b *testing.B) {
	input := `
begin
	s1 := 'Hello';
	s2 := 'World';
	s3 := s1 + ' ' + s2;

	len := Length(s3);
	upper := UpperCase(s3);
	lower := LowerCase(s3);

	if Pos('Hello', s3) > 0 then
		PrintLn('Found!');

	parts := Split(s3, ' ');
	joined := Join(parts, '-');

	sub := Copy(s3, 1, 5);
	s3 := StringReplace(s3, 'Hello', 'Hi');
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := lexer.New(input)
		p := New(l)
		_ = p.ParseProgram()
	}
}

// BenchmarkParserSmallProgram benchmarks parsing of a small, simple program
func BenchmarkParserSmallProgram(b *testing.B) {
	input := `
var x, y: Integer;
begin
	x := 10;
	y := 20;
	PrintLn(IntToStr(x + y));
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := lexer.New(input)
		p := New(l)
		_ = p.ParseProgram()
	}
}

// BenchmarkParserLargeProgram benchmarks parsing of a large, complex program
func BenchmarkParserLargeProgram(b *testing.B) {
	// Build a large program by repeating structures
	var input string
	for i := 0; i < 20; i++ {
		input += `
function Func` + string(rune('A'+i%26)) + `(n: Integer): Integer;
var i, result: Integer;
begin
	result := 0;
	for i := 1 to n do
		result := result + i;
	Result := result;
end;
`
	}

	input += `
begin
	var total: Integer;
	total := 0;
`

	for i := 0; i < 20; i++ {
		input += `	total := total + Func` + string(rune('A'+i%26)) + `(10);
`
	}

	input += `	PrintLn(IntToStr(total));
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := lexer.New(input)
		p := New(l)
		_ = p.ParseProgram()
	}
}

// BenchmarkParserParameterLists benchmarks parsing of functions with many parameters
func BenchmarkParserParameterLists(b *testing.B) {
	input := `
function ManyParams(
	a1, a2, a3, a4, a5: Integer;
	b1, b2, b3, b4, b5: String;
	c1, c2, c3, c4, c5: Float;
	d1, d2, d3, d4, d5: Boolean
): Integer;
begin
	Result := a1 + a2 + a3 + a4 + a5;
end;

procedure ProcessData(
	x, y, z: Integer;
	name, description, category: String;
	enabled, visible, active: Boolean;
	weight, height, depth: Float
);
begin
	PrintLn(name);
end;

begin
	var result: Integer;
	result := ManyParams(
		1, 2, 3, 4, 5,
		'a', 'b', 'c', 'd', 'e',
		1.0, 2.0, 3.0, 4.0, 5.0,
		True, False, True, False, True
	);

	ProcessData(
		10, 20, 30,
		'Item', 'Description', 'Category',
		True, True, False,
		1.5, 2.5, 3.5
	);
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := lexer.New(input)
		p := New(l)
		_ = p.ParseProgram()
	}
}

// BenchmarkParserComplexNestedTypes benchmarks parsing of deeply nested type declarations
func BenchmarkParserComplexNestedTypes(b *testing.B) {
	input := `
type
	TInnerRecord = record
		Value: Integer;
		Name: String;
	end;

	TMiddleRecord = record
		Inner: TInnerRecord;
		Data: array of Float;
		Status: Boolean;
	end;

	TOuterRecord = record
		Middle: TMiddleRecord;
		Items: array of TMiddleRecord;
		Callback: function(r: TMiddleRecord): Boolean;
	end;

	TComplexArray = array of array of array of Integer;
	TNestedFunction = function(f: function(x: Integer): Integer): function(y: Integer): Integer;

	TGenericCallback = function(
		data: array of record
			id: Integer;
			values: array of Float;
		end
	): array of String;

	TNode = class
	private
		FData: TOuterRecord;
		FChildren: array of TNode;
		FParent: TNode;
	public
		constructor Create(parent: TNode);
		function FindChild(predicate: function(n: TNode): Boolean): TNode;
		property Data: TOuterRecord read FData write FData;
		property Children: array of TNode read FChildren;
	end;
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := lexer.New(input)
		p := New(l)
		_ = p.ParseProgram()
	}
}

// BenchmarkParserContracts benchmarks parsing of contracts (require/ensure)
func BenchmarkParserContracts(b *testing.B) {
	input := `
function Divide(a, b: Integer): Float;
require
	b <> 0;
begin
	Result := a / b;
ensure
	Result * b = a;
end;

function Sqrt(x: Float): Float;
require
	x >= 0.0;
var
	result: Float;
begin
	result := x; // simplified
	Result := result;
ensure
	Result >= 0.0;
	Result * Result <= x + 0.001;
	Result * Result >= x - 0.001;
end;

type
	TBankAccount = class
	private
		FBalance: Float;
	public
		constructor Create(initialBalance: Float);
		require
			initialBalance >= 0.0;

		procedure Deposit(amount: Float);
		require
			amount > 0.0;
		ensure
			FBalance = old(FBalance) + amount;

		procedure Withdraw(amount: Float);
		require
			amount > 0.0;
			amount <= FBalance;
		ensure
			FBalance = old(FBalance) - amount;

		property Balance: Float read FBalance;
	end;
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := lexer.New(input)
		p := New(l)
		_ = p.ParseProgram()
	}
}

// BenchmarkParserMemberAccessChains benchmarks parsing of deep member access chains
func BenchmarkParserMemberAccessChains(b *testing.B) {
	input := `
begin
	result := obj.field1.field2.field3.method1().field4.method2(arg).field5;

	value := database.connection.session.transaction.query.results[0].columns['name'].value.ToString();

	instance := factory.builder.configure().withOptions(opts).build().getInstance().initialize();

	data := app.services.userManager.getUser(id).profile.settings.preferences.theme.colors.primary;

	config := system.registry.get('config').section('database').value('connection').parse().validate();

	chain := a.b.c.d.e.f.g.h.i.j.k.l.m.n.o.p.q.r.s.t.u.v.w.x.y.z;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := lexer.New(input)
		p := New(l)
		_ = p.ParseProgram()
	}
}

// BenchmarkParserErrorRecovery benchmarks parser performance with syntax errors
func BenchmarkParserErrorRecovery(b *testing.B) {
	// Program with multiple syntax errors to test error recovery
	input := `
function BrokenFunc(x: Integer): Integer
begin
	if x > 0 then
		Result := x  // Missing semicolon before else
	else
		Result := -x
end

var a, b: Integer  // Missing semicolon
var c: String := "test"  // Missing semicolon

begin
	x := 10
	// Missing semicolon
	y := 20

	if x > y then
		PrintLn('x is greater')
	// Missing else or semicolon

	for i := 1 to  // Missing end value
		PrintLn(IntToStr(i))

	while  // Missing condition
	begin
		count := count + 1;
	end
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := lexer.New(input)
		p := New(l)
		_ = p.ParseProgram()
		// Parser should accumulate errors but continue parsing
	}
}

// BenchmarkParserWithAllocations benchmarks memory allocations during parsing
func BenchmarkParserWithAllocations(b *testing.B) {
	input := `
function Fibonacci(n: Integer): Integer;
begin
	if n <= 1 then
		Result := n
	else
		Result := Fibonacci(n-1) + Fibonacci(n-2);
end;

type
	TPerson = class
	private
		FName: String;
		FAge: Integer;
	public
		constructor Create(name: String; age: Integer);
		function GetName: String;
		property Name: String read FName write FName;
		property Age: Integer read FAge write FAge;
	end;

var
	arr: array of Integer;
	person: TPerson;
begin
	arr := [1, 2, 3, 4, 5];
	person := TPerson.Create('John', 30);
	PrintLn(person.Name);
end.
`

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := lexer.New(input)
		p := New(l)
		_ = p.ParseProgram()
	}
}

// BenchmarkParserExpressionLists benchmarks parsing of long expression lists
func BenchmarkParserExpressionLists(b *testing.B) {
	input := `
begin
	arr := [
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
		11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
		21, 22, 23, 24, 25, 26, 27, 28, 29, 30
	];

	result := func(
		arg1, arg2, arg3, arg4, arg5,
		arg6, arg7, arg8, arg9, arg10,
		arg11, arg12, arg13, arg14, arg15
	);

	matrix := [
		[1, 2, 3, 4, 5],
		[6, 7, 8, 9, 10],
		[11, 12, 13, 14, 15],
		[16, 17, 18, 19, 20],
		[21, 22, 23, 24, 25]
	];

	sum := 1 + 2 + 3 + 4 + 5 + 6 + 7 + 8 + 9 + 10 +
	      11 + 12 + 13 + 14 + 15 + 16 + 17 + 18 + 19 + 20;

	product := a * b * c * d * e * f * g * h * i * j;
end.
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := lexer.New(input)
		p := New(l)
		_ = p.ParseProgram()
	}
}
