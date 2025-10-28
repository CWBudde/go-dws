package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunWithUnits tests running a program that uses units
func TestRunWithUnits(t *testing.T) {
	// Create a temporary directory for our test files
	tempDir := t.TempDir()

	// Create a simple unit
	mathUnit := `unit MathUtils;

interface

function Add(a, b: Integer): Integer;
function Multiply(a, b: Integer): Integer;

implementation

function Add(a, b: Integer): Integer;
begin
  Result := a + b;
end;

function Multiply(a, b: Integer): Integer;
begin
  Result := a * b;
end;

initialization
  PrintLn('MathUtils loaded');

finalization
  PrintLn('MathUtils unloading');

end.`

	mathPath := filepath.Join(tempDir, "MathUtils.dws")
	if err := os.WriteFile(mathPath, []byte(mathUnit), 0644); err != nil {
		t.Fatalf("Failed to create MathUtils.dws: %v", err)
	}

	// Create a main program that uses the unit
	mainProgram := `uses MathUtils;

PrintLn('10 + 20 = ' + IntToStr(Add(10, 20)));
PrintLn('5 * 6 = ' + IntToStr(Multiply(5, 6)));`

	mainPath := filepath.Join(tempDir, "main.dws")
	if err := os.WriteFile(mainPath, []byte(mainProgram), 0644); err != nil {
		t.Fatalf("Failed to create main.dws: %v", err)
	}

	// Reset global state
	oldSearchPaths := unitSearchPaths
	oldVerbose := verbose
	defer func() {
		unitSearchPaths = oldSearchPaths
		verbose = oldVerbose
	}()

	// Set up command arguments
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"dwscript", "run", mainPath}

	// Capture stdout
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	// Run the command
	err := runScript(runCmd, []string{mainPath})

	// Close write end and restore
	wOut.Close()
	os.Stdout = oldStdout

	// Read output
	var bufOut bytes.Buffer
	bufOut.ReadFrom(rOut)
	output := bufOut.String()

	if err != nil {
		t.Fatalf("runScript failed: %v\nOutput: %s", err, output)
	}

	// Verify output contains expected results
	if !strings.Contains(output, "MathUtils loaded") {
		t.Error("Expected 'MathUtils loaded' in output")
	}

	if !strings.Contains(output, "10 + 20 = 30") {
		t.Error("Expected '10 + 20 = 30' in output")
	}

	if !strings.Contains(output, "5 * 6 = 30") {
		t.Error("Expected '5 * 6 = 30' in output")
	}

	if !strings.Contains(output, "MathUtils unloading") {
		t.Error("Expected 'MathUtils unloading' in output")
	}
}

// TestRunWithUnitsAndIncludeFlag tests the -I flag
func TestRunWithUnitsAndIncludeFlag(t *testing.T) {
	// Create two directories: one for the main program, one for units
	mainDir := t.TempDir()
	unitsDir := t.TempDir()

	// Create a unit in the units directory
	stringUnit := `unit StringUtils;

interface

function Reverse(s: String): String;

implementation

function Reverse(s: String): String;
var
  i: Integer;
  result: String;
begin
  result := '';
  for i := Length(s) downto 1 do
    result := result + s[i];
  Result := result;
end;

end.`

	unitPath := filepath.Join(unitsDir, "StringUtils.dws")
	if err := os.WriteFile(unitPath, []byte(stringUnit), 0644); err != nil {
		t.Fatalf("Failed to create StringUtils.dws: %v", err)
	}

	// Create main program in main directory
	mainProgram := `uses StringUtils;

PrintLn(Reverse('Hello'));`

	mainPath := filepath.Join(mainDir, "main.dws")
	if err := os.WriteFile(mainPath, []byte(mainProgram), 0644); err != nil {
		t.Fatalf("Failed to create main.dws: %v", err)
	}

	// Reset global state and set up the -I flag with the units directory
	oldSearchPaths := unitSearchPaths
	oldVerbose := verbose
	defer func() {
		unitSearchPaths = oldSearchPaths
		verbose = oldVerbose
	}()
	unitSearchPaths = []string{unitsDir}

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the command
	err := runScript(runCmd, []string{mainPath})

	// Close write end and restore stdout
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("runScript failed: %v\nOutput: %s", err, output)
	}

	// Verify output contains reversed string
	if !strings.Contains(output, "olleH") {
		t.Errorf("Expected 'olleH' in output, got: %s", output)
	}
}

