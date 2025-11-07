package semantic

import (
	"testing"
)

// ============================================================================
// Enum Declaration Tests
// ============================================================================

func TestEnumDeclaration(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "simple enum",
			input: `
				type TColor = (Red, Green, Blue);
			`,
		},
		{
			name: "enum with explicit values",
			input: `
				type TStatus = (Ok = 0, Warning = 1, Error = 2);
			`,
		},
		{
			name: "enum with mixed values",
			input: `
				type TPriority = (Low, Medium = 5, High);
			`,
		},
		{
			name: "scoped enum",
			input: `
				type TEnum = enum (One, Two, Three);
			`,
		},
		{
			name: "enum type in variable declaration",
			input: `
				type TColor = (Red, Green, Blue);
				var color: TColor;
			`,
		},
		{
			name: "enum value as constant",
			input: `
				type TColor = (Red, Green, Blue);
				var color: TColor := Red;
			`,
		},
		{
			name: "multiple enums",
			input: `
				type TColor = (Red, Green, Blue);
				type TSize = (Small, Medium, Large);
				var c: TColor := Red;
				var s: TSize := Small;
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}

func TestEnumErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "duplicate enum value names",
			input: `
				type TColor = (Red, Green, Red);
			`,
			expectedError: "duplicate enum value 'Red'",
		},
		{
			name: "duplicate explicit values",
			input: `
				type TStatus = (Ok = 0, Warning = 0);
			`,
			expectedError: "duplicate enum ordinal value 0",
		},
		{
			name: "undefined enum type",
			input: `
				var color: TColor;
			`,
			expectedError: "unknown type 'TColor'",
		},
		{
			name: "undefined enum value",
			input: `
				type TColor = (Red, Green, Blue);
				var color: TColor := Yellow;
			`,
			expectedError: "undefined variable 'Yellow'",
		},
		{
			name: "duplicate enum type declaration",
			input: `
				type TColor = (Red, Green, Blue);
				type TColor = (Cyan, Magenta, Yellow);
			`,
			expectedError: "already declared",
		},
		{
			name: "type mismatch with enum",
			input: `
				type TColor = (Red, Green, Blue);
				var color: TColor := 42;
			`,
			expectedError: "cannot assign Integer to TColor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectError(t, tt.input, tt.expectedError)
		})
	}
}

func TestEnumValueResolution(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "enum value in assignment",
			input: `
				type TColor = (Red, Green, Blue);
				var color: TColor;
				color := Red;
			`,
		},
		{
			name: "enum value in comparison",
			input: `
				type TColor = (Red, Green, Blue);
				var color: TColor := Red;
				var same: Boolean := color = Red;
			`,
		},
		{
			name: "enum value in case statement",
			input: `
				type TColor = (Red, Green, Blue);
				var color: TColor := Red;
				case color of
					Red: PrintLn('Red');
					Green: PrintLn('Green');
					Blue: PrintLn('Blue');
				end;
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}

func TestEnumOrdinalValues(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "implicit ordinal values start at 0",
			input: `
				type TColor = (Red, Green, Blue);
			`,
		},
		{
			name: "explicit values override implicit",
			input: `
				type TEnum = (One = 1, Two = 2, Three = 3);
			`,
		},
		{
			name: "mixed explicit and implicit values",
			input: `
				type TEnum = (A, B = 10, C, D = 20, E);
			`,
		},
		{
			name: "negative enum values",
			input: `
				type TEnum = (Negative = -1, Zero = 0, Positive = 1);
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}

// TestEnumTypeMetaValues tests enum type names as runtime values.
func TestEnumTypeMetaValues(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "enum type name as identifier",
			input: `
				type TColor = (Red, Green, Blue);
				var x := TColor;
			`,
		},
		{
			name: "enum type name in High()",
			input: `
				type TColor = (Red, Green, Blue);
				var highest: TColor := High(TColor);
			`,
		},
		{
			name: "enum type name in Low()",
			input: `
				type TColor = (Red, Green, Blue);
				var lowest: TColor := Low(TColor);
			`,
		},
		{
			name: "multiple enum type meta-values",
			input: `
				type TColor = (Red, Green, Blue);
				type TSize = (Small, Medium, Large);
				var c := High(TColor);
				var s := Low(TSize);
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}

// ============================================================================
// Enum .Value Helper Property Tests (Task 9.31)
// ============================================================================

func TestEnumValueHelperProperty(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "enum value property access",
			input: `
				type TColor = (Red, Green, Blue);
				var c := Red;
				var ordinal: Integer := c.Value;
			`,
		},
		{
			name: "scoped enum value property",
			input: `
				type TColor = (Red, Green, Blue);
				var c := Green;
				var ordinal: Integer := c.Value;
			`,
		},
		{
			name: "enum value property in expression",
			input: `
				type TPriority = (Low, Medium = 5, High);
				var p := High;
				var doubled := p.Value * 2;
			`,
		},
		{
			name: "enum value property chaining",
			input: `
				type TEnum = (eOne=1, eTwo, eThree);
				var s: String := eTwo.Value.ToString;
			`,
		},
		{
			name: "enum value property in comparison",
			input: `
				type TPriority = (Low = 1, Medium = 5, High = 10);
				var p1 := Low;
				var p2 := High;
				var result := p1.Value < p2.Value;
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}
