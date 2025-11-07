package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
)

// TestWhileStatements tests parsing of while loop statements.
func TestWhileStatements(t *testing.T) {
	tests := []struct {
		assertions func(*testing.T, *ast.WhileStatement)
		name       string
		input      string
	}{
		{
			name:  "simple while loop",
			input: "while x < 10 do x := x + 1;",
			assertions: func(t *testing.T, stmt *ast.WhileStatement) {
				// Test condition: x < 10
				if !testInfixExpression(t, stmt.Condition, "x", "<", 10) {
					return
				}

				// Test body: x := x + 1
				assign, ok := stmt.Body.(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("body is not AssignmentStatement. got=%T", stmt.Body)
				}

				assignTarget, ok := assign.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("assign.Target is not *ast.Identifier. got=%T", assign.Target)
				}
				if assignTarget.Value != "x" {
					t.Errorf("assignment name = %q, want 'x'", assignTarget.Value)
				}

				if !testInfixExpression(t, assign.Value, "x", "+", 1) {
					return
				}
			},
		},
		{
			name: "while with block body",
			input: `while x < 10 do begin
  x := x + 1;
  PrintLn(x);
end;`,
			assertions: func(t *testing.T, stmt *ast.WhileStatement) {
				// Test condition
				if !testInfixExpression(t, stmt.Condition, "x", "<", 10) {
					return
				}

				// Test body is a block
				block, ok := stmt.Body.(*ast.BlockStatement)
				if !ok {
					t.Fatalf("body is not BlockStatement. got=%T", stmt.Body)
				}

				if len(block.Statements) != 2 {
					t.Fatalf("block has %d statements, want 2", len(block.Statements))
				}

				// First statement: x := x + 1;
				assign, ok := block.Statements[0].(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("first block statement is not AssignmentStatement. got=%T", block.Statements[0])
				}

				assignTarget, ok := assign.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("assign.Target is not *ast.Identifier. got=%T", assign.Target)
				}
				if assignTarget.Value != "x" {
					t.Errorf("assignment name = %q, want 'x'", assignTarget.Value)
				}

				// Second statement: PrintLn(x);
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
			name:  "while with complex condition",
			input: "while (x > 0) and (x < 100) do x := x * 2;",
			assertions: func(t *testing.T, stmt *ast.WhileStatement) {
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

				// Test right side: x < 100
				if !testInfixExpression(t, binExp.Right, "x", "<", 100) {
					return
				}

				// Test body
				assign, ok := stmt.Body.(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("body is not AssignmentStatement. got=%T", stmt.Body)
				}

				if !testInfixExpression(t, assign.Value, "x", "*", 2) {
					return
				}
			},
		},
		{
			name: "nested while loops",
			input: `while x < 10 do
  while y < 5 do
    y := y + 1;`,
			assertions: func(t *testing.T, stmt *ast.WhileStatement) {
				// Test outer condition
				if !testInfixExpression(t, stmt.Condition, "x", "<", 10) {
					return
				}

				// Test that body is another while statement
				innerWhile, ok := stmt.Body.(*ast.WhileStatement)
				if !ok {
					t.Fatalf("body is not WhileStatement. got=%T", stmt.Body)
				}

				// Test inner condition
				if !testInfixExpression(t, innerWhile.Condition, "y", "<", 5) {
					return
				}

				// Test inner body
				innerAssign, ok := innerWhile.Body.(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("inner while body is not AssignmentStatement. got=%T", innerWhile.Body)
				}

				innerAssignTarget, ok := innerAssign.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("innerAssign.Target is not *ast.Identifier. got=%T", innerAssign.Target)
				}
				if innerAssignTarget.Value != "y" {
					t.Errorf("inner assignment name = %q, want 'y'", innerAssignTarget.Value)
				}
			},
		},
		{
			name:  "while with function call in body",
			input: "while hasMoreData() do processItem();",
			assertions: func(t *testing.T, stmt *ast.WhileStatement) {
				// Test condition is a function call
				condCall, ok := stmt.Condition.(*ast.CallExpression)
				if !ok {
					t.Fatalf("condition is not CallExpression. got=%T", stmt.Condition)
				}

				if !testIdentifier(t, condCall.Function, "hasMoreData") {
					return
				}

				// Test body is also a function call
				bodyExpr, ok := stmt.Body.(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("body is not ExpressionStatement. got=%T", stmt.Body)
				}

				bodyCall, ok := bodyExpr.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("body expression is not CallExpression. got=%T", bodyExpr.Expression)
				}

				if !testIdentifier(t, bodyCall.Function, "processItem") {
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

			stmt, ok := program.Statements[0].(*ast.WhileStatement)
			if !ok {
				t.Fatalf("statement is not ast.WhileStatement. got=%T", program.Statements[0])
			}

			tt.assertions(t, stmt)
		})
	}
}

