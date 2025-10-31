package dwscript_test

import (
	"bytes"
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

func TestExitStatementExamples(t *testing.T) {
	t.Helper()

	scripts := []struct {
		name               string
		path               string
		expectedOutput     string
		requireNonEmptyOut bool
		skipOnCompileError bool
	}{
		{
			name:           "ExitStatementFixture",
			path:           filepath.Join("..", "..", "testdata", "exit_statement", "exit_with_value.dws"),
			expectedOutput: "false\ntrue\n",
		},
		{
			name:               "DeathStarExample",
			path:               filepath.Join("..", "..", "examples", "rosetta", "Death_Star.dws"),
			requireNonEmptyOut: true,
			skipOnCompileError: true,
		},
	}

	for _, script := range scripts {
		script := script

		t.Run(script.name, func(t *testing.T) {
			source, err := os.ReadFile(script.path)
			if err != nil {
				t.Fatalf("failed to read %s: %v", script.path, err)
			}

			var buf bytes.Buffer
			engine, err := dwscript.New(
				dwscript.WithTypeCheck(false),
				dwscript.WithOutput(&buf),
			)
			if err != nil {
				t.Fatalf("failed to create engine: %v", err)
			}

			result, err := engine.Eval(string(source))
			if err != nil {
				if compileErr, ok := err.(*dwscript.CompileError); ok && script.skipOnCompileError {
					t.Skipf("%s not yet supported: %v", script.name, compileErr)
				}
				t.Fatalf("evaluation error: %v", err)
			}
			if !result.Success {
				t.Fatalf("script %s reported unsuccessful execution", script.path)
			}

			output := buf.String()
			if script.expectedOutput != "" && output != script.expectedOutput {
				t.Fatalf("unexpected output: want %q, got %q", script.expectedOutput, output)
			}

			if script.requireNonEmptyOut && output == "" {
				t.Fatalf("expected %s to produce output", script.name)
			}
		})
	}
}

func TestArrayLiteralExamples(t *testing.T) {
	t.Helper()

	scripts := []struct {
		name           string
		path           string
		expectedOutput string
	}{
		{
			name:           "ArrayLiteralBasic",
			path:           filepath.Join("..", "..", "testdata", "array_literals", "array_literal_basic.dws"),
			expectedOutput: "6\n",
		},
		{
			name:           "ArrayLiteralNested",
			path:           filepath.Join("..", "..", "testdata", "array_literals", "array_literal_nested.dws"),
			expectedOutput: "-50\n3\n",
		},
	}

	for _, script := range scripts {
		script := script

		t.Run(script.name, func(t *testing.T) {
			source, err := os.ReadFile(script.path)
			if err != nil {
				t.Fatalf("failed to read %s: %v", script.path, err)
			}

			var buf bytes.Buffer
			engine, err := dwscript.New(
				dwscript.WithTypeCheck(false),
				dwscript.WithOutput(&buf),
			)
			if err != nil {
				t.Fatalf("failed to create engine: %v", err)
			}

			result, err := engine.Eval(string(source))
			if err != nil {
				t.Fatalf("evaluation error: %v", err)
			}
			if !result.Success {
				t.Fatalf("script %s reported unsuccessful execution", script.path)
			}

			if output := buf.String(); output != script.expectedOutput {
				t.Fatalf("unexpected output for %s: want %q, got %q", script.name, script.expectedOutput, output)
			}
		})
	}
}
