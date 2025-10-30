package ast

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// ============================================================================
// Task 9.170: AST Tests for Record Literals with Named Fields
// ============================================================================

func TestRecordLiteralExpression_Simple(t *testing.T) {
	// Test simple record: (x: 10; y: 20)
	tok := lexer.Token{Type: lexer.LPAREN, Literal: "(", Pos: lexer.Position{Line: 1, Column: 1}}

	recordLit := &RecordLiteralExpression{
		Token:    tok,
		TypeName: nil, // Anonymous record
		Fields: []*FieldInitializer{
			{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
				Name:  &Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "x"}, Value: "x"},
				Value: &IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "10"}, Value: 10},
			},
			{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "y"},
				Name:  &Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "y"}, Value: "y"},
				Value: &IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "20"}, Value: 20},
			},
		},
	}

	// Test TokenLiteral()
	if recordLit.TokenLiteral() != "(" {
		t.Errorf("TokenLiteral() = %v, want '('", recordLit.TokenLiteral())
	}

	// Test Pos()
	if recordLit.Pos().Line != 1 || recordLit.Pos().Column != 1 {
		t.Errorf("Pos() = %v, want Line:1, Column:1", recordLit.Pos())
	}

	// Test Fields count
	if len(recordLit.Fields) != 2 {
		t.Errorf("len(Fields) = %v, want 2", len(recordLit.Fields))
	}

	// Test field names
	if recordLit.Fields[0].Name.Value != "x" {
		t.Errorf("Fields[0].Name.Value = %v, want 'x'", recordLit.Fields[0].Name.Value)
	}
	if recordLit.Fields[1].Name.Value != "y" {
		t.Errorf("Fields[1].Name.Value = %v, want 'y'", recordLit.Fields[1].Name.Value)
	}

	// Test String() representation
	str := recordLit.String()
	expected := "(x: 10; y: 20)"
	if str != expected {
		t.Errorf("String() = %v, want %v", str, expected)
	}
}

func TestRecordLiteralExpression_WithTypeName(t *testing.T) {
	// Test typed record: TPoint(x: 10; y: 20)
	tok := lexer.Token{Type: lexer.IDENT, Literal: "TPoint"}

	recordLit := &RecordLiteralExpression{
		Token:    tok,
		TypeName: &Identifier{Value: "TPoint"},
		Fields: []*FieldInitializer{
			{
				Name:  &Identifier{Value: "x"},
				Value: &IntegerLiteral{Token: lexer.Token{Literal: "10"}, Value: 10},
			},
			{
				Name:  &Identifier{Value: "y"},
				Value: &IntegerLiteral{Token: lexer.Token{Literal: "20"}, Value: 20},
			},
		},
	}

	// Test that TypeName is set
	if recordLit.TypeName == nil {
		t.Fatal("TypeName should not be nil")
	}
	if recordLit.TypeName.Value != "TPoint" {
		t.Errorf("TypeName.Value = %v, want 'TPoint'", recordLit.TypeName.Value)
	}

	// Test String() representation includes type name
	str := recordLit.String()
	expected := "TPoint(x: 10; y: 20)"
	if str != expected {
		t.Errorf("String() = %v, want %v", str, expected)
	}
}

func TestRecordLiteralExpression_NestedRecords(t *testing.T) {
	// Test nested records: TRect(TopLeft: (x: 0; y: 0); BottomRight: (x: 10; y: 10))
	innerRecord1 := &RecordLiteralExpression{
		Token:    lexer.Token{Type: lexer.LPAREN, Literal: "("},
		TypeName: nil,
		Fields: []*FieldInitializer{
			{Name: &Identifier{Value: "x"}, Value: &IntegerLiteral{Token: lexer.Token{Literal: "0"}, Value: 0}},
			{Name: &Identifier{Value: "y"}, Value: &IntegerLiteral{Token: lexer.Token{Literal: "0"}, Value: 0}},
		},
	}

	innerRecord2 := &RecordLiteralExpression{
		Token:    lexer.Token{Type: lexer.LPAREN, Literal: "("},
		TypeName: nil,
		Fields: []*FieldInitializer{
			{Name: &Identifier{Value: "x"}, Value: &IntegerLiteral{Token: lexer.Token{Literal: "10"}, Value: 10}},
			{Name: &Identifier{Value: "y"}, Value: &IntegerLiteral{Token: lexer.Token{Literal: "10"}, Value: 10}},
		},
	}

	outerRecord := &RecordLiteralExpression{
		Token:    lexer.Token{Type: lexer.IDENT, Literal: "TRect"},
		TypeName: &Identifier{Value: "TRect"},
		Fields: []*FieldInitializer{
			{Name: &Identifier{Value: "TopLeft"}, Value: innerRecord1},
			{Name: &Identifier{Value: "BottomRight"}, Value: innerRecord2},
		},
	}

	// Test nested structure
	if len(outerRecord.Fields) != 2 {
		t.Errorf("len(Fields) = %v, want 2", len(outerRecord.Fields))
	}

	// Verify inner records
	if _, ok := outerRecord.Fields[0].Value.(*RecordLiteralExpression); !ok {
		t.Error("Fields[0].Value should be RecordLiteralExpression")
	}
	if _, ok := outerRecord.Fields[1].Value.(*RecordLiteralExpression); !ok {
		t.Error("Fields[1].Value should be RecordLiteralExpression")
	}

	// Test String() representation
	str := outerRecord.String()
	expected := "TRect(TopLeft: (x: 0; y: 0); BottomRight: (x: 10; y: 10))"
	if str != expected {
		t.Errorf("String() = %v, want %v", str, expected)
	}
}

