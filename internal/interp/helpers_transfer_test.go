package interp

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/frontend"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// compileAndRunWithHelperTransfer mirrors the CLI pipeline: compile, transfer
// semantic helpers, evaluate the program, and assert its output. It returns
// the interpreter so callers can inspect the helper registry.
func compileAndRunWithHelperTransfer(t *testing.T, source, name, wantOutput string) *Interpreter {
	t.Helper()

	compileResult := frontend.Compile(source, name, semantic.HintsLevelDisabled)
	if compileResult.HasFatalDiagnostics() || !compileResult.SemanticSuccessful {
		t.Fatalf("compile failed: %v", compileResult.DiagnosticStrings())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	if compileResult.SemanticInfo != nil {
		interp.SetSemanticInfo(compileResult.SemanticInfo)
	}
	interp.TransferHelpersFromSemanticAnalysis(compileResult.Analyzer.GetHelpers())

	if result := interp.Eval(compileResult.Program); result != nil && result.Type() == "ERROR" {
		t.Fatalf("evaluation failed: %s", result.String())
	}

	if got := buf.String(); got != wantOutput {
		t.Errorf("program output = %q, want %q", got, wantOutput)
	}
	return interp
}

// distinctHelperInstances scans the type system's helper registry and returns,
// for each requested (normalized) helper name, the set of distinct
// *runtime.MutableHelperInfo pointers registered under it.
func distinctHelperInstances(interp *Interpreter, names ...string) map[string]map[*runtime.MutableHelperInfo]bool {
	wanted := make(map[string]bool, len(names))
	for _, name := range names {
		wanted[ident.Normalize(name)] = true
	}

	instances := make(map[string]map[*runtime.MutableHelperInfo]bool)
	for _, helpers := range interp.typeSystem.AllHelpers() {
		for _, helperAny := range helpers {
			helperInfo, ok := helperAny.(*runtime.MutableHelperInfo)
			if !ok {
				continue
			}
			name := ident.Normalize(helperInfo.Name)
			if !wanted[name] {
				continue
			}
			if instances[name] == nil {
				instances[name] = make(map[*runtime.MutableHelperInfo]bool)
			}
			instances[name][helperInfo] = true
		}
	}
	return instances
}

// singleHelperInstance asserts exactly one registered instance for the
// (normalized) helper name and returns it (nil if the assertion failed).
func singleHelperInstance(t *testing.T, instances map[string]map[*runtime.MutableHelperInfo]bool, name string) *runtime.MutableHelperInfo {
	t.Helper()
	set := instances[ident.Normalize(name)]
	if len(set) != 1 {
		t.Errorf("helper %q registered as %d distinct MutableHelperInfo instances, want exactly 1", name, len(set))
		return nil
	}
	for h := range set {
		return h
	}
	return nil
}

// TestUserHelperReuseWithCaseMismatchedTarget verifies that VisitHelperDecl
// reuses the semantic-transfer instance even when the helper declaration
// spells the target type with different casing than the type declaration
// (DWScript is case-insensitive, but ClassType.Equals compares names
// case-sensitively). With two instances, the out-of-line method body would be
// bound into only one of them and dispatch could hit the body-less copy,
// silently returning zero values.
func TestUserHelperReuseWithCaseMismatchedTarget(t *testing.T) {
	source := `
type
   TFoo = class
      x : Integer;
   end;

type
   TFooHelper = helper for tfoo
      function GetX : Integer;
   end;

function TFooHelper.GetX : Integer;
begin
   Result := Self.x + 41;
end;

var f := TFoo.Create;
f.x := 1;
PrintLn(f.GetX);
`

	interp := compileAndRunWithHelperTransfer(t, source, "case_mismatch.dws", "42\n")
	instances := distinctHelperInstances(interp, "TFooHelper")
	singleHelperInstance(t, instances, "TFooHelper")
}

// TestUserHelperRegisteredAsSingleInstance verifies that a user helper is
// backed by exactly one *runtime.MutableHelperInfo instance, even when the
// semantic-transfer path (TransferHelpersFromSemanticAnalysis) and the
// evaluator's VisitHelperDecl both process the same declaration. Duplicate
// instances previously made first-match lookups over AllHelpers depend on Go
// map iteration order.
func TestUserHelperRegisteredAsSingleInstance(t *testing.T) {
	source := `
type
   TBase = helper for String
      function Base : String;
      begin
         Result := 'base';
      end;
   end;

type
   TChild = helper(TBase) for String
      function Doubled : String;
   end;

function TChild.Doubled : String;
begin
   Result := Self + Self;
end;

var s := 'ab';
PrintLn(s.Base);
PrintLn(s.Doubled);
`

	interp := compileAndRunWithHelperTransfer(t, source, "single_instance.dws", "base\nabab\n")
	instances := distinctHelperInstances(interp, "TBase", "TChild")

	base := singleHelperInstance(t, instances, "TBase")
	child := singleHelperInstance(t, instances, "TChild")

	// The child's parent link must point at the single registered instance.
	if base != nil && child != nil && child.ParentHelper != base {
		t.Errorf("child helper's ParentHelper is not the registered base instance")
	}
}
