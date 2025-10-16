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
		{"5 + 2.5", 7.5},    // Mixed int/float
		{"10 / 4", 2.5},     // Integer division produces float with /
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
