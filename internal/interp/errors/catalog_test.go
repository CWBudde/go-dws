package errors

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// ============================================================================
// Type Error Tests
// ============================================================================

func TestTypeMismatchError(t *testing.T) {
	pos := &token.Position{Line: 10, Column: 5}
	expr := "x + y"

	err := TypeMismatchError(pos, expr, "INTEGER", "+", "STRING")

	if err.Category != CategoryType {
		t.Errorf("Expected category Type, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "type mismatch") {
		t.Errorf("Expected 'type mismatch' in message, got %q", err.Message)
	}
	if !strings.Contains(err.Message, "INTEGER") || !strings.Contains(err.Message, "STRING") {
		t.Errorf("Expected types in message, got %q", err.Message)
	}
}

func TestUnknownOperatorError(t *testing.T) {
	pos := &token.Position{Line: 15, Column: 8}
	expr := "a ++ b"

	err := UnknownOperatorError(pos, expr, "INTEGER", "++", "INTEGER")

	if err.Category != CategoryType {
		t.Errorf("Expected category Type, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "unknown operator") {
		t.Errorf("Expected 'unknown operator' in message, got %q", err.Message)
	}
}

func TestCannotConvertError(t *testing.T) {
	pos := &token.Position{Line: 20, Column: 12}
	expr := "Integer(s)"

	err := CannotConvertError(pos, expr, "STRING", "INTEGER")

	if err.Category != CategoryType {
		t.Errorf("Expected category Type, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "cannot convert") {
		t.Errorf("Expected 'cannot convert' in message, got %q", err.Message)
	}
	if !strings.Contains(err.Message, "STRING") || !strings.Contains(err.Message, "INTEGER") {
		t.Errorf("Expected types in message, got %q", err.Message)
	}
}

func TestCannotConvertValueError(t *testing.T) {
	pos := &token.Position{Line: 25, Column: 3}
	expr := "Integer('abc')"

	err := CannotConvertValueError(pos, expr, "STRING", "abc", "INTEGER")

	if err.Category != CategoryType {
		t.Errorf("Expected category Type, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "cannot convert") {
		t.Errorf("Expected 'cannot convert' in message, got %q", err.Message)
	}
	if !strings.Contains(err.Message, "abc") {
		t.Errorf("Expected value 'abc' in message, got %q", err.Message)
	}
}

func TestCannotCastError(t *testing.T) {
	pos := &token.Position{Line: 30, Column: 7}
	expr := "MyClass(obj)"

	err := CannotCastError(pos, expr, "VARIANT", "MyClass")

	if err.Category != CategoryType {
		t.Errorf("Expected category Type, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "cannot cast") {
		t.Errorf("Expected 'cannot cast' in message, got %q", err.Message)
	}
}

func TestExpectedTypeError(t *testing.T) {
	pos := &token.Position{Line: 35, Column: 10}
	expr := "x"

	err := ExpectedTypeError(pos, expr, "INTEGER", "STRING")

	if err.Category != CategoryType {
		t.Errorf("Expected category Type, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "expected") {
		t.Errorf("Expected 'expected' in message, got %q", err.Message)
	}
	if !strings.Contains(err.Message, "INTEGER") || !strings.Contains(err.Message, "STRING") {
		t.Errorf("Expected types in message, got %q", err.Message)
	}
}

// ============================================================================
// Runtime Error Tests
// ============================================================================

func TestDivisionByZeroError(t *testing.T) {
	pos := &token.Position{Line: 40, Column: 15}
	expr := "x / y"

	err := DivisionByZeroError(pos, expr, 10, 0)

	if err.Category != CategoryRuntime {
		t.Errorf("Expected category Runtime, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "division by zero") {
		t.Errorf("Expected 'division by zero' in message, got %q", err.Message)
	}
	if !strings.Contains(err.Message, "10") {
		t.Errorf("Expected value '10' in message, got %q", err.Message)
	}
}

func TestIntegerDivByZeroError(t *testing.T) {
	pos := &token.Position{Line: 45, Column: 20}
	expr := "x div y"

	err := IntegerDivByZeroError(pos, expr, 15, 0)

	if err.Category != CategoryRuntime {
		t.Errorf("Expected category Runtime, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "division by zero") {
		t.Errorf("Expected 'division by zero' in message, got %q", err.Message)
	}
	if !strings.Contains(err.Message, "div") {
		t.Errorf("Expected 'div' in message, got %q", err.Message)
	}
}

func TestIndexOutOfBoundsError(t *testing.T) {
	pos := &token.Position{Line: 50, Column: 8}
	expr := "arr[10]"

	err := IndexOutOfBoundsError(pos, expr, 10, 5)

	if err.Category != CategoryRuntime {
		t.Errorf("Expected category Runtime, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "index out of bounds") {
		t.Errorf("Expected 'index out of bounds' in message, got %q", err.Message)
	}
	if !strings.Contains(err.Message, "10") || !strings.Contains(err.Message, "5") {
		t.Errorf("Expected index and length in message, got %q", err.Message)
	}
}

func TestIndexOutOfBoundsRangeError(t *testing.T) {
	pos := &token.Position{Line: 55, Column: 12}
	expr := "arr[10]"

	err := IndexOutOfBoundsRangeError(pos, expr, 10, 1, 5)

	if err.Category != CategoryRuntime {
		t.Errorf("Expected category Runtime, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "index out of bounds") {
		t.Errorf("Expected 'index out of bounds' in message, got %q", err.Message)
	}
	if !strings.Contains(err.Message, "1") || !strings.Contains(err.Message, "5") {
		t.Errorf("Expected range bounds in message, got %q", err.Message)
	}
}

func TestStringIndexOutOfBoundsError(t *testing.T) {
	pos := &token.Position{Line: 60, Column: 5}
	expr := "s[10]"

	err := StringIndexOutOfBoundsError(pos, expr, 10, 3)

	if err.Category != CategoryRuntime {
		t.Errorf("Expected category Runtime, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "string index out of bounds") {
		t.Errorf("Expected 'string index out of bounds' in message, got %q", err.Message)
	}
	if !strings.Contains(err.Message, "10") || !strings.Contains(err.Message, "3") {
		t.Errorf("Expected index and length in message, got %q", err.Message)
	}
}

func TestWrongArgumentCountError(t *testing.T) {
	pos := &token.Position{Line: 65, Column: 10}
	expr := "Sqrt(1, 2)"

	err := WrongArgumentCountError(pos, expr, 1, 2)

	if err.Category != CategoryRuntime {
		t.Errorf("Expected category Runtime, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "wrong number of arguments") {
		t.Errorf("Expected 'wrong number of arguments' in message, got %q", err.Message)
	}
	if !strings.Contains(err.Message, "1") || !strings.Contains(err.Message, "2") {
		t.Errorf("Expected argument counts in message, got %q", err.Message)
	}
}

func TestWrongArgumentCountForError(t *testing.T) {
	pos := &token.Position{Line: 70, Column: 15}
	expr := "Sqrt(1, 2)"

	err := WrongArgumentCountForError(pos, expr, "Sqrt", 1, 2)

	if err.Category != CategoryRuntime {
		t.Errorf("Expected category Runtime, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "wrong number of arguments") {
		t.Errorf("Expected 'wrong number of arguments' in message, got %q", err.Message)
	}
	if !strings.Contains(err.Message, "Sqrt") {
		t.Errorf("Expected function name 'Sqrt' in message, got %q", err.Message)
	}
}

// ============================================================================
// Undefined Error Tests
// ============================================================================

func TestUndefinedVariableError(t *testing.T) {
	pos := &token.Position{Line: 75, Column: 8}
	expr := "x"

	err := UndefinedVariableError(pos, expr, "x")

	if err.Category != CategoryUndefined {
		t.Errorf("Expected category Undefined, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "undefined variable") {
		t.Errorf("Expected 'undefined variable' in message, got %q", err.Message)
	}
	if !strings.Contains(err.Message, "x") {
		t.Errorf("Expected variable name 'x' in message, got %q", err.Message)
	}
}

func TestUndefinedFunctionError(t *testing.T) {
	pos := &token.Position{Line: 80, Column: 12}
	expr := "DoSomething()"

	err := UndefinedFunctionError(pos, expr, "DoSomething")

	if err.Category != CategoryUndefined {
		t.Errorf("Expected category Undefined, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "undefined function") {
		t.Errorf("Expected 'undefined function' in message, got %q", err.Message)
	}
	if !strings.Contains(err.Message, "DoSomething") {
		t.Errorf("Expected function name in message, got %q", err.Message)
	}
}

func TestFunctionNotFoundError(t *testing.T) {
	pos := &token.Position{Line: 85, Column: 5}
	expr := "Foo()"

	err := FunctionNotFoundError(pos, expr, "Foo")

	if err.Category != CategoryUndefined {
		t.Errorf("Expected category Undefined, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "function or procedure not found") {
		t.Errorf("Expected 'function or procedure not found' in message, got %q", err.Message)
	}
}

func TestUndefinedTypeError(t *testing.T) {
	pos := &token.Position{Line: 90, Column: 7}
	expr := "var x: MyType"

	err := UndefinedTypeError(pos, expr, "MyType")

	if err.Category != CategoryUndefined {
		t.Errorf("Expected category Undefined, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "undefined type") {
		t.Errorf("Expected 'undefined type' in message, got %q", err.Message)
	}
	if !strings.Contains(err.Message, "MyType") {
		t.Errorf("Expected type name in message, got %q", err.Message)
	}
}

func TestMethodNotFoundError(t *testing.T) {
	pos := &token.Position{Line: 95, Column: 10}
	expr := "obj.ToString()"

	err := MethodNotFoundError(pos, expr, "ToString", "MyClass")

	if err.Category != CategoryUndefined {
		t.Errorf("Expected category Undefined, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "method not found") {
		t.Errorf("Expected 'method not found' in message, got %q", err.Message)
	}
	if !strings.Contains(err.Message, "ToString") || !strings.Contains(err.Message, "MyClass") {
		t.Errorf("Expected method and class name in message, got %q", err.Message)
	}
}

// ============================================================================
// Contract Error Tests
// ============================================================================

func TestPreconditionFailedError(t *testing.T) {
	pos := &token.Position{Line: 100, Column: 3}
	expr := "x > 0"

	err := PreconditionFailedError(pos, expr, "x > 0")

	if err.Category != CategoryContract {
		t.Errorf("Expected category Contract, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "precondition failed") {
		t.Errorf("Expected 'precondition failed' in message, got %q", err.Message)
	}
	if !strings.Contains(err.Message, "x > 0") {
		t.Errorf("Expected condition in message, got %q", err.Message)
	}
}

func TestPreconditionNonBoolError(t *testing.T) {
	pos := &token.Position{Line: 105, Column: 5}
	expr := "x"

	err := PreconditionNonBoolError(pos, expr, "INTEGER")

	if err.Category != CategoryContract {
		t.Errorf("Expected category Contract, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "precondition must evaluate to boolean") {
		t.Errorf("Expected precondition boolean message, got %q", err.Message)
	}
	if !strings.Contains(err.Message, "INTEGER") {
		t.Errorf("Expected type in message, got %q", err.Message)
	}
}

func TestPostconditionFailedError(t *testing.T) {
	pos := &token.Position{Line: 110, Column: 7}
	expr := "result > 0"

	err := PostconditionFailedError(pos, expr, "result > 0")

	if err.Category != CategoryContract {
		t.Errorf("Expected category Contract, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "postcondition failed") {
		t.Errorf("Expected 'postcondition failed' in message, got %q", err.Message)
	}
}

func TestPostconditionNonBoolError(t *testing.T) {
	pos := &token.Position{Line: 115, Column: 9}
	expr := "result"

	err := PostconditionNonBoolError(pos, expr, "STRING")

	if err.Category != CategoryContract {
		t.Errorf("Expected category Contract, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "postcondition must evaluate to boolean") {
		t.Errorf("Expected postcondition boolean message, got %q", err.Message)
	}
}

func TestAssertionFailedError(t *testing.T) {
	pos := &token.Position{Line: 120, Column: 12}
	expr := "Assert(x > 0)"

	err := AssertionFailedError(pos, expr, "x > 0")

	if err.Category != CategoryRuntime {
		t.Errorf("Expected category Runtime, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "assertion failed") {
		t.Errorf("Expected 'assertion failed' in message, got %q", err.Message)
	}
}

// ============================================================================
// Internal Error Tests
// ============================================================================

func TestUnknownNodeError(t *testing.T) {
	pos := &token.Position{Line: 125, Column: 5}
	expr := "unknown"
	node := &ast.Identifier{Value: "test"}

	err := UnknownNodeError(pos, expr, node)

	if err.Category != CategoryInternal {
		t.Errorf("Expected category Internal, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "internal error") {
		t.Errorf("Expected 'internal error' in message, got %q", err.Message)
	}
	if !strings.Contains(err.Message, "unknown node type") {
		t.Errorf("Expected 'unknown node type' in message, got %q", err.Message)
	}
}

func TestNotImplementedError(t *testing.T) {
	pos := &token.Position{Line: 130, Column: 8}
	expr := "feature"

	err := NotImplementedError(pos, expr, "some feature")

	if err.Category != CategoryInternal {
		t.Errorf("Expected category Internal, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "not implemented") {
		t.Errorf("Expected 'not implemented' in message, got %q", err.Message)
	}
	if !strings.Contains(err.Message, "some feature") {
		t.Errorf("Expected feature name in message, got %q", err.Message)
	}
}

// ============================================================================
// Convenience Function Tests
// ============================================================================

func TestErrTypeMismatch(t *testing.T) {
	node := &ast.BinaryExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Pos: token.Position{Line: 10, Column: 5}},
			},
		},
		Left: &ast.IntegerLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.INT, Literal: "5"},
				},
			},
			Value: 5,
		},
		Operator: "+",
		Right: &ast.StringLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.STRING, Literal: "\"hello\""},
				},
			},
			Value: "hello",
		},
	}

	err := ErrTypeMismatch(node, "INTEGER", "+", "STRING")

	if err.Category != CategoryType {
		t.Errorf("Expected category Type, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "type mismatch") {
		t.Errorf("Expected 'type mismatch' in message, got %q", err.Message)
	}
}

func TestErrDivByZero(t *testing.T) {
	node := &ast.BinaryExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Pos: token.Position{Line: 15, Column: 10}},
			},
		},
		Left: &ast.IntegerLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.INT, Literal: "10"},
				},
			},
			Value: 10,
		},
		Operator: "/",
		Right: &ast.IntegerLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.INT, Literal: "0"},
				},
			},
			Value: 0,
		},
	}

	err := ErrDivByZero(node, 10, 0)

	if err.Category != CategoryRuntime {
		t.Errorf("Expected category Runtime, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "division by zero") {
		t.Errorf("Expected 'division by zero' in message, got %q", err.Message)
	}
}

func TestErrUndefinedVariable(t *testing.T) {
	node := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{
					Type:    token.IDENT,
					Literal: "x",
					Pos:     token.Position{Line: 20, Column: 5},
				},
			},
		},
		Value: "x",
	}

	err := ErrUndefinedVariable(node, "x")

	if err.Category != CategoryUndefined {
		t.Errorf("Expected category Undefined, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "undefined variable") {
		t.Errorf("Expected 'undefined variable' in message, got %q", err.Message)
	}
	if !strings.Contains(err.Message, "x") {
		t.Errorf("Expected variable name in message, got %q", err.Message)
	}
}

func TestErrWrongArgCount(t *testing.T) {
	node := &ast.CallExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Pos: token.Position{Line: 25, Column: 8}},
			},
		},
		Function: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.IDENT, Literal: "Sqrt"},
				},
			},
			Value: "Sqrt",
		},
		Arguments: []ast.Expression{},
	}

	err := ErrWrongArgCount(node, 2, 3)

	if err.Category != CategoryRuntime {
		t.Errorf("Expected category Runtime, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "wrong number of arguments") {
		t.Errorf("Expected 'wrong number of arguments' in message, got %q", err.Message)
	}
}

