# Compile stops

How a compiler stop ends compilation in upstream DWScript, how go-dws implements it today,
and the remaining steps of the model `PLAN.md` §2.1 moves towards. Line references are as of branch
`feat/compile-stop-model` (October 2026). Upstream references are to
`DWScript-Language-Server/DWScript/Source`.

## 1. Upstream model

DWScript compiles in **one pass**. The tokenizer is pulled lazily by the compiler, so lexing,
parsing, symbol resolution and type checking for a construct all happen at the same time, and
every message is appended to `TdwsCompileMessageList` in the order the compiler produced it.
`TdwsMessageList.AddMessage` (`dwsErrors.pas:433`) only appends, and nothing sorts the list
afterwards.

There are two kinds of error call (`dwsErrors.pas:762-874`):

| Call | Effect |
|---|---|
| `AddCompilerError` / `AddCompilerErrorFmt` | Appends a message. Compilation continues. |
| `AddCompilerHint` / `AddCompilerWarning` (+`Fmt`) | Appends if the hint level allows it or warnings are enabled. Compilation continues. |
| `AddCompilerStop` / `AddCompilerStopFmt` (`:841-874`) | Appends the error, sets `State := mlsStopped`, then **raises `ECompileError`** (`dwsErrors.pas:290`). |

`ECompileError` is caught in only three places:

- at the top of `TdwsCompiler.Compile` (`dwsCompiler.pas:1743`), which swallows it;
- in `RecompileInContext` (`:2262`), which does the same;
- in the `{$INCLUDE}` reader in `ReadInstrSwitch` (`:12631`), which re-raises it.

A stop therefore unwinds the whole compiler. Nothing after it is tokenized, parsed, resolved or
checked, so **no message after a stop exists**. That covers later syntax errors, later semantic
errors, and lexer messages beyond the stop point.

**Anchoring.** Most parser-shaped messages are anchored at `FTok.HotPos`
(`dwsTokenizer.pas:296`), the start of the token the compiler was looking at. Because emission
follows reading, an inner expression's error precedes the error of the expression containing it,
even when the outer error is displayed further left. A stop raised inside a call's argument list
still keeps every error already emitted for the arguments it finished reading
(`missing_parenthesis1`: `Invalid Operands` at 1:30 comes before `")" expected` at 1:35).

**Completion-point messages.** Some messages are emitted when a construct is *finished*, not
where it is declared:

- `ReadProcBody` raises `CPE_EndOfBlockExpected` as a stop (`dwsCompiler.pas:4155`) before it
  reaches `HintUnusedSymbols` / `HintUnusedResult` (`:4157-4158`). An unfinished routine body
  therefore never reports its unused locals (`block_unfinished2`), even though the declaration
  comes before the stop.
- `ReadBlocks` (`:4540`) and `ReadWith` (`:7022`) emit `HintUnusedSymbols` at block end, in the
  same way.
- `Compile` calls `HintUnusedPrivateSymbols` (`:1726`) and then `CurrentProg.Table.Initialize`
  (`:1729`) after `ReadScript` returns. `Initialize` reaches `CheckMethodsImplemented` and
  `CPE_ForwardNotImplemented` (`dwsSymbols.pas:3090`, `:4142`, `:6954`). A stop anywhere skips
  all of them.

The rule is not "drop what is positioned after the stop". It is **drop what would have been
emitted after the stop**. For a completion-point message, emission happens at the completion
point.

**Lexer and directives.** Tokenizer errors are raised while the compiler pulls tokens, so they
interleave with compiler messages. Malformed constants and invalid characters are stops
(`TTokenizer.AddCompilerStopFmtTokenBuffer`, `dwsTokenizer.pas:1305`, `:1520-1615`).
`{$FATAL 'msg'}` is `AddCompilerStop(switchPos, …, TCompilerErrorMessage)` in
`ReadInstrSwitch` (`dwsCompiler.pas:12767`). `{$ERROR}`, `{$WARNING}` and `{$HINT}` are ordinary
messages (`:12764-12766`). A directive after an earlier stop is never read.

## 2. go-dws today

go-dws runs lexer, parser and analyzer as separate passes. The lexer also runs ahead of the
parser. The parser marks what a stop cut short, the analyzer genuinely halts at its own stops,
and the frontend makes a position cut per source file before ordering the result.

### 2.1 Parser `stopped()`

