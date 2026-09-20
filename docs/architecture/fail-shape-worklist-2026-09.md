# `*Fail` message-shape worklist — September 2026

**Measured at:** `7857c64d` (the tooling commit; the compiler under test is `origin/main` =
`b7813fbe`), 2026-09-20. Read-only.

**This is a snapshot.** Five branches were changing diagnostics concurrently while this ran, so
every number below describes `b7813fbe` and nothing later. Regenerate before trusting a count.

The worklist behind [`PLAN.md`](../../PLAN.md) §4 / **F8**: every distinct diagnostic shape the
`*Fail` measurement separates go-dws from DWScript by, with the fixtures each one blocks. Companion
to [`fail-suite-audit-2026-09.md`](fail-suite-audit-2026-09.md), whose *Message-shape inventories*
section lists only the head of each of these two tables; this is the whole tail, plus the emitting
site for each spurious shape.

## Totals

| | |
| --- | --- |
| in-scope fixtures | 1,824 — 1,747 scored (1,200 pass, 547 fail), 77 unscored |
| in-scope `*Fail` failing | **433** — 143 exactly one edit away, 253 within two |
| all in-scope failing | 547 — 156 one edit away, 284 within two |
| missing shapes | **269** distinct, 1,068 lines, 442 fixtures |
| spurious shapes | **393** distinct, 1,113 lines, 427 fixtures |

**Command**

```bash
go build -o bin/dwscript ./cmd/dwscript
go run ./cmd/fixture-report --build=false --allow-stale --cli bin/dwscript \
  --in-scope --classify --list-fails --shape-top 0 --shape-fixtures
```

`--shape-top 0` prints every shape instead of the ranked top 20, and `--shape-fixtures` names the
fixtures each shape blocks. Both are off by default, so the output the harness and the audit quote
is unchanged.

**Scope.** The command classifies every in-scope category, not only the `*Fail` suites, so the
tables below also carry the execution suites' failures (`ArrayPass`, `SimpleScripts`, …). 433 of
the 547 failing fixtures are `*Fail`; 422 of the 442 fixtures with a missing shape and 341 of the
427 with a spurious one.

**Columns.** `lines` is how many diagnostic lines the shape accounts for; `fixtures` is how many
fixtures it appears in, which is the ranking key because the pass rate counts fixtures; `sole` is
how many of those fail on this shape *alone*, which is the yield if the shape is fixed and nothing
else changes; `blocks` names them.

**Shapes are normalised** the way `cmd/fixture-report` normalises them: the `[line: N, column: N]`
suffix is dropped, quoted contents become `"X"` or `'X'`, and digits become `N`. A sentence and the
same sentence about two other types are therefore one shape.

## Missing shapes (expected, never produced)

| lines | fixtures | sole | shape | blocks |
| --- | --- | --- | --- | --- |
| 70 | 63 | 24 | `Syntax Error: "X" expected` | AttributesFail/attribute_incorrect2, FailureScripts/array_error7, FailureScripts/array_index_bracket_missing1, FailureScripts/array_index_bracket_missing2, FailureScripts/array_new, FailureScripts/array_params2, FailureScripts/assert, FailureScripts/at_integer, FailureScripts/cast_string, FailureScripts/class_cast, FailureScripts/class_error2, FailureScripts/class_error3, FailureScripts/class_error4, FailureScripts/class_error8, FailureScripts/class_of, FailureScripts/class_operator5, FailureScripts/const_array2, FailureScripts/const_record1, FailureScripts/const_record4, FailureScripts/constructor_invalid_param, FailureScripts/contracts_error1, FailureScripts/contracts_error3, FailureScripts/contracts_unfinished4, FailureScripts/debugbreak, FailureScripts/dyn_array_setlength1, FailureScripts/empty_body, FailureScripts/end_implementation2, FailureScripts/enum_scoped2, FailureScripts/enums3, FailureScripts/export, FailureScripts/for_error1, FailureScripts/in_operator6, FailureScripts/in_operator8, FailureScripts/include_incorrect, FailureScripts/missing_parenthesis1, FailureScripts/missing_parenthesis2, FailureScripts/missing_semi1, FailureScripts/nested_type1, FailureScripts/new_class6, FailureScripts/new_class7, FailureScripts/params1, FailureScripts/program, FailureScripts/property_description1, FailureScripts/property_reintroduce2, FailureScripts/resourcestring3, FailureScripts/special_funcs1, FailureScripts/swap1, GenericsFail/declaration_params_error1, GenericsFail/implem_mismatch1, GenericsFail/record_constraint1, HelpersFail/helper_error3, InterfacesFail/interface_guid, InterfacesFail/partial_declaration2, InterfacesFail/partial_declaration3, OperatorOverloadFail/operator_overload1, OperatorOverloadFail/operator_overload2, OperatorOverloadFail/operator_overload3, OperatorOverloadFail/operator_overload5, PropertyExpressionsFail/expr_write_readonly_property, PropertyExpressionsFail/missing_reader_bracket, SetOfFail/bracket_left_missing, SetOfFail/bracket_right_missing, SetOfFail/include |
| 38 | 33 | 6 | `Syntax Error: Name expected` | FailureScripts/class_error2, FailureScripts/class_error3, FailureScripts/class_error5, FailureScripts/class_error6, FailureScripts/class_error7, FailureScripts/class_nested, FailureScripts/class_of, FailureScripts/const_4, FailureScripts/constructor_no_name, FailureScripts/empty_body, FailureScripts/enums2, FailureScripts/external1, FailureScripts/for_unfinished1, FailureScripts/for_var_error, FailureScripts/lazy, FailureScripts/method_implem3, FailureScripts/nested_type2, FailureScripts/proc_missing_name, FailureScripts/program, FailureScripts/property_error1, FailureScripts/property_error3, FailureScripts/property_error4, FailureScripts/property_read3, FailureScripts/resourcestring2, FailureScripts/switch_invalid2, GenericsFail/array1-2, GenericsFail/implem_mismatch1, InterfacesFail/method_decl_syntax1, InterfacesFail/partial_declaration2, OperatorOverloadFail/operator_overload5, OverloadsFail/meth_overload_simple, PropertyExpressionsFail/expr_write_readonly_property, PropertyExpressionsFail/missing_reader_bracket |
| 26 | 25 | 11 | `Syntax Error: Expression expected` | FailureScripts/array_assign_add, FailureScripts/as_invalid_right, FailureScripts/case_error6, FailureScripts/contracts_unfinished1, FailureScripts/contracts_unfinished2, FailureScripts/contracts_unfinished3, FailureScripts/default_params4, FailureScripts/dotdot, FailureScripts/enum_byname, FailureScripts/exit_result6, FailureScripts/for_var_error2, FailureScripts/for_var_error3, FailureScripts/ifthenelse_expression3, FailureScripts/include_expr, FailureScripts/inherited3, FailureScripts/invalid_cast, FailureScripts/invalid_cast2, FailureScripts/member_of_void1, FailureScripts/operator1, FailureScripts/raise_error, FailureScripts/special_funcs4, InterfacesFail/method_decl_error2, LambdaFail/invalid_expression, OverloadsFail/overload_param_missing, PropertyExpressionsFail/null_read_expression |
| 54 | 19 | 9 | `Syntax Error: Incompatible types: "X" and "X"` | FailureScripts/array_initialization4, FailureScripts/array_of_proc, FailureScripts/array_of_proc2, FailureScripts/case_error5, FailureScripts/case_range_typecheck, FailureScripts/case_typecheck, FailureScripts/coalesce, FailureScripts/coalesce_dynarray, FailureScripts/const_1, FailureScripts/const_procedure_array, FailureScripts/func_ptr3, FailureScripts/func_ptr4, FailureScripts/func_ptr5, FailureScripts/func_ptr_mismatch, FailureScripts/in_typecheck1, FailureScripts/params1, FailureScripts/property_default1, FailureScripts/swap1, SetOfFail/invalid_operand |
| 37 | 17 | 8 | `Syntax Error: Incompatible types: Cannot assign "X" to "X"` | FailureScripts/array_assign_error3, FailureScripts/array_const, FailureScripts/array_index_bracket_missing, FailureScripts/as_error, FailureScripts/assign_error, FailureScripts/assign_op_incompatible, FailureScripts/coalesce_class, FailureScripts/enums9, FailureScripts/for_in_subclass, FailureScripts/func_ptr1, FailureScripts/multi_dim_dyn_array1, FailureScripts/object_relops, FailureScripts/record_meta, InterfacesFail/assign_intf_from_intf, InterfacesFail/assign_obj_from_intf, InterfacesFail/interface_inheritence2, JSONConnectorFail/autobox |
| 34 | 15 | 1 | `Syntax Error: Method "X" of class "X" not implemented` | FailureScripts/const_expr_1, FailureScripts/constructor_no_name, FailureScripts/external_overload, FailureScripts/member_duplicates, FailureScripts/method_duplicate_override, FailureScripts/method_params, FailureScripts/missing_member1, FailureScripts/new_class3, FailureScripts/record_missing_imple, FailureScripts/reintroduce, FailureScripts/static_class1, HelpersFail/helper_not_implemented, OverloadsFail/hide_virtual, OverloadsFail/overload_missing, OverloadsFail/overloads_not_implem |
| 18 | 15 | 7 | `Syntax Error: Unknown name "X"` | AttributesFail/attribute_incorrect1, FailureScripts/class_field_type, FailureScripts/class_of, FailureScripts/enums, FailureScripts/except_error4, FailureScripts/except_error5, FailureScripts/method2, FailureScripts/method_implem, FailureScripts/missing_class1, FailureScripts/new_array, FailureScripts/property_read3, GenericsFail/declaration_params_error2, HelpersFail/static_class_method_self, InnerClassesFail/sub_outside_scope, InterfacesFail/method_decl_error1 |
| 14 | 13 | 4 | `Syntax Error: Type expected` | FailureScripts/array_static_bounds, FailureScripts/class_error8, FailureScripts/class_operator4, FailureScripts/class_type, FailureScripts/interface_type, FailureScripts/method_params, FailureScripts/open_array, GenericsFail/array1-2, GenericsFail/declaration_params_error1, GenericsFail/record_constraint1, HelpersFail/helper_error2, InterfacesFail/method_decl_syntax2, OperatorOverloadFail/operator_overload2 |
| 20 | 12 | 2 | `Syntax Error: Name "X" already exists` | FailureScripts/class_duplicate_subtype, FailureScripts/enums4, FailureScripts/for_var_overwrite1, FailureScripts/func_external, FailureScripts/nested_type2, FailureScripts/params2, FailureScripts/partial_class5, FailureScripts/result_redefine, FailureScripts/var_ambiguous_in_scope, GenericsFail/declaration_params_error2, HelpersFail/helper_duplicate_member, InterfacesFail/interface_redefine |
| 16 | 12 | 1 | `Syntax Error: Constant expression expected` | FailureScripts/array_range1, FailureScripts/class_nested, FailureScripts/const_1, FailureScripts/const_2, FailureScripts/const_array3, FailureScripts/const_record3, FailureScripts/enums2, FailureScripts/params1, FailureScripts/property_error2, FailureScripts/property_write5, FailureScripts/resourcestring3, FailureScripts/switch_invalid2 |
| 16 | 11 | 5 | `Syntax Error: Cannot assign a value to the left-side argument` | FailureScripts/assign_untyped, FailureScripts/class_const3, FailureScripts/const_1, FailureScripts/const_2, FailureScripts/const_param1, FailureScripts/const_param4, FailureScripts/lazy, FailureScripts/readonly_field, FailureScripts/record_string_field, FailureScripts/self_not_writable, JSONConnectorFail/plus_assign_field |
| 11 | 9 | 3 | `Syntax Error: Class reference expected` | FailureScripts/as_error, FailureScripts/class_operator5, FailureScripts/new_class1, FailureScripts/new_class2, FailureScripts/new_class4, FailureScripts/new_class5, FailureScripts/object_relops, FailureScripts/try_except1, HelpersFail/mixed_helper |
| 9 | 9 | 3 | `Syntax Error: Colon "X" expected` | FailureScripts/array_params1, FailureScripts/case_error4, FailureScripts/class_error6, FailureScripts/const_param3, FailureScripts/const_record3, FailureScripts/property_error11, FailureScripts/property_error2, FailureScripts/record_syntax2, OperatorOverloadFail/operator_overload4 |
| 11 | 8 | 0 | `Syntax Error: There is no overloaded version of "X" that can be called with these arguments` | FailureScripts/case_error5, FailureScripts/func_toomanyargs, FailureScripts/strict_parameter_type, HelpersFail/helper_overload_error, OverloadsFail/meth_overload_hide, OverloadsFail/meth_private_public, OverloadsFail/overload_proc_value, OverloadsFail/overload_simple |
| 15 | 7 | 1 | `Warning: Unreachable code` | FailureScripts/break_continue, FailureScripts/class_cast, FailureScripts/raise_error, FailureScripts/unreachable, FailureScripts/unreachable_case_of, SimpleScripts/exceptions2, SimpleScripts/exit |
| 12 | 7 | 2 | `Syntax Error: String expected` | FailureScripts/assert, FailureScripts/class_external, FailureScripts/enum_byname, FailureScripts/external2, FailureScripts/property_description1, FailureScripts/resourcestring1, InterfacesFail/interface_guid |
| 9 | 7 | 3 | `Hint: "X" does not match case of declaration ("X")` | ArrayPass/array_of_rec_add_create, FailureScripts/hint_pedantic, FailureScripts/params3, GenericsFail/implem_mismatch1, LambdaPass/simple_func, SimpleScripts/class_var_dyn2, SimpleScripts/string_builtin_methods |
| 9 | 7 | 2 | `Syntax Error: Class "X" isn't defined completely` | FailureScripts/class_circular, FailureScripts/class_loop, FailureScripts/class_not_defined, FailureScripts/classname_already_exists, FailureScripts/missing_class2, FailureScripts/partial_class1, FailureScripts/partial_class5 |
| 12 | 6 | 1 | `Syntax Error: Incompatible parameter types - "X" expected (instead of "X")` | AssociativeFail/delete, FailureScripts/array_assign_add, FailureScripts/array_plus_assign, GenericsFail/constraint1, GenericsFail/record_constraint1, SetOfFail/invalid_operand |
| 7 | 6 | 0 | `Syntax Error: Expected N parameters (instead of N)` | FailureScripts/declaration_mismatch2, GenericsFail/array1-2, JSONConnectorFail/add_error, JSONConnectorFail/extend, JSONConnectorFail/parameters_check, OperatorOverloadFail/operator_overload1 |
| 7 | 6 | 0 | `Syntax Error: There is already a method with name "X"` | FailureScripts/empty_body, FailureScripts/member_duplicates, FailureScripts/method_implem, GenericsFail/implem_mismatch1, InterfacesFail/method_decl_syntax1, OverloadsFail/overload_missing |
| 6 | 6 | 2 | `Syntax Error: Function expected` | FailureScripts/array_error4, FailureScripts/dyn_array3, FailureScripts/func_ptr2, FailureScripts/invalid_cast3, FailureScripts/property_error2, OperatorOverloadFail/operator_overload5 |
| 12 | 5 | 0 | `Syntax Error: Assignment's right-side-argument has no return type` | FailureScripts/array_plus_assign, FailureScripts/assign_untyped, FailureScripts/func_ptr1, FailureScripts/ifthenelse_expression3, GenericsFail/binop_constraint |
| 9 | 5 | 0 | `Syntax Error: Field/method "X" has an incompatible type` | FailureScripts/property_error1, FailureScripts/property_error3, FailureScripts/property_read3, FailureScripts/record_empty, InterfacesFail/interface_properties |
| 8 | 5 | 0 | `Hint: Name "X" could be ambiguous in its scope context` | FailureScripts/class_duplicate_subtype, FailureScripts/for_var_overwrite1, FailureScripts/var_ambiguous_in_scope, SimpleScripts/ifthenelse_optimize3, SimpleScripts/nested_declarations |
| 8 | 5 | 2 | `Syntax Error: Invalid Instruction - function or assignment expected` | FailureScripts/array_method1, FailureScripts/const_expr_1, FailureScripts/passing_prop_var, FailureScripts/passing_prop_var2, FailureScripts/property_unused |
| 8 | 5 | 0 | `Syntax Error: Method "X" has incompatible parameters` | FailureScripts/array_params2, FailureScripts/property_error3, FailureScripts/property_error4, InterfacesFail/interface_properties, InterfacesFail/method_decl_error1 |
| 8 | 5 | 0 | `Warning: Constant condition` | FailureScripts/contracts_error1, FailureScripts/contracts_error3, FailureScripts/contracts_precondition, FailureScripts/contracts_unfinished3, FailureScripts/contracts_warnings |
| 8 | 5 | 0 | `Warning: Method "X" overlaps a virtual method` | FailureScripts/final, FailureScripts/reintroduce, FailureScripts/virtual1, OverloadsFail/hide_virtual, SimpleScripts/constructor_overload |
| 6 | 5 | 1 | `Syntax Error: Class method or constructor expected` | FailureScripts/class_var_dyn2, FailureScripts/func_ptr6, HelpersFail/helper_error4, HelpersFail/helper_static, HelpersFail/integer_helper |
| 6 | 5 | 0 | `Syntax Error: Parameter N - Type "X" expected (instead of "X")` | FailureScripts/property_error4, JSONConnectorFail/add_error, JSONConnectorFail/extend, JSONConnectorFail/parameters_check, OperatorOverloadFail/operator_overload1 |
| 6 | 5 | 1 | `Syntax Error: unexpected "X"` | FailureScripts/at_integer, FailureScripts/dyn_array3, FailureScripts/field_init1, FailureScripts/func_ptr6, SetOfFail/invalid_operand |
| 5 | 5 | 5 | `Syntax Error: "X" expected but "X" found` | FailureScripts/block_unfinished3, FailureScripts/block_unfinished4, FailureScripts/double_finalization, FailureScripts/else_unexpected1, FailureScripts/else_unexpected3 |
| 5 | 5 | 4 | `Syntax Error: PROCEDURE or FUNCTION expected` | FailureScripts/class_class, FailureScripts/class_error1, FailureScripts/legacy_proc_of_object, FailureScripts/record_syntax1, HelpersFail/helper_error5 |
| 9 | 4 | 1 | `Syntax Error: Range start and range stop are of incompatible types: "X" and "X"` | FailureScripts/array_range2, FailureScripts/case_error5, FailureScripts/case_range_mismatch, FailureScripts/in_operator8 |
| 8 | 4 | 0 | `Hint: "X" parameter is a reference type passed as VAR, but never written to` | FailureScripts/hint_reference_var_params, FailureScripts/self_not_writable, FailureScripts/var_object, OperatorOverloadFail/operator_overload5 |
| 8 | 4 | 0 | `Syntax Error: There is already a field with name "X"` | FailureScripts/member_duplicates, FailureScripts/partial_class1, FailureScripts/partial_class3, FailureScripts/record_syntax2 |
| 7 | 4 | 1 | `Syntax Error: "X" is not a method of class "X"` | FailureScripts/class_missing_decl, FailureScripts/method_implem2, FailureScripts/method_implem4, FailureScripts/record_missing_imple |
| 6 | 4 | 1 | `Syntax Error: Argument N expects type "X" instead of "X"` | FailureScripts/coalesce, FailureScripts/strict_parameter_type, FailureScripts/var_object, OverloadsFail/forwards_unit |
| 6 | 4 | 1 | `Syntax Error: Declaration should be "X"` | FailureScripts/declaration_mismatch1, FailureScripts/declaration_mismatch3, FailureScripts/method_implem4, FailureScripts/nested_method |
| 5 | 4 | 0 | `Hint: Result is never used` | FailureScripts/method_implem5, HelpersFail/helper_static, OverloadsFail/overload_simple, SetOfFail/test_non_variable |
| 5 | 4 | 1 | `Syntax Error: Bound isn't of an ordinal type` | AssociativeFail/syntax1, FailureScripts/array_error4, FailureScripts/array_error7, FailureScripts/array_static_bounds |
| 5 | 4 | 0 | `Syntax Error: Procedure expected` | FailureScripts/class_error7, FailureScripts/property_error2, FailureScripts/property_error3, InterfacesFail/interface_properties |
| 4 | 4 | 0 | `Hint: Overloaded method "X" should be marked with the "X" directive` | FailureScripts/member_duplicates, FailureScripts/method_duplicate_override, FailureScripts/partial_class1, OverloadsFail/overload_missing |
| 4 | 4 | 2 | `Syntax Error: Dot "X" expected` | FailureScripts/end_implementation2, FailureScripts/end_implementation3, FailureScripts/method_implem6, GenericsFail/implem_mismatch1 |
| 4 | 4 | 1 | `Syntax Error: END expected` | FailureScripts/record_syntax2, FailureScripts/try_except1, HelpersFail/helper_scopes1, InterfacesFail/partial_declaration |
| 4 | 4 | 2 | `Syntax Error: Integer expected` | FailureScripts/for_step, FailureScripts/for_var_usage4, FailureScripts/internal_unsupported, FailureScripts/passing_prop_var |
| 4 | 4 | 1 | `Syntax Error: More arguments expected` | FailureScripts/contracts_error2, FailureScripts/func_ptr_mismatch, HelpersFail/function_helper, HelpersFail/helper_explicit |
| 10 | 3 | 2 | `Syntax Error: Numerical operand expected` | FailureScripts/cast_base_type, FailureScripts/neg_type, FailureScripts/plus_non_numeric |
| 9 | 3 | 1 | `Syntax Error: Argument N (a) cannot be passed as Var-parameter` | FailureScripts/lazy, FailureScripts/passing_const_var, FailureScripts/passing_prop_var |
| 8 | 3 | 1 | `Syntax Error: Argument N (v) cannot be passed as Var-parameter` | FailureScripts/const_param2, FailureScripts/passing_const_var2, FailureScripts/passing_prop_var2 |
| 8 | 3 | 1 | `Syntax Error: Symbol "X" cannot be captured` | LambdaFail/global_ref, LambdaFail/invalid_expression, LambdaFail/no_capture |
| 7 | 3 | 1 | `Syntax Error: Object expected` | FailureScripts/implements_error, FailureScripts/in_operator5, FailureScripts/object_relops |
| 7 | 3 | 0 | `Warning: Constructor invoked on instance outside of constructor` | FailureScripts/constructor_on_instance, FailureScripts/missing_class1, SimpleScripts/virtual_constructor2 |
| 6 | 3 | 0 | `Syntax Error: Incompatible operands` | FailureScripts/assign_op_incompatible, FailureScripts/coalesce, FailureScripts/func_ptr1 |
| 6 | 3 | 1 | `Syntax Error: Variable expected` | FailureScripts/for_loopvar1, FailureScripts/swap1, SetOfFail/test_non_variable |
| 5 | 3 | 0 | `Syntax Error: Boolean expected` | FailureScripts/assert, FailureScripts/contracts_error2, FailureScripts/switch_invalid2 |
| 4 | 3 | 2 | `Syntax Error: DO expected` | FailureScripts/for_error2, FailureScripts/for_in_str1, SetOfFail/for_in_set_missing_do |
| 4 | 3 | 1 | `Syntax Error: Integer expression expected` | FailureScripts/enums2, FailureScripts/multi_dim_dyn_array1, FailureScripts/string_get |
| 4 | 3 | 1 | `Syntax Error: Interface "X" isn't defined completely` | InterfacesFail/interface_redefine, InterfacesFail/intf_forwarded_not_implem1, InterfacesFail/intf_forwarded_not_implem2 |
| 4 | 3 | 1 | `Syntax Error: Invalid Operands` | FailureScripts/operator1, FailureScripts/ord, FailureScripts/special_funcs2 |
| 4 | 3 | 1 | `Syntax Error: Object reference needed to read/write an object field` | FailureScripts/class_property2, FailureScripts/class_var_dyn1, FailureScripts/property_error7 |
| 4 | 3 | 0 | `Syntax Error: Too many arguments` | FailureScripts/func_toomanyargs, FailureScripts/new_class4, HelpersFail/function_helper |
| 4 | 3 | 1 | `Warning: Property writer does nothing` | PropertyExpressionsFail/expr_write_readonly_property, PropertyExpressionsFail/missing_reader_bracket, PropertyExpressionsFail/null_write_expression |
| 3 | 3 | 3 | `Syntax Error: "X" or "X" expected but "X" found` | FailureScripts/double_initialization, FailureScripts/final_dot1, FailureScripts/final_dot2 |
| 3 | 3 | 1 | `Syntax Error: Class "X" does not implement interface "X"` | InterfacesFail/assign_intf_from_intf, InterfacesFail/assign_intf_from_obj, InterfacesFail/interface_inheritence1 |
| 3 | 3 | 1 | `Syntax Error: End of block expected` | FailureScripts/for_empty, FailureScripts/for_var_usage5, LambdaFail/no_begin |
| 3 | 3 | 0 | `Syntax Error: Function type expected` | FailureScripts/legacy_proc_of_object, FailureScripts/method_implem5, FailureScripts/nested_method |
| 3 | 3 | 2 | `Syntax Error: Method "X" not found in ancestor class` | FailureScripts/class_missing_decl, FailureScripts/inherited4, FailureScripts/inherited6 |
| 3 | 3 | 0 | `Syntax Error: Result type should be "X"` | FailureScripts/declaration_mismatch1, FailureScripts/method_implem5, OperatorOverloadFail/operator_overload1 |
| 8 | 2 | 1 | `Warning: Incompatible types: "X" and "X"` | FailureScripts/is_always_false, SimpleScripts/oop_is_as |
| 6 | 2 | 0 | `Syntax Error: Class "X" is static, no instances allowed` | FailureScripts/static_class1, JSONConnectorFail/create_static |
| 6 | 2 | 0 | `Syntax Error: Name "X" is reserved` | FailureScripts/classname_already_exists, FailureScripts/special_funcs3 |
| 6 | 2 | 0 | `Syntax Error: Overload of "X" will be ambiguous with a previously declared version` | OverloadsFail/meth_overload_simple, OverloadsFail/overload_simple |
| 5 | 2 | 1 | `Syntax Error: No parameters expected` | FailureScripts/string_builtin_methods1, JSONConnectorFail/parameters_check |
| 5 | 2 | 1 | `Warning: "X" expected` | FailureScripts/exit_result6, FailureScripts/proc_semi_warning |
| 4 | 2 | 1 | `Hint: Case range condition lower bound is greater than higher bound` | FailureScripts/case_range_mismatch, FailureScripts/in_operator_order |
| 4 | 2 | 0 | `Hint: Redundant specifier, visibility is already "X"` | FailureScripts/class_visibility_redundant, HelpersFail/helper_scopes1 |
| 4 | 2 | 2 | `Syntax Error: Local procedure/function cannot be used as delegate` | FailureScripts/func_ptr_local, LambdaFail/no_local_func |
| 3 | 2 | 0 | `Hint: OF OBJECT modifier is legacy and ignored` | FailureScripts/legacy_proc_of_object, SimpleScripts/func_ptr_field_no_param |
| 3 | 2 | 1 | `Syntax Error: Cannot read a write only property` | FailureScripts/inherited5, PropertyExpressionsFail/read_write_other_property |
| 3 | 2 | 0 | `Syntax Error: Interface expected` | FailureScripts/implements_error, HelpersFail/mixed_helper |
| 3 | 2 | 1 | `Syntax Error: Invalid argument combination` | FailureScripts/ifthenelse_expression3, FailureScripts/ifthenelse_expression4 |
| 3 | 2 | 0 | `Syntax Error: Parameter N (a) - Var-parameter expected` | FailureScripts/array_params2, FailureScripts/declaration_mismatch1 |
| 3 | 2 | 0 | `Warning: "X" has been deprecated` | FailureScripts/class_deprecated, FunctionsMath/randseed |
| 3 | 2 | 2 | `Warning: Infinite loop` | FailureScripts/infinite_loop, FailureScripts/loop_infinite |
| 2 | 2 | 1 | _(empty message text)_ | InterfacesPass/intf_in_record, SimpleScripts/exception_nested_call2 |
| 2 | 2 | 0 | `Hint: Property "X" reintroduced a method, you should remove empty brackets ()` | FailureScripts/property_reintroduce2, SimpleScripts/property_reintroduce |
| 2 | 2 | 0 | `Object not instantiated` | SimpleScripts/class_property, SimpleScripts/class_self |
| 2 | 2 | 0 | `Syntax Error: "X" is not a class` | FailureScripts/class_error7, FailureScripts/class_of |
| 2 | 2 | 0 | `Syntax Error: "X" is not an interface` | FailureScripts/class_error4, InterfacesFail/partial_declaration2 |
| 2 | 2 | 1 | `Syntax Error: An overload already exists for this operator and types` | OperatorOverloadFail/operator_overload1, OperatorOverloadFail/operator_overload6 |
| 2 | 2 | 1 | `Syntax Error: Argument N expects type "X"` | FailureScripts/assign_untyped, FailureScripts/func_ptr_var_param |
| 2 | 2 | 0 | `Syntax Error: BEGIN expected` | FailureScripts/export, FailureScripts/virtual2 |
| 2 | 2 | 1 | `Syntax Error: Boolean or integer operand expected` | FailureScripts/cast_base_type, FailureScripts/not_untyped |
| 2 | 2 | 0 | `Syntax Error: CLASS expected` | FailureScripts/partial_class2, FailureScripts/partial_class3 |
| 2 | 2 | 0 | `Syntax Error: Class "X" is static, instantiation not allowed` | FailureScripts/static_class1, JSONConnectorFail/create_static |
| 2 | 2 | 0 | `Syntax Error: Field/method "X" not found` | FailureScripts/class_property2, FailureScripts/property_error1 |
| 2 | 2 | 0 | `Syntax Error: Inherited method "X" isn't virtual. "X" not applicable` | FailureScripts/reintroduce, FailureScripts/virtual1 |
| 2 | 2 | 0 | `Syntax Error: Method "X" has an incompatible parameter type` | FailureScripts/class_operator4, FailureScripts/in_operator3 |
| 2 | 2 | 2 | `Syntax Error: OF expected` | FailureScripts/case_error2, SetOfFail/of_missing |
| 2 | 2 | 1 | `Syntax Error: Overloadable operator expected` | FailureScripts/class_operator3, OperatorOverloadFail/operator_overload1 |
| 2 | 2 | 0 | `Syntax Error: Parameter N (a) - Const-parameter expected` | FailureScripts/array_params2, FailureScripts/declaration_mismatch1 |
| 2 | 2 | 0 | `Syntax Error: Parameters expected` | FailureScripts/array_params1, HelpersFail/function_helper |
| 2 | 2 | 0 | `Syntax Error: Record has no field members` | FailureScripts/record_empty, OverloadsFail/meth_overload_simple |
| 2 | 2 | 1 | `Syntax Error: Single parameter expected` | FailureScripts/class_operator1, FailureScripts/in_operator3 |
| 2 | 2 | 1 | `Syntax Error: THEN expected` | FailureScripts/ifthenelse_expression1, FailureScripts/operator1 |
| 2 | 2 | 0 | `Syntax Error: TO or DOWNTO expected` | FailureScripts/for_error2, FailureScripts/for_var_error2 |
| 2 | 2 | 2 | `Syntax Error: Unexpected "X"` | FailureScripts/else_unexpected2, OverloadsFail/func_ptr_overload |
| 2 | 2 | 1 | `Syntax Error: Unexpected END` | FailureScripts/end_implementation1, FailureScripts/property_description1 |
| 2 | 2 | 0 | `Syntax Error: Unexpected method implementation` | FailureScripts/method_implem7, FailureScripts/nested_method |
| 2 | 2 | 0 | `Syntax Error: const parameter cannot have a default value` | FailureScripts/declaration_mismatch1, FailureScripts/params1 |
| 8 | 1 | 1 | `Syntax Error: Too many indices` | FailureScripts/array_index_extra |
| 6 | 1 | 0 | `Syntax Error: Type could not be inferenced` | FailureScripts/nil_type_inference |
| 4 | 1 | 1 | `Syntax Error: Default value required` | FailureScripts/default_params1 |
| 4 | 1 | 0 | `Syntax Error: Modifiers do not match previous "X" declaration of class` | FailureScripts/partial_class4 |
| 4 | 1 | 0 | `Syntax Error: Parameter N (a) - Value-parameter expected` | FailureScripts/declaration_mismatch1 |
| 4 | 1 | 0 | `Warning: Operator "X" is ambiguous` | SimpleScripts/plus_plus_minus_minus |
| 3 | 1 | 0 | `Syntax Error: Cannot mark "X" without overriding` | FailureScripts/final |
| 3 | 1 | 0 | `Syntax Error: Invalid argument type` | FailureScripts/internal_unsupported |
| 3 | 1 | 0 | `Syntax Error: No result type expected` | FailureScripts/proc_with_result |
| 3 | 1 | 0 | `Syntax Error: Ordinal expression expected` | FailureScripts/array_range1 |
| 3 | 1 | 0 | `Syntax Error: Parameter list doesn't match the inherited method` | FailureScripts/virtual1 |
| 3 | 1 | 0 | `Syntax Error: There is already a property with name "X"` | FailureScripts/member_duplicates |
| 2 | 1 | 0 | `Hint: Assigning a to itself` | FailureScripts/self_assign |
| 2 | 1 | 0 | `Hint: Private method "X" declared but never used` | OverloadsFail/meth_private_public |
| 2 | 1 | 0 | `Hint: REFERENCE TO modifier is legacy and ignored` | FailureScripts/legacy_proc_of_object |
| 2 | 1 | 0 | `Hint: Unit "X" already declared in interface section` | FailureScripts/duplicate_uses |
| 2 | 1 | 0 | `Hint: Unit "X" redeclared` | FailureScripts/duplicate_uses |
| 2 | 1 | 0 | `Syntax Error: "X" is declared "X", no implementation allowed` | FailureScripts/method_implem8 |
| 2 | 1 | 0 | `Syntax Error: "X" outside of loop` | FailureScripts/break_continue |
| 2 | 1 | 0 | `Syntax Error: Anonymous methods not allowed here` | PropertyExpressionsFail/external_class |
| 2 | 1 | 0 | `Syntax Error: Bound isn't a constant expression` | FailureScripts/array_static_bounds |
| 2 | 1 | 0 | `Syntax Error: Cannot use combined assignment on property` | FailureScripts/property_compound |
| 2 | 1 | 0 | `Syntax Error: Direct "X" or "X" not allowed in "X" block` | FailureScripts/break_in_finally |
| 2 | 1 | 0 | `Syntax Error: Helpers not allowed for delegates or function pointers` | HelpersFail/helper_of_delegate |
| 2 | 1 | 0 | `Syntax Error: Include item expected` | FailureScripts/include_expr |
| 2 | 1 | 0 | `Syntax Error: Method "X" marked as final. "X" not applicable` | FailureScripts/final |
| 2 | 1 | 0 | `Syntax Error: OF OBJECT expected` | FailureScripts/legacy_proc_of_object |
| 2 | 1 | 0 | `Syntax Error: Parameter N (i) - default value at implementation does not match declaration or forward` | FailureScripts/default_params2 |
| 2 | 1 | 0 | `Syntax Error: Parameter N - Var-parameter forbidden` | OperatorOverloadFail/operator_overload5 |
| 2 | 1 | 0 | `Syntax Error: Print is not a Type` | GenericsFail/declaration_params_error3 |
| 2 | 1 | 0 | `Syntax Error: Result type doesn't match the inherited method` | FailureScripts/virtual1 |
| 2 | 1 | 0 | `Syntax Error: Subclass of TCustomAttribute expected` | AttributesFail/attribute_incorrect1 |
| 2 | 1 | 0 | `Syntax Error: There is no overloaded version of "X" declared with these arguments` | OverloadsFail/overload_missing |
| 2 | 1 | 1 | `Syntax Error: Unexpected implementation in interface section` | FailureScripts/interface_implementation |
| 2 | 1 | 0 | `Warning: "X" has been deprecated: old stuff` | FailureScripts/class_deprecated |
| 1 | 1 | 0 | `Error: Trying to create an instance of an abstract class` | FailureScripts/static_class1 |
| 1 | 1 | 0 | `Error: Trying to use an abstract class` | FailureScripts/static_class1 |
| 1 | 1 | 0 | `Hint: "X" is meaningless for external functions` | FailureScripts/class_external |
| 1 | 1 | 0 | `Hint: Assigning TTest.F to itself` | FailureScripts/self_assign |
| 1 | 1 | 0 | `Hint: Assigning f to itself` | FailureScripts/self_assign |
| 1 | 1 | 0 | `Hint: Assigning s to itself` | FailureScripts/assign_error |
| 1 | 1 | 0 | `Hint: Constant Instruction - has no effect` | PropertyExpressionsFail/expr_write_readonly_property |
| 1 | 1 | 0 | `Hint: Private field "X" declared but never used` | OverloadsFail/overloads_not_implem |
| 1 | 1 | 1 | `Hint: Private virtual methods cannot be overridden` | FailureScripts/virtual_private |
| 1 | 1 | 0 | `Hint: Redundant "X" in clause of a case..of` | FailureScripts/case_of_else |
| 1 | 1 | 1 | `Hint: Variable "X" declared but not used` | FailureScripts/var_block1 |
| 1 | 1 | 0 | `Runtime Error: User defined exception: TEST` | SimpleScripts/exception_nested_call2 |
| 1 | 1 | 0 | `Syntax Error: "X" already has "X" as default property` | InterfacesFail/interface_properties |
| 1 | 1 | 0 | `Syntax Error: "X" is not a property` | FailureScripts/property_promotion |
| 1 | 1 | 0 | `Syntax Error: "X" is not external` | FailureScripts/class_external |
| 1 | 1 | 0 | `Syntax Error: "X" is only valid for virtual methods` | FailureScripts/missing_member1 |
| 1 | 1 | 0 | `Syntax Error: "X" not allowed in "X" block` | FailureScripts/break_in_finally |
| 1 | 1 | 0 | `Syntax Error: "X" not applicable between class and instance methods` | FailureScripts/virtual1 |
| 1 | 1 | 1 | `Syntax Error: "X" only allowed in methods` | FailureScripts/inherited1 |
| 1 | 1 | 1 | `Syntax Error: "X" or "X" expected but identifier found` | FailureScripts/block_unfinished1 |
| 1 | 1 | 1 | `Syntax Error: Ambiguous matching overloads of "X"` | FailureScripts/func_ptr_constant_ambiguous |
| 1 | 1 | 0 | `Syntax Error: Anonymous class not allowed by compiler options` | FailureScripts/class_nested |
| 1 | 1 | 0 | `Syntax Error: Argument N (o) cannot be passed as Var-parameter` | FailureScripts/self_not_writable |
| 1 | 1 | 0 | `Syntax Error: Attribute constructor expected` | AttributesFail/attribute_incorrect1 |
| 1 | 1 | 1 | `Syntax Error: Binary digit expected (found "X")` | FailureScripts/binary_literal1 |
| 1 | 1 | 0 | `Syntax Error: Cannot cast "X" as "X"` | FailureScripts/as_error |
| 1 | 1 | 0 | `Syntax Error: Cannot cast this type to "X"` | FailureScripts/cast_base_type |
| 1 | 1 | 0 | `Syntax Error: Cannot demote property visibility` | FailureScripts/property_promotion |
| 1 | 1 | 0 | `Syntax Error: Cannot override a function with a method` | FailureScripts/virtual1 |
| 1 | 1 | 0 | `Syntax Error: Cannot override a procedure with a constructor` | FailureScripts/virtual1 |
| 1 | 1 | 0 | `Syntax Error: Cannot set a value for a read-only property` | PropertyExpressionsFail/expr_write_readonly_property |
| 1 | 1 | 0 | `Syntax Error: Class "X" already defined` | FailureScripts/classname_already_exists |
| 1 | 1 | 0 | `Syntax Error: Class "X" already has "X" as default constructor` | FailureScripts/default_constructor |
| 1 | 1 | 0 | `Syntax Error: Class "X" has no default constructor` | FailureScripts/external3 |
| 1 | 1 | 1 | `Syntax Error: Class "X" has no default property` | FailureScripts/inherited7 |
| 1 | 1 | 0 | `Syntax Error: Class "X" is sealed, inheriting is not allowed` | FailureScripts/sealed |
| 1 | 1 | 0 | `Syntax Error: Class ancestor does not match with previous declaration` | FailureScripts/partial_class3 |
| 1 | 1 | 0 | `Syntax Error: Class name expected` | FailureScripts/method1 |
| 1 | 1 | 1 | `Syntax Error: Class operator already defined for type "X"` | FailureScripts/class_operator2 |
| 1 | 1 | 0 | `Syntax Error: Constant Instruction - has no effect` | FailureScripts/class_const4 |
| 1 | 1 | 0 | `Syntax Error: Dangling attribute declaration` | AttributesFail/attribute_incorrect2 |
| 1 | 1 | 0 | `Syntax Error: Declaration should start with "X"` | FailureScripts/method_implem4 |
| 1 | 1 | 0 | `Syntax Error: Declaration shouldn't start with "X"` | OverloadsFail/overload_missing |
| 1 | 1 | 0 | `Syntax Error: Destructor can only be invoked on instance` | FailureScripts/func_ptr5 |
| 1 | 1 | 0 | `Syntax Error: Enumeration element overflow` | FailureScripts/enum_flags_overflow |
| 1 | 1 | 0 | `Syntax Error: Exception object expected` | FailureScripts/raise_error |
| 1 | 1 | 0 | `Syntax Error: External classes must inherit from an external class or Object` | FailureScripts/partial_class4 |
| 1 | 1 | 0 | `Syntax Error: External properties can only have zero or one argument` | FailureScripts/external_property |
| 1 | 1 | 0 | `Syntax Error: External properties require a type` | FailureScripts/external_property |
| 1 | 1 | 1 | `Syntax Error: FOR expected` | HelpersFail/helper_error1 |
| 1 | 1 | 0 | `Syntax Error: Field "X" is readonly` | FailureScripts/readonly_field |
| 1 | 1 | 0 | `Syntax Error: Field has already been set` | FailureScripts/const_record2 |
| 1 | 1 | 0 | `Syntax Error: Flags enumerations cannot have user values` | FailureScripts/enum_scoped2 |
| 1 | 1 | 1 | `Syntax Error: For loop control variable must be simple local variable` | FailureScripts/for_var_usage3 |
| 1 | 1 | 1 | `Syntax Error: Function or value expected` | FailureScripts/contracts_old |
| 1 | 1 | 0 | `Syntax Error: Generic parameters list expected` | GenericsFail/binop_constraint |
| 1 | 1 | 0 | `Syntax Error: Helpers do not supported "X" visibility specifier` | HelpersFail/helper_scopes1 |
| 1 | 1 | 1 | `Syntax Error: IN expected` | FailureScripts/in_operator4 |
| 1 | 1 | 0 | `Syntax Error: Include item "X" unknown` | FailureScripts/include_incorrect |
| 1 | 1 | 0 | `Syntax Error: Input data of invalid size: N instead of N` | FailureScripts/string_set |
| 1 | 1 | 0 | `Syntax Error: Interface "X" already defined` | InterfacesFail/interface_redefine |
| 1 | 1 | 0 | `Syntax Error: Interface "X" already implemented` | FailureScripts/class_error5 |
| 1 | 1 | 0 | `Syntax Error: Invalid Exit argument` | FailureScripts/exit_result6 |
| 1 | 1 | 0 | `Syntax Error: Invalid const type "X"` | FailureScripts/const_3 |
| 1 | 1 | 0 | `Syntax Error: Invalid const type "X" expected "X"` | FailureScripts/const_record3 |
| 1 | 1 | 1 | `Syntax Error: Invalid integer constant "X"` | FailureScripts/binary_literal2 |
| 1 | 1 | 0 | `Syntax Error: Invalid type "X" for function result` | FailureScripts/in_operator3 |
| 1 | 1 | 0 | `Syntax Error: Invalid typecast` | FailureScripts/cast_base_type |
| 1 | 1 | 1 | `Syntax Error: Lazy parameter cannot be a function pointer` | FailureScripts/lazy_func_ptr |
| 1 | 1 | 1 | `Syntax Error: Member symbol "X" is not visible from this scope` | FailureScripts/class_const2 |
| 1 | 1 | 0 | `Syntax Error: Method "X" isn't overlapping a virtual method` | FailureScripts/reintroduce |
| 1 | 1 | 0 | `Syntax Error: Method "X" not found in connector "X"` | JSONConnectorFail/add_error |
| 1 | 1 | 1 | `Syntax Error: Missing matching method "X" for interface "X"` | InterfacesFail/implement_interface1 |
| 1 | 1 | 0 | `Syntax Error: Name expected after "X"` | FailureScripts/inherited2 |
| 1 | 1 | 0 | `Syntax Error: Name of include file expected` | FailureScripts/include_expr |
| 1 | 1 | 0 | `Syntax Error: Neither "X" nor "X" directive found` | PropertyExpressionsFail/interface_property_auto_field |
| 1 | 1 | 0 | `Syntax Error: No available specialization of operator "X" for types "X" and "X"` | GenericsFail/binop_constraint |
| 1 | 1 | 0 | `Syntax Error: No method "X" found in class: "X" not applicable` | FailureScripts/virtual2 |
| 1 | 1 | 1 | `Syntax Error: No result required` | FailureScripts/exit_result2 |
| 1 | 1 | 1 | `Syntax Error: Not a method` | FailureScripts/property_reintroduce1 |
| 1 | 1 | 1 | `Syntax Error: Number, point or exponent expected (found "X")` | FailureScripts/for_var_error4 |
| 1 | 1 | 0 | `Syntax Error: Only a constructor can be marked as default` | FailureScripts/default_constructor |
| 1 | 1 | 0 | `Syntax Error: Only non-virtual class methods can be marked as static` | HelpersFail/helper_static |
| 1 | 1 | 0 | `Syntax Error: Overload not allowed` | FailureScripts/external_overload |
| 1 | 1 | 0 | `Syntax Error: Parameter N (a) - default value at implementation does not match declaration or forward` | FailureScripts/declaration_mismatch1 |
| 1 | 1 | 0 | `Syntax Error: Parameter N - Name "X" expected` | FailureScripts/declaration_mismatch2 |
| 1 | 1 | 0 | `Syntax Error: Preconditions must be defined in the root method only` | FailureScripts/contracts_precondition |
| 1 | 1 | 0 | `Syntax Error: Previous declaration of class was not "X"` | FailureScripts/partial_class2 |
| 1 | 1 | 1 | `Syntax Error: Property cannot be read-accessed` | PropertyExpressionsFail/read_self |
| 1 | 1 | 0 | `Syntax Error: Range is too large` | FailureScripts/array_range1 |
| 1 | 1 | 1 | `Syntax Error: Record type "X" is not fully defined` | FailureScripts/record_recursive2 |
| 1 | 1 | 0 | `Syntax Error: Record type expected` | HelpersFail/mixed_helper |
| 1 | 1 | 0 | `Syntax Error: Simple type expected` | FailureScripts/cast_string |
| 1 | 1 | 0 | `Syntax Error: Symbol "X" has an incompatible type` | FailureScripts/property_error4 |
| 1 | 1 | 0 | `Syntax Error: T expected but u found` | GenericsFail/implem_mismatch1 |
| 1 | 1 | 0 | `Syntax Error: TBug is not a Type` | HelpersFail/helper_as_type |
| 1 | 1 | 0 | `Syntax Error: TDummy is not a Type` | HelpersFail/helper_as_type |
| 1 | 1 | 0 | `Syntax Error: TO expected` | FailureScripts/legacy_proc_of_object |
| 1 | 1 | 0 | `Syntax Error: There is already a class const with name "X"` | FailureScripts/property_error11 |
| 1 | 1 | 0 | `Syntax Error: There is already a class variable with name "X"` | FailureScripts/property_error11 |
| 1 | 1 | 0 | `Syntax Error: There is already a forward declaration of the "X" class` | FailureScripts/classname_already_exists |
| 1 | 1 | 0 | `Syntax Error: There is already a forward declaration of the "X" interface` | InterfacesFail/interface_redefine |
| 1 | 1 | 0 | `Syntax Error: There is no accessible member with name "X" for type JSON` | JSONConnectorFail/create_static |
| 1 | 1 | 1 | `Syntax Error: There is no accessible member with name "X" for type TMySet` | SetOfFail/invalid_method |
| 1 | 1 | 0 | `Syntax Error: There is no accessible member with name "X" for type TRec` | FailureScripts/const_record2 |
| 1 | 1 | 0 | `Syntax Error: There is no accessible member with name "X" for type TSub` | FailureScripts/property_promotion |
| 1 | 1 | 0 | `Syntax Error: There is no accessible member with name "X" for type Variant` | JSONConnectorFail/coalesce_typ |
| 1 | 1 | 0 | `Syntax Error: There is no accessible member with name "X" for type class of TClassA` | FailureScripts/class_not_defined |
| 1 | 1 | 1 | `Syntax Error: There is no accessible member with name "X" for type class of TTest` | FailureScripts/class_var_scope1 |
| 1 | 1 | 1 | `Syntax Error: There is no accessible method with name "X" for type TElement` | FailureScripts/enums8 |
| 1 | 1 | 1 | `Syntax Error: There is no accessible method with name "X" for type e` | FailureScripts/enums_alias |
| 1 | 1 | 0 | `Syntax Error: Trying to call an abstract method` | FailureScripts/inherited2 |
| 1 | 1 | 1 | `Syntax Error: Unbalanced conditional directive` | FailureScripts/conditionals2.1 |
| 1 | 1 | 0 | `Syntax Error: Unexpected "X" for a lambda statement` | LambdaFail/no_arrow |
| 1 | 1 | 0 | `Syntax Error: lazy parameter cannot have a default value` | FailureScripts/params1 |
| 1 | 1 | 0 | `Syntax Error: open array parameter must be const` | FailureScripts/params1 |
| 1 | 1 | 0 | `Syntax Error: var parameter cannot have a default value` | FailureScripts/params1 |
| 1 | 1 | 0 | `TestN` | SimpleScripts/exception_nested_call2 |
| 1 | 1 | 1 | `Unsupported character #N` | FunctionsString/toxml |
| 1 | 1 | 0 | `Warning: "X" has been deprecated: doh` | FailureScripts/func_external |
| 1 | 1 | 0 | `Warning: Assignment to FOR-Loop variable` | FailureScripts/for_var_usage5 |
| 1 | 1 | 1 | `Warning: Filename case does not match: "X" already included as "X"` | FailureScripts/include_once |

