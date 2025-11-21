package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

func TestErrorCategory(t *testing.T) {
	tests := []struct {
		name     string
		category ErrorCategory
		expected string
	}{
		{"Type category", CategoryType, "Type"},
		{"Runtime category", CategoryRuntime, "Runtime"},
		{"Undefined category", CategoryUndefined, "Undefined"},
		{"Contract category", CategoryContract, "Contract"},
		{"Internal category", CategoryInternal, "Internal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.category) != tt.expected {
				t.Errorf("Expected category %s, got %s", tt.expected, tt.category)
			}
		})
	}
}

func TestNewTypeError(t *testing.T) {
	pos := &token.Position{Line: 10, Column: 5}
	expr := "x + y"
	message := "cannot add INTEGER and STRING"

	err := NewTypeError(pos, message, expr)

	if err.Category != CategoryType {
		t.Errorf("Expected category Type, got %s", err.Category)
	}
	if err.Message != message {
		t.Errorf("Expected message %q, got %q", message, err.Message)
	}
	if err.Pos != pos {
		t.Errorf("Expected position %v, got %v", pos, err.Pos)
	}
	if err.Expression != expr {
		t.Errorf("Expected expression %q, got %q", expr, err.Expression)
	}

	expectedErr := "Type error at line 10, column 5: cannot add INTEGER and STRING"
	if err.Error() != expectedErr {
		t.Errorf("Expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestNewTypeErrorf(t *testing.T) {
	pos := &token.Position{Line: 15, Column: 20}
	expr := "a div b"

	err := NewTypeErrorf(pos, expr, "cannot divide %s by %s", "INTEGER", "STRING")

	if err.Category != CategoryType {
		t.Errorf("Expected category Type, got %s", err.Category)
	}
	expectedMsg := "cannot divide INTEGER by STRING"
	if err.Message != expectedMsg {
		t.Errorf("Expected message %q, got %q", expectedMsg, err.Message)
	}
}

func TestNewRuntimeError(t *testing.T) {
	pos := &token.Position{Line: 20, Column: 15}
	expr := "arr[100]"
	message := "index out of bounds"

	err := NewRuntimeError(pos, message, expr)

	if err.Category != CategoryRuntime {
		t.Errorf("Expected category Runtime, got %s", err.Category)
	}
	if err.Message != message {
		t.Errorf("Expected message %q, got %q", message, err.Message)
	}

	expectedErr := "Runtime error at line 20, column 15: index out of bounds"
	if err.Error() != expectedErr {
		t.Errorf("Expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestNewRuntimeErrorf(t *testing.T) {
	pos := &token.Position{Line: 25, Column: 10}
	expr := "x div y"

	err := NewRuntimeErrorf(pos, expr, "division by %s", "zero")

	if err.Category != CategoryRuntime {
		t.Errorf("Expected category Runtime, got %s", err.Category)
	}
	expectedMsg := "division by zero"
	if err.Message != expectedMsg {
		t.Errorf("Expected message %q, got %q", expectedMsg, err.Message)
	}
}

func TestNewRuntimeErrorWithValues(t *testing.T) {
	pos := &token.Position{Line: 30, Column: 5}
	expr := "x div y"
	message := "division by zero"
	values := map[string]string{
		"x": "10",
		"y": "0",
	}

	err := NewRuntimeErrorWithValues(pos, expr, message, values)

	if err.Category != CategoryRuntime {
		t.Errorf("Expected category Runtime, got %s", err.Category)
	}
	if err.Values == nil {
		t.Error("Expected values to be set")
	}
	if err.Values["x"] != "10" {
		t.Errorf("Expected x=10, got x=%s", err.Values["x"])
	}
	if err.Values["y"] != "0" {
		t.Errorf("Expected y=0, got y=%s", err.Values["y"])
	}
}

func TestNewUndefinedError(t *testing.T) {
	pos := &token.Position{Line: 35, Column: 8}
	expr := "foo"
	message := "undefined variable: foo"

	err := NewUndefinedError(pos, message, expr)

	if err.Category != CategoryUndefined {
		t.Errorf("Expected category Undefined, got %s", err.Category)
	}
	if err.Message != message {
		t.Errorf("Expected message %q, got %q", message, err.Message)
	}

	expectedErr := "Undefined error at line 35, column 8: undefined variable: foo"
	if err.Error() != expectedErr {
		t.Errorf("Expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestNewUndefinedErrorf(t *testing.T) {
	pos := &token.Position{Line: 40, Column: 12}
	expr := "bar()"

	err := NewUndefinedErrorf(pos, expr, "undefined function: %s", "bar")

	if err.Category != CategoryUndefined {
		t.Errorf("Expected category Undefined, got %s", err.Category)
	}
	expectedMsg := "undefined function: bar"
	if err.Message != expectedMsg {
		t.Errorf("Expected message %q, got %q", expectedMsg, err.Message)
	}
}

func TestNewContractError(t *testing.T) {
	pos := &token.Position{Line: 45, Column: 3}
	expr := "x > 0"
	message := "precondition failed"

	err := NewContractError(pos, message, expr)

	if err.Category != CategoryContract {
		t.Errorf("Expected category Contract, got %s", err.Category)
	}
	if err.Message != message {
		t.Errorf("Expected message %q, got %q", message, err.Message)
	}

	expectedErr := "Contract error at line 45, column 3: precondition failed"
	if err.Error() != expectedErr {
		t.Errorf("Expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestNewContractErrorf(t *testing.T) {
	pos := &token.Position{Line: 50, Column: 7}
	expr := "result > 0"

	err := NewContractErrorf(pos, expr, "%s failed: %s", "postcondition", "result <= 0")

	if err.Category != CategoryContract {
		t.Errorf("Expected category Contract, got %s", err.Category)
	}
	expectedMsg := "postcondition failed: result <= 0"
	if err.Message != expectedMsg {
		t.Errorf("Expected message %q, got %q", expectedMsg, err.Message)
	}
}

func TestNewInternalError(t *testing.T) {
	pos := &token.Position{Line: 55, Column: 20}
	expr := "unknown node"
	message := "unexpected AST node type"

	err := NewInternalError(pos, message, expr)

	if err.Category != CategoryInternal {
		t.Errorf("Expected category Internal, got %s", err.Category)
	}
	if err.Message != message {
		t.Errorf("Expected message %q, got %q", message, err.Message)
	}

	expectedErr := "Internal error at line 55, column 20: unexpected AST node type"
	if err.Error() != expectedErr {
		t.Errorf("Expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestNewInternalErrorf(t *testing.T) {
	pos := &token.Position{Line: 60, Column: 15}
	expr := "unknown"

	err := NewInternalErrorf(pos, expr, "unknown node type: %T", &ast.Identifier{})

	if err.Category != CategoryInternal {
		t.Errorf("Expected category Internal, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "unknown node type") {
		t.Errorf("Expected message to contain 'unknown node type', got %q", err.Message)
	}
}

func TestWrapError(t *testing.T) {
	baseErr := errors.New("base error")
	pos := &token.Position{Line: 65, Column: 10}
	expr := "test"

	wrapped := WrapError(baseErr, CategoryRuntime, pos, expr)

	if wrapped.Category != CategoryRuntime {
		t.Errorf("Expected category Runtime, got %s", wrapped.Category)
	}
	if wrapped.Message != "base error" {
		t.Errorf("Expected message %q, got %q", "base error", wrapped.Message)
	}
	if wrapped.Err != baseErr {
		t.Error("Expected wrapped error to contain base error")
	}

	// Test error unwrapping
	if unwrapped := wrapped.Unwrap(); unwrapped != baseErr {
		t.Error("Expected Unwrap() to return base error")
	}

	// Test errors.Is
	if !errors.Is(wrapped, baseErr) {
		t.Error("Expected errors.Is to work with wrapped error")
	}
}

func TestWrapErrorf(t *testing.T) {
	baseErr := errors.New("original error")
	pos := &token.Position{Line: 70, Column: 5}
	expr := "test"

	wrapped := WrapErrorf(baseErr, CategoryType, pos, expr, "type mismatch: %s", "details")

	if wrapped.Category != CategoryType {
		t.Errorf("Expected category Type, got %s", wrapped.Category)
	}
	expectedMsg := "type mismatch: details"
	if wrapped.Message != expectedMsg {
		t.Errorf("Expected message %q, got %q", expectedMsg, wrapped.Message)
	}
	if wrapped.Err != baseErr {
		t.Error("Expected wrapped error to contain base error")
	}
}

func TestErrorWithoutPosition(t *testing.T) {
	err := NewTypeError(nil, "no position", "expr")

	expectedErr := "Type error: no position"
	if err.Error() != expectedErr {
		t.Errorf("Expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestPositionFromNode(t *testing.T) {
	tests := []struct {
		node     ast.Node
		expected *token.Position
		name     string
	}{
		{
			name: "Identifier",
			node: &ast.Identifier{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Pos: token.Position{Line: 1, Column: 5}},
					},
				},
				Value: "x",
			},
			expected: &token.Position{Line: 1, Column: 5},
		},
		{
			name: "IntegerLiteral",
			node: &ast.IntegerLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Pos: token.Position{Line: 2, Column: 10}},
					},
				},
				Value: 42,
			},
			expected: &token.Position{Line: 2, Column: 10},
		},
		{
			name: "BinaryExpression",
			node: &ast.BinaryExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Pos: token.Position{Line: 3, Column: 15}},
					},
				},
				Operator: "+",
			},
			expected: &token.Position{Line: 3, Column: 15},
		},
		{
			name:     "Nil node",
			node:     nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos := PositionFromNode(tt.node)
			if tt.expected == nil {
				if pos != nil {
					t.Errorf("Expected nil position, got %v", pos)
				}
			} else {
				if pos == nil {
					t.Error("Expected non-nil position")
				} else if pos.Line != tt.expected.Line || pos.Column != tt.expected.Column {
					t.Errorf("Expected position %v, got %v", tt.expected, pos)
				}
			}
		})
	}
}

func TestExpressionFromNode(t *testing.T) {
	tests := []struct {
		name     string
		node     ast.Node
		contains string // Check if output contains this string
	}{
		{
			name: "Identifier",
			node: &ast.Identifier{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.IDENT, Literal: "x"},
					},
				},
				Value: "x",
			},
			contains: "x",
		},
		{
			name: "IntegerLiteral",
			node: &ast.IntegerLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.INT, Literal: "42"},
					},
				},
				Value: 42,
			},
			contains: "42",
		},
		{
			name:     "Nil node",
			node:     nil,
			contains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := ExpressionFromNode(tt.node)
			if tt.contains == "" {
				if expr != "" {
					t.Errorf("Expected empty string, got %q", expr)
				}
			} else {
				if !strings.Contains(expr, tt.contains) {
					t.Errorf("Expected expression to contain %q, got %q", tt.contains, expr)
				}
			}
		})
	}
}

