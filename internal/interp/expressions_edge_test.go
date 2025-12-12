package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
)

// testEvalEdgeCase is a helper for testing edge cases
func testEvalEdgeCase(input string, t *testing.T) (Value, string) {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		return nil, ""
	}

	analyzer := semantic.NewAnalyzer()
	_ = analyzer.Analyze(program)

	var buf bytes.Buffer
	interp := New(&buf)
	if semanticInfo := analyzer.GetSemanticInfo(); semanticInfo != nil {
		interp.SetSemanticInfo(semanticInfo)
	}
	val := interp.Eval(program)
	return val, buf.String()
}

// TestEvalAddressOfMethod tests creating method pointers
func TestEvalAddressOfMethod(t *testing.T) {
	input := `
		type TTest = class
			Value: Integer;
			function GetValue: Integer;
		end;
		function TTest.GetValue: Integer;
		begin
			Result := Value;
		end;
		var obj := TTest.Create;
		obj.Value := 99;
		var fp := @obj.GetValue;
		PrintLn(fp());
	`
	_, output := testEvalEdgeCase(input, t)
	if !strings.Contains(output, "99") {
		t.Errorf("Expected output containing 99, got %q", output)
	}
}

// TestEvalLambdaWithClosure tests lambda capturing outer variables
func TestEvalLambdaWithClosure(t *testing.T) {
	input := `
		var factor := 10;
		var multiply := lambda(x: Integer): Integer begin Result := x * factor; end;
		PrintLn(multiply(5));
	`
	_, output := testEvalEdgeCase(input, t)
	if !strings.Contains(output, "50") {
		t.Errorf("Expected output containing 50, got %q", output)
	}
}

// TestEvalVariantComparisons tests variant equality edge cases
func TestEvalVariantComparisons(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "variant unassigned equals zero",
			input: "var v: Variant; PrintLn(v = 0);",
			want:  "True",
		},
		{
			name:  "variant unassigned equals empty string",
			input: "var v: Variant; PrintLn(v = '');",
			want:  "True",
		},
		{
			name:  "variant unassigned equals false",
			input: "var v: Variant; PrintLn(v = False);",
			want:  "True",
		},
		{
			name:  "variant with value not equals unassigned",
			input: "var v: Variant := 5; PrintLn(v = 0);",
			want:  "False",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalEdgeCase(tt.input, t)
			if !strings.Contains(output, tt.want) {
				t.Errorf("Expected output containing %q, got %q", tt.want, output)
			}
		})
	}
}

// TestEvalVariantAsCasts verifies using 'as' with Variant to cast to primitives.
func TestEvalVariantAsCasts(t *testing.T) {
	input := `
		var v: Variant;
		v := 'abcd';
		PrintLn(v as String);
		v := 5;
		PrintLn((v as Integer) + 1);
		v := True;
		PrintLn(v as Boolean);
		v := False;
		PrintLn(v as Integer);
	`
	_, output := testEvalEdgeCase(input, t)
	expected := "abcd\n6\nTrue\n0\n"
	if output != expected {
		t.Fatalf("expected output:\n%s\ngot:\n%s", expected, output)
	}
}

// TestEvalEnumComparisons tests enum comparison operators
func TestEvalEnumComparisons(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name: "enum less than or equal",
			input: `
				type TColor = (Red, Green, Blue);
				var c1: TColor := Red;
				var c2: TColor := Green;
				PrintLn(c1 <= c2);
			`,
			want: "True",
		},
		{
			name: "enum greater than or equal",
			input: `
				type TColor = (Red, Green, Blue);
				var c1: TColor := Blue;
				var c2: TColor := Green;
				PrintLn(c1 >= c2);
			`,
			want: "True",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalEdgeCase(tt.input, t)
			if !strings.Contains(output, tt.want) {
				t.Errorf("Expected output containing %q, got %q", tt.want, output)
			}
		})
	}
}

// TestEvalRecordComparisons tests record equality
func TestEvalRecordComparisons(t *testing.T) {
	input := `
		type TPoint = record
			X, Y: Integer;
		end;
		var p1: TPoint;
		p1.X := 10;
		p1.Y := 20;
		var p2: TPoint;
		p2.X := 10;
		p2.Y := 20;
		PrintLn(p1 = p2);
	`
	_, output := testEvalEdgeCase(input, t)
	if !strings.Contains(output, "True") {
		t.Errorf("Expected output containing True, got %q", output)
	}
}

