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

	// Test nested array type declarations (array of array)
	t.Run("Nested dynamic arrays - 2D", func(t *testing.T) {
		input := `type Matrix = array of array of Float;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		arrayDecl, ok := program.Statements[0].(*ast.ArrayDecl)
		if !ok {
			t.Fatalf("statement is not *ast.ArrayDecl, got %T", program.Statements[0])
		}

		if arrayDecl.Name.Value != "Matrix" {
			t.Errorf("arrayDecl.Name.Value = %s, want 'Matrix'", arrayDecl.Name.Value)
		}

		// Verify the element type string representation
		expectedType := "array of Float"
		if arrayDecl.ArrayType.ElementType.String() != expectedType {
			t.Errorf("arrayDecl.ArrayType.ElementType.String() = %s, want %s",
				arrayDecl.ArrayType.ElementType.String(), expectedType)
		}
	})

	t.Run("Nested dynamic arrays - 3D", func(t *testing.T) {
		input := `type Tensor = array of array of array of Integer;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		arrayDecl, ok := program.Statements[0].(*ast.ArrayDecl)
		if !ok {
			t.Fatalf("statement is not *ast.ArrayDecl, got %T", program.Statements[0])
		}

		if arrayDecl.Name.Value != "Tensor" {
			t.Errorf("arrayDecl.Name.Value = %s, want 'Tensor'", arrayDecl.Name.Value)
		}

		// Verify the element type string representation for 3D array
		expectedType := "array of array of Integer"
		if arrayDecl.ArrayType.ElementType.String() != expectedType {
			t.Errorf("arrayDecl.ArrayType.ElementType.String() = %s, want %s",
				arrayDecl.ArrayType.ElementType.String(), expectedType)
		}
	})

	t.Run("Nested static arrays", func(t *testing.T) {
		input := `type Grid = array[1..10] of array[1..20] of Boolean;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		arrayDecl, ok := program.Statements[0].(*ast.ArrayDecl)
		if !ok {
			t.Fatalf("statement is not *ast.ArrayDecl, got %T", program.Statements[0])
		}

		if arrayDecl.Name.Value != "Grid" {
			t.Errorf("arrayDecl.Name.Value = %s, want 'Grid'", arrayDecl.Name.Value)
		}

		// Verify the element type string representation for nested static arrays
		expectedType := "array[1..20] of Boolean"
		if arrayDecl.ArrayType.ElementType.String() != expectedType {
			t.Errorf("arrayDecl.ArrayType.ElementType.String() = %s, want %s",
				arrayDecl.ArrayType.ElementType.String(), expectedType)
		}
	})

	t.Run("Mixed static and dynamic nested arrays", func(t *testing.T) {
		input := `type MixedArray = array of array[0..99] of String;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		arrayDecl, ok := program.Statements[0].(*ast.ArrayDecl)
		if !ok {
			t.Fatalf("statement is not *ast.ArrayDecl, got %T", program.Statements[0])
		}

		if arrayDecl.Name.Value != "MixedArray" {
			t.Errorf("arrayDecl.Name.Value = %s, want 'MixedArray'", arrayDecl.Name.Value)
		}

		// Verify the element type string representation
		expectedType := "array[0..99] of String"
		if arrayDecl.ArrayType.ElementType.String() != expectedType {
			t.Errorf("arrayDecl.ArrayType.ElementType.String() = %s, want %s",
				arrayDecl.ArrayType.ElementType.String(), expectedType)
		}
	})
}

// ============================================================================
// Array Literal Parser Tests
// ============================================================================

