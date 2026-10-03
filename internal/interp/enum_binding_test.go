package interp

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/frontend"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestEnumBindingSourceIdentity(t *testing.T) {
	const enums = `type TEnum1 = (Hello = 1); type TEnum2 = (Hello = 2); `
	tests := []struct{ name, source, want string }{
		{"locals", enums + `var a1: TEnum1; var a2: TEnum2; PrintLn(Ord(Hello)); PrintLn(Ord(TEnum1.Hello)); PrintLn(Ord(TEnum2.Hello));`, "1\n1\n2\n"},
		{"routines", enums + `procedure Test1; begin end; procedure Test2; begin end; PrintLn(Ord(Hello)); PrintLn(Ord(TEnum1.Hello)); PrintLn(Ord(TEnum2.Hello));`, "2\n1\n2\n"},
		// These expectations are source-derived predictions from the pinned upstream
		// symbol algorithm, unlike the two checked-in diagnostic fixture oracles.
		{"interleaved", enums + `PrintLn(Ord(Hello)); var a1: TEnum1; PrintLn(Ord(Hello)); var a2: TEnum1; PrintLn(Ord(Hello)); var a3: TEnum1; PrintLn(Ord(Hello));`, "2\n1\n1\n2\n"},
		{"variable storage", `type TEnum = (One = 1, Two); var v := One; v := Two; PrintLn(Ord(v));`, "2\n"},
		{"deferred body", enums + `procedure Test1; begin PrintLn(Ord(Hello)); end; var a1: TEnum1; var a2: TEnum1; Test1; PrintLn(Ord(Hello));`, "2\n1\n"},
		{"explicit enum ordinal", enums + `var a1: TEnum1; var a2: TEnum1; type TNext = (Next = Ord(Hello)); PrintLn(Ord(Next));`, "1\n"},
		{"type name shadow", `type TEnum1 = (Hello = 1); type TEnum2 = (Hello = 2); procedure P(TEnum1: TEnum2); begin PrintLn(Ord(TEnum1.Hello)); end; P(TEnum2.Hello);`, "2\n"},
		{"inner enum", enums + `procedure P; begin type TLocal=(Hello=3); PrintLn(Ord(Hello)); end; P; PrintLn(Ord(TEnum1.Hello));`, "3\n1\n"},
		{"constant", enums + `var a1: TEnum1; var a2: TEnum1; const chosen=Hello; PrintLn(Ord(chosen));`, "1\n"},
		{"constant array", enums + `var a1: TEnum1; var a2: TEnum1; const values: array of TEnum1=[Hello]; PrintLn(Ord(values[0]));`, "1\n"},
		{"class source order", enums + `PrintLn(Ord(Hello)); type AClass=class end; PrintLn(Ord(Hello));`, "2\n1\n"},
		{"alias source order", enums + `PrintLn(Ord(Hello)); type AAlias=TEnum1; PrintLn(Ord(Hello));`, "2\n1\n"},
		{"scoped", `type TScoped=enum(Hello=5); type TFlags=flags(Hello=8); PrintLn(Ord(TScoped.Hello)); PrintLn(Ord(TFlags.Hello));`, "5\n8\n"},
		{"case collisions", `type TEnum1=(Hello=1); type TEnum2=(hELLo=2); var a1: TEnum1; var a2: TEnum1; PrintLn(Ord(HELLO));`, "1\n"},
		{"overload registrations", enums + `procedure AProc(x: Integer); overload; begin end; PrintLn(Ord(Hello)); procedure AProc(x: String); overload; begin end; PrintLn(Ord(Hello)); procedure AProc(x: Boolean); overload; begin end; PrintLn(Ord(Hello));`, "1\n1\n2\n"},
		{"forward identity", enums + `procedure AProc; forward; PrintLn(Ord(Hello)); procedure AProc; begin end; PrintLn(Ord(Hello));`, "1\n1\n"},
		{"three duplicates", `type TEnum1=(Hello=1); type TEnum2=(Hello=2); type TEnum3=(Hello=3); PrintLn(Ord(Hello)); var a1:Integer; PrintLn(Ord(Hello)); var a2:Integer; PrintLn(Ord(Hello));`, "3\n1\n1\n"},
		{"inline body snapshot", enums + `type TZ=class class procedure P; begin PrintLn(Ord(Hello)); end; end; type AA=class end; type BB=class end; TZ.P; PrintLn(Ord(Hello));`, "2\n1\n"},
		{"parameter shadow", enums + `procedure P(Hello: Integer); begin PrintLn(Hello); end; P(42);`, "42\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiled := frontend.Compile(tt.source, "enum_binding.pas", semantic.HintsLevelNormal)
			if !compiled.SemanticSuccessful || compiled.HasFatalDiagnostics() {
				t.Fatalf("compile: %s", strings.Join(compiled.DiagnosticStrings(), "\n"))
			}
			for run := 0; run < 2; run++ {
				var output bytes.Buffer
				engine := New(&output)
				engine.SetSemanticInfo(compiled.SemanticInfo)
				result := engine.Eval(compiled.Program)
				if result != nil && result.Type() == "ERROR" {
					t.Fatal(result.String())
				}
				if got := output.String(); got != tt.want {
					t.Fatalf("run %d: got %q; want %q", run, got, tt.want)
				}
			}
		})
	}
}

