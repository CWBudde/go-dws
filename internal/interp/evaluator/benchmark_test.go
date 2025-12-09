package evaluator

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/units"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// benchSink is a package-level variable to prevent the compiler from
// eliminating benchmark calls as dead code (DCE prevention).
var benchSink any

// Benchmark helpers

func createTestEvaluator() *Evaluator {
	config := DefaultConfig()
	config.MaxRecursionDepth = 1024
	typeSystem := interptypes.NewTypeSystem()
	unitRegistry := units.NewUnitRegistry([]string{"."})
	refCountMgr := runtime.NewRefCountManager()
	return NewEvaluator(typeSystem, &bytes.Buffer{}, config, unitRegistry, nil, refCountMgr)
}

func createTestContext() *ExecutionContext {
	env := runtime.NewEnvironment()
	return NewExecutionContext(env)
}

// ============================================================================
// Simple Literal Benchmarks - Test visitor overhead for simple operations
// ============================================================================

// BenchmarkVisitIntegerLiteral tests the performance of evaluating integer literals
func BenchmarkVisitIntegerLiteral(b *testing.B) {
	eval := createTestEvaluator()
	ctx := createTestContext()
	node := &ast.IntegerLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Type: token.INT, Literal: "42"},
			},
		},
		Value: 42,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSink = eval.VisitIntegerLiteral(node, ctx)
	}
}

// BenchmarkVisitFloatLiteral tests the performance of evaluating float literals
func BenchmarkVisitFloatLiteral(b *testing.B) {
	eval := createTestEvaluator()
	ctx := createTestContext()
	node := &ast.FloatLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Type: token.FLOAT, Literal: "3.14"},
			},
		},
		Value: 3.14,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSink = eval.VisitFloatLiteral(node, ctx)
	}
}

// BenchmarkVisitStringLiteral tests the performance of evaluating string literals
func BenchmarkVisitStringLiteral(b *testing.B) {
	eval := createTestEvaluator()
	ctx := createTestContext()
	node := &ast.StringLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Type: token.STRING, Literal: "hello"},
			},
		},
		Value: "hello",
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSink = eval.VisitStringLiteral(node, ctx)
	}
}

// BenchmarkVisitBooleanLiteral tests the performance of evaluating boolean literals
func BenchmarkVisitBooleanLiteral(b *testing.B) {
	eval := createTestEvaluator()
	ctx := createTestContext()
	node := &ast.BooleanLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Type: token.TRUE, Literal: "true"},
			},
		},
		Value: true,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSink = eval.VisitBooleanLiteral(node, ctx)
	}
}

// ============================================================================
// Binary Operation Benchmarks - Test hot path performance
// ============================================================================

// BenchmarkVisitBinaryExpression_IntegerAdd tests integer addition (very hot path)
func BenchmarkVisitBinaryExpression_IntegerAdd(b *testing.B) {
	eval := createTestEvaluator()
	ctx := createTestContext()

	// 3 + 5
	node := &ast.BinaryExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Type: token.PLUS, Literal: "+"},
			},
		},
		Left: &ast.IntegerLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{Token: token.Token{Type: token.INT, Literal: "3"}},
			},
			Value: 3,
		},
		Operator: "+",
		Right: &ast.IntegerLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{Token: token.Token{Type: token.INT, Literal: "5"}},
			},
			Value: 5,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSink = eval.VisitBinaryExpression(node, ctx)
	}
}

// BenchmarkVisitBinaryExpression_IntegerMultiply tests integer multiplication
func BenchmarkVisitBinaryExpression_IntegerMultiply(b *testing.B) {
	eval := createTestEvaluator()
	ctx := createTestContext()

	// 6 * 7
	node := &ast.BinaryExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Type: token.ASTERISK, Literal: "*"},
			},
		},
		Left: &ast.IntegerLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{Token: token.Token{Type: token.INT, Literal: "6"}},
			},
			Value: 6,
		},
		Operator: "*",
		Right: &ast.IntegerLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{Token: token.Token{Type: token.INT, Literal: "7"}},
			},
			Value: 7,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSink = eval.VisitBinaryExpression(node, ctx)
	}
}

