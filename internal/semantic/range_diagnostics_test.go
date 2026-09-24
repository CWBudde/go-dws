package semantic

import (
	"os"
	"strings"
	"testing"
)

func TestRangeDiagnostics(t *testing.T) {
	for _, tt := range []struct {
		name, source, want string
		count              int
	}{
		{"array enum and integer", `type TAlpha = (alpha, beta); var a := [alpha..2];`, `Range start and range stop are of incompatible types: "TAlpha" and "Integer"`, 1},
		{"array different enums", `type TAlpha = (alpha, beta); type TNum = (one, two); var a := [alpha..two];`, `Range start and range stop are of incompatible types: "TAlpha" and "TNum"`, 1},
		{"set void endpoint", `type TItem = (first, last); procedure Test; begin end; var s: set of TItem := [first..Test()];`, `Range start and range stop are of incompatible types: "TItem" and "void"`, 1},
		{"membership void endpoint", `type TItem = (first, last); procedure Test; begin end; var item := first; if item in [first..Test()] then Test;`, `Range start and range stop are of incompatible types: "TItem" and "void"`, 1},
		{"case mismatch once", `var i: Integer; case i of 1..'z': ; end;`, `Range start and range stop are of incompatible types: "Integer" and "String"`, 1},
		{"case void endpoints", `var i: Integer; case i of TObject(nil).Free..DivMod: ; end;`, `Range start and range stop are of incompatible types: "void" and "void"`, 2},
		{"recover past invalid set element", `type TItem = (first, last); procedure Test; begin end; var s: set of TItem := [first..Test(), last];`, `Range start and range stop are of incompatible types: "TItem" and "void"`, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			a, _ := analyzeSource(t, tt.source)
			if got := a.Errors(); len(got) != tt.count || !strings.Contains(strings.Join(got, "\n"), tt.want) {
				t.Fatalf("want %d diagnostics containing %q, got %v", tt.count, tt.want, got)
			}
		})
	}
}

func TestRangeDiagnostics_ValidNeighbors(t *testing.T) {
	for _, source := range []string{
		`var v: Variant; case v of 1..3.5: ; 2.7..3: ; end;`,
		`type TItem = (first, last); var a := [first..last];`,
		`type TItem = (first, last); var s: set of TItem := [first..last];`,
		`type TItem = (first, last); var item := first; if item in [first..last] then PrintLn(item);`,
	} {
		expectNoErrors(t, source)
	}
}

func TestRangeDiagnostics_ReversedCaseHint(t *testing.T) {
	a, _ := analyzeSource(t, `var i: Integer; case i of 5..1: ; end;`)
	if got := strings.Join(a.Errors(), "\n"); !strings.Contains(got, "Case range condition lower bound is greater than higher bound") {
		t.Fatalf("missing reversed range hint: %s", got)
	}
}

func TestRangeDiagnostics_ValidCaseFixtures(t *testing.T) {
	for _, name := range []string{"case_range_enum", "case_variant_condition"} {
		t.Run(name, func(t *testing.T) {
			source, err := os.ReadFile("../../testdata/fixtures/SimpleScripts/" + name + ".pas")
			if err != nil {
				t.Fatal(err)
			}
			a, _ := analyzeSource(t, string(source))
			if got := a.Errors(); len(got) != 0 {
				t.Fatalf("unexpected diagnostics: %v", got)
			}
		})
	}
}