## Spurious shapes (produced, not expected)

| lines | fixtures | sole | shape | blocks |
| --- | --- | --- | --- | --- |
| 55 | 55 | 13 | `Syntax Error: Expression expected` | ArrayPass/array_static_enum_index, ArrayPass/dynamic_anonymous_record, AttributesFail/attribute_incorrect1, FailureScripts/block_unfinished4, FailureScripts/cast_string, FailureScripts/class_cast, FailureScripts/const_param3, FailureScripts/const_record1, FailureScripts/const_record3, FailureScripts/const_record4, FailureScripts/constructor_invalid_param, FailureScripts/contracts_error1, FailureScripts/contracts_unfinished1, FailureScripts/contracts_unfinished2, FailureScripts/contracts_unfinished3, FailureScripts/debugbreak, FailureScripts/declaration_mismatch1, FailureScripts/dotdot, FailureScripts/else_unexpected1, FailureScripts/else_unexpected2, FailureScripts/else_unexpected3, FailureScripts/end_implementation1, FailureScripts/enum_scoped2, FailureScripts/enums2, FailureScripts/enums8, FailureScripts/exit_result6, FailureScripts/for_empty, FailureScripts/for_error1, FailureScripts/for_var_error, FailureScripts/for_var_error3, FailureScripts/ifthenelse_expression1, FailureScripts/include_incorrect, FailureScripts/inherited7, FailureScripts/lazy, FailureScripts/params1, FailureScripts/partial_class2, FailureScripts/partial_class3, FailureScripts/proc_semi_warning, FailureScripts/property_description1, FailureScripts/property_reintroduce1, FailureScripts/property_reintroduce2, GenericsFail/binop_constraint, GenericsFail/declaration_params_error1, GenericsFail/declaration_params_error2, HelpersFail/helper_error1, HelpersFail/strict, InterfacesFail/partial_declaration2, InterfacesFail/partial_declaration3, OperatorOverloadPass/c_style, OperatorOverloadPass/operator_overloading2, PropertyExpressionsFail/null_write_expression, SimpleScripts/ifthenelse_optimize2, SimpleScripts/include_expr, SimpleScripts/static_class_array, SimpleScripts/variants_strict |
| 104 | 45 | 7 | `Syntax Error: Unknown name "X"` | ArrayPass/array_index_of_static, ArrayPass/array_of_proc_param, ArrayPass/dynamic_anonymous_record, AttributesFail/attribute_incorrect1, FailureScripts/array_assign_add, FailureScripts/binary_literal1, FailureScripts/case_error5, FailureScripts/cast_string, FailureScripts/class_cast, FailureScripts/class_const2, FailureScripts/class_var_dyn2, FailureScripts/const_param3, FailureScripts/contracts_error2, FailureScripts/debugbreak, FailureScripts/duplicate_uses, FailureScripts/enum_scoped2, FailureScripts/except_error4, FailureScripts/except_error5, FailureScripts/for_error2, FailureScripts/func_ptr2, FailureScripts/invalid_cast2, FailureScripts/invalid_cast3, FailureScripts/nested_method, FailureScripts/new_class4, FailureScripts/nil_type_inference, GenericsFail/binop_constraint, GenericsFail/declaration_params_error2, HelpersPass/dyn_array_create, InnerClassesFail/sub_outside_scope, InterfacesPass/intf_delegate, JSONConnectorFail/autobox, JSONConnectorFail/create_static, JSONConnectorPass/const_array, JSONConnectorPass/stringify_array_of_array, LambdaFail/invalid_expression, Memory/external, Memory/external_constructor_exception2, OperatorOverloadFail/operator_overload1, SimpleScripts/class_var_dyn2, SimpleScripts/const_array4, SimpleScripts/const_array7, SimpleScripts/const_array8, SimpleScripts/consts_expr, SimpleScripts/default_parameters_expr, SimpleScripts/defined |
| 34 | 21 | 3 | `Syntax Error: "X" expected` | ArrayPass/array_add_subclass, ArrayPass/array_static_enum_index, ArrayPass/dynamic_anonymous_record, AssociativeFail/syntax1, FailureScripts/array_index_bracket_missing2, FailureScripts/class_operator5, FailureScripts/constructor_no_name, FailureScripts/exit_result6, FailureScripts/ifthenelse_expression1, FailureScripts/include_incorrect, FailureScripts/legacy_proc_of_object, FailureScripts/method_implem3, FailureScripts/method_implem6, FailureScripts/missing_semi1, FailureScripts/new_class1, FailureScripts/proc_missing_name, FailureScripts/proc_semi_warning, GenericsFail/implem_mismatch1, InterfacesFail/method_decl_error2, SimpleScripts/const_array4, SimpleScripts/ifthenelse_optimize2 |
| 32 | 17 | 9 | `Syntax Error: Incompatible types: Cannot assign "X" to "X"` | FailureScripts/array_assign_error3, FailureScripts/array_const, FailureScripts/assign_error, FailureScripts/assign_op_incompatible, FailureScripts/assign_untyped, FailureScripts/enums9, FailureScripts/func_ptr1, FailureScripts/func_ptr3, FailureScripts/func_ptr4, FailureScripts/multi_dim_dyn_array1, FailureScripts/var_block1, InterfacesFail/assign_intf_from_intf, InterfacesFail/assign_intf_from_obj, InterfacesFail/assign_obj_from_intf, InterfacesFail/interface_inheritence2, OperatorOverloadPass/operator_implicit, SimpleScripts/func_ptr_field_no_param |
| 23 | 12 | 0 | `Syntax Error: expected 'X' after field name or method/property declaration keyword` | AttributesFail/attribute_incorrect2, FailureScripts/array_params1, FailureScripts/array_params2, FailureScripts/class_error6, FailureScripts/class_error7, FailureScripts/class_operator3, FailureScripts/class_operator4, FailureScripts/property_error2, OverloadsFail/overloads_not_implem, PropertyExpressionsFail/expr_write_readonly_property, PropertyExpressionsFail/read_write_other_property, SimpleScripts/ifthenelse_optimize2 |
| 14 | 12 | 1 | `Syntax Error: Undefined variable 'X'` | FailureScripts/as_error, FailureScripts/method2, FailureScripts/missing_class1, FailureScripts/new_array, FailureScripts/proc_missing_name, FailureScripts/proc_semi_warning, FailureScripts/string_get, HelpersPass/dyn_array_create, LambdaFail/invalid_expression, Memory/external, Memory/external_constructor_exception2, SimpleScripts/var_param_rec_method |
| 15 | 11 | 4 | `Hint: Variable "X" declared but not used` | FailureScripts/block_unfinished1, FailureScripts/block_unfinished2, FailureScripts/block_unfinished3, FailureScripts/func_ptr_local, FailureScripts/nested_method, FailureScripts/var_ambiguous_in_scope, FailureScripts/var_block1, LambdaFail/no_capture, SimpleScripts/const_block, SimpleScripts/inference2, SimpleScripts/proc_of_method |
| 12 | 11 | 0 | `Syntax Error: unknown type 'X'` | ArrayPass/array_index_of_static, FailureScripts/array_assign_add, FailureScripts/new_array, GenericsFail/array1-2, JSONConnectorFail/create_static, JSONConnectorPass/const_array, JSONConnectorPass/stringify_array_of_array, Memory/external_constructor_exception2, SimpleScripts/const_array4, SimpleScripts/const_array7, SimpleScripts/const_array8 |
| 11 | 11 | 1 | `Syntax Error: expected 'X' to close class declaration` | FailureScripts/array_params1, FailureScripts/array_params2, FailureScripts/class_error1, FailureScripts/class_error2, FailureScripts/class_error3, FailureScripts/class_error4, FailureScripts/class_error5, FailureScripts/class_error6, FailureScripts/class_error7, FailureScripts/empty_body, FailureScripts/method_params |
| 12 | 9 | 1 | `Hint: Result is never used` | ArrayPass/array_of_array, FailureScripts/result_redefine, LambdaFail/invalid_expression, LambdaFail/no_capture, LambdaFail/no_local_func, LambdaPass/immediate, OverloadsFail/overload_simple, SetOfFail/test_non_variable, SimpleScripts/exception_nested_call2 |
| 11 | 9 | 4 | `Syntax Error: Name "X" already exists` | ArrayPass/dynamic_anonymous_record, FailureScripts/enums4, FailureScripts/member_duplicates, FailureScripts/nested_type2, FailureScripts/partial_class1, InterfacesFail/interface_redefine, InterfacesFail/intf_forwarded_not_implem2, InterfacesFail/method_decl_syntax1, SimpleScripts/func_ptr_property |
| 9 | 9 | 2 | `Syntax Error: duplicate method signature for 'X'` | FailureScripts/class_missing_decl, FailureScripts/empty_body, FailureScripts/member_duplicates, FailureScripts/method_duplicate_override, FailureScripts/method_implem, FailureScripts/nested_method, FailureScripts/partial_class1, FailureScripts/proc_semi_warning, OverloadsFail/meth_overload_simple |
| 11 | 8 | 0 | `Syntax Error: expected 'X' after parameter list` | FailureScripts/const_param3, FailureScripts/declaration_mismatch1, FailureScripts/lazy, FailureScripts/method_params, FailureScripts/params1, InterfacesFail/method_decl_error2, OverloadsFail/overloads_not_implem, SimpleScripts/ifthenelse_optimize2 |
| 8 | 8 | 1 | `Syntax Error: Name expected` | AttributesFail/attribute_incorrect1, AttributesFail/attribute_incorrect2, FailureScripts/final_dot1, FailureScripts/final_dot2, FailureScripts/inherited3, FailureScripts/property_read3, FailureScripts/property_write5, InterfacesPass/interface_nil_cast_from_obj |
| 11 | 7 | 3 | `Syntax Error: method 'X' not declared in class 'X'` | FailureScripts/class_missing_decl, FailureScripts/declaration_mismatch3, FailureScripts/method_implem2, FailureScripts/method_implem4, Memory/external_constructor_exception, Memory/external_constructor_exception2, SimpleScripts/oop |
| 7 | 7 | 2 | `Syntax Error: Method "X" of class "X" not implemented` | FailureScripts/array_params2, FailureScripts/constructor_invalid_param, FailureScripts/constructor_no_name, FailureScripts/default_params3, FailureScripts/default_params3b, FailureScripts/inherited3, FailureScripts/method_implem6 |
| 7 | 7 | 0 | `Syntax Error: parent class 'X' not found` | AttributesFail/attribute_incorrect1, FailureScripts/missing_class1, Memory/external_bidicycle, Memory/external_constructor_exception, Memory/external_constructor_exception2, Memory/external_cycle, Memory/external_selfref |
| 6 | 6 | 4 | `Syntax Error: expected 'X' or 'X', got SEMICOLON` | FailureScripts/array_new, FailureScripts/dyn_array_setlength1, FailureScripts/in_operator8, FailureScripts/missing_parenthesis1, FailureScripts/new_array, SetOfFail/bracket_right_missing |
| 9 | 5 | 0 | `Syntax Error: Expression expected before ASSIGN` | ArrayPass/array_properties2, FailureScripts/for_var_error, FailureScripts/for_var_error4, SimpleScripts/property_of_as, SimpleScripts/static_class_array |
| 7 | 5 | 2 | `Syntax Error: Incompatible types: "X" and "X"` | FailureScripts/cast_base_type, FailureScripts/for_in_subclass, FailureScripts/in_typecheck1, FailureScripts/missing_class1, SimpleScripts/oop_is_as |
| 6 | 5 | 3 | `Hint: "X" does not match case of declaration ("X")` | ArrayPass/array_of_rec_add_create, ArrayPass/dynamic_anonymous_record, FailureScripts/hint_pedantic, FailureScripts/params3, HelpersPass/record_array_helper |
| 6 | 5 | 0 | `Syntax Error: Expression expected before COMMA` | ArrayPass/array_static_enum_index, AssociativeFail/syntax1, FailureScripts/dotdot, FailureScripts/duplicate_uses, FailureScripts/enum_scoped2 |
| 6 | 5 | 0 | `Syntax Error: Expression expected before TO` | FailureScripts/for_error1, FailureScripts/for_var_error, FailureScripts/for_var_error4, FailureScripts/legacy_proc_of_object, SimpleScripts/static_class_array |
| 6 | 5 | 1 | `Syntax Error: No arguments expected` | FailureScripts/enums8, FailureScripts/new_class4, HelpersFail/function_helper, HelpersFail/helper_explicit, HelpersPass/declared_helper |
| 6 | 5 | 0 | `Syntax Error: property 'X' read specifier 'X' not found in class 'X'` | FailureScripts/class_property2, FailureScripts/property_read3, PropertyExpressionsFail/read_write_other_property, PropertyExpressionsPass/read_write_other_property, SimpleScripts/class_var_as_prop |
| 6 | 5 | 0 | `Syntax Error: property 'X' write specifier 'X' not found in class 'X'` | FailureScripts/class_property2, FailureScripts/property_error1, PropertyExpressionsFail/read_write_other_property, PropertyExpressionsPass/read_write_other_property, SimpleScripts/class_var_as_prop |
| 5 | 5 | 0 | `Syntax Error: Expression expected before EQ` | FailureScripts/declaration_mismatch1, FailureScripts/params1, GenericsFail/array1-2, GenericsFail/declaration_params_error1, GenericsFail/declaration_params_error2 |
| 9 | 4 | 0 | `Syntax Error: Argument N expects type "X" instead of "X"` | ArrayPass/array_assign_subclass, FailureScripts/assign_error, FailureScripts/enum_byname, SimpleScripts/func_ptr_field_no_param |
| 9 | 4 | 0 | `Syntax Error: Cannot read a write only property` | ArrayPass/array_properties2, FailureScripts/class_property2, PropertyExpressionsPass/read_write_other_property, SimpleScripts/class_var_as_prop |
| 9 | 4 | 1 | `Syntax Error: Class method or constructor expected` | FailureScripts/class_property3, FailureScripts/const_expr_1, FailureScripts/static_class1, SimpleScripts/class_property |
| 9 | 4 | 1 | `Syntax Error: Object reference needed to read/write an object field` | FailureScripts/class_property2, FailureScripts/property_error7, SimpleScripts/class_const_as_prop, SimpleScripts/class_property |
| 6 | 4 | 1 | `Syntax Error: End of block expected` | FailureScripts/block_unfinished3, FailureScripts/final_dot1, FailureScripts/final_dot2, FailureScripts/for_var_usage5 |
| 6 | 4 | 1 | `Syntax Error: cannot infer type for variable 'X' from initializer` | FailureScripts/array_index_bracket_missing2, FailureScripts/array_range1, FailureScripts/array_range2, FailureScripts/invalid_index |
| 6 | 4 | 0 | `Syntax Error: expected parameter name` | FailureScripts/method_params, InterfacesFail/method_decl_syntax2, OverloadsFail/overloads_not_implem, SimpleScripts/ifthenelse_optimize2 |
| 6 | 4 | 2 | `Warning: Infinite loop` | FailureScripts/block_unfinished3, FailureScripts/break_in_finally, FailureScripts/infinite_loop, FailureScripts/loop_infinite |
| 4 | 4 | 2 | `Syntax Error: address-of operator (@) requires a function or procedure name` | FailureScripts/dyn_array3, FailureScripts/field_init1, FailureScripts/func_ptr6, FailureScripts/func_ptr7 |
| 4 | 4 | 0 | `Syntax Error: constant 'X' must have a value` | FailureScripts/const_array2, FailureScripts/const_record4, FailureScripts/dotdot, FailureScripts/resourcestring3 |
| 4 | 4 | 1 | `Syntax Error: expected identifier after 'X'` | FailureScripts/for_unfinished1, FailureScripts/for_var_error, FailureScripts/for_var_error4, SimpleScripts/static_class_array |
| 4 | 4 | 2 | `Warning: Unit name does not match file name` | FailureScripts/func_external, FailureScripts/interface_implementation, FailureScripts/nested_method, OverloadsFail/forwards_unit |
| 5 | 3 | 2 | `Syntax Error: Member symbol "X" is not visible from this scope` | FailureScripts/class_var_scope1, OverloadsFail/meth_private_public, SimpleScripts/class_scoping1 |
| 5 | 3 | 0 | `Syntax Error: array element N has type String, expected Integer` | FailureScripts/array_assign_error3, FailureScripts/array_const, FailureScripts/array_plus_assign |
| 5 | 3 | 1 | `Syntax Error: invalid assignment target` | FailureScripts/array_index_bracket_missing, FailureScripts/self_not_writable, PropertyExpressionsFail/expr_write_readonly_property |
| 4 | 3 | 1 | `Syntax Error: Bound isn't of an ordinal type` | FailureScripts/array_error7, FailureScripts/array_static_bounds, SimpleScripts/const_array4 |
| 4 | 3 | 0 | `Syntax Error: There is no accessible member with name "X" for type TTest` | FailureScripts/property_reintroduce1, FailureScripts/property_reintroduce2, SimpleScripts/property_reintroduce |
| 4 | 3 | 2 | `Syntax Error: cannot assign to read-only variable 'X'` | FailureScripts/const_2, FailureScripts/const_param1, FailureScripts/const_param4 |
| 4 | 3 | 2 | `Syntax Error: expected 'X' after external` | FailureScripts/class_external, FailureScripts/external2, FailureScripts/external_property |
| 3 | 3 | 0 | _(empty message text)_ | JSONConnectorPass/implicit_from_cast, SimpleScripts/class_operator3, SimpleScripts/inherited1 |
| 3 | 3 | 3 | `Hint: Empty FOR loop` | FailureScripts/for_loopvar1, FailureScripts/for_var_usage3, FailureScripts/for_var_usage4 |
| 3 | 3 | 2 | `Syntax Error: 'X' cannot be used in class 'X' which has no parent class` | FailureScripts/inherited4, FailureScripts/inherited6, FailureScripts/inherited7 |
| 3 | 3 | 1 | `Syntax Error: Argument N expects type "X"` | FailureScripts/enum_byname, FailureScripts/func_ptr_var_param, FailureScripts/lazy_func_ptr |
| 3 | 3 | 0 | `Syntax Error: Expression expected before CLASS` | AttributesFail/attribute_incorrect1, AttributesFail/attribute_incorrect2, FailureScripts/class_nested |
| 3 | 3 | 0 | `Syntax Error: Expression expected before COLON` | FailureScripts/lazy, LambdaFail/invalid_expression, SimpleScripts/ifthenelse_optimize2 |
| 3 | 3 | 0 | `Syntax Error: Expression expected before INDEX` | FailureScripts/for_var_error4, SimpleScripts/ifthenelse_optimize2, SimpleScripts/static_class_array |
| 3 | 3 | 0 | `Syntax Error: Read access of property should be a static method` | FailureScripts/const_expr_1, FailureScripts/static_class1, SimpleScripts/class_property |
| 3 | 3 | 1 | `Syntax Error: expected 'X' after if condition` | FailureScripts/ifthenelse_expression1, FailureScripts/in_operator8, FailureScripts/operator1 |
| 3 | 3 | 1 | `Syntax Error: expected 'X' in operator declaration` | OperatorOverloadFail/operator_overload3, OperatorOverloadFail/operator_overload4, OperatorOverloadFail/operator_overload5 |
| 3 | 3 | 1 | `Syntax Error: expected 'X' to close interface declaration` | InterfacesFail/method_decl_error2, InterfacesFail/method_decl_syntax2, InterfacesFail/partial_declaration |
| 3 | 3 | 0 | `Syntax Error: expected identifier in class inheritance list` | FailureScripts/class_error2, FailureScripts/class_error3, FailureScripts/class_error7 |
| 3 | 3 | 2 | `Syntax Error: expected next token to be RPAREN, got SEMICOLON instead` | FailureScripts/enums3, FailureScripts/nested_type1, PropertyExpressionsFail/missing_reader_bracket |
| 3 | 3 | 0 | `Syntax Error: expected next token to be SEMICOLON, got REINTRODUCE instead` | FailureScripts/property_reintroduce1, FailureScripts/property_reintroduce2, SimpleScripts/property_reintroduce |
| 3 | 3 | 0 | `Syntax Error: expected type after 'X' operator` | FailureScripts/as_error, FailureScripts/as_invalid_right, FailureScripts/implements_error |
| 3 | 3 | 0 | `Syntax Error: expected type expression, got )` | FailureScripts/as_invalid_right, FailureScripts/method_params, InterfacesFail/method_decl_syntax2 |
| 3 | 3 | 1 | `Syntax Error: type 'X' already declared` | FailureScripts/class_duplicate_subtype, FailureScripts/partial_class5, FailureScripts/unit_prefix4 |
| 11 | 2 | 0 | `Syntax Error: function 'X' argument N must be a variable` | FailureScripts/swap1, SimpleScripts/swap2 |
| 9 | 2 | 0 | `Syntax Error: There is no accessible member with name "X" for type TChainItem` | Memory/external_bidicycle, Memory/external_cycle |
| 8 | 2 | 1 | `Syntax Error: implementation signature for 'X' does not match forward declaration` | FailureScripts/declaration_mismatch1, FailureScripts/declaration_mismatch2 |
| 7 | 2 | 1 | `Syntax Error: Array expected` | FailureScripts/array_index_extra, InterfacesPass/interface_properties |
| 6 | 2 | 1 | `Syntax Error: There is no accessible member with name "X" for type String` | FailureScripts/string_builtin_methods1, SimpleScripts/string_builtin_methods |
| 5 | 2 | 0 | `Syntax Error: There is no accessible member with name "X" for type function` | SetOfFail/test_non_variable, SimpleScripts/default_parameters |
| 5 | 2 | 0 | `Syntax Error: expected unit name after 'X'` | FailureScripts/duplicate_uses, OperatorOverloadFail/operator_overload5 |
| 4 | 2 | 2 | `Syntax Error: circular inheritance detected in class 'X'` | FailureScripts/class_circular, FailureScripts/class_loop |
| 4 | 2 | 0 | `Syntax Error: expected 'X' after constant value` | FailureScripts/class_nested, OverloadsFail/overloads_not_implem |
| 4 | 2 | 0 | `Syntax Error: incompatible types in coalesce operator: array of TObject and array of TTest(TObject)` | FailureScripts/coalesce_dynarray, SimpleScripts/coalesce_dynarray |
| 4 | 2 | 1 | `Syntax Error: operator 'X' already defined for operand types (TObject, TObject)` | OperatorOverloadFail/operator_overload5, OperatorOverloadFail/operator_overload6 |
| 4 | 2 | 0 | `Syntax Error: unary - requires numeric operand, got String` | FailureScripts/neg_type, FailureScripts/plus_non_numeric |
| 3 | 2 | 1 | `Hint: Overloaded method "X" should be marked with the "X" directive` | SimpleScripts/inherited_constructor, SimpleScripts/new_class2 |
| 3 | 2 | 0 | `Syntax Error: Cannot set a value for a read-only property` | PropertyExpressionsPass/read_write_other_property, SimpleScripts/class_var_as_prop |
| 3 | 2 | 0 | `Syntax Error: Expression expected before PROPERTY` | FailureScripts/external_property, SimpleScripts/ifthenelse_optimize2 |
| 3 | 2 | 1 | `Syntax Error: Incompatible parameter types - "X" expected (instead of "X")` | ArrayPass/array_map, AssociativeFail/delete |
| 3 | 2 | 1 | `Syntax Error: There is no overloaded version of "X" that can be called with these arguments` | FailureScripts/strict_parameter_type, OverloadsFail/overload_simple |
| 3 | 2 | 0 | `Syntax Error: Unexpected "X"` | PropertyExpressionsPass/read_write_other_property, SimpleScripts/class_var_as_prop |
| 3 | 2 | 0 | `Syntax Error: cannot infer type for class var 'X'` | FailureScripts/class_var_dyn2, SimpleScripts/class_var_dyn2 |
| 3 | 2 | 2 | `Syntax Error: expected 'X' after for-in collection` | FailureScripts/for_in_str1, SetOfFail/for_in_set_missing_do |
| 3 | 2 | 1 | `Syntax Error: function 'X' first argument must be a variable (identifier, array element, or field)` | FailureScripts/lazy, FailureScripts/passing_const_var |
| 3 | 2 | 0 | `Syntax Error: incompatible types in coalesce operator: array of TObject and array of TSub(TTest)` | FailureScripts/coalesce_dynarray, SimpleScripts/coalesce_dynarray |
| 3 | 2 | 0 | `Syntax Error: incompatible types in coalesce operator: array of TTest(TObject) and array of TSub(TTest)` | FailureScripts/coalesce_dynarray, SimpleScripts/coalesce_dynarray |
| 3 | 2 | 1 | `Syntax Error: range start must be an ordinal type, got Float` | FailureScripts/in_operator_order, FailureScripts/in_typecheck1 |
| 3 | 2 | 1 | `Syntax Error: unknown type 'X' for field 'X'` | FailureScripts/class_field_type, FailureScripts/property_read3 |
| 3 | 2 | 0 | `Syntax Error: unknown type 'X' in type alias` | FailureScripts/class_of, FailureScripts/legacy_proc_of_object |
| 2 | 2 | 0 | `Hint: Constant Instruction - has no effect` | FailureScripts/func_toomanyargs, FailureScripts/internal_unsupported |
| 2 | 2 | 0 | `Hint: Empty THEN block` | FailureScripts/as_invalid_right, FailureScripts/in_operator6 |
| 2 | 2 | 0 | `Runtime Error: Object not instantiated` | SimpleScripts/property_sub_default, SimpleScripts/virtual_constructor2 |
| 2 | 2 | 0 | `Syntax Error: "X" expected in generic type parameter list` | GenericsFail/declaration_params_error1, GenericsFail/declaration_params_error2 |
| 2 | 2 | 2 | `Syntax Error: 'X' requires a class or class reference` | FailureScripts/new_class2, FailureScripts/new_class5 |
| 2 | 2 | 0 | `Syntax Error: Array index expected "X" but got "X"` | FailureScripts/string_get, InterfacesPass/interface_properties |
| 2 | 2 | 0 | `Syntax Error: Cannot assign to constant 'X'` | FailureScripts/const_1, FailureScripts/const_2 |
| 2 | 2 | 0 | `Syntax Error: Expression expected before ARRAY` | GenericsFail/declaration_params_error1, GenericsFail/declaration_params_error2 |
| 2 | 2 | 0 | `Syntax Error: Expression expected before DOTDOT` | ArrayPass/array_static_enum_index, LambdaFail/invalid_expression |
| 2 | 2 | 0 | `Syntax Error: Expression expected before EQ_EQ` | OperatorOverloadPass/c_style, SimpleScripts/variants_strict |
| 2 | 2 | 0 | `Syntax Error: More arguments expected` | FailureScripts/array_index_bracket_missing1, FailureScripts/enum_byname |
| 2 | 2 | 2 | `Syntax Error: There is no accessible member with name "X" for type set of TMyEnum` | SetOfFail/bracket_left_missing, SetOfFail/invalid_method |
| 2 | 2 | 1 | `Syntax Error: ambiguous overload for 'X'` | FailureScripts/method_implem5, OverloadsFail/meth_overload_simple |
| 2 | 2 | 0 | `Syntax Error: array constructor range bounds must be ordinal` | FailureScripts/array_range1, FailureScripts/in_typecheck1 |
| 2 | 2 | 1 | `Syntax Error: binding 'X' for class operator 'X' expects N parameters, got N` | FailureScripts/class_operator1, FailureScripts/in_operator3 |
| 2 | 2 | 0 | `Syntax Error: binding 'X' for operator 'X' not found` | OperatorOverloadFail/operator_overload5, OperatorOverloadPass/operator_implicit |
| 2 | 2 | 1 | `Syntax Error: cannot access private constant 'X' of class 'X'` | FailureScripts/class_const1, FailureScripts/class_const2 |
| 2 | 2 | 0 | `Syntax Error: cannot infer type for field 'X'` | FailureScripts/field_init1, FailureScripts/nil_type_inference |
| 2 | 2 | 0 | `Syntax Error: cannot infer type for variable 'X' from nil initializer` | FailureScripts/default_func1, FailureScripts/nil_type_inference |
| 2 | 2 | 1 | `Syntax Error: cannot take address of variadic built-in function 'X'` | FailureScripts/swap1, JSONConnectorFail/add_error |
| 2 | 2 | 0 | `Syntax Error: complex return types not yet supported in function pointers` | ArrayPass/array_of_proc_param, JSONConnectorFail/autobox |
| 2 | 2 | 1 | `Syntax Error: could not parse "X" as integer` | FailureScripts/binary_literal1, FailureScripts/binary_literal2 |
| 2 | 2 | 0 | `Syntax Error: expected 'X' after 'X'` | FailureScripts/partial_class2, FailureScripts/partial_class3 |
| 2 | 2 | 0 | `Syntax Error: expected 'X' after 'X' in unit declaration` | AttributesFail/attribute_incorrect2, FailureScripts/end_implementation3 |
| 2 | 2 | 2 | `Syntax Error: expected 'X' after 'X' operand, got SEMICOLON` | FailureScripts/new_class6, FailureScripts/new_class7 |
| 2 | 2 | 0 | `Syntax Error: expected 'X' after for loop variable` | FailureScripts/for_error1, HelpersFail/strict |
| 2 | 2 | 0 | `Syntax Error: expected 'X' or 'X' after const name` | FailureScripts/const_array2, FailureScripts/resourcestring3 |
| 2 | 2 | 0 | `Syntax Error: expected 'X' or 'X' in argument list, got EQ_EQ` | OperatorOverloadPass/c_style, SimpleScripts/variants_strict |
| 2 | 2 | 1 | `Syntax Error: expected 'X' or 'X' in for loop` | FailureScripts/for_error2, FailureScripts/for_var_error2 |
| 2 | 2 | 1 | `Syntax Error: expected 'X', 'X', 'X', 'X', 'X', 'X', 'X', 'X', 'X', 'X', or 'X' after 'X' in type declaration` | FailureScripts/class_error8, HelpersFail/strict |
| 2 | 2 | 1 | `Syntax Error: expected 'X', got SEMICOLON` | FailureScripts/missing_parenthesis2, PropertyExpressionsFail/missing_reader_bracket |
| 2 | 2 | 2 | `Syntax Error: expected identifier in const declaration` | FailureScripts/const_4, FailureScripts/resourcestring2 |
| 2 | 2 | 0 | `Syntax Error: expected identifier in record field declaration` | FailureScripts/record_syntax1, FailureScripts/record_syntax2 |
| 2 | 2 | 0 | `Syntax Error: expected identifier or expression after 'X'` | FailureScripts/property_error3, FailureScripts/property_error4 |
| 2 | 2 | 0 | `Syntax Error: expected next token to be DOT, got SEMICOLON instead` | AttributesFail/attribute_incorrect2, FailureScripts/end_implementation3 |
| 2 | 2 | 0 | `Syntax Error: expected next token to be IDENT, got READONLY instead` | PropertyExpressionsFail/expr_write_readonly_property, PropertyExpressionsFail/read_write_other_property |
| 2 | 2 | 0 | `Syntax Error: expected operator symbol after 'X'` | FailureScripts/class_operator3, OperatorOverloadFail/operator_overload1 |
| 2 | 2 | 1 | `Syntax Error: function 'X' expects N arguments, got N` | FailureScripts/swap1, SetOfFail/include |
| 2 | 2 | 0 | `Syntax Error: inferred lambda return type Void incompatible with expected return type Integer` | LambdaFail/invalid_expression, LambdaPass/simple_func |
| 2 | 2 | 0 | `Syntax Error: optional parameters cannot have lazy, var, or const modifiers` | FailureScripts/declaration_mismatch1, FailureScripts/params1 |
| 2 | 2 | 0 | `Syntax Error: partial class 'X' has conflicting parent classes` | FailureScripts/partial_class1, FailureScripts/partial_class3 |
| 2 | 2 | 0 | `Syntax Error: unary + requires numeric operand, got String` | FailureScripts/plus_non_numeric, SimpleScripts/include_expr |
| 6 | 1 | 0 | `Error: Trying to create an instance of an abstract class` | SimpleScripts/constructor_overload |
| 6 | 1 | 0 | `Syntax Error: incompatible types in coalesce operator: TSubN(TTest) and TSubN(TTest)` | SimpleScripts/coalesce_class_2 |
| 5 | 1 | 0 | `Syntax Error: 'X' cannot be used in class methods (static methods)` | SimpleScripts/class_self |
| 4 | 1 | 0 | `Hint: Previous declaration of class was "X"` | FailureScripts/partial_class4 |
| 3 | 1 | 0 | `Syntax Error: Class "X" already defined` | FailureScripts/classname_already_exists |
| 3 | 1 | 1 | `Syntax Error: Expression expected before EXPORT` | FailureScripts/export |
| 2 | 1 | 0 | `Hint: Private method "X" declared but never used` | InterfacesPass/intf_private |
| 2 | 1 | 0 | `Syntax Error: Array bounds are of different types` | FailureScripts/array_static_bounds |
| 2 | 1 | 0 | `Syntax Error: Expression expected before DEC` | SimpleScripts/plus_plus_minus_minus |
| 2 | 1 | 0 | `Syntax Error: Expression expected before INC` | SimpleScripts/plus_plus_minus_minus |
| 2 | 1 | 1 | `Syntax Error: Incompatible operands` | FailureScripts/in_operator5 |
| 2 | 1 | 0 | `Syntax Error: There is no accessible member with name "X" for type IMy` | InterfacesPass/intf_self_ref |
| 2 | 1 | 0 | `Syntax Error: There is no accessible member with name "X" for type TMyObj` | Memory/external_selfref |
| 2 | 1 | 0 | `Syntax Error: There is no accessible member with name "X" for type Variant` | ArrayPass/array_add_subclass |
| 2 | 1 | 0 | `Syntax Error: array element N has type Integer, expected TEnum` | ArrayPass/dynamic_array_assign |
| 2 | 1 | 0 | `Syntax Error: binding 'X' for operator 'X' expects N parameters, got N` | OperatorOverloadFail/operator_overload1 |
| 2 | 1 | 0 | `Syntax Error: break statement not allowed in finally block` | FailureScripts/break_in_finally |
| 2 | 1 | 0 | `Syntax Error: cannot compare TObject with Integer` | FailureScripts/object_relops |
| 2 | 1 | 0 | `Syntax Error: continue statement not allowed in finally block` | FailureScripts/break_in_finally |
| 2 | 1 | 0 | `Syntax Error: expected 'X' after field declaration` | FailureScripts/array_params2 |
| 2 | 1 | 0 | `Syntax Error: expected 'X' at end of operator declaration` | HelpersPass/helper_as_overload |
| 2 | 1 | 1 | `Syntax Error: expected 'X', 'X', or 'X' after 'X'` | FailureScripts/class_class |
| 2 | 1 | 0 | `Syntax Error: expected 'X', got WRITE` | PropertyExpressionsFail/missing_reader_bracket |
| 2 | 1 | 0 | `Syntax Error: expected next token to be OBJECT, got SEMICOLON instead` | FailureScripts/legacy_proc_of_object |
| 2 | 1 | 0 | `Syntax Error: expected next token to be SEMICOLON, got DESCRIPTION instead` | FailureScripts/property_description1 |
| 2 | 1 | 0 | `Syntax Error: expected parameter name in indexed property` | FailureScripts/array_params2 |
| 2 | 1 | 0 | `Syntax Error: function 'X' element argument has type procedure(), expected TMyEnum` | SetOfFail/invalid_operand |
| 2 | 1 | 0 | `Syntax Error: function 'X' expects array, enum, or type name, got class of TObject` | FailureScripts/internal_unsupported |
| 2 | 1 | 0 | `Syntax Error: function 'X' first argument must be Boolean, got String` | FailureScripts/assert |
| 2 | 1 | 0 | `Syntax Error: function 'X' second argument must be String, got Integer` | FailureScripts/assert |
| 2 | 1 | 0 | `Syntax Error: implementation return type for 'X' does not match forward declaration` | FailureScripts/declaration_mismatch1 |
| 2 | 1 | 1 | `Syntax Error: operator += not supported for type JSONVariant` | JSONConnectorFail/plus_assign_field |
| 2 | 1 | 0 | `Syntax Error: operator += not supported for type array of Integer` | FailureScripts/array_plus_assign |
| 2 | 1 | 1 | `Syntax Error: promoted property 'X' not found in an ancestor of class 'X'` | FailureScripts/property_promotion |
| 2 | 1 | 0 | `Syntax Error: property 'X' setter method 'X' has N parameters, expected N parameter` | FailureScripts/property_error3 |
| 2 | 1 | 0 | `Syntax Error: set element must be an ordinal value, got Float` | FailureScripts/in_typecheck1 |
| 2 | 1 | 0 | `Syntax Error: type mismatch in set literal: expected Integer, got String` | FailureScripts/in_typecheck1 |
| 2 | 1 | 0 | `Syntax Error: type mismatch in set literal: expected set of String, got set of Integer` | FailureScripts/in_typecheck1 |
| 2 | 1 | 0 | `Syntax Error: unary - requires numeric operand, got TObject` | FailureScripts/neg_type |
| 2 | 1 | 1 | `Syntax Error: unknown parameter type 'X' in interface method 'X'` | InterfacesFail/method_decl_error1 |
| 2 | 1 | 0 | `Warning: "X" has been deprecated` | FailureScripts/class_deprecated |
| 2 | 1 | 0 | `Warning: "X" has been deprecated: old stuff` | FailureScripts/class_deprecated |
| 1 | 1 | 1 | `Compile Error: aborted` | FailureScripts/static_methods |
| 1 | 1 | 0 | `Runtime Error: "X" is not a valid floating point value` | SimpleScripts/variants_casts |
| 1 | 1 | 0 | `Runtime Error: Low() failed: Low() expects array, enum, string, or type name, got INTEGER` | SimpleScripts/for_step_overflow |
| 1 | 1 | 0 | `Runtime Error: Ord() expects enum, boolean, integer, or string, got VARIANT` | SimpleScripts/ord |
| 1 | 1 | 0 | `Runtime Error: constructor failed: ERROR: error evaluating default value: ERROR: cannot infer type for empty array literal` | SimpleScripts/default_parameters_empty_array |
| 1 | 1 | 0 | `Runtime Error: error evaluating argument N: ERROR: cannot assign nil to class of TObject` | SimpleScripts/nil_meta_parameter |
| 1 | 1 | 0 | `Runtime Error: error evaluating argument N: ERROR: cannot determine type for array element N` | ArrayPass/array_of_metaclass |
| 1 | 1 | 0 | `Runtime Error: member 'X' not found` | SimpleScripts/virtual_constructor |
| 1 | 1 | 0 | `Runtime Error: method 'X' not found for type 'X' in Test` | JSONConnectorPass/implicit_from_cast |
| 1 | 1 | 0 | `Runtime Error: method, property, or field 'X' not found in parent class 'X' in TChild.GetBase` | SimpleScripts/inherited1 |
| 1 | 1 | 0 | `Runtime Error: type error in float operation: expected FLOAT or INTEGER, got STRING` | SimpleScripts/variants_is_bool |
| 1 | 1 | 0 | `Runtime Error: type mismatch: cannot add INTEGER to String in TTest.AppendStrings` | SimpleScripts/class_operator3 |
| 1 | 1 | 0 | `Runtime Error: undefined variable 'X'` | ArrayPass/array_filter_record |
| 1 | 1 | 0 | `Syntax Error: "X" expected in generic type argument list` | GenericsFail/array1-2 |
| 1 | 1 | 1 | `Syntax Error: 'X' can only be used inside a class method` | FailureScripts/inherited1 |
| 1 | 1 | 1 | `Syntax Error: 'X' is not a function or procedure (got Integer)` | FailureScripts/at_integer |
| 1 | 1 | 0 | `Syntax Error: 'X' operator requires a class reference for a metaclass cast, got TObject` | FailureScripts/object_relops |
| 1 | 1 | 0 | `Syntax Error: 'X' operator requires class instance or class reference, got String` | FailureScripts/implements_error |
| 1 | 1 | 0 | `Syntax Error: 'X' operator requires class instance or interface, got Integer` | FailureScripts/object_relops |
| 1 | 1 | 0 | `Syntax Error: 'X' operator requires class instance or interface, got TClass` | FailureScripts/object_relops |
| 1 | 1 | 0 | `Syntax Error: 'X' operator requires class instance, got Integer` | FailureScripts/object_relops |
| 1 | 1 | 0 | `Syntax Error: Bare raise statement is only valid inside an exception handler` | FailureScripts/raise_error |
| 1 | 1 | 0 | `Syntax Error: Cannot assign () -> Void to procedure () variable 'X'` | InterfacesPass/intf_delegate |
| 1 | 1 | 0 | `Syntax Error: Cannot assign Integer to String` | FailureScripts/const_1 |
| 1 | 1 | 0 | `Syntax Error: Colon "X" expected` | JSONConnectorFail/autobox |
| 1 | 1 | 0 | `Syntax Error: Constant "X" cannot be written to` | PropertyExpressionsFail/expr_write_readonly_property |
| 1 | 1 | 1 | `Syntax Error: Constant expression expected` | FailureScripts/special_funcs4 |
| 1 | 1 | 1 | `Syntax Error: Expression expected before FINALIZATION` | FailureScripts/double_finalization |
| 1 | 1 | 0 | `Syntax Error: Expression expected before FUNCTION` | LambdaFail/invalid_expression |
| 1 | 1 | 0 | `Syntax Error: Expression expected before GREATER` | GenericsFail/array1-2 |
| 1 | 1 | 1 | `Syntax Error: Expression expected before INITIALIZATION` | FailureScripts/double_initialization |
| 1 | 1 | 0 | `Syntax Error: Expression expected before LESS_LESS` | OperatorOverloadPass/operator_overloading2 |
| 1 | 1 | 1 | `Syntax Error: Expression expected before OVERLOAD` | OverloadsFail/func_ptr_overload |
| 1 | 1 | 0 | `Syntax Error: Expression expected before READ` | ArrayPass/dynamic_anonymous_record |
| 1 | 1 | 0 | `Syntax Error: Expression expected before READONLY` | PropertyExpressionsFail/expr_write_readonly_property |
| 1 | 1 | 0 | `Syntax Error: Expression expected before STRICT` | HelpersFail/strict |
| 1 | 1 | 0 | `Syntax Error: Invalid Operands` | GenericsFail/binop_constraint |
| 1 | 1 | 1 | `Syntax Error: Not a method` | FailureScripts/contracts_error3 |
| 1 | 1 | 0 | `Syntax Error: Record has no field members` | FailureScripts/record_syntax1 |
| 1 | 1 | 1 | `Syntax Error: Record type "X" is not fully defined` | FailureScripts/record_recursive2 |
| 1 | 1 | 1 | `Syntax Error: String expected` | FailureScripts/resourcestring1 |
| 1 | 1 | 0 | `Syntax Error: There is already a method with name "X"` | FailureScripts/func_external |
| 1 | 1 | 1 | `Syntax Error: There is no accessible member with name "X" for type TMyEnum` | FailureScripts/enums |
| 1 | 1 | 1 | `Syntax Error: There is no accessible member with name "X" for type array of Variant` | FailureScripts/member_of_void1 |
| 1 | 1 | 0 | `Syntax Error: There is no accessible member with name "X" for type class of TBase` | SimpleScripts/class_var_dyn2 |
| 1 | 1 | 1 | `Syntax Error: There is no accessible member with name "X" for type e` | FailureScripts/enums_alias |
| 1 | 1 | 1 | `Syntax Error: Unbalanced conditional directive` | FailureScripts/conditionals2.1 |
| 1 | 1 | 0 | `Syntax Error: array constructor range bounds must have the same type: got Integer and TNum` | FailureScripts/array_range2 |
| 1 | 1 | 0 | `Syntax Error: array constructor range bounds must have the same type: got TAlpha and Integer` | FailureScripts/array_range2 |
| 1 | 1 | 0 | `Syntax Error: array constructor range bounds must have the same type: got TAlpha and TNum` | FailureScripts/array_range2 |
| 1 | 1 | 0 | `Syntax Error: array dimension N must be integer, got Boolean` | FailureScripts/multi_dim_dyn_array1 |
| 1 | 1 | 0 | `Syntax Error: array element N has type Float, expected Integer` | FailureScripts/array_assign_error3 |
| 1 | 1 | 1 | `Syntax Error: array element N has type procedure(array of Float), expected Float` | FailureScripts/array_of_proc2 |
| 1 | 1 | 1 | `Syntax Error: array element N has type procedure(array of Integer), expected Integer` | FailureScripts/const_procedure_array |
| 1 | 1 | 1 | `Syntax Error: array element N has type procedure(array of procedure()), expected procedure()` | FailureScripts/array_of_proc |
| 1 | 1 | 0 | `Syntax Error: binding 'X' parameter N type Integer does not match operator operand type Float` | FailureScripts/in_operator3 |
| 1 | 1 | 0 | `Syntax Error: binding 'X' parameter N type Integer does not match operator operand type TObject` | OperatorOverloadFail/operator_overload1 |
| 1 | 1 | 0 | `Syntax Error: binding 'X' parameter N type TObject does not match operator operand type Integer` | OperatorOverloadFail/operator_overload1 |
| 1 | 1 | 0 | `Syntax Error: break statement not allowed outside loop` | FailureScripts/break_continue |
| 1 | 1 | 1 | `Syntax Error: cannot assign to constant 'X'` | FailureScripts/class_const3 |
| 1 | 1 | 0 | `Syntax Error: cannot compare TFooClass with TFoo(TObject)` | FailureScripts/class_operator5 |
| 1 | 1 | 0 | `Syntax Error: cannot infer type for field 'X' in record 'X'` | FailureScripts/nil_type_inference |
| 1 | 1 | 0 | `Syntax Error: cannot open include file 'X': failed to read file: open testdata/fixtures/FailureScripts/include.INC: no such file or directory` | FailureScripts/include_once |
| 1 | 1 | 0 | `Syntax Error: cannot open include file 'X': failed to read file: open testdata/fixtures/FailureScripts/include.inc: no such file or directory` | FailureScripts/include_once |
| 1 | 1 | 0 | `Syntax Error: case range end type Boolean incompatible with case expression type Float` | FailureScripts/case_range_typecheck |
| 1 | 1 | 0 | `Syntax Error: case range end type Boolean incompatible with case expression type Integer` | FailureScripts/case_range_typecheck |
| 1 | 1 | 0 | `Syntax Error: case range end type Boolean incompatible with case expression type String` | FailureScripts/case_range_typecheck |
| 1 | 1 | 0 | `Syntax Error: case range end type Float incompatible with case expression type Boolean` | FailureScripts/case_range_typecheck |
| 1 | 1 | 0 | `Syntax Error: case range end type Float incompatible with case expression type Integer` | FailureScripts/case_range_typecheck |
| 1 | 1 | 0 | `Syntax Error: case range end type Float incompatible with case expression type String` | FailureScripts/case_range_typecheck |
| 1 | 1 | 0 | `Syntax Error: case range end type Integer incompatible with case expression type Boolean` | FailureScripts/case_range_typecheck |
| 1 | 1 | 0 | `Syntax Error: case range end type Integer incompatible with case expression type String` | FailureScripts/case_range_typecheck |
| 1 | 1 | 0 | `Syntax Error: case range end type String incompatible with case expression type Boolean` | FailureScripts/case_range_typecheck |
| 1 | 1 | 0 | `Syntax Error: case range end type String incompatible with case expression type Float` | FailureScripts/case_range_typecheck |
| 1 | 1 | 0 | `Syntax Error: case range end type String incompatible with case expression type Integer` | FailureScripts/case_range_typecheck |
| 1 | 1 | 0 | `Syntax Error: case range start type Boolean and end type Float are incompatible` | FailureScripts/case_range_mismatch |
| 1 | 1 | 0 | `Syntax Error: case range start type Boolean incompatible with case expression type Float` | FailureScripts/case_range_typecheck |
| 1 | 1 | 0 | `Syntax Error: case range start type Boolean incompatible with case expression type Integer` | FailureScripts/case_range_typecheck |
| 1 | 1 | 0 | `Syntax Error: case range start type Boolean incompatible with case expression type String` | FailureScripts/case_range_typecheck |
| 1 | 1 | 0 | `Syntax Error: case range start type Float incompatible with case expression type Boolean` | FailureScripts/case_range_typecheck |
| 1 | 1 | 0 | `Syntax Error: case range start type Float incompatible with case expression type Integer` | FailureScripts/case_range_typecheck |
| 1 | 1 | 0 | `Syntax Error: case range start type Float incompatible with case expression type String` | FailureScripts/case_range_typecheck |
| 1 | 1 | 0 | `Syntax Error: case range start type Integer and end type String are incompatible` | FailureScripts/case_range_mismatch |
| 1 | 1 | 0 | `Syntax Error: case range start type Integer incompatible with case expression type Boolean` | FailureScripts/case_range_typecheck |
| 1 | 1 | 0 | `Syntax Error: case range start type Integer incompatible with case expression type String` | FailureScripts/case_range_typecheck |
| 1 | 1 | 0 | `Syntax Error: case range start type String and end type class of TObject are incompatible` | FailureScripts/case_range_mismatch |
| 1 | 1 | 0 | `Syntax Error: case range start type String incompatible with case expression type Boolean` | FailureScripts/case_range_typecheck |
| 1 | 1 | 0 | `Syntax Error: case range start type String incompatible with case expression type Float` | FailureScripts/case_range_typecheck |
| 1 | 1 | 0 | `Syntax Error: case range start type String incompatible with case expression type Integer` | FailureScripts/case_range_typecheck |
| 1 | 1 | 0 | `Syntax Error: case range start type Void incompatible with case expression type Integer` | FailureScripts/case_error5 |
| 1 | 1 | 1 | `Syntax Error: class 'X' does not implement interface method 'X' from interface 'X'` | InterfacesFail/implement_interface1 |
| 1 | 1 | 1 | `Syntax Error: class operator 'X' already defined for class 'X'` | FailureScripts/class_operator2 |
| 1 | 1 | 0 | `Syntax Error: continue statement not allowed outside loop` | FailureScripts/break_continue |
| 1 | 1 | 0 | `Syntax Error: duplicate property 'X' in class 'X'` | FailureScripts/member_duplicates |
| 1 | 1 | 0 | `Syntax Error: exit statement not allowed in finally block` | FailureScripts/break_in_finally |
| 1 | 1 | 0 | `Syntax Error: exit value type Void incompatible with function return type Integer` | FailureScripts/exit_result6 |
| 1 | 1 | 1 | `Syntax Error: exit with value not allowed in procedure` | FailureScripts/exit_result2 |
| 1 | 1 | 0 | `Syntax Error: expected 'X' after 'X' in metaclass type` | FailureScripts/class_type |
| 1 | 1 | 1 | `Syntax Error: expected 'X' after 'X' in set declaration` | SetOfFail/of_missing |
| 1 | 1 | 1 | `Syntax Error: expected 'X' after case expression` | FailureScripts/case_error2 |
| 1 | 1 | 1 | `Syntax Error: expected 'X' after case value` | FailureScripts/case_error4 |
| 1 | 1 | 0 | `Syntax Error: expected 'X' after empty` | FailureScripts/empty_body |
| 1 | 1 | 0 | `Syntax Error: expected 'X' after field name` | FailureScripts/record_syntax2 |
| 1 | 1 | 0 | `Syntax Error: expected 'X' after interface method declaration` | InterfacesFail/method_decl_error2 |
| 1 | 1 | 0 | `Syntax Error: expected 'X' after parent interface` | InterfacesFail/partial_declaration3 |
| 1 | 1 | 0 | `Syntax Error: expected 'X' or 'X' after indexed property parameter` | FailureScripts/array_params2 |
| 1 | 1 | 0 | `Syntax Error: expected 'X' or 'X' in argument list, got RBRACK` | FailureScripts/const_record4 |
| 1 | 1 | 0 | `Syntax Error: expected 'X' or 'X' in class inheritance list` | FailureScripts/class_error4 |
| 1 | 1 | 0 | `Syntax Error: expected 'X' or 'X' in operator operand list` | OperatorOverloadFail/operator_overload3 |
| 1 | 1 | 0 | `Syntax Error: expected 'X' or 'X', got IDENT` | AttributesFail/attribute_incorrect2 |
| 1 | 1 | 0 | `Syntax Error: expected 'X' or 'X', got INT` | FailureScripts/constructor_invalid_param |
| 1 | 1 | 0 | `Syntax Error: expected 'X' or 'X', got RPAREN` | FailureScripts/array_index_bracket_missing1 |
| 1 | 1 | 0 | `Syntax Error: expected 'X' or 'X', got THEN` | FailureScripts/in_operator6 |
| 1 | 1 | 0 | `Syntax Error: expected 'X' to close helper declaration` | HelpersFail/helper_scopes1 |
| 1 | 1 | 0 | `Syntax Error: expected 'X' to close record declaration` | FailureScripts/record_syntax2 |
| 1 | 1 | 1 | `Syntax Error: expected 'X' to close try statement` | FailureScripts/try_except1 |
| 1 | 1 | 0 | `Syntax Error: expected 'X' to start enum declaration` | FailureScripts/enum_scoped2 |
| 1 | 1 | 0 | `Syntax Error: expected 'X', 'X', 'X' or 'X' after 'X' keyword in helper` | HelpersFail/helper_error5 |
| 1 | 1 | 0 | `Syntax Error: expected 'X', 'X', 'X', 'X' or 'X' after 'X' keyword in record` | FailureScripts/record_syntax1 |
| 1 | 1 | 0 | `Syntax Error: expected 'X', 'X', 'X', 'X', 'X', or 'X' after 'X' keyword` | FailureScripts/class_error1 |
| 1 | 1 | 0 | `Syntax Error: expected class type after 'X'` | FailureScripts/class_of |
| 1 | 1 | 1 | `Syntax Error: expected condition after 'X'` | FailureScripts/in_operator4 |
| 1 | 1 | 0 | `Syntax Error: expected enum value name, got INT` | FailureScripts/enums2 |
| 1 | 1 | 0 | `Syntax Error: expected identifier after 'X' in operator declaration` | OperatorOverloadFail/operator_overload5 |
| 1 | 1 | 0 | `Syntax Error: expected identifier for method name` | InterfacesFail/method_decl_syntax1 |
| 1 | 1 | 0 | `Syntax Error: expected identifier for parent interface` | InterfacesFail/partial_declaration2 |
| 1 | 1 | 1 | `Syntax Error: expected next token to be COLON, got END instead` | FailureScripts/property_error11 |
| 1 | 1 | 0 | `Syntax Error: expected next token to be COLON, got IDENT instead` | FailureScripts/property_error2 |
| 1 | 1 | 0 | `Syntax Error: expected next token to be COLON, got RBRACK instead` | FailureScripts/array_params1 |
| 1 | 1 | 0 | `Syntax Error: expected next token to be FOR` | HelpersFail/helper_error1 |
| 1 | 1 | 0 | `Syntax Error: expected next token to be IDENT, got SEMICOLON instead` | FailureScripts/property_error1 |
| 1 | 1 | 1 | `Syntax Error: expected next token to be SEMICOLON, got DEFAULT instead` | FailureScripts/property_default1 |
| 1 | 1 | 1 | `Syntax Error: expected next token to be SEMICOLON, got INTERFACE instead` | FailureScripts/end_implementation2 |
| 1 | 1 | 0 | `Syntax Error: expected operand type in class operator declaration` | FailureScripts/class_operator4 |
| 1 | 1 | 1 | `Syntax Error: expected program name after 'X' keyword` | FailureScripts/program |
| 1 | 1 | 0 | `Syntax Error: expected record field name, got PROPERTY` | ArrayPass/dynamic_anonymous_record |
| 1 | 1 | 0 | `Syntax Error: expected type expression after 'X'` | FailureScripts/array_static_bounds |
| 1 | 1 | 0 | `Syntax Error: expected type expression, got` | InterfacesFail/method_decl_error2 |
| 1 | 1 | 0 | `Syntax Error: expected type expression, got ;` | FailureScripts/class_of |
| 1 | 1 | 0 | `Syntax Error: expected type expression, got >` | GenericsFail/array1-2 |
| 1 | 1 | 0 | `Syntax Error: expected type expression, got end` | HelpersFail/helper_error2 |
| 1 | 1 | 0 | `Syntax Error: expected type expression, got hello` | FailureScripts/as_error |
| 1 | 1 | 1 | `Syntax Error: expected type expression, got interface` | FailureScripts/interface_type |
| 1 | 1 | 1 | `Syntax Error: expected type expression, got nil` | GenericsFail/record_constraint1 |
| 1 | 1 | 0 | `Syntax Error: expected type expression, got world` | FailureScripts/implements_error |
| 1 | 1 | 0 | `Syntax Error: expected type name after 'X'` | FailureScripts/new_class1 |
| 1 | 1 | 0 | `Syntax Error: expected type name after 'X' in helper declaration` | HelpersFail/helper_error2 |
| 1 | 1 | 0 | `Syntax Error: external class 'X' cannot inherit from non-external class 'X'` | FailureScripts/partial_class4 |
| 1 | 1 | 1 | `Syntax Error: file name expected after $i` | FailureScripts/include_expr |
| 1 | 1 | 1 | `Syntax Error: for loop step must be Integer, got String` | FailureScripts/for_step |
| 1 | 1 | 0 | `Syntax Error: function 'X' arguments must have compatible types, got Integer and String` | FailureScripts/swap1 |
| 1 | 1 | 0 | `Syntax Error: function 'X' arguments must have compatible types, got Integer and Void` | FailureScripts/swap1 |
| 1 | 1 | 0 | `Syntax Error: function 'X' arguments must have compatible types, got Nil and TObject` | FailureScripts/swap1 |
| 1 | 1 | 0 | `Syntax Error: function 'X' arguments must have compatible types, got String and Integer` | FailureScripts/swap1 |
| 1 | 1 | 0 | `Syntax Error: function 'X' arguments must have compatible types, got TA(TObject) and Nil` | FailureScripts/swap1 |
| 1 | 1 | 0 | `Syntax Error: function 'X' arguments must have compatible types, got TA(TObject) and TObject` | FailureScripts/swap1 |
| 1 | 1 | 0 | `Syntax Error: function 'X' arguments must have compatible types, got TObject and TA(TObject)` | FailureScripts/swap1 |
| 1 | 1 | 0 | `Syntax Error: function 'X' delta must be Integer, got String` | FailureScripts/internal_unsupported |
| 1 | 1 | 0 | `Syntax Error: function 'X' element argument has type Integer, expected TMyEnum` | SetOfFail/invalid_operand |
| 1 | 1 | 0 | `Syntax Error: function 'X' element argument has type String, expected TMyEnum` | SetOfFail/invalid_operand |
| 1 | 1 | 0 | `Syntax Error: function 'X' expects Integer as first argument, got Float` | ArrayPass/dynamic_anonymous_record |
| 1 | 1 | 1 | `Syntax Error: function 'X' expects Integer or Enum variable, got TMyObj(TObject)` | FailureScripts/passing_prop_var |
| 1 | 1 | 1 | `Syntax Error: function 'X' expects N argument, got N` | FailureScripts/invalid_cast |
| 1 | 1 | 0 | `Syntax Error: function 'X' expects N or N arguments, got N` | FailureScripts/func_toomanyargs |
| 1 | 1 | 0 | `Syntax Error: function 'X' expects N-N arguments, got N` | FailureScripts/assert |
| 1 | 1 | 0 | `Syntax Error: function 'X' expects String as first argument, got function(): String` | SimpleScripts/exception_nested_call2 |
| 1 | 1 | 0 | `Syntax Error: function 'X' expects a type name as argument` | FailureScripts/default_func1 |
| 1 | 1 | 0 | `Syntax Error: function 'X' expects array or string, got Integer` | FailureScripts/internal_unsupported |
| 1 | 1 | 1 | `Syntax Error: function 'X' expects array, enum, or type name, got Void` | FailureScripts/special_funcs2 |
| 1 | 1 | 0 | `Syntax Error: function 'X' first argument must be Boolean, got Integer` | FailureScripts/assert |
| 1 | 1 | 0 | `Syntax Error: incompatible types in coalesce operator: String and Integer` | FailureScripts/coalesce |
| 1 | 1 | 0 | `Syntax Error: incompatible types in coalesce operator: array of TSub(TTest) and array of TObject` | FailureScripts/coalesce_dynarray |
| 1 | 1 | 0 | `Syntax Error: incompatible types in coalesce operator: array of TSub(TTest) and array of TTest(TObject)` | FailureScripts/coalesce_dynarray |
| 1 | 1 | 0 | `Syntax Error: incompatible types in coalesce operator: array of TTest(TObject) and array of TObject` | FailureScripts/coalesce_dynarray |
| 1 | 1 | 0 | `Syntax Error: incompatible types in coalesce operator: class of TObject and Integer` | FailureScripts/coalesce |
| 1 | 1 | 0 | `Syntax Error: incompatible types in coalesce operator: procedure() and String` | FailureScripts/coalesce |
| 1 | 1 | 0 | `Syntax Error: incompatible types in if-then-else: Integer and Nil` | FailureScripts/ifthenelse_expression4 |
| 1 | 1 | 0 | `Syntax Error: incompatible types in if-then-else: Integer and procedure()` | FailureScripts/ifthenelse_expression3 |
| 1 | 1 | 0 | `Syntax Error: incompatible types in if-then-else: Nil and String` | FailureScripts/ifthenelse_expression4 |
| 1 | 1 | 0 | `Syntax Error: incompatible types in if-then-else: String and Integer` | FailureScripts/ifthenelse_expression3 |
| 1 | 1 | 0 | `Syntax Error: incompatible types in if-then-else: procedure() and Integer` | FailureScripts/ifthenelse_expression3 |
| 1 | 1 | 0 | `Syntax Error: indexed property cannot have empty parameter list` | FailureScripts/array_params1 |
| 1 | 1 | 0 | `Syntax Error: inferred lambda return type function(): Integer incompatible with expected return type Integer` | LambdaFail/no_local_func |
| 1 | 1 | 1 | `Syntax Error: method, property, or field 'X' not found in parent class 'X'` | FailureScripts/inherited2 |
| 1 | 1 | 1 | `Syntax Error: non-external class 'X' cannot inherit from external class 'X'` | FailureScripts/external_overload |
| 1 | 1 | 1 | `Syntax Error: old() references undefined identifier 'X' in function 'X'` | FailureScripts/contracts_old |
| 1 | 1 | 0 | `Syntax Error: operator -= not supported for type TObject` | FailureScripts/assign_op_incompatible |
| 1 | 1 | 0 | `Syntax Error: operator /= not supported for type String` | FailureScripts/assign_op_incompatible |
| 1 | 1 | 1 | `Syntax Error: operator declaration requires at least one operand type` | OperatorOverloadFail/operator_overload2 |
| 1 | 1 | 0 | `Syntax Error: parameter modifiers are mutually exclusive` | FailureScripts/lazy |
| 1 | 1 | 0 | `Syntax Error: parent interface 'X' not found` | InterfacesFail/partial_declaration2 |
| 1 | 1 | 1 | `Syntax Error: property 'X' cannot be read-accessed` | PropertyExpressionsFail/read_self |
| 1 | 1 | 0 | `Syntax Error: property 'X' cannot combine index parameters with an index directive` | FailureScripts/property_error2 |
| 1 | 1 | 0 | `Syntax Error: property 'X' getter method 'X' parameter N has type Integer, expected String` | FailureScripts/property_error4 |
| 1 | 1 | 0 | `Syntax Error: property 'X' getter method 'X' returns Integer, expected String` | FailureScripts/property_error3 |
| 1 | 1 | 0 | `Syntax Error: property 'X' getter method 'X' returns String, expected TMypropertyType` | FailureScripts/property_error1 |
| 1 | 1 | 0 | `Syntax Error: property 'X' getter method 'X' returns Void, expected String` | FailureScripts/property_error3 |
| 1 | 1 | 0 | `Syntax Error: property 'X' index directive must be an integer literal` | FailureScripts/property_error2 |
| 1 | 1 | 1 | `Syntax Error: property 'X' read expression has type array of Variant, expected String` | PropertyExpressionsFail/null_read_expression |
| 1 | 1 | 0 | `Syntax Error: property 'X' write field 'X' has type Integer, expected String` | FailureScripts/property_error4 |
| 1 | 1 | 0 | `Syntax Error: range end must be an ordinal type, got Void` | FailureScripts/in_operator8 |
| 1 | 1 | 1 | `Syntax Error: required parameter 'X' cannot come after optional parameters in function 'X'` | FailureScripts/default_params1 |
| 1 | 1 | 0 | `Syntax Error: type mismatch in set literal: expected String, got Integer` | FailureScripts/in_typecheck1 |
| 1 | 1 | 0 | `Syntax Error: type mismatch in set literal: expected set of Integer, got set of String` | FailureScripts/in_typecheck1 |
| 1 | 1 | 0 | `Syntax Error: type parameter name expected` | GenericsFail/array1-2 |
| 1 | 1 | 0 | `Syntax Error: unary + requires numeric operand, got TObject` | FailureScripts/plus_non_numeric |
| 1 | 1 | 1 | `Syntax Error: unary not requires Boolean, Integer, or Variant operand, got procedure()` | FailureScripts/not_untyped |
| 1 | 1 | 1 | `Syntax Error: unbound method pointers (@TClass.Destroy) are not supported` | FailureScripts/func_ptr5 |
| 1 | 1 | 0 | `Syntax Error: unbound method pointers (@TClass.Free) are not supported` | FailureScripts/func_ptr6 |
| 1 | 1 | 0 | `Syntax Error: unexpected token in helper body: ;` | HelpersFail/helper_error5 |
| 1 | 1 | 0 | `Syntax Error: unexpected token in helper body: protected` | HelpersFail/helper_scopes1 |
| 1 | 1 | 0 | `Syntax Error: unknown return type 'X' in interface method 'X'` | InterfacesPass/intf_self_ref |
| 1 | 1 | 0 | `Syntax Error: unknown target type 'X' for helper 'X'` | HelpersFail/helper_as_type |
| 1 | 1 | 0 | `Syntax Error: var parameter N to function 'X' requires a variable (identifier, array element, or field), got N` | FailureScripts/passing_const_var2 |
| 1 | 1 | 0 | `Syntax Error: var parameter N to function 'X' requires a variable (identifier, array element, or field), got Self` | FailureScripts/self_not_writable |
| 1 | 1 | 0 | `Syntax Error: var parameter N to function 'X' requires a variable (identifier, array element, or field), got VarTest(i)` | FailureScripts/passing_const_var2 |
| 1 | 1 | 0 | `Syntax Error: var parameter N to function 'X' requires a variable (identifier, array element, or field), got p()` | FailureScripts/func_ptr_var_param |
| 1 | 1 | 1 | `Unsupported character #N` | FunctionsString/toxml |
| 1 | 1 | 0 | `Upper bound exceeded! Index N` | JSONConnectorPass/write_immediate_prop |
| 1 | 1 | 0 | `member assignment not supported for type STRING` | JSONConnectorPass/write_immediate_prop |

