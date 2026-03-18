# FailureScripts Classification (5.4.1)

## Baseline

`go test ./internal/interp -run 'TestDWScriptFixtures/FailureScripts'`

Current result:

- Total fixtures: `541`
- Passed: `52`
- Failed: `489`

Raw failure modes from the harness:

- `391` `Compile diagnostic mismatch`
- `82` `Expected errors ... but got none`
- `16` `Runtime error mismatch`

The captured run used for this audit is stored in `/tmp/failurescripts-5.4.1.txt`.

## Buckets

### 1. Parser recovery / parser-side position drift

Count: `94`

These are fixtures where the primary mismatch is still parser-facing:

- wrong token/position after recovery
- extra trailing parser diagnostics after the expected primary error
- lost dotted-name / unit-prefix shape before semantic analysis can normalize it

Representative fixtures:

- `array_assign_add`
- `array_error4`
- `array_error7`
- `array_index_bracket_missing`
- `array_params1`
- `block_unfinished2`
- `case_error1`
- `unit_prefix1`
- `unit_prefix8`
- `var_incomplete1`

Typical actual signatures:

- `expected ';' after ...`
- `expected 'end' to close ...`
- `expected next token to be ...`
- `Unknown name "Internal"` instead of `Unknown name "Internal.sdfq"`
- position drift on bracket / bound diagnostics

Assessment:

- This is still a front-end pipeline problem, not mostly a missing feature problem.
- The unit-prefix and incomplete-var families look especially high-leverage because they fan out into generic downstream errors.

### 2. Compile gating leaks

Count: `16`

These fixtures still execute far enough to produce runtime mismatches where DWScript expects compile-time failure.

Representative fixtures:

- `array_const_item_count`
- `var_static_array`
- `unreachable_case_of`

Typical actual signatures:

- `Runtime Error: array literal has 4 elements, expected 3`
- runtime exception output after compilation should already have failed

Assessment:

- This bucket is exactly the shared gate problem Phase 5 is meant to eliminate.
- Static-array literal size validation is the clearest representative fix because it appears in more than one fixture family.

### 3. Semantic wording / position normalization

Count: `297`

These fixtures already fail at compile time, but the emitted diagnostics differ from DWScript in wording, anchoring, hint ordering, or specialization level.

Representative fixtures:

- `abstract_method`
- `array_add_subclass`
- `array_assign_error3`
- `array_assign_subclass`
- `array_bounds`
- `array_concat`
- `array_dyn_mismatch`
- `array_in1`
- `array_index_of`
- `swap1`
- `use_proc_result1`
- `virtual1`

Recurring subclusters:

- helper/member diagnostics are too generic:
  - `There is no accessible member ...` where DWScript expects helper-method parameter diagnostics or method-specific hints
- operator diagnostics are structurally correct but too Go-port specific:
  - `arithmetic operator + requires numeric operands ...`
  - `function 'Swap' arguments must have compatible types ...`
- type names / positions drift:
  - `Void` vs `void`
  - `array[0..1] of T` vs `array [0..1] of T`
  - off-by-one columns on argument diagnostics
- expected hints are missing or emitted in a different order:
  - `Empty THEN block`
  - case-mismatch hints
  - unused-result / unused-variable hints

Assessment:

- This is the highest-volume bucket.
- Most of it is not “missing feature”; it is normalization debt across semantic wording, position anchoring, and hint ordering.

### 4. True missing feature / missing validation

Count: `82`

These fixtures expected compile diagnostics but currently produce no errors at all.

Representative fixtures:

- `array_dyn_add`
- `string_set`
- `switch_invalid1`
- `switch_invalid2`
- `switch_invalid3`
- `triple_apos1`
- `triple_apos2`
- `unused_result`
- `unused_variables`
- `use_proc_result2`
- `var_ambiguous_in_scope`
- `var_type_init`
- `virtual_private`

Typical shape:

- the analyzer accepts the construct entirely
- no parser or semantic diagnostic is emitted
- the gap is not primarily formatting; the validation is absent

Assessment:

- This bucket should not be the first target for 5.4.2.
- It is lower leverage than the wording/normalization and parser-recovery buckets because each fix is more feature-specific.

## Priority Order For 5.4.2

Recommended order:

1. semantic wording / position normalization
2. parser recovery / parser-side position drift
3. compile gating leaks
4. true missing feature / validation gaps

Reasoning:

- The first two buckets affect the largest share of current failures and align with the Phase 5 architecture work already in progress.
- Compile-gating leaks are smaller in count but strategically important once a representative family is selected.
- True missing features are real, but they are less amenable to “collapse by root cause” and should come after the pipeline buckets.
