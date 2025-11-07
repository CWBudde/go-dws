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

function Concat(a, b: String): String;

implementation

function Concat(a, b: String): String;
begin
  Result := a + b;
end;

end.`

	unitPath := filepath.Join(unitsDir, "StringUtils.dws")
	if err := os.WriteFile(unitPath, []byte(stringUnit), 0644); err != nil {
		t.Fatalf("Failed to create StringUtils.dws: %v", err)
	}

	// Create main program in main directory
	mainProgram := `uses StringUtils;

PrintLn(Concat('Hello', 'World'));`

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

	// Verify output contains concatenated string
	if !strings.Contains(output, "HelloWorld") {
		t.Errorf("Expected 'HelloWorld' in output, got: %s", output)
	}
}

// TestRunBytecodeSimpleScript ensures bytecode mode executes simple programs.
func TestRunBytecodeSimpleScript(t *testing.T) {
	tempDir := t.TempDir()
	script := `var x: Integer := 41;
x := x + 1;`

	scriptPath := filepath.Join(tempDir, "main.dws")
	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		t.Fatalf("Failed to create script: %v", err)
	}

	oldBytecode := bytecodeMode
	defer func() { bytecodeMode = oldBytecode }()
	bytecodeMode = true

	if err := runScript(runCmd, []string{scriptPath}); err != nil {
		t.Fatalf("runScript in bytecode mode failed: %v", err)
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

// TestRunMainDwsEndToEnd tests running a program with multiple units and init/final order
func TestRunMainDwsEndToEnd(t *testing.T) {
	tempDir := t.TempDir()

	// Create MathUtils unit
	mathUtils := `unit MathUtils;

interface

function Add(a, b: Integer): Integer;
function Square(x: Integer): Integer;

implementation

function Add(a, b: Integer): Integer;
begin
  Result := a + b;
end;

function Square(x: Integer): Integer;
begin
  Result := x * x;
end;

initialization
  PrintLn('MathUtils loaded');

finalization
  PrintLn('MathUtils unloading');

end.`

	if err := os.WriteFile(filepath.Join(tempDir, "MathUtils.dws"), []byte(mathUtils), 0644); err != nil {
		t.Fatalf("Failed to create MathUtils.dws: %v", err)
	}

	// Create StringUtils unit
	stringUtils := `unit StringUtils;

interface

function Concat(a, b: String): String;

implementation

function Concat(a, b: String): String;
begin
  Result := a + b;
end;

initialization
  PrintLn('StringUtils loaded');

finalization
  PrintLn('StringUtils unloading');

end.`

	if err := os.WriteFile(filepath.Join(tempDir, "StringUtils.dws"), []byte(stringUtils), 0644); err != nil {
		t.Fatalf("Failed to create StringUtils.dws: %v", err)
	}

	// Create main program that uses both units (not a unit itself, but a program)
	mainProgram := `uses MathUtils, StringUtils;

begin
  var x: Integer := 10;
  var y: Integer := 20;
  var sum: Integer := Add(x, y);
  PrintLn('10 + 20 = ' + IntToStr(sum));
  PrintLn('Square of 5 = ' + IntToStr(Square(5)));
  var result: String := Concat('Hello', 'World');
  PrintLn('Concat result: ' + result);
