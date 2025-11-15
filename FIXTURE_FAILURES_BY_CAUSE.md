# Test Failures Grouped by Implementation Cause

**Quick Reference:** Which tests will pass when implementing specific features

---

## VARIANT TYPE SUPPORT

### Basic Variant Operations (11 tests)
- assert_variant
- boolean_optimize
- case_variant_condition
- compare_vars
- variant_compound_ops
- variant_logical_ops
- variant_ops
- variants_as_casts
- variants_binary_bool_int
- variants_casts
- variants_bool
- variants_strict

### Null/Unassigned Support (5 tests)
- coalesce_variant
- var_eq_nil
- var_nan
- variant_unassigned_equal
- variants_is_bool

### Variant in For Loops (1 test)
- for_to_variant

### Variant Parameters (1 test)
- var_param_casts

---

## CLASS VAR (Class Variables)

### Basic Class Var (10 tests)
- class_var
- class_var_dyn1
- class_var_dyn2
- class_method3
- class_method4
- static_class
- static_class_array
- static_method1
- static_method2
- field_scope

### Class Var with Properties (1 test)
- class_var_as_prop

---

## CLASS CONST (Class Constants)

### Basic Class Const (3 tests)
- class_const2
- class_const3
- class_const_as_prop

### Class Const with Field Initialization (4 tests)
**Requires:** Parser support for `FField := Value` syntax in class body
- class_const4
- class_init
- const_block

---

## SELF KEYWORD

### Self in Class Methods (2 tests)
- class_self
- self

---

## FUNCTION POINTERS

### Function Pointer Types (7 tests)
**Requires:** Parse `TMyProc = procedure` syntax
- func_ptr1
- func_ptr3
- func_ptr_class_meth
- func_ptr_field
- func_ptr_field_no_param
- func_ptr_var_param
- meth_ptr1

### Function Pointer Semantics (12 tests)
**Requires:** Assignment, calling, passing as parameters
- func_ptr4
- func_ptr5
- func_ptr_assigned
- func_ptr_classname
- func_ptr_constant
- func_ptr_param
- func_ptr_property
- func_ptr_property_alias
- func_ptr_symbol_field
- func_ptr_var
- proc_of_method
- stack_of_proc

---

## FOR-IN LOOPS

### Parse For-In Syntax (4 tests)
**Requires:** `for var in collection` parsing
- for_in_array
- for_in_func_array
- for_var_in_string
- for_step_sign

### For-In Enumeration (6 tests)
**Requires:** Iterate over enum types
- enumerations2
- for_in_enum
- for_var_in_enumeration
- for_in_subclass
- for_var_in_array
- for_var_in_field_array

### For-In Strings (3 tests)
**Requires:** Iterate over string characters
- for_in_str
- for_in_str2
- for_in_str4

### For-In Sets & Records (2 tests)
- for_in_set
- for_in_record_array

---

## LAZY PARAMETERS

### Lazy Evaluation (3 tests)
**Requires:** Parse `lazy` modifier, defer evaluation
- lazy
- lazy_recursive
- lazy_sqr

---

## RECORD FEATURES

### Record Field Initialization (4 tests)
**Requires:** Parse `field := value` in record body
- record_field_init
- record_class_field_init
- record_record_field_init
- record_dynamic_init

### Record Constants (2 tests)
**Requires:** `const` in record declarations
- record_const
- record_const_as_prop

### Record Variables (2 tests)
**Requires:** `var` in record declarations
- record_var
- record_var_as_prop

### Record Methods (8 tests)
**Some parsing done, needs completion**
- record_method
- record_method2
- record_method3
- record_method4
- record_method5
- record_result
- record_result2
- record_passing

### Nested Records (2 tests)
- record_nested
- record_nested2

### Record Properties (1 test)
- record_property

### Record Advanced (8 tests)
- const_record
- record_aliased_field
- record_clone1
- record_in_copy
- record_metaclass_field
- record_recursive_dynarray
- record_result3
- record_static_array
- record_var_param1
- record_var_param2
- result_direct
- string_record_field_get_set
- var_param_rec_field
- var_param_rec_method

---

## PROPERTY FEATURES

### Index Properties (2 tests)
**Requires:** `property Items[Index: Integer]` syntax
- index_property
- property_index

### Property Advanced Syntax (7 tests)
- property_call
- property_of_as
- property_promotion
- property_reintroduce
- property_sub_default
- property_type_array
- class_property

---

## ENUM FEATURES

### Enum Parse Features (6 tests)
- enum_bool_op
- enum_bounds
- enum_byname
- enum_flags1
- enum_scoped
- enum_to_integer

