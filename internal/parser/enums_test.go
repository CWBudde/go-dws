package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// ============================================================================
// Enum Declaration Parser Tests
// ============================================================================

// Test basic enum declaration parsing
func TestParseEnumDeclaration(t *testing.T) {
	t.Run("Basic enum with implicit values", func(t *testing.T) {
		input := `type TColor = (Red, Green, Blue);`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		enumDecl, ok := program.Statements[0].(*ast.EnumDecl)
		if !ok {
			t.Fatalf("statement is not *ast.EnumDecl, got %T", program.Statements[0])
		}

		if enumDecl.Name.Value != "TColor" {
			t.Errorf("enumDecl.Name.Value = %s, want 'TColor'", enumDecl.Name.Value)
		}

		if len(enumDecl.Values) != 3 {
			t.Fatalf("enumDecl.Values should have 3 elements, got %d", len(enumDecl.Values))
		}

		expectedValues := []string{"Red", "Green", "Blue"}
		for i, expected := range expectedValues {
			if enumDecl.Values[i].Name != expected {
				t.Errorf("enumDecl.Values[%d].Name = %s, want '%s'", i, enumDecl.Values[i].Name, expected)
			}
			if enumDecl.Values[i].Value != nil {
				t.Errorf("enumDecl.Values[%d].Value should be nil (implicit), got %v", i, *enumDecl.Values[i].Value)
			}
		}
	})

	t.Run("Enum with single value", func(t *testing.T) {
		input := `type TBool = (False);`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		enumDecl := program.Statements[0].(*ast.EnumDecl)
		if len(enumDecl.Values) != 1 {
			t.Fatalf("enumDecl.Values should have 1 element, got %d", len(enumDecl.Values))
		}

		if enumDecl.Values[0].Name != "False" {
			t.Errorf("enumDecl.Values[0].Name = %s, want 'False'", enumDecl.Values[0].Name)
		}
	})
}

// Test enum with explicit values
func TestParseEnumWithExplicitValues(t *testing.T) {
	t.Run("All explicit values", func(t *testing.T) {
		input := `type TEnum = (One = 1, Two = 5, Three = 10);`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		enumDecl, ok := program.Statements[0].(*ast.EnumDecl)
		if !ok {
			t.Fatalf("statement is not *ast.EnumDecl, got %T", program.Statements[0])
		}

		expectedValues := map[string]int{
			"One":   1,
			"Two":   5,
			"Three": 10,
		}

		for i, val := range enumDecl.Values {
			expectedVal, ok := expectedValues[val.Name]
			if !ok {
				t.Errorf("unexpected enum value name: %s", val.Name)
				continue
			}

			if val.Value == nil {
				t.Errorf("enumDecl.Values[%d].Value should not be nil", i)
				continue
			}

			if *val.Value != expectedVal {
				t.Errorf("enumDecl.Values[%d].Value = %d, want %d", i, *val.Value, expectedVal)
			}
		}
	})

	t.Run("Mixed implicit and explicit values", func(t *testing.T) {
		input := `type TEnum = (First, Second = 10, Third);`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		enumDecl := program.Statements[0].(*ast.EnumDecl)

		// First - implicit (nil)
		if enumDecl.Values[0].Name != "First" {
			t.Errorf("Values[0].Name = %s, want 'First'", enumDecl.Values[0].Name)
		}
		if enumDecl.Values[0].Value != nil {
			t.Errorf("Values[0].Value should be nil (implicit)")
		}

		// Second - explicit (10)
		if enumDecl.Values[1].Name != "Second" {
			t.Errorf("Values[1].Name = %s, want 'Second'", enumDecl.Values[1].Name)
		}
		if enumDecl.Values[1].Value == nil || *enumDecl.Values[1].Value != 10 {
			t.Errorf("Values[1].Value should be 10")
		}

		// Third - implicit (nil)
		if enumDecl.Values[2].Name != "Third" {
			t.Errorf("Values[2].Name = %s, want 'Third'", enumDecl.Values[2].Name)
		}
		if enumDecl.Values[2].Value != nil {
			t.Errorf("Values[2].Value should be nil (implicit)")
		}
	})

	t.Run("Negative values", func(t *testing.T) {
		input := `type TEnum = (Negative = -1, Zero = 0, Positive = 1);`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		enumDecl := program.Statements[0].(*ast.EnumDecl)

		if enumDecl.Values[0].Value == nil || *enumDecl.Values[0].Value != -1 {
			t.Errorf("Values[0] should be -1")
		}
		if enumDecl.Values[1].Value == nil || *enumDecl.Values[1].Value != 0 {
			t.Errorf("Values[1] should be 0")
		}
		if enumDecl.Values[2].Value == nil || *enumDecl.Values[2].Value != 1 {
			t.Errorf("Values[2] should be 1")
		}
	})
}

