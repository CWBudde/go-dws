package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/frontend"
	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestExplicitHelperCalls(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{
			name: "integer receiver and class methods",
			source: `
type TNumber = helper for Integer
   function Add(n: Integer): Integer; begin Result := Self + n; end;
   class function Two: Integer; begin Result := 2; end;
end;
PrintLn(TNumber.Add(40, 2));
PrintLn(TNumber.Two);
PrintLn(TNumber.Two());
PrintLn((40).Add(2));
`,
			want: "42\n2\n2\n42\n",
		},
		{
			name: "object receiver retains mutations",
			source: `
type TBox = class
   Value: Integer;
end;
type TBoxHelper = helper for TBox
   procedure Add(n: Integer); begin Self.Value += n; end;
end;
var box := TBox.Create;
box.Value := 40;
TBoxHelper.Add(box, 2);
PrintLn(box.Value);
`,
			want: "42\n",
		},
		{
			name: "named identity and case insensitive lookup",
			source: `
type TFirst = helper for Integer
   function Label: String; begin Result := 'first'; end;
end;
type TLast = helper for Integer
   function Label: String; begin Result := 'last'; end;
end;
PrintLn(tFiRsT.lAbEl(1));
PrintLn(TLast.Label(1));
`,
			want: "first\nlast\n",
		},
		{
			name: "local variable shadows helper name",
			source: `
type TNumber = helper for Integer
   function Next: Integer; begin Result := Self + 1; end;
end;
procedure Test;
begin
   var tNuMbEr := 41;
   PrintLn(tNuMbEr.Next);
   PrintLn(tNuMbEr.Next());
end;
Test;
`,
			want: "42\n42\n",
		},
		{
			name: "class receiver and static class methods",
			source: `
type TBase = class end;
type TChild = class(TBase) end;
type TClassHelper = helper for TBase
   class function Name: String; begin Result := Self.ClassName; end;
   class function StaticName: String; static; begin Result := 'static'; end;
end;
PrintLn(TClassHelper.Name(TChild));
PrintLn(TClassHelper.StaticName);
PrintLn(TClassHelper.StaticName());
`,
			want: "TChild\nstatic\nstatic\n",
		},
		{
			name: "out of line static class method",
			source: `
type TBase = class end;
type H = helper for TBase
   class function Get: Integer; static;
end;
class function H.Get: Integer;
begin Result := 3; end;
PrintLn(H.Get());
`,
			want: "3\n",
		},
		{
			name: "overloaded lazy argument remains lazy",
			source: `
type H = helper for Integer
   class function Twice(lazy n: Integer): Integer; overload;
   begin Result := n + n; end;
   class function Twice(s: String): Integer; overload;
   begin Result := 0; end;
end;
var calls := 0;
function Next: Integer;
begin Inc(calls); Result := calls; end;
PrintLn(H.Twice(Next()));
PrintLn(calls);
`,
			want: "3\n2\n",
		},
		{
			name: "record class method explicit type receiver",
			source: `
type R = record x: Integer; end;
type H = helper for R
   class function Get: Integer; begin Result := 3; end;
end;
PrintLn(H.Get(R));
`,
			want: "3\n",
		},
		{
			name: "default lazy and var arguments",
			source: `
type TNumber = helper for Integer
   function Add(n: Integer = 2): Integer; begin Result := Self + n; end;
   procedure Change(var n: Integer; lazy unused: Integer);
   begin n := Self + n; end;
end;
function Unused: Integer;
begin PrintLn('unexpected lazy evaluation'); Result := 100; end;
PrintLn(TNumber.Add(40));
var n := 2;
TNumber.Change(40, n, Unused());
PrintLn(n);
`,
			want: "42\n42\n",
		},
		{
			name: "inherited helper method",
			source: `
type TBase = helper for Integer
   function Add(n: Integer): Integer; begin Result := Self + n; end;
end;
type TChild = helper(TBase) for Integer
   function Twice: Integer; begin Result := Self * 2; end;
end;
PrintLn(TChild.Add(40, 2));
PrintLn(TChild.Twice(21));
`,
			want: "42\n42\n",
		},
		{
			name: "same arity overloads and mixed class instance methods",
			source: `
type TNumber = helper for Integer
   function Pick(n: Integer): String; overload;
   begin Result := 'integer'; end;
   function Pick(s: String): String; overload;
   begin Result := 'string'; end;
   class function Pick(s: String; n: Integer): String; overload;
   begin Result := 'class'; end;
end;
PrintLn(TNumber.Pick(1, 2));
PrintLn(TNumber.Pick(1, 'x'));
PrintLn(TNumber.Pick('x', 2));
`,
			want: "integer\nstring\nclass\n",
		},
		{
			name: "receiver and arguments evaluate once in source order",
			source: `
type TNumber = helper for Integer
   function Combine(a, b: Integer): Integer;
   begin PrintLn('body'); Result := Self + a + b; end;
end;
function Receiver: Integer;
begin PrintLn('receiver'); Result := 1; end;
function First: Integer;
begin PrintLn('first'); Result := 2; end;
function Second: Integer;
begin PrintLn('second'); Result := 3; end;
PrintLn(TNumber.Combine(Receiver(), First(), Second()));
`,
			want: "receiver\nfirst\nsecond\nbody\n6\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compileAndRunWithHelperTransfer(t, tt.source, "explicit_helper.dws", tt.want)
		})
	}
}

