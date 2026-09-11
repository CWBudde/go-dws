package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/units"
)

// writeGreetUnit writes a unit named Greeter into dir whose Greet function
// returns marker, so that the caller can tell which copy was resolved.
func writeGreetUnit(t *testing.T, dir, marker string) {
	t.Helper()
	src := `unit Greeter;

interface

function Greet: String;

implementation

function Greet: String;
begin
  Result := '` + marker + `';
end;

end.`
	if err := os.WriteFile(filepath.Join(dir, "Greeter.dws"), []byte(src), 0o644); err != nil {
		t.Fatalf("failed to write Greeter.dws in %s: %v", dir, err)
	}
}

// runScriptCapturingStdout runs the given script file through the run command
// with the CLI globals reset, and returns everything it printed.
func runScriptCapturingStdout(t *testing.T, path string) string {
	t.Helper()

	oldSearchPaths, oldVerbose, oldEval := unitSearchPaths, verbose, evalExpr
	t.Cleanup(func() { unitSearchPaths, verbose, evalExpr = oldSearchPaths, oldVerbose, oldEval })
	unitSearchPaths, verbose, evalExpr = nil, false, ""

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	runErr := runScript(runCmd, []string{path})

	w.Close()
	os.Stdout = oldStdout

	var out bytes.Buffer
	if _, err := out.ReadFrom(r); err != nil {
		t.Fatalf("failed to read captured output: %v", err)
	}
	if runErr != nil {
		t.Fatalf("runScript(%s) failed: %v\noutput: %s", path, runErr, out.String())
	}
	return out.String()
}

// TestResolveUnitSearchPaths_EnvPath proves that a unit reachable only through
// DWSCRIPT_PATH resolves when no -I path is given, and that a same-named unit
// in the script's own directory still wins over it.
func TestResolveUnitSearchPaths_EnvPath(t *testing.T) {
	scriptDir := t.TempDir()
	libDir := t.TempDir()

	writeGreetUnit(t, libDir, "from DWSCRIPT_PATH")

	mainPath := filepath.Join(scriptDir, "main.dws")
	main := "uses Greeter;\n\nPrintLn(Greet);"
	if err := os.WriteFile(mainPath, []byte(main), 0o644); err != nil {
		t.Fatalf("failed to write main.dws: %v", err)
	}

	t.Setenv(units.SearchPathEnvVar, libDir)

	// Only the DWSCRIPT_PATH copy exists: it must be found.
	if out := runScriptCapturingStdout(t, mainPath); !strings.Contains(out, "from DWSCRIPT_PATH") {
		t.Errorf("expected the unit from DWSCRIPT_PATH to resolve, got output: %q", out)
	}

	// With a same-named unit next to the script, the script's directory wins.
	writeGreetUnit(t, scriptDir, "from script dir")
	out := runScriptCapturingStdout(t, mainPath)
	if !strings.Contains(out, "from script dir") {
		t.Errorf("expected the script directory to take priority, got output: %q", out)
	}
	if strings.Contains(out, "from DWSCRIPT_PATH") {
		t.Errorf("script directory did not shadow DWSCRIPT_PATH, got output: %q", out)
	}
}

// TestResolveUnitSearchPaths_Order pins the ordering contract: the script
// directory is only injected when no -I path is given, explicit -I paths keep
// their order and priority, and the defaults are appended last.
func TestResolveUnitSearchPaths_Order(t *testing.T) {
	libDir := t.TempDir()
	explicit := t.TempDir()
	t.Setenv(units.SearchPathEnvVar, libDir)

	oldSearchPaths := unitSearchPaths
	t.Cleanup(func() { unitSearchPaths = oldSearchPaths })

	// No -I: script directory first, then the defaults (which include ".", the
	// DWSCRIPT_PATH entry, and possibly user/system directories).
	unitSearchPaths = nil
	got := resolveUnitSearchPaths(filepath.Join("scripts", "main.dws"))
	if len(got) < 3 || got[0] != "scripts" || got[1] != "." || got[2] != libDir {
		t.Errorf("expected [scripts . %s ...], got %v", libDir, got)
	}

	// With -I: the explicit paths come first, the defaults are appended after
	// them, and the script directory is not injected.
	unitSearchPaths = []string{explicit}
	got = resolveUnitSearchPaths(filepath.Join("scripts", "main.dws"))
	if len(got) < 3 || got[0] != explicit || got[1] != "." || got[2] != libDir {
		t.Errorf("expected [%s . %s ...], got %v", explicit, libDir, got)
	}
	for _, p := range got {
		if p == "scripts" {
			t.Errorf("script directory must not be injected when -I is given, got %v", got)
		}
	}

	// Inline -e code has no script directory.
	unitSearchPaths = nil
	got = resolveUnitSearchPaths(evalFilename)
	if len(got) == 0 || got[0] != "." {
		t.Errorf("expected the defaults for inline code, got %v", got)
	}
}

// TestResolveUnitSearchPaths_Dedupe checks that a directory named twice, once
// explicitly and once through the defaults, appears only once.
func TestResolveUnitSearchPaths_Dedupe(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(units.SearchPathEnvVar, dir)

	oldSearchPaths := unitSearchPaths
	t.Cleanup(func() { unitSearchPaths = oldSearchPaths })
	unitSearchPaths = []string{dir}

	got := resolveUnitSearchPaths(filepath.Join("scripts", "main.dws"))
	count := 0
	for _, p := range got {
		if p == dir {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected %s exactly once in %v, got %d occurrences", dir, got, count)
	}
}