// TestEvalInterfaceComparisons tests interface comparisons
func TestEvalInterfaceComparisons(t *testing.T) {
	input := `
		type ITest = interface
			procedure Test;
		end;
		type TTest = class(ITest)
			procedure Test;
		end;
		procedure TTest.Test;
		begin
		end;
		var obj1: TTest := TTest.Create;
		var obj2: TTest := obj1;
		var i1: ITest := obj1 as ITest;
		var i2: ITest := obj2 as ITest;
		PrintLn(i1 = i2);
	`
	_, output := testEvalEdgeCase(input, t)
	if !strings.Contains(output, "True") {
		t.Errorf("Expected output containing True, got %q", output)
	}
}

// TestEvalClassValueComparisons tests ClassValue (metaclass) comparisons
func TestEvalClassValueComparisons(t *testing.T) {
	input := `
		type TBase = class
		end;
		type TChild = class(TBase)
		end;
		var c1: class of TBase;
		var c2: class of TBase;
		c1 := TBase;
		c2 := TBase;
		PrintLn(c1 = c2);
	`
	_, output := testEvalEdgeCase(input, t)
	if !strings.Contains(output, "True") {
		t.Errorf("Expected output containing True, got %q", output)
	}
}

// TestEvalRTTITypeInfoComparisons tests TypeOf result comparisons
func TestEvalRTTITypeInfoComparisons(t *testing.T) {
	input := `
		type TTest = class
		end;
		var obj1 := TTest.Create;
		var obj2 := TTest.Create;
		PrintLn(TypeOf(obj1) = TypeOf(obj2));
	`
	_, output := testEvalEdgeCase(input, t)
	if !strings.Contains(output, "True") {
		t.Errorf("Expected output containing True, got %q", output)
	}
}

// TestEvalStringRTTIConcatenation tests string + RTTI_TYPEINFO concatenation
func TestEvalStringRTTIConcatenation(t *testing.T) {
	input := `
		type TTest = class
		end;
		var obj := TTest.Create;
		PrintLn('Type: ' + TypeOf(obj));
	`
	_, output := testEvalEdgeCase(input, t)
	if !strings.Contains(output, "Type:") {
		t.Errorf("Expected output containing 'Type:', got %q", output)
	}
}

// TestEvalVariantMixedTypeOperations tests variant operations with mixed types
func TestEvalVariantMixedTypeOperations(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "variant int + float",
			input: "var a: Variant := 5; var b: Variant := 2.5; PrintLn(a + b);",
			want:  "7.5",
		},
		{
			name:  "variant string + int",
			input: "var a: Variant := 'Value: '; var b: Variant := 42; PrintLn(a + b);",
			want:  "Value: 42",
		},
		{
			name:  "variant boolean operations",
			input: "var a: Variant := True; var b: Variant := False; PrintLn(a and b);",
			want:  "False",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalEdgeCase(tt.input, t)
			if !strings.Contains(output, tt.want) {
				t.Errorf("Expected output containing %q, got %q", tt.want, output)
			}
		})
	}
}

// TestEvalIsExpressionWithInterface tests 'is' operator with interfaces
func TestEvalIsExpressionWithInterface(t *testing.T) {
	input := `
		type ITest = interface
			procedure Test;
		end;
		type TTest = class(ITest)
			procedure Test;
		end;
		procedure TTest.Test;
		begin
		end;
		var obj: TTest := TTest.Create;
		PrintLn(obj is ITest);
	`
	_, output := testEvalEdgeCase(input, t)
	if !strings.Contains(output, "True") {
		t.Errorf("Expected output containing True, got %q", output)
	}
}

