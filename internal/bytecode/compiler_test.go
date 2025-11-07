package bytecode

import (
	"io"
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/interp"
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
						Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(1, 5)},
						Value: "x",
					},
				},
				Type: intType,
				Value: &ast.IntegerLiteral{
					Token: lexer.Token{Type: lexer.INT, Literal: "42", Pos: pos(1, 10)},
					Value: 42,
				},
			},
			&ast.AssignmentStatement{
				Token:    lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(2, 1)},
				Operator: lexer.ASSIGN,
				Target: &ast.Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(2, 1)},
					Value: "x",
					Type:  intType,
				},
				Value: &ast.BinaryExpression{
					Token:    lexer.Token{Type: lexer.PLUS, Literal: "+", Pos: pos(2, 6)},
					Operator: "+",
					Left: &ast.Identifier{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(2, 5)},
						Value: "x",
						Type:  intType,
					},
					Right: &ast.IntegerLiteral{
						Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(2, 10)},
						Value: 1,
					},
					Type: intType,
				},
			},
			&ast.ReturnStatement{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(3, 1)},
				ReturnValue: &ast.Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(3, 9)},
					Value: "x",
					Type:  intType,
				},
			},
		},
	}

	compiler := NewCompiler("test_function")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	expectInstructions(t, chunk, []expectedInstruction{
		{OpLoadConst0, 0, 0},
		{OpStoreGlobal, 0, 0},
		{OpLoadGlobal, 0, 0},
		{OpLoadConst1, 0, 0},
		{OpAddInt, 0, 0},
		{OpStoreGlobal, 0, 0},
		{OpLoadGlobal, 0, 0},
		{OpReturn, 1, 0},
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
		Token: lexer.Token{Type: lexer.IDENT, Literal: "flag", Pos: pos(1, 5)},
		Value: "flag",
		Type:  boolType,
	}

	totalIdent := &ast.Identifier{
		Token: lexer.Token{Type: lexer.IDENT, Literal: "total", Pos: pos(2, 5)},
		Value: "total",
		Type:  intType,
	}

	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(1, 1)},
				Names: []*ast.Identifier{flagIdent},
				Type:  boolType,
				Value: &ast.BooleanLiteral{
					Token: lexer.Token{Type: lexer.TRUE, Literal: "True", Pos: pos(1, 10)},
					Value: true,
				},
			},
			&ast.VarDeclStatement{
				Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(2, 1)},
				Names: []*ast.Identifier{totalIdent},
				Type:  intType,
				Value: &ast.IntegerLiteral{
					Token: lexer.Token{Type: lexer.INT, Literal: "0", Pos: pos(2, 10)},
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
						Token: lexer.Token{Type: lexer.IDENT, Literal: "total", Pos: pos(3, 8)},
						Value: "total",
						Type:  intType,
					},
					Value: &ast.IntegerLiteral{
						Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(3, 17)},
						Value: 1,
					},
				},
				Alternative: &ast.AssignmentStatement{
					Token:    lexer.Token{Type: lexer.IDENT, Literal: "total", Pos: pos(4, 8)},
					Operator: lexer.ASSIGN,
					Target: &ast.Identifier{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "total", Pos: pos(4, 8)},
						Value: "total",
						Type:  intType,
					},
					Value: &ast.IntegerLiteral{
						Token: lexer.Token{Type: lexer.INT, Literal: "2", Pos: pos(4, 17)},
						Value: 2,
					},
				},
			},
			&ast.ReturnStatement{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(5, 1)},
				ReturnValue: &ast.Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "total", Pos: pos(5, 9)},
					Value: "total",
					Type:  intType,
				},
			},
		},
	}

	compiler := NewCompiler("test_if")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	expectInstructions(t, chunk, []expectedInstruction{
		{OpLoadTrue, 0, 0},
		{OpStoreGlobal, 0, 0},
		{OpLoadConst0, 0, 0},
		{OpStoreGlobal, 0, 1},
		{OpLoadGlobal, 0, 0},
		{OpJumpIfFalse, 0, 3},
		{OpLoadConst1, 0, 0},
		{OpStoreGlobal, 0, 1},
		{OpJump, 0, 2},
		{OpLoadConst, 0, 2},
		{OpStoreGlobal, 0, 1},
		{OpLoadGlobal, 0, 1},
		{OpReturn, 1, 0},
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

func TestCompiler_ConstantFolding(t *testing.T) {
	intType := &ast.TypeAnnotation{Name: "Integer"}

	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.ReturnStatement{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(1, 1)},
				ReturnValue: &ast.BinaryExpression{
					Token:    lexer.Token{Type: lexer.PLUS, Literal: "+", Pos: pos(1, 9)},
					Operator: "+",
					Left: &ast.IntegerLiteral{
						Token: lexer.Token{Type: lexer.INT, Literal: "2", Pos: pos(1, 8)},
						Value: 2,
						Type:  intType,
					},
					Right: &ast.IntegerLiteral{
						Token: lexer.Token{Type: lexer.INT, Literal: "3", Pos: pos(1, 12)},
						Value: 3,
						Type:  intType,
					},
					Type: intType,
				},
			},
		},
	}

	compiler := NewCompiler("fold_test")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	expectInstructions(t, chunk, []expectedInstruction{
		{OpLoadConst0, 0, 0},
		{OpReturn, 1, 0},
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
				Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(1, 1)},
				Names: []*ast.Identifier{
					{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "adder", Pos: pos(1, 5)},
						Value: "adder",
					},
				},
			},
			&ast.ReturnStatement{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(2, 1)},
				ReturnValue: &ast.CallExpression{
					Token:    lexer.Token{Type: lexer.LPAREN, Literal: "(", Pos: pos(2, 11)},
					Function: &ast.Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "adder", Pos: pos(2, 9)}, Value: "adder"},
					Arguments: []ast.Expression{
						&ast.IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(2, 15)}, Value: 1},
						&ast.IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "2", Pos: pos(2, 18)}, Value: 2},
					},
				},
			},
		},
	}

	compiler := NewCompiler("call_test")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	expectInstructions(t, chunk, []expectedInstruction{
		{OpLoadNil, 0, 0},
		{OpStoreGlobal, 0, 0},
		{OpLoadGlobal, 0, 0},
		{OpLoadConst0, 0, 0},
		{OpLoadConst1, 0, 0},
		{OpCallIndirect, 2, 0},
		{OpReturn, 1, 0},
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

func TestCompiler_LambdaCapturesLocal(t *testing.T) {
	intType := &ast.TypeAnnotation{Name: "Integer"}

	countIdent := &ast.Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "count", Pos: pos(2, 5)}, Value: "count", Type: intType}
	incrementIdent := &ast.Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "increment", Pos: pos(3, 5)}, Value: "increment"}

	innerLambda := &ast.LambdaExpression{
		Token:      lexer.Token{Type: lexer.LAMBDA, Literal: "lambda", Pos: pos(4, 10)},
		ReturnType: intType,
		Body: &ast.BlockStatement{
			Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"},
			Statements: []ast.Statement{
				&ast.AssignmentStatement{
					Token:    lexer.Token{Type: lexer.IDENT, Literal: "count", Pos: pos(5, 3)},
					Operator: lexer.ASSIGN,
					Target: &ast.Identifier{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "count", Pos: pos(5, 3)},
						Value: "count",
						Type:  intType,
					},
					Value: &ast.BinaryExpression{
						Token:    lexer.Token{Type: lexer.PLUS, Literal: "+", Pos: pos(5, 15)},
						Operator: "+",
						Left: &ast.Identifier{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "count", Pos: pos(5, 12)},
							Value: "count",
							Type:  intType,
						},
						Right: &ast.IntegerLiteral{
							Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(5, 17)},
							Value: 1,
							Type:  intType,
						},
						Type: intType,
					},
				},
				&ast.ReturnStatement{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(6, 3)},
					ReturnValue: &ast.Identifier{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "count", Pos: pos(6, 11)},
						Value: "count",
						Type:  intType,
					},
				},
			},
		},
	}

	outerLambda := &ast.LambdaExpression{
		Token: lexer.Token{Type: lexer.LAMBDA, Literal: "lambda", Pos: pos(1, 10)},
		Body: &ast.BlockStatement{
			Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"},
			Statements: []ast.Statement{
				&ast.VarDeclStatement{
					Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(2, 1)},
					Names: []*ast.Identifier{
						countIdent,
					},
					Type: intType,
					Value: &ast.IntegerLiteral{
						Token: lexer.Token{Type: lexer.INT, Literal: "0", Pos: pos(2, 15)},
						Value: 0,
						Type:  intType,
					},
				},
				&ast.VarDeclStatement{
					Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(3, 1)},
					Names: []*ast.Identifier{
						incrementIdent,
					},
					Value: innerLambda,
				},
				&ast.ReturnStatement{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(4, 1)},
					ReturnValue: &ast.Identifier{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "increment", Pos: pos(4, 9)},
						Value: "increment",
					},
				},
			},
		},
	}

	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(1, 1)},
				Names: []*ast.Identifier{
					{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "makeCounter", Pos: pos(1, 5)},
						Value: "makeCounter",
					},
				},
				Value: outerLambda,
			},
		},
	}

	compiler := NewCompiler("lambda_test")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	outerFn := firstFunctionValue(chunk.Constants)
	if outerFn == nil {
		t.Fatalf("expected outer lambda function constant")
	}
	if len(outerFn.UpvalueDefs) != 0 {
		t.Fatalf("outer lambda should not capture upvalues, got %d", len(outerFn.UpvalueDefs))
	}

	innerFn := firstFunctionValue(outerFn.Chunk.Constants)
	if innerFn == nil {
		t.Fatalf("expected inner lambda function constant")
	}
	if len(innerFn.UpvalueDefs) != 1 {
		t.Fatalf("inner lambda should capture one upvalue, got %d", len(innerFn.UpvalueDefs))
	}
	if !innerFn.UpvalueDefs[0].IsLocal {
		t.Fatalf("inner lambda upvalue should capture parent local slot")
	}

	hasLoadUpvalue := false
	hasStoreUpvalue := false
	for _, inst := range innerFn.Chunk.Code {
		switch inst.OpCode() {
		case OpLoadUpvalue:
			hasLoadUpvalue = true
		case OpStoreUpvalue:
			hasStoreUpvalue = true
		}
	}

	if !hasLoadUpvalue {
		t.Fatalf("inner lambda bytecode missing LOAD_UPVALUE instruction")
	}
	if !hasStoreUpvalue {
		t.Fatalf("inner lambda bytecode missing STORE_UPVALUE instruction")
	}
}

