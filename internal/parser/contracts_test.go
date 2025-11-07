package parser

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// TestParsePreconditionsSingle tests parsing a function with a single precondition.
func TestParsePreconditionsSingle(t *testing.T) {
	input := `
function Max(a, b: Integer): Integer;
require
	a <> b;
begin
	if a > b then
		Result := a
	else
		Result := b;
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program should have 1 statement, got=%d", len(program.Statements))
	}

	fnDecl, ok := program.Statements[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("expected *ast.FunctionDecl, got=%T", program.Statements[0])
	}

	if fnDecl.PreConditions == nil {
		t.Fatal("expected preconditions, got nil")
	}

	if len(fnDecl.PreConditions.Conditions) != 1 {
		t.Fatalf("expected 1 precondition, got=%d", len(fnDecl.PreConditions.Conditions))
	}

	// Check the condition expression
	condition := fnDecl.PreConditions.Conditions[0]
	if condition.Test == nil {
		t.Fatal("expected condition test expression, got nil")
	}

	// The test expression should be "a <> b"
	binaryExpr, ok := condition.Test.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("expected *ast.BinaryExpression, got=%T", condition.Test)
	}

	if binaryExpr.Operator != "<>" {
		t.Errorf("expected operator '<>', got='%s'", binaryExpr.Operator)
	}

	// No custom message
	if condition.Message != nil {
		t.Errorf("expected no message, got=%v", condition.Message)
	}
}

// TestParsePreconditionsMultiple tests parsing a function with multiple preconditions.
func TestParsePreconditionsMultiple(t *testing.T) {
	input := `
function DotProduct(a, b: array of Float): Float;
require
	a.Length = b.Length;
	a.Length > 0;
begin
	Result := 0.0;
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program should have 1 statement, got=%d", len(program.Statements))
	}

	fnDecl, ok := program.Statements[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("expected *ast.FunctionDecl, got=%T", program.Statements[0])
	}

	if fnDecl.PreConditions == nil {
		t.Fatal("expected preconditions, got nil")
	}

	if len(fnDecl.PreConditions.Conditions) != 2 {
		t.Fatalf("expected 2 preconditions, got=%d", len(fnDecl.PreConditions.Conditions))
	}

	// Both conditions should have test expressions
	for i, condition := range fnDecl.PreConditions.Conditions {
		if condition.Test == nil {
			t.Errorf("condition %d: expected test expression, got nil", i)
		}
	}
}

// TestParsePreconditionsWithMessage tests parsing preconditions with custom error messages.
func TestParsePreconditionsWithMessage(t *testing.T) {
	input := `
function Divide(a, b: Float): Float;
require
	b <> 0.0 : 'divisor cannot be zero';
begin
	Result := a / b;
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program should have 1 statement, got=%d", len(program.Statements))
	}

	fnDecl, ok := program.Statements[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("expected *ast.FunctionDecl, got=%T", program.Statements[0])
	}

	if fnDecl.PreConditions == nil {
		t.Fatal("expected preconditions, got nil")
	}

	if len(fnDecl.PreConditions.Conditions) != 1 {
		t.Fatalf("expected 1 precondition, got=%d", len(fnDecl.PreConditions.Conditions))
	}

	condition := fnDecl.PreConditions.Conditions[0]
	if condition.Message == nil {
		t.Fatal("expected custom message, got nil")
	}

	msgLiteral, ok := condition.Message.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("expected *ast.StringLiteral, got=%T", condition.Message)
	}

	expectedMsg := "divisor cannot be zero"
	if msgLiteral.Value != expectedMsg {
		t.Errorf("expected message '%s', got='%s'", expectedMsg, msgLiteral.Value)
	}
}

// TestParsePostconditionsSingle tests parsing a function with a single postcondition.
func TestParsePostconditionsSingle(t *testing.T) {
	input := `
function Abs(x: Integer): Integer;
begin
	if x < 0 then
		Result := -x
	else
		Result := x;
end;
ensure
	Result >= 0;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program should have 1 statement, got=%d", len(program.Statements))
	}

	fnDecl, ok := program.Statements[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("expected *ast.FunctionDecl, got=%T", program.Statements[0])
	}

	if fnDecl.PostConditions == nil {
		t.Fatal("expected postconditions, got nil")
	}

	if len(fnDecl.PostConditions.Conditions) != 1 {
		t.Fatalf("expected 1 postcondition, got=%d", len(fnDecl.PostConditions.Conditions))
	}

	condition := fnDecl.PostConditions.Conditions[0]
	if condition.Test == nil {
		t.Fatal("expected condition test expression, got nil")
	}
}

