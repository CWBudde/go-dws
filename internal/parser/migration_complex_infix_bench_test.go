package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// Task 2.2.11: Benchmarks for complex infix expression migration
//
// Benchmark parseCallExpression, parseMemberAccess, and parseIndexExpression
// in traditional vs cursor mode.

// BenchmarkParseCallExpression_Traditional_Simple benchmarks simple function call in traditional mode
func BenchmarkParseCallExpression_Traditional_Simple(b *testing.B) {
	source := "foo()"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseCallExpression_Cursor_Simple benchmarks simple function call in cursor mode
func BenchmarkParseCallExpression_Cursor_Simple(b *testing.B) {
	source := "foo()"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseCallExpression_Traditional_WithArgs benchmarks function call with arguments in traditional mode
func BenchmarkParseCallExpression_Traditional_WithArgs(b *testing.B) {
	source := "add(1, 2, 3, 4, 5)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseCallExpression_Cursor_WithArgs benchmarks function call with arguments in cursor mode
func BenchmarkParseCallExpression_Cursor_WithArgs(b *testing.B) {
	source := "add(1, 2, 3, 4, 5)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseCallExpression_Traditional_RecordLiteral benchmarks record literal in traditional mode
func BenchmarkParseCallExpression_Traditional_RecordLiteral(b *testing.B) {
	source := "Point(x: 10, y: 20)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseCallExpression_Cursor_RecordLiteral benchmarks record literal in cursor mode
func BenchmarkParseCallExpression_Cursor_RecordLiteral(b *testing.B) {
	source := "Point(x: 10, y: 20)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseMemberAccess_Traditional_Field benchmarks field access in traditional mode
func BenchmarkParseMemberAccess_Traditional_Field(b *testing.B) {
	source := "obj.field"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseMemberAccess_Cursor_Field benchmarks field access in cursor mode
func BenchmarkParseMemberAccess_Cursor_Field(b *testing.B) {
	source := "obj.field"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseMemberAccess_Traditional_Method benchmarks method call in traditional mode
func BenchmarkParseMemberAccess_Traditional_Method(b *testing.B) {
	source := "obj.Method(1, 2, 3)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseMemberAccess_Cursor_Method benchmarks method call in cursor mode
func BenchmarkParseMemberAccess_Cursor_Method(b *testing.B) {
	source := "obj.Method(1, 2, 3)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseMemberAccess_Traditional_ClassCreate benchmarks class creation in traditional mode
func BenchmarkParseMemberAccess_Traditional_ClassCreate(b *testing.B) {
	source := "TMyClass.Create()"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseMemberAccess_Cursor_ClassCreate benchmarks class creation in cursor mode
func BenchmarkParseMemberAccess_Cursor_ClassCreate(b *testing.B) {
	source := "TMyClass.Create()"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseMemberAccess_Traditional_Chain benchmarks chained member access in traditional mode
func BenchmarkParseMemberAccess_Traditional_Chain(b *testing.B) {
	source := "obj.field1.field2.field3"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseMemberAccess_Cursor_Chain benchmarks chained member access in cursor mode
func BenchmarkParseMemberAccess_Cursor_Chain(b *testing.B) {
	source := "obj.field1.field2.field3"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseIndexExpression_Traditional_Simple benchmarks simple indexing in traditional mode
func BenchmarkParseIndexExpression_Traditional_Simple(b *testing.B) {
	source := "arr[0]"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseIndexExpression_Cursor_Simple benchmarks simple indexing in cursor mode
func BenchmarkParseIndexExpression_Cursor_Simple(b *testing.B) {
	source := "arr[0]"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseIndexExpression_Traditional_MultiDim benchmarks multi-dimensional indexing in traditional mode
func BenchmarkParseIndexExpression_Traditional_MultiDim(b *testing.B) {
	source := "grid[x, y, z]"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseIndexExpression_Cursor_MultiDim benchmarks multi-dimensional indexing in cursor mode
func BenchmarkParseIndexExpression_Cursor_MultiDim(b *testing.B) {
	source := "grid[x, y, z]"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseIndexExpression_Traditional_Expression benchmarks index with expression in traditional mode
func BenchmarkParseIndexExpression_Traditional_Expression(b *testing.B) {
	source := "arr[i * 2 + j]"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkParseIndexExpression_Cursor_Expression benchmarks index with expression in cursor mode
func BenchmarkParseIndexExpression_Cursor_Expression(b *testing.B) {
	source := "arr[i * 2 + j]"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkComplexInfix_Traditional_Integration benchmarks complex combined operations in traditional mode
func BenchmarkComplexInfix_Traditional_Integration(b *testing.B) {
	source := "obj.GetArray()[0].Method(1, 2)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		_ = p.ParseProgram()
	}
}

// BenchmarkComplexInfix_Cursor_Integration benchmarks complex combined operations in cursor mode
func BenchmarkComplexInfix_Cursor_Integration(b *testing.B) {
	source := "obj.GetArray()[0].Method(1, 2)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewCursorParser(lexer.New(source))
		_ = p.ParseProgram()
	}
}
