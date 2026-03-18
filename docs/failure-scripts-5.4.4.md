# FailureScripts Delta After 5.4.2.x / 5.4.3 (5.4.4)

## Command

`go test ./internal/interp -run 'TestDWScriptFixtures/FailureScripts' -count=1`

The captured run for this audit is stored in `/tmp/failurescripts-5.4.4.txt`.

## Before / After

Baseline from `5.4.1`:

- Total fixtures: `541`
- Passed: `52`
- Failed: `489`
- Compile diagnostic mismatch: `391`
- Expected errors but got none: `82`
- Runtime error mismatch: `16`

Current snapshot after `5.4.2.1` through `5.4.2.4` and `5.4.3`:

- Total fixtures: `541`
- Passed: `76`
- Failed: `465`
- Compile diagnostic mismatch: `370`
- Expected errors but got none: `82`
- Runtime error mismatch: `13`

Delta:

- `+24` additional passing fixtures
- `-24` total failures
- `-21` compile diagnostic mismatches
- `-3` runtime mismatches
- no reduction yet in the true missing-validation bucket (`82`)

## What Landed In This Jump

The main improvement came from front-end pipeline fixes rather than new feature work:

- operator / initializer wording normalization
- array helper and array-member diagnostic normalization
- parser recovery fixes for unit-prefix and incomplete-declaration families
- compile-side static-array literal size mismatch reporting
- explicit regression coverage for mixed parser + semantic diagnostic streams

This confirms Phase 5 is reducing fixture failures by collapsing shared diagnostic root causes, not by one-off fixture patching.

## Remaining Buckets

The remaining failures are still dominated by four broad families.

### 1. Semantic wording and position normalization

This is still the largest bucket by volume. The failures are already compile-side, but the emitted diagnostics differ from DWScript in wording, specificity, hint presence, or column anchoring.

Representative fixtures:

- `abstract_method`
- `array_add_subclass`
- `array_assign_error3`
- `array_bounds`
- `array_concat`
- `array_dyn_mismatch`
- `array_in1`
- `swap1`
- `var_object`
- `virtual1`

Recurring signatures:

- helper- or operation-specific DWScript messages still rendered as generic Go-port diagnostics
- column drift on argument and assignment errors
- hint gaps or ordering drift (`Hint: Empty THEN block`, unused-result/unused-variable hints)
- overly generic parser/semantic fallout replacing a more specific DWScript diagnostic

### 2. True missing validation / missing feature coverage

This bucket is unchanged at `82`. These fixtures still expect compile diagnostics but currently produce no diagnostic at all.

Representative fixtures:

- `array_dyn_add`
- `array_map1`
- `array_of_const`
- `string_set`
- `switch_invalid1`
- `switch_invalid2`
- `switch_invalid3`
- `triple_apos1`
- `unused_result`
- `unused_variables`
- `use_proc_result2`

Assessment:

- this is now the clearest non-pipeline backlog
- these failures will require real semantic/interpreter feature work, not just normalization

### 3. Residual parser recovery / parser-side shape gaps

The earlier parser bucket shrank, but it is not closed. Some fixtures still fail before semantic normalization gets a clean AST shape.

Representative fixtures:

- `array_assign_add`
- `array_error4`
- `array_error7`
- `array_index_bracket_missing`
- `array_new`
- `strict_parameter_type`
- `static_class2`
- `static_class3`
- `try_except1`

Recurring signatures:

- expected token wording still differs from DWScript
- malformed constructs cascade into unknown-name fallout
- some parser recoveries still anchor at the wrong token after bracket/type failures

### 4. Residual compile-gating leaks

This bucket improved (`16 -> 13`) but is still open.

Representative fixtures:

- `unreachable_case_of`
- fixtures with missing warning-only compile diagnostics that still run through to runtime

Assessment:

- the static-array literal family is fixed
- the remaining leaks are now concentrated in warnings / control-flow / deferred-analysis families rather than one shared array-literal path

## Recommended Next Focus

Highest leverage next:

1. keep attacking semantic wording / position normalization families
2. pick a true missing-validation family with broad reach
3. return to parser recovery only for high-fanout grammar gaps
4. revisit the remaining runtime mismatches once the warning/deferred-semantic picture is cleaner

The important result from this checkpoint is that the front-end architecture changes are producing measurable fixture movement:

- pass count is up from `52` to `76`
- the largest reduction is in compile diagnostic mismatches
- the missing-feature bucket is now more clearly isolated