func TestCompiler_BreakAndContinue(t *testing.T) {
	boolType := &ast.TypeAnnotation{Name: "Boolean"}

	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.WhileStatement{
				Token:     lexer.Token{Type: lexer.WHILE, Literal: "while", Pos: pos(1, 1)},
				Condition: &ast.BooleanLiteral{Token: lexer.Token{Type: lexer.TRUE, Literal: "True", Pos: pos(1, 7)}, Value: true, Type: boolType},
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
		Condition: &ast.BooleanLiteral{Token: lexer.Token{Type: lexer.TRUE, Literal: "True", Pos: pos(3, 1)}, Value: true, Type: boolType},
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

func TestCompiler_MemberAccess(t *testing.T) {
	objIdent := &ast.Identifier{
		Token: lexer.Token{Type: lexer.IDENT, Literal: "obj", Pos: pos(1, 5)},
		Value: "obj",
	}
	memberName := &ast.Identifier{
		Token: lexer.Token{Type: lexer.IDENT, Literal: "value", Pos: pos(2, 8)},
		Value: "value",
	}

	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(1, 1)},
				Names: []*ast.Identifier{objIdent},
			},
			&ast.AssignmentStatement{
				Token:    lexer.Token{Type: lexer.IDENT, Literal: "obj", Pos: pos(2, 1)},
				Operator: lexer.ASSIGN,
				Target: &ast.MemberAccessExpression{
					Token:  lexer.Token{Type: lexer.DOT, Literal: ".", Pos: pos(2, 6)},
					Object: &ast.Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "obj", Pos: pos(2, 1)}, Value: "obj"},
					Member: memberName,
				},
				Value: &ast.IntegerLiteral{
					Token: lexer.Token{Type: lexer.INT, Literal: "42", Pos: pos(2, 15)},
					Value: 42,
				},
			},
			&ast.ReturnStatement{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(3, 1)},
				ReturnValue: &ast.MemberAccessExpression{
					Token:  lexer.Token{Type: lexer.DOT, Literal: ".", Pos: pos(3, 6)},
					Object: &ast.Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "obj", Pos: pos(3, 1)}, Value: "obj"},
					Member: memberName,
				},
			},
		},
	}

	compiler := NewCompiler("member_test")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	expectInstructions(t, chunk, []expectedInstruction{
		{OpLoadNil, 0, 0},
		{OpStoreGlobal, 0, 0},
		{OpLoadGlobal, 0, 0},
		{OpLoadConst0, 0, 0},
		{OpSetProperty, 0, 1},
		{OpLoadGlobal, 0, 0},
		{OpGetProperty, 0, 1},
		{OpReturn, 1, 0},
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

	objIdent := &ast.Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "obj", Pos: pos(1, 5)}, Value: "obj", Type: intType}
	methodName := &ast.Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "DoIt", Pos: pos(2, 10)}, Value: "DoIt"}

	methodCall := &ast.MethodCallExpression{
		Token: lexer.Token{Type: lexer.DOT, Literal: ".", Pos: pos(2, 8)},
		Object: &ast.Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "obj", Pos: pos(2, 5)},
			Value: "obj",
			Type:  intType,
		},
		Method: methodName,
		Arguments: []ast.Expression{
			&ast.IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(2, 15)}, Value: 1, Type: intType},
		},
	}

	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(1, 1)},
				Names: []*ast.Identifier{objIdent},
				Type:  intType,
				Value: &ast.IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "0", Pos: pos(1, 10)}, Value: 0, Type: intType},
			},
			&ast.ExpressionStatement{
				Token:      lexer.Token{Type: lexer.IDENT, Literal: "DoIt", Pos: pos(2, 5)},
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
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Self", Pos: pos(1, 1)},
				Expression: &ast.Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Self", Pos: pos(1, 1)},
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

func TestCompiler_FunctionDeclDirectCall(t *testing.T) {
	intType := &ast.TypeAnnotation{Name: "Integer"}

	paramX := &ast.Parameter{
		Name: &ast.Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(1, 20)},
			Value: "x",
			Type:  intType,
		},
		Type: intType,
	}

	addOneBody := &ast.BlockStatement{
		Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"},
		Statements: []ast.Statement{
			&ast.ReturnStatement{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(2, 5)},
				ReturnValue: &ast.BinaryExpression{
					Token:    lexer.Token{Type: lexer.PLUS, Literal: "+", Pos: pos(2, 15)},
					Operator: "+",
					Left: &ast.Identifier{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(2, 13)},
						Value: "x",
						Type:  intType,
					},
					Right: &ast.IntegerLiteral{
						Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(2, 17)},
						Value: 1,
						Type:  intType,
					},
					Type: intType,
				},
			},
		},
	}

	functionDecl := &ast.FunctionDecl{
		Name:       &ast.Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "AddOne", Pos: pos(1, 1)}, Value: "AddOne"},
		ReturnType: intType,
		Parameters: []*ast.Parameter{paramX},
		Body:       addOneBody,
		Token:      lexer.Token{Type: lexer.FUNCTION, Literal: "function", Pos: pos(1, 1)},
	}

	callAddOne := &ast.CallExpression{
		Token: lexer.Token{Type: lexer.LPAREN, Literal: "(", Pos: pos(4, 12)},
		Function: &ast.Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "AddOne", Pos: pos(4, 5)},
			Value: "AddOne",
		},
		Arguments: []ast.Expression{
			&ast.IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "41", Pos: pos(4, 13)}, Value: 41, Type: intType},
		},
	}

	program := &ast.Program{
		Statements: []ast.Statement{
			functionDecl,
			&ast.ReturnStatement{
				Token:       lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(4, 1)},
				ReturnValue: callAddOne,
			},
		},
	}

	chunk := compileProgram(t, program)

	hasCall := false
	hasCallIndirect := false
	for _, inst := range chunk.Code {
		switch inst.OpCode() {
		case OpCall:
			hasCall = true
		case OpCallIndirect:
			hasCallIndirect = true
		}
	}

	if !hasCall {
		t.Fatalf("expected OpCall instruction for direct call")
	}
	if hasCallIndirect {
		t.Fatalf("did not expect OpCallIndirect for direct call to named function")
	}
}

