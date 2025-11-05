package ast

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// ============================================================================
// ArrayTypeAnnotation Tests
// ============================================================================

func TestArrayTypeAnnotation(t *testing.T) {
	t.Run("Static array type", func(t *testing.T) {
		// array[1..10] of Integer
		tok := lexer.Token{Type: lexer.ARRAY, Literal: "array"}
		intTok := lexer.Token{Type: lexer.INT, Literal: "1"}

		arrayType := &ArrayTypeAnnotation{
			Token:       tok,
			ElementType: &TypeAnnotation{Token: tok, Name: "Integer"},
			LowBound:    &IntegerLiteral{Token: intTok, Value: 1},
			HighBound:   &IntegerLiteral{Token: intTok, Value: 10},
		}

		// Test TokenLiteral()
		if arrayType.TokenLiteral() != "array" {
			t.Errorf("TokenLiteral() = %v, want 'array'", arrayType.TokenLiteral())
		}

		// Test ElementType
		if arrayType.ElementType.Name != "Integer" {
			t.Errorf("ElementType.Name = %v, want 'Integer'", arrayType.ElementType.Name)
		}

		// Test bounds - they should be IntegerLiteral expressions
		if arrayType.LowBound == nil {
			t.Error("LowBound should not be nil")
		}
		if lowBoundLit, ok := arrayType.LowBound.(*IntegerLiteral); !ok || lowBoundLit.Value != 1 {
			t.Errorf("LowBound should be IntegerLiteral with value 1")
		}
		if arrayType.HighBound == nil {
			t.Error("HighBound should not be nil")
		}
		if highBoundLit, ok := arrayType.HighBound.(*IntegerLiteral); !ok || highBoundLit.Value != 10 {
			t.Errorf("HighBound should be IntegerLiteral with value 10")
		}

		// Test IsStatic
		if !arrayType.IsStatic() {
			t.Error("IsStatic() should be true for static array")
		}

		// Test IsDynamic
		if arrayType.IsDynamic() {
			t.Error("IsDynamic() should be false for static array")
		}
	})

	t.Run("Dynamic array type", func(t *testing.T) {
		// array of String
		tok := lexer.Token{Type: lexer.ARRAY, Literal: "array"}

		arrayType := &ArrayTypeAnnotation{
			Token:       tok,
			ElementType: &TypeAnnotation{Token: tok, Name: "String"},
			LowBound:    nil,
			HighBound:   nil,
		}

		// Test bounds are nil
		if arrayType.LowBound != nil {
			t.Errorf("LowBound should be nil for dynamic array, got %v", arrayType.LowBound)
		}
		if arrayType.HighBound != nil {
			t.Errorf("HighBound should be nil for dynamic array, got %v", arrayType.HighBound)
		}

		// Test IsDynamic
		if !arrayType.IsDynamic() {
			t.Error("IsDynamic() should be true for dynamic array")
		}

		// Test IsStatic
		if arrayType.IsStatic() {
			t.Error("IsStatic() should be false for dynamic array")
		}
	})

	t.Run("String() method for static array", func(t *testing.T) {
		tok := lexer.Token{Type: lexer.ARRAY, Literal: "array"}
		lowTok := lexer.Token{Type: lexer.INT, Literal: "1"}
		highTok := lexer.Token{Type: lexer.INT, Literal: "10"}

		arrayType := &ArrayTypeAnnotation{
			Token:       tok,
			ElementType: &TypeAnnotation{Token: tok, Name: "Integer"},
			LowBound:    &IntegerLiteral{Token: lowTok, Value: 1},
			HighBound:   &IntegerLiteral{Token: highTok, Value: 10},
		}

		str := arrayType.String()
		expected := "array[1..10] of Integer"
		if str != expected {
			t.Errorf("String() = %v, want %v", str, expected)
		}
	})

	t.Run("String() method for dynamic array", func(t *testing.T) {
		tok := lexer.Token{Type: lexer.ARRAY, Literal: "array"}

		arrayType := &ArrayTypeAnnotation{
			Token:       tok,
			ElementType: &TypeAnnotation{Token: tok, Name: "String"},
			LowBound:    nil,
			HighBound:   nil,
		}

		str := arrayType.String()
		expected := "array of String"
		if str != expected {
			t.Errorf("String() = %v, want %v", str, expected)
		}
	})
}

