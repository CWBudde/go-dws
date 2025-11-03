package interp

import (
	"testing"
)

// ============================================================================
// TDD Tests for Dynamic Array Features (happy_numbers.pas requirements)
// ============================================================================

// TestDynamicArray_AddMethod tests the .Add() method on dynamic arrays
// This is critical for happy_numbers.pas which uses: cache.Add(sum)
func TestDynamicArray_AddMethod(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected string
	}{
		{
			name: "Add single integer to empty array",
			script: `
var cache : array of Integer;
cache.Add(42);
PrintLn(Length(cache));
PrintLn(cache[0]);
`,
			expected: "1\n42\n",
		},
		{
			name: "Add multiple integers to array",
			script: `
var cache : array of Integer;
cache.Add(10);
cache.Add(20);
cache.Add(30);
PrintLn(Length(cache));
PrintLn(cache[0]);
PrintLn(cache[1]);
PrintLn(cache[2]);
`,
			expected: "3\n10\n20\n30\n",
		},
		{
			name: "Add in a loop",
			script: `
var arr : array of Integer;
var i : Integer;
for i := 1 to 5 do
   arr.Add(i * 10);

PrintLn(Length(arr));
for i := 0 to Length(arr) - 1 do
   PrintLn(arr[i]);
`,
			expected: "5\n10\n20\n30\n40\n50\n",
		},
		{
			name: "Add strings to array",
			script: `
var names : array of String;
names.Add('Alice');
names.Add('Bob');
names.Add('Charlie');
PrintLn(Length(names));
PrintLn(names[0]);
PrintLn(names[1]);
PrintLn(names[2]);
`,
			expected: "3\nAlice\nBob\nCharlie\n",
		},
		{
			name: "Add to array inside function",
			script: `
function BuildArray() : array of Integer;
var arr : array of Integer;
begin
   arr.Add(1);
   arr.Add(2);
   arr.Add(3);
   Result := arr;
end;

var myArr := BuildArray();
PrintLn(Length(myArr));
PrintLn(myArr[0]);
PrintLn(myArr[1]);
PrintLn(myArr[2]);
`,
			expected: "3\n1\n2\n3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output := testEvalWithOutput(tt.script)
			if result != nil && result.Type() == "ERROR" {
				t.Fatalf("unexpected error: %s", result.String())
			}
			if output != tt.expected {
				t.Errorf("expected output:\n%s\ngot:\n%s", tt.expected, output)
			}
		})
	}
}

// TestDynamicArray_InOperator tests the 'in' operator for checking array membership
// This is critical for happy_numbers.pas which uses: if sum in cache then
func TestDynamicArray_InOperator(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected string
	}{
		{
			name: "Check if value exists in array (true case)",
			script: `
var cache : array of Integer;
cache.Add(1);
cache.Add(4);
cache.Add(16);
cache.Add(37);

if 16 in cache then
   PrintLn('found')
else
   PrintLn('not found');
`,
			expected: "found\n",
		},
		{
			name: "Check if value does not exist in array (false case)",
			script: `
var cache : array of Integer;
cache.Add(1);
cache.Add(4);
cache.Add(16);
cache.Add(37);

if 99 in cache then
   PrintLn('found')
else
   PrintLn('not found');
`,
			expected: "not found\n",
		},
		{
			name: "Check in empty array",
			script: `
var cache : array of Integer;

if 42 in cache then
   PrintLn('found')
else
   PrintLn('not found');
`,
			expected: "not found\n",
		},
		{
			name: "Multiple in checks",
			script: `
var cache : array of Integer;
cache.Add(10);
cache.Add(20);
cache.Add(30);

if 10 in cache then PrintLn('10 found');
if 20 in cache then PrintLn('20 found');
if 30 in cache then PrintLn('30 found');
if 40 in cache then PrintLn('40 found');
if not (40 in cache) then PrintLn('40 not found');
`,
			expected: "10 found\n20 found\n30 found\n40 not found\n",
		},
		{
			name: "String array membership",
			script: `
var names : array of String;
names.Add('Alice');
names.Add('Bob');
names.Add('Charlie');

if 'Bob' in names then
   PrintLn('Bob found')
else
   PrintLn('Bob not found');

if 'Dave' in names then
   PrintLn('Dave found')
else
   PrintLn('Dave not found');
`,
			expected: "Bob found\nDave not found\n",
		},
		{
			name: "In operator within while loop (happy_numbers pattern)",
			script: `
var cache : array of Integer;
var n := 1;
var count := 0;

while count < 5 do begin
   if not (n in cache) then begin
      cache.Add(n);
      PrintLn(n);
      Inc(count);
   end;
   Inc(n);
end;
`,
			expected: "1\n2\n3\n4\n5\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output := testEvalWithOutput(tt.script)
			if result != nil && result.Type() == "ERROR" {
				t.Fatalf("unexpected error: %s", result.String())
			}
			if output != tt.expected {
				t.Errorf("expected output:\n%s\ngot:\n%s", tt.expected, output)
			}
		})
	}
}

