// Package ast defines the Abstract Syntax Tree node types for DWScript.
//
// The AST represents the hierarchical structure of a DWScript program after parsing.
// Each node type corresponds to a syntactic construct in the language.
//
// Node categories:
//   - Expressions: values that can be evaluated (literals, identifiers, binary ops, etc.)
//   - Statements: actions to be executed (assignments, declarations, control flow, etc.)
//   - Declarations: top-level constructs (functions, classes, types)
//
// All nodes implement the Node interface and can be converted to string form
// for debugging and testing purposes.
//
// Example node types:
//   - IntegerLiteral, StringLiteral, Identifier
//   - BinaryExpression, UnaryExpression
//   - VarDeclStatement, AssignmentStatement
//   - IfStatement, WhileStatement, ForStatement
//   - FunctionDecl, ClassDecl
package ast
