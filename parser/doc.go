// Package parser implements the DWScript parser.
//
// The parser consumes tokens from the lexer and builds an Abstract Syntax Tree (AST)
// representing the structure of the DWScript program. It implements a Pratt parser
// (top-down operator precedence parser) for expressions and recursive descent for statements.
//
// The parser handles:
//   - Variable declarations
//   - Assignments
//   - Expressions (arithmetic, logical, relational)
//   - Control flow (if/else, while, for, repeat/until, case)
//   - Functions and procedures
//   - Classes and object-oriented constructs
//   - Type annotations
//
// Example usage:
//
//	l := lexer.New(input)
//	p := parser.New(l)
//	program := p.ParseProgram()
//	if len(p.Errors()) > 0 {
//	    // handle errors
//	}
package parser
