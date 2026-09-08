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
		evalExpr        string
		hintsLevel      string
		diagnosticsMode string
		searchPaths     []string
		maxRecursion    int
		dumpAST         bool
		trace           bool
		typeCheck       bool
		showUnits       bool
		bytecode        bool
		testEnvelope    bool
		compileOnly     bool
		verbose         bool
		silenceUsage    bool
	}{
		evalExpr: evalExpr, hintsLevel: hintsLevel, diagnosticsMode: diagnosticsMode,
		searchPaths: unitSearchPaths, maxRecursion: maxRecursion,
		dumpAST: dumpAST, trace: trace, typeCheck: typeCheck, showUnits: showUnits,
		bytecode: bytecodeMode, testEnvelope: testEnvelope, compileOnly: compileOnly,
		verbose: verbose, silenceUsage: runCmd.SilenceUsage,
	}
	t.Cleanup(func() {
		evalExpr, hintsLevel, diagnosticsMode = saved.evalExpr, saved.hintsLevel, saved.diagnosticsMode
		dumpAST, trace, typeCheck, showUnits, bytecodeMode = saved.dumpAST, saved.trace, saved.typeCheck, saved.showUnits, saved.bytecode
		testEnvelope, compileOnly, maxRecursion, unitSearchPaths = saved.testEnvelope, saved.compileOnly, saved.maxRecursion, saved.searchPaths
		verbose, runCmd.SilenceUsage = saved.verbose, saved.silenceUsage
	})
	evalExpr, hintsLevel, diagnosticsMode = source, "off", "pretty"
	dumpAST, trace, typeCheck, showUnits, bytecodeMode, testEnvelope, compileOnly, verbose = false, false, true, false, false, false, false, false
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
	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		done <- buf.String()
	}()
	// Restore the process streams even if runScript panics, so a failure here
	// cannot wedge the rest of the package's tests behind a dangling pipe.
	defer func() {
		os.Stdout, os.Stderr = oldStdout, oldStderr
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

// Unit programs execute after the shared frontend resolves and checks their dependencies.
func TestRun_UnitsTypeChecked(t *testing.T) {
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

// Unit finalization output must be part of the (buffered) envelope output.
func TestRun_TestEnvelopeIncludesUnitFinalization(t *testing.T) {
	dir := t.TempDir()
	unit := "unit U;\ninterface\nprocedure Hello;\nimplementation\nprocedure Hello;\nbegin\n  PrintLn('hello');\nend;\ninitialization\n  PrintLn('init U');\nfinalization\n  PrintLn('final U');\nend."
	if err := os.WriteFile(filepath.Join(dir, "U.dws"), []byte(unit), 0o644); err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(dir, "main.dws")
	if err := os.WriteFile(main, []byte("uses U;\nHello;"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, envelope := range []bool{false, true} {
		out, err := captureRun(t, "", []string{main}, func() { evalExpr = ""; testEnvelope = envelope })
		if err != nil {
			t.Fatalf("envelope=%v: %v\n%s", envelope, err, out)
		}
		if out != "init U\nhello\nfinal U\n" {
			t.Fatalf("envelope=%v: got %q", envelope, out)
		}
	}
}

// Pretty mode keeps the analyzer's detail lines for type mismatches.
func TestRun_PrettyKeepsExpectedGotDetail(t *testing.T) {
	out, err := captureRun(t, "var x: Integer := 'hello';", nil, nil)
	if err == nil || !strings.Contains(out, "Expected: Integer") || !strings.Contains(out, "Got: String") {
		t.Fatalf("err=%v out=%q", err, out)
	}
}

// Harness-only modes are refused where they cannot work.
func TestRun_HarnessModesRejectBytecode(t *testing.T) {
	_, err := captureRun(t, "PrintLn(1);", nil, func() { bytecodeMode = true; testEnvelope = true })
	if err == nil || !strings.Contains(err.Error(), "--bytecode") {
		t.Fatalf("expected a rejection, got %v", err)
	}
}

// A missing argument is a usage error and must not be silenced.
func TestRun_NoArgsKeepsUsage(t *testing.T) {
	_, err := captureRun(t, "", nil, func() { evalExpr = "" })
	if err == nil || runCmd.SilenceUsage {
		t.Fatalf("expected a usage error with usage enabled, err=%v silenced=%v", err, runCmd.SilenceUsage)
	}
}

func TestRun_UnitProgramRejectsTypeErrors(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "U.dws"), []byte("unit U; interface implementation end."), 0600); err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(dir, "main.dws")
	if err := os.WriteFile(main, []byte("uses U; var n: Integer := 'bad';"), 0600); err != nil {
		t.Fatal(err)
	}
	out, err := captureRun(t, "", []string{main}, func() { evalExpr = ""; typeCheck = true })
	if err == nil || !strings.Contains(out, "Cannot assign String to Integer") {
		t.Fatalf("err=%v output=%s", err, out)
	}
}
