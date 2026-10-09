package main

import (
	"os/exec"
	"strings"
	"testing"
)

// TestMultiIndexCommaSyntax tests the CLI with multi-dimensional array comma syntax
func TestMultiIndexCommaSyntax(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	t.Run("Parse multi-index comma syntax", func(t *testing.T) {
		// Test that the parser can parse comma-separated array indices
		cmd := exec.Command(binary, "parse", "../../testdata/multi_index_comma.dws")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to parse multi_index_comma.dws: %v\nOutput: %s", err, string(output))
		}

		// Source output keeps comma groups distinct from successive brackets.
		outputStr := string(output)

		if !strings.Contains(outputStr, "value := matrix[0, 0];") {
			t.Errorf("Expected preserved 2D comma index in output: %s", outputStr)
		}

		if !strings.Contains(outputStr, "str := cube[0, 0, 0];") {
			t.Errorf("Expected preserved 3D comma index in output: %s", outputStr)
		}
	})

	t.Run("Parse Yin_and_yang.dws with comma syntax", func(t *testing.T) {
		cmd := exec.Command(binary, "parse", "../../examples/rosetta/Yin_and_yang.dws")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to parse Yin_and_yang.dws: %v\nOutput: %s", err, string(output))
		}

		outputStr := string(output)

		// Should not contain parser errors
		if strings.Contains(outputStr, "Parser errors") {
			t.Errorf("Yin_and_yang.dws should parse without errors, but got: %s", outputStr)
		}

		if !strings.Contains(outputStr, "Pix[x, y]") {
			t.Errorf("Expected preserved comma index for Pix array but not found")
		}
	})

	t.Run("Parse Levenshtein_distance.dws with comma syntax", func(t *testing.T) {
		cmd := exec.Command(binary, "parse", "../../examples/rosetta/Levenshtein_distance.dws")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to parse Levenshtein_distance.dws: %v\nOutput: %s", err, string(output))
		}

		outputStr := string(output)

		// Should not contain parser errors
		if strings.Contains(outputStr, "Parser errors") {
			t.Errorf("Levenshtein_distance.dws should parse without errors, but got: %s", outputStr)
		}

		if !strings.Contains(outputStr, "d[i, j]") {
			t.Errorf("Expected preserved comma index for d array but not found")
		}
	})

	t.Run("Preserve comma and bracket syntax", func(t *testing.T) {
		for _, want := range []string{"arr[i, j]", "arr[i][j]", "arr[i, j][k]"} {
			output, err := exec.Command(binary, "parse", "-e", want+";").CombinedOutput()
			if err != nil {
				t.Fatalf("parse %s: %v\n%s", want, err, output)
			}
			if got := strings.TrimSpace(string(output)); got != want {
				t.Errorf("printed %q; want %q", got, want)
			}
		}
	})

	t.Run("Equivalent array execution", func(t *testing.T) {
		const prefix = "var arr: array of array of Integer := [[1, 2], [3, 4]]; "
		for _, index := range []string{"arr[1, 0]", "arr[1][0]"} {
			output, err := exec.Command(binary, "run", "-e", prefix+"PrintLn("+index+");").CombinedOutput()
			if err != nil {
				t.Fatalf("run %s: %v\n%s", index, err, output)
			}
			if got := string(output); got != "3\n" {
				t.Errorf("run %s: output %q; want 3", index, got)
			}
		}
	})
}
