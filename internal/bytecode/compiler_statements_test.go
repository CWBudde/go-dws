package bytecode

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

func TestCompiler_VarAssignReturn(t *testing.T) {
	intType := &ast.TypeAnnotation{Name: "Integer"}

	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(1, 1)},
				Names: []*ast.Identifier{
					{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(1, 5)},
							},
						},
						Value: "x",
					},
				},
				Type: intType,
				Value: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.INT, Literal: "42", Pos: pos(1, 10)},
						},
					},
					Value: 42,
				},
			},
			&ast.AssignmentStatement{
				Token:    lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(2, 1)},
				Operator: lexer.ASSIGN,
				Target: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(2, 1)},
						},
						Type: intType,
					},
					Value: "x",
				},
				Value: &ast.BinaryExpression{
					Token:    lexer.Token{Type: lexer.PLUS, Literal: "+", Pos: pos(2, 6)},
					Operator: "+",
					Left: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(2, 5)},
							},
							Type: intType,
						},
						Value: "x",
					},
					Right: &ast.IntegerLiteral{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(2, 10)},
							},
						},
						Value: 1,
					},
					Type: intType,
				},
			},
			&ast.ReturnStatement{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(3, 1)},
				ReturnValue: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(3, 9)},
						},
						Type: intType,
					},
					Value: "x",
				},
			},
		},
	}

	compiler := newTestCompiler("test_function")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	expectInstructions(t, chunk, []expectedInstruction{
		{Op: OpLoadConst0, A: 0, B: 0},
		{Op: OpStoreGlobal, A: 0, B: userGlobal0},
		{Op: OpLoadGlobal, A: 0, B: userGlobal0},
		{Op: OpLoadConst1, A: 0, B: 0},
		{Op: OpAddInt, A: 0, B: 0},
		{Op: OpStoreGlobal, A: 0, B: userGlobal0},
		{Op: OpLoadGlobal, A: 0, B: userGlobal0},
		{Op: OpReturn, A: 1, B: 0},
	})

	if chunk.ConstantCount() != 2 {
		t.Fatalf("ConstantCount() = %d, want 2", chunk.ConstantCount())
	}
	if got := chunk.GetConstant(0).AsInt(); got != 42 {
		t.Errorf("Constant 0 = %d, want 42", got)
	}
	if got := chunk.GetConstant(1).AsInt(); got != 1 {
		t.Errorf("Constant 1 = %d, want 1", got)
	}
}

