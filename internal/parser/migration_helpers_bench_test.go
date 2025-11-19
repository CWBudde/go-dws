package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// Task 2.2.10 Phase 6: Benchmarks for expression helper migration
//
// Benchmark parseCallOrRecordLiteral and helper functions in traditional vs cursor mode.
// Goal: Verify cursor mode overhead is <15% (established threshold from Task 2.2.6).

// BenchmarkParseExpressionList_Traditional_Empty benchmarks empty list parsing in traditional mode
func BenchmarkParseExpressionList_Traditional_Empty(b *testing.B) {
	source := "Foo()"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		expr := p.parseExpressionCursor(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseExpressionList_Cursor_Empty benchmarks empty list parsing in cursor mode
func BenchmarkParseExpressionList_Cursor_Empty(b *testing.B) {
	source := "Foo()"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		expr := p.parseExpressionCursor(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseExpressionList_Traditional_SingleArg benchmarks single argument in traditional mode
func BenchmarkParseExpressionList_Traditional_SingleArg(b *testing.B) {
	source := "Foo(42)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		expr := p.parseExpressionCursor(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseExpressionList_Cursor_SingleArg benchmarks single argument in cursor mode
func BenchmarkParseExpressionList_Cursor_SingleArg(b *testing.B) {
	source := "Foo(42)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		expr := p.parseExpressionCursor(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseExpressionList_Traditional_MultipleArgs benchmarks multiple arguments in traditional mode
func BenchmarkParseExpressionList_Traditional_MultipleArgs(b *testing.B) {
	source := "Calculate(1, 2, 3, 4, 5)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		expr := p.parseExpressionCursor(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseExpressionList_Cursor_MultipleArgs benchmarks multiple arguments in cursor mode
func BenchmarkParseExpressionList_Cursor_MultipleArgs(b *testing.B) {
	source := "Calculate(1, 2, 3, 4, 5)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		expr := p.parseExpressionCursor(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseCallOrRecordLiteral_Traditional_FunctionCall benchmarks function call in traditional mode
func BenchmarkParseCallOrRecordLiteral_Traditional_FunctionCall(b *testing.B) {
	source := "Add(x, y)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		expr := p.parseExpressionCursor(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseCallOrRecordLiteral_Cursor_FunctionCall benchmarks function call in cursor mode
func BenchmarkParseCallOrRecordLiteral_Cursor_FunctionCall(b *testing.B) {
	source := "Add(x, y)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		expr := p.parseExpressionCursor(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseCallOrRecordLiteral_Traditional_RecordLiteral benchmarks record literal in traditional mode
func BenchmarkParseCallOrRecordLiteral_Traditional_RecordLiteral(b *testing.B) {
	source := "Point(x: 10, y: 20)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		expr := p.parseExpressionCursor(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseCallOrRecordLiteral_Cursor_RecordLiteral benchmarks record literal in cursor mode
func BenchmarkParseCallOrRecordLiteral_Cursor_RecordLiteral(b *testing.B) {
	source := "Point(x: 10, y: 20)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		expr := p.parseExpressionCursor(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseCallOrRecordLiteral_Traditional_ComplexCall benchmarks complex function call in traditional mode
func BenchmarkParseCallOrRecordLiteral_Traditional_ComplexCall(b *testing.B) {
	source := "Calculate(2 + 3, x * y, z / 2)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		expr := p.parseExpressionCursor(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseCallOrRecordLiteral_Cursor_ComplexCall benchmarks complex function call in cursor mode
func BenchmarkParseCallOrRecordLiteral_Cursor_ComplexCall(b *testing.B) {
	source := "Calculate(2 + 3, x * y, z / 2)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		expr := p.parseExpressionCursor(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseCallOrRecordLiteral_Traditional_ComplexRecord benchmarks complex record in traditional mode
func BenchmarkParseCallOrRecordLiteral_Traditional_ComplexRecord(b *testing.B) {
	source := "Complex(real: 2 + 3, imag: x * y, flag: true)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		expr := p.parseExpressionCursor(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseCallOrRecordLiteral_Cursor_ComplexRecord benchmarks complex record in cursor mode
func BenchmarkParseCallOrRecordLiteral_Cursor_ComplexRecord(b *testing.B) {
	source := "Complex(real: 2 + 3, imag: x * y, flag: true)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		expr := p.parseExpressionCursor(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseCallOrRecordLiteral_Traditional_NestedCalls benchmarks nested calls in traditional mode
func BenchmarkParseCallOrRecordLiteral_Traditional_NestedCalls(b *testing.B) {
	source := "Outer(Inner(1, 2), Middle(Inner(3, 4)))"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		expr := p.parseExpressionCursor(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseCallOrRecordLiteral_Cursor_NestedCalls benchmarks nested calls in cursor mode
func BenchmarkParseCallOrRecordLiteral_Cursor_NestedCalls(b *testing.B) {
	source := "Outer(Inner(1, 2), Middle(Inner(3, 4)))"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		expr := p.parseExpressionCursor(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseCallOrRecordLiteral_Traditional_LargeArgList benchmarks large argument list in traditional mode
func BenchmarkParseCallOrRecordLiteral_Traditional_LargeArgList(b *testing.B) {
	source := "Many(a, b, c, d, e, f, g, h, i, j)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		expr := p.parseExpressionCursor(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseCallOrRecordLiteral_Cursor_LargeArgList benchmarks large argument list in cursor mode
func BenchmarkParseCallOrRecordLiteral_Cursor_LargeArgList(b *testing.B) {
	source := "Many(a, b, c, d, e, f, g, h, i, j)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		expr := p.parseExpressionCursor(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}