// Test scoped enum syntax
func TestParseScopedEnum(t *testing.T) {
	t.Run("Scoped enum with 'enum' keyword", func(t *testing.T) {
		input := `type TEnum = enum (One, Two, Three);`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		enumDecl, ok := program.Statements[0].(*ast.EnumDecl)
		if !ok {
			t.Fatalf("statement is not *ast.EnumDecl, got %T", program.Statements[0])
		}

		if enumDecl.Name.Value != "TEnum" {
			t.Errorf("enumDecl.Name.Value = %s, want 'TEnum'", enumDecl.Name.Value)
		}

		if len(enumDecl.Values) != 3 {
			t.Fatalf("enumDecl.Values should have 3 elements, got %d", len(enumDecl.Values))
		}
	})
}

// Comprehensive parser tests
func TestEnumParserErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Missing closing parenthesis",
			input: `type TColor = (Red, Green;`,
		},
		{
			name:  "Missing comma between values",
			input: `type TColor = (Red Green Blue);`,
		},
		{
			name:  "Empty enum",
			input: `type TColor = ();`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			p.ParseProgram()

			if len(p.errors) == 0 {
				t.Errorf("expected parser errors for invalid input, got none")
			}
		})
	}
}

// Test parsing enum literals in expressions
func TestParseEnumLiteralsInExpressions(t *testing.T) {
	t.Run("Simple enum value assignment", func(t *testing.T) {
		input := `
			type TColor = (Red, Green, Blue);
			var c: TColor;
			c := Red;
		`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		// Should have 3 statements: enum decl, var decl, assignment
		if len(program.Statements) != 3 {
			t.Fatalf("program.Statements should have 3 statements, got %d", len(program.Statements))
		}

		// Third statement should be assignment with identifier "Red"
		assignStmt, ok := program.Statements[2].(*ast.AssignmentStatement)
		if !ok {
			t.Fatalf("statement is not *ast.AssignmentStatement, got %T", program.Statements[2])
		}

		// Right side should be identifier "Red"
		ident, ok := assignStmt.Value.(*ast.Identifier)
		if !ok {
			t.Fatalf("assignment value is not *ast.Identifier, got %T", assignStmt.Value)
		}

		if ident.Value != "Red" {
			t.Errorf("identifier value = %s, want 'Red'", ident.Value)
		}
	})

	t.Run("Scoped enum value reference", func(t *testing.T) {
		input := `
			type TColor = (Red, Green, Blue);
			var c: TColor;
			c := TColor.Red;
		`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		// Third statement should be assignment with member access
		assignStmt, ok := program.Statements[2].(*ast.AssignmentStatement)
		if !ok {
			t.Fatalf("statement is not *ast.AssignmentStatement, got %T", program.Statements[2])
		}

		// Right side should be member access: TColor.Red
		memberAccess, ok := assignStmt.Value.(*ast.MemberAccessExpression)
		if !ok {
			t.Fatalf("assignment value is not *ast.MemberAccessExpression, got %T", assignStmt.Value)
		}

		// Object should be identifier "TColor"
		objectIdent, ok := memberAccess.Object.(*ast.Identifier)
		if !ok {
			t.Fatalf("member object is not *ast.Identifier, got %T", memberAccess.Object)
		}

		if objectIdent.Value != "TColor" {
			t.Errorf("object identifier = %s, want 'TColor'", objectIdent.Value)
		}

		// Member should be identifier "Red"
		if memberAccess.Member.Value != "Red" {
			t.Errorf("member property = %s, want 'Red'", memberAccess.Member.Value)
		}
	})

	t.Run("Enum value in comparison", func(t *testing.T) {
		input := `
			type TColor = (Red, Green);
			var c: TColor;
			if c = Red then
				PrintLn('Red');
		`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		// Should parse without errors - enum value in comparison expression
		if len(program.Statements) != 3 {
			t.Fatalf("program.Statements should have 3 statements, got %d", len(program.Statements))
		}
	})
}

// Test .Name property access (this would be semantic analysis)
// The parser should handle it as normal member access
func TestEnumDotNamePropertyAccess(t *testing.T) {
	t.Run("Enum value .Name access", func(t *testing.T) {
		// Note: The actual .Name functionality is runtime/semantic
		// Parser just needs to handle the syntax
		input := `
			var c: TColor;
			c := Red;
			var name: String;
			name := c.Name;
		`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		// Should parse successfully - .Name is just member access
		if len(program.Statements) != 4 {
			t.Fatalf("program.Statements should have 4 statements, got %d", len(program.Statements))
		}

		// Last statement should be assignment with member access
		assignStmt, ok := program.Statements[3].(*ast.AssignmentStatement)
		if !ok {
			t.Fatalf("statement is not *ast.AssignmentStatement, got %T", program.Statements[3])
		}

		memberAccess, ok := assignStmt.Value.(*ast.MemberAccessExpression)
		if !ok {
			t.Fatalf("assignment value is not *ast.MemberAccessExpression, got %T", assignStmt.Value)
		}

		if memberAccess.Member.Value != "Name" {
			t.Errorf("member = %s, want 'Name'", memberAccess.Member.Value)
		}
	})
}