func TestCompiler_IfElse(t *testing.T) {
	intType := &ast.TypeAnnotation{Name: "Integer"}
	boolType := &ast.TypeAnnotation{Name: "Boolean"}

	flagIdent := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "flag", Pos: pos(1, 5)},
			},
			Type: boolType,
		},
		Value: "flag",
	}

	totalIdent := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "total", Pos: pos(2, 5)},
			},
			Type: intType,
		},
		Value: "total",
	}

	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(1, 1)},
				Names: []*ast.Identifier{flagIdent},
				Type:  boolType,
				Value: &ast.BooleanLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.TRUE, Literal: "True", Pos: pos(1, 10)},
						},
					},
					Value: true,
				},
			},
			&ast.VarDeclStatement{
				Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(2, 1)},
				Names: []*ast.Identifier{totalIdent},
				Type:  intType,
				Value: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.INT, Literal: "0", Pos: pos(2, 10)},
						},
					},
					Value: 0,
				},
			},
			&ast.IfStatement{
				Token:     lexer.Token{Type: lexer.IF, Literal: "if", Pos: pos(3, 1)},
				Condition: flagIdent,
				Consequence: &ast.AssignmentStatement{
					Token:    lexer.Token{Type: lexer.IDENT, Literal: "total", Pos: pos(3, 8)},
					Operator: lexer.ASSIGN,
					Target: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "total", Pos: pos(3, 8)},
							},
							Type: intType,
						},
						Value: "total",
					},
					Value: &ast.IntegerLiteral{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(3, 17)},
							},
						},
						Value: 1,
					},
				},
				Alternative: &ast.AssignmentStatement{
					Token:    lexer.Token{Type: lexer.IDENT, Literal: "total", Pos: pos(4, 8)},
					Operator: lexer.ASSIGN,
					Target: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "total", Pos: pos(4, 8)},
							},
							Type: intType,
						},
						Value: "total",
					},
					Value: &ast.IntegerLiteral{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.INT, Literal: "2", Pos: pos(4, 17)},
							},
						},
						Value: 2,
					},
				},
			},
			&ast.ReturnStatement{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(5, 1)},
				ReturnValue: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "total", Pos: pos(5, 9)},
						},
						Type: intType,
					},
					Value: "total",
				},
			},
		},
	}

	compiler := newTestCompiler("test_if")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	expectInstructions(t, chunk, []expectedInstruction{
		{Op: OpLoadTrue, A: 0, B: 0},
		{Op: OpStoreGlobal, A: 0, B: userGlobal0},
		{Op: OpLoadConst0, A: 0, B: 0},
		{Op: OpStoreGlobal, A: 0, B: userGlobal1},
		{Op: OpLoadGlobal, A: 0, B: userGlobal0},
		{Op: OpJumpIfFalse, A: 0, B: 3},
		{Op: OpLoadConst1, A: 0, B: 0},
		{Op: OpStoreGlobal, A: 0, B: userGlobal1},
		{Op: OpJump, A: 0, B: 2},
		{Op: OpLoadConst, A: 0, B: 2},
		{Op: OpStoreGlobal, A: 0, B: userGlobal1},
		{Op: OpLoadGlobal, A: 0, B: userGlobal1},
		{Op: OpReturn, A: 1, B: 0},
	})

	if chunk.ConstantCount() != 3 {
		t.Fatalf("ConstantCount() = %d, want 3", chunk.ConstantCount())
	}
	if got := chunk.GetConstant(0).AsInt(); got != 0 {
		t.Errorf("Constant 0 = %d, want 0", got)
	}
	if got := chunk.GetConstant(1).AsInt(); got != 1 {
		t.Errorf("Constant 1 = %d, want 1", got)
	}
	if got := chunk.GetConstant(2).AsInt(); got != 2 {
		t.Errorf("Constant 2 = %d, want 2", got)
	}
}

func TestCompiler_TryExceptTypedHandler(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.TryStatement{
				Token: lexer.Token{Type: lexer.TRY, Literal: "try", Pos: pos(1, 1)},
				TryBlock: &ast.BlockStatement{Statements: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.NilLiteral{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: lexer.Token{Type: lexer.NIL, Literal: "Nil", Pos: pos(1, 5)},
								},
							},
						},
					},
				}},
				ExceptClause: &ast.ExceptClause{
					Token: lexer.Token{Type: lexer.EXCEPT, Literal: "except", Pos: pos(2, 1)},
					Handlers: []*ast.ExceptionHandler{
						{
							Token: lexer.Token{Type: lexer.ON, Literal: "on", Pos: pos(2, 3)},
							Variable: &ast.Identifier{
								TypedExpressionBase: ast.TypedExpressionBase{
									BaseNode: ast.BaseNode{
										Token: lexer.Token{Type: lexer.IDENT, Literal: "E", Pos: pos(2, 6)},
									},
								},
								Value: "E",
							},
							ExceptionType: &ast.TypeAnnotation{Name: "MyError"},
							Statement:     &ast.BlockStatement{},
						},
					},
				},
			},
		},
	}

	compiler := newTestCompiler("typed_catch")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	foundGetClass := false
	for _, inst := range chunk.Code {
		if inst.OpCode() == OpGetClass {
			foundGetClass = true
			break
		}
	}
	if !foundGetClass {
		t.Fatalf("expected OpGetClass in emitted bytecode")
	}
}

