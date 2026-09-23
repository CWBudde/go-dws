package interp

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/cwbudde/go-dws/internal/fixtureconfig"
	"github.com/cwbudde/go-dws/internal/frontend"
	"github.com/cwbudde/go-dws/internal/semantic"
)

// Keep the recursive failure in a child process so a regression cannot hang
// or exhaust the test runner's stack.
func TestE15_ClassNameHelper_Bounded(t *testing.T) {
	if os.Getenv("DWS_E15_CLASSNAME_CHILD") == "1" {
		path := filepath.Join(fixturesRoot, "HelpersPass", "classname_helper1.pas")
		if got, detail := runFixtureTest(path, false, fixtureconfig.HintsLevel("HelpersPass")); got != testResultPassed {
			t.Fatalf("classname_helper1: %v: %s", got, detail)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestE15_ClassNameHelper_Bounded$")
	cmd.Env = append(os.Environ(), "DWS_E15_CLASSNAME_CHILD=1")
	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("classname_helper1 exceeded 3 seconds (recursive helper dispatch): %s", output)
	}
	if err != nil {
		t.Fatalf("classname_helper1 failed: %v\n%s", err, output)
	}
}

func TestE15_DynamicArrayCreateFixture(t *testing.T) {
	path := filepath.Join(fixturesRoot, "HelpersPass", "dyn_array_create.pas")
	if got, detail := runFixtureTest(path, false, fixtureconfig.HintsLevel("HelpersPass")); got != testResultPassed {
		t.Fatalf("dyn_array_create: %v: %s", got, detail)
	}
}

func TestE15_ClassNameInheritedFixture(t *testing.T) {
	path := filepath.Join(fixturesRoot, "HelpersPass", "classname_helper2.pas")
	if got, detail := runFixtureTest(path, false, fixtureconfig.HintsLevel("HelpersPass")); got != testResultPassed {
		t.Fatalf("classname_helper2: %v: %s", got, detail)
	}
}

func TestE15_CreateKeepsClassAndRecordDispatch(t *testing.T) {
	compileAndRunWithHelperTransfer(t, `
type TBox = class
  Value: Integer;
  constructor Create(n: Integer);
  begin Value := n; end;
end;
type TPair = record
  Value: Integer;
  class function Create(n: Integer): TPair; static;
  begin Result.Value := n; end;
end;
PrintLn(TBox.Create(42).Value);
PrintLn(TPair.Create(7).Value);
`, "e15_create_dispatch.dws", "42\n7\n")
}

func TestE15_ArrayTypeCannotUseInstanceCreateOrNew(t *testing.T) {
	for _, call := range []string{"TStrings.Create(['x'])", "new TStrings(['x'])"} {
		result := frontend.Compile(`
type TStrings = array of String;
type TStringsHelper = helper for TStrings
  function Create(values: array of String): TStrings;
  begin Result := values; end;
end;
var items := `+call+`;`, "e15_invalid_array_create.dws", semantic.HintsLevelDisabled)
		if !result.HasFatalDiagnostics() {
			t.Errorf("%s compiled despite lacking a class helper function or class constructor", call)
		}
	}
}
