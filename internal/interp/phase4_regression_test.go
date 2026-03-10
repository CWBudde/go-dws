package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
)

func testEvalWithOutputAndSemantic(t *testing.T, input string) (Value, string) {
	t.Helper()

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
	}

	analyzer := semantic.NewAnalyzer()
	analyzer.SetSource(input, "phase4_regression")
	if err := analyzer.Analyze(program); err != nil {
		t.Fatalf("semantic analysis failed: %v", err)
	}

	var buf bytes.Buffer
	interp := New(&buf)
	if info := analyzer.GetSemanticInfo(); info != nil {
		interp.SetSemanticInfo(info)
	}

	return interp.Eval(program), buf.String()
}

func TestPhase4Regression_AssignmentCluster(t *testing.T) {
	input := `
type TCounter = class
	FValue: Integer;
	class var Shared: Integer;
	property Value: Integer read FValue write FValue;
	constructor Create; begin end;
end;

var obj := TCounter.Create();
obj.Value := 7;
TCounter.Shared := 11;
PrintLn(obj.Value);
PrintLn(TCounter.Shared);
`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("unexpected error: %s", result)
	}

	expected := "7\n11\n"
	if output != expected {
		t.Fatalf("wrong output. expected=%q got=%q", expected, output)
	}
}

func TestPhase4Regression_UserFunctionCluster(t *testing.T) {
	input := `
function Add(a, b: Integer): Integer;
begin
	Result := a + b;
end;

procedure Bump(var x: Integer);
begin
	x := Add(x, 1);
end;

var n := Add(2, 3);
Bump(n);
PrintLn(n);
`

	result, output := testEvalWithOutput(input)
	if isError(result) {
		t.Fatalf("unexpected error: %s", result)
	}

	expected := "6\n"
	if output != expected {
		t.Fatalf("wrong output. expected=%q got=%q", expected, output)
	}
}

func TestPhase4Regression_DeclarationCluster(t *testing.T) {
	input := `type TOuter = class
   type TInner = class
      Value: String;
   end;
   Inner: TInner;
   procedure Init;
   begin
      Inner := new TInner;
      Inner.Value := 'ok';
   end;
end;

var o := new TOuter;
o.Init;
PrintLn(o.Inner.Value);`

	result, output := testEvalWithOutputAndSemantic(t, input)
	if isError(result) {
		t.Fatalf("unexpected error: %s", result)
	}

	if got := strings.TrimSpace(output); got != "ok" {
		t.Fatalf("wrong output. expected=%q got=%q", "ok", got)
	}
}
