package interp

import (
	"testing"
)

// TestWhileStatementExecution tests while loop execution.
func TestWhileStatementExecution(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Simple while loop - count from 0 to 5
		{
			`var x := 0; while x < 5 do begin x := x + 1; PrintLn(x) end`,
			"1\n2\n3\n4\n5\n",
		},
		// While loop with single statement body
		{
			`var x := 0; while x < 3 do x := x + 1; PrintLn(x)`,
			"3\n",
		},
		// While loop that doesn't execute (condition false from start)
		{
			`var x := 10; while x < 5 do PrintLn("should not print")`,
			"",
		},
		// While loop with complex condition
		{
			`var x := 0; var y := 0; while x < 3 and y < 2 do begin x := x + 1; y := y + 1; PrintLn(x) end`,
			"1\n2\n",
		},
		// Sum numbers with while loop
		{
			`var sum := 0; var i := 1; while i <= 5 do begin sum := sum + i; i := i + 1 end; PrintLn(sum)`,
			"15\n",
		},
		// While loop with boolean variable
		{
			`var running := true; var count := 0; while running do begin count := count + 1; if count >= 3 then running := false end; PrintLn(count)`,
			"3\n",
		},
	}

	for _, tt := range tests {
		_, output := testEvalWithOutput(tt.input)
		if output != tt.expected {
			t.Errorf("wrong output for %q.\nexpected=%q\ngot=%q", tt.input, tt.expected, output)
		}
	}
}

// TestIfStatementExecution tests if statement execution.
func TestIfStatementExecution(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Simple if with true condition - consequence should execute
		{
			`if true then PrintLn("consequence")`,
			"consequence\n",
		},
		// Simple if with false condition - consequence should NOT execute
		{
			`if false then PrintLn("consequence")`,
			"",
		},
		// If with comparison - true case
		{
			`var x := 10; if x > 5 then PrintLn("x is greater")`,
			"x is greater\n",
		},
		// If with comparison - false case
		{
			`var x := 3; if x > 5 then PrintLn("x is greater")`,
			"",
		},
		// If-else with true condition - consequence executes
		{
			`if true then PrintLn("consequence") else PrintLn("alternative")`,
			"consequence\n",
		},
		// If-else with false condition - alternative executes
		{
			`if false then PrintLn("consequence") else PrintLn("alternative")`,
			"alternative\n",
		},
		// If-else with expression condition
		{
			`var x := 10; if x > 5 then PrintLn("greater") else PrintLn("not greater")`,
			"greater\n",
		},
		// If with block statement
		{
			`if true then begin PrintLn("line1"); PrintLn("line2") end`,
			"line1\nline2\n",
		},
		// Nested if statements
		{
			`if true then if true then PrintLn("nested")`,
			"nested\n",
		},
		// If with assignment in consequence
		{
			`var x := 0; if true then x := 10; PrintLn(x)`,
			"10\n",
		},
	}

	for _, tt := range tests {
		_, output := testEvalWithOutput(tt.input)
		if output != tt.expected {
			t.Errorf("wrong output for %q.\nexpected=%q\ngot=%q", tt.input, tt.expected, output)
		}
	}
}

