package parser

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

func TestParseConstDeclaration(t *testing.T) {
	input := `const PI = 3.14;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ConstDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ConstDecl. got=%T",
			program.Statements[0])
	}

	if stmt.Name.Value != "PI" {
		t.Errorf("stmt.Name.Value not 'PI'. got=%s", stmt.Name.Value)
	}

	if stmt.Type != nil {
		t.Errorf("stmt.Type should be nil for untyped const. got=%v", stmt.Type)
	}

	floatLit, ok := stmt.Value.(*ast.FloatLiteral)
	if !ok {
		t.Fatalf("stmt.Value is not *ast.FloatLiteral. got=%T", stmt.Value)
	}

	if floatLit.Value != 3.14 {
		t.Errorf("floatLit.Value not 3.14. got=%f", floatLit.Value)
	}
}

func TestParseConstDeclarationTyped(t *testing.T) {
	input := `const MAX_USERS: Integer = 1000;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ConstDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ConstDecl. got=%T",
			program.Statements[0])
	}

	if stmt.Name.Value != "MAX_USERS" {
		t.Errorf("stmt.Name.Value not 'MAX_USERS'. got=%s", stmt.Name.Value)
	}

	if stmt.Type == nil {
		t.Fatal("stmt.Type should not be nil for typed const")
	}

	if stmt.Type.Name != "Integer" {
		t.Errorf("stmt.Type.Name not 'Integer'. got=%s", stmt.Type.Name)
	}

	intLit, ok := stmt.Value.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("stmt.Value is not *ast.IntegerLiteral. got=%T", stmt.Value)
	}

	if intLit.Value != 1000 {
		t.Errorf("intLit.Value not 1000. got=%d", intLit.Value)
	}
}

func TestParseConstDeclarationString(t *testing.T) {
	input := `const APP_NAME = 'MyApp';`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ConstDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ConstDecl. got=%T",
			program.Statements[0])
	}

	if stmt.Name.Value != "APP_NAME" {
		t.Errorf("stmt.Name.Value not 'APP_NAME'. got=%s", stmt.Name.Value)
	}

	stringLit, ok := stmt.Value.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("stmt.Value is not *ast.StringLiteral. got=%T", stmt.Value)
	}

	if stringLit.Value != "MyApp" {
		t.Errorf("stringLit.Value not 'MyApp'. got=%s", stringLit.Value)
	}
}

func TestParseMultipleConstDeclarations(t *testing.T) {
	input := `
const PI = 3.14;
const MAX = 100;
const NAME = 'test';
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d",
			len(program.Statements))
	}

	tests := []struct {
		expectedValue interface{}
		expectedName  string
	}{
		{expectedValue: 3.14, expectedName: "PI"},
		{expectedValue: int64(100), expectedName: "MAX"},
		{expectedValue: "test", expectedName: "NAME"},
	}

	for i, tt := range tests {
		stmt, ok := program.Statements[i].(*ast.ConstDecl)
		if !ok {
			t.Fatalf("program.Statements[%d] is not *ast.ConstDecl. got=%T",
				i, program.Statements[i])
		}

		if stmt.Name.Value != tt.expectedName {
			t.Errorf("stmt.Name.Value not '%s'. got=%s", tt.expectedName, stmt.Name.Value)
		}

		switch v := tt.expectedValue.(type) {
		case float64:
			floatLit, ok := stmt.Value.(*ast.FloatLiteral)
			if !ok {
				t.Fatalf("stmt.Value is not *ast.FloatLiteral. got=%T", stmt.Value)
			}
			if floatLit.Value != v {
				t.Errorf("floatLit.Value not %f. got=%f", v, floatLit.Value)
			}
		case int64:
			intLit, ok := stmt.Value.(*ast.IntegerLiteral)
			if !ok {
				t.Fatalf("stmt.Value is not *ast.IntegerLiteral. got=%T", stmt.Value)
			}
			if intLit.Value != v {
				t.Errorf("intLit.Value not %d. got=%d", v, intLit.Value)
			}
		case string:
			stringLit, ok := stmt.Value.(*ast.StringLiteral)
			if !ok {
				t.Fatalf("stmt.Value is not *ast.StringLiteral. got=%T", stmt.Value)
			}
			if stringLit.Value != v {
				t.Errorf("stringLit.Value not %s. got=%s", v, stringLit.Value)
			}
		}
	}
}

func TestParseConstDeclarationWithAssign(t *testing.T) {
	input := `const PI := 3.14;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ConstDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ConstDecl. got=%T",
			program.Statements[0])
	}

	if stmt.Name.Value != "PI" {
		t.Errorf("stmt.Name.Value not 'PI'. got=%s", stmt.Name.Value)
	}

	floatLit, ok := stmt.Value.(*ast.FloatLiteral)
	if !ok {
		t.Fatalf("stmt.Value is not *ast.FloatLiteral. got=%T", stmt.Value)
	}

	if floatLit.Value != 3.14 {
		t.Errorf("floatLit.Value not 3.14. got=%f", floatLit.Value)
	}
}

