package interp

import (
	"testing"
)

// TestGroupedExpressionEvaluation tests evaluation of grouped (parenthesized) expressions.
// Task 3.5.10: VisitGroupedExpression migration - simple delegation to inner expression evaluation.
// This test verifies that precedence is preserved correctly through grouped expressions.
func TestGroupedExpressionEvaluation(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected string
	}{
		{
			name:     "grouped integer literal",
			script:   "PrintLn((42));",
			expected: "42\n",
		},
		{
			name:     "grouped string literal",
			script:   "PrintLn(('hello'));",
			expected: "hello\n",
		},
		{
			name:     "grouped boolean literal",
			script:   "PrintLn((true));",
			expected: "True\n",
		},
		{
			name: "grouped binary expression",
			script: `
var x := (10 + 20);
PrintLn(x);
`,
			expected: "30\n",
		},
		{
			name: "nested grouped expressions",
			script: `
var x := (((99)));
PrintLn(x);
`,
			expected: "99\n",
		},
		{
			name: "multiplication before addition - ungrouped",
			// 2 + 3 * 4 = 2 + 12 = 14
			script: `
var result := 2 + 3 * 4;
PrintLn(result);
`,
			expected: "14\n",
		},
		{
			name: "grouped addition before multiplication",
			// (2 + 3) * 4 = 5 * 4 = 20
			script: `
var result := (2 + 3) * 4;
PrintLn(result);
`,
			expected: "20\n",
		},
		{
			name: "complex nested grouping",
			// ((10 + 5) * 2) - (3 * 4) = (15 * 2) - 12 = 30 - 12 = 18
			script: `
var result := ((10 + 5) * 2) - (3 * 4);
PrintLn(result);
`,
			expected: "18\n",
		},
		{
			name: "grouped subtraction and division",
			// 100 div (10 - 5) = 100 div 5 = 20
			script: `
var result := 100 div (10 - 5);
PrintLn(result);
`,
			expected: "20\n",
		},
		{
			name: "grouped comparison in boolean expression",
			// (5 > 3) and (2 < 4) = true and true = true
			script: `
var result := (5 > 3) and (2 < 4);
PrintLn(result);
`,
			expected: "True\n",
		},
		{
			name: "grouped arithmetic in comparison",
			// (2 + 3) = (1 + 4) = 5 = 5 = true
			script: `
var result := (2 + 3) = (1 + 4);
PrintLn(result);
`,
			expected: "True\n",
		},
		{
			name: "precedence override with grouping",
			// Default: 10 - 5 - 2 = (10 - 5) - 2 = 3
			// With grouping: 10 - (5 - 2) = 10 - 3 = 7
			script: `
var a := 10 - 5 - 2;
var b := 10 - (5 - 2);
PrintLn(a);
PrintLn(b);
`,
			expected: "3\n7\n",
		},
		{
			name: "float operations with grouping",
			// (1.5 + 2.5) * 2.0 = 4.0 * 2.0 = 8.0
			script: `
var result := (1.5 + 2.5) * 2.0;
PrintLn(result);
`,
			expected: "8\n",
		},
		{
			name: "string concatenation with grouping",
			// ('Hello' + ' ') + 'World' with explicit grouping
			script: `
var result := ('Hello' + ' ') + 'World';
PrintLn(result);
`,
			expected: "Hello World\n",
		},
		{
			name: "boolean logic with grouping",
			// (true or false) and (false or true) = true and true = true
			script: `
var result := (true or false) and (false or true);
PrintLn(result);
`,
			expected: "True\n",
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

// TestGroupedExpressionInConditions tests grouped expressions in various conditional contexts.
func TestGroupedExpressionInConditions(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected string
	}{
		{
			name: "if statement with grouped condition",
			script: `
if (5 > 3) then
  PrintLn('yes')
else
  PrintLn('no');
`,
			expected: "yes\n",
		},
		{
			name: "while loop with grouped condition",
			script: `
var i := 0;
while (i < 3) do begin
  PrintLn(i);
  i := i + 1;
end;
`,
			expected: "0\n1\n2\n",
		},
		{
			name: "for loop with grouped expression",
			script: `
var i: Integer;
for i := (1 + 1) to (2 + 2) do
  PrintLn(i);
`,
			expected: "2\n3\n4\n",
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
