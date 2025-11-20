package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// TestResolveClassOfTypeNode tests the resolveClassOfTypeNode function
func TestResolveClassOfTypeNode(t *testing.T) {
	code := `
		type TMyClass = class
			procedure DoSomething;
		end;

		procedure TMyClass.DoSomething;
		begin
		end;

		type TMyClassRef = class of TMyClass;

		var ref: TMyClassRef;
	`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	analyzer.Analyze(program)

	if len(analyzer.errors) > 0 {
		t.Errorf("unexpected errors: %v", analyzer.errors)
	}
}
