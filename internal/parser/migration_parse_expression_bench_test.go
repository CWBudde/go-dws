package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// Benchmark parseExpression in traditional vs cursor mode.
// Goal: Verify cursor mode overhead is <15% (established threshold from Task 2.2.6).

// BenchmarkParseExpression_Traditional_SimpleLiteral benchmarks simple literal parsing in traditional mode.
func BenchmarkParseExpression_Traditional_SimpleLiteral(b *testing.B) {
	source := "42"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseExpression_Cursor_SimpleLiteral benchmarks simple literal parsing in cursor mode.
func BenchmarkParseExpression_Cursor_SimpleLiteral(b *testing.B) {
	source := "42"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseExpression_Traditional_BinaryOp benchmarks binary operator parsing in traditional mode.
func BenchmarkParseExpression_Traditional_BinaryOp(b *testing.B) {
	source := "3 + 5"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseExpression_Cursor_BinaryOp benchmarks binary operator parsing in cursor mode.
func BenchmarkParseExpression_Cursor_BinaryOp(b *testing.B) {
	source := "3 + 5"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseExpression_Traditional_Complex benchmarks complex expression parsing in traditional mode.
func BenchmarkParseExpression_Traditional_Complex(b *testing.B) {
	source := "2 + 3 * 4 - 5 / 2"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseExpression_Cursor_Complex benchmarks complex expression parsing in cursor mode.
func BenchmarkParseExpression_Cursor_Complex(b *testing.B) {
	source := "2 + 3 * 4 - 5 / 2"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseExpression_Traditional_Nested benchmarks nested parenthesized expressions in traditional mode.
func BenchmarkParseExpression_Traditional_Nested(b *testing.B) {
	source := "((2 + 3) * (4 - 1)) / 5"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseExpression_Cursor_Nested benchmarks nested parenthesized expressions in cursor mode.
func BenchmarkParseExpression_Cursor_Nested(b *testing.B) {
	source := "((2 + 3) * (4 - 1)) / 5"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseExpression_Traditional_Logical benchmarks logical expressions in traditional mode.
func BenchmarkParseExpression_Traditional_Logical(b *testing.B) {
	source := "x > 0 and y < 10 or z = 5"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseExpression_Cursor_Logical benchmarks logical expressions in cursor mode.
func BenchmarkParseExpression_Cursor_Logical(b *testing.B) {
	source := "x > 0 and y < 10 or z = 5"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseExpression_Traditional_ChainedOps benchmarks chained operations in traditional mode.
func BenchmarkParseExpression_Traditional_ChainedOps(b *testing.B) {
	source := "a + b + c + d + e + f + g + h"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseExpression_Cursor_ChainedOps benchmarks chained operations in cursor mode.
func BenchmarkParseExpression_Cursor_ChainedOps(b *testing.B) {
	source := "a + b + c + d + e + f + g + h"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseExpression_Traditional_NotInIsAs benchmarks "not in/is/as" in traditional mode.
func BenchmarkParseExpression_Traditional_NotInIsAs(b *testing.B) {
	source := "x not in mySet"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseExpression_Cursor_NotInIsAs benchmarks "not in/is/as" in cursor mode.
func BenchmarkParseExpression_Cursor_NotInIsAs(b *testing.B) {
	source := "x not in mySet"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseExpression_Traditional_MixedPrecedence benchmarks mixed precedence in traditional mode.
func BenchmarkParseExpression_Traditional_MixedPrecedence(b *testing.B) {
	source := "2 * 3 + 4 * 5 - 6 / 2 + 7 mod 3"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseExpression_Cursor_MixedPrecedence benchmarks mixed precedence in cursor mode.
func BenchmarkParseExpression_Cursor_MixedPrecedence(b *testing.B) {
	source := "2 * 3 + 4 * 5 - 6 / 2 + 7 mod 3"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseExpression_Traditional_Identifiers benchmarks identifier-heavy expressions in traditional mode.
func BenchmarkParseExpression_Traditional_Identifiers(b *testing.B) {
	source := "firstName + lastName + middleName + suffix"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseExpression_Cursor_Identifiers benchmarks identifier-heavy expressions in cursor mode.
func BenchmarkParseExpression_Cursor_Identifiers(b *testing.B) {
	source := "firstName + lastName + middleName + suffix"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseExpression_Traditional_Coalesce benchmarks coalesce operator in traditional mode.
func BenchmarkParseExpression_Traditional_Coalesce(b *testing.B) {
	source := "a ?? b ?? c ?? d"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseExpression_Cursor_Coalesce benchmarks coalesce operator in cursor mode.
func BenchmarkParseExpression_Cursor_Coalesce(b *testing.B) {
	source := "a ?? b ?? c ?? d"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseExpression_Traditional_Comparison benchmarks comparison expressions in traditional mode.
func BenchmarkParseExpression_Traditional_Comparison(b *testing.B) {
	source := "a < b and b <= c and c <> d and d > e and e >= f"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}

// BenchmarkParseExpression_Cursor_Comparison benchmarks comparison expressions in cursor mode.
func BenchmarkParseExpression_Cursor_Comparison(b *testing.B) {
	source := "a < b and b <= c and c <> d and d > e and e >= f"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			b.Fatal("parseExpression returned nil")
		}
	}
}
