// Package token defines constants representing the lexical tokens of DWScript
// and provides types for source code positions and tokens.
//
// This package is part of the public API for the go-dws DWScript implementation
// and is designed to work with the ast package for programmatic analysis and
// manipulation of DWScript source code.
//
// # Position Tracking
//
// The Position type represents a location in source code with line, column,
// and byte offset information. This is used throughout the AST to track the
// origin of each syntactic element.
//
//	pos := token.Position{Line: 1, Column: 5, Offset: 4}
//	fmt.Println(pos) // Output: 1:5
//
// # Token Types
//
// DWScript has over 150 token types, organized into categories:
//   - Identifiers and literals (IDENT, INT, FLOAT, STRING, CHAR)
//   - Keywords (BEGIN, END, IF, FOR, CLASS, etc.)
//   - Operators (PLUS, MINUS, ASSIGN, EQ, etc.)
//   - Delimiters (LPAREN, SEMICOLON, DOT, etc.)
//
// Keywords in DWScript are case-insensitive. The LookupIdent function
// performs case-insensitive keyword recognition:
//
//	tokenType := token.LookupIdent("BEGIN")  // Returns token.BEGIN
//	tokenType := token.LookupIdent("begin")  // Also returns token.BEGIN
//
// # Token Values
//
// The Token type combines a token type with its literal text and position:
//
//	tok := token.NewToken(token.IDENT, "myVar", pos)
//	fmt.Println(tok) // Output: IDENT("myVar") at 1:5
//
// # Helper Functions
//
// The package provides several helper functions:
//   - LookupIdent: Convert identifier strings to token types
//   - IsKeyword: Check if a string is a keyword
//   - GetKeywordLiteral: Get the canonical (lowercase) form of keywords
//
// # Integration with AST
//
// This package is designed to work seamlessly with the ast package.
// Every AST node contains position information from this package:
//
//	import (
//		"github.com/cwbudde/go-dws/pkg/token"
//		"github.com/cwbudde/go-dws/pkg/ast"
//	)
//
//	// AST nodes expose positions via Pos() and End() methods
//	var node ast.Node
//	start := node.Pos()  // Returns token.Position
//	end := node.End()    // Returns token.Position
package token
