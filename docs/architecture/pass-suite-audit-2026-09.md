# Execution-suite audit — September 2026

**Measured at:** `HEAD` = `30818aed` plus the uncommitted for-loop diagnostics work and E1/E2,
2026-09-12.
**How:** `just fixture-report --in-scope --classify --list-fails`. Reproducible, not a transcript —
regenerate it rather than trusting the numbers below.

Companion to [`fail-suite-audit-2026-09.md`](fail-suite-audit-2026-09.md), which covers the `*Fail`
error-detection suites behind [`PLAN.md`](../../PLAN.md) §4. This one covers everything else: the
suites that **run** a program and compare its output. They had never been measured this way, which
is why §3 said "no open items" while 151 in-scope fixtures were failing in them. Six of those have
since closed (E1, E2); the tables below are the state after that.

## Why this exists

The pass/fail table says *how many* fail. It does not say whether a fixture is one word away from
passing or produces nothing at all, and `baselines.json` holds per-category pass-count **floors**,
so it cannot see a fixture swap one wrong line for another. Both questions now have one answer:
`cmd/fixture-report --classify` (PLAN.md T8, closed 2026-09-12).

## Method, and one number that changed

Every fixture runs through the production CLI as a subprocess:

```bash
dwscript run --diagnostics=plain --test-envelope --hints <level> <file>
```

Output and expectation are normalized identically (CRLF→LF, right-trim each line, strip the ends),
then aligned by longest common subsequence. That yields per fixture:

- **missing** — lines DWScript prints that go-dws does not;
- **spurious** — the reverse;
- **distance** — line-level edit distance, counted per diff hunk;
- **kind** — whether what differs is diagnostics, program output, both, nothing at all, a crash or
  a timeout.

⚠️ **`distance` is not the number the 2026-09-12 `*Fail` audit used.** That was a throwaway shell
script, since lost, which sorted both sides and compared with `comm` — so a diagnostic emitted in
the *wrong words* counted as 2 (one missing line, one spurious). It is one edit, and the tool now
counts it as 1. The `*Fail` near-miss figures therefore move from **68 / 194** to **156 / 281**
(one edit away / two or fewer) without anything about the port changing. The current definition
lives in `lineDiff`'s doc comment, which is the point of closing T8: the number is regenerable and
its meaning is written down next to the code that produces it.

Blank lines are kept. Normalization has already removed the leading and trailing ones, so a blank
that survives is one the program actually printed — `SimpleScripts/print_multi_args` differs from
its expectation by exactly one, and an earlier version of this tool scored it as *failing at
distance 0*.

## What it found

