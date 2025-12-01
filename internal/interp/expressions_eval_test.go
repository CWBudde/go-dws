package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
)

// testEvalExpression is a helper that evaluates an expression with semantic analysis
func testEvalExpression(input string, t *testing.T) (Value, string) {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	analyzer := semantic.NewAnalyzer()
	_ = analyzer.Analyze(program)
	// Filter out hints - only treat actual errors as blocking
	hasRealErrors := false
	for _, err := range analyzer.Errors() {
		if !strings.HasPrefix(err, "Hint:") {
			hasRealErrors = true
			break
		}
	}
	if hasRealErrors {
		// Some tests expect semantic errors
		return nil, ""
	}

	var buf bytes.Buffer
	interp := New(&buf)
	if semanticInfo := analyzer.GetSemanticInfo(); semanticInfo != nil {
		interp.SetSemanticInfo(semanticInfo)
	}
	val := interp.Eval(program)
	return val, buf.String()
}

// TestEvalIdentifierBasic tests basic identifier evaluation
func TestEvalIdentifierBasic(t *testing.T) {
	input := "var x := 42; PrintLn(x);"
	_, output := testEvalExpression(input, t)
	expected := "42\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestEvalIdentifierCaseInsensitive tests case-insensitive variable names
func TestEvalIdentifierCaseInsensitive(t *testing.T) {
	input := "var myVar := 10; PrintLn(MYVAR);"
	_, output := testEvalExpression(input, t)
	expected := "10\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestEvalUnaryMinusInteger tests unary minus on integers
func TestEvalUnaryMinusInteger(t *testing.T) {
	input := "PrintLn(-42);"
	_, output := testEvalExpression(input, t)
	expected := "-42\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestEvalUnaryMinusFloat tests unary minus on floats
func TestEvalUnaryMinusFloat(t *testing.T) {
	input := "PrintLn(-3.14);"
	_, output := testEvalExpression(input, t)
	if !strings.Contains(output, "-3.14") {
		t.Errorf("Expected output containing -3.14, got %q", output)
	}
}

// TestEvalUnaryPlusInteger tests unary plus on integers
func TestEvalUnaryPlusInteger(t *testing.T) {
	input := "PrintLn(+42);"
	_, output := testEvalExpression(input, t)
	expected := "42\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestEvalUnaryNotBoolean tests not operator on boolean
func TestEvalUnaryNotBoolean(t *testing.T) {
	input := "PrintLn(not True);"
	_, output := testEvalExpression(input, t)
	expected := "False\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestEvalUnaryNotInteger tests not operator on integer (bitwise)
func TestEvalUnaryNotInteger(t *testing.T) {
	input := "PrintLn(not 0);"
	_, output := testEvalExpression(input, t)
	expected := "-1\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestEvalBinaryIntegerAddition tests integer addition
func TestEvalBinaryIntegerAddition(t *testing.T) {
	input := "PrintLn(5 + 3);"
	_, output := testEvalExpression(input, t)
	expected := "8\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestEvalBinaryIntegerSubtraction tests integer subtraction
func TestEvalBinaryIntegerSubtraction(t *testing.T) {
	input := "PrintLn(10 - 3);"
	_, output := testEvalExpression(input, t)
	expected := "7\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestEvalBinaryIntegerMultiplication tests integer multiplication
func TestEvalBinaryIntegerMultiplication(t *testing.T) {
	input := "PrintLn(4 * 5);"
	_, output := testEvalExpression(input, t)
	expected := "20\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestEvalBinaryIntegerDivision tests integer division
func TestEvalBinaryIntegerDivision(t *testing.T) {
	input := "PrintLn(10 / 2);"
	_, output := testEvalExpression(input, t)
	expected := "5\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestEvalBinaryIntegerDiv tests integer div operator
func TestEvalBinaryIntegerDiv(t *testing.T) {
	input := "PrintLn(10 div 3);"
	_, output := testEvalExpression(input, t)
	expected := "3\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestEvalBinaryIntegerMod tests integer mod operator
func TestEvalBinaryIntegerMod(t *testing.T) {
	input := "PrintLn(10 mod 3);"
	_, output := testEvalExpression(input, t)
	expected := "1\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestEvalBinaryFloatAddition tests float addition
func TestEvalBinaryFloatAddition(t *testing.T) {
	input := "PrintLn(5.5 + 2.5);"
	_, output := testEvalExpression(input, t)
	if !strings.Contains(output, "8") {
		t.Errorf("Expected output containing 8, got %q", output)
	}
}

// TestEvalBinaryStringConcatenation tests string concatenation
func TestEvalBinaryStringConcatenation(t *testing.T) {
	input := "PrintLn('Hello' + ' ' + 'World');"
	_, output := testEvalExpression(input, t)
	expected := "Hello World\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestEvalBinaryBooleanAnd tests boolean and
func TestEvalBinaryBooleanAnd(t *testing.T) {
	input := "PrintLn(True and False);"
	_, output := testEvalExpression(input, t)
	expected := "False\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestEvalBinaryBooleanOr tests boolean or
func TestEvalBinaryBooleanOr(t *testing.T) {
	input := "PrintLn(True or False);"
	_, output := testEvalExpression(input, t)
	expected := "True\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestEvalBinaryBooleanXor tests boolean xor
func TestEvalBinaryBooleanXor(t *testing.T) {
	input := "PrintLn(True xor False);"
	_, output := testEvalExpression(input, t)
	expected := "True\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestEvalBinaryIntegerComparisons tests integer comparisons
func TestEvalBinaryIntegerComparisons(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"equals", "PrintLn(5 = 5);", "True\n"},
		{"not equals", "PrintLn(5 <> 3);", "True\n"},
		{"less than", "PrintLn(3 < 5);", "True\n"},
		{"greater than", "PrintLn(5 > 3);", "True\n"},
		{"less or equal", "PrintLn(3 <= 5);", "True\n"},
		{"greater or equal", "PrintLn(5 >= 3);", "True\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalExpression(tt.input, t)
			if output != tt.want {
				t.Errorf("Expected %q, got %q", tt.want, output)
			}
		})
	}
}

// TestEvalBinaryBitwiseOperations tests bitwise operations
func TestEvalBinaryBitwiseOperations(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"and", "PrintLn(5 and 3);", "1\n"},
		{"or", "PrintLn(5 or 3);", "7\n"},
		{"xor", "PrintLn(5 xor 3);", "6\n"},
		{"shl", "PrintLn(4 shl 2);", "16\n"},
		{"shr", "PrintLn(16 shr 2);", "4\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalExpression(tt.input, t)
			if output != tt.want {
				t.Errorf("Expected %q, got %q", tt.want, output)
			}
		})
	}
}

// TestEvalCoalesceOperator tests the ?? (coalesce) operator
func TestEvalCoalesceOperator(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"truthy left", "PrintLn(5 ?? 10);", "5\n"},
		{"falsey left zero", "PrintLn(0 ?? 10);", "10\n"},
		{"falsey left empty string", "PrintLn('' ?? 'default');", "default\n"},
		{"truthy left string", "PrintLn('hello' ?? 'default');", "hello\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalExpression(tt.input, t)
			if output != tt.want {
				t.Errorf("Expected %q, got %q", tt.want, output)
			}
		})
	}
}

// TestEvalShortCircuitAnd tests short-circuit evaluation for 'and'
func TestEvalShortCircuitAnd(t *testing.T) {
	input := `
		var sideEffect := 0;
		function HasSideEffect: Boolean;
		begin
			sideEffect := sideEffect + 1;
			Result := True;
		end;
		var result := False and HasSideEffect();
		PrintLn(result);
		PrintLn(sideEffect);
	`
	_, output := testEvalExpression(input, t)
	// Side effect should NOT execute because left side is false
	if !strings.Contains(output, "False") || !strings.Contains(output, "0") {
		t.Errorf("Expected False and 0 (no side effect), got %q", output)
	}
}

// TestEvalShortCircuitOr tests short-circuit evaluation for 'or'
func TestEvalShortCircuitOr(t *testing.T) {
	input := `
		var sideEffect := 0;
		function HasSideEffect: Boolean;
		begin
			sideEffect := sideEffect + 1;
			Result := False;
		end;
		var result := True or HasSideEffect();
		PrintLn(result);
		PrintLn(sideEffect);
	`
	_, output := testEvalExpression(input, t)
	// Side effect should NOT execute because left side is true
	if !strings.Contains(output, "True") || !strings.Contains(output, "0") {
		t.Errorf("Expected True and 0 (no side effect), got %q", output)
	}
}

// TestEvalInOperatorSet tests 'in' operator with sets
func TestEvalInOperatorSet(t *testing.T) {
	input := `
		type TColor = (Red, Green, Blue);
		var color: TColor := Red;
		PrintLn(color in [Red, Green]);
	`
	_, output := testEvalExpression(input, t)
	expected := "True\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestEvalInOperatorString tests 'in' operator with strings
func TestEvalInOperatorString(t *testing.T) {
	input := "PrintLn('a' in 'abc');"
	_, output := testEvalExpression(input, t)
	expected := "True\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestEvalIfExpression tests inline if-then-else
func TestEvalIfExpression(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"true branch", "var x := if True then 10 else 20; PrintLn(x);", "10\n"},
		{"false branch", "var x := if False then 10 else 20; PrintLn(x);", "20\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalExpression(tt.input, t)
			if output != tt.want {
				t.Errorf("Expected %q, got %q", tt.want, output)
			}
		})
	}
}

// TestGetTypeByName tests type name resolution
func TestGetTypeByName(t *testing.T) {
	var buf bytes.Buffer
	interp := New(&buf)

	tests := []struct {
		name     string
		typeName string
		wantType string
	}{
		{name: "integer type", typeName: "Integer", wantType: "Integer"},
		{name: "float type", typeName: "Float", wantType: "Float"},
		{name: "string type", typeName: "String", wantType: "String"},
		{name: "boolean type", typeName: "Boolean", wantType: "Boolean"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := interp.getTypeByName(tt.typeName)
			if typ == nil {
				t.Errorf("Expected type %s, got nil", tt.wantType)
			} else if typ.String() != tt.wantType {
				t.Errorf("Expected type %s, got %s", tt.wantType, typ.String())
			}
		})
	}
}

// TestIsFalsey tests the isFalsey helper function
func TestIsFalsey(t *testing.T) {
	tests := []struct {
		value    Value
		name     string
		expected bool
	}{
		{nil, "nil is falsey", true},
		{&IntegerValue{Value: 0}, "zero integer is falsey", true},
		{&IntegerValue{Value: 5}, "non-zero integer is truthy", false},
		{&FloatValue{Value: 0.0}, "zero float is falsey", true},
		{&StringValue{Value: ""}, "empty string is falsey", true},
		{&BooleanValue{Value: false}, "false boolean is falsey", true},
		{&BooleanValue{Value: true}, "true boolean is truthy", false},
		{&NilValue{}, "nil value is falsey", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isFalsey(tt.value)
			if result != tt.expected {
				t.Errorf("Expected isFalsey(%v) = %v, got %v", tt.value, tt.expected, result)
			}
		})
	}
}

// TestIsNumericType tests the isNumericType helper function
func TestIsNumericType(t *testing.T) {
	tests := []struct {
		name     string
		typeStr  string
		expected bool
	}{
		{"INTEGER is numeric", "INTEGER", true},
		{"FLOAT is numeric", "FLOAT", true},
		{"STRING is not numeric", "STRING", false},
		{"BOOLEAN is not numeric", "BOOLEAN", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNumericType(tt.typeStr)
			if result != tt.expected {
				t.Errorf("Expected isNumericType(%s) = %v, got %v", tt.typeStr, tt.expected, result)
			}
		})
	}
}
