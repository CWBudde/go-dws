package interp

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// runTest is a helper function to execute DWScript code and compare output
func runTest(t *testing.T, input, expected string) {
	t.Helper()
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("Evaluation error: %s", result.String())
	}

	output := buf.String()
	if output != expected {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}

// TestPropertyAccessClassVarCaseInsensitive tests that properties backed by class variables
// can be accessed with different casing
func TestPropertyAccessClassVarCaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Class var with lowercase property access",
			input: `
type
	TTest = class
	private
		class var FValue: Integer;
	public
		property Value: Integer read FValue;
	end;

begin
	TTest.FValue := 42;
	PrintLn(TTest.Value);
end.`,
			expected: "42\n",
		},
		{
			name: "Class var with uppercase in property definition",
			input: `
type
	TTest = class
	private
		class var fvalue: Integer;
	public
		property Value: Integer read FVALUE;
	end;

begin
	TTest.fvalue := 99;
	PrintLn(TTest.Value);
end.`,
			expected: "99\n",
		},
		{
			name: "Class var with mixed case variations",
			input: `
type
	TTest = class
	private
		class var FMyValue: Integer;
	public
		property Value: Integer read fmyvalue;
	end;

begin
	TTest.FMyValue := 123;
	PrintLn(TTest.Value);
end.`,
			expected: "123\n",
		},
		{
			name: "Instance property accessing class var via instance",
			input: `
type
	TTest = class
	private
		class var FShared: Integer;
	public
		property Shared: Integer read FShared;
	end;

var
	obj: TTest;
begin
	TTest.FShared := 77;
	obj := TTest.Create;
	PrintLn(obj.Shared);
end.`,
			expected: "77\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTest(t, tt.input, tt.expected)
		})
	}
}

// TestPropertyAccessConstantCaseInsensitive tests that properties backed by constants
// can be accessed with different casing
func TestPropertyAccessConstantCaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Constant with lowercase property access",
			input: `
type
	TTest = class
	private
		const CValue = 42;
	public
		property Value: Integer read CValue;
	end;

var
	obj: TTest;
begin
	obj := TTest.Create;
	PrintLn(obj.Value);
end.`,
			expected: "42\n",
		},
		{
			name: "Constant with uppercase in property definition",
			input: `
type
	TTest = class
	private
		const cvalue = 99;
	public
		property Value: Integer read CVALUE;
	end;

var
	obj: TTest;
begin
	obj := TTest.Create;
	PrintLn(obj.Value);
end.`,
			expected: "99\n",
		},
		{
			name: "Constant with mixed case variations",
			input: `
type
	TTest = class
	private
		const CMyValue = 123;
	public
		property Value: Integer read cmyvalue;
	end;

var
	obj: TTest;
begin
	obj := TTest.Create;
	PrintLn(obj.Value);
end.`,
			expected: "123\n",
		},
		{
			name: "Constant accessed via class name",
			input: `
type
	TTest = class
	private
		const CShared = 77;
	public
		property Shared: Integer read CShared;
	end;

begin
	PrintLn(TTest.Shared);
end.`,
			expected: "77\n",
		},
		{
			name: "String constant with case variation",
			input: `
type
	TTest = class
	private
		const CName = 'Hello';
	public
		property Name: String read cname;
	end;

var
	obj: TTest;
begin
	obj := TTest.Create;
	PrintLn(obj.Name);
end.`,
			expected: "Hello\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTest(t, tt.input, tt.expected)
		})
	}
}

// TestPropertyLookupOrder tests that the lookup order is correct:
// class vars -> constants -> fields -> methods
func TestPropertyLookupOrder(t *testing.T) {
	// Note: In DWScript, you cannot have a field and a constant/class var with the exact same name
	// in the same class - the semantic analyzer will reject it. So we test that each type of storage
	// works correctly rather than testing precedence when names collide.

	t.Run("Class var property access", func(t *testing.T) {
		input := `
type
	TTest = class
	private
		class var FValue: Integer;
	public
		property Value: Integer read FValue;
	end;

begin
	TTest.FValue := 50;
	PrintLn(TTest.Value);
end.`
		expected := "50\n"
		runTest(t, input, expected)
	})

	t.Run("Constant property access", func(t *testing.T) {
		input := `
type
	TTest = class
	private
		const CValue = 999;
	public
		property Value: Integer read CValue;
	end;

var
	obj: TTest;
begin
	obj := TTest.Create;
	PrintLn(obj.Value);
end.`
		expected := "999\n"
		runTest(t, input, expected)
	})

	t.Run("Field property access", func(t *testing.T) {
		input := `
type
	TTest = class
	private
		FValue: Integer;
	public
		property Value: Integer read FValue write FValue;
	end;

var
	obj: TTest;
begin
	obj := TTest.Create;
	obj.Value := 111;
	PrintLn(obj.Value);
end.`
		expected := "111\n"
		runTest(t, input, expected)
	})
}

// TestPropertyAccessInheritanceCaseInsensitive tests case-insensitive property access
// with inheritance
func TestPropertyAccessInheritanceCaseInsensitive(t *testing.T) {
	input := `
type TBase = class
private
	class var FBaseValue: Integer;
public
	property BaseValue: Integer read fbasevalue;
end;

type TDerived = class(TBase)
private
	class var FDerivedValue: Integer;
public
	property DerivedValue: Integer read FDERIVEDVALUE;
end;

begin
	TBase.FBaseValue := 10;
	TDerived.FDerivedValue := 20;
	PrintLn(TDerived.BaseValue);
	PrintLn(TDerived.DerivedValue);
end.`

	expected := "10\n20\n"
	runTest(t, input, expected)
}
