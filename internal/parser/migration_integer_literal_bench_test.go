package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// BenchmarkIntegerLiteral_Traditional benchmarks the traditional implementation
func BenchmarkIntegerLiteral_Traditional(b *testing.B) {
	source := "42"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.parseIntegerLiteralTraditional()
	}
}

// BenchmarkIntegerLiteral_Cursor benchmarks the cursor implementation
func BenchmarkIntegerLiteral_Cursor(b *testing.B) {
	source := "42"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.parseIntegerLiteralCursor()
	}
}

// BenchmarkIntegerLiteral_Dispatcher_Traditional benchmarks dispatcher in traditional mode
func BenchmarkIntegerLiteral_Dispatcher_Traditional(b *testing.B) {
	source := "42"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.parseIntegerLiteral()
	}
}

// BenchmarkIntegerLiteral_Dispatcher_Cursor benchmarks dispatcher in cursor mode
func BenchmarkIntegerLiteral_Dispatcher_Cursor(b *testing.B) {
	source := "42"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.parseIntegerLiteral()
	}
}

// BenchmarkIntegerLiteral_Hex_Traditional benchmarks hex parsing in traditional mode
func BenchmarkIntegerLiteral_Hex_Traditional(b *testing.B) {
	source := "$FF"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.parseIntegerLiteralTraditional()
	}
}

// BenchmarkIntegerLiteral_Hex_Cursor benchmarks hex parsing in cursor mode
func BenchmarkIntegerLiteral_Hex_Cursor(b *testing.B) {
	source := "$FF"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.parseIntegerLiteralCursor()
	}
}

// BenchmarkIntegerLiteral_Binary_Traditional benchmarks binary parsing in traditional mode
func BenchmarkIntegerLiteral_Binary_Traditional(b *testing.B) {
	source := "%1010"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.parseIntegerLiteralTraditional()
	}
}

// BenchmarkIntegerLiteral_Binary_Cursor benchmarks binary parsing in cursor mode
func BenchmarkIntegerLiteral_Binary_Cursor(b *testing.B) {
	source := "%1010"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.parseIntegerLiteralCursor()
	}
}

// BenchmarkIntegerLiteral_Large_Traditional benchmarks large number parsing in traditional mode
func BenchmarkIntegerLiteral_Large_Traditional(b *testing.B) {
	source := "9223372036854775807"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.parseIntegerLiteralTraditional()
	}
}

// BenchmarkIntegerLiteral_Large_Cursor benchmarks large number parsing in cursor mode
func BenchmarkIntegerLiteral_Large_Cursor(b *testing.B) {
	source := "9223372036854775807"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.parseIntegerLiteralCursor()
	}
}

// BenchmarkIntegerLiteral_Memory_Traditional measures memory allocations in traditional mode
func BenchmarkIntegerLiteral_Memory_Traditional(b *testing.B) {
	source := "42"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.parseIntegerLiteralTraditional()
	}
}

// BenchmarkIntegerLiteral_Memory_Cursor measures memory allocations in cursor mode
func BenchmarkIntegerLiteral_Memory_Cursor(b *testing.B) {
	source := "42"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.parseIntegerLiteralCursor()
	}
}
