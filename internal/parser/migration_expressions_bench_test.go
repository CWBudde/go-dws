package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// Benchmark_Identifier_Traditional benchmarks identifier parsing in traditional mode
func Benchmark_Identifier_Traditional(b *testing.B) {
	source := "myVariableName"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.parseIdentifier()
	}
}

// Benchmark_Identifier_Cursor benchmarks identifier parsing in cursor mode
func Benchmark_Identifier_Cursor(b *testing.B) {
	source := "myVariableName"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.parseIdentifier()
	}
}

// Benchmark_FloatLiteral_Traditional benchmarks float parsing in traditional mode
func Benchmark_FloatLiteral_Traditional(b *testing.B) {
	source := "3.14159265358979"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.parseFloatLiteral()
	}
}

// Benchmark_FloatLiteral_Cursor benchmarks float parsing in cursor mode
func Benchmark_FloatLiteral_Cursor(b *testing.B) {
	source := "3.14159265358979"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.parseFloatLiteral()
	}
}

// Benchmark_FloatLiteral_Scientific_Traditional benchmarks scientific notation in traditional mode
func Benchmark_FloatLiteral_Scientific_Traditional(b *testing.B) {
	source := "1.23456789e10"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.parseFloatLiteral()
	}
}

// Benchmark_FloatLiteral_Scientific_Cursor benchmarks scientific notation in cursor mode
func Benchmark_FloatLiteral_Scientific_Cursor(b *testing.B) {
	source := "1.23456789e10"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.parseFloatLiteral()
	}
}

// Benchmark_StringLiteral_Traditional benchmarks string parsing in traditional mode
func Benchmark_StringLiteral_Traditional(b *testing.B) {
	source := "'hello world this is a test string'"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.parseStringLiteral()
	}
}

// Benchmark_StringLiteral_Cursor benchmarks string parsing in cursor mode
func Benchmark_StringLiteral_Cursor(b *testing.B) {
	source := "'hello world this is a test string'"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.parseStringLiteral()
	}
}

// Benchmark_StringLiteral_Escaped_Traditional benchmarks escaped string in traditional mode
func Benchmark_StringLiteral_Escaped_Traditional(b *testing.B) {
	source := "'it''s a string with ''escaped'' quotes'"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.parseStringLiteral()
	}
}

// Benchmark_StringLiteral_Escaped_Cursor benchmarks escaped string in cursor mode
func Benchmark_StringLiteral_Escaped_Cursor(b *testing.B) {
	source := "'it''s a string with ''escaped'' quotes'"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.parseStringLiteral()
	}
}

// Benchmark_BooleanLiteral_Traditional benchmarks boolean parsing in traditional mode
func Benchmark_BooleanLiteral_Traditional(b *testing.B) {
	source := "true"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.parseBooleanLiteral()
	}
}

// Benchmark_BooleanLiteral_Cursor benchmarks boolean parsing in cursor mode
func Benchmark_BooleanLiteral_Cursor(b *testing.B) {
	source := "true"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.parseBooleanLiteral()
	}
}

// Benchmark_AllExpressions_Traditional benchmarks all expression types in traditional mode
func Benchmark_AllExpressions_Traditional(b *testing.B) {
	tests := []struct {
		name   string
		source string
	}{
		{"identifier", "myVar"},
		{"integer", "42"},
		{"float", "3.14"},
		{"string", "'hello'"},
		{"boolean", "true"},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				p := New(lexer.New(tt.source))
				switch tt.name {
				case "identifier":
					_ = p.parseIdentifier()
				case "integer":
					_ = p.parseIntegerLiteral()
				case "float":
					_ = p.parseFloatLiteral()
				case "string":
					_ = p.parseStringLiteral()
				case "boolean":
					_ = p.parseBooleanLiteral()
				}
			}
		})
	}
}

// Benchmark_AllExpressions_Cursor benchmarks all expression types in cursor mode
func Benchmark_AllExpressions_Cursor(b *testing.B) {
	tests := []struct {
		name   string
		source string
	}{
		{"identifier", "myVar"},
		{"integer", "42"},
		{"float", "3.14"},
		{"string", "'hello'"},
		{"boolean", "true"},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				p := NewCursorParser(lexer.New(tt.source))
				switch tt.name {
				case "identifier":
					_ = p.parseIdentifier()
				case "integer":
					_ = p.parseIntegerLiteral()
				case "float":
					_ = p.parseFloatLiteral()
				case "string":
					_ = p.parseStringLiteral()
				case "boolean":
					_ = p.parseBooleanLiteral()
				}
			}
		})
	}
}
