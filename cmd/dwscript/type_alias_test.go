package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestTypeAliases tests the CLI with type alias scripts
func TestTypeAliases(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../bin/dwscript", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build dwscript: %v", err)
	}

	binary := "../../bin/dwscript"

	tests := []struct {
		name         string
		scriptFile   string
		wantOutputs  []string // Strings that must appear in output
		wantExitCode int
	}{
		{
			name:       "Basic Type Aliases",
			scriptFile: "../../testdata/type_alias/basic_alias.dws",
			wantOutputs: []string{
				"User ID: 12345",
				"Filename: document.txt",
				"Price: 99.99",
				"Flag: true",
			},
			wantExitCode: 0,
		},
		{
			name:       "Type Alias Usage",
			scriptFile: "../../testdata/type_alias/alias_usage.dws",
			wantOutputs: []string{
				"ID1: 50",
				"ID2: 100",
				"Count: 50",
				"Total: 150",
			},
			wantExitCode: 0,
		},
		{
			name:       "Nested Type Aliases",
			scriptFile: "../../testdata/type_alias/nested_alias.dws",
			wantOutputs: []string{
				"a: 20",
				"b: 30",
				"c: 20",
				"sum: 70",
			},
			wantExitCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if script file exists
			if _, err := os.Stat(tt.scriptFile); os.IsNotExist(err) {
				t.Skipf("Script file not found: %s", tt.scriptFile)
			}

			// Run the CLI with the script
			cmd := exec.Command(binary, "run", tt.scriptFile)
			output, err := cmd.CombinedOutput()

			// Check exit code
			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				}
			}

			if exitCode != tt.wantExitCode {
				t.Errorf("Expected exit code %d, got %d. Output:\n%s",
					tt.wantExitCode, exitCode, string(output))
			}

			// Check for expected output strings
			outputStr := string(output)
			for _, want := range tt.wantOutputs {
				if !strings.Contains(outputStr, want) {
					t.Errorf("Expected output to contain %q, got:\n%s", want, outputStr)
				}
			}
		})
	}
}