func TestExplicitHelperCalls_InvalidReceiver(t *testing.T) {
	const declaration = `type TNumber = helper for Integer
   function Next: Integer; begin Result := Self + 1; end;
end;
`
	for _, call := range []string{
		"TNumber.Next;",
		"TNumber.Next();",
		"TNumber.Next('wrong');",
	} {
		t.Run(call, func(t *testing.T) {
			result := frontend.Compile(declaration+call, "invalid_receiver.dws", semantic.HintsLevelDisabled)
			if !result.HasFatalDiagnostics() && result.SemanticSuccessful {
				t.Fatal("invalid explicit receiver compiled successfully")
			}
			if diagnostics := strings.Join(result.DiagnosticStrings(), "\n"); !strings.Contains(diagnostics, "line: 4") && !strings.Contains(diagnostics, "4:") {
				t.Errorf("diagnostics do not identify the call site: %s", diagnostics)
			}
		})
	}
}

func TestExplicitHelperCalls_MissingClassReceiver(t *testing.T) {
	const declaration = `type TBase = class end;
type TClassHelper = helper for TBase
   class function Name: String; begin Result := Self.ClassName; end;
end;
`
	for _, call := range []string{"TClassHelper.Name;", "TClassHelper.Name();"} {
		t.Run(call, func(t *testing.T) {
			result := frontend.Compile(declaration+call, "missing_class_receiver.dws", semantic.HintsLevelDisabled)
			if !result.HasFatalDiagnostics() && result.SemanticSuccessful {
				t.Fatal("class helper method compiled without its metaclass receiver")
			}
		})
	}
}

func TestExplicitHelperCalls_NoUnrelatedHelperFallback(t *testing.T) {
	const declarations = `
type TNamed = helper for Integer
   class function Own: Integer; begin Result := 1; end;
end;
type TOther = helper for Integer
   class function Other: Integer; begin Result := 2; end;
end;
`
	for _, call := range []string{"TNamed.Other;", "TNamed.Other();"} {
		t.Run(call, func(t *testing.T) {
			result := frontend.Compile(declarations+call, "unrelated_helper.dws", semantic.HintsLevelDisabled)
			if !result.HasFatalDiagnostics() && result.SemanticSuccessful {
				t.Fatal("named helper borrowed a method from an unrelated helper")
			}
		})
	}
}

func TestExplicitHelperCalls_InvalidRecordClassReceiver(t *testing.T) {
	const declarations = `
type R = record x: Integer; end;
type H = helper for R
   class function Get: Integer; begin Result := 3; end;
end;
`
	for _, call := range []string{
		"H.Get();",
		"var value: R; H.Get(value);",
		"procedure Test; begin var R := 1; H.Get(R); end; Test;",
	} {
		t.Run(call, func(t *testing.T) {
			result := frontend.Compile(declarations+call, "invalid_record_receiver.dws", semantic.HintsLevelDisabled)
			if !result.HasFatalDiagnostics() && result.SemanticSuccessful {
				t.Fatal("record class helper method compiled without a record type receiver")
			}
		})
	}
}

func TestExplicitHelperCalls_ReceiverFailureStopsArguments(t *testing.T) {
	const source = `
type TNumber = helper for Integer
   procedure Consume(n: Integer); begin PrintLn('body'); end;
end;
function Receiver: Integer;
begin
   PrintLn('receiver');
   raise Exception.Create('receiver failed');
end;
function Argument: Integer;
begin PrintLn('argument'); Result := 1; end;
TNumber.Consume(Receiver(), Argument());
`
	compiled := frontend.Compile(source, "failed_receiver.dws", semantic.HintsLevelDisabled)
	if compiled.HasFatalDiagnostics() || !compiled.SemanticSuccessful {
		t.Fatalf("compile failed: %v", compiled.DiagnosticStrings())
	}
	var output bytes.Buffer
	i := New(&output)
	i.SetSemanticInfo(compiled.SemanticInfo)
	i.TransferHelpersFromSemanticAnalysis(compiled.Analyzer.GetHelpers())
	result := i.Eval(compiled.Program)
	if result == nil || result.Type() != "ERROR" {
		t.Fatalf("expected receiver failure, got %v", result)
	}
	if got := output.String(); got != "receiver\n" {
		t.Errorf("output = %q, want only receiver output", got)
	}
}
