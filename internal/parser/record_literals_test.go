package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Parser Tests for Record Literals with Named Fields
// ============================================================================

func TestParseRecordLiteral_Anonymous(t *testing.T) {
	// Test anonymous record: (x: 10; y: 20)
	input := `
		var p := (x: 10; y: 20);
	`

	program := testParse(t, input)
	if len(program.Statements) != 1 {
		t.Fatalf("program should have 1 statement, got %d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.VarDeclStatement)
	if !ok {
		t.Fatalf("statement is not *ast.VarDeclStatement, got %T", program.Statements[0])
	}

	recordLit, ok := stmt.Value.(*ast.RecordLiteralExpression)
	if !ok {
		t.Fatalf("value is not *ast.RecordLiteralExpression, got %T", stmt.Value)
	}

	// Check TypeName is nil (anonymous record)
	if recordLit.TypeName != nil {
		t.Errorf("TypeName should be nil for anonymous record, got %v", recordLit.TypeName)
	}

	// Check fields
	if len(recordLit.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(recordLit.Fields))
	}

	// Check first field
	if recordLit.Fields[0].Name.Value != "x" {
		t.Errorf("Fields[0].Name.Value = %v, want 'x'", recordLit.Fields[0].Name.Value)
	}
	if intLit, ok := recordLit.Fields[0].Value.(*ast.IntegerLiteral); !ok {
		t.Errorf("Fields[0].Value is not IntegerLiteral, got %T", recordLit.Fields[0].Value)
	} else if intLit.Value != 10 {
		t.Errorf("Fields[0].Value = %d, want 10", intLit.Value)
	}

	// Check second field
	if recordLit.Fields[1].Name.Value != "y" {
		t.Errorf("Fields[1].Name.Value = %v, want 'y'", recordLit.Fields[1].Name.Value)
	}
}

func TestParseRecordLiteral_Typed(t *testing.T) {
	// Test typed record: TPoint(x: 10; y: 20)
	input := `
		var p := TPoint(x: 10; y: 20);
	`

	program := testParse(t, input)
	if len(program.Statements) != 1 {
		t.Fatalf("program should have 1 statement, got %d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.VarDeclStatement)
	if !ok {
		t.Fatalf("statement is not *ast.VarDeclStatement, got %T", program.Statements[0])
	}

	recordLit, ok := stmt.Value.(*ast.RecordLiteralExpression)
	if !ok {
		t.Fatalf("value is not *ast.RecordLiteralExpression, got %T", stmt.Value)
	}

	// Check TypeName
	if recordLit.TypeName == nil {
		t.Fatal("TypeName should not be nil for typed record")
	}
	if recordLit.TypeName.Value != "TPoint" {
		t.Errorf("TypeName.Value = %v, want 'TPoint'", recordLit.TypeName.Value)
	}

	// Check fields
	if len(recordLit.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(recordLit.Fields))
	}
}

func TestParseRecordLiteral_WithSemicolons(t *testing.T) {
	// Test with semicolons: (a: 1; b: 2; c: 3)
	input := `
		var rec := (a: 1; b: 2; c: 3);
	`

	program := testParse(t, input)
	stmt := program.Statements[0].(*ast.VarDeclStatement)
	recordLit := stmt.Value.(*ast.RecordLiteralExpression)

	if len(recordLit.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(recordLit.Fields))
	}

	expectedNames := []string{"a", "b", "c"}
	for i, expected := range expectedNames {
		if recordLit.Fields[i].Name.Value != expected {
			t.Errorf("Fields[%d].Name.Value = %v, want %v", i, recordLit.Fields[i].Name.Value, expected)
		}
	}
}

func TestParseRecordLiteral_WithCommas(t *testing.T) {
	// Test with commas: (a: 1, b: 2, c: 3)
	input := `
		var rec := (a: 1, b: 2, c: 3);
	`

	program := testParse(t, input)
	stmt := program.Statements[0].(*ast.VarDeclStatement)
	recordLit := stmt.Value.(*ast.RecordLiteralExpression)

	if len(recordLit.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(recordLit.Fields))
	}

	expectedNames := []string{"a", "b", "c"}
	for i, expected := range expectedNames {
		if recordLit.Fields[i].Name.Value != expected {
			t.Errorf("Fields[%d].Name.Value = %v, want %v", i, recordLit.Fields[i].Name.Value, expected)
		}
	}
}

