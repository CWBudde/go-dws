package ast

import (
	"strings"

	"github.com/cwbudde/go-dws/pkg/token"
)

// Comment represents a single line or block comment.
// DWScript supports three comment styles:
// - Line comments: // ...
// - Curly brace block comments: { ... }
// - Parenthesis block comments: (* ... *)
type Comment struct {
	Text  string         // The comment text (including comment markers)
	Pos   token.Position // Position of the comment in the source
	Style CommentStyle   // Style of the comment
}

// CommentStyle represents the style of a comment
type CommentStyle int

const (
	// CommentStyleLine represents a line comment: // ...
	CommentStyleLine CommentStyle = iota

	// CommentStyleCurly represents a block comment: { ... }
	CommentStyleCurly

	// CommentStyleParen represents a block comment: (* ... *)
	CommentStyleParen
)

// String returns a string representation of the comment style
func (cs CommentStyle) String() string {
	switch cs {
	case CommentStyleLine:
		return "line"
	case CommentStyleCurly:
		return "curly"
	case CommentStyleParen:
		return "paren"
	default:
		return "unknown"
	}
}

// IsBlock returns true if the comment is a block comment
func (c *Comment) IsBlock() bool {
	return c.Style == CommentStyleCurly || c.Style == CommentStyleParen
}

// IsLine returns true if the comment is a line comment
func (c *Comment) IsLine() bool {
	return c.Style == CommentStyleLine
}

// End returns the end position of the comment
func (c *Comment) End() token.Position {
	lines := strings.Split(c.Text, "\n")
	lastLine := lines[len(lines)-1]

	if len(lines) == 1 {
		// Single-line comment
		return token.Position{
			Line:   c.Pos.Line,
			Column: c.Pos.Column + len(c.Text),
			Offset: c.Pos.Offset + len(c.Text),
		}
	}

	// Multi-line comment
	return token.Position{
		Line:   c.Pos.Line + len(lines) - 1,
		Column: len(lastLine) + 1,
		Offset: c.Pos.Offset + len(c.Text),
	}
}

// CommentGroup represents a sequence of comments with no other tokens between them.
// This is useful for grouping consecutive line comments or standalone block comments.
type CommentGroup struct {
	Comments []*Comment // List of comments in the group
}

// NewCommentGroup creates a new comment group with the given comments
func NewCommentGroup(comments ...*Comment) *CommentGroup {
	return &CommentGroup{Comments: comments}
}

// Pos returns the position of the first comment in the group
func (cg *CommentGroup) Pos() token.Position {
	if len(cg.Comments) > 0 {
		return cg.Comments[0].Pos
	}
	return token.Position{}
}

// End returns the end position of the last comment in the group
func (cg *CommentGroup) End() token.Position {
	if len(cg.Comments) > 0 {
		return cg.Comments[len(cg.Comments)-1].End()
	}
	return token.Position{}
}

// Text returns the concatenated text of all comments in the group
func (cg *CommentGroup) Text() string {
	if cg == nil || len(cg.Comments) == 0 {
		return ""
	}

	var buf strings.Builder
	for i, c := range cg.Comments {
		if i > 0 {
			buf.WriteByte('\n')
		}
		buf.WriteString(c.Text)
	}
	return buf.String()
}

// NodeComments stores leading and trailing comments for an AST node.
// Leading comments appear before the node, trailing comments appear after.
type NodeComments struct {
	Leading  *CommentGroup // Comments before the node
	Trailing *CommentGroup // Comments after the node (on same line)
}

// HasComments returns true if there are any comments
func (nc *NodeComments) HasComments() bool {
	return nc != nil && (nc.Leading != nil || nc.Trailing != nil)
}

// CommentMap maps AST nodes to their associated comments.
// This allows comments to be attached to nodes without modifying
// the existing AST node structures.
type CommentMap map[Node]*NodeComments

// NewCommentMap creates a new empty comment map
func NewCommentMap() CommentMap {
	return make(CommentMap)
}

// SetLeading sets the leading comments for a node
func (cm CommentMap) SetLeading(node Node, comments *CommentGroup) {
	if node == nil || comments == nil {
		return
	}
	if cm[node] == nil {
		cm[node] = &NodeComments{}
	}
	cm[node].Leading = comments
}

// SetTrailing sets the trailing comments for a node
func (cm CommentMap) SetTrailing(node Node, comments *CommentGroup) {
	if node == nil || comments == nil {
		return
	}
	if cm[node] == nil {
		cm[node] = &NodeComments{}
	}
	cm[node].Trailing = comments
}

// GetComments returns the comments for a node, or nil if none
func (cm CommentMap) GetComments(node Node) *NodeComments {
	if node == nil {
		return nil
	}
	return cm[node]
}

// HasComments returns true if the node has any comments
func (cm CommentMap) HasComments(node Node) bool {
	nc := cm.GetComments(node)
	return nc != nil && nc.HasComments()
}

// AddLeadingComment adds a single leading comment to a node
func (cm CommentMap) AddLeadingComment(node Node, comment *Comment) {
	if node == nil || comment == nil {
		return
	}

	nc := cm[node]
	if nc == nil {
		nc = &NodeComments{}
		cm[node] = nc
	}

	if nc.Leading == nil {
		nc.Leading = &CommentGroup{Comments: []*Comment{comment}}
	} else {
		nc.Leading.Comments = append(nc.Leading.Comments, comment)
	}
}

// AddTrailingComment adds a single trailing comment to a node
func (cm CommentMap) AddTrailingComment(node Node, comment *Comment) {
	if node == nil || comment == nil {
		return
	}

	nc := cm[node]
	if nc == nil {
		nc = &NodeComments{}
		cm[node] = nc
	}

	if nc.Trailing == nil {
		nc.Trailing = &CommentGroup{Comments: []*Comment{comment}}
	} else {
		nc.Trailing.Comments = append(nc.Trailing.Comments, comment)
	}
}

// ExtractCommentText returns the comment text without the comment markers.
// For example:
// - "// hello" → "hello"
// - "{ comment }" → " comment "
// - "(* comment *)" → " comment "
func ExtractCommentText(text string) string {
	text = strings.TrimSpace(text)

	// Line comment
	if strings.HasPrefix(text, "//") {
		return strings.TrimSpace(text[2:])
	}

	// Curly brace block comment
	if strings.HasPrefix(text, "{") && strings.HasSuffix(text, "}") {
		return text[1 : len(text)-1]
	}

	// Parenthesis block comment
	if strings.HasPrefix(text, "(*") && strings.HasSuffix(text, "*)") {
		return text[2 : len(text)-2]
	}

	return text
}
