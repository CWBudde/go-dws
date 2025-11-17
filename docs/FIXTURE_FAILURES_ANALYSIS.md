# SimpleScripts Test Failures Analysis

**Generated:** 2025-11-15
**Test Suite:** TestDWScriptFixtures/SimpleScripts
**Total Tests:** 292 failing tests

## Executive Summary

This document analyzes all failing tests in the SimpleScripts fixture suite, categorizing them by failure type and root cause. The failures reveal 46 distinct categories of missing features or incorrect implementations across the compiler pipeline.

### Failure Breakdown by Compiler Stage

| Stage | Count | Percentage |
|-------|-------|------------|
| Semantic Analysis | 121 | 41.4% |
| Parser | 73 | 25.0% |
| Output/Runtime Behavior | 55 | 18.8% |
| Runtime Errors | 43 | 14.7% |

---

## Priority Classification

### P0 - Critical Language Features (Must Have)

These are core language features that many tests depend on. Implementing these will unlock the most tests.

#### 1. Variant Type Support (14 tests) - SEMANTIC/RUNTIME
**Impact:** High - Affects type system fundamentals

**Missing Features:**
- Variant type operations and comparisons
- Null/Unassigned variant values
- Variant arithmetic and logical operators
- Type coercion with variants

**Tests:**
- assert_variant
- boolean_optimize
- case_variant_condition
- coalesce_variant
- compare_vars
- for_to_variant
- var_eq_nil
- var_nan
- var_param_casts
- variant_compound_ops
- variant_logical_ops
- variant_ops
- variant_unassigned_equal
- variants_as_casts
- variants_binary_bool_int
- variants_casts
- variants_is_bool

**Example:**
```pascal
var v : Variant;
v := Null;
PrintLn(v ?? 2);  // Coalesce with Null
```

---

#### 2. Class Variables (class var) (26 tests) - PARSE/SEMANTIC
**Impact:** High - Core OOP feature

**Missing Features:**
- `class var` declarations in classes
- Access to class variables via type name (e.g., `TBase.Test`)
- Access to class variables via instance
- Initialization of class variables
- Inheritance of class variables

**Tests:**
- class_var
- class_var_as_prop
- class_var_dyn1
- class_var_dyn2
- class_method3
- class_method4
- static_class
- static_class_array
- static_method1
- static_method2

**Example:**
```pascal
type TBase = class
  class var Test : Integer;
end;
TBase.Test := 123;  // Access via type name
```

---

#### 3. Class Constants (class const) (10 tests) - PARSE/SEMANTIC
**Impact:** High - Required for many class patterns

**Missing Features:**
- `const` declarations inside class bodies
- Field initialization from class constants (e.g., `FField := Value`)
- Access to class constants

**Tests:**
- class_const2
- class_const4
- class_init
- const_block
- enum_element_deprecated

**Example:**
```pascal
type TObj = class
  const Value = 5;
  FField := Value;  // Initialize from constant
end;
```

---

#### 4. Self Keyword in Class Methods (2 tests) - SEMANTIC
**Impact:** High - Fundamental OOP feature

**Missing Features:**
- `Self` keyword recognition in methods
- `Self` type resolution
- Access to `Self.ClassName`, `Self.ClassType`, etc.

**Tests:**
- class_self
- self

**Example:**
```pascal
class procedure TBase.MyProc;
begin
  if Assigned(Self) then
    PrintLn(Self.ClassName);
end;
```

---

#### 5. Function/Method Pointers (12 tests) - PARSE/SEMANTIC
**Impact:** High - Advanced but commonly used feature

**Missing Features:**
- Function pointer type declarations (e.g., `TMyProc = procedure`)
- Method pointer type declarations
- Assignment of functions to pointer variables
- Calling function pointers
- Passing function pointers as parameters

**Tests:**
- func_ptr1
- func_ptr3
- func_ptr4
- func_ptr5
- func_ptr_assigned
- func_ptr_class_meth
- func_ptr_classname
- func_ptr_constant
- func_ptr_field
- func_ptr_field_no_param
- func_ptr_param
- func_ptr_property
- func_ptr_property_alias
- func_ptr_symbol_field
- func_ptr_var
- func_ptr_var_param
- meth_ptr1
- proc_of_method
- stack_of_proc