func TestParseConstDeclarationBothSyntaxes(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"with equals", `const PI = 3.14;`},
		{"with assign", `const PI := 3.14;`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program.Statements does not contain 1 statement. got=%d",
					len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.ConstDecl)
			if !ok {
				t.Fatalf("program.Statements[0] is not *ast.ConstDecl. got=%T",
					program.Statements[0])
			}

			if stmt.Name.Value != "PI" {
				t.Errorf("stmt.Name.Value not 'PI'. got=%s", stmt.Name.Value)
			}

			floatLit, ok := stmt.Value.(*ast.FloatLiteral)
			if !ok {
				t.Fatalf("stmt.Value is not *ast.FloatLiteral. got=%T", stmt.Value)
			}

			if floatLit.Value != 3.14 {
				t.Errorf("floatLit.Value not 3.14. got=%f", floatLit.Value)
			}
		})
	}
}

func TestParseConstDeclarationErrors(t *testing.T) {
	tests := []struct {
		input         string
		expectedError string
	}{
		{"const PI;", "expected '=' or ':=' after const name"},
		{"const PI =;", "no prefix parse function for SEMICOLON found"},
		{"const = 3.14;", "expected next token to be IDENT"},
		{"const PI: = 3.14;", "expected type expression after ':' in const declaration"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		p.ParseProgram()

		if len(p.errors) == 0 {
			t.Errorf("expected parser error for input %q, but got none", tt.input)
			continue
		}

		found := false
		for _, err := range p.errors {
			if strings.Contains(err, tt.expectedError) {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("expected error containing %q, got errors: %v", tt.expectedError, p.errors)
		}
	}
}

// ============================================================================
// Type Alias Tests
// ============================================================================

// TestParseTypeAlias tests parsing simple type alias declarations
func TestParseTypeAlias(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedName string
		expectedType string
	}{
		{
			name:         "Integer alias",
			input:        `type TUserID = Integer;`,
			expectedName: "TUserID",
			expectedType: "Integer",
		},
		{
			name:         "String alias",
			input:        `type TFileName = String;`,
			expectedName: "TFileName",
			expectedType: "String",
		},
		{
			name:         "Float alias",
			input:        `type TPrice = Float;`,
			expectedName: "TPrice",
			expectedType: "Float",
		},
		{
			name:         "Boolean alias",
			input:        `type TFlag = Boolean;`,
			expectedName: "TFlag",
			expectedType: "Boolean",
		},
		{
			name:         "Nested alias (alias to another alias)",
			input:        `type TMyInt = TUserID;`,
			expectedName: "TMyInt",
			expectedType: "TUserID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program.Statements does not contain 1 statement. got=%d",
					len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.TypeDeclaration)
			if !ok {
				t.Fatalf("program.Statements[0] is not *ast.TypeDeclaration. got=%T",
					program.Statements[0])
			}

			if stmt.Name.Value != tt.expectedName {
				t.Errorf("stmt.Name.Value not %q. got=%s", tt.expectedName, stmt.Name.Value)
			}

			if !stmt.IsAlias {
				t.Error("stmt.IsAlias should be true for type alias")
			}

			if stmt.AliasedType == nil {
				t.Fatal("stmt.AliasedType should not be nil")
			}

			if stmt.AliasedType.Name != tt.expectedType {
				t.Errorf("stmt.AliasedType.Name not %q. got=%s", tt.expectedType, stmt.AliasedType.Name)
			}
		})
	}
}

