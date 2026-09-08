package dwscript

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEngine_UnitCompilationAndExecution(t *testing.T) {
	dir := t.TempDir()
	sources := map[string]string{
		"Numbers.dws": `unit Numbers; interface function Twice(n: Integer): Integer; implementation function Twice(n: Integer): Integer; begin Result := n * 2; end; end.`,
		"Facade.dws":  `unit Facade; interface function Answer: Integer; implementation uses Numbers; function Answer: Integer; begin Result := Twice(21); end; initialization PrintLn('init'); finalization PrintLn('final'); end.`,
	}
	for name, source := range sources {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var output bytes.Buffer
	engine, err := New(WithUnitSearchPaths(dir), WithOutput(&output))
	if err != nil {
		t.Fatal(err)
	}
	program, err := engine.Compile(`uses Facade; PrintLn(Facade.Answer());`)
	if err != nil {
		t.Fatal(err)
	}
	// Execution uses the analyzed unit trees and can be repeated without reparsing.
	for name := range sources {
		if err := os.Remove(filepath.Join(dir, name)); err != nil {
			t.Fatal(err)
		}
	}
	for range 2 {
		output.Reset()
		result, err := engine.Run(program)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Success || output.String() != "init\n42\nfinal\n" {
			t.Fatalf("result=%+v output=%q", result, output.String())
		}
	}
}

func TestEngine_UnitCompileError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "U.dws"), []byte("unit U; interface implementation end."), 0600); err != nil {
		t.Fatal(err)
	}
	engine, err := New(WithUnitSearchPaths(dir))
	if err != nil {
		t.Fatal(err)
	}
	_, err = engine.Compile(`uses U; var n: Integer := 'bad';`)
	if err == nil || !strings.Contains(err.Error(), "Cannot assign String to Integer") {
		t.Fatalf("expected type error, got %v", err)
	}
}