**Example:**
```pascal
type TMyProc = procedure;

procedure Proc1;
begin
  PrintLn('Proc1');
end;

var p : TMyProc;
p := Proc1;
p();  // Call via pointer
```

---

#### 6. For-In Loops (13 tests) - PARSE/SEMANTIC/RUNTIME
**Impact:** Medium-High - Modern control flow

**Missing Features:**
- `for variable in collection` syntax
- Iteration over enumerations
- Iteration over arrays
- Iteration over strings
- Iteration over sets
- For-in with type inference

**Tests:**
- enumerations2
- for_in_array
- for_in_enum
- for_in_func_array
- for_in_record_array
- for_in_set
- for_in_str
- for_in_str2
- for_in_str4
- for_in_subclass
- for_var_in_array
- for_var_in_enumeration
- for_var_in_field_array
- for_var_in_string

**Example:**
```pascal
type TMyEnum = (meA, meB, meC);
var i : TMyEnum;
for i in TMyEnum do
  Print(i);
```

---

### P1 - Important Language Features

#### 7. Lazy Parameters (3 tests) - SEMANTIC
**Impact:** Medium - Optimization feature

**Missing Features:**
- `lazy` parameter modifier
- Deferred evaluation of lazy parameters
- Lazy parameter type checking

**Tests:**
- lazy
- lazy_recursive
- lazy_sqr

**Example:**
```pascal
function CondEval(eval : Boolean; lazy a : Integer) : Integer;
begin
  if eval then
    Result := a  // Only evaluated if needed
  else Result := 0;
end;
```

---

#### 8. Record Advanced Features (19 tests) - PARSE/SEMANTIC/RUNTIME
**Impact:** Medium - Value type features

**Missing Features:**
- Record field initialization syntax
- Record constants
- Record class variables
- Record methods (already partially supported)
- Nested records
- Record properties
- Record type inference

**Tests:**
- const_record
- record_aliased_field
- record_class_field_init
- record_clone1
- record_const
- record_const_as_prop
- record_dynamic_init
- record_field_init
- record_in_copy
- record_metaclass_field
- record_method
- record_method2
- record_method3
- record_method4
- record_method5
- record_nested
- record_nested2
- record_property
- record_record_field_init
- record_result
- record_result2
- record_result3
- record_passing
- record_recursive_dynarray
- record_static_array
- record_var
- record_var_as_prop
- record_var_param1
- record_var_param2
- result_direct
- string_record_field_get_set
- var_param_rec_field
- var_param_rec_method

---

#### 9. Property Advanced Features (9 tests) - PARSE/SEMANTIC
**Impact:** Medium - OOP encapsulation

**Missing Features:**
- Index properties (e.g., `property Items[i: Integer]`)
- Property calling syntax
- Property promotion from parent
- Property reintroduce
- Array-typed properties
- Properties in records

**Tests:**
- class_var_as_prop
- index_property
- property_call
- property_index
- property_of_as
- property_promotion
- property_reintroduce
- property_sub_default
- property_type_array

---

#### 10. Inheritance & Virtual Methods Issues (14 tests) - SEMANTIC/RUNTIME/OUTPUT
**Impact:** Medium - OOP polymorphism

**Missing Features:**
- Proper `override` validation
- `inherited` keyword in constructors
- `reintroduce` keyword
- Virtual constructor handling
- Deep override chains
- Parent class member access

**Tests:**
- class_forward
- class_parent
- clear_ref_in_constructor_assignment
- clear_ref_in_destructor
- destroy
- free_destroy
- inherited1
- inherited2
- inherited_constructor
- oop
- override_deep
- reintroduce
- reintroduce_virtual
- virtual_constructor
- virtual_constructor2

---

#### 11. Enum Advanced Features (12 tests) - PARSE/SEMANTIC
**Impact:** Medium - Type system

**Missing Features:**
- Enum boolean operations
- Enum bounds (Low/High)
- EnumByName function
- Enum flags/sets
- Scoped enums
- Enum to integer casts
- Enum element deprecation

**Tests:**
- aliased_enum
- enum_bool_op
- enum_bounds
- enum_byname
- enum_casts
- enum_element_deprecated
- enum_flags1
- enum_scoped
- enum_to_integer
- enumerations
- enumerations_names
- enumerations_qualifiednames