## Spurious shapes grouped by origin

Where each spurious shape is emitted, so the shape-by-shape work in `PLAN.md` §4 / F8 can be
batched by the file that has to change and ticked off here.

**How the sites were found.** Each shape's sentence was split on its placeholders, and the
resulting literal fragments were searched for in non-test `.go` files under `internal/`, `pkg/` and
`cmd/`, counting only matches **inside a Go string literal**. The most distinctive fragment — the
one with the fewest matches, at most 30 — wins, and is quoted at the end of every line so the
search can be repeated. A shape is filed under the package holding most of its sites, ties broken
in F8's batching order (parser, lexer, calls, statements, classes, other semantic, frontend,
other). Two consequences to read the list with:

- A long site list means the fragment was **not distinctive**, not that the shape has many
  emitters. Those entries need a human look before they are worked.
- Several shapes are DWScript's own sentence emitted in the wrong place or at the wrong anchor, so
  they resolve to the shared builder in `internal/errors/errors.go` rather than to the analyzer
  site that called it. They are filed under *other* for that reason.

### Parser — `internal/parser` — 117 shapes, 312 lines, 154 fixtures

- [ ] `Syntax Error: Expression expected` (55 fixtures, 55 lines) — `internal/frontend/result.go:857`, `internal/frontend/result.go:883`, `internal/parser/expressions_calls.go:35`, `internal/parser/parser.go:326`, `internal/parser/parser.go:331`, `internal/parser/parser.go:335`, `internal/semantic/analyze_array_helpers.go:469`, `internal/semantic/analyze_expressions.go:526`, `+2 more` — matched `Expression expected`
- [ ] `Syntax Error: expected 'X' after field name or method/property declaration keyword` (12 fixtures, 23 lines) — `internal/parser/classes.go:373` — matched `after field name or method/property declaration keyword`
- [ ] `Syntax Error: expected 'X' to close class declaration` (11 fixtures, 11 lines) — `internal/frontend/result.go:854`, `internal/parser/classes.go:503` — matched `to close class declaration`
- [ ] `Syntax Error: expected 'X' after parameter list` (8 fixtures, 11 lines) — `internal/parser/functions.go:405`, `internal/parser/functions.go:601`, `internal/parser/functions.go:817`, `internal/parser/interfaces.go:626`, `internal/parser/types.go:338` — matched `after parameter list`
- [ ] `Syntax Error: expected 'X' or 'X', got SEMICOLON` (6 fixtures, 6 lines) — `internal/parser/error.go:41` — matched `SEMICOLON`
- [ ] `Syntax Error: Expression expected before ASSIGN` (5 fixtures, 9 lines) — `internal/frontend/result.go:883`, `internal/parser/parser.go:331` — matched `Expression expected before`
- [ ] `Syntax Error: Expression expected before COMMA` (5 fixtures, 6 lines) — `internal/frontend/result.go:883`, `internal/parser/parser.go:331` — matched `Expression expected before`
- [ ] `Syntax Error: Expression expected before TO` (5 fixtures, 6 lines) — `internal/frontend/result.go:883`, `internal/parser/parser.go:331` — matched `Expression expected before`
- [ ] `Syntax Error: Expression expected before EQ` (5 fixtures, 5 lines) — `internal/frontend/result.go:883`, `internal/parser/parser.go:331` — matched `Expression expected before`
- [ ] `Syntax Error: expected parameter name` (4 fixtures, 6 lines) — `internal/parser/functions.go:675`, `internal/parser/functions.go:694`, `internal/parser/properties.go:344`, `internal/parser/records.go:491` — matched `expected parameter name`
- [ ] `Syntax Error: expected identifier after 'X'` (4 fixtures, 4 lines) — `internal/frontend/result.go:725`, `internal/frontend/result.go:733`, `internal/parser/control_flow.go:688`, `internal/parser/exceptions.go:457`, `internal/parser/expressions_contracts.go:84`, `internal/parser/interfaces.go:352`, `internal/parser/operators.go:194`, `internal/parser/operators.go:79`, `+3 more` — matched `expected identifier after`
- [ ] `Syntax Error: invalid assignment target` (3 fixtures, 5 lines) — `internal/interp/evaluator/visitor_statements.go:567`, `internal/parser/statements.go:310`, `internal/semantic/analyze_statements.go:743` — matched `invalid assignment target`
- [ ] `Syntax Error: Expression expected before CLASS` (3 fixtures, 3 lines) — `internal/frontend/result.go:883`, `internal/parser/parser.go:331` — matched `Expression expected before`
- [ ] `Syntax Error: Expression expected before INDEX` (3 fixtures, 3 lines) — `internal/frontend/result.go:883`, `internal/parser/parser.go:331` — matched `Expression expected before`
- [ ] `Syntax Error: expected 'X' after if condition` (3 fixtures, 3 lines) — `internal/parser/control_flow.go:126`, `internal/parser/control_flow.go:204`, `internal/parser/error_recovery.go:104`, `internal/parser/error_recovery.go:186`, `internal/parser/error_recovery.go:22` — matched `after if condition`
- [ ] `Syntax Error: expected 'X' in operator declaration` (3 fixtures, 3 lines) — `internal/parser/operators.go:74`, `internal/parser/operators.go:79`, `internal/semantic/analyze_operators.go:29`, `internal/semantic/analyze_operators.go:40` — matched `in operator declaration`
- [ ] `Syntax Error: expected 'X' to close interface declaration` (3 fixtures, 3 lines) — `internal/parser/interfaces.go:818` — matched `to close interface declaration`
- [ ] `Syntax Error: expected identifier in class inheritance list` (3 fixtures, 3 lines) — `internal/parser/classes.go:124` — matched `expected identifier in class inheritance list`
- [ ] `Syntax Error: expected next token to be RPAREN, got SEMICOLON instead` (3 fixtures, 3 lines) — `internal/parser/helpers.go:97`, `internal/parser/parser.go:237` — matched `expected next token to be`
- [ ] `Syntax Error: expected next token to be SEMICOLON, got REINTRODUCE instead` (3 fixtures, 3 lines) — `internal/parser/helpers.go:97`, `internal/parser/parser.go:237` — matched `expected next token to be`
- [ ] `Syntax Error: expected type after 'X' operator` (3 fixtures, 3 lines) — `internal/parser/expressions_typecheck.go:66`, `internal/parser/expressions_typecheck.go:92` — matched `expected type after`
- [ ] `Syntax Error: expected type expression, got )` (3 fixtures, 3 lines) — `internal/parser/types.go:116` — matched `expected type expression, got`
- [ ] `Syntax Error: expected unit name after 'X'` (2 fixtures, 5 lines) — `internal/parser/unit.go:139`, `internal/parser/unit.go:172` — matched `expected unit name after`
- [ ] `Syntax Error: expected 'X' after constant value` (2 fixtures, 4 lines) — `internal/parser/classes.go:883` — matched `after constant value`
- [ ] `Syntax Error: Expression expected before PROPERTY` (2 fixtures, 3 lines) — `internal/frontend/result.go:883`, `internal/parser/parser.go:331` — matched `Expression expected before`
- [ ] `Syntax Error: expected 'X' after for-in collection` (2 fixtures, 3 lines) — `internal/parser/control_flow.go:945` — matched `after for-in collection`
- [ ] `Syntax Error: "X" expected in generic type parameter list` (2 fixtures, 2 lines) — `internal/parser/interfaces.go:329` — matched `expected in generic type parameter list`
- [ ] `Syntax Error: Expression expected before ARRAY` (2 fixtures, 2 lines) — `internal/frontend/result.go:883`, `internal/parser/parser.go:331` — matched `Expression expected before`
- [ ] `Syntax Error: Expression expected before DOTDOT` (2 fixtures, 2 lines) — `internal/frontend/result.go:883`, `internal/parser/parser.go:331` — matched `Expression expected before`
- [ ] `Syntax Error: Expression expected before EQ_EQ` (2 fixtures, 2 lines) — `internal/frontend/result.go:883`, `internal/parser/parser.go:331` — matched `Expression expected before`
- [ ] `Syntax Error: complex return types not yet supported in function pointers` (2 fixtures, 2 lines) — `internal/parser/types.go:375` — matched `complex return types not yet supported in function pointers`
- [ ] `Syntax Error: could not parse "X" as integer` (2 fixtures, 2 lines) — `internal/interp/runtime/conversion.go:44`, `internal/parser/expressions_literals.go:50` — matched `as integer`
- [ ] `Syntax Error: expected 'X' after 'X' in unit declaration` (2 fixtures, 2 lines) — `internal/parser/unit.go:108` — matched `in unit declaration`
- [ ] `Syntax Error: expected 'X' after for loop variable` (2 fixtures, 2 lines) — `internal/parser/control_flow.go:726`, `internal/parser/control_flow.go:886` — matched `after for loop variable`
- [ ] `Syntax Error: expected 'X' or 'X' after const name` (2 fixtures, 2 lines) — `internal/parser/declarations.go:202` — matched `after const name`
- [ ] `Syntax Error: expected 'X' or 'X' in argument list, got EQ_EQ` (2 fixtures, 2 lines) — `internal/parser/expressions_calls.go:316` — matched `in argument list, got`
- [ ] `Syntax Error: expected 'X' or 'X' in for loop` (2 fixtures, 2 lines) — `internal/parser/control_flow.go:749`, `internal/parser/control_flow.go:764`, `internal/parser/control_flow.go:795` — matched `in for loop`
- [ ] `Syntax Error: expected 'X', 'X', 'X', 'X', 'X', 'X', 'X', 'X', 'X', 'X', or 'X' after 'X' in type declaration` (2 fixtures, 2 lines) — `internal/parser/interfaces.go:361`, `internal/parser/interfaces.go:545` — matched `in type declaration`
- [ ] `Syntax Error: expected 'X', got SEMICOLON` (2 fixtures, 2 lines) — `internal/parser/error.go:41` — matched `SEMICOLON`
- [ ] `Syntax Error: expected identifier in const declaration` (2 fixtures, 2 lines) — `internal/parser/declarations.go:143` — matched `expected identifier in const declaration`
- [ ] `Syntax Error: expected identifier in record field declaration` (2 fixtures, 2 lines) — `internal/frontend/result.go:731`, `internal/parser/classes.go:124`, `internal/parser/combinators.go:582`, `internal/parser/declarations.go:143`, `internal/parser/interfaces.go:361`, `internal/parser/statements.go:549`, `internal/parser/statements.go:607`, `internal/parser/statements.go:622`, `+1 more` — matched `expected identifier in`
- [ ] `Syntax Error: expected identifier or expression after 'X'` (2 fixtures, 2 lines) — `internal/parser/properties.go:171`, `internal/parser/properties.go:197` — matched `expected identifier or expression after`
- [ ] `Syntax Error: expected next token to be DOT, got SEMICOLON instead` (2 fixtures, 2 lines) — `internal/parser/helpers.go:97`, `internal/parser/parser.go:237` — matched `expected next token to be`
- [ ] `Syntax Error: expected next token to be IDENT, got READONLY instead` (2 fixtures, 2 lines) — `internal/parser/helpers.go:97`, `internal/parser/parser.go:237` — matched `expected next token to be`
- [ ] `Syntax Error: expected operator symbol after 'X'` (2 fixtures, 2 lines) — `internal/parser/operators.go:134`, `internal/parser/operators.go:34` — matched `expected operator symbol after`
- [ ] `Syntax Error: optional parameters cannot have lazy, var, or const modifiers` (2 fixtures, 2 lines) — `internal/parser/combinators.go:864`, `internal/parser/functions.go:734` — matched `optional parameters cannot have lazy, var, or const modifiers`
- [ ] `Syntax Error: Expression expected before EXPORT` (1 fixtures, 3 lines) — `internal/frontend/result.go:883`, `internal/parser/parser.go:331` — matched `Expression expected before`
- [ ] `Syntax Error: Expression expected before DEC` (1 fixtures, 2 lines) — `internal/frontend/result.go:883`, `internal/parser/parser.go:331` — matched `Expression expected before`
- [ ] `Syntax Error: Expression expected before INC` (1 fixtures, 2 lines) — `internal/frontend/result.go:883`, `internal/parser/parser.go:331` — matched `Expression expected before`
- [ ] `Syntax Error: expected 'X' after field declaration` (1 fixtures, 2 lines) — `internal/parser/classes.go:685`, `internal/parser/records.go:353` — matched `after field declaration`
- [ ] `Syntax Error: expected 'X' at end of operator declaration` (1 fixtures, 2 lines) — `internal/parser/operators.go:93` — matched `at end of operator declaration`
- [ ] `Syntax Error: expected next token to be OBJECT, got SEMICOLON instead` (1 fixtures, 2 lines) — `internal/parser/helpers.go:97`, `internal/parser/parser.go:237` — matched `expected next token to be`
- [ ] `Syntax Error: expected next token to be SEMICOLON, got DESCRIPTION instead` (1 fixtures, 2 lines) — `internal/parser/helpers.go:97`, `internal/parser/parser.go:237` — matched `expected next token to be`
- [ ] `Syntax Error: expected parameter name in indexed property` (1 fixtures, 2 lines) — `internal/parser/properties.go:344` — matched `expected parameter name in indexed property`
- [ ] `Syntax Error: "X" expected in generic type argument list` (1 fixtures, 1 lines) — `internal/parser/types.go:221` — matched `expected in generic type argument list`
- [ ] `Syntax Error: Expression expected before FINALIZATION` (1 fixtures, 1 lines) — `internal/frontend/result.go:883`, `internal/parser/parser.go:331` — matched `Expression expected before`
- [ ] `Syntax Error: Expression expected before FUNCTION` (1 fixtures, 1 lines) — `internal/frontend/result.go:883`, `internal/parser/parser.go:331` — matched `Expression expected before`
- [ ] `Syntax Error: Expression expected before GREATER` (1 fixtures, 1 lines) — `internal/frontend/result.go:883`, `internal/parser/parser.go:331` — matched `Expression expected before`
- [ ] `Syntax Error: Expression expected before INITIALIZATION` (1 fixtures, 1 lines) — `internal/frontend/result.go:883`, `internal/parser/parser.go:331` — matched `Expression expected before`
- [ ] `Syntax Error: Expression expected before LESS_LESS` (1 fixtures, 1 lines) — `internal/frontend/result.go:883`, `internal/parser/parser.go:331` — matched `Expression expected before`
- [ ] `Syntax Error: Expression expected before OVERLOAD` (1 fixtures, 1 lines) — `internal/frontend/result.go:883`, `internal/parser/parser.go:331` — matched `Expression expected before`
- [ ] `Syntax Error: Expression expected before READ` (1 fixtures, 1 lines) — `internal/frontend/result.go:883`, `internal/parser/parser.go:331` — matched `Expression expected before`
- [ ] `Syntax Error: Expression expected before READONLY` (1 fixtures, 1 lines) — `internal/frontend/result.go:883`, `internal/parser/parser.go:331` — matched `Expression expected before`
- [ ] `Syntax Error: Expression expected before STRICT` (1 fixtures, 1 lines) — `internal/frontend/result.go:883`, `internal/parser/parser.go:331` — matched `Expression expected before`
- [ ] `Syntax Error: array dimension N must be integer, got Boolean` (1 fixtures, 1 lines) — `internal/interp/evaluator/array_helpers.go:34`, `internal/interp/evaluator/array_helpers.go:44`, `internal/parser/expressions_oop.go:285`, `internal/parser/expressions_oop.go:308`, `internal/semantic/errors.go:713` — matched `array dimension`
- [ ] `Syntax Error: expected 'X' after 'X' in metaclass type` (1 fixtures, 1 lines) — `internal/parser/types.go:720` — matched `in metaclass type`
- [ ] `Syntax Error: expected 'X' after 'X' in set declaration` (1 fixtures, 1 lines) — `internal/parser/sets.go:33` — matched `in set declaration`
- [ ] `Syntax Error: expected 'X' after case expression` (1 fixtures, 1 lines) — `internal/parser/control_flow.go:1227` — matched `after case expression`
- [ ] `Syntax Error: expected 'X' after case value` (1 fixtures, 1 lines) — `internal/parser/control_flow.go:1092` — matched `after case value`
- [ ] `Syntax Error: expected 'X' after field name` (1 fixtures, 1 lines) — `internal/parser/classes.go:373`, `internal/parser/records.go:329` — matched `after field name`
- [ ] `Syntax Error: expected 'X' after interface method declaration` (1 fixtures, 1 lines) — `internal/parser/interfaces.go:906` — matched `after interface method declaration`
- [ ] `Syntax Error: expected 'X' after parent interface` (1 fixtures, 1 lines) — `internal/parser/interfaces.go:747` — matched `after parent interface`
- [ ] `Syntax Error: expected 'X' or 'X' after indexed property parameter` (1 fixtures, 1 lines) — `internal/parser/properties.go:65` — matched `after indexed property parameter`
- [ ] `Syntax Error: expected 'X' or 'X' in argument list, got RBRACK` (1 fixtures, 1 lines) — `internal/parser/expressions_calls.go:316` — matched `in argument list, got`
- [ ] `Syntax Error: expected 'X' or 'X' in class inheritance list` (1 fixtures, 1 lines) — `internal/parser/classes.go:124`, `internal/parser/classes.go:163` — matched `in class inheritance list`
- [ ] `Syntax Error: expected 'X' or 'X' in operator operand list` (1 fixtures, 1 lines) — `internal/parser/operators.go:243` — matched `in operator operand list`
- [ ] `Syntax Error: expected 'X' to close helper declaration` (1 fixtures, 1 lines) — `internal/parser/helpers.go:261` — matched `to close helper declaration`
- [ ] `Syntax Error: expected 'X' to close record declaration` (1 fixtures, 1 lines) — `internal/parser/records.go:55` — matched `to close record declaration`
- [ ] `Syntax Error: expected 'X' to close try statement` (1 fixtures, 1 lines) — `internal/parser/exceptions.go:221` — matched `to close try statement`
- [ ] `Syntax Error: expected 'X' to start enum declaration` (1 fixtures, 1 lines) — `internal/parser/enums.go:44` — matched `to start enum declaration`
- [ ] `Syntax Error: expected 'X', 'X', 'X' or 'X' after 'X' keyword in helper` (1 fixtures, 1 lines) — `internal/parser/helpers.go:209` — matched `keyword in helper`
- [ ] `Syntax Error: expected 'X', 'X', 'X', 'X' or 'X' after 'X' keyword in record` (1 fixtures, 1 lines) — `internal/parser/records.go:215` — matched `keyword in record`
- [ ] `Syntax Error: expected 'X', 'X', 'X', 'X', 'X', or 'X' after 'X' keyword` (1 fixtures, 1 lines) — `internal/interp/evaluator/visitor_statements.go:248`, `internal/lexer/lexer.go:693`, `internal/parser/classes.go:285`, `internal/parser/classes.go:373`, `internal/parser/classes.go:87`, `internal/parser/control_flow.go:1231`, `internal/parser/control_flow.go:208`, `internal/parser/control_flow.go:351`, `+11 more` — matched `keyword`
- [ ] `Syntax Error: expected class type after 'X'` (1 fixtures, 1 lines) — `internal/parser/types.go:732` — matched `expected class type after`
- [ ] `Syntax Error: expected condition after 'X'` (1 fixtures, 1 lines) — `internal/parser/control_flow.go:119`, `internal/parser/control_flow.go:186`, `internal/parser/control_flow.go:329`, `internal/parser/control_flow.go:471` — matched `expected condition after`
- [ ] `Syntax Error: expected enum value name, got INT` (1 fixtures, 1 lines) — `internal/parser/enums.go:76` — matched `expected enum value name, got`
- [ ] `Syntax Error: expected identifier after 'X' in operator declaration` (1 fixtures, 1 lines) — `internal/parser/operators.go:74`, `internal/parser/operators.go:79`, `internal/semantic/analyze_operators.go:29`, `internal/semantic/analyze_operators.go:40` — matched `in operator declaration`
- [ ] `Syntax Error: expected identifier for method name` (1 fixtures, 1 lines) — `internal/parser/interfaces.go:858` — matched `expected identifier for method name`
- [ ] `Syntax Error: expected identifier for parent interface` (1 fixtures, 1 lines) — `internal/parser/interfaces.go:720` — matched `expected identifier for parent interface`
- [ ] `Syntax Error: expected next token to be COLON, got END instead` (1 fixtures, 1 lines) — `internal/parser/helpers.go:97`, `internal/parser/parser.go:237` — matched `expected next token to be`
- [ ] `Syntax Error: expected next token to be COLON, got IDENT instead` (1 fixtures, 1 lines) — `internal/parser/helpers.go:97`, `internal/parser/parser.go:237` — matched `expected next token to be`
- [ ] `Syntax Error: expected next token to be COLON, got RBRACK instead` (1 fixtures, 1 lines) — `internal/parser/helpers.go:97`, `internal/parser/parser.go:237` — matched `expected next token to be`
- [ ] `Syntax Error: expected next token to be FOR` (1 fixtures, 1 lines) — `internal/parser/helpers.go:97` — matched `expected next token to be FOR`
- [ ] `Syntax Error: expected next token to be IDENT, got SEMICOLON instead` (1 fixtures, 1 lines) — `internal/parser/helpers.go:97`, `internal/parser/parser.go:237` — matched `expected next token to be`
- [ ] `Syntax Error: expected next token to be SEMICOLON, got DEFAULT instead` (1 fixtures, 1 lines) — `internal/parser/helpers.go:97`, `internal/parser/parser.go:237` — matched `expected next token to be`
- [ ] `Syntax Error: expected next token to be SEMICOLON, got INTERFACE instead` (1 fixtures, 1 lines) — `internal/parser/helpers.go:97`, `internal/parser/parser.go:237` — matched `expected next token to be`
- [ ] `Syntax Error: expected operand type in class operator declaration` (1 fixtures, 1 lines) — `internal/parser/operators.go:150` — matched `expected operand type in class operator declaration`
- [ ] `Syntax Error: expected program name after 'X' keyword` (1 fixtures, 1 lines) — `internal/parser/declarations.go:39` — matched `expected program name after`
- [ ] `Syntax Error: expected record field name, got PROPERTY` (1 fixtures, 1 lines) — `internal/parser/record_expressions.go:82` — matched `expected record field name, got`
- [ ] `Syntax Error: expected type expression after 'X'` (1 fixtures, 1 lines) — `internal/parser/arrays.go:421`, `internal/parser/declarations.go:182`, `internal/parser/sets.go:225`, `internal/parser/statements.go:654`, `internal/parser/types.go:533`, `internal/parser/types.go:549` — matched `expected type expression after`
- [ ] `Syntax Error: expected type expression, got` (1 fixtures, 1 lines) — `internal/parser/types.go:116` — matched `expected type expression, got`
- [ ] `Syntax Error: expected type expression, got ;` (1 fixtures, 1 lines) — `internal/parser/types.go:116` — matched `expected type expression, got`
- [ ] `Syntax Error: expected type expression, got >` (1 fixtures, 1 lines) — `internal/parser/types.go:116` — matched `expected type expression, got`
- [ ] `Syntax Error: expected type expression, got end` (1 fixtures, 1 lines) — `internal/parser/types.go:116` — matched `expected type expression, got`
- [ ] `Syntax Error: expected type expression, got hello` (1 fixtures, 1 lines) — `internal/parser/types.go:116` — matched `expected type expression, got`
- [ ] `Syntax Error: expected type expression, got interface` (1 fixtures, 1 lines) — `internal/parser/types.go:116` — matched `expected type expression, got`
- [ ] `Syntax Error: expected type expression, got nil` (1 fixtures, 1 lines) — `internal/parser/types.go:116` — matched `expected type expression, got`
- [ ] `Syntax Error: expected type expression, got world` (1 fixtures, 1 lines) — `internal/parser/types.go:116` — matched `expected type expression, got`
- [ ] `Syntax Error: expected type name after 'X'` (1 fixtures, 1 lines) — `internal/parser/expressions_oop.go:106`, `internal/parser/helpers.go:108` — matched `expected type name after`
- [ ] `Syntax Error: expected type name after 'X' in helper declaration` (1 fixtures, 1 lines) — `internal/parser/expressions_oop.go:106`, `internal/parser/helpers.go:108` — matched `expected type name after`
- [ ] `Syntax Error: indexed property cannot have empty parameter list` (1 fixtures, 1 lines) — `internal/parser/properties.go:46` — matched `indexed property cannot have empty parameter list`
- [ ] `Syntax Error: operator declaration requires at least one operand type` (1 fixtures, 1 lines) — `internal/parser/operators.go:168`, `internal/parser/operators.go:57` — matched `operator declaration requires at least one operand type`
- [ ] `Syntax Error: parameter modifiers are mutually exclusive` (1 fixtures, 1 lines) — `internal/parser/combinators.go:823`, `internal/parser/functions.go:661` — matched `parameter modifiers are mutually exclusive`
- [ ] `Syntax Error: parent interface 'X' not found` (1 fixtures, 1 lines) — `internal/interp/evaluator/visitor_declarations.go:537`, `internal/parser/interfaces.go:720`, `internal/parser/interfaces.go:747`, `internal/semantic/analyze_interfaces.go:31` — matched `parent interface`
- [ ] `Syntax Error: type parameter name expected` (1 fixtures, 1 lines) — `internal/parser/interfaces.go:303` — matched `type parameter name expected`
- [ ] `Syntax Error: unexpected token in helper body: ;` (1 fixtures, 1 lines) — `internal/parser/helpers.go:254` — matched `unexpected token in helper body:`
- [ ] `Syntax Error: unexpected token in helper body: protected` (1 fixtures, 1 lines) — `internal/parser/helpers.go:254` — matched `unexpected token in helper body:`

