package ast_test

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// TestPositionSimpleStatements tests position tracking on simple statements (Task 10.19)
func TestPositionSimpleStatements(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		checkPos  func(t *testing.T, program *ast.Program)
	}{
		{
			name:   "variable declaration",
			source: "var x: Integer := 42;",
			checkPos: func(t *testing.T, program *ast.Program) {
				if len(program.Statements) == 0 {
					t.Fatal("Expected at least one statement")
				}
				stmt := program.Statements[0]

				// Check that Pos() returns valid position
				pos := stmt.Pos()
				if pos.Line != 1 {
					t.Errorf("Statement Pos().Line = %d, want 1", pos.Line)
				}
				if pos.Column != 1 {
					t.Errorf("Statement Pos().Column = %d, want 1", pos.Column)
				}

				// Check that End() returns valid position
				end := stmt.End()
				if end.Line != 1 {
					t.Errorf("Statement End().Line = %d, want 1", end.Line)
				}
				// End should be after start
				if end.Column <= pos.Column {
					t.Errorf("Statement End().Column (%d) should be > Pos().Column (%d)",
						end.Column, pos.Column)
				}
			},
		},
		{
			name:   "assignment statement",
			source: "begin x := 5; end;",
			checkPos: func(t *testing.T, program *ast.Program) {
				if len(program.Statements) == 0 {
					t.Fatal("Expected at least one statement")
				}

				// Should have a block statement
				blockStmt, ok := program.Statements[0].(*ast.BlockStatement)
				if !ok {
					t.Fatalf("Expected BlockStatement, got %T", program.Statements[0])
				}

				// Block should have valid position
				pos := blockStmt.Pos()
				if pos.Line != 1 {
					t.Errorf("Block Pos().Line = %d, want 1", pos.Line)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.source)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			tt.checkPos(t, program)
		})
	}
}

// TestPositionNestedExpressions tests position tracking on nested expressions (Task 10.19)
func TestPositionNestedExpressions(t *testing.T) {
	source := "var result: Integer := (1 + 2) * 3;"

	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	if len(program.Statements) == 0 {
		t.Fatal("Expected at least one statement")
	}

	varDecl, ok := program.Statements[0].(*ast.VarDeclStatement)
	if !ok {
		t.Fatalf("Expected VarDeclStatement, got %T", program.Statements[0])
	}

	// Check that the variable declaration has valid position
	pos := varDecl.Pos()
	if pos.Line != 1 {
		t.Errorf("VarDecl Pos().Line = %d, want 1", pos.Line)
	}

	// The initializer should have position info
	if varDecl.Value != nil {
		valuePos := varDecl.Value.Pos()
		if !valuePos.IsValid() {
			t.Error("Value expression should have valid position")
		}
	}
}

// TestPositionMultiLineConstructs tests position tracking on multi-line code (Task 10.19)
func TestPositionMultiLineConstructs(t *testing.T) {
	source := `begin
  x := 1;
  y := 2;
end;`

	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	if len(program.Statements) == 0 {
		t.Fatal("Expected at least one statement")
	}

	blockStmt, ok := program.Statements[0].(*ast.BlockStatement)
	if !ok {
		t.Fatalf("Expected BlockStatement, got %T", program.Statements[0])
	}

	// Block should start on line 1
	startPos := blockStmt.Pos()
	if startPos.Line != 1 {
		t.Errorf("Block start line = %d, want 1", startPos.Line)
	}

	// Block should end on line 4
	endPos := blockStmt.End()
	if endPos.Line != 4 {
		t.Errorf("Block end line = %d, want 4", endPos.Line)
	}

	// Statements inside should have correct lines
	if len(blockStmt.Statements) >= 2 {
		stmt1Pos := blockStmt.Statements[0].Pos()
		stmt2Pos := blockStmt.Statements[1].Pos()

		if stmt1Pos.Line != 2 {
			t.Errorf("First statement line = %d, want 2", stmt1Pos.Line)
		}
		if stmt2Pos.Line != 3 {
			t.Errorf("Second statement line = %d, want 3", stmt2Pos.Line)
		}
	}
}