---

### P2 - Nice to Have Features

#### 12. Compiler Hints & Warnings (6 tests) - OUTPUT
**Impact:** Low - Developer experience

**Missing Features:**
- Case mismatch hints
- Calling convention hints
- Deprecation warnings
- General hint/warning system

**Tests:**
- call_conventions
- class_const3
- class_const_as_prop
- classname_nil_call
- comments_empty
- hint_warn

---

#### 13. String Literals & Formatting (6 tests) - PARSE
**Impact:** Low - Syntax sugar

**Missing Features:**
- Binary literal with underscores (e.g., `0b_101_111`)
- Heredoc strings
- Triple apostrophe strings
- String manipulation features

**Tests:**
- binary_literal
- heredoc_indent
- heredoc_special
- string_manip
- triple_apos1
- triple_apos2

---

#### 14. Contracts & Assertions (6 tests) - PARSE/SEMANTIC/RUNTIME
**Impact:** Low - Design by contract

**Missing Features:**
- Method contracts (require/ensure)
- Procedure contracts
- `old` keyword in contracts
- `implies` operator
- Contract inheritance

**Tests:**
- contracts_code
- contracts_old
- contracts_subproc
- implies
- method_contracts
- procedure_contracts

---

#### 15. Partial Classes (4 tests) - PARSE/SEMANTIC
**Impact:** Low - Code organization

**Missing Features:**
- Partial class declarations
- Partial class in units
- Partial class subclassing

**Tests:**
- partial_class2
- partial_class3
- partial_class_subclass
- partial_class_unit

---

#### 16. Miscellaneous Parse Features (25 tests) - PARSE
**Impact:** Low - Various syntax features