func TestParseArrayLiteral(t *testing.T) {
	tests := []struct {
		assert       func(t *testing.T, lit *ast.ArrayLiteralExpression)
		name         string
		input        string
		wantElements int
	}{
		{
			name:         "SimpleIntegers",
			input:        `var arr := [1, 2, 3];`,
			wantElements: 3,
			assert: func(t *testing.T, lit *ast.ArrayLiteralExpression) {
				for i, expected := range []int64{1, 2, 3} {
					intLit, ok := lit.Elements[i].(*ast.IntegerLiteral)
					if !ok {
						t.Fatalf("element %d is not *ast.IntegerLiteral, got %T", i, lit.Elements[i])
					}
					if intLit.Value != expected {
						t.Fatalf("element %d value = %d, want %d", i, intLit.Value, expected)
					}
				}
			},
		},
		{
			name:         "WithExpressions",
			input:        `var arr := [x + 1, Length(s), 42];`,
			wantElements: 3,
			assert: func(t *testing.T, lit *ast.ArrayLiteralExpression) {
				if _, ok := lit.Elements[0].(*ast.BinaryExpression); !ok {
					t.Fatalf("element 0 is not *ast.BinaryExpression, got %T", lit.Elements[0])
				}

				callExpr, ok := lit.Elements[1].(*ast.CallExpression)
				if !ok {
					t.Fatalf("element 1 is not *ast.CallExpression, got %T", lit.Elements[1])
				}
				funcIdent, ok := callExpr.Function.(*ast.Identifier)
				if !ok || funcIdent.Value != "Length" {
					t.Fatalf("call expression function = %T (%v), want Identifier 'Length'", callExpr.Function, funcIdent)
				}

				intLit, ok := lit.Elements[2].(*ast.IntegerLiteral)
				if !ok || intLit.Value != 42 {
					t.Fatalf("element 2 = %T (value=%v), want IntegerLiteral 42", lit.Elements[2], intLit)
				}
			},
		},
		{
			name:         "NestedArrays",
			input:        `var matrix := [[1, 2], [3, 4]];`,
			wantElements: 2,
			assert: func(t *testing.T, lit *ast.ArrayLiteralExpression) {
				for i, expectedValues := range [][]int64{{1, 2}, {3, 4}} {
					nested, ok := lit.Elements[i].(*ast.ArrayLiteralExpression)
					if !ok {
						t.Fatalf("nested element %d is not *ast.ArrayLiteralExpression, got %T", i, lit.Elements[i])
					}
					if len(nested.Elements) != 2 {
						t.Fatalf("nested element %d length = %d, want 2", i, len(nested.Elements))
					}
					for j, expected := range expectedValues {
						intLit, ok := nested.Elements[j].(*ast.IntegerLiteral)
						if !ok || intLit.Value != expected {
							t.Fatalf("nested element %d[%d] = %T (value=%v), want IntegerLiteral %d", i, j, nested.Elements[j], nested.Elements[j], expected)
						}
					}
				}
			},
		},
		{
			name:         "NegativeNumbers",
			input:        `var vec := [-50.0, 30, 50];`,
			wantElements: 3,
			assert: func(t *testing.T, lit *ast.ArrayLiteralExpression) {
				unary, ok := lit.Elements[0].(*ast.UnaryExpression)
				if !ok || unary.Operator != "-" {
					t.Fatalf("element 0 = %T, want UnaryExpression with operator '-'", lit.Elements[0])
				}
				if _, ok := unary.Right.(*ast.FloatLiteral); !ok {
					t.Fatalf("unary.Right = %T, want *ast.FloatLiteral", unary.Right)
				}

				second, ok := lit.Elements[1].(*ast.IntegerLiteral)
				if !ok || second.Value != 30 {
					t.Fatalf("element 1 = %T (value=%v), want IntegerLiteral 30", lit.Elements[1], second)
				}

				third, ok := lit.Elements[2].(*ast.IntegerLiteral)
				if !ok || third.Value != 50 {
					t.Fatalf("element 2 = %T (value=%v), want IntegerLiteral 50", lit.Elements[2], third)
				}
			},
		},
		{
			name:         "Empty",
			input:        `var arrEmpty := [];`,
			wantElements: 0,
			assert: func(t *testing.T, lit *ast.ArrayLiteralExpression) {
				if len(lit.Elements) != 0 {
					t.Fatalf("len(Elements) = %d, want 0", len(lit.Elements))
				}
			},
		},
		{
			name:         "TrailingComma",
			input:        `var arr := [1, 2, 3, ];`,
			wantElements: 3,
			assert: func(t *testing.T, lit *ast.ArrayLiteralExpression) {
				if len(lit.Elements) != 3 {
					t.Fatalf("len(Elements) = %d, want 3", len(lit.Elements))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
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

			arrayLit, ok := varDecl.Value.(*ast.ArrayLiteralExpression)
			if !ok {
				t.Fatalf("Value is not *ast.ArrayLiteralExpression, got %T", varDecl.Value)
			}

			if len(arrayLit.Elements) != tt.wantElements {
				t.Fatalf("len(Elements) = %d, want %d", len(arrayLit.Elements), tt.wantElements)
			}

			if tt.assert != nil {
				tt.assert(t, arrayLit)
			}
		})
	}
}

func TestParseArrayLiteralErrors(t *testing.T) {
	t.Run("MissingCommaBetweenElements", func(t *testing.T) {
		input := `var arr := [1 2, 3];`

		l := lexer.New(input)
		p := New(l)
		_ = p.ParseProgram()

		if len(p.Errors()) == 0 {
			t.Fatalf("expected parser errors, got none")
		}

		found := false
		for _, err := range p.Errors() {
			if contains(err, "expected ',' or ']' in array literal") {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected error about missing comma, errors: %v", p.Errors())
		}
	})

	t.Run("UnclosedBracket", func(t *testing.T) {
		input := `var arr := [1, 2, 3;`

		l := lexer.New(input)
		p := New(l)
		_ = p.ParseProgram()

		if len(p.Errors()) == 0 {
			t.Fatalf("expected parser errors, got none")
		}

		found := false
		for _, err := range p.Errors() {
			if contains(err, "expected closing ']' for array literal") {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected error about missing closing bracket, errors: %v", p.Errors())
		}
	})
}

// ============================================================================
// Parenthesized Array Literal Tests
// ============================================================================

func TestParseParenthesizedArrayLiteral(t *testing.T) {
	tests := []struct {
		assert       func(t *testing.T, lit *ast.ArrayLiteralExpression)
		name         string
		input        string
		wantElements int
	}{
		{
			name:         "SimpleIntegers",
			input:        `var arr := (1, 2, 3);`,
			wantElements: 3,
			assert: func(t *testing.T, lit *ast.ArrayLiteralExpression) {
				for i, expected := range []int64{1, 2, 3} {
					intLit, ok := lit.Elements[i].(*ast.IntegerLiteral)
					if !ok {
						t.Fatalf("element %d is not *ast.IntegerLiteral, got %T", i, lit.Elements[i])
					}
					if intLit.Value != expected {
						t.Fatalf("element %d value = %d, want %d", i, intLit.Value, expected)
					}
				}
			},
		},
		{
			name:         "WithIdentifiers",
			input:        `var arr := (teOne, teTwo, teThree);`,
			wantElements: 3,
			assert: func(t *testing.T, lit *ast.ArrayLiteralExpression) {
				expectedNames := []string{"teOne", "teTwo", "teThree"}
				for i, expected := range expectedNames {
					ident, ok := lit.Elements[i].(*ast.Identifier)
					if !ok {
						t.Fatalf("element %d is not *ast.Identifier, got %T", i, lit.Elements[i])
					}
					if ident.Value != expected {
						t.Fatalf("element %d value = %s, want %s", i, ident.Value, expected)
					}
				}
			},
		},
		{
			name:         "WithExpressions",
			input:        `var arr := (x + 1, Length(s), 42);`,
			wantElements: 3,
			assert: func(t *testing.T, lit *ast.ArrayLiteralExpression) {
				if _, ok := lit.Elements[0].(*ast.BinaryExpression); !ok {
					t.Fatalf("element 0 is not *ast.BinaryExpression, got %T", lit.Elements[0])
				}

				callExpr, ok := lit.Elements[1].(*ast.CallExpression)
				if !ok {
					t.Fatalf("element 1 is not *ast.CallExpression, got %T", lit.Elements[1])
				}
				funcIdent, ok := callExpr.Function.(*ast.Identifier)
				if !ok || funcIdent.Value != "Length" {
					t.Fatalf("call expression function = %T (%v), want Identifier 'Length'", callExpr.Function, funcIdent)
				}

				intLit, ok := lit.Elements[2].(*ast.IntegerLiteral)
				if !ok || intLit.Value != 42 {
					t.Fatalf("element 2 = %T (value=%v), want IntegerLiteral 42", lit.Elements[2], intLit)
				}
			},
		},
		{
			name:         "NestedArrays",
			input:        `var matrix := ((1, 2), (3, 4));`,
			wantElements: 2,
			assert: func(t *testing.T, lit *ast.ArrayLiteralExpression) {
				for i, expectedValues := range [][]int64{{1, 2}, {3, 4}} {
					nested, ok := lit.Elements[i].(*ast.ArrayLiteralExpression)
					if !ok {
						t.Fatalf("nested element %d is not *ast.ArrayLiteralExpression, got %T", i, lit.Elements[i])
					}
					if len(nested.Elements) != 2 {
						t.Fatalf("nested element %d length = %d, want 2", i, len(nested.Elements))
					}
					for j, expected := range expectedValues {
						intLit, ok := nested.Elements[j].(*ast.IntegerLiteral)
						if !ok || intLit.Value != expected {
							t.Fatalf("nested element %d[%d] = %T (value=%v), want IntegerLiteral %d", i, j, nested.Elements[j], nested.Elements[j], expected)
						}
					}
				}
			},
		},
		{
			name:         "Empty",
			input:        `var arrEmpty := ();`,
			wantElements: 0,
			assert: func(t *testing.T, lit *ast.ArrayLiteralExpression) {
				if len(lit.Elements) != 0 {
					t.Fatalf("len(Elements) = %d, want 0", len(lit.Elements))
				}
			},
		},
		{
			name:         "TrailingComma",
			input:        `var arr := (1, 2, 3, );`,
			wantElements: 3,
			assert: func(t *testing.T, lit *ast.ArrayLiteralExpression) {
				if len(lit.Elements) != 3 {
					t.Fatalf("len(Elements) = %d, want 3", len(lit.Elements))
				}
			},
		},
		{
			name:         "InConstDeclaration",
			input:        `const arr: array [0..3] of Integer = (10, 20, 30, 40);`,
			wantElements: 4,
			assert: func(t *testing.T, lit *ast.ArrayLiteralExpression) {
				for i, expected := range []int64{10, 20, 30, 40} {
					intLit, ok := lit.Elements[i].(*ast.IntegerLiteral)
					if !ok {
						t.Fatalf("element %d is not *ast.IntegerLiteral, got %T", i, lit.Elements[i])
					}
					if intLit.Value != expected {
						t.Fatalf("element %d value = %d, want %d", i, intLit.Value, expected)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
			}

			var arrayLit *ast.ArrayLiteralExpression
			switch stmt := program.Statements[0].(type) {
			case *ast.VarDeclStatement:
				var ok bool
				arrayLit, ok = stmt.Value.(*ast.ArrayLiteralExpression)
				if !ok {
					t.Fatalf("VarDecl Value is not *ast.ArrayLiteralExpression, got %T", stmt.Value)
				}
			case *ast.ConstDecl:
				var ok bool
				arrayLit, ok = stmt.Value.(*ast.ArrayLiteralExpression)
				if !ok {
					t.Fatalf("ConstDecl Value is not *ast.ArrayLiteralExpression, got %T", stmt.Value)
				}
			default:
				t.Fatalf("statement is not VarDeclStatement or ConstDecl, got %T", program.Statements[0])
			}

			if len(arrayLit.Elements) != tt.wantElements {
				t.Fatalf("len(Elements) = %d, want %d", len(arrayLit.Elements), tt.wantElements)
			}

			if tt.assert != nil {
				tt.assert(t, arrayLit)
			}
		})
	}
}

func TestParenthesizedArrayLiteralEdgeCases(t *testing.T) {
	t.Run("SingleElementIsGroupedExpression", func(t *testing.T) {
		// A single element in parentheses should be treated as a grouped expression,
		// not an array literal with one element
		input := `var x := (42);`

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

		// Should be IntegerLiteral (unwrapped), not ArrayLiteralExpression
		intLit, ok := varDecl.Value.(*ast.IntegerLiteral)
		if !ok {
			t.Fatalf("Value should be *ast.IntegerLiteral (grouped expression unwrapped), got %T", varDecl.Value)
		}
		if intLit.Value != 42 {
			t.Fatalf("value = %d, want 42", intLit.Value)
		}
	})

	t.Run("RecordLiteralStillWorks", func(t *testing.T) {
		// Record literals should still be parsed correctly
		input := `var p := (X: 10, Y: 20);`

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

		recordLit, ok := varDecl.Value.(*ast.RecordLiteralExpression)
		if !ok {
			t.Fatalf("Value should be *ast.RecordLiteralExpression, got %T", varDecl.Value)
		}
		if len(recordLit.Fields) != 2 {
			t.Fatalf("len(Fields) = %d, want 2", len(recordLit.Fields))
		}
	})
}

// ============================================================================
// Multidimensional Array Parser Tests (Comma-Separated Syntax)
// ============================================================================

func TestParseMultiDimensionalArrayTypes(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType string // Expected desugared type string
	}{
		{
			name:         "2D array in const declaration",
			input:        `const X: array[0..1, 0..2] of Integer = [[1, 2, 3], [4, 5, 6]];`,
			expectedType: "array[0..1] of array[0..2] of Integer",
		},
		{
			name:         "3D array in const declaration",
			input:        `const cube: array[1..2, 1..3, 1..4] of Float = [[[0.0]]];`,
			expectedType: "array[1..2] of array[1..3] of array[1..4] of Float",
		},
		{
			name:         "2D array with expression bounds",
			input:        `const DIGITS = 10; const arr: array[0..1, 0..2*DIGITS] of Integer = [[0], [1]];`,
			expectedType: "array[0..1] of array[0..(2 * DIGITS)] of Integer",
		},
		{
			name:         "Nested arrays (already supported)",
			input:        `const matrix: array of array of Integer = [[1, 2], [3, 4]];`,
			expectedType: "array of array of Integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			// Find the const declaration (may not be first if we have preliminary consts like DIGITS)
			var constStmt *ast.ConstDecl
			for _, stmt := range program.Statements {
				if cs, ok := stmt.(*ast.ConstDecl); ok {
					// Skip DIGITS constant
					if cs.Name.Value != "DIGITS" {
						constStmt = cs
						break
					}
				}
			}

			if constStmt == nil {
				t.Fatalf("No const declaration found (other than DIGITS)")
			}

			// Check the type string matches expected desugared form
			if constStmt.Type == nil {
				t.Fatalf("constStmt.Type is nil")
			}

			typeName := constStmt.Type.String()
			if typeName != tt.expectedType {
				t.Errorf("type = %q, want %q", typeName, tt.expectedType)
			}
		})
	}

	t.Run("2D array in type declaration", func(t *testing.T) {
		input := `type TMatrix = array[0..9, 0..9] of Integer;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		// Check what type this actually is
		stmt := program.Statements[0]
		t.Logf("Statement type: %T", stmt)

		// Based on existing tests, type declarations should produce something
		// Let's just verify it parses without errors for now
		// The exact AST structure for type decls will be verified by running the full test suite
	})

	t.Run("3D array in type declaration", func(t *testing.T) {
		input := `type TCube = array[1..3, 1..4, 1..5] of Float;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		// Just verify it parses without errors
		t.Logf("Successfully parsed 3D array type declaration")
	})
}
