package interp

import (
	"testing"
)

// ============================================================================
// Enum Declaration Tests
// ============================================================================

func TestEnumDeclaration(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name: "simple enum declaration",
			input: `
				type TColor = (Red, Green, Blue);
				PrintLn('ok');
			`,
			expect: "ok\n",
		},
		{
			name: "enum with explicit values",
			input: `
				type TStatus = (Ok = 0, Warning = 1, Error = 2);
				PrintLn('ok');
			`,
			expect: "ok\n",
		},
		{
			name: "enum with mixed values",
			input: `
				type TPriority = (Low, Medium = 5, High);
				PrintLn('ok');
			`,
			expect: "ok\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.input)
			if output != tt.expect {
				t.Errorf("expected %q, got %q", tt.expect, output)
			}
		})
	}
}

// ============================================================================
// Enum Value Tests
// ============================================================================

func TestEnumValueStorage(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name: "enum value in variable",
			input: `
				type TColor = (Red, Green, Blue);
				var color: TColor := Red;
				PrintLn('ok');
			`,
			expect: "ok\n",
		},
		{
			name: "enum value assignment",
			input: `
				type TColor = (Red, Green, Blue);
				var color: TColor;
				color := Green;
				PrintLn('ok');
			`,
			expect: "ok\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.input)
			if output != tt.expect {
				t.Errorf("expected %q, got %q", tt.expect, output)
			}
		})
	}
}

// ============================================================================
// Enum Literal Tests
// ============================================================================

func TestEnumLiteralEvaluation(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name: "unscoped enum reference",
			input: `
				type TColor = (Red, Green, Blue);
				var color: TColor := Red;
				PrintLn('ok');
			`,
			expect: "ok\n",
		},
		{
			name: "scoped enum reference",
			input: `
				type TColor = (Red, Green, Blue);
				var color: TColor := TColor.Green;
				PrintLn('ok');
			`,
			expect: "ok\n",
		},
		{
			name: "enum in case statement",
			input: `
				type TColor = (Red, Green, Blue);
				var color: TColor := Green;
				case color of
					Red: PrintLn('Red');
					Green: PrintLn('Green');
					Blue: PrintLn('Blue');
				end;
			`,
			expect: "Green\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.input)
			if output != tt.expect {
				t.Errorf("expected %q, got %q", tt.expect, output)
			}
		})
	}
}

// ============================================================================
// Ord() Function Tests
// ============================================================================

func TestOrdFunction(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name: "Ord of first enum value",
			input: `
				type TColor = (Red, Green, Blue);
				PrintLn(Ord(Red));
			`,
			expect: "0\n",
		},
		{
			name: "Ord of second enum value",
			input: `
				type TColor = (Red, Green, Blue);
				PrintLn(Ord(Green));
			`,
			expect: "1\n",
		},
		{
			name: "Ord with explicit values",
			input: `
				type TStatus = (Ok = 10, Warning = 20, Error = 30);
				PrintLn(Ord(Warning));
			`,
			expect: "20\n",
		},
		{
			name: "Ord with mixed values",
			input: `
				type TPriority = (Low, Medium = 5, High);
				PrintLn(Ord(High));
			`,
			expect: "6\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.input)
			if output != tt.expect {
				t.Errorf("expected %q, got %q", tt.expect, output)
			}
		})
	}
}

// ============================================================================
// Integer() Casting Tests
// ============================================================================

func TestIntegerCast(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name: "Integer cast of enum",
			input: `
				type TColor = (Red, Green, Blue);
				PrintLn(Integer(Red));
			`,
			expect: "0\n",
		},
		{
			name: "Integer cast with explicit value",
			input: `
				type TStatus = (Ok = 100);
				PrintLn(Integer(Ok));
			`,
			expect: "100\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.input)
			if output != tt.expect {
				t.Errorf("expected %q, got %q", tt.expect, output)
			}
		})
	}
}

// ============================================================================
// For-In Enum Type Iteration Tests
// ============================================================================

func TestForInEnumType_Basic(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name: "iterate small enum",
			input: `
				type TColor = (Red, Green, Blue);
				for var color in TColor do
					PrintLn(color);
			`,
			expect: "0\n1\n2\n",
		},
		{
			name: "iterate enum with two values",
			input: `
				type TBool = (False, True);
				for var b in TBool do
					PrintLn(b);
			`,
			expect: "0\n1\n",
		},
		{
			name: "iterate single value enum",
			input: `
				type TSingle = (OnlyOne);
				for var s in TSingle do
					PrintLn(s);
			`,
			expect: "0\n",
		},
		{
			name: "access ordinal value in loop",
			input: `
				type TColor = (Red, Green, Blue);
				for var color in TColor do
					PrintLn(Ord(color));
			`,
			expect: "0\n1\n2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.input)
			if output != tt.expect {
				t.Errorf("expected %q, got %q", tt.expect, output)
			}
		})
	}
}

