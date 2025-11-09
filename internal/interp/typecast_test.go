package interp

import (
	"strings"
	"testing"
)

// TestBasicTypeCasts tests basic type conversions (Integer, Float, String, Boolean)
func TestBasicTypeCasts(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Integer to Integer",
			input: `
var i: Integer;
begin
	i := Integer(42);
	PrintLn(i);
end.`,
			expected: "42\n",
		},
		{
			name: "Float to Integer",
			input: `
var i: Integer;
var f: Float;
begin
	f := 3.14;
	i := Integer(f);
	PrintLn(i);
end.`,
			expected: "3\n",
		},
		{
			name: "Integer to Float",
			input: `
var f: Float;
var i: Integer;
begin
	i := 42;
	f := Float(i);
	PrintLn(f);
end.`,
			expected: "42\n",
		},
		{
			name: "Float to Float",
			input: `
var f: Float;
begin
	f := Float(1.25);
	PrintLn(f);
end.`,
			expected: "1.25\n",
		},
		{
			name: "Integer to String",
			input: `
var s: String;
begin
	s := String(42);
	PrintLn(s);
end.`,
			expected: "42\n",
		},
		{
			name: "Float to String",
			input: `
var s: String;
begin
	s := String(3.14);
	PrintLn(s);
end.`,
			expected: "3.14\n",
		},
		{
			name: "Boolean to Integer - true",
			input: `
var i: Integer;
begin
	i := Integer(True);
	PrintLn(i);
end.`,
			expected: "1\n",
		},
		{
			name: "Boolean to Integer - false",
			input: `
var i: Integer;
begin
	i := Integer(False);
	PrintLn(i);
end.`,
			expected: "0\n",
		},
		{
			name: "Integer to Boolean - non-zero",
			input: `
var b: Boolean;
begin
	b := Boolean(42);
	PrintLn(b);
end.`,
			expected: "True\n",
		},
		{
			name: "Integer to Boolean - zero",
			input: `
var b: Boolean;
begin
	b := Boolean(0);
	PrintLn(b);
end.`,
			expected: "False\n",
		},
		{
			name: "Float to Boolean - non-zero",
			input: `
var b: Boolean;
begin
	b := Boolean(3.14);
	PrintLn(b);
end.`,
			expected: "True\n",
		},
		{
			name: "Float to Boolean - zero",
			input: `
var b: Boolean;
begin
	b := Boolean(0.0);
	PrintLn(b);
end.`,
			expected: "False\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output := testEvalWithOutput(tt.input)
			if isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}
			if output != tt.expected {
				t.Errorf("Expected output:\n%s\n\nGot:\n%s", tt.expected, output)
			}
		})
	}
}

// TestClassTypeCasts tests class type casts
func TestClassTypeCasts(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Cast to same type",
			input: `
type TBase = class
	constructor Create;
end;

constructor TBase.Create;
begin
end;

var obj: TBase;
begin
	obj := TBase.Create;
	PrintLn(TBase(obj).ClassName);
end.`,
			expected: "TBase\n",
		},
		{
			name: "Upcast derived to base",
			input: `
type TBase = class
	constructor Create;
end;

type TDerived = class(TBase)
end;

constructor TBase.Create;
begin
end;

var obj: TDerived;
begin
	obj := TDerived.Create;
	PrintLn(TBase(obj).ClassName);
end.`,
			expected: "TDerived\n",
		},
		{
			name: "Downcast base to derived - valid at runtime",
			input: `
type TBase = class
	constructor Create;
end;

type TDerived = class(TBase)
end;

constructor TBase.Create;
begin
end;

var obj: TBase;
begin
	obj := TDerived.Create;
	PrintLn(TDerived(obj).ClassName);
end.`,
			expected: "TDerived\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output := testEvalWithOutput(tt.input)
			if isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}
			if output != tt.expected {
				t.Errorf("Expected output:\n%s\n\nGot:\n%s", tt.expected, output)
			}
		})
	}
}

// TestClassTypeCastFailures tests invalid class type casts
func TestClassTypeCastFailures(t *testing.T) {
	input := `
type TBase = class
	constructor Create;
end;

type TDerived = class(TBase)
end;

constructor TBase.Create;
begin
end;

var obj: TBase;
begin
	obj := TBase.Create;
	PrintLn(TDerived(obj).ClassName);
end.`

	result, _ := testEvalWithOutput(input)
	// Should return an error because TBase instance cannot be cast to TDerived
	if !isError(result) {
		t.Fatalf("Expected error for invalid cast, but got success")
	}

	// Check that the error message mentions incompatible types
	errMsg := result.String()
	if !strings.Contains(errMsg, "incompatible") && !strings.Contains(errMsg, "cannot cast") {
		t.Errorf("Expected error about incompatible types, got: %s", errMsg)
	}
}

// TestTypeCastInExpression tests type casts used in expressions
func TestTypeCastInExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Integer cast in arithmetic",
			input: `
var f: Float;
var i: Integer;
begin
	f := 3.7;
	i := Integer(f) + 5;
	PrintLn(i);
end.`,
			expected: "8\n",
		},
		{
			name: "Float cast in arithmetic",
			input: `
var i: Integer;
var result: Float;
begin
	i := 10;
	result := Float(i) / 3.0;
	PrintLn(result);
end.`,
			expected: "3.3333333333333335\n",
		},
		{
			name: "Boolean cast in condition",
			input: `
var i: Integer;
begin
	i := 5;
	if Boolean(i) then
		PrintLn('Non-zero')
	else
		PrintLn('Zero');
end.`,
			expected: "Non-zero\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output := testEvalWithOutput(tt.input)
			if isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}
			if output != tt.expected {
				t.Errorf("Expected output:\n%s\n\nGot:\n%s", tt.expected, output)
			}
		})
	}
}

// TestTypeCastCaseInsensitive tests that type names in casts are case-insensitive
func TestTypeCastCaseInsensitive(t *testing.T) {
	input := `
var i: Integer;
var f: Float;
begin
	f := 3.14;
	i := INTEGER(f);
	PrintLn(i);
	i := integer(f);
	PrintLn(i);
	i := Integer(f);
	PrintLn(i);
end.`

	expected := "3\n3\n3\n"
	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	if output != expected {
		t.Errorf("Expected output:\n%s\n\nGot:\n%s", expected, output)
	}
}
