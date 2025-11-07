package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// ============================================================================
// Raise Statement Parser Tests
// ============================================================================

// Test parsing raise statement with exception expression
func TestParseRaiseStatement(t *testing.T) {
	t.Run("Raise with constructor call", func(t *testing.T) {
		input := `raise Exception.Create('error message');`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		raiseStmt, ok := program.Statements[0].(*ast.RaiseStatement)
		if !ok {
			t.Fatalf("statement is not *ast.RaiseStatement, got %T", program.Statements[0])
		}

		if raiseStmt.Exception == nil {
			t.Fatal("raiseStmt.Exception should not be nil")
		}

		// Exception.Create() is parsed as a NewExpression in DWScript
		// This is the standard way to create exceptions
		newExpr, ok := raiseStmt.Exception.(*ast.NewExpression)
		if !ok {
			t.Fatalf("raiseStmt.Exception is not *ast.NewExpression, got %T", raiseStmt.Exception)
		}

		if newExpr.ClassName.Value != "Exception" {
			t.Errorf("newExpr.ClassName.Value = %s, want 'Exception'", newExpr.ClassName.Value)
		}

		if len(newExpr.Arguments) != 1 {
			t.Errorf("newExpr.Arguments should have 1 argument, got %d", len(newExpr.Arguments))
		}
	})

	t.Run("Raise with function call", func(t *testing.T) {
		input := `raise CreateException('custom error');`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		raiseStmt, ok := program.Statements[0].(*ast.RaiseStatement)
		if !ok {
			t.Fatalf("statement is not *ast.RaiseStatement, got %T", program.Statements[0])
		}

		if raiseStmt.Exception == nil {
			t.Fatal("raiseStmt.Exception should not be nil")
		}

		// Check it's a call expression
		callExpr, ok := raiseStmt.Exception.(*ast.CallExpression)
		if !ok {
			t.Fatalf("raiseStmt.Exception is not *ast.CallExpression, got %T", raiseStmt.Exception)
		}

		// Check function name
		ident, ok := callExpr.Function.(*ast.Identifier)
		if !ok {
			t.Fatalf("callExpr.Function is not *ast.Identifier, got %T", callExpr.Function)
		}

		if ident.Value != "CreateException" {
			t.Errorf("ident.Value = %s, want 'CreateException'", ident.Value)
		}
	})

	t.Run("Raise with variable", func(t *testing.T) {
		input := `raise myException;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		raiseStmt, ok := program.Statements[0].(*ast.RaiseStatement)
		if !ok {
			t.Fatalf("statement is not *ast.RaiseStatement, got %T", program.Statements[0])
		}

		if raiseStmt.Exception == nil {
			t.Fatal("raiseStmt.Exception should not be nil")
		}

		ident, ok := raiseStmt.Exception.(*ast.Identifier)
		if !ok {
			t.Fatalf("raiseStmt.Exception is not *ast.Identifier, got %T", raiseStmt.Exception)
		}

		if ident.Value != "myException" {
			t.Errorf("ident.Value = %s, want 'myException'", ident.Value)
		}
	})
}

// Test parsing bare raise statement
func TestParseBareRaiseStatement(t *testing.T) {
	t.Run("Bare raise", func(t *testing.T) {
		input := `raise;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		raiseStmt, ok := program.Statements[0].(*ast.RaiseStatement)
		if !ok {
			t.Fatalf("statement is not *ast.RaiseStatement, got %T", program.Statements[0])
		}

		if raiseStmt.Exception != nil {
			t.Errorf("raiseStmt.Exception should be nil for bare raise, got %T", raiseStmt.Exception)
		}
	})

	t.Run("Bare raise in block", func(t *testing.T) {
		input := `begin
			raise;
		end;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements should contain 1 statement, got %d", len(program.Statements))
		}

		blockStmt, ok := program.Statements[0].(*ast.BlockStatement)
		if !ok {
			t.Fatalf("statement is not *ast.BlockStatement, got %T", program.Statements[0])
		}

		if len(blockStmt.Statements) != 1 {
			t.Fatalf("blockStmt.Statements should contain 1 statement, got %d", len(blockStmt.Statements))
		}

		raiseStmt, ok := blockStmt.Statements[0].(*ast.RaiseStatement)
		if !ok {
			t.Fatalf("statement is not *ast.RaiseStatement, got %T", blockStmt.Statements[0])
		}

		if raiseStmt.Exception != nil {
			t.Errorf("raiseStmt.Exception should be nil for bare raise, got %T", raiseStmt.Exception)
		}
	})
}

// ============================================================================
// Try Statement Parser Tests
// ============================================================================

// Test parsing basic try...except...end statement
func TestParseTryExceptStatement(t *testing.T) {
	t.Run("Try with bare except", func(t *testing.T) {
		input := `try
			x := 42;
		except
			PrintLn('error');
		end;`

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

		if tryStmt.TryBlock == nil {
			t.Fatal("tryStmt.TryBlock should not be nil")
		}

		if tryStmt.ExceptClause == nil {
			t.Fatal("tryStmt.ExceptClause should not be nil")
		}

		if tryStmt.FinallyClause != nil {
			t.Error("tryStmt.FinallyClause should be nil for try...except form")
		}

		// Check try block has statements
		if len(tryStmt.TryBlock.Statements) == 0 {
			t.Error("tryStmt.TryBlock.Statements should not be empty")
		}
	})

	t.Run("Try with single handler", func(t *testing.T) {
		input := `try
			DoSomething();
		except
			on E: Exception do
				PrintLn(E.Message);
		end;`

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

		if len(tryStmt.ExceptClause.Handlers) != 1 {
			t.Fatalf("exceptClause should have 1 handler, got %d", len(tryStmt.ExceptClause.Handlers))
		}

		handler := tryStmt.ExceptClause.Handlers[0]
		if handler.Variable == nil {
			t.Fatal("handler.Variable should not be nil")
		}
		if handler.Variable.Value != "E" {
			t.Errorf("handler.Variable.Value = %s, want 'E'", handler.Variable.Value)
		}

		if handler.ExceptionType == nil {
			t.Fatal("handler.ExceptionType should not be nil")
		}
		if handler.ExceptionType.Name != "Exception" {
			t.Errorf("handler.ExceptionType.Name = %s, want 'Exception'", handler.ExceptionType.Name)
		}

		if handler.Statement == nil {
			t.Fatal("handler.Statement should not be nil")
		}
	})

	t.Run("Try with multiple handlers", func(t *testing.T) {
		input := `try
			DoSomething();
		except
			on E: EMyException do
				HandleMyException(E);
			on E: Exception do
				HandleGeneric(E);
		end;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		tryStmt := program.Statements[0].(*ast.TryStatement)

		if len(tryStmt.ExceptClause.Handlers) != 2 {
			t.Fatalf("exceptClause should have 2 handlers, got %d", len(tryStmt.ExceptClause.Handlers))
		}

		// Check first handler
		if tryStmt.ExceptClause.Handlers[0].ExceptionType.Name != "EMyException" {
			t.Errorf("first handler type = %s, want 'EMyException'",
				tryStmt.ExceptClause.Handlers[0].ExceptionType.Name)
		}

		// Check second handler
		if tryStmt.ExceptClause.Handlers[1].ExceptionType.Name != "Exception" {
			t.Errorf("second handler type = %s, want 'Exception'",
				tryStmt.ExceptClause.Handlers[1].ExceptionType.Name)
		}
	})
}

