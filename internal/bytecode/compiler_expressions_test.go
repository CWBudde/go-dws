package bytecode

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

func TestCompiler_ArrayLiteralAndIndex(t *testing.T) {
	arrIdent := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "arr", Pos: pos(1, 1)},
			},
		},
		Value: "arr",
	}
	arrayLiteral := &ast.ArrayLiteralExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: lexer.Token{Type: lexer.LBRACK, Literal: "[", Pos: pos(1, 10)},
			},
		},
		Elements: []ast.Expression{
			&ast.IntegerLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(1, 11)},
					},
				},
				Value: 1,
			},
			&ast.IntegerLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.INT, Literal: "2", Pos: pos(1, 14)},
					},
				},
				Value: 2,
			},
		},
	}
	assignStmt := &ast.AssignmentStatement{
		BaseNode: ast.BaseNode{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "arr", Pos: pos(2, 1)},
		},
		Operator: lexer.ASSIGN,
		Target: &ast.IndexExpression{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.LBRACK, Literal: "[", Pos: pos(2, 4)},
				},
			},
			Left: &ast.Identifier{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "arr", Pos: pos(2, 1)},
					},
				},
				Value: "arr",
			},
			Index: &ast.IntegerLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(2, 8)},
					},
				},
				Value: 1,
			},
		},
		Value: &ast.IntegerLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.INT, Literal: "5", Pos: pos(2, 13)},
				},
			},
			Value: 5,
		},
	}
	returnStmt := &ast.ReturnStatement{
		Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(3, 1)},
		ReturnValue: &ast.IndexExpression{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.LBRACK, Literal: "[", Pos: pos(3, 10)},
				},
			},
			Left: &ast.Identifier{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "arr", Pos: pos(3, 8)},
					},
				},
				Value: "arr",
			},
			Index: &ast.IntegerLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.INT, Literal: "0", Pos: pos(3, 12)},
					},
				},
				Value: 0,
			},
		},
	}

	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(1, 1)},
				},
				Names: []*ast.Identifier{arrIdent},
				Value: arrayLiteral,
			},
			assignStmt,
			returnStmt,
		},
	}

	compiler := newTestCompiler("array_index")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	expectInstructions(t, chunk, []expectedInstruction{
		{Op: OpLoadConst0, A: 0, B: 0},
		{Op: OpLoadConst1, A: 0, B: 0},
		{Op: OpNewArray, A: 0, B: 2},
		{Op: OpStoreGlobal, A: 0, B: userGlobal0},
		{Op: OpLoadConst, A: 0, B: 2},
		{Op: OpLoadGlobal, A: 0, B: userGlobal0},
		{Op: OpLoadConst0, A: 0, B: 0},
		{Op: OpRotate3, A: 0, B: 0},
		{Op: OpArraySet, A: 0, B: 0},
		{Op: OpLoadGlobal, A: 0, B: userGlobal0},
		{Op: OpLoadConst, A: 0, B: 3},
		{Op: OpArrayGet, A: 0, B: 0},
		{Op: OpReturn, A: 1, B: 0},
	})

	if chunk.ConstantCount() != 4 {
		t.Fatalf("ConstantCount() = %d, want 4", chunk.ConstantCount())
	}
	if got := chunk.GetConstant(0).AsInt(); got != 1 {
		t.Errorf("Constant 0 = %d, want 1", got)
	}
	if got := chunk.GetConstant(1).AsInt(); got != 2 {
		t.Errorf("Constant 1 = %d, want 2", got)
	}
	if got := chunk.GetConstant(2).AsInt(); got != 5 {
		t.Errorf("Constant 2 = %d, want 5", got)
	}
	if got := chunk.GetConstant(3).AsInt(); got != 0 {
		t.Errorf("Constant 3 = %d, want 0", got)
	}
}

