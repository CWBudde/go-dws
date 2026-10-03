package frontend

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestCompile_EnumBindingDiagnostics(t *testing.T) {
	for _, name := range []string{"enums9", "enums10"} {
		t.Run(name, func(t *testing.T) {
			result := Compile(fixtureSource(t, name+".pas"), name+".pas", semantic.HintsLevelPedantic)
			if got, want := result.DiagnosticStrings(), fixtureExpectation(t, name+".txt"); !reflect.DeepEqual(got, want) {
				t.Fatalf("got %q; want %q", got, want)
			}
		})
	}
}

func TestCompile_EnumBindingDeprecation(t *testing.T) {
	for _, tt := range []struct {
		name, first, second string
		wantLines           []int
	}{
		{"first deprecated", " deprecated 'first'", "", []int{5, 6}},
		{"second deprecated", "", " deprecated 'second'", []int{7}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := "type TEnum1 = (Hello" + tt.first + " = 1);\ntype TEnum2 = (Hello" + tt.second + " = 2);\nvar a1: TEnum1;\nvar a2: TEnum2;\nPrintLn(Ord(Hello));\nPrintLn(Ord(TEnum1.Hello));\nPrintLn(Ord(TEnum2.Hello));"
			result := Compile(source, "deprecation.pas", semantic.HintsLevelNormal)
			if !result.SemanticSuccessful {
				t.Fatal(result.DiagnosticStrings())
			}
			var lines []int
			for _, d := range result.Diagnostics {
				if strings.Contains(d.Message, "has been deprecated") {
					lines = append(lines, d.Line)
				}
			}
			if !reflect.DeepEqual(lines, tt.wantLines) {
				t.Fatalf("deprecated lines %v, want %v; diagnostics %q", lines, tt.wantLines, result.DiagnosticStrings())
			}
		})
	}
}

func TestCompile_EnumBindingCaseDuplicate(t *testing.T) {
	result := Compile(`type TEnum=(Hello, hello);`, "duplicate.pas", semantic.HintsLevelNormal)
	if result.SemanticSuccessful {
		t.Fatal("case-insensitive duplicate was accepted")
	}
}