func TestCompiler_TryExceptRethrowWithoutElse(t *testing.T) {
	assign := &ast.AssignmentStatement{
		Token:    lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(3, 3)},
		Operator: lexer.ASSIGN,
		Target: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(3, 3)},
				},
			},
			Value: "x",
		},
		Value: &ast.IntegerLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(3, 10)},
				},
			},
			Value: 1,
		},
	}

	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(1, 1)},
				Names: []*ast.Identifier{{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(1, 5)},
						},
					},
					Value: "x",
				}},
			},
			&ast.TryStatement{
				Token: lexer.Token{Type: lexer.TRY, Literal: "try", Pos: pos(2, 1)},
				TryBlock: &ast.BlockStatement{Statements: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.NilLiteral{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: lexer.Token{Type: lexer.NIL, Literal: "Nil", Pos: pos(2, 5)},
								},
							},
						},
					},
				}},
				ExceptClause: &ast.ExceptClause{
					Token: lexer.Token{Type: lexer.EXCEPT, Literal: "except", Pos: pos(3, 1)},
					Handlers: []*ast.ExceptionHandler{
						{
							Token:         lexer.Token{Type: lexer.ON, Literal: "on", Pos: pos(3, 3)},
							ExceptionType: &ast.TypeAnnotation{Name: "Other"},
							Statement:     &ast.BlockStatement{Statements: []ast.Statement{assign}},
						},
					},
				},
			},
		},
	}

	compiler := newTestCompiler("typed_rethrow")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	foundThrow := false
	for _, inst := range chunk.Code {
		if inst.OpCode() == OpThrow {
			foundThrow = true
			break
		}
	}
	if !foundThrow {
		t.Fatalf("expected OpThrow for unmatched typed handler")
	}
}

func TestCompiler_RaiseStatementExpression(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.RaiseStatement{
				Token: lexer.Token{Type: lexer.RAISE, Literal: "raise", Pos: pos(1, 1)},
				Exception: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.INT, Literal: "5", Pos: pos(1, 7)},
						},
					},
					Value: 5,
				},
			},
		},
	}

	compiler := newTestCompiler("raise_expr")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	expectInstructions(t, chunk, []expectedInstruction{
		{Op: OpLoadConst0, A: 0, B: 0},
		{Op: OpThrow, A: 0, B: 0},
		{Op: OpHalt, A: 0, B: 0},
	})
}

func TestCompiler_RaiseStatementBare(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.RaiseStatement{
				Token: lexer.Token{Type: lexer.RAISE, Literal: "raise", Pos: pos(1, 1)},
			},
		},
	}

	compiler := newTestCompiler("raise_bare")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	expectInstructions(t, chunk, []expectedInstruction{
		{Op: OpLoadGlobal, A: 0, B: builtinExceptObjectIndex},
		{Op: OpThrow, A: 0, B: 0},
		{Op: OpHalt, A: 0, B: 0},
	})
}

func TestCompiler_TryFinallyMetadata(t *testing.T) {
	ident := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(1, 5)},
			},
		},
		Value: "x",
	}

	assignTry := &ast.AssignmentStatement{
		Token:    lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(2, 3)},
		Operator: lexer.ASSIGN,
		Target: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(2, 3)},
				},
			},
			Value: "x",
		},
		Value: &ast.IntegerLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(2, 8)},
				},
			},
			Value: 1,
		},
	}

	assignFinally := &ast.AssignmentStatement{
		Token:    lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(3, 3)},
		Operator: lexer.ASSIGN,
		Target: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(3, 3)},
				},
			},
			Value: "x",
		},
		Value: &ast.IntegerLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.INT, Literal: "2", Pos: pos(3, 8)},
				},
			},
			Value: 2,
		},
	}

	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(1, 1)},
				Names: []*ast.Identifier{ident},
				Value: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.INT, Literal: "0", Pos: pos(1, 10)},
						},
					},
					Value: 0,
				},
			},
			&ast.TryStatement{
				Token:    lexer.Token{Type: lexer.TRY, Literal: "try", Pos: pos(2, 1)},
				TryBlock: &ast.BlockStatement{Statements: []ast.Statement{assignTry}},
				FinallyClause: &ast.FinallyClause{
					Token: lexer.Token{Type: lexer.FINALLY, Literal: "finally", Pos: pos(3, 1)},
					Block: &ast.BlockStatement{Statements: []ast.Statement{assignFinally}},
				},
			},
		},
	}

	compiler := newTestCompiler("try_finally_meta")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	tryIndex := -1
	for idx, inst := range chunk.Code {
		if inst.OpCode() == OpTry {
			tryIndex = idx
			break
		}
	}
	if tryIndex == -1 {
		t.Fatalf("no OpTry emitted")
	}

	info, ok := chunk.TryInfoAt(tryIndex)
	if !ok {
		t.Fatalf("TryInfo missing")
	}
	if !info.HasFinally {
		t.Fatalf("expected HasFinally")
	}
	if info.CatchTarget != -1 {
		t.Fatalf("expected no catch target")
	}
	if info.FinallyTarget <= tryIndex {
		t.Fatalf("invalid finally target: %d", info.FinallyTarget)
	}
}

