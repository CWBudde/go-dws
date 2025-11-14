package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

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
		// NOTE: "new Integer;" is now valid syntax
		// It will fail at semantic analysis since Integer is not a class, but parses OK
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

// ============================================================================
// Multi-Index Comma Syntax Tests
// ============================================================================

func TestParseMultiIndexCommaSyntax(t *testing.T) {
	t.Run("Two-dimensional comma syntax", func(t *testing.T) {
		input := `arr[i, j];`

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

		// arr[i, j] should desugar to (arr[i])[j]
		// So outerIndex.Left should be an IndexExpression
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

		// Test first index (i)
		firstIdx, ok := innerIndex.Index.(*ast.Identifier)
		if !ok {
			t.Fatalf("innerIndex.Index is not *ast.Identifier, got %T", innerIndex.Index)
		}
		if firstIdx.Value != "i" {
			t.Errorf("firstIdx.Value = %s, want 'i'", firstIdx.Value)
		}

		// Test second index (j)
		secondIdx, ok := outerIndex.Index.(*ast.Identifier)
		if !ok {
			t.Fatalf("outerIndex.Index is not *ast.Identifier, got %T", outerIndex.Index)
		}
		if secondIdx.Value != "j" {
			t.Errorf("secondIdx.Value = %s, want 'j'", secondIdx.Value)
		}
	})

	t.Run("Three-dimensional comma syntax", func(t *testing.T) {
		input := `arr[i, j, k];`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		stmt := program.Statements[0].(*ast.ExpressionStatement)

		// arr[i, j, k] should desugar to ((arr[i])[j])[k]
		// Outermost: [k]
		outermost, ok := stmt.Expression.(*ast.IndexExpression)
		if !ok {
			t.Fatalf("expression is not *ast.IndexExpression, got %T", stmt.Expression)
		}
		kIdx, ok := outermost.Index.(*ast.Identifier)
		if !ok || kIdx.Value != "k" {
			t.Errorf("outermost index should be 'k', got %v", outermost.Index)
		}

		// Middle: [j]
		middle, ok := outermost.Left.(*ast.IndexExpression)
		if !ok {
			t.Fatalf("outermost.Left is not *ast.IndexExpression, got %T", outermost.Left)
		}
		jIdx, ok := middle.Index.(*ast.Identifier)
		if !ok || jIdx.Value != "j" {
			t.Errorf("middle index should be 'j', got %v", middle.Index)
		}

		// Innermost: arr[i]
		innermost, ok := middle.Left.(*ast.IndexExpression)
		if !ok {
			t.Fatalf("middle.Left is not *ast.IndexExpression, got %T", middle.Left)
		}
		iIdx, ok := innermost.Index.(*ast.Identifier)
		if !ok || iIdx.Value != "i" {
			t.Errorf("innermost index should be 'i', got %v", innermost.Index)
		}

		arrIdent, ok := innermost.Left.(*ast.Identifier)
		if !ok || arrIdent.Value != "arr" {
			t.Errorf("array name should be 'arr', got %v", innermost.Left)
		}
	})

	t.Run("Comma syntax with literal indices", func(t *testing.T) {
		input := `matrix[0, 1];`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		outerIndex := stmt.Expression.(*ast.IndexExpression)
		innerIndex := outerIndex.Left.(*ast.IndexExpression)

		// Verify array name
		arrIdent := innerIndex.Left.(*ast.Identifier)
		if arrIdent.Value != "matrix" {
			t.Errorf("array name = %s, want 'matrix'", arrIdent.Value)
		}

		// Verify indices are integers
		firstIdx, ok := innerIndex.Index.(*ast.IntegerLiteral)
		if !ok || firstIdx.Value != 0 {
			t.Errorf("first index should be 0, got %v", innerIndex.Index)
		}

		secondIdx, ok := outerIndex.Index.(*ast.IntegerLiteral)
		if !ok || secondIdx.Value != 1 {
			t.Errorf("second index should be 1, got %v", outerIndex.Index)
		}
	})

	t.Run("Comma syntax with complex expressions", func(t *testing.T) {
		input := `data[i + 1, j * 2];`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		outerIndex := stmt.Expression.(*ast.IndexExpression)
		innerIndex := outerIndex.Left.(*ast.IndexExpression)

		// Verify first index is a binary expression (i + 1)
		_, ok := innerIndex.Index.(*ast.BinaryExpression)
		if !ok {
			t.Fatalf("first index is not *ast.BinaryExpression, got %T", innerIndex.Index)
		}

		// Verify second index is a binary expression (j * 2)
		_, ok = outerIndex.Index.(*ast.BinaryExpression)
		if !ok {
			t.Fatalf("second index is not *ast.BinaryExpression, got %T", outerIndex.Index)
		}
	})

	t.Run("Mixed comma and bracket syntax", func(t *testing.T) {
		input := `arr[i, j][k];`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		stmt := program.Statements[0].(*ast.ExpressionStatement)

		// arr[i, j][k] should parse as ((arr[i])[j])[k]
		// The outermost [k] is applied to the result of arr[i, j]
		outermost, ok := stmt.Expression.(*ast.IndexExpression)
		if !ok {
			t.Fatalf("expression is not *ast.IndexExpression, got %T", stmt.Expression)
		}

		// Check that it has the correct nested structure
		kIdx, ok := outermost.Index.(*ast.Identifier)
		if !ok || kIdx.Value != "k" {
			t.Errorf("outermost index should be 'k', got %v", outermost.Index)
		}

		// The left side should be the desugared arr[i, j]
		_, ok = outermost.Left.(*ast.IndexExpression)
		if !ok {
			t.Fatalf("outermost.Left is not *ast.IndexExpression, got %T", outermost.Left)
		}
	})

	t.Run("Assignment to comma-indexed array", func(t *testing.T) {
		input := `matrix[i, j] := 42;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		stmt := program.Statements[0].(*ast.AssignmentStatement)

		// Target should be a nested IndexExpression
		outerIndex, ok := stmt.Target.(*ast.IndexExpression)
		if !ok {
			t.Fatalf("assignment target is not *ast.IndexExpression, got %T", stmt.Target)
		}

		innerIndex, ok := outerIndex.Left.(*ast.IndexExpression)
		if !ok {
			t.Fatalf("outerIndex.Left is not *ast.IndexExpression, got %T", outerIndex.Left)
		}

		// Verify it's the matrix identifier
		matrixIdent, ok := innerIndex.Left.(*ast.Identifier)
		if !ok || matrixIdent.Value != "matrix" {
			t.Errorf("expected 'matrix', got %v", innerIndex.Left)
		}

		// Verify value is the integer 42
		val, ok := stmt.Value.(*ast.IntegerLiteral)
		if !ok || val.Value != 42 {
			t.Errorf("expected 42, got %v", stmt.Value)
		}
	})

	t.Run("Comma syntax equivalence with nested brackets", func(t *testing.T) {
		// Verify that arr[i, j] and arr[i][j] produce the same AST structure
		commaInput := `arr[i, j];`
		nestedInput := `arr[i][j];`

		// Parse comma syntax
		l1 := lexer.New(commaInput)
		p1 := New(l1)
		program1 := p1.ParseProgram()
		checkParserErrors(t, p1)

		// Parse nested bracket syntax
		l2 := lexer.New(nestedInput)
		p2 := New(l2)
		program2 := p2.ParseProgram()
		checkParserErrors(t, p2)

		// Both should produce the same String() representation
		stmt1 := program1.Statements[0].(*ast.ExpressionStatement)
		stmt2 := program2.Statements[0].(*ast.ExpressionStatement)

		str1 := stmt1.Expression.String()
		str2 := stmt2.Expression.String()

		if str1 != str2 {
			t.Errorf("AST structures differ:\nComma: %s\nNested: %s", str1, str2)
		}
	})

	t.Run("Single index still works", func(t *testing.T) {
		// Ensure we didn't break single-index parsing
		input := `arr[i];`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		indexExpr, ok := stmt.Expression.(*ast.IndexExpression)
		if !ok {
			t.Fatalf("expression is not *ast.IndexExpression, got %T", stmt.Expression)
		}

		// Left should be identifier 'arr', not another IndexExpression
		arrIdent, ok := indexExpr.Left.(*ast.Identifier)
		if !ok {
			t.Fatalf("indexExpr.Left is not *ast.Identifier, got %T", indexExpr.Left)
		}
		if arrIdent.Value != "arr" {
			t.Errorf("array name = %s, want 'arr'", arrIdent.Value)
		}

		// Index should be identifier 'i'
		idx, ok := indexExpr.Index.(*ast.Identifier)
		if !ok || idx.Value != "i" {
			t.Errorf("index should be 'i', got %v", indexExpr.Index)
		}
	})
}
