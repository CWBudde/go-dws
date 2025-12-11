package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
)

// TestIndexedPropertyAssignment tests indexed property assignment: obj.Prop[i] := value
func TestIndexedPropertyAssignment(t *testing.T) {
	t.Skip("Indexed property assignment requires mock object with indexed property - implement when fixture tests available")

	// This test is skipped because it requires:
	// 1. A mock object that implements PropertyAccessor
	// 2. A mock indexed property with setter method
	// 3. Proper adapter.ExecuteMethodWithSelf implementation
	//
	// The functionality will be tested via fixture tests that have real
	// DWScript class definitions with indexed properties.
	//
	// Example fixture test would use code like:
	//   type TList = class
	//     property Items[Index: Integer]: String read GetItem write SetItem;
	//   end;
	//
	//   var list: TList;
	//   list.Items[0] := 'hello';  // This calls SetItem(0, 'hello')
}

// TestMultiIndexedPropertyAssignment tests multi-index property assignment
func TestMultiIndexedPropertyAssignment(t *testing.T) {
	t.Skip("Multi-index property assignment requires mock object - implement when fixture tests available")

	// Example:
	//   type TMatrix = class
	//     property Cells[Row, Col: Integer]: Float read GetCell write SetCell;
	//   end;
	//
	//   var m: TMatrix;
	//   m.Cells[1, 2] := 3.14;  // This calls SetCell(1, 2, 3.14)
}

// TestCollectIndices verifies that multi-index expressions are collected correctly
func TestCollectIndices(t *testing.T) {
	tests := []struct {
		name        string
		expr        *ast.IndexExpression
		wantIndices int
	}{
		{
			name: "single index",
			expr: &ast.IndexExpression{
				Left: &ast.Identifier{Value: "arr"},
				Index: &ast.IntegerLiteral{Value: 1},
			},
			wantIndices: 1,
		},
		{
			name: "double index",
			expr: &ast.IndexExpression{
				Left: &ast.IndexExpression{
					Left: &ast.Identifier{Value: "arr"},
					Index: &ast.IntegerLiteral{Value: 1},
				},
				Index: &ast.IntegerLiteral{Value: 2},
			},
			wantIndices: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base, indices := CollectIndices(tt.expr)

			if base == nil {
				t.Fatal("expected base, got nil")
			}

			if len(indices) != tt.wantIndices {
				t.Errorf("CollectIndices() got %d indices, want %d", len(indices), tt.wantIndices)
			}
		})
	}
}

// TestIndexedPropertyAssignment_ErrorCases tests error handling
func TestIndexedPropertyAssignment_ErrorCases(t *testing.T) {
	t.Skip("Error cases will be tested via fixture tests with real DWScript code")

	// Error cases that will be covered by fixture tests:
	// 1. Property not found on object
	// 2. Non-indexed property accessed with index
	// 3. Read-only indexed property
	// 4. Invalid property metadata
	// 5. Type mismatches in index or value
}
