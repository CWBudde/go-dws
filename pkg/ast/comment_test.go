package ast

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/token"
)

func TestComment(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		style CommentStyle
		want  string
	}{
		{
			name:  "line comment",
			text:  "// This is a comment",
			style: CommentStyleLine,
			want:  "line",
		},
		{
			name:  "curly block comment",
			text:  "{ This is a comment }",
			style: CommentStyleCurly,
			want:  "curly",
		},
		{
			name:  "paren block comment",
			text:  "(* This is a comment *)",
			style: CommentStyleParen,
			want:  "paren",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Comment{
				Text:  tt.text,
				Pos:   token.Position{Line: 1, Column: 1},
				Style: tt.style,
			}

			if c.Style.String() != tt.want {
				t.Errorf("Style.String() = %q, want %q", c.Style.String(), tt.want)
			}
		})
	}
}

func TestCommentIsBlock(t *testing.T) {
	tests := []struct {
		name  string
		style CommentStyle
		want  bool
	}{
		{"line comment", CommentStyleLine, false},
		{"curly block comment", CommentStyleCurly, true},
		{"paren block comment", CommentStyleParen, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Comment{Style: tt.style}
			if got := c.IsBlock(); got != tt.want {
				t.Errorf("IsBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommentIsLine(t *testing.T) {
	tests := []struct {
		name  string
		style CommentStyle
		want  bool
	}{
		{"line comment", CommentStyleLine, true},
		{"curly block comment", CommentStyleCurly, false},
		{"paren block comment", CommentStyleParen, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Comment{Style: tt.style}
			if got := c.IsLine(); got != tt.want {
				t.Errorf("IsLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommentEnd(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		pos      token.Position
		wantLine int
		wantCol  int
	}{
		{
			name:     "single line",
			text:     "// comment",
			pos:      token.Position{Line: 1, Column: 1, Offset: 0},
			wantLine: 1,
			wantCol:  11,
		},
		{
			name:     "multi-line",
			text:     "{ line 1\nline 2\nline 3 }",
			pos:      token.Position{Line: 1, Column: 1, Offset: 0},
			wantLine: 3,
			wantCol:  9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Comment{
				Text: tt.text,
				Pos:  tt.pos,
			}

			end := c.End()
			if end.Line != tt.wantLine || end.Column != tt.wantCol {
				t.Errorf("End() = %d:%d, want %d:%d", end.Line, end.Column, tt.wantLine, tt.wantCol)
			}
		})
	}
}

func TestCommentGroup(t *testing.T) {
	comments := []*Comment{
		{Text: "// line 1", Pos: token.Position{Line: 1, Column: 1}},
		{Text: "// line 2", Pos: token.Position{Line: 2, Column: 1}},
		{Text: "// line 3", Pos: token.Position{Line: 3, Column: 1}},
	}

	cg := NewCommentGroup(comments...)

	if len(cg.Comments) != 3 {
		t.Errorf("got %d comments, want 3", len(cg.Comments))
	}

	if cg.Pos().Line != 1 {
		t.Errorf("Pos().Line = %d, want 1", cg.Pos().Line)
	}

	text := cg.Text()
	expected := "// line 1\n// line 2\n// line 3"
	if text != expected {
		t.Errorf("Text() = %q, want %q", text, expected)
	}
}

func TestCommentMap(t *testing.T) {
	cm := NewCommentMap()

	// Create a dummy node
	node := &Identifier{Value: "test"}

	// Test setting leading comments
	leadingComment := &Comment{Text: "// leading"}
	cm.AddLeadingComment(node, leadingComment)

	if !cm.HasComments(node) {
		t.Error("expected node to have comments")
	}

	nc := cm.GetComments(node)
	if nc == nil || nc.Leading == nil {
		t.Fatal("expected leading comments")
	}

	if len(nc.Leading.Comments) != 1 {
		t.Errorf("got %d leading comments, want 1", len(nc.Leading.Comments))
	}

	// Test setting trailing comments
	trailingComment := &Comment{Text: "// trailing"}
	cm.AddTrailingComment(node, trailingComment)

	nc = cm.GetComments(node)
	if nc == nil || nc.Trailing == nil {
		t.Fatal("expected trailing comments")
	}

	if len(nc.Trailing.Comments) != 1 {
		t.Errorf("got %d trailing comments, want 1", len(nc.Trailing.Comments))
	}

	// Test adding multiple leading comments
	cm.AddLeadingComment(node, &Comment{Text: "// another"})
	nc = cm.GetComments(node)
	if len(nc.Leading.Comments) != 2 {
		t.Errorf("got %d leading comments, want 2", len(nc.Leading.Comments))
	}

	// Test HasComments on node not in map (should not panic)
	nodeNotInMap := &Identifier{Value: "notInMap"}
	if cm.HasComments(nodeNotInMap) {
		t.Error("expected node not in map to have no comments")
	}

	// Test GetComments on node not in map
	nc = cm.GetComments(nodeNotInMap)
	if nc != nil {
		t.Error("expected GetComments to return nil for node not in map")
	}
}

func TestExtractCommentText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "line comment",
			input: "// hello",
			want:  "hello",
		},
		{
			name:  "curly block comment",
			input: "{ comment }",
			want:  " comment ",
		},
		{
			name:  "paren block comment",
			input: "(* comment *)",
			want:  " comment ",
		},
		{
			name:  "with whitespace",
			input: "  // hello  ",
			want:  "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractCommentText(tt.input)
			if got != tt.want {
				t.Errorf("ExtractCommentText(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNodeComments(t *testing.T) {
	// Test empty node comments
	var nc *NodeComments
	if nc.HasComments() {
		t.Error("nil NodeComments should not have comments")
	}

	// Test with leading comments only
	nc = &NodeComments{
		Leading: NewCommentGroup(&Comment{Text: "// leading"}),
	}
	if !nc.HasComments() {
		t.Error("expected to have comments")
	}

	// Test with trailing comments only
	nc = &NodeComments{
		Trailing: NewCommentGroup(&Comment{Text: "// trailing"}),
	}
	if !nc.HasComments() {
		t.Error("expected to have comments")
	}

	// Test with both
	nc = &NodeComments{
		Leading:  NewCommentGroup(&Comment{Text: "// leading"}),
		Trailing: NewCommentGroup(&Comment{Text: "// trailing"}),
	}
	if !nc.HasComments() {
		t.Error("expected to have comments")
	}
}
