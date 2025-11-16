package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
)

func TestCompleteSimplePrograms(t *testing.T) {
	tests := []struct {
		assertions func(*testing.T, *ast.Program)
		name       string
		input      string
		stmtCount  int
	}{
		{
			name: "variable declaration and usage",
			input: `
var x: Integer := 5;
var y: Integer := 10;
x := x + y;
`,
			stmtCount: 3,
			assertions: func(t *testing.T, program *ast.Program) {
				// First statement: var x: Integer := 5;
				varDecl1, ok := program.Statements[0].(*ast.VarDeclStatement)
				if !ok {
					t.Fatalf("statement 0 is not VarDeclStatement. got=%T", program.Statements[0])
				}
				if varDecl1.Names[0].Value != "x" {
					t.Errorf("varDecl1.Names[0].Value = %q, want 'x'", varDecl1.Names[0].Value)
				}
				if varDecl1.Type == nil || varDecl1.Type.String() != "Integer" {
					t.Errorf("varDecl1.Type = %q, want 'Integer'", varDecl1.Type)
				}
				if !testIntegerLiteral(t, varDecl1.Value, 5) {
					return
				}

				// Second statement: var y: Integer := 10;
				varDecl2, ok := program.Statements[1].(*ast.VarDeclStatement)
				if !ok {
					t.Fatalf("statement 1 is not VarDeclStatement. got=%T", program.Statements[1])
				}
				if varDecl2.Names[0].Value != "y" {
					t.Errorf("varDecl2.Names[0].Value = %q, want 'y'", varDecl2.Names[0].Value)
				}
				if !testIntegerLiteral(t, varDecl2.Value, 10) {
					return
				}

				// Third statement: x := x + y;
				assign, ok := program.Statements[2].(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("statement 2 is not AssignmentStatement. got=%T", program.Statements[2])
				}
				assignTarget, ok := assign.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("assign.Target is not *ast.Identifier. got=%T", assign.Target)
				}
				if assignTarget.Value != "x" {
					t.Errorf("assign.Target.Value = %q, want 'x'", assignTarget.Value)
				}
				if !testInfixExpression(t, assign.Value, "x", "+", "y") {
					return
				}
			},
		},
		{
			name: "program with function call",
			input: `
var message: String := 'Hello, World!';
PrintLn(message);
`,
			stmtCount: 2,
			assertions: func(t *testing.T, program *ast.Program) {
				// First statement: var message: String := 'Hello, World!';
				varDecl, ok := program.Statements[0].(*ast.VarDeclStatement)
				if !ok {
					t.Fatalf("statement 0 is not VarDeclStatement. got=%T", program.Statements[0])
				}
				if varDecl.Names[0].Value != "message" {
					t.Errorf("varDecl.Names[0].Value = %q, want 'message'", varDecl.Names[0].Value)
				}
				if !testStringLiteralExpression(t, varDecl.Value, "Hello, World!") {
					return
				}

				// Second statement: PrintLn(message);
				exprStmt, ok := program.Statements[1].(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("statement 1 is not ExpressionStatement. got=%T", program.Statements[1])
				}
				call, ok := exprStmt.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("statement 1 expression is not CallExpression. got=%T", exprStmt.Expression)
				}
				if !testIdentifier(t, call.Function, "PrintLn") {
					return
				}
				if len(call.Arguments) != 1 {
					t.Fatalf("call has %d arguments, want 1", len(call.Arguments))
				}
				if !testIdentifier(t, call.Arguments[0], "message") {
					return
				}
			},
		},
		{
			name: "program with block statement",
			input: `
var x := 0;
begin
  x := x + 1;
  x := x * 2;
end;
`,
			stmtCount: 2,
			assertions: func(t *testing.T, program *ast.Program) {
				// First statement: var x := 0;
				varDecl, ok := program.Statements[0].(*ast.VarDeclStatement)
				if !ok {
					t.Fatalf("statement 0 is not VarDeclStatement. got=%T", program.Statements[0])
				}
				if varDecl.Names[0].Value != "x" {
					t.Errorf("varDecl.Names[0].Value = %q, want 'x'", varDecl.Names[0].Value)
				}

				// Second statement: begin...end block
				block, ok := program.Statements[1].(*ast.BlockStatement)
				if !ok {
					t.Fatalf("statement 1 is not BlockStatement. got=%T", program.Statements[1])
				}
				if len(block.Statements) != 2 {
					t.Fatalf("block has %d statements, want 2", len(block.Statements))
				}

				// First block statement: x := x + 1;
				assign1, ok := block.Statements[0].(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("block statement 0 is not AssignmentStatement. got=%T", block.Statements[0])
				}
				if !testInfixExpression(t, assign1.Value, "x", "+", 1) {
					return
				}

				// Second block statement: x := x * 2;
				assign2, ok := block.Statements[1].(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("block statement 1 is not AssignmentStatement. got=%T", block.Statements[1])
				}
				if !testInfixExpression(t, assign2.Value, "x", "*", 2) {
					return
				}
			},
		},
		{
			name: "mixed statements",
			input: `
var a := 1;
var b := 2;
var sum := a + b;
PrintLn(sum);
begin
  a := a + 1;
  b := b + 1;
end;
`,
			stmtCount: 5,
			assertions: func(t *testing.T, program *ast.Program) {
				// Verify first three are variable declarations
				for i := 0; i < 3; i++ {
					if _, ok := program.Statements[i].(*ast.VarDeclStatement); !ok {
						t.Fatalf("statement %d is not VarDeclStatement. got=%T", i, program.Statements[i])
					}
				}

				// Fourth statement: PrintLn(sum);
				exprStmt, ok := program.Statements[3].(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("statement 3 is not ExpressionStatement. got=%T", program.Statements[3])
				}
				if _, ok := exprStmt.Expression.(*ast.CallExpression); !ok {
					t.Fatalf("statement 3 expression is not CallExpression. got=%T", exprStmt.Expression)
				}

				// Fifth statement: begin...end block
				if _, ok := program.Statements[4].(*ast.BlockStatement); !ok {
					t.Fatalf("statement 4 is not BlockStatement. got=%T", program.Statements[4])
				}
			},
		},
		{
			name: "arithmetic expressions",
			input: `
var result := (10 + 5) * 2 - 3;
result := result / 3;
`,
			stmtCount: 2,
			assertions: func(t *testing.T, program *ast.Program) {
				// First statement: var result := (10 + 5) * 2 - 3;
				varDecl, ok := program.Statements[0].(*ast.VarDeclStatement)
				if !ok {
					t.Fatalf("statement 0 is not VarDeclStatement. got=%T", program.Statements[0])
				}
				if varDecl.Names[0].Value != "result" {
					t.Errorf("varDecl.Names[0].Value = %q, want 'result'", varDecl.Names[0].Value)
				}
				// The value should be a complex binary expression
				if varDecl.Value == nil {
					t.Fatal("varDecl.Value is nil")
				}

				// Second statement: result := result / 3;
				assign, ok := program.Statements[1].(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("statement 1 is not AssignmentStatement. got=%T", program.Statements[1])
				}
				if !testInfixExpression(t, assign.Value, "result", "/", 3) {
					return
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != tt.stmtCount {
				t.Fatalf("program has %d statements, want %d", len(program.Statements), tt.stmtCount)
			}

			tt.assertions(t, program)
		})
	}
}

// TestImplicitProgramBlock tests that programs without explicit begin/end work correctly.
func TestImplicitProgramBlock(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		stmtCount int
	}{
		{
			name:      "single variable declaration",
			input:     "var x := 5;",
			stmtCount: 1,
		},
		{
			name:      "single assignment",
			input:     "x := 10;",
			stmtCount: 1,
		},
		{
			name: "multiple statements without begin/end",
			input: `
var x := 1;
var y := 2;
x := x + y;
`,
			stmtCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != tt.stmtCount {
				t.Fatalf("program has %d statements, want %d", len(program.Statements), tt.stmtCount)
			}
		})
	}
}