// TestEvalAsExpressionInterfaceToInterface tests interface to interface casting
func TestEvalAsExpressionInterfaceToInterface(t *testing.T) {
	input := `
		type IBase = interface
			procedure Base;
		end;
		type IChild = interface(IBase)
			procedure Child;
		end;
		type TTest = class(IChild)
			procedure Base;
			procedure Child;
		end;
		procedure TTest.Base;
		begin
		end;
		procedure TTest.Child;
		begin
		end;
		var obj: TTest := TTest.Create;
		var i1: IChild := obj as IChild;
		var i2: IBase := i1 as IBase;
		PrintLn('Success');
	`
	_, output := testEvalEdgeCase(input, t)
	if !strings.Contains(output, "Success") {
		t.Errorf("Expected output containing 'Success', got %q", output)
	}
}

func TestEvalIsExpressionInterfaceInheritance(t *testing.T) {
	input := `
		type IBase = interface
			procedure Base;
		end;
		type IChild = interface(IBase)
			procedure Child;
		end;
		type TTest = class(IChild)
			procedure Base;
			procedure Child;
		end;
		procedure TTest.Base;
		begin
		end;
		procedure TTest.Child;
		begin
		end;
		var obj: TTest := TTest.Create;
		PrintLn(obj is IChild);
		PrintLn(obj is IBase);
	`

	_, output := testEvalEdgeCase(input, t)
	if !strings.Contains(output, "True\nTrue") {
		t.Errorf("Expected both interface checks to be True, got %q", output)
	}
}

// TestEvalImplementsWithClassValue tests 'implements' with class type value
func TestEvalImplementsWithClassValue(t *testing.T) {
	input := `
		type ITest = interface
			procedure Test;
		end;
		type TTest = class(ITest)
			procedure Test;
		end;
		procedure TTest.Test;
		begin
		end;
		var cls: class of TTest;
		cls := TTest;
		PrintLn(cls implements ITest);
	`
	_, output := testEvalEdgeCase(input, t)
	if !strings.Contains(output, "True") {
		t.Errorf("Expected output containing True, got %q", output)
	}
}

// TestEvalIfExpressionDefaultValues tests if expression default values for different types
func TestEvalIfExpressionDefaultValues(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "default float",
			input: "var x: Float := if False then 5.5; PrintLn(x);",
			want:  "0",
		},
		{
			name:  "default string",
			input: "var x: String := if False then 'test'; PrintLn(x);",
			want:  "",
		},
		{
			name:  "default boolean",
			input: "var x: Boolean := if False then True; PrintLn(x);",
			want:  "False",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalEdgeCase(tt.input, t)
			if !strings.Contains(output, tt.want) {
				t.Errorf("Expected output containing %q, got %q", tt.want, output)
			}
		})
	}
}

// TestEvalCoalesceWithArrays tests coalesce with empty arrays
func TestEvalCoalesceWithArrays(t *testing.T) {
	input := `
		var arr: array of Integer;
		var result := arr ?? [1, 2, 3];
		PrintLn(Length(result));
	`
	_, output := testEvalEdgeCase(input, t)
	if !strings.Contains(output, "3") {
		t.Errorf("Expected output containing 3, got %q", output)
	}
}

// TestEvalUnaryPlusOnVariant tests unary plus on variant
func TestEvalUnaryPlusOnVariant(t *testing.T) {
	input := `
		var v: Variant := 42;
		PrintLn(+v);
	`
	_, output := testEvalEdgeCase(input, t)
	if !strings.Contains(output, "42") {
		t.Errorf("Expected output containing 42, got %q", output)
	}
}

// TestEvalNotOnVariant tests not operator on variant
func TestEvalNotOnVariant(t *testing.T) {
	input := `
		var v: Variant := True;
		var result := not v;
		PrintLn(result);
	`
	_, output := testEvalEdgeCase(input, t)
	if !strings.Contains(output, "False") {
		t.Errorf("Expected output containing False, got %q", output)
	}
}

// TestEvalInOperatorArray tests 'in' operator with arrays
func TestEvalInOperatorArray(t *testing.T) {
	input := `
		var arr := [10, 20, 30];
		PrintLn(20 in arr);
		PrintLn(99 in arr);
	`
	_, output := testEvalEdgeCase(input, t)
	if !strings.Contains(output, "True") || !strings.Contains(output, "False") {
		t.Errorf("Expected output containing True and False, got %q", output)
	}
}