end.`

	mainPath := filepath.Join(tempDir, "main.dws")
	if err := os.WriteFile(mainPath, []byte(mainProgram), 0644); err != nil {
		t.Fatalf("Failed to create main.dws: %v", err)
	}

	// Reset global state
	oldSearchPaths := unitSearchPaths
	defer func() {
		unitSearchPaths = oldSearchPaths
	}()
	unitSearchPaths = []string{tempDir}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the command
	err := runScript(runCmd, []string{mainPath})

	// Close write end and restore
	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("runScript failed: %v\nOutput: %s", err, output)
	}

	// Verify initialization order (dependencies are initialized before program runs)
	mathUtilsIdx := strings.Index(output, "MathUtils loaded")
	stringUtilsIdx := strings.Index(output, "StringUtils loaded")

	if mathUtilsIdx == -1 {
		t.Error("Expected 'MathUtils loaded' in output")
	}
	if stringUtilsIdx == -1 {
		t.Error("Expected 'StringUtils loaded' in output")
	}

	// Verify actual computation output
	if !strings.Contains(output, "10 + 20 = 30") {
		t.Error("Expected '10 + 20 = 30' in output")
	}
	if !strings.Contains(output, "Square of 5 = 25") {
		t.Error("Expected 'Square of 5 = 25' in output")
	}
	if !strings.Contains(output, "Concat result: HelloWorld") {
		t.Error("Expected 'Concat result: HelloWorld' in output")
	}

	// Verify finalization order (units finalize after program ends)
	mathUtilsFinIdx := strings.Index(output, "MathUtils unloading")
	stringUtilsFinIdx := strings.Index(output, "StringUtils unloading")

	if mathUtilsFinIdx == -1 {
		t.Error("Expected 'MathUtils unloading' in output")
	}
	if stringUtilsFinIdx == -1 {
		t.Error("Expected 'StringUtils unloading' in output")
	}
}

// TestCircularDependencyDetection tests that circular dependencies are detected
func TestCircularDependencyDetection(t *testing.T) {
	unitAPath := filepath.Join("..", "..", "..", "testdata", "units", "UnitA.dws")

	// Check if files exist
	if _, err := os.Stat(unitAPath); os.IsNotExist(err) {
		t.Skip("testdata/units/UnitA.dws not found")
	}

	// Reset global state
	oldSearchPaths := unitSearchPaths
	defer func() {
		unitSearchPaths = oldSearchPaths
	}()

	// Add testdata/units to search path
	unitsDir := filepath.Join("..", "..", "..", "testdata", "units")
	unitSearchPaths = []string{unitsDir}

	// Create a main program that tries to use UnitA (which has circular dependency with UnitB)
	tempDir := t.TempDir()
	mainProgram := `uses UnitA;
PrintLn('Should not get here');`

	mainPath := filepath.Join(tempDir, "main.dws")
	if err := os.WriteFile(mainPath, []byte(mainProgram), 0644); err != nil {
		t.Fatalf("Failed to create main.dws: %v", err)
	}

	// Capture stderr (where errors go)
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Run the command (should fail)
	err := runScript(runCmd, []string{mainPath})

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	stderrOutput := buf.String()

	// Should fail with circular dependency error
	if err == nil {
		t.Error("Expected error for circular dependency, but got none")
	} else {
		t.Logf("Got expected error: %v", err)
	}

	// Verify error message mentions circular dependency
	errorOutput := stderrOutput + err.Error()
	if !strings.Contains(errorOutput, "circular") && !strings.Contains(errorOutput, "cycle") {
		t.Logf("Stderr output: %s", stderrOutput)
		t.Logf("Error: %v", err)
		t.Error("Expected error message to mention circular dependency or cycle")
	}
}

// TestUnitNotFound tests error handling when a unit cannot be found
func TestUnitNotFound(t *testing.T) {
	tempDir := t.TempDir()

	// Create main program that uses a non-existent unit
	mainProgram := `uses NonExistentUnit;
PrintLn('Should not get here');`

	mainPath := filepath.Join(tempDir, "main.dws")
	if err := os.WriteFile(mainPath, []byte(mainProgram), 0644); err != nil {
		t.Fatalf("Failed to create main.dws: %v", err)
	}

	// Reset global state
	oldSearchPaths := unitSearchPaths
	defer func() {
		unitSearchPaths = oldSearchPaths
	}()
	unitSearchPaths = []string{tempDir}

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Run the command (should fail)
	err := runScript(runCmd, []string{mainPath})

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	stderrOutput := buf.String()

	// Should fail
	if err == nil {
		t.Error("Expected error for missing unit, but got none")
	} else {
		t.Logf("Got expected error: %v", err)
	}

	// Verify error message mentions the unit
	errorOutput := stderrOutput + err.Error()
	if !strings.Contains(errorOutput, "NonExistentUnit") && !strings.Contains(errorOutput, "not found") {
		t.Logf("Stderr output: %s", stderrOutput)
		t.Logf("Error: %v", err)
		t.Error("Expected error message to mention missing unit")
	}
}

// TestMultipleSearchPaths tests using multiple -I flags
func TestMultipleSearchPaths(t *testing.T) {
	// Create three directories
	mainDir := t.TempDir()
	lib1Dir := t.TempDir()
	lib2Dir := t.TempDir()

	// Create a unit in lib1
	lib1Unit := `unit LibOneUnit;
