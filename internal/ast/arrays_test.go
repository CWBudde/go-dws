package ast

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// ============================================================================
// ArrayTypeAnnotation Tests (Task 8.119)
// ============================================================================

func TestArrayTypeAnnotation(t *testing.T) {
	t.Run("Static array type", func(t *testing.T) {
		// array[1..10] of Integer
		tok := lexer.Token{Type: lexer.ARRAY, Literal: "array"}

		lowBound := 1
		highBound := 10
		arrayType := &ArrayTypeAnnotation{
			Token:       tok,
			ElementType: &TypeAnnotation{Token: tok, Name: "Integer"},
			LowBound:    &lowBound,
			HighBound:   &highBound,
		}

		// Test TokenLiteral()
		if arrayType.TokenLiteral() != "array" {
			t.Errorf("TokenLiteral() = %v, want 'array'", arrayType.TokenLiteral())
		}

		// Test ElementType
		if arrayType.ElementType.Name != "Integer" {
			t.Errorf("ElementType.Name = %v, want 'Integer'", arrayType.ElementType.Name)
		}

		// Test bounds
		if arrayType.LowBound == nil || *arrayType.LowBound != 1 {
			t.Errorf("LowBound = %v, want 1", arrayType.LowBound)
		}
		if arrayType.HighBound == nil || *arrayType.HighBound != 10 {
			t.Errorf("HighBound = %v, want 10", arrayType.HighBound)
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
			t.Errorf("LowBound should be nil for dynamic array, got %v", *arrayType.LowBound)
		}
		if arrayType.HighBound != nil {
			t.Errorf("HighBound should be nil for dynamic array, got %v", *arrayType.HighBound)
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

		lowBound := 1
		highBound := 10
		arrayType := &ArrayTypeAnnotation{
			Token:       tok,
			ElementType: &TypeAnnotation{Token: tok, Name: "Integer"},
			LowBound:    &lowBound,
			HighBound:   &highBound,
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
// ArrayLiteral Tests (Task 8.120)
// ============================================================================

func TestArrayLiteral(t *testing.T) {
	t.Run("Array with integer elements", func(t *testing.T) {
		// [1, 2, 3]
		tok := lexer.Token{Type: lexer.LBRACK, Literal: "["}

		arrayLit := &ArrayLiteral{
			Token: tok,
			Elements: []Expression{
				&IntegerLiteral{Token: tok, Value: 1},
				&IntegerLiteral{Token: tok, Value: 2},
				&IntegerLiteral{Token: tok, Value: 3},
			},
		}

		// Test TokenLiteral()
		if arrayLit.TokenLiteral() != "[" {
			t.Errorf("TokenLiteral() = %v, want '['", arrayLit.TokenLiteral())
		}

		// Test Elements count
		if len(arrayLit.Elements) != 3 {
			t.Errorf("len(Elements) = %v, want 3", len(arrayLit.Elements))
		}
	})

	t.Run("Empty array", func(t *testing.T) {
		// []
		tok := lexer.Token{Type: lexer.LBRACK, Literal: "["}

		arrayLit := &ArrayLiteral{
			Token:    tok,
			Elements: []Expression{},
		}

		// Test empty Elements
		if len(arrayLit.Elements) != 0 {
			t.Errorf("len(Elements) = %v, want 0", len(arrayLit.Elements))
		}
	})

	t.Run("Array with string elements", func(t *testing.T) {
		// ['hello', 'world']
		tok := lexer.Token{Type: lexer.LBRACK, Literal: "["}

		arrayLit := &ArrayLiteral{
			Token: tok,
			Elements: []Expression{
				&StringLiteral{Token: tok, Value: "hello"},
				&StringLiteral{Token: tok, Value: "world"},
			},
		}

		// Test Elements count
		if len(arrayLit.Elements) != 2 {
			t.Errorf("len(Elements) = %v, want 2", len(arrayLit.Elements))
		}
	})

	t.Run("String() method", func(t *testing.T) {
		tok := lexer.Token{Type: lexer.LBRACK, Literal: "["}

		arrayLit := &ArrayLiteral{
			Token: tok,
			Elements: []Expression{
				&IntegerLiteral{Token: tok, Value: 1},
				&IntegerLiteral{Token: tok, Value: 2},
			},
		}

		str := arrayLit.String()
		// Should contain meaningful representation
		if str == "" {
			t.Error("String() should not be empty")
		}
	})

	t.Run("Implements Expression interface", func(t *testing.T) {
		tok := lexer.Token{Type: lexer.LBRACK, Literal: "["}
		arrayLit := &ArrayLiteral{
			Token:    tok,
			Elements: []Expression{},
		}

		// Ensure it implements Expression interface
		var _ Expression = arrayLit
	})

	t.Run("Type tracking", func(t *testing.T) {
		tok := lexer.Token{Type: lexer.LBRACK, Literal: "["}
		arrayLit := &ArrayLiteral{
			Token:    tok,
			Elements: []Expression{},
		}

		// Test GetType (should be nil initially)
		if arrayLit.GetType() != nil {
			t.Error("GetType() should be nil initially")
		}

		// Test SetType
		typeAnnotation := &TypeAnnotation{Token: tok, Name: "array of Integer"}
		arrayLit.SetType(typeAnnotation)

		// Test GetType after setting
		if arrayLit.GetType() != typeAnnotation {
			t.Error("GetType() should return set type")
		}
	})
}

// ============================================================================
// IndexExpression Tests (Task 8.120)
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

	t.Run("Implements Expression interface", func(t *testing.T) {
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
// AssignmentStatement with IndexExpression Tests (Task 8.137)
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
