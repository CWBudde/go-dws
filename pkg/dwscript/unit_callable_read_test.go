package dwscript

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

const callableFactoryUnit = `unit Factories;
interface
type TProc = procedure;
type TFactory = function: TProc;
function Make: TProc;
function Scalar: Integer;
function Captured: TProc;
implementation
procedure Inner; begin PrintLn('inner'); end;
function Make: TProc; begin PrintLn('outer'); Result := @Inner; end;
function Scalar: Integer; begin PrintLn('scalar'); Result := 42; end;
function Captured: TProc;
begin
 var value := 42;
 PrintLn(value);
 Result := lambda() begin PrintLn(value); end;
end;
end.`

func TestEngine_UnitQualifiedCallableReads(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Factories.dws"), []byte(callableFactoryUnit), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, source, want string
	}{
		{"typed initializer", "var callback: TProc := Factories.Make; PrintLn('selected'); callback();", "outer\nselected\ninner\n"},
		{"assignment", "var callback: TProc; callback := fAcToRiEs.mAkE; PrintLn('selected'); callback();", "outer\nselected\ninner\n"},
		{"compatible reference", "var factory: TFactory := Factories.Make; PrintLn('captured'); var callback: TProc := factory(); callback();", "captured\nouter\ninner\n"},
		{"compatible assignment", "var factory: TFactory; factory := Factories.Make; PrintLn('captured'); var callback: TProc := factory(); callback();", "captured\nouter\ninner\n"},
		{"inferred", "var callback := Factories.Make; PrintLn('selected'); callback();", "outer\nselected\ninner\n"},
		{"explicit call", "var callback: TProc := Factories.Make(); PrintLn('selected'); callback();", "outer\nselected\ninner\n"},
		{"unqualified", "var callback: TProc := Make; PrintLn('selected'); callback();", "outer\nselected\ninner\n"},
		{"scalar result", "var number: Integer := Factories.Scalar; PrintLn(number); number := Factories.Scalar; PrintLn(number);", "scalar\n42\nscalar\n42\n"},
		{"returned closure", "var callback: TProc := Factories.Captured; PrintLn('selected'); callback(); callback();", "42\nselected\n42\n42\n"},
		{"lexical shadow", `type THolder = class Make: TFactory; end;
procedure SelectCallback(Factories: THolder);
begin var callback: TProc := Factories.Make; callback(); end;
var holder := THolder.Create; holder.Make := @Make; SelectCallback(holder);`, "outer\ninner\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			engine, err := New(WithUnitSearchPaths(dir), WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile("uses Factories;\n" + tc.source)
			if err != nil {
				t.Fatal(err)
			}
			result, err := engine.Run(program)
			if err != nil {
				t.Fatal(err)
			}
			if !result.Success || output.String() != tc.want {
				t.Fatalf("result=%+v output=%q, want %q", result, output.String(), tc.want)
			}
		})
	}
}

func TestEngine_UnitQualifiedFactoryJSONResult(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Factories.dws"), []byte(callableFactoryUnit), 0600); err != nil {
		t.Fatal(err)
	}
	engine, err := New(WithUnitSearchPaths(dir))
	if err != nil {
		t.Fatal(err)
	}
	_, err = engine.Compile("uses Factories;\nvar target: JSONVariant;\ntarget := Factories.Make;")
	var compileError *CompileError
	if !errors.As(err, &compileError) || len(compileError.Errors) != 1 {
		t.Fatalf("expected one JSON assignment error, got %v", err)
	}
	diagnostic := compileError.Errors[0]
	if diagnostic.Message != `Syntax Error: Incompatible types: Cannot assign "procedure TProc" to "JSONVariant"` || diagnostic.Line != 3 || diagnostic.Column != 8 {
		t.Fatalf("unexpected JSON assignment diagnostic: %+v", diagnostic)
	}
}
