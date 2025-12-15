package parser

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Array Type Declaration Parser Tests
// ============================================================================

func TestParseArrayTypeDeclaration(t *testing.T) {
	tests := []struct {
		checkStmt func(t *testing.T, stmt ast.Statement)
		name      string
		input     string
	}{
		{
			name:  "Static array with bounds",
			input: `type TMyArray = array[1..10] of Integer;`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				// For now, array type declarations might be parsed differently
				// We need to determine the AST structure for type declarations
				t.Logf("Statement type: %T", stmt)
			},
		},
		{
			name:  "Dynamic array without bounds",
			input: `type TStringArray = array of String;`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				// No extra checks needed beyond successful parsing
			},
		},
		{
			name:  "Array of custom type",
			input: `type TPersonArray = array[0..99] of TPerson;`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				// No extra checks needed beyond successful parsing
			},
		},
		{
			name:  "Nested dynamic arrays - 2D",
			input: `type Matrix = array of array of Float;`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				arrayDecl, ok := stmt.(*ast.ArrayDecl)
				if !ok {
					t.Fatalf("statement is not *ast.ArrayDecl, got %T", stmt)
				}

				if arrayDecl.Name.Value != "Matrix" {
					t.Errorf("arrayDecl.Name.Value = %s, want 'Matrix'", arrayDecl.Name.Value)
				}

				expectedType := "array of Float"
				if arrayDecl.ArrayType.ElementType.String() != expectedType {
					t.Errorf("arrayDecl.ArrayType.ElementType.String() = %s, want %s",
						arrayDecl.ArrayType.ElementType.String(), expectedType)
				}
			},
		},
		{
			name:  "Nested dynamic arrays - 3D",
			input: `type Tensor = array of array of array of Integer;`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				arrayDecl, ok := stmt.(*ast.ArrayDecl)
				if !ok {
					t.Fatalf("statement is not *ast.ArrayDecl, got %T", stmt)
				}

				if arrayDecl.Name.Value != "Tensor" {
					t.Errorf("arrayDecl.Name.Value = %s, want 'Tensor'", arrayDecl.Name.Value)
				}

				expectedType := "array of array of Integer"
				if arrayDecl.ArrayType.ElementType.String() != expectedType {
					t.Errorf("arrayDecl.ArrayType.ElementType.String() = %s, want %s",
						arrayDecl.ArrayType.ElementType.String(), expectedType)
				}
			},
		},
		{
			name:  "Nested static arrays",
			input: `type Grid = array[1..10] of array[1..20] of Boolean;`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				arrayDecl, ok := stmt.(*ast.ArrayDecl)
				if !ok {
					t.Fatalf("statement is not *ast.ArrayDecl, got %T", stmt)
				}

				if arrayDecl.Name.Value != "Grid" {
					t.Errorf("arrayDecl.Name.Value = %s, want 'Grid'", arrayDecl.Name.Value)
				}

				expectedType := "array[1..20] of Boolean"
				if arrayDecl.ArrayType.ElementType.String() != expectedType {
					t.Errorf("arrayDecl.ArrayType.ElementType.String() = %s, want %s",
						arrayDecl.ArrayType.ElementType.String(), expectedType)
				}
			},
		},
		{
			name:  "Mixed static and dynamic nested arrays",
			input: `type MixedArray = array of array[0..99] of String;`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				arrayDecl, ok := stmt.(*ast.ArrayDecl)
				if !ok {
					t.Fatalf("statement is not *ast.ArrayDecl, got %T", stmt)
				}

				if arrayDecl.Name.Value != "MixedArray" {
					t.Errorf("arrayDecl.Name.Value = %s, want 'MixedArray'", arrayDecl.Name.Value)
				}

				expectedType := "array[0..99] of String"
				if arrayDecl.ArrayType.ElementType.String() != expectedType {
					t.Errorf("arrayDecl.ArrayType.ElementType.String() = %s, want %s",
						arrayDecl.ArrayType.ElementType.String(), expectedType)
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
			stmt := helperParseAndGetStatement(t, tt.input)

			varDecl, ok := stmt.(*ast.VarDeclStatement)
			if !ok {
				t.Fatalf("statement is not *ast.VarDeclStatement, got %T", stmt)
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
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name:          "MissingCommaBetweenElements",
			input:         `var arr := [1 2, 3];`,
			expectedError: "expected ',' or ']'",
		},
		{
			name:          "UnclosedBracket",
			input:         `var arr := [1, 2, 3;`,
			expectedError: "expected ',' or ']'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			_ = p.ParseProgram()

			if len(p.Errors()) == 0 {
				t.Fatalf("expected parser errors, got none")
			}

			found := false
			for _, err := range p.Errors() {
				// Using strings.Contains instead of 'contains' helper
				if strings.Contains(err.Error(), tt.expectedError) {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("expected error containing %q, got: %v", tt.expectedError, p.Errors())
			}
		})
	}
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
			stmt := helperParseAndGetStatement(t, tt.input)

			var arrayLit *ast.ArrayLiteralExpression
			switch s := stmt.(type) {
			case *ast.VarDeclStatement:
				var ok bool
				arrayLit, ok = s.Value.(*ast.ArrayLiteralExpression)
				if !ok {
					t.Fatalf("VarDecl Value is not *ast.ArrayLiteralExpression, got %T", s.Value)
				}
			case *ast.ConstDecl:
				var ok bool
				arrayLit, ok = s.Value.(*ast.ArrayLiteralExpression)
				if !ok {
					t.Fatalf("ConstDecl Value is not *ast.ArrayLiteralExpression, got %T", s.Value)
				}
			default:
				t.Fatalf("statement is not VarDeclStatement or ConstDecl, got %T", stmt)
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
	tests := []struct {
		checkStmt func(t *testing.T, stmt ast.Statement)
		name      string
		input     string
	}{
		{
			name:  "SingleElementIsGroupedExpression",
			input: `var x := (42);`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				varDecl, ok := stmt.(*ast.VarDeclStatement)
				if !ok {
					t.Fatalf("statement is not *ast.VarDeclStatement, got %T", stmt)
				}

				// Should be IntegerLiteral (unwrapped), not ArrayLiteralExpression
				intLit, ok := varDecl.Value.(*ast.IntegerLiteral)
				if !ok {
					t.Fatalf("Value should be *ast.IntegerLiteral (grouped expression unwrapped), got %T", varDecl.Value)
				}
				if intLit.Value != 42 {
					t.Fatalf("value = %d, want 42", intLit.Value)
				}
			},
		},
		{
			name:  "RecordLiteralStillWorks",
			input: `var p := (X: 10, Y: 20);`,
			checkStmt: func(t *testing.T, stmt ast.Statement) {
				varDecl, ok := stmt.(*ast.VarDeclStatement)
				if !ok {
					t.Fatalf("statement is not *ast.VarDeclStatement, got %T", stmt)
				}

				recordLit, ok := varDecl.Value.(*ast.RecordLiteralExpression)
				if !ok {
					t.Fatalf("Value should be *ast.RecordLiteralExpression, got %T", varDecl.Value)
				}
				if len(recordLit.Fields) != 2 {
					t.Fatalf("len(Fields) = %d, want 2", len(recordLit.Fields))
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
// Multidimensional Array Parser Tests (Comma-Separated Syntax)
// ============================================================================

func TestParseMultiDimensionalArrayTypes(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType string // Expected desugared type string. If empty, just verify parsing.
		isConstDecl  bool   // true if checking ConstDecl, false if just checking parsing (e.g. TypeDecl)
	}{
		{
			name:         "2D array in const declaration",
			input:        `const X: array[0..1, 0..2] of Integer = [[1, 2, 3], [4, 5, 6]];`,
			expectedType: "array[0..1] of array[0..2] of Integer",
			isConstDecl:  true,
		},
		{
			name:         "3D array in const declaration",
			input:        `const cube: array[1..2, 1..3, 1..4] of Float = [[[0.0]]];`,
			expectedType: "array[1..2] of array[1..3] of array[1..4] of Float",
			isConstDecl:  true,
		},
		{
			name:         "2D array with expression bounds",
			input:        `const DIGITS = 10; const arr: array[0..1, 0..2*DIGITS] of Integer = [[0], [1]];`,
			expectedType: "array[0..1] of array[0..(2 * DIGITS)] of Integer",
			isConstDecl:  true,
		},
		{
			name:         "Nested arrays (already supported)",
			input:        `const matrix: array of array of Integer = [[1, 2], [3, 4]];`,
			expectedType: "array of array of Integer",
			isConstDecl:  true,
		},
		{
			name:        "2D array in type declaration",
			input:       `type TMatrix = array[0..9, 0..9] of Integer;`,
			isConstDecl: false,
		},
		{
			name:        "3D array in type declaration",
			input:       `type TCube = array[1..3, 1..4, 1..5] of Float;`,
			isConstDecl: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Special handling for DIGITS case where we have multiple statements
			var program *ast.Program
			if strings.Contains(tt.input, "const DIGITS = 10;") {
				l := lexer.New(tt.input)
				p := New(l)
				program = p.ParseProgram()
				checkParserErrors(t, p)
			} else {
				// For normal cases, use helper which asserts 1 statement
				stmt := helperParseAndGetStatement(t, tt.input)
				program = &ast.Program{Statements: []ast.Statement{stmt}}
			}

			if tt.isConstDecl {
				// Find the const declaration
				var constStmt *ast.ConstDecl
				for _, stmt := range program.Statements {
					if cs, ok := stmt.(*ast.ConstDecl); ok {
						// Skip DIGITS constant if present
						if cs.Name.Value != "DIGITS" {
							constStmt = cs
							break
						}
					}
				}

				if constStmt == nil {
					t.Fatalf("No const declaration found (other than DIGITS)")
				}

				if constStmt.Type == nil {
					t.Fatalf("constStmt.Type is nil")
				}

				typeName := constStmt.Type.String()
				if typeName != tt.expectedType {
					t.Errorf("type = %q, want %q", typeName, tt.expectedType)
				}
			} else {
				// Just verify it parsed (already done by helperParseAndGetStatement)
				// and maybe log the statement type
				t.Logf("Successfully parsed: %T", program.Statements[0])
			}
		})
	}
}

// helperParseAndGetStatement parses a program and returns the first statement.
// It asserts that there are no errors and exactly one statement.
func helperParseAndGetStatement(t *testing.T, input string) ast.Statement {
	t.Helper()
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
	}

	return program.Statements[0]
}
