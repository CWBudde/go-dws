package interp

import (
	"testing"
)

// ============================================================================
// Enum Declaration Tests (Task 8.48 - Interpreter Support)
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
// Enum Value Tests (Task 8.49 - Store enum values)
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
// Enum Literal Tests (Task 8.50 - Evaluate enum literals)
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
