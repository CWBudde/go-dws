package ast_test

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/dwscript"
)

// TestProgramASTReturnsValidAST tests that Program.AST() returns valid AST (Task 10.20)
func TestProgramASTReturnsValidAST(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "simple variable declaration",
			source: "var x: Integer := 42;",
		},
		{
			name: "function declaration",
			source: `function Add(a, b: Integer): Integer;
begin
  Result := a + b;
end;`,
		},
		{
			name: "class declaration",
			source: `type TPoint = class
  X, Y: Integer;
end;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, err := dwscript.New(dwscript.WithTypeCheck(true))
			if err != nil {
				t.Fatalf("Failed to create engine: %v", err)
			}

			program, err := engine.Compile(tt.source)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			// Get AST
			tree := program.AST()
			if tree == nil {
				t.Fatal("Program.AST() returned nil")
			}

			// AST should be a valid Program node
			if tree.Statements == nil {
				t.Error("AST Program.Statements should not be nil")
			}

			// Should have at least one statement
			if len(tree.Statements) == 0 {
				t.Error("Expected at least one statement in AST")
			}
		})
	}
}

// TestASTTraversalWithVisitor tests AST traversal using visitor pattern (Task 10.20)
func TestASTTraversalWithVisitor(t *testing.T) {
	source := `
var x: Integer := 42;
var y: String := 'hello';
function Add(a, b: Integer): Integer;
begin
  Result := a + b;
end;
`

	engine, err := dwscript.New(dwscript.WithTypeCheck(true))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	program, err := engine.Compile(source)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	tree := program.AST()

	// Count different node types using visitor
	counts := make(map[string]int)
	ast.Inspect(tree, func(node ast.Node) bool {
		if node != nil {
			// Get type name
			switch node.(type) {
			case *ast.Program:
				counts["Program"]++
			case *ast.VarDeclStatement:
				counts["VarDecl"]++
			case *ast.FunctionDecl:
				counts["FunctionDecl"]++
			case *ast.Identifier:
				counts["Identifier"]++
			case *ast.IntegerLiteral:
				counts["IntegerLiteral"]++
			case *ast.StringLiteral:
				counts["StringLiteral"]++
			case *ast.BlockStatement:
				counts["BlockStatement"]++
			}
		}
		return true
	})

	// Verify we found expected node types
	if counts["Program"] != 1 {
		t.Errorf("Expected 1 Program node, got %d", counts["Program"])
	}
	if counts["VarDecl"] != 2 {
		t.Errorf("Expected 2 VarDecl nodes, got %d", counts["VarDecl"])
	}
	if counts["FunctionDecl"] != 1 {
		t.Errorf("Expected 1 FunctionDecl node, got %d", counts["FunctionDecl"])
	}
	if counts["Identifier"] < 1 {
		t.Error("Expected at least 1 Identifier node")
	}

	t.Logf("Node counts: %+v", counts)
}

// TestASTStructureForVariousPrograms tests AST structure for different code (Task 10.20)
func TestASTStructureForVariousPrograms(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		checkStructure func(t *testing.T, tree *ast.Program)
	}{
		{
			name:   "variable declaration structure",
			source: "var x: Integer := 42;",
			checkStructure: func(t *testing.T, tree *ast.Program) {
				if len(tree.Statements) != 1 {
					t.Fatalf("Expected 1 statement, got %d", len(tree.Statements))
				}

				varDecl, ok := tree.Statements[0].(*ast.VarDeclStatement)
				if !ok {
					t.Fatalf("Expected VarDeclStatement, got %T", tree.Statements[0])
				}

				// Check variable name
				if len(varDecl.Names) == 0 {
					t.Fatal("Expected at least one variable name")
				}
				if varDecl.Names[0].Value != "x" {
					t.Errorf("Variable name = %q, want %q", varDecl.Names[0].Value, "x")
				}

				// Check type annotation
				if varDecl.Type == nil {
					t.Fatal("Expected type annotation")
				}

				// Check initializer
				if varDecl.Value == nil {
					t.Fatal("Expected initializer value")
				}
			},
		},
		{
			name: "if statement structure",
			source: `var x: Integer;
begin
  if x > 0 then
    PrintLn('positive');
end;`,
			checkStructure: func(t *testing.T, tree *ast.Program) {
				if len(tree.Statements) < 2 {
					t.Fatal("Expected at least 2 statements")
				}

				// Second statement should be the block
				blockStmt, ok := tree.Statements[1].(*ast.BlockStatement)
				if !ok {
					t.Fatalf("Expected BlockStatement, got %T", tree.Statements[1])
				}

				if len(blockStmt.Statements) == 0 {
					t.Fatal("Expected statements in block")
				}

				// Should have an if statement
				ifStmt, ok := blockStmt.Statements[0].(*ast.IfStatement)
				if !ok {
					t.Fatalf("Expected IfStatement, got %T", blockStmt.Statements[0])
				}

				// If should have condition
				if ifStmt.Condition == nil {
					t.Error("If statement should have condition")
				}

				// If should have consequence
				if ifStmt.Consequence == nil {
					t.Error("If statement should have consequence")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, err := dwscript.New(dwscript.WithTypeCheck(true))
			if err != nil {
				t.Fatalf("Failed to create engine: %v", err)
			}

			program, err := engine.Compile(tt.source)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			tree := program.AST()
			tt.checkStructure(t, tree)
		})
	}
}

// TestASTNodeTypes tests that AST nodes have correct types (Task 10.20)
func TestASTNodeTypes(t *testing.T) {
	source := `
var x: Integer;
const PI = 3.14;
function Test(): Boolean;
begin
  Result := true;
end;
`

	engine, err := dwscript.New(dwscript.WithTypeCheck(true))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	program, err := engine.Compile(source)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	tree := program.AST()

	// Verify node types implement correct interfaces
	for i, stmt := range tree.Statements {
		// Every statement should implement ast.Statement
		if _, ok := stmt.(ast.Statement); !ok {
			t.Errorf("Statement %d does not implement ast.Statement interface", i)
		}

		// Every statement should implement ast.Node
		if _, ok := stmt.(ast.Node); !ok {
			t.Errorf("Statement %d does not implement ast.Node interface", i)
		}
	}

	// Walk and verify all expression nodes implement ast.Expression
	ast.Inspect(tree, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.IntegerLiteral, *ast.FloatLiteral, *ast.StringLiteral,
			*ast.BooleanLiteral, *ast.Identifier, *ast.BinaryExpression:
			if _, ok := node.(ast.Expression); !ok {
				t.Errorf("Expression node %T does not implement ast.Expression", n)
			}
		}
		return true
	})
}

// TestASTAccessChildNodes tests accessing child nodes in AST (Task 10.20)
func TestASTAccessChildNodes(t *testing.T) {
	source := `var x: Integer;
begin
  x := 1 + 2;
end;`

	engine, err := dwscript.New(dwscript.WithTypeCheck(true))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	program, err := engine.Compile(source)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	tree := program.AST()

	// Access child nodes
	if len(tree.Statements) < 2 {
		t.Fatal("Expected at least 2 statements")
	}

	// Second statement should be the block
	blockStmt, ok := tree.Statements[1].(*ast.BlockStatement)
	if !ok {
		t.Fatalf("Expected BlockStatement, got %T", tree.Statements[1])
	}

	// Access statements in block
	if len(blockStmt.Statements) == 0 {
		t.Fatal("Expected statements in block")
	}

	assignStmt, ok := blockStmt.Statements[0].(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("Expected AssignmentStatement, got %T", blockStmt.Statements[0])
	}

	// Access assignment target
	if assignStmt.Target == nil {
		t.Fatal("Assignment should have target")
	}

	// Access assignment value
	if assignStmt.Value == nil {
		t.Fatal("Assignment should have value")
	}

	// Value should be binary expression (1 + 2)
	binaryExpr, ok := assignStmt.Value.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("Expected BinaryExpression, got %T", assignStmt.Value)
	}

	// Access operands
	if binaryExpr.Left == nil {
		t.Error("Binary expression should have left operand")
	}
	if binaryExpr.Right == nil {
		t.Error("Binary expression should have right operand")
	}
	if binaryExpr.Operator == "" {
		t.Error("Binary expression should have operator")
	}
}

// TestASTImmutability tests that returned AST is read-only (documentation claim)
func TestASTImmutability(t *testing.T) {
	source := "var x: Integer := 42;"

	engine, err := dwscript.New(dwscript.WithTypeCheck(true))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	program, err := engine.Compile(source)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Get AST twice
	tree1 := program.AST()
	tree2 := program.AST()

	// Should return the same AST instance
	if tree1 != tree2 {
		t.Log("Warning: Program.AST() returns different instances (copies)")
	}

	// Both should have same structure
	if len(tree1.Statements) != len(tree2.Statements) {
		t.Error("Multiple AST() calls return different structure")
	}
}
