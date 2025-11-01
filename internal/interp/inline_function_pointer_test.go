package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
)

// ============================================================================
// Task 9.51: Inline Function Pointer Interpreter Tests
// ============================================================================

// Helper function to run inline function pointer test code
func runInlineFunctionPointerTest(t *testing.T, input string) (Value, *Interpreter) {
	t.Helper()

	// Lexer
	l := lexer.New(input)

	// Parser
	p := parser.New(l)
	program := p.ParseProgram()

	// Check parser errors
	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	// Semantic analysis
	analyzer := semantic.NewAnalyzer()
	err := analyzer.Analyze(program)
	if err != nil {
		t.Fatalf("Semantic error: %v", err)
	}
	semanticErrors := analyzer.Errors()

	if len(semanticErrors) > 0 {
		t.Fatalf("Semantic errors: %v", semanticErrors)
	}

	// Interpreter
	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	return result, interp
}

// ==============================================================================
// Basic Inline Function Pointer Variable Tests
// ==============================================================================

func TestInlineFunctionPointerVariableDeclaration(t *testing.T) {
	input := `
		var f: function(x: Integer): Integer;
		begin
		end.
	`

	_, interp := runInlineFunctionPointerTest(t, input)

	// Variable should be defined (will be nil until assigned)
	_, ok := interp.env.Get("f")
	if !ok {
		t.Fatal("Variable 'f' should be defined")
	}
}

func TestInlineProcedurePointerVariableDeclaration(t *testing.T) {
	input := `
		var callback: procedure(msg: String);
		begin
		end.
	`

	_, interp := runInlineFunctionPointerTest(t, input)

	_, ok := interp.env.Get("callback")
	if !ok {
		t.Fatal("Variable 'callback' should be defined")
	}
}

// ==============================================================================
// Inline Function Pointer Assignment Tests
// ==============================================================================

func TestInlineFunctionPointerAssignment(t *testing.T) {
	input := `
		function Add(a, b: Integer): Integer;
		begin
			Result := a + b;
		end;

		var op: function(x, y: Integer): Integer;
		begin
			op := @Add;
		end.
	`

	_, interp := runInlineFunctionPointerTest(t, input)

	opVal, ok := interp.env.Get("op")
	if !ok {
		t.Fatal("Variable 'op' not found")
	}

	funcPtr, ok := opVal.(*FunctionPointerValue)
	if !ok {
		t.Fatalf("Expected FunctionPointerValue, got %T", opVal)
	}

	if funcPtr.Function == nil {
		t.Error("Expected function pointer to point to Add function")
	}

	if funcPtr.Function != nil && funcPtr.Function.Name.Value != "Add" {
		t.Errorf("Expected function 'Add', got '%s'", funcPtr.Function.Name.Value)
	}
}

func TestInlineProcedurePointerAssignment(t *testing.T) {
	input := `
		procedure PrintMsg(msg: String);
		begin
			PrintLn(msg);
		end;

		var callback: procedure(s: String);
		begin
			callback := @PrintMsg;
		end.
	`

	_, interp := runInlineFunctionPointerTest(t, input)

	callbackVal, ok := interp.env.Get("callback")
	if !ok {
		t.Fatal("Variable 'callback' not found")
	}

	funcPtr, ok := callbackVal.(*FunctionPointerValue)
	if !ok {
		t.Fatalf("Expected FunctionPointerValue, got %T", callbackVal)
	}

	if funcPtr.Function == nil {
		t.Error("Expected function pointer to point to PrintMsg procedure")
	}

	if funcPtr.Function != nil && funcPtr.Function.Name.Value != "PrintMsg" {
		t.Errorf("Expected procedure 'PrintMsg', got '%s'", funcPtr.Function.Name.Value)
	}
}

// ==============================================================================
// Inline Function Pointer Call Tests
// ==============================================================================

func TestInlineFunctionPointerCall(t *testing.T) {
	input := `
		function Add(a, b: Integer): Integer;
		begin
			Result := a + b;
		end;

		var op: function(x, y: Integer): Integer;
		var result: Integer;
		begin
			op := @Add;
			result := op(5, 3);
		end.
	`

	_, interp := runInlineFunctionPointerTest(t, input)

	resultVal, ok := interp.env.Get("result")
	if !ok {
		t.Fatal("Variable 'result' not found")
	}

	intVal, ok := resultVal.(*IntegerValue)
	if !ok {
		t.Fatalf("Expected IntegerValue, got %T", resultVal)
	}

	if intVal.Value != 8 {
		t.Errorf("Expected result 8, got %d", intVal.Value)
	}
}