### Enum Semantics (6 tests)
- aliased_enum
- enum_casts
- enum_element_deprecated
- enumerations
- enumerations_names
- enumerations_qualifiednames

---

## INHERITANCE & VIRTUAL METHODS

### Override Validation (5 tests)
- class_forward
- clear_ref_in_destructor
- override_deep
- reintroduce
- reintroduce_virtual

### Inherited Keyword (4 tests)
- clear_ref_in_constructor_assignment
- inherited1
- inherited2
- inherited_constructor

### Class Parent (1 test)
**Requires:** ClassParent property
- class_parent

### Destroy/Free (2 tests)
- destroy
- free_destroy

### Virtual Constructors (3 tests)
- virtual_constructor
- virtual_constructor2
- new_class1

### OOP Semantics (2 tests)
- oop
- oop_is_as

---

## STRING LITERALS

### Binary Literals with Underscores (1 test)
**Requires:** Parse `0b_101_111` and `%110_011`
- binary_literal

### Heredoc Strings (2 tests)
- heredoc_indent
- heredoc_special

### Triple Apostrophe (2 tests)
- triple_apos1
- triple_apos2

---

## COMPILER HINTS & WARNINGS

### Case Mismatch Hints (4 tests)
**Requires:** Track identifier case, emit hints
- class_const3
- class_const_as_prop
- classname_nil_call
- comments_empty

### Calling Convention Hints (1 test)
- call_conventions

### General Hints/Warnings (1 test)
- hint_warn

---

## CONTRACTS

### Contract Syntax (3 tests)
**Requires:** Parse require/ensure clauses
- method_contracts
- procedure_contracts
- implies

### Contract Semantics (2 tests)
- contracts_code
- contracts_old
- contracts_subproc

---

## PARTIAL CLASSES

### Partial Class Parsing (2 tests)
- partial_class3
- partial_class_unit

### Partial Class Semantics (2 tests)
- partial_class2
- partial_class_subclass

---

## OPERATORS

### Increment/Decrement (1 test)
**Requires:** Parse `++` and `--` operators
- plus_plus_minus_minus

### Implies Operator (1 test)
**Requires:** Parse `implies` operator
- implies

---

## CONDITIONAL COMPILATION

### Conditional Directives (2 tests)
**Requires:** Parse #ifndef, #ifdef, #endif
- conditionals_ifndef
- include_expr

---

## INCLUDE DIRECTIVE

### Include Semantics (4 tests)
**Requires:** Process include files correctly
- include
- includeSym
- include_once
- include_once2

---

## MISCELLANEOUS PARSE FEATURES

### With Statement (1 test)
- with1

### External Functions (1 test)
- external

### Inline Functions (1 test)
- func_inline

### Resource Strings (1 test)
- resourcestring1

### Reserved Words (1 test)
- reserved_word

### Try-Except-Finally (1 test)
- try_except_finally

### Const Arrays (4 tests)
- const_array4
- const_array6
- const_array7
- const_array8
- const_array_empty

### Const Deprecated (1 test)
- const_deprecated

### Default Function Parameters (4 tests)
- default_func
- default_parameters
- default_parameters_empty_array
- default_parameters_expr

### Exit in Case (1 test)
- exit_case

### Field Typed Default (1 test)
- field_typed_default

### If Empty Terms (1 test)
- if_empty_terms

### If-Then-Else Optimization (3 tests)
- ifthenelse_optimize1
- ifthenelse_optimize2
- ifthenelse_optimize3
- ifthenelse_expression6

### New Class Variations (4 tests)
- new_class2
- new_class3
- new_class4
- new_class_alias

### String Manipulation (1 test)
- string_manip

### Value Type Separation (1 test)
- value_type_separation

---

## SEMANTIC ANALYSIS ISSUES

### Type Resolution (5 tests)
- aliased_vs_nil
- class_of3
- class_of_cast
- conv_int_float
- record_metaclass_field

### Constructor Overload (1 test)
- constructor_overload

### Declared/Defined (2 tests)
- declared
- defined

### Empty Body (1 test)
- empty_body

### Exception Scoping (2 tests)
- exception_scoping
- exceptions

### Func Result as ByRef (1 test)
- func_result_as_byref

### Ignore Result (1 test)
- ignore_result

### Implicit Conversions (1 test)
- implicit_typed_int_to_float
- int_float

### Method Implementation (1 test)
- method_implem

### Mod Float (2 tests)
- mod_float
- mod_float_large

### Nested Declarations (1 test)
- nested_declarations

### Program Directive (2 tests)
- program
- programDot

### Stack Operations (3 tests)
- stack_depth
- stack_peek
- stacktrace

