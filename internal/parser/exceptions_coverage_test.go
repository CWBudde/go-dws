package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// ============================================================================
// Additional Exception Parser Tests for Coverage
// ============================================================================

// TestParseBareExceptClause tests parsing bare except clause without specific handlers
// This tests lines 167-202 in parseExceptClause
func TestParseBareExceptClause(t *testing.T) {
	t.Run("Bare except with statements", func(t *testing.T) {
		input := `
		try
			raise Exception.Create('test');
		except
			PrintLn('Caught any exception');
			PrintLn('Handling it');
		end;
		`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		tryStmt, ok := program.Statements[0].(*ast.TryStatement)
		if !ok {
			t.Fatalf("statement is not *ast.TryStatement, got %T", program.Statements[0])
		}

		if tryStmt.ExceptClause == nil {
			t.Fatal("tryStmt.ExceptClause should not be nil")
		}

		// Bare except creates one synthetic handler
		if len(tryStmt.ExceptClause.Handlers) != 1 {
			t.Fatalf("except clause should have 1 synthetic handler, got %d", len(tryStmt.ExceptClause.Handlers))
		}

		handler := tryStmt.ExceptClause.Handlers[0]
		if handler.Variable != nil {
			t.Error("bare except handler should have nil Variable")
		}
		if handler.ExceptionType != nil {
			t.Error("bare except handler should have nil ExceptionType")
		}

		// Check the handler statement is a block with 2 statements
		block, ok := handler.Statement.(*ast.BlockStatement)
		if !ok {
			t.Fatalf("handler.Statement should be *ast.BlockStatement, got %T", handler.Statement)
		}

		if len(block.Statements) != 2 {
			t.Errorf("bare except block should have 2 statements, got %d", len(block.Statements))
		}
	})

	t.Run("Bare except with semicolons", func(t *testing.T) {
		input := `
		try
			raise Exception.Create('test');
		except
			;
			PrintLn('Statement after semicolon');
			;
		end;
		`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		tryStmt := program.Statements[0].(*ast.TryStatement)
		handler := tryStmt.ExceptClause.Handlers[0]
		block := handler.Statement.(*ast.BlockStatement)

		// Semicolons should be skipped, leaving 1 statement
		if len(block.Statements) != 1 {
			t.Errorf("bare except block should have 1 statement after skipping semicolons, got %d", len(block.Statements))
		}
	})

	t.Run("Empty bare except", func(t *testing.T) {
		input := `
		try
			raise Exception.Create('test');
		except
		end;
		`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		tryStmt := program.Statements[0].(*ast.TryStatement)

		// Empty bare except should not create a handler
		if len(tryStmt.ExceptClause.Handlers) != 0 {
			t.Errorf("empty bare except should have 0 handlers, got %d", len(tryStmt.ExceptClause.Handlers))
		}
	})
}

// TestParseElseClause tests parsing else clause in except blocks
// This tests lines 204-230 in parseExceptClause
func TestParseElseClause(t *testing.T) {
	t.Run("Except with else clause", func(t *testing.T) {
		input := `
		try
			raise Exception.Create('test');
		except
			on E: ESpecific do
				PrintLn('Specific exception');
		else
			PrintLn('Other exception');
			PrintLn('Handling in else');
		end;
		`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		tryStmt, ok := program.Statements[0].(*ast.TryStatement)
		if !ok {
			t.Fatalf("statement is not *ast.TryStatement, got %T", program.Statements[0])
		}

		if tryStmt.ExceptClause == nil {
			t.Fatal("tryStmt.ExceptClause should not be nil")
		}

		if tryStmt.ExceptClause.ElseBlock == nil {
			t.Fatal("exceptClause.ElseBlock should not be nil")
		}

		if len(tryStmt.ExceptClause.ElseBlock.Statements) != 2 {
			t.Errorf("else block should have 2 statements, got %d", len(tryStmt.ExceptClause.ElseBlock.Statements))
		}
	})

	t.Run("Else clause with semicolons", func(t *testing.T) {
		input := `
		try
			raise Exception.Create('test');
		except
			on E: Exception do
				PrintLn('Handled');
		else
			;
			PrintLn('Else statement');
			;
		end;
		`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		tryStmt := program.Statements[0].(*ast.TryStatement)
		elseBlock := tryStmt.ExceptClause.ElseBlock

		// Semicolons should be skipped
		if len(elseBlock.Statements) != 1 {
			t.Errorf("else block should have 1 statement after skipping semicolons, got %d", len(elseBlock.Statements))
		}
	})

	t.Run("Else clause before finally", func(t *testing.T) {
		input := `
		try
			raise Exception.Create('test');
		except
			on E: Exception do
				PrintLn('Handled');
		else
			PrintLn('Else');
		finally
			PrintLn('Finally');
		end;
		`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		tryStmt := program.Statements[0].(*ast.TryStatement)

		if tryStmt.ExceptClause.ElseBlock == nil {
			t.Fatal("exceptClause.ElseBlock should not be nil")
		}

		if tryStmt.FinallyClause == nil {
			t.Fatal("tryStmt.FinallyClause should not be nil")
		}
	})
}

