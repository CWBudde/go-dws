// Package ast defines the Abstract Syntax Tree node types for DWScript.
//
// This package is part of the public API for the go-dws DWScript implementation
// and enables programmatic analysis, transformation, and tooling for DWScript
// source code.
//
// # Overview
//
// The AST represents the hierarchical structure of a DWScript program after parsing.
// Each node type corresponds to a syntactic construct in the language. All nodes
// implement the Node interface, which provides methods for accessing source
// positions and string representations.
//
// # Node Hierarchy
//
// The AST is organized into three main categories:
//
//   - Expressions: Nodes that evaluate to values (literals, identifiers, operators, etc.)
//   - Statements: Nodes that perform actions (assignments, declarations, control flow, etc.)
//   - Declarations: Top-level constructs (functions, classes, types, etc.)
//
// # Core Interfaces
//
// Node: Base interface for all AST nodes
//
//	type Node interface {
//	    TokenLiteral() string
//	    String() string
//	    Pos() token.Position
//	    End() token.Position
//	}
//
// Expression: Interface for value-producing nodes
//
//	type Expression interface {
//	    Node
//	    expressionNode()  // Marker method
//	}
//
// Statement: Interface for action-performing nodes
//
//	type Statement interface {
//	    Node
//	    statementNode()  // Marker method
//	}
//
// # Example Usage
//
// Walking the AST to find all function declarations:
//
//	import (
//	    "github.com/cwbudde/go-dws/pkg/ast"
//	    "github.com/cwbudde/go-dws/pkg/dwscript"
//	)
//
//	// Parse a program
//	engine := dwscript.New()
//	program, _ := engine.Compile("function Add(a, b: Integer): Integer; begin Result := a + b; end;")
//	tree := program.AST()
//
//	// Walk the AST
//	for _, stmt := range tree.Statements {
//	    if funcDecl, ok := stmt.(*ast.FunctionDecl); ok {
//	        fmt.Printf("Found function: %s\n", funcDecl.Name.Value)
//	    }
//	}
//
// # AST Node Types
//
// Literals:
//   - IntegerLiteral, FloatLiteral, StringLiteral, CharLiteral
//   - BooleanLiteral, NilLiteral
//
// Identifiers and Types:
//   - Identifier
//   - TypeAnnotation, TypeExpression
//
// Expressions:
//   - BinaryExpression, UnaryExpression, GroupedExpression
//   - CallExpression, IndexExpression
//   - ArrayLiteralExpression, SetLiteral
//   - NewExpression, MemberAccessExpression
//   - LambdaExpression, AddressOfExpression
//
// Statements:
//   - VarDeclStatement, AssignmentStatement
//   - ExpressionStatement, BlockStatement
//   - IfStatement, WhileStatement, RepeatStatement
//   - ForStatement, ForInStatement
//   - CaseStatement, BreakStatement, ContinueStatement, ExitStatement
//   - TryStatement, RaiseStatement
//   - ReturnStatement
//
// Declarations:
//   - FunctionDecl, ConstDecl
//   - ClassDecl, RecordDecl, InterfaceDecl
//   - EnumDecl, ArrayDecl, SetDecl
//   - PropertyDecl, OperatorDecl, HelperDecl
//   - UnitDeclaration
//
// # Position Information
//
// Every AST node tracks its source location via Pos() and End() methods,
// which return token.Position values. This enables precise error reporting
// and source code transformation:
//
//	node.Pos()  // Start position (line, column, offset)
//	node.End()  // End position
//
// # Type Information
//
// Expressions implement the TypedExpression interface, which provides
// GetType() and SetType() methods for semantic type information:
//
//	expr.GetType()          // Returns *TypeAnnotation
//	expr.SetType(typeInfo)  // Sets type information
//
// # Integration with Parser
//
// The parser (internal/parser) produces AST trees automatically. Users
// of the public API can access the AST via the dwscript.Program type:
//
//	program, err := engine.Compile(source)
//	if err == nil {
//	    ast := program.AST()  // Returns *ast.Program
//	}
package ast