// TestParseMultipleTypeAliases tests parsing multiple type alias declarations in one program
func TestParseMultipleTypeAliases(t *testing.T) {
	input := `
		type TUserID = Integer;
		type TFileName = String;
		type TPrice = Float;
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d",
			len(program.Statements))
	}

	expectedAliases := []struct {
		name string
		typ  string
	}{
		{"TUserID", "Integer"},
		{"TFileName", "String"},
		{"TPrice", "Float"},
	}

	for i, expected := range expectedAliases {
		stmt, ok := program.Statements[i].(*ast.TypeDeclaration)
		if !ok {
			t.Fatalf("program.Statements[%d] is not *ast.TypeDeclaration. got=%T",
				i, program.Statements[i])
		}

		if stmt.Name.Value != expected.name {
			t.Errorf("stmt.Name.Value not %q. got=%s", expected.name, stmt.Name.Value)
		}

		if !stmt.IsAlias {
			t.Errorf("stmt.IsAlias should be true for type alias %s", expected.name)
		}

		if stmt.AliasedType.Name != expected.typ {
			t.Errorf("stmt.AliasedType.Name not %q. got=%s", expected.typ, stmt.AliasedType.Name)
		}
	}
}

// ============================================================================
// Subrange Type Declaration Tests
// ============================================================================

// TestParseSubrangeType tests parsing subrange type declarations
func TestParseSubrangeType(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedTypeName string
		expectedLowStr   string
		expectedHighStr  string
	}{
		{
			name:             "Basic digit subrange (0..9)",
			input:            `type TDigit = 0..9;`,
			expectedTypeName: "TDigit",
			expectedLowStr:   "0",
			expectedHighStr:  "9",
		},
		{
			name:             "Percentage subrange (0..100)",
			input:            `type TPercent = 0..100;`,
			expectedTypeName: "TPercent",
			expectedLowStr:   "0",
			expectedHighStr:  "100",
		},
		{
			name:             "Negative range (-40..50)",
			input:            `type TTemperature = -40..50;`,
			expectedTypeName: "TTemperature",
			expectedLowStr:   "(-40)",
			expectedHighStr:  "50",
		},
		{
			name:             "Single value range (42..42)",
			input:            `type TAnswer = 42..42;`,
			expectedTypeName: "TAnswer",
			expectedLowStr:   "42",
			expectedHighStr:  "42",
		},
		{
			name:             "Large range (1..1000)",
			input:            `type TIndex = 1..1000;`,
			expectedTypeName: "TIndex",
			expectedLowStr:   "1",
			expectedHighStr:  "1000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program.Statements does not contain 1 statement. got=%d",
					len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.TypeDeclaration)
			if !ok {
				t.Fatalf("program.Statements[0] is not *ast.TypeDeclaration. got=%T",
					program.Statements[0])
			}

			// Verify type name
			if stmt.Name.Value != tt.expectedTypeName {
				t.Errorf("stmt.Name.Value not %q. got=%s", tt.expectedTypeName, stmt.Name.Value)
			}

			// Verify it's marked as subrange
			if !stmt.IsSubrange {
				t.Error("stmt.IsSubrange should be true for subrange type")
			}

			// Verify it's NOT marked as alias
			if stmt.IsAlias {
				t.Error("stmt.IsAlias should be false for subrange type")
			}

			// Verify low bound exists
			if stmt.LowBound == nil {
				t.Fatal("stmt.LowBound should not be nil")
			}

			// Verify high bound exists
			if stmt.HighBound == nil {
				t.Fatal("stmt.HighBound should not be nil")
			}

			// Verify String() output format
			expectedString := "type " + tt.expectedTypeName + " = " + tt.expectedLowStr + ".." + tt.expectedHighStr
			if stmt.String() != expectedString {
				t.Errorf("stmt.String() = %q, want %q", stmt.String(), expectedString)
			}

			// Verify low bound string representation
			if stmt.LowBound.String() != tt.expectedLowStr {
				t.Errorf("stmt.LowBound.String() = %q, want %q", stmt.LowBound.String(), tt.expectedLowStr)
			}

			// Verify high bound string representation
			if stmt.HighBound.String() != tt.expectedHighStr {
				t.Errorf("stmt.HighBound.String() = %q, want %q", stmt.HighBound.String(), tt.expectedHighStr)
			}
		})
	}
}

// TestParseMultipleSubrangeTypes tests parsing multiple subrange declarations
func TestParseMultipleSubrangeTypes(t *testing.T) {
	input := `
		type TDigit = 0..9;
		type TPercent = 0..100;
		type TDay = 1..31;
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements should contain 3 statements. got=%d",
			len(program.Statements))
	}

	expected := []struct {
		name    string
		lowStr  string
		highStr string
	}{
		{"TDigit", "0", "9"},
		{"TPercent", "0", "100"},
		{"TDay", "1", "31"},
	}

	for i, exp := range expected {
		stmt, ok := program.Statements[i].(*ast.TypeDeclaration)
		if !ok {
			t.Fatalf("program.Statements[%d] is not *ast.TypeDeclaration. got=%T",
				i, program.Statements[i])
		}

		if stmt.Name.Value != exp.name {
			t.Errorf("stmt[%d].Name.Value not %q. got=%s", i, exp.name, stmt.Name.Value)
		}

		if !stmt.IsSubrange {
			t.Errorf("stmt[%d].IsSubrange should be true", i)
		}

		if stmt.LowBound.String() != exp.lowStr {
			t.Errorf("stmt[%d].LowBound.String() = %q, want %q", i, stmt.LowBound.String(), exp.lowStr)
		}

		if stmt.HighBound.String() != exp.highStr {
			t.Errorf("stmt[%d].HighBound.String() = %q, want %q", i, stmt.HighBound.String(), exp.highStr)
		}
	}
}

