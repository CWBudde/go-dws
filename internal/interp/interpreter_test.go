package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
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

// TestCharLiterals tests evaluation of character literals.
func TestCharLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"#65", "A"},   // Decimal: A
		{"#$41", "A"},  // Hex: A
		{"#13", "\r"},  // Carriage return
		{"#10", "\n"},  // Line feed
		{"#$61", "a"},  // Hex: a
		{"#32", " "},   // Space
		{"#$0D", "\r"}, // Hex CR
		{"#$0A", "\n"}, // Hex LF
		{"#48", "0"},   // Digit 0
		{"#$30", "0"},  // Hex digit 0
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			val := testEval(tt.input)
			testStringValue(t, val, tt.expected)
		})
	}
}

// TestCharLiteralConcatenation tests character literal concatenation with strings.
func TestCharLiteralConcatenation(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`'Hello' + #65`, "HelloA"},
		{`#65 + 'Hello'`, "AHello"},
		{`#13 + #10`, "\r\n"},
		{`'Line1' + #13 + #10 + 'Line2'`, "Line1\r\nLine2"},
		{`#72 + #101 + #108 + #108 + #111`, "Hello"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			val := testEval(tt.input)
			testStringValue(t, val, tt.expected)
		})
	}
}

// TestCharLiteralInVariable tests character literal assignment to variables.
func TestCharLiteralInVariable(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"var s: String := #65; s", "A"},
		{"var c := #$41; c", "A"},
		{"var cr := #13; var lf := #10; cr + lf", "\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			val := testEval(tt.input)
			testStringValue(t, val, tt.expected)
		})
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
		// Bitwise shift operators
		{"2 shl 3", 16},
		{"16 shr 2", 4},
		{"1 shl 10", 1024},
		{"1024 shr 10", 1},
		{"8 shl 0", 8},
		{"8 shr 0", 8},
		// Bitwise logical operators
		{"5 and 3", 1},   // 101 & 011 = 001
		{"5 or 3", 7},    // 101 | 011 = 111
		{"5 xor 3", 6},   // 101 ^ 011 = 110
		{"12 and 10", 8}, // 1100 & 1010 = 1000
		{"12 or 10", 14}, // 1100 | 1010 = 1110
		{"12 xor 10", 6}, // 1100 ^ 1010 = 0110
		// Complex bitwise expressions
		{"(2 shl 1) or 1", 5}, // (2 << 1) | 1 = 4 | 1 = 5
		{"2 + 3 shl 2", 14},   // 2 + (3 << 2) = 2 + 12 = 14
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
		{`PrintLn("hello", "world")`, "helloworld\n"},
		{`PrintLn(5)`, "5\n"},
		{`PrintLn(5, 10)`, "510\n"},
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
		{`Print("hello", "world")`, "helloworld"},
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

		// Handle both RuntimeError and ErrorValue
		var errMsg string
		switch err := val.(type) {
		case *RuntimeError:
			errMsg = err.Message
		case *ErrorValue:
			errMsg = err.Message
		default:
			t.Errorf("unexpected error type: %T", val)
			continue
		}

		// Check for division or modulo by zero
		errMsgLower := strings.ToLower(errMsg)
		if !strings.Contains(errMsgLower, "by zero") {
			t.Errorf("wrong error message for %q. got=%q", input, errMsg)
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

// TestFunctionCalls tests user-defined function calls.
func TestFunctionCalls(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Simple function returning integer",
			input: `
				function Add(a, b: Integer): Integer;
				begin
					Result := a + b;
				end;

				PrintLn(Add(2, 3))
			`,
			expected: "5\n",
		},
		{
			name: "Function with single parameter",
			input: `
				function Double(x: Integer): Integer;
				begin
					Result := x * 2;
				end;

				PrintLn(Double(21))
			`,
			expected: "42\n",
		},
		{
			name: "Function using function name for return value",
			input: `
				function GetTen: Integer;
				begin
					GetTen := 10;
				end;

				PrintLn(GetTen())
			`,
			expected: "10\n",
		},
		{
			name: "Function called multiple times",
			input: `
				function Square(x: Integer): Integer;
				begin
					Result := x * x;
				end;

				PrintLn(Square(3));
				PrintLn(Square(4));
				PrintLn(Square(5))
			`,
			expected: "9\n16\n25\n",
		},
		{
			name: "Function with string parameter and return",
			input: `
				function Greet(name: String): String;
				begin
					Result := "Hello, " + name;
				end;

				PrintLn(Greet("World"))
			`,
			expected: "Hello, World\n",
		},
		{
			name: "Function with local variables",
			input: `
				function Calculate(x: Integer): Integer;
				begin
					var temp: Integer := x * 2;
					var result: Integer := temp + 10;
					Result := result;
				end;

				PrintLn(Calculate(5))
			`,
			expected: "20\n",
		},
		{
			name: "Multiple functions",
			input: `
				function Add(a, b: Integer): Integer;
				begin
					Result := a + b;
				end;

				function Multiply(a, b: Integer): Integer;
				begin
					Result := a * b;
				end;

				PrintLn(Add(2, 3));
				PrintLn(Multiply(4, 5))
			`,
			expected: "5\n20\n",
		},
		{
			name: "Nested function calls",
			input: `
				function Add(a, b: Integer): Integer;
				begin
					Result := a + b;
				end;

				function Multiply(a, b: Integer): Integer;
				begin
					Result := a * b;
				end;

				PrintLn(Add(Multiply(2, 3), 4))
			`,
			expected: "10\n",
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

// TestProcedures tests procedures (functions without return values).
func TestProcedures(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Simple procedure",
			input: `
				procedure SayHello;
				begin
					PrintLn("Hello!");
				end;

				SayHello()
			`,
			expected: "Hello!\n",
		},
		{
			name: "Procedure with parameters",
			input: `
				procedure Greet(name: String);
				begin
					PrintLn("Hello, " + name);
				end;

				Greet("Alice");
				Greet("Bob")
			`,
			expected: "Hello, Alice\nHello, Bob\n",
		},
		{
			name: "Procedure modifying outer variable",
			input: `
				var counter: Integer := 0;

				procedure Increment;
				begin
					counter := counter + 1;
				end;

				Increment();
				Increment();
				Increment();
				PrintLn(counter)
			`,
			expected: "3\n",
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

// TestRecursiveFunctions tests recursive function calls.
func TestRecursiveFunctions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Factorial",
			input: `
				function Factorial(n: Integer): Integer;
				begin
					if n <= 1 then
						Result := 1
					else
						Result := n * Factorial(n - 1);
				end;

				PrintLn(Factorial(5))
			`,
			expected: "120\n",
		},
		{
			name: "Fibonacci",
			input: `
				function Fibonacci(n: Integer): Integer;
				begin
					if n <= 1 then
						Result := n
					else
						Result := Fibonacci(n - 1) + Fibonacci(n - 2);
				end;

				PrintLn(Fibonacci(0));
				PrintLn(Fibonacci(1));
				PrintLn(Fibonacci(6))
			`,
			expected: "0\n1\n8\n",
		},
		{
			name: "Countdown",
			input: `
				procedure Countdown(n: Integer);
				begin
					if n > 0 then
					begin
						PrintLn(n);
						Countdown(n - 1);
					end;
				end;

				Countdown(5)
			`,
			expected: "5\n4\n3\n2\n1\n",
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

// TestFunctionScopeIsolation tests that function scopes are properly isolated.
func TestFunctionScopeIsolation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Local variable doesn't leak to global scope",
			input: `
				function Test: Integer;
				begin
					var local: Integer := 42;
					Result := local;
				end;

				var x: Integer := Test();
				PrintLn(x)
			`,
			expected: "42\n",
		},
		{
			name: "Same variable name in different scopes",
			input: `
				var x: Integer := 10;

				function GetX: Integer;
				begin
					var x: Integer := 20;
					Result := x;
				end;

				PrintLn(GetX());
				PrintLn(x)
			`,
			expected: "20\n10\n",
		},
		{
			name: "Function can access global variables",
			input: `
				var global: Integer := 100;

				function AddToGlobal(x: Integer): Integer;
				begin
					Result := global + x;
				end;

				PrintLn(AddToGlobal(23))
			`,
			expected: "123\n",
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

// TestFunctionErrors tests error handling in function calls.
func TestFunctionErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name: "Wrong number of arguments",
			input: `
				function Add(a, b: Integer): Integer;
				begin
					Result := a + b;
				end;

				PrintLn(Add(1))
			`,
			expectedErr: "wrong number of arguments",
		},
		{
			name: "Too many arguments",
			input: `
				function Add(a, b: Integer): Integer;
				begin
					Result := a + b;
				end;

				PrintLn(Add(1, 2, 3))
			`,
			expectedErr: "wrong number of arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := testEval(tt.input)
			if !isError(val) {
				t.Errorf("expected error, got %T (%+v)", val, val)
				return
			}

			errVal := val.(*ErrorValue)
			if !strings.Contains(errVal.Message, tt.expectedErr) {
				t.Errorf("wrong error message. expected to contain %q, got=%q",
					tt.expectedErr, errVal.Message)
			}
		})
	}
}

