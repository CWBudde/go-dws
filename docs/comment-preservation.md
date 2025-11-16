# Comment Preservation in go-dws

This document describes the comment preservation feature implemented in go-dws as part of task 25.1.3.

## Overview

The lexer and AST have been extended to optionally preserve comments during tokenization and parsing. This feature is essential for code formatting tools, documentation generators, and other tools that need to maintain comments in the source code.

## Architecture

### Comment Storage

Comments are stored using a multi-layered architecture:

1. **Comment** - Represents a single comment with:
   - `Text`: The comment text including delimiters
   - `Pos`: Position in source code
   - `Style`: Type of comment (line, curly block, paren block)

2. **CommentGroup** - Groups consecutive comments together
   - Useful for multi-line comment blocks
   - Provides combined text representation

3. **NodeComments** - Stores leading and trailing comments for a node
   - `Leading`: Comments before the node
   - `Trailing`: Comments after the node (on same line)

4. **CommentMap** - Maps AST nodes to their comments
   - Stored in `Program.Comments`
   - Non-intrusive design - doesn't modify existing AST nodes

### Comment Styles

DWScript supports four comment styles:

1. **Line comments**: `// comment text`
2. **Curly brace block comments**: `{ comment text }`
3. **Parenthesis block comments**: `(* comment text *)`
4. **C-style block comments**: `/* comment text */`

All styles are preserved with their original delimiters.

## Usage

### Enabling Comment Preservation in Lexer

```go
import "github.com/cwbudde/go-dws/internal/lexer"

l := lexer.New(source)
l.SetPreserveComments(true)  // Enable comment preservation

for {
    tok := l.NextToken()
    if tok.Type == token.EOF {
        break
    }
    if tok.Type == token.COMMENT {
        // Process comment
        fmt.Println("Comment:", tok.Literal)
    }
}
```

### Using CommentMap in AST

```go
import "github.com/cwbudde/go-dws/pkg/ast"

// Create a comment map
comments := ast.NewCommentMap()

// Add leading comment to a node
node := &ast.Identifier{Value: "myVar"}
comment := &ast.Comment{
    Text:  "// This is my variable",
    Pos:   token.Position{Line: 1, Column: 1},
    Style: ast.CommentStyleLine,
}
comments.AddLeadingComment(node, comment)

// Retrieve comments for a node
nodeComments := comments.GetComments(node)
if nodeComments != nil && nodeComments.Leading != nil {
    fmt.Println("Leading comments:", nodeComments.Leading.Text())
}
```

## Current Limitations

### Parser Integration Not Yet Complete

**Status**: The lexer can preserve comments as COMMENT tokens, but the parser does not yet collect and attach them to AST nodes.

**Impact**:
- Comments are tokenized correctly when `preserveComments` is enabled
- But they are not automatically attached to the appropriate AST nodes
- Parser integration is needed for full comment preservation (future work)

**What Works**:
- ✅ Lexer can return COMMENT tokens
- ✅ Comment data structures are defined (Comment, CommentGroup, CommentMap)
- ✅ CommentMap field added to Program struct
- ✅ All comment styles (line, block, C-style) are recognized

**What Doesn't Work Yet**:
- ❌ Parser doesn't collect COMMENT tokens
- ❌ Comments are not attached to AST nodes during parsing
- ❌ Formatter cannot preserve comments (would need parser support)

### Future Work Needed

To complete comment preservation (tasks for later):

1. **Parser Integration** (Phase 25.2.6):
   - Modify parser to collect COMMENT tokens
   - Implement logic to attach comments to appropriate nodes
   - Handle leading comments (before a node)
   - Handle trailing comments (after a node on same line)
   - Handle orphan comments (not attached to any node)

2. **Printer/Formatter Integration** (Phase 25.2.6):
   - Update `pkg/printer` to read comments from CommentMap
   - Output leading comments before nodes
   - Output trailing comments after nodes
   - Maintain comment formatting and indentation
   - Handle blank lines between comment groups

3. **Edge Cases**:
   - Comments between tokens in expressions
   - Comments in complex nested structures
   - Preprocessor directives vs comments
   - Multi-line comment indentation

## Design Rationale

### Why CommentMap Instead of Node Fields?

The CommentMap approach was chosen over adding comment fields to every AST node for several reasons:

1. **Non-intrusive**: Doesn't require modifying hundreds of existing AST node types
2. **Optional**: Comments can be enabled/disabled without changing AST structure
3. **Backward compatible**: Existing code using the AST continues to work
4. **Memory efficient**: Only nodes with comments consume extra memory
5. **Flexible**: Easy to add/remove/modify comments without restructuring AST

### Why Preserve Comment Delimiters?

Comments are stored with their original delimiters (`//`, `{`, `(*`, `/*`) because:

1. **Fidelity**: Preserves original source code style
2. **Style preferences**: Different projects may prefer different comment styles
3. **Semantic meaning**: In some cases, comment style may indicate intent
4. **Round-trip accuracy**: Can reconstruct exact original source

## Testing

Comprehensive tests are provided:

- **Lexer tests** (`internal/lexer/comment_test.go`):
  - Comment preservation on/off
  - All comment styles
  - Multi-line comments
  - Position tracking
  - Unterminated comments

- **AST tests** (`pkg/ast/comment_test.go`):
  - Comment structures
  - CommentGroup operations
  - CommentMap operations
  - Helper functions

Run tests with:
```bash
go test ./internal/lexer -run TestComment
go test ./pkg/ast -run TestComment
```

## Examples

### Example 1: Line Comments

```dwscript
// This is a line comment
var x := 42;  // trailing comment
```

Preserved as:
- Comment 1: `"// This is a line comment"` at line 1
- Comment 2: `"// trailing comment"` at line 2

### Example 2: Block Comments

```dwscript
{
  Multi-line block comment
  with multiple lines
}
var y := "hello";
```

Preserved as:
- Comment 1: Full block including delimiters and newlines

### Example 3: Mixed Comments

```dwscript
(* Header comment *)
// Function documentation
function Add(a, b: Integer): Integer;
begin
  Result := a + b; { inline comment }
end;
```

All four comments are preserved with their original style and position.

## See Also

- [PLAN.md](../PLAN.md) - Task 25.1.3 for implementation details
- [formatter-style-guide.md](formatter-style-guide.md) - Formatting rules
- [Token types](../pkg/token/token.go) - COMMENT token definition
- [AST comment structures](../pkg/ast/comment.go) - Comment data types
