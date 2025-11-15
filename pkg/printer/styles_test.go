package printer_test

import (
	"fmt"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/printer"
	"github.com/cwbudde/go-dws/pkg/token"
)

// Example_compactPrinter demonstrates using the CompactPrinter preset.
func Example_compactPrinter() {
	// Create a simple program: var x: Integer := 42
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

	// Use the CompactPrinter preset
	p := printer.CompactPrinter()
	output := p.Print(program)
	fmt.Println(output)

	// Output:
	// var x:Integer:=42
}

// Example_detailedPrinter demonstrates using the DetailedPrinter preset.
func Example_detailedPrinter() {
	// Create a simple if statement
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.IfStatement{
				Condition: &ast.BooleanLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.TRUE, Literal: "true"},
						},
					},
					Value: true,
				},
				Consequence: &ast.BlockStatement{
					Statements: []ast.Statement{
						&ast.ExpressionStatement{
							Expression: &ast.IntegerLiteral{
								TypedExpressionBase: ast.TypedExpressionBase{
									BaseNode: ast.BaseNode{
										Token: token.Token{Type: token.INT, Literal: "1"},
									},
								},
								Value: 1,
							},
						},
					},
				},
			},
		},
	}

	// Use the DetailedPrinter preset
	p := printer.DetailedPrinter()
	output := p.Print(program)
	fmt.Println(output)

	// Output:
	// if true then
	// begin
	//   1;
	// end
}

// Example_treePrinter demonstrates using the TreePrinter preset.
func Example_treePrinter() {
	// Create a simple expression: 10 + 20
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.ExpressionStatement{
				Expression: &ast.BinaryExpression{
					Left: &ast.IntegerLiteral{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: token.Token{Type: token.INT, Literal: "10"},
							},
						},
						Value: 10,
					},
					Operator: "+",
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

	// Use the TreePrinter preset
	p := printer.TreePrinter()
	output := p.Print(program)
	fmt.Println(output)

	// Output:
	// Program (1 statements)
	//   ExpressionStatement
}

// Example_jsonPrinter demonstrates using the JSONPrinter preset.
func Example_jsonPrinter() {
	// Create a simple literal
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.ExpressionStatement{
				Expression: &ast.IntegerLiteral{
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

	// Use the JSONPrinter preset
	p := printer.JSONPrinter()
	output := p.Print(program)
	fmt.Println(output)

	// Output:
	// {
	//   "statements": [
	//     {
	//       "expression": {
	//         "type": "IntegerLiteral",
	//         "value": 42
	//       },
	//       "type": "ExpressionStatement"
	//     }
	//   ],
	//   "type": "Program"
	// }
}

// Example_optionsPresets demonstrates using option presets directly.
func Example_optionsPresets() {
	// Create a simple expression
	expr := &ast.IntegerLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Type: token.INT, Literal: "100"},
			},
		},
		Value: 100,
	}

	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.ExpressionStatement{Expression: expr},
		},
	}

	// Use CompactOptions
	p1 := printer.New(printer.CompactOptions())
	fmt.Println("Compact:", p1.Print(program))

	// Use DetailedOptions
	p2 := printer.New(printer.DetailedOptions())
	fmt.Println("Detailed:", p2.Print(program))

	// Use TreeOptions
	p3 := printer.New(printer.TreeOptions())
	fmt.Println("Tree:", p3.Print(program))

	// Output:
	// Compact: 100
	// Detailed: 100
	// Tree: Program (1 statements)
	//   ExpressionStatement
}
