package passes

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

func TestValidationPass_TypeMismatchInAssignment(t *testing.T) {
	// Create a program with type mismatch in assignment
	// var x: Integer; x := "hello";
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				Names: []*ast.Identifier{{Value: "x"}},
				Type:  &ast.TypeAnnotation{Name: "Integer"},
				BaseNode: ast.BaseNode{
					Token: token.Token{Pos: token.Position{Line: 1, Column: 1}},
				},
			},
			&ast.AssignmentStatement{
				Target: &ast.Identifier{Value: "x"},
				Value:  &ast.StringLiteral{Value: "hello"},
				BaseNode: ast.BaseNode{
					Token: token.Token{Pos: token.Position{Line: 2, Column: 1}},
				},
			},
		},
	}

	// Run all passes
	ctx := NewPassContext()
	pass1 := NewDeclarationPass()
	pass2 := NewTypeResolutionPass()
	pass3 := NewValidationPass()

	_ = pass1.Run(program, ctx)
	_ = pass2.Run(program, ctx)
	_ = pass3.Run(program, ctx)

	// Verify type mismatch error was detected
	if len(ctx.Errors) == 0 {
		t.Fatal("Expected type mismatch error, got none")
	}

	foundMismatch := false
	for _, errMsg := range ctx.Errors {
		if contains(errMsg, "cannot assign") || contains(errMsg, "String") {
			foundMismatch = true
			break
		}
	}

	if !foundMismatch {
		t.Errorf("Expected type mismatch error, got: %v", ctx.Errors)
	}
}

func TestValidationPass_BinaryExpressionTypeCheck(t *testing.T) {
	// Create a program with arithmetic on integers
	// var result: Integer; result := 5 + 3;
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				Names: []*ast.Identifier{{Value: "result"}},
				Type:  &ast.TypeAnnotation{Name: "Integer"},
				BaseNode: ast.BaseNode{
					Token: token.Token{Pos: token.Position{Line: 1, Column: 1}},
				},
			},
			&ast.AssignmentStatement{
				Target: &ast.Identifier{Value: "result"},
				Value: &ast.BinaryExpression{
					Left:     &ast.IntegerLiteral{Value: 5},
					Operator: "+",
					Right:    &ast.IntegerLiteral{Value: 3},
				},
				BaseNode: ast.BaseNode{
					Token: token.Token{Pos: token.Position{Line: 2, Column: 1}},
				},
			},
		},
	}

	// Run all passes
	ctx := NewPassContext()
	pass1 := NewDeclarationPass()
	pass2 := NewTypeResolutionPass()
	pass3 := NewValidationPass()

	_ = pass1.Run(program, ctx)
	_ = pass2.Run(program, ctx)
	_ = pass3.Run(program, ctx)

	// Should have no errors - this is valid
	if len(ctx.Errors) > 0 {
		t.Errorf("Expected no errors, got: %v", ctx.Errors)
	}
}

func TestValidationPass_UndefinedVariable(t *testing.T) {
	// Create a program referencing an undefined variable
	// var x: Integer; x := y;
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				Names: []*ast.Identifier{{Value: "x"}},
				Type:  &ast.TypeAnnotation{Name: "Integer"},
				BaseNode: ast.BaseNode{
					Token: token.Token{Pos: token.Position{Line: 1, Column: 1}},
				},
			},
			&ast.AssignmentStatement{
				Target: &ast.Identifier{Value: "x"},
				Value:  &ast.Identifier{Value: "y"}, // undefined
				BaseNode: ast.BaseNode{
					Token: token.Token{Pos: token.Position{Line: 2, Column: 1}},
				},
			},
		},
	}

	// Run all passes
	ctx := NewPassContext()
	pass1 := NewDeclarationPass()
	pass2 := NewTypeResolutionPass()
	pass3 := NewValidationPass()

	_ = pass1.Run(program, ctx)
	_ = pass2.Run(program, ctx)
	_ = pass3.Run(program, ctx)

	// Verify undefined variable error was detected
	if len(ctx.Errors) == 0 {
		t.Fatal("Expected undefined variable error, got none")
	}

	foundUndefined := false
	for _, errMsg := range ctx.Errors {
		if contains(errMsg, "undefined") && contains(errMsg, "y") {
			foundUndefined = true
			break
		}
	}

	if !foundUndefined {
		t.Errorf("Expected undefined variable error, got: %v", ctx.Errors)
	}
}

