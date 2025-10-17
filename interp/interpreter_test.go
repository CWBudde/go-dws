package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/lexer"
	"github.com/cwbudde/go-dws/parser"
)

// testEval is a helper that parses and evaluates input.
func testEval(input string) Value {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		panic("parser errors: " + strings.Join(p.Errors(), "\n"))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	return interp.Eval(program)
}

// testEvalWithOutput is a helper that parses, evaluates, and captures output.
func testEvalWithOutput(input string) (Value, string) {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		panic("parser errors: " + strings.Join(p.Errors(), "\n"))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	val := interp.Eval(program)
	return val, buf.String()
}

// TestIntegerLiterals tests evaluation of integer literals.
func TestIntegerLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"0", 0},
	}

	for _, tt := range tests {
		val := testEval(tt.input)
		testIntegerValue(t, val, tt.expected)
	}
}

// TestFloatLiterals tests evaluation of float literals.
func TestFloatLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"5.0", 5.0},
		{"10.5", 10.5},
		{"-5.5", -5.5},
		{"0.0", 0.0},
	}

	for _, tt := range tests {
		val := testEval(tt.input)
		testFloatValue(t, val, tt.expected)
	}
}

// TestStringLiterals tests evaluation of string literals.
func TestStringLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`'world'`, "world"},
		{`""`, ""},
		{`"hello world"`, "hello world"},
	}

	for _, tt := range tests {
		val := testEval(tt.input)
		testStringValue(t, val, tt.expected)
	}
}

// TestBooleanLiterals tests evaluation of boolean literals.
func TestBooleanLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
	}

	for _, tt := range tests {
		val := testEval(tt.input)
		testBooleanValue(t, val, tt.expected)
	}
}

// TestIntegerArithmetic tests integer arithmetic operations.
func TestIntegerArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5", 10},
		{"5 - 3", 2},
		{"4 * 5", 20},
		{"5 + 2 * 3", 11},
		{"(5 + 2) * 3", 21},
		{"10 div 2", 5},
		{"10 mod 3", 1},
	}

	for _, tt := range tests {
		val := testEval(tt.input)
		testIntegerValue(t, val, tt.expected)
	}
}

// TestFloatArithmetic tests float arithmetic operations.
func TestFloatArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"5.0 + 2.5", 7.5},
		{"5.0 - 2.5", 2.5},
		{"2.0 * 3.0", 6.0},
		{"10.0 / 4.0", 2.5},
		{"5 + 2.5", 7.5}, // Mixed int/float
		{"10 / 4", 2.5},  // Integer division produces float with /
	}

	for _, tt := range tests {
		val := testEval(tt.input)
		testFloatValue(t, val, tt.expected)
	}
}

// TestStringConcatenation tests string concatenation.
func TestStringConcatenation(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello" + " " + "world"`, "hello world"},
		{`"foo" + "bar"`, "foobar"},
		{`"" + "test"`, "test"},
	}

	for _, tt := range tests {
		val := testEval(tt.input)
		testStringValue(t, val, tt.expected)
	}
}

// TestBooleanOperations tests boolean operations.
func TestBooleanOperations(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"true and true", true},
		{"true and false", false},
		{"false and false", false},
		{"true or false", true},
		{"false or false", false},
		{"true xor false", true},
		{"true xor true", false},
		{"not true", false},
		{"not false", true},
	}

	for _, tt := range tests {
		val := testEval(tt.input)
		testBooleanValue(t, val, tt.expected)
	}
}

// TestComparisons tests comparison operations.
func TestComparisons(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 = 1", true},
		{"1 <> 1", false},
		{"1 = 2", false},
		{"1 <> 2", true},
		{"1 <= 2", true},
		{"1 >= 1", true},
		{"2 <= 1", false},
		{`"a" < "b"`, true},
		{`"hello" = "hello"`, true},
		{`"hello" <> "world"`, true},
		{"true = true", true},
		{"true <> false", true},
	}

	for _, tt := range tests {
		val := testEval(tt.input)
		testBooleanValue(t, val, tt.expected)
	}
}

// TestVariableDeclarations tests variable declarations.
func TestVariableDeclarations(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"var x := 5; x", 5},
		{"var x := 5; var y := 10; x + y", 15},
		{"var x := 5; var y := x; y", 5},
		{"var x := 5; var y := x * 2; y", 10},
	}

	for _, tt := range tests {
		val := testEval(tt.input)
		testIntegerValue(t, val, tt.expected)
	}
}

// TestAssignments tests assignment statements.
func TestAssignments(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"var x := 0; x := 5; x", 5},
		{"var x := 5; x := x + 1; x", 6},
		{"var x := 0; var y := 10; x := y; x", 10},
		{"var x := 5; x := x * 2; x", 10},
	}

	for _, tt := range tests {
		val := testEval(tt.input)
		testIntegerValue(t, val, tt.expected)
	}
}

// TestBlockStatements tests block statement execution.
func TestBlockStatements(t *testing.T) {
	input := `
		begin
			var x := 5;
			var y := 10;
			x + y
		end
	`
	val := testEval(input)
	testIntegerValue(t, val, 15)
}

// TestBuiltinPrintLn tests the PrintLn built-in function.
func TestBuiltinPrintLn(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`PrintLn("hello")`, "hello\n"},
		{`PrintLn("hello", "world")`, "hello world\n"},
		{`PrintLn(5)`, "5\n"},
		{`PrintLn(5, 10)`, "5 10\n"},
		{`PrintLn(true)`, "true\n"},
	}

	for _, tt := range tests {
		_, output := testEvalWithOutput(tt.input)
		if output != tt.expected {
			t.Errorf("wrong output. expected=%q, got=%q", tt.expected, output)
		}
	}
}

