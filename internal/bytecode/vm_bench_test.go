package bytecode

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// buildBenchmarkProgram returns a simple counting loop program.
func buildBenchmarkProgram() *ast.Program {
	return &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.VAR, Literal: "var"},
				},
				Names: []*ast.Identifier{
					{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
							},
						},
						Value: "x",
					},
				},
				Value: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.INT, Literal: "0"},
						},
					},
					Value: 0,
				},
			},
			&ast.WhileStatement{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.WHILE, Literal: "while"},
				},
				Condition: &ast.BinaryExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.LESS, Literal: "<"},
						},
					},
					Operator: "<",
					Left: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
							},
						},
						Value: "x",
					},
					Right: &ast.IntegerLiteral{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.INT, Literal: "1000"},
							},
						},
						Value: 1000,
					},
				},
				Body: &ast.AssignmentStatement{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
					},
					Operator: lexer.ASSIGN,
					Target: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
							},
						},
						Value: "x",
					},
					Value: &ast.BinaryExpression{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.PLUS, Literal: "+"},
							},
						},
						Operator: "+",
						Left: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
								},
							},
							Value: "x",
						},
						Right: &ast.IntegerLiteral{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: lexer.Token{Type: lexer.INT, Literal: "1"},
								},
							},
							Value: 1,
						},
					},
				},
			},
			&ast.ReturnStatement{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Result"},
				},
				ReturnValue: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
						},
					},
					Value: "x",
				},
			},
		},
	}
}

func BenchmarkVMVsInterpreter_CountLoop(b *testing.B) {
	program := buildBenchmarkProgram()

	compiler := NewCompiler("bench")
	chunk, err := compiler.Compile(program)
	if err != nil {
		b.Fatalf("compile error: %v", err)
	}

	b.Run("vm", func(b *testing.B) {
		vm := NewVM()
		for i := 0; i < b.N; i++ {
			if _, err := vm.Run(chunk); err != nil {
				b.Fatalf("vm run error: %v", err)
			}
		}
	})

	b.Run("interp", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			interpreter := interp.New(nil)
			interpreter.Eval(program)
		}
	})
}
