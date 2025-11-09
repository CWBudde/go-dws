package main

import (
	"os"
	"os/exec"
	"testing"
)

// TestLargeSet tests the CLI with large set operations (>64 elements)
//
// NOTE: These tests are currently skipped because the parser does not yet support
// "set of" type declarations in the type section. Once parser support is added,
// remove the t.Skip() calls and the tests should pass.
func TestLargeSet(t *testing.T) {
	t.Skip("Skipping: Parser does not yet support 'set of' type declarations. -> TODO")

	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"
	scriptFile := "../../testdata/sets/large_set.dws"
	expectedOutput := "../../testdata/sets/large_set.out"

	// Check if files exist
	if _, err := os.Stat(scriptFile); os.IsNotExist(err) {
		t.Fatalf("Script file does not exist: %s", scriptFile)
	}
	if _, err := os.Stat(expectedOutput); os.IsNotExist(err) {
		t.Fatalf("Expected output file does not exist: %s", expectedOutput)
	}

	// Run the script
	cmd := exec.Command(binary, "run", scriptFile)
	actualOutput, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run script: %v\nOutput: %s", err, string(actualOutput))
	}

	// Read expected output
	expectedBytes, err := os.ReadFile(expectedOutput)
	if err != nil {
		t.Fatalf("Failed to read expected output file: %v", err)
	}

	expected := string(expectedBytes)
	actual := string(actualOutput)

	// Compare outputs
	if actual != expected {
		t.Errorf("Output mismatch:\n=== EXPECTED ===\n%s\n=== ACTUAL ===\n%s\n=== END ===", expected, actual)
	}
}

// TestForInSet tests the CLI with for-in loop over sets
//
// NOTE: This test is currently skipped because the parser does not yet support
// "set of" type declarations in the type section. Once parser support is added,
// remove the t.Skip() call and the test should pass.
func TestForInSet(t *testing.T) {
	t.Skip("Skipping: Parser does not yet support 'set of' type declarations. -> TODO")

	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"
	scriptFile := "../../testdata/sets/for_in_set.dws"
	expectedOutput := "../../testdata/sets/for_in_set.out"

	// Check if files exist
	if _, err := os.Stat(scriptFile); os.IsNotExist(err) {
		t.Fatalf("Script file does not exist: %s", scriptFile)
	}
	if _, err := os.Stat(expectedOutput); os.IsNotExist(err) {
		t.Fatalf("Expected output file does not exist: %s", expectedOutput)
	}

	// Run the script
	cmd := exec.Command(binary, "run", scriptFile)
	actualOutput, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run script: %v\nOutput: %s", err, string(actualOutput))
	}

	// Read expected output
	expectedBytes, err := os.ReadFile(expectedOutput)
	if err != nil {
		t.Fatalf("Failed to read expected output file: %v", err)
	}

	expected := string(expectedBytes)
	actual := string(actualOutput)

	// Compare outputs
	if actual != expected {
		t.Errorf("Output mismatch:\n=== EXPECTED ===\n%s\n=== ACTUAL ===\n%s\n=== END ===", expected, actual)
	}
}