func TestCompiler_NewExpression(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.ReturnStatement{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(1, 1)},
				ReturnValue: &ast.NewExpression{
					Token: lexer.Token{Type: lexer.NEW, Literal: "new", Pos: pos(1, 9)},
					ClassName: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "TPoint", Pos: pos(1, 13)},
							},
						},
						Value: "TPoint",
					},
				},
			},
		},
	}

	compiler := newTestCompiler("new_expr")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	expectInstructions(t, chunk, []expectedInstruction{
		{Op: OpNewObject, A: 0, B: 0},
		{Op: OpReturn, A: 1, B: 0},
	})

	if chunk.ConstantCount() != 1 {
		t.Fatalf("ConstantCount() = %d, want 1", chunk.ConstantCount())
	}
	if got := chunk.GetConstant(0).AsString(); got != "TPoint" {
		t.Fatalf("Constant 0 = %q, want TPoint", got)
	}
}
func TestCompiler_ConstantFolding(t *testing.T) {
	intType := &ast.TypeAnnotation{Name: "Integer"}

	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.ReturnStatement{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(1, 1)},
				ReturnValue: &ast.BinaryExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.PLUS, Literal: "+", Pos: pos(1, 9)},
						},
						Type: intType,
					},
					Operator: "+",
					Left: &ast.IntegerLiteral{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.INT, Literal: "2", Pos: pos(1, 8)},
							},
							Type: intType,
						},
						Value: 2,
					},
					Right: &ast.IntegerLiteral{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.INT, Literal: "3", Pos: pos(1, 12)},
							},
							Type: intType,
						},
						Value: 3,
					},
				},
			},
		},
	}

	compiler := newTestCompiler("fold_test")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	expectInstructions(t, chunk, []expectedInstruction{
		{Op: OpLoadConst0, A: 0, B: 0},
		{Op: OpReturn, A: 1, B: 0},
	})

	if chunk.ConstantCount() != 1 {
		t.Fatalf("ConstantCount() = %d, want 1", chunk.ConstantCount())
	}
	if got := chunk.GetConstant(0).AsInt(); got != 5 {
		t.Errorf("Constant 0 = %d, want 5", got)
	}
}

func TestCompiler_CallExpression(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(1, 1)},
				},
				Names: []*ast.Identifier{
					{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "adder", Pos: pos(1, 5)},
							},
						},
						Value: "adder",
					},
				},
			},
			&ast.ReturnStatement{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(2, 1)},
				ReturnValue: &ast.CallExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.LPAREN, Literal: "(", Pos: pos(2, 11)},
						},
					},
					Function: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "adder", Pos: pos(2, 9)},
							},
						},
						Value: "adder",
					},
					Arguments: []ast.Expression{
						&ast.IntegerLiteral{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(2, 15)},
								},
							},
							Value: 1,
						},
						&ast.IntegerLiteral{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: lexer.Token{Type: lexer.INT, Literal: "2", Pos: pos(2, 18)},
								},
							},
							Value: 2,
						},
					},
				},
			},
		},
	}

	compiler := newTestCompiler("call_test")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	expectInstructions(t, chunk, []expectedInstruction{
		{Op: OpLoadNil, A: 0, B: 0},
		{Op: OpStoreGlobal, A: 0, B: userGlobal0},
		{Op: OpLoadGlobal, A: 0, B: userGlobal0},
		{Op: OpLoadConst0, A: 0, B: 0},
		{Op: OpLoadConst1, A: 0, B: 0},
		{Op: OpCallIndirect, A: 2, B: 0},
		{Op: OpReturn, A: 1, B: 0},
	})

	if chunk.ConstantCount() != 2 {
		t.Fatalf("ConstantCount() = %d, want 2", chunk.ConstantCount())
	}
	if got := chunk.GetConstant(0).AsInt(); got != 1 {
		t.Errorf("Constant 0 = %d, want 1", got)
	}
	if got := chunk.GetConstant(1).AsInt(); got != 2 {
		t.Errorf("Constant 1 = %d, want 2", got)
	}
}
func TestCompiler_MemberAccess(t *testing.T) {
	objIdent := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "obj", Pos: pos(1, 5)},
			},
		},
		Value: "obj",
	}
	memberName := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "value", Pos: pos(2, 8)},
			},
		},
		Value: "value",
	}

	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(1, 1)},
				},
				Names: []*ast.Identifier{objIdent},
			},
			&ast.AssignmentStatement{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "obj", Pos: pos(2, 1)},
				},
				Operator: lexer.ASSIGN,
				Target: &ast.MemberAccessExpression{
					Token: lexer.Token{Type: lexer.DOT, Literal: ".", Pos: pos(2, 6)},
					Object: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "obj", Pos: pos(2, 1)},
							},
						},
						Value: "obj",
					},
					Member: memberName,
				},
				Value: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.INT, Literal: "42", Pos: pos(2, 15)},
						},
					},
					Value: 42,
				},
			},
			&ast.ReturnStatement{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(3, 1)},
				ReturnValue: &ast.MemberAccessExpression{
					Token: lexer.Token{Type: lexer.DOT, Literal: ".", Pos: pos(3, 6)},
					Object: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "obj", Pos: pos(3, 1)},
							},
						},
						Value: "obj",
					},
					Member: memberName,
				},
			},
		},
	}

	compiler := newTestCompiler("member_test")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	expectInstructions(t, chunk, []expectedInstruction{
		{Op: OpLoadNil, A: 0, B: 0},
		{Op: OpStoreGlobal, A: 0, B: userGlobal0},
		{Op: OpLoadGlobal, A: 0, B: userGlobal0},
		{Op: OpLoadConst0, A: 0, B: 0},
		{Op: OpSetProperty, A: 0, B: 1},
		{Op: OpLoadGlobal, A: 0, B: userGlobal0},
		{Op: OpGetProperty, A: 0, B: 1},
		{Op: OpReturn, A: 1, B: 0},
	})

	if chunk.ConstantCount() != 2 {
		t.Fatalf("ConstantCount() = %d, want 2", chunk.ConstantCount())
	}
	if got := chunk.GetConstant(0).AsInt(); got != 42 {
		t.Errorf("Constant 0 = %d, want 42", got)
	}
	if got := chunk.GetConstant(1).AsString(); got != "value" {
		t.Errorf("Constant 1 = %q, want \"value\"", got)
	}
}