// Test parsing try...finally...end statement
func TestParseTryFinallyStatement(t *testing.T) {
	t.Run("Try with finally", func(t *testing.T) {
		input := `try
			DoSomething();
		finally
			Cleanup();
		end;`

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

		if tryStmt.TryBlock == nil {
			t.Fatal("tryStmt.TryBlock should not be nil")
		}

		if tryStmt.ExceptClause != nil {
			t.Error("tryStmt.ExceptClause should be nil for try...finally form")
		}

		if tryStmt.FinallyClause == nil {
			t.Fatal("tryStmt.FinallyClause should not be nil")
		}

		// Check finally block has statements
		if tryStmt.FinallyClause.Block == nil {
			t.Fatal("finallyClause.Block should not be nil")
		}

		if len(tryStmt.FinallyClause.Block.Statements) == 0 {
			t.Error("finallyClause.Block.Statements should not be empty")
		}
	})
}

// Test parsing try...except...finally...end statement
func TestParseTryExceptFinallyStatement(t *testing.T) {
	t.Run("Try with except and finally", func(t *testing.T) {
		input := `try
			DoSomething();
		except
			on E: Exception do
				PrintLn(E.Message);
		finally
			Cleanup();
		end;`

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

		if tryStmt.TryBlock == nil {
			t.Fatal("tryStmt.TryBlock should not be nil")
		}

		if tryStmt.ExceptClause == nil {
			t.Fatal("tryStmt.ExceptClause should not be nil")
		}

		if tryStmt.FinallyClause == nil {
			t.Fatal("tryStmt.FinallyClause should not be nil")
		}

		// Verify all three parts exist
		if len(tryStmt.TryBlock.Statements) == 0 {
			t.Error("tryStmt.TryBlock.Statements should not be empty")
		}

		if len(tryStmt.ExceptClause.Handlers) == 0 {
			t.Error("exceptClause.Handlers should not be empty")
		}

		if len(tryStmt.FinallyClause.Block.Statements) == 0 {
			t.Error("finallyClause.Block.Statements should not be empty")
		}
	})
}

// Test error cases
func TestParseTryStatementErrors(t *testing.T) {
	t.Run("Try without except or finally", func(t *testing.T) {
		input := `try
			DoSomething();
		end;`

		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()

		if len(p.Errors()) == 0 {
			t.Fatal("expected parser error for try without except or finally")
		}

		// Should have error about missing except or finally
		hasExpectedError := false
		for _, err := range p.Errors() {
			if containsString(err, "except") || containsString(err, "finally") {
				hasExpectedError = true
				break
			}
		}

		if !hasExpectedError {
			t.Errorf("expected error about missing except or finally, got: %v", p.Errors())
		}

		_ = program // Avoid unused variable error
	})
}

// Helper function to check if a string contains a substring
func containsString(v interface{}, substr string) bool {
	var s string
	switch val := v.(type) {
	case *ParserError:
		if val == nil {
			return false
		}
		s = val.Message
	case string:
		s = val
	default:
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
