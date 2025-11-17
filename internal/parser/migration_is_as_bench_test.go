package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// Task 2.2.13: Benchmarks for IS/AS/IMPLEMENTS expression migration
//
// Benchmark parseIsExpression, parseAsExpression, and parseImplementsExpression
// in traditional vs cursor mode.

// BenchmarkParseIsExpression_Traditional_Simple benchmarks simple 'is' check in traditional mode
func BenchmarkParseIsExpression_Traditional_Simple(b *testing.B) {
	source := "obj is TMyClass"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseIsExpression_Cursor_Simple benchmarks simple 'is' check in cursor mode
func BenchmarkParseIsExpression_Cursor_Simple(b *testing.B) {
	source := "obj is TMyClass"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseIsExpression_Traditional_Complex benchmarks complex 'is' expression in traditional mode
func BenchmarkParseIsExpression_Traditional_Complex(b *testing.B) {
	source := "GetContainer().Items[i] is TClass"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseIsExpression_Cursor_Complex benchmarks complex 'is' expression in cursor mode
func BenchmarkParseIsExpression_Cursor_Complex(b *testing.B) {
	source := "GetContainer().Items[i] is TClass"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseAsExpression_Traditional_Simple benchmarks simple 'as' cast in traditional mode
func BenchmarkParseAsExpression_Traditional_Simple(b *testing.B) {
	source := "obj as IMyInterface"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseAsExpression_Cursor_Simple benchmarks simple 'as' cast in cursor mode
func BenchmarkParseAsExpression_Cursor_Simple(b *testing.B) {
	source := "obj as IMyInterface"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseAsExpression_Traditional_Chained benchmarks chained 'as' cast in traditional mode
func BenchmarkParseAsExpression_Traditional_Chained(b *testing.B) {
	source := "(obj as IFoo).Method()"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseAsExpression_Cursor_Chained benchmarks chained 'as' cast in cursor mode
func BenchmarkParseAsExpression_Cursor_Chained(b *testing.B) {
	source := "(obj as IFoo).Method()"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseImplementsExpression_Traditional_Simple benchmarks simple 'implements' check in traditional mode
func BenchmarkParseImplementsExpression_Traditional_Simple(b *testing.B) {
	source := "obj implements IMyInterface"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseImplementsExpression_Cursor_Simple benchmarks simple 'implements' check in cursor mode
func BenchmarkParseImplementsExpression_Cursor_Simple(b *testing.B) {
	source := "obj implements IMyInterface"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseImplementsExpression_Traditional_Complex benchmarks complex 'implements' in traditional mode
func BenchmarkParseImplementsExpression_Traditional_Complex(b *testing.B) {
	source := "GetContainer().Items[i] implements IInterface"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseImplementsExpression_Cursor_Complex benchmarks complex 'implements' in cursor mode
func BenchmarkParseImplementsExpression_Cursor_Complex(b *testing.B) {
	source := "GetContainer().Items[i] implements IInterface"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkTypeOperatorsIntegration_Traditional benchmarks combined type operators in traditional mode
func BenchmarkTypeOperatorsIntegration_Traditional(b *testing.B) {
	source := "(obj is IFoo) and (obj as IFoo).Method()"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkTypeOperatorsIntegration_Cursor benchmarks combined type operators in cursor mode
func BenchmarkTypeOperatorsIntegration_Cursor(b *testing.B) {
	source := "(obj is IFoo) and (obj as IFoo).Method()"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}
