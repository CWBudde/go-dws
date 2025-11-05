package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
)

// TestIfStatements tests parsing of if-then-else statements.
func TestIfStatements(t *testing.T) {
	tests := []struct {
		assertions func(*testing.T, *ast.IfStatement)
		name       string
		input      string
	}{
		{
			name:  "simple if without else",
			input: "if x > 0 then PrintLn('positive');",
			assertions: func(t *testing.T, stmt *ast.IfStatement) {
				// Test condition: x > 0
				if !testInfixExpression(t, stmt.Condition, "x", ">", 0) {
					return
				}

				// Test consequence: PrintLn('positive')
				consequence, ok := stmt.Consequence.(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("consequence is not ExpressionStatement. got=%T", stmt.Consequence)
				}

				call, ok := consequence.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("consequence expression is not CallExpression. got=%T", consequence.Expression)
				}

				if !testIdentifier(t, call.Function, "PrintLn") {
					return
				}

				if len(call.Arguments) != 1 {
					t.Fatalf("wrong number of arguments. got=%d", len(call.Arguments))
				}

				if !testStringLiteralExpression(t, call.Arguments[0], "positive") {
					return
				}

				// Test that alternative is nil
				if stmt.Alternative != nil {
					t.Errorf("alternative should be nil. got=%T", stmt.Alternative)
				}
			},
		},
		{
			name:  "if-else with expressions",
			input: "if x > 0 then PrintLn('positive') else PrintLn('non-positive');",
			assertions: func(t *testing.T, stmt *ast.IfStatement) {
				// Test condition
				if !testInfixExpression(t, stmt.Condition, "x", ">", 0) {
					return
				}

				// Test consequence
				consequence, ok := stmt.Consequence.(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("consequence is not ExpressionStatement. got=%T", stmt.Consequence)
				}

				consCall, ok := consequence.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("consequence expression is not CallExpression. got=%T", consequence.Expression)
				}

				if !testIdentifier(t, consCall.Function, "PrintLn") {
					return
				}

				if !testStringLiteralExpression(t, consCall.Arguments[0], "positive") {
					return
				}

				// Test alternative
				if stmt.Alternative == nil {
					t.Fatal("alternative should not be nil")
				}

				alternative, ok := stmt.Alternative.(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("alternative is not ExpressionStatement. got=%T", stmt.Alternative)
				}

				altCall, ok := alternative.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("alternative expression is not CallExpression. got=%T", alternative.Expression)
				}

				if !testIdentifier(t, altCall.Function, "PrintLn") {
					return
				}

				if !testStringLiteralExpression(t, altCall.Arguments[0], "non-positive") {
					return
				}
			},
		},
		{
			name: "if with block consequence",
			input: `if x > 0 then begin
  y := x * 2;
  PrintLn(y);
end;`,
			assertions: func(t *testing.T, stmt *ast.IfStatement) {
				// Test condition
				if !testInfixExpression(t, stmt.Condition, "x", ">", 0) {
					return
				}

				// Test consequence is a block
				block, ok := stmt.Consequence.(*ast.BlockStatement)
				if !ok {
					t.Fatalf("consequence is not BlockStatement. got=%T", stmt.Consequence)
				}

				if len(block.Statements) != 2 {
					t.Fatalf("block has %d statements, want 2", len(block.Statements))
				}

				// First statement in block: y := x * 2;
				assign, ok := block.Statements[0].(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("first block statement is not AssignmentStatement. got=%T", block.Statements[0])
				}

				assignTarget, ok := assign.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("assign.Target is not *ast.Identifier. got=%T", assign.Target)
				}
				if assignTarget.Value != "y" {
					t.Errorf("assignment name = %q, want 'y'", assignTarget.Value)
				}

				if !testInfixExpression(t, assign.Value, "x", "*", 2) {
					return
				}

				// Second statement in block: PrintLn(y);
				exprStmt, ok := block.Statements[1].(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("second block statement is not ExpressionStatement. got=%T", block.Statements[1])
				}

				call, ok := exprStmt.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("second block statement expression is not CallExpression. got=%T", exprStmt.Expression)
				}

				if !testIdentifier(t, call.Function, "PrintLn") {
					return
				}
			},
		},
		{
			name: "if-else with blocks",
			input: `if x > 0 then begin
  PrintLn('positive');
end else begin
  PrintLn('non-positive');
end;`,
			assertions: func(t *testing.T, stmt *ast.IfStatement) {
				// Test condition
				if !testInfixExpression(t, stmt.Condition, "x", ">", 0) {
					return
				}

				// Test consequence block
				consBlock, ok := stmt.Consequence.(*ast.BlockStatement)
				if !ok {
					t.Fatalf("consequence is not BlockStatement. got=%T", stmt.Consequence)
				}

				if len(consBlock.Statements) != 1 {
					t.Fatalf("consequence block has %d statements, want 1", len(consBlock.Statements))
				}

				// Test alternative block
				if stmt.Alternative == nil {
					t.Fatal("alternative should not be nil")
				}

				altBlock, ok := stmt.Alternative.(*ast.BlockStatement)
				if !ok {
					t.Fatalf("alternative is not BlockStatement. got=%T", stmt.Alternative)
				}

				if len(altBlock.Statements) != 1 {
					t.Fatalf("alternative block has %d statements, want 1", len(altBlock.Statements))
				}
			},
		},
		{
			name: "nested if statements",
			input: `if x > 0 then
  if y > 0 then
    PrintLn('both positive')
  else
    PrintLn('x positive, y not');`,
			assertions: func(t *testing.T, stmt *ast.IfStatement) {
				// Test outer condition
				if !testInfixExpression(t, stmt.Condition, "x", ">", 0) {
					return
				}

				// Test that consequence is another if statement
				innerIf, ok := stmt.Consequence.(*ast.IfStatement)
				if !ok {
					t.Fatalf("consequence is not IfStatement. got=%T", stmt.Consequence)
				}

				// Test inner condition
				if !testInfixExpression(t, innerIf.Condition, "y", ">", 0) {
					return
				}

				// Verify inner if has both consequence and alternative
				if innerIf.Consequence == nil {
					t.Fatal("inner if consequence is nil")
				}

				if innerIf.Alternative == nil {
					t.Fatal("inner if alternative is nil")
				}
			},
		},
		{
			name:  "if with complex condition",
			input: "if (x > 0) and (y < 10) then PrintLn('in range');",
			assertions: func(t *testing.T, stmt *ast.IfStatement) {
				// Test that condition is a binary expression with 'and'
				binExp, ok := stmt.Condition.(*ast.BinaryExpression)
				if !ok {
					t.Fatalf("condition is not BinaryExpression. got=%T", stmt.Condition)
				}

				if binExp.Operator != "and" {
					t.Errorf("condition operator = %q, want 'and'", binExp.Operator)
				}

				// Test left side: x > 0
				if !testInfixExpression(t, binExp.Left, "x", ">", 0) {
					return
				}

				// Test right side: y < 10
				if !testInfixExpression(t, binExp.Right, "y", "<", 10) {
					return
				}
			},
		},
		{
			name:  "if with assignment in consequence",
			input: `if x = 0 then x := 1;`,
			assertions: func(t *testing.T, stmt *ast.IfStatement) {
				// Test condition: x = 0
				if !testInfixExpression(t, stmt.Condition, "x", "=", 0) {
					return
				}

				// Test consequence: x := 1
				assign, ok := stmt.Consequence.(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("consequence is not AssignmentStatement. got=%T", stmt.Consequence)
				}

				assignTarget, ok := assign.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("assign.Target is not *ast.Identifier. got=%T", assign.Target)
				}
				if assignTarget.Value != "x" {
					t.Errorf("assignment name = %q, want 'x'", assignTarget.Value)
				}

				if !testIntegerLiteral(t, assign.Value, 1) {
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

			if len(program.Statements) != 1 {
				t.Fatalf("program has %d statements, want 1", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.IfStatement)
			if !ok {
				t.Fatalf("statement is not ast.IfStatement. got=%T", program.Statements[0])
			}

			tt.assertions(t, stmt)
		})
	}
}