// TestRepeatStatements tests parsing of repeat-until loop statements.
func TestRepeatStatements(t *testing.T) {
	tests := []struct {
		assertions func(*testing.T, *ast.RepeatStatement)
		name       string
		input      string
	}{
		{
			name:  "simple repeat loop",
			input: "repeat x := x + 1 until x >= 10;",
			assertions: func(t *testing.T, stmt *ast.RepeatStatement) {
				// Test body: x := x + 1
				assign, ok := stmt.Body.(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("body is not AssignmentStatement. got=%T", stmt.Body)
				}

				assignTarget, ok := assign.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("assign.Target is not *ast.Identifier. got=%T", assign.Target)
				}
				if assignTarget.Value != "x" {
					t.Errorf("assignment name = %q, want 'x'", assignTarget.Value)
				}

				if !testInfixExpression(t, assign.Value, "x", "+", 1) {
					return
				}

				// Test condition: x >= 10
				if !testInfixExpression(t, stmt.Condition, "x", ">=", 10) {
					return
				}
			},
		},
		{
			name: "repeat with block body",
			input: `repeat begin
  x := x + 1;
  PrintLn(x);
end until x >= 10;`,
			assertions: func(t *testing.T, stmt *ast.RepeatStatement) {
				// Test body is a block
				block, ok := stmt.Body.(*ast.BlockStatement)
				if !ok {
					t.Fatalf("body is not BlockStatement. got=%T", stmt.Body)
				}

				if len(block.Statements) != 2 {
					t.Fatalf("block has %d statements, want 2", len(block.Statements))
				}

				// First statement: x := x + 1;
				assign, ok := block.Statements[0].(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("first block statement is not AssignmentStatement. got=%T", block.Statements[0])
				}

				assignTarget, ok := assign.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("assign.Target is not *ast.Identifier. got=%T", assign.Target)
				}
				if assignTarget.Value != "x" {
					t.Errorf("assignment name = %q, want 'x'", assignTarget.Value)
				}

				// Second statement: PrintLn(x);
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

				// Test condition: x >= 10
				if !testInfixExpression(t, stmt.Condition, "x", ">=", 10) {
					return
				}
			},
		},
		{
			name:  "repeat with complex condition",
			input: "repeat x := x * 2 until (x > 100) or (x < 0);",
			assertions: func(t *testing.T, stmt *ast.RepeatStatement) {
				// Test body
				assign, ok := stmt.Body.(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("body is not AssignmentStatement. got=%T", stmt.Body)
				}

				if !testInfixExpression(t, assign.Value, "x", "*", 2) {
					return
				}

				// Test that condition is a binary expression with 'or'
				binExp, ok := stmt.Condition.(*ast.BinaryExpression)
				if !ok {
					t.Fatalf("condition is not BinaryExpression. got=%T", stmt.Condition)
				}

				if binExp.Operator != "or" {
					t.Errorf("condition operator = %q, want 'or'", binExp.Operator)
				}

				// Test left side: x > 100
				if !testInfixExpression(t, binExp.Left, "x", ">", 100) {
					return
				}

				// Test right side: x < 0
				if !testInfixExpression(t, binExp.Right, "x", "<", 0) {
					return
				}
			},
		},
		{
			name: "nested repeat loops",
			input: `repeat
  repeat
    y := y + 1
  until y >= 5
until x >= 10;`,
			assertions: func(t *testing.T, stmt *ast.RepeatStatement) {
				// Test that body is another repeat statement
				innerRepeat, ok := stmt.Body.(*ast.RepeatStatement)
				if !ok {
					t.Fatalf("body is not RepeatStatement. got=%T", stmt.Body)
				}

				// Test inner body
				innerAssign, ok := innerRepeat.Body.(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("inner repeat body is not AssignmentStatement. got=%T", innerRepeat.Body)
				}

				innerAssignTarget, ok := innerAssign.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("innerAssign.Target is not *ast.Identifier. got=%T", innerAssign.Target)
				}
				if innerAssignTarget.Value != "y" {
					t.Errorf("inner assignment name = %q, want 'y'", innerAssignTarget.Value)
				}

				// Test inner condition: y >= 5
				if !testInfixExpression(t, innerRepeat.Condition, "y", ">=", 5) {
					return
				}

				// Test outer condition: x >= 10
				if !testInfixExpression(t, stmt.Condition, "x", ">=", 10) {
					return
				}
			},
		},
		{
			name:  "repeat with function call in body",
			input: "repeat processItem() until isDone();",
			assertions: func(t *testing.T, stmt *ast.RepeatStatement) {
				// Test body is a function call
				bodyExpr, ok := stmt.Body.(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("body is not ExpressionStatement. got=%T", stmt.Body)
				}

				bodyCall, ok := bodyExpr.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("body expression is not CallExpression. got=%T", bodyExpr.Expression)
				}

				if !testIdentifier(t, bodyCall.Function, "processItem") {
					return
				}

				// Test condition is a function call
				condCall, ok := stmt.Condition.(*ast.CallExpression)
				if !ok {
					t.Fatalf("condition is not CallExpression. got=%T", stmt.Condition)
				}

				if !testIdentifier(t, condCall.Function, "isDone") {
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

			stmt, ok := program.Statements[0].(*ast.RepeatStatement)
			if !ok {
				t.Fatalf("statement is not ast.RepeatStatement. got=%T", program.Statements[0])
			}

			tt.assertions(t, stmt)
		})
	}
}