// TestDynamicArray_LengthProperty tests the .Length property on dynamic arrays
// This verifies that Length() works correctly with dynamically sized arrays
func TestDynamicArray_LengthProperty(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected string
	}{
		{
			name: "Length of empty array",
			script: `
var arr : array of Integer;
PrintLn(Length(arr));
`,
			expected: "0\n",
		},
		{
			name: "Length increases with Add",
			script: `
var arr : array of Integer;
PrintLn(Length(arr));
arr.Add(10);
PrintLn(Length(arr));
arr.Add(20);
PrintLn(Length(arr));
arr.Add(30);
PrintLn(Length(arr));
`,
			expected: "0\n1\n2\n3\n",
		},
		{
			name: "Length used in loop condition",
			script: `
var arr : array of Integer;
var i : Integer;

for i := 1 to 5 do
   arr.Add(i);

for i := 0 to Length(arr) - 1 do
   PrintLn(arr[i]);
`,
			expected: "1\n2\n3\n4\n5\n",
		},
		{
			name: "Length in arithmetic expression",
			script: `
var arr : array of Integer;
arr.Add(10);
arr.Add(20);
arr.Add(30);

PrintLn(Length(arr) * 2);
PrintLn(Length(arr) + 10);
`,
			expected: "6\n13\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output := testEvalWithOutput(tt.script)
			if result != nil && result.Type() == "ERROR" {
				t.Fatalf("unexpected error: %s", result.String())
			}
			if output != tt.expected {
				t.Errorf("expected output:\n%s\ngot:\n%s", tt.expected, output)
			}
		})
	}
}

// TestHappyNumbers_SimplifiedVersion tests a simplified version of the happy numbers algorithm
// This is the actual pattern from happy_numbers.pas
func TestHappyNumbers_SimplifiedVersion(t *testing.T) {
	script := `
function IsHappy(n : Integer) : Boolean;
var
   cache : array of Integer;
   sum : Integer;
   iterCount : Integer;
begin
   iterCount := 0;
   while iterCount < 100 do begin  // Safety limit instead of infinite loop
      sum := 0;
      while n>0 do begin
         sum += (n mod 10) * (n mod 10);  // Sqr not yet available
         n := n div 10;
      end;
      if sum = 1 then
         Exit(True);
      if sum in cache then
         Exit(False);
      n := sum;
      cache.Add(sum);
      Inc(iterCount);
   end;
   Exit(False);  // Safety: assume not happy after 100 iterations
end;

// Test first few happy numbers: 1, 7, 10, 13
if IsHappy(1) then PrintLn('1 is happy');
if IsHappy(7) then PrintLn('7 is happy');
if IsHappy(10) then PrintLn('10 is happy');
if not IsHappy(2) then PrintLn('2 is not happy');
if not IsHappy(3) then PrintLn('3 is not happy');
`

	expected := "1 is happy\n7 is happy\n10 is happy\n2 is not happy\n3 is not happy\n"

	result, output := testEvalWithOutput(script)
	if result != nil && result.Type() == "ERROR" {
		t.Fatalf("unexpected error: %s", result.String())
	}
	if output != expected {
		t.Errorf("expected output:\n%s\ngot:\n%s", expected, output)
	}
}
