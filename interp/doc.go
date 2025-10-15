// Package interp implements the DWScript interpreter/runtime engine.
//
// The interpreter executes DWScript programs by walking the AST and
// evaluating expressions and executing statements. It maintains the
// runtime environment including:
//   - Variable storage (symbol tables/environments)
//   - Function call stack
//   - Object instances and class metadata
//   - Built-in functions
//
// The interpreter supports:
//   - Expression evaluation (arithmetic, logical, relational operations)
//   - Statement execution (assignments, control flow, function calls)
//   - Scope management (global, local, nested scopes)
//   - Object-oriented features (classes, methods, inheritance)
//   - Runtime type checking and error handling
//
// Example usage:
//
//	program := parser.ParseProgram()
//	interp := interp.New()
//	result := interp.Eval(program)
package interp