// TestBuiltinPrint tests the Print built-in function.
func TestBuiltinPrint(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`Print("hello")`, "hello"},
		{`Print("hello", "world")`, "hello world"},
		{`Print(5); Print(10)`, "510"},
	}

	for _, tt := range tests {
		_, output := testEvalWithOutput(tt.input)
		if output != tt.expected {
			t.Errorf("wrong output. expected=%q, got=%q", tt.expected, output)
		}
	}
}

// TestCompleteProgram tests a complete program with multiple features.
func TestCompleteProgram(t *testing.T) {
	input := `
		var x := 5;
		var y := 10;
		var sum := x + y;
		PrintLn(sum);
		var product := x * y;
		PrintLn(product)
	`

	_, output := testEvalWithOutput(input)
	expected := "15\n50\n"

	if output != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, output)
	}
}

// TestUndefinedVariable tests error handling for undefined variables.
func TestUndefinedVariable(t *testing.T) {
	input := "x"
	val := testEval(input)

	if !isError(val) {
		t.Errorf("expected error, got %T (%+v)", val, val)
		return
	}

	errVal := val.(*ErrorValue)
	if !strings.Contains(errVal.Message, "undefined variable") {
		t.Errorf("wrong error message. got=%q", errVal.Message)
	}
}

// TestAssignmentToUndefinedVariable tests error handling for assignment to undefined variable.
func TestAssignmentToUndefinedVariable(t *testing.T) {
	input := "x := 5;"
	val := testEval(input)

	if !isError(val) {
		t.Errorf("expected error, got %T (%+v)", val, val)
		return
	}

	errVal := val.(*ErrorValue)
	if !strings.Contains(errVal.Message, "undefined variable") {
		t.Errorf("wrong error message. got=%q", errVal.Message)
	}
}

// TestTypeMismatch tests error handling for type mismatches.
func TestTypeMismatch(t *testing.T) {
	tests := []struct {
		input       string
		expectedErr string
	}{
		{`5 + "hello"`, "type mismatch"},
		{`"hello" - 5`, "type mismatch"},
		{`true + false`, "unknown operator"},
		{`5 and true`, "type mismatch"},
	}

	for _, tt := range tests {
		val := testEval(tt.input)

		if !isError(val) {
			t.Errorf("expected error for input %q, got %T (%+v)", tt.input, val, val)
			continue
		}

		errVal := val.(*ErrorValue)
		if !strings.Contains(errVal.Message, tt.expectedErr) {
			t.Errorf("wrong error message for %q. expected to contain %q, got=%q",
				tt.input, tt.expectedErr, errVal.Message)
		}
	}
}

// TestDivisionByZero tests error handling for division by zero.
func TestDivisionByZero(t *testing.T) {
	tests := []string{
		"5 / 0",
		"10 div 0",
		"10 mod 0",
	}

	for _, input := range tests {
		val := testEval(input)

		if !isError(val) {
			t.Errorf("expected error for input %q, got %T (%+v)", input, val, val)
			continue
		}

		errVal := val.(*ErrorValue)
		if !strings.Contains(errVal.Message, "division by zero") {
			t.Errorf("wrong error message for %q. got=%q", input, errVal.Message)
		}
	}
}

// TestCallUndefinedFunction tests error handling for calling undefined function.
func TestCallUndefinedFunction(t *testing.T) {
	input := "Foo()"
	val := testEval(input)

	if !isError(val) {
		t.Errorf("expected error, got %T (%+v)", val, val)
		return
	}

	errVal := val.(*ErrorValue)
	if !strings.Contains(errVal.Message, "undefined function") {
		t.Errorf("wrong error message. got=%q", errVal.Message)
	}
}

// Helper functions for test assertions

func testIntegerValue(t *testing.T, val Value, expected int64) bool {
	t.Helper()

	intVal, ok := val.(*IntegerValue)
	if !ok {
		t.Errorf("value is not IntegerValue. got=%T (%+v)", val, val)
		return false
	}

	if intVal.Value != expected {
		t.Errorf("intVal.Value wrong. expected=%d, got=%d", expected, intVal.Value)
		return false
	}

	return true
}

func testFloatValue(t *testing.T, val Value, expected float64) bool {
	t.Helper()

	floatVal, ok := val.(*FloatValue)
	if !ok {
		t.Errorf("value is not FloatValue. got=%T (%+v)", val, val)
		return false
	}

	if floatVal.Value != expected {
		t.Errorf("floatVal.Value wrong. expected=%f, got=%f", expected, floatVal.Value)
		return false
	}

	return true
}

func testStringValue(t *testing.T, val Value, expected string) bool {
	t.Helper()

	strVal, ok := val.(*StringValue)
	if !ok {
		t.Errorf("value is not StringValue. got=%T (%+v)", val, val)
		return false
	}

	if strVal.Value != expected {
		t.Errorf("strVal.Value wrong. expected=%q, got=%q", expected, strVal.Value)
		return false
	}

	return true
}

func testBooleanValue(t *testing.T, val Value, expected bool) bool {
	t.Helper()

	boolVal, ok := val.(*BooleanValue)
	if !ok {
		t.Errorf("value is not BooleanValue. got=%T (%+v)", val, val)
		return false
	}

	if boolVal.Value != expected {
		t.Errorf("boolVal.Value wrong. expected=%t, got=%t", expected, boolVal.Value)
		return false
	}

	return true
}

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
			expected: "1  2  3  \n2  4  6  \n3  6  9  \n",
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
