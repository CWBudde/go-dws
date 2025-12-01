package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// Helper function to evaluate DWScript code and return the result
func testEvalSimple(input string) Value {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		panic("Parser errors: " + joinParserErrorsNewline(p.Errors()))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	return interp.Eval(program)
}

// Test basic object creation without constructor
func TestBasicObjectCreation(t *testing.T) {
	input := `
		type TPoint = class
			X: Integer;
			Y: Integer;
		end;

		var p: TPoint;
		p := TPoint.Create();
	`

	result := testEvalSimple(input)

	if isError(result) {
		t.Fatalf("Evaluation error: %s", result.String())
	}

	obj, ok := AsObject(result)
	if !ok {
		t.Fatalf("Expected object, got %s", result.Type())
	}

	if obj.Class.GetName() != "TPoint" {
		t.Errorf("Expected class name TPoint, got %s", obj.Class.GetName())
	}

	// Verify fields are initialized with default values
	xVal := obj.GetField("X")
	if xVal == nil {
		t.Fatal("Field X not initialized")
	}

	intX, ok := xVal.(*IntegerValue)
	if !ok {
		t.Fatalf("Expected IntegerValue for X, got %s", xVal.Type())
	}

	if intX.Value != 0 {
		t.Errorf("Expected X = 0, got %d", intX.Value)
	}
}

// Test method call without arguments
func TestSimpleMethodCall(t *testing.T) {
	input := `
		type TCalculator = class
			function Add(a: Integer; b: Integer): Integer;
			begin
				Result := a + b;
			end;
		end;

		var calc: TCalculator;
		calc := TCalculator.Create();
		calc.Add(5, 7);
	`

	result := testEvalSimple(input)

	if isError(result) {
		t.Fatalf("Evaluation error: %s", result.String())
	}

	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("Expected IntegerValue, got %s", result.Type())
	}

	if intVal.Value != 12 {
		t.Errorf("Expected 12, got %d", intVal.Value)
	}
}

// Test direct field access (reading)
func TestDirectFieldAccess(t *testing.T) {
	input := `
		type TBox = class
			Width: Integer;
			Height: Integer;
		end;

		var box: TBox;
		box := TBox.Create();
		box.Width;
	`

	result := testEvalSimple(input)

	if isError(result) {
		t.Fatalf("Evaluation error: %s", result.String())
	}

	intVal, ok := result.(*IntegerValue)
	if !ok {
		t.Fatalf("Expected IntegerValue, got %s", result.Type())
	}

	// Should be 0 (default value)
	if intVal.Value != 0 {
		t.Errorf("Expected 0, got %d", intVal.Value)
	}
}

// Test inheritance - child inherits parent methods
func TestBasicInheritance(t *testing.T) {
	input := `
		type TAnimal = class
			function Speak(): String;
			begin
				Result := 'Some sound';
			end;
		end;

		type TDog = class(TAnimal)
			function Speak(): String;
			begin
				Result := 'Woof!';
			end;
		end;

		var dog: TDog;
		dog := TDog.Create();
		dog.Speak();
	`

	result := testEvalSimple(input)

	if isError(result) {
		t.Fatalf("Evaluation error: %s", result.String())
	}

	strVal, ok := result.(*StringValue)
	if !ok {
		t.Fatalf("Expected StringValue, got %s", result.Type())
	}

	if strVal.Value != "Woof!" {
		t.Errorf("Expected 'Woof!', got '%s'", strVal.Value)
	}
}

// Test error: class not found
func TestClassNotFound(t *testing.T) {
	input := `
		var obj: TUnknown;
		obj := TUnknown.Create();
	`

	result := testEvalSimple(input)

	if !isError(result) {
		t.Fatalf("Expected error, got %s", result.String())
	}

	errVal, ok := result.(*ErrorValue)
	if !ok {
		t.Fatalf("Expected ErrorValue, got %T", result)
	}

	if !strings.Contains(errVal.Message, "class 'TUnknown' not found") {
		t.Errorf("Expected error about class not found, got '%s'", errVal.Message)
	}
}

// Test error: parent class not found
func TestParentClassNotFound(t *testing.T) {
	input := `
		type TDog = class(TAnimal)
			Name: String;
		end;
	`

	result := testEvalSimple(input)

	if !isError(result) {
		t.Fatalf("Expected error, got %s", result.String())
	}

	errVal, ok := result.(*ErrorValue)
	if !ok {
		t.Fatalf("Expected ErrorValue, got %T", result)
	}

	if !strings.Contains(errVal.Message, "parent class 'TAnimal' not found") {
		t.Errorf("Expected error about parent not found, got '%s'", errVal.Message)
	}
}

// Test error: method not found
func TestMethodNotFound(t *testing.T) {
	input := `
		type TSimple = class
			Value: Integer;
		end;

		var obj: TSimple;
		obj := TSimple.Create();
		obj.NonExistentMethod();
	`

	result := testEvalSimple(input)

	if !isError(result) {
		t.Fatalf("Expected error, got %s", result.String())
	}

	errVal, ok := result.(*ErrorValue)
	if !ok {
		t.Fatalf("Expected ErrorValue, got %T", result)
	}

	if !strings.Contains(errVal.Message, "method 'NonExistentMethod' not found") {
		t.Errorf("Expected error about method not found, got '%s'", errVal.Message)
	}
}

// Test error: field not found
func TestFieldNotFound(t *testing.T) {
	input := `
		type TSimple = class
			Value: Integer;
		end;

		var obj: TSimple;
		obj := TSimple.Create();
		obj.NonExistentField;
	`

	result := testEvalSimple(input)

	if !isError(result) {
		t.Fatalf("Expected error, got %s", result.String())
	}

	errVal, ok := result.(*ErrorValue)
	if !ok {
		t.Fatalf("Expected ErrorValue, got %T", result)
	}

	if !strings.Contains(errVal.Message, "field 'NonExistentField' not found") {
		t.Errorf("Expected error about field not found, got '%s'", errVal.Message)
	}
}

// Test error: Self used outside method
func TestSelfOutsideMethod(t *testing.T) {
	input := `
		type TSimple = class
			Value: Integer;
		end;

		var obj: TSimple;
		obj := Self;
	`

	result := testEvalSimple(input)

	if !isError(result) {
		t.Fatalf("Expected error, got %s", result.String())
	}

	errVal, ok := result.(*ErrorValue)
	if !ok {
		t.Fatalf("Expected ErrorValue, got %T", result)
	}

	if !strings.Contains(errVal.Message, "Self used outside method context") {
		t.Errorf("Expected error about Self outside method, got '%s'", errVal.Message)
	}
}