// BenchmarkVisitBinaryExpression_IntegerComparison tests integer comparison
func BenchmarkVisitBinaryExpression_IntegerComparison(b *testing.B) {
	eval := createTestEvaluator()
	ctx := createTestContext()

	// 42 > 10
	node := &ast.BinaryExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Type: token.IDENT, Literal: ">"},
			},
		},
		Left: &ast.IntegerLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{Token: token.Token{Type: token.INT, Literal: "42"}},
			},
			Value: 42,
		},
		Operator: ">",
		Right: &ast.IntegerLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{Token: token.Token{Type: token.INT, Literal: "10"}},
			},
			Value: 10,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSink = eval.VisitBinaryExpression(node, ctx)
	}
}

// BenchmarkVisitBinaryExpression_BooleanAnd tests boolean AND with short-circuit
func BenchmarkVisitBinaryExpression_BooleanAnd(b *testing.B) {
	eval := createTestEvaluator()
	ctx := createTestContext()

	// true and false
	node := &ast.BinaryExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Type: token.AND, Literal: "and"},
			},
		},
		Left: &ast.BooleanLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{Token: token.Token{Type: token.TRUE, Literal: "true"}},
			},
			Value: true,
		},
		Operator: "and",
		Right: &ast.BooleanLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{Token: token.Token{Type: token.FALSE, Literal: "false"}},
			},
			Value: false,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSink = eval.VisitBinaryExpression(node, ctx)
	}
}

// BenchmarkVisitBinaryExpression_StringConcat tests string concatenation
func BenchmarkVisitBinaryExpression_StringConcat(b *testing.B) {
	eval := createTestEvaluator()
	ctx := createTestContext()

	// 'Hello' + ' World'
	node := &ast.BinaryExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Type: token.PLUS, Literal: "+"},
			},
		},
		Left: &ast.StringLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{Token: token.Token{Type: token.STRING, Literal: "Hello"}},
			},
			Value: "Hello",
		},
		Operator: "+",
		Right: &ast.StringLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{Token: token.Token{Type: token.STRING, Literal: " World"}},
			},
			Value: " World",
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSink = eval.VisitBinaryExpression(node, ctx)
	}
}

// ============================================================================
// Unary Operation Benchmarks
// ============================================================================

// BenchmarkVisitUnaryExpression_IntegerNegation tests integer negation
func BenchmarkVisitUnaryExpression_IntegerNegation(b *testing.B) {
	eval := createTestEvaluator()
	ctx := createTestContext()

	// -42
	node := &ast.UnaryExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Type: token.MINUS, Literal: "-"},
			},
		},
		Operator: "-",
		Right: &ast.IntegerLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{Token: token.Token{Type: token.INT, Literal: "42"}},
			},
			Value: 42,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSink = eval.VisitUnaryExpression(node, ctx)
	}
}

// BenchmarkVisitUnaryExpression_BooleanNot tests boolean NOT
func BenchmarkVisitUnaryExpression_BooleanNot(b *testing.B) {
	eval := createTestEvaluator()
	ctx := createTestContext()

	// not true
	node := &ast.UnaryExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Type: token.NOT, Literal: "not"},
			},
		},
		Operator: "not",
		Right: &ast.BooleanLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{Token: token.Token{Type: token.TRUE, Literal: "true"}},
			},
			Value: true,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSink = eval.VisitUnaryExpression(node, ctx)
	}
}

// ============================================================================
// Variable Access Benchmarks - Test identifier lookups (hot path)
// ============================================================================

// BenchmarkVisitIdentifier tests variable lookup performance
// NOTE: Skipped - requires full interpreter adapter setup
func BenchmarkVisitIdentifier(b *testing.B) {
	b.Skip("Requires full interpreter adapter setup")
}

// ============================================================================
// Complex Expression Benchmarks
// ============================================================================

// BenchmarkComplexArithmetic tests nested arithmetic operations
func BenchmarkComplexArithmetic(b *testing.B) {
	eval := createTestEvaluator()
	ctx := createTestContext()

	// (3 + 5) * 2
	innerAdd := &ast.BinaryExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: token.Token{Type: token.PLUS, Literal: "+"}},
		},
		Left: &ast.IntegerLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{Token: token.Token{Type: token.INT, Literal: "3"}},
			},
			Value: 3,
		},
		Operator: "+",
		Right: &ast.IntegerLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{Token: token.Token{Type: token.INT, Literal: "5"}},
			},
			Value: 5,
		},
	}

	node := &ast.BinaryExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: token.Token{Type: token.ASTERISK, Literal: "*"}},
		},
		Left:     innerAdd,
		Operator: "*",
		Right: &ast.IntegerLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{Token: token.Token{Type: token.INT, Literal: "2"}},
			},
			Value: 2,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSink = eval.VisitBinaryExpression(node, ctx)
	}
}

