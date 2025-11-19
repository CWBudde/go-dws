package parser

import (
	"reflect"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// NodeBuilder automatically tracks position information during node construction.
// It eliminates the need for manual EndPos assignment by capturing the current
// parser position when started and setting EndPos when finished.
//
// Usage pattern:
//
//	builder := p.StartNode()
//	// ... parse node components ...
//	stmt := &ast.IfStatement{...}
//	return builder.Finish(stmt).(*ast.IfStatement)
//
// Or when the end position comes from a child node:
//
//	builder := p.StartNode()
//	// ... parse node and children ...
//	stmt := &ast.ExpressionStatement{Expression: expr}
//	return builder.FinishWithNode(stmt, expr).(*ast.ExpressionStatement)
type NodeBuilder struct {
	p          *Parser
	startToken lexer.Token
}

// StartNode creates a new NodeBuilder that captures the current token position.
// Call this at the beginning of any parsing function that creates an AST node.
// Task 2.7.5.2: Cursor-only - always uses cursor.
//
// Example:
//
//	func (p *Parser) parseIfStatement() *ast.IfStatement {
//	    builder := p.StartNode()
//	    // ... parse if statement ...
//	    stmt := &ast.IfStatement{...}
//	    return builder.Finish(stmt).(*ast.IfStatement)
//	}
func (p *Parser) StartNode() *NodeBuilder {
	return &NodeBuilder{
		p:          p,
		startToken: p.cursor.Current(),
	}
}

// Finish sets the EndPos on a node to the current token's end position.
// This should be called after all parsing for the node is complete.
// Task 2.7.5.2: Cursor-only - always uses cursor.
//
// The EndPos is calculated as: current token position + token length
//
// Example:
//
//	builder := p.StartNode()
//	stmt := &ast.ReturnStatement{Value: expr}
//	return builder.Finish(stmt).(*ast.ReturnStatement)
//
// Returns the node for convenient chaining and type assertion.
func (nb *NodeBuilder) Finish(node ast.Node) ast.Node {
	setEndPos(node, nb.p.endPosFromToken(nb.p.cursor.Current()))
	return node
}

// FinishWithNode sets the EndPos on a node based on another node's end position.
// Use this when a node's end position should match a child node's end position.
//
// This is commonly used when a statement's last component is an expression:
//
//	builder := p.StartNode()
//	stmt := &ast.ExpressionStatement{Expression: expr}
//	return builder.FinishWithNode(stmt, expr).(*ast.ExpressionStatement)
//
// If lastChild is nil, behaves like Finish() and uses the current token position.
//
// Returns the node for convenient chaining and type assertion.
func (nb *NodeBuilder) FinishWithNode(node ast.Node, lastChild ast.Node) ast.Node {
	if lastChild != nil {
		setEndPos(node, lastChild.End())
	} else {
		setEndPos(node, nb.p.endPosFromToken(nb.p.cursor.Current()))
	}
	return node
}

// FinishWithToken sets the EndPos on a node based on a specific token's end position.
// Use this when the end position should be based on a particular token that was
// previously consumed, rather than the current token.
//
// Example:
//
//	builder := p.StartNode()
//	endToken := p.cursor.Current()  // Save the END keyword
//	p.nextToken()
//	block := &ast.BlockStatement{...}
//	return builder.FinishWithToken(block, endToken).(*ast.BlockStatement)
//
// Returns the node for convenient chaining and type assertion.
func (nb *NodeBuilder) FinishWithToken(node ast.Node, tok lexer.Token) ast.Node {
	setEndPos(node, nb.p.endPosFromToken(tok))
	return node
}

// StartToken returns the token that was current when StartNode() was called.
// This can be useful for error messages or when constructing the node.
func (nb *NodeBuilder) StartToken() lexer.Token {
	return nb.startToken
}

// setEndPos sets the EndPos field on an AST node using reflection.
// It handles both direct BaseNode embedding and TypedExpressionBase embedding.
//
// This function uses reflection to avoid needing to know the exact type of the node.
// While reflection has some overhead, it's only used during parsing (not runtime
// execution), and the simplicity and maintainability benefits outweigh the cost.
//
// Expected node structures:
//   - Statements: Direct BaseNode embedding (e.g., IfStatement, ForStatement)
//   - Expressions: TypedExpressionBase.BaseNode embedding (e.g., BinaryExpression, Identifier)
//
// Silent failure behavior:
// The function returns silently without setting EndPos when:
//   - The node is nil
//   - The node doesn't have a BaseNode field (direct embedding)
//   - The node doesn't have TypedExpressionBase.BaseNode (expression embedding)
//   - The BaseNode or EndPos fields are unexported or cannot be set
//
// This silent failure is intentional - it allows the NodeBuilder to be used uniformly
// across all node types without requiring type-specific handling. Nodes that don't
// conform to the expected structure simply won't have their EndPos set via reflection,
// which is acceptable since position tracking is primarily for error reporting.
func setEndPos(node ast.Node, pos lexer.Position) {
	if node == nil {
		return
	}

	v := reflect.ValueOf(node)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if !v.IsValid() || v.Kind() != reflect.Struct {
		return
	}

	// Try to find BaseNode field directly (used by statements)
	baseField := v.FieldByName("BaseNode")

	// If not found, try TypedExpressionBase (used by expressions)
	if !baseField.IsValid() {
		typedBase := v.FieldByName("TypedExpressionBase")
		if typedBase.IsValid() {
			baseField = typedBase.FieldByName("BaseNode")
		}
	}

	// Set EndPos if we found a valid BaseNode
	if baseField.IsValid() {
		endPosField := baseField.FieldByName("EndPos")
		if endPosField.IsValid() && endPosField.CanSet() {
			endPosField.Set(reflect.ValueOf(pos))
		}
	}
}
