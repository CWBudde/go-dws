package interp

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestConstructorAssignment_MemoryFixture(t *testing.T) {
	path := filepath.Join(fixturesRoot, "Memory", "obj_fields.pas")
	result, detail := runFixtureTest(path, false, semantic.HintsLevelNormal)
	if result != testResultPassed {
		t.Fatalf("Memory/obj_fields: %v: %s", result, detail)
	}
}

func TestConstructorAssignment_DefaultConstructor(t *testing.T) {
	for _, constructor := range []string{"TTarget.Create", "TTarget.Create()"} {
		t.Run(constructor, func(t *testing.T) {
			got := runQuickwinScript(t, `
type TTarget = class
  Value: Integer;
end;
`+constructor+`.Value := 42;
PrintLn('assigned');
`)
			assertOutput(t, got, "assigned\n")
		})
	}
}

// The constructor retains Self so writes into its temporary result remain
// observable. The counters also detect construction or setter evaluation twice.
func TestConstructorAssignment_ReceiverEvaluatedOnce(t *testing.T) {
	const declarations = `
type TTarget = class
  class var Last: TTarget;
  class var Created: Integer;
  class var Writes: Integer;
  Value: Integer;
  constructor Create;
  begin
    TTarget.Created := TTarget.Created + 1;
    TTarget.Last := Self;
    Value := 10;
  end;
  constructor Build;
  begin
    TTarget.Created := TTarget.Created + 1;
    TTarget.Last := Self;
    Value := 20;
  end;
  procedure SetValue(value: Integer);
  begin
    TTarget.Writes := TTarget.Writes + 1;
    Self.Value := value;
  end;
  property Prop: Integer read Value write SetValue;
end;
type TChild = class(TTarget) end;
type TTargetClass = class of TTarget;
`
	for _, tt := range []struct {
		name, setup, target, want string
	}{
		{"bare field", "", "TTarget.Create.Value", "1\n0\n42\nTTarget\n"},
		{"explicit field", "", "TTarget.Create().Value", "1\n0\n42\nTTarget\n"},
		{"bare property", "", "TTarget.Create.Prop", "1\n1\n42\nTTarget\n"},
		{"explicit property", "", "TTarget.Create().Prop", "1\n1\n42\nTTarget\n"},
		{"inherited constructor", "", "TChild.Create.Prop", "1\n1\n42\nTChild\n"},
		{"named constructor", "", "TTarget.Build.Prop", "1\n1\n42\nTTarget\n"},
		{"inherited named constructor", "", "TChild.Build.Value", "1\n0\n42\nTChild\n"},
		{"case insensitive", "", "ttarget.cReAtE.pRoP", "1\n1\n42\nTTarget\n"},
		{"metaclass variable", "var targetClass: TTargetClass := TChild;", "targetClass.Create.Prop", "1\n1\n42\nTChild\n"},
		{"computed metaclass", "function ChooseClass: TTargetClass; begin PrintLn('receiver'); Result := TChild; end;", "ChooseClass().Create.Prop", "receiver\n1\n1\n42\nTChild\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := runQuickwinScript(t, declarations+tt.setup+tt.target+` := 42;
PrintLn(TTarget.Created);
PrintLn(TTarget.Writes);
PrintLn(TTarget.Last.Value);
PrintLn(TTarget.Last.ClassName);
`)
			assertOutput(t, got, tt.want)
		})
	}
}

func TestConstructorAssignment_ExceptionSkipsWrite(t *testing.T) {
	for _, constructor := range []string{"TTarget.Create", "TTarget.Create()"} {
		t.Run(constructor, func(t *testing.T) {
			got := runQuickwinScript(t, `
type TTarget = class
  class var Created: Integer;
  class var Writes: Integer;
  constructor Create;
  begin
    TTarget.Created := TTarget.Created + 1;
    raise Exception.Create('construction failed');
  end;
  procedure SetValue(value: Integer);
  begin TTarget.Writes := TTarget.Writes + 1; end;
  property Prop: Integer write SetValue;
end;
try
  `+constructor+`.Prop := 42;
except
  on E: Exception do PrintLn(E.Message);
end;
PrintLn(TTarget.Created);
PrintLn(TTarget.Writes);
`)
			assertOutput(t, got, "construction failed\n1\n0\n")
		})
	}
}

// Arithmetic failures return a runtime ErrorValue, unlike a script raise, which
// stores an exception on the execution context. Both must abort the write.
func TestConstructorAssignment_RuntimeErrorSkipsWrite(t *testing.T) {
	for _, constructor := range []string{"TTarget.Create", "TTarget.Create()"} {
		t.Run(constructor, func(t *testing.T) {
			got := runQuickwinScript(t, `
type TTarget = class
  class var Created: Integer;
  class var Writes: Integer;
  constructor Create;
  begin
    TTarget.Created := TTarget.Created + 1;
    var zero := 0;
    PrintLn(1 div zero);
  end;
  procedure SetValue(value: Integer);
  begin TTarget.Writes := TTarget.Writes + 1; end;
  property Prop: Integer write SetValue;
end;
try
  `+constructor+`.Prop := 42;
except
  on E: Exception do PrintLn(E.Message);
end;
PrintLn(TTarget.Created);
PrintLn(TTarget.Writes);
`)
			lines := strings.Split(strings.TrimSpace(got), "\n")
			if len(lines) != 3 || !strings.HasPrefix(lines[0], "Division by zero") ||
				!strings.Contains(lines[0], "TTarget.Create") || lines[1] != "1" || lines[2] != "0" {
				t.Fatalf("expected original constructor error, one construction and no property write, got %q", got)
			}
		})
	}
}

func TestConstructorAssignment_StrictWritableExceptionPreserved(t *testing.T) {
	got := runQuickwinScript(t, `
type TTarget = class
  Value: Integer;
end;
function Fail: TTarget;
begin raise Exception.Create('receiver failed'); end;
try
  TryStrToInt('1', Fail().Value);
except
  on E: Exception do PrintLn(E.Message);
end;
PrintLn('continued');
`)
	// A successful conversion attempts to invoke the writable target's setter.
	// The pending receiver exception must not produce a successful binding with
	// a nil setter, which would panic in the host instead of entering except.
	assertOutput(t, got, "receiver failed\ncontinued\n")
}