### String Built-in Methods (1 test)
- string_builtin_methods

### String Operations (3 tests)
- string_class_field_get_set
- string_in_op2
- string_in_op3

### Swap (1 test)
- swap2

### Unknown Statement Types (2 tests)
**Requires:** Implement UnitDeclaration
- class_scoping1
- partial_class2

### Var Param (3 tests)
- var_param_casts
- var_param_obj_method
- var_param_rec_field
- var_param_rec_method

### Event Types (2 tests)
- event_virtual
- func_ptr_assigned

### Coalesce Operators (2 tests)
- coalesce_class_2
- coalesce_dynarray

### Associative Arrays (1 test)
**Requires:** `array[Boolean] of Type` syntax
- boolean_array_of

### Format Function (2 tests)
- format
- format_incorrect_params

### For Loop Edge Cases (2 tests)
- for_step
- for_step_overflow

### Consts Expression (1 test)
- consts_expr

---

## RUNTIME ERRORS

### Division by Zero (4 tests)
**Requires:** Add zero checks before div/mod
- div_by_zero_float
- div_by_zero_int
- mod_by_zero_int
- mul_div_optimize

### Class Methods as Fields (2 tests)
**Requires:** Fix method lookup to not allow field-style access
- class_method
- class_method2

### Field Access Issues (5 tests)
- method1
- partial_class_subclass
- record_clone1
- record_method3
- virtual_constructor

### Built-in Functions (1 test)
**Requires:** Better StrToBool validation
- bool_string

### Nil Object Access (1 test)
**Requires:** Better nil checking in assigned test
- assigned

### In/Not-In Operators (2 tests)
- in_operator
- not_in_operator

### Ord Function (2 tests)
- ord
- ord_string_op

### String Array Access (2 tests)
- string_array_item_get_set
- string_bounds2

### Field As Var Parameter (1 test)
- field_as_var

### Inc/Dec Field (1 test)
- inc_dec_field

### Mixed Nil Array (1 test)
- mixed_nil_array

### Nil Meta Parameter (1 test)
- nil_meta_parameter

### Exception Nested Call (2 tests)
- exception_nested_call
- exceptobj2

---

## OUTPUT DISCREPANCIES

### Assert Position (1 test)
**Issue:** Column position incorrect in assert messages
- assert

### Cast Behavior (3 tests)
**Issue:** Integer cast should round, not truncate; missing runtime validation
- casts_base_types (3.5 → Integer should be 4, not 3)
- class_cast (Missing "Cannot cast" error)
- params_autocast

### Case Variant (1 test)
**Issue:** Variant case comparison logic
- case_variant

### Virtual Method Dispatch (2 tests)
**Issue:** Clear_ref in static/virtual methods not working
- clear_ref_in_static_method
- clear_ref_in_virtual_method

### Virtual Inheritance (4 tests)
- inherited1
- reintroduce
- reintroduce_virtual
- virtual_constructor2

### Exception Flow (5 tests)
**Issue:** Exit in except/finally blocks
- exceptions2
- exit
- exit_constructor
- exit_finally2
- re_raise

### Unicode (2 tests)
- unicode_identifiers
- unicode_print

### Open Array Bounds (1 test)
- open_array_bounds

### Recursive Path (1 test)
- recursive_path

### Static to Dynamic Array (1 test)
- static_to_dynamic_empty_array

### Conditionals (4 tests)
- conditionals
- conditionals_else
- conditionals_nested
- conditionals_nested3
- conditionals_nested4

### Const Params (1 test)
- const_params1

### Exception Object (1 test)
- exceptobj3

### Exception Nested Call 2 (1 test)
- exception_nested_call2

### Exit Finally (1 test)
- exit_finally

### Method Condition (1 test)
- method_condition

### Method with Name of Prop (1 test)
- method_with_name_of_prop

### OOP Field (1 test)
- oop_field

---

## Summary Statistics

| Category | Test Count |
|----------|------------|
| Variant Type | 17 |
| Class Var | 11 |
| Class Const | 7 |
| Self Keyword | 2 |
| Function Pointers | 19 |
| For-In Loops | 15 |
| Lazy Parameters | 3 |
| Record Features | 26 |
| Property Features | 8 |
| Enum Features | 12 |
| Inheritance & Virtual | 20 |
| String Literals | 5 |
| Compiler Hints | 6 |
| Contracts | 5 |
| Partial Classes | 4 |
| Operators | 2 |
| Conditional Compilation | 2 |
| Include Directive | 4 |
| Misc Parse | 30 |
| Semantic Issues | 35 |
| Runtime Errors | 28 |
| Output Discrepancies | 32 |

**Total:** 292 failing tests
