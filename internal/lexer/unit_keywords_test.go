package lexer

import (
	"testing"
)

// TestUnitKeywords tests all unit-related keywords
func TestUnitKeywords(t *testing.T) {
	input := `unit MyUnit;
interface
uses System, Math;
implementation
initialization
finalization
end.`

	tests := []struct {
		expectedLiteral string
		expectedType    TokenType
	}{
		{"unit", UNIT},
		{"MyUnit", IDENT},
		{";", SEMICOLON},
		{"interface", INTERFACE},
		{"uses", USES},
		{"System", IDENT},
		{",", COMMA},
		{"Math", IDENT},
		{";", SEMICOLON},
		{"implementation", IMPLEMENTATION},
		{"initialization", INITIALIZATION},
		{"finalization", FINALIZATION},
		{"end", END},
		{".", DOT},
		{"", EOF},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestUnitKeywordsCaseInsensitive tests that unit keywords are case-insensitive
func TestUnitKeywordsCaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected TokenType
	}{
		{"lowercase unit", "unit", UNIT},
		{"uppercase UNIT", "UNIT", UNIT},
		{"mixed case Unit", "Unit", UNIT},
		{"lowercase uses", "uses", USES},
		{"uppercase USES", "USES", USES},
		{"mixed case Uses", "Uses", USES},
		{"lowercase interface", "interface", INTERFACE},
		{"uppercase INTERFACE", "INTERFACE", INTERFACE},
		{"mixed case Interface", "Interface", INTERFACE},
		{"lowercase implementation", "implementation", IMPLEMENTATION},
		{"uppercase IMPLEMENTATION", "IMPLEMENTATION", IMPLEMENTATION},
		{"mixed case Implementation", "Implementation", IMPLEMENTATION},
		{"lowercase initialization", "initialization", INITIALIZATION},
		{"uppercase INITIALIZATION", "INITIALIZATION", INITIALIZATION},
		{"mixed case Initialization", "Initialization", INITIALIZATION},
		{"lowercase finalization", "finalization", FINALIZATION},
		{"uppercase FINALIZATION", "FINALIZATION", FINALIZATION},
		{"mixed case Finalization", "Finalization", FINALIZATION},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()

			if tok.Type != tt.expected {
				t.Errorf("expected token type %q, got %q", tt.expected, tok.Type)
			}

			if !tok.Type.IsKeyword() {
				t.Errorf("token should be a keyword, got type %q", tok.Type)
			}
		})
	}
}

// TestCompleteUnitStructure tests tokenizing a complete unit structure
func TestCompleteUnitStructure(t *testing.T) {
	input := `unit TestLibrary;

interface
  uses System;

  function Add(x, y: Integer): Integer;

implementation

  function Add(x, y: Integer): Integer;
  begin
    Result := x + y;
  end;

initialization
  // Setup code

finalization
  // Cleanup code

end.`

	// Key tokens we expect to see
	expectedTokens := []TokenType{
		UNIT,           // unit
		IDENT,          // TestLibrary
		SEMICOLON,      // ;
		INTERFACE,      // interface
		USES,           // uses
		IDENT,          // System
		SEMICOLON,      // ;
		FUNCTION,       // function
		IDENT,          // Add
		LPAREN,         // (
		IDENT,          // x
		COMMA,          // ,
		IDENT,          // y
		COLON,          // :
		IDENT,          // Integer
		RPAREN,         // )
		COLON,          // :
		IDENT,          // Integer
		SEMICOLON,      // ;
		IMPLEMENTATION, // implementation
		FUNCTION,       // function
		IDENT,          // Add
		LPAREN,         // (
		IDENT,          // x
		COMMA,          // ,
		IDENT,          // y
		COLON,          // :
		IDENT,          // Integer
		RPAREN,         // )
		COLON,          // :
		IDENT,          // Integer
		SEMICOLON,      // ;
		BEGIN,          // begin
		IDENT,          // Result
		ASSIGN,         // :=
		IDENT,          // x
		PLUS,           // +
		IDENT,          // y
		SEMICOLON,      // ;
		END,            // end
		SEMICOLON,      // ;
		INITIALIZATION, // initialization
		FINALIZATION,   // finalization
		END,            // end
		DOT,            // .
		EOF,            // EOF
	}

	l := New(input)
	tokenCount := 0

	for i, expectedType := range expectedTokens {
		tok := l.NextToken()
		tokenCount++

		if tok.Type != expectedType {
			t.Fatalf("token[%d] - wrong type. expected=%q, got=%q (literal=%q)",
				i, expectedType, tok.Type, tok.Literal)
		}

		if tok.Type == EOF {
			break
		}
	}

	if tokenCount != len(expectedTokens) {
		t.Errorf("expected %d tokens, got %d", len(expectedTokens), tokenCount)
	}
}

