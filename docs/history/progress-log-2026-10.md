# Progress log — October 2026

## 2026-10-03 — Compound operands and conversion recovery (PLAN 1.3)

Compound assignment checks now validate both operands. Unsupported scalar pairs
report `Incompatible operands` at the operator and retain the ordinary RHS
assignment mismatch when needed. The complete `assign_op_incompatible` fixture
now matches all five upstream diagnostics. Missing class operators stop
compilation; class-operator and dynamic-array append dispatch remain supported,
including assignment through a function-name result.

Registered builtin operators still check assignment compatibility. In particular,
`Integer += Float` remains rejected. Upstream `TAssignExpr.TypeCheckAssign` checks
the original RHS against the target after operator lookup; registrations alone
do not establish whether the assignment is valid. This was verified in
[compiler utilities](https://github.com/EricGrange/DWScript/blob/master/Source/dwsCompilerUtils.pas),
[core expressions](https://github.com/EricGrange/DWScript/blob/master/Source/dwsCoreExprs.pas)
and [conversion expressions](https://github.com/EricGrange/DWScript/blob/master/Source/dwsConvExprs.pas).

Numeric compound assignments convert a Variant RHS before the operation using
the shared conversion path. For example, Integer `10 *= Variant(2.4)` produces
20 while retaining the RHS value. A conversion that raises preserves the target
and original exception. Pending exceptions prevent conversion fallback retries
and writeback through variables, var parameters, implicit fields, properties and
class variables. Runtime tests also retain single receiver/index evaluation,
one property getter read and no setter write after a failed conversion.

Tests were written before each fix. Exact frontend diagnostics cover the complete
fixture, invalid operand recovery, valid numeric/string/Variant pairs, class
operator stops, function-name dispatch and array append. Runtime regressions cover
all four numeric operators, integer/float/string Variant contents, fractional
conversion, builtin conversion failure and eight custom raising-converter
boundaries. Additional array tests cover typed initializer nil rejection and
accepted nil class/interface elements, verifying the existing PR review fixes.
Independent review found the raised-conversion writeback gap; the fix passed
re-review with no remaining actionable findings.

`just fixture-update` ratchets FailureScripts **287 → 288**. An independent CLI
comparison against an immutable snapshot of `68ff7e9f` confirms exactly one newly
passing fixture, `assign_op_incompatible`, and no newly failing fixtures. The
generated status scores **1,380 / 2,014**, with BuildScripts at eight. The full
race run scores **1,379 / 2,014**, with BuildScripts at seven; its previously
documented varying count remains outside this change, and its stable baseline
stays seven. Fixture sources, expectations and scoring remain unchanged.

Validation: `just test-unit` (`go test -v -race ./...`) passed on the final source.
`golangci-lint run --new-from-merge-base=origin/main --timeout 10m` reported zero
issues; `go mod tidy -diff`, tracked/new Go formatting and `git diff --check` are
clean. The rebuilt CLI matches the compound fixture and the three array recovery
fixtures exactly; a function-name class-operator run prints the expected `1` and
retains its returned object. An earlier race run exhausted temporary disk quota; clearing
only the task's Go cache and running race tests and lint sequentially resolved
it. Checks use `GOCACHE=/tmp/go-dws-lint-cache`,
`GOLANGCI_LINT_CACHE=/tmp/go-dws-golangci-cache` and
`GOFLAGS='-buildvcs=false -p=1'`.

PLAN 1.3 retains checked compound subitems and expands the open enum, record and
anonymous callable tasks with their measured dependencies. Enum investigation
now explains both `enums9` and `enums10` through duplicate-preserving sorted local
midpoint lookup, including the program's Default/Internal/System unit entries;
the original [sorting change](https://github.com/EricGrange/DWScript/commit/260d65f9629648b9ab88b7134773f89a818691bb)
updated both fixture oracles. That lookup, immutable runtime enum bindings,
routine-pointer recovery, cast recovery, record metatypes and JSON autoboxing
remain open.

## 2026-10-03 — Assignment anchors and field-backed properties (PLAN 1.3)

Ordinary scalar and interface assignment failures now anchor at their right-hand
expression instead of `:=`. This includes binary coalesce expressions, whose
diagnostic position is the `??` operator, and values on a following line. Existing
array-expression, static-array and indexed-assignment anchors remain intact.

Properties whose writer is a field now use
`Incompatible types: Cannot assign "X" to "Y"` at the supplied value. Method-backed
setters retain their `Argument 0 expects type …` sentence and accessor position.
The same distinction applies to inherited and implicit/explicit `Self` access;
read-only, write-only and metaclass access checks still run before mismatch
reporting. Existing recovery tests now expect the corrected RHS positions.

Four fixtures match their complete upstream expectations: FailureScripts
`assign_error` and `coalesce_class`, and InterfacesFail `assign_obj_from_intf` and
`interface_inheritence2`. Test-first coverage pins all four and retains array
fixtures, multiline values, inherited member paths and method-backed setters.
The local DWScript reference directory was empty; checked-in upstream fixture
expectations supplied the comparison source.

`just fixture-update` ratchets FailureScripts **282 → 284** and InterfacesFail
**6 → 8**. Its generated report scores **1,373 / 2,014** with BuildScripts at
seven passes. The previous generated report had nine BuildScripts passes while
its baseline was already seven; those previously documented varying results are
excluded from this change's four gains. No fixture source, expectation or scoring
policy changed. Remaining PLAN 1.3 groups stay open with explicit fixture names.

Validation: `go test -coverprofile=/tmp/go-dws-assignment-final-coverage.out ./...`
passed, and `golangci-lint run --new-from-merge-base=origin/main --timeout 10m`
reported zero new issues. CLI fixture reports independently agree with the
harness at **284 / 529 scored** in FailureScripts and **8 / 19** in InterfacesFail.
Independent review found no actionable issues; formatting and `git diff --check`
are clean. Validation used `GOCACHE=/tmp/go-dws-lint-cache`,
`GOFLAGS=-buildvcs=false`, and a writable temporary lint cache. The unfiltered
`just ci` lint step still reports the repository's existing backlog; the check
above matches GitHub Actions' gate for newly introduced findings.

## 2026-10-03 — Array allocation recovery and interface conversion wording (PLAN 1.3)

Array allocation dimensions now report `Integer expression expected` for every
non-integer expression. Analysis retains the inferred nested-array type so later
assignments can be checked without spurious undefined-variable diagnostics.
Allocation assignment mismatches use the opening `[` position, including after
spaces, comments or a newline. The parser saves that position separately from the
AST's ordinary `new` position; variables, fields, indexed writes, function-name
writes and field-backed properties use it. Regenerating the AST visitor produced
no changes because the added field is position metadata.

The anchor was verified against upstream `ReadNew`/`ReadNewArray` in
[dwsCompiler.pas](https://github.com/EricGrange/DWScript/blob/master/Source/dwsCompiler.pas#L8758)
and `TestDelete` in
[dwsTokenizer.pas](https://github.com/EricGrange/DWScript/blob/master/Source/dwsTokenizer.pas#L1345):
consuming `[` saves its position before dimension expressions are read. This
rules out using the element type identifier's end as an approximate anchor.

Invalid class-to-interface assignment statements now report
`Class "X" does not implement interface "Y"` at the assignment operator.
The check uses the source's declared class type. Nil, interface aliases, direct
implementations and inherited class implementations remain accepted. The same
conversion sentence applies to fields and field-backed properties; method-backed
writers retain their setter-argument diagnostic. Compatibility rules are unchanged.
`interface_inheritence1` remains open in PLAN 4.5: the analyzer still treats a
class implementing a derived interface as implicitly implementing its base.

Three complete fixtures now match upstream: FailureScripts
`multi_dim_dyn_array1`, and InterfacesFail `assign_intf_from_obj` and
`assign_intf_from_intf`. Tests through the real compile path failed first with
the old sentences, missing second dimension error and wrong anchors. Additional
coverage checks bracket whitespace, recovery typing, field/property routes,
method-backed writers and valid interface conversions. Existing semantic,
frontend, interpreter and embedding tests now expect the corrected sentences.

`just fixture-update` ratchets FailureScripts **284 → 285** and InterfacesFail
**8 → 10**, with no other category changes. The generated report scores
**1,376 / 2,014**. Independent CLI reports agree at **285 / 529 scored** for
FailureScripts and **10 / 19** for InterfacesFail. Other array constructor,
compound operator, enum, cast, record and anonymous-procedure groups in PLAN 1.3
remain open.

Validation: `just test-unit` (`go test -v -race ./...`) passed after updating two
existing interpreter/embedding expectations for the new conversion sentence.
`golangci-lint run --new-from-merge-base=origin/main --timeout 10m` reported zero
new issues, and `go mod tidy -diff` found no dependency changes. The CLI built and
both category reports agreed with the harness. Independent review identified an
explicit-field/property allocation-anchor gap; a failing regression was added
and the gap corrected before final verification. Review then found no remaining
actionable issues.

Checks used `GOCACHE=/tmp/go-dws-lint-cache`, `GOFLAGS=-buildvcs=false`, and
`GOLANGCI_LINT_CACHE=/tmp/go-dws-golangci-cache`. All tracked Go files are formatted
(`git ls-files '*.go' | xargs gofmt -l` produced no output), and `git diff --check`
passed. `just check-fmt` could not create its script under the read-only default
runtime directory; with `XDG_RUNTIME_DIR=/tmp`, it ran but also scanned ignored
local `.cache/` and `.claude/` scratch files and worktrees, reporting their existing
formatting differences. Those local artifacts were left untouched.

## 2026-10-03 — Array constructors and nested callable syntax (PLAN 1.3)

Array assignment recovery now retains incompatible constructors' inferred element
types and cardinality. Dynamic-array targets and constant-array sources use the
assignment operator's position; static targets retain the constructor position.
Writes to constant elements still check the RHS after reporting the read-only
error. The complete `array_assign_error3` and `array_const` expectations now match,
including bounds checks and parser recovery.

Untyped empty array constants can initialize typed dynamic arrays. Compatible
static-to-dynamic conversions copy the source storage and apply the destination's
type metadata, so mutation does not change the constant. Ordinary dynamic arrays
retain reference sharing and element-type invariance. Runtime regressions cover
initializers, assignments, `var` parameters, fields, properties, class variables,
nested and associative slots, `Add`, `Insert`, and nested typed constructors.
Review exposed storage paths bypassing the initial conversion; failing tests
drove the shared conversion and the additional typed-storage boundaries.

The parser retains nested callable return types and their complete source spans,
including `of object`, and recovers a following declaration after malformed nested
functions. This completes only the parser prerequisite for JSON autoboxing;
semantic signature resolution, implicit invocation and boxing diagnostics remain
open. The CLI parses `var p: function: procedure; var tail: Integer;` successfully.

PLAN 1.3 now has explicit subitems and retains checked work while the task is open,
as requested. Enum binding remains open: a blanket first-binding change fixed
`enums9` but regressed the previously passing `enums10`. That implementation was
excluded; the plan now requires measuring and matching both duplicate-name lookup
expectations before changing analyzer or evaluator registration. The merged enum
assignment-anchor fix remains intact.

`just fixture-update` raises FailureScripts **285 → 287**. Independent CLI reports
against a clean snapshot of main (`34f25e8c`) identify exactly the two array gains
and no newly failing FailureScripts fixtures, agreeing with the harness at
**287 / 529 scored**. The generated snapshot scores **1,379 / 2,014**, including
eight BuildScripts passes. That category retains its previously documented
varying results; its stable baseline remains seven. No fixture source,
expectation or scoring policy changed.

Validation: `just test-unit` (`go test -v -race ./...`) passed, and
`golangci-lint run --new-from-merge-base=origin/main --timeout 10m` reported zero
issues. `go mod tidy -diff` found no dependency changes. The CLI built, matched the
harness fixture count, and parsed the nested callable example. Independent review
found no remaining actionable issues after the storage fixes. Tracked and newly
added Go files are formatted, and `git diff --check` passed. Checks used
`GOCACHE=/tmp/go-dws-lint-cache`, `GOFLAGS='-buildvcs=false -p=1'`, and a writable
temporary lint cache. The new branch starts at the merged main commit `34f25e8c`.
