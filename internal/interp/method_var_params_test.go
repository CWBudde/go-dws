package interp

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// TestMethodVarParams verifies that var parameters of constructors and
// instance methods write back to the caller's variable (see fixture
// oop_field: TMyClass.Create(var value) halves the argument).
func TestMethodVarParams(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "constructor var param writes back",
			input: `
type
   TMyClass = class
      Field : Float;
      constructor Create(var value : Float);
   end;
constructor TMyClass.Create(var value : Float);
begin
   Field := value;
   value := value/2;
end;
var f : Float = 1.5;
var o : TMyClass = TMyClass.Create(f);
PrintLn(o.Field);
PrintLn(f);
`,
			expected: "1.5\n0.75\n",
		},
		{
			name: "instance method var param writes back",
			input: `
type
   TCounter = class
      procedure Bump(var x : Integer);
   end;
procedure TCounter.Bump(var x : Integer);
begin
   x := x + 1;
end;
var c := TCounter.Create;
var n := 10;
c.Bump(n);
c.Bump(n);
PrintLn(n);
`,
			expected: "12\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
			}

			var buf bytes.Buffer
			interp := New(&buf)
			interp.Eval(program)

			if got := buf.String(); got != tt.expected {
				t.Errorf("expected output %q, got %q", tt.expected, got)
			}
		})
	}
}