// TestEvalInOperatorSubstring tests 'in' operator with substrings
func TestEvalInOperatorSubstring(t *testing.T) {
	input := `
		PrintLn('ab' in 'abcdef');
		PrintLn('xy' in 'abcdef');
	`
	_, output := testEvalEdgeCase(input, t)
	if !strings.Contains(output, "True") || !strings.Contains(output, "False") {
		t.Errorf("Expected output containing True and False, got %q", output)
	}
}

// TestEvalInOperatorEmptyString tests 'in' operator with empty string
func TestEvalInOperatorEmptyString(t *testing.T) {
	input := `
		PrintLn('' in 'abc');
	`
	_, output := testEvalEdgeCase(input, t)
	if !strings.Contains(output, "False") {
		t.Errorf("Expected output containing False, got %q", output)
	}
}

// TestEvalIsExpressionWithNonObject tests 'is' operator with non-object values
func TestEvalIsExpressionWithNonObject(t *testing.T) {
	input := `
		type TTest = class
		end;
		var x := 42;
		PrintLn(x is TTest);
	`
	_, output := testEvalEdgeCase(input, t)
	if !strings.Contains(output, "False") {
		t.Errorf("Expected output containing False, got %q", output)
	}
}

// TestEvalIsExpressionWithNil tests 'is' operator with nil objects
func TestEvalIsExpressionWithNil(t *testing.T) {
	input := `
		type TTest = class
		end;
		var obj: TTest := nil;
		PrintLn(obj is TTest);
	`
	_, output := testEvalEdgeCase(input, t)
	if !strings.Contains(output, "False") {
		t.Errorf("Expected output containing False, got %q", output)
	}
}

// TestEvalIsExpressionWithClassHierarchy tests 'is' operator with class inheritance
func TestEvalIsExpressionWithClassHierarchy(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "object is own class",
			input: `
				type TParent = class
				end;
				type TChild = class(TParent)
				end;
				var obj := TChild.Create;
				PrintLn(obj is TChild);
			`,
			expected: "True",
		},
		{
			name: "object is parent class",
			input: `
				type TParent = class
				end;
				type TChild = class(TParent)
				end;
				var obj := TChild.Create;
				PrintLn(obj is TParent);
			`,
			expected: "True",
		},
		{
			name: "parent object is not child class",
			input: `
				type TParent = class
				end;
				type TChild = class(TParent)
				end;
				var obj := TParent.Create;
				PrintLn(obj is TChild);
			`,
			expected: "False",
		},
		{
			name: "deep hierarchy - grandchild is grandparent",
			input: `
				type TGrandparent = class
				end;
				type TParent = class(TGrandparent)
				end;
				type TChild = class(TParent)
				end;
				var obj := TChild.Create;
				PrintLn(obj is TGrandparent);
			`,
			expected: "True",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalEdgeCase(tt.input, t)
			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected output containing %q, got %q", tt.expected, output)
			}
		})
	}
}

// TestEvalIsExpressionWithInterfaceHierarchy tests 'is' operator with interface inheritance
func TestEvalIsExpressionWithInterfaceHierarchy(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "object is implemented interface",
			input: `
				type IBase = interface
					procedure Base;
				end;
				type TImpl = class(IBase)
					procedure Base;
				end;
				procedure TImpl.Base; begin end;
				var obj := TImpl.Create;
				PrintLn(obj is IBase);
			`,
			expected: "True",
		},
		{
			name: "object is not non-implemented interface",
			input: `
				type IBase = interface
					procedure Base;
				end;
				type IOther = interface
					procedure Other;
				end;
				type TImpl = class(IBase)
					procedure Base;
				end;
				procedure TImpl.Base; begin end;
				var obj := TImpl.Create;
				PrintLn(obj is IOther);
			`,
			expected: "False",
		},
		{
			name: "object inherits interface from parent class",
			input: `
				type IBase = interface
					procedure Base;
				end;
				type TParent = class(IBase)
					procedure Base;
				end;
				procedure TParent.Base; begin end;
				type TChild = class(TParent)
				end;
				var obj := TChild.Create;
				PrintLn(obj is IBase);
			`,
			expected: "True",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalEdgeCase(tt.input, t)
			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected output containing %q, got %q", tt.expected, output)
			}
		})
	}
}

