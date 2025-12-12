package astutil

import "github.com/cwbudde/go-dws/pkg/ast"

// CollectIndices flattens nested IndexExpression nodes for multi-index access.
// The parser converts multi-index syntax like obj.Data[1, 2] into nested IndexExpression nodes:
//
//	((obj.Data)[1])[2]
//
// This function returns:
//   - base: the underlying expression being indexed (e.g., obj.Data)
//   - indices: all index expressions in left-to-right order (e.g., [1, 2])
func CollectIndices(expr *ast.IndexExpression) (base ast.Expression, indices []ast.Expression) {
	indices = make([]ast.Expression, 0, 4)
	current := expr

	for {
		indices = append([]ast.Expression{current.Index}, indices...)

		if leftIndex, ok := current.Left.(*ast.IndexExpression); ok {
			current = leftIndex
			continue
		}

		base = current.Left
		break
	}

	return base, indices
}
