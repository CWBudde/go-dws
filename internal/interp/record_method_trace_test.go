package interp

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/fixtureconfig"
	"github.com/cwbudde/go-dws/internal/frontend"
	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestRecordMethodTrace_InterfaceFixture(t *testing.T) {
	const category = "InterfacesPass"
	path := filepath.Join(fixturesRoot, category, "intf_in_record.pas")
	if got, detail := runFixtureTest(path, false, fixtureconfig.HintsLevel(category)); got != testResultPassed {
		t.Fatalf("intf_in_record: %v: %s", got, detail)
	}
}

func TestRecordMethodTrace_InterfaceCallForms(t *testing.T) {
	source, err := os.ReadFile(filepath.Join(fixturesRoot, "InterfacesPass", "intf_in_record.pas"))
	if err != nil {
		t.Fatal(err)
	}
	for _, call := range []string{"r.Proc;", "r.Proc();"} {
		t.Run(call, func(t *testing.T) {
			input := strings.ReplaceAll(string(source), "r.Proc;", call)
			compiled := frontend.Compile(input, "record_trace.pas", semantic.HintsLevelDisabled)
			if compiled.HasFatalDiagnostics() || !compiled.SemanticSuccessful {
				t.Fatalf("compile diagnostics: %v", compiled.DiagnosticStrings())
			}
			var output bytes.Buffer
			engine := New(&output)
			engine.SetSemanticInfo(compiled.SemanticInfo)
			result := engine.Eval(compiled.Program)
			const want = "ERROR: Interface is nil [line: 18, column: 10]\n [line: 26, column: 3]"
			if result == nil || result.String() != want {
				t.Fatalf("runtime result = %v, want %q", result, want)
			}
		})
	}
}

func TestRecordMethodTrace_CaughtErrorKeepsCallerAndPopsFrame(t *testing.T) {
	runScriptTestWithSemantic(t, `type IIntf = interface procedure Proc; end;
type TRec = record
  FIntf: IIntf;
  procedure Proc;
  begin
    FIntf.Proc;
  end;
end;
var r: TRec;
try
  r.Proc();
except
  on E: Exception do begin
    PrintLn(E.Message);
    PrintLn(E.StackTrace);
  end;
end;
PrintLn(Length(GetStackTrace()));`, "Interface is nil [line: 6, column: 11]\n [line: 11, column: 5]\n0")
}

func TestRecordMethodTrace_ExceptionPreservesClassState(t *testing.T) {
	runScriptTestWithSemantic(t, `type TRec = record
  class var Count: Integer;
  Value: Integer;
  procedure Fail;
  begin
    Count := 42;
    Value := 7;
    raise Exception.Create('stop');
  end;
end;
var r: TRec;
try
  r.Fail;
except
  on E: Exception do PrintLn(E.Message);
end;
PrintLn(TRec.Count);
PrintLn(r.Value);
PrintLn(Length(GetStackTrace()));`, "stop\n42\n7\n0")
}
