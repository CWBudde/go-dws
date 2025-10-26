/**
 * DWScript Example Programs
 *
 * A collection of example programs that demonstrate various
 * features of the DWScript language.
 */

const EXAMPLES = {
    hello: {
        name: 'Hello World',
        description: 'A simple Hello World program',
        code: `// Hello World in DWScript
PrintLn('Hello, World!');
PrintLn('Welcome to the DWScript Playground!');`
    },

    fibonacci: {
        name: 'Fibonacci Sequence',
        description: 'Calculate Fibonacci numbers using recursion',
        code: `// Fibonacci Sequence
function Fibonacci(n: Integer): Integer;
begin
    if n <= 1 then
        Result := n
    else
        Result := Fibonacci(n - 1) + Fibonacci(n - 2);
end;

var i: Integer;

PrintLn('First 10 Fibonacci numbers:');
for i := 0 to 9 do
    PrintLn('F(' + IntToStr(i) + ') = ' + IntToStr(Fibonacci(i)));`
    },

    factorial: {
        name: 'Factorial',
        description: 'Calculate factorial using recursion and iteration',
        code: `// Factorial - Recursive and Iterative

function FactorialRecursive(n: Integer): Integer;
begin
    if n <= 1 then
        Result := 1
    else
        Result := n * FactorialRecursive(n - 1);
end;

function FactorialIterative(n: Integer): Integer;
var
    i: Integer;
begin
    Result := 1;
    for i := 2 to n do
        Result := Result * i;
end;

var n: Integer;

n := 10;
PrintLn('Factorial of ' + IntToStr(n) + ':');
PrintLn('  Recursive: ' + IntToStr(FactorialRecursive(n)));
PrintLn('  Iterative: ' + IntToStr(FactorialIterative(n)));`
    },

    loops: {
        name: 'Loops',
        description: 'Demonstrate different loop structures',
        code: `// Loop Examples in DWScript

var i: Integer;

// For loop (ascending)
PrintLn('For loop (1 to 5):');
for i := 1 to 5 do
    PrintLn('  ' + IntToStr(i));

// For loop (descending)
PrintLn('');
PrintLn('For loop (5 downto 1):');
for i := 5 downto 1 do
    PrintLn('  ' + IntToStr(i));

// While loop
PrintLn('');
PrintLn('While loop:');
i := 0;
while i < 5 do
begin
    PrintLn('  i = ' + IntToStr(i));
    i := i + 1;
end;

// Repeat-until loop
PrintLn('');
PrintLn('Repeat-until loop:');
i := 0;
repeat
    PrintLn('  i = ' + IntToStr(i));
    i := i + 1;
until i >= 5;`
    },

    functions: {
        name: 'Functions',
        description: 'Functions, procedures, and parameters',
        code: `// Functions and Procedures

// Function with return value
function Add(a, b: Integer): Integer;
begin
    Result := a + b;
end;

// Function with multiple parameters
function Max(a, b: Integer): Integer;
begin
    if a > b then
        Result := a
    else
        Result := b;
end;

// Procedure (no return value)
procedure Greet(name: String);
begin
    PrintLn('Hello, ' + name + '!');
end;

// String manipulation function
function Capitalize(s: String): String;
begin
    if Length(s) > 0 then
        Result := UpperCase(Copy(s, 1, 1)) + LowerCase(Copy(s, 2, Length(s) - 1))
    else
        Result := s;
end;

// Main program
var sum, maximum: Integer;
var greeting: String;

sum := Add(15, 27);
maximum := Max(42, 17);

PrintLn('15 + 27 = ' + IntToStr(sum));
PrintLn('Max(42, 17) = ' + IntToStr(maximum));

Greet('World');
Greet('DWScript');

greeting := Capitalize('hello');
PrintLn('Capitalized: ' + greeting);`
    },

    classes: {
        name: 'Classes (OOP)',
        description: 'Object-oriented programming with classes',
        code: `// Object-Oriented Programming

type
    TPerson = class
    private
        FName: String;
        FAge: Integer;
    public
        constructor Create(aName: String; aAge: Integer);
        procedure Introduce;
        procedure HaveBirthday;

        property Name: String read FName write FName;
        property Age: Integer read FAge write FAge;
    end;

constructor TPerson.Create(aName: String; aAge: Integer);
begin
    FName := aName;
    FAge := aAge;
end;

procedure TPerson.Introduce;
begin
    PrintLn('Hi, I am ' + FName + ' and I am ' + IntToStr(FAge) + ' years old.');
end;

procedure TPerson.HaveBirthday;
begin
    FAge := FAge + 1;
    PrintLn(FName + ' is now ' + IntToStr(FAge) + ' years old!');
end;

// Main program
var person: TPerson;

person := TPerson.Create('Alice', 25);
person.Introduce;
person.HaveBirthday;
person.Introduce;`
    },

    math: {
        name: 'Math Operations',
        description: 'Mathematical calculations and operators',
        code: `// Math Operations in DWScript

var
    a, b: Integer;
    x, y: Float;
    sum, diff, prod: Integer;
    quotient, avg: Float;

// Integer arithmetic
a := 100;
b := 25;

sum := a + b;
diff := a - b;
prod := a * b;

PrintLn('Integer Operations:');
PrintLn('a = ' + IntToStr(a));
PrintLn('b = ' + IntToStr(b));
PrintLn('a + b = ' + IntToStr(sum));
PrintLn('a - b = ' + IntToStr(diff));
PrintLn('a * b = ' + IntToStr(prod));
PrintLn('a div b = ' + IntToStr(a div b));  // Integer division
PrintLn('a mod b = ' + IntToStr(a mod b));  // Modulo

// Float arithmetic
PrintLn('');
PrintLn('Float Operations:');
x := 10.5;
y := 3.2;

PrintLn('x = ' + FloatToStr(x));
PrintLn('y = ' + FloatToStr(y));
PrintLn('x + y = ' + FloatToStr(x + y));
PrintLn('x - y = ' + FloatToStr(x - y));
PrintLn('x * y = ' + FloatToStr(x * y));
PrintLn('x / y = ' + FloatToStr(x / y));

// Average calculation
avg := (x + y) / 2;
PrintLn('Average: ' + FloatToStr(avg));

// Compound assignments
PrintLn('');
PrintLn('Compound Assignments:');
a := 10;
PrintLn('a = ' + IntToStr(a));
a += 5;
PrintLn('a += 5 → ' + IntToStr(a));
a -= 3;
PrintLn('a -= 3 → ' + IntToStr(a));
a *= 2;
PrintLn('a *= 2 → ' + IntToStr(a));`
    }
};

// Get example by key
function getExample(key) {
    return EXAMPLES[key] || null;
}

// Get all example keys
function getExampleKeys() {
    return Object.keys(EXAMPLES);
}

// Get example list for dropdown
function getExampleList() {
    return getExampleKeys().map(key => ({
        key: key,
        name: EXAMPLES[key].name,
        description: EXAMPLES[key].description
    }));
}

// Export for use in playground
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { EXAMPLES, getExample, getExampleKeys, getExampleList };
}
