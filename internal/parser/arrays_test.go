package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// ============================================================================
// Array Type Declaration Parser Tests
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
// Array Literal Parser Tests
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
// Array Indexing Parser Tests
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
// Combined Array Tests
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
// Array Assignment Parser Tests
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

// ============================================================================
// Array Instantiation with 'new' Keyword Parser Tests
// ============================================================================

func TestParseNewArrayExpression(t *testing.T) {
	t.Run("Simple 1D array instantiation", func(t *testing.T) {
		input := `new Integer[16];`

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

		newArrayExpr, ok := stmt.Expression.(*ast.NewArrayExpression)
		if !ok {
			t.Fatalf("expression is not *ast.NewArrayExpression, got %T", stmt.Expression)
		}

		// Test element type name
		if newArrayExpr.ElementTypeName.Value != "Integer" {
			t.Errorf("ElementTypeName = %s, want 'Integer'", newArrayExpr.ElementTypeName.Value)
		}

		// Test dimensions
		if len(newArrayExpr.Dimensions) != 1 {
			t.Fatalf("Dimensions should contain 1 element, got %d", len(newArrayExpr.Dimensions))
		}

		// Test dimension value
		dimLit, ok := newArrayExpr.Dimensions[0].(*ast.IntegerLiteral)
		if !ok {
			t.Fatalf("Dimension is not *ast.IntegerLiteral, got %T", newArrayExpr.Dimensions[0])
		}
		if dimLit.Value != 16 {
			t.Errorf("Dimension value = %d, want 16", dimLit.Value)
		}

		// Test String() method
		expected := "new Integer[16]"
		if newArrayExpr.String() != expected {
			t.Errorf("String() = %s, want %s", newArrayExpr.String(), expected)
		}
	})

	t.Run("2D array instantiation", func(t *testing.T) {
		input := `new String[10, 20];`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		newArrayExpr, ok := stmt.Expression.(*ast.NewArrayExpression)
		if !ok {
			t.Fatalf("expression is not *ast.NewArrayExpression, got %T", stmt.Expression)
		}

		// Test element type
		if newArrayExpr.ElementTypeName.Value != "String" {
			t.Errorf("ElementTypeName = %s, want 'String'", newArrayExpr.ElementTypeName.Value)
		}

		// Test dimensions count
		if len(newArrayExpr.Dimensions) != 2 {
			t.Fatalf("Dimensions should contain 2 elements, got %d", len(newArrayExpr.Dimensions))
		}

		// Test first dimension
		dim1, ok := newArrayExpr.Dimensions[0].(*ast.IntegerLiteral)
		if !ok {
			t.Fatalf("First dimension is not *ast.IntegerLiteral, got %T", newArrayExpr.Dimensions[0])
		}
		if dim1.Value != 10 {
			t.Errorf("First dimension = %d, want 10", dim1.Value)
		}

		// Test second dimension
		dim2, ok := newArrayExpr.Dimensions[1].(*ast.IntegerLiteral)
		if !ok {
			t.Fatalf("Second dimension is not *ast.IntegerLiteral, got %T", newArrayExpr.Dimensions[1])
		}
		if dim2.Value != 20 {
			t.Errorf("Second dimension = %d, want 20", dim2.Value)
		}

		// Test String() method
		expected := "new String[10, 20]"
		if newArrayExpr.String() != expected {
			t.Errorf("String() = %s, want %s", newArrayExpr.String(), expected)
		}
	})

	t.Run("Array with expression-based size", func(t *testing.T) {
		input := `new Float[Length(arr)+1];`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		newArrayExpr, ok := stmt.Expression.(*ast.NewArrayExpression)
		if !ok {
			t.Fatalf("expression is not *ast.NewArrayExpression, got %T", stmt.Expression)
		}

		// Test element type
		if newArrayExpr.ElementTypeName.Value != "Float" {
			t.Errorf("ElementTypeName = %s, want 'Float'", newArrayExpr.ElementTypeName.Value)
		}

		// Test dimension is an expression
		if len(newArrayExpr.Dimensions) != 1 {
			t.Fatalf("Dimensions should contain 1 element, got %d", len(newArrayExpr.Dimensions))
		}

		// Dimension should be a binary expression (Length(arr) + 1)
		binExpr, ok := newArrayExpr.Dimensions[0].(*ast.BinaryExpression)
		if !ok {
			t.Fatalf("Dimension is not *ast.BinaryExpression, got %T", newArrayExpr.Dimensions[0])
		}

		if binExpr.Operator != "+" {
			t.Errorf("Binary operator = %s, want '+'", binExpr.Operator)
		}
	})

	t.Run("3D array instantiation", func(t *testing.T) {
		input := `new Boolean[5, 10, 15];`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		newArrayExpr, ok := stmt.Expression.(*ast.NewArrayExpression)
		if !ok {
			t.Fatalf("expression is not *ast.NewArrayExpression, got %T", stmt.Expression)
		}

		// Test dimensions count for 3D array
		if len(newArrayExpr.Dimensions) != 3 {
			t.Fatalf("Dimensions should contain 3 elements, got %d", len(newArrayExpr.Dimensions))
		}

		// Verify all three dimensions are integers
		for i, dim := range newArrayExpr.Dimensions {
			_, ok := dim.(*ast.IntegerLiteral)
			if !ok {
				t.Errorf("Dimension %d is not *ast.IntegerLiteral, got %T", i, dim)
			}
		}
	})

	t.Run("Array instantiation in variable declaration", func(t *testing.T) {
		input := `var a := new Integer[16];`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		varDecl, ok := program.Statements[0].(*ast.VarDeclStatement)
		if !ok {
			t.Fatalf("statement is not *ast.VarDeclStatement, got %T", program.Statements[0])
		}

		// Value should be NewArrayExpression
		newArrayExpr, ok := varDecl.Value.(*ast.NewArrayExpression)
		if !ok {
			t.Fatalf("variable value is not *ast.NewArrayExpression, got %T", varDecl.Value)
		}

		if newArrayExpr.ElementTypeName.Value != "Integer" {
			t.Errorf("ElementTypeName = %s, want 'Integer'", newArrayExpr.ElementTypeName.Value)
		}
	})

	t.Run("Class instantiation still works (backward compatibility)", func(t *testing.T) {
		input := `new TPoint(10, 20);`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		stmt := program.Statements[0].(*ast.ExpressionStatement)

		// Should parse as NewExpression (class), not NewArrayExpression
		newExpr, ok := stmt.Expression.(*ast.NewExpression)
		if !ok {
			t.Fatalf("expression is not *ast.NewExpression, got %T", stmt.Expression)
		}

		if newExpr.ClassName.Value != "TPoint" {
			t.Errorf("ClassName = %s, want 'TPoint'", newExpr.ClassName.Value)
		}

		if len(newExpr.Arguments) != 2 {
			t.Fatalf("Arguments should contain 2 elements, got %d", len(newExpr.Arguments))
		}
	})
}

func TestParseNewArrayExpressionErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name:        "Missing opening bracket",
			input:       `new Integer;`,
			expectedErr: "expected '[' or '('",
		},
		{
			name:        "Missing closing bracket",
			input:       `new Integer[16;`,
			expectedErr: "expected next token to be RBRACK",
		},
		{
			name:        "Empty brackets",
			input:       `new Integer[];`,
			expectedErr: "expected expression for array dimension",
		},
		{
			name:        "Missing comma between dimensions",
			input:       `new Integer[10 20];`,
			expectedErr: "expected next token to be RBRACK",
		},
		{
			name:        "Trailing comma",
			input:       `new Integer[10,];`,
			expectedErr: "expected expression for array dimension",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			_ = p.ParseProgram()

			errors := p.Errors()
			if len(errors) == 0 {
				t.Fatalf("Expected parser error, but got none")
			}

			// Check that at least one error contains the expected message
			found := false
			for _, err := range errors {
				if contains(err, tt.expectedErr) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected error containing '%s', but got errors: %v", tt.expectedErr, errors)
			}
		})
	}
}
