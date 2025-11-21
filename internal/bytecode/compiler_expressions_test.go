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
		BaseNode: ast.BaseNode{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(3, 1)},
		},
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
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(1, 1)},
				},
				ReturnValue: &ast.NewExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.NEW, Literal: "new", Pos: pos(1, 9)},
						},
					},
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
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.ReturnStatement{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(1, 1)},
				},
				ReturnValue: &ast.BinaryExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.PLUS, Literal: "+", Pos: pos(1, 9)},
						},
					},
					Operator: "+",
					Left: &ast.IntegerLiteral{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.INT, Literal: "2", Pos: pos(1, 8)},
							},
						},
						Value: 2,
					},
					Right: &ast.IntegerLiteral{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.INT, Literal: "3", Pos: pos(1, 12)},
							},
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
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(2, 1)},
				},
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
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.DOT, Literal: ".", Pos: pos(2, 6)},
						},
					},
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
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(3, 1)},
				},
				ReturnValue: &ast.MemberAccessExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.DOT, Literal: ".", Pos: pos(3, 6)},
						},
					},
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
	objIdent := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "obj", Pos: pos(1, 5)},
			},
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
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: lexer.Token{Type: lexer.DOT, Literal: ".", Pos: pos(2, 8)},
			},
		},
		Object: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "obj", Pos: pos(2, 5)},
				},
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
				Value: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.INT, Literal: "0", Pos: pos(1, 10)},
						},
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

// TestCompiler_SetLiteral tests compilation of set literals (Task 3.5.29).
func TestCompiler_SetLiteral(t *testing.T) {
	tests := []struct {
		name             string
		setLiteral       *ast.SetLiteral
		expectedElements int // Number of elements that should be pushed onto stack
	}{
		{
			name: "empty set",
			setLiteral: &ast.SetLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.LBRACK, Literal: "[", Pos: pos(1, 1)},
					},
				},
				Elements: []ast.Expression{},
			},
			expectedElements: 0,
		},
		{
			name: "simple set with integers",
			setLiteral: &ast.SetLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.LBRACK, Literal: "[", Pos: pos(1, 1)},
					},
				},
				Elements: []ast.Expression{
					&ast.IntegerLiteral{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(1, 2)},
							},
						},
						Value: 1,
					},
					&ast.IntegerLiteral{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.INT, Literal: "2", Pos: pos(1, 5)},
							},
						},
						Value: 2,
					},
					&ast.IntegerLiteral{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.INT, Literal: "3", Pos: pos(1, 8)},
							},
						},
						Value: 3,
					},
				},
			},
			expectedElements: 3,
		},
		{
			name: "set with integer range",
			setLiteral: &ast.SetLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.LBRACK, Literal: "[", Pos: pos(1, 1)},
					},
				},
				Elements: []ast.Expression{
					&ast.RangeExpression{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(1, 2)},
							},
						},
						Start: &ast.IntegerLiteral{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(1, 2)},
								},
							},
							Value: 1,
						},
						RangeEnd: &ast.IntegerLiteral{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: lexer.Token{Type: lexer.INT, Literal: "5", Pos: pos(1, 5)},
								},
							},
							Value: 5,
						},
					},
				},
			},
			expectedElements: 5, // Range 1..5 expands to [1, 2, 3, 4, 5]
		},
		{
			name: "set with character range",
			setLiteral: &ast.SetLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.LBRACK, Literal: "[", Pos: pos(1, 1)},
					},
				},
				Elements: []ast.Expression{
					&ast.RangeExpression{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.CHAR, Literal: "'a'", Pos: pos(1, 2)},
							},
						},
						Start: &ast.CharLiteral{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: lexer.Token{Type: lexer.CHAR, Literal: "'a'", Pos: pos(1, 2)},
								},
							},
							Value: 'a',
						},
						RangeEnd: &ast.CharLiteral{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: lexer.Token{Type: lexer.CHAR, Literal: "'c'", Pos: pos(1, 7)},
								},
							},
							Value: 'c',
						},
					},
				},
			},
			expectedElements: 3, // Range 'a'..'c' expands to ['a', 'b', 'c']
		},
		{
			name: "set with mixed elements and range",
			setLiteral: &ast.SetLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.LBRACK, Literal: "[", Pos: pos(1, 1)},
					},
				},
				Elements: []ast.Expression{
					&ast.IntegerLiteral{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.INT, Literal: "0", Pos: pos(1, 2)},
							},
						},
						Value: 0,
					},
					&ast.RangeExpression{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.INT, Literal: "2", Pos: pos(1, 5)},
							},
						},
						Start: &ast.IntegerLiteral{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: lexer.Token{Type: lexer.INT, Literal: "2", Pos: pos(1, 5)},
								},
							},
							Value: 2,
						},
						RangeEnd: &ast.IntegerLiteral{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: lexer.Token{Type: lexer.INT, Literal: "4", Pos: pos(1, 9)},
								},
							},
							Value: 4,
						},
					},
					&ast.IntegerLiteral{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.INT, Literal: "10", Pos: pos(1, 13)},
							},
						},
						Value: 10,
					},
				},
			},
			expectedElements: 5, // [0, 2..4, 10] expands to [0, 2, 3, 4, 10]
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := &ast.Program{
				Statements: []ast.Statement{
					&ast.ExpressionStatement{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.LBRACK, Literal: "[", Pos: pos(1, 1)},
						},
						Expression: tt.setLiteral,
					},
				},
			}

			chunk := compileProgram(t, program)

			// Verify that OpNewSet is emitted with correct element count
			foundNewSet := false
			for _, inst := range chunk.Code {
				if inst.OpCode() == OpNewSet {
					foundNewSet = true
					actualCount := int(inst.B())
					if actualCount != tt.expectedElements {
						t.Errorf("OpNewSet element count = %d, want %d", actualCount, tt.expectedElements)
					}
					break
				}
			}

			if !foundNewSet {
				t.Errorf("expected OpNewSet instruction not found in compiled chunk")
			}
		})
	}
}

// TestCompiler_SetLiteralRangeExpansion tests that ranges are properly expanded at compile time.
func TestCompiler_SetLiteralRangeExpansion(t *testing.T) {
	// Test that a range like [1..10] produces 10 LOAD_CONST instructions followed by NEW_SET with count=10
	setLiteral := &ast.SetLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: lexer.Token{Type: lexer.LBRACK, Literal: "[", Pos: pos(1, 1)},
			},
		},
		Elements: []ast.Expression{
			&ast.RangeExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(1, 2)},
					},
				},
				Start: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(1, 2)},
						},
					},
					Value: 1,
				},
				RangeEnd: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.INT, Literal: "10", Pos: pos(1, 6)},
						},
					},
					Value: 10,
				},
			},
		},
	}

	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.ExpressionStatement{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.LBRACK, Literal: "[", Pos: pos(1, 1)},
				},
				Expression: setLiteral,
			},
		},
	}

	chunk := compileProgram(t, program)

	// Count LOAD_CONST instructions (should be 10, one for each value in range 1..10)
	loadConstCount := 0
	for _, inst := range chunk.Code {
		op := inst.OpCode()
		if op == OpLoadConst || op == OpLoadConst0 || op == OpLoadConst1 {
			loadConstCount++
		}
	}

	if loadConstCount != 10 {
		t.Errorf("expected 10 load constant instructions for range 1..10, got %d", loadConstCount)
	}

	// Verify the constants are correct (1, 2, 3, ..., 10)
	if len(chunk.Constants) < 10 {
		t.Fatalf("expected at least 10 constants in pool, got %d", len(chunk.Constants))
	}
}
