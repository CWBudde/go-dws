package bytecode

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/interp"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// buildBenchmarkProgram returns a simple counting loop program.
func buildBenchmarkProgram() *ast.Program {
	intType := &ast.TypeAnnotation{Name: "Integer"}
	boolType := &ast.TypeAnnotation{Name: "Boolean"}

	return &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				Token: lexer.Token{Type: lexer.VAR, Literal: "var"},
				Names: []*ast.Identifier{
					{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
							},
							Type: intType,
						},
						Value: "x",
					},
				},
				Type: intType,
				Value: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.INT, Literal: "0"},
						},
						Type: intType,
					},
					Value: 0,
				},
			},
			&ast.WhileStatement{
				Token: lexer.Token{Type: lexer.WHILE, Literal: "while"},
				Condition: &ast.BinaryExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.LESS, Literal: "<"},
						},
						Type: boolType,
					},
					Operator: "<",
					Left: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
							},
							Type: intType,
						},
						Value: "x",
					},
					Right: &ast.IntegerLiteral{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.INT, Literal: "1000"},
							},
							Type: intType,
						},
						Value: 1000,
					},
				},
				Body: &ast.AssignmentStatement{
					Token:    lexer.Token{Type: lexer.IDENT, Literal: "x"},
					Operator: lexer.ASSIGN,
					Target: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
							},
							Type: intType,
						},
						Value: "x",
					},
					Value: &ast.BinaryExpression{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.PLUS, Literal: "+"},
							},
							Type: intType,
						},
						Operator: "+",
						Left: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
								},
								Type: intType,
							},
							Value: "x",
						},
						Right: &ast.IntegerLiteral{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: lexer.Token{Type: lexer.INT, Literal: "1"},
								},
								Type: intType,
							},
							Value: 1,
						},
					},
				},
			},
			&ast.ReturnStatement{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Result"},
				ReturnValue: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
						},
						Type: intType,
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
