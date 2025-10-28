package dwscript_test

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/pkg/dwscript"
)

// TestRosettaExamplesParse ensures all Rosetta DWScript samples at least parse successfully.
func TestRosettaExamplesParse(t *testing.T) {
	t.Helper()

	scriptDir := filepath.Join("..", "..", "examples", "rosetta")
	dirEntries, err := os.ReadDir(scriptDir)
	if err != nil {
		t.Fatalf("failed to read Rosetta examples directory %s: %v", scriptDir, err)
	}

	if len(dirEntries) == 0 {
		t.Fatalf("expected Rosetta examples directory %s to contain scripts", scriptDir)
	}

	for _, entry := range dirEntries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".dws" {
			continue
		}

		entry := entry
		testName := strings.TrimSuffix(entry.Name(), ".dws")

		t.Run(testName, func(t *testing.T) {
			source, readErr := os.ReadFile(filepath.Join(scriptDir, entry.Name()))
			if readErr != nil {
				t.Fatalf("failed to read script %s: %v", entry.Name(), readErr)
			}

			engine, newErr := dwscript.New(
				dwscript.WithTypeCheck(false),
				dwscript.WithOutput(io.Discard),
			)
			if newErr != nil {
				t.Fatalf("failed to create engine: %v", newErr)
			}

			if _, compileErr := engine.Compile(string(source)); compileErr != nil {
				t.Fatalf("failed to compile %s: %v", entry.Name(), compileErr)
			}
		})
	}
}