func TestCompile_EnumBindingUnitParents(t *testing.T) {
	dir := t.TempDir()
	source := `unit AUnit; interface type TEnum1=(Hello=1); type TEnum2=(Hello=2); implementation end.`
	if err := os.WriteFile(filepath.Join(dir, "AUnit.pas"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	result := Compile(`uses AUnit; var value:TEnum1:=Hello;`, filepath.Join(dir, "Main.pas"), semantic.HintsLevelNormal)
	if !result.SemanticSuccessful {
		t.Fatalf("unit enum identity: %q", result.DiagnosticStrings())
	}
}

func TestCompile_EnumBindingImportedScopedMetadata(t *testing.T) {
	dir := t.TempDir()
	source := `unit AUnit; interface type TScoped=enum(Hello deprecated 'scoped'=5); type TFlags=flags(Hello deprecated 'flags'=8); implementation end.`
	if err := os.WriteFile(filepath.Join(dir, "AUnit.pas"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	result := Compile("uses AUnit;\nPrintLn(Ord(TScoped.Hello));\nPrintLn(Ord(TFlags.Hello));", filepath.Join(dir, "Main.pas"), semantic.HintsLevelNormal)
	if !result.SemanticSuccessful {
		t.Fatalf("compile: %q", result.DiagnosticStrings())
	}
	var warnings []string
	for _, d := range result.Diagnostics {
		if strings.Contains(d.Message, "has been deprecated") {
			warnings = append(warnings, d.Message)
		}
	}
	if len(warnings) != 2 || !strings.Contains(warnings[0], "scoped") || !strings.Contains(warnings[1], "flags") {
		t.Fatalf("owning metadata missing: %q", warnings)
	}
}

func TestCompile_EnumBindingUsesVisibility(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Imported.pas"), []byte(`unit Imported; interface type TImported=enum(Hello=7); type TClassic=(Bare=9); implementation end.`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name, source string
		success      bool
	}{
		{"bare before uses", `PrintLn(Ord(Bare)); uses Imported;`, false},
		{"qualified before uses", `PrintLn(Ord(Imported.Bare)); uses Imported;`, false},
		{"type before uses", `var v:TImported; uses Imported;`, false},
		{"body before uses", `procedure Early; begin PrintLn(Ord(Bare)); end; uses Imported; Early;`, false},
		{"qualified body before uses", `procedure Early; begin PrintLn(Ord(Imported.Bare)); end; uses Imported; Early;`, false},
		{"type body before uses", `procedure Early; begin var v:TImported; end; uses Imported; Early;`, false},
		{"local type before later import", `procedure Early; begin type TImported=(Local=1); PrintLn(Ord(Local)); end; uses Imported; Early;`, true},
		{"body after uses", `uses Imported; procedure Late; begin PrintLn(Ord(Bare)); PrintLn(Ord(TImported.Hello)); end; Late;`, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := Compile(tt.source, filepath.Join(dir, "Main.pas"), semantic.HintsLevelNormal)
			if got := result.SemanticSuccessful && !result.HasFatalDiagnostics(); got != tt.success {
				t.Fatalf("success=%v want %v; %q", got, tt.success, result.DiagnosticStrings())
			}
		})
	}
}

func TestCompile_ImportedConstantShadowsBuiltin(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Constants.pas"), []byte(`unit Constants; interface const Pi=42; implementation end.`), 0600); err != nil {
		t.Fatal(err)
	}
	result := Compile(`uses Constants; const chosen=Pi;`, filepath.Join(dir, "Main.pas"), semantic.HintsLevelNormal)
	if !result.SemanticSuccessful {
		t.Fatal(result.DiagnosticStrings())
	}
	symbol, ok := result.Analyzer.GetSymbolTable().Resolve("chosen")
	if !ok || symbol.Value != 42 {
		t.Fatalf("imported constant not selected: %+v", symbol)
	}
}

func TestCompile_EnumBindingUnitUsesVisibility(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Imported.pas"), []byte(`unit Imported; interface type TImported=(Bare=9); implementation end.`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name, body string
		success    bool
	}{
		{"body before unit uses", `procedure Early; begin PrintLn(Ord(Bare)); end; uses Imported;`, false},
		{"body after unit uses", `uses Imported; procedure Early; begin PrintLn(Ord(Bare)); end;`, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := "unit Consumer; interface " + tt.body + " implementation end."
			if err := os.WriteFile(filepath.Join(dir, "Consumer.pas"), []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			result := Compile(`uses Consumer;`, filepath.Join(dir, "Main.pas"), semantic.HintsLevelNormal)
			if got := result.SemanticSuccessful && !result.HasFatalDiagnostics(); got != tt.success {
				t.Fatalf("success=%v want %v; %q", got, tt.success, result.DiagnosticStrings())
			}
		})
	}
}

func TestCompile_EnumBindingDottedUnitEntries(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "A.B.pas"), []byte(`unit A.B; interface const Answer=42; implementation end.`), 0600); err != nil {
		t.Fatal(err)
	}
	// Both A and A.B are source unit identities. Together they shift the
	// midpoint from the second enum to the first, unlike a single placeholder.
	result := Compile(`type TEnum1=(Hello=1); type TEnum2=(Hello=2); var Z:Integer; uses A.B; var chosen:TEnum1:=Hello;`, filepath.Join(dir, "Main.pas"), semantic.HintsLevelNormal)
	if !result.SemanticSuccessful {
		t.Fatalf("namespace source table: %q", result.DiagnosticStrings())
	}
}

func TestCompile_EnumBindingFlattenedIdentity(t *testing.T) {
	result := Compile(`type TEnum1=(Hello=1); type TEnum2=(Hello=2); var a1:TEnum1; var a2:TEnum2; PrintLn(Ord(Hello));`, "identity.pas", semantic.HintsLevelNormal)
	if !result.SemanticSuccessful {
		t.Fatal(result.DiagnosticStrings())
	}
	table := result.Analyzer.GetSymbolTable()
	selected, ok := table.Resolve("Hello")
	if !ok || table.AllSymbols()["hello"] != selected {
		t.Fatal("flattened duplicate differs from canonical Resolve identity")
	}
}
