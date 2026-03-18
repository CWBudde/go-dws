# FailureScripts Next-Phase Plan

This note turns the current `FailureScripts` backlog into implementation-sized families rather than broad pipeline buckets.

Source run:

- `/tmp/failurescripts-5.4.4.txt`
- baseline: [failure-scripts-5.4.4.md](/mnt/projekte/Code/go-dws/docs/failure-scripts-5.4.4.md)

Current totals:

- `541` total fixtures
- `76` passing
- `465` failing
- `370` compile diagnostic mismatches
- `82` expected errors but got none
- `13` runtime mismatches

## Important Interpretation

The families below are planning clusters, not a second exact taxonomy of the harness output.

- they are based on repeated fixture-name patterns and recurring actual diagnostics
- they are meant to produce actionable work slices
- they are intentionally broader than one exact parser/semantic/runtime mode

## Next-Phase Families

### 1. Array Semantics and Array-Diagnostic Compatibility

Approximate scope:

- at least `41` failing fixtures already match the `array_*` / dynamic-array naming cluster
- additional array-adjacent failures also appear outside the prefix group

Representative fixtures:

- `array_assign_error3`
- `array_bounds`
- `array_concat`
- `array_dyn_mismatch`
- `array_error4`
- `array_error7`
- `array_index_bracket_missing`
- `array_index_extra`
- `array_initialization4`
- `array_method1`

Concrete work:

- collapse remaining array-literal element mismatch wording onto DWScript-compatible type mismatch diagnostics
- normalize array index diagnostics:
  - `Array expected`
  - `Too many indices`
  - integer-vs-ordinal wording
  - lower/upper bound exceeded families
- fix malformed array-type / array-index recovery where the parser still prevents semantic normalization
- align array concat / array op diagnostics with DWScript type-pair expectations
- close missing-validation array fixtures such as `array_dyn_add`, `array_map1`, `array_of_const`

### 2. Class / Property / Static / Override / Visibility Families

Approximate scope:

- about `125` failing fixtures land in the class/property/override/static naming cluster

Representative fixtures:

- `class_cast`
- `class_const3`
- `class_error*`
- `class_property1`
- `property_error7`
- `static_class2`
- `static_methods`
- `virtual1`
- `virtual2`
- `visibility5`

Concrete work:

- parser support/recovery for static-class and strict class-declaration forms that still collapse too early
- override / inherited-signature / virtual-overlap diagnostics in DWScript wording
- property read/write/class-property diagnostics that still drift from expected messages
- visibility and class-member lookup specialization beyond the already migrated buckets
- abstract/static/deferred class completeness rules vs. missing-method fallout ordering

### 3. Warning / Hint Emission and Ordering

Approximate scope:

- about `54` fixtures group around contracts, control-flow warnings, and hint ordering
- warning-only families are also a large share of the remaining runtime-gating drift

Representative fixtures:

- `abstract_method`
- `array_in1`
- `contracts_error1`
- `contracts_warnings`
- `empty_if_block`
- `ignore_result`
- `unreachable`
- `unreachable_case_of`
- `unused_result`
- `unused_variables`

Concrete work:

- restore missing hints:
  - `Empty THEN block`
  - unused-result
  - unused-variable
  - unreachable-code
- normalize warning ordering against same-line semantic errors
- tighten deferred semantic/hint emission so warning-only fixtures do not slip to runtime
- audit contracts/control-flow warnings that currently surface partial output only

### 4. Parser Header / Declaration / Delimiter Recovery

Approximate scope:

- about `58` fixtures cluster around declaration headers and malformed delimiter recovery

Representative fixtures:

- `block_unfinished2`
- `case_error1`
- `case_error7`
- `const_record4`
- `except_error*`
- `param_partial*`
- `strict_parameter_type`
- `string_error`
- `string_get`
- `var_type`

Concrete work:

- improve recovery around parameter lists, type headers, `case`, `except`, and record/method declarations
- collapse duplicate `Expression expected` / `Name expected` fallout where a more specific parser error should win
- keep recoverable parser diagnostics localized so semantic analysis still sees the intended declaration shape
- normalize remaining delimiter wording:
  - `")" expected`
  - `"]" expected`
  - `END expected`
  - `DO expected`
  - `TO or DOWNTO expected`

### 5. Missing Validation Sweep

Exact scope:

- `82` fixtures still expect errors but get none

Representative fixtures:

- `assigned`
- `coalesce_class`
- `conditionals1` through `conditionals6`
- `const_array1`
- `contracts_precondition`
- `default_params2`
- `deprecated`
- `enum_flags_overflow`
- `switch_invalid1` through `switch_invalid3`
- `use_proc_result2`

Concrete work:

- treat this as the explicit “missing semantics” queue
- classify each no-error fixture as parser-only, semantic-only, warning-only, or runtime-only missing behavior
- prioritize families that can unlock many fixtures at once:
  - conditionals
  - default params
  - deprecated / warning annotations
  - enum flags
  - switch validation

### 6. Residual Runtime-Mismatch / Compile-Gating Sweep

Exact scope:

- `13` fixtures still mismatch at runtime

Representative fixtures:

- `class_var_dyn1`
- `div_by_zero_float`
- `div_by_zero_int`
- `dyn_array_setlength3`
- `for_in_subclass`
- `ignore_result`
- `missing_param1`
- `special_funcs1`
- `unreachable_case_of`

Concrete work:

- determine which of these should become compile-time diagnostics vs. warning-only compile output
- reuse the shared frontend gate rather than fixing them in interpreter-only paths
- prioritize families already adjacent to warning/deferred-semantic work

## Recommended Execution Order

1. warning / hint emission and ordering
2. array semantics and array-diagnostic compatibility
3. parser header / declaration / delimiter recovery
4. class / property / static / override families
5. missing validation sweep
6. residual runtime-mismatch sweep

Reasoning:

- warning/hint gaps still hide real progress in many fixtures and overlap with compile-gating cleanup
- array diagnostics remain one of the densest high-leverage compatibility families
- parser header recovery is still a multiplier on downstream semantic quality
- class/property work is large but fragmented and should follow after another parser/diagnostic cleanup pass
- the explicit no-error queue is best handled once the frontend/semantic wording surface is more stable
