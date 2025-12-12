package interp_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInterpDoesNotImportEvaluator verifies the architectural boundary:
// internal/interp must not import internal/interp/evaluator (except in test files and wiring).
//
// This boundary exists to enable dependency inversion:
// - interp depends on contracts (interfaces)
// - evaluator implements those interfaces
// - runner wires them together
//
// See docs/architecture/interp-evaluator-boundary.md for rationale.
func TestInterpDoesNotImportEvaluator(t *testing.T) {
	const forbidden = "github.com/cwbudde/go-dws/internal/interp/evaluator"

	fset := token.NewFileSet()
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			// Skip evaluator package itself.
			// This test is about preventing *other* interp packages from importing evaluator.
			if path == "evaluator" {
				return fs.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files - they're allowed to import evaluator.
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Allow wiring layer to import evaluator.
		if path == filepath.Join("runner", "runner.go") || strings.HasPrefix(path, "runner"+string(filepath.Separator)) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		f, err := parser.ParseFile(fset, path, content, parser.ImportsOnly)
		if err != nil {
			return err
		}

		for _, imp := range f.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			if importPath == forbidden {
				t.Errorf("%s imports %s - violates interp/evaluator boundary.\n"+
					"Allowed only in runner/** and *_test.go.\n"+
					"Use contracts interfaces instead, or move wiring to runner.\n"+
					"See docs/architecture/interp-evaluator-boundary.md",
					path, forbidden)
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}
}