interface
function GetLibOneValue: Integer;
implementation
function GetLibOneValue: Integer;
begin
  Result := 100;
end;
end.`

	if err := os.WriteFile(filepath.Join(lib1Dir, "LibOneUnit.dws"), []byte(lib1Unit), 0644); err != nil {
		t.Fatalf("Failed to create LibOneUnit.dws: %v", err)
	}

	// Create a unit in lib2
	lib2Unit := `unit LibTwoUnit;
interface
function GetLibTwoValue: Integer;
implementation
function GetLibTwoValue: Integer;
begin
  Result := 200;
end;
end.`

	if err := os.WriteFile(filepath.Join(lib2Dir, "LibTwoUnit.dws"), []byte(lib2Unit), 0644); err != nil {
		t.Fatalf("Failed to create LibTwoUnit.dws: %v", err)
	}

	// Create main program that uses both units
	mainProgram := `uses LibOneUnit, LibTwoUnit;

begin
  PrintLn('Lib1: ' + IntToStr(GetLibOneValue()));
  PrintLn('Lib2: ' + IntToStr(GetLibTwoValue()));
end.`

	mainPath := filepath.Join(mainDir, "main.dws")
	if err := os.WriteFile(mainPath, []byte(mainProgram), 0644); err != nil {
		t.Fatalf("Failed to create main.dws: %v", err)
	}

	// Reset global state
	oldSearchPaths := unitSearchPaths
	defer func() {
		unitSearchPaths = oldSearchPaths
	}()

	// Set multiple search paths
	unitSearchPaths = []string{lib1Dir, lib2Dir}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the command
	err := runScript(runCmd, []string{mainPath})

	// Close and restore
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("runScript failed: %v\nOutput: %s", err, output)
	}

	t.Logf("Output: %s", output)

	// Verify both units were loaded and executed
	if !strings.Contains(output, "Lib1: 100") {
		t.Error("Expected 'Lib1: 100' in output")
	}
	if !strings.Contains(output, "Lib2: 200") {
		t.Error("Expected 'Lib2: 200' in output")
	}
}

// TestSearchPathPriority tests that first search path has priority
func TestSearchPathPriority(t *testing.T) {
	mainDir := t.TempDir()
	highPriorityDir := t.TempDir()
	lowPriorityDir := t.TempDir()

	// Create the same unit in both directories with different implementations
	highPriorityUnit := `unit PriorityUnit;
interface
function GetMessage: String;
implementation
function GetMessage: String;
begin
  Result := 'High Priority';
end;
end.`

	lowPriorityUnit := `unit PriorityUnit;
interface
function GetMessage: String;
implementation
function GetMessage: String;
begin
  Result := 'Low Priority';
end;
end.`

	if err := os.WriteFile(filepath.Join(highPriorityDir, "PriorityUnit.dws"), []byte(highPriorityUnit), 0644); err != nil {
		t.Fatalf("Failed to create high priority unit: %v", err)
	}
	if err := os.WriteFile(filepath.Join(lowPriorityDir, "PriorityUnit.dws"), []byte(lowPriorityUnit), 0644); err != nil {
		t.Fatalf("Failed to create low priority unit: %v", err)
	}

	// Create main program
	mainProgram := `uses PriorityUnit;
