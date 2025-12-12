package interp

import (
	"testing"
)

// ============================================================================
// Enum Constant Expression Evaluation Tests
// ============================================================================

// TestEnumConstantExpressionEvaluation tests that enum values with
// constant expressions are correctly evaluated and usable at runtime.
func TestEnumConstantExpressionEvaluation(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name: "Ord() function in enum value",
			input: `
				type TEnumAlpha = (eAlpha = Ord('A'), eBeta, eGamma);

				PrintLn(Ord(eAlpha));  // Should print 65
				PrintLn(Ord(eBeta));   // Should print 66
				PrintLn(Ord(eGamma));  // Should print 67
			`,
			expect: "65\n66\n67\n",
		},
		{
			name: "Arithmetic expressions in enum values",
			input: `
				type TEnum = (a = 1+2, b = 5*3, c = 10-2);

				PrintLn(Ord(a));  // Should print 3
				PrintLn(Ord(b));  // Should print 15
				PrintLn(Ord(c));  // Should print 8
			`,
			expect: "3\n15\n8\n",
		},
		{
			name: "Chr() function in enum value",
			input: `
				type TEnum = (a = Chr(65), b, c);

				PrintLn(Ord(a));  // Should print 65
				PrintLn(Ord(b));  // Should print 66
			`,
			expect: "65\n66\n",
		},
		{
			name: "Negative value expression",
			input: `
				type TEnum = (a = -5, b, c);

				PrintLn(Ord(a));  // Should print -5
				PrintLn(Ord(b));  // Should print -4
			`,
			expect: "-5\n-4\n",
		},
		{
			name: "Enum value used as ordinal source",
			input: `
				type TBase = (BaseA = 10, BaseB, BaseC);
				type TDerived = (First = BaseB, Second);

				PrintLn(Ord(First));   // Should print 11
				PrintLn(Ord(Second));  // Should print 12
			`,
			expect: "11\n12\n",
		},
		{
			name: "Boolean constant expression",
			input: `
				type TBoolEnum = (FalseVal = False, TrueVal = True, Next);

				PrintLn(Ord(FalseVal));    // Should print 0
				PrintLn(Ord(TrueVal));     // Should print 1
				PrintLn(Ord(Next));   // Should print 2
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