func TestCompiler_RepeatLoop(t *testing.T) {
	intType := &ast.TypeAnnotation{Name: "Integer"}
	boolType := &ast.TypeAnnotation{Name: "Boolean"}

	xIdent := &ast.Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(1, 5)}, Value: "x", Type: intType}

	repeatStmt := &ast.RepeatStatement{
		Token: lexer.Token{Type: lexer.REPEAT, Literal: "repeat", Pos: pos(2, 1)},
		Body: &ast.BlockStatement{
			Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin", Pos: pos(2, 8)},
			Statements: []ast.Statement{
				&ast.AssignmentStatement{
					Token:    lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(3, 3)},
					Operator: lexer.ASSIGN,
					Target:   &ast.Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(3, 3)}, Value: "x", Type: intType},
					Value: &ast.BinaryExpression{
						Token:    lexer.Token{Type: lexer.PLUS, Literal: "+", Pos: pos(3, 8)},
						Operator: "+",
						Left:     &ast.Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(3, 7)}, Value: "x", Type: intType},
						Right: &ast.IntegerLiteral{
							Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(3, 11)},
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
			Left:     &ast.Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(4, 9)}, Value: "x", Type: intType},
			Right: &ast.IntegerLiteral{
				Token: lexer.Token{Type: lexer.INT, Literal: "3", Pos: pos(4, 13)},
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
				Value: &ast.IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "0", Pos: pos(1, 10)}, Value: 0},
			},
			repeatStmt,
			&ast.ReturnStatement{
				Token:       lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(5, 1)},
				ReturnValue: &ast.Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(5, 9)}, Value: "x", Type: intType},
			},
		},
	}

	compiler := NewCompiler("repeat_test")
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