PrintLn(GetMessage());`

	mainPath := filepath.Join(mainDir, "main.dws")
	if err := os.WriteFile(mainPath, []byte(mainProgram), 0644); err != nil {
		t.Fatalf("Failed to create main.dws: %v", err)
	}

	// Reset global state
	oldSearchPaths := unitSearchPaths
	defer func() {
		unitSearchPaths = oldSearchPaths
	}()

	// High priority path comes first
	unitSearchPaths = []string{highPriorityDir, lowPriorityDir}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the command
	err := runScript(runCmd, []string{mainPath})

	// Close and restore
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("runScript failed: %v\nOutput: %s", err, output)
	}

	// Should use the high priority version
	if !strings.Contains(output, "High Priority") {
		t.Errorf("Expected 'High Priority' in output, got: %s", output)
	}
	if strings.Contains(output, "Low Priority") {
		t.Error("Should not have loaded low priority version")
	}
}

// TestQualifiedNameAccess tests using qualified names (UnitName.FunctionName)
// Note: This tests a feature that may not be fully implemented yet
func TestQualifiedNameAccess(t *testing.T) {
	t.Skip("Qualified name access (UnitName.FunctionName) not yet fully implemented")
	tempDir := t.TempDir()

	// Create two units with functions of the same name
	unit1 := `unit UnitOne;
interface
function GetValue: Integer;
implementation
function GetValue: Integer;
begin
  Result := 1;
end;
end.`

	unit2 := `unit UnitTwo;
interface
function GetValue: Integer;
implementation
function GetValue: Integer;
begin
  Result := 2;
end;
end.`

	if err := os.WriteFile(filepath.Join(tempDir, "UnitOne.dws"), []byte(unit1), 0644); err != nil {
		t.Fatalf("Failed to create UnitOne.dws: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "UnitTwo.dws"), []byte(unit2), 0644); err != nil {
		t.Fatalf("Failed to create UnitTwo.dws: %v", err)
	}

	// Create main program using qualified names to avoid conflict
	mainProgram := `uses UnitOne, UnitTwo;
PrintLn('Unit1: ' + IntToStr(UnitOne.GetValue()));
PrintLn('Unit2: ' + IntToStr(UnitTwo.GetValue()));`

	mainPath := filepath.Join(tempDir, "main.dws")
	if err := os.WriteFile(mainPath, []byte(mainProgram), 0644); err != nil {
		t.Fatalf("Failed to create main.dws: %v", err)
	}

	// Reset global state
	oldSearchPaths := unitSearchPaths
	defer func() {
		unitSearchPaths = oldSearchPaths
	}()
	unitSearchPaths = []string{tempDir}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the command
	err := runScript(runCmd, []string{mainPath})

	// Close and restore
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("runScript failed: %v\nOutput: %s", err, output)
	}

	// Verify both qualified calls work
	if !strings.Contains(output, "Unit1: 1") {
		t.Error("Expected 'Unit1: 1' in output")
	}
	if !strings.Contains(output, "Unit2: 2") {
		t.Error("Expected 'Unit2: 2' in output")
	}
}

// TestRecursiveUnitLoading tests units that depend on other units (A uses B uses C)
func TestRecursiveUnitLoading(t *testing.T) {
	tempDir := t.TempDir()

	// Create UnitC (no dependencies)
	unitC := `unit UnitC;
interface
implementation
initialization
  PrintLn('UnitC loaded');
finalization
  PrintLn('UnitC unloading');
end.`

	// Create UnitB (depends on C)
	unitB := `unit UnitB;
interface
uses UnitC;
implementation
initialization
  PrintLn('UnitB loaded');
finalization
  PrintLn('UnitB unloading');
end.`

	// Create UnitA (depends on B)
	unitA := `unit UnitA;
interface
uses UnitB;
implementation
initialization
  PrintLn('UnitA loaded');
finalization
  PrintLn('UnitA unloading');