// TestParseExceptionHandlerErrorPaths tests error handling in parseExceptionHandler
// This improves coverage of error paths in parseExceptionHandler
func TestParseExceptionHandlerErrorPaths(t *testing.T) {
	t.Run("Missing identifier after on", func(t *testing.T) {
		input := `
		try
			raise Exception.Create('test');
		except
			on : Exception do
				PrintLn('error');
		end;
		`

		l := lexer.New(input)
		p := New(l)
		p.ParseProgram()

		if len(p.Errors()) == 0 {
			t.Error("expected parser error for missing identifier after 'on'")
		}
	})

	t.Run("Missing colon after variable", func(t *testing.T) {
		input := `
		try
			raise Exception.Create('test');
		except
			on E Exception do
				PrintLn('error');
		end;
		`

		l := lexer.New(input)
		p := New(l)
		p.ParseProgram()

		if len(p.Errors()) == 0 {
			t.Error("expected parser error for missing ':'")
		}
	})

	t.Run("Missing type after colon", func(t *testing.T) {
		input := `
		try
			raise Exception.Create('test');
		except
			on E: do
				PrintLn('error');
		end;
		`

		l := lexer.New(input)
		p := New(l)
		p.ParseProgram()

		if len(p.Errors()) == 0 {
			t.Error("expected parser error for missing exception type")
		}
	})

	t.Run("Missing do keyword", func(t *testing.T) {
		input := `
		try
			raise Exception.Create('test');
		except
			on E: Exception
				PrintLn('error');
		end;
		`

		l := lexer.New(input)
		p := New(l)
		p.ParseProgram()

		if len(p.Errors()) == 0 {
			t.Error("expected parser error for missing 'do' keyword")
		}
	})

	t.Run("Missing statement after do", func(t *testing.T) {
		input := `
		try
			raise Exception.Create('test');
		except
			on E: Exception do
		end;
		`

		l := lexer.New(input)
		p := New(l)
		p.ParseProgram()

		if len(p.Errors()) == 0 {
			t.Error("expected parser error for missing statement after 'do'")
		}
	})
}

// TestParseMultipleExceptionHandlers tests multiple handlers with semicolons
// This covers the loop and semicolon handling in parseExceptClause
func TestParseMultipleExceptionHandlers(t *testing.T) {
	t.Run("Multiple handlers with semicolons between", func(t *testing.T) {
		input := `
		try
			raise Exception.Create('test');
		except
			on E1: EType1 do
				PrintLn('Handler 1');
			;
			on E2: EType2 do
				PrintLn('Handler 2');
			;
			on E3: EType3 do
				PrintLn('Handler 3');
		end;
		`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		tryStmt := program.Statements[0].(*ast.TryStatement)

		if len(tryStmt.ExceptClause.Handlers) != 3 {
			t.Errorf("expected 3 handlers, got %d", len(tryStmt.ExceptClause.Handlers))
		}

		// Verify each handler
		for i, handler := range tryStmt.ExceptClause.Handlers {
			if handler.Variable == nil {
				t.Errorf("handler %d should have a variable", i)
			}
			if handler.ExceptionType == nil {
				t.Errorf("handler %d should have an exception type", i)
			}
			if handler.Statement == nil {
				t.Errorf("handler %d should have a statement", i)
			}
		}
	})

	t.Run("Multiple semicolons between handlers", func(t *testing.T) {
		input := `
		try
			raise Exception.Create('test');
		except
			on E1: EType1 do
				PrintLn('Handler 1');
			;;;
			on E2: EType2 do
				PrintLn('Handler 2');
		end;
		`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		tryStmt := program.Statements[0].(*ast.TryStatement)

		if len(tryStmt.ExceptClause.Handlers) != 2 {
			t.Errorf("expected 2 handlers, got %d", len(tryStmt.ExceptClause.Handlers))
		}
	})
}

// TestParseExceptionHandlerWithBlockStatement tests handler with begin...end block
// This covers line 296-298 in parseExceptionHandler
func TestParseExceptionHandlerWithBlockStatement(t *testing.T) {
	t.Run("Handler with begin...end block", func(t *testing.T) {
		input := `
		try
			raise Exception.Create('test');
		except
			on E: Exception do begin
				PrintLn('Line 1');
				PrintLn('Line 2');
			end;
		end;
		`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		tryStmt := program.Statements[0].(*ast.TryStatement)
		handler := tryStmt.ExceptClause.Handlers[0]

		block, ok := handler.Statement.(*ast.BlockStatement)
		if !ok {
			t.Fatalf("handler.Statement should be *ast.BlockStatement, got %T", handler.Statement)
		}

		if len(block.Statements) != 2 {
			t.Errorf("block should have 2 statements, got %d", len(block.Statements))
		}
	})

	t.Run("Handler with single statement (not block)", func(t *testing.T) {
		input := `
		try
			raise Exception.Create('test');
		except
			on E: Exception do
				PrintLn('Single statement');
		end;
		`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		tryStmt := program.Statements[0].(*ast.TryStatement)
		handler := tryStmt.ExceptClause.Handlers[0]

		// Single statement should not be a BlockStatement
		_, isBlock := handler.Statement.(*ast.BlockStatement)
		if isBlock {
			t.Error("single statement should not be wrapped in BlockStatement")
		}
	})
}
