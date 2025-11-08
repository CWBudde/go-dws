package interp

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// Test implicit Self property access in methods
func TestImplicitSelfPropertyAccessWithContext(t *testing.T) {
	input := `
type
  TTest = class
  private
    FCanRun: Boolean;
    FLine: String;
    function GetLine: String;
    procedure SetLine(aLine: String);
  public
    constructor Create;
    property CanRun: Boolean read FCanRun write FCanRun;
    property Line: String read GetLine write SetLine;
  end;

constructor TTest.Create;
begin
  CanRun := true;  // Implicit Self.CanRun
  Line := 'Initial';  // Implicit Self.Line
end;

function TTest.GetLine: String;
begin
  Result := FLine;
end;

procedure TTest.SetLine(aLine: String);
begin
  FLine := aLine;
end;

var t := TTest.Create;
PrintLn(t.CanRun);
PrintLn(t.Line);
t.CanRun := false;
t.Line := 'Updated';
PrintLn(t.CanRun);
PrintLn(t.Line);
`

	var buf bytes.Buffer
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parse errors: %v", p.Errors())
	}

	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("Unexpected error: %v", result)
	}

	expected := "True\nInitial\nFalse\nUpdated\n"
	actual := buf.String()
	if actual != expected {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expected, actual)
	}
}