**Missing Features:**
- Increment/decrement operators (++/--)
- With statement
- Default function parameters in advanced scenarios
- External function declarations
- Inline function modifier
- Conditional compilation (#ifndef, #ifdef)
- Include expressions
- Reserved word handling
- Resource strings
- Try-except-finally combined
- Const array variations
- Field typed defaults
- Value type separation

**Tests:**
- class_const4
- conditionals_ifndef
- const_array4
- const_array7
- const_array8
- const_deprecated
- default_func
- exit_case
- external
- field_typed_default
- for_step_sign
- func_inline
- if_empty_terms
- ifthenelse_optimize1
- ifthenelse_optimize2
- include_expr
- new_class3
- new_class_alias
- plus_plus_minus_minus
- reserved_word
- resourcestring1
- try_except_finally
- value_type_separation
- variants_bool
- variants_strict
- with1

---

### P3 - Runtime & Output Discrepancies

These tests parse and analyze correctly but produce incorrect output or runtime behavior.

#### 17. Output Mismatches (55 tests) - OUTPUT
**Impact:** Variable - Indicates bugs in implementation

**Categories:**
- **Assert column positions (1):** Position tracking in assertions
  - assert

- **Type casting behavior (3):** Incorrect cast validation
  - casts_base_types (Integer cast rounds incorrectly)
  - class_cast (Missing runtime cast validation)
  - params_autocast

- **Virtual method dispatch (4):** Polymorphism issues
  - clear_ref_in_static_method
  - clear_ref_in_virtual_method
  - inherited1
  - reintroduce
  - reintroduce_virtual
  - virtual_constructor2

- **Exception handling (5):** Exception flow issues
  - exceptions2
  - exit
  - exit_constructor
  - exit_finally2
  - re_raise

- **For loop behavior (3):** Loop iteration bugs
  - for_step
  - for_var_in_enumeration
  - format

- **Include directive (4):** Include processing
  - include
  - includeSym
  - include_once
  - include_once2

- **Unicode (2):** Character encoding
  - unicode_identifiers
  - unicode_print

- **Record operations (5):** Record semantics
  - record_dynamic_init
  - record_method4
  - record_passing
  - record_result3
  - record_static_array

- **Case sensitivity (3):** Identifier casing
  - case_variant (variant case comparison)
  - class_const3
  - classname_nil_call

- **Other (21):** Various behavioral differences

---

#### 18. Runtime Errors (43 tests) - RUNTIME
**Impact:** High - Crashes/errors during execution

**Categories:**
- **Division by zero (4):** Missing zero checks
  - div_by_zero_float
  - div_by_zero_int
  - mod_by_zero_int
  - mul_div_optimize

- **Field access (7):** Incorrect member resolution
  - class_method (Methods accessed as fields)
  - class_method2 (Methods accessed as fields)
  - method1
  - partial_class_subclass
  - record_clone1
  - record_method3
  - virtual_constructor

- **Type resolution (2):** Runtime type lookup failures
  - aliased_vs_nil
  - func_ptr_property_alias

- **Built-in functions (1):** Incomplete stdlib
  - bool_string (StrToBool validation)

- **Other (19):** Various runtime issues

---

## Grouped by Root Cause

### TYPE SYSTEM (60 tests)
- Variant support (17 tests)
- Type aliases and resolution (5 tests)
- Enum features (12 tests)
- Metaclass/class-of (6 tests)
- Type casting (5 tests)
- Implicit conversions (5 tests)
- Associative arrays (2 tests)
- Function pointer types (19 tests)

### CLASS FEATURES (70 tests)
- Class variables (10 tests)
- Class constants (10 tests)
- Self keyword (2 tests)
- Properties (9 tests)
- Inheritance/override (14 tests)
- Class methods (5 tests)
- Partial classes (4 tests)
- Forward declarations (3 tests)
- Access control (3 tests)
- Other class issues (10 tests)

### RECORD FEATURES (19 tests)
- Field initialization (5 tests)
- Record methods (5 tests)
- Record constants (2 tests)
- Record properties (2 tests)
- Other record features (5 tests)

### CONTROL FLOW (25 tests)
- For-in loops (13 tests)
- Exception handling (7 tests)
- Exit statements (3 tests)
- With statement (1 test)
- Case statement variants (1 test)

### PARAMETERS (10 tests)
- Lazy parameters (3 tests)
- Default parameters (3 tests)
- Var parameters (2 tests)
- Open array (1 test)
- Const parameters (1 test)

### SYNTAX FEATURES (40 tests)
- String literals (6 tests)
- Binary/hex literals (1 test)
- Operators (++/--, implies) (2 tests)
- Contracts (6 tests)
- Conditional compilation (2 tests)
- Include directive (4 tests)
- Resource strings (1 test)
- Reserved words (1 test)
- Other parse issues (17 tests)

### BUILT-IN FUNCTIONS (10 tests)
- String operations (3 tests)
- Type operations (3 tests)
- Comparison operations (2 tests)
- Other stdlib (2 tests)

### OUTPUT/BEHAVIORAL (55 tests)
- Hints/warnings (6 tests)
- Position tracking (1 test)
- Virtual dispatch (6 tests)
- Type conversions (3 tests)
- Unicode (2 tests)
- Other output (37 tests)

---

## Implementation Recommendations

### Phase 1: Core Type System (60 tests → ~80 passing)
1. Implement Variant type with Null/Unassigned
2. Add function pointer types
3. Complete enum advanced features
4. Fix type resolution and aliases

### Phase 2: Class Features (70 tests → ~150 passing)
1. Add class var support
2. Add class const support with field initialization
3. Implement Self keyword
4. Fix inheritance and override validation
5. Add advanced property features

### Phase 3: Modern Language Features (38 tests → ~188 passing)
1. Implement for-in loops
2. Add lazy parameters
3. Complete record advanced features
4. Add function/method pointers

### Phase 4: Syntax & Polish (49 tests → ~237 passing)
1. Advanced string literals
2. Compiler hints/warnings
3. Contracts
4. Partial classes
5. Miscellaneous parse features

### Phase 5: Runtime & Output Fixes (55 tests → ~292 passing)
1. Fix virtual method dispatch
2. Correct type casting behavior
3. Fix exception handling
4. Resolve output discrepancies
5. Add missing runtime checks

---

## Test File Reference

Full list of all 292 failing tests organized alphabetically with their categories is available in the test output logs.

---

## Notes

- Many tests have multiple issues; fixing one category may not fully resolve a test
- Some tests (e.g., `class_method3`) fail on class var but also have other issues
- Priority estimates are based on number of tests affected and feature importance
- Implementation order should consider dependencies (e.g., class var before class const field init)