// TestRepeatStatementExecution tests repeat-until loop execution.
func TestRepeatStatementExecution(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Simple repeat-until - count from 1 to 5
		{
			`var x := 0; repeat begin x := x + 1; PrintLn(x) end until x >= 5`,
			"1\n2\n3\n4\n5\n",
		},
		// Repeat-until with single statement body
		{
			`var x := 0; repeat x := x + 1 until x >= 3; PrintLn(x)`,
			"3\n",
		},
		// Repeat-until that executes only once (condition true immediately)
		{
			`var x := 10; repeat PrintLn(x) until x >= 5`,
			"10\n",
		},
		// Repeat-until with complex condition
		{
			`var x := 0; var y := 0; repeat begin x := x + 1; y := y + 1; PrintLn(x) end until x >= 3 or y >= 3`,
			"1\n2\n3\n",
		},
		// Sum numbers with repeat-until loop
		{
			`var sum := 0; var i := 0; repeat begin i := i + 1; sum := sum + i end until i >= 5; PrintLn(sum)`,
			"15\n",
		},
		// Repeat-until with boolean variable control
		{
			`var done := false; var count := 0; repeat begin count := count + 1; if count >= 3 then done := true end until done; PrintLn(count)`,
			"3\n",
		},
		// Repeat-until always executes at least once even if condition is initially true
		{
			`var x := 100; repeat PrintLn("executed") until x >= 5`,
			"executed\n",
		},
		// Nested repeat-until loops
		{
			`var i := 0; repeat begin i := i + 1; var j := 0; repeat begin j := j + 1; Print(j) end until j >= 2; PrintLn("") end until i >= 2`,
			"12\n12\n",
		},
		// Repeat-until with multiple statements in body
		{
			`var x := 0; repeat begin x := x + 1; PrintLn("iteration"); PrintLn(x) end until x >= 2`,
			"iteration\n1\niteration\n2\n",
		},
	}

	for _, tt := range tests {
		_, output := testEvalWithOutput(tt.input)
		if output != tt.expected {
			t.Errorf("wrong output for %q.\nexpected=%q\ngot=%q", tt.input, tt.expected, output)
		}
	}
}

// TestForStatementExecution tests for loop execution.
func TestForStatementExecution(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple ascending for loop",
			input:    `for i := 1 to 5 do PrintLn(i)`,
			expected: "1\n2\n3\n4\n5\n",
		},
		{
			name:     "Ascending for loop with block body",
			input:    `for i := 1 to 3 do begin PrintLn("iteration"); PrintLn(i) end`,
			expected: "iteration\n1\niteration\n2\niteration\n3\n",
		},
		{
			name:     "Descending for loop",
			input:    `for i := 5 downto 1 do PrintLn(i)`,
			expected: "5\n4\n3\n2\n1\n",
		},
		{
			name:     "For loop with variable bounds",
			input:    `var start := 1; var finish := 3; for i := start to finish do PrintLn(i)`,
			expected: "1\n2\n3\n",
		},
		{
			name:     "For loop that doesn't execute (to)",
			input:    `for i := 5 to 3 do PrintLn(i)`,
			expected: "",
		},
		{
			name:     "For loop that doesn't execute (downto)",
			input:    `for i := 3 downto 5 do PrintLn(i)`,
			expected: "",
		},
		{
			name:     "For loop with single iteration",
			input:    `for i := 10 to 10 do PrintLn(i)`,
			expected: "10\n",
		},
		{
			name:     "Sum using for loop",
			input:    `var sum := 0; for i := 1 to 5 do sum := sum + i; PrintLn(sum)`,
			expected: "15\n",
		},
		{
			name:     "Factorial using for loop",
			input:    `var fact := 1; for i := 1 to 5 do fact := fact * i; PrintLn(fact)`,
			expected: "120\n",
		},
		{
			name:     "Nested for loops",
			input:    `for i := 1 to 3 do begin for j := 1 to 3 do Print(i * j, " "); PrintLn("") end`,
			expected: "1 2 3 \n2 4 6 \n3 6 9 \n",
		},
		{
			name:     "For loop variable scoping",
			input:    `var i := 99; for i := 1 to 3 do PrintLn(i); PrintLn(i)`,
			expected: "1\n2\n3\n99\n",
		},
		{
			name:     "For loop with expression bounds (to)",
			input:    `for i := 2 + 3 to 10 - 2 do PrintLn(i)`,
			expected: "5\n6\n7\n8\n",
		},
		{
			name:     "For loop with expression bounds (downto)",
			input:    `for i := 3 * 2 downto 10 div 5 do PrintLn(i)`,
			expected: "6\n5\n4\n3\n2\n",
		},
		{
			name:     "For loop accessing outer variables",
			input:    `var multiplier := 2; for i := 1 to 4 do PrintLn(i * multiplier)`,
			expected: "2\n4\n6\n8\n",
		},
		{
			name:     "For loop with assignment in body",
			input:    `var count := 0; for i := 1 to 5 do begin count := count + 1; PrintLn(count) end`,
			expected: "1\n2\n3\n4\n5\n",
		},
		{
			name: "For loop with declared iterator",
			input: `
				function Test;
				var
					i: Integer;
				begin
					for i := 0 to 10 do
						PrintLn(i);
				end;

				Test;
			`,
			expected: "0\n1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n",
		},
		{
			name:     "For loop with inline var declaration",
			input:    `for var i := 0 to 10 do PrintLn(i)`,
			expected: "0\n1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.input)
			if output != tt.expected {
				t.Errorf("wrong output.\nexpected=%q\ngot=%q", tt.expected, output)
			}
		})
	}
}

