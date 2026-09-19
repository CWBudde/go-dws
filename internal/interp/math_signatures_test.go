package interp

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/frontend"
	"github.com/cwbudde/go-dws/internal/semantic"
)

// TestMathSignaturesAndVariantArguments covers the math builtins' upstream
// signatures (E11) through the real compile/run path: Variant operands for Abs,
// Variant deltas for Inc/Dec/Succ/Pred, the Haversine radius and RandG's
// mean/deviation parameters.
func TestMathSignaturesAndVariantArguments(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"Abs Variant Integer", `var v : Variant := -3; PrintLn(Abs(v));`, "3\n"},
		{"Abs Variant Float", `var v : Variant := -1.5; PrintLn(Abs(v));`, "1.5\n"},
		{"Abs Variant result assignable", `var v : Variant := -4; var i : Integer := Abs(v); PrintLn(i + 1);`, "5\n"},
		{"Inc Dec Variant delta", `var i := 1; var v : Variant := 2; Inc(i, v); PrintLn(i); Dec(i, v); PrintLn(i);`, "3\n1\n"},
		{"Succ Pred Integer delta", `PrintLn(Succ(10, 3)); PrintLn(Pred(10, 3));`, "13\n7\n"},
		{"Succ Pred Variant delta", `var i := 1; var v : Variant := 2; PrintLn(Succ(i, v)); PrintLn(Pred(i, v)); PrintLn(i);`, "3\n-1\n1\n"},
		{"Succ Pred single argument", `PrintLn(Succ(1)); PrintLn(Pred(1));`, "2\n0\n"},
		{"Succ Pred enum delta", `type TE = (a, b, c, d); PrintLn(Ord(Succ(a, 2))); PrintLn(Ord(Pred(d, 3)));`, "2\n0\n"},
		{"Haversine default radius", `PrintLn(Haversine(36.12, -86.67, 33.94, -118.40).ToString(3));`, "2886.444\n"},
		{"Haversine explicit radius", `PrintLn(Haversine(36.12, -86.67, 33.94, -118.40, 6371).ToString(3));`, "2886.444\n"},
		{"Haversine scales with radius", `PrintLn(Haversine(0, 0, 0, 90, 1000).ToString(3));`, "1570.796\n"},
		{"RandG zero deviation", `PrintLn(RandG(100, 0));`, "100\n"},
		{"RandG seeded mean", `
SetRandSeed(12);
var s : Float;
var i : Integer;
for i := 1 to 100 do
   s := s + RandG(100, 5);
s := s / 100;
if (s < 95) or (s > 105) then PrintLn('RandG oddity') else PrintLn('ok');`, "ok\n"},
		{"RandG no arguments", `var g := RandG; if (g > -100) and (g < 100) then PrintLn('ok');`, "ok\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := runHelperScript(t, tt.source); got != tt.want {
				t.Fatalf("output mismatch: want %q, got %q", tt.want, got)
			}
		})
	}
}

// TestMathSignaturesRejectInvalidCalls keeps the widened signatures strict.
func TestMathSignaturesRejectInvalidCalls(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"Succ three arguments", `PrintLn(Succ(1, 2, 3));`},
		{"Pred string delta", `PrintLn(Pred(1, 'x'));`},
		{"Inc string delta", `var i := 1; Inc(i, 'x');`},
		{"Abs String", `PrintLn(Abs('x'));`},
		{"Haversine six arguments", `PrintLn(Haversine(1, 2, 3, 4, 5, 6));`},
		{"Haversine three arguments", `PrintLn(Haversine(1, 2, 3));`},
		{"RandG one argument", `PrintLn(RandG(1));`},
		{"RandG string argument", `PrintLn(RandG('a', 1));`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiled := frontend.Compile(tt.source, "math_signatures.pas", semantic.HintsLevelNormal)
			if !compiled.HasFatalDiagnostics() && compiled.SemanticSuccessful {
				t.Fatalf("expected a compile error, got none")
			}
			if len(compiled.DiagnosticStrings()) == 0 {
				t.Fatalf("expected diagnostics")
			}
			for _, d := range compiled.DiagnosticStrings() {
				if strings.Contains(d, "internal error") {
					t.Fatalf("unexpected internal error: %s", d)
				}
			}
		})
	}
}
