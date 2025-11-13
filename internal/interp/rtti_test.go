package interp

import (
	"testing"
)

// TestRTTITypeOf tests the TypeOf() built-in function.
// Task 9.25.1: TypeOf(value): TTypeInfo
func TestRTTITypeOf(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "TypeOf on Integer",
			input: `
begin
	var x: Integer := 42;
	var ti := TypeOf(x);
	PrintLn(ti);
end.`,
			expected: "Integer\n",
		},
		{
			name: "TypeOf on String",
			input: `
begin
	var s: String := "hello";
	var ti := TypeOf(s);
	PrintLn(ti);
end.`,
			expected: "String\n",
		},
		{
			name: "TypeOf on Float",
			input: `
begin
	var f: Float := 3.14;
	var ti := TypeOf(f);
	PrintLn(ti);
end.`,
			expected: "Float\n",
		},
		{
			name: "TypeOf on Boolean",
			input: `
begin
	var b: Boolean := True;
	var ti := TypeOf(b);
	PrintLn(ti);
end.`,
			expected: "Boolean\n",
		},
		{
			name: "TypeOf on object instance",
			input: `
type TMyClass = class
	constructor Create;
end;

constructor TMyClass.Create;
begin
end;

var obj: TMyClass;
begin
	obj := TMyClass.Create;
	var ti := TypeOf(obj);
	PrintLn(ti);
end.`,
			expected: "TMyClass\n",
		},
		{
			name: "TypeOf on class type",
			input: `
type TMyClass = class
end;

begin
	var ti := TypeOf(TMyClass);
	PrintLn(ti);
end.`,
			expected: "TMyClass\n",
		},
		{
			name: "TypeOf comparison same type",
			input: `
type TBase = class
end;

var a, b: TBase;
begin
	a := TBase.Create;
	b := TBase.Create;
	if TypeOf(a) = TypeOf(b) then
		PrintLn('Same type')
	else
		PrintLn('Different type');
end.`,
			expected: "Same type\n",
		},
		{
			name: "TypeOf with inheritance",
			input: `
type TBase = class
end;

type TDerived = class(TBase)
end;

var base: TBase;
var derived: TDerived;
begin
	base := TBase.Create;
	derived := TDerived.Create;

	PrintLn(TypeOf(base));
	PrintLn(TypeOf(derived));

	if TypeOf(base) = TypeOf(derived) then
		PrintLn('Same type')
	else
		PrintLn('Different type');
end.`,
			expected: "TBase\nTDerived\nDifferent type\n",
		},
		{
			name: "TypeOf on polymorphic variable",
			input: `
type TBase = class
end;

type TDerived = class(TBase)
end;

var obj: TBase;
begin
	obj := TDerived.Create;
	PrintLn(TypeOf(obj));
	PrintLn(obj.ClassName);
end.`,
			expected: "TDerived\nTDerived\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output := testEvalWithOutput(tt.input)
			if isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}
			if output != tt.expected {
				t.Errorf("expected output:\n%s\ngot:\n%s", tt.expected, output)
			}
		})
	}
}

// TestRTTITypeOfClass tests the TypeOfClass() built-in function.
// Task 9.25.2: TypeOfClass(classRef: TClass): TTypeInfo
func TestRTTITypeOfClass(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "TypeOfClass on class reference",
			input: `
type TMyClass = class
end;

var cls: TClass;
begin
	cls := TMyClass;
	var ti := TypeOfClass(cls);
	PrintLn(ti);
end.`,
			expected: "TMyClass\n",
		},
		{
			name: "TypeOfClass on ClassType property",
			input: `
type TBase = class
end;

type TDerived = class(TBase)
end;

var obj: TBase;
begin
	obj := TDerived.Create;
	var ti := TypeOfClass(obj.ClassType);
	PrintLn(ti);
end.`,
			expected: "TDerived\n",
		},
		{
			name: "TypeOfClass comparison",
			input: `
type TBase = class
end;

type TDerived = class(TBase)
end;

var base: TBase;
var derived: TDerived;
begin
	base := TBase.Create;
	derived := TDerived.Create;

	if TypeOfClass(base.ClassType) = TypeOfClass(derived.ClassType) then
		PrintLn('Same type')
	else
		PrintLn('Different type');
end.`,
			expected: "Different type\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output := testEvalWithOutput(tt.input)
			if isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}
			if output != tt.expected {
				t.Errorf("expected output:\n%s\ngot:\n%s", tt.expected, output)
			}
		})
	}
}

