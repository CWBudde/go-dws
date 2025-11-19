package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// Task 2.2.13: Test IS/AS/IMPLEMENTS expression handler migration to cursor mode
//
// This file tests the migration of type checking and casting expression handlers:
// - parseIsExpression: obj is TClass
// - parseAsExpression: obj as IInterface
// - parseImplementsExpression: obj implements IInterface
//
// Each test parses the same source in both traditional and cursor modes,
// then compares the resulting ASTs using String() representation to verify
// semantic equivalence (ignoring minor Token position differences).

// TestParseIsExpression_Traditional_vs_Cursor tests the 'is' operator
func TestParseIsExpression_Traditional_vs_Cursor(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"simple type check", "obj is TMyClass"},
		{"interface type check", "instance is IMyInterface"},
		{"nested type", "item is TNamespace.TClass"},
		{"in condition", "if obj is TTest then result := true"},
		{"in comparison", "(x is TFoo) and (y is TBar)"},
		{"negated", "not (obj is TClass)"},
		{"with method call", "GetObject() is TClass"},
		{"with member access", "container.Item is TClass"},
		{"with array index", "arr[0] is TClass"},
		{"complex left side", "GetContainer().Items[i] is TClass"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Traditional mode
			tradParser := New(lexer.New(tt.source))
			tradProgram := tradParser.ParseProgram()
			checkParserErrors(t, tradParser)

			// Cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorProgram := cursorParser.ParseProgram()
			checkParserErrors(t, cursorParser)

			// Compare ASTs
			if len(tradProgram.Statements) != len(cursorProgram.Statements) {
				t.Fatalf("Statement count mismatch: traditional=%d, cursor=%d",
					len(tradProgram.Statements), len(cursorProgram.Statements))
			}

			if len(tradProgram.Statements) == 0 {
				return
			}

			// Compare full program strings
			if tradProgram.String() != cursorProgram.String() {
				t.Errorf("Program String mismatch:\nTraditional: %s\nCursor: %s",
					tradProgram.String(), cursorProgram.String())
			}
		})
	}
}

// TestParseAsExpression_Traditional_vs_Cursor tests the 'as' operator
func TestParseAsExpression_Traditional_vs_Cursor(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"simple cast", "obj as IMyInterface"},
		{"interface cast", "instance as IDisposable"},
		{"nested type", "item as TNamespace.IInterface"},
		{"in assignment", "intf := obj as IMyInterface"},
		{"with method call", "GetObject() as IInterface"},
		{"with member access", "container.Item as IInterface"},
		{"with array index", "arr[0] as IInterface"},
		{"chained cast", "(obj as IFoo).Method()"},
		{"cast then member", "(obj as IFoo).Field"},
		{"complex expression", "GetContainer().Items[i] as IInterface"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Traditional mode
			tradParser := New(lexer.New(tt.source))
			tradProgram := tradParser.ParseProgram()
			checkParserErrors(t, tradParser)

			// Cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorProgram := cursorParser.ParseProgram()
			checkParserErrors(t, cursorParser)

			// Compare ASTs
			if len(tradProgram.Statements) != len(cursorProgram.Statements) {
				t.Fatalf("Statement count mismatch: traditional=%d, cursor=%d",
					len(tradProgram.Statements), len(cursorProgram.Statements))
			}

			if len(tradProgram.Statements) == 0 {
				return
			}

			// Compare full program strings
			if tradProgram.String() != cursorProgram.String() {
				t.Errorf("Program String mismatch:\nTraditional: %s\nCursor: %s",
					tradProgram.String(), cursorProgram.String())
			}
		})
	}
}

