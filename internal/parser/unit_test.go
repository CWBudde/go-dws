package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

func TestMinimalUnit(t *testing.T) {
	input := `unit MyUnit;
end.`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program should have 1 statement, got=%d", len(program.Statements))
	}

	unit, ok := program.Statements[0].(*ast.UnitDeclaration)
	if !ok {
		t.Fatalf("statement is not *ast.UnitDeclaration, got=%T", program.Statements[0])
	}

	if unit.Name.Value != "MyUnit" {
		t.Errorf("unit name wrong. expected=MyUnit, got=%s", unit.Name.Value)
	}
}

func TestUnitWithInterface(t *testing.T) {
	input := `unit TestUnit;
interface
end.`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program should have 1 statement, got=%d", len(program.Statements))
	}

	unit, ok := program.Statements[0].(*ast.UnitDeclaration)
	if !ok {
		t.Fatalf("statement is not *ast.UnitDeclaration, got=%T", program.Statements[0])
	}

	if unit.InterfaceSection == nil {
		t.Fatal("unit should have interface section")
	}
}

func TestUnitWithImplementation(t *testing.T) {
	input := `unit TestUnit;
implementation
end.`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	unit, ok := program.Statements[0].(*ast.UnitDeclaration)
	if !ok {
		t.Fatalf("statement is not *ast.UnitDeclaration, got=%T", program.Statements[0])
	}

	if unit.ImplementationSection == nil {
		t.Fatal("unit should have implementation section")
	}
}

func TestUnitWithUsesClause(t *testing.T) {
	input := `unit MyUnit;
interface
uses System, Math;
end.`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	unit, ok := program.Statements[0].(*ast.UnitDeclaration)
	if !ok {
		t.Fatalf("statement is not *ast.UnitDeclaration, got=%T", program.Statements[0])
	}

	if unit.InterfaceSection == nil {
		t.Fatal("unit should have interface section")
	}

	if len(unit.InterfaceSection.Statements) == 0 {
		t.Fatal("interface section should have statements")
	}

	usesClause, ok := unit.InterfaceSection.Statements[0].(*ast.UsesClause)
	if !ok {
		t.Fatalf("first statement should be UsesClause, got=%T", unit.InterfaceSection.Statements[0])
	}

	if len(usesClause.Units) != 2 {
		t.Fatalf("expected 2 units in uses clause, got=%d", len(usesClause.Units))
	}

	if usesClause.Units[0].Value != "System" {
		t.Errorf("first unit should be System, got=%s", usesClause.Units[0].Value)
	}

	if usesClause.Units[1].Value != "Math" {
		t.Errorf("second unit should be Math, got=%s", usesClause.Units[1].Value)
	}
}

func TestUnitWithInterfaceAndImplementation(t *testing.T) {
	input := `unit CompleteUnit;
interface
uses System;
implementation
uses SysUtils;
end.`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	unit, ok := program.Statements[0].(*ast.UnitDeclaration)
	if !ok {
		t.Fatalf("statement is not *ast.UnitDeclaration, got=%T", program.Statements[0])
	}

	if unit.InterfaceSection == nil {
		t.Fatal("unit should have interface section")
	}

	if unit.ImplementationSection == nil {
		t.Fatal("unit should have implementation section")
	}

	// Check interface uses clause
	if len(unit.InterfaceSection.Statements) == 0 {
		t.Fatal("interface section should have statements")
	}

	interfaceUses, ok := unit.InterfaceSection.Statements[0].(*ast.UsesClause)
	if !ok {
		t.Fatalf("first interface statement should be UsesClause, got=%T", unit.InterfaceSection.Statements[0])
	}

	if interfaceUses.Units[0].Value != "System" {
		t.Errorf("interface uses should include System, got=%s", interfaceUses.Units[0].Value)
	}

	// Check implementation uses clause
	if len(unit.ImplementationSection.Statements) == 0 {
		t.Fatal("implementation section should have statements")
	}

	implUses, ok := unit.ImplementationSection.Statements[0].(*ast.UsesClause)
	if !ok {
		t.Fatalf("first implementation statement should be UsesClause, got=%T", unit.ImplementationSection.Statements[0])
	}

	if implUses.Units[0].Value != "SysUtils" {
		t.Errorf("implementation uses should include SysUtils, got=%s", implUses.Units[0].Value)
	}
}

func TestUnitWithInitialization(t *testing.T) {
	input := `unit InitUnit;
interface
implementation
initialization
  x := 42;
end.`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	unit, ok := program.Statements[0].(*ast.UnitDeclaration)
	if !ok {
		t.Fatalf("statement is not *ast.UnitDeclaration, got=%T", program.Statements[0])
	}

	if unit.InitSection == nil {
		t.Fatal("unit should have initialization section")
	}

	if len(unit.InitSection.Statements) == 0 {
		t.Fatal("initialization section should have statements")
	}
}