- `ParserError.Stop` (`internal/parser/error.go`) marks a stop. `recordStop`, `addStructuredStop`
  and the `addExpectedStop*` helpers (`expected.go`) set it.
- `stopped()` (`parser.go:274`) scans the error list for a stop that is not `Provisional`
  (§2.2). It derives the flag from the list rather than storing it, so restoring a speculative
  parse also undoes the stop.
- `recordError` (`parser.go:254`) drops every error once `stopped()` is true. It also drops
  `{$FATAL}` truncation artifacts at the synthetic end of input (`lexer.StoppedByFatal`).
- `p.stopped()` is called at 16 sites in 9 files (`classes.go`, `records.go`, `statements.go`,
  `parser.go`, `control_flow.go`, `declarations.go`, `expected.go`, `expressions_calls.go`,
  `helpers.go`). These sites bail out of loops, so the parser does not keep building nodes
  after a stop.
- A block cut short by a missing `END` is kept with `BlockStatement.Truncated`
  (`pkg/ast/ast.go`); the analyzer then skips the block's completion hints
  (`isTruncatedBlock`, `analyze_statements.go:1513`; `unused_warnings.go`).

### 2.2 Truncated calls: one carrier

Every call form keeps its node when a stop cuts its argument list short: `CallExpression`,
`NewExpression`, `MethodCallExpression` and `InheritedExpression` carry `Truncated bool`.
`parseCallArguments` (`parser.go:301`) returns the completed arguments plus the flag;
`completedArgument` (`:309`) keeps an argument only if no stop cut it, or if it is itself a
truncated call. A record-literal-or-call list becomes a truncated `CallExpression`
(`buildTruncatedCallFromFields`, `expressions_calls.go`).

The analyzer's `analyzeTruncatedCall` (`analyze_function_calls.go:130`) analyzes the
completed arguments only and never resolves the call: no arity, overload or argument-type
checks, and no end-of-program checks. `containsParserRecovery`
(`recovery_diagnostics.go:10`) treats a truncated call like an `InvalidExpression`.
`DebugBreakExpression.Incomplete` and `RecordTypeNode.Incomplete` are separate
retain-for-diagnostics flags.

**Statement-boundary calls.** `obj.M(` directly followed by `;`, `end` or EOF is a stop only
if `M` is not a reintroduced property: upstream's `ReadPropertyExpr` recovers such a read with
a hint and an ordinary `")" expected` and goes on. The parser cannot tell, so
(`classes.go:840-856`):

- it marks the node `Truncated` and `StopDeferred` (`pkg/ast/classes.go:412`), and keeps
  parsing after the `;`;
- it records `Expression expected` as a `Provisional` stop, which `stopped()` ignores. That
  error is what parse-only callers (`ParseWithOptions`, `SkipTypeCheck`, analysis blocked by
  parser errors) see;
- if an enclosing argument list is then cut short, `confirmDeferredCallStops`
  (`parser.go:697`) turns the provisional stop into a real one;
- at the end of a source cut by `{$FATAL}`, `StopDeferred` stays false: the directive is the
  stop.

`analyzeStopDeferredCall` (`analyze_method_calls.go:15`) decides a `StopDeferred` call once
the member is resolved: unless it recorded a property read, it raises `Expression expected` as
an analyzer stop at the parser's anchor. Once analysis has run, the frontend drops the
provisional parser stops (`dropProvisional`, `result.go:347`).

### 2.3 Analyzer stops

- `addCompilerStop` (`internal/semantic/compile_stop.go`) records the diagnostic with `Stop`
  and raises the `compileStopSignal` sentinel panic. `analyzeUntilStop` recovers it at the
  units of work the analyzer runs in source order (top-level statements of a program or unit
  section, deferred routine and method bodies, signature registration), restoring the analysis
  context the unwinding skipped. `recoverCompileStop` ends an entry point. Deferred bodies
  declared before the stop still run (`insertionPrecedesStop`), because upstream compiled them
  first.
- 31 sites raise a stop, among them `"(" expected`, `Not a method`, `Class reference expected`,
  member-not-found (`NewAccessibleMemberError`, `CPE_UnknownMemberForType`), cast targets and
  the deferred boundary call above (all through `addCompilerStop` / `addPunctuationStop`).
