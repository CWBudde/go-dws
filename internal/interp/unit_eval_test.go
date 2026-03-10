package interp

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
)

func TestEval_UnitDeclarationRunsInitializationAfterDeclarations(t *testing.T) {
	input := `unit ToDoModel;

interface

type
  TToDoCategoryList = array of string;

  TToDoModel = partial class
  public
    Category: TToDoCategoryList;
  end;

var
  Model: TToDoModel;

implementation

type
  TToDoModel = partial class
  end;

initialization
  Model := TToDoModel.Create;
  PrintLn(Model.ClassName);
end.`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %s", joinParserErrorsNewline(p.Errors()))
	}

	analyzer := semantic.NewAnalyzer()
	analyzer.SetSource(input, "unit_eval_test")
	if err := analyzer.Analyze(program); err != nil {
		t.Fatalf("semantic analysis failed: %v", err)
	}

	var buf bytes.Buffer
	interp := New(&buf)
	if semanticInfo := analyzer.GetSemanticInfo(); semanticInfo != nil {
		interp.SetSemanticInfo(semanticInfo)
	}

	result := interp.Eval(program)
	if isError(result) {
		t.Fatalf("interpreter returned error: %s", result.String())
	}

	if got := buf.String(); got != "TToDoModel\n" {
		t.Fatalf("unexpected output: %q", got)
	}
}