// TestMemberAssignment tests member assignment (obj.field := value) functionality.
// This is crucial for class functionality to work properly.
func TestMemberAssignment(t *testing.T) {
	tests := []struct {
		expected interface{}
		name     string
		input    string
	}{
		{
			name: "Simple member assignment in constructor",
			input: `
				type TPoint = class
					X: Integer;
					Y: Integer;

					function Create(x: Integer; y: Integer): TPoint;
					begin
						Self.X := x;
						Self.Y := y;
					end;

					function GetX(): Integer;
					begin
						Result := Self.X;
					end;
				end;

				var p: TPoint;
				p := TPoint.Create(10, 20);
				p.GetX()
			`,
			expected: int64(10),
		},
		{
			name: "Member assignment in method",
			input: `
				type TCounter = class
					Count: Integer;

					function Create(): TCounter;
					begin
						Self.Count := 0;
					end;

					procedure Increment();
					begin
						Self.Count := Self.Count + 1;
					end;

					function GetCount(): Integer;
					begin
						Result := Self.Count;
					end;
				end;

				var c: TCounter;
				c := TCounter.Create();
				c.Increment();
				c.Increment();
				c.GetCount()
			`,
			expected: int64(2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := testEval(tt.input)

			if isError(val) {
				t.Fatalf("evaluation error: %s", val.String())
			}

			switch expected := tt.expected.(type) {
			case int64:
				testIntegerValue(t, val, expected)
			case string:
				testStringValue(t, val, expected)
			}
		})
	}
}

