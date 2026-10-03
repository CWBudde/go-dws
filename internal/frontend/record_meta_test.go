package frontend

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_RecordMetaFixture(t *testing.T) {
	source, err := os.ReadFile("../../testdata/fixtures/FailureScripts/record_meta.pas")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("../../testdata/fixtures/FailureScripts/record_meta.txt")
	if err != nil {
		t.Fatal(err)
	}
	assertDiagnostics(t, string(source), "record_meta.pas", strings.Split(strings.TrimSpace(string(want)), "\n"))
}

func TestCompile_RecordMetaQualifiedAliasesAndVisibility(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Records.pas"), []byte(`unit Records; interface type R=record x:Integer; class var Count: Integer; end; type Alias=R; implementation end.`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name, source string
		success      bool
	}{
		{"qualified alias value", `uses Records; var m := Records.Alias; var copy := m; copy.Count := 2;`, true},
		{"qualified alias instance rejection", `uses Records; var m := Records.Alias; var item: Records.Alias; m := item;`, false},
		{"unqualified alias before uses", `var m := Alias; uses Records;`, false},
		{"deferred before uses", `procedure Early; begin var m := Alias; end; uses Records; Early;`, false},
		{"namespace scalar shadow", `uses Records; procedure Test(Records: Integer); begin var m := Records.Alias; end; Test(1);`, false},
		{"qualified value remains instance", `uses Records; type Holder=record Alias: R; end; procedure Test(Records: Holder); begin var m := Records.Alias; var item: R; m:=item; end;`, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, filepath.Join(dir, "Main.pas"), semantic.HintsLevelDisabled)
			if result.SemanticSuccessful != tt.success {
				t.Fatalf("success=%v: %v", result.SemanticSuccessful, result.DiagnosticStrings())
			}
		})
	}
}

func TestCompile_RecordMetaSupplierAnchors(t *testing.T) {
	const declarations = "type R=record x:Integer; end;\ntype A=R;\nvar m:=A;\nvar item:R;\n"
	for _, tt := range []struct{ line, message string }{
		{"m := item;", `Syntax Error: Incompatible types: Cannot assign "R" to "meta of R" [line: 5, column: 6]`},
		{"item := m;", `Syntax Error: Incompatible types: Cannot assign "meta of R" to "R" [line: 5, column: 9]`},
	} {
		assertDiagnostics(t, declarations+tt.line, "suppliers.dws", []string{tt.message})
	}
}

func TestCompile_RecordMetaDeferredSourceVisibility(t *testing.T) {
	for _, source := range []string{
		`procedure Early; begin var m:=R; end; type R=record x:Integer; end; Early;`,
		`type R=record x:Integer; end; procedure Early; begin var m:=A; end; type A=R; Early;`,
	} {
		result := Compile(source, "deferred_record.dws", semantic.HintsLevelDisabled)
		if result.SemanticSuccessful {
			t.Errorf("later type became visible in earlier body: %s", source)
		}
	}
}