### Lexer — `internal/lexer` — 6 shapes, 6 lines, 5 fixtures

- [ ] `Syntax Error: Constant expression expected` (1 fixtures, 1 lines) — `internal/lexer/directives.go:478` — matched `Constant expression expected`
- [ ] `Syntax Error: String expected` (1 fixtures, 1 lines) — `internal/lexer/directive_messages.go:100`, `internal/lexer/directive_messages.go:158`, `internal/lexer/directives.go:220`, `internal/lexer/directives.go:496`, `internal/semantic/analyze_declared.go:49`, `internal/semantic/analyze_declared.go:64`, `internal/semantic/analyze_statements.go:293` — matched `String expected`
- [ ] `Syntax Error: Unbalanced conditional directive` (1 fixtures, 1 lines) — `internal/lexer/directive_messages.go:261`, `internal/lexer/directives.go:280`, `internal/lexer/directives.go:304` — matched `Unbalanced conditional directive`
- [ ] `Syntax Error: cannot open include file 'X': failed to read file: open testdata/fixtures/FailureScripts/include.INC: no such file or directory` (1 fixtures, 1 lines) — `internal/lexer/include.go:106` — matched `cannot open include file`
- [ ] `Syntax Error: cannot open include file 'X': failed to read file: open testdata/fixtures/FailureScripts/include.inc: no such file or directory` (1 fixtures, 1 lines) — `internal/lexer/include.go:106` — matched `cannot open include file`
- [ ] `Syntax Error: file name expected after $i` (1 fixtures, 1 lines) — `internal/lexer/include.go:89` — matched `file name expected after`