- `compileStopped()` is true after a raised stop or when `skipEndOfProgramChecks` is set. It
  guards the end-of-program checks: `reportUnimplementedForwards` (`analyzer.go:593`),
  `validateForwardDeclarations` (`:608`) and `validateForwardMethods` (`:629`).
  `skipEndOfProgramChecks` is set by `SetCompileStopped` (`:719`, from the frontend when the
  parser stopped, for the program and every unit), by `analyzeTruncatedCall`, and by the
  unknown name in an expression (`analyze_expr_operators.go:296`). The unknown name is still
  not a raised stop: fixture-pinned diagnostics of the enclosing reads follow it.
- `frontend.semanticDiagnostics` (`result.go:824`) still stops collecting at the first stop
  diagnostic (`:864`). Deferred bodies run after a top-level stop is raised, so their
  diagnostics are emitted after it; the `break` removes them.

### 2.4 Frontend (`internal/frontend/result.go`)

`compileParsedResult` (`:303`) runs analysis, then for the main file:

1. `dropProvisional` removes the provisional parser stops of §2.2.
2. `dropDiagnosticsAfterStop` (`:372`) is the one position cut. It finds the earliest stop of
   either phase; at a tie the analyzer's stop wins, because the analyzer only reaches a parser
   stop's token through a carrier the parser left it to decide (an interrupted typed constant
   initializer, `const C: R = ;`, is a record constant's `"(" expected` upstream, not the
   parser's `Expression expected`).
   - a parser stop cuts the semantic diagnostics after it;
   - an analyzer stop cuts the parser diagnostics, including directive diagnostics, at or
     after it.
3. `restoreStatementWarningOrder` (`diagnostic_boundary.go:10`) moves the semantic prefix ahead
   of the parser boundary for statement-boundary warnings.
4. `sortDiagnostics` (`:464`), a stable sort with roughly ten rules: directive vs. advisory,
   errors vs. hints, the deferred forward bucket, after-children stops, arity and static-class
   priority, parser order, phase, and specificity. `restoreDeclarationDiagnosticOrder`
   (`declaration_diagnostics.go:12`) runs between the two sorts.
5. A merge with unit diagnostics, then `sortDiagnostics`, `filterDiagnostics` (`:894`) and
   `sortDiagnostics` again. `filterDiagnostics` is a set of **text filters**: `Name expected`
   dedup within 8 columns; `Expression expected before COLON`; the `must have either a type
   annotation` / `variable declaration requires a type` filter; `already declared` after
   `Dot "." expected`; unknown-type relocation onto `";" expected`; unfinished-class-body
   suppression after `Unknown name` (still needed by `param_partial3`, because the unknown name
   is not a raised stop); visible/accessible member dedup; `expected 'end' to close unit
   declaration` after a fatal.

### 2.5 Lexer cutoff and units

`reachedLexerDiagnostics` (`result.go:660`) drops directive diagnostics positioned after the
earliest **parser** stop, approximating the lazy tokenizer in `ParseWithOptions`. Analyzer
stops cut directive diagnostics in `dropDiagnosticsAfterStop`.

A unit that fails to parse fails to load (`internal/units/registry.go`), so a loaded unit's
stops come from its analyzer. `analyzeUnits` (`internal/frontend/units.go`) passes the unit's
directive diagnostics and its semantic diagnostics through `dropDiagnosticsAfterStop` together,
so the same cut applies per unit. A fatal unit directive still aborts before the unit is
analyzed.

`{$FATAL}` is not a frontend stop. `lexerDiagnostics` never sets `Stop`. The lexer ends
tokenization instead (`StoppedByFatal`), and the parser suppresses the truncation noise
(§2.1). Semantic diagnostics positioned after a `{$FATAL}` are not cut by position.

## 3. Target model (PLAN §2.1)

1. **One parser carrier.** Every call form keeps its node when the argument list is cut short:
   `CallExpression`, `NewExpression`, `MethodCallExpression`, `inherited`, and statement-boundary
   calls. One marker (for example `Truncated`) replaces `nil`, `Incomplete` and `DeferredCall`.
   The analyzer analyzes the **completed** children only: arguments whose parse finished before
   the stop. It never runs arity, overload, or property-vs-call checks on a truncated list.
   Property recovery reads the same marker. A deferred stop becomes an ordinary stop that the
   analyzer can withdraw by recording the property read.