func TestCompiler_ExecuteMatchesInterpreter(t *testing.T) {
	intType := &ast.TypeAnnotation{Name: "Integer"}

	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(1, 1)},
				Names: []*ast.Identifier{{Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(1, 5)}, Value: "x", Type: intType}},
				Type:  intType,
				Value: &ast.IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "2", Pos: pos(1, 10)}, Value: 2},
			},
			&ast.VarDeclStatement{
				Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(2, 1)},
				Names: []*ast.Identifier{{Token: lexer.Token{Type: lexer.IDENT, Literal: "y", Pos: pos(2, 5)}, Value: "y", Type: intType}},
				Type:  intType,
				Value: &ast.IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "3", Pos: pos(2, 10)}, Value: 3},
			},
			&ast.ReturnStatement{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(3, 1)},
				ReturnValue: &ast.BinaryExpression{
					Token:    lexer.Token{Type: lexer.PLUS, Literal: "+", Pos: pos(3, 14)},
					Operator: "+",
					Left: &ast.BinaryExpression{
						Token:    lexer.Token{Type: lexer.ASTERISK, Literal: "*", Pos: pos(3, 9)},
						Operator: "*",
						Left:     &ast.Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(3, 8)}, Value: "x", Type: intType},
						Right:    &ast.Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "y", Pos: pos(3, 12)}, Value: "y", Type: intType},
						Type:     intType,
					},
					Right: &ast.IntegerLiteral{
						Token: lexer.Token{Type: lexer.INT, Literal: "4", Pos: pos(3, 18)},
						Value: 4,
					},
					Type: intType,
				},
			},
		},
	}

	compiler := NewCompiler("exec_test")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	v := executeChunk(t, chunk)

	interpRunner := interp.New(io.Discard)
	interpVal := interpRunner.Eval(program)
	intResult, ok := interpVal.(*interp.IntegerValue)
	if !ok {
		t.Fatalf("interpreter returned %T, want *interp.IntegerValue", interpVal)
	}

	expected := IntValue(intResult.Value)
	if !valueEqual(v, expected) {
		t.Fatalf("bytecode value = %v, want %v", v, expected)
	}
}

