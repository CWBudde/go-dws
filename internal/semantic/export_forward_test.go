package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestAnalyze_ExportForwardImplementationHasRealStop(t *testing.T) {
	p := parser.New(lexer.New("procedure P; forward;\nprocedure P; export; begin end;"))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) != 0 {
		t.Fatalf("parse: %v", errs)
	}
	a := NewAnalyzer()
	a.SetSymbolDictionaryDiagnostics(false)
	if err := a.Analyze(program); err == nil {
		t.Fatal("matching forward export accepted")
	}
	errs := a.StructuredErrors()
	if len(errs) != 1 || !errs[0].Stop || errs[0].Message != "BEGIN expected" {
		t.Fatalf("real semantic stop: %+v", errs)
	}
	if !a.compileStopped {
		t.Fatal("stop state did not suppress final checks")
	}
}

func TestAnalyzeUnit_ExportForwardImplementationHasRealStop(t *testing.T) {
	p := parser.New(lexer.New("unit U; interface procedure P; implementation\nprocedure P; export; begin end; end."))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) != 0 {
		t.Fatalf("parse: %v", errs)
	}
	a := NewAnalyzer()
	a.SetSymbolDictionaryDiagnostics(false)
	if err := a.AnalyzeUnit(program.Statements[0].(*ast.UnitDeclaration)); err == nil {
		t.Fatal("unit matching forward export accepted")
	}
	errs := a.StructuredErrors()
	if len(errs) != 1 || !errs[0].Stop || errs[0].Message != "BEGIN expected" {
		t.Fatalf("real unit stop: %+v", errs)
	}
}
