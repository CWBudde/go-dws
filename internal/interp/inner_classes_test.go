package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/pkg/ident"
)

func TestNestedClassRegistration(t *testing.T) {
	input := `type TOuter = class
   type TInner = class
      Value: Integer;
   end;
end;`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	analyzer := semantic.NewAnalyzer()
	analyzer.SetSource(input, "nested_class_registration")
	if err := analyzer.Analyze(program); err != nil {
		t.Fatalf("semantic analysis failed: %v", err)
	}

	interp := New(&bytes.Buffer{})
	val := interp.Eval(program)
	if isError(val) {
		t.Fatalf("interpreter returned error: %s", val.String())
	}

	classInfo := interp.classes[ident.Normalize("TOuter")]
	if classInfo == nil {
		t.Fatalf("outer class not registered")
	}

	if _, ok := classInfo.NestedClasses[ident.Normalize("TInner")]; !ok {
		t.Fatalf("nested class not registered on outer class")
	}
}

func TestNestedClassInstantiation(t *testing.T) {
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

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	analyzer := semantic.NewAnalyzer()
	analyzer.SetSource(input, "nested_class_instantiation")
	if err := analyzer.Analyze(program); err != nil {
		t.Fatalf("semantic analysis failed: %v", err)
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)
	if isError(result) {
		t.Fatalf("interpreter returned error: %s", result.String())
	}

	if got := strings.TrimSpace(buf.String()); got != "ok" {
		t.Fatalf("unexpected output: %q", got)
	}
}
