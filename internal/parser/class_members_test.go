package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Operator Declaration Tests
// ============================================================================

func TestParseClassOperatorDeclarations(t *testing.T) {
	input := `
type TMyRange = class
   class operator += String uses AppendString;
   class operator IN array of Integer uses ContainsArray;
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ClassDecl. got=%T",
			program.Statements[0])
	}

	if len(stmt.Operators) != 2 {
		t.Fatalf("stmt.Operators should contain 2 operators. got=%d", len(stmt.Operators))
	}

	first := stmt.Operators[0]
	if first.Kind != ast.OperatorKindClass {
		t.Fatalf("first operator kind expected OperatorKindClass, got %s", first.Kind)
	}
	if first.OperatorSymbol != "+=" {
		t.Fatalf("first operator symbol expected '+='; got %q", first.OperatorSymbol)
	}
	if first.Arity != 1 {
		t.Fatalf("first operator arity expected 1; got %d", first.Arity)
	}
	if len(first.OperandTypes) != 1 || first.OperandTypes[0].String() != "String" {
		t.Fatalf("first operator operand expected 'String'; got %v", first.OperandTypes)
	}
	if first.Binding == nil || first.Binding.Value != "AppendString" {
		t.Fatalf("first operator binding expected 'AppendString'; got %v", first.Binding)
	}

	second := stmt.Operators[1]
	if second.OperatorSymbol != "in" {
		t.Fatalf("second operator symbol expected 'in'; got %q", second.OperatorSymbol)
	}
	if len(second.OperandTypes) != 1 || second.OperandTypes[0].String() != "array of Integer" {
		t.Fatalf("second operator operand expected 'array of Integer'; got %v", second.OperandTypes)
	}
	if second.Binding == nil || second.Binding.Value != "ContainsArray" {
		t.Fatalf("second operator binding expected 'ContainsArray'; got %v", second.Binding)
	}
}

// ============================================================================
// Field Initializer Tests
// ============================================================================

func TestClassFieldInitializerWithEquals(t *testing.T) {
	input := `
type TTest = class
	Field: String = 'hello';
	Count: Integer = 42;
	Ratio: Float = 3.14;
	Active: Boolean = true;
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ClassDecl. got=%T",
			program.Statements[0])
	}

	if len(stmt.Fields) != 4 {
		t.Fatalf("stmt.Fields should contain 4 fields. got=%d", len(stmt.Fields))
	}

	// Test Field: String = 'hello'
	if stmt.Fields[0].Name.Value != "Field" {
		t.Errorf("Field[0].Name.Value not 'Field'. got=%s", stmt.Fields[0].Name.Value)
	}
	if stmt.Fields[0].InitValue == nil {
		t.Errorf("Field[0].InitValue should not be nil")
	} else {
		strLit, ok := stmt.Fields[0].InitValue.(*ast.StringLiteral)
		if !ok {
			t.Errorf("Field[0].InitValue is not *ast.StringLiteral. got=%T", stmt.Fields[0].InitValue)
		} else if strLit.Value != "hello" {
			t.Errorf("Field[0].InitValue not 'hello'. got=%s", strLit.Value)
		}
	}

	// Test Count: Integer = 42
	if stmt.Fields[1].Name.Value != "Count" {
		t.Errorf("Field[1].Name.Value not 'Count'. got=%s", stmt.Fields[1].Name.Value)
	}
	if stmt.Fields[1].InitValue == nil {
		t.Errorf("Field[1].InitValue should not be nil")
	} else {
		intLit, ok := stmt.Fields[1].InitValue.(*ast.IntegerLiteral)
		if !ok {
			t.Errorf("Field[1].InitValue is not *ast.IntegerLiteral. got=%T", stmt.Fields[1].InitValue)
		} else if intLit.Value != 42 {
			t.Errorf("Field[1].InitValue not 42. got=%d", intLit.Value)
		}
	}

	// Test Ratio: Float = 3.14
	if stmt.Fields[2].Name.Value != "Ratio" {
		t.Errorf("Field[2].Name.Value not 'Ratio'. got=%s", stmt.Fields[2].Name.Value)
	}
	if stmt.Fields[2].InitValue == nil {
		t.Errorf("Field[2].InitValue should not be nil")
	} else {
		floatLit, ok := stmt.Fields[2].InitValue.(*ast.FloatLiteral)
		if !ok {
			t.Errorf("Field[2].InitValue is not *ast.FloatLiteral. got=%T", stmt.Fields[2].InitValue)
		} else if floatLit.Value != 3.14 {
			t.Errorf("Field[2].InitValue not 3.14. got=%f", floatLit.Value)
		}
	}

	// Test Active: Boolean = true
	if stmt.Fields[3].Name.Value != "Active" {
		t.Errorf("Field[3].Name.Value not 'Active'. got=%s", stmt.Fields[3].Name.Value)
	}
	if stmt.Fields[3].InitValue == nil {
		t.Errorf("Field[3].InitValue should not be nil")
	} else {
		boolLit, ok := stmt.Fields[3].InitValue.(*ast.BooleanLiteral)
		if !ok {
			t.Errorf("Field[3].InitValue is not *ast.BooleanLiteral. got=%T", stmt.Fields[3].InitValue)
		} else if boolLit.Value != true {
			t.Errorf("Field[3].InitValue not true. got=%t", boolLit.Value)
		}
	}
}

