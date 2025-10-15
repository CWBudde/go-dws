// Package lexer implements the lexical analyzer (tokenizer) for DWScript.
//
// The lexer converts DWScript source code text into a stream of tokens
// that can be consumed by the parser. It handles:
//   - Keywords (begin, end, if, while, var, function, class, etc.)
//   - Operators (+, -, *, /, :=, =, <>, etc.)
//   - Literals (integers, floats, strings, booleans)
//   - Identifiers
//   - Comments (both { } and (* *) style)
//   - Whitespace and newlines
//
// Example usage:
//
//	input := "var x: Integer := 42;"
//	l := lexer.New(input)
//	for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
//	    fmt.Printf("%s\n", tok)
//	}
package lexer
