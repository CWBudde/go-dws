# Progress log — October 2026

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
