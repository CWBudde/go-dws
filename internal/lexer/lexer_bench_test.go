package lexer

import (
	"testing"
)

// BenchmarkNextToken benchmarks the overall tokenization throughput.
// This benchmark uses a realistic mix of tokens including operators.
func BenchmarkNextToken(b *testing.B) {
	input := `var x: Integer := 42;
var y: Float := 3.14;
function add(a, b: Integer): Integer;
begin
	result := a + b - 10 * 2 / 3;
end;
if x > 10 and y < 20 then
	x := x + 1
else
	y := y - 1;`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := New(input)
		for {
			tok := l.NextToken()
			if tok.Type == EOF {
				break
			}
		}
	}
}

// BenchmarkOperatorTokens benchmarks operator tokenization specifically.
// This focuses on the dispatch table performance for operators.
func BenchmarkOperatorTokens(b *testing.B) {
	// Heavy operator usage to stress-test the dispatch table
	input := `a + b - c * d / e % f
x == y != z < w > v <= u >= t
p && q || r ! s ? t
a += b -= c *= d /= e
i++ j-- k ** l
m << n >> o & p | q ^ r ~ s @ t`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := New(input)
		for {
			tok := l.NextToken()
			if tok.Type == EOF {
				break
			}
		}
	}
}

// BenchmarkSimpleArithmetic benchmarks simple arithmetic expressions.
// This is a common use case with lots of operator tokens.
func BenchmarkSimpleArithmetic(b *testing.B) {
	input := `1 + 2 - 3 * 4 / 5`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := New(input)
		for {
			tok := l.NextToken()
			if tok.Type == EOF {
				break
			}
		}
	}
}

// BenchmarkMixedTokens benchmarks a realistic mix of all token types.
func BenchmarkMixedTokens(b *testing.B) {
	input := `// Comment
var count: Integer := 0;
{ Block comment }
const PI = 3.14159;
type TPoint = record
	x, y: Float;
end;
function Distance(p1, p2: TPoint): Float;
begin
	result := Sqrt((p2.x - p1.x) ** 2 + (p2.y - p1.y) ** 2);
end;`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := New(input)
		for {
			tok := l.NextToken()
			if tok.Type == EOF {
				break
			}
		}
	}
}

// BenchmarkIdentifiersAndKeywords benchmarks identifier/keyword lookup.
// This tests the non-operator path to compare overall performance.
func BenchmarkIdentifiersAndKeywords(b *testing.B) {
	input := `begin end if then else while do for to repeat until
class interface function procedure var const type
private public protected virtual override abstract`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := New(input)
		for {
			tok := l.NextToken()
			if tok.Type == EOF {
				break
			}
		}
	}
}

// BenchmarkStringLiterals benchmarks string literal tokenization.
func BenchmarkStringLiterals(b *testing.B) {
	input := `'hello world' "foo bar" 'it''s a test' #13#10
'multi' 'part' #32 'concatenation'`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := New(input)
		for {
			tok := l.NextToken()
			if tok.Type == EOF {
				break
			}
		}
	}
}

// BenchmarkLargeFile benchmarks tokenizing a larger, more realistic file.
func BenchmarkLargeFile(b *testing.B) {
	// Simulate a larger source file (~500 tokens)
	input := ""
	for i := 0; i < 20; i++ {
		input += `var x: Integer := 42;
function Calculate(a, b: Integer): Float;
begin
	if a > b then
		result := a + b * 2.0
	else
		result := a - b / 2.0;
end;
`
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := New(input)
		for {
			tok := l.NextToken()
			if tok.Type == EOF {
				break
			}
		}
	}
}

// BenchmarkPeek benchmarks the Peek functionality.
func BenchmarkPeek(b *testing.B) {
	input := `var x: Integer := 42 + 10 - 5;`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := New(input)
		// Peek at first 5 tokens
		for j := 0; j < 5; j++ {
			_ = l.Peek(j)
		}
		// Then consume them
		for {
			tok := l.NextToken()
			if tok.Type == EOF {
				break
			}
		}
	}
}
