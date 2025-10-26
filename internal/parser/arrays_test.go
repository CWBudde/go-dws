package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// ============================================================================
// Array Type Declaration Parser Tests (Task 8.122)
// ============================================================================

func TestParseArrayTypeDeclaration(t *testing.T) {
	t.Run("Static array with bounds", func(t *testing.T) {
		input := `type TMyArray = array[1..10] of Integer;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		stmt := program.Statements[0]
		// For now, array type declarations might be parsed differently
		// We need to determine the AST structure for type declarations
		t.Logf("Statement type: %T", stmt)
	})

	t.Run("Dynamic array without bounds", func(t *testing.T) {
		input := `type TStringArray = array of String;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}
	})

	t.Run("Array of custom type", func(t *testing.T) {
		input := `type TPersonArray = array[0..99] of TPerson;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}
	})
}

// ============================================================================
// Array Literal Parser Tests (Task 8.123)
// ============================================================================

func TestParseArrayLiteral(t *testing.T) {
	t.Run("Array literal with integers", func(t *testing.T) {
		input := `[1, 2, 3];`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("statement is not *ast.ExpressionStatement, got %T", program.Statements[0])
		}

		// Currently this might parse as SetLiteral - we'll need to distinguish later
		// For now, just check that we can parse bracket syntax
		t.Logf("Expression type: %T", stmt.Expression)
	})

	t.Run("Empty array literal", func(t *testing.T) {
		input := `[];`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}
	})

	t.Run("Array literal with strings", func(t *testing.T) {
		input := `['hello', 'world'];`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}
	})

	t.Run("Nested array literal", func(t *testing.T) {
		input := `[[1, 2], [3, 4]];`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}
	})
}

// ============================================================================
// Array Indexing Parser Tests (Task 8.124)
// ============================================================================

func TestParseArrayIndexing(t *testing.T) {
	t.Run("Simple array indexing", func(t *testing.T) {
		input := `arr[0];`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("statement is not *ast.ExpressionStatement, got %T", program.Statements[0])
		}

		indexExpr, ok := stmt.Expression.(*ast.IndexExpression)
		if !ok {
			t.Fatalf("expression is not *ast.IndexExpression, got %T", stmt.Expression)
		}

		// Test left side (array)
		leftIdent, ok := indexExpr.Left.(*ast.Identifier)
		if !ok {
			t.Fatalf("indexExpr.Left is not *ast.Identifier, got %T", indexExpr.Left)
		}
		if leftIdent.Value != "arr" {
			t.Errorf("leftIdent.Value = %s, want 'arr'", leftIdent.Value)
		}

		// Test index
		indexLit, ok := indexExpr.Index.(*ast.IntegerLiteral)
		if !ok {
			t.Fatalf("indexExpr.Index is not *ast.IntegerLiteral, got %T", indexExpr.Index)
		}
		if indexLit.Value != 0 {
			t.Errorf("indexLit.Value = %d, want 0", indexLit.Value)
		}
	})

	t.Run("Array indexing with variable", func(t *testing.T) {
		input := `arr[i];`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		indexExpr, ok := stmt.Expression.(*ast.IndexExpression)
		if !ok {
			t.Fatalf("expression is not *ast.IndexExpression, got %T", stmt.Expression)
		}

		// Test index is identifier
		indexIdent, ok := indexExpr.Index.(*ast.Identifier)
		if !ok {
			t.Fatalf("indexExpr.Index is not *ast.Identifier, got %T", indexExpr.Index)
		}
		if indexIdent.Value != "i" {
			t.Errorf("indexIdent.Value = %s, want 'i'", indexIdent.Value)
		}
	})

	t.Run("Array indexing with expression", func(t *testing.T) {
		input := `arr[i + 1];`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		indexExpr, ok := stmt.Expression.(*ast.IndexExpression)
		if !ok {
			t.Fatalf("expression is not *ast.IndexExpression, got %T", stmt.Expression)
		}

		// Test index is binary expression
		_, ok = indexExpr.Index.(*ast.BinaryExpression)
		if !ok {
			t.Fatalf("indexExpr.Index is not *ast.BinaryExpression, got %T", indexExpr.Index)
		}
	})

	t.Run("Nested array indexing", func(t *testing.T) {
		input := `arr[i][j];`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		outerIndex, ok := stmt.Expression.(*ast.IndexExpression)
		if !ok {
			t.Fatalf("expression is not *ast.IndexExpression, got %T", stmt.Expression)
		}

		// The left side should itself be an IndexExpression
		innerIndex, ok := outerIndex.Left.(*ast.IndexExpression)
		if !ok {
			t.Fatalf("outerIndex.Left is not *ast.IndexExpression, got %T", outerIndex.Left)
		}

		// Test innermost array name
		arrIdent, ok := innerIndex.Left.(*ast.Identifier)
		if !ok {
			t.Fatalf("innerIndex.Left is not *ast.Identifier, got %T", innerIndex.Left)
		}
		if arrIdent.Value != "arr" {
			t.Errorf("arrIdent.Value = %s, want 'arr'", arrIdent.Value)
		}
	})

	t.Run("String() method for index expression", func(t *testing.T) {
		input := `arr[5];`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		indexExpr := stmt.Expression.(*ast.IndexExpression)

		str := indexExpr.String()
		expected := "(arr[5])"
		if str != expected {
			t.Errorf("String() = %s, want %s", str, expected)
		}
	})
}

// ============================================================================
// Combined Array Tests (Task 8.125)
// ============================================================================

func TestArrayDeclarationAndUsage(t *testing.T) {
	t.Run("Type declaration and array access", func(t *testing.T) {
		// Test that we can declare array types and parse array indexing expressions
		input := `
		type TMyArray = array[1..10] of Integer;
		var x: Integer;
		x := arr[5];
		`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) < 3 {
			t.Fatalf("program.Statements should contain at least 3 statements, got %d", len(program.Statements))
		}

		// Verify first statement is array type declaration
		arrayDecl, ok := program.Statements[0].(*ast.ArrayDecl)
		if !ok {
			t.Fatalf("first statement is not *ast.ArrayDecl, got %T", program.Statements[0])
		}

		if arrayDecl.Name.Value != "TMyArray" {
			t.Errorf("arrayDecl.Name.Value = %s, want 'TMyArray'", arrayDecl.Name.Value)
		}
	})

	t.Run("Reading array element", func(t *testing.T) {
		// Test array indexing in expressions
		input := `x := arr[i];`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		// This should parse as an assignment with IndexExpression on the right side
		assignStmt, ok := program.Statements[0].(*ast.AssignmentStatement)
		if !ok {
			t.Fatalf("statement is not *ast.AssignmentStatement, got %T", program.Statements[0])
		}

		// Right side should be IndexExpression
		_, ok = assignStmt.Value.(*ast.IndexExpression)
		if !ok {
			t.Fatalf("assignment value is not *ast.IndexExpression, got %T", assignStmt.Value)
		}
	})
}

// ============================================================================
// Array Assignment Parser Tests (Task 8.138)
// ============================================================================

func TestParseArrayAssignment(t *testing.T) {
	t.Run("Simple array element assignment", func(t *testing.T) {
		input := `
		begin
			arr[0] := 42;
		end
		`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		block, ok := program.Statements[0].(*ast.BlockStatement)
		if !ok {
			t.Fatalf("statement is not *ast.BlockStatement, got %T", program.Statements[0])
		}

		if len(block.Statements) != 1 {
			t.Fatalf("block.Statements should contain 1 statement, got %d", len(block.Statements))
		}

		assignStmt, ok := block.Statements[0].(*ast.AssignmentStatement)
		if !ok {
			t.Fatalf("statement is not *ast.AssignmentStatement, got %T", block.Statements[0])
		}

		// Verify Target is an IndexExpression
		indexExpr, ok := assignStmt.Target.(*ast.IndexExpression)
		if !ok {
			t.Fatalf("assignStmt.Target is not *ast.IndexExpression, got %T", assignStmt.Target)
		}

		// Verify array identifier
		arrIdent, ok := indexExpr.Left.(*ast.Identifier)
		if !ok {
			t.Fatalf("indexExpr.Left is not *ast.Identifier, got %T", indexExpr.Left)
		}
		if arrIdent.Value != "arr" {
			t.Errorf("array name = %q, want %q", arrIdent.Value, "arr")
		}

		// Verify index is integer literal
		indexLit, ok := indexExpr.Index.(*ast.IntegerLiteral)
		if !ok {
			t.Fatalf("indexExpr.Index is not *ast.IntegerLiteral, got %T", indexExpr.Index)
		}
		if indexLit.Value != 0 {
			t.Errorf("index value = %d, want %d", indexLit.Value, 0)
		}

		// Verify assigned value
		valueLit, ok := assignStmt.Value.(*ast.IntegerLiteral)
		if !ok {
			t.Fatalf("assignStmt.Value is not *ast.IntegerLiteral, got %T", assignStmt.Value)
		}
		if valueLit.Value != 42 {
			t.Errorf("assigned value = %d, want %d", valueLit.Value, 42)
		}
	})

	t.Run("Array element assignment with variable index", func(t *testing.T) {
		input := `arr[i] := value;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		assignStmt, ok := program.Statements[0].(*ast.AssignmentStatement)
		if !ok {
			t.Fatalf("statement is not *ast.AssignmentStatement, got %T", program.Statements[0])
		}

		// Verify Target is an IndexExpression
		indexExpr, ok := assignStmt.Target.(*ast.IndexExpression)
		if !ok {
			t.Fatalf("assignStmt.Target is not *ast.IndexExpression, got %T", assignStmt.Target)
		}

		// Verify index is identifier
		indexIdent, ok := indexExpr.Index.(*ast.Identifier)
		if !ok {
			t.Fatalf("indexExpr.Index is not *ast.Identifier, got %T", indexExpr.Index)
		}
		if indexIdent.Value != "i" {
			t.Errorf("index identifier = %q, want %q", indexIdent.Value, "i")
		}
	})

	t.Run("Nested array assignment", func(t *testing.T) {
		input := `matrix[i][j] := 99;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		assignStmt, ok := program.Statements[0].(*ast.AssignmentStatement)
		if !ok {
			t.Fatalf("statement is not *ast.AssignmentStatement, got %T", program.Statements[0])
		}

		// Verify Target is an IndexExpression (outer [j])
		outerIndex, ok := assignStmt.Target.(*ast.IndexExpression)
		if !ok {
			t.Fatalf("assignStmt.Target is not *ast.IndexExpression, got %T", assignStmt.Target)
		}

		// Verify Left is also an IndexExpression (inner [i])
		innerIndex, ok := outerIndex.Left.(*ast.IndexExpression)
		if !ok {
			t.Fatalf("outerIndex.Left is not *ast.IndexExpression, got %T", outerIndex.Left)
		}

		// Verify the base array identifier
		baseIdent, ok := innerIndex.Left.(*ast.Identifier)
		if !ok {
			t.Fatalf("innerIndex.Left is not *ast.Identifier, got %T", innerIndex.Left)
		}
		if baseIdent.Value != "matrix" {
			t.Errorf("matrix name = %q, want %q", baseIdent.Value, "matrix")
		}
	})

	t.Run("Array assignment with expression index", func(t *testing.T) {
		input := `arr[i + 1] := value;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		assignStmt, ok := program.Statements[0].(*ast.AssignmentStatement)
		if !ok {
			t.Fatalf("statement is not *ast.AssignmentStatement, got %T", program.Statements[0])
		}

		// Verify Target is an IndexExpression
		indexExpr, ok := assignStmt.Target.(*ast.IndexExpression)
		if !ok {
			t.Fatalf("assignStmt.Target is not *ast.IndexExpression, got %T", assignStmt.Target)
		}

		// Verify index is a binary expression
		_, ok = indexExpr.Index.(*ast.BinaryExpression)
		if !ok {
			t.Fatalf("indexExpr.Index is not *ast.BinaryExpression, got %T", indexExpr.Index)
		}
	})
}

// checkParserErrors is defined in parser_test.go
