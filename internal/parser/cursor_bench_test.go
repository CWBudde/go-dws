package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/token"
)

// BenchmarkCursor_Creation benchmarks cursor creation
func BenchmarkCursor_Creation(b *testing.B) {
	source := "var x: Integer := 42;"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := lexer.New(source)
		_ = NewTokenCursor(l)
	}
}

// BenchmarkCursor_Advance benchmarks advancing the cursor
func BenchmarkCursor_Advance(b *testing.B) {
	source := "var x: Integer := 42;"
	cursor := newCursorFromSource(source)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := cursor
		for !c.IsEOF() {
			c = c.Advance()
		}
	}
}

// BenchmarkCursor_Peek benchmarks peeking ahead
func BenchmarkCursor_Peek(b *testing.B) {
	source := "var x: Integer := 42;"
	cursor := newCursorFromSource(source)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cursor.Peek(1)
		_ = cursor.Peek(2)
		_ = cursor.Peek(3)
	}
}

// BenchmarkCursor_PeekFar benchmarks peeking far ahead
func BenchmarkCursor_PeekFar(b *testing.B) {
	source := "var x: Integer := 42; var y: String := 'hello'; var z: Float := 3.14;"
	cursor := newCursorFromSource(source)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cursor.Peek(10)
		_ = cursor.Peek(20)
	}
}

// BenchmarkCursor_Is benchmarks the Is() method
func BenchmarkCursor_Is(b *testing.B) {
	source := "var x: Integer := 42;"
	cursor := newCursorFromSource(source)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cursor.Is(token.VAR)
		_ = cursor.Is(token.CONST)
		_ = cursor.Is(token.INT)
	}
}

// BenchmarkCursor_IsAny benchmarks the IsAny() method
func BenchmarkCursor_IsAny(b *testing.B) {
	source := "var x: Integer := 42;"
	cursor := newCursorFromSource(source)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cursor.IsAny(token.VAR, token.CONST, token.TYPE)
	}
}

// BenchmarkCursor_MarkResetTo benchmarks backtracking
func BenchmarkCursor_MarkResetTo(b *testing.B) {
	source := "var x: Integer := 42;"
	cursor := newCursorFromSource(source)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mark := cursor.Mark()
		c := cursor.Advance().Advance().Advance()
		c = c.ResetTo(mark)
	}
}

// BenchmarkCursor_Clone benchmarks cloning
func BenchmarkCursor_Clone(b *testing.B) {
	source := "var x: Integer := 42;"
	cursor := newCursorFromSource(source)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cursor.Clone()
	}
}

// BenchmarkCursor_FullParse simulates a typical parsing scenario with cursor
func BenchmarkCursor_FullParse(b *testing.B) {
	source := `
		var x: Integer := 42;
		var y: String := "hello";
		var z: Float := 3.14;

		function Add(a, b: Integer): Integer;
		begin
		  Result := a + b;
		end;
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cursor := newCursorFromSource(source)

		// Simulate typical parsing operations
		for !cursor.IsEOF() {
			// Check current token
			_ = cursor.Is(token.VAR)

			// Peek ahead
			_ = cursor.Peek(1)

			// Advance
			cursor = cursor.Advance()
		}
	}
}

// BenchmarkTraditional_FullParse benchmarks the traditional approach for comparison
func BenchmarkTraditional_FullParse(b *testing.B) {
	source := `
		var x: Integer := 42;
		var y: String := "hello";
		var z: Float := 3.14;

		function Add(a, b: Integer): Integer;
		begin
		  Result := a + b;
		end;
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))

		// Simulate typical parsing operations
		for p.cursor.Current().Type != token.EOF {
			// Check current token
			_ = p.cursor.Is(token.VAR)

			// Check peek token
			_ = p.cursor.PeekIs(1, token.IDENT)

			// Advance
			p.nextToken()
		}
	}
}

// BenchmarkCursor_Expect benchmarks the Expect() method
func BenchmarkCursor_Expect(b *testing.B) {
	source := "var x: Integer := 42;"
	cursor := newCursorFromSource(source)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := cursor
		c, _ = c.Expect(token.VAR)
		c, _ = c.Expect(token.IDENT)
		c, _ = c.Expect(token.COLON)
	}
}