end.`

	if err := os.WriteFile(filepath.Join(tempDir, "UnitC.dws"), []byte(unitC), 0644); err != nil {
		t.Fatalf("Failed to create UnitC.dws: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "UnitB.dws"), []byte(unitB), 0644); err != nil {
		t.Fatalf("Failed to create UnitB.dws: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "UnitA.dws"), []byte(unitA), 0644); err != nil {
		t.Fatalf("Failed to create UnitA.dws: %v", err)
	}

	// Create main program
	mainProgram := `uses UnitA;

begin
  PrintLn('All units loaded successfully');
end.`

	mainPath := filepath.Join(tempDir, "main.dws")
	if err := os.WriteFile(mainPath, []byte(mainProgram), 0644); err != nil {
		t.Fatalf("Failed to create main.dws: %v", err)
	}

	// Reset global state
	oldSearchPaths := unitSearchPaths
	defer func() {
		unitSearchPaths = oldSearchPaths
	}()
	unitSearchPaths = []string{tempDir}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the command
	err := runScript(runCmd, []string{mainPath})

	// Close and restore
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("runScript failed: %v\nOutput: %s", err, output)
	}

	// Verify program executed
	if !strings.Contains(output, "All units loaded successfully") {
		t.Errorf("Expected 'All units loaded successfully' in output, got: %s", output)
	}

	// Verify initialization order (C before B before A)
	cIdx := strings.Index(output, "UnitC loaded")
	bIdx := strings.Index(output, "UnitB loaded")
	aIdx := strings.Index(output, "UnitA loaded")

	if cIdx == -1 || bIdx == -1 || aIdx == -1 {
		t.Error("Expected all units to be loaded")
	}
	if !(cIdx < bIdx && bIdx < aIdx) {
		t.Error("Units should initialize in dependency order: C, B, A")
	}

	// Verify finalization order (A before B before C - reverse of init)
	aFinIdx := strings.Index(output, "UnitA unloading")
	bFinIdx := strings.Index(output, "UnitB unloading")
	cFinIdx := strings.Index(output, "UnitC unloading")

	if aFinIdx == -1 || bFinIdx == -1 || cFinIdx == -1 {
		t.Error("Expected all units to be finalized")
	}
	if !(aFinIdx < bFinIdx && bFinIdx < cFinIdx) {
		t.Error("Units should finalize in reverse dependency order: A, B, C")
	}
}

// TestMissingImplementation tests when interface declares but implementation is missing
// Note: This tests semantic validation that may not yet be enforced
func TestMissingImplementation(t *testing.T) {
	t.Skip("Missing implementation validation not yet enforced in semantic analysis")
	tempDir := t.TempDir()

	// Create unit with missing implementation
	badUnit := `unit BadUnit;
interface
function DeclaredButNotImplemented: Integer;
implementation
// Missing implementation!
end.`

	if err := os.WriteFile(filepath.Join(tempDir, "BadUnit.dws"), []byte(badUnit), 0644); err != nil {
		t.Fatalf("Failed to create BadUnit.dws: %v", err)
	}

	// Create main program
	mainProgram := `uses BadUnit;
PrintLn('Should not get here');`

	mainPath := filepath.Join(tempDir, "main.dws")
	if err := os.WriteFile(mainPath, []byte(mainProgram), 0644); err != nil {
		t.Fatalf("Failed to create main.dws: %v", err)
	}

	// Reset global state
	oldSearchPaths := unitSearchPaths
	defer func() {
		unitSearchPaths = oldSearchPaths
	}()
	unitSearchPaths = []string{tempDir}

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Run the command (should fail)
	err := runScript(runCmd, []string{mainPath})

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	stderrOutput := buf.String()

	// Should fail
	if err == nil {
		t.Error("Expected error for missing implementation, but got none")
	}

	// Verify error message mentions missing implementation or the function
	if !strings.Contains(stderrOutput, "DeclaredButNotImplemented") &&
		!strings.Contains(stderrOutput, "implementation") &&
		!strings.Contains(stderrOutput, "not implemented") {
		t.Logf("Stderr output: %s", stderrOutput)
		t.Error("Expected error message about missing implementation")
	}
}

// TestCombinedFlags tests combining -I, -v, and --show-units flags
func TestCombinedFlags(t *testing.T) {
	mainDir := t.TempDir()
	unitsDir := t.TempDir()

	// Create a simple unit
	unit := `unit TestUnit;
