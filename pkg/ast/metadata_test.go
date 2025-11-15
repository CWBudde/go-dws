package ast

import (
	"sync"
	"testing"

	"github.com/cwbudde/go-dws/pkg/token"
)

func TestNewSemanticInfo(t *testing.T) {
	si := NewSemanticInfo()

	if si == nil {
		t.Fatal("NewSemanticInfo() returned nil")
	}

	if si.TypeCount() != 0 {
		t.Errorf("TypeCount() = %d, want 0", si.TypeCount())
	}

	if si.SymbolCount() != 0 {
		t.Errorf("SymbolCount() = %d, want 0", si.SymbolCount())
	}
}

func TestSemanticInfo_TypeOperations(t *testing.T) {
	si := NewSemanticInfo()

	// Create a test expression node
	expr := &IntegerLiteral{
		TypedExpressionBase: TypedExpressionBase{
			BaseNode: BaseNode{
				Token: token.Token{Type: token.INT, Literal: "42", Pos: token.Position{Line: 1, Column: 1}},
			},
		},
		Value: 42,
	}

	// Create a test type annotation
	typeAnnotation := &TypeAnnotation{
		Name: "Integer",
		Token: token.Token{
			Type:    token.IDENT,
			Literal: "Integer",
			Pos:     token.Position{Line: 1, Column: 5},
		},
	}

	// Test GetType on expression without type
	if typ := si.GetType(expr); typ != nil {
		t.Errorf("GetType() = %v, want nil", typ)
	}

	// Test HasType on expression without type
	if si.HasType(expr) {
		t.Error("HasType() = true, want false")
	}

	// Set type
	si.SetType(expr, typeAnnotation)

	// Test GetType after setting
	if typ := si.GetType(expr); typ != typeAnnotation {
		t.Errorf("GetType() = %v, want %v", typ, typeAnnotation)
	}

	// Test HasType after setting
	if !si.HasType(expr) {
		t.Error("HasType() = false, want true")
	}

	// Test TypeCount
	if count := si.TypeCount(); count != 1 {
		t.Errorf("TypeCount() = %d, want 1", count)
	}

	// Clear type
	si.ClearType(expr)

	// Test after clearing
	if typ := si.GetType(expr); typ != nil {
		t.Errorf("GetType() after ClearType = %v, want nil", typ)
	}

	if si.HasType(expr) {
		t.Error("HasType() after ClearType = true, want false")
	}

	if count := si.TypeCount(); count != 0 {
		t.Errorf("TypeCount() after ClearType = %d, want 0", count)
	}
}

func TestSemanticInfo_SymbolOperations(t *testing.T) {
	si := NewSemanticInfo()

	// Create a test identifier node
	ident := &Identifier{
		TypedExpressionBase: TypedExpressionBase{
			BaseNode: BaseNode{
				Token: token.Token{Type: token.IDENT, Literal: "x", Pos: token.Position{Line: 1, Column: 1}},
			},
		},
		Value: "x",
	}

	// Create a test symbol (using interface{} for now)
	symbol := "test_symbol"

	// Test GetSymbol on identifier without symbol
	if sym := si.GetSymbol(ident); sym != nil {
		t.Errorf("GetSymbol() = %v, want nil", sym)
	}

	// Test HasSymbol on identifier without symbol
	if si.HasSymbol(ident) {
		t.Error("HasSymbol() = true, want false")
	}

	// Set symbol
	si.SetSymbol(ident, symbol)

	// Test GetSymbol after setting
	if sym := si.GetSymbol(ident); sym != symbol {
		t.Errorf("GetSymbol() = %v, want %v", sym, symbol)
	}

	// Test HasSymbol after setting
	if !si.HasSymbol(ident) {
		t.Error("HasSymbol() = false, want true")
	}

	// Test SymbolCount
	if count := si.SymbolCount(); count != 1 {
		t.Errorf("SymbolCount() = %d, want 1", count)
	}
}