type expectedInstruction struct {
	op OpCode
	a  byte
	b  uint16
}

func firstFunctionValue(constants []Value) *FunctionObject {
	for _, val := range constants {
		if val.IsFunction() {
			return val.AsFunction()
		}
	}
	return nil
}

func expectInstructions(t *testing.T, chunk *Chunk, expected []expectedInstruction) {
	t.Helper()

	if len(chunk.Code) != len(expected) {
		t.Fatalf("instruction count = %d, want %d", len(chunk.Code), len(expected))
	}

	for i, exp := range expected {
		inst := chunk.Code[i]
		if inst.OpCode() != exp.op {
			t.Fatalf("instruction %d opcode = %v, want %v", i, inst.OpCode(), exp.op)
		}
		if inst.A() != exp.a {
			t.Fatalf("instruction %d operand A = %d, want %d", i, inst.A(), exp.a)
		}
		if inst.B() != exp.b {
			t.Fatalf("instruction %d operand B = %d, want %d", i, inst.B(), exp.b)
		}
	}
}

func pos(line, column int) lexer.Position {
	return lexer.Position{Line: line, Column: column}
}

func compileProgram(t *testing.T, program *ast.Program) *Chunk {
	t.Helper()
	compiler := NewCompiler("test_program")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	return chunk
}

