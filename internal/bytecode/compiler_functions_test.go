package bytecode

import (
	"io"
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/interp"
	"github.com/cwbudde/go-dws/internal/lexer"
)

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

	compiler := newTestCompiler("lambda_test")
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

	compiler := newTestCompiler("exec_test")
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