// TestParseSubrangeErrors tests error cases in subrange parsing
func TestParseSubrangeErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Missing DOTDOT token",
			input: `type TDigit = 0 9;`,
		},
		{
			name:  "Missing semicolon",
			input: `type TDigit = 0..9`,
		},
		{
			name:  "Missing high bound",
			input: `type TDigit = 0..;`,
		},
		{
			name:  "Missing low bound",
			input: `type TDigit = ..9;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			p.ParseProgram()

			// Should have parser errors
			if len(p.errors) == 0 {
				t.Errorf("expected parser errors for invalid subrange syntax, got none")
			}
		})
	}
}

// TestParseMixedTypeDeclarations tests parsing both aliases and subranges
func TestParseMixedTypeDeclarations(t *testing.T) {
	input := `
		type TUserID = Integer;
		type TDigit = 0..9;
		type TFileName = String;
		type TPercent = 0..100;
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 4 {
		t.Fatalf("program.Statements should contain 4 statements. got=%d",
			len(program.Statements))
	}

	// First: type alias
	stmt0, ok := program.Statements[0].(*ast.TypeDeclaration)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.TypeDeclaration")
	}
	if !stmt0.IsAlias {
		t.Error("First declaration should be alias")
	}
	if stmt0.IsSubrange {
		t.Error("First declaration should not be subrange")
	}

	// Second: subrange
	stmt1, ok := program.Statements[1].(*ast.TypeDeclaration)
	if !ok {
		t.Fatalf("program.Statements[1] is not *ast.TypeDeclaration")
	}
	if stmt1.IsAlias {
		t.Error("Second declaration should not be alias")
	}
	if !stmt1.IsSubrange {
		t.Error("Second declaration should be subrange")
	}

	// Third: type alias
	stmt2, ok := program.Statements[2].(*ast.TypeDeclaration)
	if !ok {
		t.Fatalf("program.Statements[2] is not *ast.TypeDeclaration")
	}
	if !stmt2.IsAlias {
		t.Error("Third declaration should be alias")
	}

	// Fourth: subrange
	stmt3, ok := program.Statements[3].(*ast.TypeDeclaration)
	if !ok {
		t.Fatalf("program.Statements[3] is not *ast.TypeDeclaration")
	}
	if !stmt3.IsSubrange {
		t.Error("Fourth declaration should be subrange")
	}
}

// ============================================================================
// Const Declaration with Array Types Tests
// ============================================================================

