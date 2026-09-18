package interp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

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
			if filepath.Base(req.Pas) != "obj_local.pas" {
				continue
			}
			got, detail := runFixtureTest(req.Pas, req.ExpectErrors, semantic.HintsLevel(req.Hints))
			if got != testResultPassed {
				t.Fatalf("Memory/obj_local: %v: %s", got, detail)
			}
			return
		}
	}
	t.Fatal("Memory/obj_local was not discovered")
}