func TestInlineProcedurePointerCall(t *testing.T) {
	input := `
		var counter: Integer := 0;

		procedure Increment();
		begin
			counter := counter + 1;
		end;

		var action: procedure();
		begin
			action := @Increment;
			action();
			action();
			action();
		end.
	`

	_, interp := runInlineFunctionPointerTest(t, input)

	counterVal, ok := interp.env.Get("counter")
	if !ok {
		t.Fatal("Variable 'counter' not found")
	}

	intVal, ok := counterVal.(*IntegerValue)
	if !ok {
		t.Fatalf("Expected IntegerValue, got %T", counterVal)
	}

	if intVal.Value != 3 {
		t.Errorf("Expected counter 3, got %d", intVal.Value)
	}
}

func TestInlineFunctionPointerCallWithNoParams(t *testing.T) {
	input := `
		function GetAnswer(): Integer;
		begin
			Result := 42;
		end;

		var getter: function(): Integer;
		var answer: Integer;
		begin
			getter := @GetAnswer;
			answer := getter();
		end.
	`

	_, interp := runInlineFunctionPointerTest(t, input)

	answerVal, ok := interp.env.Get("answer")
	if !ok {
		t.Fatal("Variable 'answer' not found")
	}

	intVal, ok := answerVal.(*IntegerValue)
	if !ok {
		t.Fatalf("Expected IntegerValue, got %T", answerVal)
	}

	if intVal.Value != 42 {
		t.Errorf("Expected answer 42, got %d", intVal.Value)
	}
}

// ==============================================================================
// Inline Function Pointer as Parameter Tests
// ==============================================================================

func TestInlineFunctionPointerAsParameter(t *testing.T) {
	input := `
		function Double(x: Integer): Integer;
		begin
			Result := x * 2;
		end;

		procedure Apply(f: function(x: Integer): Integer; value: Integer);
		var r: Integer;
		begin
			r := f(value);
			PrintLn(IntToStr(r));
		end;

		begin
			Apply(@Double, 5);
		end.
	`

	result, interp := runInlineFunctionPointerTest(t, input)

	if isError(result) {
		t.Errorf("Unexpected error: %v", result)
	}

	// Check output
	output := strings.TrimSpace(interp.output.(*bytes.Buffer).String())
	if output != "10" {
		t.Errorf("Expected output '10', got '%s'", output)
	}
}

func TestInlineProcedurePointerAsParameter(t *testing.T) {
	input := `
		var sum: Integer := 0;

		procedure AddToSum(x: Integer);
		begin
			sum := sum + x;
		end;

		procedure ExecuteTwice(callback: procedure(n: Integer));
		begin
			callback(5);
			callback(10);
		end;

		begin
			ExecuteTwice(@AddToSum);
		end.
	`

	_, interp := runInlineFunctionPointerTest(t, input)

	sumVal, ok := interp.env.Get("sum")
	if !ok {
		t.Fatal("Variable 'sum' not found")
	}

	intVal, ok := sumVal.(*IntegerValue)
	if !ok {
		t.Fatalf("Expected IntegerValue, got %T", sumVal)
	}

	if intVal.Value != 15 {
		t.Errorf("Expected sum 15, got %d", intVal.Value)
	}
}

// ==============================================================================
// Inline Function Pointer with Different Types Tests
// ==============================================================================

func TestInlineFunctionPointerWithStringParameters(t *testing.T) {
	input := `
		function Concat(a, b: String): String;
		begin
			Result := a + b;
		end;

		var op: function(x, y: String): String;
		var result: String;
		begin
			op := @Concat;
			result := op('Hello', 'World');
		end.
	`

	_, interp := runInlineFunctionPointerTest(t, input)

	resultVal, ok := interp.env.Get("result")
	if !ok {
		t.Fatal("Variable 'result' not found")
	}

	strVal, ok := resultVal.(*StringValue)
	if !ok {
		t.Fatalf("Expected StringValue, got %T", resultVal)
	}

	if strVal.Value != "HelloWorld" {
		t.Errorf("Expected 'HelloWorld', got '%s'", strVal.Value)
	}
}

