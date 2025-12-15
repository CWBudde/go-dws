package parser

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Helper Functions
// ============================================================================

func getExprFromStmt(t *testing.T, stmt ast.Statement) ast.Expression {
	t.Helper()
	exprStmt, ok := stmt.(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement is not *ast.ExpressionStatement, got %T", stmt)
	}
	return exprStmt.Expression
}

func asIndexExpr(t *testing.T, expr ast.Expression) *ast.IndexExpression {
	t.Helper()
	indexExpr, ok := expr.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("expression is not *ast.IndexExpression, got %T", expr)
	}
	return indexExpr
}

func asIdent(t *testing.T, node ast.Node) *ast.Identifier {
	t.Helper()
	ident, ok := node.(*ast.Identifier)
	if !ok {
		t.Fatalf("node is not *ast.Identifier, got %T", node)
	}
	return ident
}

func checkIdent(t *testing.T, node ast.Node, expected string) {
	t.Helper()
	ident := asIdent(t, node)
	if ident.Value != expected {
		t.Errorf("identifier = %s, want %s", ident.Value, expected)
	}
}

func asInt(t *testing.T, node ast.Node) *ast.IntegerLiteral {
	t.Helper()
	lit, ok := node.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("node is not *ast.IntegerLiteral, got %T", node)
	}
	return lit
}

func checkInt(t *testing.T, node ast.Node, expected int64) {
	t.Helper()
	lit := asInt(t, node)
	if lit.Value != expected {
		t.Errorf("integer literal = %d, want %d", lit.Value, expected)
	}
}

func asBinary(t *testing.T, node ast.Node) *ast.BinaryExpression {
	t.Helper()
	bin, ok := node.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("node is not *ast.BinaryExpression, got %T", node)
	}
	return bin
}

func asAssignment(t *testing.T, stmt ast.Statement) *ast.AssignmentStatement {
	t.Helper()
	assign, ok := stmt.(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("statement is not *ast.AssignmentStatement, got %T", stmt)
	}
	return assign
}

func asNewArray(t *testing.T, expr ast.Expression) *ast.NewArrayExpression {
	t.Helper()
	newArr, ok := expr.(*ast.NewArrayExpression)
	if !ok {
		t.Fatalf("expression is not *ast.NewArrayExpression, got %T", expr)
	}
	return newArr
}

// ============================================================================
// Array Indexing Parser Tests
// ============================================================================

func TestParseArrayIndexing(t *testing.T) {
	tests := []struct {
		checkStmt func(t *testing.T, stmt ast.Statement)
		name      string
		input     string
	}{
		{
			name:  "Simple array indexing",
			input: `arr[0];`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				indexExpr := asIndexExpr(t, getExprFromStmt(t, stmt))
				checkIdent(t, indexExpr.Left, "arr")
				checkInt(t, indexExpr.Index, 0)
			},
		},
		{
			name:  "Array indexing with variable",
			input: `arr[i];`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				indexExpr := asIndexExpr(t, getExprFromStmt(t, stmt))
				checkIdent(t, indexExpr.Index, "i")
			},
		},
		{
			name:  "Array indexing with expression",
			input: `arr[i + 1];`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				indexExpr := asIndexExpr(t, getExprFromStmt(t, stmt))
				asBinary(t, indexExpr.Index)
			},
		},
		{
			name:  "Nested array indexing",
			input: `arr[i][j];`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				outerIndex := asIndexExpr(t, getExprFromStmt(t, stmt))
				innerIndex := asIndexExpr(t, outerIndex.Left)
				checkIdent(t, innerIndex.Left, "arr")
			},
		},
		{
			name:  "String() method for index expression",
			input: `arr[5];`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				indexExpr := asIndexExpr(t, getExprFromStmt(t, stmt))
				expected := "(arr[5])"
				if indexExpr.String() != expected {
					t.Errorf("String() = %s, want %s", indexExpr.String(), expected)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt := helperParseAndGetStatement(t, tt.input)
			if tt.checkStmt != nil {
				tt.checkStmt(t, stmt)
			}
		})
	}
}

// ============================================================================
// Combined Array Tests
// ============================================================================

func TestArrayDeclarationAndUsage(t *testing.T) {
	t.Run("Type declaration and array access", func(t *testing.T) {
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

		arrayDecl, ok := program.Statements[0].(*ast.ArrayDecl)
		if !ok {
			t.Fatalf("first statement is not *ast.ArrayDecl, got %T", program.Statements[0])
		}
		if arrayDecl.Name.Value != "TMyArray" {
			t.Errorf("arrayDecl.Name.Value = %s, want 'TMyArray'", arrayDecl.Name.Value)
		}
	})

	t.Run("Reading array element", func(t *testing.T) {
		input := `x := arr[i];`
		stmt := helperParseAndGetStatement(t, input)
		assignStmt := asAssignment(t, stmt)
		asIndexExpr(t, assignStmt.Value)
	})
}

// ============================================================================
// Array Assignment Parser Tests
// ============================================================================