// TestParseConstDeclarationWithArrayType tests parsing const declarations with array types
func TestParseConstDeclarationWithArrayType(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedConstName string
		expectedTypeName  string
	}{
		{
			name:              "Dynamic array of integers",
			input:             `const arr: array of Integer = [1, 2, 3];`,
			expectedConstName: "arr",
			expectedTypeName:  "array of Integer",
		},
		{
			name:              "Static array with bounds",
			input:             `const good: array [0..13] of Integer = [1600,1660,1724];`,
			expectedConstName: "good",
			expectedTypeName:  "array[0..13] of Integer",
		},
		{
			name:              "Array with negative bounds",
			input:             `const temps: array [-10..10] of Float = [0.0];`,
			expectedConstName: "temps",
			expectedTypeName:  "array[(-10)..10] of Float",
		},
		{
			name:              "Nested arrays",
			input:             `const matrix: array of array of Integer = [[1, 2], [3, 4]];`,
			expectedConstName: "matrix",
			expectedTypeName:  "array of array of Integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program.Statements does not contain 1 statement. got=%d",
					len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.ConstDecl)
			if !ok {
				t.Fatalf("program.Statements[0] is not *ast.ConstDecl. got=%T",
					program.Statements[0])
			}

			if stmt.Name.Value != tt.expectedConstName {
				t.Errorf("stmt.Name.Value not %q. got=%s", tt.expectedConstName, stmt.Name.Value)
			}

			if stmt.Type == nil {
				t.Fatal("stmt.Type should not be nil for typed const")
			}

			if stmt.Type.Name != tt.expectedTypeName {
				t.Errorf("stmt.Type.Name not %q. got=%s", tt.expectedTypeName, stmt.Type.Name)
			}

			// Verify value is an array literal
			_, ok = stmt.Value.(*ast.ArrayLiteralExpression)
			if !ok {
				t.Fatalf("stmt.Value is not *ast.ArrayLiteralExpression. got=%T", stmt.Value)
			}
		})
	}
}

// TestParseConstDeclarationLeapYear tests the specific failing case from Leap_year.dws
func TestParseConstDeclarationLeapYear(t *testing.T) {
	input := `
const good : array [0..13] of Integer =
   [1600,1660,1724,1788,1848,1912,1972,2032,2092,2156,2220,2280,2344,2348];
const bad : array [0..13] of Integer =
   [1698,1699,1700,1750,1800,1810,1900,1901,1973,2100,2107,2200,2203,2289];
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 2 {
		t.Fatalf("program.Statements does not contain 2 statements. got=%d",
			len(program.Statements))
	}

	// Test first const (good)
	stmt1, ok := program.Statements[0].(*ast.ConstDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ConstDecl. got=%T",
			program.Statements[0])
	}

	if stmt1.Name.Value != "good" {
		t.Errorf("stmt1.Name.Value not 'good'. got=%s", stmt1.Name.Value)
	}

	if stmt1.Type == nil {
		t.Fatal("stmt1.Type should not be nil")
	}

	if stmt1.Type.Name != "array[0..13] of Integer" {
		t.Errorf("stmt1.Type.Name not 'array[0..13] of Integer'. got=%s", stmt1.Type.Name)
	}

	arr1, ok := stmt1.Value.(*ast.ArrayLiteralExpression)
	if !ok {
		t.Fatalf("stmt1.Value is not *ast.ArrayLiteralExpression. got=%T", stmt1.Value)
	}

	if len(arr1.Elements) != 14 {
		t.Errorf("arr1.Elements should have 14 elements. got=%d", len(arr1.Elements))
	}

	// Test second const (bad)
	stmt2, ok := program.Statements[1].(*ast.ConstDecl)
	if !ok {
		t.Fatalf("program.Statements[1] is not *ast.ConstDecl. got=%T",
			program.Statements[1])
	}

	if stmt2.Name.Value != "bad" {
		t.Errorf("stmt2.Name.Value not 'bad'. got=%s", stmt2.Name.Value)
	}

	if stmt2.Type == nil {
		t.Fatal("stmt2.Type should not be nil")
	}

	if stmt2.Type.Name != "array[0..13] of Integer" {
		t.Errorf("stmt2.Type.Name not 'array[0..13] of Integer'. got=%s", stmt2.Type.Name)
	}

	arr2, ok := stmt2.Value.(*ast.ArrayLiteralExpression)
	if !ok {
		t.Fatalf("stmt2.Value is not *ast.ArrayLiteralExpression. got=%T", stmt2.Value)
	}

	if len(arr2.Elements) != 14 {
		t.Errorf("arr2.Elements should have 14 elements. got=%d", len(arr2.Elements))
	}
}

// TestParseConstDeclarationWithFunctionPointerType tests const with function pointer types
func TestParseConstDeclarationWithFunctionPointerType(t *testing.T) {
	input := `const callback: function(x: Integer): Boolean = nil;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ConstDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ConstDecl. got=%T",
			program.Statements[0])
	}

	if stmt.Name.Value != "callback" {
		t.Errorf("stmt.Name.Value not 'callback'. got=%s", stmt.Name.Value)
	}

	if stmt.Type == nil {
		t.Fatal("stmt.Type should not be nil")
	}

	// Function pointer type is stored as string representation
	expectedType := "function(x: Integer): Boolean"
	if stmt.Type.Name != expectedType {
		t.Errorf("stmt.Type.Name not %q. got=%s", expectedType, stmt.Type.Name)
	}
}

