package semantic

import (
	"testing"
)

// ============================================================================
// Set Type Registration Tests (Task 8.99)
// ============================================================================

func TestSetTypeRegistration(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "simple set type",
			input: `
				type TWeekday = (Mon, Tue, Wed);
				type TDays = set of TWeekday;
			`,
		},
		{
			name: "set type with variable",
			input: `
				type TColor = (Red, Green, Blue);
				type TColors = set of TColor;
				var colors: TColors;
			`,
		},
		{
			name: "multiple set types",
			input: `
				type TWeekday = (Mon, Tue, Wed);
				type TColor = (Red, Green, Blue);
				type TDays = set of TWeekday;
				type TColors = set of TColor;
			`,
		},
		// Note: Set literal type checking is Task 8.101, not Task 8.99
		// Removed set literal test for now
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}

func TestSetTypeErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "duplicate set type declaration",
			input: `
				type TWeekday = (Mon, Tue, Wed);
				type TDays = set of TWeekday;
				type TDays = set of TWeekday;
			`,
			expectedError: "set type 'TDays' already declared",
		},
		{
			name: "undefined element type",
			input: `
				type TDays = set of TWeekday;
			`,
			expectedError: "unknown type 'TWeekday'",
		},
		{
			name: "undefined set type in variable",
			input: `
				var days: TDays;
			`,
			expectedError: "unknown type 'TDays'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectError(t, tt.input, tt.expectedError)
		})
	}
}

// ============================================================================
// Set Literal Tests (Task 8.101)
// ============================================================================

func TestSetLiterals(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "set literal with valid enum elements",
			input: `
				type TColor = (Red, Green, Blue);
				type TColors = set of TColor;
				var colors: TColors := [Red, Green];
			`,
		},
		{
			name: "empty set literal",
			input: `
				type TColor = (Red, Green, Blue);
				type TColors = set of TColor;
				var colors: TColors := [];
			`,
		},
		{
			name: "set literal with all enum values",
			input: `
				type TWeekday = (Mon, Tue, Wed, Thu, Fri);
				type TDays = set of TWeekday;
				var workDays: TDays := [Mon, Tue, Wed, Thu, Fri];
			`,
		},
		{
			name: "set literal with single element",
			input: `
				type TColor = (Red, Green, Blue);
				type TColors = set of TColor;
				var favorite: TColors := [Blue];
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}

func TestSetLiteralErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "set literal with wrong enum type",
			input: `
				type TColor = (Red, Green, Blue);
				type TSize = (Small, Medium, Large);
				type TColors = set of TColor;
				var colors: TColors := [Small];
			`,
			expectedError: "type mismatch",
		},
		{
			name: "set literal with mixed types",
			input: `
				type TColor = (Red, Green, Blue);
				type TSize = (Small, Medium, Large);
				type TColors = set of TColor;
				var colors: TColors := [Red, Small];
			`,
			expectedError: "type mismatch",
		},
		{
			name: "set literal with undefined value",
			input: `
				type TColor = (Red, Green, Blue);
				type TColors = set of TColor;
				var colors: TColors := [Red, Yellow];
			`,
			expectedError: "undefined variable 'Yellow'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectError(t, tt.input, tt.expectedError)
		})
	}
}

// ============================================================================
// Set Operations Tests (Task 8.102)
// ============================================================================

func TestSetOperations(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "set union",
			input: `
				type TColor = (Red, Green, Blue);
				type TColors = set of TColor;
				var s1: TColors := [Red, Green];
				var s2: TColors := [Blue];
				var s3: TColors := s1 + s2;
			`,
		},
		{
			name: "set difference",
			input: `
				type TColor = (Red, Green, Blue);
				type TColors = set of TColor;
				var s1: TColors := [Red, Green, Blue];
				var s2: TColors := [Green];
				var s3: TColors := s1 - s2;
			`,
		},
		{
			name: "set intersection",
			input: `
				type TColor = (Red, Green, Blue);
				type TColors = set of TColor;
				var s1: TColors := [Red, Green];
				var s2: TColors := [Green, Blue];
				var s3: TColors := s1 * s2;
			`,
		},
		{
			name: "chained set operations",
			input: `
				type TWeekday = (Mon, Tue, Wed, Thu, Fri);
				type TDays = set of TWeekday;
				var workDays: TDays := [Mon, Tue, Wed, Thu, Fri];
				var weekend: TDays := [];
				var allDays: TDays := workDays + weekend;
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}

func TestSetOperationErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "set union with incompatible types",
			input: `
				type TColor = (Red, Green, Blue);
				type TSize = (Small, Medium, Large);
				type TColors = set of TColor;
				type TSizes = set of TSize;
				var s1: TColors := [Red];
				var s2: TSizes := [Small];
				var s3 := s1 + s2;
			`,
			expectedError: "incompatible types",
		},
		{
			name: "set union with non-set operand",
			input: `
				type TColor = (Red, Green, Blue);
				type TColors = set of TColor;
				var s1: TColors := [Red];
				var s2: Integer := 42;
				var s3 := s1 + s2;
			`,
			expectedError: "requires set operands",
		},
		{
			name: "set difference with incompatible types",
			input: `
				type TColor = (Red, Green, Blue);
				type TSize = (Small, Medium, Large);
				type TColors = set of TColor;
				type TSizes = set of TSize;
				var s1: TColors := [Red];
				var s2: TSizes := [Small];
				var s3 := s1 - s2;
			`,
			expectedError: "incompatible types",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectError(t, tt.input, tt.expectedError)
		})
	}
}

// ============================================================================
// Set Membership Tests (Task 8.103)
// ============================================================================

func TestSetMembership(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "in operator with enum value",
			input: `
				type TColor = (Red, Green, Blue);
				type TColors = set of TColor;
				var colors: TColors := [Red, Green];
				var isRed: Boolean := Red in colors;
			`,
		},
		{
			name: "in operator with variable",
			input: `
				type TWeekday = (Mon, Tue, Wed, Thu, Fri);
				type TDays = set of TWeekday;
				var workDays: TDays := [Mon, Tue, Wed, Thu, Fri];
				var day: TWeekday := Mon;
				var isWorkDay: Boolean := day in workDays;
			`,
		},
		{
			name: "in operator in conditional",
			input: `
				type TColor = (Red, Green, Blue);
				type TColors = set of TColor;
				var colors: TColors := [Red, Green];
				if Red in colors then
					var x: Integer := 1;
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}

func TestSetMembershipErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "in operator with wrong enum type",
			input: `
				type TColor = (Red, Green, Blue);
				type TSize = (Small, Medium, Large);
				type TColors = set of TColor;
				var colors: TColors := [Red];
				var result: Boolean := Small in colors;
			`,
			expectedError: "type mismatch",
		},
		{
			name: "in operator with non-set right operand",
			input: `
				type TColor = (Red, Green, Blue);
				var x: Integer := 42;
				var result: Boolean := Red in x;
			`,
			expectedError: "requires set",
		},
		{
			name: "in operator with non-enum left operand",
			input: `
				type TColor = (Red, Green, Blue);
				type TColors = set of TColor;
				var colors: TColors := [Red];
				var x: Integer := 42;
				var result: Boolean := x in colors;
			`,
			expectedError: "requires enum value as left operand",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectError(t, tt.input, tt.expectedError)
		})
	}
}