func TestEnumBindingImportedExecution(t *testing.T) {
	dir := t.TempDir()
	source := `unit AUnit; interface type TEnum1=(Hello=1); type TEnum2=(Hello=2); type TScoped=enum(Other=5); implementation end.`
	if err := os.WriteFile(filepath.Join(dir, "AUnit.pas"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	compiled := frontend.Compile(`uses AUnit; PrintLn(Ord(TEnum1.Hello)); PrintLn(Ord(TEnum2.Hello)); PrintLn(Ord(TScoped.Other));`, filepath.Join(dir, "Main.pas"), semantic.HintsLevelNormal)
	if !compiled.SemanticSuccessful {
		t.Fatal(compiled.DiagnosticStrings())
	}
	for range 2 {
		var output bytes.Buffer
		engine := New(&output)
		engine.SetSemanticInfo(compiled.SemanticInfo)
		prepareEnumBindingUnits(t, engine, compiled)
		result := engine.Eval(compiled.Program)
		if result != nil && result.Type() == "ERROR" {
			t.Fatal(result.String())
		}
		if got := output.String(); got != "1\n2\n5\n" {
			t.Fatalf("got %q", got)
		}
	}
}

// Unchecked evaluation has no source binding metadata. Its compatibility path
// deliberately reads the environment, whose last declaration owns a bare name.
func TestEnumBindingUncheckedFallback(t *testing.T) {
	_, output := testEvalWithOutput(`type TEnum1=(Hello=1); type TEnum2=(Hello=2); var a1:TEnum1; var a2:TEnum1; PrintLn(Ord(Hello)); PrintLn(Ord(TEnum1.Hello));`)
	if output != "2\n1\n" {
		t.Fatalf("unchecked environment fallback: %q", output)
	}
}

func TestEnumBindingUsesSourceEvent(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AUnit.pas"), []byte(`unit AUnit; interface type TImported=(Pi=42, ImportedValue=7); implementation end.`), 0600); err != nil {
		t.Fatal(err)
	}
	source := `type TEnum1=(Hello=1); type TEnum2=(Hello=2); procedure Test1; begin PrintLn(Ord(Hello)); end; PrintLn(Ord(Hello)); uses AUnit; procedure ATest2; begin PrintLn(Ord(Hello)); PrintLn(Ord(Pi)); end; Test1; ATest2; PrintLn(Ord(ImportedValue));`
	compiled := frontend.Compile(source, filepath.Join(dir, "Main.pas"), semantic.HintsLevelNormal)
	if !compiled.SemanticSuccessful {
		t.Fatal(compiled.DiagnosticStrings())
	}
	for range 2 {
		var output bytes.Buffer
		engine := New(&output)
		engine.SetSemanticInfo(compiled.SemanticInfo)
		prepareEnumBindingUnits(t, engine, compiled)
		result := engine.Eval(compiled.Program)
		if result != nil && result.Type() == "ERROR" {
			t.Fatal(result.String())
		}
		if got, want := output.String(), "2\n2\n1\n42\n7\n"; got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	}
}

func TestEnumBindingImportedParentOrder(t *testing.T) {
	dir := t.TempDir()
	for name, source := range map[string]string{
		"U1": `unit U1; interface type TFirst=(Hello=1); implementation end.`,
		"U2": `unit U2; interface type TSecond=(Hello=2); implementation end.`,
	} {
		if err := os.WriteFile(filepath.Join(dir, name+".pas"), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	compiled := frontend.Compile(`uses U1; procedure Before; begin PrintLn(Ord(Hello)); end; uses U2; procedure After; begin PrintLn(Ord(Hello)); end; Before; After; uses U1; PrintLn(Ord(Hello)); PrintLn(Ord(U2.Hello));`, filepath.Join(dir, "Main.pas"), semantic.HintsLevelNormal)
	if !compiled.SemanticSuccessful {
		t.Fatal(compiled.DiagnosticStrings())
	}
	for range 2 {
		var output bytes.Buffer
		engine := New(&output)
		engine.SetSemanticInfo(compiled.SemanticInfo)
		prepareEnumBindingUnits(t, engine, compiled)
		result := engine.Eval(compiled.Program)
		if result != nil && result.Type() == "ERROR" {
			t.Fatal(result.String())
		}
		if got, want := output.String(), "1\n2\n1\n2\n"; got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	}
}

func TestEnumBindingReviewedUnitScopes(t *testing.T) {
	for _, tt := range []struct{ name, consumer, main, want string }{
		{"interface body excludes private enum", `unit Consumer; interface type TPublic=(Hello=1); procedure Early; begin var x:TPublic:=Hello; PrintLn(Ord(x)); end; implementation type TPrivate=(Hello=2); end.`, `uses Consumer;`, ""},
		{"implementation body excludes later enum", `unit Consumer; interface type TPublic=(Hello=1); procedure Early; implementation procedure Early; begin PrintLn(Ord(Hello)); end; type TPrivate=(Hello=2); end.`, `uses Consumer; Early();`, "1\n"},
		{"implementation import precedes builtin", `unit Consumer; interface procedure P; implementation uses A; procedure P; begin var x:TImported:=Pi; PrintLn(Ord(x)); end; end.`, `uses Consumer; P();`, "42\n"},
		{"own interface precedes implementation import", `unit Consumer; interface type TOwn=(Pi=1); procedure P; implementation uses A; procedure P; begin var x:TOwn:=Pi; PrintLn(Ord(x)); end; end.`, `uses Consumer; P();`, "1\n"},
		{"unit enum dependencies select later parent", `unit Consumer; interface uses U1,U2; procedure P; implementation procedure P; begin var x:TSecond:=Hello; PrintLn(Ord(x)); end; end.`, `uses Consumer; P();`, "2\n"},
		{"parameter shadows unit", `unit Consumer; interface type TRemote=(Hello=1); implementation end.`, `uses Consumer; type TLocal=(Hello=2); procedure P(Consumer:TLocal); begin var x:TLocal:=Consumer.Hello; PrintLn(Ord(x)); end; P(TLocal.Hello);`, "2\n"},
		{"local shadows unit", `unit Consumer; interface type TRemote=(Hello=1); implementation end.`, `uses Consumer; type TLocal=(Hello=2); procedure P; begin var Consumer:=TLocal.Hello; PrintLn(Ord(Consumer.Hello)); end; P;`, "2\n"},
		{"type alias shadows unit", `unit Consumer; interface type TRemote=(Hello=1); implementation end.`, `uses Consumer; type TLocal=(Hello=2); procedure P; begin type Consumer=TLocal; PrintLn(Ord(Consumer.Hello)); end; P;`, "2\n"},
		{"type shadows unit", `unit Consumer; interface type TRemote=(Hello=1); implementation end.`, `uses Consumer; procedure P; begin type Consumer=enum(Hello=2); PrintLn(Ord(Consumer.Hello)); end; P;`, "2\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, source := range map[string]string{"Consumer": tt.consumer, "A": `unit A; interface type TImported=(Pi=42); implementation end.`, "U1": `unit U1; interface type TFirst=(Hello=1); implementation end.`, "U2": `unit U2; interface type TSecond=(Hello=2); implementation end.`} {
				if err := os.WriteFile(filepath.Join(dir, name+".pas"), []byte(source), 0600); err != nil {
					t.Fatal(err)
				}
			}
			compiled := frontend.Compile(tt.main, filepath.Join(dir, "Main.pas"), semantic.HintsLevelNormal)
			if !compiled.SemanticSuccessful {
				t.Fatal(compiled.DiagnosticStrings())
			}
			if tt.want == "" {
				unit, _ := compiled.UnitRegistry.GetUnit("Consumer")
				matched := false
				ast.Inspect(unit.Declaration, func(node ast.Node) bool {
					if expression, ok := node.(ast.Expression); ok {
						if binding, ok := compiled.SemanticInfo.EnumElementBinding(expression); ok && binding.Name == "Hello" {
							if binding.EnumType.Name != "TPublic" || binding.Ordinal != 1 {
								t.Fatalf("wrong source binding: %+v", binding)
							}
							matched = true
						}
					}
					return true
				})
				if !matched {
					t.Fatal("public body enum binding missing")
				}
				return
			}
			var output bytes.Buffer
			engine := New(&output)
			engine.SetSemanticInfo(compiled.SemanticInfo)
			prepareEnumBindingUnits(t, engine, compiled)
			result := engine.Eval(compiled.Program)
			if result != nil && result.Type() == "ERROR" {
				t.Fatal(result.String())
			}
			if got := output.String(); got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestEnumBindingDuplicateOrdinalConsumers(t *testing.T) {
	for _, tt := range []struct{ name, source, want string }{
		{"subrange array bounds", `type TEnum1=(Hello=1); type TEnum2=(Hello=2); var a1:Integer; var a2:Integer; type TBounds=0+Ord(Hello)..Ord(Hello); var a:array[TBounds] of Integer; PrintLn(Low(a)); PrintLn(High(a));`, "1\n1\n"},
		{"set membership", `type TEnum1=(Hello=1); type TEnum2=(Hello=2); type TSet=set of TEnum1; var a1:Integer; var a2:Integer; var chosen:TSet:=[Hello]; PrintLn(TEnum1.Hello in chosen);`, "True\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := runForLoopRegressionScript(t, tt.source); got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

// prepareEnumBindingUnits mirrors the embedding API's unit bootstrap: importing
// declarations is separate from executing initialization sections.
func prepareEnumBindingUnits(t *testing.T, engine *Interpreter, compiled *frontend.Result) {
	t.Helper()
	registry := compiled.UnitRegistry.CloneForExecution()
	engine.SetUnitRegistry(registry)
	order, err := registry.ComputeInitializationOrder()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range order {
		unit, _ := registry.GetUnit(name)
		if err := engine.ImportUnitSymbols(unit); err != nil {
			t.Fatal(err)
		}
	}
	if err := engine.InitializeUnits(); err != nil {
		t.Fatal(err)
	}
}
