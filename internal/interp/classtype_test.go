package interp

import (
	"testing"
)

// TestClassTypeProperty tests the ClassType property on object instances
func TestClassTypeProperty(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "basic object instance ClassType",
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
	PrintLn(obj.ClassType.ClassName);
end.`,
			expected: "TBase\n",
		},
		{
			name: "derived class ClassType",
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
	PrintLn(obj.ClassType.ClassName);
end.`,
			expected: "TDerived\n",
		},
		{
			name: "ClassType in constructor",
			input: `
type TExample = class
	constructor Create;
end;

constructor TExample.Create;
begin
	PrintLn(ClassType.ClassName);
end;

var obj: TExample;
begin
	obj := TExample.Create;
end.`,
			expected: "TExample\n",
		},
		{
			name: "ClassType in instance method",
			input: `
type TExample = class
	constructor Create;
	procedure ShowType;
end;

constructor TExample.Create;
begin
end;

procedure TExample.ShowType;
begin
	PrintLn(ClassType.ClassName);
end;

var obj: TExample;
begin
	obj := TExample.Create;
	obj.ShowType;
end.`,
			expected: "TExample\n",
		},
		{
			name: "ClassType case insensitive",
			input: `
type TExample = class
	constructor Create;
end;

constructor TExample.Create;
begin
end;

var obj: TExample;
begin
	obj := TExample.Create;
	PrintLn(obj.classtype.ClassName);
	PrintLn(obj.CLASSTYPE.ClassName);
	PrintLn(obj.ClassType.ClassName);
end.`,
			expected: "TExample\nTExample\nTExample\n",
		},
		{
			name: "ClassType on TClass variable",
			input: `
type TExample = class
	constructor Create;
end;

constructor TExample.Create;
begin
end;

var meta: class of TExample;
begin
	meta := TExample;
	PrintLn(meta.ClassType.ClassName);
end.`,
			expected: "TExample\n",
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

// TestClassTypeInheritance tests ClassType with inheritance
func TestClassTypeInheritance(t *testing.T) {
	input := `
type TAnimal = class
	constructor Create;
	procedure Show;
end;

type TDog = class(TAnimal)
end;

type TCat = class(TAnimal)
end;

constructor TAnimal.Create;
begin
end;

procedure TAnimal.Show;
begin
	PrintLn(ClassType.ClassName);
end;

var a: TAnimal;
var d: TDog;
var c: TCat;
begin
	a := TAnimal.Create;
	a.Show;

	d := TDog.Create;
	d.Show;

	c := TCat.Create;
	c.Show;

	a := d;
	a.Show;

	a := c;
	a.Show;
end.`

	expected := "TAnimal\nTDog\nTCat\nTDog\nTCat\n"
	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	if output != expected {
		t.Errorf("Expected output:\n%s\n\nGot:\n%s", expected, output)
	}
}

// TestClassTypeComparison tests comparing ClassType values
func TestClassTypeComparison(t *testing.T) {
	input := `
type TBase = class
	constructor Create;
end;

type TDerived = class(TBase)
end;

constructor TBase.Create;
begin
end;

var b: TBase;
var d: TDerived;
begin
	b := TBase.Create;
	d := TDerived.Create;

	if b.ClassType = TBase then
		PrintLn('b is TBase');

	if d.ClassType = TDerived then
		PrintLn('d is TDerived');

	if d.ClassType = TBase then
		PrintLn('d is TBase (wrong)')
	else
		PrintLn('d is not TBase');

	b := d;
	if b.ClassType = TDerived then
		PrintLn('b (pointing to TDerived) has ClassType TDerived');
end.`

	expected := "b is TBase\nd is TDerived\nd is not TBase\nb (pointing to TDerived) has ClassType TDerived\n"
	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	if output != expected {
		t.Errorf("Expected output:\n%s\n\nGot:\n%s", expected, output)
	}
}

// TestClassTypeWithClassMethods tests ClassType in class methods
func TestClassTypeWithClassMethods(t *testing.T) {
	input := `
type TExample = class
	class procedure ShowType;
end;

class procedure TExample.ShowType;
begin
	PrintLn(ClassType.ClassName);
end;

begin
	TExample.ShowType;
end.`

	expected := "TExample\n"
	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	if output != expected {
		t.Errorf("Expected output:\n%s\n\nGot:\n%s", expected, output)
	}
}

// TestClassTypePropertyChaining tests chaining ClassType access
func TestClassTypePropertyChaining(t *testing.T) {
	input := `
type TExample = class
	constructor Create;
end;

constructor TExample.Create;
begin
end;

var obj: TExample;
begin
	obj := TExample.Create;
	PrintLn(obj.ClassType.ClassType.ClassName);
end.`

	// ClassType returns a ClassValue, which also has ClassType property
	// obj.ClassType returns TExample metaclass
	// obj.ClassType.ClassType should also return TExample metaclass
	expected := "TExample\n"
	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	if output != expected {
		t.Errorf("Expected output:\n%s\n\nGot:\n%s", expected, output)
	}
}
