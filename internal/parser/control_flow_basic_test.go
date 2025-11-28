package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Loop Control Statement Parser Tests
// ============================================================================

// TestParseBreakStatement tests parsing break statements
func TestParseBreakStatement(t *testing.T) {
	t.Run("Simple break statement", func(t *testing.T) {
		input := `break;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		breakStmt, ok := program.Statements[0].(*ast.BreakStatement)
		if !ok {
			t.Fatalf("statement is not *ast.BreakStatement, got %T", program.Statements[0])
		}

		if breakStmt.TokenLiteral() != "break" {
			t.Errorf("breakStmt.TokenLiteral() = %q, want 'break'", breakStmt.TokenLiteral())
		}
	})

	t.Run("Break in for loop", func(t *testing.T) {
		input := `
		for i := 1 to 10 do begin
			if i > 5 then break;
		end;
		`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		forStmt, ok := program.Statements[0].(*ast.ForStatement)
		if !ok {
			t.Fatalf("statement is not *ast.ForStatement, got %T", program.Statements[0])
		}

		block, ok := forStmt.Body.(*ast.BlockStatement)
		if !ok {
			t.Fatalf("for body is not *ast.BlockStatement, got %T", forStmt.Body)
		}

		if len(block.Statements) != 1 {
			t.Fatalf("block should have 1 statement, got %d", len(block.Statements))
		}

		ifStmt, ok := block.Statements[0].(*ast.IfStatement)
		if !ok {
			t.Fatalf("block statement is not *ast.IfStatement, got %T", block.Statements[0])
		}

		breakStmt, ok := ifStmt.Consequence.(*ast.BreakStatement)
		if !ok {
			t.Fatalf("if consequence is not *ast.BreakStatement, got %T", ifStmt.Consequence)
		}

		if breakStmt.TokenLiteral() != "break" {
			t.Errorf("breakStmt.TokenLiteral() = %q, want 'break'", breakStmt.TokenLiteral())
		}
	})

	t.Run("Break in while loop", func(t *testing.T) {
		input := `
		while true do begin
			break;
		end;
		`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		whileStmt := program.Statements[0].(*ast.WhileStatement)
		block := whileStmt.Body.(*ast.BlockStatement)

		if len(block.Statements) != 1 {
			t.Fatalf("while body should have 1 statement, got %d", len(block.Statements))
		}

		_, ok := block.Statements[0].(*ast.BreakStatement)
		if !ok {
			t.Fatalf("while body statement is not *ast.BreakStatement, got %T", block.Statements[0])
		}
	})

	t.Run("Break in repeat loop", func(t *testing.T) {
		input := `
		repeat
			break;
		until false;
		`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		repeatStmt := program.Statements[0].(*ast.RepeatStatement)

		_, ok := repeatStmt.Body.(*ast.BreakStatement)
		if !ok {
			t.Fatalf("repeat body is not *ast.BreakStatement, got %T", repeatStmt.Body)
		}
	})

	t.Run("Break missing semicolon", func(t *testing.T) {
		input := `break`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		if _, ok := program.Statements[0].(*ast.BreakStatement); !ok {
			t.Fatalf("statement is not *ast.BreakStatement, got %T", program.Statements[0])
		}
	})
}

// TestParseContinueStatement tests parsing continue statements
func TestParseContinueStatement(t *testing.T) {
	t.Run("Simple continue statement", func(t *testing.T) {
		input := `continue;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		continueStmt, ok := program.Statements[0].(*ast.ContinueStatement)
		if !ok {
			t.Fatalf("statement is not *ast.ContinueStatement, got %T", program.Statements[0])
		}

		if continueStmt.TokenLiteral() != "continue" {
			t.Errorf("continueStmt.TokenLiteral() = %q, want 'continue'", continueStmt.TokenLiteral())
		}
	})

	t.Run("Continue in for loop", func(t *testing.T) {
		input := `
		for i := 1 to 10 do begin
			if i mod 2 = 0 then continue;
			PrintLn(i);
		end;
		`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		forStmt := program.Statements[0].(*ast.ForStatement)
		block := forStmt.Body.(*ast.BlockStatement)

		if len(block.Statements) != 2 {
			t.Fatalf("for body should have 2 statements, got %d", len(block.Statements))
		}

		ifStmt := block.Statements[0].(*ast.IfStatement)
		continueStmt, ok := ifStmt.Consequence.(*ast.ContinueStatement)
		if !ok {
			t.Fatalf("if consequence is not *ast.ContinueStatement, got %T", ifStmt.Consequence)
		}

		if continueStmt.TokenLiteral() != "continue" {
			t.Errorf("continueStmt.TokenLiteral() = %q, want 'continue'", continueStmt.TokenLiteral())
		}
	})

	t.Run("Continue in while loop", func(t *testing.T) {
		input := `
		while i < 10 do begin
			i := i + 1;
			if i mod 2 = 0 then continue;
			PrintLn(i);
		end;
		`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		whileStmt := program.Statements[0].(*ast.WhileStatement)
		block := whileStmt.Body.(*ast.BlockStatement)

		if len(block.Statements) != 3 {
			t.Fatalf("while body should have 3 statements, got %d", len(block.Statements))
		}

		ifStmt := block.Statements[1].(*ast.IfStatement)
		_, ok := ifStmt.Consequence.(*ast.ContinueStatement)
		if !ok {
			t.Fatalf("if consequence is not *ast.ContinueStatement, got %T", ifStmt.Consequence)
		}
	})

	t.Run("Continue missing semicolon", func(t *testing.T) {
		input := `continue`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		if _, ok := program.Statements[0].(*ast.ContinueStatement); !ok {
			t.Fatalf("statement is not *ast.ContinueStatement, got %T", program.Statements[0])
		}
	})
}