func TestInlineFunctionPointerWithBooleanReturn(t *testing.T) {
	input := `
		function IsEven(x: Integer): Boolean;
		begin
			Result := (x mod 2) = 0;
		end;

		var predicate: function(n: Integer): Boolean;
		var test1: Boolean;
		var test2: Boolean;
		begin
			predicate := @IsEven;
			test1 := predicate(4);
			test2 := predicate(5);
		end.
	`

	_, interp := runInlineFunctionPointerTest(t, input)

	test1Val, _ := interp.env.Get("test1")
	test1Bool := test1Val.(*BooleanValue)
	if !test1Bool.Value {
		t.Error("Expected test1 to be true (4 is even)")
	}

	test2Val, _ := interp.env.Get("test2")
	test2Bool := test2Val.(*BooleanValue)
	if test2Bool.Value {
		t.Error("Expected test2 to be false (5 is odd)")
	}
}

// ==============================================================================
// Inline Function Pointer with Closures/Lambdas Tests
// ==============================================================================

func TestInlineFunctionPointerWithLambdaAssignment(t *testing.T) {
	input := `
		var f: function(x: Integer): Integer;
		begin
			f := lambda(x: Integer): Integer => x * 2;
		end.
	`

	_, interp := runInlineFunctionPointerTest(t, input)

	fVal, ok := interp.env.Get("f")
	if !ok {
		t.Fatal("Variable 'f' not found")
	}

	funcPtr, ok := fVal.(*FunctionPointerValue)
	if !ok {
		t.Fatalf("Expected FunctionPointerValue, got %T", fVal)
	}

	if funcPtr.Lambda == nil {
		t.Error("Expected function pointer to contain a lambda")
	}
}

func TestInlineFunctionPointerWithLambdaCall(t *testing.T) {
	input := `
		var double: function(x: Integer): Integer;
		var result: Integer;
		begin
			double := lambda(x: Integer): Integer => x * 2;
			result := double(7);
		end.
	`

	_, interp := runInlineFunctionPointerTest(t, input)

	resultVal, ok := interp.env.Get("result")
	if !ok {
		t.Fatal("Variable 'result' not found")
	}

	intVal, ok := resultVal.(*IntegerValue)
	if !ok {
		t.Fatalf("Expected IntegerValue, got %T", resultVal)
	}

	if intVal.Value != 14 {
		t.Errorf("Expected result 14, got %d", intVal.Value)
	}
}

func TestInlineFunctionPointerWithClosureCapture(t *testing.T) {
	// Task 9.51: Verify closure semantics with inline function pointer types
	input := `
		var factor: Integer := 10;
		var multiply: function(x: Integer): Integer;
		var result: Integer;
		begin
			multiply := lambda(x: Integer): Integer => x * factor;
			result := multiply(5);
		end.
	`

	_, interp := runInlineFunctionPointerTest(t, input)

	resultVal, ok := interp.env.Get("result")
	if !ok {
		t.Fatal("Variable 'result' not found")
	}

	intVal, ok := resultVal.(*IntegerValue)
	if !ok {
		t.Fatalf("Expected IntegerValue, got %T", resultVal)
	}

	if intVal.Value != 50 {
		t.Errorf("Expected result 50 (closure captured factor=10), got %d", intVal.Value)
	}
}

func TestInlineProcedurePointerWithClosureMutation(t *testing.T) {
	// Task 9.51: Verify closure semantics - mutations should be visible
	input := `
		var counter: Integer := 0;
		var increment: procedure();
		begin
			increment := lambda() begin counter := counter + 1; end;
			increment();
			increment();
			increment();
		end.
	`

	_, interp := runInlineFunctionPointerTest(t, input)

	counterVal, ok := interp.env.Get("counter")
	if !ok {
		t.Fatal("Variable 'counter' not found")
	}

	intVal, ok := counterVal.(*IntegerValue)
	if !ok {
		t.Fatalf("Expected IntegerValue, got %T", counterVal)
	}

	if intVal.Value != 3 {
		t.Errorf("Expected counter 3 (closure mutated), got %d", intVal.Value)
	}
}

