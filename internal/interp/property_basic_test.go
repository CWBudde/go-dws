package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// ============================================================================
// Property Interpreter Tests
// ============================================================================

// TestPropertyFieldBacked tests field-backed properties
func TestPropertyFieldBacked(t *testing.T) {
	input := `
type TTest = class
	FName: String;
	property Name: String read FName write FName;

	constructor Create; begin end;
end;

var obj := TTest.Create();
obj.Name := 'Hello';
PrintLn(obj.Name);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "Hello\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestPropertyMethodBacked tests method-backed properties
func TestPropertyMethodBacked(t *testing.T) {
	input := `
type TTest = class
	FCount: Integer;

	function GetCount: Integer;
	begin
		Result := FCount;
	end;

	procedure SetCount(value: Integer);
	begin
		FCount := value;
	end;

	property Count: Integer read GetCount write SetCount;

	constructor Create; begin end;
end;

var obj := TTest.Create();
obj.Count := 42;
PrintLn(obj.Count);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "42\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestPropertyReadOnly tests read-only properties
func TestPropertyReadOnly(t *testing.T) {
	input := `
type TTest = class
	FSize: Integer;
	property Size: Integer read FSize;

	constructor Create; begin end;
end;

var obj := TTest.Create();
obj.FSize := 100;
PrintLn(obj.Size);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "100\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestPropertyReadOnlyAssignmentError tests that writing to a read-only property fails
func TestPropertyReadOnlyAssignmentError(t *testing.T) {
	input := `
type TTest = class
	FSize: Integer;
	property Size: Integer read FSize;

	constructor Create; begin end;
end;

var obj := TTest.Create();
obj.Size := 100;
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	interp := New(nil)
	result := interp.Eval(program)

	// Should get an error
	if !isError(result) {
		t.Fatalf("expected error for read-only property assignment, got %v", result)
	}

	errorMsg := result.(*ErrorValue).Message
	// Check that the error message contains the expected text (it may have location info)
	expectedError := "property 'Size' is read-only"
	if !strings.Contains(errorMsg, expectedError) {
		t.Errorf("expected error containing '%s', got '%s'", expectedError, errorMsg)
	}
}

// TestPropertyInheritance tests that properties are inherited
func TestPropertyInheritance(t *testing.T) {
	input := `
type TBase = class
	FValue: Integer;
	property Value: Integer read FValue write FValue;

	constructor Create; begin end;
end;

type TDerived = class(TBase)
	FName: String;
	property Name: String read FName write FName;
end;

var obj := TDerived.Create();
obj.Value := 42;
obj.Name := 'Test';
PrintLn(obj.Value);
PrintLn(obj.Name);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "42\nTest\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

func TestPropertyIndexDirective(t *testing.T) {
	input := `
type TTest = class
	function GetProp(i: Integer): String; begin Result := 'Get ' + IntToStr(i); end;
	procedure SetProp(i: Integer; v: String); begin PrintLn('Set ' + IntToStr(i) + ' with ' + v); end;

	property Item1: String index 1 read GetProp write SetProp;
	property Item2: String index 2 read GetProp write SetProp;
end;

var o := TTest.Create;
PrintLn(o.Item1);
PrintLn(o.Item2);
o.Item1 := 'hello';
o.Item2 := 'world';
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	expected := "Get 1\nGet 2\nSet 1 with hello\nSet 2 with world\n"
	if buf.String() != expected {
		t.Errorf("expected output '%s', got '%s'", expected, buf.String())
	}
}

// TestPropertyAutoProperty tests auto-properties (with auto-generated backing fields)
func TestPropertyAutoProperty(t *testing.T) {
	input := `
type TTest = class
	FValue: Integer;
	property Value: Integer;

	constructor Create; begin end;
end;

var obj := TTest.Create();
obj.Value := 99;
PrintLn(obj.Value);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "99\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// ============================================================================
// Indexed Property Tests
// ============================================================================

// TestIndexedPropertyRead tests basic indexed property read functionality
func TestIndexedPropertyRead(t *testing.T) {
	input := `
type TArray = class
	FData: array of String;
	function GetItem(i: Integer): String; begin Result := FData[i]; end;
	property Items[i: Integer]: String read GetItem;
	constructor Create; begin FData := ['first', 'second', 'third']; end;
end;

var arr := TArray.Create();
PrintLn(arr.Items[0]);
PrintLn(arr.Items[1]);
PrintLn(arr.Items[2]);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "first\nsecond\nthird\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestIndexedPropertyWithVariableIndex tests indexed property with variable index
func TestIndexedPropertyWithVariableIndex(t *testing.T) {
	input := `
type TArray = class
	FData: array of Integer;
	function GetValue(i: Integer): Integer; begin Result := FData[i]; end;
	property Values[i: Integer]: Integer read GetValue;
	constructor Create; begin FData := [0, 10, 20, 30, 40]; end;
end;

var arr := TArray.Create();
var idx: Integer;
idx := 2;
PrintLn(arr.Values[idx]);
idx := 4;
PrintLn(arr.Values[idx]);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "20\n40\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestIndexedPropertyInheritance tests that indexed properties are inherited
func TestIndexedPropertyInheritance(t *testing.T) {
	input := `
type TBase = class
	FData: array of String;
	function GetItem(i: Integer): String; begin Result := FData[i]; end;
	property Items[i: Integer]: String read GetItem;
	constructor Create; begin FData := ['base0', 'base1']; end;
end;

type TDerived = class(TBase)
end;

var obj := TDerived.Create();
PrintLn(obj.Items[0]);
PrintLn(obj.Items[1]);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "base0\nbase1\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestIndexedPropertyWithoutIndexError tests error when indexed property accessed without index
func TestIndexedPropertyWithoutIndexError(t *testing.T) {
	input := `
type TArray = class
	FData: array of String;
	function GetItem(i: Integer): String; begin Result := FData[i]; end;
	property Items[i: Integer]: String read GetItem;
	constructor Create; begin end;
end;

var arr := TArray.Create();
PrintLn(arr.Items);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	interp := New(nil)
	result := interp.Eval(program)

	// Should get an error
	if !isError(result) {
		t.Fatalf("expected error for indexed property without index, got %v", result)
	}

	errorMsg := result.(*ErrorValue).Message
	// Check that the error message indicates the property requires an index
	if !strings.Contains(errorMsg, "indexed property") && !strings.Contains(errorMsg, "requires index") {
		t.Errorf("expected error about indexed property requiring index, got '%s'", errorMsg)
	}
}

// TestIndexedPropertyFieldBackedError tests that indexed properties cannot be field-backed
func TestIndexedPropertyFieldBackedError(t *testing.T) {
	input := `
type TArray = class
	FData: array of String;

	property Items[Index: Integer]: String read FData;

	constructor Create; begin end;
end;

var arr := TArray.Create();
PrintLn(arr.Items[0]);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	interp := New(nil)
	result := interp.Eval(program)

	// Should get an error
	if !isError(result) {
		t.Fatalf("expected error for field-backed indexed property, got %v", result)
	}

	errorMsg := result.(*ErrorValue).Message
	// Check that the error message indicates indexed properties require a method
	if !strings.Contains(errorMsg, "indexed property") && !strings.Contains(errorMsg, "getter method") {
		t.Errorf("expected error about indexed property requiring getter method, got '%s'", errorMsg)
	}
}

// ============================================================================
// Indexed Property Write Tests
// ============================================================================

// TestIndexedPropertyWrite tests basic indexed property write functionality
func TestIndexedPropertyWrite(t *testing.T) {
	input := `
type TArray = class
	FData: array of String;
	function GetItem(i: Integer): String; begin Result := FData[i]; end;
	procedure SetItem(i: Integer; v: String); begin FData[i] := v; end;
	property Items[i: Integer]: String read GetItem write SetItem;
	constructor Create; begin FData := ['a', 'b', 'c']; end;
end;

var arr := TArray.Create();
arr.Items[0] := 'first';
arr.Items[1] := 'second';
arr.Items[2] := 'third';
PrintLn(arr.Items[0]);
PrintLn(arr.Items[1]);
PrintLn(arr.Items[2]);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "first\nsecond\nthird\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestIndexedPropertyWriteWithVariableIndex tests indexed property write with variable index
func TestIndexedPropertyWriteWithVariableIndex(t *testing.T) {
	input := `
type TArray = class
	FData: array of Integer;
	function GetValue(i: Integer): Integer; begin Result := FData[i]; end;
	procedure SetValue(i: Integer; v: Integer); begin FData[i] := v; end;
	property Values[i: Integer]: Integer read GetValue write SetValue;
	constructor Create; begin FData := [0, 0, 0, 0, 0]; end;
end;

var arr := TArray.Create();
var idx: Integer;
idx := 2;
arr.Values[idx] := 42;
idx := 4;
arr.Values[idx] := 99;
PrintLn(arr.Values[2]);
PrintLn(arr.Values[4]);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "42\n99\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestIndexedPropertyWriteInheritance tests that indexed property writes work through inheritance
func TestIndexedPropertyWriteInheritance(t *testing.T) {
	input := `
type TBase = class
	FData: array of String;
	function GetItem(i: Integer): String; begin Result := FData[i]; end;
	procedure SetItem(i: Integer; v: String); begin FData[i] := v; end;
	property Items[i: Integer]: String read GetItem write SetItem;
	constructor Create; begin FData := ['', '']; end;
end;

type TDerived = class(TBase)
end;

var obj := TDerived.Create();
obj.Items[0] := 'derived0';
obj.Items[1] := 'derived1';
PrintLn(obj.Items[0]);
PrintLn(obj.Items[1]);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "derived0\nderived1\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestIndexedPropertyReadOnlyError tests error when writing to read-only indexed property
func TestIndexedPropertyReadOnlyError(t *testing.T) {
	input := `
type TArray = class
	FData: array of String;
	function GetItem(i: Integer): String; begin Result := FData[i]; end;
	property Items[i: Integer]: String read GetItem;
	constructor Create; begin FData := ['a', 'b']; end;
end;

var arr := TArray.Create();
arr.Items[0] := 'test';
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	interp := New(nil)
	result := interp.Eval(program)

	// Should get an error
	if !isError(result) {
		t.Fatalf("expected error for read-only indexed property write, got %v", result)
	}

	errorMsg := result.(*ErrorValue).Message
	// Check that the error message indicates the property is read-only
	if !strings.Contains(errorMsg, "read-only") {
		t.Errorf("expected error about read-only property, got '%s'", errorMsg)
	}
}

// TestIndexedPropertyWriteFieldBackedError tests that indexed property writes cannot use fields
func TestIndexedPropertyWriteFieldBackedError(t *testing.T) {
	input := `
type TArray = class
	FData: array of String;
	property Items[i: Integer]: String read FData write FData;
	constructor Create; begin end;
end;

var arr := TArray.Create();
arr.Items[0] := 'test';
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	interp := New(nil)
	result := interp.Eval(program)

	// Should get an error
	if !isError(result) {
		t.Fatalf("expected error for field-backed indexed property write, got %v", result)
	}

	errorMsg := result.(*ErrorValue).Message
	// Check that the error message indicates indexed properties require a setter method
	if !strings.Contains(errorMsg, "indexed property") && !strings.Contains(errorMsg, "setter method") {
		t.Errorf("expected error about indexed property requiring setter method, got '%s'", errorMsg)
	}
}

// ============================================================================
// Expression-Based Property Tests
// ============================================================================

// TestExpressionBasedPropertySimple tests basic expression-based property getters
func TestExpressionBasedPropertySimple(t *testing.T) {
	input := `
type TRectangle = class
	FWidth: Integer;
	FHeight: Integer;
	property Area: Integer read (FWidth * FHeight);

	constructor Create(w, h: Integer);
	begin
		FWidth := w;
		FHeight := h;
	end;
end;

var rect := TRectangle.Create(10, 5);
PrintLn(IntToStr(rect.Area));
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output - Area should be 10 * 5 = 50
	expectedOutput := "50\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestExpressionBasedPropertyComplex tests complex expression-based properties
func TestExpressionBasedPropertyComplex(t *testing.T) {
	input := `
type TRectangle = class
	FWidth: Integer;
	FHeight: Integer;
	property Perimeter: Integer read (2 * (FWidth + FHeight));

	constructor Create(w, h: Integer);
	begin
		FWidth := w;
		FHeight := h;
	end;
end;

var rect := TRectangle.Create(10, 5);
PrintLn(IntToStr(rect.Perimeter));
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output - Perimeter should be 2 * (10 + 5) = 30
	expectedOutput := "30\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestExpressionBasedPropertyWithFunctionCall tests expression with function calls
func TestExpressionBasedPropertyWithFunctionCall(t *testing.T) {
	input := `
type TDataHolder = class
	FCount: Integer;
	property Count: Integer read (FCount * 2);

	constructor Create;
	begin
		FCount := 5;
	end;
end;

var data := TDataHolder.Create();
PrintLn(IntToStr(data.Count));
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output - Count should be FCount * 2 = 10
	expectedOutput := "10\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestExpressionBasedPropertyRecalculates tests that expressions recalculate on each access
func TestExpressionBasedPropertyRecalculates(t *testing.T) {
	input := `
type TCounter = class
	FValue: Integer;
	property Doubled: Integer read (FValue * 2);

	constructor Create;
	begin
		FValue := 5;
	end;
end;

var counter := TCounter.Create();
PrintLn(IntToStr(counter.Doubled));
counter.FValue := 10;
PrintLn(IntToStr(counter.Doubled));
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output - First should be 10, then 20 after FValue changes
	expectedOutput := "10\n20\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestExpressionBasedPropertyWithDivision tests expression with integer division
func TestExpressionBasedPropertyWithDivision(t *testing.T) {
	input := `
type TTest = class
	FValue: Integer;
	property Half: Integer read (FValue div 2);

	constructor Create(v: Integer);
	begin
		FValue := v;
	end;
end;

var obj := TTest.Create(20);
PrintLn(IntToStr(obj.Half));
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output - Half should be 20 div 2 = 10
	expectedOutput := "10\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}
