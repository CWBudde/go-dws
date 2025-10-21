package semantic

import (
	"testing"
)

// ============================================================================
// Enum Declaration Tests (Task 8.43-8.46)
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
