package dwscript_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/pkg/dwscript"
)

// TestExampleScripts ensures the canonical playground examples continue to compile and run.
func TestExampleScripts(t *testing.T) {
	t.Helper()

	scriptDir := filepath.Join("..", "..", "examples", "scripts")
	if _, err := os.Stat(scriptDir); err != nil {
		t.Fatalf("expected scripts directory %s to exist: %v", scriptDir, err)
	}

	scripts := []struct {
		file     string
		expected string
	}{
		{"hello_world.dws", "Hello, World!"},
		{"fibonacci.dws", "F(9) = 34"},
		{"factorial.dws", "Iterative: 3628800"},
		{"loops.dws", "Repeat-until loop"},
		{"functions.dws", "Capitalized: Hello"},
		{"classes.dws", "is now 26 years old"},
		{"math_operations.dws", "a := a * 2 -> 24"},
		{"case_statement.dws", "Day 1 -> Monday"},
		{"palindrome_checker.dws", "dwscript is not a palindrome"},
		{"prime_numbers.dws", "Prime numbers up to 30"},
		{"multiplication_table.dws", "Multiplication table (5 x 5):"},
	}

	for _, script := range scripts {
		script := script
		name := strings.TrimSuffix(script.file, filepath.Ext(script.file))

		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer

			source, err := os.ReadFile(filepath.Join(scriptDir, script.file))
			if err != nil {
				t.Fatalf("failed to read %s: %v", script.file, err)
			}

			engine, err := dwscript.New(
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
				t.Fatalf("script %s reported unsuccessful execution", script.file)
			}

			if script.expected != "" && !strings.Contains(result.Output, script.expected) {
				t.Fatalf("script %s output missing %q\n---\n%s\n---", script.file, script.expected, result.Output)
			}

			buf.Reset()
		})
	}
}
