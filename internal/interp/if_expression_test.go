package interp

import (
	"testing"
)

// TestIfExpressionEvaluation tests inline if-then-else expression evaluation.
func TestIfExpressionEvaluation(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected string
	}{
		{
			name: "true branch with integers",
			script: `
var x := if True then 10 else 20;
PrintLn(x);
`,
			expected: "10\n",
		},
		{
			name: "false branch with integers",
			script: `
var x := if False then 10 else 20;
PrintLn(x);
`,
			expected: "20\n",
		},
		{
			name: "true branch with strings",
			script: `
var x := if True then 'yes' else 'no';
PrintLn(x);
`,
			expected: "yes\n",
		},
		{
			name: "false branch with strings",
			script: `
var x := if False then 'yes' else 'no';
PrintLn(x);
`,
			expected: "no\n",
		},
		{
			name: "true branch with floats",
			script: `
var x := if True then 3.14 else 2.71;
PrintLn(x);
`,
			expected: "3.14\n",
		},
		{
			name: "false branch with floats",
			script: `
var x := if False then 3.14 else 2.71;
PrintLn(x);
`,
			expected: "2.71\n",
		},
		{
			name: "true branch with booleans",
			script: `
var x := if True then True else False;
PrintLn(x);
`,
			expected: "True\n",
		},
		{
			name: "false branch with booleans",
			script: `
var x := if False then True else False;
PrintLn(x);
`,
			expected: "False\n",
		},
		{
			name: "condition from comparison",
			script: `
var x := if 5 > 3 then 'greater' else 'not greater';
PrintLn(x);
`,
			expected: "greater\n",
		},
		{
			name: "condition from variable",
			script: `
var cond := True;
var x := if cond then 100 else 200;
PrintLn(x);
`,
			expected: "100\n",
		},
		{
			name: "nested if expressions",
			script: `
var x := if True then (if False then 1 else 2) else 3;
PrintLn(x);
`,
			expected: "2\n",
		},
		{
			name: "if expression with arithmetic",
			script: `
var x := if True then (10 + 5) else (10 - 5);
PrintLn(x);
`,
			expected: "15\n",
		},
		{
			name: "if expression in arithmetic",
			script: `
var x := (if True then 10 else 20) + 5;
PrintLn(x);
`,
			expected: "15\n",
		},
		{
			name: "complex condition with and",
			script: `
var x := if (5 > 3) and (2 < 4) then 'yes' else 'no';
PrintLn(x);
`,
			expected: "yes\n",
		},
		{
			name: "complex condition with or",
			script: `
var x := if (5 < 3) or (2 < 4) then 'yes' else 'no';
PrintLn(x);
`,
			expected: "yes\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.script)

			if output != tt.expected {
				t.Errorf("wrong output:\n  expected: %q\n  got:      %q", tt.expected, output)
			}
		})
	}
}

// TestIfExpressionShortCircuit verifies that only the chosen branch is evaluated.
func TestIfExpressionShortCircuit(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected string
	}{
		{
			name: "true branch - else branch not evaluated",
			script: `
var counter := 0;

function IncrementAndReturn(n: Integer): Integer;
begin
  counter := counter + 1;
  Result := n;
end;

// Only the then branch should be evaluated
var x := if True then IncrementAndReturn(10) else IncrementAndReturn(20);
PrintLn(x);
PrintLn(counter); // Should be 1, not 2
`,
			expected: "10\n1\n",
		},
		{
			name: "false branch - then branch not evaluated",
			script: `
var counter := 0;

function IncrementAndReturn(n: Integer): Integer;
begin
  counter := counter + 1;
  Result := n;
end;

// Only the else branch should be evaluated
var x := if False then IncrementAndReturn(10) else IncrementAndReturn(20);
PrintLn(x);
PrintLn(counter); // Should be 1, not 2
`,
			expected: "20\n1\n",
		},
		{
			name: "true branch - error in else not triggered",
			script: `
function WillError: Integer;
begin
  raise Exception.Create('Should not be called');
  Result := 99;
end;

// The else branch should not be evaluated, so no error
var x := if True then 42 else WillError;
PrintLn(x);
`,
			expected: "42\n",
		},
		{
			name: "false branch - error in then not triggered",
			script: `
function WillError: Integer;
begin
  raise Exception.Create('Should not be called');
  Result := 99;
end;

// The then branch should not be evaluated, so no error
var x := if False then WillError else 42;
PrintLn(x);
`,
			expected: "42\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.script)

			if output != tt.expected {
				t.Errorf("wrong output:\n  expected: %q\n  got:      %q", tt.expected, output)
			}
		})
	}
}