func TestCompiler_BreakAndContinue(t *testing.T) {
	boolType := &ast.TypeAnnotation{Name: "Boolean"}

	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.WhileStatement{
				Token: lexer.Token{Type: lexer.WHILE, Literal: "while", Pos: pos(1, 1)},
				Condition: &ast.BooleanLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.TRUE, Literal: "True", Pos: pos(1, 7)},
						},
						Type: boolType,
					},
					Value: true,
				},
				Body: &ast.BlockStatement{
					Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"},
					Statements: []ast.Statement{
						&ast.ContinueStatement{Token: lexer.Token{Type: lexer.CONTINUE, Literal: "continue", Pos: pos(2, 3)}},
						&ast.BreakStatement{Token: lexer.Token{Type: lexer.BREAK, Literal: "break", Pos: pos(3, 3)}},
					},
				},
			},
		},
	}

	chunk := compileProgram(t, program)

	foundLoop := false
	foundPatchedBreak := false
	for _, inst := range chunk.Code {
		switch inst.OpCode() {
		case OpLoop:
			foundLoop = true
		case OpJump:
			if inst.B() != 0xFFFF {
				foundPatchedBreak = true
			}
		}
	}

	if !foundLoop {
		t.Fatalf("expected OpLoop instruction for continue")
	}
	if !foundPatchedBreak {
		t.Fatalf("expected patched OpJump for break")
	}
}

func TestCompiler_RepeatContinuePatchesToCondition(t *testing.T) {
	boolType := &ast.TypeAnnotation{Name: "Boolean"}

	repeatStmt := &ast.RepeatStatement{
		Token: lexer.Token{Type: lexer.REPEAT, Literal: "repeat", Pos: pos(1, 1)},
		Body: &ast.BlockStatement{
			Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"},
			Statements: []ast.Statement{
				&ast.ContinueStatement{Token: lexer.Token{Type: lexer.CONTINUE, Literal: "continue", Pos: pos(2, 3)}},
			},
		},
		Condition: &ast.BooleanLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.TRUE, Literal: "True", Pos: pos(3, 1)},
				},
				Type: boolType,
			},
			Value: true,
		},
	}

	program := &ast.Program{
		Statements: []ast.Statement{repeatStmt},
	}

	chunk := compileProgram(t, program)

	hasPatchedJump := false
	for _, inst := range chunk.Code {
		if inst.OpCode() == OpJump && inst.B() != 0xFFFF {
			hasPatchedJump = true
			break
		}
	}

	if !hasPatchedJump {
		t.Fatalf("expected repeat continue jump to be patched")
	}
}

