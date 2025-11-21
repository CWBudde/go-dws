package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

func TestPartialClassMerging(t *testing.T) {
	input := `
type TTest = partial class
  Field : Integer;
end;

type TTest = partial class
  procedure PrintMe;
  begin
    PrintLn(Field);
  end;
end;
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	analyzer.Analyze(program)

	if len(analyzer.Errors()) > 0 {
		t.Fatalf("Semantic errors: %v", analyzer.Errors())
	}

	// Check if the class was merged
	ttest, ok := analyzer.GetClasses()["ttest"]
	if !ok {
		t.Fatal("TTest class not found!")
	}

	t.Logf("TTest class found:")
	t.Logf("  IsPartial: %v", ttest.IsPartial)
	t.Logf("  Fields: %v", len(ttest.Fields))
	for name := range ttest.Fields {
		t.Logf("    - %s", name)
	}
	t.Logf("  Methods: %v", len(ttest.Methods))
	for name := range ttest.Methods {
		t.Logf("    - %s", name)
	}

	// Verify the class has both the field and the method
	if len(ttest.Fields) != 1 {
		t.Errorf("Expected 1 field, got %d", len(ttest.Fields))
	}

	if _, ok := ttest.Fields["field"]; !ok {
		t.Error("Field 'field' not found in merged class")
	}

	if len(ttest.Methods) != 1 {
		t.Errorf("Expected 1 method, got %d", len(ttest.Methods))
	}

	if _, ok := ttest.Methods["printme"]; !ok {
		t.Error("Method 'printme' not found in merged class")
	}

	// Verify IsPartial is still true
	if !ttest.IsPartial {
		t.Error("Class should still be marked as partial")
	}
}