// TestParsePostconditionsWithOld tests parsing postconditions with 'old' expressions.
func TestParsePostconditionsWithOld(t *testing.T) {
	input := `
function Increment(x: Integer): Integer;
begin
	Result := x + 1;
end;
ensure
	Result = old x + 1;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program should have 1 statement, got=%d", len(program.Statements))
	}

	fnDecl, ok := program.Statements[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("expected *ast.FunctionDecl, got=%T", program.Statements[0])
	}

	if fnDecl.PostConditions == nil {
		t.Fatal("expected postconditions, got nil")
	}

	if len(fnDecl.PostConditions.Conditions) != 1 {
		t.Fatalf("expected 1 postcondition, got=%d", len(fnDecl.PostConditions.Conditions))
	}

	condition := fnDecl.PostConditions.Conditions[0]

	// The test expression should be "Result = old x + 1"
	// This is: Result = (old x + 1)
	binaryExpr, ok := condition.Test.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("expected *ast.BinaryExpression for '=', got=%T", condition.Test)
	}

	if binaryExpr.Operator != "=" {
		t.Errorf("expected operator '=', got='%s'", binaryExpr.Operator)
	}

	// Right side should be "old x + 1"
	rightBinary, ok := binaryExpr.Right.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("expected right side to be *ast.BinaryExpression, got=%T", binaryExpr.Right)
	}

	// Left side of the addition should be "old x"
	oldExpr, ok := rightBinary.Left.(*ast.OldExpression)
	if !ok {
		t.Fatalf("expected *ast.OldExpression, got=%T", rightBinary.Left)
	}

	if oldExpr.Identifier.Value != "x" {
		t.Errorf("expected old identifier 'x', got='%s'", oldExpr.Identifier.Value)
	}
}

// TestParseCombinedContracts tests parsing a function with both pre and postconditions.
func TestParseCombinedContracts(t *testing.T) {
	input := `
function SafeDivide(a, b: Float): Float;
require
	b <> 0.0 : 'divisor must not be zero';
begin
	Result := a / b;
end;
ensure
	Result * b = a : 'result verification failed';
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program should have 1 statement, got=%d", len(program.Statements))
	}

	fnDecl, ok := program.Statements[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("expected *ast.FunctionDecl, got=%T", program.Statements[0])
	}

	// Check preconditions
	if fnDecl.PreConditions == nil {
		t.Fatal("expected preconditions, got nil")
	}
	if len(fnDecl.PreConditions.Conditions) != 1 {
		t.Fatalf("expected 1 precondition, got=%d", len(fnDecl.PreConditions.Conditions))
	}

	// Check postconditions
	if fnDecl.PostConditions == nil {
		t.Fatal("expected postconditions, got nil")
	}
	if len(fnDecl.PostConditions.Conditions) != 1 {
		t.Fatalf("expected 1 postcondition, got=%d", len(fnDecl.PostConditions.Conditions))
	}

	// Both should have custom messages
	if fnDecl.PreConditions.Conditions[0].Message == nil {
		t.Error("expected precondition message, got nil")
	}
	if fnDecl.PostConditions.Conditions[0].Message == nil {
		t.Error("expected postcondition message, got nil")
	}
}

// TestParseContractsWithLocalVars tests contracts with local variable declarations.
func TestParseContractsWithLocalVars(t *testing.T) {
	input := `
function Calculate(x: Integer): Integer;
require
	x > 0;
var
	temp: Integer;
begin
	temp := x * 2;
	Result := temp + 1;
end;
ensure
	Result = old x * 2 + 1;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program should have 1 statement, got=%d", len(program.Statements))
	}

	fnDecl, ok := program.Statements[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("expected *ast.FunctionDecl, got=%T", program.Statements[0])
	}

	if fnDecl.PreConditions == nil {
		t.Error("expected preconditions, got nil")
	}
	if fnDecl.PostConditions == nil {
		t.Error("expected postconditions, got nil")
	}
	if fnDecl.Body == nil {
		t.Error("expected function body, got nil")
	}
}

// TestParseOldOutsidePostconditionError tests that 'old' outside postconditions causes an error.
func TestParseOldOutsidePostconditionError(t *testing.T) {
	input := `
function Invalid(x: Integer): Integer;
require
	old x > 0;
begin
	Result := x;
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	// We expect an error about 'old' being used outside postconditions
	errors := p.Errors()
	if len(errors) == 0 {
		t.Fatal("expected parser error for 'old' in precondition, got none")
	}

	// Check that error message mentions postconditions
	foundRelevantError := false
	for _, err := range errors {
		if strings.Contains(err.Message, "postcondition") || strings.Contains(err.Message, "old") {
			foundRelevantError = true
			break
		}
	}

	if !foundRelevantError {
		t.Errorf("expected error about 'old' keyword, got errors: %v", errors)
	}

	// Program should still be parsed (error recovery)
	if program == nil {
		t.Fatal("expected program to be parsed despite errors")
	}
}

// TestParseMultiplePostconditions tests parsing multiple postconditions.
func TestParseMultiplePostconditions(t *testing.T) {
	input := `
function Clamp(x, min, max: Integer): Integer;
require
	min <= max;
begin
	if x < min then
		Result := min
	else if x > max then
		Result := max
	else
		Result := x;
end;
ensure
	Result >= min;
	Result <= max;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program should have 1 statement, got=%d", len(program.Statements))
	}

	fnDecl, ok := program.Statements[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("expected *ast.FunctionDecl, got=%T", program.Statements[0])
	}

	if fnDecl.PostConditions == nil {
		t.Fatal("expected postconditions, got nil")
	}

	if len(fnDecl.PostConditions.Conditions) != 2 {
		t.Fatalf("expected 2 postconditions, got=%d", len(fnDecl.PostConditions.Conditions))
	}
}