// ============================================================================
// Program Declaration Tests
// ============================================================================

// TestParseProgramDeclaration tests parsing program declarations at file start
func TestParseProgramDeclaration(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "Simple program declaration",
			input: `program Test;
begin
  PrintLn('Hello');
end`,
		},
		{
			name: "Program with variable section",
			input: `program MyProgram;
var x: Integer;
begin
  x := 42;
end`,
		},
		{
			name: "Program with const and var sections",
			input: `program Test;
const C1 = 1;
var v1: Integer;
begin
  PrintLn(C1);
end`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			// Program should parse successfully
			// The program declaration is skipped and not added to AST
			// So we just verify there are no parse errors
			if len(p.errors) > 0 {
				t.Errorf("unexpected parser errors: %v", p.errors)
			}

			// Verify the program has statements (after the program declaration)
			if len(program.Statements) == 0 {
				t.Error("program should have statements after program declaration")
			}
		})
	}
}

// TestParseProgramDeclarationOptional tests that program declaration is optional
func TestParseProgramDeclarationOptional(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantStmts  int
		checkFirst bool // Whether to check the first statement type
	}{
		{
			name: "Without program declaration",
			input: `begin
  PrintLn('Hello');
end`,
			wantStmts:  1,
			checkFirst: true,
		},
		{
			name: "With program declaration",
			input: `program Test;
begin
  PrintLn('Hello');
end`,
			wantStmts:  1,
			checkFirst: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != tt.wantStmts {
				t.Errorf("program.Statements should have %d statements. got=%d",
					tt.wantStmts, len(program.Statements))
			}

			if tt.checkFirst && len(program.Statements) > 0 {
				// First statement should be a block statement (the begin/end)
				_, ok := program.Statements[0].(*ast.BlockStatement)
				if !ok {
					t.Errorf("first statement should be *ast.BlockStatement. got=%T",
						program.Statements[0])
				}
			}
		})
	}
}

// TestParseProgramDeclarationErrors tests error cases in program declarations
func TestParseProgramDeclarationErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name:          "Missing program name",
			input:         `program;`,
			expectedError: "expected program name after 'program' keyword",
		},
		{
			name:          "Missing semicolon",
			input:         `program Test`,
			expectedError: "expected ';' after program name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			p.ParseProgram()

			if len(p.errors) == 0 {
				t.Errorf("expected parser error for input %q, but got none", tt.input)
				return
			}

			found := false
			for _, err := range p.errors {
				if strings.Contains(err, tt.expectedError) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("expected error containing %q, got errors: %v",
					tt.expectedError, p.errors)
			}
		})
	}
}

// TestParseProgramVsUnit tests that program and unit declarations are mutually exclusive
func TestParseProgramVsUnit(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{
			name: "Program declaration (valid)",
			input: `program Test;
begin
  PrintLn('Hello');
end`,
			wantError: false,
		},
		{
			name: "Unit declaration (valid)",
			input: `unit TestUnit;
interface
implementation
end.`,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			p.ParseProgram()

			hasError := len(p.errors) > 0

			if hasError != tt.wantError {
				if tt.wantError {
					t.Errorf("expected parser errors, got none")
				} else {
					t.Errorf("unexpected parser errors: %v", p.errors)
				}
			}
		})
	}
}