func TestCompiler_MethodCallEmitsCallMethod(t *testing.T) {
	intType := &ast.TypeAnnotation{Name: "Integer"}

	objIdent := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "obj", Pos: pos(1, 5)},
			},
			Type: intType,
		},
		Value: "obj",
	}
	methodName := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "DoIt", Pos: pos(2, 10)},
			},
		},
		Value: "DoIt",
	}

	methodCall := &ast.MethodCallExpression{
		Token: lexer.Token{Type: lexer.DOT, Literal: ".", Pos: pos(2, 8)},
		Object: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "obj", Pos: pos(2, 5)},
				},
				Type: intType,
			},
			Value: "obj",
		},
		Method: methodName,
		Arguments: []ast.Expression{
			&ast.IntegerLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(2, 15)},
					},
					Type: intType,
				},
				Value: 1,
			},
		},
	}

	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(1, 1)}},
				Names:    []*ast.Identifier{objIdent},
				Type:     intType,
				Value: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.INT, Literal: "0", Pos: pos(1, 10)},
						},
						Type: intType,
					},
					Value: 0,
				},
			},
			&ast.ExpressionStatement{
				BaseNode:   ast.BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "DoIt", Pos: pos(2, 5)}},
				Expression: methodCall,
			},
		},
	}

	chunk := compileProgram(t, program)

	foundCallMethod := false
	for _, inst := range chunk.Code {
		if inst.OpCode() == OpCallMethod {
			foundCallMethod = true
			break
		}
	}

	if !foundCallMethod {
		t.Fatalf("expected OpCallMethod in compiled chunk")
	}
}

func TestCompiler_SelfIdentifierEmitsGetSelf(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.ExpressionStatement{
				BaseNode: ast.BaseNode{Token: lexer.Token{Type: lexer.IDENT, Literal: "Self", Pos: pos(1, 1)}},
				Expression: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "Self", Pos: pos(1, 1)},
						},
					},
					Value: "Self",
				},
			},
		},
	}

	chunk := compileProgram(t, program)
	if len(chunk.Code) == 0 || chunk.Code[0].OpCode() != OpGetSelf {
		t.Fatalf("expected first instruction to be OpGetSelf, got %v", chunk.Code)
	}
}
