package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// captureRun executes runScript with the run flags reset to their defaults, applies
// configure, and returns the combined stdout+stderr text plus the returned error.
// Inline source is passed through -e unless configure sets evalExpr = "" and provides args.
func captureRun(t *testing.T, source string, args []string, configure func()) (string, error) {
	t.Helper()
	saved := struct {
		evalExpr, hintsLevel, diagnosticsMode          string
		dumpAST, trace, typeCheck, showUnits, bytecode bool
		testEnvelope                                   bool
		maxRecursion                                   int
		searchPaths                                    []string
	}{evalExpr, hintsLevel, diagnosticsMode, dumpAST, trace, typeCheck, showUnits, bytecodeMode, testEnvelope, maxRecursion, unitSearchPaths}
	t.Cleanup(func() {
		evalExpr, hintsLevel, diagnosticsMode = saved.evalExpr, saved.hintsLevel, saved.diagnosticsMode
		dumpAST, trace, typeCheck, showUnits, bytecodeMode = saved.dumpAST, saved.trace, saved.typeCheck, saved.showUnits, saved.bytecode
		testEnvelope, maxRecursion, unitSearchPaths = saved.testEnvelope, saved.maxRecursion, saved.searchPaths
	})
	evalExpr, hintsLevel, diagnosticsMode = source, "off", "pretty"
	dumpAST, trace, typeCheck, showUnits, bytecodeMode, testEnvelope = false, false, true, false, false, false
	maxRecursion, unitSearchPaths = 1024, nil
	if configure != nil {
		configure()
	}

	oldStdout, oldStderr := os.Stdout, os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout, os.Stderr = w, w
	done := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		done <- buf.String()
	}()
	runErr := runScript(runCmd, args)
	_ = w.Close()
	os.Stdout, os.Stderr = oldStdout, oldStderr
	return <-done, runErr
}

// A recoverable parse error followed by a semantic error must report both, like the
// harness does (frontend keeps analyzing after non-blocking parse diagnostics).
func TestRun_ReportsParseAndSemanticDiagnosticsTogether(t *testing.T) {
	out, err := captureRun(t, "var x: Integer := y;\nwhile true PrintLn(1);", nil, nil)
	if err == nil {
		t.Fatal("expected a compile failure")
	}
	if !strings.Contains(out, "Unknown name \"y\"") || !strings.Contains(out, "DO expected") {
		t.Fatalf("expected both diagnostics, got:\n%s", out)
	}
}

// --hints off must actually disable hints.
func TestRun_HintsOffSuppressesHints(t *testing.T) {
	out, err := captureRun(t, "var Foo: Integer := 1; PrintLn(foo);", nil, func() { hintsLevel = "off" })
	if err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out)
	}
	if strings.Contains(out, "Hint:") {
		t.Fatalf("hints must be off, got:\n%s", out)
	}
	out, err = captureRun(t, "var Foo: Integer := 1; PrintLn(foo);", nil, func() { hintsLevel = "pedantic" })
	if err != nil || !strings.Contains(out, "Hint: \"foo\" does not match case of declaration (\"Foo\")") {
		t.Fatalf("expected the case-mismatch hint with --hints pedantic, err=%v, got:\n%s", err, out)
	}
}

// Unit-using programs still run; semantic analysis is skipped for them (explicit bypass).
func TestRun_UnitsBypassStillRuns(t *testing.T) {
	dir := t.TempDir()
	unit := "unit U;\ninterface\nfunction Twice(a: Integer): Integer;\nimplementation\nfunction Twice(a: Integer): Integer;\nbegin\n  Result := a * 2;\nend;\nend."
	if err := os.WriteFile(filepath.Join(dir, "U.dws"), []byte(unit), 0o644); err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(dir, "main.dws")
	if err := os.WriteFile(main, []byte("uses U;\nPrintLn(Twice(21));"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := captureRun(t, "", []string{main}, func() { evalExpr = "" })
	if err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out)
	}
	if !strings.Contains(out, "42") {
		t.Fatalf("expected program output, got:\n%s", out)
	}
}
