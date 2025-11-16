package interp

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestClassVariableAccessViaClassNameRuntime(t *testing.T) {
	input := `
type TBase = class
  class var Test : Integer := 42;
end;

PrintLn(TBase.Test);
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	analyzer := semantic.NewAnalyzer()
	if err := analyzer.Analyze(program); err != nil {
		t.Fatalf("semantic errors: %v", analyzer.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	if semanticInfo := analyzer.GetSemanticInfo(); semanticInfo != nil {
		interp.SetSemanticInfo(semanticInfo)
	}

	result := interp.Eval(program)

	if result != nil && result.Type() == "ERROR" {
		t.Fatalf("runtime error: %v", result.String())
	}

	output := buf.String()
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestClassVariableAccessViaInstanceRuntime(t *testing.T) {
	input := `
type TBase = class
  class var Test : Integer := 42;
end;

var b : TBase;
PrintLn(b.Test);
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	analyzer := semantic.NewAnalyzer()
	if err := analyzer.Analyze(program); err != nil {
		t.Fatalf("semantic errors: %v", analyzer.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	if semanticInfo := analyzer.GetSemanticInfo(); semanticInfo != nil {
		interp.SetSemanticInfo(semanticInfo)
	}

	result := interp.Eval(program)

	if result != nil && result.Type() == "ERROR" {
		t.Fatalf("runtime error: %v", result.String())
	}

	output := buf.String()
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