func TestForInEnumType_ExplicitOrdinals(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name: "enum with explicit ordinals",
			input: `
				type TStatus = (Ok = 10, Warning = 20, Error = 30);
				for var s in TStatus do begin
					PrintLn(s);
					PrintLn(Ord(s));
				end;
			`,
			expect: "10\n10\n20\n20\n30\n30\n",
		},
		{
			name: "enum with mixed ordinals",
			input: `
				type TPriority = (Low, Medium = 5, High);
				for var p in TPriority do begin
					PrintLn(p);
					PrintLn(Ord(p));
				end;
			`,
			expect: "0\n0\n5\n5\n6\n6\n",
		},
		{
			name: "enum with gaps in ordinals",
			input: `
				type TLevel = (None = 0, Low = 10, High = 100);
				var count := 0;
				for var level in TLevel do
					count := count + 1;
				PrintLn(count);
			`,
			expect: "3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.input)
			if output != tt.expect {
				t.Errorf("expected %q, got %q", tt.expect, output)
			}
		})
	}
}

func TestForInEnumType_ControlFlow(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name: "break in enum for-in loop",
			input: `
				type TColor = (Red, Green, Blue, Yellow, Orange);
				for var color in TColor do begin
					PrintLn(color);
					if Ord(color) >= 2 then
						break;
				end;
			`,
			expect: "0\n1\n2\n",
		},
		{
			name: "continue in enum for-in loop",
			input: `
				type TColor = (Red, Green, Blue, Yellow);
				for var color in TColor do begin
					if Ord(color) = 1 then
						continue;
					PrintLn(color);
				end;
			`,
			expect: "0\n2\n3\n",
		},
		{
			name: "nested enum for-in loops",
			input: `
				type TFirst = (A, B);
				type TSecond = (X, Y);
				var count := 0;
				for var f in TFirst do
					for var s in TSecond do
						count := count + 1;
				PrintLn(count);
			`,
			expect: "4\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.input)
			if output != tt.expect {
				t.Errorf("expected %q, got %q", tt.expect, output)
			}
		})
	}
}

func TestForInEnumType_LargeEnum(t *testing.T) {
	// Test with an enum that has only two explicitly declared values
	// but with large ordinal values (like eratosthene.pas uses)
	input := `
		type TRange = enum (Low = 2, High = 1000);
		var count := 0;
		for var e in TRange do
			count := count + 1;
		PrintLn(count);
	`
	_, output := testEvalWithOutput(input)
	// TRange has only 2 values: Low (2) and High (1000)
	expect := "2\n"
	if output != expect {
		t.Errorf("expected %q, got %q", expect, output)
	}
}

func TestForInEnumType_OrdinalValues(t *testing.T) {
	// Test that iteration provides enum values with correct ordinals
	input := `
		type TRange = enum (Low = 2, High = 1000);
		for var e in TRange do
			PrintLn(Ord(e));
	`
	_, output := testEvalWithOutput(input)
	expect := "2\n1000\n"
	if output != expect {
		t.Errorf("expected %q, got %q", expect, output)
	}
}

func TestForInEnumType_ErrorCase(t *testing.T) {
	// Test error when trying to iterate over non-enum type metadata
	input := `
		for var i in Integer do
			PrintLn(i);
	`
	result, _ := testEvalWithOutput(input)
	if result == nil {
		t.Fatal("expected error result, got nil")
	}
	if err, ok := result.(*ErrorValue); ok {
		expected := "for-in loop: can only iterate over enum types, got Integer"
		if err.Message != expected {
			t.Errorf("expected error %q, got %q", expected, err.Message)
		}
	} else {
		t.Errorf("expected ErrorValue, got %T", result)
	}
}

// ============================================================================
// Enum .Value Helper Property Tests
// ============================================================================

