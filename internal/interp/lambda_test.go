package interp

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
)

// Task 9.225: Interpreter tests for lambda expressions and closures
// Task 9.226: Complex closure tests including nested lambdas

// Helper function to run lambda test code
func runLambdaTest(t *testing.T, input string) (Value, *Interpreter) {
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
	_ = analyzer.Analyze(program)
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
// Basic Lambda Tests
// ==============================================================================

func TestBasicLambdaCreation(t *testing.T) {
	input := `
		var f := lambda(x: Integer): Integer begin Result := x * 2; end;
	`

	_, interp := runLambdaTest(t, input)

	// Check that f is a lambda value
	fVal, ok := interp.env.Get("f")
	if !ok {
		t.Fatal("Variable 'f' not found in environment")
	}

	funcPtr, ok := fVal.(*FunctionPointerValue)
	if !ok {
		t.Fatalf("Expected FunctionPointerValue, got %T", fVal)
	}

	if funcPtr.Lambda == nil {
		t.Fatal("Lambda field is nil, expected LambdaExpression")
	}

	if funcPtr.Type() != "LAMBDA" {
		t.Errorf("Expected type 'LAMBDA', got '%s'", funcPtr.Type())
	}
}

func TestBasicLambdaCall(t *testing.T) {
	input := `
		var double := lambda(x: Integer): Integer begin Result := x * 2; end;
		var result := double(5);
	`

	_, interp := runLambdaTest(t, input)

	// Check the result
	resultVal, ok := interp.env.Get("result")
	if !ok {
		t.Fatal("Variable 'result' not found")
	}

	intVal, ok := resultVal.(*IntegerValue)
	if !ok {
		t.Fatalf("Expected IntegerValue, got %T", resultVal)
	}

	if intVal.Value != 10 {
		t.Errorf("Expected result 10, got %d", intVal.Value)
	}
}

func TestLambdaWithMultipleParameters(t *testing.T) {
	input := `
		var add := lambda(a: Integer; b: Integer): Integer begin Result := a + b; end;
		var sum := add(3, 7);
	`

	_, interp := runLambdaTest(t, input)

	sumVal, ok := interp.env.Get("sum")
	if !ok {
		t.Fatal("Variable 'sum' not found")
	}

	intVal, ok := sumVal.(*IntegerValue)
	if !ok {
		t.Fatalf("Expected IntegerValue, got %T", sumVal)
	}

	if intVal.Value != 10 {
		t.Errorf("Expected sum 10, got %d", intVal.Value)
	}
}

func TestShorthandLambdaSyntax(t *testing.T) {
	input := `
		var triple := lambda(x: Integer): Integer => x * 3;
		var result := triple(4);
	`

	_, interp := runLambdaTest(t, input)

	resultVal, ok := interp.env.Get("result")
	if !ok {
		t.Fatal("Variable 'result' not found")
	}

	intVal, ok := resultVal.(*IntegerValue)
	if !ok {
		t.Fatalf("Expected IntegerValue, got %T", resultVal)
	}

	if intVal.Value != 12 {
		t.Errorf("Expected result 12, got %d", intVal.Value)
	}
}

func TestProcedureLambda(t *testing.T) {
	input := `
		var counter: Integer := 0;
		var incrementCounter := lambda() begin counter := counter + 1; end;
		incrementCounter();
		incrementCounter();
	`

	_, interp := runLambdaTest(t, input)

	counterVal, ok := interp.env.Get("counter")
	if !ok {
		t.Fatal("Variable 'counter' not found")
	}

	intVal, ok := counterVal.(*IntegerValue)
	if !ok {
		t.Fatalf("Expected IntegerValue, got %T", counterVal)
	}

	if intVal.Value != 2 {
		t.Errorf("Expected counter 2, got %d", intVal.Value)
	}
}

// ==============================================================================
// Closure and Variable Capture Tests
// ==============================================================================

func TestSimpleClosureCapture(t *testing.T) {
	input := `
		var factor: Integer := 10;
		var multiply := lambda(x: Integer): Integer => x * factor;
		var result := multiply(5);
	`

	_, interp := runLambdaTest(t, input)

	resultVal, ok := interp.env.Get("result")
	if !ok {
		t.Fatal("Variable 'result' not found")
	}

	intVal, ok := resultVal.(*IntegerValue)
	if !ok {
		t.Fatalf("Expected IntegerValue, got %T", resultVal)
	}

	if intVal.Value != 50 {
		t.Errorf("Expected result 50, got %d", intVal.Value)
	}
}

func TestClosureCaptureMultipleVariables(t *testing.T) {
	input := `
		var a: Integer := 10;
		var b: Integer := 20;
		var sum := lambda(): Integer => a + b;
		var result := sum();
	`

	_, interp := runLambdaTest(t, input)

	resultVal, ok := interp.env.Get("result")
	if !ok {
		t.Fatal("Variable 'result' not found")
	}

	intVal, ok := resultVal.(*IntegerValue)
	if !ok {
		t.Fatalf("Expected IntegerValue, got %T", resultVal)
	}

	if intVal.Value != 30 {
		t.Errorf("Expected result 30, got %d", intVal.Value)
	}
}

func TestClosureMutatesCapturedVariable(t *testing.T) {
	// Task 9.224: Test reference semantics - mutations should be visible
	input := `
		var counter: Integer := 0;
		var increment := lambda() begin counter := counter + 1; end;

		increment();
		increment();
		increment();
	`

	_, interp := runLambdaTest(t, input)

	counterVal, ok := interp.env.Get("counter")
	if !ok {
		t.Fatal("Variable 'counter' not found")
	}

	intVal, ok := counterVal.(*IntegerValue)
	if !ok {
		t.Fatalf("Expected IntegerValue, got %T", counterVal)
	}

	if intVal.Value != 3 {
		t.Errorf("Expected counter 3 (reference semantics), got %d", intVal.Value)
	}
}

func TestClosureReadsUpdatedCapturedVariable(t *testing.T) {
	// Test that lambda sees updates to captured variables
	input := `
		var x: Integer := 5;
		var getX := lambda(): Integer => x;

		var result1 := getX();
		x := 10;
		var result2 := getX();
	`

	_, interp := runLambdaTest(t, input)

	result1Val, _ := interp.env.Get("result1")
	result1Int := result1Val.(*IntegerValue)
	if result1Int.Value != 5 {
		t.Errorf("Expected result1 = 5, got %d", result1Int.Value)
	}

	result2Val, _ := interp.env.Get("result2")
	result2Int := result2Val.(*IntegerValue)
	if result2Int.Value != 10 {
		t.Errorf("Expected result2 = 10 (updated), got %d", result2Int.Value)
	}
}

// ==============================================================================
// Lambda as First-Class Value Tests
// ==============================================================================

func TestLambdaStoredInVariable(t *testing.T) {
	input := `
		var op := lambda(x, y: Integer): Integer => x + y;
		var result := op(3, 4);
	`

	_, interp := runLambdaTest(t, input)

	resultVal, _ := interp.env.Get("result")
	intVal := resultVal.(*IntegerValue)

	if intVal.Value != 7 {
		t.Errorf("Expected result 7, got %d", intVal.Value)
	}
}

func TestLambdaReassignment(t *testing.T) {
	input := `
		var op := lambda(x: Integer): Integer => x * 2;
		var result1 := op(5);

		op := lambda(x: Integer): Integer => x * 3;
		var result2 := op(5);
	`

	_, interp := runLambdaTest(t, input)

	result1Val, _ := interp.env.Get("result1")
	result1Int := result1Val.(*IntegerValue)
	if result1Int.Value != 10 {
		t.Errorf("Expected result1 = 10, got %d", result1Int.Value)
	}

	result2Val, _ := interp.env.Get("result2")
	result2Int := result2Val.(*IntegerValue)
	if result2Int.Value != 15 {
		t.Errorf("Expected result2 = 15, got %d", result2Int.Value)
	}
}

// ==============================================================================
// Nested Lambda Tests (Task 9.226)
// ==============================================================================

func TestNestedLambdas(t *testing.T) {
	input := `
		var outer := lambda(x: Integer): Integer begin
			var inner := lambda(y: Integer): Integer => x + y;
			Result := inner(10);
		end;
		var result := outer(5);
	`

	_, interp := runLambdaTest(t, input)

	resultVal, ok := interp.env.Get("result")
	if !ok {
		t.Fatal("Variable 'result' not found")
	}

	intVal, ok := resultVal.(*IntegerValue)
	if !ok {
		t.Fatalf("Expected IntegerValue, got %T", resultVal)
	}

	if intVal.Value != 15 {
		t.Errorf("Expected result 15 (5 + 10), got %d", intVal.Value)
	}
}

func TestNestedLambdaCaptureFromMultipleLevels(t *testing.T) {
	input := `
		var a: Integer := 100;

		var outer := lambda(b: Integer): Integer begin
			var inner := lambda(c: Integer): Integer => a + b + c;
			Result := inner(1);
		end;

		var result := outer(10);
	`

	_, interp := runLambdaTest(t, input)

	resultVal, _ := interp.env.Get("result")
	intVal := resultVal.(*IntegerValue)

	if intVal.Value != 111 {
		t.Errorf("Expected result 111 (100+10+1), got %d", intVal.Value)
	}
}

// ==============================================================================
// Lambda with Different Types Tests
// ==============================================================================

func TestLambdaWithStringParameters(t *testing.T) {
	input := `
		var concat := lambda(a, b: String): String => a + b;
		var result := concat('Hello', 'World');
	`

	_, interp := runLambdaTest(t, input)

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

func TestLambdaWithFloatParameters(t *testing.T) {
	input := `
		var multiply := lambda(x, y: Float): Float => x * y;
		var result := multiply(2.5, 4.0);
	`

	_, interp := runLambdaTest(t, input)

	resultVal, _ := interp.env.Get("result")
	floatVal := resultVal.(*FloatValue)

	if floatVal.Value != 10.0 {
		t.Errorf("Expected 10.0, got %f", floatVal.Value)
	}
}

func TestLambdaWithBooleanReturn(t *testing.T) {
	input := `
		var isEven := lambda(x: Integer): Boolean => (x mod 2) = 0;
		var result1 := isEven(4);
		var result2 := isEven(5);
	`

	_, interp := runLambdaTest(t, input)

	result1Val, _ := interp.env.Get("result1")
	bool1 := result1Val.(*BooleanValue)
	if !bool1.Value {
		t.Error("Expected isEven(4) = true")
	}

	result2Val, _ := interp.env.Get("result2")
	bool2 := result2Val.(*BooleanValue)
	if bool2.Value {
		t.Error("Expected isEven(5) = false")
	}
}

// ==============================================================================
// Complex Control Flow Tests
// ==============================================================================

func TestLambdaWithIfStatement(t *testing.T) {
	input := `
		var abs := lambda(x: Integer): Integer begin
			if x < 0 then
				Result := -x
			else
				Result := x;
		end;

		var result1 := abs(-5);
		var result2 := abs(7);
	`

	_, interp := runLambdaTest(t, input)

	result1Val, _ := interp.env.Get("result1")
	int1 := result1Val.(*IntegerValue)
	if int1.Value != 5 {
		t.Errorf("Expected abs(-5) = 5, got %d", int1.Value)
	}

	result2Val, _ := interp.env.Get("result2")
	int2 := result2Val.(*IntegerValue)
	if int2.Value != 7 {
		t.Errorf("Expected abs(7) = 7, got %d", int2.Value)
	}
}

func TestLambdaWithLoop(t *testing.T) {
	input := `
		var factorial := lambda(n: Integer): Integer begin
			var result: Integer := 1;
			var i: Integer;
			for i := 1 to n do
				result := result * i;
			Result := result;
		end;

		var result := factorial(5);
	`

	_, interp := runLambdaTest(t, input)

	resultVal, _ := interp.env.Get("result")
	intVal := resultVal.(*IntegerValue)

	if intVal.Value != 120 {
		t.Errorf("Expected factorial(5) = 120, got %d", intVal.Value)
	}
}

// ==============================================================================
// Edge Cases and Error Handling
// ==============================================================================

func TestLambdaWithNoParameters(t *testing.T) {
	input := `
		var getFortyTwo := lambda(): Integer => 42;
		var result := getFortyTwo();
	`

	_, interp := runLambdaTest(t, input)

	resultVal, _ := interp.env.Get("result")
	intVal := resultVal.(*IntegerValue)

	if intVal.Value != 42 {
		t.Errorf("Expected 42, got %d", intVal.Value)
	}
}

func TestLambdaNoCapturedVariables(t *testing.T) {
	// Pure lambda - no captured variables
	input := `
		var square := lambda(x: Integer): Integer => x * x;
		var result := square(6);
	`

	_, interp := runLambdaTest(t, input)

	resultVal, _ := interp.env.Get("result")
	intVal := resultVal.(*IntegerValue)

	if intVal.Value != 36 {
		t.Errorf("Expected 36, got %d", intVal.Value)
	}
}

// ==============================================================================
// Closure Creation in Loops (Task 9.226)
// ==============================================================================

func TestLambdaCapturesLoopVariable(t *testing.T) {
	t.Skip("Requires procedure type variables - not yet implemented")
	// Test that each lambda captures the current value of the loop variable
	// Note: In DWScript, loop variables are shared, so all lambdas see final value
	input := `
		var i: Integer;
		var lastFunc: procedure;

		for i := 1 to 3 do begin
			lastFunc := lambda() begin end;
		end;

		// After loop, i should be 4 (one past the end)
	`

	_, interp := runLambdaTest(t, input)

	iVal, _ := interp.env.Get("i")
	intVal := iVal.(*IntegerValue)

	// Loop variable ends at 4 (one past 'to 3')
	if intVal.Value != 4 {
		t.Errorf("Expected i = 4 after loop, got %d", intVal.Value)
	}
}

// ==============================================================================
// Multiple Lambdas Sharing Captures (Task 9.226)
// ==============================================================================

func TestMultipleLambdasShareCapturedVariable(t *testing.T) {
	input := `
		var shared: Integer := 0;

		var increment := lambda() begin shared := shared + 1; end;
		var decrement := lambda() begin shared := shared - 1; end;
		var getValue := lambda(): Integer => shared;

		increment();
		increment();
		decrement();

		var result := getValue();
	`

	_, interp := runLambdaTest(t, input)

	resultVal, _ := interp.env.Get("result")
	intVal := resultVal.(*IntegerValue)

	if intVal.Value != 1 {
		t.Errorf("Expected result 1, got %d", intVal.Value)
	}
}

// ==============================================================================
// Lambda Type String Representation
// ==============================================================================

func TestLambdaStringRepresentation(t *testing.T) {
	input := `
		var f := lambda(x: Integer): Integer => x * 2;
	`

	_, interp := runLambdaTest(t, input)

	fVal, _ := interp.env.Get("f")
	funcPtr := fVal.(*FunctionPointerValue)

	strRep := funcPtr.String()
	if strRep != "<lambda>" {
		t.Errorf("Expected string '<lambda>', got '%s'", strRep)
	}
}

// ==============================================================================
// Integration with Existing Features
// ==============================================================================

func TestLambdaWithArrayAccess(t *testing.T) {
	t.Skip("Requires dynamic array support - not yet implemented")
	input := `
		var numbers: array of Integer;
		SetLength(numbers, 3);
		numbers[0] := 10;
		numbers[1] := 20;
		numbers[2] := 30;

		var getElement := lambda(index: Integer): Integer => numbers[index];
		var result := getElement(1);
	`

	_, interp := runLambdaTest(t, input)

	resultVal, _ := interp.env.Get("result")
	intVal := resultVal.(*IntegerValue)

	if intVal.Value != 20 {
		t.Errorf("Expected 20, got %d", intVal.Value)
	}
}

func TestLambdaModifiesArray(t *testing.T) {
	t.Skip("Requires dynamic array support - not yet implemented")
	input := `
		var numbers: array of Integer;
		SetLength(numbers, 3);
		numbers[0] := 1;
		numbers[1] := 2;
		numbers[2] := 3;

		var doubleElement := lambda(index: Integer) begin
			numbers[index] := numbers[index] * 2;
		end;

		doubleElement(1);
		var result := numbers[1];
	`

	_, interp := runLambdaTest(t, input)

	resultVal, _ := interp.env.Get("result")
	intVal := resultVal.(*IntegerValue)

	if intVal.Value != 4 {
		t.Errorf("Expected 4, got %d", intVal.Value)
	}
}

// ==============================================================================
// Higher-Order Functions Tests (Task 9.228)
// ==============================================================================

func TestMapBasic(t *testing.T) {
	input := `
		type TIntArray = array[0..4] of Integer;
		var numbers: TIntArray;
		numbers[0] := 1;
		numbers[1] := 2;
		numbers[2] := 3;
		numbers[3] := 4;
		numbers[4] := 5;

		var doubled := Map(numbers, lambda(x: Integer): Integer => x * 2);

		var sum: Integer := 0;
		var i: Integer;
		for i := 0 to 4 do
			sum := sum + doubled[i];
	`

	_, interp := runLambdaTest(t, input)

	sumVal, _ := interp.env.Get("sum")
	intVal := sumVal.(*IntegerValue)

	// doubled = [2, 4, 6, 8, 10], sum = 30
	if intVal.Value != 30 {
		t.Errorf("Expected sum 30, got %d", intVal.Value)
	}
}

func TestMapWithClosure(t *testing.T) {
	input := `
		type TIntArray = array[0..2] of Integer;
		var numbers: TIntArray;
		numbers[0] := 1;
		numbers[1] := 2;
		numbers[2] := 3;

		var multiplier: Integer := 10;
		var scaled := Map(numbers, lambda(x: Integer): Integer => x * multiplier);

		var result := scaled[1];
	`

	result, interp := runLambdaTest(t, input)

	// Check if there was an error during execution
	if isError(result) {
		t.Fatalf("Execution failed: %v", result)
	}

	resultVal, ok := interp.env.Get("result")
	if !ok {
		t.Fatal("Variable 'result' not found in environment")
	}
	if resultVal == nil {
		t.Fatal("Variable 'result' is nil")
	}
	intVal := resultVal.(*IntegerValue)

	// numbers[1] * 10 = 2 * 10 = 20
	if intVal.Value != 20 {
		t.Errorf("Expected 20, got %d", intVal.Value)
	}
}

func TestFilterBasic(t *testing.T) {
	input := `
		type TIntArray = array[0..4] of Integer;
		var numbers: TIntArray;
		numbers[0] := 1;
		numbers[1] := 2;
		numbers[2] := 3;
		numbers[3] := 4;
		numbers[4] := 5;

		var evens := Filter(numbers, lambda(x: Integer): Boolean => (x mod 2) = 0);

		var count: Integer := Length(evens);
	`

	result, interp := runLambdaTest(t, input)

	// Check if there was an error during execution
	if isError(result) {
		t.Fatalf("Execution failed: %v", result)
	}

	countVal, ok := interp.env.Get("count")
	if !ok {
		t.Fatal("Variable 'count' not found in environment")
	}
	if countVal == nil {
		t.Fatal("Variable 'count' is nil")
	}
	intVal := countVal.(*IntegerValue)

	// evens = [2, 4], count = 2
	if intVal.Value != 2 {
		t.Errorf("Expected count 2, got %d", intVal.Value)
	}
}

func TestFilterWithComplexPredicate(t *testing.T) {
	input := `
		type TIntArray = array[0..9] of Integer;
		var numbers: TIntArray;
		var i: Integer;
		for i := 0 to 9 do
			numbers[i] := i;

		var filtered := Filter(numbers, lambda(x: Integer): Boolean begin
			Result := (x > 2) and (x < 8);
		end);

		var count: Integer := Length(filtered);
	`

	result, interp := runLambdaTest(t, input)

	// Check if there was an error during execution
	if isError(result) {
		t.Fatalf("Execution failed: %v", result)
	}

	countVal, ok := interp.env.Get("count")
	if !ok {
		t.Fatal("Variable 'count' not found in environment")
	}
	if countVal == nil {
		t.Fatal("Variable 'count' is nil")
	}
	intVal := countVal.(*IntegerValue)

	// filtered = [3, 4, 5, 6, 7], count = 5
	if intVal.Value != 5 {
		t.Errorf("Expected count 5, got %d", intVal.Value)
	}
}

func TestReduceSum(t *testing.T) {
	input := `
		type TIntArray = array[0..5] of Integer;
		var numbers: TIntArray;
		numbers[0] := 1;
		numbers[1] := 2;
		numbers[2] := 3;
		numbers[3] := 4;
		numbers[4] := 5;

		var sum := Reduce(numbers, lambda(acc: Integer, x: Integer): Integer => acc + x, 0);
	`

	_, interp := runLambdaTest(t, input)

	sumVal, _ := interp.env.Get("sum")
	intVal := sumVal.(*IntegerValue)

	// sum = 1+2+3+4+5 = 15
	if intVal.Value != 15 {
		t.Errorf("Expected sum 15, got %d", intVal.Value)
	}
}

func TestReduceProduct(t *testing.T) {
	input := `
		type TIntArray = array[0..4] of Integer;
		var numbers: TIntArray;
		numbers[0] := 2;
		numbers[1] := 3;
		numbers[2] := 4;
		numbers[3] := 5;

		var product := Reduce(numbers, lambda(acc: Integer, x: Integer): Integer => acc * x, 1);
	`

	_, interp := runLambdaTest(t, input)

	productVal, _ := interp.env.Get("product")
	intVal := productVal.(*IntegerValue)

	// product = 2*3*4*5 = 120
	if intVal.Value != 120 {
		t.Errorf("Expected product 120, got %d", intVal.Value)
	}
}

func TestReduceMax(t *testing.T) {
	input := `
		type TIntArray = array[0..5] of Integer;
		var numbers: TIntArray;
		numbers[0] := 3;
		numbers[1] := 7;
		numbers[2] := 2;
		numbers[3] := 9;
		numbers[4] := 1;

		var maximum := Reduce(numbers, lambda(acc: Integer, x: Integer): Integer begin
			if x > acc then
				Result := x
			else
				Result := acc;
		end, numbers[0]);
	`

	_, interp := runLambdaTest(t, input)

	maxVal, _ := interp.env.Get("maximum")
	intVal := maxVal.(*IntegerValue)

	// maximum = 9
	if intVal.Value != 9 {
		t.Errorf("Expected maximum 9, got %d", intVal.Value)
	}
}

func TestForEachBasic(t *testing.T) {
	input := `
		type TIntArray = array[0..3] of Integer;
		var numbers: TIntArray;
		numbers[0] := 1;
		numbers[1] := 2;
		numbers[2] := 3;

		var sum: Integer := 0;
		ForEach(numbers, lambda(x: Integer) begin
			sum := sum + x;
		end);
	`

	_, interp := runLambdaTest(t, input)

	sumVal, _ := interp.env.Get("sum")
	intVal := sumVal.(*IntegerValue)

	// sum = 1+2+3 = 6
	if intVal.Value != 6 {
		t.Errorf("Expected sum 6, got %d", intVal.Value)
	}
}

func TestForEachWithOutput(t *testing.T) {
	input := `
		type TIntArray = array[0..3] of Integer;
		var numbers: TIntArray;
		numbers[0] := 10;
		numbers[1] := 20;
		numbers[2] := 30;

		ForEach(numbers, lambda(x: Integer) begin
			PrintLn(x);
		end);
	`

	_, interp := runLambdaTest(t, input)

	// Check output (via interp.output buffer if needed)
	// For now, just verify it doesn't error
	if interp.exception != nil {
		t.Errorf("Unexpected exception: %v", interp.exception)
	}
}

func TestChainedHigherOrderFunctions(t *testing.T) {
	// Map then Filter then Reduce
	input := `
		type TIntArray = array[0..5] of Integer;
		var numbers: TIntArray;
		numbers[0] := 1;
		numbers[1] := 2;
		numbers[2] := 3;
		numbers[3] := 4;
		numbers[4] := 5;

		// Double all numbers
		var doubled := Map(numbers, lambda(x: Integer): Integer => x * 2);

		// Keep only those > 4
		var filtered := Filter(doubled, lambda(x: Integer): Boolean => x > 4);

		// Sum them
		var sum := Reduce(filtered, lambda(acc: Integer, x: Integer): Integer => acc + x, 0);
	`

	_, interp := runLambdaTest(t, input)

	sumVal, _ := interp.env.Get("sum")
	intVal := sumVal.(*IntegerValue)

	// doubled = [2, 4, 6, 8, 10]
	// filtered = [6, 8, 10]
	// sum = 6+8+10 = 24
	if intVal.Value != 24 {
		t.Errorf("Expected sum 24, got %d", intVal.Value)
	}
}

func TestMapWithFunctionPointer(t *testing.T) {
	// Test that Map works with regular function pointers, not just lambdas
	input := `
		function double(x: Integer): Integer;
		begin
			Result := x * 2;
		end;

		type TIntArray = array[0..3] of Integer;
		var numbers: TIntArray;
		numbers[0] := 1;
		numbers[1] := 2;
		numbers[2] := 3;

		var doubled := Map(numbers, @double);
		var result := doubled[1];
	`

	_, interp := runLambdaTest(t, input)

	resultVal, _ := interp.env.Get("result")
	intVal := resultVal.(*IntegerValue)

	// doubled[1] = 2 * 2 = 4
	if intVal.Value != 4 {
		t.Errorf("Expected 4, got %d", intVal.Value)
	}
}

func TestHigherOrderFunctionWithStringArray(t *testing.T) {
	input := `
		type TStringArray = array[0..3] of String;
		var words: TStringArray;
		words[0] := 'hello';
		words[1] := 'world';
		words[2] := 'test';

		var lengths := Map(words, lambda(s: String): Integer => Length(s));

		var totalLength := Reduce(lengths, lambda(acc: Integer, x: Integer): Integer => acc + x, 0);
	`

	_, interp := runLambdaTest(t, input)

	totalVal, _ := interp.env.Get("totalLength")
	intVal := totalVal.(*IntegerValue)

	// lengths = [5, 5, 4], totalLength = 14
	if intVal.Value != 14 {
		t.Errorf("Expected totalLength 14, got %d", intVal.Value)
	}
}