// ==============================================================================
// Mixed Inline and Aliased Function Pointer Tests
// ==============================================================================

func TestMixedInlineAndAliasedFunctionPointers(t *testing.T) {
	input := `
		type TComparator = function(a, b: Integer): Integer;

		function Compare(x, y: Integer): Integer;
		begin
			if x < y then Result := -1
			else if x > y then Result := 1
			else Result := 0;
		end;

		var cmp1: TComparator;
		var cmp2: function(a, b: Integer): Integer;
		var result1: Integer;
		var result2: Integer;
		begin
			cmp1 := @Compare;
			cmp2 := @Compare;
			result1 := cmp1(5, 10);
			result2 := cmp2(10, 5);
		end.
	`

	_, interp := runInlineFunctionPointerTest(t, input)

	result1Val, _ := interp.env.Get("result1")
	result1Int := result1Val.(*IntegerValue)
	if result1Int.Value != -1 {
		t.Errorf("Expected result1 = -1, got %d", result1Int.Value)
	}

	result2Val, _ := interp.env.Get("result2")
	result2Int := result2Val.(*IntegerValue)
	if result2Int.Value != 1 {
		t.Errorf("Expected result2 = 1, got %d", result2Int.Value)
	}
}

// ==============================================================================
// Complex Integration Tests
// ==============================================================================

func TestInlineFunctionPointerReassignment(t *testing.T) {
	input := `
		function Add(a, b: Integer): Integer;
		begin
			Result := a + b;
		end;

		function Multiply(a, b: Integer): Integer;
		begin
			Result := a * b;
		end;

		var op: function(x, y: Integer): Integer;
		var result1: Integer;
		var result2: Integer;
		begin
			op := @Add;
			result1 := op(3, 4);

			op := @Multiply;
			result2 := op(3, 4);
		end.
	`

	_, interp := runInlineFunctionPointerTest(t, input)

	result1Val, _ := interp.env.Get("result1")
	result1Int := result1Val.(*IntegerValue)
	if result1Int.Value != 7 {
		t.Errorf("Expected result1 = 7 (add), got %d", result1Int.Value)
	}

	result2Val, _ := interp.env.Get("result2")
	result2Int := result2Val.(*IntegerValue)
	if result2Int.Value != 12 {
		t.Errorf("Expected result2 = 12 (multiply), got %d", result2Int.Value)
	}
}

func TestInlineFunctionPointerMultipleParameters(t *testing.T) {
	input := `
		function Sum(a, b, c: Integer): Integer;
		begin
			Result := a + b + c;
		end;

		var summer: function(x, y, z: Integer): Integer;
		var total: Integer;
		begin
			summer := @Sum;
			total := summer(10, 20, 30);
		end.
	`

	_, interp := runInlineFunctionPointerTest(t, input)

	totalVal, ok := interp.env.Get("total")
	if !ok {
		t.Fatal("Variable 'total' not found")
	}

	intVal, ok := totalVal.(*IntegerValue)
	if !ok {
		t.Fatalf("Expected IntegerValue, got %T", totalVal)
	}

	if intVal.Value != 60 {
		t.Errorf("Expected total 60, got %d", intVal.Value)
	}
}

func TestInlineFunctionPointerPassedToFunctionWithInlineParameter(t *testing.T) {
	// Test passing inline function pointers to functions that accept inline function pointer parameters
	input := `
		function Apply(f: function(x: Integer): Integer; value: Integer): Integer;
		begin
			Result := f(value);
		end;

		function Double(x: Integer): Integer;
		begin
			Result := x * 2;
		end;

		var result: Integer;
		begin
			result := Apply(@Double, 7);
		end.
	`

	_, interp := runInlineFunctionPointerTest(t, input)

	resultVal, ok := interp.env.Get("result")
	if !ok {
		t.Fatal("Variable 'result' not found")
	}

	intVal, ok := resultVal.(*IntegerValue)
	if !ok {
		t.Fatalf("Expected IntegerValue, got %T", resultVal)
	}

	if intVal.Value != 14 {
		t.Errorf("Expected result 14, got %d", intVal.Value)
	}
}
