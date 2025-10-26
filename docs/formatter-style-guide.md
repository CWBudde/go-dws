# DWScript Formatter Style Guide (Phase 12.1.1)

This document defines the canonical formatting rules that the Phase 12 formatter must apply. The rules pull from existing DWScript fixtures (`testdata/**/*.dws`), Delphi/DWScript conventions, and the _Object Pascal Style Guide_ (Borland/Inprise, 1997) suggested in the task description. All formatter features, whether used by the CLI, LSP, or playground, must emit code that matches this guide unless a user overrides specific options.

## Goals

- Produce deterministic, idempotent output given any valid DWScript source.
- Preserve user comments and intentional blank lines while normalizing whitespace elsewhere.
- Rely solely on the parsed AST plus trivia (leading/trailing comments) rather than best-effort token munging, ensuring formatting stays semantically aware.
- Mirror upstream DWScript/Object Pascal expectations so users switching from Delphi feel at home.

## Baseline Conventions

- **Indentation:** 4 spaces per scope level. No tabs in emitted text. `case` arms are indented one level (4 spaces) relative to the `case` keyword; labels align directly under the first indent, and statements inside an arm add another 4 spaces.
- **Line length:** Soft limit 100 columns. Long constructs (parameter lists, generic constraints) may wrap; formatter inserts line breaks before exceeding the limit and indents continuation lines by an extra 4 spaces.
- **Keyword casing:** Emit DWScript keywords in lowercase (`begin`, `then`, `repeat`). Preserve original casing for identifiers, unit names, and string literals.
- **Blank lines:** Collapse multiple consecutive blank lines into at most two and keep at most one blank line between logically related sections (e.g., between type declarations). Retain a single blank line before `begin` of a program body when declarations exist.
- **Trailing whitespace:** Never emit trailing spaces; ensure file ends with exactly one newline.

## Structural Rules

### Declarations

- `type`, `var`, `const`, and `resource string` sections start at column 1. Each declaration within the section is indented by 4 spaces.
- Class/interface members:
  - `private`, `protected`, `public`, etc., align with 4 spaces inside the class. Members under each visibility section are indented by 8 spaces.
  - Property declarations keep `read`/`write` clauses on the same line when they fit; otherwise break after the semicolon and indent continuation lines by 4 spaces.
- Procedure/function headers place the parameter list on the same line when possible; wrap after the opening parenthesis if any parameter causes the line to exceed 100 columns. Continuation lines indent by 4 spaces relative to the `function` keyword.
- Attributes (e.g., `[attribute]`) stay attached to the following declaration with no blank line between them.

### Blocks and Control Flow

- `begin`/`end` pairs align vertically; statements inside the block receive one indent level.
- Single-line statements following control keywords (`if`, `while`, `for`, `repeat`, `try`) remain on their own line with an indent, even if the body is a single statement. No `begin`/`end` omission heuristics: if the AST contains a compound statement, emit `begin`/`end`; otherwise indent the single statement.
- `else` pairs align with the matching `if`. `else if` chains emit as `else` followed by `if` on the next line for clarity.
- `case` statements:
  - `case <expr> of` stays on one line.
  - Each label is indented by 4 spaces and followed by `:`. Multiple labels share a single line separated by commas.
  - Statements associated with a label are indented an additional 4 spaces.
  - The optional `else`/`otherwise` arm is treated like another label block and indented accordingly.
- `try/finally/except` blocks align keywords; nested `except` sections indent their statements by 4 spaces.
- `repeat/until` loops align `repeat` and `until`; body statements indent by 4 spaces.

### Expressions and Operators

- Surround binary operators (`+ - * / := = <> > < >= <= and or xor`) with single spaces. Unary operators do not receive a space between operator and operand (`-Value`, `not Condition`).
- After commas and semicolons inside lists, insert a single space (unless at end of line).
- Function calls keep arguments on the same line if they fit; otherwise, break after `(` and align subsequent args with 4-space indentation.
- Parentheses appear only when needed to respect precedence; the formatter should consult the AST’s operator associativity rules rather than reprinting original parentheses unless the AST indicates explicit grouping.

### Literals and Comments

- Preserve string literal contents exactly (aside from escaping already enforced by the parser).
- Numeric literals retain their original base (decimal/hex/binary) and casing for prefixes (`$` or `%`).
- Line comments (`//`) stay attached to the following or preceding statement depending on their recorded trivia. Ensure at least one space between `//` and the first character of the comment body.
- Block comments `{ ... }` and `(* ... *)` remain in place with surrounding whitespace normalized to guarantee there is at least one newline before and after when they appear between statements.

## Formatter Behavior per Use Case

- **CLI (`dwscript fmt`):** Processes files/directories, obeying `format.Options`. Supports `-w` (write in place) and `-l` (list files that would change).
- **Playground / WASM bridge:** Expose `Format(source string) (string, error)`; use default style settings to ensure Monaco’s Format button yields identical output to the CLI.
- **Editor integrations:** Provide full-document and range formatting. Range formatting must expand to whole statements/blocks to maintain syntactic correctness but should avoid touching unrelated code.

## Extensibility & Future Work

- Document profile knobs (indent size, case style) but keep the official DWScript profile as the default.
- Track known exceptions (e.g., preprocessor directives, inline assembly) where the formatter may have to operate in token mode rather than AST mode.
- Collect feedback via `TEST_ISSUES.md` entries once the formatter lands to refine heuristics in later sub-phases.

With this guide in place, Phase 12.1.1 is complete: downstream tasks can now reference an explicit contract for the formatter’s output.