2. **Real analyzer stops.** A stop is a typed event, a sentinel `panic` that `safeAnalyze`
   (`internal/frontend/result.go:401`) recovers. It halts analysis at the point it is raised,
   just as `ECompileError` unwinds. `compileStopped` and the `semanticDiagnostics` `break` go
   away. Every upstream `AddCompilerStop` site that go-dws mirrors raises it. That includes the
   unknown name in an expression, `"(" expected`, and `CPE_UnknownMemberForType` where upstream
   uses a stop. Because the halt skips them, the end-of-program checks need no guard.
3. **One cut.** The frontend takes the earliest stop of **any** origin (lexer `{$FATAL}` and
   malformed constants, parser, analyzer) and drops everything emitted after it, once, for the
   main file and for each unit. A completion-point message, such as unused locals, unused
   private members or unimplemented forwards, survives only if its *completion point* precedes
   the stop, not its display position. This is the `block_unfinished2` rule. The analyzer
   therefore has to carry the completion position alongside such a hint.
4. **Ordering only.** Whatever remains of `sortDiagnostics` only interleaves the parser and
   analyzer streams into single-pass order. `refineTypePunctuationDiagnostics`,
   `refineDeferredPropertyCallDiagnostics`, `dropDiagnosticsAfterStop` and the stop-related text
   filters in `filterDiagnostics` are deleted. `reachedLexerDiagnostics` folds into the one cut.

**Status.** Steps 1 and 2 are in place for the call forms and the sites listed in §2.2–§2.3;
`refineTypePunctuationDiagnostics` and `refineDeferredPropertyCallDiagnostics` are deleted.
Still open:

- The unknown name in an expression sets `skipEndOfProgramChecks` but does not unwind, so the
  `semanticDiagnostics` `break` and the class-body text filter stay.
- The statement-boundary call keeps a provisional parser error as the parse-only fallback,
  and `confirmDeferredCallStops` for nested lists.
- `{$FATAL}` and malformed constants are not stops in the cut, and the cut is still by display
  position; the completion-point rule (step 3) covers only blocks cut by a missing `END`.
- `dropDiagnosticsAfterStop` and `reachedLexerDiagnostics` remain as two halves of the cut;
  the text filters in `filterDiagnostics` remain.

## 4. Invariants

These come from the old PLAN 2.4 text (`git show febda4cd:PLAN.md`). Any change here must
preserve them.

- **§1.3 enum lookup.** Snapshotting constant bindings must preserve source declaration order
  and the insertion points of routine-body diagnostics. Its symbol-table changes keep the
  compile-stop and forward-check behaviour described here.
- **§1.3 cast recovery.**
  - Restore speculative parser errors *and stops* before falling back to expression parsing. A
    stop is part of the error list (§2.1), so restoring the list restores the stop.
  - Keep semantic recovery types.
  - Keep true compiler stops ahead of later assignment and end-of-program diagnostics.
- **§1.3 record metatype.**
  - Record annotations and constants stay instance types.
  - Supplying-expression diagnostics survive through static and helper receivers.
  - Classifying type values follows the same recovery and stop rules.
- **§1.3 anonymous signatures.**
  - Nested callable types resolve from their AST nodes, keeping each node's modifiers and
    `of object` ownership.
  - Assignment recovery and implicit-call intent stay intact, without invoking a returned
    callable.
- **Unit diagnostics.** Positions only compare within one source, so the cut applies per file.

## 5. Fixture anchors

| Fixture | Role | Today |
|---|---|---|
| `FailureScripts/missing_parenthesis1` | Target (carrier, §3.1) | Passes |
| `HelpersFail/strict` | Target (analyzer stop, §3.2) | Passes |
| `FailureScripts/block_unfinished2` | Target (completion point, §3.3) | Passes |
| `FailureScripts/array_index_bracket_missing1` | Must keep passing | Passes: the truncated `[…]` is never analyzed |
| `FailureScripts/constructor_invalid_param` | Must keep passing | Passes: no arity check on the truncated `new TMyClass(1 1` |
| `FailureScripts/param_partial3` | Must keep passing | Passes through the class-body text filter (§2.4) |
| `FailureScripts/static_methods` | Known divergence, not a target | Upstream omits its own `{$FATAL}` line (`docs/decisions/known-divergences.md`) |

Run `--classify` after each step, and move the newly isolated first-diagnostic gaps into
Phase 1 or Phase 4.