### Semantic — `analyze_function_calls.go` / `analyze_method_calls.go` — 8 shapes, 10 lines, 5 fixtures

- [ ] `Syntax Error: function 'X' first argument must be Boolean, got String` (1 fixtures, 2 lines) — `internal/builtins/system.go:120`, `internal/semantic/analyze_function_calls.go:440` — matched `first argument must be Boolean, got`
- [ ] `Syntax Error: function 'X' second argument must be String, got Integer` (1 fixtures, 2 lines) — `internal/builtins/system.go:134`, `internal/semantic/analyze_builtin_string_transform.go:96`, `internal/semantic/analyze_function_calls.go:446`, `internal/semantic/analyze_function_calls.go:471` — matched `second argument must be String, got`
- [ ] `Syntax Error: Not a method` (1 fixtures, 1 lines) — `internal/frontend/result.go:713`, `internal/semantic/analyze_function_calls.go:234`, `internal/semantic/analyze_function_calls.go:240` — matched `Not a method`
- [ ] `Syntax Error: function 'X' first argument must be Boolean, got Integer` (1 fixtures, 1 lines) — `internal/builtins/system.go:120`, `internal/semantic/analyze_function_calls.go:440` — matched `first argument must be Boolean, got`
- [ ] `Syntax Error: var parameter N to function 'X' requires a variable (identifier, array element, or field), got N` (1 fixtures, 1 lines) — `internal/semantic/analyze_builtin_registry.go:101`, `internal/semantic/analyze_function_calls.go:781` — matched `to function`
- [ ] `Syntax Error: var parameter N to function 'X' requires a variable (identifier, array element, or field), got Self` (1 fixtures, 1 lines) — `internal/semantic/analyze_builtin_registry.go:101`, `internal/semantic/analyze_function_calls.go:781` — matched `to function`
- [ ] `Syntax Error: var parameter N to function 'X' requires a variable (identifier, array element, or field), got VarTest(i)` (1 fixtures, 1 lines) — `internal/semantic/analyze_builtin_registry.go:101`, `internal/semantic/analyze_function_calls.go:781` — matched `to function`
- [ ] `Syntax Error: var parameter N to function 'X' requires a variable (identifier, array element, or field), got p()` (1 fixtures, 1 lines) — `internal/semantic/analyze_builtin_registry.go:101`, `internal/semantic/analyze_function_calls.go:781` — matched `to function`

