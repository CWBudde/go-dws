package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func parseImplicitUnitTest(t *testing.T, source string) *ast.UnitDeclaration {
	t.Helper()
	p := parser.New(lexer.New(source))
	program := p.ParseProgram()
	if errors := p.Errors(); len(errors) > 0 {
		t.Fatalf("parse unit: %v", errors)
	}
	if len(program.Statements) != 1 {
		t.Fatalf("expected one unit, got %d statements", len(program.Statements))
	}
	unit, ok := program.Statements[0].(*ast.UnitDeclaration)
	if !ok {
		t.Fatalf("expected unit, got %T", program.Statements[0])
	}
	return unit
}

func TestAnalyzeUnit_ImplicitSectionExportsAndChecksBodies(t *testing.T) {
	unit := parseImplicitUnitTest(t, `unit ImplicitUnit;
type TCount = Integer;
var Count: TCount;
function First: Integer;
begin Result := Second(); end;
function Second: Integer;
begin Result := Count; end;
`)
	analyzer := NewAnalyzer()
	if err := analyzer.AnalyzeUnit(unit); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Count", "First", "Second"} {
		if _, err := analyzer.ResolveQualifiedSymbol("ImplicitUnit", name); err != nil {
			t.Errorf("implicit declaration must be public: %v", err)
		}
	}
	if exported := analyzer.GetUnitSymbols("ImplicitUnit"); len(exported.exportedTypes) == 0 {
		t.Fatal("implicit unit type declaration was not exported")
	}
	invalid := parseImplicitUnitTest(t, `unit Invalid;
function Fail: Integer;
begin Result := MissingValue; end;
`)
	if err := NewAnalyzer().AnalyzeUnit(invalid); err == nil {
		t.Fatal("implicit function body with unknown identifier must fail analysis")
	}
}

func TestAnalyzeUnit_ImplicitUsesDoesNotReexportDependencies(t *testing.T) {
	dependency := NewAnalyzer()
	if err := dependency.AnalyzeUnit(parseImplicitUnitTest(t, `unit Dependency;
function Imported: Integer;
begin Result := 42; end;
`)); err != nil {
		t.Fatal(err)
	}
	analyzer := NewAnalyzer()
	unit := parseImplicitUnitTest(t, `unit Consumer;
uses Dependency;
function Own: Integer;
begin Result := Imported(); end;
`)
	if err := analyzer.AnalyzeUnitWithDependencies(unit, map[string]*SymbolTable{
		"Dependency": dependency.GetUnitSymbols("Dependency"),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := analyzer.ResolveQualifiedSymbol("Consumer", "Own"); err != nil {
		t.Fatal(err)
	}
	if _, err := analyzer.ResolveQualifiedSymbol("Consumer", "Imported"); err == nil {
		t.Fatal("imported function was reexported")
	}
}

func TestAnalyzeUnit_ExplicitImplementationStaysPrivate(t *testing.T) {
	for _, source := range []string{
		`unit ExplicitUnit;
interface
function PublicValue: Integer;
implementation
function PublicValue: Integer;
begin Result := Hidden(); end;
function Hidden: Integer;
begin Result := 42; end;
end.`,
		`unit ExplicitUnit;
implementation
function Hidden: Integer;
begin Result := 42; end;
end.`,
	} {
		analyzer := NewAnalyzer()
		if err := analyzer.AnalyzeUnit(parseImplicitUnitTest(t, source)); err != nil {
			t.Fatal(err)
		}
		if _, err := analyzer.ResolveQualifiedSymbol("ExplicitUnit", "Hidden"); err == nil {
			t.Fatal("explicit implementation declaration was exported")
		}
	}
}
