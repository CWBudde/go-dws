package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

func TestNestedClassTypeResolution(t *testing.T) {
	input := `
type TOuter = class
   type TInner = class
      Value: Integer;
   end;
   Field: TInner;
end;

type TInner = class
   Dummy: Integer;
end;`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	analyzer.SetSource(input, "nested_classes_test")

	if err := analyzer.Analyze(program); err != nil {
		t.Fatalf("unexpected semantic errors: %v", err)
	}

	outer := analyzer.getClassType("TOuter")
	if outer == nil {
		t.Fatalf("outer class not registered")
	}

	nested := analyzer.getClassType("TOuter.TInner")
	if nested == nil {
		t.Fatalf("nested class not registered with qualified name")
	}

	if !analyzer.typeRegistry.Has("TInner") {
		t.Fatalf("global TInner should still be registered separately")
	}

	fieldType, ok := outer.Fields[ident.Normalize("Field")]
	if !ok {
		t.Fatalf("field 'Field' not found on outer class")
	}

	nestedFieldType, ok := fieldType.(*types.ClassType)
	if !ok {
		t.Fatalf("expected field type to be ClassType, got %T", fieldType)
	}

	if nestedFieldType.Name != "TOuter.TInner" {
		t.Fatalf("expected nested field to use inner class, got %s", nestedFieldType.Name)
	}
}

func TestNestedClassUsageMatchesFixture(t *testing.T) {
	input := `
type TOuter = class
   private
      type TInner = class
         Field : Integer;
         procedure PrintMe;
      end;

      FInner : TInner;

      procedure PrintInner(obj : TInner);

   public
      constructor Create;

      procedure PrintIt;
end;


constructor TOuter.Create;
begin
   FInner := TInner.Create;
   FInner.Field := 12345;
end;

procedure TOuter.PrintIt;
begin
   PrintInner(FInner);
end;

procedure TOuter.TInner.PrintMe;
begin
   // no-op
end;

procedure TOuter.PrintInner(obj : TInner);
begin
   obj.PrintMe;
end;`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	classDecl, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("expected class declaration, got %T", program.Statements[0])
	}
	if len(classDecl.NestedTypes) == 0 {
		t.Fatalf("nested types not captured in class declaration")
	}

	analyzer := NewAnalyzer()
	analyzer.SetSource(input, "nested_fixture_like")

	if err := analyzer.Analyze(program); err != nil {
		t.Fatalf("unexpected semantic errors: %v", err)
	}
}
