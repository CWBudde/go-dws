package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// ============================================================================
// Set Type Declaration Parser Tests
// ============================================================================

// Test basic set type declaration parsing
func TestParseSetDeclaration(t *testing.T) {
	t.Run("Basic set of enum type", func(t *testing.T) {
		input := `type TDays = set of TWeekday;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		setDecl, ok := program.Statements[0].(*ast.SetDecl)
		if !ok {
			t.Fatalf("statement is not *ast.SetDecl, got %T", program.Statements[0])
		}

		if setDecl.Name.Value != "TDays" {
			t.Errorf("setDecl.Name.Value = %s, want 'TDays'", setDecl.Name.Value)
		}

		if setDecl.ElementType.Name != "TWeekday" {
			t.Errorf("setDecl.ElementType.Name = %s, want 'TWeekday'", setDecl.ElementType.Name)
		}
	})

	t.Run("Set type with different element type", func(t *testing.T) {
		input := `type TOptions = set of TOption;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		setDecl := program.Statements[0].(*ast.SetDecl)

		if setDecl.Name.Value != "TOptions" {
			t.Errorf("setDecl.Name.Value = %s, want 'TOptions'", setDecl.Name.Value)
		}

		if setDecl.ElementType.Name != "TOption" {
			t.Errorf("setDecl.ElementType.Name = %s, want 'TOption'", setDecl.ElementType.Name)
		}
	})

	t.Run("String() method", func(t *testing.T) {
		input := `type TDays = set of TWeekday;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		setDecl := program.Statements[0].(*ast.SetDecl)
		str := setDecl.String()

		// Should contain "set of" and the type names
		if str == "" {
			t.Error("String() should not be empty")
		}

		// Verify it contains expected keywords
		expectedParts := []string{"type", "TDays", "set of", "TWeekday"}
		for _, part := range expectedParts {
			if !contains(str, part) {
				t.Errorf("String() = %s, should contain %s", str, part)
			}
		}
	})
}

// Test inline anonymous enum in var declaration
func TestParseInlineSetDeclaration(t *testing.T) {
	t.Run("Var with inline anonymous enum", func(t *testing.T) {
		// This is more complex - set of (Mon, Tue, Wed)
		// For now, we'll test just the named set type
		// Inline anonymous enums will be handled when we improve var declaration parsing
		input := `var s: TMySet;`

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

		if varDecl.Names[0].Value != "s" {
			t.Errorf("varDecl.Names[0].Value = %s, want 's'", varDecl.Names[0].Value)
		}

		if varDecl.Type.Name != "TMySet" {
			t.Errorf("varDecl.Type.Name = %s, want 'TMySet'", varDecl.Type.Name)
		}
	})
}

// ============================================================================
// Set Literal Parser Tests
// ============================================================================

// Test parsing set literals with elements
func TestParseSetLiteral(t *testing.T) {
	t.Run("Set literal with elements", func(t *testing.T) {
		input := `var s := [one, two, three];`

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

		setLit, ok := varDecl.Value.(*ast.SetLiteral)
		if !ok {
			t.Fatalf("varDecl.Value is not *ast.SetLiteral, got %T", varDecl.Value)
		}

		if len(setLit.Elements) != 3 {
			t.Fatalf("setLit.Elements should have 3 elements, got %d", len(setLit.Elements))
		}

		expectedElements := []string{"one", "two", "three"}
		for i, expected := range expectedElements {
			ident, ok := setLit.Elements[i].(*ast.Identifier)
			if !ok {
				t.Errorf("setLit.Elements[%d] is not *ast.Identifier, got %T", i, setLit.Elements[i])
				continue
			}
			if ident.Value != expected {
				t.Errorf("setLit.Elements[%d].Value = %s, want %s", i, ident.Value, expected)
			}
		}
	})

	t.Run("Set literal with single element", func(t *testing.T) {
		input := `var s := [one];`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		varDecl := program.Statements[0].(*ast.VarDeclStatement)
		setLit, ok := varDecl.Value.(*ast.SetLiteral)
		if !ok {
			t.Fatalf("varDecl.Value is not *ast.SetLiteral, got %T", varDecl.Value)
		}

		if len(setLit.Elements) != 1 {
			t.Fatalf("setLit.Elements should have 1 element, got %d", len(setLit.Elements))
		}
	})
}

// Test parsing empty set
func TestParseEmptySet(t *testing.T) {
	t.Run("Empty set literal", func(t *testing.T) {
		input := `var s := [];`

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

		switch lit := varDecl.Value.(type) {
		case *ast.SetLiteral:
			if len(lit.Elements) != 0 {
				t.Errorf("set literal should be empty, got %d elements", len(lit.Elements))
			}
		case *ast.ArrayLiteralExpression:
			if len(lit.Elements) != 0 {
				t.Errorf("array literal should be empty, got %d elements", len(lit.Elements))
			}
		default:
			t.Fatalf("varDecl.Value is not a recognized literal type, got %T", varDecl.Value)
		}
	})
}

// Test parsing set range literals
func TestParseSetRange(t *testing.T) {
	t.Run("Single range", func(t *testing.T) {
		input := `var s := [A..C];`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		varDecl := program.Statements[0].(*ast.VarDeclStatement)
		setLit, ok := varDecl.Value.(*ast.SetLiteral)
		if !ok {
			t.Fatalf("varDecl.Value is not *ast.SetLiteral, got %T", varDecl.Value)
		}

		// Should have one RangeExpression element
		if len(setLit.Elements) != 1 {
			t.Fatalf("setLit.Elements should have 1 element, got %d", len(setLit.Elements))
		}

		rangeExpr, ok := setLit.Elements[0].(*ast.RangeExpression)
		if !ok {
			t.Fatalf("setLit.Elements[0] is not *ast.RangeExpression, got %T", setLit.Elements[0])
		}

		// Check start and end identifiers
		startIdent, ok := rangeExpr.Start.(*ast.Identifier)
		if !ok || startIdent.Value != "A" {
			t.Errorf("rangeExpr.Start should be identifier 'A'")
		}

		endIdent, ok := rangeExpr.RangeEnd.(*ast.Identifier)
		if !ok || endIdent.Value != "C" {
			t.Errorf("rangeExpr.RangeEnd should be identifier 'C'")
		}
	})

	t.Run("Multiple ranges", func(t *testing.T) {
		input := `var s := [A..A, C..C];`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		varDecl := program.Statements[0].(*ast.VarDeclStatement)
		setLit := varDecl.Value.(*ast.SetLiteral)

		// Should have two RangeExpression elements
		if len(setLit.Elements) != 2 {
			t.Fatalf("setLit.Elements should have 2 elements, got %d", len(setLit.Elements))
		}

		for i, elem := range setLit.Elements {
			if _, ok := elem.(*ast.RangeExpression); !ok {
				t.Errorf("setLit.Elements[%d] is not *ast.RangeExpression, got %T", i, elem)
			}
		}
	})

	t.Run("Mixed elements and ranges", func(t *testing.T) {
		input := `var s := [one, three..five, seven];`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		varDecl := program.Statements[0].(*ast.VarDeclStatement)
		setLit := varDecl.Value.(*ast.SetLiteral)

		// Should have 3 elements: identifier, range, identifier
		if len(setLit.Elements) != 3 {
			t.Fatalf("setLit.Elements should have 3 elements, got %d", len(setLit.Elements))
		}

		// First element: identifier 'one'
		if ident, ok := setLit.Elements[0].(*ast.Identifier); !ok || ident.Value != "one" {
			t.Errorf("setLit.Elements[0] should be identifier 'one'")
		}

		// Second element: range 'three..five'
		if _, ok := setLit.Elements[1].(*ast.RangeExpression); !ok {
			t.Errorf("setLit.Elements[1] should be *ast.RangeExpression")
		}

		// Third element: identifier 'seven'
		if ident, ok := setLit.Elements[2].(*ast.Identifier); !ok || ident.Value != "seven" {
			t.Errorf("setLit.Elements[2] should be identifier 'seven'")
		}
	})
}

// ============================================================================
// Set Operators Parser Tests
// ============================================================================

// Test set operations and 'in' operator
func TestParseSetOperators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		operator string
	}{
		{"Union", "var r := s1 + s2;", "+"},
		{"Difference", "var r := s1 - s2;", "-"},
		{"Intersection", "var r := s1 * s2;", "*"},
		{"Membership", "var r := elem in mySet;", "in"},
		{"Equality", "var r := s1 = s2;", "="},
		{"Inequality", "var r := s1 <> s2;", "<>"},
		{"Subset", "var r := s1 <= s2;", "<="},
		{"Superset", "var r := s1 >= s2;", ">="},
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

			binExpr, ok := varDecl.Value.(*ast.BinaryExpression)
			if !ok {
				t.Fatalf("varDecl.Value is not *ast.BinaryExpression, got %T", varDecl.Value)
			}

			if binExpr.Operator != tt.operator {
				t.Errorf("binExpr.Operator = %s, want %s", binExpr.Operator, tt.operator)
			}
		})
	}
}

// Test complex set expressions
func TestParseComplexSetExpressions(t *testing.T) {
	t.Run("Nested operations", func(t *testing.T) {
		input := `var r := (s1 + s2) - s3;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		varDecl := program.Statements[0].(*ast.VarDeclStatement)

		// Should be: (s1 + s2) - s3
		binExpr, ok := varDecl.Value.(*ast.BinaryExpression)
		if !ok {
			t.Fatalf("varDecl.Value is not *ast.BinaryExpression, got %T", varDecl.Value)
		}

		if binExpr.Operator != "-" {
			t.Errorf("outer operator should be '-', got %s", binExpr.Operator)
		}

		// Left side should be either GroupedExpression or BinaryExpression
		// Both are valid depending on how the parser handles precedence
		switch binExpr.Left.(type) {
		case *ast.GroupedExpression, *ast.BinaryExpression:
			// Either is valid
		default:
			t.Errorf("left side should be GroupedExpression or BinaryExpression, got %T", binExpr.Left)
		}
	})

	t.Run("Set literal in operation", func(t *testing.T) {
		input := `var r := s1 + [one, two];`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		varDecl := program.Statements[0].(*ast.VarDeclStatement)
		binExpr := varDecl.Value.(*ast.BinaryExpression)

		// Right side should be set literal
		setLit, ok := binExpr.Right.(*ast.SetLiteral)
		if !ok {
			t.Fatalf("right side should be *ast.SetLiteral, got %T", binExpr.Right)
		}

		if len(setLit.Elements) != 2 {
			t.Errorf("setLit.Elements should have 2 elements, got %d", len(setLit.Elements))
		}
	})
}