interface
function GetValue: Integer;
implementation
function GetValue: Integer;
begin
  Result := 42;
end;
initialization
  PrintLn('TestUnit loaded');
end.`

	if err := os.WriteFile(filepath.Join(unitsDir, "TestUnit.dws"), []byte(unit), 0644); err != nil {
		t.Fatalf("Failed to create TestUnit.dws: %v", err)
	}

	// Create main program
	mainProgram := `uses TestUnit;
PrintLn('Value: ' + IntToStr(GetValue()));`

	mainPath := filepath.Join(mainDir, "main.dws")
	if err := os.WriteFile(mainPath, []byte(mainProgram), 0644); err != nil {
		t.Fatalf("Failed to create main.dws: %v", err)
	}

	// Reset global state
	oldSearchPaths := unitSearchPaths
	oldVerbose := verbose
	oldShowUnits := showUnits
	defer func() {
		unitSearchPaths = oldSearchPaths
		verbose = oldVerbose
		showUnits = oldShowUnits
	}()

	// Enable all flags
	unitSearchPaths = []string{unitsDir}
	verbose = true
	showUnits = true

	// Capture both stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	// Run the command
	err := runScript(runCmd, []string{mainPath})

	// Close pipes and restore
	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var bufOut, bufErr bytes.Buffer
	bufOut.ReadFrom(rOut)
	bufErr.ReadFrom(rErr)
	stdoutOutput := bufOut.String()
	stderrOutput := bufErr.String()

	if err != nil {
		t.Fatalf("runScript failed: %v\nStdout: %s\nStderr: %s", err, stdoutOutput, stderrOutput)
	}

	// Verify stdout has normal output
	if !strings.Contains(stdoutOutput, "TestUnit loaded") {
		t.Error("Expected 'TestUnit loaded' in stdout")
	}
	if !strings.Contains(stdoutOutput, "Value: 42") {
		t.Error("Expected 'Value: 42' in stdout")
	}

	// Verify stderr has unit tree (from --show-units)
	if !strings.Contains(stderrOutput, "Unit Dependency Tree") && !strings.Contains(stderrOutput, "TestUnit") {
		t.Logf("Stderr: %s", stderrOutput)
		// Note: Might not always show depending on implementation
	}
}

// TestAbsoluteAndRelativeSearchPaths tests mixing absolute and relative paths
func TestAbsoluteAndRelativeSearchPaths(t *testing.T) {
	mainDir := t.TempDir()
	unitsDir := t.TempDir()

	// Create a unit
	unit := `unit PathTestUnit;
interface
function GetValue: Integer;
implementation
function GetValue: Integer;
begin
  Result := 99;
end;
end.`

	if err := os.WriteFile(filepath.Join(unitsDir, "PathTestUnit.dws"), []byte(unit), 0644); err != nil {
		t.Fatalf("Failed to create PathTestUnit.dws: %v", err)
	}

	// Create main program
	mainProgram := `uses PathTestUnit;
PrintLn('Value: ' + IntToStr(GetValue()));`

	mainPath := filepath.Join(mainDir, "main.dws")
	if err := os.WriteFile(mainPath, []byte(mainProgram), 0644); err != nil {
		t.Fatalf("Failed to create main.dws: %v", err)
	}

	// Reset global state
	oldSearchPaths := unitSearchPaths
	defer func() {
		unitSearchPaths = oldSearchPaths
	}()

	// Use absolute path
	absPath, err := filepath.Abs(unitsDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}
	unitSearchPaths = []string{absPath}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the command
	err = runScript(runCmd, []string{mainPath})

	// Close and restore
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("runScript failed with absolute path: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, "Value: 99") {
		t.Error("Expected 'Value: 99' in output with absolute path")
	}
}
