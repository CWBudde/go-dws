package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/lexer"
	"github.com/cwbudde/go-dws/parser"
)

// ============================================================================
// Property Interpreter Tests (Task 8.57)
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