// TestEvalArithmeticShiftRight tests sar operator
func TestEvalArithmeticShiftRight(t *testing.T) {
	input := `
		PrintLn(-16 sar 2);
	`
	_, output := testEvalEdgeCase(input, t)
	if !strings.Contains(output, "-4") {
		t.Errorf("Expected output containing -4, got %q", output)
	}
}

// TestEvalFloatComparisons tests float comparison operators
func TestEvalFloatComparisons(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "float less or equal",
			input: "PrintLn(3.5 <= 5.5);",
			want:  "True",
		},
		{
			name:  "float greater or equal",
			input: "PrintLn(5.5 >= 3.5);",
			want:  "True",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalEdgeCase(tt.input, t)
			if !strings.Contains(output, tt.want) {
				t.Errorf("Expected output containing %q, got %q", tt.want, output)
			}
		})
	}
}

// TestEvalStringComparisons tests string comparison operators
func TestEvalStringComparisons(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "string less or equal",
			input: "PrintLn('abc' <= 'def');",
			want:  "True",
		},
		{
			name:  "string greater or equal",
			input: "PrintLn('def' >= 'abc');",
			want:  "True",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalEdgeCase(tt.input, t)
			if !strings.Contains(output, tt.want) {
				t.Errorf("Expected output containing %q, got %q", tt.want, output)
			}
		})
	}
}

// TestEvalBooleanEquality tests boolean equality operators
func TestEvalBooleanEquality(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "boolean equals",
			input: "PrintLn(True = True);",
			want:  "True",
		},
		{
			name:  "boolean not equals",
			input: "PrintLn(True <> False);",
			want:  "True",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalEdgeCase(tt.input, t)
			if !strings.Contains(output, tt.want) {
				t.Errorf("Expected output containing %q, got %q", tt.want, output)
			}
		})
	}
}

// TestEvalConvertToString tests convertToString with various types
func TestEvalConvertToString(t *testing.T) {
	var buf bytes.Buffer
	interp := New(&buf)

	tests := []struct {
		name     string
		value    Value
		expected string
	}{
		{"nil converts to empty", nil, ""},
		{"integer converts", &IntegerValue{Value: 123}, "123"},
		{"float converts", &FloatValue{Value: 4.56}, "4.56"},
		{"string converts", &StringValue{Value: "test"}, "test"},
		{"boolean true", &BooleanValue{Value: true}, "True"},
		{"boolean false", &BooleanValue{Value: false}, "False"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interp.convertToString(tt.value)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestEvalVariantArrayFalsey tests variant with empty array is falsey
func TestEvalVariantArrayFalsey(t *testing.T) {
	var emptyArr ArrayValue
	emptyArr.Elements = []Value{}
	variantWithArray := &VariantValue{Value: &emptyArr}

	result := evaluator.IsFalsey(variantWithArray)
	if !result {
		t.Errorf("Expected empty array variant to be falsey")
	}
}

// TestEvalUnassignedValueFalsey tests UnassignedValue is falsey
func TestEvalUnassignedValueFalsey(t *testing.T) {
	unassigned := &UnassignedValue{}
	result := evaluator.IsFalsey(unassigned)
	if !result {
		t.Errorf("Expected UnassignedValue to be falsey")
	}
}

// TestEvalNullValueFalsey tests NullValue is falsey
func TestEvalNullValueFalsey(t *testing.T) {
	null := &NullValue{}
	result := evaluator.IsFalsey(null)
	if !result {
		t.Errorf("Expected NullValue to be falsey")
	}
}

// TestEvalNonEmptyArrayTruthy tests non-empty array is truthy
func TestEvalNonEmptyArrayTruthy(t *testing.T) {
	arr := &ArrayValue{Elements: []Value{&IntegerValue{Value: 1}}}
	result := evaluator.IsFalsey(arr)
	if result {
		t.Errorf("Expected non-empty array to be truthy")
	}
}

// TestEvalObjectTruthy tests objects are truthy
func TestEvalObjectTruthy(t *testing.T) {
	obj := &ObjectInstance{Class: &ClassInfo{Name: "Test"}}
	result := evaluator.IsFalsey(obj)
	if result {
		t.Errorf("Expected object to be truthy")
	}
}