// TestForStatementWithStep tests for loop execution with step keyword.
func TestForStatementWithStep(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Ascending for loop with step 2",
			input:    `for i := 1 to 5 step 2 do PrintLn(i)`,
			expected: "1\n3\n5\n",
		},
		{
			name:     "Descending for loop with step 3",
			input:    `for i := 10 downto 1 step 3 do PrintLn(i)`,
			expected: "10\n7\n4\n1\n",
		},
		{
			name:     "For loop with step expression",
			input:    `var s := 2; for i := 0 to 10 step (s + 1) do PrintLn(i)`,
			expected: "0\n3\n6\n9\n",
		},
		{
			name:     "For loop with step variable",
			input:    `var stepSize := 5; for i := 0 to 20 step stepSize do PrintLn(i)`,
			expected: "0\n5\n10\n15\n20\n",
		},
		{
			name:     "Step larger than range",
			input:    `for i := 1 to 3 step 10 do PrintLn(i)`,
			expected: "1\n",
		},
		{
			name:     "Large step ascending",
			input:    `for i := 1 to 100 step 50 do PrintLn(i)`,
			expected: "1\n51\n",
		},
		{
			name:     "Sum using for loop with step",
			input:    `var sum := 0; for i := 2 to 10 step 2 do sum := sum + i; PrintLn(sum)`,
			expected: "30\n",
		},
		{
			name:     "For loop with step and inline var",
			input:    `for var i := 0 to 10 step 3 do PrintLn(i)`,
			expected: "0\n3\n6\n9\n",
		},
		{
			name:     "For loop without step still works (backward compatibility)",
			input:    `for i := 1 to 5 do PrintLn(i)`,
			expected: "1\n2\n3\n4\n5\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.input)
			if output != tt.expected {
				t.Errorf("wrong output.\nexpected=%q\ngot=%q", tt.expected, output)
			}
		})
	}
}

// TestForStatementStepErrors tests error handling for invalid step values.
func TestForStatementStepErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name:          "Step value zero",
			input:         `for i := 1 to 5 step 0 do PrintLn(i)`,
			expectedError: "FOR loop STEP should be strictly positive: 0",
		},
		{
			name:          "Step value negative",
			input:         `var s := -1; for i := 1 to 5 step s do PrintLn(i)`,
			expectedError: "FOR loop STEP should be strictly positive: -1",
		},
		{
			name:          "Step value negative literal",
			input:         `for i := 1 to 5 step -2 do PrintLn(i)`,
			expectedError: "FOR loop STEP should be strictly positive: -2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := testEvalWithOutput(tt.input)
			if result == nil {
				t.Fatalf("expected error, got nil")
			}
			errVal, ok := result.(*ErrorValue)
			if !ok {
				t.Fatalf("expected ErrorValue, got %T", result)
			}
			if errVal.Message != tt.expectedError {
				t.Errorf("wrong error message.\nexpected=%q\ngot=%q", tt.expectedError, errVal.Message)
			}
		})
	}
}