func TestErrorChaining(t *testing.T) {
	// Create a chain of errors
	baseErr := fmt.Errorf("base error")
	wrappedErr := fmt.Errorf("wrapped: %w", baseErr)

	pos := &token.Position{Line: 100, Column: 50}
	interpErr := WrapError(wrappedErr, CategoryRuntime, pos, "test")

	// Test that we can unwrap through the chain
	if !errors.Is(interpErr, baseErr) {
		t.Error("Expected errors.Is to find base error in chain")
	}

	// Test direct unwrap
	unwrapped := interpErr.Unwrap()
	if unwrapped != wrappedErr {
		t.Error("Expected Unwrap() to return immediate wrapped error")
	}
}

func TestErrorValues(t *testing.T) {
	values := map[string]string{
		"x":      "10",
		"y":      "0",
		"result": "undefined",
	}

	err := NewRuntimeErrorWithValues(
		&token.Position{Line: 80, Column: 20},
		"x div y",
		"division by zero",
		values,
	)

	if len(err.Values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(err.Values))
	}

	if err.Values["x"] != "10" {
		t.Errorf("Expected x=10, got %s", err.Values["x"])
	}
	if err.Values["y"] != "0" {
		t.Errorf("Expected y=0, got %s", err.Values["y"])
	}
	if err.Values["result"] != "undefined" {
		t.Errorf("Expected result=undefined, got %s", err.Values["result"])
	}
}