620 in-scope fixtures fail. **475 are in the `*Fail` suites** (§4's territory) and **145 are in the
execution suites**, split like this:

| Category | Fail | =1 | ≤2 | diagnostics | output | mixed | empty |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| SimpleScripts | 71 | 5 | 13 | 5 | 10 | 55 | 1 |
| ArrayPass | 19 | 3 | 7 | 1 | 4 | 14 | 0 |
| JSONConnectorPass | 14 | 4 | 7 | 0 | 8 | 5 | 1 |
| InterfacesPass | 12 | 3 | 6 | 1 | 0 | 11 | 0 |
| FunctionsMath | 10 | 0 | 0 | 0 | 0 | 10 | 0 |
| HelpersPass | 5 | 0 | 0 | 0 | 1 | 4 | 0 |
| OperatorOverloadPass | 3 | 0 | 1 | 0 | 0 | 3 | 0 |
| LambdaPass, Memory, OverloadsPass, FunctionsGlobalVars | 2 each | 0 | 0 | 0 | 0 | 8 | 0 |
| BuildScripts, FunctionsString, PropertyExpressionsPass | 1 each | 1 | 1 | 1 | 0 | 1 | 1 |
| **Total** | **145** | **16** | **35** | **8** | **23** | **111** | **3** |

**No fixture crashes and none times out**, across all 2,044 — the `*Fail` suites included. That is
the standing result of PR #401, and this is the check that keeps it standing.

### `mixed` is the headline, and it is not what it looks like

111 of 145 are `mixed`: both a diagnostic and the program's output differ. For an execution suite
that is almost always **one** fault, not two. go-dws reports a compile error the program does not
have, so the program never runs, so its output is missing too. The spurious diagnostic is the
cause; the missing output is its shadow.

Closing E1 and E2 did not touch a single `mixed` fixture, which is the shape of the remaining work.
This is why the execution suites look worse than they are, and why ranking them by distance
misleads: a one-line spurious error on a program that prints forty lines scores 41.

### The shape that dominates is a missing feature, not a wrong sentence

Unlike §4, where the work is DWScript's vocabulary, the execution suites fail on **things go-dws
cannot do**. The spurious diagnostics say so directly:

| Shape | Where | What it means |
| --- | --- | --- |
| `There is no accessible member with name "X" for type Integer/Float` | FunctionsMath (6 of 10) | primitive helpers absent: `TestBit`, `Compare`, `PopCount` |
| `There is no accessible member with name "X" for type array of Float` | FunctionsMath | array maths absent: `Pack`, `Offset`, `Multiply`, `MultiplyAdd` |
| `Unknown name "X"` | SimpleScripts 8, ArrayPass 4, JSONConnectorPass 2 | a name the script legitimately uses |
| `'X' operator requires class instance, got IInterface` | InterfacesPass 3 | interface casting and comparison |
| `unknown type 'X'` | SimpleScripts 3, JSONConnectorPass 2 | a type the script legitimately declares |

Reading these needs the `mixed` caveat above: each one is a spurious error that also swallowed the
program's output, so the fixture counts once here and its missing output is not a separate problem.

The one genuinely cross-cutting *diagnostic* family is
`Hint: "X" does not match case of declaration ("X")` — **28 in-scope fixtures want it and do not get
it, 9 get it where upstream does not**. It is implemented (`Analyzer.addCaseMismatchHint`) but wired
in by hand at twenty-odd separate resolution sites, each deciding independently what the declared
name is. Its `sole` yield is only 5, so it is not a quick win; it is a structural one, and it
belongs wherever a name finally resolves to a declaration rather than at each caller.

### The near misses worth naming

16 execution-suite fixtures are one edit from passing, down from 21: the cluster this section
named first — `SimpleScripts`' `diagnostics`-only failures, where the program runs and prints
correctly and only a message is wrong — was **runtime-message vocabulary**, F8's story moved from
compile time to run time, and it shipped as E1/E2.

| Fixture | DWScript | go-dws, before | |
| --- | --- | --- | --- |
| `div_by_zero_int` | `Division by zero` | `division by zero: 1 div 0` | closed |
| `mod_by_zero_int` | `Division by zero` | `modulo by zero: 1 mod 0` | closed |
| `string_bounds2` | `Lower bound exceeded! Index 0` | `string index out of bounds: 0 (string length is 6)`, and column 9 for upstream's 10 | closed |
| `external` | `Unhandled call to external symbol "Dummy" from` | `function 'Dummy' has no body` | closed |
| `re_raise` | keeps the original message, appends the re-raise position | wraps it as `User defined exception:` and drops the second position | closed |
| `call_conventions` | `[line: 3, column: 30]` | `Calling convention "safecall" is ignored at 3:30` | closed |
| `const_array_empty` | right sentence, column 17 | column 18 | **open — E8** |

The sixth row was a *formatting* bug rather than a vocabulary one: the analyzer built the position
into the message text at two sites, which also left `frontend.Diagnostic` with no structured
position at all. Methods and free routines now share one hint.

The seventh did not close, and the reason is worth the space. Moving the read path onto the same
anchor every *write* already uses closes `const_array_empty` and breaks
`ArrayPass/array_element_byref`, because upstream reports a genuine by-reference bind one column
further on than a read — and go-dws routes `AsString(a[a.Length])` down the read path by mistake.
`prepareArrayElementReference` evaluates the index with `e.Eval`, and in that path `a.High` and
`a.Length` evaluate to NIL, so any index expression containing a member access degrades the `var`
parameter to a copy. That is PLAN.md §3.5 **E8**; the two fixtures close together once it is fixed.

Two of the remaining near misses are the case-mismatch hint (`inherited_constructor`,
`recursive_path`). Two are spurious *compile* errors that stop the program running at all:
`ignore_result` needs `String.Replace`, and `assert_variant` needs `Assert` to accept a Variant
condition. The rest do not cluster:

- **`print_multi_args`** — `Print("")` must emit the blank line it does not.
- **`JSONConnectorPass`** — four at one edit, but not one cause: `global_var` prints `"hello"` for
  `hello` (an implicit string conversion keeping its quotes), while `implicit_to_int2` prints
  `null` for `{"test":1}`, which is a conversion that lost its value rather than a rendering
  choice.
- **`array_in_func_ptr`, `array_in_func_ptr2`** — one output line each.

## Regenerating this

```bash
just build
go run ./cmd/fixture-report --build=false --in-scope --classify --list-fails
go run ./cmd/fixture-report --build=false --category SimpleScripts --classify   # one suite
```

`--in-scope` drops the host-library categories excluded from every PLAN.md target
([`out-of-scope.md`](../decisions/out-of-scope.md)); without it the spurious tables are dominated by
unimplemented host libraries, which is honest for the headline and useless for directing work.
