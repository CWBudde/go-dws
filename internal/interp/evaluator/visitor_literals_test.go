package evaluator

import (
	"math"
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// TestVisitIntegerLiteral tests the VisitIntegerLiteral method with various integer values.
func TestVisitIntegerLiteral(t *testing.T) {
	tests := []struct {
		name     string
		value    int64
		expected int64
	}{
		{"zero", 0, 0},
		{"positive", 42, 42},
		{"negative", -123, -123},
		{"max_int64", math.MaxInt64, math.MaxInt64},
		{"min_int64", math.MinInt64, math.MinInt64},
		{"one", 1, 1},
		{"minus_one", -1, -1},
	}

	e := &Evaluator{}
	ctx := &ExecutionContext{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &ast.IntegerLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.INT, Literal: ""},
					},
				},
				Value: tt.value,
			}

			result := e.VisitIntegerLiteral(node, ctx)

			intVal, ok := result.(*runtime.IntegerValue)
			if !ok {
				t.Fatalf("expected *runtime.IntegerValue, got %T", result)
			}

			if intVal.Value != tt.expected {
				t.Errorf("expected value %d, got %d", tt.expected, intVal.Value)
			}
		})
	}
}

// TestVisitFloatLiteral tests the VisitFloatLiteral method with various float values.
func TestVisitFloatLiteral(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		expected float64
	}{
		{"zero", 0.0, 0.0},
		{"positive", 3.14159, 3.14159},
		{"negative", -2.71828, -2.71828},
		{"small", 0.000001, 0.000001},
		{"large", 1.23e10, 1.23e10},
		{"negative_exp", 1.5e-5, 1.5e-5},
		{"one", 1.0, 1.0},
		{"minus_one", -1.0, -1.0},
		{"max_float64", math.MaxFloat64, math.MaxFloat64},
		{"smallest_nonzero", math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64},
	}

	e := &Evaluator{}
	ctx := &ExecutionContext{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &ast.FloatLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.FLOAT, Literal: ""},
					},
				},
				Value: tt.value,
			}

			result := e.VisitFloatLiteral(node, ctx)

			floatVal, ok := result.(*runtime.FloatValue)
			if !ok {
				t.Fatalf("expected *runtime.FloatValue, got %T", result)
			}

			if floatVal.Value != tt.expected {
				t.Errorf("expected value %f, got %f", tt.expected, floatVal.Value)
			}
		})
	}
}

// TestVisitStringLiteral tests the VisitStringLiteral method with various string values.
func TestVisitStringLiteral(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"empty", "", ""},
		{"simple", "hello", "hello"},
		{"with_spaces", "hello world", "hello world"},
		{"with_quotes", "it's a test", "it's a test"},
		{"multiline", "line1\nline2\nline3", "line1\nline2\nline3"},
		{"unicode", "Hello 世界 🌍", "Hello 世界 🌍"},
		{"escaped_chars", "tab\there\nand newline", "tab\there\nand newline"},
		{"long_string", string(make([]byte, 1000)), string(make([]byte, 1000))},
	}

	e := &Evaluator{}
	ctx := &ExecutionContext{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &ast.StringLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.STRING, Literal: ""},
					},
				},
				Value: tt.value,
			}

			result := e.VisitStringLiteral(node, ctx)

			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected *runtime.StringValue, got %T", result)
			}

			if strVal.Value != tt.expected {
				t.Errorf("expected value %q, got %q", tt.expected, strVal.Value)
			}
		})
	}
}

// TestVisitBooleanLiteral tests the VisitBooleanLiteral method.
func TestVisitBooleanLiteral(t *testing.T) {
	tests := []struct {
		name     string
		value    bool
		expected bool
	}{
		{"true", true, true},
		{"false", false, false},
	}

	e := &Evaluator{}
	ctx := &ExecutionContext{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &ast.BooleanLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.TRUE, Literal: ""},
					},
				},
				Value: tt.value,
			}

			result := e.VisitBooleanLiteral(node, ctx)

			boolVal, ok := result.(*runtime.BooleanValue)
			if !ok {
				t.Fatalf("expected *runtime.BooleanValue, got %T", result)
			}

			if boolVal.Value != tt.expected {
				t.Errorf("expected value %v, got %v", tt.expected, boolVal.Value)
			}
		})
	}
}