// TestParseImplementsExpression_Traditional_vs_Cursor tests the 'implements' operator
func TestParseImplementsExpression_Traditional_vs_Cursor(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"simple check", "obj implements IMyInterface"},
		{"interface check", "instance implements IDisposable"},
		{"nested type", "item implements TNamespace.IInterface"},
		{"in condition", "if obj implements ITest then result := true"},
		{"in comparison", "(x implements IFoo) and (y implements IBar)"},
		{"negated", "not (obj implements IInterface)"},
		{"with method call", "GetObject() implements IInterface"},
		{"with member access", "container.Item implements IInterface"},
		{"with array index", "arr[0] implements IInterface"},
		{"complex expression", "GetContainer().Items[i] implements IInterface"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Traditional mode
			tradParser := New(lexer.New(tt.source))
			tradProgram := tradParser.ParseProgram()
			checkParserErrors(t, tradParser)

			// Cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorProgram := cursorParser.ParseProgram()
			checkParserErrors(t, cursorParser)

			// Compare ASTs
			if len(tradProgram.Statements) != len(cursorProgram.Statements) {
				t.Fatalf("Statement count mismatch: traditional=%d, cursor=%d",
					len(tradProgram.Statements), len(cursorProgram.Statements))
			}

			if len(tradProgram.Statements) == 0 {
				return
			}

			// Compare full program strings
			if tradProgram.String() != cursorProgram.String() {
				t.Errorf("Program String mismatch:\nTraditional: %s\nCursor: %s",
					tradProgram.String(), cursorProgram.String())
			}
		})
	}
}

// TestTypeOperatorsIntegration_Traditional_vs_Cursor tests combinations
func TestTypeOperatorsIntegration_Traditional_vs_Cursor(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"is then as", "(obj is IFoo) and (obj as IFoo).Method()"},
		{"as in condition", "if obj as IFoo <> nil then result := true"},
		{"implements with call", "(GetObject() implements IInterface).Method()"},
		{"multiple is checks", "(x is TFoo) or (x is TBar)"},
		{"multiple implements", "(obj implements IFoo) and (obj implements IBar)"},
		{"as chain", "((obj as IFoo) as IBar).Method()"},
		{"is with member", "(container.Item is TClass).Field"},
		{"complex integration", "if (arr[i] is TTest) and (arr[i] as ITest).IsValid() then x := 1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Traditional mode
			tradParser := New(lexer.New(tt.source))
			tradProgram := tradParser.ParseProgram()
			checkParserErrors(t, tradParser)

			// Cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorProgram := cursorParser.ParseProgram()
			checkParserErrors(t, cursorParser)

			// Compare ASTs
			if len(tradProgram.Statements) != len(cursorProgram.Statements) {
				t.Fatalf("Statement count mismatch: traditional=%d, cursor=%d",
					len(tradProgram.Statements), len(cursorProgram.Statements))
			}

			if len(tradProgram.Statements) == 0 {
				return
			}

			// Compare full program strings
			if tradProgram.String() != cursorProgram.String() {
				t.Errorf("Program String mismatch:\nTraditional: %s\nCursor: %s",
					tradProgram.String(), cursorProgram.String())
			}
		})
	}
}