func TestUnitWithFinalization(t *testing.T) {
	input := `unit FinalUnit;
interface
implementation
finalization
  Cleanup();
end.`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	unit, ok := program.Statements[0].(*ast.UnitDeclaration)
	if !ok {
		t.Fatalf("statement is not *ast.UnitDeclaration, got=%T", program.Statements[0])
	}

	if unit.FinalSection == nil {
		t.Fatal("unit should have finalization section")
	}

	if len(unit.FinalSection.Statements) == 0 {
		t.Fatal("finalization section should have statements")
	}
}

func TestCompleteUnit(t *testing.T) {
	input := `unit FullUnit;
interface
uses System, Math;

function Add(x, y: Integer): Integer;

implementation

function Add(x, y: Integer): Integer;
begin
  Result := x + y;
end;

initialization
  x := 0;

finalization
  Cleanup();

end.`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program should have 1 statement, got=%d", len(program.Statements))
	}

	unit, ok := program.Statements[0].(*ast.UnitDeclaration)
	if !ok {
		t.Fatalf("statement is not *ast.UnitDeclaration, got=%T", program.Statements[0])
	}

	if unit.Name.Value != "FullUnit" {
		t.Errorf("unit name should be FullUnit, got=%s", unit.Name.Value)
	}

	if unit.InterfaceSection == nil {
		t.Fatal("unit should have interface section")
	}

	if unit.ImplementationSection == nil {
		t.Fatal("unit should have implementation section")
	}

	if unit.InitSection == nil {
		t.Fatal("unit should have initialization section")
	}

	if unit.FinalSection == nil {
		t.Fatal("unit should have finalization section")
	}

	// Verify uses clause in interface
	if len(unit.InterfaceSection.Statements) < 1 {
		t.Fatal("interface section should have at least uses clause")
	}

	usesClause, ok := unit.InterfaceSection.Statements[0].(*ast.UsesClause)
	if !ok {
		t.Fatalf("first interface statement should be UsesClause, got=%T", unit.InterfaceSection.Statements[0])
	}

	if len(usesClause.Units) != 2 {
		t.Errorf("expected 2 units in uses clause, got=%d", len(usesClause.Units))
	}
}

func TestUnitWithInterfaceDeclarations(t *testing.T) {
	input := `unit DeclUnit;
interface
uses System;

type
  TMyClass = class
  end;

var
  GlobalVar: Integer;

function MyFunc(): String;

end.`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	unit, ok := program.Statements[0].(*ast.UnitDeclaration)
	if !ok {
		t.Fatalf("statement is not *ast.UnitDeclaration, got=%T", program.Statements[0])
	}

	if unit.InterfaceSection == nil {
		t.Fatal("unit should have interface section")
	}

	// Should have uses clause + type declaration + var declaration + function declaration
	if len(unit.InterfaceSection.Statements) < 4 {
		t.Errorf("expected at least 4 statements in interface, got=%d", len(unit.InterfaceSection.Statements))
	}
}

func TestMultipleUnitsInUsesClause(t *testing.T) {
	input := `unit TestUnit;
interface
uses System, SysUtils, Math, Graphics, Controls;
end.`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	unit, ok := program.Statements[0].(*ast.UnitDeclaration)
	if !ok {
		t.Fatalf("statement is not *ast.UnitDeclaration, got=%T", program.Statements[0])
	}

	usesClause, ok := unit.InterfaceSection.Statements[0].(*ast.UsesClause)
	if !ok {
		t.Fatalf("first statement should be UsesClause, got=%T", unit.InterfaceSection.Statements[0])
	}

	expected := []string{"System", "SysUtils", "Math", "Graphics", "Controls"}
	if len(usesClause.Units) != len(expected) {
		t.Fatalf("expected %d units, got=%d", len(expected), len(usesClause.Units))
	}

	for i, expectedName := range expected {
		if usesClause.Units[i].Value != expectedName {
			t.Errorf("unit[%d] should be %s, got=%s", i, expectedName, usesClause.Units[i].Value)
		}
	}
}

func TestUnitVsProgram(t *testing.T) {
	t.Run("Unit starts with unit keyword", func(t *testing.T) {
		input := `unit MyUnit;
end.`
		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("expected 1 statement, got=%d", len(program.Statements))
		}

		_, ok := program.Statements[0].(*ast.UnitDeclaration)
		if !ok {
			t.Fatalf("should parse as UnitDeclaration, got=%T", program.Statements[0])
		}
	})

	t.Run("Program doesn't start with unit", func(t *testing.T) {
		input := `var x: Integer;
x := 42;`
		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		// Should parse as regular program statements
		if len(program.Statements) < 2 {
			t.Fatal("expected multiple program statements")
		}

		// First statement should not be a UnitDeclaration
		_, isUnit := program.Statements[0].(*ast.UnitDeclaration)
		if isUnit {
			t.Fatal("regular program should not parse as unit")
		}
	})
}

func TestUnitErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Missing unit name",
			input: `unit;`,
		},
		{
			name:  "Missing semicolon after unit name",
			input: `unit MyUnit`,
		},
		{
			name:  "Missing end",
			input: `unit MyUnit; interface`,
		},
		{
			name:  "Missing dot after end",
			input: `unit MyUnit; end`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			p.ParseProgram()

			if len(p.Errors()) == 0 {
				t.Error("expected parser errors, got none")
			}
		})
	}
}