// TestVisitCharLiteral tests the VisitCharLiteral method with various character values.
func TestVisitCharLiteral(t *testing.T) {
	tests := []struct {
		name     string
		value    rune
		expected string
	}{
		{"ascii_a", 'a', "a"},
		{"ascii_Z", 'Z', "Z"},
		{"digit", '5', "5"},
		{"space", ' ', " "},
		{"newline", '\n', "\n"},
		{"tab", '\t', "\t"},
		{"unicode_emoji", '🌍', "🌍"},
		{"unicode_chinese", '中', "中"},
		{"null_char", '\x00', "\x00"},
	}

	e := &Evaluator{}
	ctx := &ExecutionContext{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &ast.CharLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.CHAR, Literal: ""},
					},
				},
				Value: tt.value,
			}

			result := e.VisitCharLiteral(node, ctx)

			strVal, ok := result.(*runtime.StringValue)
			if !ok {
				t.Fatalf("expected *runtime.StringValue, got %T", result)
			}

			if strVal.Value != tt.expected {
				t.Errorf("expected value %q, got %q", tt.expected, strVal.Value)
			}
		})
	}
}

// TestVisitNilLiteral tests the VisitNilLiteral method.
func TestVisitNilLiteral(t *testing.T) {
	e := &Evaluator{}
	ctx := &ExecutionContext{}

	node := &ast.NilLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Type: token.NIL, Literal: "nil"},
			},
		},
	}

	result := e.VisitNilLiteral(node, ctx)

	nilVal, ok := result.(*runtime.NilValue)
	if !ok {
		t.Fatalf("expected *runtime.NilValue, got %T", result)
	}

	// NilValue should not have a ClassType when created from a literal
	if nilVal.ClassType != "" {
		t.Errorf("expected empty ClassType for literal nil, got %q", nilVal.ClassType)
	}
}

// TestVisitNilLiteral_Multiple tests that multiple calls create distinct instances.
func TestVisitNilLiteral_Multiple(t *testing.T) {
	e := &Evaluator{}
	ctx := &ExecutionContext{}

	node := &ast.NilLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Type: token.NIL, Literal: "nil"},
			},
		},
	}

	result1 := e.VisitNilLiteral(node, ctx)
	result2 := e.VisitNilLiteral(node, ctx)

	// Both should be NilValue instances
	nilVal1, ok1 := result1.(*runtime.NilValue)
	nilVal2, ok2 := result2.(*runtime.NilValue)

	if !ok1 || !ok2 {
		t.Fatalf("expected both results to be *runtime.NilValue")
	}

	// They should be distinct instances (not the same pointer)
	// This is important for proper memory management
	if nilVal1 == nilVal2 {
		t.Errorf("expected distinct NilValue instances, got same pointer")
	}
}

// TestVisitorLiterals_NilContext tests that literal visitors work with nil context.
// This is an edge case test to ensure robustness.
func TestVisitorLiterals_NilContext(t *testing.T) {
	e := &Evaluator{}

	t.Run("integer_with_nil_context", func(t *testing.T) {
		node := &ast.IntegerLiteral{Value: 42}
		result := e.VisitIntegerLiteral(node, nil)
		if intVal, ok := result.(*runtime.IntegerValue); !ok || intVal.Value != 42 {
			t.Errorf("expected IntegerValue(42), got %v", result)
		}
	})

	t.Run("float_with_nil_context", func(t *testing.T) {
		node := &ast.FloatLiteral{Value: 3.14}
		result := e.VisitFloatLiteral(node, nil)
		if floatVal, ok := result.(*runtime.FloatValue); !ok || floatVal.Value != 3.14 {
			t.Errorf("expected FloatValue(3.14), got %v", result)
		}
	})

	t.Run("string_with_nil_context", func(t *testing.T) {
		node := &ast.StringLiteral{Value: "test"}
		result := e.VisitStringLiteral(node, nil)
		if strVal, ok := result.(*runtime.StringValue); !ok || strVal.Value != "test" {
			t.Errorf("expected StringValue('test'), got %v", result)
		}
	})

	t.Run("boolean_with_nil_context", func(t *testing.T) {
		node := &ast.BooleanLiteral{Value: true}
		result := e.VisitBooleanLiteral(node, nil)
		if boolVal, ok := result.(*runtime.BooleanValue); !ok || !boolVal.Value {
			t.Errorf("expected BooleanValue(true), got %v", result)
		}
	})

	t.Run("char_with_nil_context", func(t *testing.T) {
		node := &ast.CharLiteral{Value: 'x'}
		result := e.VisitCharLiteral(node, nil)
		if strVal, ok := result.(*runtime.StringValue); !ok || strVal.Value != "x" {
			t.Errorf("expected StringValue('x'), got %v", result)
		}
	})

	t.Run("nil_with_nil_context", func(t *testing.T) {
		node := &ast.NilLiteral{}
		result := e.VisitNilLiteral(node, nil)
		if _, ok := result.(*runtime.NilValue); !ok {
			t.Errorf("expected NilValue, got %T", result)
		}
	})
}