// ============================================================================
// ArrayLiteralExpression Tests
// ============================================================================

func TestArrayLiteralExpression_String(t *testing.T) {
	lbrackTok := lexer.Token{Type: lexer.LBRACK, Literal: "["}
	intToken := func(lit string) lexer.Token {
		return lexer.Token{Type: lexer.INT, Literal: lit}
	}
	floatToken := func(lit string) lexer.Token {
		return lexer.Token{Type: lexer.FLOAT, Literal: lit}
	}
	stringToken := func(lit string) lexer.Token {
		return lexer.Token{Type: lexer.STRING, Literal: "\"" + lit + "\""}
	}
	identToken := func(lit string) lexer.Token {
		return lexer.Token{Type: lexer.IDENT, Literal: lit}
	}

	tests := []struct {
		name     string
		elements []Expression
		want     string
	}{
		{
			name: "SimpleIntegers",
			elements: []Expression{
				&IntegerLiteral{Token: intToken("1"), Value: 1},
				&IntegerLiteral{Token: intToken("2"), Value: 2},
				&IntegerLiteral{Token: intToken("3"), Value: 3},
			},
			want: "[1, 2, 3]",
		},
		{
			name: "Expressions",
			elements: []Expression{
				&BinaryExpression{
					Token:    lexer.Token{Type: lexer.PLUS, Literal: "+"},
					Left:     &Identifier{Token: identToken("x"), Value: "x"},
					Operator: "+",
					Right:    &IntegerLiteral{Token: intToken("1"), Value: 1},
				},
				&BinaryExpression{
					Token:    lexer.Token{Type: lexer.ASTERISK, Literal: "*"},
					Left:     &Identifier{Token: identToken("y"), Value: "y"},
					Operator: "*",
					Right:    &IntegerLiteral{Token: intToken("2"), Value: 2},
				},
				&BinaryExpression{
					Token:    lexer.Token{Type: lexer.MINUS, Literal: "-"},
					Left:     &Identifier{Token: identToken("z"), Value: "z"},
					Operator: "-",
					Right:    &IntegerLiteral{Token: intToken("3"), Value: 3},
				},
			},
			want: "[(x + 1), (y * 2), (z - 3)]",
		},
		{
			name: "NestedArrays",
			elements: []Expression{
				&ArrayLiteralExpression{
					Token: lbrackTok,
					Elements: []Expression{
						&IntegerLiteral{Token: intToken("1"), Value: 1},
						&IntegerLiteral{Token: intToken("2"), Value: 2},
					},
				},
				&ArrayLiteralExpression{
					Token: lbrackTok,
					Elements: []Expression{
						&IntegerLiteral{Token: intToken("3"), Value: 3},
						&IntegerLiteral{Token: intToken("4"), Value: 4},
					},
				},
			},
			want: "[[1, 2], [3, 4]]",
		},
		{
			name: "NegativeNumbers",
			elements: []Expression{
				&FloatLiteral{Token: floatToken("-50.0"), Value: -50.0},
				&IntegerLiteral{Token: intToken("30"), Value: 30},
				&IntegerLiteral{Token: intToken("50"), Value: 50},
			},
			want: "[-50.0, 30, 50]",
		},
		{
			name:     "Empty",
			elements: []Expression{},
			want:     "[]",
		},
		{
			name: "WithIdentifiersAndStrings",
			elements: []Expression{
				&Identifier{Token: identToken("names"), Value: "names"},
				&StringLiteral{Token: stringToken("DWScript"), Value: "DWScript"},
				&StringLiteral{Token: stringToken("port"), Value: "port"},
			},
			want: "[names, \"DWScript\", \"port\"]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arrayLit := &ArrayLiteralExpression{
				Token:    lbrackTok,
				Elements: tt.elements,
			}

			if arrayLit.TokenLiteral() != "[" {
				t.Fatalf("TokenLiteral() = %q, want \"[\"", arrayLit.TokenLiteral())
			}

			if got := arrayLit.String(); got != tt.want {
				t.Fatalf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestArrayLiteralExpression_TypeTracking(t *testing.T) {
	tok := lexer.Token{Type: lexer.LBRACK, Literal: "["}
	arrayLit := &ArrayLiteralExpression{
		Token:    tok,
		Elements: []Expression{},
	}

	if arrayLit.GetType() != nil {
		t.Fatal("GetType() should be nil initially")
	}

	typeAnnotation := &TypeAnnotation{Token: tok, Name: "array of Integer"}
	arrayLit.SetType(typeAnnotation)

	if arrayLit.GetType() != typeAnnotation {
		t.Fatal("GetType() should return the type set via SetType()")
	}
}

// ============================================================================
// IndexExpression Tests
// ============================================================================

func TestIndexExpression(t *testing.T) {
	t.Run("Simple array indexing", func(t *testing.T) {
		// arr[i]
		tok := lexer.Token{Type: lexer.LBRACK, Literal: "["}

		indexExpr := &IndexExpression{
			Token: tok,
			Left:  &Identifier{Token: tok, Value: "arr"},
			Index: &Identifier{Token: tok, Value: "i"},
		}

		// Test TokenLiteral()
		if indexExpr.TokenLiteral() != "[" {
			t.Errorf("TokenLiteral() = %v, want '['", indexExpr.TokenLiteral())
		}

		// Test Left (array being indexed)
		leftIdent, ok := indexExpr.Left.(*Identifier)
		if !ok {
			t.Fatal("Left should be an Identifier")
		}
		if leftIdent.Value != "arr" {
			t.Errorf("Left.Value = %v, want 'arr'", leftIdent.Value)
		}

		// Test Index
		indexIdent, ok := indexExpr.Index.(*Identifier)
		if !ok {
			t.Fatal("Index should be an Identifier")
		}
		if indexIdent.Value != "i" {
			t.Errorf("Index.Value = %v, want 'i'", indexIdent.Value)
		}
	})

	t.Run("Array indexing with integer literal", func(t *testing.T) {
		// arr[0]
		tok := lexer.Token{Type: lexer.LBRACK, Literal: "["}

		indexExpr := &IndexExpression{
			Token: tok,
			Left:  &Identifier{Token: tok, Value: "arr"},
			Index: &IntegerLiteral{Token: tok, Value: 0},
		}

		// Test Index is integer literal
		indexInt, ok := indexExpr.Index.(*IntegerLiteral)
		if !ok {
			t.Fatal("Index should be an IntegerLiteral")
		}
		if indexInt.Value != 0 {
			t.Errorf("Index.Value = %v, want 0", indexInt.Value)
		}
	})

	t.Run("Array indexing with expression", func(t *testing.T) {
		// arr[i + 1]
		tok := lexer.Token{Type: lexer.LBRACK, Literal: "["}

		indexExpr := &IndexExpression{
			Token: tok,
			Left:  &Identifier{Token: tok, Value: "arr"},
			Index: &BinaryExpression{
				Token:    lexer.Token{Type: lexer.PLUS, Literal: "+"},
				Left:     &Identifier{Token: tok, Value: "i"},
				Operator: "+",
				Right:    &IntegerLiteral{Token: tok, Value: 1},
			},
		}

		// Test Index is binary expression
		_, ok := indexExpr.Index.(*BinaryExpression)
		if !ok {
			t.Fatal("Index should be a BinaryExpression")
		}
	})

	t.Run("Nested array indexing", func(t *testing.T) {
		// arr[i][j]
		tok := lexer.Token{Type: lexer.LBRACK, Literal: "["}

		// First create arr[i]
		innerIndex := &IndexExpression{
			Token: tok,
			Left:  &Identifier{Token: tok, Value: "arr"},
			Index: &Identifier{Token: tok, Value: "i"},
		}

		// Then create (arr[i])[j]
		outerIndex := &IndexExpression{
			Token: tok,
			Left:  innerIndex,
			Index: &Identifier{Token: tok, Value: "j"},
		}

		// Test that Left is an IndexExpression
		_, ok := outerIndex.Left.(*IndexExpression)
		if !ok {
			t.Fatal("Left should be an IndexExpression for nested indexing")
		}
	})

	t.Run("String() method", func(t *testing.T) {
		tok := lexer.Token{Type: lexer.LBRACK, Literal: "["}
		intTok := lexer.Token{Type: lexer.INT, Literal: "5"}

		indexExpr := &IndexExpression{
			Token: tok,
			Left:  &Identifier{Token: tok, Value: "arr"},
			Index: &IntegerLiteral{Token: intTok, Value: 5},
		}

		str := indexExpr.String()
		expected := "(arr[5])"
		if str != expected {
			t.Errorf("String() = %v, want %v", str, expected)
		}
	})

	t.Run("Implements Expression interface", func(_ *testing.T) {
		tok := lexer.Token{Type: lexer.LBRACK, Literal: "["}
		indexExpr := &IndexExpression{
			Token: tok,
			Left:  &Identifier{Token: tok, Value: "arr"},
			Index: &IntegerLiteral{Token: tok, Value: 0},
		}

		// Ensure it implements Expression interface
		var _ Expression = indexExpr
	})

	t.Run("Type tracking", func(t *testing.T) {
		tok := lexer.Token{Type: lexer.LBRACK, Literal: "["}
		indexExpr := &IndexExpression{
			Token: tok,
			Left:  &Identifier{Token: tok, Value: "arr"},
			Index: &IntegerLiteral{Token: tok, Value: 0},
		}

		// Test GetType (should be nil initially)
		if indexExpr.GetType() != nil {
			t.Error("GetType() should be nil initially")
		}

		// Test SetType
		typeAnnotation := &TypeAnnotation{Token: tok, Name: "Integer"}
		indexExpr.SetType(typeAnnotation)

		// Test GetType after setting
		if indexExpr.GetType() != typeAnnotation {
			t.Error("GetType() should return set type")
		}
	})
}

// ============================================================================
// AssignmentStatement with IndexExpression Tests
// ============================================================================

func TestAssignmentStatement_WithIndexExpression(t *testing.T) {
	t.Run("Assignment to array element with identifier index", func(t *testing.T) {
		// arr[i] := 42
		tok := lexer.Token{Type: lexer.ASSIGN, Literal: ":="}

		// Create the index expression (arr[i])
		indexExpr := &IndexExpression{
			Token: lexer.Token{Type: lexer.LBRACK, Literal: "["},
			Left:  &Identifier{Token: tok, Value: "arr"},
			Index: &Identifier{Token: tok, Value: "i"},
		}

		// Create the assignment statement
		assignStmt := &AssignmentStatement{
			Token:  tok,
			Target: indexExpr, // Using Target instead of Name
			Value:  &IntegerLiteral{Token: tok, Value: 42},
		}

		// Test TokenLiteral()
		if assignStmt.TokenLiteral() != ":=" {
			t.Errorf("TokenLiteral() = %v, want ':='", assignStmt.TokenLiteral())
		}

		// Test Target is an IndexExpression
		targetIndex, ok := assignStmt.Target.(*IndexExpression)
		if !ok {
			t.Fatal("Target should be an IndexExpression")
		}
		if targetIndex.Left.(*Identifier).Value != "arr" {
			t.Errorf("Target array name = %v, want 'arr'", targetIndex.Left.(*Identifier).Value)
		}

		// Test Value
		valueInt, ok := assignStmt.Value.(*IntegerLiteral)
		if !ok {
			t.Fatal("Value should be an IntegerLiteral")
		}
		if valueInt.Value != 42 {
			t.Errorf("Value = %v, want 42", valueInt.Value)
		}
	})

	t.Run("Assignment to array element with literal index", func(t *testing.T) {
		// arr[0] := 100
		tok := lexer.Token{Type: lexer.ASSIGN, Literal: ":="}

		indexExpr := &IndexExpression{
			Token: lexer.Token{Type: lexer.LBRACK, Literal: "["},
			Left:  &Identifier{Token: tok, Value: "arr"},
			Index: &IntegerLiteral{Token: tok, Value: 0},
		}

		assignStmt := &AssignmentStatement{
			Token:  tok,
			Target: indexExpr,
			Value:  &IntegerLiteral{Token: tok, Value: 100},
		}

		// Test Target is an IndexExpression
		_, ok := assignStmt.Target.(*IndexExpression)
		if !ok {
			t.Fatal("Target should be an IndexExpression")
		}
	})

	t.Run("Assignment to nested array element", func(t *testing.T) {
		// matrix[i][j] := 99
		tok := lexer.Token{Type: lexer.ASSIGN, Literal: ":="}

		// Create matrix[i]
		innerIndex := &IndexExpression{
			Token: lexer.Token{Type: lexer.LBRACK, Literal: "["},
			Left:  &Identifier{Token: tok, Value: "matrix"},
			Index: &Identifier{Token: tok, Value: "i"},
		}

		// Create (matrix[i])[j]
		outerIndex := &IndexExpression{
			Token: lexer.Token{Type: lexer.LBRACK, Literal: "["},
			Left:  innerIndex,
			Index: &Identifier{Token: tok, Value: "j"},
		}

		assignStmt := &AssignmentStatement{
			Token:  tok,
			Target: outerIndex,
			Value:  &IntegerLiteral{Token: tok, Value: 99},
		}

		// Test Target is an IndexExpression
		targetIndex, ok := assignStmt.Target.(*IndexExpression)
		if !ok {
			t.Fatal("Target should be an IndexExpression")
		}

		// Test nested structure
		_, ok = targetIndex.Left.(*IndexExpression)
		if !ok {
			t.Fatal("Target.Left should be an IndexExpression for nested indexing")
		}
	})

	t.Run("String() method with index target", func(t *testing.T) {
		assignTok := lexer.Token{Type: lexer.ASSIGN, Literal: ":="}
		bracketTok := lexer.Token{Type: lexer.LBRACK, Literal: "["}
		identTok := lexer.Token{Type: lexer.IDENT, Literal: "arr"}
		intTok := lexer.Token{Type: lexer.INT, Literal: "5"}
		valueTok := lexer.Token{Type: lexer.INT, Literal: "42"}

		indexExpr := &IndexExpression{
			Token: bracketTok,
			Left:  &Identifier{Token: identTok, Value: "arr"},
			Index: &IntegerLiteral{Token: intTok, Value: 5},
		}

		assignStmt := &AssignmentStatement{
			Token:  assignTok,
			Target: indexExpr,
			Value:  &IntegerLiteral{Token: valueTok, Value: 42},
		}

		str := assignStmt.String()
		expected := "(arr[5]) := 42"
		if str != expected {
			t.Errorf("String() = %v, want %v", str, expected)
		}
	})

	t.Run("Backward compatibility - simple identifier assignment", func(t *testing.T) {
		// x := 10
		assignTok := lexer.Token{Type: lexer.ASSIGN, Literal: ":="}
		identTok := lexer.Token{Type: lexer.IDENT, Literal: "x"}
		valueTok := lexer.Token{Type: lexer.INT, Literal: "10"}

		// Using an Identifier as the Target (backward compatibility)
		assignStmt := &AssignmentStatement{
			Token:  assignTok,
			Target: &Identifier{Token: identTok, Value: "x"},
			Value:  &IntegerLiteral{Token: valueTok, Value: 10},
		}

		// Test Target is an Identifier
		targetIdent, ok := assignStmt.Target.(*Identifier)
		if !ok {
			t.Fatal("Target should be an Identifier for backward compatibility")
		}
		if targetIdent.Value != "x" {
			t.Errorf("Target.Value = %v, want 'x'", targetIdent.Value)
		}

		// Test String() for simple assignment
		str := assignStmt.String()
		expected := "x := 10"
		if str != expected {
			t.Errorf("String() = %v, want %v", str, expected)
		}
	})
}

// ============================================================================
// NewArrayExpression Tests
// ============================================================================

func TestNewArrayExpression(t *testing.T) {
	t.Run("Simple 1D array instantiation", func(t *testing.T) {
		// new Integer[16]
		newTok := lexer.Token{Type: lexer.NEW, Literal: "new"}
		intTok := lexer.Token{Type: lexer.IDENT, Literal: "Integer"}

		newArrayExpr := &NewArrayExpression{
			Token:           newTok,
			ElementTypeName: &Identifier{Token: intTok, Value: "Integer"},
			Dimensions: []Expression{
				&IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "16"}, Value: 16},
			},
		}

		// Test TokenLiteral()
		if newArrayExpr.TokenLiteral() != "new" {
			t.Errorf("TokenLiteral() = %v, want 'new'", newArrayExpr.TokenLiteral())
		}

		// Test ElementTypeName
		if newArrayExpr.ElementTypeName == nil {
			t.Fatal("ElementTypeName should not be nil")
		}
		if newArrayExpr.ElementTypeName.Value != "Integer" {
			t.Errorf("ElementTypeName.Value = %v, want 'Integer'", newArrayExpr.ElementTypeName.Value)
		}

		// Test Dimensions count
		if len(newArrayExpr.Dimensions) != 1 {
			t.Errorf("len(Dimensions) = %v, want 1", len(newArrayExpr.Dimensions))
		}

		// Test dimension value
		dimInt, ok := newArrayExpr.Dimensions[0].(*IntegerLiteral)
		if !ok {
			t.Fatal("Dimension should be an IntegerLiteral")
		}
		if dimInt.Value != 16 {
			t.Errorf("Dimension value = %v, want 16", dimInt.Value)
		}
	})

	t.Run("2D array instantiation", func(t *testing.T) {
		// new Integer[10, 20]
		newTok := lexer.Token{Type: lexer.NEW, Literal: "new"}
		intTok := lexer.Token{Type: lexer.IDENT, Literal: "Integer"}

		newArrayExpr := &NewArrayExpression{
			Token:           newTok,
			ElementTypeName: &Identifier{Token: intTok, Value: "Integer"},
			Dimensions: []Expression{
				&IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "10"}, Value: 10},
				&IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "20"}, Value: 20},
			},
		}

		// Test Dimensions count for 2D array
		if len(newArrayExpr.Dimensions) != 2 {
			t.Errorf("len(Dimensions) = %v, want 2", len(newArrayExpr.Dimensions))
		}

		// Test first dimension
		dim1, ok := newArrayExpr.Dimensions[0].(*IntegerLiteral)
		if !ok {
			t.Fatal("First dimension should be an IntegerLiteral")
		}
		if dim1.Value != 10 {
			t.Errorf("First dimension value = %v, want 10", dim1.Value)
		}

		// Test second dimension
		dim2, ok := newArrayExpr.Dimensions[1].(*IntegerLiteral)
		if !ok {
			t.Fatal("Second dimension should be an IntegerLiteral")
		}
		if dim2.Value != 20 {
			t.Errorf("Second dimension value = %v, want 20", dim2.Value)
		}
	})

	t.Run("Array with expression-based size", func(t *testing.T) {
		// new String[Length(s)+1]
		newTok := lexer.Token{Type: lexer.NEW, Literal: "new"}
		strTok := lexer.Token{Type: lexer.IDENT, Literal: "String"}

		// Create: Length(s) + 1
		sizeExpr := &BinaryExpression{
			Token: lexer.Token{Type: lexer.PLUS, Literal: "+"},
			Left: &CallExpression{
				Token:    lexer.Token{Type: lexer.IDENT, Literal: "Length"},
				Function: &Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "Length"}, Value: "Length"},
				Arguments: []Expression{
					&Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "s"}, Value: "s"},
				},
			},
			Operator: "+",
			Right:    &IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "1"}, Value: 1},
		}

		newArrayExpr := &NewArrayExpression{
			Token:           newTok,
			ElementTypeName: &Identifier{Token: strTok, Value: "String"},
			Dimensions:      []Expression{sizeExpr},
		}

		// Test that dimension is a BinaryExpression
		_, ok := newArrayExpr.Dimensions[0].(*BinaryExpression)
		if !ok {
			t.Fatal("Dimension should be a BinaryExpression for computed size")
		}

		// Test ElementTypeName is String
		if newArrayExpr.ElementTypeName.Value != "String" {
			t.Errorf("ElementTypeName = %v, want 'String'", newArrayExpr.ElementTypeName.Value)
		}
	})

	t.Run("3D array instantiation", func(t *testing.T) {
		// new Float[5, 10, 15]
		newTok := lexer.Token{Type: lexer.NEW, Literal: "new"}
		floatTok := lexer.Token{Type: lexer.IDENT, Literal: "Float"}

		newArrayExpr := &NewArrayExpression{
			Token:           newTok,
			ElementTypeName: &Identifier{Token: floatTok, Value: "Float"},
			Dimensions: []Expression{
				&IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "5"}, Value: 5},
				&IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "10"}, Value: 10},
				&IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "15"}, Value: 15},
			},
		}

		// Test Dimensions count for 3D array
		if len(newArrayExpr.Dimensions) != 3 {
			t.Errorf("len(Dimensions) = %v, want 3", len(newArrayExpr.Dimensions))
		}
	})

	t.Run("String() method for 1D array", func(t *testing.T) {
		newTok := lexer.Token{Type: lexer.NEW, Literal: "new"}
		intTok := lexer.Token{Type: lexer.IDENT, Literal: "Integer"}

		newArrayExpr := &NewArrayExpression{
			Token:           newTok,
			ElementTypeName: &Identifier{Token: intTok, Value: "Integer"},
			Dimensions: []Expression{
				&IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "16"}, Value: 16},
			},
		}

		str := newArrayExpr.String()
		expected := "new Integer[16]"
		if str != expected {
			t.Errorf("String() = %v, want %v", str, expected)
		}
	})

	t.Run("String() method for 2D array", func(t *testing.T) {
		newTok := lexer.Token{Type: lexer.NEW, Literal: "new"}
		intTok := lexer.Token{Type: lexer.IDENT, Literal: "Integer"}

		newArrayExpr := &NewArrayExpression{
			Token:           newTok,
			ElementTypeName: &Identifier{Token: intTok, Value: "Integer"},
			Dimensions: []Expression{
				&IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "10"}, Value: 10},
				&IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "20"}, Value: 20},
			},
		}

		str := newArrayExpr.String()
		expected := "new Integer[10, 20]"
		if str != expected {
			t.Errorf("String() = %v, want %v", str, expected)
		}
	})

	t.Run("Implements Expression interface", func(_ *testing.T) {
		newTok := lexer.Token{Type: lexer.NEW, Literal: "new"}
		intTok := lexer.Token{Type: lexer.IDENT, Literal: "Integer"}

		newArrayExpr := &NewArrayExpression{
			Token:           newTok,
			ElementTypeName: &Identifier{Token: intTok, Value: "Integer"},
			Dimensions: []Expression{
				&IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "10"}, Value: 10},
			},
		}

		// Ensure it implements Expression interface
		var _ Expression = newArrayExpr
	})

	t.Run("Type tracking", func(t *testing.T) {
		newTok := lexer.Token{Type: lexer.NEW, Literal: "new"}
		intTok := lexer.Token{Type: lexer.IDENT, Literal: "Integer"}

		newArrayExpr := &NewArrayExpression{
			Token:           newTok,
			ElementTypeName: &Identifier{Token: intTok, Value: "Integer"},
			Dimensions: []Expression{
				&IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "10"}, Value: 10},
			},
		}

		// Test GetType (should be nil initially)
		if newArrayExpr.GetType() != nil {
			t.Error("GetType() should be nil initially")
		}

		// Test SetType
		typeAnnotation := &TypeAnnotation{Token: intTok, Name: "array of Integer"}
		newArrayExpr.SetType(typeAnnotation)

		// Test GetType after setting
		if newArrayExpr.GetType() != typeAnnotation {
			t.Error("GetType() should return set type")
		}
	})

	t.Run("Different element types", func(t *testing.T) {
		newTok := lexer.Token{Type: lexer.NEW, Literal: "new"}

		// Test with String type
		strTok := lexer.Token{Type: lexer.IDENT, Literal: "String"}
		stringArray := &NewArrayExpression{
			Token:           newTok,
			ElementTypeName: &Identifier{Token: strTok, Value: "String"},
			Dimensions: []Expression{
				&IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "5"}, Value: 5},
			},
		}

		if stringArray.ElementTypeName.Value != "String" {
			t.Errorf("String array ElementTypeName = %v, want 'String'", stringArray.ElementTypeName.Value)
		}

		// Test with Boolean type
		boolTok := lexer.Token{Type: lexer.IDENT, Literal: "Boolean"}
		boolArray := &NewArrayExpression{
			Token:           newTok,
			ElementTypeName: &Identifier{Token: boolTok, Value: "Boolean"},
			Dimensions: []Expression{
				&IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "3"}, Value: 3},
			},
		}

		if boolArray.ElementTypeName.Value != "Boolean" {
			t.Errorf("Boolean array ElementTypeName = %v, want 'Boolean'", boolArray.ElementTypeName.Value)
		}
	})

	t.Run("Position tracking", func(t *testing.T) {
		newTok := lexer.Token{
			Type:    lexer.NEW,
			Literal: "new",
			Pos:     lexer.Position{Line: 23, Column: 5},
		}
		intTok := lexer.Token{Type: lexer.IDENT, Literal: "Integer"}

		newArrayExpr := &NewArrayExpression{
			Token:           newTok,
			ElementTypeName: &Identifier{Token: intTok, Value: "Integer"},
			Dimensions: []Expression{
				&IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "16"}, Value: 16},
			},
		}

		// Test Pos() returns correct position
		pos := newArrayExpr.Pos()
		if pos.Line != 23 || pos.Column != 5 {
			t.Errorf("Pos() = Line %v, Column %v, want Line 23, Column 5", pos.Line, pos.Column)
		}
	})
}