func TestEnumValueProperty(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name: "enum value property with implicit ordinals",
			input: `
				type TColor = (Red, Green, Blue);
				var color := Red;
				PrintLn(color.Value);
			`,
			expect: "0\n",
		},
		{
			name: "enum value property with explicit ordinals",
			input: `
				type TStatus = (Ok = 100, Error = 200);
				PrintLn(Ok.Value);
				PrintLn(Error.Value);
			`,
			expect: "100\n200\n",
		},
		{
			name: "scoped enum value property",
			input: `
				type TColor = (Red, Green, Blue);
				PrintLn(TColor.Green.Value);
			`,
			expect: "1\n",
		},
		{
			name: "enum value property with chaining",
			input: `
				type TEnum = (eOne=1, eTwo, eThree);
				PrintLn(eTwo.Value.ToString);
			`,
			expect: "2\n",
		},
		{
			name: "enum value property in expression",
			input: `
				type TPriority = (Low, Medium = 5, High);
				var p := High;
				PrintLn(p.Value * 2);
			`,
			expect: "12\n",
		},
		{
			name: "enum value property vs Ord equivalence",
			input: `
				type TColor = (Red, Green, Blue);
				var c := Green;
				PrintLn(c.Value = Ord(c));
			`,
			expect: "True\n",
		},
		{
			name: "all values from enum with mixed ordinals",
			input: `
				type TEnum = (eOne=1, eTwo, eThree);
				PrintLn(eOne.Value);
				PrintLn(eTwo.Value);
				PrintLn(eThree.Value);
			`,
			expect: "1\n2\n3\n",
		},
		{
			name: "enum value property in variable assignment",
			input: `
				type TColor = (Red, Green, Blue);
				var c := Blue;
				var ordinal: Integer := c.Value;
				PrintLn(ordinal);
			`,
			expect: "2\n",
		},
		{
			name: "enum value property in comparison",
			input: `
				type TPriority = (Low = 1, Medium = 5, High = 10);
				var p1 := Low;
				var p2 := High;
				PrintLn(p1.Value < p2.Value);
			`,
			expect: "True\n",
		},
		{
			name: "enum value property with for-in loop",
			input: `
				type TColor = (Red, Green, Blue);
				for var c in TColor do
					PrintLn(c.Value);
			`,
			expect: "0\n1\n2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.input)
			if output != tt.expect {
				t.Errorf("expected %q, got %q", tt.expect, output)
			}
		})
	}
}

// ============================================================================
// Enum .Name Helper Property Tests
// ============================================================================

func TestEnumNameProperty(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name: "enum name property basic",
			input: `
				type TColor = (Red, Green, Blue);
				PrintLn(Red.Name);
				PrintLn(Green.Name);
				PrintLn(Blue.Name);
			`,
			expect: "Red\nGreen\nBlue\n",
		},
		{
			name: "enum name property with explicit ordinals",
			input: `
				type TStatus = (Ok = 100, Error = 200);
				PrintLn(Ok.Name);
				PrintLn(Error.Name);
			`,
			expect: "Ok\nError\n",
		},
		{
			name: "enum name property from variable",
			input: `
				type TColor = (Red, Green, Blue);
				var c := Green;
				PrintLn(c.Name);
			`,
			expect: "Green\n",
		},
		{
			name: "enum name property in for-in loop",
			input: `
				type TEnum = (enOne, enTwo, enThree);
				for var e in TEnum do
					PrintLn(e.Name);
			`,
			expect: "enOne\nenTwo\nenThree\n",
		},
		{
			name: "enum name property in string concatenation",
			input: `
				type TColor = (Red, Green, Blue);
				var c := Red;
				PrintLn('Color: ' + c.Name);
			`,
			expect: "Color: Red\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.input)
			if output != tt.expect {
				t.Errorf("expected %q, got %q", tt.expect, output)
			}
		})
	}
}

// ============================================================================
// Enum .QualifiedName Helper Property Tests
// ============================================================================

func TestEnumQualifiedNameProperty(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name: "enum qualifiedname property basic",
			input: `
				type TColor = (Red, Green, Blue);
				PrintLn(Red.QualifiedName);
				PrintLn(Green.QualifiedName);
				PrintLn(Blue.QualifiedName);
			`,
			expect: "TColor.Red\nTColor.Green\nTColor.Blue\n",
		},
		{
			name: "enum qualifiedname property with explicit ordinals",
			input: `
				type TStatus = (Ok = 100, Error = 200);
				PrintLn(Ok.QualifiedName);
				PrintLn(Error.QualifiedName);
			`,
			expect: "TStatus.Ok\nTStatus.Error\n",
		},
		{
			name: "enum qualifiedname property from variable",
			input: `
				type TColor = (Red, Green, Blue);
				var c := Green;
				PrintLn(c.QualifiedName);
			`,
			expect: "TColor.Green\n",
		},
		{
			name: "enum qualifiedname property in for-in loop",
			input: `
				type TEnum = (enOne, enTwo, enThree);
				for var e in TEnum do
					PrintLn(e.QualifiedName);
			`,
			expect: "TEnum.enOne\nTEnum.enTwo\nTEnum.enThree\n",
		},
		{
			name: "enum qualifiedname vs name difference",
			input: `
				type TPriority = (Low, High);
				var p := Low;
				PrintLn(p.Name);
				PrintLn(p.QualifiedName);
			`,
			expect: "Low\nTPriority.Low\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output := testEvalWithOutput(tt.input)
			if output != tt.expect {
				t.Errorf("expected %q, got %q", tt.expect, output)
			}
		})
	}
}