func TestSemanticInfo_MultipleExpressions(t *testing.T) {
	si := NewSemanticInfo()

	// Create multiple expression nodes
	expr1 := &IntegerLiteral{
		TypedExpressionBase: TypedExpressionBase{
			BaseNode: BaseNode{
				Token: token.Token{Type: token.INT, Literal: "42", Pos: token.Position{Line: 1, Column: 1}},
			},
		},
		Value: 42,
	}

	expr2 := &StringLiteral{
		TypedExpressionBase: TypedExpressionBase{
			BaseNode: BaseNode{
				Token: token.Token{Type: token.STRING, Literal: "hello", Pos: token.Position{Line: 2, Column: 1}},
			},
		},
		Value: "hello",
	}

	expr3 := &BooleanLiteral{
		TypedExpressionBase: TypedExpressionBase{
			BaseNode: BaseNode{
				Token: token.Token{Type: token.TRUE, Literal: "True", Pos: token.Position{Line: 3, Column: 1}},
			},
		},
		Value: true,
	}

	// Create type annotations
	intType := &TypeAnnotation{Name: "Integer"}
	strType := &TypeAnnotation{Name: "String"}
	boolType := &TypeAnnotation{Name: "Boolean"}

	// Set types
	si.SetType(expr1, intType)
	si.SetType(expr2, strType)
	si.SetType(expr3, boolType)

	// Verify all types
	if typ := si.GetType(expr1); typ != intType {
		t.Errorf("GetType(expr1) = %v, want %v", typ, intType)
	}
	if typ := si.GetType(expr2); typ != strType {
		t.Errorf("GetType(expr2) = %v, want %v", typ, strType)
	}
	if typ := si.GetType(expr3); typ != boolType {
		t.Errorf("GetType(expr3) = %v, want %v", typ, boolType)
	}

	// Verify count
	if count := si.TypeCount(); count != 3 {
		t.Errorf("TypeCount() = %d, want 3", count)
	}
}

func TestSemanticInfo_Clear(t *testing.T) {
	si := NewSemanticInfo()

	// Add some data
	expr := &IntegerLiteral{
		TypedExpressionBase: TypedExpressionBase{
			BaseNode: BaseNode{
				Token: token.Token{Type: token.INT, Literal: "42", Pos: token.Position{Line: 1, Column: 1}},
			},
		},
		Value: 42,
	}
	ident := &Identifier{
		TypedExpressionBase: TypedExpressionBase{
			BaseNode: BaseNode{
				Token: token.Token{Type: token.IDENT, Literal: "x", Pos: token.Position{Line: 1, Column: 1}},
			},
		},
		Value: "x",
	}

	si.SetType(expr, &TypeAnnotation{Name: "Integer"})
	si.SetSymbol(ident, "symbol")

	// Verify data was added
	if si.TypeCount() != 1 {
		t.Errorf("TypeCount() before Clear = %d, want 1", si.TypeCount())
	}
	if si.SymbolCount() != 1 {
		t.Errorf("SymbolCount() before Clear = %d, want 1", si.SymbolCount())
	}

	// Clear
	si.Clear()

	// Verify data was cleared
	if si.TypeCount() != 0 {
		t.Errorf("TypeCount() after Clear = %d, want 0", si.TypeCount())
	}
	if si.SymbolCount() != 0 {
		t.Errorf("SymbolCount() after Clear = %d, want 0", si.SymbolCount())
	}
	if si.GetType(expr) != nil {
		t.Error("GetType() after Clear should return nil")
	}
	if si.GetSymbol(ident) != nil {
		t.Error("GetSymbol() after Clear should return nil")
	}
}