func TestErrNotImplemented(t *testing.T) {
	node := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Pos: token.Position{Line: 30, Column: 12}},
			},
		},
		Value: "feature",
	}

	err := ErrNotImplemented(node, "advanced feature")

	if err.Category != CategoryInternal {
		t.Errorf("Expected category Internal, got %s", err.Category)
	}
	if !strings.Contains(err.Message, "not implemented") {
		t.Errorf("Expected 'not implemented' in message, got %q", err.Message)
	}
	if !strings.Contains(err.Message, "advanced feature") {
		t.Errorf("Expected feature name in message, got %q", err.Message)
	}
}

// ============================================================================
// Error Message Format Tests
// ============================================================================

func TestErrorMessageFormat(t *testing.T) {
	tests := []struct {
		name     string
		err      *InterpreterError
		contains []string
	}{
		{
			name:     "Type error format",
			err:      TypeMismatchError(&token.Position{Line: 1, Column: 1}, "x+y", "INTEGER", "+", "STRING"),
			contains: []string{"Type error", "line 1", "column 1", "type mismatch"},
		},
		{
			name:     "Runtime error format",
			err:      DivisionByZeroError(&token.Position{Line: 5, Column: 10}, "x/0", 10, 0),
			contains: []string{"Runtime error", "line 5", "column 10", "division by zero"},
		},
		{
			name:     "Undefined error format",
			err:      UndefinedVariableError(&token.Position{Line: 10, Column: 5}, "x", "x"),
			contains: []string{"Undefined error", "line 10", "column 5", "undefined variable"},
		},
		{
			name:     "Contract error format",
			err:      PreconditionFailedError(&token.Position{Line: 15, Column: 3}, "x>0", "x > 0"),
			contains: []string{"Contract error", "line 15", "column 3", "precondition failed"},
		},
		{
			name:     "Internal error format",
			err:      NotImplementedError(&token.Position{Line: 20, Column: 7}, "feat", "some feature"),
			contains: []string{"Internal error", "line 20", "column 7", "not implemented"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.err.Error()
			for _, substr := range tt.contains {
				if !strings.Contains(errMsg, substr) {
					t.Errorf("Expected error message to contain %q, got: %s", substr, errMsg)
				}
			}
		})
	}
}