// TestPosition1BasedLineNumbering verifies 1-based line numbering (Task 10.19)
func TestPosition1BasedLineNumbering(t *testing.T) {
	source := "var x: Integer;"

	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	// Check program position
	progPos := program.Pos()
	if progPos.Line < 1 {
		t.Errorf("Program position line = %d, should be >= 1 (1-based)", progPos.Line)
	}
	if progPos.Column < 1 {
		t.Errorf("Program position column = %d, should be >= 1 (1-based)", progPos.Column)
	}

	// Check statement position
	if len(program.Statements) > 0 {
		stmtPos := program.Statements[0].Pos()
		if stmtPos.Line < 1 {
			t.Errorf("Statement position line = %d, should be >= 1 (1-based)", stmtPos.Line)
		}
		if stmtPos.Column < 1 {
			t.Errorf("Statement position column = %d, should be >= 1 (1-based)", stmtPos.Column)
		}
	}
}

// TestPositionPosAndEndMethods tests Pos() and End() on various node types (Task 10.19)
func TestPositionPosAndEndMethods(t *testing.T) {
	source := `
var x: Integer := 42;
function Add(a, b: Integer): Integer;
begin
  Result := a + b;
end;
`

	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	// Walk all nodes and verify Pos() and End() are implemented
	nodeCount := 0
	ast.Inspect(program, func(node ast.Node) bool {
		if node == nil {
			return false
		}
		nodeCount++

		// Every node should implement Pos() and End()
		pos := node.Pos()
		end := node.End()

		// Pos() should be valid (line >= 1)
		if pos.Line < 1 {
			t.Errorf("Node %T has invalid Pos().Line = %d", node, pos.Line)
		}

		// End() should be valid (line >= 1)
		if end.Line < 1 {
			t.Errorf("Node %T has invalid End().Line = %d", node, end.Line)
		}

		// End should be >= Pos (same line, or later line)
		if end.Line < pos.Line {
			t.Errorf("Node %T End().Line (%d) < Pos().Line (%d)",
				node, end.Line, pos.Line)
		}
		if end.Line == pos.Line && end.Column < pos.Column {
			t.Errorf("Node %T on same line: End().Column (%d) < Pos().Column (%d)",
				node, end.Column, pos.Column)
		}

		return true
	})

	if nodeCount == 0 {
		t.Error("Expected to visit some nodes")
	}
	t.Logf("Verified position methods on %d AST nodes", nodeCount)
}

// TestPositionWithUnicode tests position tracking with Unicode characters (Task 10.19)
func TestPositionWithUnicode(t *testing.T) {
	// Unicode characters: "Hello 世界" contains multibyte chars
	source := "var message: String := 'Hello 世界';"

	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	// Parser should handle Unicode without errors
	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors with Unicode: %v", p.Errors())
	}

	if len(program.Statements) == 0 {
		t.Fatal("Expected at least one statement")
	}

	// Statement should have valid position
	stmt := program.Statements[0]
	pos := stmt.Pos()
	end := stmt.End()

	if !pos.IsValid() {
		t.Error("Position should be valid for Unicode source")
	}
	if !end.IsValid() {
		t.Error("End position should be valid for Unicode source")
	}
}

// TestPositionIsValid tests the Position.IsValid() method
func TestPositionIsValid(t *testing.T) {
	tests := []struct {
		name  string
		pos   token.Position
		want  bool
	}{
		{
			name: "valid position",
			pos:  token.Position{Line: 1, Column: 1, Offset: 0},
			want: true,
		},
		{
			name: "valid position line 10",
			pos:  token.Position{Line: 10, Column: 5, Offset: 100},
			want: true,
		},
		{
			name: "invalid position (line 0)",
			pos:  token.Position{Line: 0, Column: 1, Offset: 0},
			want: false,
		},
		{
			name: "invalid position (negative line)",
			pos:  token.Position{Line: -1, Column: 1, Offset: 0},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pos.IsValid()
			if got != tt.want {
				t.Errorf("Position.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