// TestUsesClauseMultipleUnits tests tokenizing uses clauses with multiple units
func TestUsesClauseMultipleUnits(t *testing.T) {
	input := `uses System, Math, Graphics, Controls;`

	tests := []struct {
		expectedLiteral string
		expectedType    TokenType
	}{
		{"uses", USES},
		{"System", IDENT},
		{",", COMMA},
		{"Math", IDENT},
		{",", COMMA},
		{"Graphics", IDENT},
		{",", COMMA},
		{"Controls", IDENT},
		{";", SEMICOLON},
		{"", EOF},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestUnitKeywordPositions tests that unit keywords maintain correct position information
func TestUnitKeywordPositions(t *testing.T) {
	input := `unit MyUnit;
interface
implementation
end.`

	l := New(input)

	// Test first token (unit)
	tok := l.NextToken()
	if tok.Type != UNIT {
		t.Fatalf("expected UNIT, got %q", tok.Type)
	}
	if tok.Pos.Line != 1 || tok.Pos.Column != 1 {
		t.Errorf("unit keyword position: expected line=1, col=1; got line=%d, col=%d",
			tok.Pos.Line, tok.Pos.Column)
	}

	// Skip MyUnit identifier and semicolon
	l.NextToken() // MyUnit
	l.NextToken() // ;

	// Test interface keyword
	tok = l.NextToken()
	if tok.Type != INTERFACE {
		t.Fatalf("expected INTERFACE, got %q", tok.Type)
	}
	if tok.Pos.Line != 2 {
		t.Errorf("interface keyword should be on line 2, got line %d", tok.Pos.Line)
	}

	// Test implementation keyword
	tok = l.NextToken()
	if tok.Type != IMPLEMENTATION {
		t.Fatalf("expected IMPLEMENTATION, got %q", tok.Type)
	}
	if tok.Pos.Line != 3 {
		t.Errorf("implementation keyword should be on line 3, got line %d", tok.Pos.Line)
	}

	// Test end keyword
	tok = l.NextToken()
	if tok.Type != END {
		t.Fatalf("expected END, got %q", tok.Type)
	}
	if tok.Pos.Line != 4 {
		t.Errorf("end keyword should be on line 4, got line %d", tok.Pos.Line)
	}
}

// TestInterfaceVsImplementationSections tests that interface and implementation keywords
// are properly distinguished
func TestInterfaceVsImplementationSections(t *testing.T) {
	input := `unit MyUnit;
interface
  var x: Integer;
implementation
  var y: String;
end.`

	expectedSequence := []TokenType{
		UNIT, IDENT, SEMICOLON,
		INTERFACE,
		VAR, IDENT, COLON, IDENT, SEMICOLON,
		IMPLEMENTATION,
		VAR, IDENT, COLON, IDENT, SEMICOLON,
		END, DOT, EOF,
	}

	l := New(input)

	for i, expectedType := range expectedSequence {
		tok := l.NextToken()

		if tok.Type != expectedType {
			t.Fatalf("token[%d] - wrong type. expected=%q, got=%q",
				i, expectedType, tok.Type)
		}
	}
}

// TestInitializationFinalizationOrder tests init/final section order
func TestInitializationFinalizationOrder(t *testing.T) {
	input := `unit TestUnit;
interface
implementation
initialization
  x := 1;
finalization
  x := 0;
end.`

	l := New(input)

	// Skip to initialization
	for {
		tok := l.NextToken()
		if tok.Type == INITIALIZATION {
			break
		}
		if tok.Type == EOF {
			t.Fatal("didn't find INITIALIZATION keyword")
		}
	}

	// Verify we have initialization content
	l.NextToken() // x
	l.NextToken() // :=
	l.NextToken() // 1
	l.NextToken() // ;

	// Next should be finalization
	tok := l.NextToken()
	if tok.Type != FINALIZATION {
		t.Fatalf("expected FINALIZATION after initialization section, got %q", tok.Type)
	}

	// Verify finalization content
	l.NextToken() // x
	l.NextToken() // :=
	l.NextToken() // 0
	l.NextToken() // ;

	// Should end with end.
	tok = l.NextToken()
	if tok.Type != END {
		t.Fatalf("expected END after finalization section, got %q", tok.Type)
	}

	tok = l.NextToken()
	if tok.Type != DOT {
		t.Fatalf("expected DOT after END, got %q", tok.Type)
	}
}