// ============================================================================
// Inline Set Type Parser Tests
// ============================================================================

// Test parsing inline set types in variable declarations
func TestParseInlineSetType(t *testing.T) {
	t.Run("Variable with inline set type", func(t *testing.T) {
		input := `var sieve : set of TRange;`

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

		if len(varDecl.Names) != 1 {
			t.Fatalf("varDecl should have 1 name, got %d", len(varDecl.Names))
		}

		if varDecl.Names[0].Value != "sieve" {
			t.Errorf("varDecl.Names[0].Value = %s, want 'sieve'", varDecl.Names[0].Value)
		}

		// Type should be a TypeAnnotation with InlineType set to SetTypeNode
		if varDecl.Type == nil {
			t.Fatalf("varDecl.Type should not be nil")
		}

		if varDecl.Type.Name != "set of TRange" {
			t.Errorf("varDecl.Type.Name = %s, want 'set of TRange'", varDecl.Type.Name)
		}

		// Check InlineType
		setType, ok := varDecl.Type.InlineType.(*ast.SetTypeNode)
		if !ok {
			t.Fatalf("varDecl.Type.InlineType is not *ast.SetTypeNode, got %T", varDecl.Type.InlineType)
		}

		// Check element type
		elemType, ok := setType.ElementType.(*ast.TypeAnnotation)
		if !ok {
			t.Fatalf("setType.ElementType is not *ast.TypeAnnotation, got %T", setType.ElementType)
		}

		if elemType.Name != "TRange" {
			t.Errorf("elemType.Name = %s, want 'TRange'", elemType.Name)
		}
	})

	t.Run("Multiple variables with inline set type", func(t *testing.T) {
		input := `var s1, s2 : set of TEnum;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		varDecl := program.Statements[0].(*ast.VarDeclStatement)

		if len(varDecl.Names) != 2 {
			t.Fatalf("varDecl should have 2 names, got %d", len(varDecl.Names))
		}

		if varDecl.Type.Name != "set of TEnum" {
			t.Errorf("varDecl.Type.Name = %s, want 'set of TEnum'", varDecl.Type.Name)
		}
	})

	t.Run("Function parameter with inline set type", func(t *testing.T) {
		input := `function Test(s: set of TEnum): Boolean; begin end;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		funcDecl, ok := program.Statements[0].(*ast.FunctionDecl)
		if !ok {
			t.Fatalf("statement is not *ast.FunctionDecl, got %T", program.Statements[0])
		}

		if len(funcDecl.Parameters) != 1 {
			t.Fatalf("function should have 1 parameter, got %d", len(funcDecl.Parameters))
		}

		param := funcDecl.Parameters[0]
		if param.Name.Value != "s" {
			t.Errorf("parameter name = %s, want 's'", param.Name.Value)
		}

		if param.Type.Name != "set of TEnum" {
			t.Errorf("parameter type = %s, want 'set of TEnum'", param.Type.Name)
		}
	})

	t.Run("Function return type with inline set type", func(t *testing.T) {
		input := `function GetSet(): set of TEnum; begin end;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		funcDecl, ok := program.Statements[0].(*ast.FunctionDecl)
		if !ok {
			t.Fatalf("statement is not *ast.FunctionDecl, got %T", program.Statements[0])
		}

		if funcDecl.ReturnType.Name != "set of TEnum" {
			t.Errorf("return type = %s, want 'set of TEnum'", funcDecl.ReturnType.Name)
		}
	})

	t.Run("String() method for SetTypeNode", func(t *testing.T) {
		input := `var s : set of TEnum;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		varDecl := program.Statements[0].(*ast.VarDeclStatement)
		setType := varDecl.Type.InlineType.(*ast.SetTypeNode)

		str := setType.String()
		if str != "set of TEnum" {
			t.Errorf("SetTypeNode.String() = %s, want 'set of TEnum'", str)
		}
	})
}