### Semantic — `analyze_statements.go` — 44 shapes, 64 lines, 30 fixtures

- [ ] `Syntax Error: cannot infer type for variable 'X' from initializer` (4 fixtures, 6 lines) — `internal/semantic/analyze_statements.go:202` — matched `from initializer`
- [ ] `Syntax Error: constant 'X' must have a value` (4 fixtures, 4 lines) — `internal/interp/evaluator/visitor_statements.go:403`, `internal/semantic/analyze_classes_decl.go:501`, `internal/semantic/analyze_statements.go:265` — matched `must have a value`
- [ ] `Syntax Error: cannot assign to read-only variable 'X'` (3 fixtures, 4 lines) — `internal/semantic/analyze_statements.go:482` — matched `cannot assign to read-only variable`
- [ ] `Hint: Empty FOR loop` (3 fixtures, 3 lines) — `internal/semantic/analyze_statements.go:1216`, `internal/semantic/analyze_statements.go:1361` — matched `Empty FOR loop`
- [ ] `Hint: Empty THEN block` (2 fixtures, 2 lines) — `internal/semantic/analyze_statements.go:984` — matched `Empty THEN block`
- [ ] `Syntax Error: Cannot assign to constant 'X'` (2 fixtures, 2 lines) — `internal/semantic/analyze_statements.go:480`, `internal/semantic/errors.go:305` — matched `Cannot assign to constant`
- [ ] `Syntax Error: cannot infer type for variable 'X' from nil initializer` (2 fixtures, 2 lines) — `internal/semantic/analyze_statements.go:212` — matched `from nil initializer`
- [ ] `Syntax Error: break statement not allowed in finally block` (1 fixtures, 2 lines) — `internal/semantic/analyze_statements.go:1485` — matched `break statement not allowed in finally block`
- [ ] `Syntax Error: continue statement not allowed in finally block` (1 fixtures, 2 lines) — `internal/semantic/analyze_statements.go:1503` — matched `continue statement not allowed in finally block`
- [ ] `Syntax Error: operator += not supported for type JSONVariant` (1 fixtures, 2 lines) — `internal/interp/evaluator/compound_ops.go:136`, `internal/semantic/analyze_statements.go:851`, `internal/semantic/analyze_statements.go:854` — matched `operator += not supported for type`
- [ ] `Syntax Error: operator += not supported for type array of Integer` (1 fixtures, 2 lines) — `internal/interp/evaluator/compound_ops.go:136`, `internal/semantic/analyze_statements.go:851`, `internal/semantic/analyze_statements.go:854` — matched `operator += not supported for type`
- [ ] `Syntax Error: break statement not allowed outside loop` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1491` — matched `break statement not allowed outside loop`
- [ ] `Syntax Error: cannot assign to constant 'X'` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:564` — matched `cannot assign to constant`
- [ ] `Syntax Error: case range end type Boolean incompatible with case expression type Float` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1440`, `internal/semantic/analyze_statements.go:1448`, `internal/semantic/analyze_statements.go:1465` — matched `incompatible with case expression type`
- [ ] `Syntax Error: case range end type Boolean incompatible with case expression type Integer` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1440`, `internal/semantic/analyze_statements.go:1448`, `internal/semantic/analyze_statements.go:1465` — matched `incompatible with case expression type`
- [ ] `Syntax Error: case range end type Boolean incompatible with case expression type String` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1440`, `internal/semantic/analyze_statements.go:1448`, `internal/semantic/analyze_statements.go:1465` — matched `incompatible with case expression type`
- [ ] `Syntax Error: case range end type Float incompatible with case expression type Boolean` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1440`, `internal/semantic/analyze_statements.go:1448`, `internal/semantic/analyze_statements.go:1465` — matched `incompatible with case expression type`
- [ ] `Syntax Error: case range end type Float incompatible with case expression type Integer` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1440`, `internal/semantic/analyze_statements.go:1448`, `internal/semantic/analyze_statements.go:1465` — matched `incompatible with case expression type`
- [ ] `Syntax Error: case range end type Float incompatible with case expression type String` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1440`, `internal/semantic/analyze_statements.go:1448`, `internal/semantic/analyze_statements.go:1465` — matched `incompatible with case expression type`
- [ ] `Syntax Error: case range end type Integer incompatible with case expression type Boolean` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1440`, `internal/semantic/analyze_statements.go:1448`, `internal/semantic/analyze_statements.go:1465` — matched `incompatible with case expression type`
- [ ] `Syntax Error: case range end type Integer incompatible with case expression type String` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1440`, `internal/semantic/analyze_statements.go:1448`, `internal/semantic/analyze_statements.go:1465` — matched `incompatible with case expression type`
- [ ] `Syntax Error: case range end type String incompatible with case expression type Boolean` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1440`, `internal/semantic/analyze_statements.go:1448`, `internal/semantic/analyze_statements.go:1465` — matched `incompatible with case expression type`
- [ ] `Syntax Error: case range end type String incompatible with case expression type Float` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1440`, `internal/semantic/analyze_statements.go:1448`, `internal/semantic/analyze_statements.go:1465` — matched `incompatible with case expression type`
- [ ] `Syntax Error: case range end type String incompatible with case expression type Integer` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1440`, `internal/semantic/analyze_statements.go:1448`, `internal/semantic/analyze_statements.go:1465` — matched `incompatible with case expression type`
- [ ] `Syntax Error: case range start type Boolean and end type Float are incompatible` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1440`, `internal/semantic/analyze_statements.go:1456` — matched `case range start type`
- [ ] `Syntax Error: case range start type Boolean incompatible with case expression type Float` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1440`, `internal/semantic/analyze_statements.go:1448`, `internal/semantic/analyze_statements.go:1465` — matched `incompatible with case expression type`
- [ ] `Syntax Error: case range start type Boolean incompatible with case expression type Integer` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1440`, `internal/semantic/analyze_statements.go:1448`, `internal/semantic/analyze_statements.go:1465` — matched `incompatible with case expression type`
- [ ] `Syntax Error: case range start type Boolean incompatible with case expression type String` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1440`, `internal/semantic/analyze_statements.go:1448`, `internal/semantic/analyze_statements.go:1465` — matched `incompatible with case expression type`
- [ ] `Syntax Error: case range start type Float incompatible with case expression type Boolean` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1440`, `internal/semantic/analyze_statements.go:1448`, `internal/semantic/analyze_statements.go:1465` — matched `incompatible with case expression type`
- [ ] `Syntax Error: case range start type Float incompatible with case expression type Integer` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1440`, `internal/semantic/analyze_statements.go:1448`, `internal/semantic/analyze_statements.go:1465` — matched `incompatible with case expression type`
- [ ] `Syntax Error: case range start type Float incompatible with case expression type String` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1440`, `internal/semantic/analyze_statements.go:1448`, `internal/semantic/analyze_statements.go:1465` — matched `incompatible with case expression type`
- [ ] `Syntax Error: case range start type Integer and end type String are incompatible` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1440`, `internal/semantic/analyze_statements.go:1456` — matched `case range start type`
- [ ] `Syntax Error: case range start type Integer incompatible with case expression type Boolean` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1440`, `internal/semantic/analyze_statements.go:1448`, `internal/semantic/analyze_statements.go:1465` — matched `incompatible with case expression type`
- [ ] `Syntax Error: case range start type Integer incompatible with case expression type String` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1440`, `internal/semantic/analyze_statements.go:1448`, `internal/semantic/analyze_statements.go:1465` — matched `incompatible with case expression type`
- [ ] `Syntax Error: case range start type String and end type class of TObject are incompatible` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1440`, `internal/semantic/analyze_statements.go:1456` — matched `case range start type`
- [ ] `Syntax Error: case range start type String incompatible with case expression type Boolean` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1440`, `internal/semantic/analyze_statements.go:1448`, `internal/semantic/analyze_statements.go:1465` — matched `incompatible with case expression type`
- [ ] `Syntax Error: case range start type String incompatible with case expression type Float` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1440`, `internal/semantic/analyze_statements.go:1448`, `internal/semantic/analyze_statements.go:1465` — matched `incompatible with case expression type`
- [ ] `Syntax Error: case range start type String incompatible with case expression type Integer` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1440`, `internal/semantic/analyze_statements.go:1448`, `internal/semantic/analyze_statements.go:1465` — matched `incompatible with case expression type`
- [ ] `Syntax Error: case range start type Void incompatible with case expression type Integer` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1440`, `internal/semantic/analyze_statements.go:1448`, `internal/semantic/analyze_statements.go:1465` — matched `incompatible with case expression type`
- [ ] `Syntax Error: continue statement not allowed outside loop` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1509` — matched `continue statement not allowed outside loop`
- [ ] `Syntax Error: exit statement not allowed in finally block` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1521` — matched `exit statement not allowed in finally block`
- [ ] `Syntax Error: exit value type Void incompatible with function return type Integer` (1 fixtures, 1 lines) — `internal/semantic/analyze_functions.go:308`, `internal/semantic/analyze_statements.go:1564` — matched `incompatible with function return type`
- [ ] `Syntax Error: exit with value not allowed in procedure` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1556` — matched `exit with value not allowed in procedure`
- [ ] `Syntax Error: for loop step must be Integer, got String` (1 fixtures, 1 lines) — `internal/semantic/analyze_statements.go:1174` — matched `for loop step must be Integer, got`

### Semantic — `analyze_classes*.go` — 12 shapes, 58 lines, 39 fixtures

- [ ] `Hint: Result is never used` (9 fixtures, 12 lines) — `internal/semantic/analyze_classes_decl.go:1131`, `internal/semantic/analyze_classes_decl.go:821`, `internal/semantic/analyze_functions.go:239`, `internal/semantic/unused_warnings.go:104` — matched `Result is never used`
- [ ] `Syntax Error: duplicate method signature for 'X'` (9 fixtures, 9 lines) — `internal/semantic/analyze_classes_decl.go:1018` — matched `duplicate method signature for`
- [ ] `Syntax Error: method 'X' not declared in class 'X'` (7 fixtures, 11 lines) — `internal/semantic/analyze_classes_decl.go:680` — matched `not declared in class`
- [ ] `Syntax Error: parent class 'X' not found` (7 fixtures, 7 lines) — `internal/interp/evaluator/call_helpers.go:268`, `internal/interp/evaluator/call_helpers.go:279`, `internal/interp/evaluator/call_helpers.go:283`, `internal/interp/evaluator/class_alias.go:24`, `internal/interp/evaluator/class_alias.go:41`, `internal/interp/evaluator/visitor_declarations.go:309`, `internal/interp/runtime/object.go:325`, `internal/interp/runtime/object.go:336`, `+13 more` — matched `parent class`
- [ ] `Syntax Error: circular inheritance detected in class 'X'` (2 fixtures, 4 lines) — `internal/semantic/analyze_classes_decl.go:334` — matched `circular inheritance detected in class`
- [ ] `Hint: Overloaded method "X" should be marked with the "X" directive` (2 fixtures, 3 lines) — `internal/semantic/analyze_classes_validation.go:143` — matched `should be marked with the`
- [ ] `Syntax Error: 'X' requires a class or class reference` (2 fixtures, 2 lines) — `internal/semantic/analyze_classes.go:123` — matched `requires a class or class reference`
- [ ] `Syntax Error: ambiguous overload for 'X'` (2 fixtures, 2 lines) — `internal/semantic/analyze_classes_decl.go:1022` — matched `ambiguous overload for`
- [ ] `Syntax Error: partial class 'X' has conflicting parent classes` (2 fixtures, 2 lines) — `internal/interp/evaluator/visitor_declarations.go:309`, `internal/semantic/analyze_classes_decl.go:81` — matched `has conflicting parent classes`
- [ ] `Hint: Previous declaration of class was "X"` (1 fixtures, 4 lines) — `internal/semantic/analyze_classes_decl.go:51` — matched `Previous declaration of class was`
- [ ] `Syntax Error: external class 'X' cannot inherit from non-external class 'X'` (1 fixtures, 1 lines) — `internal/semantic/analyze_classes_decl.go:310` — matched `cannot inherit from non-external class`
- [ ] `Syntax Error: non-external class 'X' cannot inherit from external class 'X'` (1 fixtures, 1 lines) — `internal/semantic/analyze_classes_decl.go:316` — matched `cannot inherit from external class`

### Semantic — other files — 121 shapes, 264 lines, 143 fixtures

- [ ] `Syntax Error: Undefined variable 'X'` (12 fixtures, 14 lines) — `internal/semantic/errors.go:142`, `internal/semantic/errors.go:241` — matched `Undefined variable`
- [ ] `Hint: Variable "X" declared but not used` (11 fixtures, 15 lines) — `internal/semantic/unused_warnings.go:108` — matched `declared but not used`
- [ ] `Hint: "X" does not match case of declaration ("X")` (5 fixtures, 6 lines) — `internal/semantic/analyzer.go:708` — matched `does not match case of declaration (`
- [ ] `Syntax Error: No arguments expected` (5 fixtures, 6 lines) — `internal/frontend/result.go:487`, `internal/semantic/analyze_arity.go:15`, `internal/semantic/analyze_array_helpers.go:226` — matched `No arguments expected`
- [ ] `Syntax Error: property 'X' write specifier 'X' not found in class 'X'` (5 fixtures, 6 lines) — `internal/interp/evaluator/class_property_helpers.go:105`, `internal/interp/evaluator/property_write.go:176`, `internal/semantic/analyze_interfaces.go:178`, `internal/semantic/analyze_properties.go:440`, `internal/semantic/analyze_properties.go:574` — matched `write specifier`
- [ ] `Syntax Error: Cannot read a write only property` (4 fixtures, 9 lines) — `internal/semantic/errors.go:567` — matched `Cannot read a write only property`
- [ ] `Warning: Infinite loop` (4 fixtures, 6 lines) — `internal/semantic/analyzer.go:1320` — matched `Infinite loop`
- [ ] `Syntax Error: address-of operator (@) requires a function or procedure name` (4 fixtures, 4 lines) — `internal/semantic/analyze_function_pointers.go:118` — matched `address-of operator (@) requires a function or procedure name`
- [ ] `Warning: Unit name does not match file name` (4 fixtures, 4 lines) — `internal/semantic/unit_analyzer.go:23` — matched `Unit name does not match file name`
- [ ] `Syntax Error: array element N has type String, expected Integer` (3 fixtures, 5 lines) — `internal/interp/evaluator/array_helpers.go:483`, `internal/interp/evaluator/array_helpers.go:509`, `internal/interp/evaluator/overload_resolution.go:430`, `internal/interp/ffi_callback.go:121`, `internal/interp/marshal.go:109`, `internal/parser/statements.go:312`, `internal/parser/types.go:537`, `internal/parser/types.go:553`, `+9 more` — matched `array element`
- [ ] `Syntax Error: Bound isn't of an ordinal type` (3 fixtures, 4 lines) — `internal/semantic/type_resolution.go:668`, `internal/semantic/type_resolution.go:675`, `internal/semantic/type_resolution.go:695` — matched `Bound isn't of an ordinal type`
- [ ] `Syntax Error: 'X' cannot be used in class 'X' which has no parent class` (3 fixtures, 3 lines) — `internal/semantic/analyze_special.go:193`, `internal/semantic/analyze_special.go:27` — matched `which has no parent class`
- [ ] `Syntax Error: Read access of property should be a static method` (3 fixtures, 3 lines) — `internal/semantic/errors.go:536` — matched `Read access of property should be a static method`
- [ ] `Syntax Error: type 'X' already declared` (3 fixtures, 3 lines) — `internal/frontend/result.go:891`, `internal/interp/evaluator/visitor_declarations.go:261`, `internal/semantic/analyze_types.go:715`, `internal/semantic/errors.go:274`, `internal/semantic/symbol_table.go:413` — matched `already declared`
- [ ] `Syntax Error: implementation signature for 'X' does not match forward declaration` (2 fixtures, 8 lines) — `internal/semantic/symbol_table.go:433`, `internal/semantic/symbol_table.go:439` — matched `implementation signature for`
- [ ] `Syntax Error: Array expected` (2 fixtures, 7 lines) — `internal/semantic/errors.go:690` — matched `Array expected`
- [ ] `Syntax Error: incompatible types in coalesce operator: array of TObject and array of TTest(TObject)` (2 fixtures, 4 lines) — `internal/semantic/analyze_expr_operators.go:371` — matched `incompatible types in coalesce operator:`
- [ ] `Syntax Error: unary - requires numeric operand, got String` (2 fixtures, 4 lines) — `internal/semantic/analyze_expr_operators.go:834` — matched `requires numeric operand, got`
- [ ] `Syntax Error: Cannot set a value for a read-only property` (2 fixtures, 3 lines) — `internal/interp/evaluator/visitor_expressions_errors.go:13`, `internal/semantic/errors.go:556` — matched `Cannot set a value for a read-only property`
- [ ] `Syntax Error: Unexpected "X"` (2 fixtures, 3 lines) — `internal/builtins/datetime_iso8601.go:88`, `internal/semantic/recovery_diagnostics.go:67` — matched `Unexpected`
- [ ] `Syntax Error: cannot infer type for class var 'X'` (2 fixtures, 3 lines) — `internal/interp/evaluator/visitor_declarations.go:1274`, `internal/semantic/analyze_classes_decl.go:540`, `internal/semantic/analyze_helpers.go:495`, `internal/semantic/analyze_records.go:213` — matched `cannot infer type for class var`
- [ ] `Syntax Error: function 'X' first argument must be a variable (identifier, array element, or field)` (2 fixtures, 3 lines) — `internal/semantic/analyze_builtin_array.go:198`, `internal/semantic/analyze_builtin_math_utils.go:21`, `internal/semantic/analyze_builtin_math_utils.go:55` — matched `first argument must be a variable (identifier, array element, or field)`
- [ ] `Syntax Error: incompatible types in coalesce operator: array of TObject and array of TSub(TTest)` (2 fixtures, 3 lines) — `internal/semantic/analyze_expr_operators.go:371` — matched `incompatible types in coalesce operator:`
- [ ] `Syntax Error: incompatible types in coalesce operator: array of TTest(TObject) and array of TSub(TTest)` (2 fixtures, 3 lines) — `internal/semantic/analyze_expr_operators.go:371` — matched `incompatible types in coalesce operator:`
- [ ] `Syntax Error: range start must be an ordinal type, got Float` (2 fixtures, 3 lines) — `internal/semantic/analyze_literals.go:479` — matched `range start must be an ordinal type, got`
- [ ] `Syntax Error: unknown type 'X' in type alias` (2 fixtures, 3 lines) — `internal/interp/evaluator/visitor_declarations.go:1457`, `internal/semantic/analyze_types.go:779` — matched `in type alias`
- [ ] `Hint: Constant Instruction - has no effect` (2 fixtures, 2 lines) — `internal/semantic/analyze_const_instruction.go:54` — matched `Constant Instruction - has no effect`
- [ ] `Syntax Error: Array index expected "X" but got "X"` (2 fixtures, 2 lines) — `internal/errors/errors.go:318`, `internal/semantic/errors.go:668` — matched `Array index expected`
- [ ] `Syntax Error: More arguments expected` (2 fixtures, 2 lines) — `internal/frontend/result.go:485`, `internal/semantic/analyze_arity.go:13`, `internal/semantic/analyze_array_helpers.go:212`, `internal/semantic/analyze_array_helpers.go:704`, `internal/semantic/analyze_array_helpers.go:710` — matched `More arguments expected`
- [ ] `Syntax Error: array constructor range bounds must be ordinal` (2 fixtures, 2 lines) — `internal/semantic/analyze_literals.go:243` — matched `array constructor range bounds must be ordinal`
- [ ] `Syntax Error: binding 'X' for class operator 'X' expects N parameters, got N` (2 fixtures, 2 lines) — `internal/builtins/array.go:357`, `internal/semantic/analyze_operators.go:209`, `internal/semantic/analyze_operators.go:63` — matched `parameters, got`
- [ ] `Syntax Error: binding 'X' for operator 'X' not found` (2 fixtures, 2 lines) — `internal/interp/evaluator/binary_ops.go:1246`, `internal/semantic/analyze_operators.go:221`, `internal/semantic/analyze_operators.go:52`, `internal/semantic/analyze_operators.go:58`, `internal/semantic/analyze_operators.go:63` — matched `for operator`
- [ ] `Syntax Error: cannot access private constant 'X' of class 'X'` (2 fixtures, 2 lines) — `internal/errors/errors.go:387`, `internal/interp/evaluator/member_assignment.go:342`, `internal/parser/operators.go:207`, `internal/semantic/analyze_classes.go:302`, `internal/semantic/analyze_classes_decl.go:51`, `internal/semantic/analyze_function_calls.go:1047`, `internal/semantic/analyze_function_pointers.go:167`, `internal/semantic/analyze_properties.go:73`, `+2 more` — matched `of class`
- [ ] `Syntax Error: cannot take address of variadic built-in function 'X'` (2 fixtures, 2 lines) — `internal/semantic/analyze_function_pointers.go:247` — matched `cannot take address of variadic built-in function`
- [ ] `Syntax Error: expected 'X' after 'X' operand, got SEMICOLON` (2 fixtures, 2 lines) — `internal/interp/evaluator/binary_ops.go:1340`, `internal/parser/expressions_oop.go:223`, `internal/semantic/analyze_expr_operators.go:745`, `internal/semantic/analyze_expr_operators.go:777`, `internal/semantic/analyze_expr_operators.go:834`, `internal/semantic/analyze_expr_operators.go:855` — matched `operand, got`
- [ ] `Syntax Error: inferred lambda return type Void incompatible with expected return type Integer` (2 fixtures, 2 lines) — `internal/semantic/analyze_lambdas.go:272`, `internal/semantic/analyze_lambdas.go:299` — matched `incompatible with expected return type`
- [ ] `Syntax Error: unary + requires numeric operand, got String` (2 fixtures, 2 lines) — `internal/semantic/analyze_expr_operators.go:834` — matched `requires numeric operand, got`
- [ ] `Syntax Error: incompatible types in coalesce operator: TSubN(TTest) and TSubN(TTest)` (1 fixtures, 6 lines) — `internal/semantic/analyze_expr_operators.go:371` — matched `incompatible types in coalesce operator:`
- [ ] `Syntax Error: 'X' cannot be used in class methods (static methods)` (1 fixtures, 5 lines) — `internal/semantic/analyze_special.go:358` — matched `cannot be used in class methods (static methods)`
- [ ] `Hint: Private method "X" declared but never used` (1 fixtures, 2 lines) — `internal/semantic/unused_warnings.go:204` — matched `Private method`
- [ ] `Syntax Error: Array bounds are of different types` (1 fixtures, 2 lines) — `internal/semantic/type_resolution.go:644` — matched `Array bounds are of different types`
- [ ] `Syntax Error: Incompatible operands` (1 fixtures, 2 lines) — `internal/semantic/errors.go:231` — matched `Incompatible operands`
- [ ] `Syntax Error: array element N has type Integer, expected TEnum` (1 fixtures, 2 lines) — `internal/interp/evaluator/array_helpers.go:483`, `internal/interp/evaluator/array_helpers.go:509`, `internal/interp/evaluator/overload_resolution.go:430`, `internal/interp/ffi_callback.go:121`, `internal/interp/marshal.go:109`, `internal/parser/statements.go:312`, `internal/parser/types.go:537`, `internal/parser/types.go:553`, `+9 more` — matched `array element`
- [ ] `Syntax Error: binding 'X' for operator 'X' expects N parameters, got N` (1 fixtures, 2 lines) — `internal/builtins/array.go:357`, `internal/semantic/analyze_operators.go:209`, `internal/semantic/analyze_operators.go:63` — matched `parameters, got`
- [ ] `Syntax Error: function 'X' element argument has type procedure(), expected TMyEnum` (1 fixtures, 2 lines) — `internal/semantic/analyze_builtin_array.go:296` — matched `element argument has type`
- [ ] `Syntax Error: function 'X' expects array, enum, or type name, got class of TObject` (1 fixtures, 2 lines) — `internal/semantic/analyze_builtin_array.go:101`, `internal/semantic/analyze_builtin_array.go:55` — matched `expects array, enum, or type name, got`
- [ ] `Syntax Error: implementation return type for 'X' does not match forward declaration` (1 fixtures, 2 lines) — `internal/semantic/symbol_table.go:436` — matched `implementation return type for`
- [ ] `Syntax Error: promoted property 'X' not found in an ancestor of class 'X'` (1 fixtures, 2 lines) — `internal/semantic/analyze_properties.go:73` — matched `not found in an ancestor of class`
- [ ] `Syntax Error: set element must be an ordinal value, got Float` (1 fixtures, 2 lines) — `internal/semantic/analyze_literals.go:514` — matched `set element must be an ordinal value, got`
- [ ] `Syntax Error: type mismatch in set literal: expected set of String, got set of Integer` (1 fixtures, 2 lines) — `internal/interp/evaluator/set_helpers.go:78`, `internal/semantic/analyze_literals.go:546` — matched `type mismatch in set literal: expected set of`
- [ ] `Syntax Error: unary - requires numeric operand, got TObject` (1 fixtures, 2 lines) — `internal/semantic/analyze_expr_operators.go:834` — matched `requires numeric operand, got`
- [ ] `Syntax Error: unknown parameter type 'X' in interface method 'X'` (1 fixtures, 2 lines) — `internal/semantic/analyze_interfaces.go:70`, `internal/semantic/analyze_interfaces.go:83` — matched `in interface method`
- [ ] `Warning: "X" has been deprecated` (1 fixtures, 2 lines) — `internal/semantic/analyze_builtin_string_transform.go:127` — matched `has been deprecated`
- [ ] `Runtime Error: constructor failed: ERROR: error evaluating default value: ERROR: cannot infer type for empty array literal` (1 fixtures, 1 lines) — `internal/interp/evaluator/array_helpers.go:272`, `internal/semantic/analyze_literals.go:78` — matched `cannot infer type for empty array literal`
- [ ] `Syntax Error: 'X' can only be used inside a class method` (1 fixtures, 1 lines) — `internal/semantic/analyze_special.go:21`, `internal/semantic/analyze_special.go:346` — matched `can only be used inside a class method`
- [ ] `Syntax Error: 'X' is not a function or procedure (got Integer)` (1 fixtures, 1 lines) — `internal/semantic/analyze_function_pointers.go:283` — matched `is not a function or procedure (got`
- [ ] `Syntax Error: 'X' operator requires a class reference for a metaclass cast, got TObject` (1 fixtures, 1 lines) — `internal/semantic/analyze_expressions.go:397` — matched `operator requires a class reference for a metaclass cast, got`
- [ ] `Syntax Error: 'X' operator requires class instance or class reference, got String` (1 fixtures, 1 lines) — `internal/semantic/analyze_expressions.go:497` — matched `operator requires class instance or class reference, got`
- [ ] `Syntax Error: 'X' operator requires class instance or interface, got Integer` (1 fixtures, 1 lines) — `internal/semantic/analyze_expressions.go:430` — matched `operator requires class instance or interface, got`
- [ ] `Syntax Error: 'X' operator requires class instance or interface, got TClass` (1 fixtures, 1 lines) — `internal/semantic/analyze_expressions.go:430` — matched `operator requires class instance or interface, got`
- [ ] `Syntax Error: 'X' operator requires class instance, got Integer` (1 fixtures, 1 lines) — `internal/semantic/analyze_expressions.go:328` — matched `operator requires class instance, got`
- [ ] `Syntax Error: Bare raise statement is only valid inside an exception handler` (1 fixtures, 1 lines) — `internal/semantic/analyze_exceptions.go:19` — matched `Bare raise statement is only valid inside an exception handler`
- [ ] `Syntax Error: Cannot assign () -> Void to procedure () variable 'X'` (1 fixtures, 1 lines) — `internal/errors/errors.go:307`, `internal/errors/errors.go:309`, `internal/errors/errors.go:393`, `internal/semantic/analyze_statements.go:480`, `internal/semantic/analyze_statements.go:691`, `internal/semantic/analyze_statements.go:695`, `internal/semantic/analyzer.go:876`, `internal/semantic/errors.go:159`, `+2 more` — matched `Cannot assign`
- [ ] `Syntax Error: Cannot assign Integer to String` (1 fixtures, 1 lines) — `internal/errors/errors.go:307`, `internal/errors/errors.go:309`, `internal/errors/errors.go:393`, `internal/semantic/analyze_statements.go:480`, `internal/semantic/analyze_statements.go:691`, `internal/semantic/analyze_statements.go:695`, `internal/semantic/analyzer.go:876`, `internal/semantic/errors.go:159`, `+2 more` — matched `Cannot assign`
- [ ] `Syntax Error: Invalid Operands` (1 fixtures, 1 lines) — `internal/semantic/errors.go:221` — matched `Invalid Operands`
- [ ] `Syntax Error: Record has no field members` (1 fixtures, 1 lines) — `internal/semantic/analyze_records.go:67` — matched `Record has no field members`
- [ ] `Syntax Error: Record type "X" is not fully defined` (1 fixtures, 1 lines) — `internal/semantic/analyze_lambdas.go:75` — matched `not fully`
- [ ] `Syntax Error: There is already a method with name "X"` (1 fixtures, 1 lines) — `internal/semantic/symbol_table.go:1044` — matched `There is already a method with name`
- [ ] `Syntax Error: array constructor range bounds must have the same type: got Integer and TNum` (1 fixtures, 1 lines) — `internal/semantic/analyze_literals.go:247` — matched `array constructor range bounds must have the same type: got`
- [ ] `Syntax Error: array constructor range bounds must have the same type: got TAlpha and Integer` (1 fixtures, 1 lines) — `internal/semantic/analyze_literals.go:247` — matched `array constructor range bounds must have the same type: got`
- [ ] `Syntax Error: array constructor range bounds must have the same type: got TAlpha and TNum` (1 fixtures, 1 lines) — `internal/semantic/analyze_literals.go:247` — matched `array constructor range bounds must have the same type: got`
- [ ] `Syntax Error: array element N has type Float, expected Integer` (1 fixtures, 1 lines) — `internal/interp/evaluator/array_helpers.go:483`, `internal/interp/evaluator/array_helpers.go:509`, `internal/interp/evaluator/overload_resolution.go:430`, `internal/interp/ffi_callback.go:121`, `internal/interp/marshal.go:109`, `internal/parser/statements.go:312`, `internal/parser/types.go:537`, `internal/parser/types.go:553`, `+9 more` — matched `array element`
- [ ] `Syntax Error: array element N has type procedure(array of Float), expected Float` (1 fixtures, 1 lines) — `internal/interp/evaluator/array_helpers.go:483`, `internal/interp/evaluator/array_helpers.go:509`, `internal/interp/evaluator/overload_resolution.go:430`, `internal/interp/ffi_callback.go:121`, `internal/interp/marshal.go:109`, `internal/parser/statements.go:312`, `internal/parser/types.go:537`, `internal/parser/types.go:553`, `+9 more` — matched `array element`
- [ ] `Syntax Error: array element N has type procedure(array of Integer), expected Integer` (1 fixtures, 1 lines) — `internal/interp/evaluator/array_helpers.go:483`, `internal/interp/evaluator/array_helpers.go:509`, `internal/interp/evaluator/overload_resolution.go:430`, `internal/interp/ffi_callback.go:121`, `internal/interp/marshal.go:109`, `internal/parser/statements.go:312`, `internal/parser/types.go:537`, `internal/parser/types.go:553`, `+9 more` — matched `array element`
- [ ] `Syntax Error: array element N has type procedure(array of procedure()), expected procedure()` (1 fixtures, 1 lines) — `internal/interp/evaluator/array_helpers.go:483`, `internal/interp/evaluator/array_helpers.go:509`, `internal/interp/evaluator/overload_resolution.go:430`, `internal/interp/ffi_callback.go:121`, `internal/interp/marshal.go:109`, `internal/parser/statements.go:312`, `internal/parser/types.go:537`, `internal/parser/types.go:553`, `+9 more` — matched `array element`
- [ ] `Syntax Error: binding 'X' parameter N type Integer does not match operator operand type Float` (1 fixtures, 1 lines) — `internal/interp/evaluator/runtime_ops.go:644`, `internal/interp/evaluator/visitor_declarations.go:489`, `internal/interp/evaluator/visitor_declarations.go:726`, `internal/interp/runtime/class_declaration.go:317`, `internal/semantic/analyze_operators.go:157`, `internal/semantic/analyze_operators.go:165`, `internal/semantic/analyze_operators.go:209`, `internal/semantic/analyze_operators.go:221`, `+7 more` — matched `binding`
- [ ] `Syntax Error: binding 'X' parameter N type Integer does not match operator operand type TObject` (1 fixtures, 1 lines) — `internal/interp/evaluator/runtime_ops.go:644`, `internal/interp/evaluator/visitor_declarations.go:489`, `internal/interp/evaluator/visitor_declarations.go:726`, `internal/interp/runtime/class_declaration.go:317`, `internal/semantic/analyze_operators.go:157`, `internal/semantic/analyze_operators.go:165`, `internal/semantic/analyze_operators.go:209`, `internal/semantic/analyze_operators.go:221`, `+7 more` — matched `binding`
- [ ] `Syntax Error: binding 'X' parameter N type TObject does not match operator operand type Integer` (1 fixtures, 1 lines) — `internal/interp/evaluator/runtime_ops.go:644`, `internal/interp/evaluator/visitor_declarations.go:489`, `internal/interp/evaluator/visitor_declarations.go:726`, `internal/interp/runtime/class_declaration.go:317`, `internal/semantic/analyze_operators.go:157`, `internal/semantic/analyze_operators.go:165`, `internal/semantic/analyze_operators.go:209`, `internal/semantic/analyze_operators.go:221`, `+7 more` — matched `binding`
- [ ] `Syntax Error: class 'X' does not implement interface method 'X' from interface 'X'` (1 fixtures, 1 lines) — `internal/semantic/analyze_interfaces.go:215` — matched `does not implement interface method`
- [ ] `Syntax Error: class operator 'X' already defined for class 'X'` (1 fixtures, 1 lines) — `internal/semantic/analyze_operators.go:261` — matched `already defined for class`
- [ ] `Syntax Error: duplicate property 'X' in class 'X'` (1 fixtures, 1 lines) — `internal/semantic/analyze_properties.go:52`, `internal/semantic/analyze_properties.go:63` — matched `duplicate property`
- [ ] `Syntax Error: function 'X' arguments must have compatible types, got Integer and String` (1 fixtures, 1 lines) — `internal/semantic/analyze_builtin_math_utils.go:155` — matched `arguments must have compatible types, got`
- [ ] `Syntax Error: function 'X' arguments must have compatible types, got Integer and Void` (1 fixtures, 1 lines) — `internal/semantic/analyze_builtin_math_utils.go:155` — matched `arguments must have compatible types, got`
- [ ] `Syntax Error: function 'X' arguments must have compatible types, got Nil and TObject` (1 fixtures, 1 lines) — `internal/semantic/analyze_builtin_math_utils.go:155` — matched `arguments must have compatible types, got`
- [ ] `Syntax Error: function 'X' arguments must have compatible types, got String and Integer` (1 fixtures, 1 lines) — `internal/semantic/analyze_builtin_math_utils.go:155` — matched `arguments must have compatible types, got`
- [ ] `Syntax Error: function 'X' arguments must have compatible types, got TA(TObject) and Nil` (1 fixtures, 1 lines) — `internal/semantic/analyze_builtin_math_utils.go:155` — matched `arguments must have compatible types, got`
- [ ] `Syntax Error: function 'X' arguments must have compatible types, got TA(TObject) and TObject` (1 fixtures, 1 lines) — `internal/semantic/analyze_builtin_math_utils.go:155` — matched `arguments must have compatible types, got`
- [ ] `Syntax Error: function 'X' arguments must have compatible types, got TObject and TA(TObject)` (1 fixtures, 1 lines) — `internal/semantic/analyze_builtin_math_utils.go:155` — matched `arguments must have compatible types, got`
- [ ] `Syntax Error: function 'X' element argument has type Integer, expected TMyEnum` (1 fixtures, 1 lines) — `internal/semantic/analyze_builtin_array.go:296` — matched `element argument has type`
- [ ] `Syntax Error: function 'X' element argument has type String, expected TMyEnum` (1 fixtures, 1 lines) — `internal/semantic/analyze_builtin_array.go:296` — matched `element argument has type`
- [ ] `Syntax Error: function 'X' expects Integer as first argument, got Float` (1 fixtures, 1 lines) — `internal/semantic/analyze_builtin_math_basic.go:122`, `internal/semantic/builtin_diagnostic_policies.go:10`, `internal/semantic/builtin_diagnostic_policies.go:11`, `internal/semantic/builtin_diagnostic_policies.go:51`, `internal/semantic/builtin_diagnostic_policies.go:52`, `internal/semantic/builtin_diagnostic_policies.go:53`, `internal/semantic/builtin_diagnostic_policies.go:7` — matched `expects Integer as first argument, got`
- [ ] `Syntax Error: function 'X' expects Integer or Enum variable, got TMyObj(TObject)` (1 fixtures, 1 lines) — `internal/semantic/analyze_builtin_math_utils.go:30`, `internal/semantic/analyze_builtin_math_utils.go:62` — matched `expects Integer or Enum variable, got`
- [ ] `Syntax Error: function 'X' expects String as first argument, got function(): String` (1 fixtures, 1 lines) — `internal/interp/evaluator/string_helpers.go:171`, `internal/interp/evaluator/var_params.go:1167`, `internal/interp/evaluator/var_params.go:1243`, `internal/semantic/builtin_diagnostic_policies.go:12`, `internal/semantic/builtin_diagnostic_policies.go:37`, `internal/semantic/builtin_diagnostic_policies.go:71` — matched `expects String as first argument, got`
- [ ] `Syntax Error: function 'X' expects a type name as argument` (1 fixtures, 1 lines) — `internal/interp/evaluator/type_casts.go:409`, `internal/semantic/analyze_builtin_convert.go:71` — matched `expects a type name as argument`
- [ ] `Syntax Error: function 'X' expects array, enum, or type name, got Void` (1 fixtures, 1 lines) — `internal/semantic/analyze_builtin_array.go:101`, `internal/semantic/analyze_builtin_array.go:55` — matched `expects array, enum, or type name, got`
- [ ] `Syntax Error: incompatible types in coalesce operator: String and Integer` (1 fixtures, 1 lines) — `internal/semantic/analyze_expr_operators.go:371` — matched `incompatible types in coalesce operator:`
- [ ] `Syntax Error: incompatible types in coalesce operator: array of TSub(TTest) and array of TObject` (1 fixtures, 1 lines) — `internal/semantic/analyze_expr_operators.go:371` — matched `incompatible types in coalesce operator:`
- [ ] `Syntax Error: incompatible types in coalesce operator: array of TSub(TTest) and array of TTest(TObject)` (1 fixtures, 1 lines) — `internal/semantic/analyze_expr_operators.go:371` — matched `incompatible types in coalesce operator:`
- [ ] `Syntax Error: incompatible types in coalesce operator: array of TTest(TObject) and array of TObject` (1 fixtures, 1 lines) — `internal/semantic/analyze_expr_operators.go:371` — matched `incompatible types in coalesce operator:`
- [ ] `Syntax Error: incompatible types in coalesce operator: class of TObject and Integer` (1 fixtures, 1 lines) — `internal/semantic/analyze_expr_operators.go:371` — matched `incompatible types in coalesce operator:`
- [ ] `Syntax Error: incompatible types in coalesce operator: procedure() and String` (1 fixtures, 1 lines) — `internal/semantic/analyze_expr_operators.go:371` — matched `incompatible types in coalesce operator:`
- [ ] `Syntax Error: incompatible types in if-then-else: Integer and Nil` (1 fixtures, 1 lines) — `internal/semantic/analyze_expressions.go:563` — matched `incompatible types in if-then-else:`
- [ ] `Syntax Error: incompatible types in if-then-else: Integer and procedure()` (1 fixtures, 1 lines) — `internal/semantic/analyze_expressions.go:563` — matched `incompatible types in if-then-else:`
- [ ] `Syntax Error: incompatible types in if-then-else: Nil and String` (1 fixtures, 1 lines) — `internal/semantic/analyze_expressions.go:563` — matched `incompatible types in if-then-else:`
- [ ] `Syntax Error: incompatible types in if-then-else: String and Integer` (1 fixtures, 1 lines) — `internal/semantic/analyze_expressions.go:563` — matched `incompatible types in if-then-else:`
- [ ] `Syntax Error: incompatible types in if-then-else: procedure() and Integer` (1 fixtures, 1 lines) — `internal/semantic/analyze_expressions.go:563` — matched `incompatible types in if-then-else:`
- [ ] `Syntax Error: inferred lambda return type function(): Integer incompatible with expected return type Integer` (1 fixtures, 1 lines) — `internal/semantic/analyze_lambdas.go:272`, `internal/semantic/analyze_lambdas.go:299` — matched `incompatible with expected return type`
- [ ] `Syntax Error: old() references undefined identifier 'X' in function 'X'` (1 fixtures, 1 lines) — `internal/semantic/analyze_functions.go:385` — matched `old() references undefined identifier`
- [ ] `Syntax Error: property 'X' cannot be read-accessed` (1 fixtures, 1 lines) — `internal/semantic/analyze_expr_operators.go:174` — matched `cannot be read-accessed`
- [ ] `Syntax Error: property 'X' cannot combine index parameters with an index directive` (1 fixtures, 1 lines) — `internal/semantic/analyze_properties.go:126` — matched `cannot combine index parameters with an index directive`
- [ ] `Syntax Error: property 'X' index directive must be an integer literal` (1 fixtures, 1 lines) — `internal/semantic/analyze_properties.go:147`, `internal/semantic/analyze_properties.go:154` — matched `index directive must be an integer literal`
- [ ] `Syntax Error: property 'X' read expression has type array of Variant, expected String` (1 fixtures, 1 lines) — `internal/semantic/analyze_properties.go:423` — matched `read expression has type`
- [ ] `Syntax Error: property 'X' write field 'X' has type Integer, expected String` (1 fixtures, 1 lines) — `internal/semantic/analyze_properties.go:468`, `internal/semantic/analyze_properties.go:487` — matched `write field`
- [ ] `Syntax Error: range end must be an ordinal type, got Void` (1 fixtures, 1 lines) — `internal/semantic/analyze_literals.go:484` — matched `range end must be an ordinal type, got`
- [ ] `Syntax Error: required parameter 'X' cannot come after optional parameters in function 'X'` (1 fixtures, 1 lines) — `internal/semantic/analyze_functions.go:89` — matched `cannot come after optional parameters in function`
- [ ] `Syntax Error: type mismatch in set literal: expected set of Integer, got set of String` (1 fixtures, 1 lines) — `internal/interp/evaluator/set_helpers.go:78`, `internal/semantic/analyze_literals.go:546` — matched `type mismatch in set literal: expected set of`
- [ ] `Syntax Error: unary + requires numeric operand, got TObject` (1 fixtures, 1 lines) — `internal/semantic/analyze_expr_operators.go:834` — matched `requires numeric operand, got`
- [ ] `Syntax Error: unary not requires Boolean, Integer, or Variant operand, got procedure()` (1 fixtures, 1 lines) — `internal/semantic/analyze_expr_operators.go:855` — matched `unary not requires Boolean, Integer, or Variant operand, got`
- [ ] `Syntax Error: unbound method pointers (@TClass.Destroy) are not supported` (1 fixtures, 1 lines) — `internal/semantic/analyze_function_pointers.go:160` — matched `unbound method pointers`
- [ ] `Syntax Error: unbound method pointers (@TClass.Free) are not supported` (1 fixtures, 1 lines) — `internal/semantic/analyze_function_pointers.go:160` — matched `unbound method pointers`
- [ ] `Syntax Error: unknown return type 'X' in interface method 'X'` (1 fixtures, 1 lines) — `internal/semantic/analyze_interfaces.go:70`, `internal/semantic/analyze_interfaces.go:83` — matched `in interface method`

### Frontend — `internal/frontend` — 6 shapes, 40 lines, 23 fixtures

- [ ] `Syntax Error: Name expected` (8 fixtures, 8 lines) — `internal/frontend/result.go:726`, `internal/frontend/result.go:732`, `internal/frontend/result.go:734`, `internal/frontend/result.go:825`, `internal/parser/classes.go:737`, `internal/parser/parser.go:328`, `internal/parser/qualified_names.go:36` — matched `Name expected`
- [ ] `Syntax Error: Class method or constructor expected` (4 fixtures, 9 lines) — `internal/frontend/result.go:539`, `internal/frontend/result.go:905`, `internal/frontend/result.go:908`, `internal/semantic/errors.go:506` — matched `Class method or constructor expected`
- [ ] `Syntax Error: Object reference needed to read/write an object field` (4 fixtures, 9 lines) — `internal/frontend/result.go:535`, `internal/frontend/result.go:904`, `internal/frontend/result.go:909`, `internal/semantic/errors.go:526` — matched `Object reference needed to read/write an object field`
- [ ] `Syntax Error: End of block expected` (4 fixtures, 6 lines) — `internal/frontend/result.go:720` — matched `End of block expected`
- [ ] `Syntax Error: Member symbol "X" is not visible from this scope` (3 fixtures, 5 lines) — `internal/frontend/result.go:706` — matched `is not visible from this scope`
- [ ] `Syntax Error: Expression expected before COLON` (3 fixtures, 3 lines) — `internal/frontend/result.go:883` — matched `Expression expected before COLON`

### Other — runtime, type system, shared error builders — 62 shapes, 289 lines, 145 fixtures

- [ ] `Syntax Error: Unknown name "X"` (45 fixtures, 104 lines) — `internal/errors/errors.go:297`, `internal/errors/errors.go:299`, `internal/frontend/result.go:838` — matched `Unknown name`
- [ ] `Syntax Error: Incompatible types: Cannot assign "X" to "X"` (17 fixtures, 32 lines) — `internal/errors/errors.go:393` — matched `Incompatible types: Cannot assign`
- [ ] `Syntax Error: Name "X" already exists` (9 fixtures, 11 lines) — `internal/errors/errors.go:356`, `internal/errors/errors.go:358`, `internal/interp/evaluator/type_resolution.go:161`, `internal/interp/types/function_registry.go:343` — matched `already exists`
- [ ] `Syntax Error: Method "X" of class "X" not implemented` (7 fixtures, 7 lines) — `internal/bytecode/vm_exec.go:1207`, `internal/frontend/result.go:505`, `internal/frontend/result.go:513`, `internal/frontend/result.go:518`, `internal/interp/errors/catalog.go:146`, `internal/interp/errors/catalog.go:299`, `internal/interp/evaluator/helper_methods.go:477`, `pkg/platform/wasm/platform.go:175`, `+1 more` — matched `not implemented`
- [ ] `Syntax Error: Incompatible types: "X" and "X"` (5 fixtures, 7 lines) — `internal/errors/errors.go:304`, `internal/errors/errors.go:393`, `internal/semantic/errors.go:680` — matched `Incompatible types`
- [ ] `Syntax Error: property 'X' read specifier 'X' not found in class 'X'` (5 fixtures, 6 lines) — `internal/interp/evaluator/class_property_helpers.go:28`, `internal/interp/evaluator/helper_methods.go:728`, `internal/interp/evaluator/property_read.go:257`, `internal/semantic/analyze_properties.go:376` — matched `read specifier`
- [ ] `Syntax Error: Argument N expects type "X" instead of "X"` (4 fixtures, 9 lines) — `internal/errors/errors.go:325`, `internal/errors/errors.go:334` — matched `expects type`
- [ ] `Syntax Error: There is no accessible member with name "X" for type TTest` (3 fixtures, 4 lines) — `internal/errors/errors.go:346`, `internal/interp/evaluator/json_namespace.go:82` — matched `There is no accessible member with name`
- [ ] `Syntax Error: Argument N expects type "X"` (3 fixtures, 3 lines) — `internal/errors/errors.go:325`, `internal/errors/errors.go:334` — matched `expects type`
- [ ] `Syntax Error: function 'X' argument N must be a variable` (2 fixtures, 11 lines) — `internal/builtins/var_param.go:279`, `internal/builtins/var_param.go:332`, `internal/builtins/var_param.go:437`, `internal/builtins/var_param.go:441`, `internal/builtins/var_param.go:473`, `internal/builtins/var_param.go:477`, `internal/builtins/var_param.go:522`, `internal/interp/evaluator/globalvars.go:54`, `+19 more` — matched `must be a variable`
- [ ] `Syntax Error: There is no accessible member with name "X" for type TChainItem` (2 fixtures, 9 lines) — `internal/errors/errors.go:346`, `internal/interp/evaluator/json_namespace.go:82` — matched `There is no accessible member with name`
- [ ] `Syntax Error: There is no accessible member with name "X" for type String` (2 fixtures, 6 lines) — `internal/errors/errors.go:346`, `internal/interp/evaluator/json_namespace.go:82` — matched `There is no accessible member with name`
- [ ] `Syntax Error: There is no accessible member with name "X" for type function` (2 fixtures, 5 lines) — `internal/errors/errors.go:346`, `internal/interp/evaluator/json_namespace.go:82` — matched `There is no accessible member with name`
- [ ] `Syntax Error: operator 'X' already defined for operand types (TObject, TObject)` (2 fixtures, 4 lines) — `internal/interp/evaluator/visitor_declarations.go:772`, `internal/interp/runtime/class_declaration.go:359`, `internal/semantic/analyze_operators.go:116` — matched `already defined for operand types`
- [ ] `Syntax Error: Incompatible parameter types - "X" expected (instead of "X")` (2 fixtures, 3 lines) — `internal/errors/errors.go:340` — matched `Incompatible parameter types -`
- [ ] `Syntax Error: There is no overloaded version of "X" that can be called with these arguments` (2 fixtures, 3 lines) — `internal/errors/errors.go:376`, `internal/interp/evaluator/overload_resolution.go:309` — matched `There is no overloaded version of`
- [ ] `Syntax Error: unknown type 'X' for field 'X'` (2 fixtures, 3 lines) — `internal/interp/evaluator/visitor_declarations.go:404`, `internal/interp/evaluator/visitor_declarations.go:415`, `internal/interp/evaluator/visitor_declarations.go:972`, `internal/interp/evaluator/visitor_declarations.go:981`, `internal/semantic/analyze_classes_decl.go:584`, `internal/semantic/analyze_classes_decl.go:592`, `internal/semantic/analyze_records.go:112`, `internal/semantic/analyze_records.go:123` — matched `for field`
- [ ] `Runtime Error: Object not instantiated` (2 fixtures, 2 lines) — `internal/interp/evaluator/member_assignment.go:125`, `internal/interp/evaluator/member_assignment.go:128`, `internal/interp/evaluator/method_dispatch.go:21`, `internal/interp/evaluator/method_dispatch.go:303`, `internal/interp/evaluator/method_dispatch.go:395`, `internal/interp/evaluator/method_dispatch.go:92`, `internal/interp/evaluator/object_method_helpers.go:223`, `internal/interp/evaluator/visitor_expressions_errors.go:42`, `+8 more` — matched `Object not instantiated`
- [ ] `Syntax Error: There is no accessible member with name "X" for type set of TMyEnum` (2 fixtures, 2 lines) — `internal/errors/errors.go:346`, `internal/interp/evaluator/json_namespace.go:82` — matched `There is no accessible member with name`
- [ ] `Syntax Error: cannot infer type for field 'X'` (2 fixtures, 2 lines) — `internal/interp/evaluator/visitor_declarations.go:415`, `internal/interp/evaluator/visitor_declarations.go:981`, `internal/semantic/analyze_classes_decl.go:592`, `internal/semantic/analyze_records.go:123` — matched `cannot infer type for field`
- [ ] `Error: Trying to create an instance of an abstract class` (1 fixtures, 6 lines) — `internal/errors/errors.go:382`, `internal/interp/evaluator/visitor_expressions_functions.go:1259`, `internal/semantic/errors.go:466` — matched `Trying to create an instance of an abstract class`
- [ ] `Syntax Error: Class "X" already defined` (1 fixtures, 3 lines) — `internal/errors/errors.go:364`, `internal/interp/evaluator/visitor_declarations.go:759`, `internal/interp/evaluator/visitor_declarations.go:772`, `internal/interp/runtime/class_declaration.go:359`, `internal/parser/functions.go:533`, `internal/semantic/analyze_operators.go:116`, `internal/semantic/analyze_operators.go:261`, `internal/semantic/analyze_operators.go:99`, `+3 more` — matched `already defined`
- [ ] `Syntax Error: There is no accessible member with name "X" for type IMy` (1 fixtures, 2 lines) — `internal/errors/errors.go:346`, `internal/interp/evaluator/json_namespace.go:82` — matched `There is no accessible member with name`
- [ ] `Syntax Error: There is no accessible member with name "X" for type TMyObj` (1 fixtures, 2 lines) — `internal/errors/errors.go:346`, `internal/interp/evaluator/json_namespace.go:82` — matched `There is no accessible member with name`
- [ ] `Syntax Error: There is no accessible member with name "X" for type Variant` (1 fixtures, 2 lines) — `internal/errors/errors.go:346`, `internal/interp/evaluator/json_namespace.go:82` — matched `There is no accessible member with name`
- [ ] `Syntax Error: cannot compare TObject with Integer` (1 fixtures, 2 lines) — `internal/interp/runtime/errors.go:84`, `internal/interp/runtime/primitives.go:163`, `internal/interp/runtime/primitives.go:194`, `internal/interp/runtime/primitives.go:238`, `internal/interp/runtime/primitives.go:251`, `internal/interp/runtime/primitives.go:323`, `internal/interp/runtime/primitives.go:57`, `internal/interp/runtime/primitives.go:98`, `+2 more` — matched `cannot compare`
- [ ] `Syntax Error: property 'X' setter method 'X' has N parameters, expected N parameter` (1 fixtures, 2 lines) — `internal/interp/evaluator/class_property_helpers.go:112`, `internal/interp/evaluator/helper_methods.go:922`, `internal/interp/evaluator/index_assignment.go:404`, `internal/interp/evaluator/index_assignment.go:501`, `internal/interp/evaluator/property_write.go:182`, `internal/interp/evaluator/property_write.go:201`, `internal/interp/evaluator/property_write.go:347`, `internal/interp/evaluator/property_write.go:355`, `+4 more` — matched `setter method`
- [ ] `Syntax Error: type mismatch in set literal: expected Integer, got String` (1 fixtures, 2 lines) — `internal/interp/evaluator/set_helpers.go:231`, `internal/interp/evaluator/set_helpers.go:235`, `internal/interp/evaluator/set_helpers.go:78`, `internal/semantic/analyze_literals.go:532`, `internal/semantic/analyze_literals.go:546` — matched `type mismatch in set literal: expected`
- [ ] `Warning: "X" has been deprecated: old stuff` (1 fixtures, 2 lines) — `cmd/dwscript/cmd/parse.go:40` — matched `deprecated:`
- [ ] `Runtime Error: "X" is not a valid floating point value` (1 fixtures, 1 lines) — `internal/builtins/conversion.go:274`, `internal/interp/evaluator/builtin_arg_coercion.go:140` — matched `is not a valid floating point value`
- [ ] `Runtime Error: Low() failed: Low() expects array, enum, string, or type name, got INTEGER` (1 fixtures, 1 lines) — `internal/interp/evaluator/context_bounds.go:79` — matched `Low() expects array, enum, string, or type name, got`
- [ ] `Runtime Error: Ord() expects enum, boolean, integer, or string, got VARIANT` (1 fixtures, 1 lines) — `internal/builtins/ordinal.go:74` — matched `Ord() expects enum, boolean, integer, or string, got`
- [ ] `Runtime Error: error evaluating argument N: ERROR: cannot assign nil to class of TObject` (1 fixtures, 1 lines) — `internal/interp/evaluator/overload_resolution.go:258` — matched `error evaluating argument`
- [ ] `Runtime Error: error evaluating argument N: ERROR: cannot determine type for array element N` (1 fixtures, 1 lines) — `internal/interp/evaluator/overload_resolution.go:258` — matched `error evaluating argument`
- [ ] `Runtime Error: member 'X' not found` (1 fixtures, 1 lines) — `internal/bytecode/compiler_expressions.go:125`, `internal/bytecode/compiler_expressions.go:160`, `internal/errors/errors.go:346`, `internal/frontend/result.go:712`, `internal/interp/errors/catalog.go:100`, `internal/interp/errors/catalog.go:101`, `internal/interp/evaluator/index_assignment.go:68`, `internal/interp/evaluator/json_namespace.go:82`, `+21 more` — matched `member`
- [ ] `Runtime Error: method 'X' not found for type 'X' in Test` (1 fixtures, 1 lines) — `internal/interp/evaluator/bytebuffer_methods.go:104`, `internal/interp/evaluator/method_dispatch.go:189`, `internal/interp/evaluator/method_dispatch.go:294`, `internal/interp/evaluator/method_dispatch.go:766`, `internal/interp/evaluator/method_dispatch.go:91` — matched `not found for type`
- [ ] `Runtime Error: method, property, or field 'X' not found in parent class 'X' in TChild.GetBase` (1 fixtures, 1 lines) — `internal/interp/evaluator/call_helpers.go:268`, `internal/interp/evaluator/call_helpers.go:283`, `internal/interp/runtime/object.go:336`, `internal/semantic/analyze_special.go:196` — matched `not found in parent class`
- [ ] `Runtime Error: type error in float operation: expected FLOAT or INTEGER, got STRING` (1 fixtures, 1 lines) — `internal/interp/evaluator/binary_ops.go:419`, `internal/interp/evaluator/binary_ops.go:429` — matched `type error in float operation: expected FLOAT or INTEGER, got`
- [ ] `Runtime Error: type mismatch: cannot add INTEGER to String in TTest.AppendStrings` (1 fixtures, 1 lines) — `internal/interp/evaluator/compound_ops.go:108`, `internal/interp/evaluator/compound_ops.go:125`, `internal/interp/evaluator/compound_ops.go:97` — matched `type mismatch: cannot add`
- [ ] `Runtime Error: undefined variable 'X'` (1 fixtures, 1 lines) — `internal/builtins/var_param.go:287`, `internal/builtins/var_param.go:340`, `internal/builtins/var_param.go:446`, `internal/builtins/var_param.go:450`, `internal/builtins/var_param.go:483`, `internal/builtins/var_param.go:487`, `internal/interp/errors/catalog.go:233`, `internal/interp/errors/catalog.go:86`, `+15 more` — matched `undefined variable`
- [ ] `Syntax Error: Constant "X" cannot be written to` (1 fixtures, 1 lines) — `cmd/dwscript/cmd/compile.go:156`, `cmd/dwscript/cmd/run.go:522`, `internal/bytecode/disasm.go:34`, `internal/bytecode/disasm.go:39`, `internal/lexer/directives.go:478`, `internal/semantic/analyze_const_instruction.go:54` — matched `Constant`
- [ ] `Syntax Error: There is no accessible member with name "X" for type TMyEnum` (1 fixtures, 1 lines) — `internal/errors/errors.go:346`, `internal/interp/evaluator/json_namespace.go:82` — matched `There is no accessible member with name`
- [ ] `Syntax Error: There is no accessible member with name "X" for type array of Variant` (1 fixtures, 1 lines) — `internal/errors/errors.go:346`, `internal/interp/evaluator/json_namespace.go:82` — matched `There is no accessible member with name`
- [ ] `Syntax Error: There is no accessible member with name "X" for type class of TBase` (1 fixtures, 1 lines) — `internal/errors/errors.go:346`, `internal/interp/evaluator/json_namespace.go:82` — matched `There is no accessible member with name`
- [ ] `Syntax Error: There is no accessible member with name "X" for type e` (1 fixtures, 1 lines) — `internal/errors/errors.go:346`, `internal/interp/evaluator/json_namespace.go:82` — matched `There is no accessible member with name`
- [ ] `Syntax Error: cannot compare TFooClass with TFoo(TObject)` (1 fixtures, 1 lines) — `internal/interp/runtime/errors.go:84`, `internal/interp/runtime/primitives.go:163`, `internal/interp/runtime/primitives.go:194`, `internal/interp/runtime/primitives.go:238`, `internal/interp/runtime/primitives.go:251`, `internal/interp/runtime/primitives.go:323`, `internal/interp/runtime/primitives.go:57`, `internal/interp/runtime/primitives.go:98`, `+2 more` — matched `cannot compare`
- [ ] `Syntax Error: cannot infer type for field 'X' in record 'X'` (1 fixtures, 1 lines) — `internal/interp/evaluator/visitor_declarations.go:415`, `internal/interp/evaluator/visitor_declarations.go:981`, `internal/semantic/analyze_classes_decl.go:592`, `internal/semantic/analyze_records.go:123` — matched `cannot infer type for field`
- [ ] `Syntax Error: expected 'X' or 'X', got INT` (1 fixtures, 1 lines) — `internal/interp/errors/catalog.go:269`, `internal/interp/errors/catalog.go:281` — matched `got INT`
- [ ] `Syntax Error: function 'X' delta must be Integer, got String` (1 fixtures, 1 lines) — `internal/builtins/ordinals.go:67`, `internal/builtins/var_param.go:184`, `internal/builtins/var_param.go:93`, `internal/interp/evaluator/var_params.go:451`, `internal/semantic/analyze_builtin_math_utils.go:114`, `internal/semantic/analyze_builtin_math_utils.go:38`, `internal/semantic/analyze_builtin_math_utils.go:71` — matched `delta must be Integer, got`
- [ ] `Syntax Error: function 'X' expects array or string, got Integer` (1 fixtures, 1 lines) — `internal/builtins/array.go:58`, `internal/bytecode/vm_exec.go:540`, `internal/semantic/analyze_builtin_string_format.go:29` — matched `expects array or string, got`
- [ ] `Syntax Error: method, property, or field 'X' not found in parent class 'X'` (1 fixtures, 1 lines) — `internal/interp/evaluator/call_helpers.go:268`, `internal/interp/evaluator/call_helpers.go:283`, `internal/interp/runtime/object.go:336`, `internal/semantic/analyze_special.go:196` — matched `not found in parent class`
- [ ] `Syntax Error: operator -= not supported for type TObject` (1 fixtures, 1 lines) — `internal/interp/evaluator/compound_ops.go:171` — matched `operator -= not supported for type`
- [ ] `Syntax Error: operator /= not supported for type String` (1 fixtures, 1 lines) — `internal/interp/evaluator/compound_ops.go:254` — matched `operator /= not supported for type`
- [ ] `Syntax Error: property 'X' getter method 'X' parameter N has type Integer, expected String` (1 fixtures, 1 lines) — `internal/interp/evaluator/class_property_helpers.go:35`, `internal/interp/evaluator/helper_methods.go:750`, `internal/interp/evaluator/property_read.go:263`, `internal/interp/evaluator/property_read.go:267`, `internal/interp/evaluator/property_read.go:294`, `internal/interp/evaluator/property_read.go:306`, `internal/interp/evaluator/property_read.go:507`, `internal/semantic/analyze_properties.go:344`, `+2 more` — matched `getter method`
- [ ] `Syntax Error: property 'X' getter method 'X' returns Integer, expected String` (1 fixtures, 1 lines) — `internal/interp/evaluator/class_property_helpers.go:35`, `internal/interp/evaluator/helper_methods.go:750`, `internal/interp/evaluator/property_read.go:263`, `internal/interp/evaluator/property_read.go:267`, `internal/interp/evaluator/property_read.go:294`, `internal/interp/evaluator/property_read.go:306`, `internal/interp/evaluator/property_read.go:507`, `internal/semantic/analyze_properties.go:344`, `+2 more` — matched `getter method`
- [ ] `Syntax Error: property 'X' getter method 'X' returns String, expected TMypropertyType` (1 fixtures, 1 lines) — `internal/interp/evaluator/class_property_helpers.go:35`, `internal/interp/evaluator/helper_methods.go:750`, `internal/interp/evaluator/property_read.go:263`, `internal/interp/evaluator/property_read.go:267`, `internal/interp/evaluator/property_read.go:294`, `internal/interp/evaluator/property_read.go:306`, `internal/interp/evaluator/property_read.go:507`, `internal/semantic/analyze_properties.go:344`, `+2 more` — matched `getter method`
- [ ] `Syntax Error: property 'X' getter method 'X' returns Void, expected String` (1 fixtures, 1 lines) — `internal/interp/evaluator/class_property_helpers.go:35`, `internal/interp/evaluator/helper_methods.go:750`, `internal/interp/evaluator/property_read.go:263`, `internal/interp/evaluator/property_read.go:267`, `internal/interp/evaluator/property_read.go:294`, `internal/interp/evaluator/property_read.go:306`, `internal/interp/evaluator/property_read.go:507`, `internal/semantic/analyze_properties.go:344`, `+2 more` — matched `getter method`
- [ ] `Syntax Error: type mismatch in set literal: expected String, got Integer` (1 fixtures, 1 lines) — `internal/interp/evaluator/set_helpers.go:231`, `internal/interp/evaluator/set_helpers.go:235`, `internal/interp/evaluator/set_helpers.go:78`, `internal/semantic/analyze_literals.go:532`, `internal/semantic/analyze_literals.go:546` — matched `type mismatch in set literal: expected`
- [ ] `Syntax Error: unknown target type 'X' for helper 'X'` (1 fixtures, 1 lines) — `internal/interp/evaluator/helper_methods.go:974`, `internal/interp/evaluator/visitor_declarations.go:1146`, `internal/semantic/analyze_helpers.go:48` — matched `unknown target type`
- [ ] `Unsupported character #N` (1 fixtures, 1 lines) — `internal/builtins/encoding.go:228`, `internal/interp/evaluator/string_helpers.go:581` — matched `Unsupported character #`
- [ ] `Upper bound exceeded! Index N` (1 fixtures, 1 lines) — `internal/interp/evaluator/json_methods.go:335`, `internal/interp/evaluator/visitor_expressions_functions.go:672`, `internal/semantic/analyze_statements.go:712` — matched `Upper bound exceeded! Index`
- [ ] `member assignment not supported for type STRING` (1 fixtures, 1 lines) — `internal/interp/evaluator/member_assignment.go:359` — matched `member assignment not supported for type`

### Unlocated — 17 shapes, 70 lines, 52 fixtures

- [ ] `Syntax Error: "X" expected` (21 fixtures, 34 lines) — not found; tried `grep -rnF "expected"`
- [ ] `Syntax Error: unknown type 'X'` (11 fixtures, 12 lines) — not found; tried `grep -rnF "unknown type"`
- [ ] `Syntax Error: expected 'X' after external` (3 fixtures, 4 lines) — not found; tried `grep -rnF "after external"`
- [ ] _(empty message text: a positioned diagnostic whose message is blank)_ (3 fixtures, 3 lines) — not found; nothing to search for
- [ ] `Syntax Error: expected 'X' after 'X'` (2 fixtures, 2 lines) — not found; tried `grep -rnF "expected"`
- [ ] `Syntax Error: function 'X' expects N arguments, got N` (2 fixtures, 2 lines) — not found; tried `grep -rnF "arguments, got"`
- [ ] `Syntax Error: expected 'X', 'X', or 'X' after 'X'` (1 fixtures, 2 lines) — not found; tried `grep -rnF "expected"`
- [ ] `Syntax Error: expected 'X', got WRITE` (1 fixtures, 2 lines) — not found; tried `grep -rnF "got WRITE"`
- [ ] `Compile Error: aborted` (1 fixtures, 1 lines) — not found; tried `grep -rnF "aborted"`
- [ ] `Syntax Error: Colon "X" expected` (1 fixtures, 1 lines) — not found; tried `grep -rnF "expected"`
- [ ] `Syntax Error: expected 'X' after empty` (1 fixtures, 1 lines) — not found; tried `grep -rnF "after empty"`
- [ ] `Syntax Error: expected 'X' or 'X', got IDENT` (1 fixtures, 1 lines) — not found; tried `grep -rnF "got IDENT"`
- [ ] `Syntax Error: expected 'X' or 'X', got RPAREN` (1 fixtures, 1 lines) — not found; tried `grep -rnF "got RPAREN"`
- [ ] `Syntax Error: expected 'X' or 'X', got THEN` (1 fixtures, 1 lines) — not found; tried `grep -rnF "expected"`
- [ ] `Syntax Error: function 'X' expects N argument, got N` (1 fixtures, 1 lines) — not found; tried `grep -rnF "argument, got"`
- [ ] `Syntax Error: function 'X' expects N or N arguments, got N` (1 fixtures, 1 lines) — not found; tried `grep -rnF "arguments, got"`
- [ ] `Syntax Error: function 'X' expects N-N arguments, got N` (1 fixtures, 1 lines) — not found; tried `grep -rnF "arguments, got"`

## How to regenerate

```bash
go build -o bin/dwscript ./cmd/dwscript
go run ./cmd/fixture-report --build=false --allow-stale --cli bin/dwscript \
  --in-scope --classify --list-fails --shape-top 0 --shape-fixtures
```

The run takes a few minutes. The two tables above are the `Missing diagnostics` and
`Spurious diagnostics` sections of its output, transcribed; the origin buckets are the greps
described above, run over the same tree. Do not hand-edit the numbers — replace them with a fresh
run, and say which commit it measured.