func TestClassFieldInitializerWithAssign(t *testing.T) {
	input := `
type TTest = class
	class var Count: Integer := 100;
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ClassDecl. got=%T",
			program.Statements[0])
	}

	if len(stmt.Fields) != 1 {
		t.Fatalf("stmt.Fields should contain 1 field. got=%d", len(stmt.Fields))
	}

	if !stmt.Fields[0].IsClassVar {
		t.Errorf("Field should be marked as class var")
	}

	if stmt.Fields[0].InitValue == nil {
		t.Errorf("Field InitValue should not be nil")
	} else {
		intLit, ok := stmt.Fields[0].InitValue.(*ast.IntegerLiteral)
		if !ok {
			t.Errorf("Field InitValue is not *ast.IntegerLiteral. got=%T", stmt.Fields[0].InitValue)
		} else if intLit.Value != 100 {
			t.Errorf("Field InitValue not 100. got=%d", intLit.Value)
		}
	}
}

func TestFieldInitializerWithExpression(t *testing.T) {
	input := `
type TTest = class
	Sum: Integer = 10 + 20;
	Product: Integer = 5 * 6;
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ClassDecl. got=%T",
			program.Statements[0])
	}

	if len(stmt.Fields) != 2 {
		t.Fatalf("stmt.Fields should contain 2 fields. got=%d", len(stmt.Fields))
	}

	// Test Sum: Integer = 10 + 20
	if stmt.Fields[0].InitValue == nil {
		t.Errorf("Field[0].InitValue should not be nil")
	} else {
		_, ok := stmt.Fields[0].InitValue.(*ast.BinaryExpression)
		if !ok {
			t.Errorf("Field[0].InitValue is not *ast.BinaryExpression. got=%T", stmt.Fields[0].InitValue)
		}
	}

	// Test Product: Integer = 5 * 6
	if stmt.Fields[1].InitValue == nil {
		t.Errorf("Field[1].InitValue should not be nil")
	} else {
		_, ok := stmt.Fields[1].InitValue.(*ast.BinaryExpression)
		if !ok {
			t.Errorf("Field[1].InitValue is not *ast.BinaryExpression. got=%T", stmt.Fields[1].InitValue)
		}
	}
}

func TestFieldWithoutInitializer(t *testing.T) {
	input := `
type TTest = class
	Field: String;
	Count: Integer;
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ClassDecl. got=%T",
			program.Statements[0])
	}

	if len(stmt.Fields) != 2 {
		t.Fatalf("stmt.Fields should contain 2 fields. got=%d", len(stmt.Fields))
	}

	// Both fields should have nil InitValue
	if stmt.Fields[0].InitValue != nil {
		t.Errorf("Field[0].InitValue should be nil. got=%v", stmt.Fields[0].InitValue)
	}

	if stmt.Fields[1].InitValue != nil {
		t.Errorf("Field[1].InitValue should be nil. got=%v", stmt.Fields[1].InitValue)
	}
}

// ============================================================================
// Class Invariant Tests
// ============================================================================

func TestParseClassInvariants(t *testing.T) {
	input := `
type TPoint = class
   FX, FY: Float;
   invariants
      FX >= 0.0;
      FY >= 0.0 : 'Y coordinate must be non-negative';
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ClassDecl. got=%T",
			program.Statements[0])
	}

	if stmt.Invariants == nil {
		t.Fatalf("stmt.Invariants should not be nil")
	}

	if len(stmt.Invariants.Conditions) != 2 {
		t.Fatalf("stmt.Invariants.Conditions should contain 2 conditions. got=%d",
			len(stmt.Invariants.Conditions))
	}

	// Check first condition (no message)
	firstCond := stmt.Invariants.Conditions[0]
	if firstCond.Test == nil {
		t.Fatalf("first condition Test should not be nil")
	}
	if firstCond.Message != nil {
		t.Errorf("first condition Message should be nil")
	}

	// Check second condition (with message)
	secondCond := stmt.Invariants.Conditions[1]
	if secondCond.Test == nil {
		t.Fatalf("second condition Test should not be nil")
	}
	if secondCond.Message == nil {
		t.Fatalf("second condition Message should not be nil")
	}
}

func TestParseClassWithMultipleInvariants(t *testing.T) {
	input := `
type TStack = class
private
   FCount: Integer;
   FCapacity: Integer;
   invariants
      FCount >= 0;
      FCount <= FCapacity;
      FCapacity > 0;
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ClassDecl. got=%T",
			program.Statements[0])
	}

	if stmt.Invariants == nil {
		t.Fatalf("stmt.Invariants should not be nil")
	}

	if len(stmt.Invariants.Conditions) != 3 {
		t.Fatalf("stmt.Invariants.Conditions should contain 3 conditions. got=%d",
			len(stmt.Invariants.Conditions))
	}
}

func TestParseClassWithoutInvariants(t *testing.T) {
	input := `
type TSimple = class
   FValue: Integer;
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ClassDecl. got=%T",
			program.Statements[0])
	}

	if stmt.Invariants != nil {
		t.Errorf("stmt.Invariants should be nil for class without invariants. got=%v",
			stmt.Invariants)
	}
}