// BenchmarkDeepNesting tests deeply nested expressions
func BenchmarkDeepNesting(b *testing.B) {
	eval := createTestEvaluator()
	ctx := createTestContext()

	// Build ((((1 + 2) + 3) + 4) + 5)
	var result ast.Expression = &ast.IntegerLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: token.Token{Type: token.INT, Literal: "1"}},
		},
		Value: 1,
	}

	for i := 2; i <= 5; i++ {
		result = &ast.BinaryExpression{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{Token: token.Token{Type: token.PLUS, Literal: "+"}},
			},
			Left:     result,
			Operator: "+",
			Right: &ast.IntegerLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{Token: token.Token{Type: token.INT, Literal: ""}},
				},
				Value: int64(i),
			},
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSink = eval.VisitBinaryExpression(result.(*ast.BinaryExpression), ctx)
	}
}

// BenchmarkWideExpression tests expressions with many operations at same level
func BenchmarkWideExpression(b *testing.B) {
	eval := createTestEvaluator()
	ctx := createTestContext()

	// Build 1 + 2 + 3 + 4 + 5 + 6 + 7 + 8 + 9 + 10
	var result ast.Expression = &ast.IntegerLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: token.Token{Type: token.INT, Literal: "1"}},
		},
		Value: 1,
	}

	for i := 2; i <= 10; i++ {
		result = &ast.BinaryExpression{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{Token: token.Token{Type: token.PLUS, Literal: "+"}},
			},
			Left:     result,
			Operator: "+",
			Right: &ast.IntegerLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{Token: token.Token{Type: token.INT, Literal: ""}},
				},
				Value: int64(i),
			},
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSink = eval.VisitBinaryExpression(result.(*ast.BinaryExpression), ctx)
	}
}

// ============================================================================
// Visitor Dispatch Overhead Benchmarks
// ============================================================================

// BenchmarkVisitorDispatchMixed tests the overhead of visitor method dispatch across different types
func BenchmarkVisitorDispatchMixed(b *testing.B) {
	eval := createTestEvaluator()
	ctx := createTestContext()

	// Create a variety of expression nodes
	intLit := &ast.IntegerLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: token.Token{Type: token.INT, Literal: "42"}},
		},
		Value: 42,
	}
	floatLit := &ast.FloatLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: token.Token{Type: token.FLOAT, Literal: "3.14"}},
		},
		Value: 3.14,
	}
	strLit := &ast.StringLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: token.Token{Type: token.STRING, Literal: "hello"}},
		},
		Value: "hello",
	}
	boolLit := &ast.BooleanLiteral{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: token.Token{Type: token.TRUE, Literal: "true"}},
		},
		Value: true,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSink = eval.VisitIntegerLiteral(intLit, ctx)
		benchSink = eval.VisitFloatLiteral(floatLit, ctx)
		benchSink = eval.VisitStringLiteral(strLit, ctx)
		benchSink = eval.VisitBooleanLiteral(boolLit, ctx)
	}
}

// ============================================================================
// Memory Allocation Benchmarks
// ============================================================================

// BenchmarkEvaluatorCreation tests the overhead of creating an evaluator
func BenchmarkEvaluatorCreation(b *testing.B) {
	config := DefaultConfig()
	typeSystem := interptypes.NewTypeSystem()
	unitRegistry := units.NewUnitRegistry([]string{"."})
	output := &bytes.Buffer{}
	refCountMgr := runtime.NewRefCountManager()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSink = NewEvaluator(typeSystem, output, config, unitRegistry, nil, refCountMgr)
	}
}

// BenchmarkContextCreation tests the overhead of creating an execution context
func BenchmarkContextCreation(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		env := runtime.NewEnvironment()
		benchSink = NewExecutionContext(env)
	}
}
