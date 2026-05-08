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
	typ, err := i.resolveType(name)
	if err != nil {
		return nil
	}
	return typ
}
