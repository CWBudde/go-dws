package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/frontend"
	"github.com/cwbudde/go-dws/internal/semantic"
)

func runRecordRegressionScript(t *testing.T, source string) string {
	t.Helper()

	compiled := frontend.Compile(source, "record_regression.pas", semantic.HintsLevelPedantic)
	if compiled.HasFatalDiagnostics() || !compiled.SemanticSuccessful {
		t.Fatalf("compile diagnostics:\n%s", strings.Join(compiled.DiagnosticStrings(), "\n"))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	if compiled.SemanticInfo != nil {
		interp.SetSemanticInfo(compiled.SemanticInfo)
	}
	result := interp.Eval(compiled.Program)
	if result != nil && result.Type() == "ERROR" {
		t.Fatalf("runtime error: %s", result.String())
	}
	return buf.String()
}

func TestRecordValueParameterCopiesRecord(t *testing.T) {
	got := runRecordRegressionScript(t, `
type
	TPoint = record
		x, y : Integer;
	end;

procedure PassCopy(p : TPoint);
begin
	p.x := 1;
	PrintLn(IntToStr(p.x) + ',' + IntToStr(p.y));
end;

procedure PassVar(var p : TPoint);
begin
	p.y := 2;
	PrintLn(IntToStr(p.x) + ',' + IntToStr(p.y));
end;

var p : TPoint;
PassCopy(p);
PassVar(p);
PrintLn(IntToStr(p.x) + ',' + IntToStr(p.y));
`)
	want := "1,0\n0,2\n0,2\n"
	if got != want {
		t.Fatalf("output mismatch:\nwant:\n%sgot:\n%s", want, got)
	}
}

func TestRecordConstParameterSharesRecordForPropertyWrite(t *testing.T) {
	got := runRecordRegressionScript(t, `
type
	TTest = record
		Field : Integer;
		property Prop : Integer read Field write Field;
	end;

procedure Stuff(const t : TTest);
begin
	t.Prop := t.Prop + 333;
end;

var t : TTest := (Field: 123);
Stuff(t);
PrintLn(t.Field);
`)
	want := "456\n"
	if got != want {
		t.Fatalf("output mismatch:\nwant:\n%sgot:\n%s", want, got)
	}
}

func TestRecordStaticMembersThroughTypeAndInstance(t *testing.T) {
	got := runRecordRegressionScript(t, `
type
	TBase = record
		class var Test : Integer;
		class function Add(a, b : Integer) : Integer;
		begin
			Result := a + b + Test;
		end;
	end;

PrintLn(TBase.Test);
TBase.Test := 5;
PrintLn(TBase.Add(1, 2));

var b : TBase;
PrintLn(b.Test);
PrintLn(b.Add(3, 4));
`)
	want := "0\n8\n5\n12\n"
	if got != want {
		t.Fatalf("output mismatch:\nwant:\n%sgot:\n%s", want, got)
	}
}

func TestRecordAliasFieldInitializesUnderlyingRecord(t *testing.T) {
	got := runRecordRegressionScript(t, `
type
	TBaseInt = Integer;
	TOther = record
		i : TBaseInt;
	end;
	TSub = TOther;
	TPoint = record
		x : TBaseInt;
		y : TSub;
	end;

var p : TPoint;
p.x := 1;
p.y.i := 2;
PrintLn(IntToStr(p.x) + ',' + IntToStr(p.y.i));
`)
	want := "1,2\n"
	if got != want {
		t.Fatalf("output mismatch:\nwant:\n%sgot:\n%s", want, got)
	}
}

func TestRecordInlineFieldTypeAndAnonymousInitializerContext(t *testing.T) {
	got := runRecordRegressionScript(t, `
type
	TParent = record
		A : String;
		Sub : record
			B : String;
		end;
	end;

	TOuter = record
		Sub : TParent = (A: 'hello'; Sub: (B: 'world'));
	end;

var value : TOuter;
PrintLn(value.Sub.A);
PrintLn(value.Sub.Sub.B);
`)
	want := "hello\nworld\n"
	if got != want {
		t.Fatalf("output mismatch:\nwant:\n%sgot:\n%s", want, got)
	}
}

func TestRecordFunctionResultCanBeUsedAsImplicitCallReceiver(t *testing.T) {
	got := runRecordRegressionScript(t, `
type
	TRec = record
		Name : String
	end;

function MakeRec : TRec;
begin
	Result.Name := 'ok';
end;

PrintLn(MakeRec.Name);
`)
	want := "ok\n"
	if got != want {
		t.Fatalf("output mismatch:\nwant:\n%sgot:\n%s", want, got)
	}
}
