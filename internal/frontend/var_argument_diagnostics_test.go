package frontend

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_VarArgumentFixtures(t *testing.T) {
	for _, fixture := range []string{"passing_const_var", "passing_const_var2", "const_param2", "self_not_writable"} {
		t.Run(fixture, func(t *testing.T) {
			base := filepath.Join("../../testdata/fixtures/FailureScripts", fixture)
			source, err := os.ReadFile(base + ".pas")
			if err != nil {
				t.Fatal(err)
			}
			expected, err := os.ReadFile(base + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			assertDiagnostics(t, string(source), base+".pas", strings.Split(strings.TrimSpace(string(expected)), "\n"))
		})
	}
}

func TestCompile_VarArgumentRecovery(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		want         []string
	}{
		{"record constant reader", "type R = record const c = 1; property p: Integer read c; end;\nprocedure Take(var v: Integer); begin end; var rec: R;\nTake(rec.p);", []string{
			"Syntax Error: Argument 0 (v) cannot be passed as Var-parameter [line: 3, column: 6]",
		}},
		{"class constant reader", "type C = class const c = 1; property p: Integer read c; end;\nprocedure Take(var v: Integer); begin end; var obj: C;\nTake(obj.p);", []string{
			"Syntax Error: Argument 0 (v) cannot be passed as Var-parameter [line: 3, column: 6]",
		}},
		{"explicit routine reference", "type F = function: Integer;\nprocedure Take(var v: F); begin end;\nfunction Get: Integer; begin Result := 1; end;\nTake(@Get);", []string{
			"Syntax Error: Argument 0 (v) cannot be passed as Var-parameter [line: 4, column: 6]",
		}},
		{"unqualified record constant", "procedure Take(var v: Integer); begin end;\ntype R = record const c = 1; procedure Test; begin Take(c); end; end;", []string{
			"Syntax Error: Argument 0 (v) cannot be passed as Var-parameter [line: 2, column: 57]",
		}},
		{"array literal recovery", "type B = array [0..2] of Integer;\nprocedure Take(var v: B); begin end;\nTake([1,2]);", []string{
			"Syntax Error: Argument 0 (v) cannot be passed as Var-parameter [line: 3, column: 6]",
		}},
		{"const array recovery", "type A = array [0..1] of Integer; type B = array [0..2] of Integer;\nprocedure Take(var v: B); begin end;\nconst c: A = [1,2];\nTake(c);", []string{
			"Syntax Error: Argument 0 (v) cannot be passed as Var-parameter [line: 4, column: 6]",
		}},
		{"declared parameter casing", "procedure Take(var MiXeD: Integer); begin end;\nconst c = 1;\nTake(c);", []string{
			"Syntax Error: Argument 0 (MiXeD) cannot be passed as Var-parameter [line: 3, column: 6]",
		}},
		{"type error precedes writability", "procedure Take(var v: Integer); begin end;\nTake('bad');", []string{
			"Syntax Error: Argument 0 expects type \"Integer\" instead of \"String\" [line: 2, column: 6]",
		}},
		{"unknown argument", "procedure Take(var v: Integer); begin end;\nTake(Missing);", []string{
			"Syntax Error: Unknown name \"Missing\" [line: 2, column: 6]",
		}},
		{"nested invalid argument", "function Take(var v: Integer): Integer; begin Result := v; end;\nTake(Take(1));", []string{
			"Syntax Error: Argument 0 (v) cannot be passed as Var-parameter [line: 2, column: 11]",
		}},
		{"decrement constant", "const c = 1;\nDec(c);", []string{
			"Syntax Error: Argument 0 (a) cannot be passed as Var-parameter [line: 2, column: 5]",
		}},
		{"decrement literal", "Dec(1);", []string{
			"Syntax Error: Argument 0 (a) cannot be passed as Var-parameter [line: 1, column: 5]",
		}},
		{"decrement unknown", "Dec(Missing);", []string{
			"Syntax Error: Unknown name \"Missing\" [line: 1, column: 5]",
		}},
		{"static array const parameter", "procedure Take(var v: Integer); begin end;\nprocedure Test(const a: array [0..1] of Integer);\nbegin Take(a[0]); end;", []string{
			"Syntax Error: Argument 0 (v) cannot be passed as Var-parameter [line: 3, column: 12]",
		}},
		{"const record field", "type R = record f: Integer; end;\nprocedure Take(var v: Integer); begin end;\nprocedure Test(const r: R);\nbegin Take(r.f); end;", []string{
			"Syntax Error: Argument 0 (v) cannot be passed as Var-parameter [line: 4, column: 12]",
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := Compile(tt.source, "<test>", semantic.HintsLevelDisabled).DiagnosticStrings()
			if strings.Join(got, "\n") != strings.Join(tt.want, "\n") {
				t.Fatalf("got %v; want %v", got, tt.want)
			}
		})
	}
}

func TestCompile_WritableVarArguments(t *testing.T) {
	for _, source := range []string{
		"procedure Take(var v: JSONVariant); begin v := 3; end; var j: JSONVariant := ParseJSON('{\"a\":1}'); Take(j.a);",
		"type H = helper for Integer class var f: Integer; property p: Integer read f; end; procedure Take(var v: Integer); begin end; Take(Integer.f); Inc(Integer.f); var n := 0; Take(n.p); Inc(n.p);",
		"type I = Integer; type H = helper for I class var f: Integer; end; procedure Take(var v: Integer); begin end; Take(I.f); Inc(I.f);",
		"type R = record class var f: Integer; property p: Integer read f; end; procedure Take(var v: Integer); begin end; Take(R.p); Inc(R.p);",

		"procedure Take(var v: Integer); begin end; type R = record f: Integer; procedure Test; begin Take(Self.f); Inc(Self.f); end; end;",
		"procedure Take(var v: Integer); begin end; type R = record a: array [0..1] of Integer; procedure Test; begin Take(Self.a[0]); Inc(Self.a[0]); end; end;",

		"procedure Take(var v: Integer); begin end; var n := 1; Take(n); Inc(n); Dec(n);",
		"procedure Take(var v: Integer); begin end; var a: array [0..1] of Integer; Take(a[0]);",
		"procedure Take(var v: Integer); begin end; var a: array of Integer; Take(a[0]);",
		"procedure Take(var v: Integer); begin end; procedure Test(const a: array of Integer); begin Take(a[0]); end;",
		"type R = record f: Integer; end; procedure Take(var v: Integer); begin end; var rec: R; Take(rec.f);",
		"type C = class f: Integer; end; procedure Take(var v: Integer); begin end; procedure Test(const c: C); begin Take(c.f); end;",
	} {
		t.Run(source, func(t *testing.T) {
			if got := Compile(source, "<test>", semantic.HintsLevelDisabled).DiagnosticStrings(); len(got) != 0 {
				t.Fatalf("unexpected diagnostics: %v", got)
			}
		})
	}
}
