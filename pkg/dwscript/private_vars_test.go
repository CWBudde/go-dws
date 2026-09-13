package dwscript

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// compilePrivateVarsProgram uses real temporary files because unit discovery is
// filesystem-backed even when an embedding application installs a platform FS.
func compilePrivateVarsProgram(t *testing.T, units map[string]string, source string) (*Engine, *Program, *bytes.Buffer) {
	t.Helper()
	dir := t.TempDir()
	for name, source := range units {
		if err := os.WriteFile(filepath.Join(dir, name+".pas"), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	output := new(bytes.Buffer)
	engine, err := New(WithUnitSearchPaths(dir), WithOutput(output))
	if err != nil {
		t.Fatal(err)
	}
	program, err := engine.Compile(source)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	return engine, program, output
}

func runPrivateVarsProgram(t *testing.T, engine *Engine, program *Program, output *bytes.Buffer, want string) {
	t.Helper()
	output.Reset()
	result, err := engine.Run(program)
	if err != nil {
		t.Fatalf("Run: %v; output=%q", err, output.String())
	}
	if !result.Success || output.String() != want {
		t.Fatalf("result=%+v output=%q, want %q", result, output.String(), want)
	}
}

func TestPrivateVars_AcceptanceFixture(t *testing.T) {
	fixtureDir := filepath.Join("..", "..", "testdata", "fixtures", "FunctionsGlobalVars")
	read := func(name string) string {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(fixtureDir, name))
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	engine, program, output := compilePrivateVarsProgram(t, map[string]string{
		"unit_private_vars1": read("unit_private_vars1.pas"),
		"unit_private_vars2": read("unit_private_vars2.pas"),
	}, read("private_vars.pas"))
	for range 2 {
		runPrivateVarsProgram(t, engine, program, output, read("private_vars.txt"))
	}
}

func TestPrivateVars_UnitExecution(t *testing.T) {
	tests := []struct {
		name   string
		units  map[string]string
		source string
		want   string
	}{
		{
			name: "lazy defaults and distinct global store",
			units: map[string]string{"PrivateValues": `unit PrivateValues;
interface procedure Run;
implementation
var calls := 0;
function DefaultValue: Integer;
begin Inc(calls); Result := 99; end;
procedure Run;
begin
  CleanupPrivateVars;
  PrintLn(WritePrivateVar('value', 12));
  PrintLn(WritePrivateVar('value', 34));
  PrintLn(ReadPrivateVar('value', DefaultValue()));
  PrintLn(calls);
  PrintLn(ReadPrivateVar('missing', DefaultValue()));
  PrintLn(calls);
  PrintLn(VarIsEmpty(ReadPrivateVar('missing')));
  WriteGlobalVar('value', 'global');
  CleanupGlobalVars;
  PrintLn(ReadPrivateVar('value'));
  WritePrivateVar('Value', 56);
  WritePrivateVar('', 78);
  PrintLn(PrivateVarsNames('').Join(','));
  CleanupPrivateVars('');
  PrintLn(PrivateVarsNames('').Join(','));
  CleanupPrivateVars('V*');
  PrintLn(PrivateVarsNames('*').Join(','));
  CleanupPrivateVars;
end;
end.`},
			source: `uses PrivateValues; PrivateValues.Run;`,
			want:   "True\nFalse\n34\n0\n99\n1\nTrue\n34\n,Value,value\nValue,value\n\n",
		},
		{
			name: "nested routines pointers and escaped lambdas",
			units: map[string]string{"PrivateClosures": `unit PrivateClosures;
interface
type TRead = function: Integer;
function ReadValue: Integer;
procedure Nested;
function MakeReader: TRead;
implementation
function ReadValue: Integer;
begin Result := ReadPrivateVar('value'); end;
procedure Nested;
begin
  procedure Inner;
  begin PrintLn(ReadPrivateVar('value')); end;
  Inner;
end;
function MakeReader: TRead;
begin Result := lambda(): Integer begin Result := ReadPrivateVar('value'); end; end;
initialization CleanupPrivateVars; WritePrivateVar('value', 42);
finalization CleanupPrivateVars;
end.`},
			source: `uses PrivateClosures;
PrintLn(PrivateClosures.ReadValue());
PrintLn(PrivateClosures.ReadValue);
PrivateClosures.Nested();
var reader: TRead := @ReadValue;
PrintLn(reader());
reader := MakeReader();
PrintLn(reader());`,
			want: "42\n42\n42\n42\n42\n",
		},
		{
			name: "overloaded routines",
			units: map[string]string{"PrivateOverloads": `unit PrivateOverloads;
interface
function ReadValue(value: Integer): Integer; overload;
function ReadValue(value: String): String; overload;
implementation
function ReadValue(value: Integer): Integer;
begin Result := ReadPrivateVar('value') + value; end;
function ReadValue(value: String): String;
begin Result := IntToStr(ReadPrivateVar('value')) + value; end;
initialization CleanupPrivateVars; WritePrivateVar('value', 15);
finalization CleanupPrivateVars;
end.`},
			source: `uses PrivateOverloads; PrintLn(ReadValue(2)); PrintLn(ReadValue('x'));`,
			want:   "17\n15x\n",
		},
		{
			name: "main callbacks and exception restoration",
			units: map[string]string{"PrivateCallbacks": `unit PrivateCallbacks;
interface
type TCallback = procedure;
procedure Invoke(callback: TCallback);
implementation
procedure Invoke(callback: TCallback);
begin
  CleanupPrivateVars;
  WritePrivateVar('value', 'unit');
  try callback(); except on E: Exception do PrintLn(E.Message); end;
  PrintLn(ReadPrivateVar('value'));
  CleanupPrivateVars;
end;
end.`},
			source: `uses PrivateCallbacks;
procedure FromMain;
begin ReadPrivateVar('value'); end;
Invoke(@FromMain);
Invoke(lambda begin WritePrivateVar('value', 'main'); end);
try ReadPrivateVar('value'); except on E: Exception do PrintLn(E.Message); end;`,
			want: "Private variables cannot be referred from main module\nunit\nPrivate variables cannot be referred from main module\nunit\nPrivate variables cannot be referred from main module\n",
		},
		{
			name: "constructors methods and properties",
			units: map[string]string{"PrivateObjects": `unit PrivateObjects;
interface
type TPrivate = class
  constructor Create;
  function ReadValue: Integer;
  procedure SetValue(value: Integer);
  class function ReadStatic: Integer;
  property Value: Integer read ReadValue write SetValue;
end;
implementation
constructor TPrivate.Create;
begin CleanupPrivateVars; WritePrivateVar('value', 73); end;
function TPrivate.ReadValue: Integer;
begin Result := ReadPrivateVar('value'); end;
procedure TPrivate.SetValue(value: Integer);
begin WritePrivateVar('value', value); end;
class function TPrivate.ReadStatic: Integer;
begin Result := ReadPrivateVar('value'); end;
finalization CleanupPrivateVars;
end.`},
			source: `uses PrivateObjects;
var instance := TPrivate.Create;
PrintLn(instance.ReadValue());
PrintLn(instance.ReadValue);
PrintLn(instance.Value);
instance.Value := 74;
PrintLn(instance.Value);
PrintLn(TPrivate.ReadStatic());`,
			want: "73\n73\n73\n74\n74\n",
		},
		{
			name: "record and helper methods",
			units: map[string]string{"PrivateRecords": `unit PrivateRecords;
interface
function ReadHelper: Integer;
type TPrivateRecord = record
  function ReadValue: Integer;
  begin Result := ReadPrivateVar('value'); end;
end;
type TPrivateHelper = helper for Integer
  function PrivateValue: Integer;
  begin Result := ReadPrivateVar('value') + Self; end;
end;
implementation
function ReadHelper: Integer;
begin var number := 2; Result := number.PrivateValue(); end;
initialization CleanupPrivateVars; WritePrivateVar('value', 21);
finalization CleanupPrivateVars;
end.`},
			source: `uses PrivateRecords;
var item: TPrivateRecord;
PrintLn(item.ReadValue());
PrintLn(ReadHelper());`,
			want: "21\n23\n",
		},
		{
			name: "inherited contracts retain base ownership",
			units: map[string]string{
				"PrivateBase": `unit PrivateBase;
interface
type TPrivateBase = class
  procedure Check; virtual;
end;
implementation
procedure TPrivateBase.Check;
require ReadPrivateVar('owner') = 'base';
begin PrintLn(ReadPrivateVar('owner'));
ensure ReadPrivateVar('owner') = 'base';
end;
initialization CleanupPrivateVars; WritePrivateVar('owner', 'base');
finalization CleanupPrivateVars;
end.`,
				"PrivateChild": `unit PrivateChild;
interface uses PrivateBase;
type TPrivateChild = class(TPrivateBase)
  procedure Check; override;
end;
implementation
procedure TPrivateChild.Check;
begin PrintLn(ReadPrivateVar('owner')); end;
initialization CleanupPrivateVars; WritePrivateVar('owner', 'child');
finalization CleanupPrivateVars;
end.`,
			},
			source: `uses PrivateChild; var instance := TPrivateChild.Create; instance.Check;`,
			want:   "child\n",
		},
		{
			name: "initialization finalization and cross unit calls",
			units: map[string]string{
				"PrivateFirst": `unit PrivateFirst;
interface function ReadValue: String;
implementation
function ReadValue: String;
begin Result := ReadPrivateVar('value'); end;
initialization CleanupPrivateVars; WritePrivateVar('value', 'first'); PrintLn(ReadPrivateVar('value'));
finalization PrintLn(ReadPrivateVar('value')); CleanupPrivateVars;
end.`,
				"PrivateSecond": `unit PrivateSecond;
interface procedure Run;
implementation uses PrivateFirst;
procedure Run;
begin PrintLn(PrivateFirst.ReadValue); PrintLn(ReadPrivateVar('value')); end;
initialization CleanupPrivateVars; WritePrivateVar('value', 'second'); PrintLn(ReadPrivateVar('value'));
finalization PrintLn(ReadPrivateVar('value')); CleanupPrivateVars;
end.`,
			},
			source: `uses PrivateSecond; PrivateSecond.Run;`,
			want:   "first\nsecond\nfirst\nsecond\nsecond\nfirst\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, program, output := compilePrivateVarsProgram(t, tt.units, tt.source)
			runPrivateVarsProgram(t, engine, program, output, tt.want)
		})
	}
}

func TestPrivateVars_MainModuleRejected(t *testing.T) {
	for _, call := range []string{
		`ReadPrivateVar('value')`,
		`WritePrivateVar('value', 1)`,
		`PrivateVarsNames('*')`,
		`CleanupPrivateVars`,
	} {
		t.Run(call, func(t *testing.T) {
			engine, program, output := compilePrivateVarsProgram(t, nil,
				`try `+call+`; except on E: Exception do PrintLn(E.Message); end;`)
			runPrivateVarsProgram(t, engine, program, output, "Private variables cannot be referred from main module\n")
		})
	}
}

func TestPrivateVars_PersistAcrossRuns(t *testing.T) {
	const unitSource = `unit PrivatePersistence;
interface procedure Increment; procedure Reset;
implementation
procedure Increment;
begin
  var value: Integer := ReadPrivateVar('counter', 0);
  Inc(value);
  WritePrivateVar('counter', value);
  PrintLn(value);
end;
procedure Reset;
begin CleanupPrivateVars; end;
end.`
	engine, program, output := compilePrivateVarsProgram(t, map[string]string{
		"PrivatePersistence": unitSource,
	}, `uses PrivatePersistence; PrivatePersistence.Increment();`)
	reset, err := engine.Compile(`uses PrivatePersistence; PrivatePersistence.Reset;`)
	if err != nil {
		t.Fatal(err)
	}
	runPrivateVarsProgram(t, engine, reset, output, "")
	defer runPrivateVarsProgram(t, engine, reset, output, "")
	runPrivateVarsProgram(t, engine, program, output, "1\n")
	runPrivateVarsProgram(t, engine, program, output, "2\n")
	// Unit identifiers share one process-wide partition regardless of casing or
	// which Engine compiled their declarations.
	otherEngine, otherProgram, otherOutput := compilePrivateVarsProgram(t, map[string]string{
		"PRIVATEPERSISTENCE": strings.Replace(unitSource, "unit PrivatePersistence;", "unit PRIVATEPERSISTENCE;", 1),
	}, `uses PRIVATEPERSISTENCE; PRIVATEPERSISTENCE.Increment();`)
	runPrivateVarsProgram(t, otherEngine, otherProgram, otherOutput, "3\n")
}
