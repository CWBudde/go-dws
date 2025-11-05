package interp

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

func TestImplicitSelfPropertyAccess(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Simple property write",
			input: `
type
  TTest = class
  private
    FValue: Integer;
  public
    constructor Create;
    property Value: Integer read FValue write FValue;
    procedure SetIt;
  end;

constructor TTest.Create;
begin
  PrintLn('Constructor');
end;

procedure TTest.SetIt;
begin
  Value := 42;
end;

var t := TTest.Create();
t.SetIt;
PrintLn(t.Value);
`,
			expected: "Constructor\n42\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parse errors: %v", p.Errors())
			}

			interp := New(&buf)
			result := interp.Eval(program)

			if result != nil && result.Type() == "ERROR" {
				t.Fatalf("Runtime error: %s", result)
			}

			output := buf.String()
			if output != tt.expected {
				t.Errorf("wrong output.\nexpected=%q\ngot=%q", tt.expected, output)
			}
		})
	}
}

func TestConstructorWithoutParentheses(t *testing.T) {
	input := `
type
  TTest = class
    constructor Create;
  end;

constructor TTest.Create;
begin
  PrintLn('Constructor');
end;

var obj := TTest.Create;
if obj = nil then
  PrintLn('NIL')
else
  PrintLn('OK');
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

	if result != nil && result.Type() == "ERROR" {
		t.Fatalf("Runtime error: %s", result)
	}

	output := buf.String()
	expected := "Constructor\nOK\n"
	if output != expected {
		t.Errorf("wrong output.\nexpected=%q\ngot=%q", expected, output)
	}
}
