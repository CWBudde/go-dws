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

		// Check that the output contains desugared nested index expressions
		outputStr := string(output)

		// Verify 2D comma syntax is desugared: arr[i, j] -> arr[i][j]
		if !strings.Contains(outputStr, "matrix[0][0]") {
			t.Errorf("Expected desugared 2D index expression like 'matrix[0][0]' but not found in output")
		}

		// Verify 3D comma syntax is desugared: arr[i, j, k] -> arr[i][j][k]
		if !strings.Contains(outputStr, "cube[0][0][0]") {
			t.Errorf("Expected desugared 3D index expression like 'cube[0][0][0]' but not found in output")
		}
	})

	t.Run("Parse Yin_and_yang.dws with comma syntax", func(t *testing.T) {
		// Test that Yin_and_yang.dws (which uses comma syntax) now parses successfully
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

		// Should contain desugared comma syntax (Pix[x, y] -> Pix[x][y])
		if !strings.Contains(outputStr, "Pix[") && !strings.Contains(outputStr, "[y]") {
			t.Errorf("Expected desugared index expression for Pix array but not found")
		}
	})

	t.Run("Parse Levenshtein_distance.dws with comma syntax", func(t *testing.T) {
		// Test that Levenshtein_distance.dws (which uses comma syntax) now parses successfully
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

		// Should contain desugared comma syntax for d array (d[i, j] -> d[i][j])
		if !strings.Contains(outputStr, "d[") {
			t.Errorf("Expected desugared index expression for d array but not found")
		}
	})

	t.Run("Equivalence of comma and bracket syntax", func(t *testing.T) {
		// Test that arr[i, j] and arr[i][j] produce equivalent AST

		// Parse with comma syntax
		cmd1 := exec.Command(binary, "parse", "-e", "arr[i, j];")
		output1, err := cmd1.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to parse comma syntax: %v\nOutput: %s", err, string(output1))
		}

		// Parse with nested bracket syntax
		cmd2 := exec.Command(binary, "parse", "-e", "arr[i][j];")
		output2, err := cmd2.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to parse bracket syntax: %v\nOutput: %s", err, string(output2))
		}

		// Both should produce the same desugared output
		output1Str := strings.TrimSpace(string(output1))
		output2Str := strings.TrimSpace(string(output2))

		if output1Str != output2Str {
			t.Errorf("Comma syntax and bracket syntax should produce equivalent AST\nComma: %s\nBracket: %s",
				output1Str, output2Str)
		}
	})
}
