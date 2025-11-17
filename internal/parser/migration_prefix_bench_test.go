package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// Task 2.2.12: Benchmarks for prefix expression handler migration
//
// Benchmark parsePrefixExpression, parseGroupedExpression, parseArrayLiteral,
// and simple literal handlers in traditional vs cursor mode.

// BenchmarkParsePrefixExpression_Traditional_Minus benchmarks minus prefix in traditional mode
func BenchmarkParsePrefixExpression_Traditional_Minus(b *testing.B) {
	source := "-x"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParsePrefixExpression_Cursor_Minus benchmarks minus prefix in cursor mode
func BenchmarkParsePrefixExpression_Cursor_Minus(b *testing.B) {
	source := "-x"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParsePrefixExpression_Traditional_Not benchmarks not prefix in traditional mode
func BenchmarkParsePrefixExpression_Traditional_Not(b *testing.B) {
	source := "not flag"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParsePrefixExpression_Cursor_Not benchmarks not prefix in cursor mode
func BenchmarkParsePrefixExpression_Cursor_Not(b *testing.B) {
	source := "not flag"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParsePrefixExpression_Traditional_Nested benchmarks nested prefix in traditional mode
func BenchmarkParsePrefixExpression_Traditional_Nested(b *testing.B) {
	source := "-(x + y)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParsePrefixExpression_Cursor_Nested benchmarks nested prefix in cursor mode
func BenchmarkParsePrefixExpression_Cursor_Nested(b *testing.B) {
	source := "-(x + y)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseGroupedExpression_Traditional_Simple benchmarks simple grouped expression in traditional mode
func BenchmarkParseGroupedExpression_Traditional_Simple(b *testing.B) {
	source := "(x)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseGroupedExpression_Cursor_Simple benchmarks simple grouped expression in cursor mode
func BenchmarkParseGroupedExpression_Cursor_Simple(b *testing.B) {
	source := "(x)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseGroupedExpression_Traditional_Complex benchmarks complex grouped expression in traditional mode
func BenchmarkParseGroupedExpression_Traditional_Complex(b *testing.B) {
	source := "((x + y) * (z - w))"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseGroupedExpression_Cursor_Complex benchmarks complex grouped expression in cursor mode
func BenchmarkParseGroupedExpression_Cursor_Complex(b *testing.B) {
	source := "((x + y) * (z - w))"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseGroupedExpression_Traditional_Empty benchmarks empty parens in traditional mode
func BenchmarkParseGroupedExpression_Traditional_Empty(b *testing.B) {
	source := "()"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseGroupedExpression_Cursor_Empty benchmarks empty parens in cursor mode
func BenchmarkParseGroupedExpression_Cursor_Empty(b *testing.B) {
	source := "()"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseArrayLiteral_Traditional_Empty benchmarks empty array in traditional mode
func BenchmarkParseArrayLiteral_Traditional_Empty(b *testing.B) {
	source := "[]"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseArrayLiteral_Cursor_Empty benchmarks empty array in cursor mode
func BenchmarkParseArrayLiteral_Cursor_Empty(b *testing.B) {
	source := "[]"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseArrayLiteral_Traditional_Simple benchmarks simple array in traditional mode
func BenchmarkParseArrayLiteral_Traditional_Simple(b *testing.B) {
	source := "[1, 2, 3]"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseArrayLiteral_Cursor_Simple benchmarks simple array in cursor mode
func BenchmarkParseArrayLiteral_Cursor_Simple(b *testing.B) {
	source := "[1, 2, 3]"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseArrayLiteral_Traditional_Range benchmarks range array in traditional mode
func BenchmarkParseArrayLiteral_Traditional_Range(b *testing.B) {
	source := "[1..10]"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseArrayLiteral_Cursor_Range benchmarks range array in cursor mode
func BenchmarkParseArrayLiteral_Cursor_Range(b *testing.B) {
	source := "[1..10]"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseArrayLiteral_Traditional_Expressions benchmarks array with expressions in traditional mode
func BenchmarkParseArrayLiteral_Traditional_Expressions(b *testing.B) {
	source := "[x + 1, y * 2, z - 3]"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseArrayLiteral_Cursor_Expressions benchmarks array with expressions in cursor mode
func BenchmarkParseArrayLiteral_Cursor_Expressions(b *testing.B) {
	source := "[x + 1, y * 2, z - 3]"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseSimpleLiterals_Traditional_Nil benchmarks nil literal in traditional mode
func BenchmarkParseSimpleLiterals_Traditional_Nil(b *testing.B) {
	source := "nil"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseSimpleLiterals_Cursor_Nil benchmarks nil literal in cursor mode
func BenchmarkParseSimpleLiterals_Cursor_Nil(b *testing.B) {
	source := "nil"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseSimpleLiterals_Traditional_Null benchmarks Null identifier in traditional mode
func BenchmarkParseSimpleLiterals_Traditional_Null(b *testing.B) {
	source := "Null"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseSimpleLiterals_Cursor_Null benchmarks Null identifier in cursor mode
func BenchmarkParseSimpleLiterals_Cursor_Null(b *testing.B) {
	source := "Null"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseSimpleLiterals_Traditional_Char benchmarks char literal in traditional mode
func BenchmarkParseSimpleLiterals_Traditional_Char(b *testing.B) {
	source := "#65"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseSimpleLiterals_Cursor_Char benchmarks char literal in cursor mode
func BenchmarkParseSimpleLiterals_Cursor_Char(b *testing.B) {
	source := "#65"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkPrefixIntegration_Traditional benchmarks complex integration in traditional mode
func BenchmarkPrefixIntegration_Traditional(b *testing.B) {
	source := "[(x + y), not flag, -count, nil]"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkPrefixIntegration_Cursor benchmarks complex integration in cursor mode
func BenchmarkPrefixIntegration_Cursor(b *testing.B) {
	source := "[(x + y), not flag, -count, nil]"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}
