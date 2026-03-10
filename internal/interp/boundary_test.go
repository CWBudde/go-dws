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

func readBoundarySource(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s failed: %v", path, err)
	}
	return string(content)
}

// TestInterpDoesNotImportEvaluator verifies the architectural boundary:
// internal/interp must not import internal/interp/evaluator except for the
// canonical construction entry point and test files.
//
// This boundary exists to enable dependency inversion:
// - interp depends on contracts (interfaces)
// - evaluator implements those interfaces
// - interp/new.go owns canonical runtime construction during Phase 4
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

		// Allow canonical construction entry points to import evaluator.
		if path == "new.go" || path == filepath.Join("runner", "runner.go") || strings.HasPrefix(path, "runner"+string(filepath.Separator)) {
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
					"Allowed only in new.go, runner/**, and *_test.go.\n"+
					"Use contracts interfaces for all non-construction code.\n"+
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

func TestConstructionDoesNotReferenceLegacyBridgeWiring(t *testing.T) {
	t.Parallel()

	checks := []struct {
		path      string
		forbidden []string
	}{
		{
			path:      "new.go",
			forbidden: []string{"SetFocusedInterfaces", "SetRuntimeBridge", "SetEnvironment", "RestoreEnvironment"},
		},
		{
			path:      filepath.Join("runner", "runner.go"),
			forbidden: []string{"SetFocusedInterfaces", "SetRuntimeBridge", "SetEnvironment", "RestoreEnvironment"},
		},
	}

	for _, tc := range checks {
		source := readBoundarySource(t, tc.path)
		for _, needle := range tc.forbidden {
			if strings.Contains(source, needle) {
				t.Fatalf("%s still references legacy construction wiring %q", tc.path, needle)
			}
		}
	}
}

func TestNoLegacyCallbackInterfaceDeclarationsRemain(t *testing.T) {
	t.Parallel()

	type forbiddenDecl struct {
		needle string
		label  string
	}

	forbidden := []forbiddenDecl{
		{needle: "type CoreEvaluator interface", label: "CoreEvaluator"},
		{needle: "type OOPEngine interface", label: "OOPEngine"},
		{needle: "type DeclHandler interface", label: "DeclHandler"},
		{needle: "func (e *Evaluator) SetFocusedInterfaces(", label: "SetFocusedInterfaces"},
	}

	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		source := readBoundarySource(t, path)
		for _, decl := range forbidden {
			if strings.Contains(source, decl.needle) {
				t.Errorf("%s still declares legacy callback surface %s", path, decl.label)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}
}
