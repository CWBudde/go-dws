package interp

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/internal/encoding"
	"github.com/cwbudde/go-dws/internal/frontend"
	"github.com/cwbudde/go-dws/internal/semantic"
)

// Only inc_expire assumes all instructions before Sleep finish within 1ms.
// Workers run fixtures sequentially; this hook also must not be used by parallel
// tests because the builtins' global store is process-wide. Sleep is overridden
// only in this interpreter, and the previous store is restored after execution.
func configureFixtureClock(interp *Interpreter, path string) (func(), error) {
	if filepath.Base(filepath.Dir(path)) != "FunctionsGlobalVars" || filepath.Base(path) != "inc_expire.pas" {
		return func() {}, nil
	}
	clock := &fixtureScriptClock{now: time.Date(2026, 9, 19, 0, 0, 0, 0, time.UTC)}
	if err := interp.externalFunctions().Register("Sleep", clock); err != nil {
		return nil, err
	}
	store := builtins.NewGlobalVarStore()
	store.SetClock(func() time.Time { return clock.now })
	previous := builtins.DefaultGlobalVars
	builtins.DefaultGlobalVars = store
	return func() { builtins.DefaultGlobalVars = previous }, nil
}

type fixtureScriptClock struct{ now time.Time }

func (c *fixtureScriptClock) Call(args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Sleep() expects exactly 1 argument, got %d", len(args))
	}
	millis, ok := args[0].(*IntegerValue)
	if !ok {
		return nil, fmt.Errorf("Sleep() expects an Integer, got %s", args[0].Type())
	}
	if millis.Value > 0 {
		c.now = c.now.Add(time.Duration(millis.Value) * time.Millisecond)
	}
	return &NilValue{}, nil
}

func (*fixtureScriptClock) GetVarParams() []bool          { return []bool{false} }
func (*fixtureScriptClock) SetInterpreter(_ *Interpreter) {}
func (*fixtureScriptClock) GetParamTypes() []string       { return []string{"Integer"} }

// Delaying the first print reliably takes longer than inc_expire's 1ms TTL,
// reproducing a slow CI host without changing the upstream script.
type fixtureDelayedWriter struct {
	bytes.Buffer
	delayed bool
}

func (w *fixtureDelayedWriter) Write(p []byte) (int, error) {
	if !w.delayed {
		time.Sleep(5 * time.Millisecond)
		w.delayed = true
	}
	return w.Buffer.Write(p)
}

func TestFixtureClock_ExpirationIgnoresHostDelay(t *testing.T) {
	path := filepath.Join(fixturesRoot, "FunctionsGlobalVars", "inc_expire.pas")
	source, err := encoding.DecodeFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want, err := encoding.DecodeFile(filepath.Join(fixturesRoot, "FunctionsGlobalVars", "inc_expire.txt"))
	if err != nil {
		t.Fatal(err)
	}
	compiled := frontend.Compile(source, path, semantic.HintsLevelNormal)
	if compiled.HasFatalDiagnostics() || !compiled.SemanticSuccessful {
		t.Fatalf("compile diagnostics: %v", compiled.DiagnosticStrings())
	}
	var output fixtureDelayedWriter
	interp := New(&output)
	interp.SetSemanticInfo(compiled.SemanticInfo)
	restore, err := configureFixtureClock(interp, path)
	if err != nil {
		t.Fatal(err)
	}
	defer restore()
	result := interp.Eval(compiled.Program)
	if result != nil && result.Type() == "ERROR" {
		t.Fatalf("runtime error: %s", result.String())
	}
	assertOutput(t, normalizeOutput(output.String()), normalizeOutput(want))
}

func TestFixtureClock_Isolation(t *testing.T) {
	original := builtins.DefaultGlobalVars
	interp := New(&bytes.Buffer{})
	restore, err := configureFixtureClock(interp, filepath.Join(fixturesRoot, "FunctionsGlobalVars", "inc_expire.pas"))
	if err != nil {
		t.Fatal(err)
	}
	defer restore()
	if builtins.DefaultGlobalVars == original {
		t.Fatal("expiration fixture must use an isolated store")
	}
	if !interp.externalFunctions().Has("Sleep") {
		t.Fatal("expiration fixture must advance its clock through Sleep")
	}
	restore()
	if builtins.DefaultGlobalVars != original {
		t.Fatal("fixture clock did not restore the original store")
	}
	other := New(&bytes.Buffer{})
	restoreOther, err := configureFixtureClock(other, filepath.Join(fixturesRoot, "FunctionsGlobalVars", "basic.pas"))
	if err != nil {
		t.Fatal(err)
	}
	defer restoreOther()
	if builtins.DefaultGlobalVars != original || other.externalFunctions().Has("Sleep") {
		t.Fatal("unrelated fixture received clock overrides")
	}
}

func TestFixtureClock_ExpirationFixture(t *testing.T) {
	path := filepath.Join(fixturesRoot, "FunctionsGlobalVars", "inc_expire.pas")
	if result, detail := runFixtureTest(path, false, semantic.HintsLevelNormal); result != testResultPassed {
		t.Fatalf("inc_expire fixture failed: %s", detail)
	}
}

func TestFixtureClock_SleepAdvancesOnlyScriptTime(t *testing.T) {
	source := `
WriteGlobalVar('clock-test', 42, 0.010);
Sleep(5);
Sleep(0);
Sleep(-1);
PrintLn(ReadGlobalVarDef('clock-test', 0));
Sleep(5);
PrintLn(ReadGlobalVarDef('clock-test', 0));
`
	compiled := frontend.Compile(source, "clock-test.pas", semantic.HintsLevelNormal)
	if compiled.HasFatalDiagnostics() || !compiled.SemanticSuccessful {
		t.Fatalf("compile diagnostics: %v", compiled.DiagnosticStrings())
	}
	output, result := evalFixture(compiled, filepath.Join(fixturesRoot, "FunctionsGlobalVars", "inc_expire.pas"))
	if result != nil && result.Type() == "ERROR" {
		t.Fatalf("runtime error: %s", result.String())
	}
	assertOutput(t, output.String(), "42\n0\n")
}