// TestIfExpressionWithBooleanCoercion tests boolean coercion in conditions.
// Note: DWScript inline if expressions require boolean conditions.
// The condition is automatically coerced to boolean for certain types.
func TestIfExpressionWithBooleanCoercion(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected string
	}{
		{
			name: "explicit boolean true",
			script: `
var x := if True then 'yes' else 'no';
PrintLn(x);
`,
			expected: "yes\n",
		},
		{
			name: "explicit boolean false",
			script: `
var x := if False then 'yes' else 'no';
PrintLn(x);
`,
			expected: "no\n",
		},
		{
			name: "comparison result",
			script: `
var x := if (10 > 5) then 'yes' else 'no';
PrintLn(x);
`,
			expected: "yes\n",
		},
		{
			name: "equality check",
			script: `
var x := if (5 = 5) then 'yes' else 'no';
PrintLn(x);
`,
			expected: "yes\n",
		},
		{
			name: "inequality check",
			script: `
var x := if (5 <> 3) then 'yes' else 'no';
PrintLn(x);
`,
			expected: "yes\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.script)

			if output != tt.expected {
				t.Errorf("wrong output:\n  expected: %q\n  got:      %q", tt.expected, output)
			}
		})
	}
}

// TestIfExpressionInComplexContexts tests if expressions in various contexts.
func TestIfExpressionInComplexContexts(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected string
	}{
		{
			name: "if expression in function call",
			script: `
function Double(n: Integer): Integer;
begin
  Result := n * 2;
end;

var x := Double(if True then 5 else 10);
PrintLn(x);
`,
			expected: "10\n",
		},
		{
			name: "if expression in array indexing",
			script: `
var arr: array of Integer;
SetLength(arr, 3);
arr[0] := 100;
arr[1] := 200;
arr[2] := 300;

var x := arr[if True then 1 else 2];
PrintLn(x);
`,
			expected: "200\n",
		},
		{
			name: "if expression as condition in if statement",
			script: `
if (if True then True else False) then
  PrintLn('yes')
else
  PrintLn('no');
`,
			expected: "yes\n",
		},
		{
			name: "multiple if expressions",
			script: `
var a := if True then 1 else 2;
var b := if False then 3 else 4;
var c := a + b;
PrintLn(c);
`,
			expected: "5\n",
		},
		{
			name: "if expression with string concatenation",
			script: `
var greeting := 'Hello, ' + (if True then 'World' else 'Universe');
PrintLn(greeting);
`,
			expected: "Hello, World\n",
		},
		{
			name: "deeply nested if expressions",
			script: `
var x := if True then
           (if True then
             (if False then 1 else 2)
           else 3)
         else 4;
PrintLn(x);
`,
			expected: "2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.script)

			if output != tt.expected {
				t.Errorf("wrong output:\n  expected: %q\n  got:      %q", tt.expected, output)
			}
		})
	}
}

// TestIfExpressionEdgeCases tests edge cases and error conditions.
func TestIfExpressionEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected string
	}{
		{
			name: "if expression with variable condition",
			script: `
var cond := False;
var x := if cond then 'a' else 'b';
PrintLn(x);
`,
			expected: "b\n",
		},
		{
			name: "if expression with negated condition",
			script: `
var x := if not True then 'a' else 'b';
PrintLn(x);
`,
			expected: "b\n",
		},
		{
			name: "if expression with complex boolean expression",
			script: `
var a := True;
var b := False;
var x := if (a and not b) or (b and not a) then 'yes' else 'no';
PrintLn(x);
`,
			expected: "yes\n",
		},
		{
			name: "if expression returning same type as condition",
			script: `
var x := if True then False else True;
PrintLn(x);
`,
			expected: "False\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.script)

			if output != tt.expected {
				t.Errorf("wrong output:\n  expected: %q\n  got:      %q", tt.expected, output)
			}
		})
	}
}
