# AST Position & Trivia Audit (Phase 12.1.2)

This note captures the current state of position tracking and comment/trivia preservation inside the DWScript AST. It enumerates every gap the formatter must close before we can emit deterministic, comment-aware output.

## High-Level Findings

- **Start-only positions:** The `Node` interface exposes only `Pos() lexer.Position` (`internal/ast/ast.go:11-23`), so every consumer has access to a single start coordinate. There is *no* symmetrical `EndPos()` anywhere in the tree, which makes it impossible to reason about node extents or reconstruct whitespace between adjacent nodes.
- **Fallback positioning:** Several nodes (notably `Program` at `internal/ast/ast.go:39-67`) return their first child’s position or a hard-coded default when empty. That breaks deterministic formatting for empty units or when we need to anchor comments that precede the first statement.
- **Comment/trivia stripping:** The lexer skips all comments and compiler directives (`internal/lexer/lexer.go:281-323`), so no AST structure currently records leading/trailing trivia. The formatter therefore has nothing to re-emit.
- **Non-node helpers lack coordinates:** Utility structs that are not `Node`s but still need formatting context (parameters, enum values, record fields, case branches, interface methods, etc.) either omit tokens entirely or only keep bare strings. We must extend them so the formatter can attach inline comments or preserve spacing.

## Detailed Node Coverage

| Category | Nodes | Current Position Data | Notes |
| --- | --- | --- | --- |
| Root & Blocks | `Program`, `BlockStatement` (`internal/ast/ast.go:39-97`, `273-291`) | Only `begin` token (if any). No closing token, no empty-block support. | Need explicit start/end spans so comments after `end` or between blocks can be preserved. |
| Control Flow | `IfStatement`, `WhileStatement`, `RepeatStatement`, `ForStatement`, `CaseStatement` (`internal/ast/control_flow.go:17-214`) | Store the keyword token for the construct. Branch helpers like `CaseBranch` (`internal/ast/control_flow.go:153-171`) do not implement `Node` nor expose `Pos()`. | Formatter cannot place comments tied to specific case labels or know where an `else` arm ends. |
| Declarations | `VarDeclStatement`, `ConstDecl`, `TypeDeclaration`, `FunctionDecl`, `ClassDecl`, `RecordDecl`, `PropertyDecl`, `OperatorDecl`, `InterfaceDecl` (`internal/ast/statements.go`, `declarations.go`, `functions.go`, `classes.go`, `records.go`, `properties.go`, `operators.go`, `interfaces.go`) | All store the leading keyword token only. No span covering headers + bodies. | Lacking end positions prevents aligning `end;` / `until` / `finally` with their owners or detecting trailing comments. |
| Helper structs | `Parameter`, `InterfaceMethodDecl`, `EnumValue`, `RecordField`, `RecordPropertyDecl`, `PropertyDecl.IndexParams`, `OperatorDecl.OperandTypes` (`internal/ast/functions.go:17-41`, `interfaces.go:16-63`, `enums.go:14-66`, `records.go:71-157`, `properties.go:18-76`, `operators.go:46-119`) | No tokens or `Pos()` at all. Data is stored as strings or nested identifiers. | Without start/end info, formatter cannot emit user comments like `a: Integer; // count` inside parameter lists. |
| Expressions | `BinaryExpression`, `UnaryExpression`, `CallExpression`, `IndexExpression`, `MemberAccessExpression`, `MethodCallExpression`, etc. (`internal/ast/ast.go:140-239`, `arrays.go:175-235`, `classes.go:216-307`) | Positions point to operator or function tokens only. Parentheses, argument lists, and selectors do not track their closing delimiters. | Need richer metadata (e.g., location of `(` and `)`, `.` and `[]`) to control spacing around them. |
| Literals / Type Nodes | `ArrayLiteral`, `SetLiteral`, `EnumLiteral`, `RecordLiteral`, `TypeAnnotation` (`internal/ast/arrays.go:119-174`, `sets.go:66-161`, `enums.go:72-101`, `records.go:119-157`, `type_annotation.go:8-34`) | Start token only; enumerated members (`EnumValue`, `RecordField`, array elements) have no individual positions. | Required for preserving inline comments inside literals and for deciding when to break long lists. |

## Comment & Trivia State

- Line and block comments are consumed entirely inside `Lexer.NextToken` (`internal/lexer/lexer.go:293-323`). No token is emitted, and the parser has no hook to attach them to surrounding nodes.
- The AST contains no storage for leading or trailing trivia on any node. Even after we add comment tokens, we will need per-node slices (e.g., `LeadingTrivia []Trivia`) to persist them.

## Nodes Missing `Pos()` Altogether

While most exported AST nodes implement `Pos()`, the following frequently formatted constructs do not expose any position data at all:

- `CaseBranch` (`internal/ast/control_flow.go:153-171`)
- `Parameter` (`internal/ast/functions.go:17-35`)
- `InterfaceMethodDecl` (`internal/ast/interfaces.go:16-63`)
- `EnumValue` (`internal/ast/enums.go:14-29`)
- `RecordField` and `RecordPropertyDecl` (`internal/ast/records.go:71-157`)
- `PropertyDecl.IndexParams` (individual `Parameter` instances again; `internal/ast/properties.go:18-76`)

Each of these will need either to implement `Node` or at least carry explicit start/end offsets so the formatter can honor inline spacing and comments.

## Required Follow-Ups

1. **Augment the node contracts:** Extend `Node` with `EndPos() lexer.Position` (or introduce a `Span` helper) so every node records a half-open range. Populate the range in the parser as tokens are consumed.
2. **Propagate tokens to helpers:** Update `Parameter`, `CaseBranch`, `EnumValue`, `RecordField`, etc., to store their defining tokens (name, colon, equals, etc.) and expose `Pos()/EndPos()`.
3. **Capture trivia:** Modify the lexer to emit comment tokens (or a lightweight trivia structure) instead of discarding them, and teach the parser to associate trivia with the nearest AST node.
4. **Test coverage:** Add fixtures that assert we can round-trip code containing: leading file comments, inline parameter comments, blank-line-separated sections, and nested block comments.

With this audit complete, Phase 12.1.2 is satisfied—the formatter work can now rely on a concrete gap list for position and trivia handling.
