# Compile stops

How a compiler stop ends compilation in upstream DWScript, how go-dws approximates it today,
and the model `PLAN.md` §2.1 replaces it with. Line references are as of branch
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

## 2. go-dws today: five mechanisms

go-dws runs lexer, parser and analyzer as separate passes. The lexer also runs ahead of the
parser. Each pass approximates "nothing after the stop" in its own way, and the frontend
reconciles the results afterwards.

### 2.1 Parser `stopped()`

- `ParserError.Stop` (`internal/parser/error.go:18`) marks a stop. `recordStop`
  (`parser.go:267`) and `addStructuredStop` (`:311`) set it, as do the `addExpectedStop*`
  helpers (`expected.go:73-103`). There are 62 stop-recording calls outside the definitions.
- `stopped()` (`parser.go:274`) scans the error list for a stop that has no `DeferredCall`. It
  derives the flag from the list rather than storing it, so restoring a speculative parse also
  undoes the stop.
- `recordError` (`parser.go:254`) drops every error once `stopped()` is true. It also drops
  `{$FATAL}` truncation artifacts at the synthetic end of input (`lexer.StoppedByFatal`,
  `internal/lexer/lexer.go:310`).
- `p.stopped()` is called at **22 sites in 10 files**: `classes.go` ×4, `records.go` ×3,
  `statements.go` ×3, `expressions_calls.go` ×3, `expressions_oop.go` ×3, `helpers.go` ×2, and
  one each in `control_flow.go`, `declarations.go`, `expected.go` and `parser.go`. These sites
  bail out of loops, so the parser does not keep building nodes after a stop.

### 2.2 Truncated calls, handled three ways

1. **Dropped (`nil`).** A call whose argument list hit a stop returns `nil`, so the analyzer
   never sees the call or the arguments it did read:
   - `expressions_calls.go:43` (non-identifier callee), `:74` (call or record literal),
     `:128` (`parseCallWithExpressionList`);
   - `expressions_oop.go:71` (`inherited X(…)`), `:198` (`new T(…)`), `:253` (`new (op)(…)`);
   - `classes.go:819` (`TClass.Create(…)` as a `NewExpression`).

   This is why `array_index_bracket_missing1` and `constructor_invalid_param` pass, and why
   `missing_parenthesis1` loses its `Invalid Operands`.
2. **`Incomplete` on `MethodCallExpression`** (`pkg/ast/classes.go:401`). It is set at
   `classes.go:849` and `:860`. The analyzer's `analyzeIncompleteMemberCall`
   (`internal/semantic/property_reintroduce.go:59`) skips the call and sets `compileStopped`.
   `DebugBreakExpression.Incomplete` (`pkg/ast/debugbreak.go:14`) and
   `RecordTypeNode.Incomplete` (`pkg/ast/type_expression.go:67`, set at `records.go:117`) are separate retain-for-diagnostics
   flags.
3. **`DeferredCall` provisional stop** (`error.go:21`). A `obj.M(` that runs into `;`, `end`
   or EOF records `Expression expected` with `Stop` and `DeferredCall` set (`classes.go:844-858`).
   `stopped()` ignores this error, so parsing continues past it. If an enclosing argument list is
   then cut short, `confirmDeferredCallStops` (`parser.go:646`, called from
   `expressions.go:329/364/395` and `expressions_calls.go:190`) makes the stop real. Otherwise
   the frontend removes it when the analyzer resolved the call as a property read (§2.4).

### 2.3 Analyzer `compileStopped`

- `Analyzer.compileStopped` (`internal/semantic/analyzer.go:118`) is a plain flag. It is set to
  true at 8 sites:
  - `analyze_expressions.go:324` (unknown name in a cast target) and `:491` (`Class reference
    expected`);
  - `analyze_expr_operators.go:292` (unknown name in an expression);
  - `analyze_function_calls.go:248`, `analyze_literals.go:304`, `analyze_statements.go:909`;
  - `property_reintroduce.go:62` (incomplete member call);
  - `type_punctuation.go:18` (`addPunctuationStop`).

  `SetCompileStopped` (`:703`) also sets it from the frontend when the parser stopped.
- It is read at only 3 sites, all of them end-of-program checks: `reportUnimplementedForwards`
  (`:578`), `validateForwardDeclarations` (`:596`) and `validateForwardMethods` (`:614`).
  **Analysis itself continues after the stop.**