// TestCaseStatementExecution tests case statement execution.
func TestCaseStatementExecution(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Simple case with integer values",
			input: `
				var x := 2;
				case x of
					1: PrintLn("one");
					2: PrintLn("two");
					3: PrintLn("three");
				end
			`,
			expected: "two\n",
		},
		{
			name: "Case with multiple values per branch",
			input: `
				var x := 3;
				case x of
					1, 2: PrintLn("one or two");
					3, 4: PrintLn("three or four");
					5: PrintLn("five");
				end
			`,
			expected: "three or four\n",
		},
		{
			name: "Case with else branch - no match",
			input: `
				var x := 10;
				case x of
					1: PrintLn("one");
					2: PrintLn("two");
				else
					PrintLn("other");
				end
			`,
			expected: "other\n",
		},
		{
			name: "Case with else branch - has match",
			input: `
				var x := 1;
				case x of
					1: PrintLn("one");
					2: PrintLn("two");
				else
					PrintLn("other");
				end
			`,
			expected: "one\n",
		},
		{
			name: "Case with string values",
			input: `
				var name := "bob";
				case name of
					"alice": PrintLn("Hello Alice");
					"bob": PrintLn("Hello Bob");
					"charlie": PrintLn("Hello Charlie");
				end
			`,
			expected: "Hello Bob\n",
		},
		{
			name: "Case with no match and no else",
			input: `
				var x := 99;
				case x of
					1: PrintLn("one");
					2: PrintLn("two");
				end
			`,
			expected: "",
		},
		{
			name: "Case with block statements",
			input: `
				var x := 2;
				case x of
					1: begin PrintLn("one"); PrintLn("first") end;
					2: begin PrintLn("two"); PrintLn("second") end;
					3: PrintLn("three");
				end
			`,
			expected: "two\nsecond\n",
		},
		{
			name: "Case with expression as case value",
			input: `
				var x := 5;
				var y := 2;
				case x + y of
					5: PrintLn("five");
					7: PrintLn("seven");
					10: PrintLn("ten");
				end
			`,
			expected: "seven\n",
		},
		{
			name: "Case with variable assignment in branch",
			input: `
				var x := 2;
				var result := 0;
				case x of
					1: result := 10;
					2: result := 20;
					3: result := 30;
				end;
				PrintLn(result)
			`,
			expected: "20\n",
		},
		{
			name: "Case with boolean values",
			input: `
				var flag := true;
				case flag of
					true: PrintLn("yes");
					false: PrintLn("no");
				end
			`,
			expected: "yes\n",
		},
		{
			name: "Case with expression values in branches",
			input: `
				var x := 10;
				var limit := 10;
				case x of
					5 + 5: PrintLn("matched");
					15: PrintLn("fifteen");
				else
					PrintLn("no match");
				end
			`,
			expected: "matched\n",
		},
		{
			name: "Nested case statements",
			input: `
				var x := 1;
				var y := 2;
				case x of
					1: case y of
						1: PrintLn("1-1");
						2: PrintLn("1-2");
					end;
					2: PrintLn("x is 2");
				end
			`,
			expected: "1-2\n",
		},
		{
			name: "Case with first matching branch executes (not all)",
			input: `
				var x := 2;
				case x of
					2: PrintLn("first");
					2: PrintLn("second");
					2: PrintLn("third");
				end
			`,
			expected: "first\n",
		},
		{
			name: "Case inside loop",
			input: `
				for i := 1 to 3 do
					case i of
						1: PrintLn("one");
						2: PrintLn("two");
						3: PrintLn("three");
					end
			`,
			expected: "one\ntwo\nthree\n",
		},
		{
			name: "Case with negative integers",
			input: `
				var x := -1;
				case x of
					-2: PrintLn("minus two");
					-1: PrintLn("minus one");
					0: PrintLn("zero");
					1: PrintLn("one");
				end
			`,
			expected: "minus one\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.input)
			if output != tt.expected {
				t.Errorf("wrong output.\nexpected=%q\ngot=%q", tt.expected, output)
			}
		})
	}
}