func TestParseRecordLiteral_Nested(t *testing.T) {
	// Test nested: TRect(TopLeft: (x: 0; y: 0); BottomRight: (x: 10; y: 10))
	input := `
		var rect := TRect(TopLeft: (x: 0; y: 0); BottomRight: (x: 10; y: 10));
	`

	program := testParse(t, input)
	stmt := program.Statements[0].(*ast.VarDeclStatement)
	recordLit := stmt.Value.(*ast.RecordLiteralExpression)

	if recordLit.TypeName.Value != "TRect" {
		t.Errorf("TypeName.Value = %v, want 'TRect'", recordLit.TypeName.Value)
	}

	if len(recordLit.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(recordLit.Fields))
	}

	// Check first field is a nested record
	topLeft, ok := recordLit.Fields[0].Value.(*ast.RecordLiteralExpression)
	if !ok {
		t.Fatalf("TopLeft value is not *ast.RecordLiteralExpression, got %T", recordLit.Fields[0].Value)
	}
	if len(topLeft.Fields) != 2 {
		t.Errorf("TopLeft should have 2 fields, got %d", len(topLeft.Fields))
	}

	// Check second field is a nested record
	bottomRight, ok := recordLit.Fields[1].Value.(*ast.RecordLiteralExpression)
	if !ok {
		t.Fatalf("BottomRight value is not *ast.RecordLiteralExpression, got %T", recordLit.Fields[1].Value)
	}
	if len(bottomRight.Fields) != 2 {
		t.Errorf("BottomRight should have 2 fields, got %d", len(bottomRight.Fields))
	}
}

func TestParseRecordLiteral_DeathStarExample(t *testing.T) {
	// Test actual Death_Star.dws example
	input := `
		const big : TSphere = (cx: 20; cy: 20; cz: 0; r: 20);
		const small : TSphere = (cx: 7; cy: 7; cz: -10; r: 15);
	`

	program := testParse(t, input)
	if len(program.Statements) != 2 {
		t.Fatalf("program should have 2 statements, got %d", len(program.Statements))
	}

	// Check first const
	stmt1, ok := program.Statements[0].(*ast.ConstDecl)
	if !ok {
		t.Fatalf("statement 0 is not *ast.ConstDecl, got %T", program.Statements[0])
	}

	recordLit1, ok := stmt1.Value.(*ast.RecordLiteralExpression)
	if !ok {
		t.Fatalf("const value is not *ast.RecordLiteralExpression, got %T", stmt1.Value)
	}

	if len(recordLit1.Fields) != 4 {
		t.Fatalf("expected 4 fields, got %d", len(recordLit1.Fields))
	}

	expectedNames := []string{"cx", "cy", "cz", "r"}
	expectedValues := []int64{20, 20, 0, 20}
	for i, expected := range expectedNames {
		if recordLit1.Fields[i].Name.Value != expected {
			t.Errorf("Fields[%d].Name.Value = %v, want %v", i, recordLit1.Fields[i].Name.Value, expected)
		}
		if intLit, ok := recordLit1.Fields[i].Value.(*ast.IntegerLiteral); ok {
			if intLit.Value != expectedValues[i] {
				t.Errorf("Fields[%d].Value = %d, want %d", i, intLit.Value, expectedValues[i])
			}
		}
	}

	// Check second const
	stmt2, ok := program.Statements[1].(*ast.ConstDecl)
	if !ok {
		t.Fatalf("statement 1 is not *ast.ConstDecl, got %T", program.Statements[1])
	}

	recordLit2, ok := stmt2.Value.(*ast.RecordLiteralExpression)
	if !ok {
		t.Fatalf("const value is not *ast.RecordLiteralExpression, got %T", stmt2.Value)
	}

	if len(recordLit2.Fields) != 4 {
		t.Fatalf("expected 4 fields, got %d", len(recordLit2.Fields))
	}
}

func TestParseRecordLiteral_ErrorMissingColon(t *testing.T) {
	// Test error: missing colon
	input := `
		var p := (x 10; y: 20);
	`

	_, errors := testParseWithErrors(t, input)
	if len(errors) == 0 {
		t.Fatal("expected parsing errors, got none")
	}
}

func TestParseRecordLiteral_ErrorMissingValue(t *testing.T) {
	// Test error: missing value after colon
	input := `
		var p := (x:; y: 20);
	`

	_, errors := testParseWithErrors(t, input)
	if len(errors) == 0 {
		t.Fatal("expected parsing errors, got none")
	}
}

func TestParseRecordLiteral_ErrorUnclosedParen(t *testing.T) {
	// Test error: unclosed parenthesis
	input := `
		var p := (x: 10; y: 20;
	`

	_, errors := testParseWithErrors(t, input)
	if len(errors) == 0 {
		t.Fatal("expected parsing errors, got none")
	}
}

func TestParseRecordLiteral_TrailingSeparator(t *testing.T) {
	// Test optional trailing separator: (x: 10; y: 20;)
	input := `
		var p := (x: 10; y: 20;);
	`

	program := testParse(t, input)
	stmt := program.Statements[0].(*ast.VarDeclStatement)
	recordLit := stmt.Value.(*ast.RecordLiteralExpression)

	if len(recordLit.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(recordLit.Fields))
	}
}

// Helper function to test parsing
func testParse(t *testing.T, input string) *ast.Program {
	t.Helper()
	program, errors := testParseWithErrors(t, input)
	if len(errors) > 0 {
		t.Fatalf("parsing errors: %v", errors)
	}
	return program
}

// Helper function to test parsing with errors
func testParseWithErrors(t *testing.T, input string) (*ast.Program, []*ParserError) {
	t.Helper()
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	return program, p.Errors()
}