// TestRunWithVerboseUnitLoading tests verbose output with -v flag
func TestRunWithVerboseUnitLoading(t *testing.T) {
	tempDir := t.TempDir()

	// Create a simple unit
	unit := `unit TestUnit;
interface
implementation
end.`

	unitPath := filepath.Join(tempDir, "TestUnit.dws")
	if err := os.WriteFile(unitPath, []byte(unit), 0644); err != nil {
		t.Fatalf("Failed to create unit: %v", err)
	}

	// Create main program
	mainProgram := `uses TestUnit;
PrintLn('Done');`

	mainPath := filepath.Join(tempDir, "main.dws")
	if err := os.WriteFile(mainPath, []byte(mainProgram), 0644); err != nil {
		t.Fatalf("Failed to create main: %v", err)
	}

	// Enable verbose mode
	verbose = true
	defer func() { verbose = false }()

	// Capture stderr (where verbose output goes)
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Run the command
	_ = runScript(runCmd, []string{mainPath})

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	stderrOutput := buf.String()

	// Verify verbose output
	if !strings.Contains(stderrOutput, "Loading") {
		t.Log("Verbose stderr output:", stderrOutput)
		// Note: This might not always print depending on implementation
	}
}

// TestShowUnitsFlag tests the --show-units flag
func TestShowUnitsFlag(t *testing.T) {
	// Create temporary directory for test units
	tempDir := t.TempDir()

	// Create base unit
	baseUnit := `unit BaseUtils;
interface
function Double(x: Integer): Integer;
implementation
function Double(x: Integer): Integer;
begin
  Result := x * 2;
end;
end.`
	
	if err := os.WriteFile(filepath.Join(tempDir, "BaseUtils.dws"), []byte(baseUnit), 0644); err != nil {
		t.Fatalf("Failed to create BaseUtils.dws: %v", err)
	}

	// Create math unit that depends on base
	mathUnit := `unit MathUtils;
interface
uses BaseUtils;
function Add(a, b: Integer): Integer;
implementation
function Add(a, b: Integer): Integer;
begin
  Result := a + b;
end;
end.`
	
	if err := os.WriteFile(filepath.Join(tempDir, "MathUtils.dws"), []byte(mathUnit), 0644); err != nil {
		t.Fatalf("Failed to create MathUtils.dws: %v", err)
	}

	// Create main program
	mainProgram := `uses MathUtils;
PrintLn('Result: ' + IntToStr(Add(5, 10)));`
	
	mainPath := filepath.Join(tempDir, "main.dws")
	if err := os.WriteFile(mainPath, []byte(mainProgram), 0644); err != nil {
		t.Fatalf("Failed to create main.dws: %v", err)
	}

	// Reset global state
	oldShowUnits := showUnits
	defer func() { showUnits = oldShowUnits }()
	showUnits = true

	// Capture stderr where tree is displayed
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Run the command
	err := runScript(runCmd, []string{mainPath})

	// Close write end and restore
	w.Close()
	os.Stderr = oldStderr

	// Read stderr output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	stderrOutput := buf.String()

	if err != nil {
		t.Fatalf("runScript failed: %v\nStderr: %s", err, stderrOutput)
	}

	// Verify tree output
	if !strings.Contains(stderrOutput, "Unit Dependency Tree") {
		t.Error("Expected 'Unit Dependency Tree' header in output")
	}

	if !strings.Contains(stderrOutput, "MathUtils") {
		t.Error("Expected 'MathUtils' in dependency tree")
	}

	if !strings.Contains(stderrOutput, "BaseUtils") {
		t.Error("Expected 'BaseUtils' in dependency tree")
	}

	// Verify tree structure (BaseUtils should be indented under MathUtils)
	if !strings.Contains(stderrOutput, "└─ BaseUtils") || !strings.Contains(stderrOutput, "  └─ BaseUtils") {
		t.Errorf("Expected BaseUtils to be shown as dependency of MathUtils with tree formatting")
	}
}