// TestRTTIClassName tests the ClassName property/method.
// Task 9.25.3: ClassName(obj: TObject): String
func TestRTTIClassName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "ClassName on object instance",
			input: `
type TMyClass = class
end;

var obj: TMyClass;
begin
	obj := TMyClass.Create;
	PrintLn(obj.ClassName);
end.`,
			expected: "TMyClass\n",
		},
		{
			name: "ClassName on class",
			input: `
type TMyClass = class
end;

begin
	PrintLn(TMyClass.ClassName);
end.`,
			expected: "TMyClass\n",
		},
		{
			name: "ClassName with inheritance",
			input: `
type TBase = class
end;

type TDerived = class(TBase)
end;

var obj: TBase;
begin
	obj := TDerived.Create;
	PrintLn(obj.ClassName);
	PrintLn(TBase.ClassName);
	PrintLn(TDerived.ClassName);
end.`,
			expected: "TDerived\nTBase\nTDerived\n",
		},
		{
			name: "ClassName in method",
			input: `
type TMyClass = class
	procedure PrintName;
end;

procedure TMyClass.PrintName;
begin
	PrintLn(ClassName);
end;

var obj: TMyClass;
begin
	obj := TMyClass.Create;
	obj.PrintName;
end.`,
			expected: "TMyClass\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output := testEvalWithOutput(tt.input)
			if isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}
			if output != tt.expected {
				t.Errorf("expected output:\n%s\ngot:\n%s", tt.expected, output)
			}
		})
	}
}

// TestRTTIClassType tests the ClassType property.
// Task 9.25.4: ClassType(obj: TObject): TClass
func TestRTTIClassType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "ClassType on object",
			input: `
type TMyClass = class
end;

var obj: TMyClass;
var cls: TClass;
begin
	obj := TMyClass.Create;
	cls := obj.ClassType;
	PrintLn(cls.ClassName);
end.`,
			expected: "TMyClass\n",
		},
		{
			name: "ClassType Create",
			input: `
type TBase = class
	constructor Create;
end;

constructor TBase.Create;
begin
	PrintLn('Creating ' + ClassName);
end;

var obj: TBase;
var cls: TClass;
begin
	obj := TBase.Create;
	cls := obj.ClassType;
	obj := cls.Create;
end.`,
			expected: "Creating TBase\nCreating TBase\n",
		},
		{
			name: "ClassType with polymorphism",
			input: `
type TBase = class
	constructor Create;
end;

type TDerived = class(TBase)
	constructor Create;
end;

constructor TBase.Create;
begin
	PrintLn('TBase.Create');
end;

constructor TDerived.Create;
begin
	PrintLn('TDerived.Create');
end;

var obj: TBase;
var cls: TClass;
begin
	obj := TDerived.Create;
	cls := obj.ClassType;
	PrintLn(cls.ClassName);
	obj := cls.Create;
end.`,
			expected: "TDerived.Create\nTDerived\nTDerived.Create\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output := testEvalWithOutput(tt.input)
			if isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}
			if output != tt.expected {
				t.Errorf("expected output:\n%s\ngot:\n%s", tt.expected, output)
			}
		})
	}
}

// TestRTTICombined tests combinations of RTTI features.
func TestRTTICombined(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "All RTTI features together",
			input: `
type TBase = class
	constructor Create;
end;

type TDerived = class(TBase)
	constructor Create;
end;

constructor TBase.Create;
begin
end;

constructor TDerived.Create;
begin
end;

var obj: TBase;
begin
	obj := TDerived.Create;

	// Test ClassName
	PrintLn('ClassName: ' + obj.ClassName);

	// Test ClassType
	var cls := obj.ClassType;
	PrintLn('ClassType.ClassName: ' + cls.ClassName);

	// Test TypeOf
	var ti := TypeOf(obj);
	PrintLn('TypeOf: ' + ti);

	// Test TypeOfClass
	var ti2 := TypeOfClass(cls);
	PrintLn('TypeOfClass: ' + ti2);

	// Test comparison
	if TypeOf(obj) = TypeOfClass(cls) then
		PrintLn('TypeOf and TypeOfClass match');
end.`,
			expected: "ClassName: TDerived\nClassType.ClassName: TDerived\nTypeOf: TDerived\nTypeOfClass: TDerived\nTypeOf and TypeOfClass match\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output := testEvalWithOutput(tt.input)
			if isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}
			if output != tt.expected {
				t.Errorf("expected output:\n%s\ngot:\n%s", tt.expected, output)
			}
		})
	}
}