func TestValidationPass_BreakOutsideLoop(t *testing.T) {
	// Create a program with break outside loop
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.BreakStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Pos: token.Position{Line: 1, Column: 1}},
				},
			},
		},
	}

	// Run validation pass
	ctx := NewPassContext()
	pass3 := NewValidationPass()
	_ = pass3.Run(program, ctx)

	// Verify break outside loop error
	if len(ctx.Errors) == 0 {
		t.Fatal("Expected break outside loop error, got none")
	}

	foundBreakError := false
	for _, errMsg := range ctx.Errors {
		if contains(errMsg, "break") && contains(errMsg, "outside") {
			foundBreakError = true
			break
		}
	}

	if !foundBreakError {
		t.Errorf("Expected break outside loop error, got: %v", ctx.Errors)
	}
}

func TestValidationPass_InvalidArithmeticOperands(t *testing.T) {
	// Create a program trying to add string + integer
	// var x: String; x := "hello" + 5;
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				Names: []*ast.Identifier{{Value: "x"}},
				Type:  &ast.TypeAnnotation{Name: "String"},
				BaseNode: ast.BaseNode{
					Token: token.Token{Pos: token.Position{Line: 1, Column: 1}},
				},
			},
			&ast.AssignmentStatement{
				Target: &ast.Identifier{Value: "x"},
				Value: &ast.BinaryExpression{
					Left:     &ast.StringLiteral{Value: "hello"},
					Operator: "+",
					Right:    &ast.IntegerLiteral{Value: 5},
				},
				BaseNode: ast.BaseNode{
					Token: token.Token{Pos: token.Position{Line: 2, Column: 1}},
				},
			},
		},
	}

	// Run all passes
	ctx := NewPassContext()
	pass1 := NewDeclarationPass()
	pass2 := NewTypeResolutionPass()
	pass3 := NewValidationPass()

	_ = pass1.Run(program, ctx)
	_ = pass2.Run(program, ctx)
	_ = pass3.Run(program, ctx)

	// Verify operator type error was detected
	if len(ctx.Errors) == 0 {
		t.Fatal("Expected operator type error, got none")
	}

	foundOperatorError := false
	for _, errMsg := range ctx.Errors {
		if contains(errMsg, "operator") || contains(errMsg, "numeric") {
			foundOperatorError = true
			break
		}
	}

	if !foundOperatorError {
		t.Errorf("Expected operator type error, got: %v", ctx.Errors)
	}
}

func TestValidationPass_BooleanCondition(t *testing.T) {
	// Create a program with non-boolean condition in if statement
	// if 5 then ...
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.IfStatement{
				Condition: &ast.IntegerLiteral{Value: 5}, // Should be boolean
				Consequence: &ast.BlockStatement{
					Statements: []ast.Statement{},
				},
				BaseNode: ast.BaseNode{
					Token: token.Token{Pos: token.Position{Line: 1, Column: 1}},
				},
			},
		},
	}

	// Run validation pass
	ctx := NewPassContext()
	pass2 := NewTypeResolutionPass()
	pass3 := NewValidationPass()

	_ = pass2.Run(program, ctx)
	_ = pass3.Run(program, ctx)

	// Verify boolean condition error
	if len(ctx.Errors) == 0 {
		t.Fatal("Expected boolean condition error, got none")
	}

	foundBooleanError := false
	for _, errMsg := range ctx.Errors {
		if contains(errMsg, "boolean") || contains(errMsg, "condition") {
			foundBooleanError = true
			break
		}
	}

	if !foundBooleanError {
		t.Errorf("Expected boolean condition error, got: %v", ctx.Errors)
	}
}