- Separately, `SemanticError.Stop` (`errors.go:117`) is set at 6 sites, the 8 above minus
  `analyze_expr_operators.go:292` and `property_reintroduce.go:62`.
  `frontend.semanticDiagnostics` (`internal/frontend/result.go:806`) stops collecting at the
  first such diagnostic (`:847`). That `break` is the only thing that truncates the analyzer's
  own output.
- `pendingClassWarnings` (`analyzer.go:511`) is emitted regardless of `compileStopped`.
- Several upstream stops have no stop counterpart in go-dws. One example is
  `ReportNoMemberForType` (`dwsCompiler.pas:13386`, `CPE_UnknownMemberForType`), which go-dws
  reports as an ordinary error (`errors.go:459`). That is why `HelpersFail/strict` still prints
  the parser errors from line 11.

### 2.4 Frontend reconciliation (`internal/frontend/result.go`)

`compileParsedResult` (`:302`) runs analysis, then post-processes in this order (`:328-340`):

1. `refineTypePunctuationDiagnostics` (`type_punctuation.go:9`) drops a parser
   `Expression expected` stop when an analyzer `"(" expected` stop sits at the same position.
   It decides this by matching **message text**.
2. `refineDeferredPropertyCallDiagnostics` (`type_punctuation.go:30`) drops a `DeferredCall`
   stop whose call the analyzer resolved as a property read.
3. `dropDiagnosticsAfterStop` (`:356`) finds the earliest stop by position:
   - a parser stop cuts the semantic diagnostics after it;
   - a semantic stop cuts the parser diagnostics at or after it.

   Main-file diagnostics only; unit diagnostics are merged back unfiltered.
4. `restoreStatementWarningOrder` (`diagnostic_boundary.go:10`) moves the semantic prefix ahead
   of the parser boundary for statement-boundary warnings.
5. `sortDiagnostics` (`:446`) is a stable sort with roughly ten rules: directive vs. advisory,
   errors vs. hints, the deferred forward bucket, after-children stops, arity and static-class
   priority, parser order, phase, and specificity. `restoreDeclarationDiagnosticOrder`
   (`declaration_diagnostics.go:12`) runs between the two sorts.
6. A merge with unit diagnostics, then `sortDiagnostics`, then `filterDiagnostics` (`:876`),
   then `sortDiagnostics` again. `filterDiagnostics` is a set of **text filters**:
   - `Name expected` dedup within 8 columns;
   - `Expression expected before COLON`;
   - the `must have either a type annotation` / `variable declaration requires a type` filter
     (`:953-956`);
   - `already declared` after `Dot "." expected`;
   - unknown-type relocation onto `";" expected`;
   - unfinished-class-body suppression after `Unknown name`;
   - visible/accessible member dedup;
   - `expected 'end' to close unit declaration` after a fatal.

### 2.5 Lexer cutoff

`reachedLexerDiagnostics` (`result.go:642`) drops directive diagnostics positioned after the
earliest **parser** stop. That approximates the lazy tokenizer in `ParseWithOptions` (`:241`).
The unit path, `analyzeUnits` (`internal/frontend/units.go:59`), appends
`unit.DirectiveDiagnostics` without this cutoff.

`{$FATAL}` is not a frontend stop. `lexerDiagnostics` (`:676`) never sets `Stop`. Instead the
lexer ends tokenization (`directives.go:179`, `StoppedByFatal`), and the parser suppresses the
truncation noise (§2.1). Semantic diagnostics positioned after a `{$FATAL}` are not cut by
position.

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
| `FailureScripts/missing_parenthesis1` | Target (carrier, §3.1) | Missing `Invalid Operands` [1:30]; the call is dropped |
| `HelpersFail/strict` | Target (analyzer stop, §3.2) | Four extra parser errors on line 11 after the 9:11 member stop |
| `FailureScripts/block_unfinished2` | Target (completion point, §3.3) | Extra `Variable "i" declared but not used` [2:5] |
| `FailureScripts/array_index_bracket_missing1` | Must keep passing | Passes: the truncated `[…]` is never analyzed |
| `FailureScripts/constructor_invalid_param` | Must keep passing | Passes: no arity check on the truncated `new TMyClass(1 1` |
| `FailureScripts/static_methods` | Known divergence, not a target | Upstream omits its own `{$FATAL}` line (`docs/decisions/known-divergences.md`) |

Run `--classify` after each step, and move the newly isolated first-diagnostic gaps into
Phase 1 or Phase 4.