func TestSemanticInfo_ConcurrentReads(t *testing.T) {
	si := NewSemanticInfo()

	// Create test data
	expr := &IntegerLiteral{
		TypedExpressionBase: TypedExpressionBase{
			BaseNode: BaseNode{
				Token: token.Token{Type: token.INT, Literal: "42", Pos: token.Position{Line: 1, Column: 1}},
			},
		},
		Value: 42,
	}
	typeAnnotation := &TypeAnnotation{Name: "Integer"}
	si.SetType(expr, typeAnnotation)

	// Concurrent reads
	const numReaders = 100
	var wg sync.WaitGroup
	wg.Add(numReaders)

	errors := make(chan error, numReaders)

	for i := 0; i < numReaders; i++ {
		go func() {
			defer wg.Done()
			typ := si.GetType(expr)
			if typ != typeAnnotation {
				errors <- nil // Signal error without blocking
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	errorCount := 0
	for range errors {
		errorCount++
	}

	if errorCount > 0 {
		t.Errorf("Concurrent reads produced %d errors", errorCount)
	}
}

func TestSemanticInfo_IndependentInstances(t *testing.T) {
	// Create two separate SemanticInfo instances
	si1 := NewSemanticInfo()
	si2 := NewSemanticInfo()

	// Create a shared expression node
	expr := &IntegerLiteral{
		TypedExpressionBase: TypedExpressionBase{
			BaseNode: BaseNode{
				Token: token.Token{Type: token.INT, Literal: "42", Pos: token.Position{Line: 1, Column: 1}},
			},
		},
		Value: 42,
	}

	// Set different types in each instance
	type1 := &TypeAnnotation{Name: "Integer"}
	type2 := &TypeAnnotation{Name: "Float"}

	si1.SetType(expr, type1)
	si2.SetType(expr, type2)

	// Verify each instance has its own type
	if typ := si1.GetType(expr); typ != type1 {
		t.Errorf("si1.GetType(expr) = %v, want %v", typ, type1)
	}
	if typ := si2.GetType(expr); typ != type2 {
		t.Errorf("si2.GetType(expr) = %v, want %v", typ, type2)
	}

	// Verify they are truly independent
	si1.ClearType(expr)
	if typ := si2.GetType(expr); typ != type2 {
		t.Errorf("si2.GetType(expr) after si1.ClearType = %v, want %v", typ, type2)
	}
}

func TestSemanticInfo_NilExpression(t *testing.T) {
	si := NewSemanticInfo()

	// Test with nil expression (edge case)
	// This should not panic
	typ := si.GetType(nil)
	if typ != nil {
		t.Errorf("GetType(nil) = %v, want nil", typ)
	}

	has := si.HasType(nil)
	if has {
		t.Error("HasType(nil) = true, want false")
	}

	// SetType with nil should not panic
	si.SetType(nil, &TypeAnnotation{Name: "Integer"})

	// Verify it was stored
	if !si.HasType(nil) {
		t.Error("HasType(nil) after SetType = false, want true")
	}
}

func TestSemanticInfo_OverwriteType(t *testing.T) {
	si := NewSemanticInfo()

	expr := &IntegerLiteral{
		TypedExpressionBase: TypedExpressionBase{
			BaseNode: BaseNode{
				Token: token.Token{Type: token.INT, Literal: "42", Pos: token.Position{Line: 1, Column: 1}},
			},
		},
		Value: 42,
	}

	// Set initial type
	type1 := &TypeAnnotation{Name: "Integer"}
	si.SetType(expr, type1)

	if typ := si.GetType(expr); typ != type1 {
		t.Errorf("GetType() = %v, want %v", typ, type1)
	}

	// Overwrite with new type
	type2 := &TypeAnnotation{Name: "Float"}
	si.SetType(expr, type2)

	if typ := si.GetType(expr); typ != type2 {
		t.Errorf("GetType() after overwrite = %v, want %v", typ, type2)
	}

	// Count should still be 1
	if count := si.TypeCount(); count != 1 {
		t.Errorf("TypeCount() after overwrite = %d, want 1", count)
	}
}