// TestCaseStatements tests parsing of case statement.
func TestCaseStatements(t *testing.T) {
	tests := []struct {
		assertions func(*testing.T, *ast.CaseStatement)
		name       string
		input      string
	}{
		{
			name: "simple case with single value branches",
			input: `case x of
  1: PrintLn('one');
  2: PrintLn('two');
  3: PrintLn('three');
end;`,
			assertions: func(t *testing.T, stmt *ast.CaseStatement) {
				// Test case expression
				if !testIdentifier(t, stmt.Expression, "x") {
					return
				}

				// Test number of case branches
				if len(stmt.Cases) != 3 {
					t.Fatalf("case has %d branches, want 3", len(stmt.Cases))
				}

				// Test first branch: 1: PrintLn('one');
				branch1 := stmt.Cases[0]
				if len(branch1.Values) != 1 {
					t.Fatalf("branch1 has %d values, want 1", len(branch1.Values))
				}
				if !testIntegerLiteral(t, branch1.Values[0], 1) {
					return
				}

				// Test second branch: 2: PrintLn('two');
				branch2 := stmt.Cases[1]
				if len(branch2.Values) != 1 {
					t.Fatalf("branch2 has %d values, want 1", len(branch2.Values))
				}
				if !testIntegerLiteral(t, branch2.Values[0], 2) {
					return
				}

				// Test third branch: 3: PrintLn('three');
				branch3 := stmt.Cases[2]
				if len(branch3.Values) != 1 {
					t.Fatalf("branch3 has %d values, want 1", len(branch3.Values))
				}
				if !testIntegerLiteral(t, branch3.Values[0], 3) {
					return
				}

				// Test that there's no else branch
				if stmt.Else != nil {
					t.Errorf("else branch should be nil, got %T", stmt.Else)
				}
			},
		},
		{
			name: "case with multiple values per branch",
			input: `case x of
  1, 2, 3: PrintLn('one to three');
  4, 5: PrintLn('four or five');
end;`,
			assertions: func(t *testing.T, stmt *ast.CaseStatement) {
				// Test case expression
				if !testIdentifier(t, stmt.Expression, "x") {
					return
				}

				// Test number of branches
				if len(stmt.Cases) != 2 {
					t.Fatalf("case has %d branches, want 2", len(stmt.Cases))
				}

				// Test first branch: 1, 2, 3
				branch1 := stmt.Cases[0]
				if len(branch1.Values) != 3 {
					t.Fatalf("branch1 has %d values, want 3", len(branch1.Values))
				}
				if !testIntegerLiteral(t, branch1.Values[0], 1) {
					return
				}
				if !testIntegerLiteral(t, branch1.Values[1], 2) {
					return
				}
				if !testIntegerLiteral(t, branch1.Values[2], 3) {
					return
				}

				// Test second branch: 4, 5
				branch2 := stmt.Cases[1]
				if len(branch2.Values) != 2 {
					t.Fatalf("branch2 has %d values, want 2", len(branch2.Values))
				}
				if !testIntegerLiteral(t, branch2.Values[0], 4) {
					return
				}
				if !testIntegerLiteral(t, branch2.Values[1], 5) {
					return
				}
			},
		},
		{
			name: "case with else branch",
			input: `case x of
  1: PrintLn('one');
  2: PrintLn('two');
else
  PrintLn('other');
end;`,
			assertions: func(t *testing.T, stmt *ast.CaseStatement) {
				// Test case expression
				if !testIdentifier(t, stmt.Expression, "x") {
					return
				}

				// Test number of branches
				if len(stmt.Cases) != 2 {
					t.Fatalf("case has %d branches, want 2", len(stmt.Cases))
				}

				// Test else branch exists
				if stmt.Else == nil {
					t.Fatal("else branch should not be nil")
				}

				// Test else branch is a PrintLn call
				elseExpr, ok := stmt.Else.(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("else branch is not ExpressionStatement. got=%T", stmt.Else)
				}

				call, ok := elseExpr.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("else expression is not CallExpression. got=%T", elseExpr.Expression)
				}

				if !testIdentifier(t, call.Function, "PrintLn") {
					return
				}

				if !testStringLiteralExpression(t, call.Arguments[0], "other") {
					return
				}
			},
		},
		{
			name: "case with block statements",
			input: `case x of
  1: begin
    y := 1;
    PrintLn(y);
  end;
  2: begin
    y := 2;
    PrintLn(y);
  end;
end;`,
			assertions: func(t *testing.T, stmt *ast.CaseStatement) {
				// Test case expression
				if !testIdentifier(t, stmt.Expression, "x") {
					return
				}

				// Test number of branches
				if len(stmt.Cases) != 2 {
					t.Fatalf("case has %d branches, want 2", len(stmt.Cases))
				}

				// Test first branch has a block statement
				branch1 := stmt.Cases[0]
				block1, ok := branch1.Statement.(*ast.BlockStatement)
				if !ok {
					t.Fatalf("branch1 statement is not BlockStatement. got=%T", branch1.Statement)
				}

				if len(block1.Statements) != 2 {
					t.Fatalf("branch1 block has %d statements, want 2", len(block1.Statements))
				}

				// Test second branch has a block statement
				branch2 := stmt.Cases[1]
				block2, ok := branch2.Statement.(*ast.BlockStatement)
				if !ok {
					t.Fatalf("branch2 statement is not BlockStatement. got=%T", branch2.Statement)
				}

				if len(block2.Statements) != 2 {
					t.Fatalf("branch2 block has %d statements, want 2", len(block2.Statements))
				}
			},
		},
		{
			name: "case with string expression and string values",
			input: `case name of
  'Alice', 'Bob': PrintLn('known person');
  'Unknown': PrintLn('stranger');
else
  PrintLn('no match');
end;`,
			assertions: func(t *testing.T, stmt *ast.CaseStatement) {
				// Test case expression
				if !testIdentifier(t, stmt.Expression, "name") {
					return
				}

				// Test number of branches
				if len(stmt.Cases) != 2 {
					t.Fatalf("case has %d branches, want 2", len(stmt.Cases))
				}

				// Test first branch has 2 string values
				branch1 := stmt.Cases[0]
				if len(branch1.Values) != 2 {
					t.Fatalf("branch1 has %d values, want 2", len(branch1.Values))
				}
				if !testStringLiteralExpression(t, branch1.Values[0], "Alice") {
					return
				}
				if !testStringLiteralExpression(t, branch1.Values[1], "Bob") {
					return
				}

				// Test second branch has 1 string value
				branch2 := stmt.Cases[1]
				if len(branch2.Values) != 1 {
					t.Fatalf("branch2 has %d values, want 1", len(branch2.Values))
				}
				if !testStringLiteralExpression(t, branch2.Values[0], "Unknown") {
					return
				}

				// Test else branch exists
				if stmt.Else == nil {
					t.Fatal("else branch should not be nil")
				}
			},
		},
		{
			name: "case with complex expression",
			input: `case x + y of
  0: PrintLn('zero');
  1: PrintLn('one');
end;`,
			assertions: func(t *testing.T, stmt *ast.CaseStatement) {
				// Test case expression is a binary expression
				if !testInfixExpression(t, stmt.Expression, "x", "+", "y") {
					return
				}

				// Test number of branches
				if len(stmt.Cases) != 2 {
					t.Fatalf("case has %d branches, want 2", len(stmt.Cases))
				}
			},
		},
		{
			name: "case with assignment in branch",
			input: `case status of
  0: result := 'failed';
  1: result := 'success';
else
  result := 'unknown';
end;`,
			assertions: func(t *testing.T, stmt *ast.CaseStatement) {
				// Test case expression
				if !testIdentifier(t, stmt.Expression, "status") {
					return
				}

				// Test number of branches
				if len(stmt.Cases) != 2 {
					t.Fatalf("case has %d branches, want 2", len(stmt.Cases))
				}

				// Test first branch is an assignment
				branch1 := stmt.Cases[0]
				assign1, ok := branch1.Statement.(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("branch1 statement is not AssignmentStatement. got=%T", branch1.Statement)
				}

				assign1Target, ok := assign1.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("assign1.Target is not *ast.Identifier. got=%T", assign1.Target)
				}
				if assign1Target.Value != "result" {
					t.Errorf("branch1 assignment name = %q, want 'result'", assign1Target.Value)
				}

				if !testStringLiteralExpression(t, assign1.Value, "failed") {
					return
				}

				// Test else branch is also an assignment
				if stmt.Else == nil {
					t.Fatal("else branch should not be nil")
				}

				elseAssign, ok := stmt.Else.(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("else branch is not AssignmentStatement. got=%T", stmt.Else)
				}

				elseAssignTarget, ok := elseAssign.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("elseAssign.Target is not *ast.Identifier. got=%T", elseAssign.Target)
				}
				if elseAssignTarget.Value != "result" {
					t.Errorf("else assignment name = %q, want 'result'", elseAssignTarget.Value)
				}

				if !testStringLiteralExpression(t, elseAssign.Value, "unknown") {
					return
				}
			},
		},
		{
			name: "case with expression values",
			input: `case x of
  min_val: PrintLn('minimum');
  max_val: PrintLn('maximum');
end;`,
			assertions: func(t *testing.T, stmt *ast.CaseStatement) {
				// Test case expression
				if !testIdentifier(t, stmt.Expression, "x") {
					return
				}

				// Test number of branches
				if len(stmt.Cases) != 2 {
					t.Fatalf("case has %d branches, want 2", len(stmt.Cases))
				}

				// Test first branch value is an identifier
				branch1 := stmt.Cases[0]
				if !testIdentifier(t, branch1.Values[0], "min_val") {
					return
				}

				// Test second branch value is an identifier
				branch2 := stmt.Cases[1]
				if !testIdentifier(t, branch2.Values[0], "max_val") {
					return
				}
			},
		},
		{
			name: "case with multiple statements in else clause",
			input: `case x of
  1: PrintLn('one');
else
  Line := 'other';
  Counter := Counter - 1;
end;`,
			assertions: func(t *testing.T, stmt *ast.CaseStatement) {
				// Test case expression
				if !testIdentifier(t, stmt.Expression, "x") {
					return
				}

				// Test number of branches
				if len(stmt.Cases) != 1 {
					t.Fatalf("case has %d branches, want 1", len(stmt.Cases))
				}

				// Test else branch exists and is a BlockStatement
				if stmt.Else == nil {
					t.Fatal("else branch should not be nil")
				}

				block, ok := stmt.Else.(*ast.BlockStatement)
				if !ok {
					t.Fatalf("else branch is not BlockStatement. got=%T", stmt.Else)
				}

				// Test that block has 2 statements
				if len(block.Statements) != 2 {
					t.Fatalf("else block has %d statements, want 2", len(block.Statements))
				}

				// Test first statement: Line := 'other'
				assign1, ok := block.Statements[0].(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("first statement is not AssignmentStatement. got=%T", block.Statements[0])
				}

				assign1Target, ok := assign1.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("assign1.Target is not *ast.Identifier. got=%T", assign1.Target)
				}
				if assign1Target.Value != "Line" {
					t.Errorf("first assignment name = %q, want 'Line'", assign1Target.Value)
				}

				if !testStringLiteralExpression(t, assign1.Value, "other") {
					return
				}

				// Test second statement: Counter := Counter - 1
				assign2, ok := block.Statements[1].(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("second statement is not AssignmentStatement. got=%T", block.Statements[1])
				}

				assign2Target, ok := assign2.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("assign2.Target is not *ast.Identifier. got=%T", assign2.Target)
				}
				if assign2Target.Value != "Counter" {
					t.Errorf("second assignment name = %q, want 'Counter'", assign2Target.Value)
				}

				// Test Counter - 1 expression
				if !testInfixExpression(t, assign2.Value, "Counter", "-", 1) {
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

			if len(program.Statements) != 1 {
				t.Fatalf("program has %d statements, want 1", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.CaseStatement)
			if !ok {
				t.Fatalf("statement is not ast.CaseStatement. got=%T", program.Statements[0])
			}

			tt.assertions(t, stmt)
		})
	}
}

