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
						Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
						Value: "x",
						Type:  intType,
					},
				},
				Type: intType,
				Value: &ast.IntegerLiteral{
					Token: lexer.Token{Type: lexer.INT, Literal: "0"},
					Value: 0,
					Type:  intType,
				},
			},
			&ast.WhileStatement{
				Token: lexer.Token{Type: lexer.WHILE, Literal: "while"},
				Condition: &ast.BinaryExpression{
					Token:    lexer.Token{Type: lexer.LESS, Literal: "<"},
					Operator: "<",
					Left: &ast.Identifier{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
						Value: "x",
						Type:  intType,
					},
					Right: &ast.IntegerLiteral{
						Token: lexer.Token{Type: lexer.INT, Literal: "1000"},
						Value: 1000,
						Type:  intType,
					},
					Type: boolType,
				},
				Body: &ast.AssignmentStatement{
					Token:    lexer.Token{Type: lexer.IDENT, Literal: "x"},
					Operator: lexer.ASSIGN,
					Target: &ast.Identifier{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
						Value: "x",
						Type:  intType,
					},
					Value: &ast.BinaryExpression{
						Token:    lexer.Token{Type: lexer.PLUS, Literal: "+"},
						Operator: "+",
						Left: &ast.Identifier{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
							Value: "x",
							Type:  intType,
						},
						Right: &ast.IntegerLiteral{
							Token: lexer.Token{Type: lexer.INT, Literal: "1"},
							Value: 1,
							Type:  intType,
						},
						Type: intType,
					},
				},
			},
			&ast.ReturnStatement{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Result"},
				ReturnValue: &ast.Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
					Value: "x",
					Type:  intType,
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