// BenchmarkCursor_Skip benchmarks the Skip() method
func BenchmarkCursor_Skip(b *testing.B) {
	source := "var x: Integer := 42;"
	cursor := newCursorFromSource(source)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := cursor
		c, _ = c.Skip(token.VAR)
		c, _ = c.Skip(token.IDENT)
		c, _ = c.Skip(token.COLON)
	}
}

// BenchmarkCursor_AdvanceN benchmarks the AdvanceN() method
func BenchmarkCursor_AdvanceN(b *testing.B) {
	source := "var x: Integer := 42; var y: String := 'hello';"
	cursor := newCursorFromSource(source)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cursor.AdvanceN(5)
	}
}

// BenchmarkCursor_PeekIs benchmarks the PeekIs() method
func BenchmarkCursor_PeekIs(b *testing.B) {
	source := "var x: Integer := 42;"
	cursor := newCursorFromSource(source)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cursor.PeekIs(1, token.IDENT)
		_ = cursor.PeekIs(2, token.COLON)
		_ = cursor.PeekIs(3, token.IDENT)
	}
}

// BenchmarkCursor_NavigationPattern benchmarks a common navigation pattern
func BenchmarkCursor_NavigationPattern(b *testing.B) {
	source := "var x: Integer := 42;"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cursor := newCursorFromSource(source)

		// Typical pattern: check, advance, peek, check
		if cursor.Is(token.VAR) {
			cursor = cursor.Advance()
			if cursor.PeekIs(1, token.COLON) {
				cursor = cursor.Advance()
				cursor = cursor.Advance()
			}
		}
	}
}

// BenchmarkTraditional_NavigationPattern benchmarks the traditional approach
func BenchmarkTraditional_NavigationPattern(b *testing.B) {
	source := "var x: Integer := 42;"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))

		// Typical pattern: check, advance, peek, check
		if p.cursor.Is(token.VAR) {
			p.nextToken()
			if p.cursor.PeekIs(1, token.COLON) {
				p.nextToken()
				p.nextToken()
			}
		}
	}
}

// BenchmarkCursor_BacktrackingScenario benchmarks a backtracking scenario
func BenchmarkCursor_BacktrackingScenario(b *testing.B) {
	source := "var x: Integer := 42; var y: String := 'hello';"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cursor := newCursorFromSource(source)

		// Try parse first pattern
		mark1 := cursor.Mark()
		cursor = cursor.Advance().Advance().Advance()

		// Backtrack
		cursor = cursor.ResetTo(mark1)

		// Try parse second pattern
		mark2 := cursor.Mark()
		cursor = cursor.Advance().Advance()

		// Backtrack
		cursor = cursor.ResetTo(mark2)
	}
}

// BenchmarkTraditional_BacktrackingScenario benchmarks traditional backtracking
func BenchmarkTraditional_BacktrackingScenario(b *testing.B) {
	source := "var x: Integer := 42; var y: String := 'hello';"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))

		// Try parse first pattern
		state1 := p.saveState()
		p.nextToken()
		p.nextToken()
		p.nextToken()

		// Backtrack
		p.restoreState(state1)

		// Try parse second pattern
		state2 := p.saveState()
		p.nextToken()
		p.nextToken()

		// Backtrack
		p.restoreState(state2)
	}
}

// BenchmarkCursor_Memory measures memory allocations
func BenchmarkCursor_Memory(b *testing.B) {
	source := "var x: Integer := 42; var y: String := 'hello'; var z: Float := 3.14;"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cursor := newCursorFromSource(source)
		for j := 0; j < 10; j++ {
			cursor = cursor.Advance()
			_ = cursor.Peek(5)
		}
	}
}

// BenchmarkTraditional_Memory measures memory allocations for traditional approach
func BenchmarkTraditional_Memory(b *testing.B) {
	source := "var x: Integer := 42; var y: String := 'hello'; var z: Float := 3.14;"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New(lexer.New(source))
		for j := 0; j < 10; j++ {
			p.nextToken()
			_ = p.peek(5)
		}
	}
}
