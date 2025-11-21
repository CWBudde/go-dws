package ast

import "github.com/cwbudde/go-dws/pkg/token"

// BaseNode provides shared storage and behavior for AST nodes that carry token
// information. Embedding this struct allows nodes to expose TokenLiteral, Pos,
// and End without reimplementing the same boilerplate in every type.
type BaseNode struct {
	Token  token.Token
	EndPos token.Position
}

// TokenLiteral returns the literal associated with the node's token.
func (n *BaseNode) TokenLiteral() string {
	return n.Token.Literal
}

// Pos returns the token's starting position. This is the default location used
// for error reporting when a node does not override Pos().
func (n *BaseNode) Pos() token.Position {
	return n.Token.Pos
}

// End returns the end position for the node. When EndPos is set explicitly it
// takes precedence. Otherwise we advance the token's starting position by the
// literal length as a sensible default for leaf nodes.
func (n *BaseNode) End() token.Position {
	if n.EndPos.Line != 0 {
		return n.EndPos
	}
	pos := n.Token.Pos
	literalLen := len(n.Token.Literal)
	pos.Column += literalLen
	pos.Offset += literalLen
	return pos
}

// TypedExpressionBase extends BaseNode and was previously used to store type
// annotations. As of Task 9.18, type information is stored separately in
// SemanticInfo rather than on AST nodes themselves. This type is kept for
// backward compatibility but is just an alias for BaseNode.
type TypedExpressionBase struct {
	BaseNode
}
