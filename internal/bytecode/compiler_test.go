package bytecode

import (
	"io"
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/interp"
	"github.com/cwbudde/go-dws/internal/lexer"
)

const (
	numBuiltinGlobals = 43 // ExceptObject + 21 string/IO builtins + 4 type casts + 2 char ops + 6 math + 9 extended math (Pi is a constant, not a global)
	userGlobal0       = numBuiltinGlobals
	userGlobal1       = numBuiltinGlobals + 1
)

var testCompileOptimizeOptions = []OptimizeOption{
	WithOptimizationPass(PassLiteralDiscard, false),
	WithOptimizationPass(PassStackShuffle, false),
	WithOptimizationPass(PassConstPropagation, false),
	WithOptimizationPass(PassDeadCode, false),
}

func newTestCompiler(name string) *Compiler {
	return NewCompiler(name, WithCompilerOptimizeOptions(testCompileOptimizeOptions...))
}

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

func TestCompiler_ArrayLiteralAndIndex(t *testing.T) {
	arrIdent := &ast.Identifier{
		Token: lexer.Token{Type: lexer.IDENT, Literal: "arr", Pos: pos(1, 1)},
		Value: "arr",
	}
	arrayLiteral := &ast.ArrayLiteralExpression{
		Token: lexer.Token{Type: lexer.LBRACK, Literal: "[", Pos: pos(1, 10)},
		Elements: []ast.Expression{
			&ast.IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(1, 11)}, Value: 1},
			&ast.IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "2", Pos: pos(1, 14)}, Value: 2},
		},
	}
	assignStmt := &ast.AssignmentStatement{
		Token:    lexer.Token{Type: lexer.IDENT, Literal: "arr", Pos: pos(2, 1)},
		Operator: lexer.ASSIGN,
		Target: &ast.IndexExpression{
			Token: lexer.Token{Type: lexer.LBRACK, Literal: "[", Pos: pos(2, 4)},
			Left: &ast.Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "arr", Pos: pos(2, 1)},
				Value: "arr",
			},
			Index: &ast.IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(2, 8)}, Value: 1},
		},
		Value: &ast.IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "5", Pos: pos(2, 13)}, Value: 5},
	}
	returnStmt := &ast.ReturnStatement{
		Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(3, 1)},
		ReturnValue: &ast.IndexExpression{
			Token: lexer.Token{Type: lexer.LBRACK, Literal: "[", Pos: pos(3, 10)},
			Left: &ast.Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "arr", Pos: pos(3, 8)},
				Value: "arr",
			},
			Index: &ast.IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "0", Pos: pos(3, 12)}, Value: 0},
		},
	}

	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(1, 1)},
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
						Token: lexer.Token{Type: lexer.IDENT, Literal: "TPoint", Pos: pos(1, 13)},
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

func TestCompiler_TryExceptTypedHandler(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.TryStatement{
				Token: lexer.Token{Type: lexer.TRY, Literal: "try", Pos: pos(1, 1)},
				TryBlock: &ast.BlockStatement{Statements: []ast.Statement{
					&ast.ExpressionStatement{Expression: &ast.NilLiteral{Token: lexer.Token{Type: lexer.NIL, Literal: "Nil", Pos: pos(1, 5)}}},
				}},
				ExceptClause: &ast.ExceptClause{
					Token: lexer.Token{Type: lexer.EXCEPT, Literal: "except", Pos: pos(2, 1)},
					Handlers: []*ast.ExceptionHandler{
						{
							Token:         lexer.Token{Type: lexer.ON, Literal: "on", Pos: pos(2, 3)},
							Variable:      &ast.Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "E", Pos: pos(2, 6)}, Value: "E"},
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
		Target:   &ast.Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(3, 3)}, Value: "x"},
		Value:    &ast.IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(3, 10)}, Value: 1},
	}

	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(1, 1)},
				Names: []*ast.Identifier{{Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(1, 5)}, Value: "x"}},
			},
			&ast.TryStatement{
				Token: lexer.Token{Type: lexer.TRY, Literal: "try", Pos: pos(2, 1)},
				TryBlock: &ast.BlockStatement{Statements: []ast.Statement{
					&ast.ExpressionStatement{Expression: &ast.NilLiteral{Token: lexer.Token{Type: lexer.NIL, Literal: "Nil", Pos: pos(2, 5)}}},
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
					Token: lexer.Token{Type: lexer.INT, Literal: "5", Pos: pos(1, 7)},
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
		Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(1, 5)},
		Value: "x",
	}

	assignTry := &ast.AssignmentStatement{
		Token:    lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(2, 3)},
		Operator: lexer.ASSIGN,
		Target: &ast.Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(2, 3)},
			Value: "x",
		},
		Value: &ast.IntegerLiteral{
			Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(2, 8)},
			Value: 1,
		},
	}

	assignFinally := &ast.AssignmentStatement{
		Token:    lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(3, 3)},
		Operator: lexer.ASSIGN,
		Target: &ast.Identifier{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(3, 3)},
			Value: "x",
		},
		Value: &ast.IntegerLiteral{
			Token: lexer.Token{Type: lexer.INT, Literal: "2", Pos: pos(3, 8)},
			Value: 2,
		},
	}

	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(1, 1)},
				Names: []*ast.Identifier{ident},
				Value: &ast.IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "0", Pos: pos(1, 10)}, Value: 0},
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

func TestCompiler_OptimizesLiteralExpressionStatements(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.ExpressionStatement{
				Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(1, 1)},
				Expression: &ast.IntegerLiteral{
					Token: lexer.Token{Type: lexer.INT, Literal: "1", Pos: pos(1, 1)},
					Value: 1,
				},
			},
			&ast.ExpressionStatement{
				Token: lexer.Token{Type: lexer.TRUE, Literal: "True", Pos: pos(2, 1)},
				Expression: &ast.BooleanLiteral{
					Token: lexer.Token{Type: lexer.TRUE, Literal: "True", Pos: pos(2, 1)},
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

type expectedInstruction struct {
	Op OpCode
	A  byte
	B  uint16
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
		if inst.OpCode() != exp.Op {
			t.Fatalf("instruction %d opcode = %v, want %v", i, inst.OpCode(), exp.Op)
		}
		if inst.A() != exp.A {
			t.Fatalf("instruction %d operand A = %d, want %d", i, inst.A(), exp.A)
		}
		if inst.B() != exp.B {
			t.Fatalf("instruction %d operand B = %d, want %d", i, inst.B(), exp.B)
		}
	}
}

func pos(line, column int) lexer.Position {
	return lexer.Position{Line: line, Column: column}
}

func compileProgram(t *testing.T, program *ast.Program) *Chunk {
	t.Helper()
	compiler := newTestCompiler("test_program")
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