func TestRecordLiteralExpression_WithExpressions(t *testing.T) {
	// Test with expressions: TSphere(cx: x+5; cy: y*2; r: radius)
	recordLit := &RecordLiteralExpression{
		Token:    lexer.Token{Type: lexer.IDENT, Literal: "TSphere"},
		TypeName: &Identifier{Value: "TSphere"},
		Fields: []*FieldInitializer{
			{
				Name: &Identifier{Value: "cx"},
				Value: &BinaryExpression{
					Left:     &Identifier{Value: "x"},
					Operator: "+",
					Right:    &IntegerLiteral{Token: lexer.Token{Literal: "5"}, Value: 5},
				},
			},
			{
				Name: &Identifier{Value: "cy"},
				Value: &BinaryExpression{
					Left:     &Identifier{Value: "y"},
					Operator: "*",
					Right:    &IntegerLiteral{Token: lexer.Token{Literal: "2"}, Value: 2},
				},
			},
			{
				Name:  &Identifier{Value: "r"},
				Value: &Identifier{Value: "radius"},
			},
		},
	}

	// Test Fields count
	if len(recordLit.Fields) != 3 {
		t.Errorf("len(Fields) = %v, want 3", len(recordLit.Fields))
	}

	// Test that values are expressions
	if _, ok := recordLit.Fields[0].Value.(*BinaryExpression); !ok {
		t.Error("Fields[0].Value should be BinaryExpression")
	}
	if _, ok := recordLit.Fields[1].Value.(*BinaryExpression); !ok {
		t.Error("Fields[1].Value should be BinaryExpression")
	}
	if _, ok := recordLit.Fields[2].Value.(*Identifier); !ok {
		t.Error("Fields[2].Value should be Identifier")
	}

	// Test String() representation
	str := recordLit.String()
	expected := "TSphere(cx: (x + 5); cy: (y * 2); r: radius)"
	if str != expected {
		t.Errorf("String() = %v, want %v", str, expected)
	}
}

func TestRecordLiteralExpression_WithNegativeNumbers(t *testing.T) {
	// Test with negative numbers: (x: -50; y: 30)
	// This is from Death_Star.dws example
	recordLit := &RecordLiteralExpression{
		Token:    lexer.Token{Type: lexer.LPAREN, Literal: "("},
		TypeName: nil,
		Fields: []*FieldInitializer{
			{
				Name: &Identifier{Value: "x"},
				Value: &UnaryExpression{
					Operator: "-",
					Right:    &IntegerLiteral{Token: lexer.Token{Literal: "50"}, Value: 50},
				},
			},
			{
				Name:  &Identifier{Value: "y"},
				Value: &IntegerLiteral{Token: lexer.Token{Literal: "30"}, Value: 30},
			},
		},
	}

	// Test Fields count
	if len(recordLit.Fields) != 2 {
		t.Errorf("len(Fields) = %v, want 2", len(recordLit.Fields))
	}

	// Test that first value is a unary expression
	if _, ok := recordLit.Fields[0].Value.(*UnaryExpression); !ok {
		t.Error("Fields[0].Value should be UnaryExpression for negative number")
	}

	// Test String() representation
	str := recordLit.String()
	expected := "(x: (-50); y: 30)"
	if str != expected {
		t.Errorf("String() = %v, want %v", str, expected)
	}
}

