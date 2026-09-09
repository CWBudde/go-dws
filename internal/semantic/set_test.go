package semantic

import (
	"testing"
)

// ============================================================================
// Set Type Registration Tests
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
			expectedError: "Name \"TDays\" already exists",
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
// Set Literal Tests
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
// Set Operations Tests
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
// Set Membership Tests
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
		{
			name: "in operator with array membership",
			input: `
				var ints: array of Integer;
				var ok: Boolean := 1 in ints;
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
			expectedError: "Incompatible operands",
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
			expectedError: "type mismatch in 'in' operator",
		},
		{
			name: "in operator with wrong array element type",
			input: `
				var ints: array of Integer;
				var bad: Boolean := 'oops' in ints;
			`,
			expectedError: `Incompatible types: "String" and "Integer"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectError(t, tt.input, tt.expectedError)
		})
	}
}

// ============================================================================
// Array <-> Set Conversion Tests (L-S6a)
// ============================================================================

// TestBracketLiteralConversions covers both directions of the bracket-literal
// disambiguation. DWScript writes array and set constructors with the same `[]`
// syntax, so the expected type — not the parser's syntactic heuristic — decides
// which one a literal becomes.
func TestBracketLiteralConversions(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "set literal shape flows into an array-typed target",
			input: `
				var a: array of Integer := [1, 2];
				var b: array of String := ['x', 'y'];
			`,
		},
		{
			name: "identifier elements flow into an array-typed target",
			input: `
				const lo = 1;
				const hi = 2;
				var a: array of Integer := [lo, hi];
			`,
		},
		{
			name: "empty literal into a set-typed target",
			input: `
				type TColor = (Red, Green, Blue);
				type TColors = set of TColor;
				var c: TColors := [];
			`,
		},
		{
			name: "single element into a set-typed target",
			input: `
				type TColor = (Red, Green, Blue);
				type TColors = set of TColor;
				var c: TColors := [Green];
			`,
		},
		{
			name: "range into a set-typed target",
			input: `
				type TColor = (Red, Green, Blue);
				type TColors = set of TColor;
				var c: TColors := [Red..Blue];
			`,
		},
		{
			// A typecast element defeats the parser's heuristic, so this literal
			// arrives as an array literal and must still convert.
			name: "typecast element into a set-typed target",
			input: `
				type TColor = (Red, Green, Blue);
				type TColors = set of TColor;
				var c: TColors;
				c := c + [TColor(1)];
			`,
		},
		{
			name: "set constant declaration",
			input: `
				type TElem = (A = 1, B = 64, C = 128);
				type TMy = set of TElem;
				const v : TMy = [A, C];
			`,
		},
		{
			name: "inline anonymous enum in a variable's set type",
			input: `
				var e : set of (et1, et2) = [];
				var f : Boolean := et1 in e;
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}

// TestSetElementBounds covers the compile-time bounds check on set elements
// (L-S6c). Only compile-time-constant ordinals are checked; anything else is
// left to run time, where an out-of-range ordinal is simply never a member.
func TestSetElementBounds(t *testing.T) {
	const setup = `
		type TColor = (Red, Green, Blue);
		type TColors = set of TColor;
		var c: TColors;
	`

	errorCases := []struct {
		name  string
		input string
	}{
		{name: "above the high bound", input: setup + "c := c + [TColor(3)];"},
		{name: "below the low bound", input: setup + "c := c + [TColor(-1)];"},
		{name: "range bound out of set", input: setup + "c := c + [TColor(0)..TColor(7)];"},
	}
	for _, tt := range errorCases {
		t.Run(tt.name, func(t *testing.T) {
			expectError(t, tt.input, "Element is out of set bounds")
		})
	}

	okCases := []struct {
		name  string
		input string
	}{
		{name: "at the low bound", input: setup + "c := c + [TColor(0)];"},
		{name: "at the high bound", input: setup + "c := c + [TColor(2)];"},
	}
	for _, tt := range okCases {
		t.Run(tt.name, func(t *testing.T) {
			expectNoErrors(t, tt.input)
		})
	}
}

// TestSetIntegerCastWidth covers the set <-> Integer cast rule (L-S6c). The
// integer form is a bitmask indexed by the element's *absolute* ordinal, so the
// rule is about the highest reachable ordinal, not the base type's span.
func TestSetIntegerCastWidth(t *testing.T) {
	t.Run("narrow set casts both ways", func(t *testing.T) {
		expectNoErrors(t, `
			type TEnum = (one, two);
			type TSet = set of TEnum;
			var s : TSet;
			var i := Integer(s);
			s := TSet(i);
		`)
	})

	t.Run("highest representable ordinal is accepted", func(t *testing.T) {
		expectNoErrors(t, `
			type TEnum = (lo = 0, hi = 31);
			type TSet = set of TEnum;
			var s : TSet;
			var i := Integer(s);
		`)
	})

	t.Run("wide set rejects the cast to integer", func(t *testing.T) {
		expectError(t, `
			type TEnum = (one = 1, fifty = 50);
			type TSet = set of TEnum;
			var s : TSet;
			var i := Integer(s);
		`, "Set has too many elements for cast to integer")
	})

	t.Run("wide set rejects the cast from integer", func(t *testing.T) {
		expectError(t, `
			type TEnum = (one = 1, fifty = 50);
			type TSet = set of TEnum;
			var i : Integer;
			var s := TSet(i);
		`, "Set has too many elements for cast to integer")
	})

	t.Run("narrow span at high ordinals is rejected", func(t *testing.T) {
		// Span 32, but ordinal 64 needs bit 64 and would be dropped silently.
		expectError(t, `
			type TEnum = (h33 = 33, h64 = 64);
			type TSet = set of TEnum;
			var s : TSet;
			var i := Integer(s);
		`, "Set has too many elements for cast to integer")
	})

	t.Run("unbounded ordinal base is rejected", func(t *testing.T) {
		// `set of Integer` works (map storage) but has no bitmask form.
		expectError(t, `
			var s : set of Integer := [65];
			var i := Integer(s);
		`, "Set has too many elements for cast to integer")
	})
}

// TestSetMutationRequiresVariable covers the receiver check on Include/Exclude.
// Both mutate in place, so a constant receiver would otherwise rewrite the
// constant at run time — `const` sets became declarable with L-S6a.
func TestSetMutationRequiresVariable(t *testing.T) {
	const setup = `
		type TElem = (A, B, C);
		type TSet = set of TElem;
	`

	t.Run("variable receiver is accepted", func(t *testing.T) {
		expectNoErrors(t, setup+`
			var v : TSet := [A];
			Include(v, B);
			v.Exclude(A);
		`)
	})

	t.Run("constant receiver is rejected in procedure form", func(t *testing.T) {
		expectError(t, setup+`
			const v : TSet = [A];
			Include(v, B);
		`, "Variable expected")
	})

	t.Run("constant receiver is rejected in method form", func(t *testing.T) {
		expectError(t, setup+`
			const v : TSet = [A];
			v.Include(B);
		`, "Variable expected")
	})

	t.Run("function result is rejected", func(t *testing.T) {
		expectError(t, setup+`
			function Make : TSet; begin end;
			Include(Make, A);
		`, "Variable expected")
	})
}

// TestInlineSetEnumScope covers where an inline `set of (a, b)` may appear.
// A parameter list has no statement to hoist the implicit enum in front of, so
// it is rejected rather than published into the scope around the routine.
func TestInlineSetEnumScope(t *testing.T) {
	t.Run("statement-level declaration is accepted", func(t *testing.T) {
		expectNoErrors(t, `
			var s : set of (et1, et2) = [];
			var b : Boolean := et1 in s;
		`)
	})

	t.Run("parameter position is rejected", func(t *testing.T) {
		expectError(t, `
			procedure Test(s : set of (pt1, pt2));
			begin
			end;
		`, "anonymous enumeration is not allowed in a parameter's set type")
	})
}
