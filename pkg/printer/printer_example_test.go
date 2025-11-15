package printer_test

import (
	"fmt"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/printer"
	"github.com/cwbudde/go-dws/pkg/token"
)

// Example demonstrates basic printer usage with DWScript format.
func Example() {
	// Create a simple AST: var x: Integer := 42
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				Names: []*ast.Identifier{
					{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: token.Token{Type: token.IDENT, Literal: "x"},
							},
						},
						Value: "x",
					},
				},
				Type: &ast.TypeAnnotation{
					Name: "Integer",
				},
				Value: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "42"},
						},
					},
					Value: 42,
				},
			},
		},
	}

	// Print using default options (DWScript format, detailed style)
	output := printer.Print(program)
	fmt.Println(output)

	// Output:
	// var x: Integer := 42
}

// Example_compactStyle demonstrates compact printing style.
func Example_compactStyle() {
	// Create a simple binary expression: 3 + 5
	expr := &ast.BinaryExpression{
		Left: &ast.IntegerLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.INT, Literal: "3"},
				},
			},
			Value: 3,
		},
		Operator: "+",
		Right: &ast.IntegerLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.INT, Literal: "5"},
				},
			},
			Value: 5,
		},
	}

	exprStmt := &ast.ExpressionStatement{
		Expression: expr,
	}

	program := &ast.Program{
		Statements: []ast.Statement{exprStmt},
	}

	// Create printer with compact style
	p := printer.New(printer.Options{
		Format: printer.FormatDWScript,
		Style:  printer.StyleCompact,
	})

	output := p.Print(program)
	fmt.Println(output)

	// Output:
	// 3 + 5
}

// Example_treeFormat demonstrates tree structure visualization.
func Example_treeFormat() {
	// Create a simple program
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				Names: []*ast.Identifier{
					{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: token.Token{Type: token.IDENT, Literal: "result"},
							},
						},
						Value: "result",
					},
				},
				Type: &ast.TypeAnnotation{
					Name: "Integer",
				},
				Value: &ast.BinaryExpression{
					Left: &ast.IntegerLiteral{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: token.Token{Type: token.INT, Literal: "10"},
							},
						},
						Value: 10,
					},
					Operator: "*",
					Right: &ast.IntegerLiteral{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: token.Token{Type: token.INT, Literal: "20"},
							},
						},
						Value: 20,
					},
				},
			},
		},
	}

	// Create printer with tree format
	p := printer.New(printer.Options{
		Format: printer.FormatTree,
	})

	output := p.Print(program)
	fmt.Println(output)

	// Output:
	// Program (1 statements)
	//   VarDeclStatement
}