// TestCaseStatementWithRanges tests case statement execution with range expressions.
func TestCaseStatementWithRanges(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Character range uppercase",
			input: `
				var ch := 'B';
				case ch of
					'A'..'Z': PrintLn('UPPER');
					'a'..'z': PrintLn('lower');
				else
					PrintLn('other');
				end
			`,
			expected: "UPPER\n",
		},
		{
			name: "Character range lowercase",
			input: `
				var ch := 'g';
				case ch of
					'A'..'Z': PrintLn('UPPER');
					'a'..'z': PrintLn('lower');
				else
					PrintLn('other');
				end
			`,
			expected: "lower\n",
		},
		{
			name: "Character range no match",
			input: `
				var ch := '5';
				case ch of
					'A'..'Z': PrintLn('UPPER');
					'a'..'z': PrintLn('lower');
				else
					PrintLn('other');
				end
			`,
			expected: "other\n",
		},
		{
			name: "Integer range",
			input: `
				var x := 5;
				case x of
					1..10: PrintLn('1-10');
					11..20: PrintLn('11-20');
				else
					PrintLn('other');
				end
			`,
			expected: "1-10\n",
		},
		{
			name: "Integer range boundary start",
			input: `
				var x := 1;
				case x of
					1..10: PrintLn('1-10');
					11..20: PrintLn('11-20');
				end
			`,
			expected: "1-10\n",
		},
		{
			name: "Integer range boundary end",
			input: `
				var x := 10;
				case x of
					1..10: PrintLn('1-10');
					11..20: PrintLn('11-20');
				end
			`,
			expected: "1-10\n",
		},
		{
			name: "Integer range second branch",
			input: `
				var x := 15;
				case x of
					1..10: PrintLn('1-10');
					11..20: PrintLn('11-20');
				end
			`,
			expected: "11-20\n",
		},
		{
			name: "Integer range no match",
			input: `
				var x := 25;
				case x of
					1..10: PrintLn('1-10');
					11..20: PrintLn('11-20');
				else
					PrintLn('other');
				end
			`,
			expected: "other\n",
		},
		{
			name: "Mixed ranges and single values",
			input: `
				var x := 4;
				case x of
					1, 3..5, 7: PrintLn('match');
					2, 6: PrintLn('even');
				else
					PrintLn('other');
				end
			`,
			expected: "match\n",
		},
		{
			name: "Mixed - single value before range",
			input: `
				var x := 1;
				case x of
					1, 3..5, 7: PrintLn('match');
					2, 6: PrintLn('even');
				end
			`,
			expected: "match\n",
		},
		{
			name: "Mixed - single value after range",
			input: `
				var x := 7;
				case x of
					1, 3..5, 7: PrintLn('match');
					2, 6: PrintLn('even');
				end
			`,
			expected: "match\n",
		},
		{
			name: "Mixed - second branch single value",
			input: `
				var x := 2;
				case x of
					1, 3..5, 7: PrintLn('match');
					2, 6: PrintLn('even');
				end
			`,
			expected: "even\n",
		},
		{
			name: "Float range",
			input: `
				var x := 5.5;
				case x of
					1.0..10.0: PrintLn('1-10');
					10.1..20.0: PrintLn('10-20');
				else
					PrintLn('other');
				end
			`,
			expected: "1-10\n",
		},
		{
			name: "Multiple ranges",
			input: `
				var x := 15;
				case x of
					0..9: PrintLn('single digit');
					10..99: PrintLn('double digit');
					100..999: PrintLn('triple digit');
				end
			`,
			expected: "double digit\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.input)
			if output != tt.expected {
				t.Errorf("wrong output.\nexpected=%q\ngot=%q", tt.expected, output)
			}
		})
	}
}