// TestExternalVarRuntime tests runtime behavior of external variables.
// External variables should raise errors when accessed until
// getter/setter functions are provided.
func TestExternalVarRuntime(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "reading external variable raises error",
			input: `
				var x: Integer external;
				x
			`,
			expectedError: "Unsupported external variable access: x",
		},
		{
			name: "writing external variable raises error",
			input: `
				var y: String external 'externalY';
				y := 'test'
			`,
			expectedError: "Unsupported external variable assignment: y",
		},
		{
			name: "reading external variable in expression raises error",
			input: `
				var z: Integer external;
				var result: Integer;
				result := z + 10
			`,
			expectedError: "Unsupported external variable access: z",
		},
		{
			name: "external variable can be declared but not used",
			input: `
				var ext: Float external;
				var regular: Float;
				regular := 3.14;
				regular
			`,
			expectedError: "", // No error - external var is declared but not accessed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := testEval(tt.input)

			if tt.expectedError == "" {
				// Should not be an error
				if isError(val) {
					t.Fatalf("unexpected error: %s", val.String())
				}
			} else {
				// Should be an error
				if !isError(val) {
					t.Fatalf("expected error containing %q, got: %s", tt.expectedError, val.String())
				}

				errVal := val.(*ErrorValue)
				if !strings.Contains(errVal.Message, tt.expectedError) {
					t.Errorf("error = %q, want to contain %q", errVal.Message, tt.expectedError)
				}
			}
		})
	}
}