func TestCompiler_OptimizesLiteralExpressionStatements(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.ExpressionStatement{
				Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(1, 1)},
				Expression: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(1, 1)},
						},
					},
					Value: 1,
				},
			},
			&ast.ExpressionStatement{
				Token: lexer.Token{Type: lexer.TRUE, Literal: "True", Pos: pos(2, 1)},
				Expression: &ast.BooleanLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.TRUE, Literal: "True", Pos: pos(2, 1)},
						},
					},
					Value: true,
				},
			},
		},
	}

	compiler := NewCompiler("literal_opt")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	if len(chunk.Code) != 1 {
		t.Fatalf("expected only OpHalt after optimizing literal expressions, got %d instructions", len(chunk.Code))
	}
	if chunk.Code[0].OpCode() != OpHalt {
		t.Fatalf("expected remaining instruction to be OpHalt, got %v", chunk.Code[0].OpCode())
	}
}

func TestCompiler_RepeatLoop(t *testing.T) {
	intType := &ast.TypeAnnotation{Name: "Integer"}
	boolType := &ast.TypeAnnotation{Name: "Boolean"}

	xIdent := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(1, 5)},
			},
			Type: intType,
		},
		Value: "x",
	}

	repeatStmt := &ast.RepeatStatement{
		Token: lexer.Token{Type: lexer.REPEAT, Literal: "repeat", Pos: pos(2, 1)},
		Body: &ast.BlockStatement{
			Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin", Pos: pos(2, 8)},
			Statements: []ast.Statement{
				&ast.AssignmentStatement{
					Token:    lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(3, 3)},
					Operator: lexer.ASSIGN,
					Target: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(3, 3)},
							},
							Type: intType,
						},
						Value: "x",
					},
					Value: &ast.BinaryExpression{
						Token:    lexer.Token{Type: lexer.PLUS, Literal: "+", Pos: pos(3, 8)},
						Operator: "+",
						Left: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(3, 7)},
								},
								Type: intType,
							},
							Value: "x",
						},
						Right: &ast.IntegerLiteral{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(3, 11)},
								},
							},
							Value: 1,
						},
						Type: intType,
					},
				},
			},
		},
		Condition: &ast.BinaryExpression{
			Token:    lexer.Token{Type: lexer.GREATER, Literal: ">", Pos: pos(4, 10)},
			Operator: ">",
			Left: &ast.Identifier{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(4, 9)},
					},
					Type: intType,
				},
				Value: "x",
			},
			Right: &ast.IntegerLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.INT, Literal: "3", Pos: pos(4, 13)},
					},
				},
				Value: 3,
			},
			Type: boolType,
		},
	}

	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(1, 1)},
				Names: []*ast.Identifier{xIdent},
				Type:  intType,
				Value: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.INT, Literal: "0", Pos: pos(1, 10)},
						},
					},
					Value: 0,
				},
			},
			repeatStmt,
			&ast.ReturnStatement{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(5, 1)},
				ReturnValue: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(5, 9)},
						},
						Type: intType,
					},
					Value: "x",
				},
			},
		},
	}

	compiler := newTestCompiler("repeat_test")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	// Verify key opcodes are emitted with expected structure.
	opcodes := make([]OpCode, len(chunk.Code))
	for i, inst := range chunk.Code {
		opcodes[i] = inst.OpCode()
	}

	expectedOpcodes := []OpCode{
		OpLoadConst0,
		OpStoreGlobal,
		OpLoadGlobal,
		OpLoadConst1,
		OpAddInt,
		OpStoreGlobal,
		OpLoadGlobal,
		OpLoadConst,
		OpGreater,
		OpJumpIfTrue,
		OpLoop,
		OpLoadGlobal,
		OpReturn,
	}

	if len(opcodes) != len(expectedOpcodes) {
		t.Fatalf("opcode count = %d, want %d", len(opcodes), len(expectedOpcodes))
	}

	for i, op := range expectedOpcodes {
		if opcodes[i] != op {
			t.Fatalf("opcode %d = %v, want %v", i, opcodes[i], op)
		}
	}

	// Validate jump offsets.
	jumpInst := chunk.Code[9]
	if jumpInst.B() != 1 {
		t.Fatalf("repeat jump offset = %d, want 1", jumpInst.B())
	}
	loopInst := chunk.Code[10]
	if loopInst.SignedB() >= 0 {
		t.Fatalf("loop offset should be negative, got %d", loopInst.SignedB())
	}
}
