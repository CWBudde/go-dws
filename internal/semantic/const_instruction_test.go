package semantic

import (
	"strings"
	"testing"
)

// A statement whose expression DWScript can prove constant computes a value and
// throws it away. The wording, the severity and the anchor at the start of the
// expression are recorded in testdata/fixtures/FailureScripts/ignore_result.txt
// and array_static_methods.txt.
func TestConstantInstructionHint(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "stateless builtin with constant arguments",
			input: "IntToStr(123);",
			want:  "Hint: Constant Instruction - has no effect [line: 1, column: 1]",
		},
		{
			// Upstream decides constness without evaluating, so a call that
			// would raise is still reported as a constant instruction.
			name:  "stateless builtin that would fail to evaluate",
			input: "StrToInt('A');",
			want:  "Hint: Constant Instruction - has no effect [line: 1, column: 1]",
		},
		{
			name:  "constant identifier",
			input: "const C = 1;\nC;",
			want:  "Hint: Constant Instruction - has no effect [line: 2, column: 1]",
		},
		{
			name:  "operator over constants",
			input: "1 + 2;",
			want:  "Hint: Constant Instruction - has no effect [line: 1, column: 1]",
		},
		{
			name:  "static array bound",
			input: "var a : array [1..10] of Integer;\na.Low;",
			want:  "Hint: Constant Instruction - has no effect [line: 2, column: 1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzeWithHints(t, tt.input, HintsLevelPedantic)
			if !containsDiagnostic(got, tt.want) {
				t.Fatalf("expected %q among %v", tt.want, got)
			}
		})
	}
}

// Statements that do something, or whose value cannot be known at compile time,
// must stay silent — a spurious hint would break every passing fixture that
// discards a call's result.
func TestConstantInstructionHintNotEmitted(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "call with a non-constant argument",
			input: "var i := 1;\nIntToStr(i);",
		},
		{
			name:  "call to a user routine",
			input: "function F : Integer;\nbegin\n   Result := 1;\nend;\nF;",
		},
		{
			name:  "a stateful builtin",
			input: "Random;",
		},
		{
			name:  "dynamic array length depends on the instance",
			input: "var a : array of Integer;\na.Length;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzeWithHints(t, tt.input, HintsLevelPedantic)
			if hasDiagnosticContaining(got, "Constant Instruction") {
				t.Fatalf("unexpected constant-instruction hint in %v", got)
			}
		})
	}
}

// The intrinsic array helpers that grow, shrink or reorder the storage exist
// only on dynamic arrays, and the ones that need the actual storage cannot be
// reached through the type name. Both wordings and both anchors come from
// testdata/fixtures/FailureScripts/array_static_methods.txt and dyn_array4.txt.
func TestArrayHelperReceiverRestrictions(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "mutating helper on a static array",
			input: "var a : array [1..10] of Integer;\na.Swap(1, 2);",
			want:  `Array method "Swap" is restricted to dynamic arrays at 2:3`,
		},
		{
			name:  "Copy on a static array",
			input: "var a : array [1..10] of Integer;\na.Copy();",
			want:  `Array method "Copy" is restricted to dynamic arrays at 2:3`,
		},
		{
			name:  "instance helper reached through a dynamic array type",
			input: "type TStrings = array of String;\nPrintLn(TStrings.High);",
			want:  "Array instance expected at 2:18",
		},
		{
			// The restriction outranks the receiver check: upstream reports the
			// static array first and says nothing about the missing instance.
			name:  "static array type takes the restriction, not the receiver error",
			input: "type TStrings2 = array [1..2] of String;\nTStrings2.Add('1');",
			want:  `Array method "Add" is restricted to dynamic arrays at 2:11`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzeWithHints(t, tt.input, HintsLevelPedantic)
			if !containsDiagnostic(got, tt.want) {
				t.Fatalf("expected %q among %v", tt.want, got)
			}
		})
	}
}

// Low is answerable from the type alone — it is 0 for every dynamic array — and
// a static array's bounds are fixed, so neither needs an instance.
func TestArrayHelperTypeReceiverAllowed(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Low of a dynamic array type",
			input: "type TStrings = array of String;\nPrintLn(TStrings.Low);",
		},
		{
			name:  "bounds of a static array type",
			input: "type TStrings2 = array [1..2] of String;\nPrintLn(TStrings2.Low);\nPrintLn(TStrings2.High);",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzeWithHints(t, tt.input, HintsLevelPedantic)
			if hasDiagnosticContaining(got, "Array instance expected") {
				t.Fatalf("unexpected receiver error in %v", got)
			}
		})
	}
}

// hasDiagnosticContaining is the substring counterpart of containsDiagnostic,
// for the negative cases, which assert that a wording never appears at all.
func hasDiagnosticContaining(diagnostics []string, want string) bool {
	for _, d := range diagnostics {
		if strings.Contains(d, want) {
			return true
		}
	}
	return false
}