// TestForStatements tests parsing of for loop statements.
func TestForStatements(t *testing.T) {
	tests := []struct {
		assertions func(*testing.T, *ast.ForStatement)
		name       string
		input      string
	}{
		{
			name:  "simple ascending for loop",
			input: "for i := 1 to 10 do PrintLn(i);",
			assertions: func(t *testing.T, stmt *ast.ForStatement) {
				// Test loop variable
				if stmt.Variable.Value != "i" {
					t.Errorf("loop variable = %q, want 'i'", stmt.Variable.Value)
				}

				// Test start expression: 1
				if !testIntegerLiteral(t, stmt.Start, 1) {
					return
				}

				// Test end expression: 10
				if !testIntegerLiteral(t, stmt.EndValue, 10) {
					return
				}

				// Test direction: to
				if stmt.Direction != ast.ForTo {
					t.Errorf("direction = %v, want ForTo", stmt.Direction)
				}

				// Test body: PrintLn(i)
				bodyExpr, ok := stmt.Body.(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("body is not ExpressionStatement. got=%T", stmt.Body)
				}

				call, ok := bodyExpr.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("body expression is not CallExpression. got=%T", bodyExpr.Expression)
				}

				if !testIdentifier(t, call.Function, "PrintLn") {
					return
				}

				if len(call.Arguments) != 1 {
					t.Fatalf("call has %d arguments, want 1", len(call.Arguments))
				}

				if !testIdentifier(t, call.Arguments[0], "i") {
					return
				}
			},
		},
		{
			name:  "simple descending for loop",
			input: "for i := 10 downto 1 do PrintLn(i);",
			assertions: func(t *testing.T, stmt *ast.ForStatement) {
				// Test loop variable
				if stmt.Variable.Value != "i" {
					t.Errorf("loop variable = %q, want 'i'", stmt.Variable.Value)
				}

				// Test start expression: 10
				if !testIntegerLiteral(t, stmt.Start, 10) {
					return
				}

				// Test end expression: 1
				if !testIntegerLiteral(t, stmt.EndValue, 1) {
					return
				}

				// Test direction: downto
				if stmt.Direction != ast.ForDownto {
					t.Errorf("direction = %v, want ForDownto", stmt.Direction)
				}

				// Test body is a PrintLn call
				bodyExpr, ok := stmt.Body.(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("body is not ExpressionStatement. got=%T", stmt.Body)
				}

				call, ok := bodyExpr.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("body expression is not CallExpression. got=%T", bodyExpr.Expression)
				}

				if !testIdentifier(t, call.Function, "PrintLn") {
					return
				}
			},
		},
		{
			name:  "for loop with inline var declaration",
			input: "for var i := 0 to 10 do PrintLn(i);",
			assertions: func(t *testing.T, stmt *ast.ForStatement) {
				if !stmt.InlineVar {
					t.Fatalf("expected inline var flag set")
				}

				if stmt.Variable.Value != "i" {
					t.Errorf("loop variable = %q, want 'i'", stmt.Variable.Value)
				}

				if !testIntegerLiteral(t, stmt.Start, 0) {
					return
				}

				if !testIntegerLiteral(t, stmt.EndValue, 10) {
					return
				}

				if stmt.Direction != ast.ForTo {
					t.Errorf("direction = %v, want ForTo", stmt.Direction)
				}

				bodyExpr, ok := stmt.Body.(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("body is not ExpressionStatement. got=%T", stmt.Body)
				}

				call, ok := bodyExpr.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("body expression is not CallExpression. got=%T", bodyExpr.Expression)
				}

				if !testIdentifier(t, call.Function, "PrintLn") {
					return
				}

				if len(call.Arguments) != 1 {
					t.Fatalf("call has %d arguments, want 1", len(call.Arguments))
				}

				if !testIdentifier(t, call.Arguments[0], "i") {
					return
				}
			},
		},
		{
			name: "for loop with block body",
			input: `for i := 1 to 5 do begin
  PrintLn(i);
  PrintLn(i * 2);
end;`,
			assertions: func(t *testing.T, stmt *ast.ForStatement) {
				// Test loop variable
				if stmt.Variable.Value != "i" {
					t.Errorf("loop variable = %q, want 'i'", stmt.Variable.Value)
				}

				// Test direction
				if stmt.Direction != ast.ForTo {
					t.Errorf("direction = %v, want ForTo", stmt.Direction)
				}

				// Test body is a block
				block, ok := stmt.Body.(*ast.BlockStatement)
				if !ok {
					t.Fatalf("body is not BlockStatement. got=%T", stmt.Body)
				}

				if len(block.Statements) != 2 {
					t.Fatalf("block has %d statements, want 2", len(block.Statements))
				}

				// First statement: PrintLn(i)
				exprStmt1, ok := block.Statements[0].(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("first block statement is not ExpressionStatement. got=%T", block.Statements[0])
				}

				call1, ok := exprStmt1.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("first block statement expression is not CallExpression. got=%T", exprStmt1.Expression)
				}

				if !testIdentifier(t, call1.Function, "PrintLn") {
					return
				}

				// Second statement: PrintLn(i * 2)
				exprStmt2, ok := block.Statements[1].(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("second block statement is not ExpressionStatement. got=%T", block.Statements[1])
				}

				call2, ok := exprStmt2.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("second block statement expression is not CallExpression. got=%T", exprStmt2.Expression)
				}

				if !testIdentifier(t, call2.Function, "PrintLn") {
					return
				}
			},
		},
		{
			name:  "for loop with variable expressions",
			input: "for i := start to finish do sum := sum + i;",
			assertions: func(t *testing.T, stmt *ast.ForStatement) {
				// Test loop variable
				if stmt.Variable.Value != "i" {
					t.Errorf("loop variable = %q, want 'i'", stmt.Variable.Value)
				}

				// Test start expression is an identifier
				if !testIdentifier(t, stmt.Start, "start") {
					return
				}

				// Test end expression is an identifier (changed from 'end' to 'finish' to avoid keyword conflict)
				if !testIdentifier(t, stmt.EndValue, "finish") {
					return
				}

				// Test direction
				if stmt.Direction != ast.ForTo {
					t.Errorf("direction = %v, want ForTo", stmt.Direction)
				}

				// Test body is an assignment
				assign, ok := stmt.Body.(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("body is not AssignmentStatement. got=%T", stmt.Body)
				}

				assignTarget, ok := assign.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("assign.Target is not *ast.Identifier. got=%T", assign.Target)
				}
				if assignTarget.Value != "sum" {
					t.Errorf("assignment name = %q, want 'sum'", assignTarget.Value)
				}

				// Test assignment value: sum + i
				if !testInfixExpression(t, assign.Value, "sum", "+", "i") {
					return
				}
			},
		},
		{
			name:  "for loop with expression boundaries",
			input: "for i := (start + 1) to (finish - 1) do process(i);",
			assertions: func(t *testing.T, stmt *ast.ForStatement) {
				// Test loop variable
				if stmt.Variable.Value != "i" {
					t.Errorf("loop variable = %q, want 'i'", stmt.Variable.Value)
				}

				// Test start expression is a binary expression: start + 1
				if !testInfixExpression(t, stmt.Start, "start", "+", 1) {
					return
				}

				// Test end expression is a binary expression: finish - 1
				if !testInfixExpression(t, stmt.EndValue, "finish", "-", 1) {
					return
				}

				// Test direction
				if stmt.Direction != ast.ForTo {
					t.Errorf("direction = %v, want ForTo", stmt.Direction)
				}
			},
		},
		{
			name: "nested for loops",
			input: `for i := 1 to 10 do
  for j := 1 to 10 do
    PrintLn(i * j);`,
			assertions: func(t *testing.T, stmt *ast.ForStatement) {
				// Test outer loop variable
				if stmt.Variable.Value != "i" {
					t.Errorf("outer loop variable = %q, want 'i'", stmt.Variable.Value)
				}

				// Test outer loop direction
				if stmt.Direction != ast.ForTo {
					t.Errorf("outer direction = %v, want ForTo", stmt.Direction)
				}

				// Test that body is another for statement
				innerFor, ok := stmt.Body.(*ast.ForStatement)
				if !ok {
					t.Fatalf("body is not ForStatement. got=%T", stmt.Body)
				}

				// Test inner loop variable
				if innerFor.Variable.Value != "j" {
					t.Errorf("inner loop variable = %q, want 'j'", innerFor.Variable.Value)
				}

				// Test inner loop direction
				if innerFor.Direction != ast.ForTo {
					t.Errorf("inner direction = %v, want ForTo", innerFor.Direction)
				}

				// Test inner body is a PrintLn call
				innerBodyExpr, ok := innerFor.Body.(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("inner body is not ExpressionStatement. got=%T", innerFor.Body)
				}

				call, ok := innerBodyExpr.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("inner body expression is not CallExpression. got=%T", innerBodyExpr.Expression)
				}

				if !testIdentifier(t, call.Function, "PrintLn") {
					return
				}
			},
		},
		{
			name:  "for loop with assignment in body",
			input: "for i := 0 to 100 do x := x + i;",
			assertions: func(t *testing.T, stmt *ast.ForStatement) {
				// Test loop variable
				if stmt.Variable.Value != "i" {
					t.Errorf("loop variable = %q, want 'i'", stmt.Variable.Value)
				}

				// Test start: 0
				if !testIntegerLiteral(t, stmt.Start, 0) {
					return
				}

				// Test end: 100
				if !testIntegerLiteral(t, stmt.EndValue, 100) {
					return
				}

				// Test direction
				if stmt.Direction != ast.ForTo {
					t.Errorf("direction = %v, want ForTo", stmt.Direction)
				}

				// Test body: x := x + i
				assign, ok := stmt.Body.(*ast.AssignmentStatement)
				if !ok {
					t.Fatalf("body is not AssignmentStatement. got=%T", stmt.Body)
				}

				assignTarget, ok := assign.Target.(*ast.Identifier)
				if !ok {
					t.Fatalf("assign.Target is not *ast.Identifier. got=%T", assign.Target)
				}
				if assignTarget.Value != "x" {
					t.Errorf("assignment name = %q, want 'x'", assignTarget.Value)
				}

				if !testInfixExpression(t, assign.Value, "x", "+", "i") {
					return
				}
			},
		},
		{
			name:  "for loop downto with larger numbers",
			input: "for count := 100 downto 0 do PrintLn(count);",
			assertions: func(t *testing.T, stmt *ast.ForStatement) {
				// Test loop variable
				if stmt.Variable.Value != "count" {
					t.Errorf("loop variable = %q, want 'count'", stmt.Variable.Value)
				}

				// Test start: 100
				if !testIntegerLiteral(t, stmt.Start, 100) {
					return
				}

				// Test end: 0
				if !testIntegerLiteral(t, stmt.EndValue, 0) {
					return
				}

				// Test direction: downto
				if stmt.Direction != ast.ForDownto {
					t.Errorf("direction = %v, want ForDownto", stmt.Direction)
				}
			},
		},
		{
			name:  "for loop ascending with step",
			input: "for i := 1 to 10 step 2 do PrintLn(i);",
			assertions: func(t *testing.T, stmt *ast.ForStatement) {
				// Test loop variable
				if stmt.Variable.Value != "i" {
					t.Errorf("loop variable = %q, want 'i'", stmt.Variable.Value)
				}

				// Test start expression: 1
				if !testIntegerLiteral(t, stmt.Start, 1) {
					return
				}

				// Test end expression: 10
				if !testIntegerLiteral(t, stmt.EndValue, 10) {
					return
				}

				// Test direction: to
				if stmt.Direction != ast.ForTo {
					t.Errorf("direction = %v, want ForTo", stmt.Direction)
				}

				// Test step expression: 2
				if stmt.Step == nil {
					t.Fatal("step should not be nil")
				}
				if !testIntegerLiteral(t, stmt.Step, 2) {
					return
				}

				// Test body: PrintLn(i)
				bodyExpr, ok := stmt.Body.(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("body is not ExpressionStatement. got=%T", stmt.Body)
				}

				call, ok := bodyExpr.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("body expression is not CallExpression. got=%T", bodyExpr.Expression)
				}

				if !testIdentifier(t, call.Function, "PrintLn") {
					return
				}
			},
		},
		{
			name:  "for loop descending with step",
			input: "for i := 10 downto 1 step 3 do PrintLn(i);",
			assertions: func(t *testing.T, stmt *ast.ForStatement) {
				// Test loop variable
				if stmt.Variable.Value != "i" {
					t.Errorf("loop variable = %q, want 'i'", stmt.Variable.Value)
				}

				// Test start expression: 10
				if !testIntegerLiteral(t, stmt.Start, 10) {
					return
				}

				// Test end expression: 1
				if !testIntegerLiteral(t, stmt.EndValue, 1) {
					return
				}

				// Test direction: downto
				if stmt.Direction != ast.ForDownto {
					t.Errorf("direction = %v, want ForDownto", stmt.Direction)
				}

				// Test step expression: 3
				if stmt.Step == nil {
					t.Fatal("step should not be nil")
				}
				if !testIntegerLiteral(t, stmt.Step, 3) {
					return
				}

				// Test body: PrintLn(i)
				bodyExpr, ok := stmt.Body.(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("body is not ExpressionStatement. got=%T", stmt.Body)
				}

				call, ok := bodyExpr.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("body expression is not CallExpression. got=%T", bodyExpr.Expression)
				}

				if !testIdentifier(t, call.Function, "PrintLn") {
					return
				}
			},
		},
		{
			name:  "for loop with step expression",
			input: "for i := 1 to 100 step (x * 2) do PrintLn(i);",
			assertions: func(t *testing.T, stmt *ast.ForStatement) {
				// Test loop variable
				if stmt.Variable.Value != "i" {
					t.Errorf("loop variable = %q, want 'i'", stmt.Variable.Value)
				}

				// Test start expression: 1
				if !testIntegerLiteral(t, stmt.Start, 1) {
					return
				}

				// Test end expression: 100
				if !testIntegerLiteral(t, stmt.EndValue, 100) {
					return
				}

				// Test direction: to
				if stmt.Direction != ast.ForTo {
					t.Errorf("direction = %v, want ForTo", stmt.Direction)
				}

				// Test step expression: x * 2
				if stmt.Step == nil {
					t.Fatal("step should not be nil")
				}
				if !testInfixExpression(t, stmt.Step, "x", "*", 2) {
					return
				}
			},
		},
		{
			name:  "for loop with inline var and step",
			input: "for var i := 0 to 20 step 5 do PrintLn(i);",
			assertions: func(t *testing.T, stmt *ast.ForStatement) {
				// Test inline var flag
				if !stmt.InlineVar {
					t.Fatal("expected inline var flag set")
				}

				// Test loop variable
				if stmt.Variable.Value != "i" {
					t.Errorf("loop variable = %q, want 'i'", stmt.Variable.Value)
				}

				// Test start expression: 0
				if !testIntegerLiteral(t, stmt.Start, 0) {
					return
				}

				// Test end expression: 20
				if !testIntegerLiteral(t, stmt.EndValue, 20) {
					return
				}

				// Test direction: to
				if stmt.Direction != ast.ForTo {
					t.Errorf("direction = %v, want ForTo", stmt.Direction)
				}

				// Test step expression: 5
				if stmt.Step == nil {
					t.Fatal("step should not be nil")
				}
				if !testIntegerLiteral(t, stmt.Step, 5) {
					return
				}
			},
		},
		{
			name:  "for loop without step still works (backward compatibility)",
			input: "for i := 1 to 10 do PrintLn(i);",
			assertions: func(t *testing.T, stmt *ast.ForStatement) {
				// Test that step is nil when not specified
				if stmt.Step != nil {
					t.Errorf("step should be nil when not specified, got %v", stmt.Step)
				}

				// Test loop variable
				if stmt.Variable.Value != "i" {
					t.Errorf("loop variable = %q, want 'i'", stmt.Variable.Value)
				}

				// Test direction
				if stmt.Direction != ast.ForTo {
					t.Errorf("direction = %v, want ForTo", stmt.Direction)
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

			stmt, ok := program.Statements[0].(*ast.ForStatement)
			if !ok {
				t.Fatalf("statement is not ast.ForStatement. got=%T", program.Statements[0])
			}

			tt.assertions(t, stmt)
		})
	}
}

