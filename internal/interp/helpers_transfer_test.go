package interp

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/frontend"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/pkg/ident"
)

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

	compileResult := frontend.Compile(source, "case_mismatch.dws", semantic.HintsLevelDisabled)
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

	if got, want := buf.String(), "42\n"; got != want {
		t.Errorf("program output = %q, want %q", got, want)
	}

	instances := make(map[*runtime.MutableHelperInfo]bool)
	for _, helpers := range interp.typeSystem.AllHelpers() {
		for _, helperAny := range helpers {
			if helperInfo, ok := helperAny.(*runtime.MutableHelperInfo); ok && ident.Equal(helperInfo.Name, "TFooHelper") {
				instances[helperInfo] = true
			}
		}
	}
	if len(instances) != 1 {
		t.Errorf("helper registered as %d distinct MutableHelperInfo instances, want exactly 1", len(instances))
	}
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

	compileResult := frontend.Compile(source, "single_instance.dws", semantic.HintsLevelDisabled)
	if compileResult.HasFatalDiagnostics() || !compileResult.SemanticSuccessful {
		t.Fatalf("compile failed: %v", compileResult.DiagnosticStrings())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	if compileResult.SemanticInfo != nil {
		interp.SetSemanticInfo(compileResult.SemanticInfo)
	}
	// Mirror the CLI pipeline: transfer semantic helpers, then evaluate the
	// program (whose declarations run through the evaluator's VisitHelperDecl).
	interp.TransferHelpersFromSemanticAnalysis(compileResult.Analyzer.GetHelpers())

	if result := interp.Eval(compileResult.Program); result != nil && result.Type() == "ERROR" {
		t.Fatalf("evaluation failed: %s", result.String())
	}

	if got, want := buf.String(), "base\nabab\n"; got != want {
		t.Errorf("program output = %q, want %q", got, want)
	}

	instances := make(map[string]map[*runtime.MutableHelperInfo]bool)
	for _, helpers := range interp.typeSystem.AllHelpers() {
		for _, helperAny := range helpers {
			helperInfo, ok := helperAny.(*runtime.MutableHelperInfo)
			if !ok {
				continue
			}
			name := ident.Normalize(helperInfo.Name)
			if name != "tbase" && name != "tchild" {
				continue
			}
			if instances[name] == nil {
				instances[name] = make(map[*runtime.MutableHelperInfo]bool)
			}
			instances[name][helperInfo] = true
		}
	}

	for _, name := range []string{"tbase", "tchild"} {
		switch got := len(instances[name]); {
		case got == 0:
			t.Errorf("helper %q not registered", name)
		case got > 1:
			t.Errorf("helper %q registered as %d distinct MutableHelperInfo instances, want exactly 1", name, got)
		}
	}

	// The child's parent link must point at the single registered instance.
	if len(instances["tbase"]) == 1 && len(instances["tchild"]) == 1 {
		var base, child *runtime.MutableHelperInfo
		for h := range instances["tbase"] {
			base = h
		}
		for h := range instances["tchild"] {
			child = h
		}
		if child.ParentHelper != base {
			t.Errorf("child helper's ParentHelper is not the registered base instance")
		}
	}
}
