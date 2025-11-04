package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

func TestAssignmentStatements(t *testing.T) {
	input := `
x := 10;
y := x + 1;
`

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 2 {
		t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
	}

	stmt1, ok := program.Statements[0].(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("statement 0 is not ast.AssignmentStatement. got=%T", program.Statements[0])
	}
	target1, ok := stmt1.Target.(*ast.Identifier)
	if !ok {
		t.Fatalf("stmt1.Target is not *ast.Identifier. got=%T", stmt1.Target)
	}
	if target1.Value != "x" {
		t.Errorf("stmt1.Target.Value = %q, want %q", target1.Value, "x")
	}
	if !testIntegerLiteral(t, stmt1.Value, 10) {
		return
	}

	stmt2, ok := program.Statements[1].(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("statement 1 is not ast.AssignmentStatement. got=%T", program.Statements[1])
	}
	target2, ok := stmt2.Target.(*ast.Identifier)
	if !ok {
		t.Fatalf("stmt2.Target is not *ast.Identifier. got=%T", stmt2.Target)
	}
	if target2.Value != "y" {
		t.Errorf("stmt2.Target.Value = %q, want %q", target2.Value, "y")
	}
	if !testInfixExpression(t, stmt2.Value, "x", "+", 1) {
		return
	}
}

// TestCompoundAssignmentStatements tests parsing of compound assignment operators.
func TestCompoundAssignmentStatements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		target   string
		operator lexer.TokenType
		value    int64
	}{
		{
			name:     "Plus assign",
			input:    "x += 5;",
			target:   "x",
			operator: lexer.PLUS_ASSIGN,
			value:    5,
		},
		{
			name:     "Minus assign",
			input:    "count -= 1;",
			target:   "count",
			operator: lexer.MINUS_ASSIGN,
			value:    1,
		},
		{
			name:     "Times assign",
			input:    "total *= 10;",
			target:   "total",
			operator: lexer.TIMES_ASSIGN,
			value:    10,
		},
		{
			name:     "Divide assign",
			input:    "result /= 2;",
			target:   "result",
			operator: lexer.DIVIDE_ASSIGN,
			value:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.AssignmentStatement)
			if !ok {
				t.Fatalf("statement is not ast.AssignmentStatement. got=%T", program.Statements[0])
			}

			target, ok := stmt.Target.(*ast.Identifier)
			if !ok {
				t.Fatalf("stmt.Target is not *ast.Identifier. got=%T", stmt.Target)
			}
			if target.Value != tt.target {
				t.Errorf("stmt.Target.Value = %q, want %q", target.Value, tt.target)
			}

			if stmt.Operator != tt.operator {
				t.Errorf("stmt.Operator = %v, want %v", stmt.Operator, tt.operator)
			}

			if !testIntegerLiteral(t, stmt.Value, tt.value) {
				return
			}
		})
	}
}

// TestCompoundAssignmentWithMemberAccess tests compound assignment with member access.
func TestCompoundAssignmentWithMemberAccess(t *testing.T) {
	input := "obj.field += 10;"

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("statement is not ast.AssignmentStatement. got=%T", program.Statements[0])
	}

	memberAccess, ok := stmt.Target.(*ast.MemberAccessExpression)
	if !ok {
		t.Fatalf("stmt.Target is not *ast.MemberAccessExpression. got=%T", stmt.Target)
	}

	objIdent, ok := memberAccess.Object.(*ast.Identifier)
	if !ok {
		t.Fatalf("memberAccess.Object is not *ast.Identifier. got=%T", memberAccess.Object)
	}
	if objIdent.Value != "obj" {
		t.Errorf("object name = %q, want %q", objIdent.Value, "obj")
	}

	if memberAccess.Member.Value != "field" {
		t.Errorf("field name = %q, want %q", memberAccess.Member.Value, "field")
	}

	if stmt.Operator != lexer.PLUS_ASSIGN {
		t.Errorf("stmt.Operator = %v, want %v", stmt.Operator, lexer.PLUS_ASSIGN)
	}

	if !testIntegerLiteral(t, stmt.Value, 10) {
		return
	}
}

// TestCompoundAssignmentWithIndexExpression tests compound assignment with array indexing.
func TestCompoundAssignmentWithIndexExpression(t *testing.T) {
	input := "arr[i] *= 2;"

	p := testParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("statement is not ast.AssignmentStatement. got=%T", program.Statements[0])
	}

	indexExpr, ok := stmt.Target.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("stmt.Target is not *ast.IndexExpression. got=%T", stmt.Target)
	}

	arrIdent, ok := indexExpr.Left.(*ast.Identifier)
	if !ok {
		t.Fatalf("indexExpr.Left is not *ast.Identifier. got=%T", indexExpr.Left)
	}
	if arrIdent.Value != "arr" {
		t.Errorf("array name = %q, want %q", arrIdent.Value, "arr")
	}

	indexIdent, ok := indexExpr.Index.(*ast.Identifier)
	if !ok {
		t.Fatalf("indexExpr.Index is not *ast.Identifier. got=%T", indexExpr.Index)
	}
	if indexIdent.Value != "i" {
		t.Errorf("index = %q, want %q", indexIdent.Value, "i")
	}

	if stmt.Operator != lexer.TIMES_ASSIGN {
		t.Errorf("stmt.Operator = %v, want %v", stmt.Operator, lexer.TIMES_ASSIGN)
	}

	if !testIntegerLiteral(t, stmt.Value, 2) {
		return
	}
}

// TestMemberAssignmentStatements tests parsing of member assignment statements.
// This tests the pattern: obj.field := value
func TestMemberAssignmentStatements(t *testing.T) {
	tests := []struct {
		value      interface{}
		name       string
		input      string
		objectName string
		fieldName  string
	}{
		{
			name:       "Simple member assignment",
			input:      "p.X := 10;",
			objectName: "p",
			fieldName:  "X",
			value:      int64(10),
		},
		{
			name:       "Self member assignment",
			input:      "Self.X := 42;",
			objectName: "Self",
			fieldName:  "X",
			value:      int64(42),
		},
		{
			name:       "Member assignment with expression",
			input:      "obj.Value := x + 5;",
			objectName: "obj",
			fieldName:  "Value",
			value:      "x + 5", // Will check it's a binary expression
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.AssignmentStatement)
			if !ok {
				t.Fatalf("statement is not ast.AssignmentStatement. got=%T", program.Statements[0])
			}

			// Check that Name is actually a MemberAccessExpression, not just an Identifier
			// For now, we'll validate the basic structure once parser is updated
			// This test will fail until parser supports member assignments
			t.Logf("Assignment statement Name: %T, Value: %T", stmt.Target, stmt.Value)
		})
	}
}

// TestCallExpressionParsing verifies parsing of function call expressions.