// TestTypeOperatorsEdgeCases_Traditional_vs_Cursor tests edge cases
func TestTypeOperatorsEdgeCases_Traditional_vs_Cursor(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		expectErrors bool
	}{
		{"is missing type", "obj is", true},
		{"as missing type", "obj as", true},
		{"implements missing type", "obj implements", true},
		{"is with expression", "obj is (x + y)", false},              // This parses as boolean comparison
		{"empty expression", "", false},                              // Empty program, no errors
		{"whitespace only", "   ", false},                            // Empty program
		{"is in expression", "result := obj is TClass", false},       // Valid: assignment of boolean
		{"as assignment", "intf := obj as IInterface", false},        // Valid: interface cast assignment
		{"implements result", "flag := obj implements ITest", false}, // Valid: boolean result
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Traditional mode
			tradParser := New(lexer.New(tt.source))
			tradProgram := tradParser.ParseProgram()
			tradHasErrors := len(tradParser.Errors()) > 0

			// Cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorProgram := cursorParser.ParseProgram()
			cursorHasErrors := len(cursorParser.Errors()) > 0

			// Both modes should agree on whether there are errors
			if tradHasErrors != cursorHasErrors {
				t.Errorf("Error state mismatch: traditional has errors=%v, cursor has errors=%v",
					tradHasErrors, cursorHasErrors)
				if tradHasErrors {
					t.Logf("Traditional errors: %v", tradParser.Errors())
				}
				if cursorHasErrors {
					t.Logf("Cursor errors: %v", cursorParser.Errors())
				}
			}

			// If we expect errors, make sure both modes have them
			if tt.expectErrors && (!tradHasErrors || !cursorHasErrors) {
				t.Errorf("Expected errors but got: traditional=%v, cursor=%v",
					tradHasErrors, cursorHasErrors)
			}

			// For valid programs, compare AST strings
			if !tt.expectErrors && !tradHasErrors && !cursorHasErrors {
				if tradProgram.String() != cursorProgram.String() {
					t.Errorf("Program String mismatch:\nTraditional: %s\nCursor: %s",
						tradProgram.String(), cursorProgram.String())
				}
			}
		})
	}
}

// TestTypeOperatorsNodeTypes_Traditional_vs_Cursor verifies correct AST node types
func TestTypeOperatorsNodeTypes_Traditional_vs_Cursor(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		expectedType string // "is", "as", or "implements"
	}{
		{"is node type", "obj is TClass", "is"},
		{"as node type", "obj as IInterface", "as"},
		{"implements node type", "obj implements IInterface", "implements"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Traditional mode
			tradParser := New(lexer.New(tt.source))
			tradProgram := tradParser.ParseProgram()
			checkParserErrors(t, tradParser)

			// Cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.source))
			cursorProgram := cursorParser.ParseProgram()
			checkParserErrors(t, cursorParser)

			if len(tradProgram.Statements) == 0 || len(cursorProgram.Statements) == 0 {
				t.Fatal("Expected at least one statement")
			}

			// Get expressions from ExpressionStatements
			tradStmt, ok := tradProgram.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("Traditional: expected ExpressionStatement, got %T", tradProgram.Statements[0])
			}
			cursorStmt, ok := cursorProgram.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("Cursor: expected ExpressionStatement, got %T", cursorProgram.Statements[0])
			}

			// Verify correct node type
			switch tt.expectedType {
			case "is":
				if _, ok := tradStmt.Expression.(*ast.IsExpression); !ok {
					t.Errorf("Traditional: expected IsExpression, got %T", tradStmt.Expression)
				}
				if _, ok := cursorStmt.Expression.(*ast.IsExpression); !ok {
					t.Errorf("Cursor: expected IsExpression, got %T", cursorStmt.Expression)
				}
			case "as":
				if _, ok := tradStmt.Expression.(*ast.AsExpression); !ok {
					t.Errorf("Traditional: expected AsExpression, got %T", tradStmt.Expression)
				}
				if _, ok := cursorStmt.Expression.(*ast.AsExpression); !ok {
					t.Errorf("Cursor: expected AsExpression, got %T", cursorStmt.Expression)
				}
			case "implements":
				if _, ok := tradStmt.Expression.(*ast.ImplementsExpression); !ok {
					t.Errorf("Traditional: expected ImplementsExpression, got %T", tradStmt.Expression)
				}
				if _, ok := cursorStmt.Expression.(*ast.ImplementsExpression); !ok {
					t.Errorf("Cursor: expected ImplementsExpression, got %T", cursorStmt.Expression)
				}
			}

			// Compare String representations
			if tradStmt.Expression.String() != cursorStmt.Expression.String() {
				t.Errorf("String mismatch:\nTraditional: %s\nCursor: %s",
					tradStmt.Expression.String(), cursorStmt.Expression.String())
			}
		})
	}
}
