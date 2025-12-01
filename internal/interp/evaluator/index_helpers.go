package evaluator

import "github.com/cwbudde/go-dws/pkg/ast"

// ============================================================================
// Multi-Index Support
// ============================================================================
//
// Moved from interp/array.go to support evaluator package direct access.
// ============================================================================

// CollectIndices flattens nested IndexExpression nodes for multi-index properties.
// The parser converts multi-index syntax like obj.Data[1, 2] into nested IndexExpression nodes:
//
//	((obj.Data)[1])[2]
//
// This function walks the chain and extracts:
//   - base: The actual object.property being accessed (e.g., obj.Data)
//   - indices: All index expressions in order (e.g., [1, 2])
//
// This supports multi-dimensional indexed properties like:
//
//	property Cells[x, y: Integer]: Float read GetCell write SetCell;
func CollectIndices(expr *ast.IndexExpression) (base ast.Expression, indices []ast.Expression) {
	indices = make([]ast.Expression, 0, 4) // Most properties have ≤4 dimensions
	current := expr

	// Walk down the chain of nested IndexExpression nodes
	for {
		// Prepend this level's index to maintain left-to-right order
		// We prepend because we're traversing from outermost to innermost
		indices = append([]ast.Expression{current.Index}, indices...)

		// Check if Left is another IndexExpression (nested)
		if leftIndex, ok := current.Left.(*ast.IndexExpression); ok {
			current = leftIndex
			continue
		}

		// Found the base expression (e.g., obj.Property or arr)
		base = current.Left
		break
	}

	return base, indices
}
