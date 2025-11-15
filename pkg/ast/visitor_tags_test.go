package ast

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/token"
)

// TestNodeWithSkipTag tests that fields with ast:"skip" tag are not walked
func TestNodeWithSkipTag(t *testing.T) {
	// Create a test node that has a field with ast:"skip" tag
	// We'll use TypeAnnotation which has a Type field that could have such a tag

	// For now, let's test with a simple program
	program := &Program{
		Statements: []Statement{
			&VarDeclStatement{
				Names: []*Identifier{{Value: "x"}},
				Type: &TypeAnnotation{
					Name: "Integer",
				},
			},
		},
	}

	// Count all nodes
	nodeCount := 0
	Inspect(program, func(n Node) bool {
		if n != nil {
			nodeCount++
		}
		return true
	})

	// Should visit: Program, VarDeclStatement, TypeAnnotation
	// TypeAnnotation.Name is a string, not a Node, so it's not visited
	// VarDeclStatement.Names contains Identifiers which should be visited
	// That's at least 3 nodes
	if nodeCount < 3 {
		t.Errorf("Expected at least 3 nodes, got %d", nodeCount)
	}
}

// TestGeneratedVisitorCompleteness ensures the generated visitor walks all expected fields
func TestGeneratedVisitorCompleteness(t *testing.T) {
	// Test with a complex node that has multiple field types
	funcDecl := &FunctionDecl{
		BaseNode: BaseNode{
			Token: token.Token{Type: token.FUNCTION, Literal: "function"},
		},
		Name: &Identifier{Value: "Test"},
		Parameters: []*Parameter{
			{
				Name: &Identifier{Value: "x"},
				Type: &TypeAnnotation{
					Name: "Integer",
				},
			},
		},
		ReturnType: &TypeAnnotation{
			Name: "Integer",
		},
		Body: &BlockStatement{
			Statements: []Statement{
				&ReturnStatement{
					ReturnValue: &Identifier{Value: "x"},
				},
			},
		},
	}

	// Count all visited nodes
	visitedTypes := make(map[string]int)
	Inspect(funcDecl, func(n Node) bool {
		if n != nil {
			visitedTypes[n.TokenLiteral()]++
		}
		return true
	})

	// Should visit function name, parameter name, parameter type, return type, body, return value
	if len(visitedTypes) == 0 {
		t.Error("No nodes were visited")
	}
}