func TestParseArrayAssignment(t *testing.T) {
	tests := []struct {
		checkStmt func(t *testing.T, stmt ast.Statement)
		name      string
		input     string
	}{
		{
			name: "Simple array element assignment",
			input: `
			begin
				arr[0] := 42;
			end
			`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				block, ok := stmt.(*ast.BlockStatement)
				if !ok {
					t.Fatalf("statement is not *ast.BlockStatement, got %T", stmt)
				}
				if len(block.Statements) != 1 {
					t.Fatalf("block.Statements should contain 1 statement, got %d", len(block.Statements))
				}

				assignStmt := asAssignment(t, block.Statements[0])
				indexExpr := asIndexExpr(t, assignStmt.Target)

				checkIdent(t, indexExpr.Left, "arr")
				checkInt(t, indexExpr.Index, 0)
				checkInt(t, assignStmt.Value, 42)
			},
		},
		{
			name:  "Array element assignment with variable index",
			input: `arr[i] := value;`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				assignStmt := asAssignment(t, stmt)
				indexExpr := asIndexExpr(t, assignStmt.Target)
				checkIdent(t, indexExpr.Index, "i")
			},
		},
		{
			name:  "Nested array assignment",
			input: `matrix[i][j] := 99;`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				assignStmt := asAssignment(t, stmt)
				outerIndex := asIndexExpr(t, assignStmt.Target)
				innerIndex := asIndexExpr(t, outerIndex.Left)
				checkIdent(t, innerIndex.Left, "matrix")
			},
		},
		{
			name:  "Array assignment with expression index",
			input: `arr[i + 1] := value;`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				assignStmt := asAssignment(t, stmt)
				indexExpr := asIndexExpr(t, assignStmt.Target)
				asBinary(t, indexExpr.Index)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt := helperParseAndGetStatement(t, tt.input)
			if tt.checkStmt != nil {
				tt.checkStmt(t, stmt)
			}
		})
	}
}

// ============================================================================
// Array Instantiation with 'new' Keyword Parser Tests
// ============================================================================

func TestParseNewArrayExpression(t *testing.T) {
	tests := []struct {
		checkStmt func(t *testing.T, stmt ast.Statement)
		name      string
		input     string
	}{
		{
			name:  "Simple 1D array instantiation",
			input: `new Integer[16];`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				newArrayExpr := asNewArray(t, getExprFromStmt(t, stmt))
				checkIdent(t, newArrayExpr.ElementTypeName, "Integer")

				if len(newArrayExpr.Dimensions) != 1 {
					t.Fatalf("Dimensions count = %d, want 1", len(newArrayExpr.Dimensions))
				}
				checkInt(t, newArrayExpr.Dimensions[0], 16)

				expected := "new Integer[16]"
				if newArrayExpr.String() != expected {
					t.Errorf("String() = %s, want %s", newArrayExpr.String(), expected)
				}
			},
		},
		{
			name:  "2D array instantiation",
			input: `new String[10, 20];`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				newArrayExpr := asNewArray(t, getExprFromStmt(t, stmt))
				checkIdent(t, newArrayExpr.ElementTypeName, "String")

				if len(newArrayExpr.Dimensions) != 2 {
					t.Fatalf("Dimensions count = %d, want 2", len(newArrayExpr.Dimensions))
				}
				checkInt(t, newArrayExpr.Dimensions[0], 10)
				checkInt(t, newArrayExpr.Dimensions[1], 20)

				expected := "new String[10, 20]"
				if newArrayExpr.String() != expected {
					t.Errorf("String() = %s, want %s", newArrayExpr.String(), expected)
				}
			},
		},
		{
			name:  "Array with expression-based size",
			input: `new Float[Length(arr)+1];`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				newArrayExpr := asNewArray(t, getExprFromStmt(t, stmt))
				checkIdent(t, newArrayExpr.ElementTypeName, "Float")

				if len(newArrayExpr.Dimensions) != 1 {
					t.Fatalf("Dimensions count = %d, want 1", len(newArrayExpr.Dimensions))
				}
				binExpr := asBinary(t, newArrayExpr.Dimensions[0])
				if binExpr.Operator != "+" {
					t.Errorf("Binary operator = %s, want '+'", binExpr.Operator)
				}
			},
		},
		{
			name:  "3D array instantiation",
			input: `new Boolean[5, 10, 15];`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				newArrayExpr := asNewArray(t, getExprFromStmt(t, stmt))
				if len(newArrayExpr.Dimensions) != 3 {
					t.Fatalf("Dimensions count = %d, want 3", len(newArrayExpr.Dimensions))
				}
				for i, dim := range newArrayExpr.Dimensions {
					asInt(t, dim) // Just verify they are int literals
					_ = i
				}
			},
		},
		{
			name:  "Array instantiation in variable declaration",
			input: `var a := new Integer[16];`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				varDecl, ok := stmt.(*ast.VarDeclStatement)
				if !ok {
					t.Fatalf("statement is not *ast.VarDeclStatement, got %T", stmt)
				}
				newArrayExpr := asNewArray(t, varDecl.Value)
				checkIdent(t, newArrayExpr.ElementTypeName, "Integer")
			},
		},
		{
			name:  "Class instantiation still works (backward compatibility)",
			input: `new TPoint(10, 20);`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				expr := getExprFromStmt(t, stmt)
				newExpr, ok := expr.(*ast.NewExpression)
				if !ok {
					t.Fatalf("expression is not *ast.NewExpression, got %T", expr)
				}
				checkIdent(t, newExpr.ClassName, "TPoint")
				if len(newExpr.Arguments) != 2 {
					t.Fatalf("Arguments count = %d, want 2", len(newExpr.Arguments))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt := helperParseAndGetStatement(t, tt.input)
			if tt.checkStmt != nil {
				tt.checkStmt(t, stmt)
			}
		})
	}
}

func TestParseNewArrayExpressionErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name:        "Missing closing bracket",
			input:       `new Integer[16;`,
			expectedErr: "expected ',' or 'RBRACK'",
		},
		{
			name:        "Empty brackets",
			input:       `new Integer[];`,
			expectedErr: "expected expression for array dimension",
		},
		{
			name:        "Missing comma between dimensions",
			input:       `new Integer[10 20];`,
			expectedErr: "expected ',' or 'RBRACK'",
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

			found := false
			for _, err := range errors {
				if strings.Contains(err.Error(), tt.expectedErr) {
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
	tests := []struct {
		checkStmt func(t *testing.T, stmt ast.Statement)
		name      string
		input     string
	}{
		{
			name:  "Two-dimensional comma syntax",
			input: `arr[i, j];`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				outerIndex := asIndexExpr(t, getExprFromStmt(t, stmt))
				innerIndex := asIndexExpr(t, outerIndex.Left)

				checkIdent(t, innerIndex.Left, "arr")
				checkIdent(t, innerIndex.Index, "i")
				checkIdent(t, outerIndex.Index, "j")
			},
		},
		{
			name:  "Three-dimensional comma syntax",
			input: `arr[i, j, k];`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				outermost := asIndexExpr(t, getExprFromStmt(t, stmt))
				checkIdent(t, outermost.Index, "k")

				middle := asIndexExpr(t, outermost.Left)
				checkIdent(t, middle.Index, "j")

				innermost := asIndexExpr(t, middle.Left)
				checkIdent(t, innermost.Index, "i")
				checkIdent(t, innermost.Left, "arr")
			},
		},
		{
			name:  "Comma syntax with literal indices",
			input: `matrix[0, 1];`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				outerIndex := asIndexExpr(t, getExprFromStmt(t, stmt))
				innerIndex := asIndexExpr(t, outerIndex.Left)

				checkIdent(t, innerIndex.Left, "matrix")
				checkInt(t, innerIndex.Index, 0)
				checkInt(t, outerIndex.Index, 1)
			},
		},
		{
			name:  "Comma syntax with complex expressions",
			input: `data[i + 1, j * 2];`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				outerIndex := asIndexExpr(t, getExprFromStmt(t, stmt))
				innerIndex := asIndexExpr(t, outerIndex.Left)

				asBinary(t, innerIndex.Index)
				asBinary(t, outerIndex.Index)
			},
		},
		{
			name:  "Mixed comma and bracket syntax",
			input: `arr[i, j][k];`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				outermost := asIndexExpr(t, getExprFromStmt(t, stmt))
				checkIdent(t, outermost.Index, "k")
				asIndexExpr(t, outermost.Left) // Ensure left side is also an index expression (arr[i,j])
			},
		},
		{
			name:  "Assignment to comma-indexed array",
			input: `matrix[i, j] := 42;`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				assignStmt := asAssignment(t, stmt)
				outerIndex := asIndexExpr(t, assignStmt.Target)
				innerIndex := asIndexExpr(t, outerIndex.Left)

				checkIdent(t, innerIndex.Left, "matrix")
				checkInt(t, assignStmt.Value, 42)
			},
		},
		{
			name:  "Single index still works",
			input: `arr[i];`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				indexExpr := asIndexExpr(t, getExprFromStmt(t, stmt))
				checkIdent(t, indexExpr.Left, "arr")
				checkIdent(t, indexExpr.Index, "i")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt := helperParseAndGetStatement(t, tt.input)
			if tt.checkStmt != nil {
				tt.checkStmt(t, stmt)
			}
		})
	}

	t.Run("Comma syntax equivalence with nested brackets", func(t *testing.T) {
		commaInput := `arr[i, j];`
		nestedInput := `arr[i][j];`

		stmt1 := helperParseAndGetStatement(t, commaInput)
		stmt2 := helperParseAndGetStatement(t, nestedInput)

		expr1 := getExprFromStmt(t, stmt1)
		expr2 := getExprFromStmt(t, stmt2)

		if expr1.String() != expr2.String() {
			t.Errorf("AST structures differ:\nComma: %s\nNested: %s", expr1.String(), expr2.String())
		}
	})
}