// TestForStatementStepErrors tests error handling for for loops with step keyword.
func TestForStatementStepErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "missing step expression after step keyword",
			input: "for i := 1 to 10 step do PrintLn(i);",
		},
		{
			name:  "missing do after step",
			input: "for i := 1 to 10 step 2 PrintLn(i);",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testParser(tt.input)
			p.ParseProgram()

			if len(p.errors) == 0 {
				t.Errorf("expected parser error for %q, got none", tt.name)
			}
		})
	}
}

// TestForInStatements tests parsing of for-in loop statements.
func TestForInStatements(t *testing.T) {
	tests := []struct {
		assertions func(*testing.T, *ast.ForInStatement)
		name       string
		input      string
	}{
		{
			name:  "simple for-in loop with identifier collection",
			input: "for e in mySet do PrintLn(e);",
			assertions: func(t *testing.T, stmt *ast.ForInStatement) {
				// Test loop variable
				if stmt.Variable.Value != "e" {
					t.Errorf("loop variable = %q, want 'e'", stmt.Variable.Value)
				}

				// Test InlineVar is false
				if stmt.InlineVar {
					t.Errorf("InlineVar = true, want false")
				}

				// Test collection expression is an identifier
				if !testIdentifier(t, stmt.Collection, "mySet") {
					return
				}

				// Test body: PrintLn(e)
				bodyExpr, ok := stmt.Body.(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("body is not ExpressionStatement. got=%T", stmt.Body)
				}

				call, ok := bodyExpr.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("body expression is not CallExpression. got=%T", bodyExpr.Expression)
				}

				if !testIdentifier(t, call.Function, "PrintLn") {
					return
				}

				if len(call.Arguments) != 1 {
					t.Fatalf("call has %d arguments, want 1", len(call.Arguments))
				}

				if !testIdentifier(t, call.Arguments[0], "e") {
					return
				}
			},
		},
		{
			name:  "for-in loop with inline var declaration",
			input: "for var item in myArray do Process(item);",
			assertions: func(t *testing.T, stmt *ast.ForInStatement) {
				// Test InlineVar is true
				if !stmt.InlineVar {
					t.Fatalf("expected inline var flag set")
				}

				// Test loop variable
				if stmt.Variable.Value != "item" {
					t.Errorf("loop variable = %q, want 'item'", stmt.Variable.Value)
				}

				// Test collection expression
				if !testIdentifier(t, stmt.Collection, "myArray") {
					return
				}

				// Test body: Process(item)
				bodyExpr, ok := stmt.Body.(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("body is not ExpressionStatement. got=%T", stmt.Body)
				}

				call, ok := bodyExpr.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("body expression is not CallExpression. got=%T", bodyExpr.Expression)
				}

				if !testIdentifier(t, call.Function, "Process") {
					return
				}

				if len(call.Arguments) != 1 {
					t.Fatalf("call has %d arguments, want 1", len(call.Arguments))
				}

				if !testIdentifier(t, call.Arguments[0], "item") {
					return
				}
			},
		},
		{
			name:  "for-in loop with string literal",
			input: `for ch in "hello" do Print(ch);`,
			assertions: func(t *testing.T, stmt *ast.ForInStatement) {
				// Test loop variable
				if stmt.Variable.Value != "ch" {
					t.Errorf("loop variable = %q, want 'ch'", stmt.Variable.Value)
				}

				// Test collection expression is a string literal
				strLit, ok := stmt.Collection.(*ast.StringLiteral)
				if !ok {
					t.Fatalf("collection is not StringLiteral. got=%T", stmt.Collection)
				}

				if strLit.Value != "hello" {
					t.Errorf("string literal value = %q, want 'hello'", strLit.Value)
				}

				// Test body: Print(ch)
				bodyExpr, ok := stmt.Body.(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("body is not ExpressionStatement. got=%T", stmt.Body)
				}

				call, ok := bodyExpr.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("body expression is not CallExpression. got=%T", bodyExpr.Expression)
				}

				if !testIdentifier(t, call.Function, "Print") {
					return
				}
			},
		},
		{
			name: "for-in loop with block body",
			input: `for var x in collection do begin
  PrintLn(x);
  Process(x);
end;`,
			assertions: func(t *testing.T, stmt *ast.ForInStatement) {
				// Test InlineVar is true
				if !stmt.InlineVar {
					t.Fatalf("expected inline var flag set")
				}

				// Test loop variable
				if stmt.Variable.Value != "x" {
					t.Errorf("loop variable = %q, want 'x'", stmt.Variable.Value)
				}

				// Test collection
				if !testIdentifier(t, stmt.Collection, "collection") {
					return
				}

				// Test body is a block statement
				block, ok := stmt.Body.(*ast.BlockStatement)
				if !ok {
					t.Fatalf("body is not BlockStatement. got=%T", stmt.Body)
				}

				if len(block.Statements) != 2 {
					t.Fatalf("block has %d statements, want 2", len(block.Statements))
				}

				// Test first statement: PrintLn(x)
				exprStmt1, ok := block.Statements[0].(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("first block statement is not ExpressionStatement. got=%T", block.Statements[0])
				}

				call1, ok := exprStmt1.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("first block statement expression is not CallExpression. got=%T", exprStmt1.Expression)
				}

				if !testIdentifier(t, call1.Function, "PrintLn") {
					return
				}

				// Test second statement: Process(x)
				exprStmt2, ok := block.Statements[1].(*ast.ExpressionStatement)
				if !ok {
					t.Fatalf("second block statement is not ExpressionStatement. got=%T", block.Statements[1])
				}

				call2, ok := exprStmt2.Expression.(*ast.CallExpression)
				if !ok {
					t.Fatalf("second block statement expression is not CallExpression. got=%T", exprStmt2.Expression)
				}

				if !testIdentifier(t, call2.Function, "Process") {
					return
				}
			},
		},
		{
			name:  "for-in loop with set literal",
			input: "for day in [Mon, Tue, Wed] do Process(day);",
			assertions: func(t *testing.T, stmt *ast.ForInStatement) {
				if stmt.Variable.Value != "day" {
					t.Errorf("loop variable = %q, want 'day'", stmt.Variable.Value)
				}

				setLit, ok := stmt.Collection.(*ast.SetLiteral)
				if !ok {
					t.Fatalf("collection is not SetLiteral. got=%T", stmt.Collection)
				}

				if len(setLit.Elements) != 3 {
					t.Fatalf("set has %d elements, want 3", len(setLit.Elements))
				}

				expected := []string{"Mon", "Tue", "Wed"}
				for i, want := range expected {
					ident, ok := setLit.Elements[i].(*ast.Identifier)
					if !ok {
						t.Fatalf("element %d is not Identifier. got=%T", i, setLit.Elements[i])
					}
					if ident.Value != want {
						t.Fatalf("element %d value = %q, want %q", i, ident.Value, want)
					}
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

			stmt, ok := program.Statements[0].(*ast.ForInStatement)
			if !ok {
				t.Fatalf("statement is not ast.ForInStatement. got=%T", program.Statements[0])
			}

			tt.assertions(t, stmt)
		})
	}
}
