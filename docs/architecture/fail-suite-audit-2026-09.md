# `*Fail` suite audit — September 2026

**Measured at:** `HEAD` = `30818aed`, 2026-09-12.
**How:** every `*Fail` fixture compiled through `bin/dwscript` as a subprocess and its output
diffed line-by-line against the sibling `.txt`; see [Method](#method). Read-only.

Measured audit behind [`PLAN.md`](../../PLAN.md) §4. It answers one question: *what exactly stands
between go-dws and DWScript's diagnostics?* The families in §4 came from a March 2026 reading of
`FailureScripts` alone; this is a full re-measurement over every `*Fail` suite, and it changed the
ordering enough to be worth keeping as a document rather than a paragraph.

Companion to [`audit-2026-09.md`](audit-2026-09.md), which does the same for §2.

## Method

Every `*Fail` fixture was run through the production CLI as a subprocess under `timeout 10`:

```bash
dwscript run --diagnostics=plain --compile-only --hints pedantic <file>
```

Diagnostics go to **stderr**, so `2>&1` is mandatory. Actual and expected output were stripped of
blank lines and CRs, sorted, and compared with `comm`, giving per fixture a count of **missing**
lines (DWScript says it, go-dws does not) and **spurious** lines (the reverse). Messages were then
clustered by shape, with positions, quoted contents and digits replaced by placeholders.

Two things this does that the standing tooling cannot:

- `baselines.json` holds per-category pass-count **floors**, so a fixture that swaps one wrong line
  for another is invisible to `just fixture-update`. The line-level diff sees it.
- Sorting before comparing separates *content* failures from *ordering* failures. Exactly two
  fixtures fail on ordering alone, which is why this count reads 160 passing where the harness
  reads 158.

**Superseded 2026-09-12, same day.** T8 is closed: the measurement now lives in
`cmd/fixture-report --classify` and this document's numbers are reproducible with
`just fixture-report --in-scope --classify --list-fails`. The tool differs from the script above in
two ways that change the figures below, neither of them a change to the port:

- It aligns the two sides by **longest common subsequence** instead of sorting them, so it also
  reports the ordering failures the sorted comparison hid — which is why the harness and this
  document disagreed by two fixtures.
- **Distance is line-level edit distance**, so a diagnostic emitted in the *wrong words* counts as
  one edit rather than two lines. The near-miss table below therefore reads **156 / 281** when
  regenerated, not 68 / 194. `PLAN.md` §4 carries the current numbers; treat the ones here as the
  historical record of what was measured on the day.

The execution suites got the same treatment once the tool existed:
[`pass-suite-audit-2026-09.md`](pass-suite-audit-2026-09.md).

## Headline

480 in-scope fixtures fail. `COMConnectorFailure`'s 8 are host-library and excluded
([`../decisions/out-of-scope.md`](../decisions/out-of-scope.md)).

| distance from the expectation | fixtures |
| --- | --- |
| exactly **one** line wrong | 68 |
| two or fewer | 194 |
| three or fewer | 269 |
| only **spurious** output, nothing missing | 35 |
| only **missing** output, nothing spurious | 90 |
| both | 355 |

More than half the remaining work is within two lines of done. That argues for working the
near-miss queue across families rather than draining one family at a time.

## The two findings that reordered `PLAN.md` §4

### go-dws's own diagnostic vocabulary is the largest single blocker

**265 of the 480** fixtures emit at least one message go-dws invented rather than inherited — 551
lines across 291 distinct shapes. They are mechanically identifiable: every DWScript sentence
begins with a capital letter and none of these do.

This had been filed as an S-sized cleanup ("convert the remaining raw `addError(...)` sites to
structured diagnostics"). It is the precondition for more than half the section. The parser is in
it as much as the analyzer — the top shapes are recovery sentences, not semantic ones.

### The missing-validation sweep was only ever counted over `FailureScripts`

**70 fixtures compile completely clean** where DWScript reports something: 43 in `FailureScripts`
and **27 in the other suites**, which had never been measured. The densest pocket is `HelpersFail`,
where 10 of 18 failures produce nothing at all — helpers accept far more than they should.

| suite | silent fixtures |
| --- | --- |
| HelpersFail | 10 (`helper_duplicate_member`, `helper_error3`, `helper_error4`, `helper_not_implemented`, `helper_of_delegate`, `helper_overload_error`, `helper_static`, `integer_helper`, `mixed_helper`, `static_class_method_self`) |
| InterfacesFail | 4 (`interface_guid`, `interface_inheritence1`, `interface_properties`, `intf_forwarded_not_implem1`) |
| JSONConnectorFail | 3 (`coalesce_typ`, `extend`, `parameters_check`) |
| LambdaFail | 3 (`global_ref`, `no_arrow`, `no_begin`) |
| OverloadsFail | 3 (`overload_func_ptr_param`, `overload_param_missing`, `overload_proc_value`) |
| GenericsFail | 2 (`constraint1`, `declaration_params_error3`) |
| PropertyExpressionsFail | 2 (`external_class`, `interface_property_auto_field`) |

## Ordering

Two fixtures produce every right line in the wrong order.

- **`array_in1`** — ~~within one source line upstream orders by **column**~~: the error at column 6
  precedes the `Empty THEN block` hint at column 17, while go-dws emits the hint first.
  **Corrected, and closed, 2026-09-12:** column is not the rule. Ordering by column also fixes
  `for_in1` but breaks `use_proc_result1`, which wants a hint at column 10 before an error at
  column 8 on the same line. The rule is **emission order** — upstream writes both streams as it
  compiles, inner before outer — and go-dws's same-line severity special case was simply removed.
  See [the September progress log](../history/progress-log-2026-09.md#2026-09-12--the-for-loop-diagnostics-4--f1-f10).
- **`infinite_loop`** — upstream reports **routine bodies before the main body**: `Trap`'s warnings
  at lines 6 and 3, then the main program's at 19, 21, 35. go-dws does the reverse. The cause is
  the deferred body checking introduced by L-S1b/L-S1c, which was right and should stay; the
  diagnostics need re-ordering on the way out, not the analysis on the way in. Nesting order within
  a single body already matches (inner `until` before outer `while`).

## Capitalization and rendering near-misses

Messages go-dws gets right except for one character:

| expected | produced | fixtures |
| --- | --- | --- |
| `Overload of "X" will be ambiguous with a previously declared version` | `overload of …` | `OverloadsFail/default_params`, `forwards` |
| `Overloaded procedure "X" must be marked with the "overload" directive` | `overloaded procedure …` | `OverloadsFail/forwards`, `forwards_unit`, `overload_missing`, `overload_simple` |
| `There is already a method with name "X"` | `there is already a method …` | `OverloadsFail/default_params`, `overload_simple` |
| `Incompatible types: "nil" and "String"` | `… "Nil" …` | `AssociativeFail/contains` |
| `Argument 0 expects type "String" instead of "Float"` | `… "string" …` | `FailureScripts/incorrect_type1` |

## Per-suite standing

Failing / one line away / two or fewer:

| suite | failing | 1 line | ≤ 2 |
| --- | --- | --- | --- |
| FailureScripts | 378 | 53 | 148 |
| HelpersFail | 18 | 5 | 9 |
| InterfacesFail | 18 | 2 | 8 |
| OverloadsFail | 14 | 1 | 5 |
| PropertyExpressionsFail | 10 | 3 | 6 |
| SetOfFail | 9 | 1 | 7 |
| GenericsFail | 8 | 0 | 2 |
| JSONConnectorFail | 7 | 1 | 2 |
| LambdaFail | 6 | 2 | 3 |
| OperatorOverloadFail | 6 | 0 | 1 |
| AssociativeFail | 3 | 0 | 2 |
| AttributesFail | 2 | 0 | 0 |
| InnerClassesFail | 1 | 0 | 1 |

## Message-shape inventories

### Missing message shapes

| lines | fixtures | shape |
| --- | --- | --- |
| 70 | 64 | `Syntax Error: "X" expected` |
| 58 | 22 | `Syntax Error: Incompatible types: "X" and "X"` |
| 38 | 33 | `Syntax Error: Name expected` |
| 37 | 17 | `Syntax Error: Incompatible types: Cannot assign "X" to "X"` |
| 34 | 15 | `Syntax Error: Method "X" of class "X" not implemented` |
| 26 | 25 | `Syntax Error: Expression expected` |
| 20 | 12 | `Syntax Error: Name "X" already exists` |
| 16 | 12 | `Syntax Error: Constant expression expected` |
| 16 | 11 | `Syntax Error: Cannot assign a value to the left-side argument` |
| 16 | 13 | `Syntax Error: Unknown name "X"` |
| 15 | 14 | `Syntax Error: Type expected` |
| 13 | 9 | `Hint: "X" does not match case of declaration ("X")` |
| 12 | 6 | `Syntax Error: Incompatible parameter types - "X" expected (instead of "X")` |
| 12 | 5 | `Syntax Error: Assignment's right-side-argument has no return type` |
| 12 | 7 | `Syntax Error: String expected` |
| 12 | 5 | `Warning: Unreachable code` |
| 11 | 9 | `Syntax Error: Class reference expected` |
| 11 | 8 | `Syntax Error: There is no overloaded version of "X" that can be called with these arguments` |
| 10 | 3 | `Syntax Error: Numerical operand expected` |
| 10 | 7 | `Syntax Error: Argument N expects type "X" instead of "X"` |
| 10 | 8 | `Syntax Error: There is already a method with name "X"` |
| 9 | 9 | `Syntax Error: Colon "X" expected` |
| 9 | 4 | `Syntax Error: Range start and range stop are of incompatible types: "X" and "X"` |
| 9 | 7 | `Syntax Error: Class "X" isn't defined completely` |
| 9 | 3 | `Syntax Error: Argument N (a) cannot be passed as Var-parameter` |

### go-dws-only message shapes (541 lines, 257 fixtures, 290 shapes)

| lines | fixtures | shape |
| --- | --- | --- |
| 20 | 11 | `expected 'X' after field name or method/property declaration keyword` |
| 14 | 7 | `expected 'X' after parameter list` |
| 11 | 11 | `variable 'X' must have either a type annotation or an initializer` |
| 11 | 11 | `expected 'X' to close class declaration` |
| 10 | 9 | `duplicate method signature for 'X'` |
| 10 | 2 | `implementation signature for 'X' does not match forward declaration` |
| 9 | 9 | `expected 'X' or 'X', got SEMICOLON` |
| 7 | 5 | `cannot infer type for variable 'X' from initializer` |
| 7 | 4 | `method 'X' not declared in class 'X'` |
| 6 | 4 | `expected parameter name` |
| 6 | 6 | `expected 'X', got SEMICOLON` |
| 5 | 3 | `array element N has type String, expected Integer` |
| 5 | 3 | `invalid assignment target` |
| 5 | 2 | `expected unit name after 'X'` |
| 5 | 5 | `expected next token to be RPAREN, got SEMICOLON instead` |
| 5 | 4 | `there is already a method with name "X"` |
| 5 | 1 | `function 'X' argument N must be a variable` |
| 4 | 4 | `unknown type 'X'` |
| 4 | 2 | `circular inheritance detected in class 'X'` |
| 4 | 3 | `expected 'X' after external` |
| 4 | 2 | `expected 'X' after constant value` |
| 4 | 3 | `cannot assign to read-only variable 'X'` |
| 4 | 4 | `constant 'X' must have a value` |
| 4 | 2 | `optional parameters cannot have lazy, var, or const modifiers` |
| 4 | 4 | `address-of operator (@) requires a function or procedure name` |
| 4 | 2 | `unary - requires numeric operand, got String` |
| 4 | 2 | `operator 'X' already defined for operand types (TObject, TObject)` |
| 4 | 2 | `overload of "X" will be ambiguous with a previously declared version` |
| 3 | 3 | `expected type after 'X' operator` |
| 3 | 3 | `expected type expression, got )` |

## The near-miss queue

The 68 fixtures that are exactly one line from passing. "add" means the line is missing, "drop"
means go-dws produces it and should not.

| fixture | needs | the one line |
| --- | --- | --- |
| `FailureScripts/array_initialization4` | add | `Syntax Error: Incompatible types: "array [0..3] of Float" and "array [0..2] of Float" [line: 1, column: 33]` |
| `FailureScripts/array_method1` | add | `Syntax Error: Invalid Instruction - function or assignment expected [line: 2, column: 1]` |
| `FailureScripts/assigned` | add | `Syntax Error: Invalid argument type [line: 1, column: 20]` |
| `FailureScripts/block_unfinished2` | drop | `Hint: Variable "i" declared but not used [line: 2, column: 5]` |
| `FailureScripts/case_error3` | drop | `Syntax Error: expected ':' after case value [line: 4, column: 8]` |
| `FailureScripts/case_of_else` | add | `Hint: Redundant "begin" in clause of a case..of [line: 5, column: 6]` |
| `FailureScripts/class_const1` | drop | `Syntax Error: cannot access private constant 'cPrivate' of class 'TBase' [line: 5, column: 14]` |
| `FailureScripts/class_const4` | add | `Syntax Error: Constant Instruction - has no effect [line: 4, column: 13]` |
| `FailureScripts/class_property3` | drop | `Syntax Error: Class method or constructor expected [line: 16, column: 14]` |
| `FailureScripts/class_var_dyn1` | add | `Syntax Error: Object reference needed to read/write an object field [line: 4, column: 35]` |
| `FailureScripts/const_param2` | add | `Syntax Error: Argument 0 (v) cannot be passed as Var-parameter [line: 8, column: 11]` |
| `FailureScripts/contracts_unfinished4` | add | `Syntax Error: ";" expected [line: 3, column: 6]` |
| `FailureScripts/default_params4` | add | `Syntax Error: Expression expected [line: 6, column: 9]` |
| `FailureScripts/dyn_array1` | drop | `Syntax Error: No arguments expected [line: 2, column: 10]` |
| `FailureScripts/dyn_array_setlength2` | drop | `Syntax Error: No arguments expected [line: 2, column: 10]` |
| `FailureScripts/enum_flags_overflow` | add | `Syntax Error: Enumeration element overflow [line: 8, column: 19]` |
| `FailureScripts/enums5` | add | `Syntax Error: "(" expected [line: 3, column: 11]` |
| `FailureScripts/enums6` | drop | `Syntax Error: expected next token to be RPAREN, got SEMICOLON instead [line: 1, column: 25]` |
| `FailureScripts/external1` | add | `Syntax Error: Name expected [line: 2, column: 29]` |
| `FailureScripts/external3` | add | `Syntax Error: Class "TFoo" has no default constructor [line: 6, column: 16]` |
| `FailureScripts/for_var_usage` | add | `Warning: Assignment to FOR-Loop variable [line: 2, column: 19]` |
| `FailureScripts/for_var_usage2` | add | `Warning: Assignment to FOR-Loop variable [line: 3, column: 9]` |
| `FailureScripts/foreach_invalid_arg` | drop | `Syntax Error: There is no overloaded version of "b" that can be called with these arguments [line: 6, column: 11]` |
| `FailureScripts/func_ptr_constant_ambiguous` | add | `Syntax Error: Ambiguous matching overloads of "IntToStr" [line: 3, column: 13]` |
| `FailureScripts/heredoc` | add | `Syntax Error: End of string constant not found (end of file) [line: 4, column: 1]` |
| `FailureScripts/hint_pedantic` | drop | `Hint: "test" does not match case of declaration ("Test") [line: 5, column: 1]` |
| `FailureScripts/ifthenelse_expression2` | drop | `Syntax Error: invalid alternative expression in if-then-else [line: 1, column: 10]` |
| `FailureScripts/in_operator7` | drop | `Hint: Empty THEN block [line: 1, column: 16]` |
| `FailureScripts/inherited5` | add | `Syntax Error: Cannot read a write only property [line: 14, column: 24]` |
| `FailureScripts/invalid_ucs2_char` | add | `Syntax Error: Invalid char constant "$200000" [line: 2, column: 19]` |
| `FailureScripts/method1` | add | `Syntax Error: Class name expected [line: 1, column: 8]` |
| `FailureScripts/method_implem7` | add | `Syntax Error: Unexpected method implementation [line: 13, column: 2]` |
| `FailureScripts/method_param_error1` | drop | `Syntax Error: argument 1 to method 'PrintMe' of class 'TMyClass' has type String, expected Integer [line: 8, column: 2]` |
| `FailureScripts/method_param_error2` | drop | `Syntax Error: argument 1 to method 'PrintMe' of class 'TMyClass' has type String, expected Integer [line: 6, column: 9]` |
| `FailureScripts/missing_operand1` | drop | `Syntax Error: expected ')', got SEMICOLON [line: 1, column: 8]` |
| `FailureScripts/missing_operand3` | drop | `Syntax Error: expected ')', got SEMICOLON [line: 4, column: 12]` |
| `FailureScripts/new_class3` | add | `Syntax Error: Method "Doh" of class "TMyClass" not implemented [line: 3, column: 19]` |
| `FailureScripts/open_array2` | add | `Syntax Error: Argument 0 expects type "array of const" instead of "array of Variant" [line: 7, column: 6]` |
| `FailureScripts/property_error6` | add | `Syntax Error: Unexpected "Integer Literal" [line: 8, column: 9]` |
| `FailureScripts/record_meta` | add | `Syntax Error: Incompatible types: Cannot assign "TSiteData" to "meta of TSiteData" [line: 16, column: 35]` |
| `FailureScripts/reserved_escape_empty` | drop | `Syntax Error: Undefined variable 'Integer' [line: 1, column: 17]` |
| `FailureScripts/reserved_escape_number` | drop | `Syntax Error: Undefined variable 'Integer' [line: 1, column: 18]` |
| `FailureScripts/sealed` | add | `Syntax Error: Class "TBase" is sealed, inheriting is not allowed [line: 5, column: 20]` |
| `FailureScripts/static_methods` | drop | `Compile Error: aborted [line: 26, column: 4]` |
| `FailureScripts/string_set` | add | `Syntax Error: Input data of invalid size: 4 instead of 1 [line: 3, column: 2]` |
| `FailureScripts/triple_apos1` | add | `Syntax Error: Incorrect triple apostrophe string indentation [line: 1, column: 10]` |
| `FailureScripts/triple_apos2` | add | `Syntax Error: Incorrect triple apostrophe string [line: 2, column: 13]` |
| `FailureScripts/unit_prefix4` | drop | `Syntax Error: type 'TMyClass' already declared [line: 2, column: 1]` |
| `FailureScripts/use_proc_result2` | add | `Syntax Error: Argument 0 expects type "Variant" [line: 1, column: 7]` |
| `FailureScripts/var_incomplete1` | drop | `Syntax Error: variable 's' must have either a type annotation or an initializer [line: 1, column: 1]` |
| `FailureScripts/var_incomplete3` | drop | `Syntax Error: variable 'i' must have either a type annotation or an initializer [line: 1, column: 1]` |
| `FailureScripts/virtual_private` | add | `Hint: Private virtual methods cannot be overridden [line: 4, column: 27]` |
| `FailureScripts/visibility5` | add | `Syntax Error: Unexpected "Integer Literal" [line: 14, column: 9]` |
| `HelpersFail/helper_error3` | add | `Syntax Error: ")" expected [line: 9, column: 21]` |
| `HelpersFail/helper_error4` | add | `Syntax Error: Class method or constructor expected [line: 11, column: 9]` |
| `HelpersFail/helper_not_implemented` | add | `Syntax Error: Method "Length" of class "THelper" not implemented [line: 3, column: 11]` |
| `HelpersFail/helper_overload_error` | add | `Syntax Error: There is no overloaded version of "Hello" that can be called with these arguments [line: 2, column: 50]` |
| `HelpersFail/static_class_method_self` | add | `Syntax Error: Unknown name "Self" [line: 8, column: 12]` |
| `InterfacesFail/interface_inheritence1` | add | `Syntax Error: Class "TImpDescendent" does not implement interface "IBase" [line: 20, column: 6]` |
| `InterfacesFail/intf_forwarded_not_implem1` | add | `Syntax Error: Interface "IIntf" isn't defined completely [line: 1, column: 23]` |
| `JSONConnectorFail/coalesce_typ` | add | `Syntax Error: There is no accessible member with name "Toto" for type Variant [line: 5, column: 11]` |
| `LambdaFail/no_arrow` | add | `Syntax Error: Unexpected "=>" for a lambda statement [line: 3, column: 21]` |
| `LambdaFail/no_begin` | add | `Syntax Error: End of block expected [line: 4, column: 10]` |
| `OverloadsFail/overload_param_missing` | add | `Syntax Error: Expression expected [line: 9, column: 10]` |
| `PropertyExpressionsFail/interface_property_auto_field` | add | `Syntax Error: Neither "read" nor "write" directive found [line: 3, column: 31]` |
| `PropertyExpressionsFail/invalid_getter` | drop | `Syntax Error: expected ')', got SEMICOLON [line: 5, column: 44]` |
| `PropertyExpressionsFail/invalid_getter2` | drop | `Syntax Error: expected ')', got SEMICOLON [line: 5, column: 56]` |
| `SetOfFail/type_missing` | add | `Syntax Error: Type expected [line: 2, column: 22]` |

Recurring themes inside that list, each worth one change:

- `Unexpected "Integer Literal"` — `property_error6`, `visibility5`.
- Triple-apostrophe string diagnostics — `triple_apos1`, `triple_apos2`.
- Spurious `No arguments expected` on an array helper called with none — `dyn_array1`,
  `dyn_array_setlength2`.
- `argument N to method 'X' of class 'Y' has type …` → `Argument N expects type "X" instead of "Y"`
  — `method_param_error1`, `method_param_error2`.
- Spurious `Undefined variable 'Integer'` for an escaped reserved word — `reserved_escape_empty`,
  `reserved_escape_number`.
- Spurious `expected ')', got SEMICOLON` — `missing_operand1`, `missing_operand3`,
  `PropertyExpressionsFail/invalid_getter`, `invalid_getter2`.
- ~~Missing `Warning: Assignment to FOR-Loop variable` — `for_var_usage`, `for_var_usage2`.~~
  Closed 2026-09-12.

## What this audit closed

**F6, the runtime-mismatch residue, is empty.** Every `FailureScripts` fixture whose expectation is
blank now compiles clean — `div_by_zero_float`, `div_by_zero_int`, `div_by_zero_int2`,
`mod_by_zero_int` — and no fixture in the suite expects a `Runtime Error` line at all.
`for_in_subclass`, the last name on the old list, turned out to be message parity: it reports
`for-in loop variable c has type TChild(TBase), cannot assign TBase(TObject)` where upstream says
`Incompatible types: Cannot assign "TBase" to "TChild"`.