func TestRecordLiteralExpression_DeathStarExample(t *testing.T) {
	// Test actual example from Death_Star.dws:
	// const big : TSphere = (cx: 20; cy: 20; cz: 0; r: 20);
	recordLit := &RecordLiteralExpression{
		Token:    lexer.Token{Type: lexer.LPAREN, Literal: "("},
		TypeName: nil, // Type comes from const declaration context
		Fields: []*FieldInitializer{
			{Name: &Identifier{Value: "cx"}, Value: &IntegerLiteral{Token: lexer.Token{Literal: "20"}, Value: 20}},
			{Name: &Identifier{Value: "cy"}, Value: &IntegerLiteral{Token: lexer.Token{Literal: "20"}, Value: 20}},
			{Name: &Identifier{Value: "cz"}, Value: &IntegerLiteral{Token: lexer.Token{Literal: "0"}, Value: 0}},
			{Name: &Identifier{Value: "r"}, Value: &IntegerLiteral{Token: lexer.Token{Literal: "20"}, Value: 20}},
		},
	}

	// Test Fields count
	if len(recordLit.Fields) != 4 {
		t.Errorf("len(Fields) = %v, want 4", len(recordLit.Fields))
	}

	// Verify all field names
	expectedNames := []string{"cx", "cy", "cz", "r"}
	for i, expected := range expectedNames {
		if recordLit.Fields[i].Name.Value != expected {
			t.Errorf("Fields[%d].Name.Value = %v, want %v", i, recordLit.Fields[i].Name.Value, expected)
		}
	}

	// Test String() representation
	str := recordLit.String()
	expected := "(cx: 20; cy: 20; cz: 0; r: 20)"
	if str != expected {
		t.Errorf("String() = %v, want %v", str, expected)
	}
}

func TestRecordLiteralExpression_EmptyRecord(t *testing.T) {
	// Test empty record: ()
	recordLit := &RecordLiteralExpression{
		Token:    lexer.Token{Type: lexer.LPAREN, Literal: "("},
		TypeName: nil,
		Fields:   []*FieldInitializer{},
	}

	// Test Fields count
	if len(recordLit.Fields) != 0 {
		t.Errorf("len(Fields) = %v, want 0", len(recordLit.Fields))
	}

	// Test String() representation
	str := recordLit.String()
	expected := "()"
	if str != expected {
		t.Errorf("String() = %v, want %v", str, expected)
	}
}

func TestRecordLiteralExpression_SingleField(t *testing.T) {
	// Test single field record: (value: 42)
	recordLit := &RecordLiteralExpression{
		Token:    lexer.Token{Type: lexer.LPAREN, Literal: "("},
		TypeName: nil,
		Fields: []*FieldInitializer{
			{Name: &Identifier{Value: "value"}, Value: &IntegerLiteral{Token: lexer.Token{Literal: "42"}, Value: 42}},
		},
	}

	// Test Fields count
	if len(recordLit.Fields) != 1 {
		t.Errorf("len(Fields) = %v, want 1", len(recordLit.Fields))
	}

	// Test String() representation (no trailing semicolon)
	str := recordLit.String()
	expected := "(value: 42)"
	if str != expected {
		t.Errorf("String() = %v, want %v", str, expected)
	}
}

// ============================================================================
// FieldInitializer Tests
// ============================================================================

func TestFieldInitializer_String(t *testing.T) {
	tests := []struct {
		name     string
		field    *FieldInitializer
		expected string
	}{
		{
			name: "Simple integer field",
			field: &FieldInitializer{
				Name:  &Identifier{Value: "count"},
				Value: &IntegerLiteral{Token: lexer.Token{Literal: "42"}, Value: 42},
			},
			expected: "count: 42",
		},
		{
			name: "Float field",
			field: &FieldInitializer{
				Name:  &Identifier{Value: "price"},
				Value: &FloatLiteral{Token: lexer.Token{Literal: "3.14"}, Value: 3.14},
			},
			expected: "price: 3.14",
		},
		{
			name: "String field",
			field: &FieldInitializer{
				Name:  &Identifier{Value: "name"},
				Value: &StringLiteral{Token: lexer.Token{Literal: "\"hello\""}, Value: "hello"},
			},
			expected: "name: \"hello\"",
		},
		{
			name: "Expression field",
			field: &FieldInitializer{
				Name: &Identifier{Value: "sum"},
				Value: &BinaryExpression{
					Left:     &IntegerLiteral{Token: lexer.Token{Literal: "2"}, Value: 2},
					Operator: "+",
					Right:    &IntegerLiteral{Token: lexer.Token{Literal: "3"}, Value: 3},
				},
			},
			expected: "sum: (2 + 3)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := tt.field.String()
			if str != tt.expected {
				t.Errorf("String() = %v, want %v", str, tt.expected)
			}
		})
	}
}
