package evaluator

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// TestErrorMessagesIncludeLocation verifies that error messages include source location information.
func TestErrorMessagesIncludeLocation(t *testing.T) {
	tests := []struct {
		name           string
		line           int
		column         int
		errorMessage   string
		expectedSubstr string
	}{
		{
			name:           "division by zero with location",
			line:           2,
			column:         15,
			errorMessage:   "division by zero",
			expectedSubstr: "division by zero at line 2, column: 15",
		},
		{
			name:           "type mismatch with location",
			line:           5,
			column:         8,
			errorMessage:   "type mismatch: expected Integer, got String",
			expectedSubstr: "type mismatch: expected Integer, got String at line 5, column: 8",
		},
		{
			name:           "undefined variable with location",
			line:           10,
			column:         3,
			errorMessage:   "undefined variable: foo",
			expectedSubstr: "undefined variable: foo at line 10, column: 3",
		},
	}

	e := &Evaluator{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a node with position information
			node := &ast.BinaryExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{
							Type:    token.SLASH,
							Literal: "/",
							Pos: lexer.Position{
								Line:   tt.line,
								Column: tt.column,
							},
						},
					},
				},
			}

			// Call newError with the node
			errVal := e.newError(node, "%s", tt.errorMessage)

			// Verify the error is of the right type
			errorValue, ok := errVal.(*ErrorValue)
			if !ok {
				t.Fatalf("expected *ErrorValue, got %T", errVal)
			}

			// Verify the error message contains location information
			errMsg := errorValue.String()
			if !strings.Contains(errMsg, tt.expectedSubstr) {
				t.Errorf("error message should contain location info\nExpected substring: %s\nGot: %s",
					tt.expectedSubstr, errMsg)
			}

			// Verify the location format is correct
			if !strings.Contains(errMsg, "at line") {
				t.Errorf("error message should contain 'at line' format, got: %s", errMsg)
			}
		})
	}
}

// TestErrorMessagesWithoutNode verifies that errors without nodes still work.
func TestErrorMessagesWithoutNode(t *testing.T) {
	e := &Evaluator{}

	errVal := e.newError(nil, "%s", "some error without location")

	errorValue, ok := errVal.(*ErrorValue)
	if !ok {
		t.Fatalf("expected *ErrorValue, got %T", errVal)
	}

	errMsg := errorValue.String()
	expectedMsg := "ERROR: some error without location"
	if errMsg != expectedMsg {
		t.Errorf("expected '%s', got '%s'", expectedMsg, errMsg)
	}

	// Should NOT contain location info when node is nil
	if strings.Contains(errMsg, "at line") {
		t.Errorf("error message should not contain location when node is nil, got: %s", errMsg)
	}
}

// TestErrorMessagesWithZeroLineNumber verifies that errors with invalid position don't add location.
func TestErrorMessagesWithZeroLineNumber(t *testing.T) {
	e := &Evaluator{}

	// Create a node with zero line number (invalid position)
	node := &ast.BinaryExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{
					Type:    token.SLASH,
					Literal: "/",
					Pos: lexer.Position{
						Line:   0, // Invalid
						Column: 0,
					},
				},
			},
		},
	}

	errVal := e.newError(node, "%s", "some error")

	errorValue, ok := errVal.(*ErrorValue)
	if !ok {
		t.Fatalf("expected *ErrorValue, got %T", errVal)
	}

	errMsg := errorValue.String()
	expectedMsg := "ERROR: some error"
	if errMsg != expectedMsg {
		t.Errorf("expected '%s', got '%s'", expectedMsg, errMsg)
	}

	// Should NOT contain location info when line number is 0
	if strings.Contains(errMsg, "at line") {
		t.Errorf("error message should not contain location when line is 0, got: %s", errMsg)
	}
}
