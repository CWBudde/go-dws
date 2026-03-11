package interp

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// getTypeFromAnnotation converts a type annotation to a types.Type
// This is a helper to extract type information from AST
func (i *Interpreter) getTypeFromAnnotation(typeExpr ast.TypeExpression) types.Type {
	if typeExpr == nil {
		return nil
	}

	// Get the type name from the type expression
	typeName := typeExpr.String()
	return i.getTypeByName(typeName)
}

// getTypeByName looks up a type by name
func (i *Interpreter) getTypeByName(name string) types.Type {
	switch name {
	case "Integer":
		return &types.IntegerType{}
	case "Float":
		return &types.FloatType{}
	case "String":
		return &types.StringType{}
	case "Boolean":
		return &types.BooleanType{}
	default:
		// Try to find in type registry (for custom types)
		// For now, return integer as placeholder
		return &types.IntegerType{}
	}
}