// TestParseExitStatement tests parsing exit statements
func TestParseExitStatement(t *testing.T) {
	t.Run("Simple exit statement", func(t *testing.T) {
		input := `exit;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		exitStmt, ok := program.Statements[0].(*ast.ExitStatement)
		if !ok {
			t.Fatalf("statement is not *ast.ExitStatement, got %T", program.Statements[0])
		}

		if exitStmt.TokenLiteral() != "exit" {
			t.Errorf("exitStmt.TokenLiteral() = %q, want 'exit'", exitStmt.TokenLiteral())
		}

		if exitStmt.ReturnValue != nil {
			t.Error("exitStmt.ReturnValue should be nil for simple exit")
		}
	})

	t.Run("Exit with integer value", func(t *testing.T) {
		input := `exit(-1);`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		exitStmt := program.Statements[0].(*ast.ExitStatement)

		if exitStmt.ReturnValue == nil {
			t.Fatal("exitStmt.ReturnValue should not be nil")
		}

		unaryExpr, ok := exitStmt.ReturnValue.(*ast.UnaryExpression)
		if !ok {
			t.Fatalf("exitStmt.ReturnValue is not *ast.UnaryExpression, got %T", exitStmt.ReturnValue)
		}

		if unaryExpr.Operator != "-" {
			t.Errorf("unary operator = %q, want '-'", unaryExpr.Operator)
		}

		intLit, ok := unaryExpr.Right.(*ast.IntegerLiteral)
		if !ok {
			t.Fatalf("unary right is not *ast.IntegerLiteral, got %T", unaryExpr.Right)
		}

		if intLit.Value != 1 {
			t.Errorf("intLit.Value = %d, want 1", intLit.Value)
		}
	})

	t.Run("Exit with identifier", func(t *testing.T) {
		input := `exit(result);`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		exitStmt := program.Statements[0].(*ast.ExitStatement)

		if exitStmt.ReturnValue == nil {
			t.Fatal("exitStmt.ReturnValue should not be nil")
		}

		ident, ok := exitStmt.ReturnValue.(*ast.Identifier)
		if !ok {
			t.Fatalf("exitStmt.ReturnValue is not *ast.Identifier, got %T", exitStmt.ReturnValue)
		}

		if ident.Value != "result" {
			t.Errorf("ident.Value = %q, want 'result'", ident.Value)
		}
	})

	t.Run("Exit with expression", func(t *testing.T) {
		input := `exit(i + 1);`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		exitStmt := program.Statements[0].(*ast.ExitStatement)

		if exitStmt.ReturnValue == nil {
			t.Fatal("exitStmt.ReturnValue should not be nil")
		}

		binExpr, ok := exitStmt.ReturnValue.(*ast.BinaryExpression)
		if !ok {
			t.Fatalf("exitStmt.ReturnValue is not *ast.BinaryExpression, got %T", exitStmt.ReturnValue)
		}

		if binExpr.Operator != "+" {
			t.Errorf("binary operator = %q, want '+'", binExpr.Operator)
		}
	})

	t.Run("Exit in function", func(t *testing.T) {
		input := `
		function Test(i: Integer): Integer;
		begin
			if i <= 0 then exit(-1);
			Result := i * 2;
		end;
		`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		funcDecl := program.Statements[0].(*ast.FunctionDecl)
		block := funcDecl.Body

		if len(block.Statements) != 2 {
			t.Fatalf("function body should have 2 statements, got %d", len(block.Statements))
		}

		ifStmt := block.Statements[0].(*ast.IfStatement)
		exitStmt, ok := ifStmt.Consequence.(*ast.ExitStatement)
		if !ok {
			t.Fatalf("if consequence is not *ast.ExitStatement, got %T", ifStmt.Consequence)
		}

		if exitStmt.ReturnValue == nil {
			t.Error("exitStmt.ReturnValue should not be nil")
		}
	})

	t.Run("Exit missing semicolon", func(t *testing.T) {
		input := `exit`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		exitStmt, ok := program.Statements[0].(*ast.ExitStatement)
		if !ok {
			t.Fatalf("statement is not *ast.ExitStatement, got %T", program.Statements[0])
		}

		if exitStmt.ReturnValue != nil {
			t.Error("exitStmt.ReturnValue should be nil for bare exit")
		}
	})

	t.Run("Exit with value missing closing paren", func(t *testing.T) {
		input := `exit(42`

		l := lexer.New(input)
		p := New(l)
		p.ParseProgram()

		if len(p.errors) == 0 {
			t.Error("expected parser error for missing closing paren, got none")
		}
	})

	t.Run("Exit with empty parens", func(t *testing.T) {
		input := `exit();`

		l := lexer.New(input)
		p := New(l)
		p.ParseProgram()

		if len(p.errors) == 0 {
			t.Error("expected parser error for empty parens, got none")
		}
	})
}

// TestBreakContinueInNestedLoops tests break and continue in nested loop structures
func TestBreakContinueInNestedLoops(t *testing.T) {
	t.Run("Break in nested for loops", func(t *testing.T) {
		input := `
		for i := 1 to 10 do begin
			for j := 1 to 10 do begin
				if i * j = 12 then break;
			end;
		end;
		`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		outerFor := program.Statements[0].(*ast.ForStatement)
		outerBlock := outerFor.Body.(*ast.BlockStatement)
		innerFor := outerBlock.Statements[0].(*ast.ForStatement)
		innerBlock := innerFor.Body.(*ast.BlockStatement)
		ifStmt := innerBlock.Statements[0].(*ast.IfStatement)

		_, ok := ifStmt.Consequence.(*ast.BreakStatement)
		if !ok {
			t.Fatalf("inner if consequence is not *ast.BreakStatement, got %T", ifStmt.Consequence)
		}
	})

	t.Run("Continue in nested loops", func(t *testing.T) {
		input := `
		for i := 1 to 5 do begin
			for j := 1 to 5 do begin
				if j = 3 then continue;
			end;
		end;
		`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		outerFor := program.Statements[0].(*ast.ForStatement)
		outerBlock := outerFor.Body.(*ast.BlockStatement)
		innerFor := outerBlock.Statements[0].(*ast.ForStatement)
		innerBlock := innerFor.Body.(*ast.BlockStatement)
		ifStmt := innerBlock.Statements[0].(*ast.IfStatement)

		_, ok := ifStmt.Consequence.(*ast.ContinueStatement)
		if !ok {
			t.Fatalf("inner if consequence is not *ast.ContinueStatement, got %T", ifStmt.Consequence)
		}
	})
}

// TestExitInCaseStatement tests exit in case branches
func TestExitInCaseStatement(t *testing.T) {
	input := `
	case x of
		1: exit;
		2: exit(42);
	else
		exit(-1);
	end;
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	caseStmt := program.Statements[0].(*ast.CaseStatement)

	// Check first branch: exit;
	if len(caseStmt.Cases) < 2 {
		t.Fatalf("case should have at least 2 branches, got %d", len(caseStmt.Cases))
	}

	exitStmt1, ok := caseStmt.Cases[0].Statement.(*ast.ExitStatement)
	if !ok {
		t.Fatalf("first case statement is not *ast.ExitStatement, got %T", caseStmt.Cases[0].Statement)
	}
	if exitStmt1.ReturnValue != nil {
		t.Error("first exit should have nil value")
	}

	// Check second branch: exit(42);
	exitStmt2, ok := caseStmt.Cases[1].Statement.(*ast.ExitStatement)
	if !ok {
		t.Fatalf("second case statement is not *ast.ExitStatement, got %T", caseStmt.Cases[1].Statement)
	}
	if exitStmt2.ReturnValue == nil {
		t.Error("second exit should have non-nil value")
	}

	// Check else branch: exit(-1);
	if caseStmt.Else == nil {
		t.Fatal("case else branch should not be nil")
	}
	exitStmt3, ok := caseStmt.Else.(*ast.ExitStatement)
	if !ok {
		t.Fatalf("case else is not *ast.ExitStatement, got %T", caseStmt.Else)
	}
	if exitStmt3.ReturnValue == nil {
		t.Error("else exit should have non-nil value")
	}
}