func executeChunk(t testing.TB, chunk *Chunk) Value {
	t.Helper()
	stack := make([]Value, 0, 16)
	globals := make(map[int]Value)
	push := func(v Value) {
		stack = append(stack, v)
	}
	pop := func() Value {
		if len(stack) == 0 {
			t.Fatalf("stack underflow during execution")
		}
		v := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		return v
	}

	localCount := chunk.LocalCount
	if localCount < 0 {
		t.Fatalf("invalid local count %d", localCount)
	}
	locals := make([]Value, localCount)
	ip := 0

	for ip < len(chunk.Code) {
		inst := chunk.Code[ip]
		ip++

		switch inst.OpCode() {
		case OpLoadConst:
			idx := int(inst.B())
			if idx >= len(chunk.Constants) {
				t.Fatalf("constant index %d out of range", idx)
			}
			push(chunk.Constants[idx])
		case OpLoadConst0:
			if len(chunk.Constants) == 0 {
				t.Fatalf("LOAD_CONST_0 with empty constant pool")
			}
			push(chunk.Constants[0])
		case OpLoadConst1:
			if len(chunk.Constants) < 2 {
				t.Fatalf("LOAD_CONST_1 requires at least two constants")
			}
			push(chunk.Constants[1])
		case OpLoadNil:
			push(NilValue())
		case OpLoadTrue:
			push(BoolValue(true))
		case OpLoadFalse:
			push(BoolValue(false))
		case OpStoreLocal:
			idx := int(inst.B())
			if idx >= len(locals) {
				t.Fatalf("local index %d out of range", idx)
			}
			locals[idx] = pop()
		case OpLoadGlobal:
			idx := int(inst.B())
			if val, ok := globals[idx]; ok {
				push(val)
			} else {
				push(NilValue())
			}
		case OpStoreGlobal:
			idx := int(inst.B())
			globals[idx] = pop()
		case OpLoadLocal:
			idx := int(inst.B())
			if idx >= len(locals) {
				t.Fatalf("local index %d out of range", idx)
			}
			push(locals[idx])
		case OpAddInt:
			b := pop()
			a := pop()
			push(IntValue(a.AsInt() + b.AsInt()))
		case OpSubInt:
			b := pop()
			a := pop()
			push(IntValue(a.AsInt() - b.AsInt()))
		case OpMulInt:
			b := pop()
			a := pop()
			push(IntValue(a.AsInt() * b.AsInt()))
		case OpDivInt:
			b := pop()
			a := pop()
			push(IntValue(a.AsInt() / b.AsInt()))
		case OpModInt:
			b := pop()
			a := pop()
			push(IntValue(a.AsInt() % b.AsInt()))
		case OpAddFloat:
			b := pop()
			a := pop()
			push(FloatValue(a.AsFloat() + b.AsFloat()))
		case OpSubFloat:
			b := pop()
			a := pop()
			push(FloatValue(a.AsFloat() - b.AsFloat()))
		case OpMulFloat:
			b := pop()
			a := pop()
			push(FloatValue(a.AsFloat() * b.AsFloat()))
		case OpDivFloat:
			b := pop()
			a := pop()
			push(FloatValue(a.AsFloat() / b.AsFloat()))
		case OpNegateInt:
			push(IntValue(-pop().AsInt()))
		case OpNegateFloat:
			push(FloatValue(-pop().AsFloat()))
		case OpStringConcat:
			b := pop()
			a := pop()
			push(StringValue(a.AsString() + b.AsString()))
		case OpEqual:
			b := pop()
			a := pop()
			eq, ok := valuesEqualForFold(a, b)
			if !ok {
				eq = valueEqual(a, b)
			}
			push(BoolValue(eq))
		case OpNotEqual:
			b := pop()
			a := pop()
			eq, ok := valuesEqualForFold(a, b)
			if !ok {
				eq = valueEqual(a, b)
			}
			push(BoolValue(!eq))
		case OpGreater:
			b := pop()
			a := pop()
			if a.Type == ValueFloat || b.Type == ValueFloat {
				push(BoolValue(a.AsFloat() > b.AsFloat()))
			} else {
				push(BoolValue(a.AsInt() > b.AsInt()))
			}
		case OpGreaterEqual:
			b := pop()
			a := pop()
			if a.Type == ValueFloat || b.Type == ValueFloat {
				push(BoolValue(a.AsFloat() >= b.AsFloat()))
			} else {
				push(BoolValue(a.AsInt() >= b.AsInt()))
			}
		case OpLess:
			b := pop()
			a := pop()
			if a.Type == ValueFloat || b.Type == ValueFloat {
				push(BoolValue(a.AsFloat() < b.AsFloat()))
			} else {
				push(BoolValue(a.AsInt() < b.AsInt()))
			}
		case OpLessEqual:
			b := pop()
			a := pop()
			if a.Type == ValueFloat || b.Type == ValueFloat {
				push(BoolValue(a.AsFloat() <= b.AsFloat()))
			} else {
				push(BoolValue(a.AsInt() <= b.AsInt()))
			}
		case OpAnd:
			b := pop()
			a := pop()
			push(BoolValue(a.AsBool() && b.AsBool()))
		case OpOr:
			b := pop()
			a := pop()
			push(BoolValue(a.AsBool() || b.AsBool()))
		case OpNot:
			push(BoolValue(!pop().AsBool()))
		case OpPop:
			_ = pop()
		case OpJump:
			ip += int(inst.SignedB())
		case OpJumpIfFalse:
			cond := pop()
			if !cond.AsBool() {
				ip += int(inst.SignedB())
			}
		case OpJumpIfTrue:
			cond := pop()
			if cond.AsBool() {
				ip += int(inst.SignedB())
			}
		case OpLoop:
			ip += int(inst.SignedB())
		case OpReturn:
			if inst.A() != 0 {
				return pop()
			}
			return NilValue()
		case OpHalt:
			if len(stack) > 0 {
				return stack[len(stack)-1]
			}
			return NilValue()
		case OpCallIndirect:
			t.Fatalf("CALL_INDIRECT not supported in test VM")
		default:
			t.Fatalf("unsupported opcode %v", inst.OpCode())
		}
	}

	return NilValue()
}

func valueEqual(a, b Value) bool {
	if a.Type != b.Type {
		if a.IsNumber() && b.IsNumber() {
			return a.AsFloat() == b.AsFloat()
		}
		return false
	}

	switch a.Type {
	case ValueNil:
		return true
	case ValueBool:
		return a.AsBool() == b.AsBool()
	case ValueInt:
		return a.AsInt() == b.AsInt()
	case ValueFloat:
		return a.AsFloat() == b.AsFloat()
	case ValueString:
		return a.AsString() == b.AsString()
	case ValueFunction:
		return a.AsFunction() == b.AsFunction()
	case ValueObject:
		return a.AsObject() == b.AsObject()
	default:
		return false
	}
}
