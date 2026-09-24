package interp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cwbudde/go-dws/internal/fixtureconfig"
	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestFixtureInterfacePrivateDictionaryPolicy(t *testing.T) {
	path := filepath.Join(fixturesRoot, "InterfacesPass", "intf_private.pas")
	if got, detail := runFixtureTest(path, false, fixtureconfig.HintsLevel("InterfacesPass")); got != testResultPassed {
		t.Fatalf("intf_private: %v: %s", got, detail)
	}
}

func TestFixtureMissingExpectation(t *testing.T) {
	for _, tt := range []struct {
		name, category, source string
		want                   testResult
	}{
		{"silent", "SimpleScripts", "var x := 1;", testResultPassed},
		{"output", "SimpleScripts", "PrintLn('unexpected');", testResultFailed},
		{"compile error", "SimpleScripts", "missing;", testResultFailed},
		{"runtime error", "SimpleScripts", "raise Exception.Create('unexpected');", testResultFailed},
		{"missing diagnostics expectation", "FailureScripts", "", testResultSkipped},
		{"build runner", "BuildScripts", "", testResultSkipped},
		{"format runner", "AutoFormat", "", testResultSkipped},
		{"external host", "External", "", testResultSkipped},
		{"delegate host", "DelegateLib", "", testResultSkipped},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), tt.category)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, "test.pas")
			if err := os.WriteFile(path, []byte(tt.source), 0o644); err != nil {
				t.Fatal(err)
			}
			got, detail := runFixtureTest(path, isErrorCategory(tt.category), semantic.HintsLevelNormal)
			if got != tt.want {
				t.Fatalf("result = %v, want %v: %s", got, tt.want, detail)
			}
		})
	}
}

func TestFixtureMemoryHintLevel(t *testing.T) {
	categories, err := discoverFixtureCategories(fixturesRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, category := range categories {
		if category.name != "Memory" {
			continue
		}
		for _, req := range buildFixtureWorkList([]fixtureCategory{category}) {
			if filepath.Base(req.Source) != "obj_local.pas" {
				continue
			}
			got, detail := runFixtureTest(req.Source, req.ExpectErrors, semantic.HintsLevel(req.Hints))
			if got != testResultPassed {
				t.Fatalf("Memory/obj_local: %v: %s", got, detail)
			}
			return
		}
	}
	t.Fatal("Memory/obj_local was not discovered")
}

func TestFixtureBuildScriptsDrivers(t *testing.T) {
	categories, err := discoverFixtureCategories(fixturesRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, category := range categories {
		if category.name != "BuildScripts" {
			continue
		}
		if len(category.sourceFiles) != 51 {
			t.Fatalf("BuildScripts source count = %d, want 51 drivers", len(category.sourceFiles))
		}
		for _, source := range category.sourceFiles {
			if filepath.Ext(source) != ".dws" {
				t.Fatalf("support unit selected as driver: %s", source)
			}
		}
		for _, name := range []string{"const_inline", "conditionals_main"} {
			path := filepath.Join(fixturesRoot, "BuildScripts", name+".dws")
			if got, detail := runFixtureTest(path, false, fixtureconfig.HintsLevel("BuildScripts")); got != testResultPassed {
				t.Fatalf("%s driver: %v: %s", name, got, detail)
			}
		}
		return
	}
	t.Fatal("BuildScripts category missing")
}