// TestCaseStatementsWithRanges tests parsing of case statements with range expressions.
func TestCaseStatementsWithRanges(t *testing.T) {
	tests := []struct {
		assertions func(*testing.T, *ast.CaseStatement)
		name       string
		input      string
	}{
		{
			name: "case with character range",
			input: `case ch of
  'A'..'Z': PrintLn('uppercase');
  'a'..'z': PrintLn('lowercase');
end;`,
			assertions: func(t *testing.T, stmt *ast.CaseStatement) {
				// Test case expression
				if !testIdentifier(t, stmt.Expression, "ch") {
					return
				}

				// Test number of branches
				if len(stmt.Cases) != 2 {
					t.Fatalf("case has %d branches, want 2", len(stmt.Cases))
				}

				// Test first branch: 'A'..'Z'
				branch1 := stmt.Cases[0]
				if len(branch1.Values) != 1 {
					t.Fatalf("branch1 has %d values, want 1", len(branch1.Values))
				}

				rangeExpr1, ok := branch1.Values[0].(*ast.RangeExpression)
				if !ok {
					t.Fatalf("branch1 value is not RangeExpression. got=%T", branch1.Values[0])
				}

				if !testStringLiteralExpression(t, rangeExpr1.Start, "A") {
					return
				}
				if !testStringLiteralExpression(t, rangeExpr1.End, "Z") {
					return
				}

				// Test second branch: 'a'..'z'
				branch2 := stmt.Cases[1]
				if len(branch2.Values) != 1 {
					t.Fatalf("branch2 has %d values, want 1", len(branch2.Values))
				}

				rangeExpr2, ok := branch2.Values[0].(*ast.RangeExpression)
				if !ok {
					t.Fatalf("branch2 value is not RangeExpression. got=%T", branch2.Values[0])
				}

				if !testStringLiteralExpression(t, rangeExpr2.Start, "a") {
					return
				}
				if !testStringLiteralExpression(t, rangeExpr2.End, "z") {
					return
				}
			},
		},
		{
			name: "case with integer range",
			input: `case x of
  1..10: PrintLn('1-10');
  11..20: PrintLn('11-20');
end;`,
			assertions: func(t *testing.T, stmt *ast.CaseStatement) {
				// Test case expression
				if !testIdentifier(t, stmt.Expression, "x") {
					return
				}

				// Test number of branches
				if len(stmt.Cases) != 2 {
					t.Fatalf("case has %d branches, want 2", len(stmt.Cases))
				}

				// Test first branch: 1..10
				branch1 := stmt.Cases[0]
				rangeExpr1, ok := branch1.Values[0].(*ast.RangeExpression)
				if !ok {
					t.Fatalf("branch1 value is not RangeExpression. got=%T", branch1.Values[0])
				}

				if !testIntegerLiteral(t, rangeExpr1.Start, 1) {
					return
				}
				if !testIntegerLiteral(t, rangeExpr1.End, 10) {
					return
				}

				// Test second branch: 11..20
				branch2 := stmt.Cases[1]
				rangeExpr2, ok := branch2.Values[0].(*ast.RangeExpression)
				if !ok {
					t.Fatalf("branch2 value is not RangeExpression. got=%T", branch2.Values[0])
				}

				if !testIntegerLiteral(t, rangeExpr2.Start, 11) {
					return
				}
				if !testIntegerLiteral(t, rangeExpr2.End, 20) {
					return
				}
			},
		},
		{
			name: "case with mixed ranges and single values",
			input: `case x of
  1, 3..5, 7: PrintLn('odd or range');
  2, 4, 6: PrintLn('even');
end;`,
			assertions: func(t *testing.T, stmt *ast.CaseStatement) {
				// Test case expression
				if !testIdentifier(t, stmt.Expression, "x") {
					return
				}

				// Test number of branches
				if len(stmt.Cases) != 2 {
					t.Fatalf("case has %d branches, want 2", len(stmt.Cases))
				}

				// Test first branch: 1, 3..5, 7
				branch1 := stmt.Cases[0]
				if len(branch1.Values) != 3 {
					t.Fatalf("branch1 has %d values, want 3", len(branch1.Values))
				}

				// First value: 1 (integer literal)
				if !testIntegerLiteral(t, branch1.Values[0], 1) {
					return
				}

				// Second value: 3..5 (range)
				rangeExpr, ok := branch1.Values[1].(*ast.RangeExpression)
				if !ok {
					t.Fatalf("branch1 value[1] is not RangeExpression. got=%T", branch1.Values[1])
				}
				if !testIntegerLiteral(t, rangeExpr.Start, 3) {
					return
				}
				if !testIntegerLiteral(t, rangeExpr.End, 5) {
					return
				}

				// Third value: 7 (integer literal)
				if !testIntegerLiteral(t, branch1.Values[2], 7) {
					return
				}

				// Test second branch: 2, 4, 6 (all single values)
				branch2 := stmt.Cases[1]
				if len(branch2.Values) != 3 {
					t.Fatalf("branch2 has %d values, want 3", len(branch2.Values))
				}

				if !testIntegerLiteral(t, branch2.Values[0], 2) {
					return
				}
				if !testIntegerLiteral(t, branch2.Values[1], 4) {
					return
				}
				if !testIntegerLiteral(t, branch2.Values[2], 6) {
					return
				}
			},
		},
		{
			name: "case with range and else",
			input: `case x of
  0..9: PrintLn('single digit');
  10..99: PrintLn('double digit');
else
  PrintLn('other');
end;`,
			assertions: func(t *testing.T, stmt *ast.CaseStatement) {
				// Test case expression
				if !testIdentifier(t, stmt.Expression, "x") {
					return
				}

				// Test number of branches
				if len(stmt.Cases) != 2 {
					t.Fatalf("case has %d branches, want 2", len(stmt.Cases))
				}

				// Test first branch: 0..9
				branch1 := stmt.Cases[0]
				rangeExpr1, ok := branch1.Values[0].(*ast.RangeExpression)
				if !ok {
					t.Fatalf("branch1 value is not RangeExpression. got=%T", branch1.Values[0])
				}

				if !testIntegerLiteral(t, rangeExpr1.Start, 0) {
					return
				}
				if !testIntegerLiteral(t, rangeExpr1.End, 9) {
					return
				}

				// Test else branch exists
				if stmt.Else == nil {
					t.Fatal("else branch should not be nil")
				}
			},
		},
		{
			name: "case with variable range bounds",
			input: `case x of
  min_val..max_val: PrintLn('in range');
end;`,
			assertions: func(t *testing.T, stmt *ast.CaseStatement) {
				// Test case expression
				if !testIdentifier(t, stmt.Expression, "x") {
					return
				}

				// Test branch has range with identifier bounds
				branch := stmt.Cases[0]
				rangeExpr, ok := branch.Values[0].(*ast.RangeExpression)
				if !ok {
					t.Fatalf("branch value is not RangeExpression. got=%T", branch.Values[0])
				}

				if !testIdentifier(t, rangeExpr.Start, "min_val") {
					return
				}
				if !testIdentifier(t, rangeExpr.End, "max_val") {
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

			if len(program.Statements) != 1 {
				t.Fatalf("program has %d statements, want 1", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.CaseStatement)
			if !ok {
				t.Fatalf("statement is not ast.CaseStatement. got=%T", program.Statements[0])
			}

			tt.assertions(t, stmt)
		})
	}
}
